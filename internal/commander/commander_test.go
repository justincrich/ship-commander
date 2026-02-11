package commander

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/ship-commander/sc3/internal/admiral"
	"github.com/ship-commander/sc3/internal/protocol"
)

func TestComputeWaves(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		missions  []Mission
		wantWaves [][]string
		wantErr   bool
	}{
		{
			name: "dependency fanout",
			missions: []Mission{
				{ID: "m1", Title: "first"},
				{ID: "m2", Title: "second", DependsOn: []string{"m1"}},
				{ID: "m3", Title: "third", DependsOn: []string{"m1"}},
			},
			wantWaves: [][]string{{"m1"}, {"m2", "m3"}},
		},
		{
			name: "independent missions share one wave",
			missions: []Mission{
				{ID: "m1", Title: "first"},
				{ID: "m2", Title: "second"},
			},
			wantWaves: [][]string{{"m1", "m2"}},
		},
		{
			name: "dependency cycle returns error",
			missions: []Mission{
				{ID: "m1", Title: "first", DependsOn: []string{"m2"}},
				{ID: "m2", Title: "second", DependsOn: []string{"m1"}},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := ComputeWaves(tt.missions)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("compute waves: %v", err)
			}

			gotIDs := make([][]string, 0, len(got))
			for _, wave := range got {
				ids := make([]string, 0, len(wave))
				for _, mission := range wave {
					ids = append(ids, mission.ID)
				}
				gotIDs = append(gotIDs, ids)
			}

			if !reflect.DeepEqual(gotIDs, tt.wantWaves) {
				t.Fatalf("waves = %v, want %v", gotIDs, tt.wantWaves)
			}
		})
	}
}

func TestGitWorktreeManagerCreate(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	runner := &fakeShellRunner{}
	manager := newGitWorktreeManagerForTest(root, runner)

	mission := Mission{
		ID:    "ship-commander-3-s6s.1",
		Title: "Commander Orchestrator",
	}
	path, err := manager.Create(context.Background(), mission)
	if err != nil {
		t.Fatalf("create worktree: %v", err)
	}

	wantPath := filepath.Join(root, ".beads", "worktrees", "MISSION-ship-commander-3-s6s-1")
	if path != wantPath {
		t.Fatalf("worktree path = %q, want %q", path, wantPath)
	}

	if runner.name != "git" {
		t.Fatalf("command name = %q, want git", runner.name)
	}
	wantArgs := []string{
		"worktree", "add", wantPath, "-b", "feature/MISSION-ship-commander-3-s6s-1-commander-orchestrator",
	}
	if !reflect.DeepEqual(runner.args, wantArgs) {
		t.Fatalf("args = %v, want %v", runner.args, wantArgs)
	}
}

func TestCommanderExecuteSingleMissionFlow(t *testing.T) {
	t.Parallel()

	sequence := make([]string, 0)
	store := &fakeManifestStore{
		manifest: []Mission{{ID: "m1", Title: "Mission One", SurfaceArea: []string{"internal/commander/**"}}},
		ready:    [][]string{{"m1"}},
	}
	worktrees := &fakeWorktreeManager{paths: map[string]string{"m1": "/tmp/worktree/m1"}}
	locks := &fakeSurfaceLocker{sequence: &sequence}
	harness := &fakeHarness{sequence: &sequence}
	verifier := &fakeVerifier{}
	demoTokens := &fakeDemoTokenValidator{}
	events := &fakeEventPublisher{}

	cmd, err := newCommanderForTest(store, worktrees, locks, harness, verifier, demoTokens, events, CommanderConfig{WIPLimit: 2})
	if err != nil {
		t.Fatalf("new commander: %v", err)
	}

	if err := cmd.Execute(context.Background(), "commission-1"); err != nil {
		t.Fatalf("execute: %v", err)
	}

	if store.readManifestCalls != 1 {
		t.Fatalf("read manifest calls = %d, want 1", store.readManifestCalls)
	}
	if len(worktrees.created) != 1 || worktrees.created[0] != "m1" {
		t.Fatalf("worktrees created = %v, want [m1]", worktrees.created)
	}
	if !reflect.DeepEqual(sequence, []string{"lock:m1", "dispatch:m1", "review:m1"}) {
		t.Fatalf("call sequence = %v, want lock before dispatch", sequence)
	}
	if len(events.events) != 1 || events.events[0].Type != EventMissionCompleted {
		t.Fatalf("events = %v, want one %s", events.events, EventMissionCompleted)
	}
	if demoTokens.CallCount() != 0 {
		t.Fatalf("demo token calls = %d, want 0 for non-standard ops mission", demoTokens.CallCount())
	}
}

func TestCommanderExecuteRequiresApprovalBeforeDispatch(t *testing.T) {
	t.Parallel()

	store := &fakeManifestStore{
		manifest: []Mission{
			{
				ID:                        "m1",
				Title:                     "Mission One",
				UseCaseIDs:                []string{"UC-1"},
				Classification:            MissionClassificationREDAlert,
				ClassificationRationale:   "Touches execution behavior",
				ClassificationCriteria:    []string{"business_logic"},
				ClassificationConfidence:  "high",
				ClassificationNeedsReview: false,
			},
		},
		ready: [][]string{{"m1"}},
	}
	worktrees := &fakeWorktreeManager{paths: map[string]string{"m1": "/tmp/worktree/m1"}}
	locks := &fakeSurfaceLocker{}
	harness := &fakeHarness{}
	verifier := &fakeVerifier{}
	demoTokens := &fakeDemoTokenValidator{}
	events := &fakeEventPublisher{}
	approval := &fakeApprovalGate{
		response: admiral.ApprovalResponse{Decision: admiral.ApprovalDecisionApproved},
	}
	feedback := &fakeFeedbackInjector{}
	shelver := &fakePlanShelver{}

	cmd, err := New(
		store,
		worktrees,
		locks,
		harness,
		verifier,
		demoTokens,
		approval,
		feedback,
		shelver,
		events,
		CommanderConfig{WIPLimit: 1},
	)
	if err != nil {
		t.Fatalf("new commander: %v", err)
	}

	if err := cmd.Execute(context.Background(), "commission-1"); err != nil {
		t.Fatalf("execute: %v", err)
	}

	if approval.callCount != 1 {
		t.Fatalf("approval calls = %d, want 1", approval.callCount)
	}
	if len(approval.lastRequest.MissionManifest) != 1 || approval.lastRequest.MissionManifest[0].ID != "m1" {
		t.Fatalf("approval mission manifest = %+v, want mission m1", approval.lastRequest.MissionManifest)
	}
	if approval.lastRequest.MissionManifest[0].Classification != MissionClassificationREDAlert {
		t.Fatalf(
			"approval mission classification = %q, want %q",
			approval.lastRequest.MissionManifest[0].Classification,
			MissionClassificationREDAlert,
		)
	}
	if approval.lastRequest.MissionManifest[0].ClassificationRationale == "" {
		t.Fatal("approval mission rationale should not be empty")
	}
	if len(approval.lastRequest.WaveAssignments) != 1 || approval.lastRequest.WaveAssignments[0].Index != 1 {
		t.Fatalf("approval wave assignments = %+v, want one wave", approval.lastRequest.WaveAssignments)
	}
	if approval.lastRequest.CoverageMap["UC-1"] != admiral.CoverageStatusCovered {
		t.Fatalf("coverage map = %+v, expected UC-1 covered", approval.lastRequest.CoverageMap)
	}
	if len(worktrees.created) != 1 || worktrees.created[0] != "m1" {
		t.Fatalf("worktrees created = %v, want [m1]", worktrees.created)
	}
}

