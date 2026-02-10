# KB Review Ecosystem Analysis

**Date**: 2026-02-10
**Scope**: Full analysis of `project-review-gates`, `code-review`, `kb-dispatch`, `kb-swarm`, `kb-submit`, `kb-approve`, `kb-reject`, `spec-review-swarm`, `test-review`, and all supporting skills/agents.

---

## 1. System Overview

The KB review ecosystem is a multi-phase software delivery pipeline that orchestrates AI agents through a TDD-enforced workflow. It spans from issue intake through implementation, verification, review, and approval - with Linear (project management) as the coordination backbone and structured comments as the inter-agent protocol.

The system has **three fundamental layers**:

| Layer | Nature | Examples |
|-------|--------|---------|
| **Orchestration** | Deterministic state machine | kb-swarm monitoring loop, Linear state transitions, comment polling |
| **Verification** | Deterministic gate execution | typecheck, lint, test runners, AC count validation |
| **Judgment** | Probabilistic AI reasoning | code-review verdicts, test quality grading, design review, AC interpretation |

---

## 2. Process Map (Mermaid)

### 2.1 End-to-End Pipeline

```mermaid
flowchart TB
    subgraph INTAKE["Phase 0: Intake & Planning"]
        PRD["PRD Document"]
        FUNNEL["kb-project-funnel<br/><i>Extract projects from PRD</i>"]
        ANALYZE["kb-project-analyze<br/><i>Spawn specialist team</i>"]
        WSJF["kb-project-wsjf<br/><i>Calculate priority score</i>"]
        IMPL_PLAN["kb-project-impl-plan<br/><i>Build task breakdown</i>"]
        TASK_CREATE["kb-task-creator<br/><i>Create Linear issues</i>"]

        PRD --> FUNNEL
        FUNNEL --> ANALYZE
        ANALYZE --> WSJF
        ANALYZE --> IMPL_PLAN
        IMPL_PLAN --> TASK_CREATE
    end

    subgraph DISPATCH["Phase 1: Dispatch"]
        KB_SWARM["kb-swarm<br/><i>Orchestrator</i>"]
        KB_DISPATCH["kb-dispatch<br/><i>Implementer agent</i>"]
        WORKTREE["Create git worktree"]
        PARSE_AC["Parse ACs from issue"]
        DETECT["Detect code state &<br/>test type"]

        KB_SWARM -->|"assigns issue"| KB_DISPATCH
        KB_DISPATCH --> WORKTREE
        WORKTREE --> PARSE_AC
        PARSE_AC --> DETECT
    end

    subgraph TDD["Phase 2: Per-AC TDD Cycle"]
        direction TB
        RED["RED Phase<br/><i>Write failing test</i>"]
        POST_RED["Post RED_COMPLETE<br/>to Linear"]
        POLL_RED["FOLLOWER MODE<br/><i>Poll Linear every 10s</i>"]
        VERIFY_RED{"Orchestrator<br/>VERIFY_RED"}
        GREEN["GREEN Phase<br/><i>Write minimal impl</i>"]
        POST_GREEN["Post GREEN_COMPLETE<br/>to Linear"]
        POLL_GREEN["FOLLOWER MODE<br/><i>Poll Linear every 10s</i>"]
        VERIFY_GREEN{"Orchestrator<br/>VERIFY_GREEN"}
        REFACTOR["REFACTOR Phase<br/><i>Optional cleanup</i>"]
        NEXT_AC{"More ACs?"}

        RED --> POST_RED
        POST_RED --> POLL_RED
        POLL_RED --> VERIFY_RED
        VERIFY_RED -->|PASSED| GREEN
        VERIFY_RED -->|FAILED| RED
        GREEN --> POST_GREEN
        POST_GREEN --> POLL_GREEN
        POLL_GREEN --> VERIFY_GREEN
        VERIFY_GREEN -->|PASSED| REFACTOR
        VERIFY_GREEN -->|FAILED| GREEN
        REFACTOR --> NEXT_AC
        NEXT_AC -->|Yes| RED
        NEXT_AC -->|No| COMPLETE
    end

    subgraph VERIFICATION["Phase 3: Orchestrator Verification Gates"]
        V_TEST["Run tests<br/><code>pnpm test</code>"]
        V_TYPE["Run typecheck<br/><code>pnpm typecheck</code>"]
        V_LINT["Run lint<br/><code>pnpm lint</code>"]
        V_E2E["E2E: Infra validation<br/><i>simulator, network, stability</i>"]
        V_CODE["Dispatch code-review<br/><i>fresh-eyes agent</i>"]
        V_COUNT["Validate AC count<br/><i>completed vs total</i>"]

        V_TEST --> V_TYPE
        V_TYPE --> V_LINT
        V_LINT --> V_E2E
        V_E2E --> V_CODE
        V_CODE --> V_COUNT
    end

    subgraph REVIEW["Phase 4: Code Review"]
        SUBMIT["kb-submit<br/><i>Move to Review state</i>"]
        ROUTE{"Route by<br/>area label"}
        CR_CONVEX["convex-reviewer"]
        CR_NEXT["nextjs-reviewer"]
        CR_RN["react-native-ui-reviewer"]
        CR_VITE["react-vite-ui-reviewer"]
        CR_CODE["code-reviewer<br/><i>default</i>"]
        CR_AI["ai-tooling-reviewer"]
        CR_PY["python-reviewer"]

        REVIEW_GATES["project-review-gates<br/><i>5-layer validation pyramid</i>"]
        VERDICT{"Verdict"}

        SUBMIT --> ROUTE
        ROUTE -->|"area:convex"| CR_CONVEX
        ROUTE -->|"area:nextjs"| CR_NEXT
        ROUTE -->|"area:ui (RN)"| CR_RN
        ROUTE -->|"area:vite"| CR_VITE
        ROUTE -->|"area:ai-tooling"| CR_AI
        ROUTE -->|"area:python"| CR_PY
        ROUTE -->|"default"| CR_CODE

        CR_CONVEX --> REVIEW_GATES
        CR_NEXT --> REVIEW_GATES
        CR_RN --> REVIEW_GATES
        CR_VITE --> REVIEW_GATES
        CR_AI --> REVIEW_GATES
        CR_PY --> REVIEW_GATES
        CR_CODE --> REVIEW_GATES
        REVIEW_GATES --> VERDICT
    end

    subgraph RESOLUTION["Phase 5: Resolution"]
        APPROVE["kb-approve<br/><i>Move to Done</i>"]
        REJECT["kb-reject<br/><i>Back to Backlog + redo label</i>"]
        UNBLOCK["Unblock dependent issues"]

        VERDICT -->|"APPROVED"| APPROVE
        VERDICT -->|"NEEDS_FIXES"| REJECT
        APPROVE --> UNBLOCK
        REJECT -->|"re-enter dispatch"| KB_DISPATCH
    end

    COMPLETE["ISSUE_COMPLETE<br/><i>All ACs verified</i>"]

    INTAKE --> DISPATCH
    DETECT --> TDD
    COMPLETE --> VERIFICATION
    V_COUNT --> SUBMIT
```

