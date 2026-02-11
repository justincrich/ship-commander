package state

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// EntityType identifies which state machine to evaluate.
type EntityType string

const (
	// EntityCommission is the commission lifecycle state machine.
	EntityCommission EntityType = "commission"
	// EntityMission is the mission lifecycle state machine.
	EntityMission EntityType = "mission"
	// EntityAC is the acceptance-criterion phase state machine.
	EntityAC EntityType = "ac"
	// EntityAgent is the agent lifecycle state machine.
	EntityAgent EntityType = "agent"
)

const (
	CommissionPlanning  = "planning"
	CommissionApproved  = "approved"
	CommissionExecuting = "executing"
	CommissionCompleted = "completed"
	CommissionShelved   = "shelved"
)

const (
	MissionBacklog    = "backlog"
	MissionInProgress = "in_progress"
	MissionReview     = "review"
	MissionApproved   = "approved"
	MissionDone       = "done"
	MissionHalted     = "halted"
)

const (
	ACRed            = "red"
	ACVerifyRed      = "verify_red"
	ACGreen          = "green"
	ACVerifyGreen    = "verify_green"
	ACRefactor       = "refactor"
	ACVerifyRefactor = "verify_refactor"
	ACComplete       = "complete"
)

const (
	AgentIdle     = "idle"
	AgentSpawning = "spawning"
	AgentRunning  = "running"
	AgentStuck    = "stuck"
	AgentDone     = "done"
	AgentDead     = "dead"
)

var allowedTransitions = map[EntityType]map[string]map[string]struct{}{
	EntityCommission: {
		CommissionPlanning: {
			CommissionApproved: {},
		},
		CommissionApproved: {
			CommissionExecuting: {},
		},
		CommissionExecuting: {
			CommissionCompleted: {},
			CommissionShelved:   {},
		},
	},
	EntityMission: {
		MissionBacklog: {
			MissionInProgress: {},
		},
		MissionInProgress: {
			MissionReview: {},
		},
		MissionReview: {
			MissionApproved: {},
		},
		MissionApproved: {
			MissionDone:   {},
			MissionHalted: {},
		},
	},
	EntityAC: {
		ACRed: {
			ACVerifyRed: {},
		},
		ACVerifyRed: {
			ACGreen: {},
		},
		ACGreen: {
			ACVerifyGreen: {},
		},
		ACVerifyGreen: {
			ACRefactor: {},
		},
		ACRefactor: {
			ACVerifyRefactor: {},
		},
		ACVerifyRefactor: {
			ACComplete: {},
		},
	},
	EntityAgent: {
		AgentIdle: {
			AgentSpawning: {},
		},
		AgentSpawning: {
			AgentRunning: {},
		},
		AgentRunning: {
			AgentStuck: {},
		},
		AgentStuck: {
			AgentDone: {},
			AgentDead: {},
		},
	},
}

// Persister writes transition outcomes into Beads state and comments.
type Persister interface {
	SetState(id, key, value string) error
	AddComment(id, comment string) error
}

// TransitionRecord stores transition metadata for local history.
type TransitionRecord struct {
	EntityType EntityType
	EntityID   string
	FromState  string
	ToState    string
	Reason     string
	Actor      string
	Timestamp  time.Time
}

// IllegalTransitionError is returned for a disallowed transition.
type IllegalTransitionError struct {
	EntityType EntityType
	EntityID   string
	FromState  string
	ToState    string
	Reason     string
}

func (e *IllegalTransitionError) Error() string {
	reason := strings.TrimSpace(e.Reason)
	if reason == "" {
		reason = "illegal transition for entity lifecycle"
	}
	return fmt.Sprintf(
		"cannot transition %s %q from %q to %q: %s",
		e.EntityType,
		e.EntityID,
		e.FromState,
		e.ToState,
		reason,
	)
}

// Is enables errors.Is checks for illegal transition failures.
func (e *IllegalTransitionError) Is(target error) bool {
	_, ok := target.(*IllegalTransitionError)
	return ok
}

// Machine validates and persists deterministic state transitions.
type Machine struct {
	persister Persister
	actor     string
	now       func() time.Time
	history   []TransitionRecord
}

// NewMachine builds a deterministic state machine persister.
func NewMachine(persister Persister, actor string) (*Machine, error) {
	if persister == nil {
		return nil, errors.New("persister is required")
	}

	normalizedActor := strings.TrimSpace(actor)
	if normalizedActor == "" {
		normalizedActor = "commander"
	}

	return &Machine{
		persister: persister,
		actor:     normalizedActor,
		now:       time.Now,
		history:   []TransitionRecord{},
	}, nil
}

// Transition validates and persists one state transition.
func (m *Machine) Transition(entityType EntityType, entityID, fromState, toState, reason string) error {
	if m == nil {
		return errors.New("machine is nil")
	}

	entityID = strings.TrimSpace(entityID)
	if entityID == "" {
		return errors.New("entity id must not be empty")
	}
	fromState = strings.TrimSpace(fromState)
	toState = strings.TrimSpace(toState)
	if fromState == "" || toState == "" {
		return errors.New("from and to states must not be empty")
	}

	if !isAllowed(entityType, fromState, toState) {
		return &IllegalTransitionError{
			EntityType: entityType,
			EntityID:   entityID,
			FromState:  fromState,
			ToState:    toState,
			Reason:     "illegal transition for entity lifecycle",
		}
	}

	timestamp := m.now().UTC()
	record := TransitionRecord{
		EntityType: entityType,
		EntityID:   entityID,
		FromState:  fromState,
		ToState:    toState,
		Reason:     strings.TrimSpace(reason),
		Actor:      m.actor,
		Timestamp:  timestamp,
	}

	if err := m.persister.SetState(entityID, stateDimension(entityType), toState); err != nil {
		return fmt.Errorf("persist state transition for %s: %w", entityID, err)
	}

	comment := fmt.Sprintf(
		"state_transition entity=%s from=%s to=%s actor=%s timestamp=%s reason=%q",
		entityType,
		fromState,
		toState,
		record.Actor,
		timestamp.Format(time.RFC3339),
		record.Reason,
	)
	if err := m.persister.AddComment(entityID, comment); err != nil {
		return fmt.Errorf("persist transition event for %s: %w", entityID, err)
	}

	m.history = append(m.history, record)
	return nil
}

// History returns transition records captured by this machine.
func (m *Machine) History() []TransitionRecord {
	if m == nil {
		return nil
	}
	out := make([]TransitionRecord, len(m.history))
	copy(out, m.history)
	return out
}

func isAllowed(entityType EntityType, fromState, toState string) bool {
	entityTransitions, ok := allowedTransitions[entityType]
	if !ok {
		return false
	}
	nextStates, ok := entityTransitions[fromState]
	if !ok {
		return false
	}
	_, ok = nextStates[toState]
	return ok
}

func stateDimension(entityType EntityType) string {
	switch entityType {
	case EntityCommission:
		return "commission_state"
	case EntityMission:
		return "mission_state"
	case EntityAC:
		return "ac_state"
	case EntityAgent:
		return "agent_state"
	default:
		return "state"
	}
}
