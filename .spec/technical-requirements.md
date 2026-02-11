# Technical Requirements: Ship Commander 3

**Generated**: 2026-02-10
**Stack**: Go + Bubble Tea + Beads + tmux

---

## System Components

| Component | Description | Technology |
|-----------|-------------|-----------|
| **CLI Entry Point** | Command-line interface: `sc3 init`, `sc3 plan`, `sc3 execute`, `sc3 tui`, `sc3 status` | Go `cobra` or custom CLI |
| **Commission Parser** | Parses Markdown PRDs into structured Commission with use cases and acceptance criteria | Go `goldmark` (Markdown AST) + YAML frontmatter |
| **Ready Room** | Planning loop orchestrator: spawns agent sessions, routes messages, validates consensus, manages Admiral gates | Go goroutines + channel-based message routing |
| **Admiral Gates** | Human approval gate + question gate: presents decisions to Admiral via TUI, suspends until resolved | Go channels + Bubble Tea components |
| **Commander** | Mission execution orchestrator: dispatches agents, runs gates, enforces termination, validates demo tokens | Go goroutines + deterministic state machine |
| **Verification Engine** | Deterministic gate execution: shell commands, exit code checking, output classification | Go `os/exec` + output parsing |
| **Protocol Service** | Structured JSON event system: schema validation, Beads persistence, event bus emission | Go JSON + Beads CLI wrapper |
| **Demo Token Validator** | Validates `demo/MISSION-<id>.md` against strict schema: YAML frontmatter, evidence sections, file references | Go YAML parser + file checks |
| **Termination Enforcer** | Deterministic halt rules: max revisions, missing demo token, AC exhaustion | Go boolean logic (no AI) |
| **Beads Client** | Go wrapper around `bd` CLI: init, create, set-state, dep, agent, comments, list, ready, graph | Go `os/exec` with `--json` output parsing |
| **State Machine** | Finite state machine for commission, mission, AC, and agent lifecycle transitions | Go with legal transition assertions |
| **Lock Service** | Surface-area lock management: acquire, conflict detection (glob matching), release, expiry | Go + Beads-backed lock records |
| **Harness Abstraction** | Interface for spawning agent sessions across CLI harnesses | Go interface + driver implementations |
| **Claude Code Driver** | Spawns `claude` CLI with flags in tmux session, captures output | Go `os/exec` + tmux commands |
| **Codex Driver** | Spawns `codex` CLI with flags in tmux session, captures output | Go `os/exec` + tmux commands |
| **tmux Manager** | Creates, lists, sends keys to, captures output from, and kills tmux sessions | Go wrapper around `tmux` CLI |
| **Event Bus** | Typed pub/sub with wildcard subscribers for state change propagation | Go channels + goroutines |
| **TUI App** | Terminal dashboard with multiple views: planning, execution, health, event log | Bubble Tea (Elm architecture) |
| **LCARS Theme** | Lipgloss styles for LCARS color palette, box-drawing borders, semantic status colors | Lipgloss style definitions |
| **Doctor** | Background health monitor: stuck agent detection, orphan recovery, heartbeat monitoring | Go goroutine on ticker interval |
| **Config** | Project and global configuration: harness defaults, WIP limits, timeouts, role-to-model mapping | TOML files |

---

## Data Model (Beads)

All state stored via Beads (Git + SQLite). Beads provides human-readable JSONL files that are grep-able, diff-able, and Git-trackable. The `bd` CLI provides instant debugging.

### Ship Beads (Persistent Agent Teams)

**Note**: Ships are **persistent agent team groupings** that exist independently of commissions. A ship serves multiple commissions over its lifetime and represents a reusable pool of specialized agents.

| Beads Field | Ship Commander Usage |
|-------------|---------------------|
| `id` | Unique ship hash ID |
| `title` | Ship name (e.g., "USS Auth", "USS Payments") |
| `type` | `ship` |
| State: `status` | `active` / `decommissioned` |
| State: `health` | `healthy` / `degraded` / `critical` |
| Body | Ship charter, domain specialization, crew composition rules |
| Labels | `crew_count:<n>`, `domain:<auth|payments|docs|generic>`, `home_port:<project-path>` |
| State: `current_commission` | Commission ID currently assigned (or null) |
| State: `commission_history` | Array of completed commission IDs |
| Children | Agent beads (crew members) |

**Ship Lifecycle**:
```
commissioned → active → [serves commissions] → decommissioned
```

**Ship → Commission Relationship**: Many-to-one over time (one ship serves many commissions sequentially).

### Commission Beads (Child of Ship)

| Beads Field | Ship Commander Usage |
|-------------|---------------------|
| `id` | Unique commission hash ID |
| `parent` | Ship bead ID (assigned ship) |
| `title` | PRD title / initiative name |
| `type` | `commission` |
| State: `status` | `planning` / `approved` / `executing` / `completed` / `shelved` |
| State: `planning_iteration` | Current Ready Room iteration number (default: 0) |
| State: `max_planning_iterations` | Max iterations before consensus failure (default: 5) |
| State: `current_wave_number` | Current wave being executed (default: 1) |
| State: `total_waves` | Total waves computed from dependency graph |
| State: `use_case_coverage` | Map of use case ID → coverage status (uncovered/partially_covered/covered) with mission IDs |
| Body | Full PRD content with use cases |
| Labels | `prd_path:<path>`, `use_case_count:<n>`, `ship_id:<ship-id>` |
| Children | Mission beads, Wave beads |

### Mission Beads (Child of Commission)

| Beads Field | Ship Commander Usage |
|-------------|---------------------|
| `id` | Unique mission hash ID |
| `parent` | Commission bead ID |
| `title` | Mission title (action verb) |
| State: `status` | `backlog` / `in_progress` / `review` / `approved` / `done` / `halted` |
| State: `classification` | `RED_ALERT` / `STANDARD_OPS` |
| State: `signoffs.captain` | Boolean - Captain functional sign-off |
| State: `signoffs.commander` | Boolean - Commander technical sign-off |
| State: `signoffs.designOfficer` | Boolean - Design Officer design sign-off |
| State: `consensus_reached` | Boolean - all three sign-offs complete |
| Labels | `wave:<n>`, `sequence:<n>`, `revision_count:<n>`, `max_revisions:<n>`, `harness:<type>`, `use_case_refs:UC-01,UC-02` |
| Body | Full mission spec with acceptance criteria |
| Dependencies | `bd dep add <mission> <blocking-mission>` |
| Children | AC beads |

### AC Beads (Child of Mission)

| Beads Field | Ship Commander Usage |
|-------------|---------------------|
| `id` | Unique AC hash ID |
| `parent` | Mission bead ID |
| `title` | AC title (e.g., "AC-1: User can login") |
| State: `tdd` | `red` / `verify_red` / `green` / `verify_green` / `refactor` / `verify_refactor` |
| State: `attempt` | Current attempt number |
| Comments | Gate evidence (exit code, output snippet, classification) |

### Agent Beads

**Note**: Agent harness and model are configured per-agent via labels, NOT in project/global config files. Each agent bead stores its own harness and model selection.

| Beads Field | Ship Commander Usage |
|-------------|---------------------|
| `id` | Unique agent instance ID |
| `parent` | Ship bead ID (crew membership) |
| `type` | `agent` |
| Labels | `role:<captain/commander/design-officer/implementer/reviewer>`, `harness:<claude/codex>`, `model:<opus/sonnet/haiku/gpt-4/etc>`, `mission:<id>`, `tmux:<session-name>` |
| State: `status` | `idle` / `spawning` / `running` / `stuck` / `done` / `dead` |
| State: `ship_id` | Ship this agent is assigned to |
| Heartbeat | `bd agent heartbeat <id>` |

**Agent Configuration Resolution** (harness + model):
```go
func GetAgentConfig(agentID string) (harness, model string, err error) {
    agentBead, err := beads.Get(agentID)
    if err != nil {
        return "", "", err
    }
    harness = agentBead.Labels["harness"]      // "claude" or "codex"
    model = agentBead.Labels["model"]          // "opus", "sonnet", "haiku", etc.
    return harness, model, nil
}
```

**Per-Agent Configuration Examples**:
- Captain agent: `role=captain, harness=claude, model=opus` (most capable)
- Implementer agent: `role=implementer, harness=claude, model=sonnet` (balanced)
- Reviewer agent: `role=reviewer, harness=claude, model=haiku` (fast, cheap)
- Design Officer: `role=design-officer, harness=claude, model=opus` (design requires highest capability)

