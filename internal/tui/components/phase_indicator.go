package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/ship-commander/sc3/internal/tui/theme"
)

type phaseLabel struct {
	key     string
	full    string
	compact string
}

var phaseLabels = []phaseLabel{
	{key: "red", full: "RED", compact: "R"},
	{key: "verify_red", full: "VERIFY_RED", compact: "VR"},
	{key: "green", full: "GREEN", compact: "G"},
	{key: "verify_green", full: "VERIFY_GREEN", compact: "VG"},
	{key: "refactor", full: "REFACTOR", compact: "RF"},
	{key: "verify_refactor", full: "VERIFY_REFACTOR", compact: "VRF"},
}

// RenderPhaseIndicator renders the 6-step TDD phase cycle in full or compact form.
func RenderPhaseIndicator(currentPhase string, phasesCompleted []string, compact bool) string {
	current := normalizePhase(currentPhase)
	completed := make(map[string]struct{}, len(phasesCompleted))
	for _, phase := range phasesCompleted {
		completed[normalizePhase(phase)] = struct{}{}
	}

	parts := make([]string, 0, len(phaseLabels)*2-1)
	for i, phase := range phaseLabels {
		label := phase.full
		if compact {
			label = phase.compact
		}

		_, isCompleted := completed[phase.key]
		parts = append(parts, phaseLabelStyle(current == phase.key, isCompleted).Render(label))

		if i < len(phaseLabels)-1 {
			parts = append(parts, phaseSeparatorStyle().Render(phaseSeparator(compact)))
		}
	}

	return strings.Join(parts, "")
}

func phaseLabelStyle(isCurrent bool, isCompleted bool) lipgloss.Style {
	switch {
	case isCurrent:
		return lipgloss.NewStyle().
			Foreground(theme.ButterscotchColor).
			Bold(true)
	case isCompleted:
		return lipgloss.NewStyle().
			Foreground(theme.GreenOkColor).
			Faint(true)
	default:
		return lipgloss.NewStyle().
			Foreground(theme.GalaxyGrayColor).
			Faint(true)
	}
}

func phaseSeparator(compact bool) string {
	if compact {
		return " > "
	}
	return " -> "
}

func phaseSeparatorStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.SpaceWhiteColor).
		Faint(true)
}

func normalizePhase(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	normalized = strings.ReplaceAll(normalized, "-", "_")
	normalized = strings.ReplaceAll(normalized, " ", "_")
	return normalized
}
