package tor

import (
	"strings"
	"testing"
)

func TestParseBridgeLine(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantTransport TransportType
		wantAddr    string
		wantFP      string
		wantErr     bool
	}{
		{
			name:          "obfs4 full line",
			input:         "obfs4 1.2.3.4:1234 A1B2C3D4E5F6A1B2C3D4E5F6A1B2C3D4E5F6A1B2 cert=abc iat-mode=0",
			wantTransport: TransportObfs4,
			wantAddr:      "1.2.3.4:1234",
			wantFP:        "A1B2C3D4E5F6A1B2C3D4E5F6A1B2C3D4E5F6A1B2",
		},
		{
			name:          "obfs4 with Bridge prefix",
			input:         "Bridge obfs4 1.2.3.4:9001 A1B2C3D4E5F6A1B2C3D4E5F6A1B2C3D4E5F6A1B2",
			wantTransport: TransportObfs4,
			wantAddr:      "1.2.3.4:9001",
			wantFP:        "A1B2C3D4E5F6A1B2C3D4E5F6A1B2C3D4E5F6A1B2",
		},
		{
			name:          "vanilla bridge bare IP",
			input:         "1.2.3.4:443",
			wantTransport: TransportDirect,
			wantAddr:      "1.2.3.4:443",
		},
		{
			name:          "vanilla bridge with fingerprint",
			input:         "1.2.3.4:443 A1B2C3D4E5F6A1B2C3D4E5F6A1B2C3D4E5F6A1B2",
			wantTransport: TransportDirect,
			wantAddr:      "1.2.3.4:443",
			wantFP:        "A1B2C3D4E5F6A1B2C3D4E5F6A1B2C3D4E5F6A1B2",
		},
		{
			name:          "meek-azure alias",
			input:         "meek-azure 0.0.2.0:2 B9E7141C594AF25699E0079C1F0146F409495296",
			wantTransport: TransportMeekAzure,
			wantAddr:      "0.0.2.0:2",
		},
		{
			name:          "snowflake",
			input:         "snowflake 192.0.2.3:1 2B280B23E1107BB62ABFC40DDCC8824814F80A72 url=https://example.com",
			wantTransport: TransportSnowflake,
			wantAddr:      "192.0.2.3:1",
		},
		{
			name:    "empty line",
			input:   "",
			wantErr: true,
		},
		{
			name:    "transport without address",
			input:   "obfs4",
			wantErr: true,
		},
		{
			name:    "invalid fingerprint length",
			input:   "1.2.3.4:443 TOOSHORT",
			wantTransport: TransportDirect,
			wantAddr: "1.2.3.4:443",
			// Short hex string that doesn't match fingerprint pattern is treated as non-fingerprint
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := ParseBridgeLine(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil (bridge=%+v)", b)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if b.Transport != tt.wantTransport {
				t.Errorf("transport: got %q, want %q", b.Transport, tt.wantTransport)
			}
			if b.Address != tt.wantAddr {
				t.Errorf("address: got %q, want %q", b.Address, tt.wantAddr)
			}
			if tt.wantFP != "" && b.Fingerprint != tt.wantFP {
				t.Errorf("fingerprint: got %q, want %q", b.Fingerprint, tt.wantFP)
			}
		})
	}
}

func TestBridgeIsValid(t *testing.T) {
	tests := []struct {
		name    string
		bridge  Bridge
		wantErr string
	}{
		{
			name:   "valid obfs4",
			bridge: Bridge{Transport: TransportObfs4, Address: "1.2.3.4:1234"},
		},
		{
			name:    "missing address",
			bridge:  Bridge{Transport: TransportObfs4},
			wantErr: "missing address",
		},
		{
			name:    "bad fingerprint",
			bridge:  Bridge{Transport: TransportDirect, Address: "1.2.3.4:443", Fingerprint: "ZZZZ"},
			wantErr: "fingerprint",
		},
		{
			name:    "bad address format",
			bridge:  Bridge{Transport: TransportDirect, Address: "notanaddress"},
			wantErr: "bad address",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.bridge.IsValid()
			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.wantErr)
				} else if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.wantErr)
				}
			}
		})
	}
}

func TestBuiltInBridgeEntries(t *testing.T) {
	for _, tt := range AllBuiltInTransports() {
		entries := BuiltInBridgeEntries(tt)
		if len(entries) == 0 {
			t.Errorf("transport %q has no built-in bridges", tt)
			continue
		}
		for _, b := range entries {
			if err := b.IsValid(); err != nil {
				t.Errorf("built-in bridge for %q is invalid: %v (bridge: %s)", tt, err, b.String())
			}
			if b.Transport != tt {
				t.Errorf("built-in bridge transport mismatch: got %q, want %q", b.Transport, tt)
			}
		}
	}
}

func TestIsValidKeyword(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"UseBridges", true},
		{"circuit-status", true},
		{"status/bootstrap-phase", true},
		{"", false},
		{"key with spaces", false},
		{"key;injection", false},
		{strings.Repeat("a", 129), false},
	}
	for _, tt := range tests {
		if got := isValidKeyword(tt.input); got != tt.valid {
			t.Errorf("isValidKeyword(%q) = %v, want %v", tt.input, got, tt.valid)
		}
	}
}

func TestGenerateTorrc(t *testing.T) {
	cfg := TorConfig{
		SocksPort:             9050,
		ControlPort:           9051,
		DataDir:               "/tmp/tor-test",
		CookieAuthentication:  true,
		SafeSocks:             true,
		ExitNodes:             []string{"US", "DE"},
		StrictExitNodes:       true,
	}

	rc := cfg.GenerateTorrc()

	checks := []string{
		"SocksPort 9050",
		"ControlPort 9051",
		"DataDirectory /tmp/tor-test",
		"CookieAuthentication 1",
		"SafeSocks 1",
		"ExitNodes {US},{DE}",
		"StrictNodes 1",
	}

	for _, check := range checks {
		if !strings.Contains(rc, check) {
			t.Errorf("torrc missing %q\n\nGot:\n%s", check, rc)
		}
	}
}