**NOT in Config Files**:
- ❌ `[roles.captain] model = "opus"` (removed)
- ❌ `[crew] default_model = "sonnet"` (removed)
- ✅ Agent beads store their own harness/model via labels

### System Health Beads

**Note**: System health beads track Doctor monitoring state for crash recovery and observability.

| Beads Field | Ship Commander Usage |
|-------------|---------------------|
| `id` | Unique health check ID (health-<uuid>) |
| `type` | `system_health` |
| `parent` | Ship bead ID (ship-level health) |
| State: `stuck_agents` | Array of agent IDs currently stuck |
| State: `orphan_missions` | Array of mission IDs with no active agent |
| State: `last_heartbeat_check` | ISO 8601 timestamp of last Doctor check |
| State: `health_status` | `healthy` / `degraded` / `critical` |
| Labels | `doctor_session:<session-id>`, `timestamp:<ISO8601>` |
| Body | Health check summary, alerts, recommendations |

**Health Check Frequency**: Every 30 seconds (configurable via `[doctor] heartbeat_interval_seconds`).

### Config Beads

**Note**: Configuration is stored in Beads for full audit trail and crash recovery. Active config can be updated via TUI or CLI and persists across runs.

| Beads Field | Ship Commander Usage |
|-------------|---------------------|
| `id` | Unique config hash ID (cfg-<uuid>) |
| `type` | `config` |
| `parent` | Project root bead OR ship bead (for ship-level config) |
| State: `config_type` | `project` / `ship` / `global` |
| State: `version` | Config schema version (e.g., "2.0") for migrations |
| State: `active` | Boolean - whether this config is currently active |
| Body | Full configuration as JSON (gates, limits, policies, observability) |
| Labels | `project:<name>`, `environment:<dev\|prod\|test>`, `applied_at:<ISO8601>`, `applied_by:<user>` |

**Config Types**:
- **Project config**: Verification gates, limits, lock policies (defaults from `sc3.toml`)
- **Ship config**: Ship-level overrides, crew composition rules (defaults from ship charter)
- **Global config**: TUI preferences, observability settings (defaults from `~/.sc3/config.json`)

**Active Config Resolution**:
1. Query active config bead by project/ship/global type
2. Fall back to `sc3.toml` if no config bead exists
3. Merge: TOML defaults → Beads active config
4. CLI flags override both

**Config Versioning**:
- Config schema version tracked in `state.version`
- On startup, if `state.version < currentVersion`, run migration
- Migrations preserve settings, update schema, save new config bead

**Example Config Bead Body** (JSON):
```json
{
  "version": "2.0",
  "verification_gates": {
    "test_command": "go test ./...",
    "test_command_timeout": 120,
    "typecheck_command": "go vet ./...",
    "typecheck_command_timeout": 30,
    "lint_command": "golangci-lint run --timeout=5m",
    "lint_command_timeout": 120,
    "build_command": "go build ./...",
    "build_command_timeout": 60
  },
  "limits": {
    "wip_limit": 3,
    "max_revisions": 3,
    "max_planning_iterations": 5,
    "gate_timeout_seconds": 120
  },
  "locks": {
    "default_ttl_seconds": 7200,
    "auto_renew_on_heartbeat": true,
    "max_renewals": 3,
    "renewal_warning_seconds": 600
  }
}
```

### Protocol Event Beads

**Note**: Protocol events are stored as individual beads for full audit trail, queryability via `bd list --type=protocol_event`, and crash recovery.

| Beads Field | Ship Commander Usage |
|-------------|---------------------|
| `id` | Unique event hash ID (evt-<uuid>) |
| `type` | `protocol_event` |
| `parent` | Commission or Mission ID (contextual) |
| State: `event_type` | `AGENT_CLAIM` / `GATE_RESULT` / `STATE_TRANSITION` / `READY_ROOM_MESSAGE` / `ADMIRAL_QUESTION` / `ADMIRAL_ANSWER` |
| State: `mission_id` | Associated mission ID (if applicable) |
| State: `commission_id` | Associated commission ID |
| State: `timestamp` | ISO 8601 event timestamp |
| State: `protocol_version` | `"1.0"` |
| State: `agent_id` | Agent that produced the event (if applicable) |
| State: `phase` | Current execution phase (if applicable) |
| Body | `description` (string, human-readable), `evidence` (object, structured evidence) |

**Query Examples**:
```bash
# All events for a mission
bd list --type=protocol_event -l mission_id:MISSION-42

# All gate results in a commission
bd list --type=protocol_event -l commission_id=<hash>,event_type=GATE_RESULT

# Audit trail since timestamp
bd list --type=protocol_event --since "2026-02-10T10:00:00Z"
```

### Surface-Area Locks

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Lock ID |
| `mission_id` | string | Owning mission |
| `surface_area` | string[] | File glob patterns (e.g., `["src/auth/*", "src/auth/**/*.test.ts"]`) |
| `acquired_at` | ISO 8601 | Lock acquisition time |
| `expires_at` | ISO 8601 | Lock expiry time |
| State: `renewal_state` | `active` / `expiring_soon` / `expired` / `renewed` |
| State: `renewal_count` | int | Number of times lock has been renewed |
| State: `last_renewed_at` | ISO 8601 | Timestamp of most recent renewal |

#### Lock Renewal State Machine

**States**: `active` → `expiring_soon` → `expired` | `renewed`

**Renewal Triggers**:
1. **Automatic renewal on heartbeat**: When agent heartbeat occurs <50% of TTL, auto-renew
2. **Explicit renewal**: Agent can request renewal via protocol event `LOCK_RENEW_REQUEST`
3. **Manual renewal**: Admiral can extend lock via TUI

**Lock Policy** (config):
```toml
[locks]
default_ttl_seconds = 7200  # 2 hours
auto_renew_on_heartbeat = true
max_renewals = 3  # Max times a lock can auto-renew
renewal_warning_seconds = 600  # Warn Admiral 10 min before expiry
```

**State Transitions**:
```
active → expiring_soon (when time_remaining < renewal_warning_seconds)
expiring_soon → renewed (on heartbeat if auto_renew && renewal_count < max_renewals)
expiring_soon → expired (when time_remaining == 0)
expired → error (any operation on expired lock fails)
```

### Admiral Question Beads

| Beads Field | Ship Commander Usage |
|-------------|---------------------|
| `id` | Unique question hash ID |
| `type` | `question` |
| `parent` | Commission bead ID (or agent bead ID) |
| State: `status` | `pending` / `answered` / `skipped` / `expired` |
| State: `broadcast` | Boolean - whether answer should be broadcast to all agents |
| Body | Question text (Markdown), domain (functional/technical/design), context |
| Labels | `question_id:<uuid>`, `agent_role:<captain\|commander\|design_officer\|ensign>` |
| Comments | Answer content (option text or free text), answered timestamp, answered by (Admiral) |

**State Machine**:
```
pending → answered (Admiral provides answer)
pending → skipped (Admiral selects "Skip" option)
pending → expired (Planning session ends without answer)
```

### Wave Beads

| Beads Field | Ship Commander Usage |
|-------------|---------------------|
| `id` | Unique wave hash ID |
| `type` | `wave` |
| `parent` | Commission bead ID |
| State: `status` | `pending` / `active` / `complete` / `halted` |
| State: `mission_ids` | Array of mission IDs in this wave (e.g., `["MISSION-01", "MISSION-03", "MISSION-05"]`) |
| State: `dependency_summary` | ASCII dependency graph or text summary (e.g., "A → B, C (no dependencies)") |
| Body | Wave number, detailed dependency graph, wave metadata |
| Labels | `commission_id:<id>`, `wave_number:<n>`, `mission_count:<n>` |
| Comments | Admiral feedback (from wave review), review timestamp |

**Feedback Injection**:
```go
type MissionDispatchContext struct {
    MissionID      string
    MissionSpec    string
    PriorFeedback  string // Wave feedback prepended here
    PreviousGateResults []GateResult
}

func BuildAgentPrompt(ctx MissionDispatchContext) string {
    prompt := ctx.MissionSpec

    if ctx.PriorFeedback != "" {
        prompt = fmt.Sprintf(`# Previous Wave Feedback

%s

Please incorporate this feedback into your work.`, ctx.PriorFeedback) + "\n\n" + prompt
    }

    return prompt
}
```

