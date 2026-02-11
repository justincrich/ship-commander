package commander

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/ship-commander/sc3/internal/admiral"
	"github.com/ship-commander/sc3/internal/protocol"
	"github.com/ship-commander/sc3/internal/telemetry"
	"github.com/ship-commander/sc3/internal/telemetry/invariants"
)

const (
	// EventMissionCompleted is emitted when a mission passes verification.
	EventMissionCompleted = "MISSION_COMPLETED"
	// EventMissionHalted is emitted when a mission fails dispatch or verification.
	EventMissionHalted = "MISSION_HALTED"
	// EventWaveFeedbackRecorded is emitted when Admiral feedback is captured at a wave checkpoint.
	EventWaveFeedbackRecorded = "WAVE_FEEDBACK_RECORDED"
	// EventCommissionHalted is emitted when Admiral halts execution during wave review.
	EventCommissionHalted = "COMMISSION_HALTED"
	// MissionClassificationStandardOps routes mission execution through the standard implementation fast path.
	MissionClassificationStandardOps = "STANDARD_OPS"
	// DefaultMaxRevisions is the deterministic default revision ceiling before halting.
	DefaultMaxRevisions = 3
	// defaultReviewPollInterval determines how often commander polls protocol store for review verdicts.
	defaultReviewPollInterval = 200 * time.Millisecond
	// defaultReviewTimeout bounds reviewer verdict waiting for deterministic mission completion.
	defaultReviewTimeout = 5 * time.Minute
)

var (
	// ErrApprovalFeedback indicates execution was paused because Admiral requested planning feedback.
	ErrApprovalFeedback = errors.New("admiral requested planning feedback")
	// ErrApprovalShelved indicates execution was paused because Admiral shelved the manifest.
	ErrApprovalShelved = errors.New("admiral shelved mission manifest")
)

// HaltReason is a deterministic reason enum for mission halts.
type HaltReason string

const (
	// HaltReasonMaxRevisionsExceeded indicates revision count reached the max revision limit.
	HaltReasonMaxRevisionsExceeded HaltReason = "MaxRevisionsExceeded"
	// HaltReasonDemoTokenInvalid indicates demo token validation failed for reasons other than missing token.
	HaltReasonDemoTokenInvalid HaltReason = "DemoTokenInvalid"
	// HaltReasonDemoTokenMissing indicates the demo token artifact was not found.
	HaltReasonDemoTokenMissing HaltReason = "DemoTokenMissing"
	// HaltReasonACExhausted indicates all AC attempts were exhausted before success.
	HaltReasonACExhausted HaltReason = "ACExhausted"
	// HaltReasonManualHalt indicates an operator-initiated or explicit manual halt.
	HaltReasonManualHalt HaltReason = "ManualHalt"
)

// Mission is an executable mission in an approved manifest.
type Mission struct {
	ID                         string
	Title                      string
	Harness                    string
	Model                      string
	Classification             string
	ClassificationRationale    string
	ClassificationCriteria     []string
	ClassificationConfidence   string
	ClassificationNeedsReview  bool
	ClassificationReviewSource string
	DependsOn                  []string
	UseCaseIDs                 []string
	SurfaceArea                []string
	WaveFeedback               string
	ReviewFeedback             string
	RevisionCount              int
	MaxRevisions               int
	// ACAttemptsExhausted indicates all AC attempts failed and mission must halt deterministically.
	ACAttemptsExhausted bool
	// ManualHalt requests deterministic dispatch stop before running mission work.
	ManualHalt bool
	// AcceptanceCriteria are forwarded to reviewer context for independent validation.
	AcceptanceCriteria []string
}

// Slug returns a URL-safe slug for branch naming.
func (m Mission) Slug() string {
	source := strings.TrimSpace(m.Title)
	if source == "" {
		source = strings.TrimSpace(m.ID)
	}
	return slugify(source)
}

// Event is a protocol event emitted by the commander.
type Event struct {
	Type      string
	MissionID string
	WaveIndex int
	Timestamp time.Time
	Message   string
	Reason    HaltReason
	NotifyTUI bool
}

// DispatchRequest contains mission dispatch details for harness implementations.
type DispatchRequest struct {
	Mission      Mission
	WorktreePath string
	// WaveFeedback is Admiral feedback from the previous wave checkpoint.
	WaveFeedback string
	// ReviewerFeedback is populated when a prior review returned NEEDS_FIXES.
	ReviewerFeedback string
}

// ReviewerDispatchRequest contains reviewer context payload.
type ReviewerDispatchRequest struct {
	Mission                     Mission
	WorktreePath                string
	CodeDiff                    string
	GateEvidence                []string
	AcceptanceCriteria          []string
	DemoTokenContent            string
	ImplementerSessionID        string
	ReadOnlyWorktree            bool
	IncludeImplementerReasoning bool
}

// DispatchResult captures dispatch metadata from a harness implementation.
type DispatchResult struct {
	SessionID string
}

// ManifestStore reads mission manifests and ready mission IDs from Beads.
type ManifestStore interface {
	ReadApprovedManifest(ctx context.Context, commissionID string) ([]Mission, error)
	ReadyMissionIDs(ctx context.Context, commissionID string) ([]string, error)
}

// WorktreeManager creates mission worktrees.
type WorktreeManager interface {
	Create(ctx context.Context, mission Mission) (string, error)
}

