package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ship-commander/sc3/internal/tui/theme"
)

func TestNavigationStackBoundedToDepthThree(t *testing.T) {
	t.Parallel()

	model := newAppModelForTest()

	enter := tea.KeyMsg{Type: tea.KeyEnter}

	for i := 0; i < 3; i++ {
		nextModel, _ := model.Update(enter)
		model = mustAppModel(t, nextModel)
	}

	stack := model.NavigationStack()
	want := []ViewID{ViewFleetOverview, "ship_bridge", "agent_detail"}
	if fmt.Sprint(stack) != fmt.Sprint(want) {
		t.Fatalf("stack after pushes = %v, want %v", stack, want)
	}

	nextModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model = mustAppModel(t, nextModel)
	stack = model.NavigationStack()
	wantAfterEsc := []ViewID{ViewFleetOverview, "ship_bridge"}
	if fmt.Sprint(stack) != fmt.Sprint(wantAfterEsc) {
		t.Fatalf("stack after esc = %v, want %v", stack, wantAfterEsc)
	}
}

func TestPanelFocusCyclesWithTabAndShiftTab(t *testing.T) {
	t.Parallel()

	model := newAppModelForTest()

	cases := []struct {
		name        string
		key         tea.KeyMsg
		wantFocused string
	}{
		{
			name:        "tab to second panel",
			key:         tea.KeyMsg{Type: tea.KeyTab},
			wantFocused: "preview_panel",
		},
		{
			name:        "tab wraps to first panel",
			key:         tea.KeyMsg{Type: tea.KeyTab},
			wantFocused: "ship_list_panel",
		},
		{
			name:        "shift tab wraps backward",
			key:         tea.KeyMsg{Type: tea.KeyShiftTab},
			wantFocused: "preview_panel",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			nextModel, _ := model.Update(tc.key)
			model = mustAppModel(t, nextModel)

			if got := model.FocusedPanel(); got != tc.wantFocused {
				t.Fatalf("focused panel = %q, want %q", got, tc.wantFocused)
			}
		})
	}
}

func TestFocusedPanelUsesMoonlitVioletBorderAndBoldTitle(t *testing.T) {
	t.Parallel()

	model := newAppModelForTest()

	focusedBorder := model.PanelBorderStyle("ship_list_panel")
	unfocusedBorder := model.PanelBorderStyle("preview_panel")

	if !focusedBorder.GetBold() {
		t.Fatal("focused border should be bold")
	}
	if unfocusedBorder.GetBold() {
		t.Fatal("unfocused border should not be bold")
	}

	focusedColor := fmt.Sprint(focusedBorder.GetBorderTopForeground())
	unfocusedColor := fmt.Sprint(unfocusedBorder.GetBorderTopForeground())
	wantFocusedColor := fmt.Sprint(theme.PanelBorderFocused.GetBorderTopForeground())
	wantUnfocusedColor := fmt.Sprint(theme.PanelBorder.GetBorderTopForeground())
	if focusedColor != wantFocusedColor {
		t.Fatalf("focused border color = %q, want %q", focusedColor, wantFocusedColor)
	}
	if unfocusedColor != wantUnfocusedColor {
		t.Fatalf("unfocused border color = %q, want %q", unfocusedColor, wantUnfocusedColor)
	}

	focusedTitle := model.PanelTitleStyle("ship_list_panel")
	unfocusedTitle := model.PanelTitleStyle("preview_panel")
	if !focusedTitle.GetBold() {
		t.Fatal("focused title should be bold")
	}
	if unfocusedTitle.GetBold() {
		t.Fatal("unfocused title should not be bold")
	}
}

