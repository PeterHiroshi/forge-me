package dashboard

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/PeterHiroshi/cfmon/internal/api"
	"github.com/PeterHiroshi/cfmon/internal/monitor"
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
	if updated.activeTab != TabAlerts {
		t.Errorf("after Shift+Tab wrap: activeTab = %d, want TabAlerts", updated.activeTab)
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

func TestScrollDownUp(t *testing.T) {
	client := api.NewClient("test-token")
	m := NewModel(client, "test-account", 30*time.Second)
	m.width = 80
	m.height = 24
	m.loading = false
	m.activeTab = TabWorkers
	m.data = &DashboardData{
		Workers: make([]api.Worker, 20),
	}

	// Scroll down with j moves selectedRow
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	updated := newModel.(Model)
	if updated.selectedRow != 1 {
		t.Errorf("selectedRow after j = %d, want 1", updated.selectedRow)
	}

	// Scroll up with k moves selectedRow back
	newModel, _ = updated.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	updated = newModel.(Model)
	if updated.selectedRow != 0 {
		t.Errorf("selectedRow after k = %d, want 0", updated.selectedRow)
	}

	// Should not go below 0
	newModel, _ = updated.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	updated = newModel.(Model)
	if updated.selectedRow != 0 {
		t.Errorf("selectedRow should not go below 0, got %d", updated.selectedRow)
	}
}

func TestScrollResetOnTabSwitch(t *testing.T) {
	client := api.NewClient("test-token")
	m := NewModel(client, "test-account", 30*time.Second)
	m.width = 80
	m.height = 24
	m.loading = false
	m.activeTab = TabWorkers
	m.scrollOffset = 5
	m.data = &DashboardData{
		Workers: make([]api.Worker, 20),
	}

	// Switch tab should reset scroll
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	updated := newModel.(Model)
	if updated.scrollOffset != 0 {
		t.Errorf("scrollOffset after tab switch = %d, want 0", updated.scrollOffset)
	}
}

func TestScrollWithArrowKeys(t *testing.T) {
	client := api.NewClient("test-token")
	m := NewModel(client, "test-account", 30*time.Second)
	m.width = 80
	m.height = 24
	m.loading = false
	m.activeTab = TabContainers
	m.data = &DashboardData{
		Containers: make([]api.Container, 20),
	}

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	updated := newModel.(Model)
	if updated.selectedRow != 1 {
		t.Errorf("selectedRow after down = %d, want 1", updated.selectedRow)
	}

	newModel, _ = updated.Update(tea.KeyMsg{Type: tea.KeyUp})
	updated = newModel.(Model)
	if updated.selectedRow != 0 {
		t.Errorf("selectedRow after up = %d, want 0", updated.selectedRow)
	}
}

func TestViewWorkersTabRendered(t *testing.T) {
	client := api.NewClient("test-token")
	m := NewModel(client, "test-account", 30*time.Second)
	m.width = 100
	m.height = 24
	m.loading = false
	m.activeTab = TabWorkers
	m.data = &DashboardData{
		Workers: []api.Worker{
			{Name: "test-worker", Status: "active", Requests: 100, Errors: 2, CPUMS: 15},
		},
	}

	view := m.View()
	if !strings.Contains(view, "test-worker") {
		t.Error("Workers tab should render worker name")
	}
	if strings.Contains(view, "Phase 2") {
		t.Error("Workers tab should no longer show Phase 2 placeholder")
	}
}

func TestViewContainersTabRendered(t *testing.T) {
	client := api.NewClient("test-token")
	m := NewModel(client, "test-account", 30*time.Second)
	m.width = 120
	m.height = 24
	m.loading = false
	m.activeTab = TabContainers
	m.data = &DashboardData{
		Containers: []api.Container{
			{Name: "test-container", Status: "running", CPUMS: 300, MemoryMB: 64},
		},
	}

	view := m.View()
	if !strings.Contains(view, "test-container") {
		t.Error("Containers tab should render container name")
	}
	if strings.Contains(view, "Phase 2") {
		t.Error("Containers tab should no longer show Phase 2 placeholder")
	}
}

func TestStatusBarShowsScrollHint(t *testing.T) {
	client := api.NewClient("test-token")
	m := NewModel(client, "test-account", 30*time.Second)
	m.width = 120
	m.height = 24

	// Overview tab should NOT show scroll hint
	m.activeTab = TabOverview
	bar := m.renderStatusBar()
	if strings.Contains(bar, "j/k") {
		t.Error("Overview tab should not show j/k scroll hint")
	}

	// Workers tab SHOULD show scroll hint
	m.activeTab = TabWorkers
	bar = m.renderStatusBar()
	if !strings.Contains(bar, "j/k") {
		t.Error("Workers tab should show j/k scroll hint")
	}

	// Containers tab SHOULD show scroll hint
	m.activeTab = TabContainers
	bar = m.renderStatusBar()
	if !strings.Contains(bar, "j/k") {
		t.Error("Containers tab should show j/k scroll hint")
	}
}

func TestSelectedRowMovesWithJ(t *testing.T) {
	m := Model{
		width: 80, height: 24,
		activeTab: TabWorkers,
		data: &DashboardData{
			Workers: make([]api.Worker, 10),
		},
	}
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	updated := newModel.(Model)
	if updated.selectedRow != 1 {
		t.Errorf("selectedRow after j = %d, want 1", updated.selectedRow)
	}
}

func TestSelectedRowMovesWithK(t *testing.T) {
	m := Model{
		width: 80, height: 24,
		activeTab:   TabWorkers,
		selectedRow: 3,
		data: &DashboardData{
			Workers: make([]api.Worker, 10),
		},
	}
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	updated := newModel.(Model)
	if updated.selectedRow != 2 {
		t.Errorf("selectedRow after k = %d, want 2", updated.selectedRow)
	}
}

func TestSelectedRowBoundsLower(t *testing.T) {
	m := Model{
		width: 80, height: 24,
		activeTab:   TabWorkers,
		selectedRow: 0,
		data: &DashboardData{
			Workers: make([]api.Worker, 10),
		},
	}
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	updated := newModel.(Model)
	if updated.selectedRow != 0 {
		t.Errorf("selectedRow should not go below 0, got %d", updated.selectedRow)
	}
}

func TestSelectedRowBoundsUpper(t *testing.T) {
	m := Model{
		width: 80, height: 24,
		activeTab:   TabWorkers,
		selectedRow: 4,
		data: &DashboardData{
			Workers: make([]api.Worker, 5),
		},
	}
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	updated := newModel.(Model)
	if updated.selectedRow != 4 {
		t.Errorf("selectedRow should not exceed len-1, got %d", updated.selectedRow)
	}
}

func TestSelectedRowResetsOnTabSwitch(t *testing.T) {
	m := Model{
		width: 80, height: 24,
		activeTab:   TabWorkers,
		selectedRow: 5,
		data: &DashboardData{
			Workers: make([]api.Worker, 10),
		},
	}
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	updated := newModel.(Model)
	if updated.selectedRow != 0 {
		t.Errorf("selectedRow should reset on tab switch, got %d", updated.selectedRow)
	}
}

func TestSelectedRowAutoScroll(t *testing.T) {
	m := Model{
		width: 80, height: 15,
		activeTab:    TabWorkers,
		selectedRow:  2,
		scrollOffset: 0,
		data: &DashboardData{
			Workers: make([]api.Worker, 20),
		},
	}
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	updated := newModel.(Model)
	if updated.selectedRow != 3 {
		t.Errorf("selectedRow = %d, want 3", updated.selectedRow)
	}
	if updated.scrollOffset < 1 {
		t.Errorf("scrollOffset should auto-adjust, got %d", updated.scrollOffset)
	}
}

func TestFilterModeActivation(t *testing.T) {
	m := Model{
		width: 80, height: 24,
		activeTab:   TabWorkers,
		filterInput: textinput.New(),
		data:        &DashboardData{Workers: make([]api.Worker, 5)},
	}
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	updated := newModel.(Model)
	if !updated.filterMode {
		t.Error("pressing / on Workers tab should activate filter mode")
	}
}

func TestFilterModeNotOnOverview(t *testing.T) {
	m := Model{
		width: 80, height: 24,
		activeTab: TabOverview,
	}
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	updated := newModel.(Model)
	if updated.filterMode {
		t.Error("pressing / on Overview tab should NOT activate filter mode")
	}
}

func TestFilterModeEscCancels(t *testing.T) {
	m := Model{
		width: 80, height: 24,
		activeTab:   TabWorkers,
		filterMode:  true,
		filterText:  "api",
		filterInput: textinput.New(),
		data:        &DashboardData{Workers: make([]api.Worker, 5)},
	}
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	updated := newModel.(Model)
	if updated.filterMode {
		t.Error("Esc should exit filter mode")
	}
	if updated.filterText != "" {
		t.Error("Esc should clear filter text")
	}
}

func TestFilterModeEnterConfirms(t *testing.T) {
	m := Model{
		width: 80, height: 24,
		activeTab:   TabWorkers,
		filterMode:  true,
		filterText:  "api",
		filterInput: textinput.New(),
		data:        &DashboardData{Workers: make([]api.Worker, 5)},
	}
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated := newModel.(Model)
	if updated.filterMode {
		t.Error("Enter should exit filter mode (but keep filter)")
	}
	if updated.filterText != "api" {
		t.Errorf("Enter should keep filter text, got %q", updated.filterText)
	}
}

func TestFilterModeSuppressesTabSwitch(t *testing.T) {
	m := Model{
		width: 80, height: 24,
		activeTab:   TabWorkers,
		filterMode:  true,
		filterInput: textinput.New(),
		data:        &DashboardData{Workers: make([]api.Worker, 5)},
	}
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	updated := newModel.(Model)
	if updated.activeTab != TabWorkers {
		t.Error("tab switch should be suppressed in filter mode")
	}
}

func TestStatusBarShowsFilterIndicator(t *testing.T) {
	m := Model{
		width: 120, height: 24,
		activeTab:  TabWorkers,
		filterText: "prod",
		data: &DashboardData{
			Workers: []api.Worker{{Name: "prod-api", Status: "active"}},
		},
	}
	bar := m.renderStatusBar()
	if !strings.Contains(bar, "filter: prod") {
		t.Errorf("status bar should show filter indicator, got: %s", bar)
	}
}

func TestStatusBarShowsHelpHint(t *testing.T) {
	m := Model{width: 120, height: 24}
	bar := m.renderStatusBar()
	if !strings.Contains(bar, "?:help") {
		t.Errorf("status bar should show ?:help hint, got: %s", bar)
	}
}

func TestStatusBarShowsRowPosition(t *testing.T) {
	m := Model{
		width: 120, height: 24,
		activeTab:   TabWorkers,
		selectedRow: 2,
		data: &DashboardData{
			Workers: make([]api.Worker, 15),
		},
	}
	bar := m.renderStatusBar()
	if !strings.Contains(bar, "row 3/15") {
		t.Errorf("status bar should show row position, got: %s", bar)
	}
}

func TestStatusBarInDetailView(t *testing.T) {
	m := Model{
		width: 120, height: 24,
		activeTab:  TabWorkers,
		showDetail: true,
		data: &DashboardData{
			Workers: []api.Worker{{Name: "w1"}},
		},
	}
	bar := m.renderStatusBar()
	if !strings.Contains(bar, "Esc: back") {
		t.Errorf("status bar in detail view should show 'Esc: back', got: %s", bar)
	}
}

func TestMouseWheelDown(t *testing.T) {
	m := Model{
		width: 80, height: 24,
		activeTab: TabWorkers,
		data: &DashboardData{
			Workers: make([]api.Worker, 20),
		},
	}
	newModel, _ := m.Update(tea.MouseMsg{Action: tea.MouseActionPress, Button: tea.MouseButtonWheelDown})
	updated := newModel.(Model)
	if updated.selectedRow != 1 {
		t.Errorf("wheel down should move selectedRow, got %d", updated.selectedRow)
	}
}

func TestMouseWheelUp(t *testing.T) {
	m := Model{
		width: 80, height: 24,
		activeTab:   TabWorkers,
		selectedRow: 5,
		data: &DashboardData{
			Workers: make([]api.Worker, 20),
		},
	}
	newModel, _ := m.Update(tea.MouseMsg{Action: tea.MouseActionPress, Button: tea.MouseButtonWheelUp})
	updated := newModel.(Model)
	if updated.selectedRow != 4 {
		t.Errorf("wheel up should move selectedRow, got %d", updated.selectedRow)
	}
}

func TestTabSwitchTo4(t *testing.T) {
	m := NewModel(api.NewClient("t"), "acc", 30*time.Second)
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'4'}})
	updated := newModel.(Model)
	if updated.activeTab != TabAlerts {
		t.Errorf("pressing '4': activeTab = %d, want TabAlerts(%d)", updated.activeTab, TabAlerts)
	}
}

