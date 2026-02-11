package views

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/ship-commander/sc3/internal/tui/components"
	"github.com/ship-commander/sc3/internal/tui/theme"
)

const (
	planReviewCompactThreshold = 120
	planReviewDefaultWidth     = 120
	planReviewPanelGap         = 1
	planReviewManifestHeight   = 16
	planReviewAnalysisHeight   = 14
)

// PlanReviewLayout identifies responsive rendering mode.
type PlanReviewLayout string

const (
	// PlanReviewLayoutStandard renders side-by-side manifest and analysis.
	PlanReviewLayoutStandard PlanReviewLayout = "standard"
	// PlanReviewLayoutCompact renders stacked manifest/analysis panels.
	PlanReviewLayoutCompact PlanReviewLayout = "compact"
)

// PlanReviewCoverageStatus identifies per-use-case coverage state.
type PlanReviewCoverageStatus string

const (
	// PlanReviewCoverageCovered marks a fully covered use case.
	PlanReviewCoverageCovered PlanReviewCoverageStatus = "covered"
	// PlanReviewCoveragePartial marks a partially covered use case.
	PlanReviewCoveragePartial PlanReviewCoverageStatus = "partial"
	// PlanReviewCoverageUncovered marks an uncovered use case.
	PlanReviewCoverageUncovered PlanReviewCoverageStatus = "uncovered"
)

// PlanReviewAnalysisTab identifies compact-mode analysis panel tabs.
type PlanReviewAnalysisTab string

const (
	// PlanReviewAnalysisCoverage shows coverage matrix tab.
	PlanReviewAnalysisCoverage PlanReviewAnalysisTab = "coverage"
	// PlanReviewAnalysisDependencies shows dependency graph tab.
	PlanReviewAnalysisDependencies PlanReviewAnalysisTab = "dependencies"
)

// PlanReviewMission captures one mission manifest row in the left viewport.
type PlanReviewMission struct {
	ID             string
	Title          string
	Classification string
	Wave           int
	UseCaseRefs    []string
	ACTotal        int
	SurfaceArea    string
}

// PlanReviewCoverageRow captures one use-case mapping in the coverage matrix.
type PlanReviewCoverageRow struct {
	UseCaseID  string
	MissionIDs []string
	Status     PlanReviewCoverageStatus
}

// PlanReviewDependencyMission captures one dependency graph node.
type PlanReviewDependencyMission struct {
	ID           string
	Title        string
	Status       string
	Dependencies []string
}

// PlanReviewDependencyWave captures dependency graph rows grouped by wave.
type PlanReviewDependencyWave struct {
	Wave     int
	Missions []PlanReviewDependencyMission
}

// PlanReviewConfig contains all render-time inputs for Plan Review.
type PlanReviewConfig struct {
	Width              int
	ShipName           string
	DirectiveTitle     string
	Missions           []PlanReviewMission
	Coverage           []PlanReviewCoverageRow
	Dependencies       []PlanReviewDependencyWave
	SignoffsDone       int
	SignoffsTotal      int
	AnalysisTab        PlanReviewAnalysisTab
	ToolbarHighlighted int
	FeedbackMode       bool
	FeedbackText       string
}

// PlanReviewQuickAction captures direct action keys supported in this view.
type PlanReviewQuickAction string

const (
	// PlanReviewQuickActionNone indicates no key match.
	PlanReviewQuickActionNone PlanReviewQuickAction = ""
	// PlanReviewQuickActionApprove approves the plan.
	PlanReviewQuickActionApprove PlanReviewQuickAction = "approve"
	// PlanReviewQuickActionFeedback opens feedback mode.
	PlanReviewQuickActionFeedback PlanReviewQuickAction = "feedback"
	// PlanReviewQuickActionShelve shelves the plan.
	PlanReviewQuickActionShelve PlanReviewQuickAction = "shelve"
	// PlanReviewQuickActionHelp opens help overlay.
	PlanReviewQuickActionHelp PlanReviewQuickAction = "help"
	// PlanReviewQuickActionReadyRoom returns to ready room.
	PlanReviewQuickActionReadyRoom PlanReviewQuickAction = "ready_room"
	// PlanReviewQuickActionCoverageTab switches compact analysis to coverage.
	PlanReviewQuickActionCoverageTab PlanReviewQuickAction = "coverage_tab"
	// PlanReviewQuickActionDependenciesTab switches compact analysis to dependencies.
	PlanReviewQuickActionDependenciesTab PlanReviewQuickAction = "dependencies_tab"
)

// ResolvePlanReviewLayout returns compact/standard mode for the given width.
func ResolvePlanReviewLayout(width int) PlanReviewLayout {
	if width > 0 && width < planReviewCompactThreshold {
		return PlanReviewLayoutCompact
	}
	return PlanReviewLayoutStandard
}

