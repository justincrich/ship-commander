# Directive Editor -- ASCII Blueprint

> Source: `directive-editor.mock.html` | Spec: `views.yaml` (directive-editor)
> Terminal: 120 cols x 30 rows (standard breakpoint)

## Layout Structure

The directive editor uses a four-row vertical stack:

```
EditorHeader (100%, 2 rows)
──────────────────────────────────────
FileInput (50%)  +  Preview (50%)
   (60% height, scrollable)
──────────────────────────────────────
StructureSummary (100%, 25% height)
──────────────────────────────────────
NavigableToolbar (100%, 1 row)
```

- **Header**: 2 rows. Row 1 is the "DIRECTIVE EDITOR" title in bold purple with mode indicator ("New" or "Editing: <title>"). Row 2 shows file path if imported.
- **Content**: Fill 60% height. File input (50% width, left) shows file path input at top and raw Markdown content in a scrollable viewport. Preview (50% width, right) shows live Glamour-rendered preview of the directive content.
- **Structure Summary**: 25% height. Parsed directive structure table showing extracted use cases (ID, title, AC count), total AC count, scope summary, and any parsing warnings.
- **Toolbar**: 1 row. NavigableToolbar with action buttons and inline ship assignment selector.

## Component Mapping (from components.yaml)

| Visual Area | Component | Charm Base | Key Props |
|---|---|---|---|
| Full screen container | **AppShell** (#1) | `lipgloss.JoinVertical` | header, body, footer |
| Editor header | Custom `lipgloss.NewStyle()` | `lipgloss.NewStyle()` | "DIRECTIVE EDITOR" bold purple, mode |
| File + Preview layout | **PanelGrid** (#2) | `lipgloss.JoinHorizontal` | ratios: `[0.5, 0.5]`, compact_threshold: 120 |
| File input panel | **FocusablePanel** (#4) | `lipgloss.NewStyle().Border(lipgloss.RoundedBorder())` | file path + raw markdown |
| File path input | `huh.Input` / `bubbles.filepicker` | `huh.Input` | path, placeholder, validation |
| Raw markdown viewport | `bubbles.viewport` | `bubbles.viewport` | scrollable raw PRD content |
| Preview panel | **FocusablePanel** (#4) + Glamour | `bubbles.viewport` | live Glamour-rendered preview |
| Structure summary | **FocusablePanel** (#4) wrapping `bubbles.table` | `bubbles.table` | UC ID, title, AC count, warnings |
| Ship assignment | `huh.Select` | `huh.Select` | inline, available DOCKED ships |
| Bottom toolbar | **NavigableToolbar** | `lipgloss.JoinHorizontal` | [Enter] Save [Tab] Cycle [?] Help [Esc] Cancel |

## Token Reference

| Element              | Token(s)                                          | Hex         |
|----------------------|---------------------------------------------------|-------------|
| Header title         | `purple` (bold)                                   | `#CC99CC`   |
| Mode indicator       | `blue`                                            | `#9999CC`   |
| File path text       | `light_gray` (faint)                              | `#CCCCCC`   |
| File loaded badge    | `green_ok`                                        | `#33FF33`   |
| Panel border default | `lipgloss.RoundedBorder()`, `galaxy_gray`         | `#52526A`   |
| Panel border focused | `lipgloss.RoundedBorder()`, `moonlit_violet`      | `#9966FF`   |
| Form label           | `butterscotch` (bold)                             | `#FF9966`   |
| Form hint            | `galaxy_gray`                                     | `#52526A`   |
| Input text           | `space_white`                                     | `#F5F6FA`   |
| Input border default | `galaxy_gray`                                     | `#52526A`   |
| Input border focused | `moonlit_violet`                                  | `#9966FF`   |
| Markdown raw text    | `space_white`                                     | `#F5F6FA`   |
| Markdown heading     | `butterscotch` (bold)                             | `#FF9966`   |
| Preview rendered     | Glamour theme (space_white base)                  | `#F5F6FA`   |
| Structure table head | `gold` (bold)                                     | `#FFAA00`   |
| Structure UC ID      | `butterscotch`                                    | `#FF9966`   |
| Structure UC title   | `space_white`                                     | `#F5F6FA`   |
| Structure AC count   | `blue`                                            | `#9999CC`   |
| Structure status ok  | `green_ok`                                        | `#33FF33`   |
| Structure warning    | `yellow_caution`                                  | `#FFCC00`   |
| Ship tag             | `butterscotch` (fg), `butterscotch` 20% (bg)      | `#FF9966`   |
| Char counter         | `galaxy_gray`                                     | `#52526A`   |
| Toolbar key          | `moonlit_violet` (bold)                           | `#9966FF`   |
| Toolbar label        | `butterscotch`                                    | `#FF9966`   |
| Toolbar separator    | `galaxy_gray`                                     | `#52526A`   |

## ASCII Blueprint (120x30 standard)

```
╭──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╮
│  DIRECTIVE EDITOR                                                                        Mode: New Directive       │
│  ./docs/api-documentation-prd.md                                                         2,340 chars  [+] loaded   │
╰──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╯
╭─ Raw Markdown ─────────────────────────────────────────────╮╭─ Glamour Preview ──────────────────────────────────────╮
│ DIRECTIVE TITLE                                            ││                                                       │
│ A short name for this directive                            ││  API Documentation Overhaul                           │
│ ┌────────────────────────────────────────────────────────┐  ││                                                       │
│ │ api-documentation                                     │  ││  Problem Statement                                    │
│ └────────────────────────────────────────────────────────┘  ││  Our API documentation is outdated, incomplete, and   │
│                                                            ││  inconsistent across endpoints. Developers report     │
│ PRD SOURCE  [+] loaded                                     ││  spending 2-3x longer integrating than expected.      │
│ ┌────────────────────────────────────────────────────────┐  ││                                                       │
│ │ # API Documentation Overhaul                          │  ││  Success Criteria                                     │
│ │                                                       │  ││  - AC1: All REST endpoints documented with request/   │
│ │ ## Problem Statement                                  │  ││    response examples                                  │
│ │ Our API documentation is outdated, incomplete, and    │  ││  - AC2: OpenAPI 3.0 spec generated from source code   │
│ │ inconsistent across endpoints. Developers report      │  ││  - AC3: Interactive API playground deployed           │
│ │ spending 2-3x longer integrating than expected.       │  ││  - AC4: Authentication flows documented with sequence │
│ │                                                       │  ││    diagrams                                           │
│ │ ## Success Criteria                                   │  ││                                                       │
│ │ - AC1: All REST endpoints documented with request/    │  ││                                                       │
│ │   response examples                                   │  ││                                                       │
│ │ - AC2: OpenAPI 3.0 spec generated from source code    │  ││                                                       │
│ │ - AC3: Interactive API playground deployed             │  ││                                                       │
│ └────────────────────────────────────────────────────────┘  ││                                                       │
│                                          2,340 / 10,000 ch ││                                                       │
╰────────────────────────────────────────────────────────────╯╰───────────────────────────────────────────────────────╯
╭─ Parsed Structure ─────────────────────────────────────────────────────────────────────────────────────────────────╮
│  ID     Title                                          ACs   Status                                               │
│  UC-1   REST endpoint documentation                    1     ✓ auto-detected                                      │
│  UC-2   OpenAPI spec generation                        1     ✓ auto-detected                                      │
│  UC-3   API playground deployment                      1     ✓ auto-detected                                      │
│  UC-4   Authentication flow documentation              1     ⚠ needs review                                       │
│                                                                                                                    │
│  Total: 4 use cases  ●  4 acceptance criteria  ●  Ship: USS Api-Documentation (auto-generated)                    │
╰────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╯
╭──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╮
│ [Enter] Save & Assign  ●  [Tab] Cycle  ●  [?] Help  ●  [Esc] Cancel                                               │
╰──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╯
```

### Annotation Notes

- The **header title** ("DIRECTIVE EDITOR") renders in bold `purple` (`#CC99CC`). The mode indicator ("New Directive") uses `blue` (`#9999CC`).
- The **file path** (`./docs/api-documentation-prd.md`) renders in `light_gray` (`#CCCCCC`) with faint style. The `[+] loaded` badge uses `green_ok` (`#33FF33`).
- The **Raw Markdown** panel label renders inline on the top border in `galaxy_gray`. When focused, the border switches to `moonlit_violet` (`#9966FF`).
- The **Glamour Preview** panel label renders inline on the top border in `galaxy_gray`. Preview content uses Glamour markdown rendering with headings in `butterscotch` bold.
- Form labels ("DIRECTIVE TITLE", "PRD SOURCE") render in `butterscotch` (`#FF9966`) bold. Hint text uses `galaxy_gray` (`#52526A`).
- Text input borders use `galaxy_gray` by default and `moonlit_violet` when focused. Input background tints with `moonlit_violet` at 10% opacity when focused.
- The **Parsed Structure** table header row ("ID", "Title", "ACs", "Status") uses `gold` (`#FFAA00`) bold.
- Use case IDs ("UC-1", etc.) use `butterscotch`. Titles use `space_white`. AC counts use `blue`.
- Status `✓ auto-detected` uses `green_ok` (`#33FF33`). Status `⚠ needs review` uses `yellow_caution` (`#FFCC00`).
- The summary line uses `●` separators in `galaxy_gray`. Ship name ("USS Api-Documentation") renders as a tag with `butterscotch` foreground and 20% opacity background.
- Toolbar shortcut keys like `[Enter]` use `moonlit_violet` (`#9966FF`) bold. Labels use `butterscotch` (`#FF9966`). Separators `●` use `galaxy_gray`.
- Character counter ("2,340 / 10,000 ch") uses `galaxy_gray` (`#52526A`) and is right-aligned below the raw content area.

## Color Annotations

| Region                      | Foreground Token    | Background Token   | Style       |
|-----------------------------|---------------------|--------------------|-------------|
| "DIRECTIVE EDITOR"          | `purple`            | --                 | Bold        |
| Mode indicator              | `blue`              | --                 | --          |
| File path                   | `light_gray`        | --                 | Faint       |
| `[+] loaded` badge          | `green_ok`          | --                 | --          |
| Char count                  | `galaxy_gray`       | --                 | --          |
| Panel border (default)      | `galaxy_gray`       | --                 | Rounded     |
| Panel border (focused)      | `moonlit_violet`    | --                 | Rounded     |
| Form labels                 | `butterscotch`      | --                 | Bold        |
| Form hints                  | `galaxy_gray`       | --                 | --          |
| Input text                  | `space_white`       | --                 | --          |
| Input border (default)      | `galaxy_gray`       | --                 | --          |
| Input border (focused)      | `moonlit_violet`    | `moonlit_violet`10%| --          |
| Raw markdown content        | `space_white`       | --                 | --          |
| Raw markdown headings       | `butterscotch`      | --                 | Bold        |
| Glamour headings            | `butterscotch`      | --                 | Bold        |
| Glamour body                | `space_white`       | --                 | --          |
| Glamour list markers        | `blue`              | --                 | --          |
| Structure table header      | `gold`              | --                 | Bold        |
| Use case IDs                | `butterscotch`      | --                 | --          |
| Use case titles             | `space_white`       | --                 | --          |
| AC counts                   | `blue`              | --                 | --          |
| Status `✓`                  | `green_ok`          | --                 | --          |
| Status `⚠`                  | `yellow_caution`    | --                 | --          |
| Ship tag                    | `butterscotch`      | `butterscotch` 20% | Bold        |
| Summary `●` separators      | `galaxy_gray`       | --                 | --          |
| Toolbar keys `[Enter]`      | `moonlit_violet`    | `moonlit_violet`20%| Bold        |
| Toolbar labels              | `butterscotch`      | --                 | --          |
| Toolbar separators `●`      | `galaxy_gray`       | --                 | --          |

## Compact Variant (80x24 minimum)

In compact mode (`<120 cols`), the preview panel is hidden. The raw markdown editor takes full width. Pressing Tab toggles between raw and preview views. Structure summary is scrollable within a reduced height.

```
╭──────────────────────────────────────────────────────────────────────────────────╮
│  DIRECTIVE EDITOR                                          Mode: New Directive  │
╰──────────────────────────────────────────────────────────────────────────────────╯
╭─ Raw Markdown ──────────────────────────────────────────────────────────────────╮
│ DIRECTIVE TITLE                                                                │
│ ┌──────────────────────────────────────────────────────────────────────────┐    │
│ │ api-documentation                                                      │    │
│ └──────────────────────────────────────────────────────────────────────────┘    │
│ PRD SOURCE  [+] loaded  ./docs/api-documentation-prd.md                        │
│ ┌──────────────────────────────────────────────────────────────────────────┐    │
│ │ # API Documentation Overhaul                                           │    │
│ │                                                                        │    │
│ │ ## Problem Statement                                                   │    │
│ │ Our API documentation is outdated, incomplete, and                     │    │
│ │ inconsistent across endpoints. Developers report                       │    │
│ │ spending 2-3x longer integrating than expected.                        │    │
│ │                                                                        │    │
│ │ ## Success Criteria                                                    │    │
│ │ - AC1: All REST endpoints documented                                   │    │
│ │ - AC2: OpenAPI 3.0 spec generated                                      │    │
│ └──────────────────────────────────────────────────────────────────────────┘    │
│                                                        2,340 / 10,000 ch       │
╰────────────────────────────────────────────────────────────────────────────────╯
╭─ Structure ────────────────────────────────────────────────────────────────────╮
│  UC-1 REST endpoint documentation  1 AC  ✓  ●  UC-2 OpenAPI spec  1 AC  ✓    │
│  Total: 4 UCs  ●  4 ACs  ●  Ship: USS Api-Documentation                      │
╰────────────────────────────────────────────────────────────────────────────────╯
╭──────────────────────────────────────────────────────────────────────────────────╮
│ [Enter] Save  ● [Tab] Cycle  ● [?] Help  ● [Esc] Cancel                        │
╰──────────────────────────────────────────────────────────────────────────────────╯
```

### Compact Variant Notes

- Preview panel is completely hidden. Tab toggles between raw markdown view and Glamour preview in the same panel space.
- Header collapses from 2 content rows to 1. File path moves inline with the PRD SOURCE label.
- Structure summary collapses to inline format. Use cases shown in a condensed single-line format with `●` separators.
- Toolbar labels are shortened: "Save" for "Save & Assign".
- The focused panel still uses `moonlit_violet` border for focus indication.
- Scrolling is required for longer PRD content since the visible textarea area is reduced.
