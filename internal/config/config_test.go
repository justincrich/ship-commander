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
`)

	writeFile(t, filepath.Join(work, ".sc3", "config.toml"), `
default_model = "project-model"
max_revisions = 7
heartbeat_interval = "45s"
gate_timeout = "3m"
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