### 2.2 Orchestrator Verification Detail (VERIFY_RED)

```mermaid
flowchart TD
    START["Agent posts RED_COMPLETE"]

    subgraph DETERMINISTIC_GATES["Deterministic Gates"]
        RUN_TEST["Run: pnpm test -- {test_file}"]
        CHECK_EXIT{"Exit code = 0?"}
        VANITY["REJECT: VANITY_TEST<br/><i>Test passes without impl</i>"]
        TEST_OK["Test correctly fails"]

        RUN_TYPE["Run: pnpm typecheck"]
        TYPE_ERR{"Errors?"}
        TYPE_REJECT["REJECT: TYPE_ERRORS_IN_TEST"]
        TYPE_OK["Typecheck passes"]

        RUN_LINT["Run: pnpm lint"]
        LINT_ERR{"Violations?"}
        LINT_REJECT["REJECT: LINT_VIOLATIONS"]
        LINT_OK["Lint passes"]

        RUN_TEST --> CHECK_EXIT
        CHECK_EXIT -->|"Yes (BAD)"| VANITY
        CHECK_EXIT -->|"No (GOOD)"| TEST_OK
        TEST_OK --> RUN_TYPE
        RUN_TYPE --> TYPE_ERR
        TYPE_ERR -->|Yes| TYPE_REJECT
        TYPE_ERR -->|No| TYPE_OK
        TYPE_OK --> RUN_LINT
        RUN_LINT --> LINT_ERR
        LINT_ERR -->|Yes| LINT_REJECT
        LINT_ERR -->|No| LINT_OK
    end

    subgraph E2E_ONLY["E2E-Only Gates (Deterministic)"]
        INFRA_CHECK["Re-check simulator status"]
        INFRA_MATCH{"Matches agent report?"}
        INFRA_REJECT["REJECT: INFRASTRUCTURE_REPORT_MISMATCH"]
        ELEMENT_SCAN["Grep for CSS/XPath selectors"]
        ELEMENT_BAD{"Found unstable locators?"}
        ELEMENT_REJECT["REJECT: UNSTABLE_LOCATORS"]
    end

    RESULT["POST: VERIFY_RED_PASSED"]

    START --> RUN_TEST
    LINT_OK --> INFRA_CHECK
    INFRA_CHECK --> INFRA_MATCH
    INFRA_MATCH -->|No| INFRA_REJECT
    INFRA_MATCH -->|Yes| ELEMENT_SCAN
    ELEMENT_SCAN --> ELEMENT_BAD
    ELEMENT_BAD -->|Yes| ELEMENT_REJECT
    ELEMENT_BAD -->|No| RESULT
    LINT_OK -->|"Unit test path"| RESULT
```

