# Ship Bridge -- ASCII Blueprint

> Source: `ship-bridge.mock.html` | Spec: `prompts/ship-bridge.spec.json`
> Terminal: 120 cols x 30 rows (standard breakpoint)

## Layout Structure

ShipHeaderBar (100%, 3 rows) | CrewPanel (40%) + MissionBoard (60%) | EventLog (100%, fill) | NavigableToolbar (100%, 1 row)

- **ShipHeaderBar**: Row 1 = ship name (bold butterscotch) + class (faint) + directive (blue) + status badge. Row 2 = health dots, crew count, mission progress, wave progress bar, question badge.
- **CrewPanel**: AgentCards with name, role badge, mission assignment, TDD phase, elapsed timer. Selected card uses moonlit_violet border. Focused panel uses moonlit_violet border.
- **MissionBoard**: Kanban columns B(acklog) IP(in-progress) R(eview) D(one) H(alted) with color-coded counts and mission cards.
- **EventLog**: Timestamped, severity-colored entries. Auto-scroll.
- **NavigableToolbar**: LAUNCHED state shortcuts with butterscotch keys, Esc in moonlit_violet.

## Component Mapping (from components.yaml)

| Visual Area | Component | Charm Base | Key Props |
|---|---|---|---|
| Full screen container | **AppShell** (#1) | `lipgloss.JoinVertical` | header, body, footer, width, height |
| Ship header row | **HeaderBar** (ship variant) | `lipgloss.NewStyle()` | ship name, class, directive, status badge, inline wave summary |
| Crew + Missions side-by-side | **PanelGrid** (#2) | `lipgloss.JoinHorizontal` | ratios: `[0.4, 0.6]`, compact_threshold: 120 |
| Crew panel | **FocusablePanel** (#4) | `lipgloss.NewStyle().Border(lipgloss.RoundedBorder())` | title: "Crew", focused: moonlit_violet |
| Agent list inside crew panel | **AgentGrid** (#12) | `bubbles.list` | filterable by role/status, keyboard nav |
| Individual agent row | **AgentCard** (#11) | `lipgloss.NewStyle()` | agent_id, role, mission_id, phase, elapsed, harness, status variants |
| TDD phase display | **PhaseIndicator** (#8) | `lipgloss.NewStyle()` | current_phase, phases_completed, compact mode |
| Elapsed timers | **ElapsedTimer** (#10) | `bubbles.stopwatch` | format: auto, warning_threshold: 5m |
| Mission board panel | **FocusablePanel** (#4) wrapping **MissionBoard** (#13) | `lipgloss.JoinHorizontal + bubbles.paginator` | 5 columns: B, IP, R, D, H |
| Mission cards in columns | **MissionCard** (#14) | `lipgloss.NewStyle()` | mission_id, wave, classification, phase, ac_progress |
| Status badges | **StatusBadge** (#6) | `lipgloss.NewStyle()` | running/done/stuck/halted/waiting variants |
| Event log panel | **FocusablePanel** (#4) wrapping **EventLog** (#15) | `bubbles.viewport` | max_lines, auto_scroll, severity_filter |
| Event log rows | **EventRow** (#16) | `lipgloss.NewStyle()` | severity (INFO/WARN/ERROR), timestamp, message |
| Health indicator | **HealthBar** (#7) | `lipgloss.NewStyle()` | 5-dot filled/empty, color by ok/warning/critical |
| Wave progress (inline header) | **WaveProgressBar** (#9) compact | `bubbles.progress` | inline "Wave K of L [bar]" |
| Bottom toolbar | **NavigableToolbar** | `lipgloss.JoinHorizontal` | Context-sensitive: DOCKED vs LAUNCHED shortcuts |

## Token Reference

| Token             | Hex       | Usage in this view                                      |
|-------------------|-----------|---------------------------------------------------------|
| butterscotch      | `#FF9966` | Ship name, panel titles, active badges, toolbar keys    |
| blue              | `#9999CC` | Directive text, panel borders (default), mission IDs    |
| gold              | `#FFAA00` | Captain role badge, question badge                      |
| green_ok          | `#33FF33` | Health dots, done status, completed missions            |
| red_alert         | `#FF3333` | RED phase indicator, halted column                      |
| yellow_caution    | `#FFCC00` | Stuck agent warning, caution events                     |
| purple            | `#CC99CC` | Review column, review status                            |
| moonlit_violet    | `#9966FF` | Focused panel border, selected card border, Esc key     |
| galaxy_gray       | `#52526A` | Inactive elements, unfilled progress, backlog IDs       |
| space_white       | `#F5F6FA` | Primary text                                            |
| light_gray        | `#CCCCCC` | Secondary text, labels, timestamps                      |
| black             | `#000000` | Terminal background                                     |

| Icon              | Char | Usage                                                    |
|-------------------|------|----------------------------------------------------------|
| running           | `▸`  | Active agent status                                      |
| done              | `✓`  | Completed agent/mission                                  |
| alert             | `⚠`  | Stuck agent warning                                      |
| failed            | `✗`  | Halted/failed status                                     |
| progress_filled   | `█`  | Wave progress bar filled                                 |
| progress_empty    | `░`  | Wave progress bar empty                                  |
| working           | `●`  | Health indicator dot                                     |
| idle              | `⏸`  | Paused/idle agent                                        |

| Border            | Chars          | Usage                                                |
|-------------------|----------------|------------------------------------------------------|
| RoundedBorder     | `╭ ╮ ╰ ╯ │ ─` | All panels (lipgloss.RoundedBorder)                  |
| Focused border    | moonlit_violet | Active/selected panel (crew panel in this mock)      |
| Default border    | blue           | Unfocused panels                                     |

## ASCII Blueprint (120x30 standard)

```
╭──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╮
│ USS Enterprise  Galaxy-class   Directive: Implement authentication system                            ┃ LAUNCHED ┃ │
│ Health: ●●●●●   Crew: 4/4   Missions: 7/12   Wave 2 of 3 ███████░░░                                      [?]    │
╰──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╯
╭─ Crew (4) ──────────────────────────────────────────╮╭─ Mission Board ──────────────────────────────────────────────╮
│ ╭────────────────────────────────────────────────╮  ││ B:2          IP:3          R:1          D:1          H:0    │
│ │ ▸ Riker              ┃ CAPTAIN ┃              │  ││─────────────────────────────────────────────────────────────│
│ │   Mission: M-003 Auth middleware              │  ││ M-008        M-003         M-002        M-001               │
│ │   Phase: RED → REFACTOR                       │  ││ Session      Auth          Login        User                │
│ │   Elapsed: 04:23                              │  ││ store        middleware    endpoint     model               │
│ ╰────────────────────────────────────────────────╯  ││ Unassigned   Riker        Under        Troi ✓              │
│ ╭────────────────────────────────────────────────╮  ││ ○○○○○○       ●●●●○○       review       ●●●●●●              │
│ │ ▸ Data               ┃ COMMANDER ┃            │  ││              M-007        ●●●●●●                            │
│ │   Mission: M-007 JWT validation               │  ││ M-012        JWT                                            │
│ │   Phase: GREEN → VERIFY_GREEN                 │  ││ OAuth        validation                                     │
│ │   Elapsed: 02:15                              │  ││ provider     Data                                           │
│ ╰────────────────────────────────────────────────╯  ││ Unassigned   ●●●●●●                                        │
│ ╭────────────────────────────────────────────────╮  ││ ○○○○○○       M-005                                          │
│ │ ⚠ La Forge            ┃ ENSIGN ┃              │  ││              Password                                       │
│ │   Mission: M-005 Password hashing             │  ││              hashing                                        │
│ │   Phase: RED STUCK                            │  ││              La Forge ⚠                                     │
│ │   Elapsed: 12:47                              │  ││              ●○○○○○                                         │
│ ╰────────────────────────────────────────────────╯  ││                                                             │
│ ╭────────────────────────────────────────────────╮  ││                                                             │
│ │ ✓ Troi               ┃ ENSIGN ┃              │  ││                                                             │
│ │   Mission: M-001 User model                   │  ││                                                             │
│ │   Phase: COMPLETE                             │  ││                                                             │
│ │   Elapsed: 08:32                              │  ││                                                             │
│ ╰────────────────────────────────────────────────╯  ││                                                             │
╰─────────────────────────────────────────────────────╯╰─────────────────────────────────────────────────────────────╯
╭─ Event Log ────────────────────────────────────────────────────────────────────────────────────────────────────────╮
│ 14:37:22  ℹ  Data completed mission M-007 gate VERIFY_GREEN ✓                                                   │
│ 14:35:18  ⚠  La Forge stuck on M-005 phase RED (attempt 3/5)                                                    │
│ 14:32:05  ℹ  Riker advancing M-003 to phase REFACTOR                                                            │
│ 14:28:41  ✓  Troi completed mission M-001 ALL_GATES_GREEN                                                       │
╰────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╯
 [h] Halt    [r] Retry    [w] Wave    [d] Dock    [?] Help    [Esc] Fleet
```

## Color Annotations

```
HEADER BAR
  "USS Enterprise"          -> butterscotch (bold)
  "Galaxy-class"            -> light_gray (faint)
  "Directive: Implement..." -> blue
  "LAUNCHED" badge          -> bg:butterscotch, fg:black
  Health dots "●●●●●"       -> green_ok
  "Crew: 4/4"              -> light_gray label, space_white value
  "Missions: 7/12"         -> light_gray label, space_white value
  "Wave 2 of 3"            -> light_gray
  Progress "███████"        -> butterscotch (filled)
  Progress "░░░"            -> galaxy_gray (empty)
  "[?]" question badge      -> bg:gold, fg:black

CREW PANEL (focused)
  Panel border              -> moonlit_violet (RoundedBorder, focused state)
  Panel title "Crew (4)"    -> butterscotch (bold)
  Agent: Riker (selected)
    Card border             -> moonlit_violet
    "▸" status icon         -> green_ok
    "Riker" name            -> space_white (bold)
    "CAPTAIN" role badge    -> bg:gold, fg:black
    "M-003" mission ID      -> blue
    "RED" phase             -> red_alert
    "REFACTOR" phase        -> yellow_caution
    "04:23" elapsed         -> space_white
  Agent: Data
    Card border             -> galaxy_gray
    "▸" status icon         -> green_ok
    "COMMANDER" role badge  -> bg:blue, fg:black
    "GREEN" phase           -> green_ok
    "VERIFY_GREEN" phase    -> butterscotch
    "02:15" elapsed         -> space_white
  Agent: La Forge
    Card border             -> galaxy_gray
    "⚠" status icon         -> yellow_caution
    "ENSIGN" role badge     -> bg:butterscotch, fg:black
    "RED" phase             -> red_alert
    "STUCK" label           -> yellow_caution
    "12:47" elapsed         -> yellow_caution (warning threshold)
  Agent: Troi
    Card border             -> galaxy_gray
    "✓" status icon         -> green_ok
    "ENSIGN" role badge     -> bg:butterscotch, fg:black
    "M-001" mission ID      -> green_ok
    "COMPLETE" phase        -> green_ok
    "08:32" elapsed         -> galaxy_gray (stopped)

MISSION BOARD
  Panel border              -> blue (RoundedBorder, unfocused)
  Panel title               -> butterscotch (bold)
  Column headers:
    "B:2"                   -> galaxy_gray
    "IP:3"                  -> butterscotch
    "R:1"                   -> purple
    "D:1"                   -> green_ok
    "H:0"                   -> red_alert
  Mission cards:
    Backlog IDs             -> galaxy_gray
    In-Progress IDs         -> butterscotch
    Review IDs              -> purple
    Done IDs                -> green_ok
  Phase dots:
    Filled (pass)           -> green_ok "●"
    Filled (red)            -> red_alert "●"
    Filled (active)         -> butterscotch "●"
    Empty                   -> galaxy_gray "○"

EVENT LOG
  Panel border              -> blue (RoundedBorder, unfocused)
  Panel title               -> butterscotch (bold)
  Timestamps                -> light_gray
  Info icon "ℹ"             -> blue
  Warning icon "⚠"          -> yellow_caution
  Done icon "✓"             -> green_ok
  Info messages             -> space_white
  Warning messages          -> yellow_caution
  Success messages          -> green_ok

TOOLBAR
  Key badges "[h]" etc.     -> bg:butterscotch, fg:black
  Key badge "[Esc]"         -> bg:moonlit_violet, fg:black
  Labels                    -> light_gray
```

## Compact Variant (80x24 minimum)

```
╭──────────────────────────────────────────────────────────────────────────────╮
│ USS Enterprise  Galaxy-class                                 ┃ LAUNCHED ┃  │
│ Health: ●●●●●  Crew: 4/4  Missions: 7/12  W2/3 ███████░░░         [?]    │
╰──────────────────────────────────────────────────────────────────────────────╯
╭─ Crew (4) ──────────────────────────────────────────────────────────────────╮
│ ▸ Riker    ┃CAPTAIN┃   M-003 Auth middleware   RED→REFACTOR   04:23       │
│ ▸ Data     ┃COMMNDR┃   M-007 JWT validation    GRN→VERIFY     02:15       │
│ ⚠ La Forge ┃ENSIGN ┃   M-005 Password hashing  RED STUCK      12:47       │
│ ✓ Troi     ┃ENSIGN ┃   M-001 User model        COMPLETE       08:32       │
╰──────────────────────────────────────────────────────────────────────────────╯
╭─ Mission Board ─────────────────────────────────────────────────────────────╮
│ B:2          IP:3          R:1          D:1          H:0                    │
│──────────────────────────────────────────────────────────────────────────── │
│ M-008 Session store      M-003 Auth middleware      M-002 Login endpoint   │
│ M-012 OAuth provider     M-007 JWT validation       M-001 User model       │
│                          M-005 Password hashing ⚠                          │
╰──────────────────────────────────────────────────────────────────────────────╯
╭─ Event Log ─────────────────────────────────────────────────────────────────╮
│ 14:37:22  ℹ  Data completed M-007 gate VERIFY_GREEN ✓                      │
│ 14:35:18  ⚠  La Forge stuck on M-005 phase RED (attempt 3/5)               │
│ 14:32:05  ℹ  Riker advancing M-003 to phase REFACTOR                      │
│ 14:28:41  ✓  Troi completed M-001 ALL_GATES_GREEN                          │
╰──────────────────────────────────────────────────────────────────────────────╯
 [h] Halt  [r] Retry  [w] Wave  [d] Dock  [?] Help  [Esc] Fleet
```
