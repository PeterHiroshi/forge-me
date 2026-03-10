package dashboard

import (
	"fmt"
	"testing"
	"time"
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
