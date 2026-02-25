package panels

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	lipgloss "charm.land/lipgloss/v2"

	"github.com/hinkers/Phorge/internal/forge"
	"github.com/hinkers/Phorge/internal/tui/theme"
)

// TreeNodeSelectedMsg is emitted when the cursor moves to a new node.
type TreeNodeSelectedMsg struct {
	Server forge.Server
	Site   *forge.Site // nil when the cursor is on a server node
}

// TreeFetchSitesMsg is emitted when a server is expanded and its sites
// have not yet been loaded.
type TreeFetchSitesMsg struct {
	ServerID int64
}

// TreeNodeKind distinguishes server nodes from site nodes.
type TreeNodeKind int

const (
	NodeServer TreeNodeKind = iota
	NodeSite
)

// TreeNode is a single entry in the flattened visible list.
type TreeNode struct {
	Kind   TreeNodeKind
	Server forge.Server
	Site   *forge.Site // non-nil only for NodeSite
	IsLast bool        // true when this is the last site under its server
}

// TreePanel is a lazygit-style tree that combines servers and their sites
// into a single navigable panel.
type TreePanel struct {
	servers       []forge.Server
	sitesByServer map[int64][]forge.Site
	expanded      map[int64]bool
	sitesLoaded   map[int64]bool
	sitesLoading  map[int64]bool
	cursor        int
	loading       bool

	// Filter state
	filterInput  textinput.Model
	filterActive bool
	filterText   string

	// Default server/site names (from .phorge project config).
	defaultServer string
	defaultSite   string

	// Keybindings
	up    key.Binding
	down  key.Binding
	enter key.Binding
	home  key.Binding
	end   key.Binding
	left  key.Binding
	right key.Binding
}

// NewTreePanel creates a new, empty tree panel.
func NewTreePanel() TreePanel {
	ti := textinput.New()
	ti.Prompt = "/ "
	ti.Placeholder = "filter..."
	ti.CharLimit = 64

	return TreePanel{
		sitesByServer: make(map[int64][]forge.Site),
		expanded:      make(map[int64]bool),
		sitesLoaded:   make(map[int64]bool),
		sitesLoading:  make(map[int64]bool),
		filterInput:   ti,
		up: key.NewBinding(
			key.WithKeys("k", "up"),
			key.WithHelp("k/up", "up"),
		),
		down: key.NewBinding(
			key.WithKeys("j", "down"),
			key.WithHelp("j/down", "down"),
		),
		enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "expand/select"),
		),
		home: key.NewBinding(
			key.WithKeys("g", "home"),
			key.WithHelp("g", "top"),
		),
		end: key.NewBinding(
			key.WithKeys("G", "end"),
			key.WithHelp("G", "bottom"),
		),
		left: key.NewBinding(
			key.WithKeys("h", "left"),
			key.WithHelp("h", "collapse"),
		),
		right: key.NewBinding(
			key.WithKeys("l", "right"),
			key.WithHelp("l", "expand"),
		),
	}
}

// SetServers replaces the server list and resets state.
func (t TreePanel) SetServers(servers []forge.Server) TreePanel {
	t.servers = servers
	t.loading = false
	t.cursor = 0
	t.filterActive = false
	t.filterText = ""
	t.filterInput.SetValue("")
	// Reset expansion state for a fresh load.
	t.expanded = make(map[int64]bool)
	t.sitesByServer = make(map[int64][]forge.Site)
	t.sitesLoaded = make(map[int64]bool)
	t.sitesLoading = make(map[int64]bool)
	return t
}

// SetSites stores the fetched sites for a server.
func (t TreePanel) SetSites(serverID int64, sites []forge.Site) TreePanel {
	t.sitesByServer[serverID] = sites
	t.sitesLoaded[serverID] = true
	t.sitesLoading[serverID] = false
	return t
}

// IsExpanded reports whether a server node is currently expanded.
func (t TreePanel) IsExpanded(serverID int64) bool {
	return t.expanded[serverID]
}

