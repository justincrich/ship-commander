package views

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ship-commander/sc3/internal/tui/components"
	"github.com/ship-commander/sc3/internal/tui/theme"
)

const (
	fleetMonitorCompactThreshold = 120
	fleetMonitorDefaultWidth     = 120
	fleetMonitorAutoRefreshEvery = 2 * time.Second
)

// FleetMonitorLayout controls standard vs compact rendering.
type FleetMonitorLayout string

const (
	// FleetMonitorLayoutStandard renders full columns.
	FleetMonitorLayoutStandard FleetMonitorLayout = "standard"
	// FleetMonitorLayoutCompact renders abbreviated rows.
	FleetMonitorLayoutCompact FleetMonitorLayout = "compact"
)

// FleetMonitorQuickAction identifies direct keyboard actions for monitor view.
type FleetMonitorQuickAction string

const (
	// FleetMonitorQuickActionNone indicates no resolved action.
	FleetMonitorQuickActionNone FleetMonitorQuickAction = ""
	// FleetMonitorQuickActionBridge drills into ship bridge.
	FleetMonitorQuickActionBridge FleetMonitorQuickAction = "bridge"
	// FleetMonitorQuickActionInbox opens inbox.
	FleetMonitorQuickActionInbox FleetMonitorQuickAction = "inbox"
	// FleetMonitorQuickActionHelp opens help.
	FleetMonitorQuickActionHelp FleetMonitorQuickAction = "help"
	// FleetMonitorQuickActionFleet returns to fleet overview.
	FleetMonitorQuickActionFleet FleetMonitorQuickAction = "fleet"
)

// FleetMonitorTickMsg is emitted by the monitor auto-refresh command.
type FleetMonitorTickMsg struct {
	At time.Time
}

// FleetMonitorConfig contains all rendering inputs for Fleet Monitor.
type FleetMonitorConfig struct {
	Width              int
	Rows               []components.ShipStatusRow
	SelectedIndex      int
	ActiveShipCount    int
	TotalAgents        int
	FleetWaveCurrent   int
	FleetWaveTotal     int
	PendingQuestions   int
	ToolbarHighlighted int
}

type fleetMonitorListItem struct {
	content string
}

func (item fleetMonitorListItem) FilterValue() string {
	return item.content
}

type fleetMonitorListDelegate struct{}

func (delegate fleetMonitorListDelegate) Height() int {
	return 1
}

func (delegate fleetMonitorListDelegate) Spacing() int {
	return 0
}

