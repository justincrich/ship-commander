package components

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
)

var eventLogANSIPattern = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func TestResolveEventLogLineCount(t *testing.T) {
	t.Parallel()

	if got := ResolveEventLogLineCount(140, 40); got != 10 {
		t.Fatalf("standard line count = %d, want 10", got)
	}
	if got := ResolveEventLogLineCount(80, 40); got != 4 {
		t.Fatalf("compact line count = %d, want 4", got)
	}
	if got := ResolveEventLogLineCount(140, 24); got != 2 {
		t.Fatalf("small terminal line count = %d, want 2", got)
	}
}

func TestBuildEventLogViewportSupportsAutoScrollPauseAndMaxEntries(t *testing.T) {
	t.Parallel()

	events := make([]EventLogEntry, 0, 60)
	for i := 0; i < 60; i++ {
		events = append(events, EventLogEntry{
			Severity:  "INFO",
			Timestamp: fmt.Sprintf("14:00:%02d", i),
			EventType: "agent.started",
			Message:   fmt.Sprintf("event-%02d", i),
		})
	}

	auto := BuildEventLogViewport(EventLogConfig{
		Width:      120,
		Height:     4,
		Events:     events,
		AutoScroll: true,
		MaxEntries: 50,
	})
	paused := BuildEventLogViewport(EventLogConfig{
		Width:      120,
		Height:     4,
		Events:     events,
		AutoScroll: false,
		MaxEntries: 50,
	})

	if auto.YOffset <= paused.YOffset {
		t.Fatalf("expected auto-scroll to move viewport to bottom (auto=%d paused=%d)", auto.YOffset, paused.YOffset)
	}

	rendered := stripANSIEventLog(auto.View())
	if strings.Contains(rendered, "event-00") {
		t.Fatalf("expected max-entry cap to drop oldest events\n%s", rendered)
	}
	if !strings.Contains(rendered, "event-59") {
		t.Fatalf("expected latest event in viewport\n%s", rendered)
	}
}

func TestRenderEventLogAppliesSeverityFiltering(t *testing.T) {
	t.Parallel()

	rendered := stripANSIEventLog(RenderEventLog(EventLogConfig{
		Width:      120,
		Height:     4,
		AutoScroll: true,
		Events: []EventLogEntry{
			{Severity: "INFO", Timestamp: "14:00:01", EventType: "agent.started", Message: "info-entry"},
			{Severity: "WARN", Timestamp: "14:00:02", EventType: "doctor.stuck", Message: "warn-entry"},
			{Severity: "ERROR", Timestamp: "14:00:03", EventType: "gate.result", Message: "error-entry"},
		},
		SeverityFilter: []string{"ERROR"},
	}))

	if !strings.Contains(rendered, "[ERROR]") {
		t.Fatalf("expected filtered ERROR row\n%s", rendered)
	}
	if strings.Contains(rendered, "[INFO]") || strings.Contains(rendered, "[WARN]") {
		t.Fatalf("expected INFO/WARN rows filtered out\n%s", rendered)
	}
}

func TestRenderEventLogRowFormatIncludesSeverityTimestampTypeAndMessage(t *testing.T) {
	t.Parallel()

	rendered := stripANSIEventLog(RenderEventLog(EventLogConfig{
		Width:      120,
		Height:     4,
		AutoScroll: true,
		Events: []EventLogEntry{
			{
				Severity:  "INFO",
				Timestamp: "14:30:05",
				EventType: "agent.started",
				Message:   "mission=MISSION-42",
			},
		},
	}))

	for _, expected := range []string{"[INFO]", "14:30:05", "agent.started", "mission=MISSION-42"} {
		if !strings.Contains(rendered, expected) {
			t.Fatalf("event row missing %q\n%s", expected, rendered)
		}
	}
}

func stripANSIEventLog(value string) string {
	return eventLogANSIPattern.ReplaceAllString(value, "")
}
