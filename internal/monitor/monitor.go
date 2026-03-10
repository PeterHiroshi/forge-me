package monitor

import (
	"fmt"
	"time"

	"github.com/PeterHiroshi/cfmon/internal/api"
)

// Thresholds defines configurable alert thresholds
type Thresholds struct {
	CPUPercent       float64
	MemoryPercent    float64
	ErrorRatePercent float64
}

// Alert represents a threshold violation for a resource
type Alert struct {
	Severity     string  `json:"severity"`
	ResourceType string  `json:"resource_type"`
	ResourceName string  `json:"resource_name"`
	Metric       string  `json:"metric"`
	Value        float64 `json:"value"`
	Threshold    float64 `json:"threshold"`
	Message      string  `json:"message"`
}

// Summary contains aggregate counts for a check result
type Summary struct {
	TotalWorkers    int    `json:"total_workers"`
	TotalContainers int    `json:"total_containers"`
	TotalAlerts     int    `json:"total_alerts"`
	MaxSeverity     string `json:"max_severity"`
}

// CheckResult contains the full result of a one-shot check
type CheckResult struct {
	Timestamp  string          `json:"timestamp"`
	AccountID  string          `json:"account_id"`
	Summary    Summary         `json:"summary"`
	Alerts     []Alert         `json:"alerts"`
	Workers    []api.Worker    `json:"workers"`
	Containers []api.Container `json:"containers"`
}

const (
	maxCPUMS    = 1000  // 1000ms CPU time = 100%
	maxMemoryMB = 1024  // 1024MB = 100%
)

// DefaultThresholds returns sensible default thresholds
func DefaultThresholds() Thresholds {
	return Thresholds{
		CPUPercent:       80,
		MemoryPercent:    85,
		ErrorRatePercent: 2.0,
	}
}

// EvaluateWorkers evaluates workers against thresholds and returns alerts
func EvaluateWorkers(workers []api.Worker, th Thresholds) []Alert {
	var alerts []Alert
	for _, w := range workers {
		if w.Requests == 0 {
			continue
		}
		errorRate := float64(w.Errors) / float64(w.Requests) * 100
		if errorRate > th.ErrorRatePercent {
			alerts = append(alerts, Alert{
				Severity:     "critical",
				ResourceType: "worker",
				ResourceName: w.Name,
				Metric:       "error_rate",
				Value:        errorRate,
				Threshold:    th.ErrorRatePercent,
				Message:      fmt.Sprintf("worker %q error rate %.1f%% exceeds threshold %.1f%%", w.Name, errorRate, th.ErrorRatePercent),
			})
		}
	}
	return alerts
}

// EvaluateContainers evaluates containers against thresholds and returns alerts
func EvaluateContainers(containers []api.Container, th Thresholds) []Alert {
	var alerts []Alert
	for _, c := range containers {
		// Check for non-running status
		if c.Status != "" && c.Status != "running" && c.Status != "active" {
			alerts = append(alerts, Alert{
				Severity:     "critical",
				ResourceType: "container",
				ResourceName: c.Name,
				Metric:       "status",
				Value:        0,
				Threshold:    0,
				Message:      fmt.Sprintf("container %q is %s", c.Name, c.Status),
			})
			continue
		}

		// Check CPU
		cpuPercent := float64(c.CPUMS) / float64(maxCPUMS) * 100
		if cpuPercent > th.CPUPercent {
			alerts = append(alerts, Alert{
				Severity:     "warning",
				ResourceType: "container",
				ResourceName: c.Name,
				Metric:       "cpu",
				Value:        cpuPercent,
				Threshold:    th.CPUPercent,
				Message:      fmt.Sprintf("container %q CPU %.1f%% exceeds threshold %.1f%%", c.Name, cpuPercent, th.CPUPercent),
			})
		}

		// Check Memory
		memPercent := float64(c.MemoryMB) / float64(maxMemoryMB) * 100
		if memPercent > th.MemoryPercent {
			alerts = append(alerts, Alert{
				Severity:     "warning",
				ResourceType: "container",
				ResourceName: c.Name,
				Metric:       "memory",
				Value:        memPercent,
				Threshold:    th.MemoryPercent,
				Message:      fmt.Sprintf("container %q memory %.1f%% exceeds threshold %.1f%%", c.Name, memPercent, th.MemoryPercent),
			})
		}
	}
	return alerts
}

// MaxSeverity returns the highest severity from a list of alerts
func MaxSeverity(alerts []Alert) string {
	if len(alerts) == 0 {
		return "ok"
	}
	for _, a := range alerts {
		if a.Severity == "critical" {
			return "critical"
		}
	}
	return "warning"
}

// RunCheck performs a one-shot health check against an account
func RunCheck(client *api.Client, accountID string, th Thresholds) (*CheckResult, error) {
	workers, err := client.ListWorkers(accountID)
	if err != nil {
		return nil, fmt.Errorf("listing workers: %w", err)
	}

	containers, err := client.ListContainers(accountID)
	if err != nil {
		return nil, fmt.Errorf("listing containers: %w", err)
	}

	var alerts []Alert
	alerts = append(alerts, EvaluateWorkers(workers, th)...)
	alerts = append(alerts, EvaluateContainers(containers, th)...)

	result := &CheckResult{
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
		AccountID:  accountID,
		Workers:    workers,
		Containers: containers,
		Alerts:     alerts,
		Summary: Summary{
			TotalWorkers:    len(workers),
			TotalContainers: len(containers),
			TotalAlerts:     len(alerts),
			MaxSeverity:     MaxSeverity(alerts),
		},
	}

	return result, nil
}
