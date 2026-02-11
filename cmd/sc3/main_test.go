package main

import (
	"bytes"
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/charmbracelet/log"
	"github.com/google/uuid"
	"github.com/ship-commander/sc3/internal/config"
	"github.com/ship-commander/sc3/internal/logging"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func TestRootCommandVersionFlag(t *testing.T) {
	originalVersion := Version
	defer func() {
		Version = originalVersion
	}()
	Version = "v0.1.0-test"
	cmd := newRootCommand(context.Background(), &config.Config{}, testLogger())

	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stdout)
	cmd.SetArgs([]string{"--version"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	output := strings.TrimSpace(stdout.String())
	if output != "v0.1.0-test" {
		t.Fatalf("version output = %q, want %q", output, "v0.1.0-test")
	}
}

func TestRootCommandHelpListsExpectedSubcommands(t *testing.T) {
	cmd := newRootCommand(context.Background(), &config.Config{}, testLogger())
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stdout)
	cmd.SetArgs([]string{"--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	output := stdout.String()
	expected := []string{"init", "plan", "execute", "tui", "status"}
	for _, name := range expected {
		if !strings.Contains(output, name) {
			t.Fatalf("help output missing %q: %s", name, output)
		}
	}
}

func testLogger() *log.Logger {
	return log.NewWithOptions(&bytes.Buffer{}, log.Options{})
}

func TestResolveCommandName(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "subcommand", args: []string{"plan"}, want: "plan"},
		{name: "flags then command", args: []string{"--verbose", "execute"}, want: "execute"},
		{name: "no command defaults to root", args: []string{"--help"}, want: "root"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if got := resolveCommandName(tc.args); got != tc.want {
				t.Fatalf("resolveCommandName(%v) = %q, want %q", tc.args, got, tc.want)
			}
		})
	}
}

func TestRedactArgs(t *testing.T) {
	input := []string{
		"execute",
		"--token",
		"abc123",
		"--password=supersecret",
		"--safe=value",
	}
	want := []string{
		"execute",
		"--token",
		"<redacted>",
		"--password=<redacted>",
		"--safe=value",
	}

	if got := redactArgs(input); !reflect.DeepEqual(got, want) {
		t.Fatalf("redactArgs(%v) = %v, want %v", input, got, want)
	}
}

func TestRunInitializesTelemetryBeforeConfigLoad(t *testing.T) {
	restore := snapshotRunHooks()
	defer restore()

	order := make([]string, 0, 2)
	initTelemetryFn = func(context.Context) (func(), error) {
		order = append(order, "telemetry")
		return func() {}, nil
	}
	loadConfigFn = func(context.Context) (*config.Config, error) {
		order = append(order, "config")
		return &config.Config{}, nil
	}
	newRuntimeLoggerFn = func(context.Context) (*logging.RuntimeLogger, error) {
		return &logging.RuntimeLogger{Logger: testLogger()}, nil
	}
	startCommandSpanFn = func(ctx context.Context, _ string, _ []attribute.KeyValue) (context.Context, commandSpan) {
		return ctx, newFakeCommandSpan()
	}

	if err := run(context.Background(), []string{"plan"}); err != nil {
		t.Fatalf("run: %v", err)
	}
	if len(order) < 2 {
		t.Fatalf("order length = %d, want at least 2", len(order))
	}
	if order[0] != "telemetry" || order[1] != "config" {
		t.Fatalf("order = %v, want telemetry before config", order)
	}
}

func TestRunSetsRootSpanStatusOnSuccessAndFailure(t *testing.T) {
	restore := snapshotRunHooks()
	defer restore()

	initTelemetryFn = func(context.Context) (func(), error) { return func() {}, nil }
	loadConfigFn = func(context.Context) (*config.Config, error) { return &config.Config{}, nil }
	newRuntimeLoggerFn = func(context.Context) (*logging.RuntimeLogger, error) {
		return &logging.RuntimeLogger{Logger: testLogger()}, nil
	}

	successSpan := newFakeCommandSpan()
	startCommandSpanFn = func(ctx context.Context, _ string, _ []attribute.KeyValue) (context.Context, commandSpan) {
		return ctx, successSpan
	}

	if err := run(context.Background(), []string{"plan"}); err != nil {
		t.Fatalf("run success: %v", err)
	}
	if !successSpan.hasStatus(codes.Ok) {
		t.Fatalf("expected OK status, got %v", successSpan.statusCodes)
	}

	failureSpan := newFakeCommandSpan()
	startCommandSpanFn = func(ctx context.Context, _ string, _ []attribute.KeyValue) (context.Context, commandSpan) {
		return ctx, failureSpan
	}

	err := run(context.Background(), []string{"invalid-command"})
	if err == nil {
		t.Fatal("expected error for invalid command")
	}
	if !failureSpan.hasStatus(codes.Error) {
		t.Fatalf("expected ERROR status, got %v", failureSpan.statusCodes)
	}
	if len(failureSpan.recordedErrors) == 0 {
		t.Fatal("expected error to be recorded on root span")
	}
}

