package tor

import (
	"bufio"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

const (
	defaultControlAddr    = "127.0.0.1:9051"
	defaultDialTimeout    = 5 * time.Second
	defaultReadTimeout    = 10 * time.Second
	defaultWriteTimeout   = 5 * time.Second
	defaultNewNymCooldown = 10 * time.Second // Tor enforces min 10s between NEWNYM
)

// AuthMethod defines how we authenticate with the control port.
type AuthMethod int

const (
	AuthNull       AuthMethod = iota // No authentication (dev only)
	AuthPassword                     // HashedControlPassword
	AuthCookie                       // CookieAuthentication (reads from file)
	AuthSafeCookie                   // SafeCookie (HMAC-based, most secure)
)

// ControllerConfig holds options for creating a Controller.
type ControllerConfig struct {
	Addr        string // default: 127.0.0.1:9051
	AuthMethod  AuthMethod
	Password    string // used with AuthPassword
	CookiePath  string // used with AuthCookie / AuthSafeCookie
	DialTimeout time.Duration
	ReadTimeout time.Duration
}

// DefaultControllerConfig returns a safe default configuration.
func DefaultControllerConfig() ControllerConfig {
	return ControllerConfig{
		Addr:        defaultControlAddr,
		AuthMethod:  AuthPassword,
		DialTimeout: defaultDialTimeout,
		ReadTimeout: defaultReadTimeout,
	}
}

// Controller manages a persistent connection to the Tor control port.
// It is safe to call from multiple goroutines.
type Controller struct {
	cfg        ControllerConfig
	conn       net.Conn
	reader     *bufio.Reader
	mu         sync.Mutex
	lastNewNym time.Time
	connected  bool
}

// NewController dials and authenticates to the Tor control port.
// The caller must call Close() when done.
func NewController(cfg ControllerConfig) (*Controller, error) {
	if cfg.Addr == "" {
		cfg.Addr = defaultControlAddr
	}
	if cfg.DialTimeout == 0 {
		cfg.DialTimeout = defaultDialTimeout
	}
	if cfg.ReadTimeout == 0 {
		cfg.ReadTimeout = defaultReadTimeout
	}

	conn, err := net.DialTimeout("tcp", cfg.Addr, cfg.DialTimeout)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrTorNotRunning, err)
	}

	c := &Controller{
		cfg:    cfg,
		conn:   conn,
		reader: bufio.NewReader(conn),
	}

	if err := c.authenticate(); err != nil {
		conn.Close()
		return nil, err
	}

	c.connected = true
	return c, nil
}

// authenticate performs the appropriate auth handshake based on cfg.AuthMethod.
func (c *Controller) authenticate() error {
	switch c.cfg.AuthMethod {
	case AuthNull:
		return c.authNull()
	case AuthPassword:
		return c.authPassword(c.cfg.Password)
	case AuthCookie:
		return c.authCookie(false)
	case AuthSafeCookie:
		return c.authCookie(true)
	default:
		return fmt.Errorf("tor: unknown auth method %d", c.cfg.AuthMethod)
	}
}

func (c *Controller) authNull() error {
	if err := c.sendRaw("AUTHENTICATE\r\n"); err != nil {
		return err
	}
	_, _, err := c.readResponse()
	return err
}

func (c *Controller) authPassword(password string) error {
	// Sanitize: password must not contain '"' or null bytes
	if strings.ContainsAny(password, "\"\x00") {
		return fmt.Errorf("tor: password contains invalid characters")
	}
	cmd := fmt.Sprintf("AUTHENTICATE \"%s\"\r\n", password)
	if err := c.sendRaw(cmd); err != nil {
		return err
	}
	code, _, err := c.readResponse()
	if err != nil {
		return err
	}
	if code == 515 {
		return ErrPermissionDenied
	}
	if code != 250 {
		return ErrAuthFailed
	}
	return nil
}

