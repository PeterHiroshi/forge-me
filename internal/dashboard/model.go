package dashboard

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/PeterHiroshi/cfmon/internal/api"
)

// Model is the main bubbletea model for the dashboard.
type Model struct {
	client          *api.Client
	accountID       string
	activeTab       TabID
	data            *DashboardData
	loading         bool
	err             error
	width           int
	height          int
	scrollOffset    int
	refreshInterval time.Duration
	spinner         spinner.Model
}

// NewModel creates a new dashboard model.
func NewModel(client *api.Client, accountID string, refresh time.Duration) Model {
	if refresh < MinRefreshInterval {
		refresh = MinRefreshInterval
	}

	s := spinner.New()
	s.Spinner = spinner.Dot

	return Model{
		client:          client,
		accountID:       accountID,
		activeTab:       TabOverview,
		loading:         true,
		refreshInterval: refresh,
		spinner:         s,
	}
}

// Init starts the initial data fetch and spinner.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		fetchData(m.client, m.accountID),
		m.spinner.Tick,
	)
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case msg.Type == tea.KeyCtrlC || (msg.Type == tea.KeyRunes && len(msg.Runes) == 1 && msg.Runes[0] == 'q'):
			return m, tea.Quit

		case msg.Type == tea.KeyRunes && len(msg.Runes) == 1 && msg.Runes[0] == 'r':
			m.loading = true
			return m, fetchData(m.client, m.accountID)

		case msg.Type == tea.KeyRunes && len(msg.Runes) == 1 && msg.Runes[0] == 'j',
			msg.Type == tea.KeyDown:
			if m.activeTab != TabOverview {
				m.scrollOffset++
			}
			return m, nil

		case msg.Type == tea.KeyRunes && len(msg.Runes) == 1 && msg.Runes[0] == 'k',
			msg.Type == tea.KeyUp:
			if m.activeTab != TabOverview && m.scrollOffset > 0 {
				m.scrollOffset--
			}
			return m, nil

		case msg.Type == tea.KeyTab:
			m.activeTab = (m.activeTab + 1) % tabCount
			m.scrollOffset = 0
			return m, nil

		case msg.Type == tea.KeyShiftTab:
			m.activeTab = (m.activeTab - 1 + tabCount) % tabCount
			m.scrollOffset = 0
			return m, nil

		case msg.Type == tea.KeyRunes && len(msg.Runes) == 1 && msg.Runes[0] == '1':
			m.activeTab = TabOverview
			m.scrollOffset = 0
			return m, nil

		case msg.Type == tea.KeyRunes && len(msg.Runes) == 1 && msg.Runes[0] == '2':
			m.activeTab = TabWorkers
			m.scrollOffset = 0
			return m, nil

		case msg.Type == tea.KeyRunes && len(msg.Runes) == 1 && msg.Runes[0] == '3':
			m.activeTab = TabContainers
			m.scrollOffset = 0
			return m, nil
		}

	case dataMsg:
		m.data = msg.data
		m.loading = false
		m.err = nil
		return m, tickCmd(m.refreshInterval)

	case errMsg:
		m.err = msg.err
		m.loading = false
		return m, tickCmd(m.refreshInterval)

	case tickMsg:
		m.loading = true
		return m, fetchData(m.client, m.accountID)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

// View renders the dashboard.
func (m Model) View() string {
	if m.width == 0 {
		return "Initializing..."
	}

	var b strings.Builder

	b.WriteString(titleStyle.Render("cfmon dashboard"))
	b.WriteString(" ")
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(m.accountID))
	b.WriteString("\n\n")

	b.WriteString(m.renderTabs())
	b.WriteString("\n\n")

	if m.loading && m.data == nil {
		b.WriteString(m.spinner.View() + " Loading dashboard data...")
	} else if m.err != nil && m.data == nil {
		b.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
	} else {
		switch m.activeTab {
		case TabOverview:
			b.WriteString(m.renderOverview())
		case TabWorkers:
			b.WriteString(m.renderWorkers())
		case TabContainers:
			b.WriteString(m.renderContainers())
		}
	}

	b.WriteString("\n")
	b.WriteString(m.renderStatusBar())

	return b.String()
}

func (m Model) renderTabs() string {
	var tabs []string
	for i := TabID(0); i < tabCount; i++ {
		label := fmt.Sprintf(" %d %s ", i+1, i.String())
		if i == m.activeTab {
			tabs = append(tabs, activeTabStyle.Render(label))
		} else {
			tabs = append(tabs, inactiveTabStyle.Render(label))
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
}

func (m Model) renderOverview() string {
	if m.data == nil {
		return "No data available"
	}

	var b strings.Builder

	gaugeWidth := m.width - 20
	if gaugeWidth < 10 {
		gaugeWidth = 10
	}
	if gaugeWidth > 50 {
		gaugeWidth = 50
	}
	b.WriteString(cardTitleStyle.Render("Health Score"))
	b.WriteString("\n")
	b.WriteString(RenderGauge(m.data.HealthScore, gaugeWidth))
	b.WriteString("\n")
	if m.data.HealthStatus != "" {
		b.WriteString(fmt.Sprintf("Status: %s", m.data.HealthStatus))
	}
	b.WriteString("\n\n")

	workerCard := cardStyle.Render(
		cardTitleStyle.Render("Workers") + "\n" +
			cardValueStyle.Render(fmt.Sprintf("%d", len(m.data.Workers))),
	)

	containerCard := cardStyle.Render(
		cardTitleStyle.Render("Containers") + "\n" +
			cardValueStyle.Render(fmt.Sprintf("%d", len(m.data.Containers))),
	)

	errorCard := cardStyle.Render(
		cardTitleStyle.Render("Error Rate") + "\n" +
			cardValueStyle.Render(fmt.Sprintf("%.1f%%", m.data.ErrorRate)),
	)

	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, workerCard, containerCard, errorCard))

	return b.String()
}

func (m Model) renderStatusBar() string {
	left := "q: quit  r: refresh  Tab/1-3: switch tabs"
	if m.activeTab != TabOverview {
		left += "  j/k: scroll"
	}
	right := ""
	if m.loading {
		right = m.spinner.View() + " refreshing..."
	}

	gap := m.width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 0 {
		gap = 0
	}

	return statusBarStyle.Render(left + strings.Repeat(" ", gap) + right)
}
