package gates

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	// GateTypeVerifyRED validates RED-phase test-failure behavior.
	GateTypeVerifyRED = "VERIFY_RED"
	// GateTypeVerifyGREEN validates GREEN-phase implementation behavior.
	GateTypeVerifyGREEN = "VERIFY_GREEN"
	// GateTypeVerifyREFACTOR validates REFACTOR-phase behavior preservation.
	GateTypeVerifyREFACTOR = "VERIFY_REFACTOR"
	// GateTypeVerifyIMPLEMENT validates STANDARD_OPS implementation quality gates.
	GateTypeVerifyIMPLEMENT = "VERIFY_IMPLEMENT"
)

const (
	// ClassificationAccept indicates gate acceptance.
	ClassificationAccept = "accept"
	// ClassificationRejectVanity indicates a vanity RED result.
	ClassificationRejectVanity = "reject_vanity"
	// ClassificationRejectSyntax indicates syntax/import failures.
	ClassificationRejectSyntax = "reject_syntax"
	// ClassificationRejectFailure indicates a generic gate failure.
	ClassificationRejectFailure = "reject_failure"
)

const (
	// DefaultTimeout is the deterministic default gate timeout.
	DefaultTimeout = 120 * time.Second
	// DefaultOutputLimitBytes is the deterministic maximum captured output size per gate run.
	DefaultOutputLimitBytes = 1024 * 1024
	// DefaultOutputSnippetBytes is the max OutputSnippet size persisted in GateResult.
	DefaultOutputSnippetBytes = 1024
)

// GateResult captures deterministic gate evidence.
type GateResult struct {
	Type           string
	ExitCode       int
	Classification string
	OutputSnippet  string
	Output         string
	Duration       time.Duration
	Attempt        int
	Timestamp      time.Time
}

// GateRunner executes one verification gate.
type GateRunner interface {
	Run(ctx context.Context, gateType string, workdir string, missionID string) (*GateResult, error)
}

// EvidenceStore persists gate evidence to durable storage.
type EvidenceStore interface {
	RecordGateEvidence(ctx context.Context, missionID string, result GateResult) error
}

// MissionCommandResolver resolves mission-scoped gate command templates.
type MissionCommandResolver interface {
	ResolveGateCommands(ctx context.Context, missionID, gateType string) ([]string, error)
}

// VariableResolver resolves command substitution variables.
type VariableResolver interface {
	ResolveGateVariables(ctx context.Context, missionID string) (map[string]string, error)
}

// RunnerConfig configures deterministic gate execution behavior.
type RunnerConfig struct {
	Timeout            time.Duration
	OutputLimitBytes   int
	OutputSnippetBytes int
	ProjectCommands    map[string][]string
	GreenInfraCommands []string
}

// Runner executes deterministic verification gates with evidence persistence.
type Runner struct {
	executor        commandExecutor
	evidence        EvidenceStore
	missionCommands MissionCommandResolver
	variables       VariableResolver
	timeout         time.Duration
	outputLimit     int
	snippetLimit    int
	projectCommands map[string][]string
	greenInfra      []string
	now             func() time.Time

	mu       sync.Mutex
	attempts map[string]int
}

// NewRunner creates a deterministic verification gate runner.
func NewRunner(
	executor commandExecutor,
	evidence EvidenceStore,
	missionCommands MissionCommandResolver,
	variables VariableResolver,
	config RunnerConfig,
) (*Runner, error) {
	if executor == nil {
		return nil, errors.New("command executor is required")
	}
	if evidence == nil {
		return nil, errors.New("evidence store is required")
	}

	timeout := config.Timeout
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	outputLimit := config.OutputLimitBytes
	if outputLimit <= 0 {
		outputLimit = DefaultOutputLimitBytes
	}
	snippetLimit := config.OutputSnippetBytes
	if snippetLimit <= 0 {
		snippetLimit = DefaultOutputSnippetBytes
	}

	projectCommands := cloneCommandMap(config.ProjectCommands)
	if config.ProjectCommands == nil {
		projectCommands = defaultProjectCommands()
	}

	return &Runner{
		executor:        executor,
		evidence:        evidence,
		missionCommands: missionCommands,
		variables:       variables,
		timeout:         timeout,
		outputLimit:     outputLimit,
		snippetLimit:    snippetLimit,
		projectCommands: projectCommands,
		greenInfra:      append([]string(nil), config.GreenInfraCommands...),
		now:             time.Now,
		attempts:        make(map[string]int),
	}, nil
}

