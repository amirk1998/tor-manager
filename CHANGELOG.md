# Changelog

All notable changes to **tor-manager** will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/)
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [Unreleased]

> Changes that are merged into `dev` but not yet released.

### Planned
- Embedded Tor daemon (no separate installation required)
- Windows installer via NSIS (`wails build -nsis`)
- Tor circuit visualizer with relay map
- System tray integration for GUI
- Auto-update for built-in bridge list

---

## [0.1.1] — 2026-02-26

> Bug fix release for circular import and controller authentication issues.

### Fixed

#### `core` — Bug Fixes
- **Controller authentication** (`tor/controller.go`)
  - Fixed `authNull()` function to correctly handle 3-value return from `readResponse()`
  - Previously only captured 2 values, causing compilation error

#### `tor-tui` — Architecture Refactoring
- **Fixed circular import** — Resolved import cycle between `ui` and `ui/views` packages
  - Created new `ui/common` subpackage for shared types
  - Moved message types (`TickMsg`, `BootstrapEventMsg`, `LogLineMsg`, etc.) to `common/messages.go`
  - Moved styles and helper functions to `common/styles.go`
  - Moved key bindings to `common/keys.go`
  - Updated all view files to import from `ui/common` instead of `ui`
- **Type compatibility fixes**
  - Updated `CensorshipBadge()` and `SpeedBadge()` to accept `int` parameters (matching `TransportMeta` struct)
  - Fixed color constant usage with proper `lipgloss.Color()` wrapper throughout views
- **Build verification**
  - All packages now compile without errors
  - `go vet ./...` passes with no warnings

---

## [0.1.0] — 2026-02-26

> Initial public release of the tor-manager suite.

### Added

#### `core` — Shared Library
- **Controller** (`tor/controller.go`) — Thread-safe Tor Control Port client with mutex protection
    - Multiple authentication methods: Null, Password, Cookie, SafeCookie (HMAC-SHA256)
    - Multi-line response parser for `250+` and `250-` formats
    - Command injection prevention via keyword allowlist (alphanumeric + `-_/`, max 128 chars)
    - Configurable timeouts: dial (5s), read (10s), write (5s)
    - `NEWNYM` with enforced 10-second cooldown
    - Bootstrap progress tracking via `GETINFO status/bootstrap-phase`
    - Context-aware operations throughout
- **Bridge management** (`tor/bridge.go`)
    - `TransportType` enum: `obfs4`, `snowflake`, `meek-azure`, `webtunnel`, `scramblesuit`, direct
    - `Bridge` struct with full validation (address host:port, 40-char hex fingerprint, options)
    - `ParseBridgeLine`: handles `Bridge` prefix, transport detection, fingerprint extraction
    - `BridgeConfig.Apply`: atomic `SETCONF` for `UseBridges`, `ClientTransportPlugin`, `Bridge` lines
    - Pluggable transport binary detection across common install paths
    - Transport metadata: censorship resistance (1–5), speed impact (1–5)
- **Built-in bridge library** (`tor/builtin_bridges.go`)
    - 6 obfs4 bridges, 2 Snowflake, 1 meek-azure, 1 WebTunnel
    - `BuiltInBridgeEntries(transport)` and `AllBuiltInTransports()` helpers
- **BridgeDB integration** (`tor/bridgedb.go`)
    - MOAT API client: POST to `https://bridges.torproject.org/moat/fetch`
    - Requests routed through Tor SOCKS5 when available (privacy)
    - Response size capped at 64 KB (memory exhaustion prevention)
    - Returns 1–10 bridges per request (configurable, default 3)
- **Torrc generation** (`tor/config.go`)
    - `TorConfig` struct covering all common options
    - Security defaults: `SafeSocks=1`, `IsolateClientAddr=1`, `IsolateClientProtocol=1`, `TestSocks=1`
    - Atomic write: temp file + rename, mode `0600`
    - Validation: port ranges, duplicate ports, auth requirement
- **Process management** (`tor/process.go`)
    - `ProcessState` enum: Stopped, Starting, Bootstrapping, Running, Error
    - Bootstrap log parsing from stdout/stderr
    - `BootstrapProgress` channel emitting structured events
    - `WaitBootstrap()` with 120-second timeout
    - Graceful stop: SIGTERM + 10s grace + SIGKILL fallback
    - Circular log buffer (last 200 lines)
    - Platform-aware Tor binary detection
