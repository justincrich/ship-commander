# Confirm Dialog -- ASCII Blueprint

> Source: `confirm-dialog.mock.html` | Spec: `tokens.yaml`
> Terminal: 120 cols x 30 rows (standard breakpoint)

## Layout Structure

The confirm dialog is a centered modal overlay rendered on top of a dimmed background view. Two variants exist: **Halt Mission** (red alert, destructive) and **Shelve Plan** (yellow caution, non-destructive).

```
Background (100%, dimmed to ~15% opacity)
──────────────────────────────────────
Modal Dialog (centered, ~54 cols x 18 rows)
  DoubleBorder (red_alert or yellow_caution)
  Icon + Title
  Description (centered text)
  Detail Box (key-value pairs)
  Action Buttons ([!] Halt / Cancel)
  Hint Line (keyboard shortcuts)
```

- **Background**: The full Ship Bridge view renders at ~15% opacity (faint). All text and borders are dimmed to indicate the modal has focus.
- **Modal**: Centered horizontally and vertically. Uses `lipgloss.DoubleBorder()` with color determined by severity. Padding of 2 on all sides per `modal_dialog` component token.
- **Action Buttons**: Two options side by side. The destructive action is highlighted (selected) by default for Halt; Cancel is highlighted by default for Shelve. Left/Right arrows move selection, Enter confirms, Esc cancels.
- **Hint Line**: Keyboard shortcut help rendered in `galaxy_gray` with key names in `moonlit_violet`.

## Component Mapping (from components.yaml)

