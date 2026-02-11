package recovery

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ship-commander/sc3/internal/events"
)

const (
	// DefaultResumeTimeout is the deterministic upper bound for startup recovery.
	DefaultResumeTimeout = 10 * time.Second
)

const (
	// CommissionExecuting marks a commission that should resume execution after recovery.
	CommissionExecuting = "executing"
	// MissionInProgress marks a mission currently being executed.
	MissionInProgress = "in_progress"
	// MissionBacklog marks a mission queued and ready for dispatch.
	MissionBacklog = "backlog"
	// MissionDone marks a mission that no longer needs execution.
	MissionDone = "done"
	// MissionHalted marks a mission that was deterministically halted.
	MissionHalted = "halted"
	// AgentRunning marks an active agent session.
	AgentRunning = "running"
	// AgentSpawning marks an agent session that is still starting.
	AgentSpawning = "spawning"
	// AgentDead marks a dead agent session.
	AgentDead = "dead"
)

// Commission captures recovery-relevant persisted commission state.
type Commission struct {
	ID    string
	State string
}

// Mission captures recovery-relevant persisted mission state.
type Mission struct {
	ID           string
	CommissionID string
	State        string
	AgentID      string
}

// Agent captures recovery-relevant persisted agent state.
type Agent struct {
	ID        string
	State     string
	SessionID string
}

// Snapshot is the full persisted startup state used by recovery.
type Snapshot struct {
	Commissions []Commission
	Missions    []Mission
	Agents      []Agent
}

// Result captures deterministic startup recovery outputs.
type Result struct {
	Snapshot            Snapshot
	OrphanedMissionIDs  []string
	CleanedDeadSessions []string
	ResumeCommissionIDs []string
	RecoveryDuration    time.Duration
}

// StateStore provides Beads-backed snapshot reads and deterministic state updates.
type StateStore interface {
	LoadSnapshot(ctx context.Context) (Snapshot, error)
	SetMissionBacklog(ctx context.Context, missionID string) error
	SetAgentDead(ctx context.Context, agentID string) error
}

// SessionManager queries and cleans up tmux-backed sessions.
type SessionManager interface {
	ActiveSessions(ctx context.Context) (map[string]struct{}, error)
	CleanupDeadSession(ctx context.Context, sessionID string) error
}

// EventBus publishes recovery audit events.
type EventBus interface {
	Publish(event events.Event)
}

// Config configures startup recovery behavior.
type Config struct {
	ResumeTimeout time.Duration
	EventBus      EventBus
}

// Manager reconstructs persisted state and repairs orphaned execution state.
type Manager struct {
	store         StateStore
	sessions      SessionManager
	bus           EventBus
	resumeTimeout time.Duration
	now           func() time.Time
}

// NewManager constructs a startup recovery manager.
func NewManager(store StateStore, sessions SessionManager, cfg Config) (*Manager, error) {
	if store == nil {
		return nil, errors.New("state store is required")
	}
	if sessions == nil {
		return nil, errors.New("session manager is required")
	}
	if cfg.ResumeTimeout <= 0 {
		cfg.ResumeTimeout = DefaultResumeTimeout
	}
	return &Manager{
		store:         store,
		sessions:      sessions,
		bus:           cfg.EventBus,
		resumeTimeout: cfg.ResumeTimeout,
		now:           time.Now,
	}, nil
}

