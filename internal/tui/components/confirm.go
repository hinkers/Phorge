package components

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/key"
	lipgloss "charm.land/lipgloss/v2"
)

// ConfirmResult is sent when the user resolves a confirmation dialog.
type ConfirmResult struct {
	Confirmed bool
	ID        string
}

// Confirm is a Y/N confirmation dialog overlay.
type Confirm struct {
	Question string
	ID       string
	Active   bool
}

// NewConfirm creates a new confirmation dialog.
func NewConfirm(id, question string) Confirm {
	return Confirm{
		Question: question,
		ID:       id,
		Active:   true,
	}
}

// Update handles key events for the confirmation dialog.
// y/Y confirms, n/N/Esc cancels.
func (c Confirm) Update(msg tea.Msg) (Confirm, tea.Cmd) {
	if !c.Active {
		return c, nil
	}

	if msg, ok := msg.(tea.KeyPressMsg); ok {
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("y", "Y"))):
			c.Active = false
			return c, func() tea.Msg {
				return ConfirmResult{Confirmed: true, ID: c.ID}
			}
		case key.Matches(msg, key.NewBinding(key.WithKeys("n", "N", "esc"))):
			c.Active = false
			return c, func() tea.Msg {
				return ConfirmResult{Confirmed: false, ID: c.ID}
			}
		}
	}

	return c, nil
}

// View renders the confirmation dialog centered on the screen.
// Returns an empty string if the dialog is not active.
func (c Confirm) View(width, height int) string {
	if !c.Active {
		return ""
	}

	// Build the dialog box content.
	question := dialogText.Render(c.Question)
	hint := dialogHint.Render("[y]es  [n]o")
	inner := lipgloss.JoinVertical(lipgloss.Center, "", question, "", hint, "")

	// Size the box to fit the content with padding.
	boxWidth := lipgloss.Width(inner) + 4
	if boxWidth < 30 {
		boxWidth = 30
	}
	if boxWidth > width-4 {
		boxWidth = width - 4
	}

	box := dialogBox.Width(boxWidth).Render(inner)

	// Center the box on the screen.
	boxH := lipgloss.Height(box)
	topPad := (height - boxH) / 2
	if topPad < 0 {
		topPad = 0
	}

	leftPad := (width - lipgloss.Width(box)) / 2
	if leftPad < 0 {
		leftPad = 0
	}

	padded := strings.Repeat("\n", topPad) + strings.Repeat(" ", leftPad) + box
	return padded
}
