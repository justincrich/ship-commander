package components

import (
	"fmt"
	"strings"
	"testing"

	"github.com/ship-commander/sc3/internal/tui/theme"
)

func TestRenderShipStatusRowFormat(t *testing.T) {
	t.Parallel()

	row := ShipStatusRow{
		ShipName:       "USS Discovery",
		DirectiveTitle: "Auth System Stabilization",
		Status:         "launched",
		CrewCount:      4,
		WaveCurrent:    2,
		WaveTotal:      3,
		MissionsDone:   8,
		MissionsTotal:  12,
		Health:         80,
	}

	rendered := RenderShipStatusRow(row)
	assertContainsAll(t, rendered, "▸", "USS Discovery", "Crew:4", "Wave 2/3", "[#####...]", "Missions:8/12", "Health:****o")
}

func TestResolveShipStatusVariant(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		status       string
		hasStuck     bool
		wantIcon     string
		wantColorKey string
	}{
		{
			name:         "launched default",
			status:       "launched",
			hasStuck:     false,
			wantIcon:     theme.IconRunning,
			wantColorKey: fmt.Sprint(theme.ButterscotchColor),
		},
		{
			name:         "complete",
			status:       "complete",
			hasStuck:     false,
			wantIcon:     theme.IconDone,
			wantColorKey: fmt.Sprint(theme.GreenOkColor),
		},
		{
			name:         "issues by status",
			status:       "halted",
			hasStuck:     false,
			wantIcon:     theme.IconAlert,
			wantColorKey: fmt.Sprint(theme.YellowCautionColor),
		},
		{
			name:         "issues by stuck flag",
			status:       "launched",
			hasStuck:     true,
			wantIcon:     theme.IconAlert,
			wantColorKey: fmt.Sprint(theme.YellowCautionColor),
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			variant := resolveShipStatusVariant(testCase.status, testCase.hasStuck)
			if variant.icon != testCase.wantIcon {
				t.Fatalf("icon = %q, want %q", variant.icon, testCase.wantIcon)
			}
			if fmt.Sprint(variant.color) != testCase.wantColorKey {
				t.Fatalf("color = %v, want %v", variant.color, testCase.wantColorKey)
			}
		})
	}
}

func TestRenderShipStatusRowPendingAndSelected(t *testing.T) {
	t.Parallel()

	row := ShipStatusRow{
		ShipName:         "USS Enterprise",
		DirectiveTitle:   "Payment API",
		Status:           "halted",
		CrewCount:        3,
		WaveCurrent:      1,
		WaveTotal:        2,
		MissionsDone:     3,
		MissionsTotal:    8,
		Health:           45,
		PendingQuestions: 2,
		Selected:         true,
	}

	rendered := RenderShipStatusRow(row)
	assertContainsAll(t, rendered, "Q:2", "⚠", "Health:**ooo")

	if selectedRowStyle(true).GetBackground() == nil {
		t.Fatal("expected selected row style to define a background color")
	}
}

func TestRenderShipStatusRowDirectiveFallbackAndTruncation(t *testing.T) {
	t.Parallel()

	noDirective := RenderShipStatusRow(ShipStatusRow{
		ShipName:      "USS Defiant",
		Status:        "launched",
		CrewCount:     2,
		WaveCurrent:   1,
		WaveTotal:     1,
		MissionsDone:  1,
		MissionsTotal: 2,
		Health:        75,
	})

	if !strings.Contains(noDirective, "No Directive") {
		t.Fatalf("expected fallback directive text, got %q", noDirective)
	}

	truncated := RenderShipStatusRow(ShipStatusRow{
		ShipName:       "USS Defiant",
		DirectiveTitle: "This directive title is intentionally long so width logic trims it",
		Status:         "launched",
		CrewCount:      2,
		WaveCurrent:    1,
		WaveTotal:      1,
		MissionsDone:   1,
		MissionsTotal:  2,
		Health:         75,
		Width:          70,
	})

	if !strings.Contains(truncated, "...") {
		t.Fatalf("expected truncated directive with ellipsis, got %q", truncated)
	}
}

func TestRenderWaveBarPlainZeroTotal(t *testing.T) {
	t.Parallel()

	bar := renderWaveBarPlain(2, 0)
	if bar != "[........]" {
		t.Fatalf("zero-total bar = %q, want %q", bar, "[........]")
	}
}

func assertContainsAll(t *testing.T, value string, fragments ...string) {
	t.Helper()

	for _, fragment := range fragments {
		if !strings.Contains(value, fragment) {
			t.Fatalf("expected %q to contain %q", value, fragment)
		}
	}
}