### Ready Room Message Beads

| Beads Field | Ship Commander Usage |
|-------------|---------------------|
| `id` | Unique message hash ID |
| `type` | `ready_room_message` |
| `parent` | Commission bead ID |
| Body | Full message content (Markdown) |
| Labels | `from_agent_id:<id>`, `to_agent_id:<id>\|broadcast`, `planning_iteration:<n>`, `domain:<functional\|technical\|design>` |
| State: `delivered` | boolean (track delivery confirmation) |
| Comments | Timestamp, message type (question/statement/acknowledgment) |

---

## Architecture Diagram

```
+========================================================================+
|                          SHIP COMMANDER 3                               |
|                                                                         |
|  +-----------+    +---------------------+    +--------------------+     |
|  |  CLI      |    |  Bubble Tea TUI     |    |  Doctor            |     |
|  |  (cobra)  |--->|  (Lipgloss/LCARS)   |    |  (Health Monitor)  |     |
|  |           |    |  - Planning view     |    |  - Heartbeat loop  |     |
|  |  sc3 plan |    |  - Execution view    |    |  - Stuck detect    |     |
|  |  sc3 exec |    |  - Admiral modals    |    |  - Orphan detect   |     |
|  |  sc3 tui  |    |  - Event log         |    |                    |     |
|  +-----------+    +---------+-----------+    +--------+-----------+     |
|                             |                         |                  |
|                             | subscribes              | queries          |
|                             v                         v                  |
|                    +--------+---------+                                  |
|                    |   Event Bus      |<---------------------------------+
|                    |  (Go channels)   |                                  |
|                    +--------+---------+                                  |
|                             ^                                            |
|                             | emits                                      |
|         +-------------------+-------------------+                        |
|         |                                       |                        |
|  +------+-------+                      +--------+---------+              |
|  | Ready Room   |                      |  Commander        |             |
|  | (Planning)   |                      |  (Execution)      |             |
|  |              |                      |                    |             |
|  | - Captain    |  approved manifest   | - Dispatch agents  |             |
|  | - Commander  |--------------------->| - Run VERIFY gates |             |
|  | - Design Off |                      | - Enforce termn    |             |
|  | - Consensus  |                      | - Validate demo    |             |
|  | - Admiral Q  |                      | - Surface locks    |             |
|  +------+-------+                      +---+----+----+-----+             |
|         |                                  |    |    |                    |
|         | [Human Input]           spawns   |    |    |  spawns            |
|         v                    +-------------+    |    +-------------+     |
|  +------+-------+           v                   v                 v      |
|  | Admiral Gate |    +------+------+   +--------+----+   +--------+--+  |
|  | (TUI modal)  |    | Claude Code |   | Codex       |   | Reviewer  |  |
|  | - Approve    |    | Driver      |   | Driver      |   | Agent     |  |
|  | - Feedback   |    +------+------+   +------+------+   +-----------+  |
|  | - Shelve     |           |                  |                         |
|  | - Questions  |           v                  v                         |
|  +--------------+    +------+------+   +------+------+                   |
|                      | tmux Session|   | tmux Session|                   |
|                      | (worktree)  |   | (worktree)  |                   |
|                      +-------------+   +-------------+                   |
|                                                                          |
|  +--------------------------------------------------------------------+ |
|  |  Beads State Layer (.beads/)                                        | |
|  |  Git + SQLite: commissions, missions, ACs, agents, deps,           | |
|  |  protocol events, locks, audit log                                  | |
|  |  CLI: bd ready | bd graph | bd agent | bd activity | bd show       | |
|  +--------------------------------------------------------------------+ |
+==========================================================================+
```

---

## Observability Architecture

Ship Commander 3 provides comprehensive observability via OpenTelemetry distributed tracing and structured JSON logging. Every operation (deterministic pipeline steps and LLM calls) emits spans correlated with logs via `run_id` and `trace_id`, enabling rapid diagnosis: WHERE something broke (which step, which LLM call) and WHY (error type, exit code, bounded output).

### Trace Correlation Model

```
run_id (UUID)         // Top-level run identifier (one per `sc3` command)
  └── trace_id (UUID)  // OpenTelemetry trace identifier (correlates all spans)
       └── span_id (UUID)  // Individual span identifier (state transition, tool exec, LLM call)

Log Record:
{
  "timestamp": "2026-02-10T14:30:01Z",
  "level": "ERROR",
  "run_id": "abc-123-def",
  "trace_id": "7bf3a8f...",
  "span_id": "3f2c1e9...",
  "component": "commander",
  "message": "Gate VERIFY_RED failed",
  "error_type": "reject_vanity",
  "mission_id": "MISSION-42",
  ...
}
```

### Span Hierarchy

```
run (root span, run_id=<uuid>)
  ├─ state.transition (commission: planning → approved)
  ├─ ready_room.spawn_agents (captain, commander, design_officer)
  │   └─ llm.call (model=opus, latency=2.3s, tokens=4500)
  ├─ state.transition (commission: approved → executing)
  ├─ mission.dispatch (MISSION-42, agent=impl-abc)
  ├─ tool.exec (git worktree add feature/MISSION-42-auth)
  ├─ state.transition (MISSION-42: backlog → in_progress)
  ├─ llm.call (model=sonnet, agent=impl-abc, harness=claude)
  │   ├─ tool.call (git write file)
  │   ├─ tool.call (git apply patch)
  │   └─ tool.call (go test ./auth/...)
  ├─ gate.verify_red (mission=MISSION-42, ac=AC-1, result=reject_vanity)
  ├─ llm.call (model=sonnet, agent=impl-abc, retry=true)
  ├─ gate.verify_green (mission=MISSION-42, ac=AC-1, result=pass)
  ├─ state.transition (MISSION-42: in_progress → review)
  └─ demo_token.validate (mission=MISSION-42, result=pass)
```

### Telemetry Package Structure

```go
// internal/telemetry/telemetry.go

package telemetry

import (
    "context"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/trace"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
    "go.opentelemetry.io/otel/sdk/resource"
    "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

type TelemetryConfig struct {
    ServiceName    string
    ServiceVersion string
    Environment    string // dev | prod | test
    OTelEndpoint   string // e.g., http://localhost:4318
    DebugMode      bool   // Enable console exporter
}

// Init initializes OpenTelemetry tracer provider and returns shutdown func
func Init(ctx context.Context, cfg TelemetryConfig) (context.Context, func(context.Context) error, error) {
    // 1. Create resource attributes (service.name, service.version, environment)
    // 2. Configure exporter (OTLP HTTP or console for debug)
    // 3. Create tracer provider with batch span processor
    // 4. Set global tracer provider
    // 5. Return context with run_id and trace_id, shutdown func
}

// RootSpan creates root span for run with run_id and trace_id
func RootSpan(ctx context.Context, command string, args []string) (context.Context, trace.Span) {
    // Generate run_id (UUID v4)
    // Create root span with attributes: command.name, command.args (redacted), cwd, git.head
    // Inject run_id and trace_id into context
}

// StateTransitionSpan records state machine transitions
func StateTransitionSpan(ctx context.Context, entityType, fromState, toState, reason string) (context.Context, trace.Span) {
    // Span name: "state.transition"
    // Attributes: entity_type, from_state, to_state, reason
}

// ToolExecSpan records tool/command execution
func ToolExecSpan(ctx context.Context, toolName, cwd string, args []string) (context.Context, trace.Span) {
    // Span name: "tool.exec"
    // Attributes: tool_name, args_redacted, cwd
    // On completion: duration_ms, exit_code, stdout_preview, stderr_preview
}

// LLMCallSpan records LLM/harness calls
func LLMCallSpan(ctx context.Context, model, harness string) (context.Context, trace.Span) {
    // Span name: "llm.call"
    // Attributes: model_name, harness
    // On completion: latency_ms, prompt_tokens, response_tokens, total_tokens, tool_calls_count
    // Events: tool.call for each tool invocation
}

// InvariantViolation records invariant violations (events, not panics)
func InvariantViolation(ctx context.Context, invariantName, severity string, context map[string]interface{}) {
    // Emit span event: "invariant.violation"
    // Attributes: invariant_name, severity (warn|error), context (structured details)
}
```

### Logging Architecture

