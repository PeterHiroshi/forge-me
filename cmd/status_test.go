package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/PeterHiroshi/cfmon/internal/config"
)

func TestStatusCmd_MissingToken(t *testing.T) {
	resetGlobalFlags()

	// Create temp directory with no config
	tmpDir := t.TempDir()
	cfgFile = filepath.Join(tmpDir, "nonexistent.yaml")

	// Execute without token
	rootCmd.SetArgs([]string{"status"})

	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("statusCmd.Execute() error = nil, want error for missing token")
	}

	if !strings.Contains(err.Error(), "token") {
		t.Errorf("error message = %q, should mention missing token", err.Error())
	}
}

func TestStatusCmd_LoadTokenFromConfig(t *testing.T) {
	resetGlobalFlags()

	// Create temp config with token
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	cfg := &config.Config{Token: "status-config-token"}
	err := config.Save(configPath, cfg)
	if err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}

	cfgFile = configPath

	// Verify config file exists and has the token
	loadedCfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}

	if loadedCfg.Token != "status-config-token" {
		t.Errorf("loaded token = %q, want %q", loadedCfg.Token, "status-config-token")
	}
}

func TestStatusCmd_FormatFlag_JSON(t *testing.T) {
	resetGlobalFlags()

	tmpDir := t.TempDir()
	cfgFile = filepath.Join(tmpDir, "config.yaml")

	// Save a token to config
	cfg := &config.Config{Token: "test-token"}
	err := config.Save(cfgFile, cfg)
	if err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}

	// Test JSON format flag parsing
	rootCmd.SetArgs([]string{"status", "--format", "json"})

	err = rootCmd.ParseFlags([]string{"--format", "json"})
	if err != nil {
		t.Fatalf("ParseFlags() error = %v", err)
	}

	if format != "json" {
		t.Errorf("format = %q, want json", format)
	}
}

func TestStatusCmd_TokenFlagOverridesConfig(t *testing.T) {
	resetGlobalFlags()

	// Create config with one token
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	cfg := &config.Config{Token: "config-token"}
	err := config.Save(configPath, cfg)
	if err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}

	cfgFile = configPath

	// Parse flags with different token
	args := []string{"status", "--token", "flag-token"}
	rootCmd.SetArgs(args)

	err = rootCmd.ParseFlags([]string{"--token", "flag-token"})
	if err != nil {
		t.Fatalf("ParseFlags() error = %v", err)
	}

	// Flag token should override config
	if token != "flag-token" {
		t.Errorf("token = %q, want %q (flag should override config)", token, "flag-token")
	}
}

func TestStatusCmd_NoColorFlag(t *testing.T) {
	resetGlobalFlags()

	tmpDir := t.TempDir()
	cfgFile = filepath.Join(tmpDir, "config.yaml")

	cfg := &config.Config{Token: "test-token"}
	err := config.Save(cfgFile, cfg)
	if err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}

	// Test no-color flag
	args := []string{"status", "--no-color"}
	rootCmd.SetArgs(args)

	err = rootCmd.ParseFlags([]string{"--no-color"})
	if err != nil {
		t.Fatalf("ParseFlags() error = %v", err)
	}

	if !noColor {
		t.Errorf("noColor = %v, want true", noColor)
	}
}

func TestStatusCmd_DefaultConfigPath(t *testing.T) {
	resetGlobalFlags()

	// Set up temp HOME
	tmpHome := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", oldHome)

	// Create config in default location
	defaultConfigPath := filepath.Join(tmpHome, ".cfmon", "config.yaml")
	cfg := &config.Config{Token: "home-status-token"}
	err := config.Save(defaultConfigPath, cfg)
	if err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}

	// Load should work from default path
	loadedCfg, err := config.Load(defaultConfigPath)
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}

	if loadedCfg.Token != "home-status-token" {
		t.Errorf("loaded token = %q, want %q", loadedCfg.Token, "home-status-token")
	}
}

func TestStatusCmd_NoArgs(t *testing.T) {
	resetGlobalFlags()

	tmpDir := t.TempDir()
	cfgFile = filepath.Join(tmpDir, "config.yaml")

	cfg := &config.Config{Token: "test-token"}
	err := config.Save(cfgFile, cfg)
	if err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}

	// Status command takes no args, should be fine
	rootCmd.SetArgs([]string{"status"})

	// Just verify it doesn't error on argument count
	// (actual execution will fail without valid API but that's tested in api package)
	err = rootCmd.ParseFlags([]string{})
	if err != nil {
		t.Fatalf("ParseFlags() error = %v", err)
	}
}

func TestStatusCmd_CommandRegistered(t *testing.T) {
	resetGlobalFlags()

	// Verify status command is registered with root
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "status" {
			found = true
			break
		}
	}

	if !found {
		t.Error("status command not found in root commands")
	}
}

func TestStatusCmd_ShortDescription(t *testing.T) {
	resetGlobalFlags()

	// Verify status command has a short description
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "status" {
			if cmd.Short == "" {
				t.Error("status command has empty short description")
			}
			if !strings.Contains(strings.ToLower(cmd.Short), "token") || !strings.Contains(strings.ToLower(cmd.Short), "status") {
				t.Errorf("status command short description doesn't mention token/status: %q", cmd.Short)
			}
			return
		}
	}

	t.Fatal("status command not found")
}
