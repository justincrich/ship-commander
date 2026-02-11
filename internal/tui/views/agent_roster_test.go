package views

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestRenderAgentRosterIncludesThreeColumnLayoutAndToolbar(t *testing.T) {
	t.Parallel()

	rendered := RenderAgentRoster(sampleAgentRosterConfig(128))
	for _, expected := range []string{
		"AGENT ROSTER",
		"Roles",
		"Agent List",
		"Agent Detail",
		"capt-alpha",
		"PROFILE",
		"MISSION PROMPT",
		"BACKSTORY",
		"[n]",
		"New",
		"[Enter]",
		"Edit",
		"[Esc]",
		"Fleet",
	} {
		if !strings.Contains(rendered, expected) {
			t.Fatalf("agent roster missing %q\n%s", expected, rendered)
		}
	}
}

func TestRenderAgentRosterEmptyState(t *testing.T) {
	t.Parallel()

	rendered := RenderAgentRoster(AgentRosterConfig{Width: 120})
	expected := "Your roster is empty. Press [n] to recruit your first crew member."
	if !strings.Contains(rendered, expected) {
		t.Fatalf("empty state missing %q\n%s", expected, rendered)
	}
}

func TestFilterAgentRosterByRole(t *testing.T) {
	t.Parallel()

	agents := []AgentRosterAgent{
		{Name: "capt-alpha", Role: "captain"},
		{Name: "cmdr-beta", Role: "commander"},
		{Name: "impl-gamma", Role: "implementer"},
		{Name: "none", Role: "unassigned"},
	}

	if got := len(FilterAgentRosterByRole(agents, AgentRosterRoleAll)); got != 4 {
		t.Fatalf("all filter count = %d, want 4", got)
	}
	if got := len(FilterAgentRosterByRole(agents, AgentRosterRoleCaptains)); got != 1 {
		t.Fatalf("captain filter count = %d, want 1", got)
	}
	if got := len(FilterAgentRosterByRole(agents, AgentRosterRoleCommanders)); got != 1 {
		t.Fatalf("commander filter count = %d, want 1", got)
	}
	if got := len(FilterAgentRosterByRole(agents, AgentRosterRoleEnsigns)); got != 1 {
		t.Fatalf("ensign filter count = %d, want 1", got)
	}
	if got := len(FilterAgentRosterByRole(agents, AgentRosterRoleUnassigned)); got != 1 {
		t.Fatalf("unassigned filter count = %d, want 1", got)
	}
}

func TestAgentRosterQuickActionForKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		key  tea.KeyMsg
		want AgentRosterQuickAction
	}{
		{key: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}, want: AgentRosterQuickActionNew},
		{key: tea.KeyMsg{Type: tea.KeyEnter}, want: AgentRosterQuickActionEdit},
		{key: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}, want: AgentRosterQuickActionAssign},
		{key: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}, want: AgentRosterQuickActionDetach},
		{key: tea.KeyMsg{Type: tea.KeyDelete}, want: AgentRosterQuickActionDelete},
		{key: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}, want: AgentRosterQuickActionSearch},
		{key: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}, want: AgentRosterQuickActionHelp},
		{key: tea.KeyMsg{Type: tea.KeyEsc}, want: AgentRosterQuickActionFleet},
		{key: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}, want: AgentRosterQuickActionNone},
	}

	for _, tt := range tests {
		if got := AgentRosterQuickActionForKey(tt.key); got != tt.want {
			t.Fatalf("quick action for key %q = %q, want %q", tt.key.String(), got, tt.want)
		}
	}
}

func TestResolveAgentRosterLayout(t *testing.T) {
	t.Parallel()

	if got := ResolveAgentRosterLayout(80); got != AgentRosterLayoutCompact {
		t.Fatalf("layout for width 80 = %q, want compact", got)
	}
	if got := ResolveAgentRosterLayout(120); got != AgentRosterLayoutStandard {
		t.Fatalf("layout for width 120 = %q, want standard", got)
	}
}

func sampleAgentRosterConfig(width int) AgentRosterConfig {
	return AgentRosterConfig{
		Width:              width,
		SelectedRoleIndex:  0,
		SelectedAgentIndex: 0,
		ToolbarHighlighted: 0,
		Agents: []AgentRosterAgent{
			{
				Name:          "capt-alpha",
				Role:          "captain",
				Model:         "claude-sonnet-4",
				Status:        "active",
				Assignment:    "SS Enterprise",
				MissionID:     "MISSION-01",
				Phase:         "PLANNING",
				Skills:        []string{"/plan", "/review", "/delegate"},
				MissionPrompt: "Lead the Enterprise crew through mission planning and execution.",
				Backstory:     "Experienced command officer with a focus on quality gates.",
			},
			{
				Name:       "impl-bravo",
				Role:       "implementer",
				Model:      "gpt-5-codex",
				Status:     "stuck",
				Assignment: "SS Enterprise",
				MissionID:  "MISSION-03",
				Phase:      "RED",
				Skills:     []string{"/implement", "/test"},
			},
			{
				Name:       "cmdr-charlie",
				Role:       "commander",
				Model:      "claude-sonnet-4",
				Status:     "idle",
				Assignment: "SS Nautilus",
				MissionID:  "",
				Phase:      "IDLE",
				Skills:     []string{"/monitor"},
			},
		},
	}
}
