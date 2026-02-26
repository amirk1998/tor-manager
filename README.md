<div align="center">

```
████████╗ ██████╗ ██████╗
   ██╔══╝██╔═══██╗██╔══██╗
   ██║   ██║   ██║██████╔╝
   ██║   ██║   ██║██╔══██╗
   ██║   ╚██████╔╝██║  ██║
   ╚═╝    ╚═════╝ ╚═╝  ╚═╝
        M A N A G E R
```

**A professional, open-source Tor management suite**
built entirely in Go — available as a Terminal UI and a Cross-Platform Desktop App.

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat-square&logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-violet?style=flat-square)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20Windows%20%7C%20macOS-lightgrey?style=flat-square)](https://github.com/yourname/tor-manager)
[![Wails](https://img.shields.io/badge/GUI-Wails%20v2-red?style=flat-square)](https://wails.io)
[![Svelte](https://img.shields.io/badge/Frontend-Svelte-FF3E00?style=flat-square&logo=svelte)](https://svelte.dev)

</div>

---

## 🧅 What is tor-manager?

**tor-manager** is a complete, production-grade suite for managing your local Tor daemon — without ever touching a config file or running cryptic commands.

It gives you real-time visibility into your Tor circuit, lets you switch bridges and exit countries in seconds, and wraps Tor's powerful control protocol in a clean, intuitive interface — whether you prefer the terminal or a native desktop window.

It is **not** a VPN. It is **not** a browser extension. It is a first-class management layer that sits on top of your existing Tor installation and exposes its full capability through an ergonomic interface.

---

## ✨ Feature Highlights

### 🔌 Connection & Identity
- **Real-time bootstrap tracking** — monitors Tor's startup progress from 0% to 100% with live status updates
- **Live exit IP verification** — checks your current IP against the official Tor Project endpoint (`check.torproject.org`) to confirm you are genuinely routing through a Tor exit node
- **One-click new identity** — triggers `SIGNAL NEWNYM` with enforced 10-second cooldown (respecting Tor's rate limit)
- **Geolocation display** — shows exit node country, city, ASN, and round-trip latency
- **SOCKS5 proxy address** — instantly copy `socks5://127.0.0.1:9050` to point any app through Tor

### 🌉 Bridge Management
- **Full transport support** — obfs4, Snowflake, WebTunnel, meek-azure, and direct (vanilla)
- **Built-in bridge library** — curated, up-to-date bridges sourced from Tor Browser for every transport type
- **BridgeDB integration** — request fresh bridges directly from the Tor Project's official MOAT API (`bridges.torproject.org`), routed through Tor for privacy
- **Custom bridge input** — paste any valid bridge line with full parsing and structural validation before applying
- **Live apply** — bridge changes are applied to the running daemon via the control port, no restart required

### 🌍 Exit Country Control
- **Multi-country exit selection** — choose one or more countries as exit node candidates
- **StrictNodes toggle** — enforce strict country filtering (with an anonymity warning)
- **Instant apply** — updates `ExitNodes` and `StrictNodes` on the running daemon immediately

### 📋 Logs
- **Real-time log stream** — Tor daemon output rendered as it happens
- **Syntax-aware coloring** — bootstrap events, circuit builds, warnings, and errors each get distinct styling
- **Level filtering** — view all, notice, warn, or error entries
- **Text search** — filter log lines by keyword
- **Follow mode** — auto-scroll to the latest entry, or lock scroll to review history

### ⚙️ Settings
- **Port configuration** — change SOCKS5 and control port without editing torrc manually
- **Authentication modes** — cookie auth (recommended, more secure) or password auth
- **Auto-refresh interval** — configure how often the exit IP check runs automatically
- **Persistent config** — settings are saved to the platform-appropriate config directory (XDG on Linux, AppData on Windows, Library on macOS)

---

## 🏗️ Architecture

This is a **monorepo** composed of three independent Go modules that share a common core library.

```
tor-manager/
│
├── core/                        ← Shared library (no UI dependencies)
│   ├── tor/
│   │   ├── controller.go        Tor Control Port client (thread-safe, multi-auth)
│   │   ├── bridge.go            Bridge types, parsing, validation, live apply
│   │   ├── builtin_bridges.go   Curated built-in bridge list (obfs4/snowflake/webtunnel/meek)
│   │   ├── bridgedb.go          BridgeDB MOAT API client (size-limited, Tor-routed)
│   │   ├── config.go            torrc generation with atomic write (mode 0600)
│   │   ├── process.go           Tor daemon lifecycle + bootstrap event tracking
│   │   ├── cookie.go            Cross-platform cookie file reader
│   │   └── errors.go            Sentinel error types
│   ├── proxy/
│   │   └── checker.go           IP verification via check.torproject.org + ipinfo.io
│   └── config/
│       └── appconfig.go         Persistent user settings (XDG-aware, atomic save)
│
├── tor-tui/                     ← Terminal UI (Bubble Tea + Lip Gloss)
│   ├── main.go
│   └── ui/
│       ├── app.go               Root model: tab routing, boot screen, help overlay
│       ├── styles.go            Centralized design system (palette, components)
│       ├── keys.go              All keybindings in one place
│       ├── messages.go          Tea message types for inter-component communication
│       └── views/
│           ├── dashboard.go     Connection status, exit IP, proxy address
│           ├── bridges.go       Bridge manager (built-in / custom / BridgeDB)
│           ├── countries.go     Exit country selector with multi-select
│           ├── logs.go          Scrollable log viewer with follow mode
│           └── settings.go      Configuration editor with live validation
│
└── tor-gui/                     ← Desktop GUI (Wails v2 + Svelte + Tailwind)
    ├── main.go                  Window configuration
    ├── app.go                   Go backend — all methods exposed to Svelte via IPC
    └── frontend/
        └── src/
            ├── App.svelte       Root: titlebar, sidebar nav, event wiring
            ├── app.css          Design system (Tailwind + custom components)
            ├── lib/
            │   ├── wails.js     Type-safe Wails IPC bridge (dev-mode mock included)
            │   └── stores.js    Svelte reactive state (connected, ipInfo, logs, ...)
            └── components/
                ├── Dashboard.svelte
                ├── Bridges.svelte
                ├── Countries.svelte
                ├── Logs.svelte
                └── Settings.svelte
```

### Data Flow

```
  Svelte Component
       │  await GetIPInfo()           (Wails IPC — zero serialization overhead)
       ▼
  frontend/lib/wails.js               Safe wrapper, works in browser dev mode too
       │  window.go.main.App.GetIPInfo()
       ▼
  tor-gui/app.go                      Go backend: mutex, context timeout, error wrap
       │  checker.CheckIP(ctx)
       ▼
  core/proxy/checker.go               SOCKS5-routed HTTP → check.torproject.org
       │
       ▼
  ──── real-time events ────
  app.go  ──EventsEmit──▶  App.svelte  ──▶  stores.js  ──▶  every component
```

---

## 🔒 Security Design

Security is a first-class concern throughout the codebase — not an afterthought.

| Layer | Measure |
|-------|---------|
| **Control Port auth** | SafeCookie HMAC challenge-response supported; password sanitized to prevent injection |
| **GETINFO / SETCONF** | Keys are allowlisted against a strict regex — no arbitrary string can reach the control port |
| **torrc generation** | Written atomically via temp-file + rename; file mode `0600` (owner-readable only) |
| **BridgeDB requests** | Response body capped at 64 KB to prevent memory exhaustion; routed through Tor when available |
| **SafeSocks** | Enabled by default (`SafeSocks 1`) — rejects SOCKS4 requests that leak DNS |
| **Stream isolation** | `IsolateClientAddr 1` and `IsolateClientProtocol 1` enabled by default |
| **HTTP client** | `DisableKeepAlives` on proxy checker — each request gets a fresh TCP connection |
| **NEWNYM cooldown** | Enforced in code (10 s minimum) — prevents accidental rapid-fire identity changes |
| **Config file** | Saved atomically with a temp+rename pattern; no partial writes |
| **Cookie path** | Validated against path traversal before reading |

---

## 🚀 Getting Started

### Prerequisites

```bash
# Tor daemon
sudo apt install tor                  # Debian / Ubuntu
brew install tor                      # macOS
# Windows: download from https://www.torproject.org

# Pluggable transports (for bridge support)
sudo apt install obfs4proxy           # obfs4 + meek-azure

# For tor-gui only
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# Linux GUI dependencies
sudo apt install libgtk-3-dev libwebkit2gtk-4.0-dev
```

**Enable the control port** in your `torrc`:

```
# /etc/tor/torrc  (Linux)
# %APPDATA%\Tor\torrc  (Windows)

ControlPort 9051
CookieAuthentication 1
```

Then restart Tor:
```bash
sudo systemctl restart tor       # Linux (systemd)
brew services restart tor        # macOS
# Windows: restart via Services or Tor Browser
```

---

### Clone & Build

```bash
git clone https://github.com/amirk1998/tor-manager.git
cd tor-manager
```

#### Terminal UI (tor-tui)

```bash
cd tor-tui

# Install dependencies
go mod download

# Run directly
go run .

# Build binary
go build -ldflags "-s -w" -o bin/tor-tui .

# Windows
go build -ldflags "-s -w" -o bin/tor-tui.exe .
```

#### Desktop GUI (tor-gui)

```bash
cd tor-gui

# Install frontend dependencies
cd frontend && npm install && cd ..

# Development — hot reload for both Go and Svelte
wails dev

# Production build
wails build -clean -ldflags "-s -w"

# With UPX compression (~40% smaller binary)
wails build -clean -upx

# Windows cross-compile (from Linux/macOS)
wails build -platform windows/amd64
```

---

## 🖥️ Usage

### TUI Keyboard Reference

| Key | Action |
|-----|--------|
| `1` – `5` | Switch tabs |
| `tab` / `shift+tab` | Next / previous tab |
| `?` | Toggle help overlay |
| `q` / `ctrl+c` | Quit |
| **Dashboard** | |
| `r` | Refresh exit IP |
| `n` | New identity (10 s cooldown enforced) |
| `c` | Copy proxy address to clipboard |
| **Bridges** | |
| `←` `→` | Switch transport type |
| `↑` `↓` / `enter` | Navigate and select bridge |
| `u` | Add a custom bridge line |
| `f` | Fetch fresh bridges from BridgeDB |
| `a` | Apply current bridge configuration |
| `x` | Disable all bridges |
| **Countries** | |
| `↑` `↓` | Navigate country list |
| `enter` / `space` | Toggle country selection |
| `s` | Toggle StrictNodes |
| `a` | Apply country filter |
| `x` | Clear filter (any country) |
| **Logs** | |
| `↑` `↓` | Scroll |
| `g` / `G` | Jump to top / bottom |
| `f` | Toggle follow mode |
| `c` | Clear log buffer |

### Pointing Applications at Tor

Once running, configure any application's proxy settings:

```
Protocol:  SOCKS5
Host:      127.0.0.1
Port:      9050
```

**Examples:**

```bash
# curl
curl --socks5 127.0.0.1:9050 https://check.torproject.org/api/ip

# git
git config --global http.proxy socks5://127.0.0.1:9050

# Environment variable (many CLI tools)
export ALL_PROXY=socks5://127.0.0.1:9050
```

---

## 🔧 Technical Stack

| Component | Technology | Why |
|-----------|-----------|-----|
| Core library | Go 1.22 | Performance, single binary, CGO-free |
| TUI framework | [Bubble Tea](https://github.com/charmbracelet/bubbletea) | The Elm Architecture — predictable state |
| TUI styling | [Lip Gloss](https://github.com/charmbracelet/lipgloss) | Declarative terminal layout |
| GUI framework | [Wails v2](https://wails.io) | Go + web UI, no CGO, no Electron |
| GUI frontend | [Svelte 4](https://svelte.dev) | Zero virtual DOM, smallest bundle |
| GUI styling | [Tailwind CSS 3](https://tailwindcss.com) | Utility-first, purged in production |
| Build tool | [Vite](https://vitejs.dev) | Fast HMR, optimized production builds |
| Compression | [UPX](https://upx.github.io) | Optional binary compression |

---

## 🛠️ Development

### Module Structure

Each module has its own `go.mod` and uses a `replace` directive during development:

```go
// tor-tui/go.mod  and  tor-gui/go.mod
replace github.com/amirk1998/tor-manager/core => ../core
```

### Running Tests

```bash
# Core library tests
cd core && go test ./... -v

# With race detector
cd core && go test -race ./...
```

### Project-wide Commands (root Makefile)

```bash
make dev-tui        # Run TUI with hot-reload
make dev-gui        # Run GUI with hot-reload (Wails)
make build-tui      # Build TUI binary
make build-gui      # Build GUI binary
make build-all      # Build both
make tidy           # go mod tidy for all modules
make test           # Run all tests
```

---

## 📦 Release Artifacts

| Platform | TUI | GUI |
|----------|-----|-----|
| Linux x64 | `tor-tui-linux-amd64` | `tor-gui` |
| Windows x64 | `tor-tui-windows.exe` | `tor-gui.exe` |
| macOS x64 | `tor-tui-macos-amd64` | `tor-gui.app` |
| macOS ARM64 | `tor-tui-macos-arm64` | `tor-gui-arm64.app` |

Binary sizes (approximate, with UPX):

| Binary | Without UPX | With UPX |
|--------|------------|---------|
| tor-tui | ~8 MB | ~3 MB |
| tor-gui | ~15 MB | ~8 MB |

---

## 🗺️ Roadmap

- [ ] **v0.2** — Embedded Tor daemon (no separate installation required)
- [ ] **v0.2** — Windows installer (NSIS) via `wails build -nsis`
- [ ] **v0.3** — Tor circuit visualizer (show relay hops on a world map)
- [ ] **v0.3** — Onion service management (create and manage `.onion` addresses)
- [ ] **v0.4** — Auto-update for built-in bridge list (fetch from Tor Project CDN)
- [ ] **v0.4** — System tray integration for the GUI

---

## 🤝 Contributing

Contributions are welcome. Please follow the commit convention:

```
<type>(<scope>): <subject>

Types:   feat | fix | chore | refactor | docs | test
Scopes:  core | tui | gui | bridges | countries | logs
```

**Branch strategy:**
- `main` — stable, tagged releases only
- `dev` — integration branch, base for all feature branches
- `feat/*` — individual features, PR into `dev`

---

## 📄 License

MIT License — see [LICENSE](LICENSE) for full text.

---

## ⚠️ Disclaimer

This software interfaces with the Tor network. Using Tor is legal in most countries, but you are responsible for understanding and complying with the laws of your jurisdiction. This tool does not provide anonymity by itself — it manages your Tor daemon. Always use Tor responsibly.

---

<div align="center">

Built with Go · Bubble Tea · Wails · Svelte · Tailwind

**If this project is useful to you, consider starring it on GitHub ⭐**

</div>
