package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/charmbracelet/log"
	"github.com/google/uuid"
	"github.com/ship-commander/sc3/internal/config"
	"github.com/ship-commander/sc3/internal/harness"
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
	expected := []string{"init", "plan", "execute", "tui", "status", "bugreport"}
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

func TestHasDebugFlag(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want bool
	}{
		{name: "long flag", args: []string{"--debug", "plan"}, want: true},
		{name: "short flag", args: []string{"-d", "plan"}, want: true},
		{name: "explicit false", args: []string{"--debug=false", "plan"}, want: false},
		{name: "unset", args: []string{"plan"}, want: false},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if got := hasDebugFlag(tc.args); got != tc.want {
				t.Fatalf("hasDebugFlag(%v) = %v, want %v", tc.args, got, tc.want)
			}
		})
	}
}

func TestHasSkipInvariantChecksFlag(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want bool
	}{
		{name: "long flag", args: []string{"--skip-invariant-checks", "plan"}, want: true},
		{name: "explicit false", args: []string{"--skip-invariant-checks=false", "plan"}, want: false},
		{name: "unset", args: []string{"plan"}, want: false},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if got := hasSkipInvariantChecksFlag(tc.args); got != tc.want {
				t.Fatalf("hasSkipInvariantChecksFlag(%v) = %v, want %v", tc.args, got, tc.want)
			}
		})
	}
}

