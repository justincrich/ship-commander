package protocol

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ship-commander/sc3/internal/events"
)

const (
	// ProtocolVersion identifies the supported event schema version.
	ProtocolVersion = "1.0"

	// EventTypeAgentClaim represents implementer claims (RED/GREEN/REFACTOR/IMPLEMENT complete).
	EventTypeAgentClaim = "AGENT_CLAIM"
	// EventTypeGateResult represents deterministic gate execution outcomes.
	EventTypeGateResult = "GATE_RESULT"
	// EventTypeStateTransition represents orchestrator state transitions.
	EventTypeStateTransition = "STATE_TRANSITION"
	// EventTypeReviewComplete represents reviewer verdict completion for a mission.
	EventTypeReviewComplete = "REVIEW_COMPLETE"
)

const (
	// ClaimTypeREDComplete indicates RED phase completion claim.
	ClaimTypeREDComplete = "RED_COMPLETE"
	// ClaimTypeGREENComplete indicates GREEN phase completion claim.
	ClaimTypeGREENComplete = "GREEN_COMPLETE"
	// ClaimTypeREFACTORComplete indicates REFACTOR phase completion claim.
	ClaimTypeREFACTORComplete = "REFACTOR_COMPLETE"
	// ClaimTypeIMPLEMENTComplete indicates IMPLEMENT phase completion claim.
	ClaimTypeIMPLEMENTComplete = "IMPLEMENT_COMPLETE"
	// ReviewVerdictApproved indicates reviewer accepted mission output.
	ReviewVerdictApproved = "APPROVED"
	// ReviewVerdictNeedsFixes indicates reviewer requested implementer changes.
	ReviewVerdictNeedsFixes = "NEEDS_FIXES"
)

const (
	defaultWaitTimeout  = 5 * time.Minute
	defaultPollInterval = 200 * time.Millisecond
)

// ErrWaitForClaimTimeout indicates claim polling exceeded timeout.
var ErrWaitForClaimTimeout = errors.New("timed out waiting for protocol claim")

// ProtocolEvent is the normalized persisted protocol event envelope.
//
//nolint:revive // Name kept for protocol schema clarity across packages.
type ProtocolEvent struct {
	ProtocolVersion string          `json:"protocol_version"`
	Type            string          `json:"type"`
	MissionID       string          `json:"mission_id"`
	ACID            string          `json:"ac_id,omitempty"`
	AgentID         string          `json:"agent_id,omitempty"`
	Payload         json.RawMessage `json:"payload"`
	Timestamp       time.Time       `json:"timestamp"`
}

// EventStore persists and reads protocol events for replay/audit.
type EventStore interface {
	Append(ctx context.Context, event ProtocolEvent) error
	ListByMission(ctx context.Context, missionID string) ([]ProtocolEvent, error)
}

// EventBus publishes real-time protocol events.
type EventBus interface {
	Publish(event events.Event)
}

// Service provides protocol event persistence, validation, and claim waiting.
type Service struct {
	store        EventStore
	bus          EventBus
	pollInterval time.Duration
	now          func() time.Time
}

// StuckEscalation is emitted when WaitForClaim times out.
type StuckEscalation struct {
	MissionID string        `json:"mission_id"`
	ACID      string        `json:"ac_id"`
	ClaimType string        `json:"claim_type"`
	Timeout   time.Duration `json:"timeout"`
	Timestamp time.Time     `json:"timestamp"`
}

// NewService constructs a protocol service with deterministic polling.
func NewService(store EventStore, bus EventBus, pollInterval time.Duration) (*Service, error) {
	if store == nil {
		return nil, errors.New("event store is required")
	}
	if bus == nil {
		return nil, errors.New("event bus is required")
	}
	if pollInterval <= 0 {
		pollInterval = defaultPollInterval
	}

	return &Service{
		store:        store,
		bus:          bus,
		pollInterval: pollInterval,
		now:          time.Now,
	}, nil
}

