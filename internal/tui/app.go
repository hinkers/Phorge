package tui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/key"
	"github.com/charmbracelet/x/ansi"
	lipgloss "charm.land/lipgloss/v2"

	"github.com/hinkers/Phorge/internal/config"
	"github.com/hinkers/Phorge/internal/forge"
	"github.com/hinkers/Phorge/internal/tui/components"
	"github.com/hinkers/Phorge/internal/tui/panels"
	"github.com/hinkers/Phorge/internal/tui/theme"
)

// Focus tracks which panel has keyboard focus.
type Focus int

const (
	FocusTree   Focus = iota
	FocusDetail
	FocusOutput
)

// panelCount is the number of focusable panels.
const panelCount = 3

// App is the root bubbletea model for the lazygit-style layout.
type App struct {
	forge   *forge.Client
	config  *config.Config
	project config.ProjectConfig

	focus         Focus
	width, height int

	// Sub-model panels.
	treePanel         panels.TreePanel
	outputPanel       panels.OutputPanel
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

	// Output polling state for auto-updating deployment/command output.
	outputPoll outputPollState

	// Keymaps
	globalKeys    GlobalKeyMap
	navKeys       NavKeyMap
	sectionKeys   SectionKeyMap
	serverActKeys ServerActionKeyMap
	siteActKeys   SiteActionKeyMap
}

// treeSitesLoadedMsg is sent when sites for a server have been fetched
// in response to expanding a tree node.
type treeSitesLoadedMsg struct {
	serverID int64
	sites    []forge.Site
}

// outputPollState tracks the active output polling context.
type outputPollState struct {
	serverID     int64
	siteID       int64
	deploymentID int64
	active       bool
}

