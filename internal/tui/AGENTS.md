# Bubble Tea TUI - Agent Coordination

## Purpose

This package implements the Ship Commander 3 Terminal User Interface using **Bubble Tea** (Elm architecture for terminals), **Lipgloss** (styling), **Glamour** (markdown rendering), and **Huh** (terminal forms).

## Context in Ship Commander 3

The TUI provides real-time visibility into:
- Fleet Overview (landing screen with ship cards)
- Ship Bridge (per-ship execution view)
- Ready Room (planning orchestrator view)
- Plan Review (mission manifest approval)
- Mission Detail (per-mission TDD pipeline view)
- Agent Detail (real-time agent output)
- Message Center (Admiral question inbox)
- Project Settings (configuration UI)
- Help Overlay (keyboard shortcuts)

## Design Artifacts References

**CRITICAL**: The TUI implementation MUST conform to the design specifications in `design/`:

### Design Documents (READ THESE FIRST!)
1. **`design/UX-DESIGN-PLAN.md`** (if exists) - Full UX/UX specifications
2. **`design/views.yaml`** - All 16 screens defined with layouts, panels, data sources
3. **`design/components.yaml`** - 35 reusable components with Charm library mappings
4. **`design/flows.yaml`** - 16 micro-level interaction flows
5. **`design/workflows.yaml`** - 8 end-to-end user workflows
6. **`design/paradigm.yaml`** - Design patterns, principles, animation categories
7. **`design/config.yaml`** - Color palette, technology stack, keyboard plan

### Binding Design Principles (from UX Plan)
1. **P1: Flows Before Screens** - Every screen emerges from a user workflow
2. **P2: Deterministic Feedback, Always** - Every gate result from deterministic data
3. **P3: Symbols + Color, Never Color Alone** - Triple redundancy for status
4. **P4: Keyboard-First, Always** - Every interaction reachable via keyboard
5. **P5: Progressive Disclosure Over Information Overload** - Basic/Advanced/Executive modes
6. **P6: Smart Main, Dumb Components** - Single root `tea.Model` owns all state
7. **P7: Charm Aesthetic Meets LCARS Soul** - Fusion of Charm playfulness with LCARS theme

## Package Organization

```
internal/tui/
├── AGENTS.md              # This file
├── app.go                 # Root Bubble Tea model (owns all state)
├── styles.go              # LCARS theme definitions
├── navigation.go          # Navigation stack management
├── state.go               # TUI state persistence (~/.sc3/tui-state.json)
│
├── views/                 # TUI views (16 screens)
│   ├── AGENTS.md         # Views package guidance
│   ├── fleet_overview.go # Fleet Overview (landing screen)
│   ├── ship_bridge.go    # Ship Bridge (execution view)
│   ├── ready_room.go     # Ready Room (planning view)
│   ├── plan_review.go    # Plan Review (approval view)
│   ├── fleet_monitor.go  # Fleet Monitor (multi-ship status)
│   ├── mission_detail.go # Mission Detail (drill-down)
│   ├── agent_detail.go   # Agent Detail (drill-down)
│   ├── specialist_detail.go # Specialist Detail (Ready Room drill-down)
│   ├── agent_roster.go   # Agent Roster (top-level)
│   ├── directive_editor.go # Directive Editor (PRD input)
│   ├── message_center.go # Message Center (Admiral questions)
│   ├── wave_manager.go   # Wave Manager (overlay)
│   ├── project_settings.go # Project Settings
│   └── help_overlay.go   # Help Overlay
│
├── components/            # Reusable TUI components (35 defined)
│   ├── AGENTS.md         # Components package guidance
│   ├── status_badge.go   # Status indicator [symbol] LABEL
│   ├── panel_frame.go    # LCARS-styled panel with title/border
│   ├── navigable_toolbar.go # Keyboard shortcuts toolbar
│   ├── phase_indicator.go # 6-phase TDD pipeline (RED > V_RED > GRN > V_REF)
│   ├── wave_bar.go       # Wave progress bar
│   └── ship_status_row.go # Compact ship status (one line)
│
├── forms/                 # Huh form wrappers
│   ├── AGENTS.md         # Forms package guidance
│   ├── question_modal.go # Admiral question form (Select + Input + Confirm)
│   ├── approval_form.go  # Plan approval (Select + Text feedback)
│   └── confirm_dialog.go # Destructive action confirmation
│
└── theme/                 # LCARS theme system
    ├── AGENTS.md         # Theme package guidance
    ├── colors.go         # LCARS color palette (16 colors)
    ├── styles.go         # Semantic styles (Active, Success, Error, etc.)
    └── borders.go        # Border definitions
```

