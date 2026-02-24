# Phorge Go + Bubbletea Rewrite Design

## Problem

Phorge is currently built with Python/Textual, which provides a widget/button/CSS-driven interaction model. This feels like a web app in the terminal rather than a keyboard-first power-user tool. The goal is to rewrite Phorge using Go + Bubbletea to achieve a lazygit-style UX: stacked panels, single-key actions, j/k navigation, context-sensitive keybinding hints.

## Decision

Full rewrite in Go using the Bubbletea framework (Elm Architecture / Model-View-Update). Target the new Forge API (org-scoped, JSON:API format) since v1 sunsets March 31, 2026.

## Architecture

### Framework Stack

- **Bubbletea** - TUI framework (Elm Architecture: Init/Update/View)
- **Bubbles** - Pre-built components (list, table, viewport, help, text input)
- **Lip Gloss** - Terminal styling (colors, borders, layout via JoinHorizontal/JoinVertical)
- **Standard library `net/http`** - Forge API client
- **pelletier/go-toml** - Configuration

### Project Structure

```
phorge/
  cmd/phorge/
    main.go                   # Entry point, config load, launch app
  internal/
    forge/                    # Forge API client (go-github pattern)
      client.go               # Client struct, NewClient(), Do(), NewRequest()
      servers.go              # ServersService
      sites.go                # SitesService
      deployments.go          # DeploymentsService
      databases.go            # DatabasesService
      environment.go          # EnvironmentService
      certificates.go         # CertificatesService
      workers.go              # WorkersService
      daemons.go              # DaemonsService
      firewall.go             # FirewallService
      jobs.go                 # JobsService
      backups.go              # BackupsService
      ssh_keys.go             # SSHKeysService
      commands.go             # SiteCommandsService
      git.go                  # GitService
      logs.go                 # LogsService
      types.go                # All data model structs
      errors.go               # ForgeError, RateLimitError, etc.
    tui/
      app.go                  # Root Bubbletea model, top-level Update/View
      keymap.go               # Keybinding definitions per context
      styles.go               # Lip Gloss styles (colors, borders, layout)
      panels/                 # Individual panel models
        servers.go            # Server list panel
        sites.go              # Site list panel
        server_info.go        # Server detail
        site_info.go          # Site detail
        deployments.go        # Deployment list + actions
        deploy_script.go      # Deployment script editor
        environment.go        # Env file viewer/editor
        databases.go          # Database list + CRUD
        database_users.go     # DB user management
        ssl.go                # SSL certificate management
        workers.go            # Queue worker management
        daemons.go            # Daemon management
        firewall.go           # Firewall rule management
        jobs.go               # Scheduled jobs
        backups.go            # Backup management
        ssh_keys.go           # SSH key management
        commands.go           # Site commands
        logs.go               # Log viewer
        git.go                # Git info panel
        domains.go            # Domain/alias management
      components/             # Reusable Bubbletea sub-models
        list.go               # Scrollable list with j/k navigation
        table.go              # Data table component
        help.go               # Context-sensitive help bar
        confirm.go            # Confirmation dialog
        input.go              # Text input modal
        viewport.go           # Scrollable text viewer
        toast.go              # Notification toast
    config/
      config.go               # TOML config load/save
  go.mod
  go.sum
```

### API Client Design (go-github pattern)

Service-oriented client with injectable HTTP transport:

```go
type Client struct {
    client  *http.Client
    BaseURL *url.URL
    Token   string

    Servers      *ServersService
    Sites        *SitesService
    Deployments  *DeploymentsService
    Databases    *DatabasesService
    // ... etc
}

type ServersService struct {
    client *Client
}

func (s *ServersService) List(ctx context.Context) ([]*Server, error) { ... }
func (s *ServersService) Get(ctx context.Context, id int64) (*Server, error) { ... }
func (s *ServersService) Reboot(ctx context.Context, id int64) error { ... }
```

Authentication via Bearer token in a custom RoundTripper transport.

## UX Design

### Layout

Three-panel lazygit-style layout with a bottom help bar:

```
+--------------------+----------------------------------------+
| Servers            | Sites (or sub-resource list)           |
|                    |                                        |
| > production-1     |  example.com                           |
|   staging-1        |  api.example.com                       |
|   dev-1            |> staging.example.com                   |
|                    |                                        |
|                    +----------------------------------------+
|                    | Detail / Preview                       |
|                    |                                        |
|                    |  PHP: 8.3  | Status: active            |
|                    |  Branch: main                          |
|                    |  Last deploy: 2m ago (success)         |
|                    |  Repository: github.com/user/repo      |
|                    |                                        |
+--------------------+----------------------------------------+
| d deploy  e env  s ssh  f sftp  D database  l logs  ? help |
+------------------------------------------------------------+
```

- **Left panel**: Server list (always visible)
- **Top-right panel**: Context list (sites, databases, keys, etc.)
- **Bottom-right panel**: Detail/preview of selected item
- **Bottom bar**: Context-sensitive keybinding hints

### Navigation

