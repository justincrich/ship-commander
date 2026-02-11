package doctor

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ship-commander/sc3/internal/events"
)

const (
	defaultHeartbeatInterval = 30 * time.Second
	defaultStuckTimeout      = 5 * time.Minute
)

const (
	missionInProgress = "in_progress"
	missionBacklog    = "backlog"

	agentRunning  = "running"
	agentSpawning = "spawning"
	agentStuck    = "stuck"
)

// Mission is the subset of mission state required by Doctor monitoring.
type Mission struct {
	ID      string
	State   string
	AgentID string
}

// Agent is the subset of agent state required by Doctor monitoring.
type Agent struct {
	ID            string
	State         string
	SessionID     string
	LastHeartbeat time.Time
}

// Snapshot captures the state Doctor inspects on each heartbeat.
type Snapshot struct {
	Missions []Mission
	Agents   []Agent
}

// StateStore provides deterministic state reads/writes for Doctor checks.
type StateStore interface {
	LoadSnapshot(ctx context.Context) (Snapshot, error)
	SetMissionBacklog(ctx context.Context, missionID string) error
	SetAgentStuck(ctx context.Context, agentID string) error
}

// SessionManager resolves active tmux sessions and cleans zombie sessions.
type SessionManager interface {
	ActiveSessions(ctx context.Context) (map[string]struct{}, error)
	CleanupDeadSession(ctx context.Context, sessionID string) error
}

// EventBus publishes health and transition events.
type EventBus interface {
	Publish(event events.Event)
}

// Config controls Doctor heartbeat cadence and stuck timeout threshold.
type Config struct {
	HeartbeatInterval time.Duration
	StuckTimeout      time.Duration
}

// HealthReport is emitted on every Doctor heartbeat.
type HealthReport struct {
	ActiveAgents     int       `json:"active_agents"`
	StuckAgents      int       `json:"stuck_agents"`
	OrphanedMissions int       `json:"orphaned_missions"`
	ZombieSessions   int       `json:"zombie_sessions"`
	DoctorHeartbeat  time.Time `json:"doctor_heartbeat"`
}

// Manager executes deterministic health checks on a periodic ticker.
type Manager struct {
	store             StateStore
	sessions          SessionManager
	bus               EventBus
	heartbeatInterval time.Duration
	stuckTimeout      time.Duration
	now               func() time.Time
	newTicker         func(time.Duration) *time.Ticker
}

// NewManager builds a Doctor manager with sane defaults.
func NewManager(store StateStore, sessions SessionManager, bus EventBus, cfg Config) (*Manager, error) {
	if store == nil {
		return nil, errors.New("state store is required")
	}
	if sessions == nil {
		return nil, errors.New("session manager is required")
	}
	if bus == nil {
		return nil, errors.New("event bus is required")
	}
	if cfg.HeartbeatInterval <= 0 {
		cfg.HeartbeatInterval = defaultHeartbeatInterval
	}
	if cfg.StuckTimeout <= 0 {
		cfg.StuckTimeout = defaultStuckTimeout
	}
	return &Manager{
		store:             store,
		sessions:          sessions,
		bus:               bus,
		heartbeatInterval: cfg.HeartbeatInterval,
		stuckTimeout:      cfg.StuckTimeout,
		now:               time.Now,
		newTicker:         time.NewTicker,
	}, nil
}

// Start runs heartbeat checks until context cancellation.
func (m *Manager) Start(ctx context.Context) {
	if m == nil {
		return
	}
	ticker := m.newTicker(m.heartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if _, err := m.RunOnce(ctx); err != nil {
				m.bus.Publish(events.Event{
					Type:       events.EventTypeSystemAlert,
					Timestamp:  m.now().UTC(),
					EntityType: "health",
					EntityID:   "doctor",
					Payload: map[string]string{
						"error": err.Error(),
					},
					Severity: events.SeverityError,
				})
			}
		}
	}
}

// RunOnce executes one deterministic health check cycle.
func (m *Manager) RunOnce(ctx context.Context) (HealthReport, error) {
	if m == nil {
		return HealthReport{}, errors.New("doctor manager is nil")
	}

	snapshot, err := m.store.LoadSnapshot(ctx)
	if err != nil {
		return HealthReport{}, fmt.Errorf("load doctor snapshot: %w", err)
	}
	activeSessions, err := m.sessions.ActiveSessions(ctx)
	if err != nil {
		return HealthReport{}, fmt.Errorf("query active sessions: %w", err)
	}

	now := m.now().UTC()
	report := HealthReport{
		DoctorHeartbeat: now,
	}

	agentByID, knownSessions, activeAgents, stuckAgents, err := m.processAgents(ctx, snapshot.Agents, now)
	if err != nil {
		return HealthReport{}, err
	}
	report.ActiveAgents = activeAgents
	report.StuckAgents = stuckAgents

	orphanedMissions, err := m.repairOrphanedMissions(ctx, snapshot.Missions, agentByID, activeSessions)
	if err != nil {
		return HealthReport{}, err
	}
	report.OrphanedMissions = orphanedMissions

	zombieSessions, err := m.cleanupZombieSessions(ctx, activeSessions, knownSessions)
	if err != nil {
		return HealthReport{}, err
	}
	report.ZombieSessions = zombieSessions

	m.bus.Publish(events.Event{
		Type:       events.EventTypeHealthCheck,
		Timestamp:  now,
		EntityType: "health",
		EntityID:   "doctor",
		Payload:    report,
		Severity:   events.SeverityInfo,
	})

	return report, nil
}

