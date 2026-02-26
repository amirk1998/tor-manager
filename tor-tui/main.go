package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/you/tor-manager/core/config"
	"github.com/you/tor-tui/ui"
)

func main() {
	// Load or create user config
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "tor-tui: cannot load config: %v\n", err)
		os.Exit(1)
	}

	// Validate config before starting
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "tor-tui: invalid config: %v\n", err)
		os.Exit(1)
	}

	// Build the root model
	app := ui.NewApp(cfg)

	// Launch Bubble Tea with alt-screen for a clean full-terminal experience
	p := tea.NewProgram(
		app,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "tor-tui: %v\n", err)
		os.Exit(1)
	}
}
