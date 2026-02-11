package telemetry

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestStartLLMCallAndEndRecordsCoreAttributes(t *testing.T) {
	recorder := installLLMSpanRecorder(t)

	ctx, llmCall := StartLLMCall(context.Background(), LLMCallRequest{
		Operation: "mission_classification",
		ModelName: "gpt-5",
		Harness:   "codex",
		Prompt:    "classify mission with token=super-secret",
	})
	if llmCall == nil {
		t.Fatal("expected llm call tracker")
	}
	if LLMCallFromContext(ctx) == nil {
		t.Fatal("expected llm call tracker in context")
	}

	llmCall.RecordToolCall("git", 25*time.Millisecond, true)
	llmCall.End("classification: RED_ALERT", nil, nil)

	span := findSpanByName(t, recorder.Ended(), "llm.call")
	if span.Status().Code != codes.Ok {
		t.Fatalf("status = %v, want %v", span.Status().Code, codes.Ok)
	}
	if got := getStringAttrByKey(span.Attributes(), "model_name"); got != "gpt-5" {
		t.Fatalf("model_name = %q, want gpt-5", got)
	}
	if got := getStringAttrByKey(span.Attributes(), "harness"); got != "codex" {
		t.Fatalf("harness = %q, want codex", got)
	}
	if got := getStringAttrByKey(span.Attributes(), "operation"); got != "mission_classification" {
		t.Fatalf("operation = %q, want mission_classification", got)
	}
	if got := getIntAttrByKey(span.Attributes(), "prompt_tokens"); got <= 0 {
		t.Fatalf("prompt_tokens = %d, want > 0", got)
	}
	if got := getIntAttrByKey(span.Attributes(), "response_tokens"); got <= 0 {
		t.Fatalf("response_tokens = %d, want > 0", got)
	}
	if got := getIntAttrByKey(span.Attributes(), "total_tokens"); got <= 0 {
		t.Fatalf("total_tokens = %d, want > 0", got)
	}
	if got := getIntAttrByKey(span.Attributes(), "tool_calls_count"); got != 1 {
		t.Fatalf("tool_calls_count = %d, want 1", got)
	}
	if got := getIntAttrByKey(span.Attributes(), "latency_ms"); got < 0 {
		t.Fatalf("latency_ms = %d, want >= 0", got)
	}

	hashValue := getStringAttrByKey(span.Attributes(), "prompt_hash")
	if len(hashValue) != 64 {
		t.Fatalf("prompt_hash length = %d, want 64", len(hashValue))
	}
	if strings.Contains(hashValue, "super-secret") {
		t.Fatalf("prompt hash unexpectedly contains secret: %q", hashValue)
	}

	toolEvent := findEventByName(t, span.Events(), "llm.tool_call")
	if got := getStringAttrByKey(toolEvent.Attributes, "tool_name"); got != "git" {
		t.Fatalf("tool event tool_name = %q, want git", got)
	}
	if got := getIntAttrByKey(toolEvent.Attributes, "duration_ms"); got != 25 {
		t.Fatalf("tool event duration_ms = %d, want 25", got)
	}
}

func TestLLMCallRecordErrorRedactsSecrets(t *testing.T) {
	recorder := installLLMSpanRecorder(t)

	_, llmCall := StartLLMCall(context.Background(), LLMCallRequest{
		ModelName: "sonnet",
		Harness:   "claude",
		Prompt:    "token=another-secret",
	})
	llmCall.RecordError("invoke_failure", "api_key=my-key token=top-secret", 2)
	llmCall.End("", nil, errors.New("authorization=bearer-private"))

	span := findSpanByName(t, recorder.Ended(), "llm.call")
	if span.Status().Code != codes.Error {
		t.Fatalf("status = %v, want %v", span.Status().Code, codes.Error)
	}

	errorEvent := findEventByName(t, span.Events(), "llm.error")
	if got := getStringAttrByKey(errorEvent.Attributes, "error_type"); got != "invoke_failure" {
		t.Fatalf("error_type = %q, want invoke_failure", got)
	}
	if got := getIntAttrByKey(errorEvent.Attributes, "retry_count"); got != 2 {
		t.Fatalf("retry_count = %d, want 2", got)
	}
	message := getStringAttrByKey(errorEvent.Attributes, "error_message")
	if strings.Contains(message, "my-key") || strings.Contains(message, "top-secret") {
		t.Fatalf("error message leaked secret: %q", message)
	}
	if !strings.Contains(message, "<redacted>") {
		t.Fatalf("expected redaction marker in error message, got %q", message)
	}
}

func installLLMSpanRecorder(t *testing.T) *tracetest.SpanRecorder {
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

func findSpanByName(t *testing.T, spans []sdktrace.ReadOnlySpan, name string) sdktrace.ReadOnlySpan {
	t.Helper()
	for _, span := range spans {
		if span.Name() == name {
			return span
		}
	}
	t.Fatalf("span %q not found in %d spans", name, len(spans))
	return nil
}

func findEventByName(t *testing.T, events []sdktrace.Event, name string) sdktrace.Event {
	t.Helper()
	for _, event := range events {
		if event.Name == name {
			return event
		}
	}
	t.Fatalf("event %q not found in %d events", name, len(events))
	return sdktrace.Event{}
}

func getStringAttrByKey(attrs []attribute.KeyValue, key string) string {
	for _, attr := range attrs {
		if string(attr.Key) == key {
			return attr.Value.AsString()
		}
	}
	return ""
}

func getIntAttrByKey(attrs []attribute.KeyValue, key string) int {
	for _, attr := range attrs {
		if string(attr.Key) == key {
			return int(attr.Value.AsInt64())
		}
	}
	return 0
}
