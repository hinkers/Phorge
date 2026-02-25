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

// SSHKeysLoadedMsg is sent when the SSH key list has been fetched.
type SSHKeysLoadedMsg struct {
	Keys []forge.SSHKey
}

// SSHKeyCreatedMsg is sent when an SSH key has been created.
type SSHKeyCreatedMsg struct {
	Key *forge.SSHKey
}

// SSHKeyDeletedMsg is sent when an SSH key has been deleted.
type SSHKeyDeletedMsg struct{}

// SSHKeysPanel shows the list of SSH keys on a server with CRUD actions.
type SSHKeysPanel struct {
	client   *forge.Client
	serverID int64

	keys    []forge.SSHKey
	cursor  int
	loading bool

	// Keybindings
	up     key.Binding
	down   key.Binding
	create key.Binding
	del    key.Binding
	home   key.Binding
	end    key.Binding
}

// NewSSHKeysPanel creates a new SSHKeysPanel.
func NewSSHKeysPanel(client *forge.Client, serverID int64) SSHKeysPanel {
	return SSHKeysPanel{
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

// LoadKeys returns a tea.Cmd that fetches the SSH key list.
func (p SSHKeysPanel) LoadKeys() tea.Cmd {
	client := p.client
	serverID := p.serverID
	return func() tea.Msg {
		keys, err := client.SSHKeys.List(context.Background(), serverID)
		if err != nil {
			return PanelErrMsg{Err: err}
		}
		return SSHKeysLoadedMsg{Keys: keys}
	}
}

// CreateKey returns a tea.Cmd that creates a new SSH key.
func (p SSHKeysPanel) CreateKey(name, keyContent, username string) tea.Cmd {
	client := p.client
	serverID := p.serverID
	return func() tea.Msg {
		k, err := client.SSHKeys.Create(context.Background(), serverID, name, keyContent, username)
		if err != nil {
			return PanelErrMsg{Err: err}
		}
		return SSHKeyCreatedMsg{Key: k}
	}
}

// DeleteKey returns a tea.Cmd that deletes the currently selected SSH key.
func (p SSHKeysPanel) DeleteKey() tea.Cmd {
	if len(p.keys) == 0 || p.cursor >= len(p.keys) {
		return nil
	}
	client := p.client
	serverID := p.serverID
	keyID := p.keys[p.cursor].ID
	return func() tea.Msg {
		err := client.SSHKeys.Delete(context.Background(), serverID, keyID)
		if err != nil {
			return PanelErrMsg{Err: err}
		}
		return SSHKeyDeletedMsg{}
	}
}

// SelectedKey returns the currently selected SSH key, or nil.
func (p SSHKeysPanel) SelectedKey() *forge.SSHKey {
	if len(p.keys) == 0 || p.cursor >= len(p.keys) {
		return nil
	}
	k := p.keys[p.cursor]
	return &k
}

// Update handles messages for the SSH keys panel.
func (p SSHKeysPanel) Update(msg tea.Msg) (Panel, tea.Cmd) {
	switch msg := msg.(type) {
	case SSHKeysLoadedMsg:
		p.keys = msg.Keys
		p.loading = false
		p.cursor = 0
		return p, nil

	case tea.KeyPressMsg:
		return p.handleKey(msg)
	}

	return p, nil
}

func (p SSHKeysPanel) handleKey(msg tea.KeyPressMsg) (Panel, tea.Cmd) {
	switch {
	case key.Matches(msg, p.down):
		if len(p.keys) > 0 {
			p.cursor = min(p.cursor+1, len(p.keys)-1)
		}
		return p, nil

	case key.Matches(msg, p.up):
		if len(p.keys) > 0 {
			p.cursor = max(p.cursor-1, 0)
		}
		return p, nil

	case key.Matches(msg, p.home):
		p.cursor = 0
		return p, nil

	case key.Matches(msg, p.end):
		if len(p.keys) > 0 {
			p.cursor = len(p.keys) - 1
		}
		return p, nil

	// 'c', 'x' are handled by the app layer.
	}

	return p, nil
}

// View renders the SSH keys panel.
func (p SSHKeysPanel) View(width, height int, focused bool) string {
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
		Render(" SSH Keys ")

	content := p.renderList(innerWidth, innerHeight-1)

	return style.
		Width(innerWidth).
		Height(innerHeight).
		Render(title + "\n" + content)
}

func (p SSHKeysPanel) renderList(width, height int) string {
	var lines []string

	if p.loading && len(p.keys) == 0 {
		lines = append(lines, theme.LoadingStyle.Render("Loading SSH keys..."))
	} else if len(p.keys) == 0 {
		lines = append(lines, theme.NormalItemStyle.Render("No SSH keys found"))
	} else {
		visibleHeight := height - 1
		if visibleHeight < 1 {
			visibleHeight = 1
		}
		startIdx := 0
		if p.cursor >= visibleHeight {
			startIdx = p.cursor - visibleHeight + 1
		}

		for i := startIdx; i < len(p.keys) && len(lines) < visibleHeight; i++ {
			k := p.keys[i]
			line := p.renderKeyLine(k, i, width)
			lines = append(lines, line)
		}
	}

	for len(lines) < height {
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}

func (p SSHKeysPanel) renderKeyLine(k forge.SSHKey, idx, maxWidth int) string {
	icon := statusIcon(k.Status)

	name := k.Name
	statusStr := fmt.Sprintf(" [%s]", k.Status)

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

// HelpBindings returns the key hints for the SSH keys panel.
func (p SSHKeysPanel) HelpBindings() []HelpBinding {
	return []HelpBinding{
		{Key: "j/k", Desc: "navigate"},
		{Key: "c", Desc: "create"},
		{Key: "x", Desc: "delete"},
		{Key: "g/G", Desc: "top/bottom"},
		{Key: "esc", Desc: "back"},
		{Key: "tab", Desc: "switch panel"},
		{Key: "q", Desc: "quit"},
	}
}
