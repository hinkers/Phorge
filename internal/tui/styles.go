package tui

import (
	lipgloss "charm.land/lipgloss/v2"
)

// Colour palette — loosely inspired by the lazygit theme.
var (
	colorPrimary   = lipgloss.Color("#7aa2f7") // blue
	colorSecondary = lipgloss.Color("#9ece6a") // green
	colorSubtle    = lipgloss.Color("#565f89") // grey
	colorHighlight = lipgloss.Color("#e0af68") // amber
	colorError     = lipgloss.Color("#f7768e") // red
	colorFg        = lipgloss.Color("#c0caf5") // light fg
	colorMuted     = lipgloss.Color("#545c7e") // muted fg
)

// Panel border styles.
var (
	ActiveBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorPrimary)

	InactiveBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorSubtle)
)

// Title style for panel headers.
var TitleStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(colorPrimary).
	Padding(0, 1)

// Help bar styles.
var (
	HelpBarStyle = lipgloss.NewStyle().
			Foreground(colorMuted)

	HelpKeyStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorHighlight)
)

// List item styles.
var (
	SelectedItemStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorSecondary)

	NormalItemStyle = lipgloss.NewStyle().
			Foreground(colorFg)

	CursorStyle = lipgloss.NewStyle().
			Foreground(colorSecondary).
			Bold(true)
)

// Status indicator styles.
var (
	ActiveStatusStyle = lipgloss.NewStyle().
				Foreground(colorSecondary).
				Bold(true)

	ErrorStatusStyle = lipgloss.NewStyle().
				Foreground(colorError).
				Bold(true)

	LoadingStyle = lipgloss.NewStyle().
			Foreground(colorHighlight).
			Italic(true)
)

// Toast styles.
var (
	ToastStyle = lipgloss.NewStyle().
			Foreground(colorFg).
			Background(colorPrimary).
			Bold(true).
			Padding(0, 1)

	ToastErrorStyle = lipgloss.NewStyle().
			Foreground(colorFg).
			Background(colorError).
			Bold(true).
			Padding(0, 1)
)

// Detail panel label/value styles.
var (
	LabelStyle = lipgloss.NewStyle().
			Foreground(colorMuted).
			Width(16)

	ValueStyle = lipgloss.NewStyle().
			Foreground(colorFg)
)
