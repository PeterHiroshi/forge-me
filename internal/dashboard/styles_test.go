package dashboard

import "testing"

func TestStatusColor(t *testing.T) {
	tests := []struct {
		status string
		want   string
	}{
		{"active", "46"},
		{"running", "46"},
		{"stopped", "196"},
		{"error", "196"},
		{"deploying", "226"},
		{"unknown", "226"},
		{"", "226"},
	}
	for _, tt := range tests {
		got := statusColor(tt.status)
		if got != tt.want {
			t.Errorf("statusColor(%q) = %q, want %q", tt.status, got, tt.want)
		}
	}
}

func TestErrorRateColor(t *testing.T) {
	tests := []struct {
		rate float64
		want string
	}{
		{0.0, "46"},
		{0.5, "46"},
		{0.99, "46"},
		{1.0, "226"},
		{3.0, "226"},
		{5.0, "226"},
		{5.01, "196"},
		{10.0, "196"},
	}
	for _, tt := range tests {
		got := errorRateColor(tt.rate)
		if got != tt.want {
			t.Errorf("errorRateColor(%f) = %q, want %q", tt.rate, got, tt.want)
		}
	}
}

func TestSeverityColor(t *testing.T) {
	tests := []struct {
		severity string
		want     string
	}{
		{"warning", "226"},
		{"critical", "196"},
		{"ok", "46"},
		{"unknown", "241"},
	}
	for _, tt := range tests {
		if got := severityColor(tt.severity); got != tt.want {
			t.Errorf("severityColor(%q) = %q, want %q", tt.severity, got, tt.want)
		}
	}
}
