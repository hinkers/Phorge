# Phorge Go + Bubbletea Rewrite Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Rewrite Phorge from Python/Textual to Go/Bubbletea for a lazygit-style keyboard-first TUI experience.

**Architecture:** Go module with Bubbletea (Elm Architecture), Bubbles components, Lip Gloss styling. Forge API client follows the go-github service-oriented pattern. Three-panel lazygit-style layout with context-sensitive keybinding hints.

**Tech Stack:** Go 1.23+, Bubbletea v2, Bubbles, Lip Gloss v2, pelletier/go-toml/v2, net/http

**Reference:** See `docs/plans/2026-02-24-go-bubbletea-rewrite-design.md` for the full design document.

---

## Task 1: Project Scaffold & Go Module

**Files:**
- Create: `cmd/phorge/main.go`
- Create: `go.mod`
- Create: `.gitignore` (update for Go)

**Step 1: Initialize Go module**

```bash
cd /path/to/new/phorge-go  # or a new directory
go mod init github.com/hinke/phorge
```

**Step 2: Create entry point**

```go
// cmd/phorge/main.go
package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea/v2"
)

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// Placeholder model - will be replaced by tui.App in Task 7
type model struct{}

func initialModel() model { return model{} }
func (m model) Init() (model, tea.Cmd)                           { return m, nil }
func (m model) Update(msg tea.Msg) (model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" {
			return m, tea.Quit
		}
	}
	return m, nil
}
func (m model) View() string { return "Phorge - press q to quit" }
```

**Step 3: Add Go dependencies**

```bash
go get github.com/charmbracelet/bubbletea/v2
go get github.com/charmbracelet/bubbles/v2
go get github.com/charmbracelet/lipgloss/v2
go get github.com/pelletier/go-toml/v2
```

**Step 4: Update .gitignore for Go**

Append Go-specific entries to existing `.gitignore`:
```
# Go
/phorge
/phorge.exe
*.test
*.out
/dist/
```

**Step 5: Verify it builds and runs**

```bash
go build -o phorge ./cmd/phorge
./phorge
# Should show "Phorge - press q to quit" in alt screen
# Press q to exit
```

**Step 6: Commit**

```bash
git add .
git commit -m "feat: initialize Go module with bubbletea scaffold"
```

---

## Task 2: Configuration System

**Files:**
- Create: `internal/config/config.go`
- Create: `internal/config/config_test.go`

**Step 1: Write tests for config loading**

```go
// internal/config/config_test.go
package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	cfg := Default()
	if cfg.Forge.SSHUser != "forge" {
		t.Errorf("expected default ssh_user 'forge', got %q", cfg.Forge.SSHUser)
	}
	if cfg.Editor.Command != "vim" {
		t.Errorf("expected default editor 'vim', got %q", cfg.Editor.Command)
	}
}

func TestLoadFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	content := `
[forge]
api_key = "test-key-123"
ssh_user = "deployer"

[editor]
command = "nano"

[server_users]
"prod-1" = "ubuntu"
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadFrom(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Forge.APIKey != "test-key-123" {
		t.Errorf("expected api_key 'test-key-123', got %q", cfg.Forge.APIKey)
	}
	if cfg.Forge.SSHUser != "deployer" {
		t.Errorf("expected ssh_user 'deployer', got %q", cfg.Forge.SSHUser)
	}
	if cfg.Editor.Command != "nano" {
		t.Errorf("expected editor 'nano', got %q", cfg.Editor.Command)
	}
	if u, ok := cfg.ServerUsers["prod-1"]; !ok || u != "ubuntu" {
		t.Errorf("expected server_users prod-1='ubuntu', got %q", u)
	}
}

func TestSaveAndReload(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	cfg := Default()
	cfg.Forge.APIKey = "saved-key"
	if err := cfg.SaveTo(path); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadFrom(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Forge.APIKey != "saved-key" {
		t.Errorf("expected api_key 'saved-key' after reload, got %q", loaded.Forge.APIKey)
	}
}

func TestConfigPath(t *testing.T) {
	p := DefaultPath()
	if p == "" {
		t.Error("config path should not be empty")
	}
}
```

**Step 2: Run tests to verify they fail**

```bash
go test ./internal/config/ -v
# Expected: FAIL - package doesn't exist yet
```

**Step 3: Implement config package**

```go
// internal/config/config.go
package config

import (
	"os"
	"path/filepath"

	toml "github.com/pelletier/go-toml/v2"
)

type Config struct {
	Forge       ForgeConfig       `toml:"forge"`
	Editor      EditorConfig      `toml:"editor"`
	ServerUsers map[string]string `toml:"server_users"`
}

type ForgeConfig struct {
	APIKey  string `toml:"api_key"`
	SSHUser string `toml:"ssh_user"`
}

type EditorConfig struct {
	Command string `toml:"command"`
}

func Default() *Config {
	return &Config{
		Forge: ForgeConfig{
			SSHUser: "forge",
		},
		Editor: EditorConfig{
			Command: "vim",
		},
		ServerUsers: make(map[string]string),
	}
}

func DefaultPath() string {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		cfgDir = filepath.Join(os.Getenv("HOME"), ".config")
	}
	return filepath.Join(cfgDir, "phorge", "config.toml")
}

func Load() (*Config, error) {
	return LoadFrom(DefaultPath())
}

func LoadFrom(path string) (*Config, error) {
	cfg := Default()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}
	if err := toml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	if cfg.ServerUsers == nil {
		cfg.ServerUsers = make(map[string]string)
	}
	return cfg, nil
}

func (c *Config) Save() error {
	return c.SaveTo(DefaultPath())
}

func (c *Config) SaveTo(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	data, err := toml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

// SSHUserFor returns the SSH user for a server, falling back to the default.
func (c *Config) SSHUserFor(serverName string) string {
	if u, ok := c.ServerUsers[serverName]; ok {
		return u
	}
	return c.Forge.SSHUser
}
```

**Step 4: Run tests to verify they pass**

```bash
go test ./internal/config/ -v
# Expected: PASS
```

**Step 5: Commit**

```bash
git add internal/config/
git commit -m "feat: add TOML configuration system"
```

---

## Task 3: Forge API Client - Core Infrastructure

**Files:**
- Create: `internal/forge/client.go`
- Create: `internal/forge/errors.go`
- Create: `internal/forge/types.go`
- Create: `internal/forge/client_test.go`

**Step 1: Write error types**

```go
// internal/forge/errors.go
package forge

import "fmt"

type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("forge API error %d: %s", e.StatusCode, e.Message)
}

type AuthenticationError struct{ APIError }
type NotFoundError struct{ APIError }
type RateLimitError struct{ APIError }

type ValidationError struct {
	APIError
	Details map[string][]string
}
```

**Step 2: Write data model structs**

```go
// internal/forge/types.go
package forge

type User struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email,omitempty"`
}

type Server struct {
	ID               int64    `json:"id"`
	Name             string   `json:"name"`
	IPAddress        string   `json:"ip_address,omitempty"`
	PrivateIPAddress string   `json:"private_ip_address,omitempty"`
	Region           string   `json:"region,omitempty"`
	PHPVersion       string   `json:"php_version,omitempty"`
	PHPCLIVersion    string   `json:"php_cli_version,omitempty"`
	Provider         string   `json:"provider,omitempty"`
	Type             string   `json:"type,omitempty"`
	Status           string   `json:"status,omitempty"`
	IsReady          bool     `json:"is_ready"`
	DatabaseType     string   `json:"database_type,omitempty"`
	SSHPort          int      `json:"ssh_port,omitempty"`
	UbuntuVersion    string   `json:"ubuntu_version,omitempty"`
	DBStatus         string   `json:"db_status,omitempty"`
	RedisStatus      string   `json:"redis_status,omitempty"`
	Network          []any    `json:"network,omitempty"`
	Tags             []any    `json:"tags,omitempty"`
}

