package cmd

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/PeterHiroshi/cfmon/internal/config"
)

func TestTailCmd_MissingArgs(t *testing.T) {
	resetGlobalFlags()

	tmpDir := t.TempDir()
	cfgFile = filepath.Join(tmpDir, "config.yaml")
	cfg := &config.Config{Token: "test-token"}
	config.Save(cfgFile, cfg)

	rootCmd.SetArgs([]string{"tail"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("Expected error for missing worker name")
	}
}

func TestTailCmd_FlagParsing(t *testing.T) {
	resetGlobalFlags()

	tailCommand, _, err := rootCmd.Find([]string{"tail"})
	if err != nil {
		t.Fatalf("tail command not found: %v", err)
	}

	if tailCommand.Use != "tail [account-id] <worker-name>" {
		t.Errorf("Use = %q, want %q", tailCommand.Use, "tail [account-id] <worker-name>")
	}

	flags := []string{"format", "status", "method", "search", "ip", "sample-rate", "header", "since", "max-events", "include-logs", "include-exceptions"}
	for _, name := range flags {
		if tailCommand.Flags().Lookup(name) == nil {
			t.Errorf("Flag --%s not found", name)
		}
	}
}

func TestTailCmd_DefaultAccountFromConfig(t *testing.T) {
	resetGlobalFlags()

	tmpDir := t.TempDir()
	cfgFile = filepath.Join(tmpDir, "config.yaml")
	cfg := &config.Config{
		Token:            "test-token",
		DefaultAccountID: "default-acc-123",
	}
	config.Save(cfgFile, cfg)

	rootCmd.SetArgs([]string{"tail", "my-worker", "--max-events", "1"})
	err := rootCmd.Execute()

	// Should fail with connection error, not argument error
	if err != nil && strings.Contains(err.Error(), "no account ID") {
		t.Errorf("Should use default account, got: %v", err)
	}
}

func TestTailCmd_MissingToken(t *testing.T) {
	resetGlobalFlags()

	tmpDir := t.TempDir()
	cfgFile = filepath.Join(tmpDir, "nonexistent.yaml")

	rootCmd.SetArgs([]string{"tail", "acc-123", "my-worker"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("Expected error for missing token")
	}
	if !strings.Contains(err.Error(), "token") {
		t.Errorf("Error should mention token, got: %v", err)
	}
}

func TestTailCmd_SampleRateValidation(t *testing.T) {
	resetGlobalFlags()

	tmpDir := t.TempDir()
	cfgFile = filepath.Join(tmpDir, "config.yaml")
	cfg := &config.Config{Token: "test-token"}
	config.Save(cfgFile, cfg)

	rootCmd.SetArgs([]string{"tail", "acc-123", "my-worker", "--sample-rate", "2.0"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("Expected error for invalid sample rate")
	}
	if !strings.Contains(err.Error(), "sample rate") {
		t.Errorf("Error should mention sample rate, got: %v", err)
	}
}
