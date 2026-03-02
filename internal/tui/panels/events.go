package panels

import (
	"context"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/key"
	lipgloss "charm.land/lipgloss/v2"

	"github.com/hinkers/Phorge/internal/forge"
	"github.com/hinkers/Phorge/internal/tui/theme"
)

// --- Messages ---

// EventsLoadedMsg is sent when the server events have been fetched.
type EventsLoadedMsg struct {
	Events []forge.Event
}

// EventsPanel shows the event history for a server.
type EventsPanel struct {
	client   *forge.Client
	serverID int64

	events  []forge.Event
	cursor  int
	loading bool

	// Keybindings
	up   key.Binding
	down key.Binding
	back key.Binding
	home key.Binding
	end  key.Binding
}

// NewEventsPanel creates a new EventsPanel. Call LoadEvents() to fetch data.
func NewEventsPanel(client *forge.Client, serverID int64) EventsPanel {
	return EventsPanel{
		client:   client,
		serverID: serverID,
		loading:  true,
		up: key.NewBinding(
			key.WithKeys("k", "up"),
			key.WithHelp("k/up", "up"),
		),
		down: key.NewBinding(
			key.WithKeys("j", "down"),
			key.WithHelp("j/down", "down"),
		),
		back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
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

// LoadEvents returns a tea.Cmd that fetches the server events.
func (p EventsPanel) LoadEvents() tea.Cmd {
	client := p.client
	serverID := p.serverID
	return func() tea.Msg {
		events, err := client.Events.List(context.Background(), serverID)
		if err != nil {
			return PanelErrMsg{Err: err}
		}
		return EventsLoadedMsg{Events: events}
	}
}

// Update handles messages for the events panel.
func (p EventsPanel) Update(msg tea.Msg) (Panel, tea.Cmd) {
	switch msg := msg.(type) {
	case EventsLoadedMsg:
		p.events = msg.Events
		p.loading = false
		p.cursor = 0
		return p, nil

	case tea.KeyPressMsg:
		return p.handleKey(msg)
	}

	return p, nil
}

// handleKey processes key events for the events panel.
func (p EventsPanel) handleKey(msg tea.KeyPressMsg) (Panel, tea.Cmd) {
	switch {
	case key.Matches(msg, p.down):
		if len(p.events) > 0 {
			p.cursor = min(p.cursor+1, len(p.events)-1)
		}
		return p, nil

	case key.Matches(msg, p.up):
		if len(p.events) > 0 {
			p.cursor = max(p.cursor-1, 0)
		}
		return p, nil

	case key.Matches(msg, p.home):
		p.cursor = 0
		return p, nil

	case key.Matches(msg, p.end):
		if len(p.events) > 0 {
			p.cursor = len(p.events) - 1
		}
		return p, nil
	}

	return p, nil
}

// View renders the events panel.
func (p EventsPanel) View(width, height int, focused bool) string {
	style := theme.InactiveBorderStyle
	titleColor := theme.ColorSubtle
	if focused {
		style = theme.ActiveBorderStyle
		titleColor = theme.ColorPrimary
	}

	innerWidth := width - 2
	innerHeight := height - 2
	if innerWidth < 0 {
		innerWidth = 0
	}
	if innerHeight < 0 {
		innerHeight = 0
	}

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(titleColor).
		Render(" Events ")
	content := p.renderList(innerWidth, innerHeight-1)

	return style.
		Width(innerWidth).
		Height(innerHeight).
		Render(title + "\n" + content)
}

// Column widths for the events table.
const (
	colEventTimeWidth = 8
	colEventUserWidth = 10
)

// eventTableOverhead is the fixed character budget for non-description columns:
// cursor(2) + time(8) + 2 spacers(4) + user(10) + border buffer(4).
const eventTableOverhead = 2 + colEventTimeWidth + 2 + colEventUserWidth + 2 + 4

// eventDescWidth returns the space available for the description column.
func eventDescWidth(maxWidth int) int {
	w := maxWidth - eventTableOverhead
	if w < 10 {
		w = 10
	}
	return w
}

// renderList renders the event list view.
func (p EventsPanel) renderList(width, height int) string {
	var lines []string

	if p.loading && len(p.events) == 0 {
		lines = append(lines, theme.LoadingStyle.Render("Loading events..."))
	} else if len(p.events) == 0 {
		lines = append(lines, theme.NormalItemStyle.Render("No events found"))
	} else {
		// Render table header.
		lines = append(lines, p.renderHeader(width))

		// Calculate visible range with scrolling.
		// Reserve 1 for the header row.
		visibleHeight := height - 2
		if visibleHeight < 1 {
			visibleHeight = 1
		}
		startIdx := 0
		if p.cursor >= visibleHeight {
			startIdx = p.cursor - visibleHeight + 1
		}

		for i := startIdx; i < len(p.events) && len(lines)-1 < visibleHeight; i++ {
			evt := p.events[i]
			line := p.renderEventLine(evt, i, width)
			lines = append(lines, line)
		}
	}

	// Pad to fill the panel height.
	for len(lines) < height {
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}

// renderHeader renders the column header row.
func (p EventsPanel) renderHeader(maxWidth int) string {
	descWidth := eventDescWidth(maxWidth)

	line := fmt.Sprintf("  %-*s  %-*s  %-*s",
		colEventTimeWidth, "TIME",
		colEventUserWidth, "USER",
		descWidth, "DESCRIPTION",
	)
	return theme.Truncate(headerStyle.Render(line), maxWidth)
}

// renderEventLine renders a single event entry as a table row.
func (p EventsPanel) renderEventLine(evt forge.Event, idx, maxWidth int) string {
	// Time.
	timeStr := relativeTime(evt.CreatedAt)
	if timeStr == "" {
		timeStr = "-"
	}

	// User (ran_as).
	user := evt.RanAs
	if user == "" {
		user = "-"
	}

	// Description.
	desc := evt.Description
	if desc == "" {
		desc = "-"
	}
	desc = strings.ReplaceAll(desc, "\n", " ")

	descWidth := eventDescWidth(maxWidth)
	desc = truncatePlain(desc, descWidth)

	timeStr = fmt.Sprintf("%-*s", colEventTimeWidth, timeStr)
	userStr := fmt.Sprintf("%-*s", colEventUserWidth, truncatePlain(user, colEventUserWidth))

	if idx == p.cursor {
		line := theme.CursorStyle.Render("> ") +
			theme.NormalItemStyle.Render(timeStr) +
			"  " + theme.NormalItemStyle.Render(userStr) +
			"  " + theme.SelectedItemStyle.Render(fmt.Sprintf("%-*s", descWidth, desc))
		return theme.Truncate(line, maxWidth)
	}

	line := "  " +
		theme.NormalItemStyle.Render(timeStr) +
		"  " + theme.NormalItemStyle.Render(userStr) +
		"  " + theme.NormalItemStyle.Render(fmt.Sprintf("%-*s", descWidth, desc))
	return theme.Truncate(line, maxWidth)
}

// HelpBindings returns the key hints for the events panel.
func (p EventsPanel) HelpBindings() []HelpBinding {
	return []HelpBinding{
		{Key: "j/k", Desc: "navigate"},
		{Key: "g/G", Desc: "top/bottom"},
		{Key: "esc", Desc: "back"},
		{Key: "tab", Desc: "next panel"},
	}
}
