package integration_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/ship-commander/sc3/internal/admiral"
	"github.com/ship-commander/sc3/internal/commander"
	"github.com/ship-commander/sc3/internal/commission"
	"github.com/ship-commander/sc3/internal/doctor"
	"github.com/ship-commander/sc3/internal/events"
	"github.com/ship-commander/sc3/internal/protocol"
	"github.com/ship-commander/sc3/internal/recovery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegrationHappyPathCommissionToCompletion(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	parsed, err := commission.ParseMarkdown(ctx, "Happy Path PRD", samplePRDMarkdown())
	require.NoError(t, err)
	require.NotEmpty(t, parsed.UseCases)
	require.NotEmpty(t, parsed.AcceptanceCriteria)

	mission := commander.Mission{
		ID:                 "mission-happy",
		Title:              "Happy mission",
		UseCaseIDs:         []string{parsed.UseCases[0].ID},
		AcceptanceCriteria: acDescriptions(parsed.AcceptanceCriteria),
	}

	store := &integrationManifestStore{manifest: []commander.Mission{mission}, ready: [][]string{{mission.ID}}}
	worktrees := newIntegrationWorktreeManager(t.TempDir(), map[string]string{mission.ID: "# demo token\n"})
	protocolStore := protocol.NewInMemoryStore()
	harness := &integrationHarness{
		protocolStore: protocolStore,
		verdicts: map[string][]integrationReviewVerdict{
			mission.ID: {{decision: protocol.ReviewVerdictApproved}},
		},
	}
	verifier := &integrationVerifier{}
	demoTokens := &integrationDemoTokenValidator{}
	approval := admiral.NewApprovalGate(2)
	approvalDone := startApprovalResponder(approval, []admiral.ApprovalResponse{{Decision: admiral.ApprovalDecisionApproved}})
	feedback := &integrationFeedbackInjector{}
	shelver := &integrationPlanShelver{}
	eventsPublisher := &integrationCommanderEventPublisher{}

	cmd, err := commander.New(
		store,
		worktrees,
		&integrationSurfaceLocker{},
		harness,
		verifier,
		demoTokens,
		approval,
		feedback,
		shelver,
		eventsPublisher,
		commander.CommanderConfig{WIPLimit: 1, ProtocolEventStore: protocolStore, ReviewPollInterval: time.Millisecond, ReviewTimeout: time.Second},
	)
	require.NoError(t, err)

	err = cmd.Execute(ctx, "commission-happy")
	require.NoError(t, err)
	require.NoError(t, waitResponder(approvalDone))

	assert.Equal(t, 1, harness.ImplementerDispatchCount())
	assert.Equal(t, 1, harness.ReviewerDispatchCount())
	assert.Equal(t, 0, demoTokens.CallCount(), "non STANDARD_OPS mission should not require demo token validator")
	assert.True(t, eventsPublisher.HasType(commander.EventMissionCompleted), "mission completion event should be emitted")
}

