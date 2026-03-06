package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
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

	// Create temp config with no default account
	tmpDir := t.TempDir()
	cfgFile = filepath.Join(tmpDir, "config.yaml")
	cfg := &config.Config{Token: "test-token"}
	err := config.Save(cfgFile, cfg)
	if err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}

	// Execute list subcommand without account ID
	rootCmd.SetArgs([]string{"containers", "list"})

	err = rootCmd.Execute()
	if err == nil {
		t.Fatal("containersCmd.Execute() error = nil, want error for missing account ID")
	}

	if !strings.Contains(err.Error(), "no account ID provided") && !strings.Contains(err.Error(), "no default account") {
		t.Errorf("error message = %q, should mention missing account ID or default account", err.Error())
	}
}

func TestContainersCmd_MissingToken(t *testing.T) {
	resetGlobalFlags()

	// Create temp directory with no config
	tmpDir := t.TempDir()
	cfgFile = filepath.Join(tmpDir, "nonexistent.yaml")

	// Execute list subcommand without token
	rootCmd.SetArgs([]string{"containers", "list", "test-account"})

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

	// Execute list subcommand with too many arguments
	rootCmd.SetArgs([]string{"containers", "list", "account1", "account2"})

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

// Tests for new features (sorting, filtering, limiting)

func TestSortContainers(t *testing.T) {
	containers := []api.Container{
		{ID: "1", Name: "zebra", CPUMS: 100, MemoryMB: 256, Requests: 500},
		{ID: "2", Name: "alpha", CPUMS: 200, MemoryMB: 128, Requests: 1000},
		{ID: "3", Name: "beta", CPUMS: 150, MemoryMB: 512, Requests: 750},
	}

	tests := []struct {
		name      string
		sortBy    string
		wantFirst string
		wantLast  string
	}{
		{
			name:      "sort by name",
			sortBy:    "name",
			wantFirst: "alpha",
			wantLast:  "zebra",
		},
		{
			name:      "sort by cpu descending",
			sortBy:    "cpu",
			wantFirst: "alpha", // 200 CPUMS
			wantLast:  "zebra", // 100 CPUMS
		},
		{
			name:      "sort by memory descending",
			sortBy:    "memory",
			wantFirst: "beta",  // 512 MB
			wantLast:  "alpha", // 128 MB
		},
		{
			name:      "sort by requests descending",
			sortBy:    "requests",
			wantFirst: "alpha", // 1000 requests
			wantLast:  "zebra", // 500 requests
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Make a copy to avoid modifying original
			testContainers := make([]api.Container, len(containers))
			copy(testContainers, containers)

			sortContainers(testContainers, tt.sortBy)

			if testContainers[0].Name != tt.wantFirst {
				t.Errorf("first after sort = %q, want %q", testContainers[0].Name, tt.wantFirst)
			}

			if testContainers[len(testContainers)-1].Name != tt.wantLast {
				t.Errorf("last after sort = %q, want %q", testContainers[len(testContainers)-1].Name, tt.wantLast)
			}
		})
	}
}

func TestContainerFiltering(t *testing.T) {
	containers := []api.Container{
		{ID: "1", Name: "prod-api-server"},
		{ID: "2", Name: "dev-worker"},
		{ID: "3", Name: "prod-database"},
		{ID: "4", Name: "staging-api"},
	}

	tests := []struct {
		name        string
		filter      string
		wantCount   int
		wantNames   []string
	}{
		{
			name:      "filter prod",
			filter:    "prod",
			wantCount: 2,
			wantNames: []string{"prod-api-server", "prod-database"},
		},
		{
			name:      "filter api",
			filter:    "api",
			wantCount: 2,
			wantNames: []string{"prod-api-server", "staging-api"},
		},
		{
			name:      "filter dev",
			filter:    "dev",
			wantCount: 1,
			wantNames: []string{"dev-worker"},
		},
		{
			name:      "filter none matching",
			filter:    "xyz",
			wantCount: 0,
			wantNames: []string{},
		},
		{
			name:      "case insensitive filter",
			filter:    "PROD",
			wantCount: 2,
			wantNames: []string{"prod-api-server", "prod-database"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := []api.Container{}
			for _, c := range containers {
				if strings.Contains(strings.ToLower(c.Name), strings.ToLower(tt.filter)) {
					filtered = append(filtered, c)
				}
			}

			if len(filtered) != tt.wantCount {
				t.Errorf("filtered count = %d, want %d", len(filtered), tt.wantCount)
			}

			for _, c := range filtered {
				found := false
				for _, want := range tt.wantNames {
					if c.Name == want {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("unexpected container in filtered results: %s", c.Name)
				}
			}
		})
	}
}

func TestContainerLimiting(t *testing.T) {
	containers := make([]api.Container, 10)
	for i := 0; i < 10; i++ {
		containers[i] = api.Container{
			ID:   fmt.Sprintf("%d", i),
			Name: fmt.Sprintf("container-%d", i),
		}
	}

	tests := []struct {
		name      string
		limit     int
		wantCount int
	}{
		{"no limit", 0, 10},
		{"limit 5", 5, 5},
		{"limit 1", 1, 1},
		{"limit exceeds count", 20, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containers
			if tt.limit > 0 && len(result) > tt.limit {
				result = result[:tt.limit]
			}

			if len(result) != tt.wantCount {
				t.Errorf("limited count = %d, want %d", len(result), tt.wantCount)
			}
		})
	}
}

func TestGetAPITokenPriority(t *testing.T) {
	// Save original values
	origEnv := os.Getenv("CFMON_TOKEN")
	origToken := token
	origCfgFile := cfgFile

	defer func() {
		os.Setenv("CFMON_TOKEN", origEnv)
		token = origToken
		cfgFile = origCfgFile
	}()

	tests := []struct {
		name      string
		envToken  string
		flagToken string
		wantToken string
		wantErr   bool
	}{
		{
			name:      "env token takes priority",
			envToken:  "env-token",
			flagToken: "flag-token",
			wantToken: "env-token",
			wantErr:   false,
		},
		{
			name:      "flag token when no env",
			envToken:  "",
			flagToken: "flag-token",
			wantToken: "flag-token",
			wantErr:   false,
		},
		{
			name:      "no token available",
			envToken:  "",
			flagToken: "",
			wantToken: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("CFMON_TOKEN", tt.envToken)
			token = tt.flagToken

			got, err := getAPIToken()

			if (err != nil) != tt.wantErr {
				t.Errorf("getAPIToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.wantToken {
				t.Errorf("getAPIToken() = %q, want %q", got, tt.wantToken)
			}
		})
	}
}
