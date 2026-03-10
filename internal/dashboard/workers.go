package dashboard

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) renderWorkers() string {
	if m.data == nil || len(m.data.Workers) == 0 {
		return emptyMessageStyle.Render("No workers found")
	}

	var b strings.Builder

	// Column widths
	nameW, statusW, reqW, errW, rateW, cpuW := 20, 10, 12, 10, 12, 10

	// Header
	header := fmt.Sprintf("%-*s %-*s %*s %*s %*s %*s",
		nameW, "Name", statusW, "Status", reqW, "Requests", errW, "Errors", rateW, "Error Rate", cpuW, "CPU (ms)")
	b.WriteString(tableHeaderStyle.Render(header))
	b.WriteString("\n")

	// Calculate visible rows
	visibleRows := m.height - 12
	if visibleRows < 3 {
		visibleRows = 3
	}

	maxScroll := len(m.data.Workers) - visibleRows
	if maxScroll < 0 {
		maxScroll = 0
	}
	offset := m.scrollOffset
	if offset > maxScroll {
		offset = maxScroll
	}

	end := offset + visibleRows
	if end > len(m.data.Workers) {
		end = len(m.data.Workers)
	}

	// Totals
	var totalReq, totalErr, totalCPU int
	for _, w := range m.data.Workers {
		totalReq += w.Requests
		totalErr += w.Errors
		totalCPU += w.CPUMS
	}

	// Rows
	for i, w := range m.data.Workers[offset:end] {
		var errRate float64
		if w.Requests > 0 {
			errRate = float64(w.Errors) / float64(w.Requests) * 100
		}

		statusStyled := lipgloss.NewStyle().Foreground(lipgloss.Color(statusColor(w.Status))).Render(fmt.Sprintf("%-*s", statusW, w.Status))
		errRateStyled := lipgloss.NewStyle().Foreground(lipgloss.Color(errorRateColor(errRate))).Render(fmt.Sprintf("%*.1f%%", rateW-1, errRate))

		row := fmt.Sprintf("%-*s %s %*d %*d %s %*d",
			nameW, truncate(w.Name, nameW), statusStyled, reqW, w.Requests, errW, w.Errors, errRateStyled, cpuW, w.CPUMS)
		if offset+i == m.selectedRow {
			b.WriteString(selectedRowStyle.Render(row))
		} else {
			b.WriteString(tableRowStyle.Render(row))
		}
		b.WriteString("\n")
	}

	// Scroll indicator
	if len(m.data.Workers) > visibleRows {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(
			fmt.Sprintf("  showing %d-%d of %d (j/k to scroll)", offset+1, end, len(m.data.Workers))))
		b.WriteString("\n")
	}

	// Totals row
	var totalErrRate float64
	if totalReq > 0 {
		totalErrRate = float64(totalErr) / float64(totalReq) * 100
	}
	totals := fmt.Sprintf("%-*s %-*s %*d %*d %*.1f%% %*d",
		nameW, "TOTAL", statusW, "", reqW, totalReq, errW, totalErr, rateW-1, totalErrRate, cpuW, totalCPU)
	b.WriteString(tableTotalsStyle.Render(totals))

	return b.String()
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max <= 3 {
		return s[:max]
	}
	return s[:max-3] + "..."
}
