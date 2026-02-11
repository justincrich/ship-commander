package commission

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// CommandRunner abstracts command execution for persistence tests.
type CommandRunner interface {
	Run(ctx context.Context, name string, args ...string) ([]byte, error)
}

type defaultCommandRunner struct{}

func (defaultCommandRunner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("run %s %s: %w; output: %s", name, strings.Join(args, " "), err, strings.TrimSpace(string(output)))
	}
	return output, nil
}

// Persist stores a commission in Beads by creating a new issue with `bd create`.
func Persist(ctx context.Context, commission *Commission) (string, error) {
	return PersistWithRunner(ctx, commission, defaultCommandRunner{})
}

// PersistWithRunner stores a commission in Beads using a custom command runner.
func PersistWithRunner(ctx context.Context, commission *Commission, runner CommandRunner) (string, error) {
	if commission == nil {
		return "", fmt.Errorf("commission must not be nil")
	}
	if runner == nil {
		return "", fmt.Errorf("runner must not be nil")
	}
	if strings.TrimSpace(commission.Title) == "" {
		return "", fmt.Errorf("commission title must not be empty")
	}

	payload, err := json.Marshal(commission)
	if err != nil {
		return "", fmt.Errorf("marshal commission: %w", err)
	}

	args := []string{
		"create",
		"--title", commission.Title,
		"--type", "feature",
		"--description", string(payload),
		"--silent",
	}

	out, err := runner.Run(ctx, "bd", args...)
	if err != nil {
		return "", fmt.Errorf("persist commission with bd create: %w", err)
	}

	id := strings.TrimSpace(string(out))
	if id == "" {
		return "", fmt.Errorf("bd create returned empty issue id")
	}

	commission.ID = id
	return id, nil
}
