package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"

	"github.com/hinke/phorge/internal/config"
	"github.com/hinke/phorge/internal/tui"
)

var version = "dev"

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Printf("phorge %s\n", version)
		os.Exit(0)
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	if cfg.Forge.APIKey == "" {
		// Run the first-run setup flow to collect the API key.
		setupProgram := tea.NewProgram(tui.NewSetup(cfg))
		if _, err := setupProgram.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Setup error: %v\n", err)
			os.Exit(1)
		}

		// Reload config after setup (the setup flow saves the key).
		cfg, err = config.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}

		// If still no API key (user cancelled), exit gracefully.
		if cfg.Forge.APIKey == "" {
			return
		}
	}

	p := tea.NewProgram(tui.NewApp(cfg))
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