// SetLoading sets the top-level loading state.
func (t TreePanel) SetLoading(loading bool) TreePanel {
	t.loading = loading
	return t
}

// SetSitesLoading marks a server's sites as currently being fetched.
func (t TreePanel) SetSitesLoading(serverID int64) TreePanel {
	t.sitesLoading[serverID] = true
	return t
}

// ExpandServer expands a server node by ID without toggling.
// If sites haven't been loaded yet, it returns a TreeFetchSitesMsg command.
func (t TreePanel) ExpandServer(serverID int64) (TreePanel, tea.Cmd) {
	t.expanded[serverID] = true
	if !t.sitesLoaded[serverID] && !t.sitesLoading[serverID] {
		t.sitesLoading[serverID] = true
		return t, func() tea.Msg {
			return TreeFetchSitesMsg{ServerID: serverID}
		}
	}
	return t, nil
}

// SetCursorToServer moves the cursor to the server node with the given ID.
// Returns true if the server was found.
func (t TreePanel) SetCursorToServer(serverID int64) (TreePanel, bool) {
	nodes := t.visibleNodes()
	for i, node := range nodes {
		if node.Kind == NodeServer && node.Server.ID == serverID {
			t.cursor = i
			return t, true
		}
	}
	return t, false
}

// FindServerByName returns the server with the given name, or nil if not found.
func (t TreePanel) FindServerByName(name string) *forge.Server {
	nameLower := strings.ToLower(name)
	for _, srv := range t.servers {
		if strings.ToLower(srv.Name) == nameLower {
			s := srv
			return &s
		}
	}
	return nil
}

// SetDefaultServer sets the name of the default server for visual indicator.
func (t TreePanel) SetDefaultServer(name string) TreePanel {
	t.defaultServer = name
	return t
}

// SetDefaultSite sets the name of the default site for visual indicator.
func (t TreePanel) SetDefaultSite(name string) TreePanel {
	t.defaultSite = name
	return t
}

// FindSiteByName returns the server and site with the given site name, or nils.
func (t TreePanel) FindSiteByName(siteName string) (*forge.Server, *forge.Site) {
	nameLower := strings.ToLower(siteName)
	for _, srv := range t.servers {
		for _, site := range t.sitesByServer[srv.ID] {
			if strings.ToLower(site.Name) == nameLower {
				s := site
				sv := srv
				return &sv, &s
			}
		}
	}
	return nil, nil
}

// SetCursorToSite moves the cursor to the site node with the given ID.
// Returns true if the site was found.
func (t TreePanel) SetCursorToSite(siteID int64) (TreePanel, bool) {
	nodes := t.visibleNodes()
	for i, node := range nodes {
		if node.Kind == NodeSite && node.Site != nil && node.Site.ID == siteID {
			t.cursor = i
			return t, true
		}
	}
	return t, false
}

// CursorOnServer reports whether the cursor is currently on a server node.
func (t TreePanel) CursorOnServer() bool {
	nodes := t.visibleNodes()
	if len(nodes) == 0 || t.cursor >= len(nodes) {
		return false
	}
	return nodes[t.cursor].Kind == NodeServer
}

// FilterActive reports whether the filter input is currently active.
func (t TreePanel) FilterActive() bool {
	return t.filterActive
}

// Selected returns the server and optional site at the current cursor position.
func (t TreePanel) Selected() (forge.Server, *forge.Site) {
	nodes := t.visibleNodes()
	if len(nodes) == 0 || t.cursor >= len(nodes) {
		return forge.Server{}, nil
	}
	node := nodes[t.cursor]
	return node.Server, node.Site
}

