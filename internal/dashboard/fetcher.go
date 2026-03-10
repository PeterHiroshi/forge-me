package dashboard

import (
	"sort"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/PeterHiroshi/cfmon/internal/api"
	"github.com/PeterHiroshi/cfmon/internal/health"
	"github.com/PeterHiroshi/cfmon/internal/monitor"
)

const (
	DefaultRefreshInterval = 30 * time.Second
	MinRefreshInterval     = 5 * time.Second
)

type dataMsg struct {
	data *DashboardData
}

type errMsg struct {
	err error
}

type tickMsg struct {
	time time.Time
}

func sortAlerts(alerts []monitor.Alert) []monitor.Alert {
	sorted := make([]monitor.Alert, len(alerts))
	copy(sorted, alerts)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Severity != sorted[j].Severity {
			if sorted[i].Severity == "critical" {
				return true
			}
			if sorted[j].Severity == "critical" {
				return false
			}
		}
		return sorted[i].ResourceName < sorted[j].ResourceName
	})
	return sorted
}

func fetchData(client *api.Client, accountID string) tea.Cmd {
	return func() tea.Msg {
		data := &DashboardData{}

		workers, err := client.ListWorkers(accountID)
		if err != nil {
			return errMsg{err: err}
		}
		data.Workers = workers

		containers, err := client.ListContainers(accountID)
		if err != nil {
			return errMsg{err: err}
		}
		data.Containers = containers

		score, err := health.CalculateScore(client, accountID)
		if err != nil {
			return errMsg{err: err}
		}
		data.HealthScore = score.Total
		data.HealthStatus = score.Status

		var totalRequests, totalErrors int
		for _, w := range workers {
			totalRequests += w.Requests
			totalErrors += w.Errors
		}
		if totalRequests > 0 {
			data.ErrorRate = float64(totalErrors) / float64(totalRequests) * 100
		}

		th := monitor.DefaultThresholds()
		workerAlerts := monitor.EvaluateWorkers(workers, th)
		containerAlerts := monitor.EvaluateContainers(containers, th, 0, 0)
		allAlerts := append(workerAlerts, containerAlerts...)
		data.Alerts = sortAlerts(allAlerts)

		return dataMsg{data: data}
	}
}

func tickCmd(interval time.Duration) tea.Cmd {
	return tea.Tick(interval, func(t time.Time) tea.Msg {
		return tickMsg{time: t}
	})
}
