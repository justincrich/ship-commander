package commander

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/ship-commander/sc3/internal/config"
	"github.com/ship-commander/sc3/internal/harness"
	"github.com/ship-commander/sc3/internal/protocol"
)

func TestClaudeHarnessAdapterDispatchImplementerUsesResolvedModelAndParsesClaim(t *testing.T) {
	t.Parallel()

	driver := &fakeHarnessDriver{
		session: &harness.Session{ID: "impl-1"},
		output:  `{"claim_type":"RED_COMPLETE","ac_id":"AC-1"}`,
	}
	store := protocol.NewInMemoryStore()
	cfg := &config.Config{
		DefaultHarness: "claude",
		DefaultModel:   "sonnet",
		Roles: map[string]config.RoleHarnessConfig{
			"ensign": {Harness: "claude", Model: "opus"},
		},
	}

	adapter, err := NewClaudeHarnessAdapter(driver, store, cfg, map[string]bool{"claude": true})
	if err != nil {
		t.Fatalf("new adapter: %v", err)
	}
	adapter.now = func() time.Time { return time.Unix(1700000000, 0).UTC() }

	result, err := adapter.DispatchImplementer(context.Background(), DispatchRequest{
		Mission:      Mission{ID: "MISSION-1", Title: "Do thing", Classification: MissionClassificationREDAlert},
		WorktreePath: "/tmp/worktree",
	})
	if err != nil {
		t.Fatalf("dispatch implementer: %v", err)
	}
	if result.SessionID != "impl-1" {
		t.Fatalf("session id = %q, want impl-1", result.SessionID)
	}
	if driver.lastSpawnOpts.Model != "opus" {
		t.Fatalf("model = %q, want opus", driver.lastSpawnOpts.Model)
	}

	events, err := store.ListByMission(context.Background(), "MISSION-1")
	if err != nil {
		t.Fatalf("list protocol events: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("event count = %d, want 1", len(events))
	}
	if events[0].Type != protocol.EventTypeAgentClaim {
		t.Fatalf("event type = %q, want %q", events[0].Type, protocol.EventTypeAgentClaim)
	}
}

func TestClaudeHarnessAdapterDispatchReviewerParsesVerdict(t *testing.T) {
	t.Parallel()

	driver := &fakeHarnessDriver{
		session: &harness.Session{ID: "rev-1"},
		output:  `{"decision":"NEEDS_FIXES","feedback":"retry with tests"}`,
	}
	store := protocol.NewInMemoryStore()
	cfg := &config.Config{DefaultHarness: "claude", DefaultModel: "sonnet", Roles: map[string]config.RoleHarnessConfig{}}
	adapter, err := NewClaudeHarnessAdapter(driver, store, cfg, map[string]bool{"claude": true})
	if err != nil {
		t.Fatalf("new adapter: %v", err)
	}

	_, err = adapter.DispatchReviewer(context.Background(), ReviewerDispatchRequest{
		Mission:              Mission{ID: "MISSION-2", Title: "Review me", Classification: MissionClassificationStandardOps},
		WorktreePath:         "/tmp/worktree",
		AcceptanceCriteria:   []string{"AC-1"},
		GateEvidence:         []string{"gate ok"},
		CodeDiff:             "diff --git",
		DemoTokenContent:     "mission_id: MISSION-2",
		ImplementerSessionID: "impl-2",
	})
	if err != nil {
		t.Fatalf("dispatch reviewer: %v", err)
	}

	events, err := store.ListByMission(context.Background(), "MISSION-2")
	if err != nil {
		t.Fatalf("list protocol events: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("event count = %d, want 1", len(events))
	}
	if events[0].Type != protocol.EventTypeReviewComplete {
		t.Fatalf("event type = %q, want %q", events[0].Type, protocol.EventTypeReviewComplete)
	}
	var payload map[string]string
	if err := json.Unmarshal(events[0].Payload, &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if payload["reviewer_session_id"] != "rev-1" {
		t.Fatalf("reviewer_session_id = %q, want rev-1", payload["reviewer_session_id"])
	}
}

func TestClaudeHarnessAdapterRejectsNonClaudeResolution(t *testing.T) {
	t.Parallel()

	driver := &fakeHarnessDriver{session: &harness.Session{ID: "impl-3"}}
	store := protocol.NewInMemoryStore()
	cfg := &config.Config{DefaultHarness: "codex", DefaultModel: "gpt-5-codex", Roles: map[string]config.RoleHarnessConfig{}}
	adapter, err := NewClaudeHarnessAdapter(driver, store, cfg, map[string]bool{"codex": true})
	if err != nil {
		t.Fatalf("new adapter: %v", err)
	}

	_, err = adapter.DispatchImplementer(context.Background(), DispatchRequest{
		Mission:      Mission{ID: "MISSION-3"},
		WorktreePath: "/tmp/worktree",
	})
	if err == nil {
		t.Fatal("expected harness resolution error")
	}
}

type fakeHarnessDriver struct {
	session       *harness.Session
	output        string
	lastSpawnOpts harness.SessionOpts
}

func (f *fakeHarnessDriver) SpawnSession(_ string, _ string, _ string, opts harness.SessionOpts) (*harness.Session, error) {
	f.lastSpawnOpts = opts
	return f.session, nil
}

func (f *fakeHarnessDriver) SendMessage(_ *harness.Session, _ string) (string, error) {
	return f.output, nil
}

func (f *fakeHarnessDriver) Terminate(_ *harness.Session) error {
	return nil
}
