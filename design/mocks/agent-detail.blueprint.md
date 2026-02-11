# Agent Detail -- ASCII Blueprint

> Source: `agent-detail.mock.html` | Spec: `prompts/agent-detail.spec.json`
> Terminal: 120 cols x 30 rows (standard breakpoint)

## Layout Structure

The agent detail is a full-screen drill-down view with a four-row vertical stack:

```
AgentProfileHeader (100%, 4 rows)
──────────────────────────────────────
OutputViewport (100%, ~70% fill height, scrollable)
──────────────────────────────────────
ErrorContextPanel (100%, ~15%, conditional on stuck/dead)
──────────────────────────────────────
NavigableToolbar (100%, 1 row)
```

- **AgentProfileHeader**: 4 rows. Row 1 = agent name (bold butterscotch) + role badge (COMMANDER=blue bg) + ship assignment (light_gray) + harness (blue) + model (blue) + mission ID (gold). Row 2 = TDD phase indicator (colored dots) + current phase label (yellow_caution) + elapsed timer + status badge (green_ok). Row 3 = mission prompt (faint galaxy_gray, italic, truncated).
- **OutputViewport**: Fill remaining height (~70%). Scrollable viewport of agent terminal output with line numbers (dim galaxy_gray), syntax-highlighted content. Auto-scroll with manual override. Shows gate execution output, test results, code reading, error traces.
- **ErrorContextPanel**: Conditional (~15%). Appears when agent is stuck or dead. Uses `lipgloss.ThickBorder()` in `red_alert`. Shows stuck/dead detection details: timestamp, last gate, timeout, last output.
- **NavigableToolbar**: 1 row. Action buttons: [h] Halt, [r] Retry, [i] Ignore, [?] Help, [Esc] Back. Arrow keys to highlight, Enter to activate, or press shortcut key directly.

## Component Mapping (from components.yaml)

