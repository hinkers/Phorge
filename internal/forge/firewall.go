package forge

import (
	"context"
	"fmt"
	"net/http"
)

// FirewallCreateOpts contains the options for creating a firewall rule.
type FirewallCreateOpts struct {
	Name      string `json:"name"`
	Port      any    `json:"port"`                  // int or string
	IPAddress string `json:"ip_address,omitempty"`   // optional
	Type      string `json:"type"`                   // default "allow"
}

// List returns all firewall rules on a server.
func (s *FirewallService) List(ctx context.Context, serverID int64) ([]FirewallRule, error) {
	var resp struct {
		Rules []FirewallRule `json:"rules"`
	}
	path := fmt.Sprintf("/servers/%d/firewall-rules", serverID)
	err := s.client.do(ctx, http.MethodGet, path, nil, &resp)
	return resp.Rules, err
}

// Get returns a single firewall rule by ID.
func (s *FirewallService) Get(ctx context.Context, serverID, ruleID int64) (*FirewallRule, error) {
	var resp struct {
		Rule FirewallRule `json:"rule"`
	}
	path := fmt.Sprintf("/servers/%d/firewall-rules/%d", serverID, ruleID)
	err := s.client.do(ctx, http.MethodGet, path, nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp.Rule, nil
}

// Create creates a new firewall rule on a server.
func (s *FirewallService) Create(ctx context.Context, serverID int64, opts FirewallCreateOpts) (*FirewallRule, error) {
	var resp struct {
		Rule FirewallRule `json:"rule"`
	}
	path := fmt.Sprintf("/servers/%d/firewall-rules", serverID)
	err := s.client.do(ctx, http.MethodPost, path, opts, &resp)
	if err != nil {
		return nil, err
	}
	return &resp.Rule, nil
}

// Delete removes a firewall rule from a server.
func (s *FirewallService) Delete(ctx context.Context, serverID, ruleID int64) error {
	path := fmt.Sprintf("/servers/%d/firewall-rules/%d", serverID, ruleID)
	return s.client.do(ctx, http.MethodDelete, path, nil, nil)
}
