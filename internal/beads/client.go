package beads

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	defaultCommand = "bd"
	defaultTimeout = 30 * time.Second
)

type commandRunner interface {
	Run(ctx context.Context, dir string, name string, args ...string) ([]byte, []byte, error)
}

type defaultCommandRunner struct{}

func (defaultCommandRunner) Run(ctx context.Context, dir string, name string, args ...string) ([]byte, []byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	return stdout.Bytes(), stderr.Bytes(), err
}

// Client wraps the `bd` CLI and returns typed results.
type Client struct {
	workDir string
	command string
	timeout time.Duration
	runner  commandRunner
}

// NewClient creates a Beads client rooted at workDir and validates bd availability.
func NewClient(workDir string) (*Client, error) {
	return newClient(workDir, defaultCommand, defaultTimeout, defaultCommandRunner{})
}

func newClient(workDir, command string, timeout time.Duration, runner commandRunner) (*Client, error) {
	if strings.TrimSpace(workDir) == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("resolve working directory: %w", err)
		}
		workDir = cwd
	}
	if strings.TrimSpace(command) == "" {
		return nil, errors.New("command must not be empty")
	}
	if timeout <= 0 {
		return nil, errors.New("timeout must be positive")
	}
	if runner == nil {
		return nil, errors.New("runner must not be nil")
	}

	client := &Client{
		workDir: workDir,
		command: command,
		timeout: timeout,
		runner:  runner,
	}

	if err := client.checkCLI(); err != nil {
		return nil, err
	}

	return client, nil
}

func (c *Client) checkCLI() error {
	if _, err := exec.LookPath(c.command); err != nil {
		return fmt.Errorf("find %s on PATH: %w", c.command, err)
	}

	_, err := c.run("version")
	if err != nil {
		return fmt.Errorf("check %s availability: %w", c.command, err)
	}

	return nil
}

// Init initializes Beads with `bd init` when `.beads/` does not exist.
func (c *Client) Init() error {
	beadsDir := filepath.Join(c.workDir, ".beads")
	_, err := os.Stat(beadsDir)
	switch {
	case err == nil:
		return nil
	case errors.Is(err, os.ErrNotExist):
		_, runErr := c.run("init")
		if runErr != nil {
			return fmt.Errorf("initialize beads: %w", runErr)
		}
		return nil
	default:
		return fmt.Errorf("stat beads directory %q: %w", beadsDir, err)
	}
}

// Create creates a Beads issue and returns the new bead ID.
func (c *Client) Create(opts CreateOpts) (string, error) {
	title := strings.TrimSpace(opts.Title)
	if title == "" {
		return "", errors.New("create title must not be empty")
	}

	issueType := strings.TrimSpace(opts.Type)
	if issueType == "" {
		issueType = "task"
	}

	args := []string{
		"create",
		"--type", issueType,
		"--title", title,
	}
	if strings.TrimSpace(opts.Description) != "" {
		args = append(args, "--description", opts.Description)
	}
	if opts.Parent != nil && strings.TrimSpace(*opts.Parent) != "" {
		args = append(args, "--parent", strings.TrimSpace(*opts.Parent))
	}
	if len(opts.Labels) > 0 {
		args = append(args, "--labels", strings.Join(opts.Labels, ","))
	}
	if strings.TrimSpace(opts.Priority) != "" {
		args = append(args, "--priority", strings.TrimSpace(opts.Priority))
	}

	out, err := c.run(args...)
	if err != nil {
		return "", fmt.Errorf("create bead: %w", err)
	}

	var created Bead
	if err := decodeJSON(out, &created); err != nil {
		return "", fmt.Errorf("parse create output JSON: %w", err)
	}
	if strings.TrimSpace(created.ID) == "" {
		return "", errors.New("create output missing id")
	}

	return created.ID, nil
}

// Show returns one bead by ID.
func (c *Client) Show(id string) (*Bead, error) {
	if strings.TrimSpace(id) == "" {
		return nil, errors.New("issue id must not be empty")
	}

	out, err := c.run("show", id)
	if err != nil {
		return nil, fmt.Errorf("show bead %q: %w", id, err)
	}

	bead, err := decodeSingleBead(out)
	if err != nil {
		return nil, fmt.Errorf("parse show output JSON: %w", err)
	}
	return bead, nil
}

// List returns issues with optional filtering.
func (c *Client) List(opts ListOpts) ([]Bead, error) {
	args := []string{"list"}
	if strings.TrimSpace(opts.Type) != "" {
		args = append(args, "--type", strings.TrimSpace(opts.Type))
	}
	if strings.TrimSpace(opts.Status) != "" {
		args = append(args, "--status", strings.TrimSpace(opts.Status))
	}
	if opts.Limit > 0 {
		args = append(args, "--limit", fmt.Sprintf("%d", opts.Limit))
	}
	if strings.TrimSpace(opts.Parent) != "" {
		args = append(args, "--parent", strings.TrimSpace(opts.Parent))
	}
	for _, label := range opts.Labels {
		if strings.TrimSpace(label) == "" {
			continue
		}
		args = append(args, "--label", strings.TrimSpace(label))
	}

	out, err := c.run(args...)
	if err != nil {
		return nil, fmt.Errorf("list beads: %w", err)
	}

	issues, err := decodeBeadList(out)
	if err != nil {
		return nil, fmt.Errorf("parse list output JSON: %w", err)
	}
	return issues, nil
}

