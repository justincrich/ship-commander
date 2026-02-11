package harness

import "time"

// SessionStatus represents the lifecycle state of one harness-backed session.
type SessionStatus string

const (
	// SessionStatusRunning indicates a live session that can receive messages.
	SessionStatusRunning SessionStatus = "running"
	// SessionStatusExited indicates the underlying process has exited.
	SessionStatusExited SessionStatus = "exited"
	// SessionStatusTerminated indicates the session was terminated by orchestrator action.
	SessionStatusTerminated SessionStatus = "terminated"
)

// SessionOpts configures one harness-backed agent session.
type SessionOpts struct {
	Model    string
	MaxTurns int
	Timeout  time.Duration
	OnOutput func(chunk string)
}

// SessionResult captures structured process output from one harness interaction.
type SessionResult struct {
	ExitCode int
	Stdout   string
	Stderr   string
	Duration time.Duration
}

// Session is the runtime descriptor for one harness-backed agent session.
type Session struct {
	ID          string
	Role        string
	PID         int
	TmuxSession string
	StartedAt   time.Time
	Status      SessionStatus
	LastResult  SessionResult
}

// HarnessDriver provides a common session abstraction for CLI harness adapters.
//
//nolint:revive // Contract name is explicitly required by issue acceptance criteria.
type HarnessDriver interface {
	SpawnSession(role string, prompt string, workdir string, opts SessionOpts) (*Session, error)
	SendMessage(session *Session, message string) (string, error)
	Terminate(session *Session) error
}
