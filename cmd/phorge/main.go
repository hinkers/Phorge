package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"

	"github.com/hinke/phorge/internal/config"
	"github.com/hinke/phorge/internal/tui"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	if cfg.Forge.APIKey == "" {
		fmt.Fprintln(os.Stderr, "No API key configured. Set it in ~/.config/phorge/config.toml")
		os.Exit(1)
	}

	p := tea.NewProgram(tui.NewApp(cfg))
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
