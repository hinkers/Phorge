// Package panels provides the Panel interface and concrete panel
// implementations for the three-panel TUI layout.
package panels

import tea "charm.land/bubbletea/v2"

// Panel is the interface all detail/context panels implement.
type Panel interface {
	// Update handles messages and returns the updated panel plus any command.
	Update(msg tea.Msg) (Panel, tea.Cmd)

	// View renders the panel into a string that fits within the given
	// dimensions. The focused flag controls whether the panel draws an
	// active (highlighted) or inactive border.
	View(width, height int, focused bool) string

	// HelpBindings returns the context-sensitive key hints to display in
	// the help bar when this panel is focused.
	HelpBindings() []HelpBinding
}

// HelpBinding pairs a key label with a short description for the help bar.
type HelpBinding struct {
	Key  string
	Desc string
}
