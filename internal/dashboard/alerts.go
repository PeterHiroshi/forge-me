package dashboard

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/PeterHiroshi/cfmon/internal/monitor"
)

func (m Model) renderAlerts() string {
	if m.data == nil {
		return "No data available"
	}

	alerts := m.data.Alerts
	if m.filterText != "" {
		alerts = filterAlerts(alerts, m.filterText)
	}

	var b strings.Builder

	// Filter input
	if m.filterMode {
		b.WriteString("Filter: " + m.filterInput.View())
		b.WriteString("\n\n")
	} else if m.filterText != "" {
		b.WriteString(filterActiveStyle.Render("filter: " + m.filterText))
		b.WriteString("\n\n")
	}

	// Empty state
	if len(alerts) == 0 && m.filterText == "" {
		b.WriteString(successStyle.Render("All systems healthy — no alerts"))
		b.WriteString("\n")
		b.WriteString(m.renderEventLog())
		return b.String()
	}
	if len(alerts) == 0 {
		b.WriteString(emptyMessageStyle.Render("No alerts match filter"))
		b.WriteString("\n")
		b.WriteString(m.renderEventLog())
		return b.String()
	}

	// Summary
	b.WriteString(m.renderAlertSummary(alerts))
	b.WriteString("\n\n")

	// Alert rows with scrolling
	visibleRows := m.height - 18
	if visibleRows < 3 {
		visibleRows = 3
	}

	maxScroll := len(alerts) - visibleRows
	if maxScroll < 0 {
		maxScroll = 0
	}
	offset := m.scrollOffset
	if offset > maxScroll {
		offset = maxScroll
	}
	end := offset + visibleRows
	if end > len(alerts) {
		end = len(alerts)
	}

	for i, a := range alerts[offset:end] {
		row := formatAlertRow(a)
		if offset+i == m.selectedRow {
			b.WriteString(selectedRowStyle.Render(row))
		} else {
			style := warningStyle
			if a.Severity == "critical" {
				style = criticalStyle
			}
			b.WriteString(style.Render(row))
		}
		b.WriteString("\n")
	}

	// Scroll indicator
	if len(alerts) > visibleRows {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(
			fmt.Sprintf("  showing %d-%d of %d (j/k to scroll)", offset+1, end, len(alerts))))
		b.WriteString("\n")
	}

	// Event log
	b.WriteString(m.renderEventLog())

	return b.String()
}

func (m Model) renderAlertSummary(alerts []monitor.Alert) string {
	var warnings, criticals int
	for _, a := range alerts {
		switch a.Severity {
		case "warning":
			warnings++
		case "critical":
			criticals++
		}
	}

	var parts []string
	if warnings > 0 {
		label := "warning"
		if warnings > 1 {
			label = "warnings"
		}
		parts = append(parts, warningStyle.Render(fmt.Sprintf("%d %s", warnings, label)))
	}
	if criticals > 0 {
		parts = append(parts, criticalStyle.Render(fmt.Sprintf("%d critical", criticals)))
	}

	return strings.Join(parts, ", ")
}

func formatAlertRow(a monitor.Alert) string {
	label := severityLabel(a.Severity)
	return fmt.Sprintf("[%s] %s %q — %s %.1f%% (threshold: %.1f%%)",
		label, a.ResourceType, a.ResourceName, a.Metric, a.Value, a.Threshold)
}

func (m Model) renderEventLog() string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(cardTitleStyle.Render("Event Log"))
	b.WriteString("\n")

	if len(m.events) == 0 {
		b.WriteString(emptyMessageStyle.Render("  No events yet"))
		return b.String()
	}

	// Show last 20 events (most recent first)
	start := len(m.events) - 20
	if start < 0 {
		start = 0
	}
	for i := len(m.events) - 1; i >= start; i-- {
		e := m.events[i]
		ts := e.Time.Format("15:04:05")
		style := lipgloss.NewStyle().Foreground(lipgloss.Color(severityColor(e.Severity)))
		b.WriteString(fmt.Sprintf("  %s %s\n", ts, style.Render(e.Text)))
	}

	return b.String()
}
