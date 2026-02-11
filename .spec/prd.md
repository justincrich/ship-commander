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
- Provides a **Bubble Tea TUI** with LCARS-themed dashboard, Glamour markdown rendering, and Huh terminal forms for real-time visibility and structured Admiral interactions
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
| **Ensign** (Specialist Agent) | Mission execution | Specialist agents assigned by domain (backend, frontend, infra, etc.). Each mission has at most 1 implementer ensign and at least 1 reviewer ensign. Commander assigns the right ensign based on mission domain. | Developer / Code Reviewer |

The Admiral _commissions and approves_; the Captain is _in charge of commission adherence_; the Commander is _in control of mission execution_; the Design Officer is _in control of design_. Ensigns are **specialist workers** -- the role is Ensign, the assignment (implementer or reviewer) is per-mission. The Commander selects ensigns by domain fit (e.g., a backend ensign for API work, a frontend ensign for UI work). A mission's implementer and reviewer must always be different ensigns for context isolation.

---

## Core Philosophy (Non-Negotiable)

| # | Principle | Enforcement |
|---|-----------|-------------|
| 1 | **Probabilistic systems explore; deterministic systems decide** | Commander runs all gates as shell commands + exit codes. No AI in verification. |
| 2 | **No agent self-certifies** | Agent posts claim → Commander runs independent verification → Commander transitions state |
| 3 | **Every mission terminates** | Commander enforces `maxRevisions` and demo token production |
| 4 | **Every mission yields proof** | Demo token (`demo/MISSION-<id>.md`) required before completion |
| 5 | **Plan once, execute many** | Ready Room plans once per commission (not per mission). Admiral approves the full manifest before any mission dispatches. Inter-wave demos provide feedback checkpoints without re-planning. |
| 6 | **Persistent state over session memory** | Beads survives agent crashes, session loss, orchestrator restarts |
| 7 | **Observability is mandatory** | Event bus, protocol events, TUI dashboard, charmbracelet/log structured logging -- every state change is visible |
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
  └── [Commander Execution (wave loop)]
        ├── Dispatch wave of missions in parallel (worktree per mission)
        ├── Per mission: Agent claims → Commander verifies → state transitions
        ├── Commander validates demo tokens
        ├── Wave complete → [Admiral Wave Review]
        │     ├── Admiral reviews completed wave demos
        │     ├── Continue → advance to next wave
        │     ├── Feedback → inject into next wave context
        │     └── Halt → stop commission execution
        └── Repeat for subsequent waves until commission complete
```

### Dual Track: RED_ALERT vs STANDARD_OPS

| Track | When | Phases | Proof Required |
|-------|------|--------|---------------|
| **RED_ALERT** (TDD) | Business logic, APIs, auth, data integrity, bug fixes | Per-AC: RED → VERIFY_RED → GREEN → VERIFY_GREEN → REFACTOR → VERIFY_REFACTOR | `tests` + (`commands` or `diff_refs`) |
| **STANDARD_OPS** | Styling, non-behavioral refactors, tooling, docs | IMPLEMENT → VERIFY_IMPLEMENT → Demo Token | At least one of: `commands`, `manual_steps`, `diff_refs` |

### Wave Execution

Missions execute in waves. Planning happens **once** during the Ready Room -- individual missions are tasks, not planning units. Between waves, the Admiral reviews completed demos and provides feedback that feeds into the next wave's context.

```
[WAVE 1] - No dependencies (dispatched in parallel)
    ├── MISSION-A (worktree-1, RED_ALERT)  → TDD → Review
    ├── MISSION-B (worktree-2, RED_ALERT)  → TDD → Review
    └── MISSION-C (worktree-3, STANDARD_OPS) → Implement → Review

[ADMIRAL WAVE REVIEW] - Human reviews Wave 1 demos
    ├── Review demo tokens from completed missions
    ├── Continue → proceed to Wave 2
    ├── Feedback → injected as context into Wave 2 dispatches
    └── Halt → stop commission (remaining missions shelved)

[WAVE 2] - Depends on Wave 1
    ├── MISSION-D (blocked by A, RED_ALERT)  → TDD → Review
    └── MISSION-E (blocked by B, STANDARD_OPS) → Implement → Review

[ADMIRAL WAVE REVIEW] - Human reviews Wave 2 demos
    └── ...
