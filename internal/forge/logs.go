package forge

import (
	"context"
	"net/http"
)

type logContent struct {
	Content string `json:"content"`
}

// GetServerLog returns the log content for a server.
func (s *LogsService) GetServerLog(ctx context.Context, serverID int64) (string, error) {
	out, err := getResource[logContent](s.client, ctx, s.client.orgPath("/servers/%d/logs/syslog", serverID))
	if err != nil {
		return "", err
	}
	return out.Content, nil
}

// GetSiteLog returns the log content for a site.
func (s *LogsService) GetSiteLog(ctx context.Context, serverID, siteID int64) (string, error) {
	out, err := getResource[logContent](s.client, ctx, s.client.orgPath("/servers/%d/sites/%d/logs/application", serverID, siteID))
	if err != nil {
		return "", err
	}
	return out.Content, nil
}

// ClearSiteLog clears the log for a site.
func (s *LogsService) ClearSiteLog(ctx context.Context, serverID, siteID int64) error {
	return s.client.do(ctx, http.MethodDelete, s.client.orgPath("/servers/%d/sites/%d/logs/application", serverID, siteID), nil, nil)
}
