package health

import (
	"fmt"
	"time"

	"github.com/PeterHiroshi/cfmon/internal/api"
)

// Score represents a health score with breakdown
type Score struct {
	Total              int     `json:"total"`
	APIConnectivity    int     `json:"api_connectivity"`
	WorkerHealth       int     `json:"worker_health"`
	ContainerHealth    int     `json:"container_health"`
	APIConnectivityMax int     `json:"api_connectivity_max"`
	WorkerHealthMax    int     `json:"worker_health_max"`
	ContainerHealthMax int     `json:"container_health_max"`
	Status             string  `json:"status"`
	Message            string  `json:"message"`
	Timestamp          string  `json:"timestamp"`
}

const (
	APIConnectivityMax = 30
	WorkerHealthMax    = 35
	ContainerHealthMax = 35
	MaxScore           = 100
)

// CalculateScore calculates the overall health score for an account
func CalculateScore(client *api.Client, accountID string) (*Score, error) {
	score := &Score{
		APIConnectivityMax: APIConnectivityMax,
		WorkerHealthMax:    WorkerHealthMax,
		ContainerHealthMax: ContainerHealthMax,
		Timestamp:          time.Now().UTC().Format(time.RFC3339),
	}

	// Test API connectivity (30 points)
	apiScore, err := checkAPIConnectivity(client)
	if err != nil {
		score.Message = fmt.Sprintf("API connectivity check failed: %v", err)
		score.Status = "critical"
		return score, nil
	}
	score.APIConnectivity = apiScore

	// Check worker health (35 points)
	workerScore, err := checkWorkerHealth(client, accountID)
	if err != nil {
		// If we can't check workers, give partial credit
		score.WorkerHealth = 0
		score.Message = fmt.Sprintf("Worker health check failed: %v", err)
	} else {
		score.WorkerHealth = workerScore
	}

	// Check container health (35 points)
	containerScore, err := checkContainerHealth(client, accountID)
	if err != nil {
		// If we can't check containers, give partial credit
		score.ContainerHealth = 0
		if score.Message == "" {
			score.Message = fmt.Sprintf("Container health check failed: %v", err)
		}
	} else {
		score.ContainerHealth = containerScore
	}

	// Calculate total
	score.Total = score.APIConnectivity + score.WorkerHealth + score.ContainerHealth

	// Determine status
	score.Status = determineStatus(score.Total)

	// Set message if not already set
	if score.Message == "" {
		score.Message = "All systems operational"
	}

	return score, nil
}

// checkAPIConnectivity tests if the API is reachable and responsive
func checkAPIConnectivity(client *api.Client) (int, error) {
	// Try to list accounts as a connectivity test
	_, err := client.ListAccounts()
	if err != nil {
		return 0, err
	}
	return APIConnectivityMax, nil
}

// checkWorkerHealth checks the health of workers in the account
func checkWorkerHealth(client *api.Client, accountID string) (int, error) {
	workers, err := client.ListWorkers(accountID)
	if err != nil {
		return 0, err
	}

	if len(workers) == 0 {
		// No workers, full credit (nothing to be unhealthy)
		return WorkerHealthMax, nil
	}

	// Calculate health based on error rates
	totalWorkers := len(workers)
	healthyWorkers := 0
	for _, w := range workers {
		// Consider a worker healthy if it has low error rate
		if w.Requests == 0 {
			// No requests, assume healthy
			healthyWorkers++
		} else {
			errorRate := float64(w.Errors) / float64(w.Requests) * 100
			if errorRate < 5 { // Less than 5% error rate
				healthyWorkers++
			}
		}
	}

	// Score proportional to healthy workers
	healthPercent := float64(healthyWorkers) / float64(totalWorkers)
	return int(float64(WorkerHealthMax) * healthPercent), nil
}

// checkContainerHealth checks the health of containers in the account
func checkContainerHealth(client *api.Client, accountID string) (int, error) {
	containers, err := client.ListContainers(accountID)
	if err != nil {
		return 0, err
	}

	if len(containers) == 0 {
		// No containers, full credit (nothing to be unhealthy)
		return ContainerHealthMax, nil
	}

	// Calculate health based on status
	totalContainers := len(containers)
	healthyContainers := 0
	for _, c := range containers {
		// Consider a container healthy if status is running or empty (assumed running)
		if c.Status == "" || c.Status == "running" || c.Status == "active" {
			healthyContainers++
		}
	}

	// Score proportional to healthy containers
	healthPercent := float64(healthyContainers) / float64(totalContainers)
	return int(float64(ContainerHealthMax) * healthPercent), nil
}

// determineStatus determines the overall status based on total score
func determineStatus(total int) string {
	switch {
	case total >= 90:
		return "excellent"
	case total >= 75:
		return "good"
	case total >= 50:
		return "fair"
	case total >= 25:
		return "poor"
	default:
		return "critical"
	}
}
