package panels

import (
	"context"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/key"
	lipgloss "charm.land/lipgloss/v2"

	"github.com/hinkers/Phorge/internal/forge"
	"github.com/hinkers/Phorge/internal/tui/theme"
)

// --- Messages ---

// CommandsLoadedMsg is sent when the commands list has been fetched.
type CommandsLoadedMsg struct {
	Commands []forge.SiteCommand
}

// CommandCreatedMsg is sent when a command has been executed.
type CommandCreatedMsg struct {
	Command *forge.SiteCommand
}

// CommandDetailMsg is sent when a single command's details have been fetched.
type CommandDetailMsg struct {
	Command *forge.SiteCommand
}

// CommandsPanel shows the list of executed commands on a site.
type CommandsPanel struct {
	client   *forge.Client
	serverID int64
	siteID   int64

	commands []forge.SiteCommand
	cursor   int
	loading  bool

	// Detail sub-view state.
	showDetail    bool
	detailCommand *forge.SiteCommand

	// Keybindings
	up     key.Binding
	down   key.Binding
	create key.Binding
	enter  key.Binding
	home   key.Binding
	end    key.Binding
}

// NewCommandsPanel creates a new CommandsPanel.
func NewCommandsPanel(client *forge.Client, serverID, siteID int64) CommandsPanel {
	return CommandsPanel{
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
		create: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "run command"),
		),
		enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "view details"),
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

// LoadCommands returns a tea.Cmd that fetches the commands list.
func (p CommandsPanel) LoadCommands() tea.Cmd {
	client := p.client
	serverID := p.serverID
	siteID := p.siteID
	return func() tea.Msg {
		cmds, err := client.Commands.List(context.Background(), serverID, siteID)
		if err != nil {
			return PanelErrMsg{Err: err}
		}
		return CommandsLoadedMsg{Commands: cmds}
	}
}

// CreateCommand returns a tea.Cmd that executes a new command on the site.
func (p CommandsPanel) CreateCommand(command string) tea.Cmd {
	client := p.client
	serverID := p.serverID
	siteID := p.siteID
	return func() tea.Msg {
		cmd, err := client.Commands.Create(context.Background(), serverID, siteID, command)
		if err != nil {
			return PanelErrMsg{Err: err}
		}
		return CommandCreatedMsg{Command: cmd}
	}
}

// FetchCommandDetail returns a tea.Cmd that fetches a single command's details.
func (p CommandsPanel) FetchCommandDetail() tea.Cmd {
	if len(p.commands) == 0 || p.cursor >= len(p.commands) {
		return nil
	}
	client := p.client
	serverID := p.serverID
	siteID := p.siteID
	cmdID := p.commands[p.cursor].ID
	return func() tea.Msg {
		cmd, err := client.Commands.Get(context.Background(), serverID, siteID, cmdID)
		if err != nil {
			return PanelErrMsg{Err: err}
		}
		return CommandDetailMsg{Command: cmd}
	}
}

// ShowingDetail reports whether the detail sub-view is active.
func (p CommandsPanel) ShowingDetail() bool {
	return p.showDetail
}

// Update handles messages for the commands panel.
func (p CommandsPanel) Update(msg tea.Msg) (Panel, tea.Cmd) {
	switch msg := msg.(type) {
	case CommandsLoadedMsg:
		p.commands = msg.Commands
		p.loading = false
		p.cursor = 0
		return p, nil

	case CommandDetailMsg:
		p.detailCommand = msg.Command
		p.showDetail = true
		return p, nil

	case tea.KeyPressMsg:
		return p.handleKey(msg)
	}

	return p, nil
}

func (p CommandsPanel) handleKey(msg tea.KeyPressMsg) (Panel, tea.Cmd) {
	// If showing detail, Esc goes back to list.
	if p.showDetail {
		if key.Matches(msg, key.NewBinding(key.WithKeys("esc", "backspace"))) {
			p.showDetail = false
			p.detailCommand = nil
			return p, nil
		}
		// No other keys in detail view.
		return p, nil
	}

	switch {
	case key.Matches(msg, p.down):
		if len(p.commands) > 0 {
			p.cursor = min(p.cursor+1, len(p.commands)-1)
		}
		return p, nil

	case key.Matches(msg, p.up):
		if len(p.commands) > 0 {
			p.cursor = max(p.cursor-1, 0)
		}
		return p, nil

	case key.Matches(msg, p.home):
		p.cursor = 0
		return p, nil

	case key.Matches(msg, p.end):
		if len(p.commands) > 0 {
			p.cursor = len(p.commands) - 1
		}
		return p, nil

	case key.Matches(msg, p.enter):
		return p, p.FetchCommandDetail()

	// 'c' is handled by the app layer.
	}

	return p, nil
}

// View renders the commands panel.
func (p CommandsPanel) View(width, height int, focused bool) string {
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
		Render(" Commands ")

	var content string
	if p.showDetail && p.detailCommand != nil {
		content = p.renderDetail(innerWidth, innerHeight-1)
	} else {
		content = p.renderList(innerWidth, innerHeight-1)
	}

	return style.
		Width(innerWidth).
		Height(innerHeight).
		Render(title + "\n" + content)
}

