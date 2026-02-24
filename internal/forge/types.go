package forge

// User represents a Forge user account.
type User struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email,omitempty"`
}

// Server represents a provisioned server managed by Forge.
type Server struct {
	ID               int64  `json:"id"`
	Name             string `json:"name"`
	IPAddress        string `json:"ip_address,omitempty"`
	PrivateIPAddress string `json:"private_ip_address,omitempty"`
	Region           string `json:"region,omitempty"`
	PHPVersion       string `json:"php_version,omitempty"`
	PHPCLIVersion    string `json:"php_cli_version,omitempty"`
	Provider         string `json:"provider,omitempty"`
	Type             string `json:"type,omitempty"`
	Status           string `json:"status,omitempty"`
	IsReady          bool   `json:"is_ready"`
	DatabaseType     string `json:"database_type,omitempty"`
	SSHPort          int    `json:"ssh_port,omitempty"`
	UbuntuVersion    string `json:"ubuntu_version,omitempty"`
	DBStatus         string `json:"db_status,omitempty"`
	RedisStatus      string `json:"redis_status,omitempty"`
	Network          []any  `json:"network,omitempty"`
	Tags             []any  `json:"tags,omitempty"`
}

// Site represents a website/application hosted on a server.
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

// Deployment represents a site deployment event.
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

// Database represents a database on a server.
type Database struct {
	ID       int64  `json:"id"`
	ServerID int64  `json:"server_id,omitempty"`
	Name     string `json:"name"`
	Status   string `json:"status,omitempty"`
	IsSynced bool   `json:"is_synced"`
}

// DatabaseUser represents a database user on a server.
type DatabaseUser struct {
	ID        int64   `json:"id"`
	ServerID  int64   `json:"server_id,omitempty"`
	Name      string  `json:"name"`
	Status    string  `json:"status,omitempty"`
	Databases []int64 `json:"databases,omitempty"`
}

// SSHKey represents an SSH key installed on a server.
type SSHKey struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status,omitempty"`
}

// Daemon represents a daemon (supervisor) process on a server.
type Daemon struct {
	ID        int64  `json:"id"`
	ServerID  int64  `json:"server_id,omitempty"`
	Command   string `json:"command"`
	User      string `json:"user,omitempty"`
	Directory string `json:"directory,omitempty"`
	Processes int    `json:"processes,omitempty"`
	StartSecs int    `json:"startsecs,omitempty"`
	Status    string `json:"status,omitempty"`
}

// FirewallRule represents a firewall rule on a server.
type FirewallRule struct {
	ID        int64  `json:"id"`
	ServerID  int64  `json:"server_id,omitempty"`
	Name      string `json:"name"`
	Port      any    `json:"port,omitempty"`
	IPAddress string `json:"ip_address,omitempty"`
	Type      string `json:"type,omitempty"`
	Status    string `json:"status,omitempty"`
}

// ScheduledJob represents a cron job on a server.
type ScheduledJob struct {
	ID        int64  `json:"id"`
	ServerID  int64  `json:"server_id,omitempty"`
	Command   string `json:"command"`
	User      string `json:"user,omitempty"`
	Frequency string `json:"frequency,omitempty"`
	Cron      string `json:"cron,omitempty"`
	Status    string `json:"status,omitempty"`
}

// Worker represents a queue worker on a site.
type Worker struct {
	ID         int64  `json:"id"`
	Connection string `json:"connection,omitempty"`
	Queue      string `json:"queue,omitempty"`
	Timeout    int    `json:"timeout,omitempty"`
	Sleep      int    `json:"sleep,omitempty"`
	Processes  int    `json:"processes,omitempty"`
	DaemonMode bool   `json:"daemon"`
	Force      bool   `json:"force"`
	Status     string `json:"status,omitempty"`
}

// Certificate represents an SSL certificate on a site.
type Certificate struct {
	ID       int64  `json:"id"`
	Domain   string `json:"domain,omitempty"`
	Type     string `json:"type,omitempty"`
	Active   bool   `json:"active"`
	Status   string `json:"status,omitempty"`
	Existing bool   `json:"existing"`
}

// Backup represents a single backup snapshot.
type Backup struct {
	ID                    int64  `json:"id"`
	BackupConfigurationID int64  `json:"backup_configuration_id"`
	Status                string `json:"status,omitempty"`
	Date                  string `json:"date,omitempty"`
	Size                  any    `json:"size,omitempty"`
	Duration              any    `json:"duration,omitempty"`
}

// BackupConfig represents a backup configuration on a server.
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

// SiteCommand represents a command that was executed on a site.
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

// RedirectRule represents a redirect rule on a site.
type RedirectRule struct {
	ID     int64  `json:"id"`
	From   string `json:"from"`
	To     string `json:"to,omitempty"`
	Type   string `json:"type,omitempty"`
	Status string `json:"status,omitempty"`
}
