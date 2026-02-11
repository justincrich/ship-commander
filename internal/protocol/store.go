package protocol

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/ship-commander/sc3/internal/beads"
)

const protocolCommentPrefix = "[sc3-protocol] "

// InMemoryStore stores protocol events in process memory.
type InMemoryStore struct {
	mu      sync.RWMutex
	events  map[string][]ProtocolEvent
	overall []ProtocolEvent
}

// NewInMemoryStore creates a memory-backed protocol event store.
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		events:  make(map[string][]ProtocolEvent),
		overall: make([]ProtocolEvent, 0),
	}
}

// Append persists one protocol event.
func (s *InMemoryStore) Append(_ context.Context, event ProtocolEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.events[event.MissionID] = append(s.events[event.MissionID], event)
	s.overall = append(s.overall, event)
	return nil
}

// ListByMission returns protocol events for one mission.
func (s *InMemoryStore) ListByMission(_ context.Context, missionID string) ([]ProtocolEvent, error) {
	missionID = strings.TrimSpace(missionID)
	if missionID == "" {
		return nil, fmt.Errorf("mission id must not be empty")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	items := s.events[missionID]
	out := make([]ProtocolEvent, len(items))
	copy(out, items)
	return out, nil
}

type beadsClient interface {
	AddComment(id, comment string) error
	Show(id string) (*beads.Bead, error)
}

// BeadsStore persists protocol events as structured comments on mission beads.
type BeadsStore struct {
	client beadsClient
}

// NewBeadsStore creates a Beads-backed protocol event store.
func NewBeadsStore(client beadsClient) (*BeadsStore, error) {
	if client == nil {
		return nil, fmt.Errorf("beads client is required")
	}
	return &BeadsStore{client: client}, nil
}

// Append persists one protocol event to Beads comments.
func (s *BeadsStore) Append(_ context.Context, event ProtocolEvent) error {
	missionID := strings.TrimSpace(event.MissionID)
	if missionID == "" {
		return fmt.Errorf("mission id must not be empty")
	}
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal protocol event: %w", err)
	}
	if err := s.client.AddComment(missionID, protocolCommentPrefix+string(body)); err != nil {
		return fmt.Errorf("persist protocol event comment: %w", err)
	}
	return nil
}

// ListByMission reads protocol events from Beads comments.
func (s *BeadsStore) ListByMission(_ context.Context, missionID string) ([]ProtocolEvent, error) {
	missionID = strings.TrimSpace(missionID)
	if missionID == "" {
		return nil, fmt.Errorf("mission id must not be empty")
	}

	bead, err := s.client.Show(missionID)
	if err != nil {
		return nil, fmt.Errorf("show mission bead: %w", err)
	}

	events := make([]ProtocolEvent, 0)
	for _, comment := range bead.Comments {
		raw := strings.TrimSpace(comment.Text)
		if !strings.HasPrefix(raw, protocolCommentPrefix) {
			continue
		}
		payload := strings.TrimPrefix(raw, protocolCommentPrefix)
		var event ProtocolEvent
		if err := json.Unmarshal([]byte(payload), &event); err != nil {
			return nil, fmt.Errorf("decode protocol comment %d: %w", comment.ID, err)
		}
		events = append(events, event)
	}
	return events, nil
}