## Charmbracelet Dependencies

### Core Libraries (Required)
| Library | Min Version | Purpose | Documentation |
|---------|-------------|---------|---------------|
| **Bubble Tea** | >=1.2.0 | TUI framework (Elm architecture) | https://github.com/charmbracelet/bubbletea |
| **Lipgloss** | >=1.0.0 | Styling, borders, layout composition | https://github.com/charmbracelet/lipgloss |
| **Bubbles** | >=0.20.0 | Pre-built components (spinner, progress, table, list, viewport) | https://github.com/charmbracelet/bubbles |
| **Huh** | >=0.6.0 | Terminal forms for Admiral interactions | https://github.com/charmbracelet/huh |
| **Glamour** | >=0.8.0 | Markdown rendering (PRD, specs, demo tokens) | https://github.com/charmbracelet/glamour |
| **Harmonica** | >=0.2.0 | Spring-based physics animations | https://github.com/charmbracelet/harmonica |
| **Log** | >=0.4.0 | Structured logging with Lipgloss styling | https://github.com/charmbracelet/log |

### Import Paths
```go
import (
    tea "github.com/charmbracelet/bubbletea"
    lp "github.com/charmbracelet/lipgloss"
    "github.com/charmbracelet/bubbles/textinput"
    "github.com/charmbracelet/bubbles/viewport"
    "github.com/charmbracelet/huh"
    "github.com/charmbracelet/glamour"
    "github.com/charmbracelet/harmonica"
    "github.com/charmbracelet/log"
)
```

## Key Principles

### 1. Elm Architecture (Model-Update-View)
Bubble Tea uses the Elm architecture pattern:
- **Model**: Application state (struct)
- **Update**: Pure function: `(Model, Msg) → (Model, Cmd)`
- **View**: Pure function: `Model → string` (UI rendering)

```go
// ✅ GOOD: Elm architecture
type model struct {
    cursor   int
    selected map[string]bool
    quitting bool
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg {
    case tea.KeyMsg:
        return m, nil // Handle input
    }
    return m, nil
}

func (m model) View() string {
    return "Render UI here"
}
```

### 2. Single Root Model
- ONE root `tea.Model` owns all TUI state
- Child views are pure render functions (no state)
- State flows down, events bubble up

```go
// ✅ GOOD: Single root model
type rootModel struct {
    currentView    ViewType
    navStack       []ViewType
    scrollPositions map[string]int
    commissions    []Commission
    missions       []Mission
    agents         []Agent
    // ALL STATE HERE
}

// ❌ BAD: Multiple independent models
type fleetModel struct { ... }  // Separate state
type bridgeModel struct { ... } // Separate state
```

### 3. Dumb Components
- Components receive data as arguments
- Components hold NO state
- Components emit NO commands
- Components are pure functions: `data → styled string`

```go
// ✅ GOOD: Dumb component
func RenderStatusBadge(status string) string {
    symbol, color := getStatusSymbol(status)
    return lp.NewStyle().Foreground(color).Render(fmt.Sprintf("[%s] %s", symbol, status))
}

// ❌ BAD: Smart component (stateful)
type StatusBadge struct {
    status string
    func()  // Can emit commands!
}
```

### 4. Event Bus Integration
- Subscribe to protocol events in `Init()`
- Update model when events received
- Re-render view on state changes

```go
func (m *rootModel) Init() tea.Cmd {
    // Subscribe to events
    return tea.Batch(
        m.tickCmd(),
        m.subscribeEvents(),
    )
}

func (m *rootModel) subscribeEvents() tea.Cmd {
    return func() tea.Msg {
        // Wait for event from protocol bus
        event := <-m.eventCh
        return EventMsg{Event: event}
    }
}
```

## Dependencies

### Internal Dependencies
- `internal/beads` - Query commissions, missions, agents
- `internal/protocol` - Subscribe to event bus
- `internal/config` - TUI preferences (mode, animations)
- `internal/telemetry` - Trace user interactions

