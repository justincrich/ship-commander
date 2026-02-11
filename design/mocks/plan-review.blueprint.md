# Plan Review -- ASCII Blueprint

> Source: `plan-review.mock.html` | Spec: `prompts/plan-review.spec.json`
> Terminal: 120 cols x 30 rows (standard breakpoint)

## Layout Structure

```
Row 1-3:   PlanReviewHeader (100%)
           Line 1: "PLAN REVIEW -- <Ship Name>" bold butterscotch (#FF9966)
           Line 2: Stats row -- Missions: N  Waves: N  Coverage: N%  Sign-offs: N/M
           Line 3: Separator ─── (galaxy_gray #52526A)

Row 4-27:  Content Grid (two-column, 50/50 split)
           Left:   ManifestViewport (scrollable, 50%)
           Right:  AnalysisPanel (stacked, 50%)
                   Top:    CoverageMatrix
                   Bottom: DependencyGraph

Row 28-29: NavigableToolbar (100%)
           [a] Approve  [f] Feedback  [s] Shelve  [?] Help  [Esc] Ready Room

Row 30:    Scroll hint (right-aligned, galaxy_gray)
```

## Component Mapping (from components.yaml)

| Visual Area | Component | Charm Base | Key Props |
|---|---|---|---|
| Full screen container | **AppShell** (#1) | `lipgloss.JoinVertical` | header, body, footer |
| Plan review header | **HeaderBar** (plan-review variant) | `lipgloss.NewStyle()` | directive title, iteration, coverage % |
| Manifest + Analysis side-by-side | **PanelGrid** (#2) | `lipgloss.JoinHorizontal` | ratios: `[0.5, 0.5]` |
| Manifest viewport | **FocusablePanel** (#4) wrapping `bubbles.viewport` | `bubbles.viewport` | scrollable mission manifest |
| Coverage matrix | **FocusablePanel** (#4) wrapping **CoverageMatrix** (#18) | `bubbles.table` | UC ID, Mapped Missions, Coverage Status |
| Dependency graph | **FocusablePanel** (#4) wrapping **DependencyGraph** (#17) | `lipgloss.NewStyle()` | wave headers, tree branches, dep lines |
| Mission cards in manifest | **MissionCard** (#14) | `lipgloss.NewStyle()` | mission_id, classification, wave, ACs |
| Classification badges | Custom `lipgloss.NewStyle()` | `lipgloss.NewStyle()` | [RED_ALERT] red_alert, [STANDARD_OPS] blue |
| Gate result rows | **GateResultRow** (#19) | `lipgloss.NewStyle()` | accept/reject/running/skipped variants |
| Coverage status icons | Custom `lipgloss.NewStyle()` | `lipgloss.NewStyle()` | ✓ green_ok, ✗ red_alert, ⚠ yellow_caution |
| Wave progress bars | **WaveProgressBar** (#9) | `bubbles.progress` | per-wave completion |
| Status badges | **StatusBadge** (#6) | `lipgloss.NewStyle()` | approved/review/planning variants |
| Bottom toolbar | **NavigableToolbar** | `lipgloss.JoinHorizontal` | [a] Approve [f] Feedback [s] Shelve [?] Help [Esc] Ready Room |

## Token Reference

| Token            | Hex       | Usage in this view                                    |
|------------------|-----------|-------------------------------------------------------|
| butterscotch     | `#FF9966` | Header title, mission seq IDs, toolbar keys, active   |
| almond           | `#FFAA90` | Mission sequence numbers (M-001, M-002, ...)          |
| blue             | `#9999CC` | Panel titles, stat labels, STANDARD_OPS badge bg      |
| green_ok         | `#33FF33` | Coverage %, covered checkmarks (✓)                    |
| yellow_caution   | `#FFCC00` | Sign-off count, partial coverage (⚠), ELEVATED_RISK   |
| red_alert        | `#FF3333` | RED_ALERT badge, uncovered cross (✗)                  |
| space_white      | `#F5F6FA` | Primary text, mission titles                          |
| light_gray       | `#CCCCCC` | Metadata, mission meta line, toolbar labels           |
| galaxy_gray      | `#52526A` | Borders, separators, dependency lines, scroll hints   |
| dark_blue        | `#1B4F8F` | Mission item separator lines, STANDARD_OPS badge bg   |
| moonlit_violet   | `#9966FF` | Focused panel border (not shown in default state)     |
| black            | `#000000` | Background                                            |

**Icons** (from `tokens.yaml`):
- `✓` (U+2713) -- covered / pass
- `✗` (U+2717) -- uncovered / fail
- `⚠` (U+26A0) -- partial coverage / warning
- `█` (U+2588) -- progress filled
- `░` (U+2591) -- progress empty
- `▸` (U+25B8) -- running / expand

**Borders**: Lipgloss `RoundedBorder()` -- chars: `╭ ╮ ╰ ╯ │ ─`

## ASCII Blueprint (120x30 standard)

```
╭──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╮
│ PLAN REVIEW ── USS Enterprise                                                                                      │
│ Missions: 8    Waves: 3    Coverage: 92%    Sign-offs: 3/4                                                         │
├──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
│ ╭─ Mission Manifest ──────────────────────────╮ ╭─ Coverage Matrix ─────────────────────────────────────╮           │
│ │                                             │ │ Use Case                       ✓  Missions            │           │
│ │ M-001  Initialize warp core systems         │ │ ────────────────────────────────────────────────────── │           │
│ │        [STANDARD_OPS]                       │ │ UC-001: Warp initialization     ✓  M-001, M-007       │           │
│ │        Wave: 1 │ UC: 001,002 │ ACs: 4      │ │ UC-002: Power distribution      ✓  M-001, M-007       │           │
│ │        Surface: core/warp                   │ │ UC-003: Shield modulation       ✓  M-002, M-007       │           │
│ │ ─────────────────────────────────────────── │ │ UC-004: Long-range scan         ✓  M-003, M-007       │           │
│ │ M-002  Configure shield harmonics           │ │ UC-005: Anomaly detection       ✓  M-003, M-007       │           │
│ │        [STANDARD_OPS]                       │ │ UC-006: Phaser targeting        ✓  M-004, M-007       │           │
│ │        Wave: 1 │ UC: 003 │ ACs: 3          │ │ UC-007: Torpedo guidance         ✓  M-004, M-007       │           │
│ │        Surface: defense/shields             │ │ UC-008: Atmosphere control      ✓  M-005, M-007       │           │
│ │ ─────────────────────────────────────────── │ │ UC-009: Emergency containment   ✓  M-006, M-007       │           │
│ │ M-003  Establish sensor array baseline      │ │ UC-010: Failsafe procedures     ⚠  M-006              │           │
│ │        [STANDARD_OPS]                       │ │ UC-011: Subspace relay          ✓  M-008              │           │
│ │        Wave: 1 │ UC: 004,005 │ ACs: 5      │ │ UC-012: Distress beacon         ✗  ─                  │           │
│ │        Surface: sensors/array               │ ├─ Dependency Graph ──────────────────────────────────── │           │
│ │ ─────────────────────────────────────────── │ │ Wave 1                                                │           │
│ │ M-004  Calibrate weapons targeting systems  │ │ ├─ M-001 Initialize warp core systems                 │           │
│ │        [ELEVATED_RISK]                      │ │ ├─ M-002 Configure shield harmonics                   │           │
│ │        Wave: 2 │ UC: 006,007 │ ACs: 6      │ │ └─ M-003 Establish sensor array baseline              │           │
│ │        Surface: weapons/phasers             │ │                                                       │           │
│ │ ─────────────────────────────────────────── │ │ Wave 2 (depends on Wave 1)                            │           │
│ │ M-005  Synchronize life support networks    │ │ ├─ M-004 Calibrate weapons targeting                  │           │
│ │        ...                        ▲ ▼ scroll│ │ │  └─ requires M-003 (sensor data)                    │           │
│ │                                             │ │ ├─ M-005 Synchronize life support                     │           │
│ ╰─────────────────────────────────────────────╯ │ │  └─ requires M-001 (power grid)                     │           │
│                                                 │ └─ M-006 Emergency containment                        │           │
│                                                 │    ├─ requires M-001 (power)                          │           │
│                                                 │    └─ requires M-002 (shields)                        │           │
│                                                 │                                                       │           │
│                                                 │ Wave 3 (final integration)                            │           │
│                                                 │ ├─ M-007 Full bridge integration                      │           │
│                                                 │ │  └─ requires M-001 through M-006                    │           │
│                                                 │ └─ M-008 Communication relay                          │           │
│                                                 │    └─ requires M-003 (sensors)                        │           │
│                                                 ╰───────────────────────────────────────────────────────╯           │
├──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
│ [a] Approve    [f] Feedback    [s] Shelve    [?] Help    [Esc] Ready Room                                          │
╰──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╯
```

## Color Annotations

Overlaid on the blueprint above, these annotations indicate which token colors apply to each element.

```
Line  Element                          Color Token        Hex
────  ───────────────────────────────  ─────────────────  ─────────
  1   Outer border                     galaxy_gray        #52526A
  2   "PLAN REVIEW ── USS Enterprise"  butterscotch       #FF9966
  3   "Missions:" label                light_gray         #CCCCCC
  3   "8" value                        blue               #9999CC
  3   "Coverage:" label                light_gray         #CCCCCC
  3   "92%" value                      green_ok           #33FF33
  3   "Sign-offs:" label               light_gray         #CCCCCC
  3   "3/4" value                      yellow_caution     #FFCC00
  4   Horizontal rule                  galaxy_gray        #52526A
  5   "╭─ Mission Manifest ─╮"         blue               #9999CC
  5   "╭─ Coverage Matrix ─╮"          blue               #9999CC
  7   "M-001"                          almond             #FFAA90
  7   Mission title text               space_white        #F5F6FA
  8   [STANDARD_OPS] badge             blue bg            #1B4F8F on #9999CC
  8   [ELEVATED_RISK] badge            yellow_caution bg  #FFCC00 on #000000
  8   [RED_ALERT] badge                red_alert bg       #FF3333 on #000000
  9   Meta line (Wave, UC, ACs)        light_gray         #CCCCCC
  -   Coverage ✓ (covered)             green_ok           #33FF33
  -   Coverage ⚠ (partial)             yellow_caution     #FFCC00
  -   Coverage ✗ (uncovered)           red_alert          #FF3333
  -   Coverage table headers           blue               #9999CC
  -   Coverage use-case text           space_white        #F5F6FA
  -   Coverage mission refs            space_white        #F5F6FA
  -   "Wave 1", "Wave 2", "Wave 3"    blue               #9999CC
  -   "M-001".."M-008" in dep graph    butterscotch       #FF9966
  -   Dependency descriptions          galaxy_gray        #52526A
  -   Mission item separator           dark_blue          #1B4F8F
  -   "▲ ▼ scroll" hint                galaxy_gray        #52526A
  -   Panel border chars ╭╮╰╯│─       galaxy_gray        #52526A
  -   Toolbar [a] key                  green_ok           #33FF33
  -   Toolbar [f] [s] [?] [Esc] keys   butterscotch       #FF9966
  -   Toolbar label text               light_gray         #CCCCCC
```

## Compact Variant (80x24 minimum)

When the terminal is narrower than 120 columns, the layout degrades to a stacked
single-column view. The manifest and analysis panels stack vertically with tabbed
navigation between Coverage Matrix and Dependency Graph.

```
╭──────────────────────────────────────────────────────────────────────────────╮
│ PLAN REVIEW ── USS Enterprise                                              │
│ Missions: 8  Waves: 3  Coverage: 92%  Sign-offs: 3/4                      │
├──────────────────────────────────────────────────────────────────────────────┤
│ ╭─ Mission Manifest ──────────────────────────────────────────────────────╮ │
│ │ M-001  Initialize warp core systems           [STANDARD_OPS]          │ │
│ │        Wave: 1 │ UC: 001,002 │ ACs: 4 │ Surface: core/warp           │ │
│ │ ───────────────────────────────────────────────────────────────────    │ │
│ │ M-002  Configure shield harmonics             [STANDARD_OPS]          │ │
│ │        Wave: 1 │ UC: 003 │ ACs: 3 │ Surface: defense/shields         │ │
│ │ ───────────────────────────────────────────────────────────────────    │ │
│ │ M-003  Establish sensor array baseline        [STANDARD_OPS]          │ │
│ │        Wave: 1 │ UC: 004,005 │ ACs: 5 │ Surface: sensors/array       │ │
│ │ ───────────────────────────────────────────────────────────────────    │ │
│ │ M-004  Calibrate weapons targeting systems    [ELEVATED_RISK]         │ │
│ │        Wave: 2 │ UC: 006,007 │ ACs: 6 │ Surface: weapons/phasers     │ │
│ │                                                          ▲ ▼ scroll   │ │
│ ╰─────────────────────────────────────────────────────────────────────── ╯ │
│ ╭─ [1] Coverage  [2] Dependencies ────────────────────────────────────── ╮ │
│ │ UC-001: Warp initialization          ✓  M-001, M-007                  │ │
│ │ UC-002: Power distribution           ✓  M-001, M-007                  │ │
│ │ UC-003: Shield modulation            ✓  M-002, M-007                  │ │
│ │ ...                                                     ▲ ▼ scroll    │ │
│ ╰──────────────────────────────────────────────────────────────────────  ╯ │
├──────────────────────────────────────────────────────────────────────────────┤
│ [a] Approve  [f] Feedback  [s] Shelve  [?] Help  [Esc] Ready Room         │
╰──────────────────────────────────────────────────────────────────────────────╯
```

**Compact differences from standard:**
- Single column: manifest on top, analysis below (stacked, not side-by-side)
- Analysis panel uses tab selector: `[1] Coverage  [2] Dependencies`
- Tab keys `1` and `2` switch between Coverage Matrix and Dependency Graph
- Mission manifest gets ~10 rows, analysis panel gets ~5 rows
- Panel titles shortened
- Scroll indicators on both panels
- No borders on inner mission items (save horizontal space per `breakpoints.minimum`)