// Publish validates, persists, and emits one protocol event.
func (s *Service) Publish(ctx context.Context, event ProtocolEvent) (ProtocolEvent, error) {
	if s == nil {
		return ProtocolEvent{}, errors.New("service is nil")
	}

	normalized, err := normalizeEvent(event, s.now().UTC())
	if err != nil {
		return ProtocolEvent{}, err
	}
	if err := validateEvent(normalized); err != nil {
		return ProtocolEvent{}, err
	}
	if err := s.store.Append(ctx, normalized); err != nil {
		return ProtocolEvent{}, fmt.Errorf("persist protocol event: %w", err)
	}

	s.bus.Publish(events.Event{
		Type:       events.EventTypeProtocolEvent,
		Timestamp:  normalized.Timestamp.UTC(),
		EntityType: "mission",
		EntityID:   normalized.MissionID,
		Payload:    normalized,
		Severity:   events.SeverityInfo,
	})

	return normalized, nil
}

// WaitForClaim polls persisted events until the expected claim appears or timeout elapses.
func (s *Service) WaitForClaim(
	ctx context.Context,
	missionID string,
	acID string,
	claimType string,
	timeout time.Duration,
) (*ProtocolEvent, error) {
	if s == nil {
		return nil, errors.New("service is nil")
	}
	missionID = strings.TrimSpace(missionID)
	acID = strings.TrimSpace(acID)
	claimType = strings.TrimSpace(claimType)
	if missionID == "" {
		return nil, errors.New("mission id must not be empty")
	}
	if acID == "" {
		return nil, errors.New("acceptance criterion id must not be empty")
	}
	if claimType == "" {
		return nil, errors.New("claim type must not be empty")
	}
	if timeout <= 0 {
		timeout = defaultWaitTimeout
	}

	waitCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if match, err := s.findClaim(waitCtx, missionID, acID, claimType); err != nil {
		return nil, err
	} else if match != nil {
		return match, nil
	}

	ticker := time.NewTicker(s.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-waitCtx.Done():
			escalation := StuckEscalation{
				MissionID: missionID,
				ACID:      acID,
				ClaimType: claimType,
				Timeout:   timeout,
				Timestamp: s.now().UTC(),
			}
			s.bus.Publish(events.Event{
				Type:       events.EventTypeSystemAlert,
				Timestamp:  escalation.Timestamp,
				EntityType: "mission",
				EntityID:   missionID,
				Payload:    escalation,
				Severity:   events.SeverityError,
			})
			return nil, fmt.Errorf("%w: mission=%s ac=%s claim=%s", ErrWaitForClaimTimeout, missionID, acID, claimType)
		case <-ticker.C:
			match, err := s.findClaim(waitCtx, missionID, acID, claimType)
			if err != nil {
				return nil, err
			}
			if match != nil {
				return match, nil
			}
		}
	}
}

func (s *Service) findClaim(
	ctx context.Context,
	missionID string,
	acID string,
	claimType string,
) (*ProtocolEvent, error) {
	events, err := s.store.ListByMission(ctx, missionID)
	if err != nil {
		return nil, fmt.Errorf("list protocol events for mission %s: %w", missionID, err)
	}

	for i := len(events) - 1; i >= 0; i-- {
		event := events[i]
		if matchesClaim(event, acID, claimType) {
			return &event, nil
		}
	}
	return nil, nil
}

func normalizeEvent(event ProtocolEvent, now time.Time) (ProtocolEvent, error) {
	event.ProtocolVersion = strings.TrimSpace(event.ProtocolVersion)
	if event.ProtocolVersion == "" {
		event.ProtocolVersion = ProtocolVersion
	}
	event.Type = strings.TrimSpace(event.Type)
	event.MissionID = strings.TrimSpace(event.MissionID)
	event.ACID = strings.TrimSpace(event.ACID)
	event.AgentID = strings.TrimSpace(event.AgentID)
	if event.Timestamp.IsZero() {
		event.Timestamp = now.UTC()
	}
	if len(event.Payload) == 0 {
		event.Payload = json.RawMessage("{}")
	}
	return event, nil
}

