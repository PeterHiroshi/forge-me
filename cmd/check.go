package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/PeterHiroshi/cfmon/internal/api"
	"github.com/PeterHiroshi/cfmon/internal/config"
	"github.com/PeterHiroshi/cfmon/internal/monitor"
	"github.com/PeterHiroshi/cfmon/internal/output"
	"github.com/spf13/cobra"
)

var (
	cpuThreshold    float64
	memoryThreshold float64
	errorThreshold  float64
)

var checkCmd = &cobra.Command{
	Use:   "check [account-id]",
	Short: "One-shot health check with threshold-based alerts",
	Long: `Run a one-shot health check against your Cloudflare account.

Evaluates workers and containers against configurable thresholds and
generates alerts with severity levels (ok, warning, critical).

Exit codes:
  0 = healthy (no alerts)
  1 = warnings detected
  2 = critical alerts detected`,
	Example: `  # Check with default thresholds
  cfmon check abc123

  # Check with custom thresholds
  cfmon check abc123 --cpu-threshold 70 --memory-threshold 80 --error-threshold 1

  # JSON output for automation
  cfmon check abc123 --output json

  # Use default account
  cfmon check`,
	Args: cobra.MaximumNArgs(1),
	RunE: runCheck,
}

func init() {
	rootCmd.AddCommand(checkCmd)

	checkCmd.Flags().Float64Var(&cpuThreshold, "cpu-threshold", 80, "CPU usage warning threshold (percent)")
	checkCmd.Flags().Float64Var(&memoryThreshold, "memory-threshold", 85, "Memory usage warning threshold (percent)")
	checkCmd.Flags().Float64Var(&errorThreshold, "error-threshold", 2, "Error rate warning threshold (percent)")
}

func runCheck(cmd *cobra.Command, args []string) error {
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

	apiToken, err := getAPIToken()
	if err != nil {
		return err
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Debug: Running check for account: %s\n", accountID)
	}

	client := api.NewClient(apiToken)
	if timeout > 0 {
		client.SetTimeout(timeout)
	}

	th := monitor.Thresholds{
		CPUPercent:       cpuThreshold,
		MemoryPercent:    memoryThreshold,
		ErrorRatePercent: errorThreshold,
	}

	result, err := monitor.RunCheck(client, accountID, th)
	if err != nil {
		return fmt.Errorf("running check: %w", err)
	}

	outFormat := getOutputFormat()

	switch outFormat {
	case "json", "jsonl":
		jsonStr, err := output.FormatJSON(result)
		if err != nil {
			return fmt.Errorf("formatting JSON: %w", err)
		}
		fmt.Println(jsonStr)

	default:
		printCheckTable(result)
	}

	// Set exit code based on severity
	switch result.Summary.MaxSeverity {
	case "critical":
		os.Exit(2)
	case "warning":
		os.Exit(1)
	}

	return nil
}

func printCheckTable(result *monitor.CheckResult) {
	if !quiet {
		fmt.Println(colorize("Health Check", "cyan", true))
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	}

	sevColor := "green"
	switch result.Summary.MaxSeverity {
	case "warning":
		sevColor = "yellow"
	case "critical":
		sevColor = "red"
	}

	if !quiet {
		fmt.Printf("\n%s %s\n", colorize("Status:", "yellow", true), colorize(result.Summary.MaxSeverity, sevColor, true))
		fmt.Printf("%s Workers: %d  Containers: %d  Alerts: %d\n\n",
			colorize("Summary:", "yellow", true),
			result.Summary.TotalWorkers,
			result.Summary.TotalContainers,
			result.Summary.TotalAlerts,
		)
	}

	if len(result.Alerts) == 0 {
		if !quiet {
			fmt.Println(colorize("No alerts - all resources within thresholds.", "green", false))
		}
		return
	}

	// Print alerts table
	headers := []string{"SEVERITY", "TYPE", "RESOURCE", "METRIC", "VALUE", "THRESHOLD", "MESSAGE"}
	rows := make([][]string, len(result.Alerts))
	for i, a := range result.Alerts {
		rows[i] = []string{
			a.Severity,
			a.ResourceType,
			a.ResourceName,
			a.Metric,
			fmt.Sprintf("%.1f%%", a.Value),
			fmt.Sprintf("%.1f%%", a.Threshold),
			a.Message,
		}
	}

	fmt.Print(output.FormatTable(headers, rows))

	if !quiet {
		fmt.Printf("\n%s %s\n", colorize("Checked at:", "yellow", true), result.Timestamp)
	}
}
