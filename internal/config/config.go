// Package config reads and writes TOML config files for Phorge.
//
// The config file lives at ~/.config/phorge/config.toml and is compatible
// with the existing Python app's configuration format.
package config

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	toml "github.com/pelletier/go-toml/v2"
)

// Config is the top-level configuration structure.
type Config struct {
	Forge       ForgeConfig       `toml:"forge"`
	Editor      EditorConfig      `toml:"editor"`
	ServerUsers map[string]string `toml:"server_users,omitempty"`
}

// ForgeConfig holds Laravel Forge API settings.
type ForgeConfig struct {
	APIKey  string `toml:"api_key"`
	SSHUser string `toml:"ssh_user"`
}

// EditorConfig holds external editor settings.
type EditorConfig struct {
	Command string `toml:"command"`
}

// Default returns a Config populated with sensible defaults.
func Default() *Config {
	return &Config{
		Forge: ForgeConfig{
			SSHUser: "forge",
		},
		Editor: EditorConfig{
			Command: "vim",
		},
		ServerUsers: make(map[string]string),
	}
}

// DefaultPath returns the platform-appropriate path to the config file.
// On most systems this is ~/.config/phorge/config.toml.
func DefaultPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		// Fall back to HOME/.config on failure.
		home, _ := os.UserHomeDir()
		dir = filepath.Join(home, ".config")
	}
	return filepath.Join(dir, "phorge", "config.toml")
}

// Load reads the config from the default path.
// If the file does not exist, it returns a default Config (no error).
func Load() (*Config, error) {
	return LoadFrom(DefaultPath())
}

// LoadFrom reads the config from the given path.
// If the file does not exist, it returns a default Config (no error).
func LoadFrom(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return Default(), nil
		}
		return nil, err
	}

	cfg := Default()
	if err := toml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	// Ensure the map is never nil after unmarshalling.
	if cfg.ServerUsers == nil {
		cfg.ServerUsers = make(map[string]string)
	}

	return cfg, nil
}

// Save writes the config to the default path.
func (c *Config) Save() error {
	return c.SaveTo(DefaultPath())
}

// SaveTo writes the config to the given path.
// It creates the parent directory with mode 0o700 and the file with mode 0o600.
func (c *Config) SaveTo(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}

	data, err := toml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o600)
}

// SSHUserFor returns the SSH user for a given server.
// It checks the per-server override map first, then falls back to the
// global Forge.SSHUser setting.
func (c *Config) SSHUserFor(serverName string) string {
	if user, ok := c.ServerUsers[serverName]; ok && user != "" {
		return user
	}
	return c.Forge.SSHUser
}
