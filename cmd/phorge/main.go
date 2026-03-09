package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"

	"github.com/hinkers/Phorge/internal/config"
	"github.com/hinkers/Phorge/internal/tui"
)

var version = "dev"

func main() {
	// Parse arguments: phorge [nickname] [--ssh|--sftp|--db] [--version|-v]
	var jumpTarget string
	var action tui.LaunchAction

	for _, arg := range os.Args[1:] {
		switch arg {
		case "--version", "-v":
			fmt.Printf("phorge %s\n", version)
			os.Exit(0)
		case "--ssh", "-s":
			action = tui.LaunchSSH
		case "--sftp", "-f":
			action = tui.LaunchSFTP
		case "--db", "-d":
			action = tui.LaunchDB
		default:
			jumpTarget = arg
		}
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

	p := tea.NewProgram(tui.NewApp(cfg, jumpTarget, action))
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
