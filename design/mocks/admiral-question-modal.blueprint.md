# Admiral Question Modal -- ASCII Blueprint

> Source: `admiral-question-modal.mock.html` | Spec: `prompts/admiral-question-modal.spec.json`
> Terminal: 120 cols x 30 rows (standard breakpoint)

## Layout Structure

The admiral question modal is a non-dismissable overlay that appears over any active view when a crew member surfaces a question. It steals all keyboard input until the admiral submits a response.

```
DimmedBackground (100%, full viewport)
══════════════════════════════════════
  ModalOverlay (60% width, centered)
    ModalHeader (100%, 2 rows)
    QuestionContent (100%, auto)
    OptionSelect (100%, auto)
    FreeTextInput (100%, auto)
    BroadcastToggle (100%, 1 row)
    SubmitHint (100%, 1 row)
══════════════════════════════════════
```

- **Dimmed Background**: Underlying view rendered at ~15% opacity. Represents whatever screen the admiral was viewing (Ship Bridge, Fleet Overview, etc.).
- **Modal Header**: 2 rows. Title "ADMIRAL -- QUESTION FROM \<Agent Name\>" in bold butterscotch. Ship name in galaxy_gray faint. Role badge (Commander=blue, Captain=gold, Ensign=butterscotch). Domain badge (technical=butterscotch, functional=blue, design=purple).
- **Question Content**: Auto-height. Question text rendered via Glamour markdown in space_white.
- **Option Select**: Auto-height. huh.Select with predefined answer options. Selected option highlighted in moonlit_violet. Up/Down to navigate.
- **Free Text Input**: Auto-height. huh.Input for custom response. Placeholder text in galaxy_gray.
- **Broadcast Toggle**: 1 row. huh.Confirm checkbox: "Broadcast answer to all crew?"
- **Submit Hint**: 1 row. "Press Enter to submit" centered in galaxy_gray.

## Component Mapping (from components.yaml)

