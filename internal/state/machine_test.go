package state

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestTransitionEnforcesAllowedStateMachines(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		entity   EntityType
		entityID string
		sequence [][2]string
		stateKey string
	}{
		{
			name:     "commission lifecycle",
			entity:   EntityCommission,
			entityID: "COMM-1",
			sequence: [][2]string{
				{CommissionPlanning, CommissionApproved},
				{CommissionApproved, CommissionExecuting},
				{CommissionExecuting, CommissionCompleted},
			},
			stateKey: "commission_state",
		},
		{
			name:     "mission lifecycle",
			entity:   EntityMission,
			entityID: "MISSION-1",
			sequence: [][2]string{
				{MissionBacklog, MissionInProgress},
				{MissionInProgress, MissionReview},
				{MissionReview, MissionApproved},
				{MissionApproved, MissionDone},
			},
			stateKey: "mission_state",
		},
		{
			name:     "ac lifecycle",
			entity:   EntityAC,
			entityID: "AC-1",
			sequence: [][2]string{
				{ACRed, ACVerifyRed},
				{ACVerifyRed, ACGreen},
				{ACGreen, ACVerifyGreen},
				{ACVerifyGreen, ACRefactor},
				{ACRefactor, ACVerifyRefactor},
				{ACVerifyRefactor, ACComplete},
			},
			stateKey: "ac_state",
		},
		{
			name:     "agent lifecycle",
			entity:   EntityAgent,
			entityID: "agent-1",
			sequence: [][2]string{
				{AgentIdle, AgentSpawning},
				{AgentSpawning, AgentRunning},
				{AgentRunning, AgentStuck},
				{AgentStuck, AgentDone},
			},
			stateKey: "agent_state",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			persister := &fakePersister{}
			machine, err := NewMachine(persister, "commander")
			if err != nil {
				t.Fatalf("new machine: %v", err)
			}

			for _, step := range tt.sequence {
				err := machine.Transition(context.Background(), tt.entity, tt.entityID, step[0], step[1], "transition")
				if err != nil {
					t.Fatalf("transition %s -> %s: %v", step[0], step[1], err)
				}
			}

			if len(persister.setStateCalls) != len(tt.sequence) {
				t.Fatalf("set-state calls = %d, want %d", len(persister.setStateCalls), len(tt.sequence))
			}
			for _, call := range persister.setStateCalls {
				if call.key != tt.stateKey {
					t.Fatalf("state key = %q, want %q", call.key, tt.stateKey)
				}
			}
		})
	}
}

func TestTransitionRejectsIllegalTransitionWithTypedError(t *testing.T) {
	t.Parallel()

	persister := &fakePersister{}
	machine, err := NewMachine(persister, "commander")
	if err != nil {
		t.Fatalf("new machine: %v", err)
	}

	err = machine.Transition(
		context.Background(),
		EntityMission,
		"MISSION-42",
		MissionBacklog,
		MissionDone,
		"skip stages",
	)
	if err == nil {
		t.Fatal("expected illegal transition error, got nil")
	}

	var illegalErr *IllegalTransitionError
	if !errors.As(err, &illegalErr) {
		t.Fatalf("error = %T, want *IllegalTransitionError", err)
	}
	if !errors.Is(err, &IllegalTransitionError{}) {
		t.Fatalf("errors.Is(%v, IllegalTransitionError{}) = false, want true", err)
	}
	if illegalErr.EntityType != EntityMission {
		t.Fatalf("entity type = %s, want %s", illegalErr.EntityType, EntityMission)
	}
	if illegalErr.EntityID != "MISSION-42" {
		t.Fatalf("entity id = %s, want MISSION-42", illegalErr.EntityID)
	}
	if illegalErr.FromState != MissionBacklog || illegalErr.ToState != MissionDone {
		t.Fatalf("illegal transition = %s -> %s", illegalErr.FromState, illegalErr.ToState)
	}
	if illegalErr.Reason != "illegal transition for entity lifecycle" {
		t.Fatalf("reason = %q, want lifecycle reason", illegalErr.Reason)
	}
	if !strings.Contains(err.Error(), "illegal transition for entity lifecycle") {
		t.Fatalf("error text missing reason: %v", err)
	}
}

