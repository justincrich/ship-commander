package commission

import (
	"fmt"
	"time"
)

// Status is the lifecycle state of a commission.
type Status string

const (
	// StatusPlanning indicates the commission is being defined.
	StatusPlanning Status = "planning"
	// StatusApproved indicates the commission was approved for execution.
	StatusApproved Status = "approved"
	// StatusExecuting indicates missions are actively being executed.
	StatusExecuting Status = "executing"
	// StatusCompleted indicates all mission work is complete.
	StatusCompleted Status = "completed"
	// StatusShelved indicates the commission was intentionally paused/stopped.
	StatusShelved Status = "shelved"
)

// AC represents an acceptance criterion item.
type AC struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

// UseCase is a use case extracted from a PRD.
type UseCase struct {
	ID                 string `json:"id"`
	Title              string `json:"title"`
	Description        string `json:"description"`
	AcceptanceCriteria []AC   `json:"acceptanceCriteria"`
}

// ScopeConfig represents high-level scope boundaries from a PRD.
type ScopeConfig struct {
	InScope    []string `json:"inScope"`
	OutOfScope []string `json:"outOfScope"`
}

// MissionTrace keeps use-case-to-mission references for traceability.
type MissionTrace struct {
	ID          string   `json:"id"`
	UseCaseRefs []string `json:"useCaseRefs"`
}

// Commission is the parsed and execution-ready representation of a PRD.
type Commission struct {
	ID                 string         `json:"id"`
	Title              string         `json:"title"`
	Status             Status         `json:"status"`
	UseCases           []UseCase      `json:"useCases"`
	AcceptanceCriteria []AC           `json:"acceptanceCriteria"`
	FunctionalGroups   []string       `json:"functionalGroups"`
	ScopeBoundaries    ScopeConfig    `json:"scopeBoundaries"`
	PRDContent         string         `json:"prdContent"`
	Missions           []MissionTrace `json:"missions"`
	CreatedAt          time.Time      `json:"createdAt"`
}

var allowedTransitions = map[Status]map[Status]struct{}{
	StatusPlanning: {
		StatusApproved: {},
	},
	StatusApproved: {
		StatusExecuting: {},
	},
	StatusExecuting: {
		StatusCompleted: {},
		StatusShelved:   {},
	},
}

// TransitionTo enforces legal lifecycle transitions.
func (c *Commission) TransitionTo(next Status) error {
	if c == nil {
		return fmt.Errorf("commission is nil")
	}
	if c.Status == next {
		return nil
	}

	allowed, ok := allowedTransitions[c.Status]
	if !ok {
		return fmt.Errorf("cannot transition commission from %q", c.Status)
	}
	if _, ok := allowed[next]; !ok {
		return fmt.Errorf("cannot transition commission from %q to %q", c.Status, next)
	}

	c.Status = next
	return nil
}
