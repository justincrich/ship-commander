package logging

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/log"
)

const (
	defaultMaxSizeBytes int64 = 10 * 1024 * 1024
	defaultMaxFiles           = 5
)

// Option configures RuntimeLogger creation.
type Option func(*newOptions)

type newOptions struct {
	runID           string
	traceID         string
	spanID          string
	maxSizeBytes    int64
	maxFiles        int
	consoleToStderr bool
	consoleWriter   io.Writer
	level           log.Level
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

// WithMaxSizeBytes configures max file size before rotating the active log file.
func WithMaxSizeBytes(maxSizeBytes int64) Option {
	return func(opts *newOptions) {
		opts.maxSizeBytes = maxSizeBytes
	}
}

// WithMaxFiles configures how many log files to keep, including the active file.
func WithMaxFiles(maxFiles int) Option {
	return func(opts *newOptions) {
		opts.maxFiles = maxFiles
	}
}

// WithConsoleStderr enables mirroring logs to stderr.
func WithConsoleStderr(enabled bool) Option {
	return func(opts *newOptions) {
		opts.consoleToStderr = enabled
	}
}

// WithConsoleWriter configures an alternate console sink for mirrored logs.
func WithConsoleWriter(writer io.Writer) Option {
	return func(opts *newOptions) {
		opts.consoleWriter = writer
	}
}

// WithLevel configures the runtime logger level.
func WithLevel(level log.Level) Option {
	return func(opts *newOptions) {
		opts.level = level
	}
}

// RuntimeLogger writes structured JSON logs to disk.
type RuntimeLogger struct {
	Logger     *log.Logger
	closer     io.Closer
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
	if resolved.maxSizeBytes <= 0 {
		return nil, fmt.Errorf("max log size must be > 0")
	}
	if resolved.maxFiles <= 0 {
		return nil, fmt.Errorf("max log files must be > 0")
	}
	if err := pruneHistoricalLogs(logDir, resolved.maxFiles-1); err != nil {
		return nil, fmt.Errorf("prune log directory: %w", err)
	}

	timestamp := time.Now().UTC().Format("20060102-150405")
	fileName := fmt.Sprintf("sc3-%s.log", timestamp)
	if resolved.runID != "" {
		fileName = fmt.Sprintf("sc3-%s-%s.log", timestamp, resolved.runID)
	}
	filePath := filepath.Join(logDir, fileName)

	fileWriter, err := newRotatingFileWriter(filePath, resolved.maxSizeBytes, resolved.maxFiles)
	if err != nil {
		return nil, fmt.Errorf("open log file: %w", err)
	}

	sink := io.Writer(fileWriter)
	if resolved.consoleToStderr {
		consoleSink := resolved.consoleWriter
		if consoleSink == nil {
			consoleSink = os.Stderr
		}
		if consoleSink != nil {
			sink = io.MultiWriter(fileWriter, consoleSink)
		}
	}

	logger := log.NewWithOptions(sink, log.Options{
		Level:           resolved.level,
		ReportTimestamp: true,
		TimeFormat:      time.RFC3339,
	})
	logger.SetFormatter(log.JSONFormatter)

	runtimeLogger := &RuntimeLogger{
		closer:     fileWriter,
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
	if r == nil || r.closer == nil {
		return nil
	}
	return r.closer.Close()
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
	resolved := newOptions{
		maxSizeBytes: defaultMaxSizeBytes,
		maxFiles:     defaultMaxFiles,
		level:        log.InfoLevel,
	}
	for _, option := range options {
		if option == nil {
			continue
		}
		option(&resolved)
	}
	return resolved
}

type rotatingFileWriter struct {
	path         string
	maxSizeBytes int64
	maxFiles     int
	file         *os.File
	size         int64
	mu           sync.Mutex
}

func newRotatingFileWriter(path string, maxSizeBytes int64, maxFiles int) (*rotatingFileWriter, error) {
	writer := &rotatingFileWriter{
		path:         path,
		maxSizeBytes: maxSizeBytes,
		maxFiles:     maxFiles,
	}
	if err := writer.openLocked(os.O_CREATE | os.O_APPEND | os.O_WRONLY); err != nil {
		return nil, err
	}
	return writer, nil
}

func (w *rotatingFileWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file == nil {
		if err := w.openLocked(os.O_CREATE | os.O_APPEND | os.O_WRONLY); err != nil {
			return 0, err
		}
	}
	if w.maxSizeBytes > 0 && w.size+int64(len(p)) > w.maxSizeBytes {
		if err := w.rotateLocked(); err != nil {
			return 0, err
		}
	}

	n, err := w.file.Write(p)
	w.size += int64(n)
	if err != nil {
		return n, fmt.Errorf("write log file %s: %w", w.path, err)
	}
	return n, nil
}

func (w *rotatingFileWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file == nil {
		return nil
	}
	if err := w.file.Close(); err != nil {
		return fmt.Errorf("close log file %s: %w", w.path, err)
	}
	w.file = nil
	return nil
}

func (w *rotatingFileWriter) rotateLocked() error {
	if w.file != nil {
		if err := w.file.Close(); err != nil {
			return fmt.Errorf("close log file for rotation %s: %w", w.path, err)
		}
		w.file = nil
	}

	backupLimit := w.maxFiles - 1
	if backupLimit > 0 {
		oldestBackup := fmt.Sprintf("%s.%d", w.path, backupLimit)
		if err := os.Remove(oldestBackup); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("remove oldest rotated log %s: %w", oldestBackup, err)
		}
		for idx := backupLimit - 1; idx >= 1; idx-- {
			source := fmt.Sprintf("%s.%d", w.path, idx)
			target := fmt.Sprintf("%s.%d", w.path, idx+1)
			if err := os.Rename(source, target); err != nil && !errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("rotate log %s to %s: %w", source, target, err)
			}
		}
		firstBackup := fmt.Sprintf("%s.1", w.path)
		if err := os.Rename(w.path, firstBackup); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("rotate active log %s to %s: %w", w.path, firstBackup, err)
		}
	}

	return w.openLocked(os.O_CREATE | os.O_TRUNC | os.O_WRONLY)
}

func (w *rotatingFileWriter) openLocked(flags int) error {
	// #nosec G304 -- path is constructed from trusted local paths.
	file, err := os.OpenFile(w.path, flags, 0o600)
	if err != nil {
		return fmt.Errorf("open log file %s: %w", w.path, err)
	}
	info, err := file.Stat()
	if err != nil {
		if closeErr := file.Close(); closeErr != nil {
			return fmt.Errorf("stat log file %s: %w (close: %v)", w.path, err, closeErr)
		}
		return fmt.Errorf("stat log file %s: %w", w.path, err)
	}
	w.file = file
	w.size = info.Size()
	return nil
}

type historicalLog struct {
	path    string
	modTime time.Time
}

func pruneHistoricalLogs(logDir string, keep int) error {
	if keep < 0 {
		keep = 0
	}
	entries, err := os.ReadDir(logDir)
	if err != nil {
		return err
	}

	history := make([]historicalLog, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasPrefix(entry.Name(), "sc3-") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		history = append(history, historicalLog{
			path:    filepath.Join(logDir, entry.Name()),
			modTime: info.ModTime(),
		})
	}

	sort.Slice(history, func(i, j int) bool {
		return history[i].modTime.After(history[j].modTime)
	})
	for idx := keep; idx < len(history); idx++ {
		if err := os.Remove(history[idx].path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("remove old log %s: %w", history[idx].path, err)
		}
	}
	return nil
}
