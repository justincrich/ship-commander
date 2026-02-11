# Beads State Layer Client - Agent Guidance

## Purpose

This package provides a Go wrapper around the `bd` CLI for Beads (Git + SQLite) state management. It handles all persistent state operations for Ship Commander 3.

## Context in Ship Commander 3

Beads is the **source of truth** for all state in Ship Commander 3:
- Commissions are parent beads
- Missions are child beads of commissions
- Agents are beads (crew members)
- Protocol events are beads (audit trail)
- Locks are managed via Beads

The Beads package provides the interface between Go code and the `bd` CLI.

## Package Organization

```
internal/beads/
├── AGENTS.md              # This file
├── client.go              # Beads CLI wrapper and HTTP client
├── types.go               # Bead type definitions (Commission, Mission, Agent, etc.)
├── commission.go           # Commission bead operations
├── mission.go              # Mission bead operations
├── agent.go               # Agent bead operations
├── protocol.go            # Protocol event operations
├── locks.go               # Lock management
└── state_machine.go       # State transition enforcement
```

## Key Responsibilities

### 1. Beads CLI Wrapper
- Execute `bd` commands via `os/exec`
- Parse JSON output from `--json` flag
- Handle errors and exit codes
- Provide Go-native API

### 2. Type Definitions
Define Go structs for all Bead types:
```go
type Bead struct {
    ID       string            `json:"id"`
    Type     string            `json:"type"`
    Title    string            `json:"title"`
    Body     string            `json:"body"`
    State    map[string]string `json:"state"`
    Labels   map[string]string `json:"labels"`
    Parent   string            `json:"parent,omitempty"`
}

type Commission Bead
type Mission Bead
type Agent Bead
type ProtocolEvent Bead
```

### 3. CRUD Operations
- Create beads: `bd create --type=<type> --title=<title>`
- Read beads: `bd show <id> --json`
- Update beads: `bd set-state <id> <key>=<value>`
- Delete beads: `bd delete <id>` (rare, use state changes instead)

### 4. State Machine Enforcement
- Validate legal state transitions
- Reject illegal transitions with clear errors
- Record all transitions in audit log

### 5. Lock Management
- Acquire locks with surface area declarations
- Check for conflicts (glob pattern matching)
- Release locks on mission completion
- Handle lock expiry and renewal

## Dependencies

### External Dependencies
- None (only `os/exec` for CLI calls)

### Internal Dependencies
- None (this is a leaf package)

**Used By**: All other packages

## Coding Standards

Follow Go standards from `.spec/research-findings/GO_CODING_STANDARDS.md`:

### Error Handling
- Wrap all `bd` command errors with context
- Distinguish between CLI errors and Beads errors
- Parse error output from `bd` commands

### Command Execution
```go
// ✅ GOOD: Command execution with timeout
func (c *Client) runCommand(ctx context.Context, args ...string) (string, error) {
    cmd := exec.CommandContext(ctx, "bd", args...)
    cmd.Args = append([]string{"bd", "--json"}, args...)

    output, err := cmd.CombinedOutput()
    if err != nil {
        return "", fmt.Errorf("bd %v: %w (output: %s)", strings.Join(args, " "), err, string(output))
    }

    return string(output), nil
}
```

### JSON Parsing
```go
// ✅ GOOD: Parse JSON output
func (c *Client) GetBead(ctx context.Context, id string) (*Bead, error) {
    output, err := c.runCommand(ctx, "show", id, "--json")
    if err != nil {
        return nil, err
    }

    var bead Bead
    if err := json.Unmarshal([]byte(output), &bead); err != nil {
        return nil, fmt.Errorf("parse bead JSON: %w", err)
    }

    return &bead, nil
}
```

## Common Patterns

### Client Initialization
```go
type Client struct {
    basePath string   // Path to .beads/ directory
    timeout  time.Duration
}

func NewClient(basePath string) (*Client, error) {
    if err := os.MkdirAll(basePath, 0755); err != nil {
        return nil, fmt.Errorf("create beads dir: %w", err)
    }

    // Initialize Beads if not present
    if _, err := os.Stat(filepath.Join(basePath, "db.sqlite")); os.IsNotExist(err) {
        if err := exec.Command("bd", "init", "-C", basePath).Run(); err != nil {
            return nil, fmt.Errorf("bd init: %w", err)
        }
    }

    return &Client{
        basePath: basePath,
        timeout:  30 * time.Second,
    }, nil
}
```

### Commission Operations
```go
// Create commission from PRD
func (c *Client) CreateCommission(ctx context.Context, prd *PRD) (*Commission, error) {
    output, err := c.runCommand(ctx, "create",
        "--type=commission",
        "--title", prd.Title,
        "--body", prd.Content,
    )
    if err != nil {
        return nil, err
    }

    var commission Commission
    if err := json.Unmarshal([]byte(output), &commission); err != nil {
        return nil, fmt.Errorf("parse commission: %w", err)
    }

    return &commission, nil
}

// Update commission state
func (c *Client) SetCommissionState(ctx context.Context, id, state string) error {
    _, err := c.runCommand(ctx, "set-state", id, "status="+state)
    return err
}
```

