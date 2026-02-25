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

// DaemonsLoadedMsg is sent when the daemon list has been fetched.
type DaemonsLoadedMsg struct {
	Daemons []forge.Daemon
}

// DaemonCreatedMsg is sent when a daemon has been created.
type DaemonCreatedMsg struct {
	Daemon *forge.Daemon
}

// DaemonRestartedMsg is sent when a daemon has been restarted.
type DaemonRestartedMsg struct{}

// DaemonDeletedMsg is sent when a daemon has been deleted.
type DaemonDeletedMsg struct{}

// DaemonsPanel shows the daemons on a server with CRUD actions.
// Daemons are server-level resources (not site-level).
type DaemonsPanel struct {
	client   *forge.Client
	serverID int64

	daemons []forge.Daemon
	cursor  int
	loading bool

	// Keybindings
	up      key.Binding
	down    key.Binding
	create  key.Binding
	restart key.Binding
	del     key.Binding
	home    key.Binding
	end     key.Binding
}

// NewDaemonsPanel creates a new DaemonsPanel.
func NewDaemonsPanel(client *forge.Client, serverID int64) DaemonsPanel {
	return DaemonsPanel{
		client:   client,
		serverID: serverID,
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
			key.WithHelp("c", "create"),
		),
		restart: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "restart"),
		),
		del: key.NewBinding(
			key.WithKeys("x"),
			key.WithHelp("x", "delete"),
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

// LoadDaemons returns a tea.Cmd that fetches the daemon list.
func (p DaemonsPanel) LoadDaemons() tea.Cmd {
	client := p.client
	serverID := p.serverID
	return func() tea.Msg {
		daemons, err := client.Daemons.List(context.Background(), serverID)
		if err != nil {
			return PanelErrMsg{Err: err}
		}
		return DaemonsLoadedMsg{Daemons: daemons}
	}
}

// CreateDaemon returns a tea.Cmd that creates a new daemon with the given command.
func (p DaemonsPanel) CreateDaemon(command string) tea.Cmd {
	client := p.client
	serverID := p.serverID
	return func() tea.Msg {
		opts := forge.DaemonCreateOpts{
			Command:   command,
			User:      "forge",
			Processes: 1,
			StartSecs: 1,
		}
		daemon, err := client.Daemons.Create(context.Background(), serverID, opts)
		if err != nil {
			return PanelErrMsg{Err: err}
		}
		return DaemonCreatedMsg{Daemon: daemon}
	}
}

// RestartDaemon returns a tea.Cmd that restarts the currently selected daemon.
func (p DaemonsPanel) RestartDaemon() tea.Cmd {
	if len(p.daemons) == 0 || p.cursor >= len(p.daemons) {
		return nil
	}
	client := p.client
	serverID := p.serverID
	daemonID := p.daemons[p.cursor].ID
	return func() tea.Msg {
		err := client.Daemons.Restart(context.Background(), serverID, daemonID)
		if err != nil {
			return PanelErrMsg{Err: err}
		}
		return DaemonRestartedMsg{}
	}
}

// DeleteDaemon returns a tea.Cmd that deletes the currently selected daemon.
func (p DaemonsPanel) DeleteDaemon() tea.Cmd {
	if len(p.daemons) == 0 || p.cursor >= len(p.daemons) {
		return nil
	}
	client := p.client
	serverID := p.serverID
	daemonID := p.daemons[p.cursor].ID
	return func() tea.Msg {
		err := client.Daemons.Delete(context.Background(), serverID, daemonID)
		if err != nil {
			return PanelErrMsg{Err: err}
		}
		return DaemonDeletedMsg{}
	}
}

// SelectedDaemon returns the currently selected daemon, or nil.
func (p DaemonsPanel) SelectedDaemon() *forge.Daemon {
	if len(p.daemons) == 0 || p.cursor >= len(p.daemons) {
		return nil
	}
	d := p.daemons[p.cursor]
	return &d
}

// Update handles messages for the daemons panel.
func (p DaemonsPanel) Update(msg tea.Msg) (Panel, tea.Cmd) {
	switch msg := msg.(type) {
	case DaemonsLoadedMsg:
		p.daemons = msg.Daemons
		p.loading = false
		p.cursor = 0
		return p, nil

	case tea.KeyPressMsg:
		return p.handleKey(msg)
	}

	return p, nil
}

func (p DaemonsPanel) handleKey(msg tea.KeyPressMsg) (Panel, tea.Cmd) {
	switch {
	case key.Matches(msg, p.down):
		if len(p.daemons) > 0 {
			p.cursor = min(p.cursor+1, len(p.daemons)-1)
		}
		return p, nil

	case key.Matches(msg, p.up):
		if len(p.daemons) > 0 {
			p.cursor = max(p.cursor-1, 0)
		}
		return p, nil

	case key.Matches(msg, p.home):
		p.cursor = 0
		return p, nil

	case key.Matches(msg, p.end):
		if len(p.daemons) > 0 {
			p.cursor = len(p.daemons) - 1
		}
		return p, nil

	// 'c', 'r', 'x' are handled by the app layer.
	}

	return p, nil
}

// View renders the daemons panel.
func (p DaemonsPanel) View(width, height int, focused bool) string {
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
		Render(" Daemons ")

	content := p.renderList(innerWidth, innerHeight-1)

	return style.
		Width(innerWidth).
		Height(innerHeight).
		Render(title + "\n" + content)
}

func (p DaemonsPanel) renderList(width, height int) string {
	var lines []string

	if p.loading && len(p.daemons) == 0 {
		lines = append(lines, theme.LoadingStyle.Render("Loading daemons..."))
	} else if len(p.daemons) == 0 {
		lines = append(lines, theme.NormalItemStyle.Render("No daemons found"))
	} else {
		visibleHeight := height - 1
		if visibleHeight < 1 {
			visibleHeight = 1
		}
		startIdx := 0
		if p.cursor >= visibleHeight {
			startIdx = p.cursor - visibleHeight + 1
		}

		for i := startIdx; i < len(p.daemons) && len(lines) < visibleHeight; i++ {
			d := p.daemons[i]
			line := p.renderDaemonLine(d, i, width)
			lines = append(lines, line)
		}
	}

	for len(lines) < height {
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}

func (p DaemonsPanel) renderDaemonLine(d forge.Daemon, idx, maxWidth int) string {
	icon := statusIcon(d.Status)

	command := d.Command
	if command == "" {
		command = "-"
	}

	user := d.User
	if user == "" {
		user = "forge"
	}
	procs := fmt.Sprintf("%d procs", d.Processes)
	statusStr := fmt.Sprintf(" [%s]", d.Status)

	// Leave room for: cursor(2) + icon(2) + user(~8) + procs(~8) + status(~14) + spacing(8)
	overhead := 42
	cmdWidth := maxWidth - overhead
	if cmdWidth < 10 {
		cmdWidth = 10
	}
	command = truncatePlain(command, cmdWidth)

	userStr := fmt.Sprintf("%-8s", truncatePlain(user, 8))

	if idx == p.cursor {
		line := theme.CursorStyle.Render("> ") +
			icon + " " +
			theme.SelectedItemStyle.Render(command) +
			"  " + theme.NormalItemStyle.Render(userStr) +
			"  " + theme.NormalItemStyle.Render(procs) +
			"  " + theme.NormalItemStyle.Render(statusStr)
		return theme.Truncate(line, maxWidth)
	}

	line := "  " +
		icon + " " +
		theme.NormalItemStyle.Render(command) +
		"  " + theme.NormalItemStyle.Render(userStr) +
		"  " + theme.NormalItemStyle.Render(procs) +
		"  " + theme.NormalItemStyle.Render(statusStr)
	return theme.Truncate(line, maxWidth)
}

// HelpBindings returns the key hints for the daemons panel.
func (p DaemonsPanel) HelpBindings() []HelpBinding {
	return []HelpBinding{
		{Key: "j/k", Desc: "navigate"},
		{Key: "c", Desc: "create"},
		{Key: "r", Desc: "restart"},
		{Key: "x", Desc: "delete"},
		{Key: "g/G", Desc: "top/bottom"},
		{Key: "esc", Desc: "back"},
		{Key: "tab", Desc: "switch panel"},
		{Key: "q", Desc: "quit"},
	}
}
