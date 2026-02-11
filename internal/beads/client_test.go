package beads

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type fakeResult struct {
	stdout []byte
	stderr []byte
	err    error
}

type fakeCall struct {
	dir  string
	name string
	args []string
}

type fakeCommandRunner struct {
	results []fakeResult
	calls   []fakeCall
}

func (f *fakeCommandRunner) Run(_ context.Context, dir string, name string, args ...string) ([]byte, []byte, error) {
	f.calls = append(f.calls, fakeCall{
		dir:  dir,
		name: name,
		args: append([]string{}, args...),
	})

	if len(f.results) == 0 {
		return nil, nil, errors.New("unexpected command call")
	}

	next := f.results[0]
	f.results = f.results[1:]
	return next.stdout, next.stderr, next.err
}

func TestNewClientChecksCLIAvailability(t *testing.T) {
	t.Parallel()

	workDir := t.TempDir()
	runner := &fakeCommandRunner{
		results: []fakeResult{
			{stdout: []byte(`{"version":"1.0.0"}`)},
		},
	}

	client, err := newClient(workDir, "sh", time.Second, runner)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	if client == nil {
		t.Fatal("client is nil")
	}

	if len(runner.calls) != 1 {
		t.Fatalf("calls = %d, want 1", len(runner.calls))
	}
	if !containsArgsInOrder(runner.calls[0].args, []string{"version", "--json"}) {
		t.Fatalf("version call args = %v, want version --json", runner.calls[0].args)
	}
}

func TestNewClientWrapsVersionErrors(t *testing.T) {
	t.Parallel()

	workDir := t.TempDir()
	runner := &fakeCommandRunner{
		results: []fakeResult{
			{
				stderr: []byte("boom"),
				err:    errors.New("exit 1"),
			},
		},
	}

	_, err := newClient(workDir, "sh", time.Second, runner)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "check sh availability") {
		t.Fatalf("error = %v, want availability context", err)
	}
	if !strings.Contains(err.Error(), "run sh version --json") {
		t.Fatalf("error = %v, want command context", err)
	}
	if !strings.Contains(err.Error(), "stderr: boom") {
		t.Fatalf("error = %v, want stderr context", err)
	}
}

func TestInitRunsInitWhenBeadsMissing(t *testing.T) {
	t.Parallel()

	workDir := t.TempDir()
	runner := &fakeCommandRunner{
		results: []fakeResult{
			{stdout: []byte(`{"version":"1.0.0"}`)},
			{stdout: []byte(`{"ok":true}`)},
		},
	}

	client, err := newClient(workDir, "sh", time.Second, runner)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	if err := client.Init(); err != nil {
		t.Fatalf("init: %v", err)
	}

	if len(runner.calls) != 2 {
		t.Fatalf("calls = %d, want 2", len(runner.calls))
	}
	if !containsArgsInOrder(runner.calls[1].args, []string{"init", "--json"}) {
		t.Fatalf("init call args = %v, want init --json", runner.calls[1].args)
	}
}

func TestInitReusesExistingBeadsDirectory(t *testing.T) {
	t.Parallel()

	workDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(workDir, ".beads"), 0o750); err != nil {
		t.Fatalf("mkdir .beads: %v", err)
	}

	runner := &fakeCommandRunner{
		results: []fakeResult{
			{stdout: []byte(`{"version":"1.0.0"}`)},
		},
	}

	client, err := newClient(workDir, "sh", time.Second, runner)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	if err := client.Init(); err != nil {
		t.Fatalf("init: %v", err)
	}
	if len(runner.calls) != 1 {
		t.Fatalf("calls = %d, want 1 (version check only)", len(runner.calls))
	}
}

