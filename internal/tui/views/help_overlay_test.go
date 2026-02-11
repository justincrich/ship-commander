package views

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestRenderHelpOverlayIncludesGlobalAndContextSections(t *testing.T) {
	t.Parallel()

	rendered := RenderHelpOverlay(HelpOverlayConfig{
		Width:   120,
		Height:  30,
		Context: HelpOverlayContextShipBridge,
	})

	for _, expected := range []string{
		"KEYBOARD SHORTCUTS",
		"Context: Ship Bridge",
		"GLOBAL",
		"NAVIGATION",
		"SHIP BRIDGE",
		"Tab",
		"Enter",
		"Launch ship",
		"Press ? or Escape to close",
		"╔",
	} {
		if !strings.Contains(rendered, expected) {
			t.Fatalf("help overlay missing %q\n%s", expected, rendered)
		}
	}
}

func TestBuildHelpOverlaySectionsAlwaysIncludesGlobalAndNavigation(t *testing.T) {
	t.Parallel()

	sections := BuildHelpOverlaySections(HelpOverlayContextMessageCenter)
	if len(sections) != 3 {
		t.Fatalf("section count = %d, want 3", len(sections))
	}
	if sections[0].Title != "Global" {
		t.Fatalf("first section = %q, want Global", sections[0].Title)
	}
	if sections[1].Title != "Navigation" {
		t.Fatalf("second section = %q, want Navigation", sections[1].Title)
	}
	if sections[2].Title != "Message Center" {
		t.Fatalf("third section = %q, want Message Center", sections[2].Title)
	}
}

func TestHelpOverlayQuickActionForKey(t *testing.T) {
	t.Parallel()

	cases := []struct {
		key  tea.KeyMsg
		want HelpOverlayQuickAction
	}{
		{key: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}, want: HelpOverlayQuickActionClose},
		{key: tea.KeyMsg{Type: tea.KeyEsc}, want: HelpOverlayQuickActionClose},
		{key: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}, want: HelpOverlayQuickActionNone},
	}

	for _, testCase := range cases {
		if got := HelpOverlayQuickActionForKey(testCase.key); got != testCase.want {
			t.Fatalf("quick action for %q = %q, want %q", testCase.key.String(), got, testCase.want)
		}
	}
}

func TestFormatHelpContextLabel(t *testing.T) {
	t.Parallel()

	if got := formatHelpContextLabel(HelpOverlayContextFleetOverview); got != "Fleet Overview" {
		t.Fatalf("fleet context label = %q", got)
	}
	if got := formatHelpContextLabel(HelpOverlayContextGlobal); got != "Global" {
		t.Fatalf("global context label = %q", got)
	}
}