// Recover reads persisted state, resets orphaned missions, cleans dead sessions, and returns commissions to resume.
func (m *Manager) Recover(ctx context.Context) (Result, error) {
	if m == nil {
		return Result{}, errors.New("recovery manager is nil")
	}
	started := m.now()
	auditTimestamp := started.UTC()

	snapshot, err := m.store.LoadSnapshot(ctx)
	if err != nil {
		return Result{}, fmt.Errorf("load recovery snapshot: %w", err)
	}
	activeSessions, err := m.sessions.ActiveSessions(ctx)
	if err != nil {
		return Result{}, fmt.Errorf("query active tmux sessions: %w", err)
	}

	agentByID := buildAgentIndex(snapshot.Agents)

	result := Result{Snapshot: snapshot}
	markedDeadAgents := map[string]struct{}{}
	orphanedMissions, err := m.recoverOrphanedMissions(
		ctx,
		snapshot.Missions,
		agentByID,
		activeSessions,
		markedDeadAgents,
		auditTimestamp,
	)
	if err != nil {
		return Result{}, err
	}
	result.OrphanedMissionIDs = orphanedMissions

	cleanedSessions, err := m.cleanupDeadSessions(
		ctx,
		snapshot.Agents,
		activeSessions,
		markedDeadAgents,
		auditTimestamp,
	)
	if err != nil {
		return Result{}, err
	}
	result.CleanedDeadSessions = cleanedSessions

	result.ResumeCommissionIDs = commissionsToResume(snapshot.Commissions)
	result.RecoveryDuration = m.now().Sub(started)
	if err := validateRecoveryDuration(result.RecoveryDuration, m.resumeTimeout); err != nil {
		return Result{}, err
	}
	m.publishRecoverySummary(result, auditTimestamp)

	return result, nil
}

func buildAgentIndex(agents []Agent) map[string]Agent {
	index := map[string]Agent{}
	for _, agent := range agents {
		if strings.TrimSpace(agent.ID) == "" {
			continue
		}
		index[agent.ID] = agent
	}
	return index
}

func (m *Manager) recoverOrphanedMissions(
	ctx context.Context,
	missions []Mission,
	agentByID map[string]Agent,
	activeSessions map[string]struct{},
	markedDeadAgents map[string]struct{},
	auditTimestamp time.Time,
) ([]string, error) {
	orphanedMissionIDs := make([]string, 0)
	for _, mission := range missions {
		if !isInProgressMission(mission) {
			continue
		}

		agent, hasAgent := agentByID[strings.TrimSpace(mission.AgentID)]
		if hasLiveAgentSession(agent, activeSessions) {
			continue
		}
		if err := m.store.SetMissionBacklog(ctx, mission.ID); err != nil {
			return nil, fmt.Errorf("set orphaned mission %s to backlog: %w", mission.ID, err)
		}
		orphanedMissionIDs = append(orphanedMissionIDs, mission.ID)
		m.publishAuditEvent(events.Event{
			Type:       events.EventTypeStateTransition,
			Timestamp:  auditTimestamp,
			EntityType: "mission",
			EntityID:   mission.ID,
			Payload: map[string]string{
				"from": MissionInProgress,
				"to":   MissionBacklog,
			},
			Severity: events.SeverityWarn,
		})

		if hasAgent && isActiveAgentState(agent.State) {
			if err := m.store.SetAgentDead(ctx, agent.ID); err != nil {
				return nil, fmt.Errorf("mark orphaned agent %s dead: %w", agent.ID, err)
			}
			markedDeadAgents[agent.ID] = struct{}{}
			m.publishAuditEvent(events.Event{
				Type:       events.EventTypeStateTransition,
				Timestamp:  auditTimestamp,
				EntityType: "agent",
				EntityID:   agent.ID,
				Payload: map[string]string{
					"from": strings.ToLower(strings.TrimSpace(agent.State)),
					"to":   AgentDead,
				},
				Severity: events.SeverityWarn,
			})
		}
	}
	return orphanedMissionIDs, nil
}

