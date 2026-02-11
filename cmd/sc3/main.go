package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/google/uuid"
	"github.com/ship-commander/sc3/internal/config"
	"github.com/ship-commander/sc3/internal/logging"
	"github.com/ship-commander/sc3/internal/telemetry"
	"github.com/ship-commander/sc3/internal/telemetry/invariants"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Version is set at build time.
var Version = "dev"

type contextKey string

const runIDContextKey contextKey = "run_id"

type commandSpan interface {
	SetAttributes(...attribute.KeyValue)
	RecordError(error, ...trace.EventOption)
	SetStatus(codes.Code, string)
	SpanContext() trace.SpanContext
	End(...trace.SpanEndOption)
}

type traceSpanAdapter struct {
	span trace.Span
}

func (a traceSpanAdapter) SetAttributes(attrs ...attribute.KeyValue) {
	a.span.SetAttributes(attrs...)
}

func (a traceSpanAdapter) RecordError(err error, options ...trace.EventOption) {
	a.span.RecordError(err, options...)
}

func (a traceSpanAdapter) SetStatus(code codes.Code, description string) {
	a.span.SetStatus(code, description)
}

func (a traceSpanAdapter) SpanContext() trace.SpanContext {
	return a.span.SpanContext()
}

func (a traceSpanAdapter) End(options ...trace.SpanEndOption) {
	a.span.End(options...)
}

var (
	loadConfigFn       = config.Load
	newRuntimeLoggerFn = func(ctx context.Context, options ...logging.Option) (*logging.RuntimeLogger, error) {
		return logging.New(ctx, options...)
	}
	setTelemetryEndpointOverrideFn     = telemetry.SetEndpointOverride
	setTelemetryDebugConsoleExporterFn = telemetry.SetDebugConsoleExporter
	initTelemetryFn                    = telemetry.Init
	setInvariantChecksEnabledFn        = invariants.SetEnabled
	newRootCommandFn                   = newRootCommand
	startCommandSpanFn                 = func(ctx context.Context, commandName string, attrs []attribute.KeyValue) (context.Context, commandSpan) {
		spanCtx, span := otel.Tracer("sc3/command").Start(
			ctx,
			"sc3."+commandName,
			trace.WithAttributes(attrs...),
		)
		return spanCtx, traceSpanAdapter{span: span}
	}
)

func main() {
	if err := run(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	setTelemetryEndpointOverrideFn(resolveOTelEndpointFlag(args))
	defer setTelemetryEndpointOverrideFn("")
	commandName := resolveCommandName(args)
	debugEnabled := hasDebugFlag(args)
	debugConsoleExporterEnabled := debugEnabled && commandName != "tui"
	setTelemetryDebugConsoleExporterFn(debugConsoleExporterEnabled)
	defer setTelemetryDebugConsoleExporterFn(false)

	telemetry.ServiceVersion = Version
	shutdownTelemetry, err := initTelemetryFn(ctx)
	if err != nil {
		return fmt.Errorf("initialize telemetry: %w", err)
	}
	defer func() {
		if shutdownTelemetry != nil {
			shutdownTelemetry()
		}
	}()

	spanContext := ctx
	var rootSpan commandSpan
	loggerOptions := make([]logging.Option, 0, 3)
	skipInvariantChecks := hasSkipInvariantChecksFlag(args)
	if commandName != "bugreport" {
		runID := uuid.NewString()
		attrs := rootSpanAttributes(commandName, runID, args)
		spanContext = context.WithValue(spanContext, runIDContextKey, runID)
		spanContext, rootSpan = startCommandSpanFn(spanContext, commandName, attrs)
		traceID := rootSpan.SpanContext().TraceID().String()
		spanID := rootSpan.SpanContext().SpanID().String()
		rootSpan.SetAttributes(attribute.String("trace_id", traceID))
		loggerOptions = append(
			loggerOptions,
			logging.WithRunID(runID),
			logging.WithTraceID(traceID),
			logging.WithSpanID(spanID),
		)
		defer rootSpan.End()
	}

	cfg, err := loadConfigFn(spanContext)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	setInvariantChecksEnabledFn(!skipInvariantChecks)
	loggerOptions = append(
		loggerOptions,
		logging.WithMaxSizeBytes(cfg.LogMaxSizeBytes),
		logging.WithMaxFiles(cfg.LogMaxFiles),
	)
	if debugEnabled && commandName != "tui" {
		loggerOptions = append(
			loggerOptions,
			logging.WithConsoleStderr(true),
			logging.WithLevel(log.DebugLevel),
		)
	}

	logger, err := newRuntimeLoggerFn(spanContext, loggerOptions...)
	if err != nil {
		return fmt.Errorf("initialize logging: %w", err)
	}
	defer func() {
		if closeErr := logger.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "failed to close logger: %v\n", closeErr)
		}
	}()
	if debugConsoleExporterEnabled {
		logger.Logger.With("logging", "DEBUG", "otel_exporter", "console").Info("debug mode enabled")
	}

	cmd := newRootCommandFn(spanContext, cfg, logger.Logger)
	cmd.SetArgs(args)

	if err := cmd.ExecuteContext(spanContext); err != nil {
		if rootSpan != nil {
			rootSpan.RecordError(err)
			rootSpan.SetStatus(codes.Error, err.Error())
		}
		return err
	}
	if rootSpan != nil {
		rootSpan.SetStatus(codes.Ok, "command completed")
	}

	return nil
}

