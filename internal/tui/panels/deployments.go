package panels

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/key"
	lipgloss "charm.land/lipgloss/v2"

	"github.com/hinke/phorge/internal/forge"
	"github.com/hinke/phorge/internal/tui/theme"
)

// --- Messages ---

// DeploymentsLoadedMsg is sent when the deployment history has been fetched.
type DeploymentsLoadedMsg struct {
	Deployments []forge.Deployment
}

// DeployOutputMsg is sent when a deployment's output has been fetched.
type DeployOutputMsg struct {
	Output string
}

// DeployResetMsg is sent when a deployment status reset completes.
type DeployResetMsg struct {
	Err error
}

// DeploymentsPanel shows the deployment history for a site and allows
// triggering deploys, viewing output, and resetting deployment status.
type DeploymentsPanel struct {
	client   *forge.Client
	serverID int64
	siteID   int64

	deployments []forge.Deployment
	cursor      int
	loading     bool

	// Output view state.
	outputView   string
	showOutput   bool
	outputScroll int // line offset for scrolling output

	// Keybindings
	up     key.Binding
	down   key.Binding
	enter  key.Binding
	deploy key.Binding
	reset  key.Binding
	back   key.Binding
	home   key.Binding
	end    key.Binding
}

// NewDeploymentsPanel creates a new DeploymentsPanel. Call LoadDeployments()
// to kick off the initial data fetch.
func NewDeploymentsPanel(client *forge.Client, serverID, siteID int64) DeploymentsPanel {
	return DeploymentsPanel{
		client:   client,
		serverID: serverID,
		siteID:   siteID,
		loading:  true,
		up: key.NewBinding(
			key.WithKeys("k", "up"),
			key.WithHelp("k/up", "up"),
		),
		down: key.NewBinding(
			key.WithKeys("j", "down"),
			key.WithHelp("j/down", "down"),
		),
		enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "view output"),
		),
		deploy: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "deploy"),
		),
		reset: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "reset status"),
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

// LoadDeployments returns a tea.Cmd that fetches the deployment history.
func (p DeploymentsPanel) LoadDeployments() tea.Cmd {
	client := p.client
	serverID := p.serverID
	siteID := p.siteID
	return func() tea.Msg {
		deployments, err := client.Deployments.List(context.Background(), serverID, siteID)
		if err != nil {
			return PanelErrMsg{Err: err}
		}
		return DeploymentsLoadedMsg{Deployments: deployments}
	}
}

// TriggerDeploy returns a tea.Cmd that triggers a new deployment.
func (p DeploymentsPanel) TriggerDeploy() tea.Cmd {
	client := p.client
	serverID := p.serverID
	siteID := p.siteID
	return func() tea.Msg {
		err := client.Deployments.Deploy(context.Background(), serverID, siteID)
		if err != nil {
			return PanelErrMsg{Err: err}
		}
		return DeployTriggerMsg{}
	}
}

// LoadOutput returns a tea.Cmd that fetches the output for a deployment.
func (p DeploymentsPanel) LoadOutput(deployID int64) tea.Cmd {
	client := p.client
	serverID := p.serverID
	siteID := p.siteID
	return func() tea.Msg {
		output, err := client.Deployments.GetOutput(context.Background(), serverID, siteID, deployID)
		if err != nil {
			return PanelErrMsg{Err: err}
		}
		return DeployOutputMsg{Output: output}
	}
}

// ResetDeployStatus returns a tea.Cmd that resets the deployment status.
func (p DeploymentsPanel) ResetDeployStatus() tea.Cmd {
	client := p.client
	serverID := p.serverID
	siteID := p.siteID
	return func() tea.Msg {
		err := client.Deployments.ResetStatus(context.Background(), serverID, siteID)
		return DeployResetMsg{Err: err}
	}
}

// DeployTriggerMsg is sent when a deploy has been successfully triggered.
// The app layer handles showing a toast and refreshing the list.
type DeployTriggerMsg struct{}

// PanelErrMsg is sent when a panel API call fails.
// The app layer should catch this and display the error.
type PanelErrMsg struct {
	Err error
}

// Update handles messages for the deployments panel.
func (p DeploymentsPanel) Update(msg tea.Msg) (Panel, tea.Cmd) {
	switch msg := msg.(type) {
	case DeploymentsLoadedMsg:
		p.deployments = msg.Deployments
		p.loading = false
		p.cursor = 0
		return p, nil

	case DeployOutputMsg:
		p.outputView = msg.Output
		p.showOutput = true
		p.outputScroll = 0
		p.loading = false
		return p, nil

	case tea.KeyPressMsg:
		if p.showOutput {
			return p.handleOutputKey(msg)
		}
		return p.handleListKey(msg)
	}

	return p, nil
}