```

---

## In Scope

1. **Commission type** -- formal PRD-level initiative with use cases, acceptance criteria, and use-case-to-mission traceability
2. **Captain's Ready Room** -- one-time collaborative planning loop where Captain, Commander, and Design Officer decompose a commission into all missions up front with domain ownership and three-way sign-off
3. **Admiral approval gate** -- human reviews full mission manifest once before any dispatch; feedback reconvenes planning loop; plans can be saved, shelved, resumed
4. **Agent → Admiral question gate** -- any agent can ask the Admiral questions mid-planning via TUI prompt; loop suspends until answer
5. **Admiral wave review** -- between waves, Admiral reviews completed mission demos and provides feedback or halt; lightweight checkpoint without re-planning
6. **Commander orchestrator** -- mission execution authority that dispatches agents, runs verification gates independently, enforces termination, and reports status
7. **Verification gate engine** -- Commander-owned deterministic gate execution (shell commands + exit codes) for VERIFY_RED, VERIFY_GREEN, VERIFY_REFACTOR, VERIFY_IMPLEMENT
8. **Protocol event system** -- structured JSON events for agent ↔ Commander coordination with schema validation and persistence
9. **Demo token validator** -- strict Markdown schema enforcement for `demo/MISSION-<id>.md` proof artifacts
10. **Termination enforcer** -- deterministic rules for when Commander halts a mission (max revisions, missing demo token, AC exhaustion)
11. **Surface-area locking** -- Commander acquires locks before dispatch to prevent concurrent missions modifying same subsystem
12. **Beads state layer** -- persistent state via Git + SQLite for commissions, missions, agents, protocol events, gate evidence, and audit log
13. **Multi-harness support** -- Claude Code (default) and Codex drivers with harness-agnostic session abstraction
14. **tmux session management** -- agent sessions managed via tmux for process isolation and recovery
15. **Bubble Tea TUI** -- LCARS-themed terminal dashboard with planning view, execution view, Admiral question modal, and health monitoring
16. **Standard Ops fast path** -- Commander routes STANDARD_OPS missions through simplified IMPLEMENT → VERIFY → Demo Token flow
17. **Worktree-per-mission isolation** -- Git worktrees for code isolation with branch naming conventions
18. **Crash recovery** -- Commander reconstructs state from Beads on restart; orphaned missions returned to backlog

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
| **Ensign** (Specialist Agent) | Domain specialist (backend, frontend, infra, etc.) assigned per mission by Commander. Each mission has at most 1 implementer ensign and at least 1 reviewer ensign (must be different ensigns). As **implementer**: receives mission spec, executes TDD phases, posts protocol claims (harness session in worktree). As **reviewer**: receives diff + gate evidence + acceptance criteria, context-isolated from implementer (harness session, read-only). Ephemeral per mission. | Harness session |
| **Doctor** (Process) | Background health monitor. Detects stuck agents, orphaned missions, heartbeat failures. | Deterministic Go goroutine |

---

## Functional Groups

### COMM -- Commission & Planning

**Prefix**: COMM | **Purpose**: Commission lifecycle, Captain's Ready Room planning loop, Admiral approval, agent questions, plan persistence

The commission is the **PRD-level initiative** that connects product requirements to their implementing missions. The Ready Room is a **one-time planning event** where Captain, Commander, and Design Officer collaboratively decompose a commission's use cases into all executable missions before any code is written. Planning happens once per commission -- individual missions are tasks, not planning units. Each agent operates in **their own harness session** (separate context) -- collaboration happens through structured message passing between sessions.

| UC ID | Title | Description |
|-------|-------|-------------|
| UC-COMM-01 | Create commission from PRD | Parse a PRD file into a `Commission` struct with use cases, acceptance criteria, functional groups, scope boundaries, and persisted ID. PRD content rendered via Glamour for TUI display. |
| UC-COMM-02 | Spawn Ready Room sessions | Create isolated harness sessions for Captain, Commander, and Design Officer -- each receives the commission context but maintains separate context windows |
| UC-COMM-03 | Captain functional analysis | Captain agent analyzes commission use cases and produces functional requirements -- mapping which use cases should become which missions and validating AC coverage |
| UC-COMM-04 | Design Officer analysis | Design Officer agent reviews commission for design requirements, UX implications, component architecture, and design system adherence |
| UC-COMM-05 | Commander mission decomposition | Commander agent decomposes use cases into missions (split, combine, or 1:1), drafts technical requirements, and sequences by importance and dependency |
| UC-COMM-06 | Inter-session message routing | Route structured `ReadyRoomMessage` between agent sessions -- messages arrive as additional context in target session, orchestrator controls information flow |
| UC-COMM-07 | Consensus validation | Deterministic check that all missions have three-way sign-off (Captain functional, Commander technical, Design Officer design) and all commission use cases are covered |
| UC-COMM-08 | Agent → Admiral question | Any agent can surface a question to the Admiral mid-planning -- planning loop suspends, TUI renders question prompt, answer routed back to agent session |
| UC-COMM-09 | Admiral approval gate | Present mission manifest to Admiral via TUI (Glamour markdown + Huh approval form) -- Admiral can approve, provide feedback (reconvenes loop), or shelve for later |
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
| UC-EXEC-02 | Execute RED phase (per AC) | Dispatch ensign (implementer) with mission spec and AC context; Commander selects ensign by mission domain. Wait for `RED_COMPLETE` protocol claim |
| UC-EXEC-03 | Execute GREEN phase (per AC) | After VERIFY_RED passes, dispatch ensign for implementation; wait for `GREEN_COMPLETE` claim |
| UC-EXEC-04 | Execute REFACTOR phase (per AC) | After VERIFY_GREEN passes, dispatch ensign for refactor; wait for `REFACTOR_COMPLETE` claim |
| UC-EXEC-05 | Route STANDARD_OPS fast path | For STANDARD_OPS missions, dispatch single IMPLEMENT phase and validate demo token without per-AC TDD |
| UC-EXEC-06 | Enforce mission termination | Halt mission when `revisionCount >= maxRevisions`, demo token missing/invalid, or AC attempts exhausted |
| UC-EXEC-07 | Validate demo token | Before mission completion, validate `demo/MISSION-<id>.md` exists, has correct YAML frontmatter, required evidence sections, and referenced files exist in worktree. Render preview via Glamour in TUI. |
| UC-EXEC-08 | Dispatch domain reviewer | After all ACs verified, spawn fresh ensign (reviewer) with diff context, gate evidence, and acceptance criteria -- must be a different ensign than the implementer for context isolation |
| UC-EXEC-09 | Handle review verdict | Process APPROVED (complete mission) or NEEDS_FIXES (increment revision count, re-dispatch implementer with feedback) |
| UC-EXEC-10 | Report status to Captain | Commander reports mission completion/halt status back for commission-level progress tracking |
| UC-EXEC-11 | Admiral wave review | After all missions in a wave complete, present demo tokens to Admiral via TUI for review. Admiral can continue (next wave), provide feedback (injected into next wave context), or halt (remaining missions shelved). No re-planning -- this is a lightweight checkpoint. |

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

- UC-EXEC-11:
  - `☐` After all missions in a wave complete (or halt), Commander triggers wave review
  - `☐` Demo tokens from all completed missions in the wave rendered via Glamour in TUI
  - `☐` Admiral can continue (advance to next wave), provide feedback (text injected as context into next wave's mission dispatches), or halt (remaining waves cancelled, commission status updated)
  - `☐` Wave review is a lightweight checkpoint -- no Ready Room re-planning, no manifest revision
  - `☐` Admiral feedback persisted to Beads and linked to wave

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
| UC-HARN-08 | Configure harness per role | Allow role-level and domain-level harness and model configuration via TOML settings (e.g., Captain uses Opus, ensigns use Sonnet, backend ensigns use a specific model) |

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
  - `☐` TOML config allows per-role and per-domain harness and model: `[roles.captain] harness = "claude" model = "opus"`, `[roles.ensign.backend] model = "sonnet"`
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

The TUI is built with **Bubble Tea** (framework), **Lipgloss** (styling), **Glamour** (markdown rendering), and **Huh** (terminal forms). Glamour (the embeddable library behind the Glow CLI) renders PRD content, mission specs, demo tokens, and gate output as styled terminal markdown. Huh provides structured form components (Select, Input, Text, Confirm) that implement `tea.Model` and compose directly into Bubble Tea -- used for all Admiral interactions (approval gates, question modals, feedback input, operator commands). It provides real-time visibility into the commission planning loop, mission execution, agent status, and system health. The Admiral interacts with the system exclusively through the TUI -- approving plans, answering questions, and monitoring progress.

| UC ID | Title | Description |
|-------|-------|-------------|
| UC-TUI-01 | Display planning dashboard | Show Ready Room status: agent sessions (Captain/Commander/Design Officer), planning iteration, sign-off progress, pending questions |
| UC-TUI-02 | Display Admiral question modal | LCARS-themed modal presenting agent questions via Huh form (Select for options, Input for free-text, Confirm for broadcast toggle) with skip option |
| UC-TUI-03 | Display plan review overlay | Mission manifest rendered via Glamour markdown with approve/feedback/shelve controls via Huh form |
| UC-TUI-04 | Display execution dashboard | Show active missions, per-AC TDD phase progress, agent status, gate results, wave progress |
| UC-TUI-05 | Display agent status grid | Show all active agents with role, assigned mission, current phase, harness type, and elapsed time |
| UC-TUI-06 | Display live event log | Scrolling log of recent events via charmbracelet/log (state transitions, gate results, protocol events, agent spawns/exits) with leveled, colorized output |
| UC-TUI-07 | Accept operator commands | Huh Input field for operator commands: halt mission, retry mission, approve plan, change WIP limit, quit |
| UC-TUI-08 | Display system health | WIP utilization, agent uptime, stuck detection warnings, heartbeat status, Doctor state |
| UC-TUI-09 | Display wave execution view | Current wave, parallel missions, dependency graph, completed vs active waves |
| UC-TUI-10 | Apply LCARS theme | Lipgloss styles for LCARS color palette (orange/blue/purple), box-drawing borders, semantic status colors, Harmonica spring animations for view transitions and status changes |
| UC-TUI-11 | Implement fleet-centric navigation | Stack-based view navigation with NavigableToolbar on every view, panel focus cycling (Tab/Shift+Tab), drill-down/back stack bounded to depth 3, and global keyboard shortcuts |
| UC-TUI-12 | Display Fleet Overview screen | Landing screen with ship card grid, ship preview panel, ship count/health header, and NavigableToolbar for fleet-level navigation |
| UC-TUI-13 | Display Ship Bridge screen | Per-ship execution view with crew panel, mission board (Kanban summary), event log, inline wave summary in header, and context-sensitive NavigableToolbar |
| UC-TUI-14 | Display Fleet Monitor screen | Condensed multi-ship status grid with one ShipStatusRow per ship showing inline progress, crew health, and wave status |
| UC-TUI-15 | Display Mission Detail view | Per-mission drill-down with per-AC TDD phase pipeline (ACPhaseDetail), gate evidence, output viewport, and context-sensitive actions |
| UC-TUI-16 | Display Agent Detail view | Per-agent drill-down with real-time output stream, assignment, TDD phase, harness type, health indicators, and elapsed time |
| UC-TUI-17 | Display Specialist Detail view | Per-specialist drill-down in Ready Room showing role badge, analysis output rendered via Glamour, status, and assignment |
| UC-TUI-18 | Display Directive Editor screen | PRD input form with Huh Text field for content, ship assignment, and Glamour markdown preview |
| UC-TUI-19 | Display Message Center screen | Cross-ship Admiral question inbox showing pending questions from all ships with response interface and badge count |
| UC-TUI-20 | Display Project Settings screen | Global configuration with 4-tab interface (Gates, Crew, Fleet, Export/Import) for verification gate commands, crew defaults, fleet defaults, and settings portability |
| UC-TUI-21 | Display Help Overlay | Contextual keyboard shortcut overlay using Bubbles help + key components, showing global and view-specific shortcuts |
| UC-TUI-22 | Display Confirmation Dialog | Destructive action confirmation modal using Huh Confirm for halt, force-kill, and shelve actions with descriptive context |
| UC-TUI-23 | Implement responsive layout engine | Standard (>=120 cols) and Compact (<120 cols) layout modes with breakpoint detection on tea.WindowSizeMsg and per-element adaptations |
| UC-TUI-24 | Implement progressive disclosure modes | Basic, Advanced, and Executive display modes with mode-specific panel visibility, terminology translation, auto-detection heuristic, and runtime toggle (Ctrl+M) |
| UC-TUI-25 | Implement accessibility features | Symbol-based status indicators (never color alone), accessible Huh theme mode, contrast compliance, and full keyboard navigation coverage |
| UC-TUI-26 | Implement style token system | Centralized theme module (internal/tui/theme/lcars.go) with reference-tier LCARS colors, semantic-tier styles, and panel/overlay border definitions |
| UC-TUI-27 | Implement animation system | Harmonica spring-based animations with 9 named presets, performance budget (60fps, max 3 concurrent), and user preference controls |
| UC-TUI-28 | Implement custom component library | Seven reusable custom components: StatusBadge, PanelFrame, NavigableToolbar, PhaseIndicator, WaveBar, ShipStatusRow, ACPhaseDetail |

**Acceptance Criteria**:

- UC-TUI-01:
  - `☐` Planning dashboard shows each agent session with status indicator: `[✓]` done, `[●]` working, `[⏸]` waiting for Admiral, `[✗]` failed
  - `☐` Dashboard shows planning iteration count (e.g., "Iteration 2/5")
  - `☐` Dashboard shows sign-off progress per mission
  - `☐` Dashboard shows pending question count: `Questions: N pending`
  - `☐` Ready Room displays specialist grid with status cards for Captain, Commander, and Design Officer
  - `☐` Directive viewport toggles visibility via `[t]` shortcut, renders PRD content via Glamour
  - `☐` Panel focus order: Specialist Grid → Directive Viewport → NavigableToolbar
  - `☐` NavigableToolbar shows: `[v]` Review Plan, `[t]` Toggle Directive, `[Esc]` Bridge

- UC-TUI-02:
  - `☐` Modal identifies asking agent: "ADMIRAL — QUESTION FROM CAPTAIN"
  - `☐` Modal shows question text, context, and domain (functional/technical/design) rendered via Glamour
  - `☐` Modal uses Huh `Select` for arrow-key selection of pre-defined options
  - `☐` Modal uses Huh `Input` for free-text entry, Huh `Confirm` for broadcast toggle, with skip as a Select option
  - `☐` Huh form groups question display and input fields into a single `huh.Form` composed as `tea.Model`
  - `☐` Modal resolves the `AdmiralQuestionGate` promise on form submission
  - `☐` Modal renders with rounded border in butterscotch (#FF9966) with drop shadow effect (1-char offset in dark_blue)
  - `☐` Modal pushed onto overlay stack; captures all input until dismissed

- UC-TUI-03:
  - `☐` Plan Review renders as full-screen drill-down view (not overlay) from Ready Room
  - `☐` Mission manifest rendered as Glamour-styled markdown (mission list, sequence numbers, titles, use-case references)
  - `☐` Coverage matrix shows use-case to mission mapping (covered/partial/uncovered)
  - `☐` Dependency graph visualization shows mission dependencies
  - `☐` Multi-step Huh form: Step 1 review manifest, Step 2 select action (approve/feedback/shelve), Step 3 feedback text (conditional on "feedback" selection)
  - `☐` Feedback mode uses Huh `Text` (multi-line, 2000 char limit) for Admiral feedback input
  - `☐` NavigableToolbar shows: `[a]` Approve, `[f]` Feedback, `[s]` Shelve, `[Esc]` Ready Room

- UC-TUI-04:
  - `☐` Execution dashboard (Ship Bridge) shows per-mission progress through TDD phases
  - `☐` Each AC shows phase indicators: RED → VERIFY_RED → GREEN → VERIFY_GREEN → REFACTOR → VERIFY_REFACTOR
  - `☐` Gate pass/fail results shown inline with classification
  - `☐` Missions color-coded by status: active (butterscotch), complete (green_ok), halted (red_alert)
  - `☐` Ship Bridge is the primary execution view, entered via Enter on a ship in Fleet Overview

- UC-TUI-05:
  - `☐` Agent grid updates in real-time as agents spawn, progress, and exit via event bus subscription
  - `☐` Each agent card shows: role, mission ID, TDD phase, elapsed time, harness name
  - `☐` Stuck agents highlighted in warning color (yellow_caution) with `[~]` symbol
  - `☐` Agent grid rendered as Bubbles table with sortable columns
  - `☐` Enter on an agent navigates to Agent Detail view (drill-down)

- UC-TUI-06:
  - `☐` Event log shows last 50 events with timestamp, type, actor, and summary
  - `☐` Events color-coded by severity: info=blue, warn=yellow_caution, error=red_alert
  - `☐` Log is scrollable with keyboard navigation (PgUp/PgDn) using Bubbles viewport
  - `☐` Log shows 8-12 lines in standard layout, 4 lines in compact layout, 2 lines when height < 30 rows
  - `☐` Events emitted via charmbracelet/log structured logging with Lipgloss styling

- UC-TUI-07:
  - `☐` Command input bar at bottom of TUI uses Huh `Input` for command entry: `halt <id>`, `retry <id>`, `approve <id>`, `wip <n>`, `quit`
  - `☐` Commands provide immediate feedback in event log
  - `☐` Destructive commands (halt, force-kill) trigger Confirmation Dialog (UC-TUI-22) before execution

- UC-TUI-08:
  - `☐` WIP utilization bar (e.g., "3/5 agents active") using Bubbles progress component
  - `☐` Doctor heartbeat indicator (green_ok/red_alert) with symbol: `[+]` healthy, `[!]` failed
  - `☐` Count of stuck agents and orphaned missions

- UC-TUI-09:
  - `☐` Current wave number and parallel mission count displayed
  - `☐` Completed waves visually distinguished from active wave
  - `☐` Dependency relationships shown between related missions
  - `☐` Wave Manager accessible via `[w]` from Ship Bridge as overlay
  - `☐` Wave Manager shows dependency graph (ASCII rendering) and merge controls

- UC-TUI-10:
  - `☐` LCARS color palette applied via Lipgloss styles from centralized theme module (see UC-TUI-26)
  - `☐` Box-drawing borders using Lipgloss border styles (NormalBorder, DoubleBorder, RoundedBorder)
  - `☐` Semantic color mapping: active=butterscotch, success=green_ok, error=red_alert, warning=yellow_caution, info=blue, planning=purple
  - `☐` Harmonica spring animations for view transitions (see UC-TUI-27 for full animation system)
  - `☐` Lipgloss `AdaptiveColor` used for light/dark terminal detection
  - `☐` Lipgloss `ColorProfile()` used to detect ANSI/256/TrueColor and degrade gracefully

- UC-TUI-11:
  - `☐` System implements a navigation stack (bounded to depth 3) supporting push (Enter) and pop (Esc) operations
  - `☐` Every view displays a NavigableToolbar at the bottom showing context-sensitive shortcut buttons
  - `☐` Operator can navigate toolbar buttons via Left/Right arrows and activate via Enter, or press quick key directly
  - `☐` Tab/Shift+Tab cycles focus between panels within the current view in a fixed order
  - `☐` Focused panel displays double-line border in moonlit_violet (#9966FF) with bold title; unfocused panels show single-line border in galaxy_gray (#52526A)
  - `☐` Active overlay displays rounded border in butterscotch (#FF9966) with drop shadow effect
  - `☐` Global shortcuts work from any view: `?` (help), `q`/`Ctrl+C` (quit with confirmation if missions active), `Tab`/`Shift+Tab` (panel cycle), `Enter` (select/drill-down), `Esc` (back/dismiss)
  - `☐` Attempting to push beyond stack depth 3 replaces the deepest entry

- UC-TUI-12:
  - `☐` Fleet Overview displays as the initial screen when TUI launches via `sc3 tui`
  - `☐` Ship cards show name, directive title, status badge (StatusBadge component), and progress bar in a scrollable list
  - `☐` Header shows ship count, active/complete summary, and fleet health status
  - `☐` Selected ship shows preview panel with crew roster, mission summary, and wave progress
  - `☐` In standard layout (>=120 cols), ship list (40%) and preview panel (60%) display side-by-side via lipgloss.JoinHorizontal
  - `☐` In compact layout (<120 cols), ship list displays full width; preview hidden or stacked below
  - `☐` NavigableToolbar shows: `[Enter]` Bridge, `[n]` New, `[f]` Monitor, `[a]` Roster, `[i]` Inbox, `[s]` Settings
  - `☐` Panel focus order: Ship List → Ship Preview → NavigableToolbar

- UC-TUI-13:
  - `☐` Ship Bridge displays when operator selects a ship from Fleet Overview via Enter
  - `☐` Header shows ship name, directive, health status, and inline wave summary: `Wave K of L [====....] M/T`
  - `☐` Crew Panel shows agents via Bubbles table with columns: role, mission, phase, elapsed time
  - `☐` Mission Board shows Kanban summary line: `B:N IP:N R:N D:N H:N` with color-coded column keys
  - `☐` Event Log shows scrollable recent events with severity coding (8-12 lines standard, 4 lines compact)
  - `☐` In standard layout, Crew Panel (40%) and Mission Board (60%) side-by-side with Event Log below
  - `☐` In compact layout, panels stack vertically: Crew → Board summary → Log
  - `☐` NavigableToolbar shows: `[r]` Ready Room, `[w]` Waves, `[h]` Halt, `[Space]` Pause, `[Esc]` Fleet
  - `☐` Panel focus order: Crew Panel → Mission Board → Event Log → NavigableToolbar
  - `☐` Enter on agent in Crew Panel navigates to Agent Detail; Enter on mission navigates to Mission Detail

- UC-TUI-14:
  - `☐` Fleet Monitor displays when operator presses `[f]` from Fleet Overview
  - `☐` Each ship renders as a single ShipStatusRow with name, directive, progress bar, crew health, and wave status
  - `☐` Operator can select a ship and press Enter to navigate to its Ship Bridge
  - `☐` NavigableToolbar shows: `[Enter]` Bridge, `[i]` Inbox, `[Esc]` Fleet

- UC-TUI-15:
  - `☐` Mission Detail displays when operator selects a mission from the Mission Board via Enter
  - `☐` ACPhaseDetail component shows scrollable list of ACs with 6-phase TDD pipeline: `RED > V_RED > GRN > V_GRN > REF > V_REF`
  - `☐` Each AC phase shows status symbol (`[!]` fail expected, `[>]` active, `[+]` passed, `[.]` pending) with semantic color
  - `☐` Gate evidence displays inline per phase: gate type, exit code, classification (accept/reject_vanity/reject_syntax/reject_failure), output snippet, timestamp, attempt number
  - `☐` Output viewport shows agent output for selected AC via scrollable Bubbles viewport
  - `☐` NavigableToolbar shows context-sensitive actions based on mission state (halt/retry/approve)

- UC-TUI-16:
  - `☐` Agent Detail displays when operator selects an agent from Crew Panel via Enter
  - `☐` View shows agent role, assigned mission, current TDD phase, harness type, and elapsed time via Bubbles timer
  - `☐` Output viewport shows real-time agent stdout/stderr captured via tmux capture-pane and propagated through event bus
  - `☐` Health indicators show heartbeat status and stuck detection (symbol + color)
  - `☐` NavigableToolbar shows: `[Esc]` to return to Ship Bridge

- UC-TUI-17:
  - `☐` Specialist Detail displays when operator selects a specialist from Ready Room grid via Enter
  - `☐` View shows specialist role badge (Captain/Commander/Design Officer), status indicator, and current assignment
  - `☐` Output viewport shows specialist analysis output rendered via Glamour markdown
  - `☐` NavigableToolbar shows: `[Esc]` to return to Ready Room

- UC-TUI-18:
  - `☐` Directive Editor displays when operator presses `[n]` from Fleet Overview
  - `☐` Operator can input or paste PRD content via Huh Text field
  - `☐` Operator can assign the directive to a new ship name via Huh Input
  - `☐` PRD content previews as rendered Glamour markdown in side panel
  - `☐` On submit, creates new commission and navigates to Ready Room

- UC-TUI-19:
  - `☐` Message Center displays when operator presses `[i]` from any view with NavigableToolbar
  - `☐` View shows list of pending Admiral questions across all ships with agent name, domain, and timestamp
  - `☐` Operator can select a question and respond using the Admiral Question Modal (UC-TUI-02)
  - `☐` Answered questions move to resolved section with answer summary
  - `☐` Badge count indicator displays in NavigableToolbar when pending questions exist

- UC-TUI-20:
  - `☐` Project Settings displays when operator presses `[s]` from Fleet Overview
  - `☐` View shows 4 tabs navigable via `[1]`, `[2]`, `[3]`, `[4]` keys: Gates, Crew, Fleet, Export
  - `☐` Gates tab displays configured gate commands (test_command, typecheck_command, lint_command, build_command) with Edit and Remove actions
  - `☐` Gates tab supports adding new gate commands; commands support variable substitution: `{test_file}`, `{worktree_dir}`, `{mission_id}`
  - `☐` Crew tab displays and edits crew defaults: default harness, model, WIP limits, timeouts
  - `☐` Fleet tab displays and edits fleet defaults: naming convention, wave strategy, merge policy
  - `☐` Export tab supports exporting all settings to JSON file and importing from JSON file
  - `☐` Settings persist to `~/.sc3/config.json` (global, not per-project)
  - `☐` NavigableToolbar shows: `[1]` Gates, `[2]` Crew, `[3]` Fleet, `[4]` Export, `[Esc]` Fleet

- UC-TUI-21:
  - `☐` Help Overlay displays when operator presses `?` from any screen
  - `☐` Overlay shows keyboard shortcuts relevant to the current view using Bubbles `help` + `key` components
  - `☐` Global shortcuts always visible at top; view-specific shortcuts shown below
  - `☐` Overlay dismisses with `Esc` or `?` key
  - `☐` Overlay renders with double border in blue (#9999CC)

- UC-TUI-22:
  - `☐` Confirmation Dialog displays before executing any destructive action (halt mission, force-kill agent, shelve plan)
  - `☐` Dialog uses Huh `Confirm` with descriptive title (e.g., "Halt mission MISSION-42?"), context description, and affirmative/negative options
  - `☐` Dialog renders with double border in yellow_caution (#FFCC00)
  - `☐` Escape cancels the action; affirmative button executes it
  - `☐` Dialog pushed onto overlay stack above current view

- UC-TUI-23:
  - `☐` System detects terminal dimensions at startup and on every `tea.WindowSizeMsg`
  - `☐` Standard mode (>=120 cols, >=30 rows) renders side-by-side panel layouts with full detail
  - `☐` Compact mode (<120 cols) renders stacked vertical layouts with abbreviated labels
  - `☐` Compact mode adaptations: Crew + Mission Board stacked (not side-by-side), Event Log reduced to 4 lines, wave summary abbreviated to `W2/3`, NavigableToolbar labels shortened (e.g., `[r]Ready`)
  - `☐` Height adaptation (<30 rows): Event Log reduces to 2 lines, Crew Panel shows max 3 agents (scrollable), Ship Preview hidden in Fleet Overview
  - `☐` Layout transitions use Harmonica spring animation for smooth panel reflow (frequency 4.0, damping 1.2)
  - `☐` Root model stores `width` and `height` and passes layout variant to all child view render functions

- UC-TUI-24:
  - `☐` System supports three display modes: Basic (low density, simplified terminology), Advanced (full detail, LCARS terminology), Executive (aggregated metrics)
  - `☐` Mode auto-detected on first launch: no `.sc3/` directory = Basic; existing `.sc3/` = Advanced; `--executive`/`--basic`/`--advanced` flags override
  - `☐` Runtime mode toggle via `Ctrl+M` cycles Basic → Advanced → Executive → Basic; also settable via command bar: `mode basic`, `mode advanced`, `mode executive`
  - `☐` Basic mode uses simplified terminology: Task (not Mission), Batch (not Wave), Helper (not Agent), Check (not Gate), Planning (not Ready Room), Dashboard (not Ship Bridge)
  - `☐` Basic mode hides: inter-agent message log, dependency graph, wave summary detail, gate details; shows simplified "Active Tasks" list instead of Kanban board
  - `☐` Executive mode shows fleet-level aggregates: ship progress percentages, velocity metrics (missions/day with trend), blocker summary (actionable items only); hides individual agent details, gate details, event logs
  - `☐` Advanced mode shows all panels with full LCARS terminology and complete detail
  - `☐` Mode-specific panel visibility rules applied on every `View()` render cycle; mode stored in root model and passed to all view render functions
  - `☐` Configuration via `[tui] mode` in TOML config file persists mode preference across sessions

- UC-TUI-25:
  - `☐` All status indicators use consistent symbol + color vocabulary: `[>]` running/butterscotch, `[+]` success/green_ok, `[!]` failed/red_alert, `[~]` warning/yellow_caution, `[-]` info/blue, `[?]` review/purple, `[.]` waiting/light_gray, `[x]` skipped/galaxy_gray
  - `☐` TDD phase indicators combine phase name text, status symbol, and color for triple-redundancy
  - `☐` Accessible mode activates via `SC3_ACCESSIBLE` env var, `--accessible` CLI flag, or `[tui] accessible = true` config
  - `☐` In accessible mode: Huh forms use `huh.ThemeBase()` with high-contrast overrides, focus indicators use reverse video (not color-only), form validation errors prefixed with `ERROR:` text
  - `☐` galaxy_gray (#52526A) used only for decorative borders and disabled elements (never essential text); disabled elements include `(disabled)` text label
  - `☐` All interactive elements reachable via keyboard without mouse; mouse support optional enhancement only
  - `☐` Terminal bell (`\a`) emitted on critical events (mission halted, gate failure requiring Admiral); suppressible via `[tui] bell = false`

- UC-TUI-26:
  - `☐` Theme module defined at `internal/tui/theme/lcars.go` with exported color constants and semantic styles
  - `☐` Reference tier defines 16 LCARS color values: Butterscotch (#FF9966), Blue (#9999CC), Purple (#CC99CC), Pink (#FF99CC), Gold (#FFAA00), Almond (#FFAA90), RedAlert (#FF3333), YellowCaution (#FFCC00), GreenOk (#33FF33), Black (#000000), DarkBlue (#1B4F8F), GalaxyGray (#52526A), SpaceWhite (#F5F6FA), LightGray (#CCCCCC), MoonlitViolet (#9966FF), Ice (#99CCFF)
  - `☐` Semantic tier defines meaning-assigned Lipgloss styles: ActiveStyle, SuccessStyle, ErrorStyle, WarningStyle, InfoStyle, PlanningStyle, NotifyStyle, FocusStyle
  - `☐` Panel border styles defined: PanelBorder (NormalBorder, GalaxyGray), PanelBorderFocused (DoubleBorder, MoonlitViolet), OverlayBorder (RoundedBorder, Butterscotch)
  - `☐` HeaderStyle defined with Bold, SpaceWhite foreground, DarkBlue background, Padding(0,1)
  - `☐` No hardcoded colors or styles in component code; all styles reference theme module
  - `☐` `lipgloss.AdaptiveColor` used for light/dark terminal detection where appropriate

- UC-TUI-27:
  - `☐` Animation system uses Harmonica springs driven by `tea.Tick` at ~60fps during active animations, 0fps when idle
  - `☐` System implements 9 named animations: Modal Open (freq 6.0, damp 1.0), Modal Close (freq 6.0, damp 1.0), View Transition (freq 5.0, damp 1.0), Red Alert Pulse (freq 3.0, damp 0.4), Panel Resize (freq 4.0, damp 1.2), Progress Bar Update (freq 3.0, damp 1.5), Notification Badge (freq 8.0, damp 0.3), Phase Advance (freq 5.0, damp 0.8), Success Checkmark (freq 6.0, damp 0.6)
  - `☐` Maximum 3 concurrent animations enforced; additional animations queued
  - `☐` Animations disabled when terminal width < 80 cols
  - `☐` User preference `[tui] animations = false` disables all animations (instant transitions)
  - `☐` User preference `[tui] animation_speed` supports `"fast"` (1.5x frequency) and `"slow"` (0.7x frequency)
  - `☐` Executive mode disables animations by default
  - `☐` Animation state is cheap (few floats per spring); render cost is in the `View()` pass

- UC-TUI-28:
  - `☐` **StatusBadge**: Renders `[symbol] LABEL` with semantic color; used in 8+ locations (Crew Panel, Mission Board, Ready Room, Fleet Monitor, Agent Roster)
  - `☐` **PanelFrame**: Renders Lipgloss-styled border with title, focus state (unfocused/focused/active overlay), and optional subtitle; used in 10+ locations (every panel in every view)
  - `☐` **NavigableToolbar**: Renders shortcut buttons supporting both quick-key activation and arrow-key (Left/Right) + Enter navigation; used in every view (14 instances)
  - `☐` **PhaseIndicator**: Renders 6-phase TDD pipeline as horizontal strip `RED > V_RED > GRN > V_GRN > REF > V_REF` with current phase highlighted and status symbols
  - `☐` **WaveBar**: Renders `Wave N [====....] M/T` with progress fill using Bubbles progress component internally
  - `☐` **ShipStatusRow**: Renders single-line ship representation with inline progress, crew health, and wave status for Fleet Monitor
  - `☐` **ACPhaseDetail**: Renders scrollable list of ACs with 6-dot phase pipeline, inline gate results, and attempt counts for Mission Detail
  - `☐` All custom components follow dumb-component pattern: receive data as arguments, return styled strings, hold no state, emit no commands

---

### OBS -- Observability

**Prefix**: OBS | **Purpose**: Distributed tracing, structured logging, metrics, and debuggability for both deterministic logic and LLM calls

Ship Commander 3 provides comprehensive observability via OpenTelemetry tracing and structured JSON logging. Every run produces a complete trace correlated with logs, enabling quick diagnosis of failures: WHERE something broke (which pipeline step, which LLM call) and WHY (error type, exit code, bounded output). Debug bundles capture all context needed for reproduction without interactive access.

| UC ID | Title | Description |
|-------|-------|-------------|
| UC-OBS-01 | Initialize telemetry on startup | Create tracer provider with OTLP exporter (configurable via env vars), resource attributes (service.name=ship-commander-3, service.version, environment), batch span processor, and shutdown handler. Return shutdown func for graceful cleanup. |
| UC-OBS-02 | Emit root span for every run | Each `sc3` command creates a root span with `run_id` (UUID) and `trace_id` (correlates all child spans). Attributes: command name, arguments (redacted), working directory, Git HEAD, environment. |
| UC-OBS-03 | Trace deterministic pipeline steps | Emit spans for each major operation: state transitions (from_state, to_state, reason), tool execution (tool name, args redacted, cwd, duration, exit_code, bounded stdout/stderr), Git operations (checkout, worktree, apply_patch, diff summary), test runs (command, duration, pass/fail, failing snippet). |
| UC-OBS-04 | Trace LLM/harness calls | Emit spans for each agent session: model name, latency (ms), token usage (prompt/response if available), prompt version/hash, tool calls count, harness type (claude/codex). Redact secrets in prompts/tool args. |
| UC-OBS-05 | Structured logging to JSON file | Never log to stdout while TUI is active (corrupts terminal). Write JSON logs to file with run_id, trace_id, timestamp, level, component, message, structured fields. Rotate logs by size (default 10MB) with retention (default 5 files). |
| UC-OBS-06 | Correlate logs with traces | Every log record includes `run_id` and `trace_id` attributes. Trace span IDs referenced in logs enable jumping from log line to span and vice versa. |
| UC-OBS-07 | Detect invariant violations | Emit telemetry events (not panics) when system invariants violated: patch didn't apply cleanly, repo dirty before merge, retries exceeded, edited outside allowed paths, unexpected state transition. Event includes context (what, where, why, stack trace if applicable). |
| UC-OBS-08 | Support debug mode | `--debug` flag enables console exporter (spans to stderr) and verbose logging (DEBUG level). Useful for development and troubleshooting without full OTel setup. |
| UC-OBS-09 | Override OTel endpoint | `--otel-endpoint` flag or `OTEL_EXPORTER_OTLP_ENDPOINT` env var to target remote collector. Defaults to `http://localhost:4318` (standard OTLP HTTP port). |
| UC-OBS-10 | Generate debug bundle | `sc3 bugreport` command collects: last N log files (configurable, default 3), redacted config (secrets masked), version/build info, last trace_id/run_id, Git diff/HEAD, failing test output if present. Outputs tarball (`.sc3-bugreport-<timestamp>.tar.gz`). |
| UC-OBS-11 | Phase 1: Traces + structured logs | Implement UC-OBS-01 through UC-OBS-10 with focus on traces and structured logging. Metrics and GenAI semantic conventions deferred to Phase 2. |