// SetState sets one state dimension on an issue.
func (c *Client) SetState(id, key, value string) error {
	if strings.TrimSpace(id) == "" {
		return errors.New("issue id must not be empty")
	}
	if strings.TrimSpace(key) == "" {
		return errors.New("state key must not be empty")
	}

	out, err := c.run("set-state", id, fmt.Sprintf("%s=%s", key, value))
	if err != nil {
		return fmt.Errorf("set state %q on %q: %w", key, id, err)
	}
	if err := decodeJSON(out, &map[string]any{}); err != nil {
		return fmt.Errorf("parse set-state output JSON: %w", err)
	}
	return nil
}

// AddDep adds a dependency edge `childID -> parentID`.
func (c *Client) AddDep(childID, parentID string) error {
	if strings.TrimSpace(childID) == "" {
		return errors.New("child issue id must not be empty")
	}
	if strings.TrimSpace(parentID) == "" {
		return errors.New("parent issue id must not be empty")
	}

	out, err := c.run("dep", "add", childID, parentID)
	if err != nil {
		return fmt.Errorf("add dependency %q -> %q: %w", childID, parentID, err)
	}
	if err := decodeJSON(out, &map[string]any{}); err != nil {
		return fmt.Errorf("parse dep output JSON: %w", err)
	}
	return nil
}

// Ready returns currently ready issues from Beads.
func (c *Client) Ready() ([]Bead, error) {
	out, err := c.run("ready")
	if err != nil {
		return nil, fmt.Errorf("query ready beads: %w", err)
	}

	issues, err := decodeBeadList(out)
	if err != nil {
		return nil, fmt.Errorf("parse ready output JSON: %w", err)
	}
	return issues, nil
}

// Graph returns the dependency graph JSON for one issue.
func (c *Client) Graph(id string) (string, error) {
	if strings.TrimSpace(id) == "" {
		return "", errors.New("issue id must not be empty")
	}

	out, err := c.run("graph", id)
	if err != nil {
		return "", fmt.Errorf("graph issue %q: %w", id, err)
	}

	var payload map[string]any
	if err := decodeJSON(out, &payload); err != nil {
		return "", fmt.Errorf("parse graph output JSON: %w", err)
	}

	normalized, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal graph JSON: %w", err)
	}

	return string(normalized), nil
}

// AddComment appends a comment to an issue.
func (c *Client) AddComment(id, comment string) error {
	if strings.TrimSpace(id) == "" {
		return errors.New("issue id must not be empty")
	}
	if strings.TrimSpace(comment) == "" {
		return errors.New("comment must not be empty")
	}

	out, err := c.run("comments", "add", id, comment)
	if err != nil {
		return fmt.Errorf("add comment to %q: %w", id, err)
	}
	if err := decodeJSON(out, &map[string]any{}); err != nil {
		return fmt.Errorf("parse comments output JSON: %w", err)
	}
	return nil
}

// AgentHeartbeat updates the heartbeat timestamp for an agent bead.
func (c *Client) AgentHeartbeat(id string) error {
	if strings.TrimSpace(id) == "" {
		return errors.New("agent id must not be empty")
	}

	out, err := c.run("agent", "heartbeat", id)
	if err != nil {
		return fmt.Errorf("send agent heartbeat for %q: %w", id, err)
	}
	if err := decodeJSON(out, &map[string]any{}); err != nil {
		return fmt.Errorf("parse heartbeat output JSON: %w", err)
	}
	return nil
}

func (c *Client) run(args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	commandArgs := append([]string{}, args...)
	if !hasJSONFlag(commandArgs) {
		commandArgs = append(commandArgs, "--json")
	}

	stdout, stderr, err := c.runner.Run(ctx, c.workDir, c.command, commandArgs...)
	if err != nil {
		return nil, fmt.Errorf(
			"run %s %s: %w (stderr: %s)",
			c.command,
			strings.Join(commandArgs, " "),
			err,
			strings.TrimSpace(string(stderr)),
		)
	}

	return bytes.TrimSpace(stdout), nil
}

func hasJSONFlag(args []string) bool {
	for _, arg := range args {
		if arg == "--json" {
			return true
		}
	}
	return false
}

func decodeJSON(data []byte, out any) error {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return errors.New("empty JSON output")
	}
	if err := json.Unmarshal(trimmed, out); err != nil {
		return fmt.Errorf("unmarshal JSON: %w", err)
	}
	return nil
}

func decodeSingleBead(data []byte) (*Bead, error) {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return nil, errors.New("empty show output")
	}

	switch trimmed[0] {
	case '[':
		var items []Bead
		if err := decodeJSON(trimmed, &items); err != nil {
			return nil, err
		}
		if len(items) == 0 {
			return nil, errors.New("show returned no issue")
		}
		return &items[0], nil
	case '{':
		var item Bead
		if err := decodeJSON(trimmed, &item); err != nil {
			return nil, err
		}
		if strings.TrimSpace(item.ID) == "" {
			return nil, errors.New("show output missing id")
		}
		return &item, nil
	default:
		return nil, errors.New("unexpected show output format")
	}
}

func decodeBeadList(data []byte) ([]Bead, error) {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return nil, errors.New("empty list output")
	}

	switch trimmed[0] {
	case '[':
		var items []Bead
		if err := decodeJSON(trimmed, &items); err != nil {
			return nil, err
		}
		return items, nil
	case '{':
		var wrapped struct {
			Issues []Bead `json:"issues"`
		}
		if err := decodeJSON(trimmed, &wrapped); err != nil {
			return nil, err
		}
		return wrapped.Issues, nil
	default:
		return nil, errors.New("unexpected list output format")
	}
}
