package commander

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tooltrace "github.com/ship-commander/sc3/internal/tracing"
)

type shellRunner interface {
	Run(ctx context.Context, dir string, name string, args ...string) ([]byte, []byte, error)
}

type commandRunner struct{}

func (commandRunner) Run(ctx context.Context, dir string, name string, args ...string) ([]byte, []byte, error) {
	_, stdout, stderr, err := tooltrace.ExecuteTool(ctx, name, args, dir)
	return []byte(stdout), []byte(stderr), err
}

// GitWorktreeManager creates per-mission git worktrees with deterministic naming.
type GitWorktreeManager struct {
	projectRoot string
	runner      shellRunner
}

// NewGitWorktreeManager returns a worktree manager rooted at projectRoot.
func NewGitWorktreeManager(projectRoot string) (*GitWorktreeManager, error) {
	root := strings.TrimSpace(projectRoot)
	if root == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("resolve current directory: %w", err)
		}
		root = cwd
	}

	return &GitWorktreeManager{
		projectRoot: root,
		runner:      commandRunner{},
	}, nil
}

func newGitWorktreeManagerForTest(projectRoot string, runner shellRunner) *GitWorktreeManager {
	return &GitWorktreeManager{
		projectRoot: projectRoot,
		runner:      runner,
	}
}

// Create creates a mission worktree in .beads/worktrees with feature branch naming.
func (m *GitWorktreeManager) Create(ctx context.Context, mission Mission) (string, error) {
	if m == nil {
		return "", fmt.Errorf("worktree manager is nil")
	}
	if strings.TrimSpace(mission.ID) == "" {
		return "", fmt.Errorf("mission id must not be empty")
	}
	if m.runner == nil {
		return "", fmt.Errorf("worktree runner is nil")
	}

	token := missionToken(mission.ID)
	worktreePath := filepath.Join(m.projectRoot, ".beads", "worktrees", token)
	branch := fmt.Sprintf("feature/%s-%s", token, mission.Slug())

	args := []string{"worktree", "add", worktreePath, "-b", branch}
	if _, stderr, err := m.runner.Run(ctx, m.projectRoot, "git", args...); err != nil {
		return "", fmt.Errorf("git %s: %w (stderr: %s)", strings.Join(args, " "), err, strings.TrimSpace(string(stderr)))
	}

	return worktreePath, nil
}
