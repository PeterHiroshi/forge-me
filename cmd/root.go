// cfmon CLI - Cloudflare management tool
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	format   string
	token    string
	cfgFile  string
	noColor  bool
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "cfmon",
	Short: "A CLI tool for Cloudflare API",
	Long:  `cfmon is a CLI tool that talks to the Cloudflare API to list containers and workers with their resource usage.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Version = "0.1.0"
	// Global flags
	rootCmd.PersistentFlags().StringVar(&format, "format", "table", "Output format (table or json)")
	rootCmd.PersistentFlags().StringVar(&token, "token", "", "Cloudflare API token")
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "Config file (default is $HOME/.cfmon/config.yaml)")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable colored output")
}
