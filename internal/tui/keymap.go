package tui

import "charm.land/bubbles/v2/key"

// GlobalKeyMap contains keybindings available in every context.
type GlobalKeyMap struct {
	Quit     key.Binding
	Refresh  key.Binding
	SSH      key.Binding
	SFTP     key.Binding
	Database key.Binding
	Help     key.Binding
	Tab      key.Binding
	ShiftTab key.Binding
}

// DefaultGlobalKeyMap returns the default global keybindings.
func DefaultGlobalKeyMap() GlobalKeyMap {
	return GlobalKeyMap{
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("ctrl+r"),
			key.WithHelp("ctrl+r", "refresh"),
		),
		SSH: key.NewBinding(
			key.WithKeys("ctrl+s"),
			key.WithHelp("ctrl+s", "ssh"),
		),
		SFTP: key.NewBinding(
			key.WithKeys("ctrl+f"),
			key.WithHelp("ctrl+f", "sftp"),
		),
		Database: key.NewBinding(
			key.WithKeys("ctrl+d"),
			key.WithHelp("ctrl+d", "database"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next panel"),
		),
		ShiftTab: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "prev panel"),
		),
	}
}

// NavKeyMap contains keybindings for list navigation.
type NavKeyMap struct {
	Up        key.Binding
	Down      key.Binding
	Enter     key.Binding
	Back      key.Binding
	Search    key.Binding
	Home      key.Binding
	End       key.Binding
	PageUp    key.Binding
	PageDown  key.Binding
}

// DefaultNavKeyMap returns the default navigation keybindings.
func DefaultNavKeyMap() NavKeyMap {
	return NavKeyMap{
		Up: key.NewBinding(
			key.WithKeys("k", "up"),
			key.WithHelp("k/up", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("j", "down"),
			key.WithHelp("j/down", "down"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc", "backspace"),
			key.WithHelp("esc", "back"),
		),
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
		),
		Home: key.NewBinding(
			key.WithKeys("g", "home"),
			key.WithHelp("g", "top"),
		),
		End: key.NewBinding(
			key.WithKeys("G", "end"),
			key.WithHelp("G", "bottom"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup", "ctrl+u"),
			key.WithHelp("pgup", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("pgdown", "ctrl+d"),
			key.WithHelp("pgdn", "page down"),
		),
	}
}

// SectionKeyMap contains keybindings for switching detail panel tabs (1-9).
type SectionKeyMap struct {
	Deployments key.Binding // 1
	Environment key.Binding // 2
	Databases   key.Binding // 3
	SSL         key.Binding // 4
	Workers     key.Binding // 5
	Commands    key.Binding // 6
	Logs        key.Binding // 7
	Git         key.Binding // 8
	Domains     key.Binding // 9
}

// DefaultSectionKeyMap returns the default section keybindings.
func DefaultSectionKeyMap() SectionKeyMap {
	return SectionKeyMap{
		Deployments: key.NewBinding(
			key.WithKeys("1"),
			key.WithHelp("1", "deployments"),
		),
		Environment: key.NewBinding(
			key.WithKeys("2"),
			key.WithHelp("2", "environment"),
		),
		Databases: key.NewBinding(
			key.WithKeys("3"),
			key.WithHelp("3", "databases"),
		),
		SSL: key.NewBinding(
			key.WithKeys("4"),
			key.WithHelp("4", "ssl"),
		),
		Workers: key.NewBinding(
			key.WithKeys("5"),
			key.WithHelp("5", "workers"),
		),
		Commands: key.NewBinding(
			key.WithKeys("6"),
			key.WithHelp("6", "commands"),
		),
		Logs: key.NewBinding(
			key.WithKeys("7"),
			key.WithHelp("7", "logs"),
		),
		Git: key.NewBinding(
			key.WithKeys("8"),
			key.WithHelp("8", "git"),
		),
		Domains: key.NewBinding(
			key.WithKeys("9"),
			key.WithHelp("9", "domains"),
		),
	}
}

// ServerActionKeyMap contains keybindings for server-level actions.
type ServerActionKeyMap struct {
	SSH    key.Binding
	SFTP   key.Binding
	Reboot key.Binding
}

// DefaultServerActionKeyMap returns the default server action keybindings.
func DefaultServerActionKeyMap() ServerActionKeyMap {
	return ServerActionKeyMap{
		SSH: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "ssh"),
		),
		SFTP: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "sftp"),
		),
		Reboot: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "reboot"),
		),
	}
}

// SiteActionKeyMap contains keybindings for site-level actions.
type SiteActionKeyMap struct {
	Deploy   key.Binding
	Env      key.Binding
	SSH      key.Binding
	Database key.Binding
	Logs     key.Binding
}

// DefaultSiteActionKeyMap returns the default site action keybindings.
func DefaultSiteActionKeyMap() SiteActionKeyMap {
	return SiteActionKeyMap{
		Deploy: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "deploy"),
		),
		Env: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "env"),
		),
		SSH: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "ssh"),
		),
		Database: key.NewBinding(
			key.WithKeys("D"),
			key.WithHelp("D", "database"),
		),
		Logs: key.NewBinding(
			key.WithKeys("l"),
			key.WithHelp("l", "logs"),
		),
	}
}
