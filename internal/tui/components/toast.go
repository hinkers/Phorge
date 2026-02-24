package components

import (
	"time"

	tea "charm.land/bubbletea/v2"
)

// toastTimeoutMsg is sent when the toast auto-dismiss timer fires.
type toastTimeoutMsg struct{}

// Toast is a timed notification bar that auto-dismisses after a duration.
type Toast struct {
	Message string
	IsError bool
	Active  bool
}

// ShowToast creates an active toast and returns a tick command that will
// dismiss it after 3 seconds.
func ShowToast(message string, isError bool) (Toast, tea.Cmd) {
	t := Toast{
		Message: message,
		IsError: isError,
		Active:  true,
	}
	cmd := tea.Tick(3*time.Second, func(time.Time) tea.Msg {
		return toastTimeoutMsg{}
	})
	return t, cmd
}

// Update handles the toast timeout message.
func (t Toast) Update(msg tea.Msg) (Toast, tea.Cmd) {
	if _, ok := msg.(toastTimeoutMsg); ok {
		t.Active = false
	}
	return t, nil
}

// View renders the toast notification bar spanning the given width.
// Returns an empty string if the toast is not active.
func (t Toast) View(width int) string {
	if !t.Active {
		return ""
	}

	style := toastNormal
	if t.IsError {
		style = toastError
	}

	return style.Width(width).Render(t.Message)
}
