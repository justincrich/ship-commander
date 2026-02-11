package events

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestPublishDeliversToSpecificSubscribers(t *testing.T) {
	t.Parallel()

	bus := New(WithLogger(&captureLogger{}))

	stateEvents := make(chan Event, 1)
	gateEvents := make(chan Event, 1)

	bus.Subscribe(EventTypeStateTransition, func(event Event) {
		stateEvents <- event
	})
	bus.Subscribe(EventTypeGateResult, func(event Event) {
		gateEvents <- event
	})

	bus.Publish(Event{
		Type:       EventTypeStateTransition,
		EntityType: "mission",
		EntityID:   "m-1",
		Severity:   SeverityInfo,
	})

	select {
	case got := <-stateEvents:
		if got.Type != EventTypeStateTransition {
			t.Fatalf("received type = %q, want %q", got.Type, EventTypeStateTransition)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for state subscriber event")
	}

	select {
	case got := <-gateEvents:
		t.Fatalf("unexpected gate event delivered: %#v", got)
	case <-time.After(150 * time.Millisecond):
	}
}

func TestSubscribeAllReceivesEveryEvent(t *testing.T) {
	t.Parallel()

	bus := New(WithLogger(&captureLogger{}))
	all := make(chan Event, 2)

	bus.SubscribeAll(func(event Event) {
		all <- event
	})

	bus.Publish(Event{
		Type:       EventTypeAgentSpawn,
		EntityType: "agent",
		EntityID:   "a-1",
		Severity:   SeverityInfo,
	})
	bus.Publish(Event{
		Type:       EventTypeSystemAlert,
		EntityType: "system",
		EntityID:   "s-1",
		Severity:   SeverityWarn,
	})

	gotFirst := waitForEvent(t, all)
	gotSecond := waitForEvent(t, all)
	got := []string{gotFirst.Type, gotSecond.Type}

	if !containsType(got, EventTypeAgentSpawn) {
		t.Fatalf("wildcard subscriber missing %q event; got %v", EventTypeAgentSpawn, got)
	}
	if !containsType(got, EventTypeSystemAlert) {
		t.Fatalf("wildcard subscriber missing %q event; got %v", EventTypeSystemAlert, got)
	}
}

func TestPublishDropsWhenSubscriberBufferIsFullAndReturnsQuickly(t *testing.T) {
	t.Parallel()

	logger := &captureLogger{}
	bus := New(WithBufferSize(1), WithLogger(logger))

	started := make(chan struct{}, 1)
	unblock := make(chan struct{})

	bus.Subscribe(EventTypeProtocolEvent, func(Event) {
		select {
		case started <- struct{}{}:
		default:
		}
		<-unblock
	})

	baseEvent := Event{
		Type:       EventTypeProtocolEvent,
		EntityType: "mission",
		EntityID:   "m-42",
		Severity:   SeverityWarn,
	}

	bus.Publish(baseEvent)
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for handler to block")
	}

	bus.Publish(baseEvent)

	start := time.Now()
	bus.Publish(baseEvent)
	if elapsed := time.Since(start); elapsed > 100*time.Millisecond {
		t.Fatalf("publish blocked for %s; expected non-blocking behavior", elapsed)
	}

	close(unblock)

	if !logger.contains("dropping event") {
		t.Fatalf("expected drop warning log, got %v", logger.messages())
	}
}

func TestPublishPopulatesTimestampAndPreservesMetadata(t *testing.T) {
	t.Parallel()

	bus := New(WithLogger(&captureLogger{}))
	ch := make(chan Event, 1)

	bus.Subscribe(EventTypeHealthCheck, func(event Event) {
		ch <- event
	})

	bus.Publish(Event{
		Type:       EventTypeHealthCheck,
		EntityType: "health",
		EntityID:   "hc-1",
		Payload:    map[string]any{"status": "ok"},
		Severity:   SeverityInfo,
	})

	got := waitForEvent(t, ch)
	if got.Timestamp.IsZero() {
		t.Fatal("timestamp is zero; expected publish to populate timestamp")
	}
	if got.EntityType != "health" {
		t.Fatalf("entity type = %q, want %q", got.EntityType, "health")
	}
	if got.EntityID != "hc-1" {
		t.Fatalf("entity id = %q, want %q", got.EntityID, "hc-1")
	}
	if got.Severity != SeverityInfo {
		t.Fatalf("severity = %q, want %q", got.Severity, SeverityInfo)
	}
}

func TestBusSupportsConcurrentPublishAndSubscribe(t *testing.T) {
	t.Parallel()

	bus := New(WithBufferSize(5000), WithLogger(&captureLogger{}))
	const publisherCount = 20
	const eventsPerPublisher = 100

	var received atomic.Int64
	expectedFromWildcard := int64(publisherCount * eventsPerPublisher)

	bus.SubscribeAll(func(Event) {
		received.Add(1)
	})

	var wg sync.WaitGroup
	for i := 0; i < publisherCount; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < eventsPerPublisher; j++ {
				bus.Publish(Event{
					Type:       EventTypeProtocolEvent,
					EntityType: "mission",
					EntityID:   "m-concurrent",
					Payload:    map[string]int{"publisher": i, "index": j},
					Severity:   SeverityInfo,
				})
			}
		}()
	}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			bus.Subscribe(EventTypeProtocolEvent, func(Event) {})
		}()
	}

	wg.Wait()
	waitForCount(t, &received, expectedFromWildcard, 2*time.Second)
}

func containsType(types []string, want string) bool {
	for _, eventType := range types {
		if eventType == want {
			return true
		}
	}
	return false
}

func waitForEvent(t *testing.T, ch <-chan Event) Event {
	t.Helper()

	select {
	case event := <-ch:
		return event
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for event")
		return Event{}
	}
}

func waitForCount(t *testing.T, got *atomic.Int64, want int64, timeout time.Duration) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if got.Load() >= want {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("received count = %d, want at least %d", got.Load(), want)
}

type captureLogger struct {
	mu   sync.Mutex
	logs []string
}

func (c *captureLogger) Printf(format string, args ...any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.logs = append(c.logs, fmt.Sprintf(format, args...))
}

func (c *captureLogger) contains(fragment string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, message := range c.logs {
		if strings.Contains(message, fragment) {
			return true
		}
	}
	return false
}

func (c *captureLogger) messages() []string {
	c.mu.Lock()
	defer c.mu.Unlock()

	out := make([]string, len(c.logs))
	copy(out, c.logs)
	return out
}
