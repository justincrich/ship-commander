package commander

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/ship-commander/sc3/internal/gates"
	"github.com/ship-commander/sc3/internal/protocol"
)

func TestVerifyRequiresMissionIDAndWorktree(t *testing.T) {
	adapter := &GateVerifierAdapter{runner: &fakeGateRunner{}}

	if err := adapter.Verify(context.Background(), Mission{}, "/tmp/worktree"); err == nil {
		t.Fatal("expected mission id validation error")
	}
	if err := adapter.Verify(context.Background(), Mission{ID: "mission-1"}, ""); err == nil {
		t.Fatal("expected worktree validation error")
	}
}

func TestVerifyImplementRunsVerifyImplementGate(t *testing.T) {
	runner := &fakeGateRunner{result: &gates.GateResult{Classification: gates.ClassificationAccept}}
	adapter := &GateVerifierAdapter{runner: runner}

	err := adapter.VerifyImplement(context.Background(), Mission{ID: "mission-1"}, "/tmp/worktree")
	if err != nil {
		t.Fatalf("VerifyImplement() error = %v", err)
	}
	if runner.gateType != gates.GateTypeVerifyIMPLEMENT {
		t.Fatalf("gateType = %q, want %q", runner.gateType, gates.GateTypeVerifyIMPLEMENT)
	}
}

func TestVerifyRunsGreenThenRefactor(t *testing.T) {
	runner := &sequenceGateRunner{results: []*gates.GateResult{
		{Classification: gates.ClassificationAccept},
		{Classification: gates.ClassificationAccept},
	}}
	adapter := &GateVerifierAdapter{runner: runner}

	err := adapter.Verify(context.Background(), Mission{ID: "mission-2"}, "/tmp/worktree")
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if len(runner.calls) != 2 {
		t.Fatalf("calls = %d, want 2", len(runner.calls))
	}
	if runner.calls[0] != gates.GateTypeVerifyGREEN || runner.calls[1] != gates.GateTypeVerifyREFACTOR {
		t.Fatalf("calls = %v, want [VERIFY_GREEN VERIFY_REFACTOR]", runner.calls)
	}
}

func TestRunGateRejectsNonAcceptClassification(t *testing.T) {
	runner := &fakeGateRunner{result: &gates.GateResult{Classification: gates.ClassificationRejectFailure}}
	adapter := &GateVerifierAdapter{runner: runner}

	err := adapter.runGate(context.Background(), "mission-3", "/tmp/worktree", gates.GateTypeVerifyGREEN)
	if err == nil {
		t.Fatal("expected rejection error")
	}
	if !strings.Contains(err.Error(), "classification=reject_failure") {
		t.Fatalf("error = %q, want classification context", err.Error())
	}
}

func TestProtocolGateEvidenceStoreRecordGateEvidence(t *testing.T) {
	store := &fakeProtocolStore{}
	evidence := &protocolGateEvidenceStore{
		store: store,
		now: func() time.Time {
			return time.Date(2026, 2, 11, 12, 0, 0, 0, time.UTC)
		},
	}

	err := evidence.RecordGateEvidence(context.Background(), "mission-4", gates.GateResult{
		Type:           gates.GateTypeVerifyGREEN,
		Classification: gates.ClassificationAccept,
		ExitCode:       0,
	})
	if err != nil {
		t.Fatalf("RecordGateEvidence() error = %v", err)
	}
	if len(store.events) != 1 {
		t.Fatalf("events = %d, want 1", len(store.events))
	}
	event := store.events[0]
	if event.Type != protocol.EventTypeGateResult {
		t.Fatalf("event type = %q, want %q", event.Type, protocol.EventTypeGateResult)
	}
	if event.MissionID != "mission-4" {
		t.Fatalf("mission id = %q, want mission-4", event.MissionID)
	}
	var payload gates.GateResult
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if payload.Type != gates.GateTypeVerifyGREEN {
		t.Fatalf("payload type = %q, want %q", payload.Type, gates.GateTypeVerifyGREEN)
	}
}

func TestProtocolGateEvidenceStoreRequiresStore(t *testing.T) {
	evidence := &protocolGateEvidenceStore{}
	err := evidence.RecordGateEvidence(context.Background(), "mission-5", gates.GateResult{})
	if err == nil {
		t.Fatal("expected nil store validation error")
	}
}

type fakeGateRunner struct {
	result    *gates.GateResult
	err       error
	gateType  string
	workdir   string
	missionID string
}

func (f *fakeGateRunner) Run(_ context.Context, gateType string, workdir string, missionID string) (*gates.GateResult, error) {
	f.gateType = gateType
	f.workdir = workdir
	f.missionID = missionID
	if f.err != nil {
		return nil, f.err
	}
	if f.result == nil {
		return nil, errors.New("no result configured")
	}
	return f.result, nil
}

type sequenceGateRunner struct {
	results []*gates.GateResult
	calls   []string
}

func (s *sequenceGateRunner) Run(_ context.Context, gateType string, _ string, _ string) (*gates.GateResult, error) {
	s.calls = append(s.calls, gateType)
	if len(s.results) == 0 {
		return nil, errors.New("no results configured")
	}
	result := s.results[0]
	s.results = s.results[1:]
	return result, nil
}

type fakeProtocolStore struct {
	events []protocol.ProtocolEvent
}

func (f *fakeProtocolStore) Append(_ context.Context, event protocol.ProtocolEvent) error {
	f.events = append(f.events, event)
	return nil
}

func (f *fakeProtocolStore) ListByMission(_ context.Context, _ string) ([]protocol.ProtocolEvent, error) {
	return nil, nil
}
