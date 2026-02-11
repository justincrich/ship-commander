# Task: Commander Risk Classification System

**Epic**: Core Commission & Planning (COMM)
**Priority**: HIGH
**Classification**: RED_ALERT (affects mission execution correctness)

## Problem Statement

The Commander agent (UC-COMM-05) currently lacks a systematic framework for classifying missions as RED_ALERT vs STANDARD_OPS during the Ready Room planning phase. Classification is documented in the PRD but not operationalized as:

1. An assessable decision framework
2. An invokeable prompt for the Commander
3. A reproducible workflow for task-by-task classification

## Requirements

### AC-1: Commander Risk Assessment Framework

**Given**: A use case decomposed into one or more missions
**When**: The Commander analyzes each mission during Ready Room
**Then**: The mission is classified using the following deterministic framework:

#### Decision Tree

```
Does the change affect system correctness or behavior?
│
├─ YES → RED_ALERT
│  │
│  ├─ Does it change business logic?
│  ├─ Does it modify API contracts?
│  ├─ Does it touch auth/security?
│  ├─ Does it affect data integrity?
│  └─ Is it a bug fix?
│
└─ NO → STANDARD_OPS
   │
   ├─ Is it purely visual/styling?
   ├─ Is it a non-behavioral refactor?
   ├─ Is it tooling/automation?
   └─ Is it documentation?
```

**Acceptance Criteria**:
- [ ] Framework documented in `internal/commander/classification.go`
- [ ] Each criterion has explicit yes/no questions
- [ ] Default to RED_ALERT on ambiguity
- [ ] Classification rationale persisted to mission bead

### AC-2: Commander Classification Prompt

**Given**: The Commander agent during Ready Room planning phase
**When**: The Commander needs to classify a mission
**Then**: The Commander invokes a structured prompt that produces:

```yaml
mission_id: "MISSION-XXX"
title: "Mission title here"
classification: "RED_ALERT" | "STANDARD_OPS"
rationale:
  affects_behavior: true | false
  criteria_matched:
    - "business_logic" | "api_changes" | "auth_security" |
      "data_integrity" | "bug_fix" | "styling" |
      "non_behavioral_refactor" | "tooling" | "documentation"
  risk_assessment: "Detailed explanation"
  confidence: "high" | "medium" | "low"
```

**Acceptance Criteria**:
- [ ] Prompt template defined in `internal/commander/prompts.go`
- [ ] Template receives: mission title, functional requirements, design requirements, use case context
- [ ] Returns structured YAML output
- [ ] Includes confidence score (requires Admiral review if low)

### AC-3: Classification Validation Gate

**Given**: A classified mission from the Commander
**When**: Classification is complete
**Then**: System validates that:

1. RED_ALERT missions require:
   - `tests` in demo token proof
   - At least one of: `commands` or `diff_refs`

2. STANDARD_OPS missions require:
   - At least one of: `commands`, `manual_steps`, `diff_refs`

**Acceptance Criteria**:
- [ ] Validator function in `internal/state/validation.go`
- [ ] Returns error if classification doesn't match proof requirements
- [ ] Validation happens during Ready Room consensus check (UC-COMM-07)

### AC-4: Task-by-Task Classification Workflow

**Given**: A commission with N missions
**When**: Ready Room planning produces mission manifest
**Then**: Commander classifies each mission sequentially:

```
For each mission in manifest:
  1. Extract functional requirements (Captain output)
  2. Extract design requirements (Design Officer output)
  3. Invoke classification prompt
  4. Validate classification
  5. Persist classification to mission bead state
  6. If confidence < high, flag for Admiral review
```

**Acceptance Criteria**:
- [ ] Workflow implemented in `internal/commander/readyroom.go`
- [ ] Classification happens before consensus validation
- [ ] Low-confidence classifications trigger Admiral question (UC-COMM-08)
- [ ] All classifications visible in Plan Review (UC-TUI-03)

### AC-5: TUI Classification Display

**Given**: Admiral reviewing mission manifest
**When**: Displaying mission list in Plan Review
**Then**: Each mission shows:

```
MISSION-042: Add login button
  Classification: [RED_ALERT] 🔴
  Criteria: business_logic, auth_security
  Confidence: high

MISSION-043: Update button colors
  Classification: [STANDARD_OPS] 🟢
  Criteria: styling
  Confidence: high
```

