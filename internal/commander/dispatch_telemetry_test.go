package commander

import (
	"context"
	"testing"
	"time"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func TestDispatchImplementerEmitsLLMCallSpan(t *testing.T) {
	recorder := installClassificationSpanRecorder(t)
	harness := &fakeHarness{}
	events := &fakeEventPublisher{}
	cmd := &Commander{
		harness: harness,
		events:  events,
		now:     time.Now,
	}

	_, err := cmd.dispatchImplementer(context.Background(), Mission{
		ID:             "m1",
		Title:          "Mission One",
		Harness:        "codex",
		Model:          "gpt-5",
		RevisionCount:  1,
		WaveFeedback:   "focus reliability",
		ReviewFeedback: "add guard clauses",
	}, t.TempDir(), 2)
	if err != nil {
		t.Fatalf("dispatch implementer: %v", err)
	}

	span := findDispatchSpanByOperation(t, recorder.Ended(), "dispatch_implementer")
	if got := getClassificationStringAttr(span.Attributes(), "operation"); got != "dispatch_implementer" {
		t.Fatalf("operation = %q, want dispatch_implementer", got)
	}
	if got := getClassificationStringAttr(span.Attributes(), "harness"); got != "codex" {
		t.Fatalf("harness = %q, want codex", got)
	}
	if got := getClassificationStringAttr(span.Attributes(), "model_name"); got != "gpt-5" {
		t.Fatalf("model_name = %q, want gpt-5", got)
	}
	if got := getClassificationIntAttr(span.Attributes(), "latency_ms"); got < 0 {
		t.Fatalf("latency_ms = %d, want >= 0", got)
	}
}

func TestDispatchReviewerAndAwaitVerdictEmitsLLMCallSpan(t *testing.T) {
	recorder := installClassificationSpanRecorder(t)
	harness := &fakeHarness{reviewerSessionIDs: []string{"rev-42"}}
	events := &fakeEventPublisher{}
	cmd := &Commander{
		harness:       harness,
		events:        events,
		now:           time.Now,
		reviewPoll:    10 * time.Millisecond,
		reviewTimeout: 50 * time.Millisecond,
	}

	verdict, err := cmd.dispatchReviewerAndAwaitVerdict(
		context.Background(),
		Mission{
			ID:                 "m2",
			Title:              "Mission Two",
			Harness:            "claude",
			Model:              "sonnet",
			AcceptanceCriteria: []string{"AC-1"},
		},
		t.TempDir(),
		1,
		"impl-21",
	)
	if err != nil {
		t.Fatalf("dispatch reviewer: %v", err)
	}
	if verdict.Decision != "APPROVED" {
		t.Fatalf("verdict = %q, want APPROVED", verdict.Decision)
	}

	span := findDispatchSpanByOperation(t, recorder.Ended(), "dispatch_reviewer")
	if got := getClassificationStringAttr(span.Attributes(), "operation"); got != "dispatch_reviewer" {
		t.Fatalf("operation = %q, want dispatch_reviewer", got)
	}
	if got := getClassificationStringAttr(span.Attributes(), "harness"); got != "claude" {
		t.Fatalf("harness = %q, want claude", got)
	}
	if got := getClassificationStringAttr(span.Attributes(), "model_name"); got != "sonnet" {
		t.Fatalf("model_name = %q, want sonnet", got)
	}
}

func findDispatchSpanByOperation(t *testing.T, spans []sdktrace.ReadOnlySpan, operation string) sdktrace.ReadOnlySpan {
	t.Helper()
	for _, span := range spans {
		if span.Name() != "llm.call" {
			continue
		}
		if got := getClassificationStringAttr(span.Attributes(), "operation"); got == operation {
			return span
		}
	}
	t.Fatalf("llm.call span with operation %q not found in %d spans", operation, len(spans))
	return nil
}