// handleListKey processes key events when viewing the deployment list.
func (p DeploymentsPanel) handleListKey(msg tea.KeyPressMsg) (Panel, tea.Cmd) {
	switch {
	case key.Matches(msg, p.down):
		if len(p.deployments) > 0 {
			p.cursor = min(p.cursor+1, len(p.deployments)-1)
		}
		return p, nil

	case key.Matches(msg, p.up):
		if len(p.deployments) > 0 {
			p.cursor = max(p.cursor-1, 0)
		}
		return p, nil

	case key.Matches(msg, p.home):
		p.cursor = 0
		return p, nil

	case key.Matches(msg, p.end):
		if len(p.deployments) > 0 {
			p.cursor = len(p.deployments) - 1
		}
		return p, nil

	case key.Matches(msg, p.enter):
		if len(p.deployments) > 0 {
			dep := p.deployments[p.cursor]
			p.loading = true
			return p, p.LoadOutput(dep.ID)
		}
		return p, nil

	// 'd' and 'r' are handled by the app layer which shows the confirm dialog.
	// We just return nil here; the app inspects the key before delegating.
	}

	return p, nil
}

// handleOutputKey processes key events when viewing deployment output.
func (p DeploymentsPanel) handleOutputKey(msg tea.KeyPressMsg) (Panel, tea.Cmd) {
	switch {
	case key.Matches(msg, p.back):
		p.showOutput = false
		p.outputView = ""
		p.outputScroll = 0
		return p, nil

	case key.Matches(msg, p.down):
		p.outputScroll++
		return p, nil

	case key.Matches(msg, p.up):
		if p.outputScroll > 0 {
			p.outputScroll--
		}
		return p, nil

	case key.Matches(msg, p.home):
		p.outputScroll = 0
		return p, nil
	}

	return p, nil
}

// ShowingOutput reports whether the panel is currently showing deployment output.
func (p DeploymentsPanel) ShowingOutput() bool {
	return p.showOutput
}

// View renders the deployments panel.
func (p DeploymentsPanel) View(width, height int, focused bool) string {
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

	var title string
	var content string

	if p.showOutput {
		title = lipgloss.NewStyle().
			Bold(true).
			Foreground(titleColor).
			Render(" Deploy Output ")
		content = p.renderOutput(innerWidth, innerHeight)
	} else {
		title = lipgloss.NewStyle().
			Bold(true).
			Foreground(titleColor).
			Render(" Deployments ")
		content = p.renderList(innerWidth, innerHeight)
	}

	return style.
		Width(innerWidth).
		Height(innerHeight).
		Render(title + "\n" + content)
}

