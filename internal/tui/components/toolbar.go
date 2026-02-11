package components

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ship-commander/sc3/internal/tui/theme"
)

const toolbarSeparator = "  "

// ToolbarButton defines a single keyboard-accessible button in the toolbar.
type ToolbarButton struct {
	Key     string
	Label   string
	Action  tea.Cmd
	Enabled bool
}

// RenderNavigableToolbar renders `[key] Label` buttons separated by two spaces.
func RenderNavigableToolbar(buttons []ToolbarButton, highlighted int) string {
	if len(buttons) == 0 {
		return ""
	}

	parts := make([]string, 0, len(buttons))
	for i, button := range buttons {
		parts = append(parts, renderToolbarButton(button, i == highlighted))
	}

	return strings.Join(parts, toolbarSeparator)
}

// NextToolbarIndex returns the next enabled button index, wrapping around.
func NextToolbarIndex(buttons []ToolbarButton, current int) int {
	return walkToolbar(buttons, current, 1)
}

// PreviousToolbarIndex returns the previous enabled button index, wrapping around.
func PreviousToolbarIndex(buttons []ToolbarButton, current int) int {
	return walkToolbar(buttons, current, -1)
}

// ActivateHighlightedButton executes the highlighted button action when enabled.
func ActivateHighlightedButton(buttons []ToolbarButton, highlighted int) tea.Msg {
	if highlighted < 0 || highlighted >= len(buttons) {
		return nil
	}

	button := buttons[highlighted]
	if !button.Enabled || button.Action == nil {
		return nil
	}

	return button.Action()
}

// ActivateToolbarByKey executes the action for the first enabled key match.
func ActivateToolbarByKey(buttons []ToolbarButton, key string) tea.Msg {
	for _, button := range buttons {
		if strings.EqualFold(strings.TrimSpace(button.Key), strings.TrimSpace(key)) && button.Enabled {
			if button.Action == nil {
				return nil
			}
			return button.Action()
		}
	}
	return nil
}

func walkToolbar(buttons []ToolbarButton, current int, step int) int {
	if len(buttons) == 0 {
		return -1
	}
	if !hasEnabledButton(buttons) {
		return -1
	}

	index := current
	if index < 0 || index >= len(buttons) {
		index = 0
	}

	for attempt := 0; attempt < len(buttons); attempt++ {
		index = (index + step + len(buttons)) % len(buttons)
		if buttons[index].Enabled {
			return index
		}
	}

	return -1
}

func hasEnabledButton(buttons []ToolbarButton) bool {
	for _, button := range buttons {
		if button.Enabled {
			return true
		}
	}
	return false
}

func renderToolbarButton(button ToolbarButton, highlighted bool) string {
	style := toolbarButtonStyle(button, highlighted)

	keyStyle := lipgloss.NewStyle().Foreground(theme.ButterscotchColor)
	labelStyle := lipgloss.NewStyle().Foreground(theme.SpaceWhiteColor)

	if !button.Enabled {
		keyStyle = lipgloss.NewStyle().Foreground(theme.GalaxyGrayColor)
		labelStyle = keyStyle
	}

	content := keyStyle.Render("["+button.Key+"]") + " " + labelStyle.Render(button.Label)
	return style.Render(content)
}

func toolbarButtonStyle(button ToolbarButton, highlighted bool) lipgloss.Style {
	if !button.Enabled {
		return lipgloss.NewStyle().Foreground(theme.GalaxyGrayColor)
	}
	if highlighted {
		return lipgloss.NewStyle().
			Background(theme.MoonlitVioletColor).
			Foreground(theme.SpaceWhiteColor).
			Bold(true)
	}
	return lipgloss.NewStyle()
}