// SurfaceLocker acquires and releases mission surface-area locks.
type SurfaceLocker interface {
	Acquire(ctx context.Context, missionID string, patterns []string) (func() error, error)
}

// Harness dispatches implementer sessions.
type Harness interface {
	DispatchImplementer(ctx context.Context, req DispatchRequest) (DispatchResult, error)
	DispatchReviewer(ctx context.Context, req ReviewerDispatchRequest) (DispatchResult, error)
}

// Verifier verifies mission output independently from the implementer agent.
type Verifier interface {
	Verify(ctx context.Context, mission Mission, worktreePath string) error
	VerifyImplement(ctx context.Context, mission Mission, worktreePath string) error
}

// DemoTokenValidator validates a mission's demo token before completion.
type DemoTokenValidator interface {
	Validate(ctx context.Context, mission Mission, worktreePath string) error
}

// ApprovalGate presents mission manifests to Admiral and blocks until a decision is made.
type ApprovalGate interface {
	AwaitDecision(ctx context.Context, request admiral.ApprovalRequest) (admiral.ApprovalResponse, error)
}

// FeedbackInjector reinjects Admiral feedback into Ready Room planning sessions.
type FeedbackInjector interface {
	InjectPlanningFeedback(ctx context.Context, commissionID, feedbackText string) error
}

// PlanShelver persists a shelved plan for later resume/re-execute flows.
type PlanShelver interface {
	ShelvePlan(ctx context.Context, commissionID, feedbackText string) error
}

// EventPublisher publishes protocol events for mission status changes.
type EventPublisher interface {
	Publish(ctx context.Context, event Event) error
}

// ProtocolEventStore provides mission-scoped protocol history used by reviewer flows.
type ProtocolEventStore interface {
	ListByMission(ctx context.Context, missionID string) ([]protocol.ProtocolEvent, error)
}

// ReviewVerdict captures reviewer decision and feedback.
type ReviewVerdict struct {
	Decision string
	Feedback string
}

// CommanderConfig configures commander runtime behavior.
type CommanderConfig struct {
	WIPLimit           int
	ProtocolEventStore ProtocolEventStore
	ReviewPollInterval time.Duration
	ReviewTimeout      time.Duration
}

// Commander orchestrates mission execution from approved manifest through verification.
type Commander struct {
	manifestStore ManifestStore
	worktrees     WorktreeManager
	locks         SurfaceLocker
	harness       Harness
	verifier      Verifier
	demoTokens    DemoTokenValidator
	approvalGate  ApprovalGate
	feedback      FeedbackInjector
	shelver       PlanShelver
	events        EventPublisher
	protocolStore ProtocolEventStore
	wipLimit      int
	reviewPoll    time.Duration
	reviewTimeout time.Duration
	missionPaths  sync.Map
	now           func() time.Time
}

// New creates a Commander with required dependencies.
func New(
	store ManifestStore,
	worktrees WorktreeManager,
	locks SurfaceLocker,
	harness Harness,
	verifier Verifier,
	demoTokens DemoTokenValidator,
	approvalGate ApprovalGate,
	feedback FeedbackInjector,
	shelver PlanShelver,
	events EventPublisher,
	cfg CommanderConfig,
) (*Commander, error) {
	if store == nil {
		return nil, errors.New("manifest store is required")
	}
	if worktrees == nil {
		return nil, errors.New("worktree manager is required")
	}
	if locks == nil {
		return nil, errors.New("surface locker is required")
	}
	if harness == nil {
		return nil, errors.New("harness is required")
	}
	if verifier == nil {
		return nil, errors.New("verifier is required")
	}
	if demoTokens == nil {
		return nil, errors.New("demo token validator is required")
	}
	if approvalGate == nil {
		return nil, errors.New("approval gate is required")
	}
	if feedback == nil {
		return nil, errors.New("feedback injector is required")
	}
	if shelver == nil {
		return nil, errors.New("plan shelver is required")
	}
	if events == nil {
		return nil, errors.New("event publisher is required")
	}
	if cfg.WIPLimit <= 0 {
		return nil, errors.New("wip limit must be positive")
	}

	return &Commander{
		manifestStore: store,
		worktrees:     worktrees,
		locks:         locks,
		harness:       harness,
		verifier:      verifier,
		demoTokens:    demoTokens,
		approvalGate:  approvalGate,
		feedback:      feedback,
		shelver:       shelver,
		events:        events,
		protocolStore: cfg.ProtocolEventStore,
		wipLimit:      cfg.WIPLimit,
		reviewPoll:    pickDuration(cfg.ReviewPollInterval, defaultReviewPollInterval),
		reviewTimeout: pickDuration(cfg.ReviewTimeout, defaultReviewTimeout),
		now:           time.Now,
	}, nil
}