func TestCreateReturnsIDAndBuildsExpectedArgs(t *testing.T) {
	t.Parallel()

	workDir := t.TempDir()
	parent := "ship-commander-3-abc"
	runner := &fakeCommandRunner{
		results: []fakeResult{
			{stdout: []byte(`{"version":"1.0.0"}`)},
			{stdout: []byte(`{"id":"ship-commander-3-abc.1","title":"test"}`)},
		},
	}

	client, err := newClient(workDir, "sh", time.Second, runner)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	id, err := client.Create(CreateOpts{
		Title:       "test",
		Type:        "task",
		Description: "desc",
		Parent:      &parent,
		Labels:      []string{"phase:1", "type:mission"},
		Priority:    "1",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if id != "ship-commander-3-abc.1" {
		t.Fatalf("id = %q, want %q", id, "ship-commander-3-abc.1")
	}

	call := runner.calls[1]
	if !containsArgsInOrder(call.args, []string{
		"create",
		"--type", "task",
		"--title", "test",
		"--description", "desc",
		"--parent", parent,
		"--labels", "phase:1,type:mission",
		"--priority", "1",
		"--json",
	}) {
		t.Fatalf("create args = %v", call.args)
	}
}

func TestShowParsesArrayOutput(t *testing.T) {
	t.Parallel()

	workDir := t.TempDir()
	runner := &fakeCommandRunner{
		results: []fakeResult{
			{stdout: []byte(`{"version":"1.0.0"}`)},
			{stdout: []byte(`[{"id":"ship-commander-3-1","title":"a","status":"open","priority":2,"issue_type":"task"}]`)},
		},
	}

	client, err := newClient(workDir, "sh", time.Second, runner)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	bead, err := client.Show("ship-commander-3-1")
	if err != nil {
		t.Fatalf("show: %v", err)
	}
	if bead.ID != "ship-commander-3-1" {
		t.Fatalf("bead id = %q, want ship-commander-3-1", bead.ID)
	}
}

func TestListAppliesFilters(t *testing.T) {
	t.Parallel()

	workDir := t.TempDir()
	runner := &fakeCommandRunner{
		results: []fakeResult{
			{stdout: []byte(`{"version":"1.0.0"}`)},
			{stdout: []byte(`[{"id":"ship-commander-3-1","title":"a","status":"open","priority":2,"issue_type":"task"}]`)},
		},
	}

	client, err := newClient(workDir, "sh", time.Second, runner)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	issues, err := client.List(ListOpts{
		Type:   "task",
		Status: "open",
		Limit:  25,
		Parent: "ship-commander-3-9sd",
		Labels: []string{"phase:1", "type:mission"},
	})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("issues length = %d, want 1", len(issues))
	}

	call := runner.calls[1]
	if !containsArgsInOrder(call.args, []string{
		"list",
		"--type", "task",
		"--status", "open",
		"--limit", "25",
		"--parent", "ship-commander-3-9sd",
		"--label", "phase:1",
		"--label", "type:mission",
		"--json",
	}) {
		t.Fatalf("list args = %v", call.args)
	}
}

func TestSetStateIncludesCommandContextOnFailure(t *testing.T) {
	t.Parallel()

	workDir := t.TempDir()
	runner := &fakeCommandRunner{
		results: []fakeResult{
			{stdout: []byte(`{"version":"1.0.0"}`)},
			{
				stderr: []byte("state failure"),
				err:    errors.New("exit status 1"),
			},
		},
	}

	client, err := newClient(workDir, "sh", time.Second, runner)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	err = client.SetState("ship-commander-3-1", "phase", "green")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "set state \"phase\"") {
		t.Fatalf("error = %v, want set state context", err)
	}
	if !strings.Contains(err.Error(), "run sh set-state ship-commander-3-1 phase=green --json") {
		t.Fatalf("error = %v, want command/args context", err)
	}
	if !strings.Contains(err.Error(), "stderr: state failure") {
		t.Fatalf("error = %v, want stderr context", err)
	}
}

func TestClientIntegrationInitCreateShowList(t *testing.T) {
	if _, err := exec.LookPath("bd"); err != nil {
		t.Skipf("bd not installed: %v", err)
	}

	workDir := t.TempDir()
	runInDir(t, workDir, "git", "init", "-q")

	client, err := NewClient(workDir)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	if err := client.Init(); err != nil {
		t.Fatalf("init: %v", err)
	}
	if _, err := os.Stat(filepath.Join(workDir, ".beads")); err != nil {
		t.Fatalf("stat .beads: %v", err)
	}

	rootID, err := client.Create(CreateOpts{
		Title:       "integration-root",
		Type:        "task",
		Description: "root bead",
	})
	if err != nil {
		t.Fatalf("create root: %v", err)
	}

	shown, err := client.Show(rootID)
	if err != nil {
		t.Fatalf("show root: %v", err)
	}
	if shown.ID != rootID {
		t.Fatalf("show id = %q, want %q", shown.ID, rootID)
	}

	if err := client.SetState(rootID, "phase", "red"); err != nil {
		t.Fatalf("set state: %v", err)
	}

	blockerID, err := client.Create(CreateOpts{
		Title:       "integration-blocker",
		Type:        "task",
		Description: "dependency",
	})
	if err != nil {
		t.Fatalf("create blocker: %v", err)
	}

	if err := client.AddDep(rootID, blockerID); err != nil {
		t.Fatalf("add dep: %v", err)
	}

	ready, err := client.Ready()
	if err != nil {
		t.Fatalf("ready: %v", err)
	}
	if containsIssueID(ready, rootID) {
		t.Fatalf("ready list unexpectedly contains blocked issue %q", rootID)
	}

	list, err := client.List(ListOpts{Type: "task", Status: "open", Limit: 100})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if !containsIssueID(list, blockerID) {
		t.Fatalf("open task list missing blocker issue %q", blockerID)
	}

	graph, err := client.Graph(rootID)
	if err != nil {
		t.Fatalf("graph: %v", err)
	}
	if !strings.Contains(graph, rootID) {
		t.Fatalf("graph output missing root id %q", rootID)
	}

	if err := client.AddComment(rootID, "integration comment"); err != nil {
		t.Fatalf("add comment: %v", err)
	}

	agentID, err := client.Create(CreateOpts{
		Title:       "integration-agent",
		Type:        "task",
		Description: "agent bead",
		Labels:      []string{"gt:agent"},
	})
	if err != nil {
		t.Fatalf("create agent: %v", err)
	}
	if err := client.AgentHeartbeat(agentID); err != nil {
		t.Fatalf("agent heartbeat: %v", err)
	}
}

