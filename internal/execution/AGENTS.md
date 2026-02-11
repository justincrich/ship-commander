# Commander Execution Orchestrator - Agent Guidance

## Purpose

This package implements the **Commander**—the mission execution authority that owns each mission from dispatch through completion or halt. The Commander runs verification gates independently, enforces termination rules, and validates demo tokens.

## Context in Ship Commander 3

The Commander is the **deterministic authority** between probabilistic agents:
- Dispatches missions in waves (dependency-aware parallel execution)
- Runs TDD cycle per AC: RED → VERIFY → GREEN → VERIFY → REFACTOR → VERIFY
- Runs STANDARD_OPS fast path: IMPLEMENT → VERIFY
- Enforces max revisions (default: 3)
- Validates demo tokens before completion
- Spawns reviewer (different ensign than implementer) for final validation

## Package Organization

```
internal/execution/
├── AGENTS.md          # This file
├── commander.go       # Main execution orchestrator
├── dispatcher.go      # Mission dispatch logic
├── missions.go        # Mission lifecycle management
├── waves.go           # Wave computation and execution
└── gates.go          # Gate runner (wraps internal/gates)
```

## Key Responsibilities

### 1. Wave Management
- Compute waves from dependency graph
- Dispatch unblocked missions in parallel
- Wait for wave completion before next wave
- Admiral wave review between waves

### 2. Mission Dispatch
- Create Git worktree per mission
- Acquire surface-area lock
- Select ensign by domain match
- Dispatch with full mission spec + context

### 3. TDD Cycle Execution
- RED phase: Dispatch ensign → Wait for RED_COMPLETE → Run VERIFY_RED
- GREEN phase: On VERIFY_RED pass → Dispatch ensign → Wait for GREEN_COMPLETE → Run VERIFY_GREEN
- REFACTOR phase: On VERIFY_GREEN pass → Dispatch ensign → Wait for REFACTOR_COMPLETE → Run VERIFY_REFACTOR
- Loop on failure until AC complete or max attempts

### 4. STANDARD_OPS Fast Path
- Single IMPLEMENT phase
- VERIFY_IMPLEMENT gate (typecheck, lint, build)
- Validate demo token

### 5. Review Orchestration
- After all ACs complete → Spawn reviewer (different ensign!)
- Reviewer receives: diff, gate evidence, acceptance criteria
- APPROVED → Mission complete
- NEEDS_FIXES → Increment revision count, re-dispatch

### 6. Termination Enforcement
- Check `revisionCount >= maxRevisions` before each dispatch
- Check demo token validity before completion
- Halt mission if termination conditions met

## Dependencies

### Internal Dependencies
- `internal/beads` - All state operations
- `internal/gates` - Verification gate execution
- `internal/harness` - Agent session management
- `internal/protocol` - Event publishing
- `internal/demo` - Demo token validation
- `internal/util` - Worktree operations

### External Dependencies
- `context` - Cancellation and timeout
- `os/exec` - Git worktree commands

## Coding Standards

Follow Go standards from `.spec/research-findings/GO_CODING_STANDARDS.md`:

### Error Handling
- Wrap all errors with context
- Never lose mission state (persist to Beads)
- Provide clear halt reasons

### Concurrency
- Use goroutines for parallel mission execution
- Use sync.WaitGroup for wave completion
- Context for wave cancellation

### State Machine
- Enforce legal state transitions
- Record all transitions to Beads
- Validate transitions before executing

## Common Patterns

### Commander Structure
```go
type Commander struct {
    beads    *beads.Client
    gates    *gates.Engine
    harness  *harness.Manager
    ensignPool *EnsignPool

    // Configuration
    maxRevisions int // Default: 3
    wipLimit    int // Default: 3
}

func (c *Commander) ExecuteCommission(ctx context.Context, commissionID string) error {
    // 1. Load commission
    commission, err := c.beads.GetCommission(ctx, commissionID)
    if err != nil {
        return fmt.Errorf("get commission: %w", err)
    }

    // 2. Compute waves
    waves := c.computeWaves(commission.Missions)

    // 3. Execute waves
    for waveNum, wave := range waves {
        c.publishWaveStart(ctx, waveNum, wave)

        // Execute wave
        if err := c.executeWave(ctx, wave); err != nil {
            return fmt.Errorf("execute wave %d: %w", waveNum, err)
        }

        // Admiral wave review
        if err := c.admiralWaveReview(ctx, waveNum); err != nil {
            return fmt.Errorf("wave review: %w", err)
        }
    }

    return nil
}
```

### Wave Execution
```go
func (c *Commander) executeWave(ctx context.Context, wave []*Mission) error {
    var wg sync.WaitGroup
    errCh := make(chan error, len(wave))

    // Dispatch missions in parallel
    for _, mission := range wave {
        wg.Add(1)
        go func(m *Mission) {
            defer wg.Done()

            if err := c.executeMission(ctx, m); err != nil {
                errCh <- fmt.Errorf("mission %s: %w", m.ID, err)
            }
        }(mission)
    }

    // Wait for all missions
    go func() {
        wg.Wait()
        close(errCh)
    }()

    // Collect errors
    var errors []error
    for err := range errCh {
        errors = append(errors, err)
    }

    if len(errors) > 0 {
        return fmt.Errorf("wave execution errors: %v", errors)
    }

    return nil
}
```