// PlanReviewToolbarButtons returns canonical approval actions for Plan Review.
func PlanReviewToolbarButtons() []components.ToolbarButton {
	return []components.ToolbarButton{
		{Key: "a", Label: "Approve", Enabled: true},
		{Key: "f", Label: "Feedback", Enabled: true},
		{Key: "s", Label: "Shelve", Enabled: true},
		{Key: "?", Label: "Help", Enabled: true},
		{Key: "Esc", Label: "Ready Room", Enabled: true},
	}
}

// PlanReviewQuickActionForKey resolves direct keyboard shortcuts.
func PlanReviewQuickActionForKey(msg tea.KeyMsg) PlanReviewQuickAction {
	switch strings.ToLower(strings.TrimSpace(msg.String())) {
	case "a":
		return PlanReviewQuickActionApprove
	case "f":
		return PlanReviewQuickActionFeedback
	case "s":
		return PlanReviewQuickActionShelve
	case "?":
		return PlanReviewQuickActionHelp
	case "esc":
		return PlanReviewQuickActionReadyRoom
	case "1":
		return PlanReviewQuickActionCoverageTab
	case "2":
		return PlanReviewQuickActionDependenciesTab
	default:
		return PlanReviewQuickActionNone
	}
}

// RenderPlanReview renders the full-screen Plan Review view in standard/compact layout.
func RenderPlanReview(config PlanReviewConfig) string {
	width := config.Width
	if width <= 0 {
		width = planReviewDefaultWidth
	}

	layout := ResolvePlanReviewLayout(width)
	header := renderPlanReviewHeader(config)
	toolbar := components.RenderNavigableToolbar(PlanReviewToolbarButtons(), config.ToolbarHighlighted)

	if layout == PlanReviewLayoutCompact {
		manifestPanel := renderManifestPanel(config.Missions, width, 10)
		analysisPanel := renderCompactAnalysisPanel(config, width)
		blocks := []string{header, manifestPanel, analysisPanel}
		if config.FeedbackMode {
			blocks = append(blocks, renderFeedbackInput(config.FeedbackText, width))
		}
		blocks = append(blocks, toolbar)
		return lipgloss.JoinVertical(lipgloss.Left, blocks...)
	}

	leftWidth := int(float64(width)*0.5) - planReviewPanelGap
	if leftWidth < 50 {
		leftWidth = 50
	}
	rightWidth := width - leftWidth - planReviewPanelGap
	if rightWidth < 50 {
		rightWidth = 50
	}

	manifestPanel := lipgloss.NewStyle().Width(leftWidth).Render(renderManifestPanel(config.Missions, leftWidth, planReviewManifestHeight))
	analysisPanel := lipgloss.NewStyle().Width(rightWidth).Render(renderStandardAnalysisPanel(config, rightWidth))
	content := lipgloss.JoinHorizontal(
		lipgloss.Top,
		manifestPanel,
		lipgloss.NewStyle().Width(planReviewPanelGap).Render(""),
		analysisPanel,
	)

	blocks := []string{header, content}
	if config.FeedbackMode {
		blocks = append(blocks, renderFeedbackInput(config.FeedbackText, width))
	}
	blocks = append(blocks, toolbar)
	return lipgloss.JoinVertical(lipgloss.Left, blocks...)
}

func renderPlanReviewHeader(config PlanReviewConfig) string {
	shipName := strings.TrimSpace(config.ShipName)
	if shipName == "" {
		shipName = "Unnamed Ship"
	}
	directive := strings.TrimSpace(config.DirectiveTitle)
	if directive == "" {
		directive = "No Directive"
	}

	title := lipgloss.NewStyle().
		Foreground(theme.ButterscotchColor).
		Bold(true).
		Render("PLAN REVIEW -- " + shipName)
	subtitle := lipgloss.NewStyle().
		Foreground(theme.BlueColor).
		Render("Directive: " + directive)

	stats := strings.Join([]string{
		fmt.Sprintf("Missions: %d", len(config.Missions)),
		fmt.Sprintf("Waves: %d", countWaves(config)),
		fmt.Sprintf("Coverage: %d%%", coveragePercent(config.Coverage)),
		fmt.Sprintf("Sign-offs: %d/%d", clampNonNegative(config.SignoffsDone), clampNonNegative(config.SignoffsTotal)),
	}, "   ")

	return theme.PanelBorder.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			title,
			subtitle,
			lipgloss.NewStyle().Foreground(theme.SpaceWhiteColor).Render(stats),
		),
	)
}

func renderManifestPanel(missions []PlanReviewMission, width int, height int) string {
	contentWidth := max(20, width-4)
	contentHeight := max(4, height)
	markdown := buildManifestMarkdown(missions)
	rendered := renderMarkdown(markdown, contentWidth)

	viewportModel := viewport.New(contentWidth, contentHeight)
	viewportModel.SetContent(rendered)

	title := lipgloss.NewStyle().Foreground(theme.BlueColor).Bold(true).Render("Mission Manifest")
	return theme.PanelBorder.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			title,
			viewportModel.View(),
		),
	)
}

