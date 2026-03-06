package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/PeterHiroshi/cfmon/internal/api"
	"github.com/PeterHiroshi/cfmon/internal/config"
	"github.com/PeterHiroshi/cfmon/internal/watch"
	"github.com/spf13/cobra"
)

var (
	watchInterval  time.Duration
	watchEventsOnly bool
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch resources for changes in real-time",
	Long: `Watch Cloudflare resources for changes in real-time.

Continuously monitors resources and outputs changes as JSON lines.
Use Ctrl+C to stop watching gracefully.`,
	Example: `  # Watch containers for changes
  cfmon watch containers abc123

  # Watch workers with custom interval
  cfmon watch workers abc123 --interval 10s

  # Only show events (no heartbeats)
  cfmon watch containers --events-only`,
}

var watchContainersCmd = &cobra.Command{
	Use:   "containers [account-id]",
	Short: "Watch containers for changes",
	Long:  `Continuously monitor Cloudflare Containers and output changes as JSON lines.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runWatchContainers,
}

var watchWorkersCmd = &cobra.Command{
	Use:   "workers [account-id]",
	Short: "Watch workers for changes",
	Long:  `Continuously monitor Cloudflare Workers and output changes as JSON lines.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runWatchWorkers,
}

func init() {
	rootCmd.AddCommand(watchCmd)
	watchCmd.AddCommand(watchContainersCmd)
	watchCmd.AddCommand(watchWorkersCmd)

	// Add flags
	watchCmd.PersistentFlags().DurationVar(&watchInterval, "interval", 30*time.Second, "Check interval")
	watchCmd.PersistentFlags().BoolVar(&watchEventsOnly, "events-only", false, "Only output change events (no heartbeats)")
}

func runWatchContainers(cmd *cobra.Command, args []string) error {
	var accountID string

	// If no account ID provided, try to use default from config
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
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("loading config: %w", err)
		}

		if cfg == nil || cfg.DefaultAccountID == "" {
			return fmt.Errorf("no account ID provided and no default account set. Use 'cfmon accounts set-default <account-id>' to set a default")
		}

		accountID = cfg.DefaultAccountID
		if verbose {
			fmt.Fprintf(os.Stderr, "Debug: Using default account ID: %s\n", accountID)
		}
	} else {
		accountID = args[0]
	}

	// Get token
	apiToken, err := getAPIToken()
	if err != nil {
		return err
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Debug: Watching containers for account: %s\n", accountID)
		fmt.Fprintf(os.Stderr, "Debug: Interval: %s\n", watchInterval)
		fmt.Fprintf(os.Stderr, "Debug: Events only: %v\n", watchEventsOnly)
	}

	// Create API client
	client := api.NewClient(apiToken)
	if timeout > 0 {
		client.SetTimeout(timeout)
	}

	// Start watching
	options := watch.WatchOptions{
		Interval:   watchInterval,
		EventsOnly: watchEventsOnly,
	}

	return watch.WatchContainers(client, accountID, options)
}

func runWatchWorkers(cmd *cobra.Command, args []string) error {
	var accountID string

	// If no account ID provided, try to use default from config
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
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("loading config: %w", err)
		}

		if cfg == nil || cfg.DefaultAccountID == "" {
			return fmt.Errorf("no account ID provided and no default account set. Use 'cfmon accounts set-default <account-id>' to set a default")
		}

		accountID = cfg.DefaultAccountID
		if verbose {
			fmt.Fprintf(os.Stderr, "Debug: Using default account ID: %s\n", accountID)
		}
	} else {
		accountID = args[0]
	}

	// Get token
	apiToken, err := getAPIToken()
	if err != nil {
		return err
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Debug: Watching workers for account: %s\n", accountID)
		fmt.Fprintf(os.Stderr, "Debug: Interval: %s\n", watchInterval)
		fmt.Fprintf(os.Stderr, "Debug: Events only: %v\n", watchEventsOnly)
	}

	// Create API client
	client := api.NewClient(apiToken)
	if timeout > 0 {
		client.SetTimeout(timeout)
	}

	// Start watching
	options := watch.WatchOptions{
		Interval:   watchInterval,
		EventsOnly: watchEventsOnly,
	}

	return watch.WatchWorkers(client, accountID, options)
}
