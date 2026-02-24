package tui

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/key"
	lipgloss "charm.land/lipgloss/v2"

	"github.com/hinke/phorge/internal/config"
	"github.com/hinke/phorge/internal/forge"
	"github.com/hinke/phorge/internal/tui/components"
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
	serverList        panels.ServerList
	siteList          panels.SiteList
	serverInfo        panels.ServerInfo
	siteInfo          panels.SiteInfo
	deploymentsPanel  panels.DeploymentsPanel
	deployScriptPanel panels.DeployScriptPanel
	environmentPanel  panels.EnvironmentPanel
	databasesPanel    panels.DatabasesPanel
	dbUsersPanel      panels.DBUsersPanel
	sslPanel          panels.SSLPanel
	workersPanel      panels.WorkersPanel
	daemonsPanel      panels.DaemonsPanel
	firewallPanel     panels.FirewallPanel
	jobsPanel         panels.JobsPanel
	sshKeysPanel      panels.SSHKeysPanel
	commandsPanel     panels.CommandsPanel
	logsPanel         panels.LogsPanel
	gitPanel          panels.GitPanel
	domainsPanel      panels.DomainsPanel

	// showDeployScript is true when viewing the deploy script sub-view
	// from within the deployments tab.
	showDeployScript bool

	// showDBUsers is true when viewing the database users sub-view
	// from within the databases tab.
	showDBUsers bool

	// Confirmation dialog state.
	confirm *components.Confirm

	// Input dialog state.
	inputDialog *components.Input

	// pendingInputValue stores a value from a multi-step input dialog
	// (e.g. SSH key name before prompting for key content).
	pendingInputValue string

	// Data kept at the app level for cross-panel concerns.
	selectedSrv  *forge.Server
	selectedSite *forge.Site
	activeTab    int // 1-9 for detail section tabs

	// UI state
	toast      string
	toastIsErr bool
	loading    bool

	// tunnelProc holds the SSH tunnel process for database connections.
	// It is killed when the external database client exits.
	tunnelProc *os.Process

	// Help modal overlay.
	helpModal HelpModal

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
		serverInfo:    panels.NewServerInfo(),
		siteInfo:      panels.NewSiteInfo(),
		helpModal:     NewHelpModal(),
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
	// If the help modal is active, route all key events to it.
	if m.helpModal.Active() {
		if _, ok := msg.(tea.KeyPressMsg); ok {
			var cmd tea.Cmd
			m.helpModal, cmd = m.helpModal.Update(msg)
			return m, cmd
		}
	}

	// If an input dialog is active, route all key events to it.
	if m.inputDialog != nil && m.inputDialog.Active {
		if _, ok := msg.(tea.KeyPressMsg); ok {
			i, cmd := m.inputDialog.Update(msg)
			m.inputDialog = &i
			return m, cmd
		}
	}

	// If a confirmation dialog is active, route all key events to it.
	if m.confirm != nil && m.confirm.Active {
		if _, ok := msg.(tea.KeyPressMsg); ok {
			c, cmd := m.confirm.Update(msg)
			m.confirm = &c
			return m, cmd
		}
	}

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
		m.serverInfo = m.serverInfo.SetServer(sel)
		if sel != nil {
			m.siteList = m.siteList.SetServerName(sel.Name).SetLoading(true)
			return m, m.fetchSites(sel.ID)
		}
		return m, nil

	case sitesLoadedMsg:
		m.siteList = m.siteList.SetSites(msg.sites)
		m.selectedSite = m.siteList.Selected()
		m.siteInfo = m.siteInfo.SetSite(m.selectedSite)
		return m, nil

	// Panel-emitted messages: a server was navigated to.
	case panels.ServerSelectedMsg:
		srv := msg.Server
		m.selectedSrv = &srv
		m.serverInfo = m.serverInfo.SetServer(&srv)
		m.selectedSite = nil
		m.siteInfo = m.siteInfo.SetSite(nil)
		m.siteList = m.siteList.SetServerName(srv.Name).SetLoading(true)
		return m, m.fetchSites(srv.ID)

	// Panel-emitted messages: a site was navigated to.
	case panels.SiteSelectedMsg:
		site := msg.Site
		m.selectedSite = &site
		m.siteInfo = m.siteInfo.SetSite(&site)
		// Re-initialise the active detail panel.
		if m.focus == FocusDetailPanel && m.selectedSrv != nil {
			return m.initTabPanel(m.activeTab, m.selectedSrv.ID, site.ID)
		}
		return m, nil

	// Deployment panel messages.
	case panels.DeploymentsLoadedMsg:
		p, cmd := m.deploymentsPanel.Update(msg)
		m.deploymentsPanel = p.(panels.DeploymentsPanel)
		return m, cmd

	case panels.DeployOutputMsg:
		p, cmd := m.deploymentsPanel.Update(msg)
		m.deploymentsPanel = p.(panels.DeploymentsPanel)
		return m, cmd

	case panels.DeployResetMsg:
		if msg.Err != nil {
			m.toast = fmt.Sprintf("Reset failed: %v", msg.Err)
			m.toastIsErr = true
		} else {
			m.toast = "Deployment status reset"
			m.toastIsErr = false
		}
		// Refresh the deployments list after reset.
		return m, tea.Batch(
			m.clearToastAfter(3*time.Second),
			m.deploymentsPanel.LoadDeployments(),
		)

	// Deploy triggered (from deployments panel commands).
	case panels.DeployTriggerMsg:
		m.toast = "Deployment started"
		m.toastIsErr = false
		cmds := []tea.Cmd{m.clearToastAfter(3 * time.Second)}
		if m.activeTab == 1 {
			cmds = append(cmds, m.deploymentsPanel.LoadDeployments())
		}
		return m, tea.Batch(cmds...)

	// Deploy script panel messages.
	case panels.ScriptLoadedMsg:
		p, cmd := m.deployScriptPanel.Update(msg)
		m.deployScriptPanel = p.(panels.DeployScriptPanel)
		return m, cmd

	case panels.ScriptSavedMsg:
		if msg.Err != nil {
			m.toast = fmt.Sprintf("Script save failed: %v", msg.Err)
			m.toastIsErr = true
		} else {
			m.toast = "Deploy script updated"
			m.toastIsErr = false
		}
		return m, m.clearToastAfter(3 * time.Second)

	// Environment panel messages.
	case panels.EnvLoadedMsg:
		p, cmd := m.environmentPanel.Update(msg)
		m.environmentPanel = p.(panels.EnvironmentPanel)
		return m, cmd

	case panels.EnvSavedMsg:
		if msg.Err != nil {
			m.toast = fmt.Sprintf("Environment save failed: %v", msg.Err)
			m.toastIsErr = true
		} else {
			m.toast = "Environment updated"
			m.toastIsErr = false
		}
		return m, m.clearToastAfter(3 * time.Second)

	// Databases panel messages.
	case panels.DatabasesLoadedMsg:
		p, cmd := m.databasesPanel.Update(msg)
		m.databasesPanel = p.(panels.DatabasesPanel)
		return m, cmd

	case panels.DatabaseCreatedMsg:
		m.toast = "Database created"
		m.toastIsErr = false
		return m, tea.Batch(
			m.clearToastAfter(3*time.Second),
			m.databasesPanel.LoadDatabases(),
		)

	case panels.DatabaseDeletedMsg:
		m.toast = "Database deleted"
		m.toastIsErr = false
		return m, tea.Batch(
			m.clearToastAfter(3*time.Second),
			m.databasesPanel.LoadDatabases(),
		)

	// Database users panel messages.
	case panels.DBUsersLoadedMsg:
		p, cmd := m.dbUsersPanel.Update(msg)
		m.dbUsersPanel = p.(panels.DBUsersPanel)
		return m, cmd

	case panels.DBUserCreatedMsg:
		m.toast = "Database user created"
		m.toastIsErr = false
		return m, tea.Batch(
			m.clearToastAfter(3*time.Second),
			m.dbUsersPanel.LoadUsers(),
		)

	case panels.DBUserDeletedMsg:
		m.toast = "Database user deleted"
		m.toastIsErr = false
		return m, tea.Batch(
			m.clearToastAfter(3*time.Second),
			m.dbUsersPanel.LoadUsers(),
		)

	// SSL panel messages.
	case panels.CertsLoadedMsg:
		p, cmd := m.sslPanel.Update(msg)
		m.sslPanel = p.(panels.SSLPanel)
		return m, cmd

	case panels.CertCreatedMsg:
		m.toast = "Certificate created"
		m.toastIsErr = false
		return m, tea.Batch(
			m.clearToastAfter(3*time.Second),
			m.sslPanel.LoadCerts(),
		)

	case panels.CertActivatedMsg:
		m.toast = "Certificate activated"
		m.toastIsErr = false
		return m, tea.Batch(
			m.clearToastAfter(3*time.Second),
			m.sslPanel.LoadCerts(),
		)

	case panels.CertDeletedMsg:
		m.toast = "Certificate deleted"
		m.toastIsErr = false
		return m, tea.Batch(
			m.clearToastAfter(3*time.Second),
			m.sslPanel.LoadCerts(),
		)

	// Workers panel messages.
	case panels.WorkersLoadedMsg:
		p, cmd := m.workersPanel.Update(msg)
		m.workersPanel = p.(panels.WorkersPanel)
		return m, cmd

	case panels.WorkerCreatedMsg:
		m.toast = "Worker created"
		m.toastIsErr = false
		return m, tea.Batch(
			m.clearToastAfter(3*time.Second),
			m.workersPanel.LoadWorkers(),
		)

	case panels.WorkerRestartedMsg:
		m.toast = "Worker restarted"
		m.toastIsErr = false
		return m, tea.Batch(
			m.clearToastAfter(3*time.Second),
			m.workersPanel.LoadWorkers(),
		)

	case panels.WorkerDeletedMsg:
		m.toast = "Worker deleted"
		m.toastIsErr = false
		return m, tea.Batch(
			m.clearToastAfter(3*time.Second),
			m.workersPanel.LoadWorkers(),
		)

	// Daemons panel messages.
	case panels.DaemonsLoadedMsg:
		p, cmd := m.daemonsPanel.Update(msg)
		m.daemonsPanel = p.(panels.DaemonsPanel)
		return m, cmd

	case panels.DaemonCreatedMsg:
		m.toast = "Daemon created"
		m.toastIsErr = false
		return m, tea.Batch(
			m.clearToastAfter(3*time.Second),
			m.daemonsPanel.LoadDaemons(),
		)

	case panels.DaemonRestartedMsg:
		m.toast = "Daemon restarted"
		m.toastIsErr = false
		return m, tea.Batch(
			m.clearToastAfter(3*time.Second),
			m.daemonsPanel.LoadDaemons(),
		)

	case panels.DaemonDeletedMsg:
		m.toast = "Daemon deleted"
		m.toastIsErr = false
		return m, tea.Batch(
			m.clearToastAfter(3*time.Second),
			m.daemonsPanel.LoadDaemons(),
		)

	// Firewall panel messages.
	case panels.FirewallLoadedMsg:
		p, cmd := m.firewallPanel.Update(msg)
		m.firewallPanel = p.(panels.FirewallPanel)
		return m, cmd

	case panels.FirewallCreatedMsg:
		m.toast = "Firewall rule created"
		m.toastIsErr = false
		return m, tea.Batch(
			m.clearToastAfter(3*time.Second),
			m.firewallPanel.LoadRules(),
		)

	case panels.FirewallDeletedMsg:
		m.toast = "Firewall rule deleted"
		m.toastIsErr = false
		return m, tea.Batch(
			m.clearToastAfter(3*time.Second),
			m.firewallPanel.LoadRules(),
		)

	// Jobs panel messages.
	case panels.JobsLoadedMsg:
		p, cmd := m.jobsPanel.Update(msg)
		m.jobsPanel = p.(panels.JobsPanel)
		return m, cmd

	// SSH Keys panel messages.
	case panels.SSHKeysLoadedMsg:
		p, cmd := m.sshKeysPanel.Update(msg)
		m.sshKeysPanel = p.(panels.SSHKeysPanel)
		return m, cmd

	case panels.SSHKeyCreatedMsg:
		m.toast = "SSH key created"
		m.toastIsErr = false
		return m, tea.Batch(
			m.clearToastAfter(3*time.Second),
			m.sshKeysPanel.LoadKeys(),
		)

	case panels.SSHKeyDeletedMsg:
		m.toast = "SSH key deleted"
		m.toastIsErr = false
		return m, tea.Batch(
			m.clearToastAfter(3*time.Second),
			m.sshKeysPanel.LoadKeys(),
		)

	// Commands panel messages.
	case panels.CommandsLoadedMsg:
		p, cmd := m.commandsPanel.Update(msg)
		m.commandsPanel = p.(panels.CommandsPanel)
		return m, cmd

	case panels.CommandCreatedMsg:
		m.toast = "Command executed"
		m.toastIsErr = false
		return m, tea.Batch(
			m.clearToastAfter(3*time.Second),
			m.commandsPanel.LoadCommands(),
		)

	case panels.CommandDetailMsg:
		p, cmd := m.commandsPanel.Update(msg)
		m.commandsPanel = p.(panels.CommandsPanel)
		return m, cmd

	// Logs panel messages.
	case panels.LogsLoadedMsg:
		p, cmd := m.logsPanel.Update(msg)
		m.logsPanel = p.(panels.LogsPanel)
		return m, cmd

	// Domains panel messages.
	case panels.DomainsLoadedMsg:
		p, cmd := m.domainsPanel.Update(msg)
		m.domainsPanel = p.(panels.DomainsPanel)
		return m, cmd

	case panels.DomainsSavedMsg:
		if msg.Err != nil {
			m.toast = fmt.Sprintf("Domain update failed: %v", msg.Err)
			m.toastIsErr = true
		} else {
			m.toast = "Domains updated"
			m.toastIsErr = false
		}
		return m, tea.Batch(
			m.clearToastAfter(3*time.Second),
			m.domainsPanel.RefreshAliases(),
		)

	// Input dialog results.
	case components.InputResult:
		m.inputDialog = nil
		return m.handleInputResult(msg)

	case components.InputCancelled:
		m.inputDialog = nil
		return m, nil

	// Panel-level errors (from panel API commands).
	case panels.PanelErrMsg:
		m.loading = false
		m.toast = fmt.Sprintf("Error: %v", msg.Err)
		m.toastIsErr = true
		return m, m.clearToastAfter(5 * time.Second)

	// Confirmation dialog result.
	case components.ConfirmResult:
		m.confirm = nil
		return m.handleConfirmResult(msg)

	case errMsg:
		m.loading = false
		m.serverList = m.serverList.SetLoading(false)
		m.siteList = m.siteList.SetLoading(false)
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
		// Refresh the deployments list after a successful deploy trigger.
		cmds := []tea.Cmd{m.clearToastAfter(3 * time.Second)}
		if msg.err == nil && m.activeTab == 1 {
			cmds = append(cmds, m.deploymentsPanel.LoadDeployments())
		}
		return m, tea.Batch(cmds...)

	case externalExitMsg:
		// Clean up any lingering tunnel process.
		m.cleanupTunnel()
		if msg.err != nil {
			m.toast = fmt.Sprintf("External process error: %v", msg.err)
			m.toastIsErr = true
			return m, m.clearToastAfter(5 * time.Second)
		}
		return m, nil

	case dbReadyMsg:
		m.toast = ""
		m.toastIsErr = false
		var cmd tea.Cmd
		m, cmd = m.handleDBReady(msg)
		return m, cmd
	}

	return m, nil
}