func TestIntegrationFeedbackLoopReconvenesThenExecutes(t *testing.T) {
	t.Parallel()

	mission := commander.Mission{ID: "mission-feedback", Title: "Feedback mission", UseCaseIDs: []string{"UC-1"}}
	store := &integrationManifestStore{manifest: []commander.Mission{mission}, ready: [][]string{{mission.ID}}}
	worktrees := newIntegrationWorktreeManager(t.TempDir(), map[string]string{mission.ID: "# demo token\n"})
	protocolStore := protocol.NewInMemoryStore()
	feedback := &integrationFeedbackInjector{}
	shelver := &integrationPlanShelver{}

	firstHarness := &integrationHarness{protocolStore: protocolStore}
	firstApproval := admiral.NewApprovalGate(2)
	firstDone := startApprovalResponder(firstApproval, []admiral.ApprovalResponse{{
		Decision:     admiral.ApprovalDecisionFeedback,
		FeedbackText: "Split mission into backend and ui tracks",
	}})

	firstCmd, err := commander.New(
		store,
		worktrees,
		&integrationSurfaceLocker{},
		firstHarness,
		&integrationVerifier{},
		&integrationDemoTokenValidator{},
		firstApproval,
		feedback,
		shelver,
		&integrationCommanderEventPublisher{},
		commander.CommanderConfig{WIPLimit: 1, ProtocolEventStore: protocolStore, ReviewPollInterval: time.Millisecond, ReviewTimeout: time.Second},
	)
	require.NoError(t, err)

	err = firstCmd.Execute(context.Background(), "commission-feedback")
	require.Error(t, err)
	assert.ErrorIs(t, err, commander.ErrApprovalFeedback)
	require.NoError(t, waitResponder(firstDone))
	assert.Equal(t, 1, feedback.CallCount())
	assert.Equal(t, 0, firstHarness.ImplementerDispatchCount(), "no dispatch should happen when admiral requests planning feedback")

	secondHarness := &integrationHarness{
		protocolStore: protocolStore,
		verdicts: map[string][]integrationReviewVerdict{
			mission.ID: {{decision: protocol.ReviewVerdictApproved}},
		},
	}
	secondApproval := admiral.NewApprovalGate(2)
	secondDone := startApprovalResponder(secondApproval, []admiral.ApprovalResponse{{Decision: admiral.ApprovalDecisionApproved}})

	secondCmd, err := commander.New(
		store,
		worktrees,
		&integrationSurfaceLocker{},
		secondHarness,
		&integrationVerifier{},
		&integrationDemoTokenValidator{},
		secondApproval,
		feedback,
		shelver,
		&integrationCommanderEventPublisher{},
		commander.CommanderConfig{WIPLimit: 1, ProtocolEventStore: protocolStore, ReviewPollInterval: time.Millisecond, ReviewTimeout: time.Second},
	)
	require.NoError(t, err)

	err = secondCmd.Execute(context.Background(), "commission-feedback")
	require.NoError(t, err)
	require.NoError(t, waitResponder(secondDone))
	assert.Equal(t, 1, secondHarness.ImplementerDispatchCount(), "execution should proceed after reconvened approval")
}

func TestIntegrationShelveAndResumeLifecycle(t *testing.T) {
	t.Parallel()

	mission := commander.Mission{ID: "mission-shelve", Title: "Shelve mission"}
	store := &integrationManifestStore{manifest: []commander.Mission{mission}, ready: [][]string{{mission.ID}}}
	worktrees := newIntegrationWorktreeManager(t.TempDir(), map[string]string{mission.ID: "# demo token\n"})
	protocolStore := protocol.NewInMemoryStore()
	feedback := &integrationFeedbackInjector{}
	shelver := &integrationPlanShelver{}

	firstHarness := &integrationHarness{protocolStore: protocolStore}
	firstApproval := admiral.NewApprovalGate(2)
	firstDone := startApprovalResponder(firstApproval, []admiral.ApprovalResponse{{
		Decision:     admiral.ApprovalDecisionShelved,
		FeedbackText: "Pause for dependency validation",
	}})

	firstCmd, err := commander.New(
		store,
		worktrees,
		&integrationSurfaceLocker{},
		firstHarness,
		&integrationVerifier{},
		&integrationDemoTokenValidator{},
		firstApproval,
		feedback,
		shelver,
		&integrationCommanderEventPublisher{},
		commander.CommanderConfig{WIPLimit: 1, ProtocolEventStore: protocolStore, ReviewPollInterval: time.Millisecond, ReviewTimeout: time.Second},
	)
	require.NoError(t, err)

	err = firstCmd.Execute(context.Background(), "commission-shelve")
	require.Error(t, err)
	assert.ErrorIs(t, err, commander.ErrApprovalShelved)
	require.NoError(t, waitResponder(firstDone))
	assert.Equal(t, 1, shelver.CallCount())
	assert.Equal(t, 0, firstHarness.ImplementerDispatchCount(), "shelved plan must not dispatch implementation")

	secondHarness := &integrationHarness{
		protocolStore: protocolStore,
		verdicts: map[string][]integrationReviewVerdict{
			mission.ID: {{decision: protocol.ReviewVerdictApproved}},
		},
	}
	secondApproval := admiral.NewApprovalGate(2)
	secondDone := startApprovalResponder(secondApproval, []admiral.ApprovalResponse{{Decision: admiral.ApprovalDecisionApproved}})

	secondCmd, err := commander.New(
		store,
		worktrees,
		&integrationSurfaceLocker{},
		secondHarness,
		&integrationVerifier{},
		&integrationDemoTokenValidator{},
		secondApproval,
		feedback,
		shelver,
		&integrationCommanderEventPublisher{},
		commander.CommanderConfig{WIPLimit: 1, ProtocolEventStore: protocolStore, ReviewPollInterval: time.Millisecond, ReviewTimeout: time.Second},
	)
	require.NoError(t, err)

	err = secondCmd.Execute(context.Background(), "commission-shelve")
	require.NoError(t, err)
	require.NoError(t, waitResponder(secondDone))
	assert.Equal(t, 1, secondHarness.ImplementerDispatchCount(), "resumed plan should dispatch implementation")
}

