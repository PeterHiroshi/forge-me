package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// Version is set by GoReleaser
	Version = "dev"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("cfmon version %s\n", Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