```go
// internal/logging/json_logger.go

package logging

import (
    "context"
    "os"
    "go.opentelemetry.io/otel/trace"
)

type JSONLogger struct {
    file   *os.File
    runID  string
    traceID string
}

// LogRecord is structured JSON log entry
type LogRecord struct {
    Timestamp string                 `json:"timestamp"`
    Level     string                 `json:"level"`
    RunID     string                 `json:"run_id"`
    TraceID   string                 `json:"trace_id,omitempty"`
    SpanID    string                 `json:"span_id,omitempty"`
    Component string                 `json:"component"`
    Message   string                 `json:"message"`
    Fields    map[string]interface{} `json:"fields,omitempty"`
}

// Emit writes log record to JSON file (never stdout while TUI active)
func (l *JSONLogger) Emit(ctx context.Context, level, component, message string, fields map[string]interface{}) {
    // Extract trace_id and span_id from OpenTelemetry context
    // Write JSON log record to file: .sc3/logs/sc3-<date>-<run_id>.json
    // Rotate at 10MB, retain 5 files (configurable)
}
```

### Redaction Policy

Secrets redacted from span attributes and log fields:

| Pattern | Redacted To | Example |
|---------|-------------|---------|
| API keys | `***REDACTED***` | `sk-ant-***REDACTED***` |
| Passwords | `***REDACTED***` | `password: ***REDACTED***` |
| Tokens | `***REDACTED***` | `Authorization: Bearer ***REDACTED***` |
| Sensitive paths | `***REDACTED***` | `/home/user/.ssh/***REDACTED***` |

### Debug Bundle Contents

```
.sc3-bugreport-20260210-143001.tar.gz
├── README.txt                          # Collection summary
├── version.txt                         # `sc3 --version` output
├── config.json                         # Redacted config (secrets masked)
├── last_run_id.txt                     # Last run_id/trace_id
├── logs/
│   ├── sc3-2026-02-10-abc123.json     # Last 3 log files
│   ├── sc3-2026-02-10-def456.json
│   └── sc3-2026-02-10-ghi789.json
├── git-state.txt                       # Git HEAD, branch, status
├── git-diff.patch                      # Git diff/HEAD
└── failing-test-output.txt             # If last run failed
```

### Configuration

```toml
# sc3.toml (project config)

[observability]
enabled = true
otel_endpoint = "http://localhost:4318"  # OTLP collector
environment = "dev"                      # dev | prod | test

[logging]
level = "INFO"                           # DEBUG | INFO | WARN | ERROR
max_size_mb = 10                        # Log rotation size
max_files = 5                           # Retention count

[debug]
console_exporter = false                # Enable for --debug flag
verbose_logging = false                 # Enable for --debug flag
```

### Environment Variables

| Variable | Purpose | Default |
|----------|---------|---------|
| `OTEL_EXPORTER_OTLP_ENDPOINT` | OTLP collector endpoint | `http://localhost:4318` |
| `OTEL_SERVICE_NAME` | Service name override | `ship-commander-3` |
| `OTEL_ENVIRONMENT` | Environment (dev/prod/test) | `dev` |
| `SC3_DEBUG` | Enable debug mode | `false` |
| `SC3_LOG_LEVEL` | Logging level | `INFO` |

---

## External Dependencies

### Core Runtime

| Dependency | Purpose | Documentation |
|-----------|---------|--------------|
| **Go** 1.22+ | Core language, single binary distribution | https://go.dev/doc/ |
| **Beads** (`bd` CLI) | Persistent state layer (Git + SQLite) | https://github.com/steveyegge/beads |
| **tmux** | Session management for agent process isolation | https://github.com/tmux/tmux |

### Go Libraries

| Dependency | Purpose | Documentation |
|-----------|---------|--------------|
| **Bubble Tea** | TUI framework (Elm architecture for terminals) | https://github.com/charmbracelet/bubbletea |
| **Lipgloss** | TUI styling (LCARS color palette, borders) | https://github.com/charmbracelet/lipgloss |
| **Bubbles** | TUI components (text input, list, spinner, table) | https://github.com/charmbracelet/bubbles |
| **Huh** | TUI forms (Admiral question modal, approval prompts) | https://github.com/charmbracelet/huh |
| **Log** | Structured logging | https://github.com/charmbracelet/log |
| **OpenTelemetry Go SDK** | Distributed tracing, metrics (OpenTelemetry) | https://opentelemetry.io/docs/instrumentation/go/ |
| **OpenTelemetry OTLP Exporter** | Trace export over OTLP protocol | https://pkg.go.dev/go.opentelemetry.io/otel/exporters/otlp/otlptrace |
| **OpenTelemetry Resource Detectors** | Auto-detect service name, version, environment | https://pkg.go.dev/go.opentelemetry.io/contrib/detectors |

### CLI Harnesses (External, user-installed)

| Dependency | Purpose | Documentation |
|-----------|---------|--------------|
| **claude** CLI | Claude Code agent harness (default) | https://docs.anthropic.com/en/docs/claude-code |
| **codex** CLI | OpenAI Codex agent harness (optional) | https://github.com/openai/codex |

### Development

| Dependency | Purpose | Documentation |
|-----------|---------|--------------|
| **Go test** | Built-in test runner | https://pkg.go.dev/testing |
| **testify** | Test assertions and mocks | https://github.com/stretchr/testify |
| **golangci-lint** | Linting | https://golangci-lint.run/ |

### Distribution

| Channel | Method | Command |
|---------|--------|---------|
| **Homebrew** | Tap + formula | `brew install sc3` |
| **go install** | Direct from source | `go install github.com/.../sc3@latest` |
| **GitHub Releases** | Pre-built binaries | Download from releases page |

---

## Config Format (Beads + Project TOML)

Ship Commander 3 uses **Beads for persistent configuration** combined with a human-editable project TOML file. Configuration is stored in Beads for auditability and crash recovery, while the TOML file provides git-tracked project defaults.

### Config Beads

Configuration is stored as Beads for full audit trail and crash recovery:

| Beads Field | Ship Commander Usage |
|-------------|---------------------|
| `id` | Unique config hash ID |
| `type` | `config` |
| `parent` | Project root bead (or ship bead for ship-level config) |
| State: `config_type` | `project` / `ship` / `global` |
| State: `version` | Config schema version (for migrations) |
| State: `active` | Boolean - whether this config is currently active |
| Body | Full configuration as JSON (gates, limits, policies) |
| Labels | `project:<name>`, `environment:<dev|prod|test>`, `applied_at:<ISO8601>` |

**Config Loading Flow**:
1. Read active config bead from Beads (by project/ship/global type)
2. Fall back to `sc3.toml` if no config bead exists
3. Merge: TOML defaults → Beads active config
4. CLI flags override both

### Project Config: `sc3.toml` (TOML - Human-Editable, Git-Trackable)

Project-specific settings committed to repository. These are **defaults** that can be overridden via Beads config or CLI flags.

**IMPORTANT**: Agent harness and model selection are **NOT** configured here — they are set at the agent level (see Agent Beads schema).

```toml
[project]
name = "my-app"
beads_path = ".beads"
domain = "auth"  # Optional: ship specialization domain

[limits]
wip_limit = 3              # Max concurrent missions
max_revisions = 3          # Max revision attempts per mission
max_planning_iterations = 5  # Max Ready Room iterations
gate_timeout_seconds = 120  # Default gate command timeout

[verification_gates]
# Project-level bash commands for verification gates
# Commands run in mission worktree with variable substitution

# Test gate (VERIFY_RED, VERIFY_GREEN, VERIFY_REFACTOR)
test_command = "go test ./..."
test_command_timeout = 120

# Typecheck gate (VERIFY_IMPLEMENT)
typecheck_command = "go vet ./..."
typecheck_command_timeout = 30

# Lint gate (VERIFY_IMPLEMENT)
lint_command = "golangci-lint run --timeout=5m"
lint_command_timeout = 120

# Build gate (VERIFY_IMPLEMENT)
build_command = "go build ./..."
build_command_timeout = 60

# Infrastructure-specific: run tests 3x for flakiness detection
# Use {test_file} variable for per-AC filtering
test_command_flaky = "go test {test_file} -count=3"

[locks]
default_ttl_seconds = 7200      # 2 hours
auto_renew_on_heartbeat = true   # Auto-renew locks on agent heartbeat
max_renewals = 3                 # Max times a lock can auto-renew
renewal_warning_seconds = 600    # Warn Admiral 10 min before expiry
```

