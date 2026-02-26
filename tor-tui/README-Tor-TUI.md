# tor-tui 🧅

A professional terminal UI for managing your Tor daemon — built with Go and Bubble Tea.

```
┌─────────────────────────────────────────────────────────────┐
│ 🧅 tor-manager                                      v0.1.0  │
├──────────────────────────────────────────────────────────────│
│  ◈ Dashboard   ⬡ Bridges   ◎ Countries   ≡ Logs   ⚙ Settings│
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  ▸ Connection Status                                         │
│                                                              │
│    Status          ● Connected                               │
│    Tor Version     0.4.8.10                                  │
│    Last Checked    14:32:05                                  │
│                                                              │
│  ▸ Exit Node Info                                            │
│                                                              │
│    Exit IP         185.220.xxx.xxx  ✓ TOR EXIT               │
│    Location        Frankfurt, DE                             │
│    Network         AS24940 Hetzner                           │
│    Latency         1.24s                                     │
│                                                              │
│  ▸ Proxy Addresses                                           │
│                                                              │
│    SOCKS5          socks5://127.0.0.1:9050                   │
│                                                              │
│  [r] Refresh IP  ·  [n] New Identity  ·  [c] Copy Proxy     │
├──────────────────────────────────────────────────────────────┤
│  [1-5] tabs  ·  [tab] next  ·  [?] help  ·  [q] quit        │
└─────────────────────────────────────────────────────────────┘
```

## Features

| Tab | Description |
|-----|-------------|
| **Dashboard** | Real-time IP, exit node location, connection status |
| **Bridges** | Switch transport (obfs4/snowflake/meek/webtunnel), add custom bridges, fetch fresh from BridgeDB |
| **Countries** | Pick one or more exit countries, toggle StrictNodes |
| **Logs** | Scrollable Tor daemon log with follow mode and syntax coloring |
| **Settings** | Ports, auth method, auto-refresh interval |

## Prerequisites

```bash
# Debian / Ubuntu
sudo apt install tor obfs4proxy

# macOS
brew install tor

# Arch
sudo pacman -S tor obfs4proxy
```

Ensure Tor is running and its control port is enabled:
```
# /etc/tor/torrc
ControlPort 9051
CookieAuthentication 1
```

## Build & Run

```bash
# Download dependencies
make deps

# Run directly (dev mode)
make run

# Build binary
make build
./bin/tor-tui

# Install to PATH
make install
tor-tui
```

## Cross-compile

```bash
make cross
# Produces:
#   bin/tor-tui-linux-amd64
#   bin/tor-tui-macos-amd64
#   bin/tor-tui-macos-arm64
#   bin/tor-tui-windows.exe
```

## Keyboard Reference

| Key | Action |
|-----|--------|
| `1–5` | Switch tabs |
| `tab / shift+tab` | Next / previous tab |
| `?` | Toggle help overlay |
| `q / ctrl+c` | Quit |
| **Dashboard** | |
| `r` | Refresh IP |
| `n` | New identity (10s cooldown) |
| `c` | Copy proxy address |
| **Bridges** | |
| `← →` | Switch transport type |
| `↑ ↓ / enter` | Navigate & select bridge |
| `u` | Add a custom bridge |
| `f` | Fetch bridges from BridgeDB |
| `a` | Apply bridge configuration |
| `x` | Disable bridges |
| **Countries** | |
| `↑ ↓ / enter` | Toggle exit country |
| `s` | Toggle StrictNodes |
| `a` | Apply |
| **Logs** | |
| `↑ ↓` | Scroll |
| `g / G` | Top / bottom |
| `f` | Toggle follow mode |
| **Settings** | |
| `tab` | Next field |
| `ctrl+s` | Save |
| `esc` | Revert changes |

## Architecture

```
tor-tui/
├── main.go              — entry point
├── ui/
│   ├── app.go           — root Bubble Tea model, tab routing
│   ├── styles.go        — centralized design system (palette, components)
│   ├── keys.go          — all keybindings
│   ├── messages.go      — Tea message types
│   └── views/
│       ├── dashboard.go — connection status & IP info
│       ├── bridges.go   — bridge manager (builtin/custom/BridgeDB)
│       ├── countries.go — exit country selector
│       ├── logs.go      — scrollable log viewer
│       └── settings.go  — configuration editor
└── Makefile
```

Depends on `../core` (the shared library from Phase 1).
