package dashboard

import "github.com/PeterHiroshi/cfmon/internal/api"

// TabID identifies a dashboard tab.
type TabID int

const (
	TabOverview   TabID = 0
	TabWorkers    TabID = 1
	TabContainers TabID = 2
	tabCount            = 3
)

var tabNames = [tabCount]string{"Overview", "Workers", "Containers"}

// String returns the display name of the tab.
func (t TabID) String() string {
	if t >= 0 && int(t) < len(tabNames) {
		return tabNames[t]
	}
	return "Unknown"
}

// DashboardData holds all fetched data for the dashboard.
type DashboardData struct {
	Workers      []api.Worker
	Containers   []api.Container
	HealthScore  int
	HealthStatus string
	ErrorRate    float64
}
