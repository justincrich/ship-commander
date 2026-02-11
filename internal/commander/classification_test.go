package commander

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestBuildClassificationPromptIncludesMissionContext(t *testing.T) {
	t.Parallel()

	prompt, err := BuildClassificationPrompt(ClassificationContext{
		MissionID:              "MISSION-42",
		Title:                  "Add authentication flow",
		UseCase:                "UC-COMM-05",
		FunctionalRequirements: "Users must log in with email and password",
		DesignRequirements:     "Add sign-in form to landing view",
		CommissionTitle:        "Commander planning",
		Domain:                 "backend",
		Dependencies:           []string{"MISSION-41"},
	})
	if err != nil {
		t.Fatalf("build prompt: %v", err)
	}

	for _, needle := range []string{
		"MISSION-42",
		"Add authentication flow",
		"UC-COMM-05",
		"Users must log in with email and password",
		"Add sign-in form to landing view",
		"MISSION-41",
	} {
		if !strings.Contains(prompt, needle) {
			t.Fatalf("prompt missing %q", needle)
		}
	}
}

func TestClassifierClassifyMissionUsesConfiguredHarnessAndModel(t *testing.T) {
	t.Parallel()

	invoker := &fakeClassificationInvoker{
		response: `
mission_id: "MISSION-7"
title: "Fix mission state transition bug"
classification: "RED_ALERT"
rationale:
  affects_behavior: true
  criteria_matched: ["bug_fix"]
  risk_assessment: "Fixes incorrect mission lifecycle behavior."
  confidence: "high"
`,
	}
	classifier, err := NewClassifier(invoker)
	if err != nil {
		t.Fatalf("new classifier: %v", err)
	}

	result, err := classifier.ClassifyMission(context.Background(), ClassificationContext{
		MissionID:              "MISSION-7",
		Title:                  "Fix mission state transition bug",
		FunctionalRequirements: "Machine must reject illegal transitions",
		Harness:                "codex",
		Model:                  "gpt-5",
	})
	if err != nil {
		t.Fatalf("classify mission: %v", err)
	}

	if invoker.lastRequest.Harness != "codex" {
		t.Fatalf("harness = %q, want codex", invoker.lastRequest.Harness)
	}
	if invoker.lastRequest.Model != "gpt-5" {
		t.Fatalf("model = %q, want gpt-5", invoker.lastRequest.Model)
	}
	if result.Classification != MissionClassificationREDAlert {
		t.Fatalf("classification = %q, want %q", result.Classification, MissionClassificationREDAlert)
	}
	if result.Rationale.Confidence != "high" {
		t.Fatalf("confidence = %q, want high", result.Rationale.Confidence)
	}
}

func TestClassifierClassifyMissionReturnsLowConfidenceError(t *testing.T) {
	t.Parallel()

	invoker := &fakeClassificationInvoker{
		response: `
mission_id: "MISSION-8"
title: "Refactor planner templates"
classification: "STANDARD_OPS"
rationale:
  affects_behavior: false
  criteria_matched: ["tooling"]
  risk_assessment: "Touches only planning templates."
  confidence: "low"
`,
	}
	classifier, err := NewClassifier(invoker)
	if err != nil {
		t.Fatalf("new classifier: %v", err)
	}

	_, err = classifier.ClassifyMission(context.Background(), ClassificationContext{
		MissionID: "MISSION-8",
		Title:     "Refactor planner templates",
		Harness:   "claude",
		Model:     "sonnet",
	})
	if err == nil {
		t.Fatal("expected low confidence error")
	}
	if !errors.Is(err, ErrLowConfidenceClassification) {
		t.Fatalf("expected ErrLowConfidenceClassification, got %v", err)
	}

	var lowErr *LowConfidenceClassificationError
	if !errors.As(err, &lowErr) {
		t.Fatalf("expected LowConfidenceClassificationError, got %T", err)
	}
	if lowErr.Result.Classification != MissionClassificationStandardOps {
		t.Fatalf("classification = %q, want %q", lowErr.Result.Classification, MissionClassificationStandardOps)
	}
}

func TestClassifierClassifyMissionRejectsInvalidYAML(t *testing.T) {
	t.Parallel()

	invoker := &fakeClassificationInvoker{response: "not: [valid"}
	classifier, err := NewClassifier(invoker)
	if err != nil {
		t.Fatalf("new classifier: %v", err)
	}

	_, err = classifier.ClassifyMission(context.Background(), ClassificationContext{
		MissionID: "MISSION-9",
		Title:     "Broken response",
		Harness:   "codex",
		Model:     "gpt-5",
	})
	if err == nil {
		t.Fatal("expected parse error")
	}
	if !strings.Contains(err.Error(), "parse classification YAML") {
		t.Fatalf("unexpected error: %v", err)
	}
}

type fakeClassificationInvoker struct {
	response    string
	err         error
	lastRequest ClassificationRequest
}

func (f *fakeClassificationInvoker) ClassifyMission(_ context.Context, request ClassificationRequest) (string, error) {
	f.lastRequest = request
	if f.err != nil {
		return "", f.err
	}
	return f.response, nil
}