func TestIntegrationTerminationWhenMaxRevisionsExceeded(t *testing.T) {
	t.Parallel()

	mission := commander.Mission{ID: "mission-terminate", Title: "Terminate mission", MaxRevisions: 1}
	store := &integrationManifestStore{manifest: []commander.Mission{mission}, ready: [][]string{{mission.ID}}}
	worktrees := newIntegrationWorktreeManager(t.TempDir(), map[string]string{mission.ID: "# demo token\n"})
	protocolStore := protocol.NewInMemoryStore()
	harness := &integrationHarness{
		protocolStore: protocolStore,
		verdicts: map[string][]integrationReviewVerdict{
			mission.ID: {{decision: protocol.ReviewVerdictNeedsFixes, feedback: "needs refactor"}},
		},
	}
	eventsPublisher := &integrationCommanderEventPublisher{}
	approval := admiral.NewApprovalGate(2)
	approvalDone := startApprovalResponder(approval, []admiral.ApprovalResponse{{Decision: admiral.ApprovalDecisionApproved}})

	cmd, err := commander.New(
		store,
		worktrees,
		&integrationSurfaceLocker{},
		harness,
		&integrationVerifier{},
		&integrationDemoTokenValidator{},
		approval,
		&integrationFeedbackInjector{},
		&integrationPlanShelver{},
		eventsPublisher,
		commander.CommanderConfig{WIPLimit: 1, ProtocolEventStore: protocolStore, ReviewPollInterval: time.Millisecond, ReviewTimeout: time.Second},
	)
	require.NoError(t, err)

	err = cmd.Execute(context.Background(), "commission-terminate")
	require.Error(t, err)
	require.NoError(t, waitResponder(approvalDone))
	assert.Contains(t, err.Error(), "halted after review")
	assert.True(t, eventsPublisher.HasHaltReason(commander.HaltReasonMaxRevisionsExceeded), "max revision halt must be emitted")
}

