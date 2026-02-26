package tor

import (
	"context"
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"
)

// TransportType defines the pluggable transport protocol.
type TransportType string

const (
	TransportDirect       TransportType = ""
	TransportObfs4        TransportType = "obfs4"
	TransportMeekAzure    TransportType = "meek_lite"
	TransportSnowflake    TransportType = "snowflake"
	TransportWebTunnel    TransportType = "webtunnel"
	TransportScrambleSuit TransportType = "scramblesuit"
)

// TransportMeta holds human-readable info about a transport type.
type TransportMeta struct {
	Name                 string
	Description          string
	CensorshipResistance int // 1 (low) – 5 (very high)
	SpeedImpact          int // 1 (minimal) – 5 (very slow)
	HighCensorshipOK     bool
}

// Transports provides metadata about each known pluggable transport.
var Transports = map[TransportType]TransportMeta{
	TransportDirect: {
		Name:                 "Direct (No Bridge)",
		Description:          "Connect to Tor directly without any bridge.",
		CensorshipResistance: 1,
		SpeedImpact:          1,
		HighCensorshipOK:     false,
	},
	TransportObfs4: {
		Name:                 "obfs4",
		Description:          "Obfuscates traffic to look like random noise. Fast and widely available.",
		CensorshipResistance: 4,
		SpeedImpact:          2,
		HighCensorshipOK:     true,
	},
	TransportSnowflake: {
		Name:                 "Snowflake",
		Description:          "Uses WebRTC via volunteer proxies. Proxies rotate constantly — very hard to block.",
		CensorshipResistance: 5,
		SpeedImpact:          3,
		HighCensorshipOK:     true,
	},
	TransportMeekAzure: {
		Name:                 "meek-azure",
		Description:          "Domain-fronts traffic through Microsoft Azure CDN. Extremely hard to block.",
		CensorshipResistance: 5,
		SpeedImpact:          5,
		HighCensorshipOK:     true,
	},
	TransportWebTunnel: {
		Name:                 "WebTunnel",
		Description:          "Makes traffic indistinguishable from regular HTTPS. Newest and most stealthy.",
		CensorshipResistance: 5,
		SpeedImpact:          2,
		HighCensorshipOK:     true,
	},
}

// Bridge represents a single Tor bridge entry.
type Bridge struct {
	Transport   TransportType
	Address     string // host:port
	Fingerprint string // 40-char hex SHA-1 (optional for some transports)
	Options     string // extra params like cert=... iat-mode=...
	Raw         string // original unparsed line (preserved for display)
}

// String formats the bridge as a torrc-compatible line.
func (b Bridge) String() string {
	if b.Raw != "" {
		return b.Raw
	}
	var parts []string
	if b.Transport != TransportDirect {
		parts = append(parts, string(b.Transport))
	}
	parts = append(parts, b.Address)
	if b.Fingerprint != "" {
		parts = append(parts, b.Fingerprint)
	}
	if b.Options != "" {
		parts = append(parts, b.Options)
	}
	return strings.Join(parts, " ")
}

// IsValid performs structural validation on the bridge.
func (b Bridge) IsValid() error {
	if b.Address == "" {
		return fmt.Errorf("%w: missing address", ErrBridgeInvalid)
	}
	host, port, err := net.SplitHostPort(b.Address)
	if err != nil {
		return fmt.Errorf("%w: bad address %q: %v", ErrBridgeInvalid, b.Address, err)
	}
	if host == "" || port == "" {
		return fmt.Errorf("%w: address %q has empty host or port", ErrBridgeInvalid, b.Address)
	}
	if b.Fingerprint != "" && !isValidFingerprint(b.Fingerprint) {
		return fmt.Errorf("%w: fingerprint must be 40 hex chars, got %q", ErrBridgeInvalid, b.Fingerprint)
	}
	return nil
}

var fingerprintRE = regexp.MustCompile(`^[0-9A-Fa-f]{40}$`)

func isValidFingerprint(fp string) bool {
	return fingerprintRE.MatchString(fp)
}