// handleKey processes key events, routing to global keys first, then focus-specific keys.
func (m App) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	// When a list panel's filter input is active, route all key events
	// directly to that panel (bypass global keys like q, tab, etc.).
	if m.focus == FocusServerList && m.serverList.FilterActive() {
		var cmd tea.Cmd
		panel, cmd := m.serverList.Update(msg)
		m.serverList = panel.(panels.ServerList)
		return m, cmd
	}
	if m.focus == FocusContextList && m.siteList.FilterActive() {
		var cmd tea.Cmd
		panel, cmd := m.siteList.Update(msg)
		m.siteList = panel.(panels.SiteList)
		return m, cmd
	}

	// Global keys take priority.
	switch {
	case key.Matches(msg, m.globalKeys.Quit):
		return m, tea.Quit
	case key.Matches(msg, m.globalKeys.Help):
		m.helpModal = m.helpModal.Toggle()
		return m, nil
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
	case key.Matches(msg, m.globalKeys.SSH):
		cmd := m.sshCmd()
		if cmd == nil {
			m.toast = "No server selected"
			m.toastIsErr = true
			return m, m.clearToastAfter(3 * time.Second)
		}
		return m, cmd
	case key.Matches(msg, m.globalKeys.SFTP):
		cmd := m.sftpCmd()
		if cmd == nil {
			m.toast = "No server selected"
			m.toastIsErr = true
			return m, m.clearToastAfter(3 * time.Second)
		}
		return m, cmd
	case key.Matches(msg, m.globalKeys.Database):
		cmd := m.databaseCmd()
		if cmd == nil {
			m.toast = "Select a server and site first"
			m.toastIsErr = true
			return m, m.clearToastAfter(3 * time.Second)
		}
		m.toast = "Fetching database credentials..."
		m.toastIsErr = false
		return m, cmd
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
		// Initialise the active tab panel when entering the detail panel.
		if m.selectedSite != nil && m.selectedSrv != nil {
			return m.initTabPanel(m.activeTab, m.selectedSrv.ID, m.selectedSite.ID)
		}
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
	// If the deploy script sub-view is active, route keys to it.
	if m.activeTab == 1 && m.selectedSite != nil && m.showDeployScript {
		if key.Matches(msg, m.navKeys.Back) {
			m.showDeployScript = false
			return m, nil
		}
		p, cmd := m.deployScriptPanel.Update(msg)
		m.deployScriptPanel = p.(panels.DeployScriptPanel)
		return m, cmd
	}

	// If the DB users sub-view is active, route keys to it.
	if m.activeTab == 3 && m.showDBUsers {
		if key.Matches(msg, m.navKeys.Back) {
			m.showDBUsers = false
			return m, nil
		}
		return m.handleDBUsersKey(msg)
	}

	// If the commands panel is showing detail and user presses Esc,
	// go back to the commands list (not up to context panel).
	if m.activeTab == 6 && m.selectedSite != nil && m.commandsPanel.ShowingDetail() {
		if key.Matches(msg, m.navKeys.Back) {
			p, cmd := m.commandsPanel.Update(msg)
			m.commandsPanel = p.(panels.CommandsPanel)
			return m, cmd
		}
	}

	// If the deployments panel is showing output and user presses Esc,
	// go back to the deployments list (not up to context panel).
	if m.activeTab == 1 && m.selectedSite != nil && m.deploymentsPanel.ShowingOutput() {
		if key.Matches(msg, m.navKeys.Back) {
			p, cmd := m.deploymentsPanel.Update(msg)
			m.deploymentsPanel = p.(panels.DeploymentsPanel)
			return m, cmd
		}
	}

	switch {
	case key.Matches(msg, m.navKeys.Back):
		m.focus = FocusContextList
		return m, nil

	// Section tab switching (1-9).
	case key.Matches(msg, m.sectionKeys.Deployments):
		return m.switchToTab(1)
	case key.Matches(msg, m.sectionKeys.Environment):
		return m.switchToTab(2)
	case key.Matches(msg, m.sectionKeys.Databases):
		return m.switchToTab(3)
	case key.Matches(msg, m.sectionKeys.SSL):
		return m.switchToTab(4)
	case key.Matches(msg, m.sectionKeys.Workers):
		return m.switchToTab(5)
	case key.Matches(msg, m.sectionKeys.Daemons):
		return m.switchToTab(6)
	case key.Matches(msg, m.sectionKeys.Firewall):
		return m.switchToTab(7)
	case key.Matches(msg, m.sectionKeys.Jobs):
		return m.switchToTab(8)
	case key.Matches(msg, m.sectionKeys.Domains):
		return m.switchToTab(9)
	}

	// If the deployments panel is active, handle deployment-specific keys.
	if m.activeTab == 1 && m.selectedSite != nil {
		return m.handleDeploymentsKey(msg)
	}

	// If the environment panel is active, delegate keys.
	if m.activeTab == 2 && m.selectedSite != nil {
		return m.handleEnvironmentKey(msg)
	}

	// Databases (tab 3) - server-level.
	if m.activeTab == 3 && m.selectedSrv != nil {
		return m.handleDatabasesKey(msg)
	}

	// SSL (tab 4) - site-level.
	if m.activeTab == 4 && m.selectedSite != nil {
		return m.handleSSLKey(msg)
	}

	// Workers (tab 5) - site-level.
	if m.activeTab == 5 && m.selectedSite != nil {
		return m.handleWorkersKey(msg)
	}

	// Tab 6: Commands (site) or Daemons (server).
	if m.activeTab == 6 {
		if m.selectedSite != nil {
			return m.handleCommandsKey(msg)
		}
		if m.selectedSrv != nil {
			return m.handleDaemonsKey(msg)
		}
	}

	// Tab 7: Logs (site) or Firewall (server).
	if m.activeTab == 7 {
		if m.selectedSite != nil {
			p, cmd := m.logsPanel.Update(msg)
			m.logsPanel = p.(panels.LogsPanel)
			return m, cmd
		}
		if m.selectedSrv != nil {
			return m.handleFirewallKey(msg)
		}
	}

	// Tab 8: Git (site, read-only) or Jobs (server, read-only).
	if m.activeTab == 8 {
		if m.selectedSite != nil {
			// Git panel is read-only, no key handling needed.
			return m, nil
		}
		if m.selectedSrv != nil {
			p, cmd := m.jobsPanel.Update(msg)
			m.jobsPanel = p.(panels.JobsPanel)
			return m, cmd
		}
	}

	// Tab 9: Domains (site) or SSH Keys (server).
	if m.activeTab == 9 {
		if m.selectedSite != nil {
			return m.handleDomainsKey(msg)
		}
		if m.selectedSrv != nil {
			return m.handleSSHKeysKey(msg)
		}
	}

	return m, nil
}