type Site struct {
	ID                 int64    `json:"id"`
	ServerID           int64    `json:"server_id,omitempty"`
	Name               string   `json:"name"`
	Directory          string   `json:"directory,omitempty"`
	WebDirectory       string   `json:"web_directory,omitempty"`
	Repository         string   `json:"repository,omitempty"`
	RepositoryProvider string   `json:"repository_provider,omitempty"`
	RepositoryBranch   string   `json:"repository_branch,omitempty"`
	RepositoryStatus   string   `json:"repository_status,omitempty"`
	QuickDeploy        bool     `json:"quick_deploy"`
	DeploymentURL      string   `json:"deployment_url,omitempty"`
	Status             string   `json:"status,omitempty"`
	ProjectType        string   `json:"project_type,omitempty"`
	PHPVersion         string   `json:"php_version,omitempty"`
	App                string   `json:"app,omitempty"`
	Wildcards          bool     `json:"wildcards"`
	Aliases            []string `json:"aliases,omitempty"`
	IsSecured          bool     `json:"is_secured"`
	Tags               []any    `json:"tags,omitempty"`
}

type Deployment struct {
	ID              int64  `json:"id"`
	ServerID        int64  `json:"server_id,omitempty"`
	SiteID          int64  `json:"site_id"`
	Type            int    `json:"type,omitempty"`
	CommitHash      string `json:"commit_hash,omitempty"`
	CommitAuthor    string `json:"commit_author,omitempty"`
	CommitMessage   string `json:"commit_message,omitempty"`
	StartedAt       string `json:"started_at,omitempty"`
	EndedAt         string `json:"ended_at,omitempty"`
	Status          string `json:"status,omitempty"`
	DisplayableType string `json:"displayable_type,omitempty"`
}

type Database struct {
	ID       int64  `json:"id"`
	ServerID int64  `json:"server_id,omitempty"`
	Name     string `json:"name"`
	Status   string `json:"status,omitempty"`
	IsSynced bool   `json:"is_synced"`
}

type DatabaseUser struct {
	ID        int64   `json:"id"`
	ServerID  int64   `json:"server_id,omitempty"`
	Name      string  `json:"name"`
	Status    string  `json:"status,omitempty"`
	Databases []int64 `json:"databases,omitempty"`
}

type SSHKey struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status,omitempty"`
}

type Daemon struct {
	ID        int64  `json:"id"`
	ServerID  int64  `json:"server_id,omitempty"`
	Command   string `json:"command"`
	User      string `json:"user,omitempty"`
	Directory string `json:"directory,omitempty"`
	Processes int    `json:"processes"`
	StartSecs int    `json:"startsecs"`
	Status    string `json:"status,omitempty"`
}

type FirewallRule struct {
	ID        int64  `json:"id"`
	ServerID  int64  `json:"server_id,omitempty"`
	Name      string `json:"name"`
	Port      any    `json:"port,omitempty"` // can be int or string
	IPAddress string `json:"ip_address,omitempty"`
	Type      string `json:"type,omitempty"`
	Status    string `json:"status,omitempty"`
}

type ScheduledJob struct {
	ID        int64  `json:"id"`
	ServerID  int64  `json:"server_id,omitempty"`
	Command   string `json:"command"`
	User      string `json:"user,omitempty"`
	Frequency string `json:"frequency,omitempty"`
	Cron      string `json:"cron,omitempty"`
	Status    string `json:"status,omitempty"`
}

type Worker struct {
	ID         int64  `json:"id"`
	Connection string `json:"connection,omitempty"`
	Queue      string `json:"queue,omitempty"`
	Timeout    int    `json:"timeout"`
	Sleep      int    `json:"sleep"`
	Processes  int    `json:"processes"`
	DaemonMode bool   `json:"daemon"`
	Force      bool   `json:"force"`
	Status     string `json:"status,omitempty"`
}

type Certificate struct {
	ID       int64  `json:"id"`
	Domain   string `json:"domain,omitempty"`
	Type     string `json:"type,omitempty"`
	Active   bool   `json:"active"`
	Status   string `json:"status,omitempty"`
	Existing bool   `json:"existing"`
}

type Backup struct {
	ID                    int64  `json:"id"`
	BackupConfigurationID int64  `json:"backup_configuration_id"`
	Status                string `json:"status,omitempty"`
	Date                  string `json:"date,omitempty"`
	Size                  any    `json:"size,omitempty"`
	Duration              any    `json:"duration,omitempty"`
}

type BackupConfig struct {
	ID         int64    `json:"id"`
	ServerID   int64    `json:"server_id,omitempty"`
	DayOfWeek  *int     `json:"day_of_week,omitempty"`
	Time       string   `json:"time,omitempty"`
	Provider   string   `json:"provider,omitempty"`
	Frequency  string   `json:"frequency,omitempty"`
	Databases  []int64  `json:"databases,omitempty"`
	Backups    []Backup `json:"backups,omitempty"`
	BackupTime string   `json:"backup_time,omitempty"`
}

type SiteCommand struct {
	ID              int64  `json:"id"`
	ServerID        int64  `json:"server_id,omitempty"`
	SiteID          int64  `json:"site_id"`
	UserID          int64  `json:"user_id,omitempty"`
	Command         string `json:"command"`
	Status          string `json:"status,omitempty"`
	CreatedAt       string `json:"created_at,omitempty"`
	Duration        any    `json:"duration,omitempty"`
	ProfilePhotoURL string `json:"profile_photo_url,omitempty"`
	UserName        string `json:"user_name,omitempty"`
}

type RedirectRule struct {
	ID      int64  `json:"id"`
	FromURL string `json:"from"`
	To      string `json:"to"`
	Type    string `json:"type,omitempty"`
	Status  string `json:"status,omitempty"`
}
```

**Step 3: Write client core with test**

```go
// internal/forge/client_test.go
package forge

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func testClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	c := NewClient("test-token")
	c.BaseURL = srv.URL
	return c
}

func TestListServers(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Error("missing auth header")
		}
		if r.URL.Path != "/servers" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"servers": []map[string]any{
				{"id": 1, "name": "prod-1", "ip_address": "1.2.3.4"},
				{"id": 2, "name": "staging-1", "ip_address": "5.6.7.8"},
			},
		})
	})
	servers, err := c.Servers.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(servers) != 2 {
		t.Fatalf("expected 2 servers, got %d", len(servers))
	}
	if servers[0].Name != "prod-1" {
		t.Errorf("expected name 'prod-1', got %q", servers[0].Name)
	}
}

func TestAuthError(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		w.Write([]byte(`{"message": "Unauthenticated."}`))
	})
	_, err := c.Servers.List(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	if _, ok := err.(*AuthenticationError); !ok {
		t.Errorf("expected AuthenticationError, got %T", err)
	}
}

func TestRateLimitError(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(429)
		w.Write([]byte(`{"message": "Too Many Requests"}`))
	})
	_, err := c.Servers.List(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	if _, ok := err.(*RateLimitError); !ok {
		t.Errorf("expected RateLimitError, got %T", err)
	}
}
```

**Step 4: Implement client core**

```go
// internal/forge/client.go
package forge

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const defaultBaseURL = "https://forge.laravel.com/api/v1"

type Client struct {
	BaseURL string
	token   string
	http    *http.Client

	// Services
	Servers      *ServersService
	Sites        *SitesService
	Deployments  *DeploymentsService
	Databases    *DatabasesService
	Environment  *EnvironmentService
	Certificates *CertificatesService
	Workers      *WorkersService
	Daemons      *DaemonsService
	Firewall     *FirewallService
	Jobs         *JobsService
	Backups      *BackupsService
	SSHKeys      *SSHKeysService
	Commands     *CommandsService
	Git          *GitService
	Logs         *LogsService
}

