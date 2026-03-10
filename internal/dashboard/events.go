package dashboard

import (
	"fmt"

	"github.com/PeterHiroshi/cfmon/internal/monitor"
)

const maxEvents = 50

func alertKey(a monitor.Alert) string {
	return fmt.Sprintf("%s:%s:%s", a.ResourceType, a.ResourceName, a.Metric)
}

func diffAlerts(prev, curr []monitor.Alert) []DashboardEvent {
	prevMap := make(map[string]monitor.Alert, len(prev))
	for _, a := range prev {
		prevMap[alertKey(a)] = a
	}
	currMap := make(map[string]monitor.Alert, len(curr))
	for _, a := range curr {
		currMap[alertKey(a)] = a
	}

	var events []DashboardEvent

	// New alerts
	for key, a := range currMap {
		if _, existed := prevMap[key]; !existed {
			events = append(events, DashboardEvent{
				Severity: a.Severity,
				Text:     fmt.Sprintf("NEW: [%s] %s %s %s %.2f%%", severityLabel(a.Severity), a.ResourceType, a.ResourceName, a.Metric, a.Value),
			})
		}
	}

	// Cleared alerts
	for key, a := range prevMap {
		if _, exists := currMap[key]; !exists {
			events = append(events, DashboardEvent{
				Severity: "success",
				Text:     fmt.Sprintf("CLEARED: %s %s %s", a.ResourceType, a.ResourceName, a.Metric),
			})
		}
	}

	return events
}

func severityLabel(s string) string {
	switch s {
	case "critical":
		return "CRIT"
	case "warning":
		return "WARN"
	default:
		return "INFO"
	}
}

func addEvent(events []DashboardEvent, e DashboardEvent) []DashboardEvent {
	events = append(events, e)
	if len(events) > maxEvents {
		events = events[len(events)-maxEvents:]
	}
	return events
}
