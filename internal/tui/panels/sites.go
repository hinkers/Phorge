package panels

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/key"
	lipgloss "charm.land/lipgloss/v2"

	"github.com/hinke/phorge/internal/forge"
	"github.com/hinke/phorge/internal/tui/theme"
)

// SiteSelectedMsg is emitted when the user presses Enter on a site.
type SiteSelectedMsg struct {
	Site forge.Site
}

// SiteList is a scrollable panel that displays the list of sites for the
// currently selected server.
type SiteList struct {
	sites      []forge.Site
	cursor     int
	selected   *forge.Site
	serverName string // displayed in the panel title

	// Keybindings
	up    key.Binding
	down  key.Binding
	enter key.Binding
	home  key.Binding
	end   key.Binding
}

// NewSiteList creates a new, empty SiteList panel.
func NewSiteList() SiteList {
	return SiteList{
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

// SetSites replaces the site list and resets the cursor.
func (s SiteList) SetSites(sites []forge.Site) SiteList {
	s.sites = sites
	s.cursor = 0
	if len(sites) > 0 {
		site := sites[0]
		s.selected = &site
	} else {
		s.selected = nil
	}
	return s
}

// SetServerName updates the server name shown in the panel title.
func (s SiteList) SetServerName(name string) SiteList {
	s.serverName = name
	return s
}

// Selected returns a pointer to the currently highlighted site, or nil.
func (s SiteList) Selected() *forge.Site {
	return s.selected
}

// Cursor returns the current cursor index.
func (s SiteList) Cursor() int {
	return s.cursor
}

// Update handles key events for the site list.
func (s SiteList) Update(msg tea.Msg) (Panel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		return s.handleKey(msg)
	}
	return s, nil
}

func (s SiteList) handleKey(msg tea.KeyPressMsg) (Panel, tea.Cmd) {
	switch {
	case key.Matches(msg, s.down):
		if len(s.sites) > 0 {
			s.cursor = min(s.cursor+1, len(s.sites)-1)
			site := s.sites[s.cursor]
			s.selected = &site
			return s, func() tea.Msg { return SiteSelectedMsg{Site: site} }
		}

	case key.Matches(msg, s.up):
		if len(s.sites) > 0 {
			s.cursor = max(s.cursor-1, 0)
			site := s.sites[s.cursor]
			s.selected = &site
			return s, func() tea.Msg { return SiteSelectedMsg{Site: site} }
		}

	case key.Matches(msg, s.enter):
		if s.selected != nil {
			site := *s.selected
			return s, func() tea.Msg { return SiteSelectedMsg{Site: site} }
		}

	case key.Matches(msg, s.home):
		if len(s.sites) > 0 {
			s.cursor = 0
			site := s.sites[0]
			s.selected = &site
			return s, func() tea.Msg { return SiteSelectedMsg{Site: site} }
		}

	case key.Matches(msg, s.end):
		if len(s.sites) > 0 {
			s.cursor = len(s.sites) - 1
			site := s.sites[s.cursor]
			s.selected = &site
			return s, func() tea.Msg { return SiteSelectedMsg{Site: site} }
		}
	}

	return s, nil
}

// View renders the site list panel.
func (s SiteList) View(width, height int, focused bool) string {
	style := theme.InactiveBorderStyle
	titleColor := theme.ColorSubtle
	if focused {
		style = theme.ActiveBorderStyle
		titleColor = theme.ColorPrimary
	}

	panelTitle := "Sites"
	if s.serverName != "" {
		panelTitle = fmt.Sprintf("Sites (%s)", s.serverName)
	}
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(titleColor).
		Render(" " + panelTitle + " ")

	// Account for border size.
	innerWidth := width - 2
	innerHeight := height - 2
	if innerWidth < 0 {
		innerWidth = 0
	}
	if innerHeight < 0 {
		innerHeight = 0
	}

	var lines []string

	if s.serverName == "" && len(s.sites) == 0 {
		lines = append(lines, theme.NormalItemStyle.Render("Select a server"))
	} else if len(s.sites) == 0 {
		lines = append(lines, theme.NormalItemStyle.Render("No sites found"))
	} else {
		for i, site := range s.sites {
			name := theme.Truncate(site.Name, innerWidth-4)
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

// HelpBindings returns the key hints for the site list.
func (s SiteList) HelpBindings() []HelpBinding {
	return []HelpBinding{
		{Key: "j/k", Desc: "navigate"},
		{Key: "enter", Desc: "select"},
		{Key: "g/G", Desc: "top/bottom"},
		{Key: "esc", Desc: "back"},
		{Key: "tab", Desc: "switch panel"},
		{Key: "q", Desc: "quit"},
	}
}
