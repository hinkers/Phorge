package components

import lipgloss "charm.land/lipgloss/v2"

// Colour palette — matches the parent tui package theme.
var (
	colorPrimary = lipgloss.Color("#7aa2f7") // blue
	colorError   = lipgloss.Color("#f7768e") // red
	colorFg      = lipgloss.Color("#c0caf5") // light fg
	colorMuted   = lipgloss.Color("#545c7e") // muted fg
	colorBg      = lipgloss.Color("#1a1b26") // dark bg
)

// Dialog box style — rounded border, centered content.
var dialogBox = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(colorPrimary).
	Padding(1, 2).
	Background(colorBg)

// Dialog text style for the question/label.
var dialogText = lipgloss.NewStyle().
	Foreground(colorFg).
	Bold(true)

// Dialog hint style for key hints (e.g. "[y]es [n]o").
var dialogHint = lipgloss.NewStyle().
	Foreground(colorMuted)

// Toast styles.
var (
	toastNormal = lipgloss.NewStyle().
			Foreground(colorFg).
			Background(colorPrimary).
			Bold(true).
			Padding(0, 1)

	toastError = lipgloss.NewStyle().
			Foreground(colorFg).
			Background(colorError).
			Bold(true).
			Padding(0, 1)
)
