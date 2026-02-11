package tmux

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/ship-commander/sc3/internal/events"
)

const (
	defaultCaptureStartLine = "-2000"

	// DefaultOutputChunkLimitBytes caps emitted output chunk size at 1MB.
	DefaultOutputChunkLimitBytes = 1 << 20
	// DefaultOutputPollInterval is the default interval for capture-pane streaming.
	DefaultOutputPollInterval = 2 * time.Second
	// DefaultTerminationGracePeriod is the SIGTERM grace window before SIGKILL.
	DefaultTerminationGracePeriod = 5 * time.Second

	defaultTerminationPollInterval = 100 * time.Millisecond
	defaultForcedExitWait          = 2 * time.Second
)

const (
	// EventTypeSessionOutputChunk identifies streamed tmux capture-pane output chunks.
	EventTypeSessionOutputChunk = "SessionOutputChunk"
)

var sessionNamePattern = regexp.MustCompile(`^sc3-[a-z0-9][a-z0-9-]*-[a-z0-9][a-z0-9-]*$`)

// CommandRunner executes tmux commands.
type CommandRunner interface {
	Run(ctx context.Context, name string, args ...string) ([]byte, error)
}

// ProcessSignaler sends unix signals to a process ID.
type ProcessSignaler interface {
	Signal(pid int, signal syscall.Signal) error
}

// ProcessChecker checks whether a process is still alive.
type ProcessChecker interface {
	Alive(pid int) (bool, error)
}

type defaultCommandRunner struct{}

func (defaultCommandRunner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
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

type defaultProcessSignaler struct{}

func (defaultProcessSignaler) Signal(pid int, signal syscall.Signal) error {
	return syscall.Kill(pid, signal)
}

type defaultProcessChecker struct{}

func (defaultProcessChecker) Alive(pid int) (bool, error) {
	err := syscall.Kill(pid, 0)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, syscall.ESRCH) {
		return false, nil
	}
	if errors.Is(err, syscall.EPERM) {
		return true, nil
	}
	return false, err
}

// TmuxSession is one active tmux session descriptor.
//
//nolint:revive // Name is fixed by issue acceptance criteria.
type TmuxSession struct {
	Name string
}

// OutputChunk is a streamed capture-pane event payload.
type OutputChunk struct {
	SessionName string
	Chunk       string
	Truncated   bool
	CapturedAt  time.Time
}

// Options configures a tmux manager.
type Options struct {
	Runner                  CommandRunner
	Signaler                ProcessSignaler
	Checker                 ProcessChecker
	Bus                     events.Bus
	OutputChunkLimitBytes   int
	TerminationPollInterval time.Duration
	ForcedExitWait          time.Duration
}

// Manager executes tmux lifecycle operations and timeout cleanup.
type Manager struct {
	runner                  CommandRunner
	signaler                ProcessSignaler
	checker                 ProcessChecker
	bus                     events.Bus
	outputChunkLimitBytes   int
	terminationPollInterval time.Duration
	forcedExitWait          time.Duration
	now                     func() time.Time
	sleep                   func(time.Duration)
}

// New creates a tmux lifecycle manager with default dependencies where omitted.
func New(opts Options) (*Manager, error) {
	runner := opts.Runner
	if runner == nil {
		runner = defaultCommandRunner{}
	}

	signaler := opts.Signaler
	if signaler == nil {
		signaler = defaultProcessSignaler{}
	}

	checker := opts.Checker
	if checker == nil {
		checker = defaultProcessChecker{}
	}

	chunkLimit := opts.OutputChunkLimitBytes
	if chunkLimit <= 0 {
		chunkLimit = DefaultOutputChunkLimitBytes
	}

	pollInterval := opts.TerminationPollInterval
	if pollInterval <= 0 {
		pollInterval = defaultTerminationPollInterval
	}

	forcedExitWait := opts.ForcedExitWait
	if forcedExitWait <= 0 {
		forcedExitWait = defaultForcedExitWait
	}

	return &Manager{
		runner:                  runner,
		signaler:                signaler,
		checker:                 checker,
		bus:                     opts.Bus,
		outputChunkLimitBytes:   chunkLimit,
		terminationPollInterval: pollInterval,
		forcedExitWait:          forcedExitWait,
		now:                     time.Now,
		sleep:                   time.Sleep,
	}, nil
}

// CreateSession creates a detached named tmux session.
func (m *Manager) CreateSession(ctx context.Context, name string, cmd string, workdir string) error {
	if m == nil {
		return errors.New("tmux manager is nil")
	}
	if err := validateSessionName(name); err != nil {
		return err
	}

	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return errors.New("command is required")
	}

	workdir = strings.TrimSpace(workdir)
	if workdir == "" {
		return errors.New("workdir is required")
	}

	if _, err := m.runner.Run(ctx, "tmux", "new-session", "-d", "-s", name, "-c", workdir, cmd); err != nil {
		return fmt.Errorf("create tmux session %s: %w", name, err)
	}

	return nil
}

