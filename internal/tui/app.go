package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/key"
	lipgloss "charm.land/lipgloss/v2"

	"github.com/hinke/phorge/internal/config"
	"github.com/hinke/phorge/internal/forge"
	"github.com/hinke/phorge/internal/tui/panels"
	"github.com/hinke/phorge/internal/tui/theme"
)

// Focus tracks which panel has keyboard focus.
type Focus int

const (
	FocusServerList  Focus = iota
	FocusContextList       // sites or other context items
	FocusDetailPanel
)

// panelCount is the number of focusable panels.
const panelCount = 3

// App is the root bubbletea model for the three-panel lazygit-style layout.
type App struct {
	forge  *forge.Client
	config *config.Config

	focus         Focus
	width, height int

	// Sub-model panels.
	serverList panels.ServerList
	siteList   panels.SiteList

	// Data kept at the app level for cross-panel concerns.
	selectedSrv  *forge.Server
	selectedSite *forge.Site
	activeTab    int // 1-9 for detail section tabs

	// UI state
	toast      string
	toastIsErr bool
	loading    bool

	// Keymaps
	globalKeys    GlobalKeyMap
	navKeys       NavKeyMap
	sectionKeys   SectionKeyMap
	serverActKeys ServerActionKeyMap
	siteActKeys   SiteActionKeyMap
}

// NewApp creates a new App model with the given configuration.
func NewApp(cfg *config.Config) App {
	client := forge.NewClient(cfg.Forge.APIKey)
	return App{
		forge:         client,
		config:        cfg,
		focus:         FocusServerList,
		activeTab:     1,
		serverList:    panels.NewServerList(),
		siteList:      panels.NewSiteList(),
		globalKeys:    DefaultGlobalKeyMap(),
		navKeys:       DefaultNavKeyMap(),
		sectionKeys:   DefaultSectionKeyMap(),
		serverActKeys: DefaultServerActionKeyMap(),
		siteActKeys:   DefaultSiteActionKeyMap(),
	}
}

// Init fetches the initial server list.
func (m App) Init() tea.Cmd {
	return m.fetchServers()
}

// Update handles all incoming messages.
func (m App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyPressMsg:
		return m.handleKey(msg)

	case serversLoadedMsg:
		m.loading = false
		m.serverList = m.serverList.SetServers(msg.servers).SetLoading(false)
		sel := m.serverList.Selected()
		m.selectedSrv = sel
		if sel != nil {
			m.siteList = m.siteList.SetServerName(sel.Name)
			return m, m.fetchSites(sel.ID)
		}
		return m, nil

	case sitesLoadedMsg:
		m.siteList = m.siteList.SetSites(msg.sites)
		m.selectedSite = m.siteList.Selected()
		return m, nil

	// Panel-emitted messages: a server was navigated to.
	case panels.ServerSelectedMsg:
		srv := msg.Server
		m.selectedSrv = &srv
		m.siteList = m.siteList.SetServerName(srv.Name)
		return m, m.fetchSites(srv.ID)

	// Panel-emitted messages: a site was navigated to.
	case panels.SiteSelectedMsg:
		site := msg.Site
		m.selectedSite = &site
		return m, nil

	case errMsg:
		m.loading = false
		m.toast = fmt.Sprintf("Error: %v", msg.err)
		m.toastIsErr = true
		return m, m.clearToastAfter(5 * time.Second)

	case toastMsg:
		m.toast = msg.message
		m.toastIsErr = msg.isError
		return m, m.clearToastAfter(3 * time.Second)

	case clearToastMsg:
		m.toast = ""
		m.toastIsErr = false
		return m, nil

	case rebootResultMsg:
		if msg.err != nil {
			m.toast = fmt.Sprintf("Reboot failed: %v", msg.err)
			m.toastIsErr = true
		} else {
			m.toast = "Server reboot initiated"
			m.toastIsErr = false
		}
		return m, m.clearToastAfter(3 * time.Second)

	case deployResultMsg:
		if msg.err != nil {
			m.toast = fmt.Sprintf("Deploy failed: %v", msg.err)
			m.toastIsErr = true
		} else {
			m.toast = "Deployment started"
			m.toastIsErr = false
		}
		return m, m.clearToastAfter(3 * time.Second)
	}

	return m, nil
}

