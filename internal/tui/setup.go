package tui

import (
	"context"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	lipgloss "charm.land/lipgloss/v2"

	"github.com/hinke/phorge/internal/config"
	"github.com/hinke/phorge/internal/forge"
	"github.com/hinke/phorge/internal/tui/theme"
)

// setupValidateMsg is returned after attempting to validate the API key.
type setupValidateMsg struct {
	user *forge.User
	err  error
}

// Setup is a standalone bubbletea model for the first-run API key setup.
// It runs before the main App when no API key is configured.
type Setup struct {
	config     *config.Config
	input      textinput.Model
	err        error
	validating bool
	done       bool
	userName   string
	width      int
	height     int
}

// NewSetup creates a new Setup model with the given configuration.
func NewSetup(cfg *config.Config) Setup {
	ti := textinput.New()
	ti.Placeholder = "paste your API key here"
	ti.Prompt = "> "
	ti.EchoMode = textinput.EchoPassword
	ti.Focus()

	return Setup{
		config: cfg,
		input:  ti,
	}
}

// Init returns no initial command.
func (s Setup) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles messages for the setup flow.
func (s Setup) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height
		return s, nil

	case tea.KeyPressMsg:
		// If setup is done (success screen), Enter exits.
		if s.done {
			if key.Matches(msg, key.NewBinding(key.WithKeys("enter"))) {
				return s, tea.Quit
			}
			return s, nil
		}

		// If currently validating, ignore key input.
		if s.validating {
			return s, nil
		}

		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			apiKey := strings.TrimSpace(s.input.Value())
			if apiKey == "" {
				s.err = nil
				return s, nil
			}
			s.validating = true
			s.err = nil
			return s, s.validateKey(apiKey)

		case key.Matches(msg, key.NewBinding(key.WithKeys("esc", "ctrl+c"))):
			return s, tea.Quit
		}

		// Delegate to the textinput for regular character input.
		var cmd tea.Cmd
		s.input, cmd = s.input.Update(msg)
		return s, cmd

	case setupValidateMsg:
		s.validating = false
		if msg.err != nil {
			s.err = msg.err
			return s, nil
		}

		// Success: save the API key to the config.
		s.config.Forge.APIKey = strings.TrimSpace(s.input.Value())
		if err := s.config.Save(); err != nil {
			s.err = err
			return s, nil
		}

		s.userName = msg.user.Name
		s.done = true
		return s, nil
	}

	return s, nil
}

// View renders the setup screen.
func (s Setup) View() tea.View {
	var content string

	if s.done {
		content = s.viewSuccess()
	} else {
		content = s.viewInput()
	}

	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

// viewInput renders the API key input screen.
func (s Setup) viewInput() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.ColorPrimary)

	subtitleStyle := lipgloss.NewStyle().
		Foreground(theme.ColorFg)

	hintStyle := lipgloss.NewStyle().
		Foreground(theme.ColorMuted)

	errorStyle := lipgloss.NewStyle().
		Foreground(theme.ColorError).
		Bold(true)

	var lines []string
	lines = append(lines, "")
	lines = append(lines, titleStyle.Render("  Welcome to Phorge"))
	lines = append(lines, "")
	lines = append(lines, subtitleStyle.Render("  Laravel Forge TUI"))
	lines = append(lines, "")

	if s.validating {
		lines = append(lines, hintStyle.Render("  Validating API key..."))
	} else {
		lines = append(lines, subtitleStyle.Render("  Enter your Forge API key:"))
		lines = append(lines, "  "+s.input.View())
	}

	lines = append(lines, "")

	if s.err != nil {
		lines = append(lines, errorStyle.Render("  "+s.err.Error()))
		lines = append(lines, "")
	}

	lines = append(lines, hintStyle.Render("  Get your key from:"))
	lines = append(lines, hintStyle.Render("  forge.laravel.com/user/profile"))
	lines = append(lines, "")

	inner := strings.Join(lines, "\n")

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.ColorPrimary).
		Padding(0, 2).
		Width(40).
		Render(inner)

	return s.center(box)
}

// viewSuccess renders the success screen after validation.
func (s Setup) viewSuccess() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.ColorSecondary)

	subtitleStyle := lipgloss.NewStyle().
		Foreground(theme.ColorFg)

	hintStyle := lipgloss.NewStyle().
		Foreground(theme.ColorMuted)

	configPath := config.DefaultPath()

	var lines []string
	lines = append(lines, "")
	lines = append(lines, titleStyle.Render("  Welcome, "+s.userName+"!"))
	lines = append(lines, "")
	lines = append(lines, subtitleStyle.Render("  Config saved to:"))
	lines = append(lines, hintStyle.Render("  "+configPath))
	lines = append(lines, "")
	lines = append(lines, hintStyle.Render("  Press Enter to continue..."))
	lines = append(lines, "")

	inner := strings.Join(lines, "\n")

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.ColorSecondary).
		Padding(0, 2).
		Width(40).
		Render(inner)

	return s.center(box)
}

// center places content in the center of the terminal.
func (s Setup) center(box string) string {
	boxH := lipgloss.Height(box)
	boxW := lipgloss.Width(box)

	topPad := (s.height - boxH) / 2
	if topPad < 0 {
		topPad = 0
	}

	leftPad := (s.width - boxW) / 2
	if leftPad < 0 {
		leftPad = 0
	}

	// Build centered output line by line.
	boxLines := strings.Split(box, "\n")
	var out strings.Builder
	for i := 0; i < topPad; i++ {
		out.WriteString("\n")
	}
	for _, line := range boxLines {
		out.WriteString(strings.Repeat(" ", leftPad))
		out.WriteString(line)
		out.WriteString("\n")
	}

	return out.String()
}

// validateKey creates a command that validates the API key by calling the Forge API.
func (s Setup) validateKey(apiKey string) tea.Cmd {
	return func() tea.Msg {
		client := forge.NewClient(apiKey)
		user, err := client.Servers.GetUser(context.Background())
		return setupValidateMsg{user: user, err: err}
	}
}