func NewClient(token string) *Client {
	c := &Client{
		BaseURL: defaultBaseURL,
		token:   token,
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
	c.Servers = &ServersService{client: c}
	c.Sites = &SitesService{client: c}
	c.Deployments = &DeploymentsService{client: c}
	c.Databases = &DatabasesService{client: c}
	c.Environment = &EnvironmentService{client: c}
	c.Certificates = &CertificatesService{client: c}
	c.Workers = &WorkersService{client: c}
	c.Daemons = &DaemonsService{client: c}
	c.Firewall = &FirewallService{client: c}
	c.Jobs = &JobsService{client: c}
	c.Backups = &BackupsService{client: c}
	c.SSHKeys = &SSHKeysService{client: c}
	c.Commands = &CommandsService{client: c}
	c.Git = &GitService{client: c}
	c.Logs = &LogsService{client: c}
	return c
}

func (c *Client) request(ctx context.Context, method, path string, body any) (*http.Response, error) {
	url := c.BaseURL + path

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return c.http.Do(req)
}

func (c *Client) do(ctx context.Context, method, path string, body any, result any) error {
	resp, err := c.request(ctx, method, path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return parseError(resp)
	}

	if result == nil || resp.StatusCode == 204 {
		return nil
	}

	return json.NewDecoder(resp.Body).Decode(result)
}

func (c *Client) getText(ctx context.Context, path string) (string, error) {
	resp, err := c.request(ctx, http.MethodGet, path, nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", parseError(resp)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func parseError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)
	msg := string(body)

	// Try to extract message from JSON
	var envelope struct {
		Message string `json:"message"`
	}
	if json.Unmarshal(body, &envelope) == nil && envelope.Message != "" {
		msg = envelope.Message
	}

	base := APIError{StatusCode: resp.StatusCode, Message: msg}

	switch resp.StatusCode {
	case 401:
		return &AuthenticationError{base}
	case 404:
		return &NotFoundError{base}
	case 422:
		ve := &ValidationError{APIError: base}
		var details struct {
			Errors map[string][]string `json:"errors"`
		}
		if json.Unmarshal(body, &details) == nil {
			ve.Details = details.Errors
		}
		return ve
	case 429:
		return &RateLimitError{base}
	default:
		return &base
	}
}

// Service stubs - each will be implemented in its own file.
// For now, declare the types so the client compiles.

type ServersService struct{ client *Client }
type SitesService struct{ client *Client }
type DeploymentsService struct{ client *Client }
type DatabasesService struct{ client *Client }
type EnvironmentService struct{ client *Client }
type CertificatesService struct{ client *Client }
type WorkersService struct{ client *Client }
type DaemonsService struct{ client *Client }
type FirewallService struct{ client *Client }
type JobsService struct{ client *Client }
type BackupsService struct{ client *Client }
type SSHKeysService struct{ client *Client }
type CommandsService struct{ client *Client }
type GitService struct{ client *Client }
type LogsService struct{ client *Client }
```

**Step 5: Run tests**

```bash
go test ./internal/forge/ -v
# Expected: PASS (ListServers, AuthError, RateLimitError tests)
```

**Step 6: Commit**

```bash
git add internal/forge/
git commit -m "feat: add Forge API client core with types and error handling"
```

---

## Task 4: Forge API Services - All Endpoints

**Files:**
- Create: `internal/forge/servers.go`
- Create: `internal/forge/sites.go`
- Create: `internal/forge/deployments.go`
- Create: `internal/forge/databases.go`
- Create: `internal/forge/environment.go`
- Create: `internal/forge/certificates.go`
- Create: `internal/forge/workers.go`
- Create: `internal/forge/daemons.go`
- Create: `internal/forge/firewall.go`
- Create: `internal/forge/jobs.go`
- Create: `internal/forge/backups.go`
- Create: `internal/forge/ssh_keys.go`
- Create: `internal/forge/commands.go`
- Create: `internal/forge/git.go`
- Create: `internal/forge/logs.go`

Each service file follows the same pattern. Here is the full implementation for `servers.go` as the exemplar:

**Step 1: Implement servers service**

```go
// internal/forge/servers.go
package forge

import (
	"context"
	"fmt"
)

func (s *ServersService) List(ctx context.Context) ([]Server, error) {
	var resp struct {
		Servers []Server `json:"servers"`
	}
	err := s.client.do(ctx, "GET", "/servers", nil, &resp)
	return resp.Servers, err
}

func (s *ServersService) Get(ctx context.Context, serverID int64) (*Server, error) {
	var resp struct {
		Server Server `json:"server"`
	}
	err := s.client.do(ctx, "GET", fmt.Sprintf("/servers/%d", serverID), nil, &resp)
	return &resp.Server, err
}

func (s *ServersService) Reboot(ctx context.Context, serverID int64) error {
	return s.client.do(ctx, "POST", fmt.Sprintf("/servers/%d/reboot", serverID), nil, nil)
}

func (s *ServersService) GetUser(ctx context.Context) (*User, error) {
	var resp struct {
		User User `json:"user"`
	}
	err := s.client.do(ctx, "GET", "/user", nil, &resp)
	return &resp.User, err
}
```

**Step 2: Implement all remaining services**

Each service follows the same pattern. The exact endpoints (from the current Python app):

**sites.go:**
- `List(ctx, serverID) → []Site` — `GET /servers/{id}/sites`
- `Get(ctx, serverID, siteID) → *Site` — `GET /servers/{id}/sites/{id}`
- `UpdateAliases(ctx, serverID, siteID, aliases []string) → *Site` — `PUT /servers/{id}/sites/{id}/aliases`
- `UpdatePHP(ctx, serverID, siteID, version string)` — `PUT /servers/{id}/sites/{id}/php`

**deployments.go:**
- `List(ctx, serverID, siteID) → []Deployment` — `GET /servers/{id}/sites/{id}/deployment-history`
- `Get(ctx, serverID, siteID, deployID) → *Deployment` — `GET /servers/{id}/sites/{id}/deployment-history/{id}`
- `GetOutput(ctx, serverID, siteID, deployID) → string` — `GET .../output` (json, key "output")
- `Deploy(ctx, serverID, siteID)` — `POST /servers/{id}/sites/{id}/deployment/deploy`
- `GetLog(ctx, serverID, siteID) → string` — `GET .../deployment/log` (json, key "output")
- `GetScript(ctx, serverID, siteID) → string` — `GET .../deployment/script` (plain text, use getText)
- `UpdateScript(ctx, serverID, siteID, content string)` — `PUT .../deployment/script`
- `EnableQuickDeploy(ctx, serverID, siteID)` — `POST .../deployment`
- `DisableQuickDeploy(ctx, serverID, siteID)` — `DELETE .../deployment`
- `ResetStatus(ctx, serverID, siteID)` — `POST .../deployment/reset`

**databases.go:**
- `List(ctx, serverID) → []Database` — `GET /servers/{id}/databases`
- `Get(ctx, serverID, dbID) → *Database` — `GET /servers/{id}/databases/{id}`
- `Create(ctx, serverID, name, user, password string) → *Database` — `POST /servers/{id}/databases`
- `Delete(ctx, serverID, dbID)` — `DELETE /servers/{id}/databases/{id}`
- `Sync(ctx, serverID)` — `POST /servers/{id}/databases/sync`
- `ListUsers(ctx, serverID) → []DatabaseUser` — `GET /servers/{id}/database-users`
- `GetUser(ctx, serverID, userID) → *DatabaseUser` — `GET /servers/{id}/database-users/{id}`
- `CreateUser(ctx, serverID, name, password string, databases []int64) → *DatabaseUser` — `POST`
- `UpdateUser(ctx, serverID, userID int64, databases []int64) → *DatabaseUser` — `PUT`
- `DeleteUser(ctx, serverID, userID)` — `DELETE /servers/{id}/database-users/{id}`

**environment.go:**
- `Get(ctx, serverID, siteID) → string` — `GET .../env` (plain text, use getText)
- `Update(ctx, serverID, siteID, content string)` — `PUT .../env`

**certificates.go:**
- `List(ctx, serverID, siteID) → []Certificate` — `GET .../certificates`
- `Get(ctx, serverID, siteID, certID) → *Certificate` — `GET .../certificates/{id}`
- `CreateLetsEncrypt(ctx, serverID, siteID, domains []string) → *Certificate` — `POST .../certificates/letsencrypt`
- `Activate(ctx, serverID, siteID, certID)` — `POST .../certificates/{id}/activate`
- `Delete(ctx, serverID, siteID, certID)` — `DELETE .../certificates/{id}`

**workers.go:**
- `List(ctx, serverID, siteID) → []Worker` — `GET .../workers`
- `Get(ctx, serverID, siteID, workerID) → *Worker` — `GET .../workers/{id}`
- `Create(ctx, serverID, siteID, opts WorkerCreateOpts) → *Worker` — `POST .../workers`
- `Restart(ctx, serverID, siteID, workerID)` — `POST .../workers/{id}/restart`
- `Delete(ctx, serverID, siteID, workerID)` — `DELETE .../workers/{id}`

**daemons.go:**
- `List(ctx, serverID) → []Daemon` — `GET /servers/{id}/daemons`
- `Get(ctx, serverID, daemonID) → *Daemon` — `GET /servers/{id}/daemons/{id}`
- `Create(ctx, serverID, opts DaemonCreateOpts) → *Daemon` — `POST /servers/{id}/daemons`
- `Restart(ctx, serverID, daemonID)` — `POST /servers/{id}/daemons/{id}/restart`
- `Delete(ctx, serverID, daemonID)` — `DELETE /servers/{id}/daemons/{id}`

**firewall.go:**
- `List(ctx, serverID) → []FirewallRule` — `GET /servers/{id}/firewall-rules`
- `Get(ctx, serverID, ruleID) → *FirewallRule` — `GET /servers/{id}/firewall-rules/{id}`
- `Create(ctx, serverID, opts FirewallCreateOpts) → *FirewallRule` — `POST`
- `Delete(ctx, serverID, ruleID)` — `DELETE /servers/{id}/firewall-rules/{id}`

**jobs.go:**
- `List(ctx, serverID) → []ScheduledJob` — `GET /servers/{id}/jobs`
- `Get(ctx, serverID, jobID) → *ScheduledJob` — `GET /servers/{id}/jobs/{id}`
- `Create(ctx, serverID, opts JobCreateOpts) → *ScheduledJob` — `POST /servers/{id}/jobs`
- `Delete(ctx, serverID, jobID)` — `DELETE /servers/{id}/jobs/{id}`

**backups.go:**
- `ListConfigs(ctx, serverID) → []BackupConfig` — `GET /servers/{id}/backup-configs`
- `GetConfig(ctx, serverID, configID) → *BackupConfig` — `GET /servers/{id}/backup-configs/{id}`
- `CreateConfig(ctx, serverID, opts BackupConfigCreateOpts) → *BackupConfig` — `POST`
- `DeleteConfig(ctx, serverID, configID)` — `DELETE /servers/{id}/backup-configs/{id}`
- `RunBackup(ctx, serverID, configID)` — `POST /servers/{id}/backup-configs/{id}`
- `RestoreBackup(ctx, serverID, configID, backupID)` — `POST .../backups/{id}`
- `DeleteBackup(ctx, serverID, configID, backupID)` — `DELETE .../backups/{id}`

**ssh_keys.go:**
- `List(ctx, serverID) → []SSHKey` — `GET /servers/{id}/keys`
- `Get(ctx, serverID, keyID) → *SSHKey` — `GET /servers/{id}/keys/{id}`
- `Create(ctx, serverID, name, key, username string) → *SSHKey` — `POST /servers/{id}/keys`
- `Delete(ctx, serverID, keyID)` — `DELETE /servers/{id}/keys/{id}`

**commands.go:**
- `List(ctx, serverID, siteID) → []SiteCommand` — `GET .../commands`
- `Get(ctx, serverID, siteID, cmdID) → *SiteCommand` — `GET .../commands/{id}`
- `Create(ctx, serverID, siteID, command string) → *SiteCommand` — `POST .../commands`

**git.go:**
- `Install(ctx, serverID, siteID, provider, repo, branch string, composer bool)` — `POST .../git`
- `UpdateBranch(ctx, serverID, siteID, branch string)` — `PUT .../git`
- `Remove(ctx, serverID, siteID)` — `DELETE .../git`

**logs.go:**
- `GetServerLog(ctx, serverID) → string` — `GET /servers/{id}/logs` (json, key "content")
- `GetSiteLog(ctx, serverID, siteID) → string` — `GET .../sites/{id}/logs` (json, key "content")
- `ClearSiteLog(ctx, serverID, siteID)` — `DELETE .../sites/{id}/logs`

**Step 3: Move service type declarations out of client.go**

Remove the service type stubs from the bottom of `client.go` since each service file now declares its own type.

**Step 4: Write tests for 2-3 representative services**

Test `Sites.List`, `Deployments.Deploy`, `Environment.Get` (text endpoint) to cover the main patterns: JSON list, POST action, plain text response.

**Step 5: Run all tests**

```bash
go test ./internal/forge/ -v
# Expected: PASS
```

**Step 6: Commit**

```bash
git add internal/forge/
git commit -m "feat: implement all Forge API service endpoints"
```

---

## Task 5: TUI Foundation - Styles, Keymaps, Root Model

**Files:**
- Create: `internal/tui/styles.go`
- Create: `internal/tui/keymap.go`
- Create: `internal/tui/app.go`
- Create: `internal/tui/messages.go`

**Step 1: Define Lip Gloss styles**

```go
// internal/tui/styles.go
package tui

import "github.com/charmbracelet/lipgloss/v2"

var (
	// Panel borders
	ActiveBorderStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")) // blue-ish

	InactiveBorderStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")) // gray

	// Panel titles
	TitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("62"))

	// Help bar at bottom
	HelpBarStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	HelpKeyStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("62"))

	// List items
	SelectedItemStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("62")).
		Bold(true)

	NormalItemStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	// Status indicators
	ActiveStatusStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("42")) // green

	ErrorStatusStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")) // red

	// Toast
	ToastStyle = lipgloss.NewStyle().
		Padding(0, 1).
		Background(lipgloss.Color("62")).
		Foreground(lipgloss.Color("230"))

	ToastErrorStyle = lipgloss.NewStyle().
		Padding(0, 1).
		Background(lipgloss.Color("196")).
		Foreground(lipgloss.Color("230"))
)
```

**Step 2: Define keybindings**

```go
// internal/tui/keymap.go
package tui

