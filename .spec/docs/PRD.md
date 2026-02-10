# Product Requirements Document (PRD)

## Product Name

**Ship Commander -- Agent Orchestration Runtime**

---

## One-Line Description

A Bun/TypeScript orchestration runtime that manages 3-5 parallel AI coding agents through Kanban pull-flow, TDD verification gates, Beads-backed persistent state, and multi-harness CLI execution -- bringing Gas Town-scale agent coordination to Fleet Command's structured authority model.

---

## Background & Motivation

### The Problem

AI coding workflows degrade rapidly when scaled beyond one or two agents due to:

- **Lost state**: Agent context windows are unreliable; sessions crash and lose progress
- **No verification**: Agents claim test results without independent verification (vanity tests)
- **Unclear ownership**: Multiple agents make conflicting changes with no authority boundaries
- **No visibility**: Operators cannot see what 3-5 agents are doing simultaneously
- **Manual coordination**: Humans become bottlenecks routing work between agents

### Inspiration Sources

**Steve Yegge's Gas Town** (Jan 2026) proved that 20-30 agents can work in parallel when you:

1. Store work state **externally** (not in agent memory)
2. Use **specialized roles** (Mayor, Polecats, Witness, TransporterRoom, Deacon)
3. Apply the **Propulsion Principle**: "If there is work on your hook, you MUST run it"
4. Provide **real-time observability** into every agent's activity

**Fleet Command's ROLES.md** established the authority model:

1. **Strict role hierarchy**: Admiral > Captain > First Officer > Ensigns
2. **TDD verification gates**: Orchestrator runs tests, never trusts agent claims
3. **Kanban pull-flow**: Work pulled from Backlog when WIP capacity allows
4. **Three-layer state**: Beads (truth) + TaskList (runtime) + Worktree (isolation)
5. **SPEC framework**: Self-contained Directive contracts

### What Ship Commander Builds

Ship Commander is the **runtime engine** that executes Fleet Command's model. It:

- Uses Gas Town's **Beads** (`bd` CLI) as the persistent state layer natively
- Replaces Gas Town's tmux-based visibility with a **Terminal UI (TUI)**
- Implements Fleet Command's **TDD verification gates** as programmatic enforcement
- Supports **three CLI harnesses**: Claude Code, Codex, and OpenCode
- Targets **3-5 parallel agents** (Stage 6-7) with architecture designed for 20+

---

## Core Philosophy (Non-Negotiable)

| #   | Principle                                          | Gas Town Origin                       | Fleet Command Origin                  |
| --- | -------------------------------------------------- | ------------------------------------- | ------------------------------------- |
| 1   | **Persistent state over session memory**           | Beads (`bd` CLI)                      | Three-layer state model               |
| 2   | **Agents write code; orchestrator verifies**       | Witness role                          | First Officer VERIFY gates            |
| 3   | **Specialized roles, not generalists**             | Mayor/Polecat/Witness/TransporterRoom | Captain/FO/Implementer/Reviewer       |
| 4   | **Propulsion: if work exists, execute it**         | GUPP principle                        | Pull from Backlog when WIP allows     |
| 5   | **Structure beats autonomy**                       | Hooks + Convoys                       | Verb whitelists + state machine       |
| 6   | **Ephemeral execution, persistent accountability** | Sessions crash, state survives        | Worktrees disposable, Beads persists |
| 7   | **Observability is mandatory**                     | Dashboard sees all agents             | TUI shows real-time agent status      |

---

## Execution Model: Kanban + Dual Track + Propulsion

### Kanban Flow

```
BACKLOG ──pull──> IN_PROGRESS ──verify──> REVIEW ──approve──> DONE
  (n)              (WIP: 3-5)              (n)                (n)
    ^                                       |
    └────────── redo label ─────────────────┘
```

### Dual Track: TDD vs INFRA

Directives are assigned one of two **execution tracks** based on their nature:

| Track | Purpose | Phases | When to Use |
|-------|---------|--------|-------------|
| **TDD** | Feature code with testable acceptance criteria | RED -> VERIFY_RED -> GREEN -> VERIFY_GREEN -> REFACTOR -> VERIFY_REFACTOR -> Review | Business logic, APIs, components, algorithms |
| **INFRA** | Platform setup, configuration, tooling with no meaningful test targets | IMPLEMENT -> Review | Dependency installation, CI/CD config, env setup, build tooling, scaffolding |

The track is declared in the Directive's SPEC framework via `track: tdd | infra`. The orchestrator uses the track to determine which verification flow to run.

### Propulsion Principle (Fleet Command GUPP)

> "If there is work in the Backlog and WIP capacity allows, the orchestrator MUST pull and dispatch."

The orchestrator runs a **continuous propulsion loop**:

```
LOOP (every N seconds OR on event):
  1. CHECK: WIP count < WIP limit?
  2. CHECK: Backlog has unblocked Directives?
  3. IF both true: PULL next Directive, CREATE worktree, DISPATCH agent
  4. CHECK track:
     - TDD: Any agent completed a TDD phase? RUN verification gate
     - INFRA: Agent completed implementation? DISPATCH reviewer
  5. CHECK: Any Directive fully verified/reviewed? DISPATCH reviewer (TDD) or NOTIFY human gate (INFRA)
  6. CHECK: Any review complete? NOTIFY human gate
```

### TDD Track Phases (Per Acceptance Criterion)

```
FOR EACH AC:
  Implementer -> RED: Write ONE failing test
  Orchestrator -> VERIFY_RED: Confirm failure (catch vanity tests)
  Implementer -> GREEN: Write minimal implementation
  Orchestrator -> VERIFY_GREEN: Confirm all tests pass
  Implementer -> REFACTOR: Clean up code
  Orchestrator -> VERIFY_REFACTOR: Confirm still green

ALL ACs complete -> DISPATCH Domain Reviewer
Reviewer APPROVED + Human Gate -> DONE
```

### INFRA Track Phases

```
Implementer -> IMPLEMENT: Execute all ACs (no test writing required)
Orchestrator -> VERIFY_IMPLEMENT: Run any configured validation commands (typecheck, lint, build) if present
DISPATCH Domain Reviewer
Reviewer APPROVED + Human Gate -> DONE
```