func TestWindowSizeSwitchesLayoutMode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		width  int
		height int
		want   LayoutMode
	}{
		{
			name:   "standard at threshold",
			width:  120,
			height: 30,
			want:   LayoutStandard,
		},
		{
			name:   "compact below threshold",
			width:  119,
			height: 30,
			want:   LayoutCompact,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			model := newAppModelForTest()
			nextModel, _ := model.Update(tea.WindowSizeMsg{Width: tt.width, Height: tt.height})
			updated, ok := nextModel.(*AppModel)
			if !ok {
				t.Fatalf("update return type = %T, want *AppModel", nextModel)
			}
			if got := updated.LayoutMode(); got != tt.want {
				t.Fatalf("layout mode = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGlobalShortcutsAndOverlayStack(t *testing.T) {
	t.Parallel()

	model := newAppModelForTest()

	nextModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	model = mustAppModel(t, nextModel)
	if got := model.OverlayDepth(); got != 1 {
		t.Fatalf("overlay depth after help = %d, want 1", got)
	}
	if top, ok := model.CurrentOverlay(); !ok || top.Kind != OverlayKindHelp {
		t.Fatalf("top overlay = %+v, ok=%v, want help overlay", top, ok)
	}

	nextModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	model = mustAppModel(t, nextModel)
	if got := model.OverlayDepth(); got != 2 {
		t.Fatalf("overlay depth after quit shortcut = %d, want 2", got)
	}
	if top, ok := model.CurrentOverlay(); !ok || top.Kind != OverlayKindConfirmQuit {
		t.Fatalf("top overlay = %+v, ok=%v, want confirm_quit overlay", top, ok)
	}

	nextModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model = mustAppModel(t, nextModel)
	if got := model.OverlayDepth(); got != 1 {
		t.Fatalf("overlay depth after esc = %d, want 1", got)
	}

	nextModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model = mustAppModel(t, nextModel)
	if got := model.OverlayDepth(); got != 0 {
		t.Fatalf("overlay depth after second esc = %d, want 0", got)
	}

	nextModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	model = mustAppModel(t, nextModel)
	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("enter on quit confirmation should return tea.Quit cmd")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatalf("quit command result type = %T, want tea.QuitMsg", cmd())
	}
	if !model.Quitting() {
		t.Fatal("model should be in quitting state after confirming quit")
	}
}

func TestViewDispatchUsesTopOfNavigationStack(t *testing.T) {
	t.Parallel()

	model := newAppModelForTest()
	if got := strings.TrimSpace(model.View()); got != "fleet-overview-view" {
		t.Fatalf("root view render = %q, want fleet-overview-view", got)
	}

	nextModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = mustAppModel(t, nextModel)
	if got := strings.TrimSpace(model.View()); got != "ship-bridge-view" {
		t.Fatalf("entered view render = %q, want ship-bridge-view", got)
	}

	nextModel, _ = model.Update(OverlayPushMsg{
		Overlay: Overlay{
			Kind:    OverlayKindAdmiralQuestion,
			Payload: "Need decision on mission sequencing",
		},
	})
	model = mustAppModel(t, nextModel)
	rendered := model.View()
	if !strings.Contains(rendered, "Overlay: admiral_question") {
		t.Fatalf("overlay render missing overlay title: %q", rendered)
	}
	if !strings.Contains(rendered, "Need decision on mission sequencing") {
		t.Fatalf("overlay render missing payload: %q", rendered)
	}
}

func newAppModelForTest() *AppModel {
	return NewAppModel(ViewFleetOverview, map[ViewID]ViewDefinition{
		ViewFleetOverview: {
			FocusOrder:  []string{"ship_list_panel", "preview_panel"},
			EnterTarget: "ship_bridge",
			Render: func(_ AppModel) string {
				return "fleet-overview-view"
			},
		},
		"ship_bridge": {
			FocusOrder:  []string{"crew_panel", "mission_board", "event_log"},
			EnterTarget: "mission_detail",
			Render: func(_ AppModel) string {
				return "ship-bridge-view"
			},
		},
		"mission_detail": {
			FocusOrder:  []string{"ac_phase_detail", "gate_history"},
			EnterTarget: "agent_detail",
			Render: func(_ AppModel) string {
				return "mission-detail-view"
			},
		},
		"agent_detail": {
			FocusOrder: []string{"summary", "logs"},
			Render: func(_ AppModel) string {
				return "agent-detail-view"
			},
		},
	})
}

func mustAppModel(t *testing.T, model tea.Model) *AppModel {
	t.Helper()

	typed, ok := model.(*AppModel)
	if !ok {
		t.Fatalf("update return type = %T, want *AppModel", model)
	}
	return typed
}
