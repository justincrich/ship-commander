package theme

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

const (
	// Butterscotch is the primary active LCARS accent color.
	Butterscotch = "#FF9966"
	// Blue is the LCARS informational blue.
	Blue = "#9999CC"
	// Purple is the LCARS planning/review purple.
	Purple = "#CC99CC"
	// Pink is the LCARS notification pink.
	Pink = "#FF99CC"
	// Gold is the LCARS selection highlight gold.
	Gold = "#FFAA00"
	// Almond is the LCARS secondary warm accent.
	Almond = "#FFAA90"
	// RedAlert is the LCARS high-severity red.
	RedAlert = "#FF3333"
	// YellowCaution is the LCARS caution yellow.
	YellowCaution = "#FFCC00"
	// GreenOk is the LCARS success green.
	GreenOk = "#33FF33"
	// Black is the LCARS primary background.
	Black = "#000000"
	// DarkBlue is the LCARS panel surface color.
	DarkBlue = "#1B4F8F"
	// GalaxyGray is the LCARS muted neutral.
	GalaxyGray = "#52526A"
	// SpaceWhite is the LCARS primary text color.
	SpaceWhite = "#F5F6FA"
	// LightGray is the LCARS secondary text color.
	LightGray = "#CCCCCC"
	// MoonlitViolet is the LCARS focus ring color.
	MoonlitViolet = "#9966FF"
	// Ice is the LCARS cool accent for highlights.
	Ice = "#99CCFF"
)

const (
	// IconDone indicates completed work.
	IconDone = "✓"
	// IconWorking indicates active work.
	IconWorking = "●"
	// IconWaiting indicates waiting state.
	IconWaiting = "⏸"
	// IconSkipped indicates skipped work.
	IconSkipped = "⊘"
	// IconFailed indicates failed work.
	IconFailed = "✗"
	// IconAlert indicates warning/alert state.
	IconAlert = "⚠"
	// IconRunning indicates actively running state.
	IconRunning = "▸"
	// IconBlocked indicates blocked state.
	IconBlocked = "◆"
)

var (
	// ButterscotchColor is the profile-aware terminal color for Butterscotch.
	ButterscotchColor = lcarsColor(Butterscotch, "209", "11")
	// BlueColor is the profile-aware terminal color for Blue.
	BlueColor = lcarsColor(Blue, "146", "12")
	// PurpleColor is the profile-aware terminal color for Purple.
	PurpleColor = lcarsColor(Purple, "182", "13")
	// PinkColor is the profile-aware terminal color for Pink.
	PinkColor = lcarsColor(Pink, "218", "13")
	// GoldColor is the profile-aware terminal color for Gold.
	GoldColor = lcarsColor(Gold, "214", "11")
	// AlmondColor is the profile-aware terminal color for Almond.
	AlmondColor = lcarsColor(Almond, "216", "7")
	// RedAlertColor is the profile-aware terminal color for RedAlert.
	RedAlertColor = lcarsColor(RedAlert, "203", "9")
	// YellowCautionColor is the profile-aware terminal color for YellowCaution.
	YellowCautionColor = lcarsColor(YellowCaution, "220", "11")
	// GreenOkColor is the profile-aware terminal color for GreenOk.
	GreenOkColor = lcarsColor(GreenOk, "46", "10")
	// BlackColor is the profile-aware terminal color for Black.
	BlackColor = lcarsColor(Black, "16", "0")
	// DarkBlueColor is the profile-aware terminal color for DarkBlue.
	DarkBlueColor = lcarsColor(DarkBlue, "25", "4")
	// GalaxyGrayColor is the profile-aware terminal color for GalaxyGray.
	GalaxyGrayColor = lcarsColor(GalaxyGray, "60", "8")
	// SpaceWhiteColor is the profile-aware terminal color for SpaceWhite.
	SpaceWhiteColor = lcarsColor(SpaceWhite, "255", "15")
	// LightGrayColor is the profile-aware terminal color for LightGray.
	LightGrayColor = lcarsColor(LightGray, "252", "7")
	// MoonlitVioletColor is the profile-aware terminal color for MoonlitViolet.
	MoonlitVioletColor = lcarsColor(MoonlitViolet, "99", "5")
	// IceColor is the profile-aware terminal color for Ice.
	IceColor = lcarsColor(Ice, "153", "14")
)

var (
	// ActiveStyle marks currently active interface elements.
	ActiveStyle = lipgloss.NewStyle().Foreground(ButterscotchColor).Bold(true)
	// SuccessStyle marks successful/completed states.
	SuccessStyle = lipgloss.NewStyle().Foreground(GreenOkColor).Bold(true)
	// ErrorStyle marks error/failure states.
	ErrorStyle = lipgloss.NewStyle().Foreground(RedAlertColor).Bold(true)
	// WarningStyle marks warning/caution states.
	WarningStyle = lipgloss.NewStyle().Foreground(YellowCautionColor).Bold(true)
	// InfoStyle marks informational states.
	InfoStyle = lipgloss.NewStyle().Foreground(BlueColor)
	// PlanningStyle marks planning/review states.
	PlanningStyle = lipgloss.NewStyle().Foreground(PurpleColor)
	// NotifyStyle marks notifications and inbox counts.
	NotifyStyle = lipgloss.NewStyle().Foreground(PinkColor).Bold(true)
	// FocusStyle marks currently focused controls.
	FocusStyle = lipgloss.NewStyle().Foreground(MoonlitVioletColor).Bold(true)
)

var (
	// PanelBorder is the default panel border style.
	PanelBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(GalaxyGrayColor)

	// PanelBorderFocused is the focused panel border with emphasized title styling.
	PanelBorderFocused = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(MoonlitVioletColor).
				Bold(true)

	// OverlayBorder is the modal/overlay border style.
	OverlayBorder = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(ButterscotchColor)

	// PanelTitleFocusedStyle is the focused panel-title style helper.
	PanelTitleFocusedStyle = lipgloss.NewStyle().Foreground(MoonlitVioletColor).Bold(true)
)

var colorProfileFn = lipgloss.ColorProfile

func lcarsColor(hex string, ansi256 string, ansi string) lipgloss.TerminalColor {
	switch colorProfileFn() {
	case termenv.TrueColor:
		// Use AdaptiveColor even in TrueColor mode for light/dark terminal detection.
		return lipgloss.AdaptiveColor{Light: hex, Dark: hex}
	case termenv.ANSI256, termenv.ANSI:
		return lipgloss.CompleteAdaptiveColor{
			Light: lipgloss.CompleteColor{
				TrueColor: hex,
				ANSI256:   ansi256,
				ANSI:      ansi,
			},
			Dark: lipgloss.CompleteColor{
				TrueColor: hex,
				ANSI256:   ansi256,
				ANSI:      ansi,
			},
		}
	default:
		return lipgloss.AdaptiveColor{Light: hex, Dark: hex}
	}
}
