package main

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

const (
	bugreportLogLimit = 3
)

var (
	bugreportNowFn = func() time.Time {
		return time.Now().UTC()
	}
	bugreportHomeDirFn = os.UserHomeDir
	bugreportGetwdFn   = os.Getwd
	bugreportRunCmdFn  = func(ctx context.Context, name string, args ...string) ([]byte, error) {
		return exec.CommandContext(ctx, name, args...).CombinedOutput()
	}
)

func newBugreportCommand(logger *log.Logger) *cobra.Command {
	return &cobra.Command{
		Use:   "bugreport",
		Short: "Collect a diagnostic bundle for debugging",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if logger != nil {
				logger.With("command", "bugreport").Info("collecting diagnostic bundle")
			}
			return runBugReport(cmd.Context(), cmd.OutOrStdout())
		},
	}
}

func runBugReport(ctx context.Context, out io.Writer) error {
	homeDir, err := bugreportHomeDirFn()
	if err != nil {
		return fmt.Errorf("resolve home directory: %w", err)
	}
	homeDir = filepath.Clean(homeDir)
	if strings.TrimSpace(homeDir) == "" || homeDir == "." {
		return fmt.Errorf("home directory is not valid")
	}

	cwd, err := bugreportGetwdFn()
	if err != nil {
		return fmt.Errorf("resolve current directory: %w", err)
	}
	cwd = filepath.Clean(cwd)

	timestamp := bugreportNowFn().Format("20060102-150405")
	bundleName := fmt.Sprintf(".sc3-bugreport-%s.tar.gz", timestamp)
	bundlePath := filepath.Join(cwd, bundleName)

	stagingDir, err := os.MkdirTemp("", "sc3-bugreport-*")
	if err != nil {
		return fmt.Errorf("create staging directory: %w", err)
	}
	defer func() {
		if removeErr := os.RemoveAll(stagingDir); removeErr != nil {
			_ = removeErr
		}
	}()

	report, err := collectBugreportArtifacts(ctx, homeDir, cwd, stagingDir)
	if err != nil {
		return err
	}
	if err := writeBugreportREADME(stagingDir, report); err != nil {
		return err
	}
	if err := archiveBugreport(stagingDir, bundlePath); err != nil {
		return err
	}

	if out == nil {
		out = os.Stdout
	}
	if _, err := fmt.Fprintf(out, "Bug report written to: %s. Share for debugging.\n", bundlePath); err != nil {
		return fmt.Errorf("write bugreport output: %w", err)
	}
	return nil
}

type bugreportSummary struct {
	Timestamp string
	Version   string
	LogFiles  []string
	RunID     string
	TraceID   string
	Warnings  []string
}

func collectBugreportArtifacts(
	ctx context.Context,
	homeDir string,
	cwd string,
	stagingDir string,
) (bugreportSummary, error) {
	summary := bugreportSummary{
		Timestamp: bugreportNowFn().Format(time.RFC3339),
		Version:   Version,
		Warnings:  make([]string, 0),
	}

	logFiles, warnings := copyRecentLogs(homeDir, stagingDir, bugreportLogLimit)
	summary.LogFiles = logFiles
	summary.Warnings = append(summary.Warnings, warnings...)

	runID, traceID := extractLastCorrelation(logFiles)
	summary.RunID = runID
	summary.TraceID = traceID
	if runID == "" && traceID == "" {
		summary.Warnings = append(summary.Warnings, "no run_id/trace_id found in copied logs")
	}

	if err := writeLastRunFile(stagingDir, runID, traceID); err != nil {
		return bugreportSummary{}, err
	}
	if err := writeVersionFile(stagingDir, summary.Version); err != nil {
		return bugreportSummary{}, err
	}
	if err := copyRedactedConfig(homeDir, stagingDir, &summary); err != nil {
		return bugreportSummary{}, err
	}
	if err := writeGitState(ctx, cwd, stagingDir); err != nil {
		return bugreportSummary{}, err
	}
	if err := copyLatestTestOutput(homeDir, stagingDir, &summary); err != nil {
		return bugreportSummary{}, err
	}

	return summary, nil
}