func newRootCommand(ctx context.Context, cfg *config.Config, logger *log.Logger) *cobra.Command {
	root := &cobra.Command{
		Use:           "sc3",
		Short:         "Ship Commander 3 orchestration runtime",
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       Version,
	}

	root.SetVersionTemplate("{{printf \"%s\\n\" .Version}}")
	root.PersistentFlags().BoolP("debug", "d", false, "Enable debug logging to stderr for non-TUI commands")
	root.PersistentFlags().String("otel-endpoint", "", "Override OTLP endpoint URL (e.g. http://localhost:4318)")
	root.PersistentFlags().Bool("skip-invariant-checks", false, "Disable invariant violation telemetry checks (emergency only)")
	root.AddCommand(
		newLeafCommand("init", "Initialize Ship Commander 3 project state", logger),
		newLeafCommand("plan", "Run Ready Room mission planning", logger),
		newLeafCommand("execute", "Execute approved missions", logger),
		newLeafCommand("tui", "Launch terminal dashboard", logger),
		newLeafCommand("status", "Show commission and mission status", logger),
		newBugreportCommand(logger),
	)

	root.PersistentPreRunE = func(cmd *cobra.Command, _ []string) error {
		if cmd.Name() == "help" || cmd.Name() == "completion" {
			return nil
		}
		if logger == nil {
			return errors.New("logger is required")
		}
		if cfg == nil {
			return errors.New("config is required")
		}
		logger.With("command", cmd.Name()).Debug("command invocation")
		return nil
	}

	_ = ctx
	return root
}

func newLeafCommand(name, short string, logger *log.Logger) *cobra.Command {
	return &cobra.Command{
		Use:   name,
		Short: short,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if logger != nil {
				logger.With("command", cmd.Name()).Info("command scaffold executed")
			}
			return nil
		},
	}
}

func resolveCommandName(args []string) string {
	for _, arg := range args {
		trimmed := strings.TrimSpace(arg)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "-") {
			continue
		}
		return trimmed
	}
	return "root"
}

func hasDebugFlag(args []string) bool {
	debugEnabled := false
	for _, arg := range args {
		trimmed := strings.TrimSpace(arg)
		switch {
		case trimmed == "--debug" || trimmed == "-d":
			debugEnabled = true
		case strings.HasPrefix(trimmed, "--debug="):
			debugEnabled = parseTruthyFlag(strings.TrimSpace(strings.TrimPrefix(trimmed, "--debug=")))
		case strings.HasPrefix(trimmed, "-d="):
			debugEnabled = parseTruthyFlag(strings.TrimSpace(strings.TrimPrefix(trimmed, "-d=")))
		}
	}
	return debugEnabled
}

func hasSkipInvariantChecksFlag(args []string) bool {
	enabled := false
	for _, arg := range args {
		trimmed := strings.TrimSpace(arg)
		switch {
		case trimmed == "--skip-invariant-checks":
			enabled = true
		case strings.HasPrefix(trimmed, "--skip-invariant-checks="):
			enabled = parseTruthyFlag(strings.TrimSpace(strings.TrimPrefix(trimmed, "--skip-invariant-checks=")))
		}
	}
	return enabled
}

func resolveOTelEndpointFlag(args []string) string {
	for i := 0; i < len(args); i++ {
		trimmed := strings.TrimSpace(args[i])
		switch {
		case strings.HasPrefix(trimmed, "--otel-endpoint="):
			return strings.TrimSpace(strings.TrimPrefix(trimmed, "--otel-endpoint="))
		case trimmed == "--otel-endpoint":
			if i+1 >= len(args) {
				return ""
			}
			return strings.TrimSpace(args[i+1])
		}
	}
	return ""
}

func parseTruthyFlag(value string) bool {
	lower := strings.ToLower(strings.TrimSpace(value))
	switch lower {
	case "", "0", "false", "no", "off":
		return false
	default:
		return true
	}
}

func rootSpanAttributes(commandName, runID string, args []string) []attribute.KeyValue {
	workingDir := ""
	if cwd, err := os.Getwd(); err == nil {
		workingDir = cwd
	}
	gitHead := readGitValue("rev-parse", "--short", "HEAD")
	gitBranch := readGitValue("rev-parse", "--abbrev-ref", "HEAD")
	environment := resolveEnvironment()

	return []attribute.KeyValue{
		attribute.String("run_id", runID),
		attribute.String("command.name", commandName),
		attribute.StringSlice("command.args", redactArgs(args)),
		attribute.String("working_dir", workingDir),
		attribute.String("git.head", gitHead),
		attribute.String("git.branch", gitBranch),
		attribute.String("environment", environment),
	}
}

func readGitValue(args ...string) string {
	cmd := exec.Command("git", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

func resolveEnvironment() string {
	for _, key := range []string{"SC3_ENV", "ENVIRONMENT", "ENV"} {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return strings.ToLower(value)
		}
	}
	return "dev"
}

func redactArgs(args []string) []string {
	redacted := make([]string, 0, len(args))
	maskNext := false

	for _, arg := range args {
		if maskNext {
			redacted = append(redacted, "<redacted>")
			maskNext = false
			continue
		}

		trimmed := strings.TrimSpace(arg)
		if strings.Contains(trimmed, "=") {
			parts := strings.SplitN(trimmed, "=", 2)
			if len(parts) == 2 && isSensitiveToken(strings.ToLower(parts[0])) {
				redacted = append(redacted, parts[0]+"=<redacted>")
				continue
			}
		}

		lower := strings.ToLower(trimmed)
		if isSensitiveToken(lower) {
			maskNext = true
			redacted = append(redacted, trimmed)
			continue
		}

		redacted = append(redacted, trimmed)
	}

	return redacted
}

func isSensitiveToken(value string) bool {
	sensitiveSubstrings := []string{
		"token",
		"password",
		"passwd",
		"secret",
		"api-key",
		"api_key",
		"apikey",
		"auth",
		"bearer",
	}
	for _, candidate := range sensitiveSubstrings {
		if strings.Contains(value, candidate) {
			return true
		}
	}
	return false
}
