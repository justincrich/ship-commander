package protocol

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/ship-commander/sc3/internal/beads"
)

func TestInMemoryStoreAppendAndListByMission(t *testing.T) {
	t.Parallel()

	store := NewInMemoryStore()
	event := ProtocolEvent{
		ProtocolVersion: ProtocolVersion,
		Type:            EventTypeStateTransition,
		MissionID:       "mission-1",
		Payload:         json.RawMessage(`{"from":"red","to":"green"}`),
		Timestamp:       time.Now().UTC(),
	}

	if err := store.Append(context.Background(), event); err != nil {
		t.Fatalf("append: %v", err)
	}

	got, err := store.ListByMission(context.Background(), "mission-1")
	if err != nil {
		t.Fatalf("list by mission: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("event count = %d, want 1", len(got))
	}
	if got[0].Type != EventTypeStateTransition {
		t.Fatalf("event type = %q, want %q", got[0].Type, EventTypeStateTransition)
	}
}

func TestBeadsStoreAppendAndListByMission(t *testing.T) {
	t.Parallel()

	client := &fakeBeadsClient{}
	store, err := NewBeadsStore(client)
	if err != nil {
		t.Fatalf("new beads store: %v", err)
	}

	event := ProtocolEvent{
		ProtocolVersion: ProtocolVersion,
		Type:            EventTypeAgentClaim,
		MissionID:       "mission-2",
		ACID:            "AC-7",
		AgentID:         "agent-1",
		Payload:         json.RawMessage(`{"claim_type":"IMPLEMENT_COMPLETE"}`),
		Timestamp:       time.Date(2026, 2, 11, 12, 0, 0, 0, time.UTC),
	}
	if err := store.Append(context.Background(), event); err != nil {
		t.Fatalf("append beads event: %v", err)
	}

	events, err := store.ListByMission(context.Background(), "mission-2")
	if err != nil {
		t.Fatalf("list beads events: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("events count = %d, want 1", len(events))
	}
	if events[0].Type != EventTypeAgentClaim {
		t.Fatalf("event type = %q, want %q", events[0].Type, EventTypeAgentClaim)
	}
}

func TestBeadsStoreIgnoresNonProtocolComments(t *testing.T) {
	t.Parallel()

	client := &fakeBeadsClient{
		commentsByIssue: map[string][]beads.Comment{
			"mission-9": {
				{ID: 1, Text: "plain comment"},
				{ID: 2, Text: protocolCommentPrefix + `{"protocol_version":"1.0","type":"STATE_TRANSITION","mission_id":"mission-9","payload":{},"timestamp":"2026-02-11T12:00:00Z"}`},
			},
		},
	}
	store, err := NewBeadsStore(client)
	if err != nil {
		t.Fatalf("new beads store: %v", err)
	}

	events, err := store.ListByMission(context.Background(), "mission-9")
	if err != nil {
		t.Fatalf("list events: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("events count = %d, want 1", len(events))
	}
}

func TestBeadsStoreValidationAndErrors(t *testing.T) {
	t.Parallel()

	if _, err := NewBeadsStore(nil); err == nil {
		t.Fatal("expected nil client error")
	}

	client := &fakeBeadsClient{addErr: errors.New("bd unavailable")}
	store, err := NewBeadsStore(client)
	if err != nil {
		t.Fatalf("new beads store: %v", err)
	}

	appendErr := store.Append(context.Background(), ProtocolEvent{
		ProtocolVersion: ProtocolVersion,
		Type:            EventTypeStateTransition,
		MissionID:       "mission-4",
		Payload:         json.RawMessage(`{}`),
		Timestamp:       time.Now().UTC(),
	})
	if appendErr == nil {
		t.Fatal("expected append error")
	}
}

type fakeBeadsClient struct {
	addErr          error
	showErr         error
	commentsByIssue map[string][]beads.Comment
}

func (f *fakeBeadsClient) AddComment(id, comment string) error {
	if f.addErr != nil {
		return f.addErr
	}
	if f.commentsByIssue == nil {
		f.commentsByIssue = map[string][]beads.Comment{}
	}
	nextID := len(f.commentsByIssue[id]) + 1
	f.commentsByIssue[id] = append(f.commentsByIssue[id], beads.Comment{
		ID:      nextID,
		IssueID: id,
		Text:    comment,
	})
	return nil
}

func (f *fakeBeadsClient) Show(id string) (*beads.Bead, error) {
	if f.showErr != nil {
		return nil, f.showErr
	}
	return &beads.Bead{
		ID:       id,
		Comments: append([]beads.Comment(nil), f.commentsByIssue[id]...),
	}, nil
}