**Acceptance Criteria**:

- UC-OBS-01:
  - `☐` Telemetry package at `internal/telemetry` with `Init(ctx) -> shutdown func`
  - `☐` Resource attributes: service.name, service.version, environment (dev/prod)
  - `☐` Tracer provider configured with batch span processor (default batch size: 512, timeout: 5s)
  - `☐` OTLP exporter over HTTP (protobuf), endpoint configurable via `OTEL_EXPORTER_OTLP_ENDPOINT`
  - `☐` Shutdown func flushes pending spans and closes connections

- UC-OBS-02:
  - `☐` Root span created on every `sc3` command execution (except `bugreport` itself)
  - `☐` Root span includes `run_id` (UUID v4) and `trace_id` attributes
  - `☐` Root span attributes: command.name, command.args (redacted secrets), working_dir, git.head, git.branch, environment (dev/prod/test)
  - `☐` Root span status: OK on success, ERROR on failure with error message

- UC-OBS-03:
  - `☐` State transition span: `state.transition` with attributes from_state, to_state, reason, entity_type (commission/mission/ac/agent)
  - `☐` Tool execution span: `tool.exec` with attributes tool_name, args_redacted (secrets masked), cwd, duration_ms, exit_code, stdout_preview (first 1KB), stderr_preview (first 1KB)
  - `☐` Git operation span: `git.op` with attributes operation (checkout/worktree/apply_patch/diff), target (branch/path), duration_ms, success (boolean), changed_files (count)
  - `☐` Test run span: `tests.run` with attributes command, duration_ms, passed (boolean), failed_count, error_summary (first 500 chars of failure)

