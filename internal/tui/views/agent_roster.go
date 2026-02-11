package views

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ship-commander/sc3/internal/tui/components"
	"github.com/ship-commander/sc3/internal/tui/theme"
)

const (
	agentRosterCompactThreshold = 120
	agentRosterDefaultWidth     = 120
	agentRosterPanelGap         = 1
)

// AgentRosterLayout identifies standard vs compact rendering.
type AgentRosterLayout string

const (
	// AgentRosterLayoutStandard renders 20/40/40 columns.
	AgentRosterLayoutStandard AgentRosterLayout = "standard"
	// AgentRosterLayoutCompact stacks filter and list for narrow terminals.
	AgentRosterLayoutCompact AgentRosterLayout = "compact"
)

// AgentRosterRoleCategory identifies role filter entries.
type AgentRosterRoleCategory string

const (
	// AgentRosterRoleAll includes all agents.
	AgentRosterRoleAll AgentRosterRoleCategory = "All"
	// AgentRosterRoleCommanders includes commander-role agents.
	AgentRosterRoleCommanders AgentRosterRoleCategory = "Commanders"
	// AgentRosterRoleCaptains includes captain-role agents.
	AgentRosterRoleCaptains AgentRosterRoleCategory = "Captains"
	// AgentRosterRoleEnsigns includes ensign-role agents.
	AgentRosterRoleEnsigns AgentRosterRoleCategory = "Ensigns"
	// AgentRosterRoleUnassigned includes unassigned-role agents.
	AgentRosterRoleUnassigned AgentRosterRoleCategory = "Unassigned"
)

// AgentRosterQuickAction identifies direct keyboard actions.
type AgentRosterQuickAction string

const (
	// AgentRosterQuickActionNone indicates no quick action.
	AgentRosterQuickActionNone AgentRosterQuickAction = ""
	// AgentRosterQuickActionNew creates an agent.
	AgentRosterQuickActionNew AgentRosterQuickAction = "new"
	// AgentRosterQuickActionEdit edits selected agent.
	AgentRosterQuickActionEdit AgentRosterQuickAction = "edit"
	// AgentRosterQuickActionAssign assigns selected agent.
	AgentRosterQuickActionAssign AgentRosterQuickAction = "assign"
	// AgentRosterQuickActionDetach detaches selected agent.
	AgentRosterQuickActionDetach AgentRosterQuickAction = "detach"
	// AgentRosterQuickActionDelete deletes selected agent.
	AgentRosterQuickActionDelete AgentRosterQuickAction = "delete"
	// AgentRosterQuickActionSearch starts search.
	AgentRosterQuickActionSearch AgentRosterQuickAction = "search"
	// AgentRosterQuickActionHelp opens help.
	AgentRosterQuickActionHelp AgentRosterQuickAction = "help"
	// AgentRosterQuickActionFleet returns to fleet view.
	AgentRosterQuickActionFleet AgentRosterQuickAction = "fleet"
)

// AgentRosterAgent captures one row/profile entry in roster view.
type AgentRosterAgent struct {
	Name          string
	Role          string
	Model         string
	Status        string
	Assignment    string
	MissionID     string
	Phase         string
	Skills        []string
	MissionPrompt string
	Backstory     string
}

// AgentRosterConfig contains render inputs for Agent Roster view.
type AgentRosterConfig struct {
	Width              int
	Agents             []AgentRosterAgent
	SelectedRoleIndex  int
	SelectedAgentIndex int
	ToolbarHighlighted int
}

type agentRosterListItem struct {
	content string
}

func (item agentRosterListItem) FilterValue() string {
	return item.content
}

type agentRosterListDelegate struct{}

func (agentRosterListDelegate) Height() int {
	return 1
}

func (agentRosterListDelegate) Spacing() int {
	return 0
}