// NewShellRunner creates a gate runner backed by shell command execution.
func NewShellRunner(
	evidence EvidenceStore,
	missionCommands MissionCommandResolver,
	variables VariableResolver,
	config RunnerConfig,
) (*Runner, error) {
	return NewRunner(shellExecutor{}, evidence, missionCommands, variables, config)
}

// Run executes one verification gate and persists evidence.
func (r *Runner) Run(ctx context.Context, gateType string, workdir string, missionID string) (*GateResult, error) {
	if r == nil {
		return nil, errors.New("runner is nil")
	}
	gateType = strings.TrimSpace(gateType)
	workdir = strings.TrimSpace(workdir)
	missionID = strings.TrimSpace(missionID)
	if gateType == "" {
		return nil, errors.New("gate type must not be empty")
	}
	if workdir == "" {
		return nil, errors.New("workdir must not be empty")
	}
	if missionID == "" {
		return nil, errors.New("mission id must not be empty")
	}

	commands, err := r.resolveCommands(ctx, gateType, missionID)
	if err != nil {
		return nil, err
	}

	vars := map[string]string{
		"{mission_id}":   missionID,
		"{worktree_dir}": workdir,
		"{test_file}":    "",
	}
	if r.variables != nil {
		extra, resolveErr := r.variables.ResolveGateVariables(ctx, missionID)
		if resolveErr != nil {
			return nil, fmt.Errorf("resolve gate variables: %w", resolveErr)
		}
		for key, value := range extra {
			normalized := normalizeVariableKey(key)
			vars[normalized] = value
		}
	}

	substituted := substituteAll(commands, vars)
	attempt := r.nextAttempt(gateType, missionID)
	start := r.now().UTC()

	result, runErr := r.executeGate(ctx, gateType, workdir, substituted)
	result.Type = gateType
	result.Attempt = attempt
	result.Timestamp = start
	result.OutputSnippet = snippetForOutput(result.Output, r.snippetLimit, gateType)

	if persistErr := r.evidence.RecordGateEvidence(ctx, missionID, result); persistErr != nil {
		return nil, fmt.Errorf("record gate evidence: %w", persistErr)
	}
	if runErr != nil {
		return nil, runErr
	}

	return &result, nil
}

func (r *Runner) resolveCommands(ctx context.Context, gateType, missionID string) ([]string, error) {
	if r.missionCommands != nil {
		commands, err := r.missionCommands.ResolveGateCommands(ctx, missionID, gateType)
		if err != nil {
			return nil, fmt.Errorf("resolve mission gate commands: %w", err)
		}
		if len(commands) > 0 {
			return dedupeNonEmpty(commands), nil
		}
	}

	commands := dedupeNonEmpty(r.projectCommands[gateType])
	if len(commands) > 0 {
		return commands, nil
	}
	if gateType == GateTypeVerifyIMPLEMENT {
		return []string{}, nil
	}
	return nil, fmt.Errorf("no commands configured for gate type %s", gateType)
}

func (r *Runner) executeGate(ctx context.Context, gateType, workdir string, commands []string) (GateResult, error) {
	switch gateType {
	case GateTypeVerifyRED:
		return r.executeVerifyRED(ctx, workdir, commands)
	case GateTypeVerifyGREEN:
		return r.executeVerifyGREEN(ctx, workdir, commands)
	case GateTypeVerifyREFACTOR:
		return r.executeVerifyREFACTOR(ctx, workdir, commands)
	case GateTypeVerifyIMPLEMENT:
		return r.executeVerifyIMPLEMENT(ctx, workdir, commands)
	default:
		return GateResult{}, fmt.Errorf("unsupported gate type %q", gateType)
	}
}

