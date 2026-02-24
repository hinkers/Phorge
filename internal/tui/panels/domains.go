package panels

import (
	"context"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/key"
	lipgloss "charm.land/lipgloss/v2"

	"github.com/hinke/phorge/internal/forge"
	"github.com/hinke/phorge/internal/tui/theme"
)

// --- Messages ---

// DomainsLoadedMsg is sent when the domain aliases have been loaded from the site.
type DomainsLoadedMsg struct {
	Aliases []string
}

// DomainsSavedMsg is sent after the domain aliases have been updated.
type DomainsSavedMsg struct {
	Err error
}

// DomainsPanel shows the domain aliases for a site with add/remove actions.
type DomainsPanel struct {
	client   *forge.Client
	serverID int64
	siteID   int64

	aliases []string
	cursor  int
	loading bool

	// Keybindings
	up     key.Binding
	down   key.Binding
	add    key.Binding
	remove key.Binding
	home   key.Binding
	end    key.Binding
}

// NewDomainsPanel creates a new DomainsPanel.
func NewDomainsPanel(client *forge.Client, serverID, siteID int64, aliases []string) DomainsPanel {
	return DomainsPanel{
		client:   client,
		serverID: serverID,
		siteID:   siteID,
		aliases:  aliases,
		loading:  false,
		up: key.NewBinding(
			key.WithKeys("k", "up"),
			key.WithHelp("k/up", "up"),
		),
		down: key.NewBinding(
			key.WithKeys("j", "down"),
			key.WithHelp("j/down", "down"),
		),
		add: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "add alias"),
		),
		remove: key.NewBinding(
			key.WithKeys("x"),
			key.WithHelp("x", "remove"),
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

// AddAlias adds a new alias and saves the full list via the API.
func (p DomainsPanel) AddAlias(alias string) tea.Cmd {
	newAliases := make([]string, len(p.aliases))
	copy(newAliases, p.aliases)
	newAliases = append(newAliases, alias)

	client := p.client
	serverID := p.serverID
	siteID := p.siteID
	return func() tea.Msg {
		_, err := client.Sites.UpdateAliases(context.Background(), serverID, siteID, newAliases)
		if err != nil {
			return PanelErrMsg{Err: err}
		}
		return DomainsSavedMsg{Err: nil}
	}
}

// RemoveAlias removes the currently selected alias and saves the full list via the API.
func (p DomainsPanel) RemoveAlias() tea.Cmd {
	if len(p.aliases) == 0 || p.cursor >= len(p.aliases) {
		return nil
	}

	newAliases := make([]string, 0, len(p.aliases)-1)
	for i, a := range p.aliases {
		if i != p.cursor {
			newAliases = append(newAliases, a)
		}
	}

	client := p.client
	serverID := p.serverID
	siteID := p.siteID
	return func() tea.Msg {
		_, err := client.Sites.UpdateAliases(context.Background(), serverID, siteID, newAliases)
		if err != nil {
			return PanelErrMsg{Err: err}
		}
		return DomainsSavedMsg{Err: nil}
	}
}

// SelectedAlias returns the currently selected alias, or empty string.
func (p DomainsPanel) SelectedAlias() string {
	if len(p.aliases) == 0 || p.cursor >= len(p.aliases) {
		return ""
	}
	return p.aliases[p.cursor]
}

// RefreshAliases returns a tea.Cmd that fetches the latest site data to update aliases.
func (p DomainsPanel) RefreshAliases() tea.Cmd {
	client := p.client
	serverID := p.serverID
	siteID := p.siteID
	return func() tea.Msg {
		site, err := client.Sites.Get(context.Background(), serverID, siteID)
		if err != nil {
			return PanelErrMsg{Err: err}
		}
		return DomainsLoadedMsg{Aliases: site.Aliases}
	}
}

// Update handles messages for the domains panel.
func (p DomainsPanel) Update(msg tea.Msg) (Panel, tea.Cmd) {
	switch msg := msg.(type) {
	case DomainsLoadedMsg:
		p.aliases = msg.Aliases
		p.loading = false
		if p.cursor >= len(p.aliases) {
			p.cursor = max(len(p.aliases)-1, 0)
		}
		return p, nil

	case tea.KeyPressMsg:
		return p.handleKey(msg)
	}

	return p, nil
}

func (p DomainsPanel) handleKey(msg tea.KeyPressMsg) (Panel, tea.Cmd) {
	switch {
	case key.Matches(msg, p.down):
		if len(p.aliases) > 0 {
			p.cursor = min(p.cursor+1, len(p.aliases)-1)
		}
		return p, nil

	case key.Matches(msg, p.up):
		if len(p.aliases) > 0 {
			p.cursor = max(p.cursor-1, 0)
		}
		return p, nil

	case key.Matches(msg, p.home):
		p.cursor = 0
		return p, nil

	case key.Matches(msg, p.end):
		if len(p.aliases) > 0 {
			p.cursor = len(p.aliases) - 1
		}
		return p, nil

	// 'a', 'x' are handled by the app layer.
	}

	return p, nil
}

// View renders the domains panel.
func (p DomainsPanel) View(width, height int, focused bool) string {
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
		Render(" Domains ")

	content := p.renderList(innerWidth, innerHeight)

	return style.
		Width(innerWidth).
		Height(innerHeight).
		Render(title + "\n" + content)
}

func (p DomainsPanel) renderList(width, height int) string {
	var lines []string

	if p.loading {
		lines = append(lines, theme.LoadingStyle.Render("Loading domains..."))
	} else if len(p.aliases) == 0 {
		lines = append(lines, theme.NormalItemStyle.Render("No domain aliases"))
	} else {
		visibleHeight := height - 1
		if visibleHeight < 1 {
			visibleHeight = 1
		}
		startIdx := 0
		if p.cursor >= visibleHeight {
			startIdx = p.cursor - visibleHeight + 1
		}

		for i := startIdx; i < len(p.aliases) && len(lines) < visibleHeight; i++ {
			alias := p.aliases[i]
			line := p.renderAliasLine(alias, i, width)
			lines = append(lines, line)
		}
	}

	for len(lines) < height {
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}

func (p DomainsPanel) renderAliasLine(alias string, idx, maxWidth int) string {
	nameWidth := maxWidth - 6
	if nameWidth < 10 {
		nameWidth = 10
	}
	alias = truncatePlain(alias, nameWidth)

	if idx == p.cursor {
		line := theme.CursorStyle.Render("> ") +
			theme.SelectedItemStyle.Render(alias)
		return theme.Truncate(line, maxWidth)
	}

	line := "  " +
		theme.NormalItemStyle.Render(alias)
	return theme.Truncate(line, maxWidth)
}

// HelpBindings returns the key hints for the domains panel.
func (p DomainsPanel) HelpBindings() []HelpBinding {
	return []HelpBinding{
		{Key: "j/k", Desc: "navigate"},
		{Key: "a", Desc: "add alias"},
		{Key: "x", Desc: "remove"},
		{Key: "g/G", Desc: "top/bottom"},
		{Key: "esc", Desc: "back"},
		{Key: "tab", Desc: "switch panel"},
		{Key: "q", Desc: "quit"},
	}
}
