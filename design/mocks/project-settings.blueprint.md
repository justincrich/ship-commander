# Project Settings -- ASCII Blueprint

> Source: `project-settings.mock.html` | Spec: `prompts/project-settings.spec.json`
> Terminal: 120 cols x 30 rows (standard breakpoint)

## Layout Structure

The project settings screen uses a three-row vertical stack:

```
SettingsHeader (100%, 2 rows)
──────────────────────────────────────
SettingsTabs (100%, fill height)
──────────────────────────────────────
NavigableToolbar (100%, 1 row)
```

- **Header**: 2 rows. Row 1 is the PROJECT SETTINGS title in bold butterscotch. Row 2 shows project path and config file location in light_gray labels with space_white values.
- **Settings content**: Fill remaining height. Tabbed panel with 4 sections navigable with Left/Right or [1]-[4]: Tab 1 Verification Gates, Tab 2 Crew Defaults, Tab 3 Fleet Defaults, Tab 4 Export/Import.
- **Toolbar**: 1 row. NavigableToolbar with tab number keys and shortcuts.

## Component Mapping (from components.yaml)

| Visual Area | Component | Charm Base | Key Props |
|---|---|---|---|
| Full screen container | **AppShell** (#1) | `lipgloss.JoinVertical` | header, body, footer |
| Settings header | Custom `lipgloss.NewStyle()` | `lipgloss.NewStyle()` | "PROJECT SETTINGS" bold butterscotch, path/config info |
| Tab bar | Custom `lipgloss.JoinHorizontal` | `lipgloss.JoinHorizontal` | `▸` active indicator (butterscotch), inactive (galaxy_gray), `│` separators |
| Tab content area | **FocusablePanel** (#4) | `lipgloss.NewStyle().Border(lipgloss.RoundedBorder())` | scrollable tab content |
| Gate item cards | **FocusablePanel** (#4) nested | `lipgloss.NewStyle().Border(lipgloss.RoundedBorder())` | ✓/⊘ icon, gate label, command, edit hint |
| Gate edit form | `huh.Input` | `huh.Input` | command text editing, validation |
| Gate enable toggle | `huh.Confirm` | `huh.Confirm` | enable/disable gate |
| Crew defaults form | `huh.Form` wrapping `huh.Select` + `huh.Input` | `huh.Form` | model select, harness select, default role |
| Fleet defaults form | `huh.Form` wrapping `huh.Input` + `huh.Select` | `huh.Form` | default ship config, naming conventions |
| Export/import actions | `huh.Confirm` + `huh.FilePicker` | `huh.Confirm` | export/import confirmation, file path selection |
| Bottom toolbar | **NavigableToolbar** | `lipgloss.JoinHorizontal` | [1] Gates [2] Crew [3] Fleet [4] Export [?] Help [Esc] Fleet |

## Token Reference

| Element                | Token(s)                                          | Hex         |
|------------------------|---------------------------------------------------|-------------|
| Header title           | `butterscotch` (bold)                             | `#FF9966`   |
| Header info label      | `blue`                                            | `#9999CC`   |
| Header info value      | `space_white`                                     | `#F5F6FA`   |
| Header divider         | `galaxy_gray`                                     | `#52526A`   |
| Tab active             | `butterscotch`                                    | `#FF9966`   |
| Tab active indicator   | `▸` prefix in `butterscotch`                      | `#FF9966`   |
| Tab inactive           | `galaxy_gray`                                     | `#52526A`   |
| Tab separator          | `galaxy_gray` (`│`)                               | `#52526A`   |
| Tab divider line       | `galaxy_gray`                                     | `#52526A`   |
| Gate border            | `lipgloss.RoundedBorder()`, `galaxy_gray`         | `#52526A`   |
| Gate border (focused)  | `lipgloss.RoundedBorder()`, `moonlit_violet`      | `#9966FF`   |
| Gate icon enabled      | `✓` in `green_ok`                                 | `#33FF33`   |
| Gate icon disabled     | `⊘` in `galaxy_gray`                              | `#52526A`   |
| Gate label             | `blue`                                            | `#9999CC`   |
| Gate command           | `space_white`                                     | `#F5F6FA`   |
| Gate command disabled  | `galaxy_gray` (faint, italic)                     | `#52526A`   |
| Gate edit hint         | `galaxy_gray`                                     | `#52526A`   |
| Toolbar key            | `butterscotch` (bold)                             | `#FF9966`   |
| Toolbar label          | `space_white`                                     | `#F5F6FA`   |
| Toolbar separator      | `galaxy_gray`                                     | `#52526A`   |
| Toolbar help key       | `moonlit_violet` (bold)                           | `#9966FF`   |
| Toolbar Esc label      | `light_gray`                                      | `#CCCCCC`   |

## ASCII Blueprint (120x30 standard)

```
╭──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╮
│  PROJECT SETTINGS                                                                                                  │
│  Path: ~/Projects/my-app                              Config: ~/.sc3/config.json                                   │
╰──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╯
╭──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╮
│  ▸ Gates │  Crew │  Fleet │  Export                                                                                │
│──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────│
│                                                                                                                    │
│  ╭────────────────────────────────────────────────────────────────────────────────────────────────────────────────╮  │
│  │  ✓  Lint:          npm run lint                                                       [Enter to edit]        │  │
│  ╰────────────────────────────────────────────────────────────────────────────────────────────────────────────────╯  │
│                                                                                                                    │
│  ╭────────────────────────────────────────────────────────────────────────────────────────────────────────────────╮  │
│  │  ✓  Typecheck:     npx tsc --noEmit                                                   [Enter to edit]        │  │
│  ╰────────────────────────────────────────────────────────────────────────────────────────────────────────────────╯  │
│                                                                                                                    │
│  ╭────────────────────────────────────────────────────────────────────────────────────────────────────────────────╮  │
│  │  ✓  Test:          npm test -- --watchAll=false                                        [Enter to edit]        │  │
│  ╰────────────────────────────────────────────────────────────────────────────────────────────────────────────────╯  │
│                                                                                                                    │
│  ╭────────────────────────────────────────────────────────────────────────────────────────────────────────────────╮  │
│  │  ⊘  Build:         (disabled)                                                         [Enter to edit]        │  │
│  ╰────────────────────────────────────────────────────────────────────────────────────────────────────────────────╯  │
│                                                                                                                    │
╰──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╯
╭──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╮
│ [1] Gates  ●  [2] Crew  ●  [3] Fleet  ●  [4] Export  ●  [?] Help  ●  [Esc] Fleet                                  │
╰──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╯
```

### Annotation Notes

- The **header title** (`PROJECT SETTINGS`) renders in bold `butterscotch` (`#FF9966`).
- The **header info labels** (`Path:`, `Config:`) use `blue` (`#9999CC`). The values use `space_white` (`#F5F6FA`).
- The **active tab** (`▸ Gates`) renders in `butterscotch` (`#FF9966`) with the `▸` prefix indicator. Inactive tabs (`Crew`, `Fleet`, `Export`) render in `galaxy_gray` (`#52526A`).
- Tab separators `│` use `galaxy_gray` (`#52526A`).
- The tab content divider line uses `galaxy_gray` box-drawing `─` characters.
- Each **gate item** is a rounded-border card. The first gate (Lint) uses `moonlit_violet` (`#9966FF`) border to indicate focus. Others use `galaxy_gray` (`#52526A`).
- **Enabled gate icons** (`✓`) render in `green_ok` (`#33FF33`). The **disabled gate icon** (`⊘`) renders in `galaxy_gray` (`#52526A`).
- **Gate labels** (`Lint:`, `Typecheck:`, `Test:`, `Build:`) render in `blue` (`#9999CC`).
- **Gate commands** render in `space_white` (`#F5F6FA`). The disabled gate command text `(disabled)` renders in `galaxy_gray` faint.
- **Edit hints** (`[Enter to edit]`) render in `galaxy_gray` (`#52526A`).
- Toolbar `●` separators use `galaxy_gray`. Shortcut keys like `[1]` use `butterscotch`. Labels use `space_white`. The `[?]` key uses `moonlit_violet`. `[Esc]` label uses `light_gray`.

## Color Annotations

| Region                      | Foreground Token    | Background Token   | Style       |
|-----------------------------|---------------------|--------------------|-------------|
| "PROJECT SETTINGS"          | `butterscotch`      | --                 | Bold        |
| Header info labels          | `blue`              | --                 | --          |
| Header info values          | `space_white`       | --                 | --          |
| Header divider line         | `galaxy_gray`       | --                 | --          |
| Tab active `▸ Gates`        | `butterscotch`      | --                 | --          |
| Tab inactive                | `galaxy_gray`       | --                 | --          |
| Tab separator `│`           | `galaxy_gray`       | --                 | --          |
| Tab content divider         | `galaxy_gray`       | --                 | --          |
| Gate border (default)       | `galaxy_gray`       | --                 | Rounded     |
| Gate border (focused)       | `moonlit_violet`    | --                 | Rounded     |
| Gate icon `✓` (enabled)     | `green_ok`          | --                 | --          |
| Gate icon `⊘` (disabled)    | `galaxy_gray`       | --                 | --          |
| Gate label                  | `blue`              | --                 | --          |
| Gate command (enabled)      | `space_white`       | --                 | --          |
| Gate command (disabled)     | `galaxy_gray`       | --                 | Faint       |
| Gate edit hint              | `galaxy_gray`       | --                 | --          |
| Gate item bg (enabled)      | --                  | `blue` (5% alpha)  | --          |
| Gate item bg (disabled)     | --                  | `galaxy_gray` (5%) | --          |
| Toolbar shortcut keys       | `butterscotch`      | --                 | Bold        |
| Toolbar labels              | `space_white`       | --                 | --          |
| Toolbar separators `●`      | `galaxy_gray`       | --                 | --          |
| Toolbar `[?]`               | `moonlit_violet`    | --                 | Bold        |
| Toolbar `[Esc]` label       | `light_gray`        | --                 | --          |

## Compact Variant (80x24 minimum)

In compact mode (`<120 cols`), tab headers are abbreviated. Form fields stack vertically. Gate list shows name and status only, hiding the edit hint. Toolbar labels are shortened.

```
╭──────────────────────────────────────────────────────────────────────────────────╮
│  PROJECT SETTINGS   Path: ~/Projects/my-app   Config: ~/.sc3/config.json       │
╰──────────────────────────────────────────────────────────────────────────────────╯
╭──────────────────────────────────────────────────────────────────────────────────╮
│  ▸ Gates │  Crew │  Fleet │  Export                                             │
│────────────────────────────────────────────────────────────────────────────────  │
│                                                                                 │
│  ╭──────────────────────────────────────────────────────────────────────────╮    │
│  │  ✓  Lint:          npm run lint                                        │    │
│  ╰──────────────────────────────────────────────────────────────────────────╯    │
│  ╭──────────────────────────────────────────────────────────────────────────╮    │
│  │  ✓  Typecheck:     npx tsc --noEmit                                    │    │
│  ╰──────────────────────────────────────────────────────────────────────────╯    │
│  ╭──────────────────────────────────────────────────────────────────────────╮    │
│  │  ✓  Test:          npm test -- --watchAll=false                         │    │
│  ╰──────────────────────────────────────────────────────────────────────────╯    │
│  ╭──────────────────────────────────────────────────────────────────────────╮    │
│  │  ⊘  Build:         (disabled)                                          │    │
│  ╰──────────────────────────────────────────────────────────────────────────╯    │
│                                                                                 │
╰──────────────────────────────────────────────────────────────────────────────────╯
╭──────────────────────────────────────────────────────────────────────────────────╮
│ [1] Gates  ● [2] Crew  ● [3] Fleet  ● [4] Export  ● [?] Help  ● [Esc] Fleet   │
╰──────────────────────────────────────────────────────────────────────────────────╯
```

### Compact Variant Notes

- Header collapses from 2 content rows to 1. Title, path, and config render inline on a single line.
- Tab headers retain full names since they are already short.
- Gate items hide the `[Enter to edit]` hint to save horizontal space.
- Gate cards still use rounded borders with the same color rules (focused in `moonlit_violet`, default in `galaxy_gray`).
- Toolbar labels remain unabbreviated since they are already concise.
- The focused gate card still uses `moonlit_violet` border for focus indication.
