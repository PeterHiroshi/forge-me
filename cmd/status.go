package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/PeterHiroshi/cfmon/internal/api"
	"github.com/PeterHiroshi/cfmon/internal/config"
	"github.com/PeterHiroshi/cfmon/internal/output"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check API token status and account information",
	RunE: func(cmd *cobra.Command, args []string) error {
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

		// Get token status
		tokenStatus, err := client.GetStatus()
		if err != nil {
			return fmt.Errorf("checking token status: %w", err)
		}

		if !tokenStatus.Valid {
			fmt.Println("Token Status: Invalid")
			return fmt.Errorf("API token is invalid or unauthorized")
		}

		// Get account info
		accountInfo, err := client.GetAccountInfo()
		if err != nil {
			return fmt.Errorf("getting account info: %w", err)
		}

		// Format output
		if format == "json" {
			result := map[string]interface{}{
				"token_valid":  tokenStatus.Valid,
				"token_status": tokenStatus.Status,
				"account_id":   accountInfo.ID,
				"account_name": accountInfo.Name,
				"plan_type":    accountInfo.PlanType,
			}
			jsonOutput, err := output.FormatJSON(result)
			if err != nil {
				return fmt.Errorf("formatting JSON: %w", err)
			}
			fmt.Println(jsonOutput)
		} else {
			// Table format with colors
			headers := []string{"Property", "Value"}
			rows := [][]string{
				{"Token Valid", fmt.Sprintf("%v", tokenStatus.Valid)},
				{"Token Status", tokenStatus.Status},
				{"Account ID", accountInfo.ID},
				{"Account Name", accountInfo.Name},
				{"Plan Type", accountInfo.PlanType},
			}
			result := output.FormatColoredTable(headers, rows, !noColor)
			fmt.Print(result)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
