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
	Frequency   string         `json:"frequency"`              // default "daily"
	Databases   []int64        `json:"databases,omitempty"`
	Time        string         `json:"time,omitempty"`
	DayOfWeek   *int           `json:"day_of_week,omitempty"`
}

// ListConfigs returns all backup configurations on a server.
func (s *BackupsService) ListConfigs(ctx context.Context, serverID int64) ([]BackupConfig, error) {
	var resp struct {
		Backups []BackupConfig `json:"backups"`
	}
	path := fmt.Sprintf("/servers/%d/backup-configs", serverID)
	err := s.client.do(ctx, http.MethodGet, path, nil, &resp)
	return resp.Backups, err
}

// GetConfig returns a single backup configuration by ID.
func (s *BackupsService) GetConfig(ctx context.Context, serverID, configID int64) (*BackupConfig, error) {
	var resp struct {
		Backup BackupConfig `json:"backup"`
	}
	path := fmt.Sprintf("/servers/%d/backup-configs/%d", serverID, configID)
	err := s.client.do(ctx, http.MethodGet, path, nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp.Backup, nil
}

// CreateConfig creates a new backup configuration on a server.
func (s *BackupsService) CreateConfig(ctx context.Context, serverID int64, opts BackupConfigCreateOpts) (*BackupConfig, error) {
	var resp struct {
		Backup BackupConfig `json:"backup"`
	}
	path := fmt.Sprintf("/servers/%d/backup-configs", serverID)
	err := s.client.do(ctx, http.MethodPost, path, opts, &resp)
	if err != nil {
		return nil, err
	}
	return &resp.Backup, nil
}

// DeleteConfig removes a backup configuration from a server.
func (s *BackupsService) DeleteConfig(ctx context.Context, serverID, configID int64) error {
	path := fmt.Sprintf("/servers/%d/backup-configs/%d", serverID, configID)
	return s.client.do(ctx, http.MethodDelete, path, nil, nil)
}

// RunBackup triggers a backup for a configuration.
func (s *BackupsService) RunBackup(ctx context.Context, serverID, configID int64) error {
	path := fmt.Sprintf("/servers/%d/backup-configs/%d", serverID, configID)
	return s.client.do(ctx, http.MethodPost, path, nil, nil)
}

// RestoreBackup restores a specific backup.
func (s *BackupsService) RestoreBackup(ctx context.Context, serverID, configID, backupID int64) error {
	path := fmt.Sprintf("/servers/%d/backup-configs/%d/backups/%d", serverID, configID, backupID)
	return s.client.do(ctx, http.MethodPost, path, nil, nil)
}

// DeleteBackup removes a specific backup.
func (s *BackupsService) DeleteBackup(ctx context.Context, serverID, configID, backupID int64) error {
	path := fmt.Sprintf("/servers/%d/backup-configs/%d/backups/%d", serverID, configID, backupID)
	return s.client.do(ctx, http.MethodDelete, path, nil, nil)
}
