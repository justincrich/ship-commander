package phases

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

const (
	// EventRefactorComplete is emitted when the implementer completes REFACTOR work.
	EventRefactorComplete = "REFACTOR_COMPLETE"
	// GateTypeVerifyRefactor is the deterministic gate for REFACTOR validation.
	GateTypeVerifyRefactor = "VERIFY_REFACTOR"
	// PhaseComplete is the terminal AC phase after REFACTOR verification passes.
	PhaseComplete       = "complete"
	refactorInstruction = "REFACTOR: clean up implementation without changing test behavior"
)

// RefactorInput is the required runtime input for one REFACTOR attempt.
type RefactorInput struct {
	MissionID string
	ACID      string
}

// RefactorDispatchRequest contains the payload sent to the implementer.
type RefactorDispatchRequest struct {
	MissionID   string
	ACID        string
	Instruction string
}

// RefactorOutcome summarizes REFACTOR execution results.
type RefactorOutcome struct {
	NextPhase     string
	Rejected      bool
	FailureOutput string
}

// RefactorDispatcher sends REFACTOR instructions to the implementer.
type RefactorDispatcher interface {
	DispatchRefactor(ctx context.Context, req RefactorDispatchRequest) error
}

// ACCompletionStore records successful AC completion transitions.
type ACCompletionStore interface {
	MarkACComplete(ctx context.Context, missionID, acID, reason string) error
}

// RefactorRejectionStore records rejected REFACTOR attempts.
type RefactorRejectionStore interface {
	RejectRefactor(ctx context.Context, missionID, acID, output string) error
}

// RefactorRunner executes REFACTOR dispatch -> wait -> verify -> state handling.
type RefactorRunner struct {
	dispatcher RefactorDispatcher
	waiter     ClaimWaiter
	gates      GateRunner
	completion ACCompletionStore
	rejections RefactorRejectionStore
}

// NewRefactorRunner creates a REFACTOR phase runner.
func NewRefactorRunner(
	dispatcher RefactorDispatcher,
	waiter ClaimWaiter,
	gates GateRunner,
	completion ACCompletionStore,
	rejections RefactorRejectionStore,
) (*RefactorRunner, error) {
	if dispatcher == nil {
		return nil, errors.New("dispatcher is required")
	}
	if waiter == nil {
		return nil, errors.New("claim waiter is required")
	}
	if gates == nil {
		return nil, errors.New("gate runner is required")
	}
	if completion == nil {
		return nil, errors.New("completion store is required")
	}
	if rejections == nil {
		return nil, errors.New("rejection store is required")
	}

	return &RefactorRunner{
		dispatcher: dispatcher,
		waiter:     waiter,
		gates:      gates,
		completion: completion,
		rejections: rejections,
	}, nil
}

// Run executes one REFACTOR attempt and returns pass/fail outcome.
func (r *RefactorRunner) Run(ctx context.Context, input RefactorInput) (RefactorOutcome, error) {
	if strings.TrimSpace(input.MissionID) == "" {
		return RefactorOutcome{}, errors.New("mission id must not be empty")
	}
	if strings.TrimSpace(input.ACID) == "" {
		return RefactorOutcome{}, errors.New("acceptance criterion id must not be empty")
	}

	if err := r.dispatcher.DispatchRefactor(ctx, RefactorDispatchRequest{
		MissionID:   input.MissionID,
		ACID:        input.ACID,
		Instruction: refactorInstruction,
	}); err != nil {
		return RefactorOutcome{}, fmt.Errorf("dispatch REFACTOR: %w", err)
	}

	if err := r.waiter.WaitFor(ctx, input.MissionID, input.ACID, EventRefactorComplete); err != nil {
		return RefactorOutcome{}, fmt.Errorf("wait for %s: %w", EventRefactorComplete, err)
	}

	result, err := r.gates.Run(ctx, GateRequest{
		Type:          GateTypeVerifyRefactor,
		MissionID:     input.MissionID,
		ACID:          input.ACID,
		FullTestSuite: true,
	})
	if err != nil {
		return RefactorOutcome{}, fmt.Errorf("run %s: %w", GateTypeVerifyRefactor, err)
	}

	if result.ExitCode == 0 {
		if err := r.completion.MarkACComplete(ctx, input.MissionID, input.ACID, "VERIFY_REFACTOR accepted"); err != nil {
			return RefactorOutcome{}, fmt.Errorf("mark AC complete: %w", err)
		}
		return RefactorOutcome{
			NextPhase: PhaseComplete,
			Rejected:  false,
		}, nil
	}

	output := strings.TrimSpace(result.Output)
	if err := r.rejections.RejectRefactor(ctx, input.MissionID, input.ACID, output); err != nil {
		return RefactorOutcome{}, fmt.Errorf("reject refactor attempt: %w", err)
	}

	return RefactorOutcome{
		NextPhase:     PhaseRefactor,
		Rejected:      true,
		FailureOutput: output,
	}, fmt.Errorf("%s rejected: %s", GateTypeVerifyRefactor, fallbackOutput(output))
}

func fallbackOutput(output string) string {
	if strings.TrimSpace(output) == "" {
		return "no gate output"
	}
	return output
}
