package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRunBugReportCreatesArchiveWithRedactedConfigAndArtifacts(t *testing.T) {
	restore := snapshotBugreportHooks()
	defer restore()

	fixture := setupBugreportFixture(t)

	var out bytes.Buffer
	if err := runBugReport(context.Background(), &out); err != nil {
		t.Fatalf("run bugreport: %v", err)
	}
	output := strings.TrimSpace(out.String())
	if !strings.Contains(output, "Bug report written to:") {
		t.Fatalf("unexpected output: %q", output)
	}

	archivePath := filepath.Join(fixture.cwd, ".sc3-bugreport-20260211-100000.tar.gz")
	contents := extractTarballTextFiles(t, archivePath)

	assertBugreportCoreArtifacts(t, contents)

	logCount := 0
	for name := range contents {
		if strings.HasPrefix(name, "logs/") {
			logCount++
		}
	}
	if logCount != 3 {
		t.Fatalf("log file count = %d, want 3 most recent logs", logCount)
	}
	if strings.Contains(contents["config.yaml"], "supersecret") || strings.Contains(contents["config.yaml"], "pass123") {
		t.Fatalf("config should be redacted: %q", contents["config.yaml"])
	}
	if !strings.Contains(contents["config.yaml"], "***REDACTED***") {
		t.Fatalf("config redaction marker missing: %q", contents["config.yaml"])
	}
	if !strings.Contains(contents["last-run.txt"], "run-123") || !strings.Contains(contents["last-run.txt"], "trace-abc") {
		t.Fatalf("missing run/trace IDs: %q", contents["last-run.txt"])
	}
}

func TestRunBugReportHandlesMissingOptionalArtifacts(t *testing.T) {
	restore := snapshotBugreportHooks()
	defer restore()

	home := filepath.Join(t.TempDir(), "home")
	cwd := filepath.Join(t.TempDir(), "cwd")
	if err := os.MkdirAll(home, 0o750); err != nil {
		t.Fatalf("create home: %v", err)
	}
	if err := os.MkdirAll(cwd, 0o750); err != nil {
		t.Fatalf("create cwd: %v", err)
	}

	bugreportHomeDirFn = func() (string, error) { return home, nil }
	bugreportGetwdFn = func() (string, error) { return cwd, nil }
	bugreportNowFn = func() time.Time { return time.Date(2026, 2, 11, 11, 0, 0, 0, time.UTC) }
	bugreportRunCmdFn = func(context.Context, string, ...string) ([]byte, error) { return []byte(""), nil }

	var out bytes.Buffer
	if err := runBugReport(context.Background(), &out); err != nil {
		t.Fatalf("run bugreport: %v", err)
	}

	archivePath := filepath.Join(cwd, ".sc3-bugreport-20260211-110000.tar.gz")
	contents := extractTarballTextFiles(t, archivePath)
	readme := contents["README.txt"]
	if !strings.Contains(readme, "unable to read logs directory") {
		t.Fatalf("readme should include missing logs warning: %q", readme)
	}
	if !strings.Contains(readme, "no failing test output found") {
		t.Fatalf("readme should include missing test output warning: %q", readme)
	}
	if !strings.Contains(contents["config.yaml"], "config unavailable") {
		t.Fatalf("expected config placeholder, got: %q", contents["config.yaml"])
	}
	if !strings.Contains(contents["test-output.txt"], "No failing test output found.") {
		t.Fatalf("expected test output placeholder, got: %q", contents["test-output.txt"])
	}
}

func snapshotBugreportHooks() func() {
	prevNow := bugreportNowFn
	prevHomeDir := bugreportHomeDirFn
	prevGetwd := bugreportGetwdFn
	prevRunCmd := bugreportRunCmdFn
	return func() {
		bugreportNowFn = prevNow
		bugreportHomeDirFn = prevHomeDir
		bugreportGetwdFn = prevGetwd
		bugreportRunCmdFn = prevRunCmd
	}
}

func extractTarballTextFiles(t *testing.T, archivePath string) map[string]string {
	t.Helper()

	// #nosec G304 -- archivePath is generated in the test-owned temp directory.
	archiveFile, err := os.Open(archivePath)
	if err != nil {
		t.Fatalf("open archive: %v", err)
	}
	defer func() {
		if closeErr := archiveFile.Close(); closeErr != nil {
			t.Fatalf("close archive file: %v", closeErr)
		}
	}()

	gzipReader, err := gzip.NewReader(archiveFile)
	if err != nil {
		t.Fatalf("create gzip reader: %v", err)
	}
	defer func() {
		if closeErr := gzipReader.Close(); closeErr != nil {
			t.Fatalf("close gzip reader: %v", closeErr)
		}
	}()

	tarReader := tar.NewReader(gzipReader)
	files := make(map[string]string)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("read tar entry: %v", err)
		}
		data, err := io.ReadAll(tarReader)
		if err != nil {
			t.Fatalf("read tar entry %s: %v", header.Name, err)
		}
		files[header.Name] = string(data)
	}
	if len(files) == 0 {
		t.Fatalf("archive %s is empty", archivePath)
	}
	return files
}

