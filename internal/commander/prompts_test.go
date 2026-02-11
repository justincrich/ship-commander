package commander

import (
	"strings"
	"testing"
)

func TestBuildPlanningPromptIncludesManifestInstructions(t *testing.T) {
	t.Parallel()

	prompt, err := BuildPlanningPrompt(PlanningPromptContext{
		CommissionTitle: "Headless PRD",
		CommissionPRD:   "Build the commander loop",
		UseCases:        []string{"UC-PROMPT-01", "UC-PROMPT-02"},
	})
	if err != nil {
		t.Fatalf("build planning prompt: %v", err)
	}

	for _, needle := range []string{"Headless PRD", "Build the commander loop", "UC-PROMPT-01", "Return ONLY YAML mission manifest"} {
		if !strings.Contains(prompt, needle) {
			t.Fatalf("prompt missing %q", needle)
		}
	}
}

func TestBuildImplementerPromptsIncludeDemoTokenInstruction(t *testing.T) {
	t.Parallel()

	input := ImplementerPromptContext{
		MissionID:           "MISSION-101",
		Title:               "Wire prompt rendering",
		Classification:      MissionClassificationREDAlert,
		UseCases:            []string{"UC-PROMPT-02"},
		WorktreePath:        "/tmp/worktree",
		AcceptanceCriterion: "Must include AC context",
		MissionSpec:         "Implement prompt templates",
		PriorContext:        "RED failed due to missing test",
		GateFeedback:        "Gate failed: no tests",
		ValidationCommands:  []string{"go test ./..."},
	}

	builds := []struct {
		name string
		fn   func(ImplementerPromptContext) (string, error)
		must []string
	}{
		{name: "red", fn: BuildREDPrompt, must: []string{"RED phase", "Write a failing test", "demo/MISSION-MISSION-101.md"}},
		{name: "green", fn: BuildGREENPrompt, must: []string{"GREEN phase", "Gate feedback", "demo/MISSION-MISSION-101.md"}},
		{name: "refactor", fn: BuildREFACTORPrompt, must: []string{"REFACTOR phase", "Do not change externally observable behavior", "demo/MISSION-MISSION-101.md"}},
		{name: "standard", fn: BuildStandardOpsPrompt, must: []string{"STANDARD_OPS", "Validation commands", "demo/MISSION-MISSION-101.md"}},
	}

	for _, tc := range builds {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			prompt, err := tc.fn(input)
			if err != nil {
				t.Fatalf("build prompt: %v", err)
			}
			for _, needle := range tc.must {
				if !strings.Contains(prompt, needle) {
					t.Fatalf("prompt missing %q", needle)
				}
			}
		})
	}
}

func TestBuildReviewerPromptIncludesVerdictContract(t *testing.T) {
	t.Parallel()

	prompt, err := BuildReviewerPrompt(ReviewerPromptContext{
		MissionID:          "MISSION-202",
		Title:              "Review command wiring",
		Classification:     MissionClassificationStandardOps,
		AcceptanceCriteria: []string{"AC1", "AC2"},
		GateEvidence:       []string{"go test ./... passed"},
		CodeDiff:           "diff --git a/main.go b/main.go",
		DemoTokenContent:   "mission_id: MISSION-202",
	})
	if err != nil {
		t.Fatalf("build reviewer prompt: %v", err)
	}

	for _, needle := range []string{"Review command wiring", "AC1", "go test ./... passed", "decision: \"APPROVED\" | \"NEEDS_FIXES\"", "Do not rely on implementer chain-of-thought"} {
		if !strings.Contains(prompt, needle) {
			t.Fatalf("prompt missing %q", needle)
		}
	}
}

func TestBuildPromptRejectsMissingMissionID(t *testing.T) {
	t.Parallel()

	if _, err := BuildREDPrompt(ImplementerPromptContext{}); err == nil {
		t.Fatal("expected missing mission id error")
	}
	if _, err := BuildReviewerPrompt(ReviewerPromptContext{}); err == nil {
		t.Fatal("expected missing mission id error")
	}
}
