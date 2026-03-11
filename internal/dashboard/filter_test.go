package dashboard

import (
	"strings"
	"testing"

	"github.com/PeterHiroshi/cfmon/internal/api"
	"github.com/PeterHiroshi/cfmon/internal/monitor"
)

func TestFilterWorkersByName(t *testing.T) {
	workers := []api.Worker{
		{Name: "api-gateway", Status: "active"},
		{Name: "auth-service", Status: "active"},
		{Name: "api-internal", Status: "stopped"},
	}
	result := filterWorkers(workers, "api")
	if len(result) != 2 {
		t.Errorf("expected 2 matches for 'api', got %d", len(result))
	}
}

func TestFilterWorkersCaseInsensitive(t *testing.T) {
	workers := []api.Worker{
		{Name: "API-Gateway", Status: "active"},
		{Name: "auth-service", Status: "active"},
	}
	result := filterWorkers(workers, "api")
	if len(result) != 1 {
		t.Errorf("expected 1 match for 'api' (case-insensitive), got %d", len(result))
	}
}

func TestFilterWorkersEmptyText(t *testing.T) {
	workers := []api.Worker{
		{Name: "w1", Status: "active"},
		{Name: "w2", Status: "active"},
	}
	result := filterWorkers(workers, "")
	if len(result) != 2 {
		t.Errorf("empty filter should return all, got %d", len(result))
	}
}

func TestFilterContainersByName(t *testing.T) {
	containers := []api.Container{
		{Name: "web-app", Status: "running"},
		{Name: "worker-bg", Status: "running"},
		{Name: "web-admin", Status: "stopped"},
	}
	result := filterContainers(containers, "web")
	if len(result) != 2 {
		t.Errorf("expected 2 matches for 'web', got %d", len(result))
	}
}

func TestFilterContainersCaseInsensitive(t *testing.T) {
	containers := []api.Container{
		{Name: "Web-App", Status: "running"},
		{Name: "worker-bg", Status: "running"},
	}
	result := filterContainers(containers, "WEB")
	if len(result) != 1 {
		t.Errorf("expected 1 match for 'WEB', got %d", len(result))
	}
}

func TestFilterContainersEmptyText(t *testing.T) {
	containers := []api.Container{
		{Name: "c1", Status: "running"},
		{Name: "c2", Status: "running"},
	}
	result := filterContainers(containers, "")
	if len(result) != 2 {
		t.Errorf("empty filter should return all, got %d", len(result))
	}
}

func TestFilterNoMatches(t *testing.T) {
	workers := []api.Worker{
		{Name: "api-gw", Status: "active"},
	}
	result := filterWorkers(workers, "zzz")
	if len(result) != 0 {
		t.Errorf("expected 0 matches for 'zzz', got %d", len(result))
	}
}

func TestRenderWorkersWithFilterShowsFilteredRows(t *testing.T) {
	m := Model{
		width:      100,
		height:     24,
		activeTab:  TabWorkers,
		filterText: "api",
		data: &DashboardData{
			Workers: []api.Worker{
				{Name: "api-gateway", Status: "active", Requests: 100, Errors: 0, CPUMS: 5},
				{Name: "auth-service", Status: "active", Requests: 200, Errors: 1, CPUMS: 8},
				{Name: "api-internal", Status: "stopped", Requests: 50, Errors: 0, CPUMS: 3},
			},
		},
	}
	result := m.renderWorkers()
	if !strings.Contains(result, "api-gateway") {
		t.Error("filtered view should contain api-gateway")
	}
	if strings.Contains(result, "auth-service") {
		t.Error("filtered view should NOT contain auth-service")
	}
	if !strings.Contains(result, "api-internal") {
		t.Error("filtered view should contain api-internal")
	}
}

func TestFilterAlerts(t *testing.T) {
	alerts := []monitor.Alert{
		{ResourceName: "api-proxy", Metric: "error_rate", Severity: "warning"},
		{ResourceName: "db-svc", Metric: "memory", Severity: "critical"},
		{ResourceName: "web-app", Metric: "cpu", Severity: "warning"},
	}

	// Filter by resource name
	result := filterAlerts(alerts, "api")
	if len(result) != 1 {
		t.Errorf("filter 'api': got %d alerts, want 1", len(result))
	}

	// Filter by metric name
	result = filterAlerts(alerts, "memory")
	if len(result) != 1 {
		t.Errorf("filter 'memory': got %d alerts, want 1", len(result))
	}

	// Empty filter returns all
	result = filterAlerts(alerts, "")
	if len(result) != 3 {
		t.Errorf("empty filter: got %d alerts, want 3", len(result))
	}

	// No match
	result = filterAlerts(alerts, "xyz")
	if len(result) != 0 {
		t.Errorf("filter 'xyz': got %d alerts, want 0", len(result))
	}
}