func (agentRosterListDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd {
	return nil
}

func (agentRosterListDelegate) Render(writer io.Writer, _ list.Model, _ int, item list.Item) {
	row, ok := item.(agentRosterListItem)
	if !ok {
		return
	}
	if _, err := io.WriteString(writer, row.content); err != nil {
		return
	}
}

// ResolveAgentRosterLayout resolves compact/standard layout mode.
func ResolveAgentRosterLayout(width int) AgentRosterLayout {
	if width > 0 && width < agentRosterCompactThreshold {
		return AgentRosterLayoutCompact
	}
	return AgentRosterLayoutStandard
}

// AgentRosterToolbarButtons returns roster CRUD actions.
func AgentRosterToolbarButtons() []components.ToolbarButton {
	return []components.ToolbarButton{
		{Key: "n", Label: "New", Enabled: true},
		{Key: "Enter", Label: "Edit", Enabled: true},
		{Key: "a", Label: "Assign", Enabled: true},
		{Key: "d", Label: "Detach", Enabled: true},
		{Key: "Del", Label: "Delete", Enabled: true},
		{Key: "/", Label: "Search", Enabled: true},
		{Key: "?", Label: "Help", Enabled: true},
		{Key: "Esc", Label: "Fleet", Enabled: true},
	}
}

// AgentRosterQuickActionForKey resolves direct action keys.
func AgentRosterQuickActionForKey(msg tea.KeyMsg) AgentRosterQuickAction {
	switch strings.ToLower(strings.TrimSpace(msg.String())) {
	case "n":
		return AgentRosterQuickActionNew
	case "enter":
		return AgentRosterQuickActionEdit
	case "a":
		return AgentRosterQuickActionAssign
	case "d":
		return AgentRosterQuickActionDetach
	case "delete", "backspace":
		return AgentRosterQuickActionDelete
	case "/":
		return AgentRosterQuickActionSearch
	case "?":
		return AgentRosterQuickActionHelp
	case "esc":
		return AgentRosterQuickActionFleet
	default:
		return AgentRosterQuickActionNone
	}
}

// RenderAgentRoster renders the full roster view.
func RenderAgentRoster(config AgentRosterConfig) string {
	width := config.Width
	if width <= 0 {
		width = agentRosterDefaultWidth
	}
	layout := ResolveAgentRosterLayout(width)
	roles := agentRosterRoleCategories()
	selectedRole := normalizeSelectedIndex(config.SelectedRoleIndex, len(roles))
	if selectedRole < 0 {
		selectedRole = 0
	}

	selectedCategory := roles[selectedRole]
	filtered := FilterAgentRosterByRole(config.Agents, selectedCategory)
	agentIndex := normalizeSelectedIndex(config.SelectedAgentIndex, len(filtered))

	header := renderAgentRosterHeader(config.Agents)
	toolbar := components.RenderNavigableToolbar(AgentRosterToolbarButtons(), config.ToolbarHighlighted)

	if len(config.Agents) == 0 {
		empty := theme.PanelBorder.Render(panelWithTitle(
			"Agent List",
			lipgloss.NewStyle().Foreground(theme.ButterscotchColor).Bold(true).Render("Your roster is empty. Press [n] to recruit your first crew member."),
		))
		return lipgloss.JoinVertical(lipgloss.Left, header, empty, toolbar)
	}

	if layout == AgentRosterLayoutCompact {
		filter := renderAgentRosterFilterPanel(roles, selectedRole, config.Agents, width)
		listPanel := renderAgentRosterListPanel(filtered, agentIndex, width)
		return lipgloss.JoinVertical(lipgloss.Left, header, filter, listPanel, toolbar)
	}

	roleWidth := max(20, int(float64(width)*0.2)-agentRosterPanelGap)
	listWidth := max(36, int(float64(width)*0.4)-agentRosterPanelGap)
	detailWidth := max(36, width-roleWidth-listWidth-(agentRosterPanelGap*2))

	filter := lipgloss.NewStyle().Width(roleWidth).Render(renderAgentRosterFilterPanel(roles, selectedRole, config.Agents, roleWidth))
	listPanel := lipgloss.NewStyle().Width(listWidth).Render(renderAgentRosterListPanel(filtered, agentIndex, listWidth))
	detailPanel := lipgloss.NewStyle().Width(detailWidth).Render(renderAgentRosterDetailPanel(filtered, agentIndex, detailWidth))

	content := lipgloss.JoinHorizontal(
		lipgloss.Top,
		filter,
		lipgloss.NewStyle().Width(agentRosterPanelGap).Render(""),
		listPanel,
		lipgloss.NewStyle().Width(agentRosterPanelGap).Render(""),
		detailPanel,
	)

	return lipgloss.JoinVertical(lipgloss.Left, header, content, toolbar)
}

// FilterAgentRosterByRole applies role-filter selection to roster agents.
func FilterAgentRosterByRole(agents []AgentRosterAgent, category AgentRosterRoleCategory) []AgentRosterAgent {
	filtered := make([]AgentRosterAgent, 0, len(agents))
	for _, agent := range agents {
		if matchesAgentRosterRole(agent, category) {
			filtered = append(filtered, agent)
		}
	}
	sort.SliceStable(filtered, func(i int, j int) bool {
		return strings.ToLower(strings.TrimSpace(filtered[i].Name)) < strings.ToLower(strings.TrimSpace(filtered[j].Name))
	})
	return filtered
}

func renderAgentRosterHeader(agents []AgentRosterAgent) string {
	counts := map[AgentRosterRoleCategory]int{
		AgentRosterRoleAll:        len(agents),
		AgentRosterRoleCommanders: 0,
		AgentRosterRoleCaptains:   0,
		AgentRosterRoleEnsigns:    0,
		AgentRosterRoleUnassigned: 0,
	}
	active := 0
	idle := 0
	stuck := 0

	for _, agent := range agents {
		if matchesAgentRosterRole(agent, AgentRosterRoleCommanders) {
			counts[AgentRosterRoleCommanders]++
		}
		if matchesAgentRosterRole(agent, AgentRosterRoleCaptains) {
			counts[AgentRosterRoleCaptains]++
		}
		if matchesAgentRosterRole(agent, AgentRosterRoleEnsigns) {
			counts[AgentRosterRoleEnsigns]++
		}
		if matchesAgentRosterRole(agent, AgentRosterRoleUnassigned) {
			counts[AgentRosterRoleUnassigned]++
		}

		switch normalizeAgentRosterStatus(agent.Status) {
		case "active":
			active++
		case "stuck":
			stuck++
		default:
			idle++
		}
	}

	stuckBadge := ""
	if stuck > 0 {
		stuckBadge = " " + lipgloss.NewStyle().Foreground(theme.BlackColor).Background(theme.PinkColor).Bold(true).Render(fmt.Sprintf("[%d]", stuck))
	}

	title := lipgloss.NewStyle().Foreground(theme.BlueColor).Bold(true).Render("AGENT ROSTER")
	stats := fmt.Sprintf(
		"Agents: %d   Active: %d   Idle: %d   Stuck:%s   Commanders: %d | Captains: %d | Ensigns: %d | Unassigned: %d",
		counts[AgentRosterRoleAll],
		active,
		idle,
		stuckBadge,
		counts[AgentRosterRoleCommanders],
		counts[AgentRosterRoleCaptains],
		counts[AgentRosterRoleEnsigns],
		counts[AgentRosterRoleUnassigned],
	)

	return theme.PanelBorder.Render(lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		lipgloss.NewStyle().Foreground(theme.SpaceWhiteColor).Render(stats),
	))
}