func TestTabCycleIncludesAlerts(t *testing.T) {
	m := NewModel(api.NewClient("t"), "acc", 30*time.Second)
	m.activeTab = TabContainers
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	updated := newModel.(Model)
	if updated.activeTab != TabAlerts {
		t.Errorf("Tab from Containers should go to Alerts, got %d", updated.activeTab)
	}
}

func TestAlertsTabSupportsScrolling(t *testing.T) {
	m := Model{
		width: 80, height: 24,
		activeTab: TabAlerts,
		data: &DashboardData{
			Alerts: make([]monitor.Alert, 10),
		},
	}
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	updated := newModel.(Model)
	if updated.selectedRow != 1 {
		t.Errorf("j on alerts tab: selectedRow = %d, want 1", updated.selectedRow)
	}
}

func TestAlertsTabSupportsFilter(t *testing.T) {
	m := Model{
		width: 80, height: 24,
		activeTab:   TabAlerts,
		filterInput: textinput.New(),
		data: &DashboardData{
			Alerts: make([]monitor.Alert, 5),
		},
	}
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	updated := newModel.(Model)
	if !updated.filterMode {
		t.Error("/ on Alerts tab should activate filter mode")
	}
}

func TestAlertsTabSupportsDetailView(t *testing.T) {
	m := Model{
		width: 80, height: 24,
		activeTab: TabAlerts,
		data: &DashboardData{
			Alerts: []monitor.Alert{{ResourceName: "api"}},
		},
	}
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated := newModel.(Model)
	if !updated.showDetail {
		t.Error("Enter on Alerts tab should open detail view")
	}
}