// switchToTab changes the active detail tab and initialises the panel if needed.
func (m App) switchToTab(tab int) (tea.Model, tea.Cmd) {
	m.activeTab = tab
	m.showDeployScript = false // always reset sub-view when switching tabs
	m.showDBUsers = false      // always reset sub-view when switching tabs

	if m.selectedSrv == nil {
		return m, nil
	}

	// Server-level tabs work even without a selected site.
	// Site-level tabs require both server and site.
	siteID := int64(0)
	if m.selectedSite != nil {
		siteID = m.selectedSite.ID
	}

	return m.initTabPanel(tab, m.selectedSrv.ID, siteID)
}

// initTabPanel creates and loads the panel for the given tab.
// Tabs 1-5 are always the same: Deploy, Env, DB, SSL, Workers.
// Tabs 6-9 are context-sensitive:
//   - With a site selected: Commands, Logs, Git, Domains
//   - Without a site (server-only): Daemons, Firewall, Jobs, SSH Keys
func (m App) initTabPanel(tab int, serverID, siteID int64) (tea.Model, tea.Cmd) {
	switch tab {
	case 1:
		if siteID == 0 {
			return m, nil
		}
		m.showDeployScript = false
		m.deploymentsPanel = panels.NewDeploymentsPanel(m.forge, serverID, siteID)
		return m, m.deploymentsPanel.LoadDeployments()
	case 2:
		if siteID == 0 {
			return m, nil
		}
		m.environmentPanel = panels.NewEnvironmentPanel(
			m.forge, serverID, siteID, m.config.Editor.Command,
		)
		return m, m.environmentPanel.LoadEnv()
	case 3:
		// Databases are server-level.
		m.showDBUsers = false
		m.databasesPanel = panels.NewDatabasesPanel(m.forge, serverID)
		return m, m.databasesPanel.LoadDatabases()
	case 4:
		if siteID == 0 {
			return m, nil
		}
		m.sslPanel = panels.NewSSLPanel(m.forge, serverID, siteID)
		return m, m.sslPanel.LoadCerts()
	case 5:
		if siteID == 0 {
			return m, nil
		}
		m.workersPanel = panels.NewWorkersPanel(m.forge, serverID, siteID)
		return m, m.workersPanel.LoadWorkers()
	case 6:
		if siteID > 0 {
			// Site context: Commands.
			m.commandsPanel = panels.NewCommandsPanel(m.forge, serverID, siteID)
			return m, m.commandsPanel.LoadCommands()
		}
		// Server context: Daemons.
		m.daemonsPanel = panels.NewDaemonsPanel(m.forge, serverID)
		return m, m.daemonsPanel.LoadDaemons()
	case 7:
		if siteID > 0 {
			// Site context: Logs (site-level).
			m.logsPanel = panels.NewLogsPanel(m.forge, serverID, siteID)
			return m, m.logsPanel.LoadLogs()
		}
		// Server context: Firewall.
		m.firewallPanel = panels.NewFirewallPanel(m.forge, serverID)
		return m, m.firewallPanel.LoadRules()
	case 8:
		if siteID > 0 {
			// Site context: Git info (read-only).
			m.gitPanel = panels.NewGitPanel(m.selectedSite)
			return m, nil
		}
		// Server context: Scheduled jobs.
		m.jobsPanel = panels.NewJobsPanel(m.forge, serverID)
		return m, m.jobsPanel.LoadJobs()
	case 9:
		if siteID > 0 {
			// Site context: Domains.
			aliases := []string{}
			if m.selectedSite != nil {
				aliases = m.selectedSite.Aliases
			}
			m.domainsPanel = panels.NewDomainsPanel(m.forge, serverID, siteID, aliases)
			return m, nil
		}
		// Server context: SSH Keys.
		m.sshKeysPanel = panels.NewSSHKeysPanel(m.forge, serverID)
		return m, m.sshKeysPanel.LoadKeys()
	}
	return m, nil
}

