package forge

import (
	"context"
	"fmt"
	"net/http"
)

// List returns all commands that have been executed on a site.
func (s *CommandsService) List(ctx context.Context, serverID, siteID int64) ([]SiteCommand, error) {
	var resp struct {
		Commands []SiteCommand `json:"commands"`
	}
	path := fmt.Sprintf("/servers/%d/sites/%d/commands", serverID, siteID)
	err := s.client.do(ctx, http.MethodGet, path, nil, &resp)
	return resp.Commands, err
}

// Get returns a single site command by ID.
func (s *CommandsService) Get(ctx context.Context, serverID, siteID, cmdID int64) (*SiteCommand, error) {
	var resp struct {
		Command SiteCommand `json:"command"`
	}
	path := fmt.Sprintf("/servers/%d/sites/%d/commands/%d", serverID, siteID, cmdID)
	err := s.client.do(ctx, http.MethodGet, path, nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp.Command, nil
}

// Create executes a new command on a site.
func (s *CommandsService) Create(ctx context.Context, serverID, siteID int64, command string) (*SiteCommand, error) {
	body := map[string]string{"command": command}
	var resp struct {
		Command SiteCommand `json:"command"`
	}
	path := fmt.Sprintf("/servers/%d/sites/%d/commands", serverID, siteID)
	err := s.client.do(ctx, http.MethodPost, path, body, &resp)
	if err != nil {
		return nil, err
	}
	return &resp.Command, nil
}
