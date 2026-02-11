package views

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestRenderPlanReviewIncludesManifestCoverageDependencyAndToolbar(t *testing.T) {
	t.Parallel()

	rendered := RenderPlanReview(samplePlanReviewConfig(128))
	for _, expected := range []string{
		"PLAN REVIEW -- USS Enterprise",
		"Directive: Validate launch manifest",
		"Missions: 3",
		"Coverage Matrix",
		"Dependency Graph",
		"UC-TUI-01",
		"UC-TUI-03",
		"Wave 1",
		"requires M-001",
		"[a]",
		"Approve",
		"[Esc]",
		"Ready Room",
	} {
		if !strings.Contains(rendered, expected) {
			t.Fatalf("plan review missing %q\n%s", expected, rendered)
		}
	}
}

func TestRenderPlanReviewCompactUsesTabbedAnalysis(t *testing.T) {
	t.Parallel()

	config := samplePlanReviewConfig(80)
	config.AnalysisTab = PlanReviewAnalysisDependencies
	rendered := RenderPlanReview(config)

	for _, expected := range []string{
		"[1] Coverage",
		"[2] Dependencies",
		"Dependency Graph",
		"Wave 1",
	} {
		if !strings.Contains(rendered, expected) {
			t.Fatalf("compact plan review missing %q\n%s", expected, rendered)
		}
	}
}

func TestRenderPlanReviewFeedbackModeShowsInlineInput(t *testing.T) {
	t.Parallel()

	config := samplePlanReviewConfig(120)
	config.FeedbackMode = true
	config.FeedbackText = "Need stronger coverage for UC-TUI-03."

	rendered := RenderPlanReview(config)
	if !strings.Contains(rendered, "Admiral Feedback") {
		t.Fatalf("feedback mode should show inline input title\n%s", rendered)
	}
}

func TestResolvePlanReviewLayout(t *testing.T) {
	t.Parallel()

	if got := ResolvePlanReviewLayout(80); got != PlanReviewLayoutCompact {
		t.Fatalf("layout for width 80 = %q, want compact", got)
	}
	if got := ResolvePlanReviewLayout(120); got != PlanReviewLayoutStandard {
		t.Fatalf("layout for width 120 = %q, want standard", got)
	}
}

func TestPlanReviewQuickActionForKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		key  tea.KeyMsg
		want PlanReviewQuickAction
	}{
		{key: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}, want: PlanReviewQuickActionApprove},
		{key: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}}, want: PlanReviewQuickActionFeedback},
		{key: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}}, want: PlanReviewQuickActionShelve},
		{key: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}, want: PlanReviewQuickActionHelp},
		{key: tea.KeyMsg{Type: tea.KeyEsc}, want: PlanReviewQuickActionReadyRoom},
		{key: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}}, want: PlanReviewQuickActionCoverageTab},
		{key: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}}, want: PlanReviewQuickActionDependenciesTab},
		{key: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}, want: PlanReviewQuickActionNone},
	}

	for _, tt := range tests {
		if got := PlanReviewQuickActionForKey(tt.key); got != tt.want {
			t.Fatalf("quick action for key %q = %q, want %q", tt.key.String(), got, tt.want)
		}
	}
}

func samplePlanReviewConfig(width int) PlanReviewConfig {
	return PlanReviewConfig{
		Width:          width,
		ShipName:       "USS Enterprise",
		DirectiveTitle: "Validate launch manifest",
		Missions: []PlanReviewMission{
			{
				ID:             "M-001",
				Title:          "Initialize bridge systems",
				Classification: "STANDARD_OPS",
				Wave:           1,
				UseCaseRefs:    []string{"UC-TUI-01", "UC-TUI-03"},
				ACTotal:        4,
				SurfaceArea:    "internal/tui/views",
			},
			{
				ID:             "M-002",
				Title:          "Implement coverage matrix",
				Classification: "STANDARD_OPS",
				Wave:           1,
				UseCaseRefs:    []string{"UC-TUI-03"},
				ACTotal:        3,
				SurfaceArea:    "internal/tui/components",
			},
			{
				ID:             "M-003",
				Title:          "Review risk boundaries",
				Classification: "RED_ALERT",
				Wave:           2,
				UseCaseRefs:    []string{"UC-TUI-15"},
				ACTotal:        2,
				SurfaceArea:    "internal/commander",
			},
		},
		Coverage: []PlanReviewCoverageRow{
			{UseCaseID: "UC-TUI-01", MissionIDs: []string{"M-001"}, Status: PlanReviewCoverageCovered},
			{UseCaseID: "UC-TUI-03", MissionIDs: []string{"M-001", "M-002"}, Status: PlanReviewCoveragePartial},
			{UseCaseID: "UC-TUI-15", MissionIDs: nil, Status: PlanReviewCoverageUncovered},
		},
		Dependencies: []PlanReviewDependencyWave{
			{
				Wave: 1,
				Missions: []PlanReviewDependencyMission{
					{ID: "M-001", Title: "Initialize bridge systems", Status: "done"},
					{ID: "M-002", Title: "Implement coverage matrix", Status: "running", Dependencies: []string{"M-001"}},
				},
			},
			{
				Wave: 2,
				Missions: []PlanReviewDependencyMission{
					{ID: "M-003", Title: "Review risk boundaries", Status: "waiting", Dependencies: []string{"M-001", "M-002"}},
				},
			},
		},
		SignoffsDone:       2,
		SignoffsTotal:      3,
		AnalysisTab:        PlanReviewAnalysisCoverage,
		ToolbarHighlighted: 1,
	}
}
