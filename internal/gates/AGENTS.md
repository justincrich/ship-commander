# Verification Gate Engine - Agent Guidance

## Purpose

This package implements the **deterministic verification gate engine**—a core non-negotiable principle of Ship Commander 3. Gates are 100% deterministic: shell commands + exit code checking with NO AI involvement.

## Context in Ship Commander 3

Verification gates are the **boundary between probabilistic agents and deterministic authority**:
- Agent posts claim → Commander runs independent verification → Commander transitions state
- Gates NEVER use AI—only shell commands and exit code analysis
- Gates produce evidence (exit code, output snippet, classification, timestamp)

## Gate Types

| Gate | Purpose | Typical Command | Exit Code Meaning |
|------|---------|-----------------|-------------------|
| **VERIFY_RED** | Confirm test fails before fix | `go test -v {test_file}` | Non-zero = red confirmed |
| **VERIFY_GREEN** | Confirm test passes after fix | `go test -v {test_file}` | Zero = green confirmed |
| **VERIFY_REFACTOR** | Confirm behavior preserved | `go test ./...` | Zero = refactor safe |
| **VERIFY_IMPLEMENT** | Validate STANDARD_OPS work | `go vet ./... && golangci-lint run` | Zero = passes |

## Package Organization

```
internal/gates/
├── AGENTS.md          # This file
├── gate.go            # Gate interface and types
├── verify_red.go      # RED phase gate implementation
├── verify_green.go    # GREEN phase gate implementation
├── verify_refactor.go # REFACTOR phase gate implementation
├── verify_implement.go # STANDARD_OPS gate implementation
├── executor.go        # Gate execution engine (shell commands)
└── evidence.go         # Gate evidence recording
```

## Key Responsibilities

### 1. Deterministic Gate Execution
- Run shell commands in mission worktree
- Capture stdout/stderr (max 1MB default)
- Check exit code
- Parse output for patterns

### 2. Output Classification
Classify gate results based on exit code and output:
- **accept**: Gate passed (test failed as expected, test passed, etc.)
- **reject_vanity**: Test passed without implementation (vanity test)
- **reject_syntax**: Syntax/import error in output
- **reject_failure**: Test failed unexpectedly

### 3. Evidence Recording
Persist gate results to Beads for audit trail:
- Exit code
- Classification
- Output snippet (first 1KB)
- Timestamp
- Attempt number

### 4. Variable Substitution
Support variable substitution in gate commands:
- `{mission_id}` - Mission identifier
- `{worktree_dir}` - Path to mission worktree
- `{test_file}` - Path to test file
- `{test_dir}` - Directory containing test file
- `{ac_index}` - Acceptance Criteria index
- `{ac_title}` - URL-safe version of AC title
- `{surface_area}` - Comma-separated glob patterns
- `{project_root}` - Original project root

## Dependencies

### Internal Dependencies
- `internal/beads` - Persist gate evidence
- `internal/protocol` - Publish gate events
- `internal/config` - Gate command configuration

### External Dependencies
- `os/exec` - Shell command execution
- `context` - Timeout and cancellation
- `time` - Timeout handling

## Coding Standards

Follow Go standards from `.spec/research-findings/GO_CODING_STANDARDS.md`:

### Error Handling
- Never panic in gate execution
- Return structured errors with classification
- Wrap shell command errors with context

### Command Execution
```go
// ✅ GOOD: Command execution with timeout
func ExecuteGate(ctx context.Context, workdir, command string) (*GateResult, error) {
    cmd := exec.CommandContext(ctx, "sh", "-c", command)
    cmd.Dir = workdir

    // Create timeout context
    ctx, cancel := context.WithTimeout(ctx, 120*time.Second)
    defer cancel()

    output, err := cmd.CombinedOutput()
    if err != nil {
        exitCode := cmd.ProcessState.ExitCode()
        return &GateResult{
            ExitCode: exitCode,
            Output:   string(output),
            Success:  false,
        }, nil
    }

    return &GateResult{
        ExitCode: 0,
        Output:   string(output),
        Success:  true,
    }, nil
}
```

### Output Classification
```go
func ClassifyRedPhase(result *GateResult) string {
    // Exit 0 without implementation = vanity test
    if result.ExitCode == 0 {
        if hasSyntaxErrors(result.Output) {
            return "reject_syntax"
        }
        return "reject_vanity"
    }

    // Non-zero with test failure = expected (test is red)
    if hasTestFailurePattern(result.Output) {
        return "accept"
    }

    // Non-zero without clear failure = unexpected
    return "reject_failure"
}
```

## Common Patterns

