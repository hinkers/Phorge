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

// JobsLoadedMsg is sent when the scheduled job list has been fetched.
type JobsLoadedMsg struct {
	Jobs []forge.ScheduledJob
}

// JobsPanel shows the scheduled jobs on a server (read-only list).
// Jobs are server-level resources.
type JobsPanel struct {
	client   *forge.Client
	serverID int64

	jobs    []forge.ScheduledJob
	cursor  int
	loading bool

	// Keybindings
	up   key.Binding
	down key.Binding
	home key.Binding
	end  key.Binding
}

// NewJobsPanel creates a new JobsPanel.
func NewJobsPanel(client *forge.Client, serverID int64) JobsPanel {
	return JobsPanel{
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

// LoadJobs returns a tea.Cmd that fetches the scheduled job list.
func (p JobsPanel) LoadJobs() tea.Cmd {
	client := p.client
	serverID := p.serverID
	return func() tea.Msg {
		jobs, err := client.Jobs.List(context.Background(), serverID)
		if err != nil {
			return PanelErrMsg{Err: err}
		}
		return JobsLoadedMsg{Jobs: jobs}
	}
}

// Update handles messages for the jobs panel.
func (p JobsPanel) Update(msg tea.Msg) (Panel, tea.Cmd) {
	switch msg := msg.(type) {
	case JobsLoadedMsg:
		p.jobs = msg.Jobs
		p.loading = false
		p.cursor = 0
		return p, nil

	case tea.KeyPressMsg:
		return p.handleKey(msg)
	}

	return p, nil
}

func (p JobsPanel) handleKey(msg tea.KeyPressMsg) (Panel, tea.Cmd) {
	switch {
	case key.Matches(msg, p.down):
		if len(p.jobs) > 0 {
			p.cursor = min(p.cursor+1, len(p.jobs)-1)
		}
		return p, nil

	case key.Matches(msg, p.up):
		if len(p.jobs) > 0 {
			p.cursor = max(p.cursor-1, 0)
		}
		return p, nil

	case key.Matches(msg, p.home):
		p.cursor = 0
		return p, nil

	case key.Matches(msg, p.end):
		if len(p.jobs) > 0 {
			p.cursor = len(p.jobs) - 1
		}
		return p, nil
	}

	return p, nil
}

// View renders the jobs panel.
func (p JobsPanel) View(width, height int, focused bool) string {
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
		Render(" Scheduled Jobs ")

	content := p.renderList(innerWidth, innerHeight-1)

	return style.
		Width(innerWidth).
		Height(innerHeight).
		Render(title + "\n" + content)
}

// Column widths for jobs table.
const (
	jobColSchedWidth = 14
	jobColUserWidth  = 8
)

const jobTableOverhead = 2 + colStatusWidth + 2 + 2 + jobColSchedWidth + 2 + jobColUserWidth + 4

func jobCmdWidth(maxWidth int) int {
	w := maxWidth - jobTableOverhead
	if w < 10 {
		w = 10
	}
	return w
}

func (p JobsPanel) renderList(width, height int) string {
	var lines []string

	if p.loading && len(p.jobs) == 0 {
		lines = append(lines, theme.LoadingStyle.Render("Loading scheduled jobs..."))
	} else if len(p.jobs) == 0 {
		lines = append(lines, theme.NormalItemStyle.Render("No scheduled jobs found"))
	} else {
		lines = append(lines, p.renderJobHeader(width))

		visibleHeight := height - 2
		if visibleHeight < 1 {
			visibleHeight = 1
		}
		startIdx := 0
		if p.cursor >= visibleHeight {
			startIdx = p.cursor - visibleHeight + 1
		}

		for i := startIdx; i < len(p.jobs) && len(lines)-1 < visibleHeight; i++ {
			job := p.jobs[i]
			line := p.renderJobLine(job, i, width)
			lines = append(lines, line)
		}
	}

	for len(lines) < height {
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}

func (p JobsPanel) renderJobHeader(maxWidth int) string {
	cmdWidth := jobCmdWidth(maxWidth)
	line := fmt.Sprintf("  %-*s  %-*s  %-*s  %-*s",
		colStatusWidth, "STATUS",
		cmdWidth, "COMMAND",
		jobColSchedWidth, "SCHEDULE",
		jobColUserWidth, "USER",
	)
	return theme.Truncate(headerStyle.Render(line), maxWidth)
}

func (p JobsPanel) renderJobLine(job forge.ScheduledJob, idx, maxWidth int) string {
	icon := statusIcon(job.Status)
	statusText := job.Status
	if statusText == "" {
		statusText = "unknown"
	}

	command := job.Command
	if command == "" {
		command = "-"
	}

	freq := job.Cron
	if freq == "" {
		freq = job.Frequency
	}
	if freq == "" {
		freq = "-"
	}

	user := job.User
	if user == "" {
		user = "forge"
	}

	cmdWidth := jobCmdWidth(maxWidth)
	command = truncatePlain(command, cmdWidth)

	statusPad := colStatusWidth - 2
	statusStr := icon + " " + fmt.Sprintf("%-*s", statusPad, truncatePlain(statusText, statusPad))
	freqStr := fmt.Sprintf("%-*s", jobColSchedWidth, truncatePlain(freq, jobColSchedWidth))
	userStr := fmt.Sprintf("%-*s", jobColUserWidth, truncatePlain(user, jobColUserWidth))

	if idx == p.cursor {
		line := theme.CursorStyle.Render("> ") +
			statusStr +
			"  " + theme.SelectedItemStyle.Render(fmt.Sprintf("%-*s", cmdWidth, command)) +
			"  " + theme.NormalItemStyle.Render(freqStr) +
			"  " + theme.NormalItemStyle.Render(userStr)
		return theme.Truncate(line, maxWidth)
	}

	line := "  " +
		statusStr +
		"  " + theme.NormalItemStyle.Render(fmt.Sprintf("%-*s", cmdWidth, command)) +
		"  " + theme.NormalItemStyle.Render(freqStr) +
		"  " + theme.NormalItemStyle.Render(userStr)
	return theme.Truncate(line, maxWidth)
}

// HelpBindings returns the key hints for the jobs panel.
func (p JobsPanel) HelpBindings() []HelpBinding {
	return []HelpBinding{
		{Key: "j/k", Desc: "navigate"},
		{Key: "g/G", Desc: "top/bottom"},
		{Key: "esc", Desc: "back"},
		{Key: "tab", Desc: "switch panel"},
		{Key: "q", Desc: "quit"},
	}
}
