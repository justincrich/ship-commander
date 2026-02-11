package harness

import (
	"errors"
	"strings"
	"testing"
)

func TestResolveConfiguredHarnessUsesConfiguredWhenAvailable(t *testing.T) {
	t.Parallel()

	resolved, availability, warnings, err := resolveConfiguredHarness(
		"claude",
		fakeLookPath(map[string]bool{
			"claude": true,
			"codex":  true,
			"tmux":   true,
			"bd":     true,
		}),
	)
	if err != nil {
		t.Fatalf("resolve configured harness: %v", err)
	}
	if resolved != "claude" {
		t.Fatalf("resolved harness = %q, want %q", resolved, "claude")
	}
	if !availability.Claude || !availability.Codex || !availability.Tmux || !availability.BD {
		t.Fatalf("unexpected availability flags: %#v", availability)
	}
	if len(warnings) != 0 {
		t.Fatalf("warnings = %v, want none", warnings)
	}
}

func TestResolveConfiguredHarnessFallsBackWithWarning(t *testing.T) {
	t.Parallel()

	resolved, _, warnings, err := resolveConfiguredHarness(
		"claude",
		fakeLookPath(map[string]bool{
			"codex": true,
			"tmux":  true,
			"bd":    true,
		}),
	)
	if err != nil {
		t.Fatalf("resolve configured harness: %v", err)
	}
	if resolved != "codex" {
		t.Fatalf("resolved harness = %q, want %q", resolved, "codex")
	}
	if len(warnings) != 1 {
		t.Fatalf("warning count = %d, want 1", len(warnings))
	}
	if !strings.Contains(warnings[0], `configured harness "claude" unavailable`) {
		t.Fatalf("warning = %q", warnings[0])
	}
}

func TestResolveConfiguredHarnessFailsWhenRequiredToolMissing(t *testing.T) {
	t.Parallel()

	_, _, _, err := resolveConfiguredHarness(
		"codex",
		fakeLookPath(map[string]bool{
			"codex": true,
			"bd":    true,
		}),
	)
	if err == nil {
		t.Fatal("expected missing tmux error")
	}
	if !strings.Contains(err.Error(), "tmux") {
		t.Fatalf("error = %v, want tmux dependency message", err)
	}
}

func TestResolveConfiguredHarnessFailsWhenNoHarnessAvailable(t *testing.T) {
	t.Parallel()

	_, _, _, err := resolveConfiguredHarness(
		"codex",
		fakeLookPath(map[string]bool{
			"tmux": true,
			"bd":   true,
		}),
	)
	if err == nil {
		t.Fatal("expected no harness available error")
	}
	if !strings.Contains(err.Error(), "no available harness") {
		t.Fatalf("error = %v, want no harness message", err)
	}
}

func TestResolveConfiguredHarnessRejectsNilLookPath(t *testing.T) {
	t.Parallel()

	_, _, _, err := resolveConfiguredHarness("codex", nil)
	if err == nil {
		t.Fatal("expected nil lookPath error")
	}
}

func fakeLookPath(available map[string]bool) func(file string) (string, error) {
	return func(file string) (string, error) {
		if available[file] {
			return "/usr/bin/" + file, nil
		}
		return "", errors.New("not found")
	}
}