INFRA Directives skip the RED/GREEN/REFACTOR cycle entirely. The reviewer confirms the implementation is correct, complete, and follows project conventions. If validation commands are configured (e.g., `typecheck_command`, `lint_command`, `build_command`), the orchestrator runs them as a basic sanity gate before review.

### Wave Execution

```
[WAVE 1] - No dependencies
    +-- DIRECTIVE-A (worktree-1, track:infra) -> Implement -> Review
    +-- DIRECTIVE-B (worktree-2, track:tdd)   -> TDD -> Review
    +-- DIRECTIVE-C (worktree-3, track:tdd)   -> TDD -> Review

[WAIT for Wave 1 Review/Done]

[WAVE 2] - Depends on Wave 1
    +-- DIRECTIVE-D (blocked by A, track:tdd)   -> TDD -> Review
    +-- DIRECTIVE-E (blocked by B, track:infra) -> Implement -> Review
```

---

## Roles (Mapped from Fleet Command + Gas Town)

| Fleet Command Role         | Gas Town Equivalent | Ship Commander Implementation                                  |
| -------------------------- | ------------------- | -------------------------------------------------------------- |
| **Admiral** (Human)        | Overseer            | Human operator via TUI commands                                |
| **Captain**                | Mayor               | Orchestrator planning layer (decomposes goals into Directives) |
| **First Officer**          | Mayor + Witness     | Propulsion loop + dual-track verification engine               |
| **TDD Implementer Ensign** | Polecat             | Ephemeral CLI agent in isolated worktree (TDD track)           |
| **INFRA Implementer Ensign** | Polecat           | Ephemeral CLI agent in isolated worktree (INFRA track)         |
| **Domain Reviewer Ensign** | Polecat (review)    | Ephemeral CLI agent with read-only diff context (both tracks)  |
| **Doctor**                 | Deacon              | Background health monitor (stuck detection, heartbeats)        |
| _(v2)_ **TransporterRoom** | Refinery            | Merge conflict resolution agent                                |

**Key Simplification**: For v1 (3-5 agents), Captain and First Officer logic coexist in the orchestrator process. They are **logical roles** enforced by verb whitelists, not separate processes.

---

## In Scope

1. Persistent state layer via Beads (`bd` CLI) for Directives, agents, TDD phase tracking, and dependency graphs
2. Multi-harness CLI agent spawning (Claude Code, Codex, OpenCode) with process lifecycle management
3. Propulsion loop: continuous Backlog polling, WIP-limited dispatch, and verification gate execution
4. TDD verification gates (VERIFY_RED, VERIFY_GREEN, VERIFY_REFACTOR) run by orchestrator
5. Wave execution with dependency-aware scheduling (blocked-by relationships)
6. Terminal UI (TUI) showing agent status, Directive progress, event log, and TDD phase tracking
7. Event bus for real-time state change propagation between orchestrator and TUI
8. Doctor process for stuck agent detection, orphaned Directive recovery, and heartbeat monitoring
9. Graceful cancellation with SIGTERM -> SIGKILL escalation and cleanup verification
10. SPEC framework validation for Directive contracts before dispatch
11. Local-only directive state sync using Beads transitions (Backlog/InProgress/Review/Done)
12. Worktree-per-Directive isolation with branch naming conventions

## Out of Scope

1. Web dashboard (React/WebSocket) -- TUI only for v1
2. More than 5 concurrent agents -- architecture supports 20+, v1 targets 3-5
3. Merge conflict resolution (TransporterRoom role) -- human handles merges in v1
4. Formula/workflow templates (Gas Town Molecules/Formulas) -- ad-hoc Directives only
5. Session recording/replay for debugging
6. Feed curation, event filtering, or deduplication layer
7. Agent-to-agent direct communication (all routing through orchestrator)
8. Sparse checkout / worktree exclusions
9. Cost tracking / token usage monitoring
10. Remote/distributed agent execution (all agents local)

---

## Functional Groups

### STATE -- Persistent State Management

**Prefix**: STATE | **Purpose**: Use Beads (`bd` CLI) as the persistent state layer for Directives, agents, and dependencies

The state layer is the **single source of truth** for all runtime data. It survives agent crashes, orchestrator restarts, and session loss. This is the core innovation from Gas Town: agents are ephemeral, state is permanent.

Ship Commander uses **Beads natively** rather than a custom Drizzle+SQLite schema. Directives are beads. Dependencies use `bd dep`. Agent state uses `bd agent`. Gates use `bd gate`. TDD phase tracking uses state dimensions (`bd set-state`). The `.beads/` directory contains human-readable JSONL files that are grep-able, diff-able, and Git-trackable. The `bd` CLI provides instant debugging: `bd ready`, `bd graph`, `bd show`, `bd activity`.

All `bd` interactions go through a thin TypeScript wrapper (`BeadsClient`) that calls `Bun.spawn("bd", [...args], { stdout: "pipe" })` with `--json` output parsing.

| UC ID | Title | Description |
|-------|-------|-------------|
| UC-STATE-01 | Initialize beads database | Run `bd init` in project root to create `.beads/` directory with JSONL storage and SQLite cache |
| UC-STATE-02 | Track Directive lifecycle | Use `bd set-state <id> status=<STATE>` for transitions (BACKLOG -> IN_PROGRESS -> REVIEW -> DONE -> HALTED) with `--reason` for evidence |
| UC-STATE-03 | Track TDD phase per AC | Create child beads per AC (`bd create --parent <directive-id>`), track phase via `bd set-state <ac-id> tdd=<phase>`, record evidence in comments (`bd comments add`) |
| UC-STATE-04 | Manage dependency graph | Use `bd dep add <child> <parent>` for blocked-by relationships; query with `bd ready` (unblocked) and `bd blocked` (blocked) |
| UC-STATE-05 | Track agent instances | Use `bd agent state <agent-id> <state>` and `bd agent heartbeat <agent-id>` for agent lifecycle tracking |
| UC-STATE-06 | Persist event log | Use `bd audit` for append-only event recording; `bd activity` for real-time feed; in-memory ring buffer for TUI event display |
| UC-STATE-07 | Recover from crash | On restart, query `bd list --json` + `bd agent show --json` to reconstruct state; `bd stale` to detect stuck items |

**Acceptance Criteria**:

- UC-STATE-01:
  - `[ ]` System runs `bd init` if `.beads/` directory does not exist
  - `[ ]` System reuses existing `.beads/` database on restart without data loss
  - `[ ]` System verifies `bd` CLI is available on PATH at startup

