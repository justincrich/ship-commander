---
title: "Ship Commander 2: Manifesto Alignment Analysis & Tactical Recommendations"
date: "2026-02-10"
time: "16:25"
category: "research"
tags:
  ["ship-commander2", "manifesto", "architecture", "tactical-recommendations"]
status: "complete"
research_type: "codebase_analysis"
iterations: 1
sources_consulted: 4
confidence: "HIGH"
team_tier: 4
teammates_used: ["research-director", "codebase-analysis"]
---

# Ship Commander 2: Manifesto Alignment Analysis & Tactical Recommendations

## Executive Summary

**Ship Commander 2** represents a **significant evolutionary step** toward the manifesto's vision of single-loop development, with **strong foundations already in place** but **critical gaps remaining** in mission lifecycle enforcement, verification gate independence, the separation of Captain and Commander authority, and the distinction between commissions (PRD-level initiatives) and missions (individual coding tasks).

The codebase has successfully **escaped the KB ecosystem's complexity trap** while preserving essential patterns. However, it currently operates as a **"light manifesto"** system—implementing the vocabulary and data structures without fully enforcing the behavioral constraints that make manifesto principles work in practice.

**Key roles that must be clearly separated**:

- **Admiral (Human)**: Final approval authority. No mission dispatched without Admiral sign-off. Can provide feedback (reconvenes planning loop), shelve plans, or approve for execution. Commissions the work (hence "commission"). **Sizing rule**: a commission should amortize approval cost over ≥5 missions — if you're approving commissions that produce 1-2 missions, you're slipping back toward big-loop behavior.
- **Captain**: Owns **functional requirements / commission-level adherence** — ensures what gets built matches the commission's use cases and acceptance criteria. Strategic oversight. The _what_.
- **Commander**: The **orchestrator** — decomposes commissions into missions, evaluates, and sequences missions by importance and dependency. Owns **technical requirements / implementation**. Dispatches agents, enforces gates, determines when missions are complete. Like an engineering manager. The _how_ and _when_.
- **Design Officer**: Owns **design requirements** — UX/UI, component architecture, design system adherence. The _look and feel_.

The Admiral _commissions and approves_; the Captain is _in charge of commission adherence_; the Commander is _in control of mission execution_; the Design Officer is _in control of design_. Before execution begins, all three agents must convene in the **Captain's Ready Room** — a planning loop for a specific commission where they collaboratively decompose use cases into missions, draft, negotiate, and sign off on mission specs with clear domain ownership (like KB's `kb-project-analyze`). The resulting mission manifest goes to the Admiral for final approval. Plans can be saved, shelved, and executed later.

Today these roles are collapsed into a single intake pipeline with no distinct Commander orchestrating mission lifecycle, no collaborative planning loop, and no formal commission concept connecting the PRD to its missions.

**Key finding**: Ship Commander 2 is **80% aligned** with manifesto structures but only **40% aligned** with manifesto behavioral enforcement. The gap is not architectural—it's the **absence of a Commander** that independently orchestrates and verifies mission execution, the **absence of a Ready Room** where specialists collaboratively plan before execution begins, and the **absence of a commission concept** that ties missions back to their originating PRD.

---

## Execution Model: Harness Invocation vs Conditional Logic

Throughout this document, every process and step is labeled with one of three execution types:

| Label | Meaning | Example |
|-------|---------|---------|
| **`[Harness → Agent Name]`** | AI agent invoked via a coding harness session. The harness and model are **interchangeable** — configured via role/app settings (Claude Code, Codex, OpenCode, or any agentic IDE). | Captain analyzing functional requirements |
| **`[Conditional Logic]`** | Deterministic code — state machine transitions, shell command execution, exit code checking, message routing, lock management. No AI involved. | Verification gate running `pnpm test` and checking exit code |
| **`[Human Input → Admiral]`** | System presents information to the human (Admiral) and waits for a decision. | Admiral reviewing mission manifest for approval |

> **Harness interchangeability**: Ship Commander 2 is **harness-agnostic**. Agent sessions are spawned via a `HarnessSession` abstraction. The concrete harness (Claude Code, Codex, OpenCode, Windsurf, etc.) and model (Opus, Sonnet, GPT-4, etc.) are configured per-role in app settings — not hardcoded. All references to "harness session" in this document mean "any agentic coding IDE session."

> **Reference materials**: Required Reading (brain/ document references) and Commission/Mission terminology definitions are in **Appendix A** and **Appendix B** at the end of this document.

---

## 1. What Ship Commander 2 Does Right (DO NOT CHANGE)

### 1.1 Deterministic State Machine Foundation ✅

**Status**: **EXCELLENT - PRESERVE AS-IS**

The state management layer is rock-solid and fully deterministic:

```typescript
// src/domain/types.ts lines 7-15
export type MissionStatus =
  | 'backlog'
  | 'in_progress'
  | 'review'
  | 'approved'
  | 'done'
  | 'halted';

// src/state/state-service.ts lines 13-20
async transitionMission(
  id: string,
  from: MissionStatus,
  to: MissionStatus,
  reason: string
): Promise<void> {
  assertLegalMissionTransition(from, to);  // ✅ Deterministic enforcement
  await this.beads.setMissionStatus(id, to, reason);
}
```

**Why this works**:

- Finite state machine with legal transition enforcement
- State transitions are **deterministic assertions**, not AI judgment
- Beads backend provides persistent, queryable state
- No probabilistic agent can override the state machine

**Action**: **DO NOT TOUCH**. This is the load-bearing structure.

---

### 1.2 Mission Classification System ✅

**Status**: **EXCELLENT - MANIFESTO-ALIGNED**

The risk classification system directly implements manifesto's **Red Alert vs Standard Ops** distinction:

```typescript
// src/domain/types.ts lines 96-104
export interface MissionClassification {
  class: MissionClass; // 'RED_ALERT' | 'STANDARD_OPS'
  risk_score: number;
  reasons: string[];
  required_gates: string[];
  loop_mode: "RGR" | "OPS"; // ✅ Red-Green-Refactor vs Standard Ops
  demo_token_required: boolean;
  test_required: boolean;
}
```

**Why this matters**:

- **Explicit risk encoding**: Not all missions need full TDD
- **Demo token awareness**: System knows when proof is required
- **Gate tailoring**: High-risk missions get stricter gates

**Action**: **EXPAND THIS**. The classifier works but isn't used everywhere it should be.

---

### 1.3 Per-AC TDD State Tracking ✅

**Status**: **EXCELLENT - KB PATTERN PRESERVED**

Ship Commander 2 preserved the KB ecosystem's most successful pattern—**per-acceptance-criteria TDD phase tracking**:

```typescript
// src/domain/types.ts lines 336-344
export interface AcExecutionState {
  acId: string;
  index: number;
  title: string;
  phase: AcTddState; // 'red' | 'verify_red' | 'green' | 'verify_green' | 'refactor' | 'verify_refactor'
  attempt: number;
  completed: boolean;
  attempts: ExecutionAttempt[];
}
```

**Why this is critical**:

- Enforces **one AC = one loop** manifesto principle
- Each phase has independent verification gates
- Attempt counting prevents infinite loops
- Fully deterministic state machine per AC

**Action**: **DO NOT CHANGE**. This is the heartbeat of single-loop enforcement.

> **Progressive disclosure**: See `brain/docs/TDD-METHODOLOGY.md` for the full TDD cycle specification this structure implements. See `brain/skills/kb-dispatch/SKILL.md` "REQUIRED WORKFLOW" for the per-AC: Write test → POST status → WAIT for verification → Implement → POST status → WAIT for verification pattern.

---

### 1.4 Demo Token Primitive ✅

**Status**: **PRESENT - MANIFESTO-ALIGNED**

The system has the **DemoToken** type as a first-class concept:

```typescript
// src/domain/types.ts lines 244-248
export interface DemoToken {
  type: "command" | "screenshot" | "trace" | "diff" | "storybook";
  how_to_verify: string;
  expected: string;
}
```

**Why this matters**:

- Generalizes "proof" beyond tests (manifesto requirement)
- Supports multiple verification modalities
- Structured and machine-readable

**Gap**: **Not enforced**. Demo tokens are optional decorations rather than mandatory exit criteria.

**V1 formalization**: Section 4.8 defines the **V1 Demo Token specification** — a strict Markdown file (`demo/MISSION-<id>.md`) with YAML frontmatter schema, 4 allowed evidence types (`commands`, `tests`, `manual_steps`, `diff_refs`), mode-dependent requirements (RED_ALERT requires `tests`), and deterministic enforcement by the Commander's `DemoTokenValidator`. The in-memory `DemoToken` type above becomes the parsed representation; the Markdown file is the canonical artifact.

---

### 1.5 Event-Driven Architecture ✅

**Status**: **EXCELLENT - OBSERVABILITY FOUNDATION**

The event system provides full observability of mission lifecycle:

```typescript
// src/domain/types.ts lines 250-261
export type RuntimeEventPayload =
  | { channel: 'dispatch'; data: DispatchEventPayload }
  | { channel: 'gate'; data: GateEventPayload }
  | { channel: 'review'; data: ReviewEventPayload }
  | { channel: 'integrate'; data: IntegrateEventPayload }
  ...
```

**Why this works**:

- Every state transition, gate execution, and agent action emits events
- Enables TUI dashboards, logging, and post-hoc analysis
- Provides "observable system effect" evidence manifesto requires

**Action**: **EXPAND**. Add events for mission termination, demo token production, and attempt exhaustion.

---

## 2. Critical Gaps: What's Missing for Full Manifesto Alignment

### 2.1 MISSION LIFECycle: No Explicit Termination Enforcement ❌

**Status**: **CRITICAL GAP - BLOCKS MANIFESTO COMPLIANCE**

**The problem**: The system **tracks** mission state but does not **enforce** termination rules.

**Evidence from codebase**:

```typescript
// src/domain/types.ts lines 355-372
export interface MissionExecutionSession {
  directiveId: string;
  missionId?: string;
  projectId: string;
  status: SessionStatus;
  revisionCount: number;
  maxRevisions: number; // ✅ Exists but...
  startedAt: string;
  updatedAt: string;
  worktreePath?: string;
  mission: MissionClassification;
  trackState: MissionTrackState;
  activeAgents: AgentAssignment[];
  demoToken?: DemoToken; // ❌ Optional, not required
}
```

**What's missing**:

1. **No `attempt_count` field** - Mission can retry infinitely
2. **No termination condition** - Nothing checks `revisionCount >= maxRevisions`
3. **No mandatory demo token** - `demoToken?` is optional
4. **No "single mission must terminate" invariant** - Multiple missions can run concurrently without coordination

**Legacy KB comparison**:

- KB enforced **per-AC termination** via TDD cycle completion
- KB enforced **issue-level termination** via AC count validation (`completed_acs == total_acs`)
- Ship Commander 2 has the fields but **no enforcement logic**

---

### 2.2 VERIFICATION GATES: No Independent Commander Verification ❌

**Status**: **CRITICAL GAP - BREAKS "DON'T TRUST THE AGENT" PRINCIPLE**

**The problem**: There are **gate types defined** but **no Commander that runs them independently**. Gate execution is the Commander's core responsibility — the Commander must sit between every agent claim and every state transition, running deterministic verification that the agent cannot influence.

**Evidence from codebase**:

```typescript
// src/domain/types.ts lines 34-38
export type GateType =
  | "verify_red"
  | "verify_green"
  | "verify_refactor"
  | "verify_implement";

// lines 139-148
export interface GateEvidence {
  gateType: GateType;
  directiveId: string;
  acId?: string;
  exitCode: number | null;
  classification:
    | "accept"
    | "reject_vanity"
    | "reject_syntax"
    | "reject_failure";
  snippet: string;
  timestamp: string;
  attempt: number;
}
```

**What exists**:

- Gate type definitions
- Evidence structure for recording gate results
- Event types for gate pass/fail

**What's missing**:

1. **No gate execution engine** - There's no `runVerifyRedGate()`, `runVerifyGreenGate()` function owned by the Commander
2. **No independent test execution** - Agents likely run tests themselves; the Commander should run them independently
3. **No vanity test detection** - No check that RED phase tests actually fail
4. **No infrastructure re-check** - No validation that agent-reported infra state matches reality
5. **No gate result enforcement** - Nothing blocks agent progress on gate failure; the Commander should be the gatekeeper

**Legacy KB comparison**:

- KB's **kb-swarm** (the Commander-equivalent) had **explicit verification functions**:
  - `pnpm test -- {test_file}` and exit code checking
  - Vanity test detection (test passes without impl = reject)
  - Infrastructure mismatch checks
  - Flakiness detection (run 3x, pass_rate < 100% = reject)

**Ship Commander 2 has the data structures but no Commander to run the verification engine.**

> **Progressive disclosure**: See `brain/docs/VERIFICATION-GATES.md` for the mandatory gate commands and `brain/docs/TDD-METHODOLOGY.md` Section "The Problem with Traditional Agent TDD" for why independent verification is non-negotiable. See `brain/skills/kb-swarm/SKILL.md` "ROLE BOUNDARIES" table for the exact split between orchestrator (runs gates) and implementer (writes code).

---

### 2.3 ORCHESTRATOR AUTHORITY: No "Cage" Around Probabilistic Execution ❌

**Status**: **CRITICAL GAP - BREAKS DETERMINISTIC/PROBABILISTIC BOUNDARY**

**Role clarity — Captain, Commander, and Design Officer**:

