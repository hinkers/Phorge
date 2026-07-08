package forge

import (
	"context"
	"fmt"
	"net/http"
)

// BackupConfigCreateOpts contains the options for creating a backup configuration.
type BackupConfigCreateOpts struct {
	Provider    string         `json:"provider"`
	Credentials map[string]any `json:"credentials"`
	Frequency   string         `json:"frequency"` // default "daily"
	Databases   []int64        `json:"databases,omitempty"`
	Time        string         `json:"time,omitempty"`
	DayOfWeek   *int           `json:"day_of_week,omitempty"`
}

func (s *BackupsService) dbPath(serverID int64, suffix string) string {
	return s.client.orgPath("/servers/%d/database/backups%s", serverID, suffix)
}

// ListConfigs returns all backup configurations on a server.
func (s *BackupsService) ListConfigs(ctx context.Context, serverID int64) ([]BackupConfig, error) {
	return listResources[BackupConfig](s.client, ctx, s.dbPath(serverID, ""))
}

// GetConfig returns a single backup configuration by ID.
func (s *BackupsService) GetConfig(ctx context.Context, serverID, configID int64) (*BackupConfig, error) {
	return getResource[BackupConfig](s.client, ctx, s.dbPath(serverID, fmt.Sprintf("/%d", configID)))
}

// CreateConfig creates a new backup configuration on a server.
func (s *BackupsService) CreateConfig(ctx context.Context, serverID int64, opts BackupConfigCreateOpts) (*BackupConfig, error) {
	return createResource[BackupConfig](s.client, ctx, s.dbPath(serverID, ""), opts)
}

// DeleteConfig removes a backup configuration from a server.
func (s *BackupsService) DeleteConfig(ctx context.Context, serverID, configID int64) error {
	return s.client.do(ctx, http.MethodDelete, s.dbPath(serverID, fmt.Sprintf("/%d", configID)), nil, nil)
}

// RunBackup triggers a backup for a configuration.
func (s *BackupsService) RunBackup(ctx context.Context, serverID, configID int64) error {
	return s.client.do(ctx, http.MethodPost, s.dbPath(serverID, fmt.Sprintf("/%d/instances", configID)), nil, nil)
}

// RestoreBackup restores a specific backup.
func (s *BackupsService) RestoreBackup(ctx context.Context, serverID, configID, backupID int64) error {
	path := s.dbPath(serverID, fmt.Sprintf("/%d/instances/%d/restores", configID, backupID))
	return s.client.do(ctx, http.MethodPost, path, nil, nil)
}

// DeleteBackup removes a specific backup.
func (s *BackupsService) DeleteBackup(ctx context.Context, serverID, configID, backupID int64) error {
	path := s.dbPath(serverID, fmt.Sprintf("/%d/instances/%d", configID, backupID))
	return s.client.do(ctx, http.MethodDelete, path, nil, nil)
}
