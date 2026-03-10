package monitor

import (
	"fmt"
	"time"

	"github.com/PeterHiroshi/cfmon/internal/api"
)

// Thresholds defines the alert thresholds for monitoring
type Thresholds struct {
	CPUPercent       float64
	MemoryPercent    float64
	ErrorRatePercent float64
}

// Alert represents a monitoring alert
type Alert struct {
	Severity     string  `json:"severity"`       // "warning" or "critical"
	ResourceType string  `json:"resource_type"`  // "worker" or "container"
	ResourceName string  `json:"resource_name"`
	Metric       string  `json:"metric"`         // "cpu", "memory", "error_rate"
	Value        float64 `json:"value"`
	Threshold    float64 `json:"threshold"`
	Message      string  `json:"message"`
}

// Summary provides an overview of the check results
type Summary struct {
	TotalWorkers    int `json:"total_workers"`
	TotalContainers int `json:"total_containers"`
	Warnings        int `json:"warnings"`
	Criticals       int `json:"criticals"`
}

// CheckResult contains the results of a monitoring check
type CheckResult struct {
	Timestamp  string          `json:"timestamp"`
	AccountID  string          `json:"account_id"`
	Summary    Summary         `json:"summary"`
	Alerts     []Alert         `json:"alerts"`
	Workers    []api.Worker    `json:"workers"`
	Containers []api.Container `json:"containers"`
}

// DefaultThresholds returns the default threshold values
func DefaultThresholds() Thresholds {
	return Thresholds{
		CPUPercent:       80.0,
		MemoryPercent:    85.0,
		ErrorRatePercent: 2.0,
	}
}

// MaxSeverity returns the maximum severity level from the check result
func (r *CheckResult) MaxSeverity() string {
	if r.Summary.Criticals > 0 {
		return "critical"
	}
	if r.Summary.Warnings > 0 {
		return "warning"
	}
	return "ok"
}

// EvaluateWorkers evaluates workers against thresholds and returns alerts
func EvaluateWorkers(workers []api.Worker, t Thresholds) []Alert {
	var alerts []Alert

	for _, worker := range workers {
		// Skip workers with no requests to avoid division by zero
		if worker.Requests == 0 {
			continue
		}

		errorRate := (float64(worker.Errors) / float64(worker.Requests)) * 100.0

		// Check for critical threshold (2x)
		if errorRate >= t.ErrorRatePercent*2.0 {
			alerts = append(alerts, Alert{
				Severity:     "critical",
				ResourceType: "worker",
				ResourceName: worker.Name,
				Metric:       "error_rate",
				Value:        errorRate,
				Threshold:    t.ErrorRatePercent * 2.0,
				Message:      fmt.Sprintf("Worker %s has critical error rate: %.2f%% (threshold: %.2f%%)", worker.Name, errorRate, t.ErrorRatePercent*2.0),
			})
		} else if errorRate >= t.ErrorRatePercent {
			// Check for warning threshold
			alerts = append(alerts, Alert{
				Severity:     "warning",
				ResourceType: "worker",
				ResourceName: worker.Name,
				Metric:       "error_rate",
				Value:        errorRate,
				Threshold:    t.ErrorRatePercent,
				Message:      fmt.Sprintf("Worker %s has high error rate: %.2f%% (threshold: %.2f%%)", worker.Name, errorRate, t.ErrorRatePercent),
			})
		}
	}

	return alerts
}

// EvaluateContainers evaluates containers against thresholds and returns alerts
func EvaluateContainers(containers []api.Container, t Thresholds, cpuLimitMS int, memoryLimitMB int) []Alert {
	var alerts []Alert

	for _, container := range containers {
		// Evaluate CPU if limit is set
		if cpuLimitMS > 0 {
			cpuPercent := (float64(container.CPUMS) / float64(cpuLimitMS)) * 100.0

			// Check for critical threshold (1.25x)
			if cpuPercent >= t.CPUPercent*1.25 {
				alerts = append(alerts, Alert{
					Severity:     "critical",
					ResourceType: "container",
					ResourceName: container.Name,
					Metric:       "cpu",
					Value:        cpuPercent,
					Threshold:    t.CPUPercent * 1.25,
					Message:      fmt.Sprintf("Container %s has critical CPU usage: %.2f%% (threshold: %.2f%%)", container.Name, cpuPercent, t.CPUPercent*1.25),
				})
			} else if cpuPercent >= t.CPUPercent {
				// Check for warning threshold
				alerts = append(alerts, Alert{
					Severity:     "warning",
					ResourceType: "container",
					ResourceName: container.Name,
					Metric:       "cpu",
					Value:        cpuPercent,
					Threshold:    t.CPUPercent,
					Message:      fmt.Sprintf("Container %s has high CPU usage: %.2f%% (threshold: %.2f%%)", container.Name, cpuPercent, t.CPUPercent),
				})
			}
		}

		// Evaluate memory if limit is set
		if memoryLimitMB > 0 {
			memoryPercent := (float64(container.MemoryMB) / float64(memoryLimitMB)) * 100.0

			// Check for critical threshold (1.15x)
			if memoryPercent >= t.MemoryPercent*1.15 {
				alerts = append(alerts, Alert{
					Severity:     "critical",
					ResourceType: "container",
					ResourceName: container.Name,
					Metric:       "memory",
					Value:        memoryPercent,
					Threshold:    t.MemoryPercent * 1.15,
					Message:      fmt.Sprintf("Container %s has critical memory usage: %.2f%% (threshold: %.2f%%)", container.Name, memoryPercent, t.MemoryPercent*1.15),
				})
			} else if memoryPercent >= t.MemoryPercent {
				// Check for warning threshold
				alerts = append(alerts, Alert{
					Severity:     "warning",
					ResourceType: "container",
					ResourceName: container.Name,
					Metric:       "memory",
					Value:        memoryPercent,
					Threshold:    t.MemoryPercent,
					Message:      fmt.Sprintf("Container %s has high memory usage: %.2f%% (threshold: %.2f%%)", container.Name, memoryPercent, t.MemoryPercent),
				})
			}
		}
	}

	return alerts
}

// RunCheck performs a complete monitoring check
func RunCheck(client *api.Client, accountID string, t Thresholds, cpuLimitMS int, memoryLimitMB int) (*CheckResult, error) {
	// Fetch workers
	workers, err := client.ListWorkers(accountID)
	if err != nil {
		return nil, fmt.Errorf("listing workers: %w", err)
	}

	// Fetch containers
	containers, err := client.ListContainers(accountID)
	if err != nil {
		return nil, fmt.Errorf("listing containers: %w", err)
	}

	// Evaluate workers and containers
	workerAlerts := EvaluateWorkers(workers, t)
	containerAlerts := EvaluateContainers(containers, t, cpuLimitMS, memoryLimitMB)

	// Combine alerts
	allAlerts := append(workerAlerts, containerAlerts...)

	// Calculate summary
	summary := Summary{
		TotalWorkers:    len(workers),
		TotalContainers: len(containers),
	}

	for _, alert := range allAlerts {
		if alert.Severity == "critical" {
			summary.Criticals++
		} else if alert.Severity == "warning" {
			summary.Warnings++
		}
	}

	// Create result
	result := &CheckResult{
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
		AccountID:  accountID,
		Summary:    summary,
		Alerts:     allAlerts,
		Workers:    workers,
		Containers: containers,
	}

	return result, nil
}
