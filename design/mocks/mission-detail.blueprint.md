# Mission Detail -- ASCII Blueprint

> Source: `mission-detail.mock.html` | Spec: `prompts/mission-detail.spec.json`
> Terminal: 120 cols x 30 rows (standard breakpoint)

## Layout Structure

The mission detail uses a three-row vertical stack:

```
MissionHeader (100%, 4 rows)
──────────────────────────────────────
ACPhaseDetail (40%) + GateEvidence (30%) + DemoTokenPreview (30%)
   (fill height, scrollable)
──────────────────────────────────────
NavigableToolbar (100%, 1 row)
```

- **Header**: 4 rows. Row 1 is mission ID and title in butterscotch bold, with classification badge (STANDARD_OPS in blue bg or RED_ALERT in red bg) and status badge. Row 2 is metadata inline (Wave, Revision, Agent, Ship).
- **Main content**: Fill remaining height. AC phase detail (40% width, left) shows scrollable AC list with 6-dot TDD phase pipelines. Gate evidence (30% width, center) shows a table of gate results sorted newest-first. Demo token preview (30% width, right) shows Glamour-rendered demo token content with validation badge.
- **Toolbar**: 1 row. Context-sensitive NavigableToolbar. Actions vary by mission state: in-progress shows Halt, halted shows Retry, review shows Approve/Reject, done shows Worktree.

## Component Mapping (from components.yaml)