// Execute runs the propulsion loop for an approved commission manifest.
func (c *Commander) Execute(ctx context.Context, commissionID string) error {
	if strings.TrimSpace(commissionID) == "" {
		return errors.New("commission id must not be empty")
	}

	manifest, err := c.manifestStore.ReadApprovedManifest(ctx, commissionID)
	if err != nil {
		return fmt.Errorf("read approved manifest: %w", err)
	}
	waves, err := ComputeWaves(manifest)
	if err != nil {
		return fmt.Errorf("compute waves: %w", err)
	}
	if err := c.resolveAdmiralDecision(ctx, commissionID, manifest, waves); err != nil {
		return err
	}

	waveFeedback := ""
	for i, wave := range waves {
		waveIndex := i + 1
		if err := c.executeWave(ctx, commissionID, waveIndex, wave, waveFeedback); err != nil {
			return fmt.Errorf("execute wave %d: %w", i+1, err)
		}
		waveFeedback = ""
		if i == len(waves)-1 {
			continue
		}
		nextWaveFeedback, err := c.runWaveReview(ctx, commissionID, waveIndex, wave)
		if err != nil {
			return err
		}
		waveFeedback = nextWaveFeedback
	}

	return nil
}

func (c *Commander) executeWave(
	ctx context.Context,
	commissionID string,
	waveIndex int,
	missions []Mission,
	waveFeedback string,
) error {
	if len(missions) == 0 {
		return nil
	}

	pending := make(map[string]Mission, len(missions))
	order := make([]string, 0, len(missions))
	for _, mission := range missions {
		mission.WaveFeedback = strings.TrimSpace(waveFeedback)
		pending[mission.ID] = mission
		order = append(order, mission.ID)
	}

	for len(pending) > 0 {
		readyIDs, err := c.manifestStore.ReadyMissionIDs(ctx, commissionID)
		if err != nil {
			return fmt.Errorf("query ready missions: %w", err)
		}

		readySet := make(map[string]struct{}, len(readyIDs))
		for _, id := range readyIDs {
			readySet[id] = struct{}{}
		}

		batch := make([]Mission, 0, c.wipLimit)
		for _, id := range order {
			mission, ok := pending[id]
			if !ok {
				continue
			}
			if _, ok := readySet[id]; !ok {
				continue
			}
			batch = append(batch, mission)
			if len(batch) == c.wipLimit {
				break
			}
		}

		if len(batch) == 0 {
			return fmt.Errorf("no unblocked missions available while %d missions remain in wave", len(pending))
		}

		if err := c.runBatch(ctx, waveIndex, batch); err != nil {
			return err
		}
		for _, mission := range batch {
			delete(pending, mission.ID)
		}
	}

	return nil
}

func (c *Commander) runBatch(ctx context.Context, waveIndex int, batch []Mission) error {
	var wg sync.WaitGroup
	errCh := make(chan error, len(batch))

	for _, mission := range batch {
		mission := mission
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := c.runMission(ctx, waveIndex, mission); err != nil {
				errCh <- err
			}
		}()
	}

	wg.Wait()
	close(errCh)

	var errs []error
	for err := range errCh {
		errs = append(errs, err)
	}
	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}

func (c *Commander) runMission(ctx context.Context, waveIndex int, mission Mission) error {
	if reason, message, shouldHalt := haltBeforeDispatch(mission); shouldHalt {
		if reason == HaltReasonMaxRevisionsExceeded {
			maxRevisions := mission.MaxRevisions
			if maxRevisions <= 0 {
				maxRevisions = DefaultMaxRevisions
			}
			invariants.CheckMaxRetriesNotExceeded(
				ctx,
				"commander.runMission",
				mission.RevisionCount,
				maxRevisions,
			)
		}
		_ = c.publishHalt(ctx, waveIndex, mission.ID, reason, message)
		return fmt.Errorf("mission %s halted before dispatch: %s", mission.ID, message)
	}

	worktreePath, err := c.worktrees.Create(ctx, mission)
	if err != nil {
		_ = c.publishHalt(ctx, waveIndex, mission.ID, HaltReasonManualHalt, fmt.Sprintf("worktree creation failed: %v", err))
		return fmt.Errorf("create worktree for %s: %w", mission.ID, err)
	}
	c.missionPaths.Store(mission.ID, worktreePath)
	cleanRepo, repoStatus := isGitWorktreeClean(ctx, worktreePath)
	invariants.CheckRepoCleanBeforeMerge(
		ctx,
		"commander.runMission",
		cleanRepo,
		repoStatus,
	)
	invariants.CheckEditsWithinAllowedPaths(
		ctx,
		"commander.runMission",
		mission.SurfaceArea,
		nil,
	)

	release, err := c.locks.Acquire(ctx, mission.ID, mission.SurfaceArea)
	if err != nil {
		_ = c.publishHalt(ctx, waveIndex, mission.ID, HaltReasonManualHalt, fmt.Sprintf("surface-area lock failed: %v", err))
		return fmt.Errorf("acquire lock for %s: %w", mission.ID, err)
	}
	defer func() {
		_ = release()
	}()

	maxRevisions := mission.MaxRevisions
	if maxRevisions <= 0 {
		maxRevisions = DefaultMaxRevisions
	}
	currentMission := mission

	for {
		implementerResult, err := c.dispatchImplementer(ctx, currentMission, worktreePath, waveIndex)
		if err != nil {
			return err
		}

		if err := c.verifyMissionOutput(ctx, currentMission, worktreePath, waveIndex); err != nil {
			return err
		}

		verdict, err := c.dispatchReviewerAndAwaitVerdict(
			ctx,
			currentMission,
			worktreePath,
			waveIndex,
			implementerResult.SessionID,
		)
		if err != nil {
			return err
		}

		done, err := c.handleReviewVerdict(ctx, mission.ID, waveIndex, &currentMission, maxRevisions, verdict)
		if err != nil {
			return err
		}
		if done {
			return nil
		}
	}
}

