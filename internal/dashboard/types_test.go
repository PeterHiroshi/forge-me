package dashboard

import (
	"testing"
	"time"
)

func TestTabIDConstants(t *testing.T) {
	if TabOverview != 0 {
		t.Errorf("TabOverview = %d, want 0", TabOverview)
	}
	if TabWorkers != 1 {
		t.Errorf("TabWorkers = %d, want 1", TabWorkers)
	}
	if TabContainers != 2 {
		t.Errorf("TabContainers = %d, want 2", TabContainers)
	}
	if TabAlerts != 3 {
		t.Errorf("TabAlerts = %d, want 3", TabAlerts)
	}
}

func TestTabIDCount(t *testing.T) {
	if tabCount != 4 {
		t.Errorf("tabCount = %d, want 4", tabCount)
	}
}

func TestTabNames(t *testing.T) {
	tests := []struct {
		tab  TabID
		want string
	}{
		{TabOverview, "Overview"},
		{TabWorkers, "Workers"},
		{TabContainers, "Containers"},
		{TabAlerts, "Alerts"},
	}
	for _, tt := range tests {
		if got := tt.tab.String(); got != tt.want {
			t.Errorf("TabID(%d).String() = %q, want %q", tt.tab, got, tt.want)
		}
	}
}

func TestDashboardDataDefaults(t *testing.T) {
	d := &DashboardData{}
	if d.Workers != nil {
		t.Error("expected nil Workers")
	}
	if d.Containers != nil {
		t.Error("expected nil Containers")
	}
	if d.HealthScore != 0 {
		t.Error("expected 0 HealthScore")
	}
}

func TestTabAlertsConstant(t *testing.T) {
	if TabAlerts != 3 {
		t.Errorf("TabAlerts = %d, want 3", TabAlerts)
	}
}

func TestTabCountIsFour(t *testing.T) {
	if tabCount != 4 {
		t.Errorf("tabCount = %d, want 4", tabCount)
	}
}

func TestTabAlertsName(t *testing.T) {
	if got := TabAlerts.String(); got != "Alerts" {
		t.Errorf("TabAlerts.String() = %q, want %q", got, "Alerts")
	}
}

func TestDashboardEventStruct(t *testing.T) {
	now := time.Now()
	e := DashboardEvent{Time: now, Text: "test event", Severity: "info"}
	if e.Time != now {
		t.Error("expected matching time")
	}
	if e.Text != "test event" {
		t.Errorf("Text = %q, want %q", e.Text, "test event")
	}
	if e.Severity != "info" {
		t.Errorf("Severity = %q, want %q", e.Severity, "info")
	}
}

func TestDashboardDataAlerts(t *testing.T) {
	d := &DashboardData{}
	if d.Alerts != nil {
		t.Error("expected nil Alerts on zero value")
	}
}
