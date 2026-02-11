package protocol

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ship-commander/sc3/internal/events"
)

func TestPublishValidatesPersistsAndEmits(t *testing.T) {
	t.Parallel()

	store := NewInMemoryStore()
	bus := &fakeBus{}
	service, err := NewService(store, bus, 10*time.Millisecond)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	published, err := service.Publish(context.Background(), ProtocolEvent{
		Type:      EventTypeAgentClaim,
		MissionID: "mission-1",
		ACID:      "AC-1",
		AgentID:   "agent-77",
		Payload:   json.RawMessage(`{"claim_type":"RED_COMPLETE"}`),
	})
	if err != nil {
		t.Fatalf("publish protocol event: %v", err)
	}

	if published.ProtocolVersion != ProtocolVersion {
		t.Fatalf("protocol version = %q, want %q", published.ProtocolVersion, ProtocolVersion)
	}
	if published.Timestamp.IsZero() {
		t.Fatal("timestamp should be populated")
	}

	persisted, err := store.ListByMission(context.Background(), "mission-1")
	if err != nil {
		t.Fatalf("list mission events: %v", err)
	}
	if len(persisted) != 1 {
		t.Fatalf("persisted events = %d, want 1", len(persisted))
	}

	if bus.Count() != 1 {
		t.Fatalf("bus publish count = %d, want 1", bus.Count())
	}
	event := bus.Last()
	if event.Type != events.EventTypeProtocolEvent {
		t.Fatalf("bus event type = %q, want %q", event.Type, events.EventTypeProtocolEvent)
	}
}

func TestPublishRejectsInvalidSchema(t *testing.T) {
	t.Parallel()

	store := NewInMemoryStore()
	bus := &fakeBus{}
	service, err := NewService(store, bus, 10*time.Millisecond)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	_, err = service.Publish(context.Background(), ProtocolEvent{
		ProtocolVersion: "2.0",
		Type:            EventTypeStateTransition,
		MissionID:       "mission-1",
		Payload:         json.RawMessage(`{}`),
	})
	if err == nil {
		t.Fatal("expected schema validation error")
	}
	if !strings.Contains(err.Error(), "unsupported protocol version") {
		t.Fatalf("error = %v, want version validation", err)
	}

	_, err = service.Publish(context.Background(), ProtocolEvent{
		Type:      EventTypeAgentClaim,
		MissionID: "mission-1",
		ACID:      "AC-1",
		AgentID:   "agent-1",
		Payload:   json.RawMessage(`{"claim_type":"UNKNOWN"}`),
	})
	if err == nil {
		t.Fatal("expected unsupported claim type error")
	}
}

func TestWaitForClaimFindsPersistedClaim(t *testing.T) {
	t.Parallel()

	store := NewInMemoryStore()
	bus := &fakeBus{}
	service, err := NewService(store, bus, 5*time.Millisecond)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	publishErrCh := make(chan error, 1)
	go func() {
		time.Sleep(40 * time.Millisecond)
		_, err := service.Publish(context.Background(), ProtocolEvent{
			Type:      EventTypeAgentClaim,
			MissionID: "mission-2",
			ACID:      "AC-2",
			AgentID:   "agent-99",
			Payload:   json.RawMessage(`{"claim_type":"GREEN_COMPLETE"}`),
		})
		publishErrCh <- err
	}()

	found, err := service.WaitForClaim(context.Background(), "mission-2", "AC-2", ClaimTypeGREENComplete, 500*time.Millisecond)
	if err != nil {
		t.Fatalf("wait for claim: %v", err)
	}
	if found == nil {
		t.Fatal("expected claim event")
	}
	if found.Type != EventTypeAgentClaim {
		t.Fatalf("claim event type = %q, want %q", found.Type, EventTypeAgentClaim)
	}
	select {
	case publishErr := <-publishErrCh:
		if publishErr != nil {
			t.Fatalf("publish protocol event: %v", publishErr)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for publish goroutine")
	}
}

func TestWaitForClaimTimeoutPublishesEscalation(t *testing.T) {
	t.Parallel()

	store := NewInMemoryStore()
	bus := &fakeBus{}
	service, err := NewService(store, bus, 5*time.Millisecond)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	_, err = service.WaitForClaim(context.Background(), "mission-3", "AC-9", ClaimTypeREFACTORComplete, 60*time.Millisecond)
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if !errors.Is(err, ErrWaitForClaimTimeout) {
		t.Fatalf("error = %v, want ErrWaitForClaimTimeout", err)
	}

	alerts := bus.EventsByType(events.EventTypeSystemAlert)
	if len(alerts) != 1 {
		t.Fatalf("system alert count = %d, want 1", len(alerts))
	}
}

func TestWaitForClaimInputValidation(t *testing.T) {
	t.Parallel()

	store := NewInMemoryStore()
	bus := &fakeBus{}
	service, err := NewService(store, bus, 5*time.Millisecond)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	_, err = service.WaitForClaim(context.Background(), "", "AC-1", ClaimTypeREDComplete, time.Second)
	if err == nil {
		t.Fatal("expected mission id validation error")
	}

	_, err = service.WaitForClaim(context.Background(), "mission-1", "", ClaimTypeREDComplete, time.Second)
	if err == nil {
		t.Fatal("expected AC id validation error")
	}

	_, err = service.WaitForClaim(context.Background(), "mission-1", "AC-1", "", time.Second)
	if err == nil {
		t.Fatal("expected claim type validation error")
	}
}

type fakeBus struct {
	mu     sync.Mutex
	events []events.Event
}

func (f *fakeBus) Publish(event events.Event) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.events = append(f.events, event)
}

func (f *fakeBus) Count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.events)
}

func (f *fakeBus) Last() events.Event {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.events) == 0 {
		return events.Event{}
	}
	return f.events[len(f.events)-1]
}

func (f *fakeBus) EventsByType(eventType string) []events.Event {
	f.mu.Lock()
	defer f.mu.Unlock()

	out := make([]events.Event, 0)
	for _, event := range f.events {
		if event.Type == eventType {
			out = append(out, event)
		}
	}
	return out
}
