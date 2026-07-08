package forge

import (
	"context"
	"fmt"
	"net/http"
)

type deploymentOutput struct {
	Output string `json:"output"`
}

type deploymentScript struct {
	Content    string `json:"content"`
	AutoSource bool   `json:"auto_source"`
}

func (s *DeploymentsService) sitePath(serverID, siteID int64, suffix string) string {
	return s.client.orgPath("/servers/%d/sites/%d/deployments%s", serverID, siteID, suffix)
}

// List returns deployment history for a site.
func (s *DeploymentsService) List(ctx context.Context, serverID, siteID int64) ([]Deployment, error) {
	return listResources[Deployment](s.client, ctx, s.sitePath(serverID, siteID, ""))
}

// Get returns a single deployment by ID.
func (s *DeploymentsService) Get(ctx context.Context, serverID, siteID, deployID int64) (*Deployment, error) {
	return getResource[Deployment](s.client, ctx, s.sitePath(serverID, siteID, fmt.Sprintf("/%d", deployID)))
}

// GetOutput returns the output of a specific deployment.
func (s *DeploymentsService) GetOutput(ctx context.Context, serverID, siteID, deployID int64) (string, error) {
	out, err := getResource[deploymentOutput](s.client, ctx, s.sitePath(serverID, siteID, fmt.Sprintf("/%d/log", deployID)))
	if err != nil {
		return "", err
	}
	return out.Output, nil
}

// Deploy triggers a new deployment for the site.
func (s *DeploymentsService) Deploy(ctx context.Context, serverID, siteID int64) error {
	return s.client.do(ctx, http.MethodPost, s.sitePath(serverID, siteID, ""), nil, nil)
}

// GetLog returns the latest deployment log for the site. The new API has no
// separate "live log" endpoint, so this fetches the most recent deployment
// via List and then its output via GetOutput.
func (s *DeploymentsService) GetLog(ctx context.Context, serverID, siteID int64) (string, error) {
	deployments, err := s.List(ctx, serverID, siteID)
	if err != nil {
		return "", err
	}
	if len(deployments) == 0 {
		return "", nil
	}
	return s.GetOutput(ctx, serverID, siteID, deployments[0].ID)
}

// GetScript returns the deployment script contents.
func (s *DeploymentsService) GetScript(ctx context.Context, serverID, siteID int64) (string, error) {
	script, err := getResource[deploymentScript](s.client, ctx, s.sitePath(serverID, siteID, "/script"))
	if err != nil {
		return "", err
	}
	return script.Content, nil
}

// UpdateScript replaces the deployment script content.
func (s *DeploymentsService) UpdateScript(ctx context.Context, serverID, siteID int64, content string) error {
	path := s.sitePath(serverID, siteID, "/script")
	body := map[string]string{"content": content}
	return s.client.do(ctx, http.MethodPut, path, body, nil)
}

// EnableQuickDeploy enables quick deploy (push-to-deploy) for the site.
func (s *DeploymentsService) EnableQuickDeploy(ctx context.Context, serverID, siteID int64) error {
	return s.client.do(ctx, http.MethodPost, s.sitePath(serverID, siteID, "/push-to-deploy"), nil, nil)
}

// DisableQuickDeploy disables quick deploy for the site.
func (s *DeploymentsService) DisableQuickDeploy(ctx context.Context, serverID, siteID int64) error {
	return s.client.do(ctx, http.MethodDelete, s.sitePath(serverID, siteID, "/push-to-deploy"), nil, nil)
}

// ResetStatus resets the deployment status for the site.
func (s *DeploymentsService) ResetStatus(ctx context.Context, serverID, siteID int64) error {
	return s.client.do(ctx, http.MethodDelete, s.sitePath(serverID, siteID, "/status"), nil, nil)
}