func runInDir(t *testing.T, dir, name string, args ...string) {
	t.Helper()

	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf(
			"run %s %s: %v\noutput:\n%s",
			name,
			strings.Join(args, " "),
			err,
			string(output),
		)
	}
}

func containsIssueID(issues []Bead, issueID string) bool {
	for _, issue := range issues {
		if issue.ID == issueID {
			return true
		}
	}
	return false
}

func containsArgsInOrder(args, orderedSubsequence []string) bool {
	if len(orderedSubsequence) == 0 {
		return true
	}

	index := 0
	for _, arg := range args {
		if arg != orderedSubsequence[index] {
			continue
		}
		index++
		if index == len(orderedSubsequence) {
			return true
		}
	}

	return false
}

func TestDecodeBeadListFromGraphWrapper(t *testing.T) {
	t.Parallel()

	input := []byte(`{"issues":[{"id":"ship-commander-3-1","title":"a","status":"open","priority":2,"issue_type":"task"}]}`)
	issues, err := decodeBeadList(input)
	if err != nil {
		t.Fatalf("decode bead list: %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("issues length = %d, want 1", len(issues))
	}
}

func TestGraphReturnsNormalizedJSON(t *testing.T) {
	t.Parallel()

	workDir := t.TempDir()
	runner := &fakeCommandRunner{
		results: []fakeResult{
			{stdout: []byte(`{"version":"1.0.0"}`)},
			{stdout: []byte(`{"root":{"id":"ship-commander-3-1"}}`)},
		},
	}

	client, err := newClient(workDir, "sh", time.Second, runner)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	out, err := client.Graph("ship-commander-3-1")
	if err != nil {
		t.Fatalf("graph: %v", err)
	}
	if !strings.Contains(out, `"root"`) {
		t.Fatalf("graph output = %q, want root field", out)
	}
}

func TestDecodeSingleBeadRejectsEmptyArray(t *testing.T) {
	t.Parallel()

	_, err := decodeSingleBead([]byte(`[]`))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "show returned no issue") {
		t.Fatalf("error = %v, want empty show result error", err)
	}
}

func TestCreateRejectsMissingTitle(t *testing.T) {
	t.Parallel()

	workDir := t.TempDir()
	runner := &fakeCommandRunner{
		results: []fakeResult{
			{stdout: []byte(`{"version":"1.0.0"}`)},
		},
	}

	client, err := newClient(workDir, "sh", time.Second, runner)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	_, err = client.Create(CreateOpts{})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "title must not be empty") {
		t.Fatalf("error = %v, want title validation error", err)
	}
}

