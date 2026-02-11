package doctor

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/ship-commander/sc3/internal/events"
)

func TestNewManagerValidatesInputsAndDefaults(t *testing.T) {
	store := &fakeStateStore{}
	sessions := &fakeSessionManager{}
	bus := &fakeEventBus{}

	if _, err := NewManager(nil, sessions, bus, Config{}); err == nil {
		t.Fatal("expected error for nil store")
	}
	if _, err := NewManager(store, nil, bus, Config{}); err == nil {
		t.Fatal("expected error for nil session manager")
	}
	if _, err := NewManager(store, sessions, nil, Config{}); err == nil {
		t.Fatal("expected error for nil event bus")
	}

	manager, err := NewManager(store, sessions, bus, Config{})
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}
	if manager.heartbeatInterval != defaultHeartbeatInterval {
		t.Fatalf("heartbeatInterval = %s, want %s", manager.heartbeatInterval, defaultHeartbeatInterval)
	}
	if manager.stuckTimeout != defaultStuckTimeout {
		t.Fatalf("stuckTimeout = %s, want %s", manager.stuckTimeout, defaultStuckTimeout)
	}
}

func TestRunOnceDetectsStuckAgentsOrphansAndZombies(t *testing.T) {
	now := time.Date(2026, 2, 11, 8, 30, 0, 0, time.UTC)
	store := &fakeStateStore{
		snapshot: Snapshot{
			Agents: []Agent{
				{ID: "agent-live", State: agentRunning, SessionID: "session-live", LastHeartbeat: now.Add(-1 * time.Minute)},
				{ID: "agent-stale", State: agentRunning, SessionID: "session-missing", LastHeartbeat: now.Add(-10 * time.Minute)},
				{ID: "agent-already-stuck", State: agentStuck, SessionID: "session-stuck", LastHeartbeat: now.Add(-12 * time.Minute)},
			},
			Missions: []Mission{
				{ID: "mission-live", State: missionInProgress, AgentID: "agent-live"},
				{ID: "mission-orphan-no-agent", State: missionInProgress, AgentID: "agent-does-not-exist"},
				{ID: "mission-orphan-missing-session", State: missionInProgress, AgentID: "agent-stale"},
			},
		},
	}
	sessions := &fakeSessionManager{
		activeSessions: map[string]struct{}{
			"session-live":   {},
			"session-zombie": {},
		},
	}
	bus := &fakeEventBus{}

	manager, err := NewManager(store, sessions, bus, Config{
		HeartbeatInterval: 50 * time.Millisecond,
		StuckTimeout:      5 * time.Minute,
	})
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}
	manager.now = func() time.Time { return now }

	report, err := manager.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("run once: %v", err)
	}

	if report.ActiveAgents != 3 {
		t.Fatalf("ActiveAgents = %d, want 3", report.ActiveAgents)
	}
	if report.StuckAgents != 2 {
		t.Fatalf("StuckAgents = %d, want 2", report.StuckAgents)
	}
	if report.OrphanedMissions != 2 {
		t.Fatalf("OrphanedMissions = %d, want 2", report.OrphanedMissions)
	}
	if report.ZombieSessions != 1 {
		t.Fatalf("ZombieSessions = %d, want 1", report.ZombieSessions)
	}

	if !reflect.DeepEqual(store.setAgentStuck, []string{"agent-stale"}) {
		t.Fatalf("setAgentStuck = %v, want [agent-stale]", store.setAgentStuck)
	}
	if !reflect.DeepEqual(store.setMissionBacklog, []string{"mission-orphan-no-agent", "mission-orphan-missing-session"}) {
		t.Fatalf("setMissionBacklog = %v", store.setMissionBacklog)
	}
	if !reflect.DeepEqual(sessions.cleanedSessions, []string{"session-zombie"}) {
		t.Fatalf("cleanedSessions = %v, want [session-zombie]", sessions.cleanedSessions)
	}

	if count := bus.countByType(events.EventTypeStateTransition); count != 1 {
		t.Fatalf("state transition events = %d, want 1", count)
	}
	if count := bus.countByType(events.EventTypeHealthCheck); count != 1 {
		t.Fatalf("health check events = %d, want 1", count)
	}
}

func TestStartRunsUntilCancelled(t *testing.T) {
	store := &fakeStateStore{
		snapshot: Snapshot{
			Agents:   []Agent{},
			Missions: []Mission{},
		},
	}
	sessions := &fakeSessionManager{activeSessions: map[string]struct{}{}}
	bus := &fakeEventBus{}

	manager, err := NewManager(store, sessions, bus, Config{
		HeartbeatInterval: 20 * time.Millisecond,
		StuckTimeout:      5 * time.Minute,
	})
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		manager.Start(ctx)
	}()

	time.Sleep(75 * time.Millisecond)
	cancel()

	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("doctor start did not stop on context cancellation")
	}
	if count := bus.countByType(events.EventTypeHealthCheck); count < 2 {
		t.Fatalf("health check event count = %d, want at least 2", count)
	}
}

func TestRunOncePropagatesStoreErrors(t *testing.T) {
	store := &fakeStateStore{
		loadSnapshotErr: errors.New("snapshot unavailable"),
	}
	sessions := &fakeSessionManager{activeSessions: map[string]struct{}{}}
	bus := &fakeEventBus{}

	manager, err := NewManager(store, sessions, bus, Config{})
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}
	if _, err := manager.RunOnce(context.Background()); err == nil {
		t.Fatal("expected run once error when snapshot load fails")
	}
}

type fakeStateStore struct {
	snapshot          Snapshot
	loadSnapshotErr   error
	setMissionBacklog []string
	setAgentStuck     []string
}

func (f *fakeStateStore) LoadSnapshot(context.Context) (Snapshot, error) {
	if f.loadSnapshotErr != nil {
		return Snapshot{}, f.loadSnapshotErr
	}
	return f.snapshot, nil
}

func (f *fakeStateStore) SetMissionBacklog(_ context.Context, missionID string) error {
	f.setMissionBacklog = append(f.setMissionBacklog, missionID)
	return nil
}

func (f *fakeStateStore) SetAgentStuck(_ context.Context, agentID string) error {
	f.setAgentStuck = append(f.setAgentStuck, agentID)
	return nil
}

type fakeSessionManager struct {
	activeSessions    map[string]struct{}
	activeSessionsErr error
	cleanedSessions   []string
}

func (f *fakeSessionManager) ActiveSessions(context.Context) (map[string]struct{}, error) {
	if f.activeSessionsErr != nil {
		return nil, f.activeSessionsErr
	}
	copyMap := make(map[string]struct{}, len(f.activeSessions))
	for key := range f.activeSessions {
		copyMap[key] = struct{}{}
	}
	return copyMap, nil
}

func (f *fakeSessionManager) CleanupDeadSession(_ context.Context, sessionID string) error {
	f.cleanedSessions = append(f.cleanedSessions, sessionID)
	return nil
}

type fakeEventBus struct {
	mu     sync.Mutex
	events []events.Event
}

func (f *fakeEventBus) Publish(event events.Event) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.events = append(f.events, event)
}

func (f *fakeEventBus) countByType(eventType string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	count := 0
	for _, event := range f.events {
		if event.Type == eventType {
			count++
		}
	}
	return count
}