func renderAgentRosterFilterPanel(categories []AgentRosterRoleCategory, selected int, agents []AgentRosterAgent, width int) string {
	lines := make([]string, 0, len(categories))
	for idx, category := range categories {
		count := len(FilterAgentRosterByRole(agents, category))
		prefix := "  "
		style := lipgloss.NewStyle().Foreground(theme.GalaxyGrayColor)
		if idx == selected {
			prefix = "▸ "
			style = lipgloss.NewStyle().Foreground(theme.MoonlitVioletColor).Bold(true)
		}
		lines = append(lines, style.Render(fmt.Sprintf("%s%s (%d)", prefix, string(category), count)))
	}

	content := strings.Join(lines, "\n")
	if width > 0 {
		content = lipgloss.NewStyle().Width(width - 4).Render(content)
	}
	return theme.PanelBorder.Render(panelWithTitle("Roles", content))
}

func renderAgentRosterListPanel(agents []AgentRosterAgent, selected int, width int) string {
	if len(agents) == 0 {
		empty := lipgloss.NewStyle().Foreground(theme.GalaxyGrayColor).Faint(true).Render("No agents in selected role category")
		return theme.PanelBorder.Render(panelWithTitle("Agent List", empty))
	}

	items := make([]list.Item, 0, len(agents))
	for idx, agent := range agents {
		items = append(items, agentRosterListItem{content: renderAgentRosterRow(agent, idx == selected)})
	}

	model := list.New(items, agentRosterListDelegate{}, max(30, width-4), max(8, len(items)+2))
	model.SetShowTitle(false)
	model.SetShowStatusBar(false)
	model.SetShowHelp(false)
	model.SetShowPagination(false)
	model.SetFilteringEnabled(false)
	model.Select(max(0, selected))

	return theme.PanelBorder.Render(panelWithTitle("Agent List", model.View()))
}