### Mission Dispatch
```go
func (c *Commander) executeMission(ctx context.Context, mission *Mission) error {
    // 1. Acquire lock
    lock, err := c.beads.AcquireLock(ctx, mission.ID, mission.SurfaceArea)
    if err != nil {
        return fmt.Errorf("acquire lock: %w", err)
    }
    defer c.beads.ReleaseLock(ctx, lock.ID)

    // 2. Create worktree
    worktree, err := util.CreateWorktree(ctx, mission.ID)
    if err != nil {
        return fmt.Errorf("create worktree: %w", err)
    }

    // 3. Select ensign by domain
    ensign, err := c.ensignPool.SelectImplementer(mission.Domain)
    if err != nil {
        return fmt.Errorf("select ensign: %w", err)
    }

    // 4. Execute based on classification
    switch mission.Classification {
    case "RED_ALERT":
        return c.executeRED_ALERT(ctx, mission, ensign, worktree)
    case "STANDARD_OPS":
        return c.executeSTANDARD_OPS(ctx, mission, ensign, worktree)
    }
}
```

### RED_ALERT TDD Cycle
```go
func (c *Commander) executeRED_ALERT(ctx context.Context, mission *Mission, ensign *Ensign, worktree string) error {
    // For each AC...
    for _, ac := range mission.AcceptanceCriteria {
        attempt := 0

        for attempt < mission.MaxAttempts {
            // RED phase
            if err := c.executePhase(ctx, ensign, worktree, ac, "RED"); err != nil {
                return err
            }

            // VERIFY_RED
            result := c.gates.ExecuteVerifyRed(ctx, worktree, ac.ID)
            if result.Classification == "accept" {
                break // Test failed as expected
            }

            if result.Classification == "reject_vanity" {
                return fmt.Errorf("vanity test detected for AC %s", ac.ID)
            }

            attempt++
        }

        // GREEN phase
        if err := c.executePhase(ctx, ensign, worktree, ac, "GREEN"); err != nil {
            return err
        }

        // VERIFY_GREEN
        result := c.gates.ExecuteVerifyGreen(ctx, worktree, ac.ID)
        if !result.Success {
            attempt++
            continue // Retry GREEN
        }

        // REFACTOR phase
        if err := c.executePhase(ctx, ensign, worktree, ac, "REFACTOR"); err != nil {
            return err
        }

        // VERIFY_REFACTOR
        result := c.gates.ExecuteVerifyRefactor(ctx, worktree, ac.ID)
        if !result.Success {
            attempt++
            continue // Retry REFACTOR
        }

        // AC complete
        break
    }

    return nil
}
```

### Review Orchestration
```go
func (c *Commander) requestReview(ctx context.Context, mission *Mission) error {
    // 1. Spawn reviewer (different ensign!)
    reviewer, err := c.ensignPool.SelectReviewer(mission.Domain, mission.ImplementerID)
    if err != nil {
        return fmt.Errorf("select reviewer: %w", err)
    }

    // 2. Prepare review context (diff, gate evidence, ACs)
    reviewCtx := ReviewContext{
        MissionID:     mission.ID,
        Diff:          c.getDiff(mission),
        GateEvidence:  c.getGateEvidence(mission),
        AcceptanceCriteria: mission.AcceptanceCriteria,
    }

    // 3. Dispatch reviewer
    verdict, err := c.dispatchReviewer(ctx, reviewer, reviewCtx)
    if err != nil {
        return fmt.Errorf("dispatch reviewer: %w", err)
    }

    // 4. Process verdict
    switch verdict {
    case "APPROVED":
        return c.beads.SetMissionState(ctx, mission.ID, "done")
    case "NEEDS_FIXES":
        mission.RevisionCount++
        if mission.RevisionCount >= mission.MaxRevisions {
            return c.beads.SetMissionState(ctx, mission.ID, "halted")
        }
        return c.reDispatchMission(ctx, mission)
    }

    return nil
}
```

### Termination Enforcement
```go
func (c *Commander) checkTerminationConditions(ctx context.Context, mission *Mission) error {
    // 1. Check max revisions
    if mission.RevisionCount >= mission.MaxRevisions {
        return fmt.Errorf("max revisions (%d) exceeded", mission.MaxRevisions)
    }

    // 2. Check demo token
    if err := c.demo.Validate(ctx, mission.ID); err != nil {
        return fmt.Errorf("demo token validation failed: %w", err)
    }

    // 3. Check AC exhaustion
    if mission.Attempts >= mission.MaxAttempts {
        return fmt.Errorf("max attempts (%d) exceeded", mission.MaxAttempts)
    }

    return nil
}
```

## Anti-Patterns to Avoid

### ❌ DON'T: Trust agent claims
```go
// BAD! Agent says "I tested it"
if agent.Claim == "TEST_PASSED" {
    return nil // NO!
}
```

### ✅ DO: Independent verification
```go
// GOOD! Run independent gate
result := gates.ExecuteVerifyGreen(ctx, worktree, acID)
if !result.Success {
    return fmt.Errorf("gate failed")
}
```

### ❌ DON'T: Implementer as reviewer
```go
// BAD! Same ensign reviews own work
reviewer := implementer // NO!
```

### ✅ DO: Different ensigns
```go
// GOOD! Context isolation
reviewer := c.ensignPool.SelectReviewer(domain, implementer.ID)
if reviewer.ID == implementer.ID {
    return errors.New("cannot review own work")
}
```

## Testing Requirements

### Unit Tests
- Test wave computation logic
- Test termination condition checks
- Test state machine transitions (with mocked Beads)

### Integration Tests
- Test full mission execution with mock harness
- Test wave parallel execution
- Test review orchestration

## References

- `.spec/prd.md` - UC-EXEC-01 through UC-EXEC-11 (11 Execution use cases)
- `.spec/technical-requirements.md` - Commander architecture (lines 273-362)
- `.spec/research-findings/GO_CODING_STANDARDS.md` - Go coding standards

---

**Version**: 1.0
**Last Updated**: 2025-02-10
