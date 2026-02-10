# Starfleet Software Engineering Doctrine

## Issued by Captain Jean-Luc Picard

### For the Operation of Autonomous Engineering Systems

---

## Purpose

Starfleet employs autonomous and semi-autonomous systems to accelerate software development.  
Such systems increase velocity, but also amplify error when improperly constrained.

This doctrine defines **software engineering practice**, not metaphor.  
Starfleet terminology is used to establish shared mental models, never to obscure technical intent.

---

## Foundational Principle

**We optimize for verified learning, not speculative completeness.**

In software systems of sufficient complexity, correctness cannot be inferred from intent.  
It must be demonstrated through execution and measurement.

If a change cannot be evaluated quickly, it is too large.  
If a change cannot be observed, it is not yet real.

---

## Engineering Rule 1: Single-Mission Completion

All engineering work must complete within a **single mission**.

A mission consists of:

**Define → Verify → Implement → Validate → Decide**

A mission must:

- Fit within one bounded execution context
- Produce a definitive outcome (accept / reject / split)
- Leave the system in a reversible state

Work that exceeds a single mission is not an implementation task — it is planning input and must be decomposed.

This rule is enforced by orchestration, not by convention.

---

## Engineering Rule 2: Observable System Change

Every mission must produce an **observable system change**.

Acceptable artifacts include:

- A test that previously failed and now passes
- A CLI or API invocation with altered output
- A visible UI or state change
- A schema, invariant, or diagnostic difference

Aesthetic quality is secondary.  
Observability is mandatory.

If no concrete change can be demonstrated, the mission has not completed.

---

## Deterministic Systems vs Engineering Judgment

Starfleet explicitly separates **deterministic engineering systems** from **probabilistic judgment systems**.

### Deterministic Systems (Must Be Trusted)

- State machines and workflow transitions
- Test, typecheck, lint, and build execution
- Parsing, counting, routing, and locking
- Tool invocation and exit-code evaluation

These systems must be predictable, testable, and free from interpretation.

They are treated as production infrastructure.

---

### Judgment Systems (Must Be Verified)

- Writing implementation code
- Writing or selecting tests
- Interpreting acceptance criteria
- Code review and design evaluation

Judgment systems may vary in quality.

Their outputs are **never accepted without deterministic verification**.

No agent validates its own work.

---

## Orchestration Authority

Autonomous agents act as implementers, not decision-makers.

The orchestration system holds authority over:

- Mission scope and decomposition
- Execution order and concurrency
- Verification requirements
- Acceptance or rejection of results

Agents do not define success criteria.  
They produce candidates for verification.

---

## Operational Risk Modes

Starfleet distinguishes engineering work by operational risk.

### Red Alert Missions (Test-Required)

Declared when failure risks:

- Business or domain logic
- Data integrity or migrations
- Authentication, authorization, billing
- Public APIs or contracts
- Bug fixes or regressions

Requirements:

- Verification defined before implementation
- Mandatory automated tests
- Full deterministic gate execution
- Observable mission artifact

---

### Standard Operations Missions (Test-Optional)

Declared for:

- Styling and presentation changes
- Non-behavioral refactors
- Tooling and documentation
- Low-risk infrastructure updates

Requirements:

- Deterministic baseline gates (typecheck, lint, build)
- Observable artifact demonstrating the change

Tests are used where they reduce uncertainty, not by default.

---

## Human Oversight

Human engineers are not expected to review all missions.

Their responsibilities are to:

- Define what constitutes sufficient evidence
- Calibrate risk thresholds
- Review missions involving high uncertainty
- Adjust orchestration rules as the system evolves

Human attention is reserved for ambiguity, not volume.

---

## Compounding Engineering Confidence

Early development emphasizes frequent demonstration and inspection.

As behavior becomes well understood:

- Tests encode expectations
- Gates enforce invariants
- Manual review frequency decreases

When uncertainty reappears, rigor increases immediately.

Confidence compounds only when learning is preserved in code and verification.

---

## System Health Criteria

At any time, the system must be able to answer:

1. What mission is currently executing?
2. What observable result will it produce?
3. What occurs if the mission fails?
4. How easily can the change be reverted?

If these questions are answerable, the system is operating correctly — regardless of current velocity.

---

## Standing Engineering Directive

**No change is accepted unless it completes within a single mission.  
No change is complete unless it produces an observable system effect.**

Starfleet advances through discipline, not assumption.  
The same must be true of our software.
