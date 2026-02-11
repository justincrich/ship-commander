package phases

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestRefactorRunnerRun(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		gateExitCode     int
		gateOutput       string
		wantErr          bool
		wantNextPhase    string
		wantRejected     bool
		wantFailure      string
		wantCompleteCall int
		wantRejectCall   int
	}{
		{
			name:             "passes and marks AC complete",
			gateExitCode:     0,
			wantNextPhase:    PhaseComplete,
			wantRejected:     false,
			wantCompleteCall: 1,
			wantRejectCall:   0,
		},
		{
			name:             "fails and rejects refactor",
			gateExitCode:     1,
			gateOutput:       "test suite regression",
			wantErr:          true,
			wantNextPhase:    PhaseRefactor,
			wantRejected:     true,
			wantFailure:      "test suite regression",
			wantCompleteCall: 0,
			wantRejectCall:   1,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dispatcher := &fakeRefactorDispatcher{}
			waiter := &fakeWaiter{}
			gates := &fakeGates{responses: []gateCall{{exitCode: tt.gateExitCode, output: tt.gateOutput}}}
			completion := &fakeCompletionStore{}
			rejections := &fakeRejectionStore{}

			runner, err := NewRefactorRunner(dispatcher, waiter, gates, completion, rejections)
			if err != nil {
				t.Fatalf("new refactor runner: %v", err)
			}

			outcome, err := runner.Run(context.Background(), RefactorInput{
				MissionID: "mission-1",
				ACID:      "AC-1",
			})
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
			} else if err != nil {
				t.Fatalf("run refactor: %v", err)
			}

			if len(dispatcher.calls) != 1 {
				t.Fatalf("dispatch calls = %d, want 1", len(dispatcher.calls))
			}
			if dispatcher.calls[0].Instruction != refactorInstruction {
				t.Fatalf("instruction = %q, want %q", dispatcher.calls[0].Instruction, refactorInstruction)
			}

			if len(waiter.calls) != 1 {
				t.Fatalf("wait calls = %d, want 1", len(waiter.calls))
			}
			if waiter.calls[0].eventType != EventRefactorComplete {
				t.Fatalf("wait event = %q, want %q", waiter.calls[0].eventType, EventRefactorComplete)
			}

			if len(gates.requests) != 1 {
				t.Fatalf("gate requests = %d, want 1", len(gates.requests))
			}
			req := gates.requests[0]
			if req.Type != GateTypeVerifyRefactor {
				t.Fatalf("gate type = %q, want %q", req.Type, GateTypeVerifyRefactor)
			}
			if !req.FullTestSuite {
				t.Fatal("expected FullTestSuite=true for VERIFY_REFACTOR")
			}

			if outcome.NextPhase != tt.wantNextPhase {
				t.Fatalf("next phase = %q, want %q", outcome.NextPhase, tt.wantNextPhase)
			}
			if outcome.Rejected != tt.wantRejected {
				t.Fatalf("rejected = %v, want %v", outcome.Rejected, tt.wantRejected)
			}
			if outcome.FailureOutput != tt.wantFailure {
				t.Fatalf("failure output = %q, want %q", outcome.FailureOutput, tt.wantFailure)
			}

			if len(completion.calls) != tt.wantCompleteCall {
				t.Fatalf("completion calls = %d, want %d", len(completion.calls), tt.wantCompleteCall)
			}
			if len(rejections.calls) != tt.wantRejectCall {
				t.Fatalf("rejection calls = %d, want %d", len(rejections.calls), tt.wantRejectCall)
			}
		})
	}
}

