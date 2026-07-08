package forge

import (
	"context"
	"net/http"
)

// FirewallCreateOpts contains the options for creating a firewall rule.
type FirewallCreateOpts struct {
	Name      string `json:"name"`
	Port      string `json:"port,omitempty"`
	IPAddress string `json:"ip_address,omitempty"` // optional
	Type      string `json:"type"`                 // default "allow"
}

// List returns all firewall rules on a server.
func (s *FirewallService) List(ctx context.Context, serverID int64) ([]FirewallRule, error) {
	return listResources[FirewallRule](s.client, ctx, s.client.orgPath("/servers/%d/firewall-rules", serverID))
}

// Get returns a single firewall rule by ID.
func (s *FirewallService) Get(ctx context.Context, serverID, ruleID int64) (*FirewallRule, error) {
	return getResource[FirewallRule](s.client, ctx, s.client.orgPath("/servers/%d/firewall-rules/%d", serverID, ruleID))
}

// Create creates a new firewall rule on a server.
func (s *FirewallService) Create(ctx context.Context, serverID int64, opts FirewallCreateOpts) (*FirewallRule, error) {
	return createResource[FirewallRule](s.client, ctx, s.client.orgPath("/servers/%d/firewall-rules", serverID), opts)
}

// Delete removes a firewall rule from a server.
func (s *FirewallService) Delete(ctx context.Context, serverID, ruleID int64) error {
	return s.client.do(ctx, http.MethodDelete, s.client.orgPath("/servers/%d/firewall-rules/%d", serverID, ruleID), nil, nil)
}