func renderAgentRosterRow(agent AgentRosterAgent, selected bool) string {
	status := normalizeAgentRosterStatus(agent.Status)
	statusIcon := "○"
	statusColor := theme.LightGrayColor
	if status == "active" {
		statusIcon = "▸"
		statusColor = theme.ButterscotchColor
	} else if status == "stuck" {
		statusIcon = "⚠"
		statusColor = theme.YellowCautionColor
	}

	name := strings.TrimSpace(agent.Name)
	if name == "" {
		name = "unnamed-agent"
	}
	role := strings.TrimSpace(agent.Role)
	if role == "" {
		role = "unassigned"
	}
	assignment := strings.TrimSpace(agent.Assignment)
	if assignment == "" {
		assignment = "Unassigned"
	}
	phase := strings.TrimSpace(agent.Phase)
	if phase == "" {
		phase = "IDLE"
	}

	row := strings.Join([]string{
		lipgloss.NewStyle().Foreground(statusColor).Bold(true).Render(statusIcon),
		lipgloss.NewStyle().Foreground(theme.ButterscotchColor).Bold(true).Render(name),
		lipgloss.NewStyle().Foreground(roleColor(role)).Render(strings.ToLower(role)),
		lipgloss.NewStyle().Foreground(theme.LightGrayColor).Render(strings.TrimSpace(agent.Model)),
		lipgloss.NewStyle().Foreground(theme.BlueColor).Render(assignment),
		lipgloss.NewStyle().Foreground(phaseColor(phase)).Render(strings.ToUpper(phase)),
	}, "  ")

	if !selected {
		return row
	}
	return lipgloss.NewStyle().Foreground(theme.MoonlitVioletColor).Bold(true).Render("▸ ") + row
}

