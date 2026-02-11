package invariants

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestInvariantViolationAddsEventToActiveSpan(t *testing.T) {
	previous := Enabled()
	SetEnabled(true)
	t.Cleanup(func() {
		SetEnabled(previous)
	})

	recorder, restore := installTracerProvider()
	defer restore()

	ctx, span := otel.Tracer("test/invariants").Start(context.Background(), "operation")
	InvariantViolation(ctx, InvariantPatchApplyClean, SeverityError, ViolationDetails{
		WhatInvariant: "patch clean apply",
		WhereDetected: "commander.verifyMissionOutput",
		WhyViolated:   "patch had fuzzy hunks",
		StackTrace:    "trace",
		Additional: map[string]string{
			"mission_id": "mission-1",
		},
	})
	span.End()

	events := spanEventsByName(recorder, "operation")
	require.Len(t, events, 1)
	assert.Equal(t, "invariant.violation", events[0].Name)
	assert.Equal(t, InvariantPatchApplyClean, eventAttr(events[0], "invariant_name"))
	assert.Equal(t, SeverityError, eventAttr(events[0], "severity"))
	assert.Equal(t, "commander.verifyMissionOutput", eventAttr(events[0], "where_detected"))
	assert.Equal(t, "mission-1", eventAttr(events[0], "context.mission_id"))
}

func TestInvariantViolationDisabledSkipsEmission(t *testing.T) {
	previous := Enabled()
	SetEnabled(false)
	t.Cleanup(func() {
		SetEnabled(previous)
	})

	recorder, restore := installTracerProvider()
	defer restore()

	ctx, span := otel.Tracer("test/invariants").Start(context.Background(), "operation")
	InvariantViolation(ctx, InvariantPatchApplyClean, SeverityError, ViolationDetails{
		WhereDetected: "commander.verifyMissionOutput",
	})
	span.End()

	events := spanEventsByName(recorder, "operation")
	require.Len(t, events, 0)
}

func TestPredefinedInvariantChecksEmitExpectedNames(t *testing.T) {
	previous := Enabled()
	SetEnabled(true)
	t.Cleanup(func() {
		SetEnabled(previous)
	})

	tests := []struct {
		name          string
		wantInvariant string
		run           func(ctx context.Context) bool
	}{
		{
			name:          "patch_apply_clean",
			wantInvariant: InvariantPatchApplyClean,
			run: func(ctx context.Context) bool {
				return CheckPatchApplyClean(ctx, "commander.verifyMissionOutput", false, "reject file created")
			},
		},
		{
			name:          "repo_clean_before_merge",
			wantInvariant: InvariantRepoCleanBeforeMerge,
			run: func(ctx context.Context) bool {
				return CheckRepoCleanBeforeMerge(ctx, "commander.handleReviewVerdict", false, "M internal/commander/commander.go")
			},
		},
		{
			name:          "max_retries_not_exceeded",
			wantInvariant: InvariantMaxRetriesNotExceeded,
			run: func(ctx context.Context) bool {
				return CheckMaxRetriesNotExceeded(ctx, "commander.handleReviewVerdict", 4, 3)
			},
		},
		{
			name:          "edits_within_allowed_paths",
			wantInvariant: InvariantEditsWithinAllowedPaths,
			run: func(ctx context.Context) bool {
				return CheckEditsWithinAllowedPaths(ctx, "commander.runMission", []string{"internal/commander/**"}, []string{"cmd/sc3/main.go"})
			},
		},
		{
			name:          "state_transition_legal",
			wantInvariant: InvariantStateTransitionLegal,
			run: func(ctx context.Context) bool {
				return CheckStateTransitionLegal(ctx, "state.machine.transition", "mission", "backlog", "invalid", false)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			recorder, restore := installTracerProvider()
			defer restore()

			ctx, span := otel.Tracer("test/invariants").Start(context.Background(), "operation")
			assert.False(t, tt.run(ctx))
			span.End()

			events := spanEventsByName(recorder, "operation")
			require.Len(t, events, 1)
			assert.Equal(t, tt.wantInvariant, eventAttr(events[0], "invariant_name"))
		})
	}
}

func TestCheckRepoCleanBeforeMergeUsesWarnSeverity(t *testing.T) {
	previous := Enabled()
	SetEnabled(true)
	t.Cleanup(func() {
		SetEnabled(previous)
	})

	recorder, restore := installTracerProvider()
	defer restore()

	ctx, span := otel.Tracer("test/invariants").Start(context.Background(), "operation")
	assert.False(t, CheckRepoCleanBeforeMerge(ctx, "commander.handleReviewVerdict", false, "M file.go"))
	span.End()

	events := spanEventsByName(recorder, "operation")
	require.Len(t, events, 1)
	assert.Equal(t, SeverityWarn, eventAttr(events[0], "severity"))
}

func installTracerProvider() (*tracetest.SpanRecorder, func()) {
	recorder := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	previous := otel.GetTracerProvider()
	otel.SetTracerProvider(provider)

	return recorder, func() {
		if err := provider.Shutdown(context.Background()); err != nil {
			otel.Handle(err)
		}
		otel.SetTracerProvider(previous)
	}
}

func spanEventsByName(recorder *tracetest.SpanRecorder, spanName string) []sdktrace.Event {
	for _, finished := range recorder.Ended() {
		if finished.Name() != spanName {
			continue
		}
		return finished.Events()
	}
	return nil
}

func eventAttr(event sdktrace.Event, key string) string {
	for _, attr := range event.Attributes {
		if string(attr.Key) != key {
			continue
		}
		return attr.Value.AsString()
	}
	return ""
}