func renderStandardAnalysisPanel(config PlanReviewConfig, width int) string {
	topHeight := max(5, planReviewAnalysisHeight/2)
	bottomHeight := max(5, planReviewAnalysisHeight-topHeight)

	coverage := renderCoverageMatrixPanel(config.Coverage, width, topHeight)
	dependencies := renderDependencyGraphPanel(config.Dependencies, width, bottomHeight)
	return lipgloss.JoinVertical(lipgloss.Left, coverage, dependencies)
}

func renderCompactAnalysisPanel(config PlanReviewConfig, width int) string {
	tab := normalizeAnalysisTab(config.AnalysisTab)
	coverageActive := tab == PlanReviewAnalysisCoverage
	tabStyle := lipgloss.NewStyle().Foreground(theme.LightGrayColor)

	coverageLabel := "[1] Coverage"
	dependenciesLabel := "[2] Dependencies"
	if coverageActive {
		coverageLabel = lipgloss.NewStyle().Foreground(theme.ButterscotchColor).Bold(true).Render(coverageLabel)
	} else {
		dependenciesLabel = lipgloss.NewStyle().Foreground(theme.ButterscotchColor).Bold(true).Render(dependenciesLabel)
	}
	tabLine := tabStyle.Render(fmt.Sprintf("%s  %s", coverageLabel, dependenciesLabel))

	body := ""
	if coverageActive {
		body = renderCoverageMatrixPanel(config.Coverage, width, 6)
	} else {
		body = renderDependencyGraphPanel(config.Dependencies, width, 6)
	}
	return lipgloss.JoinVertical(lipgloss.Left, tabLine, body)
}

func renderCoverageMatrixPanel(rows []PlanReviewCoverageRow, width int, height int) string {
	columns := []table.Column{
		{Title: "Use Case", Width: max(10, (width-8)/3)},
		{Title: "Missions", Width: max(16, (width-8)/2)},
		{Title: "Status", Width: max(10, (width-8)/6)},
	}

	tableRows := make([]table.Row, 0, len(rows))
	for _, row := range rows {
		useCase := strings.TrimSpace(row.UseCaseID)
		if useCase == "" {
			useCase = "UC-?"
		}
		missions := strings.Join(normalizeNonEmpty(row.MissionIDs), ", ")
		if missions == "" {
			missions = "-"
		}
		icon, label := coverageBadge(row.Status)
		tableRows = append(tableRows, table.Row{useCase, missions, icon + " " + label})
	}

	matrix := table.New(
		table.WithColumns(columns),
		table.WithRows(tableRows),
		table.WithHeight(max(3, height)),
		table.WithFocused(false),
	)
	styles := table.DefaultStyles()
	styles.Header = styles.Header.Foreground(theme.BlueColor).Bold(true)
	styles.Cell = styles.Cell.Foreground(theme.SpaceWhiteColor)
	matrix.SetStyles(styles)

	title := lipgloss.NewStyle().Foreground(theme.BlueColor).Bold(true).Render("Coverage Matrix")
	return theme.PanelBorder.Render(lipgloss.JoinVertical(lipgloss.Left, title, matrix.View()))
}

func renderDependencyGraphPanel(waves []PlanReviewDependencyWave, width int, height int) string {
	lines := renderDependencyLines(waves)
	viewportModel := viewport.New(max(20, width-4), max(4, height))
	viewportModel.SetContent(strings.Join(lines, "\n"))

	title := lipgloss.NewStyle().Foreground(theme.BlueColor).Bold(true).Render("Dependency Graph")
	return theme.PanelBorder.Render(lipgloss.JoinVertical(lipgloss.Left, title, viewportModel.View()))
}

func renderDependencyLines(waves []PlanReviewDependencyWave) []string {
	if len(waves) == 0 {
		return []string{"No dependencies mapped."}
	}

	ordered := append([]PlanReviewDependencyWave(nil), waves...)
	sort.SliceStable(ordered, func(i, j int) bool {
		return ordered[i].Wave < ordered[j].Wave
	})

	lines := make([]string, 0, 32)
	for _, wave := range ordered {
		lines = append(lines, lipgloss.NewStyle().Foreground(theme.BlueColor).Bold(true).Render(fmt.Sprintf("Wave %d", wave.Wave)))
		for _, mission := range wave.Missions {
			status := strings.TrimSpace(mission.Status)
			if status == "" {
				status = "waiting"
			}
			missionLine := fmt.Sprintf("├─ %s %s", strings.TrimSpace(mission.ID), strings.TrimSpace(mission.Title))
			lines = append(lines, lipgloss.NewStyle().Foreground(theme.ButterscotchColor).Render(missionLine)+" "+components.RenderStatusBadge(status, components.WithBadgeIcon(false)))
			for _, dep := range normalizeNonEmpty(mission.Dependencies) {
				lines = append(lines, lipgloss.NewStyle().Foreground(theme.GalaxyGrayColor).Render("│  └─ requires "+dep))
			}
		}
	}
	return lines
}

