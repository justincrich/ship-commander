package commander

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/ship-commander/sc3/internal/admiral"
)

const (
	// EventMissionCompleted is emitted when a mission passes verification.
	EventMissionCompleted = "MISSION_COMPLETED"
	// EventMissionHalted is emitted when a mission fails dispatch or verification.
	EventMissionHalted = "MISSION_HALTED"
	// MissionClassificationStandardOps routes mission execution through the standard implementation fast path.
	MissionClassificationStandardOps = "STANDARD_OPS"
	// DefaultMaxRevisions is the deterministic default revision ceiling before halting.
	DefaultMaxRevisions = 3
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
	ID             string
	Title          string
	Classification string
	DependsOn      []string
	UseCaseIDs     []string
	SurfaceArea    []string
	RevisionCount  int
	MaxRevisions   int
	// ACAttemptsExhausted indicates all AC attempts failed and mission must halt deterministically.
	ACAttemptsExhausted bool
	// ManualHalt requests deterministic dispatch stop before running mission work.
	ManualHalt bool
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

// CommanderConfig configures commander runtime behavior.
type CommanderConfig struct {
	WIPLimit int
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
	wipLimit      int
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
		wipLimit:      cfg.WIPLimit,
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

	for i, wave := range waves {
		if err := c.executeWave(ctx, commissionID, i+1, wave); err != nil {
			return fmt.Errorf("execute wave %d: %w", i+1, err)
		}
	}

	return nil
}

func (c *Commander) executeWave(ctx context.Context, commissionID string, waveIndex int, missions []Mission) error {
	if len(missions) == 0 {
		return nil
	}

	pending := make(map[string]Mission, len(missions))
	order := make([]string, 0, len(missions))
	for _, mission := range missions {
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
		_ = c.publishHalt(ctx, waveIndex, mission.ID, reason, message)
		return fmt.Errorf("mission %s halted before dispatch: %s", mission.ID, message)
	}

	worktreePath, err := c.worktrees.Create(ctx, mission)
	if err != nil {
		_ = c.publishHalt(ctx, waveIndex, mission.ID, HaltReasonManualHalt, fmt.Sprintf("worktree creation failed: %v", err))
		return fmt.Errorf("create worktree for %s: %w", mission.ID, err)
	}

	release, err := c.locks.Acquire(ctx, mission.ID, mission.SurfaceArea)
	if err != nil {
		_ = c.publishHalt(ctx, waveIndex, mission.ID, HaltReasonManualHalt, fmt.Sprintf("surface-area lock failed: %v", err))
		return fmt.Errorf("acquire lock for %s: %w", mission.ID, err)
	}
	defer func() {
		_ = release()
	}()

	if _, err := c.harness.DispatchImplementer(ctx, DispatchRequest{
		Mission:      mission,
		WorktreePath: worktreePath,
	}); err != nil {
		_ = c.publishHalt(ctx, waveIndex, mission.ID, HaltReasonManualHalt, fmt.Sprintf("dispatch failed: %v", err))
		return fmt.Errorf("dispatch implementer for %s: %w", mission.ID, err)
	}

	if isStandardOpsMission(mission) {
		if err := c.verifier.VerifyImplement(ctx, mission, worktreePath); err != nil {
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
	} else {
		if err := c.verifier.Verify(ctx, mission, worktreePath); err != nil {
			_ = c.publishHalt(ctx, waveIndex, mission.ID, HaltReasonManualHalt, fmt.Sprintf("verification failed: %v", err))
			return fmt.Errorf("verify mission %s: %w", mission.ID, err)
		}
	}

	if err := c.publish(ctx, Event{
		Type:      EventMissionCompleted,
		MissionID: mission.ID,
		WaveIndex: waveIndex,
		Timestamp: c.now().UTC(),
		Message:   "mission verified successfully",
	}); err != nil {
		return fmt.Errorf("publish completion event for %s: %w", mission.ID, err)
	}

	return nil
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
			ID:         mission.ID,
			Title:      mission.Title,
			DependsOn:  append([]string(nil), mission.DependsOn...),
			UseCaseIDs: append([]string(nil), mission.UseCaseIDs...),
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
