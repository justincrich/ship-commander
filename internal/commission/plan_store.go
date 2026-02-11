package commission

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	planStorageVersion = "v1"
)

// PlanningStatus is the persisted plan lifecycle status.
type PlanningStatus string

const (
	// PlanningStatusApproved indicates the persisted plan is approved for execution.
	PlanningStatusApproved PlanningStatus = "approved"
	// PlanningStatusShelved indicates the persisted plan is intentionally shelved.
	PlanningStatusShelved PlanningStatus = "shelved"
)

// PlanMission captures one persisted mission entry in a manifest plan.
//
//nolint:revive // Field names match issue contract.
type PlanMission struct {
	ID                         string   `json:"id"`
	Title                      string   `json:"title,omitempty"`
	UseCaseIDs                 []string `json:"useCaseIds,omitempty"`
	Classification             string   `json:"classification,omitempty"`
	ClassificationRationale    string   `json:"classificationRationale,omitempty"`
	ClassificationCriteria     []string `json:"classificationCriteria,omitempty"`
	ClassificationConfidence   string   `json:"classificationConfidence,omitempty"`
	ClassificationNeedsReview  bool     `json:"classificationNeedsReview,omitempty"`
	ClassificationReviewSource string   `json:"classificationReviewSource,omitempty"`
}

// PlanMessage captures a persisted Ready Room message.
//
//nolint:revive // Field names match issue contract.
type PlanMessage struct {
	From      string    `json:"from,omitempty"`
	To        string    `json:"to,omitempty"`
	Type      string    `json:"type,omitempty"`
	Domain    string    `json:"domain,omitempty"`
	Content   string    `json:"content,omitempty"`
	Timestamp time.Time `json:"timestamp,omitempty"`
}

// PlanSignoff captures persisted captain/commander/design signoff state for a mission.
//
//nolint:revive // Field names match issue contract.
type PlanSignoff struct {
	Captain       bool `json:"captain"`
	Commander     bool `json:"commander"`
	DesignOfficer bool `json:"designOfficer"`
}

// PlanWave captures persisted wave assignments for mission execution.
//
//nolint:revive // Field names match issue contract.
type PlanWave struct {
	Index      int      `json:"index"`
	MissionIDs []string `json:"missionIds"`
}

// PlanState is the full persisted mission manifest planning state.
//
//nolint:revive // Field names match issue contract.
type PlanState struct {
	MissionList       []PlanMission          `json:"missionList"`
	ReadyRoomMessages []PlanMessage          `json:"readyRoomMessages"`
	SignoffMap        map[string]PlanSignoff `json:"signoffMap"`
	IterationCount    int                    `json:"iterationCount"`
	CoverageMap       map[string]string      `json:"coverageMap"`
	WaveAssignments   []PlanWave             `json:"waveAssignments"`
}

// ResumeResult captures state needed to resume planning and re-spawn agent sessions.
type ResumeResult struct {
	State        PlanState
	RespawnRoles []string
}

type persistedPlanEnvelope struct {
	Version          string         `json:"version"`
	CommissionID     string         `json:"commissionId"`
	CommissionStatus PlanningStatus `json:"commissionStatus"`
	FeedbackText     string         `json:"feedbackText,omitempty"`
	SavedAt          time.Time      `json:"savedAt"`
	State            PlanState      `json:"state"`
}

type beadsShowRecord struct {
	ID    string `json:"id"`
	Notes string `json:"notes"`
}

// SavePlan persists full mission manifest state to Beads notes.
func SavePlan(ctx context.Context, commissionID string, state PlanState) error {
	return SavePlanWithRunner(ctx, commissionID, state, defaultCommandRunner{})
}

// SavePlanWithRunner persists full mission manifest state to Beads notes with a custom runner.
func SavePlanWithRunner(ctx context.Context, commissionID string, state PlanState, runner CommandRunner) error {
	envelope, err := newPlanEnvelope(commissionID, PlanningStatusApproved, state, "")
	if err != nil {
		return err
	}
	return writePlanEnvelope(ctx, envelope, runner)
}

// ShelvePlan persists full state and marks commission status as shelved.
func ShelvePlan(ctx context.Context, commissionID string, feedbackText string) error {
	return ShelvePlanWithRunner(ctx, commissionID, feedbackText, defaultCommandRunner{})
}

// ShelvePlanWithRunner persists full state and marks commission status as shelved using a custom runner.
func ShelvePlanWithRunner(ctx context.Context, commissionID string, feedbackText string, runner CommandRunner) error {
	if runner == nil {
		return errors.New("runner must not be nil")
	}
	commissionID = strings.TrimSpace(commissionID)
	if commissionID == "" {
		return errors.New("commission id must not be empty")
	}

	current, err := readPlanEnvelope(ctx, commissionID, runner)
	if err != nil {
		return err
	}

	current.CommissionStatus = PlanningStatusShelved
	current.FeedbackText = strings.TrimSpace(feedbackText)
	current.SavedAt = time.Now().UTC()
	return writePlanEnvelope(ctx, current, runner)
}

// LoadPlan loads full persisted mission manifest state.
func LoadPlan(ctx context.Context, commissionID string) (PlanState, error) {
	return LoadPlanWithRunner(ctx, commissionID, defaultCommandRunner{})
}

