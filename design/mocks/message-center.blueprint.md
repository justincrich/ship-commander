# Message Center -- ASCII Blueprint

> Source: `message-center.mock.html` | Spec: `prompts/message-center.spec.json`
> Terminal: 120 cols x 30 rows (standard breakpoint)

## Layout Structure

InboxHeader (100%, 2 rows) | MessageList (40%) + MessageDetail (60%) | NavigableToolbar (100%, 1 row)

- **InboxHeader**: Row 1 = "MESSAGE CENTER" in bold butterscotch. Row 2 = inline metrics: Total count, Pending badge (pink bg), Blocking badge (red_alert bg), Answered count. Filter context indicator.
- **MessageList**: Scrollable inbox list sorted by priority then timestamp. Each message row shows sender name (role-colored), ship + mission reference, priority badge (BLOCKING/QUESTION/ANSWERED), preview text, relative timestamp. Selected row uses moonlit_violet left border highlight. Unread rows use pink left border.
- **MessageDetail**: Full message content for selected message. Header with sender, role badge, ship, mission, priority badge. Body with question text. For questions: inline option selector with cursor indicator. Footer with action hints.
- **NavigableToolbar**: Message action shortcuts with moonlit_violet keys and butterscotch labels.

## Component Mapping (from components.yaml)

