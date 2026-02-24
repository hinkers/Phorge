package forge

import (
	"context"
	"fmt"
	"net/http"
)

// Get returns the environment file contents for a site as plain text.
func (s *EnvironmentService) Get(ctx context.Context, serverID, siteID int64) (string, error) {
	path := fmt.Sprintf("/servers/%d/sites/%d/env", serverID, siteID)
	return s.client.getText(ctx, path)
}

// Update replaces the environment file contents for a site.
func (s *EnvironmentService) Update(ctx context.Context, serverID, siteID int64, content string) error {
	body := map[string]string{"content": content}
	path := fmt.Sprintf("/servers/%d/sites/%d/env", serverID, siteID)
	return s.client.do(ctx, http.MethodPut, path, body, nil)
}
