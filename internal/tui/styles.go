package tui

// This file re-exports styles and colours from the shared theme package so that
// existing code within the tui package continues to compile unchanged.  New
// code should import theme directly when possible.

import "github.com/hinkers/Phorge/internal/tui/theme"

// Colour aliases.
var (
	colorPrimary   = theme.ColorPrimary
	colorSecondary = theme.ColorSecondary
	colorSubtle    = theme.ColorSubtle
	colorHighlight = theme.ColorHighlight
	colorError     = theme.ColorError
	colorFg        = theme.ColorFg
	colorMuted     = theme.ColorMuted
)

// Panel border styles.
var (
	ActiveBorderStyle   = theme.ActiveBorderStyle
	InactiveBorderStyle = theme.InactiveBorderStyle
)

// Title style for panel headers.
var TitleStyle = theme.TitleStyle

// Help bar styles.
var (
	HelpBarStyle = theme.HelpBarStyle
	HelpKeyStyle = theme.HelpKeyStyle
)

// List item styles.
var (
	SelectedItemStyle = theme.SelectedItemStyle
	NormalItemStyle   = theme.NormalItemStyle
	CursorStyle       = theme.CursorStyle
)

// Status indicator styles.
var (
	ActiveStatusStyle = theme.ActiveStatusStyle
	ErrorStatusStyle  = theme.ErrorStatusStyle
	LoadingStyle      = theme.LoadingStyle
)

// Toast styles.
var (
	ToastStyle      = theme.ToastStyle
	ToastErrorStyle = theme.ToastErrorStyle
)

// Detail panel label/value styles.
var (
	LabelStyle = theme.LabelStyle
	ValueStyle = theme.ValueStyle
)
