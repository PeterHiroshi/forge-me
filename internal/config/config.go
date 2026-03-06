package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config holds the application configuration
type Config struct {
	Token           string `yaml:"token"`
	APIEndpoint     string `yaml:"api_endpoint,omitempty"`
	DefaultFormat   string `yaml:"default_format,omitempty"`
	DefaultAccountID string `yaml:"default_account_id,omitempty"`
}

// New creates a new Config instance
func New() *Config {
	return &Config{}
}

// GetConfigPath returns the default configuration file path
func (c *Config) GetConfigPath() string {
	// Check environment variable first
	if path := os.Getenv("CFMON_CONFIG"); path != "" {
		return path
	}

	// Use default path
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".cfmon", "config.yaml")
	}
	return filepath.Join(home, ".cfmon", "config.yaml")
}

// Load reads configuration from the default or specified path
func (c *Config) Load() error {
	path := c.GetConfigPath()

	// If file doesn't exist, return nil (empty config is valid)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(data, c); err != nil {
		return err
	}

	return nil
}

// Load reads configuration from the specified path
func Load(path string) (*Config, error) {
	cfg := &Config{}

	// If file doesn't exist, return empty config
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Save writes configuration to the specified path
func Save(path string, cfg *Config) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}