// NewApp creates a new App model with the given configuration.
func NewApp(cfg *config.Config) App {
	client := forge.NewClient(cfg.Forge.APIKey)
	project := config.LoadProjectConfig()
	return App{
		forge:       client,
		config:      cfg,
		project:     project,
		focus:       FocusTree,
		activeTab:   1,
		treePanel:   panels.NewTreePanel().SetDefaultServer(project.Server).SetDefaultSite(project.Site),
		outputPanel: panels.NewOutputPanel(),
		serverInfo:  panels.NewServerInfo(),
		siteInfo:    panels.NewSiteInfo(),
		helpModal:   NewHelpModal(),
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
		m.treePanel = m.treePanel.SetServers(msg.servers).SetLoading(false)

		// Auto-expand default server from per-directory .phorge config.
		var cmds []tea.Cmd
		if defaultName := m.project.Server; defaultName != "" {
			if srv := m.treePanel.FindServerByName(defaultName); srv != nil {
				var cmd tea.Cmd
				m.treePanel, cmd = m.treePanel.ExpandServer(srv.ID)
				if cmd != nil {
					cmds = append(cmds, cmd)
				}
				var found bool
				m.treePanel, found = m.treePanel.SetCursorToServer(srv.ID)
				_ = found
			}
		}

		// Select whatever the cursor is on.
		srv, site := m.treePanel.Selected()
		if srv.ID != 0 {
			m.selectedSrv = &srv
			m.serverInfo = m.serverInfo.SetServer(&srv)
		} else {
			m.selectedSrv = nil
			m.serverInfo = m.serverInfo.SetServer(nil)
		}
		m.selectedSite = site
		m.siteInfo = m.siteInfo.SetSite(site)
		if len(cmds) > 0 {
			return m, tea.Batch(cmds...)
		}
		return m, nil

	// Tree panel: user navigated to a node.
	case panels.TreeNodeSelectedMsg:
		srv := msg.Server
		m.selectedSrv = &srv
		m.serverInfo = m.serverInfo.SetServer(&srv)
		if msg.Site != nil {
			site := *msg.Site
			m.selectedSite = &site
			m.siteInfo = m.siteInfo.SetSite(&site)
		} else {
			m.selectedSite = nil
			m.siteInfo = m.siteInfo.SetSite(nil)
		}
		return m, nil

	// Tree panel: needs sites for a server.
	case panels.TreeFetchSitesMsg:
		m.treePanel = m.treePanel.SetSitesLoading(msg.ServerID)
		return m, m.fetchSitesForTree(msg.ServerID)

	// Sites loaded for tree expansion.
	case treeSitesLoadedMsg:
		m.treePanel = m.treePanel.SetSites(msg.serverID, msg.sites)

		// If a default site is configured, navigate to it when its server's
		// sites are first loaded.
		if defaultSite := m.project.Site; defaultSite != "" {
			if defaultServer := m.project.Server; defaultServer != "" {
				if srv := m.treePanel.FindServerByName(defaultServer); srv != nil && srv.ID == msg.serverID {
					_, site := m.treePanel.FindSiteByName(defaultSite)
					if site != nil {
						m.treePanel, _ = m.treePanel.SetCursorToSite(site.ID)
						m.selectedSite = site
						m.siteInfo = m.siteInfo.SetSite(site)
					}
				}
			}
		}
		return m, nil

	// Deployment panel messages.
	case panels.DeploymentsLoadedMsg:
		p, cmd := m.deploymentsPanel.Update(msg)
		m.deploymentsPanel = p.(panels.DeploymentsPanel)
		return m, cmd

	// User pressed Enter on a deployment to view output.
	case panels.DeployViewOutputMsg:
		// Start polling if the deployment might still be running.
		m.outputPoll = outputPollState{
			serverID:     msg.ServerID,
			siteID:       msg.SiteID,
			deploymentID: msg.DeploymentID,
			active:       true,
		}
		return m, m.fetchDeployOutputWithStatus(msg.ServerID, msg.SiteID, msg.DeploymentID)

	// Deploy output fetched — route to output panel.
	case panels.DeployOutputMsg:
		m.outputPanel = m.outputPanel.SetContent("Deploy Output", msg.Output)
		m.focus = FocusOutput
		return m, nil

	// Polled output+status result.
	case pollOutputResultMsg:
		title := "Deploy Output"
		if !msg.finished {
			title = "Deploy Output (deploying...)"
		}
		m.outputPanel = m.outputPanel.SetContent(title, msg.output)
		m.focus = FocusOutput
		if msg.finished {
			m.outputPoll.active = false
			// Refresh the deployments list to show updated status.
			if m.activeTab == 1 {
				return m, m.deploymentsPanel.LoadDeployments()
			}
			return m, nil
		}
		// Continue polling.
		return m, m.pollOutputTick()

	// Poll timer fired.
	case pollOutputTickMsg:
		if !m.outputPoll.active {
			return m, nil
		}
		return m, m.fetchDeployOutputWithStatus(
			m.outputPoll.serverID,
			m.outputPoll.siteID,
			m.outputPoll.deploymentID,
		)

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

	case setDefaultMsg:
		if msg.err != nil {
			m.toast = fmt.Sprintf("Failed to save default: %v", msg.err)
			m.toastIsErr = true
		} else if msg.serverName == "" && msg.siteName == "" {
			m.project.Server = ""
			m.project.Site = ""
			m.treePanel = m.treePanel.SetDefaultServer("").SetDefaultSite("")
			m.toast = "Cleared default"
			m.toastIsErr = false
		} else {
			m.project.Server = msg.serverName
			m.project.Site = msg.siteName
			m.treePanel = m.treePanel.SetDefaultServer(msg.serverName).SetDefaultSite(msg.siteName)
			if msg.siteName != "" {
				m.toast = fmt.Sprintf("Set %s/%s as default", msg.serverName, msg.siteName)
			} else {
				m.toast = fmt.Sprintf("Set %s as default server", msg.serverName)
			}
			m.toastIsErr = false
		}
		return m, m.clearToastAfter(3 * time.Second)

	case errMsg:
		m.loading = false
		m.treePanel = m.treePanel.SetLoading(false)
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
	// When the tree panel's filter is active, route all key events to it.
	if m.focus == FocusTree && m.treePanel.FilterActive() {
		var cmd tea.Cmd
		panel, cmd := m.treePanel.Update(msg)
		m.treePanel = panel.(panels.TreePanel)
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
		m.treePanel = m.treePanel.SetLoading(true)
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
	case FocusTree:
		return m.handleTreeKey(msg)
	case FocusDetail:
		return m.handleDetailKey(msg)
	case FocusOutput:
		return m.handleOutputKey(msg)
	}

	return m, nil
}

// handleTreeKey processes keys when the tree panel is focused.
func (m App) handleTreeKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	onServer := m.treePanel.CursorOnServer()

	// Server-level action keys.
	if onServer && m.selectedSrv != nil {
		switch {
		case key.Matches(msg, m.serverActKeys.Reboot):
			return m, m.rebootServer(m.selectedSrv.ID)
		case key.Matches(msg, m.serverActKeys.SSH):
			cmd := m.sshCmd()
			if cmd != nil {
				return m, cmd
			}
			return m, nil
		case key.Matches(msg, m.serverActKeys.SFTP):
			cmd := m.sftpCmd()
			if cmd != nil {
				return m, cmd
			}
			return m, nil
		case key.Matches(msg, key.NewBinding(key.WithKeys("D"))):
			// Toggle default server for this directory (.phorge file).
			return m, m.toggleDefault(m.selectedSrv.Name, "")
		}
	}

	// Site-level action keys.
	if !onServer && m.selectedSite != nil && m.selectedSrv != nil {
		switch {
		case key.Matches(msg, m.siteActKeys.Deploy):
			c := components.NewConfirm("deploy", "Deploy site now?")
			m.confirm = &c
			return m, nil
		case key.Matches(msg, m.siteActKeys.SSH):
			cmd := m.sshCmd()
			if cmd != nil {
				return m, cmd
			}
			return m, nil
		case key.Matches(msg, key.NewBinding(key.WithKeys("D"))):
			// Toggle default site for this directory (.phorge file).
			return m, m.toggleDefault(m.selectedSrv.Name, m.selectedSite.Name)
		}
	}

	// Enter focuses the detail panel for both server and site nodes.
	if key.Matches(msg, m.navKeys.Enter) {
		srv, site := m.treePanel.Selected()
		if site != nil {
			m.focus = FocusDetail
			if m.selectedSrv != nil {
				return m.initTabPanel(m.activeTab, m.selectedSrv.ID, site.ID)
			}
			return m, nil
		}
		if srv.ID != 0 {
			// Server node: if not expanded, expand first to load sites.
			if !m.treePanel.IsExpanded(srv.ID) {
				var cmd tea.Cmd
				panel, cmd := m.treePanel.Update(msg)
				m.treePanel = panel.(panels.TreePanel)
				return m, cmd
			}
			// Already expanded: focus the detail panel.
			m.focus = FocusDetail
			return m, nil
		}
	}

	// l on a site node focuses the detail panel (vim-style: right = drill in).
	if key.Matches(msg, key.NewBinding(key.WithKeys("l", "right"))) {
		_, site := m.treePanel.Selected()
		if site != nil {
			m.focus = FocusDetail
			if m.selectedSrv != nil {
				return m.initTabPanel(m.activeTab, m.selectedSrv.ID, site.ID)
			}
			return m, nil
		}
		// For server nodes, fall through to tree panel (which handles expand).
	}

	// Server-level tab switching from tree (so the detail panel updates).
	if onServer && m.selectedSrv != nil {
		switch {
		case key.Matches(msg, m.sectionKeys.Databases):
			return m.switchToServerTab(3)
		case key.Matches(msg, m.sectionKeys.Daemons):
			return m.switchToServerTab(6)
		case key.Matches(msg, m.sectionKeys.Firewall):
			return m.switchToServerTab(7)
		case key.Matches(msg, m.sectionKeys.Jobs):
			return m.switchToServerTab(8)
		case key.Matches(msg, m.sectionKeys.Domains):
			return m.switchToServerTab(9)
		}
	}

	// Delegate navigation keys to the tree panel.
	var cmd tea.Cmd
	panel, cmd := m.treePanel.Update(msg)
	m.treePanel = panel.(panels.TreePanel)
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
	// go back to the commands list (not up to tree panel).
	if m.activeTab == 6 && m.selectedSite != nil && m.commandsPanel.ShowingDetail() {
		if key.Matches(msg, m.navKeys.Back) {
			p, cmd := m.commandsPanel.Update(msg)
			m.commandsPanel = p.(panels.CommandsPanel)
			return m, cmd
		}
	}

	switch {
	case key.Matches(msg, m.navKeys.Back):
		// In server-only context with a server tab active, go back to Info first.
		if m.selectedSite == nil && m.selectedSrv != nil && serverTabNums[m.activeTab] {
			m.activeTab = 1 // Reset to default (shows Info in server context).
			return m, nil
		}
		m.focus = FocusTree
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

// handleOutputKey processes keys when the output panel is focused.
func (m App) handleOutputKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.navKeys.Back):
		m.focus = FocusDetail
		m.outputPoll.active = false // Stop polling when leaving output.
		return m, nil
	}

	// Delegate scrolling keys to the output panel.
	p, cmd := m.outputPanel.Update(msg)
	m.outputPanel = p.(panels.OutputPanel)
	return m, cmd
}

