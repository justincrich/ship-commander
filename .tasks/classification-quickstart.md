# Commander Risk Classification - Quick Start

## Overview

This system enables the Commander agent to classify missions as **RED_ALERT** (full TDD) or **STANDARD_OPS** (fast path) during the Ready Room planning phase.

## Files Created

1. **`.tasks/commander-risk-classification.md`** - Full task specification with ACs, implementation plan, and testing strategy
2. **`.prompts/classify-mission-risk.md`** - Invokeable prompt template for the Commander agent

## Usage Workflow

### Step 1: During Ready Room (UC-COMM-05)

When the Commander decomposes use cases into missions:

```go
// Pseudocode for Commander workflow
for each mission in manifest {
    // 1. Gather requirements from agents
    funcReqs := Captain.GetFunctionalRequirements(mission)
    designReqs := DesignOfficer.GetDesignRequirements(mission)

    // 2. Build prompt context
    context := ClassificationContext{
        MissionID: mission.ID,
        Title: mission.Title,
        UseCase: mission.UseCase,
        FunctionalRequirements: funcReqs,
        DesignRequirements: designReqs,
        CommissionTitle: commission.Title,
        Domain: commission.Domain,
        Dependencies: mission.Dependencies,
    }

    // 3. Invoke classification prompt
    result := ClassifyMission(context)

    // 4. Validate classification
    if err := ValidateClassification(result); err != nil {
        // Flag for Admiral review
        result.AdmiralReviewRequired = true
    }

    // 5. Persist to mission bead
    mission.State.Classification = result.Classification
    mission.State.ClassificationRationale = result.Rationale
    mission.State.Confidence = result.Confidence

    // 6. Trigger Admiral question if low confidence
    if result.Confidence != "high" {
        AskAdmiralQuestion(
            Question: fmt.Sprintf("Mission %s classified as %s with %s confidence. Confirm?",
                mission.ID, result.Classification, result.Confidence),
            Options: ["Confirm", "Reclassify as RED_ALERT", "Reclassify as STANDARD_OPS"],
        )
    }
}
```

### Step 2: Validate Proof Requirements (UC-EXEC-07)

When mission completes and demo token is submitted:

```go
func ValidateDemoTokenClassification(mission Mission, token DemoToken) error {
    switch mission.State.Classification {
    case "RED_ALERT":
        // Must have tests
        if !token.HasTests() {
            return fmt.Errorf("RED_ALERT mission requires tests in demo token")
        }
        // Must have commands OR diff_refs
        if !token.HasCommands() && !token.HasDiffRefs() {
            return fmt.Errorf("RED_ALERT mission requires commands or diff_refs")
        }

    case "STANDARD_OPS":
        // Must have at least one proof type
        if !token.HasCommands() && !token.HasManualSteps() && !token.HasDiffRefs() {
            return fmt.Errorf("STANDARD_OPS mission requires at least one proof type")
        }
    }
    return nil
}
```

### Step 3: Display in TUI (UC-TUI-03)

Plan Review view shows:

```go
// Render mission with classification
func RenderMissionRow(mission Mission) string {
    badge := GetClassificationBadge(mission.State.Classification)
    confidence := GetConfidenceIndicator(mission.State.Confidence)

    return fmt.Sprintf("%s: %s\n  Classification: %s %s\n  Criteria: %s\n  Confidence: %s",
        mission.ID,
        mission.Title,
        badge,
        GetClassificationEmoji(mission.State.Classification),
        mission.State.Rationale.PrimaryCriterion,
        confidence,
    )
}

// Badge styles
const (
    RED_ALERTBadge   = "[RED_ALERT] " + redAlertColor   // 🔴
    STANDARD_OPSBadge = "[STANDARD_OPS] " + greenOkColor // 🟢
)

// Confidence indicators
const (
    HighConfidence   = "●"   // solid
    MediumConfidence = "◐"   // half-full
    LowConfidence    = "○"   // empty + warning color
)
```

## Prompt Template Variables

The `.prompts/classify-mission-risk.md` template uses these variables:

```go
type ClassificationContext struct {
    MissionID              string
    Title                  string
    UseCase                string
    FunctionalRequirements string
    DesignRequirements     string
    CommissionTitle        string
    Domain                 string
    Dependencies           []string
}
```

## Decision Tree Reference

```
┌─────────────────────────────────────┐
│ Does mission affect behavior?       │
└─────────────────┬───────────────────┘
                  │
         ┌────────┴────────┐
         │                 │
        YES               NO
         │                 │
    RED_ALERT      STANDARD_OPS
         │                 │
    ◤────────┴────────◐    │
    │                 │    │
business_logic   styling
api_changes      non_behavioral_refactor
auth_security    tooling
data_integrity   documentation
bug_fix
```

## Examples at a Glance

