package dashboard

import (
	"fmt"
	"testing"
	"time"

	"github.com/PeterHiroshi/cfmon/internal/api"
	"github.com/PeterHiroshi/cfmon/internal/monitor"
)

func TestDataMsg(t *testing.T) {
	msg := dataMsg{
		data: &DashboardData{
			HealthScore: 85,
			Workers:     nil,
		},
	}
	if msg.data.HealthScore != 85 {
		t.Errorf("HealthScore = %d, want 85", msg.data.HealthScore)
	}
}

func TestErrMsg(t *testing.T) {
	msg := errMsg{err: fmt.Errorf("test error")}
	if msg.err.Error() != "test error" {
		t.Errorf("err = %q, want %q", msg.err.Error(), "test error")
	}
}

func TestTickMsg(t *testing.T) {
	msg := tickMsg{time: time.Now()}
	if msg.time.IsZero() {
		t.Error("expected non-zero time in tickMsg")
	}
}

func TestDefaultRefreshInterval(t *testing.T) {
	if DefaultRefreshInterval != 30*time.Second {
		t.Errorf("DefaultRefreshInterval = %v, want 30s", DefaultRefreshInterval)
	}
}

func TestMinRefreshInterval(t *testing.T) {
	if MinRefreshInterval != 5*time.Second {
		t.Errorf("MinRefreshInterval = %v, want 5s", MinRefreshInterval)
	}
}

func TestFetchDataIncludesAlerts(t *testing.T) {
	data := &DashboardData{
		Workers: []api.Worker{
			{Name: "high-err", Requests: 100, Errors: 10}, // 10% error rate > 2% threshold
		},
		Containers: []api.Container{
			{Name: "c1", CPUMS: 100, MemoryMB: 64},
		},
	}
	th := monitor.DefaultThresholds()
	alerts := monitor.EvaluateWorkers(data.Workers, th)
	data.Alerts = alerts

	if len(data.Alerts) == 0 {
		t.Error("expected alerts for high error rate worker")
	}
	if data.Alerts[0].ResourceName != "high-err" {
		t.Errorf("alert resource = %q, want high-err", data.Alerts[0].ResourceName)
	}
}

func TestAlertsSortedCriticalsFirst(t *testing.T) {
	alerts := []monitor.Alert{
		{Severity: "warning", ResourceName: "b"},
		{Severity: "critical", ResourceName: "a"},
		{Severity: "warning", ResourceName: "a"},
		{Severity: "critical", ResourceName: "b"},
	}
	sorted := sortAlerts(alerts)
	if sorted[0].Severity != "critical" {
		t.Errorf("first alert should be critical, got %q", sorted[0].Severity)
	}
	if sorted[1].Severity != "critical" {
		t.Errorf("second alert should be critical, got %q", sorted[1].Severity)
	}
	if sorted[0].ResourceName != "a" {
		t.Errorf("first critical should be 'a', got %q", sorted[0].ResourceName)
	}
}