- UC-OBS-04:
  - `☐` LLM span: `llm.call` with attributes model_name, harness (claude/codex), latency_ms, prompt_tokens, response_tokens (if available), total_tokens, prompt_hash (SHA-256 of prompt template), tool_calls_count
  - `☐` Secrets redacted from span attributes: API keys, passwords, tokens, sensitive paths
  - `☐` Span events for tool calls within LLM session: tool_name, duration_ms, success (boolean)
  - `☐` Error tracking: LLM failures recorded as span events with error_type, error_message, retry_count

- UC-OBS-05:
  - `☐` JSON logs written to `.sc3/logs/sc3-<date>-<run_id>.json`
  - `☐` Log record schema: timestamp (ISO 8601), level (DEBUG/INFO/WARN/ERROR), run_id, trace_id, span_id, component, message, structured fields (key-value pairs)
  - `☐` No stdout logging while TUI active (corrupts terminal display)
  - `☐` Log rotation at 10MB, retain 5 files (configurable via `[logging] max_size_mb`, `max_files`)
  - `☐` Debug mode (`--debug`) allows console logging (stderr) for non-TUI commands

- UC-OBS-06:
  - `☐` Every log record includes `run_id` and `trace_id` fields
  - `☐` Active span context propagated via `trace_id` and `span_id` fields
  - `☐` Log viewer tool (optional): `sc3 logs --run-id <uuid>` filters logs by run
  - `☐` Trace viewer integration: logs link to trace in OTel backend (Jaeger/Tempo)

