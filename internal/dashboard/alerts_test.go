package dashboard

import (
	"strings"
	"testing"
	"time"

	"github.com/PeterHiroshi/cfmon/internal/monitor"
)

func TestRenderAlertsEmpty(t *testing.T) {
	m := Model{
		width: 80, height: 24,
		activeTab: TabAlerts,
		data: &DashboardData{
			Alerts: []monitor.Alert{},
		},
	}
	result := m.renderAlerts()
	if !strings.Contains(result, "All systems healthy") {
		t.Errorf("empty alerts should show healthy message, got: %s", result)
	}
}

func TestRenderAlertsSummary(t *testing.T) {
	m := Model{
		width: 80, height: 24,
		activeTab: TabAlerts,
		data: &DashboardData{
			Alerts: []monitor.Alert{
				{Severity: "warning", ResourceName: "api", Metric: "error_rate", Value: 4.2, Threshold: 2.0},
				{Severity: "critical", ResourceName: "db", Metric: "memory", Value: 92.3, Threshold: 85.0},
			},
		},
	}
	result := m.renderAlerts()
	if !strings.Contains(result, "1 warning") {
		t.Errorf("should show warning count, got: %s", result)
	}
	if !strings.Contains(result, "1 critical") {
		t.Errorf("should show critical count, got: %s", result)
	}
}

func TestRenderAlertsRows(t *testing.T) {
	m := Model{
		width: 100, height: 24,
		activeTab: TabAlerts,
		data: &DashboardData{
			Alerts: []monitor.Alert{
				{Severity: "warning", ResourceType: "worker", ResourceName: "api-proxy", Metric: "error_rate", Value: 4.2, Threshold: 2.0},
			},
		},
	}
	result := m.renderAlerts()
	if !strings.Contains(result, "WARN") {
		t.Errorf("should show WARN label, got: %s", result)
	}
	if !strings.Contains(result, "api-proxy") {
		t.Errorf("should show resource name, got: %s", result)
	}
}

func TestRenderAlertsSelectedRow(t *testing.T) {
	m := Model{
		width: 100, height: 24,
		activeTab:   TabAlerts,
		selectedRow: 1,
		data: &DashboardData{
			Alerts: []monitor.Alert{
				{Severity: "warning", ResourceName: "a"},
				{Severity: "critical", ResourceName: "b"},
			},
		},
	}
	result := m.renderAlerts()
	if !strings.Contains(result, "a") || !strings.Contains(result, "b") {
		t.Errorf("should render both alerts, got: %s", result)
	}
}

func TestRenderAlertsWithFilter(t *testing.T) {
	m := Model{
		width: 100, height: 24,
		activeTab:  TabAlerts,
		filterText: "api",
		data: &DashboardData{
			Alerts: []monitor.Alert{
				{Severity: "warning", ResourceName: "api-proxy", Metric: "error_rate"},
				{Severity: "critical", ResourceName: "db-svc", Metric: "memory"},
			},
		},
	}
	result := m.renderAlerts()
	if !strings.Contains(result, "api-proxy") {
		t.Errorf("filtered alerts should show api-proxy, got: %s", result)
	}
	if strings.Contains(result, "db-svc") {
		t.Errorf("filtered alerts should not show db-svc, got: %s", result)
	}
}

func TestRenderAlertsEventLog(t *testing.T) {
	now := time.Now()
	m := Model{
		width: 100, height: 40,
		activeTab: TabAlerts,
		data: &DashboardData{
			Alerts: []monitor.Alert{},
		},
		events: []DashboardEvent{
			{Time: now, Text: "Data refreshed — 5 workers, 3 containers", Severity: "info"},
		},
	}
	result := m.renderAlerts()
	if !strings.Contains(result, "Event Log") {
		t.Errorf("alerts tab should show event log, got: %s", result)
	}
	if !strings.Contains(result, "Data refreshed") {
		t.Errorf("event log should show refresh event, got: %s", result)
	}
}
