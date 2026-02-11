# Internal Packages - Agent Coordination

## Purpose

The `internal/` directory contains all private application code for Ship Commander 3. Packages in `internal/` cannot be imported by external projects (Go compiler enforcement). This document provides coordination guidance across all internal packages.

## Context in Ship Commander 3

The internal packages implement the core orchestration logic:
- **Commission**: Parse PRDs and manage commission lifecycle
- **Planning**: Ready Room orchestrator (multi-agent planning loop)
- **Execution**: Commander orchestrator (wave-based mission dispatch)
- **Gates**: Deterministic verification gate engine
- **Harness**: AI harness abstraction (Claude Code, Codex)
- **Beads**: State layer client (Git + SQLite wrapper)
- **Protocol**: Event bus and protocol events
- **Telemetry**: OpenTelemetry distributed tracing
- **Logging**: Structured JSON logging
- **TUI**: Bubble Tea terminal UI with 16 screens
- **Config**: Configuration management (TOML + Beads)
- **Doctor**: Health monitoring and stuck detection
- **Util**: Shared utilities (templates, worktree, retry)

## Package Organization Principles

### 1. Single Responsibility
Each package has ONE clear responsibility:
- `commission/` ONLY handles commission parsing and lifecycle
- `planning/` ONLY handles Ready Room orchestrator
- `execution/` ONLY handles Commander orchestrator
- etc.

### 2. Avoid Circular Dependencies
Package dependency graph MUST be acyclic:
```
beads/ (lowest level)
  ↑
protocol/ → beads/
  ↑
config/ → beads/
  ↑
planning/ → beads/ + protocol/ + config/
execution/ → beads/ + protocol/ + gates/ + harness/
  ↑
cmd/ (highest level)
```

### 3. Communication via Interfaces
- Define interfaces in consuming package
- Implement in dependency package
- Enables testing with mocks

### 4. Event Bus for Decoupling
- Use protocol event bus for cross-package communication
- Publish events for state changes
- Subscribe to events for reactive updates

## Inter-Package Dependencies

### Core Dependencies (Low-Level)
These packages have minimal dependencies and are used by many others:

#### `beads/`
- **Purpose**: Beads CLI wrapper and type definitions
- **Dependencies**: None (except os/exec for CLI calls)
- **Used By**: ALL packages

#### `protocol/`
- **Purpose**: Event bus and protocol events
- **Dependencies**: `beads/`
- **Used By**: ALL packages for state change notifications

#### `config/`
- **Purpose**: Configuration loading and resolution
- **Dependencies**: `beads/`
- **Used By**: ALL packages

#### `telemetry/`
- **Purpose**: OpenTelemetry tracing
- **Dependencies**: None (except OTEL libraries)
- **Used By**: ALL packages for observability

#### `logging/`
- **Purpose**: JSON structured logging
- **Dependencies**: `telemetry/` (for trace correlation)
- **Used By**: ALL packages

### Orchestrator Dependencies (Mid-Level)

#### `planning/`
- **Dependencies**: `beads/`, `protocol/`, `config/`, `harness/`
- **Purpose**: Ready Room orchestrator

#### `execution/`
- **Dependencies**: `beads/`, `protocol/`, `config/`, `gates/`, `harness/`
- **Purpose**: Commander orchestrator

#### `gates/`
- **Dependencies**: `beads/`, `protocol/`, `config/`
- **Purpose**: Verification gate engine

#### `harness/`
- **Dependencies**: `config/`, `telemetry/`, `logging/`
- **Purpose**: AI harness abstraction

### UI Dependencies (High-Level)

#### `tui/`
- **Dependencies**: ALL packages (via event bus subscription)
- **Purpose**: Bubble Tea TUI application

### Utility Dependencies (Shared)

#### `util/`
- **Dependencies**: None (pure utilities)
- **Purpose**: Shared helpers (templates, worktree, retry)
- **Used By**: Any package that needs these utilities

#### `demo/`
- **Dependencies**: `beads/`
- **Purpose**: Demo token validation

#### `doctor/`
- **Dependencies**: `beads/`, `protocol/`, `harness/`
- **Purpose**: Health monitoring

## State Management Patterns

### Source of Truth: Beads
- ALL persistent state stored in Beads
- No in-memory state survives restart
- Reconstruct state from Beads on startup

### State Changes via Protocol Events
1. Write state change to Beads
2. Publish protocol event to event bus
3. Subscribers react to event

