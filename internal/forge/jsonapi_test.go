package forge

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetResource(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.api+json")
		w.Write([]byte(`{
			"data": {
				"id": "1",
				"type": "servers",
				"attributes": {
					"id": 1,
					"name": "test-server",
					"ip_address": "1.2.3.4",
					"is_ready": true
				}
			}
		}`))
	}))
	defer srv.Close()

	c := NewClient("test-token", "test-org")
	c.BaseURL = srv.URL

	server, err := getResource[Server](c, context.Background(), "/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if server.Name != "test-server" {
		t.Errorf("Name = %q, want %q", server.Name, "test-server")
	}
	if server.ID != 1 {
		t.Errorf("ID = %d, want 1", server.ID)
	}
}

func TestListResources(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.api+json")
		w.Write([]byte(`{
			"data": [
				{"id": "1", "type": "servers", "attributes": {"id": 1, "name": "srv-1", "is_ready": true}},
				{"id": "2", "type": "servers", "attributes": {"id": 2, "name": "srv-2", "is_ready": false}}
			],
			"meta": {"per_page": 30, "next_cursor": null, "prev_cursor": null}
		}`))
	}))
	defer srv.Close()

	c := NewClient("test-token", "test-org")
	c.BaseURL = srv.URL

	servers, err := listResources[Server](c, context.Background(), "/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(servers) != 2 {
		t.Fatalf("got %d servers, want 2", len(servers))
	}
	if servers[0].Name != "srv-1" {
		t.Errorf("servers[0].Name = %q, want %q", servers[0].Name, "srv-1")
	}
}

func TestCreateResource(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if got := r.Header.Get("Content-Type"); got != "application/vnd.api+json" {
			t.Errorf("Content-Type = %q, want %q", got, "application/vnd.api+json")
		}
		w.Header().Set("Content-Type", "application/vnd.api+json")
		w.Write([]byte(`{
			"data": {
				"id": "3",
				"type": "servers",
				"attributes": {"id": 3, "name": "new-server", "is_ready": false}
			}
		}`))
	}))
	defer srv.Close()

	c := NewClient("test-token", "test-org")
	c.BaseURL = srv.URL

	server, err := createResource[Server](c, context.Background(), "/test", map[string]string{"name": "new-server"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if server.Name != "new-server" {
		t.Errorf("Name = %q, want %q", server.Name, "new-server")
	}
	if server.ID != 3 {
		t.Errorf("ID = %d, want 3", server.ID)
	}
}

func TestUpdateResource(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("method = %s, want PUT", r.Method)
		}
		w.Header().Set("Content-Type", "application/vnd.api+json")
		w.Write([]byte(`{
			"data": {
				"id": "1",
				"type": "servers",
				"attributes": {"id": 1, "name": "renamed-server", "is_ready": true}
			}
		}`))
	}))
	defer srv.Close()

	c := NewClient("test-token", "test-org")
	c.BaseURL = srv.URL

	server, err := updateResource[Server](c, context.Background(), "/test/1", map[string]string{"name": "renamed-server"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if server.Name != "renamed-server" {
		t.Errorf("Name = %q, want %q", server.Name, "renamed-server")
	}
}

func TestOrgPath(t *testing.T) {
	c := NewClient("tok", "myorg")
	got := c.orgPath("/servers/%d", 42)
	want := "/orgs/myorg/servers/42"
	if got != want {
		t.Errorf("orgPath = %q, want %q", got, want)
	}
}

func TestOrgPathNoArgs(t *testing.T) {
	c := NewClient("tok", "myorg")
	got := c.orgPath("/servers")
	want := "/orgs/myorg/servers"
	if got != want {
		t.Errorf("orgPath = %q, want %q", got, want)
	}
}