// ListSessions returns active tmux session names.
func (m *Manager) ListSessions(ctx context.Context) ([]TmuxSession, error) {
	if m == nil {
		return nil, errors.New("tmux manager is nil")
	}

	out, err := m.runner.Run(ctx, "tmux", "list-sessions", "-F", "#{session_name}")
	if err != nil {
		if isNoTmuxServerError(err) {
			return []TmuxSession{}, nil
		}
		return nil, fmt.Errorf("list tmux sessions: %w", err)
	}

	lines := strings.Split(string(out), "\n")
	sessions := make([]TmuxSession, 0, len(lines))
	for _, line := range lines {
		sessionName := strings.TrimSpace(line)
		if sessionName == "" {
			continue
		}
		sessions = append(sessions, TmuxSession{Name: sessionName})
	}
	return sessions, nil
}

// SendKeys sends keys followed by Enter to the target tmux session.
func (m *Manager) SendKeys(ctx context.Context, name string, keys string) error {
	if m == nil {
		return errors.New("tmux manager is nil")
	}
	if err := validateSessionName(name); err != nil {
		return err
	}

	keys = strings.TrimSpace(keys)
	if keys == "" {
		return errors.New("keys are required")
	}

	if _, err := m.runner.Run(ctx, "tmux", "send-keys", "-t", name, keys, "Enter"); err != nil {
		return fmt.Errorf("send keys to tmux session %s: %w", name, err)
	}

	return nil
}

// CapturePanes captures the latest pane output for the target tmux session.
func (m *Manager) CapturePanes(ctx context.Context, name string) (string, error) {
	if m == nil {
		return "", errors.New("tmux manager is nil")
	}
	if err := validateSessionName(name); err != nil {
		return "", err
	}

	out, err := m.runner.Run(ctx, "tmux", "capture-pane", "-pt", name, "-S", defaultCaptureStartLine)
	if err != nil {
		return "", fmt.Errorf("capture tmux panes for %s: %w", name, err)
	}
	return strings.TrimSpace(string(out)), nil
}

// KillSession kills a tmux session and ignores already-missing session errors.
func (m *Manager) KillSession(ctx context.Context, name string) error {
	if m == nil {
		return errors.New("tmux manager is nil")
	}
	if err := validateSessionName(name); err != nil {
		return err
	}

	if _, err := m.runner.Run(ctx, "tmux", "kill-session", "-t", name); err != nil {
		if isMissingSessionError(err) || isNoTmuxServerError(err) {
			return nil
		}
		return fmt.Errorf("kill tmux session %s: %w", name, err)
	}
	return nil
}

// StreamOutput polls capture-pane on an interval and emits output chunks to the event bus.
func (m *Manager) StreamOutput(ctx context.Context, name string, interval time.Duration, bus events.Bus) error {
	if m == nil {
		return errors.New("tmux manager is nil")
	}
	if err := validateSessionName(name); err != nil {
		return err
	}

	if interval <= 0 {
		interval = DefaultOutputPollInterval
	}

	eventBus := bus
	if eventBus == nil {
		eventBus = m.bus
	}
	if eventBus == nil {
		return errors.New("event bus is required for output streaming")
	}

	var lastSnapshot string

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			output, err := m.CapturePanes(ctx, name)
			if err != nil {
				if isMissingSessionError(err) || isNoTmuxServerError(err) {
					return nil
				}
				return err
			}

			chunk := diffChunk(lastSnapshot, output)
			lastSnapshot = output
			if chunk == "" {
				continue
			}

			truncatedChunk, truncated := truncateChunk(chunk, m.outputChunkLimitBytes)
			eventBus.Publish(events.Event{
				Type:       EventTypeSessionOutputChunk,
				EntityType: "session",
				EntityID:   name,
				Severity:   events.SeverityInfo,
				Payload: OutputChunk{
					SessionName: name,
					Chunk:       truncatedChunk,
					Truncated:   truncated,
					CapturedAt:  m.now().UTC(),
				},
			})
		}
	}
}

