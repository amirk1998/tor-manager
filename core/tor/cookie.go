package tor

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// defaultCookiePath returns the platform-specific default location
// for Tor's control auth cookie.
func defaultCookiePath() string {
	switch runtime.GOOS {
	case "linux":
		// System Tor (apt/dnf) vs user Tor
		paths := []string{
			"/run/tor/control.authcookie",
			filepath.Join(os.Getenv("HOME"), ".local/share/tor/control.authcookie"),
		}
		for _, p := range paths {
			if _, err := os.Stat(p); err == nil {
				return p
			}
		}
		return paths[0]
	case "darwin":
		return filepath.Join(os.Getenv("HOME"), "Library/Application Support/TorBrowser-Data/Tor/control_auth_cookie")
	case "windows":
		return filepath.Join(os.Getenv("APPDATA"), "Tor Browser", "Browser", "TorBrowser", "Data", "Tor", "control_auth_cookie")
	default:
		return "/var/lib/tor/control_auth_cookie"
	}
}

// readCookieFile reads and validates a Tor auth cookie file.
// Cookie files must be exactly 32 bytes.
func readCookieFile(path string) ([]byte, error) {
	// Prevent path traversal
	clean := filepath.Clean(path)
	if clean != path {
		return nil, fmt.Errorf("tor: suspicious cookie path: %q", path)
	}

	data, err := os.ReadFile(clean)
	if err != nil {
		return nil, err
	}
	if len(data) != 32 {
		return nil, fmt.Errorf("tor: cookie file is %d bytes, expected 32", len(data))
	}
	return data, nil
}
