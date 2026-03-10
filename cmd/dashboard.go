package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/PeterHiroshi/cfmon/internal/api"
	"github.com/PeterHiroshi/cfmon/internal/config"
	"github.com/PeterHiroshi/cfmon/internal/dashboard"
	"github.com/spf13/cobra"
)

var dashboardRefresh time.Duration

var dashboardCmd = &cobra.Command{
	Use:   "dashboard [account-id]",
	Short: "Interactive TUI dashboard for monitoring Cloudflare resources",
	Long: `Launch an interactive terminal dashboard showing health, workers, and containers.

Use Tab or number keys (1-3) to switch between tabs.
Press 'r' to force refresh, 'q' to quit.`,
	Example: `  # Launch dashboard with explicit account ID
  cfmon dashboard abc123

  # Launch with custom refresh interval
  cfmon dashboard abc123 --refresh 10s

  # Launch using default account
  cfmon dashboard`,
	Args: cobra.MaximumNArgs(1),
	RunE: runDashboard,
}

func init() {
	rootCmd.AddCommand(dashboardCmd)
	dashboardCmd.Flags().DurationVar(&dashboardRefresh, "refresh", 30*time.Second, "Auto-refresh interval (minimum 5s)")
}

func runDashboard(cmd *cobra.Command, args []string) error {
	var accountID string

	if len(args) == 0 {
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

		if cfg == nil || cfg.DefaultAccountID == "" {
			return fmt.Errorf("no account ID provided and no default account set. Use 'cfmon accounts set-default <account-id>' to set a default")
		}

		accountID = cfg.DefaultAccountID
	} else {
		accountID = args[0]
	}

	apiToken, err := getAPIToken()
	if err != nil {
		return err
	}

	client := api.NewClient(apiToken)
	if timeout > 0 {
		client.SetTimeout(timeout)
	}

	m := dashboard.NewModel(client, accountID, dashboardRefresh)
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("running dashboard: %w", err)
	}

	return nil
}