// EnforceTimeout applies deterministic SIGTERM -> grace -> SIGKILL escalation and cleans tmux session state.
func (m *Manager) EnforceTimeout(ctx context.Context, name string, pid int, gracePeriod time.Duration) error {
	if m == nil {
		return errors.New("tmux manager is nil")
	}
	if err := validateSessionName(name); err != nil {
		return err
	}

	if gracePeriod <= 0 {
		gracePeriod = DefaultTerminationGracePeriod
	}

	if pid <= 0 {
		return m.KillSession(ctx, name)
	}

	if err := m.signaler.Signal(pid, syscall.SIGTERM); err != nil && !isProcessGoneError(err) {
		return fmt.Errorf("send SIGTERM to pid %d: %w", pid, err)
	}

	exited, err := m.waitForExit(ctx, pid, gracePeriod)
	if err != nil {
		return fmt.Errorf("wait for pid %d after SIGTERM: %w", pid, err)
	}
	if !exited {
		if err := m.signaler.Signal(pid, syscall.SIGKILL); err != nil && !isProcessGoneError(err) {
			return fmt.Errorf("send SIGKILL to pid %d: %w", pid, err)
		}
		if _, waitErr := m.waitForExit(ctx, pid, m.forcedExitWait); waitErr != nil {
			return fmt.Errorf("wait for pid %d after SIGKILL: %w", pid, waitErr)
		}
	}

	if err := m.KillSession(ctx, name); err != nil {
		return err
	}

	alive, err := m.checker.Alive(pid)
	if err != nil {
		return fmt.Errorf("verify pid %d termination: %w", pid, err)
	}
	if alive {
		return fmt.Errorf("pid %d still alive after timeout enforcement", pid)
	}

	return nil
}

func (m *Manager) waitForExit(ctx context.Context, pid int, window time.Duration) (bool, error) {
	if window <= 0 {
		window = m.terminationPollInterval
	}

	deadline := m.now().Add(window)
	for {
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		default:
		}

		alive, err := m.checker.Alive(pid)
		if err != nil {
			return false, err
		}
		if !alive {
			return true, nil
		}
		if !m.now().Before(deadline) {
			return false, nil
		}
		m.sleep(m.terminationPollInterval)
	}
}

func validateSessionName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("session name is required")
	}
	if !sessionNamePattern.MatchString(name) {
		return fmt.Errorf("session name %q must match sc3-<role>-<mission-id>", name)
	}
	return nil
}

func isNoTmuxServerError(err error) bool {
	if err == nil {
		return false
	}
	text := strings.ToLower(err.Error())
	return strings.Contains(text, "no server running") || strings.Contains(text, "failed to connect to server")
}

func isMissingSessionError(err error) bool {
	if err == nil {
		return false
	}
	text := strings.ToLower(err.Error())
	return strings.Contains(text, "can't find session") || strings.Contains(text, "no such session")
}

func isProcessGoneError(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, syscall.ESRCH)
}

func diffChunk(previous string, current string) string {
	previous = strings.TrimSpace(previous)
	current = strings.TrimSpace(current)

	if current == "" || current == previous {
		return ""
	}
	if previous == "" {
		return current
	}
	if strings.HasPrefix(current, previous) {
		return strings.TrimSpace(current[len(previous):])
	}
	return current
}

func truncateChunk(chunk string, maxBytes int) (string, bool) {
	if maxBytes <= 0 {
		maxBytes = DefaultOutputChunkLimitBytes
	}
	if len(chunk) <= maxBytes {
		return chunk, false
	}
	return chunk[:maxBytes], true
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

// CreateSession creates a detached named tmux session.
func CreateSession(name string, cmd string, workdir string) error {
	manager, err := New(Options{})
	if err != nil {
		return err
	}
	return manager.CreateSession(context.Background(), name, cmd, workdir)
}

// ListSessions returns active tmux sessions.
func ListSessions() ([]TmuxSession, error) {
	manager, err := New(Options{})
	if err != nil {
		return nil, err
	}
	return manager.ListSessions(context.Background())
}

// SendKeys sends keys to a tmux session.
func SendKeys(name string, keys string) error {
	manager, err := New(Options{})
	if err != nil {
		return err
	}
	return manager.SendKeys(context.Background(), name, keys)
}

// CapturePanes captures output from a tmux session.
func CapturePanes(name string) (string, error) {
	manager, err := New(Options{})
	if err != nil {
		return "", err
	}
	return manager.CapturePanes(context.Background(), name)
}

// KillSession kills one tmux session.
func KillSession(name string) error {
	manager, err := New(Options{})
	if err != nil {
		return err
	}
	return manager.KillSession(context.Background(), name)
}

// StreamOutput polls and emits output chunks to an event bus.
func StreamOutput(ctx context.Context, name string, interval time.Duration, bus events.Bus) error {
	manager, err := New(Options{Bus: bus})
	if err != nil {
		return err
	}
	return manager.StreamOutput(ctx, name, interval, bus)
}

// EnforceTimeout escalates process termination then kills tmux session state.
func EnforceTimeout(ctx context.Context, name string, pid int) error {
	manager, err := New(Options{})
	if err != nil {
		return err
	}
	return manager.EnforceTimeout(ctx, name, pid, DefaultTerminationGracePeriod)
}

var _ CommandRunner = defaultCommandRunner{}
var _ ProcessSignaler = defaultProcessSignaler{}
var _ ProcessChecker = defaultProcessChecker{}