func TestEventGenerationOnDataMsg(t *testing.T) {
	m := NewModel(api.NewClient("t"), "acc", 30*time.Second)
	m.width = 80
	m.height = 24

	data1 := &DashboardData{
		Workers:    []api.Worker{{Name: "w1"}},
		Containers: []api.Container{{Name: "c1"}},
		Alerts:     []monitor.Alert{{ResourceType: "worker", ResourceName: "api", Metric: "error_rate", Severity: "warning"}},
	}
	newModel, _ := m.Update(dataMsg{data: data1})
	updated := newModel.(Model)
	if len(updated.events) < 1 {
		t.Errorf("expected events after first data, got %d", len(updated.events))
	}
}

func TestViewContainsAlertsTab(t *testing.T) {
	m := NewModel(api.NewClient("t"), "acc", 30*time.Second)
	m.width = 80
	m.height = 24
	view := m.View()
	if !strings.Contains(view, "Alerts") {
		t.Error("view should contain 'Alerts' tab")
	}
}

func TestCurrentItemCountAlerts(t *testing.T) {
	m := Model{
		activeTab: TabAlerts,
		data: &DashboardData{
			Alerts: make([]monitor.Alert, 7),
		},
	}
	if got := m.currentItemCount(); got != 7 {
		t.Errorf("currentItemCount for alerts = %d, want 7", got)
	}
}