import "github.com/charmbracelet/bubbles/v2/key"

// Global keybindings (always active)
type GlobalKeyMap struct {
	Quit    key.Binding
	Refresh key.Binding
	SSH     key.Binding
	SFTP    key.Binding
	DB      key.Binding
	Help    key.Binding
	Tab     key.Binding
	ShiftTab key.Binding
}

var GlobalKeys = GlobalKeyMap{
	Quit:     key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	Refresh:  key.NewBinding(key.WithKeys("ctrl+r"), key.WithHelp("C-r", "refresh")),
	SSH:      key.NewBinding(key.WithKeys("ctrl+s"), key.WithHelp("C-s", "ssh")),
	SFTP:     key.NewBinding(key.WithKeys("ctrl+f"), key.WithHelp("C-f", "sftp")),
	DB:       key.NewBinding(key.WithKeys("ctrl+d"), key.WithHelp("C-d", "database")),
	Help:     key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
	Tab:      key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next panel")),
	ShiftTab: key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("S-tab", "prev panel")),
}

// Navigation keybindings (within lists)
type NavKeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Enter  key.Binding
	Back   key.Binding
	Search key.Binding
	Home   key.Binding
	End    key.Binding
}

var NavKeys = NavKeyMap{
	Up:     key.NewBinding(key.WithKeys("k", "up"), key.WithHelp("k", "up")),
	Down:   key.NewBinding(key.WithKeys("j", "down"), key.WithHelp("j", "down")),
	Enter:  key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
	Back:   key.NewBinding(key.WithKeys("esc", "backspace"), key.WithHelp("esc", "back")),
	Search: key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "search")),
	Home:   key.NewBinding(key.WithKeys("g", "home"), key.WithHelp("g", "top")),
	End:    key.NewBinding(key.WithKeys("G", "end"), key.WithHelp("G", "bottom")),
}

