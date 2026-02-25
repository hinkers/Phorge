package components

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	lipgloss "charm.land/lipgloss/v2"
)

// InputResult is sent when the user submits the input dialog.
type InputResult struct {
	Value string
	ID    string
}

// InputCancelled is sent when the user cancels the input dialog.
type InputCancelled struct {
	ID string
}

// Input is a text input modal overlay using the bubbles textinput widget.
type Input struct {
	Label  string
	ID     string
	Active bool
	input  textinput.Model
}

// NewInput creates a new text input dialog with the given label and placeholder.
func NewInput(id, label, placeholder string) Input {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Prompt = "  "
	ti.Focus()

	return Input{
		Label:  label,
		ID:     id,
		Active: true,
		input:  ti,
	}
}

// NewInputWide creates a text input dialog with no character limit, suitable
// for long values like file paths or SSH keys.
func NewInputWide(id, label, placeholder string) Input {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Prompt = "  "
	ti.CharLimit = 0 // unlimited
	ti.Focus()

	return Input{
		Label:  label,
		ID:     id,
		Active: true,
		input:  ti,
	}
}

// Update handles key events for the input dialog.
// Enter submits, Esc cancels, other keys are delegated to the textinput.
func (i Input) Update(msg tea.Msg) (Input, tea.Cmd) {
	if !i.Active {
		return i, nil
	}

	if msg, ok := msg.(tea.KeyPressMsg); ok {
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			i.Active = false
			value := i.input.Value()
			id := i.ID
			return i, func() tea.Msg {
				return InputResult{Value: value, ID: id}
			}
		case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
			i.Active = false
			id := i.ID
			return i, func() tea.Msg {
				return InputCancelled{ID: id}
			}
		}
	}

	// Delegate to the textinput for regular character input.
	var cmd tea.Cmd
	i.input, cmd = i.input.Update(msg)
	return i, cmd
}

// View renders the input dialog centered on the screen.
// Returns an empty string if the dialog is not active.
func (i Input) View(width, height int) string {
	if !i.Active {
		return ""
	}

	// Build the dialog content.
	label := dialogText.Render(i.Label)
	inputView := i.input.View()
	hint := dialogHint.Render("enter confirm  esc cancel")
	inner := lipgloss.JoinVertical(lipgloss.Left, "", label, "", inputView, "", hint, "")

	// Size the box to fit the content with padding.
	boxWidth := lipgloss.Width(inner) + 4
	if boxWidth < 40 {
		boxWidth = 40
	}
	if boxWidth > width-4 {
		boxWidth = width - 4
	}

	box := dialogBox.Width(boxWidth).Render(inner)

	return box
}