// visibleNodes builds the flat list of visible tree nodes.
func (t TreePanel) visibleNodes() []TreeNode {
	filterLower := strings.ToLower(t.filterText)
	var nodes []TreeNode

	for _, srv := range t.servers {
		srvMatches := filterLower == "" || strings.Contains(strings.ToLower(srv.Name), filterLower)

		sites := t.sitesByServer[srv.ID]

		// Collect sites that match the filter (used when expanded or auto-expanding).
		var matchingSites []forge.Site
		for _, site := range sites {
			if filterLower == "" || strings.Contains(strings.ToLower(site.Name), filterLower) || srvMatches {
				matchingSites = append(matchingSites, site)
			}
		}

		// Determine whether to show children: either manually expanded, or
		// auto-expand when filtering reveals matching child sites.
		showChildren := t.expanded[srv.ID]
		if filterLower != "" && !srvMatches && len(matchingSites) > 0 {
			showChildren = true
		}

		// A server is visible if it matches the filter, or has matching sites.
		if !srvMatches && len(matchingSites) == 0 {
			continue
		}

		nodes = append(nodes, TreeNode{
			Kind:   NodeServer,
			Server: srv,
		})

		if showChildren {
			for i, site := range matchingSites {
				s := site
				nodes = append(nodes, TreeNode{
					Kind:   NodeSite,
					Server: srv,
					Site:   &s,
					IsLast: i == len(matchingSites)-1,
				})
			}
		}
	}

	return nodes
}

// Update handles key events for the tree panel.
func (t TreePanel) Update(msg tea.Msg) (Panel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if t.filterActive {
			return t.handleFilterKey(msg)
		}
		return t.handleKey(msg)
	}
	return t, nil
}

// handleFilterKey processes keys when the filter input is active.
func (t TreePanel) handleFilterKey(msg tea.KeyPressMsg) (Panel, tea.Cmd) {
	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
		t.filterActive = false
		t.filterText = t.filterInput.Value()
		t.cursor = 0
		return t, t.emitSelected()

	case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
		t.filterActive = false
		t.filterText = ""
		t.filterInput.SetValue("")
		t.cursor = 0
		return t, t.emitSelected()
	}

	var cmd tea.Cmd
	t.filterInput, cmd = t.filterInput.Update(msg)
	t.filterText = t.filterInput.Value()
	t.cursor = 0
	return t, cmd
}

// handleKey processes normal (non-filter) key events.
func (t TreePanel) handleKey(msg tea.KeyPressMsg) (Panel, tea.Cmd) {
	nodes := t.visibleNodes()

	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("/"))):
		t.filterActive = true
		t.filterInput.SetValue(t.filterText)
		t.filterInput.Focus()
		return t, textinput.Blink

	case key.Matches(msg, t.down):
		if len(nodes) > 0 {
			t.cursor = min(t.cursor+1, len(nodes)-1)
			return t, t.emitSelected()
		}

	case key.Matches(msg, t.up):
		if len(nodes) > 0 {
			t.cursor = max(t.cursor-1, 0)
			return t, t.emitSelected()
		}

	case key.Matches(msg, t.home):
		if len(nodes) > 0 {
			t.cursor = 0
			return t, t.emitSelected()
		}

	case key.Matches(msg, t.end):
		if len(nodes) > 0 {
			t.cursor = len(nodes) - 1
			return t, t.emitSelected()
		}

	case key.Matches(msg, t.enter):
		// Enter: emit selected — app.go handles focus change for both
		// server and site nodes.
		if t.cursor < len(nodes) {
			node := nodes[t.cursor]
			if node.Kind == NodeServer && !t.expanded[node.Server.ID] {
				// First expand the server so sites load.
				return t.toggleServer(node.Server)
			}
			return t, t.emitSelected()
		}

	case key.Matches(msg, key.NewBinding(key.WithKeys(" "))):
		// Space: toggle expand/collapse for servers.
		if t.cursor < len(nodes) {
			node := nodes[t.cursor]
			if node.Kind == NodeServer {
				return t.toggleServer(node.Server)
			}
		}

	case key.Matches(msg, t.right):
		// l / right: expand server or move to first child site.
		if t.cursor < len(nodes) {
			node := nodes[t.cursor]
			if node.Kind == NodeServer {
				if !t.expanded[node.Server.ID] {
					// Expand the server.
					return t.toggleServer(node.Server)
				}
				// Already expanded — move cursor to first child site.
				if t.cursor+1 < len(nodes) && nodes[t.cursor+1].Kind == NodeSite {
					t.cursor++
					return t, t.emitSelected()
				}
			}
			// On a site node, l is handled by app.go (focus detail panel).
		}

	case key.Matches(msg, t.left):
		// h / left: collapse server or move to parent server from site.
		if t.cursor < len(nodes) {
			node := nodes[t.cursor]
			if node.Kind == NodeServer {
				// Collapse the server if expanded.
				if t.expanded[node.Server.ID] {
					t.expanded[node.Server.ID] = false
					newNodes := t.visibleNodes()
					if t.cursor >= len(newNodes) && len(newNodes) > 0 {
						t.cursor = len(newNodes) - 1
					}
					return t, t.emitSelected()
				}
			} else if node.Kind == NodeSite {
				// Move cursor to parent server and collapse it.
				parentID := node.Server.ID
				t.expanded[parentID] = false
				newNodes := t.visibleNodes()
				// Find the parent server node.
				for i, n := range newNodes {
					if n.Kind == NodeServer && n.Server.ID == parentID {
						t.cursor = i
						break
					}
				}
				return t, t.emitSelected()
			}
		}
	}

	return t, nil
}

