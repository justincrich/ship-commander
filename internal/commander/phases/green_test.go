package phases

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestGreenRunnerRun(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		maxAttempts      int
		gateCalls        []gateCall
		wantErr          bool
		wantAttempts     int
		wantNextPhase    string
		wantLastFailure  string
		wantDispatches   int
		wantStatePhases  []string
		wantLastFeedback string
	}{
		{
			name:            "passes on first attempt and advances to refactor",
			maxAttempts:     3,
			gateCalls:       []gateCall{{exitCode: 0}},
			wantAttempts:    1,
			wantNextPhase:   PhaseRefactor,
			wantDispatches:  1,
			wantStatePhases: []string{PhaseRefactor},
		},
		{
			name:             "fails then retries with feedback and passes",
			maxAttempts:      3,
			gateCalls:        []gateCall{{exitCode: 1, output: "first failure"}, {exitCode: 0}},
			wantAttempts:     2,
			wantNextPhase:    PhaseRefactor,
			wantDispatches:   2,
			wantStatePhases:  []string{PhaseGreen, PhaseRefactor},
			wantLastFeedback: "first failure",
		},
		{
			name:             "fails until attempts exhausted",
			maxAttempts:      2,
			gateCalls:        []gateCall{{exitCode: 1, output: "compile failed"}, {exitCode: 1, output: "still failing"}},
			wantErr:          true,
			wantAttempts:     2,
			wantNextPhase:    PhaseGreen,
			wantLastFailure:  "still failing",
			wantDispatches:   2,
			wantStatePhases:  []string{PhaseGreen, PhaseGreen},
			wantLastFeedback: "compile failed",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dispatcher := &fakeDispatcher{}
			waiter := &fakeWaiter{}
			gates := &fakeGates{responses: tt.gateCalls}
			store := &fakeStateStore{}

			runner, err := NewGreenRunner(dispatcher, waiter, gates, store, tt.maxAttempts)
			if err != nil {
				t.Fatalf("new runner: %v", err)
			}

			outcome, err := runner.Run(context.Background(), GreenInput{
				MissionID: "mission-1",
				ACID:      "AC-1",
				TestFile:  "internal/commander/phases/green_test.go",
				RedOutput: "expected failing test",
			})

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
			} else if err != nil {
				t.Fatalf("run green: %v", err)
			}

			if outcome.Attempts != tt.wantAttempts {
				t.Fatalf("attempts = %d, want %d", outcome.Attempts, tt.wantAttempts)
			}
			if outcome.NextPhase != tt.wantNextPhase {
				t.Fatalf("next phase = %q, want %q", outcome.NextPhase, tt.wantNextPhase)
			}
			if outcome.LastFailure != tt.wantLastFailure {
				t.Fatalf("last failure = %q, want %q", outcome.LastFailure, tt.wantLastFailure)
			}

			if len(dispatcher.calls) != tt.wantDispatches {
				t.Fatalf("dispatch calls = %d, want %d", len(dispatcher.calls), tt.wantDispatches)
			}
			firstDispatch := dispatcher.calls[0]
			if firstDispatch.Instruction != greenInstruction {
				t.Fatalf("instruction = %q, want %q", firstDispatch.Instruction, greenInstruction)
			}
			if firstDispatch.TestFile != "internal/commander/phases/green_test.go" {
				t.Fatalf("test file = %q, want green test path", firstDispatch.TestFile)
			}
			if firstDispatch.RedOutput != "expected failing test" {
				t.Fatalf("red output = %q, want %q", firstDispatch.RedOutput, "expected failing test")
			}
			if tt.wantLastFeedback != "" && len(dispatcher.calls) > 1 {
				secondDispatch := dispatcher.calls[1]
				if secondDispatch.Feedback != tt.wantLastFeedback {
					t.Fatalf("second dispatch feedback = %q, want %q", secondDispatch.Feedback, tt.wantLastFeedback)
				}
			}

			if len(waiter.calls) != len(dispatcher.calls) {
				t.Fatalf("wait calls = %d, want %d", len(waiter.calls), len(dispatcher.calls))
			}
			for _, call := range waiter.calls {
				if call.eventType != EventGreenComplete {
					t.Fatalf("wait event = %q, want %q", call.eventType, EventGreenComplete)
				}
			}

			for _, req := range gates.requests {
				if req.Type != GateTypeVerifyGreen {
					t.Fatalf("gate type = %q, want %q", req.Type, GateTypeVerifyGreen)
				}
				if !req.FullTestSuite {
					t.Fatal("expected FullTestSuite=true for VERIFY_GREEN")
				}
			}

			if !reflect.DeepEqual(store.phases, tt.wantStatePhases) {
				t.Fatalf("state phases = %v, want %v", store.phases, tt.wantStatePhases)
			}
		})
	}
}

