package dashboard

import "testing"

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
}

func TestTabIDCount(t *testing.T) {
	if tabCount != 3 {
		t.Errorf("tabCount = %d, want 3", tabCount)
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