### External Dependencies
- All Charmbracelet libraries (see table above)
- No other UI framework dependencies

## Coding Standards

Follow Go standards from `.spec/research-findings/GO_CODING_STANDARDS.md`:

### Lipgloss Styling
```go
// ✅ GOOD: Reusable styles from theme package
import "github.com/ship-commander-3/internal/tui/theme"

var HeaderStyle = theme.HeaderStyle // Defined in theme/
var ActiveStyle = theme.ActiveStyle
var SuccessStyle = theme.SuccessStyle

// ❌ BAD: Hardcoded colors everywhere
lp.NewStyle().Foreground(lp.Color("#FF9966")) // Not reusable
```

### View Rendering
```go
// ✅ GOOD: Compose views with Lipgloss
func (m model) View() string {
    leftPanel := m.renderFleetList()
    rightPanel := m.renderShipPreview()
    return lp.JoinHorizontal(lp.Top, leftPanel, rightPanel)
}

// ❌ BAD: String concatenation
func (m model) View() string {
    return leftPanel + rightPanel // No layout control
}
```

### Navigation
```go
// ✅ GOOD: Stack-based navigation
func (m rootModel) navigateTo(view ViewType) (tea.Model, tea.Cmd) {
    m.navStack = append(m.navStack, m.currentView)
    m.currentView = view
    return m, nil
}

func (m rootModel) navigateBack() (tea.Model, tea.Cmd) {
    if len(m.navStack) == 0 {
        return m, tea.Quit
    }
    m.currentView = m.navStack[len(m.navStack)-1]
    m.navStack = m.navStack[:len(m.navStack)-1]
    return m, nil
}
```

## Common Patterns

### Root Model Initialization
```go
type Model struct {
    ctx         context.Context
    config      *config.Config
    beads       *beads.Client
    eventBus    *protocol.Bus

    // Navigation state
    currentView ViewType
    navStack    []ViewType

    // Data
    commissions []Commission
    missions    []Mission
    agents      []Agent

    // UI state
    ready       bool
    quitting     bool
}

func NewModel(ctx context.Context, cfg *config.Config) (*Model, error) {
    beadsClient, err := beads.NewClient(cfg.BeadsPath)
    if err != nil {
        return nil, err
    }

    return &Model{
        ctx:       ctx,
        config:    cfg,
        beads:     beadsClient,
        eventBus:  protocol.NewBus(),
        ready:     false,
    }, nil
}
```

### View Switching
```go
type ViewType string

const (
    ViewFleetOverview  ViewType = "fleet_overview"
    ViewShipBridge    ViewType = "ship_bridge"
    ViewReadyRoom     ViewType = "ready_room"
    // ... 13 more views
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg {
    case NavigateMsg:
        return m.navigateTo(msg.View)
    case BackMsg:
        return m.navigateBack()
    case QuitMsg:
        m.quitting = true
        return m, tea.Quit
    }
    return m, nil
}

func (m Model) View() string {
    if !m.ready {
        return "Loading...\n"
    }

    switch m.currentView {
    case ViewFleetOverview:
        return m.renderFleetOverview()
    case ViewShipBridge:
        return m.renderShipBridge()
    // ... all 16 views
    default:
        return "Unknown view"
    }
}
```

### Glamour Markdown Rendering
```go
import "github.com/charmbracelet/glamour"

func (m Model) renderPRD(prdContent string) string {
    // Create Glamour renderer with LCARS theme
    renderer, err := glamour.NewTermRenderer(
        glamour.WithAutoStyle(),
        glamour.WithWordWrap(m.width),
    )
    if err != nil {
        return fmt.Sprintf("Error creating renderer: %v", err)
    }

    // Render markdown
    out, err := renderer.Render(prdContent)
    if err != nil {
        return fmt.Sprintf("Error rendering markdown: %v", err)
    }

    return out
}
```

### Huh Form Integration
```go
import "github.com/charmbracelet/huh"

func (m Model) ShowQuestionModal(question Question) (tea.Model, tea.Cmd) {
    form := huh.NewForm(
        huh.NewGroup(
            huh.NewSelect[string]().
                Title("Question from " + question.AgentRole).
                Options(huh.NewOptions(question.Options...)...).
                Value(&m.answer),

            huh.NewConfirm().
                Title("Broadcast answer to all agents?").
                Value(&m.broadcast),
        ),
    )

    // Return form as model
    return form, tea.Sequential(form)
}
```

