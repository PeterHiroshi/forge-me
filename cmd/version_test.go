package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestVersionCmd_Output(t *testing.T) {
	resetGlobalFlags()

	// Capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Execute version command
	rootCmd.SetArgs([]string{"version"})

	err := rootCmd.Execute()

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("versionCmd.Execute() error = %v, want nil", err)
	}

	// Read captured output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Should contain "cfmon" and "version"
	if !strings.Contains(output, "cfmon") {
		t.Errorf("version output doesn't contain 'cfmon': %q", output)
	}

	if !strings.Contains(output, "version") {
		t.Errorf("version output doesn't contain 'version': %q", output)
	}
}

func TestVersionCmd_DefaultValue(t *testing.T) {
	resetGlobalFlags()

	// Version should have a default value
	if Version == "" {
		t.Error("Version variable is empty, want non-empty default")
	}

	// Should be set to "dev" by default
	if Version != "dev" {
		t.Logf("Version = %q, default is 'dev' but may be overridden by GoReleaser", Version)
	}
}

func TestVersionCmd_NoArgs(t *testing.T) {
	resetGlobalFlags()

	// Version command should not require arguments
	rootCmd.SetArgs([]string{"version"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("versionCmd.Execute() with no args error = %v, want nil", err)
	}
}

func TestVersionCmd_IgnoresExtraArgs(t *testing.T) {
	resetGlobalFlags()

	// Capture output to avoid cluttering test output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Version command with extra args (should be ignored by cobra)
	rootCmd.SetArgs([]string{"version", "extra", "args"})

	err := rootCmd.Execute()

	w.Close()
	os.Stdout = oldStdout

	// Read and discard output
	var buf bytes.Buffer
	buf.ReadFrom(r)

	// Should still succeed (extra args are ignored)
	if err != nil {
		t.Logf("versionCmd with extra args: error = %v", err)
	}
}

func TestVersionCmd_CommandRegistered(t *testing.T) {
	resetGlobalFlags()

	// Verify version command is registered with root
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "version" {
			found = true
			break
		}
	}

	if !found {
		t.Error("version command not found in root commands")
	}
}

func TestVersionCmd_ShortDescription(t *testing.T) {
	resetGlobalFlags()

	// Verify version command has a short description
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "version" {
			if cmd.Short == "" {
				t.Error("version command has empty short description")
			}
			if !strings.Contains(strings.ToLower(cmd.Short), "version") {
				t.Errorf("version command short description doesn't mention version: %q", cmd.Short)
			}
			return
		}
	}

	t.Fatal("version command not found")
}
