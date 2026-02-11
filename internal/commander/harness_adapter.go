package commander

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ship-commander/sc3/internal/config"
	"github.com/ship-commander/sc3/internal/harness"
	"github.com/ship-commander/sc3/internal/protocol"
)

const (
	implementerRoleKey = "ensign"
	reviewerRoleKey    = "reviewer"
)

// ClaudeHarnessAdapter implements Commander Harness using a tmux-backed harness driver.
type ClaudeHarnessAdapter struct {
	driver       harness.HarnessDriver
	protocol     protocol.EventStore
	cfg          *config.Config
	availability map[string]bool
	now          func() time.Time
}

// NewClaudeHarnessAdapter constructs a Commander harness adapter.
func NewClaudeHarnessAdapter(
	driver harness.HarnessDriver,
	protocolStore protocol.EventStore,
	cfg *config.Config,
	availability map[string]bool,
) (*ClaudeHarnessAdapter, error) {
	if driver == nil {
		return nil, errors.New("harness driver is required")
	}
	if protocolStore == nil {
		return nil, errors.New("protocol event store is required")
	}
	if cfg == nil {
		return nil, errors.New("config is required")
	}
	copiedAvailability := make(map[string]bool, len(availability))
	for key, value := range availability {
		copiedAvailability[strings.ToLower(strings.TrimSpace(key))] = value
	}
	return &ClaudeHarnessAdapter{
		driver:       driver,
		protocol:     protocolStore,
		cfg:          cfg,
		availability: copiedAvailability,
		now:          time.Now,
	}, nil
}

// DispatchImplementer builds a mission prompt, dispatches a session, then captures and parses claims.
func (a *ClaudeHarnessAdapter) DispatchImplementer(ctx context.Context, req DispatchRequest) (DispatchResult, error) {
	if a == nil {
		return DispatchResult{}, errors.New("adapter is nil")
	}
	missionID := strings.TrimSpace(req.Mission.ID)
	if missionID == "" {
		return DispatchResult{}, errors.New("mission id is required")
	}

	prompt, err := a.buildImplementerPrompt(req)
	if err != nil {
		return DispatchResult{}, err
	}

	model, err := a.resolveRoleModel(implementerRoleKey, req.Mission, req.Mission.Model)
	if err != nil {
		return DispatchResult{}, err
	}

	session, err := a.driver.SpawnSession(
		implementerRoleKey,
		prompt,
		req.WorktreePath,
		harness.SessionOpts{Model: model, MaxTurns: 1},
	)
	if err != nil {
		return DispatchResult{}, fmt.Errorf("spawn implementer session for %s: %w", missionID, err)
	}
	if session == nil || strings.TrimSpace(session.ID) == "" {
		return DispatchResult{}, fmt.Errorf("spawn implementer session for %s: empty session", missionID)
	}

	if output, captureErr := a.driver.SendMessage(session, ""); captureErr == nil {
		if parseErr := a.persistImplementerClaims(ctx, req.Mission, session.ID, output); parseErr != nil {
			return DispatchResult{}, parseErr
		}
	}

	return DispatchResult{SessionID: strings.TrimSpace(session.ID)}, nil
}

// DispatchReviewer builds reviewer context, dispatches independent reviewer, and records review verdict events.
func (a *ClaudeHarnessAdapter) DispatchReviewer(ctx context.Context, req ReviewerDispatchRequest) (DispatchResult, error) {
	if a == nil {
		return DispatchResult{}, errors.New("adapter is nil")
	}
	missionID := strings.TrimSpace(req.Mission.ID)
	if missionID == "" {
		return DispatchResult{}, errors.New("mission id is required")
	}

	prompt, err := BuildReviewerPrompt(ReviewerPromptContext{
		MissionID:          req.Mission.ID,
		Title:              req.Mission.Title,
		Classification:     req.Mission.Classification,
		AcceptanceCriteria: req.AcceptanceCriteria,
		GateEvidence:       req.GateEvidence,
		CodeDiff:           req.CodeDiff,
		DemoTokenContent:   req.DemoTokenContent,
	})
	if err != nil {
		return DispatchResult{}, fmt.Errorf("build reviewer prompt for %s: %w", missionID, err)
	}

	model, err := a.resolveRoleModel(reviewerRoleKey, req.Mission, req.Mission.Model)
	if err != nil {
		return DispatchResult{}, err
	}

	session, err := a.driver.SpawnSession(
		reviewerRoleKey,
		prompt,
		req.WorktreePath,
		harness.SessionOpts{Model: model, MaxTurns: 1},
	)
	if err != nil {
		return DispatchResult{}, fmt.Errorf("spawn reviewer session for %s: %w", missionID, err)
	}
	if session == nil || strings.TrimSpace(session.ID) == "" {
		return DispatchResult{}, fmt.Errorf("spawn reviewer session for %s: empty session", missionID)
	}

	if output, captureErr := a.driver.SendMessage(session, ""); captureErr == nil {
		if parseErr := a.persistReviewVerdict(ctx, req.Mission, req.ImplementerSessionID, session.ID, output); parseErr != nil {
			return DispatchResult{}, parseErr
		}
	}

	return DispatchResult{SessionID: strings.TrimSpace(session.ID)}, nil
}