func (m *Manager) processAgents(
	ctx context.Context,
	agents []Agent,
	now time.Time,
) (map[string]Agent, map[string]struct{}, int, int, error) {
	agentByID := map[string]Agent{}
	knownSessions := map[string]struct{}{}
	activeCount := 0
	stuckCount := 0

	for _, agent := range agents {
		agentByID[strings.TrimSpace(agent.ID)] = agent
		sessionID := strings.TrimSpace(agent.SessionID)
		if sessionID != "" {
			knownSessions[sessionID] = struct{}{}
		}
		if isActiveAgentState(agent.State) {
			activeCount++
		}
		if strings.EqualFold(strings.TrimSpace(agent.State), agentStuck) {
			stuckCount++
			continue
		}
		if !shouldTransitionToStuck(agent, now, m.stuckTimeout) {
			continue
		}
		if err := m.store.SetAgentStuck(ctx, agent.ID); err != nil {
			return nil, nil, 0, 0, fmt.Errorf("set agent %s stuck: %w", agent.ID, err)
		}
		stuckCount++
		m.publishStuckTransition(agent, now)
	}

	return agentByID, knownSessions, activeCount, stuckCount, nil
}

func (m *Manager) publishStuckTransition(agent Agent, now time.Time) {
	m.bus.Publish(events.Event{
		Type:       events.EventTypeStateTransition,
		Timestamp:  now,
		EntityType: "agent",
		EntityID:   agent.ID,
		Payload: map[string]string{
			"from": strings.ToLower(strings.TrimSpace(agent.State)),
			"to":   agentStuck,
		},
		Severity: events.SeverityWarn,
	})
}

func (m *Manager) repairOrphanedMissions(
	ctx context.Context,
	missions []Mission,
	agentByID map[string]Agent,
	activeSessions map[string]struct{},
) (int, error) {
	orphanedCount := 0
	for _, mission := range missions {
		if !strings.EqualFold(strings.TrimSpace(mission.State), missionInProgress) {
			continue
		}
		if !missionHasLiveSession(mission, agentByID, activeSessions) {
			if err := m.store.SetMissionBacklog(ctx, mission.ID); err != nil {
				return 0, fmt.Errorf("set orphaned mission %s backlog: %w", mission.ID, err)
			}
			orphanedCount++
		}
	}
	return orphanedCount, nil
}

func missionHasLiveSession(mission Mission, agentByID map[string]Agent, activeSessions map[string]struct{}) bool {
	agent, hasAgent := agentByID[strings.TrimSpace(mission.AgentID)]
	if !hasAgent {
		return false
	}
	sessionID := strings.TrimSpace(agent.SessionID)
	if sessionID == "" {
		return false
	}
	_, alive := activeSessions[sessionID]
	return alive
}

func (m *Manager) cleanupZombieSessions(
	ctx context.Context,
	activeSessions map[string]struct{},
	knownSessions map[string]struct{},
) (int, error) {
	cleaned := 0
	for sessionID := range activeSessions {
		if _, known := knownSessions[sessionID]; known {
			continue
		}
		if err := m.sessions.CleanupDeadSession(ctx, sessionID); err != nil {
			return 0, fmt.Errorf("cleanup zombie session %s: %w", sessionID, err)
		}
		cleaned++
	}
	return cleaned, nil
}

func shouldTransitionToStuck(agent Agent, now time.Time, timeout time.Duration) bool {
	if !isRunnableAgentState(agent.State) {
		return false
	}
	if timeout <= 0 {
		return false
	}
	if agent.LastHeartbeat.IsZero() {
		return true
	}
	return now.Sub(agent.LastHeartbeat.UTC()) > timeout
}

func isRunnableAgentState(state string) bool {
	normalized := strings.ToLower(strings.TrimSpace(state))
	return normalized == agentRunning || normalized == agentSpawning
}

func isActiveAgentState(state string) bool {
	normalized := strings.ToLower(strings.TrimSpace(state))
	return normalized == agentRunning || normalized == agentSpawning || normalized == agentStuck
}
