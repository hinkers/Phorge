package forge

import (
	"context"
	"net/http"
)

// DaemonCreateOpts contains the options for creating a daemon.
type DaemonCreateOpts struct {
	Command   string `json:"command"`
	User      string `json:"user"`                // default "forge"
	Directory string `json:"directory,omitempty"` // optional
	Processes int    `json:"processes"`           // default 1
}

// List returns all background processes on a server.
func (s *DaemonsService) List(ctx context.Context, serverID int64) ([]BackgroundProcess, error) {
	return listResources[BackgroundProcess](s.client, ctx, s.client.orgPath("/servers/%d/background-processes", serverID))
}

// Get returns a single background process by ID.
func (s *DaemonsService) Get(ctx context.Context, serverID, daemonID int64) (*BackgroundProcess, error) {
	return getResource[BackgroundProcess](s.client, ctx, s.client.orgPath("/servers/%d/background-processes/%d", serverID, daemonID))
}

// Create creates a new background process on a server.
func (s *DaemonsService) Create(ctx context.Context, serverID int64, opts DaemonCreateOpts) (*BackgroundProcess, error) {
	return createResource[BackgroundProcess](s.client, ctx, s.client.orgPath("/servers/%d/background-processes", serverID), opts)
}

// Restart restarts a background process.
func (s *DaemonsService) Restart(ctx context.Context, serverID, daemonID int64) error {
	path := s.client.orgPath("/servers/%d/background-processes/%d/actions", serverID, daemonID)
	body := map[string]string{"action": "restart"}
	return s.client.do(ctx, http.MethodPost, path, body, nil)
}

// Delete removes a background process from a server.
func (s *DaemonsService) Delete(ctx context.Context, serverID, daemonID int64) error {
	return s.client.do(ctx, http.MethodDelete, s.client.orgPath("/servers/%d/background-processes/%d", serverID, daemonID), nil, nil)
}
