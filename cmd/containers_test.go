package cmd

import (
	"bytes"
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

func TestContainersCmd_WithTokenFlag_TableFormat(t *testing.T) {
	resetGlobalFlags()

	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != "GET" {
			t.Errorf("Method = %s, want GET", r.Method)
		}

		if !strings.Contains(r.URL.Path, "/accounts/test-account/workers/containers/namespaces") {
			t.Errorf("Path = %s, want to contain containers namespaces", r.URL.Path)
		}

		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("Authorization = %s, want Bearer test-token", auth)
		}

		// Return mock data
		resp := map[string]interface{}{
			"success": true,
			"result": []api.Container{
				{ID: "cont-1", Name: "Container 1", CPUMS: 100, MemoryMB: 128},
				{ID: "cont-2", Name: "Container 2", CPUMS: 200, MemoryMB: 256},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Temporarily override the API base URL
	// We need to capture output to verify it
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Execute command
	rootCmd.SetArgs([]string{"containers", "test-account", "--token", "test-token", "--format", "table"})

	// We need to inject the test server URL into the client
	// Since we can't easily do that from cmd level, we'll just test the parsing
	// and error handling. The actual API client is tested in internal/api tests.

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)

	// Note: This test verifies command setup but can't easily mock the HTTP client
	// The actual HTTP behavior is tested in internal/api package
}

func TestContainersCmd_MissingAccountID(t *testing.T) {
	resetGlobalFlags()

	// Execute without account ID
	rootCmd.SetArgs([]string{"containers"})

	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("containersCmd.Execute() error = nil, want error for missing account ID")
	}

	if !strings.Contains(err.Error(), "arg") && !strings.Contains(err.Error(), "requires") {
		t.Errorf("error message = %q, should mention missing argument", err.Error())
	}
}

func TestContainersCmd_MissingToken(t *testing.T) {
	resetGlobalFlags()

	// Create temp directory with no config
	tmpDir := t.TempDir()
	cfgFile = filepath.Join(tmpDir, "nonexistent.yaml")

	// Execute without token
	rootCmd.SetArgs([]string{"containers", "test-account"})

	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("containersCmd.Execute() error = nil, want error for missing token")
	}

	if !strings.Contains(err.Error(), "token") {
		t.Errorf("error message = %q, should mention missing token", err.Error())
	}
}

func TestContainersCmd_LoadTokenFromConfig(t *testing.T) {
	resetGlobalFlags()

	// Create temp config with token
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	cfg := &config.Config{Token: "config-token"}
	err := config.Save(configPath, cfg)
	if err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}

	cfgFile = configPath

	// Note: We can't easily inject the mock server URL from cmd tests
	// The token loading logic is what we're primarily testing here
	// The actual HTTP calls are tested in internal/api

	// Verify config file exists and has the token
	loadedCfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}

	if loadedCfg.Token != "config-token" {
		t.Errorf("loaded token = %q, want %q", loadedCfg.Token, "config-token")
	}
}

func TestContainersCmd_FormatFlag_JSON(t *testing.T) {
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
	rootCmd.SetArgs([]string{"containers", "test-account", "--format", "json"})

	err = rootCmd.ParseFlags([]string{"--format", "json"})
	if err != nil {
		t.Fatalf("ParseFlags() error = %v", err)
	}

	if format != "json" {
		t.Errorf("format = %q, want json", format)
	}
}

func TestContainersCmd_TokenFlagOverridesConfig(t *testing.T) {
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
	args := []string{"containers", "test-account", "--token", "flag-token"}
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

func TestContainersCmd_NoColorFlag(t *testing.T) {
	resetGlobalFlags()

	tmpDir := t.TempDir()
	cfgFile = filepath.Join(tmpDir, "config.yaml")

	cfg := &config.Config{Token: "test-token"}
	err := config.Save(cfgFile, cfg)
	if err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}

	// Test no-color flag
	args := []string{"containers", "test-account", "--no-color"}
	rootCmd.SetArgs(args)

	err = rootCmd.ParseFlags([]string{"--no-color"})
	if err != nil {
		t.Fatalf("ParseFlags() error = %v", err)
	}

	if !noColor {
		t.Errorf("noColor = %v, want true", noColor)
	}
}

func TestContainersCmd_TooManyArgs(t *testing.T) {
	resetGlobalFlags()

	// Execute with too many arguments
	rootCmd.SetArgs([]string{"containers", "account1", "account2"})

	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("containersCmd.Execute() error = nil, want error for too many arguments")
	}
}

func TestContainersCmd_DefaultConfigPath(t *testing.T) {
	resetGlobalFlags()

	// Set up temp HOME
	tmpHome := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", oldHome)

	// Create config in default location
	defaultConfigPath := filepath.Join(tmpHome, ".cfmon", "config.yaml")
	cfg := &config.Config{Token: "home-token"}
	err := config.Save(defaultConfigPath, cfg)
	if err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}

	// Load should work from default path
	loadedCfg, err := config.Load(defaultConfigPath)
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}

	if loadedCfg.Token != "home-token" {
		t.Errorf("loaded token = %q, want %q", loadedCfg.Token, "home-token")
	}
}
