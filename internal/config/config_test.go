package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	home := t.TempDir()
	work := t.TempDir()
	t.Setenv("HOME", home)

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() {
		if chdirErr := os.Chdir(cwd); chdirErr != nil {
			t.Fatalf("restore cwd: %v", chdirErr)
		}
	})
	if err := os.Chdir(work); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	cfg, err := Load(context.Background())
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.DefaultHarness != defaultHarness {
		t.Fatalf("default_harness = %q, want %q", cfg.DefaultHarness, defaultHarness)
	}
	if cfg.DefaultModel != defaultModel {
		t.Fatalf("default_model = %q, want %q", cfg.DefaultModel, defaultModel)
	}
	if cfg.WIPLimit != defaultWIPLimit {
		t.Fatalf("wip_limit = %d, want %d", cfg.WIPLimit, defaultWIPLimit)
	}
	if cfg.MaxRevisions != defaultMaxRevisions {
		t.Fatalf("max_revisions = %d, want %d", cfg.MaxRevisions, defaultMaxRevisions)
	}
	if cfg.PlanningMaxIterations != defaultPlanningIterations {
		t.Fatalf("planning_max_iterations = %d, want %d", cfg.PlanningMaxIterations, defaultPlanningIterations)
	}
	if cfg.StuckTimeout != defaultStuckTimeout {
		t.Fatalf("stuck_timeout = %s, want %s", cfg.StuckTimeout, defaultStuckTimeout)
	}
	if cfg.HeartbeatInterval != defaultHeartbeatInterval {
		t.Fatalf("heartbeat_interval = %s, want %s", cfg.HeartbeatInterval, defaultHeartbeatInterval)
	}
	if cfg.GateTimeout != defaultGateTimeout {
		t.Fatalf("gate_timeout = %s, want %s", cfg.GateTimeout, defaultGateTimeout)
	}
	if cfg.LogMaxSizeBytes != defaultLogMaxSizeBytes {
		t.Fatalf("log_max_size_bytes = %d, want %d", cfg.LogMaxSizeBytes, defaultLogMaxSizeBytes)
	}
	if cfg.LogMaxFiles != defaultLogMaxFiles {
		t.Fatalf("log_max_files = %d, want %d", cfg.LogMaxFiles, defaultLogMaxFiles)
	}
}

func TestLoadOverlayProjectOverHome(t *testing.T) {
	home := t.TempDir()
	work := t.TempDir()
	t.Setenv("HOME", home)

	writeFile(t, filepath.Join(home, ".sc3", "config.toml"), `
default_harness = "home-harness"
default_model = "home-model"
wip_limit = 9
stuck_timeout = "9m"
log_max_size_mb = 20
	`)

	writeFile(t, filepath.Join(work, ".sc3", "config.toml"), `
default_model = "project-model"
max_revisions = 7
heartbeat_interval = "45s"
gate_timeout = "3m"
log_max_files = 7
	`)

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() {
		if chdirErr := os.Chdir(cwd); chdirErr != nil {
			t.Fatalf("restore cwd: %v", chdirErr)
		}
	})
	if err := os.Chdir(work); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	cfg, err := Load(context.Background())
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.DefaultHarness != "home-harness" {
		t.Fatalf("default_harness = %q, want %q", cfg.DefaultHarness, "home-harness")
	}
	if cfg.DefaultModel != "project-model" {
		t.Fatalf("default_model = %q, want %q", cfg.DefaultModel, "project-model")
	}
	if cfg.WIPLimit != 9 {
		t.Fatalf("wip_limit = %d, want 9", cfg.WIPLimit)
	}
	if cfg.MaxRevisions != 7 {
		t.Fatalf("max_revisions = %d, want 7", cfg.MaxRevisions)
	}
	if cfg.StuckTimeout != 9*time.Minute {
		t.Fatalf("stuck_timeout = %s, want 9m", cfg.StuckTimeout)
	}
	if cfg.HeartbeatInterval != 45*time.Second {
		t.Fatalf("heartbeat_interval = %s, want 45s", cfg.HeartbeatInterval)
	}
	if cfg.GateTimeout != 3*time.Minute {
		t.Fatalf("gate_timeout = %s, want 3m", cfg.GateTimeout)
	}
	if cfg.LogMaxSizeBytes != 20*1024*1024 {
		t.Fatalf("log_max_size_bytes = %d, want %d", cfg.LogMaxSizeBytes, 20*1024*1024)
	}
	if cfg.LogMaxFiles != 7 {
		t.Fatalf("log_max_files = %d, want 7", cfg.LogMaxFiles)
	}
}

func TestLoadRoleAndDomainHarnessModelConfig(t *testing.T) {
	home := t.TempDir()
	work := t.TempDir()
	t.Setenv("HOME", home)

	writeFile(t, filepath.Join(work, ".sc3", "config.toml"), `
[defaults]
harness = "claude"
model = "sonnet"

[roles.captain]
harness = "claude"
model = "opus"

[roles.ensign]
harness = "claude"
model = "haiku"

[roles.ensign.backend]
model = "sonnet"
`)

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() {
		if chdirErr := os.Chdir(cwd); chdirErr != nil {
			t.Fatalf("restore cwd: %v", chdirErr)
		}
	})
	if err := os.Chdir(work); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	cfg, err := Load(context.Background())
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.DefaultHarness != "claude" {
		t.Fatalf("default harness = %q, want claude", cfg.DefaultHarness)
	}
	if cfg.DefaultModel != "sonnet" {
		t.Fatalf("default model = %q, want sonnet", cfg.DefaultModel)
	}
	captain := cfg.Roles["captain"]
	if captain.Harness != "claude" || captain.Model != "opus" {
		t.Fatalf("captain role config = %#v", captain)
	}
	backend := cfg.Roles["ensign"].Domains["backend"]
	if backend.Harness != "claude" || backend.Model != "sonnet" {
		t.Fatalf("ensign backend config = %#v", backend)
	}
}

func TestResolveHarnessModelPriorityAndFallback(t *testing.T) {
	cfg := defaults()
	cfg.DefaultHarness = "codex"
	cfg.DefaultModel = "gpt-5-codex"
	cfg.Roles["ensign"] = RoleHarnessConfig{
		Harness: "claude",
		Model:   "haiku",
		Domains: map[string]HarnessModelConfig{
			"backend": {
				Model: "sonnet",
			},
		},
	}

	harnessName, modelName, warnings, err := cfg.ResolveHarnessModel(
		"ensign",
		"backend",
		map[string]bool{"claude": true, "codex": true},
	)
	if err != nil {
		t.Fatalf("resolve role/domain config: %v", err)
	}
	if harnessName != "claude" {
		t.Fatalf("resolved harness = %q, want claude", harnessName)
	}
	if modelName != "sonnet" {
		t.Fatalf("resolved model = %q, want sonnet (domain override)", modelName)
	}
	if len(warnings) != 0 {
		t.Fatalf("warnings = %v, want none", warnings)
	}

	harnessName, modelName, warnings, err = cfg.ResolveHarnessModel(
		"ensign",
		"backend",
		map[string]bool{"claude": false, "codex": true},
	)
	if err != nil {
		t.Fatalf("resolve fallback config: %v", err)
	}
	if harnessName != "codex" {
		t.Fatalf("fallback harness = %q, want codex", harnessName)
	}
	if modelName != "sonnet" {
		t.Fatalf("fallback model = %q, want sonnet", modelName)
	}
	if len(warnings) != 1 {
		t.Fatalf("warning count = %d, want 1", len(warnings))
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}
}
