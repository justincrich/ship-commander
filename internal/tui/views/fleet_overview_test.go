package views

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestRenderFleetOverviewIncludesHeaderCardsAndToolbar(t *testing.T) {
	t.Parallel()

	config := FleetOverviewConfig{
		Width:            120,
		SelectedIndex:    0,
		PendingMessages:  4,
		FleetHealthLabel: "Optimal",
		Ships: []FleetOverviewShip{
			{
				Name:           "USS Reliant",
				Class:          "Miranda-class",
				Status:         "complete",
				DirectiveTitle: "Survey anomalies",
				CrewCount:      4,
				MissionsDone:   6,
				MissionsTotal:  6,
			},
			{
				Name:             "USS Enterprise",
				Class:            "Constitution-class",
				Status:           "launched",
				DirectiveTitle:   "Explore strange new worlds",
				DirectiveSummary: "Explore strange new worlds and seek out new life.",
				CrewCount:        5,
				WaveCurrent:      3,
				WaveTotal:        5,
				MissionsDone:     3,
				MissionsTotal:    5,
				Crew: []FleetOverviewCrewMember{
					{Name: "kirk", Role: "Commander"},
				},
				RecentActivity: []string{"Wave 3 started"},
			},
			{
				Name:           "USS Discovery",
				Class:          "Crossfield-class",
				Status:         "docked",
				DirectiveTitle: "",
				CrewCount:      3,
				MissionsDone:   0,
				MissionsTotal:  8,
			},
		},
	}

	rendered := RenderFleetOverview(config)

	for _, expected := range []string{
		"FLEET COMMAND",
		"Ships: 3",
		"Launched: 1",
		"Messages:",
		"[4]",
		"Ship List",
		"Ship Preview",
		"USS Enterprise",
		"Explore strange new worlds",
		"RUNNING",
		"[n]",
		"New",
		"[q]",
		"Quit",
	} {
		if !strings.Contains(rendered, expected) {
			t.Fatalf("fleet overview missing %q\n%s", expected, rendered)
		}
	}
}

func TestRenderFleetOverviewSortsLaunchedDockedComplete(t *testing.T) {
	t.Parallel()

	rendered := RenderFleetOverview(FleetOverviewConfig{
		Width:         120,
		SelectedIndex: 0,
		Ships: []FleetOverviewShip{
			{Name: "Complete Ship", Status: "complete"},
			{Name: "Docked Ship", Status: "docked"},
			{Name: "Launched Ship", Status: "launched"},
		},
	})

	launchedIndex := strings.Index(rendered, "Launched Ship")
	dockedIndex := strings.Index(rendered, "Docked Ship")
	completeIndex := strings.Index(rendered, "Complete Ship")
	if launchedIndex == -1 || dockedIndex == -1 || completeIndex == -1 {
		t.Fatalf("expected all ship names in output, got:\n%s", rendered)
	}
	if !(launchedIndex < dockedIndex && dockedIndex < completeIndex) {
		t.Fatalf("unexpected sort order launched/docked/complete in output:\n%s", rendered)
	}
}

func TestRenderFleetOverviewCompactHidesPreview(t *testing.T) {
	t.Parallel()

	rendered := RenderFleetOverview(FleetOverviewConfig{
		Width: 80,
		Ships: []FleetOverviewShip{
			{Name: "USS Enterprise", Status: "launched"},
		},
	})

	if strings.Contains(rendered, "Ship Preview") {
		t.Fatalf("compact layout should hide preview panel:\n%s", rendered)
	}
	if !strings.Contains(rendered, "Ship List") {
		t.Fatalf("compact layout must keep ship list panel:\n%s", rendered)
	}
}

func TestRenderFleetOverviewEmptyState(t *testing.T) {
	t.Parallel()

	rendered := RenderFleetOverview(FleetOverviewConfig{
		Width: 120,
	})
	expected := "Your fleet is empty. Press [n] to commission your first starship."
	if !strings.Contains(rendered, expected) {
		t.Fatalf("empty state missing expected message %q\n%s", expected, rendered)
	}
}

func TestFleetOverviewSelectionHelpers(t *testing.T) {
	t.Parallel()

	ships := []FleetOverviewShip{{Name: "A"}, {Name: "B"}, {Name: "C"}}

	if got := NextFleetSelection(ships, 0); got != 1 {
		t.Fatalf("next selection = %d, want 1", got)
	}
	if got := NextFleetSelection(ships, 2); got != 0 {
		t.Fatalf("wrapped next selection = %d, want 0", got)
	}
	if got := PreviousFleetSelection(ships, 0); got != 2 {
		t.Fatalf("wrapped previous selection = %d, want 2", got)
	}
	if got := NextFleetSelection(nil, 0); got != -1 {
		t.Fatalf("next selection on empty = %d, want -1", got)
	}
	if got := PreviousFleetSelection(nil, 0); got != -1 {
		t.Fatalf("previous selection on empty = %d, want -1", got)
	}
	if target := FleetOverviewEnterTarget(); target != "ship_bridge" {
		t.Fatalf("enter target = %q, want ship_bridge", target)
	}
}

func TestResolveFleetOverviewLayout(t *testing.T) {
	t.Parallel()

	if got := ResolveFleetOverviewLayout(80); got != FleetOverviewLayoutCompact {
		t.Fatalf("layout for width 80 = %q, want compact", got)
	}
	if got := ResolveFleetOverviewLayout(120); got != FleetOverviewLayoutStandard {
		t.Fatalf("layout for width 120 = %q, want standard", got)
	}
}

func TestFleetOverviewQuickActionForKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		key  tea.KeyMsg
		want FleetOverviewQuickAction
	}{
		{key: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}, want: FleetQuickActionNew},
		{key: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}, want: FleetQuickActionDirective},
		{key: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}, want: FleetQuickActionRoster},
		{key: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}}, want: FleetQuickActionInbox},
		{key: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}}, want: FleetQuickActionSettings},
		{key: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}, want: FleetQuickActionHelp},
		{key: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}, want: FleetQuickActionQuit},
		{key: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}, want: FleetQuickActionNone},
	}

	for _, tt := range tests {
		if got := FleetOverviewQuickActionForKey(tt.key); got != tt.want {
			t.Fatalf("quick action for key %q = %q, want %q", tt.key.String(), got, tt.want)
		}
	}
}
