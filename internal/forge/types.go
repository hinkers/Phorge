package forge

// User represents a Forge user account.
type User struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email,omitempty"`
}

// Organization represents a Forge organization/team.
type Organization struct {
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

// Server represents a provisioned server managed by Forge.
type Server struct {
	ID               int64  `json:"id"`
	CredentialID     *int64 `json:"credential_id,omitempty"`
	Name             string `json:"name"`
	Slug             string `json:"slug,omitempty"`
	Type             string `json:"type,omitempty"`
	UbuntuVersion    string `json:"ubuntu_version,omitempty"`
	SSHPort          int    `json:"ssh_port,omitempty"`
	Provider         string `json:"provider,omitempty"`
	Identifier       string `json:"identifier,omitempty"`
	Size             string `json:"size,omitempty"`
	Region           string `json:"region,omitempty"`
	PHPVersion       string `json:"php_version,omitempty"`
	PHPCLIVersion    string `json:"php_cli_version,omitempty"`
	OpcacheStatus    string `json:"opcache_status,omitempty"`
	DatabaseType     string `json:"database_type,omitempty"`
	DBStatus         string `json:"db_status,omitempty"`
	RedisStatus      string `json:"redis_status,omitempty"`
	IPAddress        string `json:"ip_address,omitempty"`
	PrivateIPAddress string `json:"private_ip_address,omitempty"`
	Revoked          bool   `json:"revoked"`
	IsReady          bool   `json:"is_ready"`
	ConnectionStatus string `json:"connection_status,omitempty"`
	Timezone         string `json:"timezone,omitempty"`
	LocalPublicKey   string `json:"local_public_key,omitempty"`
	CreatedAt        string `json:"created_at,omitempty"`
	UpdatedAt        string `json:"updated_at,omitempty"`
}

// SiteRepository represents the source-control repository backing a site.
type SiteRepository struct {
	Provider string `json:"provider,omitempty"`
	URL      string `json:"url,omitempty"`
	Branch   string `json:"branch,omitempty"`
	Status   string `json:"status,omitempty"`
}

// MaintenanceMode represents a site's maintenance mode state.
type MaintenanceMode struct {
	Enabled bool   `json:"enabled"`
	Status  string `json:"status,omitempty"`
}

// Site represents a website/application hosted on a server.
type Site struct {
	ID                      int64            `json:"id,omitempty"`
	Name                    string           `json:"name"`
	Status                  string           `json:"status,omitempty"`
	URL                     string           `json:"url,omitempty"`
	User                    string           `json:"user,omitempty"`
	HTTPS                   bool             `json:"https"`
	Directory               string           `json:"directory,omitempty"`
	WebDirectory            string           `json:"web_directory,omitempty"`
	RootDirectory           string           `json:"root_directory,omitempty"`
	Aliases                 []string         `json:"aliases,omitempty"`
	PHPVersion              string           `json:"php_version,omitempty"`
	DeploymentStatus        string           `json:"deployment_status,omitempty"`
	QuickDeploy             *bool            `json:"quick_deploy,omitempty"`
	Isolated                bool             `json:"isolated"`
	Repository              *SiteRepository  `json:"repository,omitempty"`
	Database                string           `json:"database,omitempty"`
	MaintenanceMode         *MaintenanceMode `json:"maintenance_mode,omitempty"`
	ZeroDowntimeDeployments bool             `json:"zero_downtime_deployments"`
	DeploymentScript        string           `json:"deployment_script,omitempty"`
	Wildcards               *bool            `json:"wildcards,omitempty"`
	AppType                 string           `json:"app_type,omitempty"`
	UsesEnvoyer             bool             `json:"uses_envoyer"`
	DeploymentURL           string           `json:"deployment_url,omitempty"`
	HealthcheckURL          string           `json:"healthcheck_url,omitempty"`
	IsSecured               bool             `json:"is_secured"`
	CreatedAt               string           `json:"created_at,omitempty"`
	UpdatedAt               string           `json:"updated_at,omitempty"`
}

// DeploymentCommit represents the commit associated with a deployment.
type DeploymentCommit struct {
	Hash    string `json:"hash,omitempty"`
	Author  string `json:"author,omitempty"`
	Message string `json:"message,omitempty"`
	Branch  string `json:"branch,omitempty"`
}

// Deployment represents a site deployment event.
type Deployment struct {
	ID        int64            `json:"id,omitempty"`
	Commit    DeploymentCommit `json:"commit"`
	Type      string           `json:"type,omitempty"`
	Status    string           `json:"status,omitempty"`
	StartedAt string           `json:"started_at,omitempty"`
	EndedAt   string           `json:"ended_at,omitempty"`
	CreatedAt string           `json:"created_at,omitempty"`
	UpdatedAt string           `json:"updated_at,omitempty"`
}

// Database represents a database on a server.
type Database struct {
	ID        int64  `json:"id,omitempty"`
	Name      string `json:"name"`
	Status    string `json:"status,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

// DatabaseUser represents a database user on a server.
type DatabaseUser struct {
	ID        int64  `json:"id,omitempty"`
	Name      string `json:"name"`
	Status    string `json:"status,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

// SSHKey represents an SSH key installed on a server.
type SSHKey struct {
	ID        int64  `json:"id,omitempty"`
	Name      string `json:"name"`
	User      string `json:"user,omitempty"`
	Status    string `json:"status,omitempty"`
	CreatedBy *int64 `json:"created_by,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

// BackgroundProcess represents a daemon (supervisor) process on a server.
type BackgroundProcess struct {
	ID        int64  `json:"id,omitempty"`
	Command   string `json:"command"`
	User      string `json:"user,omitempty"`
	Directory string `json:"directory,omitempty"`
	Processes int    `json:"processes,omitempty"`
	Status    string `json:"status,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
}

// FirewallRule represents a firewall rule on a server.
type FirewallRule struct {
	ID        int64  `json:"id,omitempty"`
	Name      string `json:"name"`
	Port      string `json:"port,omitempty"`
	IPAddress string `json:"ip_address,omitempty"`
	Type      string `json:"type,omitempty"`
	Status    string `json:"status,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

// ScheduledJob represents a cron job on a server.
type ScheduledJob struct {
	ID          int64  `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Command     string `json:"command"`
	User        string `json:"user,omitempty"`
	Frequency   string `json:"frequency,omitempty"`
	Cron        string `json:"cron,omitempty"`
	NextRunTime string `json:"next_run_time,omitempty"`
	Status      string `json:"status,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
	UpdatedAt   string `json:"updated_at,omitempty"`
}

// Certificate represents an SSL certificate on a site.
type Certificate struct {
	ID                 int64  `json:"id,omitempty"`
	Type               string `json:"type,omitempty"`
	VerificationMethod string `json:"verification_method,omitempty"`
	KeyType            string `json:"key_type,omitempty"`
	PreferredChain     string `json:"preferred_chain,omitempty"`
	RequestStatus      string `json:"request_status,omitempty"`
	Status             string `json:"status,omitempty"`
	Active             bool   `json:"active"`
	CreatedAt          string `json:"created_at,omitempty"`
	UpdatedAt          string `json:"updated_at,omitempty"`
}

// Backup represents a single backup snapshot.
type Backup struct {
	ID         int64  `json:"id,omitempty"`
	Status     string `json:"status,omitempty"`
	IsPartial  string `json:"is_partial,omitempty"`
	Size       int    `json:"size,omitempty"`
	FinishedAt string `json:"finished_at,omitempty"`
}

// BackupConfig represents a backup configuration on a server.
type BackupConfig struct {
	ID                  int64   `json:"id,omitempty"`
	Name                string  `json:"name,omitempty"`
	StorageProviderID   *int64  `json:"storage_provider_id,omitempty"`
	Provider            string  `json:"provider,omitempty"`
	Bucket              string  `json:"bucket,omitempty"`
	Directory           string  `json:"directory,omitempty"`
	Schedule            string  `json:"schedule,omitempty"`
	DisplayableSchedule string  `json:"displayable_schedule,omitempty"`
	NextRunTime         string  `json:"next_run_time,omitempty"`
	Status              string  `json:"status,omitempty"`
	DayOfWeek           *int    `json:"day_of_week,omitempty"`
	Time                string  `json:"time,omitempty"`
	CronSchedule        string  `json:"cron_schedule,omitempty"`
	DatabaseIDs         []int64 `json:"database_ids,omitempty"`
	Retention           int     `json:"retention,omitempty"`
	NotifyEmail         string  `json:"notify_email,omitempty"`
}

// Command represents a command that was executed on a site.
type Command struct {
	ID          int64  `json:"id,omitempty"`
	UserID      int64  `json:"user_id,omitempty"`
	Command     string `json:"command"`
	Status      string `json:"status,omitempty"`
	ExitCode    *int   `json:"exit_code,omitempty"`
	ErrorOutput string `json:"error_output,omitempty"`
	Duration    string `json:"duration,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
	UpdatedAt   string `json:"updated_at,omitempty"`
}

// Event represents a server activity event (e.g. deployment, reboot).
type Event struct {
	ID          int64  `json:"id,omitempty"`
	Description string `json:"description,omitempty"`
	RanAs       string `json:"ran_as,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
	UpdatedAt   string `json:"updated_at,omitempty"`
}

// RedirectRule represents a redirect rule on a site.
type RedirectRule struct {
	ID        int64  `json:"id,omitempty"`
	From      string `json:"from"`
	To        string `json:"to,omitempty"`
	Type      string `json:"type,omitempty"`
	Status    string `json:"status,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}
