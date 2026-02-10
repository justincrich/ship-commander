# Ship Commander TUI Design Specification

**Version**: 1.0
**Date**: 2025-02-09
**Designer**: TUI Designer Agent
**Status**: Draft - Pending Product Manager Review

---

## Table of Contents

1. [Design Vision](#design-vision)
2. [Visual Language](#visual-language)
3. [User Flows](#user-flows)
4. [Layout Specifications](#layout-specifications)
5. [Component Design](#component-design)
6. [Interaction Patterns](#interaction-patterns)
7. [Implementation Guide](#implementation-guide)
8. [Accessibility Considerations](#accessibility-considerations)

---

## Design Vision

### Core Philosophy

**"Make it feel like a fun Star Trek video game while maintaining productivity"**

The TUI should balance:
- **Playful immersion**: Star Trek LCARS aesthetic with terminal-friendly adaptations
- **Information density**: Maximize useful data display without overwhelming
- **Operational clarity**: Clear status, actions, and feedback at all times
- **Responsive interaction**: Fast keyboard-driven navigation with minimal latency

### Target Experience

The operator should feel like they're:
1. Standing on a starship bridge monitoring missions
2. Coordinating specialist roles (Captain, Commander, Design Officer, Synthesizer)
3. Controlling agent teams like away teams
4. Making strategic decisions about resource allocation (missions, WIP limits)

### Design Principles

1. **Color-Coded Status**: Use LCARS color palette for instant recognition
2. **Spatial Layout**: Fixed panels for muscle memory (same place, same function)
3. **Progressive Disclosure**: Show overview first, drill down on demand
4. **Terminal-Aesthetic**: Work within terminal constraints (no rounded corners, limited colors)
5. **Star Trek Terminology**: Use mission/role-based language throughout

---

## Visual Language

### Color Palette (LCARS-Inspired, Terminal-Friendly)

Based on [LCARS Color Guide](https://www.thelcars.com/colors.php) and adapted for 256-color terminals:

```typescript
// LCARS-inspired terminal colors
const LCARS_COLORS = {
  // Primary colors (bold/bright for emphasis)
  ORANGE: '#FF9966',      // Primary actions, active elements
  BLUE: '#9999CC',         // Standard information, status
  PURPLE: '#CC99CC',       // Secondary information, warnings
  PINK: '#FF99CC',         // Alerts, notifications

  // System colors (terminal-compatible)
  RED_ALERT: '#FF3333',    // Critical errors, halted missions
  YELLOW_CAUTION: '#FFCC00', // Warnings, stuck agents
  GREEN_OK: '#33FF33',     // Success, completed, healthy

  // Backgrounds
  BLACK: '#000000',        // Primary background
  DARK_BLUE: '#1B4F8F',    // Panel backgrounds
  GRAY: '#333333',         // Inactive/disabled elements

  // Text
  WHITE: '#FFFFFF',        // Primary text
  LIGHT_GRAY: '#CCCCCC'    // Secondary text
} as const;
```

#### Semantic Color Mapping

| Status/State | LCARS Color | Usage |
|-------------|-------------|-------|
| **Active/Running** | Orange | Agents working, missions in progress |
| **Success/Done** | Green | Completed directives, passed gates |
| **Error/Halted** | Red Alert | Failed gates, halted missions, critical errors |
| **Warning/Stuck** | Yellow Caution | Stuck agents, doctor intervention needed |
| **Information** | Blue | General status, agent roles, metadata |
| **Planning/Review** | Purple | Synthesis, review phases, pending approval |
| **Notification** | Pink | System messages, alerts, feedback |

### Typography

Terminal fonts are monospaced by nature. Use:
- **Bold** for headers, active elements, emphasis
- **Dim** for secondary information, disabled items
- **Underline** for links, interactive elements (keyboard-driven)
- **Reverse video** for selected/focused items

```
BOLD_WHITE     = Major headers, critical status
BOLD_ORANGE    = Active items, current phase
BOLD_BLUE      = Section headers
DIM_GRAY       = Metadata, timestamps
UNDERLINE      = Interactive elements, keyboard shortcuts
REVERSE        = Selected item, focused panel
```

### Layout Patterns

#### Terminal-Friendly LCARS Adaptations

Traditional LCARS uses curved elbow segments. In terminal, we approximate with:

```
┌─────────────────────────────────────────────────────────┐
│ SHIP COMMANDER │ project=default │ Health: ████░░░ 80% │
├─────────────────────────────────────────────────────────┤
│ ┌─ Mission Control ──────────────────────────────────┐ │
│ │ Agents (3)     │ Mission Board                     │ │
│ │ ├─ cmdr-abc    │ B:2 IP:3 R:1 D:5 H:0             │ │
│ │ ├─ impl-def    │ Wave 1: [████░░] 2/4 done        │ │
│ │ └─ rev-ghi     │ Wave 2: [██░░░░] 1/4 done        │ │
│ └────────────────────────────────────────────────────┘ │
│ ┌─ Tactical Display ──────────────────────────────────┐ │
│ │ Phase Tracker  │ Wave Summary                      │ │
│ │ RED→GREEN→OK  │ w1-d1 started  │                  │ │
│ │ VERIFY_PASS    │ w1-d2 started  │                  │ │
│ └────────────────────────────────────────────────────┘ │
│ ┌─ Engineering Logs ──────────────────────────────────┐ │
│ │ [INFO] propulsion.dispatch wave=1 directive=d1     │ │
│ │ [WARN] doctor.stuck agent=cmdr-abc timeout=300s    │ │
│ └────────────────────────────────────────────────────┘ │
│ ┌─ Command Interface ─────────────────────────────────┐ │
│ │ > approve d1                                        │ │
│ │ Commands: halt|retry|approve|max_missions|quit      │ │
│ └────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────┘
```

**Key Terminal Constraints:**
- Use box-drawing characters (`┌─┐│├┤└┘`) for panel borders
- No actual curved corners (use standard ASCII/Unicode box chars)
- Maximize horizontal space (panels side-by-side)
- Vertical stacking for temporal/event information
- Consistent panel sizes for muscle memory

---

## User Flows

### Flow 1: Runtime Monitoring (Captain's Chair)

**Purpose**: Operator monitors active missions, agents, and system health in real-time

**Entry Point**: `ship-commander tui --project default` or `ship-commander start --with-tui`

#### Screens & Transitions

```
┌─────────────────────────────────────────────────────────┐
│                    MAIN BRIDGE VIEW                      │
└─────────────────────────────────────────────────────────┘
                            │
            ┌───────────────┼───────────────┐
            │               │               │
    [1] Health View  [2] Mission View  [3] Agent Detail
            │               │               │
            └───────────────┴───────────────┘
                            │
                    [Tab] Cycle Panels
                    [Enter] Drill Down
                    [q] Quit
```

#### Detailed Flow: Health View → Agent Detail

```
MAIN BRIDGE (Health Summary)
├─ Shows: Missions 3/3 | Stuck: 1 | Doctor: degraded
├─ Operator notices: "stuck=1 doctor=degraded"
├─ Action: Press [a] to focus Agents panel
│
├─ AGENTS PANEL (List View)
│   ├─ Shows: All active agents with status
│   ├─ Operator sees: "cmdr-abc STATUS: STUCK phase=RED timeout=300s"
│   ├─ Action: Press [Enter] on cmdr-abc
│   │
│   ├─ AGENT DETAIL (Drill-down)
│   │   ├─ Shows: Full agent session details
│   │   │   ├─ Directive ID, role, harness, model
│   │   │   ├─ Current phase, elapsed time
│   │   │   ├─ Recent output (last 10 lines)
│   │   │   ├─ Error context (if stuck)
│   │   │   └─ Actions: [h]alt [r]etry [i]gnore
│   │   │
│   │   ├─ Operator Action: Press [r] to retry
│   │   │
│   │   └─ Return to AGENTS PANEL (refreshed)
│   │       └─ "cmdr-abc STATUS: RUNNING phase=RED"
│   │
│   └─ Operator Action: Press [Escape] to return to MAIN BRIDGE
│
└─ Health updates to: "Missions 3/3 | Stuck: 0 | Doctor: ok"
```

### Flow 2: Planning & Intake (Mission Briefing)

**Purpose**: Operator ingests PRD, reviews AI-generated plan, approves directive creation

**Entry Point**: `ship-commander intake --prd ./prd.md --with-tui`

#### Screens & Transitions

```
┌─────────────────────────────────────────────────────────┐
│                 MISSION BRIEFING ROOM                    │
└─────────────────────────────────────────────────────────┘
                            │
            ┌───────────────┼───────────────┐
            │               │               │
    [1] PRD Parse    [2] Analysis     [3] Plan Review
            │               │               │
            └───────────────┴───────────────┘
                            │
                    Auto-progress through phases
                    Operator intervention at Plan Review
```

#### Detailed Flow: PRD → Directives

```
PHASE 1: PRD INGEST
├─ Screen: PRD PARSING
│   ├─ Shows: "Parsing PRD: ./prd.md"
│   ├─ Progress bar: [████████░░] 80%
│   ├─ Extracted sections: Product Name, Scope, Use Cases
│   └─ Operator action: Review parsed summary
│       ├─ [Enter] Continue to analysis
│       ├─ [e] Edit PRD (exit TUI, manual edit)
│       └─ [q] Cancel intake
│
PHASE 2: AI ANALYSIS (Parallel Specialists)
├─ Screen: SPECIALIST DASHBOARD
│   ├─ Panel 1: Captain (Triage)
│   │   ├─ Status: ANALYZING...
│   │   ├─ Harness: claude (opus)
│   │   ├─ Elapsed: 00:45
│   │   └─ Output: "Validating PRD completeness..."
│   │
│   ├─ Panel 2: Commander (Technical)
│   │   ├─ Status: ANALYZING...
│   │   ├─ Harness: codex (gpt-5-codex)
│   │   ├─ Elapsed: 00:42
│   │   └─ Output: "Identifying infrastructure tasks..."
│   │
│   ├─ Panel 3: Design Officer (Conditional)
│   │   ├─ Status: SKIPPED (no UI use cases)
│   │   └─ Note: "Captain determined no UI analysis needed"
│   │
│   └─ Panel 4: Synthesizer (Integration)
│       ├─ Status: WAITING (for Captain + Commander)
│       └─ Note: "Will merge specialist outputs"
│
├─ Auto-progress: All specialists complete
│
PHASE 3: PLAN REVIEW (Human Gate)
├─ Screen: PROPOSED DIRECTIVES
│   ├─ Panel 1: Coverage Summary
│   │   ├─ Use Cases: 12 found, 12 covered ✓
│   │   ├─ Directives: 8 proposed
│   │   └─ Waves: 3 execution waves
│   │
│   ├─ Panel 2: Directive List (Scrollable)
│   │   ├─ d1: "Create Beads client adapter" [INFRA] Wave 1
│   │   │   ├─ Dependencies: none
│   │   │   ├─ Ensign: ensign-implementer
│   │   │   └─ ACs: 3 acceptance criteria
│   │   │
│   │   ├─ d2: "Implement control service" [FEATURE] Wave 1
│   │   │   ├─ Dependencies: none
│   │   │   ├─ Ensign: ensign-backend-implementer
│   │   │   └─ ACs: 5 acceptance criteria
│   │   │
│   │   ├─ d3: "Add TUI health panel" [FEATURE] Wave 2
│   │   │   ├─ Dependencies: d1, d2
│   │   │   ├─ Ensign: ensign-ui-implementer
│   │   │   └─ ACs: 4 acceptance criteria
│   │   │
│   │   └─ [d3-detail] Press [Enter] to view full directive
│   │
│   ├─ Panel 3: Dependency Graph (ASCII Tree)
│   │   ├─ Wave 1:
│   │   │   ├─ d1 (infra)
│   │   │   └─ d2 (feature)
│   │   │
│   │   └─ Wave 2:
│   │       └─ d3 (feature)
│   │           ├─ depends on: d1
│   │           └─ depends on: d2
│   │
│   └─ Panel 4: UC Coverage Matrix
│       ├─ UC-INTAKE-01 → d1, d2 ✓
│       ├─ UC-ROLES-02 → d1 ✓
│       ├─ UC-ANALYZE-01 → d1, d2 ✓
│       └─ ... (all 29 UCs mapped)
│
├─ Operator Actions:
│   ├─ [Enter] on directive → View full details (ACs, implementation direction)
│   ├─ [Up/Down] → Scroll directive list
│   ├─ [Tab] → Cycle panels (Coverage → Directives → Dependencies)
│   ├─ [a] Approve all → Create directives, exit to runtime
│   ├─ [r] Request revision → Re-run specialists with feedback
│   ├─ [d] Dry run → Export JSON plan, exit without creating
│   └─ [q] Cancel → Exit without changes
│
└─ On Approve:
    ├─ System: Creates Beads directives
    ├─ System: Wires dependency edges
    ├─ System: Assigns wave numbers
    ├─ Feedback: "✓ Created 8 directives in 3 waves"
    └─ Exit: TUI closes, directives ready for propulsion
```

### Flow 3: Directive Control (Tactical)

**Purpose**: Operator intervenes on specific directives (halt, retry, approve)

**Entry Point**: From MAIN BRIDGE VIEW, navigate to Mission Board panel

#### Detailed Flow: Halt Stuck Directive

```
MAIN BRIDGE (Mission Board Panel)
├─ Shows: "IP: d1,d2,d3 R: d4 H: d5"
├─ Operator notices: d5 in HALTED state (Red Alert color)
├─ Action: Press [Enter] on Mission Board panel
│
├─ MISSION BOARD DETAIL
│   ├─ Panel 1: Backlog (2)
│   ├─ Panel 2: In Progress (3)
│   │   ├─ d1: Implement control service [Running] 00:03:12
│   │   ├─ d2: Create Beads adapter [Running] 00:02:45
│   │   └─ d3: Add TUI health panel [Review] 00:00:30
│   │
│   ├─ Panel 3: Review (1)
│   │   └─ d4: Merge workflow [APPROVED] awaiting operator
│   │
│   └─ Panel 4: Halted (1)
│       └─ d5: Doctor timeout [HALTED] gate failure
│
├─ Operator Action: Press [Enter] on d5 (Halted panel)
│
├─ DIRECTIVE DETAIL VIEW
│   ├─ Header: "Directive d5: Doctor timeout"
│   ├─ Status: HALTED (Red Alert)
│   ├─ Track: INFRA
│   ├─ Ensign: ensign-implementer
│   │
│   ├─ Panel: Recent Events
│   │   ├─ [ERROR] gate.result gate=lint classification=reject_syntax
│   │   ├─ [WARN] doctor.stuck agent=impl-xyz directive=d5
│   │   └─ [INFO] control.halt directive=d5 operator=admiral
│   │
│   ├─ Panel: Gate Results
│   │   ├─ Typecheck: PASS ✓
│   │   ├─ Lint: FAIL ✗ (syntax errors in 3 files)
│   │   └─ Build: SKIPPED
│   │
│   └─ Actions:
│       ├─ [r] Retry → Send back to backlog, restart execution
│       ├─ [h] Halt → Keep halted (manual intervention needed)
│       ├─ [s] Show worktree → Open terminal in directive worktree
│       └─ [Esc] Back to Mission Board
│
├─ Operator Action: Press [r] to retry
│
├─ CONFIRMATION DIALOG
│   ├─ "Retry directive d5? Worktree will be reset."
│   ├─ [y] Yes, retry
│   ├─ [n] Cancel
│   └─ [Esc] Cancel
│
├─ Operator Action: Press [y]
│
└─ RESULT
    ├─ System: Resets directive to backlog
    ├─ System: Cleans worktree
    ├─ Feedback: "✓ Directive d5 retried, moved to backlog"
    └─ Return: Mission Board (d5 now in Backlog panel)
```

### Flow 4: Batch Operations (Wave Commander)

**Purpose**: Operator approves multiple completed directives at once (wave-level merge)

**Entry Point**: Auto-triggered when wave completes

#### Detailed Flow: Batch Wave Approval

```
NOTIFICATION: "Wave 2 Complete: 4 directives ready for merge"
├─ Operator Action: Press [w] to open Wave Manager
│
├─ WAVE MANAGER SCREEN
│   ├─ Panel 1: Wave Overview
│   │   ├─ Wave 1: [████████] 4/4 done (2 merged, 2 pending)
│   │   ├─ Wave 2: [████████] 4/4 done (0 merged, 4 pending)
│   │   └─ Wave 3: [██░░░░░░] 2/4 in progress
│   │
│   ├─ Panel 2: Wave 2 Directives (Pending Merge)
│   │   ├─ d6: Control bead persistence [APPROVED]
│   │   │   ├─ Gates: All pass ✓
│   │   │   ├─ Conflicts: None detected
│   │   │   └─ Ready: Yes
│   │   │
│   │   ├─ d7: Event bus enhancements [APPROVED]
│   │   │   ├─ Gates: All pass ✓
│   │   │   ├─ Conflicts: None detected
│   │   │   └─ Ready: Yes
│   │   │
│   │   ├─ d8: TUI command parser [APPROVED]
│   │   │   ├─ Gates: All pass ✓
│   │   │   ├─ Conflicts: 1 file (src/tui/commands.ts)
│   │   │   └─ Ready: No (manual resolution needed)
│   │   │
│   │   └─ d9: Ink component library [APPROVED]
│   │       ├─ Gates: All pass ✓
│   │       ├─ Conflicts: None detected
│   │       └─ Ready: Yes
│   │
│   ├─ Panel 3: Batch Actions
│   │   ├─ [m] Merge all ready (3 directives)
│   │   ├─ [c] Merge with conflicts (requires resolution)
│   │   ├─ [s] Skip all (approve individually later)
│   │   └─ [Esc] Close Wave Manager
│   │
│   └─ Panel 4: Conflict Preview (if any)
│       ├─ File: src/tui/commands.ts
│       ├─ Lines: 45-52 conflict marker
│       └─ Action: [r] Resolve manually → Open worktree
│
├─ Operator Action: Press [m] to merge all ready
│
├─ MERGE PROGRESS (Animated)
│   ├─ Merging d6... [████████] ✓
│   ├─ Merging d7... [████████] ✓
│   ├─ Skipping d8 (conflicts)... [░░░░░░░░] ⚠
│   ├─ Merging d9... [████████] ✓
│   └─ Post-merge verification... [████████] ✓
│
└─ RESULT
    ├─ System: Merged 3 directives to main branch
    ├─ System: Ran verification gates (all passed)
    ├─ System: Cleaned up worktrees
    ├─ Feedback: "✓ Merged 3/4 Wave 2 directives (1 skipped due to conflicts)"
    └─ Wave 2 panel updates: "Wave 2: [████████] 4/4 done (3 merged, 1 pending)"
```

---

## Layout Specifications

### Screen Architecture

The TUI uses a **fixed panel layout** with **keyboard navigation** between panels.

#### Standard Layout: Main Bridge (Runtime Monitoring)

```
┌─────────────────────────────────────────────────────────────────────────────┐
│ SHIP COMMANDER │ project: default │ Health: ●●●●○ │ Missions: 3/3         │
├─────────────────────────────────────────────────────────────────────────────┤
│ ┌─ Mission Control (40%) ─────────────┐ ┌─ Mission Board (60%) ──────────┐ │
│ │ Active Agents (3)                   │ │ B:2 IP:3 R:1 D:5 H:0          │ │
│ │ ▸ cmdr-abc [d1] impl RED 00:03:12  │ │                                │ │
│ │ ▸ rev-def [d4] rev APPROVED 00:45  │ │ In Progress:                   │ │
│ │ ▸ impl-ghi [d2] impl GREEN 00:02:30│ │ • d1, d2, d3                  │ │
│ │                                    │ │                                │ │
│ │ [Space] to pause/resume            │ │ Review:                       │ │
│ │ [Enter] for agent detail           │ │ • d4                          │ │
│ │                                    │ │                                │ │
│ └────────────────────────────────────┘ │ Done (5):                      │ │
│                                       │ • d5, d6, d7, d8, d9           │ │
│ ┌─ Tactical Display (40%) ───────────┤ │                                │ │
│ │ AC Phase Tracker (6 shown)         │ │ Halted (0):                    │ │
│ │ • d1 RED→GREEN PASS ✓              │ │                                │ │
│ │ • d1 VERIFY_GREEN RUNNING...       │ │                                │ │
│ │ • d2 GREEN→REFACTOR WAIT           │ │ Wave Progress:                 │ │
│ │ • d3 IMPLEMENT DONE ✓              │ │ Wave 1: [████████] 4/4 done   │ │
│ │ • d4 REVIEW APPROVED ✓             │ │ Wave 2: [████░░░░] 2/4 done   │ │
│ │ • d5 HALTED gate:lint FAIL         │ │ Wave 3: [██░░░░░░] 1/4 done   │ │
│ │                                    │ │                                │ │
│ └────────────────────────────────────┘ │ [Enter] for directive detail   │ │
│                                       └─────────────────────────────────┘ │
│ ┌─ Engineering Logs (100%) ──────────────────────────────────────────────┐ │
│ │ [INFO] propulsion.dispatch wave=2 directive=d3 started=2025-02-09T14:30│ │
│ │ [WARN] doctor.stuck agent=cmdr-abc directive=d1 timeout=300s          │ │
│ │ [INFO] gate.result directive=d1 ac=ac1 gate=typecheck classification=accept│ │
│ │ [ERROR] agent.force_stop agent=cmdr-abc directive=d1 reason=timeout   │ │
│ │ [INFO] control.retry directive=d1 operator=admiral                    │ │
│ │ [INFO] agent.started agent=cmdr-retry directive=d1 role=implementer   │ │
│ └──────────────────────────────────────────────────────────────────────┘ │
│ ┌─ Command Interface ───────────────────────────────────────────────────┐ │
│ │ > approve d4                                                          │ │
│ │ Commands: halt|retry|approve|max_missions|wave|quit [Tab] panels     │ │
│ └──────────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────────┘
```

#### Alternative Layout: Planning Dashboard

```
┌─────────────────────────────────────────────────────────────────────────────┐
│ MISSION BRIEFING │ PRD: project-intake-prd.md │ Phase: Synthesis │ 00:02:15│
├─────────────────────────────────────────────────────────────────────────────┤
│ ┌─ PRD Summary (30%) ──────────────┐ ┌─ Specialist Status (70%) ─────────┐│
│ │ Product: Project Intake          │ │ Captain (Triage)                   ││
│ │ Use Cases: 29                    │ │ Status: ✓ DONE                    ││
│ │ Scope: 8 in-scope, 5 out-scope   │ │ Harness: claude (opus)            ││
│ │                                   │ │ Elapsed: 00:00:45                 ││
│ │ Functional Groups:                │ │ Output: "PRD complete, all        ││
│ │ • INTAKE (3 UCs)                 │ │         specialists needed"        ││
│ │ • ROLES (5 UCs)                  │ │                                   ││
│ │ • ANALYZE (6 UCs)                │ │ Commander (Technical)             ││
│ │ • COMMIT (4 UCs)                 │ │ Status: ⏳ WORKING...             ││
│ │ • INTEGRATE (6 UCs)              │ │ Harness: codex (gpt-5-codex)      ││
│ │ • SETTINGS (3 UCs)               │ │ Elapsed: 00:01:32                 ││
│ │ • TUI (2 UCs)                    │ │ Output: "Identifying infra...     ││
│ │                                   │ │         tasks (12 found)"          ││
│ │ [Enter] Full PRD view            │ │                                   ││
│ └───────────────────────────────────┘ │ Design Officer                    ││
│                                       │ Status: ⊘ SKIPPED                 ││
│ ┌─ Coverage Report (100%) ───────────┤ Note: "No UI use cases detected"  ││
│ │ Use Case Coverage: 12/12 (100%) ✓ │ │                                   ││
│ │                                    │ │ Synthesizer (Integration)        ││
│ │ UC-INTAKE-01 → d1, d2 ✓           │ │ Status: ⏸ WAITING               ││
│ │ UC-ROLES-02 → d1 ✓                │ │ Note: "Awaiting Commander output"││
│ │ UC-ANALYZE-01 → d1, d2 ✓          │ │                                   ││
│ │ ...                                │ └───────────────────────────────────┘│
│ │                                    │                                     │
│ │ Gaps: None detected                │ ┌─ Task Inventory (100%) ──────────┐│
│ └────────────────────────────────────┘ │ Proposed Directives: 8           ││
│                                       │ Wave 1 (4 tasks):                 ││
│ ┌─ Dependency Graph (100%) ───────────┤ • d1: Beads client adapter [INFRA]││
│ │ Wave 1:                            │ • d2: Control service [FEATURE]   ││
│ │ ├─ d1 (infra)                     │ • d3: Event bus [FEATURE]         ││
│ │ └─ d2 (feature)                   │ • d4: State persistence [INFRA]   ││
│ │                                    │ │                                   ││
│ │ Wave 2:                            │ Wave 2 (3 tasks):                 ││
│ │ └─ d3 (feature)                   │ • d5: Intake parsing [FEATURE]    ││
│ │     ├─ dep: d1                    │ • d6: Role prompts [FEATURE]      ││
│ │     └─ dep: d2                    │ • d7: AI decomposition [FEATURE]  ││
│ │                                    │ │                                   ││
│ │ Wave 3:                            │ Wave 3 (1 task):                  ││
│ │ └─ d4 (infra)                     │ • d8: TUI planning [FEATURE]      ││
│ │     └─ dep: d3                    │ │                                   ││
│ │                                    │ [Enter] View directive details    ││
│ └────────────────────────────────────┘ └───────────────────────────────────┘│
│ ┌─ Approval Gate ─────────────────────────────────────────────────────────┐│
│ │ Plan Review: 8 directives in 3 waves, 12/12 UCs covered                ││
│ │                                                                         ││
│ │ [a] Approve & Create    [r] Request Revision    [d] Dry Run    [q] Quit││
│ └────────────────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────────────────┘
```

### Panel Sizes & Responsibilities

| Panel | Width | Height | Purpose | Update Frequency |
|-------|-------|--------|---------|------------------|
| **Header** | 100% | 1 line | System status, health summary | On change |
| **Mission Control** | 40% | 8 lines | Active agents list, real-time status | Every 1s |
| **Mission Board** | 60% | 8 lines | Kanban-style directive counts, wave progress | Every 5s |
| **Tactical Display** | 40% | 6 lines | Phase tracker, wave summary | Every 2s |
| **Engineering Logs** | 100% | 8 lines | Event log with severity filtering | Every 1s |
| **Command Interface** | 100% | 2 lines | Command input, help text | On input |

**Total Dimensions**: 80+ columns × 24-30 rows (standard terminal sizes)

### Responsive Behavior

For smaller terminals (< 80 columns):

1. **Collapse side-by-side panels** to vertical stacking
2. **Reduce log lines** from 8 → 4
3. **Simplify header** to one-line status
4. **Hide wave progress** behind a detail view

For larger terminals (> 120 columns):

1. **Add "Directive Details" panel** (rightmost column)
2. **Expand log lines** from 8 → 12
3. **Show full directive titles** instead of truncated IDs
4. **Add keyboard shortcut hints** in panel headers

---

## Component Design

### C1: Status Indicator (LCARS Color Block)

**Purpose**: Instant visual recognition of system/agent/directive status

**Design**:
```
●●●●○ (80% healthy)
```

**Variants**:
- `●●●●○` (Green): All systems normal
- `●●●○☆` (Yellow): 1+ warnings, no errors
- `●●○☆☆` (Orange): 1+ errors, degraded operation
- `●○☆☆☆` (Red): Critical, intervention needed

**Implementation**:
```tsx
interface StatusIndicatorProps {
  health: number; // 0-100
  status: 'ok' | 'warning' | 'error' | 'critical';
}

function StatusIndicator({ health, status }: StatusIndicatorProps): JSX.Element {
  const filled = Math.round(health / 20); // 0-5 dots
  const color = LCARS_COLORS[status.toUpperCase()];
  const empty = 5 - filled;

  return (
    <Text color={color}>
      {'●'.repeat(filled)}
      <Text dimColor>{'○'.repeat(empty)}</Text>
    </Text>
  );
}
```

### C2: Agent Card (Mission Control)

**Purpose**: Display active agent with role, directive, phase, elapsed time

**Design**:
```
▸ cmdr-abc [d1] impl RED 00:03:12
```

**Elements**:
- `▸` (Expand/collapse indicator)
- `cmdr-abc` (Agent ID, truncated to 8 chars)
- `[d1]` (Directive ID in brackets)
- `impl` (Role: implementer, reviewer, commander)
- `RED` (Current TDD phase or status)
- `00:03:12` (Elapsed time MM:SS)

**Color Coding**:
- `RED`, `GREEN`, `REFACTOR`: Phase status (Blue)
- `RUNNING`: Orange
- `DONE`: Green
- `STUCK`: Red Alert

**Implementation**:
```tsx
interface AgentCardProps {
  agentId: string;
  directiveId: string;
  role: 'implementer' | 'reviewer' | 'commander' | 'captain';
  phase: string;
  status: 'running' | 'done' | 'stuck';
  elapsed: number; // seconds
}

function AgentCard({ agentId, directiveId, role, phase, status, elapsed }: AgentCardProps): JSX.Element {
  const color = status === 'stuck' ? LCARS_COLORS.RED_ALERT
             : status === 'done' ? LCARS_COLORS.GREEN_OK
             : LCARS_COLORS.ORANGE;

  const elapsedStr = formatElapsed(elapsed); // "00:03:12"
  const roleCode = ROLE_CODES[role]; // "impl", "rev", "cmdr"

  return (
    <Text>
      <Text color={color}>▸</Text>
      <Text> {agentId.substring(0, 8)} </Text>
      <Text dimColor>[{directiveId}]</Text>
      <Text> {roleCode} </Text>
      <Text bold color={LCARS_COLORS.BLUE}>{phase}</Text>
      <Text dimColor> {elapsedStr}</Text>
    </Text>
  );
}
```

### C3: Mission Board (Kanban Summary)

**Purpose**: Show directive counts per state with drill-down capability

**Design**:
```
B:2 IP:3 R:1 D:5 H:0
```

**Elements**:
- `B` (Backlog): Count of directives awaiting dispatch
- `IP` (In Progress): Currently executing
- `R` (Review): Awaiting human approval
- `D` (Done): Completed and approved
- `H` (Halted): Failed/stuck

**Interactive**: Press `[Enter]` to expand full board

**Expanded View**:
```
┌─ Mission Board Detail ────────────────────────────────────────┐
│ In Progress (3):                                              │
│ • d1: Control service [cmdr-abc] RED 00:03:12                │
│ • d2: Beads adapter [impl-def] GREEN 00:02:45                │
│ • d3: TUI panel [rev-ghi] REVIEW 00:00:30                    │
│                                                               │
│ Review (1):                                                   │
│ • d4: Event bus [APPROVED] awaiting operator                  │
│                                                               │
│ Halted (0):                                                   │
│ • None                                                       │
└───────────────────────────────────────────────────────────────┘
```

**Implementation**:
```tsx
interface MissionBoardProps {
  board: {
    backlog: string[];
    inProgress: string[];
    review: string[];
    done: string[];
    halted: string[];
  };
  waveProgress: Array<{ wave: number; done: number; total: number }>;
}

function MissionBoard({ board, waveProgress }: MissionBoardProps): JSX.Element {
  return (
    <Box flexDirection="column">
      <Text>
        <Text color={LCARS_COLORS.BLUE}>B:{board.backlog.length}</Text>
        {' '}
        <Text color={LCARS_COLORS.ORANGE}>IP:{board.inProgress.length}</Text>
        {' '}
        <Text color={LCARS_COLORS.PURPLE}>R:{board.review.length}</Text>
        {' '}
        <Text color={LCARS_COLORS.GREEN_OK}>D:{board.done.length}</Text>
        {' '}
        <Text color={LCARS_COLORS.RED_ALERT}>H:{board.halted.length}</Text>
      </Text>

      {waveProgress.map(wp => (
        <Text key={wp.wave}>
          Wave {wp.wave}: [{'█'.repeat(wp.done)}{'░'.repeat(wp.total - wp.done)}] {wp.done}/{wp.total}
        </Text>
      ))}
    </Box>
  );
}
```

### C4: Phase Tracker (Gate Results)

**Purpose**: Show recent gate results with directive, AC, and classification

**Design**:
```
• d1 RED→GREEN PASS ✓
• d1 VERIFY_GREEN RUNNING...
• d2 GREEN→REFACTOR WAIT
• d3 IMPLEMENT DONE ✓
• d4 REVIEW APPROVED ✓
• d5 HALTED gate:lint FAIL
```

**Elements**:
- Bullet point (`•`) for visual scanning
- Directive ID (`d1`)
- Phase transition (`RED→GREEN`, `VERIFY_GREEN`, `IMPLEMENT`)
- Result (`PASS`, `RUNNING`, `WAIT`, `DONE`, `APPROVED`, `FAIL`)
- Checkmark (`✓`) or cross (`✗`) for pass/fail

**Color Coding**:
- `PASS`, `DONE`, `APPROVED`: Green
- `RUNNING`: Orange
- `WAIT`: Blue
- `FAIL`, `HALTED`: Red Alert

**Implementation**:
```tsx
interface PhaseTrackerProps {
  events: Array<{
    directiveId: string;
    acId?: string;
    phase: string;
    result: 'pass' | 'fail' | 'running' | 'wait' | 'done';
    gate?: string;
  }>;
}

function PhaseTracker({ events }: PhaseTrackerProps): JSX.Element {
  return (
    <Box flexDirection="column">
      {events.map(event => {
        const color = RESULT_COLORS[event.result];
        const icon = event.result === 'pass' || event.result === 'done' ? '✓'
                  : event.result === 'fail' ? '✗'
                  : '•';

        return (
          <Text key={`${event.directiveId}-${event.acId || 'none'}`}>
            <Text color={color}>{icon} {event.directiveId} {event.phase}</Text>
            {event.gate && <Text dimColor> gate:{event.gate}</Text>}
            <Text> {event.result.toUpperCase()}</Text>
          </Text>
        );
      })}
    </Box>
  );
}
```

### C5: Command Interface (Input + Help)

**Purpose**: Accept operator commands with inline help and feedback

**Design**:
```
> approve d4
Commands: halt|retry|approve|max_missions|wave|quit [Tab] panels
```

**Elements**:
- `>` (Prompt indicator)
- Command input (editable, shows current keystrokes)
- Help text (dimmed, shows available commands)
- Feedback (above input, shows last command result)

**Features**:
- **Tab completion**: Type `app` → `[Tab]` → `approve`
- **Command history**: `[Up/Down]` arrows to cycle previous commands
- **Inline validation**: Red text for unknown commands
- **Panel navigation**: `[Tab]` cycles focus between panels

**Implementation**:
```tsx
interface CommandInterfaceProps {
  command: string;
  feedback: string;
  onSubmit: (cmd: string) => void;
  onChange: (cmd: string) => void;
}

function CommandInterface({ command, feedback, onSubmit, onChange }: CommandInterfaceProps): JSX.Element {
  useInput((input, key) => {
    if (key.return) {
      onSubmit(command);
    } else if (key.backspace || key.delete) {
      onChange(command.slice(0, -1));
    } else if (!key.ctrl && !key.meta && input) {
      onChange(command + input);
    }
  });

  return (
    <Box flexDirection="column">
      {feedback && (
        <Text color={LCARS_COLORS.GREEN_OK}>✓ {feedback}</Text>
      )}
      <Text>
        <Text bold color={LCARS_COLORS.ORANGE}>&gt; </Text>
        <Text>{command}</Text>
        <Text dimColor>_</Text>
      </Text>
      <Text dimColor>
        Commands: halt|retry|approve|max_missions|wave|quit [Tab] panels
      </Text>
    </Box>
  );
}
```

### C6: Specialist Dashboard (Planning)

**Purpose**: Show parallel AI specialist execution with real-time output

**Design**:
```
┌─ Captain (Triage) ────────────────────────────────────┐
│ Status: ✓ DONE                                        │
│ Harness: claude (opus)                                │
│ Elapsed: 00:00:45                                     │
│ Output: "PRD complete, all specialists needed"        │
└───────────────────────────────────────────────────────┘
┌─ Commander (Technical) ───────────────────────────────┐
│ Status: ⏳ WORKING...                                 │
│ Harness: codex (gpt-5-codex)                          │
│ Elapsed: 00:01:32                                     │
│ Output: "Identifying infra tasks (12 found)..."       │
└───────────────────────────────────────────────────────┘
┌─ Design Officer ──────────────────────────────────────┐
│ Status: ⊘ SKIPPED                                     │
│ Note: "No UI use cases detected"                      │
└───────────────────────────────────────────────────────┘
┌─ Synthesizer (Integration) ───────────────────────────┐
│ Status: ⏸ WAITING                                    │
│ Note: "Awaiting Commander output"                     │
└───────────────────────────────────────────────────────┘
```

**Status Icons**:
- `✓` (DONE): Green checkmark
- `⏳` (WORKING): Hourglass (orange)
- `⏸` (WAITING): Paused (blue)
- `⊘` (SKIPPED): Circle with slash (gray)
- `✗` (FAILED): Red cross

**Implementation**:
```tsx
interface SpecialistPanelProps {
  specialist: {
    name: string;
    title: string;
    status: 'done' | 'working' | 'waiting' | 'skipped' | 'failed';
    harness: string;
    model: string;
    elapsed: number;
    output: string;
    note?: string;
  };
}

function SpecialistPanel({ specialist }: SpecialistPanelProps): JSX.Element {
  const statusIcon = STATUS_ICONS[specialist.status];
  const statusColor = STATUS_COLORS[specialist.status];

  return (
    <Box flexDirection="column" borderStyle="single" paddingX={1}>
      <Text bold color={LCARS_COLORS.BLUE}>
        {specialist.title}
      </Text>
      <Text>
        Status: <Text color={statusColor}>{statusIcon} {specialist.status.toUpperCase()}</Text>
      </Text>
      <Text dimColor>
        Harness: {specialist.harness} ({specialist.model})
      </Text>
      <Text dimColor>
        Elapsed: {formatElapsed(specialist.elapsed)}
      </Text>
      <Text>"{specialist.output}"</Text>
      {specialist.note && (
        <Text dimColor>Note: "{specialist.note}"</Text>
      )}
    </Box>
  );
}
```

### C7: Dependency Graph (ASCII Tree)

**Purpose**: Visualize directive dependencies and wave ordering

**Design**:
```
Wave 1:
├─ d1 (infra)
└─ d2 (feature)

Wave 2:
└─ d3 (feature)
    ├─ dep: d1
    └─ dep: d2

Wave 3:
└─ d4 (infra)
    └─ dep: d3
```

**Color Coding**:
- Wave numbers: Blue (bold)
- Directive IDs: Orange
- Types: Purple (infra, feature, design)
- Dependencies: Gray (dim)

**Implementation**:
```tsx
interface DependencyGraphProps {
  waves: Array<{
    wave: number;
    directives: Array<{
      id: string;
      type: 'infra' | 'feature' | 'design';
      dependencies: string[];
    }>;
  }>;
}

function DependencyGraph({ waves }: DependencyGraphProps): JSX.Element {
  return (
    <Box flexDirection="column">
      {waves.map((wave, waveIdx) => (
        <Box key={wave.wave} flexDirection="column">
          <Text bold color={LCARS_COLORS.BLUE}>
            Wave {wave.wave}:
          </Text>
          {wave.directives.map((directive, dirIdx) => {
            const isLast = dirIdx === wave.directives.length - 1;
            const prefix = isLast ? '└─ ' : '├─ ';
            const depPrefix = isLast ? '    ' : '│   ';

            return (
              <Box key={directive.id} flexDirection="column">
                <Text>
                  {prefix}
                  <Text color={LCARS_COLORS.ORANGE}>{directive.id}</Text>
                  <Text dimColor> ({directive.type})</Text>
                </Text>
                {directive.dependencies.map((dep, depIdx) => (
                  <Text key={dep} dimColor>
                    {depPrefix}
                    {depIdx === directive.dependencies.length - 1 ? '└─ dep: ' : '├─ dep: '}
                    {dep}
                  </Text>
                ))}
              </Box>
            );
          })}
        </Box>
      ))}
    </Box>
  );
}
```

### C8: UC Coverage Matrix

**Purpose**: Map use cases to directives for validation

**Design**:
```
UC-INTAKE-01 → d1, d2 ✓
UC-ROLES-02 → d1 ✓
UC-ANALYZE-01 → d1, d2 ✓
...
```

**Elements**:
- Use Case ID (left)
- Arrow (`→`) with mapped directive IDs
- Checkmark (`✓`) for covered, `✗` for gaps

**Color Coding**:
- Covered (✓): Green
- Gaps (✗): Red Alert
- Partial coverage (⚠): Yellow Caution

**Implementation**:
```tsx
interface CoverageMatrixProps {
  coverage: Array<{
    ucId: string;
    directives: string[];
    status: 'covered' | 'partial' | 'gap';
  }>;
}

function CoverageMatrix({ coverage }: CoverageMatrixProps): JSX.Element {
  return (
    <Box flexDirection="column">
      {coverage.map(({ ucId, directives, status }) => {
        const icon = status === 'covered' ? '✓'
                  : status === 'gap' ? '✗'
                  : '⚠';
        const color = status === 'covered' ? LCARS_COLORS.GREEN_OK
                   : status === 'gap' ? LCARS_COLORS.RED_ALERT
                   : LCARS_COLORS.YELLOW_CAUTION;

        return (
          <Text key={ucId}>
            <Text dimColor>{ucId}</Text>
            {' → '}
            <Text>{directives.join(', ')}</Text>
            {' '}
            <Text color={color}>{icon}</Text>
          </Text>
        );
      })}
    </Box>
  );
}
```

---

## Interaction Patterns

### Keyboard Shortcuts (Global)

| Key | Action | Context |
|-----|--------|---------|
| `Tab` | Cycle panel focus | All screens |
| `Shift+Tab` | Reverse panel focus | All screens |
| `Enter` | Drill down / Select | All contexts |
| `Escape` | Go back / Cancel | Detail views |
| `q` | Quit TUI | All screens |
| `?` | Show help overlay | All screens |
| `Ctrl+C` | Force quit | All screens |

### Keyboard Shortcuts (Main Bridge)

| Key | Action | Panel |
|-----|--------|-------|
| `Space` | Pause/resume propulsion | Mission Control |
| `a` | Focus Agents panel | Any |
| `m` | Focus Mission Board | Any |
| `t` | Focus Tactical Display | Any |
| `l` | Focus Engineering Logs | Any |
| `c` | Focus Command Interface | Any |
| `Up/Down` | Navigate list items | Mission Board, Agents |
| `Page Up/Down` | Scroll logs | Engineering Logs |

### Keyboard Shortcuts (Planning)

| Key | Action | Context |
|-----|--------|---------|
| `a` | Approve plan | Plan Review |
| `r` | Request revision | Plan Review |
| `d` | Dry run (export JSON) | Plan Review |
| `e` | Edit PRD | PRD Parse |
| `s` | Skip specialist | Analysis (if hung) |
| `Tab` | Cycle specialist panels | Analysis |

### Navigation Patterns

#### Pattern 1: Panel Focus Cycle

```
┌─────┐   ┌─────┐   ┌─────┐   ┌─────┐   ┌─────┐
│  A  │ → │  M  │ → │  T  │ → │  L  │ → │  C  │
└─────┘   └─────┘   └─────┘   └─────┘   └─────┘
  Agent    Mission   Tactical  Logs    Command
  Control  Board     Display            Interface
    ↑                                                   │
    └───────────────────────────────────────────────────┘
                    Shift+Tab (reverse)
```

**Visual Feedback**: Focused panel shows `*` in header or `::` borders

#### Pattern 2: Drill Down / Back

```
Main Bridge
    │ [Enter] on agent card
    ▼
Agent Detail View
    │ [Enter] on directive link
    ▼
Directive Detail View
    │ [Escape] or [q]
    ▼
Agent Detail View
    │ [Escape] or [q]
    ▼
Main Bridge
```

**Stack-Based Navigation**: Maintain navigation stack for `[Escape]` to pop

#### Pattern 3: Command Input Flow

```
Command Interface (focused)
    │ Type "halt d1"
    ▼
[Enter] → Execute command
    │
    ├─ Success → Feedback: "✓ Halted d1" (green)
    └─ Error → Feedback: "✗ Unknown directive: d1" (red)
    │
    ▼
Return to command input (clear line)
```

**Auto-Repeat**: Press `[Up]` to show last command, `[Enter]` to re-execute

### State Transitions

#### Panel Focus State

```typescript
type PanelFocus = {
  focusedPanel: 'agents' | 'mission' | 'tactical' | 'logs' | 'command';
  navigationStack: Array<string>; // For drill-down navigation
};

function handleTab(current: PanelFocus): PanelFocus {
  const panels = ['agents', 'mission', 'tactical', 'logs', 'command'];
  const idx = panels.indexOf(current.focusedPanel);
  const nextIdx = (idx + 1) % panels.length;
  return { ...current, focusedPanel: panels[nextIdx] };
}
```

#### Selection State

```typescript
type SelectionState = {
  panel: string;
  selectedIndex: number;
  selectedItem: string | null;
};

function handleEnter(state: SelectionState): NavigationAction {
  if (state.selectedItem) {
    return { type: 'DRILL_DOWN', item: state.selectedItem };
  }
  return { type: 'NONE' };
}
```

#### Command Input State

```typescript
type CommandInputState = {
  buffer: string;
  cursor: number;
  history: Array<string>;
  historyIndex: number;
  feedback: string | null;
  feedbackType: 'success' | 'error' | 'info';
};

function handleSubmit(state: CommandInputState): CommandInputState {
  return {
    ...state,
    history: [...state.history, state.buffer],
    historyIndex: state.history.length,
    buffer: '',
    cursor: 0,
  };
}
```

---

## Implementation Guide

### Ink Component Architecture

```typescript
// src/tui/components/StatusIndicator.tsx
export function StatusIndicator({ health, status }: StatusIndicatorProps): JSX.Element;

// src/tui/components/AgentCard.tsx
export function AgentCard({ agentId, directiveId, role, phase, status, elapsed }: AgentCardProps): JSX.Element;

// src/tui/components/MissionBoard.tsx
export function MissionBoard({ board, waveProgress }: MissionBoardProps): JSX.Element;

// src/tui/components/PhaseTracker.tsx
export function PhaseTracker({ events }: PhaseTrackerProps): JSX.Element;

// src/tui/components/CommandInterface.tsx
export function CommandInterface({ command, feedback, onSubmit, onChange }: CommandInterfaceProps): JSX.Element;

// src/tui/components/SpecialistPanel.tsx
export function SpecialistPanel({ specialist }: SpecialistPanelProps): JSX.Element;

// src/tui/components/DependencyGraph.tsx
export function DependencyGraph({ waves }: DependencyGraphProps): JSX.Element;

// src/tui/components/CoverageMatrix.tsx
export function CoverageMatrix({ coverage }: CoverageMatrixProps): JSX.Element;
```

### Layout Composition

```typescript
// src/tui/layouts/MainBridge.tsx
export function MainBridge({ eventBus, control, projectId }: MainBridgeProps): JSX.Element {
  const [focusedPanel, setFocusedPanel] = useState<Panel>('mission');
  const [selectedAgent, setSelectedAgent] = useState<string | null>(null);

  // Subscribe to events, build state
  const agents = useAgents(eventBus);
  const board = useMissionBoard(eventBus);
  const phases = usePhaseTracker(eventBus);
  const waves = useWaveSummary(eventBus);
  const health = useHealth(eventBus, projectId);

  return (
    <Box flexDirection="column">
      <Header project={projectId} health={health} />

      <Box flexDirection="row">
        <MissionControl
          agents={agents}
          focused={focusedPanel === 'mission'}
          onFocus={() => setFocusedPanel('mission')}
          onSelectAgent={setSelectedAgent}
        />
        <MissionBoard
          board={board}
          waves={waves}
          focused={focusedPanel === 'board'}
          onFocus={() => setFocusedPanel('board')}
        />
      </Box>

      <Box flexDirection="row">
        <TacticalDisplay
          phases={phases}
          waves={waves}
          focused={focusedPanel === 'tactical'}
          onFocus={() => setFocusedPanel('tactical')}
        />
      </Box>

      <EngineeringLogs
        events={useEvents(eventBus)}
        focused={focusedPanel === 'logs'}
        onFocus={() => setFocusedPanel('logs')}
      />

      <CommandInterface
        focused={focusedPanel === 'command'}
        onFocus={() => setFocusedPanel('command')}
        onSubmit={handleCommand}
      />
    </Box>
  );
}
```

### Event Bus Integration

```typescript
// src/tui/hooks/useAgents.ts
export function useAgents(eventBus: EventBus): AgentRow[] {
  const [agents, setAgents] = useState<AgentRow[]>([]);

  useEffect(() => {
    const unsub = eventBus.subscribe(() => {
      const events = eventBus.snapshot(150);
      setAgents(buildAgentRows(events));
    });
    return unsub;
  }, [eventBus]);

  return agents;
}

// src/tui/hooks/useMissionBoard.ts
export function useMissionBoard(eventBus: EventBus): MissionBoardState {
  const [board, setBoard] = useState<MissionBoardState>(initialBoard);

  useEffect(() => {
    const unsub = eventBus.subscribe(() => {
      const events = eventBus.snapshot(150);
      setBoard(buildDirectiveBoard(events));
    });
    return unsub;
  }, [eventBus]);

  return board;
}
```

### Planning TUI Architecture

```typescript
// src/tui/layouts/PlanningDashboard.tsx
export function PlanningDashboard({ prdPath, eventBus }: PlanningDashboardProps): JSX.Element {
  const [phase, setPhase] = useState<'parse' | 'analyze' | 'review'>('parse');
  const [prd, setPrd] = useState<PRD | null>(null);
  const [specialists, setSpecialists] = useState<SpecialistOutputs>({});
  const [plan, setPlan] = useState<DirectivePlan | null>(null);

  return (
    <Box flexDirection="column">
      <Header
        title="Mission Briefing"
        prd={prdPath}
        phase={phase}
      />

      {phase === 'parse' && (
        <PRDParseView
          prdPath={prdPath}
          onParsed={(prd) => {
            setPrd(prd);
            setPhase('analyze');
          }}
        />
      )}

      {phase === 'analyze' && (
        <AnalysisView
          prd={prd}
          specialists={specialists}
          onComplete={(outputs) => {
            setSpecialists(outputs);
            setPlan(outputs.synthesizer.plan);
            setPhase('review');
          }}
        />
      )}

      {phase === 'review' && (
        <PlanReviewView
          plan={plan}
          onApprove={handleApprove}
          onRequestRevision={handleRevision}
          onDryRun={handleDryRun}
        />
      )}
    </Box>
  );
}
```

### Color Constants Module

```typescript
// src/tui/theme/colors.ts
export const LCARS_COLORS = {
  ORANGE: '#FF9966',
  BLUE: '#9999CC',
  PURPLE: '#CC99CC',
  PINK: '#FF99CC',
  RED_ALERT: '#FF3333',
  YELLOW_CAUTION: '#FFCC00',
  GREEN_OK: '#33FF33',
  BLACK: '#000000',
  DARK_BLUE: '#1B4F8F',
  GRAY: '#333333',
  WHITE: '#FFFFFF',
  LIGHT_GRAY: '#CCCCCC',
} as const;

export const STATUS_COLORS = {
  ok: LCARS_COLORS.GREEN_OK,
  warning: LCARS_COLORS.YELLOW_CAUTION,
  error: LCARS_COLORS.RED_ALERT,
  critical: LCARS_COLORS.RED_ALERT,
  running: LCARS_COLORS.ORANGE,
  done: LCARS_COLORS.GREEN_OK,
  stuck: LCARS_COLORS.RED_ALERT,
  waiting: LCARS_COLORS.BLUE,
  skipped: LCARS_COLORS.GRAY,
} as const;

export const RESULT_COLORS = {
  pass: LCARS_COLORS.GREEN_OK,
  fail: LCARS_COLORS.RED_ALERT,
  running: LCARS_COLORS.ORANGE,
  wait: LCARS_COLORS.BLUE,
  done: LCARS_COLORS.GREEN_OK,
} as const;
```

### Keyboard Navigation Handler

```typescript
// src/tui/hooks/useKeyboardNav.ts
export function useKeyboardNav(
  panels: string[],
  initialFocus: string
): {
  focusedPanel: string;
  focusPanel: (panel: string) => void;
  nextPanel: () => void;
  prevPanel: () => void;
} {
  const [focusedPanel, setFocusedPanel] = useState(initialFocus);
  const [navigationStack, setNavigationStack] = useState<string[]>([]);

  useInput((input, key) => {
    if (key.tab) {
      const idx = panels.indexOf(focusedPanel);
      const nextIdx = key.shift
        ? (idx - 1 + panels.length) % panels.length
        : (idx + 1) % panels.length;
      setFocusedPanel(panels[nextIdx]);
    }
  });

  const focusPanel = useCallback((panel: string) => {
    setNavigationStack([...navigationStack, focusedPanel]);
    setFocusedPanel(panel);
  }, [focusedPanel, navigationStack]);

  const goBack = useCallback(() => {
    if (navigationStack.length > 0) {
      const prev = navigationStack[navigationStack.length - 1];
      setNavigationStack(navigationStack.slice(0, -1));
      setFocusedPanel(prev);
    }
  }, [navigationStack]);

  return {
    focusedPanel,
    focusPanel,
    nextPanel: () => {
      const idx = panels.indexOf(focusedPanel);
      setFocusedPanel(panels[(idx + 1) % panels.length]);
    },
    prevPanel: () => {
      const idx = panels.indexOf(focusedPanel);
      setFocusedPanel(panels[(idx - 1 + panels.length) % panels.length]);
    },
  };
}
```

---

## Accessibility Considerations

### Color Blindness Support

The LCARS color palette relies heavily on color differentiation. To support color-blind operators:

1. **Add Text Labels**: Always pair color with text labels
   ```
   [●●●●○] Health: 80% (OK)
   [d1] RUNNING (Orange)
   ```

2. **Use Symbols**: Include status symbols alongside colors
   - `✓` (pass), `✗` (fail), `⚠` (warning), `⊘` (skipped)

3. **High Contrast Mode**: Option to switch to grayscale with symbols only
   ```typescript
   interface AccessibilitySettings {
     highContrast: boolean;
     colorBlindMode: 'none' | 'protanopia' | 'deuteranopia' | 'tritanopia';
     symbolEnhanced: boolean;
   }
   ```

### Screen Reader Compatibility

While TUIs are inherently visual, we can support terminal screen readers (e.g., `brltty`):

1. **Semantic HTML**: Use Ink's accessible output modes
2. **Alt Text**: Add descriptive text for visual elements
   ```typescript
   <Text>
     <Text color={color}>●●●●○</Text>
     {' '}
     <Text dimColor>(Health: 80%, 4 of 5 systems OK)</Text>
   </Text>
   ```

3. **Audio Cues**: Optional beep/bell for alerts
   ```bash
   # Terminal bell on critical errors
   echo -e "\a"  # ASCII BEL character
   ```

### Keyboard-Only Navigation

All interactions must be keyboard-driven (no mouse required):

- `[Tab]` / `[Shift+Tab]`: Panel focus
- `[Up]` / `[Down]`: List navigation
- `[Enter]`: Select / Drill down
- `[Escape]`: Go back / Cancel
- `[Space]`: Pause/resume / Toggle
- `q`: Quit (always available)

### Font Size & Terminal Constraints

Support varying terminal sizes:

```typescript
interface TerminalSize {
  columns: number;
  rows: number;
}

function getLayoutConfig(size: TerminalSize): LayoutConfig {
  if (size.columns < 80) {
    return {
      agentCards: 4,      // Reduce from 8
      logLines: 4,        // Reduce from 8
      panelLayout: 'stacked',  // Vertical instead of side-by-side
    };
  } else if (size.columns > 120) {
    return {
      agentCards: 12,
      logLines: 12,
      panelLayout: 'expanded',
    };
  }
  return {
    agentCards: 8,
    logLines: 8,
    panelLayout: 'standard',
  };
}
```

---

## Next Steps

### For Product Manager Review

1. **Validate Visual Direction**: Is "moderate Star Trek" the right balance?
2. **Confirm Scope**: Design both planning TUI and runtime TUI refresh?
3. **Prioritize Components**: Which components are MVP vs. nice-to-have?
4. **Validate User Flows**: Do the flows match operational needs?

### For Implementation

1. **Phase 1**: Core components (StatusIndicator, AgentCard, MissionBoard)
2. **Phase 2**: Main Bridge layout with keyboard navigation
3. **Phase 3**: Planning TUI (PRD parse, specialist dashboard, plan review)
4. **Phase 4**: Polish (animations, accessibility, help overlay)

### Design Artifacts to Create

1. **Component Storybook**: Ink component showcase (`npm run tui:storybook`)
2. **Interactive Prototype**: Working TUI mock with fake data
3. **Operator Handbook**: User guide for keyboard shortcuts and flows
4. **Video Demo**: Screen recording of TUI in action

---

## Appendix: Star Trek Terminology Mapping

| Ship Commander Term | Star Trek Equivalent | Usage Context |
|---------------------|----------------------|---------------|
| **Operator** | Captain / Admiral | Human controlling the system |
| **Agent** | Ensign / Crewmember | AI worker executing tasks |
| **Directive** | Mission Order | Task to be completed |
| **Propulsion** | Warp Drive | Engine pulling work |
| **Doctor** | Chief Medical Officer | Health monitoring system |
| **Mission Board** | Tactical Display | Kanban-style status view |
| **Engineering Logs** | Captain's Log | Event history |
| **Command Interface** | Helm Control | Operator input |
| **Wave** | Formation Group | Execution batch |
| **Health** | Ship Status | System health indicator |
| **Halted** | Red Alert | Critical failure state |
| **Review** | Inspection Phase | Pre-approval gate |
| **Approved** | Mission Complete | Ready for merge |
| **Stuck** | Disabled / Damaged | Agent not progressing |
| **Specialist** | Department Head | Captain, Commander, Design Officer |
| **Intake** | Mission Briefing | PRD ingestion phase |
| **Synthesizer** | First Officer | Integration role |

**Sources:**
- [LCARS Color Guide](https://www.thelcars.com/colors.php)
- [The user interfaces of Star Trek – LCARS](https://craftofcoding.wordpress.com/2015/10/13/the-user-interfaces-of-star-trek-lcars/)
- [Ink GitHub Repository](https://github.com/vadimdemedes/ink)
- [Build a System Monitor TUI in Go](https://penchev.com/posts/create-tui-with-go/)
- [Ink.js Terminal UI Design - Claude Code Skill](https://mcpmarket.com/es/tools/skills/ink-js-terminal-ui-design)
