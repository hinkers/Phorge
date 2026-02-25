package panels

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/key"
	lipgloss "charm.land/lipgloss/v2"

	"github.com/hinkers/Phorge/internal/tui/theme"
)

// OutputPanel is a scrollable text viewer that displays command output,
// deploy logs, or other textual content in the bottom-right area.
type OutputPanel struct {
	title   string
	content string
	scroll  int

	// Keybindings
	up   key.Binding
	down key.Binding
	home key.Binding
	end  key.Binding
	back key.Binding
}

// NewOutputPanel creates a new, empty output panel.
func NewOutputPanel() OutputPanel {
	return OutputPanel{
		up: key.NewBinding(
			key.WithKeys("k", "up"),
			key.WithHelp("k/up", "scroll up"),
		),
		down: key.NewBinding(
			key.WithKeys("j", "down"),
			key.WithHelp("j/down", "scroll down"),
		),
		home: key.NewBinding(
			key.WithKeys("g", "home"),
			key.WithHelp("g", "top"),
		),
		end: key.NewBinding(
			key.WithKeys("G", "end"),
			key.WithHelp("G", "bottom"),
		),
		back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
	}
}

// SetContent replaces the output content and scrolls to the bottom.
// View() clamps the scroll to the valid max, so 999999 just means "end".
func (o OutputPanel) SetContent(title, content string) OutputPanel {
	o.title = title
	o.content = content
	o.scroll = 999999
	return o
}

// Clear removes all content from the output panel.
func (o OutputPanel) Clear() OutputPanel {
	o.title = ""
	o.content = ""
	o.scroll = 0
	return o
}

// HasContent reports whether the panel has any content to display.
func (o OutputPanel) HasContent() bool {
	return o.content != ""
}

// Update handles key events for the output panel.
func (o OutputPanel) Update(msg tea.Msg) (Panel, tea.Cmd) {
	if msg, ok := msg.(tea.KeyPressMsg); ok {
		return o.handleKey(msg)
	}
	return o, nil
}

func (o OutputPanel) handleKey(msg tea.KeyPressMsg) (Panel, tea.Cmd) {
	switch {
	case key.Matches(msg, o.down):
		o.scroll++
		return o, nil

	case key.Matches(msg, o.up):
		if o.scroll > 0 {
			o.scroll--
		}
		return o, nil

	case key.Matches(msg, o.home):
		o.scroll = 0
		return o, nil

	case key.Matches(msg, o.end):
		// Set to a large value; View will clamp it.
		o.scroll = 999999
		return o, nil
	}

	return o, nil
}

// View renders the output panel.
func (o OutputPanel) View(width, height int, focused bool) string {
	style := theme.InactiveBorderStyle
	titleColor := theme.ColorSubtle
	if focused {
		style = theme.ActiveBorderStyle
		titleColor = theme.ColorPrimary
	}

	panelTitle := "Output"
	if o.title != "" {
		panelTitle = o.title
	}
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(titleColor).
		Render(" " + panelTitle + " ")

	innerWidth := width - 2
	innerHeight := height - 3
	if innerWidth < 0 {
		innerWidth = 0
	}
	if innerHeight < 0 {
		innerHeight = 0
	}

	var lines []string

	if o.content == "" {
		lines = append(lines, theme.NormalItemStyle.Render("No output"))
	} else {
		allLines := strings.Split(o.content, "\n")

		// Clamp scroll.
		maxScroll := len(allLines) - innerHeight
		if maxScroll < 0 {
			maxScroll = 0
		}
		scroll := o.scroll
		if scroll > maxScroll {
			scroll = maxScroll
		}

		for i := scroll; i < len(allLines) && len(lines) < innerHeight; i++ {
			line := theme.Truncate(allLines[i], innerWidth)
			lines = append(lines, theme.NormalItemStyle.Render(line))
		}
	}

	// Pad to fill.
	totalHeight := height - 2
	if totalHeight < 0 {
		totalHeight = 0
	}
	for len(lines) < totalHeight-1 {
		lines = append(lines, "")
	}

	content := strings.Join(lines, "\n")

	return style.
		Width(width - 2).
		Height(totalHeight).
		Render(title + "\n" + content)
}

// HelpBindings returns the key hints for the output panel.
func (o OutputPanel) HelpBindings() []HelpBinding {
	return []HelpBinding{
		{Key: "j/k", Desc: "scroll"},
		{Key: "g/G", Desc: "top/bottom"},
		{Key: "esc", Desc: "back"},
		{Key: "tab", Desc: "next panel"},
	}
}