// ParseBridgeLine parses a raw bridge line into a Bridge struct.
// Accepts lines with or without a leading "Bridge " prefix.
func ParseBridgeLine(line string) (Bridge, error) {
	line = strings.TrimSpace(line)
	if strings.HasPrefix(strings.ToUpper(line), "BRIDGE ") {
		line = line[7:]
	}
	if line == "" {
		return Bridge{}, fmt.Errorf("%w: empty line", ErrBridgeInvalid)
	}

	b := Bridge{Raw: line}
	parts := strings.Fields(line)

	knownTransports := map[string]TransportType{
		"obfs4":        TransportObfs4,
		"meek_lite":    TransportMeekAzure,
		"meek-azure":   TransportMeekAzure,
		"snowflake":    TransportSnowflake,
		"webtunnel":    TransportWebTunnel,
		"scramblesuit": TransportScrambleSuit,
	}

	if tt, ok := knownTransports[strings.ToLower(parts[0])]; ok {
		if len(parts) < 2 {
			return Bridge{}, fmt.Errorf("%w: transport %q missing address", ErrBridgeInvalid, parts[0])
		}
		b.Transport = tt
		b.Address = parts[1]
		if len(parts) > 2 && isValidFingerprint(parts[2]) {
			b.Fingerprint = strings.ToUpper(parts[2])
			b.Options = strings.Join(parts[3:], " ")
		} else if len(parts) > 2 {
			b.Options = strings.Join(parts[2:], " ")
		}
	} else {
		// Vanilla bridge — first token is the address
		b.Transport = TransportDirect
		b.Address = parts[0]
		if len(parts) > 1 && isValidFingerprint(parts[1]) {
			b.Fingerprint = strings.ToUpper(parts[1])
		}
	}

	if err := b.IsValid(); err != nil {
		return Bridge{}, err
	}
	return b, nil
}

// BridgeConfig describes a bridge configuration to apply to a live Tor instance.
type BridgeConfig struct {
	Enabled   bool
	Bridges   []Bridge
	Transport TransportType
	PTPath    string // override default pluggable transport binary path
}

// Apply pushes the bridge configuration to a running Tor controller.
func (bc BridgeConfig) Apply(ctx context.Context, c *Controller) error {
	if !bc.Enabled || len(bc.Bridges) == 0 {
		return c.SetConf(ctx, "UseBridges", "0")
	}

	if ptPlugin := bc.buildPTPlugin(); ptPlugin != "" {
		if err := c.SetConf(ctx, "ClientTransportPlugin", ptPlugin); err != nil {
			return fmt.Errorf("tor: set transport plugin: %w", err)
		}
	}

	if err := c.SetConf(ctx, "UseBridges", "1"); err != nil {
		return fmt.Errorf("tor: enable bridge mode: %w", err)
	}

	// Acquire lock once for all bridge SETCONF calls
	c.mu.Lock()
	defer c.mu.Unlock()

	replacer := strings.NewReplacer(`\`, `\\`, `"`, `\"`)
	for _, br := range bc.Bridges {
		if err := br.IsValid(); err != nil {
			return err
		}
		safe := replacer.Replace(br.String())
		cmd := fmt.Sprintf("SETCONF Bridge=\"%s\"\r\n", safe)
		if err := c.sendRaw(cmd); err != nil {
			return err
		}
		code, msg, err := c.readResponse()
		if err != nil {
			return err
		}
		if code != 250 {
			return fmt.Errorf("tor: SETCONF Bridge failed (%d): %s", code, msg)
		}
	}
	return nil
}

func (bc BridgeConfig) buildPTPlugin() string {
	bin := bc.PTPath
	if bin == "" {
		bin = defaultPTBinary(bc.Transport)
	}
	if bin == "" {
		return ""
	}
	switch bc.Transport {
	case TransportObfs4:
		return fmt.Sprintf("obfs4 exec %s", bin)
	case TransportMeekAzure:
		return fmt.Sprintf("meek_lite exec %s", bin)
	case TransportSnowflake:
		return fmt.Sprintf(
			"snowflake exec %s -url https://snowflake-broker.torproject.org/ -front cdn.sstatic.net -ice stun:stun.l.google.com:19302",
			bin,
		)
	case TransportWebTunnel:
		return fmt.Sprintf("webtunnel exec %s", bin)
	}
	return ""
}

func defaultPTBinary(tt TransportType) string {
	var candidates []string
	switch tt {
	case TransportObfs4, TransportMeekAzure:
		candidates = []string{"/usr/bin/obfs4proxy", "/usr/local/bin/obfs4proxy"}
	case TransportSnowflake:
		candidates = []string{"/usr/bin/snowflake-client", "/usr/local/bin/snowflake-client"}
	case TransportWebTunnel:
		candidates = []string{"/usr/bin/webtunnel", "/usr/local/bin/webtunnel"}
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}