// handleDeploymentsKey handles keys specific to the deployments panel tab.
func (m App) handleDeploymentsKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	// Check for action keys before delegating to the panel.
	switch {
	case key.Matches(msg, m.siteActKeys.Deploy):
		c := components.NewConfirm("deploy", "Deploy site now?")
		m.confirm = &c
		return m, nil

	case key.Matches(msg, key.NewBinding(key.WithKeys("r"))):
		c := components.NewConfirm("reset-deploy", "Reset deployment status?")
		m.confirm = &c
		return m, nil

	case key.Matches(msg, key.NewBinding(key.WithKeys("S"))):
		// Open the deploy script sub-view.
		if m.selectedSrv != nil && m.selectedSite != nil {
			m.showDeployScript = true
			m.deployScriptPanel = panels.NewDeployScriptPanel(
				m.forge, m.selectedSrv.ID, m.selectedSite.ID, m.config.Editor.Command,
			)
			return m, m.deployScriptPanel.LoadScript()
		}
		return m, nil
	}

	// Delegate navigation and other keys to the deployments panel.
	p, cmd := m.deploymentsPanel.Update(msg)
	m.deploymentsPanel = p.(panels.DeploymentsPanel)
	return m, cmd
}

// handleEnvironmentKey handles keys specific to the environment panel tab.
func (m App) handleEnvironmentKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	// Delegate all keys to the environment panel.
	p, cmd := m.environmentPanel.Update(msg)
	m.environmentPanel = p.(panels.EnvironmentPanel)
	return m, cmd
}