// Section tab keybindings (1-9 for sub-sections)
type SectionKeyMap struct {
	Deployments key.Binding
	Environment key.Binding
	Databases   key.Binding
	SSL         key.Binding
	Workers     key.Binding
	Commands    key.Binding
	Logs        key.Binding
	Git         key.Binding
	Domains     key.Binding
}

var SectionKeys = SectionKeyMap{
	Deployments: key.NewBinding(key.WithKeys("1"), key.WithHelp("1", "deploys")),
	Environment: key.NewBinding(key.WithKeys("2"), key.WithHelp("2", "env")),
	Databases:   key.NewBinding(key.WithKeys("3"), key.WithHelp("3", "databases")),
	SSL:         key.NewBinding(key.WithKeys("4"), key.WithHelp("4", "ssl")),
	Workers:     key.NewBinding(key.WithKeys("5"), key.WithHelp("5", "workers")),
	Commands:    key.NewBinding(key.WithKeys("6"), key.WithHelp("6", "commands")),
	Logs:        key.NewBinding(key.WithKeys("7"), key.WithHelp("7", "logs")),
	Git:         key.NewBinding(key.WithKeys("8"), key.WithHelp("8", "git")),
	Domains:     key.NewBinding(key.WithKeys("9"), key.WithHelp("9", "domains")),
}

// Context-sensitive action keys (change depending on what's selected)
type ServerActionKeyMap struct {
	SSH    key.Binding
	SFTP   key.Binding
	Reboot key.Binding
}

var ServerActionKeys = ServerActionKeyMap{
	SSH:    key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "ssh")),
	SFTP:   key.NewBinding(key.WithKeys("f"), key.WithHelp("f", "sftp")),
	Reboot: key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "reboot")),
}

type SiteActionKeyMap struct {
	Deploy   key.Binding
	Env      key.Binding
	SSH      key.Binding
	Database key.Binding
	Logs     key.Binding
}

var SiteActionKeys = SiteActionKeyMap{
	Deploy:   key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "deploy")),
	Env:      key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "env")),
	SSH:      key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "ssh")),
	Database: key.NewBinding(key.WithKeys("D"), key.WithHelp("D", "database")),
	Logs:     key.NewBinding(key.WithKeys("l"), key.WithHelp("l", "logs")),
}
```

**Step 3: Define message types**

```go
// internal/tui/messages.go
package tui

import (
	"github.com/hinke/phorge/internal/forge"
)

// Data loading messages
type serversLoadedMsg struct{ servers []forge.Server }
type sitesLoadedMsg struct{ sites []forge.Site }
type errMsg struct{ err error }

// Selection messages
type serverSelectedMsg struct{ server *forge.Server }
type siteSelectedMsg struct{ site *forge.Site }

// Action result messages
type deployResultMsg struct{ err error }
type rebootResultMsg struct{ err error }

// Toast messages
type toastMsg struct {
	message string
	isError bool
}

// SSH/external tool exit
type externalExitMsg struct{ err error }
```

**Step 4: Create root app model (skeleton)**

```go
// internal/tui/app.go
package tui

import (
	"github.com/charmbracelet/bubbles/v2/key"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/hinke/phorge/internal/config"
	"github.com/hinke/phorge/internal/forge"
)

type Focus int

const (
	FocusServerList Focus = iota
	FocusContextList
	FocusDetailPanel
)

type App struct {
	forge  *forge.Client
	config *config.Config

	focus     Focus
	width     int
	height    int

	// Data
	servers      []forge.Server
	selectedSrv  *forge.Server
	sites        []forge.Site
	selectedSite *forge.Site
	activeTab    int

	// UI state
	serverCursor int
	siteCursor   int
	toast        string
	toastIsErr   bool
	loading      bool
}

func NewApp(cfg *config.Config) App {
	return App{
		forge:     forge.NewClient(cfg.Forge.APIKey),
		config:    cfg,
		focus:     FocusServerList,
		activeTab: 1,
	}
}

func (m App) Init() (App, tea.Cmd) {
	return m, m.fetchServers()
}

func (m App) Update(msg tea.Msg) (App, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		// Global keys first
		switch {
		case key.Matches(msg, GlobalKeys.Quit):
			return m, tea.Quit
		case key.Matches(msg, GlobalKeys.Tab):
			m.focus = (m.focus + 1) % 3
			return m, nil
		case key.Matches(msg, GlobalKeys.ShiftTab):
			m.focus = (m.focus + 2) % 3
			return m, nil
		case key.Matches(msg, GlobalKeys.Refresh):
			return m, m.fetchServers()
		}

		// Route to focused panel
		switch m.focus {
		case FocusServerList:
			return m.updateServerList(msg)
		case FocusContextList:
			return m.updateContextList(msg)
		case FocusDetailPanel:
			return m.updateDetailPanel(msg)
		}

	case serversLoadedMsg:
		m.servers = msg.servers
		m.loading = false
		if len(m.servers) > 0 && m.selectedSrv == nil {
			m.selectedSrv = &m.servers[0]
			return m, m.fetchSites(m.selectedSrv.ID)
		}
		return m, nil

	case sitesLoadedMsg:
		m.sites = msg.sites
		if len(m.sites) > 0 {
			m.selectedSite = &m.sites[0]
		}
		return m, nil

	case toastMsg:
		m.toast = msg.message
		m.toastIsErr = msg.isError
		return m, nil

	case errMsg:
		m.toast = msg.err.Error()
		m.toastIsErr = true
		m.loading = false
		return m, nil
	}

	return m, nil
}

func (m App) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	leftWidth := m.width / 4
	rightWidth := m.width - leftWidth - 3 // border padding

	// Left panel: server list
	left := m.renderServerList(leftWidth)

	// Right panels: context + detail
	contextHeight := (m.height - 4) / 2
	detailHeight := m.height - 4 - contextHeight

	context := m.renderContextList(rightWidth, contextHeight)
	detail := m.renderDetailPanel(rightWidth, detailHeight)
	right := lipgloss.JoinVertical(lipgloss.Left, context, detail)

	main := lipgloss.JoinHorizontal(lipgloss.Top, left, right)

	// Help bar at bottom
	helpBar := m.renderHelpBar()

	return lipgloss.JoinVertical(lipgloss.Left, main, helpBar)
}

// --- Panel update handlers ---

func (m App) updateServerList(msg tea.KeyMsg) (App, tea.Cmd) {
	switch {
	case key.Matches(msg, NavKeys.Down):
		if m.serverCursor < len(m.servers)-1 {
			m.serverCursor++
			m.selectedSrv = &m.servers[m.serverCursor]
			m.siteCursor = 0
			m.selectedSite = nil
			return m, m.fetchSites(m.selectedSrv.ID)
		}
	case key.Matches(msg, NavKeys.Up):
		if m.serverCursor > 0 {
			m.serverCursor--
			m.selectedSrv = &m.servers[m.serverCursor]
			m.siteCursor = 0
			m.selectedSite = nil
			return m, m.fetchSites(m.selectedSrv.ID)
		}
	case key.Matches(msg, NavKeys.Enter):
		m.focus = FocusContextList
	}
	return m, nil
}

