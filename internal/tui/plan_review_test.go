package tui

import (
	"strings"
	"testing"

	"github.com/ship-commander/sc3/internal/admiral"
)

func TestClassificationBadge(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		classification string
		wantLabel      string
		wantColor      string
	}{
		{
			name:           "red alert badge",
			classification: "RED_ALERT",
			wantLabel:      "[RED_ALERT] 🔴",
			wantColor:      BadgeColorRedAlert,
		},
		{
			name:           "standard ops badge",
			classification: "STANDARD_OPS",
			wantLabel:      "[STANDARD_OPS] 🟢",
			wantColor:      BadgeColorGreenOK,
		},
		{
			name:           "unknown badge",
			classification: "",
			wantLabel:      "[UNCLASSIFIED] ⚪",
			wantColor:      BadgeColorWarning,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			label, color := ClassificationBadge(tt.classification)
			if label != tt.wantLabel {
				t.Fatalf("label = %q, want %q", label, tt.wantLabel)
			}
			if color != tt.wantColor {
				t.Fatalf("color = %q, want %q", color, tt.wantColor)
			}
		})
	}
}

func TestRenderPlanReviewMissionIncludesExpandedRationaleAndWarning(t *testing.T) {
	t.Parallel()

	mission := admiral.Mission{
		ID:                        "M-42",
		Title:                     "Classify mission risk",
		Classification:            "RED_ALERT",
		ClassificationCriteria:    []string{"business_logic", "bug_fix"},
		ClassificationRationale:   "Touches mission execution behavior.",
		ClassificationConfidence:  "low",
		ClassificationNeedsReview: true,
	}

	rendered := RenderPlanReviewMission(mission, true)

	for _, expected := range []string{
		"M-42: Classify mission risk",
		"Classification: [RED_ALERT] 🔴",
		"Confidence: low",
		"Warning: low-confidence classification",
		"Criteria: business_logic, bug_fix",
		"Rationale: Touches mission execution behavior.",
	} {
		if !strings.Contains(rendered, expected) {
			t.Fatalf("rendered output missing %q\n%s", expected, rendered)
		}
	}
}
