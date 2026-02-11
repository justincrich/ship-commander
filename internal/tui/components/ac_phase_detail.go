package components

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ship-commander/sc3/internal/tui/theme"
)

const (
	defaultACPhaseDetailWidth  = 72
	defaultACPhaseDetailHeight = 14
)

var (
	acPhaseOrder = []string{
		"RED",
		"VERIFY_RED",
		"GREEN",
		"VERIFY_GREEN",
		"REFACTOR",
		"VERIFY_REFACTOR",
	}
	acPhaseAbbreviations = map[string]string{
		"RED":             "R",
		"VERIFY_RED":      "VR",
		"GREEN":           "G",
		"VERIFY_GREEN":    "VG",
		"REFACTOR":        "RF",
		"VERIFY_REFACTOR": "VRF",
	}
)

// ACGateResult captures one gate execution result associated to an AC phase.
type ACGateResult struct {
	Phase          string
	GateType       string
	ExitCode       int
	Classification string
	Message        string
}

// ACPhaseData is the render data for one acceptance-criterion row.
type ACPhaseData struct {
	ACIndex          int
	ACTitle          string
	CurrentPhase     string
	PhasesCompleted  []string
	AttemptCount     int
	GateResults      []ACGateResult
	ExpandedGateRows bool
}

// ACPhaseDetailConfig defines rendering inputs for ACPhaseDetail.
type ACPhaseDetailConfig struct {
	AcceptanceCriteria []ACPhaseData
	SelectedIndex      int
	Height             int
	Width              int
	Compact            bool
	PulseFrame         bool
}

type acPhaseListItem struct {
	data      ACPhaseData
	compact   bool
	pulse     bool
	phaseView []string
}

func (item acPhaseListItem) FilterValue() string {
	return item.data.ACTitle
}

type acPhaseItemDelegate struct{}

func (delegate acPhaseItemDelegate) Height() int {
	return 2
}

func (delegate acPhaseItemDelegate) Spacing() int {
	return 0
}

