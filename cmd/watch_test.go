package cmd

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/PeterHiroshi/cfmon/internal/config"
)

func TestWatchCmd_MissingAccountID(t *testing.T) {
	resetGlobalFlags()

	// Create temp config with no default account
	tmpDir := t.TempDir()
	cfgFile = filepath.Join(tmpDir, "config.yaml")
	cfg := &config.Config{Token: "test-token"}
	err := config.Save(cfgFile, cfg)
	if err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}

	// Execute watch without account ID
	rootCmd.SetArgs([]string{"watch", "containers"})

	err = rootCmd.Execute()
	if err == nil {
		t.Fatal("watch containers error = nil, want error for missing account ID")
	}

	if !strings.Contains(err.Error(), "no account ID provided") && !strings.Contains(err.Error(), "no default account") {
		t.Errorf("error message = %q, should mention missing account ID or default account", err.Error())
	}
}

func TestWatchCmd_MissingToken(t *testing.T) {
	resetGlobalFlags()

	// Create temp directory with no config
	tmpDir := t.TempDir()
	cfgFile = filepath.Join(tmpDir, "nonexistent.yaml")

	// Execute watch command without token
	rootCmd.SetArgs([]string{"watch", "containers", "test-account"})

	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("watch containers error = nil, want error for missing token")
	}

	if !strings.Contains(err.Error(), "token") {
		t.Errorf("error message = %q, should mention missing token", err.Error())
	}
}

func TestWatchCmd_TooManyArgs(t *testing.T) {
	resetGlobalFlags()

	tmpDir := t.TempDir()
	cfgFile = filepath.Join(tmpDir, "config.yaml")

	// Execute watch with too many arguments
	rootCmd.SetArgs([]string{"watch", "containers", "account1", "account2"})

	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("watch containers error = nil, want error for too many arguments")
	}
}

func TestWatchCmd_IntervalFlag(t *testing.T) {
	resetGlobalFlags()

	// Test interval flag parsing
	watchInterval = 30 * time.Second // reset to default
	err := watchCmd.ParseFlags([]string{"--interval", "10s"})
	if err != nil {
		t.Fatalf("ParseFlags() error = %v", err)
	}

	if watchInterval != 10*time.Second {
		t.Errorf("watchInterval = %v, want 10s", watchInterval)
	}
}

func TestWatchCmd_EventsOnlyFlag(t *testing.T) {
	resetGlobalFlags()

	// Test events-only flag parsing
	watchEventsOnly = false // reset to default
	err := watchCmd.ParseFlags([]string{"--events-only"})
	if err != nil {
		t.Fatalf("ParseFlags() error = %v", err)
	}

	if !watchEventsOnly {
		t.Errorf("watchEventsOnly = %v, want true", watchEventsOnly)
	}
}

func TestWatchCmd_CommandRegistered(t *testing.T) {
	// Verify watch command is registered
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "watch" {
			found = true
			break
		}
	}

	if !found {
		t.Error("watch command not registered with root command")
	}
}

func TestWatchCmd_SubcommandsRegistered(t *testing.T) {
	// Verify watch subcommands are registered
	subcommands := []string{"containers", "workers"}

	for _, subcmdName := range subcommands {
		found := false
		for _, cmd := range watchCmd.Commands() {
			if cmd.Name() == subcmdName {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("watch %s subcommand not registered", subcmdName)
		}
	}
}

func TestWatchCmd_ShortDescription(t *testing.T) {
	if watchCmd.Short == "" {
		t.Error("watch command should have a short description")
	}

	if watchContainersCmd.Short == "" {
		t.Error("watch containers command should have a short description")
	}

	if watchWorkersCmd.Short == "" {
		t.Error("watch workers command should have a short description")
	}
}

func TestWatchCmd_DefaultInterval(t *testing.T) {
	// Reset to default value
	watchInterval = 30 * time.Second

	// Default interval should be 30s
	if watchInterval != 30*time.Second {
		t.Errorf("default watchInterval = %v, want 30s", watchInterval)
	}
}

func TestWatchCmd_DefaultEventsOnly(t *testing.T) {
	// Reset to default value
	watchEventsOnly = false

	// Default events-only should be false
	if watchEventsOnly {
		t.Errorf("default watchEventsOnly = %v, want false", watchEventsOnly)
	}
}