| Visual Area | Component | Charm Base | Key Props |
|---|---|---|---|
| Full screen container | **AppShell** (#1) | `lipgloss.JoinVertical` | header, body, footer |
| Inbox + Detail layout | **PanelGrid** (#2) | `lipgloss.JoinHorizontal` | ratios: `[0.4, 0.6]`, compact_threshold: 120 |
| Inbox header | Custom `lipgloss.NewStyle()` | `lipgloss.NewStyle()` | "MESSAGE CENTER" bold butterscotch, inline metrics |
| Message list panel | **FocusablePanel** (#4) wrapping `bubbles.list` | `bubbles.list` | filterable, keyboard nav, scrollable |
| Message rows | **AgentCard** (#11) variant | `lipgloss.NewStyle()` | sender (role-colored), ship/mission ref, priority badge, preview |
| Priority badges | **StatusBadge** (#6) variant | `lipgloss.NewStyle()` | BLOCKING (red_alert bg), QUESTION (purple bg), ANSWERED (green_ok bg) |
| Message detail panel | **FocusablePanel** (#4) | `lipgloss.NewStyle().Border(lipgloss.RoundedBorder())` | full message content, question body |
| Question body | Glamour markdown | `bubbles.viewport` | Glamour-rendered question text |
| Option selector | `huh.Select` | `huh.Select` | predefined answer options, cursor in moonlit_violet |
| Broadcast toggle | `huh.Confirm` | `huh.Confirm` | "Broadcast answer to all crew?" checkbox |
| Bottom toolbar | **NavigableToolbar** | `lipgloss.JoinHorizontal` | [Enter] Answer [j/k] Navigate [f] Filter [?] Help [Esc] Fleet |

## Token Reference

| Token             | Hex       | Usage in this view                                      |
|-------------------|-----------|---------------------------------------------------------|
| butterscotch      | `#FF9966` | Header title, detail title, sender (implementer), labels|
| blue              | `#9999CC` | Sender (commander), detail meta values, stat labels     |
| gold              | `#FFAA00` | Sender (captain), captain role badge                    |
| green_ok          | `#33FF33` | ANSWERED badge bg                                       |
| red_alert         | `#FF3333` | BLOCKING badge bg, blocking indicator                   |
| purple            | `#CC99CC` | QUESTION badge bg                                       |
| pink              | `#FF99CC` | Pending badge bg, unread left border                    |
| moonlit_violet    | `#9966FF` | Selected message border, focused panel border, keys     |
| galaxy_gray       | `#52526A` | Timestamps, ship/mission refs, separators, inactive     |
| space_white       | `#F5F6FA` | Primary text, message body, stat values                 |
| light_gray        | `#CCCCCC` | Message preview text, secondary labels                  |
| black             | `#000000` | Terminal background, badge foregrounds                  |

| Icon              | Char | Usage                                                    |
|-------------------|------|----------------------------------------------------------|
| bullet            | `*`  | Toolbar separators                                       |
| separator         | `│`  | Inline dividers, panel section dividers                  |
| expand            | `▸`  | Cursor indicator for option selection                    |

| Border            | Chars          | Usage                                                |
|-------------------|----------------|------------------------------------------------------|
| RoundedBorder     | `╭ ╮ ╰ ╯ │ ─` | All panels (lipgloss.RoundedBorder)                  |
| Focused border    | moonlit_violet | Active/selected panel (message detail in this mock)  |
| Default border    | galaxy_gray    | Unfocused panels (message list)                      |

## ASCII Blueprint (120x30 standard)

```
╭──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╮
│  MESSAGE CENTER                                                                                                    │
│  Total: 8 messages   Pending: [3]   Blocking: [1]   Answered: 4                                                   │
╰──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╯
╭─ Inbox │ 3 pending ────────────────────────────────────────────╮╭─ Message Detail ─────────────────────────────────────────────╮
│ ┃ capt-alpha                                       2 min ago  ││ ADMIRAL -- QUESTION FROM capt-alpha                          │
│ ┃ SS Enterprise · MISSION-03              BLOCKING             ││ Ship: SS Enterprise   Mission: MISSION-03   CAPTAIN  BLOCKING│
│ ┃ Which session store approach for JWT...                      ││─────────────────────────────────────────────────────────────  │
│─────────────────────────────────────────────────────────────── ││                                                              │
│ ┃ cmdr-beta                                       15 min ago  ││ The authentication module requires a session store.           │
│ ┃ Nautilus · MISSION-11                   QUESTION             ││ Which approach should I use for the JWT token storage?        │
│ ┃ Should API versioning use path or header...                  ││                                                              │
│─────────────────────────────────────────────────────────────── ││ Context: The PRD specifies "secure token management" but     │
│ ┃ impl-charlie                                    32 min ago  ││ does not prescribe a specific implementation. Current         │
│ ┃ SS Enterprise · MISSION-04              QUESTION             ││ infrastructure includes Redis and PostgreSQL.                 │
│ ┃ The existing test helper uses a deprecated...                ││                                                              │
│─────────────────────────────────────────────────────────────── ││ Select response:                                             │
│   capt-echo                                        1h ago     ││ ▸ Redis-backed session store (recommended for production)     │
│   Nautilus · MISSION-10                   ANSWERED             ││   Browser localStorage with httpOnly cookies                  │
│   Confirmed: use PostgreSQL for persistence                    ││   In-memory store (simplest, no persistence)                  │
│─────────────────────────────────────────────────────────────── ││   Database-backed sessions                                   │
│   cmdr-lima                                        2h ago     ││                                                              │
│   Discovery · MISSION-20                  ANSWERED             ││                                                              │
│   Design approved with minor revisions...                      ││                                                              │
│                                                                ││─────────────────────────────────────────────────────────────  │
│                                                                ││    Press Enter to answer · b to broadcast · Esc to close     │
╰────────────────────────────────────────────────────────────────╯╰──────────────────────────────────────────────────────────────╯
╭──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╮
│ [Enter] Answer  ●  [j/k] Navigate  ●  [f] Filter  ●  [?] Help  ●  [Esc] Fleet                                    │
╰──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╯
```

### Annotation Notes

- The **selected message** (capt-alpha, first row) uses `moonlit_violet` (`#9966FF`) for its left border indicator `┃`. All unread messages also show `pink` (`#FF99CC`) left border. Read messages have no left border accent.
- The **Message List** panel label (`Inbox | 3 pending`) renders inline on the top border in `galaxy_gray`. The count "3 pending" uses `pink`.
- The **Message Detail** panel uses `moonlit_violet` border (focused state). The message list panel uses `galaxy_gray` border (unfocused).
- Priority badges are rendered with colored backgrounds: `BLOCKING` in red_alert bg, `QUESTION` in purple bg, `ANSWERED` in green_ok bg. All use `black` foreground.
- Sender names are color-coded by role: captain names use `gold`, commander names use `blue`, implementer names use `butterscotch`.
- The `[3]` pending badge renders with `pink` (`#FF99CC`) background and `black` foreground. The `[1]` blocking badge renders with `red_alert` (`#FF3333`) background and `black` foreground.
- The option selector `▸` cursor indicator uses `moonlit_violet` for the selected option. Unselected options use `light_gray`.
- Toolbar `●` separators use `galaxy_gray`. Shortcut keys like `[Enter]` use `moonlit_violet`. Labels use `butterscotch`.

## Color Annotations

```
INBOX HEADER
  "MESSAGE CENTER"            -> butterscotch (bold)
  "Total:" label              -> blue
  "8 messages" value          -> space_white (bold)
  "Pending:" label            -> blue
  "[3]" pending badge         -> bg:pink, fg:black (bold)
  "Blocking:" label           -> blue
  "[1]" blocking badge        -> bg:red_alert, fg:black (bold)
  "Answered:" label           -> blue
  "4" value                   -> space_white (bold)

MESSAGE LIST (unfocused)
  Panel border                -> galaxy_gray (RoundedBorder, unfocused state)
  Panel title "Inbox"         -> butterscotch (bold)
  "3 pending" count           -> pink

  Message: capt-alpha (selected + unread)
    Left border "┃"           -> moonlit_violet (selected)
    "capt-alpha" sender       -> gold (captain role)
    "2 min ago" timestamp     -> galaxy_gray
    "SS Enterprise" ship      -> galaxy_gray
    "MISSION-03" ref          -> galaxy_gray
    "BLOCKING" badge          -> bg:red_alert, fg:black
    Preview text              -> light_gray

  Message: cmdr-beta (unread)
    Left border "┃"           -> pink (unread)
    "cmdr-beta" sender        -> blue (commander role)
    "15 min ago" timestamp    -> galaxy_gray
    "QUESTION" badge          -> bg:purple, fg:black

  Message: impl-charlie (unread)
    Left border "┃"           -> pink (unread)
    "impl-charlie" sender     -> butterscotch (implementer role)
    "32 min ago" timestamp    -> galaxy_gray
    "QUESTION" badge          -> bg:purple, fg:black

  Message: capt-echo (read)
    No left border accent
    "capt-echo" sender        -> gold (captain role)
    "1h ago" timestamp        -> galaxy_gray
    "ANSWERED" badge          -> bg:green_ok, fg:black

  Message: cmdr-lima (read)
    No left border accent
    "cmdr-lima" sender        -> blue (commander role)
    "2h ago" timestamp        -> galaxy_gray
    "ANSWERED" badge          -> bg:green_ok, fg:black

MESSAGE DETAIL (focused)
  Panel border                -> moonlit_violet (RoundedBorder, focused state)
  Panel title                 -> butterscotch (bold)
  "ADMIRAL -- QUESTION..."    -> butterscotch (bold)
  "Ship:" label               -> galaxy_gray
  "SS Enterprise" value       -> blue
  "Mission:" label            -> galaxy_gray
  "MISSION-03" value          -> blue
  "CAPTAIN" role badge        -> bg:gold, fg:black
  "BLOCKING" badge            -> bg:red_alert, fg:black
  Question body text          -> space_white
  "Context:" text             -> space_white
  "Select response:" label    -> blue
  "▸" cursor (selected)       -> moonlit_violet
  Selected option text        -> moonlit_violet
  Unselected option text      -> light_gray
  Footer separator            -> galaxy_gray
  "Enter" key hint            -> moonlit_violet (bold)
  "b" key hint                -> moonlit_violet (bold)
  "Esc" key hint              -> galaxy_gray
  Footer text                 -> galaxy_gray

TOOLBAR
  Key badges "[Enter]" etc.   -> moonlit_violet (bold)
  Labels "Answer" etc.        -> butterscotch
  Separators "●"              -> galaxy_gray
```

## Compact Variant (80x24 minimum)

In compact mode (`<120 cols`), the message detail panel is hidden. The message list takes full width. Enter drills into a full-screen message detail view. Header collapses to a single line with abbreviated labels.

```
╭──────────────────────────────────────────────────────────────────────────────────╮
│  MESSAGE CENTER   Total: 8  Pending: [3]  Blocking: [1]  Answered: 4           │
╰──────────────────────────────────────────────────────────────────────────────────╯
╭─ Inbox │ 3 pending ────────────────────────────────────────────────────────────╮
│ ┃ capt-alpha                                                      2 min ago   │
│ ┃ SS Enterprise · MISSION-03                             BLOCKING             │
│ ┃ Which session store approach for JWT...                                     │
│────────────────────────────────────────────────────────────────────────────── │
│ ┃ cmdr-beta                                                      15 min ago   │
│ ┃ Nautilus · MISSION-11                                  QUESTION             │
│ ┃ Should API versioning use path or header...                                 │
│────────────────────────────────────────────────────────────────────────────── │
│ ┃ impl-charlie                                                   32 min ago   │
│ ┃ SS Enterprise · MISSION-04                             QUESTION             │
│ ┃ The existing test helper uses a deprecated...                               │
│────────────────────────────────────────────────────────────────────────────── │
│   capt-echo                                                       1h ago      │
│   Nautilus · MISSION-10                                  ANSWERED             │
│   Confirmed: use PostgreSQL for persistence                                   │
│────────────────────────────────────────────────────────────────────────────── │
│   cmdr-lima                                                       2h ago      │
│   Discovery · MISSION-20                                 ANSWERED             │
│   Design approved with minor revisions...                                     │
╰──────────────────────────────────────────────────────────────────────────────────╯
╭──────────────────────────────────────────────────────────────────────────────────╮
│ [Enter] Answer  ●  [j/k] Nav  ●  [f] Filter  ●  [?] Help  ●  [Esc] Fleet      │
╰──────────────────────────────────────────────────────────────────────────────────╯
```

### Compact Variant Notes

- Message detail panel is completely hidden. Enter on a selected message opens a full-screen detail view.
- Header collapses from 2 content rows to 1. Labels remain but values are tightened.
- Toolbar labels are shortened: "Nav" for Navigate.
- Message rows maintain the same structure (sender, ship/mission, badge, preview) but use full width.
- Scrolling is required to see all messages since only ~5 messages fit in the visible area at 24 rows.
- The selected message still uses `moonlit_violet` left border for focus indication. Unread messages use `pink` left border.
