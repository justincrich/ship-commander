# Ship Commander 3 - Root Agent Coordination

## Purpose

Ship Commander 3 is a **Go-based CLI orchestration runtime** that manages parallel AI coding agents through deterministic verification gates and Beads-backed persistent state. This document provides project-wide coordination for all agents (human and AI).

## Quick Context

Ship Commander 3 implements the **Starfleet Software Engineering Doctrine**:
- **Single-Mission Completion**: All work completes within a single mission (Define → Verify → Implement → Validate → Decide)
- **Observable System Change**: Every mission produces an observable artifact (test, CLI output, UI change, schema diff)
- **Deterministic Systems Decide**: State machines, gates, and routing are deterministic. Probabilistic agents explore; deterministic systems decide.
- **No Agent Self-Certification**: All agent claims are independently verified by the orchestrator
- **Risk-Aware Execution**: RED_ALERT missions require full TDD; STANDARD_OPS missions require proof but not necessarily tests

## Architecture Overview

```
COMMISSION (PRD)
  │
  ├── [Captain's Ready Room] → Planning loop with 3 agents (Captain, Commander, Design Officer)
  │     ├── Collaborative mission decomposition
  │     ├── Inter-agent message routing
  │     ├── Agent → Admiral questions
  │     └── Consensus validation
  │
  ├── [Admiral Approval Gate] → Human approves full mission manifest
  │
  └── [Commander Execution] → Wave-based mission dispatch
        ├── Per-mission TDD cycle (RED → VERIFY → GREEN → VERIFY → REFACTOR → VERIFY)
        ├── Independent verification gates (shell commands + exit codes)
        ├── Demo token validation
        └── Admiral wave reviews (feedback checkpoints)
```

## Technology Stack

| Component | Technology | Purpose |
|-----------|------------|---------|
| **Language** | Go 1.22+ | Single binary distribution, goroutine concurrency |
| **CLI Framework** | Cobra | Command organization and flag parsing |
| **TUI Framework** | Bubble Tea + Lipgloss | Elm architecture TUI, LCARS styling |
| **State Layer** | Beads (Git + SQLite) | Persistent state, crash recovery |
| **Session Mgmt** | tmux | Agent process isolation |
| **Observability** | OpenTelemetry + JSON logs | Distributed tracing, structured logging |
| **Testing** | testing + testify | Table-driven tests, assertions |

## Non-Negotiable Principles

### 1. Deterministic Gates, Probabilistic Agents
- **Gates are 100% deterministic**: Shell commands + exit code checking. No AI in gates.
- **Agents explore**: Agents can try different approaches, but all claims must pass independent verification.

### 2. No Agent Self-Certification
- Agent posts claim → Commander runs independent verification → Commander transitions state
- No "I tested it and it works" → Only "I ran this command and here's the output"

### 3. Every Mission Terminates
- Commander enforces `maxRevisions` (default: 3)
- Commander enforces demo token production
- No infinite retry loops

### 4. Every Mission Yields Proof
- `demo/MISSION-<id>.md` required before completion
- Schema validation (YAML frontmatter, evidence sections, file references)
- Human-reviewable in minutes

### 5. Beads as Source of Truth
- All persistent state goes through Beads (commissions, missions, agents, events)
- Crash recovery: Commander reconstructs state from Beads on restart
- Human-readable: JSONL files, grep-able, Git-trackable

## High-Level Workflow

### Phase 1: Commission Definition
1. Parse PRD Markdown into Commission with use cases
2. Persist to Beads with unique ID

### Phase 2: Ready Room Planning (One-Time)
1. Spawn 3 agent sessions: Captain, Commander, Design Officer
2. Collaborative decomposition: Use cases → Missions
3. Three-way sign-off: Captain (functional), Commander (technical), Design Officer (design)
4. Admiral questions: Any agent can surface questions mid-planning
5. Produce mission manifest (sequenced by dependency)
6. **Admiral Approval**: Human approves full manifest before execution

