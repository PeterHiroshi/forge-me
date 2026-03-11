package dashboard

import (
	"testing"
	"time"

	"github.com/PeterHiroshi/cfmon/internal/monitor"
)

func TestAlertKey(t *testing.T) {
	a := monitor.Alert{ResourceType: "worker", ResourceName: "api", Metric: "error_rate"}
	if got := alertKey(a); got != "worker:api:error_rate" {
		t.Errorf("alertKey = %q, want %q", got, "worker:api:error_rate")
	}
}

func TestDiffAlertsNewAlert(t *testing.T) {
	prev := []monitor.Alert{}
	curr := []monitor.Alert{
		{Severity: "warning", ResourceType: "worker", ResourceName: "api", Metric: "error_rate", Value: 4.2},
	}
	events := diffAlerts(prev, curr)
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Severity != "warning" {
		t.Errorf("severity = %q, want warning", events[0].Severity)
	}
}

func TestDiffAlertsClearedAlert(t *testing.T) {
	prev := []monitor.Alert{
		{Severity: "warning", ResourceType: "worker", ResourceName: "api", Metric: "error_rate"},
	}
	curr := []monitor.Alert{}
	events := diffAlerts(prev, curr)
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Severity != "success" {
		t.Errorf("severity = %q, want success", events[0].Severity)
	}
}

func TestDiffAlertsNoChange(t *testing.T) {
	alerts := []monitor.Alert{
		{ResourceType: "worker", ResourceName: "api", Metric: "error_rate"},
	}
	events := diffAlerts(alerts, alerts)
	if len(events) != 0 {
		t.Errorf("expected 0 events for no change, got %d", len(events))
	}
}

func TestAddEventRingBuffer(t *testing.T) {
	events := make([]DashboardEvent, 0, maxEvents)
	for i := 0; i < maxEvents+10; i++ {
		events = addEvent(events, DashboardEvent{
			Time: time.Now(), Text: "event", Severity: "info",
		})
	}
	if len(events) != maxEvents {
		t.Errorf("ring buffer len = %d, want %d", len(events), maxEvents)
	}
}
