package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

// resetGlobalFlags resets all global variables between tests
func resetGlobalFlags() {
	format = "table"
	token = ""
	cfgFile = ""
	noColor = false
}

func TestExecute(t *testing.T) {
	// Test that Execute doesn't panic with no args
	// We can't easily test Execute() itself since it calls os.Exit on error,
	// but we can test the rootCmd directly
	resetGlobalFlags()

	// Reset rootCmd args
	rootCmd.SetArgs([]string{})

	// Capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// This should succeed (just shows help)
	err := rootCmd.Execute()

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}

	// Read captured output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Should contain help text about the tool or commands
	// Running with no args shows help which should mention Cloudflare or the commands
	if !strings.Contains(output, "Cloudflare") && !strings.Contains(output, "cfmon") {
		t.Errorf("Execute() output doesn't contain expected help text")
	}
}

func TestRootCmd_GlobalFlags(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantFormat string
		wantToken string
		wantNoColor bool
	}{
		{
			name:     "default values",
			args:     []string{},
			wantFormat: "table",
			wantToken: "",
			wantNoColor: false,
		},
		{
			name:     "format flag json",
			args:     []string{"--format", "json"},
			wantFormat: "json",
			wantToken: "",
			wantNoColor: false,
		},
		{
			name:     "token flag",
			args:     []string{"--token", "test-token-123"},
			wantFormat: "table",
			wantToken: "test-token-123",
			wantNoColor: false,
		},
		{
			name:     "no-color flag",
			args:     []string{"--no-color"},
			wantFormat: "table",
			wantToken: "",
			wantNoColor: true,
		},
		{
			name:     "all flags combined",
			args:     []string{"--format", "json", "--token", "my-token", "--no-color", "--config", "/tmp/config.yaml"},
			wantFormat: "json",
			wantToken: "my-token",
			wantNoColor: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetGlobalFlags()

			rootCmd.SetArgs(tt.args)

			// Parse flags
			err := rootCmd.ParseFlags(tt.args)
			if err != nil {
				t.Fatalf("ParseFlags() error = %v", err)
			}

			if format != tt.wantFormat {
				t.Errorf("format = %q, want %q", format, tt.wantFormat)
			}

			if token != tt.wantToken {
				t.Errorf("token = %q, want %q", token, tt.wantToken)
			}

			if noColor != tt.wantNoColor {
				t.Errorf("noColor = %v, want %v", noColor, tt.wantNoColor)
			}
		})
	}
}

func TestRootCmd_Version(t *testing.T) {
	resetGlobalFlags()

	// Test that version is set
	if rootCmd.Version == "" {
		t.Error("rootCmd.Version is empty, want non-empty")
	}

	if rootCmd.Version != "0.1.0" {
		t.Errorf("rootCmd.Version = %q, want %q", rootCmd.Version, "0.1.0")
	}
}

func TestRootCmd_Subcommands(t *testing.T) {
	resetGlobalFlags()

	// Check that expected subcommands are registered
	expectedCommands := []string{
		"login",
		"containers",
		"workers",
		"status",
		"version",
		"ping",
		"completion",
		"help", // cobra adds this automatically
	}

	commands := rootCmd.Commands()
	commandNames := make(map[string]bool)
	for _, cmd := range commands {
		commandNames[cmd.Name()] = true
	}

	for _, expected := range expectedCommands {
		if !commandNames[expected] {
			t.Errorf("expected command %q not found in rootCmd", expected)
		}
	}
}

func TestRootCmd_Use(t *testing.T) {
	resetGlobalFlags()

	if rootCmd.Use != "cfmon" {
		t.Errorf("rootCmd.Use = %q, want %q", rootCmd.Use, "cfmon")
	}
}

func TestRootCmd_ConfigFlag(t *testing.T) {
	resetGlobalFlags()

	args := []string{"--config", "/custom/path/config.yaml"}
	rootCmd.SetArgs(args)

	err := rootCmd.ParseFlags(args)
	if err != nil {
		t.Fatalf("ParseFlags() error = %v", err)
	}

	if cfgFile != "/custom/path/config.yaml" {
		t.Errorf("cfgFile = %q, want %q", cfgFile, "/custom/path/config.yaml")
	}
}
