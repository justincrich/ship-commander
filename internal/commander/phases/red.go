package phases

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	// REDCompleteEventType identifies the claim emitted by implementer completion.
	REDCompleteEventType = "RED_COMPLETE"
	// REDInstruction is the fixed instruction sent to implementers during RED phase.
	REDInstruction    = "RED: write failing test for this acceptance criterion"
	defaultREDTimeout = 5 * time.Minute
)

// ACPhase represents the current acceptance-criterion phase.
type ACPhase string

const (
	// ACPhaseRed is the RED phase.
	ACPhaseRed ACPhase = "red"
	// ACPhaseGreen is the GREEN phase entered after VERIFY_RED accepts.
	ACPhaseGreen ACPhase = "green"
)

// GateClassification describes VERIFY_RED outcome semantics.
type GateClassification string

const (
	// GateClassificationAccept means VERIFY_RED accepted the failing-test result.
	GateClassificationAccept GateClassification = "accept"
	// GateClassificationReject means VERIFY_RED rejected the RED attempt.
	GateClassificationReject GateClassification = "reject"
)

// ErrREDClaimTimeout indicates RED_COMPLETE was not received before timeout.
var ErrREDClaimTimeout = errors.New("timed out waiting for RED_COMPLETE")

// REDExecutionInput is the runtime input for one RED phase execution.
type REDExecutionInput struct {
	MissionID     string
	MissionSpec   string
	ACID          string
	ACDescription string
	WorktreePath  string
	Attempt       int
	Timeout       time.Duration
}

// REDDispatchRequest is sent to the implementer dispatcher.
type REDDispatchRequest struct {
	MissionID     string
	MissionSpec   string
	ACID          string
	ACDescription string
	WorktreePath  string
	Instruction   string
}

// REDClaim is the RED_COMPLETE protocol claim payload.
type REDClaim struct {
	EventType string
	MissionID string
	ACID      string
}

// VerifyREDRequest is used to run VERIFY_RED independently.
type VerifyREDRequest struct {
	MissionID    string
	ACID         string
	WorktreePath string
}

// VerifyREDResult captures gate verdict and useful output.
type VerifyREDResult struct {
	Classification GateClassification
	Output         string
}

// REDFailure is recorded whenever VERIFY_RED rejects.
type REDFailure struct {
	MissionID string
	ACID      string
	Attempt   int
	Reason    string
	Output    string
	Timestamp time.Time
}

// REDFeedback is sent back to implementer for the next retry.
type REDFeedback struct {
	MissionID string
	ACID      string
	Attempt   int
	Message   string
}

// REDStuckEscalation is emitted when RED claim wait times out.
type REDStuckEscalation struct {
	MissionID string
	ACID      string
	Attempt   int
	Timeout   time.Duration
	Timestamp time.Time
}

// REDExecutionResult summarizes phase state changes after execution.
type REDExecutionResult struct {
	NextPhase ACPhase
	Attempt   int
	Feedback  string
	Gate      VerifyREDResult
}

// REDDispatcher dispatches the implementer for RED work.
type REDDispatcher interface {
	DispatchRED(ctx context.Context, req REDDispatchRequest) error
}

// REDClaimWaiter waits for a RED_COMPLETE claim for mission + AC.
type REDClaimWaiter interface {
	WaitREDComplete(ctx context.Context, missionID, acID string) (REDClaim, error)
}

// REDGateVerifier runs VERIFY_RED independently.
type REDGateVerifier interface {
	VerifyRED(ctx context.Context, req VerifyREDRequest) (VerifyREDResult, error)
}

// ACTransitionStore persists AC phase transitions.
type ACTransitionStore interface {
	TransitionToGreen(ctx context.Context, missionID, acID, reason string) error
}

// REDFailureStore persists rejected RED attempts.
type REDFailureStore interface {
	RecordREDFailure(ctx context.Context, failure REDFailure) error
}

// REDFeedbackSender sends retry feedback to the implementer.
type REDFeedbackSender interface {
	SendREDFeedback(ctx context.Context, feedback REDFeedback) error
}

// REDEscalationPublisher publishes stuck-escalation events.
type REDEscalationPublisher interface {
	PublishREDStuck(ctx context.Context, escalation REDStuckEscalation) error
}

// REDExecutor executes one deterministic RED phase loop.
type REDExecutor struct {
	dispatcher  REDDispatcher
	waiter      REDClaimWaiter
	verifier    REDGateVerifier
	transitions ACTransitionStore
	failures    REDFailureStore
	feedback    REDFeedbackSender
	escalations REDEscalationPublisher
	now         func() time.Time
}

// NewREDExecutor constructs a RED executor with required dependencies.
func NewREDExecutor(
	dispatcher REDDispatcher,
	waiter REDClaimWaiter,
	verifier REDGateVerifier,
	transitions ACTransitionStore,
	failures REDFailureStore,
	feedback REDFeedbackSender,
	escalations REDEscalationPublisher,
) (*REDExecutor, error) {
	if dispatcher == nil {
		return nil, errors.New("dispatcher is required")
	}
	if waiter == nil {
		return nil, errors.New("waiter is required")
	}
	if verifier == nil {
		return nil, errors.New("verifier is required")
	}
	if transitions == nil {
		return nil, errors.New("transition store is required")
	}
	if failures == nil {
		return nil, errors.New("failure store is required")
	}
	if feedback == nil {
		return nil, errors.New("feedback sender is required")
	}
	if escalations == nil {
		return nil, errors.New("escalation publisher is required")
	}

	return &REDExecutor{
		dispatcher:  dispatcher,
		waiter:      waiter,
		verifier:    verifier,
		transitions: transitions,
		failures:    failures,
		feedback:    feedback,
		escalations: escalations,
		now:         time.Now,
	}, nil
}

