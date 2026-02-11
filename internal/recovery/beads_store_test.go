package recovery

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
)

type fakeCommandRunner struct {
	listPayload     []byte
	listErr         error
	setStateCalls   [][]string
	setStateCallErr error
}

func (f *fakeCommandRunner) Run(_ context.Context, _ string, args ...string) ([]byte, error) {
	if len(args) == 0 {
		return nil, errors.New("missing args")
	}
	if args[0] == "list" {
		if f.listErr != nil {
			return nil, f.listErr
		}
		return f.listPayload, nil
	}
	if args[0] == "set-state" {
		f.setStateCalls = append(f.setStateCalls, append([]string(nil), args...))
		if f.setStateCallErr != nil {
			return nil, f.setStateCallErr
		}
		return []byte(`{"ok":true}`), nil
	}
	return nil, errors.New("unexpected command")
}

func TestBeadsStoreLoadSnapshotParsesCommissionsMissionsAndAgents(t *testing.T) {
	t.Parallel()

	payload, err := json.Marshal([]map[string]any{
		{
			"id":         "comm-1",
			"issue_type": "commission",
			"status":     "open",
			"state": map[string]any{
				"commission_state": "executing",
			},
		},
		{
			"id":         "mission-1",
			"issue_type": "mission",
			"status":     "in_progress",
			"labels":     []string{"agent:agent-1"},
			"dependencies": []map[string]any{
				{"issue_id": "mission-1", "depends_on_id": "comm-1", "type": "parent-child"},
			},
			"state": map[string]any{
				"mission_state": "in_progress",
			},
		},
		{
			"id":         "agent-1",
			"issue_type": "agent",
			"status":     "open",
			"labels":     []string{"session:session-1"},
			"state": map[string]any{
				"agent_state": "running",
			},
		},
	})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	runner := &fakeCommandRunner{listPayload: payload}
	store, err := NewBeadsStoreWithRunner(runner)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}

	snapshot, err := store.LoadSnapshot(context.Background())
	if err != nil {
		t.Fatalf("load snapshot: %v", err)
	}

	wantCommissions := []Commission{{ID: "comm-1", State: "executing"}}
	if !reflect.DeepEqual(snapshot.Commissions, wantCommissions) {
		t.Fatalf("commissions = %#v, want %#v", snapshot.Commissions, wantCommissions)
	}

	wantMissions := []Mission{{
		ID:           "mission-1",
		CommissionID: "comm-1",
		State:        "in_progress",
		AgentID:      "agent-1",
	}}
	if !reflect.DeepEqual(snapshot.Missions, wantMissions) {
		t.Fatalf("missions = %#v, want %#v", snapshot.Missions, wantMissions)
	}

	wantAgents := []Agent{{
		ID:        "agent-1",
		State:     "running",
		SessionID: "session-1",
	}}
	if !reflect.DeepEqual(snapshot.Agents, wantAgents) {
		t.Fatalf("agents = %#v, want %#v", snapshot.Agents, wantAgents)
	}
}

func TestBeadsStoreSetStateOperations(t *testing.T) {
	t.Parallel()

	runner := &fakeCommandRunner{listPayload: []byte(`[]`)}
	store, err := NewBeadsStoreWithRunner(runner)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}

	if err := store.SetMissionBacklog(context.Background(), "mission-1"); err != nil {
		t.Fatalf("set mission backlog: %v", err)
	}
	if err := store.SetAgentDead(context.Background(), "agent-1"); err != nil {
		t.Fatalf("set agent dead: %v", err)
	}

	wantCalls := [][]string{
		{"set-state", "mission-1", "mission_state=backlog", "--json"},
		{"set-state", "agent-1", "agent_state=dead", "--json"},
	}
	if !reflect.DeepEqual(runner.setStateCalls, wantCalls) {
		t.Fatalf("set-state calls = %#v, want %#v", runner.setStateCalls, wantCalls)
	}
}

func TestNewBeadsStoreWithRunnerRejectsNil(t *testing.T) {
	t.Parallel()

	if _, err := NewBeadsStoreWithRunner(nil); err == nil {
		t.Fatal("expected nil runner error")
	}
}
