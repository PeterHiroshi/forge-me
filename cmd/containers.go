package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/PeterHiroshi/cfmon/internal/api"
	"github.com/PeterHiroshi/cfmon/internal/config"
	"github.com/PeterHiroshi/cfmon/internal/output"
	"github.com/spf13/cobra"
)

var containersCmd = &cobra.Command{
	Use:   "containers [account-id]",
	Short: "List containers and their resource usage",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		accountID := args[0]

		// Get token from flag or config
		apiToken := token
		if apiToken == "" {
			configPath := cfgFile
			if configPath == "" {
				home, err := os.UserHomeDir()
				if err != nil {
					return fmt.Errorf("getting home directory: %w", err)
				}
				configPath = filepath.Join(home, ".cfmon", "config.yaml")
			}

			cfg, err := config.Load(configPath)
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}

			apiToken = cfg.Token
		}

		if apiToken == "" {
			return fmt.Errorf("no API token provided. Use --token flag or run 'cfmon login' first")
		}

		// Create API client
		client := api.NewClient(apiToken)

		// List containers
		containers, err := client.ListContainers(accountID)
		if err != nil {
			return fmt.Errorf("listing containers: %w", err)
		}

		// Format output
		if format == "json" {
			result, err := output.FormatJSON(containers)
			if err != nil {
				return fmt.Errorf("formatting JSON: %w", err)
			}
			fmt.Println(result)
		} else {
			// Table format with colors
			headers := []string{"ID", "Name", "CPU (ms)", "Memory (MB)"}
			rows := make([][]string, len(containers))
			for i, c := range containers {
				rows[i] = []string{
					c.ID,
					c.Name,
					strconv.Itoa(c.CPUMS),
					strconv.Itoa(c.MemoryMB),
				}
			}
			result := output.FormatColoredTable(headers, rows, !noColor)
			fmt.Print(result)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(containersCmd)
}
