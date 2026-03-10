package dashboard

import (
	"fmt"
	"strings"
	"testing"

	"github.com/PeterHiroshi/cfmon/internal/api"
)

func TestRenderWorkersEmpty(t *testing.T) {
	m := Model{width: 80, height: 24, data: &DashboardData{}}
	result := m.renderWorkers()
	if !strings.Contains(result, "No workers found") {
		t.Errorf("empty workers should show 'No workers found', got: %s", result)
	}
}

func TestRenderWorkersTable(t *testing.T) {
	m := Model{
		width:  100,
		height: 24,
		data: &DashboardData{
			Workers: []api.Worker{
				{Name: "api-worker", Status: "active", Requests: 1000, Errors: 5, CPUMS: 12},
				{Name: "auth-svc", Status: "stopped", Requests: 500, Errors: 50, CPUMS: 8},
			},
		},
	}
	result := m.renderWorkers()

	if !strings.Contains(result, "Name") {
		t.Error("should contain Name header")
	}
	if !strings.Contains(result, "Status") {
		t.Error("should contain Status header")
	}
	if !strings.Contains(result, "Requests") {
		t.Error("should contain Requests header")
	}
	if !strings.Contains(result, "Errors") {
		t.Error("should contain Errors header")
	}
	if !strings.Contains(result, "Error Rate") {
		t.Error("should contain Error Rate header")
	}
	if !strings.Contains(result, "CPU (ms)") {
		t.Error("should contain CPU (ms) header")
	}
	if !strings.Contains(result, "api-worker") {
		t.Error("should contain worker name 'api-worker'")
	}
	if !strings.Contains(result, "auth-svc") {
		t.Error("should contain worker name 'auth-svc'")
	}
	// Totals row: 1000+500=1500, 5+50=55
	if !strings.Contains(result, "1500") {
		t.Error("should contain total requests 1500")
	}
	if !strings.Contains(result, "55") {
		t.Error("should contain total errors 55")
	}
}

func TestRenderWorkersErrorRate(t *testing.T) {
	m := Model{
		width:  100,
		height: 24,
		data: &DashboardData{
			Workers: []api.Worker{
				{Name: "w1", Status: "active", Requests: 100, Errors: 0, CPUMS: 5},
			},
		},
	}
	result := m.renderWorkers()
	if !strings.Contains(result, "0.0%") {
		t.Errorf("0 errors should show 0.0%%, got: %s", result)
	}
}

func TestRenderWorkersSelectedRowHighlight(t *testing.T) {
	m := Model{
		width:       100,
		height:      24,
		selectedRow: 1,
		data: &DashboardData{
			Workers: []api.Worker{
				{Name: "worker-a", Status: "active", Requests: 100, Errors: 0, CPUMS: 5},
				{Name: "worker-b", Status: "active", Requests: 200, Errors: 1, CPUMS: 8},
			},
		},
	}
	result := m.renderWorkers()
	if !strings.Contains(result, "worker-b") {
		t.Error("selected row worker-b should be rendered")
	}
	if !strings.Contains(result, "worker-a") {
		t.Error("non-selected row worker-a should be rendered")
	}
}

func TestRenderWorkersScroll(t *testing.T) {
	workers := make([]api.Worker, 30)
	for i := range workers {
		workers[i] = api.Worker{Name: fmt.Sprintf("worker-%02d", i), Status: "active", Requests: 10}
	}
	m := Model{
		width:        100,
		height:       15,
		scrollOffset: 5,
		data:         &DashboardData{Workers: workers},
	}
	result := m.renderWorkers()
	if !strings.Contains(result, "worker-05") {
		t.Errorf("scrolled view should contain worker-05, got: %s", result)
	}
}