func TestIntegrationStandardOpsFastPath(t *testing.T) {
	t.Parallel()

	mission := commander.Mission{
		ID:             "mission-standard",
		Title:          "Standard ops mission",
		Classification: commander.MissionClassificationStandardOps,
	}
	store := &integrationManifestStore{manifest: []commander.Mission{mission}, ready: [][]string{{mission.ID}}}
	worktrees := newIntegrationWorktreeManager(t.TempDir(), map[string]string{mission.ID: "# demo token\n"})
	protocolStore := protocol.NewInMemoryStore()
	harness := &integrationHarness{
		protocolStore: protocolStore,
		verdicts: map[string][]integrationReviewVerdict{
			mission.ID: {{decision: protocol.ReviewVerdictApproved}},
		},
	}
	verifier := &integrationVerifier{}
	demoTokens := &integrationDemoTokenValidator{}
	eventsPublisher := &integrationCommanderEventPublisher{}
	approval := admiral.NewApprovalGate(2)
	approvalDone := startApprovalResponder(approval, []admiral.ApprovalResponse{{Decision: admiral.ApprovalDecisionApproved}})

	cmd, err := commander.New(
		store,
		worktrees,
		&integrationSurfaceLocker{},
		harness,
		verifier,
		demoTokens,
		approval,
		&integrationFeedbackInjector{},
		&integrationPlanShelver{},
		eventsPublisher,
		commander.CommanderConfig{WIPLimit: 1, ProtocolEventStore: protocolStore, ReviewPollInterval: time.Millisecond, ReviewTimeout: time.Second},
	)
	require.NoError(t, err)

	err = cmd.Execute(context.Background(), "commission-standard")
	require.NoError(t, err)
	require.NoError(t, waitResponder(approvalDone))

	assert.Equal(t, 0, verifier.VerifyCallCount(), "STANDARD_OPS should bypass RED_ALERT verifier")
	assert.Equal(t, 1, verifier.VerifyImplementCallCount(), "STANDARD_OPS should use VerifyImplement")
	assert.Equal(t, 1, demoTokens.CallCount(), "STANDARD_OPS should validate demo token")
	assert.True(t, eventsPublisher.HasType(commander.EventMissionCompleted))
}

func TestIntegrationMultiWaveExecutionWithReviewCheckpoint(t *testing.T) {
	t.Parallel()

	missionOne := commander.Mission{ID: "mission-wave-1", Title: "Wave one"}
	missionTwo := commander.Mission{ID: "mission-wave-2", Title: "Wave two", DependsOn: []string{missionOne.ID}}
	store := &integrationManifestStore{
		manifest: []commander.Mission{missionOne, missionTwo},
		ready:    [][]string{{missionOne.ID}, {missionTwo.ID}},
	}
	worktrees := newIntegrationWorktreeManager(t.TempDir(), map[string]string{
		missionOne.ID: "# demo wave 1\n",
		missionTwo.ID: "# demo wave 2\n",
	})
	protocolStore := protocol.NewInMemoryStore()
	harness := &integrationHarness{
		protocolStore: protocolStore,
		verdicts: map[string][]integrationReviewVerdict{
			missionOne.ID: {{decision: protocol.ReviewVerdictApproved}},
			missionTwo.ID: {{decision: protocol.ReviewVerdictApproved}},
		},
	}
	eventsPublisher := &integrationCommanderEventPublisher{}
	approval := admiral.NewApprovalGate(3)
	approvalDone := startApprovalResponder(approval, []admiral.ApprovalResponse{
		{Decision: admiral.ApprovalDecisionApproved},
		{Decision: admiral.ApprovalDecisionFeedback, FeedbackText: "carry wave checkpoint feedback into next mission"},
	})

	cmd, err := commander.New(
		store,
		worktrees,
		&integrationSurfaceLocker{},
		harness,
		&integrationVerifier{},
		&integrationDemoTokenValidator{},
		approval,
		&integrationFeedbackInjector{},
		&integrationPlanShelver{},
		eventsPublisher,
		commander.CommanderConfig{WIPLimit: 1, ProtocolEventStore: protocolStore, ReviewPollInterval: time.Millisecond, ReviewTimeout: time.Second},
	)
	require.NoError(t, err)

	err = cmd.Execute(context.Background(), "commission-wave")
	require.NoError(t, err)
	require.NoError(t, waitResponder(approvalDone))

	require.Equal(t, 2, harness.ImplementerDispatchCount())
	dispatches := harness.ImplementerDispatches()
	require.Len(t, dispatches, 2)
	assert.Equal(t, missionOne.ID, dispatches[0].Mission.ID)
	assert.Equal(t, missionTwo.ID, dispatches[1].Mission.ID)
	assert.Equal(t, "carry wave checkpoint feedback into next mission", dispatches[1].WaveFeedback)
	assert.True(t, eventsPublisher.HasType(commander.EventWaveFeedbackRecorded), "wave feedback should be recorded between waves")
	assert.Len(t, approval.History(), 2, "initial approval and one inter-wave review should be recorded")
}