- **Admiral (Human)**: The **final approval authority** and the one who **commissions** the work. No mission is dispatched without Admiral sign-off. The Admiral reviews the Commander's mission manifest, provides feedback, and can reconvene the planning loop at any time. Plans can be saved, shelved, and executed later at the Admiral's discretion. The Admiral is the only role that can greenlight execution.
- **Captain**: Responsible for **commission-level requirement adherence** (functional requirements). The Captain ensures missions collectively satisfy the commission's use cases and acceptance criteria, validates that what gets built matches what was specified, and maintains strategic oversight. Owns the _what_.
- **Commander**: The **orchestrator**. Responsible for **decomposing commissions into missions, evaluating, and sequencing them** (technical requirements / implementation). The Commander takes the commission's use cases and turns them into executable missions — deciding what gets built in what order, splitting or combining use cases as needed, dispatching agents, enforcing gates, and determining when missions are complete. Like an **engineering manager**, the Commander drafts and sequences missions. Owns the _how_ and _when_.
- **Design Officer**: Responsible for **design requirements**. The Design Officer ensures UX/UI concerns, component architecture, and design system adherence are captured in the mission spec. Owns the _look and feel_.

The Admiral _commissions and approves_; the Captain is _in charge of commission adherence_; the Commander is _in control of mission execution_; the Design Officer is _in control of design_.

**The Captain's Ready Room — Collaborative Planning Loop**:

Before missions are dispatched, these three agents must convene in a **planning loop** (the "Captain's Ready Room") to collaboratively plan within the context of a specific **commission** (PRD). The commission provides the use cases and acceptance criteria; the Ready Room decomposes them into executable missions. This mirrors KB's `kb-project-analyze` pattern where specialists collaborate to produce a complete implementation plan:

1. **`[Harness → Captain Agent]`** **Captain** analyzes the commission's use cases and identifies functional requirements — determining which use cases map to which missions and validating that no commission-level acceptance criteria are lost
2. **`[Harness → Design Officer Agent]`** **Design Officer** reviews the commission's use cases for design requirements and UX implications
3. **`[Harness → Commander Agent]`** **Commander** decomposes use cases into missions (splitting, combining, or mapping 1:1 as appropriate), drafts technical requirements, **sequences them by importance and dependency**, and proposes the execution order
4. **`[Harness → All Agents]`** **All three converse** — messaging each other (via **`[Conditional Logic]`** message routing) to resolve conflicts, fill gaps, and reach agreement on the final task list
5. **`[Harness → All Agents]`** **Each agent takes concrete ownership** of their domain within each mission spec:
   - Captain signs off on functional requirements
   - Design Officer signs off on design requirements
   - Commander signs off on technical requirements and owns the final sequencing
6. **`[Harness → Commander Agent]`** **Commander produces the final mission manifest** — the agreed-upon, sequenced task list that adheres to standard task protocol
7. **`[Human Input → Admiral]`** **Admiral (human) reviews the manifest** — the plan is presented for approval before any mission is dispatched
8. **`[Human Input → Admiral]`** **Any agent can ask the Admiral questions mid-planning** — if an agent encounters ambiguity, missing context, or needs a human decision, it can pause and escalate a question to the Admiral. The planning loop suspends until the Admiral responds. The answer is routed back to the asking agent's session (and optionally broadcast to all sessions if it affects the whole plan).
9. **`[Conditional Logic]`** **Admiral feedback reconvenes the loop** — if the Admiral has feedback, objections, or scope changes, the planning loop restarts with that feedback **`[Conditional Logic]`** injected into all three agent sessions
10. **`[Conditional Logic]`** **Plan is saved and can be executed later** — approved or in-progress plans are persisted and can be shelved, resumed, or re-executed at the Admiral's discretion

**Session isolation — each agent in their own harness context**:

Each agent in the Ready Room operates in **their own coding harness session** (separate context window). They do not share a single context. Collaboration happens through **structured message passing between sessions**:

- **Captain** runs in Session A — analyzes PRD, produces functional requirements
- **Design Officer** runs in Session B — analyzes design implications, produces design requirements
- **Commander** runs in Session C — drafts technical requirements, sequences missions
- When agents need to collaborate, they **send messages to each other's sessions** — these arrive as additional messages in the target session's context
- The **planning loop orchestrator** (deterministic) routes messages between sessions and collects outputs, but does **not** inject or merge session contexts
- Each agent's session is **bound to their domain** — they see the PRD, their own work, and messages from other agents, but not the full internal context of other agents

This mirrors how harness-level agent teams work today (e.g., Claude Code Teams, Codex multi-agent, OpenCode agent collaboration): each teammate is a separate agent session, and structured messages deliver between them. Ship Commander 2 must internalize this pattern so the planning loop controls the session lifecycle and message routing at the system level — independent of which harness is used.

Today this collaborative planning happens at the harness level. Ship Commander 2 must **internalize this process** so the planning loop is a first-class system concept, not an external harness dependency. The harness is interchangeable — what matters is the session abstraction (`HarnessSession`) and the message routing protocol.

> **Progressive disclosure**: See `brain/skills/kb-project-analyze/SKILL.md` for the team architecture (product-manager + engineering-manager + ui-designer) that is the closest existing pattern to the Ready Room. See `brain/skills/kb-project-analyze/docs/task-description-format.md` for the standard task protocol (v4.0 TDD-integrated) that missions must conform to. See `brain/docs/UNIVERSAL-TASK-SCHEMA.md` for the portable task schema.

**The problem**: The **intake orchestrator** (src/intake/orchestrator.ts) is a **pipeline coordinator**, not a **mission execution authority**. It conflates the Captain and Commander roles — there is no distinct Commander that owns mission lifecycle from dispatch through verification.

**Evidence from codebase**:

```typescript
// src/intake/orchestrator.ts lines 125-307
export async function runIntakePipeline(
  options: IntakePipelineOptions,
): Promise<IntakePipelineResult> {
  // [Conditional Logic] ✅ Deterministic: Parse PRD
  const parsed = parsePRD(readPrd(options.repoRoot, options.prdPath));

  // [Harness → Captain Agent] ❌ Probabilistic: Run Captain (AI agent)
  captain = await runCaptain({
    /* ... */
  });

  // [Harness → Commander + Design Officer Agents] ❌ Probabilistic: Run in parallel (AI agents)
  const parallel = await runSpecialistsInParallel([
    /* ... */
  ]);

  // [Harness → Synthesizer Agent] ❌ Probabilistic: Run Synthesizer (AI agent)
  const synthesized = await runSynthesizer({
    /* ... */
  });

  // [Human Input → Admiral] ✅ Deterministic: Approval gate
  const approval = await runApprovalGate({
    /* ... */
  });

  // [Conditional Logic] ❌ No verification: Commit missions (creates work)
  const commit = await commitMissions({
    /* ... */
  });
}
```

**What's wrong**:

1. **No Commander exists as a distinct orchestrator** - The intake pipeline treats the Commander as just another specialist agent, not as the mission execution authority
2. **No collaborative planning loop** - Captain, Commander, and Design Officer run in parallel but don't converse or negotiate on mission specs. There is no "Ready Room" where they collaborate to draft and validate the task list
3. **No domain ownership per mission** - No clear delineation of who owns functional (Captain), design (Design Officer), and technical (Commander) requirements within each mission spec
4. **No commission concept** - The PRD is parsed into use cases but there's no formal commission type that persists beyond intake and connects use cases to the missions that implement them. The system loses traceability from PRD → use case → mission
5. **No mission execution orchestrator exists** - Who runs RED → GREEN → REFACTOR? Who checks gates? This is the Commander's job.
6. **No "don't trust the agent" enforcement** - Agent claims are not independently verified
7. **Probabilistic agents can emit state changes** - No deterministic gate between agent and state transition
8. **Captain and Commander roles are collapsed** - The Captain should validate _what_ to build; the Commander should own _how_ and _when_ it gets built
9. **Planning is a harness-level concern, not a system concern** - The collaborative planning pattern (harness-level agent teams) is external to Ship Commander 2, meaning the system can't enforce standard task protocol

**What's needed**:

**A. Captain's Ready Room (Planning Loop)**:

1. **`[Conditional Logic]`** A **commission** (PRD) is created with use cases and acceptance criteria
2. **`[Harness → Captain + Commander + Design Officer Agents]`** Captain, Commander, and Design Officer convene to plan within the commission's context
3. **`[Harness → All Agents]`** Each takes ownership of their domain: functional (Captain), design (Design Officer), technical/implementation (Commander)
4. **`[Harness → Commander Agent]`** Commander **decomposes use cases into missions** — splitting, combining, or mapping 1:1 as appropriate
5. **`[Harness → All Agents]`** via **`[Conditional Logic]`** message routing — Agents **message each other** to resolve gaps, conflicts, and dependencies
6. **`[Harness → Captain Agent]`** Captain validates **use case coverage** — every commission AC maps to at least one mission
7. **`[Harness → Commander Agent]`** Commander drafts the mission manifest — sequenced by importance and dependency
8. **`[Conditional Logic]`** All three sign off — missions adhere to standard task protocol and cover the commission (consensus check is deterministic)
9. **`[Conditional Logic]`** Planning loop iterates until consensus (like `kb-project-analyze`)
10. **`[Human Input → Admiral]`** **Admiral (human) approves the manifest** — no mission dispatched without Admiral sign-off
11. **`[Conditional Logic]`** Admiral feedback **reconvenes the planning loop** — agents re-enter the Ready Room with feedback injected into sessions
12. **`[Conditional Logic]`** Plans are **saved and persistable** — can be shelved, resumed, and executed at a later time

**B. Commander as mission execution orchestrator**:

1. **`[Conditional Logic]`** Receives the agreed mission manifest from the Ready Room
2. **`[Conditional Logic]`** **Sequences and prioritizes missions** based on dependencies, complexity, and constraints
3. **`[Harness → Implementer Agent]`** Dispatches agent to implement (probabilistic)
4. **`[Harness → Implementer Agent]`** **Agent claims "RED_COMPLETE"** via protocol event
5. **`[Conditional Logic]`** **Commander runs VERIFY_RED gate independently** (shell commands + exit codes — deterministic)
6. **`[Conditional Logic]`** **Gate result determines next state** (deterministic state machine transition)
7. Repeat for GREEN, REFACTOR (same **`[Harness]`** → **`[Conditional Logic]`** pattern)
8. **`[Conditional Logic]`** **Enforce termination** when `attempt_count >= max` or `demo_token` not produced
9. **`[Conditional Logic]`** **Reports mission status back to Captain** for PRD-level requirement tracking

**Legacy KB comparison**:

- KB had **kb-swarm** as master orchestrator (≈ Commander role)
- **kb-dispatch** as mission dispatcher (≈ Commander dispatching individual missions)
- **Comment polling loop** as coordination mechanism
- **VERIFY_RED / VERIFY_GREEN gates** as independent checks

Ship Commander 2 has **no equivalent** to the kb-swarm / kb-dispatch monitoring loop. The Commander role exists as a specialist agent but not as the orchestrator it needs to be.

> **Progressive disclosure**: See `brain/skills/kb-swarm/SKILL.md` for the full Commander-equivalent orchestrator pattern and `brain/skills/kb-swarm/docs/monitoring-loop.md` for the wave execution + AC progress tracking algorithm. See `brain/skills/kb-dispatch/SKILL.md` for the implementer-under-Commander pattern (FOLLOWER mode, blocked until orchestrator approves).

---

### 2.4 PROTOCOL: No Structured Inter-Agent Communication ❌

**Status**: **HIGH PRIORITY GAP - PREVENTS OBSERVABLE COORDINATION**

**The problem**: Agents communicate via **return values and function calls**, not **structured protocol events**. The Commander needs a protocol layer to receive agent claims, dispatch verification, and coordinate mission flow. Without it, the Commander has no way to observe or control what agents are doing.

**Evidence from codebase**:

```typescript
// src/intake/orchestrator.ts lines 144-163
const captain = await runCaptain({
  projectId: options.projectId,
  missionId: "INTAKE-CAPTAIN",
  context: parsed,
  cwd: options.repoRoot,
  timeoutMs,
  settings,
  harnesses: options.harnesses,
  eventBus: options.eventBus,
});
// ❌ Captain output is just a return value, not a persisted event
```

**What's missing**:

1. **No protocol versioning** - No `{protocol_version: "1.0"}` in events
2. **No structured event schema** - Events are ad-hoc, not validated against schema
3. **No event persistence** - Events are emitted to EventBus but not stored for replay
4. **No "agent claim" pattern** - Agent doesn't post "I claim X" and wait for Commander verification

**Legacy KB comparison**:

- KB used **Linear comments as inter-agent protocol**:
  - `TDD_PHASE: RED_COMPLETE`
  - `AC: 3/5`
  - `TEST_FILE: path/to/test.spec.ts`
  - `VERIFY_RED_PASSED` / `VERIFY_RED_FAILED`

**Manifesto-structures requirement**:

```markdown
## B. Protocol Event

**Deterministic**

- `event_type`
- `mission_id`
- `timestamp`
- `protocol_version`

**Probabilistic**

- Agent-supplied descriptions

**What's different**

- JSON schema instead of regex parsing
- Versioned protocol
```

Ship Commander 2 has **none of this**. Agent outputs are TypeScript return values, not persisted protocol events.

> **Progressive disclosure**: See `brain/docs/SHARED-FORMATS.md` for existing agent ↔ orchestrator communication standards (promise formats, intervention patterns). See `brain/skills/kb-dispatch/docs/comment-formats.md` for the exact TDD_PHASE comment protocol used in KB (the structured inter-agent protocol this section describes).

---

### 2.5 WORK ISOLATION: No Surface-Area Locking ❌

**Status**: **MEDIUM PRIORITY GAP - RISK OF CONCURRENT CORRUPTION**

**The problem**: Worktree isolation exists but **no surface-area locking** to prevent multiple agents from modifying the same subsystem. The Commander, as the orchestrator responsible for sequencing missions, must enforce isolation at dispatch time — not after conflicts arise.

**Evidence from codebase**:

```typescript
// src/domain/types.ts lines 364
worktreePath?: string;  // ❌ Optional, no evidence of enforcement
```

**What's missing**:

1. **No lock acquisition before mission start** - Commander doesn't check what other missions are touching before dispatching
2. **No surface-area declaration** - Missions don't declare "I will modify auth/\*"
3. **No conflict detection up front** - Only during merge phase, not at Commander dispatch time