### Mission Operations
```go
// Create mission as child of commission
func (c *Client) CreateMission(ctx context.Context, commissionID string, spec *MissionSpec) (*Mission, error) {
    output, err := c.runCommand(ctx, "create",
        "--type=mission",
        "--parent", commissionID,
        "--title", spec.Title,
        "--body", spec.Body,
    )
    if err != nil {
        return nil, err
    }

    var mission Mission
    if err := json.Unmarshal([]byte(output), &mission); err != nil {
        return nil, fmt.Errorf("parse mission: %w", err)
    }

    return &mission, nil
}

// Add mission dependency
func (c *Client) AddDependency(ctx context.Context, missionID, dependsOnID string) error {
    _, err := c.runCommand(ctx, "dep", "add", missionID, dependsOnID)
    return err
}

// Query unblocked missions (ready for dispatch)
func (c *Client) GetReadyMissions(ctx context.Context, commissionID string) ([]*Mission, error) {
    output, err := c.runCommand(ctx, "ready", "--commission", commissionID, "--json")
    if err != nil {
        return nil, err
    }

    var missions []*Mission
    if err := json.Unmarshal([]byte(output), &missions); err != nil {
        return nil, fmt.Errorf("parse missions: %w", err)
    }

    return missions, nil
}
```

### Lock Operations
```go
type Lock struct {
    ID         string
    MissionID  string
    SurfaceArea []string  // Glob patterns
    AcquiredAt  time.Time
    ExpiresAt   time.Time
}

// Acquire lock
func (c *Client) AcquireLock(ctx context.Context, missionID string, surfaceArea []string) (*Lock, error) {
    // Check for conflicts
    locks, err := c.GetActiveLocks(ctx)
    if err != nil {
        return nil, err
    }

    if conflicts := checkConflicts(surfaceArea, locks); len(conflicts) > 0 {
        return nil, fmt.Errorf("lock conflict with missions: %v", conflicts)
    }

    // Create lock bead
    lock := &Lock{
        ID:         generateLockID(),
        MissionID:  missionID,
        SurfaceArea: surfaceArea,
        AcquiredAt:  time.Now(),
        ExpiresAt:   time.Now().Add(2 * time.Hour), // Default TTL
    }

    // Persist to Beads
    if err := c.persistLock(ctx, lock); err != nil {
        return nil, err
    }

    return lock, nil
}
```

## State Machine Enforcement

```go
// Legal commission state transitions
var commissionStateTransitions = map[string][]string{
    "planning":   {"approved", "shelved"},
    "approved":   {"executing", "shelved"},
    "executing":  {"completed", "shelved"},
    "completed":  {}, // Terminal
    "shelved":    {"planning", "approved"}, // Can resume
}

// Validate state transition
func (c *Client) ValidateTransition(entityType, fromState, toState string) error {
    var transitions map[string][]string

    switch entityType {
    case "commission":
        transitions = commissionStateTransitions
    case "mission":
        transitions = missionStateTransitions
    default:
        return fmt.Errorf("unknown entity type: %s", entityType)
    }

    validStates, ok := transitions[fromState]
    if !ok {
        return fmt.Errorf("invalid from state: %s", fromState)
    }

    for _, valid := range validStates {
        if valid == toState {
            return nil // Valid transition
        }
    }

    return fmt.Errorf("invalid transition: %s → %s", fromState, toState)
}
```

## Anti-Patterns to Avoid

### ❌ DON'T: Bypass state machine
```go
// Direct state update without validation
bead.State["status"] = "done" // BAD!
```

### ✅ DO: Use state machine
```go
// Validate and transition
if err := client.SetMissionState(ctx, id, "done"); err != nil {
    return fmt.Errorf("set state: %w", err)
}
```

### ❌ DON'T: Ignore errors
```go
client.CreateCommission(ctx, prd) // BAD! Ignoring error
```

### ✅ DO: Always handle errors
```go
commission, err := client.CreateCommission(ctx, prd)
if err != nil {
    return fmt.Errorf("create commission: %w", err)
}
```

## Testing Requirements

### Unit Tests
- Mock `bd` CLI execution
- Test JSON parsing
- Test state machine validation
- Test lock conflict detection

### Integration Tests
- Test with real `bd` CLI
- Test Beads initialization
- Test CRUD operations

## References

- `.spec/technical-requirements.md` - Beads data model (lines 36-365)
- [Beads Documentation](https://github.com/steveyegge/beads) - State layer
- `.spec/research-findings/GO_CODING_STANDARDS.md` - Go coding standards

---

**Version**: 1.0
**Last Updated**: 2025-02-10
