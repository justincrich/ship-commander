package views

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestRenderShipBridgeIncludesHeaderCrewMissionEventAndLaunchedToolbar(t *testing.T) {
	t.Parallel()

	config := ShipBridgeConfig{
		Width:             128,
		ShipName:          "USS Enterprise",
		ShipClass:         "Galaxy-class",
		DirectiveTitle:    "Explore anomalies",
		Status:            ShipBridgeStatusLaunched,
		FleetHealthLabel:  "Optimal",
		PendingQuestions:  2,
		WaveCurrent:       2,
		WaveTotal:         3,
		MissionsDone:      3,
		MissionsTotal:     8,
		SelectedCrewIndex: 0,
		Crew: []ShipBridgeCrewMember{
			{Name: "Riker", Role: "Captain", MissionID: "M-003", Phase: "GREEN", Elapsed: "04:23", Status: "running"},
			{Name: "Data", Role: "Commander", MissionID: "M-007", Phase: "VERIFY_GREEN", Elapsed: "02:15", Status: "running"},
		},
		SelectedMissionIndex: 1,
		Missions: []ShipBridgeMission{
			{ID: "M-001", Title: "Backlog item", Column: "backlog", Classification: "STANDARD_OPS", AssignedAgent: "Unassigned", Phase: "PENDING", ACCompleted: 0, ACTotal: 4},
			{ID: "M-003", Title: "Auth middleware", Column: "in_progress", Classification: "RED_ALERT", AssignedAgent: "Riker", Phase: "GREEN", ACCompleted: 2, ACTotal: 4},
			{ID: "M-004", Title: "Review docs", Column: "review", Classification: "STANDARD_OPS", AssignedAgent: "Data", Phase: "VERIFY_GREEN", ACCompleted: 3, ACTotal: 4},
		},
		Events: []ShipBridgeEvent{
			{Timestamp: "14:37:22", Severity: "info", Actor: "Data", Message: "Completed gate VERIFY_GREEN"},
			{Timestamp: "14:35:18", Severity: "warn", Actor: "Riker", Message: "Awaiting approval"},
		},
	}

	rendered := RenderShipBridge(config)
	for _, expected := range []string{
		"USS Enterprise",
		"Directive: Explore anomalies",
		"Wave 2 of 3",
		"Crew (2)",
		"Mission Board",
		"B:1",
		"IP:1",
		"R:1",
		"Event Log",
		"14:37:22",
		"[h]",
		"Halt",
		"[r]",
		"Retry",
		"[d]",
		"Dock",
	} {
		if !strings.Contains(rendered, expected) {
			t.Fatalf("ship bridge missing %q\n%s", expected, rendered)
		}
	}
}

func TestRenderShipBridgeDockedToolbarAndEmptyMissionState(t *testing.T) {
	t.Parallel()

	rendered := RenderShipBridge(ShipBridgeConfig{
		Width:          120,
		ShipName:       "USS Cerritos",
		DirectiveTitle: "Survey outer rim",
		Status:         ShipBridgeStatusDocked,
		Crew: []ShipBridgeCrewMember{
			{Name: "Boimler", Role: "Ensign", Status: "idle"},
		},
	})

	for _, expected := range []string{"No missions yet. Press [p] to plan.", "[p]", "Plan", "[a]", "Assign", "[l]", "Launch"} {
		if !strings.Contains(rendered, expected) {
			t.Fatalf("docked bridge missing %q\n%s", expected, rendered)
		}
	}
	if strings.Contains(rendered, "[h]") || strings.Contains(rendered, "Halt") {
		t.Fatalf("docked toolbar must not include launched actions\n%s", rendered)
	}
}

func TestRenderShipBridgeCompactLayoutRendersAllPanels(t *testing.T) {
	t.Parallel()

	rendered := RenderShipBridge(ShipBridgeConfig{
		Width: 80,
		Crew: []ShipBridgeCrewMember{
			{Name: "Tendi", Role: "Ensign", Status: "running"},
		},
		Missions: []ShipBridgeMission{{ID: "M-001", Title: "Initialize logs", Column: "in_progress", ACCompleted: 1, ACTotal: 2}},
		Events:   []ShipBridgeEvent{{Timestamp: "09:00:00", Severity: "info", Actor: "system", Message: "Event line"}},
	})

	for _, expected := range []string{"Crew (1)", "Mission Board", "Event Log"} {
		if !strings.Contains(rendered, expected) {
			t.Fatalf("compact bridge missing %q\n%s", expected, rendered)
		}
	}
}

func TestShipBridgeEnterTargetForPanel(t *testing.T) {
	t.Parallel()

	if got := ShipBridgeEnterTargetForPanel("crew_panel"); got != "agent_detail" {
		t.Fatalf("enter target for crew_panel = %q, want agent_detail", got)
	}
	if got := ShipBridgeEnterTargetForPanel("mission_board"); got != "mission_detail" {
		t.Fatalf("enter target for mission_board = %q, want mission_detail", got)
	}
	if got := ShipBridgeEnterTargetForPanel("event_log"); got != "" {
		t.Fatalf("enter target for event_log = %q, want empty", got)
	}
}

func TestShipBridgeQuickActionForKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		key    tea.KeyMsg
		status ShipBridgeStatus
		want   ShipBridgeQuickAction
	}{
		{key: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}}, status: ShipBridgeStatusDocked, want: ShipBridgeQuickActionPlan},
		{key: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}}, status: ShipBridgeStatusDocked, want: ShipBridgeQuickActionLaunch},
		{key: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}}, status: ShipBridgeStatusLaunched, want: ShipBridgeQuickActionHalt},
		{key: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}, status: ShipBridgeStatusLaunched, want: ShipBridgeQuickActionDock},
		{key: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}, status: ShipBridgeStatusLaunched, want: ShipBridgeQuickActionHelp},
		{key: tea.KeyMsg{Type: tea.KeyEsc}, status: ShipBridgeStatusLaunched, want: ShipBridgeQuickActionFleet},
		{key: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}, status: ShipBridgeStatusLaunched, want: ShipBridgeQuickActionNone},
	}

	for _, tt := range tests {
		if got := ShipBridgeQuickActionForKey(tt.key, tt.status); got != tt.want {
			t.Fatalf("quick action for %q (%s) = %q, want %q", tt.key.String(), tt.status, got, tt.want)
		}
	}
}

func TestRenderShipBridgeEventLogClampsToLast50(t *testing.T) {
	t.Parallel()

	events := make([]ShipBridgeEvent, 0, 60)
	for i := 0; i < 60; i++ {
		events = append(events, ShipBridgeEvent{
			Timestamp: fmt.Sprintf("14:00:%02d", i),
			Severity:  "info",
			Actor:     "system",
			Message:   fmt.Sprintf("event-%02d", i),
		})
	}

	rendered := RenderShipBridge(ShipBridgeConfig{Width: 120, Events: events, Crew: []ShipBridgeCrewMember{{Name: "Riker", Status: "running"}}})
	if strings.Contains(rendered, "event-00") {
		t.Fatalf("event log should drop older entries beyond last 50\n%s", rendered)
	}
	if !strings.Contains(rendered, "event-59") {
		t.Fatalf("event log should include newest entries\n%s", rendered)
	}
}
