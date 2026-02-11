package views

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ship-commander/sc3/internal/tui/components"
)

func TestRenderFleetMonitorIncludesHeaderGridAndToolbar(t *testing.T) {
	t.Parallel()

	rendered := RenderFleetMonitor(FleetMonitorConfig{
		Width:            120,
		Rows:             sampleFleetMonitorRows(),
		SelectedIndex:    0,
		ActiveShipCount:  3,
		TotalAgents:      9,
		FleetWaveCurrent: 4,
		FleetWaveTotal:   8,
		PendingQuestions: 2,
	})

	for _, expected := range []string{
		"FLEET MONITOR",
		"Active Ships: 3",
		"Ship Status Grid",
		"SS Enterprise",
		"[Enter]",
		"Bridge",
		"[i]",
		"Inbox",
		"[Esc]",
		"Fleet",
	} {
		if !strings.Contains(rendered, expected) {
			t.Fatalf("fleet monitor output missing %q\n%s", expected, rendered)
		}
	}
}

func TestBuildFleetMonitorRowsFiltersToLaunchedShips(t *testing.T) {
	t.Parallel()

	rows := BuildFleetMonitorRows([]components.ShipStatusRow{
		{ShipName: "Launched", Status: "launched", CrewCount: 4},
		{ShipName: "Docked", Status: "docked", CrewCount: 2},
		{ShipName: "Halted", Status: "halted", CrewCount: 3},
	}, 120, 0)

	if len(rows) != 2 {
		t.Fatalf("launched row count = %d, want 2", len(rows))
	}
	if rows[0].ShipName != "Launched" || rows[1].ShipName != "Halted" {
		t.Fatalf("unexpected launched rows order: %+v", rows)
	}
}

func TestFleetMonitorEnterTargetAndQuickActions(t *testing.T) {
	t.Parallel()

	if target := FleetMonitorEnterTarget(); target != "ship_bridge" {
		t.Fatalf("enter target = %q, want ship_bridge", target)
	}

	cases := []struct {
		key  tea.KeyMsg
		want FleetMonitorQuickAction
	}{
		{key: tea.KeyMsg{Type: tea.KeyEnter}, want: FleetMonitorQuickActionBridge},
		{key: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}}, want: FleetMonitorQuickActionInbox},
		{key: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}, want: FleetMonitorQuickActionHelp},
		{key: tea.KeyMsg{Type: tea.KeyEsc}, want: FleetMonitorQuickActionFleet},
		{key: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}, want: FleetMonitorQuickActionNone},
	}

	for _, testCase := range cases {
		if got := FleetMonitorQuickActionForKey(testCase.key); got != testCase.want {
			t.Fatalf("quick action for %q = %q, want %q", testCase.key.String(), got, testCase.want)
		}
	}
}

func TestFleetMonitorAutoRefreshCmd(t *testing.T) {
	t.Parallel()

	if got := FleetMonitorAutoRefreshInterval(); got != 2*time.Second {
		t.Fatalf("auto-refresh interval = %s, want 2s", got)
	}
	msg := FleetMonitorAutoRefreshCmd(0)()
	if _, ok := msg.(FleetMonitorTickMsg); !ok {
		t.Fatalf("refresh message type = %T, want FleetMonitorTickMsg", msg)
	}
}

func TestResolveFleetMonitorLayout(t *testing.T) {
	t.Parallel()

	if got := ResolveFleetMonitorLayout(100); got != FleetMonitorLayoutCompact {
		t.Fatalf("layout(100) = %q, want compact", got)
	}
	if got := ResolveFleetMonitorLayout(120); got != FleetMonitorLayoutStandard {
		t.Fatalf("layout(120) = %q, want standard", got)
	}
}

func sampleFleetMonitorRows() []components.ShipStatusRow {
	return []components.ShipStatusRow{
		{
			ShipName:       "SS Enterprise",
			DirectiveTitle: "Implement auth system",
			Status:         "launched",
			CrewCount:      4,
			WaveCurrent:    3,
			WaveTotal:      6,
			MissionsDone:   12,
			MissionsTotal:  18,
			Health:         80,
		},
		{
			ShipName:       "HMS Beagle",
			DirectiveTitle: "Database migration",
			Status:         "complete",
			CrewCount:      2,
			WaveCurrent:    5,
			WaveTotal:      5,
			MissionsDone:   15,
			MissionsTotal:  15,
			Health:         100,
		},
		{
			ShipName:       "Discovery",
			DirectiveTitle: "API docs",
			Status:         "halted",
			CrewCount:      3,
			WaveCurrent:    1,
			WaveTotal:      4,
			MissionsDone:   3,
			MissionsTotal:  12,
			Health:         40,
		},
	}
}
