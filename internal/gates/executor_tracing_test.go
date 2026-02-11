package gates

import (
	"context"
	"strings"
	"testing"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestShellExecutorRunEmitsToolExecSpan(t *testing.T) {
	spanRecorder := installExecutorSpanRecorder(t)
	workdir := t.TempDir()

	result, err := (shellExecutor{}).Run(context.Background(), workdir, "echo gate-pass", time.Second, 1024)
	if err != nil {
		t.Fatalf("run shell executor: %v", err)
	}
	if result.ExitCode != 0 {
		t.Fatalf("exit code = %d, want 0", result.ExitCode)
	}

	span := findToolExecSpanForExecutor(t, spanRecorder.Ended())
	if span.Status().Code != codes.Ok {
		t.Fatalf("status code = %v, want %v", span.Status().Code, codes.Ok)
	}
	if got := getStringAttrForExecutor(span.Attributes(), "tool_name"); got != "sh" {
		t.Fatalf("tool_name = %q, want sh", got)
	}
	if got := getStringAttrForExecutor(span.Attributes(), "cwd"); got != workdir {
		t.Fatalf("cwd = %q, want %q", got, workdir)
	}
	if got := getIntAttrForExecutor(span.Attributes(), "exit_code"); got != 0 {
		t.Fatalf("exit_code = %d, want 0", got)
	}
}

func TestShellExecutorFailureCapturesStdoutStderrEvents(t *testing.T) {
	spanRecorder := installExecutorSpanRecorder(t)
	workdir := t.TempDir()

	result, err := (shellExecutor{}).Run(
		context.Background(),
		workdir,
		"echo std-out; echo std-err 1>&2; exit 1",
		time.Second,
		1024,
	)
	if err != nil {
		t.Fatalf("run shell executor: %v", err)
	}
	if result.ExitCode != 1 {
		t.Fatalf("exit code = %d, want 1", result.ExitCode)
	}

	span := findToolExecSpanForExecutor(t, spanRecorder.Ended())
	if span.Status().Code != codes.Error {
		t.Fatalf("status code = %v, want %v", span.Status().Code, codes.Error)
	}
	stdoutEvent := findEventForExecutor(t, span.Events(), "tool.stdout")
	stderrEvent := findEventForExecutor(t, span.Events(), "tool.stderr")
	if !strings.Contains(getStringAttrForExecutor(stdoutEvent.Attributes, "output"), "std-out") {
		t.Fatalf("stdout event = %+v, expected std-out output", stdoutEvent)
	}
	if !strings.Contains(getStringAttrForExecutor(stderrEvent.Attributes, "output"), "std-err") {
		t.Fatalf("stderr event = %+v, expected std-err output", stderrEvent)
	}
}

func installExecutorSpanRecorder(t *testing.T) *tracetest.SpanRecorder {
	t.Helper()

	spanRecorder := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(spanRecorder))
	previous := otel.GetTracerProvider()
	otel.SetTracerProvider(provider)

	t.Cleanup(func() {
		if err := provider.Shutdown(context.Background()); err != nil {
			t.Errorf("shutdown tracer provider: %v", err)
		}
		otel.SetTracerProvider(previous)
	})

	return spanRecorder
}

func findToolExecSpanForExecutor(t *testing.T, spans []sdktrace.ReadOnlySpan) sdktrace.ReadOnlySpan {
	t.Helper()
	for _, span := range spans {
		if span.Name() == "tool.exec" {
			return span
		}
	}
	t.Fatalf("tool.exec span not found in %d spans", len(spans))
	return nil
}

func findEventForExecutor(t *testing.T, events []sdktrace.Event, name string) sdktrace.Event {
	t.Helper()
	for _, event := range events {
		if event.Name == name {
			return event
		}
	}
	t.Fatalf("event %q not found in %d events", name, len(events))
	return sdktrace.Event{}
}

func getStringAttrForExecutor(attrs []attribute.KeyValue, key string) string {
	for _, attr := range attrs {
		if string(attr.Key) == key {
			return attr.Value.AsString()
		}
	}
	return ""
}

func getIntAttrForExecutor(attrs []attribute.KeyValue, key string) int {
	for _, attr := range attrs {
		if string(attr.Key) == key {
			return int(attr.Value.AsInt64())
		}
	}
	return 0
}