- UC-STATE-02:
  - `[ ]` System transitions Directive via `bd set-state <id> status=<STATE> --reason "<evidence>"`
  - `[ ]` BeadsClient validates legal state transitions before calling `bd` (e.g., BACKLOG -> DONE rejected)
  - `[ ]` All transitions recorded with timestamp and actor via beads audit trail

- UC-STATE-03:
  - `[ ]` System creates child beads for each AC on Directive dispatch (`bd create "AC-1: <title>" --parent <directive-id>`)
  - `[ ]` System tracks TDD phase via `bd set-state <ac-id> tdd=red|verify_red|green|verify_green|refactor|verify_refactor`
  - `[ ]` VERIFY gate evidence recorded via `bd comments add <ac-id> "<exit_code, output, classification>"`
  - `[ ]` System prevents phase skipping by reading current state before advancing

- UC-STATE-04:
  - `[ ]` System queries unblocked Directives via `bd ready --json`
  - `[ ]` System computes wave assignments from `bd graph --json` dependency data
  - `[ ]` System detects circular dependencies via `bd graph` cycle detection

- UC-STATE-05:
  - `[ ]` System creates agent bead on spawn via `bd create --type agent` with role/harness labels
  - `[ ]` System updates agent state via `bd agent state <id> running|stuck|done|dead`
  - `[ ]` System updates heartbeat via `bd agent heartbeat <id>`
  - `[ ]` System queries active agent count via `bd list --label role_type:agent --label status:active --json | jq length`

- UC-STATE-06:
  - `[ ]` System records structured events via `bd audit` for all state transitions
  - `[ ]` TUI reads events from in-memory ring buffer (populated by orchestrator event bus)
  - `[ ]` `bd activity` available as CLI debugging tool for operators

- UC-STATE-07:
  - `[ ]` System queries all Directives via `bd list --json` on startup to reconstruct state
  - `[ ]` System checks `bd stale` to detect stuck Directives (no activity beyond timeout)
  - `[ ]` System transitions orphaned Directives back to BACKLOG via `bd set-state <id> status=backlog --reason "orphan_recovery"`

---

### ORCH -- Orchestration Engine

**Prefix**: ORCH | **Purpose**: Propulsion loop, wave execution, dual-track verification (TDD + INFRA), dispatch

The orchestrator is the **First Officer**: it pulls work, dispatches agents, runs verification gates, and enforces the state machine. It never writes code. It reads the Directive's `track` field to determine whether to run TDD phases or the INFRA implement/review cycle.

| UC ID | Title | Description |
|-------|-------|-------------|
| UC-ORCH-01 | Run propulsion loop | Continuously check Backlog for unblocked Directives, enforce WIP limits, and dispatch agents when capacity allows |
| UC-ORCH-02 | Dispatch TDD phase | (TDD track) Send specific TDD phase instruction (RED/GREEN/REFACTOR) to an agent with Directive context and AC details |
| UC-ORCH-03 | Execute VERIFY_RED gate | (TDD track) Run test command in worktree, confirm non-zero exit (test correctly fails), reject vanity tests |
| UC-ORCH-04 | Execute VERIFY_GREEN gate | (TDD track) Run full test suite in worktree, confirm zero exit (all tests pass) |
| UC-ORCH-05 | Execute VERIFY_REFACTOR gate | (TDD track) Run full test suite after refactor, confirm still green |
| UC-ORCH-06 | Dispatch domain review | After all ACs verified (TDD) or implementation complete (INFRA), spawn a fresh reviewer agent with diff context |
| UC-ORCH-07 | Handle review verdict | Process APPROVED (move to REVIEW for human gate) or NEEDS_FIXES (revision loop with count tracking) |
| UC-ORCH-08 | Compute wave schedule | Given dependency graph, compute which Directives can execute in parallel within each wave |
| UC-ORCH-09 | Enforce WIP limits | Reject dispatch attempts when active agent count >= configured WIP limit |
| UC-ORCH-10 | Handle escalation | When Directive is stuck (2x consecutive VERIFY failure or max revisions), halt and notify human |
| UC-ORCH-11 | Dispatch INFRA implementation | (INFRA track) Send all ACs to agent in single dispatch with full SPEC context, no test writing required |
| UC-ORCH-12 | Execute VERIFY_IMPLEMENT gate | (INFRA track) Run configured validation commands (typecheck, lint, build) if present; skip if none configured |

**Acceptance Criteria**:

- UC-ORCH-01:
  - `[ ]` Orchestrator polls Backlog on configurable interval (default 10s) or on event trigger
  - `[ ]` Orchestrator pulls highest-priority unblocked Directive when WIP allows
  - `[ ]` Orchestrator reads Directive `track` field to determine dispatch flow (TDD or INFRA)
  - `[ ]` Orchestrator creates isolated worktree before dispatching agent
  - `[ ]` Orchestrator logs all dispatch decisions to event log

- UC-ORCH-02:
  - `[ ]` Dispatch includes full SPEC framework content for the Directive
  - `[ ]` Dispatch includes specific AC being worked and expected TDD phase
  - `[ ]` Dispatch includes prior feedback if this is a revision iteration

- UC-ORCH-03:
  - `[ ]` Orchestrator runs exact test command in worktree directory
  - `[ ]` Exit code 0 -> REJECT with "Vanity test detected" (test passes without implementation)
  - `[ ]` Syntax/import error in output -> REJECT with "Test has syntax error"
  - `[ ]` Non-zero exit with test failure -> ACCEPT, proceed to GREEN

- UC-ORCH-04:
  - `[ ]` Orchestrator runs full test suite (not just new test)
  - `[ ]` Exit code 0 -> ACCEPT, proceed to REFACTOR
  - `[ ]` Exit code != 0 -> Loop back to GREEN with failure output

- UC-ORCH-05:
  - `[ ]` Orchestrator runs full test suite after refactor
  - `[ ]` Exit code 0 -> ACCEPT, move to next AC (or REVIEW if last AC)
  - `[ ]` Exit code != 0 -> REJECT refactor, revert to pre-refactor state

