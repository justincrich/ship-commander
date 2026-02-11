# Ready Room -- ASCII Blueprint

> Source: `ready-room.mock.html` | Spec: `prompts/ready-room.spec.json`
> Terminal: 120 cols x 30 rows (standard breakpoint)

## Layout Structure

```
ReadyRoomHeader | DirectiveSidebar (30%) + CrewSessionGrid (70%) | NavigableToolbar
```

## Component Mapping (from components.yaml)

| Visual Area | Component | Charm Base | Key Props |
|---|---|---|---|
| Full screen container | **AppShell** (#1) | `lipgloss.JoinVertical` | header, body, footer |
| Ready Room header | **HeaderBar** (ready-room variant) | `lipgloss.NewStyle()` | ship name, directive, iteration, purple theme |
| Directive + Crew side-by-side | **PanelGrid** (#2) | `lipgloss.JoinHorizontal` | ratios: `[0.3, 0.7]`, compact_threshold: 120 |
| Directive sidebar | **FocusablePanel** (#4) wrapping `bubbles.viewport` + Glamour | `bubbles.viewport` | Glamour-rendered PRD markdown, scrollable |
| Crew session grid | **SpecialistGrid** (#25) | `lipgloss.JoinHorizontal` | 2-3 columns of SpecialistCards |
| Individual specialist panel | **SpecialistCard** (#24) | `lipgloss.NewStyle().Border(lipgloss.RoundedBorder())` | name, role, status, harness, model, elapsed, output_snippet |
| Status badges | **StatusBadge** (#6) | `lipgloss.NewStyle()` | working/done/waiting/skipped/failed variants |
| Elapsed timers | **ElapsedTimer** (#10) | `bubbles.stopwatch` | per-specialist timing |
| Loading spinners | `bubbles.spinner` | `bubbles.spinner` | during specialist startup |
| Planning iteration badge | Custom `lipgloss.NewStyle()` | `lipgloss.NewStyle()` | "Iteration N/M" in purple |
| Sign-off tracker | Custom `lipgloss.NewStyle()` | `lipgloss.NewStyle()` | Captain: ✓, Commander: ✓ |
| Bottom toolbar | **NavigableToolbar** | `lipgloss.JoinHorizontal` | [a] Approve [f] Feedback [s] Shelve [v] Review [t] Toggle [?] Help [Esc] Bridge |

## Token Reference

| Token             | Hex       | Usage in Ready Room                                      |
|-------------------|-----------|----------------------------------------------------------|
| `purple`          | `#CC99CC` | Header title, planning-phase borders, spinner             |
| `blue`            | `#9999CC` | Directive sidebar border, section headers, info text      |
| `butterscotch`    | `#FF9966` | Active/working borders, toolbar keys, directive title     |
| `green_ok`        | `#33FF33` | Done status badges, completed specialist borders          |
| `galaxy_gray`     | `#52526A` | Waiting/inactive borders, separator lines                 |
| `space_white`     | `#F5F6FA` | Primary text                                              |
| `light_gray`      | `#CCCCCC` | Secondary text, metadata, elapsed times                   |
| `dark_blue`       | `#1B4F8F` | Planning status bar background                            |
| `red_alert`       | `#FF3333` | Uncovered use case marker                                 |
| `yellow_caution`  | `#FFCC00` | Partial coverage marker                                   |
| `moonlit_violet`  | `#9966FF` | Focused panel border                                      |
| `pink`            | `#FF99CC` | Notifications (not shown in default state)                |

### Icons (from tokens.yaml)

| Icon  | Unicode  | Meaning              |
|-------|----------|----------------------|
| `✓`   | `\u2713` | Done / Pass          |
| `●`   | `\u25CF` | Working              |
| `⏸`   | `\u23F8` | Waiting              |
| `✗`   | `\u2717` | Failed / Uncovered   |
| `⚠`   | `\u26A0` | Warning / Partial    |
| `▸`   | `\u25B8` | Running agent        |

## ASCII Blueprint (120x30 standard)

```
╭─ READY ROOM ── USS ENTERPRISE ──────────────────────────────────────────────────────────────────────────────────────╮
│ Directive: Multi-Region Deployment System          Iteration 2/3    Questions: 0    Crew: 4 active                  │
╰────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╯
╭─ Planning ─────────────────────────────────────────────────────────────────────────────────────────────────────────╮
│  Planning: Iteration 2/3          Coverage: 85%          Sign-offs: 2/4                                            │
╰────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╯
╭─ Directive ──────────────────╮ ╭─ Cmdr. Data ──────────────────── ● WORKING ─╮╭─ Lt.Cmdr. LaForge ──────── ✓ DONE ─╮
│                              │ │  System Architect                            ││  Infrastructure Lead               │
│  Multi-Region Deployment     │ │  claude-opus-4-6        ⏱ 2m 34s            ││  claude-sonnet-4-5     ⏱ 4m 12s   │
│  System                      │ │ ─────────────────────────────────────────────││─────────────────────────────────── │
│                              │ │  Analyzing deployment topology               ││  ✓ Regional cluster config         │
│  Implement a robust multi-   │ │  constraints...                              ││    validated                       │
│  region deployment pipeline  │ │  Proposing orchestration layer               ││  ✓ Network topology for            │
│  that enables seamless       │ │  using Kubernetes operators                   ││    cross-region traffic defined    │
│  rollouts across global      │ │  Cross-region state sync requires            ││  ✓ Multi-AZ deployment strategy    │
│  infrastructure with zero-   │ │  distributed consensus                       ││    with 3 zones per region         │
│  downtime guarantees and     │ ╰──────────────────────────────────────────────╯╰────────────────────────────────────╯
│  automated rollback.         │ ╭─ Lt. Torres ─────────────────── ● WORKING ─╮╭─ Ens. Kim ──────────────── ✓ DONE ─╮
│                              │ │  Backend Engineer                            ││  DevOps Specialist                 │
│  Use Cases                   │ │  claude-sonnet-4-5      ⏱ 1m 58s            ││  claude-sonnet-4-5     ⏱ 3m 47s   │
│  ✓ Deploy to multiple        │ │ ─────────────────────────────────────────────││─────────────────────────────────── │
│    regions simultaneously    │ │  Designing health check endpoints            ││  ✓ CI/CD pipeline stages defined   │
│  ✓ Progressive rollout with  │ │  for regional services...                    ││    for multi-region flow           │
│    canary testing            │ │  HTTP /health and /ready endpoints           ││  ✓ Canary deployment strategy:     │
│  ⚠ Automated health checks  │ │  with dependency checks                      ││    5% → 25% → 50% → 100%         │
│    and rollback              │ │  Implementing circuit breaker                ││  ✓ Rollback automation triggers    │
│  ✗ Cross-region traffic      │ │  pattern for inter-region calls              ││    on error rate > 2%              │
│    management                │ ╰──────────────────────────────────────────────╯╰────────────────────────────────────╯
│  ⚠ Deployment status        │
│    monitoring dashboard      │
│                              │
│  Acceptance Criteria         │
│  • Zero-downtime deployments │
│    verified                  │
│  • Rollback within 60s       │
│  • Regional health active    │
│  • Deployment metrics tracked│
│                              │
│  Scope                       │
│  In: Orchestration, health   │
│   checks, rollback, failover │
│  Out: Infra provisioning,    │
│   DNS mgmt, CDN config       │
╰──────────────────────────────╯
 [a] Approve  [f] Feedback  [s] Shelve  [v] Review Plan  [t] Toggle Directive  [?] Help  [Esc] Bridge
```

## Color Annotations

The blueprint above uses structural characters only. Below maps regions to their token colors.

```
REGION                          BORDER COLOR       TEXT COLOR         BACKGROUND
────────────────────────────────────────────────────────────────────────────────
ReadyRoomHeader (row 1-3)       purple #CC99CC     space_white        black #000000
  "READY ROOM" title            --                 purple #CC99CC     --
  "Directive: ..." subtitle     --                 blue #9999CC       --
  meta items                    --                 light_gray #CCCCCC --

Planning Status Bar (row 4-6)   blue #9999CC       space_white        dark_blue #1B4F8F
  labels ("Planning:", etc.)    --                 light_gray #CCCCCC --
  values ("Iteration 2/3")     --                 butterscotch       --

Directive Sidebar (row 7-38)    purple #CC99CC     --                 black #000000
  h3 title                      --                 butterscotch       --
  h4 section headers            --                 blue #9999CC       --
  body text                     --                 light_gray #CCCCCC --
  ✓ covered                     --                 green_ok #33FF33   --
  ⚠ partial                    --                 yellow_caution     --
  ✗ uncovered                   --                 red_alert #FF3333  --

Specialist Card (WORKING)       butterscotch       --                 #0a0a0a
  agent name                    --                 blue #9999CC       --
  role badge                    --                 purple #CC99CC     --
  status badge "● WORKING"      --                 black #000000      butterscotch
  model + timer                 --                 light_gray #CCCCCC --
  output lines                  --                 light_gray #CCCCCC --

Specialist Card (DONE)          green_ok #33FF33   --                 #0a0a0a
  agent name                    --                 blue #9999CC       --
  role badge                    --                 purple #CC99CC     --
  status badge "✓ DONE"         --                 black #000000      green_ok
  model + timer                 --                 light_gray #CCCCCC --
  output lines                  --                 light_gray #CCCCCC --

Specialist Card (WAITING)       galaxy_gray        --                 #0a0a0a
  status badge "⏸ WAITING"     --                 space_white        galaxy_gray

NavigableToolbar (bottom row)   --                 --                 black #000000
  key brackets [a], [f], etc.   --                 butterscotch       --
  labels "Approve", etc.        --                 space_white        --
```

## Compact Variant (80x24 minimum)

In compact mode (`<120 cols`), the directive sidebar collapses (togglable with `[t]`),
and crew session cards stack vertically in a single column.

```
╭─ READY ROOM ── USS ENTERPRISE ────────────────────────────────────────────────╮
│ Directive: Multi-Region Deployment System    Iter 2/3  Q: 0  Crew: 4         │
╰──────────────────────────────────────────────────────────────────────────────╯
╭─ Planning: Iter 2/3 ──── Coverage: 85% ──── Sign-offs: 2/4 ─────────────────╮
╰──────────────────────────────────────────────────────────────────────────────╯
╭─ Cmdr. Data ── System Architect ──────────────────────────── ● WORKING ─────╮
│  claude-opus-4-6  ⏱ 2m 34s                                                  │
│  Analyzing deployment topology constraints...                                │
│  Proposing orchestration layer using Kubernetes operators                     │
╰──────────────────────────────────────────────────────────────────────────────╯
╭─ Lt. Cmdr. LaForge ── Infrastructure Lead ────────────────── ✓ DONE ────────╮
│  claude-sonnet-4-5  ⏱ 4m 12s                                                │
│  ✓ Regional cluster configuration validated                                  │
│  ✓ Network topology for cross-region traffic defined                         │
╰──────────────────────────────────────────────────────────────────────────────╯
╭─ Lt. Torres ── Backend Engineer ──────────────────────────── ● WORKING ─────╮
│  claude-sonnet-4-5  ⏱ 1m 58s                                                │
│  Designing health check endpoints for regional services...                   │
│  HTTP /health and /ready endpoints with dependency checks                    │
╰──────────────────────────────────────────────────────────────────────────────╯
╭─ Ens. Kim ── DevOps Specialist ───────────────────────────── ✓ DONE ────────╮
│  claude-sonnet-4-5  ⏱ 3m 47s                                                │
│  ✓ CI/CD pipeline stages defined for multi-region flow                       │
│  ✓ Canary deployment strategy: 5% → 25% → 50% → 100%                       │
╰──────────────────────────────────────────────────────────────────────────────╯
 [a] Approve  [f] Feedback  [s] Shelve  [v] Review  [t] Directive  [?]  [Esc]
```

### Compact Variant Notes

- **Directive sidebar**: Hidden by default in compact mode. Press `[t]` to toggle it
  as a full-width panel above the crew grid.
- **Crew cards**: Stack vertically, single column. Each card shows 2 output lines
  instead of 3.
- **Toolbar**: Abbreviated labels to fit 80 columns.
- **Status bar**: Condensed to single line with shortened labels.
- **No side-by-side panels**: All content flows top-to-bottom.