### 2.3 Review Agent Internal Flow

```mermaid
flowchart TD
    START["Reviewer receives issue"]

    subgraph MANDATORY["Mandatory Gates (Deterministic)"]
        TYPECHECK["pnpm typecheck<br/>zero errors required"]
        LINT["pnpm lint<br/>zero errors required"]
        BUILD["Build/Runtime check<br/><i>framework-specific</i>"]
        GATE_PASS{"All gates pass?"}
        IMMEDIATE_REJECT["Immediate NEEDS_FIXES<br/><i>No further review</i>"]
    end

    subgraph TDD_VERIFY["TDD Quality Check (Hybrid)"]
        PARSE_COMMENTS["Parse Linear comments<br/>for VERIFY_RED evidence"]
        COUNT_TESTS["Count: 1 test per AC?"]
        VANITY_SCAN["Detect vanity tests"]
        TDD_OK{"TDD quality OK?"}
        TDD_FAIL["NEEDS_FIXES: TDD violations"]
    end

    subgraph CODE_REVIEW["Code Review (Probabilistic)"]
        READ_FILES["Read full file contents<br/><i>not just diffs</i>"]
        AC_VERIFY["Verify each AC met<br/><i>interpret GIVEN-WHEN-THEN</i>"]
        STANDARDS["Check coding standards<br/><i>patterns, types, imports</i>"]
        SECURITY["Security scan<br/><i>OWASP Top 10</i>"]
        DOMAIN["Domain-specific checks<br/><i>Convex/React/Next.js rules</i>"]
        JUDGMENT["Form verdict<br/><i>weigh severity of findings</i>"]
    end

    subgraph PROJECT_GATES["Project Review Gates (Deterministic)"]
        L1["Layer 1: Syntax"]
        L2["Layer 2: Type Safety"]
        L3["Layer 3: Security"]
        L4["Layer 4: Schema"]
        L5["Layer 5: Specification"]
    end

    VERDICT_OUT{"Final Verdict"}
    APPROVED["APPROVED"]
    NEEDS_FIXES["NEEDS_FIXES<br/>+ revision context"]

    START --> TYPECHECK
    TYPECHECK --> LINT
    LINT --> BUILD
    BUILD --> GATE_PASS
    GATE_PASS -->|No| IMMEDIATE_REJECT
    GATE_PASS -->|Yes| PARSE_COMMENTS

    PARSE_COMMENTS --> COUNT_TESTS
    COUNT_TESTS --> VANITY_SCAN
    VANITY_SCAN --> TDD_OK
    TDD_OK -->|No| TDD_FAIL
    TDD_OK -->|Yes| READ_FILES

    READ_FILES --> AC_VERIFY
    AC_VERIFY --> STANDARDS
    STANDARDS --> SECURITY
    SECURITY --> DOMAIN
    DOMAIN --> JUDGMENT
    JUDGMENT --> L1
    L1 --> L2
    L2 --> L3
    L3 --> L4
    L4 --> L5
    L5 --> VERDICT_OUT

    VERDICT_OUT -->|"All pass"| APPROVED
    VERDICT_OUT -->|"Issues found"| NEEDS_FIXES
```

