package admiral

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestApprovalGateAwaitDecisionBlocksAndPersistsHistory(t *testing.T) {
	t.Parallel()

	gate := NewApprovalGate(1)
	request := ApprovalRequest{
		CommissionID: "commission-1",
		MissionManifest: []Mission{
			{
				ID:         "M-1",
				Title:      "Bootstrap runtime",
				UseCaseIDs: []string{"UC-1"},
			},
		},
		WaveAssignments: []Wave{
			{
				Index:      1,
				MissionIDs: []string{"M-1"},
			},
		},
		CoverageMap:   map[string]CoverageStatus{"UC-1": CoverageStatusCovered},
		Iteration:     1,
		MaxIterations: 3,
	}

	done := make(chan struct{})
	var response ApprovalResponse
	var err error
	go func() {
		defer close(done)
		response, err = gate.AwaitDecision(context.Background(), request)
	}()

	select {
	case surfaced := <-gate.Requests():
		if surfaced.CommissionID != "commission-1" {
			t.Fatalf("commission id = %q, want commission-1", surfaced.CommissionID)
		}
		if len(surfaced.MissionManifest) != 1 {
			t.Fatalf("mission manifest length = %d, want 1", len(surfaced.MissionManifest))
		}
		if len(surfaced.WaveAssignments) != 1 || surfaced.WaveAssignments[0].Index != 1 {
			t.Fatalf("wave assignments = %+v, want one wave with index 1", surfaced.WaveAssignments)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for approval request")
	}

	if err := gate.Respond(ApprovalResponse{Decision: ApprovalDecisionApproved}); err != nil {
		t.Fatalf("respond: %v", err)
	}

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for await decision")
	}

	if err != nil {
		t.Fatalf("await decision: %v", err)
	}
	if response.Decision != ApprovalDecisionApproved {
		t.Fatalf("decision = %q, want %q", response.Decision, ApprovalDecisionApproved)
	}

	history := gate.History()
	if len(history) != 1 {
		t.Fatalf("history entries = %d, want 1", len(history))
	}
	if history[0].Request.CommissionID != "commission-1" {
		t.Fatalf("history commission id = %q, want commission-1", history[0].Request.CommissionID)
	}
}

func TestApprovalGateResponseValidation(t *testing.T) {
	t.Parallel()

	gate := NewApprovalGate(1)

	if err := gate.Respond(ApprovalResponse{Decision: "invalid"}); err == nil {
		t.Fatal("expected invalid decision error, got nil")
	}

	if err := gate.Respond(ApprovalResponse{Decision: ApprovalDecisionFeedback}); err == nil {
		t.Fatal("expected missing feedback text error, got nil")
	}

	if err := gate.Respond(ApprovalResponse{
		Decision:     ApprovalDecisionFeedback,
		FeedbackText: "reconvene with mission split",
	}); err != nil {
		t.Fatalf("respond feedback: %v", err)
	}
}

func TestApprovalGateRejectsInvalidRequest(t *testing.T) {
	t.Parallel()

	gate := NewApprovalGate(1)
	_, err := gate.AwaitDecision(context.Background(), ApprovalRequest{})
	if err == nil {
		t.Fatal("expected invalid request error, got nil")
	}
	if !errors.Is(err, context.Canceled) && err.Error() != "commission id is required" {
		t.Fatalf("unexpected error: %v", err)
	}
}
