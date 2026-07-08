package forge

import (
	"context"
	"net/http"
)

// envContent wraps the "content" attribute of the environment JSON:API resource.
type envContent struct {
	Content string `json:"content"`
}

// Get returns the environment file contents for a site.
func (s *EnvironmentService) Get(ctx context.Context, serverID, siteID int64) (string, error) {
	out, err := getResource[envContent](s.client, ctx, s.client.orgPath("/servers/%d/sites/%d/environment", serverID, siteID))
	if err != nil {
		return "", err
	}
	return out.Content, nil
}

// Update replaces the environment file contents for a site.
func (s *EnvironmentService) Update(ctx context.Context, serverID, siteID int64, content string) error {
	path := s.client.orgPath("/servers/%d/sites/%d/environment", serverID, siteID)
	body := map[string]string{"content": content}
	return s.client.do(ctx, http.MethodPut, path, body, nil)
}
