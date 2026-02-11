package recovery

import (
	"context"
	"errors"
	"reflect"
	"sort"
	"testing"
	"time"
)

type fakeStateStore struct {
	snapshot           Snapshot
	setMissionBacklog  []string
	setAgentDead       []string
	loadSnapshotErr    error
	setMissionStateErr error
	setAgentStateErr   error
}

func (f *fakeStateStore) LoadSnapshot(_ context.Context) (Snapshot, error) {
	if f.loadSnapshotErr != nil {
		return Snapshot{}, f.loadSnapshotErr
	}
	return f.snapshot, nil
}

func (f *fakeStateStore) SetMissionBacklog(_ context.Context, missionID string) error {
	if f.setMissionStateErr != nil {
		return f.setMissionStateErr
	}
	f.setMissionBacklog = append(f.setMissionBacklog, missionID)
	return nil
}

func (f *fakeStateStore) SetAgentDead(_ context.Context, agentID string) error {
	if f.setAgentStateErr != nil {
		return f.setAgentStateErr
	}
	f.setAgentDead = append(f.setAgentDead, agentID)
	return nil
}

type fakeSessionManager struct {
	activeSessions map[string]struct{}
	cleaned        []string
	activeErr      error
	cleanupErr     error
}

func (f *fakeSessionManager) ActiveSessions(_ context.Context) (map[string]struct{}, error) {
	if f.activeErr != nil {
		return nil, f.activeErr
	}
	out := map[string]struct{}{}
	for sessionID := range f.activeSessions {
		out[sessionID] = struct{}{}
	}
	return out, nil
}

func (f *fakeSessionManager) CleanupDeadSession(_ context.Context, sessionID string) error {
	if f.cleanupErr != nil {
		return f.cleanupErr
	}
	f.cleaned = append(f.cleaned, sessionID)
	return nil
}

func TestRecoverRepairsOrphansCleansDeadSessionsAndReturnsResumeCommissions(t *testing.T) {
	t.Parallel()

	store := &fakeStateStore{
		snapshot: Snapshot{
			Commissions: []Commission{
				{ID: "comm-1", State: CommissionExecuting},
				{ID: "comm-2", State: "completed"},
			},
			Missions: []Mission{
				{ID: "mission-live", CommissionID: "comm-1", State: MissionInProgress, AgentID: "agent-live"},
				{ID: "mission-orphan", CommissionID: "comm-1", State: MissionInProgress, AgentID: "agent-missing"},
				{ID: "mission-no-agent", CommissionID: "comm-1", State: MissionInProgress},
				{ID: "mission-done", CommissionID: "comm-2", State: MissionDone, AgentID: "agent-done"},
			},
			Agents: []Agent{
				{ID: "agent-live", State: AgentRunning, SessionID: "session-live"},
				{ID: "agent-missing", State: AgentRunning, SessionID: "session-missing"},
				{ID: "agent-stale", State: AgentSpawning, SessionID: "session-stale"},
				{ID: "agent-done", State: AgentDead, SessionID: "session-done"},
			},
		},
	}
	sessions := &fakeSessionManager{activeSessions: map[string]struct{}{"session-live": {}}}

	manager, err := NewManager(store, sessions, Config{ResumeTimeout: 10 * time.Second})
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}

	start := time.Date(2026, 2, 11, 0, 0, 0, 0, time.UTC)
	ticks := 0
	manager.now = func() time.Time {
		defer func() { ticks++ }()
		return start.Add(time.Duration(ticks) * time.Second)
	}

	result, err := manager.Recover(context.Background())
	if err != nil {
		t.Fatalf("recover: %v", err)
	}

	sort.Strings(result.OrphanedMissionIDs)
	if !reflect.DeepEqual(result.OrphanedMissionIDs, []string{"mission-no-agent", "mission-orphan"}) {
		t.Fatalf("orphaned missions = %v", result.OrphanedMissionIDs)
	}

	sort.Strings(result.CleanedDeadSessions)
	if !reflect.DeepEqual(result.CleanedDeadSessions, []string{"session-missing", "session-stale"}) {
		t.Fatalf("cleaned sessions = %v", result.CleanedDeadSessions)
	}

	if !reflect.DeepEqual(result.ResumeCommissionIDs, []string{"comm-1"}) {
		t.Fatalf("resume commissions = %v, want [comm-1]", result.ResumeCommissionIDs)
	}

	sort.Strings(store.setMissionBacklog)
	if !reflect.DeepEqual(store.setMissionBacklog, []string{"mission-no-agent", "mission-orphan"}) {
		t.Fatalf("set mission backlog calls = %v", store.setMissionBacklog)
	}

	sort.Strings(store.setAgentDead)
	if !reflect.DeepEqual(store.setAgentDead, []string{"agent-missing", "agent-stale"}) {
		t.Fatalf("set agent dead calls = %v", store.setAgentDead)
	}

	if result.RecoveryDuration != time.Second {
		t.Fatalf("recovery duration = %s, want 1s", result.RecoveryDuration)
	}
}

func TestRecoverReturnsErrorWhenRecoveryExceedsTimeout(t *testing.T) {
	t.Parallel()

	store := &fakeStateStore{snapshot: Snapshot{}}
	sessions := &fakeSessionManager{activeSessions: map[string]struct{}{}}
	manager, err := NewManager(store, sessions, Config{ResumeTimeout: 500 * time.Millisecond})
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}

	start := time.Date(2026, 2, 11, 0, 0, 0, 0, time.UTC)
	ticks := 0
	manager.now = func() time.Time {
		defer func() { ticks++ }()
		if ticks == 0 {
			return start
		}
		return start.Add(2 * time.Second)
	}

	_, err = manager.Recover(context.Background())
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestRecoverWrapsStoreAndSessionErrors(t *testing.T) {
	t.Parallel()

	t.Run("snapshot error", func(t *testing.T) {
		manager, err := NewManager(
			&fakeStateStore{loadSnapshotErr: errors.New("load failed")},
			&fakeSessionManager{},
			Config{},
		)
		if err != nil {
			t.Fatalf("new manager: %v", err)
		}
		_, runErr := manager.Recover(context.Background())
		if runErr == nil {
			t.Fatal("expected snapshot error")
		}
	})

	t.Run("active sessions error", func(t *testing.T) {
		manager, err := NewManager(
			&fakeStateStore{snapshot: Snapshot{}},
			&fakeSessionManager{activeErr: errors.New("tmux failed")},
			Config{},
		)
		if err != nil {
			t.Fatalf("new manager: %v", err)
		}
		_, runErr := manager.Recover(context.Background())
		if runErr == nil {
			t.Fatal("expected active sessions error")
		}
	})
}
