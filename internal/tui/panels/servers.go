package panels

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	lipgloss "charm.land/lipgloss/v2"

	"github.com/hinkers/Phorge/internal/forge"
	"github.com/hinkers/Phorge/internal/tui/theme"
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
	errMsg   string // error message from API call

	// Filter state
	filterInput  textinput.Model
	filterActive bool   // true when the text input is visible and capturing keys
	filterText   string // the accepted filter text (persists after Enter)
	filtered     []int  // indices into the full servers slice

	// Keybindings
	up    key.Binding
	down  key.Binding
	enter key.Binding
	home  key.Binding
	end   key.Binding
}

// NewServerList creates a new, empty ServerList panel.
func NewServerList() ServerList {
	ti := textinput.New()
	ti.Prompt = "/ "
	ti.Placeholder = "filter servers..."
	ti.CharLimit = 64

	return ServerList{
		filterInput: ti,
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
	s.errMsg = ""
	s.filterActive = false
	s.filterText = ""
	s.filterInput.SetValue("")
	s.refilter()
	s.cursor = 0
	s.updateSelected()
	return s
}

// SetLoading sets the loading indicator state.
func (s ServerList) SetLoading(loading bool) ServerList {
	s.loading = loading
	return s
}

// SetError sets an error message to display in the panel.
func (s ServerList) SetError(err string) ServerList {
	s.errMsg = err
	s.loading = false
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

// FilterActive reports whether the filter input is currently active.
func (s ServerList) FilterActive() bool {
	return s.filterActive
}

// visibleList returns the servers to display (filtered or all).
func (s ServerList) visibleList() []forge.Server {
	if len(s.filtered) == 0 && s.filterText == "" {
		return s.servers
	}
	result := make([]forge.Server, 0, len(s.filtered))
	for _, idx := range s.filtered {
		if idx < len(s.servers) {
			result = append(result, s.servers[idx])
		}
	}
	return result
}

// refilter rebuilds the filtered index list based on the current filter text.
func (s *ServerList) refilter() {
	text := strings.ToLower(s.filterText)
	if text == "" {
		// Show all servers.
		s.filtered = make([]int, len(s.servers))
		for i := range s.servers {
			s.filtered[i] = i
		}
		return
	}
	s.filtered = nil
	for i, srv := range s.servers {
		if strings.Contains(strings.ToLower(srv.Name), text) {
			s.filtered = append(s.filtered, i)
		}
	}
}

// updateSelected updates the selected pointer based on the cursor and filtered list.
func (s *ServerList) updateSelected() {
	visible := s.visibleList()
	if len(visible) > 0 && s.cursor < len(visible) {
		srv := visible[s.cursor]
		s.selected = &srv
	} else {
		s.selected = nil
	}
}

// Update handles key events for the server list.
func (s ServerList) Update(msg tea.Msg) (Panel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if s.filterActive {
			return s.handleFilterKey(msg)
		}
		return s.handleKey(msg)
	}
	return s, nil
}

// handleFilterKey processes keys when the filter input is active.
func (s ServerList) handleFilterKey(msg tea.KeyPressMsg) (Panel, tea.Cmd) {
	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
		// Accept filter, hide input but keep filter text.
		s.filterActive = false
		s.filterText = s.filterInput.Value()
		s.refilter()
		s.cursor = 0
		s.updateSelected()
		if s.selected != nil {
			srv := *s.selected
			return s, func() tea.Msg { return ServerSelectedMsg{Server: srv} }
		}
		return s, nil

	case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
		// Cancel filter, clear everything.
		s.filterActive = false
		s.filterText = ""
		s.filterInput.SetValue("")
		s.refilter()
		s.cursor = 0
		s.updateSelected()
		if s.selected != nil {
			srv := *s.selected
			return s, func() tea.Msg { return ServerSelectedMsg{Server: srv} }
		}
		return s, nil
	}

	// Delegate to textinput for character input.
	var cmd tea.Cmd
	s.filterInput, cmd = s.filterInput.Update(msg)

	// Live-filter as the user types.
	s.filterText = s.filterInput.Value()
	s.refilter()
	s.cursor = 0
	s.updateSelected()

	return s, cmd
}