func (p CommandsPanel) renderDetail(width, height int) string {
	cmd := p.detailCommand
	var lines []string

	lines = append(lines, renderInfoKV("Command", cmd.Command, width))
	lines = append(lines, renderInfoKV("Status", cmd.Status, width))
	lines = append(lines, renderInfoKV("User", cmd.UserName, width))
	lines = append(lines, renderInfoKV("Created", cmd.CreatedAt, width))
	if cmd.Duration != nil {
		lines = append(lines, renderInfoKV("Duration", fmt.Sprintf("%v", cmd.Duration), width))
	}

	lines = append(lines, "")
	lines = append(lines, theme.LabelStyle.Render("Press Esc to go back"))

	for len(lines) < height {
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}

// Column widths for commands table.
const (
	cmdColUserWidth = 10
	cmdColDateWidth = 12
)

const cmdTableOverhead = 2 + colStatusWidth + 2 + 2 + cmdColUserWidth + 2 + cmdColDateWidth + 4

func cmdFlexWidth(maxWidth int) int {
	w := maxWidth - cmdTableOverhead
	if w < 10 {
		w = 10
	}
	return w
}

func (p CommandsPanel) renderList(width, height int) string {
	var lines []string

	if p.loading && len(p.commands) == 0 {
		lines = append(lines, theme.LoadingStyle.Render("Loading commands..."))
	} else if len(p.commands) == 0 {
		lines = append(lines, theme.NormalItemStyle.Render("No commands found"))
	} else {
		lines = append(lines, p.renderCommandHeader(width))

		visibleHeight := height - 2
		if visibleHeight < 1 {
			visibleHeight = 1
		}
		startIdx := 0
		if p.cursor >= visibleHeight {
			startIdx = p.cursor - visibleHeight + 1
		}

		for i := startIdx; i < len(p.commands) && len(lines)-1 < visibleHeight; i++ {
			cmd := p.commands[i]
			line := p.renderCommandLine(cmd, i, width)
			lines = append(lines, line)
		}
	}

	for len(lines) < height {
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}

func (p CommandsPanel) renderCommandHeader(maxWidth int) string {
	flexW := cmdFlexWidth(maxWidth)
	line := fmt.Sprintf("  %-*s  %-*s  %-*s  %-*s",
		colStatusWidth, "STATUS",
		flexW, "COMMAND",
		cmdColUserWidth, "USER",
		cmdColDateWidth, "DATE",
	)
	return theme.Truncate(headerStyle.Render(line), maxWidth)
}

func (p CommandsPanel) renderCommandLine(cmd forge.SiteCommand, idx, maxWidth int) string {
	icon := statusIcon(cmd.Status)
	statusText := cmd.Status
	if statusText == "" {
		statusText = "unknown"
	}

	command := cmd.Command
	if command == "" {
		command = "-"
	}

	user := cmd.UserName
	if user == "" {
		user = "-"
	}

	date := cmd.CreatedAt
	if date == "" {
		date = "-"
	}
	// Trim to date portion if it contains a time.
	if len(date) > cmdColDateWidth {
		date = date[:cmdColDateWidth]
	}

	flexW := cmdFlexWidth(maxWidth)
	command = truncatePlain(command, flexW)

	statusPad := colStatusWidth - 2
	statusStr := icon + " " + fmt.Sprintf("%-*s", statusPad, truncatePlain(statusText, statusPad))
	userStr := fmt.Sprintf("%-*s", cmdColUserWidth, truncatePlain(user, cmdColUserWidth))
	dateStr := fmt.Sprintf("%-*s", cmdColDateWidth, truncatePlain(date, cmdColDateWidth))

	if idx == p.cursor {
		line := theme.CursorStyle.Render("> ") +
			statusStr +
			"  " + theme.SelectedItemStyle.Render(fmt.Sprintf("%-*s", flexW, command)) +
			"  " + theme.NormalItemStyle.Render(userStr) +
			"  " + theme.NormalItemStyle.Render(dateStr)
		return theme.Truncate(line, maxWidth)
	}

	line := "  " +
		statusStr +
		"  " + theme.NormalItemStyle.Render(fmt.Sprintf("%-*s", flexW, command)) +
		"  " + theme.NormalItemStyle.Render(userStr) +
		"  " + theme.NormalItemStyle.Render(dateStr)
	return theme.Truncate(line, maxWidth)
}

// HelpBindings returns the key hints for the commands panel.
func (p CommandsPanel) HelpBindings() []HelpBinding {
	if p.showDetail {
		return []HelpBinding{
			{Key: "esc", Desc: "back to list"},
			{Key: "tab", Desc: "switch panel"},
			{Key: "q", Desc: "quit"},
		}
	}
	return []HelpBinding{
		{Key: "j/k", Desc: "navigate"},
		{Key: "enter", Desc: "view details"},
		{Key: "c", Desc: "run command"},
		{Key: "g/G", Desc: "top/bottom"},
		{Key: "esc", Desc: "back"},
		{Key: "tab", Desc: "switch panel"},
		{Key: "q", Desc: "quit"},
	}
}
