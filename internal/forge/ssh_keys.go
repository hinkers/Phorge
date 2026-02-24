package forge

import (
	"context"
	"fmt"
	"net/http"
)

// List returns all SSH keys on a server.
func (s *SSHKeysService) List(ctx context.Context, serverID int64) ([]SSHKey, error) {
	var resp struct {
		Keys []SSHKey `json:"keys"`
	}
	path := fmt.Sprintf("/servers/%d/keys", serverID)
	err := s.client.do(ctx, http.MethodGet, path, nil, &resp)
	return resp.Keys, err
}

// Get returns a single SSH key by ID.
func (s *SSHKeysService) Get(ctx context.Context, serverID, keyID int64) (*SSHKey, error) {
	var resp struct {
		Key SSHKey `json:"key"`
	}
	path := fmt.Sprintf("/servers/%d/keys/%d", serverID, keyID)
	err := s.client.do(ctx, http.MethodGet, path, nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp.Key, nil
}

// Create installs a new SSH key on a server.
func (s *SSHKeysService) Create(ctx context.Context, serverID int64, name, key, username string) (*SSHKey, error) {
	body := map[string]string{
		"name":     name,
		"key":      key,
		"username": username,
	}
	var resp struct {
		Key SSHKey `json:"key"`
	}
	path := fmt.Sprintf("/servers/%d/keys", serverID)
	err := s.client.do(ctx, http.MethodPost, path, body, &resp)
	if err != nil {
		return nil, err
	}
	return &resp.Key, nil
}

// Delete removes an SSH key from a server.
func (s *SSHKeysService) Delete(ctx context.Context, serverID, keyID int64) error {
	path := fmt.Sprintf("/servers/%d/keys/%d", serverID, keyID)
	return s.client.do(ctx, http.MethodDelete, path, nil, nil)
}
