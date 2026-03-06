package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/PeterHiroshi/cfmon/internal/api"
	"github.com/PeterHiroshi/cfmon/internal/config"
	"github.com/PeterHiroshi/cfmon/internal/health"
	"github.com/PeterHiroshi/cfmon/internal/output"
	"github.com/spf13/cobra"
)

var healthCmd = &cobra.Command{
	Use:   "health [account-id]",
	Short: "Check health score of Cloudflare account",
	Long: `Check the health score of your Cloudflare account.

The health score is calculated based on:
  - API Connectivity (30 points)
  - Worker Health (35 points)
  - Container Health (35 points)

Total score ranges from 0-100, with higher scores indicating better health.`,
	Example: `  # Check health with explicit account ID
  cfmon health abc123

  # Check health using default account
  cfmon health

  # Get health score in JSON format
  cfmon health --output json`,
	Args: cobra.MaximumNArgs(1),
	RunE: runHealth,
}

func init() {
	rootCmd.AddCommand(healthCmd)
}

func runHealth(cmd *cobra.Command, args []string) error {
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
		fmt.Fprintf(os.Stderr, "Debug: Calculating health score for account: %s\n", accountID)
	}

	// Create API client
	client := api.NewClient(apiToken)
	if timeout > 0 {
		client.SetTimeout(timeout)
	}

	// Calculate health score
	score, err := health.CalculateScore(client, accountID)
	if err != nil {
		return fmt.Errorf("calculating health score: %w", err)
	}

	// Format output
	outFormat := getOutputFormat()

	switch outFormat {
	case "json", "jsonl":
		result, err := output.FormatJSON(score)
		if err != nil {
			return fmt.Errorf("formatting JSON: %w", err)
		}
		fmt.Println(result)

	default:
		// Colored output
		if !quiet {
			fmt.Println(colorize("Health Score", "cyan", true))
			fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		}

		// Display score with color based on status
		scoreColor := getScoreColor(score.Status)
		if !quiet {
			fmt.Printf("\n%s %s\n", colorize("Overall Score:", "yellow", true), colorize(fmt.Sprintf("%d/100", score.Total), scoreColor, true))
			fmt.Printf("%s %s\n\n", colorize("Status:", "yellow", true), colorize(score.Status, scoreColor, false))
		} else {
			fmt.Printf("%d\n", score.Total)
			return nil
		}

		// Breakdown
		fmt.Println(colorize("Score Breakdown:", "yellow", true))
		fmt.Printf("  • API Connectivity:  %s/%d\n", colorize(fmt.Sprintf("%d", score.APIConnectivity), getPointsColor(score.APIConnectivity, score.APIConnectivityMax), false), score.APIConnectivityMax)
		fmt.Printf("  • Worker Health:     %s/%d\n", colorize(fmt.Sprintf("%d", score.WorkerHealth), getPointsColor(score.WorkerHealth, score.WorkerHealthMax), false), score.WorkerHealthMax)
		fmt.Printf("  • Container Health:  %s/%d\n", colorize(fmt.Sprintf("%d", score.ContainerHealth), getPointsColor(score.ContainerHealth, score.ContainerHealthMax), false), score.ContainerHealthMax)

		if score.Message != "" {
			fmt.Printf("\n%s %s\n", colorize("Message:", "yellow", true), score.Message)
		}

		fmt.Printf("\n%s %s\n", colorize("Checked at:", "yellow", true), score.Timestamp)
	}

	return nil
}

// getScoreColor returns the appropriate color based on status
func getScoreColor(status string) string {
	switch status {
	case "excellent", "good":
		return "green"
	case "fair":
		return "yellow"
	case "poor", "critical":
		return "red"
	default:
		return "white"
	}
}

// getPointsColor returns color based on points achieved vs max
func getPointsColor(points, max int) string {
	percent := float64(points) / float64(max) * 100
	switch {
	case percent >= 90:
		return "green"
	case percent >= 70:
		return "yellow"
	default:
		return "red"
	}
}
