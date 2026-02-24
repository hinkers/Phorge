package panels

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	lipgloss "charm.land/lipgloss/v2"

	"github.com/hinkers/Phorge/internal/forge"
	"github.com/hinkers/Phorge/internal/tui/theme"
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
	loading    bool
	errMsg     string // error message from API call

	// Filter state
	filterInput  textinput.Model
	filterActive bool   // true when the text input is visible and capturing keys
	filterText   string // the accepted filter text (persists after Enter)
	filtered     []int  // indices into the full sites slice

	// Keybindings
	up    key.Binding
	down  key.Binding
	enter key.Binding
	home  key.Binding
	end   key.Binding
}

// NewSiteList creates a new, empty SiteList panel.
func NewSiteList() SiteList {
	ti := textinput.New()
	ti.Prompt = "/ "
	ti.Placeholder = "filter sites..."
	ti.CharLimit = 64

	return SiteList{
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

// SetSites replaces the site list and resets the cursor.
func (s SiteList) SetSites(sites []forge.Site) SiteList {
	s.sites = sites
	s.errMsg = ""
	s.loading = false
	s.filterActive = false
	s.filterText = ""
	s.filterInput.SetValue("")
	s.refilter()
	s.cursor = 0
	s.updateSelected()
	return s
}

// SetServerName updates the server name shown in the panel title.
func (s SiteList) SetServerName(name string) SiteList {
	s.serverName = name
	return s
}

// SetLoading sets the loading indicator state.
func (s SiteList) SetLoading(loading bool) SiteList {
	s.loading = loading
	return s
}

// SetError sets an error message to display in the panel.
func (s SiteList) SetError(err string) SiteList {
	s.errMsg = err
	s.loading = false
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

// FilterActive reports whether the filter input is currently active.
func (s SiteList) FilterActive() bool {
	return s.filterActive
}

// visibleList returns the sites to display (filtered or all).
func (s SiteList) visibleList() []forge.Site {
	if len(s.filtered) == 0 && s.filterText == "" {
		return s.sites
	}
	result := make([]forge.Site, 0, len(s.filtered))
	for _, idx := range s.filtered {
		if idx < len(s.sites) {
			result = append(result, s.sites[idx])
		}
	}
	return result
}

// refilter rebuilds the filtered index list based on the current filter text.
func (s *SiteList) refilter() {
	text := strings.ToLower(s.filterText)
	if text == "" {
		// Show all sites.
		s.filtered = make([]int, len(s.sites))
		for i := range s.sites {
			s.filtered[i] = i
		}
		return
	}
	s.filtered = nil
	for i, site := range s.sites {
		if strings.Contains(strings.ToLower(site.Name), text) {
			s.filtered = append(s.filtered, i)
		}
	}
}

// updateSelected updates the selected pointer based on the cursor and filtered list.
func (s *SiteList) updateSelected() {
	visible := s.visibleList()
	if len(visible) > 0 && s.cursor < len(visible) {
		site := visible[s.cursor]
		s.selected = &site
	} else {
		s.selected = nil
	}
}

// Update handles key events for the site list.
func (s SiteList) Update(msg tea.Msg) (Panel, tea.Cmd) {
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
func (s SiteList) handleFilterKey(msg tea.KeyPressMsg) (Panel, tea.Cmd) {
	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
		// Accept filter, hide input but keep filter text.
		s.filterActive = false
		s.filterText = s.filterInput.Value()
		s.refilter()
		s.cursor = 0
		s.updateSelected()
		if s.selected != nil {
			site := *s.selected
			return s, func() tea.Msg { return SiteSelectedMsg{Site: site} }
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
			site := *s.selected
			return s, func() tea.Msg { return SiteSelectedMsg{Site: site} }
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

func (s SiteList) handleKey(msg tea.KeyPressMsg) (Panel, tea.Cmd) {
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
				site := *s.selected
				return s, func() tea.Msg { return SiteSelectedMsg{Site: site} }
			}
		}

	case key.Matches(msg, s.up):
		if len(visible) > 0 {
			s.cursor = max(s.cursor-1, 0)
			s.updateSelected()
			if s.selected != nil {
				site := *s.selected
				return s, func() tea.Msg { return SiteSelectedMsg{Site: site} }
			}
		}

	case key.Matches(msg, s.enter):
		if s.selected != nil {
			site := *s.selected
			return s, func() tea.Msg { return SiteSelectedMsg{Site: site} }
		}

	case key.Matches(msg, s.home):
		if len(visible) > 0 {
			s.cursor = 0
			s.updateSelected()
			if s.selected != nil {
				site := *s.selected
				return s, func() tea.Msg { return SiteSelectedMsg{Site: site} }
			}
		}

	case key.Matches(msg, s.end):
		if len(visible) > 0 {
			s.cursor = len(visible) - 1
			s.updateSelected()
			if s.selected != nil {
				site := *s.selected
				return s, func() tea.Msg { return SiteSelectedMsg{Site: site} }
			}
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

	if s.loading && len(s.sites) == 0 {
		lines = append(lines, theme.LoadingStyle.Render("Loading sites..."))
	} else if s.errMsg != "" {
		lines = append(lines, theme.ErrorStatusStyle.Render(s.errMsg))
		lines = append(lines, theme.NormalItemStyle.Render("Press ctrl+r to retry"))
	} else if s.serverName == "" && len(s.sites) == 0 {
		lines = append(lines, theme.NormalItemStyle.Render("Select a server"))
	} else if len(s.sites) == 0 {
		lines = append(lines, theme.NormalItemStyle.Render("No sites on this server"))
	} else if len(visible) == 0 && s.filterText != "" {
		lines = append(lines, theme.NormalItemStyle.Render("No matching sites"))
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

		for i := startIdx; i < len(visible) && len(lines)-siteFilterLines(s) < visibleHeight; i++ {
			site := visible[i]
			name := theme.Truncate(site.Name, innerWidth-4)
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

// siteFilterLines returns the number of header lines used by the filter UI.
func siteFilterLines(s SiteList) int {
	if s.filterActive || s.filterText != "" {
		return 1
	}
	return 0
}

// HelpBindings returns the key hints for the site list.
func (s SiteList) HelpBindings() []HelpBinding {
	if s.filterActive {
		return []HelpBinding{
			{Key: "enter", Desc: "accept filter"},
			{Key: "esc", Desc: "clear filter"},
		}
	}
	return []HelpBinding{
		{Key: "j/k", Desc: "navigate"},
		{Key: "enter", Desc: "select"},
		{Key: "/", Desc: "filter"},
		{Key: "g/G", Desc: "top/bottom"},
		{Key: "esc", Desc: "back"},
		{Key: "tab", Desc: "switch panel"},
		{Key: "q", Desc: "quit"},
	}
}