func (r *Runner) executeVerifyRED(ctx context.Context, workdir string, commands []string) (GateResult, error) {
	exitCode, output, duration, err := r.runSequential(ctx, workdir, commands)
	if err != nil {
		return GateResult{}, err
	}

	classification := ClassificationRejectFailure
	switch {
	case hasSyntaxError(output):
		classification = ClassificationRejectSyntax
	case exitCode == 0:
		classification = ClassificationRejectVanity
	case hasTestFailure(output):
		classification = ClassificationAccept
	}

	return GateResult{
		ExitCode:       exitCode,
		Classification: classification,
		Output:         output,
		Duration:       duration,
	}, nil
}

func (r *Runner) executeVerifyGREEN(ctx context.Context, workdir string, commands []string) (GateResult, error) {
	exitCode, output, duration, err := r.runSequential(ctx, workdir, commands)
	if err != nil {
		return GateResult{}, err
	}
	if exitCode != 0 {
		return GateResult{
			ExitCode:       exitCode,
			Classification: ClassificationRejectFailure,
			Output:         output,
			Duration:       duration,
		}, nil
	}

	infraOutput := newLimitedBuffer(r.outputLimit)
	infraExitCode, infraDuration, infraErr := r.runInfraConsistency(ctx, workdir, infraOutput)
	duration += infraDuration
	output = mergeOutput(output, infraOutput.String(), r.outputLimit)
	if infraErr != nil {
		return GateResult{}, infraErr
	}
	if infraExitCode != 0 {
		return GateResult{
			ExitCode:       infraExitCode,
			Classification: ClassificationRejectFailure,
			Output:         output,
			Duration:       duration,
		}, nil
	}

	return GateResult{
		ExitCode:       0,
		Classification: ClassificationAccept,
		Output:         output,
		Duration:       duration,
	}, nil
}

func (r *Runner) executeVerifyREFACTOR(ctx context.Context, workdir string, commands []string) (GateResult, error) {
	exitCode, output, duration, err := r.runSequential(ctx, workdir, commands)
	if err != nil {
		return GateResult{}, err
	}

	classification := ClassificationAccept
	if exitCode != 0 {
		classification = ClassificationRejectFailure
	}
	return GateResult{
		ExitCode:       exitCode,
		Classification: classification,
		Output:         output,
		Duration:       duration,
	}, nil
}

func (r *Runner) executeVerifyIMPLEMENT(ctx context.Context, workdir string, commands []string) (GateResult, error) {
	if len(commands) == 0 {
		return GateResult{
			ExitCode:       0,
			Classification: ClassificationAccept,
			Output:         "no VERIFY_IMPLEMENT commands configured",
			Duration:       0,
		}, nil
	}

	exitCode, output, duration, err := r.runSequential(ctx, workdir, commands)
	if err != nil {
		return GateResult{}, err
	}
	classification := ClassificationAccept
	if exitCode != 0 {
		classification = ClassificationRejectFailure
	}
	return GateResult{
		ExitCode:       exitCode,
		Classification: classification,
		Output:         output,
		Duration:       duration,
	}, nil
}

func (r *Runner) runInfraConsistency(
	ctx context.Context,
	workdir string,
	output *limitedBuffer,
) (int, time.Duration, error) {
	commands := dedupeNonEmpty(r.greenInfra)
	if len(commands) == 0 {
		return 0, 0, nil
	}

	var duration time.Duration
	failures := 0
	lastExitCode := 0
	for _, command := range commands {
		for attempt := 0; attempt < 3; attempt++ {
			result, err := r.executor.Run(ctx, workdir, command, r.timeout, r.outputLimit)
			if err != nil {
				return 0, 0, err
			}
			duration += result.Duration
			output.WriteString(fmt.Sprintf("infra(%d/%d): %s\n", attempt+1, 3, command))
			output.WriteString(result.Output)
			lastExitCode = result.ExitCode
			if result.ExitCode != 0 {
				failures++
			}
		}
	}
	if failures > 0 {
		if strings.TrimSpace(output.String()) == "" {
			output.WriteString("infra test flaky")
		}
		if lastExitCode == 0 {
			lastExitCode = 1
		}
		return lastExitCode, duration, nil
	}
	return 0, duration, nil
}

