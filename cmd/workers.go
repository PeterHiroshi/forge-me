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

var workersCmd = &cobra.Command{
	Use:   "workers [account-id]",
	Short: "List workers and their resource usage",
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

		// List workers
		workers, err := client.ListWorkers(accountID)
		if err != nil {
			return fmt.Errorf("listing workers: %w", err)
		}

		// Format output
		if format == "json" {
			result, err := output.FormatJSON(workers)
			if err != nil {
				return fmt.Errorf("formatting JSON: %w", err)
			}
			fmt.Println(result)
		} else {
			// Table format with colors
			headers := []string{"ID", "Name", "CPU (ms)", "Requests"}
			rows := make([][]string, len(workers))
			for i, w := range workers {
				rows[i] = []string{
					w.ID,
					w.Name,
					strconv.Itoa(w.CPUMS),
					strconv.Itoa(w.Requests),
				}
			}
			result := output.FormatColoredTable(headers, rows, !noColor)
			fmt.Print(result)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(workersCmd)
}