- UC-OBS-07:
  - `☐` Invariant violations emit telemetry events (not panic): `invariant.violation` with attributes invariant_name, severity (warn/error), context (structured details)
  - `☐` Pre-defined invariants: patch_apply_clean (patch applied without fuzz), repo_clean_before_merge (no uncommitted changes), max_retries_not_exceeded, edits_within_allowed_paths, state_transition_legal
  - `☐` Event includes: what_invariant, where_detected (component/line), why_violated (context), stack_trace (if applicable)
  - `☐` Invariant checks enabled by default; can be disabled via `--skip-invariant-checks` (emergency only)

- UC-OBS-08:
  - `☐` `--debug` flag enables: console exporter (spans to stderr in human-readable format), verbose logging (DEBUG level), span events printed to console
  - `☐` Debug mode useful for: development without collector, troubleshooting in CI/CD, quick diagnostics
  - `☐` Debug mode clearly indicated in logs: `"logging": "DEBUG", "otel_exporter": "console"`
  - `☐` Console span format: `[SPAN] span_name {duration} {status}` with indented child spans

- UC-OBS-09:
  - `☐` OTel endpoint configurable via: `--otel-endpoint` CLI flag, `OTEL_EXPORTER_OTLP_ENDPOINT` env var, `[otel] endpoint` in config
  - `☐` Default: `http://localhost:4318` (OTLP HTTP)
  - `☐` Falls back to console exporter if endpoint unreachable (with warning)
  - `☐` Supports TLS: `https://` for remote collectors, cert verification via `OTEL_EXPORTER_OTLP_CERTIFICATE`