func (r *Runner) runSequential(
	ctx context.Context,
	workdir string,
	commands []string,
) (int, string, time.Duration, error) {
	output := newLimitedBuffer(r.outputLimit)
	var totalDuration time.Duration

	lastExitCode := 0
	for _, command := range commands {
		result, err := r.executor.Run(ctx, workdir, command, r.timeout, r.outputLimit)
		if err != nil {
			return 0, "", 0, err
		}
		totalDuration += result.Duration
		lastExitCode = result.ExitCode
		output.WriteString(result.Output)
		if strings.TrimSpace(result.Output) != "" {
			output.WriteString("\n")
		}
		if result.ExitCode != 0 {
			break
		}
	}
	return lastExitCode, strings.TrimSpace(output.String()), totalDuration, nil
}

func (r *Runner) nextAttempt(gateType, missionID string) int {
	key := missionID + "|" + gateType
	r.mu.Lock()
	defer r.mu.Unlock()

	r.attempts[key]++
	return r.attempts[key]
}

func substituteAll(commands []string, vars map[string]string) []string {
	out := make([]string, 0, len(commands))
	keys := make([]string, 0, len(vars))
	for key := range vars {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, command := range commands {
		replaced := command
		for _, key := range keys {
			replaced = strings.ReplaceAll(replaced, key, vars[key])
		}
		out = append(out, replaced)
	}
	return out
}

func normalizeVariableKey(key string) string {
	key = strings.TrimSpace(key)
	if key == "" {
		return key
	}
	if strings.HasPrefix(key, "{") && strings.HasSuffix(key, "}") {
		return key
	}
	return "{" + key + "}"
}

func snippetForOutput(output string, limit int, gateType string) string {
	output = strings.TrimSpace(output)
	if output == "" {
		return "no gate output"
	}

	if gateType == GateTypeVerifyGREEN || gateType == GateTypeVerifyREFACTOR || gateType == GateTypeVerifyIMPLEMENT {
		if failure := firstFailureLine(output); failure != "" {
			return trimToLimit(failure, limit)
		}
	}
	return trimToLimit(output, limit)
}

func trimToLimit(value string, limit int) string {
	if limit <= 0 {
		return value
	}
	if len(value) <= limit {
		return value
	}
	if limit <= 3 {
		return value[:limit]
	}
	return value[:limit-3] + "..."
}

func firstFailureLine(output string) string {
	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.Contains(trimmed, "--- FAIL:") ||
			strings.HasPrefix(trimmed, "FAIL") ||
			strings.Contains(trimmed, "panic:") {
			return trimmed
		}
	}
	return ""
}

func hasSyntaxError(output string) bool {
	lower := strings.ToLower(output)
	patterns := []string{
		"syntax error",
		"import cycle",
		"cannot find package",
		"undefined:",
		"expected ';'",
	}
	for _, pattern := range patterns {
		if strings.Contains(lower, pattern) {
			return true
		}
	}
	return false
}

func hasTestFailure(output string) bool {
	if strings.Contains(output, "--- FAIL:") {
		return true
	}
	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "FAIL") {
			return true
		}
	}
	return false
}

func cloneCommandMap(in map[string][]string) map[string][]string {
	if len(in) == 0 {
		return map[string][]string{}
	}
	out := make(map[string][]string, len(in))
	for key, commands := range in {
		out[key] = append([]string(nil), commands...)
	}
	return out
}

func dedupeNonEmpty(commands []string) []string {
	out := make([]string, 0, len(commands))
	seen := make(map[string]struct{}, len(commands))
	for _, command := range commands {
		command = strings.TrimSpace(command)
		if command == "" {
			continue
		}
		if _, exists := seen[command]; exists {
			continue
		}
		seen[command] = struct{}{}
		out = append(out, command)
	}
	return out
}

func mergeOutput(base, extra string, limit int) string {
	buffer := newLimitedBuffer(limit)
	buffer.WriteString(base)
	if strings.TrimSpace(base) != "" && strings.TrimSpace(extra) != "" {
		buffer.WriteString("\n")
	}
	buffer.WriteString(extra)
	return strings.TrimSpace(buffer.String())
}

func defaultProjectCommands() map[string][]string {
	return map[string][]string{
		GateTypeVerifyRED:       {"go test {test_file}"},
		GateTypeVerifyGREEN:     {"go test ./..."},
		GateTypeVerifyREFACTOR:  {"go test ./..."},
		GateTypeVerifyIMPLEMENT: {},
	}
}
