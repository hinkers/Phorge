package forge

import (
	"context"
	"fmt"
	"net/http"
)

// List returns all servers for the authenticated user.
func (s *ServersService) List(ctx context.Context) ([]Server, error) {
	var resp struct {
		Servers []Server `json:"servers"`
	}
	err := s.client.do(ctx, http.MethodGet, "/servers", nil, &resp)
	return resp.Servers, err
}

// Get returns a single server by ID.
func (s *ServersService) Get(ctx context.Context, serverID int64) (*Server, error) {
	var resp struct {
		Server Server `json:"server"`
	}
	err := s.client.do(ctx, http.MethodGet, fmt.Sprintf("/servers/%d", serverID), nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp.Server, nil
}

// Reboot initiates a server reboot.
func (s *ServersService) Reboot(ctx context.Context, serverID int64) error {
	return s.client.do(ctx, http.MethodPost, fmt.Sprintf("/servers/%d/reboot", serverID), nil, nil)
}

// GetUser returns the authenticated Forge user.
func (s *ServersService) GetUser(ctx context.Context) (*User, error) {
	var resp struct {
		User User `json:"user"`
	}
	err := s.client.do(ctx, http.MethodGet, "/user", nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp.User, nil
}