### 2.4 Players Map

```mermaid
flowchart LR
    subgraph ORCHESTRATORS["Orchestrators (Deterministic Controllers)"]
        KBS["kb-swarm<br/><i>Master orchestrator</i>"]
        KBD["kb-dispatch<br/><i>Issue dispatcher</i>"]
        KBSub["kb-submit<br/><i>Review router</i>"]
        KBA["kb-approve / kb-reject<br/><i>State transitions</i>"]
    end

    subgraph IMPLEMENTERS["Implementers (Probabilistic Coders)"]
        CONV_IMPL["convex-implementer"]
        NEXT_IMPL["nextjs-implementer"]
        RN_IMPL["react-native-ui-implementer"]
        VITE_IMPL["react-vite-ui-implementer"]
        AI_IMPL["ai-tooling-implementer"]
        PY_IMPL["python-implementer"]
    end

    subgraph REVIEWERS["Reviewers (Probabilistic Judges)"]
        CONV_REV["convex-reviewer"]
        NEXT_REV["nextjs-reviewer"]
        RN_REV["react-native-ui-reviewer"]
        VITE_REV["react-vite-ui-reviewer"]
        AI_REV["ai-tooling-reviewer"]
        PY_REV["python-reviewer"]
        CODE_REV["code-reviewer<br/><i>default/general</i>"]
    end

    subgraph VALIDATORS["Validators (Deterministic Tools)"]
        REACT_GATES["react-validation-gates<br/><i>6 universal gates</i>"]
        PROJ_GATES["project-review-gates<br/><i>5-layer pyramid</i>"]
        INT_VAL["integration-validator<br/><i>stub migration</i>"]
        TDD_LIB["tdd-lib<br/><i>test execution utils</i>"]
    end

    subgraph PLANNERS["Planners (Probabilistic Strategists)"]
        PM["product-manager"]
        EM["engineering-manager"]
        UID["ui-designer"]
        PLNR["planner"]
    end

    subgraph INFRASTRUCTURE["Infrastructure (Deterministic)"]
        LINEAR["Linear API<br/><i>State machine + comments</i>"]
        GIT["Git worktrees"]
        CLI["CLI tools<br/><i>pnpm, typecheck, lint</i>"]
    end

    KBS --> KBD
    KBD --> CONV_IMPL
    KBD --> NEXT_IMPL
    KBD --> RN_IMPL
    KBSub --> CONV_REV
    KBSub --> CODE_REV
    CONV_REV --> PROJ_GATES
    CONV_REV --> REACT_GATES
    KBS --> LINEAR
    KBD --> GIT
    PROJ_GATES --> CLI
```

---

## 3. Deterministic vs Probabilistic Analysis

### 3.1 Classification Framework

**Deterministic** = Traditional software engineering: the same input always produces the same output. Testable with unit tests. Bugs are logic errors, not judgment errors.

**Probabilistic** = AI judgment: the output depends on LLM reasoning, context interpretation, and heuristic weighing. The same input may produce different outputs across runs. Quality is measured by consistency and calibration, not correctness.

### 3.2 Deterministic Processes

These are the parts where traditional software engineering should be applied. They should be rock-solid, well-tested, and never left to AI judgment.

