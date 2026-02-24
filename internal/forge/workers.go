package forge

import (
	"context"
	"fmt"
	"net/http"
)

// WorkerCreateOpts contains the options for creating a queue worker.
type WorkerCreateOpts struct {
	Connection string `json:"connection"`            // default "redis"
	Queue      string `json:"queue"`                 // default "default"
	Timeout    int    `json:"timeout"`               // default 60
	Sleep      int    `json:"sleep"`                 // default 3
	Processes  int    `json:"processes"`             // default 1
	Daemon     bool   `json:"daemon"`                // default true
	Force      bool   `json:"force"`                 // default false
	PHPVersion string `json:"php_version,omitempty"` // optional
}

// List returns all queue workers for a site.
func (s *WorkersService) List(ctx context.Context, serverID, siteID int64) ([]Worker, error) {
	var resp struct {
		Workers []Worker `json:"workers"`
	}
	path := fmt.Sprintf("/servers/%d/sites/%d/workers", serverID, siteID)
	err := s.client.do(ctx, http.MethodGet, path, nil, &resp)
	return resp.Workers, err
}

// Get returns a single worker by ID.
func (s *WorkersService) Get(ctx context.Context, serverID, siteID, workerID int64) (*Worker, error) {
	var resp struct {
		Worker Worker `json:"worker"`
	}
	path := fmt.Sprintf("/servers/%d/sites/%d/workers/%d", serverID, siteID, workerID)
	err := s.client.do(ctx, http.MethodGet, path, nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp.Worker, nil
}

// Create creates a new queue worker on a site.
func (s *WorkersService) Create(ctx context.Context, serverID, siteID int64, opts WorkerCreateOpts) (*Worker, error) {
	var resp struct {
		Worker Worker `json:"worker"`
	}
	path := fmt.Sprintf("/servers/%d/sites/%d/workers", serverID, siteID)
	err := s.client.do(ctx, http.MethodPost, path, opts, &resp)
	if err != nil {
		return nil, err
	}
	return &resp.Worker, nil
}

// Restart restarts a queue worker.
func (s *WorkersService) Restart(ctx context.Context, serverID, siteID, workerID int64) error {
	path := fmt.Sprintf("/servers/%d/sites/%d/workers/%d/restart", serverID, siteID, workerID)
	return s.client.do(ctx, http.MethodPost, path, nil, nil)
}

// Delete removes a queue worker.
func (s *WorkersService) Delete(ctx context.Context, serverID, siteID, workerID int64) error {
	path := fmt.Sprintf("/servers/%d/sites/%d/workers/%d", serverID, siteID, workerID)
	return s.client.do(ctx, http.MethodDelete, path, nil, nil)
}