// renderList renders the deployment list view.
func (p DeploymentsPanel) renderList(width, height int) string {
	var lines []string

	if p.loading && len(p.deployments) == 0 {
		lines = append(lines, theme.LoadingStyle.Render("Loading deployments..."))
	} else if len(p.deployments) == 0 {
		lines = append(lines, theme.NormalItemStyle.Render("No deployments found"))
	} else {
		// Calculate visible range with scrolling.
		visibleHeight := height - 1 // reserve 1 for title already rendered above
		if visibleHeight < 1 {
			visibleHeight = 1
		}
		startIdx := 0
		if p.cursor >= visibleHeight {
			startIdx = p.cursor - visibleHeight + 1
		}

		for i := startIdx; i < len(p.deployments) && len(lines) < visibleHeight; i++ {
			dep := p.deployments[i]
			line := p.renderDeploymentLine(dep, i, width)
			lines = append(lines, line)
		}
	}

	// Pad to fill the panel height.
	for len(lines) < height {
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}

// renderDeploymentLine renders a single deployment entry.
func (p DeploymentsPanel) renderDeploymentLine(dep forge.Deployment, idx, maxWidth int) string {
	// Status icon.
	icon := statusIcon(dep.Status)

	// Commit message (truncated).
	msg := dep.CommitMessage
	if msg == "" {
		msg = dep.DisplayableType
	}
	if msg == "" {
		msg = "No message"
	}
	// Remove newlines from the commit message.
	msg = strings.ReplaceAll(msg, "\n", " ")

	// Author.
	author := dep.CommitAuthor
	if author == "" {
		author = "-"
	}

	// Time.
	timeStr := relativeTime(dep.EndedAt)
	if timeStr == "" {
		timeStr = relativeTime(dep.StartedAt)
	}
	if timeStr == "" {
		timeStr = "-"
	}

	// Build the line: icon  message  author  time
	// Leave room for: icon(3) + author(~12) + time(~8) + spacing(6) = ~29 chars overhead
	overhead := 29
	msgWidth := maxWidth - overhead
	if msgWidth < 10 {
		msgWidth = 10
	}
	msg = truncatePlain(msg, msgWidth)

	// Format author and time portions.
	authorStr := fmt.Sprintf("%-10s", truncatePlain(author, 10))
	timeStr = fmt.Sprintf("%8s", timeStr)

	if idx == p.cursor {
		line := theme.CursorStyle.Render("> ") +
			icon + " " +
			theme.SelectedItemStyle.Render(msg) +
			"  " + theme.NormalItemStyle.Render(authorStr) +
			"  " + theme.NormalItemStyle.Render(timeStr)
		return theme.Truncate(line, maxWidth)
	}

	line := "  " +
		icon + " " +
		theme.NormalItemStyle.Render(msg) +
		"  " + theme.NormalItemStyle.Render(authorStr) +
		"  " + theme.NormalItemStyle.Render(timeStr)
	return theme.Truncate(line, maxWidth)
}

// renderOutput renders the deployment output with scrolling.
func (p DeploymentsPanel) renderOutput(width, height int) string {
	if p.loading {
		return theme.LoadingStyle.Render("Loading output...")
	}

	if p.outputView == "" {
		return theme.NormalItemStyle.Render("No output available")
	}

	allLines := strings.Split(p.outputView, "\n")

	// Clamp scroll offset.
	maxScroll := len(allLines) - height
	if maxScroll < 0 {
		maxScroll = 0
	}
	scroll := p.outputScroll
	if scroll > maxScroll {
		scroll = maxScroll
	}

	var lines []string
	for i := scroll; i < len(allLines) && len(lines) < height; i++ {
		line := theme.Truncate(allLines[i], width)
		lines = append(lines, theme.NormalItemStyle.Render(line))
	}

	// Pad remaining height.
	for len(lines) < height {
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}

// HelpBindings returns the key hints for the deployments panel.
func (p DeploymentsPanel) HelpBindings() []HelpBinding {
	if p.showOutput {
		return []HelpBinding{
			{Key: "j/k", Desc: "scroll"},
			{Key: "g", Desc: "top"},
			{Key: "esc", Desc: "back"},
			{Key: "tab", Desc: "switch panel"},
			{Key: "q", Desc: "quit"},
		}
	}
	return []HelpBinding{
		{Key: "j/k", Desc: "navigate"},
		{Key: "enter", Desc: "output"},
		{Key: "d", Desc: "deploy"},
		{Key: "S", Desc: "script"},
		{Key: "r", Desc: "reset status"},
		{Key: "g/G", Desc: "top/bottom"},
		{Key: "esc", Desc: "back"},
		{Key: "tab", Desc: "switch panel"},
		{Key: "q", Desc: "quit"},
	}
}

// --- Helpers ---

// statusIcon returns a coloured status indicator for a deployment.
func statusIcon(status string) string {
	switch strings.ToLower(status) {
	case "finished":
		return lipgloss.NewStyle().Foreground(theme.ColorSecondary).Render("✓")
	case "failed":
		return lipgloss.NewStyle().Foreground(theme.ColorError).Render("✗")
	case "deploying":
		return lipgloss.NewStyle().Foreground(theme.ColorHighlight).Render("●")
	default:
		return lipgloss.NewStyle().Foreground(theme.ColorSubtle).Render("?")
	}
}

// relativeTime converts a Forge timestamp string into a human-readable
// relative duration like "2m ago", "1h ago", etc.
func relativeTime(ts string) string {
	if ts == "" {
		return ""
	}

	// Forge timestamps are typically in ISO 8601 / RFC 3339 format.
	layouts := []string{
		time.RFC3339,
		"2006-01-02T15:04:05.000000Z",
		"2006-01-02 15:04:05",
	}

	var t time.Time
	var err error
	for _, layout := range layouts {
		t, err = time.Parse(layout, ts)
		if err == nil {
			break
		}
	}
	if err != nil {
		return ts // fall back to raw string
	}

	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		days := int(d.Hours()) / 24
		return fmt.Sprintf("%dd ago", days)
	}
}

// truncatePlain truncates a plain (no ANSI) string to the given width.
func truncatePlain(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= maxWidth {
		return s
	}
	if maxWidth <= 3 {
		return string(runes[:maxWidth])
	}
	return string(runes[:maxWidth-3]) + "..."
}
