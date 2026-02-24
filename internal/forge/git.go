package forge

import (
	"context"
	"fmt"
	"net/http"
)

// Install installs a Git repository on a site.
func (s *GitService) Install(ctx context.Context, serverID, siteID int64, provider, repo, branch string, composer bool) error {
	body := map[string]any{
		"provider":   provider,
		"repository": repo,
		"branch":     branch,
		"composer":   composer,
	}
	path := fmt.Sprintf("/servers/%d/sites/%d/git", serverID, siteID)
	return s.client.do(ctx, http.MethodPost, path, body, nil)
}

// UpdateBranch changes the deployed branch for a site.
func (s *GitService) UpdateBranch(ctx context.Context, serverID, siteID int64, branch string) error {
	body := map[string]string{"branch": branch}
	path := fmt.Sprintf("/servers/%d/sites/%d/git", serverID, siteID)
	return s.client.do(ctx, http.MethodPut, path, body, nil)
}

// Remove removes Git integration from a site.
func (s *GitService) Remove(ctx context.Context, serverID, siteID int64) error {
	path := fmt.Sprintf("/servers/%d/sites/%d/git", serverID, siteID)
	return s.client.do(ctx, http.MethodDelete, path, nil, nil)
}
