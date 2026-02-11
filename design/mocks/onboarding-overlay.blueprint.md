# Onboarding Overlay -- ASCII Blueprint

> Source: `onboarding-overlay.mock.html` | Spec: `prompts/onboarding-overlay.spec.json`
> Terminal: 120 cols x 30 rows (standard breakpoint)

## Layout Structure

The onboarding overlay is a first-run tour that renders on top of the dimmed Fleet Overview. It consists of two co-rendered elements: a highlighted panel region (full opacity) and an anchored tooltip card with step content.

```
┌─────────────────────────────────────────────────────────┐
│                Fleet Overview (dimmed 20%)               │
│                                                         │
│  ╔═ Highlighted Panel ═╗   ╔═══ Tooltip Card ═══════╗  │
│  ║  (full opacity)      ║◀──║  Step N of 5  ● ○ ○ ○  ║  │
│  ║  current target      ║   ║  Title                  ║  │
│  ║  panel content       ║   ║  Description text...    ║  │
│  ╚══════════════════════╝   ║  [Enter] Next [Esc] Skip║  │
│                             ╚═════════════════════════╝  │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

- **Background**: The Fleet Overview renders at 20% opacity (dimmed). All panels, header, and toolbar are visible but faded.
- **Highlighted panel**: The current tour step's target panel renders at full opacity with a `gold` double border and glow effect. Only one panel is highlighted at a time.
- **Tooltip card**: Anchored to the right of the highlighted panel. Uses `butterscotch` double border. Contains step indicator, title, description, and navigation hints.
- **Navigation**: `Enter` or `Right` advances, `Left` goes back, `Esc` skips the tour entirely.

## Tour Steps

| Step | Target Panel       | Title         | Highlighted Region     |
|------|--------------------|---------------|------------------------|
| 1    | `ship-list`        | Your Fleet    | Ship List panel (left) |
| 2    | `crew-panel`       | Your Crew     | Crew panel area        |
| 3    | `mission-board`    | Mission Board | Mission Board panel    |
| 4    | `event-log`        | Event Log     | Event Log panel        |
| 5    | `navigable-toolbar`| Navigation    | Bottom toolbar         |

## Component Mapping (from components.yaml)

| Visual Area | Component | Charm Base | Key Props |
|---|---|---|---|
| Dimmed background | `lipgloss.NewStyle()` | `lipgloss.NewStyle()` | Fleet Overview at 20% opacity |
| Highlighted panel | **FocusablePanel** (#4) variant | `lipgloss.NewStyle().Border(lipgloss.DoubleBorder())` | gold border, glow via `alert_pulse` spring |
| Tooltip card | **ModalOverlay** (#5) variant | `lipgloss.Place` + `lipgloss.DoubleBorder()` | butterscotch border, anchored to highlighted panel |
| Step indicator | Custom `lipgloss.NewStyle()` | `lipgloss.NewStyle()` | "Step N of 5", `●` (green_ok) / `○` (galaxy_gray) dots |
| Tooltip title | Custom `lipgloss.NewStyle()` | `lipgloss.NewStyle()` | butterscotch bold |
| Tooltip description | Custom `lipgloss.NewStyle()` | `lipgloss.NewStyle()` | space_white body text |
| Navigation hints | Custom `lipgloss.NewStyle()` | `lipgloss.NewStyle()` | `[Enter]`/`[Esc]` key badges with dark_blue bg |
| Pointer arrow | Custom `lipgloss.NewStyle()` | `lipgloss.NewStyle()` | `◀──` in butterscotch connecting tooltip to panel |

## Token Reference

| Element                  | Token(s)                                            | Hex         |
|--------------------------|-----------------------------------------------------|-------------|
| Overlay border           | `lipgloss.DoubleBorder()`, `butterscotch`           | `#FF9966`   |
| Highlighted panel border | `lipgloss.DoubleBorder()`, `gold`                   | `#FFAA00`   |
| Background (dimmed)      | Fleet Overview at 20% opacity                       | --          |
| Modal surface            | `surface_modal` (`black`)                           | `#000000`   |
| Step indicator text      | `butterscotch` (bold)                               | `#FF9966`   |
| Step dot (complete)      | `green_ok`                                          | `#33FF33`   |
| Step dot (current)       | `green_ok` (with glow)                              | `#33FF33`   |
| Step dot (pending)       | `galaxy_gray`                                       | `#52526A`   |
| Tooltip title            | `butterscotch` (bold)                               | `#FF9966`   |
| Tooltip description      | `space_white`                                       | `#F5F6FA`   |
| Navigation key labels    | `dark_blue` (bg), `space_white` (fg)                | `#1B4F8F`   |
| Navigation hint text     | `light_gray`                                        | `#CCCCCC`   |
| Separator lines          | `galaxy_gray`                                       | `#52526A`   |
| Ship name (highlighted)  | `butterscotch` (bold)                               | `#FF9966`   |
| Ship status text         | `galaxy_gray`                                       | `#52526A`   |
| Panel title (highlighted)| `butterscotch` (bold)                               | `#FF9966`   |
| Pointer arrow            | `butterscotch`                                      | `#FF9966`   |