// switchToServerTab changes to a server-level tab without changing focus.
func (m App) switchToServerTab(tab int) (tea.Model, tea.Cmd) {
	m.activeTab = tab
	m.showDeployScript = false
	m.showDBUsers = false
	if m.selectedSrv == nil {
		return m, nil
	}
	return m.initTabPanel(tab, m.selectedSrv.ID, 0)
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
		i := components.NewInputWide("create-sshkey-path", "Path to public key (or paste key directly):", "~/.ssh/id_rsa.pub")
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

// handleSSHKeyCreate handles the result of the SSH key creation input.
// If the input looks like a file path, it reads the file; otherwise it
// treats the input as raw key content and prompts for a name.
func (m App) handleSSHKeyCreate(input string) (tea.Model, tea.Cmd) {
	// Try to expand ~ and read as file path.
	path := input
	if strings.HasPrefix(path, "~/") || path == "~" {
		home, err := os.UserHomeDir()
		if err == nil {
			path = filepath.Join(home, path[2:])
		}
	}

	content, err := os.ReadFile(path)
	if err == nil {
		// Successfully read file. Derive key name from filename.
		name := filepath.Base(path)
		name = strings.TrimSuffix(name, ".pub")
		keyContent := strings.TrimSpace(string(content))
		if keyContent == "" {
			m.toast = "File is empty"
			m.toastIsErr = true
			return m, m.clearToastAfter(3 * time.Second)
		}
		return m, m.sshKeysPanel.CreateKey(name, keyContent, "forge")
	}

	// Not a file — treat as raw key content. Prompt for a name.
	m.pendingInputValue = input
	i := components.NewInput("create-sshkey-name", "Key name:", "my-key")
	m.inputDialog = &i
	return m, nil
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
	case "create-sshkey-path":
		return m.handleSSHKeyCreate(value)
	case "create-sshkey-name":
		// Second step: user provided a name for a pasted key.
		keyContent := m.pendingInputValue
		m.pendingInputValue = ""
		return m, m.sshKeysPanel.CreateKey(value, keyContent, "forge")
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

// View renders the layout: tree (left), detail+output (right), footer (bottom).
func (m App) View() tea.View {
	if m.width == 0 || m.height == 0 {
		v := tea.NewView("Loading...")
		v.AltScreen = true
		return v
	}

	// Reserve space for the footer (1 line) and optional toast (1 line).
	footerHeight := 1
	toastHeight := 0
	if m.toast != "" {
		toastHeight = 1
	}
	contentHeight := m.height - footerHeight - toastHeight

	// Left panel = ~30% width, right panel = rest.
	leftWidth := m.width * 3 / 10
	if leftWidth < 20 {
		leftWidth = 20
	}
	rightWidth := m.width - leftWidth

	// Tree panel on the left, full content height.
	treeView := m.treePanel.View(leftWidth, contentHeight, m.focus == FocusTree)

	// Right side: detail panel on top, output panel on bottom.
	// Adaptive: if output has no content, give detail more space.
	var detailHeight, outputHeight int
	if m.outputPanel.HasContent() {
		detailHeight = contentHeight * 60 / 100
		outputHeight = contentHeight - detailHeight
	} else {
		detailHeight = contentHeight * 85 / 100
		outputHeight = contentHeight - detailHeight
	}
	if detailHeight < 4 {
		detailHeight = 4
	}
	if outputHeight < 3 {
		outputHeight = 3
	}
	// Re-balance if sum exceeds content height.
	if detailHeight+outputHeight > contentHeight {
		outputHeight = contentHeight - detailHeight
		if outputHeight < 3 {
			outputHeight = 3
			detailHeight = contentHeight - outputHeight
		}
	}

	detailView := m.renderDetailPanel(rightWidth, detailHeight)
	outputView := m.outputPanel.View(rightWidth, outputHeight, m.focus == FocusOutput)

	// Join the right panels vertically.
	rightSide := lipgloss.JoinVertical(lipgloss.Left, detailView, outputView)

	// Join left and right horizontally.
	mainContent := lipgloss.JoinHorizontal(lipgloss.Top, treeView, rightSide)

	// Hard-clip mainContent to exactly contentHeight lines so the footer
	// is never pushed off-screen. String-based truncation is more reliable
	// than lipgloss Height/MaxHeight which can add padding or interact
	// unexpectedly with border rendering.
	if mainLines := strings.Split(mainContent, "\n"); len(mainLines) > contentHeight {
		mainContent = strings.Join(mainLines[:contentHeight], "\n")
	}

	// Build the footer.
	footer := m.renderFooter()

	// Assemble everything.
	var parts []string
	parts = append(parts, mainContent)
	if m.toast != "" {
		parts = append(parts, m.renderToast())
	}
	parts = append(parts, footer)

	content := lipgloss.JoinVertical(lipgloss.Left, parts...)

	// Overlay the input dialog if active (float on top of existing UI).
	if m.inputDialog != nil && m.inputDialog.Active {
		overlay := m.inputDialog.View(m.width, m.height)
		if overlay != "" {
			content = overlayCenter(overlay, content, m.width, m.height)
		}
	}

	// Overlay the confirmation dialog if active (float on top of existing UI).
	if m.confirm != nil && m.confirm.Active {
		overlay := m.confirm.View(m.width, m.height)
		if overlay != "" {
			content = overlayCenter(overlay, content, m.width, m.height)
		}
	}

	// Overlay the help modal on top of the existing UI.
	if m.helpModal.Active() {
		box := m.helpModal.View(m.width, m.height)
		if box != "" {
			content = overlayCenter(box, content, m.width, m.height)
		}
	}

	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

// renderDetailPanel renders the top-right detail panel.
// When a site is selected it shows a tab bar with site-level panels;
// when only a server is selected it shows a tab bar with server-level panels;
// otherwise it shows server info.
func (m App) renderDetailPanel(width, height int) string {
	focused := m.focus == FocusDetail

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

	// Server-only context: always show server tab bar.
	if m.selectedSrv != nil {
		tabBar := m.renderServerTabBar(width)
		tabBarHeight := lipgloss.Height(tabBar)

		sectionHeight := height - tabBarHeight
		if sectionHeight < 2 {
			sectionHeight = 2
		}

		var sectionPanel string
		switch m.activeTab {
		case 3:
			if m.showDBUsers {
				sectionPanel = m.dbUsersPanel.View(width, sectionHeight, focused)
			} else {
				sectionPanel = m.databasesPanel.View(width, sectionHeight, focused)
			}
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

// serverTabNums lists which activeTab values correspond to server-level panels.
var serverTabNums = map[int]bool{3: true, 6: true, 7: true, 8: true, 9: true}

// renderServerTabBar renders the server-level tab bar.
func (m App) renderServerTabBar(width int) string {
	tabs := []struct {
		num  int
		name string
	}{
		{0, "Info"}, {3, "DB"}, {6, "Daemons"}, {7, "Firewall"}, {8, "Jobs"}, {9, "SSH Keys"},
	}

	// If the active tab isn't a server-level tab, highlight Info.
	activeForBar := m.activeTab
	if !serverTabNums[activeForBar] {
		activeForBar = 0
	}

	var parts []string
	for _, t := range tabs {
		var label string
		if t.num == 0 {
			label = t.name
		} else {
			label = fmt.Sprintf("%d:%s", t.num, t.name)
		}
		if t.num == activeForBar {
			parts = append(parts, SelectedItemStyle.Render(label))
		} else {
			parts = append(parts, HelpBarStyle.Render(label))
		}
	}

	bar := strings.Join(parts, "  ")
	return theme.Truncate(bar, width)
}

// renderFooter renders the context-sensitive footer with pipe-separated keybindings.
func (m App) renderFooter() string {
	var helpBindings []panels.HelpBinding

	switch m.focus {
	case FocusTree:
		helpBindings = m.treePanel.HelpBindings()
	case FocusOutput:
		helpBindings = m.outputPanel.HelpBindings()
	case FocusDetail:
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

	// Append context-sensitive global keybindings.
	if m.selectedSrv != nil {
		helpBindings = append(helpBindings,
			panels.HelpBinding{Key: "ctrl+s", Desc: "SSH"},
			panels.HelpBinding{Key: "ctrl+f", Desc: "SFTP"},
		)
		if m.selectedSite != nil {
			helpBindings = append(helpBindings,
				panels.HelpBinding{Key: "ctrl+d", Desc: "Database"},
			)
		}
	}
	helpBindings = append(helpBindings, panels.HelpBinding{Key: "?", Desc: "help"})

	var formatted []string
	for _, b := range helpBindings {
		formatted = append(formatted, helpBinding(b.Key, b.Desc))
	}

	bar := strings.Join(formatted, HelpBarStyle.Render(" \u2502 "))

	return HelpBarStyle.Width(m.width).Render(bar)
}

// renderToast renders the toast notification bar.
func (m App) renderToast() string {
	style := ToastStyle
	if m.toastIsErr {
		style = ToastErrorStyle
	}
	return style.Width(m.width).Render(m.toast)
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

// fetchSitesForTree returns a command that fetches sites for a server and
// sends a treeSitesLoadedMsg (instead of the old sitesLoadedMsg).
func (m App) fetchSitesForTree(serverID int64) tea.Cmd {
	client := m.forge
	return func() tea.Msg {
		sites, err := client.Sites.List(context.Background(), serverID)
		if err != nil {
			return errMsg{err}
		}
		return treeSitesLoadedMsg{serverID: serverID, sites: sites}
	}
}

// fetchDeployOutput returns a command that fetches deployment output and
// sends a DeployOutputMsg to be routed to the output panel.
func (m App) fetchDeployOutput(serverID, siteID, deployID int64) tea.Cmd {
	client := m.forge
	return func() tea.Msg {
		output, err := client.Deployments.GetOutput(context.Background(), serverID, siteID, deployID)
		if err != nil {
			return panels.PanelErrMsg{Err: err}
		}
		return panels.DeployOutputMsg{Output: output}
	}
}

// fetchDeployOutputWithStatus fetches both output and deployment status,
// returning a pollOutputResultMsg. Used for auto-updating output.
func (m App) fetchDeployOutputWithStatus(serverID, siteID, deployID int64) tea.Cmd {
	client := m.forge
	return func() tea.Msg {
		output, err := client.Deployments.GetOutput(context.Background(), serverID, siteID, deployID)
		if err != nil {
			return panels.PanelErrMsg{Err: err}
		}
		dep, err := client.Deployments.Get(context.Background(), serverID, siteID, deployID)
		if err != nil {
			// If we can't get status, still show the output and stop polling.
			return pollOutputResultMsg{output: output, finished: true}
		}
		finished := dep.Status != "deploying"
		return pollOutputResultMsg{output: output, finished: finished}
	}
}

// pollOutputTick returns a command that sends a pollOutputTickMsg after 2 seconds.
func (m App) pollOutputTick() tea.Cmd {
	return tea.Tick(2*time.Second, func(time.Time) tea.Msg {
		return pollOutputTickMsg{}
	})
}

// toggleDefault saves or clears the default server/site in .phorge.
// If siteName is empty, it toggles only the server default.
// If siteName is non-empty, it sets/clears both server and site.
func (m App) toggleDefault(serverName, siteName string) tea.Cmd {
	currentServer := m.project.Server
	currentSite := m.project.Site
	return func() tea.Msg {
		var newServer, newSite string
		if siteName != "" {
			// Site toggle.
			if strings.EqualFold(currentServer, serverName) && strings.EqualFold(currentSite, siteName) {
				// Already the default — clear both.
				newServer = ""
				newSite = ""
			} else {
				newServer = serverName
				newSite = siteName
			}
		} else {
			// Server toggle.
			if strings.EqualFold(currentServer, serverName) && currentSite == "" {
				// Already the default — clear.
				newServer = ""
			} else {
				newServer = serverName
				// Clear any site default when setting server-only default.
				newSite = ""
			}
		}
		err := config.SaveProjectConfig(config.ProjectConfig{Server: newServer, Site: newSite})
		return setDefaultMsg{serverName: newServer, siteName: newSite, err: err}
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

// helpBinding formats a single key-description pair for the footer.
func helpBinding(k, desc string) string {
	return HelpKeyStyle.Render(k) + " " + HelpBarStyle.Render(desc)
}

// overlayCenter places fg centered on top of bg. Lines outside the overlay
// area keep the background content. Lines within the overlay area preserve
// background content on both the left and right sides of the overlay box,
// giving a true floating-popup effect.
func overlayCenter(fg, bg string, width, height int) string {
	fgLines := strings.Split(fg, "\n")
	bgLines := strings.Split(bg, "\n")

	// Pad bg to full height.
	for len(bgLines) < height {
		bgLines = append(bgLines, "")
	}

	fgH := len(fgLines)
	fgW := lipgloss.Width(fg)

	startY := (height - fgH) / 2
	if startY < 0 {
		startY = 0
	}
	leftPad := (width - fgW) / 2
	if leftPad < 0 {
		leftPad = 0
	}

	result := make([]string, len(bgLines))
	for i, bgLine := range bgLines {
		fgIdx := i - startY
		if fgIdx >= 0 && fgIdx < fgH {
			bgW := lipgloss.Width(bgLine)

			// Left portion: truncate background to leftPad visual width.
			left := ansi.Truncate(bgLine, leftPad, "")
			leftW := lipgloss.Width(left)
			if leftW < leftPad {
				left += strings.Repeat(" ", leftPad-leftW)
			}

			// Right portion: background content after the overlay area.
			rightStart := leftPad + fgW
			right := ""
			if rightStart < bgW {
				right = ansiCutLeft(bgLine, rightStart)
			}

			result[i] = left + fgLines[fgIdx] + right
		} else {
			result[i] = bgLine
		}
	}

	return strings.Join(result, "\n")
}

// ansiCutLeft returns the portion of an ANSI string starting at visual
// position `skip`. It walks through the string, counting visible characters
// and preserving ANSI escape sequences.
func ansiCutLeft(s string, skip int) string {
	var visWidth int
	var result strings.Builder
	inEsc := false
	skipping := true

	for _, r := range s {
		if inEsc {
			if !skipping {
				result.WriteRune(r)
			}
			if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
				inEsc = false
			}
			continue
		}
		if r == '\x1b' {
			inEsc = true
			if !skipping {
				result.WriteRune(r)
			}
			continue
		}

		visWidth++
		if visWidth > skip {
			if skipping {
				skipping = false
			}
			result.WriteRune(r)
		}
	}

	return result.String()
}
