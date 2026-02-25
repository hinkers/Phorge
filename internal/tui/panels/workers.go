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

// WorkersLoadedMsg is sent when the worker list has been fetched.
type WorkersLoadedMsg struct {
	Workers []forge.Worker
}

// WorkerCreatedMsg is sent when a worker has been created.
type WorkerCreatedMsg struct {
	Worker *forge.Worker
}

// WorkerRestartedMsg is sent when a worker has been restarted.
type WorkerRestartedMsg struct{}

// WorkerDeletedMsg is sent when a worker has been deleted.
type WorkerDeletedMsg struct{}

// WorkersPanel shows the queue workers for a site with CRUD actions.
type WorkersPanel struct {
	client   *forge.Client
	serverID int64
	siteID   int64

	workers []forge.Worker
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

// NewWorkersPanel creates a new WorkersPanel.
func NewWorkersPanel(client *forge.Client, serverID, siteID int64) WorkersPanel {
	return WorkersPanel{
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

// LoadWorkers returns a tea.Cmd that fetches the worker list.
func (p WorkersPanel) LoadWorkers() tea.Cmd {
	client := p.client
	serverID := p.serverID
	siteID := p.siteID
	return func() tea.Msg {
		workers, err := client.Workers.List(context.Background(), serverID, siteID)
		if err != nil {
			return PanelErrMsg{Err: err}
		}
		return WorkersLoadedMsg{Workers: workers}
	}
}

// CreateWorker returns a tea.Cmd that creates a new worker with default settings.
func (p WorkersPanel) CreateWorker() tea.Cmd {
	client := p.client
	serverID := p.serverID
	siteID := p.siteID
	return func() tea.Msg {
		opts := forge.WorkerCreateOpts{
			Connection: "redis",
			Queue:      "default",
			Timeout:    60,
			Sleep:      3,
			Processes:  1,
			Daemon:     true,
		}
		worker, err := client.Workers.Create(context.Background(), serverID, siteID, opts)
		if err != nil {
			return PanelErrMsg{Err: err}
		}
		return WorkerCreatedMsg{Worker: worker}
	}
}

// RestartWorker returns a tea.Cmd that restarts the currently selected worker.
func (p WorkersPanel) RestartWorker() tea.Cmd {
	if len(p.workers) == 0 || p.cursor >= len(p.workers) {
		return nil
	}
	client := p.client
	serverID := p.serverID
	siteID := p.siteID
	workerID := p.workers[p.cursor].ID
	return func() tea.Msg {
		err := client.Workers.Restart(context.Background(), serverID, siteID, workerID)
		if err != nil {
			return PanelErrMsg{Err: err}
		}
		return WorkerRestartedMsg{}
	}
}

// DeleteWorker returns a tea.Cmd that deletes the currently selected worker.
func (p WorkersPanel) DeleteWorker() tea.Cmd {
	if len(p.workers) == 0 || p.cursor >= len(p.workers) {
		return nil
	}
	client := p.client
	serverID := p.serverID
	siteID := p.siteID
	workerID := p.workers[p.cursor].ID
	return func() tea.Msg {
		err := client.Workers.Delete(context.Background(), serverID, siteID, workerID)
		if err != nil {
			return PanelErrMsg{Err: err}
		}
		return WorkerDeletedMsg{}
	}
}

// SelectedWorker returns the currently selected worker, or nil.
func (p WorkersPanel) SelectedWorker() *forge.Worker {
	if len(p.workers) == 0 || p.cursor >= len(p.workers) {
		return nil
	}
	w := p.workers[p.cursor]
	return &w
}

// Update handles messages for the workers panel.
func (p WorkersPanel) Update(msg tea.Msg) (Panel, tea.Cmd) {
	switch msg := msg.(type) {
	case WorkersLoadedMsg:
		p.workers = msg.Workers
		p.loading = false
		p.cursor = 0
		return p, nil

	case tea.KeyPressMsg:
		return p.handleKey(msg)
	}

	return p, nil
}

func (p WorkersPanel) handleKey(msg tea.KeyPressMsg) (Panel, tea.Cmd) {
	switch {
	case key.Matches(msg, p.down):
		if len(p.workers) > 0 {
			p.cursor = min(p.cursor+1, len(p.workers)-1)
		}
		return p, nil

	case key.Matches(msg, p.up):
		if len(p.workers) > 0 {
			p.cursor = max(p.cursor-1, 0)
		}
		return p, nil

	case key.Matches(msg, p.home):
		p.cursor = 0
		return p, nil

	case key.Matches(msg, p.end):
		if len(p.workers) > 0 {
			p.cursor = len(p.workers) - 1
		}
		return p, nil

	// 'c', 'r', 'x' are handled by the app layer.
	}

	return p, nil
}

// View renders the workers panel.
func (p WorkersPanel) View(width, height int, focused bool) string {
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
		Render(" Workers ")

	content := p.renderList(innerWidth, innerHeight-1)

	return style.
		Width(innerWidth).
		Height(innerHeight).
		Render(title + "\n" + content)
}

func (p WorkersPanel) renderList(width, height int) string {
	var lines []string

	if p.loading && len(p.workers) == 0 {
		lines = append(lines, theme.LoadingStyle.Render("Loading workers..."))
	} else if len(p.workers) == 0 {
		lines = append(lines, theme.NormalItemStyle.Render("No workers found"))
	} else {
		visibleHeight := height - 1
		if visibleHeight < 1 {
			visibleHeight = 1
		}
		startIdx := 0
		if p.cursor >= visibleHeight {
			startIdx = p.cursor - visibleHeight + 1
		}

		for i := startIdx; i < len(p.workers) && len(lines) < visibleHeight; i++ {
			w := p.workers[i]
			line := p.renderWorkerLine(w, i, width)
			lines = append(lines, line)
		}
	}

	for len(lines) < height {
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}

func (p WorkersPanel) renderWorkerLine(w forge.Worker, idx, maxWidth int) string {
	icon := statusIcon(w.Status)

	conn := w.Connection
	if conn == "" {
		conn = "-"
	}
	queue := w.Queue
	if queue == "" {
		queue = "default"
	}
	procs := fmt.Sprintf("%d procs", w.Processes)
	statusStr := fmt.Sprintf(" [%s]", w.Status)

	// Build: connection:queue  procs  status
	connQueue := fmt.Sprintf("%s:%s", conn, queue)

	// Leave room for: cursor(2) + icon(2) + procs(~8) + status(~14) + spacing(6)
	overhead := 32
	connWidth := maxWidth - overhead
	if connWidth < 10 {
		connWidth = 10
	}
	connQueue = truncatePlain(connQueue, connWidth)

	if idx == p.cursor {
		line := theme.CursorStyle.Render("> ") +
			icon + " " +
			theme.SelectedItemStyle.Render(connQueue) +
			"  " + theme.NormalItemStyle.Render(procs) +
			"  " + theme.NormalItemStyle.Render(statusStr)
		return theme.Truncate(line, maxWidth)
	}

	line := "  " +
		icon + " " +
		theme.NormalItemStyle.Render(connQueue) +
		"  " + theme.NormalItemStyle.Render(procs) +
		"  " + theme.NormalItemStyle.Render(statusStr)
	return theme.Truncate(line, maxWidth)
}

// HelpBindings returns the key hints for the workers panel.
func (p WorkersPanel) HelpBindings() []HelpBinding {
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
