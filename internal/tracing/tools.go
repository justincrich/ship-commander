package tracing

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const maxOutputEventBytes = 1024

// ExecuteTool runs a shell tool and emits deterministic tracing metadata.
func ExecuteTool(
	ctx context.Context,
	toolName string,
	args []string,
	cwd string,
) (int, string, string, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	toolName = strings.TrimSpace(toolName)
	cwd = strings.TrimSpace(cwd)
	if toolName == "" {
		return 0, "", "", errors.New("tool name must not be empty")
	}
	if cwd == "" {
		return 0, "", "", errors.New("cwd must not be empty")
	}

	redactedArgs := redactArgs(args)
	spanCtx, span := otel.Tracer("sc3/tracing/tools").Start(
		ctx,
		"tool.exec",
		trace.WithAttributes(
			attribute.String("tool_name", toolName),
			attribute.String("args_redacted", strings.Join(redactedArgs, " ")),
			attribute.String("cwd", cwd),
		),
	)
	_ = spanCtx

	started := time.Now()
	defer func() {
		span.SetAttributes(attribute.Int64("duration_ms", time.Since(started).Milliseconds()))
		span.End()
	}()

	cmd := exec.CommandContext(ctx, toolName, args...)
	cmd.Dir = cwd

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	exitCode := resolveExitCode(cmd, err, ctx)
	stdoutText := strings.TrimSpace(stdout.String())
	stderrText := strings.TrimSpace(stderr.String())

	span.SetAttributes(attribute.Int("exit_code", exitCode))
	if strings.EqualFold(toolName, "git") {
		operation := ""
		if len(args) > 0 {
			operation = strings.TrimSpace(args[0])
		}
		span.SetAttributes(
			attribute.String("operation", operation),
			attribute.Int("changed_files", estimateChangedFiles(operation, stdoutText)),
		)
	}
	if stdoutText != "" {
		span.AddEvent(
			"tool.stdout",
			trace.WithAttributes(attribute.String("output", truncateOutput(stdoutText, maxOutputEventBytes))),
		)
	}
	if stderrText != "" {
		span.AddEvent(
			"tool.stderr",
			trace.WithAttributes(attribute.String("output", truncateOutput(stderrText, maxOutputEventBytes))),
		)
	}

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return exitCode, stdoutText, stderrText, err
	}

	span.SetStatus(codes.Ok, "tool command completed")
	return exitCode, stdoutText, stderrText, nil
}

func resolveExitCode(cmd *exec.Cmd, runErr error, ctx context.Context) int {
	if runErr == nil {
		return 0
	}
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return -1
	}

	var exitErr *exec.ExitError
	if errors.As(runErr, &exitErr) {
		return exitErr.ExitCode()
	}
	if cmd != nil && cmd.ProcessState != nil {
		return cmd.ProcessState.ExitCode()
	}
	return 0
}

func estimateChangedFiles(operation, stdout string) int {
	if strings.TrimSpace(operation) != "diff" {
		return 0
	}
	count := 0
	for _, line := range strings.Split(stdout, "\n") {
		if strings.TrimSpace(line) != "" {
			count++
		}
	}
	return count
}

func truncateOutput(value string, limit int) string {
	if limit <= 0 || len(value) <= limit {
		return value
	}
	const marker = "...[truncated]"
	if limit <= len(marker) {
		return value[:limit]
	}
	return value[:limit-len(marker)] + marker
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

// FormatCommand returns a deterministic, shell-safe command preview for traces/logs.
func FormatCommand(toolName string, args []string) string {
	parts := append([]string{strings.TrimSpace(toolName)}, args...)
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		out = append(out, part)
	}
	return strings.Join(out, " ")
}

// WrapExecutionError annotates execution failures with command identity.
func WrapExecutionError(toolName string, args []string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("run %s: %w", FormatCommand(toolName, args), err)
}