func copyRecentLogs(homeDir string, stagingDir string, limit int) ([]string, []string) {
	logsDir := filepath.Join(homeDir, ".sc3", "logs")
	files, err := newestFiles(logsDir, limit)
	if err != nil {
		return nil, []string{fmt.Sprintf("unable to read logs directory: %v", err)}
	}

	destDir := filepath.Join(stagingDir, "logs")
	if err := os.MkdirAll(destDir, 0o750); err != nil {
		return nil, []string{fmt.Sprintf("unable to create logs staging directory: %v", err)}
	}

	warnings := make([]string, 0)
	copiedPaths := make([]string, 0, len(files))
	for _, file := range files {
		// #nosec G304 -- source path comes from deterministic ~/.sc3/logs enumeration.
		data, readErr := os.ReadFile(file.path)
		if readErr != nil {
			warnings = append(warnings, fmt.Sprintf("unable to read log %s: %v", file.path, readErr))
			continue
		}
		dstPath := filepath.Join(destDir, filepath.Base(file.path))
		if writeErr := os.WriteFile(dstPath, data, 0o600); writeErr != nil {
			warnings = append(warnings, fmt.Sprintf("unable to stage log %s: %v", file.path, writeErr))
			continue
		}
		copiedPaths = append(copiedPaths, file.path)
	}
	return copiedPaths, warnings
}

func extractLastCorrelation(logPaths []string) (string, string) {
	for _, logPath := range logPaths {
		// #nosec G304 -- log paths are selected from deterministic ~/.sc3/logs files.
		data, err := os.ReadFile(logPath)
		if err != nil {
			continue
		}
		lines := strings.Split(strings.TrimSpace(string(data)), "\n")
		for i := len(lines) - 1; i >= 0; i-- {
			line := strings.TrimSpace(lines[i])
			if line == "" {
				continue
			}
			record := map[string]any{}
			if err := json.Unmarshal([]byte(line), &record); err != nil {
				continue
			}
			runID := asString(record["run_id"])
			traceID := asString(record["trace_id"])
			if runID == "" && traceID == "" {
				continue
			}
			return runID, traceID
		}
	}
	return "", ""
}

func writeLastRunFile(stagingDir, runID, traceID string) error {
	content := strings.TrimSpace(fmt.Sprintf("run_id: %s\ntrace_id: %s\n", runID, traceID))
	path := filepath.Join(stagingDir, "last-run.txt")
	if err := os.WriteFile(path, []byte(content+"\n"), 0o600); err != nil {
		return fmt.Errorf("write last-run.txt: %w", err)
	}
	return nil
}

func writeVersionFile(stagingDir, version string) error {
	content := fmt.Sprintf("sc3 version: %s\n", strings.TrimSpace(version))
	path := filepath.Join(stagingDir, "version.txt")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		return fmt.Errorf("write version.txt: %w", err)
	}
	return nil
}

func copyRedactedConfig(homeDir, stagingDir string, summary *bugreportSummary) error {
	configPath := filepath.Join(homeDir, ".sc3", "config.yaml")
	// #nosec G304 -- config path is deterministic under ~/.sc3.
	configData, err := os.ReadFile(configPath)
	if err != nil {
		summary.Warnings = append(summary.Warnings, fmt.Sprintf("unable to read config: %v", err))
		configData = []byte("# config unavailable\n")
	}
	redacted := redactSensitiveConfig(string(configData))
	if err := os.WriteFile(filepath.Join(stagingDir, "config.yaml"), []byte(redacted), 0o600); err != nil {
		return fmt.Errorf("write redacted config: %w", err)
	}
	return nil
}