### Example: Mission Dispatch
```go
// 1. Update Beads state
if err := beads.SetMissionState(missionID, "in_progress"); err != nil {
    return fmt.Errorf("set mission state: %w", err)
}

// 2. Publish event
evt := protocol.Event{
    Type:      "MISSION_DISPATCHED",
    MissionID: missionID,
    Timestamp: time.Now(),
}
if err := protocol.Publish(ctx, evt); err != nil {
    return fmt.Errorf("publish event: %w", err)
}

// 3. TUI receives event via subscription and updates display
```

## Communication Patterns

### 1. Function Calls (Synchronous)
For direct dependencies:
```go
// execution/ calls gates/
result, err := gates.Execute(ctx, gateReq)
```

### 2. Protocol Events (Asynchronous)
For loose coupling:
```go
// Publish event
protocol.Publish(ctx, protocol.Event{Type: "AGENT_CLAIM"})

// Subscribe in TUI
ch := protocol.Subscribe(filter)
for event := range ch {
    // Handle event
}
```

### 3. Context Cancellation (Control Flow)
For cancellation and timeout:
```go
ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
defer cancel()

if err := agent.Run(ctx); err != nil {
    return fmt.Errorf("agent run: %w", err)
}
```

## Coding Standards

Follow Go standards from `.spec/research-findings/GO_CODING_STANDARDS.md`:

### Package Names
- Single lowercase word: `commission`, `planning`, `execution`
- No underscores: `use_case` → `usecase`
- No plurals: `missions` → `mission`

### Exported Symbols
- Capitalized for export: `ParseCommission`, `Mission`
- Godoc comments required on all exports

### Error Handling
- Early returns with guard clauses
- Wrap errors with context: `fmt.Errorf("operation: %w", err)`
- Never ignore errors

### Context
- Always pass `context.Context` as first parameter
- Use context for cancellation, timeout, and trace propagation

### Goroutines
- Control lifecycle with context cancellation
- No fire-and-forget goroutines
- Use channels for communication

## Common Patterns

### Package Initialization
```go
// ✅ GOOD: Explicit initialization
func NewClient(cfg *config.Config) (*Client, error) {
    if cfg == nil {
        return nil, errors.New("config cannot be nil")
    }
    return &Client{cfg: cfg}, nil
}

// ❌ BAD: Package-level mutable state
var client *Client // Shared, not thread-safe!
```

### Interface Definition
```go
// ✅ GOOD: Small, focused interface
type Parser interface {
    Parse(ctx context.Context, path string) (*Commission, error)
}

// ❌ BAD: Large interface
type Manager interface {
    Parse(ctx context.Context, path string) (*Commission, error)
    Validate(*Commission) error
    Persist(*Commission) error
    // Too many methods!
}
```

### Event Publishing
```go
// ✅ GOOD: Publish after state change
func (m *Mission) Dispatch(ctx context.Context) error {
    // 1. Update Beads
    if err := m.beads.SetState(m.id, "dispatched"); err != nil {
        return err
    }

    // 2. Publish event
    return m.events.Publish(ctx, protocol.Event{
        Type:      "MISSION_DISPATCHED",
        MissionID: m.id,
    })
}
```

## Anti-Patterns to Avoid

### ❌ DON'T: Package-level mutable state
```go
var globalState map[string]string // BAD! Not thread-safe
```

### ✅ DO: Encapsulate state in structs
```go
type Service struct {
    mu    sync.Mutex
    state map[string]string
}
```

### ❌ DON'T: Circular imports
```go
// planning/ imports execution/
// execution/ imports planning/
// Circular! BAD!
```

### ✅ DO: Use protocol events
```go
// planning/ publishes events
protocol.Publish(ctx, Event{Type: "PLANNING_COMPLETE"})

// execution/ subscribes to events
ch := protocol.Subscribe(filter)
```

### ❌ DON'T: Ignore context
```go
func DoWork() error {  // No context!
    time.Sleep(10 * time.Second)  // Can't cancel!
}
```

### ✅ DO: Accept and respect context
```go
func DoWork(ctx context.Context) error {
    select {
    case <-time.After(10 * time.Second):
        return nil
    case <-ctx.Done():
        return ctx.Err()  // Cancellable!
    }
}
```

## Testing Requirements

### Unit Tests
- Table-driven tests for all logic
- Mock external dependencies (Beads, harness)
- Test all error paths
- Target >80% coverage

### Integration Tests
- Test inter-package communication
- Test event bus publish/subscribe
- Test state transitions via Beads

### Race Detector
- Run `go test -race ./...` in CI
- Fix all data races

## References

- `.spec/technical-requirements.md` - System architecture and data model
- `.spec/research-findings/GO_CODING_STANDARDS.md` - Go coding standards
- Package-specific `AGENTS.md` files in each subdirectory

---

**Version**: 1.0
**Last Updated**: 2025-02-10
