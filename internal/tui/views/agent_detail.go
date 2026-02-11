package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ship-commander/sc3/internal/tui/components"
	"github.com/ship-commander/sc3/internal/tui/theme"
)

const agentDetailDefaultWidth = 120

// AgentDetailConfig captures all render input for the Agent Detail drill-down.
type AgentDetailConfig struct {
	Width              int
	AgentName          string
	Role               string
	ShipName           string
	Harness            string
	Model              string
	MissionID          string
	Phase              string
	Elapsed            string
	Status             string
	OutputLines        []string
	AutoScroll         bool
	StuckTimestamp     string
	LastGateFailure    string
	TimeoutInfo        string
	LastMeaningfulLine string
	ToolbarHighlighted int
}

// AgentDetailQuickAction captures direct action keys for agent detail.
type AgentDetailQuickAction string

const (
	// AgentDetailQuickActionNone indicates no action key match.
	AgentDetailQuickActionNone AgentDetailQuickAction = ""
	// AgentDetailQuickActionHalt halts the current agent.
	AgentDetailQuickActionHalt AgentDetailQuickAction = "halt"
	// AgentDetailQuickActionRetry retries the mission.
	AgentDetailQuickActionRetry AgentDetailQuickAction = "retry"
	// AgentDetailQuickActionIgnoreStuck acknowledges stuck warning.
	AgentDetailQuickActionIgnoreStuck AgentDetailQuickAction = "ignore_stuck"
	// AgentDetailQuickActionHelp opens help overlay.
	AgentDetailQuickActionHelp AgentDetailQuickAction = "help"
	// AgentDetailQuickActionBack returns to previous screen.
	AgentDetailQuickActionBack AgentDetailQuickAction = "back"
)

// AgentDetailToolbarButtons returns action buttons for the drill-down toolbar.
func AgentDetailToolbarButtons() []components.ToolbarButton {
	return []components.ToolbarButton{
		{Key: "h", Label: "Halt", Enabled: true},
		{Key: "r", Label: "Retry", Enabled: true},
		{Key: "i", Label: "Ignore Stuck", Enabled: true},
		{Key: "?", Label: "Help", Enabled: true},
		{Key: "Esc", Label: "Back", Enabled: true},
	}
}

// AgentDetailQuickActionForKey maps key messages to quick actions.
func AgentDetailQuickActionForKey(msg tea.KeyMsg) AgentDetailQuickAction {
	switch strings.ToLower(strings.TrimSpace(msg.String())) {
	case "h":
		return AgentDetailQuickActionHalt
	case "r":
		return AgentDetailQuickActionRetry
	case "i":
		return AgentDetailQuickActionIgnoreStuck
	case "?":
		return AgentDetailQuickActionHelp
	case "esc":
		return AgentDetailQuickActionBack
	default:
		return AgentDetailQuickActionNone
	}
}

