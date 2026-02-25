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

// FirewallLoadedMsg is sent when the firewall rule list has been fetched.
type FirewallLoadedMsg struct {
	Rules []forge.FirewallRule
}

// FirewallCreatedMsg is sent when a firewall rule has been created.
type FirewallCreatedMsg struct {
	Rule *forge.FirewallRule
}

// FirewallDeletedMsg is sent when a firewall rule has been deleted.
type FirewallDeletedMsg struct{}

// FirewallPanel shows the firewall rules on a server with CRUD actions.
// Firewall rules are server-level resources.
type FirewallPanel struct {
	client   *forge.Client
	serverID int64

	rules   []forge.FirewallRule
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

// NewFirewallPanel creates a new FirewallPanel.
func NewFirewallPanel(client *forge.Client, serverID int64) FirewallPanel {
	return FirewallPanel{
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

// LoadRules returns a tea.Cmd that fetches the firewall rule list.
func (p FirewallPanel) LoadRules() tea.Cmd {
	client := p.client
	serverID := p.serverID
	return func() tea.Msg {
		rules, err := client.Firewall.List(context.Background(), serverID)
		if err != nil {
			return PanelErrMsg{Err: err}
		}
		return FirewallLoadedMsg{Rules: rules}
	}
}

// CreateRule returns a tea.Cmd that creates a new firewall rule.
func (p FirewallPanel) CreateRule(name string, port string) tea.Cmd {
	client := p.client
	serverID := p.serverID
	return func() tea.Msg {
		opts := forge.FirewallCreateOpts{
			Name: name,
			Port: port,
			Type: "allow",
		}
		rule, err := client.Firewall.Create(context.Background(), serverID, opts)
		if err != nil {
			return PanelErrMsg{Err: err}
		}
		return FirewallCreatedMsg{Rule: rule}
	}
}

// DeleteRule returns a tea.Cmd that deletes the currently selected firewall rule.
func (p FirewallPanel) DeleteRule() tea.Cmd {
	if len(p.rules) == 0 || p.cursor >= len(p.rules) {
		return nil
	}
	client := p.client
	serverID := p.serverID
	ruleID := p.rules[p.cursor].ID
	return func() tea.Msg {
		err := client.Firewall.Delete(context.Background(), serverID, ruleID)
		if err != nil {
			return PanelErrMsg{Err: err}
		}
		return FirewallDeletedMsg{}
	}
}

// SelectedRule returns the currently selected firewall rule, or nil.
func (p FirewallPanel) SelectedRule() *forge.FirewallRule {
	if len(p.rules) == 0 || p.cursor >= len(p.rules) {
		return nil
	}
	r := p.rules[p.cursor]
	return &r
}

// Update handles messages for the firewall panel.
func (p FirewallPanel) Update(msg tea.Msg) (Panel, tea.Cmd) {
	switch msg := msg.(type) {
	case FirewallLoadedMsg:
		p.rules = msg.Rules
		p.loading = false
		p.cursor = 0
		return p, nil

	case tea.KeyPressMsg:
		return p.handleKey(msg)
	}

	return p, nil
}

func (p FirewallPanel) handleKey(msg tea.KeyPressMsg) (Panel, tea.Cmd) {
	switch {
	case key.Matches(msg, p.down):
		if len(p.rules) > 0 {
			p.cursor = min(p.cursor+1, len(p.rules)-1)
		}
		return p, nil

	case key.Matches(msg, p.up):
		if len(p.rules) > 0 {
			p.cursor = max(p.cursor-1, 0)
		}
		return p, nil

	case key.Matches(msg, p.home):
		p.cursor = 0
		return p, nil

	case key.Matches(msg, p.end):
		if len(p.rules) > 0 {
			p.cursor = len(p.rules) - 1
		}
		return p, nil

	// 'c', 'x' are handled by the app layer.
	}

	return p, nil
}

// View renders the firewall panel.
func (p FirewallPanel) View(width, height int, focused bool) string {
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
		Render(" Firewall Rules ")

	content := p.renderList(innerWidth, innerHeight-1)

	return style.
		Width(innerWidth).
		Height(innerHeight).
		Render(title + "\n" + content)
}

func (p FirewallPanel) renderList(width, height int) string {
	var lines []string

	if p.loading && len(p.rules) == 0 {
		lines = append(lines, theme.LoadingStyle.Render("Loading firewall rules..."))
	} else if len(p.rules) == 0 {
		lines = append(lines, theme.NormalItemStyle.Render("No firewall rules found"))
	} else {
		visibleHeight := height - 1
		if visibleHeight < 1 {
			visibleHeight = 1
		}
		startIdx := 0
		if p.cursor >= visibleHeight {
			startIdx = p.cursor - visibleHeight + 1
		}

		for i := startIdx; i < len(p.rules) && len(lines) < visibleHeight; i++ {
			r := p.rules[i]
			line := p.renderRuleLine(r, i, width)
			lines = append(lines, line)
		}
	}

	for len(lines) < height {
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}

func (p FirewallPanel) renderRuleLine(r forge.FirewallRule, idx, maxWidth int) string {
	icon := statusIcon(r.Status)

	name := r.Name
	if name == "" {
		name = "-"
	}
	port := fmt.Sprintf("%v", r.Port)
	ip := r.IPAddress
	if ip == "" {
		ip = "any"
	}
	ruleType := r.Type
	if ruleType == "" {
		ruleType = "allow"
	}
	statusStr := fmt.Sprintf(" [%s]", r.Status)

	// Leave room for: cursor(2) + icon(2) + port(~8) + ip(~16) + type(~6) + status(~14) + spacing(10)
	overhead := 58
	nameWidth := maxWidth - overhead
	if nameWidth < 8 {
		nameWidth = 8
	}
	name = truncatePlain(name, nameWidth)

	portStr := fmt.Sprintf("%-6s", truncatePlain(port, 6))
	ipStr := fmt.Sprintf("%-15s", truncatePlain(ip, 15))
	typeStr := fmt.Sprintf("%-5s", truncatePlain(ruleType, 5))

	if idx == p.cursor {
		line := theme.CursorStyle.Render("> ") +
			icon + " " +
			theme.SelectedItemStyle.Render(name) +
			"  " + theme.NormalItemStyle.Render(portStr) +
			"  " + theme.NormalItemStyle.Render(ipStr) +
			"  " + theme.NormalItemStyle.Render(typeStr) +
			"  " + theme.NormalItemStyle.Render(statusStr)
		return theme.Truncate(line, maxWidth)
	}

	line := "  " +
		icon + " " +
		theme.NormalItemStyle.Render(name) +
		"  " + theme.NormalItemStyle.Render(portStr) +
		"  " + theme.NormalItemStyle.Render(ipStr) +
		"  " + theme.NormalItemStyle.Render(typeStr) +
		"  " + theme.NormalItemStyle.Render(statusStr)
	return theme.Truncate(line, maxWidth)
}

// HelpBindings returns the key hints for the firewall panel.
func (p FirewallPanel) HelpBindings() []HelpBinding {
	return []HelpBinding{
		{Key: "j/k", Desc: "navigate"},
		{Key: "c", Desc: "create rule"},
		{Key: "x", Desc: "delete"},
		{Key: "g/G", Desc: "top/bottom"},
		{Key: "esc", Desc: "back"},
		{Key: "tab", Desc: "switch panel"},
		{Key: "q", Desc: "quit"},
	}
}
