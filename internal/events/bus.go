package events

import (
	"log"
	"strings"
	"sync"
	"time"
)

const (
	// DefaultBufferSize is the default per-subscriber channel capacity.
	DefaultBufferSize = 100

	// EventTypeStateTransition identifies state transition events.
	EventTypeStateTransition = "StateTransition"
	// EventTypeGateResult identifies gate execution result events.
	EventTypeGateResult = "GateResult"
	// EventTypeProtocolEvent identifies protocol-layer events.
	EventTypeProtocolEvent = "ProtocolEvent"
	// EventTypeAgentSpawn identifies agent spawn lifecycle events.
	EventTypeAgentSpawn = "AgentSpawn"
	// EventTypeAgentExit identifies agent exit lifecycle events.
	EventTypeAgentExit = "AgentExit"
	// EventTypeHealthCheck identifies system health check events.
	EventTypeHealthCheck = "HealthCheck"
	// EventTypeAdmiralQuestion identifies question events surfaced to Admiral.
	EventTypeAdmiralQuestion = "AdmiralQuestion"
	// EventTypeSystemAlert identifies high-severity system alert events.
	EventTypeSystemAlert = "SystemAlert"
)

const (
	// SeverityInfo indicates informational event severity.
	SeverityInfo = "INFO"
	// SeverityWarn indicates warning event severity.
	SeverityWarn = "WARN"
	// SeverityError indicates error event severity.
	SeverityError = "ERROR"
)

// Event is the normalized message delivered through the in-process event bus.
type Event struct {
	Type       string
	Timestamp  time.Time
	EntityType string
	EntityID   string
	Payload    any
	Severity   string
}

// Handler consumes a published event.
type Handler func(Event)

// Logger captures warning logs for dropped events.
type Logger interface {
	Printf(format string, args ...any)
}

// Bus defines event subscription and publish behavior.
type Bus interface {
	Subscribe(eventType string, handler Handler)
	SubscribeAll(handler Handler)
	Publish(event Event)
}

// Option customizes bus construction.
type Option func(*InMemoryBus)

// WithBufferSize configures per-subscriber channel capacity.
func WithBufferSize(size int) Option {
	return func(bus *InMemoryBus) {
		if size > 0 {
			bus.bufferSize = size
		}
	}
}

// WithLogger configures log sink used for dropped-event warnings.
func WithLogger(logger Logger) Option {
	return func(bus *InMemoryBus) {
		if logger != nil {
			bus.logger = logger
		}
	}
}

// InMemoryBus is a thread-safe in-process pub/sub bus backed by buffered channels.
type InMemoryBus struct {
	mu             sync.RWMutex
	bufferSize     int
	logger         Logger
	typedSubs      map[string][]*subscriber
	wildcardSubs   []*subscriber
	nextSubscriber uint64
}

type subscriber struct {
	id uint64
	ch chan Event
}

// New creates an in-memory event bus with optional configuration.
func New(options ...Option) *InMemoryBus {
	bus := &InMemoryBus{
		bufferSize:   DefaultBufferSize,
		logger:       log.Default(),
		typedSubs:    make(map[string][]*subscriber),
		wildcardSubs: make([]*subscriber, 0),
	}
	for _, option := range options {
		option(bus)
	}
	return bus
}

// Subscribe registers a handler for a specific event type.
func (b *InMemoryBus) Subscribe(eventType string, handler Handler) {
	normalizedType := strings.TrimSpace(eventType)
	if normalizedType == "" || handler == nil {
		return
	}
	sub := b.newSubscriber()

	b.mu.Lock()
	b.typedSubs[normalizedType] = append(b.typedSubs[normalizedType], sub)
	b.mu.Unlock()

	go b.consume(sub, handler)
}

// SubscribeAll registers a handler that receives every published event.
func (b *InMemoryBus) SubscribeAll(handler Handler) {
	if handler == nil {
		return
	}
	sub := b.newSubscriber()

	b.mu.Lock()
	b.wildcardSubs = append(b.wildcardSubs, sub)
	b.mu.Unlock()

	go b.consume(sub, handler)
}

// Publish delivers an event to typed subscribers and wildcard subscribers.
func (b *InMemoryBus) Publish(event Event) {
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}

	typed, wildcard := b.snapshotSubscribers(strings.TrimSpace(event.Type))
	for _, sub := range typed {
		b.deliver(sub, event)
	}
	for _, sub := range wildcard {
		b.deliver(sub, event)
	}
}

func (b *InMemoryBus) snapshotSubscribers(eventType string) ([]*subscriber, []*subscriber) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	typed := make([]*subscriber, len(b.typedSubs[eventType]))
	copy(typed, b.typedSubs[eventType])

	wildcard := make([]*subscriber, len(b.wildcardSubs))
	copy(wildcard, b.wildcardSubs)

	return typed, wildcard
}

func (b *InMemoryBus) deliver(sub *subscriber, event Event) {
	select {
	case sub.ch <- event:
	default:
		b.logger.Printf(
			"events: dropping event for subscriber=%d type=%s entity_type=%s entity_id=%s",
			sub.id,
			event.Type,
			event.EntityType,
			event.EntityID,
		)
	}
}

func (b *InMemoryBus) newSubscriber() *subscriber {
	b.mu.Lock()
	b.nextSubscriber++
	id := b.nextSubscriber
	b.mu.Unlock()

	return &subscriber{
		id: id,
		ch: make(chan Event, b.bufferSize),
	}
}

func (b *InMemoryBus) consume(sub *subscriber, handler Handler) {
	for event := range sub.ch {
		handler(event)
	}
}
