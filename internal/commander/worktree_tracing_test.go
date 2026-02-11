package commander

import (
	"context"
	"os/exec"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestCommandRunnerGitOperationEmitsTracingAttributes(t *testing.T) {
	spanRecorder := installWorktreeSpanRecorder(t)
	repo := t.TempDir()

	runCommand(t, repo, "git", "init")

	_, _, err := (commandRunner{}).Run(context.Background(), repo, "git", "worktree", "list")
	if err != nil {
		t.Fatalf("run git worktree list: %v", err)
	}

	span := findToolExecSpanForWorktree(t, spanRecorder.Ended())
	if got := getStringAttrForWorktree(span.Attributes(), "tool_name"); got != "git" {
		t.Fatalf("tool_name = %q, want git", got)
	}
	if got := getStringAttrForWorktree(span.Attributes(), "operation"); got != "worktree" {
		t.Fatalf("operation = %q, want worktree", got)
	}
	if got := getIntAttrForWorktree(span.Attributes(), "changed_files"); got != 0 {
		t.Fatalf("changed_files = %d, want 0", got)
	}
}

func installWorktreeSpanRecorder(t *testing.T) *tracetest.SpanRecorder {
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

func runCommand(t *testing.T, dir string, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("run %s %v: %v (output: %s)", name, args, err, string(output))
	}
}

func findToolExecSpanForWorktree(t *testing.T, spans []sdktrace.ReadOnlySpan) sdktrace.ReadOnlySpan {
	t.Helper()
	for _, span := range spans {
		if span.Name() == "tool.exec" {
			return span
		}
	}
	t.Fatalf("tool.exec span not found in %d spans", len(spans))
	return nil
}

func getStringAttrForWorktree(attrs []attribute.KeyValue, key string) string {
	for _, attr := range attrs {
		if string(attr.Key) == key {
			return attr.Value.AsString()
		}
	}
	return ""
}

func getIntAttrForWorktree(attrs []attribute.KeyValue, key string) int {
	for _, attr := range attrs {
		if string(attr.Key) == key {
			return int(attr.Value.AsInt64())
		}
	}
	return 0
}
