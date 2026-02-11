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

func main() {
	if err := run(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	cfg, err := config.Load(ctx)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	logger, err := logging.New(ctx)
	if err != nil {
		return fmt.Errorf("initialize logging: %w", err)
	}
	defer func() {
		if closeErr := logger.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "failed to close logger: %v\n", closeErr)
		}
	}()

	telemetry.ServiceVersion = Version
	shutdownTelemetry, err := telemetry.Init(ctx)
	if err != nil {
		return fmt.Errorf("initialize telemetry: %w", err)
	}
	defer shutdownTelemetry()

	cmd := newRootCommand(ctx, cfg, logger.Logger)
	cmd.SetArgs(args)

	spanContext := ctx
	var rootSpan trace.Span
	commandName := resolveCommandName(args)
	if commandName != "bugreport" {
		runID := uuid.NewString()
		attrs := rootSpanAttributes(commandName, runID, args)
		spanContext = context.WithValue(spanContext, runIDContextKey, runID)
		spanContext, rootSpan = otel.Tracer("sc3/command").Start(
			spanContext,
			"sc3."+commandName,
			trace.WithAttributes(attrs...),
		)
		rootSpan.SetAttributes(attribute.String("trace_id", rootSpan.SpanContext().TraceID().String()))
		defer rootSpan.End()
	}

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
	root.AddCommand(
		newLeafCommand("init", "Initialize Ship Commander 3 project state", logger),
		newLeafCommand("plan", "Run Ready Room mission planning", logger),
		newLeafCommand("execute", "Execute approved missions", logger),
		newLeafCommand("tui", "Launch terminal dashboard", logger),
		newLeafCommand("status", "Show commission and mission status", logger),
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
