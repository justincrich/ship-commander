package commander

import (
	"context"
	"errors"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
	"time"
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

	cmd, err := New(store, worktrees, locks, harness, verifier, demoTokens, events, CommanderConfig{WIPLimit: 2})
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
	if !reflect.DeepEqual(sequence, []string{"lock:m1", "dispatch:m1"}) {
		t.Fatalf("call sequence = %v, want lock before dispatch", sequence)
	}
	if len(events.events) != 1 || events.events[0].Type != EventMissionCompleted {
		t.Fatalf("events = %v, want one %s", events.events, EventMissionCompleted)
	}
	if demoTokens.CallCount() != 0 {
		t.Fatalf("demo token calls = %d, want 0 for non-standard ops mission", demoTokens.CallCount())
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

	cmd, err := New(store, worktrees, locks, harness, verifier, demoTokens, events, CommanderConfig{WIPLimit: 1})
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

	cmd, err := New(store, worktrees, locks, harness, verifier, demoTokens, events, CommanderConfig{WIPLimit: 2})
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
			"m1": "/tmp/worktree/m1",
			"m2": "/tmp/worktree/m2",
		},
	}
	locks := &fakeSurfaceLocker{}
	harness := &fakeHarness{sequence: &sequence}
	verifier := &fakeVerifier{}
	demoTokens := &fakeDemoTokenValidator{}
	events := &fakeEventPublisher{}

	cmd, err := New(store, worktrees, locks, harness, verifier, demoTokens, events, CommanderConfig{WIPLimit: 2})
	if err != nil {
		t.Fatalf("new commander: %v", err)
	}

	if err := cmd.Execute(context.Background(), "commission-1"); err != nil {
		t.Fatalf("execute: %v", err)
	}

	if len(sequence) != 2 {
		t.Fatalf("dispatch sequence = %v, want two dispatches", sequence)
	}
	if sequence[0] != "dispatch:m1" || sequence[1] != "dispatch:m2" {
		t.Fatalf("dispatch sequence = %v, want [dispatch:m1 dispatch:m2]", sequence)
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

	cmd, err := New(store, worktrees, locks, harness, verifier, demoTokens, events, CommanderConfig{WIPLimit: 1})
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

	cmd, err := New(store, worktrees, locks, harness, verifier, demoTokens, events, CommanderConfig{WIPLimit: 1})
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

	cmd, err := New(store, worktrees, locks, harness, verifier, demoTokens, events, CommanderConfig{WIPLimit: 1})
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
}

type fakeManifestStore struct {
	manifest          []Mission
	ready             [][]string
	readManifestCalls int
	readyCalls        int
	mu                sync.Mutex
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
	mu            sync.Mutex
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
	f.mu.Unlock()

	if f.delay > 0 {
		time.Sleep(f.delay)
	}

	f.mu.Lock()
	f.current--
	f.mu.Unlock()

	return DispatchResult{SessionID: "session-" + req.Mission.ID}, nil
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

func (f *fakeShellRunner) Run(_ context.Context, dir string, name string, args ...string) ([]byte, []byte, error) {
	f.dir = dir
	f.name = name
	f.args = append([]string{}, args...)
	return []byte{}, []byte{}, nil
}
