package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/PeterHiroshi/cfmon/internal/api"
	"github.com/PeterHiroshi/cfmon/internal/config"
	"github.com/PeterHiroshi/cfmon/internal/output"
	"github.com/spf13/cobra"
)

var (
	// Worker command flags
	workerSort   string
	workerLimit  int
	workerFilter string
)

var workersCmd = &cobra.Command{
	Use:   "workers",
	Short: "Manage and monitor Cloudflare Workers",
	Long: `Manage and monitor Cloudflare Workers.

List workers with metrics, check status, and apply filters.`,
	Example: `  # List all workers
  cfmon workers list <account-id>

  # List workers with filters
  cfmon workers list <account-id> --filter "api" --limit 10

  # Sort by requests
  cfmon workers list <account-id> --sort requests

  # Get worker status
  cfmon workers status <account-id> <worker-name>`,
}

var workersListCmd = &cobra.Command{
	Use:   "list [account-id]",
	Short: "List workers with metrics",
	Long:  `List all Cloudflare Workers with their metrics including CPU usage and request counts.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runWorkersList,
}

var workersStatusCmd = &cobra.Command{
	Use:   "status [account-id] [worker-name]",
	Short: "Get detailed status of a specific worker",
	Long:  `Get detailed status information for a specific Cloudflare Worker.`,
	Args:  cobra.ExactArgs(2),
	RunE:  runWorkersStatus,
}

func init() {
	rootCmd.AddCommand(workersCmd)
	workersCmd.AddCommand(workersListCmd)
	workersCmd.AddCommand(workersStatusCmd)

	// Add flags to list command
	workersListCmd.Flags().StringVar(&workerSort, "sort", "", "Sort results by field (name, cpu, requests)")
	workersListCmd.Flags().IntVar(&workerLimit, "limit", 0, "Limit number of results (0 = unlimited)")
	workersListCmd.Flags().StringVar(&workerFilter, "filter", "", "Filter by name (substring match)")
}

func runWorkersList(cmd *cobra.Command, args []string) error {
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
		fmt.Fprintf(os.Stderr, "Debug: Using account ID: %s\n", accountID)
		fmt.Fprintf(os.Stderr, "Debug: API timeout: %s\n", timeout)
	}

	// Create API client
	client := api.NewClient(apiToken)
	if timeout > 0 {
		client.SetTimeout(timeout)
	}

	// List workers
	workers, err := client.ListWorkers(accountID)
	if err != nil {
		return fmt.Errorf("listing workers: %w", err)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Debug: Retrieved %d workers\n", len(workers))
	}

	// Apply filter if specified
	if workerFilter != "" {
		filtered := []api.Worker{}
		for _, w := range workers {
			if strings.Contains(strings.ToLower(w.Name), strings.ToLower(workerFilter)) {
				filtered = append(filtered, w)
			}
		}
		workers = filtered

		if verbose {
			fmt.Fprintf(os.Stderr, "Debug: Filtered to %d workers matching '%s'\n", len(workers), workerFilter)
		}
	}

	// Sort if specified
	if workerSort != "" {
		sortWorkers(workers, workerSort)
		if verbose {
			fmt.Fprintf(os.Stderr, "Debug: Sorted by %s\n", workerSort)
		}
	}

	// Apply limit if specified
	if workerLimit > 0 && len(workers) > workerLimit {
		workers = workers[:workerLimit]
		if verbose {
			fmt.Fprintf(os.Stderr, "Debug: Limited to %d results\n", workerLimit)
		}
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
		if len(workers) == 0 {
			if workerFilter != "" {
				fmt.Printf("No workers found matching filter '%s'\n", workerFilter)
			} else {
				fmt.Println("No workers found")
			}
			return nil
		}

		headers := []string{"ID", "Name", "CPU (ms)", "Requests", "Errors", "Status"}
		rows := make([][]string, len(workers))
		for i, w := range workers {
			errors := "0"
			if w.Errors > 0 {
				errors = strconv.Itoa(w.Errors)
			}
			status := "active"
			if w.Status != "" {
				status = w.Status
			}
			rows[i] = []string{
				w.ID,
				w.Name,
				strconv.Itoa(w.CPUMS),
				strconv.Itoa(w.Requests),
				errors,
				status,
			}
		}
		result := output.FormatColoredTable(headers, rows, !noColor)
		fmt.Print(result)

		// Summary
		if !noColor {
			fmt.Printf("\n\033[36mTotal: %d worker(s)\033[0m\n", len(workers))
		} else {
			fmt.Printf("\nTotal: %d worker(s)\n", len(workers))
		}
	}

	return nil
}

func runWorkersStatus(cmd *cobra.Command, args []string) error {
	accountID := args[0]
	workerName := args[1]

	// Get token
	apiToken, err := getAPIToken()
	if err != nil {
		return err
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Debug: Getting status for worker %s in account %s\n", workerName, accountID)
	}

	// Create API client
	client := api.NewClient(apiToken)
	if timeout > 0 {
		client.SetTimeout(timeout)
	}

	// Get worker status (for now, we'll list and find the specific one)
	workers, err := client.ListWorkers(accountID)
	if err != nil {
		return fmt.Errorf("getting worker status: %w", err)
	}

	var worker *api.Worker
	for _, w := range workers {
		if w.ID == workerName || w.Name == workerName {
			worker = &w
			break
		}
	}

	if worker == nil {
		return fmt.Errorf("worker not found: %s", workerName)
	}

	// Format output
	if format == "json" {
		result, err := output.FormatJSON(worker)
		if err != nil {
			return fmt.Errorf("formatting JSON: %w", err)
		}
		fmt.Println(result)
	} else {
		// Detailed view
		fmt.Println(colorize("Worker Details", "cyan", true))
		fmt.Println(strings.Repeat("=", 50))
		fmt.Printf("\n%s %s\n", colorize("ID:", "yellow", true), worker.ID)
		fmt.Printf("%s %s\n", colorize("Name:", "yellow", true), worker.Name)
		fmt.Printf("%s %d ms\n", colorize("CPU Usage:", "yellow", true), worker.CPUMS)
		fmt.Printf("%s %d\n", colorize("Requests:", "yellow", true), worker.Requests)
		if worker.Errors > 0 {
			fmt.Printf("%s %s\n", colorize("Errors:", "yellow", true), colorize(strconv.Itoa(worker.Errors), "red", false))
		} else {
			fmt.Printf("%s 0\n", colorize("Errors:", "yellow", true))
		}
		if worker.Status != "" {
			statusColor := "green"
			if worker.Status != "active" && worker.Status != "running" {
				statusColor = "yellow"
			}
			fmt.Printf("%s %s\n", colorize("Status:", "yellow", true), colorize(worker.Status, statusColor, false))
		} else {
			fmt.Printf("%s %s\n", colorize("Status:", "yellow", true), colorize("active", "green", false))
		}

		// Additional metrics if available
		if worker.SuccessRate > 0 {
			rateColor := "green"
			if worker.SuccessRate < 95 {
				rateColor = "yellow"
			}
			if worker.SuccessRate < 90 {
				rateColor = "red"
			}
			fmt.Printf("%s %s%%\n", colorize("Success Rate:", "yellow", true), colorize(fmt.Sprintf("%.2f", worker.SuccessRate), rateColor, false))
		}
	}

	return nil
}

func sortWorkers(workers []api.Worker, sortBy string) {
	switch strings.ToLower(sortBy) {
	case "name":
		sort.Slice(workers, func(i, j int) bool {
			return strings.ToLower(workers[i].Name) < strings.ToLower(workers[j].Name)
		})
	case "cpu":
		sort.Slice(workers, func(i, j int) bool {
			return workers[i].CPUMS > workers[j].CPUMS // Descending
		})
	case "requests":
		sort.Slice(workers, func(i, j int) bool {
			return workers[i].Requests > workers[j].Requests // Descending
		})
	case "errors":
		sort.Slice(workers, func(i, j int) bool {
			return workers[i].Errors > workers[j].Errors // Descending
		})
	}
}
