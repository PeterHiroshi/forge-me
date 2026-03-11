package dashboard

import "github.com/charmbracelet/lipgloss"

var (
	activeTabStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Background(lipgloss.Color("236")).
		Padding(0, 2)

	inactiveTabStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")).
		Padding(0, 2)

	tabGapStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("238"))

	cardStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2).
		MarginRight(1)

	cardTitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205"))

	cardValueStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("86"))

	statusBarStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(1)

	errorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		Bold(true)

	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	tableHeaderStyle = lipgloss.NewStyle().
		Bold(true).
		Underline(true).
		Foreground(lipgloss.Color("252"))

	tableRowStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	tableTotalsStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("86"))

	emptyMessageStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Italic(true)

	selectedRowStyle = lipgloss.NewStyle().
		Background(lipgloss.Color("237")).
		Foreground(lipgloss.Color("252"))

	helpOverlayStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 3).
		Width(50)

	helpTitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	filterActiveStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("226"))

	warningStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("226"))

	criticalStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		Bold(true)

	successStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("46"))
)

func statusColor(status string) string {
	switch status {
	case "active", "running":
		return "46"
	case "stopped", "error":
		return "196"
	default:
		return "226"
	}
}

func errorRateColor(rate float64) string {
	switch {
	case rate < 1.0:
		return "46"
	case rate <= 5.0:
		return "226"
	default:
		return "196"
	}
}

func severityColor(severity string) string {
	switch severity {
	case "warning":
		return "226"
	case "critical":
		return "196"
	case "ok":
		return "46"
	default:
		return "241"
	}
}
