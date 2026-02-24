package forge

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSitesList(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/servers/1/sites" {
			t.Errorf("path = %s, want /servers/1/sites", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"sites": [
				{
					"id": 10,
					"server_id": 1,
					"name": "example.com",
					"status": "installed",
					"repository": "user/repo",
					"repository_branch": "main",
					"quick_deploy": true
				},
				{
					"id": 11,
					"server_id": 1,
					"name": "staging.example.com",
					"status": "installed",
					"quick_deploy": false
				}
			]
		}`))
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	sites, err := client.Sites.List(context.Background(), 1)
	if err != nil {
		t.Fatalf("Sites.List: %v", err)
	}

	if len(sites) != 2 {
		t.Fatalf("got %d sites, want 2", len(sites))
	}

	if sites[0].ID != 10 {
		t.Errorf("sites[0].ID = %d, want 10", sites[0].ID)
	}
	if sites[0].Name != "example.com" {
		t.Errorf("sites[0].Name = %q, want %q", sites[0].Name, "example.com")
	}
	if sites[0].Repository != "user/repo" {
		t.Errorf("sites[0].Repository = %q, want %q", sites[0].Repository, "user/repo")
	}
	if !sites[0].QuickDeploy {
		t.Error("sites[0].QuickDeploy = false, want true")
	}

	if sites[1].ID != 11 {
		t.Errorf("sites[1].ID = %d, want 11", sites[1].ID)
	}
	if sites[1].Name != "staging.example.com" {
		t.Errorf("sites[1].Name = %q, want %q", sites[1].Name, "staging.example.com")
	}
	if sites[1].QuickDeploy {
		t.Error("sites[1].QuickDeploy = true, want false")
	}
}

func TestDeploymentsDeploy(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/servers/1/sites/10/deployment/deploy" {
			t.Errorf("path = %s, want /servers/1/sites/10/deployment/deploy", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Errorf("Authorization = %q, want %q", got, "Bearer test-token")
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	err := client.Deployments.Deploy(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("Deployments.Deploy: %v", err)
	}
}

func TestEnvironmentGet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/servers/1/sites/10/env" {
			t.Errorf("path = %s, want /servers/1/sites/10/env", r.URL.Path)
		}
		if got := r.Header.Get("Accept"); got != "text/plain" {
			t.Errorf("Accept = %q, want %q", got, "text/plain")
		}

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("APP_NAME=Laravel\nAPP_ENV=production\nDB_HOST=127.0.0.1\n"))
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	env, err := client.Environment.Get(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("Environment.Get: %v", err)
	}

	expected := "APP_NAME=Laravel\nAPP_ENV=production\nDB_HOST=127.0.0.1\n"
	if env != expected {
		t.Errorf("Environment.Get = %q, want %q", env, expected)
	}
}

func TestDatabaseCreate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/servers/1/databases" {
			t.Errorf("path = %s, want /servers/1/databases", r.URL.Path)
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Errorf("Content-Type = %q, want %q", got, "application/json")
		}

		// Verify request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("reading body: %v", err)
		}
		var req map[string]any
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("unmarshalling body: %v", err)
		}

		if req["name"] != "my_db" {
			t.Errorf("body.name = %v, want %q", req["name"], "my_db")
		}
		if req["user"] != "my_user" {
			t.Errorf("body.user = %v, want %q", req["user"], "my_user")
		}
		if req["password"] != "secret" {
			t.Errorf("body.password = %v, want %q", req["password"], "secret")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"database": {
				"id": 100,
				"server_id": 1,
				"name": "my_db",
				"status": "installing",
				"is_synced": false
			}
		}`))
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	user := "my_user"
	password := "secret"
	db, err := client.Databases.Create(context.Background(), 1, "my_db", &user, &password)
	if err != nil {
		t.Fatalf("Databases.Create: %v", err)
	}

	if db.ID != 100 {
		t.Errorf("db.ID = %d, want 100", db.ID)
	}
	if db.Name != "my_db" {
		t.Errorf("db.Name = %q, want %q", db.Name, "my_db")
	}
	if db.Status != "installing" {
		t.Errorf("db.Status = %q, want %q", db.Status, "installing")
	}
	if db.IsSynced {
		t.Error("db.IsSynced = true, want false")
	}
}

func TestDatabaseCreateWithoutOptionals(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request body does NOT contain user/password
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("reading body: %v", err)
		}
		var req map[string]any
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("unmarshalling body: %v", err)
		}

		if _, ok := req["user"]; ok {
			t.Errorf("body should not contain 'user' when nil, got %v", req["user"])
		}
		if _, ok := req["password"]; ok {
			t.Errorf("body should not contain 'password' when nil, got %v", req["password"])
		}
		if req["name"] != "my_db" {
			t.Errorf("body.name = %v, want %q", req["name"], "my_db")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"database": {
				"id": 101,
				"name": "my_db",
				"status": "installing",
				"is_synced": false
			}
		}`))
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	db, err := client.Databases.Create(context.Background(), 1, "my_db", nil, nil)
	if err != nil {
		t.Fatalf("Databases.Create: %v", err)
	}

	if db.ID != 101 {
		t.Errorf("db.ID = %d, want 101", db.ID)
	}
}
