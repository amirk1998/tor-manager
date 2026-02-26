# tor-manager/core

Core library for managing a Tor daemon, configuring bridges, and verifying connectivity.
Used by both `tor-tui` and `tor-gui` applications.

## Package Structure

```
core/
├── tor/
│   ├── errors.go          — sentinel error types
│   ├── controller.go      — Control Port connection & auth (password / cookie / safecookie)
│   ├── cookie.go          — platform-aware cookie file helpers
│   ├── bridge.go          — bridge types, parsing, validation, applying to live Tor
│   ├── builtin_bridges.go — curated built-in bridges (obfs4, snowflake, meek-azure, webtunnel)
│   ├── bridgedb.go        — fetch fresh bridges from Tor Project's BridgeDB MOAT API
│   ├── config.go          — torrc generation with security hardening
│   └── process.go         — Tor daemon lifecycle (start/stop/bootstrap tracking)
├── proxy/
│   └── checker.go         — IP check & geolocation through SOCKS5
└── config/
    └── appconfig.go       — persistent user preferences (JSON, XDG-aware)
```

## Quick Start

```go
import (
    "context"
    "fmt"
    "github.com/you/tor-manager/core/tor"
    "github.com/you/tor-manager/core/proxy"
)

func main() {
    ctx := context.Background()

    // 1. Generate torrc
    cfg := tor.DefaultConfig()
    cfg.HashedControlPassword = "" // use cookie auth (default)
    if err := cfg.WriteToFile("/tmp/torrc"); err != nil {
        panic(err)
    }

    // 2. Start Tor
    proc := tor.NewProcess(cfg, "/tmp/torrc")
    if err := proc.Start(ctx); err != nil {
        panic(err)
    }
    defer proc.Stop()

    // 3. Wait for bootstrap
    if err := proc.WaitBootstrap(ctx); err != nil {
        panic(err)
    }

    // 4. Connect to control port
    ctrl, err := tor.NewController(tor.DefaultControllerConfig())
    if err != nil {
        panic(err)
    }
    defer ctrl.Close()

    // 5. Apply built-in obfs4 bridges
    bridges := tor.BuiltInBridgeEntries(tor.TransportObfs4)
    bridgeCfg := tor.BridgeConfig{
        Enabled:   true,
        Bridges:   bridges,
        Transport: tor.TransportObfs4,
    }
    if err := bridgeCfg.Apply(ctx, ctrl); err != nil {
        panic(err)
    }

    // 6. Request new identity
    if err := ctrl.NewNym(ctx); err != nil {
        panic(err)
    }

    // 7. Check current IP
    checker, _ := proxy.NewChecker(cfg.SocksPort)
    result := checker.CheckIP(ctx)
    if result.Error != nil {
        panic(result.Error)
    }
    fmt.Println(proxy.FormatIPInfo(result.IPInfo))
    // Output: 185.220.x.x | Frankfurt, DE [Tor exit]
}
```

## Requesting Fresh Bridges from BridgeDB

```go
bridges, err := tor.RequestBridgesFromBridgeDB(ctx, tor.BridgeDBOptions{
    Transport: tor.TransportObfs4,
    SocksPort: 9050, // route through Tor for privacy (set 0 if Tor not running yet)
    Count:     3,
})
```

## Parsing Custom Bridge Lines

```go
line := "obfs4 1.2.3.4:1234 FINGERPRINT cert=abc iat-mode=0"
bridge, err := tor.ParseBridgeLine(line)
if err != nil {
    // invalid bridge
}
```

## Security Notes

- torrc is written with **mode 0600** (owner-readable only)
- Control port password sanitization prevents **command injection**
- GETINFO/SETCONF keys are **allowlisted** (alphanumeric + `-_/` only)
- HTTP responses are **size-limited** (64 KB) to prevent memory exhaustion
- BridgeDB requests are routed **through Tor** when available
- SafeSocks=1 is enabled by default to **prevent DNS leaks**
- Stream isolation (IsolateClientAddr) is enabled by default

## Prerequisites

- **Tor** installed and in PATH (`apt install tor` / `brew install tor`)
- For obfs4/meek: `obfs4proxy` (`apt install obfs4proxy`)
- For snowflake: `snowflake-client` (from Tor Project)

## Running Tests

```bash
cd core
go test ./tor/... -v
```
