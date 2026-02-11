package components

import (
	"regexp"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

var confirmDialogANSIPattern = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func TestRenderConfirmDialogVariantsAndContent(t *testing.T) {
	t.Parallel()

	destructive := stripANSIConfirmDialog(RenderConfirmDialog(ConfirmDialogConfig{
		Width:       120,
		Height:      30,
		Destructive: true,
		Title:       "HALT MISSION?",
		Message:     "Agent impl-bravo will be terminated.",
		Consequence: "Worktree will be preserved.",
	}))
	for _, expected := range []string{"[!]", "HALT MISSION?", "Agent impl-bravo will be terminated.", "Worktree will be preserved.", "Yes", "No", "╔"} {
		if !strings.Contains(destructive, expected) {
			t.Fatalf("destructive dialog missing %q\n%s", expected, destructive)
		}
	}

	standard := stripANSIConfirmDialog(RenderConfirmDialog(ConfirmDialogConfig{
		Width:       120,
		Height:      30,
		Destructive: false,
		Title:       "SHELVE PLAN?",
		Message:     "This will save the current mission manifest.",
		Consequence: "Ready Room specialists will be released.",
	}))
	for _, expected := range []string{"[~]", "SHELVE PLAN?", "This will save the current mission manifest.", "Yes", "No", "╔"} {
		if !strings.Contains(standard, expected) {
			t.Fatalf("standard dialog missing %q\n%s", expected, standard)
		}
	}
}

func TestConfirmDialogQuickActionForKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		key  tea.KeyMsg
		want ConfirmDialogQuickAction
	}{
		{key: tea.KeyMsg{Type: tea.KeyLeft}, want: ConfirmDialogQuickActionSelectConfirm},
		{key: tea.KeyMsg{Type: tea.KeyRight}, want: ConfirmDialogQuickActionSelectCancel},
		{key: tea.KeyMsg{Type: tea.KeyEnter}, want: ConfirmDialogQuickActionSubmit},
		{key: tea.KeyMsg{Type: tea.KeyEsc}, want: ConfirmDialogQuickActionDismiss},
		{key: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}, want: ConfirmDialogQuickActionNone},
	}

	for _, tt := range tests {
		if got := ConfirmDialogQuickActionForKey(tt.key); got != tt.want {
			t.Fatalf("action for key %q = %q, want %q", tt.key.String(), got, tt.want)
		}
	}
}

func TestToggleConfirmDialogSelection(t *testing.T) {
	t.Parallel()

	if got := ToggleConfirmDialogSelection(false, ConfirmDialogQuickActionSelectConfirm); !got {
		t.Fatal("select confirm should set selection true")
	}
	if got := ToggleConfirmDialogSelection(true, ConfirmDialogQuickActionSelectCancel); got {
		t.Fatal("select cancel should set selection false")
	}
	if got := ToggleConfirmDialogSelection(true, ConfirmDialogQuickActionNone); !got {
		t.Fatal("no action should preserve selection")
	}
}

func TestResolveConfirmDialogDecision(t *testing.T) {
	t.Parallel()

	finished, confirmed := ResolveConfirmDialogDecision(true, ConfirmDialogQuickActionSubmit)
	if !finished || !confirmed {
		t.Fatalf("submit with affirmative selection should confirm (finished=%v confirmed=%v)", finished, confirmed)
	}

	finished, confirmed = ResolveConfirmDialogDecision(true, ConfirmDialogQuickActionDismiss)
	if !finished || confirmed {
		t.Fatalf("dismiss should cancel (finished=%v confirmed=%v)", finished, confirmed)
	}

	finished, confirmed = ResolveConfirmDialogDecision(false, ConfirmDialogQuickActionNone)
	if finished || confirmed {
		t.Fatalf("non-terminal action should not resolve (finished=%v confirmed=%v)", finished, confirmed)
	}
}

func stripANSIConfirmDialog(value string) string {
	return confirmDialogANSIPattern.ReplaceAllString(value, "")
}