## ASCII Blueprint -- Step 1: Your Fleet (120x30 standard)

```
╭──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╮
│  FLEET COMMAND                                                                                                     │
│  Ships: 6   Launched: 3   Messages: [4]   Fleet Health: ● Optimal   Completion: 67%                               │
╰──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╯
                                                                                                    ░░░░░░░░░░░░░░░░░░
                                                                                                    ░░░░░░░░░░░░░░░░░░
╔═ YOUR FLEET ═══════════════════════════════════════════════╗                                       ░░░░░░░░░░░░░░░░░░
║                                                           ║   ╔══════════════════════════════════════════════════════╗
║  ▸ USS ENTERPRISE                                         ║   ║  Step 1 of 5   ● ○ ○ ○ ○                            ║
║    Active • 5 missions • 3 agents                         ║◀──║──────────────────────────────────────────────────────║
║                                                           ║   ║  Your Fleet                                          ║
║    USS VOYAGER                                             ║   ║                                                      ║
║    Idle • 4 missions • 2 agents                           ║   ║  This is your fleet -- each ship represents a        ║
║                                                           ║   ║  project commission with its own crew and missions.   ║
║    USS DEFIANT                                             ║   ║  Press Enter to drill into any ship.                 ║
║    Active • 3 missions • 4 agents                         ║   ║                                                      ║
║                                                           ║   ║──────────────────────────────────────────────────────║
╚═══════════════════════════════════════════════════════════╝   ║  [Enter] Next                          [Esc] Skip Tour║
                                                                ╚══════════════════════════════════════════════════════╝
░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░
░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░
░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░
░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░
╭──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╮
│ [n] New Ship  ●  [d] Directive  ●  [r] Roster  ●  [i] Inbox  ●  [m] Monitor  ●  [s] Settings  ●  [?] Help  ● [q] Q│
╰──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╯
```

## ASCII Blueprint -- Step 5: Navigation (120x30 standard)

```
╭──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╮
│  FLEET COMMAND                                                                                                     │
│  Ships: 6   Launched: 3   Messages: [4]   Fleet Health: ● Optimal   Completion: 67%                               │
╰──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╯
░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░
░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░
░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░
░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░
░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░
░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░
░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░
░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░
░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░
░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░
░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░
░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░
╔══════════════════════════════════════════════════════════════════════════════════════════════════════════════════════╗
║  ╔══════════════════════════════════════════════════════════════════════════════════════════════╗                   ║
║  ║  Step 5 of 5   ● ● ● ● ●                                                                 ║                   ║
║  ║──────────────────────────────────────────────────────────────────────────────────────────────║                   ║
║  ║  Navigation                                                                                ║                   ║
║  ║                                                                                            ║                   ║
║  ║  Use the toolbar at the bottom for quick actions. Arrow keys navigate, Enter activates.    ║                   ║
║  ║  Press ? anytime for keyboard shortcuts.                                                   ║                   ║
║  ║                                                                                            ║                   ║
║  ║──────────────────────────────────────────────────────────────────────────────────────────────║                   ║
║  ║  [Enter] Start                                                              [Esc] Skip Tour║                   ║
║  ╚══════════════════════════════════════════════════════════════════════════════════════════════╝                   ║
║ [n] New Ship  ●  [d] Directive  ●  [r] Roster  ●  [i] Inbox  ●  [m] Monitor  ●  [s] Settings  ●  [?] Help  ● [q] Q║
╚══════════════════════════════════════════════════════════════════════════════════════════════════════════════════════╝
```

### Annotation Notes

- The **overlay background** uses `░` (light shade) to represent the dimmed Fleet Overview at 20% opacity. In implementation, the actual Fleet Overview content renders but with reduced opacity via lipgloss styling.
- The **highlighted panel** (e.g., Ship List in Step 1) uses `lipgloss.DoubleBorder()` with `gold` (`#FFAA00`) border color and a glow effect (pulsing box-shadow equivalent via `alert_pulse`-style spring animation).
- The **tooltip card** uses `lipgloss.DoubleBorder()` with `butterscotch` (`#FF9966`) border color. It is anchored to the right of the highlighted panel, with a `◀──` pointer arrow connecting them.
- The **pointer arrow** (`◀──`) renders in `butterscotch` to visually connect the tooltip to its target panel.
- Step indicator uses `●` (`green_ok` `#33FF33`) for completed/current steps and `○` (`galaxy_gray` `#52526A`) for pending steps.
- On the **last step** (Step 5), the navigation changes from `[Enter] Next` to `[Enter] Start` to indicate tour completion.
- The **header** and **toolbar** of the Fleet Overview remain visible (dimmed) to provide spatial context during the tour.
- In Step 5, the toolbar itself is the highlighted panel, so it renders at full opacity with a `gold` double border while everything above is dimmed.

