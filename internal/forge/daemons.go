package forge

import (
	"context"
	"fmt"
	"net/http"
)

// DaemonCreateOpts contains the options for creating a daemon.
type DaemonCreateOpts struct {
	Command   string `json:"command"`
	User      string `json:"user"`                // default "forge"
	Directory string `json:"directory,omitempty"`  // optional
	Processes int    `json:"processes"`            // default 1
	StartSecs int    `json:"startsecs"`            // default 1
}

// List returns all daemons on a server.
func (s *DaemonsService) List(ctx context.Context, serverID int64) ([]Daemon, error) {
	var resp struct {
		Daemons []Daemon `json:"daemons"`
	}
	path := fmt.Sprintf("/servers/%d/daemons", serverID)
	err := s.client.do(ctx, http.MethodGet, path, nil, &resp)
	return resp.Daemons, err
}

// Get returns a single daemon by ID.
func (s *DaemonsService) Get(ctx context.Context, serverID, daemonID int64) (*Daemon, error) {
	var resp struct {
		Daemon Daemon `json:"daemon"`
	}
	path := fmt.Sprintf("/servers/%d/daemons/%d", serverID, daemonID)
	err := s.client.do(ctx, http.MethodGet, path, nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp.Daemon, nil
}

// Create creates a new daemon on a server.
func (s *DaemonsService) Create(ctx context.Context, serverID int64, opts DaemonCreateOpts) (*Daemon, error) {
	var resp struct {
		Daemon Daemon `json:"daemon"`
	}
	path := fmt.Sprintf("/servers/%d/daemons", serverID)
	err := s.client.do(ctx, http.MethodPost, path, opts, &resp)
	if err != nil {
		return nil, err
	}
	return &resp.Daemon, nil
}

// Restart restarts a daemon.
func (s *DaemonsService) Restart(ctx context.Context, serverID, daemonID int64) error {
	path := fmt.Sprintf("/servers/%d/daemons/%d/restart", serverID, daemonID)
	return s.client.do(ctx, http.MethodPost, path, nil, nil)
}

// Delete removes a daemon from a server.
func (s *DaemonsService) Delete(ctx context.Context, serverID, daemonID int64) error {
	path := fmt.Sprintf("/servers/%d/daemons/%d", serverID, daemonID)
	return s.client.do(ctx, http.MethodDelete, path, nil, nil)
}
