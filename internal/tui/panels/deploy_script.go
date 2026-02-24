package panels

import (
	"context"
	"os"
	"os/exec"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/key"
	lipgloss "charm.land/lipgloss/v2"

	"github.com/hinke/phorge/internal/forge"
	"github.com/hinke/phorge/internal/tui/theme"
)

// --- Messages ---

// ScriptLoadedMsg is sent when the deployment script has been fetched.
type ScriptLoadedMsg struct {
	Content string
}

// ScriptSavedMsg is sent after the deployment script has been uploaded.
type ScriptSavedMsg struct {
	Err error
}

// scriptEditorDoneMsg is sent after the external editor exits for the deploy script.
type scriptEditorDoneMsg struct {
	newContent string
	changed    bool
	err        error
}

// DeployScriptPanel shows the deployment script with option to edit in an
// external editor.
type DeployScriptPanel struct {
	client   *forge.Client
	serverID int64
	siteID   int64

	content string // the script text
	scrollY int    // scroll offset (line)
	loading bool
	editor  string // editor command from config

	// Keybindings
	up   key.Binding
	down key.Binding
	edit key.Binding
	back key.Binding
	home key.Binding
	end  key.Binding
}

// NewDeployScriptPanel creates a new DeployScriptPanel. Call LoadScript() to
// kick off the initial data fetch.
func NewDeployScriptPanel(client *forge.Client, serverID, siteID int64, editor string) DeployScriptPanel {
	if editor == "" {
		editor = "vim"
	}
	return DeployScriptPanel{
		client:   client,
		serverID: serverID,
		siteID:   siteID,
		loading:  true,
		editor:   editor,
		up: key.NewBinding(
			key.WithKeys("k", "up"),
			key.WithHelp("k/up", "scroll up"),
		),
		down: key.NewBinding(
			key.WithKeys("j", "down"),
			key.WithHelp("j/down", "scroll down"),
		),
		edit: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "edit"),
		),
		back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
		home: key.NewBinding(
			key.WithKeys("g", "home"),
			key.WithHelp("g", "top"),
		),
		end: key.NewBinding(
			key.WithKeys("G", "end"),
			key.WithHelp("G", "bottom"),
		),
	}
}

// LoadScript returns a tea.Cmd that fetches the deployment script.
func (p DeployScriptPanel) LoadScript() tea.Cmd {
	client := p.client
	serverID := p.serverID
	siteID := p.siteID
	return func() tea.Msg {
		content, err := client.Deployments.GetScript(context.Background(), serverID, siteID)
		if err != nil {
			return PanelErrMsg{Err: err}
		}
		return ScriptLoadedMsg{Content: content}
	}
}

// saveScript returns a tea.Cmd that uploads the deployment script.
func (p DeployScriptPanel) saveScript(content string) tea.Cmd {
	client := p.client
	serverID := p.serverID
	siteID := p.siteID
	return func() tea.Msg {
		err := client.Deployments.UpdateScript(context.Background(), serverID, siteID, content)
		return ScriptSavedMsg{Err: err}
	}
}

// Update handles messages for the deploy script panel.
func (p DeployScriptPanel) Update(msg tea.Msg) (Panel, tea.Cmd) {
	switch msg := msg.(type) {
	case ScriptLoadedMsg:
		p.content = msg.Content
		p.loading = false
		p.scrollY = 0
		return p, nil

	case scriptEditorDoneMsg:
		if msg.err != nil {
			return p, func() tea.Msg {
				return PanelErrMsg{Err: msg.err}
			}
		}
		if msg.changed {
			p.content = msg.newContent
			return p, p.saveScript(msg.newContent)
		}
		return p, nil

	case ScriptSavedMsg:
		// The app layer will handle showing a toast based on this message.
		return p, nil

	case tea.KeyPressMsg:
		return p.handleKey(msg)
	}

	return p, nil
}