| Key | Action |
|-----|--------|
| `Tab` / `Shift+Tab` | Cycle focus between panels |
| `j` / `k` | Move up/down in focused list |
| `Enter` | Drill into item |
| `Esc` / `Backspace` | Go back up |
| `1`-`9` | Jump to section tab (Deployments, Env, Databases, etc.) |
| `/` | Search/filter current list |
| `?` | Full help modal |
| `q` | Quit |

### Section Tabs (when a site is selected)

| Key | Section |
|-----|---------|
| `1` | Deployments |
| `2` | Environment |
| `3` | Databases |
| `4` | SSL Certificates |
| `5` | Workers |
| `6` | Commands |
| `7` | Logs |
| `8` | Git |
| `9` | Domains |

### Context-Sensitive Actions

Actions shown in the bottom help bar change based on what's focused:

**Server selected**: `s` ssh, `f` sftp, `r` reboot, `Enter` view sites
**Site selected**: `d` deploy, `e` env, `s` ssh, `D` database, `l` logs
**Deployment list**: `d` deploy now, `r` reset status, `Enter` view output
**Database list**: `c` create, `x` delete, `u` users

### Global Keys

| Key | Action |
|-----|--------|
| `Ctrl+S` | SSH to selected server/site |
| `Ctrl+F` | SFTP via termscp |
| `Ctrl+D` | Database via SSH tunnel + lazysql |
| `Ctrl+R` | Refresh current view |

## State Management

### Root Model

Holds global state and routes messages to active panel:

```go
type model struct {
    forge       *forge.Client
    config      *config.Config

    // Panel focus
    focus       Focus  // ServerList | ContextList | DetailPanel

    // Context
    servers     []forge.Server
    selectedSrv *forge.Server
    selectedSite *forge.Site
    activeTab   int  // 1-9 section tabs

    // Sub-models
    serverList  panels.ServerList
    contextList panels.Panel  // polymorphic - changes with tab
    detailPanel panels.Panel
    helpBar     components.Help
    toast       components.Toast

    // State
    width, height int
}
```

### Message Flow

```
KeyMsg → root.Update()
  → Route to focused panel's Update()
  → Panel returns tea.Cmd (API call, etc.)
  → Cmd runs async, returns result Msg
  → Result Msg arrives at root.Update()
  → Root forwards to panel, panel updates state
  → View() re-renders
```

### API Calls as Commands

```go
func fetchServers(client *forge.Client) tea.Cmd {
    return func() tea.Msg {
        servers, err := client.Servers.List(context.Background())
        if err != nil {
            return errMsg{err}
        }
        return serversLoadedMsg{servers}
    }
}
```

### External Tools

Bubbletea's `tea.ExecProcess()` suspends the TUI and runs a subprocess:

```go
case "s":  // SSH
    cmd := exec.Command("ssh", fmt.Sprintf("%s@%s", sshUser, server.IPAddress))
    return m, tea.ExecProcess(cmd, func(err error) tea.Msg {
        return sshExitMsg{err}
    })
```

## Configuration

Same TOML format and location as current Python version:

```toml
# ~/.config/phorge/config.toml

[forge]
api_key = "your-key"
ssh_user = "forge"

[editor]
command = "vim"

[ui]
vim_keys = true

[server_users]
"production-1" = "deployer"
```

The Go version reads the existing config file - users just replace the binary.

## Distribution

- Single binary via `go build` (zero runtime dependencies)
- Cross-compiled for linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64
- GitHub Releases with prebuilt binaries
- `go install github.com/user/phorge@latest`
- Homebrew tap for macOS
- Scoop manifest for Windows

## Error Handling

- **API errors**: Toast notification with error message
- **Rate limiting**: Back off, show "Rate limited, retrying in Xs" in status bar
- **Network errors**: Inline error with `r` to retry hint
- **Auth errors**: Prompt to re-enter API key (setup flow)
- **Validation errors**: Show field-level errors in input modals

## Feature Parity

All 18 current panel types carry over:

| Current Panel | Go Equivalent |
|---------------|---------------|
| ServerInfoPanel | panels/server_info.go |
| SiteInfoPanel | panels/site_info.go |
| DeploymentsPanel | panels/deployments.go |
| DeploymentScriptPanel | panels/deploy_script.go |
| EnvironmentPanel | panels/environment.go |
| DatabasesPanel | panels/databases.go |
| DatabaseUsersPanel | panels/database_users.go |
| SSLPanel | panels/ssl.go |
| WorkersPanel | panels/workers.go |
| DaemonsPanel | panels/daemons.go |
| FirewallPanel | panels/firewall.go |
| ScheduledJobsPanel | panels/jobs.go |
| BackupsPanel | panels/backups.go |
| SSHKeysPanel | panels/ssh_keys.go |
| CommandsPanel | panels/commands.go |
| LogsPanel | panels/logs.go |
| GitPanel | panels/git.go |
| DomainsPanel | panels/domains.go |

All integrations carry over:
- SSH via `ssh` subprocess
- SFTP via `termscp` subprocess
- Database via SSH tunnel + `lazysql` subprocess
- Environment editing via external editor subprocess

## Testing Strategy

- **API client**: Unit tests with `net/http/httptest.Server` mocking Forge responses
- **TUI models**: Bubbletea's `teatest` package for state transition testing
- **Integration**: Manual testing against real Forge account