func (c *Commander) dispatchImplementer(
	ctx context.Context,
	mission Mission,
	worktreePath string,
	waveIndex int,
) (DispatchResult, error) {
	dispatchCtx, llmCall := telemetry.StartLLMCall(ctx, telemetry.LLMCallRequest{
		Operation: "dispatch_implementer",
		ModelName: mission.Model,
		Harness:   mission.Harness,
		Prompt:    buildDispatchTelemetryPrompt(mission, waveIndex),
	})

	result, err := c.harness.DispatchImplementer(dispatchCtx, DispatchRequest{
		Mission:          mission,
		WorktreePath:     worktreePath,
		WaveFeedback:     mission.WaveFeedback,
		ReviewerFeedback: mission.ReviewFeedback,
	})
	if err != nil {
		llmCall.RecordError("implementer_dispatch_error", err.Error(), mission.RevisionCount)
		llmCall.End("", nil, err)
		_ = c.publishHalt(ctx, waveIndex, mission.ID, HaltReasonManualHalt, fmt.Sprintf("dispatch failed: %v", err))
		return DispatchResult{}, fmt.Errorf("dispatch implementer for %s: %w", mission.ID, err)
	}
	llmCall.End(result.SessionID, nil, nil)
	return result, nil
}

func (c *Commander) verifyMissionOutput(
	ctx context.Context,
	mission Mission,
	worktreePath string,
	waveIndex int,
) error {
	if isStandardOpsMission(mission) {
		if err := c.verifier.VerifyImplement(ctx, mission, worktreePath); err != nil {
			invariants.CheckPatchApplyClean(
				ctx,
				"commander.verifyMissionOutput",
				!looksLikePatchFailure(err),
				err.Error(),
			)
			_ = c.publishHalt(ctx, waveIndex, mission.ID, HaltReasonManualHalt, fmt.Sprintf("verification failed: %v", err))
			return fmt.Errorf("verify implement mission %s: %w", mission.ID, err)
		}
		if err := c.demoTokens.Validate(ctx, mission, worktreePath); err != nil {
			_ = c.publishHalt(
				ctx,
				waveIndex,
				mission.ID,
				classifyDemoTokenHaltReason(err),
				fmt.Sprintf("demo token validation failed: %v", err),
			)
			return fmt.Errorf("validate demo token for %s: %w", mission.ID, err)
		}
		return nil
	}

	if err := c.verifier.Verify(ctx, mission, worktreePath); err != nil {
		invariants.CheckPatchApplyClean(
			ctx,
			"commander.verifyMissionOutput",
			!looksLikePatchFailure(err),
			err.Error(),
		)
		_ = c.publishHalt(ctx, waveIndex, mission.ID, HaltReasonManualHalt, fmt.Sprintf("verification failed: %v", err))
		return fmt.Errorf("verify mission %s: %w", mission.ID, err)
	}
	return nil
}

func (c *Commander) dispatchReviewerAndAwaitVerdict(
	ctx context.Context,
	mission Mission,
	worktreePath string,
	waveIndex int,
	implementerSessionID string,
) (ReviewVerdict, error) {
	reviewerReq, err := c.buildReviewerDispatchRequest(ctx, mission, worktreePath, implementerSessionID)
	if err != nil {
		_ = c.publishHalt(ctx, waveIndex, mission.ID, HaltReasonManualHalt, fmt.Sprintf("build reviewer context failed: %v", err))
		return ReviewVerdict{}, fmt.Errorf("build reviewer context for %s: %w", mission.ID, err)
	}

	reviewCtx, llmCall := telemetry.StartLLMCall(ctx, telemetry.LLMCallRequest{
		Operation: "dispatch_reviewer",
		ModelName: mission.Model,
		Harness:   mission.Harness,
		Prompt:    buildReviewerTelemetryPrompt(mission, reviewerReq, waveIndex),
	})

	reviewerResult, err := c.harness.DispatchReviewer(reviewCtx, reviewerReq)
	if err != nil {
		llmCall.RecordError("reviewer_dispatch_error", err.Error(), mission.RevisionCount)
		llmCall.End("", nil, err)
		_ = c.publishHalt(ctx, waveIndex, mission.ID, HaltReasonManualHalt, fmt.Sprintf("reviewer dispatch failed: %v", err))
		return ReviewVerdict{}, fmt.Errorf("dispatch reviewer for %s: %w", mission.ID, err)
	}

	reviewerSession := strings.TrimSpace(reviewerResult.SessionID)
	implementerSession := strings.TrimSpace(implementerSessionID)
	if reviewerSession == "" {
		llmCall.RecordError("reviewer_session_invalid", "reviewer dispatch returned empty session id", mission.RevisionCount)
		llmCall.End("", nil, errors.New("reviewer dispatch returned empty session id"))
		_ = c.publishHalt(ctx, waveIndex, mission.ID, HaltReasonManualHalt, "reviewer dispatch returned empty session id")
		return ReviewVerdict{}, fmt.Errorf("dispatch reviewer for %s: empty reviewer session id", mission.ID)
	}
	if implementerSession != "" && reviewerSession == implementerSession {
		llmCall.RecordError(
			"reviewer_session_invalid",
			"reviewer must be a different ensign session than implementer",
			mission.RevisionCount,
		)
		llmCall.End("", nil, errors.New("reviewer and implementer session ids must differ"))
		_ = c.publishHalt(ctx, waveIndex, mission.ID, HaltReasonManualHalt, "reviewer must be a different ensign session than implementer")
		return ReviewVerdict{}, fmt.Errorf("dispatch reviewer for %s: reviewer and implementer session ids must differ", mission.ID)
	}

	verdict, err := c.awaitReviewVerdict(reviewCtx, mission.ID, implementerSession, reviewerSession)
	if err != nil {
		llmCall.RecordError("review_verdict_wait_error", err.Error(), mission.RevisionCount)
		llmCall.End(reviewerSession, nil, err)
		_ = c.publishHalt(ctx, waveIndex, mission.ID, HaltReasonManualHalt, fmt.Sprintf("review verdict wait failed: %v", err))
		return ReviewVerdict{}, fmt.Errorf("await review verdict for %s: %w", mission.ID, err)
	}
	llmCall.End(fmt.Sprintf("%s:%s", reviewerSession, verdict.Decision), nil, nil)
	return verdict, nil
}

