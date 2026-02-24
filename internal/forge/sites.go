package forge

import (
	"context"
	"fmt"
	"net/http"
)

// List returns all sites on a server.
func (s *SitesService) List(ctx context.Context, serverID int64) ([]Site, error) {
	var resp struct {
		Sites []Site `json:"sites"`
	}
	err := s.client.do(ctx, http.MethodGet, fmt.Sprintf("/servers/%d/sites", serverID), nil, &resp)
	return resp.Sites, err
}

// Get returns a single site by ID.
func (s *SitesService) Get(ctx context.Context, serverID, siteID int64) (*Site, error) {
	var resp struct {
		Site Site `json:"site"`
	}
	path := fmt.Sprintf("/servers/%d/sites/%d", serverID, siteID)
	err := s.client.do(ctx, http.MethodGet, path, nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp.Site, nil
}

// UpdateAliases updates the domain aliases for a site.
func (s *SitesService) UpdateAliases(ctx context.Context, serverID, siteID int64, aliases []string) (*Site, error) {
	body := map[string]any{"aliases": aliases}
	var resp struct {
		Site Site `json:"site"`
	}
	path := fmt.Sprintf("/servers/%d/sites/%d/aliases", serverID, siteID)
	err := s.client.do(ctx, http.MethodPut, path, body, &resp)
	if err != nil {
		return nil, err
	}
	return &resp.Site, nil
}

// UpdatePHP changes the PHP version for a site.
func (s *SitesService) UpdatePHP(ctx context.Context, serverID, siteID int64, version string) error {
	body := map[string]string{"version": version}
	path := fmt.Sprintf("/servers/%d/sites/%d/php", serverID, siteID)
	return s.client.do(ctx, http.MethodPut, path, body, nil)
}
