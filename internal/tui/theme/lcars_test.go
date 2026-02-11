package theme

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func TestLCARSColorConstantsAndIcons(t *testing.T) {
	t.Parallel()

	colors := map[string]string{
		"Butterscotch":  Butterscotch,
		"Blue":          Blue,
		"Purple":        Purple,
		"Pink":          Pink,
		"Gold":          Gold,
		"Almond":        Almond,
		"RedAlert":      RedAlert,
		"YellowCaution": YellowCaution,
		"GreenOk":       GreenOk,
		"Black":         Black,
		"DarkBlue":      DarkBlue,
		"GalaxyGray":    GalaxyGray,
		"SpaceWhite":    SpaceWhite,
		"LightGray":     LightGray,
		"MoonlitViolet": MoonlitViolet,
		"Ice":           Ice,
	}
	if len(colors) != 16 {
		t.Fatalf("color constant count = %d, want 16", len(colors))
	}
	for name, value := range colors {
		if value == "" {
			t.Fatalf("%s constant is empty", name)
		}
	}

	icons := []string{
		IconDone,
		IconWorking,
		IconWaiting,
		IconSkipped,
		IconFailed,
		IconAlert,
		IconRunning,
		IconBlocked,
	}
	for i, icon := range icons {
		if icon == "" {
			t.Fatalf("icon index %d is empty", i)
		}
	}
}

func TestSemanticStylesAndBordersExported(t *testing.T) {
	t.Parallel()

	semantic := []lipgloss.Style{
		ActiveStyle,
		SuccessStyle,
		ErrorStyle,
		WarningStyle,
		InfoStyle,
		PlanningStyle,
		NotifyStyle,
		FocusStyle,
	}
	if len(semantic) != 8 {
		t.Fatalf("semantic style count = %d, want 8", len(semantic))
	}
	for i, style := range semantic {
		if style.GetForeground() == nil {
			t.Fatalf("semantic style %d has nil foreground", i)
		}
	}

	if border, _, _, _, _ := PanelBorder.GetBorder(); border.Top != lipgloss.RoundedBorder().Top {
		t.Fatalf("panel border style top = %q, want rounded top %q", border.Top, lipgloss.RoundedBorder().Top)
	}
	if border, _, _, _, _ := PanelBorderFocused.GetBorder(); border.Top != lipgloss.RoundedBorder().Top {
		t.Fatalf("focused panel border style top = %q, want rounded top %q", border.Top, lipgloss.RoundedBorder().Top)
	}
	if border, _, _, _, _ := OverlayBorder.GetBorder(); border.Top != lipgloss.DoubleBorder().Top {
		t.Fatalf("overlay border style top = %q, want double top %q", border.Top, lipgloss.DoubleBorder().Top)
	}
	if !PanelBorderFocused.GetBold() {
		t.Fatal("focused panel border should be bold")
	}
}

func TestLCARSColorRespectsProfileAndAdaptiveColor(t *testing.T) {
	original := colorProfileFn
	t.Cleanup(func() {
		colorProfileFn = original
	})

	colorProfileFn = func() termenv.Profile { return termenv.TrueColor }
	trueColor := lcarsColor(Butterscotch, "209", "11")
	if _, ok := trueColor.(lipgloss.AdaptiveColor); !ok {
		t.Fatalf("truecolor result type = %T, want lipgloss.AdaptiveColor", trueColor)
	}

	colorProfileFn = func() termenv.Profile { return termenv.ANSI256 }
	ansi256Color := lcarsColor(Butterscotch, "209", "11")
	complete, ok := ansi256Color.(lipgloss.CompleteAdaptiveColor)
	if !ok {
		t.Fatalf("ansi256 result type = %T, want lipgloss.CompleteAdaptiveColor", ansi256Color)
	}
	if complete.Dark.ANSI256 != "209" || complete.Light.ANSI != "11" {
		t.Fatalf("complete adaptive color = %#v", complete)
	}
}
