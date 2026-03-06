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

var accountsCmd = &cobra.Command{
	Use:   "accounts",
	Short: "Manage Cloudflare accounts",
	Long: `Manage Cloudflare accounts and set default account ID.

Allows you to set a default account ID to avoid providing it for every command,
list available accounts, and check the current default.`,
	Example: `  # Set default account
  cfmon accounts set-default abc123

  # Get current default account
  cfmon accounts get-default

  # List all accounts
  cfmon accounts list`,
}

var accountsSetDefaultCmd = &cobra.Command{
	Use:   "set-default [account-id]",
	Short: "Set the default account ID",
	Long:  `Set the default account ID to be used when no account ID is specified in commands.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runAccountsSetDefault,
}

var accountsGetDefaultCmd = &cobra.Command{
	Use:   "get-default",
	Short: "Get the default account ID",
	Long:  `Display the currently configured default account ID.`,
	Args:  cobra.NoArgs,
	RunE:  runAccountsGetDefault,
}

var accountsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available accounts",
	Long:  `List all Cloudflare accounts accessible with the current API token.`,
	Args:  cobra.NoArgs,
	RunE:  runAccountsList,
}

func init() {
	rootCmd.AddCommand(accountsCmd)
	accountsCmd.AddCommand(accountsSetDefaultCmd)
	accountsCmd.AddCommand(accountsGetDefaultCmd)
	accountsCmd.AddCommand(accountsListCmd)
}

func runAccountsSetDefault(cmd *cobra.Command, args []string) error {
	accountID := args[0]

	// Get config path
	configPath := cfgFile
	if configPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("getting home directory: %w", err)
		}
		configPath = filepath.Join(home, ".cfmon", "config.yaml")
	}

	// Load existing config or create new one
	cfg, err := config.Load(configPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("loading config: %w", err)
	}
	if cfg == nil {
		cfg = config.New()
	}

	// Set default account ID
	cfg.DefaultAccountID = accountID

	// Save config
	if err := config.Save(configPath, cfg); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	if format == "json" {
		result := map[string]string{
			"default_account_id": accountID,
			"config_path":        configPath,
			"status":             "success",
		}
		output, err := output.FormatJSON(result)
		if err != nil {
			return fmt.Errorf("formatting JSON: %w", err)
		}
		fmt.Println(output)
	} else {
		fmt.Printf("Default account ID set to: %s\n", colorize(accountID, "green", true))
		if verbose {
			fmt.Printf("Configuration saved to: %s\n", configPath)
		}
	}

	return nil
}

func runAccountsGetDefault(cmd *cobra.Command, args []string) error {
	// Get config path
	configPath := cfgFile
	if configPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("getting home directory: %w", err)
		}
		configPath = filepath.Join(home, ".cfmon", "config.yaml")
	}

	// Load config
	cfg, err := config.Load(configPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("loading config: %w", err)
	}

	if cfg == nil || cfg.DefaultAccountID == "" {
		if format == "json" {
			result := map[string]interface{}{
				"default_account_id": nil,
				"status":             "not_set",
			}
			output, err := output.FormatJSON(result)
			if err != nil {
				return fmt.Errorf("formatting JSON: %w", err)
			}
			fmt.Println(output)
		} else {
			fmt.Println("No default account ID set")
			fmt.Println("Use 'cfmon accounts set-default <account-id>' to set one")
		}
		return nil
	}

	if format == "json" {
		result := map[string]string{
			"default_account_id": cfg.DefaultAccountID,
			"status":             "set",
		}
		output, err := output.FormatJSON(result)
		if err != nil {
			return fmt.Errorf("formatting JSON: %w", err)
		}
		fmt.Println(output)
	} else {
		fmt.Println(cfg.DefaultAccountID)
	}

	return nil
}

func runAccountsList(cmd *cobra.Command, args []string) error {
	// Get token
	apiToken, err := getAPIToken()
	if err != nil {
		return err
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Debug: Listing accounts\n")
		fmt.Fprintf(os.Stderr, "Debug: API timeout: %s\n", timeout)
	}

	// Create API client
	client := api.NewClient(apiToken)
	if timeout > 0 {
		client.SetTimeout(timeout)
	}

	// List accounts
	accounts, err := client.ListAccounts()
	if err != nil {
		return fmt.Errorf("listing accounts: %w", err)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Debug: Retrieved %d accounts\n", len(accounts))
	}

	// Get current default account ID
	configPath := cfgFile
	if configPath == "" {
		home, err := os.UserHomeDir()
		if err == nil {
			configPath = filepath.Join(home, ".cfmon", "config.yaml")
		}
	}
	var defaultAccountID string
	if configPath != "" {
		cfg, _ := config.Load(configPath)
		if cfg != nil {
			defaultAccountID = cfg.DefaultAccountID
		}
	}

	// Format output
	if format == "json" {
		// Add default flag to accounts
		type AccountWithDefault struct {
			api.Account
			IsDefault bool `json:"is_default"`
		}
		accountsWithDefault := make([]AccountWithDefault, len(accounts))
		for i, acc := range accounts {
			accountsWithDefault[i] = AccountWithDefault{
				Account:   acc,
				IsDefault: acc.ID == defaultAccountID,
			}
		}
		result, err := output.FormatJSON(accountsWithDefault)
		if err != nil {
			return fmt.Errorf("formatting JSON: %w", err)
		}
		fmt.Println(result)
	} else {
		// Table format with colors
		if len(accounts) == 0 {
			fmt.Println("No accounts found")
			return nil
		}

		headers := []string{"ID", "Name", "Type", "Default"}
		rows := make([][]string, len(accounts))
		for i, acc := range accounts {
			isDefault := ""
			if acc.ID == defaultAccountID {
				isDefault = "✓"
			}
			accountType := "standard"
			if acc.Type != "" {
				accountType = acc.Type
			}
			rows[i] = []string{
				acc.ID,
				acc.Name,
				accountType,
				isDefault,
			}
		}
		result := output.FormatColoredTable(headers, rows, !noColor)
		fmt.Print(result)

		// Summary
		if !noColor {
			fmt.Printf("\n\033[36mTotal: %d account(s)\033[0m\n", len(accounts))
		} else {
			fmt.Printf("\nTotal: %d account(s)\n", len(accounts))
		}

		if defaultAccountID != "" {
			fmt.Printf("Default account: %s\n", colorize(defaultAccountID, "green", true))
		}
	}

	return nil
}