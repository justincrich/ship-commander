package components

import (
	"strings"
	"testing"
)

func TestRenderPhaseIndicatorFullVariant(t *testing.T) {
	t.Parallel()

	output := RenderPhaseIndicator("green", []string{"red", "verify_red"}, false)

	if !containsAll(output,
		"RED",
		"VERIFY_RED",
		"GREEN",
		"VERIFY_GREEN",
		"REFACTOR",
		"VERIFY_REFACTOR",
	) {
		t.Fatalf("full output missing expected phase labels: %q", output)
	}
	if !strings.Contains(output, " -> ") {
		t.Fatalf("full output should include arrow separator, got %q", output)
	}
}

func TestRenderPhaseIndicatorCompactVariant(t *testing.T) {
	t.Parallel()

	output := RenderPhaseIndicator("green", []string{"red", "verify_red"}, true)
	expected := "R > VR > G > VG > RF > VRF"
	if !strings.Contains(output, expected) {
		t.Fatalf("compact output mismatch, expected %q in %q", expected, output)
	}
}

func TestRenderPhaseIndicatorNormalizesInput(t *testing.T) {
	t.Parallel()

	output := RenderPhaseIndicator("VERIFY-GREEN", []string{"VERIFY RED", "red"}, false)
	if !containsAll(output, "VERIFY_GREEN", "VERIFY_RED", "RED") {
		t.Fatalf("normalized output should still include canonical labels, got %q", output)
	}
}

func TestPhaseLabelStyleStates(t *testing.T) {
	t.Parallel()

	current := phaseLabelStyle(true, true)
	if !current.GetBold() {
		t.Fatal("current phase should be bold")
	}
	if current.GetForeground() == nil {
		t.Fatal("current phase should set foreground color")
	}

	completed := phaseLabelStyle(false, true)
	if !completed.GetFaint() {
		t.Fatal("completed phase should be dim/faint")
	}
	if completed.GetForeground() == nil {
		t.Fatal("completed phase should set foreground color")
	}

	future := phaseLabelStyle(false, false)
	if !future.GetFaint() {
		t.Fatal("future phase should be dim/faint")
	}
	if future.GetForeground() == nil {
		t.Fatal("future phase should set foreground color")
	}
}

func TestPhaseSeparatorVariants(t *testing.T) {
	t.Parallel()

	if got := phaseSeparator(false); got != " -> " {
		t.Fatalf("full separator = %q, want %q", got, " -> ")
	}
	if got := phaseSeparator(true); got != " > " {
		t.Fatalf("compact separator = %q, want %q", got, " > ")
	}
}
