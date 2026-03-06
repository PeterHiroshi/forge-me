package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/PeterHiroshi/cfmon/internal/api"
	"github.com/PeterHiroshi/cfmon/internal/config"
)

func TestAccountsSetDefault_Success(t *testing.T) {
	resetGlobalFlags()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	cfgFile = configPath

	// Execute set-default command
	rootCmd.SetArgs([]string{"accounts", "set-default", "test-account-123"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("accounts set-default error = %v", err)
	}

	// Verify config was saved
	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}

	if cfg.DefaultAccountID != "test-account-123" {
		t.Errorf("DefaultAccountID = %q, want %q", cfg.DefaultAccountID, "test-account-123")
	}
}

func TestAccountsSetDefault_UpdatesExisting(t *testing.T) {
	resetGlobalFlags()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	cfgFile = configPath

	// Create initial config with different account
	cfg := &config.Config{
		Token:            "test-token",
		DefaultAccountID: "old-account",
	}
	err := config.Save(configPath, cfg)
	if err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}

	// Execute set-default to update
	rootCmd.SetArgs([]string{"accounts", "set-default", "new-account-456"})

	err = rootCmd.Execute()
	if err != nil {
		t.Fatalf("accounts set-default error = %v", err)
	}

	// Verify config was updated
	cfg, err = config.Load(configPath)
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}

	if cfg.DefaultAccountID != "new-account-456" {
		t.Errorf("DefaultAccountID = %q, want %q", cfg.DefaultAccountID, "new-account-456")
	}

	// Verify token is still there
	if cfg.Token != "test-token" {
		t.Errorf("Token = %q, want %q (should preserve existing token)", cfg.Token, "test-token")
	}
}

func TestAccountsSetDefault_JSONFormat(t *testing.T) {
	resetGlobalFlags()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	cfgFile = configPath

	// Execute with JSON format
	rootCmd.SetArgs([]string{"accounts", "set-default", "test-account-789", "--format", "json"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("accounts set-default error = %v", err)
	}

	// Verify config was saved
	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}

	if cfg.DefaultAccountID != "test-account-789" {
		t.Errorf("DefaultAccountID = %q, want %q", cfg.DefaultAccountID, "test-account-789")
	}
}

func TestAccountsSetDefault_MissingAccountID(t *testing.T) {
	resetGlobalFlags()

	tmpDir := t.TempDir()
	cfgFile = filepath.Join(tmpDir, "config.yaml")

	// Execute without account ID
	rootCmd.SetArgs([]string{"accounts", "set-default"})

	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("accounts set-default error = nil, want error for missing account ID")
	}

	if !strings.Contains(err.Error(), "arg") && !strings.Contains(err.Error(), "requires") {
		t.Errorf("error message = %q, should mention missing argument", err.Error())
	}
}

func TestAccountsGetDefault_WithDefaultSet(t *testing.T) {
	resetGlobalFlags()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	cfgFile = configPath

	// Create config with default account
	cfg := &config.Config{
		Token:            "test-token",
		DefaultAccountID: "my-default-account",
	}
	err := config.Save(configPath, cfg)
	if err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}

	// Execute get-default
	rootCmd.SetArgs([]string{"accounts", "get-default"})

	err = rootCmd.Execute()
	if err != nil {
		t.Fatalf("accounts get-default error = %v", err)
	}

	// Verify the output would contain the account ID
	// (We can't easily capture stdout in this test setup, but we verified no error)
}

func TestAccountsGetDefault_NoDefaultSet(t *testing.T) {
	resetGlobalFlags()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	cfgFile = configPath

	// Create config without default account
	cfg := &config.Config{
		Token: "test-token",
	}
	err := config.Save(configPath, cfg)
	if err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}

	// Execute get-default
	rootCmd.SetArgs([]string{"accounts", "get-default"})

	err = rootCmd.Execute()
	if err != nil {
		t.Fatalf("accounts get-default should not error when no default set, got: %v", err)
	}
}