**Manifesto-structures requirement**:

```markdown
### D. Work Isolation & Concurrency Control

**Deterministic aspects**

- Worktree per issue/mission
- Lock acquisition/release rules (future)

**What's different**

- Explicit _surface-area locks_ (DB, auth, core utils)
- Orchestrator enforcement instead of "best effort"
```

Ship Commander 2 has worktree isolation but **no Commander-enforced lock acquisition at dispatch time**.

---

## 3. Comparison: Ship Commander 2 vs Legacy KB Ecosystem

### 3.1 What Was Preserved ✅

| KB Pattern                     | Ship Commander 2 Implementation                       | Status                                                        |
| ------------------------------ | ----------------------------------------------------- | ------------------------------------------------------------- |
| **Per-AC TDD cycle**           | `AcExecutionState` with phase tracking                | ✅ **PRESERVED** - Lines 336-344                              |
| **Mission status FSM**         | `MissionStatus` with legal transitions                | ✅ **PRESERVED** - Lines 7-15, `assertLegalMissionTransition` |
| **Risk classification**        | `MissionClassification` with RED_ALERT / STANDARD_OPS | ✅ **PRESERVED & IMPROVED** - Lines 96-104                    |
| **Demo token concept**         | `DemoToken` type definition                           | ✅ **PRESERVED** - Lines 244-248                              |
| **Event-driven observability** | `RuntimeEventPayload` with channels                   | ✅ **PRESERVED & EXPANDED** - Lines 250-261                   |

**Assessment**: Ship Commander 2 successfully **extracted the kernel** of KB's best ideas and cleaned them up.

---

### 3.2 What Was Abandoned ❌

| KB Pattern                         | KB Role                         | Ship Commander 2 Status                      | Impact                                                           |
| ---------------------------------- | ------------------------------- | -------------------------------------------- | ---------------------------------------------------------------- |
| **Linear as state backbone**       | Infrastructure                  | **REMOVED** - Now uses Beads (local state)   | ✅ **IMPROVEMENT** - No external dependency, faster              |
| **Comment-based protocol**         | Commander ↔ Agent communication | **REMOVED** - No inter-agent messaging       | ❌ **REGRESSION** - Commander has no way to observe agent claims |
| **Per-AC polling loop**            | Commander monitoring pattern    | **REMOVED** - No monitoring orchestrator     | ❌ **REGRESSION** - Commander can't enforce "don't trust agent"  |
| **Independent verification gates** | Commander gate execution        | **REMOVED** - Gates defined but not executed | ❌ **CRITICAL REGRESSION** - Commander has no deterministic cage |
| **WSJF prioritization**            | Captain sequencing input        | **REMOVED** - No project prioritization      | ⚠️ **CONTEXT DEPENDENT** - Maybe out of scope for v1             |

**Assessment**: Some removals are improvements (Linear dependency), but the **Commander's core capabilities** — monitoring, verification, and gate enforcement — were lost. The Commander role (≈ KB's kb-swarm + kb-dispatch) needs to be rebuilt as a distinct orchestrator, separate from the Captain's PRD-level oversight.

---

### 3.3 What's Improved 🚀

| Area                    | KB Approach                    | Ship Commander 2 Approach                          | Why Better                                |
| ----------------------- | ------------------------------ | -------------------------------------------------- | ----------------------------------------- |
| **State persistence**   | Linear comments (remote, slow) | Beads (local, fast)                                | No network latency, offline-capable       |
| **Type safety**         | Regex parsing from comments    | TypeScript types, Zod validation                   | Catch errors at compile time, not runtime |
| **Risk classification** | Implicit via agent choice      | Explicit `MissionClassification` type              | Clear, auditable, machine-readable        |
| **Event system**        | Ad-hoc console logs            | Structured `RuntimeEventPayload`                   | Queryable, dashboard-ready                |
| **Mission modes**       | Not explicitly modeled         | `RED_ALERT` vs `STANDARD_OPS` with different gates | Manifesto-aligned, risk-appropriate       |

**Assessment**: Ship Commander 2 successfully **modernized the kernel** while shedding complexity.

---

## 4. Tactical Recommendations: Path to Full Manifesto Alignment

### 4.1 CRITICAL: Build the Commander — Mission Execution Orchestrator (Priority: P0)

**Problem**: No Commander exists to orchestrate mission execution. The Captain validates _what_ to build (PRD adherence), but there is no distinct Commander that owns _how_ and _when_ — building, evaluating, sequencing, and verifying missions.

**Solution**: Build the Commander as `MissionExecutionOrchestrator` — the deterministic authority that sits between agents and state transitions. The Commander receives validated mission specs from the Captain and owns the entire execution lifecycle:

```typescript
// NEW FILE: src/execution/mission-orchestrator.ts

export class MissionExecutionOrchestrator {
  async executeMission(missionId: string): Promise<MissionResult> {
    const spec = await this.beads.getMissionSpec(missionId);
    const session = await this.createSession(missionId, spec);

    // [Conditional Logic] ENFORCE: Single mission must terminate
    while (session.attemptCount < spec.maxRevisions) {
      for (const ac of spec.acceptanceCriteria) {
        const acState = session.trackState.acs[ac.index];

        // RED PHASE
        if (acState.phase === "red") {
          await this.dispatchAgent(session, ac, "red"); // [Harness → Implementer Agent]
          await this.waitForAgentClaim(ac.id, "RED_COMPLETE"); // [Conditional Logic] poll for claim

          // [Conditional Logic] CRITICAL: Independent verification
          const result = await this.runVerifyRedGate(ac.id); // [Conditional Logic] shell cmd + exit code
          if (!result.passed) {
            await this.handleGateFailure(session, ac, result); // [Conditional Logic]
            break; // Restart AC or terminate mission
          }
          await this.transitionAcPhase(ac.id, "green"); // [Conditional Logic] state machine
        }

        // GREEN PHASE
        if (acState.phase === "green") {
          await this.dispatchAgent(session, ac, "green"); // [Harness → Implementer Agent]
          await this.waitForAgentClaim(ac.id, "GREEN_COMPLETE"); // [Conditional Logic] poll for claim

          const result = await this.runVerifyGreenGate(ac.id); // [Conditional Logic] shell cmd + exit code
          if (!result.passed) {
            await this.handleGateFailure(session, ac, result); // [Conditional Logic]
            break;
          }
          await this.transitionAcPhase(ac.id, "refactor"); // [Conditional Logic] state machine
        }

        // REFACTOR PHASE (optional)
        // ...
      }

      // [Conditional Logic] All ACs complete - validate demo token file (Section 4.8)
      if (spec.classification.demo_token_required) {
        const validation = await this.demoTokenValidator.validate(
          missionId, spec.classification, session.worktreePath);
        if (!validation.valid) {
          throw new Error(`Mission ${missionId}: ${validation.message}`);
        }
      }

      return { status: "completed", session };
    }

    // [Conditional Logic] Exhausted attempts
    return { status: "halted", reason: "max_revisions_exceeded", session };
  }

  private async runVerifyRedGate(acId: string): Promise<GateResult> {
    // [Conditional Logic] INDEPENDENT TEST EXECUTION - Don't trust agent's claim
    const ac = await this.beads.getAc(acId);
    const result = await this.runCommand(ac.testFile); // [Conditional Logic] shell execution

    // [Conditional Logic] VANITY TEST DETECTION
    if (result.exitCode === 0) {
      return {
        passed: false,
        classification: "reject_vanity",
        reason: "Test passed without implementation - vanity test detected",
      };
    }

    // [Conditional Logic] SYNTAX CHECK
    if (result.stderr.includes("SyntaxError")) {
      return {
        passed: false,
        classification: "reject_syntax",
        reason: "Test has syntax errors",
      };
    }

    return { passed: true, classification: "accept" };
  }
}
```

**Key invariants the Commander enforces**:

1. **`[Harness → Implementer Agent]`** dispatch → **`[Harness → Implementer Agent]`** claim → **`[Conditional Logic]`** Commander verification → **`[Conditional Logic]`** State transition
2. No agent can mark its own work complete — only the Commander **`[Conditional Logic]`** transitions state
3. Vanity tests rejected **`[Conditional Logic]`** deterministically by the Commander's gate engine
4. Mission terminates after `maxRevisions` — **`[Conditional Logic]`** Commander enforces termination
5. Captain receives mission status reports — **`[Conditional Logic]`** Commander reports up, agents report to Commander

**Relationship to Captain and Commission**: The Captain validates the mission spec against the commission (use-case coverage, acceptance criteria quality, risk classification). Once the commission's mission manifest is approved by the Admiral, the Commander receives the manifest and owns each mission from dispatch through completion/halt. The Commander reports mission status back to the Captain, who tracks progress against the commission's use cases — ensuring that completed missions collectively satisfy the commission's acceptance criteria.

> **Progressive disclosure**: See `brain/skills/kb-swarm/SKILL.md` for the full orchestrator pattern this Commander must implement. See `brain/skills/kb-swarm/docs/monitoring-loop.md` for the wave execution algorithm (dispatch → monitor → verify → proceed). See `brain/docs/TDD-METHODOLOGY.md` "Core Principles" — especially principle 2: "Orchestrator verifies (doesn't trust agent claims)."

**Implementation effort**: ~3-5 days

---

### 4.1b CRITICAL: Implement Captain's Ready Room — Collaborative Planning Loop (Priority: P0)

**Problem**: Captain, Commander, and Design Officer currently run in parallel as independent specialist agents during intake. They don't converse, negotiate, or collaboratively decompose a commission's use cases into missions. There is no planning loop where they reach consensus on use-case-to-mission mapping, sequencing, and domain ownership. There is no formal commission concept connecting the PRD to its missions. Today this collaborative process happens at the harness level (harness-native agent teams) rather than as a first-class system concept.

**Solution**: Build the **Captain's Ready Room** — a planning loop for a specific **commission** where the three senior agents collaborate to decompose use cases into missions and produce a validated mission manifest before any execution begins. The commission (PRD) provides the use cases and acceptance criteria; the Ready Room produces missions that collectively cover them. This mirrors KB's `kb-project-analyze` pattern.