- UC-ORCH-06:
  - `[ ]` Reviewer agent is a different instance than the implementer (context isolation)
  - `[ ]` For TDD track: Reviewer receives code diff, TDD evidence (VERIFY gate results), acceptance criteria
  - `[ ]` For INFRA track: Reviewer receives code diff, validation results (if any), acceptance criteria
  - `[ ]` Reviewer does NOT receive implementer's internal reasoning

- UC-ORCH-07:
  - `[ ]` APPROVED verdict moves Directive to REVIEW state (awaiting human gate)
  - `[ ]` NEEDS_FIXES increments revision_count, transitions REVIEW -> IN_PROGRESS
  - `[ ]` revision_count >= 3 triggers HALTED state with escalation

- UC-ORCH-08:
  - `[ ]` Wave computation respects blocked-by relationships
  - `[ ]` Directives with no dependencies are assigned to Wave 1
  - `[ ]` Wave N+1 contains Directives whose dependencies are all in Wave <= N

- UC-ORCH-09:
  - `[ ]` WIP limit is configurable (default 3, max 5 for v1)
  - `[ ]` Active agent count is derived from Beads state, not process list
  - `[ ]` WIP enforcement is checked before every dispatch

- UC-ORCH-10:
  - `[ ]` 2x consecutive VERIFY_GREEN failure for same AC triggers escalation (TDD track)
  - `[ ]` revision_count >= max_revisions triggers HALTED (both tracks)
  - `[ ]` Escalation posts structured message to TUI and event log

- UC-ORCH-11:
  - `[ ]` INFRA dispatch sends all ACs in a single agent session (no per-AC iteration)
  - `[ ]` Dispatch prompt instructs agent to implement without writing tests
  - `[ ]` Dispatch includes prior feedback if this is a revision iteration

- UC-ORCH-12:
  - `[ ]` If `typecheck_command`, `lint_command`, or `build_command` configured, run each in worktree
  - `[ ]` All configured commands must pass (exit 0) to proceed to review
  - `[ ]` If no validation commands configured, proceed directly to review
  - `[ ]` Validation failures loop back to implementer with failure output

---

### HARN -- Harness Abstraction

**Prefix**: HARN | **Purpose**: Spawn and manage CLI agent subprocesses across Claude Code, Codex, and OpenCode

The harness layer abstracts away which CLI tool executes the agent. All harnesses implement the same interface: spawn an agent with a prompt, capture output, enforce timeout, return result

See `./old_wine/harness` for an example of a harness wrapper that worked in another project

| UC ID      | Title                           | Description                                                                                                            |
| ---------- | ------------------------------- | ---------------------------------------------------------------------------------------------------------------------- |
| UC-HARN-01 | Define harness driver interface | Common interface for spawning implementer and reviewer agents across any CLI harness                                   |
| UC-HARN-02 | Implement Claude Code driver    | Spawn `claude` CLI with `-p`, model selection, timeout, and working directory                                          |
| UC-HARN-03 | Implement Codex driver          | Spawn `codex` CLI with sandbox mode, approval policy, and model configuration                                          |
| UC-HARN-04 | Implement OpenCode driver       | Spawn `opencode` CLI with agent selection, model, and working directory                                                |
| UC-HARN-05 | Manage process lifecycle        | Spawn with piped stdin/stdout/stderr, capture output with size limits, enforce timeout with SIGTERM/SIGKILL escalation |
| UC-HARN-06 | Stream agent output             | Real-time streaming of agent stdout/stderr to event bus for TUI display                                                |
| UC-HARN-07 | Detect harness availability     | Check which CLI tools are installed and on PATH at startup; fail fast if none available                                |
| UC-HARN-08 | Configure harness per Directive | Allow Directives to specify preferred harness (or use system default)                                                  |

**Acceptance Criteria**:

- UC-HARN-01:
  - `[ ]` Interface defines `spawnImplementer(session, options)` and `spawnReviewer(session, reviewer, options)` methods
  - `[ ]` Interface supports optional `onOutput` callback for real-time streaming
  - `[ ]` Interface returns typed `ImplementerOutput` and `ReviewerOutput` structs

- UC-HARN-02:
  - `[ ]` Claude Code driver constructs correct CLI flags: `-p --model <model> --verbose --max-turns 20`
  - `[ ]` Driver supports model selection (haiku, sonnet, opus)
  - `[ ]` Driver writes prompt to stdin and captures stdout/stderr

- UC-HARN-03:
  - `[ ]` Codex driver constructs correct CLI flags: `--sandbox <mode> -m <model> exec -`
  - `[ ]` Driver supports sandbox modes (read-only, workspace-write, danger-full-access)
  - `[ ]` Driver supports approval policies (untrusted, on-failure, on-request, never)

- UC-HARN-04:
  - `[ ]` OpenCode driver spawns `opencode` CLI with correct arguments
  - `[ ]` Driver supports agent selection and model configuration
  - `[ ]` Driver captures output in same format as other drivers

- UC-HARN-05:
  - `[ ]` Process output truncated at configurable limit (default 1MB) to prevent memory bloat
  - `[ ]` Timeout enforced with 5-second SIGTERM grace period before SIGKILL
  - `[ ]` No zombie processes left after timeout or crash

- UC-HARN-06:
  - `[ ]` Agent stdout/stderr chunks emitted as events in real-time
  - `[ ]` TUI can subscribe to agent output stream by agent ID
  - `[ ]` Output buffered if no subscribers (no data loss)

- UC-HARN-07:
  - `[ ]` System checks for `claude`, `codex`, `opencode` on PATH at startup
  - `[ ]` System reports which harnesses are available
  - `[ ]` System errors if zero harnesses available

- UC-HARN-08:
  - `[ ]` Directive SPEC can include `harness: claude | codex | opencode` field
  - `[ ]` If not specified, system uses configured default harness
  - `[ ]` If specified harness unavailable, falls back to default with warning

---

### TUI -- Terminal User Interface

**Prefix**: TUI | **Purpose**: Real-time visibility into agent activity, Directive progress, and system health

The TUI is the operator's window into the orchestration runtime. It replaces Gas Town's tmux-based monitoring with a structured terminal dashboard.

