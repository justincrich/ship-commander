package components

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/ship-commander/sc3/internal/tui/theme"
)

const (
	confirmDialogDefaultWidth      = 120
	confirmDialogDefaultHeight     = 30
	confirmDialogStandardWidthPct  = 0.54
	confirmDialogCompactWidthPct   = 0.80
	confirmDialogCompactThreshold  = 120
	confirmDialogMinimumModalWidth = 48
)

// ConfirmDialogQuickAction captures direct keyboard intents.
type ConfirmDialogQuickAction string

const (
	// ConfirmDialogQuickActionNone indicates no action key matched.
	ConfirmDialogQuickActionNone ConfirmDialogQuickAction = ""
	// ConfirmDialogQuickActionSelectConfirm highlights confirm action.
	ConfirmDialogQuickActionSelectConfirm ConfirmDialogQuickAction = "select_confirm"
	// ConfirmDialogQuickActionSelectCancel highlights cancel action.
	ConfirmDialogQuickActionSelectCancel ConfirmDialogQuickAction = "select_cancel"
	// ConfirmDialogQuickActionSubmit confirms current selection.
	ConfirmDialogQuickActionSubmit ConfirmDialogQuickAction = "submit"
	// ConfirmDialogQuickActionDismiss cancels and dismisses the dialog.
	ConfirmDialogQuickActionDismiss ConfirmDialogQuickAction = "dismiss"
)

// ConfirmDialogConfig defines render payload for destructive/standard confirm modal.
type ConfirmDialogConfig struct {
	Width            int
	Height           int
	Title            string
	Message          string
	Consequence      string
	ConfirmLabel     string
	CancelLabel      string
	Destructive      bool
	ConfirmSelected  bool
	SelectionDefined bool
}

// ConfirmDialogQuickActionForKey resolves keyboard actions for the confirm modal.
func ConfirmDialogQuickActionForKey(msg tea.KeyMsg) ConfirmDialogQuickAction {
	switch strings.ToLower(strings.TrimSpace(msg.String())) {
	case "left", "h":
		return ConfirmDialogQuickActionSelectConfirm
	case "right", "l":
		return ConfirmDialogQuickActionSelectCancel
	case "enter":
		return ConfirmDialogQuickActionSubmit
	case "esc":
		return ConfirmDialogQuickActionDismiss
	default:
		return ConfirmDialogQuickActionNone
	}
}

// ToggleConfirmDialogSelection applies directional selection changes.
func ToggleConfirmDialogSelection(current bool, action ConfirmDialogQuickAction) bool {
	switch action {
	case ConfirmDialogQuickActionSelectConfirm:
		return true
	case ConfirmDialogQuickActionSelectCancel:
		return false
	default:
		return current
	}
}

// ResolveConfirmDialogDecision resolves completion status and final decision.
func ResolveConfirmDialogDecision(selectedConfirm bool, action ConfirmDialogQuickAction) (finished bool, confirmed bool) {
	switch action {
	case ConfirmDialogQuickActionDismiss:
		return true, false
	case ConfirmDialogQuickActionSubmit:
		return true, selectedConfirm
	default:
		return false, false
	}
}