func TestIntegrationDoctorStuckDetection(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 2, 11, 4, 0, 0, 0, time.UTC)
	store := &integrationDoctorStore{
		snapshot: doctor.Snapshot{
			Missions: []doctor.Mission{{ID: "mission-stuck", State: "in_progress", AgentID: "agent-1"}},
			Agents: []doctor.Agent{{
				ID:            "agent-1",
				State:         "running",
				SessionID:     "session-missing",
				LastHeartbeat: now.Add(-10 * time.Minute),
			}},
		},
	}
	sessions := &integrationDoctorSessions{active: map[string]struct{}{"zombie-session": {}}}
	bus := &integrationEventBus{}

	manager, err := doctor.NewManager(store, sessions, bus, doctor.Config{
		HeartbeatInterval: time.Second,
		StuckTimeout:      time.Minute,
	})
	require.NoError(t, err)
	managerNowOverride(manager, now)

	report, err := manager.RunOnce(context.Background())
	require.NoError(t, err)

	assert.Equal(t, 1, report.StuckAgents)
	assert.Equal(t, 1, report.OrphanedMissions)
	assert.Equal(t, 1, report.ZombieSessions)
	assert.Equal(t, []string{"agent-1"}, store.stuckAgents)
	assert.Equal(t, []string{"mission-stuck"}, store.backlogMissions)
	assert.Equal(t, []string{"zombie-session"}, sessions.cleaned)
	assert.True(t, bus.HasEventType(events.EventTypeHealthCheck), "health check event should be emitted")
}

func TestIntegrationCrashRecoveryResume(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 2, 11, 5, 0, 0, 0, time.UTC)
	store := &integrationRecoveryStore{
		snapshot: recovery.Snapshot{
			Commissions: []recovery.Commission{
				{ID: "commission-1", State: recovery.CommissionExecuting},
				{ID: "commission-2", State: "paused"},
			},
			Missions: []recovery.Mission{
				{ID: "mission-1", CommissionID: "commission-1", State: recovery.MissionInProgress, AgentID: "agent-1"},
				{ID: "mission-2", CommissionID: "commission-1", State: recovery.MissionDone, AgentID: "agent-2"},
			},
			Agents: []recovery.Agent{
				{ID: "agent-1", State: recovery.AgentRunning, SessionID: "session-dead"},
				{ID: "agent-2", State: recovery.AgentRunning, SessionID: "session-live"},
			},
		},
	}
	sessions := &integrationRecoverySessions{active: map[string]struct{}{"session-live": {}}}
	bus := &integrationEventBus{}

	manager, err := recovery.NewManager(store, sessions, recovery.Config{ResumeTimeout: 2 * time.Second, EventBus: bus})
	require.NoError(t, err)
	recoveryNowOverride(manager, now)

	result, err := manager.Recover(context.Background())
	require.NoError(t, err)

	assert.Equal(t, []string{"mission-1"}, result.OrphanedMissionIDs)
	assert.Equal(t, []string{"session-dead"}, result.CleanedDeadSessions)
	assert.Equal(t, []string{"commission-1"}, result.ResumeCommissionIDs)
	assert.Equal(t, []string{"mission-1"}, store.backlogMissions)
	assert.Equal(t, []string{"agent-1"}, store.deadAgents)
	assert.Equal(t, []string{"session-dead"}, sessions.cleaned)
	assert.True(t, bus.HasEventType(events.EventTypeHealthCheck), "recovery summary event should be emitted")
}