```typescript
// NEW FILE: src/planning/ready-room.ts

// COMMISSION — the PRD-level initiative that the Ready Room is planning against
export interface Commission {
  id: string; // Unique commission ID
  prdPath: string; // Path to source PRD
  productName: string;
  description: string;
  useCases: UseCaseRef[]; // Use cases from the PRD
  inScope: string[];
  outScope: string[];
  createdAt: string;
  status: "planning" | "approved" | "executing" | "completed" | "shelved";
}

export interface UseCaseRef {
  id: string; // Original use case ID from PRD
  title: string;
  acceptanceCriteria: string[];
  functionalGroup: string;
  missionIds: string[]; // Which missions implement this use case (many-to-many)
}

export interface ReadyRoomMessage {
  from: "captain" | "commander" | "design-officer" | "admiral";
  to: "captain" | "commander" | "design-officer" | "admiral" | "all";
  type:
    | "proposal"
    | "review"
    | "objection"
    | "approval"
    | "revision"
    | "question"
    | "answer";
  missionId?: string;
  domain: "functional" | "technical" | "design";
  content: string;
  timestamp: string;
  questionId?: string; // Links question→answer pairs
}

export interface MissionDraft {
  id: string;
  commissionId: string; // Which commission this mission belongs to
  useCaseRefs: string[]; // Which use case(s) this mission implements (often 1, can be many)
  title: string;
  functionalRequirements: RequirementSignoff; // Captain owns
  designRequirements: RequirementSignoff; // Design Officer owns
  technicalRequirements: RequirementSignoff; // Commander owns
  acceptanceCriteria: AcceptanceCriterion[];
  sequence: number; // Commander determines order (by importance + dependency)
  dependencies: string[]; // Mission IDs this depends on
  importance: "critical" | "high" | "medium" | "low"; // Commander assesses
  signoffs: {
    captain: boolean;
    commander: boolean;
    designOfficer: boolean;
  };
}

export interface RequirementSignoff {
  owner: "captain" | "commander" | "design-officer";
  requirements: string[];
  signedOff: boolean;
  signedOffAt?: string;
}

export type ManifestStatus =
  | "drafting" // Agents still collaborating
  | "awaiting_approval" // Consensus reached, waiting for Admiral
  | "approved" // Admiral approved, ready for execution
  | "feedback" // Admiral provided feedback, reconvening loop
  | "shelved" // Saved for later execution
  | "executing" // Commander is dispatching missions
  | "completed"; // All missions done

export interface MissionManifest {
  id: string; // Unique manifest ID for save/resume
  commission: Commission; // The PRD-level initiative being planned
  missions: MissionDraft[];
  useCaseCoverage: UseCaseCoverageMap; // Tracks which use cases are covered by which missions
  planningMessages: ReadyRoomMessage[]; // Full conversation log
  consensusReached: boolean;
  iterations: number;
  status: ManifestStatus;
  admiralApproval?: {
    approved: boolean;
    feedback?: string; // Admiral's feedback if not approved
    approvedAt?: string;
    approvedBy: string; // Admiral identifier
  };
  savedAt?: string; // When plan was last persisted
  resumedFrom?: string; // Manifest ID if resumed from a saved plan
}

// Tracks which commission use cases are covered by which missions
// Captain validates this is complete before sign-off
export type UseCaseCoverageMap = Record<
  string,
  {
    // keyed by use case ID
    useCaseId: string;
    useCaseTitle: string;
    missionIds: string[]; // Missions that implement this use case
    acceptanceCriteriaCovered: number; // How many of the UC's ACs are covered
    acceptanceCriteriaTotal: number; // Total ACs on this use case
    status: "covered" | "partial" | "uncovered";
  }
>;

export interface AdmiralQuestion {
  text: string; // The question being asked
  context?: string; // Background context for the Admiral
  domain: "functional" | "technical" | "design"; // Which domain this affects
  options?: AdmiralQuestionOption[]; // Optional pre-defined choices
  broadcastAnswer?: boolean; // Should the answer go to all agents?
}

export interface AdmiralQuestionOption {
  label: string; // Short display label (e.g., "OAuth 2.0")
  description: string; // Explanation of what this choice means
}

export class ReadyRoom {
  private messages: ReadyRoomMessage[] = [];
  private missions: Map<string, MissionDraft> = new Map();

  // Each agent operates in their OWN harness session (separate context)
  private captainSession: HarnessSession; // Session A
  private designSession: HarnessSession; // Session B
  private commanderSession: HarnessSession; // Session C

  async runPlanningLoop(
    commission: Commission, // The PRD-level initiative being planned
    maxIterations: number = 5,
  ): Promise<MissionManifest> {
    // [Harness → Captain/Design Officer/Commander Agents] Spawn each in isolated session
    // Each session receives the commission (PRD context + use cases) as input
    this.captainSession = await this.spawnSession("captain", { commission }); // [Harness → Captain Agent]
    this.designSession = await this.spawnSession("design-officer", {
      commission,
    }); // [Harness → Design Officer Agent]
    this.commanderSession = await this.spawnSession("commander", {
      commission,
    }); // [Harness → Commander Agent]

    for (let iteration = 0; iteration < maxIterations; iteration++) {
      // STEP 1 [Harness → Captain Agent]: Analyzes commission use cases → functional requirements
      // Captain maps use cases to potential missions, validates AC coverage
      // If agent needs clarification, it returns a question instead of analysis
      const captainAnalysis =
        await this.captainSession.run("analyze_functional"); // [Harness → Captain Agent]
      if (captainAnalysis.questions?.length) {
        await this.routeQuestionsToAdmiral(
          captainAnalysis.questions,
          "captain",
        ); // [Human Input → Admiral]
      }
      this.recordMessages(captainAnalysis.messages); // [Conditional Logic]
      this.updateMissions(captainAnalysis.missionDrafts, "functional"); // [Conditional Logic]

      // STEP 2 [Harness → Design Officer Agent]: Reviews missions → design requirements
      await this.sendToSession(this.designSession, {
        // [Conditional Logic] message routing
        from: "captain",
        content: captainAnalysis.summary, // Not full context, just structured output
      });
      const designAnalysis = await this.designSession.run("analyze_design"); // [Harness → Design Officer Agent]
      if (designAnalysis.questions?.length) {
        await this.routeQuestionsToAdmiral(
          designAnalysis.questions,
          "design-officer",
        ); // [Human Input → Admiral]
      }
      this.recordMessages(designAnalysis.messages); // [Conditional Logic]
      this.updateMissions(designAnalysis.missionDrafts, "design"); // [Conditional Logic]

      // STEP 3 [Harness → Commander Agent]: Decomposes use cases into missions + sequences
      // Commander may split, combine, or reorder use cases into missions
      await this.sendToSession(this.commanderSession, {
        // [Conditional Logic] message routing
        from: "captain",
        content: captainAnalysis.summary,
      });
      await this.sendToSession(this.commanderSession, {
        // [Conditional Logic] message routing
        from: "design-officer",
        content: designAnalysis.summary,
      });
      const commanderAnalysis =
        await this.commanderSession.run("draft_missions"); // [Harness → Commander Agent]
      if (commanderAnalysis.questions?.length) {
        await this.routeQuestionsToAdmiral(
          commanderAnalysis.questions,
          "commander",
        ); // [Human Input → Admiral]
      }
      this.recordMessages(commanderAnalysis.messages); // [Conditional Logic]
      this.updateMissions(commanderAnalysis.missionDrafts, "technical"); // [Conditional Logic]

      // STEP 4 [Conditional Logic]: Check for consensus — all three signed off
      if (this.hasConsensus()) {
        // [Conditional Logic]
        const manifest = this.buildManifest(iteration + 1, "awaiting_approval");

        // STEP 5 [Human Input → Admiral]: ADMIRAL APPROVAL GATE
        await this.savePlan(manifest); // [Conditional Logic] persistence

        const admiralDecision = await this.requestAdmiralApproval(manifest); // [Human Input → Admiral]

        if (admiralDecision.approved) {
          manifest.status = "approved"; // [Conditional Logic]
          manifest.admiralApproval = admiralDecision;
          await this.savePlan(manifest); // [Conditional Logic] persistence
          return manifest;
        }

        // [Conditional Logic] ADMIRAL FEEDBACK → reconvene the planning loop
        // Inject Admiral feedback into all three agent sessions
        await this.sendToSession(this.captainSession, {
          // [Conditional Logic] message routing
          from: "admiral",
          content: admiralDecision.feedback!,
        });
        await this.sendToSession(this.designSession, {
          // [Conditional Logic] message routing
          from: "admiral",
          content: admiralDecision.feedback!,
        });
        await this.sendToSession(this.commanderSession, {
          // [Conditional Logic] message routing
          from: "admiral",
          content: admiralDecision.feedback!,
        });
        // Loop continues — agents will see Admiral feedback and revise
        continue;
      }

      // STEP 6 [Conditional Logic]: If objections exist, route messages between sessions
      // Each agent stays in their own session; messages are delivered as
      // additional messages to the target session's context
      await this.resolveObjections(); // [Conditional Logic]

      // [Conditional Logic] Auto-save plan state after each iteration (enables resume)
      await this.savePlan(this.buildManifest(iteration + 1, "drafting"));
    }

    // [Conditional Logic] Max iterations reached without consensus — save for later
    const manifest = this.buildManifest(maxIterations, "shelved");
    await this.savePlan(manifest);
    return manifest;
  }

  // [Conditional Logic] PLAN PERSISTENCE — save, resume, shelve
  async savePlan(manifest: MissionManifest): Promise<void> {
    manifest.savedAt = new Date().toISOString();
    await this.beads.saveMissionManifest(manifest);
  }

  static async resumePlan(
    manifestId: string,
    beads: BeadsService,
  ): Promise<ReadyRoom> {
    const saved = await beads.getMissionManifest(manifestId);
    const room = new ReadyRoom();
    room.messages = saved.planningMessages;
    room.missions = new Map(saved.missions.map((m) => [m.id, m]));
    // Re-spawn sessions with saved context
    // Agents resume with their prior work + any Admiral feedback
    return room;
  }

  private async requestAdmiralApproval(
    manifest: MissionManifest,
  ): Promise<{ approved: boolean; feedback?: string }> {
    // Present manifest to Admiral (human) via TUI or CLI prompt
    // Admiral can: approve, provide feedback, or shelve for later
    return this.humanApprovalGate.request(manifest);
  }

  // [Human Input → Admiral] AGENT → ADMIRAL QUESTIONS — any agent can ask the Admiral during planning
  private async routeQuestionsToAdmiral(
    questions: AdmiralQuestion[],
    fromAgent: "captain" | "commander" | "design-officer",
  ): Promise<void> {
    for (const question of questions) {
      // [Conditional Logic] Record question in planning log
      const questionId = `q-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`;
      this.recordMessages([
        {
          from: fromAgent,
          to: "admiral",
          type: "question",
          domain: question.domain,
          content: question.text,
          timestamp: new Date().toISOString(),
          questionId,
        },
      ]);

      // [Human Input → Admiral] SUSPEND planning loop — present question to Admiral via TUI
      const answer = await this.humanQuestionGate.ask({
        questionId,
        fromAgent,
        text: question.text,
        context: question.context,
        options: question.options, // Optional pre-defined choices
        domain: question.domain,
      });

      // [Conditional Logic] Record Admiral's answer
      this.recordMessages([
        {
          from: "admiral",
          to: question.broadcastAnswer ? "all" : fromAgent,
          type: "answer",
          domain: question.domain,
          content: answer.text,
          timestamp: new Date().toISOString(),
          questionId,
        },
      ]);

      // [Conditional Logic] Route answer to the asking agent's session
      await this.sendToSession(this.getSession(fromAgent), {
        from: "admiral",
        content: answer.text,
      });

      // [Conditional Logic] If answer affects all agents, broadcast to all sessions
      if (question.broadcastAnswer || answer.broadcastToAll) {
        for (const session of [
          this.captainSession,
          this.designSession,
          this.commanderSession,
        ]) {
          if (session !== this.getSession(fromAgent)) {
            await this.sendToSession(session, {
              from: "admiral",
              content: `[Re: ${fromAgent}'s question] ${answer.text}`,
            });
          }
        }
      }
    }
  }

  // [Conditional Logic] — pure boolean checks, no AI
  private hasConsensus(): boolean {
    // All missions signed off AND all commission use cases covered
    const allSignedOff = Array.from(this.missions.values()).every(
      (m) =>
        m.signoffs.captain && m.signoffs.commander && m.signoffs.designOfficer,
    );
    const allUseCasesCovered = this.getUseCaseCoverage().every(
      (uc) => uc.status === "covered",
    );
    return allSignedOff && allUseCasesCovered;
  }

  // [Conditional Logic] routing + [Harness → Agent] for responses
  private async resolveObjections(): Promise<void> {
    const openObjections = this.messages.filter(
      (m) => m.type === "objection" && !this.isResolved(m),
    );

    for (const objection of openObjections) {
      // Route objection to the appropriate domain owner's session
      const response = await this.getResponse(objection);
      this.recordMessages([response]);
    }
  }
}
```

**Domain ownership per mission (within commission context)**:

| Domain                      | Owner          | Responsibility                                                                                      | Sign-off means                                                                   |
| --------------------------- | -------------- | --------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------- |
| **Functional requirements** | Captain        | Commission use-case coverage, AC completeness, scope — ensures all use cases are mapped to missions | "This mission delivers the right behavior and covers the commission's use cases" |
| **Design requirements**     | Design Officer | UX/UI, component architecture, design system                                                        | "This mission meets design standards"                                            |
| **Technical requirements**  | Commander      | Implementation approach, use-case-to-mission decomposition, sequencing, dependencies                | "This mission is buildable and correctly ordered"                                |

**Use case → Mission mapping responsibility**:

The **Commander** proposes how use cases decompose into missions (split, combine, or 1:1). The **Captain** validates that the mapping covers all commission-level acceptance criteria. Common patterns:

| Pattern            | When                                    | Example                                                                                             |
| ------------------ | --------------------------------------- | --------------------------------------------------------------------------------------------------- |
| **1:1**            | Use case maps cleanly to one mission    | UC "Login Form" → Mission "Implement Login Form"                                                    |
| **Split**          | Use case is too large for one mission   | UC "Auth System" → Mission "JWT Infrastructure" + Mission "Login UI" + Mission "Session Management" |
| **Combine**        | Multiple use cases share implementation | UC "Password Reset" + UC "Email Verification" → Mission "Email Service + Flows"                     |
| **Infrastructure** | No use case but technically required    | (none) → Mission "Database Migration for Auth Tables"                                               |

**How agents communicate in the Ready Room (session-isolated messaging)**:

Each agent operates in **their own harness session** (separate context window). They do not share context. The **`[Conditional Logic]`** planning loop orchestrator (deterministic) routes `ReadyRoomMessage` between sessions:

- **`[Conditional Logic]`** Messages arrive as **additional messages** in the target agent's session — they don't see the sender's full context, only the structured message
- **`[Conditional Logic]`** Each agent's **outputs are collected by the orchestrator**, not directly injected into other sessions
- **`[Conditional Logic]`** The orchestrator decides **what to pass and when** — it controls the information flow between sessions

Example messages routed between sessions:

- **Captain (Session A) → Commander (Session C)**: "Mission 3 needs auth middleware before Mission 5 can use it"
- **Commander (Session C) → Design Officer (Session B)**: "I'm splitting the dashboard into 2 missions — can you split the design specs?"
- **Design Officer (Session B) → Captain (Session A)**: "AC-2 mentions 'responsive layout' but the PRD doesn't specify breakpoints"
- **Commander (Session C) → All sessions**: "Proposed sequence: M1 → M2 → [M3, M4 parallel] → M5"

This mirrors harness-level agent teams: each teammate is a separate agent session with structured message delivery. Ship Commander 2 internalizes this so the system controls session lifecycle and message routing — the harness and model are interchangeable via settings.

**Legacy KB comparison**:

- KB's `kb-project-analyze` had specialists (product-manager, engineering-manager) collaborate to produce implementation plans
- KB's `kb-task-plan` enriched individual tasks with SPEC framework
- Ship Commander 2 needs to **internalize this** rather than depending on harness-level Teams/TasksList

**Key invariants**:

1. **Every planning loop targets a specific commission** — agents always know which PRD and which use cases they are planning against
2. No mission enters execution without **all three agent sign-offs AND Admiral (human) approval**
3. **All commission use cases must be covered** — the Captain validates that every use case has at least one mission implementing it before consensus is reached
4. Commander owns the final sequencing (by importance and dependency) and the **use-case-to-mission decomposition** — the Captain and Design Officer can object but the Commander resolves execution order
5. **Admiral feedback reconvenes the planning loop** — agents re-enter the Ready Room with feedback injected into their sessions
6. **Any agent can ask the Admiral questions mid-planning** — the loop suspends, surfaces the question via TUI, routes the answer back, and resumes (see Section 4.1c)
7. **Plans are persistable** — can be saved, shelved, resumed from saved state, and executed at a later time
8. **Each agent operates in their own harness session** — no shared context, only structured message passing between sessions
9. **The planning loop orchestrator (deterministic) controls message routing** — agents cannot directly access each other's sessions
10. Planning conversation is **persisted** — auditable, replayable, feeds into TUI dashboard
11. Standard task protocol is enforced — missions conform to system-level schema, not agent-invented formats
12. **Traceability is maintained** — every mission links back to its commission and the use case(s) it implements via `commissionId` and `useCaseRefs`
13. **Commission sizing rule of thumb** — a commission should amortize its Admiral approval cost over **5 or more missions**. If commissions routinely produce only 1-2 missions, the system is slipping back toward big-loop behavior (Admiral bottleneck on every small task). Conversely, if the planning loop is too eager to escalate questions or if Admiral feedback restarts loops too often, the approval cost inflates. The Ready Room should batch work at the commission level, not the task level

> **Progressive disclosure**: See `brain/skills/kb-project-analyze/SKILL.md` "TEAM ARCHITECTURE" for the existing 3-agent planning team pattern. See `brain/skills/kb-project-analyze/docs/task-description-format.md` for the v4.0 standard task format the Ready Room must produce. See `brain/skills/kb-project-analyze/docs/output-formats.md` for structured output with execution phases. See `brain/docs/AGENT-RULES.md` "Planning Mode" for agent planning discipline.

**Implementation effort**: ~3-4 days

---

### 4.1c HIGH: Agent → Admiral Questions During Planning (Priority: P1)

**Problem**: During the Ready Room planning loop, agents may encounter ambiguity, missing context, or decisions that only the human (Admiral) can resolve. Without a structured question/answer flow, agents either guess (introducing drift from intent) or block indefinitely. The planning loop needs an escape hatch where any agent can pause, surface a question to the Admiral, and resume with the answer.

**Solution**: Add a **question gate** to the Ready Room that any agent can trigger mid-planning. When an agent returns a question instead of (or alongside) analysis output, the planning loop **`[Conditional Logic]`** suspends, **`[Human Input → Admiral]`** surfaces the question to the Admiral via the TUI, waits for the answer, **`[Conditional Logic]`** routes it back to the agent's session, and resumes.

This follows the `AskUserQuestion` pattern (harness-agnostic) — the system presents a structured question with optional choices, waits for human input, and continues. The concrete harness (Claude Code, Codex, OpenCode) is interchangeable.

**TUI Design — Admiral Question Prompt**

The question appears as a modal overlay during the Planning Dashboard phase, using the existing LCARS theme. Three variants are shown below.

**Variant 1: Multiple-choice question (with pre-defined options)**

```
╔══════════════════════════════════════════════════════════════════════╗
║  ◆ ADMIRAL — QUESTION FROM CAPTAIN                                 ║
╠══════════════════════════════════════════════════════════════════════╣
║                                                                    ║
║  The PRD references "user authentication" but doesn't specify      ║
║  the auth strategy. Which approach should we implement?            ║
║                                                                    ║
║  Context: Mission M-003 (Login Flow) depends on this decision.     ║
║  This affects functional and technical requirements.               ║
║                                                                    ║
╠══════════════════════════════════════════════════════════════════════╣
║                                                                    ║
║    ► OAuth 2.0 + PKCE                                              ║
║      Industry standard, supports SSO, more complex setup           ║
║                                                                    ║
║      Email/Password + JWT                                          ║
║      Simpler, self-contained, faster to implement                  ║
║                                                                    ║
║      Magic Link (passwordless)                                     ║
║      Modern UX, requires email service integration                 ║
║                                                                    ║
╠══════════════════════════════════════════════════════════════════════╣
║  [↑/↓] Navigate  [Enter] Select  [t] Type custom answer  [s] Skip ║
╚══════════════════════════════════════════════════════════════════════╝
```

**Variant 2: Open-ended question (free-text input)**

```
╔══════════════════════════════════════════════════════════════════════╗
║  ◆ ADMIRAL — QUESTION FROM COMMANDER                               ║
╠══════════════════════════════════════════════════════════════════════╣
║                                                                    ║
║  The PRD lists 12 acceptance criteria but doesn't prioritize them. ║
║  Should all 12 be P0 for v1, or can some be deferred to v2?       ║
║                                                                    ║
║  Context: This affects mission sequencing and total scope.         ║
║  Domain: technical                                                 ║
║                                                                    ║
╠══════════════════════════════════════════════════════════════════════╣
║                                                                    ║
║  ADMIRAL > All 12 are needed for v1, but AC-7 through AC-12 can   ║
║           be parallel missions after the core auth flow ships._    ║
║                                                                    ║
╠══════════════════════════════════════════════════════════════════════╣
║  [Enter] Submit  [Esc] Cancel  [Tab] Toggle broadcast to all      ║
╚══════════════════════════════════════════════════════════════════════╝
```

**Variant 3: Inline question in Planning Dashboard (non-modal)**

When a question is pending, it appears as an alert row in the planning phase view:

```
╔══════════════════════════════════════════════════════════════════════╗
║  CAPTAIN'S READY ROOM — Planning Iteration 2/5                     ║
╠══════════════════════════════════════════════════════════════════════╣
║                                                                    ║
║  [✓] Captain        Functional analysis complete       12.3s       ║
║  [●] Design Officer Analyzing design requirements...    8.1s       ║
║  [⏸] Commander      Waiting for Admiral answer          ---        ║
║                                                                    ║
╠══ ⚠ ADMIRAL ATTENTION REQUIRED ═════════════════════════════════════╣
║                                                                    ║
║  Commander asks: "Should the API follow REST or GraphQL            ║
║  conventions? The PRD mentions both in different sections."         ║
║                                                                    ║
║  [1] REST (Recommended)  [2] GraphQL  [3] Both  [t] Type answer   ║
║                                                                    ║
╠══════════════════════════════════════════════════════════════════════╣
║  Missions: 5 drafted  │  Sign-offs: 3/5  │  Questions: 1 pending  ║
╚══════════════════════════════════════════════════════════════════════╝
```

**Key TUI design decisions**:

1. **LCARS-themed modal** — uses the existing `╔═╗║╚═╝` border system, LCARS_ORANGE header, LCARS_BLUE borders
2. **Identifies the asking agent** — "QUESTION FROM CAPTAIN" so the Admiral knows who needs the answer and in what domain
3. **Context is always shown** — the question includes which mission it relates to and what domain it affects
4. **Options follow the harness `AskUserQuestion` pattern** — structured choices with descriptions, plus a free-text escape hatch (`[t] Type custom answer`)
5. **Broadcast toggle** — Admiral can choose to send the answer to all agents (`[Tab]`) when the answer affects the whole plan
6. **Skip option** — Admiral can skip/defer a question (`[s]`) and let the agent use its best judgment
7. **Non-blocking variant** — the inline version (Variant 3) lets the Admiral see overall planning progress while a question is pending; other agents can continue their work if they don't depend on the answer
8. **Pending question counter** — the status bar shows `Questions: N pending` so the Admiral knows how many need attention

**Architectural changes**:

```typescript
// NEW FILE: src/planning/admiral-question-gate.ts