| Visual Area | Component | Charm Base | Key Props |
|---|---|---|---|
| Dimmed background | `lipgloss.NewStyle()` | `lipgloss.NewStyle()` | underlying view at ~15% opacity |
| Modal container | **ModalOverlay** (#5) | `lipgloss.Place` + `lipgloss.DoubleBorder()` | centered, red_alert or yellow_caution border, `modal_open`/`modal_close` spring |
| Alert/caution icon | Custom `lipgloss.NewStyle()` | `lipgloss.NewStyle()` | `[!]` red_alert bold or `[~]` yellow_caution bold |
| Modal title | Custom `lipgloss.NewStyle()` | `lipgloss.NewStyle()` | "HALT MISSION?" or "SHELVE PLAN?" centered, severity-colored bold |
| Description text | Custom `lipgloss.NewStyle()` | `lipgloss.NewStyle()` | centered space_white |
| Detail box | **Panel** (#3) | `lipgloss.NewStyle().Border(lipgloss.NormalBorder())` | galaxy_gray border, key-value pairs (blue labels, space_white values) |
| Action buttons | **ConfirmDialog** (#23) | `huh.Confirm` | Left/Right selection, Enter confirm, Esc cancel |
| Hint line | Custom `lipgloss.NewStyle()` | `lipgloss.NewStyle()` | galaxy_gray text, moonlit_violet key names |

## Token Reference

| Element               | Token(s)                                          | Hex         |
|-----------------------|---------------------------------------------------|-------------|
| Modal border (halt)   | `lipgloss.DoubleBorder()`, `red_alert`            | `#FF3333`   |
| Modal border (shelve) | `lipgloss.DoubleBorder()`, `yellow_caution`       | `#FFCC00`   |
| Modal background      | `surface_modal` (`black`)                         | `#000000`   |
| Alert icon `[!]`      | `red_alert` (bold)                                | `#FF3333`   |
| Caution icon `[~]`    | `yellow_caution` (bold)                           | `#FFCC00`   |
| Modal title (halt)    | `red_alert` (bold)                                | `#FF3333`   |
| Modal title (shelve)  | `yellow_caution` (bold)                           | `#FFCC00`   |
| Description text      | `space_white`                                     | `#F5F6FA`   |
| Detail label          | `blue`                                            | `#9999CC`   |
| Detail value          | `space_white`                                     | `#F5F6FA`   |
| Detail value warning  | `yellow_caution`                                  | `#FFCC00`   |
| Detail box border     | `galaxy_gray`                                     | `#52526A`   |
| Button danger (sel)   | `black` (fg), `red_alert` (bg)                    | `#FF3333`   |
| Button danger (unsel) | `red_alert` (fg), `red_alert` 20% (bg)            | `#FF3333`   |
| Button cancel (sel)   | `space_white` (fg), `galaxy_gray` (bg)            | `#52526A`   |
| Button cancel (unsel) | `light_gray` (fg), `galaxy_gray` 20% (bg)         | `#CCCCCC`   |
| Button shelve (sel)   | `black` (fg), `yellow_caution` (bg)               | `#FFCC00`   |
| Button shelve (unsel) | `yellow_caution` (fg), `yellow_caution` 20% (bg)  | `#FFCC00`   |
| Hint text             | `galaxy_gray`                                     | `#52526A`   |
| Hint keys             | `moonlit_violet` (bold)                           | `#9966FF`   |
| Background (dimmed)   | `space_white` at 15% opacity                      | `#F5F6FA`   |

## ASCII Blueprint -- Halt Mission (120x30 standard)

```
┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄
┊ (dimmed background -- Ship Bridge at 15% opacity)                                                                        ┊
┊  ╔═══════════════════════════════════════════════════════════════════════════╗                                             ┊
┊  ║                          SHIP BRIDGE - SS ENTERPRISE                    ║                                             ┊
┊  ╚═══════════════════════════════════════════════════════════════════════════╝                                             ┊
┊  │ CREW ROSTER                                                             │                                             ┊
┊  │  capt-alpha       ▸ ACTIVE    MISSION-01                                │                                             ┊
┊  │  impl-bravo       ▸ ACTIVE    MISSION-03  R                             │                                             ┊
┊  │  impl-charlie     ▸ ACTIVE    MISSION-04  G                             │                                             ┊
┊  │  rev-delta        ⏸ REVIEW    MISSION-05                                │                                             ┊
┊                                                                                                                          ┊
┊                       ╔══════════════════════════════════════════════════╗                                                ┊
┊                       ║                                                  ║                                                ┊
┊                       ║               [!] RED ALERT                      ║                                                ┊
┊                       ║                                                  ║                                                ┊
┊                       ║              HALT MISSION?                       ║                                                ┊
┊                       ║                                                  ║                                                ┊
┊                       ║   This will stop the active agent and return     ║                                                ┊
┊                       ║   the mission to the backlog. The worktree       ║                                                ┊
┊                       ║   will be preserved.                             ║                                                ┊
┊                       ║                                                  ║                                                ┊
┊                       ║   ┌──────────────────────────────────────────┐   ║                                                ┊
┊                       ║   │ Mission:   MISSION-03                    │   ║                                                ┊
┊                       ║   │ Agent:     impl-bravo                    │   ║                                                ┊
┊                       ║   │ Phase:     RED (02:15 elapsed)           │   ║                                                ┊
┊                       ║   │ Worktree:  .worktrees/mission-03/        │   ║                                                ┊
┊                       ║   └──────────────────────────────────────────┘   ║                                                ┊
┊                       ║                                                  ║                                                ┊
┊                       ║          [!] Halt          Cancel                ║                                                ┊
┊                       ║                                                  ║                                                ┊
┊                       ║   Left/Right to select  Enter confirm  Esc cancel║                                                ┊
┊                       ║                                                  ║                                                ┊
┊                       ╚══════════════════════════════════════════════════╝                                                ┊
┊                                                                                                                          ┊
┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄
```

### Annotation Notes

- The **dimmed background** is indicated by `┄` and `┊` dotted borders. In the TUI, the full Ship Bridge view renders at ~15% opacity behind the modal. All background text uses `space_white` with heavy faint styling.
- The **modal dialog** uses `lipgloss.DoubleBorder()` with `red_alert` (`#FF3333`) border color. The `╔ ╗ ╚ ╝ ║ ═` characters are the double border set.
- The **alert icon** `[!] RED ALERT` renders centered in `red_alert` with bold styling.
- The **title** `HALT MISSION?` renders centered in `red_alert` with bold styling.
- The **description** text renders centered in `space_white`.
- The **detail box** uses a thin `galaxy_gray` border (`┌ ┐ └ ┘ │ ─`). Labels are in `blue` (`#9999CC`), values in `space_white`. The Phase value `RED (02:15 elapsed)` uses `yellow_caution` (`#FFCC00`) to indicate warning state.
- The **[!] Halt** button is selected (highlighted) by default: `red_alert` background with `black` foreground. The **Cancel** button is unselected: `galaxy_gray` 20% background with `light_gray` foreground.
- The **hint line** uses `galaxy_gray` for descriptive text and `moonlit_violet` (`#9966FF`) bold for key names.
- Modal animation uses `modal_open` spring (angular_frequency: 6.0, damping_ratio: 0.8) for a gentle bounce entrance.

## ASCII Blueprint -- Shelve Plan Variant

```
                       ╔══════════════════════════════════════════════════╗
                       ║                                                  ║
                       ║               [~] CAUTION                        ║
                       ║                                                  ║
                       ║              SHELVE PLAN?                        ║
                       ║                                                  ║
                       ║   This will save the current mission manifest    ║
                       ║   for later review. The Ready Room specialists   ║
                       ║   will be released.                              ║
                       ║                                                  ║
                       ║   ┌──────────────────────────────────────────┐   ║
                       ║   │ Ship:       SS Enterprise                │   ║
                       ║   │ Directive:  auth-system                  │   ║
                       ║   │ Missions:   10 planned (3 waves)         │   ║
                       ║   └──────────────────────────────────────────┘   ║
                       ║                                                  ║
                       ║           Cancel          [~] Shelve             ║
                       ║                                                  ║
                       ║   Left/Right to select  Enter confirm  Esc cancel║
                       ║                                                  ║
                       ╚══════════════════════════════════════════════════╝
```

### Shelve Variant Notes

- The **modal border** uses `yellow_caution` (`#FFCC00`) instead of `red_alert`.
- The **icon** `[~] CAUTION` and **title** `SHELVE PLAN?` render in `yellow_caution` bold.
- The **detail box** background tints slightly toward `yellow_caution` (5% opacity) instead of `red_alert`.
- The **Cancel** button is selected (highlighted) by default for non-destructive actions: `galaxy_gray` background with `space_white` foreground. The **[~] Shelve** button is unselected: `yellow_caution` 20% background with `yellow_caution` foreground.
- Selection defaults differ by severity: destructive dialogs default to the action button (Halt), non-destructive dialogs default to Cancel.

## Color Annotations

| Region                        | Foreground Token     | Background Token    | Style       |
|-------------------------------|----------------------|---------------------|-------------|
| Modal border (halt)           | `red_alert`          | --                  | Double      |
| Modal border (shelve)         | `yellow_caution`     | --                  | Double      |
| Modal interior                | --                   | `black`             | --          |
| `[!] RED ALERT`               | `red_alert`          | --                  | Bold        |
| `[~] CAUTION`                 | `yellow_caution`     | --                  | Bold        |
| `HALT MISSION?`               | `red_alert`          | --                  | Bold        |
| `SHELVE PLAN?`                | `yellow_caution`     | --                  | Bold        |
| Description text              | `space_white`        | --                  | --          |
| Detail box border             | `galaxy_gray`        | --                  | Normal      |
| Detail labels                 | `blue`               | --                  | --          |
| Detail values                 | `space_white`        | --                  | --          |
| Detail warning value          | `yellow_caution`     | --                  | Bold        |
| `[!] Halt` (selected)         | `black`              | `red_alert`         | Bold        |
| `[!] Halt` (unselected)       | `red_alert`          | `red_alert` 20%     | Bold        |
| `Cancel` (selected)           | `space_white`        | `galaxy_gray`       | Bold        |
| `Cancel` (unselected)         | `light_gray`         | `galaxy_gray` 20%   | --          |
| `[~] Shelve` (selected)       | `black`              | `yellow_caution`    | Bold        |
| `[~] Shelve` (unselected)     | `yellow_caution`     | `yellow_caution` 20%| Bold        |
| Hint text                     | `galaxy_gray`        | --                  | Faint       |
| Hint key names                | `moonlit_violet`     | --                  | Bold        |
| Background (dimmed)           | `space_white`        | --                  | Faint (15%) |
