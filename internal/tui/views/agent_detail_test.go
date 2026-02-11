package views

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestRenderAgentDetailIncludesHeaderAndOutputViewport(t *testing.T) {
	t.Parallel()

	rendered := RenderAgentDetail(AgentDetailConfig{
		Width:      120,
		AgentName:  "Cmdr. Data",
		Role:       "Commander",
		ShipName:   "USS Enterprise",
		Harness:    "claude-code",
		Model:      "sonnet-4",
		MissionID:  "M-005",
		Phase:      "VERIFY_GREEN",
		Elapsed:    "12m 34s",
		Status:     "running",
		AutoScroll: true,
		OutputLines: []string{
			"[14:27:48] Running test suite...",
			"✓ should generate valid JWT token",
			"✗ should enforce rate limiting",
		},
	})

	for _, expected := range []string{"Cmdr. Data", "COMMANDER", "USS Enterprise", "M-005", "VERIFY_GREEN", "Agent Output [AUTO]", "Running test suite"} {
		if !strings.Contains(rendered, expected) {
			t.Fatalf("agent detail missing %q\n%s", expected, rendered)
		}
	}
}

func TestRenderAgentDetailConditionalErrorContext(t *testing.T) {
	t.Parallel()

	stuck := RenderAgentDetail(AgentDetailConfig{
		Width:              120,
		AgentName:          "Cmdr. Data",
		Status:             "stuck",
		OutputLines:        []string{"last line"},
		StuckTimestamp:     "14:28:05",
		LastGateFailure:    "VERIFY_GREEN failed (exit 1)",
		TimeoutInfo:        "180s exceeded",
		LastMeaningfulLine: "expected 3 assertions, got 2",
	})
	if !strings.Contains(stuck, "STUCK DETECTED") || !strings.Contains(stuck, "VERIFY_GREEN failed") {
		t.Fatalf("stuck agent should render error context\n%s", stuck)
	}

	running := RenderAgentDetail(AgentDetailConfig{
		Width:       120,
		AgentName:   "Cmdr. Data",
		Status:      "running",
		OutputLines: []string{"line"},
	})
	if strings.Contains(running, "STUCK DETECTED") {
		t.Fatalf("running agent should not render error context\n%s", running)
	}
}

func TestAgentDetailToolbarAndQuickActions(t *testing.T) {
	t.Parallel()

	rendered := RenderAgentDetail(AgentDetailConfig{
		Width:       120,
		AgentName:   "Cmdr. Data",
		Status:      "running",
		OutputLines: []string{"line"},
	})

	for _, expected := range []string{"[h]", "Halt", "[r]", "Retry", "[i]", "Ignore Stuck", "[Esc]", "Back"} {
		if !strings.Contains(rendered, expected) {
			t.Fatalf("agent toolbar missing %q\n%s", expected, rendered)
		}
	}

	tests := []struct {
		key  tea.KeyMsg
		want AgentDetailQuickAction
	}{
		{key: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}}, want: AgentDetailQuickActionHalt},
		{key: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}, want: AgentDetailQuickActionRetry},
		{key: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}}, want: AgentDetailQuickActionIgnoreStuck},
		{key: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}, want: AgentDetailQuickActionHelp},
		{key: tea.KeyMsg{Type: tea.KeyEsc}, want: AgentDetailQuickActionBack},
		{key: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}, want: AgentDetailQuickActionNone},
	}

	for _, tt := range tests {
		if got := AgentDetailQuickActionForKey(tt.key); got != tt.want {
			t.Fatalf("quick action for %q = %q, want %q", tt.key.String(), got, tt.want)
		}
	}
}
