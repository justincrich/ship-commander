package recovery

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

const (
	entityTypeCommission = "commission"
	entityTypeMission    = "mission"
	entityTypeAgent      = "agent"
)

// CommandRunner executes shell commands for recovery stores.
type CommandRunner interface {
	Run(ctx context.Context, name string, args ...string) ([]byte, error)
}

type defaultCommandRunner struct{}

func (defaultCommandRunner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("run %s %s: %w (%s)", name, strings.Join(args, " "), err, strings.TrimSpace(string(output)))
	}
	return output, nil
}

// BeadsStore reads commissions, missions, and agents from Beads and persists recovery updates.
type BeadsStore struct {
	runner CommandRunner
}

// NewBeadsStore creates a Beads-backed recovery store.
func NewBeadsStore() (*BeadsStore, error) {
	return NewBeadsStoreWithRunner(defaultCommandRunner{})
}

// NewBeadsStoreWithRunner creates a Beads-backed recovery store with a custom runner.
func NewBeadsStoreWithRunner(runner CommandRunner) (*BeadsStore, error) {
	if runner == nil {
		return nil, errors.New("runner must not be nil")
	}
	return &BeadsStore{runner: runner}, nil
}

// LoadSnapshot reads all persisted commissions, missions, and agents from Beads.
func (s *BeadsStore) LoadSnapshot(ctx context.Context) (Snapshot, error) {
	if s == nil {
		return Snapshot{}, errors.New("beads store is nil")
	}
	out, err := s.runner.Run(ctx, "bd", "list", "--json")
	if err != nil {
		return Snapshot{}, fmt.Errorf("list beads issues: %w", err)
	}

	var issues []beadsIssue
	if err := json.Unmarshal(out, &issues); err != nil {
		return Snapshot{}, fmt.Errorf("parse beads list JSON: %w", err)
	}

	snapshot := Snapshot{}
	for _, issue := range issues {
		switch normalizedEntityType(issue) {
		case entityTypeCommission:
			snapshot.Commissions = append(snapshot.Commissions, Commission{
				ID:    strings.TrimSpace(issue.ID),
				State: extractState(issue, "commission_state"),
			})
		case entityTypeMission:
			snapshot.Missions = append(snapshot.Missions, Mission{
				ID:           strings.TrimSpace(issue.ID),
				CommissionID: extractCommissionID(issue),
				State:        extractState(issue, "mission_state"),
				AgentID:      extractAgentID(issue),
			})
		case entityTypeAgent:
			snapshot.Agents = append(snapshot.Agents, Agent{
				ID:        strings.TrimSpace(issue.ID),
				State:     extractState(issue, "agent_state"),
				SessionID: extractSessionID(issue),
			})
		}
	}

	return snapshot, nil
}

// SetMissionBacklog sets a mission bead's deterministic state to backlog.
func (s *BeadsStore) SetMissionBacklog(ctx context.Context, missionID string) error {
	if s == nil {
		return errors.New("beads store is nil")
	}
	missionID = strings.TrimSpace(missionID)
	if missionID == "" {
		return errors.New("mission id must not be empty")
	}
	if _, err := s.runner.Run(ctx, "bd", "set-state", missionID, "mission_state="+MissionBacklog, "--json"); err != nil {
		return fmt.Errorf("set mission %s backlog state: %w", missionID, err)
	}
	return nil
}

// SetAgentDead sets an agent bead's deterministic state to dead.
func (s *BeadsStore) SetAgentDead(ctx context.Context, agentID string) error {
	if s == nil {
		return errors.New("beads store is nil")
	}
	agentID = strings.TrimSpace(agentID)
	if agentID == "" {
		return errors.New("agent id must not be empty")
	}
	if _, err := s.runner.Run(ctx, "bd", "set-state", agentID, "agent_state="+AgentDead, "--json"); err != nil {
		return fmt.Errorf("set agent %s dead state: %w", agentID, err)
	}
	return nil
}

type beadsIssue struct {
	ID           string            `json:"id"`
	IssueType    string            `json:"issue_type"`
	Status       string            `json:"status"`
	Parent       string            `json:"parent"`
	Labels       []string          `json:"labels"`
	State        map[string]any    `json:"state"`
	Dependencies []beadsDependency `json:"dependencies"`
}

type beadsDependency struct {
	IssueID      string `json:"issue_id"`
	DependsOnID  string `json:"depends_on_id"`
	Type         string `json:"type"`
	DependencyID string `json:"id"`
}

func normalizedEntityType(issue beadsIssue) string {
	if kind := strings.ToLower(strings.TrimSpace(issue.IssueType)); kind != "" {
		return kind
	}
	for _, label := range issue.Labels {
		normalized := strings.ToLower(strings.TrimSpace(label))
		switch normalized {
		case entityTypeCommission, "type:commission":
			return entityTypeCommission
		case entityTypeMission, "type:mission":
			return entityTypeMission
		case entityTypeAgent, "type:agent":
			return entityTypeAgent
		}
	}
	return ""
}

func extractState(issue beadsIssue, key string) string {
	if issue.State != nil {
		if raw, ok := issue.State[key]; ok {
			if text := valueToString(raw); text != "" {
				return strings.ToLower(text)
			}
		}
		if raw, ok := issue.State["status"]; ok {
			if text := valueToString(raw); text != "" {
				return strings.ToLower(text)
			}
		}
	}
	return strings.ToLower(strings.TrimSpace(issue.Status))
}

func extractCommissionID(issue beadsIssue) string {
	if parent := strings.TrimSpace(issue.Parent); parent != "" {
		return parent
	}
	for _, dep := range issue.Dependencies {
		if strings.EqualFold(strings.TrimSpace(dep.Type), "parent-child") {
			if strings.TrimSpace(dep.DependsOnID) != "" {
				return strings.TrimSpace(dep.DependsOnID)
			}
			if strings.TrimSpace(dep.DependencyID) != "" {
				return strings.TrimSpace(dep.DependencyID)
			}
		}
	}
	return ""
}

func extractAgentID(issue beadsIssue) string {
	if issue.State != nil {
		if raw, ok := issue.State["agent_id"]; ok {
			if value := valueToString(raw); value != "" {
				return value
			}
		}
		if raw, ok := issue.State["agentId"]; ok {
			if value := valueToString(raw); value != "" {
				return value
			}
		}
	}
	for _, label := range issue.Labels {
		if agentID, ok := extractLabelValue(label, "agent:"); ok {
			return agentID
		}
	}
	return ""
}

func extractSessionID(issue beadsIssue) string {
	if issue.State != nil {
		if raw, ok := issue.State["session_id"]; ok {
			if value := valueToString(raw); value != "" {
				return value
			}
		}
		if raw, ok := issue.State["sessionId"]; ok {
			if value := valueToString(raw); value != "" {
				return value
			}
		}
	}
	for _, label := range issue.Labels {
		if sessionID, ok := extractLabelValue(label, "session:"); ok {
			return sessionID
		}
	}
	return ""
}

func valueToString(value any) string {
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case fmt.Stringer:
		return strings.TrimSpace(typed.String())
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", typed))
	}
}

func extractLabelValue(label, prefix string) (string, bool) {
	normalized := strings.TrimSpace(label)
	if normalized == "" {
		return "", false
	}
	if !strings.HasPrefix(strings.ToLower(normalized), strings.ToLower(prefix)) {
		return "", false
	}
	value := strings.TrimSpace(normalized[len(prefix):])
	if value == "" {
		return "", false
	}
	return value, true
}
