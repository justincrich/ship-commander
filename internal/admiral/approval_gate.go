package admiral

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

const defaultApprovalGateBuffer = 1

// CoverageStatus captures use-case coverage status presented during manifest approval.
type CoverageStatus string

const (
	// CoverageStatusCovered indicates a use case is fully covered by signed missions.
	CoverageStatusCovered CoverageStatus = "covered"
	// CoverageStatusPartial indicates a use case is only partially covered.
	CoverageStatusPartial CoverageStatus = "partial"
	// CoverageStatusUncovered indicates a use case is currently uncovered.
	CoverageStatusUncovered CoverageStatus = "uncovered"
)

// Mission is a mission summary sent to Admiral during manifest approval.
//
//nolint:revive // Field names follow the issue contract.
type Mission struct {
	ID                        string
	Title                     string
	DependsOn                 []string
	UseCaseIDs                []string
	Classification            string
	ClassificationRationale   string
	ClassificationCriteria    []string
	ClassificationConfidence  string
	ClassificationNeedsReview bool
}

// Wave is one deterministic execution wave assignment for approval review.
//
//nolint:revive // Field names follow the issue contract.
type Wave struct {
	Index      int
	MissionIDs []string
}

// ApprovalDecision is the deterministic decision value returned by Admiral.
type ApprovalDecision string

const (
	// ApprovalDecisionApproved means execution can begin.
	ApprovalDecisionApproved ApprovalDecision = "Approved"
	// ApprovalDecisionFeedback means planning must reconvene with feedback.
	ApprovalDecisionFeedback ApprovalDecision = "Feedback"
	// ApprovalDecisionShelved means the plan is persisted and paused.
	ApprovalDecisionShelved ApprovalDecision = "Shelved"
)

// ApprovalRequest is the manifest approval payload presented to Admiral.
//
//nolint:revive // Field names follow the issue contract.
type ApprovalRequest struct {
	CommissionID    string
	MissionManifest []Mission
	WaveAssignments []Wave
	CoverageMap     map[string]CoverageStatus
	Iteration       int
	MaxIterations   int
}

// ApprovalResponse is the Admiral decision payload for manifest review.
//
//nolint:revive // Field names follow the issue contract.
type ApprovalResponse struct {
	Decision     ApprovalDecision
	FeedbackText string
}

// ApprovalRecord captures one approval request/response interaction.
type ApprovalRecord struct {
	Request    ApprovalRequest
	Response   ApprovalResponse
	AskedAt    time.Time
	AnsweredAt time.Time
}

// ApprovalGate is a blocking gate that prevents mission dispatch before Admiral decision.
type ApprovalGate struct {
	requests  chan ApprovalRequest
	responses chan ApprovalResponse
	now       func() time.Time

	mu      sync.Mutex
	history []ApprovalRecord
}

// NewApprovalGate constructs a blocking approval gate.
func NewApprovalGate(bufferSize int) *ApprovalGate {
	if bufferSize <= 0 {
		bufferSize = defaultApprovalGateBuffer
	}
	return &ApprovalGate{
		requests:  make(chan ApprovalRequest, bufferSize),
		responses: make(chan ApprovalResponse, bufferSize),
		now:       time.Now,
		history:   make([]ApprovalRecord, 0),
	}
}

// Requests exposes approval requests to subscribers (for example, TUI approval modal handling).
func (g *ApprovalGate) Requests() <-chan ApprovalRequest {
	return g.requests
}

// Respond publishes an Admiral decision for a pending approval request.
func (g *ApprovalGate) Respond(response ApprovalResponse) error {
	if g == nil {
		return errors.New("approval gate is nil")
	}
	normalized, err := normalizeApprovalResponse(response)
	if err != nil {
		return err
	}

	g.responses <- normalized
	return nil
}

// AwaitDecision presents one approval request and blocks until Admiral responds or context is canceled.
func (g *ApprovalGate) AwaitDecision(ctx context.Context, request ApprovalRequest) (ApprovalResponse, error) {
	if g == nil {
		return ApprovalResponse{}, errors.New("approval gate is nil")
	}

	normalizedRequest, err := normalizeApprovalRequest(request)
	if err != nil {
		return ApprovalResponse{}, err
	}
	askedAt := g.now().UTC()

	select {
	case g.requests <- normalizedRequest:
	case <-ctx.Done():
		return ApprovalResponse{}, ctx.Err()
	}

	select {
	case response := <-g.responses:
		record := ApprovalRecord{
			Request:    normalizedRequest,
			Response:   response,
			AskedAt:    askedAt,
			AnsweredAt: g.now().UTC(),
		}
		g.mu.Lock()
		g.history = append(g.history, record)
		g.mu.Unlock()
		return response, nil
	case <-ctx.Done():
		return ApprovalResponse{}, ctx.Err()
	}
}

