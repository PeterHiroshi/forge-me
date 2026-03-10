package dashboard

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/PeterHiroshi/cfmon/internal/api"
	"github.com/PeterHiroshi/cfmon/internal/monitor"
)

func TestDetailOpenWithEnter(t *testing.T) {
	m := Model{
		width: 80, height: 24,
		activeTab:   TabWorkers,
		selectedRow: 0,
		data: &DashboardData{
			Workers: []api.Worker{{Name: "w1", Status: "active"}},
		},
	}
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated := newModel.(Model)
	if !updated.showDetail {
		t.Error("Enter should open detail view")
	}
}

func TestDetailNotOnOverview(t *testing.T) {
	m := Model{
		width: 80, height: 24,
		activeTab: TabOverview,
	}
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated := newModel.(Model)
	if updated.showDetail {
		t.Error("Enter on Overview should not open detail")
	}
}

func TestDetailCloseWithEsc(t *testing.T) {
	m := Model{
		width: 80, height: 24,
		activeTab:  TabWorkers,
		showDetail: true,
		data: &DashboardData{
			Workers: []api.Worker{{Name: "w1", Status: "active"}},
		},
	}
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	updated := newModel.(Model)
	if updated.showDetail {
		t.Error("Esc should close detail view")
	}
}

func TestDetailTabSwitchClosesDetail(t *testing.T) {
	m := Model{
		width: 80, height: 24,
		activeTab:  TabWorkers,
		showDetail: true,
		data: &DashboardData{
			Workers: []api.Worker{{Name: "w1", Status: "active"}},
		},
	}
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	updated := newModel.(Model)
	if updated.showDetail {
		t.Error("Tab switch should close detail view")
	}
	if updated.activeTab != TabContainers {
		t.Errorf("should switch to Containers tab, got %d", updated.activeTab)
	}
}

func TestRenderWorkerDetail(t *testing.T) {
	m := Model{
		width: 80, height: 24,
		activeTab:   TabWorkers,
		selectedRow: 0,
		showDetail:  true,
		data: &DashboardData{
			Workers: []api.Worker{
				{ID: "w-123", Name: "api-gateway", Status: "active", Requests: 1000, Errors: 5, CPUMS: 12, SuccessRate: 99.5},
			},
		},
	}
	result := m.renderWorkerDetail()
	checks := []string{"api-gateway", "w-123", "active", "1000", "12"}
	for _, c := range checks {
		if !strings.Contains(result, c) {
			t.Errorf("worker detail should contain %q", c)
		}
	}
}

func TestRenderContainerDetail(t *testing.T) {
	m := Model{
		width: 80, height: 24,
		activeTab:   TabContainers,
		selectedRow: 0,
		showDetail:  true,
		data: &DashboardData{
			Containers: []api.Container{
				{ID: "c-456", Name: "web-app", Status: "running", CPUMS: 500, MemoryMB: 64, Requests: 2000},
			},
		},
	}
	result := m.renderContainerDetail()
	checks := []string{"web-app", "c-456", "running", "500", "64", "2000"}
	for _, c := range checks {
		if !strings.Contains(result, c) {
			t.Errorf("container detail should contain %q", c)
		}
	}
}

func TestDetailSuppressesJK(t *testing.T) {
	m := Model{
		width: 80, height: 24,
		activeTab:   TabWorkers,
		showDetail:  true,
		selectedRow: 2,
		data: &DashboardData{
			Workers: make([]api.Worker, 10),
		},
	}
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	updated := newModel.(Model)
	if updated.selectedRow != 2 {
		t.Errorf("j should not move selectedRow in detail view, got %d", updated.selectedRow)
	}
}

func TestRenderAlertDetail(t *testing.T) {
	m := Model{
		width: 80, height: 24,
		activeTab:   TabAlerts,
		selectedRow: 0,
		showDetail:  true,
		data: &DashboardData{
			Alerts: []monitor.Alert{
				{
					Severity:     "warning",
					ResourceType: "worker",
					ResourceName: "api-proxy",
					Metric:       "error_rate",
					Value:        4.2,
					Threshold:    2.0,
					Message:      "Worker api-proxy has high error rate: 4.20%",
				},
			},
			Workers: []api.Worker{
				{Name: "api-proxy", ID: "w-123", Status: "active", Requests: 1000, Errors: 42},
			},
		},
	}
	result := m.renderAlertDetail()
	checks := []string{"warning", "worker", "api-proxy", "error_rate", "4.2", "2.0"}
	for _, c := range checks {
		if !strings.Contains(result, c) {
			t.Errorf("alert detail should contain %q, got: %s", c, result)
		}
	}
}

func TestRenderAlertDetailWithRelatedWorker(t *testing.T) {
	m := Model{
		width: 80, height: 24,
		activeTab:   TabAlerts,
		selectedRow: 0,
		showDetail:  true,
		data: &DashboardData{
			Alerts: []monitor.Alert{
				{Severity: "warning", ResourceType: "worker", ResourceName: "api-proxy", Metric: "error_rate", Value: 4.2, Threshold: 2.0},
			},
			Workers: []api.Worker{
				{Name: "api-proxy", ID: "w-123", Status: "active", Requests: 1000, Errors: 42},
			},
		},
	}
	result := m.renderAlertDetail()
	if !strings.Contains(result, "w-123") {
		t.Errorf("alert detail should show related worker ID, got: %s", result)
	}
}

func TestRenderAlertDetailContainer(t *testing.T) {
	m := Model{
		width: 80, height: 24,
		activeTab:   TabAlerts,
		selectedRow: 0,
		showDetail:  true,
		data: &DashboardData{
			Alerts: []monitor.Alert{
				{Severity: "critical", ResourceType: "container", ResourceName: "db-svc", Metric: "memory", Value: 92.3, Threshold: 85.0},
			},
			Containers: []api.Container{
				{Name: "db-svc", ID: "c-789", Status: "running", MemoryMB: 120},
			},
		},
	}
	result := m.renderAlertDetail()
	if !strings.Contains(result, "c-789") {
		t.Errorf("alert detail should show related container ID, got: %s", result)
	}
}

func TestRenderAlertDetailNoSelection(t *testing.T) {
	m := Model{
		width: 80, height: 24,
		activeTab:   TabAlerts,
		selectedRow: 5,
		showDetail:  true,
		data: &DashboardData{
			Alerts: []monitor.Alert{},
		},
	}
	result := m.renderAlertDetail()
	if !strings.Contains(result, "No alert selected") {
		t.Errorf("should show 'No alert selected', got: %s", result)
	}
}
