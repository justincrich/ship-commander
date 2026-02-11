package phases

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestREDExecutorDispatchIncludesMissionAndACContext(t *testing.T) {
	t.Parallel()

	dispatcher := &fakeREDDispatcher{}
	waiter := &fakeREDWaiter{
		claim: REDClaim{EventType: REDCompleteEventType, MissionID: "m1", ACID: "ac-1"},
	}
	verifier := &fakeREDVerifier{
		result: VerifyREDResult{Classification: GateClassificationAccept},
	}
	transitions := &fakeTransitionStore{}
	failures := &fakeFailureStore{}
	feedback := &fakeFeedbackSender{}
	escalations := &fakeEscalationPublisher{}

	executor, err := NewREDExecutor(dispatcher, waiter, verifier, transitions, failures, feedback, escalations)
	if err != nil {
		t.Fatalf("new executor: %v", err)
	}

	input := REDExecutionInput{
		MissionID:     "m1",
		MissionSpec:   "full mission spec",
		ACID:          "ac-1",
		ACDescription: "must fail first",
		WorktreePath:  "/tmp/worktree/m1",
		Attempt:       0,
	}
	if _, err := executor.Execute(context.Background(), input); err != nil {
		t.Fatalf("execute: %v", err)
	}

	if len(dispatcher.requests) != 1 {
		t.Fatalf("dispatch requests = %d, want 1", len(dispatcher.requests))
	}
	req := dispatcher.requests[0]
	if req.MissionSpec != input.MissionSpec || req.ACID != input.ACID || req.ACDescription != input.ACDescription {
		t.Fatalf("dispatch request mismatch: %+v", req)
	}
	if req.Instruction != REDInstruction {
		t.Fatalf("instruction = %q, want %q", req.Instruction, REDInstruction)
	}
}

func TestREDExecutorPassTransitionsToGreen(t *testing.T) {
	t.Parallel()

	dispatcher := &fakeREDDispatcher{}
	waiter := &fakeREDWaiter{
		claim: REDClaim{EventType: REDCompleteEventType, MissionID: "m1", ACID: "ac-1"},
	}
	verifier := &fakeREDVerifier{
		result: VerifyREDResult{Classification: GateClassificationAccept, Output: "failing test observed"},
	}
	transitions := &fakeTransitionStore{}
	failures := &fakeFailureStore{}
	feedback := &fakeFeedbackSender{}
	escalations := &fakeEscalationPublisher{}

	executor, err := NewREDExecutor(dispatcher, waiter, verifier, transitions, failures, feedback, escalations)
	if err != nil {
		t.Fatalf("new executor: %v", err)
	}

	result, err := executor.Execute(context.Background(), REDExecutionInput{
		MissionID:     "m1",
		MissionSpec:   "spec",
		ACID:          "ac-1",
		ACDescription: "desc",
		WorktreePath:  "/tmp/worktree/m1",
		Attempt:       2,
	})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}

	if result.NextPhase != ACPhaseGreen {
		t.Fatalf("next phase = %q, want %q", result.NextPhase, ACPhaseGreen)
	}
	if result.Attempt != 2 {
		t.Fatalf("attempt = %d, want 2", result.Attempt)
	}
	if len(transitions.calls) != 1 {
		t.Fatalf("transitions = %d, want 1", len(transitions.calls))
	}
	if len(failures.items) != 0 {
		t.Fatalf("failures recorded = %d, want 0", len(failures.items))
	}
	if len(feedback.items) != 0 {
		t.Fatalf("feedback messages = %d, want 0", len(feedback.items))
	}
}

func TestREDExecutorRejectRecordsFailureAndIncrementsAttempt(t *testing.T) {
	t.Parallel()

	dispatcher := &fakeREDDispatcher{}
	waiter := &fakeREDWaiter{
		claim: REDClaim{EventType: REDCompleteEventType, MissionID: "m2", ACID: "ac-2"},
	}
	verifier := &fakeREDVerifier{
		result: VerifyREDResult{
			Classification: GateClassificationReject,
			Output:         "test did not fail as expected",
		},
	}
	transitions := &fakeTransitionStore{}
	failures := &fakeFailureStore{}
	feedback := &fakeFeedbackSender{}
	escalations := &fakeEscalationPublisher{}

	executor, err := NewREDExecutor(dispatcher, waiter, verifier, transitions, failures, feedback, escalations)
	if err != nil {
		t.Fatalf("new executor: %v", err)
	}

	result, err := executor.Execute(context.Background(), REDExecutionInput{
		MissionID:     "m2",
		MissionSpec:   "spec",
		ACID:          "ac-2",
		ACDescription: "desc",
		WorktreePath:  "/tmp/worktree/m2",
		Attempt:       3,
	})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}

	if result.NextPhase != ACPhaseRed {
		t.Fatalf("next phase = %q, want %q", result.NextPhase, ACPhaseRed)
	}
	if result.Attempt != 4 {
		t.Fatalf("attempt = %d, want 4", result.Attempt)
	}
	if len(failures.items) != 1 {
		t.Fatalf("failure records = %d, want 1", len(failures.items))
	}
	if failures.items[0].Attempt != 4 {
		t.Fatalf("failure attempt = %d, want 4", failures.items[0].Attempt)
	}
	if len(feedback.items) != 1 {
		t.Fatalf("feedback items = %d, want 1", len(feedback.items))
	}
	if !strings.Contains(feedback.items[0].Message, "attempt 4") {
		t.Fatalf("feedback message missing attempt: %q", feedback.items[0].Message)
	}
	if len(transitions.calls) != 0 {
		t.Fatalf("transition calls = %d, want 0", len(transitions.calls))
	}
}

