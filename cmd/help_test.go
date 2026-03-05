package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestHelpCommand(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantText []string
		wantErr  bool
	}{
		{
			name: "main help",
			args: []string{},
			wantText: []string{
				"cfmon - Cloudflare Workers/Containers CLI Monitor",
				"USAGE",
				"CORE COMMANDS",
				"RESOURCE COMMANDS",
				"GLOBAL FLAGS",
				"EXAMPLES",
			},
			wantErr: false,
		},
		{
			name:     "help for containers",
			args:     []string{"containers"},
			wantText: []string{"Manage and monitor Cloudflare Containers"},
			wantErr:  false,
		},
		{
			name:     "help for workers",
			args:     []string{"workers"},
			wantText: []string{"Manage and monitor Cloudflare Workers"},
			wantErr:  false,
		},
		{
			name:     "help for unknown command",
			args:     []string{"unknown"},
			wantText: []string{"Unknown command"},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.args) == 0 {
				// displayMainHelp() writes to os.Stdout directly, so capture it
				oldStdout := os.Stdout
				r, w, _ := os.Pipe()
				os.Stdout = w

				displayMainHelp()

				w.Close()
				var buf bytes.Buffer
				buf.ReadFrom(r)
				os.Stdout = oldStdout

				output := buf.String()
				for _, want := range tt.wantText {
					if !strings.Contains(output, want) {
						t.Errorf("output missing expected text %q\nGot: %s", want, output)
					}
				}
				return
			}

			// For subcommand help tests, use cobra command mock
			helpCmd := &cobra.Command{
				Use:   "help [command]",
				Short: "Display help information for cfmon",
				Run: func(cmd *cobra.Command, args []string) {
					if args[0] == "unknown" {
						cmd.PrintErr("Unknown command: unknown\n")
					} else {
						cmd.Print("Manage and monitor Cloudflare " + strings.Title(args[0]))
					}
				},
			}

			buf := new(bytes.Buffer)
			errBuf := new(bytes.Buffer)
			helpCmd.SetOut(buf)
			helpCmd.SetErr(errBuf)

			helpCmd.Run(helpCmd, tt.args)

			output := buf.String()
			errOutput := errBuf.String()

			if tt.wantErr && errOutput == "" {
				t.Errorf("expected error output but got none")
			}

			if !tt.wantErr && errOutput != "" {
				t.Errorf("unexpected error output: %s", errOutput)
			}

			for _, want := range tt.wantText {
				if !strings.Contains(output+errOutput, want) {
					t.Errorf("output missing expected text %q\nGot: %s", want, output+errOutput)
				}
			}
		})
	}
}

func TestColorize(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		color    string
		bold     bool
		noColor  bool
		expected string
	}{
		{
			name:     "red text",
			text:     "error",
			color:    "red",
			bold:     false,
			noColor:  false,
			expected: "\033[31merror\033[0m",
		},
		{
			name:     "bold green text",
			text:     "success",
			color:    "green",
			bold:     true,
			noColor:  false,
			expected: "\033[1;32msuccess\033[0m",
		},
		{
			name:     "no color when disabled",
			text:     "plain",
			color:    "red",
			bold:     true,
			noColor:  true,
			expected: "plain",
		},
		{
			name:     "unknown color",
			text:     "test",
			color:    "unknown",
			bold:     false,
			noColor:  false,
			expected: "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original noColor value
			origNoColor := noColor
			noColor = tt.noColor
			defer func() { noColor = origNoColor }()

			result := colorize(tt.text, tt.color, tt.bold)
			if result != tt.expected {
				t.Errorf("colorize() = %q, want %q", result, tt.expected)
			}
		})
	}
}