func (c *Commander) handleReviewVerdict(
	ctx context.Context,
	missionID string,
	waveIndex int,
	mission *Mission,
	maxRevisions int,
	verdict ReviewVerdict,
) (bool, error) {
	switch verdict.Decision {
	case protocol.ReviewVerdictApproved:
		if err := c.publish(ctx, Event{
			Type:      EventMissionCompleted,
			MissionID: missionID,
			WaveIndex: waveIndex,
			Timestamp: c.now().UTC(),
			Message:   "mission verified and reviewer approved",
		}); err != nil {
			return false, fmt.Errorf("publish completion event for %s: %w", missionID, err)
		}
		return true, nil
	case protocol.ReviewVerdictNeedsFixes:
		mission.RevisionCount++
		mission.ReviewFeedback = strings.TrimSpace(verdict.Feedback)
		if mission.RevisionCount >= maxRevisions {
			invariants.CheckMaxRetriesNotExceeded(
				ctx,
				"commander.handleReviewVerdict",
				mission.RevisionCount,
				maxRevisions,
			)
			message := fmt.Sprintf(
				"review requested fixes and revision count %d reached max revisions %d",
				mission.RevisionCount,
				maxRevisions,
			)
			_ = c.publishHalt(ctx, waveIndex, missionID, HaltReasonMaxRevisionsExceeded, message)
			return false, fmt.Errorf("mission %s halted after review: %s", missionID, message)
		}
		return false, nil
	default:
		_ = c.publishHalt(
			ctx,
			waveIndex,
			missionID,
			HaltReasonManualHalt,
			fmt.Sprintf("unsupported reviewer verdict %q", verdict.Decision),
		)
		return false, fmt.Errorf("unsupported reviewer verdict %q for mission %s", verdict.Decision, missionID)
	}
}

func (c *Commander) runWaveReview(
	ctx context.Context,
	commissionID string,
	waveIndex int,
	missions []Mission,
) (string, error) {
	demoTokens, err := c.collectWaveDemoTokens(missions)
	if err != nil {
		return "", fmt.Errorf("collect wave %d demo tokens: %w", waveIndex, err)
	}

	response, err := c.approvalGate.AwaitDecision(ctx, buildWaveReviewRequest(commissionID, waveIndex, missions, demoTokens))
	if err != nil {
		return "", fmt.Errorf("await wave %d review decision: %w", waveIndex, err)
	}

	switch response.Decision {
	case admiral.ApprovalDecisionApproved:
		return "", nil
	case admiral.ApprovalDecisionFeedback:
		feedbackText := strings.TrimSpace(response.FeedbackText)
		if err := c.publish(ctx, Event{
			Type:      EventWaveFeedbackRecorded,
			WaveIndex: waveIndex,
			Timestamp: c.now().UTC(),
			Message:   feedbackText,
			NotifyTUI: true,
		}); err != nil {
			return "", fmt.Errorf("publish wave %d feedback: %w", waveIndex, err)
		}
		return feedbackText, nil
	case admiral.ApprovalDecisionHalted, admiral.ApprovalDecisionShelved:
		message := strings.TrimSpace(response.FeedbackText)
		if message == "" {
			message = fmt.Sprintf("admiral halted execution after wave %d review", waveIndex)
		}
		if err := c.publish(ctx, Event{
			Type:      EventCommissionHalted,
			WaveIndex: waveIndex,
			Timestamp: c.now().UTC(),
			Message:   message,
			NotifyTUI: true,
		}); err != nil {
			return "", fmt.Errorf("publish commission halt after wave %d review: %w", waveIndex, err)
		}
		return "", fmt.Errorf("execution halted after wave %d review: %s", waveIndex, message)
	default:
		return "", fmt.Errorf("unsupported wave review decision %q", response.Decision)
	}
}

