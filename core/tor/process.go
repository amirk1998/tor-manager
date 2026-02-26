package tor

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

const (
	bootstrapPollInterval = 500 * time.Millisecond
	bootstrapTimeout      = 120 * time.Second
)

// ProcessState describes the current lifecycle state of the Tor process.
type ProcessState int

const (
	StateStopped     ProcessState = iota
	StateStarting
	StateBootstrapping
	StateRunning
	StateError
)

func (s ProcessState) String() string {
	switch s {
	case StateStopped:
		return "Stopped"
	case StateStarting:
		return "Starting"
	case StateBootstrapping:
		return "Bootstrapping"
	case StateRunning:
		return "Running"
	case StateError:
		return "Error"
	default:
		return "Unknown"
	}
}

// BootstrapEvent is sent on the BootstrapProgress channel during startup.
type BootstrapEvent struct {
	Progress int    // 0–100
	Tag      string // e.g. "conn", "handshake", "done"
	Summary  string // human-readable description
}

// Process manages the lifecycle of an embedded Tor daemon.
type Process struct {
	cfg        TorConfig
	configPath string
	cmd        *exec.Cmd
	mu         sync.RWMutex
	state      ProcessState
	logLines   []string // circular buffer of recent log lines
	logCap     int

	// BootstrapProgress emits events as Tor bootstraps. Closed when done.
	BootstrapProgress chan BootstrapEvent
}

// NewProcess creates a Process manager for the given TorConfig.
// configPath is where the torrc will be written.
func NewProcess(cfg TorConfig, configPath string) *Process {
	return &Process{
		cfg:               cfg,
		configPath:        configPath,
		state:             StateStopped,
		logCap:            200,
		BootstrapProgress: make(chan BootstrapEvent, 16),
	}
}

// Start writes the torrc, then spawns the Tor daemon.
// It returns after the process has started (not after bootstrapping).
// Use WaitBootstrap to block until Tor is fully connected.
func (p *Process) Start(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.state == StateRunning || p.state == StateStarting {
		return fmt.Errorf("tor: process already running")
	}

	// Write config
	if err := p.cfg.WriteToFile(p.configPath); err != nil {
		return err
	}

	torBin, err := findTorBinary()
	if err != nil {
		return err
	}

	p.cmd = exec.CommandContext(ctx, torBin, "-f", p.configPath)
	p.cmd.Env = append(os.Environ(), "TOR_SKIP_LAUNCH=0")

	// Capture stdout/stderr for log parsing
	stdout, err := p.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("tor: stdout pipe: %w", err)
	}
	p.cmd.Stderr = p.cmd.Stdout // merge

	if err := p.cmd.Start(); err != nil {
		p.state = StateError
		return fmt.Errorf("tor: start failed: %w", err)
	}

	p.state = StateStarting
	go p.watchLogs(bufio.NewScanner(stdout))

	return nil
}

// Stop sends SIGTERM to the Tor process and waits for it to exit.
func (p *Process) Stop() error {
	p.mu.Lock()
	cmd := p.cmd
	p.mu.Unlock()

	if cmd == nil || cmd.Process == nil {
		return nil
	}

	if err := cmd.Process.Signal(os.Interrupt); err != nil {
		// Fall back to Kill if Interrupt isn't supported (Windows)
		cmd.Process.Kill() //nolint:errcheck
	}

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		cmd.Process.Kill() //nolint:errcheck
		<-done
	}

	p.mu.Lock()
	p.state = StateStopped
	p.cmd = nil
	p.mu.Unlock()

	return nil
}

// WaitBootstrap blocks until Tor completes bootstrapping (progress=100)
// or the context is cancelled.
func (p *Process) WaitBootstrap(ctx context.Context) error {
	deadline := time.Now().Add(bootstrapTimeout)
	for {
		if time.Now().After(deadline) {
			return ErrOperationTimeout
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(bootstrapPollInterval):
		}

		p.mu.RLock()
		state := p.state
		p.mu.RUnlock()

		if state == StateRunning {
			return nil
		}
		if state == StateError {
			return fmt.Errorf("tor: process entered error state during bootstrap")
		}
		if state == StateStopped {
			return fmt.Errorf("tor: process stopped before completing bootstrap")
		}
	}
}

// State returns the current process state (thread-safe).
func (p *Process) State() ProcessState {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.state
}

