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

// DatabasesLoadedMsg is sent when the database list has been fetched.
type DatabasesLoadedMsg struct {
	Databases []forge.Database
}

// DatabaseCreatedMsg is sent when a database has been created.
type DatabaseCreatedMsg struct {
	Database *forge.Database
}

// DatabaseDeletedMsg is sent when a database has been deleted.
type DatabaseDeletedMsg struct{}

// DatabasesPanel shows the list of databases on a server with CRUD actions.
type DatabasesPanel struct {
	client   *forge.Client
	serverID int64

	databases []forge.Database
	cursor    int
	loading   bool

	// Keybindings
	up     key.Binding
	down   key.Binding
	create key.Binding
	del    key.Binding
	users  key.Binding
	home   key.Binding
	end    key.Binding
}

// NewDatabasesPanel creates a new DatabasesPanel.
func NewDatabasesPanel(client *forge.Client, serverID int64) DatabasesPanel {
	return DatabasesPanel{
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
		del: key.NewBinding(
			key.WithKeys("x"),
			key.WithHelp("x", "delete"),
		),
		users: key.NewBinding(
			key.WithKeys("u"),
			key.WithHelp("u", "users"),
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

// LoadDatabases returns a tea.Cmd that fetches the database list.
func (p DatabasesPanel) LoadDatabases() tea.Cmd {
	client := p.client
	serverID := p.serverID
	return func() tea.Msg {
		databases, err := client.Databases.List(context.Background(), serverID)
		if err != nil {
			return PanelErrMsg{Err: err}
		}
		return DatabasesLoadedMsg{Databases: databases}
	}
}

// CreateDatabase returns a tea.Cmd that creates a new database.
func (p DatabasesPanel) CreateDatabase(name string) tea.Cmd {
	client := p.client
	serverID := p.serverID
	return func() tea.Msg {
		db, err := client.Databases.Create(context.Background(), serverID, name, nil, nil)
		if err != nil {
			return PanelErrMsg{Err: err}
		}
		return DatabaseCreatedMsg{Database: db}
	}
}

// DeleteDatabase returns a tea.Cmd that deletes the currently selected database.
func (p DatabasesPanel) DeleteDatabase() tea.Cmd {
	if len(p.databases) == 0 || p.cursor >= len(p.databases) {
		return nil
	}
	client := p.client
	serverID := p.serverID
	dbID := p.databases[p.cursor].ID
	return func() tea.Msg {
		err := client.Databases.Delete(context.Background(), serverID, dbID)
		if err != nil {
			return PanelErrMsg{Err: err}
		}
		return DatabaseDeletedMsg{}
	}
}

// SelectedDatabase returns the currently selected database, or nil.
func (p DatabasesPanel) SelectedDatabase() *forge.Database {
	if len(p.databases) == 0 || p.cursor >= len(p.databases) {
		return nil
	}
	db := p.databases[p.cursor]
	return &db
}

// Update handles messages for the databases panel.
func (p DatabasesPanel) Update(msg tea.Msg) (Panel, tea.Cmd) {
	switch msg := msg.(type) {
	case DatabasesLoadedMsg:
		p.databases = msg.Databases
		p.loading = false
		p.cursor = 0
		return p, nil

	case tea.KeyPressMsg:
		return p.handleKey(msg)
	}

	return p, nil
}

func (p DatabasesPanel) handleKey(msg tea.KeyPressMsg) (Panel, tea.Cmd) {
	switch {
	case key.Matches(msg, p.down):
		if len(p.databases) > 0 {
			p.cursor = min(p.cursor+1, len(p.databases)-1)
		}
		return p, nil

	case key.Matches(msg, p.up):
		if len(p.databases) > 0 {
			p.cursor = max(p.cursor-1, 0)
		}
		return p, nil

	case key.Matches(msg, p.home):
		p.cursor = 0
		return p, nil

	case key.Matches(msg, p.end):
		if len(p.databases) > 0 {
			p.cursor = len(p.databases) - 1
		}
		return p, nil

	// 'c', 'x', 'u' are handled by the app layer.
	}

	return p, nil
}

// View renders the databases panel.
func (p DatabasesPanel) View(width, height int, focused bool) string {
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
		Render(" Databases ")

	content := p.renderList(innerWidth, innerHeight-1)

	return style.
		Width(innerWidth).
		Height(innerHeight).
		Render(title + "\n" + content)
}

func (p DatabasesPanel) renderList(width, height int) string {
	var lines []string

	if p.loading && len(p.databases) == 0 {
		lines = append(lines, theme.LoadingStyle.Render("Loading databases..."))
	} else if len(p.databases) == 0 {
		lines = append(lines, theme.NormalItemStyle.Render("No databases found"))
	} else {
		visibleHeight := height - 1
		if visibleHeight < 1 {
			visibleHeight = 1
		}
		startIdx := 0
		if p.cursor >= visibleHeight {
			startIdx = p.cursor - visibleHeight + 1
		}

		for i := startIdx; i < len(p.databases) && len(lines) < visibleHeight; i++ {
			db := p.databases[i]
			line := p.renderDatabaseLine(db, i, width)
			lines = append(lines, line)
		}
	}

	for len(lines) < height {
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}

func (p DatabasesPanel) renderDatabaseLine(db forge.Database, idx, maxWidth int) string {
	icon := statusIcon(db.Status)

	name := db.Name
	statusStr := fmt.Sprintf(" [%s]", db.Status)

	// Leave room for: cursor(2) + icon(2) + status(~14) + spacing(4)
	overhead := 22
	nameWidth := maxWidth - overhead
	if nameWidth < 10 {
		nameWidth = 10
	}
	name = truncatePlain(name, nameWidth)

	if idx == p.cursor {
		line := theme.CursorStyle.Render("> ") +
			icon + " " +
			theme.SelectedItemStyle.Render(name) +
			"  " + theme.NormalItemStyle.Render(statusStr)
		return theme.Truncate(line, maxWidth)
	}

	line := "  " +
		icon + " " +
		theme.NormalItemStyle.Render(name) +
		"  " + theme.NormalItemStyle.Render(statusStr)
	return theme.Truncate(line, maxWidth)
}

// HelpBindings returns the key hints for the databases panel.
func (p DatabasesPanel) HelpBindings() []HelpBinding {
	return []HelpBinding{
		{Key: "j/k", Desc: "navigate"},
		{Key: "c", Desc: "create"},
		{Key: "x", Desc: "delete"},
		{Key: "u", Desc: "users"},
		{Key: "g/G", Desc: "top/bottom"},
		{Key: "esc", Desc: "back"},
		{Key: "tab", Desc: "switch panel"},
		{Key: "q", Desc: "quit"},
	}
}
