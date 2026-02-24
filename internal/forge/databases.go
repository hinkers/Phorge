package forge

import (
	"context"
	"fmt"
	"net/http"
)

// List returns all databases on a server.
func (s *DatabasesService) List(ctx context.Context, serverID int64) ([]Database, error) {
	var resp struct {
		Databases []Database `json:"databases"`
	}
	path := fmt.Sprintf("/servers/%d/databases", serverID)
	err := s.client.do(ctx, http.MethodGet, path, nil, &resp)
	return resp.Databases, err
}

// Get returns a single database by ID.
func (s *DatabasesService) Get(ctx context.Context, serverID, dbID int64) (*Database, error) {
	var resp struct {
		Database Database `json:"database"`
	}
	path := fmt.Sprintf("/servers/%d/databases/%d", serverID, dbID)
	err := s.client.do(ctx, http.MethodGet, path, nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp.Database, nil
}

// Create creates a new database on a server.
// The user and password parameters are optional; pass nil to omit them.
func (s *DatabasesService) Create(ctx context.Context, serverID int64, name string, user, password *string) (*Database, error) {
	body := map[string]any{"name": name}
	if user != nil {
		body["user"] = *user
	}
	if password != nil {
		body["password"] = *password
	}

	var resp struct {
		Database Database `json:"database"`
	}
	path := fmt.Sprintf("/servers/%d/databases", serverID)
	err := s.client.do(ctx, http.MethodPost, path, body, &resp)
	if err != nil {
		return nil, err
	}
	return &resp.Database, nil
}

// Delete removes a database from a server.
func (s *DatabasesService) Delete(ctx context.Context, serverID, dbID int64) error {
	path := fmt.Sprintf("/servers/%d/databases/%d", serverID, dbID)
	return s.client.do(ctx, http.MethodDelete, path, nil, nil)
}

// Sync triggers a database sync on the server.
func (s *DatabasesService) Sync(ctx context.Context, serverID int64) error {
	path := fmt.Sprintf("/servers/%d/databases/sync", serverID)
	return s.client.do(ctx, http.MethodPost, path, nil, nil)
}

// ListUsers returns all database users on a server.
func (s *DatabasesService) ListUsers(ctx context.Context, serverID int64) ([]DatabaseUser, error) {
	var resp struct {
		Users []DatabaseUser `json:"users"`
	}
	path := fmt.Sprintf("/servers/%d/database-users", serverID)
	err := s.client.do(ctx, http.MethodGet, path, nil, &resp)
	return resp.Users, err
}

// GetUser returns a single database user by ID.
func (s *DatabasesService) GetUser(ctx context.Context, serverID, userID int64) (*DatabaseUser, error) {
	var resp struct {
		User DatabaseUser `json:"user"`
	}
	path := fmt.Sprintf("/servers/%d/database-users/%d", serverID, userID)
	err := s.client.do(ctx, http.MethodGet, path, nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp.User, nil
}

// CreateUser creates a new database user on a server.
func (s *DatabasesService) CreateUser(ctx context.Context, serverID int64, name, password string, databases []int64) (*DatabaseUser, error) {
	body := map[string]any{
		"name":      name,
		"password":  password,
		"databases": databases,
	}
	var resp struct {
		User DatabaseUser `json:"user"`
	}
	path := fmt.Sprintf("/servers/%d/database-users", serverID)
	err := s.client.do(ctx, http.MethodPost, path, body, &resp)
	if err != nil {
		return nil, err
	}
	return &resp.User, nil
}

// UpdateUser updates the database access for a database user.
func (s *DatabasesService) UpdateUser(ctx context.Context, serverID, userID int64, databases []int64) (*DatabaseUser, error) {
	body := map[string]any{"databases": databases}
	var resp struct {
		User DatabaseUser `json:"user"`
	}
	path := fmt.Sprintf("/servers/%d/database-users/%d", serverID, userID)
	err := s.client.do(ctx, http.MethodPut, path, body, &resp)
	if err != nil {
		return nil, err
	}
	return &resp.User, nil
}

// DeleteUser removes a database user from a server.
func (s *DatabasesService) DeleteUser(ctx context.Context, serverID, userID int64) error {
	path := fmt.Sprintf("/servers/%d/database-users/%d", serverID, userID)
	return s.client.do(ctx, http.MethodDelete, path, nil, nil)
}