| Process | Skill/Component | Why Deterministic |
|---------|----------------|-------------------|
| **Linear state transitions** | kb-move, kb-approve, kb-reject, kb-submit | Finite state machine: Backlog → In Progress → Review → Done. Binary transitions with no ambiguity. |
| **Comment polling loop** | kb-dispatch (FOLLOWER MODE) | Poll every 10s, pattern-match for `VERIFY_RED_PASSED` or `VERIFY_RED_FAILED`. Pure string matching. |
| **Comment format parsing** | kb-swarm monitoring loop | Regex/structured parsing of `TDD_PHASE:`, `AC:`, `TEST_FILE:`, etc. Fixed grammar. |
| **Test execution** | Orchestrator VERIFY gates | `pnpm test` returns exit code 0 or non-0. Binary. No interpretation. |
| **Typecheck execution** | All reviewers + orchestrator | `pnpm typecheck` returns errors or clean. Binary. |
| **Lint execution** | All reviewers + orchestrator | `pnpm lint` returns violations or clean. Binary. |
| **Vanity test detection** | VERIFY_RED gate | If test passes (exit 0) without implementation → vanity. Pure logic: exit code check. |
| **AC count validation** | ISSUE_COMPLETE verification | `len(completed_acs) == total_acs`. Arithmetic comparison. |
| **E2E infrastructure check** | VERIFY_RED (E2E path) | `xcrun simctl list`, `curl localhost:8081/status`, `ping`. Binary: up or down. |
| **Locator stability scan** | VERIFY_RED (E2E path) | `grep -c 'getByTestId'` vs `grep -c 'css\|xpath'`. Pattern count comparison. |
| **Infrastructure report mismatch** | VERIFY_GREEN (E2E path) | Compare agent-reported values vs orchestrator-measured values. Equality check. |
| **Flakiness detection** | VERIFY_GREEN (E2E path) | Run test 3 times, `pass_rate < 100%` → flaky. Arithmetic. |
| **Stability metrics mismatch** | VERIFY_GREEN (E2E path) | Compare agent-reported `pass_rate` vs orchestrator-measured. Equality check. |
| **Area-label routing** | kb-submit | `area:convex → convex-reviewer`, `area:ui → react-native-ui-reviewer`. Lookup table. |
| **WSJF calculation** | kb-project-wsjf | `(BV + TC + RR) / multiplier`. Pure arithmetic formula once inputs are set. |
| **Worktree lifecycle** | kb-dispatch, kb-swarm | `git worktree add/remove`. File system operations with deterministic outcomes. |
| **State file tracking** | .kb-swarm-state.json | JSON read/write of session state. CRUD operations. |
| **Dependency unblocking** | kb-approve | When issue moves to Done, query for issues `blockedBy` this one. Graph traversal. |
| **Review gates pyramid** | project-review-gates layers 1-2 | Syntax (parseable?) and Type Safety (typecheck passes?) are binary tool outputs. |
| **Workflow selection** | kb-dispatch code state detection | File missing → CLASSIC_TDD, file exists → GAP_ANALYSIS. File existence check. |
| **WIP limit enforcement** | kb-move | Count issues in state, compare to limit. Arithmetic. |

### 3.3 Probabilistic Processes

These are the parts that are inherently AI tasks - they require judgment, interpretation, and reasoning that doesn't reduce to binary logic.

| Process | Skill/Component | Why Probabilistic |
|---------|----------------|-------------------|
| **Writing test code** | kb-dispatch RED phase | Interpreting a GIVEN-WHEN-THEN AC and producing a test that faithfully captures the intent. Requires understanding of the domain, choosing assertions, mocking strategy. |
| **Writing implementation code** | kb-dispatch GREEN phase | "Minimal code to make the test pass" requires judgment about what's minimal, what patterns to use, how to structure the solution. |
| **Refactoring decisions** | kb-dispatch REFACTOR phase | Identifying duplication, naming improvements, extraction candidates. Aesthetic and structural judgment. |
| **AC interpretation** | All reviewers | Reading "GIVEN a user is logged in WHEN they click profile THEN they see their name" and determining whether the code actually satisfies this. Semantic understanding. |
| **Code quality judgment** | code-review, all domain reviewers | "Is this code clean?" "Does this follow the pattern?" "Is this abstraction appropriate?" These are judgment calls even with standards documents. |
| **Security vulnerability detection** | security-review, all reviewers | Identifying OWASP Top 10 issues requires understanding attack vectors in context. A `user_input` in one context is safe, in another it's XSS. |
| **Test quality grading** | test-review | Assigning A-F grades to test suites. Detecting "vanity tests" that technically pass but don't test real behavior. Requires understanding test intent. |
| **Coding standards interpretation** | code-review | "Context-dependent" rules like `useCallback` being valid only with `React.memo` parent. Requires tracing component relationships. |
| **Design review** | design-reviewer | Evaluating if a mockup is 80%+ aligned with a specification. Visual and conceptual judgment. |
| **PRD/TRD analysis** | spec-review-swarm | Cross-referencing specification documents against codebase implementation. Understanding intent vs. implementation. |
| **WSJF input scoring** | kb-project-wsjf | Business Value (1-10), Time Criticality (1-10), Risk Reduction (1-10) are subjective scores that an AI must assign. The formula is deterministic, the inputs are not. |
| **Implementation planning** | kb-project-analyze | Deciding task breakdown, dependency ordering, agent assignment. Multiple valid decompositions exist. |
| **Revision feedback** | All reviewers (NEEDS_FIXES) | Writing actionable feedback: "what to fix, why, how." Requires understanding the developer's intent and suggesting alternatives. |
| **Cross-track synthesis** | spec-review-swarm | Merging findings from 4 parallel reviewers, resolving conflicts, aggregating severity. Weighting disagreements. |
| **Gap analysis triage** | kb-dispatch GAP_ANALYSIS path | When existing code partially satisfies an AC, deciding whether to augment or rewrite. Judgment call. |
| **Review gates pyramid layers 3-5** | project-review-gates | Security (OWASP interpretation), Schema (migration safety assessment), Specification (TRD compliance) require contextual AI judgment. |
| **Attempt log reasoning** | All reviewers | "What was tried, what failed, what to avoid" requires understanding failure patterns and suggesting novel approaches. |
| **Domain routing edge cases** | kb-submit | When an issue touches multiple areas (e.g., Convex + UI), choosing the primary reviewer requires judgment. |

