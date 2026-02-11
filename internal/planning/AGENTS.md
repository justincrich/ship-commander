# Ready Room Planning Orchestrator - Agent Guidance

## Purpose

This package implements the **Ready Room** orchestrator—a multi-agent planning loop where Captain, Commander, and Design Officer collaboratively decompose a commission into missions before any code is written.

## Context in Ship Commander 3

The Ready Room is a **one-time planning event** that produces a complete mission manifest:

1. **Spawn Sessions**: Create 3 separate harness sessions (Captain, Commander, Design Officer)
2. **Analyze Commission**: Each agent analyzes commission from their perspective
3. **Decompose Use Cases**: Collaboratively break down use cases into missions
4. **Route Messages**: Pass structured messages between agent sessions
5. **Surface Questions**: Agents can ask Admiral questions mid-planning
6. **Validate Consensus**: Ensure three-way sign-off (Captain, Commander, Design Officer)
7. **Produce Manifest**: Sequenced mission list with dependencies
8. **Admiral Approval**: Human approves full manifest before execution

## Package Organization

```
internal/planning/
├── AGENTS.md          # This file
├── readyroom.go       # Main planning orchestrator
├── message.go         # Inter-agent message routing
├── consensus.go        # Consensus validation
└── questions.go       # Agent → Admiral question handling
```

## Key Responsibilities

### 1. Session Management
- Spawn 3 separate harness sessions via `internal/harness`
- Each session has isolated context window
- Terminate sessions after planning complete

### 2. Message Routing
- Route structured messages between agent sessions
- Messages arrive as additional context (not full context merge)
- Orchestrator controls information flow

### 3. Consensus Validation
- All missions must have three-way sign-off
- All commission use cases must be covered
- Planning loop iterates until consensus or max iterations

### 4. Admiral Questions
- Any agent can surface questions to Admiral
- Planning loop suspends until answer received
- Answer routed back to asking agent

## Dependencies

### Internal Dependencies
- `internal/harness` - Spawn agent sessions
- `internal/beads` - Persist planning state and messages
- `internal/protocol` - Publish planning events
- `internal/config` - Role-to-model configuration

### External Dependencies
- Context for cancellation and timeout

## Coding Standards

Follow Go standards from `.spec/research-findings/GO_CODING_STANDARDS.md`:

### Error Handling
- Wrap harness errors with context
- Provide clear feedback on planning failures
- Never lose planning state (persist to Beads)

### Concurrency
- Use goroutines for parallel agent sessions
- Use channels for message passing
- Context for cancellation and timeout

### Session Isolation
- Each agent in separate harness session
- No shared memory between sessions
- Communication via structured messages only

## Common Patterns

### Ready Room Orchestrator
```go
type ReadyRoom struct {
    beads       *beads.Client
    harness     *harness.Manager
    eventBus    *protocol.Bus

    // Agent sessions
    captain     *harness.Session
    commander    *harness.Session
    designOff    *harness.Session

    // Planning state
    commission  *Commission
    missions    []*Mission
    messages    []*ReadyRoomMessage
    iteration   int
    maxIter     int // Default: 5
}

func (rr *ReadyRoom) Plan(ctx context.Context, prdPath string) (*MissionManifest, error) {
    // 1. Parse commission
    commission, err := commission.Parse(ctx, prdPath)
    if err != nil {
        return nil, fmt.Errorf("parse commission: %w", err)
    }

    // 2. Persist to Beads
    if err := rr.beads.CreateCommission(ctx, commission); err != nil {
        return nil, fmt.Errorf("persist commission: %w", err)
    }

    // 3. Spawn agent sessions
    rr.captain, rr.commander, rr.designOff, err = rr.spawnAgents(ctx, commission)
    if err != nil {
        return nil, fmt.Errorf("spawn agents: %w", err)
    }
    defer rr.terminateSessions()

    // 4. Run planning loop
    manifest, err := rr.planningLoop(ctx)
    if err != nil {
        return nil, fmt.Errorf("planning loop: %w", err)
    }

    return manifest, nil
}
```

### Planning Loop
```go
func (rr *ReadyRoom) planningLoop(ctx context.Context) (*MissionManifest, error) {
    for rr.iteration < rr.maxIter {
        // 1. Send commission context to all agents
        if err := rr.sendCommissionToAgents(ctx); err != nil {
            return nil, err
        }

        // 2. Collect agent analyses
        captainAnal, commanderAnal, designAnal, err := rr.collectAnalyses(ctx)
        if err != nil {
            return nil, err
        }

        // 3. Check for questions
        if question := rr.extractQuestion(captainAnal, commanderAnal, designAnal); question != nil {
            answer, err := rr.askAdmiral(ctx, question)
            if err != nil {
                return nil, err
            }
            rr.sendAnswerToAgent(ctx, question.AgentRole, answer)
            continue
        }

        // 4. Validate consensus
        if rr.validateConsensus(captainAnal, commanderAnal, designAnal) {
            break // Consensus reached
        }

        // 5. Route inter-agent messages
        if err := rr.routeMessages(ctx); err != nil {
            return nil, err
        }

        rr.iteration++
    }

    if rr.iteration >= rr.maxIter {
        return nil, errors.New("planning failed: max iterations exceeded")
    }

    // 6. Produce mission manifest
    return rr.buildManifest(), nil
}
```