func (s ServerList) handleKey(msg tea.KeyPressMsg) (Panel, tea.Cmd) {
	visible := s.visibleList()

	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("/"))):
		// Activate filter mode.
		s.filterActive = true
		s.filterInput.SetValue(s.filterText)
		s.filterInput.Focus()
		return s, textinput.Blink

	case key.Matches(msg, s.down):
		if len(visible) > 0 {
			s.cursor = min(s.cursor+1, len(visible)-1)
			s.updateSelected()
			if s.selected != nil {
				srv := *s.selected
				return s, func() tea.Msg { return ServerSelectedMsg{Server: srv} }
			}
		}

	case key.Matches(msg, s.up):
		if len(visible) > 0 {
			s.cursor = max(s.cursor-1, 0)
			s.updateSelected()
			if s.selected != nil {
				srv := *s.selected
				return s, func() tea.Msg { return ServerSelectedMsg{Server: srv} }
			}
		}

	case key.Matches(msg, s.enter):
		if s.selected != nil {
			srv := *s.selected
			return s, func() tea.Msg { return ServerSelectedMsg{Server: srv} }
		}

	case key.Matches(msg, s.home):
		if len(visible) > 0 {
			s.cursor = 0
			s.updateSelected()
			if s.selected != nil {
				srv := *s.selected
				return s, func() tea.Msg { return ServerSelectedMsg{Server: srv} }
			}
		}

	case key.Matches(msg, s.end):
		if len(visible) > 0 {
			s.cursor = len(visible) - 1
			s.updateSelected()
			if s.selected != nil {
				srv := *s.selected
				return s, func() tea.Msg { return ServerSelectedMsg{Server: srv} }
			}
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

	// Render the filter input at the top when active, or show filter indicator.
	if s.filterActive {
		filterLine := s.filterInput.View()
		lines = append(lines, theme.Truncate(filterLine, innerWidth))
		innerHeight-- // account for the filter input line
	} else if s.filterText != "" {
		indicator := theme.FilterIndicatorStyle.
			Render("filter: " + s.filterText)
		lines = append(lines, theme.Truncate(indicator, innerWidth))
		innerHeight-- // account for the filter indicator line
	}

	visible := s.visibleList()

	if s.loading && len(s.servers) == 0 {
		lines = append(lines, theme.LoadingStyle.Render("Loading servers..."))
	} else if s.errMsg != "" {
		lines = append(lines, theme.ErrorStatusStyle.Render(s.errMsg))
		lines = append(lines, theme.NormalItemStyle.Render("Press ctrl+r to retry"))
	} else if len(s.servers) == 0 {
		lines = append(lines, theme.NormalItemStyle.Render("No servers found"))
	} else if len(visible) == 0 && s.filterText != "" {
		lines = append(lines, theme.NormalItemStyle.Render("No matching servers"))
	} else {
		// Calculate visible range with scrolling.
		visibleHeight := innerHeight
		if visibleHeight < 1 {
			visibleHeight = 1
		}
		startIdx := 0
		if s.cursor >= visibleHeight {
			startIdx = s.cursor - visibleHeight + 1
		}

		for i := startIdx; i < len(visible) && len(lines)-countFilterLines(s) < visibleHeight; i++ {
			srv := visible[i]
			name := theme.Truncate(srv.Name, innerWidth-4)
			if i == s.cursor {
				line := theme.CursorStyle.Render("> ") + theme.SelectedItemStyle.Render(name)
				lines = append(lines, line)
			} else {
				line := "  " + theme.NormalItemStyle.Render(name)
				lines = append(lines, line)
			}
		}
	}

	// Pad to fill the panel height.
	totalHeight := height - 2
	if totalHeight < 0 {
		totalHeight = 0
	}
	for len(lines) < totalHeight {
		lines = append(lines, "")
	}

	content := strings.Join(lines, "\n")

	return style.
		Width(width - 2).
		Height(totalHeight).
		Render(title + "\n" + content)
}

// countFilterLines returns the number of header lines used by the filter UI.
func countFilterLines(s ServerList) int {
	if s.filterActive || s.filterText != "" {
		return 1
	}
	return 0
}

// HelpBindings returns the key hints for the server list.
func (s ServerList) HelpBindings() []HelpBinding {
	if s.filterActive {
		return []HelpBinding{
			{Key: "enter", Desc: "accept filter"},
			{Key: "esc", Desc: "clear filter"},
		}
	}
	bindings := []HelpBinding{
		{Key: "j/k", Desc: "navigate"},
		{Key: "enter", Desc: "select"},
		{Key: "/", Desc: "filter"},
		{Key: "g/G", Desc: "top/bottom"},
		{Key: "r", Desc: "reboot"},
		{Key: "tab", Desc: "switch panel"},
		{Key: "ctrl+r", Desc: "refresh"},
		{Key: "q", Desc: "quit"},
	}
	return bindings
}
