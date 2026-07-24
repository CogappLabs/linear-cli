package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config is persisted to ~/.config/linear-cli/config.json.
type Config struct {
	APIKey      string `json:"apiKey,omitempty"`
	DefaultTeam string `json:"defaultTeam,omitempty"` // team key, e.g. "ENG"
}

func dir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "linear-cli"), nil
}

func path() (string, error) {
	d, err := dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, "config.json"), nil
}

// Load reads the config file. A missing file yields an empty Config, not an error.
func Load() (*Config, error) {
	p, err := path()
	if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(p)
	if os.IsNotExist(err) {
		return &Config{}, nil
	}
	if err != nil {
		return nil, err
	}
	var c Config
	if err := json.Unmarshal(b, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

// Save writes the config file with 0600 perms (it holds the API key).
func Save(c *Config) error {
	d, err := dir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(d, 0o700); err != nil {
		return err
	}
	p := filepath.Join(d, "config.json")
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, b, 0o600)
}

// ResolveAPIKey returns the key from LINEAR_API_KEY if set, else the config file.
func ResolveAPIKey() string {
	if v := os.Getenv("LINEAR_API_KEY"); v != "" {
		return v
	}
	c, err := Load()
	if err != nil {
		return ""
	}
	return c.APIKey
}
