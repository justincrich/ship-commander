package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ship-commander/sc3/internal/tui/theme"
)

const (
	// MaxNavigationDepth bounds the stack depth for Enter-driven navigation.
	MaxNavigationDepth = 3
	// StandardLayoutMinWidth is the terminal width threshold for side-by-side panels.
	StandardLayoutMinWidth = 120
)

// ViewID identifies a top-level or nested TUI view.
type ViewID string

const (
	// ViewFleetOverview is the default landing screen.
	ViewFleetOverview ViewID = "fleet_overview"
	// ViewShipBridge is the per-ship execution dashboard.
	ViewShipBridge ViewID = "ship_bridge"
	// ViewReadyRoom is the planning session view.
	ViewReadyRoom ViewID = "ready_room"
	// ViewPlanReview is the plan review drill-down view.
	ViewPlanReview ViewID = "plan_review"
	// ViewMissionDetail is the mission drill-down view.
	ViewMissionDetail ViewID = "mission_detail"
	// ViewAgentDetail is the agent drill-down view.
	ViewAgentDetail ViewID = "agent_detail"
)

// LayoutMode identifies responsive AppShell layout mode.
type LayoutMode string

const (
	// LayoutStandard renders side-by-side panel arrangements.
	LayoutStandard LayoutMode = "standard"
	// LayoutCompact renders vertically stacked panel arrangements.
	LayoutCompact LayoutMode = "compact"
)

// OverlayKind identifies modal overlays rendered above base views.
type OverlayKind string

const (
	// OverlayKindAdmiralQuestion presents a blocking Admiral prompt.
	OverlayKindAdmiralQuestion OverlayKind = "admiral_question"
	// OverlayKindHelp presents global keyboard and view help.
	OverlayKindHelp OverlayKind = "help"
	// OverlayKindConfirmQuit presents a quit confirmation modal.
	OverlayKindConfirmQuit OverlayKind = "confirm_quit"
)

// Overlay represents one modal layer in the AppShell overlay stack.
type Overlay struct {
	Kind    OverlayKind
	Payload string
}

// ViewRenderer renders one view from root model state.
type ViewRenderer func(model AppModel) string

// ViewDefinition configures AppShell behavior for one view.
type ViewDefinition struct {
	FocusOrder  []string
	EnterTarget ViewID
	Render      ViewRenderer
}

// NavigateMsg requests stack push navigation to a specific view.
type NavigateMsg struct {
	View ViewID
}

// OverlayPushMsg pushes an overlay onto the modal stack.
type OverlayPushMsg struct {
	Overlay Overlay
}

// OverlayPopMsg pops the top overlay if one exists.
type OverlayPopMsg struct{}

// SetViewFocusOrderMsg sets the panel focus cycle order for a view.
type SetViewFocusOrderMsg struct {
	View   ViewID
	Panels []string
}

// AppModel is the root Bubble Tea model that owns all TUI AppShell state.
type AppModel struct {
	viewDefs      map[ViewID]ViewDefinition
	navStack      []ViewID
	overlays      []Overlay
	focusByView   map[ViewID]int
	width         int
	height        int
	layoutMode    LayoutMode
	quitting      bool
	standardWidth int
}

// NewAppModel constructs a root AppShell model with an initial view.
func NewAppModel(initialView ViewID, defs map[ViewID]ViewDefinition) *AppModel {
	if initialView == "" {
		initialView = ViewFleetOverview
	}

	model := &AppModel{
		viewDefs:      make(map[ViewID]ViewDefinition, len(defs)),
		navStack:      []ViewID{initialView},
		overlays:      make([]Overlay, 0, 3),
		focusByView:   map[ViewID]int{initialView: 0},
		layoutMode:    LayoutStandard,
		standardWidth: StandardLayoutMinWidth,
	}

	for viewID, def := range defs {
		model.viewDefs[viewID] = def
	}
	if _, exists := model.focusByView[initialView]; !exists {
		model.focusByView[initialView] = 0
	}

	return model
}

// Init satisfies tea.Model.
func (m *AppModel) Init() tea.Cmd {
	return nil
}

// Update handles global AppShell events and keyboard shortcuts.
func (m *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch typed := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = typed.Width
		m.height = typed.Height
		m.layoutMode = resolveLayoutMode(typed.Width, m.standardWidth)
		return m, nil
	case NavigateMsg:
		m.PushView(typed.View)
		return m, nil
	case OverlayPushMsg:
		m.PushOverlay(typed.Overlay)
		return m, nil
	case OverlayPopMsg:
		m.PopOverlay()
		return m, nil
	case SetViewFocusOrderMsg:
		def := m.viewDefs[typed.View]
		def.FocusOrder = cloneStrings(typed.Panels)
		m.viewDefs[typed.View] = def
		if _, exists := m.focusByView[typed.View]; !exists {
			m.focusByView[typed.View] = 0
		}
		return m, nil
	case tea.KeyMsg:
		return m.handleGlobalKey(typed)
	default:
		return m, nil
	}
}

func (m *AppModel) handleGlobalKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "tab":
		m.CyclePanelFocus(true)
		return m, nil
	case "shift+tab":
		m.CyclePanelFocus(false)
		return m, nil
	case "?":
		m.PushOverlay(Overlay{Kind: OverlayKindHelp})
		return m, nil
	case "q", "ctrl+c":
		m.PushOverlay(Overlay{Kind: OverlayKindConfirmQuit})
		return m, nil
	case "esc":
		if m.PopOverlay() {
			return m, nil
		}
		m.PopView()
		return m, nil
	case "enter":
		if current, ok := m.CurrentOverlay(); ok {
			if current.Kind == OverlayKindConfirmQuit {
				m.quitting = true
				return m, tea.Quit
			}
			m.PopOverlay()
			return m, nil
		}

		def := m.viewDefs[m.CurrentView()]
		if def.EnterTarget != "" {
			m.PushView(def.EnterTarget)
		}
		return m, nil
	default:
		return m, nil
	}
}

