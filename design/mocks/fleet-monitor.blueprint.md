# Fleet Monitor -- ASCII Blueprint

> Source: `fleet-monitor.mock.html` | Spec: `prompts/fleet-monitor.spec.json`
> Terminal: 120 cols x 30 rows (standard breakpoint)

## Layout Structure

```
Row 1-2:   MonitorHeader    [100%]  -- Title, active ships, agents, wave progress, pending badge
Row 3-4:   GridHeader       [100%]  -- Column labels for ship status grid
Row 5-25:  ShipStatusGrid   [100%]  -- Scrollable ship rows (one row per ship)
Row 26:    Separator         ---    -- Border-bottom closing the grid
Row 27-28: NavigableToolbar [100%]  -- [Enter] Bridge  [i] Inbox  [?] Help  [Esc] Fleet
```

Vertical flow: `MonitorHeader | ShipStatusGrid (fill) | NavigableToolbar`

## Component Mapping (from components.yaml)

| Visual Area | Component | Charm Base | Key Props |
|---|---|---|---|
| Full screen container | **AppShell** (#1) | `lipgloss.JoinVertical` | header, body, footer |
| Monitor header | **HeaderBar** (fleet-monitor variant) | `lipgloss.NewStyle()` | fleet health, active ships, agent count |
| Ship status grid | **PanelGrid** (#2) as table layout | `lipgloss.JoinVertical` | one row per ship |
| Per-ship status row | Custom `lipgloss.NewStyle()` table row | `lipgloss.NewStyle()` | ship name, directive, crew count, wave, health |
| Status badges | **StatusBadge** (#6) | `lipgloss.NewStyle()` | LAUNCHED/DOCKED/COMPLETE/HALTED variants |
| Health indicators | **HealthBar** (#7) | `lipgloss.NewStyle()` | compact 5-dot per ship |
| Wave progress bars | **WaveProgressBar** (#9) compact | `bubbles.progress` | inline progress per ship |
| Pending question badges | Custom `lipgloss.NewStyle()` | `lipgloss.NewStyle()` | pink badge with count |
| Bottom toolbar | **NavigableToolbar** | `lipgloss.JoinHorizontal` | [Enter] Bridge [i] Inbox [?] Help [Esc] Fleet |

## Token Reference

| Element             | Token            | Hex       | Lipgloss Call                          |
|---------------------|------------------|-----------|----------------------------------------|
| Title text          | butterscotch     | `#FF9966` | `style.Foreground(lipgloss.Color("#FF9966"))` |
| Stat labels         | blue             | `#9999CC` | `style.Foreground(lipgloss.Color("#9999CC"))` |
| Stat values         | space_white      | `#F5F6FA` | `style.Foreground(lipgloss.Color("#F5F6FA")).Bold(true)` |
| Pending badge       | pink             | `#FF99CC` | `style.Background(lipgloss.Color("#FF99CC")).Foreground(lipgloss.Color("#000000"))` |
| Selected row        | moonlit_violet   | `#9966FF` | `style.BorderLeft(true).BorderForeground(lipgloss.Color("#9966FF"))` |
| LAUNCHED row        | butterscotch     | `#FF9966` | `style.Foreground(lipgloss.Color("#FF9966"))` |
| COMPLETE row        | green_ok         | `#33FF33` | `style.Foreground(lipgloss.Color("#33FF33"))` |
| HALTED row          | red_alert        | `#FF3333` | `style.Foreground(lipgloss.Color("#FF3333"))` |
| STUCK row           | yellow_caution   | `#FFCC00` | `style.Foreground(lipgloss.Color("#FFCC00"))` |
| Directive text      | blue             | `#9999CC` | `style.Foreground(lipgloss.Color("#9999CC"))` |
| Crew/mission text   | light_gray       | `#CCCCCC` | `style.Foreground(lipgloss.Color("#CCCCCC"))` |
| Progress filled     | (status color)   | varies    | `icon: \u2588 (full block)` |
| Progress empty      | galaxy_gray      | `#52526A` | `icon: \u2591 (light shade)` |
| Panel border        | blue             | `#9999CC` | `lipgloss.RoundedBorder()` |
| Toolbar key         | moonlit_violet   | `#9966FF` | `style.Foreground(lipgloss.Color("#9966FF")).Bold(true)` |
| Toolbar label       | butterscotch     | `#FF9966` | `style.Foreground(lipgloss.Color("#FF9966"))` |
| Health star filled   | (status color)  | varies    | `icon: \u2605 (black star)` |
| Health star empty    | galaxy_gray     | `#52526A` | `icon: \u2606 (white star)` |
| Borders             | blue             | `#9999CC` | `lipgloss.RoundedBorder()` chars: \u256D \u256E \u2570 \u256F \u2502 \u2500` |
| Icons: launched     | --               | --        | `\u25B8` (right-pointing triangle) |
| Icons: complete     | --               | --        | `\u2713` (checkmark) |
| Icons: halted       | --               | --        | `\u26A0` (warning sign) |

## ASCII Blueprint (120x30 standard)

```
╭─ FLEET MONITOR ──────────────────────────────────────────────────────────────────────────────────────────────────────╮
│  Active Ships: 3    Total Agents: 12    Fleet Progress: Wave 4/8    Pending: [2]                                   │
├──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
│ STATUS  SHIP NAME              DIRECTIVE                       CREW     WAVE          MISSIONS        HEALTH       │
├──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
┃ ▸      SS Enterprise           Implement auth system           Crew: 4  3/6 ████████░░░░  Missions: 12/18  ★★★★☆  ┃
│ ✓      HMS Beagle              Database migration tooling      Crew: 2  5/5 ████████████  Missions: 15/15  ★★★★★  │
│ ▸      Nautilus                Refactor payment gateway        Crew: 3  2/7 ████░░░░░░░░  Missions:  8/21  ★★★☆☆  │
│ ⚠      Discovery               API documentation overhaul     Crew: 3  1/4 ███░░░░░░░░░  Missions:  3/12  ★★☆☆☆  │
│ ▸      Voyager                 Build CI/CD pipeline            Crew: 2  4/8 ██████░░░░░░  Missions: 16/24  ★★★★☆  │
│                                                                                                                    │
│                                                                                                                    │
│                                                                                                                    │
│                                                                                                                    │
│                                                                                                                    │
│                                                                                                                    │
│                                                                                                                    │
│                                                                                                                    │
│                                                                                                                    │
│                                                                                                                    │
│                                                                                                                    │
│                                                                                                                    │
│                                                                                                                    │
│                                                                                                                    │
│                                                                                                                    │
│                                                                                                                    │
╰──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╯
╭──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╮
│                          [Enter] Bridge       [i] Inbox       [?] Help       [Esc] Fleet                           │
╰──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╯
```

### Key Rendering Notes

- **Row 1**: `╭─ FLEET MONITOR ─...─╮` -- title embedded in top border, butterscotch bold
- **Row 2**: Stat labels in blue (`#9999CC`), stat values in space_white bold; `[2]` badge rendered with pink background
- **Row 3**: Horizontal rule separator `├─...─┤`
- **Row 4**: Column headers in butterscotch bold
- **Row 5**: Separator `├─...─┤`
- **Row 6** (selected): Left border accent `┃` in moonlit_violet (`#9966FF`), row background tinted violet
  - `▸` icon in butterscotch, ship name bold, directive in blue, progress bar `████████░░░░` in butterscotch/gray
  - Health stars: `★★★★☆` (4 green_ok filled, 1 galaxy_gray empty)
- **Row 7**: Complete ship -- `✓` icon and entire row in green_ok (`#33FF33`), progress fully filled
- **Row 8**: Launched ship -- `▸` in butterscotch, health `★★★☆☆` in yellow_caution
- **Row 9**: Halted ship -- `⚠` icon, entire row in red_alert (`#FF3333`), health `★★☆☆☆` in red_alert
- **Row 10**: Launched ship -- standard butterscotch treatment
- **Rows 11-26**: Empty scrollable area (fills remaining terminal height)
- **Row 27**: Grid bottom border `╰─...─╯`
- **Rows 28-30**: NavigableToolbar -- keys in moonlit_violet brackets, labels in butterscotch

## Color Annotations

```
╭─ FLEET MONITOR ──...──╮                     butterscotch (#FF9966) bold, border in blue (#9999CC)
│  Active Ships: 3  ...  │                     "Active Ships:" = blue (#9999CC), "3" = space_white (#F5F6FA) bold
│  ...  Pending: [2]     │                     "[2]" = pink bg (#FF99CC), black text (#000000)
├────────────────────────┤                     border = blue (#9999CC)
│ STATUS  SHIP NAME  ... │                     column headers = butterscotch (#FF9966) bold
├────────────────────────┤                     border = blue (#9999CC)
┃ ▸  SS Enterprise  ...  ┃  <-- selected       moonlit_violet (#9966FF) left border, violet bg tint
│ ✓  HMS Beagle     ...  │  <-- complete        green_ok (#33FF33) text + icon
│ ▸  Nautilus       ...  │  <-- launched        butterscotch (#FF9966) icon, health = yellow_caution (#FFCC00)
│ ⚠  Discovery      ... │  <-- halted          red_alert (#FF3333) text + icon
│ ▸  Voyager        ...  │  <-- launched        butterscotch (#FF9966) icon, health = green_ok (#33FF33)
╰────────────────────────╯                     border = blue (#9999CC)
╭────────────────────────╮                     toolbar bg tint = dark_blue (#1B4F8F) 20% opacity
│  [Enter] Bridge  ...   │                     keys = moonlit_violet (#9966FF), labels = butterscotch (#FF9966)
╰────────────────────────╯                     border = galaxy_gray (#52526A)
```

### Per-Row Color Breakdown

| Ship Status | Icon  | Icon Color   | Name Color   | Directive    | Progress Bar      | Health Color     |
|-------------|-------|--------------|--------------|--------------|-------------------|------------------|
| LAUNCHED    | `▸`   | butterscotch | butterscotch | blue         | butterscotch/gray | green_ok         |
| COMPLETE    | `✓`   | green_ok     | green_ok     | green_ok     | green_ok          | green_ok         |
| HALTED      | `⚠`   | red_alert    | red_alert    | red_alert    | butterscotch/gray | red_alert        |
| STUCK       | `▸`   | yellow_caution| yellow_caution| blue        | butterscotch/gray | yellow_caution   |
| (selected)  | any   | (as above)   | (as above)   | (as above)   | (as above)        | (as above)       |

Selected row overlay: `background: rgba(153, 102, 255, 0.25)` + moonlit_violet left border accent

## Compact Variant (80x24 minimum)

When terminal width falls below 120 columns, rows truncate per the spec:
`icon + name + wave bar + mission count` (directive, crew, and health hidden).

```
╭─ FLEET MONITOR ──────────────────────────────────────────────────────────────╮
│  Ships: 3  Agents: 12  Wave: 4/8  Pending: [2]                             │
├──────────────────────────────────────────────────────────────────────────────┤
│ ST  SHIP NAME              WAVE          MISSIONS                           │
├──────────────────────────────────────────────────────────────────────────────┤
┃ ▸   SS Enterprise          3/6 ████████░░░░  12/18                          ┃
│ ✓   HMS Beagle             5/5 ████████████  15/15                          │
│ ▸   Nautilus               2/7 ████░░░░░░░░   8/21                          │
│ ⚠   Discovery              1/4 ███░░░░░░░░░   3/12                          │
│ ▸   Voyager                4/8 ██████░░░░░░  16/24                          │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
╰──────────────────────────────────────────────────────────────────────────────╯
╭──────────────────────────────────────────────────────────────────────────────╮
│                [Enter] Bridge    [i] Inbox    [?] Help    [Esc] Fleet       │
╰──────────────────────────────────────────────────────────────────────────────╯
```

### Compact Differences

| Attribute       | Standard (120+)                        | Compact (<120)                     |
|-----------------|----------------------------------------|------------------------------------|
| Header stats    | Full labels                            | Abbreviated: `Ships`, `Wave`       |
| Columns shown   | Status, Name, Directive, Crew, Wave, Missions, Health | Status, Name, Wave, Missions |
| Directive       | Visible (200px equivalent)             | Hidden                             |
| Crew count      | `Crew: N`                              | Hidden (drill-down only)           |
| Health stars    | `★★★★☆`                                | Hidden (drill-down only)           |
| Wave bar        | 12-char bar                            | 12-char bar (preserved)            |
| Mission count   | `Missions: N/M`                        | `N/M` (label removed)              |
