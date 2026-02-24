package forge

import (
	"context"
	"fmt"
	"net/http"
)

// GetServerLog returns the log content for a server.
func (s *LogsService) GetServerLog(ctx context.Context, serverID int64) (string, error) {
	var resp struct {
		Content string `json:"content"`
	}
	path := fmt.Sprintf("/servers/%d/logs", serverID)
	err := s.client.do(ctx, http.MethodGet, path, nil, &resp)
	return resp.Content, err
}

// GetSiteLog returns the log content for a site.
func (s *LogsService) GetSiteLog(ctx context.Context, serverID, siteID int64) (string, error) {
	var resp struct {
		Content string `json:"content"`
	}
	path := fmt.Sprintf("/servers/%d/sites/%d/logs", serverID, siteID)
	err := s.client.do(ctx, http.MethodGet, path, nil, &resp)
	return resp.Content, err
}

// ClearSiteLog clears the log for a site.
func (s *LogsService) ClearSiteLog(ctx context.Context, serverID, siteID int64) error {
	path := fmt.Sprintf("/servers/%d/sites/%d/logs", serverID, siteID)
	return s.client.do(ctx, http.MethodDelete, path, nil, nil)
}