// RenderAgentDetail renders full-screen agent profile, output viewport, optional error context, and toolbar.
func RenderAgentDetail(config AgentDetailConfig) string {
	width := config.Width
	if width <= 0 {
		width = agentDetailDefaultWidth
	}

	header := renderAgentDetailHeader(config)
	outputPanel := renderAgentDetailOutput(config, width)
	sections := []string{header, outputPanel}
	if shouldShowAgentErrorContext(config.Status) {
		sections = append(sections, renderAgentDetailErrorContext(config))
	}
	sections = append(sections, components.RenderNavigableToolbar(AgentDetailToolbarButtons(), config.ToolbarHighlighted))
	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func renderAgentDetailHeader(config AgentDetailConfig) string {
	name := strings.TrimSpace(config.AgentName)
	if name == "" {
		name = "Unnamed Agent"
	}
	role := strings.TrimSpace(config.Role)
	if role == "" {
		role = "ENSIGN"
	}
	ship := strings.TrimSpace(config.ShipName)
	if ship == "" {
		ship = "Unassigned"
	}
	harness := strings.TrimSpace(config.Harness)
	if harness == "" {
		harness = "unknown"
	}
	model := strings.TrimSpace(config.Model)
	if model == "" {
		model = "default"
	}
	missionID := strings.TrimSpace(config.MissionID)
	if missionID == "" {
		missionID = "M-000"
	}
	phase := strings.TrimSpace(config.Phase)
	if phase == "" {
		phase = "PENDING"
	}
	elapsed := strings.TrimSpace(config.Elapsed)
	if elapsed == "" {
		elapsed = "00:00"
	}

	roleBadge := renderAgentRoleBadge(role)
	status := components.RenderStatusBadge(mapAgentDetailStatusToBadge(config.Status), components.WithBadgeBold(true))

	lineOne := lipgloss.JoinHorizontal(
		lipgloss.Left,
		lipgloss.NewStyle().Foreground(theme.ButterscotchColor).Bold(true).Render(name),
		"  ",
		roleBadge,
		"  ",
		lipgloss.NewStyle().Foreground(theme.LightGrayColor).Render(ship),
		"  ",
		lipgloss.NewStyle().Foreground(theme.BlueColor).Render(harness+"-"+model),
		"  ",
		lipgloss.NewStyle().Foreground(theme.GoldColor).Bold(true).Render(missionID),
	)

	lineTwo := lipgloss.JoinHorizontal(
		lipgloss.Left,
		lipgloss.NewStyle().Foreground(theme.LightGrayColor).Render("Phase: "+phase),
		"   ",
		lipgloss.NewStyle().Foreground(theme.LightGrayColor).Render("Elapsed: "+elapsed),
		"   ",
		status,
	)

	return theme.PanelBorder.Render(lipgloss.JoinVertical(lipgloss.Left, lineOne, lineTwo))
}

func renderAgentRoleBadge(role string) string {
	label := strings.ToUpper(strings.TrimSpace(role))
	if label == "" {
		label = "ENSIGN"
	}

	style := lipgloss.NewStyle().Background(theme.ButterscotchColor).Foreground(theme.BlackColor).Bold(true)
	switch label {
	case "COMMANDER":
		style = lipgloss.NewStyle().Background(theme.BlueColor).Foreground(theme.BlackColor).Bold(true)
	case "CAPTAIN":
		style = lipgloss.NewStyle().Background(theme.GoldColor).Foreground(theme.BlackColor).Bold(true)
	}

	return style.Render(label)
}

func renderAgentDetailOutput(config AgentDetailConfig, width int) string {
	lines := config.OutputLines
	if len(lines) == 0 {
		lines = []string{"No output captured yet."}
	}

	rendered := make([]string, 0, len(lines))
	for i, line := range lines {
		rendered = append(rendered,
			lipgloss.JoinHorizontal(
				lipgloss.Left,
				lipgloss.NewStyle().Foreground(theme.GalaxyGrayColor).Render(fmt.Sprintf("%4d", i+1)),
				"  ",
				line,
			),
		)
	}

	viewWidth := width - 6
	if viewWidth < 24 {
		viewWidth = 24
	}
	viewHeight := 15
	if width < 80 {
		viewHeight = 10
	}

	vp := viewport.New(viewWidth, viewHeight)
	vp.SetContent(strings.Join(rendered, "\n"))
	if config.AutoScroll {
		vp.GotoBottom()
	} else {
		vp.GotoTop()
	}

	title := "Agent Output"
	if config.AutoScroll {
		title = "Agent Output [AUTO]"
	}
	return theme.PanelBorder.Render(panelWithTitle(title, vp.View()))
}

func shouldShowAgentErrorContext(status string) bool {
	normalized := strings.ToLower(strings.TrimSpace(status))
	return normalized == "stuck" || normalized == "dead"
}

func renderAgentDetailErrorContext(config AgentDetailConfig) string {
	timestamp := strings.TrimSpace(config.StuckTimestamp)
	if timestamp == "" {
		timestamp = "Unknown"
	}
	lastGate := strings.TrimSpace(config.LastGateFailure)
	if lastGate == "" {
		lastGate = "Unavailable"
	}
	timeout := strings.TrimSpace(config.TimeoutInfo)
	if timeout == "" {
		timeout = "Unavailable"
	}
	lastLine := strings.TrimSpace(config.LastMeaningfulLine)
	if lastLine == "" {
		lastLine = "Unavailable"
	}

	header := lipgloss.NewStyle().Foreground(theme.RedAlertColor).Bold(true).Render(theme.IconAlert + " STUCK DETECTED")
	body := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		fmt.Sprintf("Timestamp: %s", timestamp),
		fmt.Sprintf("Last gate: %s", lastGate),
		fmt.Sprintf("Timeout: %s", timeout),
		fmt.Sprintf("Last output: %s", lastLine),
	)

	panel := lipgloss.NewStyle().
		Border(lipgloss.ThickBorder()).
		BorderForeground(theme.RedAlertColor).
		Render(body)
	return panel
}

func mapAgentDetailStatusToBadge(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "done", "complete", "completed":
		return "done"
	case "waiting", "idle":
		return "waiting"
	case "stuck":
		return "stuck"
	case "dead", "failed", "halted":
		return "halted"
	default:
		return "running"
	}
}