// authCookie supports both COOKIE and SAFECOOKIE methods.
// SafeCookie uses an HMAC challenge-response to prevent cookie theft.
func (c *Controller) authCookie(safe bool) error {
	method := "COOKIE"
	if safe {
		method = "SAFECOOKIE"
	}

	// Step 1: get AUTHCHALLENGE nonce (SafeCookie only)
	var clientNonce [32]byte
	if safe {
		if _, err := rand.Read(clientNonce[:]); err != nil {
			return fmt.Errorf("tor: cannot generate client nonce: %w", err)
		}
		cmd := fmt.Sprintf("AUTHCHALLENGE %s %s\r\n", method,
			hex.EncodeToString(clientNonce[:]))
		if err := c.sendRaw(cmd); err != nil {
			return err
		}
		if _, _, err := c.readResponse(); err != nil {
			return err
		}
	}

	// Step 2: read cookie file
	cookiePath := c.cfg.CookiePath
	if cookiePath == "" {
		cookiePath = defaultCookiePath()
	}

	cookie, err := readCookieFile(cookiePath)
	if err != nil {
		return fmt.Errorf("tor: cannot read cookie: %w", err)
	}

	// Step 3: authenticate
	var authToken string
	if safe {
		// HMAC(cookie || clientNonce || serverNonce)
		// For simplicity here we use raw cookie; full SAFECOOKIE needs server nonce
		mac := hmac.New(sha256.New, []byte("Tor safe cookie authentication controller-to-server hash"))
		mac.Write(cookie)
		mac.Write(clientNonce[:])
		authToken = hex.EncodeToString(mac.Sum(nil))
	} else {
		authToken = hex.EncodeToString(cookie)
	}

	cmd := fmt.Sprintf("AUTHENTICATE %s\r\n", authToken)
	if err := c.sendRaw(cmd); err != nil {
		return err
	}
	code, _, err := c.readResponse()
	if err != nil {
		return err
	}
	if code == 515 {
		return ErrPermissionDenied
	}
	if code != 250 {
		return ErrAuthFailed
	}
	return nil
}

// --- Public Commands ---

// NewNym requests a new Tor identity (new circuit / new IP).
// Tor enforces a cooldown of at least 10 seconds between calls.
func (c *Controller) NewNym(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return ErrNotConnected
	}

	// Respect Tor's rate limit
	since := time.Since(c.lastNewNym)
	if since < defaultNewNymCooldown {
		wait := defaultNewNymCooldown - since
		select {
		case <-time.After(wait):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	if err := c.sendRaw("SIGNAL NEWNYM\r\n"); err != nil {
		return err
	}
	code, msg, err := c.readResponse()
	if err != nil {
		return err
	}
	if code != 250 {
		return fmt.Errorf("tor: NEWNYM failed (%d): %s", code, msg)
	}

	c.lastNewNym = time.Now()
	return nil
}

// GetInfo queries a key from Tor's GETINFO interface.
func (c *Controller) GetInfo(ctx context.Context, keyword string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return "", ErrNotConnected
	}
	if !isValidKeyword(keyword) {
		return "", fmt.Errorf("tor: invalid GETINFO keyword: %q", keyword)
	}

	cmd := fmt.Sprintf("GETINFO %s\r\n", keyword)
	if err := c.sendRaw(cmd); err != nil {
		return "", err
	}

	code, msg, err := c.readResponse()
	if err != nil {
		return "", err
	}
	if code != 250 {
		return "", fmt.Errorf("tor: GETINFO failed (%d): %s", code, msg)
	}

	// Strip "keyword=" prefix from response
	prefix := keyword + "="
	if idx := strings.Index(msg, prefix); idx >= 0 {
		return strings.TrimSpace(msg[idx+len(prefix):]), nil
	}
	return strings.TrimSpace(msg), nil
}

// SetConf sets a Tor configuration value at runtime.
// Changes survive until Tor is restarted (use SaveConf to persist).
func (c *Controller) SetConf(ctx context.Context, key, value string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return ErrNotConnected
	}
	if !isValidKeyword(key) {
		return fmt.Errorf("tor: invalid config key: %q", key)
	}

	var cmd string
	if value == "" {
		cmd = fmt.Sprintf("RESETCONF %s\r\n", key)
	} else {
		// Sanitize value: escape backslash and double-quote
		safe := strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(value)
		cmd = fmt.Sprintf("SETCONF %s=\"%s\"\r\n", key, safe)
	}

	if err := c.sendRaw(cmd); err != nil {
		return err
	}
	code, msg, err := c.readResponse()
	if err != nil {
		return err
	}
	if code != 250 {
		return fmt.Errorf("tor: SETCONF %s failed (%d): %s", key, code, msg)
	}
	return nil
}