func (m App) updateContextList(msg tea.KeyMsg) (App, tea.Cmd) {
	switch {
	case key.Matches(msg, NavKeys.Down):
		if m.siteCursor < len(m.sites)-1 {
			m.siteCursor++
			m.selectedSite = &m.sites[m.siteCursor]
		}
	case key.Matches(msg, NavKeys.Up):
		if m.siteCursor > 0 {
			m.siteCursor--
			m.selectedSite = &m.sites[m.siteCursor]
		}
	case key.Matches(msg, NavKeys.Back):
		m.focus = FocusServerList
	}
	return m, nil
}

func (m App) updateDetailPanel(msg tea.KeyMsg) (App, tea.Cmd) {
	switch {
	case key.Matches(msg, NavKeys.Back):
		m.focus = FocusContextList
	}
	return m, nil
}

// --- Render helpers ---

func (m App) renderServerList(width int) string {
	style := InactiveBorderStyle
	if m.focus == FocusServerList {
		style = ActiveBorderStyle
	}

	content := TitleStyle.Render("Servers") + "\n"
	for i, s := range m.servers {
		name := s.Name
		if i == m.serverCursor {
			content += SelectedItemStyle.Render("> " + name) + "\n"
		} else {
			content += NormalItemStyle.Render("  " + name) + "\n"
		}
	}

	return style.Width(width).Height(m.height - 4).Render(content)
}

func (m App) renderContextList(width, height int) string {
	style := InactiveBorderStyle
	if m.focus == FocusContextList {
		style = ActiveBorderStyle
	}

	// Tab bar
	tabs := ""
	tabNames := []string{"Deploys", "Env", "DB", "SSL", "Workers", "Cmds", "Logs", "Git", "Domains"}
	for i, name := range tabNames {
		if i+1 == m.activeTab {
			tabs += HelpKeyStyle.Render(name) + " "
		} else {
			tabs += HelpBarStyle.Render(name) + " "
		}
	}

	content := TitleStyle.Render("Sites") + "\n"
	for i, s := range m.sites {
		if i == m.siteCursor {
			content += SelectedItemStyle.Render("> " + s.Name) + "\n"
		} else {
			content += NormalItemStyle.Render("  " + s.Name) + "\n"
		}
	}

	return style.Width(width).Height(height).Render(tabs + "\n" + content)
}

func (m App) renderDetailPanel(width, height int) string {
	style := InactiveBorderStyle
	if m.focus == FocusDetailPanel {
		style = ActiveBorderStyle
	}

	content := TitleStyle.Render("Detail") + "\n"
	if m.selectedSite != nil {
		content += "Site: " + m.selectedSite.Name + "\n"
		if m.selectedSite.Repository != "" {
			content += "Repo: " + m.selectedSite.Repository + "\n"
		}
		if m.selectedSite.RepositoryBranch != "" {
			content += "Branch: " + m.selectedSite.RepositoryBranch + "\n"
		}
		content += "Status: " + m.selectedSite.Status + "\n"
	} else if m.selectedSrv != nil {
		content += "Server: " + m.selectedSrv.Name + "\n"
		content += "IP: " + m.selectedSrv.IPAddress + "\n"
		content += "PHP: " + m.selectedSrv.PHPVersion + "\n"
		content += "Status: " + m.selectedSrv.Status + "\n"
	}

	return style.Width(width).Height(height).Render(content)
}

func (m App) renderHelpBar() string {
	bindings := ""
	switch m.focus {
	case FocusServerList:
		bindings = HelpKeyStyle.Render("s") + " ssh  " +
			HelpKeyStyle.Render("f") + " sftp  " +
			HelpKeyStyle.Render("r") + " reboot  " +
			HelpKeyStyle.Render("enter") + " sites"
	case FocusContextList:
		bindings = HelpKeyStyle.Render("d") + " deploy  " +
			HelpKeyStyle.Render("e") + " env  " +
			HelpKeyStyle.Render("s") + " ssh  " +
			HelpKeyStyle.Render("D") + " database  " +
			HelpKeyStyle.Render("l") + " logs"
	case FocusDetailPanel:
		bindings = HelpKeyStyle.Render("esc") + " back"
	}

	global := HelpKeyStyle.Render("tab") + " switch  " +
		HelpKeyStyle.Render("?") + " help  " +
		HelpKeyStyle.Render("q") + " quit"

	return HelpBarStyle.Render(bindings + "  │  " + global)
}

// --- API command helpers ---

func (m App) fetchServers() tea.Cmd {
	client := m.forge
	return func() tea.Msg {
		servers, err := client.Servers.List(nil)
		if err != nil {
			return errMsg{err}
		}
		return serversLoadedMsg{servers}
	}
}

func (m App) fetchSites(serverID int64) tea.Cmd {
	client := m.forge
	return func() tea.Msg {
		sites, err := client.Sites.List(nil, serverID)
		if err != nil {
			return errMsg{err}
		}
		return sitesLoadedMsg{sites}
	}
}
```

**Step 5: Wire up main.go to use the App**

Update `cmd/phorge/main.go`:
```go
package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/hinke/phorge/internal/config"
	"github.com/hinke/phorge/internal/tui"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	if cfg.Forge.APIKey == "" {
		fmt.Fprintln(os.Stderr, "No API key configured. Set it in ~/.config/phorge/config.toml")
		os.Exit(1)
	}

	p := tea.NewProgram(tui.NewApp(cfg), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
```

**Step 6: Build and verify**

```bash
go build ./cmd/phorge
# Should compile. Manual test with a real API key to verify the three-panel layout renders.
```

**Step 7: Commit**

```bash
git add internal/tui/ cmd/phorge/
git commit -m "feat: add TUI foundation with three-panel layout and keybindings"
```

---

## Task 6: Reusable Components - Confirm, Input, Toast

**Files:**
- Create: `internal/tui/components/confirm.go`
- Create: `internal/tui/components/input.go`
- Create: `internal/tui/components/toast.go`

These are overlay/modal sub-models that the root app can activate.

**Step 1: Implement confirm dialog**

A confirmation modal that shows a question and accepts y/n:

```go
// internal/tui/components/confirm.go
package components

import (
	"github.com/charmbracelet/bubbles/v2/key"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
)

type ConfirmResult struct {
	Confirmed bool
	ID        string // caller-defined identifier
}

type Confirm struct {
	Question string
	ID       string
	Active   bool
}

func NewConfirm(id, question string) Confirm {
	return Confirm{ID: id, Question: question, Active: true}
}

func (c Confirm) Update(msg tea.Msg) (Confirm, tea.Cmd) {
	if !c.Active {
		return c, nil
	}
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("y", "Y"))):
			c.Active = false
			return c, func() tea.Msg { return ConfirmResult{Confirmed: true, ID: c.ID} }
		case key.Matches(msg, key.NewBinding(key.WithKeys("n", "N", "esc"))):
			c.Active = false
			return c, func() tea.Msg { return ConfirmResult{Confirmed: false, ID: c.ID} }
		}
	}
	return c, nil
}

func (c Confirm) View(width, height int) string {
	if !c.Active {
		return ""
	}
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2).
		Width(40).
		Align(lipgloss.Center).
		Render(c.Question + "\n\n[y]es  [n]o")

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, box)
}
```

**Step 2: Implement text input modal**

```go
// internal/tui/components/input.go
package components

import (
	"github.com/charmbracelet/bubbles/v2/key"
	"github.com/charmbracelet/bubbles/v2/textinput"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
)

type InputResult struct {
	Value string
	ID    string
}

type InputCancelled struct {
	ID string
}

type Input struct {
	Label  string
	ID     string
	Active bool
	input  textinput.Model
}

func NewInput(id, label, placeholder string) Input {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Focus()
	return Input{ID: id, Label: label, Active: true, input: ti}
}