**Variable Substitution** (in gate commands):
- `{test_file}` - Test file path for per-AC filtering (e.g., `auth_test.go`)
- `{worktree_dir}` - Mission worktree directory
- `{mission_id}` - Mission identifier (e.g., `MISSION-42`)
- `{ac_title}` - Acceptance criteria title (slugified)

### Global Config: `~/.sc3/config.json` (JSON - TUI Preferences Only)

**DEPRECATED**: Global config now ONLY contains TUI user preferences. All other configuration moved to Beads + project TOML.

```json
{
  "version": "1.0",
  "tui": {
    "mode": "advanced",
    "accessible": false,
    "animations": true,
    "bell": false
  },
  "observability": {
    "enabled": true,
    "otel_endpoint": "http://localhost:4318",
    "environment": "dev",
    "logging": {
      "level": "INFO",
      "max_size_mb": 10,
      "max_files": 5
    },
    "debug": {
      "console_exporter": false,
      "verbose_logging": false
    }
  }
}
```

### Config Loading Example

```go
type Config struct {
    Beads    *ConfigBead    // Active config from Beads
    Project  *ProjectConfig // sc3.toml defaults
    Global   *GlobalConfig  // ~/.sc3/config.json (TUI prefs only)
}

func LoadConfig(projectPath string) (*Config, error) {
    // 1. Load active config bead from Beads
    beadsConfig, err := loadActiveConfigBead(projectPath)
    if err != nil {
        // Fall back to defaults if no bead exists
        beadsConfig = getDefaultConfig()
    }

    // 2. Load project TOML (defaults)
    projectConfig := loadProjectConfig("sc3.toml")
    if projectConfig == nil {
        projectConfig = getDefaultProjectConfig()
    }

    // 3. Load global TUI preferences
    globalConfig := loadGlobalConfig("~/.sc3/config.json")

    // 4. Merge: TOML defaults → Beads config → TUI prefs
    return &Config{
        Beads:   beadsConfig,
        Project: projectConfig,
        Global:  globalConfig,
    }, nil
}

// Agent harness/model is retrieved from Agent bead, not config
func (c *Config) GetAgentConfig(agentID string) (harness, model string, err error) {
    agentBead, err := beads.Get(agentID)
    if err != nil {
        return "", "", err
    }
    harness = agentBead.Labels["harness"]
    model = agentBead.Labels["model"]
    return harness, model, nil
}
```

### Config File Responsibilities

| Setting | Storage | Rationale |
|---------|---------|-----------|
| **Project name, beads_path, domain** | `sc3.toml` | Project metadata |
| **Verification gate commands** | `sc3.toml` | Project-level bash commands, git-tracked |
| **Limits (WIP, revisions, timeouts)** | `sc3.toml` | Project constraints, git-tracked |
| **Lock policies** | `sc3.toml` | Project lock defaults |
| **Agent harness/model** | **Agent bead** (labels) | Per-agent configuration, NOT in files |
| **TUI preferences** | `~/.sc3/config.json` | User preference, not domain state |
| **Observability settings** | `~/.sc3/config.json` | User preference, not domain state |

### Config Migrations

Config schema version tracked in Beads. On startup, if `beadsConfig.version < currentVersion`, run migration:

```go
const currentConfigVersion = "2.0"

func migrateConfig(config *ConfigBead) error {
    if config.Version >= currentConfigVersion {
        return nil
    }

    // Run migrations based on version
    switch config.Version {
    case "1.0":
        // Migrate gates from global to project config
        migrateV1ToV2(config)
        fallthrough
    case "2.0":
        // Current version
        config.Version = currentConfigVersion
        // Save migrated config to Beads
    }

    return nil
}
```

**V1 → V2 Migration** (when upgrading from old config format):
- Move `gates.*` from `~/.sc3/config.json` to project `sc3.toml`
- Remove `crew.*` (harness/model moved to agent beads)
- Preserve TUI preferences in global config
- Create config bead with migrated settings

---

## Verification Gates

Ship Commander 3 uses deterministic shell commands to verify work quality at critical phases. Gates execute commands, check exit codes, and capture output as evidence.

### Gate Phases

| Phase | Purpose | Typical Command | Exit Code Meaning |
|-------|---------|-----------------|-------------------|
| **VERIFY_RED** | Confirm test fails before fix | `go test -v {test_file}` | Non-zero = red confirmed |
| **VERIFY_GREEN** | Confirm test passes after fix | `go test -v {test_file}` | Zero = green confirmed |
| **VERIFY_REFACTOR** | Confirm behavior preserved | `go test ./...` | Zero = refactor safe |
| **TYPECHECK** | Validate type correctness | `go vet ./...` | Zero = types valid |
| **LINT** | Validate code quality | `golangci-lint run {surface_area}` | Zero = style compliant |
| **BUILD** | Validate compilation | `go build ./...` | Zero = builds successfully |

### Variable Substitution

Gate commands support variable substitution to provide context-specific execution:

#### Variable Catalog

| Variable | Substitution | Example | Usage |
|----------|--------------|---------|-------|
| `{mission_id}` | Mission identifier (e.g., `MISSION-42`) | `MISSION-123` | Naming artifacts |
| `{worktree_dir}` | Path to mission worktree | `/project/.git/worktrees/MISSION-42` | Working directory |
| `{test_file}` | Path to test file for current AC | `/project/src/auth/auth_test.go` | Targeted testing |
| `{test_dir}` | Directory containing test file | `/project/src/auth/` | Directory-scoped testing |
| `{ac_index}` | Acceptance Criteria index | `AC-1`, `AC-2` | Log messages |
| `{ac_title}` | URL-safe version of AC title | `user-can-login` | Artifact naming |
| `{surface_area}` | Comma-separated glob patterns | `src/auth/*,src/auth/**/*.go` | Linting/typecheck scope |
| `{project_root}` | Original project root (not worktree) | `/project/` | Absolute paths |
| `{timestamp}` | ISO 8601 timestamp | `2026-02-10T14:30:00Z` | Logging |
| `{agent_id}` | Agent identifier | `ensign-42` | Audit trail |
| `{harness}` | Harness type | `claude`, `codex` | Harness-specific logic |

#### Substitution Rules

1. **Undefined variable** → Error (fail fast)
2. **Double braces** → Literal single brace: `{{` → `{`
3. **Syntax** → `{variable_name}`
4. **Scope** → Variables resolved per-mission at gate execution time

#### Example Commands

```json
// ~/.sc3/config.json (TUI-writable)
{
  "gates": {
    // RED phase: Test specific file (targeted)
    "verify_red_command": "go test -v {test_file}",

    // GREEN phase: Test specific file (targeted)
    "verify_green_command": "go test -v {test_file}",

    // REFACTOR phase: Test all (comprehensive)
    "verify_refactor_command": "go test ./...",

    // BUILD: Binary named by mission
    "build_command": "go build -o {worktree_dir}/bin/{mission_id}",

    // LINT: Surface area only (efficient)
    "lint_command": "golangci-lint run {surface_area}",

    // TYPECHECK: Surface area only (efficient)
    "typecheck_command": "go vet {surface_area}",

    // Custom: Run tests with coverage report
    "coverage_command": "go test -coverprofile={worktree_dir}/coverage/{mission_id}.cov {test_dir}"
  }
}
```

### Gate Execution

```go
// Gate execution with variable substitution
type GateExecution struct {
    Command string       // Template with variables
    Context GateContext  // Variable values
}

type GateContext struct {
    MissionID    string
    WorktreeDir  string
    TestFile     string
    TestDir      string
    ACIndex      string
    ACTitle      string
    SurfaceArea  []string
    ProjectRoot  string
    AgentID      string
    Harness      string
}

func ExecuteGate(ctx GateContext) (GateResult, error) {
    // 1. Substitute variables
    command := SubstituteVariables(ctx.Command, ctx)

    // 2. Execute in worktree directory
    cmd := exec.Command("sh", "-c", command)
    cmd.Dir = ctx.WorktreeDir

    // 3. Capture output
    output, err := cmd.CombinedOutput()

    // 4. Return result
    return GateResult{
        ExitCode: cmd.ProcessState.ExitCode(),
        Output:   string(output),
        Success:  err == nil,
    }, nil
}
```

### Gate Evidence

Gate results are stored in Beads as evidence:

```go
type GateResult struct {
    GateID       string    // Unique gate execution ID
    MissionID    string    // Mission this gate validated
    Phase        string    // Phase: "VERIFY_RED" | "VERIFY_GREEN" | etc.
    Command      string    // Command executed (after substitution)
    ExitCode     int       // Shell exit code
    Output       string    // Captured stdout/stderr
    Success      bool      // Exit code == 0
    Timestamp    time.Time // Execution time
}
```

### Gate Configuration

Gates are configured per-phase to support different verification strategies:

| Phase | Required Gates | Optional Gates | Skip Condition |
|-------|----------------|----------------|----------------|
| **RED** | `verify_red` | `typecheck` | Test already exists |
| **GREEN** | `verify_green` | `typecheck`, `lint` | N/A |
| **REFACTOR** | `verify_refactor` | `typecheck`, `lint`, `build` | No refactor performed |
| **COMPLETE** | `build` | `lint`, `typecheck` | N/A |

---

## Protocol Versioning

Ship Commander 3 uses semantic versioning for protocol events to ensure backward and forward compatibility as the system evolves.

### Versioning Rules

1. **MAJOR version** (1.0 → 2.0): Breaking changes, removed fields
2. **MINOR version** (1.0 → 1.1): Additive changes, backward compatible
3. **PATCH version** (1.0.0 → 1.0.1): Bug fixes, no schema changes

### Compatibility Policy

- **Reader supports**: Current MINOR version + 2 prior MINOR versions
  - Example: Reader v1.3 reads v1.1, v1.2, v1.3
- **Writer writes**: Current MINOR version
- **Forward compatibility**: Reader ignores unknown fields
- **Backward compatibility**: Writer includes all fields from oldest supported MINOR

### Event Schema Evolution

```go
type Event struct {
    ProtocolVersion string `json:"protocol_version"` // "1.1"
    EventType       string `json:"event_type"`

    // v1.0 fields (always present for backward compat)
    MissionID       string `json:"mission_id,omitempty"`
    AgentID         string `json:"agent_id,omitempty"`
    Phase           string `json:"phase,omitempty"`

    // v1.1 fields (added, optional)
    WaveID          string `json:"wave_id,omitempty"` // NEW in v1.1

    // v1.0 fields (continued)
    Timestamp       time.Time       `json:"timestamp"`
    Description     string          `json:"description"`
    Evidence        json.RawMessage `json:"evidence"`
}
```

### Version Negotiation

- On startup, components broadcast supported version range
- Components agree on minimum common version
- Event bus validates event versions and rejects unsupported versions

### Deprecation Policy

- MINOR versions deprecated after 2 releases
- MAJOR versions require coordinated upgrade

---

## Doctor (Health Monitor)

The Doctor is a background goroutine that monitors system health and detects anomalies requiring intervention. It runs on a ticker interval and queries Beads for inconsistencies.

### Detection Capabilities

| Detection | Description | Event Emitted |
|-----------|-------------|---------------|
| **Stuck Agent** | Agent session exceeds expected duration without state change | `STUCK_AGENT_DETECTED` |
| **Orphan Mission** | Mission in `in_progress` without active agent session | `ORPHAN_MISSION_FOUND` |
| **Heartbeat Failure** | Agent fails to send heartbeat within timeout window | `HEARTBEAT_FAILED` |
| **tmux Orphan** | tmux session exists without corresponding Beads agent record | `TMUX_ORPHAN_FOUND` |
| **Lock Expiry** | Surface-area lock nearing expiration without renewal | `LOCK_EXPIRING_SOON` |

### Remediation Policy

The Doctor follows policy-based remediation to automatically resolve detected issues:

```go
// Remediation Policy - Add to TRD

type RemediationAction string

const (
    AlertOnly       RemediationAction = "alert"        // Just emit event
    KillAgent       RemediationAction = "kill"         // Terminate agent
    RestartAgent    RemediationAction = "restart"      // Spawn new agent
    ReturnMission   RemediationAction = "return"       // Return to backlog
    RequestHelp     RemediationAction = "request_help" // Ask Admiral
)
```

### Policy Configuration

Remediation actions are configurable via global config:

```toml
# ~/.sc3/config.json (TUI-writable)
{
  "doctor": {
    "remediation": {
      "stuck_agent": "request_help",
      "orphaned_mission": "return",
      "heartbeat_failure": "kill",
      "tmux_orphan": "kill"
    },
    "detection_intervals": {
      "stuck_agent_seconds": 600,
      "orphan_mission_seconds": 300,
      "heartbeat_timeout_seconds": 120
    }
  }
}
```

### Remediation State Machine

```
STUCK_AGENT_DETECTED
  → if policy == kill: Kill agent, emit AGENT_TERMINATED
  → if policy == restart: Kill + spawn new, emit AGENT_RESTARTED
  → if policy == return: Kill + return to backlog, emit MISSION_RETURNED
  → if policy == request_help: Emit event, wait for Admiral
```

### Admiral Controls (TUI)

When `request_help` policy is triggered, TUI presents Admiral with remediation options:

```
┌─────────────────────────────────────────────────────────┐
│  STUCK AGENT DETECTED                                   │
│                                                         │
│  Agent: ensign-42 (implementer)                         │
│  Mission: MISSION-123                                   │
│  Duration: 45 minutes (expected: 15)                    │
│                                                         │
│  [k] Kill agent and return mission to backlog          │
│  [r] Restart agent with fresh context                   │
│  [i] Ignore (continue waiting)                          │
│  [?] View agent output log                              │
└─────────────────────────────────────────────────────────┘
```

### Heartbeat Monitoring

All agents send heartbeat events every 30 seconds:

```go
// Agent heartbeat event
type HeartbeatEvent struct {
    ProtocolVersion string    `json:"protocol_version"`
    EventType       string    `json:"event_type"` // "HEARTBEAT"
    AgentID         string    `json:"agent_id"`
    MissionID       string    `json:"mission_id"`
    Timestamp       time.Time `json:"timestamp"`
    CurrentPhase    string    `json:"current_phase"`
    LastProgress    string    `json:"last_progress"`
}
```

Doctor tracks last heartbeat timestamp per agent and emits `HEARTBEAT_FAILED` if timeout exceeded.

---

## UI State Persistence (Non-Beads)

TUI navigation and display state are persisted separately from domain state (Beads) to enable crash recovery and session restoration.

### TUI State File: `~/.sc3/tui-state.json`

TUI session state is ephemeral UI position, not system state:

```json
{
  "version": "1.0",
  "session_id": "uuid",
  "last_view": "mission-detail",
  "last_view_params": {
    "mission_id": "MISSION-42"
  },
  "nav_stack": ["fleet-overview", "ship-bridge"],
  "scroll_positions": {
    "event_log": 500,
    "ac_phase_list": 3
  },
  "focused_panel": "ac-phase-detail",
  "layout_mode": "standard",
  "display_mode": "advanced",
  "timestamp": "2026-02-10T14:30:00Z"
}
```

### Crash Recovery Behavior

On TUI startup:

1. **Check for recovery file**: If `~/.sc3/tui-state.json` exists < 1 hour old
2. **Offer restoration**: Prompt Admiral: "Restore previous session? [Y/n]"
3. **Restore on accept**: Rebuild navigation stack, scroll positions, selections
4. **Fresh start on decline/no file**: Start at Fleet Overview view

### What NOT to Persist in TUI State

TUI state does NOT duplicate domain state stored in Beads:

| Setting | Storage | Rationale |
|---------|---------|-----------|
| Commission/mission/agent state | Beads | Source of truth for domain |
| Event log | Event Bus (replay) | Replay from subscription |
| Agent status | Beads | Query current state |
| Navigation stack | `tui-state.json` | Ephemeral session position |
| Scroll positions | `tui-state.json` | Session-specific |
| Focused panel | `tui-state.json` | Session-specific |
| Layout mode | `tui-state.json` | Session-specific |
| Display mode (advanced/simple) | `~/.sc3/config.json` | User preference (global) |
| TUI preferences (animations, bell) | `~/.sc3/config.json` | User preference (global) |
| Accessibility mode | `~/.sc3/config.json` | User preference (global) |

### TUI State vs Global Config

**Rule of thumb**: If it affects **where** the user is in the UI → `tui-state.json`. If it affects **how** the UI behaves → `~/.sc3/config.json`.

### State Write Frequency

TUI state is written on:

- Navigation view changes (push to nav stack)
- Scroll position changes (debounced, 100ms)
- Panel focus changes
- TUI mode changes (standard/compact, advanced/simple)
- Session shutdown (graceful exit)

