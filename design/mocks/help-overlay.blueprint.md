# Help Overlay -- ASCII Blueprint

> Source: `help-overlay.mock.html` | Spec: `prompts/help-overlay.spec.json`
> Terminal: 120 cols x 30 rows (standard breakpoint)

## Layout Structure

The help overlay is a centered modal that floats above the current view with a dimmed background:

```
[Dimmed Background -- underlying view at 15% opacity]
       ╔══════════════════════════════════════════════╗
       ║  HelpTitle (100%, 2 rows)                    ║
       ║──────────────────────────────────────────────║
       ║  KeybindingSections (100%, auto height)      ║
       ║    GLOBAL                                    ║
       ║    NAVIGATION                                ║
       ║    <CONTEXT-SPECIFIC>                        ║
       ║──────────────────────────────────────────────║
       ║  DismissHint (100%, 1 row)                   ║
       ╚══════════════════════════════════════════════╝
```

- **Overlay**: 70% terminal width, auto height, centered horizontally and vertically. Background is `black` at ~92% opacity. Uses `lipgloss.DoubleBorder()` in `ice` (`#99CCFF`).
- **Title**: 2 rows. Row 1 is "KEYBOARD SHORTCUTS" in bold `butterscotch`. Row 2 is "Context: \<active view name\>" in dim `light_gray`.
- **Sections**: Variable height. Keybindings grouped by category. Section headers in bold `ice`. Keys in bold `butterscotch`. Descriptions in `space_white`. Sections are: GLOBAL (always), NAVIGATION (always), then context-specific section (varies by active view).
- **Dismiss hint**: 1 row. "Press ? or Escape to close" centered in dim `light_gray`.
- **Animation**: Opens with `modal_open` spring (freq=6.0, ratio=0.8, gentle bounce). Closes with `modal_close` spring (freq=10.0, ratio=1.0, crisp dismissal).

## Component Mapping (from components.yaml)