// handleKey processes key events, routing to global keys first, then focus-specific keys.
func (m App) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	// Global keys take priority.
	switch {
	case key.Matches(msg, m.globalKeys.Quit):
		return m, tea.Quit
	case key.Matches(msg, m.globalKeys.Tab):
		m.focus = (m.focus + 1) % panelCount
		return m, nil
	case key.Matches(msg, m.globalKeys.ShiftTab):
		m.focus = (m.focus + panelCount - 1) % panelCount
		return m, nil
	case key.Matches(msg, m.globalKeys.Refresh):
		m.loading = true
		m.serverList = m.serverList.SetLoading(true)
		return m, m.fetchServers()
	}

	// Panel-specific keys.
	switch m.focus {
	case FocusServerList:
		return m.handleServerListKey(msg)
	case FocusContextList:
		return m.handleContextListKey(msg)
	case FocusDetailPanel:
		return m.handleDetailKey(msg)
	}

	return m, nil
}

// handleServerListKey processes keys when the server list panel is focused.
func (m App) handleServerListKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	// Check for server-specific action keys first (reboot, etc.).
	switch {
	case key.Matches(msg, m.navKeys.Enter):
		m.focus = FocusContextList
		return m, nil
	case key.Matches(msg, m.serverActKeys.Reboot):
		if m.selectedSrv != nil {
			return m, m.rebootServer(m.selectedSrv.ID)
		}
		return m, nil
	}

	// Delegate navigation keys to the server list panel.
	var cmd tea.Cmd
	panel, cmd := m.serverList.Update(msg)
	m.serverList = panel.(panels.ServerList)
	return m, cmd
}

// handleContextListKey processes keys when the context (sites) list is focused.
func (m App) handleContextListKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	// Handle app-level keys (focus changes).
	switch {
	case key.Matches(msg, m.navKeys.Enter):
		m.focus = FocusDetailPanel
		return m, nil
	case key.Matches(msg, m.navKeys.Back):
		m.focus = FocusServerList
		return m, nil
	}

	// Delegate navigation keys to the site list panel.
	var cmd tea.Cmd
	panel, cmd := m.siteList.Update(msg)
	m.siteList = panel.(panels.SiteList)
	return m, cmd
}

// handleDetailKey processes keys when the detail panel is focused.
func (m App) handleDetailKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.navKeys.Back):
		m.focus = FocusContextList
		return m, nil

	// Section tab switching (1-9).
	case key.Matches(msg, m.sectionKeys.Deployments):
		m.activeTab = 1
	case key.Matches(msg, m.sectionKeys.Environment):
		m.activeTab = 2
	case key.Matches(msg, m.sectionKeys.Databases):
		m.activeTab = 3
	case key.Matches(msg, m.sectionKeys.SSL):
		m.activeTab = 4
	case key.Matches(msg, m.sectionKeys.Workers):
		m.activeTab = 5
	case key.Matches(msg, m.sectionKeys.Commands):
		m.activeTab = 6
	case key.Matches(msg, m.sectionKeys.Logs):
		m.activeTab = 7
	case key.Matches(msg, m.sectionKeys.Git):
		m.activeTab = 8
	case key.Matches(msg, m.sectionKeys.Domains):
		m.activeTab = 9
	}

	return m, nil
}

