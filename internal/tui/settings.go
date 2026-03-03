package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/key"
	lipgloss "charm.land/lipgloss/v2"

	"github.com/hinkers/Phorge/internal/config"
	"github.com/hinkers/Phorge/internal/tui/theme"
)

// settingsField represents a single editable config field.
type settingsField struct {
	label   string
	value   string
	inputID string // ID used for the Input dialog
	mask    bool   // true to mask the value (e.g. API key)
}

// SettingsModal is a floating overlay for editing config values.
type SettingsModal struct {
	active bool
	cursor int
	fields []settingsField
	config *config.Config
}

// NewSettingsModal creates a new (inactive) settings modal.
func NewSettingsModal() SettingsModal {
	return SettingsModal{}
}

// Open activates the settings modal and populates fields from the config.
func (s SettingsModal) Open(cfg *config.Config) SettingsModal {
	s.active = true
	s.cursor = 0
	s.config = cfg
	s.fields = []settingsField{
		{label: "API Key", value: cfg.Forge.APIKey, inputID: "settings-api-key", mask: true},
		{label: "SSH User", value: cfg.Forge.SSHUser, inputID: "settings-ssh-user"},
		{label: "Editor", value: cfg.Editor.Command, inputID: "settings-editor"},
		{label: "Default SSH Key", value: cfg.Forge.DefaultSSHKey, inputID: "settings-default-ssh-key"},
	}
	return s
}

// Close deactivates the settings modal.
func (s SettingsModal) Close() SettingsModal {
	s.active = false
	return s
}

// Active returns whether the settings modal is currently visible.
func (s SettingsModal) Active() bool {
	return s.active
}

// Cursor returns the current cursor position.
func (s SettingsModal) Cursor() int {
	return s.cursor
}

// SelectedField returns the inputID of the currently selected field.
func (s SettingsModal) SelectedField() string {
	if s.cursor < 0 || s.cursor >= len(s.fields) {
		return ""
	}
	return s.fields[s.cursor].inputID
}

// SelectedLabel returns the label of the currently selected field.
func (s SettingsModal) SelectedLabel() string {
	if s.cursor < 0 || s.cursor >= len(s.fields) {
		return ""
	}
	return s.fields[s.cursor].label
}

// SelectedValue returns the current value of the selected field.
func (s SettingsModal) SelectedValue() string {
	if s.cursor < 0 || s.cursor >= len(s.fields) {
		return ""
	}
	return s.fields[s.cursor].value
}

// ApplyValue updates the config with a new value for the given inputID
// and returns the updated modal. Does NOT save to disk.
func (s SettingsModal) ApplyValue(inputID, value string) SettingsModal {
	switch inputID {
	case "settings-api-key":
		s.config.Forge.APIKey = value
	case "settings-ssh-user":
		s.config.Forge.SSHUser = value
	case "settings-editor":
		s.config.Editor.Command = value
	case "settings-default-ssh-key":
		s.config.Forge.DefaultSSHKey = value
	}
	// Refresh fields from config.
	for i := range s.fields {
		switch s.fields[i].inputID {
		case "settings-api-key":
			s.fields[i].value = s.config.Forge.APIKey
		case "settings-ssh-user":
			s.fields[i].value = s.config.Forge.SSHUser
		case "settings-editor":
			s.fields[i].value = s.config.Editor.Command
		case "settings-default-ssh-key":
			s.fields[i].value = s.config.Forge.DefaultSSHKey
		}
	}
	return s
}

// Update handles key events when the settings modal is active.
func (s SettingsModal) Update(msg tea.Msg) (SettingsModal, tea.Cmd) {
	if !s.active {
		return s, nil
	}

	if msg, ok := msg.(tea.KeyPressMsg); ok {
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("esc", "ctrl+o"))):
			s.active = false
			return s, nil
		case key.Matches(msg, key.NewBinding(key.WithKeys("j", "down"))):
			if s.cursor < len(s.fields)-1 {
				s.cursor++
			}
			return s, nil
		case key.Matches(msg, key.NewBinding(key.WithKeys("k", "up"))):
			if s.cursor > 0 {
				s.cursor--
			}
			return s, nil
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			// Signal to app layer to open an input dialog for this field.
			return s, func() tea.Msg {
				return settingsEditMsg{
					inputID: s.fields[s.cursor].inputID,
					label:   s.fields[s.cursor].label,
					current: s.fields[s.cursor].value,
				}
			}
		case key.Matches(msg, key.NewBinding(key.WithKeys("e"))):
			// Signal to app layer to open config in external editor.
			return s, func() tea.Msg {
				return settingsOpenEditorMsg{}
			}
		}
	}

	return s, nil
}

// settingsEditMsg signals the app to open an input dialog for a settings field.
type settingsEditMsg struct {
	inputID string
	label   string
	current string
}

// settingsOpenEditorMsg signals the app to open config.toml in the external editor.
type settingsOpenEditorMsg struct{}

// settingsEditorDoneMsg is sent when the external editor closes.
type settingsEditorDoneMsg struct {
	err error
}

// View renders the settings modal as a box suitable for overlay.
func (s SettingsModal) View(width, height int) string {
	if !s.active {
		return ""
	}

	// Style definitions.
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.ColorPrimary).
		Align(lipgloss.Center)

	labelStyle := lipgloss.NewStyle().
		Foreground(theme.ColorSubtle).
		Width(18).
		Align(lipgloss.Right)

	valueStyle := lipgloss.NewStyle().
		Foreground(theme.ColorFg)

	selectedLabelStyle := lipgloss.NewStyle().
		Foreground(theme.ColorHighlight).
		Bold(true).
		Width(18).
		Align(lipgloss.Right)

	selectedValueStyle := lipgloss.NewStyle().
		Foreground(theme.ColorHighlight).
		Bold(true)

	cursorStyle := lipgloss.NewStyle().
		Foreground(theme.ColorHighlight).
		Bold(true)

	hintStyle := lipgloss.NewStyle().
		Foreground(theme.ColorMuted).
		Align(lipgloss.Center)

	contentWidth := 54
	if width < contentWidth+6 {
		contentWidth = width - 6
	}
	if contentWidth < 30 {
		contentWidth = 30
	}

	var lines []string
	lines = append(lines, titleStyle.Width(contentWidth).Render("Settings"))
	lines = append(lines, "")

	for i, f := range s.fields {
		displayValue := f.value
		if displayValue == "" {
			displayValue = "(not set)"
		}
		if f.mask && len(displayValue) > 8 {
			displayValue = displayValue[:4] + strings.Repeat("*", len(displayValue)-8) + displayValue[len(displayValue)-4:]
		}

		var line string
		if i == s.cursor {
			line = cursorStyle.Render("> ") +
				selectedLabelStyle.Render(f.label+": ") +
				selectedValueStyle.Render(displayValue)
		} else {
			line = "  " +
				labelStyle.Render(f.label+": ") +
				valueStyle.Render(displayValue)
		}
		lines = append(lines, line)
	}

	lines = append(lines, "")
	lines = append(lines, hintStyle.Width(contentWidth).Render("enter edit  e open in editor  esc close"))

	inner := strings.Join(lines, "\n")

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.ColorPrimary).
		Padding(1, 2).
		Background(theme.ColorBg).
		Width(contentWidth + 4).
		Render(inner)
}