// View satisfies tea.Model by dispatching to the current view renderer.
func (m *AppModel) View() string {
	base := m.renderCurrentView()
	if overlay, ok := m.CurrentOverlay(); ok {
		overlayBody := fmt.Sprintf("Overlay: %s", overlay.Kind)
		if overlay.Payload != "" {
			overlayBody = overlayBody + "\n" + overlay.Payload
		}
		return lipgloss.JoinVertical(lipgloss.Left, base, theme.OverlayBorder.Render(overlayBody))
	}
	return base
}

func (m *AppModel) renderCurrentView() string {
	currentView := m.CurrentView()
	def, ok := m.viewDefs[currentView]
	if ok && def.Render != nil {
		return def.Render(*m)
	}
	return fmt.Sprintf("view: %s", currentView)
}

// PushView appends a view onto the stack, replacing current when max depth is reached.
func (m *AppModel) PushView(view ViewID) {
	if view == "" {
		return
	}

	if len(m.navStack) == 0 {
		m.navStack = append(m.navStack, view)
		return
	}

	if len(m.navStack) >= MaxNavigationDepth {
		m.navStack[len(m.navStack)-1] = view
		return
	}

	m.navStack = append(m.navStack, view)
	if _, exists := m.focusByView[view]; !exists {
		m.focusByView[view] = 0
	}
}

// PopView pops one view from stack while retaining at least one root entry.
func (m *AppModel) PopView() bool {
	if len(m.navStack) <= 1 {
		return false
	}
	m.navStack = m.navStack[:len(m.navStack)-1]
	return true
}

// CurrentView returns the active view at the top of the navigation stack.
func (m AppModel) CurrentView() ViewID {
	if len(m.navStack) == 0 {
		return ViewFleetOverview
	}
	return m.navStack[len(m.navStack)-1]
}

// NavigationStack returns a copy of the full current navigation stack.
func (m AppModel) NavigationStack() []ViewID {
	out := make([]ViewID, len(m.navStack))
	copy(out, m.navStack)
	return out
}

// PushOverlay appends a modal overlay to the stack.
func (m *AppModel) PushOverlay(overlay Overlay) {
	if overlay.Kind == "" {
		return
	}
	m.overlays = append(m.overlays, overlay)
}

// PopOverlay removes and returns true when the top overlay exists.
func (m *AppModel) PopOverlay() bool {
	if len(m.overlays) == 0 {
		return false
	}
	m.overlays = m.overlays[:len(m.overlays)-1]
	return true
}

// OverlayDepth returns the number of stacked overlays.
func (m AppModel) OverlayDepth() int {
	return len(m.overlays)
}

// CurrentOverlay returns the top modal overlay, if present.
func (m AppModel) CurrentOverlay() (Overlay, bool) {
	if len(m.overlays) == 0 {
		return Overlay{}, false
	}
	return m.overlays[len(m.overlays)-1], true
}

// CyclePanelFocus advances or reverses panel focus for the active view.
func (m *AppModel) CyclePanelFocus(forward bool) {
	currentView := m.CurrentView()
	def, ok := m.viewDefs[currentView]
	if !ok || len(def.FocusOrder) == 0 {
		return
	}

	index := m.focusByView[currentView]
	if index < 0 || index >= len(def.FocusOrder) {
		index = 0
	}

	if forward {
		index = (index + 1) % len(def.FocusOrder)
	} else {
		index = (index - 1 + len(def.FocusOrder)) % len(def.FocusOrder)
	}
	m.focusByView[currentView] = index
}

// FocusedPanel returns the active panel ID for the current view.
func (m AppModel) FocusedPanel() string {
	currentView := m.CurrentView()
	def, ok := m.viewDefs[currentView]
	if !ok || len(def.FocusOrder) == 0 {
		return ""
	}

	index := m.focusByView[currentView]
	if index < 0 || index >= len(def.FocusOrder) {
		return def.FocusOrder[0]
	}

	return def.FocusOrder[index]
}

// PanelBorderStyle returns focused/unfocused border styles for view panel framing.
func (m AppModel) PanelBorderStyle(panelID string) lipgloss.Style {
	if panelID != "" && panelID == m.FocusedPanel() {
		return theme.PanelBorderFocused
	}
	return theme.PanelBorder
}

// PanelTitleStyle returns focused/unfocused title style for panel frames.
func (m AppModel) PanelTitleStyle(panelID string) lipgloss.Style {
	if panelID != "" && panelID == m.FocusedPanel() {
		return theme.PanelTitleFocusedStyle
	}
	return lipgloss.NewStyle().Foreground(theme.GalaxyGrayColor)
}

// LayoutMode reports the active responsive layout mode.
func (m AppModel) LayoutMode() LayoutMode {
	return m.layoutMode
}

// Dimensions reports current terminal width and height.
func (m AppModel) Dimensions() (int, int) {
	return m.width, m.height
}

// Quitting reports whether a quit confirmation was accepted.
func (m AppModel) Quitting() bool {
	return m.quitting
}

func resolveLayoutMode(width int, standardMinWidth int) LayoutMode {
	if width < standardMinWidth {
		return LayoutCompact
	}
	return LayoutStandard
}

func cloneStrings(in []string) []string {
	out := make([]string, len(in))
	copy(out, in)
	return out
}
