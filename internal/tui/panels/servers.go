package panels

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/key"
	lipgloss "charm.land/lipgloss/v2"

	"github.com/hinke/phorge/internal/forge"
	"github.com/hinke/phorge/internal/tui/theme"
)

// ServerSelectedMsg is emitted when the user presses Enter on a server.
type ServerSelectedMsg struct {
	Server forge.Server
}

// ServerList is a scrollable panel that displays the list of Forge servers.
type ServerList struct {
	servers  []forge.Server
	cursor   int
	selected *forge.Server
	loading  bool

	// Keybindings
	up    key.Binding
	down  key.Binding
	enter key.Binding
	home  key.Binding
	end   key.Binding
}

// NewServerList creates a new, empty ServerList panel.
func NewServerList() ServerList {
	return ServerList{
		up: key.NewBinding(
			key.WithKeys("k", "up"),
			key.WithHelp("k/up", "up"),
		),
		down: key.NewBinding(
			key.WithKeys("j", "down"),
			key.WithHelp("j/down", "down"),
		),
		enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		home: key.NewBinding(
			key.WithKeys("g", "home"),
			key.WithHelp("g", "top"),
		),
		end: key.NewBinding(
			key.WithKeys("G", "end"),
			key.WithHelp("G", "bottom"),
		),
	}
}

// SetServers replaces the server list and resets the cursor.
func (s ServerList) SetServers(servers []forge.Server) ServerList {
	s.servers = servers
	s.cursor = 0
	if len(servers) > 0 {
		srv := servers[0]
		s.selected = &srv
	} else {
		s.selected = nil
	}
	return s
}

// SetLoading sets the loading indicator state.
func (s ServerList) SetLoading(loading bool) ServerList {
	s.loading = loading
	return s
}

// Selected returns a pointer to the currently highlighted server, or nil.
func (s ServerList) Selected() *forge.Server {
	return s.selected
}

// Cursor returns the current cursor index.
func (s ServerList) Cursor() int {
	return s.cursor
}

// Update handles key events for the server list.
func (s ServerList) Update(msg tea.Msg) (Panel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		return s.handleKey(msg)
	}
	return s, nil
}

func (s ServerList) handleKey(msg tea.KeyPressMsg) (Panel, tea.Cmd) {
	switch {
	case key.Matches(msg, s.down):
		if len(s.servers) > 0 {
			s.cursor = min(s.cursor+1, len(s.servers)-1)
			srv := s.servers[s.cursor]
			s.selected = &srv
			return s, func() tea.Msg { return ServerSelectedMsg{Server: srv} }
		}

	case key.Matches(msg, s.up):
		if len(s.servers) > 0 {
			s.cursor = max(s.cursor-1, 0)
			srv := s.servers[s.cursor]
			s.selected = &srv
			return s, func() tea.Msg { return ServerSelectedMsg{Server: srv} }
		}

	case key.Matches(msg, s.enter):
		if s.selected != nil {
			srv := *s.selected
			return s, func() tea.Msg { return ServerSelectedMsg{Server: srv} }
		}

	case key.Matches(msg, s.home):
		if len(s.servers) > 0 {
			s.cursor = 0
			srv := s.servers[0]
			s.selected = &srv
			return s, func() tea.Msg { return ServerSelectedMsg{Server: srv} }
		}

	case key.Matches(msg, s.end):
		if len(s.servers) > 0 {
			s.cursor = len(s.servers) - 1
			srv := s.servers[s.cursor]
			s.selected = &srv
			return s, func() tea.Msg { return ServerSelectedMsg{Server: srv} }
		}
	}

	return s, nil
}

// View renders the server list panel.
func (s ServerList) View(width, height int, focused bool) string {
	style := theme.InactiveBorderStyle
	titleColor := theme.ColorSubtle
	if focused {
		style = theme.ActiveBorderStyle
		titleColor = theme.ColorPrimary
	}

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(titleColor).
		Render(" Servers ")

	// Account for border size (1px on each side).
	innerWidth := width - 2
	innerHeight := height - 2
	if innerWidth < 0 {
		innerWidth = 0
	}
	if innerHeight < 0 {
		innerHeight = 0
	}

	var lines []string

	if s.loading && len(s.servers) == 0 {
		lines = append(lines, theme.LoadingStyle.Render("Loading servers..."))
	} else if len(s.servers) == 0 {
		lines = append(lines, theme.NormalItemStyle.Render("No servers found"))
	} else {
		for i, srv := range s.servers {
			name := theme.Truncate(srv.Name, innerWidth-4)
			if i == s.cursor {
				line := theme.CursorStyle.Render("> ") + theme.SelectedItemStyle.Render(name)
				lines = append(lines, line)
			} else {
				line := "  " + theme.NormalItemStyle.Render(name)
				lines = append(lines, line)
			}
			if i >= innerHeight-1 {
				break
			}
		}
	}

	// Pad to fill the panel height.
	for len(lines) < innerHeight {
		lines = append(lines, "")
	}

	content := strings.Join(lines, "\n")

	return style.
		Width(innerWidth).
		Height(innerHeight).
		Render(title + "\n" + content)
}

// HelpBindings returns the key hints for the server list.
func (s ServerList) HelpBindings() []HelpBinding {
	return []HelpBinding{
		{Key: "j/k", Desc: "navigate"},
		{Key: "enter", Desc: "select"},
		{Key: "g/G", Desc: "top/bottom"},
		{Key: "r", Desc: "reboot"},
		{Key: "tab", Desc: "switch panel"},
		{Key: "ctrl+r", Desc: "refresh"},
		{Key: "q", Desc: "quit"},
	}
}