### Concurrency Considerations

- Single TUI instance expected per user
- File locked during write to prevent corruption
- Stale sessions (>1 hour old) ignored on startup

---

## Execution Labels (Process Classification)

Every process and step in Ship Commander 3 falls into one of three categories:

| Label | Meaning | Example |
|-------|---------|---------|
| **`[Harness → Agent]`** | AI agent invoked via a coding harness session. The harness and model are interchangeable via config. | Captain analyzing functional requirements |
| **`[Conditional Logic]`** | Deterministic code: state machine transitions, shell execution, exit code checking, message routing, lock management. No AI. | VERIFY_RED gate running `go test` and checking exit code |
| **`[Human Input → Admiral]`** | System presents information to the human and waits for a decision. | Admiral reviewing mission manifest for approval |

**Key invariant**: All `[Harness → Agent]` output must pass through `[Conditional Logic]` before affecting system state. No probabilistic output directly transitions state.

---

## Demo Token V1 Specification

### File Convention

```
demo/MISSION-<id>.md
```

One file per mission, committed to the mission's worktree branch.

### Schema

```markdown
---
mission_id: "MISSION-42"
title: "Add login button to auth page"
classification: "RED_ALERT"
status: "complete"
created_at: "2026-02-10T14:30:00Z"
agent_id: "implementer-1"
---

## Evidence

### commands

- `go test ./auth/...`
  - exit_code: 0
  - summary: "3 tests passed"

### tests

- file: `auth/login_test.go`
  - added_tests:
    - "TestLoginButton_Renders"
    - "TestLoginButton_SubmitsCredentials"
    - "TestLoginButton_ValidatesEmptyFields"
  - passing: true

### manual_steps

1. Run `go run ./cmd/server`
2. Navigate to `/auth`
3. Observe login button in top-right corner

### diff_refs

- `auth/login.go` — added LoginButton handler (lines 12-45)
- `auth/login_test.go` — added 3 test cases (lines 8-52)
```

### Allowed Evidence Types (V1)

| Type | Required Fields | Purpose |
|------|----------------|---------|
| `commands` | command, `exit_code`, `summary` | Prove a shell command ran successfully |
| `tests` | `file`, `added_tests[]`, `passing` | Prove tests exist and pass |
| `manual_steps` | Numbered step list | Human-walkable verification path |
| `diff_refs` | File path + description | Link proof to actual code changes |

### Mode-Dependent Requirements

| Classification | Required Sections |
|---------------|-------------------|
| `RED_ALERT` | `tests` + at least one of (`commands`, `diff_refs`) |
| `STANDARD_OPS` | At least one of (`commands`, `manual_steps`, `diff_refs`) |

---

## Event Bus Architecture

### Event Catalog (V1)

Ship Commander 3 uses a comprehensive event catalog for real-time system observability and coordination.

#### Agent Lifecycle Events

| Event Type | Description |
|------------|-------------|
| `AGENT_CLAIM` | Agent completes phase (RED_COMPLETE, GREEN_COMPLETE, etc.) |
| `AGENT_SPAWNED` | Agent session started |
| `AGENT_STUCK` | Stuck detection triggered |
| `AGENT_TIMEOUT` | Gate timeout occurred |
| `AGENT_TERMINATED` | Agent session ended |

#### Mission Lifecycle Events

| Event Type | Description |
|------------|-------------|
| `MISSION_DISPATCHED` | Mission dispatched to agent |
| `MISSION_REVIEW_REQUESTED` | Mission ready for review |
| `MISSION_APPROVED` | Review approved, mission complete |
| `MISSION_NEEDS_FIXES` | Review requires changes |
| `MISSION_HALTED` | Mission halted (termination conditions met) |

#### Wave Lifecycle Events

| Event Type | Description |
|------------|-------------|
| `WAVE_STARTED` | Wave execution started |
| `WAVE_COMPLETE` | Wave execution completed |
| `WAVE_REVIEW_REQUESTED` | Wave ready for Admiral review |
| `WAVE_FEEDBACK_PROVIDED` | Admiral provided feedback |

#### Gate Execution Events

| Event Type | Description |
|------------|-------------|
| `GATE_RESULT` | Deterministic gate result |
| `GATE_FAILURE` | Gate failed with error |

#### Planning Events

| Event Type | Description |
|------------|-------------|
| `READY_ROOM_MESSAGE` | Inter-agent message in Ready Room |
| `PLANNING_ITERATION_ADVANCED` | Planning loop iteration advanced |
| `CONSENSUS_REACHED` | All agents signed off on manifest |

#### Admiral Interaction Events

| Event Type | Description |
|------------|-------------|
| `ADMIRAL_QUESTION` | Question surfaced to Admiral |
| `ADMIRAL_ANSWER` | Admiral answered question |
| `PLAN_APPROVAL` | Admiral approved manifest |
| `PLAN_FEEDBACK` | Admiral provided feedback on manifest |

#### State Transition Events

| Event Type | Description |
|------------|-------------|
| `STATE_TRANSITION` | Legal state transition completed |
| `STATE_TRANSITION_REJECTED` | Illegal transition attempted and rejected |

#### Health Monitoring Events

| Event Type | Description |
|------------|-------------|
| `STUCK_AGENT_DETECTED` | Doctor detected stuck agent |
| `ORPHAN_MISSION_FOUND` | Doctor found orphaned mission |
| `HEARTBEAT_FAILED` | Agent heartbeat failed |

#### System Events

| Event Type | Description |
|------------|-------------|
| `LOCK_CONFLICT` | Surface-area lock conflict detected |
| `LOCK_RENEWED` | Lock renewed via heartbeat or explicit request |
| `HARNESS_UNAVAILABLE` | Harness CLI not found on PATH |
| `ERROR` | General system error |

### Event Subscription Model

```go
// Event Subscription API
type EventFilter struct {
    CommissionID string   // Filter to commission
    MissionID    string   // Filter to mission
    AgentID      string   // Filter to agent
    EventType    string   // Filter to event type
    Since        time.Time // Replay since timestamp
}

type EventBus interface {
    Publish(event Event) error
    Subscribe(filter EventFilter) <-chan Event
    Unsubscribe(sub <-chan Event)
    Replay(filter EventFilter) ([]Event, error)
}
```

### Delivery Guarantees

- **Best effort**: Events delivered if subscriber keeps up
- **No backpressure**: Oldest events dropped from buffer if slow
- **Replay on reconnect**: Subscriber can request missed events on reconnect

### Event Bus Configuration

```toml
[event_bus]
max_buffer_size = 1000  # Max events buffered per subscriber
replay_limit = 10000    # Max events returned by Replay()
```

---

## Harness Abstraction Layer

Ship Commander 3 uses a standardized harness abstraction layer that provides a unified interface for AI coding CLIs (Claude Code and Codex).

### Harness Interface

```go
// Harness interface - common operations for all AI coding CLIs
type Harness interface {
    // SpawnSession creates a new coding session
    SpawnSession(ctx context.Context, req SessionRequest) (*Session, error)

    // SendMessage sends additional context to running session
    SendMessage(ctx context.Context, sessionID string, message string) error

    // GetSession retrieves session status
    GetSession(ctx context.Context, sessionID string) (*SessionStatus, error)

    // Terminate ends a session
    Terminate(ctx context.Context, sessionID string) error

    // StreamOutput captures real-time output
    StreamOutput(ctx context.Context, sessionID string) (<-chan OutputChunk, error)

    // GetModelCatalog returns available models
    GetModelCatalog() ModelCatalog
}
```

### Session Request Configuration

```go
// SessionRequest defines session configuration
type SessionRequest struct {
    // Model selection
    Model string `json:"model"`

    // Skills integration (Claude Code only)
    Skills []SkillConfig `json:"skills,omitempty"`

    // Working directory
    WorkingDir string `json:"workingDir,omitempty"`

    // Timeout configuration
    Timeout time.Duration `json:"timeout,omitempty"`

    // System prompt
    SystemPrompt string `json:"systemPrompt,omitempty"`

    // Initial message/prompt
    Prompt string `json:"prompt"`

    // Session-specific flags
    Flags map[string]interface{} `json:"flags,omitempty"`

    // Role context
    Role string `json:"role"` // captain|commander|design_officer|implementer|reviewer

    // Mission assignment
    MissionID string `json:"missionID"`
}

// SkillConfig for Claude Code skills
type SkillConfig struct {
    Name string `json:"name"`
    Path string `json:"path"`
    Description string `json:"description"`
    Parameters map[string]string `json:"parameters"`
}
```