func TestResolveOTelEndpointFlag(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "inline value", args: []string{"--otel-endpoint=https://otel.example.com:4318", "plan"}, want: "https://otel.example.com:4318"},
		{name: "separate value", args: []string{"--otel-endpoint", "http://localhost:4318", "plan"}, want: "http://localhost:4318"},
		{name: "unset", args: []string{"plan"}, want: ""},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if got := resolveOTelEndpointFlag(tc.args); got != tc.want {
				t.Fatalf("resolveOTelEndpointFlag(%v) = %q, want %q", tc.args, got, tc.want)
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
		return testRuntimeConfig(), nil
	}
	newRuntimeLoggerFn = func(context.Context, ...logging.Option) (*logging.RuntimeLogger, error) {
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
	loadConfigFn = func(context.Context) (*config.Config, error) { return testRuntimeConfig(), nil }
	newRuntimeLoggerFn = func(context.Context, ...logging.Option) (*logging.RuntimeLogger, error) {
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
	loadConfigFn = func(context.Context) (*config.Config, error) { return testRuntimeConfig(), nil }
	newRuntimeLoggerFn = func(context.Context, ...logging.Option) (*logging.RuntimeLogger, error) {
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

func TestRunWritesCorrelatedLogFieldsFromRootSpan(t *testing.T) {
	restore := snapshotRunHooks()
	defer restore()

	home := t.TempDir()
	t.Setenv("HOME", home)

	initTelemetryFn = func(context.Context) (func(), error) { return func() {}, nil }
	loadConfigFn = func(context.Context) (*config.Config, error) { return testRuntimeConfig(), nil }
	newRuntimeLoggerFn = func(ctx context.Context, options ...logging.Option) (*logging.RuntimeLogger, error) {
		return logging.New(ctx, options...)
	}
	newRootCommandFn = func(_ context.Context, _ *config.Config, logger *log.Logger) *cobra.Command {
		return &cobra.Command{
			Use:                "sc3",
			DisableFlagParsing: true,
			RunE: func(*cobra.Command, []string) error {
				logger.Info("integration-correlation-entry")
				return nil
			},
		}
	}

	var capturedAttrs []attribute.KeyValue
	startCommandSpanFn = func(ctx context.Context, _ string, attrs []attribute.KeyValue) (context.Context, commandSpan) {
		capturedAttrs = append([]attribute.KeyValue(nil), attrs...)
		return ctx, newFakeCommandSpan()
	}

	if err := run(context.Background(), []string{"plan"}); err != nil {
		t.Fatalf("run: %v", err)
	}

	runID, ok := attrString(capturedAttrs, "run_id")
	if !ok || strings.TrimSpace(runID) == "" {
		t.Fatalf("captured run_id missing: %q", runID)
	}

	logFiles, err := filepath.Glob(filepath.Join(home, ".sc3", "logs", "sc3-*.log"))
	if err != nil {
		t.Fatalf("glob log files: %v", err)
	}
	if len(logFiles) == 0 {
		t.Fatal("expected at least one log file")
	}

	content, err := os.ReadFile(logFiles[0])
	if err != nil {
		t.Fatalf("read log file: %v", err)
	}

	var correlatedRecord map[string]any
	for _, line := range strings.Split(strings.TrimSpace(string(content)), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		record := map[string]any{}
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			t.Fatalf("unmarshal log line: %v", err)
		}
		if asStringAny(record["msg"]) == "integration-correlation-entry" {
			correlatedRecord = record
			break
		}
	}
	if correlatedRecord == nil {
		t.Fatal("did not find integration-correlation-entry record in log output")
	}

	if got := asStringAny(correlatedRecord["run_id"]); got != runID {
		t.Fatalf("run_id = %q, want %q", got, runID)
	}
	if got := asStringAny(correlatedRecord["trace_id"]); got != newFakeCommandSpan().spanContext.TraceID().String() {
		t.Fatalf("trace_id = %q, want %q", got, newFakeCommandSpan().spanContext.TraceID().String())
	}
	if got := asStringAny(correlatedRecord["span_id"]); got != newFakeCommandSpan().spanContext.SpanID().String() {
		t.Fatalf("span_id = %q, want %q", got, newFakeCommandSpan().spanContext.SpanID().String())
	}
}

func TestRunDebugMirrorsLogsToStderrForNonTUICommands(t *testing.T) {
	restore := snapshotRunHooks()
	defer restore()

	home := t.TempDir()
	t.Setenv("HOME", home)

	initTelemetryFn = func(context.Context) (func(), error) { return func() {}, nil }
	loadConfigFn = func(context.Context) (*config.Config, error) { return testRuntimeConfig(), nil }
	newRuntimeLoggerFn = func(ctx context.Context, options ...logging.Option) (*logging.RuntimeLogger, error) {
		return logging.New(ctx, options...)
	}
	startCommandSpanFn = func(ctx context.Context, _ string, _ []attribute.KeyValue) (context.Context, commandSpan) {
		return ctx, newFakeCommandSpan()
	}

	stderr := captureRunStderr(t, func() error {
		return run(context.Background(), []string{"--debug", "plan"})
	})
	if !strings.Contains(stderr, "command scaffold executed") {
		t.Fatalf("expected debug mode console output on stderr, got: %q", stderr)
	}
	if !strings.Contains(stderr, "\"logging\":\"DEBUG\"") {
		t.Fatalf("expected debug logging marker in stderr output, got: %q", stderr)
	}
	if !strings.Contains(stderr, "\"otel_exporter\":\"console\"") {
		t.Fatalf("expected console exporter marker in stderr output, got: %q", stderr)
	}
}

func TestRunDebugDoesNotMirrorLogsForTUICommand(t *testing.T) {
	restore := snapshotRunHooks()
	defer restore()

	home := t.TempDir()
	t.Setenv("HOME", home)

	initTelemetryFn = func(context.Context) (func(), error) { return func() {}, nil }
	loadConfigFn = func(context.Context) (*config.Config, error) { return testRuntimeConfig(), nil }
	newRuntimeLoggerFn = func(ctx context.Context, options ...logging.Option) (*logging.RuntimeLogger, error) {
		return logging.New(ctx, options...)
	}
	startCommandSpanFn = func(ctx context.Context, _ string, _ []attribute.KeyValue) (context.Context, commandSpan) {
		return ctx, newFakeCommandSpan()
	}

	stderr := captureRunStderr(t, func() error {
		return run(context.Background(), []string{"--debug", "tui"})
	})
	if strings.Contains(stderr, "command scaffold executed") {
		t.Fatalf("expected no mirrored stderr logs for tui command, got: %q", stderr)
	}
}

func TestRunSetsInvariantChecksFromFlags(t *testing.T) {
	restore := snapshotRunHooks()
	defer restore()

	initTelemetryFn = func(context.Context) (func(), error) { return func() {}, nil }
	loadConfigFn = func(context.Context) (*config.Config, error) { return testRuntimeConfig(), nil }
	newRuntimeLoggerFn = func(context.Context, ...logging.Option) (*logging.RuntimeLogger, error) {
		return &logging.RuntimeLogger{Logger: testLogger()}, nil
	}
	startCommandSpanFn = func(ctx context.Context, _ string, _ []attribute.KeyValue) (context.Context, commandSpan) {
		return ctx, newFakeCommandSpan()
	}

	enabledStates := make([]bool, 0, 2)
	setInvariantChecksEnabledFn = func(enabled bool) {
		enabledStates = append(enabledStates, enabled)
	}

	if err := run(context.Background(), []string{"plan"}); err != nil {
		t.Fatalf("run default invariants enabled: %v", err)
	}
	if err := run(context.Background(), []string{"--skip-invariant-checks", "plan"}); err != nil {
		t.Fatalf("run skip invariants: %v", err)
	}

	if !reflect.DeepEqual(enabledStates, []bool{true, false}) {
		t.Fatalf("invariant enabled states = %v, want [true false]", enabledStates)
	}
}

func TestRunSetsTelemetryEndpointOverrideFromFlags(t *testing.T) {
	restore := snapshotRunHooks()
	defer restore()

	initTelemetryFn = func(context.Context) (func(), error) { return func() {}, nil }
	loadConfigFn = func(context.Context) (*config.Config, error) { return testRuntimeConfig(), nil }
	newRuntimeLoggerFn = func(context.Context, ...logging.Option) (*logging.RuntimeLogger, error) {
		return &logging.RuntimeLogger{Logger: testLogger()}, nil
	}
	startCommandSpanFn = func(ctx context.Context, _ string, _ []attribute.KeyValue) (context.Context, commandSpan) {
		return ctx, newFakeCommandSpan()
	}

	values := make([]string, 0, 2)
	setTelemetryEndpointOverrideFn = func(endpoint string) {
		values = append(values, endpoint)
	}

	if err := run(context.Background(), []string{"--otel-endpoint=https://collector.example.com:4318", "plan"}); err != nil {
		t.Fatalf("run with otel endpoint: %v", err)
	}

	if len(values) < 2 {
		t.Fatalf("expected override setter to be called at least twice (set + reset), got %d", len(values))
	}
	if values[0] != "https://collector.example.com:4318" {
		t.Fatalf("first endpoint override = %q, want https://collector.example.com:4318", values[0])
	}
	if values[len(values)-1] != "" {
		t.Fatalf("final endpoint override reset = %q, want empty string", values[len(values)-1])
	}
}

func TestRunSetsTelemetryDebugConsoleExporterFromFlags(t *testing.T) {
	restore := snapshotRunHooks()
	defer restore()

	initTelemetryFn = func(context.Context) (func(), error) { return func() {}, nil }
	loadConfigFn = func(context.Context) (*config.Config, error) { return testRuntimeConfig(), nil }
	newRuntimeLoggerFn = func(context.Context, ...logging.Option) (*logging.RuntimeLogger, error) {
		return &logging.RuntimeLogger{Logger: testLogger()}, nil
	}
	startCommandSpanFn = func(ctx context.Context, _ string, _ []attribute.KeyValue) (context.Context, commandSpan) {
		return ctx, newFakeCommandSpan()
	}

	values := make([]bool, 0, 4)
	setTelemetryDebugConsoleExporterFn = func(enabled bool) {
		values = append(values, enabled)
	}

	if err := run(context.Background(), []string{"--debug", "plan"}); err != nil {
		t.Fatalf("run with --debug plan: %v", err)
	}
	if err := run(context.Background(), []string{"--debug", "tui"}); err != nil {
		t.Fatalf("run with --debug tui: %v", err)
	}

	if len(values) < 4 {
		t.Fatalf("expected at least four setter calls (set+reset for two runs), got %d", len(values))
	}
	if !values[0] {
		t.Fatalf("first setter call = %v, want true for non-tui debug", values[0])
	}
	if values[1] {
		t.Fatalf("second setter call = %v, want false reset", values[1])
	}
	if values[2] {
		t.Fatalf("third setter call = %v, want false for tui debug", values[2])
	}
	if values[3] {
		t.Fatalf("fourth setter call = %v, want false reset", values[3])
	}
}

func TestRunFailsWhenHarnessAvailabilityCheckFails(t *testing.T) {
	restore := snapshotRunHooks()
	defer restore()

	initTelemetryFn = func(context.Context) (func(), error) { return func() {}, nil }
	loadConfigFn = func(context.Context) (*config.Config, error) { return testRuntimeConfig(), nil }
	newRuntimeLoggerFn = func(context.Context, ...logging.Option) (*logging.RuntimeLogger, error) {
		return &logging.RuntimeLogger{Logger: testLogger()}, nil
	}
	startCommandSpanFn = func(ctx context.Context, _ string, _ []attribute.KeyValue) (context.Context, commandSpan) {
		return ctx, newFakeCommandSpan()
	}
	resolveHarnessAvailabilityFn = func(string) (string, harness.Availability, []string, error) {
		return "", harness.Availability{}, nil, errors.New("tmux unavailable")
	}

	err := run(context.Background(), []string{"plan"})
	if err == nil {
		t.Fatal("expected harness availability error")
	}
	if !strings.Contains(err.Error(), "check harness availability") {
		t.Fatalf("error = %v, want check harness availability context", err)
	}
}

func TestRunAppliesHarnessFallbackToConfig(t *testing.T) {
	restore := snapshotRunHooks()
	defer restore()

	initTelemetryFn = func(context.Context) (func(), error) { return func() {}, nil }
	loadConfigFn = func(context.Context) (*config.Config, error) {
		cfg := testRuntimeConfig()
		cfg.DefaultHarness = "claude"
		return cfg, nil
	}
	newRuntimeLoggerFn = func(context.Context, ...logging.Option) (*logging.RuntimeLogger, error) {
		return &logging.RuntimeLogger{Logger: testLogger()}, nil
	}
	startCommandSpanFn = func(ctx context.Context, _ string, _ []attribute.KeyValue) (context.Context, commandSpan) {
		return ctx, newFakeCommandSpan()
	}
	resolveHarnessAvailabilityFn = func(string) (string, harness.Availability, []string, error) {
		return "codex", harness.Availability{Codex: true, Tmux: true, BD: true}, []string{"fallback warning"}, nil
	}

	capturedHarness := ""
	newRootCommandFn = func(_ context.Context, cfg *config.Config, _ *log.Logger) *cobra.Command {
		capturedHarness = cfg.DefaultHarness
		return &cobra.Command{
			Use:                "sc3",
			DisableFlagParsing: true,
			RunE: func(*cobra.Command, []string) error {
				return nil
			},
		}
	}

	if err := run(context.Background(), []string{"plan"}); err != nil {
		t.Fatalf("run: %v", err)
	}
	if capturedHarness != "codex" {
		t.Fatalf("captured default harness = %q, want %q", capturedHarness, "codex")
	}
}

func snapshotRunHooks() func() {
	prevLoadConfig := loadConfigFn
	prevNewLogger := newRuntimeLoggerFn
	prevSetTelemetryEndpointOverride := setTelemetryEndpointOverrideFn
	prevSetTelemetryDebugConsoleExporter := setTelemetryDebugConsoleExporterFn
	prevInitTelemetry := initTelemetryFn
	prevSetInvariantChecks := setInvariantChecksEnabledFn
	prevResolveHarnessAvailability := resolveHarnessAvailabilityFn
	prevRootCommand := newRootCommandFn
	prevStartSpan := startCommandSpanFn

	resolveHarnessAvailabilityFn = func(configured string) (string, harness.Availability, []string, error) {
		candidate := strings.TrimSpace(configured)
		if candidate == "" {
			candidate = "codex"
		}
		return candidate, harness.Availability{Claude: true, Codex: true, Tmux: true, BD: true}, nil, nil
	}

	return func() {
		loadConfigFn = prevLoadConfig
		newRuntimeLoggerFn = prevNewLogger
		setTelemetryEndpointOverrideFn = prevSetTelemetryEndpointOverride
		setTelemetryDebugConsoleExporterFn = prevSetTelemetryDebugConsoleExporter
		initTelemetryFn = prevInitTelemetry
		setInvariantChecksEnabledFn = prevSetInvariantChecks
		resolveHarnessAvailabilityFn = prevResolveHarnessAvailability
		newRootCommandFn = prevRootCommand
		startCommandSpanFn = prevStartSpan
	}
}

func captureRunStderr(t *testing.T, runFn func() error) string {
	t.Helper()

	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe stderr: %v", err)
	}
	originalStderr := os.Stderr
	os.Stderr = writer
	t.Cleanup(func() {
		os.Stderr = originalStderr
	})

	runErr := runFn()
	if closeErr := writer.Close(); closeErr != nil {
		t.Fatalf("close stderr writer: %v", closeErr)
	}
	data, readErr := io.ReadAll(reader)
	if readErr != nil {
		t.Fatalf("read stderr: %v", readErr)
	}
	if closeErr := reader.Close(); closeErr != nil {
		t.Fatalf("close stderr reader: %v", closeErr)
	}
	if runErr != nil {
		t.Fatalf("run: %v", runErr)
	}
	return string(data)
}

func testRuntimeConfig() *config.Config {
	return &config.Config{
		DefaultHarness:  "codex",
		LogMaxSizeBytes: 10 * 1024 * 1024,
		LogMaxFiles:     5,
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

func asStringAny(value any) string {
	text, ok := value.(string)
	if !ok {
		return ""
	}
	return text
}