### Message Routing
```go
type ReadyRoomMessage struct {
    ID        string
    FromAgent string  // "captain" | "commander" | "design_officer"
    ToAgent   string  // Agent name or "broadcast"
    Domain    string  // "functional" | "technical" | "design"
    Content   string
    Timestamp time.Time
}

func (rr *ReadyRoom) routeMessages(ctx context.Context) error {
    // Get messages from all agents
    messages := rr.collectOutboundMessages()

    // Route to recipients
    for _, msg := range messages {
        if msg.ToAgent == "broadcast" {
            // Send to all agents
            rr.sendMessageToAgent(ctx, rr.captain, msg)
            rr.sendMessageToAgent(ctx, rr.commander, msg)
            rr.sendMessageToAgent(ctx, rr.designOff, msg)
        } else {
            // Send to specific agent
            session := rr.getAgentSession(msg.ToAgent)
            rr.sendMessageToAgent(ctx, session, msg)
        }
    }

    // Persist messages to Beads
    return rr.persistMessages(ctx, messages)
}
```

### Consensus Validation
```go
func (rr *ReadyRoom) validateConsensus(captainAnal, commanderAnal, designAnal *AgentAnalysis) bool {
    // 1. Check all missions have three-way sign-off
    for _, mission := range rr.missions {
        if !mission.Signoffs.Captain || !mission.Signoffs.Commander || !mission.Signoffs.DesignOfficer {
            return false // Missing sign-off
        }
    }

    // 2. Check all use cases covered
    covered := rr.buildUseCaseCoverage()
    for _, uc := range rr.commission.UseCases {
        if covered[uc.ID] != "covered" {
            return false // Uncovered use case
        }
    }

    return true // Consensus reached
}
```

### Admiral Question Handling
```go
func (rr *ReadyRoom) askAdmiral(ctx context.Context, question *Question) (string, error) {
    // Publish question event (TUI will show modal)
    if err := rr.eventBus.Publish(ctx, protocol.Event{
        Type:      "ADMIRAL_QUESTION",
        Body:      question.Text,
        Metadata:  map[string]string{"agent_role": question.AgentRole},
    }); err != nil {
        return "", err
    }

    // Wait for answer event
    answerCh := rr.eventBus.Subscribe(filter{
        Type: "ADMIRAL_ANSWER",
    })
    defer rr.eventBus.Unsubscribe(answerCh)

    select {
    case event := <-answerCh:
        return event.Body, nil
    case <-time.After(5 * time.Minute):
        return "", errors.New("question timeout")
    case <-ctx.Done():
        return "", ctx.Err()
    }
}
```

## Anti-Patterns to Avoid

### ❌ DON'T: Share context between agents
```go
// BAD! Agents share context window
ctx := context.Background()
captain := harness.NewSession(ctx)
commander := harness.NewSession(ctx) // Same context!
```

### ✅ DO: Isolated sessions
```go
// GOOD! Separate contexts
captainCtx := context.WithValue(context.Background(), "role", "captain")
commanderCtx := context.WithValue(context.Background(), "role", "commander")
```

### ❌ DON'T: Allow agents to communicate directly
```go
// BAD! Direct agent-to-agent communication
captain.SendMessageTo(commander, "...")
```

### ✅ DO: Route via orchestrator
```go
// GOOD! Orchestrator controls routing
rr.routeMessage(ReadyRoomMessage{
    FromAgent: "captain",
    ToAgent:   "commander",
    Content:   "...",
})
```

## Testing Requirements

### Unit Tests
- Test consensus validation logic
- Test message routing
- Test use case coverage calculation

### Integration Tests
- Test full planning loop with mock harness
- Test Admiral question flow
- Test planning iteration limits

## References

- `.spec/prd.md` - UC-COMM-01 through UC-COMM-10 (10 Planning use cases)
- `.spec/technical-requirements.md` - Commission & Planning architecture (lines 198-270)
- `.spec/research-findings/GO_CODING_STANDARDS.md` - Go coding standards

---

**Version**: 1.0
**Last Updated**: 2025-02-10
