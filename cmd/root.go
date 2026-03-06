// cfmon CLI - Cloudflare management tool
package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	// Output flags
	format     string // Deprecated: use outputFormat
	outputFormat string
	fields     string
	quiet      bool
	noHeader   bool
	noColor    bool

	// Authentication flags
	token    string
	cfgFile  string

	// Global flags
	verbose  bool
	timeout  time.Duration
)

// Version info (set by build flags)
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "cfmon",
	Short: "Cloudflare Workers/Containers CLI monitoring tool",
	Long: `cfmon - A powerful CLI tool for monitoring and managing Cloudflare Workers and Containers

cfmon provides an intuitive command-line interface to interact with Cloudflare's API,
allowing you to monitor resource usage, check status, and manage your infrastructure
efficiently from the terminal.

Features:
  • List and monitor Cloudflare Workers
  • List and monitor Cloudflare Containers
  • Check resource usage (CPU, memory, requests)
  • Multiple output formats (table, JSON)
  • Secure token management
  • Shell completion support

Quick Start:
  1. Set your Cloudflare API token:
     $ cfmon login <your-token>

  2. List your resources:
     $ cfmon containers list
     $ cfmon workers list

  3. Check system status:
     $ cfmon doctor

For more information, visit: https://github.com/PeterHiroshi/cfmon`,
	Example: `  # Set up authentication
  cfmon login <token>

  # List containers with filters
  cfmon containers list --filter "prod" --limit 10

  # Get worker status in JSON format
  cfmon workers status my-worker --json

  # Check system health
  cfmon doctor`,
	Version: Version,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	// Set custom help template for better formatting
	rootCmd.SetHelpTemplate(helpTemplate)

	if err := rootCmd.Execute(); err != nil {
		// Enhanced error handling with suggestions
		handleError(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Output flags
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, json, jsonl, csv)")
	rootCmd.PersistentFlags().StringVar(&format, "format", "", "Output format (deprecated: use --output instead)")
	rootCmd.PersistentFlags().StringVar(&fields, "fields", "", "Comma-separated list of fields to include (e.g., name,cpu,memory)")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Suppress headers, summaries, and decorations")
	rootCmd.PersistentFlags().BoolVar(&noHeader, "no-header", false, "Suppress column headers in table/csv output")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable colored output")

	// Authentication flags
	rootCmd.PersistentFlags().StringVar(&token, "token", "", "Cloudflare API token (overrides config)")
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "Config file (default: $HOME/.cfmon/config.yaml)")

	// Global flags
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose debug output")
	rootCmd.PersistentFlags().DurationVar(&timeout, "timeout", 30*time.Second, "API request timeout")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// Handle deprecated --format flag
	if format != "" && outputFormat == "table" {
		outputFormat = format
	}
}

// getOutputFormat returns the effective output format
func getOutputFormat() string {
	if format != "" && outputFormat == "table" {
		return format
	}
	return outputFormat
}

// parseFieldsList parses the comma-separated fields list
func parseFieldsList() []string {
	if fields == "" {
		return nil
	}
	parts := strings.Split(fields, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// handleError provides user-friendly error messages with suggestions
func handleError(err error) {
	fmt.Fprintf(os.Stderr, "\033[31m✗ Error:\033[0m %v\n", err)

	// Provide helpful suggestions based on error type
	switch {
	case err.Error() == "token required" || err.Error() == "authentication failed":
		fmt.Fprintf(os.Stderr, "\n\033[33mSuggestion:\033[0m Run 'cfmon login <token>' to set your Cloudflare API token\n")
	case err.Error() == "network error" || err.Error() == "connection failed":
		fmt.Fprintf(os.Stderr, "\n\033[33mSuggestion:\033[0m Check your internet connection and try again\n")
	case err.Error() == "timeout":
		fmt.Fprintf(os.Stderr, "\n\033[33mSuggestion:\033[0m Try increasing the timeout with --timeout flag\n")
	default:
		fmt.Fprintf(os.Stderr, "\n\033[33mFor help:\033[0m Run 'cfmon help' or 'cfmon <command> --help'\n")
	}
}

// Custom help template for better formatting
const helpTemplate = `{{with (or .Long .Short)}}{{. | trimTrailingWhitespaces}}

{{end}}{{if or .Runnable .HasSubCommands}}{{.UsageString}}{{end}}`
