// Package config handles persistent application-level settings for tor-manager.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

const appName = "tor-manager"

// AppConfig stores user preferences that persist between sessions.
type AppConfig struct {
	// Tor daemon settings
	SocksPort   int    `json:"socks_port"`
	ControlPort int    `json:"control_port"`
	ControlAuth string `json:"control_auth"` // "password" | "cookie"
	Password    string `json:"password"`     // stored only if auth=password

	// Bridge settings
	UseBridges       bool   `json:"use_bridges"`
	BridgeTransport  string `json:"bridge_transport"`  // transport type string
	CustomBridges    []string `json:"custom_bridges"` // user-added bridge lines

	// Exit node preferences
	ExitCountries   []string `json:"exit_countries"`
	StrictExitNodes bool     `json:"strict_exit_nodes"`

	// UI preferences
	AutoRefreshSecs int  `json:"auto_refresh_secs"`
	ShowBootstrapLog bool `json:"show_bootstrap_log"`

	mu sync.Mutex `json:"-"`
}

// Default returns a sensible default AppConfig.
func Default() *AppConfig {
	return &AppConfig{
		SocksPort:        9050,
		ControlPort:      9051,
		ControlAuth:      "cookie",
		UseBridges:       false,
		BridgeTransport:  "obfs4",
		AutoRefreshSecs:  30,
		ShowBootstrapLog: true,
	}
}

// ConfigDir returns the platform-appropriate config directory.
func ConfigDir() (string, error) {
	var base string
	switch runtime.GOOS {
	case "windows":
		base = os.Getenv("APPDATA")
	case "darwin":
		base = filepath.Join(os.Getenv("HOME"), "Library", "Application Support")
	default: // Linux / BSD / etc.
		if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
			base = xdg
		} else {
			base = filepath.Join(os.Getenv("HOME"), ".config")
		}
	}
	if base == "" {
		return "", fmt.Errorf("config: cannot determine config directory")
	}
	return filepath.Join(base, appName), nil
}

// Load reads the AppConfig from disk. Returns Default() if the file doesn't exist.
func Load() (*AppConfig, error) {
	dir, err := ConfigDir()
	if err != nil {
		return Default(), nil
	}

	path := filepath.Join(dir, "config.json")
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return Default(), nil
	}
	if err != nil {
		return nil, fmt.Errorf("config: read %s: %w", path, err)
	}

	cfg := Default()
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("config: parse %s: %w", path, err)
	}
	return cfg, nil
}

// Save writes the config to disk atomically.
func (c *AppConfig) Save() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	dir, err := ConfigDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("config: create dir: %w", err)
	}

	path := filepath.Join(dir, "config.json")
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("config: marshal: %w", err)
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return fmt.Errorf("config: write: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp) //nolint:errcheck
		return fmt.Errorf("config: rename: %w", err)
	}
	return nil
}

// Validate checks the config for obvious errors.
func (c *AppConfig) Validate() error {
	if c.SocksPort < 1 || c.SocksPort > 65535 {
		return fmt.Errorf("config: invalid SocksPort %d", c.SocksPort)
	}
	if c.ControlPort < 1 || c.ControlPort > 65535 {
		return fmt.Errorf("config: invalid ControlPort %d", c.ControlPort)
	}
	if c.SocksPort == c.ControlPort {
		return fmt.Errorf("config: SocksPort and ControlPort must differ")
	}
	if c.AutoRefreshSecs < 5 {
		return fmt.Errorf("config: AutoRefreshSecs must be >= 5")
	}
	return nil
}
