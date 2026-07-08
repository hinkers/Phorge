package forge

import (
	"context"
	"net/http"
)

// List returns all servers for the authenticated organization.
func (s *ServersService) List(ctx context.Context) ([]Server, error) {
	return listResources[Server](s.client, ctx, s.client.orgPath("/servers"))
}

// Get returns a single server by ID.
func (s *ServersService) Get(ctx context.Context, serverID int64) (*Server, error) {
	return getResource[Server](s.client, ctx, s.client.orgPath("/servers/%d", serverID))
}

// Reboot initiates a server reboot.
func (s *ServersService) Reboot(ctx context.Context, serverID int64) error {
	path := s.client.orgPath("/servers/%d/actions", serverID)
	body := map[string]string{"action": "reboot"}
	return s.client.do(ctx, http.MethodPost, path, body, nil)
}

// GetUser returns the authenticated Forge user.
func (s *ServersService) GetUser(ctx context.Context) (*User, error) {
	return getResource[User](s.client, ctx, "/me")
}
