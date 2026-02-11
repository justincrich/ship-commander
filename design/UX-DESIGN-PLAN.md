# UX Design Plan: Ship Commander 3

**Version**: 2.1
**Date**: 2026-02-10
**Platform**: Terminal/CLI (Bubble Tea + Lipgloss)
**Aesthetic**: Charm-LCARS Fusion

---

## Design Artifact Index

This document is the **broad overview** of Ship Commander 3's UX design. Detailed specifications live in dedicated YAML artifacts. Use this document for orientation; drill into the YAML files for implementation detail.

| Artifact | Path | Contents |
|----------|------|----------|
| **Views** | `design/views.yaml` | Per-view layout, panels, data sources, keyboard shortcuts, responsive behavior |
| **Screens** | `design/screens.yaml` | Navigation hierarchy, screen routing, transitions, animation presets |
| **Components** | `design/components.yaml` | Component library (35 components), Charm library mappings, composition patterns |
| **Flows** | `design/flows.yaml` | Micro-level interaction flows (16 flows), step-by-step user interactions |
| **Workflows** | `design/workflows.yaml` | End-to-end user workflows (8 workflows), multi-screen journeys |
| **Paradigm** | `design/paradigm.yaml` | Design patterns, principles, animation categories, terminology glossary |
| **Config** | `design/config.yaml` | Color palette, technology stack, keyboard plan, display modes |
| **Mockups** | `design/mocks/*.mock.html` | HTML visual mockups for all 17 screens/overlays (see Screen Inventory for per-screen mapping) |

---

## Table of Contents

