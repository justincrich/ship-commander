package components

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

type toolbarTestMsg string

func testCmd(value string) tea.Cmd {
	return func() tea.Msg {
		return toolbarTestMsg(value)
	}
}

func TestRenderNavigableToolbar(t *testing.T) {
	t.Parallel()

	output := RenderNavigableToolbar([]ToolbarButton{
		{Key: "n", Label: "New", Enabled: true},
		{Key: "q", Label: "Quit", Enabled: true},
	}, 0)

	if output == "" {
		t.Fatal("expected non-empty toolbar render output")
	}
	if !containsAll(output, "[n]", "New", "[q]", "Quit") {
		t.Fatalf("expected toolbar output to include key+label pairs, got %q", output)
	}
	if !containsAll(output, "  ") {
		t.Fatalf("expected toolbar output to include two-space separator, got %q", output)
	}
}

func TestToolbarIndexNavigation(t *testing.T) {
	t.Parallel()

	buttons := []ToolbarButton{
		{Key: "n", Label: "New", Enabled: true},
		{Key: "d", Label: "Directive", Enabled: false},
		{Key: "q", Label: "Quit", Enabled: true},
	}

	if got := NextToolbarIndex(buttons, 0); got != 2 {
		t.Fatalf("next index = %d, want 2", got)
	}
	if got := NextToolbarIndex(buttons, 2); got != 0 {
		t.Fatalf("wrapped next index = %d, want 0", got)
	}
	if got := PreviousToolbarIndex(buttons, 2); got != 0 {
		t.Fatalf("previous index = %d, want 0", got)
	}
	if got := PreviousToolbarIndex(buttons, 0); got != 2 {
		t.Fatalf("wrapped previous index = %d, want 2", got)
	}
}

func TestActivateHighlightedButton(t *testing.T) {
	t.Parallel()

	buttons := []ToolbarButton{
		{Key: "n", Label: "New", Enabled: true, Action: testCmd("new")},
		{Key: "d", Label: "Disabled", Enabled: false, Action: testCmd("disabled")},
	}

	msg := ActivateHighlightedButton(buttons, 0)
	if msg != toolbarTestMsg("new") {
		t.Fatalf("highlighted activation message = %#v, want %#v", msg, toolbarTestMsg("new"))
	}
	if msg := ActivateHighlightedButton(buttons, 1); msg != nil {
		t.Fatalf("disabled highlighted button should not activate, got %#v", msg)
	}
}

func TestActivateToolbarByKey(t *testing.T) {
	t.Parallel()

	buttons := []ToolbarButton{
		{Key: "n", Label: "New", Enabled: true, Action: testCmd("new")},
		{Key: "q", Label: "Quit", Enabled: true, Action: testCmd("quit")},
		{Key: "d", Label: "Disabled", Enabled: false, Action: testCmd("disabled")},
	}

	if msg := ActivateToolbarByKey(buttons, "q"); msg != toolbarTestMsg("quit") {
		t.Fatalf("quick-key activation message = %#v, want %#v", msg, toolbarTestMsg("quit"))
	}
	if msg := ActivateToolbarByKey(buttons, "D"); msg != nil {
		t.Fatalf("disabled quick-key should not activate, got %#v", msg)
	}
	if msg := ActivateToolbarByKey(buttons, "x"); msg != nil {
		t.Fatalf("unknown quick-key should not activate, got %#v", msg)
	}
}

func TestToolbarButtonStyleStates(t *testing.T) {
	t.Parallel()

	disabled := toolbarButtonStyle(ToolbarButton{Enabled: false}, false)
	if disabled.GetForeground() == nil {
		t.Fatal("disabled button should have foreground color")
	}

	highlighted := toolbarButtonStyle(ToolbarButton{Enabled: true}, true)
	if highlighted.GetBackground() == nil {
		t.Fatal("highlighted button should have background color")
	}
}

func TestNavigationWithNoEnabledButtons(t *testing.T) {
	t.Parallel()

	buttons := []ToolbarButton{
		{Enabled: false},
		{Enabled: false},
	}
	if got := NextToolbarIndex(buttons, 0); got != -1 {
		t.Fatalf("next index with no enabled buttons = %d, want -1", got)
	}
	if got := PreviousToolbarIndex(buttons, 0); got != -1 {
		t.Fatalf("previous index with no enabled buttons = %d, want -1", got)
	}
}

func containsAll(text string, parts ...string) bool {
	for _, part := range parts {
		if !contains(text, part) {
			return false
		}
	}
	return true
}

func contains(text string, part string) bool {
	return strings.Contains(text, part)
}
