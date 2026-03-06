package watch

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/PeterHiroshi/cfmon/internal/api"
)

// Event represents a change detected during watching
type Event struct {
	Timestamp string      `json:"timestamp"`
	Type      string      `json:"type"`
	Resource  string      `json:"resource"`
	Name      string      `json:"name"`
	Change    string      `json:"change"`
	Data      interface{} `json:"data,omitempty"`
}

// WatchOptions configures the watch behavior
type WatchOptions struct {
	Interval   time.Duration
	EventsOnly bool
}

// WatchContainers continuously monitors containers for changes
func WatchContainers(client *api.Client, accountID string, options WatchOptions) error {
	// Set up signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		cancel()
	}()

	// Keep track of previous state
	var previousContainers []api.Container

	ticker := time.NewTicker(options.Interval)
	defer ticker.Stop()

	// Initial fetch
	containers, err := client.ListContainers(accountID)
	if err != nil {
		return fmt.Errorf("initial containers fetch: %w", err)
	}

	if !options.EventsOnly {
		// Output all containers initially
		for _, c := range containers {
			event := Event{
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				Type:      "snapshot",
				Resource:  "container",
				Name:      c.Name,
				Data:      c,
			}
			outputEvent(event)
		}
	}

	previousContainers = containers

	// Watch loop
	for {
		select {
		case <-ctx.Done():
			// Graceful shutdown
			return nil
		case <-ticker.C:
			containers, err := client.ListContainers(accountID)
			if err != nil {
				// Log error as event
				event := Event{
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					Type:      "error",
					Resource:  "container",
					Change:    fmt.Sprintf("fetch error: %v", err),
				}
				outputEvent(event)
				continue
			}

			// Detect changes
			events := detectContainerChanges(previousContainers, containers)
			for _, event := range events {
				outputEvent(event)
			}

			if !options.EventsOnly && len(events) == 0 {
				// Output all containers even if no changes (for periodic snapshots)
				event := Event{
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					Type:      "heartbeat",
					Resource:  "container",
					Change:    fmt.Sprintf("checked %d containers", len(containers)),
				}
				outputEvent(event)
			}

			previousContainers = containers
		}
	}
}

// WatchWorkers continuously monitors workers for changes
func WatchWorkers(client *api.Client, accountID string, options WatchOptions) error {
	// Set up signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		cancel()
	}()

	// Keep track of previous state
	var previousWorkers []api.Worker

	ticker := time.NewTicker(options.Interval)
	defer ticker.Stop()

	// Initial fetch
	workers, err := client.ListWorkers(accountID)
	if err != nil {
		return fmt.Errorf("initial workers fetch: %w", err)
	}

	if !options.EventsOnly {
		// Output all workers initially
		for _, w := range workers {
			event := Event{
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				Type:      "snapshot",
				Resource:  "worker",
				Name:      w.Name,
				Data:      w,
			}
			outputEvent(event)
		}
	}

	previousWorkers = workers

	// Watch loop
	for {
		select {
		case <-ctx.Done():
			// Graceful shutdown
			return nil
		case <-ticker.C:
			workers, err := client.ListWorkers(accountID)
			if err != nil {
				// Log error as event
				event := Event{
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					Type:      "error",
					Resource:  "worker",
					Change:    fmt.Sprintf("fetch error: %v", err),
				}
				outputEvent(event)
				continue
			}

			// Detect changes
			events := detectWorkerChanges(previousWorkers, workers)
			for _, event := range events {
				outputEvent(event)
			}

			if !options.EventsOnly && len(events) == 0 {
				// Output heartbeat
				event := Event{
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					Type:      "heartbeat",
					Resource:  "worker",
					Change:    fmt.Sprintf("checked %d workers", len(workers)),
				}
				outputEvent(event)
			}

			previousWorkers = workers
		}
	}
}

