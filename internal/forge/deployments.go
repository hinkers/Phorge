package forge

import (
	"context"
	"fmt"
	"net/http"
)

// List returns deployment history for a site.
func (s *DeploymentsService) List(ctx context.Context, serverID, siteID int64) ([]Deployment, error) {
	var resp struct {
		Deployments []Deployment `json:"deployments"`
	}
	path := fmt.Sprintf("/servers/%d/sites/%d/deployment-history", serverID, siteID)
	err := s.client.do(ctx, http.MethodGet, path, nil, &resp)
	return resp.Deployments, err
}

// Get returns a single deployment by ID.
func (s *DeploymentsService) Get(ctx context.Context, serverID, siteID, deployID int64) (*Deployment, error) {
	var resp struct {
		Deployment Deployment `json:"deployment"`
	}
	path := fmt.Sprintf("/servers/%d/sites/%d/deployment-history/%d", serverID, siteID, deployID)
	err := s.client.do(ctx, http.MethodGet, path, nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp.Deployment, nil
}

// GetOutput returns the output of a specific deployment.
func (s *DeploymentsService) GetOutput(ctx context.Context, serverID, siteID, deployID int64) (string, error) {
	var resp struct {
		Output string `json:"output"`
	}
	path := fmt.Sprintf("/servers/%d/sites/%d/deployment-history/%d/output", serverID, siteID, deployID)
	err := s.client.do(ctx, http.MethodGet, path, nil, &resp)
	return resp.Output, err
}

// Deploy triggers a new deployment for the site.
func (s *DeploymentsService) Deploy(ctx context.Context, serverID, siteID int64) error {
	path := fmt.Sprintf("/servers/%d/sites/%d/deployment/deploy", serverID, siteID)
	return s.client.do(ctx, http.MethodPost, path, nil, nil)
}

// GetLog returns the latest deployment log for the site.
func (s *DeploymentsService) GetLog(ctx context.Context, serverID, siteID int64) (string, error) {
	var resp struct {
		Output string `json:"output"`
	}
	path := fmt.Sprintf("/servers/%d/sites/%d/deployment/log", serverID, siteID)
	err := s.client.do(ctx, http.MethodGet, path, nil, &resp)
	return resp.Output, err
}

// GetScript returns the deployment script contents as plain text.
func (s *DeploymentsService) GetScript(ctx context.Context, serverID, siteID int64) (string, error) {
	path := fmt.Sprintf("/servers/%d/sites/%d/deployment/script", serverID, siteID)
	return s.client.getText(ctx, path)
}

// UpdateScript replaces the deployment script content.
func (s *DeploymentsService) UpdateScript(ctx context.Context, serverID, siteID int64, content string) error {
	body := map[string]string{"content": content}
	path := fmt.Sprintf("/servers/%d/sites/%d/deployment/script", serverID, siteID)
	return s.client.do(ctx, http.MethodPut, path, body, nil)
}

// EnableQuickDeploy enables quick deploy (push-to-deploy) for the site.
func (s *DeploymentsService) EnableQuickDeploy(ctx context.Context, serverID, siteID int64) error {
	path := fmt.Sprintf("/servers/%d/sites/%d/deployment", serverID, siteID)
	return s.client.do(ctx, http.MethodPost, path, nil, nil)
}

// DisableQuickDeploy disables quick deploy for the site.
func (s *DeploymentsService) DisableQuickDeploy(ctx context.Context, serverID, siteID int64) error {
	path := fmt.Sprintf("/servers/%d/sites/%d/deployment", serverID, siteID)
	return s.client.do(ctx, http.MethodDelete, path, nil, nil)
}

// ResetStatus resets the deployment status for the site.
func (s *DeploymentsService) ResetStatus(ctx context.Context, serverID, siteID int64) error {
	path := fmt.Sprintf("/servers/%d/sites/%d/deployment/reset", serverID, siteID)
	return s.client.do(ctx, http.MethodPost, path, nil, nil)
}