// handleKey processes key events for the deploy script panel.
func (p DeployScriptPanel) handleKey(msg tea.KeyPressMsg) (Panel, tea.Cmd) {
	switch {
	case key.Matches(msg, p.down):
		p.scrollY++
		return p, nil

	case key.Matches(msg, p.up):
		if p.scrollY > 0 {
			p.scrollY--
		}
		return p, nil

	case key.Matches(msg, p.home):
		p.scrollY = 0
		return p, nil

	case key.Matches(msg, p.end):
		lines := strings.Split(p.content, "\n")
		p.scrollY = len(lines) // will be clamped during render
		return p, nil

	case key.Matches(msg, p.edit):
		if p.loading {
			return p, nil
		}
		// Write content to temp file synchronously (fast local I/O).
		tmpFile, err := os.CreateTemp("", "phorge-deploy-*.sh")
		if err != nil {
			return p, func() tea.Msg {
				return PanelErrMsg{Err: err}
			}
		}
		if _, err := tmpFile.WriteString(p.content); err != nil {
			tmpFile.Close()
			os.Remove(tmpFile.Name())
			return p, func() tea.Msg {
				return PanelErrMsg{Err: err}
			}
		}
		tmpFile.Close()
		original := p.content
		path := tmpFile.Name()

		c := exec.Command(p.editor, path)
		return p, tea.ExecProcess(c, func(err error) tea.Msg {
			defer os.Remove(path)
			if err != nil {
				return scriptEditorDoneMsg{err: err}
			}
			newContent, readErr := os.ReadFile(path)
			if readErr != nil {
				return scriptEditorDoneMsg{err: readErr}
			}
			return scriptEditorDoneMsg{
				newContent: string(newContent),
				changed:    string(newContent) != original,
			}
		})
	}

	return p, nil
}

// View renders the deploy script panel.
func (p DeployScriptPanel) View(width, height int, focused bool) string {
	style := theme.InactiveBorderStyle
	titleColor := theme.ColorSubtle
	if focused {
		style = theme.ActiveBorderStyle
		titleColor = theme.ColorPrimary
	}

	innerWidth := width - 2
	innerHeight := height - 2
	if innerWidth < 0 {
		innerWidth = 0
	}
	if innerHeight < 0 {
		innerHeight = 0
	}

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(titleColor).
		Render(" Deploy Script ")

	content := p.renderContent(innerWidth, innerHeight-1) // -1 for title line

	return style.
		Width(innerWidth).
		Height(innerHeight).
		Render(title + "\n" + content)
}

// renderContent renders the script content with scrolling.
func (p DeployScriptPanel) renderContent(width, height int) string {
	if height < 1 {
		height = 1
	}

	if p.loading {
		return theme.LoadingStyle.Render("Loading deploy script...")
	}

	if p.content == "" {
		return theme.NormalItemStyle.Render("No deployment script found")
	}

	allLines := strings.Split(p.content, "\n")

	// Clamp scroll offset.
	maxScroll := len(allLines) - height
	if maxScroll < 0 {
		maxScroll = 0
	}
	if p.scrollY > maxScroll {
		p.scrollY = maxScroll
	}

	var lines []string
	for i := p.scrollY; i < len(allLines) && len(lines) < height; i++ {
		line := theme.Truncate(allLines[i], width)
		lines = append(lines, theme.NormalItemStyle.Render(line))
	}

	// Pad remaining height.
	for len(lines) < height {
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}

// HelpBindings returns the key hints for the deploy script panel.
func (p DeployScriptPanel) HelpBindings() []HelpBinding {
	return []HelpBinding{
		{Key: "e", Desc: "edit"},
		{Key: "j/k", Desc: "scroll"},
		{Key: "g/G", Desc: "top/bottom"},
		{Key: "esc", Desc: "back"},
		{Key: "tab", Desc: "switch panel"},
		{Key: "q", Desc: "quit"},
	}
}
