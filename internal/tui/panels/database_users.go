package panels

import (
	"context"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/key"
	lipgloss "charm.land/lipgloss/v2"

	"github.com/hinke/phorge/internal/forge"
	"github.com/hinke/phorge/internal/tui/theme"
)

// --- Messages ---

// DBUsersLoadedMsg is sent when the database user list has been fetched.
type DBUsersLoadedMsg struct {
	Users []forge.DatabaseUser
}

// DBUserCreatedMsg is sent when a database user has been created.
type DBUserCreatedMsg struct {
	User *forge.DatabaseUser
}

// DBUserDeletedMsg is sent when a database user has been deleted.
type DBUserDeletedMsg struct{}

// DBUsersPanel shows the list of database users on a server.
type DBUsersPanel struct {
	client   *forge.Client
	serverID int64

	users   []forge.DatabaseUser
	cursor  int
	loading bool

	// Keybindings
	up     key.Binding
	down   key.Binding
	create key.Binding
	del    key.Binding
	back   key.Binding
	home   key.Binding
	end    key.Binding
}

// NewDBUsersPanel creates a new DBUsersPanel.
func NewDBUsersPanel(client *forge.Client, serverID int64) DBUsersPanel {
	return DBUsersPanel{
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

// LoadUsers returns a tea.Cmd that fetches the database user list.
func (p DBUsersPanel) LoadUsers() tea.Cmd {
	client := p.client
	serverID := p.serverID
	return func() tea.Msg {
		users, err := client.Databases.ListUsers(context.Background(), serverID)
		if err != nil {
			return PanelErrMsg{Err: err}
		}
		return DBUsersLoadedMsg{Users: users}
	}
}

// CreateUser returns a tea.Cmd that creates a new database user.
// For simplicity, password is auto-generated and databases is empty initially.
func (p DBUsersPanel) CreateUser(name, password string) tea.Cmd {
	client := p.client
	serverID := p.serverID
	return func() tea.Msg {
		user, err := client.Databases.CreateUser(context.Background(), serverID, name, password, nil)
		if err != nil {
			return PanelErrMsg{Err: err}
		}
		return DBUserCreatedMsg{User: user}
	}
}

// DeleteUser returns a tea.Cmd that deletes the currently selected database user.
func (p DBUsersPanel) DeleteUser() tea.Cmd {
	if len(p.users) == 0 || p.cursor >= len(p.users) {
		return nil
	}
	client := p.client
	serverID := p.serverID
	userID := p.users[p.cursor].ID
	return func() tea.Msg {
		err := client.Databases.DeleteUser(context.Background(), serverID, userID)
		if err != nil {
			return PanelErrMsg{Err: err}
		}
		return DBUserDeletedMsg{}
	}
}

// SelectedUser returns the currently selected database user, or nil.
func (p DBUsersPanel) SelectedUser() *forge.DatabaseUser {
	if len(p.users) == 0 || p.cursor >= len(p.users) {
		return nil
	}
	u := p.users[p.cursor]
	return &u
}

// Update handles messages for the database users panel.
func (p DBUsersPanel) Update(msg tea.Msg) (Panel, tea.Cmd) {
	switch msg := msg.(type) {
	case DBUsersLoadedMsg:
		p.users = msg.Users
		p.loading = false
		p.cursor = 0
		return p, nil

	case tea.KeyPressMsg:
		return p.handleKey(msg)
	}

	return p, nil
}

func (p DBUsersPanel) handleKey(msg tea.KeyPressMsg) (Panel, tea.Cmd) {
	switch {
	case key.Matches(msg, p.down):
		if len(p.users) > 0 {
			p.cursor = min(p.cursor+1, len(p.users)-1)
		}
		return p, nil

	case key.Matches(msg, p.up):
		if len(p.users) > 0 {
			p.cursor = max(p.cursor-1, 0)
		}
		return p, nil

	case key.Matches(msg, p.home):
		p.cursor = 0
		return p, nil

	case key.Matches(msg, p.end):
		if len(p.users) > 0 {
			p.cursor = len(p.users) - 1
		}
		return p, nil

	// 'c', 'x' are handled by the app layer.
	}

	return p, nil
}

// View renders the database users panel.
func (p DBUsersPanel) View(width, height int, focused bool) string {
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
		Render(" Database Users ")

	content := p.renderList(innerWidth, innerHeight)

	return style.
		Width(innerWidth).
		Height(innerHeight).
		Render(title + "\n" + content)
}

func (p DBUsersPanel) renderList(width, height int) string {
	var lines []string

	if p.loading && len(p.users) == 0 {
		lines = append(lines, theme.LoadingStyle.Render("Loading database users..."))
	} else if len(p.users) == 0 {
		lines = append(lines, theme.NormalItemStyle.Render("No database users found"))
	} else {
		visibleHeight := height - 1
		if visibleHeight < 1 {
			visibleHeight = 1
		}
		startIdx := 0
		if p.cursor >= visibleHeight {
			startIdx = p.cursor - visibleHeight + 1
		}

		for i := startIdx; i < len(p.users) && len(lines) < visibleHeight; i++ {
			user := p.users[i]
			line := p.renderUserLine(user, i, width)
			lines = append(lines, line)
		}
	}

	for len(lines) < height {
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}

func (p DBUsersPanel) renderUserLine(user forge.DatabaseUser, idx, maxWidth int) string {
	icon := statusIcon(user.Status)

	name := user.Name
	dbCount := fmt.Sprintf(" %d dbs", len(user.Databases))
	statusStr := fmt.Sprintf(" [%s]", user.Status)

	// Leave room for: cursor(2) + icon(2) + dbCount(~8) + status(~14) + spacing(6)
	overhead := 32
	nameWidth := maxWidth - overhead
	if nameWidth < 10 {
		nameWidth = 10
	}
	name = truncatePlain(name, nameWidth)

	if idx == p.cursor {
		line := theme.CursorStyle.Render("> ") +
			icon + " " +
			theme.SelectedItemStyle.Render(name) +
			"  " + theme.NormalItemStyle.Render(dbCount) +
			"  " + theme.NormalItemStyle.Render(statusStr)
		return theme.Truncate(line, maxWidth)
	}

	line := "  " +
		icon + " " +
		theme.NormalItemStyle.Render(name) +
		"  " + theme.NormalItemStyle.Render(dbCount) +
		"  " + theme.NormalItemStyle.Render(statusStr)
	return theme.Truncate(line, maxWidth)
}

// HelpBindings returns the key hints for the database users panel.
func (p DBUsersPanel) HelpBindings() []HelpBinding {
	return []HelpBinding{
		{Key: "j/k", Desc: "navigate"},
		{Key: "c", Desc: "create"},
		{Key: "x", Desc: "delete"},
		{Key: "g/G", Desc: "top/bottom"},
		{Key: "esc", Desc: "back to databases"},
		{Key: "tab", Desc: "switch panel"},
		{Key: "q", Desc: "quit"},
	}
}