func (i Input) Update(msg tea.Msg) (Input, tea.Cmd) {
	if !i.Active {
		return i, nil
	}
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			i.Active = false
			val := i.input.Value()
			id := i.ID
			return i, func() tea.Msg { return InputResult{Value: val, ID: id} }
		case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
			i.Active = false
			id := i.ID
			return i, func() tea.Msg { return InputCancelled{ID: id} }
		}
	}
	var cmd tea.Cmd
	i.input, cmd = i.input.Update(msg)
	return i, cmd
}

func (i Input) View(width, height int) string {
	if !i.Active {
		return ""
	}
	content := i.Label + "\n\n" + i.input.View()
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2).
		Width(50).
		Render(content)

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, box)
}
```

**Step 3: Implement toast notification**

```go
// internal/tui/components/toast.go
package components

import (
	"time"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
)

type toastTimeoutMsg struct{}

type Toast struct {
	Message string
	IsError bool
	Active  bool
}

func ShowToast(message string, isError bool) (Toast, tea.Cmd) {
	t := Toast{Message: message, IsError: isError, Active: true}
	return t, t.timeout()
}

func (t Toast) timeout() tea.Cmd {
	return tea.Tick(3*time.Second, func(time.Time) tea.Msg {
		return toastTimeoutMsg{}
	})
}

func (t Toast) Update(msg tea.Msg) (Toast, tea.Cmd) {
	if _, ok := msg.(toastTimeoutMsg); ok {
		t.Active = false
	}
	return t, nil
}

func (t Toast) View(width int) string {
	if !t.Active {
		return ""
	}
	style := lipgloss.NewStyle().
		Padding(0, 1).
		Background(lipgloss.Color("62")).
		Foreground(lipgloss.Color("230"))

	if t.IsError {
		style = style.Background(lipgloss.Color("196"))
	}

	return lipgloss.Place(width, 1, lipgloss.Center, lipgloss.Center, style.Render(t.Message))
}
```

**Step 4: Commit**

```bash
git add internal/tui/components/
git commit -m "feat: add confirm, input, and toast components"
```

---

## Task 7: Panel Interface & Server/Site List Panels

**Files:**
- Create: `internal/tui/panels/panel.go` (interface)
- Create: `internal/tui/panels/servers.go`
- Create: `internal/tui/panels/sites.go`

**Step 1: Define panel interface**

```go
// internal/tui/panels/panel.go
package panels

import tea "github.com/charmbracelet/bubbletea/v2"

// Panel is the interface all detail/context panels implement.
type Panel interface {
	Update(msg tea.Msg) (Panel, tea.Cmd)
	View(width, height int, focused bool) string
	// HelpBindings returns the context-sensitive key hints for the help bar.
	HelpBindings() []HelpBinding
}