// RenderConfirmDialog renders a centered modal with destructive/standard variant styling.
func RenderConfirmDialog(config ConfirmDialogConfig) string {
	width := config.Width
	if width <= 0 {
		width = confirmDialogDefaultWidth
	}
	height := config.Height
	if height <= 0 {
		height = confirmDialogDefaultHeight
	}

	modalWidth := int(float64(width) * confirmDialogStandardWidthPct)
	if width < confirmDialogCompactThreshold {
		modalWidth = int(float64(width) * confirmDialogCompactWidthPct)
	}
	if modalWidth < confirmDialogMinimumModalWidth {
		modalWidth = confirmDialogMinimumModalWidth
	}
	if modalWidth > width {
		modalWidth = width
	}

	confirmLabel := strings.TrimSpace(config.ConfirmLabel)
	if confirmLabel == "" {
		confirmLabel = "Yes"
	}
	cancelLabel := strings.TrimSpace(config.CancelLabel)
	if cancelLabel == "" {
		cancelLabel = "No"
	}

	title := strings.TrimSpace(config.Title)
	if title == "" {
		if config.Destructive {
			title = "HALT MISSION?"
		} else {
			title = "SHELVE PLAN?"
		}
	}

	message := strings.TrimSpace(config.Message)
	if message == "" {
		message = "Confirm this action?"
	}

	consequence := strings.TrimSpace(config.Consequence)
	if consequence == "" {
		consequence = "This action cannot be undone."
	}

	confirmSelected := config.ConfirmSelected
	if !config.SelectionDefined {
		confirmSelected = config.Destructive
	}

	icon := "[~]"
	borderColor := theme.GoldColor
	titleColor := theme.YellowCautionColor
	if config.Destructive {
		icon = "[!]"
		borderColor = theme.RedAlertColor
		titleColor = theme.RedAlertColor
	}

	value := confirmSelected
	confirmField := huh.NewConfirm().
		Title("Proceed?").
		Affirmative(confirmLabel).
		Negative(cancelLabel).
		Value(&value)
	_ = confirmField.Init()
	confirmView := strings.TrimSpace(confirmField.View())
	if confirmView == "" {
		confirmView = renderConfirmDialogFallbackButtons(confirmSelected, confirmLabel, cancelLabel, config.Destructive)
	}

	body := lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.NewStyle().Foreground(titleColor).Bold(true).Align(lipgloss.Center).Width(maxInt(20, modalWidth-6)).Render(icon+" "+title),
		lipgloss.NewStyle().Foreground(theme.SpaceWhiteColor).Align(lipgloss.Center).Width(maxInt(20, modalWidth-6)).Render(message),
		lipgloss.NewStyle().Foreground(theme.LightGrayColor).Align(lipgloss.Center).Width(maxInt(20, modalWidth-6)).Render(consequence),
		lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(theme.GalaxyGrayColor).Padding(0, 1).Render(consequence),
		confirmView,
		lipgloss.NewStyle().
			Foreground(theme.GalaxyGrayColor).
			Faint(true).
			Align(lipgloss.Center).
			Width(maxInt(20, modalWidth-6)).
			Render("Left/Right to select  Enter confirm  Esc cancel"),
	)

	modal := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(borderColor).
		Padding(1, 2).
		Width(modalWidth).
		Render(body)

	return lipgloss.Place(
		width,
		height,
		lipgloss.Center,
		lipgloss.Center,
		modal,
		lipgloss.WithWhitespaceChars("┄"),
		lipgloss.WithWhitespaceForeground(theme.GalaxyGrayColor),
	)
}

func renderConfirmDialogFallbackButtons(confirmSelected bool, confirmLabel string, cancelLabel string, destructive bool) string {
	confirmStyle := lipgloss.NewStyle().Foreground(theme.LightGrayColor).Border(lipgloss.RoundedBorder()).BorderForeground(theme.GalaxyGrayColor).Padding(0, 1)
	cancelStyle := lipgloss.NewStyle().Foreground(theme.LightGrayColor).Border(lipgloss.RoundedBorder()).BorderForeground(theme.GalaxyGrayColor).Padding(0, 1)

	if confirmSelected {
		bg := theme.YellowCautionColor
		if destructive {
			bg = theme.RedAlertColor
		}
		confirmStyle = lipgloss.NewStyle().Background(bg).Foreground(theme.BlackColor).Bold(true).Padding(0, 1)
	} else {
		cancelStyle = lipgloss.NewStyle().Background(theme.GalaxyGrayColor).Foreground(theme.SpaceWhiteColor).Bold(true).Padding(0, 1)
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		confirmStyle.Render(confirmLabel),
		"  ",
		cancelStyle.Render(cancelLabel),
	)
}

func maxInt(a int, b int) int {
	if a > b {
		return a
	}
	return b
}
