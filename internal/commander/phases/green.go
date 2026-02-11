package phases

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

const (
	// EventGreenComplete is the protocol claim emitted when GREEN work is ready for verification.
	EventGreenComplete = "GREEN_COMPLETE"
	// GateTypeVerifyGreen is the deterministic gate used to validate GREEN phase output.
	GateTypeVerifyGreen = "VERIFY_GREEN"
	// PhaseGreen is the AC phase used while iterating on implementation failures.
	PhaseGreen = "green"
	// PhaseRefactor is the next AC phase after a successful GREEN verification.
	PhaseRefactor    = "refactor"
	greenInstruction = "GREEN: implement minimal code to pass this test"
)

// GreenInput holds the data required to run a GREEN attempt for one acceptance criterion.
type GreenInput struct {
	MissionID string
	ACID      string
	TestFile  string
	RedOutput string
}

// GreenOutcome captures the final state of GREEN execution.
type GreenOutcome struct {
	NextPhase   string
	Attempts    int
	LastFailure string
}

// DispatchRequest provides the dispatch payload for GREEN attempts.
type DispatchRequest struct {
	MissionID   string
	ACID        string
	Attempt     int
	Instruction string
	TestFile    string
	RedOutput   string
	Feedback    string
}

// Dispatcher sends GREEN implementation instructions to the implementer agent.
type Dispatcher interface {
	DispatchGreen(ctx context.Context, req DispatchRequest) error
}

// ClaimWaiter waits for protocol claims from the agent execution loop.
type ClaimWaiter interface {
	WaitFor(ctx context.Context, missionID, acID, eventType string) error
}

// GateRequest is a deterministic verification gate invocation request.
type GateRequest struct {
	Type          string
	MissionID     string
	ACID          string
	FullTestSuite bool
}

// GateResult is the deterministic gate output summary.
type GateResult struct {
	ExitCode int
	Output   string
}

// GateRunner executes verification gates independently of implementer claims.
type GateRunner interface {
	Run(ctx context.Context, req GateRequest) (GateResult, error)
}

// PhaseStateStore persists AC phase transitions.
type PhaseStateStore interface {
	SetACPhase(ctx context.Context, missionID, acID, phase string) error
}

// GreenRunner executes GREEN attempts until verification passes or attempts are exhausted.
type GreenRunner struct {
	dispatcher  Dispatcher
	waiter      ClaimWaiter
	gates       GateRunner
	stateStore  PhaseStateStore
	maxAttempts int
}

// NewGreenRunner creates a GREEN phase runner.
func NewGreenRunner(
	dispatcher Dispatcher,
	waiter ClaimWaiter,
	gates GateRunner,
	stateStore PhaseStateStore,
	maxAttempts int,
) (*GreenRunner, error) {
	if dispatcher == nil {
		return nil, errors.New("dispatcher is required")
	}
	if waiter == nil {
		return nil, errors.New("claim waiter is required")
	}
	if gates == nil {
		return nil, errors.New("gate runner is required")
	}
	if stateStore == nil {
		return nil, errors.New("phase state store is required")
	}
	if maxAttempts <= 0 {
		return nil, errors.New("max attempts must be positive")
	}

	return &GreenRunner{
		dispatcher:  dispatcher,
		waiter:      waiter,
		gates:       gates,
		stateStore:  stateStore,
		maxAttempts: maxAttempts,
	}, nil
}

// Run executes GREEN dispatch/wait/verify loops for a single acceptance criterion.
func (r *GreenRunner) Run(ctx context.Context, input GreenInput) (GreenOutcome, error) {
	if strings.TrimSpace(input.MissionID) == "" {
		return GreenOutcome{}, errors.New("mission id must not be empty")
	}
	if strings.TrimSpace(input.ACID) == "" {
		return GreenOutcome{}, errors.New("acceptance criterion id must not be empty")
	}

	feedback := ""
	for attempt := 1; attempt <= r.maxAttempts; attempt++ {
		if err := r.dispatcher.DispatchGreen(ctx, DispatchRequest{
			MissionID:   input.MissionID,
			ACID:        input.ACID,
			Attempt:     attempt,
			Instruction: greenInstruction,
			TestFile:    input.TestFile,
			RedOutput:   input.RedOutput,
			Feedback:    feedback,
		}); err != nil {
			return GreenOutcome{}, fmt.Errorf("dispatch GREEN attempt %d: %w", attempt, err)
		}

		if err := r.waiter.WaitFor(ctx, input.MissionID, input.ACID, EventGreenComplete); err != nil {
			return GreenOutcome{}, fmt.Errorf("wait for %s attempt %d: %w", EventGreenComplete, attempt, err)
		}

		result, err := r.gates.Run(ctx, GateRequest{
			Type:          GateTypeVerifyGreen,
			MissionID:     input.MissionID,
			ACID:          input.ACID,
			FullTestSuite: true,
		})
		if err != nil {
			return GreenOutcome{}, fmt.Errorf("run %s attempt %d: %w", GateTypeVerifyGreen, attempt, err)
		}

		if result.ExitCode == 0 {
			if err := r.stateStore.SetACPhase(ctx, input.MissionID, input.ACID, PhaseRefactor); err != nil {
				return GreenOutcome{}, fmt.Errorf("set AC phase to %s: %w", PhaseRefactor, err)
			}
			return GreenOutcome{
				NextPhase: PhaseRefactor,
				Attempts:  attempt,
			}, nil
		}

		feedback = strings.TrimSpace(result.Output)
		if err := r.stateStore.SetACPhase(ctx, input.MissionID, input.ACID, PhaseGreen); err != nil {
			return GreenOutcome{}, fmt.Errorf("set AC phase to %s: %w", PhaseGreen, err)
		}
	}

	message := feedback
	if message == "" {
		message = "no gate output"
	}
	return GreenOutcome{
		NextPhase:   PhaseGreen,
		Attempts:    r.maxAttempts,
		LastFailure: feedback,
	}, fmt.Errorf("%s failed after %d attempts: %s", GateTypeVerifyGreen, r.maxAttempts, message)
}