func TestCommanderExecuteFeedbackReconvenesPlanningWithoutDispatch(t *testing.T) {
	t.Parallel()

	store := &fakeManifestStore{
		manifest: []Mission{{ID: "m1", Title: "Mission One"}},
		ready:    [][]string{{"m1"}},
	}
	worktrees := &fakeWorktreeManager{paths: map[string]string{"m1": "/tmp/worktree/m1"}}
	locks := &fakeSurfaceLocker{}
	harness := &fakeHarness{}
	verifier := &fakeVerifier{}
	demoTokens := &fakeDemoTokenValidator{}
	events := &fakeEventPublisher{}
	approval := &fakeApprovalGate{
		response: admiral.ApprovalResponse{
			Decision:     admiral.ApprovalDecisionFeedback,
			FeedbackText: "split mission into backend and tui slices",
		},
	}
	feedback := &fakeFeedbackInjector{}
	shelver := &fakePlanShelver{}

	cmd, err := New(
		store,
		worktrees,
		locks,
		harness,
		verifier,
		demoTokens,
		approval,
		feedback,
		shelver,
		events,
		CommanderConfig{WIPLimit: 1},
	)
	if err != nil {
		t.Fatalf("new commander: %v", err)
	}

	err = cmd.Execute(context.Background(), "commission-1")
	if !errors.Is(err, ErrApprovalFeedback) {
		t.Fatalf("execute error = %v, want ErrApprovalFeedback", err)
	}
	if feedback.callCount != 1 {
		t.Fatalf("feedback injector calls = %d, want 1", feedback.callCount)
	}
	if feedback.lastFeedback != "split mission into backend and tui slices" {
		t.Fatalf("feedback text = %q", feedback.lastFeedback)
	}
	if len(worktrees.created) != 0 {
		t.Fatalf("worktrees created = %v, want none when feedback requested", worktrees.created)
	}
	if shelver.callCount != 0 {
		t.Fatalf("shelver calls = %d, want 0", shelver.callCount)
	}
}

func TestCommanderExecuteShelvePersistsPlanWithoutDispatch(t *testing.T) {
	t.Parallel()

	store := &fakeManifestStore{
		manifest: []Mission{{ID: "m1", Title: "Mission One"}},
		ready:    [][]string{{"m1"}},
	}
	worktrees := &fakeWorktreeManager{paths: map[string]string{"m1": "/tmp/worktree/m1"}}
	locks := &fakeSurfaceLocker{}
	harness := &fakeHarness{}
	verifier := &fakeVerifier{}
	demoTokens := &fakeDemoTokenValidator{}
	events := &fakeEventPublisher{}
	approval := &fakeApprovalGate{
		response: admiral.ApprovalResponse{
			Decision:     admiral.ApprovalDecisionShelved,
			FeedbackText: "hold until dependencies land",
		},
	}
	feedback := &fakeFeedbackInjector{}
	shelver := &fakePlanShelver{}

	cmd, err := New(
		store,
		worktrees,
		locks,
		harness,
		verifier,
		demoTokens,
		approval,
		feedback,
		shelver,
		events,
		CommanderConfig{WIPLimit: 1},
	)
	if err != nil {
		t.Fatalf("new commander: %v", err)
	}

	err = cmd.Execute(context.Background(), "commission-1")
	if !errors.Is(err, ErrApprovalShelved) {
		t.Fatalf("execute error = %v, want ErrApprovalShelved", err)
	}
	if shelver.callCount != 1 {
		t.Fatalf("shelver calls = %d, want 1", shelver.callCount)
	}
	if shelver.lastFeedback != "hold until dependencies land" {
		t.Fatalf("shelve feedback text = %q", shelver.lastFeedback)
	}
	if len(worktrees.created) != 0 {
		t.Fatalf("worktrees created = %v, want none when shelved", worktrees.created)
	}
	if feedback.callCount != 0 {
		t.Fatalf("feedback injector calls = %d, want 0", feedback.callCount)
	}
}

func TestCommanderExecutePublishesHaltedOnVerifyFailure(t *testing.T) {
	t.Parallel()

	store := &fakeManifestStore{
		manifest: []Mission{{ID: "m1", Title: "Mission One"}},
		ready:    [][]string{{"m1"}},
	}
	worktrees := &fakeWorktreeManager{paths: map[string]string{"m1": "/tmp/worktree/m1"}}
	locks := &fakeSurfaceLocker{}
	harness := &fakeHarness{}
	verifier := &fakeVerifier{verifyErr: errors.New("verification failed")}
	demoTokens := &fakeDemoTokenValidator{}
	events := &fakeEventPublisher{}

	cmd, err := newCommanderForTest(store, worktrees, locks, harness, verifier, demoTokens, events, CommanderConfig{WIPLimit: 1})
	if err != nil {
		t.Fatalf("new commander: %v", err)
	}

	if err := cmd.Execute(context.Background(), "commission-1"); err == nil {
		t.Fatal("expected execute error, got nil")
	}

	if len(events.events) == 0 {
		t.Fatal("expected halted event, got none")
	}
	if events.events[0].Type != EventMissionHalted {
		t.Fatalf("first event = %s, want %s", events.events[0].Type, EventMissionHalted)
	}
	if events.events[0].Reason != HaltReasonManualHalt {
		t.Fatalf("halt reason = %s, want %s", events.events[0].Reason, HaltReasonManualHalt)
	}
	if !events.events[0].NotifyTUI {
		t.Fatal("expected TUI notification on halted mission event")
	}
}

