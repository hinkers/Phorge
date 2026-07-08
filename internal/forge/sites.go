package forge

import (
	"context"
	"net/http"
)

// List returns all sites on a server.
func (s *SitesService) List(ctx context.Context, serverID int64) ([]Site, error) {
	return listResources[Site](s.client, ctx, s.client.orgPath("/servers/%d/sites", serverID))
}

// Get returns a single site by ID.
func (s *SitesService) Get(ctx context.Context, serverID, siteID int64) (*Site, error) {
	return getResource[Site](s.client, ctx, s.client.orgPath("/sites/%d", siteID))
}

// UpdateAliases updates the domain aliases for a site.
func (s *SitesService) UpdateAliases(ctx context.Context, serverID, siteID int64, aliases []string) (*Site, error) {
	path := s.client.orgPath("/servers/%d/sites/%d", serverID, siteID)
	body := map[string]any{"aliases": aliases}
	return updateResource[Site](s.client, ctx, path, body)
}

// UpdatePHP changes the PHP version for a site.
func (s *SitesService) UpdatePHP(ctx context.Context, serverID, siteID int64, version string) error {
	path := s.client.orgPath("/servers/%d/sites/%d", serverID, siteID)
	body := map[string]string{"php_version": version}
	return s.client.do(ctx, http.MethodPut, path, body, nil)
}