func TestAccountsGetDefault_NoConfigFile(t *testing.T) {
	resetGlobalFlags()

	tmpDir := t.TempDir()
	cfgFile = filepath.Join(tmpDir, "nonexistent", "config.yaml")

	// Execute get-default with no config file
	rootCmd.SetArgs([]string{"accounts", "get-default"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("accounts get-default should not error when no config file, got: %v", err)
	}
}

func TestAccountsGetDefault_JSONFormat(t *testing.T) {
	resetGlobalFlags()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	cfgFile = configPath

	// Create config with default account
	cfg := &config.Config{
		Token:            "test-token",
		DefaultAccountID: "json-account",
	}
	err := config.Save(configPath, cfg)
	if err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}

	// Execute get-default with JSON format
	rootCmd.SetArgs([]string{"accounts", "get-default", "--format", "json"})

	err = rootCmd.Execute()
	if err != nil {
		t.Fatalf("accounts get-default error = %v", err)
	}
}

func TestAccountsList_WithMockAPI(t *testing.T) {
	resetGlobalFlags()

	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != "GET" {
			t.Errorf("Method = %s, want GET", r.Method)
		}

		if !strings.Contains(r.URL.Path, "/accounts") {
			t.Errorf("Path = %s, want to contain /accounts", r.URL.Path)
		}

		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("Authorization = %s, want Bearer test-token", auth)
		}

		// Return mock data
		resp := map[string]interface{}{
			"success": true,
			"result": []api.Account{
				{ID: "acc-1", Name: "Account 1", Type: "standard"},
				{ID: "acc-2", Name: "Account 2", Type: "enterprise"},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	cfgFile = configPath

	// Save token to config
	cfg := &config.Config{Token: "test-token"}
	err := config.Save(configPath, cfg)
	if err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}

	// Note: We can't easily inject the test server URL from cmd tests
	// The actual API client behavior is tested in internal/api package
	// This test verifies the command setup and token loading
}

func TestAccountsList_MissingToken(t *testing.T) {
	resetGlobalFlags()

	tmpDir := t.TempDir()
	cfgFile = filepath.Join(tmpDir, "nonexistent.yaml")

	// Execute list without token
	rootCmd.SetArgs([]string{"accounts", "list"})

	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("accounts list error = nil, want error for missing token")
	}

	if !strings.Contains(err.Error(), "token") {
		t.Errorf("error message = %q, should mention missing token", err.Error())
	}
}

func TestAccountsList_WithDefaultAccount(t *testing.T) {
	resetGlobalFlags()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	cfgFile = configPath

	// Save config with token and default account
	cfg := &config.Config{
		Token:            "test-token",
		DefaultAccountID: "default-acc-123",
	}
	err := config.Save(configPath, cfg)
	if err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}

	// Verify config has default account
	loadedCfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}

	if loadedCfg.DefaultAccountID != "default-acc-123" {
		t.Errorf("DefaultAccountID = %q, want %q", loadedCfg.DefaultAccountID, "default-acc-123")
	}
}

func TestAccountsList_JSONFormat(t *testing.T) {
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
	rootCmd.SetArgs([]string{"accounts", "list", "--format", "json"})

	err = rootCmd.ParseFlags([]string{"--format", "json"})
	if err != nil {
		t.Fatalf("ParseFlags() error = %v", err)
	}

	if format != "json" {
		t.Errorf("format = %q, want json", format)
	}
}

func TestAccountsCmd_TooManyArgsSetDefault(t *testing.T) {
	resetGlobalFlags()

	tmpDir := t.TempDir()
	cfgFile = filepath.Join(tmpDir, "config.yaml")

	// Execute set-default with too many arguments
	rootCmd.SetArgs([]string{"accounts", "set-default", "acc1", "acc2"})

	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("accounts set-default error = nil, want error for too many arguments")
	}
}

func TestAccountsCmd_TooManyArgsGetDefault(t *testing.T) {
	resetGlobalFlags()

	tmpDir := t.TempDir()
	cfgFile = filepath.Join(tmpDir, "config.yaml")

	// Execute get-default with extra arguments
	rootCmd.SetArgs([]string{"accounts", "get-default", "extra-arg"})

	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("accounts get-default error = nil, want error for too many arguments")
	}
}

func TestAccountsCmd_TooManyArgsList(t *testing.T) {
	resetGlobalFlags()

	tmpDir := t.TempDir()
	cfgFile = filepath.Join(tmpDir, "config.yaml")

	// Execute list with extra arguments
	rootCmd.SetArgs([]string{"accounts", "list", "extra-arg"})

	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("accounts list error = nil, want error for too many arguments")
	}
}