### 3.4 Hybrid Processes (Deterministic Shell, Probabilistic Core)

These are processes where the **workflow** is deterministic but the **substance** within each step is probabilistic.

| Process | Deterministic Shell | Probabilistic Core |
|---------|--------------------|--------------------|
| **TDD cycle** | RED → GREEN → REFACTOR sequence is fixed | What code to write at each step is AI judgment |
| **Per-AC iteration** | "For each AC, do RED then GREEN" is a loop | Interpreting each AC and producing code is judgment |
| **Reviewer dispatch** | Area label → reviewer mapping is a lookup | The review itself is probabilistic reasoning |
| **Monitoring loop** | "Poll, parse, gate-check, post" is mechanical | code-review dispatch within VERIFY_GREEN is probabilistic |
| **Review gates pyramid** | "Run layers 1-5 in order, fail-fast" is deterministic | Layers 3-5 content analysis is probabilistic |
| **kb-project-analyze** | "Spawn PM, EM, UID → collect → merge" is orchestration | Each agent's analysis is probabilistic |
| **spec-review-swarm** | "Decompose into tracks, dispatch, synthesize" is orchestration | Each track's review is probabilistic |
| **design-review-loop** | "Generate → review → regenerate until approved" is a loop | Review judgment and regeneration are probabilistic |

---

## 4. Architectural Implications

### 4.1 Where to Invest in Traditional Engineering

The deterministic layer is the **load-bearing structure** of the system. Failures here cascade into the probabilistic layer and are hard to diagnose. Invest in:

1. **State machine correctness**: The Linear state transitions (Backlog → In Progress → Review → Done) should be formally modeled and tested. Invalid transitions should be impossible, not just discouraged.

2. **Comment protocol parsing**: The structured comment format (`TDD_PHASE: RED_COMPLETE`, etc.) is the inter-agent communication protocol. A parsing error here breaks the entire feedback loop. This deserves a formal grammar and parser tests.

3. **Gate execution reliability**: `pnpm test`, `pnpm typecheck`, `pnpm lint` are the ground truth. Flaky execution, timeout handling, and exit code interpretation must be bulletproof.

4. **AC count arithmetic**: The `completed_acs == total_acs` check is the final integrity gate. Off-by-one errors or parsing failures here let incomplete work through.

5. **Polling/timeout logic**: The 10-second poll + 5-minute timeout in FOLLOWER MODE is a distributed coordination mechanism. Race conditions, missed messages, and timeout cascades are traditional concurrency bugs.

### 4.2 Where to Invest in AI Quality

The probabilistic layer determines the **output quality** of the system. The code it writes and the reviews it performs are only as good as the prompts, context, and constraints it operates under. Invest in:

1. **Prompt engineering for reviewers**: The adversarial mindset ("assume bugs exist") is calibrated by prompt. Too aggressive → false positives that waste cycles. Too lenient → real bugs slip through.

2. **Context window management**: Reviewers must "read every line" but context windows are finite. The code-review-setup pattern (diff → full file read) is a pragmatic tradeoff. Monitoring whether reviewers actually read vs. skim.

3. **Consistency across runs**: The same code reviewed twice may get different verdicts. For high-stakes decisions (APPROVED vs NEEDS_FIXES), consider requiring consensus from multiple reviewers or using structured checklists to reduce variance.

4. **Feedback quality**: NEEDS_FIXES feedback must be actionable. Vague feedback ("improve error handling") creates expensive re-review cycles. Structured feedback templates help.

5. **Calibration of WSJF inputs**: Business Value, Time Criticality, and Risk Reduction scores are subjective. Track whether AI-assigned scores correlate with actual outcomes to calibrate over time.

### 4.3 The Critical Boundary

The most important architectural boundary is where **deterministic gates guard against probabilistic failures**:

```
Probabilistic: Agent writes code and claims "all tests pass"
                    ↓
Deterministic: Orchestrator ACTUALLY runs tests independently
                    ↓
Deterministic: Exit code 0 or not. No AI involved.
```

This pattern - **never trust the agent's self-assessment, always verify with deterministic tools** - is the fundamental safety mechanism of the system. It appears in:

- Vanity test detection (agent says RED, orchestrator verifies test actually fails)
- Infrastructure report validation (agent says sim is booted, orchestrator re-checks)
- Stability metrics verification (agent says 100% pass rate, orchestrator re-runs)
- AC count validation (agent says "all done", orchestrator counts)

**Any weakening of this boundary is a systemic risk.** If the orchestrator ever trusts an agent's claim without independent verification, the entire quality guarantee collapses.

---

## 5. Risk Analysis

### 5.1 High-Risk Deterministic Failures

| Risk | Impact | Mitigation |
|------|--------|------------|
| Comment parsing breaks silently | Agent stuck in FOLLOWER MODE forever (5min timeout) | Integration tests for all comment formats |
| Linear API rate limiting | Polling loop fails, orchestrator loses sync | Exponential backoff, state recovery from .kb-swarm-state.json |
| Git worktree conflicts | Multiple agents clobber each other's work | Worktree-per-issue isolation (already implemented) |
| Typecheck/lint version drift | Same code passes locally, fails in gate | Pin tool versions, run in consistent environment |

### 5.2 High-Risk Probabilistic Failures

| Risk | Impact | Mitigation |
|------|--------|------------|
| Reviewer approves buggy code | Bug reaches production | Multiple reviewer rounds, project-review-gates as second check |
| Agent writes test that doesn't actually test the AC | False confidence in coverage | test-review skill grades test quality, vanity test detection |
| WSJF scores are miscalibrated | Wrong work prioritized | Track outcomes, recalibrate scoring rubric |
| Reviewer and implementer disagree on AC meaning | Infinite revision loop | AC format standardization (GIVEN-WHEN-THEN), human escalation |
| Code review feedback is vague | Expensive re-review cycles | Structured feedback templates with file:line specificity |

---

## 6. Summary

The KB review ecosystem is a **deterministic orchestration framework with probabilistic actors**. The brilliance of the design is that AI agents are never trusted - they operate inside a cage of deterministic verification gates that independently confirm every claim. The deterministic layer (state machines, polling, parsing, tool execution) should be treated as critical infrastructure with full test coverage. The probabilistic layer (code writing, reviewing, planning) should be treated as a quality optimization problem - measured by consistency, calibration, and outcome tracking rather than binary correctness.

### Classification Ratio

Across the ~45 distinct processes identified in this analysis:

- **~18 (40%) are purely deterministic** - state transitions, tool execution, arithmetic, pattern matching
- **~15 (33%) are purely probabilistic** - code writing, reviewing, planning, judgment
- **~12 (27%) are hybrid** - deterministic workflow wrapping probabilistic substance

The system's reliability comes from ensuring that **every probabilistic output passes through at least one deterministic gate** before being accepted as truth.