func TestCommanderExecuteEnforcesWIPLimit(t *testing.T) {
	t.Parallel()

	store := &fakeManifestStore{
		manifest: []Mission{
			{ID: "m1", Title: "Mission One"},
			{ID: "m2", Title: "Mission Two"},
			{ID: "m3", Title: "Mission Three"},
		},
		ready: [][]string{
			{"m1", "m2", "m3"},
			{"m1", "m2", "m3"},
		},
	}
	worktrees := &fakeWorktreeManager{
		paths: map[string]string{
			"m1": "/tmp/worktree/m1",
			"m2": "/tmp/worktree/m2",
			"m3": "/tmp/worktree/m3",
		},
	}
	locks := &fakeSurfaceLocker{}
	harness := &fakeHarness{delay: 30 * time.Millisecond}
	verifier := &fakeVerifier{}
	demoTokens := &fakeDemoTokenValidator{}
	events := &fakeEventPublisher{}

	cmd, err := newCommanderForTest(store, worktrees, locks, harness, verifier, demoTokens, events, CommanderConfig{WIPLimit: 2})
	if err != nil {
		t.Fatalf("new commander: %v", err)
	}

	if err := cmd.Execute(context.Background(), "commission-1"); err != nil {
		t.Fatalf("execute: %v", err)
	}

	if harness.maxConcurrent > 2 {
		t.Fatalf("max concurrent dispatches = %d, want <= 2", harness.maxConcurrent)
	}
	if store.readyCalls < 2 {
		t.Fatalf("ready calls = %d, want at least 2 for propulsion loop advance", store.readyCalls)
	}
}

