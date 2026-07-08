package forge

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOrganizationsList(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/orgs" {
			t.Errorf("path = %q, want /orgs", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/vnd.api+json")
		w.Write([]byte(`{
			"data": [
				{"id": "1", "type": "organizations", "attributes": {"name": "My Org", "slug": "my-org"}}
			],
			"meta": {"per_page": 30, "next_cursor": null, "prev_cursor": null}
		}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	orgs, err := c.Organizations.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(orgs) != 1 {
		t.Fatalf("got %d orgs, want 1", len(orgs))
	}
	if orgs[0].Slug != "my-org" {
		t.Errorf("Slug = %q, want %q", orgs[0].Slug, "my-org")
	}
	if orgs[0].Name != "My Org" {
		t.Errorf("Name = %q, want %q", orgs[0].Name, "My Org")
	}
}

func TestOrganizationsGet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/orgs/my-org" {
			t.Errorf("path = %q, want /orgs/my-org", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/vnd.api+json")
		w.Write([]byte(`{
			"data": {"id": "1", "type": "organizations", "attributes": {"name": "My Org", "slug": "my-org", "created_at": "2024-01-01T00:00:00Z"}}
		}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	org, err := c.Organizations.Get(context.Background(), "my-org")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if org.Slug != "my-org" {
		t.Errorf("Slug = %q, want %q", org.Slug, "my-org")
	}
	if org.Name != "My Org" {
		t.Errorf("Name = %q, want %q", org.Name, "My Org")
	}
	if org.CreatedAt != "2024-01-01T00:00:00Z" {
		t.Errorf("CreatedAt = %q, want %q", org.CreatedAt, "2024-01-01T00:00:00Z")
	}
}

func TestOrganizationsNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message": "Resource not found."}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.Organizations.Get(context.Background(), "missing-org")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
