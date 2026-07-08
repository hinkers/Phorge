package forge

import (
	"context"
	"net/http"
)

// List returns all databases on a server.
func (s *DatabasesService) List(ctx context.Context, serverID int64) ([]Database, error) {
	return listResources[Database](s.client, ctx, s.client.orgPath("/servers/%d/database/schemas", serverID))
}

// Get returns a single database by ID.
func (s *DatabasesService) Get(ctx context.Context, serverID, dbID int64) (*Database, error) {
	return getResource[Database](s.client, ctx, s.client.orgPath("/servers/%d/database/schemas/%d", serverID, dbID))
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
	return createResource[Database](s.client, ctx, s.client.orgPath("/servers/%d/database/schemas", serverID), body)
}

// Delete removes a database from a server.
func (s *DatabasesService) Delete(ctx context.Context, serverID, dbID int64) error {
	return s.client.do(ctx, http.MethodDelete, s.client.orgPath("/servers/%d/database/schemas/%d", serverID, dbID), nil, nil)
}

// Sync triggers a database sync on the server.
func (s *DatabasesService) Sync(ctx context.Context, serverID int64) error {
	return s.client.do(ctx, http.MethodPost, s.client.orgPath("/servers/%d/database/schemas/synchronizations", serverID), nil, nil)
}

// ListUsers returns all database users on a server.
func (s *DatabasesService) ListUsers(ctx context.Context, serverID int64) ([]DatabaseUser, error) {
	return listResources[DatabaseUser](s.client, ctx, s.client.orgPath("/servers/%d/database/users", serverID))
}

// GetUser returns a single database user by ID.
func (s *DatabasesService) GetUser(ctx context.Context, serverID, userID int64) (*DatabaseUser, error) {
	return getResource[DatabaseUser](s.client, ctx, s.client.orgPath("/servers/%d/database/users/%d", serverID, userID))
}

// CreateUser creates a new database user on a server.
func (s *DatabasesService) CreateUser(ctx context.Context, serverID int64, name, password string, databases []int64) (*DatabaseUser, error) {
	body := map[string]any{
		"name":     name,
		"password": password,
	}
	if databases != nil {
		body["databases"] = databases
	}
	return createResource[DatabaseUser](s.client, ctx, s.client.orgPath("/servers/%d/database/users", serverID), body)
}

// UpdateUser updates the database access for a database user.
func (s *DatabasesService) UpdateUser(ctx context.Context, serverID, userID int64, databases []int64) (*DatabaseUser, error) {
	path := s.client.orgPath("/servers/%d/database/users/%d", serverID, userID)
	body := map[string]any{"databases": databases}
	return updateResource[DatabaseUser](s.client, ctx, path, body)
}

// DeleteUser removes a database user from a server.
func (s *DatabasesService) DeleteUser(ctx context.Context, serverID, userID int64) error {
	return s.client.do(ctx, http.MethodDelete, s.client.orgPath("/servers/%d/database/users/%d", serverID, userID), nil, nil)
}
