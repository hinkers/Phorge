package components

import (
	lipgloss "charm.land/lipgloss/v2"

	"github.com/hinke/phorge/internal/tui/theme"
)

// Dialog box style — rounded border, centered content.
var dialogBox = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(theme.ColorPrimary).
	Padding(1, 2).
	Background(theme.ColorBg)

// Dialog text style for the question/label.
var dialogText = lipgloss.NewStyle().
	Foreground(theme.ColorFg).
	Bold(true)

// Dialog hint style for key hints (e.g. "[y]es [n]o").
var dialogHint = lipgloss.NewStyle().
	Foreground(theme.ColorMuted)

// Toast styles.
var (
	toastNormal = lipgloss.NewStyle().
			Foreground(theme.ColorFg).
			Background(theme.ColorPrimary).
			Bold(true).
			Padding(0, 1)

	toastError = lipgloss.NewStyle().
			Foreground(theme.ColorFg).
			Background(theme.ColorError).
			Bold(true).
			Padding(0, 1)
)