// handleDatabasesKey handles keys specific to the databases panel tab.
func (m App) handleDatabasesKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("c"))):
		i := components.NewInput("create-db", "Database name:", "my_database")
		m.inputDialog = &i
		return m, nil

	case key.Matches(msg, key.NewBinding(key.WithKeys("x"))):
		if db := m.databasesPanel.SelectedDatabase(); db != nil {
			c := components.NewConfirm("delete-db", fmt.Sprintf("Delete database %q?", db.Name))
			m.confirm = &c
		}
		return m, nil

	case key.Matches(msg, key.NewBinding(key.WithKeys("u"))):
		if m.selectedSrv != nil {
			m.showDBUsers = true
			m.dbUsersPanel = panels.NewDBUsersPanel(m.forge, m.selectedSrv.ID)
			return m, m.dbUsersPanel.LoadUsers()
		}
		return m, nil
	}

	p, cmd := m.databasesPanel.Update(msg)
	m.databasesPanel = p.(panels.DatabasesPanel)
	return m, cmd
}

// handleDBUsersKey handles keys specific to the database users sub-view.
func (m App) handleDBUsersKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("c"))):
		i := components.NewInput("create-dbuser", "Username:", "forge_user")
		m.inputDialog = &i
		return m, nil

	case key.Matches(msg, key.NewBinding(key.WithKeys("x"))):
		if u := m.dbUsersPanel.SelectedUser(); u != nil {
			c := components.NewConfirm("delete-dbuser", fmt.Sprintf("Delete user %q?", u.Name))
			m.confirm = &c
		}
		return m, nil
	}

	p, cmd := m.dbUsersPanel.Update(msg)
	m.dbUsersPanel = p.(panels.DBUsersPanel)
	return m, cmd
}