// View renders the three-panel layout with a help bar at the bottom.
func (m App) View() tea.View {
	if m.width == 0 || m.height == 0 {
		v := tea.NewView("Loading...")
		v.AltScreen = true
		return v
	}

	// Reserve space for the help bar (1 line) and optional toast (1 line).
	helpHeight := 1
	toastHeight := 0
	if m.toast != "" {
		toastHeight = 1
	}
	contentHeight := m.height - helpHeight - toastHeight

	// Left panel = ~30% width, right panel = rest.
	leftWidth := m.width * 3 / 10
	if leftWidth < 20 {
		leftWidth = 20
	}
	rightWidth := m.width - leftWidth

	// Build the three panels using sub-models.
	serverPanel := m.serverList.View(leftWidth, contentHeight, m.focus == FocusServerList)
	contextPanel := m.siteList.View(rightWidth, contentHeight/2, m.focus == FocusContextList)
	detailPanel := m.renderDetailPanel(rightWidth, contentHeight-contentHeight/2)

	// Join the right panels vertically.
	rightSide := lipgloss.JoinVertical(lipgloss.Left, contextPanel, detailPanel)

	// Join left and right horizontally.
	mainContent := lipgloss.JoinHorizontal(lipgloss.Top, serverPanel, rightSide)

	// Build the help bar.
	helpBar := m.renderHelpBar()

	// Assemble everything.
	var parts []string
	parts = append(parts, mainContent)
	if m.toast != "" {
		parts = append(parts, m.renderToast())
	}
	parts = append(parts, helpBar)

	content := lipgloss.JoinVertical(lipgloss.Left, parts...)
	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

// renderDetailPanel renders the bottom-right detail/preview panel.
func (m App) renderDetailPanel(width, height int) string {
	style := InactiveBorderStyle
	titleColor := colorSubtle
	if m.focus == FocusDetailPanel {
		style = ActiveBorderStyle
		titleColor = colorPrimary
	}

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(titleColor).
		Render(" " + m.tabName() + " ")

	innerWidth := width - 2
	innerHeight := height - 2
	if innerWidth < 0 {
		innerWidth = 0
	}
	if innerHeight < 0 {
		innerHeight = 0
	}

	var lines []string

	// Show server/site info as a placeholder.
	if m.selectedSrv != nil {
		lines = append(lines, renderKV("Server", m.selectedSrv.Name, innerWidth))
		lines = append(lines, renderKV("IP", m.selectedSrv.IPAddress, innerWidth))
		lines = append(lines, renderKV("PHP", m.selectedSrv.PHPVersion, innerWidth))
		lines = append(lines, renderKV("Provider", m.selectedSrv.Provider, innerWidth))
		lines = append(lines, renderKV("Region", m.selectedSrv.Region, innerWidth))
		lines = append(lines, renderKV("Status", m.selectedSrv.Status, innerWidth))
	}
	if m.selectedSite != nil {
		if len(lines) > 0 {
			lines = append(lines, "")
		}
		lines = append(lines, renderKV("Site", m.selectedSite.Name, innerWidth))
		lines = append(lines, renderKV("Repository", m.selectedSite.Repository, innerWidth))
		lines = append(lines, renderKV("Branch", m.selectedSite.RepositoryBranch, innerWidth))
		lines = append(lines, renderKV("Type", m.selectedSite.ProjectType, innerWidth))
		lines = append(lines, renderKV("Status", m.selectedSite.Status, innerWidth))
	}
	if m.selectedSrv == nil {
		lines = append(lines, NormalItemStyle.Render("No server selected"))
	}

	// Tab bar
	tabBar := m.renderTabBar(innerWidth)
	lines = append([]string{tabBar, ""}, lines...)

	for len(lines) < innerHeight {
		lines = append(lines, "")
	}

	content := strings.Join(lines, "\n")

	return style.
		Width(innerWidth).
		Height(innerHeight).
		Render(title + "\n" + content)
}

// renderTabBar renders the numbered section tabs at the top of the detail panel.
func (m App) renderTabBar(width int) string {
	tabs := []struct {
		num  int
		name string
	}{
		{1, "Deploy"}, {2, "Env"}, {3, "DB"},
		{4, "SSL"}, {5, "Workers"}, {6, "Cmds"},
		{7, "Logs"}, {8, "Git"}, {9, "Domains"},
	}

	var parts []string
	for _, t := range tabs {
		label := fmt.Sprintf("%d:%s", t.num, t.name)
		if t.num == m.activeTab {
			parts = append(parts, SelectedItemStyle.Render(label))
		} else {
			parts = append(parts, HelpBarStyle.Render(label))
		}
	}

	bar := strings.Join(parts, "  ")
	return theme.Truncate(bar, width)
}

// renderHelpBar renders the context-sensitive help bar at the bottom.
func (m App) renderHelpBar() string {
	var helpBindings []panels.HelpBinding

	switch m.focus {
	case FocusServerList:
		helpBindings = m.serverList.HelpBindings()
	case FocusContextList:
		helpBindings = m.siteList.HelpBindings()
	case FocusDetailPanel:
		helpBindings = []panels.HelpBinding{
			{Key: "1-9", Desc: "sections"},
			{Key: "esc", Desc: "back"},
			{Key: "tab", Desc: "switch panel"},
			{Key: "q", Desc: "quit"},
		}
	}

	var formatted []string
	for _, b := range helpBindings {
		formatted = append(formatted, helpBinding(b.Key, b.Desc))
	}

	bar := strings.Join(formatted, "  ")

	// Pad to full width.
	barWidth := lipgloss.Width(bar)
	if barWidth < m.width {
		bar += strings.Repeat(" ", m.width-barWidth)
	}

	return HelpBarStyle.Render(bar)
}

// renderToast renders the toast notification bar.
func (m App) renderToast() string {
	style := ToastStyle
	if m.toastIsErr {
		style = ToastErrorStyle
	}
	return style.Width(m.width).Render(m.toast)
}

// tabName returns the display name for the current active tab.
func (m App) tabName() string {
	names := map[int]string{
		1: "Deployments",
		2: "Environment",
		3: "Databases",
		4: "SSL",
		5: "Workers",
		6: "Commands",
		7: "Logs",
		8: "Git",
		9: "Domains",
	}
	if name, ok := names[m.activeTab]; ok {
		return name
	}
	return "Detail"
}

// --- Commands (tea.Cmd factories) ---

// fetchServers returns a command that fetches the server list from the API.
func (m App) fetchServers() tea.Cmd {
	client := m.forge
	return func() tea.Msg {
		servers, err := client.Servers.List(context.Background())
		if err != nil {
			return errMsg{err}
		}
		return serversLoadedMsg{servers}
	}
}

// fetchSites returns a command that fetches the sites for a server.
func (m App) fetchSites(serverID int64) tea.Cmd {
	client := m.forge
	return func() tea.Msg {
		sites, err := client.Sites.List(context.Background(), serverID)
		if err != nil {
			return errMsg{err}
		}
		return sitesLoadedMsg{sites}
	}
}

// rebootServer returns a command that initiates a server reboot.
func (m App) rebootServer(serverID int64) tea.Cmd {
	client := m.forge
	return func() tea.Msg {
		err := client.Servers.Reboot(context.Background(), serverID)
		return rebootResultMsg{err}
	}
}

// clearToastAfter returns a command that clears the toast after a delay.
func (m App) clearToastAfter(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(time.Time) tea.Msg {
		return clearToastMsg{}
	})
}

// --- Helpers ---

// helpBinding formats a single key-description pair for the help bar.
func helpBinding(k, desc string) string {
	return HelpKeyStyle.Render(k) + " " + HelpBarStyle.Render(desc)
}

// renderKV renders a label-value pair for the detail panel.
func renderKV(label, value string, maxWidth int) string {
	if value == "" {
		value = "-"
	}
	l := LabelStyle.Render(label + ":")
	v := ValueStyle.Render(value)
	line := l + " " + v
	return theme.Truncate(line, maxWidth)
}
