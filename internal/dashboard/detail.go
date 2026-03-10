package dashboard

import (
	"fmt"
	"strings"
)

func (m Model) renderWorkerDetail() string {
	workers := m.data.Workers
	if m.filterText != "" {
		workers = filterWorkers(workers, m.filterText)
	}
	if m.selectedRow >= len(workers) || m.selectedRow < 0 {
		return "No worker selected"
	}
	w := workers[m.selectedRow]

	var errRate float64
	if w.Requests > 0 {
		errRate = float64(w.Errors) / float64(w.Requests) * 100
	}

	var b strings.Builder
	b.WriteString(cardTitleStyle.Render("Worker Detail"))
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("  Name:         %s\n", w.Name))
	b.WriteString(fmt.Sprintf("  ID:           %s\n", w.ID))
	b.WriteString(fmt.Sprintf("  Status:       %s\n", w.Status))
	b.WriteString(fmt.Sprintf("  Requests:     %d\n", w.Requests))
	b.WriteString(fmt.Sprintf("  Errors:       %d\n", w.Errors))
	b.WriteString(fmt.Sprintf("  Error Rate:   %.1f%%\n", errRate))
	b.WriteString(fmt.Sprintf("  CPU (ms):     %d\n", w.CPUMS))
	b.WriteString(fmt.Sprintf("  Success Rate: %.1f%%\n", w.SuccessRate))

	barWidth := 40
	b.WriteString("\n")
	b.WriteString("  Requests: ")
	b.WriteString(renderBar(w.Requests, w.Requests, barWidth))
	b.WriteString("\n")
	b.WriteString("  Errors:   ")
	if w.Requests > 0 {
		b.WriteString(renderBar(w.Errors, w.Requests, barWidth))
	} else {
		b.WriteString(renderBar(0, 1, barWidth))
	}

	return b.String()
}

func (m Model) renderContainerDetail() string {
	containers := m.data.Containers
	if m.filterText != "" {
		containers = filterContainers(containers, m.filterText)
	}
	if m.selectedRow >= len(containers) || m.selectedRow < 0 {
		return "No container selected"
	}
	c := containers[m.selectedRow]

	largeBarWidth := 40

	var b strings.Builder
	b.WriteString(cardTitleStyle.Render("Container Detail"))
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("  Name:       %s\n", c.Name))
	b.WriteString(fmt.Sprintf("  ID:         %s\n", c.ID))
	b.WriteString(fmt.Sprintf("  Status:     %s\n", c.Status))
	b.WriteString(fmt.Sprintf("  CPU (ms):   %d\n", c.CPUMS))
	b.WriteString(fmt.Sprintf("  Memory (MB):%d\n", c.MemoryMB))
	b.WriteString(fmt.Sprintf("  Requests:   %d\n", c.Requests))

	b.WriteString("\n")
	b.WriteString("  CPU:    ")
	b.WriteString(renderBar(c.CPUMS, c.CPUMS, largeBarWidth))
	b.WriteString("\n")
	b.WriteString("  Memory: ")
	b.WriteString(renderBar(c.MemoryMB, c.MemoryMB, largeBarWidth))

	return b.String()
}

func (m Model) renderAlertDetail() string {
	alerts := m.data.Alerts
	if m.filterText != "" {
		alerts = filterAlerts(alerts, m.filterText)
	}
	if m.selectedRow >= len(alerts) || m.selectedRow < 0 {
		return "No alert selected"
	}
	a := alerts[m.selectedRow]

	var b strings.Builder
	b.WriteString(cardTitleStyle.Render("Alert Detail"))
	b.WriteString("\n\n")

	sevStyle := warningStyle
	if a.Severity == "critical" {
		sevStyle = criticalStyle
	}
	b.WriteString(fmt.Sprintf("  Severity:     %s\n", sevStyle.Render(a.Severity)))
	b.WriteString(fmt.Sprintf("  Resource:     %s %q\n", a.ResourceType, a.ResourceName))
	b.WriteString(fmt.Sprintf("  Metric:       %s\n", a.Metric))
	b.WriteString(fmt.Sprintf("  Value:        %.1f%%\n", a.Value))
	b.WriteString(fmt.Sprintf("  Threshold:    %.1f%%\n", a.Threshold))
	if a.Message != "" {
		b.WriteString(fmt.Sprintf("  Message:      %s\n", a.Message))
	}

	// Show related resource detail
	b.WriteString("\n")
	switch a.ResourceType {
	case "worker":
		for _, w := range m.data.Workers {
			if w.Name == a.ResourceName {
				b.WriteString(cardTitleStyle.Render("Related Worker"))
				b.WriteString("\n")
				b.WriteString(fmt.Sprintf("  Name:       %s\n", w.Name))
				b.WriteString(fmt.Sprintf("  ID:         %s\n", w.ID))
				b.WriteString(fmt.Sprintf("  Status:     %s\n", w.Status))
				b.WriteString(fmt.Sprintf("  Requests:   %d\n", w.Requests))
				b.WriteString(fmt.Sprintf("  Errors:     %d\n", w.Errors))
				b.WriteString(fmt.Sprintf("  CPU (ms):   %d\n", w.CPUMS))
				break
			}
		}
	case "container":
		for _, c := range m.data.Containers {
			if c.Name == a.ResourceName {
				b.WriteString(cardTitleStyle.Render("Related Container"))
				b.WriteString("\n")
				b.WriteString(fmt.Sprintf("  Name:       %s\n", c.Name))
				b.WriteString(fmt.Sprintf("  ID:         %s\n", c.ID))
				b.WriteString(fmt.Sprintf("  Status:     %s\n", c.Status))
				b.WriteString(fmt.Sprintf("  CPU (ms):   %d\n", c.CPUMS))
				b.WriteString(fmt.Sprintf("  Memory (MB):%d\n", c.MemoryMB))
				break
			}
		}
	}

	return b.String()
}