// handleSSLKey handles keys specific to the SSL panel tab.
func (m App) handleSSLKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("c"))):
		i := components.NewInput("create-cert", "Domain(s) (comma-separated):", "example.com")
		m.inputDialog = &i
		return m, nil

	case key.Matches(msg, key.NewBinding(key.WithKeys("a"))):
		if cert := m.sslPanel.SelectedCert(); cert != nil {
			c := components.NewConfirm("activate-cert", fmt.Sprintf("Activate certificate for %q?", cert.Domain))
			m.confirm = &c
		}
		return m, nil

	case key.Matches(msg, key.NewBinding(key.WithKeys("x"))):
		if cert := m.sslPanel.SelectedCert(); cert != nil {
			c := components.NewConfirm("delete-cert", fmt.Sprintf("Delete certificate for %q?", cert.Domain))
			m.confirm = &c
		}
		return m, nil
	}

	p, cmd := m.sslPanel.Update(msg)
	m.sslPanel = p.(panels.SSLPanel)
	return m, cmd
}

// handleWorkersKey handles keys specific to the workers panel tab.
func (m App) handleWorkersKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("c"))):
		c := components.NewConfirm("create-worker", "Create worker with defaults (redis/default/1 proc)?")
		m.confirm = &c
		return m, nil

	case key.Matches(msg, key.NewBinding(key.WithKeys("r"))):
		if w := m.workersPanel.SelectedWorker(); w != nil {
			c := components.NewConfirm("restart-worker", fmt.Sprintf("Restart worker %s:%s?", w.Connection, w.Queue))
			m.confirm = &c
		}
		return m, nil

	case key.Matches(msg, key.NewBinding(key.WithKeys("x"))):
		if w := m.workersPanel.SelectedWorker(); w != nil {
			c := components.NewConfirm("delete-worker", fmt.Sprintf("Delete worker %s:%s?", w.Connection, w.Queue))
			m.confirm = &c
		}
		return m, nil
	}

	p, cmd := m.workersPanel.Update(msg)
	m.workersPanel = p.(panels.WorkersPanel)
	return m, cmd
}

// handleDaemonsKey handles keys specific to the daemons panel tab.
func (m App) handleDaemonsKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("c"))):
		i := components.NewInput("create-daemon", "Daemon command:", "php artisan queue:work")
		m.inputDialog = &i
		return m, nil

	case key.Matches(msg, key.NewBinding(key.WithKeys("r"))):
		if d := m.daemonsPanel.SelectedDaemon(); d != nil {
			c := components.NewConfirm("restart-daemon", fmt.Sprintf("Restart daemon %q?", truncateStr(d.Command, 30)))
			m.confirm = &c
		}
		return m, nil

	case key.Matches(msg, key.NewBinding(key.WithKeys("x"))):
		if d := m.daemonsPanel.SelectedDaemon(); d != nil {
			c := components.NewConfirm("delete-daemon", fmt.Sprintf("Delete daemon %q?", truncateStr(d.Command, 30)))
			m.confirm = &c
		}
		return m, nil
	}

	p, cmd := m.daemonsPanel.Update(msg)
	m.daemonsPanel = p.(panels.DaemonsPanel)
	return m, cmd
}

// handleFirewallKey handles keys specific to the firewall panel tab.
func (m App) handleFirewallKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("c"))):
		i := components.NewInput("create-firewall", "Rule name and port (name:port):", "HTTP:80")
		m.inputDialog = &i
		return m, nil

	case key.Matches(msg, key.NewBinding(key.WithKeys("x"))):
		if r := m.firewallPanel.SelectedRule(); r != nil {
			c := components.NewConfirm("delete-firewall", fmt.Sprintf("Delete rule %q?", r.Name))
			m.confirm = &c
		}
		return m, nil
	}

	p, cmd := m.firewallPanel.Update(msg)
	m.firewallPanel = p.(panels.FirewallPanel)
	return m, cmd
}

// handleCommandsKey handles keys specific to the commands panel tab.
func (m App) handleCommandsKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("c"))):
		i := components.NewInput("run-command", "Command to execute:", "php artisan migrate")
		m.inputDialog = &i
		return m, nil
	}

	p, cmd := m.commandsPanel.Update(msg)
	m.commandsPanel = p.(panels.CommandsPanel)
	return m, cmd
}

