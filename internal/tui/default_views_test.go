package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewDefaultAppModelStartsOnFleetOverview(t *testing.T) {
	t.Parallel()

	model := NewDefaultAppModel()
	if got := model.CurrentView(); got != ViewFleetOverview {
		t.Fatalf("current view = %q, want %q", got, ViewFleetOverview)
	}

	rendered := model.View()
	if !strings.Contains(rendered, "FLEET COMMAND") {
		t.Fatalf("expected default model to render Fleet Overview header, got:\n%s", rendered)
	}
}

func TestDefaultViewDefinitionsIncludeFleetOverviewFocusOrder(t *testing.T) {
	t.Parallel()

	definitions := DefaultViewDefinitions()
	definition, ok := definitions[ViewFleetOverview]
	if !ok {
		t.Fatalf("missing %q view definition", ViewFleetOverview)
	}
	if len(definition.FocusOrder) == 0 {
		t.Fatal("fleet overview focus order should not be empty")
	}
	if definition.FocusOrder[0] != "ship_list_panel" {
		t.Fatalf("first focused panel = %q, want ship_list_panel", definition.FocusOrder[0])
	}
}

func TestDefaultViewDefinitionsIncludeShipBridgeFocusOrder(t *testing.T) {
	t.Parallel()

	definitions := DefaultViewDefinitions()
	definition, ok := definitions[ViewShipBridge]
	if !ok {
		t.Fatalf("missing %q view definition", ViewShipBridge)
	}
	if len(definition.FocusOrder) != 4 {
		t.Fatalf("ship bridge focus order length = %d, want 4", len(definition.FocusOrder))
	}
	if definition.FocusOrder[0] != "crew_panel" {
		t.Fatalf("first ship bridge focus panel = %q, want crew_panel", definition.FocusOrder[0])
	}
}

func TestDefaultViewDefinitionsIncludePlanReviewFocusOrder(t *testing.T) {
	t.Parallel()

	definitions := DefaultViewDefinitions()
	definition, ok := definitions[ViewPlanReview]
	if !ok {
		t.Fatalf("missing %q view definition", ViewPlanReview)
	}
	if len(definition.FocusOrder) != 3 {
		t.Fatalf("plan review focus order length = %d, want 3", len(definition.FocusOrder))
	}
	if definition.FocusOrder[0] != "manifest_panel" {
		t.Fatalf("first plan review focus panel = %q, want manifest_panel", definition.FocusOrder[0])
	}
}

func TestDefaultModelEnterNavigatesToShipBridge(t *testing.T) {
	t.Parallel()

	model := NewDefaultAppModel()
	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	typed, ok := next.(*AppModel)
	if !ok {
		t.Fatalf("update return type = %T, want *AppModel", next)
	}
	if got := typed.CurrentView(); got != ViewShipBridge {
		t.Fatalf("current view after enter = %q, want %q", got, ViewShipBridge)
	}

	rendered := typed.View()
	if !strings.Contains(rendered, "Mission Board") {
		t.Fatalf("expected ship bridge render after enter, got:\n%s", rendered)
	}
}

func TestDefaultPlanReviewRendererContainsKeySections(t *testing.T) {
	t.Parallel()

	definitions := DefaultViewDefinitions()
	definition, ok := definitions[ViewPlanReview]
	if !ok {
		t.Fatalf("missing %q view definition", ViewPlanReview)
	}

	model := NewDefaultAppModel()
	rendered := definition.Render(*model)
	for _, expected := range []string{"PLAN REVIEW", "Coverage Matrix", "Dependency Graph"} {
		if !strings.Contains(rendered, expected) {
			t.Fatalf("plan review default render missing %q\n%s", expected, rendered)
		}
	}
}