**Acceptance Criteria**:
- [ ] Classification badge in Plan Review view (UC-TUI-03)
- [ ] Color-coded: RED_ALERT = red_alert, STANDARD_OPS = green_ok
- [ ] Hover/expand shows rationale and criteria
- [ ] Low-confidence missions highlighted with warning

## Implementation Plan

### Phase 1: Core Classification (RED → VERIFY_RED → GREEN)

1. Create `internal/commander/classification.go` with decision tree
2. Create `internal/commander/prompts.go` with classification prompt template
3. Add `classification` field to Mission bead state schema
4. Implement `ClassifyMission()` function

### Phase 2: Validation & Workflow (VERIFY_GREEN → REFACTOR)

1. Create validation function in `internal/state/validation.go`
2. Integrate classification into Ready Room workflow
3. Add low-confidence Admiral question trigger
4. Persist classification rationale to mission comments

### Phase 3: TUI Display (VERIFY_REFACTOR)

1. Add classification badge component to TUI theme
2. Update Plan Review view with classification display
3. Add rationale expansion on user interaction
4. Update manifest Glamour rendering

## Commander Classification Prompt Template

```go
const ClassificationPrompt = `
You are the Commander assessing mission risk level for Ship Commander 3.

MISSION TITLE: {{ .Title }}
USE CASE: {{ .UseCase }}

FUNCTIONAL REQUIREMENTS (Captain):
{{ .FunctionalReqs }}

DESIGN REQUIREMENTS (Design Officer):
{{ .DesignReqs }}

CLASSIFY THIS MISSION by answering:

1. Does this change affect system correctness or behavior? (YES/NO)
2. If YES, which criteria apply?
   - business_logic: Changes how system behaves from user perspective?
   - api_changes: Modifies endpoints, request/response formats?
   - auth_security: Touches authentication, authorization, permissions?
   - data_integrity: Could corrupt data or violate invariants?
   - bug_fix: Fixes incorrect behavior?

3. If NO, which criteria apply?
   - styling: Purely visual (CSS, tokens, colors)?
   - non_behavioral_refactor: Structural change without behavior change?
   - tooling: Developer tooling, not production code?
   - documentation: Documentation only?

4. What is your confidence level? (high/medium/low)

OUTPUT YAML:
classification: RED_ALERT | STANDARD_OPS
rationale:
  affects_behavior: true | false
  criteria_matched: [list of applicable criteria]
  risk_assessment: "2-3 sentence explanation"
  confidence: high | medium | low
`
```

## Testing Strategy

### Unit Tests

```go
func TestClassifyBusinessLogic(t *testing.T) {
    mission := Mission{
        Title: "Add user registration flow",
        FunctionalReqs: "System must allow users to register with email",
    }
    result := ClassifyMission(mission)
    assert.Equal(t, "RED_ALERT", result.Classification)
    assert.Contains(t, result.Criteria, "business_logic")
}

func TestClassifyStyling(t *testing.T) {
    mission := Mission{
        Title: "Update button color to butterscotch",
        FunctionalReqs: "Change button color from blue to butterscotch",
    }
    result := ClassifyMission(mission)
    assert.Equal(t, "STANDARD_OPS", result.Classification)
    assert.Contains(t, result.Criteria, "styling")
}
```

### Integration Tests

- Ready Room with 3 missions → all classified correctly
- Low-confidence classification → Admiral question triggered
- Invalid classification (RED_ALERT without test requirement) → validation error

## Dependencies

- UC-COMM-05 (Commander mission decomposition)
- UC-COMM-07 (Consensus validation)
- UC-COMM-08 (Admiral questions)
- UC-TUI-03 (Plan Review display)

## Definition of Done

- [ ] All ACs pass with TDD (RED → VERIFY_RED → GREEN → VERIFY_GREEN → REFACTOR → VERIFY_REFACTOR)
- [ ] Commander can classify missions during Ready Room
- [ ] Classification persisted to mission beads
- [ ] TUI displays classification in Plan Review
- [ ] Low-confidence classifications trigger Admiral review
- [ ] Demo token validation enforces classification-proof consistency
- [ ] Tests cover all classification criteria
- [ ] Documentation updated in TRD

---

**Created**: 2026-02-10
**Assignee**: Commander Agent
**Reviewer**: TBD
**Estimated**: 4-6 hours