func (delegate acPhaseItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd {
	return nil
}

func (delegate acPhaseItemDelegate) Render(writer io.Writer, model list.Model, index int, item list.Item) {
	acItem, ok := item.(acPhaseListItem)
	if !ok {
		return
	}

	selected := index == model.Index()

	prefix := "  "
	if selected {
		prefix = lipgloss.NewStyle().
			Foreground(theme.MoonlitVioletColor).
			Bold(true).
			Render(theme.IconRunning + " ")
	}

	titleLine := prefix +
		lipgloss.NewStyle().Foreground(theme.BlueColor).Bold(true).Render(fmt.Sprintf("AC-%d", acItem.data.ACIndex)) +
		" " +
		lipgloss.NewStyle().Foreground(theme.SpaceWhiteColor).Render(strings.TrimSpace(acItem.data.ACTitle))

	pipelineLine := "   " + strings.Join(acItem.phaseView, "  ")
	if acItem.data.AttemptCount > 1 {
		pipelineLine += "  " + lipgloss.NewStyle().
			Foreground(theme.YellowCautionColor).
			Bold(true).
			Render(fmt.Sprintf("Attempt: %d", acItem.data.AttemptCount))
	}

	rendered := titleLine + "\n" + pipelineLine
	if selected {
		rendered = lipgloss.NewStyle().
			Foreground(theme.SpaceWhiteColor).
			Render(rendered)
	}

	if _, err := io.WriteString(writer, rendered); err != nil {
		return
	}
}

// RenderACPhaseDetail renders a scrollable AC list with phase pipeline and selected gate detail.
func RenderACPhaseDetail(config ACPhaseDetailConfig) string {
	normalized := normalizeACPhaseDetailConfig(config)
	model := newACPhaseListModel(normalized)
	view := model.View()

	selectedAC, ok := selectedACPhaseData(normalized.AcceptanceCriteria, normalized.SelectedIndex)
	if !ok {
		return view
	}

	detail := renderSelectedACGateDetail(selectedAC, normalized.Width)
	if strings.TrimSpace(detail) == "" {
		return view
	}

	return lipgloss.JoinVertical(lipgloss.Left, view, detail)
}

func normalizeACPhaseDetailConfig(config ACPhaseDetailConfig) ACPhaseDetailConfig {
	width := config.Width
	if width <= 0 {
		width = defaultACPhaseDetailWidth
	}

	height := config.Height
	if height <= 0 {
		height = defaultACPhaseDetailHeight
	}

	selected := config.SelectedIndex
	if selected < 0 {
		selected = 0
	}

	if len(config.AcceptanceCriteria) > 0 && selected >= len(config.AcceptanceCriteria) {
		selected = len(config.AcceptanceCriteria) - 1
	}

	return ACPhaseDetailConfig{
		AcceptanceCriteria: config.AcceptanceCriteria,
		SelectedIndex:      selected,
		Height:             height,
		Width:              width,
		Compact:            config.Compact,
		PulseFrame:         config.PulseFrame,
	}
}

func newACPhaseListModel(config ACPhaseDetailConfig) list.Model {
	items := make([]list.Item, 0, len(config.AcceptanceCriteria))
	for _, ac := range config.AcceptanceCriteria {
		items = append(items, acPhaseListItem{
			data:      ac,
			compact:   config.Compact,
			pulse:     config.PulseFrame,
			phaseView: renderACPhasePipeline(ac, config.Compact, config.PulseFrame),
		})
	}

	model := list.New(items, acPhaseItemDelegate{}, config.Width, config.Height)
	model.SetShowTitle(false)
	model.SetShowStatusBar(false)
	model.SetShowHelp(false)
	model.SetShowPagination(false)
	model.SetFilteringEnabled(false)
	model.Styles.NoItems = lipgloss.NewStyle().Foreground(theme.GalaxyGrayColor)

	if len(items) > 0 {
		model.Select(config.SelectedIndex)
	}

	return model
}

func renderACPhasePipeline(ac ACPhaseData, compact bool, pulseFrame bool) []string {
	completed := make(map[string]bool, len(ac.PhasesCompleted))
	for _, phase := range ac.PhasesCompleted {
		completed[normalizePhaseName(phase)] = true
	}

	failed := phaseFailuresByName(ac.GateResults)
	current := normalizePhaseName(ac.CurrentPhase)

	segments := make([]string, 0, len(acPhaseOrder))
	for _, phase := range acPhaseOrder {
		label := phase
		if compact {
			if abbreviation, ok := acPhaseAbbreviations[phase]; ok {
				label = abbreviation
			}
		}

		symbol := " "
		style := lipgloss.NewStyle().Foreground(theme.GalaxyGrayColor).Faint(true)

		switch {
		case failed[phase]:
			symbol = "x"
			style = lipgloss.NewStyle().Foreground(theme.RedAlertColor).Bold(true)
		case current == phase:
			symbol = ">"
			style = lipgloss.NewStyle().Foreground(theme.ButterscotchColor).Bold(true)
			if pulseFrame {
				style = style.Underline(true)
			}
		case completed[phase]:
			symbol = "*"
			style = lipgloss.NewStyle().Foreground(theme.GreenOkColor).Faint(true)
		}

		segments = append(segments, style.Render(fmt.Sprintf("[%s] %s", symbol, label)))
	}

	return segments
}

func phaseFailuresByName(results []ACGateResult) map[string]bool {
	failed := make(map[string]bool, len(results))
	for _, result := range results {
		phase := normalizePhaseName(result.Phase)
		if phase == "" {
			phase = phaseFromGateType(result.GateType)
		}
		if phase == "" {
			continue
		}

		failed[phase] = gateResultIsFailure(result)
	}
	return failed
}

func gateResultIsFailure(result ACGateResult) bool {
	if result.ExitCode != 0 {
		return true
	}
	classification := strings.ToLower(strings.TrimSpace(result.Classification))
	return strings.Contains(classification, "fail") || strings.Contains(classification, "reject")
}

func phaseFromGateType(gateType string) string {
	switch normalizePhaseName(gateType) {
	case "VERIFY_RED":
		return "VERIFY_RED"
	case "VERIFY_GREEN":
		return "VERIFY_GREEN"
	case "VERIFY_REFACTOR":
		return "VERIFY_REFACTOR"
	case "RED":
		return "RED"
	case "GREEN":
		return "GREEN"
	case "REFACTOR":
		return "REFACTOR"
	default:
		return ""
	}
}

func normalizePhaseName(phase string) string {
	normalized := strings.ToUpper(strings.TrimSpace(phase))
	normalized = strings.ReplaceAll(normalized, "-", "_")
	normalized = strings.ReplaceAll(normalized, " ", "_")

	switch normalized {
	case "VERIFYRED":
		return "VERIFY_RED"
	case "VERIFYGREEN":
		return "VERIFY_GREEN"
	case "VERIFYREFACTOR":
		return "VERIFY_REFACTOR"
	default:
		return normalized
	}
}

func selectedACPhaseData(acceptanceCriteria []ACPhaseData, selectedIndex int) (ACPhaseData, bool) {
	if selectedIndex < 0 || selectedIndex >= len(acceptanceCriteria) {
		return ACPhaseData{}, false
	}
	return acceptanceCriteria[selectedIndex], true
}

func renderSelectedACGateDetail(ac ACPhaseData, width int) string {
	header := lipgloss.NewStyle().
		Foreground(theme.MoonlitVioletColor).
		Bold(true).
		Render("Gate detail (selected AC)")

	lines := []string{header}
	if len(ac.GateResults) == 0 {
		lines = append(lines, lipgloss.NewStyle().Foreground(theme.GalaxyGrayColor).Faint(true).Render("No gate results yet."))
	} else {
		for _, gateResult := range ac.GateResults {
			phase := normalizePhaseName(gateResult.Phase)
			if phase == "" {
				phase = phaseFromGateType(gateResult.GateType)
			}
			if phase == "" {
				phase = "UNKNOWN"
			}

			line := fmt.Sprintf("%s | %s | exit %d | %s", phase, strings.TrimSpace(gateResult.GateType), gateResult.ExitCode, strings.TrimSpace(gateResult.Classification))
			if message := strings.TrimSpace(gateResult.Message); message != "" {
				line += " | " + message
			}

			lineStyle := lipgloss.NewStyle().Foreground(theme.GreenOkColor)
			if gateResultIsFailure(gateResult) {
				lineStyle = lipgloss.NewStyle().Foreground(theme.RedAlertColor)
			}
			lines = append(lines, lineStyle.Render(line))
		}
	}

	detailStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), true, false, false, false).
		BorderForeground(theme.MoonlitVioletColor).
		Padding(0, 1)
	if width > 0 {
		detailStyle = detailStyle.Width(width)
	}

	return detailStyle.Render(strings.Join(lines, "\n"))
}