type bugreportFixture struct {
	home string
	cwd  string
}

func setupBugreportFixture(t *testing.T) bugreportFixture {
	t.Helper()

	home := filepath.Join(t.TempDir(), "home")
	cwd := filepath.Join(t.TempDir(), "cwd")
	if err := os.MkdirAll(filepath.Join(home, ".sc3", "logs"), 0o750); err != nil {
		t.Fatalf("create logs dir: %v", err)
	}
	if err := os.MkdirAll(cwd, 0o750); err != nil {
		t.Fatalf("create cwd: %v", err)
	}

	baseTime := time.Date(2026, 2, 11, 9, 0, 0, 0, time.UTC)
	writeBugreportLog(t, home, "log-1.log", `{"msg":"older"}`, baseTime.Add(-4*time.Minute))
	writeBugreportLog(t, home, "log-2.log", `{"msg":"middle"}`, baseTime.Add(-3*time.Minute))
	writeBugreportLog(
		t,
		home,
		"log-3.log",
		`{"msg":"newer","run_id":"run-123","trace_id":"trace-abc"}`,
		baseTime.Add(-2*time.Minute),
	)
	writeBugreportLog(t, home, "log-4.log", `{"msg":"newest"}`, baseTime.Add(-1*time.Minute))

	configText := "api_key: supersecret\npassword=pass123\nnormal: value\n"
	if err := os.WriteFile(filepath.Join(home, ".sc3", "config.yaml"), []byte(configText), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(home, ".sc3", "last-test-output.txt"), []byte("failing test output"), 0o600); err != nil {
		t.Fatalf("write test output: %v", err)
	}

	bugreportHomeDirFn = func() (string, error) { return home, nil }
	bugreportGetwdFn = func() (string, error) { return cwd, nil }
	bugreportNowFn = func() time.Time { return time.Date(2026, 2, 11, 10, 0, 0, 0, time.UTC) }
	bugreportRunCmdFn = stubBugreportGitCommands

	return bugreportFixture{home: home, cwd: cwd}
}

func writeBugreportLog(t *testing.T, home, name, content string, modTime time.Time) {
	t.Helper()

	path := filepath.Join(home, ".sc3", "logs", name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write log %s: %v", name, err)
	}
	if err := os.Chtimes(path, modTime, modTime); err != nil {
		t.Fatalf("chtimes %s: %v", name, err)
	}
}

func stubBugreportGitCommands(_ context.Context, name string, args ...string) ([]byte, error) {
	joined := name + " " + strings.Join(args, " ")
	switch {
	case strings.Contains(joined, "rev-parse HEAD"):
		return []byte("deadbeef\n"), nil
	case strings.Contains(joined, "rev-parse --abbrev-ref HEAD"):
		return []byte("main\n"), nil
	case strings.Contains(joined, "status --short"):
		return []byte(" M internal/commander/commander.go\n"), nil
	case strings.Contains(joined, "git -C") && strings.Contains(joined, " diff"):
		return []byte("diff --git a/file b/file\n"), nil
	default:
		return []byte(""), nil
	}
}

func assertBugreportCoreArtifacts(t *testing.T, contents map[string]string) {
	t.Helper()

	required := []string{
		"README.txt",
		"config.yaml",
		"version.txt",
		"last-run.txt",
		"git-state.txt",
		"test-output.txt",
	}
	for _, path := range required {
		if _, ok := contents[path]; !ok {
			t.Fatalf("missing artifact %q in bugreport archive", path)
		}
	}
}

func TestRedactSensitiveConfig(t *testing.T) {
	input := "api_key: abc\npassword=def\nnormal: value\n"
	got := redactSensitiveConfig(input)
	if strings.Contains(got, "abc") || strings.Contains(got, "def") {
		t.Fatalf("expected sensitive values to be redacted: %q", got)
	}
	if strings.Count(got, "***REDACTED***") != 2 {
		t.Fatalf("expected two redactions, got %q", got)
	}
}

func TestNewestFiles(t *testing.T) {
	dir := t.TempDir()
	base := time.Date(2026, 2, 11, 12, 0, 0, 0, time.UTC)
	for i := 1; i <= 4; i++ {
		path := filepath.Join(dir, fmt.Sprintf("log-%d.log", i))
		if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
			t.Fatalf("write file %d: %v", i, err)
		}
		mod := base.Add(time.Duration(i) * time.Minute)
		if err := os.Chtimes(path, mod, mod); err != nil {
			t.Fatalf("set modtime %d: %v", i, err)
		}
	}

	files, err := newestFiles(dir, 2)
	if err != nil {
		t.Fatalf("newestFiles: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("file count = %d, want 2", len(files))
	}
	if !strings.HasSuffix(files[0].path, "log-4.log") {
		t.Fatalf("first file = %s, want log-4.log", files[0].path)
	}
	if !strings.HasSuffix(files[1].path, "log-3.log") {
		t.Fatalf("second file = %s, want log-3.log", files[1].path)
	}
}