func (delegate fleetMonitorListDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd {
	return nil
}

func (delegate fleetMonitorListDelegate) Render(writer io.Writer, _ list.Model, _ int, item list.Item) {
	row, ok := item.(fleetMonitorListItem)
	if !ok {
		return
	}
	if _, err := io.WriteString(writer, row.content); err != nil {
		return
	}
}

// ResolveFleetMonitorLayout resolves responsive layout mode.
func ResolveFleetMonitorLayout(width int) FleetMonitorLayout {
	if width > 0 && width < fleetMonitorCompactThreshold {
		return FleetMonitorLayoutCompact
	}
	return FleetMonitorLayoutStandard
}

// FleetMonitorToolbarButtons returns monitor-specific wayfinding toolbar buttons.
func FleetMonitorToolbarButtons() []components.ToolbarButton {
	return []components.ToolbarButton{
		{Key: "Enter", Label: "Bridge", Enabled: true},
		{Key: "i", Label: "Inbox", Enabled: true},
		{Key: "?", Label: "Help", Enabled: true},
		{Key: "Esc", Label: "Fleet", Enabled: true},
	}
}

// FleetMonitorEnterTarget returns drill-down destination for Enter action.
func FleetMonitorEnterTarget() string {
	return "ship_bridge"
}

// FleetMonitorQuickActionForKey resolves direct key actions.
func FleetMonitorQuickActionForKey(msg tea.KeyMsg) FleetMonitorQuickAction {
	switch strings.ToLower(strings.TrimSpace(msg.String())) {
	case "enter":
		return FleetMonitorQuickActionBridge
	case "i":
		return FleetMonitorQuickActionInbox
	case "?":
		return FleetMonitorQuickActionHelp
	case "esc":
		return FleetMonitorQuickActionFleet
	default:
		return FleetMonitorQuickActionNone
	}
}

// FleetMonitorAutoRefreshInterval returns the default refresh cadence.
func FleetMonitorAutoRefreshInterval() time.Duration {
	return fleetMonitorAutoRefreshEvery
}

// FleetMonitorAutoRefreshCmd emits a periodic refresh tick.
func FleetMonitorAutoRefreshCmd(interval time.Duration) tea.Cmd {
	if interval <= 0 {
		interval = FleetMonitorAutoRefreshInterval()
	}
	return tea.Tick(interval, func(at time.Time) tea.Msg {
		return FleetMonitorTickMsg{At: at}
	})
}

// BuildFleetMonitorRows filters to launched ships and applies default widths.
func BuildFleetMonitorRows(rows []components.ShipStatusRow, width int, selectedIndex int) []components.ShipStatusRow {
	launched := make([]components.ShipStatusRow, 0, len(rows))
	for _, row := range rows {
		if !isLaunchedStatus(row.Status) {
			continue
		}
		launched = append(launched, row)
	}

	selected := normalizeSelectedIndex(selectedIndex, len(launched))
	layout := ResolveFleetMonitorLayout(width)
	rowWidth := width
	if rowWidth <= 0 {
		rowWidth = fleetMonitorDefaultWidth
	}
	if layout == FleetMonitorLayoutCompact {
		rowWidth = 80
	}

	for idx := range launched {
		launched[idx].Selected = idx == selected
		launched[idx].Width = rowWidth
	}

	return launched
}

// RenderFleetMonitor renders monitor header, ship grid, and toolbar.
func RenderFleetMonitor(config FleetMonitorConfig) string {
	width := config.Width
	if width <= 0 {
		width = fleetMonitorDefaultWidth
	}
	layout := ResolveFleetMonitorLayout(width)
	rows := BuildFleetMonitorRows(config.Rows, width, config.SelectedIndex)

	header := renderFleetMonitorHeader(config, rows, layout)
	grid := renderFleetMonitorGrid(rows, width)
	toolbar := components.RenderNavigableToolbar(FleetMonitorToolbarButtons(), config.ToolbarHighlighted)

	return lipgloss.JoinVertical(lipgloss.Left, header, grid, toolbar)
}

func renderFleetMonitorHeader(config FleetMonitorConfig, rows []components.ShipStatusRow, layout FleetMonitorLayout) string {
	activeShips := config.ActiveShipCount
	if activeShips <= 0 {
		activeShips = len(rows)
	}
	totalAgents := config.TotalAgents
	if totalAgents <= 0 {
		for _, row := range rows {
			totalAgents += row.CrewCount
		}
	}
	waveCurrent := shipWaveNumber(config.FleetWaveCurrent)
	waveTotal := config.FleetWaveTotal
	if waveTotal < waveCurrent {
		waveTotal = waveCurrent
	}

	pending := "0"
	if config.PendingQuestions > 0 {
		pending = lipgloss.NewStyle().Background(theme.PinkColor).Foreground(theme.BlackColor).Bold(true).Render(fmt.Sprintf("[%d]", config.PendingQuestions))
	}

	title := lipgloss.NewStyle().Foreground(theme.ButterscotchColor).Bold(true).Render("FLEET MONITOR")
	stats := fmt.Sprintf("Active Ships: %d    Total Agents: %d    Fleet Progress: Wave %d/%d    Pending: %s", activeShips, totalAgents, waveCurrent, waveTotal, pending)
	if layout == FleetMonitorLayoutCompact {
		stats = fmt.Sprintf("Ships: %d  Agents: %d  Wave: %d/%d  Pending: %s", activeShips, totalAgents, waveCurrent, waveTotal, pending)
	}

	return theme.PanelBorder.Render(lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		lipgloss.NewStyle().Foreground(theme.SpaceWhiteColor).Render(stats),
	))
}

func renderFleetMonitorGrid(rows []components.ShipStatusRow, width int) string {
	if len(rows) == 0 {
		empty := lipgloss.NewStyle().Foreground(theme.GalaxyGrayColor).Faint(true).Render("No launched ships to monitor")
		return theme.PanelBorder.Render(panelWithTitle("Ship Status", empty))
	}

	items := make([]list.Item, 0, len(rows))
	for _, row := range rows {
		items = append(items, fleetMonitorListItem{content: components.RenderShipStatusRow(row)})
	}

	listHeight := len(rows) + 2
	if listHeight < 8 {
		listHeight = 8
	}
	listWidth := width - 4
	if listWidth < 60 {
		listWidth = 60
	}

	model := list.New(items, fleetMonitorListDelegate{}, listWidth, listHeight)
	model.SetShowTitle(false)
	model.SetShowStatusBar(false)
	model.SetShowHelp(false)
	model.SetShowPagination(false)
	model.SetFilteringEnabled(false)

	return theme.PanelBorder.Render(panelWithTitle("Ship Status Grid", model.View()))
}

func isLaunchedStatus(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "launched", "running", "in_progress", "halted", "complete", "completed", "done", "stuck", "has_issues":
		return true
	default:
		return false
	}
}
