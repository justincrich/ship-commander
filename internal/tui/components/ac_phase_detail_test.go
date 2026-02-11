package components

import (
	"regexp"
	"strings"
	"testing"
)

var acPhaseANSIPattern = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func TestRenderACPhaseDetailPipelineStates(t *testing.T) {
	t.Parallel()

	rendered := stripACPhaseANSI(RenderACPhaseDetail(ACPhaseDetailConfig{
		AcceptanceCriteria: []ACPhaseData{
			{
				ACIndex:         1,
				ACTitle:         "User can log in",
				CurrentPhase:    "GREEN",
				PhasesCompleted: []string{"RED", "VERIFY_RED"},
				GateResults: []ACGateResult{
					{Phase: "VERIFY_GREEN", GateType: "VERIFY_GREEN", ExitCode: 1, Classification: "reject_failure"},
				},
			},
		},
		SelectedIndex: 0,
		Width:         100,
		Height:        8,
	}))

	if !strings.Contains(rendered, "[*] RED") {
		t.Fatalf("expected completed RED phase marker, got %q", rendered)
	}
	if !strings.Contains(rendered, "[>] GREEN") {
		t.Fatalf("expected active GREEN phase marker, got %q", rendered)
	}
	if !strings.Contains(rendered, "[x] VERIFY_GREEN") {
		t.Fatalf("expected failed VERIFY_GREEN phase marker, got %q", rendered)
	}
	if !strings.Contains(rendered, "[ ] REFACTOR") {
		t.Fatalf("expected pending REFACTOR phase marker, got %q", rendered)
	}
}

func TestRenderACPhaseDetailAttemptCountShownOnlyWhenOverOne(t *testing.T) {
	t.Parallel()

	withAttempts := stripACPhaseANSI(RenderACPhaseDetail(ACPhaseDetailConfig{
		AcceptanceCriteria: []ACPhaseData{
			{
				ACIndex:         1,
				ACTitle:         "Attempted AC",
				CurrentPhase:    "RED",
				PhasesCompleted: nil,
				AttemptCount:    3,
			},
		},
		SelectedIndex: 0,
		Width:         100,
		Height:        8,
	}))
	if !strings.Contains(withAttempts, "Attempt: 3") {
		t.Fatalf("expected attempt count when over 1, got %q", withAttempts)
	}

	withoutAttempts := stripACPhaseANSI(RenderACPhaseDetail(ACPhaseDetailConfig{
		AcceptanceCriteria: []ACPhaseData{
			{
				ACIndex:         1,
				ACTitle:         "First attempt AC",
				CurrentPhase:    "RED",
				PhasesCompleted: nil,
				AttemptCount:    1,
			},
		},
		SelectedIndex: 0,
		Width:         100,
		Height:        8,
	}))
	if strings.Contains(withoutAttempts, "Attempt: 1") {
		t.Fatalf("did not expect attempt count for 1, got %q", withoutAttempts)
	}
}

func TestRenderACPhaseDetailIncludesSelectedGateDetails(t *testing.T) {
	t.Parallel()

	rendered := stripACPhaseANSI(RenderACPhaseDetail(ACPhaseDetailConfig{
		AcceptanceCriteria: []ACPhaseData{
			{
				ACIndex:         1,
				ACTitle:         "First AC",
				CurrentPhase:    "RED",
				AttemptCount:    2,
				GateResults:     []ACGateResult{{Phase: "RED", GateType: "RED", ExitCode: 1, Classification: "reject_failure", Message: "expected command to fail"}},
				PhasesCompleted: nil,
			},
			{
				ACIndex:         2,
				ACTitle:         "Second AC",
				CurrentPhase:    "GREEN",
				AttemptCount:    1,
				GateResults:     []ACGateResult{{Phase: "GREEN", GateType: "GREEN", ExitCode: 0, Classification: "accept", Message: "ok"}},
				PhasesCompleted: []string{"RED", "VERIFY_RED"},
			},
		},
		SelectedIndex: 1,
		Width:         100,
		Height:        10,
	}))

	if !strings.Contains(rendered, "Gate detail (selected AC)") {
		t.Fatalf("expected selected AC detail header, got %q", rendered)
	}
	if !strings.Contains(rendered, "GREEN | GREEN | exit 0 | accept | ok") {
		t.Fatalf("expected selected AC gate details to be rendered, got %q", rendered)
	}
	if strings.Contains(rendered, "RED | RED | exit 1 | reject_failure | expected command to fail") {
		t.Fatalf("did not expect non-selected AC gate details, got %q", rendered)
	}
}

func TestRenderACPhaseDetailCompactVariantUsesAbbreviations(t *testing.T) {
	t.Parallel()

	rendered := stripACPhaseANSI(RenderACPhaseDetail(ACPhaseDetailConfig{
		AcceptanceCriteria: []ACPhaseData{
			{
				ACIndex:      1,
				ACTitle:      "Compact AC",
				CurrentPhase: "VERIFY_REFACTOR",
			},
		},
		SelectedIndex: 0,
		Compact:       true,
		Width:         100,
		Height:        8,
	}))

	if !strings.Contains(rendered, "[>] VRF") {
		t.Fatalf("expected compact active phase abbreviation VRF, got %q", rendered)
	}
	if strings.Contains(rendered, "VERIFY_REFACTOR") {
		t.Fatalf("did not expect full phase label in compact mode, got %q", rendered)
	}
}

func stripACPhaseANSI(input string) string {
	return acPhaseANSIPattern.ReplaceAllString(input, "")
}
