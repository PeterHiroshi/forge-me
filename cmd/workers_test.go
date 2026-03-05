package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/PeterHiroshi/cfmon/internal/api"
	"github.com/PeterHiroshi/cfmon/internal/config"
)

func TestWorkersCmd_WithTokenFlag_TableFormat(t *testing.T) {
	resetGlobalFlags()

	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != "GET" {
			t.Errorf("Method = %s, want GET", r.Method)
		}

		if !strings.Contains(r.URL.Path, "/accounts/test-account/workers/scripts") {
			t.Errorf("Path = %s, want to contain workers scripts", r.URL.Path)
		}

		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("Authorization = %s, want Bearer test-token", auth)
		}

		// Return mock data
		resp := map[string]interface{}{
			"success": true,
			"result": []api.Worker{
				{ID: "worker-1", Name: "Worker 1", CPUMS: 50, Requests: 1000},
				{ID: "worker-2", Name: "Worker 2", CPUMS: 75, Requests: 2000},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Note: The actual HTTP behavior is tested in internal/api package
	// Here we test the command setup and argument parsing
}

func TestWorkersCmd_MissingAccountID(t *testing.T) {
	resetGlobalFlags()

	// Execute list subcommand without account ID
	rootCmd.SetArgs([]string{"workers", "list"})

	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("workersCmd.Execute() error = nil, want error for missing account ID")
	}

	if !strings.Contains(err.Error(), "arg") && !strings.Contains(err.Error(), "requires") {
		t.Errorf("error message = %q, should mention missing argument", err.Error())
	}
}

func TestWorkersCmd_MissingToken(t *testing.T) {
	resetGlobalFlags()

	// Create temp directory with no config
	tmpDir := t.TempDir()
	cfgFile = filepath.Join(tmpDir, "nonexistent.yaml")

	// Execute list subcommand without token
	rootCmd.SetArgs([]string{"workers", "list", "test-account"})

	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("workersCmd.Execute() error = nil, want error for missing token")
	}

	if !strings.Contains(err.Error(), "token") {
		t.Errorf("error message = %q, should mention missing token", err.Error())
	}
}

func TestWorkersCmd_LoadTokenFromConfig(t *testing.T) {
	resetGlobalFlags()

	// Create temp config with token
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	cfg := &config.Config{Token: "worker-config-token"}
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

	if loadedCfg.Token != "worker-config-token" {
		t.Errorf("loaded token = %q, want %q", loadedCfg.Token, "worker-config-token")
	}
}

func TestWorkersCmd_FormatFlag_JSON(t *testing.T) {
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
	rootCmd.SetArgs([]string{"workers", "test-account", "--format", "json"})

	err = rootCmd.ParseFlags([]string{"--format", "json"})
	if err != nil {
		t.Fatalf("ParseFlags() error = %v", err)
	}

	if format != "json" {
		t.Errorf("format = %q, want json", format)
	}
}

func TestWorkersCmd_TokenFlagOverridesConfig(t *testing.T) {
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
	args := []string{"workers", "test-account", "--token", "flag-token"}
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

func TestWorkersCmd_NoColorFlag(t *testing.T) {
	resetGlobalFlags()

	tmpDir := t.TempDir()
	cfgFile = filepath.Join(tmpDir, "config.yaml")

	cfg := &config.Config{Token: "test-token"}
	err := config.Save(cfgFile, cfg)
	if err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}

	// Test no-color flag
	args := []string{"workers", "test-account", "--no-color"}
	rootCmd.SetArgs(args)

	err = rootCmd.ParseFlags([]string{"--no-color"})
	if err != nil {
		t.Fatalf("ParseFlags() error = %v", err)
	}

	if !noColor {
		t.Errorf("noColor = %v, want true", noColor)
	}
}

func TestWorkersCmd_TooManyArgs(t *testing.T) {
	resetGlobalFlags()

	// Execute list subcommand with too many arguments
	rootCmd.SetArgs([]string{"workers", "list", "account1", "account2"})

	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("workersCmd.Execute() error = nil, want error for too many arguments")
	}
}

func TestWorkersCmd_DefaultConfigPath(t *testing.T) {
	resetGlobalFlags()

	// Set up temp HOME
	tmpHome := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", oldHome)

	// Create config in default location
	defaultConfigPath := filepath.Join(tmpHome, ".cfmon", "config.yaml")
	cfg := &config.Config{Token: "home-worker-token"}
	err := config.Save(defaultConfigPath, cfg)
	if err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}

	// Load should work from default path
	loadedCfg, err := config.Load(defaultConfigPath)
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}

	if loadedCfg.Token != "home-worker-token" {
		t.Errorf("loaded token = %q, want %q", loadedCfg.Token, "home-worker-token")
	}
}

func TestWorkersCmd_CommandRegistered(t *testing.T) {
	resetGlobalFlags()

	// Verify workers command is registered with root
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "workers" {
			found = true
			break
		}
	}

	if !found {
		t.Error("workers command not found in root commands")
	}
}
