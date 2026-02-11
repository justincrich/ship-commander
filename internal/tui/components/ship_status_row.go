package components

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/ship-commander/sc3/internal/tui/theme"
)

const (
	shipHealthSlots  = 5
	shipWaveBarSlots = 8
)

type shipStatusVariant struct {
	icon  string
	color lipgloss.TerminalColor
}

// ShipStatusRow defines the display payload for a Fleet Monitor row.
type ShipStatusRow struct {
	ShipName         string
	DirectiveTitle   string
	Status           string
	CrewCount        int
	WaveCurrent      int
	WaveTotal        int
	MissionsDone     int
	MissionsTotal    int
	Health           int
	HasStuck         bool
	PendingQuestions int
	Selected         bool
	Width            int
}

// RenderShipStatusRow renders a compact single-line ship summary for Fleet Monitor.
func RenderShipStatusRow(row ShipStatusRow) string {
	variant := resolveShipStatusVariant(row.Status, row.HasStuck)

	shipName := strings.TrimSpace(row.ShipName)
	if shipName == "" {
		shipName = "Unnamed Ship"
	}

	hasDirective := strings.TrimSpace(row.DirectiveTitle) != ""
	directive := strings.TrimSpace(row.DirectiveTitle)
	if !hasDirective {
		directive = "No Directive"
	}

	crewText := fmt.Sprintf("Crew:%d", clampInt(row.CrewCount, 0, row.CrewCount))
	waveText := fmt.Sprintf("Wave %d/%d", clampInt(row.WaveCurrent, 0, row.WaveCurrent), clampInt(row.WaveTotal, 0, row.WaveTotal))
	waveBarPlain := renderWaveBarPlain(row.MissionsDone, row.MissionsTotal)
	missionText := fmt.Sprintf("Missions:%d/%d", clampInt(row.MissionsDone, 0, row.MissionsDone), clampInt(row.MissionsTotal, 0, row.MissionsTotal))

	healthDots, healthColor := renderHealthDots(row.Health)
	healthText := "Health:" + healthDots

	if row.Width > 0 {
		directive = fitDirectiveToWidth(row.Width, variant.icon, shipName, directive, crewText, waveText, waveBarPlain, missionText, healthText, row.PendingQuestions)
	}

	statusRendered := lipgloss.NewStyle().Foreground(variant.color).Bold(true).Render(variant.icon)
	shipRendered := lipgloss.NewStyle().Foreground(theme.ButterscotchColor).Bold(true).Render(shipName)

	directiveColor := theme.BlueColor
	if !hasDirective {
		directiveColor = theme.GalaxyGrayColor
	}
	directiveRendered := lipgloss.NewStyle().Foreground(directiveColor).Render(directive)

	metricsStyle := lipgloss.NewStyle().Foreground(theme.LightGrayColor)
	waveBarRendered := lipgloss.NewStyle().Foreground(variant.color).Render(waveBarPlain)
	healthRendered := lipgloss.NewStyle().Foreground(healthColor).Render(healthText)

	parts := []string{
		statusRendered,
		shipRendered,
		directiveRendered,
		metricsStyle.Render(crewText),
		metricsStyle.Render(waveText),
		waveBarRendered,
		metricsStyle.Render(missionText),
		healthRendered,
	}

	if row.PendingQuestions > 0 {
		parts = append(parts, lipgloss.NewStyle().Foreground(theme.PinkColor).Bold(true).Render(fmt.Sprintf("Q:%d", row.PendingQuestions)))
	}

	rendered := strings.Join(parts, " ")
	return selectedRowStyle(row.Selected).Render(rendered)
}

func resolveShipStatusVariant(status string, hasStuck bool) shipStatusVariant {
	if hasStuck {
		return shipStatusVariant{
			icon:  theme.IconAlert,
			color: theme.YellowCautionColor,
		}
	}

	switch strings.ToLower(strings.TrimSpace(status)) {
	case "complete", "completed", "done":
		return shipStatusVariant{
			icon:  theme.IconDone,
			color: theme.GreenOkColor,
		}
	case "halted", "failed", "stuck", "has_issues":
		return shipStatusVariant{
			icon:  theme.IconAlert,
			color: theme.YellowCautionColor,
		}
	default:
		return shipStatusVariant{
			icon:  theme.IconRunning,
			color: theme.ButterscotchColor,
		}
	}
}

func renderWaveBarPlain(done int, total int) string {
	safeDone := clampInt(done, 0, done)
	safeTotal := clampInt(total, 0, total)
	if safeTotal == 0 {
		return "[" + strings.Repeat(".", shipWaveBarSlots) + "]"
	}

	ratio := float64(safeDone) / float64(safeTotal)
	ratio = math.Max(0, math.Min(1, ratio))
	filled := int(math.Round(ratio * shipWaveBarSlots))
	if filled > shipWaveBarSlots {
		filled = shipWaveBarSlots
	}

	return "[" + strings.Repeat("#", filled) + strings.Repeat(".", shipWaveBarSlots-filled) + "]"
}

func renderHealthDots(health int) (string, lipgloss.TerminalColor) {
	safeHealth := clampInt(health, 0, 100)
	filled := int(math.Round(float64(safeHealth) / float64(100) * shipHealthSlots))
	filled = clampInt(filled, 0, shipHealthSlots)

	dots := strings.Repeat("*", filled) + strings.Repeat("o", shipHealthSlots-filled)
	switch {
	case safeHealth >= 80:
		return dots, theme.GreenOkColor
	case safeHealth >= 50:
		return dots, theme.YellowCautionColor
	default:
		return dots, theme.RedAlertColor
	}
}

func selectedRowStyle(selected bool) lipgloss.Style {
	if !selected {
		return lipgloss.NewStyle()
	}

	return lipgloss.NewStyle().
		Background(theme.MoonlitVioletColor).
		Foreground(theme.SpaceWhiteColor)
}

func fitDirectiveToWidth(
	maxWidth int,
	icon string,
	shipName string,
	directive string,
	crewText string,
	waveText string,
	waveBar string,
	missionText string,
	healthText string,
	pendingQuestions int,
) string {
	pendingText := ""
	if pendingQuestions > 0 {
		pendingText = fmt.Sprintf("Q:%d", pendingQuestions)
	}

	buildPlain := func(currentDirective string) string {
		parts := []string{
			icon,
			shipName,
			currentDirective,
			crewText,
			waveText,
			waveBar,
			missionText,
			healthText,
		}
		if pendingText != "" {
			parts = append(parts, pendingText)
		}
		return strings.Join(parts, " ")
	}

	plain := buildPlain(directive)
	if len([]rune(plain)) <= maxWidth {
		return directive
	}

	overflow := len([]rune(plain)) - maxWidth
	target := len([]rune(directive)) - overflow
	if target < 0 {
		target = 0
	}

	return truncateText(directive, target)
}

func truncateText(value string, maxRunes int) string {
	runes := []rune(value)
	if len(runes) <= maxRunes {
		return value
	}
	if maxRunes <= 0 {
		return ""
	}
	if maxRunes <= 3 {
		return string(runes[:maxRunes])
	}
	return string(runes[:maxRunes-3]) + "..."
}

func clampInt(value int, min int, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
