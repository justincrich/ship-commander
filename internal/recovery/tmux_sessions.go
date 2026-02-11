package recovery

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// TmuxSessionManager checks active tmux sessions and cleans up dead session handles.
type TmuxSessionManager struct {
	runner CommandRunner
}

// NewTmuxSessionManager creates a tmux session manager.
func NewTmuxSessionManager() (*TmuxSessionManager, error) {
	return NewTmuxSessionManagerWithRunner(defaultCommandRunner{})
}

// NewTmuxSessionManagerWithRunner creates a tmux session manager with a custom runner.
func NewTmuxSessionManagerWithRunner(runner CommandRunner) (*TmuxSessionManager, error) {
	if runner == nil {
		return nil, errors.New("runner must not be nil")
	}
	return &TmuxSessionManager{runner: runner}, nil
}

// ActiveSessions returns all currently running tmux session names.
func (m *TmuxSessionManager) ActiveSessions(ctx context.Context) (map[string]struct{}, error) {
	if m == nil {
		return nil, errors.New("tmux session manager is nil")
	}
	out, err := m.runner.Run(ctx, "tmux", "list-sessions", "-F", "#{session_name}")
	if err != nil {
		if isNoTmuxServerError(err) {
			return map[string]struct{}{}, nil
		}
		return nil, fmt.Errorf("list tmux sessions: %w", err)
	}

	sessions := map[string]struct{}{}
	for _, line := range strings.Split(string(out), "\n") {
		sessionID := strings.TrimSpace(line)
		if sessionID == "" {
			continue
		}
		sessions[sessionID] = struct{}{}
	}
	return sessions, nil
}

// CleanupDeadSession clears tmux session state for a dead session identifier.
func (m *TmuxSessionManager) CleanupDeadSession(ctx context.Context, sessionID string) error {
	if m == nil {
		return errors.New("tmux session manager is nil")
	}
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return errors.New("session id must not be empty")
	}
	_, err := m.runner.Run(ctx, "tmux", "kill-session", "-t", sessionID)
	if err != nil {
		if isMissingSessionError(err) || isNoTmuxServerError(err) {
			return nil
		}
		return fmt.Errorf("kill tmux session %s: %w", sessionID, err)
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

var _ CommandRunner = defaultCommandRunner{}
