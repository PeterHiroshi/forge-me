package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/PeterHiroshi/cfmon/internal/api"
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
	tailFormat         string
	tailStatusFilter   []string
	tailMethodFilter   []string
	tailSearchFilter   string
	tailIPFilter       []string
	tailHeaderFilters  []string
	tailSamplingRate   float64
	tailMaxEvents      int
	tailSince          string
	tailNoColor        bool
	tailIncludeLogs    bool
	tailIncludeErrors  bool
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
	var accountID, workerName string
	switch len(args) {
	case 1:
		workerName = args[0]
		accountID = mustGetDefaultAccount()
	case 2:
		accountID, workerName = args[0], args[1]
	}

	apiClient := api.NewClient(mustGetAPIToken())

	filter := tail.TailFilter{
		Status:       tailStatusFilter,
		Method:       tailMethodFilter,
		SamplingRate: tailSamplingRate,
	}

	// Parse since duration
	var sinceTime time.Time
	if tailSince != "" {
		duration, err := time.ParseDuration(tailSince)
		if err != nil {
			return fmt.Errorf("invalid duration: %w", err)
		}
		sinceTime = time.Now().Add(-duration)
	}

	tailSession, err := apiClient.CreateTail(accountID, workerName, filter)
	if err != nil {
		return fmt.Errorf("create tail session: %w", err)
	}
	defer apiClient.DeleteTail(accountID, workerName, tailSession.ID)

	ctx, cancel := context.WithCancel(cmd.Context())
	defer cancel()

	format := tail.OutputFormat(tailFormat)
	engine := tail.NewTailEngine(tailSession.URL, filter, format, tailMaxEvents)
	engine.Start(ctx)

	formatter := tail.NewFormatter(format)
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case event := <-engine.GetEvents():
			if event.EventTimestamp >= sinceTime.UnixNano()/int64(time.Millisecond) {
				fmt.Println(formatter.Format(event))
			}
		case err := <-engine.GetErrors():
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		case <-signalChan:
			cancel()
			return nil
		case <-ctx.Done():
			return nil
		}
	}
}

// Placeholder functions - implement with actual config loading logic
func mustGetDefaultAccount() string {
	// Load from config file
	return "default-account-id"
}

func mustGetAPIToken() string {
	// Load from config or environment
	return "api-token"
}