func (a *ClaudeHarnessAdapter) buildImplementerPrompt(req DispatchRequest) (string, error) {
	input := ImplementerPromptContext{
		MissionID:           req.Mission.ID,
		Title:               req.Mission.Title,
		Classification:      req.Mission.Classification,
		UseCases:            req.Mission.UseCaseIDs,
		WorktreePath:        req.WorktreePath,
		AcceptanceCriterion: req.ReviewerFeedback,
		MissionSpec:         req.Mission.ClassificationRationale,
		PriorContext:        req.WaveFeedback,
		GateFeedback:        req.ReviewerFeedback,
	}
	if isStandardOpsMission(req.Mission) {
		return BuildStandardOpsPrompt(input)
	}
	if strings.TrimSpace(req.ReviewerFeedback) != "" {
		return BuildGREENPrompt(input)
	}
	return BuildREDPrompt(input)
}

func (a *ClaudeHarnessAdapter) resolveRoleModel(role string, mission Mission, fallbackModel string) (string, error) {
	domain := ""
	if len(mission.UseCaseIDs) > 0 {
		domain = mission.UseCaseIDs[0]
	}
	harnessName, modelName, _, err := a.cfg.ResolveHarnessModel(role, domain, a.availability)
	if err != nil {
		return "", fmt.Errorf("resolve harness/model for role %s mission %s: %w", role, mission.ID, err)
	}
	if strings.TrimSpace(harnessName) != "claude" {
		return "", fmt.Errorf("resolved harness %q for mission %s; claude adapter requires claude", harnessName, mission.ID)
	}
	if strings.TrimSpace(modelName) == "" {
		modelName = strings.TrimSpace(fallbackModel)
	}
	if strings.TrimSpace(modelName) == "" {
		modelName = "sonnet"
	}
	return modelName, nil
}

func (a *ClaudeHarnessAdapter) persistImplementerClaims(ctx context.Context, mission Mission, sessionID, output string) error {
	claims := parseImplementerClaims(output)
	for _, claim := range claims {
		payload, err := json.Marshal(map[string]string{
			"claim_type": claim.claimType,
			"source":     "harness-output",
		})
		if err != nil {
			return fmt.Errorf("marshal claim payload for mission %s: %w", mission.ID, err)
		}
		if err := a.protocol.Append(ctx, protocol.ProtocolEvent{
			ProtocolVersion: protocol.ProtocolVersion,
			Type:            protocol.EventTypeAgentClaim,
			MissionID:       mission.ID,
			ACID:            claim.acID,
			AgentID:         strings.TrimSpace(sessionID),
			Payload:         payload,
			Timestamp:       a.now().UTC(),
		}); err != nil {
			return fmt.Errorf("append claim event for mission %s: %w", mission.ID, err)
		}
	}
	return nil
}

func (a *ClaudeHarnessAdapter) persistReviewVerdict(
	ctx context.Context,
	mission Mission,
	implementerSessionID,
	reviewerSessionID,
	output string,
) error {
	verdict, feedback, ok := parseReviewVerdictOutput(output)
	if !ok {
		return nil
	}
	payload, err := json.Marshal(map[string]string{
		"verdict":                verdict,
		"feedback":               feedback,
		"implementer_session_id": strings.TrimSpace(implementerSessionID),
		"reviewer_session_id":    strings.TrimSpace(reviewerSessionID),
	})
	if err != nil {
		return fmt.Errorf("marshal review verdict payload for mission %s: %w", mission.ID, err)
	}
	if err := a.protocol.Append(ctx, protocol.ProtocolEvent{
		ProtocolVersion: protocol.ProtocolVersion,
		Type:            protocol.EventTypeReviewComplete,
		MissionID:       mission.ID,
		Payload:         payload,
		Timestamp:       a.now().UTC(),
	}); err != nil {
		return fmt.Errorf("append review verdict event for mission %s: %w", mission.ID, err)
	}
	return nil
}

type parsedClaim struct {
	claimType string
	acID      string
}

func parseImplementerClaims(output string) []parsedClaim {
	lines := strings.Split(output, "\n")
	claims := make([]parsedClaim, 0)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var payload map[string]any
		if err := json.Unmarshal([]byte(line), &payload); err != nil {
			continue
		}
		claimType := strings.ToUpper(strings.TrimSpace(firstNonEmptyMap(payload, "claim_type", "claimType", "event_type")))
		if !isSupportedClaimType(claimType) {
			continue
		}
		acID := strings.TrimSpace(firstNonEmptyMap(payload, "ac_id", "acID"))
		if acID == "" {
			acID = "AC-UNSPECIFIED"
		}
		claims = append(claims, parsedClaim{claimType: claimType, acID: acID})
	}
	return claims
}

func parseReviewVerdictOutput(output string) (string, string, bool) {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var payload map[string]any
		if err := json.Unmarshal([]byte(line), &payload); err != nil {
			continue
		}
		verdict := strings.ToUpper(strings.TrimSpace(firstNonEmptyMap(payload, "verdict", "decision")))
		if verdict != protocol.ReviewVerdictApproved && verdict != protocol.ReviewVerdictNeedsFixes {
			continue
		}
		feedback := strings.TrimSpace(firstNonEmptyMap(payload, "feedback", "feedback_text", "feedbackText"))
		return verdict, feedback, true
	}
	return "", "", false
}

func isSupportedClaimType(value string) bool {
	switch strings.TrimSpace(value) {
	case protocol.ClaimTypeREDComplete, protocol.ClaimTypeGREENComplete, protocol.ClaimTypeREFACTORComplete, protocol.ClaimTypeIMPLEMENTComplete:
		return true
	default:
		return false
	}
}
