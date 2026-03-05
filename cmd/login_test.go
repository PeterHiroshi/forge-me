package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/PeterHiroshi/cfmon/internal/config"
)

func TestLoginCmd_Success(t *testing.T) {
	resetGlobalFlags()

	// Create temp directory for config
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Set custom config file
	cfgFile = configPath

	// Execute login command
	rootCmd.SetArgs([]string{"login", "test-token-123"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("loginCmd.Execute() error = %v, want nil", err)
	}

	// Verify config was saved
	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}

	if cfg.Token != "test-token-123" {
		t.Errorf("saved token = %q, want %q", cfg.Token, "test-token-123")
	}

	// Verify file permissions
	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("os.Stat() error = %v", err)
	}

	mode := info.Mode().Perm()
	// Should be 0600 (owner read/write only)
	if mode != 0600 {
		t.Errorf("config file permissions = %o, want 0600", mode)
	}
}

func TestLoginCmd_MissingArgs(t *testing.T) {
	resetGlobalFlags()

	// Create temp directory
	tmpDir := t.TempDir()
	cfgFile = filepath.Join(tmpDir, "config.yaml")

	// Execute login without token argument
	rootCmd.SetArgs([]string{"login"})

	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("loginCmd.Execute() error = nil, want error for missing argument")
	}

	// Error message should mention arguments
	if !strings.Contains(err.Error(), "arg") {
		t.Errorf("error message = %q, want to contain 'arg'", err.Error())
	}
}

func TestLoginCmd_TooManyArgs(t *testing.T) {
	resetGlobalFlags()

	tmpDir := t.TempDir()
	cfgFile = filepath.Join(tmpDir, "config.yaml")

	// Execute login with too many arguments
	rootCmd.SetArgs([]string{"login", "token1", "token2"})

	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("loginCmd.Execute() error = nil, want error for too many arguments")
	}
}

func TestLoginCmd_DefaultConfigPath(t *testing.T) {
	resetGlobalFlags()

	// Don't set cfgFile, let it use default
	// We'll use a temp HOME to avoid affecting real config
	tmpHome := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", oldHome)

	// Execute login command
	rootCmd.SetArgs([]string{"login", "my-secret-token"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("loginCmd.Execute() error = %v, want nil", err)
	}

	// Verify config was saved to default location
	expectedPath := filepath.Join(tmpHome, ".cfmon", "config.yaml")
	cfg, err := config.Load(expectedPath)
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}

	if cfg.Token != "my-secret-token" {
		t.Errorf("saved token = %q, want %q", cfg.Token, "my-secret-token")
	}

	// Verify directory was created
	dirInfo, err := os.Stat(filepath.Join(tmpHome, ".cfmon"))
	if err != nil {
		t.Fatalf("config directory not created: %v", err)
	}

	if !dirInfo.IsDir() {
		t.Error("config path is not a directory")
	}
}

func TestLoginCmd_OverwriteExistingToken(t *testing.T) {
	resetGlobalFlags()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	cfgFile = configPath

	// Save initial token
	initialCfg := &config.Config{Token: "old-token"}
	err := config.Save(configPath, initialCfg)
	if err != nil {
		t.Fatalf("initial Save() error = %v", err)
	}

	// Login with new token
	rootCmd.SetArgs([]string{"login", "new-token"})

	err = rootCmd.Execute()
	if err != nil {
		t.Fatalf("loginCmd.Execute() error = %v, want nil", err)
	}

	// Verify token was overwritten
	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}

	if cfg.Token != "new-token" {
		t.Errorf("saved token = %q, want %q (should overwrite old token)", cfg.Token, "new-token")
	}
}

func TestLoginCmd_EmptyToken(t *testing.T) {
	resetGlobalFlags()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	cfgFile = configPath

	// Login with empty string as token (should still work)
	rootCmd.SetArgs([]string{"login", ""})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("loginCmd.Execute() error = %v, want nil", err)
	}

	// Verify config was saved with empty token
	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}

	if cfg.Token != "" {
		t.Errorf("saved token = %q, want empty string", cfg.Token)
	}
}