### Panel Frame Component
```go
// internal/tui/components/panel_frame.go

type PanelFrameConfig struct {
    Title    string
    Subtitle string
    Focused  bool
    Active   bool  // Overlay state
}

func RenderPanelFrame(config PanelFrameConfig, content string) string {
    var style lp.Style

    switch {
    case config.Active:
        style = theme.OverlayBorder // Butterscotch, rounded
    case config.Focused:
        style = theme.PanelBorderFocused // Double, violet
    default:
        style = theme.PanelBorder // Normal, gray
    }

    title := config.Title
    if config.Subtitle != "" {
        title = fmt.Sprintf("%s — %s", config.Title, config.Subtitle)
    }

    frame := lp.NewStyle().
        Border(style).
        BorderForeground(theme.ColorsButterscotch).
        BorderTitle(title, true)

    return frame.Render(content)
}
```

## Anti-Patterns to Avoid

### ❌ DON'T: Store state in components
```go
type StatusBadge struct {
    status string // BAD! Component should be stateless
}
```

### ✅ DO: Pass data to render functions
```go
func RenderStatusBadge(status string) string {
    // GOOD! Pure function
}
```

### ❌ DON'T: Use goroutines in Bubble Tea
```go
func (m model) Init() tea.Cmd {
    go func() {
        // BAD! Uncontrolled goroutine
    }()
}
```

### ✅ DO: Use tea.Cmd for async operations
```go
func (m model) Init() tea.Cmd {
    return tea.Tick(time.Second, func(t time.Time) tea.Msg {
        return TickMsg(t)
    })
}
```

### ❌ DON'T: Mutate model in View()
```go
func (m model) View() string {
    m.counter++ // BAD! View should be pure
}
```

### ✅ DO: Only read model in View()
```go
func (m model) View() string {
    return fmt.Sprintf("Count: %d", m.counter) // GOOD!
}
```

## Testing Requirements

### Unit Tests
- Test Update() logic for all messages
- Test navigation stack operations
- Test component rendering with different inputs

### Integration Tests
- Test view switching
- Test keyboard input handling
- Test form submission

### TUI Testing
- Use `github.com/charmbracelet/bubbles/teatest` for end-to-end tests
- Simulate keyboard input
- Capture and validate output

## LCARS Theme System

See `internal/tui/theme/AGENTS.md` for complete theme specification:
- **Reference Colors**: 16 LCARS color constants
- **Semantic Styles**: Active, Success, Error, Warning, Info, Planning
- **Border Styles**: Normal, Focused, Active Overlay
- **Accessibility**: Symbol + color, never color alone

## Animation System (Harmonica)

See `design/paradigm.yaml` for animation categories:
- **Modal Open/Close**: freq 6.0, damp 1.0
- **View Transition**: freq 5.0, damp 1.0
- **Panel Resize**: freq 4.0, damp 1.2
- **Progress Bar Update**: freq 3.0, damp 1.5
- **Phase Advance**: freq 5.0, damp 0.8

Max 3 concurrent animations enforced.

## References

### Design Artifacts (READ THESE!)
- `design/views.yaml` - All 16 screens defined
- `design/components.yaml` - 35 reusable components
- `design/flows.yaml` - 16 interaction flows
- `design/workflows.yaml` - 8 end-to-end workflows
- `design/paradigm.yaml` - Design patterns and principles
- `design/config.yaml` - Color palette and technology stack

### External Documentation
- [Bubble Tea Documentation](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lipgloss Documentation](https://github.com/charmbracelet/lipgloss) - Styling
- [Bubbles Documentation](https://github.com/charmbracelet/bubbles) - Components
- [Huh Documentation](https://github.com/charmbracelet/huh) - Forms
- [Glamour Documentation](https://github.com/charmbracelet/glamour) - Markdown
- [Harmonica Documentation](https://github.com/charmbracelet/harmonica) - Animations

### Internal Documentation
- `.spec/prd.md` - UC-TUI-01 through UC-TUI-28 (28 TUI use cases)
- `.spec/technical-requirements.md` - TUI Architecture (lines 571-843)
- `.spec/research-findings/GO_CODING_STANDARDS.md` - Go coding standards

---

**Version**: 1.0
**Last Updated**: 2025-02-10