export interface AdmiralQuestionRequest {
  questionId: string;
  fromAgent: "captain" | "commander" | "design-officer";
  text: string;
  context?: string;
  options?: AdmiralQuestionOption[];
  domain: "functional" | "technical" | "design";
  missionId?: string;
  iteration: number;
}

export interface AdmiralQuestionResponse {
  questionId: string;
  text: string; // Admiral's answer (free text or selected option)
  selectedOption?: string; // Which option was selected (if applicable)
  broadcastToAll: boolean; // Should answer go to all agent sessions?
  skipped: boolean; // Admiral chose to skip — agent uses best judgment
  answeredAt: string;
}

// ALL METHODS IN THIS CLASS ARE [Conditional Logic] except where noted
export class AdmiralQuestionGate {
  private pendingQuestions: Map<string, AdmiralQuestionRequest> = new Map();
  private resolvers: Map<string, (response: AdmiralQuestionResponse) => void> =
    new Map();

  async ask(request: AdmiralQuestionRequest): Promise<AdmiralQuestionResponse> {
    this.pendingQuestions.set(request.questionId, request); // [Conditional Logic]

    // [Conditional Logic] Emit event for TUI to render the question prompt
    this.eventBus.emit("planning.admiral_question", {
      channel: "planning",
      type: "admiral_question",
      data: request,
    });

    // [Conditional Logic] SUSPEND — Promise awaits [Human Input → Admiral] via TUI
    return new Promise<AdmiralQuestionResponse>((resolve) => {
      this.resolvers.set(request.questionId, resolve);
    });
  }

  // [Human Input → Admiral] Called by TUI when Admiral submits an answer
  resolve(questionId: string, response: AdmiralQuestionResponse): void {
    const resolver = this.resolvers.get(questionId);
    if (resolver) {
      resolver(response);
      this.resolvers.delete(questionId);
      this.pendingQuestions.delete(questionId);
    }
  }

  getPendingQuestions(): AdmiralQuestionRequest[] {
    return Array.from(this.pendingQuestions.values());
  }

  hasPendingQuestions(): boolean {
    return this.pendingQuestions.size > 0;
  }
}
```

```typescript
// NEW FILE: src/tui/components/AdmiralQuestionModal.tsx
// React/Ink component that renders the LCARS-themed question modal

import { Box, Text, useInput } from "ink";

interface Props {
  question: AdmiralQuestionRequest;
  onAnswer: (response: AdmiralQuestionResponse) => void;
}

// Component renders:
// 1. LCARS header with "ADMIRAL — QUESTION FROM {agent}"
// 2. Question text + context
// 3. Options (if provided) with ▸ selector, or free-text input
// 4. Keyboard hints: [↑/↓] Navigate [Enter] Select [t] Type [s] Skip
// 5. Broadcast toggle: [Tab] to send answer to all agents
//
// Uses existing patterns:
// - LCARSPanel for borders
// - useInput() for keyboard handling (matches app.tsx pattern)
// - eventBus.emit('planning.admiral_answer', ...) to resolve
// - approvalResolverRef pattern for async resolution
```

**How questions flow through the system**:

```
Agent Session (e.g., Commander)
  │
  ├─ Agent returns { questions: [{ text, domain, options }] }
  │
  ▼
ReadyRoom.routeQuestionsToAdmiral()
  │
  ├─ Records question as ReadyRoomMessage (type: 'question')
  ├─ Calls AdmiralQuestionGate.ask(request)
  │
  ▼
AdmiralQuestionGate
  │
  ├─ Emits 'planning.admiral_question' event
  ├─ SUSPENDS (Promise awaits resolution)
  │
  ▼
TUI: AdmiralQuestionModal renders
  │
  ├─ Admiral reads question + context
  ├─ Admiral selects option or types answer
  ├─ Admiral optionally toggles "broadcast to all"
  │
  ▼
AdmiralQuestionGate.resolve(questionId, response)
  │
  ├─ Promise resolves → ReadyRoom continues
  │
  ▼
ReadyRoom routes answer
  │
  ├─ Records answer as ReadyRoomMessage (type: 'answer')
  ├─ Sends answer to asking agent's session
  ├─ If broadcast: sends to all other agent sessions
  │
  ▼
Planning loop resumes
```

**Integration with existing patterns**:

| Pattern               | Existing                                   | Admiral Questions                                                        |
| --------------------- | ------------------------------------------ | ------------------------------------------------------------------------ |
| **Async resolution**  | `approvalResolverRef` in app.tsx           | Same pattern — `AdmiralQuestionGate` holds Promise resolvers             |
| **Event bus**         | `planning.approval.*` events               | New `planning.admiral_question` / `planning.admiral_answer` events       |
| **TUI modal**         | Plan review overlay (PlanReviewPhase)      | New `AdmiralQuestionModal` component, same LCARS styling                 |
| **Keyboard handling** | `useInput()` in app.tsx                    | Same hook — arrows navigate options, Enter selects, `t` for free text    |
| **Message routing**   | `sendToSession()` for inter-agent messages | Same method — answer is routed as a message to agent session(s)          |
| **Status tracking**   | Phase indicators in PlanningDashboard      | `[⏸] Waiting for Admiral answer` status + `Questions: N pending` counter |

**Key invariants**:

1. Planning loop **suspends** on question — no forward progress until Admiral responds (or skips)
2. Questions are **persisted** in the planning message log — auditable, replayable
3. Admiral can **skip** — agent uses best judgment, but the skip is logged
4. Admiral can **broadcast** — one answer updates all agent sessions
5. **Non-blocking for other agents** — if Commander asks a question, Captain and Design Officer can continue their phases (only the asking agent's next step blocks)
6. Question/answer pairs are **linked by `questionId`** — traceable in the audit log

**Implementation effort**: ~2 days (1 day for gate + ReadyRoom integration, 1 day for TUI component)

---

### 4.2 CRITICAL: Implement Commander's Verification Gate Engine (Priority: P0)

**Problem**: Gate types exist but no Commander-owned execution engine runs them. The Commander needs a gate service to independently verify every agent claim before allowing state transitions.

**Solution**: Build `VerificationGateService` as the Commander's gate execution engine. **This entire service is `[Conditional Logic]`** — it runs shell commands, checks exit codes, and returns deterministic pass/fail verdicts. No AI is involved in gate execution:

```typescript
// NEW FILE: src/execution/verification-gate-service.ts

export class VerificationGateService {
  async executeGate(
    gateType: GateType,
    targetId: string, // missionId or acId
    context: GateContext,
  ): Promise<GateEvidence> {
    switch (gateType) {
      case "verify_red":
        return this.verifyRedGate(targetId, context);
      case "verify_green":
        return this.verifyGreenGate(targetId, context);
      case "verify_refactor":
        return this.verifyRefactorGate(targetId, context);
      case "verify_implement":
        return this.verifyImplementGate(targetId, context);
    }
  }