type HelpBinding struct {
	Key  string
	Desc string
}
```

**Step 2: Implement server list panel with j/k navigation, cursor, and rendering**

This replaces the inline server rendering in `app.go` with a proper sub-model.

**Step 3: Implement site list panel**

Same pattern as servers but for the context list area.

**Step 4: Integrate panels into root app model**

Refactor `app.go` to delegate rendering and updates to the panel sub-models rather than inline methods.

**Step 5: Verify build and test manually**

```bash
go build ./cmd/phorge && ./phorge
```

**Step 6: Commit**

```bash
git add internal/tui/panels/
git commit -m "feat: add panel interface with server and site list panels"
```

---

## Task 8: Detail Panels - Server Info & Site Info

**Files:**
- Create: `internal/tui/panels/server_info.go`
- Create: `internal/tui/panels/site_info.go`

These render formatted detail views of the selected server/site in the bottom-right panel.

**Step 1: Implement server info panel**

Displays: name, IP, provider, region, PHP version, Ubuntu version, status, database type, SSH port. Uses Lip Gloss for formatting key-value pairs.

**Step 2: Implement site info panel**

Displays: name, directory, repository, branch, PHP version, status, quick deploy status, SSL status, project type.

**Step 3: Wire into root app - when a server/site is selected and the detail panel is focused, show the appropriate info panel**

**Step 4: Commit**

```bash
git commit -m "feat: add server and site info detail panels"
```

---

## Task 9: Deployments Panel

**Files:**
- Create: `internal/tui/panels/deployments.go`

**Step 1: Implement deployment list panel**

- Fetches deployment history via `forge.Deployments.List()`
- Renders as a scrollable list with: commit message, author, status, time ago
- Actions: `d` deploy now, `r` reset status, `Enter` view output
- Deploy action triggers confirmation dialog, then calls `forge.Deployments.Deploy()`
- Output view uses a viewport component for scrollable text

**Step 2: Wire into tab system - when activeTab=1 and a site is selected, show DeploymentsPanel**

**Step 3: Test manually - navigate to a site, press `1`, verify deployment list loads**

**Step 4: Commit**

```bash
git commit -m "feat: add deployments panel with deploy and view output"
```

---

## Task 10: Deployment Script & Environment Panels

**Files:**
- Create: `internal/tui/panels/deploy_script.go`
- Create: `internal/tui/panels/environment.go`

**Step 1: Implement deployment script panel**

- Fetches script via `forge.Deployments.GetScript()` (plain text)
- Shows in a viewport with syntax-like rendering
- Action: `e` opens in external editor using `tea.ExecProcess()`
- After editor exits, reads the modified file and calls `forge.Deployments.UpdateScript()`

**Step 2: Implement environment panel**

- Fetches .env via `forge.Environment.Get()` (plain text)
- Shows in a viewport
- Action: `e` opens in external editor
- After editor exits, detects changes and uploads via `forge.Environment.Update()`
- Both panels use temp files: write content to temp file, open editor, read back, compare, upload if changed

**Step 3: Wire into tab system (deployment script accessible from deployments panel, env is tab 2)**

**Step 4: Commit**

```bash
git commit -m "feat: add deployment script and environment editor panels"
```

---

## Task 11: Database, SSL, Workers, Daemons, Firewall, Jobs Panels

**Files:**
- Create: `internal/tui/panels/databases.go`
- Create: `internal/tui/panels/database_users.go`
- Create: `internal/tui/panels/ssl.go`
- Create: `internal/tui/panels/workers.go`
- Create: `internal/tui/panels/daemons.go`
- Create: `internal/tui/panels/firewall.go`
- Create: `internal/tui/panels/jobs.go`

All CRUD panels follow the same pattern:
1. Fetch list via API on mount
2. Render as scrollable list with key info columns
3. `c` create (shows input modal for required fields)
4. `x` or `d` delete (shows confirmation dialog)
5. Additional actions specific to each resource

**Step 1: Implement databases panel**

- List databases, `c` create (name input), `x` delete (confirm), `u` switch to users view
- Create request: `{name, user?, password?}` - show multi-field input

**Step 2: Implement database users panel**

- List users, `c` create, `x` delete, `Enter` edit database access

**Step 3: Implement SSL panel**

- List certificates, `c` create Let's Encrypt (domain input), `a` activate, `x` delete

**Step 4: Implement workers panel**

- List workers, `c` create (connection, queue, timeout fields), `r` restart, `x` delete

**Step 5: Implement daemons panel**

- List daemons, `c` create (command, user, directory), `r` restart, `x` delete

**Step 6: Implement firewall panel**

- List rules, `c` create (name, port, IP), `x` delete

**Step 7: Implement jobs panel**

- List scheduled jobs (read-only list showing command, frequency, user)

**Step 8: Wire all panels into tab system (tabs 3-9)**

**Step 9: Commit**

```bash
git commit -m "feat: add database, SSL, workers, daemons, firewall, and jobs panels"
```

---

## Task 12: Remaining Panels - SSH Keys, Commands, Logs, Git, Domains

**Files:**
- Create: `internal/tui/panels/ssh_keys.go`
- Create: `internal/tui/panels/commands.go`
- Create: `internal/tui/panels/logs.go`
- Create: `internal/tui/panels/git.go`
- Create: `internal/tui/panels/domains.go`

**Step 1: Implement SSH keys panel**

- List keys, `c` create (name + key content input), `x` delete

**Step 2: Implement commands panel**

- List site commands with status, `c` run new command (input), `Enter` view output

**Step 3: Implement logs panel**

- Fetch log content, render in scrollable viewport
- Toggle between server log and site log

**Step 4: Implement git panel**

- Show repo info (provider, URL, branch) in detail view
- Read-only info panel (no CRUD actions needed for basic use)

**Step 5: Implement domains panel**

- Show site aliases as a list
- `a` add alias (input), `x` remove alias (confirm)
- Uses `forge.Sites.UpdateAliases()` to save changes

**Step 6: Commit**

```bash
git commit -m "feat: add SSH keys, commands, logs, git, and domains panels"
```

---

## Task 13: External Tool Integrations

**Files:**
- Modify: `internal/tui/app.go`
- Create: `internal/tui/external.go`

**Step 1: Implement SSH integration**

```go
func (m App) sshToServer() tea.Cmd {
	if m.selectedSrv == nil {
		return nil
	}
	user := m.config.SSHUserFor(m.selectedSrv.Name)
	args := []string{fmt.Sprintf("%s@%s", user, m.selectedSrv.IPAddress)}

	// If a site is selected, cd to its directory
	if m.selectedSite != nil && m.selectedSite.Directory != "" {
		args = append(args, "-t", fmt.Sprintf("cd %s && exec $SHELL -l", m.selectedSite.Directory))
	}

	cmd := exec.Command("ssh", args...)
	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		return externalExitMsg{err}
	})
}
```

**Step 2: Implement SFTP via termscp**

```go
func (m App) sftpToServer() tea.Cmd {
	if m.selectedSrv == nil {
		return nil
	}
	user := m.config.SSHUserFor(m.selectedSrv.Name)
	remotePath := "/"
	if m.selectedSite != nil && m.selectedSite.Directory != "" {
		remotePath = m.selectedSite.Directory
	}
	target := fmt.Sprintf("scp://%s@%s:%s", user, m.selectedSrv.IPAddress, remotePath)
	cmd := exec.Command("termscp", target)
	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		return externalExitMsg{err}
	})
}
```

**Step 3: Implement database tunnel + lazysql**

This is the most complex integration:
1. Fetch .env from the selected site to extract DB credentials
2. Parse DB_HOST, DB_PORT, DB_DATABASE, DB_USERNAME, DB_PASSWORD from env content
3. Open SSH tunnel: `ssh -L localPort:dbHost:dbPort user@server -N` as a background process
4. Launch lazysql connecting to `localhost:localPort`
5. Kill tunnel process on exit

```go
func (m App) openDatabase() tea.Cmd {
	// Implementation will:
	// 1. Call forge.Environment.Get() to get .env content
	// 2. Parse DB_* variables
	// 3. Start SSH tunnel as background goroutine
	// 4. Wait for tunnel to be ready
	// 5. Return tea.ExecProcess for lazysql
	// 6. Kill tunnel on lazysql exit
}
```

**Step 4: Wire keybindings in app.go Update()**

Add `Ctrl+S` → sshToServer, `Ctrl+F` → sftpToServer, `Ctrl+D` → openDatabase to the global key handler.

**Step 5: Manual test each integration**

**Step 6: Commit**

```bash
git commit -m "feat: add SSH, SFTP, and database tunnel integrations"
```

---

## Task 14: Setup Flow & Help Modal

**Files:**
- Create: `internal/tui/setup.go`
- Create: `internal/tui/help.go`
- Modify: `cmd/phorge/main.go`

**Step 1: Implement setup flow**

When no API key is configured, show a setup screen:
- Text input for API key
- Validate by calling `forge.Servers.GetUser()`
- On success, save to config and proceed to main app
- On failure, show error and let user retry

This can be a separate Bubbletea model that runs before the main App model.

**Step 2: Implement help modal**

A full-screen overlay showing all keybindings organized by context:
- Global, Navigation, Server Actions, Site Actions, Section Tabs
- Toggle with `?`
- Dismiss with `esc` or `?`

**Step 3: Wire into app - check API key in main.go, show setup if missing**

**Step 4: Commit**

```bash
git commit -m "feat: add first-run setup flow and help modal"
```

---

## Task 15: Search/Filter & Polish

**Files:**
- Modify: `internal/tui/app.go`
- Modify: `internal/tui/panels/servers.go`
- Modify: `internal/tui/panels/sites.go`

**Step 1: Implement search/filter**

When `/` is pressed in a list panel:
- Show a text input at the top of the panel
- Filter the list in real-time as the user types (case-insensitive substring match)
- `Enter` accepts the filter, `Esc` clears it
- Filtered items are a subset view - cursor stays within filtered results

**Step 2: Add loading indicators**

When API calls are in progress, show a spinner in the panel title.

**Step 3: Handle terminal resize**

Ensure all panels re-render correctly when `tea.WindowSizeMsg` is received. Propagate width/height to all sub-models.

**Step 4: Edge cases**

- Empty server list: show "No servers found"
- Empty site list: show "No sites on this server"
- Network offline: show connection error with retry hint

**Step 5: Commit**

```bash
git commit -m "feat: add search/filter, loading spinners, and edge case handling"
```

---

## Task 16: Build & Distribution

**Files:**
- Create: `Makefile`
- Create: `.goreleaser.yml`

**Step 1: Create Makefile**

```makefile
.PHONY: build clean test

build:
	go build -o phorge ./cmd/phorge

test:
	go test ./... -v

clean:
	rm -f phorge phorge.exe

install:
	go install ./cmd/phorge
```

**Step 2: Create GoReleaser config for cross-compilation and GitHub releases**

```yaml
# .goreleaser.yml
version: 2
builds:
  - main: ./cmd/phorge
    binary: phorge
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64
archives:
  - formats: ['tar.gz']
    format_overrides:
      - goos: windows
        formats: ['zip']
```

**Step 3: Verify cross-compilation**

```bash
GOOS=linux GOARCH=amd64 go build -o phorge-linux ./cmd/phorge
GOOS=darwin GOARCH=arm64 go build -o phorge-darwin ./cmd/phorge
GOOS=windows GOARCH=amd64 go build -o phorge.exe ./cmd/phorge
```

**Step 4: Commit**

```bash
git commit -m "feat: add Makefile and GoReleaser config for distribution"
```

---

## Task 17: Final Integration Testing & Cleanup

**Step 1: Full manual test with real Forge account**

Test every feature against a real Forge API:
- [ ] Server list loads
- [ ] Site list loads when server selected
- [ ] Tab switching (1-9) shows correct panels
- [ ] j/k navigation in all lists
- [ ] Tab/Shift+Tab panel focus cycling
- [ ] Deploy action (d → confirm → deploy)
- [ ] Environment viewer and editor
- [ ] Database create/delete
- [ ] SSH integration (Ctrl+S)
- [ ] SFTP integration (Ctrl+F)
- [ ] Database tunnel (Ctrl+D)
- [ ] Refresh (Ctrl+R)
- [ ] Search/filter (/)
- [ ] Help modal (?)
- [ ] Setup flow (fresh config)
- [ ] All remaining panels (SSL, workers, daemons, firewall, jobs, keys, commands, logs, git, domains)

**Step 2: Fix any issues found during testing**

**Step 3: Remove old Python code or archive it**

Move the Python source to a `legacy/` directory or delete it, depending on preference.

**Step 4: Update README**

Update installation instructions to reflect the Go binary distribution.

**Step 5: Final commit**

```bash
git commit -m "feat: complete Go rewrite with full feature parity"
```
