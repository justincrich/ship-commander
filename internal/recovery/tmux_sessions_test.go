package recovery

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

type fakeTmuxRunner struct {
	outputs map[string][]byte
	errors  map[string]error
	calls   [][]string
}

func (f *fakeTmuxRunner) Run(_ context.Context, _ string, args ...string) ([]byte, error) {
	f.calls = append(f.calls, append([]string(nil), args...))
	key := commandKey(args)
	if err, ok := f.errors[key]; ok {
		return nil, err
	}
	if out, ok := f.outputs[key]; ok {
		return out, nil
	}
	return []byte{}, nil
}

func commandKey(args []string) string {
	if len(args) == 0 {
		return ""
	}
	return args[0]
}

func TestActiveSessionsParsesTmuxOutput(t *testing.T) {
	t.Parallel()

	runner := &fakeTmuxRunner{
		outputs: map[string][]byte{"list-sessions": []byte("captain\ncommander\n")},
		errors:  map[string]error{},
	}
	manager, err := NewTmuxSessionManagerWithRunner(runner)
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}

	sessions, err := manager.ActiveSessions(context.Background())
	if err != nil {
		t.Fatalf("active sessions: %v", err)
	}

	want := map[string]struct{}{"captain": {}, "commander": {}}
	if !reflect.DeepEqual(sessions, want) {
		t.Fatalf("sessions = %#v, want %#v", sessions, want)
	}
}

func TestActiveSessionsNoTmuxServerReturnsEmptySet(t *testing.T) {
	t.Parallel()

	runner := &fakeTmuxRunner{
		outputs: map[string][]byte{},
		errors:  map[string]error{"list-sessions": errors.New("failed to connect to server")},
	}
	manager, err := NewTmuxSessionManagerWithRunner(runner)
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}

	sessions, err := manager.ActiveSessions(context.Background())
	if err != nil {
		t.Fatalf("active sessions: %v", err)
	}
	if len(sessions) != 0 {
		t.Fatalf("sessions len = %d, want 0", len(sessions))
	}
}

func TestCleanupDeadSessionIgnoresMissingSession(t *testing.T) {
	t.Parallel()

	runner := &fakeTmuxRunner{
		outputs: map[string][]byte{},
		errors:  map[string]error{"kill-session": errors.New("can't find session")},
	}
	manager, err := NewTmuxSessionManagerWithRunner(runner)
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}

	if err := manager.CleanupDeadSession(context.Background(), "session-missing"); err != nil {
		t.Fatalf("cleanup dead session: %v", err)
	}
}

func TestCleanupDeadSessionWrapsUnexpectedError(t *testing.T) {
	t.Parallel()

	runner := &fakeTmuxRunner{
		outputs: map[string][]byte{},
		errors:  map[string]error{"kill-session": errors.New("permission denied")},
	}
	manager, err := NewTmuxSessionManagerWithRunner(runner)
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}

	if err := manager.CleanupDeadSession(context.Background(), "session-1"); err == nil {
		t.Fatal("expected cleanup error")
	}
}
