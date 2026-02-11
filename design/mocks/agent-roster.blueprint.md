# Agent Roster -- ASCII Blueprint

> Source: `agent-roster.mock.html` | Spec: `views.yaml` (agent-roster section)
> Terminal: 120 cols x 30 rows (standard breakpoint)

## Layout Structure

The agent roster uses a three-row vertical stack:

```
RosterHeader (100%, 2 rows)
──────────────────────────────────────
RoleFilter (20%) + AgentList (40%) + AgentDetail (40%)
   (fill height, scrollable)
──────────────────────────────────────
NavigableToolbar (100%, 1 row)
```

- **Header**: 2 rows. Row 1 is the AGENT ROSTER title in bold blue. Row 2 is roster-wide metrics inline (total agents, active count, idle count, stuck count with pink badge).
- **Main content**: Fill remaining height. Role filter sidebar (20% width, left) shows role categories with counts. Agent list (40% width, center) shows scrollable agent rows with status, name, role, ship, and phase. Agent detail (40% width, right) shows full profile for the selected agent.
- **Toolbar**: 1 row. NavigableToolbar with shortcut buttons. Arrow Left/Right to highlight, Enter to activate, or press shortcut key directly.

## Component Mapping (from components.yaml)

| Visual Area | Component | Charm Base | Key Props |
|---|---|---|---|
| Full screen container | **AppShell** (#1) | `lipgloss.JoinVertical` | header, body, footer |
| Roster header | Custom `lipgloss.NewStyle()` | `lipgloss.NewStyle()` | "AGENT ROSTER" bold blue, role counts |
| 3-column layout | **PanelGrid** (#2) | `lipgloss.JoinHorizontal` | ratios: `[0.2, 0.4, 0.4]`, compact_threshold: 120 |
| Role filter sidebar | **FocusablePanel** (#4) wrapping `bubbles.list` | `bubbles.list` | All/Commanders/Captains/Ensigns/Unassigned |
| Agent list | **FocusablePanel** (#4) wrapping **AgentGrid** (#12) | `bubbles.list` | filterable, keyboard nav |
| Agent rows | **AgentCard** (#11) | `lipgloss.NewStyle()` | name, role badge, ship, status |
| Agent detail panel | **FocusablePanel** (#4) | `lipgloss.NewStyle().Border(lipgloss.RoundedBorder())` | full profile, backstory (Glamour) |
| Backstory viewport | `bubbles.viewport` + Glamour | `bubbles.viewport` | Glamour markdown |
| Status badges | **StatusBadge** (#6) | `lipgloss.NewStyle()` | idle/active/stuck variants |
| Bottom toolbar | **NavigableToolbar** | `lipgloss.JoinHorizontal` | [n] New [Enter] Edit [a] Assign [d] Detach [Del] Delete [/] Search [?] Help [Esc] Fleet |

## Token Reference

| Element               | Token(s)                                          | Hex         |
|-----------------------|---------------------------------------------------|-------------|
| Header title          | `blue` (bold)                                     | `#9999CC`   |
| Header metrics text   | `space_white`                                     | `#F5F6FA`   |
| Header metrics label  | `light_gray`                                      | `#CCCCCC`   |
| Stuck badge           | `pink` (bg), `black` (fg)                         | `#FF99CC`   |
| Panel border default  | `lipgloss.RoundedBorder()`, `galaxy_gray`         | `#52526A`   |
| Panel border focused  | `lipgloss.RoundedBorder()`, `moonlit_violet`      | `#9966FF`   |
| Agent name            | `butterscotch` (bold)                             | `#FF9966`   |
| Role captain          | `gold`                                            | `#FFAA00`   |
| Role commander        | `blue`                                            | `#9999CC`   |
| Role implementer      | `butterscotch`                                    | `#FF9966`   |
| Role reviewer         | `purple`                                          | `#CC99CC`   |
| Role design officer   | `pink`                                            | `#FF99CC`   |
| Status active `▸`     | `butterscotch`                                    | `#FF9966`   |
| Status idle `○`       | `light_gray`                                      | `#CCCCCC`   |
| Status stuck `⚠`      | `yellow_caution`                                  | `#FFCC00`   |
| Ship assignment       | `blue`                                            | `#9999CC`   |
| Unassigned text       | `galaxy_gray`                                     | `#52526A`   |
| Phase RED             | `red_alert`                                       | `#FF3333`   |
| Phase GREEN           | `green_ok`                                        | `#33FF33`   |
| Phase REVIEW          | `purple`                                          | `#CC99CC`   |
| Phase PLANNING        | `blue`                                            | `#9999CC`   |
| Phase IDLE            | `light_gray`                                      | `#CCCCCC`   |
| Filter selected       | `moonlit_violet`                                  | `#9966FF`   |
| Filter unselected     | `galaxy_gray`                                     | `#52526A`   |
| Detail section head   | `gold` (bold, uppercase)                          | `#FFAA00`   |
| Detail label          | `light_gray`                                      | `#CCCCCC`   |
| Detail value          | `space_white`                                     | `#F5F6FA`   |
| Model valid `✓`       | `green_ok`                                        | `#33FF33`   |
| Model invalid `✗`     | `red_alert`                                       | `#FF3333`   |
| Toolbar key           | `butterscotch` (bold)                             | `#FF9966`   |
| Toolbar label         | `space_white`                                     | `#F5F6FA`   |
| Toolbar separator     | `galaxy_gray`                                     | `#52526A`   |

## ASCII Blueprint (120x30 standard)

```
╭──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╮
│  AGENT ROSTER                                                                                                      │
│  Agents: 12   Active: 7   Idle: 3   Stuck: [2]   Commanders: 2 │ Captains: 2 │ Ensigns: 8                         │
╰──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╯
╭─ Roles ──────────────╮╭─ Agent List ──────────────────────────────────────╮╭─ Agent Detail ───────────────────────────╮
│                      ││ ●  AGENT          ROLE         SHIP       PHASE  ││ capt-alpha                               │
│  ▸ All          (12) ││──────────────────────────────────────────────────││ captain                                  │
│    Captains      (2) ││ ▸  capt-alpha     captain      Enterprise PLAN  ││─────────────────────────────────────────  │
│    Commanders    (2) ││ ▸  impl-bravo     implementer  Enterprise RED   ││ PROFILE                                  │
│    Implementers  (5) ││ ▸  impl-charlie   implementer  Enterprise GREEN ││ Role:    captain                         │
│    Reviewers     (2) ││ ▸  rev-delta      reviewer     Enterprise REVW  ││ Model:   claude-sonnet-4  ✓              │
│    Design        (1) ││ ▸  capt-echo      captain      Nautilus   PLAN  ││ Harness: claude                          │
│                      ││ ▸  impl-foxtrot   implementer  Nautilus   GREEN ││ Created: 2025-01-15                      │
│                      ││ ⚠  impl-golf      implementer  Nautilus   RED   ││                                          │
│                      ││ ▸  design-hotel   design off.  Discovery  PLAN  ││ ASSIGNMENT                               │
│                      ││ ○  impl-india     implementer  Discovery  IDLE  ││ Ship:      SS Enterprise                 │
│                      ││ ○  impl-juliet    implementer  Discovery  IDLE  ││ Mission:   MISSION-01                    │
│                      ││ ⚠  rev-kilo       reviewer     Discovery  REVW  ││ Phase:     PLANNING                      │
│                      ││ ○  cmdr-lima      commander    Discovery  IDLE  ││ Elapsed:   04:12                         │
│                      ││                                                  ││                                          │
│                      ││                                                  ││ SKILLS                                   │
│                      ││                                                  ││ /plan  /review  /delegate                │
│                      ││                                                  ││ /commission  /monitor                    │
│                      ││                                                  ││                                          │
│                      ││                                                  ││ MISSION PROMPT                           │
│                      ││                                                  ││ ┃ Lead the Enterprise crew through       │
│                      ││                                                  ││ ┃ mission planning and execution.        │
│                      ││                                                  ││ ┃ Coordinate specialists and ensure      │
│                      ││                                                  ││ ┃ quality gates pass before merge.       │
╰──────────────────────╯╰──────────────────────────────────────────────────╯╰──────────────────────────────────────────╯
╭──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╮
│ [n] New  ●  [Enter] Edit  ●  [a] Assign  ●  [d] Detach  ●  [Del] Delete  ●  [/] Search  ●  [?] Help  ●  [Esc] Flt│
╰──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╯
```

### Annotation Notes

- The **selected agent row** (capt-alpha) uses `moonlit_violet` (`#9966FF`) for its left border indicator. All other agent rows use default text colors.
- The **Roles** panel label (`Roles`) renders inline on the top border in `galaxy_gray`.
- The **Agent List** panel label (`Agent List`) renders inline on the top border in `galaxy_gray`.
- The **Agent Detail** panel label (`Agent Detail`) renders inline on the top border in `galaxy_gray`.
- The currently selected role filter (`All`) is shown with `▸` prefix in `moonlit_violet`. Other role categories use `galaxy_gray`.
- Status indicators: `▸` (active, `butterscotch`), `○` (idle, `light_gray`), `⚠` (stuck, `yellow_caution`).
- Phase abbreviations are color-coded: PLAN (`blue`), RED (`red_alert`), GREEN (`green_ok`), REVW (`purple`), IDLE (`light_gray`).
- The `[2]` stuck badge renders with `pink` (`#FF99CC`) background and `black` foreground.
- The model validation checkmark `✓` uses `green_ok` (`#33FF33`). Invalid models show `✗` in `red_alert` (`#FF3333`).
- The detail mission prompt uses `┃` left-border in `blue` for blockquote style.
- Toolbar `●` separators use `galaxy_gray`. Shortcut keys like `[n]` use `butterscotch`. Labels use `space_white`.

## Color Annotations

| Region                      | Foreground Token    | Background Token   | Style       |
|-----------------------------|---------------------|--------------------|-------------|
| "AGENT ROSTER"              | `blue`              | --                 | Bold        |
| Header metric labels        | `light_gray`        | --                 | --          |
| Header metric values        | `space_white`       | --                 | Bold        |
| Stuck `[2]`                 | `black`             | `pink`             | Bold        |
| Role count separator `│`    | `galaxy_gray`       | --                 | --          |
| Panel border (default)      | `galaxy_gray`       | --                 | Rounded     |
| Panel border (focused)      | `moonlit_violet`    | --                 | Rounded     |
| Filter selected `▸`         | `moonlit_violet`    | --                 | Bold        |
| Filter unselected           | `galaxy_gray`       | --                 | --          |
| Filter count `(N)`          | `light_gray`        | --                 | Faint       |
| Agent name (selected)       | `butterscotch`      | --                 | Bold        |
| Agent name (unselected)     | `space_white`       | --                 | --          |
| Role captain                | `gold`              | --                 | --          |
| Role commander              | `blue`              | --                 | --          |
| Role implementer            | `butterscotch`      | --                 | --          |
| Role reviewer               | `purple`            | --                 | --          |
| Role design officer         | `pink`              | --                 | --          |
| Status `▸` (active)         | `butterscotch`      | --                 | --          |
| Status `○` (idle)           | `light_gray`        | --                 | --          |
| Status `⚠` (stuck)          | `yellow_caution`    | --                 | --          |
| Ship name                   | `blue`              | --                 | --          |
| Phase PLAN                  | `blue`              | --                 | --          |
| Phase RED                   | `red_alert`         | --                 | Bold        |
| Phase GREEN                 | `green_ok`          | --                 | Bold        |
| Phase REVW                  | `purple`            | --                 | --          |
| Phase IDLE                  | `light_gray`        | --                 | Faint       |
| Column header row           | `butterscotch`      | --                 | Bold        |
| Detail agent name           | `butterscotch`      | --                 | Bold        |
| Detail section titles       | `gold`              | --                 | Bold, Upper |
| Detail labels               | `light_gray`        | --                 | --          |
| Detail values               | `space_white`       | --                 | --          |
| Detail role badge           | `gold`              | --                 | --          |
| Model valid `✓`             | `green_ok`          | --                 | --          |
| Model invalid `✗`           | `red_alert`         | --                 | --          |
| Mission prompt `┃`          | `blue`              | --                 | --          |
| Mission prompt text         | `space_white`       | --                 | Faint       |
| Skills slash commands        | `butterscotch`      | --                 | --          |
| Toolbar shortcut keys       | `butterscotch`      | --                 | Bold        |
| Toolbar labels              | `space_white`       | --                 | --          |
| Toolbar separators `●`      | `galaxy_gray`       | --                 | --          |

## Compact Variant (80x24 minimum)

In compact mode (`<120 cols`), the role filter collapses to a dropdown filter bar at top. The agent detail panel is hidden. The agent list takes full width. Pressing Enter on an agent drills down into a full-screen detail view.

```
╭──────────────────────────────────────────────────────────────────────────────╮
│  AGENT ROSTER   Agents: 12  Active: 7  Idle: 3  Stuck: [2]                 │
╰──────────────────────────────────────────────────────────────────────────────╯
╭─ Filter: [All] Captains Commanders Implementers Reviewers Design ───────────╮
╰──────────────────────────────────────────────────────────────────────────────╯
╭─ Agent List ────────────────────────────────────────────────────────────────╮
│ ●  AGENT            ROLE          SHIP          PHASE    TIME              │
│────────────────────────────────────────────────────────────────────────────│
│ ▸  capt-alpha       captain       Enterprise    PLAN     04:12            │
│ ▸  impl-bravo       implementer   Enterprise    RED      02:15            │
│ ▸  impl-charlie     implementer   Enterprise    GREEN    01:45            │
│ ▸  rev-delta        reviewer      Enterprise    REVW     00:30            │
│ ▸  capt-echo        captain       Nautilus      PLAN     01:00            │
│ ▸  impl-foxtrot     implementer   Nautilus      GREEN    03:20            │
│ ⚠  impl-golf        implementer   Nautilus      RED      05:45            │
│ ▸  design-hotel     design off.   Discovery     PLAN     00:15            │
│ ○  impl-india       implementer   Discovery     IDLE     --              │
│ ○  impl-juliet      implementer   Discovery     IDLE     --              │
│ ⚠  rev-kilo         reviewer      Discovery     REVW     06:10            │
│ ○  cmdr-lima        commander     Discovery     IDLE     --              │
│────────────────────────────────────────────────────────────────────────────│
│ 12 of 12 agents │ ▸ Active: 7  ○ Idle: 3  ⚠ Stuck: 2                     │
╰──────────────────────────────────────────────────────────────────────────────╯
╭──────────────────────────────────────────────────────────────────────────────╮
│ [n] New  ● [Enter] Edit  ● [a] Assign  ● [d] Detach  ● [/] Srch  ● [?] H │
╰──────────────────────────────────────────────────────────────────────────────╯
```

### Compact Variant Notes

- Role filter sidebar collapses to a single-line dropdown filter bar at top. The active filter (`All`) is shown in brackets with `moonlit_violet`. Other categories use `galaxy_gray`.
- Agent detail panel is completely hidden. Enter on a selected agent navigates to a full-screen agent detail view.
- Header collapses from 2 content rows to 1. Labels are abbreviated.
- Toolbar labels are shortened: "Srch" for Search, "H" for Help. Delete and Detach are hidden behind `[?]` help overlay.
- The agent list adds a TIME column for elapsed time, taking advantage of the full width.
- A summary footer row shows agent counts by status.
- The selected row still uses `moonlit_violet` left-border for focus indication.
- Scrolling is required if the agent list exceeds the visible area.