- UC-OBS-10:
  - `☐` `sc3 bugreport` command generates: tarball with logs (last N files, default 3), config (redacted secrets), version info (`sc3 --version` output), last run_id/trace_id, Git state (HEAD, diff, uncommitted changes), failing test output (if last run failed)
  - `☐` Secrets redaction pattern: API keys, tokens, passwords, credentials masked as `***REDACTED***`
  - `☐` Tarball named `.sc3-bugreport-<timestamp>.tar.gz`, includes `README.txt` with collection summary
  - `☐` Command exits with path to bugreport and instructions: `"Bug report written to: /path/to/file. Share for debugging."`

- UC-OBS-11:
  - `☐` Phase 1 (this PRD) implements: traces (deterministic + LLM), structured JSON logging, debug bundles, basic invariants
  - `☐` Phase 2 deferred: metrics (counters, gauges, histograms), GenAI semantic conventions (gen_ai.prompt, gen_ai.completion, etc.), advanced invariants (performance SLAs, resource limits)
  - `☐` Phase 1 acceptance: all runs produce traces, failure diagnosis via trace + logs, debug bundle sufficient for reproduction

---

## UI Design Requirements

> **Reference**: Full UI/UX specifications are defined in `design/UX-DESIGN-PLAN.md` and its companion YAML artifacts. The UX Design Plan is the authoritative source for visual layout, interaction patterns, and component behavior. All TUI use cases (UC-TUI-*) must conform to the specifications therein.

