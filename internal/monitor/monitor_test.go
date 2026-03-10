package monitor

import (
	"testing"

	"github.com/PeterHiroshi/cfmon/internal/api"
)

func TestDefaultThresholds(t *testing.T) {
	thresholds := DefaultThresholds()

	if thresholds.CPUPercent != 80.0 {
		t.Errorf("Expected CPUPercent 80.0, got %.2f", thresholds.CPUPercent)
	}
	if thresholds.MemoryPercent != 85.0 {
		t.Errorf("Expected MemoryPercent 85.0, got %.2f", thresholds.MemoryPercent)
	}
	if thresholds.ErrorRatePercent != 2.0 {
		t.Errorf("Expected ErrorRatePercent 2.0, got %.2f", thresholds.ErrorRatePercent)
	}
}

func TestCheckResult_MaxSeverity(t *testing.T) {
	tests := []struct {
		name      string
		criticals int
		warnings  int
		want      string
	}{
		{
			name:      "ok when no alerts",
			criticals: 0,
			warnings:  0,
			want:      "ok",
		},
		{
			name:      "warning when only warnings",
			criticals: 0,
			warnings:  2,
			want:      "warning",
		},
		{
			name:      "critical when only criticals",
			criticals: 1,
			warnings:  0,
			want:      "critical",
		},
		{
			name:      "critical when both exist",
			criticals: 1,
			warnings:  2,
			want:      "critical",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &CheckResult{
				Summary: Summary{
					Criticals: tt.criticals,
					Warnings:  tt.warnings,
				},
			}

			got := result.MaxSeverity()
			if got != tt.want {
				t.Errorf("MaxSeverity() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEvaluateWorkers(t *testing.T) {
	thresholds := DefaultThresholds()

	tests := []struct {
		name          string
		workers       []api.Worker
		thresholds    Thresholds
		wantAlerts    int
		wantWarnings  int
		wantCriticals int
	}{
		{
			name:          "no workers",
			workers:       []api.Worker{},
			thresholds:    thresholds,
			wantAlerts:    0,
			wantWarnings:  0,
			wantCriticals: 0,
		},
		{
			name: "healthy workers",
			workers: []api.Worker{
				{Name: "worker1", Requests: 1000, Errors: 10}, // 1% error rate
				{Name: "worker2", Requests: 500, Errors: 5},   // 1% error rate
			},
			thresholds:    thresholds,
			wantAlerts:    0,
			wantWarnings:  0,
			wantCriticals: 0,
		},
		{
			name: "worker at warning threshold",
			workers: []api.Worker{
				{Name: "worker1", Requests: 1000, Errors: 20}, // 2% error rate
			},
			thresholds:    thresholds,
			wantAlerts:    1,
			wantWarnings:  1,
			wantCriticals: 0,
		},
		{
			name: "worker at critical threshold",
			workers: []api.Worker{
				{Name: "worker1", Requests: 1000, Errors: 40}, // 4% error rate (2x threshold)
			},
			thresholds:    thresholds,
			wantAlerts:    1,
			wantWarnings:  0,
			wantCriticals: 1,
		},
		{
			name: "worker with zero requests skipped",
			workers: []api.Worker{
				{Name: "worker1", Requests: 0, Errors: 10},
			},
			thresholds:    thresholds,
			wantAlerts:    0,
			wantWarnings:  0,
			wantCriticals: 0,
		},
		{
			name: "custom thresholds",
			workers: []api.Worker{
				{Name: "worker1", Requests: 1000, Errors: 50}, // 5% error rate
			},
			thresholds: Thresholds{
				ErrorRatePercent: 5.0,
			},
			wantAlerts:    1,
			wantWarnings:  1,
			wantCriticals: 0,
		},
		{
			name: "multiple workers with different severities",
			workers: []api.Worker{
				{Name: "worker1", Requests: 1000, Errors: 10}, // 1% - ok
				{Name: "worker2", Requests: 1000, Errors: 20}, // 2% - warning
				{Name: "worker3", Requests: 1000, Errors: 40}, // 4% - critical
			},
			thresholds:    thresholds,
			wantAlerts:    2,
			wantWarnings:  1,
			wantCriticals: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alerts := EvaluateWorkers(tt.workers, tt.thresholds)

			if len(alerts) != tt.wantAlerts {
				t.Errorf("Expected %d alerts, got %d", tt.wantAlerts, len(alerts))
			}

			warnings := 0
			criticals := 0
			for _, alert := range alerts {
				if alert.Severity == "warning" {
					warnings++
				} else if alert.Severity == "critical" {
					criticals++
				}

				// Validate alert fields
				if alert.ResourceType != "worker" {
					t.Errorf("Expected ResourceType 'worker', got '%s'", alert.ResourceType)
				}
				if alert.Metric != "error_rate" {
					t.Errorf("Expected Metric 'error_rate', got '%s'", alert.Metric)
				}
				if alert.Message == "" {
					t.Error("Alert message should not be empty")
				}
			}

			if warnings != tt.wantWarnings {
				t.Errorf("Expected %d warnings, got %d", tt.wantWarnings, warnings)
			}
			if criticals != tt.wantCriticals {
				t.Errorf("Expected %d criticals, got %d", tt.wantCriticals, criticals)
			}
		})
	}
}

func TestEvaluateContainers(t *testing.T) {
	thresholds := DefaultThresholds()

	tests := []struct {
		name           string
		containers     []api.Container
		thresholds     Thresholds
		cpuLimitMS     int
		memoryLimitMB  int
		wantAlerts     int
		wantWarnings   int
		wantCriticals  int
		wantCPUAlerts  int
		wantMemAlerts  int
	}{
		{
			name:          "no containers",
			containers:    []api.Container{},
			thresholds:    thresholds,
			cpuLimitMS:    1000,
			memoryLimitMB: 128,
			wantAlerts:    0,
			wantWarnings:  0,
			wantCriticals: 0,
		},
		{
			name: "healthy containers",
			containers: []api.Container{
				{Name: "container1", CPUMS: 500, MemoryMB: 64},  // 50% CPU, 50% memory
				{Name: "container2", CPUMS: 600, MemoryMB: 80},  // 60% CPU, 62.5% memory
			},
			thresholds:    thresholds,
			cpuLimitMS:    1000,
			memoryLimitMB: 128,
			wantAlerts:    0,
			wantWarnings:  0,
			wantCriticals: 0,
		},
		{
			name: "CPU warning threshold",
			containers: []api.Container{
				{Name: "container1", CPUMS: 800, MemoryMB: 64}, // 80% CPU (at threshold)
			},
			thresholds:     thresholds,
			cpuLimitMS:     1000,
			memoryLimitMB:  128,
			wantAlerts:     1,
			wantWarnings:   1,
			wantCriticals:  0,
			wantCPUAlerts:  1,
		},
		{
			name: "CPU critical threshold",
			containers: []api.Container{
				{Name: "container1", CPUMS: 1000, MemoryMB: 64}, // 100% CPU (1.25x threshold)
			},
			thresholds:     thresholds,
			cpuLimitMS:     1000,
			memoryLimitMB:  128,
			wantAlerts:     1,
			wantWarnings:   0,
			wantCriticals:  1,
			wantCPUAlerts:  1,
		},
		{
			name: "memory warning threshold",
			containers: []api.Container{
				{Name: "container1", CPUMS: 500, MemoryMB: 109}, // 85.15% memory (at threshold)
			},
			thresholds:     thresholds,
			cpuLimitMS:     1000,
			memoryLimitMB:  128,
			wantAlerts:     1,
			wantWarnings:   1,
			wantCriticals:  0,
			wantMemAlerts:  1,
		},
		{
			name: "memory critical threshold",
			containers: []api.Container{
				{Name: "container1", CPUMS: 500, MemoryMB: 126}, // 98.44% memory (above 1.15x threshold)
			},
			thresholds:     thresholds,
			cpuLimitMS:     1000,
			memoryLimitMB:  128,
			wantAlerts:     1,
			wantWarnings:   0,
			wantCriticals:  1,
			wantMemAlerts:  1,
		},
		{
			name: "skip CPU when limit is 0",
			containers: []api.Container{
				{Name: "container1", CPUMS: 9999, MemoryMB: 64},
			},
			thresholds:     thresholds,
			cpuLimitMS:     0,
			memoryLimitMB:  128,
			wantAlerts:     0,
			wantWarnings:   0,
			wantCriticals:  0,
		},
		{
			name: "skip memory when limit is 0",
			containers: []api.Container{
				{Name: "container1", CPUMS: 500, MemoryMB: 9999},
			},
			thresholds:     thresholds,
			cpuLimitMS:     1000,
			memoryLimitMB:  0,
			wantAlerts:     0,
			wantWarnings:   0,
			wantCriticals:  0,
		},
		{
			name: "multiple alerts from one container",
			containers: []api.Container{
				{Name: "container1", CPUMS: 1000, MemoryMB: 126}, // Both CPU and memory critical
			},
			thresholds:     thresholds,
			cpuLimitMS:     1000,
			memoryLimitMB:  128,
			wantAlerts:     2,
			wantWarnings:   0,
			wantCriticals:  2,
			wantCPUAlerts:  1,
			wantMemAlerts:  1,
		},
		{
			name: "mixed severities across containers",
			containers: []api.Container{
				{Name: "container1", CPUMS: 800, MemoryMB: 64},  // CPU warning
				{Name: "container2", CPUMS: 500, MemoryMB: 109}, // Memory warning
				{Name: "container3", CPUMS: 1000, MemoryMB: 126}, // Both critical
			},
			thresholds:     thresholds,
			cpuLimitMS:     1000,
			memoryLimitMB:  128,
			wantAlerts:     4,
			wantWarnings:   2,
			wantCriticals:  2,
			wantCPUAlerts:  2,
			wantMemAlerts:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alerts := EvaluateContainers(tt.containers, tt.thresholds, tt.cpuLimitMS, tt.memoryLimitMB)

			if len(alerts) != tt.wantAlerts {
				t.Errorf("Expected %d alerts, got %d", tt.wantAlerts, len(alerts))
			}

			warnings := 0
			criticals := 0
			cpuAlerts := 0
			memAlerts := 0

			for _, alert := range alerts {
				if alert.Severity == "warning" {
					warnings++
				} else if alert.Severity == "critical" {
					criticals++
				}

				if alert.Metric == "cpu" {
					cpuAlerts++
				} else if alert.Metric == "memory" {
					memAlerts++
				}

				// Validate alert fields
				if alert.ResourceType != "container" {
					t.Errorf("Expected ResourceType 'container', got '%s'", alert.ResourceType)
				}
				if alert.Metric != "cpu" && alert.Metric != "memory" {
					t.Errorf("Expected Metric 'cpu' or 'memory', got '%s'", alert.Metric)
				}
				if alert.Message == "" {
					t.Error("Alert message should not be empty")
				}
			}

			if warnings != tt.wantWarnings {
				t.Errorf("Expected %d warnings, got %d", tt.wantWarnings, warnings)
			}
			if criticals != tt.wantCriticals {
				t.Errorf("Expected %d criticals, got %d", tt.wantCriticals, criticals)
			}
			if tt.wantCPUAlerts > 0 && cpuAlerts != tt.wantCPUAlerts {
				t.Errorf("Expected %d CPU alerts, got %d", tt.wantCPUAlerts, cpuAlerts)
			}
			if tt.wantMemAlerts > 0 && memAlerts != tt.wantMemAlerts {
				t.Errorf("Expected %d memory alerts, got %d", tt.wantMemAlerts, memAlerts)
			}
		})
	}
}
