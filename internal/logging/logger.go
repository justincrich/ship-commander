package logging

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/log"
)

// Option configures RuntimeLogger creation.
type Option func(*newOptions)

type newOptions struct {
	runID   string
	traceID string
	spanID  string
}

// WithRunID configures the run_id field used in emitted log records.
func WithRunID(runID string) Option {
	return func(opts *newOptions) {
		opts.runID = strings.TrimSpace(runID)
	}
}

// WithTraceID configures the trace_id field used in emitted log records.
func WithTraceID(traceID string) Option {
	return func(opts *newOptions) {
		opts.traceID = strings.TrimSpace(traceID)
	}
}

// WithSpanID configures the span_id field used in emitted log records.
func WithSpanID(spanID string) Option {
	return func(opts *newOptions) {
		opts.spanID = strings.TrimSpace(spanID)
	}
}

// RuntimeLogger writes structured JSON logs to disk.
type RuntimeLogger struct {
	Logger     *log.Logger
	file       *os.File
	path       string
	baseLogger *log.Logger
	runID      string
	traceID    string
	spanID     string
}

// New initializes logging under ~/.sc3/logs without writing to stdout.
func New(ctx context.Context, options ...Option) (*RuntimeLogger, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("resolve home directory: %w", err)
	}

	logDir := filepath.Join(homeDir, ".sc3", "logs")
	if err := os.MkdirAll(logDir, 0o750); err != nil {
		return nil, fmt.Errorf("create log directory: %w", err)
	}

	resolved := resolveOptions(options)
	timestamp := time.Now().UTC().Format("20060102-150405")
	fileName := fmt.Sprintf("sc3-%s.log", timestamp)
	if resolved.runID != "" {
		fileName = fmt.Sprintf("sc3-%s-%s.log", timestamp, resolved.runID)
	}
	filePath := filepath.Join(logDir, fileName)
	// #nosec G304 -- filePath is constructed from trusted local paths.
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, fmt.Errorf("open log file: %w", err)
	}

	logger := log.NewWithOptions(file, log.Options{
		Level:           log.InfoLevel,
		ReportTimestamp: true,
		TimeFormat:      time.RFC3339,
	})
	logger.SetFormatter(log.JSONFormatter)

	runtimeLogger := &RuntimeLogger{
		file:       file,
		path:       filePath,
		baseLogger: logger,
		runID:      resolved.runID,
		traceID:    resolved.traceID,
		spanID:     resolved.spanID,
	}
	runtimeLogger.rebuildLogger()
	runtimeLogger.Logger.With("log_file", filePath).Info("logger initialized")

	_ = ctx
	return runtimeLogger, nil
}

// WithRunID updates the run_id field for subsequent log records.
func (r *RuntimeLogger) WithRunID(runID string) *RuntimeLogger {
	if r == nil {
		return nil
	}
	r.runID = strings.TrimSpace(runID)
	r.rebuildLogger()
	return r
}

// WithTraceID updates the trace_id field for subsequent log records.
func (r *RuntimeLogger) WithTraceID(traceID string) *RuntimeLogger {
	if r == nil {
		return nil
	}
	r.traceID = strings.TrimSpace(traceID)
	r.rebuildLogger()
	return r
}

// WithSpanID updates the span_id field for subsequent log records.
func (r *RuntimeLogger) WithSpanID(spanID string) *RuntimeLogger {
	if r == nil {
		return nil
	}
	r.spanID = strings.TrimSpace(spanID)
	r.rebuildLogger()
	return r
}

// Close flushes and closes the log file.
func (r *RuntimeLogger) Close() error {
	if r == nil || r.file == nil {
		return nil
	}
	return r.file.Close()
}

// Path returns the current log file path.
func (r *RuntimeLogger) Path() string {
	if r == nil {
		return ""
	}
	return r.path
}

func (r *RuntimeLogger) rebuildLogger() {
	if r == nil || r.baseLogger == nil {
		return
	}
	r.Logger = r.baseLogger.With(
		"run_id", r.runID,
		"trace_id", r.traceID,
		"span_id", r.spanID,
	)
}

func resolveOptions(options []Option) newOptions {
	resolved := newOptions{}
	for _, option := range options {
		if option == nil {
			continue
		}
		option(&resolved)
	}
	return resolved
}