  // [Conditional Logic] — entire method is deterministic shell execution
  private async verifyRedGate(
    acId: string,
    context: GateContext,
  ): Promise<GateEvidence> {
    const ac = await this.beads.getAc(acId);
    const commands = await this.getMissionCommands(ac.missionId);

    // 1. [Conditional Logic] RUN TEST INDEPENDENTLY
    const testResult = await this.runCommand(commands.test, {
      fileFilter: ac.testFile,
    });

    // 2. CHECK FOR VANITY TEST
    if (testResult.exitCode === 0) {
      return {
        gateType: "verify_red",
        directiveId: ac.missionId,
        acId: ac.id,
        exitCode: testResult.exitCode,
        classification: "reject_vanity",
        snippet: testResult.stdout,
        timestamp: new Date().toISOString(),
        attempt: context.attempt,
      };
    }

    // 3. CHECK FOR TYPE ERRORS
    if (commands.typecheck) {
      const typeResult = await this.runCommand(commands.typecheck);
      if (typeResult.exitCode !== 0) {
        return {
          gateType: "verify_red",
          directiveId: ac.missionId,
          acId: ac.id,
          exitCode: typeResult.exitCode,
          classification: "reject_syntax",
          snippet: typeResult.stderr,
          timestamp: new Date().toISOString(),
          attempt: context.attempt,
        };
      }
    }

    // 4. CHECK FOR LINT VIOLATIONS
    if (commands.lint) {
      const lintResult = await this.runCommand(commands.lint);
      if (lintResult.exitCode !== 0) {
        return {
          gateType: "verify_red",
          directiveId: ac.missionId,
          acId: ac.id,
          exitCode: lintResult.exitCode,
          classification: "reject_syntax",
          snippet: lintResult.stdout,
          timestamp: new Date().toISOString(),
          attempt: context.attempt,
        };
      }
    }

    // ALL CHECKS PASSED
    return {
      gateType: "verify_red",
      directiveId: ac.missionId,
      acId: ac.id,
      exitCode: testResult.exitCode,
      classification: "accept",
      snippet: testResult.stdout,
      timestamp: new Date().toISOString(),
      attempt: context.attempt,
    };
  }

  // [Conditional Logic] — entire method is deterministic shell execution
  private async verifyGreenGate(
    acId: string,
    context: GateContext,
  ): Promise<GateEvidence> {
    const ac = await this.beads.getAc(acId);
    const commands = await this.getMissionCommands(ac.missionId);

    // 1. RUN TEST INDEPENDENTLY
    const testResult = await this.runCommand(commands.test, {
      fileFilter: ac.testFile,
    });

    // 2. TEST MUST NOW PASS
    if (testResult.exitCode !== 0) {
      return {
        gateType: "verify_green",
        directiveId: ac.missionId,
        acId: ac.id,
        exitCode: testResult.exitCode,
        classification: "reject_failure",
        snippet: testResult.stdout,
        timestamp: new Date().toISOString(),
        attempt: context.attempt,
      };
    }

    // 3. CHECK FOR FLAKINESS (E2E ONLY)
    if (context.track === "infra") {
      const results = [];
      for (let i = 0; i < 3; i++) {
        results.push(
          await this.runCommand(commands.test, { fileFilter: ac.testFile }),
        );
      }
      const passRate = results.filter((r) => r.exitCode === 0).length / 3;
      if (passRate < 1.0) {
        return {
          gateType: "verify_green",
          directiveId: ac.missionId,
          acId: ac.id,
          exitCode: testResult.exitCode,
          classification: "reject_failure",
          snippet: `Flaky test detected: ${passRate * 100}% pass rate (3 runs)`,
          timestamp: new Date().toISOString(),
          attempt: context.attempt,
        };
      }
    }

    return {
      gateType: "verify_green",
      directiveId: ac.missionId,
      acId: ac.id,
      exitCode: testResult.exitCode,
      classification: "accept",
      snippet: testResult.stdout,
      timestamp: new Date().toISOString(),
      attempt: context.attempt,
    };
  }
}
```

**Note**: The `VerificationGateService` handles per-AC TDD gates (red/green/refactor). For **mission-level demo token validation** — checking that the agent produced a valid `demo/MISSION-<id>.md` file with correct schema and evidence sections — see the `DemoTokenValidator` in **Section 4.8**.

**Implementation effort**: ~2-3 days

---

### 4.3 HIGH: Implement Commander's Protocol Event System (Priority: P1)

**Problem**: Agent outputs are transient return values, not persisted protocol events. The Commander needs a structured protocol to receive agent claims, dispatch verification results, and maintain an auditable coordination log.

**Solution**: Build `ProtocolEventService` as the Commander's coordination backbone. **This entire service is `[Conditional Logic]`** — it validates JSON schemas, persists events to Beads, emits to EventBus, and polls for claims. The only **`[Harness → Implementer Agent]`** involvement is the agent posting its claim event:

```typescript
// NEW FILE: src/protocol/protocol-event-service.ts

export interface ProtocolEvent {
  event_type: string; // 'AGENT_CLAIM' | 'GATE_RESULT' | 'STATE_TRANSITION'
  mission_id: string;
  timestamp: string;
  protocol_version: string; // "1.0"

  // Deterministic fields
  agent_id?: string;
  phase?: ExecutionPhase;

  // Probabilistic fields (agent-supplied)
  description?: string;
  evidence?: Record<string, unknown>;
}

export class ProtocolEventService {
  private readonly PROTOCOL_VERSION = "1.0";

  async publishEvent(
    event: Omit<ProtocolEvent, "timestamp" | "protocol_version">,
  ): Promise<void> {
    const protocolEvent: ProtocolEvent = {
      ...event,
      timestamp: new Date().toISOString(),
      protocol_version: this.PROTOCOL_VERSION,
    };

    // VALIDATE against JSON schema
    const validated = await this.validateProtocolEvent(protocolEvent);

    // PERSIST to Beads
    await this.beads.recordProtocolEvent(validated);

    // EMIT to EventBus for TUI
    this.eventBus.emit("protocol.event", validated);
  }

  async waitForClaim(
    missionId: string,
    expectedClaim: string,
    timeoutMs: number = 300_000,
  ): Promise<ProtocolEvent> {
    const startTime = Date.now();

    while (Date.now() - startTime < timeoutMs) {
      const events = await this.beads.getProtocolEvents(missionId, {
        event_type: "AGENT_CLAIM",
        since: startTime,
      });

      const claim = events.find((e) => e.description === expectedClaim);
      if (claim) {
        return claim;
      }

      await this.sleep(5_000); // Poll every 5 seconds
    }

    throw new Error(`Timeout waiting for agent claim: ${expectedClaim}`);
  }
}
```

**Agent protocol**:

```typescript
// [Harness → Implementer Agent] Agent posts claim (probabilistic output)
await protocolService.publishEvent({
  event_type: "AGENT_CLAIM",
  mission_id: mission.id,
  agent_id: "ensign-42",
  phase: "red",
  description: "RED_COMPLETE",
  evidence: {
    test_file: "src/auth/login.spec.ts",
    assertions_written: 3,
  },
});

// [Conditional Logic] Commander waits for claim (deterministic coordination)
const claim = await protocolService.waitForClaim(mission.id, "RED_COMPLETE");

// [Conditional Logic] Commander runs verification gate (deterministic enforcement)
const gateResult = await verificationGateService.executeGate(
  "verify_red",
  ac.id,
  context,
);

// [Conditional Logic] Commander records gate result (deterministic persistence)
await protocolService.publishEvent({
  event_type: "GATE_RESULT",
  mission_id: mission.id,
  gate_type: "verify_red",
  classification: gateResult.classification,
  exit_code: gateResult.exitCode,
});
```

**Implementation effort**: ~2 days

---

### 4.4 HIGH: Commander Enforces Mission Termination Rules (Priority: P1)

**Problem**: Missions can retry infinitely; no "single mission must terminate" invariant. The Commander must enforce termination — it is the authority that decides when a mission is halted, not the agent.

**Solution**: Add termination enforcement to the Commander's `MissionExecutionOrchestrator`. **This entire service is `[Conditional Logic]`** — deterministic boolean checks against session state (revision counts, demo token presence, AC completion). No AI involved:

```typescript
// EXISTING: src/domain/types.ts

export interface MissionExecutionSession {
  directiveId: string;
  missionId?: string;
  projectId: string;
  status: SessionStatus;
  revisionCount: number;
  maxRevisions: number;
  attemptCount: number; // ✅ ADD THIS FIELD
  startedAt: string;
  updatedAt: string;
  worktreePath?: string;
  mission: MissionClassification;
  trackState: MissionTrackState;
  activeAgents: AgentAssignment[];
  demoToken?: DemoToken;
  terminationReason?: string; // ✅ ADD THIS FIELD
}
```

```typescript
// NEW FILE: src/execution/termination-enforcer.ts

export class TerminationEnforcer {
  async checkTermination(
    session: MissionExecutionSession,
  ): Promise<TerminationCheck> {
    const { mission, revisionCount, maxRevisions, demoToken, trackState } =
      session;

    // CHECK 1: Max revisions exceeded
    if (revisionCount >= maxRevisions) {
      return {
        shouldTerminate: true,
        reason: "max_revisions_exceeded",
        message: `Mission halted after ${revisionCount} revisions (max: ${maxRevisions})`,
      };
    }

    // CHECK 2: Demo token required but not produced
    // V1: "produced" means demo/MISSION-<id>.md exists and passes
    // DemoTokenValidator schema checks (see Section 4.8)
    if (mission.demo_token_required) {
      const validation = await this.demoTokenValidator.validate(
        session.missionId,
        mission,
        session.worktreePath,
      );
      if (!validation.valid) {
        return {
          shouldTerminate: true,
          reason: validation.reason,  // e.g. "demo_token_file_missing", "red_alert_missing_tests"
          message: validation.message,
        };
      }
    }

    // CHECK 3: All ACs completed but demo token missing
    const allAcsComplete = trackState.acs.every((ac) => ac.completed);
    if (allAcsComplete && mission.demo_token_required) {
      const validation = await this.demoTokenValidator.validate(
        session.missionId,
        mission,
        session.worktreePath,
      );
      if (!validation.valid) {
        return {
          shouldTerminate: true,
          reason: "incomplete_without_proof",
          message: "All ACs complete but demo token invalid: " + validation.message,
        };
      }
    }

    return { shouldTerminate: false };
  }
}
```

**Note**: CHECK 2 and CHECK 3 now delegate to `DemoTokenValidator` (Section 4.8) instead of checking an in-memory `demoToken` field. The canonical proof is the `demo/MISSION-<id>.md` file in the worktree, not a transient object.

**Implementation effort**: ~1 day

---

### 4.5 MEDIUM: Commander Enforces Surface-Area Locking (Priority: P2)

**Problem**: No prevention of concurrent missions modifying same subsystem. The Commander sequences and dispatches missions — it must also enforce isolation by acquiring surface-area locks before dispatch.

**Solution**: Build `LockService` as the Commander's concurrency control mechanism. **This entire service is `[Conditional Logic]`** — deterministic lock acquisition, conflict detection via glob matching, and lock release. No AI involved:

```typescript
// NEW FILE: src/concurrency/lock-service.ts

export interface SurfaceAreaLock {
  id: string;
  missionId: string;
  surfaceArea: string[]; // ['auth/*', 'database/users/*']
  acquiredAt: string;
  expiresAt: string;
}

export class LockService {
  async acquireLock(
    missionId: string,
    surfaceArea: string[],
    timeoutMs: number = 300_000,
  ): Promise<SurfaceAreaLock> {
    // CHECK for conflicts
    const existingLocks = await this.beads.getActiveLocks();
    const conflict = this.findConflict(existingLocks, surfaceArea);

    if (conflict) {
      throw new Error(
        `Cannot acquire lock for ${surfaceArea.join(", ")}: ` +
          `conflicts with mission ${conflict.missionId} holding ${conflict.surfaceArea.join(", ")}`,
      );
    }

    // CREATE lock
    const lock: SurfaceAreaLock = {
      id: `lock-${Date.now()}`,
      missionId,
      surfaceArea,
      acquiredAt: new Date().toISOString(),
      expiresAt: new Date(Date.now() + timeoutMs).toISOString(),
    };

    await this.beads.createLock(lock);
    return lock;
  }

  async releaseLock(missionId: string): Promise<void> {
    await this.beads.releaseLocksForMission(missionId);
  }
}
```

**Usage in mission dispatch**:

```typescript
// Mission spec declares surface area
const mission: MissionSpec = {
  id: "MISSION-42",
  title: "Add login button",
  surfaceArea: ["src/auth/login.tsx", "src/auth/login.test.ts"],
  // ...
};

// Commander acquires lock before dispatch
await lockService.acquireLock(mission.id, mission.surfaceArea);

try {
  await commander.executeMission(mission.id);
} finally {
  await lockService.releaseLock(mission.id);
}
```

**Implementation effort**: ~2 days

---

### 4.6 LOW: Add Attempt Counter to Mission Session (Priority: P3)

**Problem**: No tracking of how many times a mission has been retried.

**Solution**: Already partially addressed in section 4.4, but ensure it's persisted:

```typescript
// UPDATE: MissionExecutionSession interface
attemptCount: number; // Increment on each revision loop
```

```typescript
// In MissionExecutionOrchestrator.executeMission()
session.attemptCount += 1;
await this.beads.updateSession(session.id, {
  attemptCount: session.attemptCount,
});
```

**Implementation effort**: ~0.5 days

---

### 4.7 RECOMMENDED: Commander Routes Standard Ops Fast Path (Priority: P3)

**Problem**: Low-risk missions (Standard Ops) still pay TDD overhead when tests add little value. The Commander should route missions through different execution paths based on the Captain's risk classification.

**Solution**: Commander implements fast path routing for STANDARD_OPS missions:

```typescript
// [Conditional Logic] In Commander's MissionExecutionOrchestrator.executeMission()
if (mission.classification.class === "STANDARD_OPS") {
  // FAST PATH: Implement → Verify → Demo Token
  await this.dispatchAgent(session, mission, "implement");        // [Harness → Implementer Agent]
  await this.waitForAgentClaim(mission.id, "IMPLEMENT_COMPLETE"); // [Conditional Logic] poll

  // [Conditional Logic] Single verification gate
  const result = await this.runVerifyImplementGate(mission.id);   // [Conditional Logic] shell cmd

  // [Conditional Logic] Validate demo token file (Section 4.8)
  const validation = await this.demoTokenValidator.validate(
    mission.id, mission.classification, session.worktreePath);
  if (!validation.valid) {
    return { status: "halted", reason: validation.reason };
  }

  return { status: "completed", session };
}