| Visual Area | Component | Charm Base | Key Props |
|---|---|---|---|
| Full screen container | **AppShell** (#1) | `lipgloss.JoinVertical` | header, body, footer |
| Mission header | Custom `lipgloss.NewStyle()` | `lipgloss.NewStyle()` | mission_id, title, classification, status badge |
| Classification badge | Custom `lipgloss.NewStyle()` | `lipgloss.NewStyle()` | [RED_ALERT] red_alert bg, [STANDARD_OPS] blue bg |
| Status badge | **StatusBadge** (#6) | `lipgloss.NewStyle()` | in_progress/review/done/halted variants |
| 3-column content | **PanelGrid** (#2) | `lipgloss.JoinHorizontal` | ratios: `[0.4, 0.3, 0.3]` |
| AC phase detail | **FocusablePanel** (#4) wrapping `bubbles.viewport` | `bubbles.viewport` | scrollable AC list |
| TDD phase pipeline per AC | **PhaseIndicator** (#8) compact | `lipgloss.NewStyle()` | 6-dot per AC |
| Gate evidence table | **FocusablePanel** (#4) wrapping `bubbles.table` | `bubbles.table` | gate_type, classification, exit_code |
| Gate result rows | **GateResultRow** (#19) | `lipgloss.NewStyle()` | accept/reject/running/skipped |
| Demo token preview | **FocusablePanel** (#4) + Glamour | `bubbles.viewport` | Glamour-rendered content |
| Bottom toolbar | **NavigableToolbar** | `lipgloss.JoinHorizontal` | Context-sensitive by mission state |

## Token Reference

| Element                | Token(s)                                          | Hex         |
|------------------------|---------------------------------------------------|-------------|
| Mission title          | `butterscotch` (bold)                             | `#FF9966`   |
| Classification STANDARD| `blue` (bg), `space_white` (fg)                   | `#9999CC`   |
| Classification RED_ALERT| `red_alert` (bg), `black` (fg)                   | `#FF3333`   |
| Status IN PROGRESS     | `galaxy_gray` (bg), `butterscotch` (fg)           | `#52526A`   |
| Header labels          | `galaxy_gray`                                     | `#52526A`   |
| Header values          | `space_white`                                     | `#F5F6FA`   |
| Panel border default   | `lipgloss.RoundedBorder()`, `blue`                | `#9999CC`   |
| Panel border focused   | `lipgloss.RoundedBorder()`, `moonlit_violet`      | `#9966FF`   |
| Panel title            | `blue` (bold)                                     | `#9999CC`   |
| AC index               | `blue` (bold)                                     | `#9999CC`   |
| AC title               | `space_white`                                     | `#F5F6FA`   |
| Phase dot complete     | `green_ok`                                        | `#33FF33`   |
| Phase dot active       | `butterscotch` (pulsing)                          | `#FF9966`   |
| Phase dot pending      | `galaxy_gray`                                     | `#52526A`   |
| Phase dot failed       | `red_alert`                                       | `#FF3333`   |
| Gate time              | `light_gray`                                      | `#CCCCCC`   |
| Gate type              | `blue`                                            | `#9999CC`   |
| Gate AC ref            | `space_white`                                     | `#F5F6FA`   |
| Gate exit pass         | `green_ok`                                        | `#33FF33`   |
| Gate exit fail         | `red_alert`                                       | `#FF3333`   |
| Gate classification    | `light_gray`                                      | `#CCCCCC`   |
| Demo heading           | `butterscotch` (bold)                             | `#FF9966`   |
| Demo text              | `light_gray`                                      | `#CCCCCC`   |
| Demo badge VALID       | `green_ok` (fg), `dark_blue` (bg)                 | `#33FF33`   |
| Toolbar key            | `butterscotch` (bold)                             | `#FF9966`   |
| Toolbar label          | `space_white`                                     | `#F5F6FA`   |
| Toolbar separator      | `galaxy_gray`                                     | `#52526A`   |

## ASCII Blueprint (120x30 standard)

```
╭──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╮
│  M-005: Implement User Authentication                                    STANDARD_OPS   IN PROGRESS                │
│  Wave: Wave 2   Rev: 1   Agent: Cmdr. Data   Ship: USS Enterprise                                                 │
╰──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╯
╭─ Acceptance Criteria & TDD Pipeline ───────────╮╭─ Gate Evidence ──────────────────╮╭─ Demo Token Preview ────────────╮
│                                                ││ Time  Gate     AC   Exit  Class  ││ ## Auth Module                  │
│ AC-1  User can log in with email and password  ││───────────────────────────────── ││                                 │
│       ✓ ── ✓ ── ✓ ── ✓ ── ✓ ── ✓              ││ 14:23 V_GREEN  AC-3  ✓  STANDARD ││ Users can now log in with       │
│       R    VR   G    VG   RF   VRF             ││ 14:19 GREEN   AC-3  ✓  STANDARD ││ email and password credentials. │
│                                                ││ 14:15 V_RED   AC-3  ✓  STANDARD ││ The system validates input,     │
│ AC-2  Invalid credentials show error message   ││ 14:11 RED     AC-3  ✗  STANDARD ││ manages secure sessions, and    │
│       ✓ ── ✓ ── ✓ ── ✓ ── ✓ ── ✓              ││ 14:08 V_REFACT AC-2  ✓  STANDARD ││ provides appropriate error      │
│       R    VR   G    VG   RF   VRF             ││ 14:04 REFACTOR AC-2  ✓  STANDARD ││ feedback for invalid attempts.  │
│                                                ││ 13:58 V_GREEN AC-2  ✓  STANDARD ││                                 │
│ AC-3  Session token is stored securely         ││ 13:54 GREEN   AC-2  ✓  STANDARD ││ Session tokens are stored using │
│       ✓ ── ✓ ── ✓ ── ✓ ── ▸ ── ●              ││                                  ││ industry-standard encryption    │
│       R    VR   G    VG   RF   VRF             ││                                  ││ and automatically expire after  │
│                                                ││                                  ││ 24 hours of inactivity.         │
│ AC-4  User can log out and clear session       ││                                  ││                                 │
│       ● ── ● ── ● ── ● ── ● ── ●              ││                                  ││ Password complexity requires:   │
│       R    VR   G    VG   RF   VRF             ││                                  ││ • Minimum 8 characters          │
│                                                ││                                  ││ • At least one uppercase letter │
│ AC-5  Password must meet complexity reqs       ││                                  ││ • At least one number           │
│       ● ── ● ── ● ── ● ── ● ── ●              ││                                  ││ • At least one special character│
│       R    VR   G    VG   RF   VRF             ││                                  ││                                 │
│                                                ││                                  ││ ✓ VALID                         │
╰────────────────────────────────────────────────╯╰──────────────────────────────────╯╰─────────────────────────────────╯
╭──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╮
│ [h] Halt  ●  [?] Help  ●  [Esc] Bridge                                                                            │
╰──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╯
```

### Annotation Notes

- The **focused panel** uses `moonlit_violet` (`#9966FF`) for its border `╭╮╰╯`. The default focused panel on entry is the AC Phase Detail panel. All other panels use `blue` (`#9999CC`) borders.
- The **mission title** (`M-005: Implement User Authentication`) renders in `butterscotch` (`#FF9966`) bold.
- The **STANDARD_OPS** classification badge renders with `dark_blue` (`#1B4F8F`) background and `blue` (`#9999CC`) foreground. A `RED_ALERT` classification would use `red_alert` (`#FF3333`) background with `black` foreground.
- The **IN PROGRESS** status badge renders with `galaxy_gray` (`#52526A`) background and `butterscotch` (`#FF9966`) foreground.
- Phase pipeline dots use Unicode icons from tokens.yaml: `✓` for complete phases (green_ok), `▸` for active phase (butterscotch, pulsing via `alert_pulse` spring), `●` for pending phases (galaxy_gray), `✗` for failed phases (red_alert).
- Phase labels (R, VR, G, VG, RF, VRF) render in `galaxy_gray` below the dots as contextual reference for RED, VERIFY_RED, GREEN, VERIFY_GREEN, REFACTOR, VERIFY_REFACTOR.
- Gate evidence rows with pass (`✓`) use a faint `green_ok` row tint. Rows with fail (`✗`) use a faint `red_alert` row tint.
- The gate evidence table header separator uses `galaxy_gray` (`#52526A`).
- Demo token content renders via Glamour markdown. The `## Auth Module` heading uses `butterscotch` bold. Body text uses `light_gray`.
- The `✓ VALID` demo token badge uses `green_ok` foreground on `dark_blue` background.
- Toolbar `●` separators use `galaxy_gray`. Shortcut keys like `[h]` use `butterscotch`. Labels use `space_white`.
- The toolbar is context-sensitive: in-progress shows `[h] Halt`, halted shows `[r] Retry`, review shows `[a] Approve [f] Reject`, done shows `[w] Worktree`. All variants include `[?] Help` and `[Esc] Bridge`.

## Color Annotations

| Region                      | Foreground Token    | Background Token   | Style       |
|-----------------------------|---------------------|--------------------|-------------|
| Mission title "M-005:..."   | `butterscotch`      | --                 | Bold        |
| Classification STANDARD_OPS | `blue`              | `dark_blue`        | Bold        |
| Classification RED_ALERT    | `black`             | `red_alert`        | Bold        |
| Status IN PROGRESS          | `butterscotch`      | `galaxy_gray`      | Bold        |
| Header labels "Wave:" etc   | `galaxy_gray`       | --                 | --          |
| Header values               | `space_white`       | --                 | --          |
| Panel border (default)      | `blue`              | --                 | Rounded     |
| Panel border (focused)      | `moonlit_violet`    | --                 | Rounded     |
| Panel title                 | `blue`              | --                 | Bold        |
| AC index "AC-1"             | `blue`              | --                 | Bold        |
| AC title                    | `space_white`       | --                 | --          |
| Phase dot complete `✓`      | `green_ok`          | --                 | --          |
| Phase dot active `▸`        | `butterscotch`      | --                 | Pulsing     |
| Phase dot pending `●`       | `galaxy_gray`       | --                 | --          |
| Phase dot failed `✗`        | `red_alert`         | --                 | --          |
| Phase labels "R VR G..."    | `galaxy_gray`       | --                 | Faint       |
| Gate time "14:23"           | `light_gray`        | --                 | --          |
| Gate type "V_GREEN"         | `blue`              | --                 | --          |
| Gate AC ref "AC-3"          | `space_white`       | --                 | --          |
| Gate exit pass `✓`          | `green_ok`          | --                 | --          |
| Gate exit fail `✗`          | `red_alert`         | --                 | --          |
| Gate row (pass)             | --                  | `green_ok` (5%)    | --          |
| Gate row (fail)             | --                  | `red_alert` (5%)   | --          |
| Gate classification         | `light_gray`        | --                 | --          |
| Demo heading "## Auth..."   | `butterscotch`      | --                 | Bold        |
| Demo body text              | `light_gray`        | --                 | --          |
| Demo badge `✓ VALID`        | `green_ok`          | `dark_blue`        | Bold        |
| Toolbar keys `[h]`          | `butterscotch`      | --                 | Bold        |
| Toolbar labels              | `space_white`       | --                 | --          |
| Toolbar separators `●`      | `galaxy_gray`       | --                 | --          |

## Compact Variant (80x24 minimum)

In compact mode (`<120 cols`), the three-column content area stacks vertically. The AC list takes full width. Gate evidence and demo token collapse into a tabbed area below the AC list, switchable with Tab. Header metadata abbreviates.

```
╭──────────────────────────────────────────────────────────────────────────────────╮
│  M-005: Implement User Authentication        STANDARD_OPS   IN PROGRESS        │
│  Wave: 2  Rev: 1  Agent: Cmdr. Data  Ship: USS Enterprise                      │
╰──────────────────────────────────────────────────────────────────────────────────╯
╭─ AC & TDD Pipeline ────────────────────────────────────────────────────────────╮
│                                                                                │
│ AC-1  User can log in with email and password                                  │
│       ✓ ── ✓ ── ✓ ── ✓ ── ✓ ── ✓                                              │
│       R    VR   G    VG   RF   VRF                                             │
│                                                                                │
│ AC-2  Invalid credentials show error message                                   │
│       ✓ ── ✓ ── ✓ ── ✓ ── ✓ ── ✓                                              │
│       R    VR   G    VG   RF   VRF                                             │
│                                                                                │
│ AC-3  Session token is stored securely                                         │
│       ✓ ── ✓ ── ✓ ── ✓ ── ▸ ── ●                                              │
│       R    VR   G    VG   RF   VRF                                             │
│                                                                                │
│ AC-4  User can log out and clear session                                       │
│       ● ── ● ── ● ── ● ── ● ── ●                                              │
│       R    VR   G    VG   RF   VRF                                             │
│                                                                                │
╰────────────────────────────────────────────────────────────────────────────────╯
╭──────────────────────────────────────────────────────────────────────────────────╮
│ [h] Halt  ●  [?] Help  ●  [Esc] Bridge  ●  [Tab] Gates/Demo                   │
╰──────────────────────────────────────────────────────────────────────────────────╯
```

### Compact Variant Notes

- The three content panels (AC Phase Detail, Gate Evidence, Demo Token) cannot fit side-by-side at 80 columns. The AC list takes full width as the primary panel.
- Gate Evidence and Demo Token Preview are accessible via Tab cycling. A `[Tab] Gates/Demo` hint appears in the toolbar.
- Header collapses from 2 content rows to 2 abbreviated rows. "Wave: Wave 2" shortens to "Wave: 2".
- AC-5 is scrolled off-screen at 24 rows -- scrolling via Up/Down is required to see all ACs.
- Toolbar labels remain unabbreviated since the in-progress toolbar is compact enough to fit at 80 columns.
- The focused panel still uses `moonlit_violet` border for focus indication.