func TestCommanderExecuteUsesDependencyOrderAcrossWaves(t *testing.T) {
	t.Parallel()

	m1Path := filepath.Join(t.TempDir(), "m1")
	if err := os.MkdirAll(filepath.Join(m1Path, "demo"), 0o750); err != nil {
		t.Fatalf("create m1 demo dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(m1Path, "demo", "MISSION-m1.md"), []byte("# demo evidence"), 0o600); err != nil {
		t.Fatalf("write m1 demo token: %v", err)
	}
	m2Path := filepath.Join(t.TempDir(), "m2")

	sequence := make([]string, 0)
	store := &fakeManifestStore{
		manifest: []Mission{
			{ID: "m1", Title: "First"},
			{ID: "m2", Title: "Second", DependsOn: []string{"m1"}},
		},
		ready: [][]string{
			{"m1", "m2"},
			{"m1", "m2"},
		},
	}
	worktrees := &fakeWorktreeManager{
		paths: map[string]string{
			"m1": m1Path,
			"m2": m2Path,
		},
	}
	locks := &fakeSurfaceLocker{}
	harness := &fakeHarness{sequence: &sequence}
	verifier := &fakeVerifier{}
	demoTokens := &fakeDemoTokenValidator{}
	events := &fakeEventPublisher{}

	cmd, err := newCommanderForTest(store, worktrees, locks, harness, verifier, demoTokens, events, CommanderConfig{WIPLimit: 2})
	if err != nil {
		t.Fatalf("new commander: %v", err)
	}

	if err := cmd.Execute(context.Background(), "commission-1"); err != nil {
		t.Fatalf("execute: %v", err)
	}

	if len(sequence) != 4 {
		t.Fatalf("dispatch sequence = %v, want dispatch/review for two missions", sequence)
	}
	if sequence[0] != "dispatch:m1" || sequence[1] != "review:m1" || sequence[2] != "dispatch:m2" || sequence[3] != "review:m2" {
		t.Fatalf(
			"dispatch sequence = %v, want [dispatch:m1 review:m1 dispatch:m2 review:m2]",
			sequence,
		)
	}
}

func TestCommanderExecuteTriggersWaveReviewCheckpointAndContinues(t *testing.T) {
	t.Parallel()

	m1Path := filepath.Join(t.TempDir(), "m1")
	if err := os.MkdirAll(filepath.Join(m1Path, "demo"), 0o750); err != nil {
		t.Fatalf("create m1 demo dir: %v", err)
	}
	m1Evidence := "# MISSION-m1 demo evidence"
	if err := os.WriteFile(filepath.Join(m1Path, "demo", "MISSION-m1.md"), []byte(m1Evidence), 0o600); err != nil {
		t.Fatalf("write m1 demo token: %v", err)
	}

	store := &fakeManifestStore{
		manifest: []Mission{
			{ID: "m1", Title: "First"},
			{ID: "m2", Title: "Second", DependsOn: []string{"m1"}},
		},
		ready: [][]string{
			{"m1", "m2"},
			{"m1", "m2"},
		},
	}
	worktrees := &fakeWorktreeManager{
		paths: map[string]string{
			"m1": m1Path,
			"m2": filepath.Join(t.TempDir(), "m2"),
		},
	}
	locks := &fakeSurfaceLocker{}
	harness := &fakeHarness{}
	verifier := &fakeVerifier{}
	demoTokens := &fakeDemoTokenValidator{}
	events := &fakeEventPublisher{}
	approval := &fakeApprovalGate{
		responses: []admiral.ApprovalResponse{
			{Decision: admiral.ApprovalDecisionApproved},
			{Decision: admiral.ApprovalDecisionApproved},
		},
	}
	feedback := &fakeFeedbackInjector{}
	shelver := &fakePlanShelver{}

	cmd, err := New(
		store,
		worktrees,
		locks,
		harness,
		verifier,
		demoTokens,
		approval,
		feedback,
		shelver,
		events,
		CommanderConfig{WIPLimit: 2},
	)
	if err != nil {
		t.Fatalf("new commander: %v", err)
	}

	if err := cmd.Execute(context.Background(), "commission-1"); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if approval.callCount != 2 {
		t.Fatalf("approval calls = %d, want 2 (manifest + wave review)", approval.callCount)
	}
	if len(approval.requests) != 2 {
		t.Fatalf("approval requests = %d, want 2", len(approval.requests))
	}
	waveReviewReq := approval.requests[1]
	if waveReviewReq.WaveReview == nil {
		t.Fatal("wave review request should include WaveReview payload")
	}
	if waveReviewReq.WaveReview.WaveIndex != 1 {
		t.Fatalf("wave review index = %d, want 1", waveReviewReq.WaveReview.WaveIndex)
	}
	if got := waveReviewReq.WaveReview.DemoTokens["m1"]; got != m1Evidence {
		t.Fatalf("wave review demo token for m1 = %q, want %q", got, m1Evidence)
	}
	if len(harness.implementerDispatches) != 2 {
		t.Fatalf("implementer dispatches = %d, want 2 (wave2 should continue)", len(harness.implementerDispatches))
	}
}

func TestCommanderExecuteInjectsWaveFeedbackIntoNextWaveDispatch(t *testing.T) {
	t.Parallel()

	m1Path := filepath.Join(t.TempDir(), "m1")
	if err := os.MkdirAll(filepath.Join(m1Path, "demo"), 0o750); err != nil {
		t.Fatalf("create m1 demo dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(m1Path, "demo", "MISSION-m1.md"), []byte("# m1 demo evidence"), 0o600); err != nil {
		t.Fatalf("write m1 demo token: %v", err)
	}

	store := &fakeManifestStore{
		manifest: []Mission{
			{ID: "m1", Title: "First"},
			{ID: "m2", Title: "Second", DependsOn: []string{"m1"}},
		},
		ready: [][]string{
			{"m1", "m2"},
			{"m1", "m2"},
		},
	}
	worktrees := &fakeWorktreeManager{
		paths: map[string]string{
			"m1": m1Path,
			"m2": filepath.Join(t.TempDir(), "m2"),
		},
	}
	locks := &fakeSurfaceLocker{}
	harness := &fakeHarness{}
	verifier := &fakeVerifier{}
	demoTokens := &fakeDemoTokenValidator{}
	events := &fakeEventPublisher{}
	approval := &fakeApprovalGate{
		responses: []admiral.ApprovalResponse{
			{Decision: admiral.ApprovalDecisionApproved},
			{Decision: admiral.ApprovalDecisionFeedback, FeedbackText: "focus on reliability checks"},
		},
	}
	feedback := &fakeFeedbackInjector{}
	shelver := &fakePlanShelver{}

	cmd, err := New(
		store,
		worktrees,
		locks,
		harness,
		verifier,
		demoTokens,
		approval,
		feedback,
		shelver,
		events,
		CommanderConfig{WIPLimit: 2},
	)
	if err != nil {
		t.Fatalf("new commander: %v", err)
	}

	if err := cmd.Execute(context.Background(), "commission-1"); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if len(harness.implementerDispatches) != 2 {
		t.Fatalf("implementer dispatches = %d, want 2", len(harness.implementerDispatches))
	}
	if harness.implementerDispatches[1].WaveFeedback != "focus on reliability checks" {
		t.Fatalf("wave feedback = %q, want propagated feedback", harness.implementerDispatches[1].WaveFeedback)
	}

	foundWaveFeedbackEvent := false
	for _, event := range events.events {
		if event.Type == EventWaveFeedbackRecorded && event.WaveIndex == 1 {
			foundWaveFeedbackEvent = true
			break
		}
	}
	if !foundWaveFeedbackEvent {
		t.Fatal("expected wave feedback event to be published")
	}
}

func TestCommanderExecuteHaltsOnWaveReviewHaltDecision(t *testing.T) {
	t.Parallel()

	m1Path := filepath.Join(t.TempDir(), "m1")
	if err := os.MkdirAll(filepath.Join(m1Path, "demo"), 0o750); err != nil {
		t.Fatalf("create m1 demo dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(m1Path, "demo", "MISSION-m1.md"), []byte("# m1 demo evidence"), 0o600); err != nil {
		t.Fatalf("write m1 demo token: %v", err)
	}

	store := &fakeManifestStore{
		manifest: []Mission{
			{ID: "m1", Title: "First"},
			{ID: "m2", Title: "Second", DependsOn: []string{"m1"}},
		},
		ready: [][]string{
			{"m1", "m2"},
			{"m1", "m2"},
		},
	}
	worktrees := &fakeWorktreeManager{
		paths: map[string]string{
			"m1": m1Path,
			"m2": filepath.Join(t.TempDir(), "m2"),
		},
	}
	locks := &fakeSurfaceLocker{}
	harness := &fakeHarness{}
	verifier := &fakeVerifier{}
	demoTokens := &fakeDemoTokenValidator{}
	events := &fakeEventPublisher{}
	approval := &fakeApprovalGate{
		responses: []admiral.ApprovalResponse{
			{Decision: admiral.ApprovalDecisionApproved},
			{Decision: admiral.ApprovalDecisionHalted, FeedbackText: "stop after wave one"},
		},
	}
	feedback := &fakeFeedbackInjector{}
	shelver := &fakePlanShelver{}

	cmd, err := New(
		store,
		worktrees,
		locks,
		harness,
		verifier,
		demoTokens,
		approval,
		feedback,
		shelver,
		events,
		CommanderConfig{WIPLimit: 2},
	)
	if err != nil {
		t.Fatalf("new commander: %v", err)
	}

	if err := cmd.Execute(context.Background(), "commission-1"); err == nil {
		t.Fatal("expected halt error from wave review decision")
	}
	if len(harness.implementerDispatches) != 1 {
		t.Fatalf("implementer dispatches = %d, want 1 (wave2 should not run)", len(harness.implementerDispatches))
	}

	foundHaltEvent := false
	for _, event := range events.events {
		if event.Type == EventCommissionHalted && event.WaveIndex == 1 {
			foundHaltEvent = true
			break
		}
	}
	if !foundHaltEvent {
		t.Fatal("expected commission halted event from wave review decision")
	}
}

func TestCommanderExecuteDispatchesReviewerWithContextAndWaitsForVerdict(t *testing.T) {
	t.Parallel()

	store := &fakeManifestStore{
		manifest: []Mission{{
			ID:                 "m1",
			Title:              "Mission One",
			AcceptanceCriteria: []string{"AC-1", "AC-2"},
		}},
		ready: [][]string{{"m1"}},
	}
	worktrees := &fakeWorktreeManager{paths: map[string]string{"m1": "/tmp/worktree/m1"}}
	locks := &fakeSurfaceLocker{}
	harness := &fakeHarness{
		implementerSessionIDs: []string{"impl-1"},
		reviewerSessionIDs:    []string{"rev-1"},
	}
	verifier := &fakeVerifier{}
	demoTokens := &fakeDemoTokenValidator{}
	events := &fakeEventPublisher{}
	protocolStore := &fakeProtocolEventStore{
		responses: [][]protocol.ProtocolEvent{
			{{
				Type:      protocol.EventTypeGateResult,
				MissionID: "m1",
				Payload:   json.RawMessage(`{"gate":"go test ./...","result":"pass"}`),
				Timestamp: time.Date(2026, 2, 11, 12, 0, 0, 0, time.UTC),
			}},
			{},
			{reviewCompleteEvent("m1", "APPROVED", "impl-1", "rev-1", "looks good")},
		},
	}

	cmd, err := newCommanderForTest(
		store,
		worktrees,
		locks,
		harness,
		verifier,
		demoTokens,
		events,
		CommanderConfig{
			WIPLimit:           1,
			ProtocolEventStore: protocolStore,
			ReviewPollInterval: 1 * time.Millisecond,
			ReviewTimeout:      200 * time.Millisecond,
		},
	)
	if err != nil {
		t.Fatalf("new commander: %v", err)
	}

	if err := cmd.Execute(context.Background(), "commission-1"); err != nil {
		t.Fatalf("execute: %v", err)
	}

	if len(harness.reviewerDispatches) != 1 {
		t.Fatalf("reviewer dispatch count = %d, want 1", len(harness.reviewerDispatches))
	}
	reviewerReq := harness.reviewerDispatches[0]
	if reviewerReq.ImplementerSessionID != "impl-1" {
		t.Fatalf("reviewer implementer session = %q, want impl-1", reviewerReq.ImplementerSessionID)
	}
	if !reviewerReq.ReadOnlyWorktree {
		t.Fatal("reviewer request should enforce read-only worktree")
	}
	if reviewerReq.IncludeImplementerReasoning {
		t.Fatal("reviewer request should not include implementer reasoning")
	}
	if !reflect.DeepEqual(reviewerReq.AcceptanceCriteria, []string{"AC-1", "AC-2"}) {
		t.Fatalf("acceptance criteria = %v, want [AC-1 AC-2]", reviewerReq.AcceptanceCriteria)
	}
	if len(reviewerReq.GateEvidence) != 1 {
		t.Fatalf("gate evidence count = %d, want 1", len(reviewerReq.GateEvidence))
	}
	if protocolStore.calls < 3 {
		t.Fatalf("protocol store calls = %d, want at least 3 to prove polling", protocolStore.calls)
	}
	if len(events.events) != 1 || events.events[0].Type != EventMissionCompleted {
		t.Fatalf("events = %v, want one %s", events.events, EventMissionCompleted)
	}
}

func TestCommanderExecuteNeedsFixesRedispatchesImplementerWithFeedback(t *testing.T) {
	t.Parallel()

	store := &fakeManifestStore{
		manifest: []Mission{{ID: "m1", Title: "Mission One", MaxRevisions: 3}},
		ready:    [][]string{{"m1"}},
	}
	worktrees := &fakeWorktreeManager{paths: map[string]string{"m1": "/tmp/worktree/m1"}}
	locks := &fakeSurfaceLocker{}
	harness := &fakeHarness{
		implementerSessionIDs: []string{"impl-1", "impl-2"},
		reviewerSessionIDs:    []string{"rev-1", "rev-2"},
	}
	verifier := &fakeVerifier{}
	demoTokens := &fakeDemoTokenValidator{}
	events := &fakeEventPublisher{}
	protocolStore := &fakeProtocolEventStore{
		responses: [][]protocol.ProtocolEvent{
			{},
			{reviewCompleteEvent("m1", "NEEDS_FIXES", "impl-1", "rev-1", "add edge-case guard")},
			{},
			{reviewCompleteEvent("m1", "APPROVED", "impl-2", "rev-2", "resolved")},
		},
	}

	cmd, err := newCommanderForTest(
		store,
		worktrees,
		locks,
		harness,
		verifier,
		demoTokens,
		events,
		CommanderConfig{
			WIPLimit:           1,
			ProtocolEventStore: protocolStore,
			ReviewPollInterval: 1 * time.Millisecond,
			ReviewTimeout:      300 * time.Millisecond,
		},
	)
	if err != nil {
		t.Fatalf("new commander: %v", err)
	}

	if err := cmd.Execute(context.Background(), "commission-1"); err != nil {
		t.Fatalf("execute: %v", err)
	}

	if len(harness.implementerDispatches) != 2 {
		t.Fatalf("implementer dispatches = %d, want 2", len(harness.implementerDispatches))
	}
	if harness.implementerDispatches[1].ReviewerFeedback != "add edge-case guard" {
		t.Fatalf("second dispatch feedback = %q, want propagated reviewer feedback", harness.implementerDispatches[1].ReviewerFeedback)
	}
	if len(events.events) != 1 || events.events[0].Type != EventMissionCompleted {
		t.Fatalf("events = %v, want one %s", events.events, EventMissionCompleted)
	}
}

func TestCommanderExecuteNeedsFixesHaltsWhenMaxRevisionsReached(t *testing.T) {
	t.Parallel()

	store := &fakeManifestStore{
		manifest: []Mission{{ID: "m1", Title: "Mission One", RevisionCount: 2, MaxRevisions: 3}},
		ready:    [][]string{{"m1"}},
	}
	worktrees := &fakeWorktreeManager{paths: map[string]string{"m1": "/tmp/worktree/m1"}}
	locks := &fakeSurfaceLocker{}
	harness := &fakeHarness{
		implementerSessionIDs: []string{"impl-1"},
		reviewerSessionIDs:    []string{"rev-1"},
	}
	verifier := &fakeVerifier{}
	demoTokens := &fakeDemoTokenValidator{}
	events := &fakeEventPublisher{}
	protocolStore := &fakeProtocolEventStore{
		responses: [][]protocol.ProtocolEvent{
			{},
			{reviewCompleteEvent("m1", "NEEDS_FIXES", "impl-1", "rev-1", "still broken")},
		},
	}

	cmd, err := newCommanderForTest(
		store,
		worktrees,
		locks,
		harness,
		verifier,
		demoTokens,
		events,
		CommanderConfig{
			WIPLimit:           1,
			ProtocolEventStore: protocolStore,
			ReviewPollInterval: 1 * time.Millisecond,
			ReviewTimeout:      300 * time.Millisecond,
		},
	)
	if err != nil {
		t.Fatalf("new commander: %v", err)
	}

	if err := cmd.Execute(context.Background(), "commission-1"); err == nil {
		t.Fatal("expected execute error when max revisions reached")
	}
	if len(events.events) == 0 || events.events[0].Type != EventMissionHalted {
		t.Fatalf("events = %v, want first event %s", events.events, EventMissionHalted)
	}
	if events.events[0].Reason != HaltReasonMaxRevisionsExceeded {
		t.Fatalf("halt reason = %s, want %s", events.events[0].Reason, HaltReasonMaxRevisionsExceeded)
	}
}

func TestCommanderExecuteReviewerMustDifferFromImplementer(t *testing.T) {
	t.Parallel()

	store := &fakeManifestStore{
		manifest: []Mission{{ID: "m1", Title: "Mission One"}},
		ready:    [][]string{{"m1"}},
	}
	worktrees := &fakeWorktreeManager{paths: map[string]string{"m1": "/tmp/worktree/m1"}}
	locks := &fakeSurfaceLocker{}
	harness := &fakeHarness{
		implementerSessionIDs: []string{"shared-session"},
		reviewerSessionIDs:    []string{"shared-session"},
	}
	verifier := &fakeVerifier{}
	demoTokens := &fakeDemoTokenValidator{}
	events := &fakeEventPublisher{}

	cmd, err := newCommanderForTest(store, worktrees, locks, harness, verifier, demoTokens, events, CommanderConfig{WIPLimit: 1})
	if err != nil {
		t.Fatalf("new commander: %v", err)
	}

	if err := cmd.Execute(context.Background(), "commission-1"); err == nil {
		t.Fatal("expected execute error for same-session reviewer/implementer")
	}
	if len(events.events) == 0 || events.events[0].Type != EventMissionHalted {
		t.Fatalf("events = %v, want first event %s", events.events, EventMissionHalted)
	}
}

func TestCommanderExecuteStandardOpsUsesVerifyImplementAndDemoToken(t *testing.T) {
	t.Parallel()

	store := &fakeManifestStore{
		manifest: []Mission{{ID: "m1", Title: "Mission One", Classification: MissionClassificationStandardOps}},
		ready:    [][]string{{"m1"}},
	}
	worktrees := &fakeWorktreeManager{paths: map[string]string{"m1": "/tmp/worktree/m1"}}
	locks := &fakeSurfaceLocker{}
	harness := &fakeHarness{}
	verifier := &fakeVerifier{}
	demoTokens := &fakeDemoTokenValidator{}
	events := &fakeEventPublisher{}

	cmd, err := newCommanderForTest(store, worktrees, locks, harness, verifier, demoTokens, events, CommanderConfig{WIPLimit: 1})
	if err != nil {
		t.Fatalf("new commander: %v", err)
	}

	if err := cmd.Execute(context.Background(), "commission-1"); err != nil {
		t.Fatalf("execute: %v", err)
	}

	if verifier.VerifyCallCount() != 0 {
		t.Fatalf("verify calls = %d, want 0", verifier.VerifyCallCount())
	}
	if verifier.VerifyImplementCallCount() != 1 {
		t.Fatalf("verify implement calls = %d, want 1", verifier.VerifyImplementCallCount())
	}
	if demoTokens.CallCount() != 1 {
		t.Fatalf("demo token calls = %d, want 1", demoTokens.CallCount())
	}
}

func TestCommanderExecuteStandardOpsHaltsOnVerifyImplementFailure(t *testing.T) {
	t.Parallel()

	store := &fakeManifestStore{
		manifest: []Mission{{ID: "m1", Title: "Mission One", Classification: MissionClassificationStandardOps}},
		ready:    [][]string{{"m1"}},
	}
	worktrees := &fakeWorktreeManager{paths: map[string]string{"m1": "/tmp/worktree/m1"}}
	locks := &fakeSurfaceLocker{}
	harness := &fakeHarness{}
	verifier := &fakeVerifier{verifyImplementErr: errors.New("verify implement failed")}
	demoTokens := &fakeDemoTokenValidator{}
	events := &fakeEventPublisher{}

	cmd, err := newCommanderForTest(store, worktrees, locks, harness, verifier, demoTokens, events, CommanderConfig{WIPLimit: 1})
	if err != nil {
		t.Fatalf("new commander: %v", err)
	}

	if err := cmd.Execute(context.Background(), "commission-1"); err == nil {
		t.Fatal("expected execute error, got nil")
	}
	if demoTokens.CallCount() != 0 {
		t.Fatalf("demo token calls = %d, want 0 when verify implement fails", demoTokens.CallCount())
	}
	if len(events.events) == 0 || events.events[0].Type != EventMissionHalted {
		t.Fatalf("events = %v, want first event %s", events.events, EventMissionHalted)
	}
}

func TestCommanderExecuteStandardOpsHaltsOnDemoTokenFailure(t *testing.T) {
	t.Parallel()

	store := &fakeManifestStore{
		manifest: []Mission{{ID: "m1", Title: "Mission One", Classification: MissionClassificationStandardOps}},
		ready:    [][]string{{"m1"}},
	}
	worktrees := &fakeWorktreeManager{paths: map[string]string{"m1": "/tmp/worktree/m1"}}
	locks := &fakeSurfaceLocker{}
	harness := &fakeHarness{}
	verifier := &fakeVerifier{}
	demoTokens := &fakeDemoTokenValidator{err: errors.New("demo token invalid")}
	events := &fakeEventPublisher{}

	cmd, err := newCommanderForTest(store, worktrees, locks, harness, verifier, demoTokens, events, CommanderConfig{WIPLimit: 1})
	if err != nil {
		t.Fatalf("new commander: %v", err)
	}

	if err := cmd.Execute(context.Background(), "commission-1"); err == nil {
		t.Fatal("expected execute error, got nil")
	}
	if verifier.VerifyImplementCallCount() != 1 {
		t.Fatalf("verify implement calls = %d, want 1", verifier.VerifyImplementCallCount())
	}
	if len(events.events) == 0 || events.events[0].Type != EventMissionHalted {
		t.Fatalf("events = %v, want first event %s", events.events, EventMissionHalted)
	}
	if events.events[0].Reason != HaltReasonDemoTokenInvalid {
		t.Fatalf("halt reason = %s, want %s", events.events[0].Reason, HaltReasonDemoTokenInvalid)
	}
	if !events.events[0].NotifyTUI {
		t.Fatal("expected TUI notification on halted mission event")
	}
}

func TestCommanderExecuteHaltsBeforeDispatchWhenRevisionLimitReached(t *testing.T) {
	t.Parallel()

	store := &fakeManifestStore{
		manifest: []Mission{{ID: "m1", Title: "Mission One", RevisionCount: 3, MaxRevisions: 3}},
		ready:    [][]string{{"m1"}},
	}
	worktrees := &fakeWorktreeManager{paths: map[string]string{"m1": "/tmp/worktree/m1"}}
	locks := &fakeSurfaceLocker{}
	harness := &fakeHarness{}
	verifier := &fakeVerifier{}
	demoTokens := &fakeDemoTokenValidator{}
	events := &fakeEventPublisher{}

	cmd, err := newCommanderForTest(store, worktrees, locks, harness, verifier, demoTokens, events, CommanderConfig{WIPLimit: 1})
	if err != nil {
		t.Fatalf("new commander: %v", err)
	}

	if err := cmd.Execute(context.Background(), "commission-1"); err == nil {
		t.Fatal("expected execute error, got nil")
	}

	if len(worktrees.created) != 0 {
		t.Fatalf("worktrees created = %v, want none because mission halts before dispatch", worktrees.created)
	}
	if len(events.events) == 0 || events.events[0].Type != EventMissionHalted {
		t.Fatalf("events = %v, want first event %s", events.events, EventMissionHalted)
	}
	if events.events[0].Reason != HaltReasonMaxRevisionsExceeded {
		t.Fatalf("halt reason = %s, want %s", events.events[0].Reason, HaltReasonMaxRevisionsExceeded)
	}
	if !events.events[0].NotifyTUI {
		t.Fatal("expected TUI notification on halted mission event")
	}
}

func TestCommanderExecuteHaltsBeforeDispatchWhenACAttemptsExhausted(t *testing.T) {
	t.Parallel()

	store := &fakeManifestStore{
		manifest: []Mission{{ID: "m1", Title: "Mission One", ACAttemptsExhausted: true}},
		ready:    [][]string{{"m1"}},
	}
	worktrees := &fakeWorktreeManager{paths: map[string]string{"m1": "/tmp/worktree/m1"}}
	locks := &fakeSurfaceLocker{}
	harness := &fakeHarness{}
	verifier := &fakeVerifier{}
	demoTokens := &fakeDemoTokenValidator{}
	events := &fakeEventPublisher{}

	cmd, err := newCommanderForTest(store, worktrees, locks, harness, verifier, demoTokens, events, CommanderConfig{WIPLimit: 1})
	if err != nil {
		t.Fatalf("new commander: %v", err)
	}

	if err := cmd.Execute(context.Background(), "commission-1"); err == nil {
		t.Fatal("expected execute error, got nil")
	}

	if len(worktrees.created) != 0 {
		t.Fatalf("worktrees created = %v, want none because mission halts before dispatch", worktrees.created)
	}
	if len(events.events) == 0 || events.events[0].Type != EventMissionHalted {
		t.Fatalf("events = %v, want first event %s", events.events, EventMissionHalted)
	}
	if events.events[0].Reason != HaltReasonACExhausted {
		t.Fatalf("halt reason = %s, want %s", events.events[0].Reason, HaltReasonACExhausted)
	}
	if !events.events[0].NotifyTUI {
		t.Fatal("expected TUI notification on halted mission event")
	}
}

func TestCommanderExecuteStandardOpsHaltsOnMissingDemoToken(t *testing.T) {
	t.Parallel()

	store := &fakeManifestStore{
		manifest: []Mission{{ID: "m1", Title: "Mission One", Classification: MissionClassificationStandardOps}},
		ready:    [][]string{{"m1"}},
	}
	worktrees := &fakeWorktreeManager{paths: map[string]string{"m1": "/tmp/worktree/m1"}}
	locks := &fakeSurfaceLocker{}
	harness := &fakeHarness{}
	verifier := &fakeVerifier{}
	demoTokens := &fakeDemoTokenValidator{err: os.ErrNotExist}
	events := &fakeEventPublisher{}

	cmd, err := newCommanderForTest(store, worktrees, locks, harness, verifier, demoTokens, events, CommanderConfig{WIPLimit: 1})
	if err != nil {
		t.Fatalf("new commander: %v", err)
	}

	if err := cmd.Execute(context.Background(), "commission-1"); err == nil {
		t.Fatal("expected execute error, got nil")
	}
	if verifier.VerifyImplementCallCount() != 1 {
		t.Fatalf("verify implement calls = %d, want 1", verifier.VerifyImplementCallCount())
	}
	if len(events.events) == 0 || events.events[0].Type != EventMissionHalted {
		t.Fatalf("events = %v, want first event %s", events.events, EventMissionHalted)
	}
	if events.events[0].Reason != HaltReasonDemoTokenMissing {
		t.Fatalf("halt reason = %s, want %s", events.events[0].Reason, HaltReasonDemoTokenMissing)
	}
	if !events.events[0].NotifyTUI {
		t.Fatal("expected TUI notification on halted mission event")
	}
}

type fakeManifestStore struct {
	manifest          []Mission
	ready             [][]string
	readManifestCalls int
	readyCalls        int
	mu                sync.Mutex
}

func newCommanderForTest(
	store ManifestStore,
	worktrees WorktreeManager,
	locks SurfaceLocker,
	harness Harness,
	verifier Verifier,
	demoTokens DemoTokenValidator,
	events EventPublisher,
	cfg CommanderConfig,
) (*Commander, error) {
	return New(
		store,
		worktrees,
		locks,
		harness,
		verifier,
		demoTokens,
		&fakeApprovalGate{
			response: admiral.ApprovalResponse{Decision: admiral.ApprovalDecisionApproved},
		},
		&fakeFeedbackInjector{},
		&fakePlanShelver{},
		events,
		cfg,
	)
}

func (f *fakeManifestStore) ReadApprovedManifest(_ context.Context, _ string) ([]Mission, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.readManifestCalls++
	out := make([]Mission, len(f.manifest))
	copy(out, f.manifest)
	return out, nil
}

func (f *fakeManifestStore) ReadyMissionIDs(_ context.Context, _ string) ([]string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.readyCalls++
	if len(f.ready) == 0 {
		return []string{}, nil
	}
	idx := f.readyCalls - 1
	if idx >= len(f.ready) {
		idx = len(f.ready) - 1
	}
	out := make([]string, len(f.ready[idx]))
	copy(out, f.ready[idx])
	return out, nil
}

type fakeWorktreeManager struct {
	paths   map[string]string
	created []string
	mu      sync.Mutex
}

func (f *fakeWorktreeManager) Create(_ context.Context, mission Mission) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.created = append(f.created, mission.ID)
	if path, ok := f.paths[mission.ID]; ok {
		return path, nil
	}
	return "/tmp/worktree/" + mission.ID, nil
}

type fakeSurfaceLocker struct {
	sequence *[]string
}

func (f *fakeSurfaceLocker) Acquire(_ context.Context, missionID string, _ []string) (func() error, error) {
	if f.sequence != nil {
		*f.sequence = append(*f.sequence, "lock:"+missionID)
	}
	return func() error { return nil }, nil
}

type fakeHarness struct {
	sequence      *[]string
	delay         time.Duration
	current       int
	maxConcurrent int
	dispatchErr   error
	reviewErr     error

	implementerSessionIDs []string
	reviewerSessionIDs    []string
	implementerDispatches []DispatchRequest
	reviewerDispatches    []ReviewerDispatchRequest

	mu sync.Mutex
}

func (f *fakeHarness) DispatchImplementer(_ context.Context, req DispatchRequest) (DispatchResult, error) {
	f.mu.Lock()
	if f.sequence != nil {
		*f.sequence = append(*f.sequence, "dispatch:"+req.Mission.ID)
	}
	f.current++
	if f.current > f.maxConcurrent {
		f.maxConcurrent = f.current
	}
	f.implementerDispatches = append(f.implementerDispatches, req)
	if f.dispatchErr != nil {
		f.current--
		f.mu.Unlock()
		return DispatchResult{}, f.dispatchErr
	}
	sessionID := "session-" + req.Mission.ID
	if len(f.implementerSessionIDs) > 0 {
		sessionID = f.implementerSessionIDs[0]
		f.implementerSessionIDs = f.implementerSessionIDs[1:]
	}
	f.mu.Unlock()

	if f.delay > 0 {
		time.Sleep(f.delay)
	}

	f.mu.Lock()
	f.current--
	f.mu.Unlock()

	return DispatchResult{SessionID: sessionID}, nil
}

func (f *fakeHarness) DispatchReviewer(_ context.Context, req ReviewerDispatchRequest) (DispatchResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.sequence != nil {
		*f.sequence = append(*f.sequence, "review:"+req.Mission.ID)
	}
	f.reviewerDispatches = append(f.reviewerDispatches, req)
	if f.reviewErr != nil {
		return DispatchResult{}, f.reviewErr
	}
	sessionID := "review-session-" + req.Mission.ID
	if len(f.reviewerSessionIDs) > 0 {
		sessionID = f.reviewerSessionIDs[0]
		f.reviewerSessionIDs = f.reviewerSessionIDs[1:]
	}
	return DispatchResult{SessionID: sessionID}, nil
}

type fakeVerifier struct {
	verifyErr            error
	verifyImplementErr   error
	verifyCalls          int
	verifyImplementCalls int
	mu                   sync.Mutex
}

func (f *fakeVerifier) Verify(_ context.Context, _ Mission, _ string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.verifyCalls++
	return f.verifyErr
}

func (f *fakeVerifier) VerifyImplement(_ context.Context, _ Mission, _ string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.verifyImplementCalls++
	return f.verifyImplementErr
}

func (f *fakeVerifier) VerifyCallCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.verifyCalls
}

func (f *fakeVerifier) VerifyImplementCallCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.verifyImplementCalls
}