// RED ALERT: Full TDD cycle
if (mission.classification.class === "RED_ALERT") {
  // Per-AC RED → GREEN → REFACTOR loop
  // Same [Harness → Implementer Agent] → [Conditional Logic] pattern per-AC
}
```

**Implementation effort**: ~1 day

---

### 4.8 V1 Demo Token Specification: Strict Markdown Schema (Priority: P1)

**Problem**: The `DemoToken` type exists (Section 1.4) but is an in-memory object with no enforced format, no file artifact, and optional status. Agents can claim completion without producing verifiable proof. The manifesto requires **observable system effects** — V1 must formalize what a demo token *is*, where it lives, and how the Commander validates it.

**Design constraint**: A demo token must be **verifiable by a human without reading all the code**. It is the artifact the Admiral reviews to confirm a mission actually did what it claimed.

**V1 philosophy**: Intentionally boring. A single Markdown file with a strict schema plus a small set of allowed evidence types that you can verify in minutes. Defer rich media (screenshots, E2E reports, trace files, videos) to V2+.

#### 4.8.1 File Convention

```
demo/MISSION-<id>.md
```

One file per mission. The file is committed to the mission's worktree branch. The Commander validates its existence and schema before allowing mission completion.

#### 4.8.2 Schema: YAML Frontmatter + Markdown Body

```markdown
---
mission_id: "MISSION-42"
title: "Add login button to auth page"
classification: "RED_ALERT"        # or STANDARD_OPS
status: "complete"                 # complete | partial
created_at: "2026-02-10T14:30:00Z"
agent_id: "implementer-1"
---

## Evidence

### commands

- `pnpm test src/auth/login.test.ts`
  - exit_code: 0
  - summary: "3 tests passed (login renders, submits, validates)"

### tests

- file: `src/auth/login.test.ts`
  - added_tests:
    - "should render login button"
    - "should submit credentials on click"
    - "should show validation error for empty fields"
  - passing: true

### manual_steps

1. Run `pnpm dev`
2. Navigate to `/auth`
3. Observe login button in top-right corner
4. Click button → login form appears

### diff_refs

- `src/auth/login.tsx` — added LoginButton component (lines 12-45)
- `src/auth/login.test.ts` — added 3 test cases (lines 8-52)
- `src/auth/index.ts` — re-exported LoginButton
```

#### 4.8.3 Allowed Evidence Types (V1)

| Type | Required Fields | Purpose |
|------|----------------|---------|
| **`commands`** | command string, `exit_code`, `summary` | Prove a shell command ran successfully. Commander can re-run to verify. |
| **`tests`** | `file`, `added_tests[]`, `passing` | Prove tests exist and pass. Commander cross-references with gate results. |
| **`manual_steps`** | Numbered step list | Give the Admiral a human-walkable verification path. |
| **`diff_refs`** | File path + description | Link proof to actual code changes. Commander validates files exist in worktree. |

**V1 scope — explicitly deferred**:
- Screenshots (requires headless browser infrastructure)
- E2E test reports (requires test runner integration beyond shell exit codes)
- Trace files (requires OpenTelemetry/profiler setup)
- Video recordings (requires screen capture tooling)

These belong in V2+ when the infrastructure exists to capture and validate them automatically.

#### 4.8.4 Mode-Dependent Requirements

**`[Conditional Logic]`** — the Commander enforces these rules based on the mission's classification:

| Classification | Required Sections | Rationale |
|---------------|-------------------|-----------|
| **`RED_ALERT`** | `tests` + at least one of (`commands`, `diff_refs`) | TDD missions must prove tests exist and pass |
| **`STANDARD_OPS`** | At least one of (`commands`, `manual_steps`, `diff_refs`) | Non-TDD missions need some verifiable proof |

If the classification requires `tests` and the `tests` section is missing or has `passing: false`, the Commander rejects the demo token.

#### 4.8.5 Enforcement Rules

**`[Conditional Logic]`** — all enforcement is deterministic. The Commander runs these checks before allowing mission completion:

```typescript
// NEW FILE: src/execution/demo-token-validator.ts

export class DemoTokenValidator {
  // [Conditional Logic] — entire class is deterministic file/schema checks

  async validate(
    missionId: string,
    classification: MissionClassification,
    worktreePath: string,
  ): Promise<DemoTokenValidation> {
    const filePath = path.join(worktreePath, "demo", `MISSION-${missionId}.md`);

    // RULE 1: File must exist
    if (!await fileExists(filePath)) {
      return { valid: false, reason: "demo_token_file_missing",
        message: `Expected ${filePath} but file does not exist` };
    }

    // RULE 2: YAML frontmatter must parse
    const { frontmatter, body } = parseFrontmatter(await readFile(filePath));
    if (!frontmatter) {
      return { valid: false, reason: "yaml_parse_error",
        message: "Demo token file has invalid or missing YAML frontmatter" };
    }

    // RULE 3: Required frontmatter fields present
    const required = ["mission_id", "title", "classification", "status", "created_at", "agent_id"];
    for (const field of required) {
      if (!frontmatter[field]) {
        return { valid: false, reason: "missing_required_field",
          message: `Demo token missing required field: ${field}` };
      }
    }

    // RULE 4: mission_id matches
    if (frontmatter.mission_id !== missionId) {
      return { valid: false, reason: "mission_id_mismatch",
        message: `Demo token mission_id "${frontmatter.mission_id}" does not match "${missionId}"` };
    }

    // RULE 5: diff_refs reference real files in worktree
    const diffRefs = extractDiffRefs(body);
    for (const ref of diffRefs) {
      if (!await fileExists(path.join(worktreePath, ref.filePath))) {
        return { valid: false, reason: "diff_ref_file_missing",
          message: `diff_ref references "${ref.filePath}" but file does not exist in worktree` };
      }
    }

    // RULE 6: Mode-dependent section requirements
    const sections = extractEvidenceSections(body);
    if (classification.class === "RED_ALERT") {
      if (!sections.has("tests")) {
        return { valid: false, reason: "red_alert_missing_tests",
          message: "RED_ALERT missions require a 'tests' evidence section" };
      }
      if (!sections.has("commands") && !sections.has("diff_refs")) {
        return { valid: false, reason: "red_alert_missing_corroboration",
          message: "RED_ALERT missions require 'commands' or 'diff_refs' in addition to 'tests'" };
      }
    } else {
      // STANDARD_OPS: at least one evidence section
      if (!sections.has("commands") && !sections.has("manual_steps") && !sections.has("diff_refs")) {
        return { valid: false, reason: "standard_ops_no_evidence",
          message: "STANDARD_OPS missions require at least one evidence section" };
      }
    }

    return { valid: true };
  }
}
```

#### 4.8.6 Integration with Commander Lifecycle

The `DemoTokenValidator` is called by the Commander at two points:

1. **Before mission completion** — when all ACs are marked complete, the Commander validates the demo token file before transitioning to `done`. This is in the `MissionExecutionOrchestrator` (Section 4.1).
2. **In the TerminationEnforcer** (Section 4.4) — when checking `demo_token_required`, the enforcer delegates to `DemoTokenValidator` instead of just checking for the presence of an in-memory `DemoToken` object.

```typescript
// In MissionExecutionOrchestrator.completeMission()
// [Conditional Logic] — validate demo token file before completing

const validation = await this.demoTokenValidator.validate(
  missionId,
  session.mission,
  session.worktreePath,
);

if (!validation.valid) {
  this.eventBus.emit("gate", {
    type: "demo_token_rejected",
    missionId,
    reason: validation.reason,
    message: validation.message,
  });
  return { status: "halted", reason: validation.reason };
}