// History returns a copy of approval request/response records.
func (g *ApprovalGate) History() []ApprovalRecord {
	if g == nil {
		return nil
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	history := make([]ApprovalRecord, len(g.history))
	copy(history, g.history)
	return history
}

func normalizeApprovalRequest(request ApprovalRequest) (ApprovalRequest, error) {
	request.CommissionID = strings.TrimSpace(request.CommissionID)
	if request.CommissionID == "" {
		return ApprovalRequest{}, errors.New("commission id is required")
	}

	request.MissionManifest = normalizeApprovalMissions(request.MissionManifest)
	if len(request.MissionManifest) == 0 {
		return ApprovalRequest{}, errors.New("mission manifest is required")
	}

	request.WaveAssignments = normalizeWaves(request.WaveAssignments)

	coverage := make(map[string]CoverageStatus, len(request.CoverageMap))
	for useCaseID, status := range request.CoverageMap {
		trimmedID := strings.TrimSpace(useCaseID)
		if trimmedID == "" {
			continue
		}
		coverage[trimmedID] = normalizeCoverageStatus(status)
	}
	request.CoverageMap = coverage

	if request.Iteration <= 0 {
		request.Iteration = 1
	}
	if request.MaxIterations <= 0 {
		request.MaxIterations = request.Iteration
	}
	if request.Iteration > request.MaxIterations {
		return ApprovalRequest{}, fmt.Errorf("iteration %d exceeds max iterations %d", request.Iteration, request.MaxIterations)
	}

	return request, nil
}

func normalizeApprovalMissions(missions []Mission) []Mission {
	normalized := make([]Mission, 0, len(missions))
	for _, mission := range missions {
		mission.ID = strings.TrimSpace(mission.ID)
		if mission.ID == "" {
			continue
		}
		mission.Title = strings.TrimSpace(mission.Title)
		mission.DependsOn = normalizeStringSlice(mission.DependsOn)
		mission.UseCaseIDs = normalizeStringSlice(mission.UseCaseIDs)
		mission.Classification = strings.ToUpper(strings.TrimSpace(mission.Classification))
		mission.ClassificationRationale = strings.TrimSpace(mission.ClassificationRationale)
		mission.ClassificationCriteria = normalizeStringSlice(mission.ClassificationCriteria)
		mission.ClassificationConfidence = strings.ToLower(strings.TrimSpace(mission.ClassificationConfidence))
		normalized = append(normalized, mission)
	}
	return normalized
}

func normalizeWaves(waves []Wave) []Wave {
	normalized := make([]Wave, 0, len(waves))
	for _, wave := range waves {
		if wave.Index <= 0 {
			continue
		}
		wave.MissionIDs = normalizeStringSlice(wave.MissionIDs)
		normalized = append(normalized, wave)
	}
	return normalized
}

func normalizeStringSlice(values []string) []string {
	normalized := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		normalized = append(normalized, value)
	}
	return normalized
}

func normalizeCoverageStatus(status CoverageStatus) CoverageStatus {
	switch strings.ToLower(strings.TrimSpace(string(status))) {
	case string(CoverageStatusCovered):
		return CoverageStatusCovered
	case string(CoverageStatusPartial):
		return CoverageStatusPartial
	case string(CoverageStatusUncovered):
		return CoverageStatusUncovered
	default:
		return CoverageStatusUncovered
	}
}

func normalizeApprovalResponse(response ApprovalResponse) (ApprovalResponse, error) {
	response.FeedbackText = strings.TrimSpace(response.FeedbackText)

	switch strings.ToLower(strings.TrimSpace(string(response.Decision))) {
	case strings.ToLower(string(ApprovalDecisionApproved)):
		response.Decision = ApprovalDecisionApproved
	case strings.ToLower(string(ApprovalDecisionFeedback)):
		response.Decision = ApprovalDecisionFeedback
	case strings.ToLower(string(ApprovalDecisionShelved)):
		response.Decision = ApprovalDecisionShelved
	default:
		return ApprovalResponse{}, fmt.Errorf("invalid approval decision %q", response.Decision)
	}

	if response.Decision == ApprovalDecisionFeedback && response.FeedbackText == "" {
		return ApprovalResponse{}, errors.New("feedback text is required when decision is Feedback")
	}

	return response, nil
}