func renderAgentRosterDetailPanel(agents []AgentRosterAgent, selected int, width int) string {
	if selected < 0 || selected >= len(agents) {
		return theme.PanelBorder.Render(panelWithTitle("Agent Detail", lipgloss.NewStyle().Foreground(theme.GalaxyGrayColor).Faint(true).Render("Select an agent to view profile details")))
	}

	agent := agents[selected]
	name := strings.TrimSpace(agent.Name)
	if name == "" {
		name = "unnamed-agent"
	}
	role := strings.TrimSpace(agent.Role)
	if role == "" {
		role = "unassigned"
	}
	model := strings.TrimSpace(agent.Model)
	if model == "" {
		model = "unknown"
	}
	assignment := strings.TrimSpace(agent.Assignment)
	if assignment == "" {
		assignment = "Unassigned"
	}
	mission := strings.TrimSpace(agent.MissionID)
	if mission == "" {
		mission = "-"
	}

	skills := "-"
	if len(agent.Skills) > 0 {
		parts := make([]string, 0, len(agent.Skills))
		for _, skill := range agent.Skills {
			skill = strings.TrimSpace(skill)
			if skill == "" {
				continue
			}
			parts = append(parts, skill)
		}
		if len(parts) > 0 {
			skills = strings.Join(parts, "  ")
		}
	}

	prompt := strings.TrimSpace(agent.MissionPrompt)
	if prompt == "" {
		prompt = "No mission prompt configured."
	}
	promptRendered := renderMarkdown(prompt, max(30, width-8))

	backstory := strings.TrimSpace(agent.Backstory)
	if backstory == "" {
		backstory = "No backstory provided."
	}
	backstoryRendered := renderMarkdown(backstory, max(30, width-8))

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.NewStyle().Foreground(theme.ButterscotchColor).Bold(true).Render(name),
		lipgloss.NewStyle().Foreground(roleColor(role)).Render(strings.ToLower(role)),
		lipgloss.NewStyle().Foreground(theme.GoldColor).Bold(true).Render("PROFILE"),
		lipgloss.NewStyle().Foreground(theme.LightGrayColor).Render("Model:")+" "+lipgloss.NewStyle().Foreground(theme.SpaceWhiteColor).Render(model),
		lipgloss.NewStyle().Foreground(theme.LightGrayColor).Render("Assignment:")+" "+lipgloss.NewStyle().Foreground(theme.SpaceWhiteColor).Render(assignment),
		lipgloss.NewStyle().Foreground(theme.LightGrayColor).Render("Mission:")+" "+lipgloss.NewStyle().Foreground(theme.SpaceWhiteColor).Render(mission),
		lipgloss.NewStyle().Foreground(theme.GoldColor).Bold(true).Render("SKILLS"),
		lipgloss.NewStyle().Foreground(theme.SpaceWhiteColor).Render(skills),
		lipgloss.NewStyle().Foreground(theme.GoldColor).Bold(true).Render("MISSION PROMPT"),
		promptRendered,
		lipgloss.NewStyle().Foreground(theme.GoldColor).Bold(true).Render("BACKSTORY"),
		backstoryRendered,
	)

	return theme.PanelBorder.Render(panelWithTitle("Agent Detail", content))
}

func agentRosterRoleCategories() []AgentRosterRoleCategory {
	return []AgentRosterRoleCategory{
		AgentRosterRoleAll,
		AgentRosterRoleCommanders,
		AgentRosterRoleCaptains,
		AgentRosterRoleEnsigns,
		AgentRosterRoleUnassigned,
	}
}

func matchesAgentRosterRole(agent AgentRosterAgent, category AgentRosterRoleCategory) bool {
	role := strings.ToLower(strings.TrimSpace(agent.Role))
	switch category {
	case AgentRosterRoleCommanders:
		return strings.Contains(role, "commander")
	case AgentRosterRoleCaptains:
		return strings.Contains(role, "captain")
	case AgentRosterRoleEnsigns:
		return strings.Contains(role, "ensign") || strings.Contains(role, "implement") || strings.Contains(role, "review") || strings.Contains(role, "design")
	case AgentRosterRoleUnassigned:
		return role == "" || strings.Contains(role, "unassigned")
	default:
		return true
	}
}

func normalizeAgentRosterStatus(status string) string {
	normalized := strings.ToLower(strings.TrimSpace(status))
	switch normalized {
	case "active", "running", "working":
		return "active"
	case "stuck", "failed", "halted":
		return "stuck"
	default:
		return "idle"
	}
}

func roleColor(role string) lipgloss.TerminalColor {
	normalized := strings.ToLower(strings.TrimSpace(role))
	switch {
	case strings.Contains(normalized, "captain"):
		return theme.GoldColor
	case strings.Contains(normalized, "commander"):
		return theme.BlueColor
	case strings.Contains(normalized, "review"):
		return theme.PurpleColor
	case strings.Contains(normalized, "design"):
		return theme.PinkColor
	default:
		return theme.ButterscotchColor
	}
}

func phaseColor(phase string) lipgloss.TerminalColor {
	normalized := strings.ToLower(strings.TrimSpace(phase))
	switch {
	case strings.Contains(normalized, "red"):
		return theme.RedAlertColor
	case strings.Contains(normalized, "green"):
		return theme.GreenOkColor
	case strings.Contains(normalized, "review"):
		return theme.PurpleColor
	case strings.Contains(normalized, "plan"):
		return theme.BlueColor
	default:
		return theme.LightGrayColor
	}
}
