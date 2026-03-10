package dashboard

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/PeterHiroshi/cfmon/internal/api"
	"github.com/PeterHiroshi/cfmon/internal/health"
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

		return dataMsg{data: data}
	}
}

func tickCmd(interval time.Duration) tea.Cmd {
	return tea.Tick(interval, func(t time.Time) tea.Msg {
		return tickMsg{time: t}
	})
}
