package logging

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/charmbracelet/log"
)

// RuntimeLogger writes structured JSON logs to disk.
type RuntimeLogger struct {
	Logger *log.Logger
	file   *os.File
	path   string
}

// New initializes logging under ~/.sc3/logs without writing to stdout.
func New(ctx context.Context) (*RuntimeLogger, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("resolve home directory: %w", err)
	}

	logDir := filepath.Join(homeDir, ".sc3", "logs")
	if err := os.MkdirAll(logDir, 0o750); err != nil {
		return nil, fmt.Errorf("create log directory: %w", err)
	}

	filePath := filepath.Join(logDir, fmt.Sprintf("sc3-%s.log", time.Now().UTC().Format("20060102-150405")))
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

	logger.With("log_file", filePath).Info("logger initialized")

	_ = ctx
	return &RuntimeLogger{
		Logger: logger,
		file:   file,
		path:   filePath,
	}, nil
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
