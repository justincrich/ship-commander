package components

import (
	"fmt"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"github.com/ship-commander/sc3/internal/tui/theme"
)

const (
	eventLogCompactWidthThreshold = 120
	eventLogStandardLines         = 10
	eventLogCompactLines          = 4
	eventLogSmallTerminalLines    = 2
	eventLogDefaultMaxEntries     = 50
)

// EventLogEntry represents one event line in the viewport.
type EventLogEntry struct {
	Severity  string
	Timestamp string
	EventType string
	Message   string
}

// EventLogConfig contains render-time settings for the EventLog component.
type EventLogConfig struct {
	Width          int
	Height         int
	TerminalHeight int
	Events         []EventLogEntry
	MaxEntries     int
	AutoScroll     bool
	SeverityFilter []string
}

// ResolveEventLogLineCount computes visible viewport lines for the current layout.
func ResolveEventLogLineCount(width int, terminalHeight int) int {
	if terminalHeight > 0 && terminalHeight < 30 {
		return eventLogSmallTerminalLines
	}
	if width > 0 && width < eventLogCompactWidthThreshold {
		return eventLogCompactLines
	}
	return eventLogStandardLines
}

// BuildEventLogViewport constructs a viewport model for event log content.
func BuildEventLogViewport(config EventLogConfig) viewport.Model {
	viewWidth := config.Width
	if viewWidth < 24 {
		viewWidth = 24
	}

	viewHeight := ResolveEventLogLineCount(config.Width, config.TerminalHeight)
	if config.Height > 0 && config.Height < viewHeight {
		viewHeight = config.Height
	}
	if viewHeight < 2 {
		viewHeight = 2
	}

	lines := formatEventLines(config.Events, config.SeverityFilter, config.MaxEntries)
	if len(lines) == 0 {
		lines = []string{lipgloss.NewStyle().Foreground(theme.GalaxyGrayColor).Faint(true).Render("No recent events")}
	}

	model := viewport.New(viewWidth, viewHeight)
	model.SetContent(strings.Join(lines, "\n"))
	if config.AutoScroll {
		model.GotoBottom()
	}
	return model
}

// RenderEventLog renders a scrollable event viewport string.
func RenderEventLog(config EventLogConfig) string {
	return BuildEventLogViewport(config).View()
}

func formatEventLines(events []EventLogEntry, severityFilter []string, maxEntries int) []string {
	allowed := normalizeSeverityFilter(severityFilter)
	filtered := make([]string, 0, len(events))
	for _, event := range events {
		if !isSeverityAllowed(event.Severity, allowed) {
			continue
		}
		filtered = append(filtered, renderEventRow(event))
	}

	limit := maxEntries
	if limit <= 0 {
		limit = eventLogDefaultMaxEntries
	}
	if len(filtered) > limit {
		filtered = append([]string(nil), filtered[len(filtered)-limit:]...)
	}
	return filtered
}

func renderEventRow(event EventLogEntry) string {
	severity := normalizeSeverity(event.Severity)
	timestamp := strings.TrimSpace(event.Timestamp)
	if timestamp == "" {
		timestamp = "--:--:--"
	}
	eventType := strings.TrimSpace(event.EventType)
	if eventType == "" {
		eventType = "event.unknown"
	}
	message := strings.TrimSpace(event.Message)
	if message == "" {
		message = "(no message)"
	}

	severityText := strings.ToUpper(severity)
	severityStyle := lipgloss.NewStyle().Foreground(theme.BlueColor).Bold(true)
	switch severity {
	case "WARN":
		severityStyle = lipgloss.NewStyle().Foreground(theme.YellowCautionColor).Bold(true)
	case "ERROR":
		severityStyle = lipgloss.NewStyle().Foreground(theme.RedAlertColor).Bold(true)
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		severityStyle.Render(fmt.Sprintf("[%s]", severityText)),
		" ",
		lipgloss.NewStyle().Foreground(theme.LightGrayColor).Render(timestamp),
		" ",
		lipgloss.NewStyle().Foreground(theme.SpaceWhiteColor).Render(eventType),
		" ",
		lipgloss.NewStyle().Foreground(theme.SpaceWhiteColor).Render(message),
	)
}

func normalizeSeverityFilter(filter []string) []string {
	if len(filter) == 0 {
		return []string{"INFO", "WARN", "ERROR"}
	}

	allowed := make([]string, 0, len(filter))
	for _, severity := range filter {
		normalized := normalizeSeverity(severity)
		if normalized == "" {
			continue
		}
		if !slices.Contains(allowed, normalized) {
			allowed = append(allowed, normalized)
		}
	}
	if len(allowed) == 0 {
		return []string{"INFO", "WARN", "ERROR"}
	}
	return allowed
}

func isSeverityAllowed(severity string, allowed []string) bool {
	normalized := normalizeSeverity(severity)
	if normalized == "" {
		return false
	}
	return slices.Contains(allowed, normalized)
}

func normalizeSeverity(severity string) string {
	switch strings.ToUpper(strings.TrimSpace(severity)) {
	case "INFO":
		return "INFO"
	case "WARN", "WARNING":
		return "WARN"
	case "ERROR", "FAILED":
		return "ERROR"
	default:
		return ""
	}
}