| UC ID     | Title                       | Description                                                                                         |
| --------- | --------------------------- | --------------------------------------------------------------------------------------------------- |
| UC-TUI-01 | Display agent status grid   | Show all active agents with role, assigned Directive, current TDD phase, harness type, and duration |
| UC-TUI-02 | Display Directive board     | Kanban-style columns showing Directives in each state (Backlog, In Progress, Review, Done, Halted)  |
| UC-TUI-03 | Display live event log      | Scrolling log of recent events (state transitions, VERIFY results, agent spawns/exits, errors)      |
| UC-TUI-04 | Display TDD phase tracker   | Per-Directive view showing AC progress through RED/GREEN/REFACTOR with pass/fail indicators         |
| UC-TUI-05 | Display wave execution view | Show current wave, which Directives are in parallel, dependency graph visualization                 |
| UC-TUI-06 | Accept operator commands    | Command input for: halt agent, force-retry Directive, approve human gate, change WIP limit          |
| UC-TUI-07 | Display system health       | WIP utilization, agent uptime, stuck detection warnings, heartbeat status                           |

**Acceptance Criteria**:

- UC-TUI-01:
  - `[ ]` Agent grid updates in real-time as agents spawn, progress, and exit
  - `[ ]` Each agent card shows: role icon, Directive ID, TDD phase, elapsed time, harness name
  - `[ ]` Stuck agents (no heartbeat > timeout) highlighted in warning color

- UC-TUI-02:
  - `[ ]` Board shows count per column and Directive identifiers
  - `[ ]` Directives display priority, blocked-by count, and current assignee
  - `[ ]` Board refreshes on every state transition event

- UC-TUI-03:
  - `[ ]` Event log shows last 50 events with timestamp, type, actor, and summary
  - `[ ]` Events color-coded by severity (info, warn, error)
  - `[ ]` Log is scrollable with keyboard navigation

- UC-TUI-04:
  - `[ ]` Each AC shows checkboxes for RED, VERIFY_RED, GREEN, VERIFY_GREEN, REFACTOR, VERIFY_REFACTOR
  - `[ ]` Current active phase highlighted
  - `[ ]` VERIFY rejections shown with reason

- UC-TUI-05:
  - `[ ]` Current wave number and parallel Directive count displayed
  - `[ ]` Dependency arrows shown between related Directives
  - `[ ]` Completed waves visually distinguished from active wave

- UC-TUI-06:
  - `[ ]` Operator can type commands in input bar at bottom of TUI
  - `[ ]` Commands: `halt <id>`, `retry <id>`, `approve <id>`, `wip <n>`, `quit`
  - `[ ]` Command feedback shown in event log

- UC-TUI-07:
  - `[ ]` WIP utilization bar (e.g., "3/5 agents active")
  - `[ ]` Doctor heartbeat indicator (green/red)
  - `[ ]` Count of stuck agents and orphaned Directives

---

### GATE -- Verification Gates

**Prefix**: GATE | **Purpose**: Orchestrator-enforced test execution preventing vanity tests and false claims

Verification gates are the core correctness mechanism from Fleet Command. The orchestrator independently runs tests -- it never trusts agent claims about results.

| UC ID      | Title                                 | Description                                                                                                           |
| ---------- | ------------------------------------- | --------------------------------------------------------------------------------------------------------------------- |
| UC-GATE-01 | Execute bash commands in worktree     | Run arbitrary shell commands (test, typecheck, lint) in a Directive's isolated worktree                               |
| UC-GATE-02 | Parse test output                     | Extract pass/fail count, failure messages, and exit code from test runner output                                      |
| UC-GATE-03 | Classify VERIFY_RED result            | Determine if test correctly fails (ACCEPT), passes without impl (REJECT: vanity), or has syntax error (REJECT: error) |
| UC-GATE-04 | Classify VERIFY_GREEN result          | Determine if all tests pass (ACCEPT) or any fail (REJECT: loop back)                                                  |
| UC-GATE-05 | Record gate evidence                  | Persist gate result (pass/fail, exit code, output snippet, classification) to Beads comments and event log            |
| UC-GATE-06 | Configure test commands per Directive | SPEC framework includes test_command, typecheck_command, lint_command fields                                          |

**Acceptance Criteria**:

- UC-GATE-01:
  - `[ ]` Commands execute in the worktree directory (not project root)
  - `[ ]` Command output captured with configurable size limit
  - `[ ]` Timeout enforced per command (default 120s)

- UC-GATE-02:
  - `[ ]` Parser extracts pass count, fail count, and error count from common test runners (vitest, jest, pytest, go test)
  - `[ ]` Parser extracts first failure message for feedback to implementer
  - `[ ]` Raw output preserved in gate evidence for debugging

- UC-GATE-03:
  - `[ ]` Exit 0 with no implementation files -> REJECT vanity test
  - `[ ]` Non-zero exit with SYNTAX/IMPORT error pattern -> REJECT syntax error
  - `[ ]` Non-zero exit with test FAIL/FAILURE pattern -> ACCEPT (correct failure)
  - `[ ]` Classification persisted as gate evidence

- UC-GATE-04:
  - `[ ]` Exit 0 -> ACCEPT
  - `[ ]` Non-zero exit -> REJECT with failure details for implementer
  - `[ ]` Runs full test suite, not just the new test file

- UC-GATE-05:
  - `[ ]` Gate evidence includes: gate type, AC ID, exit code, classification, output snippet, timestamp
  - `[ ]` Evidence queryable by Directive ID for reviewer consumption
  - `[ ]` Evidence emitted as event for TUI display

- UC-GATE-06:
  - `[ ]` SPEC framework includes `test_command`, `typecheck_command`, `lint_command`
  - `[ ]` If not specified, system uses project-level defaults from config
  - `[ ]` Commands support variable substitution (e.g., `{test_file}` for VERIFY_RED)

---

### RESIL -- Operational Resilience

**Prefix**: RESIL | **Purpose**: Health monitoring, crash recovery, stuck detection, and graceful degradation

Resilience patterns from both Gas Town (Deacon patrols, handoffs) and Fleet Command (Doctor, cancellation) ensure the system self-heals.