### Phase 3: Commander Execution (Wave-Based)
1. **Wave K**: Dispatch unblocked missions in parallel
2. **Per Mission**:
   - Create Git worktree
   - Acquire surface-area lock
   - Dispatch ensign (implementer)
   - Execute TDD cycle: RED → VERIFY_RED → GREEN → VERIFY_GREEN → REFACTOR → VERIFY_REFACTOR
   - Validate demo token
   - Dispatch ensign (reviewer, different from implementer)
   - Review verdict: APPROVED (done) or NEEDS_FIXES (re-dispatch)
3. **Wave Review**: Admiral reviews completed demos, provides feedback
4. **Wave K+1**: Repeat with feedback injected

### Phase 4: Completion
- All missions complete → Commission status: `completed`
- Demo tokens available for human review

## Coding Standards (MANDATORY)

All agents MUST follow the Go coding standards in `.spec/research-findings/GO_CODING_STANDARDS.md`:

### Essential Tooling
- **gofmt**: Non-negotiable canonical formatting
- **goimports**: Import management (superset of gofmt)
- **golangci-lint**: Meta-linter with 50+ linters
- **staticcheck**: Advanced static analysis

### Testing Requirements
- **Table-driven tests**: Idiomatic Go pattern
- **testify**: Assertion library
- **Coverage**: Target >80%
- **Race detector**: Run `go test -race ./...` in CI

### Error Handling
- **Early returns**: Use guard clauses, avoid deep nesting
- **Error wrapping**: Use `fmt.Errorf("context: %w", err)` with `%w` verb
- **No panics**: Return errors instead of panicking in production code

### Concurrency
- **Context first**: Always pass `context.Context` as first parameter
- **Controlled goroutines**: Use context cancellation, no fire-and-forget
- **Channels for orchestration**: Use channels for communication, mutexes for serialization

## Project Structure

```
ship-commander-3/
├── cmd/                    # CLI entry points (Cobra commands)
├── internal/               # Private application code
│   ├── commission/         # Commission lifecycle
│   ├── planning/           # Ready Room orchestrator
│   ├── execution/          # Commander orchestrator
│   ├── gates/              # Verification gate engine
│   ├── harness/            # AI harness abstraction
│   ├── beads/              # Beads state layer client
│   ├── protocol/           # Protocol event system
│   ├── telemetry/          # OpenTelemetry tracing
│   ├── logging/            # JSON structured logging
│   ├── tui/                # Bubble Tea TUI
│   ├── config/             # Configuration management
│   ├── doctor/             # Health monitoring
│   └── util/               # Internal utilities
├── pkg/                    # Public library code (optional)
├── test/                   # Test data and tools
└── scripts/                # Build and install scripts
```

Each directory has its own `AGENTS.md` with package-specific guidance.

## Beads Workflow

This project uses **Beads** (Git + SQLite) for issue tracking and state persistence:

### Quick Reference
```bash
bd ready              # Find available work
bd show <id>          # View issue details
bd update <id> --status in_progress  # Claim work
bd close <id>         # Complete work
bd sync               # Sync with git
```

### Beads in Ship Commander 3
- **Commissions are beads**: Parent beads with use cases and acceptance criteria
- **Missions are child beads**: Linked to parent commission
- **State transitions**: Enforced by deterministic state machine
- **Protocol events**: All state changes, gate results, agent communications stored as beads
- **Crash recovery**: On restart, reconstruct state from Beads

### Landing the Plane (Session Completion)

**When ending a work session**, you MUST complete ALL steps below. Work is NOT complete until `git push` succeeds.

**MANDATORY WORKFLOW:**

1. **File issues for remaining work** - Create issues for anything that needs follow-up
2. **Run quality gates** (if code changed) - Tests, linters, builds
3. **Update issue status** - Close finished work, update in-progress items
4. **PUSH TO REMOTE** - This is MANDATORY:
   ```bash
   git pull --rebase
   bd sync
   git push
   git status  # MUST show "up to date with origin"
   ```
5. **Clean up** - Clear stashes, prune remote branches
6. **Verify** - All changes committed AND pushed
7. **Hand off** - Provide context for next session

**CRITICAL RULES:**
- Work is NOT complete until `git push` succeeds
- NEVER stop before pushing - that leaves work stranded locally
- NEVER say "ready to push when you are" - YOU must push
- If push fails, resolve and retry until it succeeds

## Testing Strategy

