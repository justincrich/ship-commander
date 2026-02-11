package locks

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"
)

type memoryStore struct {
	mu    sync.Mutex
	locks []Lock
}

func (m *memoryStore) Load(_ context.Context) ([]Lock, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]Lock, len(m.locks))
	copy(out, m.locks)
	return out, nil
}

func (m *memoryStore) Save(_ context.Context, locks []Lock) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.locks = make([]Lock, len(locks))
	copy(m.locks, locks)
	return nil
}

type fakeBeadsRunner struct {
	mu      sync.Mutex
	issueID string
	notes   string
}

func (r *fakeBeadsRunner) Run(_ context.Context, _ string, args ...string) ([]byte, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(args) == 0 {
		return nil, errors.New("missing command args")
	}
	switch args[0] {
	case "show":
		payload, err := json.Marshal([]map[string]string{{
			"id":    r.issueID,
			"notes": r.notes,
		}})
		if err != nil {
			return nil, err
		}
		return payload, nil
	case "update":
		if len(args) >= 4 && args[2] == "--notes" {
			r.notes = args[3]
		}
		return []byte(`{"ok":true}`), nil
	default:
		return nil, errors.New("unexpected command")
	}
}

func TestAcquireConflictReleaseFlow(t *testing.T) {
	t.Parallel()

	store := &memoryStore{}
	mgr, err := NewManager(store, ManagerConfig{ExpiryTimeout: 10 * time.Minute})
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}

	if err := mgr.Acquire("mission-1", []string{"internal/commander/**"}); err != nil {
		t.Fatalf("acquire lock: %v", err)
	}

	conflicts, err := mgr.CheckConflict([]string{"internal/commander/commander.go"})
	if err != nil {
		t.Fatalf("check conflict: %v", err)
	}
	if len(conflicts) != 1 {
		t.Fatalf("conflicts = %d, want 1", len(conflicts))
	}

	err = mgr.Acquire("mission-2", []string{"internal/commander/commander.go"})
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("acquire conflict err = %v, want ErrConflict", err)
	}

	if err := mgr.Release("mission-1"); err != nil {
		t.Fatalf("release lock: %v", err)
	}

	conflicts, err = mgr.CheckConflict([]string{"internal/commander/commander.go"})
	if err != nil {
		t.Fatalf("check conflict after release: %v", err)
	}
	if len(conflicts) != 0 {
		t.Fatalf("conflicts after release = %d, want 0", len(conflicts))
	}
}

func TestLockExpiryAndConfigurableTimeout(t *testing.T) {
	t.Parallel()

	store := &memoryStore{}
	mgr, err := NewManager(store, ManagerConfig{ExpiryTimeout: time.Second})
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}

	t0 := time.Date(2026, 2, 11, 0, 0, 0, 0, time.UTC)
	mgr.now = func() time.Time { return t0 }

	if err := mgr.Acquire("mission-1", []string{"internal/gates/**"}); err != nil {
		t.Fatalf("acquire lock: %v", err)
	}

	mgr.now = func() time.Time { return t0.Add(2 * time.Second) }
	conflicts, err := mgr.CheckConflict([]string{"internal/gates/gate.go"})
	if err != nil {
		t.Fatalf("check conflicts: %v", err)
	}
	if len(conflicts) != 0 {
		t.Fatalf("expired lock should not conflict, got %d conflicts", len(conflicts))
	}
}

func TestLocksPersistAcrossManagerRestart(t *testing.T) {
	t.Parallel()

	store := &memoryStore{}
	mgr1, err := NewManager(store, ManagerConfig{ExpiryTimeout: 5 * time.Minute})
	if err != nil {
		t.Fatalf("new manager1: %v", err)
	}
	if err := mgr1.Acquire("mission-1", []string{"internal/readyroom/**"}); err != nil {
		t.Fatalf("acquire lock: %v", err)
	}

	mgr2, err := NewManager(store, ManagerConfig{ExpiryTimeout: 5 * time.Minute})
	if err != nil {
		t.Fatalf("new manager2: %v", err)
	}
	conflicts, err := mgr2.CheckConflict([]string{"internal/readyroom/readyroom.go"})
	if err != nil {
		t.Fatalf("check conflict: %v", err)
	}
	if len(conflicts) != 1 {
		t.Fatalf("conflicts = %d, want 1", len(conflicts))
	}
}

func TestBeadsStoreRoundTripPersistsLocks(t *testing.T) {
	t.Parallel()

	runner := &fakeBeadsRunner{issueID: "ship-commander-3-9sd"}
	store, err := NewBeadsStoreWithRunner("ship-commander-3-9sd", runner)
	if err != nil {
		t.Fatalf("new beads store: %v", err)
	}

	locks := []Lock{{
		MissionID:  "mission-1",
		Patterns:   []string{"internal/state/**"},
		AcquiredAt: time.Date(2026, 2, 11, 0, 0, 0, 0, time.UTC),
		ExpiresAt:  time.Date(2026, 2, 11, 0, 5, 0, 0, time.UTC),
	}}
	if err := store.Save(context.Background(), locks); err != nil {
		t.Fatalf("save locks: %v", err)
	}

	loaded, err := store.Load(context.Background())
	if err != nil {
		t.Fatalf("load locks: %v", err)
	}
	if len(loaded) != 1 {
		t.Fatalf("loaded locks = %d, want 1", len(loaded))
	}
	if loaded[0].MissionID != "mission-1" {
		t.Fatalf("mission id = %q, want mission-1", loaded[0].MissionID)
	}
}

func TestCommanderSurfaceLockerAcquireRelease(t *testing.T) {
	t.Parallel()

	store := &memoryStore{}
	mgr, err := NewManager(store, ManagerConfig{ExpiryTimeout: 5 * time.Minute})
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}
	locker, err := NewCommanderSurfaceLocker(mgr)
	if err != nil {
		t.Fatalf("new commander surface locker: %v", err)
	}

	release, err := locker.Acquire(context.Background(), "mission-1", []string{"internal/protocol/**"})
	if err != nil {
		t.Fatalf("acquire: %v", err)
	}
	if release == nil {
		t.Fatal("release closure should not be nil")
	}

	conflicts, err := mgr.CheckConflict([]string{"internal/protocol/protocol.go"})
	if err != nil {
		t.Fatalf("check conflicts: %v", err)
	}
	if len(conflicts) != 1 {
		t.Fatalf("conflicts = %d, want 1", len(conflicts))
	}

	if err := release(); err != nil {
		t.Fatalf("release: %v", err)
	}
	conflicts, err = mgr.CheckConflict([]string{"internal/protocol/protocol.go"})
	if err != nil {
		t.Fatalf("check conflicts after release: %v", err)
	}
	if len(conflicts) != 0 {
		t.Fatalf("conflicts after release = %d, want 0", len(conflicts))
	}
}
