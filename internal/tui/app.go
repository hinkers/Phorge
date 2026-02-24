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

	// Data
	servers      []forge.Server
	selectedSrv  *forge.Server
	sites        []forge.Site
	selectedSite *forge.Site
	activeTab    int // 1-9 for detail section tabs

	// Cursors
	serverCursor int
	siteCursor   int

	// UI state
	toast      string
	toastIsErr bool
	loading    bool

	// Keymaps
	globalKeys     GlobalKeyMap
	navKeys        NavKeyMap
	sectionKeys    SectionKeyMap
	serverActKeys  ServerActionKeyMap
	siteActKeys    SiteActionKeyMap
}

// NewApp creates a new App model with the given configuration.
func NewApp(cfg *config.Config) App {
	client := forge.NewClient(cfg.Forge.APIKey)
	return App{
		forge:         client,
		config:        cfg,
		focus:         FocusServerList,
		activeTab:     1,
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
		m.servers = msg.servers
		m.loading = false
		m.serverCursor = 0
		if len(m.servers) > 0 {
			srv := m.servers[0]
			m.selectedSrv = &srv
			return m, m.fetchSites(srv.ID)
		}
		return m, nil

	case sitesLoadedMsg:
		m.sites = msg.sites
		m.siteCursor = 0
		if len(m.sites) > 0 {
			site := m.sites[0]
			m.selectedSite = &site
		} else {
			m.selectedSite = nil
		}
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
	switch {
	case key.Matches(msg, m.navKeys.Down):
		if len(m.servers) > 0 {
			m.serverCursor = min(m.serverCursor+1, len(m.servers)-1)
			srv := m.servers[m.serverCursor]
			m.selectedSrv = &srv
			return m, m.fetchSites(srv.ID)
		}
	case key.Matches(msg, m.navKeys.Up):
		if len(m.servers) > 0 {
			m.serverCursor = max(m.serverCursor-1, 0)
			srv := m.servers[m.serverCursor]
			m.selectedSrv = &srv
			return m, m.fetchSites(srv.ID)
		}
	case key.Matches(msg, m.navKeys.Enter):
		m.focus = FocusContextList
		return m, nil
	case key.Matches(msg, m.navKeys.Home):
		if len(m.servers) > 0 {
			m.serverCursor = 0
			srv := m.servers[0]
			m.selectedSrv = &srv
			return m, m.fetchSites(srv.ID)
		}
	case key.Matches(msg, m.navKeys.End):
		if len(m.servers) > 0 {
			m.serverCursor = len(m.servers) - 1
			srv := m.servers[m.serverCursor]
			m.selectedSrv = &srv
			return m, m.fetchSites(srv.ID)
		}
	case key.Matches(msg, m.serverActKeys.Reboot):
		if m.selectedSrv != nil {
			return m, m.rebootServer(m.selectedSrv.ID)
		}
	}

	return m, nil
}

// handleContextListKey processes keys when the context (sites) list is focused.
func (m App) handleContextListKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.navKeys.Down):
		if len(m.sites) > 0 {
			m.siteCursor = min(m.siteCursor+1, len(m.sites)-1)
			site := m.sites[m.siteCursor]
			m.selectedSite = &site
		}
	case key.Matches(msg, m.navKeys.Up):
		if len(m.sites) > 0 {
			m.siteCursor = max(m.siteCursor-1, 0)
			site := m.sites[m.siteCursor]
			m.selectedSite = &site
		}
	case key.Matches(msg, m.navKeys.Enter):
		m.focus = FocusDetailPanel
		return m, nil
	case key.Matches(msg, m.navKeys.Back):
		m.focus = FocusServerList
		return m, nil
	case key.Matches(msg, m.navKeys.Home):
		if len(m.sites) > 0 {
			m.siteCursor = 0
			site := m.sites[0]
			m.selectedSite = &site
		}
	case key.Matches(msg, m.navKeys.End):
		if len(m.sites) > 0 {
			m.siteCursor = len(m.sites) - 1
			site := m.sites[m.siteCursor]
			m.selectedSite = &site
		}
	}

	return m, nil
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

	// Build the three panels.
	serverPanel := m.renderServerList(leftWidth, contentHeight)
	contextPanel := m.renderContextList(rightWidth, contentHeight/2)
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

