package views

import (
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ship-commander/sc3/internal/tui/theme"
)

const (
	helpOverlayDefaultWidth      = 120
	helpOverlayDefaultHeight     = 30
	helpOverlayStandardWidthPct  = 0.70
	helpOverlayCompactWidthPct   = 0.85
	helpOverlayCompactThreshold  = 120
	helpOverlayMinimumModalWidth = 56
)

// HelpOverlayContext identifies the active view context for help filtering.
type HelpOverlayContext string

const (
	// HelpOverlayContextGlobal represents generic/global help.
	HelpOverlayContextGlobal HelpOverlayContext = "global"
	// HelpOverlayContextFleetOverview represents Fleet Overview help.
	HelpOverlayContextFleetOverview HelpOverlayContext = "fleet_overview"
	// HelpOverlayContextShipBridge represents Ship Bridge help.
	HelpOverlayContextShipBridge HelpOverlayContext = "ship_bridge"
	// HelpOverlayContextReadyRoom represents Ready Room help.
	HelpOverlayContextReadyRoom HelpOverlayContext = "ready_room"
	// HelpOverlayContextAgentRoster represents Agent Roster help.
	HelpOverlayContextAgentRoster HelpOverlayContext = "agent_roster"
	// HelpOverlayContextMessageCenter represents Message Center help.
	HelpOverlayContextMessageCenter HelpOverlayContext = "message_center"
)

// HelpOverlayQuickAction captures direct close actions in the help overlay.
type HelpOverlayQuickAction string

const (
	// HelpOverlayQuickActionNone indicates no action.
	HelpOverlayQuickActionNone HelpOverlayQuickAction = ""
	// HelpOverlayQuickActionClose closes the help overlay.
	HelpOverlayQuickActionClose HelpOverlayQuickAction = "close"
)

// HelpOverlayConfig contains all rendering inputs for help overlay.
type HelpOverlayConfig struct {
	Width   int
	Height  int
	Context HelpOverlayContext
}

// HelpOverlaySection represents one logical keybinding section.
type HelpOverlaySection struct {
	Title    string
	Bindings []key.Binding
}

// HelpOverlayQuickActionForKey resolves close actions for `?` and `Esc`.
func HelpOverlayQuickActionForKey(msg tea.KeyMsg) HelpOverlayQuickAction {
	switch strings.ToLower(strings.TrimSpace(msg.String())) {
	case "?", "esc":
		return HelpOverlayQuickActionClose
	default:
		return HelpOverlayQuickActionNone
	}
}

// RenderHelpOverlay renders the full-screen keyboard shortcut modal.
func RenderHelpOverlay(config HelpOverlayConfig) string {
	width := config.Width
	if width <= 0 {
		width = helpOverlayDefaultWidth
	}
	height := config.Height
	if height <= 0 {
		height = helpOverlayDefaultHeight
	}

	modalWidth := int(float64(width) * helpOverlayStandardWidthPct)
	if width < helpOverlayCompactThreshold {
		modalWidth = int(float64(width) * helpOverlayCompactWidthPct)
	}
	if modalWidth < helpOverlayMinimumModalWidth {
		modalWidth = helpOverlayMinimumModalWidth
	}
	if modalWidth > width {
		modalWidth = width
	}

	context := normalizeHelpOverlayContext(config.Context)
	sections := BuildHelpOverlaySections(context)
	contentWidth := max(24, modalWidth-4)

	helpModel := help.New()
	helpModel.Width = contentWidth
	helpModel.ShowAll = true

	title := lipgloss.NewStyle().
		Foreground(theme.ButterscotchColor).
		Bold(true).
		Align(lipgloss.Center).
		Width(contentWidth).
		Render("KEYBOARD SHORTCUTS")
	subtitle := lipgloss.NewStyle().
		Foreground(theme.LightGrayColor).
		Faint(true).
		Align(lipgloss.Center).
		Width(contentWidth).
		Render("Context: " + formatHelpContextLabel(context))

	sectionBlocks := make([]string, 0, len(sections))
	for _, section := range sections {
		header := lipgloss.NewStyle().Foreground(theme.IceColor).Bold(true).Render(strings.ToUpper(section.Title))
		body := strings.TrimSpace(helpModel.FullHelpView([][]key.Binding{section.Bindings}))
		if body == "" {
			body = "No shortcuts defined."
		}
		sectionBlocks = append(sectionBlocks, header, body)
	}

	hint := lipgloss.NewStyle().
		Foreground(theme.LightGrayColor).
		Faint(true).
		Align(lipgloss.Center).
		Width(contentWidth).
		Render("Press ? or Escape to close")

	body := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		subtitle,
		lipgloss.NewStyle().Foreground(theme.GalaxyGrayColor).Render(strings.Repeat("─", contentWidth)),
		lipgloss.JoinVertical(lipgloss.Left, sectionBlocks...),
		lipgloss.NewStyle().Foreground(theme.GalaxyGrayColor).Render(strings.Repeat("─", contentWidth)),
		hint,
	)

	modal := lipgloss.NewStyle().
		Width(modalWidth).
		Border(lipgloss.DoubleBorder()).
		BorderForeground(theme.BlueColor).
		Padding(0, 1).
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