func (c *Commander) collectWaveDemoTokens(missions []Mission) (map[string]string, error) {
	demoTokens := make(map[string]string, len(missions))
	for _, mission := range missions {
		worktreePathRaw, ok := c.missionPaths.Load(mission.ID)
		if !ok {
			return nil, fmt.Errorf("worktree path missing for mission %s", mission.ID)
		}
		worktreePath, ok := worktreePathRaw.(string)
		if !ok || strings.TrimSpace(worktreePath) == "" {
			return nil, fmt.Errorf("worktree path invalid for mission %s", mission.ID)
		}
		token, err := readDemoToken(worktreePath, mission.ID)
		if err != nil {
			return nil, fmt.Errorf("read demo token for mission %s: %w", mission.ID, err)
		}
		demoTokens[mission.ID] = token
	}
	return demoTokens, nil
}

func (c *Commander) buildReviewerDispatchRequest(
	ctx context.Context,
	mission Mission,
	worktreePath string,
	implementerSessionID string,
) (ReviewerDispatchRequest, error) {
	diff, err := gitDiff(ctx, worktreePath)
	if err != nil {
		diff = fmt.Sprintf("diff unavailable: %v", err)
	}

	gateEvidence, err := c.collectGateEvidence(ctx, mission.ID)
	if err != nil {
		return ReviewerDispatchRequest{}, fmt.Errorf("collect gate evidence: %w", err)
	}

	demoToken, err := readDemoToken(worktreePath, mission.ID)
	if err != nil {
		demoToken = fmt.Sprintf("demo token unavailable: %v", err)
	}

	return ReviewerDispatchRequest{
		Mission:                     mission,
		WorktreePath:                worktreePath,
		CodeDiff:                    diff,
		GateEvidence:                gateEvidence,
		AcceptanceCriteria:          append([]string(nil), mission.AcceptanceCriteria...),
		DemoTokenContent:            demoToken,
		ImplementerSessionID:        strings.TrimSpace(implementerSessionID),
		ReadOnlyWorktree:            true,
		IncludeImplementerReasoning: false,
	}, nil
}

func (c *Commander) collectGateEvidence(ctx context.Context, missionID string) ([]string, error) {
	if c.protocolStore == nil {
		return []string{"gate evidence unavailable: protocol store not configured"}, nil
	}

	events, err := c.protocolStore.ListByMission(ctx, missionID)
	if err != nil {
		return nil, fmt.Errorf("list protocol events for mission %s: %w", missionID, err)
	}

	gateEvidence := make([]string, 0, len(events))
	for _, event := range events {
		if event.Type != protocol.EventTypeGateResult {
			continue
		}
		payload := strings.TrimSpace(string(event.Payload))
		if payload == "" {
			payload = "{}"
		}
		gateEvidence = append(gateEvidence, fmt.Sprintf("%s %s", event.Timestamp.UTC().Format(time.RFC3339), payload))
	}
	if len(gateEvidence) == 0 {
		return []string{"no gate evidence events recorded for mission"}, nil
	}

	return gateEvidence, nil
}

func (c *Commander) awaitReviewVerdict(
	ctx context.Context,
	missionID string,
	implementerSessionID string,
	reviewerSessionID string,
) (ReviewVerdict, error) {
	if c.protocolStore == nil {
		return ReviewVerdict{Decision: protocol.ReviewVerdictApproved}, nil
	}

	waitCtx, cancel := context.WithTimeout(ctx, c.reviewTimeout)
	defer cancel()

	for {
		verdict, found, err := c.findReviewVerdict(waitCtx, missionID, implementerSessionID, reviewerSessionID)
		if err != nil {
			return ReviewVerdict{}, err
		}
		if found {
			return verdict, nil
		}

		select {
		case <-waitCtx.Done():
			return ReviewVerdict{}, fmt.Errorf(
				"timed out waiting for review verdict event %q for mission %s",
				protocol.EventTypeReviewComplete,
				missionID,
			)
		case <-time.After(c.reviewPoll):
		}
	}
}

func (c *Commander) findReviewVerdict(
	ctx context.Context,
	missionID string,
	implementerSessionID string,
	reviewerSessionID string,
) (ReviewVerdict, bool, error) {
	events, err := c.protocolStore.ListByMission(ctx, missionID)
	if err != nil {
		return ReviewVerdict{}, false, fmt.Errorf("list protocol events for mission %s: %w", missionID, err)
	}

	for i := len(events) - 1; i >= 0; i-- {
		verdict, verdictImplementerSessionID, verdictReviewerSessionID, ok := parseReviewVerdict(events[i])
		if !ok {
			continue
		}
		if implementerSessionID != "" && verdictImplementerSessionID != "" && verdictImplementerSessionID != implementerSessionID {
			continue
		}
		if reviewerSessionID != "" && verdictReviewerSessionID != "" && verdictReviewerSessionID != reviewerSessionID {
			continue
		}
		return ReviewVerdict{
			Decision: verdict,
			Feedback: firstNonEmptyString(
				extractJSONString(events[i].Payload, "feedback"),
				extractJSONString(events[i].Payload, "feedback_text"),
				extractJSONString(events[i].Payload, "feedbackText"),
			),
		}, true, nil
	}
	return ReviewVerdict{}, false, nil
}

