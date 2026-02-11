package commander

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ship-commander/sc3/internal/gates"
	"github.com/ship-commander/sc3/internal/protocol"
)

// GateVerifierConfig configures the gates runner used by the commander verifier adapter.
type GateVerifierConfig struct {
	Timeout            time.Duration
	OutputLimitBytes   int
	OutputSnippetBytes int
	ProjectCommands    map[string][]string
	GreenInfraCommands []string
}

// GateVerifierAdapter implements Verifier using the deterministic gates runner.
type GateVerifierAdapter struct {
	runner gates.GateRunner
}

// NewGateVerifierAdapter creates a commander verifier backed by gates.NewShellRunner.
func NewGateVerifierAdapter(
	store protocol.EventStore,
	missionCommands gates.MissionCommandResolver,
	variables gates.VariableResolver,
	config GateVerifierConfig,
) (*GateVerifierAdapter, error) {
	if store == nil {
		return nil, errors.New("protocol event store is required")
	}

	evidence := &protocolGateEvidenceStore{
		store: store,
		now:   time.Now,
	}

	runner, err := gates.NewShellRunner(evidence, missionCommands, variables, gates.RunnerConfig{
		Timeout:            config.Timeout,
		OutputLimitBytes:   config.OutputLimitBytes,
		OutputSnippetBytes: config.OutputSnippetBytes,
		ProjectCommands:    config.ProjectCommands,
		GreenInfraCommands: config.GreenInfraCommands,
	})
	if err != nil {
		return nil, fmt.Errorf("create gates runner: %w", err)
	}

	return &GateVerifierAdapter{runner: runner}, nil
}

// Verify runs TDD verification gates for non-standard missions.
func (v *GateVerifierAdapter) Verify(ctx context.Context, mission Mission, worktreePath string) error {
	missionID := strings.TrimSpace(mission.ID)
	if missionID == "" {
		return errors.New("mission id must not be empty")
	}
	worktreePath = strings.TrimSpace(worktreePath)
	if worktreePath == "" {
		return errors.New("worktree path must not be empty")
	}

	if err := v.runGate(ctx, missionID, worktreePath, gates.GateTypeVerifyGREEN); err != nil {
		return err
	}
	if err := v.runGate(ctx, missionID, worktreePath, gates.GateTypeVerifyREFACTOR); err != nil {
		return err
	}
	return nil
}

// VerifyImplement runs STANDARD_OPS verification.
func (v *GateVerifierAdapter) VerifyImplement(ctx context.Context, mission Mission, worktreePath string) error {
	missionID := strings.TrimSpace(mission.ID)
	if missionID == "" {
		return errors.New("mission id must not be empty")
	}
	worktreePath = strings.TrimSpace(worktreePath)
	if worktreePath == "" {
		return errors.New("worktree path must not be empty")
	}

	return v.runGate(ctx, missionID, worktreePath, gates.GateTypeVerifyIMPLEMENT)
}

func (v *GateVerifierAdapter) runGate(ctx context.Context, missionID, worktreePath, gateType string) error {
	if v == nil || v.runner == nil {
		return errors.New("gate verifier runner is required")
	}

	result, err := v.runner.Run(ctx, gateType, worktreePath, missionID)
	if err != nil {
		return fmt.Errorf("run %s for %s: %w", gateType, missionID, err)
	}
	if result == nil {
		return fmt.Errorf("run %s for %s: empty gate result", gateType, missionID)
	}
	if strings.TrimSpace(result.Classification) != gates.ClassificationAccept {
		return fmt.Errorf("%s rejected mission %s with classification=%s", gateType, missionID, result.Classification)
	}
	return nil
}

type protocolGateEvidenceStore struct {
	store protocol.EventStore
	now   func() time.Time
}

func (s *protocolGateEvidenceStore) RecordGateEvidence(ctx context.Context, missionID string, result gates.GateResult) error {
	if s == nil || s.store == nil {
		return errors.New("protocol event store is required")
	}

	missionID = strings.TrimSpace(missionID)
	if missionID == "" {
		return errors.New("mission id must not be empty")
	}

	payload, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("marshal gate result payload: %w", err)
	}

	now := time.Now
	if s.now != nil {
		now = s.now
	}

	if err := s.store.Append(ctx, protocol.ProtocolEvent{
		ProtocolVersion: protocol.ProtocolVersion,
		Type:            protocol.EventTypeGateResult,
		MissionID:       missionID,
		Payload:         payload,
		Timestamp:       now().UTC(),
	}); err != nil {
		return fmt.Errorf("append gate result protocol event: %w", err)
	}

	return nil
}
