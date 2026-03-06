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
	// Container command flags
	containerAccountID string
	containerSort      string
	containerLimit     int
	containerFilter    string
)

var containersCmd = &cobra.Command{
	Use:   "containers",
	Short: "Manage and monitor Cloudflare Containers",
	Long: `Manage and monitor Cloudflare Containers.

List containers with resource usage, check status, and apply filters.`,
	Example: `  # List all containers
  cfmon containers list <account-id>

  # List containers with filters
  cfmon containers list <account-id> --filter "prod" --limit 10

  # Sort by CPU usage
  cfmon containers list <account-id> --sort cpu

  # Get container status
  cfmon containers status <account-id> <container-id>`,
}

var containersListCmd = &cobra.Command{
	Use:   "list [account-id]",
	Short: "List containers with resource usage",
	Long:  `List all Cloudflare Containers with their resource usage including CPU and memory.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runContainersList,
}

var containersStatusCmd = &cobra.Command{
	Use:   "status [account-id] [container-id]",
	Short: "Get detailed status of a specific container",
	Long:  `Get detailed status information for a specific Cloudflare Container.`,
	Args:  cobra.ExactArgs(2),
	RunE:  runContainersStatus,
}

func init() {
	rootCmd.AddCommand(containersCmd)
	containersCmd.AddCommand(containersListCmd)
	containersCmd.AddCommand(containersStatusCmd)

	// Add flags to list command
	containersListCmd.Flags().StringVar(&containerSort, "sort", "", "Sort results by field (name, cpu, memory, requests)")
	containersListCmd.Flags().IntVar(&containerLimit, "limit", 0, "Limit number of results (0 = unlimited)")
	containersListCmd.Flags().StringVar(&containerFilter, "filter", "", "Filter by name (substring match)")
}

func runContainersList(cmd *cobra.Command, args []string) error {
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

	// List containers
	containers, err := client.ListContainers(accountID)
	if err != nil {
		return fmt.Errorf("listing containers: %w", err)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Debug: Retrieved %d containers\n", len(containers))
	}

	// Apply filter if specified
	if containerFilter != "" {
		filtered := []api.Container{}
		for _, c := range containers {
			if strings.Contains(strings.ToLower(c.Name), strings.ToLower(containerFilter)) {
				filtered = append(filtered, c)
			}
		}
		containers = filtered

		if verbose {
			fmt.Fprintf(os.Stderr, "Debug: Filtered to %d containers matching '%s'\n", len(containers), containerFilter)
		}
	}

	// Sort if specified
	if containerSort != "" {
		sortContainers(containers, containerSort)
		if verbose {
			fmt.Fprintf(os.Stderr, "Debug: Sorted by %s\n", containerSort)
		}
	}

	// Apply limit if specified
	if containerLimit > 0 && len(containers) > containerLimit {
		containers = containers[:containerLimit]
		if verbose {
			fmt.Fprintf(os.Stderr, "Debug: Limited to %d results\n", containerLimit)
		}
	}

	// Format output
	outFormat := getOutputFormat()
	fieldsList := parseFieldsList()

	switch outFormat {
	case "json":
		// For JSON, convert to map slice if fields filtering is requested
		if len(fieldsList) > 0 {
			data := make([]map[string]interface{}, len(containers))
			for i, c := range containers {
				data[i] = map[string]interface{}{
					"id":       c.ID,
					"name":     c.Name,
					"cpu":      c.CPUMS,
					"cpums":    c.CPUMS,
					"memory":   c.MemoryMB,
					"memorymb": c.MemoryMB,
					"requests": c.Requests,
					"status":   c.Status,
				}
			}
			filtered := output.FilterFields(data, fieldsList)
			result, err := output.FormatJSON(filtered)
			if err != nil {
				return fmt.Errorf("formatting JSON: %w", err)
			}
			fmt.Println(result)
		} else {
			result, err := output.FormatJSON(containers)
			if err != nil {
				return fmt.Errorf("formatting JSON: %w", err)
			}
			fmt.Println(result)
		}

	case "jsonl":
		// Convert to interface slice for JSONL
		data := make([]interface{}, len(containers))
		if len(fieldsList) > 0 {
			maps := make([]map[string]interface{}, len(containers))
			for i, c := range containers {
				maps[i] = map[string]interface{}{
					"id":       c.ID,
					"name":     c.Name,
					"cpu":      c.CPUMS,
					"cpums":    c.CPUMS,
					"memory":   c.MemoryMB,
					"memorymb": c.MemoryMB,
					"requests": c.Requests,
					"status":   c.Status,
				}
			}
			filtered := output.FilterFields(maps, fieldsList)
			for i, m := range filtered {
				data[i] = m
			}
		} else {
			for i, c := range containers {
				data[i] = c
			}
		}
		result, err := output.FormatJSONL(data)
		if err != nil {
			return fmt.Errorf("formatting JSONL: %w", err)
		}
		fmt.Print(result)

	case "csv", "table":
		// Build headers and rows
		headers := []string{"ID", "Name", "CPU", "Memory", "Requests"}
		rows := make([][]string, len(containers))
		for i, c := range containers {
			requests := "0"
			if c.Requests > 0 {
				requests = strconv.Itoa(c.Requests)
			}
			rows[i] = []string{
				c.ID,
				c.Name,
				strconv.Itoa(c.CPUMS),
				strconv.Itoa(c.MemoryMB),
				requests,
			}
		}

		// Filter fields if requested
		if len(fieldsList) > 0 {
			headers, rows = output.FilterTableFields(headers, rows, fieldsList)
		}

		if outFormat == "csv" {
			result := output.FormatCSV(headers, rows, !noHeader)
			fmt.Print(result)
		} else {
			// Table format
			if len(containers) == 0 {
				if !quiet {
					if containerFilter != "" {
						fmt.Printf("No containers found matching filter '%s'\n", containerFilter)
					} else {
						fmt.Println("No containers found")
					}
				}
				return nil
			}

			// Format table with or without headers
			if noHeader {
				// Print rows without headers
				for _, row := range rows {
					for i, cell := range row {
						if i > 0 {
							fmt.Print("  ")
						}
						fmt.Print(cell)
					}
					fmt.Println()
				}
			} else {
				result := output.FormatColoredTable(headers, rows, !noColor)
				fmt.Print(result)
			}

			// Summary (suppressed in quiet mode)
			if !quiet {
				if !noColor {
					fmt.Printf("\n\033[36mTotal: %d container(s)\033[0m\n", len(containers))
				} else {
					fmt.Printf("\nTotal: %d container(s)\n", len(containers))
				}
			}
		}

	default:
		return fmt.Errorf("unsupported output format: %s (use: table, json, jsonl, csv)", outFormat)
	}

	return nil
}

func runContainersStatus(cmd *cobra.Command, args []string) error {
	accountID := args[0]
	containerID := args[1]

	// Get token
	apiToken, err := getAPIToken()
	if err != nil {
		return err
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Debug: Getting status for container %s in account %s\n", containerID, accountID)
	}

	// Create API client
	client := api.NewClient(apiToken)
	if timeout > 0 {
		client.SetTimeout(timeout)
	}

	// Get container status (for now, we'll list and find the specific one)
	containers, err := client.ListContainers(accountID)
	if err != nil {
		return fmt.Errorf("getting container status: %w", err)
	}

	var container *api.Container
	for _, c := range containers {
		if c.ID == containerID || c.Name == containerID {
			container = &c
			break
		}
	}

	if container == nil {
		return fmt.Errorf("container not found: %s", containerID)
	}

	// Format output
	if format == "json" {
		result, err := output.FormatJSON(container)
		if err != nil {
			return fmt.Errorf("formatting JSON: %w", err)
		}
		fmt.Println(result)
	} else {
		// Detailed view
		fmt.Println(colorize("Container Details", "cyan", true))
		fmt.Println(strings.Repeat("=", 50))
		fmt.Printf("\n%s %s\n", colorize("ID:", "yellow", true), container.ID)
		fmt.Printf("%s %s\n", colorize("Name:", "yellow", true), container.Name)
		fmt.Printf("%s %d ms\n", colorize("CPU Usage:", "yellow", true), container.CPUMS)
		fmt.Printf("%s %d MB\n", colorize("Memory Usage:", "yellow", true), container.MemoryMB)
		if container.Requests > 0 {
			fmt.Printf("%s %d\n", colorize("Requests:", "yellow", true), container.Requests)
		}
		if container.Status != "" {
			statusColor := "green"
			if container.Status != "running" {
				statusColor = "yellow"
			}
			fmt.Printf("%s %s\n", colorize("Status:", "yellow", true), colorize(container.Status, statusColor, false))
		}
	}

	return nil
}

func sortContainers(containers []api.Container, sortBy string) {
	switch strings.ToLower(sortBy) {
	case "name":
		sort.Slice(containers, func(i, j int) bool {
			return strings.ToLower(containers[i].Name) < strings.ToLower(containers[j].Name)
		})
	case "cpu":
		sort.Slice(containers, func(i, j int) bool {
			return containers[i].CPUMS > containers[j].CPUMS // Descending
		})
	case "memory":
		sort.Slice(containers, func(i, j int) bool {
			return containers[i].MemoryMB > containers[j].MemoryMB // Descending
		})
	case "requests":
		sort.Slice(containers, func(i, j int) bool {
			return containers[i].Requests > containers[j].Requests // Descending
		})
	}
}

// getAPIToken retrieves the API token from various sources
func getAPIToken() (string, error) {
	// Priority: env var > command flag > config file
	if envToken := os.Getenv("CFMON_TOKEN"); envToken != "" {
		return envToken, nil
	}

	if token != "" {
		return token, nil
	}

	// Load from config
	configPath := cfgFile
	if configPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("getting home directory: %w", err)
		}
		configPath = filepath.Join(home, ".cfmon", "config.yaml")
	}

	cfg, err := config.Load(configPath)
	if err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("loading config: %w", err)
	}

	if cfg != nil && cfg.Token != "" {
		return cfg.Token, nil
	}

	return "", fmt.Errorf("no API token provided. Use --token flag or run 'cfmon login <token>' first")
}
