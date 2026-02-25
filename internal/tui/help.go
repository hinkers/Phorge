package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/key"
	lipgloss "charm.land/lipgloss/v2"

	"github.com/hinkers/Phorge/internal/tui/theme"
)

// helpSection groups keybindings under a section heading.
type helpSection struct {
	title    string
	bindings []helpEntry
}

// helpEntry is a single key-description pair.
type helpEntry struct {
	key  string
	desc string
}

// HelpModal is a full-screen overlay showing all keybindings.
type HelpModal struct {
	active  bool
	scrollY int
	height  int
}

// NewHelpModal creates a new (inactive) help modal.
func NewHelpModal() HelpModal {
	return HelpModal{}
}

// Toggle switches the help modal on or off.
func (h HelpModal) Toggle() HelpModal {
	h.active = !h.active
	if h.active {
		h.scrollY = 0
	}
	return h
}

// Active returns whether the help modal is currently visible.
func (h HelpModal) Active() bool {
	return h.active
}

// Update handles key events when the help modal is active.
func (h HelpModal) Update(msg tea.Msg) (HelpModal, tea.Cmd) {
	if !h.active {
		return h, nil
	}

	if msg, ok := msg.(tea.KeyPressMsg); ok {
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("esc", "?", "q"))):
			h.active = false
			return h, nil
		case key.Matches(msg, key.NewBinding(key.WithKeys("j", "down"))):
			h.scrollY++
			return h, nil
		case key.Matches(msg, key.NewBinding(key.WithKeys("k", "up"))):
			if h.scrollY > 0 {
				h.scrollY--
			}
			return h, nil
		case key.Matches(msg, key.NewBinding(key.WithKeys("g", "home"))):
			h.scrollY = 0
			return h, nil
		case key.Matches(msg, key.NewBinding(key.WithKeys("G", "end"))):
			// Scroll to bottom — will be clamped in View.
			h.scrollY = 999
			return h, nil
		case key.Matches(msg, key.NewBinding(key.WithKeys("pgdown", "ctrl+d"))):
			h.scrollY += 10
			return h, nil
		case key.Matches(msg, key.NewBinding(key.WithKeys("pgup", "ctrl+u"))):
			h.scrollY -= 10
			if h.scrollY < 0 {
				h.scrollY = 0
			}
			return h, nil
		}
	}

	return h, nil
}

// View renders the help modal as a box suitable for overlay on the existing UI.
func (h HelpModal) View(width, height int) string {
	if !h.active {
		return ""
	}

	h.height = height

	sections := helpSections()

	// Style definitions.
	sectionStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.ColorPrimary)

	keyStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.ColorHighlight).
		Width(14).
		Align(lipgloss.Right)

	descStyle := lipgloss.NewStyle().
		Foreground(theme.ColorFg)

	separatorStyle := lipgloss.NewStyle().
		Foreground(theme.ColorSubtle)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.ColorPrimary).
		Align(lipgloss.Center)

	hintStyle := lipgloss.NewStyle().
		Foreground(theme.ColorMuted).
		Align(lipgloss.Center)

	// Content width for the help box.
	contentWidth := 44
	if width < contentWidth+6 {
		contentWidth = width - 6
	}
	if contentWidth < 20 {
		contentWidth = 20
	}

	// Build all lines.
	var lines []string
	lines = append(lines, titleStyle.Width(contentWidth).Render("Keybindings"))
	lines = append(lines, "")

	for i, section := range sections {
		// Section separator.
		titleLen := len(section.title) + 2
		dashCount := contentWidth - titleLen
		if dashCount < 2 {
			dashCount = 2
		}
		header := separatorStyle.Render(strings.Repeat("─", 1)+" ") +
			sectionStyle.Render(section.title) +
			separatorStyle.Render(" "+strings.Repeat("─", dashCount-1))
		lines = append(lines, header)

		for _, entry := range section.bindings {
			line := keyStyle.Render(entry.key) + "  " + descStyle.Render(entry.desc)
			lines = append(lines, line)
		}

		if i < len(sections)-1 {
			lines = append(lines, "")
		}
	}

	lines = append(lines, "")
	lines = append(lines, hintStyle.Width(contentWidth).Render("esc/? close  j/k scroll"))

	totalLines := len(lines)

	// Available height for content inside the box (border + padding takes space).
	boxPadding := 2 // top + bottom padding
	boxBorder := 2  // top + bottom border
	availLines := height - boxPadding - boxBorder - 4 // margin for overlay
	if availLines < 5 {
		availLines = 5
	}

	// Clamp scroll.
	maxScroll := totalLines - availLines
	if maxScroll < 0 {
		maxScroll = 0
	}
	if h.scrollY > maxScroll {
		h.scrollY = maxScroll
	}

	// Slice visible lines.
	start := h.scrollY
	end := start + availLines
	if end > totalLines {
		end = totalLines
	}
	visibleLines := lines[start:end]

	// Add scroll indicators.
	if h.scrollY > 0 {
		indicator := hintStyle.Width(contentWidth).Render(fmt.Sprintf("  (%d more above)", h.scrollY))
		visibleLines = append([]string{indicator}, visibleLines...)
	}
	if end < totalLines {
		indicator := hintStyle.Width(contentWidth).Render(fmt.Sprintf("  (%d more below)", totalLines-end))
		visibleLines = append(visibleLines, indicator)
	}

	inner := strings.Join(visibleLines, "\n")

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.ColorPrimary).
		Padding(1, 2).
		Background(theme.ColorBg).
		Width(contentWidth + 4). // add padding
		Render(inner)
}

