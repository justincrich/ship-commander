package claude

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/ship-commander/sc3/internal/harness"
)

func TestSpawnSessionConstructsClaudeCLIFlags(t *testing.T) {
	runner := &fakeRunner{
		outputs: map[string][]byte{
			"tmux list-panes -t sc3-captain-mission-42 -F #{pane_pid}": []byte("1234\n"),
		},
	}
	driver, err := NewWithRunner(runner, DriverConfig{})
	if err != nil {
		t.Fatalf("new driver: %v", err)
	}
	driver.now = fixedNow

	session, err := driver.SpawnSession(
		"captain",
		"Work mission MISSION-42 immediately",
		"/tmp/worktree",
		harness.SessionOpts{Model: "opus", MaxTurns: 4},
	)
	if err != nil {
		t.Fatalf("spawn session: %v", err)
	}

	call := runner.findCall(t, "tmux", "new-session")
	if !containsInOrder(call.args, []string{"-d", "-s", "sc3-captain-mission-42", "-c", "/tmp/worktree"}) {
		t.Fatalf("new-session args = %v, missing expected tmux session args", call.args)
	}
	commandArg := call.args[len(call.args)-1]
	for _, expected := range []string{
		"claude -p",
		"--model opus",
		"--verbose",
		"--max-turns 4",
	} {
		if !strings.Contains(commandArg, expected) {
			t.Fatalf("claude command = %q, missing %q", commandArg, expected)
		}
	}

	if session.TmuxSession != "sc3-captain-mission-42" {
		t.Fatalf("tmux session = %q, want sc3-captain-mission-42", session.TmuxSession)
	}
	if session.PID != 1234 {
		t.Fatalf("pid = %d, want 1234", session.PID)
	}
}

func TestSpawnSessionUsesRoleModelFallback(t *testing.T) {
	runner := &fakeRunner{
		outputs: map[string][]byte{
			"tmux list-panes -t sc3-ensign-backend-session -F #{pane_pid}": []byte("12\n"),
		},
	}
	driver, err := NewWithRunner(runner, DriverConfig{
		RoleModels: map[string]string{
			"ensign-backend": "haiku",
		},
	})
	if err != nil {
		t.Fatalf("new driver: %v", err)
	}

	_, err = driver.SpawnSession(
		"ensign-backend",
		"No mission id in this prompt",
		"/tmp/worktree",
		harness.SessionOpts{MaxTurns: 2},
	)
	if err != nil {
		t.Fatalf("spawn session: %v", err)
	}

	call := runner.findCall(t, "tmux", "new-session")
	commandArg := call.args[len(call.args)-1]
	if !strings.Contains(commandArg, "--model haiku") {
		t.Fatalf("claude command = %q, want role-model fallback haiku", commandArg)
	}
}

func TestSendMessageCapturesOutputViaTmuxCapturePane(t *testing.T) {
	runner := &fakeRunner{
		outputs: map[string][]byte{
			"tmux capture-pane -pt sc3-captain-mission-42 -S -200": []byte("assistant response\n"),
		},
	}
	driver, err := NewWithRunner(runner, DriverConfig{})
	if err != nil {
		t.Fatalf("new driver: %v", err)
	}

	session := &harness.Session{
		ID:          "sc3-captain-mission-42",
		TmuxSession: "sc3-captain-mission-42",
	}
	driver.sessionOpts[session.ID] = harness.SessionOpts{}

	response, err := driver.SendMessage(session, "continue")
	if err != nil {
		t.Fatalf("send message: %v", err)
	}
	if response != "assistant response" {
		t.Fatalf("response = %q, want assistant response", response)
	}

	runner.findCall(t, "tmux", "send-keys")
	runner.findCall(t, "tmux", "capture-pane")
	if session.LastResult.Stdout != "assistant response" {
		t.Fatalf("session stdout = %q, want assistant response", session.LastResult.Stdout)
	}
}

func TestTerminateKillsTmuxSession(t *testing.T) {
	runner := &fakeRunner{}
	driver, err := NewWithRunner(runner, DriverConfig{})
	if err != nil {
		t.Fatalf("new driver: %v", err)
	}

	session := &harness.Session{ID: "sc3-captain-mission-42", TmuxSession: "sc3-captain-mission-42"}
	if err := driver.Terminate(session); err != nil {
		t.Fatalf("terminate: %v", err)
	}
	if session.Status != harness.SessionStatusTerminated {
		t.Fatalf("session status = %q, want %q", session.Status, harness.SessionStatusTerminated)
	}
	runner.findCall(t, "tmux", "kill-session")
}

func TestSpawnSessionRejectsUnsupportedModel(t *testing.T) {
	driver, err := NewWithRunner(&fakeRunner{}, DriverConfig{})
	if err != nil {
		t.Fatalf("new driver: %v", err)
	}

	_, err = driver.SpawnSession("captain", "MISSION-4", "/tmp/worktree", harness.SessionOpts{
		Model: "invalid",
	})
	if err == nil {
		t.Fatal("expected unsupported model error")
	}
	if !strings.Contains(err.Error(), "unsupported claude model") {
		t.Fatalf("unexpected error: %v", err)
	}
}

type runnerCall struct {
	name string
	args []string
}

type fakeRunner struct {
	calls   []runnerCall
	outputs map[string][]byte
	errors  map[string]error
}

func (f *fakeRunner) Run(_ context.Context, name string, args ...string) ([]byte, error) {
	call := runnerCall{
		name: name,
		args: append([]string(nil), args...),
	}
	f.calls = append(f.calls, call)

	key := callKey(name, args)
	if err, ok := f.errors[key]; ok {
		return nil, err
	}
	if out, ok := f.outputs[key]; ok {
		return out, nil
	}
	return []byte{}, nil
}

func (f *fakeRunner) findCall(t *testing.T, name string, subcommand string) runnerCall {
	t.Helper()
	for _, call := range f.calls {
		if call.name != name {
			continue
		}
		if len(call.args) == 0 {
			continue
		}
		if call.args[0] == subcommand {
			return call
		}
	}
	t.Fatalf("call %s %s not found in %v", name, subcommand, f.calls)
	return runnerCall{}
}

func callKey(name string, args []string) string {
	parts := append([]string{name}, args...)
	return strings.Join(parts, " ")
}

func containsInOrder(args []string, expected []string) bool {
	if len(expected) == 0 {
		return true
	}
	index := 0
	for _, arg := range args {
		if arg == expected[index] {
			index++
			if index == len(expected) {
				return true
			}
		}
	}
	return false
}

func fixedNow() time.Time {
	return time.Date(2026, 2, 11, 12, 0, 0, 0, time.UTC)
}

var _ CommandRunner = (*fakeRunner)(nil)

func TestNewWithRunnerRejectsNilRunner(t *testing.T) {
	_, err := NewWithRunner(nil, DriverConfig{})
	if err == nil {
		t.Fatal("expected nil runner error")
	}
	if !strings.Contains(err.Error(), "runner is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}
