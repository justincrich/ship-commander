# Specialist Detail -- ASCII Blueprint

> Source: `specialist-detail.mock.html` | Spec: `prompts/specialist-detail.spec.json`
> Terminal: 120 cols x 30 rows (standard breakpoint)

## Layout Structure

The specialist detail uses a four-row vertical stack:

```
AgentProfileHeader (100%, 4 rows)
──────────────────────────────────────
OutputViewport (100%, 70% fill height, scrollable)
──────────────────────────────────────
SignoffStatus (100%, 15% fill height)
──────────────────────────────────────
NavigableToolbar (100%, 1 row)
```

- **Agent Profile Header**: 4 rows. Row 1 is specialist name in bold butterscotch, role badge in purple bg, and ship assignment in blue. Row 2 is specialty, harness, and model metadata in light_gray labels with space_white values. Row 3 is status badge (e.g. PLANNING) in butterscotch, directive title, and iteration count. Row 4 is elapsed timer in almond.
- **Output Viewport**: Fill ~70% remaining height. Scrollable panel showing the specialist's real-time planning output. Includes section headers, bullet analysis, mission proposals with acceptance criteria, and questions for command. Blue rounded border. Line numbers in galaxy_gray on the left gutter.
- **Signoff Status**: ~15% height. Compact panel tracking approval status from crew members. Shows checkmark (green_ok) for approved, open circle (galaxy_gray) for pending. Blue rounded border.
- **Toolbar**: 1 row. Minimal NavigableToolbar with [?] Help and [Esc] Ready Room. This is a read-only monitoring view.

## Component Mapping (from components.yaml)