- **Cookie auth** (`tor/cookie.go`) — Platform-aware cookie file reader with path traversal prevention
- **IP checking** (`proxy/checker.go`)
    - Two-step verification: `check.torproject.org/api/ip` + `ipinfo.io/json`
    - `DisableKeepAlives` for privacy (new TCP connection per request)
    - Response size capped at 8 KB
    - Latency measurement included in result
- **App config** (`config/appconfig.go`)
    - Platform-aware config directory (XDG / AppData / Library)
    - JSON persistence with atomic write
    - Validation: port ranges, duplicate ports, `AutoRefreshSecs ≥ 5`

#### `tor-tui` — Terminal UI
- Full-terminal alt-screen application with Bubble Tea (Elm Architecture)
- Five tabs: **Dashboard**, **Bridges**, **Countries**, **Logs**, **Settings**
- Centralized design system (`ui/styles.go`) — "Noir Onion" palette (`#7C3AED` violet base)
- Centralized keybindings (`ui/keys.go`) — all bindings documented in one place
- Typed Tea message bus (`ui/messages.go`) — zero shared state between views
- **Dashboard view**
    - Real-time exit IP with country, city, org, latency
    - `● Connected` / `◉ Bootstrapping N%` / `○ Disconnected` status dot
    - Bootstrap progress bar
    - NEWNYM with enforced cooldown countdown
    - Proxy address display + clipboard copy
    - Auto-refresh on configurable interval
- **Bridges view**
    - Transport type tabs with censorship/speed badges
    - Built-in bridge list with cursor navigation
    - Custom bridge paste input with validation
    - BridgeDB fetch with loading spinner
    - Live apply to running daemon
- **Countries view**
    - Two-column layout: country list + active filter panel
    - Flag emoji display for 19 countries
    - Multi-select with toggle
    - StrictNodes toggle with anonymity warning
    - Live apply to running daemon
- **Logs view**
    - Viewport-based scrollable log with 500-line circular buffer
    - Follow mode (auto-scroll to bottom)
    - Syntax coloring: bootstrap events, circuit builds, errors, warnings
    - `pgup`/`pgdn`, `g`/`G` (top/bottom), clear
- **Settings view**
    - Four-field form: SOCKS5 port, control port, password, auto-refresh
    - Cookie vs password auth selector
    - Live validation with inline error messages
    - `ctrl+s` to save, `esc` to revert
- **Help overlay** — full keyboard reference accessible with `?`
- **Boot screen** with ASCII logo and spinner
- **Error screen** with actionable hint when control port is unreachable

#### `tor-gui` — Desktop GUI
- Wails v2 desktop application (Go backend + Svelte frontend, ~15 MB binary)
- No CGO, no Electron, uses native OS WebView (WebView2 on Windows, WebKit on Linux/macOS)
- Svelte 4 + Tailwind CSS 3 + Vite frontend stack
- "Noir Onion" dark theme with `DM Sans` + `JetBrains Mono` typography
- Sidebar navigation (Dashboard / Bridges / Countries / Logs / Settings)
- Custom draggable titlebar (frameless native window)
- **Go backend** (`app.go`) — 15+ exported methods auto-bound to JS via Wails IPC
    - Thread-safe with `sync.RWMutex`
    - All operations context-aware with timeouts
    - Real-time events: `tor:connecting`, `tor:connected`, `tor:bootstrap`, `log:line`
- **Svelte stores** (`lib/stores.js`) — reactive global state for all views
- **Wails IPC bridge** (`lib/wails.js`) — typed wrapper with browser dev-mode mock
- **Dashboard** — status card + exit IP card + proxy address + quick stats row
- **Bridges** — transport sidebar + built-in list + custom input + BridgeDB fetch tabs
- **Countries** — 3-column flag grid + selection panel + StrictNodes toggle
- **Logs** — real-time stream with level filter + text search + follow mode
- **Settings** — ports, auth mode cards, preferences, unsaved-changes indicator
- **Toast notification system** — 4-second ephemeral messages in bottom-right corner
- Windows 11: Dark theme via `windows.Theme` option
- Linux: GTK3 + WebKit2GTK

### Security
- SafeCookie HMAC-SHA256 challenge-response authentication
- GETINFO/SETCONF keyword allowlist (prevents control port injection)
- torrc written atomically at mode `0600`
- BridgeDB response capped at 64 KB
- IP check response capped at 8 KB
- SafeSocks=1 enabled by default
- Stream isolation enabled by default
- Cookie path validated against traversal

---

[Unreleased]: https://github.com/amirk1998/tor-manager/compare/v0.1.1...HEAD
[0.1.1]: https://github.com/amirk1998/tor-manager/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/amirk1998/tor-manager/releases/tag/v0.1.0