# Product Requirements Document (PRD)

## Product Name

**Ship Commander 3 -- AI Coding Orchestration CLI**

---

## One-Line Description

A Go CLI orchestration runtime that manages parallel AI coding agents through commission-driven planning, deterministic verification gates, Beads-backed persistent state, and a Bubble Tea TUI -- implementing the Starfleet manifesto's strict separation of probabilistic execution from deterministic authority.

---

## Background & Motivation

### The Problem

AI coding workflows degrade rapidly when scaled beyond one or two agents due to:

- **No planning discipline**: Agents jump straight to code without collaborative decomposition of requirements into executable tasks
- **Agent self-certification**: Agents claim test results without independent verification (vanity tests)
- **Collapsed authority**: Planning (what to build), orchestration (how to build), and execution (writing code) happen in the same context with no separation of concerns
- **No termination guarantee**: Missions can retry infinitely with no enforced halt conditions
- **No verifiable proof**: Agents claim completion without producing artifacts a human can review in minutes
- **No human approval gate**: Work is dispatched without human sign-off on the plan

### The Manifesto

Ship Commander 3 implements the **Starfleet Software Engineering Doctrine** -- a set of non-negotiable engineering principles:

1. **Single-Mission Completion**: All work completes within a single mission (Define → Verify → Implement → Validate → Decide)
2. **Observable System Change**: Every mission produces an observable artifact (test, CLI output, UI change, schema diff)
3. **Deterministic Systems Decide**: State machines, gates, and routing are deterministic. Probabilistic agents explore; deterministic systems decide.
4. **No Agent Self-Certification**: All agent claims are independently verified by the orchestrator
5. **Risk-Aware Execution**: RED_ALERT missions require full TDD; STANDARD_OPS missions require proof but not necessarily tests

### What Ship Commander 3 Builds

Ship Commander 3 is the **orchestration runtime** that enforces these principles. It:

- Implements a **commission-driven planning loop** (Captain's Ready Room) where specialists collaboratively decompose PRDs into missions before any code is written
- Provides a **Commander** as a distinct orchestrator that owns mission execution, verification gates, and termination enforcement -- separate from the Captain's requirement validation
- Requires **Admiral (human) approval** before any mission is dispatched, with plan persistence for save/shelve/resume
- Uses **Beads** (Git + SQLite) as the persistent state layer natively
- Provides a **Bubble Tea TUI** with LCARS-themed dashboard for real-time visibility
- Supports **Claude Code** (default) and **Codex** (optional) as interchangeable AI harnesses
- Uses **tmux** for session management of concurrent agent sessions
- Produces **demo tokens** -- verifiable proof artifacts that a human can review without reading all code

### Authority Model

| Role | Authority | Responsibility | Analogy |
|------|-----------|---------------|---------|
| **Admiral** (Human) | Final approval authority | Commissions work, approves plans, answers agent questions, can shelve/resume plans | Product Owner |
| **Captain** (Agent) | Commission adherence | Validates functional requirements, ensures use-case coverage, strategic oversight | Product Manager |
| **Commander** (Orchestrator) | Mission execution | Decomposes commissions into missions, sequences by dependency, dispatches agents, runs gates, enforces termination | Engineering Manager |
| **Design Officer** (Agent) | Design requirements | UX/UI concerns, component architecture, design system adherence | UI/UX Lead |
| **Implementer** (Agent) | Code production | Writes code and tests within a single mission, reports claims to Commander | Developer |

The Admiral _commissions and approves_; the Captain is _in charge of commission adherence_; the Commander is _in control of mission execution_; the Design Officer is _in control of design_.

---

## Core Philosophy (Non-Negotiable)

| # | Principle | Enforcement |
|---|-----------|-------------|
| 1 | **Probabilistic systems explore; deterministic systems decide** | Commander runs all gates as shell commands + exit codes. No AI in verification. |
| 2 | **No agent self-certifies** | Agent posts claim → Commander runs independent verification → Commander transitions state |
| 3 | **Every mission terminates** | Commander enforces `maxRevisions` and demo token production |
| 4 | **Every mission yields proof** | Demo token (`demo/MISSION-<id>.md`) required before completion |
| 5 | **Commission before execution** | No mission dispatched without Ready Room planning + Admiral approval |
| 6 | **Persistent state over session memory** | Beads survives agent crashes, session loss, orchestrator restarts |
| 7 | **Observability is mandatory** | Event bus, protocol events, TUI dashboard -- every state change is visible |
| 8 | **Isolation by default** | Git worktree per mission, surface-area locks before dispatch |

---

## Execution Model

### Commission → Planning → Execution Flow

```
COMMISSION (PRD)
  │
  ├── [Captain's Ready Room]
  │     ├── Captain analyzes functional requirements
  │     ├── Design Officer analyzes design requirements
  │     ├── Commander decomposes use cases into missions
  │     ├── Agents converse via structured message passing
  │     ├── Agent → Admiral questions (mid-planning)
  │     └── Commander produces mission manifest (sequenced)
  │
  ├── [Admiral Approval Gate]
  │     ├── Admiral reviews manifest
  │     ├── Approve → execute
  │     ├── Feedback → reconvene Ready Room
  │     └── Shelve → save for later
  │
  └── [Commander Execution]
        ├── Dispatch agent to mission worktree
        ├── Agent claims "RED_COMPLETE" (protocol event)
        ├── Commander runs VERIFY_RED gate (shell + exit code)
        ├── Commander transitions state (deterministic)
        ├── Repeat for GREEN, REFACTOR
        ├── Commander validates demo token file
        └── Mission complete or halted
```

### Dual Track: RED_ALERT vs STANDARD_OPS

| Track | When | Phases | Proof Required |
|-------|------|--------|---------------|
| **RED_ALERT** (TDD) | Business logic, APIs, auth, data integrity, bug fixes | Per-AC: RED → VERIFY_RED → GREEN → VERIFY_GREEN → REFACTOR → VERIFY_REFACTOR | `tests` + (`commands` or `diff_refs`) |
| **STANDARD_OPS** | Styling, non-behavioral refactors, tooling, docs | IMPLEMENT → VERIFY_IMPLEMENT → Demo Token | At least one of: `commands`, `manual_steps`, `diff_refs` |

### Wave Execution

```
[WAVE 1] - No dependencies (dispatched in parallel)
    ├── MISSION-A (worktree-1, RED_ALERT)  → TDD → Review
    ├── MISSION-B (worktree-2, RED_ALERT)  → TDD → Review
    └── MISSION-C (worktree-3, STANDARD_OPS) → Implement → Review

[WAIT for Wave 1 completion]

[WAVE 2] - Depends on Wave 1
    ├── MISSION-D (blocked by A, RED_ALERT)  → TDD → Review
    └── MISSION-E (blocked by B, STANDARD_OPS) → Implement → Review
```

---

## In Scope

1. **Commission type** -- formal PRD-level initiative with use cases, acceptance criteria, and use-case-to-mission traceability
2. **Captain's Ready Room** -- collaborative planning loop where Captain, Commander, and Design Officer decompose commissions into missions with domain ownership and three-way sign-off
3. **Admiral approval gate** -- human reviews mission manifest before dispatch; feedback reconvenes planning loop; plans can be saved, shelved, resumed
4. **Agent → Admiral question gate** -- any agent can ask the Admiral questions mid-planning via TUI prompt; loop suspends until answer
5. **Commander orchestrator** -- mission execution authority that dispatches agents, runs verification gates independently, enforces termination, and reports status
6. **Verification gate engine** -- Commander-owned deterministic gate execution (shell commands + exit codes) for VERIFY_RED, VERIFY_GREEN, VERIFY_REFACTOR, VERIFY_IMPLEMENT
7. **Protocol event system** -- structured JSON events for agent ↔ Commander coordination with schema validation and persistence
8. **Demo token validator** -- strict Markdown schema enforcement for `demo/MISSION-<id>.md` proof artifacts
9. **Termination enforcer** -- deterministic rules for when Commander halts a mission (max revisions, missing demo token, AC exhaustion)
10. **Surface-area locking** -- Commander acquires locks before dispatch to prevent concurrent missions modifying same subsystem
11. **Beads state layer** -- persistent state via Git + SQLite for commissions, missions, agents, protocol events, gate evidence, and audit log
12. **Multi-harness support** -- Claude Code (default) and Codex drivers with harness-agnostic session abstraction
13. **tmux session management** -- agent sessions managed via tmux for process isolation and recovery
14. **Bubble Tea TUI** -- LCARS-themed terminal dashboard with planning view, execution view, Admiral question modal, and health monitoring
15. **Standard Ops fast path** -- Commander routes STANDARD_OPS missions through simplified IMPLEMENT → VERIFY → Demo Token flow
16. **Worktree-per-mission isolation** -- Git worktrees for code isolation with branch naming conventions
17. **Crash recovery** -- Commander reconstructs state from Beads on restart; orphaned missions returned to backlog

## Out of Scope

1. Web dashboard -- TUI only for v1
2. More than 5 concurrent agents -- architecture supports 20+, v1 targets 3-5
3. Merge conflict resolution -- human handles merges in v1
4. OpenCode harness -- Claude Code and Codex only for v1
5. Screenshots, E2E reports, trace files, video recordings as demo token evidence types -- V2+ when infrastructure exists
6. Remote/distributed agent execution -- all agents local
7. Agent-to-agent direct communication -- all routing through Commander
8. Cost tracking / token usage monitoring
9. Session recording/replay for debugging
10. WSJF prioritization algorithm -- manual priority ordering in v1
11. Formula/workflow templates -- ad-hoc commissions only

---

## Roles

| Role | Description | Session Type |
|------|-------------|-------------|
| **Admiral** (Human) | Final authority. Commissions work, approves plans, answers questions, can shelve/resume. Operates via TUI. | Human at terminal |
| **Captain** (Agent) | Owns functional requirements. Analyzes commission use cases, validates coverage, ensures missions collectively satisfy PRD. | Harness session (own context) |
| **Commander** (Orchestrator) | Owns mission execution. Decomposes commissions into missions, sequences, dispatches, runs gates, enforces termination. Part deterministic logic, part agent session. | Harness session + deterministic loop |
| **Design Officer** (Agent) | Owns design requirements. Reviews commission for UX/UI concerns, component architecture, design system adherence. | Harness session (own context) |
| **Implementer** (Agent) | Writes code. Receives mission spec from Commander, executes TDD phases or implementation, posts protocol claims. Ephemeral. | Harness session in worktree |
| **Reviewer** (Agent) | Reviews code. Receives diff + gate evidence + acceptance criteria. Context-isolated from implementer. | Harness session (read-only) |
| **Doctor** (Process) | Background health monitor. Detects stuck agents, orphaned missions, heartbeat failures. | Deterministic Go goroutine |

---

## Functional Groups

### COMM -- Commission & Planning

**Prefix**: COMM | **Purpose**: Commission lifecycle, Captain's Ready Room planning loop, Admiral approval, agent questions, plan persistence

The commission is the **PRD-level initiative** that connects product requirements to their implementing missions. The Ready Room is where Captain, Commander, and Design Officer collaboratively decompose a commission's use cases into executable missions before any code is written. Each agent operates in **their own harness session** (separate context) -- collaboration happens through structured message passing between sessions.

| UC ID | Title | Description |
|-------|-------|-------------|
| UC-COMM-01 | Create commission from PRD | Parse a PRD file into a `Commission` struct with use cases, acceptance criteria, functional groups, scope boundaries, and persisted ID |
| UC-COMM-02 | Spawn Ready Room sessions | Create isolated harness sessions for Captain, Commander, and Design Officer -- each receives the commission context but maintains separate context windows |
| UC-COMM-03 | Captain functional analysis | Captain agent analyzes commission use cases and produces functional requirements -- mapping which use cases should become which missions and validating AC coverage |
| UC-COMM-04 | Design Officer analysis | Design Officer agent reviews commission for design requirements, UX implications, component architecture, and design system adherence |
| UC-COMM-05 | Commander mission decomposition | Commander agent decomposes use cases into missions (split, combine, or 1:1), drafts technical requirements, and sequences by importance and dependency |
| UC-COMM-06 | Inter-session message routing | Route structured `ReadyRoomMessage` between agent sessions -- messages arrive as additional context in target session, orchestrator controls information flow |
| UC-COMM-07 | Consensus validation | Deterministic check that all missions have three-way sign-off (Captain functional, Commander technical, Design Officer design) and all commission use cases are covered |
| UC-COMM-08 | Agent → Admiral question | Any agent can surface a question to the Admiral mid-planning -- planning loop suspends, TUI renders question prompt, answer routed back to agent session |
| UC-COMM-09 | Admiral approval gate | Present mission manifest to Admiral via TUI -- Admiral can approve, provide feedback (reconvenes loop), or shelve for later |
| UC-COMM-10 | Plan persistence | Save, shelve, resume, and re-execute mission manifests -- plans persist across sessions and can be executed at a later time |

**Acceptance Criteria**:

- UC-COMM-01:
  - `☐` System can parse a Markdown PRD file into a structured Commission with use cases and acceptance criteria
  - `☐` Commission persists to Beads with unique ID and tracks status (planning → approved → executing → completed → shelved)
  - `☐` Commission maintains use-case-to-mission traceability via `useCaseRefs` on each mission

- UC-COMM-02:
  - `☐` System can spawn three separate harness sessions (Captain, Design Officer, Commander) with the commission context injected
  - `☐` Each session operates in its own context window -- no shared memory between sessions
  - `☐` Session lifecycle (spawn, send message, read output, terminate) is managed by a deterministic planning loop

- UC-COMM-03:
  - `☐` Captain agent can return functional requirements for each mission, including use-case mapping and AC coverage analysis
  - `☐` Captain agent can return questions for the Admiral if ambiguity is encountered
  - `☐` Captain signs off on functional requirements per mission (persisted as `signoffs.captain: true`)

- UC-COMM-04:
  - `☐` Design Officer agent can return design requirements for each mission
  - `☐` Design Officer can return questions for the Admiral if design ambiguity exists
  - `☐` Design Officer signs off on design requirements per mission (persisted as `signoffs.designOfficer: true`)

- UC-COMM-05:
  - `☐` Commander agent can decompose use cases into missions using split, combine, or 1:1 mapping
  - `☐` Commander agent produces sequenced mission list ordered by importance and dependency
  - `☐` Commander signs off on technical requirements per mission (persisted as `signoffs.commander: true`)

- UC-COMM-06:
  - `☐` Messages between sessions are structured (`ReadyRoomMessage` with from, to, type, domain, content)
  - `☐` Messages arrive as additional context in the target session (not full context merge)
  - `☐` Orchestrator controls message routing -- agents cannot directly access other sessions

- UC-COMM-07:
  - `☐` Consensus check is deterministic: all missions signed off by all three agents AND all commission use cases covered by at least one mission
  - `☐` Use-case coverage map tracks which use cases are covered, partially covered, or uncovered
  - `☐` Planning loop iterates until consensus is reached or max iterations exceeded

- UC-COMM-08:
  - `☐` Agent can return a question as part of its output (alongside or instead of analysis)
  - `☐` Planning loop suspends and surfaces question to Admiral via TUI
  - `☐` Admiral can select from pre-defined options, type free text, or skip (agent uses best judgment)
  - `☐` Answer is routed back to asking agent's session; optionally broadcast to all sessions
  - `☐` Question/answer pairs linked by `questionId` and persisted in planning log

- UC-COMM-09:
  - `☐` Mission manifest presented to Admiral via TUI with full mission list, sequencing, and use-case coverage
  - `☐` Admiral can approve (missions enter execution), provide feedback (loop reconvenes with feedback injected into all sessions), or shelve (plan saved for later)
  - `☐` No mission is dispatched without Admiral approval

- UC-COMM-10:
  - `☐` Mission manifests can be saved to Beads with full state (missions, messages, sign-offs, iterations)
  - `☐` Saved plans can be resumed from their last state with agent sessions re-spawned
  - `☐` Shelved plans persist indefinitely and can be executed at any time

---

### EXEC -- Mission Execution

**Prefix**: EXEC | **Purpose**: Commander orchestrator, mission dispatch, per-AC TDD cycle, termination enforcement, demo token validation

The Commander is the **mission execution authority** -- a distinct orchestrator that sits between every agent and every state transition. It receives the approved mission manifest from the Ready Room and owns each mission from dispatch through completion or halt. The Commander runs verification gates independently (deterministic shell commands), enforces termination rules, and validates demo tokens before allowing mission completion.

| UC ID | Title | Description |
|-------|-------|-------------|
| UC-EXEC-01 | Sequence and dispatch missions | Commander sequences missions from the approved manifest by dependency and importance, creates worktrees, and dispatches agents |
| UC-EXEC-02 | Execute RED phase (per AC) | Dispatch implementer agent with mission spec and AC context; wait for `RED_COMPLETE` protocol claim |
| UC-EXEC-03 | Execute GREEN phase (per AC) | After VERIFY_RED passes, dispatch implementer for implementation; wait for `GREEN_COMPLETE` claim |
| UC-EXEC-04 | Execute REFACTOR phase (per AC) | After VERIFY_GREEN passes, dispatch implementer for refactor; wait for `REFACTOR_COMPLETE` claim |
| UC-EXEC-05 | Route STANDARD_OPS fast path | For STANDARD_OPS missions, dispatch single IMPLEMENT phase and validate demo token without per-AC TDD |
| UC-EXEC-06 | Enforce mission termination | Halt mission when `revisionCount >= maxRevisions`, demo token missing/invalid, or AC attempts exhausted |
| UC-EXEC-07 | Validate demo token | Before mission completion, validate `demo/MISSION-<id>.md` exists, has correct YAML frontmatter, required evidence sections, and referenced files exist in worktree |
| UC-EXEC-08 | Dispatch domain reviewer | After all ACs verified, spawn fresh reviewer agent with diff context, gate evidence, and acceptance criteria -- context-isolated from implementer |
| UC-EXEC-09 | Handle review verdict | Process APPROVED (complete mission) or NEEDS_FIXES (increment revision count, re-dispatch implementer with feedback) |
| UC-EXEC-10 | Report status to Captain | Commander reports mission completion/halt status back for commission-level progress tracking |

**Acceptance Criteria**:

- UC-EXEC-01:
  - `☐` Commander can read approved mission manifest and determine execution order based on dependencies
  - `☐` Commander can create isolated Git worktree per mission with branch naming convention `feature/MISSION-<id>-<slug>`
  - `☐` Commander can acquire surface-area lock before dispatch (see UC-STATE-08)
  - `☐` Commander can dispatch implementer agent via harness session abstraction

- UC-EXEC-02:
  - `☐` Dispatch includes full mission spec, current AC context, and TDD phase instruction (RED: write failing test)
  - `☐` Commander waits for `RED_COMPLETE` protocol event via polling or event subscription
  - `☐` Commander runs VERIFY_RED gate independently after receiving claim (see UC-GATE-01)
  - `☐` On gate pass: Commander transitions AC phase to `green` (deterministic state machine)
  - `☐` On gate fail: Commander records failure, provides feedback to agent, increments attempt count

- UC-EXEC-03:
  - `☐` Dispatch includes prior AC context and instruction to implement minimal code to pass test
  - `☐` Commander waits for `GREEN_COMPLETE` protocol event
  - `☐` Commander runs VERIFY_GREEN gate independently (see UC-GATE-02)
  - `☐` On gate pass: Commander transitions AC phase to `refactor`
  - `☐` On gate fail: Commander loops back to GREEN with failure output

- UC-EXEC-04:
  - `☐` Dispatch includes refactor instruction (clean up without changing behavior)
  - `☐` Commander waits for `REFACTOR_COMPLETE` protocol event
  - `☐` Commander runs VERIFY_REFACTOR gate (see UC-GATE-03)
  - `☐` On gate pass: Commander marks AC complete, advances to next AC or mission review

- UC-EXEC-05:
  - `☐` Commander detects STANDARD_OPS classification and routes through fast path
  - `☐` Single IMPLEMENT dispatch with full mission spec (no per-AC TDD iteration)
  - `☐` Commander runs VERIFY_IMPLEMENT gate (typecheck, lint, build if configured)
  - `☐` Commander validates demo token before completion

- UC-EXEC-06:
  - `☐` Commander checks `revisionCount >= maxRevisions` before each dispatch (deterministic)
  - `☐` Commander checks demo token validity before allowing mission completion
  - `☐` Commander halts mission with structured reason when termination conditions met
  - `☐` Halted missions emit protocol event and TUI notification

- UC-EXEC-07:
  - `☐` Validator checks `demo/MISSION-<id>.md` exists in worktree
  - `☐` Validator parses YAML frontmatter and validates required fields (mission_id, title, classification, status, created_at, agent_id)
  - `☐` Validator checks `mission_id` matches the actual mission
  - `☐` Validator checks `diff_refs` reference real files in the worktree
  - `☐` Validator enforces mode-dependent section requirements: RED_ALERT requires `tests` + (`commands` or `diff_refs`); STANDARD_OPS requires at least one evidence section
  - `☐` Validator returns structured pass/fail with reason

- UC-EXEC-08:
  - `☐` Reviewer agent is a different harness session than implementer (context isolation)
  - `☐` Reviewer receives: code diff, gate evidence history, acceptance criteria, demo token content
  - `☐` Reviewer does NOT receive implementer's internal reasoning

- UC-EXEC-09:
  - `☐` APPROVED verdict transitions mission to `done` state
  - `☐` NEEDS_FIXES increments `revisionCount`, transitions mission back to `in_progress`
  - `☐` `revisionCount >= maxRevisions` after NEEDS_FIXES triggers HALTED state

- UC-EXEC-10:
  - `☐` Commander emits mission status events (completed/halted) as protocol events
  - `☐` Captain can query commission-level progress (completed missions vs total missions, use-case coverage)

---

### GATE -- Verification Gates

**Prefix**: GATE | **Purpose**: Commander-owned deterministic gate execution, protocol events for agent ↔ Commander coordination

Verification gates are **entirely deterministic** -- shell commands, exit code checking, output parsing. No AI is involved. The Commander runs these independently after every agent claim. The protocol event system provides structured JSON communication between agents and Commander.

| UC ID | Title | Description |
|-------|-------|-------------|
| UC-GATE-01 | Execute VERIFY_RED gate | Run test command in worktree; exit 0 = REJECT vanity test; syntax error = REJECT; non-zero with test failure = ACCEPT |
| UC-GATE-02 | Execute VERIFY_GREEN gate | Run full test suite in worktree; exit 0 = ACCEPT; non-zero = REJECT with failure details. For infra track: run 3x for flakiness detection |
| UC-GATE-03 | Execute VERIFY_REFACTOR gate | Run full test suite after refactor; exit 0 = ACCEPT (behavior unchanged); non-zero = REJECT |
| UC-GATE-04 | Execute VERIFY_IMPLEMENT gate | Run configured validation commands (typecheck, lint, build) in worktree; all must pass |
| UC-GATE-05 | Record gate evidence | Persist gate result (type, exit code, classification, output snippet, timestamp, attempt) to Beads |
| UC-GATE-06 | Publish protocol event | Validate JSON event against schema, persist to Beads, emit to event bus. Supports: AGENT_CLAIM, GATE_RESULT, STATE_TRANSITION |
| UC-GATE-07 | Wait for agent claim | Poll Beads for protocol event matching expected claim type and mission ID, with configurable timeout |
| UC-GATE-08 | Configure gate commands per mission | Mission spec includes `test_command`, `typecheck_command`, `lint_command`, `build_command` with variable substitution |

**Acceptance Criteria**:

- UC-GATE-01:
  - `☐` Gate runs exact test command in worktree directory (not project root)
  - `☐` Exit code 0 without implementation → REJECT with classification `reject_vanity`
  - `☐` Syntax/import error in output → REJECT with classification `reject_syntax`
  - `☐` Non-zero exit with test failure pattern → ACCEPT with classification `accept`
  - `☐` Gate output captured with configurable size limit (default 1MB)
  - `☐` Gate timeout enforced (default 120s)

- UC-GATE-02:
  - `☐` Runs full test suite (not just new test file)
  - `☐` Exit 0 → ACCEPT
  - `☐` Non-zero → REJECT with first failure message extracted for implementer feedback
  - `☐` For infrastructure-sensitive tests: run 3x, pass rate < 100% → REJECT with `Flaky test detected: {rate}% pass rate`

- UC-GATE-03:
  - `☐` Runs full test suite after refactor
  - `☐` Exit 0 → ACCEPT (behavior preserved)
  - `☐` Non-zero → REJECT refactor

- UC-GATE-04:
  - `☐` Runs each configured command (typecheck, lint, build) sequentially
  - `☐` All commands must exit 0 to pass
  - `☐` If no commands configured, gate passes automatically
  - `☐` Failure output captured for implementer feedback

- UC-GATE-05:
  - `☐` Evidence includes: gate type, mission/AC ID, exit code, classification (accept/reject_vanity/reject_syntax/reject_failure), output snippet, timestamp, attempt number
  - `☐` Evidence persisted to Beads and queryable by mission ID
  - `☐` Evidence emitted as event for TUI display

- UC-GATE-06:
  - `☐` Protocol events validated against JSON schema before persistence
  - `☐` Events include `protocol_version: "1.0"` for forward compatibility
  - `☐` Events persisted to Beads for replay and audit
  - `☐` Events emitted to event bus for real-time TUI updates

- UC-GATE-07:
  - `☐` Commander polls for expected claim (e.g., `RED_COMPLETE`) by mission ID
  - `☐` Configurable timeout (default 5 minutes)
  - `☐` Timeout triggers escalation event (mission marked stuck)

- UC-GATE-08:
  - `☐` Mission spec includes `test_command`, `typecheck_command`, `lint_command`, `build_command` fields
  - `☐` Commands support variable substitution: `{test_file}` for per-AC test file filtering
  - `☐` If not specified, system uses project-level defaults from config

---

### HARN -- Harness & Sessions

**Prefix**: HARN | **Purpose**: Spawn and manage AI agent sessions via Claude Code and Codex, with tmux for process isolation

The harness layer abstracts which CLI tool executes the agent. All harnesses implement the same interface: spawn an agent with a prompt in a working directory, capture output, enforce timeout, return result. tmux provides session management for concurrent agent processes.

| UC ID | Title | Description |
|-------|-------|-------------|
| UC-HARN-01 | Define harness session interface | Common Go interface for spawning, messaging, and terminating agent sessions across any CLI harness |
| UC-HARN-02 | Implement Claude Code driver | Spawn `claude` CLI with `-p`, model selection, timeout, and working directory in a tmux session |
| UC-HARN-03 | Implement Codex driver | Spawn `codex` CLI with sandbox mode, approval policy, and model configuration in a tmux session |
| UC-HARN-04 | Manage tmux sessions | Create, list, send-keys, capture-pane, and kill tmux sessions for agent process isolation |
| UC-HARN-05 | Stream agent output | Real-time capture of agent stdout/stderr via tmux capture-pane for event bus propagation |
| UC-HARN-06 | Enforce session timeout | SIGTERM → grace period → SIGKILL escalation with cleanup verification and no zombie processes |
| UC-HARN-07 | Detect harness availability | Check which CLI tools (`claude`, `codex`) are on PATH at startup; fail fast if none available |
| UC-HARN-08 | Configure harness per role | Allow role-level harness and model configuration via TOML settings (e.g., Captain uses Opus, Implementer uses Sonnet) |

**Acceptance Criteria**:

- UC-HARN-01:
  - `☐` Go interface defines `SpawnSession(role, prompt, workdir, opts) (*Session, error)` and `SendMessage(session, message) (string, error)` and `Terminate(session) error`
  - `☐` Interface supports optional `OnOutput` callback for real-time streaming
  - `☐` Interface returns structured output (exit code, stdout, stderr, duration)

- UC-HARN-02:
  - `☐` Claude Code driver constructs correct CLI flags: `-p --model <model> --verbose --max-turns <n>`
  - `☐` Driver supports model selection (haiku, sonnet, opus) per configuration
  - `☐` Driver runs claude CLI within a tmux session for isolation
  - `☐` Driver captures output via tmux capture-pane

- UC-HARN-03:
  - `☐` Codex driver constructs correct CLI flags: `--sandbox <mode> -m <model> exec -`
  - `☐` Driver supports sandbox modes (read-only, workspace-write, danger-full-access)
  - `☐` Driver supports approval policies (untrusted, on-failure, on-request, never)
  - `☐` Driver runs codex CLI within a tmux session for isolation

- UC-HARN-04:
  - `☐` System can create named tmux sessions: `sc3-<role>-<mission-id>`
  - `☐` System can list active tmux sessions for agent inventory
  - `☐` System can send keystrokes to tmux sessions for agent input
  - `☐` System can kill tmux sessions for cleanup

- UC-HARN-05:
  - `☐` Agent output captured via `tmux capture-pane` on configurable interval
  - `☐` Output chunks emitted as events to event bus for TUI display
  - `☐` Output truncated at configurable limit (default 1MB) to prevent memory bloat

- UC-HARN-06:
  - `☐` SIGTERM sent first with 5-second grace period
  - `☐` If process still running after grace period, SIGKILL sent
  - `☐` tmux session killed after process termination
  - `☐` No zombie processes after timeout or crash

- UC-HARN-07:
  - `☐` System checks for `claude` and `codex` on PATH at startup
  - `☐` System reports which harnesses are available
  - `☐` System errors if zero harnesses available
  - `☐` System checks for `tmux` on PATH (required dependency)

- UC-HARN-08:
  - `☐` TOML config allows per-role harness and model: `[roles.captain] harness = "claude" model = "opus"`
  - `☐` If configured harness unavailable, falls back to default with warning
  - `☐` Default harness configurable: `[defaults] harness = "claude" model = "sonnet"`

---

### STATE -- State Persistence

**Prefix**: STATE | **Purpose**: Beads-backed persistent state for commissions, missions, agents, protocol events, locks, and audit trail

All state is stored via **Beads** (Git + SQLite). Commissions are beads. Missions are child beads of commissions. Agent state, gate evidence, protocol events, and surface-area locks are tracked through Beads dimensions, comments, and labels. The state machine enforces legal transitions deterministically.

| UC ID | Title | Description |
|-------|-------|-------------|
| UC-STATE-01 | Initialize Beads database | Run `bd init` if `.beads/` does not exist; reuse existing database on restart without data loss |
| UC-STATE-02 | Track commission lifecycle | Commission as parent bead with status dimension: `planning` → `approved` → `executing` → `completed` / `shelved` |
| UC-STATE-03 | Track mission lifecycle | Mission as child bead of commission with status: `backlog` → `in_progress` → `review` → `approved` → `done` / `halted` |
| UC-STATE-04 | Track TDD phase per AC | AC as child bead of mission with phase dimension: `red` → `verify_red` → `green` → `verify_green` → `refactor` → `verify_refactor` |
| UC-STATE-05 | Enforce legal state transitions | Deterministic assertion that validates transition legality before executing (e.g., `backlog` → `done` rejected) |
| UC-STATE-06 | Manage dependency graph | Use Beads dep system for mission dependencies; query unblocked missions for wave computation |
| UC-STATE-07 | Track agent instances | Agent state tracking via Beads with role, harness, mission assignment, heartbeat |
| UC-STATE-08 | Manage surface-area locks | Acquire, check conflicts, and release file-path locks per mission to prevent concurrent modification |
| UC-STATE-09 | Persist event log | Append-only event recording for all state transitions, gate results, protocol events, and system health |
| UC-STATE-10 | Recover from crash | On restart, reconstruct state from Beads; identify orphaned missions; resume execution loop |

**Acceptance Criteria**:

- UC-STATE-01:
  - `☐` System runs `bd init` if `.beads/` directory does not exist
  - `☐` System reuses existing `.beads/` database on restart without data loss
  - `☐` System verifies `bd` CLI is available on PATH at startup

- UC-STATE-02:
  - `☐` Commission created as parent bead with PRD content, use cases, and scope
  - `☐` Commission status transitions enforced: `planning` → `approved` → `executing` → `completed`/`shelved`
  - `☐` Commission links to child mission beads via Beads parent-child relationship

- UC-STATE-03:
  - `☐` Mission created as child bead of commission with full spec content
  - `☐` Mission status transitions enforced by deterministic state machine
  - `☐` Mission tracks: `revisionCount`, `maxRevisions`, `attemptCount`, `classification`, `commissionId`, `useCaseRefs`

- UC-STATE-04:
  - `☐` AC created as child bead of mission with title and index
  - `☐` TDD phase tracked via state dimension with legal transition enforcement
  - `☐` Gate evidence recorded as comments on AC bead

- UC-STATE-05:
  - `☐` State machine validates every transition before execution
  - `☐` Illegal transitions return error with reason (never silently skip)
  - `☐` All transitions recorded with timestamp, actor, and reason

- UC-STATE-06:
  - `☐` `bd dep add <child> <parent>` for mission dependencies
  - `☐` `bd ready` queries unblocked missions for wave computation
  - `☐` Wave assignments computed from dependency graph

- UC-STATE-07:
  - `☐` Agent bead created on spawn with role, harness, mission assignment labels
  - `☐` Agent state tracked: `idle` → `spawning` → `running` → `stuck` → `done` / `dead`
  - `☐` Heartbeat updated on activity; stale agents detected by Doctor

- UC-STATE-08:
  - `☐` Lock acquired before mission dispatch with surface area declaration (file glob patterns)
  - `☐` Conflict detection checks existing locks for overlapping patterns
  - `☐` Lock released on mission completion or halt
  - `☐` Lock has configurable expiry timeout

- UC-STATE-09:
  - `☐` All state transitions, gate results, and protocol events recorded in append-only log
  - `☐` Events queryable by mission ID, type, and time range
  - `☐` Event log survives crashes (Beads-backed)

- UC-STATE-10:
  - `☐` On startup, system reads all commissions, missions, agents from Beads
  - `☐` Orphaned missions (in_progress with no active agent) returned to backlog
  - `☐` Execution loop resumes within 10 seconds of restart

---

### TUI -- Terminal User Interface

**Prefix**: TUI | **Purpose**: Bubble Tea TUI with LCARS theme for planning visibility, execution monitoring, Admiral interactions, and system health

The TUI is built with **Bubble Tea** (framework) and **Lipgloss** (styling). It provides real-time visibility into the commission planning loop, mission execution, agent status, and system health. The Admiral interacts with the system exclusively through the TUI -- approving plans, answering questions, and monitoring progress.

| UC ID | Title | Description |
|-------|-------|-------------|
| UC-TUI-01 | Display planning dashboard | Show Ready Room status: agent sessions (Captain/Commander/Design Officer), planning iteration, sign-off progress, pending questions |
| UC-TUI-02 | Display Admiral question modal | LCARS-themed modal presenting agent questions with options, free-text input, broadcast toggle, and skip option |
| UC-TUI-03 | Display plan review overlay | Mission manifest review showing all missions, sequencing, use-case coverage, and approve/feedback/shelve controls |
| UC-TUI-04 | Display execution dashboard | Show active missions, per-AC TDD phase progress, agent status, gate results, wave progress |
| UC-TUI-05 | Display agent status grid | Show all active agents with role, assigned mission, current phase, harness type, and elapsed time |
| UC-TUI-06 | Display live event log | Scrolling log of recent events (state transitions, gate results, protocol events, agent spawns/exits) |
| UC-TUI-07 | Accept operator commands | Command input for: halt mission, retry mission, approve plan, change WIP limit, quit |
| UC-TUI-08 | Display system health | WIP utilization, agent uptime, stuck detection warnings, heartbeat status, Doctor state |
| UC-TUI-09 | Display wave execution view | Current wave, parallel missions, dependency graph, completed vs active waves |
| UC-TUI-10 | Apply LCARS theme | Lipgloss styles for LCARS color palette (orange/blue/purple), box-drawing borders, semantic status colors |

**Acceptance Criteria**:

- UC-TUI-01:
  - `☐` Planning dashboard shows each agent session with status indicator: `[✓]` done, `[●]` working, `[⏸]` waiting for Admiral, `[✗]` failed
  - `☐` Dashboard shows planning iteration count (e.g., "Iteration 2/5")
  - `☐` Dashboard shows sign-off progress per mission
  - `☐` Dashboard shows pending question count: `Questions: N pending`

- UC-TUI-02:
  - `☐` Modal identifies asking agent: "ADMIRAL — QUESTION FROM CAPTAIN"
  - `☐` Modal shows question text, context, and domain (functional/technical/design)
  - `☐` Modal supports arrow-key selection of pre-defined options
  - `☐` Modal supports `[t]` for free-text input, `[s]` for skip, `[Tab]` for broadcast toggle
  - `☐` Modal resolves the `AdmiralQuestionGate` promise on submission

- UC-TUI-03:
  - `☐` Overlay shows full mission list with sequence numbers, titles, and use-case references
  - `☐` Overlay shows use-case coverage summary (covered/partial/uncovered)
  - `☐` Overlay supports `[a]` approve, `[f]` feedback, `[s]` shelve controls

- UC-TUI-04:
  - `☐` Execution dashboard shows per-mission progress through TDD phases
  - `☐` Each AC shows phase indicators: RED → VERIFY_RED → GREEN → VERIFY_GREEN → REFACTOR → VERIFY_REFACTOR
  - `☐` Gate pass/fail results shown inline with classification
  - `☐` Missions color-coded by status: active (orange), complete (green), halted (red)

- UC-TUI-05:
  - `☐` Agent grid updates in real-time as agents spawn, progress, and exit
  - `☐` Each agent card shows: role, mission ID, TDD phase, elapsed time, harness name
  - `☐` Stuck agents highlighted in warning color (yellow)

- UC-TUI-06:
  - `☐` Event log shows last 50 events with timestamp, type, actor, and summary
  - `☐` Events color-coded by severity (info=blue, warn=yellow, error=red)
  - `☐` Log is scrollable with keyboard navigation

- UC-TUI-07:
  - `☐` Command input bar at bottom of TUI accepts: `halt <id>`, `retry <id>`, `approve <id>`, `wip <n>`, `quit`
  - `☐` Commands provide immediate feedback in event log
  - `☐` Destructive commands (halt) require confirmation

- UC-TUI-08:
  - `☐` WIP utilization bar (e.g., "3/5 agents active")
  - `☐` Doctor heartbeat indicator (green/red)
  - `☐` Count of stuck agents and orphaned missions

- UC-TUI-09:
  - `☐` Current wave number and parallel mission count displayed
  - `☐` Completed waves visually distinguished from active wave
  - `☐` Dependency relationships shown between related missions

- UC-TUI-10:
  - `☐` LCARS color palette applied via Lipgloss styles
  - `☐` Box-drawing borders using Unicode characters (`╔═╗║╚═╝` or `┌─┐│└─┘`)
  - `☐` Semantic color mapping: active=orange, success=green, error=red, warning=yellow, info=blue, planning=purple

---

## Use Case Summary

| Group | Prefix | Use Cases | Description |
|-------|--------|-----------|-------------|
| Commission & Planning | COMM | 10 | Ready Room, Admiral approval, agent questions, plan persistence |
| Mission Execution | EXEC | 10 | Commander orchestrator, TDD cycle, termination, demo tokens |
| Verification Gates | GATE | 8 | Independent gate execution, protocol events |
| Harness & Sessions | HARN | 8 | Claude Code, Codex, tmux session management |
| State Persistence | STATE | 10 | Beads integration, state machine, locks, recovery |
| Terminal Interface | TUI | 10 | Bubble Tea dashboard, LCARS theme, Admiral modals |
| **Total** | | **56** | |

---

## Constraints

| Constraint | Value | Rationale |
|-----------|-------|-----------|
| Core language | Go (Golang) | Performance, single binary distribution, goroutine concurrency |
| TUI framework | Bubble Tea + Lipgloss | Best-in-class Go TUI library, Elm architecture |
| State persistence | Beads (Git + SQLite) | Human-readable, grep-able, Git-trackable |
| Session management | tmux | Battle-tested process isolation, recoverable sessions |
| Max concurrent agents (v1) | 5 | Prove architecture before scaling |
| Default WIP limit | 3 | Conservative start |
| Max revisions per mission | 3 | Prevent infinite loops |
| Planning loop max iterations | 5 | Prevent infinite planning |
| Commission sizing | 5+ missions per commission | Amortize Admiral approval cost; <5 suggests slipping toward big-loop behavior |
| Process output limit | 1 MB | Prevent memory bloat |
| SIGTERM grace period | 5 seconds | Balance cleanup vs responsiveness |
| Doctor heartbeat interval | 30 seconds | Responsive without polling overhead |
| Stuck agent timeout | 5 minutes | Detect genuinely stuck, not slow |
| Gate command timeout | 120 seconds | Most test suites finish in < 60s |
| Protocol version | "1.0" | Forward-compatible JSON schema |
| Demo token evidence types (V1) | commands, tests, manual_steps, diff_refs | Intentionally boring; rich media in V2+ |
| Harnesses (v1) | Claude Code (default), Codex (optional) | Two harnesses from day one |
| Config format | TOML (primary), JSON (data interchange) | TOML for human config, JSON for Beads/protocol |
| Distribution | Homebrew, go install | Standard Go distribution |

---

## Terminology

| Term | Definition | Analogy |
|------|-----------|---------|
| **Commission** | PRD-level initiative with use cases and acceptance criteria. One commission produces many missions. | Epic / Initiative |
| **Mission** | Atomic coding task that goes through TDD cycle or implementation. Produces verifiable output. | Task / Issue |
| **Ready Room** | Collaborative planning loop where Captain, Commander, and Design Officer decompose a commission into missions. | Sprint planning meeting |
| **Mission Manifest** | The agreed-upon, sequenced list of missions produced by the Ready Room. Requires Admiral approval. | Sprint backlog |
| **Demo Token** | Verifiable proof artifact (`demo/MISSION-<id>.md`) that a human can review without reading all code. | Acceptance test evidence |
| **Gate** | Deterministic verification (shell command + exit code) run by Commander independently of agent. | CI/CD check |
| **Protocol Event** | Structured JSON event for agent ↔ Commander coordination. Versioned, validated, persisted. | Message queue event |
| **Surface-Area Lock** | File-path lock acquired by Commander before dispatch to prevent concurrent modification. | Database row lock |
| **Wave** | Group of missions with no inter-dependencies that execute in parallel. | Sprint batch |
| **Propulsion** | Commander's continuous loop: check backlog → dispatch → verify → advance. | CI/CD pipeline |

---

## Implementation Roadmap

### Phase 1: Foundation (Week 1-2)

1. **Go project scaffold** -- CLI entry point, TOML config, Beads wrapper, state machine
2. **Commission type + parser** -- Parse PRD Markdown into Commission struct with use cases
3. **Captain's Ready Room** -- Planning loop with harness session spawning, message routing, consensus validation
4. **Admiral approval gate** -- TUI-based plan review with approve/feedback/shelve
5. **Agent → Admiral question gate** -- Question modal with options, free text, broadcast

**Deliverable**: Commission planning loop works end-to-end with Admiral approval.

### Phase 2: Commander Execution (Week 3-4)

1. **Commander orchestrator** -- Mission dispatch, per-AC TDD cycle, state transitions
2. **Verification gate engine** -- VERIFY_RED, VERIFY_GREEN, VERIFY_REFACTOR, VERIFY_IMPLEMENT
3. **Protocol event system** -- JSON events, schema validation, Beads persistence
4. **Demo token validator** -- Markdown schema enforcement for proof artifacts
5. **Termination enforcer** -- Max revisions, missing demo token, AC exhaustion

**Deliverable**: Full mission execution with Commander-owned gates and termination.

### Phase 3: Harness & TUI (Week 5-6)

1. **Claude Code driver** -- Spawn, capture output, timeout enforcement via tmux
2. **Codex driver** -- Same interface, Codex-specific CLI flags
3. **Bubble Tea TUI** -- Planning dashboard, execution dashboard, event log
4. **LCARS theme** -- Lipgloss styles, color palette, box-drawing borders
5. **Admiral question modal** -- Full TUI component with options and free text

**Deliverable**: Working TUI with real harness integration.

### Phase 4: Advanced & Hardening (Week 7-8)

1. **Surface-area locking** -- Lock acquisition, conflict detection, release
2. **Standard Ops fast path** -- Simplified execution for low-risk missions
3. **Doctor process** -- Health monitoring, stuck detection, orphan recovery
4. **Crash recovery** -- Beads-based state reconstruction on restart
5. **Integration testing** -- Full commission → planning → execution → completion flow

**Deliverable**: Production-ready manifesto-aligned system.

---

## Success Criteria

The system is manifesto-aligned when:

1. **Admiral, Captain, Commander, and Design Officer are distinct** -- roles are not collapsed
2. **Commission→Mission traceability** -- every mission links to a commission and the use case(s) it implements
3. **Ready Room planning loop** -- missions collaboratively decomposed with domain ownership and three-way sign-off
4. **No mission dispatched without Admiral approval** -- human gate enforced before execution
5. **Agent → Admiral questions work** -- agents can ask, loop suspends, answer routes back
6. **Plans are persistable** -- save, shelve, resume, re-execute
7. **No agent self-certifies** -- Commander independently verifies every claim
8. **Every mission terminates** -- Commander enforces maxRevisions and demo token production
9. **Every mission yields proof** -- `demo/MISSION-<id>.md` required and validated
10. **Isolation by default** -- surface-area locks before dispatch
11. **Zero vanity tests pass** -- Commander's VERIFY_RED gate catches all
12. **Review is a gate** -- probabilistic review inside Commander's deterministic cage

**Measured by**:

- 100% of missions have Commander-logged termination reason
- 100% of completed missions have valid `demo/MISSION-<id>.md`
- 100% of commission use cases covered by at least one mission
- Zero concurrent modification conflicts (Commander enforces locks)
- Zero missions dispatched without Admiral greenlight
- Ready Room conversation log is persisted and auditable