### Unit Tests
- **Table-driven tests** for all packages
- **Mock interfaces** for external dependencies (Beads, harness)
- **Godoc comments** on all exported symbols

### Integration Tests
- **End-to-end** commission → planning → execution flow
- **Gate execution** with real shell commands
- **TUI interaction** with simulated user input

### Observability Testing
- **Trace validation**: Every run produces complete trace
- **Log validation**: All logs include run_id and trace_id
- **Debug bundle**: `sc3 bugreport` collects logs, config, state

## Common Anti-Patterns to Avoid

### ❌ DON'T: Panic in production code
```go
if data == "" {
    panic("data is empty")  // BAD!
}
```

### ✅ DO: Return errors
```go
if data == "" {
    return errors.New("data is empty")  // GOOD
}
```

### ❌ DON'T: Ignore errors
```go
data, _ := ioutil.ReadFile(path)  // BAD!
```

### ✅ DO: Always handle errors
```go
data, err := ioutil.ReadFile(path)
if err != nil {
    return fmt.Errorf("read file: %w", err)  // GOOD
}
```

### ❌ DON'T: Use goroutines without lifecycle control
```go
go func() {
    for { doWork() }  // Never stops!
}()
```

### ✅ DO: Control goroutine lifecycles
```go
ctx, cancel := context.WithCancel(context.Background())
go func() {
    for {
        select {
        case <-ctx.Done():
            return  // Controlled shutdown
        default:
            doWork()
        }
    }
}()
// Remember to call cancel() when done!
```

## References

### Specifications
- `.spec/prd.md` - Complete product requirements (86 use cases)
- `.spec/technical-requirements.md` - System architecture and data model
- `.spec/research-findings/GO_CODING_STANDARDS.md` - Go coding standards

### Design Artifacts
- `design/views.yaml` - All 16 TUI screens
- `design/components.yaml` - 35 reusable components
- `design/paradigm.yaml` - Design patterns and principles

### External Documentation
- [Effective Go](https://go.dev/doc/effective_go) - Go idioms and conventions
- [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md) - Production standards
- [Bubble Tea Documentation](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Beads Documentation](https://github.com/steveyegge/beads) - State layer

## Agent Coordination

### For Human Agents (Admirals)
- You approve commissions and plans before execution
- You answer agent questions during planning
- You review completed demos between waves
- You can halt execution at any time

### For AI Agents (Captain, Commander, Design Officer, Ensigns)
- Read package-specific `AGENTS.md` before making changes
- Follow Go coding standards religiously
- Write tests before implementation (TDD)
- Document all exported code with Godoc comments
- Never panic, always return errors
- Use context for cancellation
- Verify your work produces demo tokens

### Agent Work Completion Requirements (MANDATORY)

**Before marking any work as complete**, ALL agents MUST:

1. **Format Code**: Run `make fmt` to ensure all code is formatted
2. **Run Linters**: Run `make lint` and fix all issues
3. **Run Tests**: Run `make test` and ensure all tests pass
4. **Check Coverage**: Run `make test-coverage` and verify >80% coverage
5. **Clean Up Artifacts**: Remove temporary files, build artifacts, and test caches
6. **Verify No Warnings**: Ensure `go vet ./...` passes with no warnings
7. **Update Documentation**: Update AGENTS.md if new patterns were introduced

**Quick Completion Checklist**:
```bash
# 1. Format and lint
make fmt
make lint

# 2. Run full test suite
make test

# 3. Check coverage
make test-coverage

# 4. Clean build artifacts
make clean

# 5. Verify pre-commit hooks pass
make check

# 6. Only then mark work as complete
```

**Quality Gate Enforcement**:
- Work is NOT complete until all checks pass
- Document any exceptions with clear reasoning
- Never skip linting or testing to finish faster
- If linters have false positives, document in code comments

### For Codex Agents
- Start by reading the package-specific `AGENTS.md` in the directory you're working
- Follow the directory structure laid out in this document
- Reference the technical requirements for data models and API contracts
- Use table-driven tests for all logic
- Implement Elm architecture (Model-Update-View) for TUI code
- Wrap all errors with context using `%w` verb

---

**Version**: 1.0
**Last Updated**: 2025-02-10
**Maintained By**: Ship Commander 3 Project
