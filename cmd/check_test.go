package cmd

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/PeterHiroshi/cfmon/internal/config"
)

func TestCheckCmd_CommandRegistered(t *testing.T) {
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "check" {
			found = true
			break
		}
	}
	if !found {
		t.Error("check command not registered with root command")
	}
}

func TestCheckCmd_ShortDescription(t *testing.T) {
	if checkCmd.Short == "" {
		t.Error("check command should have a short description")
	}
}

func TestCheckCmd_MissingAccountID(t *testing.T) {
	resetGlobalFlags()

	tmpDir := t.TempDir()
	cfgFile = filepath.Join(tmpDir, "config.yaml")
	cfg := &config.Config{Token: "test-token"}
	err := config.Save(cfgFile, cfg)
	if err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}

	rootCmd.SetArgs([]string{"check"})

	err = rootCmd.Execute()
	if err == nil {
		t.Fatal("check command error = nil, want error for missing account ID")
	}

	if !strings.Contains(err.Error(), "no account ID provided") && !strings.Contains(err.Error(), "no default account") {
		t.Errorf("error message = %q, should mention missing account ID or default account", err.Error())
	}
}

func TestCheckCmd_WithDefaultAccount(t *testing.T) {
	resetGlobalFlags()

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

	rootCmd.SetArgs([]string{"check"})

	err = rootCmd.Execute()
	// Should fail on API call, not on missing account
	if err != nil && (strings.Contains(err.Error(), "no account ID provided") || strings.Contains(err.Error(), "no default account")) {
		t.Errorf("should not fail with account ID error when default is set, got: %v", err)
	}
}

func TestCheckCmd_WithExplicitAccount(t *testing.T) {
	resetGlobalFlags()

	tmpDir := t.TempDir()
	cfgFile = filepath.Join(tmpDir, "config.yaml")
	cfg := &config.Config{Token: "test-token"}
	err := config.Save(cfgFile, cfg)
	if err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}

	rootCmd.SetArgs([]string{"check", "explicit-account-456"})

	err = rootCmd.Execute()
	// Should fail on API call, not on missing account
	if err != nil && (strings.Contains(err.Error(), "no account ID provided") || strings.Contains(err.Error(), "no default account")) {
		t.Errorf("should not fail with account ID error when explicitly provided, got: %v", err)
	}
}

func TestCheckCmd_MissingToken(t *testing.T) {
	resetGlobalFlags()

	tmpDir := t.TempDir()
	cfgFile = filepath.Join(tmpDir, "nonexistent.yaml")

	rootCmd.SetArgs([]string{"check", "test-account"})

	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("check command error = nil, want error for missing token")
	}

	if !strings.Contains(err.Error(), "token") {
		t.Errorf("error message = %q, should mention missing token", err.Error())
	}
}

func TestCheckCmd_TooManyArgs(t *testing.T) {
	resetGlobalFlags()

	tmpDir := t.TempDir()
	cfgFile = filepath.Join(tmpDir, "config.yaml")

	rootCmd.SetArgs([]string{"check", "account1", "account2"})

	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("check command error = nil, want error for too many arguments")
	}
}

func TestCheckCmd_ThresholdFlags(t *testing.T) {
	// Verify threshold flags are registered
	f := checkCmd.Flags()

	cpuFlag := f.Lookup("cpu-threshold")
	if cpuFlag == nil {
		t.Fatal("--cpu-threshold flag not registered")
	}
	if cpuFlag.DefValue != "80" {
		t.Errorf("--cpu-threshold default = %q, want %q", cpuFlag.DefValue, "80")
	}

	memFlag := f.Lookup("memory-threshold")
	if memFlag == nil {
		t.Fatal("--memory-threshold flag not registered")
	}
	if memFlag.DefValue != "85" {
		t.Errorf("--memory-threshold default = %q, want %q", memFlag.DefValue, "85")
	}

	errFlag := f.Lookup("error-threshold")
	if errFlag == nil {
		t.Fatal("--error-threshold flag not registered")
	}
	if errFlag.DefValue != "2" {
		t.Errorf("--error-threshold default = %q, want %q", errFlag.DefValue, "2")
	}
}

func TestCheckCmd_JSONFormat(t *testing.T) {
	resetGlobalFlags()

	tmpDir := t.TempDir()
	cfgFile = filepath.Join(tmpDir, "config.yaml")
	cfg := &config.Config{Token: "test-token"}
	err := config.Save(cfgFile, cfg)
	if err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}

	rootCmd.SetArgs([]string{"check", "test-account", "--output", "json"})

	err = rootCmd.ParseFlags([]string{"--output", "json"})
	if err != nil {
		t.Fatalf("ParseFlags() error = %v", err)
	}

	if getOutputFormat() != "json" {
		t.Errorf("output format = %q, want json", getOutputFormat())
	}
}
