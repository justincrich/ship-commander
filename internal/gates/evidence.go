package gates

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type stateSetter interface {
	SetState(id, key, value string) error
}

// BeadsEvidenceStore persists gate evidence using Beads issue state keys.
type BeadsEvidenceStore struct {
	states stateSetter
}

// NewBeadsEvidenceStore creates an EvidenceStore backed by Beads state storage.
func NewBeadsEvidenceStore(states stateSetter) (*BeadsEvidenceStore, error) {
	if states == nil {
		return nil, fmt.Errorf("state setter is required")
	}
	return &BeadsEvidenceStore{states: states}, nil
}

// RecordGateEvidence persists one gate result into Beads issue state.
func (s *BeadsEvidenceStore) RecordGateEvidence(_ context.Context, missionID string, result GateResult) error {
	missionID = strings.TrimSpace(missionID)
	if missionID == "" {
		return fmt.Errorf("mission id must not be empty")
	}
	gateType := strings.TrimSpace(strings.ToLower(result.Type))
	if gateType == "" {
		return fmt.Errorf("gate type must not be empty")
	}
	if result.Attempt <= 0 {
		return fmt.Errorf("attempt must be positive")
	}

	base := fmt.Sprintf("gates.%s.attempt_%d", gateType, result.Attempt)
	updates := map[string]string{
		base + ".exit_code":      strconv.Itoa(result.ExitCode),
		base + ".classification": strings.TrimSpace(result.Classification),
		base + ".output_snippet": strings.TrimSpace(result.OutputSnippet),
		base + ".duration_ms":    strconv.FormatInt(result.Duration.Milliseconds(), 10),
		base + ".timestamp":      result.Timestamp.UTC().Format(time.RFC3339Nano),
	}

	for key, value := range updates {
		if err := s.states.SetState(missionID, key, value); err != nil {
			return fmt.Errorf("set state %q: %w", key, err)
		}
	}
	return nil
}