func redactSensitiveConfig(configText string) string {
	lines := strings.Split(configText, "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		separator := ":"
		if strings.Contains(line, "=") && (!strings.Contains(line, ":") || strings.Index(line, "=") < strings.Index(line, ":")) {
			separator = "="
		}
		parts := strings.SplitN(line, separator, 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		if !isSensitiveToken(strings.ToLower(key)) {
			continue
		}
		prefix := parts[0] + separator
		lines[i] = prefix + " ***REDACTED***"
	}
	return strings.Join(lines, "\n")
}

func writeGitState(ctx context.Context, cwd, stagingDir string) error {
	head := runCommandForBugreport(ctx, "git", "-C", cwd, "rev-parse", "HEAD")
	branch := runCommandForBugreport(ctx, "git", "-C", cwd, "rev-parse", "--abbrev-ref", "HEAD")
	status := runCommandForBugreport(ctx, "git", "-C", cwd, "status", "--short")
	diff := runCommandForBugreport(ctx, "git", "-C", cwd, "diff")

	content := strings.Join([]string{
		"[HEAD]",
		head,
		"",
		"[BRANCH]",
		branch,
		"",
		"[STATUS]",
		status,
		"",
		"[DIFF]",
		diff,
		"",
	}, "\n")
	if err := os.WriteFile(filepath.Join(stagingDir, "git-state.txt"), []byte(content), 0o600); err != nil {
		return fmt.Errorf("write git-state.txt: %w", err)
	}
	return nil
}

func runCommandForBugreport(ctx context.Context, name string, args ...string) string {
	output, err := bugreportRunCmdFn(ctx, name, args...)
	text := strings.TrimSpace(string(output))
	if err == nil {
		return text
	}
	if text == "" {
		return fmt.Sprintf("error: %v", err)
	}
	return text + "\nerror: " + err.Error()
}

func copyLatestTestOutput(homeDir, stagingDir string, summary *bugreportSummary) error {
	testOutputPath := filepath.Join(homeDir, ".sc3", "last-test-output.txt")
	// #nosec G304 -- test output path is deterministic under ~/.sc3.
	testOutput, err := os.ReadFile(testOutputPath)
	if err != nil {
		summary.Warnings = append(summary.Warnings, "no failing test output found")
		testOutput = []byte("No failing test output found.\n")
	}
	if err := os.WriteFile(filepath.Join(stagingDir, "test-output.txt"), testOutput, 0o600); err != nil {
		return fmt.Errorf("write test-output.txt: %w", err)
	}
	return nil
}

func writeBugreportREADME(stagingDir string, summary bugreportSummary) error {
	builder := strings.Builder{}
	builder.WriteString("Ship Commander 3 Bug Report\n")
	builder.WriteString("===========================\n\n")
	builder.WriteString(fmt.Sprintf("Generated: %s\n", summary.Timestamp))
	builder.WriteString(fmt.Sprintf("Version: %s\n", summary.Version))
	builder.WriteString(fmt.Sprintf("run_id: %s\n", summary.RunID))
	builder.WriteString(fmt.Sprintf("trace_id: %s\n\n", summary.TraceID))
	builder.WriteString("Included artifacts:\n")
	builder.WriteString("- logs/ (up to last 3 log files)\n")
	builder.WriteString("- config.yaml (redacted)\n")
	builder.WriteString("- version.txt\n")
	builder.WriteString("- last-run.txt\n")
	builder.WriteString("- git-state.txt\n")
	builder.WriteString("- test-output.txt\n\n")
	builder.WriteString("Usage:\n")
	builder.WriteString("- Share this archive with maintainers for debugging.\n")
	builder.WriteString("- Use run_id/trace_id to correlate logs with traces.\n")
	if len(summary.Warnings) > 0 {
		builder.WriteString("\nWarnings:\n")
		for _, warning := range summary.Warnings {
			builder.WriteString("- " + warning + "\n")
		}
	}

	if err := os.WriteFile(filepath.Join(stagingDir, "README.txt"), []byte(builder.String()), 0o600); err != nil {
		return fmt.Errorf("write README.txt: %w", err)
	}
	return nil
}

func archiveBugreport(stagingDir, destination string) error {
	// #nosec G304 -- destination is generated in current working directory with deterministic file name.
	archiveFile, err := os.OpenFile(destination, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("create archive %s: %w", destination, err)
	}
	defer func() {
		if closeErr := archiveFile.Close(); closeErr != nil {
			_ = closeErr
		}
	}()

	gzipWriter := gzip.NewWriter(archiveFile)
	defer func() {
		if closeErr := gzipWriter.Close(); closeErr != nil {
			_ = closeErr
		}
	}()

	tarWriter := tar.NewWriter(gzipWriter)
	defer func() {
		if closeErr := tarWriter.Close(); closeErr != nil {
			_ = closeErr
		}
	}()

	walkErr := filepath.WalkDir(stagingDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return fmt.Errorf("read file info for %s: %w", path, err)
		}

		relPath, err := filepath.Rel(stagingDir, path)
		if err != nil {
			return fmt.Errorf("compute archive path for %s: %w", path, err)
		}

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return fmt.Errorf("create tar header for %s: %w", path, err)
		}
		header.Name = filepath.ToSlash(relPath)

		if err := tarWriter.WriteHeader(header); err != nil {
			return fmt.Errorf("write tar header for %s: %w", path, err)
		}

		// #nosec G304 -- walk paths originate from controlled staging directory.
		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("open %s for archive: %w", path, err)
		}
		if _, err := io.Copy(tarWriter, file); err != nil {
			if closeErr := file.Close(); closeErr != nil {
				return fmt.Errorf("close %s after copy failure: %w", path, closeErr)
			}
			return fmt.Errorf("copy %s into archive: %w", path, err)
		}
		if err := file.Close(); err != nil {
			return fmt.Errorf("close %s: %w", path, err)
		}
		return nil
	})
	if walkErr != nil {
		return fmt.Errorf("archive bugreport: %w", walkErr)
	}

	return nil
}

type datedFile struct {
	path    string
	modTime time.Time
}

func newestFiles(dir string, limit int) ([]datedFile, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	files := make([]datedFile, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		files = append(files, datedFile{
			path:    filepath.Join(dir, entry.Name()),
			modTime: info.ModTime(),
		})
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].modTime.After(files[j].modTime)
	})
	if limit > 0 && len(files) > limit {
		files = files[:limit]
	}
	return files, nil
}

func asString(value any) string {
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	default:
		return ""
	}
}