// LoadPlanWithRunner loads full persisted mission manifest state with a custom runner.
func LoadPlanWithRunner(ctx context.Context, commissionID string, runner CommandRunner) (PlanState, error) {
	envelope, err := readPlanEnvelope(ctx, commissionID, runner)
	if err != nil {
		return PlanState{}, err
	}
	return envelope.State, nil
}

// ResumePlan loads state and returns the deterministic set of planning sessions to re-spawn.
func ResumePlan(ctx context.Context, commissionID string) (ResumeResult, error) {
	return ResumePlanWithRunner(ctx, commissionID, defaultCommandRunner{})
}

// ResumePlanWithRunner loads state and returns deterministic planning sessions to re-spawn with a custom runner.
func ResumePlanWithRunner(ctx context.Context, commissionID string, runner CommandRunner) (ResumeResult, error) {
	state, err := LoadPlanWithRunner(ctx, commissionID, runner)
	if err != nil {
		return ResumeResult{}, err
	}
	return ResumeResult{
		State:        state,
		RespawnRoles: []string{"captain", "commander", "designOfficer"},
	}, nil
}

// ReexecutePlan loads a previously approved plan for execution pipeline re-entry.
func ReexecutePlan(ctx context.Context, commissionID string) (PlanState, error) {
	return ReexecutePlanWithRunner(ctx, commissionID, defaultCommandRunner{})
}

// ReexecutePlanWithRunner loads a previously approved plan for execution pipeline re-entry with a custom runner.
func ReexecutePlanWithRunner(ctx context.Context, commissionID string, runner CommandRunner) (PlanState, error) {
	envelope, err := readPlanEnvelope(ctx, commissionID, runner)
	if err != nil {
		return PlanState{}, err
	}
	if envelope.CommissionStatus != PlanningStatusApproved && envelope.CommissionStatus != PlanningStatusShelved {
		return PlanState{}, fmt.Errorf("plan for %s is not approved for re-execution (status=%s)", envelope.CommissionID, envelope.CommissionStatus)
	}
	return envelope.State, nil
}

func newPlanEnvelope(
	commissionID string,
	status PlanningStatus,
	state PlanState,
	feedbackText string,
) (persistedPlanEnvelope, error) {
	commissionID = strings.TrimSpace(commissionID)
	if commissionID == "" {
		return persistedPlanEnvelope{}, errors.New("commission id must not be empty")
	}
	if status == "" {
		return persistedPlanEnvelope{}, errors.New("commission status must not be empty")
	}
	if state.IterationCount <= 0 {
		return persistedPlanEnvelope{}, errors.New("iteration count must be positive")
	}
	if len(state.MissionList) == 0 {
		return persistedPlanEnvelope{}, errors.New("mission list must not be empty")
	}

	return persistedPlanEnvelope{
		Version:          planStorageVersion,
		CommissionID:     commissionID,
		CommissionStatus: status,
		FeedbackText:     strings.TrimSpace(feedbackText),
		SavedAt:          time.Now().UTC(),
		State:            state,
	}, nil
}

func writePlanEnvelope(ctx context.Context, envelope persistedPlanEnvelope, runner CommandRunner) error {
	if runner == nil {
		return errors.New("runner must not be nil")
	}
	payload, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("marshal plan envelope: %w", err)
	}

	_, err = runner.Run(ctx, "bd", "update", envelope.CommissionID, "--notes", string(payload))
	if err != nil {
		return fmt.Errorf("persist plan envelope for %s: %w", envelope.CommissionID, err)
	}
	return nil
}

func readPlanEnvelope(ctx context.Context, commissionID string, runner CommandRunner) (persistedPlanEnvelope, error) {
	if runner == nil {
		return persistedPlanEnvelope{}, errors.New("runner must not be nil")
	}
	commissionID = strings.TrimSpace(commissionID)
	if commissionID == "" {
		return persistedPlanEnvelope{}, errors.New("commission id must not be empty")
	}

	out, err := runner.Run(ctx, "bd", "show", commissionID, "--json")
	if err != nil {
		return persistedPlanEnvelope{}, fmt.Errorf("read persisted plan for %s: %w", commissionID, err)
	}

	var records []beadsShowRecord
	if err := json.Unmarshal(out, &records); err != nil {
		return persistedPlanEnvelope{}, fmt.Errorf("parse persisted plan record: %w", err)
	}
	if len(records) == 0 {
		return persistedPlanEnvelope{}, fmt.Errorf("commission %s not found", commissionID)
	}
	notes := strings.TrimSpace(records[0].Notes)
	if notes == "" {
		return persistedPlanEnvelope{}, fmt.Errorf("commission %s has no persisted plan", commissionID)
	}

	var envelope persistedPlanEnvelope
	if err := json.Unmarshal([]byte(notes), &envelope); err != nil {
		return persistedPlanEnvelope{}, fmt.Errorf("parse persisted plan envelope: %w", err)
	}
	if envelope.Version != planStorageVersion {
		return persistedPlanEnvelope{}, fmt.Errorf("unsupported plan storage version %q", envelope.Version)
	}
	if envelope.CommissionID != commissionID {
		return persistedPlanEnvelope{}, fmt.Errorf("persisted plan commission id mismatch: got %q want %q", envelope.CommissionID, commissionID)
	}
	return envelope, nil
}
