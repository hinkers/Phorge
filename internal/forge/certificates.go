package forge

import (
	"context"
	"fmt"
	"net/http"
)

// sitePath builds the org-scoped certificates collection path for a site.
func (s *CertificatesService) sitePath(serverID, siteID int64) string {
	return s.client.orgPath("/servers/%d/sites/%d/certificates", serverID, siteID)
}

// List returns all SSL certificates for a site.
func (s *CertificatesService) List(ctx context.Context, serverID, siteID int64) ([]Certificate, error) {
	return listResources[Certificate](s.client, ctx, s.sitePath(serverID, siteID))
}

// Get returns a single certificate by ID.
func (s *CertificatesService) Get(ctx context.Context, serverID, siteID, certID int64) (*Certificate, error) {
	return getResource[Certificate](s.client, ctx, fmt.Sprintf("%s/%d", s.sitePath(serverID, siteID), certID))
}

// CreateLetsEncrypt creates a new Let's Encrypt certificate for the given domains.
func (s *CertificatesService) CreateLetsEncrypt(ctx context.Context, serverID, siteID int64, domains []string) (*Certificate, error) {
	body := map[string]any{"domains": domains, "type": "letsencrypt"}
	return createResource[Certificate](s.client, ctx, s.sitePath(serverID, siteID), body)
}

// Activate activates an SSL certificate.
func (s *CertificatesService) Activate(ctx context.Context, serverID, siteID, certID int64) error {
	path := fmt.Sprintf("%s/%d/actions", s.sitePath(serverID, siteID), certID)
	body := map[string]string{"action": "activate"}
	return s.client.do(ctx, http.MethodPost, path, body, nil)
}

// Delete removes an SSL certificate.
func (s *CertificatesService) Delete(ctx context.Context, serverID, siteID, certID int64) error {
	return s.client.do(ctx, http.MethodDelete, fmt.Sprintf("%s/%d", s.sitePath(serverID, siteID), certID), nil, nil)
}