1. [Design Vision and Principles](#1-design-vision-and-principles)
2. [Information Architecture](#2-information-architecture)
3. [Navigation Model](#3-navigation-model)
4. [Responsive Behavior](#4-responsive-behavior)
5. [Accessibility](#5-accessibility)
6. [Component Strategy](#6-component-strategy)
7. [Animation Strategy](#7-animation-strategy)
8. [Progressive Disclosure](#8-progressive-disclosure)
9. [Verification Gates and Settings](#9-verification-gates-and-settings)

---

## 1. Design Vision and Principles

### Vision Statement

Ship Commander 3 merges Charmbracelet's playful, glamorous terminal aesthetic with Star Trek LCARS thematic elements. The result is a polished Charm CLI tool themed as a starship command console -- glamorous agentic coding meets Starfleet. The TUI must make an operator feel like they are commanding AI coding agents from a bridge while providing the transparency and debuggability developers demand.

### Core Design Principles

**P1: Flows Before Screens**
Every screen emerges from a user workflow, not imagination. The Ready Room exists because planning is a collaborative act. The Ship Bridge exists because execution demands real-time monitoring. No screen exists without a workflow that demands it.

**P2: Deterministic Feedback, Always**
The manifesto separates probabilistic exploration from deterministic decision-making. The TUI must reflect this: every gate result, every state transition, and every status indicator is drawn from deterministic data. No ambiguous states. No "maybe it worked" indicators.

**P3: Symbols + Color, Never Color Alone**
Every semantic status is communicated through a symbol AND a color. A color-blind operator can operate Ship Commander with full confidence. The LCARS palette is beautiful, but it is always supplemented with textual or symbolic differentiation.

**P4: Keyboard-First, Always**
Every interaction is reachable via keyboard. No mouse is required. Power users navigate at keyboard speed. Novice users discover shortcuts through help overlays. The command bar and Huh forms are the primary interaction surfaces.

**P5: Progressive Disclosure Over Information Overload**
Basic mode shows what matters. Advanced mode shows everything. Executive mode shows only aggregates. The operator never sees more than their current task demands.

**P6: Smart Main, Dumb Components**
Following the Charmbracelet Crush architecture pattern: a single "smart" root `tea.Model` owns all application state and message routing. Child components are "dumb" -- they receive data as arguments, render views, and emit messages upward. No child component manages its own subscriptions or state persistence.

**P7: Charm Aesthetic Meets LCARS Soul**
The Charm side provides: clean Lipgloss styling, immutable style composition, rounded borders where appropriate, responsive compact mode, and a "fun but productive" energy. The LCARS side provides: the butterscotch/blue/purple/pink semantic color palette, box-drawing panel structure, Star Trek terminology, and the feeling of a starship bridge. Neither aesthetic dominates; they fuse.

> **Detail**: Full principle definitions in `design/paradigm.yaml` (principles section). Library references in `design/paradigm.yaml` (references section).

### Emotional Design Goals

| Persona | Primary Emotion | Secondary Emotion |
|---------|----------------|-------------------|
| Staff Commander | Confidence and control | Delight at the aesthetic |
| Junior Lieutenant | Safety and clarity | Growing competence |
| Fleet Admiral | Efficiency and oversight | Trust in the system |

---

## 2. Information Architecture

### Top-Level Structure

Ship Commander 3 uses a **fleet-centric navigation hierarchy**. There are no separate planning/execution modes -- all views exist in a unified drill-down tree. Every view ends with a **NavigableToolbar** showing available navigation and actions.

```
sc3 tui
  |
  +-- [Fleet Overview] Landing screen (one card per ship/directive)
  |     |-- Ship card grid (name, directive, status, progress)
  |     |-- Ship preview panel (selected ship details)
  |     +-- NavigableToolbar: [Enter] Bridge  [n] New  [f] Monitor  [a] Roster  [i] Inbox  [s] Settings
  |
  +-- [Ship Bridge] Per-ship execution monitoring (drill-down from Fleet)
  |     |-- Crew panel (agents with role, mission, phase, time)
  |     |-- Mission board (Kanban: Backlog / In Progress / Review / Done / Halted)
  |     |-- Event log (scrollable, severity-filtered)
  |     |-- Header with inline wave summary ("Wave K of L [compact bar]")
  |     +-- NavigableToolbar: [r] Ready Room  [w] Waves  [h] Halt  [Space] Pause  [Esc] Fleet
  |
  +-- [Ready Room] Per-ship planning (drill-down from Ship Bridge)
  |     |-- Specialist grid (Captain / Commander / Design Officer status)
  |     |-- Directive viewport (PRD rendered via Glamour)
  |     |-- Specialist Detail (drill-down from specialist grid, Enter on specialist)
  |     +-- NavigableToolbar: [v] Review Plan  [t] Toggle Directive  [Esc] Bridge
  |
  +-- [Plan Review] Full-screen manifest review (drill-down from Ready Room)
  |     |-- Mission manifest viewport (Glamour markdown)
  |     |-- Coverage matrix + dependency graph (analysis panel)
  |     +-- NavigableToolbar: [a] Approve  [f] Feedback  [s] Shelve  [Esc] Ready Room
  |
  +-- [Fleet Monitor] Condensed multi-ship status (drill-down from Fleet)
  |     |-- ShipStatusRow grid (one line per ship with inline progress)
  |     +-- NavigableToolbar: [Enter] Bridge  [i] Inbox  [Esc] Fleet
  |
  +-- [Drill-down views] (pushed onto navigation stack)
  |     |-- Mission Detail (per-AC TDD phase pipeline, gate evidence, output)
  |     |-- Agent Detail (output stream, assignment, phase, health)
  |     |-- Specialist Detail (per-specialist output, status, assignment)
  |     +-- Wave Manager (dependency graph, merge controls)
  |
  +-- [Project Settings] Global configuration (top-level from Fleet)
  |     |-- Verification Gates (bash commands for test/lint/typecheck/build)
  |     |-- Crew Defaults (harness, model, WIP limits, timeouts)
  |     |-- Fleet Defaults (naming, wave strategy, merge policy)
  |     |-- Export/Import (JSON settings file for cross-project portability)
  |     +-- NavigableToolbar: [1] Gates  [2] Crew  [3] Fleet  [4] Export  [Esc] Fleet
  |
  +-- [Global overlays] (modal layer, atop any screen)
        |-- Admiral Question Modal (Huh Select + Input + Confirm)
        |-- Help Overlay (contextual keyboard reference)
        |-- Onboarding Overlay (first-run welcome tour, Basic mode)
        +-- Confirmation Dialog (Huh Confirm for destructive actions)
```

> **Detail**: Full view specifications (layout, panels, data sources, keyboard shortcuts) in `design/views.yaml`. Screen routing, navigation hierarchy, and transition animations in `design/screens.yaml`.

### Screen Inventory

| Screen ID | Name | Type | Purpose | Entry Points | Mockup |
|-----------|------|------|---------|--------------|--------|
| `fleet_overview` | Fleet Overview | primary | Landing screen, ship card grid with preview | `sc3 tui`, Esc from Ship Bridge | [mock](design/mocks/fleet-overview.mock.html) |
| `ship_bridge` | Ship Bridge | drill-down | Per-ship execution monitoring | Enter on ship in Fleet Overview | [mock](design/mocks/ship-bridge.mock.html) |
| `ready_room` | Ready Room | drill-down | Per-ship planning (specialists + directive) | `[r]` from Ship Bridge | [mock](design/mocks/ready-room.mock.html) |
| `plan_review` | Plan Review | drill-down | Manifest review with coverage/dependency analysis | `[v]` from Ready Room, auto on consensus | [mock](design/mocks/plan-review.mock.html) |
| `fleet_monitor` | Fleet Monitor | drill-down | Condensed multi-ship status rows | `[f]` from Fleet Overview | [mock](design/mocks/fleet-monitor.mock.html) |
| `mission_detail` | Mission Detail | drill-down | Per-mission AC phases, gates, demo token | Enter on mission in Mission Board | [mock](design/mocks/mission-detail.mock.html) |
| `agent_detail` | Agent Detail | drill-down | Per-agent output, phase, health | Enter on agent in Crew Panel | [mock](design/mocks/agent-detail.mock.html) |
| `specialist_detail` | Specialist Detail | drill-down | Per-specialist output, status, assignment in Ready Room | Enter on specialist in Ready Room grid | [mock](design/mocks/specialist-detail.mock.html) |
| `agent_roster` | Agent Roster | top-level | Cross-ship agent inventory | `[a]` from Fleet Overview | [mock](design/mocks/agent-roster.mock.html) |
| `directive_editor` | Directive Editor | top-level | PRD input form with ship assignment | `[n]` from Fleet Overview | [mock](design/mocks/directive-editor.mock.html) |
| `message_center` | Message Center | top-level | Cross-ship Admiral question inbox | `[i]` from any view | [mock](design/mocks/message-center.mock.html) |
| `wave_manager` | Wave Manager | overlay | Wave dependency graph and merge controls | `[w]` from Ship Bridge | [mock](design/mocks/wave-manager.mock.html) |
| `project_settings` | Project Settings | top-level | Global config: verification gates, crew/fleet defaults, export/import | `[s]` from Fleet Overview | [mock](design/mocks/project-settings.mock.html) |
| `admiral_question` | Admiral Question | modal | Agent-to-Admiral structured input | Auto when agent surfaces question | [mock](design/mocks/admiral-question-modal.mock.html) |
| `help_overlay` | Help | overlay | Contextual keyboard shortcuts | `?` key from any screen | [mock](design/mocks/help-overlay.mock.html) |
| `onboarding_overlay` | Onboarding | overlay | First-run welcome tour (key concepts + navigation) | First launch in Basic mode | [mock](design/mocks/onboarding-overlay.mock.html) |
| `confirm_dialog` | Confirmation | modal | Destructive action confirmation | Halt, force-kill, shelve actions | [mock](design/mocks/confirm-dialog.mock.html) |

### Information Hierarchy per Screen

#### Fleet Overview

```
Priority 1 (always visible):
  - Ship cards (name, directive title, status badge, progress bar)
  - Ship count and active/complete summary in header
  - NavigableToolbar with navigation shortcuts

Priority 2 (visible when ship selected):
  - Ship preview panel (crew roster, mission summary, wave progress)
```

#### Ship Bridge

```
Priority 1 (always visible):
  - Header with ship name, directive, health, inline wave summary
  - Crew panel (agents with role, mission, phase, elapsed time)
  - Mission board summary (column counts: B:2 IP:3 R:1 D:5 H:0)
  - NavigableToolbar with context-sensitive actions

Priority 2 (visible in standard+ width):
  - Event log (last 8-12 lines, severity coded)

Priority 3 (available via drill-down):
  - Per-AC TDD phase detail (drill into Mission Detail)
  - Gate evidence details (drill into Mission Detail)
  - Wave dependency graph (open Wave Manager)
  - Agent output streams (drill into Agent Detail)
```

#### Ready Room

```
Priority 1 (always visible):
  - Specialist status indicators (Captain [*] Commander [*] Design Officer [*])
  - Planning iteration and pending question count in header
  - NavigableToolbar with [v] Review Plan action

Priority 2 (toggle with [t]):
  - Directive viewport (PRD content rendered via Glamour)

Priority 3 (available via drill-down to Plan Review):
  - Coverage matrix (use-case to mission mapping)
  - Dependency graph
```

### Data Flow: Event Bus to TUI

```
Beads State Layer
      |
      v
  Event Bus (Go channels, typed pub/sub)
      |
      +---> TUI Root Model (smart main)
      |       |
      |       +---> Active View (from navStack top)
      |       |       +---> Fleet Overview / Ship Bridge / Ready Room / etc.
      |       +---> Overlay stack (modals, question, help)
      |       +---> Notification queue
      |
      +---> charmbracelet/log (structured logging)
```

The root `tea.Model` subscribes to the event bus and translates events into Bubble Tea `tea.Msg` types. Child views are dumb render functions that receive data as arguments and return styled strings. State flows downward; messages flow upward. See `design/paradigm.yaml` for the smart-main/dumb-components pattern.

---

## 3. Navigation Model

### Global Navigation Paradigm

Ship Commander 3 uses a **fleet-centric drill-down hierarchy** with a **NavigableToolbar** on every view and a **modal/overlay layer** for focused input.

There are three navigation layers:

1. **View layer**: Fleet Overview → Ship Bridge → Ready Room → Plan Review (stack-based, Enter/Esc)
2. **Panel layer**: Tab/Shift+Tab cycles focus between panels within a view
3. **Overlay layer**: Modals pushed onto a stack, dismissed with Esc or completion

### NavigableToolbar Pattern

Every view has a bottom toolbar showing available actions as labeled buttons. This provides consistent wayfinding across all screens (like htop or Midnight Commander):

```
[r] Ready Room  [w] Waves  [h] Halt  [Space] Pause  [?] Help  [Esc] Fleet
     ^^^                                                ^^^^^^^^
     dimmed                                             highlighted (arrow navigation)
```

- **Quick key**: Press the shortcut key directly (e.g., `r` for Ready Room)
- **Arrow nav**: Left/Right arrows highlight buttons, Enter activates
- **Context-sensitive**: Each view shows only relevant actions

### Panel Focus Cycle

Within each view, panels are ordered for logical workflow. Tab advances forward; Shift+Tab reverses. The NavigableToolbar is always the last tab stop.

#### Fleet Overview Panel Order

```
[1] Ship List  -->  [2] Ship Preview  -->  [3] NavigableToolbar
       ^                                           |
       +-------------------------------------------+
```

#### Ship Bridge Panel Order

```
[1] Crew Panel  -->  [2] Mission Board  -->  [3] Event Log  -->  [4] NavigableToolbar
       ^                                                                |
       +----------------------------------------------------------------+
```

#### Ready Room Panel Order

```
[1] Specialist Grid  -->  [2] Directive Viewport  -->  [3] NavigableToolbar
       ^                                                       |
       +-------------------------------------------------------+
```

### Focus Visual Indicator

The focused panel receives a distinct border treatment:

- **Unfocused panel**: Single-line border in `galaxy_gray` (#52526A)
- **Focused panel**: Double-line border in `moonlit_violet` (#9966FF) with bold panel title
- **Active overlay**: Rounded border in `butterscotch` (#FF9966) with drop shadow effect (1-char offset in `dark_blue`)

### Drill-Down / Back Stack

Navigation into detail views follows a stack model. Pressing Enter on an interactive element pushes a detail view. Pressing Esc pops back.

```
Fleet Overview
  |-- [Enter] on ship "USS Auth"
  |     +-- Ship Bridge (pushed)
  |           |-- [Enter] on Agent "impl-abc" in Crew Panel
  |           |     +-- Agent Detail View (pushed)
  |           |           |-- [Esc] pops back to Ship Bridge
  |           |-- [Enter] on Mission "MISSION-42" in Mission Board
  |           |     +-- Mission Detail View (pushed)
  |           |           |-- [Esc] pops back to Ship Bridge
  |           |-- [Esc] pops back to Fleet Overview
```

The stack is bounded to depth 3. Attempting to push beyond 3 replaces the deepest entry.

### Global Keyboard Shortcuts

These work from any view, any panel, any overlay depth.

| Key | Action | Context |
|-----|--------|---------|
| `?` | Toggle help overlay | Always |
| `Ctrl+C` | Quit (with confirmation if missions active) | Always |
| `q` | Quit (same as Ctrl+C) | When not in text input |
| `Tab` | Focus next panel | Within current view |
| `Shift+Tab` | Focus previous panel | Within current view |
| `Enter` | Select / drill down / activate toolbar button | On interactive element |
| `Esc` | Back / pop navigation stack / dismiss overlay | Views, overlays |
| `Left/Right` | Highlight toolbar buttons | When NavigableToolbar focused |

> **Detail**: Complete navigation architecture, transitions, and animation presets in `design/screens.yaml`. Per-view keyboard shortcuts in `design/views.yaml`. Interaction flows in `design/flows.yaml`.

### View-Specific Shortcuts (via NavigableToolbar)

Each view shows context-sensitive shortcuts in its NavigableToolbar. These are the primary navigation shortcuts:

#### Fleet Overview

| Key | Action | Shown In |
|-----|--------|----------|
| `Enter` | Open Ship Bridge for selected ship | NavigableToolbar |
| `n` | New directive | NavigableToolbar |
| `f` | Fleet Monitor | NavigableToolbar |
| `a` | Agent Roster | NavigableToolbar |
| `i` | Message Center (inbox) | NavigableToolbar |
| `s` | Project Settings | NavigableToolbar |

#### Ship Bridge

| Key | Action | Shown In |
|-----|--------|----------|
| `r` | Ready Room (planning) | NavigableToolbar |
| `w` | Wave Manager | NavigableToolbar |
| `h` | Halt selected mission | NavigableToolbar |
| `Space` | Pause/resume propulsion | NavigableToolbar |
| `Esc` | Back to Fleet Overview | NavigableToolbar |

#### Ready Room

| Key | Action | Shown In |
|-----|--------|----------|
| `v` | Review Plan (Plan Review view) | NavigableToolbar |
| `t` | Toggle directive sidebar | NavigableToolbar |
| `Esc` | Back to Ship Bridge | NavigableToolbar |

#### Plan Review

| Key | Action | Shown In |
|-----|--------|----------|
| `a` | Approve plan | NavigableToolbar |
| `f` | Provide feedback (reconvene loop) | NavigableToolbar |
| `s` | Shelve plan | NavigableToolbar |
| `Esc` | Back to Ready Room | NavigableToolbar |

---

## 4. Responsive Behavior

### Breakpoint Strategy

Ship Commander 3 defines two layout modes based on terminal width. Height is handled by reducing visible rows within panels, not by collapsing layout.

| Mode | Width | Height | Description |
|------|-------|--------|-------------|
| **Standard** | >= 120 cols | >= 30 rows | Full side-by-side layout, all panels visible |
| **Compact** | < 120 cols | any | Stacked vertical layout, panels simplified |

These breakpoints are detected at startup and on `tea.WindowSizeMsg`. The root model stores `width` and `height` and passes the appropriate layout variant to child models.

### Standard Layout: Fleet Overview (>= 120 cols)

```
+============================================================================+
| FLEET COMMAND | Ships: 3 active  1 complete | Health: [OK] | WIP 3/5      |
+============================================================================+
| Ship List (40%)                | Ship Preview (60%)                        |
| +----------------------------+ | +--------------------------------------+ |
| | > USS Auth       [>] W2/3  | | | USS Auth -- auth-system              | |
| |   USS Payments   [>] W1/2  | | | Crew: capt-abc, impl-def, impl-ghi  | |
| |   USS Docs       [+] Done  | | | Board: B:2 IP:3 R:1 D:5 H:0        | |
| |                            | | | Wave 2 of 3 [====....] 2/4          | |
| |                            | | | Questions: 0 pending                | |
| +----------------------------+ | +--------------------------------------+ |
+============================================================================+
| [Enter] Bridge  [n] New  [f] Monitor  [a] Roster  [i] Inbox  [?] Help    |
+============================================================================+
```

### Standard Layout: Ship Bridge (>= 120 cols)

```
+============================================================================+
| USS Auth | auth-system | [OK] 4/5 crew | Wave 2 of 3 [====....] 2/4      |
+============================================================================+
| Crew Panel (40%)               | Mission Board (60%)                       |
| +----------------------------+ | +--------------------------------------+ |
| | [*] capt-abc  MISSION-01   | | | B:2  IP:3  R:1  D:5  H:0            | |
| |     captain   DONE  01:30  | | |                                      | |
| | [>] impl-def  MISSION-03   | | | MISSION-03 [impl] RED       02:15   | |
| |     implementer RED  02:15 | | | MISSION-04 [impl] GREEN     01:45   | |
| | [>] impl-ghi  MISSION-04   | | | MISSION-05 [rev]  REVIEW    00:30   | |
| |     implementer GRN  01:45 | | |                                      | |
| +----------------------------+ | +--------------------------------------+ |
+================================+=========================================+
| Event Log                                                                 |
| +-----------------------------------------------------------------------+ |
| | 14:30:01 [INFO] commander.dispatch mission=M-03 agent=impl-def        | |
| | 14:30:15 [WARN] doctor.stuck agent=impl-xyz timeout=300s              | |
| | 14:30:22 [INFO] gate.verify_red mission=M-03 ac=AC-1 result=REJECT   | |
| | 14:30:45 [INFO] gate.verify_green mission=M-04 ac=AC-2 result=PASS   | |
| +-----------------------------------------------------------------------+ |
+===========================================================================+
| [r] Ready Room  [w] Waves  [h] Halt  [Space] Pause  [?] Help  [Esc] Fleet|
+===========================================================================+
```

### Compact Layout (< 120 cols)

In compact mode, side-by-side panels stack vertically. Phase detail is always in Mission Detail drill-down.

```
+=============================================+
| USS Auth | [OK] | W2/3 [====....] | 3/5    |
+=============================================+
| Crew (3)                                    |
| impl-def  M-03 RED     02:15 [>]           |
| impl-ghi  M-04 GREEN   01:45 [>]           |
| rev-jkl   M-05 REVIEW  00:30 [~]           |
+---------------------------------------------+
| Board  B:2  IP:3  R:1  D:5  H:0            |
+---------------------------------------------+
| Log                                         |
| 14:30:22 [INFO] gate M-03 AC-1 REJECT      |
| 14:30:45 [INFO] gate M-04 AC-2 PASS        |
+---------------------------------------------+
| [r]Ready [w]Wave [h]Halt [Esc]Fleet [?]Help |
+=============================================+
```

### Compact Mode Adaptations

| Element | Standard | Compact |
|---------|----------|---------|
| Crew Panel + Mission Board | Side-by-side (40/60 split) | Stacked, Crew first, Board summary below |
| Event Log | 8 lines | 4 lines |
| Wave Summary | Inline in header with progress bar | Abbreviated fraction: `W2/3` |
| Header | Full ship name, directive, health, wave bar | Abbreviated: ship, health icon, wave fraction |
| NavigableToolbar | Full labels with shortcut keys | Abbreviated: `[r]Ready [w]Wave [h]Halt` |
| Panel borders | Double-line for focused, single for unfocused | Single-line only (save horizontal space) |
| Phase detail | Always in Mission Detail drill-down | Same (not on Ship Bridge) |

### Height Adaptations

When terminal height < 30 rows:

- Event log reduces to 2 lines
- Crew Panel shows max 3 agents (scrollable)
- Ship Preview panel hidden in Fleet Overview
- NavigableToolbar abbreviates labels

### Dynamic Resize

On `tea.WindowSizeMsg`, the root model:

1. Updates stored `width` and `height`
2. Recalculates layout mode (standard vs compact)
3. Reflows all panels in a single render cycle
4. If transitioning between modes, uses a Harmonica spring for smooth panel reflow (see Section 7)

---

## 5. Accessibility

### Guiding Principle

No information is conveyed by color alone. Every color-coded element has a co-located text label, symbol, or shape that independently communicates the same information.

### Symbol System

All status indicators use a consistent symbol vocabulary alongside color.

| Status | Symbol | Color | Text Label |
|--------|--------|-------|------------|
| Running / Active | `[>]` | butterscotch (#FF9966) | `RUNNING` |
| Success / Done | `[+]` | green_ok (#33FF33) | `PASS` or `DONE` |
| Failed / Halted | `[!]` | red_alert (#FF3333) | `FAIL` or `HALTED` |
| Warning / Stuck | `[~]` | yellow_caution (#FFCC00) | `STUCK` or `WARN` |
| Info / Neutral | `[-]` | blue (#9999CC) | `INFO` |
| Planning / Review | `[?]` | purple (#CC99CC) | `REVIEW` or `PENDING` |
| Waiting / Queued | `[.]` | light_gray (#CCCCCC) | `WAITING` or `QUEUED` |
| Skipped / Disabled | `[x]` | galaxy_gray (#52526A) | `SKIPPED` |

### TDD Phase Indicators

The TDD pipeline for each AC uses a combined symbol + color + label system.

```
RED [!] --> VERIFY_RED [>] --> GREEN [>] --> VERIFY_GREEN [+] --> REFACTOR [>] --> VERIFY_REFACTOR [+]
```

Each phase shows:
- Phase name (text)
- Status symbol (`[!]` fail expected, `[>]` active, `[+]` passed, `[.]` pending)
- Color (semantic from palette)

### Huh Accessible Mode

Huh provides built-in accessible themes. Ship Commander 3 activates accessible mode when:

1. The `SC3_ACCESSIBLE` environment variable is set
2. The `--accessible` CLI flag is passed
3. The `[tui] accessible = true` TOML config is set

In accessible mode:
- Huh forms use the `huh.ThemeBase()` theme with high-contrast overrides
- Focus indicators use reverse video (inverted foreground/background) instead of color-only borders
- All Huh Select options include text labels alongside any color indicators
- Form validation errors are announced with a leading `ERROR:` text prefix, not just red coloring

### Contrast Ratios

The LCARS color palette on a black (#000000) background provides the following approximate contrast ratios:

| Color | Hex | Contrast vs Black | WCAG AA (4.5:1) |
|-------|-----|-------------------|------------------|
| space_white | #F5F6FA | 18.9:1 | PASS |
| butterscotch | #FF9966 | 8.2:1 | PASS |
| green_ok | #33FF33 | 10.5:1 | PASS |
| yellow_caution | #FFCC00 | 12.1:1 | PASS |
| red_alert | #FF3333 | 4.7:1 | PASS |
| blue | #9999CC | 6.4:1 | PASS |
| purple | #CC99CC | 8.5:1 | PASS |
| light_gray | #CCCCCC | 12.6:1 | PASS |
| galaxy_gray | #52526A | 2.8:1 | FAIL (decorative only) |

`galaxy_gray` is used only for decorative borders and disabled elements -- never for text that conveys essential information. When used for disabled text, a `(disabled)` label accompanies it.

### Keyboard Navigation Completeness

Every interactive element is reachable via keyboard:

- Panel cycling: `Tab` / `Shift+Tab`
- List navigation: `Up` / `Down` arrows
- Selection: `Enter`
- Dismissal: `Esc`
- Command entry: `:` or direct typing in command bar
- Help: `?`
- Quit: `q` or `Ctrl+C`

No feature requires a mouse. Mouse support is available as an optional enhancement (clicking panels to focus, clicking list items to select) but is never the only path.

### Terminal Bell

On critical events (mission halted, gate failure requiring Admiral attention), Ship Commander 3 emits a terminal bell character (`\a`) to trigger the user's system notification. This is suppressible via `[tui] bell = false`.

---

## 6. Component Strategy

### Design Philosophy: Charm Libraries First

Ship Commander 3 uses Charmbracelet's component libraries as the primary building blocks. Custom components are created only when no Bubbles/Huh component satisfies the need and the pattern appears in 2+ locations.

> **Detail**: Full component library (35 components) with props, variants, tokens, Charm library mappings, and composition patterns in `design/components.yaml`.

### Component Map

#### Bubbles Components (from github.com/charmbracelet/bubbles)

| Bubbles Component | Ship Commander Usage | Screen(s) |
|-------------------|---------------------|-----------|
| **spinner** | Agent "working" indicator, gate "running" animation | Ready Room, Ship Bridge |
| **progress** | Wave progress bars, WIP utilization bar | Ship Bridge |
| **table** | Crew Panel agent list (columns: role, mission, phase, time, harness) | Ship Bridge |
| **list** | Mission Board items (scrollable, filterable per column), event log entries | Ship Bridge |
| **viewport** | Event Log (scrollable region with PgUp/PgDn), Glamour-rendered markdown content | Ship Bridge, Plan Review, Mission Detail |
| **textinput** | Command bar input (single-line operator commands) | Ship Bridge |
| **paginator** | Multi-page mission list in Plan Review (when > 10 missions) | Plan Review |
| **help** | Contextual keyboard shortcut display (keybinding definitions) | Help Overlay, panel footer hints |
| **key** | Keybinding definitions fed to `help` component | All screens |
| **timer** | Agent elapsed time display, gate timeout countdown | Ship Bridge, Agent Detail |
| **stopwatch** | Planning iteration elapsed time | Ready Room |

#### Huh Components (from github.com/charmbracelet/huh)

All Admiral interactions use Huh forms. Huh forms implement `tea.Model` and compose natively into Bubble Tea.

| Huh Component | Ship Commander Usage | Screen(s) |
|---------------|---------------------|-----------|
| **huh.Select** | Admiral question options, plan approval (approve/feedback/shelve), mission action (halt/retry/approve) | Admiral Question Modal, Plan Review, Confirm Dialog |
| **huh.Input** | Free-text answer to Admiral questions, command bar alternative, feedback text (single-line) | Admiral Question Modal, Command Bar |
| **huh.Text** | Multi-line Admiral feedback when providing plan revision notes | Plan Review (feedback mode) |
| **huh.Confirm** | Destructive action confirmation (halt mission, force-kill agent, shelve plan), broadcast toggle on questions | Confirm Dialog, Admiral Question Modal |
| **huh.Form** | Groups multiple fields: question display + Select + Input + Confirm into a single flow | Admiral Question Modal, Plan Review |
| **huh.Group** | Multi-step form: step 1 = review manifest, step 2 = select action, step 3 = provide feedback | Plan Review |
| **huh themes** | `huh.ThemeCharm()` as base, customized with LCARS colors; `huh.ThemeBase()` for accessible mode | All Huh forms |

#### Lipgloss Styling Patterns

| Lipgloss Feature | Usage |
|-----------------|-------|
| **lipgloss.NewStyle()** | Base style factory for all component styles |
| **Border (Rounded)** | Overlay and modal borders (Charm aesthetic) |
| **Border (Normal/Thick)** | Panel borders on Ship Bridge and Ready Room (LCARS feel) |
| **Border (Double)** | Focused panel border indicator |
| **JoinHorizontal** | Side-by-side panel composition in standard layout |
| **JoinVertical** | Stacked panel composition in compact layout and within panels |
| **Place** | Centering modals and overlays within the terminal viewport |
| **Width/Height/MaxWidth/MaxHeight** | Constraining panel dimensions based on terminal size |
| **Foreground/Background** | LCARS color application |
| **Bold/Italic/Faint/Underline** | Text emphasis hierarchy |
| **Adaptive colors** | `lipgloss.AdaptiveColor` for light/dark terminal detection |
| **Color profiles** | `lipgloss.ColorProfile()` to detect ANSI/256/TrueColor support and degrade gracefully |

#### Custom Components (2+ Uses Justified)

| Custom Component | Usage Count | Locations | Justification |
|-----------------|-------------|-----------|---------------|
| **StatusBadge** | 8+ | Crew Panel, Mission Board, Ready Room specialists, Fleet Monitor, Agent Roster | Renders `[symbol] LABEL` with semantic color. Used everywhere status is shown. |
| **PanelFrame** | 10+ | Every panel in every view | Renders a Lipgloss-styled border with title, focus state, and optional subtitle. Wraps all panel content. |
| **NavigableToolbar** | 14 | Every view (bottom bar) | Renders shortcut buttons that are both pressable via quick key and arrow-navigable + Enter-selectable. Like htop/Midnight Commander bottom bar. |
| **PhaseIndicator** | 4+ | ACPhaseDetail (per AC), Crew Panel (inline phase), Mission Detail | Renders the 6-phase TDD pipeline as a horizontal indicator strip: `RED > V_RED > GRN > V_GRN > REF > V_REF` with current phase highlighted. |
| **WaveBar** | 3+ | Ship Bridge header (inline), Fleet Monitor (per ship), Plan Review | Renders `Wave N [====....] M/T` with progress fill using Bubbles progress component internally. |
| **ShipStatusRow** | 1+ (Fleet Monitor, potentially Fleet Overview compact) | Fleet Monitor grid | Single-line ship representation with inline progress, crew health, wave status. |
| **ACPhaseDetail** | 1 (Mission Detail) | Mission Detail view | Scrollable list of ACs with 6-dot TDD phase pipeline, inline gate results, attempt counts. |

Components used in only one location (like the dependency graph ASCII renderer in Wave Manager) are implemented inline within their parent model, not extracted as reusable components.

### Huh Form Compositions

#### Admiral Question Modal

```
huh.NewForm(
  huh.NewGroup(
    // Question display (read-only, rendered by Glamour)
    huh.NewNote().Title("ADMIRAL -- QUESTION FROM CAPTAIN").Description(glamourRendered),

    // Options selection
    huh.NewSelect[string]().
      Title("Select response").
      Options(huh.NewOptions(questionOptions...)...),

    // Free-text input (optional)
    huh.NewInput().
      Title("Additional context (optional)").
      Placeholder("Type here or press Enter to skip"),

    // Broadcast toggle
    huh.NewConfirm().
      Title("Broadcast answer to all agents?").
      Affirmative("Yes").
      Negative("No"),
  ),
).WithTheme(lcarsHuhTheme)
```

#### Plan Review Overlay

```
huh.NewForm(
  // Step 1: Review manifest (Glamour viewport, non-interactive)
  huh.NewGroup(
    huh.NewNote().Title("MISSION MANIFEST").Description(glamourManifest),
  ),

  // Step 2: Decision
  huh.NewGroup(
    huh.NewSelect[string]().
      Title("Admiral decision").
      Options(
        huh.NewOption("Approve -- dispatch missions", "approve"),
        huh.NewOption("Feedback -- reconvene Ready Room", "feedback"),
        huh.NewOption("Shelve -- save for later", "shelve"),
      ),
  ),

  // Step 3: Feedback (conditional, shown only if "feedback" selected)
  huh.NewGroup(
    huh.NewText().
      Title("Feedback for Ready Room").
      Placeholder("Describe what needs revision...").
      CharLimit(2000),
  ),
).WithTheme(lcarsHuhTheme)
```

#### Operator Command Bar

```
huh.NewInput().
  Title("").
  Prompt("> ").
  Placeholder("halt <id> | retry <id> | approve <id> | wip <n> | quit").
  WithTheme(lcarsHuhTheme)
```

For destructive commands (halt, force-kill):

```
huh.NewConfirm().
  Title("Halt mission MISSION-42?").
  Description("Worktree will be preserved. Mission returns to backlog.").
  Affirmative("Halt").
  Negative("Cancel").
  WithTheme(lcarsHuhTheme)
```

### Style Token System

All Lipgloss styles are defined in a centralized theme module. No hardcoded colors or styles in component code.

```go
// internal/tui/theme/lcars.go

package theme

import "github.com/charmbracelet/lipgloss"

// Reference tier: raw color values
var (
    Butterscotch   = lipgloss.Color("#FF9966")
    Blue           = lipgloss.Color("#9999CC")
    Purple         = lipgloss.Color("#CC99CC")
    Pink           = lipgloss.Color("#FF99CC")
    Gold           = lipgloss.Color("#FFAA00")
    Almond         = lipgloss.Color("#FFAA90")
    RedAlert       = lipgloss.Color("#FF3333")
    YellowCaution  = lipgloss.Color("#FFCC00")
    GreenOk        = lipgloss.Color("#33FF33")
    Black          = lipgloss.Color("#000000")
    DarkBlue       = lipgloss.Color("#1B4F8F")
    GalaxyGray     = lipgloss.Color("#52526A")
    SpaceWhite     = lipgloss.Color("#F5F6FA")
    LightGray      = lipgloss.Color("#CCCCCC")
    MoonlitViolet  = lipgloss.Color("#9966FF")
    Ice            = lipgloss.Color("#99CCFF")
)

// Semantic tier: meaning-assigned styles
var (
    ActiveStyle   = lipgloss.NewStyle().Foreground(Butterscotch)
    SuccessStyle  = lipgloss.NewStyle().Foreground(GreenOk)
    ErrorStyle    = lipgloss.NewStyle().Foreground(RedAlert)
    WarningStyle  = lipgloss.NewStyle().Foreground(YellowCaution)
    InfoStyle     = lipgloss.NewStyle().Foreground(Blue)
    PlanningStyle = lipgloss.NewStyle().Foreground(Purple)
    NotifyStyle   = lipgloss.NewStyle().Foreground(Pink)
    FocusStyle    = lipgloss.NewStyle().Foreground(MoonlitViolet)

    PanelBorder        = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(GalaxyGray)
    PanelBorderFocused = lipgloss.NewStyle().Border(lipgloss.DoubleBorder()).BorderForeground(MoonlitViolet)
    OverlayBorder      = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(Butterscotch)

    HeaderStyle = lipgloss.NewStyle().
                    Bold(true).
                    Foreground(SpaceWhite).
                    Background(DarkBlue).
                    Padding(0, 1)

    DimStyle = lipgloss.NewStyle().Foreground(LightGray).Faint(true)
)
```

---

## 7. Animation Strategy

### Harmonica Integration

Ship Commander 3 uses Harmonica (github.com/charmbracelet/harmonica) for physics-based animations in the TUI. Harmonica provides spring oscillators with configurable frequency and damping ratio, producing natural-feeling motion.

> **Detail**: Animation presets (warp_engage, beam_materialize, red_alert_klaxon, shields_up, etc.) and per-transition spring configs in `design/screens.yaml` (transitions section). Animation categories in `design/paradigm.yaml` (animation section).

### Spring Configuration Reference

| Damping Ratio | Behavior | Use Case |
|---------------|----------|----------|
| < 1.0 (under-damped) | Oscillates/bounces before settling | Attention-drawing alerts, notification badges |
| = 1.0 (critically damped) | Snaps to target without overshoot | View transitions, panel resize, modal open/close |
| > 1.0 (over-damped) | Slowly approaches target, no overshoot | Subtle background animations, progress bar smoothing |

### Animation Catalog

| Animation | Trigger | Spring Config | Target Property | Duration |
|-----------|---------|---------------|-----------------|----------|
| **Modal Open** | Admiral question appears, plan review opens | Frequency: 6.0, Damping: 1.0 (critically damped) | Modal Y position (slides in from top) and opacity (0 to 1) | ~200ms |
| **Modal Close** | Esc or form submission | Frequency: 6.0, Damping: 1.0 | Modal Y position (slides out) and opacity (1 to 0) | ~150ms |
| **View Transition** | Navigating between views (push/pop) | Frequency: 5.0, Damping: 1.0 (critically damped) | Content X position (horizontal slide) | ~250ms |
| **Red Alert Pulse** | Mission halted, critical gate failure | Frequency: 3.0, Damping: 0.4 (under-damped, bouncy) | Border color intensity oscillation (fades in/out) | 3 bounces over ~1.5s |
| **Panel Resize** | Terminal resize triggers compact/standard reflow | Frequency: 4.0, Damping: 1.2 (slightly over-damped, smooth) | Panel width/height dimensions | ~300ms |
| **Progress Bar Update** | Wave progress changes, WIP bar changes | Frequency: 3.0, Damping: 1.5 (over-damped, smooth) | Bar fill value | ~400ms |
| **Notification Badge** | New pending question, agent stuck detected | Frequency: 8.0, Damping: 0.3 (under-damped, very bouncy) | Badge scale (pops in) | 4 bounces over ~800ms |
| **Phase Advance** | AC transitions to next TDD phase | Frequency: 5.0, Damping: 0.8 (slightly under-damped) | Phase indicator position highlight shift | ~200ms with slight overshoot |
| **Success Checkmark** | Gate passes, mission completes | Frequency: 6.0, Damping: 0.6 (under-damped) | Checkmark scale (pops in with bounce) | ~300ms |

### Animation Implementation Pattern

Harmonica animations are driven by Bubble Tea's `tea.Tick` mechanism. The root model maintains animation state and dispatches tick messages at ~60fps during active animations.

```go
// Conceptual pattern (not exact API)

type animationState struct {
    spring   harmonica.Spring
    value    float64
    target   float64
    velocity float64
    active   bool
}

func (a *animationState) Update(dt float64) {
    a.value, a.velocity = a.spring.Update(a.value, a.velocity, a.target)
    if math.Abs(a.value - a.target) < 0.001 && math.Abs(a.velocity) < 0.001 {
        a.value = a.target
        a.active = false
    }
}
```

### Animation Preferences

Animations respect user preferences:

| Setting | Effect |
|---------|--------|
| `[tui] animations = true` | All animations enabled (default) |
| `[tui] animations = false` | All animations disabled; transitions are instant |
| `[tui] animation_speed = "fast"` | Frequency multiplied by 1.5 (quicker settling) |
| `[tui] animation_speed = "slow"` | Frequency multiplied by 0.7 (slower, more visible) |
| Executive mode | Animations disabled by default |
| `TERM` without TrueColor | Animations that depend on color interpolation are simplified to instant transitions |

### Performance Budget

Animations must not degrade TUI responsiveness:

- Animation tick rate: 60 FPS during active animation, 0 FPS when idle
- Maximum concurrent animations: 3
- If terminal width < 80 cols: animations disabled (not enough visual space to justify computation)
- Animation state is cheap (a few floats per spring); render cost is in the View() pass, which is already rate-limited by Bubble Tea

---

## 8. Progressive Disclosure

### Three Display Modes

Ship Commander 3 provides three display modes that control information density. The mode affects which panels are visible, how much detail each panel shows, and what terminology is used.

| Mode | Target Persona | Information Density | Default For |
|------|---------------|--------------------:|-------------|
| **Basic** | Junior Lieutenant | Low | First-time users (no `.sc3/` dir), `--basic` flag |
| **Advanced** | Staff Commander | High | Users with existing `.sc3/` dir, `--advanced` flag |
| **Executive** | Fleet Admiral | Aggregated | `--executive` flag, `[tui] mode = "executive"` |

Mode is toggled at runtime via `Ctrl+M` (cycles Basic -> Advanced -> Executive -> Basic) or via the command bar: `mode basic`, `mode advanced`, `mode executive`.

### Basic Mode

**Philosophy**: Show what matters. Hide complexity. Guide the user.

#### Ready Room (Basic)

```
+=============================================+
| PLANNING | auth-system | Iteration 2/5     |
+=============================================+
| Agents                                      |
|   Captain        [+] Done                   |
|   Commander      [>] Working...             |
|   Design Officer [>] Working...             |
|                                             |
| Questions: 0 pending                        |
+---------------------------------------------+
| [v] Review Plan  [?] Help  [Esc] Bridge    |
+=============================================+
```

Differences from Advanced:
- Inter-agent message log hidden (too noisy for novices)
- Directive viewport hidden by default (user knows what they submitted)
- Help shortcut prominently displayed
- Simplified terminology: "Agents" not "Specialists", "Done" not "CONSENSUS_REACHED"

#### Ship Bridge (Basic)

```
+=============================================+
| USS Auth | [OK] | 3 tasks active            |
+=============================================+
| Active Tasks                                |
|   Task 3  writing tests     02:15  [>]      |
|   Task 4  implementing      01:45  [>]      |
|   Task 5  under review      00:30  [~]      |
+---------------------------------------------+
| Progress                                    |
|   Batch 1 [========] 4/4 complete           |
|   Batch 2 [====....] 2/4 in progress        |
+---------------------------------------------+
| Recent Activity                             |
|   Test check for Task 3 -- needs fix  [!]   |
|   Task 4 tests passing                [+]   |
+---------------------------------------------+
| [r] Planning  [?] Help  [Esc] Fleet        |
+=============================================+
```

Differences from Advanced:
- Terminology: "Task" instead of "Mission", "Batch" instead of "Wave", "Active Tasks" instead of "Crew Panel"
- Kanban columns replaced with simple "Active Tasks" list
- Event Log simplified to "Recent Activity" with plain language
- NavigableToolbar shows simplified labels
- Phase detail always in drill-down (same as Advanced -- no Phase Tracker on Bridge)

### Advanced Mode

**Philosophy**: Show everything. Assume expertise. Full LCARS aesthetic.

This is the standard layout described in Section 4 with all panels visible:
- Crew Panel with full detail (role, mission ID, harness, phase, time)
- Mission Board with Kanban columns and counts
- Event Log with severity levels and structured log format
- NavigableToolbar with all context-sensitive shortcuts
- Inline wave summary in Ship Bridge header
- LCARS terminology throughout (Mission, Wave, Gate, Admiral, Ensign)
- Per-AC TDD phase detail available via Mission Detail drill-down
- Wave dependency graph available via Wave Manager

### Executive Mode

**Philosophy**: Aggregates only. No details unless drilled into.

#### Fleet Overview (Executive)

```
+============================================================+
| FLEET STATUS | 3 ships active                              |
+============================================================+
| Ship                | Progress | Missions | Health         |
|---------------------|----------|----------|----------------|
| USS Auth            | 60%      | 6/10     | [OK]           |
| USS Payments        | 25%      | 2/8      | [~] 1 stuck    |
| USS Docs            | 100%     | 4/4      | [+] complete   |
+============================================================+
| Velocity: 3.2 missions/day (trending up)                   |
| Blockers: 1 stuck agent (USS Payments)                     |
| Next milestone: USS Auth Wave 3 (est. 2h)                  |
+------------------------------------------------------------+
| [Enter] Bridge  [f] Monitor  [a] Approve all  [?] Help    |
+============================================================+
```

Differences from Advanced:
- No individual agent details (too granular)
- No event log (too noisy)
- Ship-level progress percentages (uses Fleet Overview naturally)
- Velocity metrics and trend direction
- Blocker summary (actionable items only)
- Drill-down available: Enter on a ship shows its Ship Bridge view
- Terminology: "Ship" (business term), "Progress" (percentage), "Health" (traffic light)

### Mode-Specific Panel Visibility

| Panel | Basic | Advanced | Executive |
|-------|-------|----------|-----------|
| Fleet Overview Header | Simplified ship count | Full with health, WIP | Fleet-level metrics |
| Ship Bridge Header | Ship name + health icon | Full ship name, directive, wave bar | Ship-level summary |
| Crew Panel | Simplified "Active Tasks" | Full agent table | Hidden |
| Mission Board | Hidden (merged into task list) | Full Kanban | Ship summary table |
| Event Log | Simplified "Recent Activity" | Full structured log | Hidden |
| NavigableToolbar | Simplified labels | Full labels with all shortcuts | Simplified with bulk actions |
| Wave Summary | Abbreviated in header | Inline in header with progress bar | Hidden (% shown per ship) |
| Phase Detail (Mission Detail) | Available via drill-down | Available via drill-down | Hidden |
| Dependency Graph (Wave Manager) | Hidden | Available via `[w]` | Hidden |
| Directive Viewport (Ready Room) | Hidden by default | Toggle with `[t]` | Hidden |

### Terminology Translation

| Concept | Basic Mode | Advanced Mode | Executive Mode |
|---------|-----------|---------------|----------------|
| Mission | Task | Mission | Mission |
| Wave | Batch | Wave | Sprint |
| Agent | Helper | Agent / Ensign | Worker |
| Gate | Check | Gate | Quality gate |
| Commission | Project | Commission | Initiative |
| Admiral | You | Admiral | Approver |
| Ready Room | Planning | Ready Room | Planning phase |
| Ship Bridge | Dashboard | Ship Bridge | Status board |
| Halted | Stopped | Halted / Red Alert | Blocked |
| Doctor | Health check | Doctor | Monitoring |

### Auto-Detection Heuristic

On first launch, Ship Commander 3 selects the initial mode:

| Signal | Detected Mode |
|--------|--------------|
| No `.sc3/` directory exists | Basic (first-time user) |
| `--executive` flag | Executive |
| `--basic` flag | Basic |
| `--advanced` flag | Advanced |
| `.sc3/` exists + no flag | Advanced (returning user) |
| `[tui] mode` in config | Configured mode |

The mode can always be overridden at runtime via `Ctrl+M` or the `mode` command.

### Onboarding (Basic Mode Only)

On first launch in Basic mode, a one-time help overlay appears:

```
+=============================================+
|          Welcome to Ship Commander          |
|                                             |
|  Ship Commander orchestrates AI coding      |
|  agents to build software from your         |
|  requirements.                              |
|                                             |
|  KEY CONCEPTS:                              |
|                                             |
|  Tasks   -- Work items for AI agents        |
|  Batches -- Groups of tasks run together    |
|  Checks  -- Automated quality gates         |
|  Helpers -- AI agents that write code       |
|                                             |
|  NAVIGATION:                                |
|                                             |
|  Tab       -- Move between panels           |
|  Enter     -- Select / view details         |
|  Esc       -- Go back                       |
|  ?         -- Show this help anytime        |
|  Ctrl+M    -- Switch to advanced mode       |
|                                             |
|         [Enter] to begin                    |
+=============================================+
```

This overlay is shown once and never again (tracked in `.sc3/onboarded`). The user can re-trigger it via `? onboarding` in the command bar.

> **Detail**: Display mode logic and per-mode panel visibility defined in `design/views.yaml` (per-view `mode_differences`) and `design/screens.yaml` (per-screen `mode_availability`).

---

## 9. Verification Gates and Settings

### Design Philosophy: Language-Agnostic Gates

Verification gates are **bash commands** -- not hardcoded language-specific logic. The Commander runs these commands in the mission's worktree directory and checks exit codes. This means Ship Commander works for Go, TypeScript, Python, Rust, or any language without modification.

Gate commands are configured in the **Project Settings** view (accessible via `[s]` from Fleet Overview) and stored globally at `~/.sc3/config.json`. They are NOT per-project -- the same gates apply regardless of which project directory you launch from.

### Default Gate Commands

| Gate | Default Command | Purpose |
|------|----------------|---------|
| `test_command` | `go test ./...` | Run test suite (VERIFY_RED, VERIFY_GREEN, VERIFY_REFACTOR) |
| `typecheck_command` | `go vet ./...` | Type checking (VERIFY_IMPLEMENT) |
| `lint_command` | `golangci-lint run` | Linting (VERIFY_IMPLEMENT) |
| `build_command` | `go build ./...` | Build verification (VERIFY_IMPLEMENT) |

Commands support variable substitution: `{test_file}`, `{worktree_dir}`, `{mission_id}`.

### Settings Export/Import

Settings can be exported to a JSON file and imported in a different project:

```
Export:  All settings (gates + crew + fleet) → sc3-settings.json
Import:  sc3-settings.json → Apply to current environment
```

**What's included**: Verification gate commands, crew defaults (harness, model, WIP limits, timeouts), fleet defaults (naming, wave strategy, merge policy).

**What's NOT included**: Tasks, missions, commissions, beads state. These are project-specific.

This is the primary mechanism for reusing configuration across projects. When switching projects, import your settings file and everything is restored.

### Project Settings View

```
+============================================================================+
| PROJECT SETTINGS | ~/.sc3/config.json                                      |
+============================================================================+
| [Gates]  Crew  Fleet  Export                                               |
| ──────────────────────────────────────────────────────────────────────────  |
|  test_command:       go test ./...                          [Edit] [x]     |
|  typecheck_command:  go vet ./...                           [Edit] [x]     |
|  lint_command:       golangci-lint run                      [Edit] [x]     |
|  build_command:      go build ./...                         [Edit] [x]     |
|                                                   [+ Add Gate Command]     |
+============================================================================+
| [1] Gates  [2] Crew  [3] Fleet  [4] Export  [?] Help  [Esc] Fleet        |
+============================================================================+
```

> **Detail**: Full view specification in `design/views.yaml` (VIEW 15: project-settings). Flow in `design/flows.yaml` (flow-16: project-settings-configuration). Components in `design/components.yaml` (SettingsTabs, GateCommandEditor).

---

## Appendix A: Screen Relationship Map

```
                        +-------------+
                        |   sc3 tui   |
                        +------+------+
                               |
                    +----------+----------+
                    |                     |
            +-------+--------+    +------+-------+
            | Fleet Overview |    | (top-level)  |
            | (landing)      |    | Agent Roster |
            +--+------+------+    | Dir. Editor  |
               |      |          | Msg Center   |
               |      |          | Settings     |
    +----------+      +-------+  +------+-------+
    |                         |
+---+----------+     +-------+--------+
| Ship Bridge  |     | Fleet Monitor  |
| (per-ship)   |     | (condensed)    |
+--+-----+--+--+     +----------------+
   |     |  |
   |     |  +-- Wave Manager (overlay)
   |     |
+--+---+ +--+--------+
|Ready | |Mission Det.|
|Room  | |Agent Detail|
+--+--++ +------------+
   |  |
   |  +-- Specialist Detail
   |
   |
+--+----------+
| Plan Review |
| (full-screen)|
+-------------+

    OVERLAYS (modal layer, atop any view):

      +-------------------+    +-----------------+
      | Admiral Question  |    | Help Overlay    |
      | (Huh Form)        |    | (keybindings)   |
      +-------------------+    +-----------------+

      +-------------------+    +-------------------+
      | Confirm Dialog    |    | Onboarding Overlay|
      | (Huh Confirm)     |    | (first-run tour)  |
      +-------------------+    +-------------------+
```

## Appendix B: Bubble Tea Model Architecture

```
RootModel (smart main)
  |
  |-- navStack: []ViewID (fleet-overview → ship-bridge → ready-room → plan-review)
  |-- mode: DisplayMode (Basic | Advanced | Executive)
  |-- width, height: int
  |-- overlayStack: []OverlayModel
  |-- animations: []AnimationState
  |-- eventBusSub: <-chan Event
  |-- toolbarHighlight: int (-1 = none)
  |
  |-- Init() -> tea.Cmd
  |     |-- Subscribe to event bus
  |     |-- Detect terminal size
  |     |-- Auto-detect display mode
  |     |-- Show onboarding if first launch
  |     |-- Push fleet-overview onto navStack
  |
  |-- Update(msg tea.Msg) -> (tea.Model, tea.Cmd)
  |     |-- tea.WindowSizeMsg -> recalculate layout
  |     |-- tea.KeyMsg -> route to overlay stack, toolbar, or active view
  |     |-- EventBusMsg -> update state, trigger animations
  |     |-- AnimationTickMsg -> advance spring states
  |     |-- HuhFormCompleteMsg -> process Admiral responses
  |     |-- NavigateMsg -> push/pop navStack
  |
  |-- View() -> string
        |-- Render active view (top of navStack)
        |-- Each view renders: header + content + NavigableToolbar
        |-- If overlay stack non-empty: render overlay on top
        |-- Apply animations to rendered output
        |-- lipgloss.Place() for final composition

FleetOverviewView (dumb component)
  |-- Receives: ships[], selectedIndex, previewShip, mode
  |-- Returns: string (header + ship list + preview + toolbar)
  |-- Toolbar: [Enter] Bridge  [n] New  [f] Monitor  [a] Roster  [i] Inbox  [s] Settings

ShipBridgeView (dumb component)
  |-- Receives: ship, agents, missions, events, waveProgress, mode
  |-- Returns: string (header with wave summary + crew + board + log + toolbar)
  |-- Toolbar: [r] Ready Room  [w] Waves  [h] Halt  [Space] Pause  [Esc] Fleet

ReadyRoomView (dumb component)
  |-- Receives: directive, specialists, planningStatus, mode
  |-- Returns: string (header + specialist grid + directive viewport + toolbar)
  |-- Toolbar: [v] Review Plan  [t] Toggle Directive  [Esc] Bridge

PlanReviewView (dumb component)
  |-- Receives: manifest, coverageMatrix, dependencyGraph, mode
  |-- Returns: string (header + manifest viewport + analysis panel + toolbar)
  |-- Toolbar: [a] Approve  [f] Feedback  [s] Shelve  [Esc] Ready Room

SpecialistDetailView (dumb component)
  |-- Receives: specialist, outputLog, status, assignment, mode
  |-- Returns: string (header + role badge + output viewport + status + toolbar)
  |-- Toolbar: [Esc] Ready Room

MissionDetailView (dumb component)
  |-- Receives: mission, acPhaseData[], gateEvidence, outputViewport
  |-- Returns: string (header + ACPhaseDetail + evidence + output + toolbar)
  |-- Toolbar: context-sensitive (halt/retry/approve based on mission state)

FleetMonitorView (dumb component)
  |-- Receives: ships[], mode
  |-- Returns: string (header + ShipStatusRow grid + toolbar)
  |-- Toolbar: [Enter] Bridge  [i] Inbox  [Esc] Fleet

ProjectSettingsView (dumb component)
  |-- Receives: gateCommands[], crewDefaults, fleetDefaults, activeTab, mode
  |-- Returns: string (header + tabbed settings + toolbar)
  |-- Toolbar: [1] Gates  [2] Crew  [3] Fleet  [4] Export  [Esc] Fleet

NavigableToolbar (shared render function)
  |-- Receives: buttons[], highlightedIndex, width
  |-- Returns: string (bottom bar with labeled shortcut buttons)
  |-- Arrow Left/Right changes highlight, Enter activates, quick keys work directly

OverlayModel (interface)
  |-- AdmiralQuestionOverlay (wraps huh.Form)
  |-- WaveManagerOverlay (wraps dependency graph + merge controls)
  |-- HelpOverlay (wraps help.Model)
  |-- OnboardingOverlay (first-run welcome tour, Basic mode only)
  |-- ConfirmOverlay (wraps huh.Confirm)
```

## Appendix C: Color Palette Quick Reference

```
PRIMARY LCARS                    SEMANTIC STATUS
+--------+--------+             +--------+-----------+
|  butt  |  blue  |             | red    | error     |
| #FF99  | #9999  |             | #FF33  | halt      |
|   66   |   CC   |             |   33   | fail      |
+--------+--------+             +--------+-----------+
| purple |  pink  |             | yellow | warning   |
| #CC99  | #FF99  |             | #FFCC  | stuck     |
|   CC   |   CC   |             |   00   | caution   |
+--------+--------+             +--------+-----------+
|  gold  | almond |             | green  | success   |
| #FFAA  | #FFAA  |             | #33FF  | done      |
|   00   |   90   |             |   33   | pass      |
+--------+--------+             +--------+-----------+

BACKGROUNDS & NEUTRALS           CHARM ACCENTS
+--------+--------+             +--------+-----------+
| black  | dk_blu |             | violet | focus     |
| #0000  | #1B4F  |             | #9966  | highlight |
|   00   |   8F   |             |   FF   | selected  |
+--------+--------+             +--------+-----------+
| galaxy | sp_wht |             |  ice   | cool info |
| #5252  | #F5F6  |             | #99CC  | secondary |
|   6A   |   FA   |             |   FF   | panels    |
+--------+--------+             +--------+-----------+
| lt_gry |        |
| #CCCC  |        |
|   CC   |        |
+--------+--------+
```

---

**END OF DOCUMENT**
