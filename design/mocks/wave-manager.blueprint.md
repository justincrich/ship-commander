# Wave Manager -- ASCII Blueprint

> Source: `wave-manager.mock.html` | Spec: `prompts/wave-manager.spec.json`
> Terminal: 120 cols x 30 rows (standard breakpoint)

## Layout Structure

The wave manager is an **overlay** on top of the Ship Bridge. The underlying bridge content is dimmed to 15% opacity, and the modal fills 95% width / 90% height centered.

```
[Dimmed Ship Bridge background at 15% opacity]
  ╔═══════════════════════════════════════════════════════════════╗
  ║ WaveHeader (100%, 2 rows)                                    ║
  ║──────────────────────────────────────────────────────────────║
  ║ WaveList (30%)  │  WaveMissionDetail (70%)                   ║
  ║   (fill height, scrollable)                                  ║
  ║──────────────────────────────────────────────────────────────║
  ║ NavigableToolbar (100%, 1 row)                               ║
  ╚═══════════════════════════════════════════════════════════════╝
```

- **WaveHeader**: 2 rows. Row 1 is the overlay title "WAVE MANAGER -- <Ship Name>" in bold butterscotch. Row 2 is summary stats: "N/M waves complete, P missions merged." in blue. Separated from content by a gold horizontal rule.
- **Content area**: Fill remaining height. Wave list (30% width, left) shows all waves with progress bars and mission counts. Selected wave highlighted with moonlit_violet left border. Wave mission detail (70% width, right) shows a table of missions in the selected wave with ID, title, status badge, conflict indicator, and merge checkbox. Separated by a galaxy_gray vertical divider.
- **Toolbar**: 1 row. NavigableToolbar with action buttons. Separated from content by a gold horizontal rule.

## Component Mapping (from components.yaml)

