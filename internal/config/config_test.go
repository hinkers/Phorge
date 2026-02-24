package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultValues(t *testing.T) {
	cfg := Default()

	if cfg.Forge.SSHUser != "forge" {
		t.Errorf("Default ssh_user = %q, want %q", cfg.Forge.SSHUser, "forge")
	}
	if cfg.Editor.Command != "vim" {
		t.Errorf("Default editor command = %q, want %q", cfg.Editor.Command, "vim")
	}
	if cfg.Forge.APIKey != "" {
		t.Errorf("Default api_key = %q, want empty", cfg.Forge.APIKey)
	}
	if cfg.ServerUsers == nil {
		t.Error("Default ServerUsers is nil, want initialized map")
	}
}

func TestDefaultPathNotEmpty(t *testing.T) {
	p := DefaultPath()
	if p == "" {
		t.Fatal("DefaultPath() returned empty string")
	}
	if filepath.Base(p) != "config.toml" {
		t.Errorf("DefaultPath() basename = %q, want %q", filepath.Base(p), "config.toml")
	}
}

func TestLoadFromMissingFile(t *testing.T) {
	cfg, err := LoadFrom(filepath.Join(t.TempDir(), "nonexistent.toml"))
	if err != nil {
		t.Fatalf("LoadFrom missing file: %v", err)
	}

	// Should return defaults.
	if cfg.Forge.SSHUser != "forge" {
		t.Errorf("ssh_user = %q, want %q", cfg.Forge.SSHUser, "forge")
	}
	if cfg.Editor.Command != "vim" {
		t.Errorf("editor command = %q, want %q", cfg.Editor.Command, "vim")
	}
}

func TestLoadFromTOML(t *testing.T) {
	content := `
[forge]
api_key = "test-key-123"
ssh_user = "ubuntu"

[editor]
command = "nano"

[server_users]
my-server = "root"
staging = "deployer"
`
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadFrom(path)
	if err != nil {
		t.Fatalf("LoadFrom: %v", err)
	}

	if cfg.Forge.APIKey != "test-key-123" {
		t.Errorf("api_key = %q, want %q", cfg.Forge.APIKey, "test-key-123")
	}
	if cfg.Forge.SSHUser != "ubuntu" {
		t.Errorf("ssh_user = %q, want %q", cfg.Forge.SSHUser, "ubuntu")
	}
	if cfg.Editor.Command != "nano" {
		t.Errorf("editor command = %q, want %q", cfg.Editor.Command, "nano")
	}
	if len(cfg.ServerUsers) != 2 {
		t.Fatalf("ServerUsers length = %d, want 2", len(cfg.ServerUsers))
	}
	if cfg.ServerUsers["my-server"] != "root" {
		t.Errorf("ServerUsers[my-server] = %q, want %q", cfg.ServerUsers["my-server"], "root")
	}
	if cfg.ServerUsers["staging"] != "deployer" {
		t.Errorf("ServerUsers[staging] = %q, want %q", cfg.ServerUsers["staging"], "deployer")
	}
}

func TestLoadFromPartialTOML(t *testing.T) {
	// Only forge section present; editor and server_users should get defaults.
	content := `
[forge]
api_key = "partial-key"
`
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadFrom(path)
	if err != nil {
		t.Fatalf("LoadFrom: %v", err)
	}

	if cfg.Forge.APIKey != "partial-key" {
		t.Errorf("api_key = %q, want %q", cfg.Forge.APIKey, "partial-key")
	}
	// Defaults should be preserved for fields not in the file.
	if cfg.Editor.Command != "vim" {
		t.Errorf("editor command = %q, want default %q", cfg.Editor.Command, "vim")
	}
	if cfg.Forge.SSHUser != "forge" {
		t.Errorf("ssh_user = %q, want default %q", cfg.Forge.SSHUser, "forge")
	}
}

func TestSaveAndReload(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "subdir", "config.toml")

	cfg := Default()
	cfg.Forge.APIKey = "round-trip-key"
	cfg.Forge.SSHUser = "admin"
	cfg.Editor.Command = "emacs"
	cfg.ServerUsers["prod"] = "www-data"

	if err := cfg.SaveTo(path); err != nil {
		t.Fatalf("SaveTo: %v", err)
	}

	// Verify the directory was created.
	info, err := os.Stat(filepath.Dir(path))
	if err != nil {
		t.Fatalf("Stat dir: %v", err)
	}
	if !info.IsDir() {
		t.Fatal("Expected directory to be created")
	}

	// Reload and verify.
	loaded, err := LoadFrom(path)
	if err != nil {
		t.Fatalf("LoadFrom after save: %v", err)
	}

	if loaded.Forge.APIKey != "round-trip-key" {
		t.Errorf("api_key = %q, want %q", loaded.Forge.APIKey, "round-trip-key")
	}
	if loaded.Forge.SSHUser != "admin" {
		t.Errorf("ssh_user = %q, want %q", loaded.Forge.SSHUser, "admin")
	}
	if loaded.Editor.Command != "emacs" {
		t.Errorf("editor command = %q, want %q", loaded.Editor.Command, "emacs")
	}
	if loaded.ServerUsers["prod"] != "www-data" {
		t.Errorf("ServerUsers[prod] = %q, want %q", loaded.ServerUsers["prod"], "www-data")
	}
}

func TestSSHUserForFallback(t *testing.T) {
	cfg := Default()
	cfg.Forge.SSHUser = "forge"
	cfg.ServerUsers["special"] = "root"

	// Per-server override.
	if got := cfg.SSHUserFor("special"); got != "root" {
		t.Errorf("SSHUserFor(special) = %q, want %q", got, "root")
	}

	// Fallback to global default.
	if got := cfg.SSHUserFor("unknown"); got != "forge" {
		t.Errorf("SSHUserFor(unknown) = %q, want %q", got, "forge")
	}

	// Empty string in map should fall back.
	cfg.ServerUsers["empty"] = ""
	if got := cfg.SSHUserFor("empty"); got != "forge" {
		t.Errorf("SSHUserFor(empty) = %q, want %q (should fallback)", got, "forge")
	}
}

func TestSaveCreatesDirectoryAndFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "deep", "nested", "config.toml")

	cfg := Default()
	if err := cfg.SaveTo(path); err != nil {
		t.Fatalf("SaveTo: %v", err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("File not created: %v", err)
	}
}

func TestLoadFromInvalidTOML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.toml")
	if err := os.WriteFile(path, []byte("{{invalid toml"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := LoadFrom(path)
	if err == nil {
		t.Fatal("Expected error for invalid TOML, got nil")
	}
}
