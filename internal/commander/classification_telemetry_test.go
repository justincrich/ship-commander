package commander

import (
	"context"
	"errors"
	"strings"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestClassifierClassifyMissionEmitsLLMCallSpan(t *testing.T) {
	recorder := installClassificationSpanRecorder(t)
	invoker := &fakeClassificationInvoker{
		response: `
mission_id: "MISSION-55"
title: "Telemetry classification"
classification: "RED_ALERT"
rationale:
  affects_behavior: true
  criteria_matched: ["bug_fix"]
  risk_assessment: "Mission touches critical behavior."
  confidence: "high"
`,
	}
	classifier, err := NewClassifier(invoker)
	if err != nil {
		t.Fatalf("new classifier: %v", err)
	}

	_, err = classifier.ClassifyMission(context.Background(), ClassificationContext{
		MissionID: "MISSION-55",
		Title:     "Telemetry classification",
		Harness:   "codex",
		Model:     "gpt-5",
	})
	if err != nil {
		t.Fatalf("classify mission: %v", err)
	}

	span := findClassificationSpanByName(t, recorder.Ended(), "llm.call")
	if span.Status().Code != codes.Ok {
		t.Fatalf("status = %v, want %v", span.Status().Code, codes.Ok)
	}
	if got := getClassificationStringAttr(span.Attributes(), "harness"); got != "codex" {
		t.Fatalf("harness = %q, want codex", got)
	}
	if got := getClassificationStringAttr(span.Attributes(), "model_name"); got != "gpt-5" {
		t.Fatalf("model_name = %q, want gpt-5", got)
	}
	if got := getClassificationIntAttr(span.Attributes(), "prompt_tokens"); got <= 0 {
		t.Fatalf("prompt_tokens = %d, want > 0", got)
	}
	if got := getClassificationIntAttr(span.Attributes(), "response_tokens"); got <= 0 {
		t.Fatalf("response_tokens = %d, want > 0", got)
	}
	if got := getClassificationIntAttr(span.Attributes(), "total_tokens"); got <= 0 {
		t.Fatalf("total_tokens = %d, want > 0", got)
	}
}

func TestClassifierClassifyMissionErrorEventRedactsSecrets(t *testing.T) {
	recorder := installClassificationSpanRecorder(t)
	invoker := &fakeClassificationInvoker{err: errors.New("token=super-secret upstream failure")}
	classifier, err := NewClassifier(invoker)
	if err != nil {
		t.Fatalf("new classifier: %v", err)
	}

	_, err = classifier.ClassifyMission(context.Background(), ClassificationContext{
		MissionID: "MISSION-88",
		Title:     "Telemetry failure",
		Harness:   "claude",
		Model:     "sonnet",
	})
	if err == nil {
		t.Fatal("expected classify error")
	}

	span := findClassificationSpanByName(t, recorder.Ended(), "llm.call")
	if span.Status().Code != codes.Error {
		t.Fatalf("status = %v, want %v", span.Status().Code, codes.Error)
	}
	errorEvent := findClassificationEventByName(t, span.Events(), "llm.error")
	if got := getClassificationStringAttr(errorEvent.Attributes, "error_type"); got != "classification_invoke_error" {
		t.Fatalf("error_type = %q, want classification_invoke_error", got)
	}
	message := getClassificationStringAttr(errorEvent.Attributes, "error_message")
	if strings.Contains(message, "super-secret") {
		t.Fatalf("error_message leaked secret: %q", message)
	}
	if !strings.Contains(message, "<redacted>") {
		t.Fatalf("error_message should contain redaction marker: %q", message)
	}
}

func installClassificationSpanRecorder(t *testing.T) *tracetest.SpanRecorder {
	t.Helper()

	recorder := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	previous := otel.GetTracerProvider()
	otel.SetTracerProvider(provider)

	t.Cleanup(func() {
		if err := provider.Shutdown(context.Background()); err != nil {
			t.Errorf("shutdown tracer provider: %v", err)
		}
		otel.SetTracerProvider(previous)
	})

	return recorder
}

func findClassificationSpanByName(t *testing.T, spans []sdktrace.ReadOnlySpan, name string) sdktrace.ReadOnlySpan {
	t.Helper()
	for _, span := range spans {
		if span.Name() == name {
			return span
		}
	}
	t.Fatalf("span %q not found in %d spans", name, len(spans))
	return nil
}

func findClassificationEventByName(t *testing.T, events []sdktrace.Event, name string) sdktrace.Event {
	t.Helper()
	for _, event := range events {
		if event.Name == name {
			return event
		}
	}
	t.Fatalf("event %q not found in %d events", name, len(events))
	return sdktrace.Event{}
}

func getClassificationStringAttr(attrs []attribute.KeyValue, key string) string {
	for _, attr := range attrs {
		if string(attr.Key) == key {
			return attr.Value.AsString()
		}
	}
	return ""
}

func getClassificationIntAttr(attrs []attribute.KeyValue, key string) int {
	for _, attr := range attrs {
		if string(attr.Key) == key {
			return int(attr.Value.AsInt64())
		}
	}
	return 0
}
