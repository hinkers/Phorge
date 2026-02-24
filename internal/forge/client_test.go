package forge

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// newTestClient creates a Client pointed at the given httptest.Server.
func newTestClient(t *testing.T, srv *httptest.Server) *Client {
	t.Helper()
	c := NewClient("test-token")
	c.BaseURL = srv.URL
	return c
}

func TestListServers(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/servers" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Errorf("Authorization = %q, want %q", got, "Bearer test-token")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"servers": [
				{
					"id": 1,
					"name": "production",
					"ip_address": "10.0.0.1",
					"is_ready": true,
					"status": "installed"
				},
				{
					"id": 2,
					"name": "staging",
					"ip_address": "10.0.0.2",
					"is_ready": false,
					"status": "installing"
				}
			]
		}`))
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	servers, err := client.Servers.List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	if len(servers) != 2 {
		t.Fatalf("got %d servers, want 2", len(servers))
	}

	if servers[0].ID != 1 {
		t.Errorf("servers[0].ID = %d, want 1", servers[0].ID)
	}
	if servers[0].Name != "production" {
		t.Errorf("servers[0].Name = %q, want %q", servers[0].Name, "production")
	}
	if servers[0].IPAddress != "10.0.0.1" {
		t.Errorf("servers[0].IPAddress = %q, want %q", servers[0].IPAddress, "10.0.0.1")
	}
	if !servers[0].IsReady {
		t.Error("servers[0].IsReady = false, want true")
	}

	if servers[1].ID != 2 {
		t.Errorf("servers[1].ID = %d, want 2", servers[1].ID)
	}
	if servers[1].Name != "staging" {
		t.Errorf("servers[1].Name = %q, want %q", servers[1].Name, "staging")
	}
	if servers[1].IsReady {
		t.Error("servers[1].IsReady = true, want false")
	}
}

func TestAuthError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message": "Unauthenticated."}`))
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	_, err := client.Servers.List(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var authErr *AuthenticationError
	if !errors.As(err, &authErr) {
		t.Fatalf("expected AuthenticationError, got %T: %v", err, err)
	}
	if authErr.StatusCode != 401 {
		t.Errorf("StatusCode = %d, want 401", authErr.StatusCode)
	}
	if authErr.Message != "Unauthenticated." {
		t.Errorf("Message = %q, want %q", authErr.Message, "Unauthenticated.")
	}
}

func TestRateLimitError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"message": "Too Many Attempts."}`))
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	_, err := client.Servers.List(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var rlErr *RateLimitError
	if !errors.As(err, &rlErr) {
		t.Fatalf("expected RateLimitError, got %T: %v", err, err)
	}
	if rlErr.StatusCode != 429 {
		t.Errorf("StatusCode = %d, want 429", rlErr.StatusCode)
	}
}

func TestNotFoundError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message": "Resource not found."}`))
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	_, err := client.Servers.List(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var nfErr *NotFoundError
	if !errors.As(err, &nfErr) {
		t.Fatalf("expected NotFoundError, got %T: %v", err, err)
	}
	if nfErr.StatusCode != 404 {
		t.Errorf("StatusCode = %d, want 404", nfErr.StatusCode)
	}
	if nfErr.Message != "Resource not found." {
		t.Errorf("Message = %q, want %q", nfErr.Message, "Resource not found.")
	}
}

func TestValidationError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{
			"message": "The given data was invalid.",
			"errors": {
				"name": ["The name field is required."],
				"provider": ["The provider field is required.", "The provider must be valid."]
			}
		}`))
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	_, err := client.Servers.List(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var valErr *ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("expected ValidationError, got %T: %v", err, err)
	}
	if valErr.StatusCode != 422 {
		t.Errorf("StatusCode = %d, want 422", valErr.StatusCode)
	}
	if valErr.Message != "The given data was invalid." {
		t.Errorf("Message = %q, want %q", valErr.Message, "The given data was invalid.")
	}

	if len(valErr.Details) != 2 {
		t.Fatalf("Details has %d keys, want 2", len(valErr.Details))
	}

	nameErrs := valErr.Details["name"]
	if len(nameErrs) != 1 {
		t.Fatalf("Details[name] has %d entries, want 1", len(nameErrs))
	}
	if nameErrs[0] != "The name field is required." {
		t.Errorf("Details[name][0] = %q, want %q", nameErrs[0], "The name field is required.")
	}

	providerErrs := valErr.Details["provider"]
	if len(providerErrs) != 2 {
		t.Fatalf("Details[provider] has %d entries, want 2", len(providerErrs))
	}
	if providerErrs[0] != "The provider field is required." {
		t.Errorf("Details[provider][0] = %q, want %q", providerErrs[0], "The provider field is required.")
	}
	if providerErrs[1] != "The provider must be valid." {
		t.Errorf("Details[provider][1] = %q, want %q", providerErrs[1], "The provider must be valid.")
	}
}

func TestGetText(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/servers/1/sites/2/env" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("Accept"); got != "text/plain" {
			t.Errorf("Accept = %q, want %q", got, "text/plain")
		}

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("APP_NAME=Laravel\nAPP_ENV=production\n"))
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	text, err := client.getText(context.Background(), "/servers/1/sites/2/env")
	if err != nil {
		t.Fatalf("getText: %v", err)
	}

	expected := "APP_NAME=Laravel\nAPP_ENV=production\n"
	if text != expected {
		t.Errorf("getText = %q, want %q", text, expected)
	}
}

func TestNewClientDefaults(t *testing.T) {
	c := NewClient("my-token")

	if c.BaseURL != "https://forge.laravel.com/api/v1" {
		t.Errorf("BaseURL = %q, want %q", c.BaseURL, "https://forge.laravel.com/api/v1")
	}
	if c.http == nil {
		t.Fatal("http client is nil")
	}
	if c.Servers == nil {
		t.Fatal("Servers service is nil")
	}
	if c.Sites == nil {
		t.Fatal("Sites service is nil")
	}
	if c.Deployments == nil {
		t.Fatal("Deployments service is nil")
	}
	if c.Databases == nil {
		t.Fatal("Databases service is nil")
	}
	if c.Environment == nil {
		t.Fatal("Environment service is nil")
	}
	if c.Certificates == nil {
		t.Fatal("Certificates service is nil")
	}
	if c.Workers == nil {
		t.Fatal("Workers service is nil")
	}
	if c.Daemons == nil {
		t.Fatal("Daemons service is nil")
	}
	if c.Firewall == nil {
		t.Fatal("Firewall service is nil")
	}
	if c.Jobs == nil {
		t.Fatal("Jobs service is nil")
	}
	if c.Backups == nil {
		t.Fatal("Backups service is nil")
	}
	if c.SSHKeys == nil {
		t.Fatal("SSHKeys service is nil")
	}
	if c.Commands == nil {
		t.Fatal("Commands service is nil")
	}
	if c.Git == nil {
		t.Fatal("Git service is nil")
	}
	if c.Logs == nil {
		t.Fatal("Logs service is nil")
	}
}