func TestRunRootSpanAttributesContainRunIDAndRedactedArgs(t *testing.T) {
	restore := snapshotRunHooks()
	defer restore()

	t.Setenv("SC3_ENV", "test")

	initTelemetryFn = func(context.Context) (func(), error) { return func() {}, nil }
	loadConfigFn = func(context.Context) (*config.Config, error) { return &config.Config{}, nil }
	newRuntimeLoggerFn = func(context.Context) (*logging.RuntimeLogger, error) {
		return &logging.RuntimeLogger{Logger: testLogger()}, nil
	}
	newRootCommandFn = func(_ context.Context, _ *config.Config, _ *log.Logger) *cobra.Command {
		return &cobra.Command{
			Use:                "sc3",
			DisableFlagParsing: true,
			RunE: func(*cobra.Command, []string) error {
				return nil
			},
		}
	}

	var capturedAttrs []attribute.KeyValue
	startCommandSpanFn = func(ctx context.Context, _ string, attrs []attribute.KeyValue) (context.Context, commandSpan) {
		capturedAttrs = append([]attribute.KeyValue(nil), attrs...)
		return ctx, newFakeCommandSpan()
	}

	if err := run(context.Background(), []string{"plan", "--token", "secret", "--safe=value"}); err != nil {
		t.Fatalf("run: %v", err)
	}

	runID, ok := attrString(capturedAttrs, "run_id")
	if !ok {
		t.Fatal("run_id attribute missing")
	}
	if _, err := uuid.Parse(runID); err != nil {
		t.Fatalf("run_id is not UUID v4-compatible: %q", runID)
	}

	commandName, ok := attrString(capturedAttrs, "command.name")
	if !ok || commandName != "plan" {
		t.Fatalf("command.name = %q (present=%v), want plan", commandName, ok)
	}

	commandArgs, ok := attrStringSlice(capturedAttrs, "command.args")
	if !ok {
		t.Fatal("command.args attribute missing")
	}
	if !reflect.DeepEqual(commandArgs, []string{"plan", "--token", "<redacted>", "--safe=value"}) {
		t.Fatalf("command.args = %v", commandArgs)
	}

	if environment, ok := attrString(capturedAttrs, "environment"); !ok || environment != "test" {
		t.Fatalf("environment = %q (present=%v), want test", environment, ok)
	}
	if workingDir, ok := attrString(capturedAttrs, "working_dir"); !ok || strings.TrimSpace(workingDir) == "" {
		t.Fatalf("working_dir = %q (present=%v), expected non-empty", workingDir, ok)
	}
	if _, ok := attrString(capturedAttrs, "git.head"); !ok {
		t.Fatal("git.head attribute missing")
	}
	if _, ok := attrString(capturedAttrs, "git.branch"); !ok {
		t.Fatal("git.branch attribute missing")
	}
}

func snapshotRunHooks() func() {
	prevLoadConfig := loadConfigFn
	prevNewLogger := newRuntimeLoggerFn
	prevInitTelemetry := initTelemetryFn
	prevRootCommand := newRootCommandFn
	prevStartSpan := startCommandSpanFn

	return func() {
		loadConfigFn = prevLoadConfig
		newRuntimeLoggerFn = prevNewLogger
		initTelemetryFn = prevInitTelemetry
		newRootCommandFn = prevRootCommand
		startCommandSpanFn = prevStartSpan
	}
}

type fakeCommandSpan struct {
	statusCodes    []codes.Code
	recordedErrors []error
	spanContext    trace.SpanContext
}

func newFakeCommandSpan() *fakeCommandSpan {
	return &fakeCommandSpan{
		statusCodes: make([]codes.Code, 0),
		spanContext: trace.NewSpanContext(trace.SpanContextConfig{
			TraceID: trace.TraceID{1, 2, 3, 4},
			SpanID:  trace.SpanID{1, 2, 3, 4, 5, 6, 7, 8},
		}),
	}
}

func (f *fakeCommandSpan) SetAttributes(_ ...attribute.KeyValue) {}

func (f *fakeCommandSpan) RecordError(err error, _ ...trace.EventOption) {
	if err != nil {
		f.recordedErrors = append(f.recordedErrors, err)
	}
}

func (f *fakeCommandSpan) SetStatus(code codes.Code, _ string) {
	f.statusCodes = append(f.statusCodes, code)
}

func (f *fakeCommandSpan) SpanContext() trace.SpanContext {
	return f.spanContext
}

func (f *fakeCommandSpan) End(_ ...trace.SpanEndOption) {}

func (f *fakeCommandSpan) hasStatus(code codes.Code) bool {
	for _, candidate := range f.statusCodes {
		if candidate == code {
			return true
		}
	}
	return false
}

func attrString(attrs []attribute.KeyValue, key string) (string, bool) {
	for _, attr := range attrs {
		if string(attr.Key) != key {
			continue
		}
		return attr.Value.AsString(), true
	}
	return "", false
}

func attrStringSlice(attrs []attribute.KeyValue, key string) ([]string, bool) {
	for _, attr := range attrs {
		if string(attr.Key) != key {
			continue
		}
		return attr.Value.AsStringSlice(), true
	}
	return nil, false
}
