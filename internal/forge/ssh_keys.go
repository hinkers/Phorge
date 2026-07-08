package forge

import (
	"context"
	"net/http"
)

// List returns all SSH keys on a server.
func (s *SSHKeysService) List(ctx context.Context, serverID int64) ([]SSHKey, error) {
	return listResources[SSHKey](s.client, ctx, s.client.orgPath("/servers/%d/ssh-keys", serverID))
}

// Get returns a single SSH key by ID.
func (s *SSHKeysService) Get(ctx context.Context, serverID, keyID int64) (*SSHKey, error) {
	return getResource[SSHKey](s.client, ctx, s.client.orgPath("/servers/%d/ssh-keys/%d", serverID, keyID))
}

// Create installs a new SSH key on a server.
func (s *SSHKeysService) Create(ctx context.Context, serverID int64, name, key, username string) (*SSHKey, error) {
	body := map[string]string{
		"name":     name,
		"key":      key,
		"username": username,
	}
	return createResource[SSHKey](s.client, ctx, s.client.orgPath("/servers/%d/ssh-keys", serverID), body)
}

// Delete removes an SSH key from a server.
func (s *SSHKeysService) Delete(ctx context.Context, serverID, keyID int64) error {
	return s.client.do(ctx, http.MethodDelete, s.client.orgPath("/servers/%d/ssh-keys/%d", serverID, keyID), nil, nil)
}