| Visual Area | Component | Charm Base | Key Props |
|---|---|---|---|
| Full screen container | **AppShell** (#1) | `lipgloss.JoinVertical` | header, body, footer |
| Agent profile header | **SpecialistCard** (#24) header variant | `lipgloss.NewStyle()` | name, role (purple badge), ship, harness, model |
| Status badge | **StatusBadge** (#6) | `lipgloss.NewStyle()` | working/done/waiting/skipped/failed |
| Elapsed timer | **ElapsedTimer** (#10) | `bubbles.stopwatch` | format: auto |
| Output viewport | **FocusablePanel** (#4) wrapping `bubbles.viewport` | `bubbles.viewport` | scrollable planning output, line numbers |
| Signoff status panel | **FocusablePanel** (#4) | `lipgloss.NewStyle().Border(lipgloss.RoundedBorder())` | ✓/○ per crew member |
| Bottom toolbar | **NavigableToolbar** | `lipgloss.JoinHorizontal` | [?] Help [Esc] Ready Room |

## Token Reference

| Element                | Token(s)                                          | Hex         |
|------------------------|---------------------------------------------------|-------------|
| Agent name             | `butterscotch` (bold)                             | `#FF9966`   |
| Role badge             | `purple` (bg), `black` (fg)                       | `#CC99CC`   |
| Ship name              | `blue`                                            | `#9999CC`   |
| Specialty/harness/model| `light_gray` (label), `space_white` (value)       | `#CCCCCC`   |
| Status badge PLANNING  | `butterscotch`                                    | `#FF9966`   |
| Status badge EXECUTING | `butterscotch`                                    | `#FF9966`   |
| Status badge DONE      | `green_ok`                                        | `#33FF33`   |
| Status badge WAITING   | `galaxy_gray`                                     | `#52526A`   |
| Status badge FAILED    | `red_alert`                                       | `#FF3333`   |
| Elapsed timer          | `almond`                                          | `#FFAA90`   |
| Separator `│`          | `galaxy_gray`                                     | `#52526A`   |
| Panel border default   | `lipgloss.RoundedBorder()`, `blue`                | `#9999CC`   |
| Panel border focused   | `lipgloss.RoundedBorder()`, `moonlit_violet`      | `#9966FF`   |
| Viewport title         | `butterscotch` (bold)                             | `#FF9966`   |
| Line numbers           | `galaxy_gray`                                     | `#52526A`   |
| Section headers        | `blue` (bold)                                     | `#9999CC`   |
| Bullet `●`             | `butterscotch`                                    | `#FF9966`   |
| Highlight text         | `gold` (bold)                                     | `#FFAA00`   |
| Arrow `→`              | `galaxy_gray`                                     | `#52526A`   |
| AC labels              | `galaxy_gray`                                     | `#52526A`   |
| Output text            | `space_white`                                     | `#F5F6FA`   |
| Signoff title          | `butterscotch` (bold)                             | `#FF9966`   |
| Signoff approved `✓`   | `green_ok`                                        | `#33FF33`   |
| Signoff pending `○`    | `galaxy_gray`                                     | `#52526A`   |
| Signoff name           | `space_white`                                     | `#F5F6FA`   |
| Signoff annotation     | `galaxy_gray`                                     | `#52526A`   |
| Toolbar key            | `butterscotch` (bold)                             | `#FF9966`   |
| Toolbar label          | `light_gray`                                      | `#CCCCCC`   |
| Toolbar separator      | `galaxy_gray`                                     | `#52526A`   |

## ASCII Blueprint (120x30 standard)

```
╭──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╮
│  Lt. Cmdr. La Forge  SPECIALIST  │ USS Enterprise                                                                  │
│  Specialty: Backend Architecture │ Harness: claude-code │ Model: sonnet-4.5                                        │
│  ▸ PLANNING │ Directive: Build terminal UI for Ready Room │ Iteration: 2                                           │
│  Elapsed: 08m 12s                                                                                                  │
╰──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╯
╭─ PLANNING OUTPUT ────────────────────────────────────────────────────────────────────────────────────────────────────╮
│   1  ═══ MISSION DECOMPOSITION ANALYSIS ═══                                                                        │
│   2                                                                                                                │
│   3  Directive: Build terminal UI for Ready Room (iteration 2)                                                     │
│   4  Status: Analyzing surface area and defining acceptance criteria                                               │
│   5                                                                                                                │
│   6  ─── Surface Area Analysis ───                                                                                 │
│   7                                                                                                                │
│   8  ● Core Components Required:                                                                                   │
│   9    → AgentProfileCard (4x components for Captain, Commander, Design Officer, Specialist)                       │
│  10    → DirectivePanel (displays current mission context)                                                         │
│  11    → IterationTracker (shows planning cycle progress)                                                          │
│  12    → NavigableToolbar (keyboard navigation and actions)                                                        │
│  13                                                                                                                │
│  14  ● Backend Integration Points:                                                                                 │
│  15    → Agent state management (status, harness, model, elapsed time)                                             │
│  16    → Real-time updates via observer pattern or event channels                                                  │
│  17    → Keyboard event handling and focus management                                                              │
│  18    → Navigation state transitions (drill-down to specialist detail)                                            │
│  19                                                                                                                │
│  20  ─── Proposed Mission Breakdown ───                                                                            │
│  21                                                                                                                │
│  22  Mission 1: Implement AgentProfileCard Component                                                               │
│  23    AC1: Component renders agent name in butterscotch color with bold weight                                    │
│  24    AC2: Role badge displays with correct color mapping (Captain=gold, Commander=blue, etc.)                    │
╰──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╯
╭─ SIGN-OFF STATUS ────────────────────────────────────────────────────────────────────────────────────────────────────╮
│  ✓ Capt. Picard (approved)   ✓ Cmdr. Riker (approved)   ○ Lt. Troi (pending)   ○ Ens. Crusher (pending)           │
╰──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╯
╭──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╮
│ [?] Help  ●  [Esc] Ready Room                                                                                      │
╰──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╯
```

### Annotation Notes

- The **Agent Profile Header** panel uses `galaxy_gray` (`#52526A`) rounded border. The specialist name `Lt. Cmdr. La Forge` renders in `butterscotch` (`#FF9966`) bold. The `SPECIALIST` role badge renders with `purple` (`#CC99CC`) background and `black` foreground.
- The **Output Viewport** panel uses `blue` (`#9999CC`) rounded border by default. When focused via Tab cycling, it switches to `moonlit_violet` (`#9966FF`) border.
- Line numbers in the output viewport are right-aligned in `galaxy_gray` (`#52526A`). A 2-character gutter separates line numbers from content.
- Section headers like `═══ MISSION DECOMPOSITION ANALYSIS ═══` render in `blue` (`#9999CC`) bold.
- Highlight labels like `Mission 1:` and `Directive:` render in `gold` (`#FFAA00`) bold.
- Bullet `●` characters render in `butterscotch` (`#FF9966`). Arrow `→` characters render in `galaxy_gray`.
- AC labels (`AC1:`, `AC2:`, etc.) render in `galaxy_gray` with content in `space_white`.
- The **Signoff Panel** uses `blue` (`#9999CC`) rounded border. Approved items show `✓` in `green_ok` (`#33FF33`). Pending items show `○` in `galaxy_gray` (`#52526A`). Names render in `space_white`, annotations like `(approved)` in `galaxy_gray`.
- The **status badge** `▸ PLANNING` renders in `butterscotch` (`#FF9966`). The `▸` icon is the `running` icon from tokens.yaml.
- The **elapsed timer** `08m 12s` renders in `almond` (`#FFAA90`). After 15 minutes, the color changes to `yellow_caution` (`#FFCC00`) as a warning.
- Inline separators `│` between metadata fields render in `galaxy_gray` (`#52526A`).
- Toolbar `●` separator uses `galaxy_gray`. Shortcut keys like `[?]` use `butterscotch`. Labels use `light_gray`.
- The output viewport supports auto-scroll when new output arrives. Manual scroll (Up/Down/PageUp/PageDown) disables auto-scroll until the user scrolls back to the bottom.

## Color Annotations

| Region                      | Foreground Token    | Background Token   | Style       |
|-----------------------------|---------------------|--------------------|-------------|
| Agent name                  | `butterscotch`      | --                 | Bold        |
| Role badge `SPECIALIST`     | `black`             | `purple`           | Bold        |
| Ship name `USS Enterprise`  | `blue`              | --                 | --          |
| Metadata labels             | `galaxy_gray`       | --                 | --          |
| Metadata values             | `light_gray`        | --                 | --          |
| Status `▸ PLANNING`         | `butterscotch`      | --                 | --          |
| Directive text              | `light_gray`        | --                 | --          |
| Elapsed timer               | `almond`            | --                 | --          |
| Elapsed timer (warning)     | `yellow_caution`    | --                 | --          |
| Separator `│`               | `galaxy_gray`       | --                 | --          |
| Panel border (default)      | `blue`              | --                 | Rounded     |
| Panel border (focused)      | `moonlit_violet`    | --                 | Rounded     |
| Viewport title              | `butterscotch`      | --                 | Bold        |
| Line numbers                | `galaxy_gray`       | --                 | --          |
| Section headers             | `blue`              | --                 | Bold        |
| Bullet `●`                  | `butterscotch`      | --                 | --          |
| Highlight labels            | `gold`              | --                 | Bold        |
| Arrow `→`                   | `galaxy_gray`       | --                 | --          |
| AC labels                   | `galaxy_gray`       | --                 | --          |
| Output text                 | `space_white`       | --                 | --          |
| Signoff title               | `butterscotch`      | --                 | Bold        |
| Signoff `✓`                 | `green_ok`          | --                 | --          |
| Signoff `○`                 | `galaxy_gray`       | --                 | --          |
| Signoff names               | `space_white`       | --                 | --          |
| Signoff annotations         | `galaxy_gray`       | --                 | --          |
| Toolbar shortcut keys       | `butterscotch`      | --                 | Bold        |
| Toolbar labels              | `light_gray`        | --                 | --          |
| Toolbar separators `●`      | `galaxy_gray`       | --                 | --          |

## Compact Variant (80x24 minimum)

In compact mode (`<120 cols`), the layout compresses. The agent profile header collapses to 2 lines. The signoff panel hides behind a toggle. The output viewport takes maximum available height.

```
╭──────────────────────────────────────────────────────────────────────────────────╮
│  Lt. Cmdr. La Forge  SPECIALIST  │ USS Enterprise  ▸ PLANNING                  │
│  Backend Architecture │ claude-code │ sonnet-4.5 │ Iter: 2 │ 08m 12s          │
╰──────────────────────────────────────────────────────────────────────────────────╯
╭─ PLANNING OUTPUT ────────────────────────────────────────────────────────────────╮
│   1  ═══ MISSION DECOMPOSITION ANALYSIS ═══                                     │
│   2                                                                             │
│   3  Directive: Build terminal UI for Ready Room (iteration 2)                  │
│   4  Status: Analyzing surface area and defining acceptance criteria            │
│   5                                                                             │
│   6  ─── Surface Area Analysis ───                                              │
│   7                                                                             │
│   8  ● Core Components Required:                                                │
│   9    → AgentProfileCard (4x components)                                       │
│  10    → DirectivePanel (displays mission context)                              │
│  11    → IterationTracker (planning cycle progress)                             │
│  12    → NavigableToolbar (keyboard nav and actions)                            │
│  13                                                                             │
│  14  ● Backend Integration Points:                                              │
│  15    → Agent state management                                                 │
│  16    → Real-time updates via observer pattern                                 │
│  17    → Keyboard event handling and focus mgmt                                 │
│  18    → Navigation state transitions                                           │
╰──────────────────────────────────────────────────────────────────────────────────╯
╭──────────────────────────────────────────────────────────────────────────────────╮
│ [?] Help  ● [s] Signoffs  ● [Esc] Back                                         │
╰──────────────────────────────────────────────────────────────────────────────────╯
```

### Compact Variant Notes

- Agent profile header collapses from 4 rows to 2. Row 1 combines name, role badge, ship, and status. Row 2 combines specialty, harness, model, iteration, and elapsed time with abbreviated labels.
- Signoff panel is hidden by default to maximize output viewport space. Press `[s]` to toggle signoff panel visibility.
- Toolbar adds `[s] Signoffs` action and abbreviates `[Esc] Ready Room` to `[Esc] Back`.
- Output viewport text may be truncated at the right edge for long lines. Horizontal scroll is not supported; content wraps or truncates.
- Line numbers and gutter are preserved in compact mode for output readability.
- The focused panel still uses `moonlit_violet` border for focus indication.