### Gate Interface
```go
type Gate interface {
    Execute(ctx context.Context, missionID, acID string) (*GateResult, error)
    Classify(result *GateResult) string
}

type GateResult struct {
    GateID       string
    MissionID    string
    Phase        string // "VERIFY_RED" | "VERIFY_GREEN" | etc.
    Command      string // Command executed (after substitution)
    ExitCode     int
    Output       string // Captured stdout/stderr
    Success      bool
    Classification string // "accept" | "reject_vanity" | "reject_syntax" | "reject_failure"
    Timestamp    time.Time
    Attempt      int
}
```

### VERIFY_RED Gate
```go
type VerifyRedGate struct {
    beads    *beads.Client
    config   *config.Config
}

func (g *VerifyRedGate) Execute(ctx context.Context, missionID, acID string) (*GateResult, error) {
    // 1. Build command with variable substitution
    testFile := filepath.Join(worktree, acTestFile(acID))
    command := strings.ReplaceAll(g.config.TestCommand, "{test_file}", testFile)

    // 2. Execute gate
    result, err := gates.ExecuteGate(ctx, worktree, command)
    if err != nil {
        return nil, fmt.Errorf("execute gate: %w", err)
    }

    // 3. Classify result
    result.Classification = g.Classify(result)

    // 4. Record evidence
    if err := g.beads.RecordGateEvidence(ctx, result); err != nil {
        return nil, fmt.Errorf("record evidence: %w", err)
    }

    return result, nil
}

func (g *VerifyRedGate) Classify(result *GateResult) string {
    // Exit 0 = test passed (vanity if no implementation exists)
    if result.ExitCode == 0 {
        if hasSyntaxErrors(result.Output) {
            return "reject_syntax"
        }
        if !hasImplementation(result.Output) {
            return "reject_vanity"
        }
    }

    // Non-zero = test failed (expected for RED phase)
    if hasTestFailure(result.Output) {
        return "accept"
    }

    return "reject_failure"
}
```

### Variable Substitution
```go
type GateContext struct {
    MissionID    string
    WorktreeDir  string
    TestFile     string
    TestDir      string
    ACIndex      int
    ACTitle      string
    SurfaceArea  []string
    ProjectRoot  string
}

func SubstituteVariables(command string, ctx GateContext) string {
    replacements := map[string]string{
        "{mission_id}":   ctx.MissionID,
        "{worktree_dir}":  ctx.WorktreeDir,
        "{test_file}":     ctx.TestFile,
        "{test_dir}":      ctx.TestDir,
        "{ac_index}":      fmt.Sprintf("AC-%d", ctx.ACIndex),
        "{ac_title}":      slugify(ctx.ACTitle),
        "{surface_area}":  strings.Join(ctx.SurfaceArea, ","),
        "{project_root}":  ctx.ProjectRoot,
    }

    for varName, value := range replacements {
        command = strings.ReplaceAll(command, varName, value)
    }

    return command
}
```

### Gate Evidence Recording
```go
func (g *VerifyRedGate) RecordEvidence(ctx context.Context, result *GateResult) error {
    evidence := map[string]interface{}{
        "gate_type":     "VERIFY_RED",
        "mission_id":    result.MissionID,
        "ac_id":         result.ACID,
        "exit_code":     result.ExitCode,
        "classification": result.Classification,
        "output":        truncate(result.Output, 1024), // First 1KB
        "timestamp":     result.Timestamp.Format(time.RFC3339),
        "attempt":       result.Attempt,
    }

    // Create protocol event bead
    return g.beads.CreateEvent(ctx, protocol.Event{
        Type:      "GATE_RESULT",
        Body:      evidence,
        Metadata:  map[string]string{"phase": "VERIFY_RED"},
    })
}
```

## Anti-Patterns to Avoid

### ❌ DON'T: Use AI in gates
```go
// BAD! Gate using LLM to judge output
func (g *Gate) JudgeOutput(output string) (string, error) {
    return llm.Ask("Did this test pass?", output) // NO!
}
```

### ✅ DO: Deterministic analysis only
```go
// GOOD! Pattern matching
func (g *Gate) JudgeOutput(output string) string {
    if strings.Contains(output, "PASS") {
        return "accept"
    }
    return "reject_failure"
}
```

### ❌ DON'T: Ignore exit codes
```go
if err != nil {
    // BAD! Ignoring actual error
}
```

### ✅ DO: Always check exit code
```go
if err != nil {
    exitCode := cmd.ProcessState.ExitCode()
    // Use exit code for classification
}
```

## Testing Requirements

### Unit Tests
- Test variable substitution logic
- Test output classification patterns
- Test evidence recording (with mocked Beads)

### Integration Tests
- Test gate execution with real shell commands
- Test timeout handling
- Test evidence persistence

## References

- `.spec/prd.md` - UC-GATE-01 through UC-GATE-08 (8 Gate use cases)
- `.spec/technical-requirements.md` - Gate execution (lines 897-1010)
- `.spec/research-findings/GO_CODING_STANDARDS.md` - Go coding standards

---

**Version**: 1.0
**Last Updated**: 2025-02-10