func validateEvent(event ProtocolEvent) error {
	if event.ProtocolVersion != ProtocolVersion {
		return fmt.Errorf("unsupported protocol version %q", event.ProtocolVersion)
	}
	if !isSupportedType(event.Type) {
		return fmt.Errorf("unsupported protocol event type %q", event.Type)
	}
	if event.MissionID == "" {
		return errors.New("mission id must not be empty")
	}
	if event.Timestamp.IsZero() {
		return errors.New("timestamp must not be zero")
	}
	if !json.Valid(event.Payload) {
		return errors.New("payload must be valid JSON")
	}
	if event.Type == EventTypeAgentClaim {
		if event.ACID == "" {
			return errors.New("agent claim ac id must not be empty")
		}
		if event.AgentID == "" {
			return errors.New("agent claim agent id must not be empty")
		}
		claimType, ok := extractClaimType(event.Payload)
		if !ok {
			return errors.New("agent claim payload missing claim_type")
		}
		if !isSupportedClaimType(claimType) {
			return fmt.Errorf("unsupported claim type %q", claimType)
		}
	}
	if event.Type == EventTypeReviewComplete {
		verdict, ok := extractReviewVerdict(event.Payload)
		if !ok {
			return errors.New("review complete payload missing verdict")
		}
		if !isSupportedReviewVerdict(verdict) {
			return fmt.Errorf("unsupported review verdict %q", verdict)
		}
	}
	return nil
}

func isSupportedType(value string) bool {
	switch value {
	case EventTypeAgentClaim, EventTypeGateResult, EventTypeStateTransition, EventTypeReviewComplete:
		return true
	default:
		return false
	}
}

func isSupportedClaimType(value string) bool {
	switch strings.TrimSpace(value) {
	case ClaimTypeREDComplete, ClaimTypeGREENComplete, ClaimTypeREFACTORComplete, ClaimTypeIMPLEMENTComplete:
		return true
	default:
		return false
	}
}

func matchesClaim(event ProtocolEvent, acID string, claimType string) bool {
	if strings.TrimSpace(event.MissionID) == "" || strings.TrimSpace(event.ACID) != strings.TrimSpace(acID) {
		return false
	}

	if strings.EqualFold(strings.TrimSpace(event.Type), strings.TrimSpace(claimType)) {
		return true
	}

	if event.Type != EventTypeAgentClaim {
		return false
	}
	gotClaimType, ok := extractClaimType(event.Payload)
	if !ok {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(gotClaimType), strings.TrimSpace(claimType))
}

func extractReviewVerdict(payload json.RawMessage) (string, bool) {
	var envelope map[string]any
	if err := json.Unmarshal(payload, &envelope); err != nil {
		return "", false
	}

	for _, key := range []string{"verdict", "decision"} {
		raw, ok := envelope[key]
		if !ok {
			continue
		}
		value, ok := raw.(string)
		if !ok {
			continue
		}
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		return strings.ToUpper(value), true
	}
	return "", false
}

func isSupportedReviewVerdict(value string) bool {
	switch strings.TrimSpace(strings.ToUpper(value)) {
	case ReviewVerdictApproved, ReviewVerdictNeedsFixes:
		return true
	default:
		return false
	}
}

func extractClaimType(payload json.RawMessage) (string, bool) {
	var envelope map[string]any
	if err := json.Unmarshal(payload, &envelope); err != nil {
		return "", false
	}

	for _, key := range []string{"claim_type", "claimType", "event_type", "eventType"} {
		raw, ok := envelope[key]
		if !ok {
			continue
		}
		value, ok := raw.(string)
		if !ok {
			continue
		}
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		return value, true
	}
	return "", false
}