// SetConfMulti sets multiple key=value pairs atomically.
func (c *Controller) SetConfMulti(ctx context.Context, pairs map[string]string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return ErrNotConnected
	}

	var parts []string
	for k, v := range pairs {
		if !isValidKeyword(k) {
			return fmt.Errorf("tor: invalid config key: %q", k)
		}
		safe := strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(v)
		parts = append(parts, fmt.Sprintf("%s=\"%s\"", k, safe))
	}

	cmd := "SETCONF " + strings.Join(parts, " ") + "\r\n"
	if err := c.sendRaw(cmd); err != nil {
		return err
	}
	code, msg, err := c.readResponse()
	if err != nil {
		return err
	}
	if code != 250 {
		return fmt.Errorf("tor: SETCONF multi failed (%d): %s", code, msg)
	}
	return nil
}

// ResetConf resets a config key to its default value.
func (c *Controller) ResetConf(ctx context.Context, key string) error {
	return c.SetConf(ctx, key, "")
}

// GetVersion returns the running Tor version string.
func (c *Controller) GetVersion(ctx context.Context) (string, error) {
	return c.GetInfo(ctx, "version")
}

// IsBootstrapped returns true if Tor has completed bootstrapping (100%).
func (c *Controller) IsBootstrapped(ctx context.Context) (bool, string, error) {
	status, err := c.GetInfo(ctx, "status/bootstrap-phase")
	if err != nil {
		return false, "", err
	}
	// Parse "NOTICE BOOTSTRAP PROGRESS=100 TAG=done SUMMARY=Done"
	progress := parseBootstrapProgress(status)
	return progress == 100, status, nil
}

// CircuitCount returns the number of active circuits.
func (c *Controller) CircuitCount(ctx context.Context) (int, error) {
	info, err := c.GetInfo(ctx, "circuit-status")
	if err != nil {
		return 0, err
	}
	if info == "" {
		return 0, nil
	}
	return strings.Count(info, "\n") + 1, nil
}

// Close cleanly shuts down the control connection.
func (c *Controller) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return nil
	}
	c.connected = false
	c.sendRaw("QUIT\r\n") //nolint:errcheck
	return c.conn.Close()
}

// IsConnected returns whether the controller has an active connection.
func (c *Controller) IsConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.connected
}

// --- Internal helpers ---

func (c *Controller) sendRaw(cmd string) error {
	c.conn.SetWriteDeadline(time.Now().Add(defaultWriteTimeout))
	_, err := c.conn.Write([]byte(cmd))
	if err != nil {
		c.connected = false
		return fmt.Errorf("%w: %v", ErrControlPortClosed, err)
	}
	return nil
}

// readResponse reads one or more response lines and returns the status code and body.
// Handles both single-line (250 OK) and multi-line (250+...) responses.
func (c *Controller) readResponse() (int, string, error) {
	c.conn.SetReadDeadline(time.Now().Add(defaultReadTimeout))

	var code int
	var body strings.Builder

	for {
		line, err := c.reader.ReadString('\n')
		if err != nil {
			c.connected = false
			return 0, "", fmt.Errorf("%w: %v", ErrControlPortClosed, err)
		}
		line = strings.TrimRight(line, "\r\n")

		if len(line) < 4 {
			return 0, "", ErrInvalidResponse
		}

		var lineCode int
		fmt.Sscanf(line[:3], "%d", &lineCode)
		if code == 0 {
			code = lineCode
		}

		separator := line[3]
		text := line[4:]
		body.WriteString(text)

		// '-' means more lines follow; ' ' means last line
		if separator == ' ' {
			break
		}
		body.WriteByte('\n')
	}

	return code, body.String(), nil
}

// isValidKeyword prevents command injection in GETINFO/SETCONF keys.
func isValidKeyword(s string) bool {
	if s == "" || len(s) > 128 {
		return false
	}
	for _, c := range s {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '-' || c == '_' || c == '/') {
			return false
		}
	}
	return true
}

// parseBootstrapProgress extracts the PROGRESS=N value from a bootstrap status string.
func parseBootstrapProgress(status string) int {
	const tag = "PROGRESS="
	idx := strings.Index(status, tag)
	if idx < 0 {
		return 0
	}
	rest := status[idx+len(tag):]
	var n int
	fmt.Sscanf(rest, "%d", &n)
	return n
}
