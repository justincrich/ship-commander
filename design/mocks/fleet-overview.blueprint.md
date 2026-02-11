# Fleet Overview -- ASCII Blueprint

> Source: `fleet-overview.mock.html` | Spec: `prompts/fleet-overview.spec.json`
> Terminal: 120 cols x 30 rows (standard breakpoint)

## Layout Structure

The fleet overview uses a three-row vertical stack:

```
HeaderBar (100%, 3 rows)
──────────────────────────────────────
ShipList (60%)  +  ShipPreview (40%)
   (fill height, scrollable)
──────────────────────────────────────
NavigableToolbar (100%, 1 row)
```

- **Header**: 3 rows. Row 1 is the FLEET COMMAND title in bold gold. Row 2 is fleet-wide metrics inline (ship count, launched count, pending messages pink badge, fleet health, completion).
- **Main content**: Fill remaining height. Ship list (60% width, left) shows scrollable ShipCards sorted launched-first, then docked, then complete. Ship preview (40% width, right) shows detail for the selected ship.
- **Toolbar**: 1 row. NavigableToolbar with shortcut buttons. Arrow Left/Right to highlight, Enter to activate, or press shortcut key directly.

## Component Mapping (from components.yaml)

| Visual Area | Component | Charm Base | Key Props |
|---|---|---|---|
| Full screen container | **AppShell** (#1) | `lipgloss.JoinVertical` | header, body, footer, width, height |
| Header row | **HeaderBar** (fleet variant) | `lipgloss.NewStyle()` | fleet metrics inline, health bar |
| Ship List + Preview layout | **PanelGrid** (#2) | `lipgloss.JoinHorizontal` | ratios: `[0.6, 0.4]`, compact_threshold: 120 |
| Ship List panel | **FocusablePanel** (#4) | `lipgloss.NewStyle().Border(lipgloss.RoundedBorder())` | title: "Ship List", focused: moonlit_violet, default: blue |
| Ship card list (scrollable) | **AgentGrid** pattern via `bubbles.list` | `bubbles.list` | filterable, keyboard nav (j/k, arrows) |
| Individual ship card | Custom `lipgloss.NewStyle()` | `lipgloss.NewStyle().Border(lipgloss.RoundedBorder())` | name, class, directive, crew count, progress |
| Ship Preview panel | **FocusablePanel** (#4) | `lipgloss.NewStyle().Border(lipgloss.RoundedBorder())` | title: "Ship Preview" |
| Status badges (LAUNCHED/DOCKED/COMPLETE/HALTED) | **StatusBadge** (#6) | `lipgloss.NewStyle()` | Variants: running, done, waiting, halted |
| Progress bars on ship cards | **WaveProgressBar** (#9) | `bubbles.progress` | gradient_start: butterscotch, gradient_end: gold |
| Health indicator in header | **HealthBar** (#7) | `lipgloss.NewStyle()` | 5-dot filled/empty, color by status (ok/warning/critical) |
| Elapsed timers | **ElapsedTimer** (#10) | `bubbles.stopwatch` | format: auto, warning_threshold: 5m |
| Crew entries in preview | **AgentCard** (#11) compact | `lipgloss.NewStyle()` | name, role, status icon |
| Bottom toolbar | **NavigableToolbar** | `lipgloss.JoinHorizontal` | Shortcut buttons, arrow nav + Enter, moonlit_violet highlight |

## Token Reference

| Element              | Token(s)                                          | Hex         |
|----------------------|---------------------------------------------------|-------------|
| Header title         | `gold` (bold)                                     | `#FFAA00`   |
| Header metrics text  | `space_white`                                     | `#F5F6FA`   |
| Header metrics label | `light_gray`                                      | `#CCCCCC`   |
| Messages badge       | `pink` (bg), `black` (fg)                         | `#FF99CC`   |
| Health indicator OK  | `green_ok`                                        | `#33FF33`   |
| Panel border default | `lipgloss.RoundedBorder()`, `galaxy_gray`         | `#52526A`   |
| Panel border focused | `lipgloss.RoundedBorder()`, `moonlit_violet`      | `#9966FF`   |
| Ship name            | `butterscotch` (bold)                             | `#FF9966`   |
| Ship class           | `galaxy_gray` (faint)                             | `#52526A`   |
| Directive text       | `blue`                                            | `#9999CC`   |
| No Directive text    | `galaxy_gray`                                     | `#52526A`   |
| Status LAUNCHED      | `butterscotch` (bg), `black` (fg)                 | `#FF9966`   |
| Status DOCKED        | `blue` (bg), `black` (fg)                         | `#9999CC`   |
| Status COMPLETE      | `green_ok` (bg), `black` (fg)                     | `#33FF33`   |
| Status HALTED        | `red_alert` (bg), `black` (fg)                    | `#FF3333`   |
| Progress filled      | `butterscotch` (active), `green_ok` (complete)    | `#FF9966`   |
| Progress empty       | `galaxy_gray`                                     | `#52526A`   |
| Preview section head | `gold` (bold, uppercase)                          | `#FFAA00`   |
| Crew name            | `space_white`                                     | `#F5F6FA`   |
| Crew role            | `blue`                                            | `#9999CC`   |
| Activity time        | `galaxy_gray`                                     | `#52526A`   |
| Activity text        | `space_white`                                     | `#F5F6FA`   |
| Toolbar key          | `butterscotch` (bold)                             | `#FF9966`   |
| Toolbar label        | `space_white`                                     | `#F5F6FA`   |
| Toolbar separator    | `galaxy_gray`                                     | `#52526A`   |
| Modal border         | `lipgloss.DoubleBorder()`, `butterscotch`         | `#FF9966`   |

## ASCII Blueprint (120x30 standard)

```
╭──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╮
│  FLEET COMMAND                                                                                                     │
│  Ships: 6   Launched: 3   Messages: [4]   Fleet Health: ● Optimal   Completion: 67%                               │
╰──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╯
╭─ Ship List ──────────────────────────────────────────────────────╮╭─ Ship Preview ───────────────────────────────────╮
│ ╭────────────────────────────────────────────────────────────╮   ││ USS Enterprise                                   │
│ │ USS Enterprise  Constitution-class            LAUNCHED     │   ││ Constitution-class  ● LAUNCHED                   │
│ │ Explore strange new worlds and seek out new life           │   ││─────────────────────────────────────────────────  │
│ │ Crew: 5   Mission 3/5                                     │   ││ DIRECTIVE                                        │
│ │ ██████████████████░░░░░░░░░░░░                  60%        │   ││ ┃ Explore strange new worlds and seek out new    │
│ ╰────────────────────────────────────────────────────────────╯   ││ ┃ life and new civilizations. Document all        │
│ ╭────────────────────────────────────────────────────────────╮   ││ ┃ encounters and maintain peaceful first contact  │
│ │ USS Defiant  Defiant-class                    LAUNCHED     │   ││ ┃ protocols.                                     │
│ │ Tactical defense operations in the Gamma Quadrant          │   ││                                                  │
│ │ Crew: 4   Mission 7/10                                    │   ││ CREW ROSTER                                      │
│ │ ██████████████████████░░░░░░░░░░                70%        │   ││ captain-kirk                       Commander     │
│ ╰────────────────────────────────────────────────────────────╯   ││ spock                         Science Officer     │
│ ╭────────────────────────────────────────────────────────────╮   ││ scotty                              Engineer     │
│ │ USS Voyager  Intrepid-class                   LAUNCHED     │   ││ bones                                Medical     │
│ │ Return journey from Delta Quadrant via sustainable routes  │   ││ uhura                         Communications     │
│ │ Crew: 6   Mission 12/20                                   │   ││                                                  │
│ │ ██████████████████░░░░░░░░░░░░                  60%        │   ││ MISSION PROGRESS                                 │
│ ╰────────────────────────────────────────────────────────────╯   ││ Wave 3 of 5  ● 60% Complete                      │
│ ╭────────────────────────────────────────────────────────────╮   ││ ██████████████████░░░░░░░░░░░░                    │
│ │ USS Discovery  Crossfield-class                 DOCKED     │   ││                                                  │
│ │ Research and development of spore drive technology         │   ││ RECENT ACTIVITY                                  │
│ │ Crew: 3   Mission 0/8                                     │   ││  2m ago  Wave 3 started                          │
│ │ ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░                   0%       │   ││ 15m ago  Encounter logged: Species 4713          │
│ ╰────────────────────────────────────────────────────────────╯   ││  1h ago  Wave 2 completed successfully            │
│ ╭────────────────────────────────────────────────────────────╮   ││  3h ago  First contact protocol initiated         │
│ │ USS Reliant  Miranda-class                   COMPLETE      │   ││  5h ago  Entered uncharted sector M-113           │
│ │ Survey and catalog Regula I sector anomalies               │   ││                                                  │
│ │ Crew: 4   Mission 6/6                                     │   ││ ACTIONS                                          │
│ │ ██████████████████████████████                 100%         │   ││ [Enter] Bridge  [p] Planning  [a] Crew  [d] Dir  │
│ ╰────────────────────────────────────────────────────────────╯   ││                                                  │
╰──────────────────────────────────────────────────────────────────╯╰──────────────────────────────────────────────────╯
╭──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╮
│ [n] New Ship  ●  [d] Directive  ●  [r] Roster  ●  [i] Inbox  ●  [m] Monitor  ●  [s] Settings  ●  [?] Help  ● [q] Q│
╰──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╯
```

### Annotation Notes

- The **selected ship card** (USS Enterprise) uses `moonlit_violet` (`#9966FF`) for its inner border `╭╮╰╯`. All other ship cards use `galaxy_gray` (`#52526A`) borders.
- The **Ship List** panel label (`Ship List`) renders inline on the top border in `galaxy_gray`.
- The **Ship Preview** panel label (`Ship Preview`) renders inline on the top border in `galaxy_gray`.
- Status badges are rendered with colored backgrounds: `LAUNCHED` in butterscotch bg, `DOCKED` in blue bg, `COMPLETE` in green_ok bg.
- Progress bars use `█` (`progress_filled`) and `░` (`progress_empty`). Active ships use `butterscotch` fill; complete ships use `green_ok` fill.
- The `[4]` messages badge renders with `pink` (`#FF99CC`) background and `black` foreground.
- Fleet health `●` indicator uses `green_ok` (`#33FF33`) for Optimal status.
- Toolbar `●` separators use `galaxy_gray`. Shortcut keys like `[n]` use `butterscotch`. Labels use `space_white`.

## Color Annotations

| Region                      | Foreground Token    | Background Token   | Style       |
|-----------------------------|---------------------|--------------------|-------------|
| "FLEET COMMAND"             | `gold`              | --                 | Bold        |
| Header metric labels        | `light_gray`        | --                 | --          |
| Header metric values        | `space_white`       | --                 | Bold        |
| Messages `[4]`              | `black`             | `pink`             | Bold        |
| Health `●`                  | `green_ok`          | --                 | --          |
| Panel border (default)      | `galaxy_gray`       | --                 | Rounded     |
| Panel border (focused)      | `moonlit_violet`    | --                 | Rounded     |
| Ship name                   | `butterscotch`      | --                 | Bold        |
| Ship class                  | `galaxy_gray`       | --                 | Faint       |
| Directive text              | `blue`              | --                 | --          |
| "No Directive"              | `galaxy_gray`       | --                 | Faint       |
| Status `LAUNCHED`           | `black`             | `butterscotch`     | Bold        |
| Status `DOCKED`             | `black`             | `blue`             | Bold        |
| Status `COMPLETE`           | `black`             | `green_ok`         | Bold        |
| Progress `█`                | `butterscotch`      | --                 | --          |
| Progress `█` (complete)     | `green_ok`          | --                 | --          |
| Progress `░`                | `galaxy_gray`       | --                 | --          |
| Preview section titles      | `gold`              | --                 | Bold, Upper |
| Preview directive `┃`       | `blue`              | --                 | --          |
| Crew names                  | `space_white`       | --                 | --          |
| Crew roles                  | `blue`              | --                 | --          |
| Activity timestamps         | `galaxy_gray`       | --                 | --          |
| Activity text               | `space_white`       | --                 | --          |
| Action keys `[Enter]`       | `butterscotch`      | --                 | Bold        |
| Toolbar shortcut keys       | `butterscotch`      | --                 | Bold        |
| Toolbar labels              | `space_white`       | --                 | --          |
| Toolbar separators `●`      | `galaxy_gray`       | --                 | --          |

## Compact Variant (80x24 minimum)

In compact mode (`<120 cols`), the ship preview panel is hidden. The ship list takes full width. Fleet summary collapses to a single line. Pressing Enter on a ship goes directly to the Ship Bridge.

```
╭──────────────────────────────────────────────────────────────────────────────────╮
│  FLEET COMMAND   Ships: 6  Launched: 3  Msgs: [4]  Health: ● OK  Done: 67%     │
╰──────────────────────────────────────────────────────────────────────────────────╯
╭─ Ship List ──────────────────────────────────────────────────────────────────────╮
│ ╭────────────────────────────────────────────────────────────────────────────╮   │
│ │ USS Enterprise  Constitution-class                          LAUNCHED      │   │
│ │ Explore strange new worlds and seek out new life                          │   │
│ │ Crew: 5   Mission 3/5   ██████████████████░░░░░░░░░░░░             60%   │   │
│ ╰────────────────────────────────────────────────────────────────────────────╯   │
│ ╭────────────────────────────────────────────────────────────────────────────╮   │
│ │ USS Defiant  Defiant-class                                 LAUNCHED      │   │
│ │ Tactical defense operations in the Gamma Quadrant                         │   │
│ │ Crew: 4   Mission 7/10  ██████████████████████░░░░░░░░░░           70%   │   │
│ ╰────────────────────────────────────────────────────────────────────────────╯   │
│ ╭────────────────────────────────────────────────────────────────────────────╮   │
│ │ USS Voyager  Intrepid-class                                LAUNCHED      │   │
│ │ Return journey from Delta Quadrant via sustainable routes                 │   │
│ │ Crew: 6   Mission 12/20 ██████████████████░░░░░░░░░░░░             60%   │   │
│ ╰────────────────────────────────────────────────────────────────────────────╯   │
│ ╭────────────────────────────────────────────────────────────────────────────╮   │
│ │ USS Discovery  Crossfield-class                              DOCKED      │   │
│ │ Research and development of spore drive technology                        │   │
│ │ Crew: 3   Mission 0/8   ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░              0%  │   │
│ ╰────────────────────────────────────────────────────────────────────────────╯   │
╰──────────────────────────────────────────────────────────────────────────────────╯
╭──────────────────────────────────────────────────────────────────────────────────╮
│ [n] New  ● [d] Dir  ● [r] Roster  ● [i] Inbox  ● [s] Set  ● [?] Help  ● [q] Q │
╰──────────────────────────────────────────────────────────────────────────────────╯
```

### Compact Variant Notes

- Ship preview panel is completely hidden. Enter on a selected ship navigates directly to Ship Bridge.
- Header collapses from 2 content rows to 1. Labels are abbreviated: "Msgs" for Messages, "OK" for Optimal, "Done" for Completion.
- Toolbar labels are shortened: "Dir" for Directive, "Set" for Settings, "Q" for Quit.
- Ship cards combine the progress bar and meta line into a single row for vertical space savings.
- Scrolling is required to see all ships since only ~4 cards fit in the visible area at 24 rows.
- The selected card still uses `moonlit_violet` border for focus indication.