type fakeDemoTokenValidator struct {
	err   error
	calls int
	mu    sync.Mutex
}

type fakeApprovalGate struct {
	response    admiral.ApprovalResponse
	responses   []admiral.ApprovalResponse
	err         error
	callCount   int
	lastRequest admiral.ApprovalRequest
	requests    []admiral.ApprovalRequest
	mu          sync.Mutex
}

func (f *fakeApprovalGate) AwaitDecision(
	_ context.Context,
	request admiral.ApprovalRequest,
) (admiral.ApprovalResponse, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.callCount++
	f.lastRequest = request
	f.requests = append(f.requests, request)
	if f.err != nil {
		return admiral.ApprovalResponse{}, f.err
	}
	if len(f.responses) > 0 {
		response := f.responses[0]
		f.responses = f.responses[1:]
		return response, nil
	}
	return f.response, nil
}

type fakeFeedbackInjector struct {
	callCount    int
	lastCID      string
	lastFeedback string
	err          error
	mu           sync.Mutex
}

func (f *fakeFeedbackInjector) InjectPlanningFeedback(_ context.Context, commissionID, feedbackText string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.callCount++
	f.lastCID = commissionID
	f.lastFeedback = feedbackText
	return f.err
}

type fakePlanShelver struct {
	callCount    int
	lastCID      string
	lastFeedback string
	err          error
	mu           sync.Mutex
}