func (m *Manager) cleanupDeadSessions(
	ctx context.Context,
	agents []Agent,
	activeSessions map[string]struct{},
	markedDeadAgents map[string]struct{},
	auditTimestamp time.Time,
) ([]string, error) {
	cleanedSessionIDs := make([]string, 0)
	alreadyCleaned := map[string]struct{}{}

	for _, agent := range agents {
		if !isActiveAgentState(agent.State) {
			continue
		}
		sessionID := strings.TrimSpace(agent.SessionID)
		if sessionID == "" {
			continue
		}
		if _, ok := activeSessions[sessionID]; ok {
			continue
		}
		if _, exists := alreadyCleaned[sessionID]; exists {
			continue
		}

		if err := m.sessions.CleanupDeadSession(ctx, sessionID); err != nil {
			return nil, fmt.Errorf("cleanup dead session %s: %w", sessionID, err)
		}
		if _, alreadyMarked := markedDeadAgents[agent.ID]; !alreadyMarked {
			if err := m.store.SetAgentDead(ctx, agent.ID); err != nil {
				return nil, fmt.Errorf("mark dead agent %s: %w", agent.ID, err)
			}
			markedDeadAgents[agent.ID] = struct{}{}
			m.publishAuditEvent(events.Event{
				Type:       events.EventTypeStateTransition,
				Timestamp:  auditTimestamp,
				EntityType: "agent",
				EntityID:   agent.ID,
				Payload: map[string]string{
					"from": strings.ToLower(strings.TrimSpace(agent.State)),
					"to":   AgentDead,
				},
				Severity: events.SeverityWarn,
			})
		}
		cleanedSessionIDs = append(cleanedSessionIDs, sessionID)
		alreadyCleaned[sessionID] = struct{}{}
		m.publishAuditEvent(events.Event{
			Type:       events.EventTypeHealthCheck,
			Timestamp:  auditTimestamp,
			EntityType: "session",
			EntityID:   sessionID,
			Payload: map[string]string{
				"action": "cleanup_dead_session",
			},
			Severity: events.SeverityInfo,
		})
	}

	return cleanedSessionIDs, nil
}

func commissionsToResume(commissions []Commission) []string {
	resumeCommissionIDs := make([]string, 0)
	for _, commission := range commissions {
		if strings.EqualFold(strings.TrimSpace(commission.State), CommissionExecuting) {
			resumeCommissionIDs = append(resumeCommissionIDs, commission.ID)
		}
	}
	return resumeCommissionIDs
}

func validateRecoveryDuration(duration, timeout time.Duration) error {
	if duration <= timeout {
		return nil
	}
	return fmt.Errorf(
		"startup recovery exceeded timeout: duration=%s timeout=%s",
		duration,
		timeout,
	)
}

func isInProgressMission(m Mission) bool {
	state := strings.ToLower(strings.TrimSpace(m.State))
	return state == MissionInProgress
}

func hasLiveAgentSession(agent Agent, activeSessions map[string]struct{}) bool {
	if !isActiveAgentState(agent.State) {
		return false
	}
	sessionID := strings.TrimSpace(agent.SessionID)
	if sessionID == "" {
		return false
	}
	_, ok := activeSessions[sessionID]
	return ok
}

func isActiveAgentState(state string) bool {
	normalized := strings.ToLower(strings.TrimSpace(state))
	return normalized == AgentRunning || normalized == AgentSpawning
}

func (m *Manager) publishRecoverySummary(result Result, auditTimestamp time.Time) {
	if m == nil || m.bus == nil {
		return
	}
	m.bus.Publish(events.Event{
		Type:       events.EventTypeHealthCheck,
		Timestamp:  auditTimestamp,
		EntityType: "recovery",
		EntityID:   "startup",
		Payload: map[string]any{
			"orphaned_mission_ids":   append([]string(nil), result.OrphanedMissionIDs...),
			"cleaned_dead_sessions":  append([]string(nil), result.CleanedDeadSessions...),
			"resume_commission_ids":  append([]string(nil), result.ResumeCommissionIDs...),
			"recovery_duration_msec": result.RecoveryDuration.Milliseconds(),
		},
		Severity: events.SeverityInfo,
	})
}

func (m *Manager) publishAuditEvent(event events.Event) {
	if m == nil || m.bus == nil {
		return
	}
	m.bus.Publish(event)
}