| Visual Area | Component | Charm Base | Key Props |
|---|---|---|---|
| Modal container | **ModalOverlay** (#5) | `lipgloss.Place` + `lipgloss.DoubleBorder()` | 95% width, 90% height, gold border |
| Dimmed background | `lipgloss.NewStyle().Faint(true)` | `lipgloss.NewStyle()` | Ship Bridge at 15% opacity |
| Wave header | Custom `lipgloss.NewStyle()` | `lipgloss.NewStyle()` | title, wave/mission counts |
| Wave list + Detail layout | **PanelGrid** (#2) | `lipgloss.JoinHorizontal` | ratios: `[0.3, 0.7]` |
| Wave list | **FocusablePanel** (#4) wrapping `bubbles.list` | `bubbles.list` | wave items with progress bars |
| Wave progress bars | **WaveProgressBar** (#9) | `bubbles.progress` | active/complete/pending variants |
| Mission table | `bubbles.table` | `bubbles.table` | ID, title, status, conflict indicator |
| Status badges | **StatusBadge** (#6) | `lipgloss.NewStyle()` | done/in_progress/backlog/halted |
| Bottom toolbar | **NavigableToolbar** | `lipgloss.JoinHorizontal` | [m] Merge [r] Reorder [d] Deps [?] Help [Esc] Bridge |

## Token Reference

| Element                  | Token(s)                                          | Hex         |
|--------------------------|---------------------------------------------------|-------------|
| Overlay border           | `lipgloss.DoubleBorder()`, `gold`                 | `#FFAA00`   |
| Overlay title            | `butterscotch` (bold)                             | `#FF9966`   |
| Summary stats            | `blue`                                            | `#9999CC`   |
| Divider rules            | `gold`                                            | `#FFAA00`   |
| Wave title (active)      | `butterscotch` (bold)                             | `#FF9966`   |
| Wave title (complete)    | `green_ok` (bold)                                 | `#33FF33`   |
| Wave title (pending)     | `galaxy_gray`                                     | `#52526A`   |
| Wave left border default | `galaxy_gray`                                     | `#52526A`   |
| Wave left border select  | `moonlit_violet`                                  | `#9966FF`   |
| Wave left border done    | `green_ok`                                        | `#33FF33`   |
| Progress filled          | `butterscotch` (active), `green_ok` (complete)    | `#FF9966`   |
| Progress empty           | `galaxy_gray`                                     | `#52526A`   |
| Wave stats               | `light_gray`                                      | `#CCCCCC`   |
| Wave icon active         | `butterscotch` (▸)                                | `#FF9966`   |
| Wave icon complete       | `green_ok` (✓)                                    | `#33FF33`   |
| Wave icon pending        | `galaxy_gray` (○)                                 | `#52526A`   |
| Detail header            | `moonlit_violet` (bold)                           | `#9966FF`   |
| Detail divider           | `galaxy_gray`                                     | `#52526A`   |
| Table column headers     | `gold` (bold)                                     | `#FFAA00`   |
| Mission ID               | `blue`                                            | `#9999CC`   |
| Mission title            | `space_white`                                     | `#F5F6FA`   |
| Status READY             | `green_ok` (fg), `green_ok` 20% (bg/border)       | `#33FF33`   |
| Status ACTIVE            | `butterscotch` (fg), `butterscotch` 20% (bg/border)| `#FF9966`  |
| Status DONE              | `blue` (fg), `blue` 20% (bg/border)               | `#9999CC`   |
| Conflict icon (none)     | `galaxy_gray`                                     | `#52526A`   |
| Conflict icon (warning)  | `red_alert`                                       | `#FF3333`   |
| Merge checkbox (enabled) | `green_ok`                                        | `#33FF33`   |
| Merge checkbox (disabled)| `galaxy_gray`                                     | `#52526A`   |
| Toolbar key              | `butterscotch` (bold)                             | `#FF9966`   |
| Toolbar label            | `light_gray`                                      | `#CCCCCC`   |
| Toolbar key (highlight)  | `moonlit_violet` (bold)                           | `#9966FF`   |

## ASCII Blueprint (120x30 standard)

```
  ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░
  ░░ USS ENTERPRISE - SHIP BRIDGE (dimmed 15%)  STATUS: NOMINAL   DIRECTIVE: feature/quantum-drive   MISSIONS: 12/15 ░░
  ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░
  ╔════════════════════════════════════════════════════════════════════════════════════════════════════════════════════╗
  ║  WAVE MANAGER -- USS Enterprise                                                                                  ║
  ║  3/4 waves complete, 12 missions merged.                                                                         ║
  ║══════════════════════════════════════════════════════════════════════════════════════════════════════════════════════║
  ║                                  │                                                                               ║
  ║  ┃ Wave 1                        │  Wave 2 - Missions (5)                                                        ║
  ║  ┃ ████████████  4/4 ✓           │──────────────────────────────────────────────────────────────────────────────── ║
  ║  ┃                               │  ID      TITLE                           STATUS       CONF     MERGE          ║
  ║                                  │──────────────────────────────────────────────────────────────────────────────── ║
  ║  ┃ Wave 2                        │  M-201   Sensor calibration fix           READY         —        ☑             ║
  ║  ┃ ██████░░░░░░  3/5 ▸           │                                                                               ║
  ║  ┃                               │  M-202   Deflector shield optimization   READY         —        ☑             ║
  ║                                  │                                                                               ║
  ║  ┃ Wave 3                        │  M-203   Warp signature analysis          DONE         ⚠        ☐             ║
  ║  ┃ ░░░░░░░░░░░░  0/3 ○          │                                                                               ║
  ║  ┃                               │  M-204   Transporter buffer upgrade      ACTIVE        —        ☐             ║
  ║                                  │                                                                               ║
  ║  ┃ Wave 4                        │  M-205   Navigation chart update          READY        ⚠        ☐             ║
  ║  ┃ ░░░░░░░░░░░░  0/2 ○          │                                                                               ║
  ║                                  │                                                                               ║
  ║                                  │                                                                               ║
  ║                                  │                                                                               ║
  ║══════════════════════════════════════════════════════════════════════════════════════════════════════════════════════║
  ║  [m] Merge Ready   [c] Merge Conflicts   [?] Help   [Esc] Close                                                 ║
  ╚════════════════════════════════════════════════════════════════════════════════════════════════════════════════════╝
```

### Annotation Notes

- The entire overlay uses `lipgloss.DoubleBorder()` in `gold` (`#FFAA00`) with characters `╔ ╗ ╚ ╝ ║ ═`, distinguishing it from standard rounded-border panels.
- The **dimmed background** (top 3 rows with `░`) represents the Ship Bridge content at 15% opacity. In the actual TUI, the full bridge view renders behind the modal.
- The **overlay title** "WAVE MANAGER -- USS Enterprise" renders in `butterscotch` (`#FF9966`) bold.
- The **summary stats** "3/4 waves complete, 12 missions merged." render in `blue` (`#9999CC`).
- **Horizontal rules** (`═══`) within the overlay use `gold` (`#FFAA00`) to separate header, content, and toolbar.
- The **wave list** (left 30%) uses a left-border indicator `┃` for each wave. Wave 1 (complete) uses `green_ok` for `┃` and title. Wave 2 (selected/active) uses `moonlit_violet` for `┃` and title, with a subtle violet background tint. Waves 3 and 4 (pending) use `galaxy_gray` for `┃` and title.
- The **vertical divider** `│` between wave list and mission detail uses `galaxy_gray` (`#52526A`).
- The **detail header** "Wave 2 - Missions (5)" renders in `moonlit_violet` (`#9966FF`) bold.
- **Table column headers** (ID, TITLE, STATUS, CONF, MERGE) render in `gold` (`#FFAA00`) bold.
- **Status badges** are outlined style: `READY` in green_ok, `ACTIVE` in butterscotch, `DONE` in blue.
- **Conflict icons**: `—` (dash) in galaxy_gray means no conflict; `⚠` in red_alert means conflict detected.
- **Merge checkboxes**: `☑` in green_ok means merge-ready and selected; `☐` in galaxy_gray means not available for merge.
- **Toolbar** keys `[m]`, `[c]`, `[?]` render in `butterscotch` bold. `[Esc]` renders in `moonlit_violet` bold (highlighted). Labels render in `light_gray`.

## Color Annotations

| Region                        | Foreground Token    | Background Token   | Style       |
|-------------------------------|---------------------|--------------------|-------------|
| Overlay border `╔╗╚╝║═`      | `gold`              | --                 | Double      |
| "WAVE MANAGER -- USS..."      | `butterscotch`      | --                 | Bold        |
| Summary stats                 | `blue`              | --                 | --          |
| Internal dividers `═══`       | `gold`              | --                 | --          |
| Wave title (complete)         | `green_ok`          | --                 | Bold        |
| Wave title (selected/active)  | `moonlit_violet`    | `moonlit_violet` 20% | Bold      |
| Wave title (pending)          | `galaxy_gray`       | --                 | Faint       |
| Wave left border `┃` complete | `green_ok`          | --                 | --          |
| Wave left border `┃` selected | `moonlit_violet`    | --                 | --          |
| Wave left border `┃` pending  | `galaxy_gray`       | --                 | --          |
| Progress `█` (active)         | `butterscotch`      | --                 | --          |
| Progress `█` (complete)       | `green_ok`          | --                 | --          |
| Progress `░`                  | `galaxy_gray`       | --                 | --          |
| Wave stats (e.g., "4/4")      | `light_gray`        | --                 | --          |
| Wave icon `✓`                 | `green_ok`          | --                 | --          |
| Wave icon `▸`                 | `butterscotch`      | --                 | --          |
| Wave icon `○`                 | `galaxy_gray`       | --                 | --          |
| Vertical divider `│`          | `galaxy_gray`       | --                 | --          |
| Detail header                 | `moonlit_violet`    | --                 | Bold        |
| Table headers                 | `gold`              | --                 | Bold        |
| Table header divider `───`    | `galaxy_gray`       | --                 | --          |
| Mission ID                    | `blue`              | --                 | --          |
| Mission title                 | `space_white`       | --                 | --          |
| Status `READY`                | `green_ok`          | `green_ok` 20%     | Bold        |
| Status `ACTIVE`               | `butterscotch`      | `butterscotch` 20% | Bold        |
| Status `DONE`                 | `blue`              | `blue` 20%         | Bold        |
| Conflict `—` (none)           | `galaxy_gray`       | --                 | --          |
| Conflict `⚠` (alert)         | `red_alert`         | --                 | --          |
| Merge `☑` (ready)             | `green_ok`          | --                 | --          |
| Merge `☐` (disabled)          | `galaxy_gray`       | --                 | --          |
| Toolbar keys `[m]` `[c]` `[?]`| `butterscotch`     | --                 | Bold        |
| Toolbar key `[Esc]`           | `moonlit_violet`    | --                 | Bold        |
| Toolbar labels                | `light_gray`        | --                 | --          |
| Dimmed background `░`         | `galaxy_gray`       | --                 | Faint       |

## Compact Variant (80x24 minimum)

In compact mode (`<120 cols`), the mission detail panel is hidden. The wave list takes full width with inline progress. Pressing Enter on a selected wave shows missions in a stacked view replacing the wave list.

```
░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░
░░ USS ENTERPRISE - SHIP BRIDGE (dimmed)                                        ░░
░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░
╔══════════════════════════════════════════════════════════════════════════════════╗
║  WAVE MANAGER -- USS Enterprise   3/4 complete, 12 merged.                     ║
║══════════════════════════════════════════════════════════════════════════════════║
║                                                                                ║
║  ┃ Wave 1   ████████████████████████████████  4/4 ✓                            ║
║                                                                                ║
║  ┃ Wave 2   ████████████████████░░░░░░░░░░░░  3/5 ▸                            ║
║                                                                                ║
║  ┃ Wave 3   ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░  0/3 ○                            ║
║                                                                                ║
║  ┃ Wave 4   ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░  0/2 ○                            ║
║                                                                                ║
║                                                                                ║
║                                                                                ║
║                                                                                ║
║══════════════════════════════════════════════════════════════════════════════════║
║  [m] Merge  [c] Conflicts  [?] Help  [Esc] Close                              ║
╚══════════════════════════════════════════════════════════════════════════════════╝
```

### Compact Variant Notes

- Mission detail panel is completely hidden. Enter on a selected wave navigates to an inline mission list replacing the wave list, with Back/Escape to return.
- Header collapses from 2 content rows to 1. Summary is abbreviated: "3/4 complete, 12 merged." instead of full text.
- Toolbar labels are shortened: "Merge" for "Merge Ready", "Conflicts" for "Merge Conflicts".
- Wave list items use wider progress bars since they have the full width available.
- The selected wave (Wave 2) still uses `moonlit_violet` left border `┃` for focus indication.
- Scrolling is required if more than ~4 waves exist since only 4 fit in the visible area at 24 rows.