func TestOverviewShowsAlertsSummary(t *testing.T) {
	m := Model{
		width: 100, height: 24,
		activeTab: TabOverview,
		data: &DashboardData{
			HealthScore:  85,
			HealthStatus: "good",
			Workers:      []api.Worker{{Name: "w1"}},
			Containers:   []api.Container{{Name: "c1"}},
			Alerts: []monitor.Alert{
				{Severity: "warning"},
				{Severity: "critical"},
			},
		},
	}
	result := m.renderOverview()
	if !strings.Contains(result, "Alerts") {
		t.Errorf("overview should contain alerts card, got: %s", result)
	}
}

func TestOverviewAlertsCardNoAlerts(t *testing.T) {
	m := Model{
		width: 100, height: 24,
		activeTab: TabOverview,
		data: &DashboardData{
			HealthScore: 85,
			Workers:     []api.Worker{{Name: "w1"}},
			Containers:  []api.Container{{Name: "c1"}},
			Alerts:      []monitor.Alert{},
		},
	}
	result := m.renderOverview()
	if !strings.Contains(result, "Alerts") {
		t.Errorf("overview should always show alerts card, got: %s", result)
	}
}

func TestTabBadgeShowsAlertCount(t *testing.T) {
	m := Model{
		width: 100, height: 24,
		activeTab: TabOverview, // not on alerts tab
		data: &DashboardData{
			Alerts: []monitor.Alert{
				{Severity: "warning"},
				{Severity: "critical"},
				{Severity: "warning"},
			},
		},
	}
	tabs := m.renderTabs()
	if !strings.Contains(tabs, "3") {
		t.Errorf("tab badge should show alert count, got: %s", tabs)
	}
}

func TestErrorEventGeneration(t *testing.T) {
	m := NewModel(api.NewClient("t"), "acc", 30*time.Second)
	newModel, _ := m.Update(errMsg{err: fmt.Errorf("network timeout")})
	updated := newModel.(Model)
	found := false
	for _, e := range updated.events {
		if strings.Contains(e.Text, "network timeout") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected error event for fetch failure")
	}
}

func TestCurrentItemCountAlertsWithFilter(t *testing.T) {
	m := Model{
		activeTab:  TabAlerts,
		filterText: "api",
		data: &DashboardData{
			Alerts: []monitor.Alert{
				{ResourceName: "api-proxy", Metric: "error_rate"},
				{ResourceName: "db-svc", Metric: "memory"},
			},
		},
	}
	if got := m.currentItemCount(); got != 1 {
		t.Errorf("currentItemCount with filter = %d, want 1", got)
	}
}
