package tmux

import (
	"context"
	"errors"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/ship-commander/sc3/internal/events"
)

func TestCreateSessionRunsTmuxNewSession(t *testing.T) {
	t.Parallel()

	runner := &fakeRunner{}
	manager, err := New(Options{Runner: runner})
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}

	err = manager.CreateSession(context.Background(), "sc3-captain-mission-42", "echo ready", "/tmp/worktree")
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	call := runner.findCall(t, "tmux", "new-session")
	if !containsInOrder(call.args, []string{"-d", "-s", "sc3-captain-mission-42", "-c", "/tmp/worktree", "echo ready"}) {
		t.Fatalf("new-session args = %v", call.args)
	}
}

func TestCreateSessionRejectsInvalidSessionName(t *testing.T) {
	t.Parallel()

	manager, err := New(Options{Runner: &fakeRunner{}})
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}

	err = manager.CreateSession(context.Background(), "captain-mission-42", "echo ready", "/tmp/worktree")
	if err == nil {
		t.Fatal("expected invalid session name error")
	}
}

func TestListSessionsParsesTmuxOutput(t *testing.T) {
	t.Parallel()

	runner := &fakeRunner{
		outputs: map[string][][]byte{
			"tmux list-sessions -F #{session_name}": {
				[]byte("sc3-captain-mission-1\nsc3-commander-mission-2\n"),
			},
		},
	}
	manager, err := New(Options{Runner: runner})
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}

	sessions, err := manager.ListSessions(context.Background())
	if err != nil {
		t.Fatalf("list sessions: %v", err)
	}

	if len(sessions) != 2 {
		t.Fatalf("session count = %d, want 2", len(sessions))
	}
	if sessions[0].Name != "sc3-captain-mission-1" {
		t.Fatalf("first session = %q", sessions[0].Name)
	}
}

func TestListSessionsNoServerReturnsEmpty(t *testing.T) {
	t.Parallel()

	runner := &fakeRunner{
		errors: map[string]error{
			"tmux list-sessions -F #{session_name}": errors.New("failed to connect to server"),
		},
	}
	manager, err := New(Options{Runner: runner})
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}

	sessions, err := manager.ListSessions(context.Background())
	if err != nil {
		t.Fatalf("list sessions: %v", err)
	}
	if len(sessions) != 0 {
		t.Fatalf("session count = %d, want 0", len(sessions))
	}
}

func TestSendKeysCaptureAndKillSession(t *testing.T) {
	t.Parallel()

	runner := &fakeRunner{
		outputs: map[string][][]byte{
			"tmux capture-pane -pt sc3-captain-mission-42 -S -2000": {
				[]byte("assistant output\n"),
			},
		},
	}
	manager, err := New(Options{Runner: runner})
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}

	if err := manager.SendKeys(context.Background(), "sc3-captain-mission-42", "continue"); err != nil {
		t.Fatalf("send keys: %v", err)
	}
	output, err := manager.CapturePanes(context.Background(), "sc3-captain-mission-42")
	if err != nil {
		t.Fatalf("capture panes: %v", err)
	}
	if output != "assistant output" {
		t.Fatalf("capture output = %q, want %q", output, "assistant output")
	}
	if err := manager.KillSession(context.Background(), "sc3-captain-mission-42"); err != nil {
		t.Fatalf("kill session: %v", err)
	}

	runner.findCall(t, "tmux", "send-keys")
	runner.findCall(t, "tmux", "capture-pane")
	runner.findCall(t, "tmux", "kill-session")
}

func TestStreamOutputPublishesChunksWithTruncation(t *testing.T) {
	t.Parallel()

	runner := &fakeRunner{
		outputs: map[string][][]byte{
			"tmux capture-pane -pt sc3-captain-mission-42 -S -2000": {
				[]byte("hello"),
				[]byte("hello-0123456789"),
				[]byte("hello-0123456789"),
			},
		},
	}

	bus := newCaptureBus()
	manager, err := New(Options{
		Runner:                runner,
		Bus:                   bus,
		OutputChunkLimitBytes: 8,
	})
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- manager.StreamOutput(ctx, "sc3-captain-mission-42", 2*time.Millisecond, nil)
	}()

	first := bus.waitForEvent(t, 2*time.Second)
	second := bus.waitForEvent(t, 2*time.Second)
	cancel()

	select {
	case streamErr := <-errCh:
		if streamErr != nil {
			t.Fatalf("stream output: %v", streamErr)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for stream shutdown")
	}

	firstPayload, ok := first.Payload.(OutputChunk)
	if !ok {
		t.Fatalf("first payload type = %T, want OutputChunk", first.Payload)
	}
	if firstPayload.Chunk != "hello" {
		t.Fatalf("first chunk = %q, want %q", firstPayload.Chunk, "hello")
	}
	if firstPayload.Truncated {
		t.Fatal("first chunk should not be truncated")
	}

	secondPayload, ok := second.Payload.(OutputChunk)
	if !ok {
		t.Fatalf("second payload type = %T, want OutputChunk", second.Payload)
	}
	if secondPayload.Chunk != "-0123456" {
		t.Fatalf("second chunk = %q, want %q", secondPayload.Chunk, "-0123456")
	}
	if !secondPayload.Truncated {
		t.Fatal("second chunk should be truncated")
	}
}