// toggleServer expands or collapses a server node.
func (t TreePanel) toggleServer(srv forge.Server) (Panel, tea.Cmd) {
	if t.expanded[srv.ID] {
		// Collapse.
		t.expanded[srv.ID] = false
		// Clamp cursor in case it was on a child site that's now hidden.
		nodes := t.visibleNodes()
		if t.cursor >= len(nodes) && len(nodes) > 0 {
			t.cursor = len(nodes) - 1
		}
		return t, t.emitSelected()
	}

	// Expand.
	t.expanded[srv.ID] = true
	if !t.sitesLoaded[srv.ID] && !t.sitesLoading[srv.ID] {
		t.sitesLoading[srv.ID] = true
		serverID := srv.ID
		return t, func() tea.Msg {
			return TreeFetchSitesMsg{ServerID: serverID}
		}
	}
	return t, t.emitSelected()
}

// emitSelected returns a command that emits a TreeNodeSelectedMsg for the
// currently highlighted node.
func (t TreePanel) emitSelected() tea.Cmd {
	nodes := t.visibleNodes()
	if len(nodes) == 0 || t.cursor >= len(nodes) {
		return nil
	}
	node := nodes[t.cursor]
	srv := node.Server
	var site *forge.Site
	if node.Site != nil {
		s := *node.Site
		site = &s
	}
	return func() tea.Msg {
		return TreeNodeSelectedMsg{Server: srv, Site: site}
	}
}

// View renders the tree panel.
func (t TreePanel) View(width, height int, focused bool) string {
	style := theme.InactiveBorderStyle
	titleColor := theme.ColorSubtle
	if focused {
		style = theme.ActiveBorderStyle
		titleColor = theme.ColorPrimary
	}

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(titleColor).
		Render(" Servers ")

	innerWidth := width - 2
	innerHeight := height - 3
	if innerWidth < 0 {
		innerWidth = 0
	}
	if innerHeight < 0 {
		innerHeight = 0
	}

	var lines []string

	// Render filter UI.
	if t.filterActive {
		filterLine := t.filterInput.View()
		lines = append(lines, theme.Truncate(filterLine, innerWidth))
		innerHeight--
	} else if t.filterText != "" {
		indicator := theme.FilterIndicatorStyle.
			Render("filter: " + t.filterText)
		lines = append(lines, theme.Truncate(indicator, innerWidth))
		innerHeight--
	}

	nodes := t.visibleNodes()

	if t.loading && len(t.servers) == 0 {
		lines = append(lines, theme.LoadingStyle.Render("Loading servers..."))
	} else if len(t.servers) == 0 {
		lines = append(lines, theme.NormalItemStyle.Render("No servers found"))
	} else if len(nodes) == 0 && t.filterText != "" {
		lines = append(lines, theme.NormalItemStyle.Render("No matches"))
	} else {
		filterLines := 0
		if t.filterActive || t.filterText != "" {
			filterLines = 1
		}

		visibleHeight := innerHeight
		if visibleHeight < 1 {
			visibleHeight = 1
		}
		startIdx := 0
		if t.cursor >= visibleHeight {
			startIdx = t.cursor - visibleHeight + 1
		}

		for i := startIdx; i < len(nodes) && len(lines)-filterLines < visibleHeight; i++ {
			node := nodes[i]
			line := t.renderNode(node, i, innerWidth)
			lines = append(lines, line)
		}
	}

	// Pad to fill.
	totalHeight := height - 2
	if totalHeight < 0 {
		totalHeight = 0
	}
	for len(lines) < totalHeight-1 {
		lines = append(lines, "")
	}

	content := strings.Join(lines, "\n")

	return style.
		Width(width - 2).
		Height(totalHeight).
		Render(title + "\n" + content)
}

