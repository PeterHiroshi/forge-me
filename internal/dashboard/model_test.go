package dashboard

import (
	"fmt"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/PeterHiroshi/cfmon/internal/api"
)

func TestNewModel(t *testing.T) {
	client := api.NewClient("test-token")
	m := NewModel(client, "test-account", 30*time.Second)

	if m.activeTab != TabOverview {
		t.Errorf("activeTab = %d, want TabOverview", m.activeTab)
	}
	if m.accountID != "test-account" {
		t.Errorf("accountID = %q, want %q", m.accountID, "test-account")
	}
	if !m.loading {
		t.Error("expected loading=true on new model")
	}
	if m.refreshInterval != 30*time.Second {
		t.Errorf("refreshInterval = %v, want 30s", m.refreshInterval)
	}
}

func TestNewModelClampsMinInterval(t *testing.T) {
	client := api.NewClient("test-token")
	m := NewModel(client, "test-account", 2*time.Second)

	if m.refreshInterval != MinRefreshInterval {
		t.Errorf("refreshInterval = %v, want %v", m.refreshInterval, MinRefreshInterval)
	}
}

func TestUpdateTabSwitch(t *testing.T) {
	client := api.NewClient("test-token")
	m := NewModel(client, "test-account", 30*time.Second)

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	updated := newModel.(Model)
	if updated.activeTab != TabWorkers {
		t.Errorf("after pressing '2': activeTab = %d, want TabWorkers(%d)", updated.activeTab, TabWorkers)
	}

	newModel, _ = updated.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
	updated = newModel.(Model)
	if updated.activeTab != TabContainers {
		t.Errorf("after pressing '3': activeTab = %d, want TabContainers(%d)", updated.activeTab, TabContainers)
	}

	newModel, _ = updated.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	updated = newModel.(Model)
	if updated.activeTab != TabOverview {
		t.Errorf("after pressing '1': activeTab = %d, want TabOverview(%d)", updated.activeTab, TabOverview)
	}
}

func TestUpdateTabCycle(t *testing.T) {
	client := api.NewClient("test-token")
	m := NewModel(client, "test-account", 30*time.Second)

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	updated := newModel.(Model)
	if updated.activeTab != TabWorkers {
		t.Errorf("after Tab: activeTab = %d, want TabWorkers", updated.activeTab)
	}

	newModel, _ = updated.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	updated = newModel.(Model)
	if updated.activeTab != TabOverview {
		t.Errorf("after Shift+Tab: activeTab = %d, want TabOverview", updated.activeTab)
	}

	newModel, _ = updated.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	updated = newModel.(Model)
	if updated.activeTab != TabContainers {
		t.Errorf("after Shift+Tab wrap: activeTab = %d, want TabContainers", updated.activeTab)
	}
}

func TestUpdateQuit(t *testing.T) {
	client := api.NewClient("test-token")
	m := NewModel(client, "test-account", 30*time.Second)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Error("expected quit command, got nil")
	}
}

func TestUpdateDataMsg(t *testing.T) {
	client := api.NewClient("test-token")
	m := NewModel(client, "test-account", 30*time.Second)

	data := &DashboardData{
		HealthScore:  85,
		HealthStatus: "good",
		Workers:      []api.Worker{{Name: "w1"}},
		Containers:   []api.Container{{Name: "c1"}},
	}

	newModel, _ := m.Update(dataMsg{data: data})
	updated := newModel.(Model)

	if updated.loading {
		t.Error("expected loading=false after dataMsg")
	}
	if updated.data == nil {
		t.Fatal("expected non-nil data after dataMsg")
	}
	if updated.data.HealthScore != 85 {
		t.Errorf("HealthScore = %d, want 85", updated.data.HealthScore)
	}
	if updated.err != nil {
		t.Errorf("expected nil error, got %v", updated.err)
	}
}

func TestUpdateErrMsg(t *testing.T) {
	client := api.NewClient("test-token")
	m := NewModel(client, "test-account", 30*time.Second)

	newModel, _ := m.Update(errMsg{err: fmt.Errorf("fetch failed")})
	updated := newModel.(Model)

	if updated.loading {
		t.Error("expected loading=false after errMsg")
	}
	if updated.err == nil {
		t.Error("expected non-nil error after errMsg")
	}
}

func TestUpdateWindowSizeMsg(t *testing.T) {
	client := api.NewClient("test-token")
	m := NewModel(client, "test-account", 30*time.Second)

	newModel, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	updated := newModel.(Model)

	if updated.width != 120 {
		t.Errorf("width = %d, want 120", updated.width)
	}
	if updated.height != 40 {
		t.Errorf("height = %d, want 40", updated.height)
	}
}

func TestViewContainsTabNames(t *testing.T) {
	client := api.NewClient("test-token")
	m := NewModel(client, "test-account", 30*time.Second)
	m.width = 80
	m.height = 24

	view := m.View()
	if !strings.Contains(view, "Overview") {
		t.Error("view should contain 'Overview' tab")
	}
	if !strings.Contains(view, "Workers") {
		t.Error("view should contain 'Workers' tab")
	}
	if !strings.Contains(view, "Containers") {
		t.Error("view should contain 'Containers' tab")
	}
}

func TestViewOverviewWithData(t *testing.T) {
	client := api.NewClient("test-token")
	m := NewModel(client, "test-account", 30*time.Second)
	m.width = 80
	m.height = 24
	m.loading = false
	m.data = &DashboardData{
		HealthScore:  92,
		HealthStatus: "excellent",
		Workers:      []api.Worker{{Name: "w1"}, {Name: "w2"}},
		Containers:   []api.Container{{Name: "c1"}},
		ErrorRate:    1.5,
	}

	view := m.View()
	if !strings.Contains(view, "92") {
		t.Error("overview should display health score 92")
	}
}
