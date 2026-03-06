package watch

import (
	"testing"

	"github.com/PeterHiroshi/cfmon/internal/api"
)

func TestDetectContainerChanges_NewContainer(t *testing.T) {
	previous := []api.Container{}
	current := []api.Container{
		{ID: "1", Name: "container1", Status: "running"},
	}

	events := detectContainerChanges(previous, current)

	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}

	if events[0].Type != "added" {
		t.Errorf("Expected event type 'added', got '%s'", events[0].Type)
	}

	if events[0].Name != "container1" {
		t.Errorf("Expected name 'container1', got '%s'", events[0].Name)
	}
}

func TestDetectContainerChanges_RemovedContainer(t *testing.T) {
	previous := []api.Container{
		{ID: "1", Name: "container1", Status: "running"},
	}
	current := []api.Container{}

	events := detectContainerChanges(previous, current)

	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}

	if events[0].Type != "removed" {
		t.Errorf("Expected event type 'removed', got '%s'", events[0].Type)
	}
}

func TestDetectContainerChanges_StatusChange(t *testing.T) {
	previous := []api.Container{
		{ID: "1", Name: "container1", Status: "running"},
	}
	current := []api.Container{
		{ID: "1", Name: "container1", Status: "stopped"},
	}

	events := detectContainerChanges(previous, current)

	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}

	if events[0].Type != "modified" {
		t.Errorf("Expected event type 'modified', got '%s'", events[0].Type)
	}
}

func TestDetectContainerChanges_NoChanges(t *testing.T) {
	previous := []api.Container{
		{ID: "1", Name: "container1", Status: "running", CPUMS: 100},
	}
	current := []api.Container{
		{ID: "1", Name: "container1", Status: "running", CPUMS: 100},
	}

	events := detectContainerChanges(previous, current)

	if len(events) != 0 {
		t.Errorf("Expected 0 events, got %d", len(events))
	}
}

func TestDetectContainerChanges_CPUChange(t *testing.T) {
	previous := []api.Container{
		{ID: "1", Name: "container1", Status: "running", CPUMS: 100},
	}
	current := []api.Container{
		{ID: "1", Name: "container1", Status: "running", CPUMS: 120}, // 20% increase
	}

	events := detectContainerChanges(previous, current)

	if len(events) != 1 {
		t.Errorf("Expected 1 event for significant CPU change, got %d", len(events))
	}

	if events[0].Type != "modified" {
		t.Errorf("Expected event type 'modified', got '%s'", events[0].Type)
	}
}

func TestDetectWorkerChanges_NewWorker(t *testing.T) {
	previous := []api.Worker{}
	current := []api.Worker{
		{ID: "1", Name: "worker1", Status: "active"},
	}

	events := detectWorkerChanges(previous, current)

	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}

	if events[0].Type != "added" {
		t.Errorf("Expected event type 'added', got '%s'", events[0].Type)
	}
}

func TestDetectWorkerChanges_RemovedWorker(t *testing.T) {
	previous := []api.Worker{
		{ID: "1", Name: "worker1", Status: "active"},
	}
	current := []api.Worker{}

	events := detectWorkerChanges(previous, current)

	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}

	if events[0].Type != "removed" {
		t.Errorf("Expected event type 'removed', got '%s'", events[0].Type)
	}
}

func TestDetectWorkerChanges_ErrorCountChange(t *testing.T) {
	previous := []api.Worker{
		{ID: "1", Name: "worker1", Errors: 5},
	}
	current := []api.Worker{
		{ID: "1", Name: "worker1", Errors: 10},
	}

	events := detectWorkerChanges(previous, current)

	if len(events) != 1 {
		t.Errorf("Expected 1 event for error count change, got %d", len(events))
	}

	if events[0].Type != "modified" {
		t.Errorf("Expected event type 'modified', got '%s'", events[0].Type)
	}
}

func TestDetectWorkerChanges_NoChanges(t *testing.T) {
	previous := []api.Worker{
		{ID: "1", Name: "worker1", Status: "active", Errors: 5},
	}
	current := []api.Worker{
		{ID: "1", Name: "worker1", Status: "active", Errors: 5},
	}

	events := detectWorkerChanges(previous, current)

	if len(events) != 0 {
		t.Errorf("Expected 0 events, got %d", len(events))
	}
}

func TestAbs(t *testing.T) {
	tests := []struct {
		input int
		want  int
	}{
		{5, 5},
		{-5, 5},
		{0, 0},
		{100, 100},
		{-100, 100},
	}

	for _, tt := range tests {
		got := abs(tt.input)
		if got != tt.want {
			t.Errorf("abs(%d) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestEvent_Structure(t *testing.T) {
	event := Event{
		Timestamp: "2024-01-01T00:00:00Z",
		Type:      "added",
		Resource:  "container",
		Name:      "test-container",
		Change:    "new container detected",
		Data:      map[string]string{"id": "123"},
	}

	if event.Timestamp != "2024-01-01T00:00:00Z" {
		t.Errorf("Timestamp = %q, want %q", event.Timestamp, "2024-01-01T00:00:00Z")
	}

	if event.Type != "added" {
		t.Errorf("Type = %q, want %q", event.Type, "added")
	}

	if event.Resource != "container" {
		t.Errorf("Resource = %q, want %q", event.Resource, "container")
	}
}

func TestWatchOptions(t *testing.T) {
	opts := WatchOptions{
		Interval:   30,
		EventsOnly: true,
	}

	if opts.Interval != 30 {
		t.Errorf("Interval = %d, want 30", opts.Interval)
	}

	if !opts.EventsOnly {
		t.Errorf("EventsOnly = %v, want true", opts.EventsOnly)
	}
}