### Model Catalog

Ship Commander 3 uses role-based model selection to optimize for capability vs cost.

#### Model Capability Types

- `code_generation` - Generate implementation code
- `code_review` - Review code for correctness and style
- `system_design` - Design system architecture
- `debugging` - Debug issues and errors
- `documentation` - Generate documentation
- `testing` - Write and understand tests
- `refactoring` - Refactor code for quality
- `requirements_analysis` - Analyze requirements

#### Claude Code Models

| Model | Capabilities | Recommended For | Max Tokens | Cost/1K Tokens |
|-------|--------------|----------------|------------|---------------|
| **opus** | All capabilities | captain, design_officer | 200K | $0.015 |
| **sonnet** | code, review, debug, test, refactor, docs | commander, implementer (backend/frontend) | 200K | $0.003 |
| **haiku** | code, review, test | reviewer | 200K | $0.00025 |

#### OpenAI Codex Models

| Model | Capabilities | Recommended For | Max Tokens | Cost/1K Tokens |
|-------|--------------|----------------|------------|---------------|
| **gpt-4** | All capabilities | implementer, complex_tasks | 128K | $0.03 |
| **gpt-3.5-turbo** | code, review, test | reviewer, simple_implementation, backend | 16K | $0.0015 |

#### Role-to-Model Mappings

| Role | Recommended Model | Rationale |
|------|------------------|-----------|
| Captain | opus | Most capable for complex analysis and planning |
| Commander | sonnet | Balanced for orchestration tasks |
| Design Officer | opus | Design requires highest capability |
| Implementer | sonnet | Balanced capability for most development |
| Reviewer | haiku | Fast for code review |

### Configuration Resolution

**Agent configuration is resolved from Agent beads, not config files**:

```go
// When spawning an agent for a mission
agentBead, _ := beads.Get(agentID)

harness := agentBead.Labels["harness"]  // "claude" or "codex"
model := agentBead.Labels["model"]      // "opus", "sonnet", "haiku", etc.

// Spawn agent session with resolved config
session := harness.SpawnSession(ctx, SessionRequest{
    Model: model,
    // ... other params
})
```

**Agent Bead Creation Example**:

```go
// Create Captain agent (high-capability model for planning)
captainAgent := beads.Create(beads.Bead{
    Type: "agent",
    Labels: map[string]string{
        "role":    "captain",
        "harness": "claude",
        "model":   "opus",
        "ship_id": shipID,
    },
})

// Create Implementer agent (balanced model for development)
implementerAgent := beads.Create(beads.Bead{
    Type: "agent",
    Labels: map[string]string{
        "role":    "implementer",
        "harness": "claude",
        "model":   "sonnet",
        "ship_id": shipID,
        "domain":  "backend",  // Optional domain specialization
    },
})

// Create Reviewer agent (fast model for code review)
reviewerAgent := beads.Create(beads.Bead{
    Type: "agent",
    Labels: map[string]string{
        "role":    "reviewer",
        "harness": "claude",
        "model":   "haiku",
        "ship_id": shipID,
    },
})
```

**Role-to-Model Recommendations** (for Agent bead creation):

| Role | Recommended Harness | Recommended Model | Rationale |
|------|-------------------|-------------------|-----------|
| Captain | claude | opus | Most capable for complex analysis and planning |
| Commander | claude | sonnet | Balanced for orchestration tasks |
| Design Officer | claude | opus | Design requires highest capability |
| Implementer | claude/codex | sonnet | Balanced capability for most development |
| Reviewer | claude/codex | haiku | Fast for code review |

### Driver Implementations

#### Claude Code Driver

```go
type ClaudeCodeDriver struct {
    config        HarnessConfig
    modelCatalog  ModelCatalog
    skillsManager *SkillsManager
    tmuxManager   *tmux.Manager
}

func (d *ClaudeCodeDriver) SpawnSession(ctx context.Context, req SessionRequest) (*Session, error) {
    // Build command:
    // claude -p "prompt" --model <model> --verbose --timeout <timeout> \
    //   --add-dir <workdir> --system-prompt <prompt> --agents <skills-json>

    // Create tmux session: sc3-<role>-<mission-id>
    // Send command to tmux
    // Capture output via tmux capture-pane
    // Return session object
}
```

#### Codex Driver

```go
type CodexDriver struct {
    config       HarnessConfig
    modelCatalog ModelCatalog
    tmuxManager  *tmux.Manager
}

func (d *CodexDriver) SpawnSession(ctx context.Context, req SessionRequest) (*Session, error) {
    // Build command:
    // codex exec - --model <model> --timeout <timeout> \
    //   --cd <workdir> --sandbox <mode> --approval <policy>

    // Create tmux session: sc3-<role>-<mission-id>
    // Send command to tmux
    // Capture output via tmux capture-pane
    // Return session object
}
```

### CLI Flag Mapping

| Feature | Claude Code | Codex | SessionRequest Field |
|---------|-------------|-------|---------------------|
| Model | `--model` | `--model` | `Model` |
| Prompt | `-p` | stdin | `Prompt` |
| Workdir | `--add-dir` | `--cd` | `WorkingDir` |
| Timeout | `--timeout` | `--timeout` | `Timeout` |
| Skills | `--agents` | N/A | `Skills` |
| Sandbox | `--tools` | `--sandbox` | `Flags["sandbox"]` |

### Ensign Pool Management

The Ensign Pool manages isolation between implementer and reviewer agents to ensure context separation (PRD requirement: implementer ≠ reviewer for each mission).

```go
// Ensign - Agent worker instance
type Ensign struct {
    ID          string
    Name        string
    Domains     []string // ["backend", "frontend", "fullstack"]
    Harness     string   // "claude" | "codex"
    Model       string
    CurrentRole string   // "implementer" | "reviewer" | "idle"
    CurrentMission string | null
}

// EnsignPool - Manages available ensigns
type EnsignPool struct {
    Ensigns []Ensign
}

func (p *EnsignPool) SelectImplementer(domain string) (*Ensign, error)
func (p *EnsignPool) SelectReviewer(domain string, excludeID string) (*Ensign, error)
```

### Isolation Guarantees

The Commander enforces that implementer and reviewer for a given mission are different ensigns:

1. **Different tmux sessions**: `sc3-impl-<mission>` vs `sc3-rev-<mission>`
2. **Different working directories**:
   - Implementer gets writable worktree: `.git/worktrees/MISSION-<id>`
   - Reviewer gets read-only copy or original project root
3. **Different context prompts**:
   - Implementer: Receives mission spec + prior wave feedback
   - Reviewer: Receives diff summary + gate evidence + test results
4. **Session ID validation**: Commander validates `implementer_id != reviewer_id` before spawning reviewer

### Fallback Policy

When only 1 ensign is available (e.g., single-user local development):

```toml
# ~/.sc3/config.json (TUI-writable)
{
  "roles": {
    "ensign": {
      "allow_self_review": false,  # Enforce PRD requirement
      "fallback_strategy": "block"  # | "warn" | "allow_generic"
    }
  }
}
```

**Fallback Strategies**:

| Strategy | Behavior | Use Case |
|----------|----------|----------|
| `block` | Fail mission until second ensign available | Production, multi-ensign environments |
| `warn` | Show Admiral warning, allow self-review | Single-user development (dangerous) |
| `allow_generic` | Use generic haiku reviewer for all missions | Demo mode (not production) |

### Domain Assignment

Ensigns are assigned to missions based on domain expertise:

| Mission Domain | Required Ensign Domains | Example |
|----------------|------------------------|---------|
| Backend | `backend` or `fullstack` | API development, database changes |
| Frontend | `frontend` or `fullstack` | UI components, React views |
| Fullstack | `fullstack` | End-to-end features |
| Infrastructure | `backend` or `fullstack` | DevOps, CI/CD, config |

### Session Lifecycle

```
Mission Dispatched
  → SelectImplementer(domain) → ensign_42
  → SpawnSession(ensign_42, implementer)
  → [Implementer works...]
  → Mission Complete
  → SelectReviewer(domain, exclude=ensign_42) → ensign_73
  → SpawnSession(ensign_73, reviewer)
  → [Reviewer validates...]
  → Review Complete
  → Both ensigns returned to pool (role: idle)
```

---

## Execution Labels (Process Classification)