// Execute runs RED dispatch -> claim wait -> VERIFY_RED -> state update.
func (e *REDExecutor) Execute(ctx context.Context, input REDExecutionInput) (REDExecutionResult, error) {
	if e == nil {
		return REDExecutionResult{}, errors.New("executor is nil")
	}
	if err := validateREDInput(input); err != nil {
		return REDExecutionResult{}, err
	}

	timeout := input.Timeout
	if timeout <= 0 {
		timeout = defaultREDTimeout
	}

	dispatchReq := REDDispatchRequest{
		MissionID:     input.MissionID,
		MissionSpec:   input.MissionSpec,
		ACID:          input.ACID,
		ACDescription: input.ACDescription,
		WorktreePath:  input.WorktreePath,
		Instruction:   REDInstruction,
	}
	if err := e.dispatcher.DispatchRED(ctx, dispatchReq); err != nil {
		return REDExecutionResult{}, fmt.Errorf("dispatch RED phase: %w", err)
	}

	waitCtx, cancel := context.WithTimeout(ctx, timeout)
	claim, err := e.waiter.WaitREDComplete(waitCtx, input.MissionID, input.ACID)
	cancel()
	if err != nil {
		if errors.Is(waitCtx.Err(), context.DeadlineExceeded) || errors.Is(err, context.DeadlineExceeded) {
			escalation := REDStuckEscalation{
				MissionID: input.MissionID,
				ACID:      input.ACID,
				Attempt:   input.Attempt,
				Timeout:   timeout,
				Timestamp: e.now().UTC(),
			}
			if escalateErr := e.escalations.PublishREDStuck(ctx, escalation); escalateErr != nil {
				return REDExecutionResult{}, fmt.Errorf("publish RED stuck escalation: %w", escalateErr)
			}
			return REDExecutionResult{
				NextPhase: ACPhaseRed,
				Attempt:   input.Attempt,
			}, fmt.Errorf("%w for mission %s AC %s", ErrREDClaimTimeout, input.MissionID, input.ACID)
		}
		return REDExecutionResult{}, fmt.Errorf("wait for RED_COMPLETE claim: %w", err)
	}
	if claim.EventType != "" && claim.EventType != REDCompleteEventType {
		return REDExecutionResult{}, fmt.Errorf("unexpected RED claim type %q", claim.EventType)
	}

	gateResult, err := e.verifier.VerifyRED(ctx, VerifyREDRequest{
		MissionID:    input.MissionID,
		ACID:         input.ACID,
		WorktreePath: input.WorktreePath,
	})
	if err != nil {
		return REDExecutionResult{}, fmt.Errorf("run VERIFY_RED: %w", err)
	}

	switch gateResult.Classification {
	case GateClassificationAccept:
		if err := e.transitions.TransitionToGreen(ctx, input.MissionID, input.ACID, "VERIFY_RED accepted"); err != nil {
			return REDExecutionResult{}, fmt.Errorf("transition AC to green: %w", err)
		}
		return REDExecutionResult{
			NextPhase: ACPhaseGreen,
			Attempt:   input.Attempt,
			Gate:      gateResult,
		}, nil
	case GateClassificationReject:
		nextAttempt := input.Attempt + 1
		failure := REDFailure{
			MissionID: input.MissionID,
			ACID:      input.ACID,
			Attempt:   nextAttempt,
			Reason:    "VERIFY_RED rejected",
			Output:    strings.TrimSpace(gateResult.Output),
			Timestamp: e.now().UTC(),
		}
		if err := e.failures.RecordREDFailure(ctx, failure); err != nil {
			return REDExecutionResult{}, fmt.Errorf("record RED failure: %w", err)
		}

		feedbackMessage := fmt.Sprintf("VERIFY_RED rejected attempt %d: %s", nextAttempt, strings.TrimSpace(gateResult.Output))
		if err := e.feedback.SendREDFeedback(ctx, REDFeedback{
			MissionID: input.MissionID,
			ACID:      input.ACID,
			Attempt:   nextAttempt,
			Message:   feedbackMessage,
		}); err != nil {
			return REDExecutionResult{}, fmt.Errorf("send RED feedback: %w", err)
		}

		return REDExecutionResult{
			NextPhase: ACPhaseRed,
			Attempt:   nextAttempt,
			Feedback:  feedbackMessage,
			Gate:      gateResult,
		}, nil
	default:
		return REDExecutionResult{}, fmt.Errorf("unsupported VERIFY_RED classification %q", gateResult.Classification)
	}
}

func validateREDInput(input REDExecutionInput) error {
	if strings.TrimSpace(input.MissionID) == "" {
		return errors.New("mission id must not be empty")
	}
	if strings.TrimSpace(input.MissionSpec) == "" {
		return errors.New("mission spec must not be empty")
	}
	if strings.TrimSpace(input.ACID) == "" {
		return errors.New("AC id must not be empty")
	}
	if strings.TrimSpace(input.ACDescription) == "" {
		return errors.New("AC description must not be empty")
	}
	if strings.TrimSpace(input.WorktreePath) == "" {
		return errors.New("worktree path must not be empty")
	}
	if input.Attempt < 0 {
		return errors.New("attempt must be non-negative")
	}
	return nil
}