func TestRefactorRunnerValidation(t *testing.T) {
	t.Parallel()

	dispatcher := &fakeRefactorDispatcher{}
	waiter := &fakeWaiter{}
	gates := &fakeGates{responses: []gateCall{{exitCode: 0}}}
	completion := &fakeCompletionStore{}
	rejections := &fakeRejectionStore{}

	tests := []struct {
		name       string
		dispatcher RefactorDispatcher
		waiter     ClaimWaiter
		gates      GateRunner
		completion ACCompletionStore
		rejections RefactorRejectionStore
	}{
		{name: "missing dispatcher", waiter: waiter, gates: gates, completion: completion, rejections: rejections},
		{name: "missing waiter", dispatcher: dispatcher, gates: gates, completion: completion, rejections: rejections},
		{name: "missing gates", dispatcher: dispatcher, waiter: waiter, completion: completion, rejections: rejections},
		{name: "missing completion", dispatcher: dispatcher, waiter: waiter, gates: gates, rejections: rejections},
		{name: "missing rejection store", dispatcher: dispatcher, waiter: waiter, gates: gates, completion: completion},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := NewRefactorRunner(tt.dispatcher, tt.waiter, tt.gates, tt.completion, tt.rejections)
			if err == nil {
				t.Fatal("expected constructor validation error")
			}
		})
	}
}

func TestRefactorRunnerInputValidation(t *testing.T) {
	t.Parallel()

	runner, err := NewRefactorRunner(
		&fakeRefactorDispatcher{},
		&fakeWaiter{},
		&fakeGates{responses: []gateCall{{exitCode: 0}}},
		&fakeCompletionStore{},
		&fakeRejectionStore{},
	)
	if err != nil {
		t.Fatalf("new refactor runner: %v", err)
	}

	_, err = runner.Run(context.Background(), RefactorInput{ACID: "AC-1"})
	if err == nil || !strings.Contains(err.Error(), "mission id") {
		t.Fatalf("expected mission id validation error, got %v", err)
	}

	_, err = runner.Run(context.Background(), RefactorInput{MissionID: "mission-1"})
	if err == nil || !strings.Contains(err.Error(), "acceptance criterion id") {
		t.Fatalf("expected AC id validation error, got %v", err)
	}
}

type fakeRefactorDispatcher struct {
	calls []RefactorDispatchRequest
	err   error
}

func (f *fakeRefactorDispatcher) DispatchRefactor(_ context.Context, req RefactorDispatchRequest) error {
	if f.err != nil {
		return f.err
	}
	f.calls = append(f.calls, req)
	return nil
}

type completionCall struct {
	missionID string
	acID      string
	reason    string
}

type fakeCompletionStore struct {
	calls []completionCall
	err   error
}

func (f *fakeCompletionStore) MarkACComplete(_ context.Context, missionID, acID, reason string) error {
	if f.err != nil {
		return f.err
	}
	f.calls = append(f.calls, completionCall{
		missionID: missionID,
		acID:      acID,
		reason:    reason,
	})
	return nil
}

type rejectionCall struct {
	missionID string
	acID      string
	output    string
}

type fakeRejectionStore struct {
	calls []rejectionCall
	err   error
}

func (f *fakeRejectionStore) RejectRefactor(_ context.Context, missionID, acID, output string) error {
	if f.err != nil {
		return f.err
	}
	f.calls = append(f.calls, rejectionCall{
		missionID: missionID,
		acID:      acID,
		output:    output,
	})
	return nil
}

func TestRefactorRunnerWrapsRejectionStoreError(t *testing.T) {
	t.Parallel()

	runner, err := NewRefactorRunner(
		&fakeRefactorDispatcher{},
		&fakeWaiter{},
		&fakeGates{responses: []gateCall{{exitCode: 1, output: "failed"}}},
		&fakeCompletionStore{},
		&fakeRejectionStore{err: errors.New("reject store failed")},
	)
	if err != nil {
		t.Fatalf("new refactor runner: %v", err)
	}

	_, err = runner.Run(context.Background(), RefactorInput{
		MissionID: "mission-1",
		ACID:      "AC-1",
	})
	if err == nil {
		t.Fatal("expected rejection store error")
	}
	if !strings.Contains(err.Error(), "reject refactor attempt") {
		t.Fatalf("error = %v, want reject refactor context", err)
	}
}
