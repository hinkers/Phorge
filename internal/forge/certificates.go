package forge

import (
	"context"
	"fmt"
	"net/http"
)

// List returns all SSL certificates for a site.
func (s *CertificatesService) List(ctx context.Context, serverID, siteID int64) ([]Certificate, error) {
	var resp struct {
		Certificates []Certificate `json:"certificates"`
	}
	path := fmt.Sprintf("/servers/%d/sites/%d/certificates", serverID, siteID)
	err := s.client.do(ctx, http.MethodGet, path, nil, &resp)
	return resp.Certificates, err
}

// Get returns a single certificate by ID.
func (s *CertificatesService) Get(ctx context.Context, serverID, siteID, certID int64) (*Certificate, error) {
	var resp struct {
		Certificate Certificate `json:"certificate"`
	}
	path := fmt.Sprintf("/servers/%d/sites/%d/certificates/%d", serverID, siteID, certID)
	err := s.client.do(ctx, http.MethodGet, path, nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp.Certificate, nil
}

// CreateLetsEncrypt creates a new Let's Encrypt certificate for the given domains.
func (s *CertificatesService) CreateLetsEncrypt(ctx context.Context, serverID, siteID int64, domains []string) (*Certificate, error) {
	body := map[string]any{"domains": domains}
	var resp struct {
		Certificate Certificate `json:"certificate"`
	}
	path := fmt.Sprintf("/servers/%d/sites/%d/certificates/letsencrypt", serverID, siteID)
	err := s.client.do(ctx, http.MethodPost, path, body, &resp)
	if err != nil {
		return nil, err
	}
	return &resp.Certificate, nil
}

// Activate activates an SSL certificate.
func (s *CertificatesService) Activate(ctx context.Context, serverID, siteID, certID int64) error {
	path := fmt.Sprintf("/servers/%d/sites/%d/certificates/%d/activate", serverID, siteID, certID)
	return s.client.do(ctx, http.MethodPost, path, nil, nil)
}

// Delete removes an SSL certificate.
func (s *CertificatesService) Delete(ctx context.Context, serverID, siteID, certID int64) error {
	path := fmt.Sprintf("/servers/%d/sites/%d/certificates/%d", serverID, siteID, certID)
	return s.client.do(ctx, http.MethodDelete, path, nil, nil)
}
