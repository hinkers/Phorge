package forge

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const defaultBaseURL = "https://forge.laravel.com/api/v1"

// Client is the entry point for the Laravel Forge API.
// Services are accessed through the exported fields (e.g. client.Servers).
type Client struct {
	BaseURL string
	token   string
	http    *http.Client

	// Services
	Servers      *ServersService
	Sites        *SitesService
	Deployments  *DeploymentsService
	Databases    *DatabasesService
	Environment  *EnvironmentService
	Certificates *CertificatesService
	Workers      *WorkersService
	Daemons      *DaemonsService
	Firewall     *FirewallService
	Jobs         *JobsService
	Backups      *BackupsService
	SSHKeys      *SSHKeysService
	Commands     *CommandsService
	Git          *GitService
	Logs         *LogsService
}

// Service types -- each holds a back-pointer to the parent Client.

type ServersService struct{ client *Client }
type SitesService struct{ client *Client }
type DeploymentsService struct{ client *Client }
type DatabasesService struct{ client *Client }
type EnvironmentService struct{ client *Client }
type CertificatesService struct{ client *Client }
type WorkersService struct{ client *Client }
type DaemonsService struct{ client *Client }
type FirewallService struct{ client *Client }
type JobsService struct{ client *Client }
type BackupsService struct{ client *Client }
type SSHKeysService struct{ client *Client }
type CommandsService struct{ client *Client }
type GitService struct{ client *Client }
type LogsService struct{ client *Client }

// NewClient creates a new Forge API client authenticated with the given token.
func NewClient(token string) *Client {
	c := &Client{
		BaseURL: defaultBaseURL,
		token:   token,
		http:    &http.Client{},
	}

	c.Servers = &ServersService{client: c}
	c.Sites = &SitesService{client: c}
	c.Deployments = &DeploymentsService{client: c}
	c.Databases = &DatabasesService{client: c}
	c.Environment = &EnvironmentService{client: c}
	c.Certificates = &CertificatesService{client: c}
	c.Workers = &WorkersService{client: c}
	c.Daemons = &DaemonsService{client: c}
	c.Firewall = &FirewallService{client: c}
	c.Jobs = &JobsService{client: c}
	c.Backups = &BackupsService{client: c}
	c.SSHKeys = &SSHKeysService{client: c}
	c.Commands = &CommandsService{client: c}
	c.Git = &GitService{client: c}
	c.Logs = &LogsService{client: c}

	return c
}

// do executes an API request. If body is non-nil it is marshalled as JSON.
// If result is non-nil the response body is decoded into it.
func (c *Client) do(ctx context.Context, method, path string, body any, result any) error {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshalling request body: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, reqBody)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return parseError(resp)
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("decoding response: %w", err)
		}
	}

	return nil
}

// getText fetches a plain-text response (e.g. environment files, deploy scripts).
func (c *Client) getText(ctx context.Context, path string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+path, nil)
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "text/plain")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", parseError(resp)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading response body: %w", err)
	}

	return string(data), nil
}

// parseError maps an HTTP error response to the appropriate error type.
func parseError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)

	// Try to extract a message from the JSON body.
	var payload struct {
		Message string              `json:"message"`
		Errors  map[string][]string `json:"errors"`
	}
	_ = json.Unmarshal(body, &payload)

	if payload.Message == "" {
		payload.Message = http.StatusText(resp.StatusCode)
	}

	base := APIError{
		StatusCode: resp.StatusCode,
		Message:    payload.Message,
	}

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return &AuthenticationError{APIError: base}
	case http.StatusNotFound:
		return &NotFoundError{APIError: base}
	case http.StatusUnprocessableEntity:
		return &ValidationError{APIError: base, Details: payload.Errors}
	case http.StatusTooManyRequests:
		return &RateLimitError{APIError: base}
	default:
		return &base
	}
}

