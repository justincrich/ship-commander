package locks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	// DefaultExpiryTimeout is the default lock lease duration when no config override is provided.
	DefaultExpiryTimeout = 5 * time.Minute
)

var (
	// ErrConflict indicates an attempted lock acquisition overlaps with an existing lock.
	ErrConflict = errors.New("surface-area lock conflict")
)

// Lock tracks one mission's surface-area reservation.
//
//nolint:revive // Field names are specified by the issue contract.
type Lock struct {
	MissionID  string    `json:"missionId"`
	Patterns   []string  `json:"patterns"`
	AcquiredAt time.Time `json:"acquiredAt"`
	ExpiresAt  time.Time `json:"expiresAt"`
}

// ManagerConfig controls lock manager behavior.
type ManagerConfig struct {
	ExpiryTimeout time.Duration
}

// Store persists lock state.
type Store interface {
	Load(ctx context.Context) ([]Lock, error)
	Save(ctx context.Context, locks []Lock) error
}

// Manager manages surface-area lock acquisition, conflict checks, and release.
type Manager struct {
	store         Store
	now           func() time.Time
	expiryTimeout time.Duration
}

// NewManager constructs a lock manager with configured lock expiry timeout.
func NewManager(store Store, cfg ManagerConfig) (*Manager, error) {
	if store == nil {
		return nil, errors.New("store is required")
	}
	if cfg.ExpiryTimeout <= 0 {
		cfg.ExpiryTimeout = DefaultExpiryTimeout
	}
	return &Manager{
		store:         store,
		now:           time.Now,
		expiryTimeout: cfg.ExpiryTimeout,
	}, nil
}

// Acquire reserves a mission's declared surface-area patterns.
func (m *Manager) Acquire(missionID string, patterns []string) error {
	if m == nil {
		return errors.New("manager is nil")
	}
	missionID = strings.TrimSpace(missionID)
	if missionID == "" {
		return errors.New("mission id must not be empty")
	}
	patterns = normalizePatterns(patterns)
	if len(patterns) == 0 {
		return errors.New("at least one lock pattern is required")
	}

	ctx := context.Background()
	locks, err := m.store.Load(ctx)
	if err != nil {
		return fmt.Errorf("load locks: %w", err)
	}

	now := m.now().UTC()
	locks = onlyActiveLocks(locks, now)
	locks = withoutMission(locks, missionID)

	conflicts := findConflicts(locks, patterns)
	if len(conflicts) > 0 {
		return fmt.Errorf("%w: mission=%s conflicts=%d", ErrConflict, missionID, len(conflicts))
	}

	locks = append(locks, Lock{
		MissionID:  missionID,
		Patterns:   append([]string(nil), patterns...),
		AcquiredAt: now,
		ExpiresAt:  now.Add(m.expiryTimeout),
	})

	if err := m.store.Save(ctx, locks); err != nil {
		return fmt.Errorf("save locks: %w", err)
	}
	return nil
}

// Release removes a mission lock.
func (m *Manager) Release(missionID string) error {
	if m == nil {
		return errors.New("manager is nil")
	}
	missionID = strings.TrimSpace(missionID)
	if missionID == "" {
		return errors.New("mission id must not be empty")
	}

	ctx := context.Background()
	locks, err := m.store.Load(ctx)
	if err != nil {
		return fmt.Errorf("load locks: %w", err)
	}
	locks = withoutMission(onlyActiveLocks(locks, m.now().UTC()), missionID)
	if err := m.store.Save(ctx, locks); err != nil {
		return fmt.Errorf("save locks: %w", err)
	}
	return nil
}

// CheckConflict returns existing locks overlapping requested patterns.
func (m *Manager) CheckConflict(patterns []string) ([]Lock, error) {
	if m == nil {
		return nil, errors.New("manager is nil")
	}
	patterns = normalizePatterns(patterns)
	if len(patterns) == 0 {
		return nil, errors.New("at least one lock pattern is required")
	}

	locks, err := m.store.Load(context.Background())
	if err != nil {
		return nil, fmt.Errorf("load locks: %w", err)
	}
	locks = onlyActiveLocks(locks, m.now().UTC())
	return findConflicts(locks, patterns), nil
}

func findConflicts(existing []Lock, requested []string) []Lock {
	conflicts := make([]Lock, 0)
	for _, lock := range existing {
		if lockOverlaps(lock, requested) {
			conflicts = append(conflicts, lock)
		}
	}
	return conflicts
}

func lockOverlaps(lock Lock, requested []string) bool {
	for _, existingPattern := range lock.Patterns {
		for _, requestedPattern := range requested {
			if patternsOverlap(existingPattern, requestedPattern) {
				return true
			}
		}
	}
	return false
}

func patternsOverlap(a, b string) bool {
	a = filepath.ToSlash(strings.TrimSpace(a))
	b = filepath.ToSlash(strings.TrimSpace(b))
	if a == "" || b == "" {
		return false
	}
	if a == b {
		return true
	}
	if prefix, ok := doubleStarPrefix(a); ok && hasPathPrefix(b, prefix) {
		return true
	}
	if prefix, ok := doubleStarPrefix(b); ok && hasPathPrefix(a, prefix) {
		return true
	}
	if matched, err := filepath.Match(a, b); err == nil && matched {
		return true
	}
	if matched, err := filepath.Match(b, a); err == nil && matched {
		return true
	}
	return false
}