func samplePRDMarkdown() string {
	return `# Ship Commander Integration

## Use Cases

| UC ID | Title | Description |
| --- | --- | --- |
| UC-1 | Execute mission lifecycle | Parse PRD, approve manifest, execute and complete |

## Acceptance Criteria

- [ ] Commander executes deterministic mission lifecycle
- [ ] Admiral approval gate blocks execution until decision
- [ ] Mission completion emits observable event
`
}

func acDescriptions(criteria []commission.AC) []string {
	out := make([]string, 0, len(criteria))
	for _, criterion := range criteria {
		if criterion.Description == "" {
			continue
		}
		out = append(out, criterion.Description)
	}
	return out
}

type integrationManifestStore struct {
	manifest []commander.Mission
	ready    [][]string

	mu        sync.Mutex
	readyCall int
}

func (s *integrationManifestStore) ReadApprovedManifest(_ context.Context, _ string) ([]commander.Mission, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]commander.Mission, len(s.manifest))
	copy(out, s.manifest)
	return out, nil
}

func (s *integrationManifestStore) ReadyMissionIDs(_ context.Context, _ string) ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.readyCall++
	if len(s.ready) == 0 {
		return nil, nil
	}
	idx := s.readyCall - 1
	if idx >= len(s.ready) {
		idx = len(s.ready) - 1
	}
	out := make([]string, len(s.ready[idx]))
	copy(out, s.ready[idx])
	return out, nil
}

type integrationWorktreeManager struct {
	root       string
	demoTokens map[string]string

	mu      sync.Mutex
	created []string
}

func newIntegrationWorktreeManager(root string, demoTokens map[string]string) *integrationWorktreeManager {
	return &integrationWorktreeManager{root: root, demoTokens: demoTokens}
}

func (m *integrationWorktreeManager) Create(_ context.Context, mission commander.Mission) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.created = append(m.created, mission.ID)
	path := filepath.Join(m.root, mission.ID)
	if err := os.MkdirAll(filepath.Join(path, "demo"), 0o750); err != nil {
		return "", fmt.Errorf("create worktree dirs: %w", err)
	}
	if token := m.demoTokens[mission.ID]; token != "" {
		filename := filepath.Join(path, "demo", fmt.Sprintf("MISSION-%s.md", mission.ID))
		if err := os.WriteFile(filename, []byte(token), 0o600); err != nil {
			return "", fmt.Errorf("write demo token: %w", err)
		}
	}
	return path, nil
}

type integrationSurfaceLocker struct{}

func (integrationSurfaceLocker) Acquire(_ context.Context, _ string, _ []string) (func() error, error) {
	return func() error { return nil }, nil
}

type integrationReviewVerdict struct {
	decision string
	feedback string
}

type integrationHarness struct {
	protocolStore *protocol.InMemoryStore
	verdicts      map[string][]integrationReviewVerdict

	mu                    sync.Mutex
	implementerDispatches []commander.DispatchRequest
	reviewerDispatches    []commander.ReviewerDispatchRequest
	implementerCounter    map[string]int
	reviewerCounter       map[string]int
}

func (h *integrationHarness) DispatchImplementer(_ context.Context, req commander.DispatchRequest) (commander.DispatchResult, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.implementerCounter == nil {
		h.implementerCounter = map[string]int{}
	}
	h.implementerCounter[req.Mission.ID]++
	h.implementerDispatches = append(h.implementerDispatches, req)

	sessionID := fmt.Sprintf("implementer-%s-%d", req.Mission.ID, h.implementerCounter[req.Mission.ID])
	return commander.DispatchResult{SessionID: sessionID}, nil
}