// handleDomainsKey handles keys specific to the domains panel tab.
func (m App) handleDomainsKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("a"))):
		i := components.NewInput("add-domain", "Domain alias:", "example.com")
		m.inputDialog = &i
		return m, nil

	case key.Matches(msg, key.NewBinding(key.WithKeys("x"))):
		if alias := m.domainsPanel.SelectedAlias(); alias != "" {
			c := components.NewConfirm("remove-domain", fmt.Sprintf("Remove alias %q?", alias))
			m.confirm = &c
		}
		return m, nil
	}

	p, cmd := m.domainsPanel.Update(msg)
	m.domainsPanel = p.(panels.DomainsPanel)
	return m, cmd
}

// handleSSHKeysKey handles keys specific to the SSH keys panel tab.
func (m App) handleSSHKeysKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("c"))):
		i := components.NewInput("create-sshkey", "SSH key name:", "my-key")
		m.inputDialog = &i
		return m, nil

	case key.Matches(msg, key.NewBinding(key.WithKeys("x"))):
		if k := m.sshKeysPanel.SelectedKey(); k != nil {
			c := components.NewConfirm("delete-sshkey", fmt.Sprintf("Delete SSH key %q?", k.Name))
			m.confirm = &c
		}
		return m, nil
	}

	p, cmd := m.sshKeysPanel.Update(msg)
	m.sshKeysPanel = p.(panels.SSHKeysPanel)
	return m, cmd
}

// handleInputResult processes the result of an input dialog.
func (m App) handleInputResult(msg components.InputResult) (tea.Model, tea.Cmd) {
	value := strings.TrimSpace(msg.Value)
	if value == "" {
		return m, nil
	}

	switch msg.ID {
	case "create-db":
		return m, m.databasesPanel.CreateDatabase(value)
	case "create-dbuser":
		// Use the username as both name and password for simplicity.
		return m, m.dbUsersPanel.CreateUser(value, value)
	case "create-cert":
		// Split comma-separated domains.
		domains := strings.Split(value, ",")
		for i := range domains {
			domains[i] = strings.TrimSpace(domains[i])
		}
		return m, m.sslPanel.CreateLetsEncrypt(domains)
	case "create-daemon":
		return m, m.daemonsPanel.CreateDaemon(value)
	case "create-firewall":
		// Parse "name:port" format.
		parts := strings.SplitN(value, ":", 2)
		name := strings.TrimSpace(parts[0])
		port := "80"
		if len(parts) > 1 {
			port = strings.TrimSpace(parts[1])
		}
		return m, m.firewallPanel.CreateRule(name, port)
	case "run-command":
		return m, m.commandsPanel.CreateCommand(value)
	case "add-domain":
		return m, m.domainsPanel.AddAlias(value)
	case "create-sshkey":
		// For SSH key creation, we need the key content too.
		// Store the name and prompt for the key content.
		m.pendingInputValue = value
		i := components.NewInput("create-sshkey-content", "Public key content:", "ssh-rsa AAAA...")
		m.inputDialog = &i
		return m, nil
	case "create-sshkey-content":
		name := m.pendingInputValue
		m.pendingInputValue = ""
		return m, m.sshKeysPanel.CreateKey(name, value, "forge")
	}

	return m, nil
}

// handleConfirmResult processes the result of a confirmation dialog.
func (m App) handleConfirmResult(msg components.ConfirmResult) (tea.Model, tea.Cmd) {
	if !msg.Confirmed {
		return m, nil
	}

	switch msg.ID {
	case "deploy":
		if m.selectedSite != nil && m.selectedSrv != nil {
			m.toast = "Deploying..."
			m.toastIsErr = false
			return m, m.deploymentsPanel.TriggerDeploy()
		}
	case "reset-deploy":
		if m.selectedSite != nil && m.selectedSrv != nil {
			return m, m.deploymentsPanel.ResetDeployStatus()
		}
	case "delete-db":
		return m, m.databasesPanel.DeleteDatabase()
	case "delete-dbuser":
		return m, m.dbUsersPanel.DeleteUser()
	case "activate-cert":
		return m, m.sslPanel.ActivateCert()
	case "delete-cert":
		return m, m.sslPanel.DeleteCert()
	case "create-worker":
		return m, m.workersPanel.CreateWorker()
	case "restart-worker":
		return m, m.workersPanel.RestartWorker()
	case "delete-worker":
		return m, m.workersPanel.DeleteWorker()
	case "restart-daemon":
		return m, m.daemonsPanel.RestartDaemon()
	case "delete-daemon":
		return m, m.daemonsPanel.DeleteDaemon()
	case "delete-firewall":
		return m, m.firewallPanel.DeleteRule()
	case "remove-domain":
		return m, m.domainsPanel.RemoveAlias()
	case "delete-sshkey":
		return m, m.sshKeysPanel.DeleteKey()
	}

	return m, nil
}

// truncateStr truncates a string for display in confirmation dialogs.
func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
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

	// Overlay the input dialog if active.
	if m.inputDialog != nil && m.inputDialog.Active {
		overlay := m.inputDialog.View(m.width, m.height)
		if overlay != "" {
			content = overlay
		}
	}

	// Overlay the confirmation dialog if active.
	if m.confirm != nil && m.confirm.Active {
		overlay := m.confirm.View(m.width, m.height)
		if overlay != "" {
			content = overlay
		}
	}

	// Overlay the help modal if active.
	if m.helpModal.Active() {
		overlay := m.helpModal.View(m.width, m.height)
		if overlay != "" {
			content = overlay
		}
	}

	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