func TestNewGreenRunnerValidation(t *testing.T) {
	t.Parallel()

	dispatcher := &fakeDispatcher{}
	waiter := &fakeWaiter{}
	gates := &fakeGates{}
	store := &fakeStateStore{}

	tests := []struct {
		name        string
		dispatcher  Dispatcher
		waiter      ClaimWaiter
		gates       GateRunner
		store       PhaseStateStore
		maxAttempts int
	}{
		{name: "missing dispatcher", waiter: waiter, gates: gates, store: store, maxAttempts: 1},
		{name: "missing waiter", dispatcher: dispatcher, gates: gates, store: store, maxAttempts: 1},
		{name: "missing gates", dispatcher: dispatcher, waiter: waiter, store: store, maxAttempts: 1},
		{name: "missing store", dispatcher: dispatcher, waiter: waiter, gates: gates, maxAttempts: 1},
		{name: "invalid max attempts", dispatcher: dispatcher, waiter: waiter, gates: gates, store: store, maxAttempts: 0},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := NewGreenRunner(tt.dispatcher, tt.waiter, tt.gates, tt.store, tt.maxAttempts)
			if err == nil {
				t.Fatal("expected validation error, got nil")
			}
		})
	}
}

func TestGreenRunnerRunInputValidation(t *testing.T) {
	t.Parallel()

	runner, err := NewGreenRunner(&fakeDispatcher{}, &fakeWaiter{}, &fakeGates{responses: []gateCall{{exitCode: 0}}}, &fakeStateStore{}, 1)
	if err != nil {
		t.Fatalf("new runner: %v", err)
	}

	_, err = runner.Run(context.Background(), GreenInput{ACID: "AC-1"})
	if err == nil || !strings.Contains(err.Error(), "mission id") {
		t.Fatalf("expected mission id validation error, got %v", err)
	}

	_, err = runner.Run(context.Background(), GreenInput{MissionID: "mission-1"})
	if err == nil || !strings.Contains(err.Error(), "acceptance criterion id") {
		t.Fatalf("expected AC id validation error, got %v", err)
	}
}

type gateCall struct {
	exitCode int
	output   string
}

type fakeDispatcher struct {
	calls []DispatchRequest
	err   error
}

func (f *fakeDispatcher) DispatchGreen(_ context.Context, req DispatchRequest) error {
	if f.err != nil {
		return f.err
	}
	f.calls = append(f.calls, req)
	return nil
}

type waitCall struct {
	missionID string
	acID      string
	eventType string
}

type fakeWaiter struct {
	calls []waitCall
	err   error
}

func (f *fakeWaiter) WaitFor(_ context.Context, missionID, acID, eventType string) error {
	if f.err != nil {
		return f.err
	}
	f.calls = append(f.calls, waitCall{
		missionID: missionID,
		acID:      acID,
		eventType: eventType,
	})
	return nil
}

type fakeGates struct {
	requests  []GateRequest
	responses []gateCall
	err       error
	callIndex int
}

func (f *fakeGates) Run(_ context.Context, req GateRequest) (GateResult, error) {
	if f.err != nil {
		return GateResult{}, f.err
	}
	f.requests = append(f.requests, req)
	if len(f.responses) == 0 {
		return GateResult{}, errors.New("no gate response configured")
	}
	idx := f.callIndex
	if idx >= len(f.responses) {
		idx = len(f.responses) - 1
	}
	f.callIndex++
	response := f.responses[idx]
	return GateResult{
		ExitCode: response.exitCode,
		Output:   response.output,
	}, nil
}

type fakeStateStore struct {
	phases []string
	err    error
}

func (f *fakeStateStore) SetACPhase(_ context.Context, _, _ string, phase string) error {
	if f.err != nil {
		return f.err
	}
	f.phases = append(f.phases, phase)
	return nil
}