| Visual Area | Component | Charm Base | Key Props |
|---|---|---|---|
| Full screen container | **AppShell** (#1) | `lipgloss.JoinVertical` | header, body, footer |
| Agent profile header | Custom `lipgloss.NewStyle()` | `lipgloss.NewStyle()` | name, role badge, ship, harness, model, mission_id |
| Role badge | **StatusBadge** (#6) variant | `lipgloss.NewStyle()` | COMMANDER=blue, CAPTAIN=gold, ENSIGN=butterscotch |
| TDD phase indicator | **PhaseIndicator** (#8) | `lipgloss.NewStyle()` | 6-phase: R→VR→G→VG→RF→VRF, compact mode |
| Elapsed timer | **ElapsedTimer** (#10) | `bubbles.stopwatch` | format: auto, warning_threshold: 5m |
| Status badge | **StatusBadge** (#6) | `lipgloss.NewStyle()` | running/done/stuck/dead variants |
| Output viewport | **FocusablePanel** (#4) wrapping `bubbles.viewport` | `bubbles.viewport` | scrollable output, auto-scroll, line numbers |
| Error context panel | **Panel** (#3) alert variant | `lipgloss.NewStyle().Border(lipgloss.ThickBorder())` | red_alert border, stuck/dead details |
| Bottom toolbar | **NavigableToolbar** | `lipgloss.JoinHorizontal` | [h] Halt [r] Retry [i] Ignore [?] Help [Esc] Back |

## Token Reference

| Token             | Hex       | Usage in this view                                       |
|-------------------|-----------|----------------------------------------------------------|
| butterscotch      | `#FF9966` | Agent name, toolbar keys, gate title                     |
| blue              | `#9999CC` | Panel border (default), harness/model, info log lines    |
| gold              | `#FFAA00` | Mission ID                                               |
| green_ok          | `#33FF33` | Running status, passed tests, GREEN phase dot            |
| red_alert         | `#FF3333` | Error context border, failed tests, RED phase dot        |
| yellow_caution    | `#FFCC00` | Stuck detection, warnings, VERIFY phase label            |
| moonlit_violet    | `#9966FF` | Keywords in output, toolbar label text                   |
| galaxy_gray       | `#52526A` | Line numbers, inactive phase dots, faint text, border    |
| space_white       | `#F5F6FA` | Primary text in output viewport                          |
| light_gray        | `#CCCCCC` | Ship info, elapsed timer, secondary text                 |
| black             | `#000000` | Terminal background, role badge foreground                |

| Icon              | Char | Usage                                                    |
|-------------------|------|----------------------------------------------------------|
| running           | `▸`  | Active agent status badge                                |
| done              | `✓`  | Passed test output                                       |
| alert             | `⚠`  | Stuck detection header                                   |
| failed            | `✗`  | Failed test output                                       |
| working           | `●`  | TDD phase dots (colored by phase)                        |

| Border            | Chars          | Usage                                                |
|-------------------|----------------|------------------------------------------------------|
| RoundedBorder     | `╭ ╮ ╰ ╯ │ ─` | Profile header, output viewport (lipgloss.Rounded)   |
| ThickBorder       | `┏ ┓ ┗ ┛ ┃ ━` | Error context panel (lipgloss.ThickBorder)           |

## ASCII Blueprint (120x30 standard)

```
╭──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╮
│ Cmdr. Data    ┃ COMMANDER ┃   USS Enterprise   claude-code   sonnet-4   M-005                                     │
│ TDD: ● ● ● ● ● ●  VERIFY_GREEN                                                       ⏱ 12m 34s    ▸ RUNNING    │
│ Mission: "Implement user authentication with JWT tokens and rate limiting..."                                     │
╰──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╯
╭──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╮
│    1  ╭─────────────────────────────────────────────────────╮                                                      │
│    2  │ VERIFY_GREEN Gate Execution                         │                                                      │
│    3  ╰─────────────────────────────────────────────────────╯                                                      │
│    4                                                                                                               │
│    5  [14:27:48] Running test suite...                                                                             │
│    6                                                                                                               │
│    7  $ npm test -- auth.test.js                                                                                   │
│    8                                                                                                               │
│    9    ✓ should generate valid JWT token                                                                          │
│   10    ✓ should validate token signature                                                                          │
│   11    ✗ should enforce rate limiting after 100 requests                                                          │
│   12                                                                                                               │
│   13    AssertionError: expected 3 passing, got 2                                                                  │
│   14        at Context.<anonymous> (test/auth.test.js:45:12)                                                       │
│   15        at processImmediate (node:internal/timers:478:21)                                                      │
│   16                                                                                                               │
│   17    2 passing (1.2s)                                                                                           │
│   18    1 failing                                                                                                  │
│   19                                                                                                               │
│   20  [14:28:05] Test suite failed with exit code 1                                                                │
│   21                                                                                                               │
│   22  [14:28:05] Analyzing test failure...                                                                         │
│   23                                                                                                               │
│   24  Reading file: src/middleware/rateLimit.js                                                                     │
│   25    1: const rateLimit = require('express-rate-limit');                                                         │
│   26    2:                                                                                                         │
│   27    3: module.exports = rateLimit({                                                                            │
│   28    4:   windowMs: 15 * 60 * 1000, // 15 minutes                                                              │
│   29    5:   max: 50, // <- Issue: Should be 100                                                                   │
│   30    6: });                                                                                                     │
│   31                                                                                                               │
│   32  [14:28:12] Timeout exceeded (180s). Stuck detection triggered.                                               │
╰──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╯
┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃ ⚠ STUCK DETECTED                                                                                                 ┃
┃ Timestamp:       14:28:05                                                                                         ┃
┃ Last gate:       VERIFY_GREEN failed (exit 1)                                                                     ┃
┃ Timeout:         180s exceeded                                                                                    ┃
┃ Last output:     "Error: expected 3 assertions, got 2"                                                            ┃
┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
 [h] Halt    [r] Retry    [i] Ignore    [?] Help    [Esc] Back
```

### Annotation Notes

- The **AgentProfileHeader** uses `lipgloss.RoundedBorder()` in `blue` (`#9999CC`) for the panel border.
- The agent name **"Cmdr. Data"** renders in `butterscotch` (`#FF9966`) bold.
- The **role badge** `┃ COMMANDER ┃` uses `blue` (`#9999CC`) background with `black` foreground text.
- TDD **phase dots** `●` are individually colored to show progress: red_alert, yellow_caution, green_ok, yellow_caution, galaxy_gray (inactive), galaxy_gray (inactive).
- The **current phase label** "VERIFY_GREEN" renders in `yellow_caution` (`#FFCC00`).
- The **elapsed timer** `12m 34s` renders in `light_gray` (`#CCCCCC`).
- The **status badge** `▸ RUNNING` renders in `green_ok` (`#33FF33`).
- The **mission prompt** (row 3) renders in `galaxy_gray` (`#52526A`) italic/faint.
- The **OutputViewport** uses `lipgloss.RoundedBorder()` in `galaxy_gray` (`#52526A`) since it is not the focused panel in this state.
- **Line numbers** (left column) render in `galaxy_gray` (`#52526A`).
- The inner gate header box uses dim `galaxy_gray` border characters. The gate title "VERIFY_GREEN Gate Execution" renders in `butterscotch` (`#FF9966`).
- Test pass lines `✓` render in `green_ok` (`#33FF33`). Test fail lines `✗` render in `red_alert` (`#FF3333`).
- Info log timestamps `[14:27:48]` render in `blue` (`#9999CC`). Warning timestamps `[14:28:05]` render in `yellow_caution` (`#FFCC00`).
- Stack trace lines render in `galaxy_gray` (`#52526A`) faint.
- File path `src/middleware/rateLimit.js` renders in `pink` (`#FF99CC`). Keywords `const`, `require`, `module.exports` render in `moonlit_violet` (`#9966FF`). Numbers render in `ice` (`#99CCFF`). Error comment `// <- Issue: Should be 100` renders in `red_alert` (`#FF3333`).
- The **ErrorContextPanel** uses `lipgloss.ThickBorder()` (`┏ ┓ ┗ ┛ ┃ ━`) in `red_alert` (`#FF3333`). This panel only appears when the agent is stuck or dead.
- Error header `⚠ STUCK DETECTED` renders in `red_alert` (`#FF3333`) bold.
- Error detail labels (Timestamp, Last gate, etc.) render in `butterscotch` (`#FF9966`). Values render in `light_gray` (`#CCCCCC`).
- The **NavigableToolbar** renders inline (no border). Shortcut keys `[h]` etc. use `butterscotch` (`#FF9966`) bold. Labels use `moonlit_violet` (`#9966FF`). `[Esc]` uses `moonlit_violet`.

## Color Annotations

```
AGENT PROFILE HEADER
  "Cmdr. Data"               -> butterscotch (bold)
  "COMMANDER" badge          -> bg:blue, fg:black
  "USS Enterprise"           -> light_gray
  "claude-code"              -> blue
  "sonnet-4"                 -> blue
  "M-005"                    -> gold
  TDD label "TDD:"          -> light_gray
  Phase dot ● (RED)          -> red_alert
  Phase dot ● (YELLOW)       -> yellow_caution
  Phase dot ● (GREEN)        -> green_ok
  Phase dot ● (YELLOW)       -> yellow_caution
  Phase dot ● (inactive)     -> galaxy_gray
  Phase dot ● (inactive)     -> galaxy_gray
  "VERIFY_GREEN" phase       -> yellow_caution
  "⏱ 12m 34s" timer          -> light_gray
  "▸ RUNNING" status         -> green_ok
  Mission prompt text        -> galaxy_gray (faint, italic)
  Panel border               -> blue (RoundedBorder)

OUTPUT VIEWPORT
  Panel border               -> galaxy_gray (RoundedBorder)
  Line numbers "1"-"32"      -> galaxy_gray
  Gate box border chars      -> galaxy_gray (dim)
  Gate title                 -> butterscotch
  Info timestamps            -> blue
  Warning timestamps         -> yellow_caution
  "$ npm test" command       -> space_white
  "✓" pass icon              -> green_ok
  Pass test descriptions     -> green_ok
  "✗" fail icon              -> red_alert
  Fail test descriptions     -> red_alert
  AssertionError message     -> red_alert
  Stack trace lines          -> galaxy_gray (faint)
  "2 passing" / "1 failing"  -> red_alert
  File path string           -> pink
  Keywords (const, require)  -> moonlit_violet
  Number literals            -> ice
  Error inline comment       -> red_alert
  Stuck detection warning    -> yellow_caution

ERROR CONTEXT PANEL (conditional: stuck/dead)
  Panel border               -> red_alert (ThickBorder)
  "⚠ STUCK DETECTED" header  -> red_alert (bold)
  Detail labels              -> butterscotch
  Detail values              -> light_gray

NAVIGABLE TOOLBAR
  Key badges "[h]" etc.      -> butterscotch (bold)
  Key badge "[Esc]"          -> moonlit_violet (bold)
  Labels                     -> moonlit_violet
```

## Compact Variant (80x24 minimum)

In compact mode (`<120 cols`), the profile header compresses to 2 rows. The output viewport shows the last 20 lines only. The error context panel is hidden behind a toggle (`[e]` to expand). Mission prompt is truncated or hidden.

```
╭──────────────────────────────────────────────────────────────────────────────────╮
│ Cmdr. Data  ┃COMMNDR┃  USS Enterprise  claude-code  sonnet-4  M-005            │
│ TDD: ●●●●●●  VERIFY_GREEN   ⏱ 12m 34s   ▸ RUNNING                            │
╰──────────────────────────────────────────────────────────────────────────────────╯
╭──────────────────────────────────────────────────────────────────────────────────╮
│    5  [14:27:48] Running test suite...                                          │
│    7  $ npm test -- auth.test.js                                                │
│    9    ✓ should generate valid JWT token                                       │
│   10    ✓ should validate token signature                                      │
│   11    ✗ should enforce rate limiting after 100 requests                       │
│   13    AssertionError: expected 3 passing, got 2                               │
│   14        at Context.<anonymous> (test/auth.test.js:45:12)                    │
│   17    2 passing (1.2s)                                                        │
│   18    1 failing                                                               │
│   20  [14:28:05] Test suite failed with exit code 1                             │
│   22  [14:28:05] Analyzing test failure...                                      │
│   24  Reading file: src/middleware/rateLimit.js                                  │
│   29    5:   max: 50, // <- Issue: Should be 100                                │
│   32  [14:28:12] Timeout exceeded (180s). Stuck detection triggered.            │
╰──────────────────────────────────────────────────────────────────────────────────╯
┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃ ⚠ STUCK DETECTED  Gate: VERIFY_GREEN failed  Timeout: 180s exceeded            ┃
┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
 [h] Halt  [r] Retry  [i] Ignore  [?] Help  [Esc] Back
```

### Compact Variant Notes

- Profile header compresses from 3 content rows to 2. Mission prompt row is hidden entirely.
- Role badge abbreviates to 6 chars: "COMMNDR" for Commander, "CAPTAIN" stays, "ENSIGN" stays.
- TDD phase dots render inline without spaces between them for horizontal savings.
- Output viewport shows last 20 lines only (basic mode). Blank lines are omitted to maximize visible content.
- Error context panel collapses to a single row showing the most critical details inline.
- Toolbar labels remain the same but spacing is tighter.
- The selected/focused panel still uses `moonlit_violet` border when applicable.