// Demo token valid — proceed with completion
await this.stateService.transitionMission(missionId, "review", "done", "demo_token_validated");
```

**Implementation effort**: ~2 days (validator + integration + tests)

---

## 5. What NOT To Change

### 5.1 Preserve: State Machine Foundation

**DO NOT TOUCH**:

- `MissionStatus` enum and transitions
- `AcTddState` enum and transitions
- `assertLegalMissionTransition()` function
- `StateService` class

**Why**: This is the load-bearing structure. It's correct, tested, and manifesto-aligned.

---

### 5.2 Preserve: Mission Classification System

**DO NOT TOUCH**:

- `MissionClassification` interface
- `RED_ALERT` vs `STANDARD_OPS` distinction
- Risk scoring logic
- `loop_mode` field

**Why**: It directly implements manifesto's risk-aware approach.

---

### 5.3 Preserve: Event System

**DO NOT TOUCH**:

- `RuntimeEventPayload` type
- Event channel structure
- EventBus integration

**Why**: It provides the observability foundation. Expand it, don't restructure it.

---

### 5.4 Preserve: Per-AC State Tracking

**DO NOT TOUCH**:

- `AcExecutionState` interface
- `ExecutionAttempt` array
- Phase tracking per AC

**Why**: This is the kernel of "one AC = one loop" enforcement.

---

## 6. Implementation Roadmap

### Phase 1: Build the Commander, Ready Room & Admiral Gate (Week 1-2)

1. **Commission type + Captain's Ready Room (Planning Loop)** (3-4 days) — formalize the `Commission` type (from PRDContext). Collaborative planning within a commission's context where Captain (functional), Design Officer (design), and Commander (technical) decompose use cases into missions, converse in isolated sessions, validate use-case coverage, take domain ownership, and produce a signed-off mission manifest. Includes **Admiral (human) approval gate** — no dispatch without Admiral sign-off. Admiral feedback reconvenes the loop. **Plans are persistable** — save, shelve, resume, execute later.
   - **`[Harness → Captain/Commander/Design Officer Agents]`**: `spawnSession()` for each agent, `session.run()` for analysis/drafting
   - **`[Conditional Logic]`**: `sendToSession()` message routing, `hasConsensus()` checks, `savePlan()` persistence, feedback injection, planning loop iteration
   - **`[Human Input → Admiral]`**: `requestAdmiralApproval()` gate, Admiral feedback
2. **Agent → Admiral Question Gate** (2 days) — any agent can ask the Admiral questions mid-planning via TUI prompt. Includes `AdmiralQuestionGate` service and `AdmiralQuestionModal` TUI component.
   - **`[Conditional Logic]`**: `AdmiralQuestionGate.ask()` (Promise suspension, event emission, answer routing)
   - **`[Human Input → Admiral]`**: TUI modal presentation, Admiral answers/skips
3. **Commander (Mission Execution Orchestrator)** (3-5 days) — the distinct orchestrator that owns mission lifecycle, separate from the Captain
   - **`[Harness → Implementer Agent]`**: `dispatchAgent()` to spawn coding agent sessions
   - **`[Conditional Logic]`**: `waitForAgentClaim()` polling, `runVerifyRedGate()`/`runVerifyGreenGate()` shell execution, `transitionAcPhase()` state machine, termination enforcement
4. **Commander's Verification Gate Engine** (2-3 days) — independent gate execution that the Commander runs against agent claims
   - **Entirely `[Conditional Logic]`**: Shell command execution, exit code checking, vanity test detection, flakiness detection. No AI involved.

**Deliverable**: Planning happens in the Ready Room with domain ownership, agent sign-offs, and Admiral approval. Plans can be saved and resumed. Commander exists as a distinct execution authority. Missions execute with Commander-owned verification gates.

---

### Phase 2: Commander Protocol, Termination & Demo Tokens (Week 3)

1. **Commander's Protocol Event Service** (2 days) — structured communication backbone for agent ↔ Commander coordination
   - **`[Conditional Logic]`**: JSON schema validation, event persistence to Beads, EventBus emission, claim polling
   - **`[Harness → Implementer Agent]`**: Only the agent's `publishEvent()` call posting its claim
2. **Commander's Termination Enforcer** (1 day) — deterministic rules for when the Commander halts a mission
   - **Entirely `[Conditional Logic]`**: Boolean checks on revision count, demo token validation, AC completion status
3. **V1 Demo Token Validator** (2 days) — `DemoTokenValidator` enforces strict Markdown schema for `demo/MISSION-<id>.md` files (Section 4.8)
   - **Entirely `[Conditional Logic]`**: File existence, YAML frontmatter parsing, required field checks, diff_ref file validation, mode-dependent section requirements
   - Integrates with TerminationEnforcer (CHECK 2/3) and MissionExecutionOrchestrator (completion gate)
4. **Integrate protocol, termination, and demo tokens with Commander** (2 days)
   - **`[Conditional Logic]`**: Wiring protocol events, termination checks, and demo token validation into the Commander's execution loop

**Deliverable**: Commander coordinates via persisted protocol events. Missions terminate when Commander's rules are met. Demo tokens are validated Markdown files with strict schema — no mission completes without verifiable proof. Captain receives status reports from Commander.

---

### Phase 3: Commander Advanced Features (Week 4)

1. **Commander's Surface-Area Locking** (2 days) — Commander enforces isolation at dispatch time
   - **Entirely `[Conditional Logic]`**: Lock acquisition, conflict detection via glob matching, lock release
2. **Commander's Standard Ops Fast Path** (1 day) — Commander routes low-risk missions through simplified execution
   - **`[Conditional Logic]`**: Classification check determines routing
   - **`[Harness → Implementer Agent]`**: `dispatchAgent()` for single-pass implementation
3. **Add Attempt Counter** (0.5 days)
   - **Entirely `[Conditional Logic]`**: Increment and persist counter on each revision
4. **Integration Testing** (1.5 days) — validate Captain → Commander → Agent → Commander → Captain flow
   - Tests exercise the full **`[Harness]`** → **`[Conditional Logic]`** → **`[Human Input]`** cycle

**Deliverable**: Full manifesto compliance with Commander-managed concurrent mission support.

---

### Phase 4: Hardening (Week 5)

1. **Add missing event types** for Commander protocol — **`[Conditional Logic]`**
2. **Post-hoc analysis tools** — replay Commander decisions from protocol log — **`[Conditional Logic]`**
3. **Recovery logic refinement** — Commander handles agent crashes, timeouts, partial state — **`[Conditional Logic]`**
4. **Documentation** — Captain vs Commander responsibilities, mission lifecycle, protocol spec

**Deliverable**: Production-ready manifesto-aligned system with clear Captain/Commander separation.

---

## 7. Success Criteria

**System is manifesto-aligned when**:

1. ✅ **Admiral, Captain, Commander, and Design Officer are distinct**: Admiral commissions and approves; Captain owns commission-level adherence; Commander owns mission execution; Design Officer owns design requirements — roles are not collapsed
2. ✅ **Commission→Mission traceability**: Every mission links to a commission and the use case(s) it implements; every commission use case is covered by at least one mission
3. ✅ **Ready Room planning loop**: All missions are collaboratively decomposed from commission use cases with domain ownership, three-way agent sign-off, use-case coverage validation, and **Admiral (human) approval** before execution begins
4. ✅ **Admiral feedback reconvenes the loop**: Any Admiral feedback restarts the planning loop with feedback injected into agent sessions
5. ✅ **Agent → Admiral questions**: Any agent can ask the Admiral questions mid-planning; the loop suspends, surfaces the question via TUI, routes the answer back, and resumes
6. ✅ **Plans are persistable**: Mission manifests can be saved, shelved, resumed, and executed at a later time
7. ✅ **No agent self-certifies**: Every agent claim is independently verified by the Commander
8. ✅ **Every mission terminates**: Commander enforces `maxRevisions` or demo token production
9. ✅ **Every mission yields proof**: Demo tokens are required, not optional — V1 enforces `demo/MISSION-<id>.md` with strict schema (Section 4.8)
10. ✅ **Isolation by default**: Commander acquires surface-area locks before dispatch
11. ✅ **Review is a gate**: Probabilistic review happens inside Commander's deterministic cage
12. ✅ **Failure is cheap**: Commander can discard and restart missions without carryover state

**Measured by**:

- Zero "agent claimed X but it was false" incidents (Commander catches all)
- 100% of commission use cases covered by at least one mission (no orphaned use cases)
- 100% of missions link back to their commission and use case(s) (full traceability)
- 100% of missions have Commander-logged termination reason
- 100% of completed missions have valid `demo/MISSION-<id>.md` file passing `DemoTokenValidator` checks
- Zero concurrent modification conflicts (Commander enforces locks)
- All review verdicts persisted as structured Commander protocol events
- Clear audit trail of Ready Room → Commander → Agent → Commander → Captain flow
- 100% of missions have three-way agent sign-off + Admiral approval before execution
- Zero missions dispatched without Admiral greenlight
- Ready Room conversation log is persisted and auditable
- All agent → Admiral questions and answers are logged with linked `questionId`
- Plans can be saved, resumed, and re-executed from saved state

---

## 8. Key Insights

### Insight 1: The Gap is Enforcement, Not Architecture

Ship Commander 2 has the **right data structures** and **right vocabulary**. The gap is **behavioral enforcement**—the code that says "STOP" when an invariant is violated.

**Implication**: This is **not a rewrite**. It's **adding enforcement logic** around existing structures.

---

### Insight 2: "Don't Trust the Agent" Requires the Commander

KB's comment polling loop was ugly but it **enforced the invariant**:

- Agent posts claim
- Commander (kb-swarm) sees claim
- Commander runs verification independently
- Commander posts result
- Agent sees result and proceeds (or retries)

Ship Commander 2 needs the **Commander as a distinct orchestrator** — not a specialist agent, but the execution authority that sits between every agent and every state transition.

**Implication**: Build the Commander (`MissionExecutionOrchestrator`) as the authority that owns mission execution. The Captain validates requirements; the Commander enforces execution discipline.

---

### Insight 3: Demo Tokens Are the Manifesto's Killer Feature

KB required **tests only**. Manifesto generalizes to **any proof**:

- Screenshot for UI changes
- Command output for CLI tools
- Trace for performance changes
- Diff for refactors

Ship Commander 2 has the `DemoToken` type but doesn't **require** it.

**Implication**: Make `demo_token_required` true by default, false only for explicitly exempted work.

**V1 answer** (Section 4.8): Intentionally boring — a single Markdown file (`demo/MISSION-<id>.md`) with YAML frontmatter and 4 evidence types (`commands`, `tests`, `manual_steps`, `diff_refs`). Verifiable by a human without reading all code. Screenshots, E2E reports, traces, and videos deferred to V2+ when infrastructure exists.

---

### Insight 4: Protocol Versioning Enables Evolution

KB's comment protocol was **brittle**—regex parsing breaks easily.

Ship Commander 2 should use **JSON schema validation** for protocol events:

- `protocol_version: "1.0"`
- Schema validation on publish
- Backward compatibility on read

**Implication**: Protocol changes are non-breaking if versioned properly.

---

## 9. Conclusion

**Ship Commander 2 is 80% of the way to manifesto compliance.**

The **hard problems are solved**:

- ✅ State machine
- ✅ Risk classification
- ✅ Per-AC TDD tracking
- ✅ Event system
- ✅ Demo token type

The **remaining work is building the Commission type, the Commander, the Ready Room, and the Admiral gate**:

- ❌ **Commission concept** — formalize the PRD-level initiative as a first-class `Commission` type with use cases, acceptance criteria, and use-case-to-mission traceability
- ❌ **Captain's Ready Room** — collaborative planning loop within a commission's context where Captain (functional), Commander (technical), and Design Officer (design) decompose use cases into missions, converse, take domain ownership, validate use-case coverage, and sign off on mission specs — with **Admiral (human) approval** before any dispatch, **agent → Admiral questions** for mid-planning clarification, and **plan persistence** for save/resume/shelve
- ❌ **Commander as mission execution orchestrator** — distinct from Captain, owns execution lifecycle
- ❌ **Commander's verification gate engine** — independent gate execution
- ❌ **Commander's protocol event system** — structured agent ↔ Commander coordination
- ❌ **Commander's termination enforcer** — deterministic halt rules
- ❌ **V1 Demo Token validator** — enforce strict Markdown schema for `demo/MISSION-<id>.md` with YAML frontmatter, 4 evidence types, mode-dependent requirements (Section 4.8)
- ❌ **Commander's surface-area locking** — concurrency control at dispatch

**The core architectural gaps are**: (1) the **absence of a commission concept** connecting PRD-level initiatives to their implementing missions, (2) the **collapsed Captain/Commander/Design Officer boundary** — these roles run in parallel but don't collaborate or take domain ownership, (3) the **absence of a planning loop** where specialists decompose commissions into missions and reach consensus before execution, and (4) the **absence of an Admiral (human) approval gate** with plan persistence so manifests can be reviewed, revised, shelved, and executed on the human's timeline. Formalizing commissions, building the Ready Room with Admiral gate, and the Commander as a distinct orchestrator, are the most impactful changes.

**This is a 5-week journey from "light manifesto" to "full manifesto compliance."**

The codebase is **well-structured, type-safe, and ready** for these additions. None of the recommendations require rewrites—all are **additive enforcement layers** around existing structures, with the Commander as the central new authority.

**Final assessment**: Ship Commander 2 successfully **escaped KB's complexity trap** while preserving its best ideas. With the **Commission concept** formalizing PRD-to-mission traceability, the **Captain's Ready Room** for commission-driven collaborative planning, the **Admiral gate** for human approval and plan persistence, and the **Commander** built as a distinct execution orchestrator, it will achieve the manifesto's vision of **disciplined, observable, single-loop development** — where the Admiral _commissions and approves_, the Captain owns _what_ (commission adherence), the Design Officer owns _look and feel_, and the Commander owns _how_ and _when_ (mission execution).

---

**"Starfleet advances through discipline, not assumption. The same must be true of our software."**

✓ **The foundation is solid.**
✓ **The path is clear.**
✓ **The discipline is enforceable.**

**Engage.**

---

## Appendix A: Required Reading (Progressive Disclosure)

The following `brain/` documents define the patterns Ship Commander 2 must internalize. Read in order of relevance to your current section.

### Core Architecture References

| Document                  | Path                                  | Relevance                                                                                                                                                                                                             |
| ------------------------- | ------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **TDD Methodology**       | `brain/docs/TDD-METHODOLOGY.md`       | Defines the RED → GREEN → REFACTOR cycle, per-AC iteration, and the principle that "agents cannot verify themselves — orchestrators must enforce verification gates." Core reference for Sections 2.2, 2.3, 4.1, 4.2. |
| **Verification Gates**    | `brain/docs/VERIFICATION-GATES.md`    | Mandatory verification commands by domain. Every agent MUST run these before marking a task complete. Core reference for Section 2.2, 4.2.                                                                            |
| **Agent Rules**           | `brain/docs/AGENT-RULES.md`           | Agent operating modes (Planning vs Execution), task creation, task execution, handoff documentation. Defines the agent discipline that Captain/Commander must enforce.                                                |
| **Universal Task Schema** | `brain/docs/UNIVERSAL-TASK-SCHEMA.md` | Portable task management schema — cross-session persistence, multi-agent coordination, dependency tracking. Defines the "standard task protocol" missions must conform to.                                            |
| **Shared Formats**        | `brain/docs/SHARED-FORMATS.md`        | Standard message and promise formats for agent ↔ orchestrator communication. Defines `USER_INTERVENTION_REQUIRED` and structured coordination patterns.                                                               |
| **Agent Workflows**       | `brain/docs/AGENT-WORKFLOWS.md`       | Multi-turn, sequential, and parallel skill invocation patterns. Defines how agents should collaborate across domains.                                                                                                 |

### Legacy KB Skill References (Commander Patterns)

| Skill                                     | Path                                       | Relevance                                                                                                                                                                                                                                                                                                                                                                    |
| ----------------------------------------- | ------------------------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **kb-swarm** (≈ Commander)                | `brain/skills/kb-swarm/SKILL.md`           | The orchestrator that runs verification gates directly, delegates code to kb-dispatch, manages worktrees. **This is the closest existing pattern to the Commander role.** See `brain/skills/kb-swarm/docs/monitoring-loop.md` for the wave execution and AC progress tracking algorithm.                                                                                     |
| **kb-dispatch** (≈ Agent under Commander) | `brain/skills/kb-dispatch/SKILL.md`        | TDD workflow executor for a single issue. Per-AC iteration with orchestrator verification gates. Implements code only — delegates verification to orchestrator. **The implementer role that the Commander dispatches and monitors.**                                                                                                                                         |
| **kb-project-analyze** (≈ Ready Room)     | `brain/skills/kb-project-analyze/SKILL.md` | Spawns specialist team (product-manager + engineering-manager + ui-designer) to collaboratively build implementation plans. **The closest existing pattern to the Captain's Ready Room.** See `brain/skills/kb-project-analyze/docs/task-description-format.md` for standard task format and `brain/skills/kb-project-analyze/docs/output-formats.md` for structured output. |

### Task Protocol References

| Document                    | Path                                                              | Relevance                                                                                                                                          |
| --------------------------- | ----------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Task Description Format** | `brain/skills/kb-project-analyze/docs/task-description-format.md` | v4.0 TDD-integrated task template with TDD_PHASE, CURRENT_AC, TEST_COMMAND fields. Defines the standard task protocol the Ready Room must produce. |
| **Monitoring Loop**         | `brain/skills/kb-swarm/docs/monitoring-loop.md`                   | Wave execution algorithm, AC progress tracking, polling, verification gate execution. The Commander's execution loop pattern.                      |
| **Coding Agent**            | `brain/docs/CODING-AGENT.md`                                      | Session protocol for incremental feature implementation. Understand → Select → Implement → Test → Commit → Checkpoint.                             |

---

## Appendix B: Terminology — Commission vs Mission

Throughout this document, two distinct levels of work are referenced. Keeping these separate is critical to understanding the planning and execution architecture.

| Term           | Definition                                                                                                                                                                                                              | Analogy           | Example                                                                                                            |
| -------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------- | ------------------------------------------------------------------------------------------------------------------ |
| **Commission** | A **PRD-level initiative** — a product requirement with specific use cases and acceptance criteria. A commission represents the _what_ and _why_ at the product level. One commission may produce one or many missions. | Epic / Initiative | "Build user authentication system" (PRD with 4 use cases: login, registration, password reset, session management) |
| **Mission**    | An **individual coding task** that executes specific implementation work. A mission is the atomic unit of execution — it goes through the TDD cycle, gets dispatched to an agent, and produces verifiable output.       | Task / Issue      | "Implement login form with OAuth 2.0 flow" (single task with 3 acceptance criteria)                                |

**Key relationships**:

- A **commission** is parsed from a PRD and contains **use cases** (`UseCase[]` in the current codebase)
- Use cases **often but not always** become missions — the Ready Room planning process decomposes, combines, splits, or resequences use cases into missions based on technical dependencies, design constraints, and implementation feasibility
- The **Captain's Ready Room** always plans within the context of a specific commission — agents know which PRD they're working against
- A **mission manifest** belongs to exactly one commission — it is the execution plan for that commission
- The **Commander** sequences and dispatches **missions**; the **Captain** validates that completed missions satisfy the **commission's** use cases and acceptance criteria

```
Commission (PRD)
  ├── Use Case 1  ──→  Mission A
  ├── Use Case 2  ──→  Mission B + Mission C  (split for technical reasons)
  ├── Use Case 3  ──→  Mission D
  └── Use Case 4  ──→  (merged into Mission B)  (combined for efficiency)
```

> **Existing codebase mapping**: `PRDContext` (src/intake/types.ts) is the closest existing type to a commission. It contains `useCases: UseCase[]`, `productName`, scope, and functional groups. The document recommends formalizing this as a `Commission` type that persists beyond the intake pipeline and provides the planning context for the Ready Room.
