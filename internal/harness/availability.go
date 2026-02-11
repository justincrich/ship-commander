package harness

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// Availability captures which harness and runtime tools are present on PATH.
type Availability struct {
	Claude bool
	Codex  bool
	Tmux   bool
	BD     bool
}

// AvailableHarnesses returns available harness binaries in deterministic order.
func (a Availability) AvailableHarnesses() []string {
	harnesses := make([]string, 0, 2)
	if a.Claude {
		harnesses = append(harnesses, "claude")
	}
	if a.Codex {
		harnesses = append(harnesses, "codex")
	}
	return harnesses
}

// ResolveConfiguredHarness validates startup tool availability and resolves one active harness.
//
// It fails fast when required dependencies are missing:
//   - tmux must exist on PATH
//   - bd must exist on PATH
//   - at least one of claude/codex must exist on PATH
//
// When the configured harness is unavailable, the function falls back to one
// available harness and returns a warning message.
func ResolveConfiguredHarness(configured string) (string, Availability, []string, error) {
	return resolveConfiguredHarness(configured, exec.LookPath)
}

func resolveConfiguredHarness(
	configured string,
	lookPath func(file string) (string, error),
) (string, Availability, []string, error) {
	if lookPath == nil {
		return "", Availability{}, nil, errors.New("lookPath function is required")
	}

	availability := detectAvailability(lookPath)
	if err := validateAvailability(availability); err != nil {
		return "", availability, nil, err
	}

	requested := strings.ToLower(strings.TrimSpace(configured))
	fallback := preferredFallback(availability)
	if fallback == "" {
		return "", availability, nil, errors.New("no harness available for fallback")
	}

	if requested == "" {
		return fallback, availability, nil, nil
	}
	if availability.supportsHarness(requested) {
		return requested, availability, nil, nil
	}

	warnings := []string{
		fmt.Sprintf("configured harness %q unavailable; falling back to %q", requested, fallback),
	}
	return fallback, availability, warnings, nil
}

func detectAvailability(lookPath func(file string) (string, error)) Availability {
	return Availability{
		Claude: toolAvailable(lookPath, "claude"),
		Codex:  toolAvailable(lookPath, "codex"),
		Tmux:   toolAvailable(lookPath, "tmux"),
		BD:     toolAvailable(lookPath, "bd"),
	}
}

func toolAvailable(lookPath func(file string) (string, error), binary string) bool {
	_, err := lookPath(binary)
	return err == nil
}

func validateAvailability(availability Availability) error {
	if !availability.Tmux {
		return errors.New("required dependency tmux not found on PATH")
	}
	if !availability.BD {
		return errors.New("required dependency bd not found on PATH")
	}
	if len(availability.AvailableHarnesses()) == 0 {
		return errors.New("no available harness binaries found on PATH (claude/codex)")
	}
	return nil
}

func preferredFallback(availability Availability) string {
	// Keep codex preference aligned with project defaults while remaining deterministic.
	if availability.Codex {
		return "codex"
	}
	if availability.Claude {
		return "claude"
	}
	return ""
}

func (a Availability) supportsHarness(harnessName string) bool {
	switch strings.ToLower(strings.TrimSpace(harnessName)) {
	case "claude":
		return a.Claude
	case "codex":
		return a.Codex
	default:
		return false
	}
}
