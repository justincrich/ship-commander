# Essential Processes & Structures

## Deterministic vs Probabilistic Alignment Review

### For KB / ship-commander / Future Direction

**Purpose**  
This document extracts the _essential, load-bearing processes and data structures_ that support the manifesto’s goals of nimble, fast, safe iteration.  
It explicitly classifies each element as **Deterministic** or **Probabilistic**, and highlights:

- **What already exists** in the KB / Claude / ship-commander system
- **What is new or different** relative to current practice

Use this as a shared reference for agent collaboration and path-forward decisions.

---

## 1. Core System Shape (High-Level)

The system is intentionally split into two layers:

| Layer                           | Nature             | Responsibility                              |
| ------------------------------- | ------------------ | ------------------------------------------- |
| **Deterministic Orchestration** | Stable, testable   | Enforce rules, run tools, decide acceptance |
| **Probabilistic Execution**     | Adaptive, creative | Generate code, tests, plans, reviews        |

**Key invariant:**

> _All probabilistic output must pass through at least one deterministic gate before acceptance._

This invariant is already present in KB and ship-commander and must never be weakened.

---

## 2. Essential Processes

### A. Mission Execution (Single-Mission Completion)

**Description**  
A _Mission_ is the smallest executable unit of work that can be accepted or rejected.

**Deterministic aspects**

- Mission lifecycle states (queued → running → verified → accepted/rejected)
- Attempt counting and termination rules
- Enforcement of “single mission must terminate”

**Probabilistic aspects**

- Writing code
- Writing tests or other proof artifacts
- Interpreting objectives/ACs

**Exists today**

- Per-AC TDD cycles in KB
- RED → GREEN → REFACTOR sequencing
- Orchestrator VERIFY gates after each phase

**What’s different**

- “Mission” becomes the explicit primitive (not just implicit AC loops)
- Mission may require **proof**, not strictly a test
- Mission mode (Red Alert vs Standard Ops) is explicit

---

### B. Independent Verification (“Don’t Trust the Agent”)

**Description**  
All claims made by agents are independently verified by the orchestrator.

**Deterministic aspects**

- `pnpm test`, `pnpm typecheck`, `pnpm lint`, `build`
- Exit-code evaluation
- Infra re-checks, flake detection, vanity-test detection

**Probabilistic aspects**

- Agent’s description of what they did
- Agent’s belief that tests cover the objective

**Exists today**

- VERIFY_RED / VERIFY_GREEN gates
- Infra mismatch checks
- AC count validation

**What’s different**

- Verification generalized beyond tests (demo tokens)
- Proof verification becomes a first-class concept

---

### C. Deterministic Orchestration & State Management

**Description**  
The orchestrator governs execution using a finite state machine.

**Deterministic aspects**

- Linear state transitions
- Polling loops
- Timeouts and retries
- `.kb-swarm-state.json` persistence
- Worktree lifecycle

**Probabilistic aspects**

- None (this layer must be purely deterministic)

**Exists today**

- Linear-driven workflow
- Comment polling + parsing
- Session state file

**What’s different**

- Protocol should become versioned & structured (JSON, not regex)
- Mission state tracked explicitly, not inferred

---

### D. Work Isolation & Concurrency Control

**Description**  
Parallelism without corruption.

**Deterministic aspects**

- Worktree per issue/mission
- Lock acquisition/release rules (future)

**Probabilistic aspects**

- None

**Exists today**

- Worktree-per-issue isolation

**What’s different**

- Explicit _surface-area locks_ (DB, auth, core utils)
- Orchestrator enforcement instead of “best effort”

---

### E. Agent Routing & Specialization

**Description**  
Tasks are routed to domain-specific agents.

**Deterministic aspects**

- Area-label → agent mapping
- Reviewer dispatch rules

**Probabilistic aspects**

- Actual implementation and review quality

**Exists today**

- Convex / Next / RN / Vite / Python routing
- Domain reviewers

**What’s different**

- Explicit breaker / multi-review for Red Alert missions
- Clear separation between implementers and judges

---

### F. Review & Resolution

**Description**  
Reviews decide acceptance, not negotiate scope.

**Deterministic aspects**

- Mandatory gates before review
- Verdict states (APPROVED / NEEDS_FIXES / ESCALATE)
- State transitions on verdict

**Probabilistic aspects**

- Code quality judgment
- AC interpretation
- Security reasoning

**Exists today**

- Review pyramid
- NEEDS_FIXES → re-dispatch loop

**What’s different**

- Verdicts should be fully structured
- Review outcomes feed analytics and retry limits

---

## 3. Essential Data Structures

### A. Mission

**Deterministic fields**

- `mission_id`
- `issue_id`
- `mode` (RED_ALERT | STANDARD_OPS)
- `status`
- `attempt_count`

**Probabilistic fields**

- `objective`
- `risk_assessment` (input)

**Exists today**

- ACs + TDD phase markers

**What’s different**

- Mission becomes explicit, mode-aware, proof-driven

---

### B. Protocol Event

**Deterministic**

- `event_type`
- `mission_id`
- `timestamp`
- `protocol_version`

**Probabilistic**

- Agent-supplied descriptions

**Exists today**

- Structured comments with markers

**What’s different**

- JSON schema instead of regex parsing
- Versioned protocol

---

### C. Gate Result

**Deterministic**

- `gate_name`
- `command`
- `exit_code`
- `passed`
- `stdout/stderr`

**Probabilistic**

- None

**Exists today**

- Implicit via logs

**What’s different**

- First-class persisted artifacts

---

### D. Review Verdict

**Deterministic**

- `verdict`
- `must_fix` flags
- Gate confirmation

**Probabilistic**

- Findings content
- Severity judgment

**Exists today**

- Review comments

**What’s different**

- Machine-readable verdict structure

---

### E. Demo Token (New but aligned)

**Deterministic**

- `demo_type`
- `how_to_verify`
- `expected_result`

**Probabilistic**

- Agent explanation

**Exists today**

- Implicit via tests

**What’s different**

- Generalized proof mechanism
- Enables test-optional missions

---

## 4. Invariants (Must Always Hold)

| Invariant                       | Deterministic | Probabilistic       |
| ------------------------------- | ------------- | ------------------- |
| No agent self-certifies         | Enforced      | Violations rejected |
| Every mission terminates        | Enforced      | —                   |
| Every mission yields proof      | Enforced      | Proof created       |
| Isolation by default            | Enforced      | —                   |
| Review is a gate                | Enforced      | Judgment inside     |
| Failure is cheap, thrash is not | Enforced      | Retried or split    |

---

## 5. Summary for the Team

**What is already right**

- Deterministic cage around AI
- Per-unit termination
- Independent verification
- Isolation and routing

**What changes directionally**

- Make _Mission_ explicit
- Separate _proof_ from _test_
- Encode risk modes
- Harden protocol & verdict structures

**What must not change**

- Deterministic authority over acceptance
- Fast rejection and regeneration
- Small, terminating units of work

---

## Final Alignment Rule

> **Probabilistic systems explore.  
> Deterministic systems decide.**

Any future design that blurs this line will slow the system and increase risk.
