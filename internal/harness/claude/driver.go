package claude

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ship-commander/sc3/internal/harness"
)

const (
	defaultModel       = "sonnet"
	defaultSessionTail = "session"
)

var (
	missionIDPattern = regexp.MustCompile(`(?i)\bmission[-_: ]*([a-z0-9-]+)\b`)
	slugPattern      = regexp.MustCompile(`[^a-z0-9]+`)
	allowedModels    = map[string]struct{}{
		"haiku":  {},
		"sonnet": {},
		"opus":   {},
	}
)

// CommandRunner executes shell commands for tmux orchestration.
type CommandRunner interface {
	Run(ctx context.Context, name string, args ...string) ([]byte, error)
}

type defaultCommandRunner struct{}

func (d defaultCommandRunner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		trimmed := strings.TrimSpace(string(out))
		if trimmed == "" {
			return nil, fmt.Errorf("run %s: %w", formatCommand(name, args), err)
		}
		return nil, fmt.Errorf("run %s: %w (%s)", formatCommand(name, args), err, trimmed)
	}
	return out, nil
}

// DriverConfig configures model-selection behavior for the Claude harness driver.
type DriverConfig struct {
	RoleModels map[string]string
}

// Driver implements harness.HarnessDriver using Claude Code in tmux sessions.
type Driver struct {
	runner     CommandRunner
	roleModels map[string]string
	now        func() time.Time

	mu          sync.Mutex
	sessionOpts map[string]harness.SessionOpts
}

// New constructs a Claude harness driver with default command execution.
func New(cfg DriverConfig) *Driver {
	return &Driver{
		runner:      defaultCommandRunner{},
		roleModels:  normalizedRoleModels(cfg.RoleModels),
		now:         time.Now,
		sessionOpts: map[string]harness.SessionOpts{},
	}
}

// NewWithRunner constructs a Claude harness driver with an injectable command runner.
func NewWithRunner(runner CommandRunner, cfg DriverConfig) (*Driver, error) {
	if runner == nil {
		return nil, errors.New("runner is required")
	}

	return &Driver{
		runner:      runner,
		roleModels:  normalizedRoleModels(cfg.RoleModels),
		now:         time.Now,
		sessionOpts: map[string]harness.SessionOpts{},
	}, nil
}

// SpawnSession starts a Claude CLI command in a detached tmux session.
func (d *Driver) SpawnSession(
	role string,
	prompt string,
	workdir string,
	opts harness.SessionOpts,
) (*harness.Session, error) {
	if d == nil {
		return nil, errors.New("driver is nil")
	}
	roleSlug := normalizeSlug(role)
	if roleSlug == "" {
		return nil, errors.New("role is required")
	}
	if strings.TrimSpace(prompt) == "" {
		return nil, errors.New("prompt is required")
	}
	workdir = strings.TrimSpace(workdir)
	if workdir == "" {
		return nil, errors.New("workdir is required")
	}

	model, err := d.resolveModel(roleSlug, opts.Model)
	if err != nil {
		return nil, err
	}
	maxTurns := opts.MaxTurns
	if maxTurns <= 0 {
		maxTurns = 1
	}

	sessionName := fmt.Sprintf("sc3-%s-%s", roleSlug, extractMissionID(prompt))
	command := buildClaudeCommand(prompt, model, maxTurns)

	ctx, cancel := d.spawnContext(opts.Timeout)
	defer cancel()
	if _, err := d.runner.Run(ctx, "tmux", "new-session", "-d", "-s", sessionName, "-c", workdir, command); err != nil {
		return nil, fmt.Errorf("create claude tmux session %s: %w", sessionName, err)
	}

	pid := 0
	if out, err := d.runner.Run(context.Background(), "tmux", "list-panes", "-t", sessionName, "-F", "#{pane_pid}"); err == nil {
		pid = parsePanePID(string(out))
	}

	session := &harness.Session{
		ID:          sessionName,
		Role:        roleSlug,
		PID:         pid,
		TmuxSession: sessionName,
		StartedAt:   d.now().UTC(),
		Status:      harness.SessionStatusRunning,
	}

	d.mu.Lock()
	d.sessionOpts[session.ID] = opts
	d.mu.Unlock()

	return session, nil
}

