package cmd

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/PeterHiroshi/cfmon/internal/config"
)

func TestHealthCmd_MissingAccountID(t *testing.T) {
	resetGlobalFlags()

	// Create temp config with no default account
	tmpDir := t.TempDir()
	cfgFile = filepath.Join(tmpDir, "config.yaml")
	cfg := &config.Config{Token: "test-token"}
	err := config.Save(cfgFile, cfg)
	if err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}

	// Execute health without account ID
	rootCmd.SetArgs([]string{"health"})

	err = rootCmd.Execute()
	if err == nil {
		t.Fatal("health command error = nil, want error for missing account ID")
	}

	if !strings.Contains(err.Error(), "no account ID provided") && !strings.Contains(err.Error(), "no default account") {
		t.Errorf("error message = %q, should mention missing account ID or default account", err.Error())
	}
}

func TestHealthCmd_WithDefaultAccount(t *testing.T) {
	resetGlobalFlags()

	// Create temp config with default account
	tmpDir := t.TempDir()
	cfgFile = filepath.Join(tmpDir, "config.yaml")
	cfg := &config.Config{
		Token:            "test-token",
		DefaultAccountID: "default-account-123",
	}
	err := config.Save(cfgFile, cfg)
	if err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}

	// Note: This will fail because we can't actually connect to the API
	// But it verifies that the command accepts the default account
	rootCmd.SetArgs([]string{"health"})

	err = rootCmd.Execute()
	// Should fail on API call, not on missing account
	if err != nil && (strings.Contains(err.Error(), "no account ID provided") || strings.Contains(err.Error(), "no default account")) {
		t.Errorf("should not fail with account ID error when default is set, got: %v", err)
	}
}

func TestHealthCmd_WithExplicitAccount(t *testing.T) {
	resetGlobalFlags()

	// Create temp config
	tmpDir := t.TempDir()
	cfgFile = filepath.Join(tmpDir, "config.yaml")
	cfg := &config.Config{Token: "test-token"}
	err := config.Save(cfgFile, cfg)
	if err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}

	// Execute health with explicit account ID
	rootCmd.SetArgs([]string{"health", "explicit-account-456"})

	err = rootCmd.Execute()
	// Should fail on API call, not on missing account
	if err != nil && (strings.Contains(err.Error(), "no account ID provided") || strings.Contains(err.Error(), "no default account")) {
		t.Errorf("should not fail with account ID error when explicitly provided, got: %v", err)
	}
}

func TestHealthCmd_MissingToken(t *testing.T) {
	resetGlobalFlags()

	// Create temp directory with no config
	tmpDir := t.TempDir()
	cfgFile = filepath.Join(tmpDir, "nonexistent.yaml")

	// Execute health command without token
	rootCmd.SetArgs([]string{"health", "test-account"})

	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("health command error = nil, want error for missing token")
	}

	if !strings.Contains(err.Error(), "token") {
		t.Errorf("error message = %q, should mention missing token", err.Error())
	}
}

func TestHealthCmd_TooManyArgs(t *testing.T) {
	resetGlobalFlags()

	tmpDir := t.TempDir()
	cfgFile = filepath.Join(tmpDir, "config.yaml")

	// Execute health with too many arguments
	rootCmd.SetArgs([]string{"health", "account1", "account2"})

	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("health command error = nil, want error for too many arguments")
	}
}

func TestHealthCmd_JSONFormat(t *testing.T) {
	resetGlobalFlags()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	cfgFile = configPath

	// Save token to config
	cfg := &config.Config{Token: "test-token"}
	err := config.Save(configPath, cfg)
	if err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}

	// Test JSON format flag parsing
	rootCmd.SetArgs([]string{"health", "test-account", "--output", "json"})

	err = rootCmd.ParseFlags([]string{"--output", "json"})
	if err != nil {
		t.Fatalf("ParseFlags() error = %v", err)
	}

	if getOutputFormat() != "json" {
		t.Errorf("output format = %q, want json", getOutputFormat())
	}
}

func TestHealthCmd_CommandRegistered(t *testing.T) {
	// Verify health command is registered
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "health" {
			found = true
			break
		}
	}

	if !found {
		t.Error("health command not registered with root command")
	}
}

func TestHealthCmd_ShortDescription(t *testing.T) {
	if healthCmd.Short == "" {
		t.Error("health command should have a short description")
	}
}

func TestGetScoreColor(t *testing.T) {
	tests := []struct {
		status string
		want   string
	}{
		{"excellent", "green"},
		{"good", "green"},
		{"fair", "yellow"},
		{"poor", "red"},
		{"critical", "red"},
		{"unknown", "white"},
	}

	for _, tt := range tests {
		got := getScoreColor(tt.status)
		if got != tt.want {
			t.Errorf("getScoreColor(%q) = %q, want %q", tt.status, got, tt.want)
		}
	}
}

func TestGetPointsColor(t *testing.T) {
	tests := []struct {
		points int
		max    int
		want   string
	}{
		{100, 100, "green"}, // 100%
		{90, 100, "green"},  // 90%
		{89, 100, "yellow"}, // 89%
		{70, 100, "yellow"}, // 70%
		{69, 100, "red"},    // 69%
		{50, 100, "red"},    // 50%
		{0, 100, "red"},     // 0%
		{30, 30, "green"},   // 100%
		{27, 30, "green"},   // 90%
		{21, 30, "yellow"},  // 70%
		{20, 30, "red"},     // 66%
	}

	for _, tt := range tests {
		got := getPointsColor(tt.points, tt.max)
		if got != tt.want {
			percent := float64(tt.points) / float64(tt.max) * 100
			t.Errorf("getPointsColor(%d, %d) [%.0f%%] = %q, want %q", tt.points, tt.max, percent, got, tt.want)
		}
	}
}