### Design Artifact Reference

| Artifact | Path | Contents |
|----------|------|----------|
| **UX Design Plan** | `design/UX-DESIGN-PLAN.md` | Navigation model, responsive behavior, accessibility, component strategy, animation strategy, progressive disclosure |
| **Views** | `design/views.yaml` | Per-view layout, panels, data sources, keyboard shortcuts, responsive behavior |
| **Screens** | `design/screens.yaml` | Navigation hierarchy, screen routing, transitions, animation presets |
| **Components** | `design/components.yaml` | Component library (35 components), Charm library mappings, composition patterns |
| **Flows** | `design/flows.yaml` | Micro-level interaction flows (16 flows), step-by-step user interactions |
| **Workflows** | `design/workflows.yaml` | End-to-end user workflows (8 workflows), multi-screen journeys |
| **Paradigm** | `design/paradigm.yaml` | Design patterns, principles, animation categories, terminology glossary |
| **Config** | `design/config.yaml` | Color palette, technology stack, keyboard plan, display modes |
| **Mockups** | `design/mocks/*.mock.html` | HTML visual mockups for all screens/overlays |

### Binding Design Principles

The following principles from the UX Design Plan are **binding** on all TUI implementation:

1. **P1: Flows Before Screens** -- Every screen emerges from a user workflow, not imagination
2. **P2: Deterministic Feedback, Always** -- Every gate result and state transition drawn from deterministic data; no ambiguous states
3. **P3: Symbols + Color, Never Color Alone** -- Every status communicated through symbol AND color (see UC-TUI-25)
4. **P4: Keyboard-First, Always** -- Every interaction reachable via keyboard; mouse optional (see UC-TUI-11)
5. **P5: Progressive Disclosure Over Information Overload** -- Basic/Advanced/Executive modes control density (see UC-TUI-24)
6. **P6: Smart Main, Dumb Components** -- Single root `tea.Model` owns all state; child views are pure render functions (see Appendix B in UX plan)
7. **P7: Charm Aesthetic Meets LCARS Soul** -- Fusion of Charmbracelet playfulness with LCARS thematic elements (see UC-TUI-26)

### Screen Inventory (16 Screens)

| Screen | Type | Priority | Entry Point | UX Plan Reference |
|--------|------|----------|-------------|-------------------|
| Fleet Overview | primary/landing | high | `sc3 tui` | Section 2, 4 |
| Ship Bridge | drill-down | high | Enter on ship in Fleet | Section 2, 4 |
| Ready Room | drill-down | high | `[r]` from Ship Bridge | Section 2 |
| Plan Review | drill-down | high | `[v]` from Ready Room | Section 2 |
| Fleet Monitor | drill-down | medium | `[f]` from Fleet Overview | Section 2 |
| Mission Detail | drill-down | high | Enter on mission | Section 2 |
| Agent Detail | drill-down | medium | Enter on agent | Section 2 |
| Specialist Detail | drill-down | medium | Enter on specialist | Section 2 |
| Agent Roster | top-level | medium | `[a]` from Fleet Overview | Section 2 |
| Directive Editor | top-level | medium | `[n]` from Fleet Overview | Section 2 |
| Message Center | top-level | medium | `[i]` from any view | Section 2 |
| Wave Manager | overlay | low | `[w]` from Ship Bridge | Section 2 |
| Project Settings | top-level | medium | `[s]` from Fleet Overview | Section 9 |
| Admiral Question | modal | high | Auto on agent question | Section 6 |
| Help Overlay | overlay | medium | `?` from any screen | Section 5 |
| Confirm Dialog | modal | medium | Destructive actions | Section 6 |

### Charm Library Dependencies

| Library | Min Version | Role | Import |
|---------|-------------|------|--------|
| Bubble Tea | >=1.2.0 | Elm architecture TUI framework | `github.com/charmbracelet/bubbletea` |
| Bubbles | >=0.20.0 | Pre-built components (spinner, progress, table, list, viewport, etc.) | `github.com/charmbracelet/bubbles` |
| Lipgloss | >=1.0.0 | Styling, borders, layout composition, color system | `github.com/charmbracelet/lipgloss` |
| Huh | >=0.6.0 | Terminal forms for all Admiral interactions | `github.com/charmbracelet/huh` |
| Glamour | >=0.8.0 | Markdown rendering (PRD, specs, demo tokens, gate output) | `github.com/charmbracelet/glamour` |
| Harmonica | >=0.2.0 | Spring-based physics animations | `github.com/charmbracelet/harmonica` |
| Log | >=0.4.0 | Structured logging with Lipgloss styling | `github.com/charmbracelet/log` |

### Cross-Cutting Implementation Notes

The UX Design Plan introduces several cross-cutting concerns that affect multiple PRD functional groups:

1. **Event Bus → TUI Bridge** (STATE + TUI): The Beads state layer publishes typed events through Go channels. The TUI root model subscribes and translates events into `tea.Msg` types. This bridges UC-STATE-09 (event log persistence) with UC-TUI-06 (live event log display) and UC-TUI-05 (agent status updates).

2. **Project Settings → Gate Config** (TUI + GATE): The Project Settings view (UC-TUI-20) provides the UI for configuring verification gate commands (UC-GATE-08). Gate commands are language-agnostic bash strings stored globally at `~/.sc3/config.json`.

3. **Progressive Disclosure → All Views** (TUI-wide): Display mode (UC-TUI-24) affects panel visibility and terminology in every view. All view render functions must accept a `mode` parameter and adapt accordingly.

4. **Responsive Layout → All Views** (TUI-wide): Layout mode (UC-TUI-23) affects every view's panel composition. All views must support both standard (side-by-side) and compact (stacked) layouts.

5. **Style Token System → All Components** (TUI-wide): The centralized theme (UC-TUI-26) is a dependency for every visual component. No component should define its own colors or border styles.

---

## Use Case Summary

| Group | Prefix | Use Cases | Description |
|-------|--------|-----------|-------------|
| Commission & Planning | COMM | 10 | Ready Room, Admiral approval, agent questions, plan persistence |
| Mission Execution | EXEC | 11 | Commander orchestrator, TDD cycle, termination, demo tokens, wave review |
| Verification Gates | GATE | 8 | Independent gate execution, protocol events |
| Harness & Sessions | HARN | 8 | Claude Code, Codex, tmux session management |
| State Persistence | STATE | 10 | Beads integration, state machine, locks, recovery |
| Observability | OBS | 11 | Distributed tracing, structured logging, LLM observability, debug bundles |
| Terminal Interface | TUI | 28 | Navigation system, 16 screens/overlays, responsive layout, progressive disclosure, accessibility, style tokens, animations, custom components, Huh forms, LCARS theme |
| **Total** | | **86** | |

---

## Constraints