// RecentLogs returns the last N log lines captured from Tor's output.
func (p *Process) RecentLogs(n int) []string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if n >= len(p.logLines) {
		result := make([]string, len(p.logLines))
		copy(result, p.logLines)
		return result
	}
	result := make([]string, n)
	copy(result, p.logLines[len(p.logLines)-n:])
	return result
}

// watchLogs reads Tor's stdout line by line, updates state, and emits bootstrap events.
func (p *Process) watchLogs(scanner *bufio.Scanner) {
	defer close(p.BootstrapProgress)

	for scanner.Scan() {
		line := scanner.Text()
		p.appendLog(line)

		// Parse bootstrap progress lines:
		// Nov 01 12:00:00.000 [notice] Bootstrapped 25% (conn): Connecting to a relay
		if strings.Contains(line, "Bootstrapped") {
			evt := parseBootstrapEvent(line)
			p.mu.Lock()
			if evt.Progress < 100 {
				p.state = StateBootstrapping
			} else {
				p.state = StateRunning
			}
			p.mu.Unlock()

			select {
			case p.BootstrapProgress <- evt:
			default: // non-blocking; consumer may be slow
			}
		}

		if strings.Contains(line, "[err]") || strings.Contains(line, "[warn] Failed to") {
			p.mu.Lock()
			p.state = StateError
			p.mu.Unlock()
		}
	}
}

func (p *Process) appendLog(line string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.logLines) >= p.logCap {
		// Shift out oldest
		p.logLines = p.logLines[1:]
	}
	p.logLines = append(p.logLines, line)
}

// parseBootstrapEvent extracts progress/tag/summary from a Tor log line.
func parseBootstrapEvent(line string) BootstrapEvent {
	evt := BootstrapEvent{}

	// Extract PROGRESS
	idx := strings.Index(line, "Bootstrapped ")
	if idx < 0 {
		return evt
	}
	rest := line[idx+len("Bootstrapped "):]
	fmt.Sscanf(rest, "%d", &evt.Progress)

	// Extract TAG and SUMMARY from structured log:
	// "Bootstrapped 50% (loading_descriptors): Loading relay descriptors"
	if paren := strings.Index(rest, "("); paren >= 0 {
		end := strings.Index(rest[paren:], ")")
		if end >= 0 {
			evt.Tag = rest[paren+1 : paren+end]
		}
	}
	if colon := strings.Index(rest, "): "); colon >= 0 {
		evt.Summary = strings.TrimSpace(rest[colon+3:])
	}

	return evt
}

// findTorBinary locates the tor executable on the current platform.
func findTorBinary() (string, error) {
	// 1. Check PATH first (handles system-installed Tor)
	if path, err := exec.LookPath("tor"); err == nil {
		return path, nil
	}

	// 2. Platform-specific fallback locations
	var candidates []string
	switch runtime.GOOS {
	case "linux":
		candidates = []string{"/usr/bin/tor", "/usr/sbin/tor", "/usr/local/bin/tor"}
	case "darwin":
		candidates = []string{
			"/usr/local/bin/tor",                                    // homebrew (Intel)
			"/opt/homebrew/bin/tor",                                  // homebrew (Apple Silicon)
			"/Applications/Tor Browser.app/Contents/MacOS/Tor/tor",  // Tor Browser bundle
		}
	case "windows":
		candidates = []string{
			filepath.Join(os.Getenv("ProgramFiles"), "Tor", "tor.exe"),
			filepath.Join(os.Getenv("ProgramFiles(x86)"), "Tor", "tor.exe"),
			filepath.Join(os.Getenv("APPDATA"), "Tor Browser", "Browser", "TorBrowser", "Tor", "tor.exe"),
		}
	}

	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}

	return "", fmt.Errorf("tor: cannot find tor binary — install Tor and ensure it is in PATH")
}

// HashPassword runs `tor --hash-password` and returns the hashed value.
// Useful for generating HashedControlPassword for the torrc.
func HashPassword(password string) (string, error) {
	if strings.ContainsAny(password, "\"\x00\r\n") {
		return "", fmt.Errorf("tor: password contains invalid characters")
	}

	torBin, err := findTorBinary()
	if err != nil {
		return "", err
	}

	out, err := exec.Command(torBin, "--hash-password", password).Output()
	if err != nil {
		return "", fmt.Errorf("tor: hash-password failed: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "16:") {
			return line, nil
		}
	}
	return "", fmt.Errorf("tor: unexpected hash-password output: %q", string(out))
}