// renderNode renders a single tree node line.
func (t TreePanel) renderNode(node TreeNode, idx, maxWidth int) string {
	isCursor := idx == t.cursor

	if node.Kind == NodeServer {
		icon := "▶"
		if t.expanded[node.Server.ID] {
			icon = "▼"
		}

		// Show * next to the default server.
		suffix := ""
		if t.defaultServer != "" && strings.EqualFold(node.Server.Name, t.defaultServer) {
			suffix = " *"
		}

		name := theme.Truncate(node.Server.Name+suffix, maxWidth-6)
		if isCursor {
			return theme.CursorStyle.Render("> ") +
				theme.SelectedItemStyle.Render(icon+" "+name)
		}
		return "  " + theme.NormalItemStyle.Render(icon+" "+name)
	}

	// Site node.
	prefix := "├ "
	if node.IsLast {
		prefix = "└ "
	}

	siteName := ""
	if node.Site != nil {
		siteName = node.Site.Name
	}

	// Show * next to the default site.
	siteSuffix := ""
	if t.defaultSite != "" && strings.EqualFold(siteName, t.defaultSite) {
		siteSuffix = " *"
	}

	name := theme.Truncate(siteName+siteSuffix, maxWidth-8)

	if isCursor {
		return theme.CursorStyle.Render("> ") +
			"  " + theme.SelectedItemStyle.Render(prefix+name)
	}
	return "    " + theme.NormalItemStyle.Render(prefix+name)
}

// HelpBindings returns the key hints for the tree panel.
// The bindings are context-aware based on cursor position.
func (t TreePanel) HelpBindings() []HelpBinding {
	if t.filterActive {
		return []HelpBinding{
			{Key: "enter", Desc: "accept filter"},
			{Key: "esc", Desc: "clear filter"},
		}
	}

	bindings := []HelpBinding{
		{Key: "j/k", Desc: "navigate"},
		{Key: "h/l", Desc: "collapse/expand"},
	}

	if t.CursorOnServer() {
		bindings = append(bindings,
			HelpBinding{Key: "enter", Desc: "select → detail"},
			HelpBinding{Key: "space", Desc: "expand/collapse"},
			HelpBinding{Key: "s", Desc: "SSH"},
			HelpBinding{Key: "r", Desc: "reboot"},
			HelpBinding{Key: "D", Desc: "set default"},
		)
	} else {
		bindings = append(bindings,
			HelpBinding{Key: "enter", Desc: "select → detail"},
			HelpBinding{Key: "d", Desc: "deploy"},
			HelpBinding{Key: "s", Desc: "SSH"},
			HelpBinding{Key: "D", Desc: "set default"},
		)
	}

	bindings = append(bindings,
		HelpBinding{Key: "/", Desc: "filter"},
		HelpBinding{Key: "tab", Desc: "next panel"},
	)

	return bindings
}