func TestReadyParsesListOutput(t *testing.T) {
	t.Parallel()

	workDir := t.TempDir()
	runner := &fakeCommandRunner{
		results: []fakeResult{
			{stdout: []byte(`{"version":"1.0.0"}`)},
			{stdout: []byte(`[{"id":"ship-commander-3-1","title":"a","status":"open","priority":2,"issue_type":"task"}]`)},
		},
	}

	client, err := newClient(workDir, "sh", time.Second, runner)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	issues, err := client.Ready()
	if err != nil {
		t.Fatalf("ready: %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("issues length = %d, want 1", len(issues))
	}
}

func TestAddDepParsesJSONOutput(t *testing.T) {
	t.Parallel()

	workDir := t.TempDir()
	runner := &fakeCommandRunner{
		results: []fakeResult{
			{stdout: []byte(`{"version":"1.0.0"}`)},
			{stdout: []byte(`{"status":"added"}`)},
		},
	}

	client, err := newClient(workDir, "sh", time.Second, runner)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	if err := client.AddDep("a", "b"); err != nil {
		t.Fatalf("add dep: %v", err)
	}
}

func TestAddCommentParsesJSONOutput(t *testing.T) {
	t.Parallel()

	workDir := t.TempDir()
	runner := &fakeCommandRunner{
		results: []fakeResult{
			{stdout: []byte(`{"version":"1.0.0"}`)},
			{stdout: []byte(`{"id":1}`)},
		},
	}

	client, err := newClient(workDir, "sh", time.Second, runner)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	if err := client.AddComment("ship-commander-3-1", "hello"); err != nil {
		t.Fatalf("add comment: %v", err)
	}
}

func TestAgentHeartbeatParsesJSONOutput(t *testing.T) {
	t.Parallel()

	workDir := t.TempDir()
	runner := &fakeCommandRunner{
		results: []fakeResult{
			{stdout: []byte(`{"version":"1.0.0"}`)},
			{stdout: []byte(`{"agent":"ship-commander-3-agent"}`)},
		},
	}

	client, err := newClient(workDir, "sh", time.Second, runner)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	if err := client.AgentHeartbeat("ship-commander-3-agent"); err != nil {
		t.Fatalf("agent heartbeat: %v", err)
	}
}

func TestDecodeJSONRejectsEmpty(t *testing.T) {
	t.Parallel()

	err := decodeJSON([]byte(""), &map[string]any{})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "empty JSON output") {
		t.Fatalf("error = %v, want empty JSON output error", err)
	}
}

func TestRunAppendsJSONFlagOnce(t *testing.T) {
	t.Parallel()

	workDir := t.TempDir()
	runner := &fakeCommandRunner{
		results: []fakeResult{
			{stdout: []byte(`{"version":"1.0.0"}`)},
			{stdout: []byte(`[]`)},
		},
	}

	client, err := newClient(workDir, "sh", time.Second, runner)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	_, err = client.run("ready", "--json")
	if err != nil {
		t.Fatalf("run: %v", err)
	}

	call := runner.calls[1]
	jsonCount := 0
	for _, arg := range call.args {
		if arg == "--json" {
			jsonCount++
		}
	}
	if jsonCount != 1 {
		t.Fatalf("--json count = %d, want 1 (args: %v)", jsonCount, call.args)
	}
}

func TestListRejectsInvalidJSON(t *testing.T) {
	t.Parallel()

	workDir := t.TempDir()
	runner := &fakeCommandRunner{
		results: []fakeResult{
			{stdout: []byte(`{"version":"1.0.0"}`)},
			{stdout: []byte(`not-json`)},
		},
	}

	client, err := newClient(workDir, "sh", time.Second, runner)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	_, err = client.List(ListOpts{})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "parse list output JSON") {
		t.Fatalf("error = %v, want parse list output context", err)
	}
}

func TestSetStateRejectsEmptyKey(t *testing.T) {
	t.Parallel()

	workDir := t.TempDir()
	runner := &fakeCommandRunner{
		results: []fakeResult{
			{stdout: []byte(`{"version":"1.0.0"}`)},
		},
	}

	client, err := newClient(workDir, "sh", time.Second, runner)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	if err := client.SetState("ship-commander-3-1", "", "x"); err == nil {
		t.Fatal("expected error for empty key")
	}
}

func TestNewClientRejectsInvalidTimeout(t *testing.T) {
	t.Parallel()

	_, err := newClient(t.TempDir(), "sh", 0, &fakeCommandRunner{})
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if !strings.Contains(err.Error(), "timeout must be positive") {
		t.Fatalf("error = %v, want timeout validation error", err)
	}
}

func TestRunWrapsRunnerError(t *testing.T) {
	t.Parallel()

	workDir := t.TempDir()
	runner := &fakeCommandRunner{
		results: []fakeResult{
			{stdout: []byte(`{"version":"1.0.0"}`)},
			{
				stderr: []byte("broken"),
				err:    fmt.Errorf("runner failed"),
			},
		},
	}

	client, err := newClient(workDir, "sh", time.Second, runner)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	_, err = client.run("ready")
	if err == nil {
		t.Fatal("expected run error")
	}
	if !strings.Contains(err.Error(), "run sh ready --json") {
		t.Fatalf("error = %v, want command context", err)
	}
}
