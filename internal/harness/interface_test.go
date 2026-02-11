package harness

import (
	"reflect"
	"testing"
	"time"
)

type fakeDriver struct{}

func (f *fakeDriver) SpawnSession(role string, prompt string, workdir string, opts SessionOpts) (*Session, error) {
	return &Session{
		ID:          "session-1",
		Role:        role,
		PID:         1234,
		TmuxSession: "sc3-captain-m1",
		StartedAt:   time.Now().UTC(),
		Status:      SessionStatusRunning,
		LastResult: SessionResult{
			ExitCode: 0,
			Stdout:   prompt,
			Stderr:   "",
			Duration: opts.Timeout,
		},
	}, nil
}

func (f *fakeDriver) SendMessage(session *Session, message string) (string, error) {
	if session == nil {
		return "", nil
	}
	session.LastResult.Stdout = message
	return message, nil
}

func (f *fakeDriver) Terminate(session *Session) error {
	if session != nil {
		session.Status = SessionStatusTerminated
	}
	return nil
}

func TestHarnessDriverDefinesRequiredMethods(t *testing.T) {
	var _ HarnessDriver = (*fakeDriver)(nil)

	driverType := reflect.TypeOf((*HarnessDriver)(nil)).Elem()
	if _, ok := driverType.MethodByName("SpawnSession"); !ok {
		t.Fatal("expected SpawnSession on HarnessDriver")
	}
	if _, ok := driverType.MethodByName("SendMessage"); !ok {
		t.Fatal("expected SendMessage on HarnessDriver")
	}
	if _, ok := driverType.MethodByName("Terminate"); !ok {
		t.Fatal("expected Terminate on HarnessDriver")
	}
}

func TestSessionOptsSupportsConfiguredFields(t *testing.T) {
	timeout := 30 * time.Second
	received := ""

	opts := SessionOpts{
		Model:    "gpt-5",
		MaxTurns: 7,
		Timeout:  timeout,
		OnOutput: func(chunk string) {
			received = chunk
		},
	}
	opts.OnOutput("chunk")

	if opts.Model != "gpt-5" {
		t.Fatalf("model = %q, want gpt-5", opts.Model)
	}
	if opts.MaxTurns != 7 {
		t.Fatalf("max turns = %d, want 7", opts.MaxTurns)
	}
	if opts.Timeout != timeout {
		t.Fatalf("timeout = %s, want %s", opts.Timeout, timeout)
	}
	if received != "chunk" {
		t.Fatalf("on output callback = %q, want chunk", received)
	}
}

func TestSessionResultCarriesStructuredCommandOutput(t *testing.T) {
	result := SessionResult{
		ExitCode: 42,
		Stdout:   "stdout content",
		Stderr:   "stderr content",
		Duration: 1250 * time.Millisecond,
	}
	session := Session{
		ID:         "s-1",
		Role:       "captain",
		LastResult: result,
	}

	if session.LastResult.ExitCode != 42 {
		t.Fatalf("exit code = %d, want 42", session.LastResult.ExitCode)
	}
	if session.ID != "s-1" {
		t.Fatalf("session ID = %q, want s-1", session.ID)
	}
	if session.Role != "captain" {
		t.Fatalf("session role = %q, want captain", session.Role)
	}
	if session.LastResult.Stdout != "stdout content" {
		t.Fatalf("stdout = %q, want stdout content", session.LastResult.Stdout)
	}
	if session.LastResult.Stderr != "stderr content" {
		t.Fatalf("stderr = %q, want stderr content", session.LastResult.Stderr)
	}
	if session.LastResult.Duration != 1250*time.Millisecond {
		t.Fatalf("duration = %s, want 1.25s", session.LastResult.Duration)
	}
}
