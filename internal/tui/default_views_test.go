package tui

import (
	"strings"
	"testing"
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