func renderFeedbackInput(value string, width int) string {
	current := value
	text := huh.NewText().
		Title("Admiral Feedback").
		Placeholder("Provide review feedback for the crew").
		Lines(3).
		Value(&current)
	form := huh.NewForm(huh.NewGroup(text))
	title := lipgloss.NewStyle().Foreground(theme.BlueColor).Bold(true).Render("Admiral Feedback")

	return theme.PanelBorderFocused.
		Width(max(20, width)).
		Render(lipgloss.JoinVertical(lipgloss.Left, title, form.View()))
}

func buildManifestMarkdown(missions []PlanReviewMission) string {
	if len(missions) == 0 {
		return "No missions in manifest."
	}

	entries := make([]string, 0, len(missions))
	for _, mission := range missions {
		id := strings.TrimSpace(mission.ID)
		if id == "" {
			id = "M-?"
		}
		title := strings.TrimSpace(mission.Title)
		if title == "" {
			title = "Untitled mission"
		}
		classification := strings.TrimSpace(mission.Classification)
		if classification == "" {
			classification = "UNCLASSIFIED"
		}
		useCaseText := strings.Join(normalizeNonEmpty(mission.UseCaseRefs), ", ")
		if useCaseText == "" {
			useCaseText = "-"
		}
		surface := strings.TrimSpace(mission.SurfaceArea)
		if surface == "" {
			surface = "-"
		}

		entries = append(entries, strings.Join([]string{
			fmt.Sprintf("### %s %s", id, title),
			fmt.Sprintf("- Classification: %s", classification),
			fmt.Sprintf("- Wave: %d", max(0, mission.Wave)),
			fmt.Sprintf("- Use Cases: %s", useCaseText),
			fmt.Sprintf("- AC Count: %d", max(0, mission.ACTotal)),
			fmt.Sprintf("- Surface Area: %s", surface),
		}, "\n"))
	}

	return strings.Join(entries, "\n\n---\n\n")
}

func renderMarkdown(markdown string, width int) string {
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(max(40, width)),
	)
	if err != nil {
		return markdown
	}
	rendered, err := renderer.Render(markdown)
	if err != nil {
		return markdown
	}
	return rendered
}

func coverageBadge(status PlanReviewCoverageStatus) (string, string) {
	switch normalizeCoverageStatus(status) {
	case PlanReviewCoverageCovered:
		return theme.IconDone, "covered"
	case PlanReviewCoveragePartial:
		return theme.IconAlert, "partial"
	default:
		return theme.IconFailed, "uncovered"
	}
}

func normalizeCoverageStatus(status PlanReviewCoverageStatus) PlanReviewCoverageStatus {
	switch PlanReviewCoverageStatus(strings.ToLower(strings.TrimSpace(string(status)))) {
	case PlanReviewCoverageCovered:
		return PlanReviewCoverageCovered
	case PlanReviewCoveragePartial:
		return PlanReviewCoveragePartial
	default:
		return PlanReviewCoverageUncovered
	}
}

func normalizeAnalysisTab(tab PlanReviewAnalysisTab) PlanReviewAnalysisTab {
	if strings.EqualFold(strings.TrimSpace(string(tab)), string(PlanReviewAnalysisDependencies)) {
		return PlanReviewAnalysisDependencies
	}
	return PlanReviewAnalysisCoverage
}

func countWaves(config PlanReviewConfig) int {
	maxWave := 0
	for _, mission := range config.Missions {
		if mission.Wave > maxWave {
			maxWave = mission.Wave
		}
	}
	for _, wave := range config.Dependencies {
		if wave.Wave > maxWave {
			maxWave = wave.Wave
		}
	}
	if maxWave == 0 && len(config.Missions) > 0 {
		return 1
	}
	return maxWave
}

func coveragePercent(rows []PlanReviewCoverageRow) int {
	if len(rows) == 0 {
		return 0
	}
	score := 0.0
	for _, row := range rows {
		switch normalizeCoverageStatus(row.Status) {
		case PlanReviewCoverageCovered:
			score += 1.0
		case PlanReviewCoveragePartial:
			score += 0.5
		}
	}
	return int((score / float64(len(rows))) * 100)
}

func normalizeNonEmpty(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		out = append(out, trimmed)
	}
	return out
}

func max(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

func clampNonNegative(value int) int {
	if value < 0 {
		return 0
	}
	return value
}
