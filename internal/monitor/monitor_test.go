package monitor

import (
	"testing"

	"github.com/PeterHiroshi/cfmon/internal/api"
)

func TestDefaultThresholds(t *testing.T) {
	th := DefaultThresholds()
	if th.CPUPercent != 80 {
		t.Errorf("CPUPercent = %f, want 80", th.CPUPercent)
	}
	if th.MemoryPercent != 85 {
		t.Errorf("MemoryPercent = %f, want 85", th.MemoryPercent)
	}
	if th.ErrorRatePercent != 2.0 {
		t.Errorf("ErrorRatePercent = %f, want 2.0", th.ErrorRatePercent)
	}
}

func TestEvaluateWorkers_NoWorkers(t *testing.T) {
	th := DefaultThresholds()
	alerts := EvaluateWorkers(nil, th)
	if len(alerts) != 0 {
		t.Errorf("got %d alerts for nil workers, want 0", len(alerts))
	}

	alerts = EvaluateWorkers([]api.Worker{}, th)
	if len(alerts) != 0 {
		t.Errorf("got %d alerts for empty workers, want 0", len(alerts))
	}
}

func TestEvaluateWorkers_HighErrorRate(t *testing.T) {
	th := DefaultThresholds()
	workers := []api.Worker{
		{Name: "api-worker", Requests: 1000, Errors: 30}, // 3% error rate > 2% threshold
	}

	alerts := EvaluateWorkers(workers, th)
	if len(alerts) != 1 {
		t.Fatalf("got %d alerts, want 1", len(alerts))
	}

	a := alerts[0]
	if a.Severity != "critical" {
		t.Errorf("Severity = %q, want %q", a.Severity, "critical")
	}
	if a.ResourceType != "worker" {
		t.Errorf("ResourceType = %q, want %q", a.ResourceType, "worker")
	}
	if a.ResourceName != "api-worker" {
		t.Errorf("ResourceName = %q, want %q", a.ResourceName, "api-worker")
	}
	if a.Metric != "error_rate" {
		t.Errorf("Metric = %q, want %q", a.Metric, "error_rate")
	}
	if a.Value != 3.0 {
		t.Errorf("Value = %f, want 3.0", a.Value)
	}
	if a.Threshold != 2.0 {
		t.Errorf("Threshold = %f, want 2.0", a.Threshold)
	}
}

func TestEvaluateWorkers_HealthyWorker(t *testing.T) {
	th := DefaultThresholds()
	workers := []api.Worker{
		{Name: "healthy-worker", Requests: 1000, Errors: 5}, // 0.5% error rate
	}

	alerts := EvaluateWorkers(workers, th)
	if len(alerts) != 0 {
		t.Errorf("got %d alerts for healthy worker, want 0", len(alerts))
	}
}

func TestEvaluateWorkers_NoRequests(t *testing.T) {
	th := DefaultThresholds()
	workers := []api.Worker{
		{Name: "idle-worker", Requests: 0, Errors: 0},
	}

	alerts := EvaluateWorkers(workers, th)
	if len(alerts) != 0 {
		t.Errorf("got %d alerts for idle worker, want 0", len(alerts))
	}
}

func TestEvaluateContainers_NoContainers(t *testing.T) {
	th := DefaultThresholds()
	alerts := EvaluateContainers(nil, th)
	if len(alerts) != 0 {
		t.Errorf("got %d alerts for nil containers, want 0", len(alerts))
	}
}

func TestEvaluateContainers_HighCPU(t *testing.T) {
	th := DefaultThresholds()
	containers := []api.Container{
		{Name: "web-app", CPUMS: 900, MemoryMB: 256}, // 90% CPU > 80% threshold (assuming 1000ms max)
	}

	alerts := EvaluateContainers(containers, th)
	found := false
	for _, a := range alerts {
		if a.Metric == "cpu" && a.ResourceName == "web-app" {
			found = true
			if a.Severity != "warning" {
				t.Errorf("Severity = %q, want %q", a.Severity, "warning")
			}
			if a.ResourceType != "container" {
				t.Errorf("ResourceType = %q, want %q", a.ResourceType, "container")
			}
		}
	}
	if !found {
		t.Error("expected CPU alert for web-app container, got none")
	}
}

func TestEvaluateContainers_HighMemory(t *testing.T) {
	th := DefaultThresholds()
	containers := []api.Container{
		{Name: "mem-hog", CPUMS: 100, MemoryMB: 900}, // 90% memory > 85% threshold (assuming 1024MB max)
	}

	alerts := EvaluateContainers(containers, th)
	found := false
	for _, a := range alerts {
		if a.Metric == "memory" && a.ResourceName == "mem-hog" {
			found = true
			if a.Severity != "warning" {
				t.Errorf("Severity = %q, want %q", a.Severity, "warning")
			}
		}
	}
	if !found {
		t.Error("expected memory alert for mem-hog container, got none")
	}
}

func TestEvaluateContainers_Healthy(t *testing.T) {
	th := DefaultThresholds()
	containers := []api.Container{
		{Name: "small-app", CPUMS: 100, MemoryMB: 128},
	}

	alerts := EvaluateContainers(containers, th)
	if len(alerts) != 0 {
		t.Errorf("got %d alerts for healthy container, want 0", len(alerts))
	}
}

func TestEvaluateContainers_NotRunning(t *testing.T) {
	th := DefaultThresholds()
	containers := []api.Container{
		{Name: "stopped", Status: "stopped", CPUMS: 0, MemoryMB: 0},
	}

	alerts := EvaluateContainers(containers, th)
	found := false
	for _, a := range alerts {
		if a.ResourceName == "stopped" {
			found = true
			if a.Severity != "critical" {
				t.Errorf("Severity = %q, want %q", a.Severity, "critical")
			}
		}
	}
	if !found {
		t.Error("expected alert for stopped container")
	}
}

func TestMaxSeverity_NoAlerts(t *testing.T) {
	sev := MaxSeverity(nil)
	if sev != "ok" {
		t.Errorf("MaxSeverity(nil) = %q, want %q", sev, "ok")
	}
}

func TestMaxSeverity_WarningsOnly(t *testing.T) {
	alerts := []Alert{
		{Severity: "warning"},
		{Severity: "warning"},
	}
	sev := MaxSeverity(alerts)
	if sev != "warning" {
		t.Errorf("MaxSeverity = %q, want %q", sev, "warning")
	}
}

func TestMaxSeverity_CriticalPresent(t *testing.T) {
	alerts := []Alert{
		{Severity: "warning"},
		{Severity: "critical"},
	}
	sev := MaxSeverity(alerts)
	if sev != "critical" {
		t.Errorf("MaxSeverity = %q, want %q", sev, "critical")
	}
}

func TestCustomThresholds(t *testing.T) {
	th := Thresholds{
		CPUPercent:       70,
		MemoryPercent:    75,
		ErrorRatePercent: 1.0,
	}

	// 1.5% error rate should trigger with 1.0% threshold
	workers := []api.Worker{
		{Name: "sensitive-worker", Requests: 1000, Errors: 15},
	}

	alerts := EvaluateWorkers(workers, th)
	if len(alerts) != 1 {
		t.Fatalf("got %d alerts with custom threshold, want 1", len(alerts))
	}
	if alerts[0].Threshold != 1.0 {
		t.Errorf("Threshold = %f, want 1.0", alerts[0].Threshold)
	}
}
