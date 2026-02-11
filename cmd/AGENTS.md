# CLI Commands Package - Agent Guidance

## Purpose

This package contains all CLI entry points for Ship Commander 3. Each subdirectory under `cmd/` represents a top-level command in the `sc3` CLI.

## Context in Ship Commander 3

The `cmd/` package provides the user-facing CLI interface:
- `sc3` - Root command
- `sc3 plan` - Launch Ready Room planning session
- `sc3 execute` - Execute approved mission manifest
- `sc3 tui` - Launch Bubble Tea TUI dashboard

## Package Organization

```
cmd/
├── AGENTS.md              # This file
├── root/                  # Root command (sc3)
│   ├── main.go           # Entry point
│   └── root.go           # Root Cobra command setup
├── plan/                  # sc3 plan command
│   └── plan.go           # Planning orchestrator entry point
├── execute/               # sc3 execute command
│   └── execute.go        # Execution orchestrator entry point
└── tui/                   # sc3 tui command
    └── tui.go            # TUI entry point
```

## Dependencies

### External Dependencies
- `github.com/spf13/cobra` - CLI framework
- `github.com/spf13/viper` - Flag parsing and configuration

### Internal Dependencies
- `internal/commission` - Commission parsing and creation
- `internal/planning` - Ready Room orchestrator
- `internal/execution` - Commander orchestrator
- `internal/tui` - Bubble Tea TUI application
- `internal/config` - Configuration loading

## Coding Standards

Follow Go standards from `.spec/research-findings/GO_CODING_STANDARDS.md`:

### Cobra Command Pattern
```go
// ✅ GOOD: Standard Cobra command pattern
var cmdPlan = &cobra.Command{
    Use:   "plan <prd-file>",
    Short: "Launch Ready Room planning session",
    Long:  `Parse PRD and run Ready Room planning loop...`,
    Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        prdPath := args[0]
        return runPlan(ctx, prdPath)
    },
}
```

### Flag Conventions
- Use `kebab-case` for long flags: `--max-revisions`
- Use short flags for common options: `-v` for verbose
- Provide default values in flags
- Bind to Viper for configuration file support

### Error Handling
- Always return errors from `RunE`, never use `Run`
- Wrap errors with context: `fmt.Errorf("failed to parse PRD: %w", err)`
- Never panic in command handlers

## Common Patterns

### Command Entry Point
```go
package main

import (
    "context"
    "os"

    "github.com/ship-commander-3/internal/config"
    "github.com/spf13/cobra"
)

func main() {
    ctx := context.Background()

    // Load configuration
    cfg, err := config.Load()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
        os.Exit(1)
    }

    // Execute root command
    if err := ExecuteRoot(ctx, cfg); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}
```

### Subcommand Registration
```go
// root.go
package root

import (
    "github.com/spf13/cobra"
    "github.com/ship-commander-3/cmd/plan"
    "github.com/ship-commander-3/cmd/execute"
    "github.com/ship-commander-3/cmd/tui"
)

func ExecuteRoot(ctx context.Context, cfg *config.Config) error {
    rootCmd := &cobra.Command{
        Use:   "sc3",
        Short: "Ship Commander 3 - AI Agent Orchestration CLI",
    }

    // Register subcommands
    rootCmd.AddCommand(plan.NewCommand(ctx, cfg))
    rootCmd.AddCommand(execute.NewCommand(ctx, cfg))
    rootCmd.AddCommand(tui.NewCommand(ctx, cfg))

    return rootCmd.Execute()
}
```

### Context Propagation
```go
// Always pass context as first parameter to internal packages
func runPlan(ctx context.Context, prdPath string) error {
    commission, err := commission.Parse(ctx, prdPath)
    if err != nil {
        return fmt.Errorf("parse commission: %w", err)
    }

    // Continue with planning...
}
```

## Integration Points

### Configuration Loading
Commands must load configuration before execution:
1. Read `sc3.toml` (project defaults)
2. Read active config bead from Beads
3. Merge: TOML → Beads → CLI flags
4. Pass merged config to internal packages

### Telemetry Initialization
Commands must initialize OpenTelemetry before work:
1. Create tracer provider with OTLP exporter
2. Create root span with run_id and trace_id
3. Pass context with tracing to internal packages
4. Shutdown tracer on exit

### Beads Integration
Commands must ensure Beads is initialized:
1. Check if `.beads/` exists
2. Run `bd init` if not present
3. Use Beads client for all state operations

## Anti-Patterns to Avoid

### ❌ DON'T: Use Run instead of RunE
```go
var cmdBad = &cobra.Command{
    Run: func(cmd *cobra.Command, args []string) {
        // Can't return errors!
    },
}
```

### ✅ DO: Use RunE for error handling
```go
var cmdGood = &cobra.Command{
    RunE: func(cmd *cobra.Command, args []string) error {
        // Can return errors
        return doWork()
    },
}
```

### ❌ DON'T: Ignore context
```go
func runPlan(prdPath string) error {
    // No context!
}
```

### ✅ DO: Accept and propagate context
```go
func runPlan(ctx context.Context, prdPath string) error {
    // Context flows through
}
```

## Testing Requirements

### Unit Tests
- Test command registration
- Test flag parsing
- Test error handling paths
- Mock internal package dependencies

### Integration Tests
- Test full command execution with real Beads
- Test flag combinations
- Test configuration file loading

## References

- [Cobra Documentation](https://github.com/spf13/cobra) - CLI framework
- `.spec/technical-requirements.md` - CLI entry point specifications
- `.spec/research-findings/GO_CODING_STANDARDS.md` - Go coding standards

---

**Version**: 1.0
**Last Updated**: 2025-02-10