| Mission | Classification | Why? |
|---------|----------------|------|
| Add user registration | RED_ALERT | Business logic + API + data |
| Update button colors | STANDARD_OPS | Styling only |
| Fix null pointer crash | RED_ALERT | Bug fix |
| Extract utility function | STANDARD_OPS | Non-behavioral refactor |
| Add rate limiting | RED_ALERT | Business logic + API |
| Update README | STANDARD_OPS | Documentation |
| Refactor component tree | STANDARD_OPS | Non-behavioral refactor |
| Add authentication | RED_ALERT | Auth security |

## Testing the Classifier

### Unit Test Example

```go
func TestClassifier(t *testing.T) {
    tests := []struct {
        name     string
        mission  ClassificationContext
        want     ClassificationResult
    }{
        {
            name: "Business logic → RED_ALERT",
            mission: ClassificationContext{
                Title: "Add user registration",
                FunctionalRequirements: "Users can register with email",
            },
            want: ClassificationResult{
                Classification: "RED_ALERT",
                Rationale: Rationale{
                    AffectsBehavior:  true,
                    PrimaryCriterion: "business_logic",
                },
                Confidence: "high",
            },
        },
        {
            name: "Styling → STANDARD_OPS",
            mission: ClassificationContext{
                Title: "Update button colors",
                FunctionalRequirements: "Change blue to butterscotch",
            },
            want: ClassificationResult{
                Classification: "STANDARD_OPS",
                Rationale: Rationale{
                    AffectsBehavior:  false,
                    PrimaryCriterion: "styling",
                },
                Confidence: "high",
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := ClassifyMission(tt.mission)
            assert.Equal(t, tt.want, got)
        })
    }
}
```

### Integration Test

```go
func TestReadyRoomClassification(t *testing.T) {
    // Setup: Create commission with 3 missions
    commission := createTestCommission()

    // Run Ready Room
    manifest := RunReadyRoom(commission)

    // Assert: All missions classified
    for _, mission := range manifest.Missions {
        assert.NotEmpty(t, mission.State.Classification,
            "Mission %s should have classification", mission.ID)
        assert.NotEmpty(t, mission.State.Rationale,
            "Mission %s should have rationale", mission.ID)
    }

    // Assert: Low confidence triggers Admiral question
    lowConfMissions := filterByConfidence(manifest, "low")
    assert.Equal(t, 1, len(lowConfMissions),
        "Should trigger Admiral review for low confidence")
}
```

## Admiral Review Questions

When `confidence != "high"`, the system prompts the Admiral:

```
┌─────────────────────────────────────────────────┐
│ ADMIRAL — QUESTION FROM COMMANDER              │
│                                                 │
│ Mission MISSION-042 classified as RED_ALERT     │
│ with LOW confidence. Please review.             │
│                                                 │
│ Rationale:                                      │
│ Primary: non_behavioral_refactor                │
│ Risk: Refactor touches auth logic               │
│                                                 │
│ Options:                                        │
│ [1] Confirm RED_ALERT (proceed with TDD)       │
│ [2] Reclassify as STANDARD_OPS (fast path)     │
│ [3] Request additional analysis                │
│                                                 │
│ [Enter] to select, [Esc] to skip                │
└─────────────────────────────────────────────────┘
```

## Implementation Checklist

- [ ] Create `internal/commander/classification.go` with `ClassifyMission()`
- [ ] Create `internal/commander/prompts.go` with prompt template
- [ ] Add `classification` field to Mission bead state
- [ ] Implement validation in `internal/state/validation.go`
- [ ] Integrate into Ready Room workflow (UC-COMM-05)
- [ ] Add TUI classification badge (UC-TUI-03)
- [ ] Add demo token proof validation (UC-EXEC-07)
- [ ] Write unit tests for all classification criteria
- [ ] Write integration test for Ready Room flow
- [ ] Update TRD with classification framework

## Quick Reference: Classification Rules

| Criterion | Classification | Tests Required | Examples |
|-----------|----------------|----------------|----------|
| **business_logic** | RED_ALERT | ✅ Yes | User registration, payment flow |
| **api_changes** | RED_ALERT | ✅ Yes | New endpoint, modified schema |
| **auth_security** | RED_ALERT | ✅ Yes | Login, permissions, RBAC |
| **data_integrity** | RED_ALERT | ✅ Yes | DB migration, data processing |
| **bug_fix** | RED_ALERT | ✅ Yes | Crash fix, logic error |
| **styling** | STANDARD_OPS | ❌ No | Colors, fonts, spacing |
| **non_behavioral_refactor** | STANDARD_OPS | ❌ No | Extract function, rename |
| **tooling** | STANDARD_OPS | ❌ No | Scripts, configs, build tools |
| **documentation** | STANDARD_OPS | ❌ No | README, API docs |

**Default**: When uncertain → **RED_ALERT**

---

**Next Steps**:
1. Review task specification in `.tasks/commander-risk-classification.md`
2. Test prompt with sample missions using Commander agent
3. Implement classifier in `internal/commander/classification.go`
4. Add validation and TUI integration
5. Run integration tests and update documentation