func TestTransitionRecordsTimestampActorAndReason(t *testing.T) {
	t.Parallel()

	persister := &fakePersister{}
	machine, err := NewMachine(persister, "captain")
	if err != nil {
		t.Fatalf("new machine: %v", err)
	}

	fixed := time.Date(2026, 2, 11, 5, 0, 0, 0, time.UTC)
	machine.now = func() time.Time { return fixed }

	if err := machine.Transition(
		context.Background(),
		EntityCommission,
		"COMM-1",
		CommissionPlanning,
		CommissionApproved,
		"manifest complete",
	); err != nil {
		t.Fatalf("transition: %v", err)
	}

	history := machine.History()
	if len(history) != 1 {
		t.Fatalf("history length = %d, want 1", len(history))
	}

	record := history[0]
	if record.Actor != "captain" {
		t.Fatalf("actor = %q, want %q", record.Actor, "captain")
	}
	if record.Timestamp != fixed {
		t.Fatalf("timestamp = %s, want %s", record.Timestamp, fixed)
	}
	if record.Reason != "manifest complete" {
		t.Fatalf("reason = %q, want %q", record.Reason, "manifest complete")
	}
}

func TestTransitionPersistsSetStateAndComment(t *testing.T) {
	t.Parallel()

	persister := &fakePersister{}
	machine, err := NewMachine(persister, "commander")
	if err != nil {
		t.Fatalf("new machine: %v", err)
	}

	if err := machine.Transition(
		context.Background(),
		EntityAC,
		"MISSION-1",
		ACRed,
		ACVerifyRed,
		"red claim verified",
	); err != nil {
		t.Fatalf("transition: %v", err)
	}

	if len(persister.setStateCalls) != 1 {
		t.Fatalf("set-state calls = %d, want 1", len(persister.setStateCalls))
	}
	call := persister.setStateCalls[0]
	if call.id != "MISSION-1" || call.key != "ac_state" || call.value != ACVerifyRed {
		t.Fatalf("unexpected set-state call: %+v", call)
	}

	if len(persister.comments) != 1 {
		t.Fatalf("comments = %d, want 1", len(persister.comments))
	}
	comment := persister.comments[0]
	if !strings.Contains(comment.text, "actor=commander") {
		t.Fatalf("comment missing actor field: %q", comment.text)
	}
	if !strings.Contains(comment.text, `reason="red claim verified"`) {
		t.Fatalf("comment missing reason field: %q", comment.text)
	}
}

func TestTransitionWrapsSetStateAndCommentErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		persister *fakePersister
		wantText  string
	}{
		{
			name: "set-state error",
			persister: &fakePersister{
				setStateErr: errors.New("set-state failed"),
			},
			wantText: "persist state transition",
		},
		{
			name: "comment error",
			persister: &fakePersister{
				addCommentErr: errors.New("comment failed"),
			},
			wantText: "persist transition event",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			machine, err := NewMachine(tt.persister, "commander")
			if err != nil {
				t.Fatalf("new machine: %v", err)
			}

			err = machine.Transition(
				context.Background(),
				EntityCommission,
				"COMM-1",
				CommissionPlanning,
				CommissionApproved,
				"transition",
			)
			if err == nil {
				t.Fatal("expected wrapped persistence error")
			}
			if !strings.Contains(err.Error(), tt.wantText) {
				t.Fatalf("error %q missing %q", err.Error(), tt.wantText)
			}
		})
	}
}

