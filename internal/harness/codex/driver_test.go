package codex

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/ship-commander/sc3/internal/harness"
)

func TestSpawnSessionConstructsCodexCLIFlags(t *testing.T) {
	runner := &fakeRunner{}
	driver, err := NewWithRunner(runner, DriverConfig{
		SandboxMode:    "workspace-write",
		ApprovalPolicy: "on-request",
	})
	if err != nil {
		t.Fatalf("new driver: %v", err)
	}
	driver.now = fixedNow

	session, err := driver.SpawnSession(
		"ensign-backend",
		"Implement feature for MISSION-88",
		"/tmp/worktree",
		harness.SessionOpts{Model: "gpt-5-codex"},
	)
	if err != nil {
		t.Fatalf("spawn session: %v", err)
	}

	call := runner.findCall(t, "tmux", "new-session")
	commandArg := call.args[len(call.args)-1]
	for _, expected := range []string{
		"codex --sandbox workspace-write",
		"--approval-policy on-request",
		"-m gpt-5-codex",
		"exec -",
	} {
		if !strings.Contains(commandArg, expected) {
			t.Fatalf("codex command = %q, missing %q", commandArg, expected)
		}
	}

	if session.TmuxSession != "sc3-ensign-backend-mission-88" {
		t.Fatalf("tmux session = %q, want sc3-ensign-backend-mission-88", session.TmuxSession)
	}
}

func TestSpawnSessionUsesRoleModelFallback(t *testing.T) {
	runner := &fakeRunner{}
	driver, err := NewWithRunner(runner, DriverConfig{
		RoleModels: map[string]string{
			"ensign-backend": "gpt-5-mini",
		},
		SandboxMode:    "read-only",
		ApprovalPolicy: "never",
	})
	if err != nil {
		t.Fatalf("new driver: %v", err)
	}

	_, err = driver.SpawnSession(
		"ensign-backend",
		"No mission id prompt",
		"/tmp/worktree",
		harness.SessionOpts{},
	)
	if err != nil {
		t.Fatalf("spawn session: %v", err)
	}

	call := runner.findCall(t, "tmux", "new-session")
	commandArg := call.args[len(call.args)-1]
	if !strings.Contains(commandArg, "-m gpt-5-mini") {
		t.Fatalf("codex command = %q, missing role model fallback", commandArg)
	}
}

func TestSendMessageCapturesOutputViaTmuxCapturePane(t *testing.T) {
	runner := &fakeRunner{
		outputs: map[string][]byte{
			"tmux capture-pane -pt sc3-ensign-mission-7 -S -200": []byte("codex response\n"),
		},
	}
	driver, err := NewWithRunner(runner, DriverConfig{})
	if err != nil {
		t.Fatalf("new driver: %v", err)
	}

	session := &harness.Session{
		ID:          "sc3-ensign-mission-7",
		TmuxSession: "sc3-ensign-mission-7",
	}
	driver.sessionOpts[session.ID] = harness.SessionOpts{}

	response, err := driver.SendMessage(session, "continue")
	if err != nil {
		t.Fatalf("send message: %v", err)
	}
	if response != "codex response" {
		t.Fatalf("response = %q, want codex response", response)
	}

	runner.findCall(t, "tmux", "send-keys")
	runner.findCall(t, "tmux", "capture-pane")
}

func TestTerminateKillsTmuxSession(t *testing.T) {
	runner := &fakeRunner{}
	driver, err := NewWithRunner(runner, DriverConfig{})
	if err != nil {
		t.Fatalf("new driver: %v", err)
	}

	session := &harness.Session{ID: "sc3-ensign-mission-7", TmuxSession: "sc3-ensign-mission-7"}
	if err := driver.Terminate(session); err != nil {
		t.Fatalf("terminate: %v", err)
	}
	if session.Status != harness.SessionStatusTerminated {
		t.Fatalf("session status = %q, want %q", session.Status, harness.SessionStatusTerminated)
	}
	runner.findCall(t, "tmux", "kill-session")
}

func TestNewWithRunnerRejectsUnsupportedSandboxOrApproval(t *testing.T) {
	if _, err := NewWithRunner(&fakeRunner{}, DriverConfig{SandboxMode: "invalid"}); err == nil {
		t.Fatal("expected sandbox mode validation error")
	}
	if _, err := NewWithRunner(&fakeRunner{}, DriverConfig{ApprovalPolicy: "invalid"}); err == nil {
		t.Fatal("expected approval policy validation error")
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

func fixedNow() time.Time {
	return time.Date(2026, 2, 11, 12, 0, 0, 0, time.UTC)
}

var _ CommandRunner = (*fakeRunner)(nil)
