package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/PeterHiroshi/cfmon/internal/config"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login [token]",
	Short: "Save API token to config",
	Long:  `Save your Cloudflare API token to ~/.cfmon/config.yaml for future use.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		token := args[0]

		// Get config path
		configPath := cfgFile
		if configPath == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("getting home directory: %w", err)
			}
			configPath = filepath.Join(home, ".cfmon", "config.yaml")
		}

		// Save config
		cfg := &config.Config{
			Token: token,
		}

		if err := config.Save(configPath, cfg); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}

		fmt.Printf("Token saved to %s\n", configPath)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