| Constraint | Value | Rationale |
|-----------|-------|-----------|
| Core language | Go (Golang) | Performance, single binary distribution, goroutine concurrency |
| TUI framework | Bubble Tea + Lipgloss + Glamour + Huh + Harmonica | Elm architecture TUI (Bubble Tea), styling (Lipgloss), markdown rendering (Glamour -- Glow's embeddable engine), terminal forms (Huh -- native `tea.Model` for structured input), physics-based animations (Harmonica -- spring/damping for smooth TUI transitions) |
| Logging | charmbracelet/log | Minimal, colorful, leveled structured logging with Lipgloss styling. Slog handler compatible. Used for ALL application logging (orchestrator, gates, harness, state, TUI events). |
| Observability | OpenTelemetry Go SDK + OTLP | Distributed tracing, structured logging, debuggability for deterministic and LLM operations |
| Tracing backend | OTLP over HTTP/4318 | Standard OpenTelemetry protocol, compatible with Jaeger/Tempo collectors |
| Log format | JSON to file | Machine-parseable, correlates with traces via run_id/trace_id, no stdout while TUI active |
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
| **Ship** | Persistent agent team/pool that serves multiple commissions over time. Reusable grouping of specialized agents. | Team / Squad |
| **Mission** | Atomic coding task that goes through TDD cycle or implementation. Produces verifiable output. | Task / Issue |
| **Ready Room** | Collaborative planning loop where Captain, Commander, and Design Officer decompose a commission into missions. | Sprint planning meeting |
| **Mission Manifest** | The agreed-upon, sequenced list of missions produced by the Ready Room. Requires Admiral approval. | Sprint backlog |
| **Demo Token** | Verifiable proof artifact (`demo/MISSION-<id>.md`) that a human can review without reading all code. | Acceptance test evidence |
| **Gate** | Deterministic verification (shell command + exit code) run by Commander independently of agent. | CI/CD check |
| **Protocol Event** | Structured JSON event for agent ↔ Commander coordination. Versioned, validated, persisted. | Message queue event |
| **Surface-Area Lock** | File-path lock acquired by Commander before dispatch to prevent concurrent modification. | Database row lock |
| **Wave** | Group of missions with no inter-dependencies that execute in parallel. | Sprint batch |
| **Propulsion** | Commander's continuous loop: check backlog → dispatch → verify → advance. | CI/CD pipeline |
| **Trace** | Distributed trace (OpenTelemetry) correlating all operations in a run via run_id and trace_id. Spans cover deterministic steps (state transitions, tools, Git, tests) and LLM calls (model, latency, tokens). | Distributed tracing |
| **Debug Bundle** | Tarball (`sc3 bugreport`) containing logs, config, version info, Git state, and test output for offline diagnosis. | Incident response package |

---

## Implementation Roadmap

### Phase 1: Foundation + Observability Setup (Week 1-2)

1. **Go project scaffold** -- CLI entry point, TOML config, Beads wrapper, state machine, charmbracelet/log structured logging throughout
2. **Commission type + parser** -- Parse PRD Markdown into Commission struct with use cases
3. **Captain's Ready Room** -- Planning loop with harness session spawning, message routing, consensus validation
4. **Admiral approval gate** -- TUI-based plan review with approve/feedback/shelve
5. **Agent → Admiral question gate** -- Question modal with options, free text, broadcast
6. **[OBSERVABILITY] Telemetry package** -- OpenTelemetry Go SDK setup, tracer provider, OTLP exporter, resource attributes, shutdown handler (UC-OBS-01)
7. **[OBSERVABILITY] Root spans** -- Every command creates run span with run_id/trace_id, command name, args redacted, working dir, Git HEAD (UC-OBS-02)

**Deliverable**: Commission planning loop works end-to-end with Admiral approval + basic telemetry infrastructure.

### Phase 2: Commander Execution + Pipeline Tracing (Week 3-4)

1. **Commander orchestrator** -- Mission dispatch, per-AC TDD cycle, state transitions
2. **Verification gate engine** -- VERIFY_RED, VERIFY_GREEN, VERIFY_REFACTOR, VERIFY_IMPLEMENT
3. **Protocol event system** -- JSON events, schema validation, Beads persistence
4. **Demo token validator** -- Markdown schema enforcement for proof artifacts
5. **Termination enforcer** -- Max revisions, missing demo token, AC exhaustion
6. **[OBSERVABILITY] Deterministic tracing** -- Spans for state transitions, tool execution, Git ops, test runs (UC-OBS-03)
7. **[OBSERVABILITY] Structured logging** -- JSON logs to file, run_id/trace_id correlation, log rotation (UC-OBS-05, UC-OBS-06)

**Deliverable**: Full mission execution with Commander-owned gates and termination + complete pipeline traces.

### Phase 3: Harness & TUI + LLM Observability (Week 5-6)

1. **Claude Code driver** -- Spawn, capture output, timeout enforcement via tmux
2. **Codex driver** -- Same interface, Codex-specific CLI flags
3. **Bubble Tea TUI** -- Planning dashboard, execution dashboard, event log
4. **Glamour markdown rendering** -- PRD content, mission specs, demo tokens, gate output rendered as styled terminal markdown
5. **Huh terminal forms** -- Admiral question modal (Select, Input, Confirm), approval gate (Select + Text), operator command bar (Input + Confirm)
6. **LCARS theme** -- Lipgloss styles, color palette, box-drawing borders, Harmonica spring animations for transitions
7. **[OBSERVABILITY] LLM tracing** -- Spans for agent sessions: model name, latency, token usage, prompt hash, tool calls, secret redaction (UC-OBS-04)
8. **[OBSERVABILITY] Debug mode** -- `--debug` flag for console exporter + verbose logging (UC-OBS-08)

**Deliverable**: Working TUI with real harness integration + full observability coverage (deterministic + LLM).

### Phase 4: Advanced & Hardening + Observability Polish (Week 7-8)

1. **Surface-area locking** -- Lock acquisition, conflict detection, release
2. **Standard Ops fast path** -- Simplified execution for low-risk missions
3. **Doctor process** -- Health monitoring, stuck detection, orphan recovery
4. **Crash recovery** -- Beads-based state reconstruction on restart
5. **Integration testing** -- Full commission → planning → execution → completion flow
6. **[OBSERVABILITY] Invariant violations** -- Detect and emit events for: patch failures, repo dirty, retries exceeded, illegal edits (UC-OBS-07)
7. **[OBSERVABILITY] Debug bundle** -- `sc3 bugreport` command: logs, config, version, Git state, test output (UC-OBS-10)
8. **[OBSERVABILITY] OTel endpoint config** -- `--otel-endpoint` flag, env var support, fallback to console (UC-OBS-09)

**Deliverable**: Production-ready manifesto-aligned system with complete observability: traces + logs + debug bundles.

### Observability Phase 2 (Future, Post-V1)

**Deferred to future release**:
- Metrics: Counters, gauges, histograms for operations, latency, resource usage
- GenAI semantic conventions: gen_ai.prompt, gen_ai.completion, gen_ai.tool.name per OpenTelemetry spec
- Advanced invariants: Performance SLAs, resource limit checks, anomaly detection
- Metrics dashboards: Grafana integration, alerting rules

---

## Success Criteria

The system is manifesto-aligned when:

1. **Admiral, Captain, Commander, and Design Officer are distinct** -- roles are not collapsed
2. **Commission→Mission traceability** -- every mission links to a commission and the use case(s) it implements
3. **Plan once per commission** -- Ready Room decomposes all missions up front with domain ownership and three-way sign-off; no per-mission re-planning
4. **No mission dispatched without Admiral approval** -- human approves full manifest once before any execution begins
5. **Agent → Admiral questions work** -- agents can ask, loop suspends, answer routes back
6. **Plans are persistable** -- save, shelve, resume, re-execute
7. **No agent self-certifies** -- Commander independently verifies every claim
8. **Every mission terminates** -- Commander enforces maxRevisions and demo token production
9. **Every mission yields proof** -- `demo/MISSION-<id>.md` required and validated
10. **Isolation by default** -- surface-area locks before dispatch
11. **Zero vanity tests pass** -- Commander's VERIFY_RED gate catches all
12. **Review is a gate** -- probabilistic review inside Commander's deterministic cage
13. **Wave reviews provide feedback checkpoints** -- Admiral reviews demos between waves without re-planning
14. **Observability is first-class** -- Every run produces traces (deterministic + LLM) and correlated logs; failures diagnosable via run_id/trace_id; debug bundles sufficient for reproduction

**Measured by**:

- 100% of missions have Commander-logged termination reason
- 100% of completed missions have valid `demo/MISSION-<id>.md`
- 100% of commission use cases covered by at least one mission
- Zero concurrent modification conflicts (Commander enforces locks)
- Zero missions dispatched without Admiral greenlight
- Ready Room conversation log is persisted and auditable
- 100% of runs produce complete traces with run_id and trace_id
- 100% of tool failures have span with exit code, bounded stderr, and error classification
- 100% of LLM calls have span with model name, latency, token usage (if available)
- Zero logs to stdout while TUI active (JSON logs to file only)
- Debug bundle includes: logs (last 3), redacted config, version info, last run_id/trace_id, Git state, failing test output