// detectContainerChanges compares two container lists and returns events for changes
func detectContainerChanges(previous, current []api.Container) []Event {
	events := []Event{}
	timestamp := time.Now().UTC().Format(time.RFC3339)

	// Create maps for easier lookup
	prevMap := make(map[string]api.Container)
	for _, c := range previous {
		prevMap[c.ID] = c
	}

	currMap := make(map[string]api.Container)
	for _, c := range current {
		currMap[c.ID] = c
	}

	// Check for new and changed containers
	for id, curr := range currMap {
		if prev, exists := prevMap[id]; !exists {
			// New container
			events = append(events, Event{
				Timestamp: timestamp,
				Type:      "added",
				Resource:  "container",
				Name:      curr.Name,
				Change:    "new container detected",
				Data:      curr,
			})
		} else {
			// Check for status changes
			if prev.Status != curr.Status {
				events = append(events, Event{
					Timestamp: timestamp,
					Type:      "modified",
					Resource:  "container",
					Name:      curr.Name,
					Change:    fmt.Sprintf("status changed from %s to %s", prev.Status, curr.Status),
					Data:      curr,
				})
			}
			// Check for significant metric changes (>10% change)
			if prev.CPUMS > 0 && abs(curr.CPUMS-prev.CPUMS)*100/prev.CPUMS > 10 {
				events = append(events, Event{
					Timestamp: timestamp,
					Type:      "modified",
					Resource:  "container",
					Name:      curr.Name,
					Change:    fmt.Sprintf("CPU changed from %d to %d ms", prev.CPUMS, curr.CPUMS),
					Data:      curr,
				})
			}
		}
	}

	// Check for removed containers
	for id, prev := range prevMap {
		if _, exists := currMap[id]; !exists {
			events = append(events, Event{
				Timestamp: timestamp,
				Type:      "removed",
				Resource:  "container",
				Name:      prev.Name,
				Change:    "container removed",
			})
		}
	}

	return events
}

// detectWorkerChanges compares two worker lists and returns events for changes
func detectWorkerChanges(previous, current []api.Worker) []Event {
	events := []Event{}
	timestamp := time.Now().UTC().Format(time.RFC3339)

	// Create maps for easier lookup
	prevMap := make(map[string]api.Worker)
	for _, w := range previous {
		prevMap[w.ID] = w
	}

	currMap := make(map[string]api.Worker)
	for _, w := range current {
		currMap[w.ID] = w
	}

	// Check for new and changed workers
	for id, curr := range currMap {
		if prev, exists := prevMap[id]; !exists {
			// New worker
			events = append(events, Event{
				Timestamp: timestamp,
				Type:      "added",
				Resource:  "worker",
				Name:      curr.Name,
				Change:    "new worker detected",
				Data:      curr,
			})
		} else {
			// Check for status changes
			if prev.Status != curr.Status {
				events = append(events, Event{
					Timestamp: timestamp,
					Type:      "modified",
					Resource:  "worker",
					Name:      curr.Name,
					Change:    fmt.Sprintf("status changed from %s to %s", prev.Status, curr.Status),
					Data:      curr,
				})
			}
			// Check for error rate changes
			if prev.Errors != curr.Errors {
				events = append(events, Event{
					Timestamp: timestamp,
					Type:      "modified",
					Resource:  "worker",
					Name:      curr.Name,
					Change:    fmt.Sprintf("errors changed from %d to %d", prev.Errors, curr.Errors),
					Data:      curr,
				})
			}
		}
	}

	// Check for removed workers
	for id, prev := range prevMap {
		if _, exists := currMap[id]; !exists {
			events = append(events, Event{
				Timestamp: timestamp,
				Type:      "removed",
				Resource:  "worker",
				Name:      prev.Name,
				Change:    "worker removed",
			})
		}
	}

	return events
}

// outputEvent outputs an event as JSON line
func outputEvent(event Event) {
	data, err := json.Marshal(event)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling event: %v\n", err)
		return
	}
	fmt.Println(string(data))
}

// abs returns the absolute value of an integer
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