func TestTransitionCreatesSpanWithRequiredAttributes(t *testing.T) {
	t.Parallel()

	spanRecorder := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(spanRecorder))
	t.Cleanup(func() {
		if err := provider.Shutdown(context.Background()); err != nil {
			t.Errorf("shutdown tracer provider: %v", err)
		}
	})

	persister := &fakePersister{}
	machine, err := NewMachine(persister, "commander", WithTracer(provider.Tracer("state-test")))
	if err != nil {
		t.Fatalf("new machine: %v", err)
	}

	if err := machine.Transition(
		context.Background(),
		EntityCommission,
		"COMM-7",
		CommissionPlanning,
		CommissionApproved,
		"admiral approved",
	); err != nil {
		t.Fatalf("transition: %v", err)
	}

	span := findTransitionSpan(t, spanRecorder.Ended())
	attrs := attributesToMap(span.Attributes())

	if span.Name() != "state.transition" {
		t.Fatalf("span name = %q, want %q", span.Name(), "state.transition")
	}
	if got := attrs["entity_type"]; got != string(EntityCommission) {
		t.Fatalf("entity_type = %q, want %q", got, string(EntityCommission))
	}
	if got := attrs["entity_id"]; got != "COMM-7" {
		t.Fatalf("entity_id = %q, want %q", got, "COMM-7")
	}
	if got := attrs["from_state"]; got != CommissionPlanning {
		t.Fatalf("from_state = %q, want %q", got, CommissionPlanning)
	}
	if got := attrs["to_state"]; got != CommissionApproved {
		t.Fatalf("to_state = %q, want %q", got, CommissionApproved)
	}
	if got := attrs["reason"]; got != "admiral approved" {
		t.Fatalf("reason = %q, want %q", got, "admiral approved")
	}
	if _, ok := attrs["duration_ms"]; !ok {
		t.Fatal("duration_ms attribute missing")
	}
}

func TestTransitionRecordsErrorsAndUsesParentContext(t *testing.T) {
	t.Parallel()

	spanRecorder := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(spanRecorder))
	t.Cleanup(func() {
		if err := provider.Shutdown(context.Background()); err != nil {
			t.Errorf("shutdown tracer provider: %v", err)
		}
	})

	tracer := provider.Tracer("state-test")
	persister := &fakePersister{setStateErr: errors.New("store failed")}
	machine, err := NewMachine(persister, "commander", WithTracer(tracer))
	if err != nil {
		t.Fatalf("new machine: %v", err)
	}

	parentCtx, parentSpan := tracer.Start(context.Background(), "parent")
	err = machine.Transition(
		parentCtx,
		EntityCommission,
		"COMM-9",
		CommissionPlanning,
		CommissionApproved,
		"persist failure",
	)
	parentSpan.End()

	if err == nil {
		t.Fatal("expected transition error, got nil")
	}

	transitionSpan := findTransitionSpan(t, spanRecorder.Ended())
	if transitionSpan.Parent().SpanID() != parentSpan.SpanContext().SpanID() {
		t.Fatalf(
			"transition span parent = %s, want %s",
			transitionSpan.Parent().SpanID(),
			parentSpan.SpanContext().SpanID(),
		)
	}
	if transitionSpan.Status().Code != codes.Error {
		t.Fatalf("status code = %v, want %v", transitionSpan.Status().Code, codes.Error)
	}
	if len(transitionSpan.Events()) == 0 {
		t.Fatal("expected at least one event recorded on error span")
	}
}

func findTransitionSpan(t *testing.T, spans []sdktrace.ReadOnlySpan) sdktrace.ReadOnlySpan {
	t.Helper()
	for _, span := range spans {
		if span.Name() == "state.transition" {
			return span
		}
	}
	t.Fatalf("state.transition span not found in %d spans", len(spans))
	return nil
}

func attributesToMap(attrs []attribute.KeyValue) map[string]string {
	out := make(map[string]string, len(attrs))
	for _, attr := range attrs {
		out[string(attr.Key)] = attr.Value.Emit()
	}
	return out
}

type setStateCall struct {
	id    string
	key   string
	value string
}

type commentCall struct {
	id   string
	text string
}

type fakePersister struct {
	setStateCalls []setStateCall
	comments      []commentCall
	setStateErr   error
	addCommentErr error
}

func (f *fakePersister) SetState(id, key, value string) error {
	if f.setStateErr != nil {
		return fmt.Errorf("set-state: %w", f.setStateErr)
	}
	f.setStateCalls = append(f.setStateCalls, setStateCall{id: id, key: key, value: value})
	return nil
}

func (f *fakePersister) AddComment(id, comment string) error {
	if f.addCommentErr != nil {
		return fmt.Errorf("add-comment: %w", f.addCommentErr)
	}
	f.comments = append(f.comments, commentCall{id: id, text: comment})
	return nil
}