| UC ID       | Title                         | Description                                                                                           |
| ----------- | ----------------------------- | ----------------------------------------------------------------------------------------------------- |
| UC-RESIL-01 | Run Doctor heartbeat loop     | Background process checks agent health, Directive progress, and system state on configurable interval |
| UC-RESIL-02 | Detect stuck agents           | Identify agents with no progress (no events) beyond configurable timeout                              |
| UC-RESIL-03 | Detect orphaned Directives    | Find Directives in IN_PROGRESS state with no active agent instance                                    |
| UC-RESIL-04 | Graceful cancellation         | Cancel agent with SIGTERM, wait grace period, check cleanup, then SIGKILL if needed                   |
| UC-RESIL-05 | Command safety validation     | Pre-validate shell commands against dangerous patterns before execution                               |
| UC-RESIL-06 | Orchestrator restart recovery | On restart, read state from Beads (`bd list --json`), identify inconsistencies, and resume propulsion loop |
| UC-RESIL-07 | Worktree lifecycle management | Create worktree on dispatch, clean up on completion/halt, detect orphaned worktrees                   |

**Acceptance Criteria**:

- UC-RESIL-01:
  - `[ ]` Doctor runs on configurable interval (default 30s)
  - `[ ]` Each heartbeat cycle checks: agent health, Directive staleness, worktree integrity
  - `[ ]` Heartbeat results emitted as events for TUI health display

- UC-RESIL-02:
  - `[ ]` Agent considered stuck if no events emitted for configurable timeout (default 5 min)
  - `[ ]` Stuck agents logged as warnings in event log
  - `[ ]` After 2x timeout, agent force-killed and Directive returned to BACKLOG

- UC-RESIL-03:
  - `[ ]` Orphan detection runs on every Doctor cycle and on orchestrator startup
  - `[ ]` Orphaned Directives transitioned to BACKLOG with "orphan_recovery" reason
  - `[ ]` Event logged with original agent ID and Directive state

- UC-RESIL-04:
  - `[ ]` SIGTERM sent first, with 5-second grace period
  - `[ ]` If process still running after grace period, SIGKILL sent
  - `[ ]` Process confirmed exited before cleanup proceeds
  - `[ ]` No zombie processes after cancellation

- UC-RESIL-05:
  - `[ ]` Commands checked against deny list: `rm -rf /`, `sudo`, `chmod 777`, `> /dev/sda`
  - `[ ]` Commands checked against allow list for known safe patterns
  - `[ ]` Rejected commands logged as security events

- UC-RESIL-06:
  - `[ ]` On startup, system reads all Directives, agents, and events from Beads (`bd list --json`, `bd agent show --json`)
  - `[ ]` System identifies process-alive vs registered-dead agents (PID check)
  - `[ ]` Propulsion loop resumes from persisted state within 10 seconds of restart

- UC-RESIL-07:
  - `[ ]` Worktree created at `.worktrees/{directive-id}/` with branch `feature/{directive-id}-{slug}`
  - `[ ]` Worktree cleaned up after Directive reaches DONE or HALTED
  - `[ ]` Orphaned worktrees (no matching Directive in non-terminal state) detected and reported

---

## Use Case Summary

| Group                       | Prefix | Use Cases | Description                                                 |
| --------------------------- | ------ | --------- | ----------------------------------------------------------- |
| Persistent State Management | STATE  | 7         | Beads (`bd` CLI) state layer with JSONL persistence         |
| Orchestration Engine        | ORCH   | 12        | Propulsion loop, dual-track dispatch (TDD + INFRA), wave execution, verification |
| Harness Abstraction         | HARN   | 8         | Multi-CLI agent spawning (Claude Code, Codex, OpenCode)     |
| Terminal User Interface     | TUI    | 7         | Real-time visibility into agents, Directives, and health    |
| Verification Gates          | GATE   | 6         | Orchestrator-enforced test execution and evidence recording |
| Operational Resilience      | RESIL  | 7         | Health monitoring, crash recovery, stuck detection          |
| **Total**                   |        | **47**    |                                                             |

---

## Technical Requirements

### System Components

| Component           | Description                                                                    | Technology                                 |
| ------------------- | ------------------------------------------------------------------------------ | ------------------------------------------ |
| **State Layer**     | Persistent state for Directives, agents, TDD phases, events via Beads          | `bd` CLI + BeadsClient TypeScript wrapper  |
| **Orchestrator**    | Propulsion loop, dispatch, verification gates, state machine enforcement       | Bun/TypeScript + XState (state machines)   |
| **Harness Layer**   | CLI subprocess spawning and lifecycle management                               | Bun.spawn + Node child_process             |
| **Event Bus**       | Typed pub/sub with wildcard subscribers for state change propagation            | Custom EventBus (typed events + `*` subscriptions) |
| **TUI Renderer**    | Terminal dashboard with panels for agents, board, events, health               | Ink (React for terminal UIs)               |
| **Doctor**          | Background health monitor running on heartbeat interval                        | Bun setInterval + `bd agent` state queries |
| **Tracing**         | Distributed tracing for agent execution, TDD phases, and gate results          | OpenTelemetry (`@opentelemetry/api`)       |
| **CLI Entry Point** | Command-line interface for starting, configuring, and operating Ship Commander | Bun CLI                                    |

### Data Model (Beads)

All state is stored in `.beads/` as JSONL files with SQLite cache. No custom schema is needed -- Ship Commander uses Beads' native data model with conventions:

#### Directive Beads

Created via `bd create "<title>" --type task -p <priority>`. Tracked using:

| Beads Field | Ship Commander Usage |
|-------------|---------------------|
| `id` | Unique hash ID (e.g., `bd-a1b2c3`) |
| `title` | Human-readable Directive title |
| `type` | `task` (standard Beads type) |
| `priority` | WSJF priority score (0 = highest) |
| `status` | `open` / `closed` (Beads native) |
| State dimension `status` | `backlog` / `in_progress` / `review` / `done` / `halted` (via `bd set-state`) |
| State dimension `wave` | Computed wave assignment (via `bd set-state`) |
| Labels | `harness:<type>`, `revision_count:<n>` |
| Dependencies | `bd dep add <child> <parent>` |
| Body/description | Full SPEC framework content |

#### AC Child Beads (TDD Phase Tracking)

Created via `bd create "AC-1: <title>" --parent <directive-id>`. Each AC is a child bead:

| Beads Field | Ship Commander Usage |
|-------------|---------------------|
| `id` | Unique hash ID |
| `parent` | Directive bead ID |
| State dimension `tdd` | `red` / `verify_red` / `green` / `verify_green` / `refactor` / `verify_refactor` |
| Comments | VERIFY gate evidence (exit code, output snippet, classification) |

#### Agent Beads

Created via `bd create --type agent` with `bd agent state` management:

| Beads Field | Ship Commander Usage |
|-------------|---------------------|
| `id` | Unique agent instance ID |
| Labels | `role_type:tdd_implementer` or `role_type:domain_reviewer`, `harness:<type>`, `directive:<id>` |
| Agent state | `idle` / `spawning` / `running` / `working` / `stuck` / `done` / `dead` (via `bd agent state`) |
| Heartbeat | `bd agent heartbeat <id>` updates `last_activity` |

#### Dependencies

Managed natively by Beads:
- `bd dep add <child> <parent>` -- blocked-by relationship
- `bd ready` -- query unblocked items
- `bd blocked` -- query blocked items
- `bd graph` -- visualize dependency tree

#### Event Log

Two layers:
- **Beads audit** (`bd audit`): Append-only JSONL for durable event recording
- **In-memory ring buffer**: EventEmitter-based for TUI real-time display (last 200 events)

### Architecture Diagram

```
+------------------------------------------------------------------+
|                        SHIP COMMANDER                              |
|                                                                    |
|  +-----------+     +----------------+     +-------------------+    |
|  |  CLI      |     |  TUI Renderer  |     |  Doctor           |    |
|  |  Entry    |---->|  (Ink)         |     |  (Health Monitor) |    |
|  |  Point    |     |                |     |  - Heartbeat loop |    |
|  +-----------+     +-------+--------+     |  - Stuck detect   |    |
|                            |              |  - Orphan detect   |    |
|                            | subscribes   +--------+----------+    |
|                            v                       |               |
|                   +--------+--------+              | queries       |
|                   |   Event Bus     |<-------------+               |
|                   |  (pub/sub)      |                              |
|                   +--------+--------+                              |
|                            ^                                       |
|                            | emits                                 |
|                   +--------+--------+                              |
|                   |  Orchestrator   |                              |
|                   |  (First Officer)|                              |
|                   |  - Propulsion   |                              |
|                   |  - Dispatch     |                              |
|                   |  - VERIFY gates |                              |
|                   |  - Wave sched   |                              |
|                   +---+----+----+---+                              |
|                       |    |    |                                   |
|              spawns   |    |    |  spawns                          |
|         +-------------+    |    +-------------+                    |
|         v                  v                  v                    |
|  +------+------+   +------+------+   +-------+-----+              |
|  | Claude Code |   | Codex       |   | OpenCode    |              |
|  | Driver      |   | Driver      |   | Driver      |              |
|  +------+------+   +------+------+   +-------+-----+              |
|         |                  |                  |                    |
|         v                  v                  v                    |
|  +------+------+   +------+------+   +-------+-----+              |
|  | Worktree 1  |   | Worktree 2  |   | Worktree 3  |              |
|  | DIR-001     |   | DIR-002     |   | DIR-003     |              |
|  +-------------+   +-------------+   +-------------+              |
|                                                                    |
|  +--------------------------------------------------------------+  |
|  |  Beads State Layer (.beads/)                                  |  |
|  |  JSONL: directives, agents, AC children, deps, audit log     |  |
|  |  CLI: bd ready | bd graph | bd agent | bd gate | bd activity |  |
|  +--------------------------------------------------------------+  |
+------------------------------------------------------------------+
```

### External Dependencies

#### Runtime

| Dependency         | Purpose                                                     | Documentation                              |
| ------------------ | ----------------------------------------------------------- | ------------------------------------------ |
| **Bun** 1.1+       | JavaScript runtime for subprocess spawning and fast startup | https://bun.sh/docs                           |
| **Beads** (`bd`)   | Persistent state layer -- issue tracking, deps, agents, gates | https://github.com/steveyegge/beads         |
| **XState** 5+      | State machine enforcement for Kanban flow and TDD phases      | https://stately.ai/docs/xstate              |

#### Observability

| Dependency              | Purpose                                                     | Documentation                                        |
| ----------------------- | ----------------------------------------------------------- | ---------------------------------------------------- |
| **@opentelemetry/api**  | Tracing API for spans across agent execution and gate runs  | https://opentelemetry.io/docs/languages/js/           |
| **@opentelemetry/sdk-node** | Node/Bun SDK for trace collection and export            | https://opentelemetry.io/docs/languages/js/getting-started/nodejs/ |
| **OTLP Exporter**       | Export traces to Jaeger, Grafana Tempo, or other backends   | https://opentelemetry.io/docs/specs/otlp/             |

#### TUI

| Dependency       | Purpose                                                  | Documentation                              |
| ---------------- | -------------------------------------------------------- | ------------------------------------------ |
| **Ink**          | React for terminal UIs -- build interactive CLI apps with components | https://github.com/vadimdemedes/ink |

#### CLI Harnesses (External, user-installed)

| Dependency       | Purpose                    | Documentation                                  |
| ---------------- | -------------------------- | ---------------------------------------------- |
| **claude** CLI   | Claude Code agent harness  | https://docs.anthropic.com/en/docs/claude-code |
| **codex** CLI    | OpenAI Codex agent harness | https://github.com/openai/codex                |
| **opencode** CLI | OpenCode agent harness     | https://opencode.ai/docs                       |

#### Development

| Dependency          | Purpose     | Documentation                  |
| ------------------- | ----------- | ------------------------------ |
| **Vitest**          | Test runner | https://vitest.dev             |
| **TypeScript** 5.5+ | Type safety | https://www.typescriptlang.org |

---

## Terminology Mapping

For contributors familiar with Gas Town or Fleet Command:

| Gas Town Term | Fleet Command Term            | Ship Commander Implementation                                                                   |
| ------------- | ----------------------------- | ----------------------------------------------------------------------------------------------- |
| Bead          | Directive                     | Bead in `.beads/` with SPEC content in body                     |
| Hook          | Work Queue                    | `bd ready --json` (returns unblocked, open beads by priority)   |
| Convoy        | Wave                          | Computed wave assignment from `bd graph --json`                 |
| GUPP          | Propulsion Loop               | `setInterval` + event-driven dispatch cycle                     |
| Mayor         | Captain + First Officer       | Orchestrator process (logical role separation)                  |
| Polecat       | Ensign (Implementer/Reviewer) | CLI subprocess via harness driver                               |
| Witness       | Doctor                        | Heartbeat loop with stuck/orphan detection                      |
| Refinery      | TransporterRoom _(v2)_        | Human handles merges in v1                                      |
| Deacon        | Doctor                        | Same concept, different name                                    |
| Sling         | Dispatch                      | `orchestrator.dispatch(directive, harness)`                     |
| Town          | State Layer                   | `.beads/` directory (JSONL + SQLite cache)                      |
| Rig           | Starship/Project              | Working directory with `.worktrees/`                            |
| Crew          | Operator                      | Human at TUI                                                    |
| Handoff       | Crash Recovery                | Beads JSONL enables session-less restart                        |
| Seance        | Event Log Query               | `bd activity` / `bd audit`                                      |
| Formula       | _(out of scope)_              | Ad-hoc Directives only in v1                                    |
| Molecule      | Dependency Graph              | `bd dep` + `bd graph` with wave computation                     |