func parseReviewVerdict(event protocol.ProtocolEvent) (string, string, string, bool) {
	if event.Type != protocol.EventTypeReviewComplete {
		return "", "", "", false
	}

	var payload map[string]any
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return "", "", "", false
	}

	verdict := strings.ToUpper(strings.TrimSpace(firstNonEmptyMap(payload, "verdict", "decision")))
	if verdict != protocol.ReviewVerdictApproved && verdict != protocol.ReviewVerdictNeedsFixes {
		return "", "", "", false
	}

	return verdict,
		strings.TrimSpace(firstNonEmptyMap(payload, "implementer_session_id", "implementerSessionID", "implementer_session")),
		strings.TrimSpace(firstNonEmptyMap(payload, "reviewer_session_id", "reviewerSessionID", "reviewer_session")),
		true
}

func firstNonEmptyMap(values map[string]any, keys ...string) string {
	for _, key := range keys {
		raw, ok := values[key]
		if !ok {
			continue
		}
		value, ok := raw.(string)
		if !ok {
			continue
		}
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

func extractJSONString(raw json.RawMessage, keys ...string) string {
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return ""
	}
	return firstNonEmptyMap(payload, keys...)
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

func gitDiff(ctx context.Context, worktreePath string) (string, error) {
	out, err := exec.CommandContext(ctx, "git", "-C", worktreePath, "diff", "--").CombinedOutput()
	if err != nil {
		trimmed := strings.TrimSpace(string(out))
		if trimmed == "" {
			return "", fmt.Errorf("git diff: %w", err)
		}
		return "", fmt.Errorf("git diff: %w (%s)", err, trimmed)
	}
	return string(out), nil
}

func isGitWorktreeClean(ctx context.Context, worktreePath string) (bool, string) {
	out, err := exec.CommandContext(ctx, "git", "-C", worktreePath, "status", "--porcelain").CombinedOutput()
	if err != nil {
		trimmed := strings.TrimSpace(string(out))
		if trimmed == "" {
			trimmed = err.Error()
		}
		return false, trimmed
	}
	trimmed := strings.TrimSpace(string(out))
	return trimmed == "", trimmed
}

func looksLikePatchFailure(err error) bool {
	if err == nil {
		return false
	}
	text := strings.ToLower(err.Error())
	return strings.Contains(text, "patch") ||
		strings.Contains(text, "hunk") ||
		strings.Contains(text, "fuzz") ||
		strings.Contains(text, "reject")
}

func readDemoToken(worktreePath string, missionID string) (string, error) {
	root := filepath.Clean(worktreePath)
	if root == "." || root == "" {
		return "", errors.New("worktree path must not be empty")
	}
	tokenPath := filepath.Clean(filepath.Join(root, "demo", fmt.Sprintf("MISSION-%s.md", missionID)))
	rootWithSep := root + string(os.PathSeparator)
	if tokenPath != root && !strings.HasPrefix(tokenPath, rootWithSep) {
		return "", fmt.Errorf("demo token path escapes worktree root: %s", tokenPath)
	}
	// #nosec G304 -- tokenPath is constrained to worktree root and deterministic mission filename.
	content, err := os.ReadFile(tokenPath)
	if err != nil {
		return "", fmt.Errorf("read demo token %s: %w", tokenPath, err)
	}
	return string(content), nil
}

func pickDuration(value time.Duration, fallback time.Duration) time.Duration {
	if value > 0 {
		return value
	}
	return fallback
}

func (c *Commander) publishHalt(
	ctx context.Context,
	waveIndex int,
	missionID string,
	reason HaltReason,
	message string,
) error {
	return c.publish(ctx, Event{
		Type:      EventMissionHalted,
		MissionID: missionID,
		WaveIndex: waveIndex,
		Timestamp: c.now().UTC(),
		Message:   message,
		Reason:    reason,
		NotifyTUI: true,
	})
}

func (c *Commander) publish(ctx context.Context, event Event) error {
	return c.events.Publish(ctx, event)
}

func haltBeforeDispatch(mission Mission) (HaltReason, string, bool) {
	if mission.ManualHalt {
		return HaltReasonManualHalt, "mission manually halted before dispatch", true
	}
	if mission.ACAttemptsExhausted {
		return HaltReasonACExhausted, "all acceptance criteria attempts exhausted", true
	}

	maxRevisions := mission.MaxRevisions
	if maxRevisions <= 0 {
		maxRevisions = DefaultMaxRevisions
	}
	if mission.RevisionCount >= maxRevisions {
		return HaltReasonMaxRevisionsExceeded,
			fmt.Sprintf("revision count %d reached max revisions %d", mission.RevisionCount, maxRevisions),
			true
	}

	return "", "", false
}

func classifyDemoTokenHaltReason(err error) HaltReason {
	if err == nil {
		return HaltReasonDemoTokenInvalid
	}
	if errors.Is(err, os.ErrNotExist) {
		return HaltReasonDemoTokenMissing
	}

	lowerMessage := strings.ToLower(err.Error())
	if strings.Contains(lowerMessage, "does not exist") ||
		strings.Contains(lowerMessage, "not found") ||
		strings.Contains(lowerMessage, "missing") {
		return HaltReasonDemoTokenMissing
	}

	return HaltReasonDemoTokenInvalid
}

func isStandardOpsMission(mission Mission) bool {
	return strings.EqualFold(strings.TrimSpace(mission.Classification), MissionClassificationStandardOps)
}

func buildDispatchTelemetryPrompt(mission Mission, waveIndex int) string {
	return fmt.Sprintf(
		"mission_id=%s title=%s wave=%d wave_feedback=%s reviewer_feedback=%s",
		strings.TrimSpace(mission.ID),
		strings.TrimSpace(mission.Title),
		waveIndex,
		strings.TrimSpace(mission.WaveFeedback),
		strings.TrimSpace(mission.ReviewFeedback),
	)
}

func buildReviewerTelemetryPrompt(mission Mission, req ReviewerDispatchRequest, waveIndex int) string {
	return fmt.Sprintf(
		"mission_id=%s title=%s wave=%d gate_evidence=%d acceptance_criteria=%d",
		strings.TrimSpace(mission.ID),
		strings.TrimSpace(mission.Title),
		waveIndex,
		len(req.GateEvidence),
		len(req.AcceptanceCriteria),
	)
}

func (c *Commander) resolveAdmiralDecision(
	ctx context.Context,
	commissionID string,
	manifest []Mission,
	waves [][]Mission,
) error {
	response, err := c.approvalGate.AwaitDecision(ctx, buildApprovalRequest(commissionID, manifest, waves))
	if err != nil {
		return fmt.Errorf("await admiral approval: %w", err)
	}

	switch response.Decision {
	case admiral.ApprovalDecisionApproved:
		return nil
	case admiral.ApprovalDecisionFeedback:
		feedbackText := strings.TrimSpace(response.FeedbackText)
		if err := c.feedback.InjectPlanningFeedback(ctx, commissionID, feedbackText); err != nil {
			return fmt.Errorf("inject planning feedback: %w", err)
		}
		return fmt.Errorf("%w: %s", ErrApprovalFeedback, feedbackText)
	case admiral.ApprovalDecisionShelved:
		if err := c.shelver.ShelvePlan(ctx, commissionID, strings.TrimSpace(response.FeedbackText)); err != nil {
			return fmt.Errorf("shelve plan: %w", err)
		}
		return ErrApprovalShelved
	default:
		return fmt.Errorf("unsupported approval decision %q", response.Decision)
	}
}

func buildApprovalRequest(
	commissionID string,
	manifest []Mission,
	waves [][]Mission,
) admiral.ApprovalRequest {
	requestMissions := make([]admiral.Mission, 0, len(manifest))
	coverage := make(map[string]admiral.CoverageStatus)
	for _, mission := range manifest {
		requestMissions = append(requestMissions, admiral.Mission{
			ID:                        mission.ID,
			Title:                     mission.Title,
			DependsOn:                 append([]string(nil), mission.DependsOn...),
			UseCaseIDs:                append([]string(nil), mission.UseCaseIDs...),
			Classification:            mission.Classification,
			ClassificationRationale:   mission.ClassificationRationale,
			ClassificationCriteria:    append([]string(nil), mission.ClassificationCriteria...),
			ClassificationConfidence:  mission.ClassificationConfidence,
			ClassificationNeedsReview: mission.ClassificationNeedsReview,
		})
		for _, useCaseID := range mission.UseCaseIDs {
			useCaseID = strings.TrimSpace(useCaseID)
			if useCaseID == "" {
				continue
			}
			coverage[useCaseID] = admiral.CoverageStatusCovered
		}
	}

	assignments := make([]admiral.Wave, 0, len(waves))
	for i, wave := range waves {
		missionIDs := make([]string, 0, len(wave))
		for _, mission := range wave {
			missionIDs = append(missionIDs, mission.ID)
		}
		assignments = append(assignments, admiral.Wave{
			Index:      i + 1,
			MissionIDs: missionIDs,
		})
	}

	return admiral.ApprovalRequest{
		CommissionID:    commissionID,
		MissionManifest: requestMissions,
		WaveAssignments: assignments,
		CoverageMap:     coverage,
		Iteration:       1,
		MaxIterations:   1,
	}
}

func buildWaveReviewRequest(
	commissionID string,
	waveIndex int,
	missions []Mission,
	demoTokens map[string]string,
) admiral.ApprovalRequest {
	requestMissions := make([]admiral.Mission, 0, len(missions))
	missionIDs := make([]string, 0, len(missions))
	for _, mission := range missions {
		requestMissions = append(requestMissions, admiral.Mission{
			ID:                        mission.ID,
			Title:                     mission.Title,
			DependsOn:                 append([]string(nil), mission.DependsOn...),
			UseCaseIDs:                append([]string(nil), mission.UseCaseIDs...),
			Classification:            mission.Classification,
			ClassificationRationale:   mission.ClassificationRationale,
			ClassificationCriteria:    append([]string(nil), mission.ClassificationCriteria...),
			ClassificationConfidence:  mission.ClassificationConfidence,
			ClassificationNeedsReview: mission.ClassificationNeedsReview,
		})
		missionIDs = append(missionIDs, mission.ID)
	}

	return admiral.ApprovalRequest{
		CommissionID:    commissionID,
		MissionManifest: requestMissions,
		WaveAssignments: []admiral.Wave{{
			Index:      waveIndex,
			MissionIDs: missionIDs,
		}},
		CoverageMap:   map[string]admiral.CoverageStatus{},
		Iteration:     1,
		MaxIterations: 1,
		WaveReview: &admiral.WaveReview{
			WaveIndex:  waveIndex,
			DemoTokens: demoTokens,
		},
	}
}