// renderDetailPanel renders the bottom-right detail/preview panel.
// When a site is selected it shows a tab bar with site-level panels;
// when only a server is selected it shows a tab bar with server-level panels;
// otherwise it shows server info.
func (m App) renderDetailPanel(width, height int) string {
	focused := m.focus == FocusDetailPanel

	if m.selectedSite != nil {
		// Render the tab bar as a single line above the section panel.
		tabBar := m.renderTabBar(width)
		tabBarHeight := lipgloss.Height(tabBar)

		// The section panel gets the remaining height below the tab bar.
		sectionHeight := height - tabBarHeight
		if sectionHeight < 2 {
			sectionHeight = 2
		}

		var sectionPanel string
		switch m.activeTab {
		case 1:
			if m.showDeployScript {
				sectionPanel = m.deployScriptPanel.View(width, sectionHeight, focused)
			} else {
				sectionPanel = m.deploymentsPanel.View(width, sectionHeight, focused)
			}
		case 2:
			sectionPanel = m.environmentPanel.View(width, sectionHeight, focused)
		case 3:
			if m.showDBUsers {
				sectionPanel = m.dbUsersPanel.View(width, sectionHeight, focused)
			} else {
				sectionPanel = m.databasesPanel.View(width, sectionHeight, focused)
			}
		case 4:
			sectionPanel = m.sslPanel.View(width, sectionHeight, focused)
		case 5:
			sectionPanel = m.workersPanel.View(width, sectionHeight, focused)
		case 6:
			sectionPanel = m.commandsPanel.View(width, sectionHeight, focused)
		case 7:
			sectionPanel = m.logsPanel.View(width, sectionHeight, focused)
		case 8:
			sectionPanel = m.gitPanel.View(width, sectionHeight, focused)
		case 9:
			sectionPanel = m.domainsPanel.View(width, sectionHeight, focused)
		default:
			sectionPanel = m.siteInfo.View(width, sectionHeight, focused)
		}

		return lipgloss.JoinVertical(lipgloss.Left, tabBar, sectionPanel)
	}

	// Server-only context: show server-level tab bar for tabs 6-9.
	if m.selectedSrv != nil && m.focus == FocusDetailPanel && m.activeTab >= 6 {
		tabBar := m.renderServerTabBar(width)
		tabBarHeight := lipgloss.Height(tabBar)

		sectionHeight := height - tabBarHeight
		if sectionHeight < 2 {
			sectionHeight = 2
		}

		var sectionPanel string
		switch m.activeTab {
		case 6:
			sectionPanel = m.daemonsPanel.View(width, sectionHeight, focused)
		case 7:
			sectionPanel = m.firewallPanel.View(width, sectionHeight, focused)
		case 8:
			sectionPanel = m.jobsPanel.View(width, sectionHeight, focused)
		case 9:
			sectionPanel = m.sshKeysPanel.View(width, sectionHeight, focused)
		default:
			sectionPanel = m.serverInfo.View(width, sectionHeight, focused)
		}

		return lipgloss.JoinVertical(lipgloss.Left, tabBar, sectionPanel)
	}

	return m.serverInfo.View(width, height, focused)
}

// renderTabBar renders the numbered section tabs at the top of the detail panel.
func (m App) renderTabBar(width int) string {
	// Tabs 6-9 change based on context (site selected vs server only).
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

// renderServerTabBar renders the server-level tab bar for tabs 6-9.
func (m App) renderServerTabBar(width int) string {
	tabs := []struct {
		num  int
		name string
	}{
		{6, "Daemons"}, {7, "Firewall"}, {8, "Jobs"}, {9, "SSH Keys"},
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
		if m.selectedSite != nil && m.activeTab == 1 && m.showDeployScript {
			helpBindings = m.deployScriptPanel.HelpBindings()
		} else if m.selectedSite != nil && m.activeTab == 1 {
			helpBindings = m.deploymentsPanel.HelpBindings()
		} else if m.selectedSite != nil && m.activeTab == 2 {
			helpBindings = m.environmentPanel.HelpBindings()
		} else if m.activeTab == 3 && m.showDBUsers {
			helpBindings = m.dbUsersPanel.HelpBindings()
		} else if m.activeTab == 3 {
			helpBindings = m.databasesPanel.HelpBindings()
		} else if m.selectedSite != nil && m.activeTab == 4 {
			helpBindings = m.sslPanel.HelpBindings()
		} else if m.selectedSite != nil && m.activeTab == 5 {
			helpBindings = m.workersPanel.HelpBindings()
		} else if m.activeTab == 6 && m.selectedSite != nil {
			helpBindings = m.commandsPanel.HelpBindings()
		} else if m.activeTab == 6 {
			helpBindings = m.daemonsPanel.HelpBindings()
		} else if m.activeTab == 7 && m.selectedSite != nil {
			helpBindings = m.logsPanel.HelpBindings()
		} else if m.activeTab == 7 {
			helpBindings = m.firewallPanel.HelpBindings()
		} else if m.activeTab == 8 && m.selectedSite != nil {
			helpBindings = m.gitPanel.HelpBindings()
		} else if m.activeTab == 8 {
			helpBindings = m.jobsPanel.HelpBindings()
		} else if m.activeTab == 9 && m.selectedSite != nil {
			helpBindings = m.domainsPanel.HelpBindings()
		} else if m.activeTab == 9 {
			helpBindings = m.sshKeysPanel.HelpBindings()
		} else if m.selectedSite != nil {
			helpBindings = m.siteInfo.HelpBindings()
		} else {
			helpBindings = m.serverInfo.HelpBindings()
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

