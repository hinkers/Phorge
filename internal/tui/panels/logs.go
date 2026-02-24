package panels

import (
	"context"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/key"
	lipgloss "charm.land/lipgloss/v2"

	"github.com/hinkers/Phorge/internal/forge"
	"github.com/hinkers/Phorge/internal/tui/theme"
)

// --- Messages ---

// LogsLoadedMsg is sent when log content has been fetched.
type LogsLoadedMsg struct {
	Content string
}

// LogsPanel shows log content in a scrollable viewport.
// If siteID > 0 it shows site logs, otherwise server logs.
type LogsPanel struct {
	client   *forge.Client
	serverID int64
	siteID   int64

	content string
	scrollY int
	loading bool

	// Keybindings
	up      key.Binding
	down    key.Binding
	refresh key.Binding
	home    key.Binding
	end     key.Binding
}

// NewLogsPanel creates a new LogsPanel.
func NewLogsPanel(client *forge.Client, serverID, siteID int64) LogsPanel {
	return LogsPanel{
		client:   client,
		serverID: serverID,
		siteID:   siteID,
		loading:  true,
		up: key.NewBinding(
			key.WithKeys("k", "up"),
			key.WithHelp("k/up", "scroll up"),
		),
		down: key.NewBinding(
			key.WithKeys("j", "down"),
			key.WithHelp("j/down", "scroll down"),
		),
		refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
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

// LoadLogs returns a tea.Cmd that fetches log content.
func (p LogsPanel) LoadLogs() tea.Cmd {
	client := p.client
	serverID := p.serverID
	siteID := p.siteID
	return func() tea.Msg {
		var content string
		var err error
		if siteID > 0 {
			content, err = client.Logs.GetSiteLog(context.Background(), serverID, siteID)
		} else {
			content, err = client.Logs.GetServerLog(context.Background(), serverID)
		}
		if err != nil {
			return PanelErrMsg{Err: err}
		}
		return LogsLoadedMsg{Content: content}
	}
}

// Update handles messages for the logs panel.
func (p LogsPanel) Update(msg tea.Msg) (Panel, tea.Cmd) {
	switch msg := msg.(type) {
	case LogsLoadedMsg:
		p.content = msg.Content
		p.loading = false
		p.scrollY = 0
		return p, nil

	case tea.KeyPressMsg:
		return p.handleKey(msg)
	}

	return p, nil
}

func (p LogsPanel) handleKey(msg tea.KeyPressMsg) (Panel, tea.Cmd) {
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

	case key.Matches(msg, p.refresh):
		p.loading = true
		return p, p.LoadLogs()
	}

	return p, nil
}

// View renders the logs panel.
func (p LogsPanel) View(width, height int, focused bool) string {
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

	titleText := " Logs "
	if p.siteID > 0 {
		titleText = " Site Logs "
	} else {
		titleText = " Server Logs "
	}

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(titleColor).
		Render(titleText)

	content := p.renderContent(innerWidth, innerHeight-1) // -1 for title line

	return style.
		Width(innerWidth).
		Height(innerHeight).
		Render(title + "\n" + content)
}

// renderContent renders the log content with scrolling.
func (p LogsPanel) renderContent(width, height int) string {
	if height < 1 {
		height = 1
	}

	if p.loading {
		return theme.LoadingStyle.Render("Loading logs...")
	}

	if p.content == "" {
		return theme.NormalItemStyle.Render("No log content available")
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

// HelpBindings returns the key hints for the logs panel.
func (p LogsPanel) HelpBindings() []HelpBinding {
	return []HelpBinding{
		{Key: "j/k", Desc: "scroll"},
		{Key: "g/G", Desc: "top/bottom"},
		{Key: "r", Desc: "refresh"},
		{Key: "esc", Desc: "back"},
		{Key: "tab", Desc: "switch panel"},
		{Key: "q", Desc: "quit"},
	}
}