func (h *integrationHarness) DispatchReviewer(ctx context.Context, req commander.ReviewerDispatchRequest) (commander.DispatchResult, error) {
	h.mu.Lock()
	if h.reviewerCounter == nil {
		h.reviewerCounter = map[string]int{}
	}
	h.reviewerCounter[req.Mission.ID]++
	h.reviewerDispatches = append(h.reviewerDispatches, req)

	reviewerSession := fmt.Sprintf("reviewer-%s-%d", req.Mission.ID, h.reviewerCounter[req.Mission.ID])
	verdict := integrationReviewVerdict{decision: protocol.ReviewVerdictApproved}
	if queue := h.verdicts[req.Mission.ID]; len(queue) > 0 {
		verdict = queue[0]
		h.verdicts[req.Mission.ID] = queue[1:]
	}
	h.mu.Unlock()

	if h.protocolStore != nil {
		payload, err := json.Marshal(map[string]string{
			"verdict":                verdict.decision,
			"feedback":               verdict.feedback,
			"implementer_session_id": req.ImplementerSessionID,
			"reviewer_session_id":    reviewerSession,
		})
		if err != nil {
			return commander.DispatchResult{}, fmt.Errorf("marshal verdict payload: %w", err)
		}
		if err := h.protocolStore.Append(ctx, protocol.ProtocolEvent{
			Type:      protocol.EventTypeReviewComplete,
			MissionID: req.Mission.ID,
			Payload:   payload,
			Timestamp: time.Now().UTC(),
		}); err != nil {
			return commander.DispatchResult{}, fmt.Errorf("append review verdict event: %w", err)
		}
	}

	return commander.DispatchResult{SessionID: reviewerSession}, nil
}

func (h *integrationHarness) ImplementerDispatchCount() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.implementerDispatches)
}

func (h *integrationHarness) ReviewerDispatchCount() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.reviewerDispatches)
}

func (h *integrationHarness) ImplementerDispatches() []commander.DispatchRequest {
	h.mu.Lock()
	defer h.mu.Unlock()
	out := make([]commander.DispatchRequest, len(h.implementerDispatches))
	copy(out, h.implementerDispatches)
	return out
}

type integrationVerifier struct {
	verifyErr          error
	verifyImplementErr error

	mu                   sync.Mutex
	verifyCalls          int
	verifyImplementCalls int
}

func (v *integrationVerifier) Verify(_ context.Context, _ commander.Mission, _ string) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.verifyCalls++
	return v.verifyErr
}

func (v *integrationVerifier) VerifyImplement(_ context.Context, _ commander.Mission, _ string) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.verifyImplementCalls++
	return v.verifyImplementErr
}

func (v *integrationVerifier) VerifyCallCount() int {
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.verifyCalls
}

func (v *integrationVerifier) VerifyImplementCallCount() int {
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.verifyImplementCalls
}

type integrationDemoTokenValidator struct {
	err error

	mu    sync.Mutex
	calls int
}

func (v *integrationDemoTokenValidator) Validate(_ context.Context, _ commander.Mission, _ string) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.calls++
	return v.err
}

func (v *integrationDemoTokenValidator) CallCount() int {
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.calls
}

type integrationFeedbackInjector struct {
	mu           sync.Mutex
	callCount    int
	commissionID string
	feedbackText string
}

func (f *integrationFeedbackInjector) InjectPlanningFeedback(_ context.Context, commissionID, feedbackText string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.callCount++
	f.commissionID = commissionID
	f.feedbackText = feedbackText
	return nil
}

func (f *integrationFeedbackInjector) CallCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.callCount
}

type integrationPlanShelver struct {
	mu           sync.Mutex
	callCount    int
	commissionID string
	feedbackText string
}

func (s *integrationPlanShelver) ShelvePlan(_ context.Context, commissionID, feedbackText string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.callCount++
	s.commissionID = commissionID
	s.feedbackText = feedbackText
	return nil
}

func (s *integrationPlanShelver) CallCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.callCount
}

type integrationCommanderEventPublisher struct {
	mu     sync.Mutex
	events []commander.Event
}

func (p *integrationCommanderEventPublisher) Publish(_ context.Context, event commander.Event) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.events = append(p.events, event)
	return nil
}

func (p *integrationCommanderEventPublisher) HasType(eventType string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, event := range p.events {
		if event.Type == eventType {
			return true
		}
	}
	return false
}