// SendMessage injects a message into the tmux session and captures pane output.
func (d *Driver) SendMessage(session *harness.Session, message string) (string, error) {
	if d == nil {
		return "", errors.New("driver is nil")
	}
	target, err := sessionTarget(session)
	if err != nil {
		return "", err
	}

	started := d.now()
	if _, err := d.runner.Run(context.Background(), "tmux", "send-keys", "-t", target, message, "Enter"); err != nil {
		return "", fmt.Errorf("send message to tmux session %s: %w", target, err)
	}

	out, err := d.runner.Run(context.Background(), "tmux", "capture-pane", "-pt", target, "-S", "-200")
	duration := d.now().Sub(started)
	if duration < 0 {
		duration = 0
	}

	if err != nil {
		if session != nil {
			session.LastResult = harness.SessionResult{
				ExitCode: 1,
				Stdout:   "",
				Stderr:   err.Error(),
				Duration: duration,
			}
		}
		return "", fmt.Errorf("capture tmux output for %s: %w", target, err)
	}

	stdout := strings.TrimSpace(string(out))
	if session != nil {
		session.LastResult = harness.SessionResult{
			ExitCode: 0,
			Stdout:   stdout,
			Stderr:   "",
			Duration: duration,
		}
	}

	if opts, ok := d.lookupSessionOpts(session); ok && opts.OnOutput != nil && stdout != "" {
		opts.OnOutput(stdout)
	}
	return stdout, nil
}

// Terminate ends the tmux-backed Claude session.
func (d *Driver) Terminate(session *harness.Session) error {
	if d == nil {
		return errors.New("driver is nil")
	}
	target, err := sessionTarget(session)
	if err != nil {
		return err
	}

	if _, err := d.runner.Run(context.Background(), "tmux", "kill-session", "-t", target); err != nil {
		return fmt.Errorf("kill tmux session %s: %w", target, err)
	}

	if session != nil {
		session.Status = harness.SessionStatusTerminated
	}

	d.mu.Lock()
	delete(d.sessionOpts, target)
	d.mu.Unlock()
	return nil
}

func (d *Driver) resolveModel(role string, explicitModel string) (string, error) {
	model := strings.ToLower(strings.TrimSpace(explicitModel))
	if model == "" {
		model = strings.ToLower(strings.TrimSpace(d.roleModels[role]))
	}
	if model == "" {
		model = defaultModel
	}
	if _, ok := allowedModels[model]; !ok {
		return "", fmt.Errorf("unsupported claude model %q", model)
	}
	return model, nil
}

func (d *Driver) spawnContext(timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout > 0 {
		return context.WithTimeout(context.Background(), timeout)
	}
	return context.WithCancel(context.Background())
}

func (d *Driver) lookupSessionOpts(session *harness.Session) (harness.SessionOpts, bool) {
	if session == nil {
		return harness.SessionOpts{}, false
	}
	target := strings.TrimSpace(session.TmuxSession)
	if target == "" {
		target = strings.TrimSpace(session.ID)
	}
	if target == "" {
		return harness.SessionOpts{}, false
	}

	d.mu.Lock()
	defer d.mu.Unlock()
	opts, ok := d.sessionOpts[target]
	return opts, ok
}

func buildClaudeCommand(prompt string, model string, maxTurns int) string {
	return fmt.Sprintf(
		"claude -p --model %s --verbose --max-turns %d %s",
		model,
		maxTurns,
		shellQuote(prompt),
	)
}

func shellQuote(value string) string {
	if strings.TrimSpace(value) == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}

func extractMissionID(prompt string) string {
	match := missionIDPattern.FindStringSubmatch(strings.ToLower(prompt))
	if len(match) > 1 {
		slug := normalizeSlug(match[1])
		if slug != "" {
			if strings.HasPrefix(slug, "mission-") {
				return slug
			}
			return "mission-" + slug
		}
	}
	return defaultSessionTail
}

func normalizeSlug(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return ""
	}
	value = slugPattern.ReplaceAllString(value, "-")
	value = strings.Trim(value, "-")
	return value
}

func sessionTarget(session *harness.Session) (string, error) {
	if session == nil {
		return "", errors.New("session is required")
	}
	target := strings.TrimSpace(session.TmuxSession)
	if target == "" {
		target = strings.TrimSpace(session.ID)
	}
	if target == "" {
		return "", errors.New("session target is required")
	}
	return target, nil
}

func parsePanePID(output string) int {
	firstLine := strings.TrimSpace(strings.SplitN(strings.TrimSpace(output), "\n", 2)[0])
	if firstLine == "" {
		return 0
	}
	pid, err := strconv.Atoi(firstLine)
	if err != nil || pid < 0 {
		return 0
	}
	return pid
}

func formatCommand(name string, args []string) string {
	parts := append([]string{strings.TrimSpace(name)}, args...)
	sanitized := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		sanitized = append(sanitized, part)
	}
	return strings.Join(sanitized, " ")
}

func normalizedRoleModels(input map[string]string) map[string]string {
	roleModels := make(map[string]string, len(input))
	for role, model := range input {
		roleModels[normalizeSlug(role)] = strings.ToLower(strings.TrimSpace(model))
	}
	return roleModels
}

var _ harness.HarnessDriver = (*Driver)(nil)