// BuildHelpOverlaySections returns global + navigation + context-specific bindings.
func BuildHelpOverlaySections(context HelpOverlayContext) []HelpOverlaySection {
	sections := []HelpOverlaySection{
		{
			Title: "Global",
			Bindings: []key.Binding{
				newHelpBinding([]string{"tab"}, "Tab", "Cycle panel focus"),
				newHelpBinding([]string{"shift+tab"}, "Shift+Tab", "Reverse panel focus"),
				newHelpBinding([]string{"?"}, "?", "Toggle help"),
				newHelpBinding([]string{"q"}, "q", "Quit"),
				newHelpBinding([]string{"ctrl+c"}, "Ctrl+C", "Force quit"),
			},
		},
		{
			Title: "Navigation",
			Bindings: []key.Binding{
				newHelpBinding([]string{"enter"}, "Enter", "Drill down / Select"),
				newHelpBinding([]string{"esc"}, "Escape", "Go back / Cancel"),
				newHelpBinding([]string{"up", "down"}, "Up/Down", "Navigate items"),
				newHelpBinding([]string{"pgup", "pgdown"}, "PgUp/PgDn", "Scroll viewport"),
			},
		},
	}

	contextTitle, contextBindings := contextHelpBindings(normalizeHelpOverlayContext(context))
	if len(contextBindings) > 0 {
		sections = append(sections, HelpOverlaySection{
			Title:    contextTitle,
			Bindings: contextBindings,
		})
	}
	return sections
}

func contextHelpBindings(context HelpOverlayContext) (string, []key.Binding) {
	switch context {
	case HelpOverlayContextFleetOverview:
		return "Fleet Overview", []key.Binding{
			newHelpBinding([]string{"n"}, "n", "New ship"),
			newHelpBinding([]string{"d"}, "d", "Directive"),
			newHelpBinding([]string{"r"}, "r", "Roster"),
			newHelpBinding([]string{"i"}, "i", "Inbox"),
			newHelpBinding([]string{"m"}, "m", "Monitor"),
			newHelpBinding([]string{"s"}, "s", "Settings"),
		}
	case HelpOverlayContextShipBridge:
		return "Ship Bridge", []key.Binding{
			newHelpBinding([]string{"p"}, "p", "Open Ready Room"),
			newHelpBinding([]string{"l"}, "l", "Launch ship"),
			newHelpBinding([]string{"a"}, "a", "Agent detail"),
			newHelpBinding([]string{"h"}, "h", "Halt mission/agent"),
			newHelpBinding([]string{"r"}, "r", "Retry mission"),
			newHelpBinding([]string{"w"}, "w", "Wave manager"),
			newHelpBinding([]string{" "}, "Space", "Pause/resume"),
		}
	case HelpOverlayContextReadyRoom:
		return "Ready Room", []key.Binding{
			newHelpBinding([]string{"a"}, "a", "Approve"),
			newHelpBinding([]string{"f"}, "f", "Feedback"),
			newHelpBinding([]string{"s"}, "s", "Shelve"),
		}
	case HelpOverlayContextAgentRoster:
		return "Agent Roster", []key.Binding{
			newHelpBinding([]string{"n"}, "n", "New agent"),
			newHelpBinding([]string{"enter"}, "Enter", "Edit agent"),
			newHelpBinding([]string{"a"}, "a", "Assign"),
			newHelpBinding([]string{"d"}, "d", "Detach"),
			newHelpBinding([]string{"backspace"}, "Backspace", "Delete"),
		}
	case HelpOverlayContextMessageCenter:
		return "Message Center", []key.Binding{
			newHelpBinding([]string{"g"}, "g", "Go to question"),
			newHelpBinding([]string{"d"}, "d", "Dismiss question"),
			newHelpBinding([]string{"shift+d"}, "D", "Dismiss all"),
			newHelpBinding([]string{"1", "2", "3", "0"}, "1/2/3/0", "Severity filter"),
		}
	default:
		return "Global", nil
	}
}

func newHelpBinding(keys []string, helpKey string, description string) key.Binding {
	return key.NewBinding(
		key.WithKeys(keys...),
		key.WithHelp(helpKey, description),
	)
}

func normalizeHelpOverlayContext(context HelpOverlayContext) HelpOverlayContext {
	switch HelpOverlayContext(strings.ToLower(strings.TrimSpace(string(context)))) {
	case HelpOverlayContextFleetOverview:
		return HelpOverlayContextFleetOverview
	case HelpOverlayContextShipBridge:
		return HelpOverlayContextShipBridge
	case HelpOverlayContextReadyRoom:
		return HelpOverlayContextReadyRoom
	case HelpOverlayContextAgentRoster:
		return HelpOverlayContextAgentRoster
	case HelpOverlayContextMessageCenter:
		return HelpOverlayContextMessageCenter
	default:
		return HelpOverlayContextGlobal
	}
}

func formatHelpContextLabel(context HelpOverlayContext) string {
	switch context {
	case HelpOverlayContextFleetOverview:
		return "Fleet Overview"
	case HelpOverlayContextShipBridge:
		return "Ship Bridge"
	case HelpOverlayContextReadyRoom:
		return "Ready Room"
	case HelpOverlayContextAgentRoster:
		return "Agent Roster"
	case HelpOverlayContextMessageCenter:
		return "Message Center"
	default:
		return "Global"
	}
}