func TestREDExecutorTimeoutTriggersEscalation(t *testing.T) {
	t.Parallel()

	dispatcher := &fakeREDDispatcher{}
	waiter := &fakeREDWaiter{
		waitForDone: true,
	}
	verifier := &fakeREDVerifier{}
	transitions := &fakeTransitionStore{}
	failures := &fakeFailureStore{}
	feedback := &fakeFeedbackSender{}
	escalations := &fakeEscalationPublisher{}

	executor, err := NewREDExecutor(dispatcher, waiter, verifier, transitions, failures, feedback, escalations)
	if err != nil {
		t.Fatalf("new executor: %v", err)
	}
	executor.now = func() time.Time { return time.Date(2026, 2, 11, 0, 0, 0, 0, time.UTC) }

	_, err = executor.Execute(context.Background(), REDExecutionInput{
		MissionID:     "m3",
		MissionSpec:   "spec",
		ACID:          "ac-3",
		ACDescription: "desc",
		WorktreePath:  "/tmp/worktree/m3",
		Attempt:       1,
		Timeout:       5 * time.Millisecond,
	})
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if !errors.Is(err, ErrREDClaimTimeout) {
		t.Fatalf("error = %v, want ErrREDClaimTimeout", err)
	}
	if len(escalations.items) != 1 {
		t.Fatalf("escalations = %d, want 1", len(escalations.items))
	}
	if escalations.items[0].Timeout != 5*time.Millisecond {
		t.Fatalf("escalation timeout = %s, want 5ms", escalations.items[0].Timeout)
	}
	if verifier.calls != 0 {
		t.Fatalf("verifier calls = %d, want 0 on timeout", verifier.calls)
	}
}

type fakeREDDispatcher struct {
	requests []REDDispatchRequest
	err      error
}

func (f *fakeREDDispatcher) DispatchRED(_ context.Context, req REDDispatchRequest) error {
	f.requests = append(f.requests, req)
	return f.err
}

type fakeREDWaiter struct {
	claim       REDClaim
	err         error
	waitForDone bool
	calls       int
	missionID   string
	acID        string
}

func (f *fakeREDWaiter) WaitREDComplete(ctx context.Context, missionID, acID string) (REDClaim, error) {
	f.calls++
	f.missionID = missionID
	f.acID = acID
	if f.waitForDone {
		<-ctx.Done()
		return REDClaim{}, ctx.Err()
	}
	return f.claim, f.err
}

type fakeREDVerifier struct {
	result VerifyREDResult
	err    error
	calls  int
}

func (f *fakeREDVerifier) VerifyRED(_ context.Context, _ VerifyREDRequest) (VerifyREDResult, error) {
	f.calls++
	return f.result, f.err
}

type transitionCall struct {
	missionID string
	acID      string
	reason    string
}

type fakeTransitionStore struct {
	calls []transitionCall
	err   error
}

func (f *fakeTransitionStore) TransitionToGreen(_ context.Context, missionID, acID, reason string) error {
	f.calls = append(f.calls, transitionCall{
		missionID: missionID,
		acID:      acID,
		reason:    reason,
	})
	return f.err
}

type fakeFailureStore struct {
	items []REDFailure
	err   error
}

func (f *fakeFailureStore) RecordREDFailure(_ context.Context, failure REDFailure) error {
	f.items = append(f.items, failure)
	return f.err
}

type fakeFeedbackSender struct {
	items []REDFeedback
	err   error
}

func (f *fakeFeedbackSender) SendREDFeedback(_ context.Context, feedback REDFeedback) error {
	f.items = append(f.items, feedback)
	return f.err
}

type fakeEscalationPublisher struct {
	items []REDStuckEscalation
	err   error
}

func (f *fakeEscalationPublisher) PublishREDStuck(_ context.Context, escalation REDStuckEscalation) error {
	f.items = append(f.items, escalation)
	return f.err
}
