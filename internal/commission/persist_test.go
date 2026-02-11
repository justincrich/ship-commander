package commission

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

type fakeRunner struct {
	output   []byte
	err      error
	lastName string
	lastArgs []string
}

func (f *fakeRunner) Run(_ context.Context, name string, args ...string) ([]byte, error) {
	f.lastName = name
	f.lastArgs = append([]string{}, args...)
	return f.output, f.err
}

func TestPersistWithRunnerCreatesBeadAndSetsID(t *testing.T) {
	t.Parallel()

	runner := &fakeRunner{output: []byte("ship-commander-3-abc\n")}
	commission := &Commission{
		Title:     "Commission Parser",
		Status:    StatusPlanning,
		CreatedAt: time.Now().UTC(),
	}

	id, err := PersistWithRunner(context.Background(), commission, runner)
	if err != nil {
		t.Fatalf("persist commission: %v", err)
	}

	if id != "ship-commander-3-abc" {
		t.Fatalf("id = %q, want %q", id, "ship-commander-3-abc")
	}
	if commission.ID != "ship-commander-3-abc" {
		t.Fatalf("commission id = %q, want %q", commission.ID, "ship-commander-3-abc")
	}
	if runner.lastName != "bd" {
		t.Fatalf("command name = %q, want %q", runner.lastName, "bd")
	}
	if !containsArgPair(runner.lastArgs, "--title", "Commission Parser") {
		t.Fatalf("missing --title argument in %v", runner.lastArgs)
	}
	if !containsArgPair(runner.lastArgs, "--silent", "") {
		t.Fatalf("missing --silent argument in %v", runner.lastArgs)
	}
}

func TestPersistWithRunnerRejectsEmptyID(t *testing.T) {
	t.Parallel()

	runner := &fakeRunner{output: []byte(" \n")}
	commission := &Commission{
		Title:     "Commission Parser",
		Status:    StatusPlanning,
		CreatedAt: time.Now().UTC(),
	}

	_, err := PersistWithRunner(context.Background(), commission, runner)
	if err == nil {
		t.Fatal("expected error for empty id output")
	}
	if !strings.Contains(err.Error(), "empty issue id") {
		t.Fatalf("error = %v, want empty issue id error", err)
	}
}

func TestCommissionJSONSerialization(t *testing.T) {
	t.Parallel()

	commission := Commission{
		ID:     "ship-commander-3-123",
		Title:  "Commission Parser",
		Status: StatusPlanning,
		UseCases: []UseCase{
			{ID: "UC-COMM-01", Title: "Parse PRD", Description: "Parse markdown"},
		},
		AcceptanceCriteria: []AC{
			{ID: "AC-001", Description: "Extract use cases", Status: "open"},
		},
		FunctionalGroups: []string{"Commission"},
		ScopeBoundaries: ScopeConfig{
			InScope:    []string{"Parser"},
			OutOfScope: []string{"UI"},
		},
		PRDContent: "## Commission",
		CreatedAt:  time.Now().UTC(),
	}

	data, err := json.Marshal(commission)
	if err != nil {
		t.Fatalf("marshal commission: %v", err)
	}
	if !strings.Contains(string(data), "\"useCases\"") {
		t.Fatalf("json output missing useCases field: %s", string(data))
	}
	if !strings.Contains(string(data), "\"acceptanceCriteria\"") {
		t.Fatalf("json output missing acceptanceCriteria field: %s", string(data))
	}
}

func TestCommissionTransitionTo(t *testing.T) {
	t.Parallel()

	commission := &Commission{Status: StatusPlanning}

	sequence := []Status{StatusApproved, StatusExecuting, StatusCompleted}
	for _, next := range sequence {
		if err := commission.TransitionTo(next); err != nil {
			t.Fatalf("transition to %q: %v", next, err)
		}
	}
	if commission.Status != StatusCompleted {
		t.Fatalf("final status = %q, want %q", commission.Status, StatusCompleted)
	}
}

func TestCommissionTransitionToRejectsIllegalTransition(t *testing.T) {
	t.Parallel()

	commission := &Commission{Status: StatusPlanning}
	if err := commission.TransitionTo(StatusCompleted); err == nil {
		t.Fatal("expected transition error")
	}
}

func containsArgPair(args []string, key, value string) bool {
	for i := 0; i < len(args); i++ {
		if args[i] != key {
			continue
		}
		if value == "" {
			return true
		}
		if i+1 < len(args) && args[i+1] == value {
			return true
		}
	}
	return false
}