## Color Annotations

| Region                      | Foreground Token    | Background Token   | Style       |
|-----------------------------|---------------------|--------------------|-------------|
| Dimmed background `░`       | `galaxy_gray`       | --                 | Faint       |
| Highlighted panel border    | `gold`              | --                 | Double      |
| Highlighted panel title     | `butterscotch`      | --                 | Bold        |
| Ship name (selected) `▸`   | `butterscotch`      | --                 | Bold        |
| Ship name (unselected)      | `space_white`       | --                 | --          |
| Ship status text            | `galaxy_gray`       | --                 | Faint       |
| Pointer arrow `◀──`        | `butterscotch`      | --                 | --          |
| Tooltip border              | `butterscotch`      | --                 | Double      |
| Step indicator "Step N of 5"| `butterscotch`      | --                 | Bold        |
| Step dot (active) `●`      | `green_ok`          | --                 | --          |
| Step dot (pending) `○`     | `galaxy_gray`       | --                 | --          |
| Separator `──────`          | `galaxy_gray`       | --                 | --          |
| Tooltip title               | `butterscotch`      | --                 | Bold        |
| Tooltip description text    | `space_white`       | --                 | --          |
| Navigation `[Enter]`        | `space_white`       | `dark_blue`        | Bold        |
| Navigation label "Next"     | `light_gray`        | --                 | --          |
| Navigation `[Esc]`          | `space_white`       | `dark_blue`        | Bold        |
| Navigation label "Skip Tour"| `light_gray`        | --                 | --          |
| Header "FLEET COMMAND"      | `gold`              | --                 | Bold        |
| Header metrics              | `space_white`       | --                 | --          |
| Toolbar shortcut keys       | `butterscotch`      | --                 | Bold        |
| Toolbar labels              | `space_white`       | --                 | --          |
| Toolbar separators `●`     | `galaxy_gray`       | --                 | --          |

## Animation Notes

| Event                | Animation Preset    | Spring Parameters              | Feel                                    |
|----------------------|---------------------|---------------------------------|-----------------------------------------|
| Tour opens           | `modal_open`        | freq=6.0, ratio=0.8            | Gentle bounce on tooltip arrival        |
| Step transition      | `view_switch`       | freq=8.0, ratio=1.0            | Snappy slide to next panel anchor       |
| Tour closes          | `modal_close`       | freq=10.0, ratio=1.0           | Fast, crisp dismissal                   |
| Panel highlight glow | `alert_pulse`       | freq=4.0, ratio=0.3            | Pulsing gold border on highlighted panel|

## Compact Variant (80x24 minimum)

In compact mode (`<120 cols`), the tooltip renders centered below the highlighted region instead of anchored to the right. Text wraps to narrower width.

```
╭────────────────────────────────────────────────────────────────────────────────╮
│  FLEET COMMAND   Ships: 6  Launched: 3  Msgs: [4]  Health: ● OK  Done: 67%   │
╰────────────────────────────────────────────────────────────────────────────────╯
░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░
╔═ YOUR FLEET ═════════════════════════════════════════════════════════════════╗
║  ▸ USS ENTERPRISE                                                          ║
║    Active • 5 missions • 3 agents                                          ║
║    USS VOYAGER                                                              ║
║    Idle • 4 missions • 2 agents                                            ║
║    USS DEFIANT                                                              ║
║    Active • 3 missions • 4 agents                                          ║
╚══════════════════════════════════════════════════════════════════════════════╝
╔══════════════════════════════════════════════════════════════════════════════╗
║  Step 1 of 5   ● ○ ○ ○ ○                                                  ║
║────────────────────────────────────────────────────────────────────────────║
║  Your Fleet                                                                ║
║                                                                            ║
║  This is your fleet -- each ship represents a project commission           ║
║  with its own crew and missions. Press Enter to drill into any ship.       ║
║                                                                            ║
║────────────────────────────────────────────────────────────────────────────║
║  [Enter] Next                                               [Esc] Skip Tour║
╚══════════════════════════════════════════════════════════════════════════════╝
╭────────────────────────────────────────────────────────────────────────────────╮
│ [n] New  ● [d] Dir  ● [r] Roster  ● [i] Inbox  ● [s] Set  ● [?] Help  ● [q]│
╰────────────────────────────────────────────────────────────────────────────────╯
```

### Compact Variant Notes

- Tooltip renders directly below the highlighted panel instead of anchored to the right. No pointer arrow.
- Header collapses to a single line with abbreviated labels.
- Toolbar labels are shortened to fit 80 columns.
- Highlighted panel takes full width.
- The tooltip card also takes full width, centered below.
- Scrolling may be needed if the highlighted panel is large and the tooltip pushes content below the fold.