func (f *fakePlanShelver) ShelvePlan(_ context.Context, commissionID, feedbackText string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.callCount++
	f.lastCID = commissionID
	f.lastFeedback = feedbackText
	return f.err
}

func (f *fakeDemoTokenValidator) Validate(_ context.Context, _ Mission, _ string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.calls++
	return f.err
}

func (f *fakeDemoTokenValidator) CallCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.calls
}

type fakeEventPublisher struct {
	events []Event
	mu     sync.Mutex
}

func (f *fakeEventPublisher) Publish(_ context.Context, event Event) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.events = append(f.events, event)
	return nil
}

type fakeShellRunner struct {
	dir  string
	name string
	args []string
}

type fakeProtocolEventStore struct {
	responses [][]protocol.ProtocolEvent
	calls     int
	listErr   error
	mu        sync.Mutex
}

func (f *fakeProtocolEventStore) ListByMission(_ context.Context, _ string) ([]protocol.ProtocolEvent, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.listErr != nil {
		return nil, f.listErr
	}
	f.calls++
	if len(f.responses) == 0 {
		return nil, nil
	}
	index := f.calls - 1
	if index >= len(f.responses) {
		index = len(f.responses) - 1
	}
	events := f.responses[index]
	out := make([]protocol.ProtocolEvent, len(events))
	copy(out, events)
	return out, nil
}

func reviewCompleteEvent(
	missionID string,
	verdict string,
	implementerSessionID string,
	reviewerSessionID string,
	feedback string,
) protocol.ProtocolEvent {
	return protocol.ProtocolEvent{
		Type:      protocol.EventTypeReviewComplete,
		MissionID: missionID,
		Payload: json.RawMessage(
			fmt.Sprintf(
				`{"verdict":"%s","implementer_session_id":"%s","reviewer_session_id":"%s","feedback":"%s"}`,
				verdict,
				implementerSessionID,
				reviewerSessionID,
				feedback,
			),
		),
		Timestamp: time.Now().UTC(),
	}
}

func (f *fakeShellRunner) Run(_ context.Context, dir string, name string, args ...string) ([]byte, []byte, error) {
	f.dir = dir
	f.name = name
	f.args = append([]string{}, args...)
	return []byte{}, []byte{}, nil
}