func TestEnforceTimeoutEscalatesTermThenKill(t *testing.T) {
	t.Parallel()

	runner := &fakeRunner{}
	signaler := &fakeSignaler{}
	checker := &fakeChecker{alive: true}
	signaler.onSignal = func(signal syscall.Signal) {
		if signal == syscall.SIGKILL {
			checker.setAlive(false)
		}
	}

	manager, err := New(Options{
		Runner:                  runner,
		Signaler:                signaler,
		Checker:                 checker,
		TerminationPollInterval: time.Millisecond,
		ForcedExitWait:          5 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}

	err = manager.EnforceTimeout(
		context.Background(),
		"sc3-captain-mission-42",
		1234,
		3*time.Millisecond,
	)
	if err != nil {
		t.Fatalf("enforce timeout: %v", err)
	}

	if len(signaler.signals) != 2 {
		t.Fatalf("signal count = %d, want 2", len(signaler.signals))
	}
	if signaler.signals[0] != syscall.SIGTERM {
		t.Fatalf("first signal = %v, want SIGTERM", signaler.signals[0])
	}
	if signaler.signals[1] != syscall.SIGKILL {
		t.Fatalf("second signal = %v, want SIGKILL", signaler.signals[1])
	}

	runner.findCall(t, "tmux", "kill-session")
}

type runnerCall struct {
	name string
	args []string
}

type fakeRunner struct {
	mu      sync.Mutex
	calls   []runnerCall
	outputs map[string][][]byte
	errors  map[string]error
}

func (f *fakeRunner) Run(_ context.Context, name string, args ...string) ([]byte, error) {
	call := runnerCall{
		name: name,
		args: append([]string(nil), args...),
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	f.calls = append(f.calls, call)

	key := callKey(name, args)
	if err, ok := f.errors[key]; ok {
		return nil, err
	}
	if queue, ok := f.outputs[key]; ok && len(queue) > 0 {
		next := queue[0]
		f.outputs[key] = queue[1:]
		return next, nil
	}
	return []byte{}, nil
}

func (f *fakeRunner) findCall(t *testing.T, name string, subcommand string) runnerCall {
	t.Helper()

	f.mu.Lock()
	defer f.mu.Unlock()

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
	idx := 0
	for _, arg := range args {
		if arg == expected[idx] {
			idx++
			if idx == len(expected) {
				return true
			}
		}
	}
	return false
}

type fakeSignaler struct {
	mu       sync.Mutex
	signals  []syscall.Signal
	onSignal func(syscall.Signal)
}

func (f *fakeSignaler) Signal(_ int, signal syscall.Signal) error {
	f.mu.Lock()
	f.signals = append(f.signals, signal)
	callback := f.onSignal
	f.mu.Unlock()
	if callback != nil {
		callback(signal)
	}
	return nil
}

type fakeChecker struct {
	mu        sync.Mutex
	responses []bool
	index     int
	alive     bool
}

func (f *fakeChecker) Alive(_ int) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if len(f.responses) == 0 {
		return f.alive, nil
	}
	if f.index >= len(f.responses) {
		return f.responses[len(f.responses)-1], nil
	}
	value := f.responses[f.index]
	f.index++
	return value, nil
}

func (f *fakeChecker) setAlive(value bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.alive = value
}

type captureBus struct {
	events chan events.Event
}

func newCaptureBus() *captureBus {
	return &captureBus{events: make(chan events.Event, 8)}
}

func (b *captureBus) Subscribe(_ string, _ events.Handler) {}

func (b *captureBus) SubscribeAll(_ events.Handler) {}

func (b *captureBus) Publish(event events.Event) {
	b.events <- event
}

func (b *captureBus) waitForEvent(t *testing.T, timeout time.Duration) events.Event {
	t.Helper()

	select {
	case event := <-b.events:
		return event
	case <-time.After(timeout):
		t.Fatal("timed out waiting for event")
		return events.Event{}
	}
}

var _ CommandRunner = (*fakeRunner)(nil)
var _ ProcessSignaler = (*fakeSignaler)(nil)
var _ ProcessChecker = (*fakeChecker)(nil)
var _ events.Bus = (*captureBus)(nil)