// helpSections returns all help sections with their keybindings.
func helpSections() []helpSection {
	return []helpSection{
		{
			title: "Tree Panel",
			bindings: []helpEntry{
				{"j/k", "Move up/down"},
				{"h/l", "Collapse/expand node"},
				{"g/G", "Top/bottom"},
				{"Enter", "Select → detail panel"},
				{"Space", "Expand/collapse server"},
				{"/", "Filter servers & sites"},
				{"Esc", "Clear filter"},
			},
		},
		{
			title: "Navigation",
			bindings: []helpEntry{
				{"Tab", "Next panel"},
				{"Shift+Tab", "Previous panel"},
				{"Esc", "Go back"},
			},
		},
		{
			title: "Output Panel",
			bindings: []helpEntry{
				{"j/k", "Scroll up/down"},
				{"g/G", "Top/bottom"},
				{"Esc", "Back to detail"},
			},
		},
		{
			title: "Global",
			bindings: []helpEntry{
				{"Ctrl+S", "SSH to server"},
				{"Ctrl+F", "SFTP via termscp"},
				{"Ctrl+D", "Database tunnel"},
				{"Ctrl+R", "Refresh"},
				{"?", "Toggle help"},
				{"q", "Quit"},
			},
		},
		{
			title: "Server Actions",
			bindings: []helpEntry{
				{"s", "SSH"},
				{"f", "SFTP"},
				{"r", "Reboot server"},
				{"D", "Set/clear default"},
			},
		},
		{
			title: "Site Actions",
			bindings: []helpEntry{
				{"d", "Deploy"},
				{"e", "Edit env/script"},
				{"s", "SSH"},
				{"D", "Set/clear default"},
				{"l", "View logs"},
			},
		},
		{
			title: "Section Tabs",
			bindings: []helpEntry{
				{"1", "Deployments"},
				{"2", "Environment"},
				{"3", "Databases"},
				{"4", "SSL"},
				{"5", "Workers"},
				{"6", "Commands/Daemons"},
				{"7", "Logs/Firewall"},
				{"8", "Git/Jobs"},
				{"9", "Domains/SSH Keys"},
			},
		},
		{
			title: "Panel Actions",
			bindings: []helpEntry{
				{"c", "Create new"},
				{"x", "Delete"},
				{"a", "Add/activate"},
				{"r", "Restart"},
				{"u", "Users (databases)"},
				{"S", "Deploy script"},
			},
		},
	}
}