| Visual Area | Component | Charm Base | Key Props |
|---|---|---|---|
| Dimmed background | `lipgloss.NewStyle()` | `lipgloss.NewStyle()` | underlying view at ~15% opacity |
| Overlay container | **ModalOverlay** (#5) | `lipgloss.Place` + `lipgloss.DoubleBorder()` | 70% width centered, ice border, `modal_open`/`modal_close` spring |
| Title section | Custom `lipgloss.NewStyle()` | `lipgloss.NewStyle()` | "KEYBOARD SHORTCUTS" bold butterscotch, context subtitle |
| Keybinding sections | `bubbles.help` + `bubbles.key` | `bubbles.help` | GLOBAL, NAVIGATION, context-specific sections |
| Key bindings | `bubbles.key.Binding` | `bubbles.key` | key in butterscotch bold, description in space_white |
| Dismiss hint | Custom `lipgloss.NewStyle()` | `lipgloss.NewStyle()` | "Press ? or Escape to close" centered, light_gray faint |

## Token Reference

| Element              | Token(s)                                          | Hex         |
|----------------------|---------------------------------------------------|-------------|
| Overlay border       | `lipgloss.DoubleBorder()`, `ice`                  | `#99CCFF`   |
| Overlay background   | `black` (92% opacity)                             | `#000000`   |
| Title text           | `butterscotch` (bold)                             | `#FF9966`   |
| Context subtitle     | `light_gray` (faint)                              | `#CCCCCC`   |
| Section headers      | `ice` (bold)                                      | `#99CCFF`   |
| Shortcut keys        | `butterscotch` (bold)                             | `#FF9966`   |
| Key descriptions     | `space_white`                                     | `#F5F6FA`   |
| Dismiss hint         | `light_gray` (faint)                              | `#CCCCCC`   |
| Dimmed background    | underlying view at 15% opacity                    | --          |

## ASCII Blueprint (120x30 standard)

This shows the help overlay opened from the Ship Bridge context. The underlying Ship Bridge view is dimmed behind the overlay.

```
┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄
┄  SHIP BRIDGE - USS ENDEAVOR                              (dimmed background at 15% opacity)                              ┄
┄  STATUS: Active Mission: Deploy Authentication System                                                                    ┄
┄  Current Agent: backend-engineer   Progress: 67%                                                                         ┄
┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄
               ╔══════════════════════════════════════════════════════════════════════════════════╗
               ║                                                                                ║
               ║                              KEYBOARD SHORTCUTS                                ║
               ║                              Context: Ship Bridge                              ║
               ║                                                                                ║
               ║  GLOBAL                                                                        ║
               ║    Tab / Shift+Tab              Cycle panel focus                               ║
               ║    ?                            Toggle this help                                ║
               ║    q                            Quit                                            ║
               ║    Ctrl+C                       Force quit                                      ║
               ║                                                                                ║
               ║  NAVIGATION                                                                    ║
               ║    Enter                        Drill down / Select                             ║
               ║    Escape                       Go back / Cancel                                ║
               ║    Up / Down                    Navigate list items                              ║
               ║    Page Up / Page Down          Scroll viewport                                 ║
               ║                                                                                ║
               ║  SHIP BRIDGE                                                                   ║
               ║    p                            Open Ready Room (plan)                          ║
               ║    l                            Launch ship                                     ║
               ║    a                            Agent detail                                    ║
               ║    h                            Halt mission/agent                              ║
               ║    r                            Retry mission                                   ║
               ║    w                            Wave manager                                    ║
               ║    Space                        Pause/resume                                    ║
               ║                                                                                ║
               ║                           Press ? or Escape to close                           ║
               ║                                                                                ║
               ╚══════════════════════════════════════════════════════════════════════════════════╝
```

### Annotation Notes

- The **overlay border** uses `lipgloss.DoubleBorder()` with double-line characters (`╔ ╗ ╚ ╝ ║ ═`) in `ice` (`#99CCFF`). This distinguishes it from standard panels which use `lipgloss.RoundedBorder()`.
- The **dimmed background** is represented by `┄` dashed lines. In the actual TUI, the underlying view (Ship Bridge, Fleet Overview, etc.) renders at ~15% opacity behind the overlay.
- The **title** "KEYBOARD SHORTCUTS" renders centered in bold `butterscotch` (`#FF9966`).
- The **context subtitle** "Context: Ship Bridge" renders centered in faint `light_gray` (`#CCCCCC`). This dynamically updates based on which view the help was opened from.
- **Section headers** ("GLOBAL", "NAVIGATION", "SHIP BRIDGE") render in bold `ice` (`#99CCFF`) with left alignment and no indent.
- **Shortcut keys** (e.g., `Tab / Shift+Tab`, `p`, `l`) render in bold `butterscotch` (`#FF9966`) with a fixed-width column (~30 chars) for alignment.
- **Descriptions** render in `space_white` (`#F5F6FA`) in the second column.
- The **dismiss hint** "Press ? or Escape to close" renders centered in faint `light_gray` (`#CCCCCC`).
- The context-specific section (third section) changes dynamically. When opened from Fleet Overview, it shows "FLEET OVERVIEW" with shortcuts like `n` (New Ship), `d` (Directive), etc. When opened from Ship Bridge, it shows "SHIP BRIDGE" shortcuts as shown above.
- The overlay is **not pushed to the navigation stack**. It toggles on/off with `?` or `Escape` and does not affect breadcrumb navigation.
- The overlay is **blocked from opening** when another modal is active (AdmiralQuestionModal, ConfirmDialog). The `?` key is simply ignored in that case.

## Color Annotations

| Region                      | Foreground Token    | Background Token   | Style       |
|-----------------------------|---------------------|--------------------|-------------|
| Overlay border `╔╗╚╝║═`    | `ice`               | --                 | Double      |
| Overlay fill                | --                  | `black`            | 92% opacity |
| "KEYBOARD SHORTCUTS"        | `butterscotch`      | --                 | Bold        |
| "Context: Ship Bridge"      | `light_gray`        | --                 | Faint       |
| Section headers             | `ice`               | --                 | Bold        |
| Shortcut keys               | `butterscotch`      | --                 | Bold        |
| Key descriptions            | `space_white`       | --                 | --          |
| "Press ? or Escape to close"| `light_gray`        | --                 | Faint       |
| Dimmed background content   | (original colors)   | --                 | 15% opacity |

## Context Variants

The help overlay adapts its third section based on the active view. Below are the context-specific keybindings for each view.

### Fleet Overview Context

```
               ╔══════════════════════════════════════════════════════════════════════════════════╗
               ║                                                                                ║
               ║                              KEYBOARD SHORTCUTS                                ║
               ║                            Context: Fleet Overview                             ║
               ║                                                                                ║
               ║  GLOBAL                                                                        ║
               ║    Tab / Shift+Tab              Cycle panel focus                               ║
               ║    ?                            Toggle this help                                ║
               ║    q                            Quit                                            ║
               ║    Ctrl+C                       Force quit                                      ║
               ║                                                                                ║
               ║  NAVIGATION                                                                    ║
               ║    Enter                        Drill down / Select                             ║
               ║    Escape                       Go back / Cancel                                ║
               ║    Up / Down                    Navigate list items                              ║
               ║    Page Up / Page Down          Scroll viewport                                 ║
               ║                                                                                ║
               ║  FLEET OVERVIEW                                                                ║
               ║    n                            New ship                                        ║
               ║    d                            Directive                                       ║
               ║    r                            Roster                                          ║
               ║    i                            Inbox                                           ║
               ║    m                            Monitor                                         ║
               ║    s                            Settings                                        ║
               ║                                                                                ║
               ║                           Press ? or Escape to close                           ║
               ║                                                                                ║
               ╚══════════════════════════════════════════════════════════════════════════════════╝
```

## Compact Variant (80x24 minimum)

In compact mode (`<120 cols`), the overlay takes ~85% width. Key column is narrower. Layout remains single-column but more condensed.

```
       ╔══════════════════════════════════════════════════════════════════╗
       ║                                                                ║
       ║                       KEYBOARD SHORTCUTS                       ║
       ║                       Context: Ship Bridge                     ║
       ║                                                                ║
       ║  GLOBAL                                                        ║
       ║    Tab / Shift+Tab        Cycle panel focus                     ║
       ║    ?                      Toggle this help                      ║
       ║    q                      Quit                                  ║
       ║    Ctrl+C                 Force quit                            ║
       ║                                                                ║
       ║  NAVIGATION                                                    ║
       ║    Enter                  Drill down / Select                   ║
       ║    Escape                 Go back / Cancel                      ║
       ║    Up / Down              Navigate list items                   ║
       ║    PgUp / PgDn            Scroll viewport                      ║
       ║                                                                ║
       ║  SHIP BRIDGE                                                   ║
       ║    p  Ready Room    l  Launch    a  Agent                       ║
       ║    h  Halt          r  Retry     w  Waves                       ║
       ║    Space  Pause/resume                                          ║
       ║                                                                ║
       ║                    Press ? or Escape to close                   ║
       ║                                                                ║
       ╚══════════════════════════════════════════════════════════════════╝
```

### Compact Variant Notes

- Overlay width increases to ~85% of terminal to compensate for narrower terminal.
- Key column width reduced from ~30 chars to ~22 chars.
- "Page Up / Page Down" abbreviated to "PgUp / PgDn".
- Context-specific shortcuts may use a condensed multi-column layout (3 shortcuts per row) when there are many bindings, to reduce vertical space.
- GLOBAL and NAVIGATION sections remain in standard two-column layout since they have fewer entries.
- Dimmed background still renders at 15% opacity behind the overlay.