| Visual Area | Component | Charm Base | Key Props |
|---|---|---|---|
| Dimmed background | `lipgloss.NewStyle()` | `lipgloss.NewStyle()` | underlying view at ~15% opacity |
| Modal container | **ModalOverlay** (#5) | `lipgloss.Place` + `lipgloss.DoubleBorder()` | 60% width centered, butterscotch border, `modal_open`/`modal_close` spring |
| Modal header | Custom `lipgloss.NewStyle()` | `lipgloss.NewStyle()` | title, ship, role badge, domain badge |
| Question content | Glamour markdown | `bubbles.viewport` | Glamour-rendered question text in space_white |
| Option selector | `huh.Select` | `huh.Select` | predefined answer options, `>` cursor in moonlit_violet |
| Free text input | `huh.Input` | `huh.Input` | custom response, placeholder in galaxy_gray |
| Broadcast toggle | `huh.Confirm` | `huh.Confirm` | "Broadcast answer to all crew?" checkbox |
| Full form wrapper | **AdmiralQuestionForm** (#21) | `huh.Form` wrapping `huh.Note` + `huh.Select` + `huh.Input` + `huh.Confirm` | non-dismissable, Enter to submit |

## Token Reference

| Element                | Token(s)                                          | Hex         |
|------------------------|---------------------------------------------------|-------------|
| Modal border           | `lipgloss.DoubleBorder()`, `butterscotch`         | `#FF9966`   |
| Modal title            | `butterscotch` (bold)                             | `#FF9966`   |
| Ship name              | `galaxy_gray` (faint)                             | `#52526A`   |
| Role badge COMMANDER   | `blue` (bg), `black` (fg)                         | `#9999CC`   |
| Role badge CAPTAIN     | `gold` (bg), `black` (fg)                         | `#FFAA00`   |
| Role badge ENSIGN      | `butterscotch` (bg), `black` (fg)                 | `#FF9966`   |
| Domain badge TECHNICAL | `butterscotch` (bg), `black` (fg)                 | `#FF9966`   |
| Domain badge FUNCTIONAL| `blue` (bg), `black` (fg)                         | `#9999CC`   |
| Domain badge DESIGN    | `purple` (bg), `black` (fg)                       | `#CC99CC`   |
| Question label         | `space_white` (bold)                              | `#F5F6FA`   |
| Question text          | `space_white`                                     | `#F5F6FA`   |
| Option selected        | `moonlit_violet`                                  | `#9966FF`   |
| Option unselected      | `light_gray`                                      | `#CCCCCC`   |
| Option cursor `>`      | `moonlit_violet`                                  | `#9966FF`   |
| Input label            | `galaxy_gray`                                     | `#52526A`   |
| Input placeholder      | `galaxy_gray`                                     | `#52526A`   |
| Input border           | `galaxy_gray`, focused: `blue`                    | `#52526A`   |
| Input text             | `space_white`                                     | `#F5F6FA`   |
| Broadcast checkbox     | `light_gray`                                      | `#CCCCCC`   |
| Submit hint            | `galaxy_gray`                                     | `#52526A`   |
| Dimmed background      | `space_white` at ~15% opacity                     | `#F5F6FA`   |
| Question icon          | `gold`                                            | `#FFAA00`   |

## ASCII Blueprint (120x30 standard)

```
░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░
░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░
░░░ SHIP BRIDGE - USS ENTERPRISE ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░
░░░ CREW ROSTER ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░
░░░ Commander Data  ACTIVE  Technical ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░
░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░
░░░░░░░░░░░░░░░░░░░╔══════════════════════════════════════════════════════════════════════════════╗░░░░░░░░░░░░░░░░░░░░░░░░░
░░░░░░░░░░░░░░░░░░░║  ADMIRAL -- QUESTION FROM Cmdr. Data                                       ║░░░░░░░░░░░░░░░░░░░░░░░░░
░░░░░░░░░░░░░░░░░░░║  USS Enterprise   COMMANDER   TECHNICAL                                    ║░░░░░░░░░░░░░░░░░░░░░░░░░
░░░░░░░░░░░░░░░░░░░╠══════════════════════════════════════════════════════════════════════════════╣░░░░░░░░░░░░░░░░░░░░░░░░░
░░░░░░░░░░░░░░░░░░░║                                                                            ║░░░░░░░░░░░░░░░░░░░░░░░░░
░░░░░░░░░░░░░░░░░░░║  Question:                                                                 ║░░░░░░░░░░░░░░░░░░░░░░░░░
░░░░░░░░░░░░░░░░░░░║  The authentication module requires a session store. Which approach         ║░░░░░░░░░░░░░░░░░░░░░░░░░
░░░░░░░░░░░░░░░░░░░║  should I use for the JWT token storage?                                   ║░░░░░░░░░░░░░░░░░░░░░░░░░
░░░░░░░░░░░░░░░░░░░║                                                                            ║░░░░░░░░░░░░░░░░░░░░░░░░░
░░░░░░░░░░░░░░░░░░░║  > Redis-backed session store (recommended for production)                  ║░░░░░░░░░░░░░░░░░░░░░░░░░
░░░░░░░░░░░░░░░░░░░║    Browser localStorage with httpOnly cookies                              ║░░░░░░░░░░░░░░░░░░░░░░░░░
░░░░░░░░░░░░░░░░░░░║    In-memory store (simplest, no persistence)                              ║░░░░░░░░░░░░░░░░░░░░░░░░░
░░░░░░░░░░░░░░░░░░░║    Database-backed sessions                                                ║░░░░░░░░░░░░░░░░░░░░░░░░░
░░░░░░░░░░░░░░░░░░░║                                                                            ║░░░░░░░░░░░░░░░░░░░░░░░░░
░░░░░░░░░░░░░░░░░░░║  Or provide a custom response:                                             ║░░░░░░░░░░░░░░░░░░░░░░░░░
░░░░░░░░░░░░░░░░░░░║  ┌──────────────────────────────────────────────────────────────────────┐   ║░░░░░░░░░░░░░░░░░░░░░░░░░
░░░░░░░░░░░░░░░░░░░║  │ Type a custom response...                                          │   ║░░░░░░░░░░░░░░░░░░░░░░░░░
░░░░░░░░░░░░░░░░░░░║  └──────────────────────────────────────────────────────────────────────┘   ║░░░░░░░░░░░░░░░░░░░░░░░░░
░░░░░░░░░░░░░░░░░░░║                                                                            ║░░░░░░░░░░░░░░░░░░░░░░░░░
░░░░░░░░░░░░░░░░░░░║  [ ] Broadcast answer to all crew?                                         ║░░░░░░░░░░░░░░░░░░░░░░░░░
░░░░░░░░░░░░░░░░░░░║                                                                            ║░░░░░░░░░░░░░░░░░░░░░░░░░
░░░░░░░░░░░░░░░░░░░║  ─────────────────────────────────────────────────────────────────────────  ║░░░░░░░░░░░░░░░░░░░░░░░░░
░░░░░░░░░░░░░░░░░░░║                          Press Enter to submit                             ║░░░░░░░░░░░░░░░░░░░░░░░░░
░░░░░░░░░░░░░░░░░░░╚══════════════════════════════════════════════════════════════════════════════╝░░░░░░░░░░░░░░░░░░░░░░░░░
░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░
░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░
```

### Annotation Notes

- The **modal border** uses `lipgloss.DoubleBorder()` characters (`╔ ╗ ╚ ╝ ║ ═`) in `butterscotch` (`#FF9966`). This is the standard modal border defined in `tokens.yaml`.
- The **dimmed background** is represented by `░` (light shade) characters. In implementation, the underlying view is rendered at ~15% opacity using lipgloss styling.
- The **modal title** "ADMIRAL -- QUESTION FROM Cmdr. Data" renders in bold `butterscotch`.
- The **ship name** "USS Enterprise" renders in faint `galaxy_gray` (`#52526A`).
- **Role badge** `COMMANDER` renders with `blue` (`#9999CC`) background and `black` foreground.
- **Domain badge** `TECHNICAL` renders with `butterscotch` (`#FF9966`) background and `black` foreground.
- The **selected option** (first item with `>` cursor) renders in `moonlit_violet` (`#9966FF`). Unselected options render in `light_gray` (`#CCCCCC`).
- The **text input** border uses `galaxy_gray` (`#52526A`), switching to `blue` (`#9999CC`) when focused. Placeholder text is `galaxy_gray`.
- The **broadcast checkbox** `[ ]` and label render in `light_gray`.
- The **submit hint** separator and text render in `galaxy_gray`.
- The **horizontal divider** (`╠═══╣`) separates header from content using the same `butterscotch` double-border style.
- The modal is **non-dismissable**: Escape does not close it. The admiral must submit a response via Enter.
- The modal uses `modal_open` spring animation (freq=6.0, ratio=0.8) on entrance for a gentle bounce effect, and `modal_close` (freq=10.0, ratio=1.0) for crisp dismissal.

## Color Annotations

| Region                         | Foreground Token    | Background Token   | Style       |
|--------------------------------|---------------------|--------------------|-------------|
| Modal border `╔═║╗╚╝`         | `butterscotch`      | --                 | Double      |
| "ADMIRAL -- QUESTION FROM..."  | `butterscotch`      | --                 | Bold        |
| "USS Enterprise"               | `galaxy_gray`       | --                 | Faint       |
| Badge `COMMANDER`              | `black`             | `blue`             | Bold        |
| Badge `TECHNICAL`              | `black`             | `butterscotch`     | Bold        |
| "Question:" label              | `space_white`       | --                 | Bold        |
| Question body text             | `space_white`       | --                 | --          |
| Selected option `>`            | `moonlit_violet`    | --                 | --          |
| Selected option text           | `moonlit_violet`    | --                 | --          |
| Unselected option text         | `light_gray`        | --                 | --          |
| "Or provide a custom response" | `galaxy_gray`       | --                 | --          |
| Input placeholder text         | `galaxy_gray`       | --                 | Faint       |
| Input border                   | `galaxy_gray`       | --                 | Normal      |
| Input border (focused)         | `blue`              | --                 | Normal      |
| Checkbox `[ ]`                 | `light_gray`        | --                 | --          |
| Broadcast label                | `light_gray`        | --                 | --          |
| Separator line                 | `dark_blue`         | --                 | --          |
| "Press Enter to submit"        | `galaxy_gray`       | --                 | Faint       |
| Dimmed background `░`          | `galaxy_gray`       | --                 | Faint       |

## Compact Variant (80x24 minimum)

In compact mode (`<120 cols`), the modal expands to 80% terminal width. Options stack tighter with no blank line separation. Free-text input height is reduced. The dimmed background remains.

```
░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░
░░░░ SHIP BRIDGE - USS ENTERPRISE ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░
░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░
░░░░░░░╔════════════════════════════════════════════════════════════════╗░░░░░░░░░░░
░░░░░░░║  ADMIRAL -- QUESTION FROM Cmdr. Data                         ║░░░░░░░░░░░
░░░░░░░║  USS Enterprise   COMMANDER   TECHNICAL                      ║░░░░░░░░░░░
░░░░░░░╠════════════════════════════════════════════════════════════════╣░░░░░░░░░░░
░░░░░░░║  Question:                                                   ║░░░░░░░░░░░
░░░░░░░║  The authentication module requires a session store.         ║░░░░░░░░░░░
░░░░░░░║  Which approach should I use for the JWT token storage?      ║░░░░░░░░░░░
░░░░░░░║                                                              ║░░░░░░░░░░░
░░░░░░░║  > Redis-backed session store (recommended for production)   ║░░░░░░░░░░░
░░░░░░░║    Browser localStorage with httpOnly cookies                ║░░░░░░░░░░░
░░░░░░░║    In-memory store (simplest, no persistence)                ║░░░░░░░░░░░
░░░░░░░║    Database-backed sessions                                  ║░░░░░░░░░░░
░░░░░░░║  Or provide a custom response:                               ║░░░░░░░░░░░
░░░░░░░║  ┌──────────────────────────────────────────────────────┐    ║░░░░░░░░░░░
░░░░░░░║  │ Type a custom response...                            │    ║░░░░░░░░░░░
░░░░░░░║  └──────────────────────────────────────────────────────┘    ║░░░░░░░░░░░
░░░░░░░║  [ ] Broadcast answer to all crew?                           ║░░░░░░░░░░░
░░░░░░░║  ────────────────────────────────────────────────────────    ║░░░░░░░░░░░
░░░░░░░║                    Press Enter to submit                     ║░░░░░░░░░░░
░░░░░░░╚════════════════════════════════════════════════════════════════╝░░░░░░░░░░░
░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░
░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░
```

### Compact Variant Notes

- Modal takes 80% of terminal width instead of 60% to maximize usable space.
- Blank lines between sections are removed; options stack directly after the question.
- Free-text input maintains full width within the modal interior.
- Broadcast toggle and submit hint are tightened with no extra vertical spacing.
- The underlying dimmed background still renders to maintain modal context.
- Keyboard navigation remains identical: Up/Down for options, Tab to cycle fields, Enter to submit.