---

## Constraints

| Constraint                  | Value              | Rationale                               |
| --------------------------- | ------------------ | --------------------------------------- |
| Max concurrent agents (v1)  | 5                  | Prove architecture before scaling       |
| Default WIP limit           | 3                  | Conservative start                      |
| Max revisions per Directive | 3                  | Prevent infinite loops                  |
| Process output limit        | 1 MB               | Prevent memory bloat                    |
| SIGTERM grace period        | 5 seconds          | Balance cleanup vs. responsiveness      |
| Doctor heartbeat interval   | 30 seconds         | Responsive without polling overhead     |
| Stuck agent timeout         | 5 minutes          | Detect genuinely stuck, not slow        |
| Test command timeout        | 120 seconds        | Most test suites finish in < 60s        |
| State layer                 | Beads `.beads/` dir | Human-readable JSONL, no external services |
| Runtime                     | Bun only           | Required for Bun.spawn and fast startup |

---

## Appendix A -- Machine-Readable Policy (Normative)

```yaml
ship_commander:
  execution_model: kanban_with_tdd_and_propulsion
  target_scale: 3-5 agents (v1), 20+ (architecture)
  runtime: bun_typescript

  state_layer:
    backend: beads_cli
    storage: jsonl_with_sqlite_cache
    location: .beads/
    entities: [directive_beads, ac_child_beads, agent_beads, dependencies, audit_log]
    crash_recovery: bd_list_on_restart
    debugging: [bd ready, bd graph, bd show, bd activity, bd stale]

  propulsion:
    trigger: interval_10s_or_event
    check: wip_available AND backlog_has_unblocked
    action: pull_dispatch_verify

  kanban_states: [BACKLOG, IN_PROGRESS, REVIEW, DONE, HALTED]
  terminal_states: [DONE, HALTED]
  wip_limits:
    in_progress: 3 (default), 5 (max v1)

  tracks:
    tdd:
      phases_per_ac:
        - RED
        - VERIFY_RED
        - GREEN
        - VERIFY_GREEN
        - REFACTOR
        - VERIFY_REFACTOR
      use_for: [business_logic, apis, components, algorithms]
    infra:
      phases:
        - IMPLEMENT
        - VERIFY_IMPLEMENT (optional, runs typecheck/lint/build if configured)
        - REVIEW
      use_for: [dependency_setup, ci_cd_config, env_setup, build_tooling, scaffolding]

  harnesses:
    supported: [claude, codex, opencode]
    interface: [spawnImplementer, spawnReviewer]
    process_mgmt: bun_spawn_with_timeout

  tui:
    framework: ink
    panels: [agent_grid, directive_board, event_log, tdd_tracker, health]
    commands: [halt, retry, approve, wip, quit]

  event_bus:
    pattern: typed_pub_sub_with_wildcard
    events: [agent:spawned, agent:thinking, agent:action, agent:done, agent:stuck,
             task:assigned, task:phase_change, task:complete, task:halted,
             gate:start, gate:pass, gate:fail,
             doctor:heartbeat, doctor:stuck_detected, doctor:orphan_detected]

  state_machines:
    framework: xstate
    machines:
      kanban: [BACKLOG, IN_PROGRESS, REVIEW, DONE, HALTED]
      tdd_phase: [RED, VERIFY_RED, GREEN, VERIFY_GREEN, REFACTOR, VERIFY_REFACTOR]
      infra_phase: [IMPLEMENT, VERIFY_IMPLEMENT, REVIEW]
      agent_lifecycle: [idle, spawning, running, working, stuck, done, dead]

  observability:
    tracing: opentelemetry
    span_types: [agent_execution, tdd_phase, verification_gate, propulsion_cycle, wave_execution]
    exporters: [otlp_grpc, console (dev)]
    backends: [jaeger, grafana_tempo]

  resilience:
    doctor_interval: 30s
    stuck_timeout: 5min
    orphan_detection: on_startup_and_heartbeat
    cancellation: sigterm_then_sigkill

  roles:
    captain:
      allowed: [approve, reject, assign, halt, escalate, prioritize]
      forbidden: [implement, test, merge]
    first_officer:
      allowed:
        [pull, dispatch, verify, create_worktree, teardown_worktree, escalate]
      forbidden: [implement, write_code, review_code, merge]
    tdd_implementer:
      allowed: [write_test, implement, refactor, run_tests, report]
      forbidden: [assign, complete, merge, review_own_work, skip_red]
    infra_implementer:
      allowed: [implement, configure, install, scaffold, run_validation, report]
      forbidden: [assign, complete, merge, review_own_work]
    domain_reviewer:
      allowed: [inspect_code, run_tests, verify, pass, fail, report]
      forbidden: [implement, refactor, merge, write_code]

  verification_gates:
    tdd_track:
      verify_red:
        run_by: orchestrator
        expected: non_zero_exit
        reject_if: [exit_0_vanity, syntax_error]
      verify_green:
        run_by: orchestrator
        expected: exit_0
        reject_if: [any_failure]
      verify_refactor:
        run_by: orchestrator
        expected: exit_0
        reject_if: [any_failure]
    infra_track:
      verify_implement:
        run_by: orchestrator
        commands: [typecheck_command, lint_command, build_command]
        required: false  # skip if no commands configured
        expected: exit_0
        reject_if: [any_failure]

  hard_fail_conditions:
    - vanity_test_undetected (TDD track)
    - implementer_skips_red (TDD track)
    - missing_verify_evidence
    - unauthorized_transition
    - agent_claims_results_without_verification
```