func (p *integrationCommanderEventPublisher) HasHaltReason(reason commander.HaltReason) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, event := range p.events {
		if event.Type == commander.EventMissionHalted && event.Reason == reason {
			return true
		}
	}
	return false
}

type integrationDoctorStore struct {
	snapshot doctor.Snapshot

	mu              sync.Mutex
	backlogMissions []string
	stuckAgents     []string
}

func (s *integrationDoctorStore) LoadSnapshot(_ context.Context) (doctor.Snapshot, error) {
	return s.snapshot, nil
}

func (s *integrationDoctorStore) SetMissionBacklog(_ context.Context, missionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.backlogMissions = append(s.backlogMissions, missionID)
	return nil
}

func (s *integrationDoctorStore) SetAgentStuck(_ context.Context, agentID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stuckAgents = append(s.stuckAgents, agentID)
	return nil
}

type integrationDoctorSessions struct {
	active map[string]struct{}

	mu      sync.Mutex
	cleaned []string
}

func (s *integrationDoctorSessions) ActiveSessions(_ context.Context) (map[string]struct{}, error) {
	out := make(map[string]struct{}, len(s.active))
	for id := range s.active {
		out[id] = struct{}{}
	}
	return out, nil
}

func (s *integrationDoctorSessions) CleanupDeadSession(_ context.Context, sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cleaned = append(s.cleaned, sessionID)
	return nil
}

type integrationRecoveryStore struct {
	snapshot recovery.Snapshot

	mu              sync.Mutex
	backlogMissions []string
	deadAgents      []string
}

func (s *integrationRecoveryStore) LoadSnapshot(_ context.Context) (recovery.Snapshot, error) {
	return s.snapshot, nil
}

func (s *integrationRecoveryStore) SetMissionBacklog(_ context.Context, missionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.backlogMissions = append(s.backlogMissions, missionID)
	return nil
}

func (s *integrationRecoveryStore) SetAgentDead(_ context.Context, agentID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.deadAgents = append(s.deadAgents, agentID)
	return nil
}

type integrationRecoverySessions struct {
	active map[string]struct{}

	mu      sync.Mutex
	cleaned []string
}

func (s *integrationRecoverySessions) ActiveSessions(_ context.Context) (map[string]struct{}, error) {
	out := make(map[string]struct{}, len(s.active))
	for id := range s.active {
		out[id] = struct{}{}
	}
	return out, nil
}

func (s *integrationRecoverySessions) CleanupDeadSession(_ context.Context, sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cleaned = append(s.cleaned, sessionID)
	return nil
}

type integrationEventBus struct {
	mu     sync.Mutex
	events []events.Event
}

func (b *integrationEventBus) Publish(event events.Event) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.events = append(b.events, event)
}

func (b *integrationEventBus) HasEventType(eventType string) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, event := range b.events {
		if event.Type == eventType {
			return true
		}
	}
	return false
}

func startApprovalResponder(
	gate *admiral.ApprovalGate,
	responses []admiral.ApprovalResponse,
) <-chan error {
	done := make(chan error, 1)
	go func() {
		for i, response := range responses {
			select {
			case <-gate.Requests():
			case <-time.After(2 * time.Second):
				done <- fmt.Errorf("timed out waiting for approval request %d", i+1)
				return
			}
			if err := gate.Respond(response); err != nil {
				done <- fmt.Errorf("respond approval %d: %w", i+1, err)
				return
			}
		}
		done <- nil
	}()
	return done
}

func waitResponder(done <-chan error) error {
	select {
	case err := <-done:
		return err
	case <-time.After(3 * time.Second):
		return errors.New("approval responder did not complete")
	}
}

func managerNowOverride(manager *doctor.Manager, now time.Time) {
	// The package does not export clock injection directly; this helper intentionally uses runtime behavior only.
	_ = manager
	_ = now
}

func recoveryNowOverride(manager *recovery.Manager, now time.Time) {
	// The package does not export clock injection directly; this helper intentionally uses runtime behavior only.
	_ = manager
	_ = now
}
