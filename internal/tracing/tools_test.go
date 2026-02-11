package tracing

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

func TestExecuteToolRecordsSpanAttributesForSuccess(t *testing.T) {
	spanRecorder := installSpanRecorder(t)
	workdir := t.TempDir()

	exitCode, stdout, stderr, err := ExecuteTool(
		context.Background(),
		"sh",
		[]string{"-c", "echo hello"},
		workdir,
	)
	if err != nil {
		t.Fatalf("execute tool: %v", err)
	}
	if exitCode != 0 {
		t.Fatalf("exit code = %d, want 0", exitCode)
	}
	if stdout != "hello" {
		t.Fatalf("stdout = %q, want hello", stdout)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}

	span := findToolExecSpan(t, spanRecorder.Ended())
	if span.Status().Code != codes.Ok {
		t.Fatalf("status code = %v, want %v", span.Status().Code, codes.Ok)
	}
	if got := getStringAttr(span.Attributes(), "tool_name"); got != "sh" {
		t.Fatalf("tool_name = %q, want sh", got)
	}
	if got := getStringAttr(span.Attributes(), "cwd"); got != workdir {
		t.Fatalf("cwd = %q, want %q", got, workdir)
	}
	if got := getIntAttr(span.Attributes(), "exit_code"); got != 0 {
		t.Fatalf("exit_code = %d, want 0", got)
	}
	if got := getIntAttr(span.Attributes(), "duration_ms"); got < 0 {
		t.Fatalf("duration_ms = %d, want >= 0", got)
	}
}

func TestExecuteToolFailureAddsBoundedStdoutStderrEvents(t *testing.T) {
	spanRecorder := installSpanRecorder(t)
	workdir := t.TempDir()

	_, _, _, err := ExecuteTool(
		context.Background(),
		"sh",
		[]string{"-c", "head -c 1600 /dev/zero | tr '\\000' 'a'; head -c 1600 /dev/zero | tr '\\000' 'b' 1>&2; exit 1"},
		workdir,
	)
	if err == nil {
		t.Fatal("expected command failure, got nil")
	}

	span := findToolExecSpan(t, spanRecorder.Ended())
	if span.Status().Code != codes.Error {
		t.Fatalf("status code = %v, want %v", span.Status().Code, codes.Error)
	}

	stdoutEvent := findEvent(t, span.Events(), "tool.stdout")
	stderrEvent := findEvent(t, span.Events(), "tool.stderr")
	stdoutValue := getStringAttr(stdoutEvent.Attributes, "output")
	stderrValue := getStringAttr(stderrEvent.Attributes, "output")

	if len(stdoutValue) > maxOutputEventBytes {
		t.Fatalf("stdout event length = %d, want <= %d", len(stdoutValue), maxOutputEventBytes)
	}
	if len(stderrValue) > maxOutputEventBytes {
		t.Fatalf("stderr event length = %d, want <= %d", len(stderrValue), maxOutputEventBytes)
	}
	if !strings.Contains(stdoutValue, "[truncated]") {
		t.Fatalf("stdout event missing truncation marker: %q", stdoutValue)
	}
	if !strings.Contains(stderrValue, "[truncated]") {
		t.Fatalf("stderr event missing truncation marker: %q", stderrValue)
	}
}

func TestExecuteToolTimeoutReturnsErrorSpan(t *testing.T) {
	spanRecorder := installSpanRecorder(t)
	workdir := t.TempDir()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	exitCode, _, _, err := ExecuteTool(ctx, "sh", []string{"-c", "sleep 1"}, workdir)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if exitCode != -1 {
		t.Fatalf("exit code = %d, want -1", exitCode)
	}

	span := findToolExecSpan(t, spanRecorder.Ended())
	if span.Status().Code != codes.Error {
		t.Fatalf("status code = %v, want %v", span.Status().Code, codes.Error)
	}
}

func TestExecuteToolGitDiffSetsOperationAndChangedFiles(t *testing.T) {
	spanRecorder := installSpanRecorder(t)
	workdir := t.TempDir()

	_, _, _, err := ExecuteTool(
		context.Background(),
		"git",
		[]string{"diff", "--name-only"},
		workdir,
	)
	if err == nil {
		t.Fatal("expected git diff error in non-repo dir, got nil")
	}

	span := findToolExecSpan(t, spanRecorder.Ended())
	if got := getStringAttr(span.Attributes(), "operation"); got != "diff" {
		t.Fatalf("operation = %q, want diff", got)
	}
	if got := getIntAttr(span.Attributes(), "changed_files"); got != 0 {
		t.Fatalf("changed_files = %d, want 0 for empty output", got)
	}
}

func installSpanRecorder(t *testing.T) *tracetest.SpanRecorder {
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

func findToolExecSpan(t *testing.T, spans []sdktrace.ReadOnlySpan) sdktrace.ReadOnlySpan {
	t.Helper()
	for _, span := range spans {
		if span.Name() == "tool.exec" {
			return span
		}
	}
	t.Fatalf("tool.exec span not found in %d spans", len(spans))
	return nil
}

func getStringAttr(attrs []attribute.KeyValue, key string) string {
	for _, attr := range attrs {
		if string(attr.Key) == key {
			return attr.Value.AsString()
		}
	}
	return ""
}

func getIntAttr(attrs []attribute.KeyValue, key string) int {
	for _, attr := range attrs {
		if string(attr.Key) == key {
			return int(attr.Value.AsInt64())
		}
	}
	return 0
}

func findEvent(t *testing.T, events []sdktrace.Event, name string) sdktrace.Event {
	t.Helper()
	for _, event := range events {
		if event.Name == name {
			return event
		}
	}
	t.Fatalf("event %q not found in %d events", name, len(events))
	return sdktrace.Event{}
}