func hasPathPrefix(value, prefix string) bool {
	value = filepath.ToSlash(strings.TrimSpace(value))
	prefix = filepath.ToSlash(strings.TrimSpace(prefix))
	if value == prefix {
		return true
	}
	return strings.HasPrefix(value, prefix+"/")
}

func doubleStarPrefix(pattern string) (string, bool) {
	pattern = filepath.ToSlash(strings.TrimSpace(pattern))
	if strings.HasSuffix(pattern, "/**") {
		return strings.TrimSuffix(pattern, "/**"), true
	}
	return "", false
}

func normalizePatterns(patterns []string) []string {
	normalized := make([]string, 0, len(patterns))
	for _, pattern := range patterns {
		pattern = filepath.ToSlash(strings.TrimSpace(pattern))
		if pattern == "" {
			continue
		}
		normalized = append(normalized, pattern)
	}
	return normalized
}

func onlyActiveLocks(locks []Lock, now time.Time) []Lock {
	active := make([]Lock, 0, len(locks))
	for _, lock := range locks {
		if lock.ExpiresAt.IsZero() || lock.ExpiresAt.After(now) {
			active = append(active, lock)
		}
	}
	return active
}

func withoutMission(locks []Lock, missionID string) []Lock {
	filtered := make([]Lock, 0, len(locks))
	for _, lock := range locks {
		if strings.TrimSpace(lock.MissionID) == missionID {
			continue
		}
		filtered = append(filtered, lock)
	}
	return filtered
}

// CommanderSurfaceLocker adapts Manager to the commander locker interface.
type CommanderSurfaceLocker struct {
	manager *Manager
}

// NewCommanderSurfaceLocker constructs a commander-compatible surface locker.
func NewCommanderSurfaceLocker(manager *Manager) (*CommanderSurfaceLocker, error) {
	if manager == nil {
		return nil, errors.New("manager is required")
	}
	return &CommanderSurfaceLocker{manager: manager}, nil
}

// Acquire reserves the surface area and returns a release closure.
func (l *CommanderSurfaceLocker) Acquire(_ context.Context, missionID string, patterns []string) (func() error, error) {
	if l == nil || l.manager == nil {
		return nil, errors.New("surface locker is not initialized")
	}
	if err := l.manager.Acquire(missionID, patterns); err != nil {
		return nil, err
	}
	return func() error {
		return l.manager.Release(missionID)
	}, nil
}

// CommandRunner executes Beads CLI commands for lock persistence.
type CommandRunner interface {
	Run(ctx context.Context, name string, args ...string) ([]byte, error)
}

type defaultCommandRunner struct{}

func (defaultCommandRunner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("run %s %s: %w (%s)", name, strings.Join(args, " "), err, strings.TrimSpace(string(output)))
	}
	return output, nil
}

// BeadsStore persists lock state in one Beads issue's notes field.
type BeadsStore struct {
	issueID string
	runner  CommandRunner
}

// NewBeadsStore constructs a Beads-backed lock store.
func NewBeadsStore(issueID string) (*BeadsStore, error) {
	return NewBeadsStoreWithRunner(issueID, defaultCommandRunner{})
}

// NewBeadsStoreWithRunner constructs a Beads-backed lock store with a custom runner.
func NewBeadsStoreWithRunner(issueID string, runner CommandRunner) (*BeadsStore, error) {
	issueID = strings.TrimSpace(issueID)
	if issueID == "" {
		return nil, errors.New("issue id must not be empty")
	}
	if runner == nil {
		return nil, errors.New("runner must not be nil")
	}
	return &BeadsStore{issueID: issueID, runner: runner}, nil
}

// Load reads locks from Beads notes.
func (s *BeadsStore) Load(ctx context.Context) ([]Lock, error) {
	if s == nil {
		return nil, errors.New("beads store is nil")
	}
	out, err := s.runner.Run(ctx, "bd", "show", s.issueID, "--json")
	if err != nil {
		return nil, fmt.Errorf("show issue %s: %w", s.issueID, err)
	}
	var records []struct {
		Notes string `json:"notes"`
	}
	if err := json.Unmarshal(out, &records); err != nil {
		return nil, fmt.Errorf("parse show output: %w", err)
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("issue %s not found", s.issueID)
	}
	notes := strings.TrimSpace(records[0].Notes)
	if notes == "" {
		return []Lock{}, nil
	}
	var locks []Lock
	if err := json.Unmarshal([]byte(notes), &locks); err != nil {
		return nil, fmt.Errorf("parse lock notes payload: %w", err)
	}
	return locks, nil
}

// Save writes locks to Beads notes.
func (s *BeadsStore) Save(ctx context.Context, locks []Lock) error {
	if s == nil {
		return errors.New("beads store is nil")
	}
	payload, err := json.Marshal(locks)
	if err != nil {
		return fmt.Errorf("marshal locks: %w", err)
	}
	_, err = s.runner.Run(ctx, "bd", "update", s.issueID, "--notes", string(payload))
	if err != nil {
		return fmt.Errorf("update issue %s notes: %w", s.issueID, err)
	}
	return nil
}