// renderServerList renders the left-side server list panel.
func (m App) renderServerList(width, height int) string {
	style := InactiveBorderStyle
	titleColor := colorSubtle
	if m.focus == FocusServerList {
		style = ActiveBorderStyle
		titleColor = colorPrimary
	}

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(titleColor).
		Render(" Servers ")

	// Account for border size.
	innerWidth := width - 2
	innerHeight := height - 2
	if innerWidth < 0 {
		innerWidth = 0
	}
	if innerHeight < 0 {
		innerHeight = 0
	}

	var lines []string

	if m.loading && len(m.servers) == 0 {
		lines = append(lines, LoadingStyle.Render("Loading servers..."))
	} else if len(m.servers) == 0 {
		lines = append(lines, NormalItemStyle.Render("No servers found"))
	} else {
		for i, srv := range m.servers {
			name := truncate(srv.Name, innerWidth-4)
			if i == m.serverCursor {
				line := CursorStyle.Render("> ") + SelectedItemStyle.Render(name)
				lines = append(lines, line)
			} else {
				line := "  " + NormalItemStyle.Render(name)
				lines = append(lines, line)
			}
			if i >= innerHeight-1 {
				break
			}
		}
	}

	// Pad to fill the panel height.
	for len(lines) < innerHeight {
		lines = append(lines, "")
	}

	content := strings.Join(lines, "\n")

	return style.
		Width(innerWidth).
		Height(innerHeight).
		Render(title + "\n" + content)
}

// renderContextList renders the top-right sites/context list panel.
func (m App) renderContextList(width, height int) string {
	style := InactiveBorderStyle
	titleColor := colorSubtle
	if m.focus == FocusContextList {
		style = ActiveBorderStyle
		titleColor = colorPrimary
	}

	panelTitle := "Sites"
	if m.selectedSrv != nil {
		panelTitle = fmt.Sprintf("Sites (%s)", m.selectedSrv.Name)
	}
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(titleColor).
		Render(" " + panelTitle + " ")

	innerWidth := width - 2
	innerHeight := height - 2
	if innerWidth < 0 {
		innerWidth = 0
	}
	if innerHeight < 0 {
		innerHeight = 0
	}

	var lines []string

	if m.selectedSrv == nil {
		lines = append(lines, NormalItemStyle.Render("Select a server"))
	} else if len(m.sites) == 0 {
		lines = append(lines, NormalItemStyle.Render("No sites found"))
	} else {
		for i, site := range m.sites {
			name := truncate(site.Name, innerWidth-4)
			if i == m.siteCursor {
				line := CursorStyle.Render("> ") + SelectedItemStyle.Render(name)
				lines = append(lines, line)
			} else {
				line := "  " + NormalItemStyle.Render(name)
				lines = append(lines, line)
			}
			if i >= innerHeight-1 {
				break
			}
		}
	}

	for len(lines) < innerHeight {
		lines = append(lines, "")
	}

	content := strings.Join(lines, "\n")

	return style.
		Width(innerWidth).
		Height(innerHeight).
		Render(title + "\n" + content)
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
	return truncate(bar, width)
}

// renderHelpBar renders the context-sensitive help bar at the bottom.
func (m App) renderHelpBar() string {
	var bindings []string

	switch m.focus {
	case FocusServerList:
		bindings = []string{
			helpBinding("j/k", "navigate"),
			helpBinding("enter", "select"),
			helpBinding("r", "reboot"),
			helpBinding("tab", "switch panel"),
			helpBinding("ctrl+r", "refresh"),
			helpBinding("?", "help"),
			helpBinding("q", "quit"),
		}
	case FocusContextList:
		bindings = []string{
			helpBinding("j/k", "navigate"),
			helpBinding("enter", "select"),
			helpBinding("esc", "back"),
			helpBinding("tab", "switch panel"),
			helpBinding("q", "quit"),
		}
	case FocusDetailPanel:
		bindings = []string{
			helpBinding("1-9", "sections"),
			helpBinding("esc", "back"),
			helpBinding("tab", "switch panel"),
			helpBinding("q", "quit"),
		}
	}

	bar := strings.Join(bindings, "  ")

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
	return truncate(line, maxWidth)
}

// truncate shortens a string to fit within the given width, accounting for
// ANSI escape sequences by using lipgloss.Width for measurement.
func truncate(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	w := lipgloss.Width(s)
	if w <= maxWidth {
		return s
	}
	// Brute-force truncation: trim runes until we fit.
	runes := []rune(s)
	for len(runes) > 0 {
		runes = runes[:len(runes)-1]
		candidate := string(runes) + "..."
		if lipgloss.Width(candidate) <= maxWidth {
			return candidate
		}
	}
	return ""
}
