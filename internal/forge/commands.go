package forge

import (
	"context"
	"fmt"
)

// sitePath builds the org-scoped commands collection path for a site.
func (s *CommandsService) sitePath(serverID, siteID int64) string {
	return s.client.orgPath("/servers/%d/sites/%d/commands", serverID, siteID)
}

// List returns all commands that have been executed on a site.
func (s *CommandsService) List(ctx context.Context, serverID, siteID int64) ([]Command, error) {
	return listResources[Command](s.client, ctx, s.sitePath(serverID, siteID))
}

// Get returns a single command by ID.
func (s *CommandsService) Get(ctx context.Context, serverID, siteID, cmdID int64) (*Command, error) {
	return getResource[Command](s.client, ctx, fmt.Sprintf("%s/%d", s.sitePath(serverID, siteID), cmdID))
}

// Create executes a new command on a site.
func (s *CommandsService) Create(ctx context.Context, serverID, siteID int64, command string) (*Command, error) {
	body := map[string]string{"command": command}
	return createResource[Command](s.client, ctx, s.sitePath(serverID, siteID), body)
}
