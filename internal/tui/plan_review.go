package tui

import (
	"fmt"
	"strings"

	"github.com/ship-commander/sc3/internal/admiral"
)

const (
	// BadgeColorRedAlert is the semantic color token for RED_ALERT missions.
	BadgeColorRedAlert = "red_alert"
	// BadgeColorGreenOK is the semantic color token for STANDARD_OPS missions.
	BadgeColorGreenOK = "green_ok"
	// BadgeColorWarning is the semantic color token for warning callouts.
	BadgeColorWarning = "warning"
)

// ClassificationBadge returns a plan-review badge label and semantic color token.
func ClassificationBadge(classification string) (label string, color string) {
	switch strings.ToUpper(strings.TrimSpace(classification)) {
	case "RED_ALERT":
		return "[RED_ALERT] 🔴", BadgeColorRedAlert
	case "STANDARD_OPS":
		return "[STANDARD_OPS] 🟢", BadgeColorGreenOK
	default:
		return "[UNCLASSIFIED] ⚪", BadgeColorWarning
	}
}

// RenderPlanReviewMission renders one mission row for Plan Review with optional expanded rationale details.
func RenderPlanReviewMission(mission admiral.Mission, expanded bool) string {
	badge, badgeColor := ClassificationBadge(mission.Classification)
	confidence := strings.ToLower(strings.TrimSpace(mission.ClassificationConfidence))
	if confidence == "" {
		confidence = "unknown"
	}

	lines := []string{
		fmt.Sprintf("%s: %s", strings.TrimSpace(mission.ID), strings.TrimSpace(mission.Title)),
		fmt.Sprintf("  Classification: %s (%s)", badge, badgeColor),
		fmt.Sprintf("  Confidence: %s", confidence),
	}

	if mission.ClassificationNeedsReview {
		lines = append(lines, fmt.Sprintf("  Warning: low-confidence classification (%s)", BadgeColorWarning))
	}

	if expanded {
		criteria := strings.Join(mission.ClassificationCriteria, ", ")
		if strings.TrimSpace(criteria) == "" {
			criteria = "(none)"
		}
		rationale := strings.TrimSpace(mission.ClassificationRationale)
		if rationale == "" {
			rationale = "(none)"
		}

		lines = append(lines, fmt.Sprintf("  Criteria: %s", criteria))
		lines = append(lines, fmt.Sprintf("  Rationale: %s", rationale))
	}

	return strings.Join(lines, "\n")
}
