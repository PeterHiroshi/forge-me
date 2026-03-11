package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/PeterHiroshi/cfmon/internal/api"
	"github.com/PeterHiroshi/cfmon/internal/config"
	"github.com/PeterHiroshi/cfmon/internal/tail"
	"github.com/spf13/cobra"
)

var tailCmd = &cobra.Command{
	Use:   "tail [account-id] <worker-name>",
	Short: "Stream real-time logs from Cloudflare Workers and Containers",
	Long: `Stream real-time logs with advanced filtering and formatting.

More powerful than wrangler tail with:
- Colored output
- Multiple formats
- Advanced filtering
- Auto-reconnect
- Error handling`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runTail,
}

var (
	tailFormat        string
	tailStatusFilter  []string
	tailMethodFilter  []string
	tailSearchFilter  string
	tailIPFilter      []string
	tailHeaderFilters []string
	tailSamplingRate  float64
	tailMaxEvents     int
	tailSince         string
	tailNoColor       bool
	tailIncludeLogs   bool
	tailIncludeErrors bool
)

func init() {
	rootCmd.AddCommand(tailCmd)

	tailCmd.Flags().StringVarP(&tailFormat, "format", "f", "pretty", "Output format: pretty, json, compact")
	tailCmd.Flags().StringSliceVar(&tailStatusFilter, "status", nil, "Filter by HTTP status: ok, error, 200, 500")
	tailCmd.Flags().StringSliceVar(&tailMethodFilter, "method", nil, "Filter by HTTP method: GET, POST, etc")
	tailCmd.Flags().StringVar(&tailSearchFilter, "search", "", "Filter logs containing this string")
	tailCmd.Flags().StringSliceVar(&tailIPFilter, "ip", nil, "Filter by client IP address")
	tailCmd.Flags().StringSliceVar(&tailHeaderFilters, "header", nil, "Filter by header (key:value)")
	tailCmd.Flags().Float64Var(&tailSamplingRate, "sample-rate", 1.0, "Sampling rate (0.0-1.0)")
	tailCmd.Flags().IntVarP(&tailMaxEvents, "max-events", "n", 0, "Stop after N events")
	tailCmd.Flags().StringVar(&tailSince, "since", "", "Only show events after duration (e.g. '5m', '1h')")
	tailCmd.Flags().BoolVar(&tailNoColor, "no-color", false, "Disable colored output")
	tailCmd.Flags().BoolVar(&tailIncludeLogs, "include-logs", true, "Show console.log() output")
	tailCmd.Flags().BoolVar(&tailIncludeErrors, "include-exceptions", true, "Show exceptions")
}

func runTail(cmd *cobra.Command, args []string) error {
	// Validate sample rate
	if tailSamplingRate < 0 || tailSamplingRate > 1.0 {
		return fmt.Errorf("invalid sample rate %.2f: must be between 0.0 and 1.0", tailSamplingRate)
	}

	var accountID, workerName string
	switch len(args) {
	case 1:
		workerName = args[0]
		// Try to load default account from config
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
		if cfg != nil && cfg.DefaultAccountID != "" {
			accountID = cfg.DefaultAccountID
		} else {
			return fmt.Errorf("no account ID provided and no default account set")
		}
	case 2:
		accountID, workerName = args[0], args[1]
	}

	apiToken, err := getAPIToken()
	if err != nil {
		return err
	}

	apiClient := api.NewClient(apiToken)

	filter := api.TailFilter{
		Status:       tailStatusFilter,
		Method:       tailMethodFilter,
		SamplingRate: tailSamplingRate,
	}

	// Parse since duration
	var sinceDuration time.Duration
	if tailSince != "" {
		sinceDuration, err = time.ParseDuration(tailSince)
		if err != nil {
			return fmt.Errorf("invalid duration: %w", err)
		}
	}

	tailSession, err := apiClient.CreateTail(accountID, workerName, filter)
	if err != nil {
		return fmt.Errorf("create tail session: %w", err)
	}
	defer apiClient.DeleteTail(accountID, workerName, tailSession.ID)

	ctx, cancel := context.WithCancel(cmd.Context())
	defer cancel()

	formatter := tail.NewFormatter(tailFormat, tailNoColor)
	formatter.IncludeLogs = tailIncludeLogs
	formatter.IncludeExceptions = tailIncludeErrors

	engine := tail.NewEngine(tail.EngineConfig{
		WebSocketURL: tailSession.URL,
		MaxEvents:    tailMaxEvents,
		Search:       tailSearchFilter,
		Since:        sinceDuration,
		OnEvent: func(event tail.TailEvent) {
			fmt.Println(formatter.Format(event))
		},
		OnError: func(err error) {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		},
	})

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		select {
		case <-signalChan:
			engine.Stop()
			cancel()
		case <-ctx.Done():
			engine.Stop()
		}
	}()

	engine.Run()
	return nil
}
