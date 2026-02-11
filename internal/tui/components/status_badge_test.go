package components

import (
	"strings"
	"testing"
)

func TestRenderStatusBadgeVariants(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		status string
		icon   string
		label  string
	}{
		{name: "running", status: "running", icon: "●", label: "RUNNING"},
		{name: "done", status: "done", icon: "✓", label: "DONE"},
		{name: "waiting", status: "waiting", icon: "⏸", label: "WAITING"},
		{name: "skipped", status: "skipped", icon: "⊘", label: "SKIPPED"},
		{name: "failed", status: "failed", icon: "✗", label: "FAILED"},
		{name: "stuck", status: "stuck", icon: "●", label: "STUCK"},
		{name: "halted", status: "halted", icon: "✗", label: "HALTED"},
		{name: "planning", status: "planning", icon: "●", label: "PLANNING"},
		{name: "review", status: "review", icon: "●", label: "REVIEW"},
		{name: "approved", status: "approved", icon: "✓", label: "APPROVED"},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			rendered := RenderStatusBadge(testCase.status)
			expected := testCase.icon + " " + testCase.label
			if !strings.Contains(rendered, expected) {
				t.Fatalf("rendered badge %q does not include %q", rendered, expected)
			}
		})
	}
}

func TestRenderStatusBadgeOptions(t *testing.T) {
	t.Parallel()

	withoutIcon := RenderStatusBadge("running", WithBadgeIcon(false))
	if strings.Contains(withoutIcon, "● RUNNING") {
		t.Fatalf("expected icon to be omitted, got %q", withoutIcon)
	}
	if !strings.Contains(withoutIcon, "RUNNING") {
		t.Fatalf("expected label to remain when icon hidden, got %q", withoutIcon)
	}

	standard := RenderStatusBadge("review")
	bold := RenderStatusBadge("review", WithBadgeBold(true))
	if !strings.Contains(bold, "● REVIEW") {
		t.Fatalf("expected bold rendering to preserve icon+label, got %q", bold)
	}
	if !strings.Contains(standard, "● REVIEW") {
		t.Fatalf("expected standard rendering to include icon+label, got %q", standard)
	}
}

func TestRenderStatusBadgeUnknownStatus(t *testing.T) {
	t.Parallel()

	rendered := RenderStatusBadge("  in_progress  ")
	if !strings.Contains(rendered, "⚠ IN_PROGRESS") {
		t.Fatalf("unexpected unknown rendering: %q", rendered)
	}

	empty := RenderStatusBadge("")
	if !strings.Contains(empty, "⚠ UNKNOWN") {
		t.Fatalf("empty status should render UNKNOWN, got %q", empty)
	}
}
