// Package theme provides shared colours and styles used across TUI packages.
// Extracting these into a standalone package avoids circular imports between
// the root tui package and its sub-packages (panels, components, etc.).
package theme

import lipgloss "charm.land/lipgloss/v2"

// Colour palette — loosely inspired by the lazygit theme.
var (
	ColorPrimary   = lipgloss.Color("#7aa2f7") // blue
	ColorSecondary = lipgloss.Color("#9ece6a") // green
	ColorSubtle    = lipgloss.Color("#565f89") // grey
	ColorHighlight = lipgloss.Color("#e0af68") // amber
	ColorError     = lipgloss.Color("#f7768e") // red
	ColorFg        = lipgloss.Color("#c0caf5") // light fg
	ColorMuted     = lipgloss.Color("#545c7e") // muted fg
	ColorBg        = lipgloss.Color("#1a1b26") // dark bg
)

// Panel border styles.
var (
	ActiveBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorPrimary)

	InactiveBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorSubtle)
)

// Title style for panel headers.
var TitleStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(ColorPrimary).
	Padding(0, 1)

// Help bar styles.
var (
	HelpBarBg = lipgloss.Color("#24283b") // slightly lighter than bg

	HelpBarStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Background(HelpBarBg)

	HelpKeyStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorHighlight).
			Background(HelpBarBg)
)

// List item styles.
var (
	SelectedItemStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorSecondary)

	NormalItemStyle = lipgloss.NewStyle().
			Foreground(ColorFg)

	CursorStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary).
			Bold(true)
)

// Filter indicator style (shown when a filter is active but the input is hidden).
var FilterIndicatorStyle = lipgloss.NewStyle().
	Foreground(ColorHighlight).
	Italic(true)

// Status indicator styles.
var (
	ActiveStatusStyle = lipgloss.NewStyle().
				Foreground(ColorSecondary).
				Bold(true)

	ErrorStatusStyle = lipgloss.NewStyle().
				Foreground(ColorError).
				Bold(true)

	LoadingStyle = lipgloss.NewStyle().
			Foreground(ColorHighlight).
			Italic(true)
)

// Toast styles.
var (
	ToastStyle = lipgloss.NewStyle().
			Foreground(ColorFg).
			Background(ColorPrimary).
			Bold(true).
			Padding(0, 1)

	ToastErrorStyle = lipgloss.NewStyle().
				Foreground(ColorFg).
				Background(ColorError).
				Bold(true).
				Padding(0, 1)
)

// Detail panel label/value styles.
var (
	LabelStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Width(16)

	ValueStyle = lipgloss.NewStyle().
			Foreground(ColorFg)
)

// Truncate shortens a string to fit within the given width, accounting for
// ANSI escape sequences by using lipgloss.Width for measurement.
func Truncate(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	w := lipgloss.Width(s)
	if w <= maxWidth {
		return s
	}
	// Brute-force truncation: trim runes until we fit.
	runes := []rune(s)
	for len(runes) > 0 {
		runes = runes[:len(runes)-1]
		candidate := string(runes) + "..."
		if lipgloss.Width(candidate) <= maxWidth {
			return candidate
		}
	}
	return ""
}
