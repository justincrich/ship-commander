package views

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ship-commander/sc3/internal/tui/components"
	"github.com/ship-commander/sc3/internal/tui/theme"
)

const (
	shipBridgeCompactThreshold = 120
	shipBridgePanelGap         = 1
	shipBridgeDefaultWidth     = 120
)

// ShipBridgeLayout defines responsive rendering mode for ship bridge.
type ShipBridgeLayout string

const (
	// ShipBridgeLayoutStandard renders crew and mission panels side by side.
	ShipBridgeLayoutStandard ShipBridgeLayout = "standard"
	// ShipBridgeLayoutCompact renders stacked panels for narrow terminals.
	ShipBridgeLayoutCompact ShipBridgeLayout = "compact"
)

// ShipBridgeStatus identifies the ship lifecycle state used for toolbar variants.
type ShipBridgeStatus string

const (
	// ShipBridgeStatusDocked uses planning toolbar actions.
	ShipBridgeStatusDocked ShipBridgeStatus = "docked"
	// ShipBridgeStatusLaunched uses execution toolbar actions.
	ShipBridgeStatusLaunched ShipBridgeStatus = "launched"
	// ShipBridgeStatusComplete indicates all mission work is complete.
	ShipBridgeStatusComplete ShipBridgeStatus = "complete"
	// ShipBridgeStatusHalted indicates execution is paused due to failure/halt.
	ShipBridgeStatusHalted ShipBridgeStatus = "halted"
)

// ShipBridgeCrewMember captures one crew row in the bridge panel.
type ShipBridgeCrewMember struct {
	Name      string
	Role      string
	MissionID string
	Phase     string
	Elapsed   string
	Status    string
}

// ShipBridgeMission captures one mission row in the mission board.
type ShipBridgeMission struct {
	ID             string
	Title          string
	Column         string
	Classification string
	AssignedAgent  string
	Phase          string
	ACCompleted    int
	ACTotal        int
	Stuck          bool
}

// ShipBridgeEvent captures one event log line.
type ShipBridgeEvent struct {
	Timestamp string
	Severity  string
	Actor     string
	Message   string
}

// ShipBridgeConfig contains all render inputs for the ship bridge view.
type ShipBridgeConfig struct {
	Width                int
	ShipName             string
	ShipClass            string
	DirectiveTitle       string
	Status               ShipBridgeStatus
	FleetHealthLabel     string
	PendingQuestions     int
	WaveCurrent          int
	WaveTotal            int
	MissionsDone         int
	MissionsTotal        int
	Crew                 []ShipBridgeCrewMember
	SelectedCrewIndex    int
	Missions             []ShipBridgeMission
	SelectedMissionIndex int
	Events               []ShipBridgeEvent
	ToolbarHighlighted   int
}

// ShipBridgeQuickAction captures direct keyboard actions supported in this view.
type ShipBridgeQuickAction string

const (
	// ShipBridgeQuickActionNone indicates no matched quick action.
	ShipBridgeQuickActionNone ShipBridgeQuickAction = ""
	// ShipBridgeQuickActionPlan opens planning flow.
	ShipBridgeQuickActionPlan ShipBridgeQuickAction = "plan"
	// ShipBridgeQuickActionAssign opens crew assignment flow.
	ShipBridgeQuickActionAssign ShipBridgeQuickAction = "assign"
	// ShipBridgeQuickActionLaunch launches a docked ship.
	ShipBridgeQuickActionLaunch ShipBridgeQuickAction = "launch"
	// ShipBridgeQuickActionWave opens wave manager.
	ShipBridgeQuickActionWave ShipBridgeQuickAction = "wave"
	// ShipBridgeQuickActionHelp opens help overlay.
	ShipBridgeQuickActionHelp ShipBridgeQuickAction = "help"
	// ShipBridgeQuickActionFleet returns to fleet overview.
	ShipBridgeQuickActionFleet ShipBridgeQuickAction = "fleet"
	// ShipBridgeQuickActionHalt halts launched mission execution.
	ShipBridgeQuickActionHalt ShipBridgeQuickAction = "halt"
	// ShipBridgeQuickActionRetry retries halted work.
	ShipBridgeQuickActionRetry ShipBridgeQuickAction = "retry"
	// ShipBridgeQuickActionDock docks a launched ship.
	ShipBridgeQuickActionDock ShipBridgeQuickAction = "dock"
)

type shipBridgeListItem struct {
	content string
}

func (item shipBridgeListItem) FilterValue() string {
	return item.content
}

type shipBridgeCrewDelegate struct{}

func (delegate shipBridgeCrewDelegate) Height() int {
	return 5
}

func (delegate shipBridgeCrewDelegate) Spacing() int {
	return 1
}

func (delegate shipBridgeCrewDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd {
	return nil
}

func (delegate shipBridgeCrewDelegate) Render(writer io.Writer, _ list.Model, _ int, listItem list.Item) {
	item, ok := listItem.(shipBridgeListItem)
	if !ok {
		return
	}
	if _, err := io.WriteString(writer, item.content); err != nil {
		return
	}
}

type shipBridgeMissionDelegate struct{}

func (delegate shipBridgeMissionDelegate) Height() int {
	return 4
}

func (delegate shipBridgeMissionDelegate) Spacing() int {
	return 1
}

func (delegate shipBridgeMissionDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd {
	return nil
}

func (delegate shipBridgeMissionDelegate) Render(writer io.Writer, _ list.Model, _ int, listItem list.Item) {
	item, ok := listItem.(shipBridgeListItem)
	if !ok {
		return
	}
	if _, err := io.WriteString(writer, item.content); err != nil {
		return
	}
}

// ResolveShipBridgeLayout returns compact or standard layout based on terminal width.
func ResolveShipBridgeLayout(width int) ShipBridgeLayout {
	if width > 0 && width < shipBridgeCompactThreshold {
		return ShipBridgeLayoutCompact
	}
	return ShipBridgeLayoutStandard
}

// ShipBridgeToolbarButtons returns the context-sensitive toolbar for ship status.
func ShipBridgeToolbarButtons(status ShipBridgeStatus) []components.ToolbarButton {
	normalized := normalizeShipBridgeStatus(status)
	if normalized == ShipBridgeStatusDocked {
		return []components.ToolbarButton{
			{Key: "p", Label: "Plan", Enabled: true},
			{Key: "a", Label: "Assign", Enabled: true},
			{Key: "l", Label: "Launch", Enabled: true},
			{Key: "w", Label: "Wave", Enabled: true},
			{Key: "?", Label: "Help", Enabled: true},
			{Key: "Esc", Label: "Fleet", Enabled: true},
		}
	}

	return []components.ToolbarButton{
		{Key: "h", Label: "Halt", Enabled: true},
		{Key: "r", Label: "Retry", Enabled: true},
		{Key: "w", Label: "Wave", Enabled: true},
		{Key: "d", Label: "Dock", Enabled: true},
		{Key: "?", Label: "Help", Enabled: true},
		{Key: "Esc", Label: "Fleet", Enabled: true},
	}
}

// ShipBridgeEnterTargetForPanel resolves Enter-key drill-down by focused panel.
func ShipBridgeEnterTargetForPanel(panelID string) string {
	switch strings.ToLower(strings.TrimSpace(panelID)) {
	case "crew_panel":
		return "agent_detail"
	case "mission_board":
		return "mission_detail"
	default:
		return ""
	}
}

// ShipBridgeQuickActionForKey resolves direct action keys for docked/launched states.
func ShipBridgeQuickActionForKey(msg tea.KeyMsg, status ShipBridgeStatus) ShipBridgeQuickAction {
	key := strings.ToLower(strings.TrimSpace(msg.String()))
	normalized := normalizeShipBridgeStatus(status)

	switch key {
	case "?":
		return ShipBridgeQuickActionHelp
	case "esc":
		return ShipBridgeQuickActionFleet
	case "w":
		return ShipBridgeQuickActionWave
	}

	if normalized == ShipBridgeStatusDocked {
		switch key {
		case "p":
			return ShipBridgeQuickActionPlan
		case "a":
			return ShipBridgeQuickActionAssign
		case "l":
			return ShipBridgeQuickActionLaunch
		default:
			return ShipBridgeQuickActionNone
		}
	}

	switch key {
	case "h":
		return ShipBridgeQuickActionHalt
	case "r":
		return ShipBridgeQuickActionRetry
	case "d":
		return ShipBridgeQuickActionDock
	default:
		return ShipBridgeQuickActionNone
	}
}

// RenderShipBridge renders the Ship Bridge dashboard in standard or compact layout.
func RenderShipBridge(config ShipBridgeConfig) string {
	width := config.Width
	if width <= 0 {
		width = shipBridgeDefaultWidth
	}

	layout := ResolveShipBridgeLayout(width)
	status := normalizeShipBridgeStatus(config.Status)

	selectedCrew := normalizeSelectedIndex(config.SelectedCrewIndex, len(config.Crew))
	selectedMission := normalizeSelectedIndex(config.SelectedMissionIndex, len(config.Missions))

	header := renderShipBridgeHeader(config, status)
	toolbar := components.RenderNavigableToolbar(ShipBridgeToolbarButtons(status), config.ToolbarHighlighted)

	if layout == ShipBridgeLayoutCompact {
		crewPanel := renderCrewPanel(config.Crew, selectedCrew, width)
		missionPanel := renderMissionBoardPanel(config.Missions, selectedMission, status, width)
		eventPanel := renderEventLogPanel(config.Events, width, 4)
		return lipgloss.JoinVertical(lipgloss.Left, header, crewPanel, missionPanel, eventPanel, toolbar)
	}

	leftWidth := int(float64(width)*0.4) - shipBridgePanelGap
	if leftWidth < 36 {
		leftWidth = 36
	}
	rightWidth := width - leftWidth - shipBridgePanelGap
	if rightWidth < 52 {
		rightWidth = 52
	}

	crewPanel := lipgloss.NewStyle().Width(leftWidth).Render(renderCrewPanel(config.Crew, selectedCrew, leftWidth))
	missionPanel := lipgloss.NewStyle().Width(rightWidth).Render(renderMissionBoardPanel(config.Missions, selectedMission, status, rightWidth))
	panelRow := lipgloss.JoinHorizontal(lipgloss.Top, crewPanel, lipgloss.NewStyle().Width(shipBridgePanelGap).Render(""), missionPanel)
	eventPanel := renderEventLogPanel(config.Events, width, 5)

	return lipgloss.JoinVertical(lipgloss.Left, header, panelRow, eventPanel, toolbar)
}

func renderShipBridgeHeader(config ShipBridgeConfig, status ShipBridgeStatus) string {
	shipName := strings.TrimSpace(config.ShipName)
	if shipName == "" {
		shipName = "Unnamed Ship"
	}
	shipClass := strings.TrimSpace(config.ShipClass)
	if shipClass == "" {
		shipClass = "Unknown-class"
	}
	directive := strings.TrimSpace(config.DirectiveTitle)
	if directive == "" {
		directive = "No Directive"
	}

	rowOne := lipgloss.JoinHorizontal(
		lipgloss.Left,
		lipgloss.NewStyle().Foreground(theme.ButterscotchColor).Bold(true).Render(shipName),
		"  ",
		lipgloss.NewStyle().Foreground(theme.GalaxyGrayColor).Faint(true).Render(shipClass),
		"  ",
		lipgloss.NewStyle().Foreground(theme.BlueColor).Render("Directive: "+directive),
		"  ",
		components.RenderStatusBadge(mapShipStatusToBadge(status), components.WithBadgeBold(true)),
	)

	pendingBadge := ""
	if config.PendingQuestions > 0 {
		pendingBadge = " " + lipgloss.NewStyle().
			Background(theme.GoldColor).
			Foreground(theme.BlackColor).
			Bold(true).
			Render(fmt.Sprintf("[%d]", config.PendingQuestions))
	}

	crewCount := len(config.Crew)
	waveSummary := renderInlineWaveSummary(config.WaveCurrent, config.WaveTotal, config.MissionsDone, config.MissionsTotal)
	if strings.TrimSpace(config.FleetHealthLabel) == "" {
		config.FleetHealthLabel = "Optimal"
	}

	rowTwo := strings.Join([]string{
		fmt.Sprintf("Health: %s %s", renderHealthDots(status), strings.TrimSpace(config.FleetHealthLabel)),
		fmt.Sprintf("Crew: %d", crewCount),
		fmt.Sprintf("Missions: %d/%d", clampToZero(config.MissionsDone), clampToZero(config.MissionsTotal)),
		waveSummary,
		"Questions:" + pendingBadge,
	}, "   ")

	return theme.PanelBorder.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			rowOne,
			lipgloss.NewStyle().Foreground(theme.SpaceWhiteColor).Render(rowTwo),
		),
	)
}

func renderCrewPanel(crew []ShipBridgeCrewMember, selected int, width int) string {
	if len(crew) == 0 {
		empty := lipgloss.NewStyle().Foreground(theme.GalaxyGrayColor).Faint(true).Render("No crew assigned. Press [a] to assign crew.")
		return theme.PanelBorder.Render(panelWithTitle("Crew", empty))
	}

	items := make([]list.Item, 0, len(crew))
	for i, member := range crew {
		items = append(items, shipBridgeListItem{content: renderCrewCard(member, i == selected, width-6)})
	}

	listHeight := len(items)*6 + 1
	if listHeight < 10 {
		listHeight = 10
	}

	model := list.New(items, shipBridgeCrewDelegate{}, width-4, listHeight)
	model.SetShowTitle(false)
	model.SetShowStatusBar(false)
	model.SetShowHelp(false)
	model.SetShowPagination(false)
	model.SetFilteringEnabled(false)
	model.Select(selected)

	return theme.PanelBorder.Render(panelWithTitle(fmt.Sprintf("Crew (%d)", len(crew)), model.View()))
}

func renderCrewCard(member ShipBridgeCrewMember, selected bool, width int) string {
	name := strings.TrimSpace(member.Name)
	if name == "" {
		name = "Unnamed Agent"
	}
	role := renderCrewRoleBadge(member.Role)
	mission := strings.TrimSpace(member.MissionID)
	if mission == "" {
		mission = "Unassigned"
	}
	phase := strings.TrimSpace(member.Phase)
	if phase == "" {
		phase = "IDLE"
	}
	elapsed := strings.TrimSpace(member.Elapsed)
	if elapsed == "" {
		elapsed = "--:--"
	}
	statusBadge := components.RenderStatusBadge(mapCrewStatusToBadge(member.Status), components.WithBadgeBold(true))

	body := lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.JoinHorizontal(
			lipgloss.Left,
			lipgloss.NewStyle().Foreground(theme.SpaceWhiteColor).Bold(true).Render(name),
			"  ",
			role,
			"  ",
			statusBadge,
		),
		lipgloss.NewStyle().Foreground(theme.BlueColor).Render("Mission: "+mission),
		lipgloss.NewStyle().Foreground(theme.LightGrayColor).Render(fmt.Sprintf("Phase: %s   Elapsed: %s", phase, elapsed)),
	)

	if width > 0 {
		body = lipgloss.NewStyle().Width(width).Render(body)
	}

	cardStyle := theme.PanelBorder
	if selected {
		cardStyle = theme.PanelBorderFocused
	}

	return cardStyle.Render(body)
}

func renderCrewRoleBadge(role string) string {
	label := strings.ToUpper(strings.TrimSpace(role))
	if label == "" {
		label = "ENSIGN"
	}

	style := lipgloss.NewStyle().
		Background(theme.ButterscotchColor).
		Foreground(theme.BlackColor).
		Bold(true)

	switch label {
	case "CAPTAIN":
		style = lipgloss.NewStyle().Background(theme.GoldColor).Foreground(theme.BlackColor).Bold(true)
	case "COMMANDER":
		style = lipgloss.NewStyle().Background(theme.BlueColor).Foreground(theme.BlackColor).Bold(true)
	}

	return style.Render(label)
}

func renderMissionBoardPanel(missions []ShipBridgeMission, selected int, status ShipBridgeStatus, width int) string {
	summary := renderMissionSummary(missions)

	if len(missions) == 0 {
		empty := lipgloss.NewStyle().Foreground(theme.GalaxyGrayColor).Faint(true).Render(missionBoardEmptyMessage(status))
		content := lipgloss.JoinVertical(lipgloss.Left, summary, empty)
		return theme.PanelBorder.Render(panelWithTitle("Mission Board", content))
	}

	sorted := sortMissionBoardMissions(missions)
	items := make([]list.Item, 0, len(sorted))
	for i, mission := range sorted {
		items = append(items, shipBridgeListItem{content: renderMissionCard(mission, i == selected, width-6)})
	}

	listHeight := len(items)*5 + 1
	if listHeight < 10 {
		listHeight = 10
	}

	model := list.New(items, shipBridgeMissionDelegate{}, width-4, listHeight)
	model.SetShowTitle(false)
	model.SetShowStatusBar(false)
	model.SetShowHelp(false)
	model.SetShowPagination(false)
	model.SetFilteringEnabled(false)
	model.Select(selected)

	content := lipgloss.JoinVertical(lipgloss.Left, summary, model.View())
	return theme.PanelBorder.Render(panelWithTitle("Mission Board", content))
}

func renderMissionSummary(missions []ShipBridgeMission) string {
	counts := map[string]int{"B": 0, "IP": 0, "R": 0, "D": 0, "H": 0}
	for _, mission := range missions {
		counts[missionBoardColumnKey(mission.Column)]++
	}

	return strings.Join([]string{
		lipgloss.NewStyle().Foreground(theme.GalaxyGrayColor).Bold(true).Render(fmt.Sprintf("B:%d", counts["B"])),
		lipgloss.NewStyle().Foreground(theme.ButterscotchColor).Bold(true).Render(fmt.Sprintf("IP:%d", counts["IP"])),
		lipgloss.NewStyle().Foreground(theme.PurpleColor).Bold(true).Render(fmt.Sprintf("R:%d", counts["R"])),
		lipgloss.NewStyle().Foreground(theme.GreenOkColor).Bold(true).Render(fmt.Sprintf("D:%d", counts["D"])),
		lipgloss.NewStyle().Foreground(theme.RedAlertColor).Bold(true).Render(fmt.Sprintf("H:%d", counts["H"])),
	}, "    ")
}

func sortMissionBoardMissions(missions []ShipBridgeMission) []ShipBridgeMission {
	sorted := append([]ShipBridgeMission(nil), missions...)
	sort.SliceStable(sorted, func(i int, j int) bool {
		left := missionBoardColumnRank(sorted[i].Column)
		right := missionBoardColumnRank(sorted[j].Column)
		if left != right {
			return left < right
		}
		return strings.ToLower(strings.TrimSpace(sorted[i].ID)) < strings.ToLower(strings.TrimSpace(sorted[j].ID))
	})
	return sorted
}

func renderMissionCard(mission ShipBridgeMission, selected bool, width int) string {
	id := strings.TrimSpace(mission.ID)
	if id == "" {
		id = "M-000"
	}
	title := strings.TrimSpace(mission.Title)
	if title == "" {
		title = "Untitled mission"
	}
	if len(title) > 42 {
		title = title[:39] + "..."
	}

	classification := strings.ToUpper(strings.TrimSpace(mission.Classification))
	if classification == "" {
		classification = "STANDARD_OPS"
	}

	classificationStyle := lipgloss.NewStyle().Foreground(theme.BlueColor).Bold(true)
	if classification == "RED_ALERT" {
		classificationStyle = lipgloss.NewStyle().Foreground(theme.RedAlertColor).Bold(true)
	}

	agent := strings.TrimSpace(mission.AssignedAgent)
	if agent == "" {
		agent = "Unassigned"
	}
	phase := strings.TrimSpace(mission.Phase)
	if phase == "" {
		phase = "PENDING"
	}

	progress := fmt.Sprintf("AC %d/%d", clampToZero(mission.ACCompleted), clampToZero(mission.ACTotal))
	if mission.ACTotal == 0 {
		progress = "AC 0/0"
	}

	stuck := ""
	if mission.Stuck {
		stuck = "  " + lipgloss.NewStyle().Foreground(theme.YellowCautionColor).Bold(true).Render(theme.IconAlert+" STUCK")
	}

	body := lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.JoinHorizontal(
			lipgloss.Left,
			lipgloss.NewStyle().Foreground(theme.ButterscotchColor).Bold(true).Render(id),
			"  ",
			classificationStyle.Render(classification),
			stuck,
		),
		lipgloss.NewStyle().Foreground(theme.SpaceWhiteColor).Render(title),
		lipgloss.NewStyle().Foreground(theme.LightGrayColor).Render(fmt.Sprintf("Agent: %s   Phase: %s   %s", agent, phase, progress)),
	)

	if width > 0 {
		body = lipgloss.NewStyle().Width(width).Render(body)
	}

	cardStyle := theme.PanelBorder
	if selected {
		cardStyle = theme.PanelBorderFocused
	}

	return cardStyle.Render(body)
}

func renderEventLogPanel(events []ShipBridgeEvent, width int, height int) string {
	if len(events) == 0 {
		empty := lipgloss.NewStyle().Foreground(theme.GalaxyGrayColor).Faint(true).Render("No recent events")
		return theme.PanelBorder.Render(panelWithTitle("Event Log", empty))
	}

	start := 0
	if len(events) > 50 {
		start = len(events) - 50
	}

	lines := make([]string, 0, len(events)-start)
	for _, event := range events[start:] {
		lines = append(lines, renderEventLine(event))
	}

	viewWidth := width - 6
	if viewWidth < 24 {
		viewWidth = 24
	}
	viewHeight := height
	if viewHeight < 4 {
		viewHeight = 4
	}

	logViewport := viewport.New(viewWidth, viewHeight)
	logViewport.SetContent(strings.Join(lines, "\n"))
	logViewport.GotoBottom()

	return theme.PanelBorder.Render(panelWithTitle("Event Log", logViewport.View()))
}

func renderEventLine(event ShipBridgeEvent) string {
	timestamp := strings.TrimSpace(event.Timestamp)
	if timestamp == "" {
		timestamp = "--:--:--"
	}
	actor := strings.TrimSpace(event.Actor)
	if actor == "" {
		actor = "system"
	}
	message := strings.TrimSpace(event.Message)
	if message == "" {
		message = "(no message)"
	}

	severity := strings.ToLower(strings.TrimSpace(event.Severity))
	icon := "ℹ"
	style := lipgloss.NewStyle().Foreground(theme.BlueColor)
	switch severity {
	case "warn", "warning":
		icon = theme.IconAlert
		style = lipgloss.NewStyle().Foreground(theme.YellowCautionColor)
	case "error", "failed":
		icon = theme.IconFailed
		style = lipgloss.NewStyle().Foreground(theme.RedAlertColor)
	case "success", "done":
		icon = theme.IconDone
		style = lipgloss.NewStyle().Foreground(theme.GreenOkColor)
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		lipgloss.NewStyle().Foreground(theme.LightGrayColor).Render(timestamp),
		"  ",
		style.Render(icon),
		"  ",
		lipgloss.NewStyle().Foreground(theme.GalaxyGrayColor).Render(actor),
		"  ",
		style.Render(message),
	)
}

func missionBoardColumnRank(column string) int {
	switch missionBoardColumnKey(column) {
	case "B":
		return 0
	case "IP":
		return 1
	case "R":
		return 2
	case "D":
		return 3
	case "H":
		return 4
	default:
		return 5
	}
}

func missionBoardColumnKey(column string) string {
	switch strings.ToLower(strings.TrimSpace(column)) {
	case "backlog", "b":
		return "B"
	case "in_progress", "ip", "in-progress", "inprogress":
		return "IP"
	case "review", "r":
		return "R"
	case "done", "complete", "d":
		return "D"
	case "halted", "h":
		return "H"
	default:
		return "B"
	}
}

func missionBoardEmptyMessage(status ShipBridgeStatus) string {
	if normalizeShipBridgeStatus(status) == ShipBridgeStatusDocked {
		return "No missions yet. Press [p] to plan."
	}
	return "No active missions in this ship."
}

func mapShipStatusToBadge(status ShipBridgeStatus) string {
	switch normalizeShipBridgeStatus(status) {
	case ShipBridgeStatusDocked:
		return "waiting"
	case ShipBridgeStatusComplete:
		return "done"
	case ShipBridgeStatusHalted:
		return "halted"
	default:
		return "running"
	}
}

func mapCrewStatusToBadge(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "done", "complete", "completed":
		return "done"
	case "stuck":
		return "stuck"
	case "halted", "failed":
		return "halted"
	case "idle", "waiting", "docked":
		return "waiting"
	default:
		return "running"
	}
}

func normalizeShipBridgeStatus(status ShipBridgeStatus) ShipBridgeStatus {
	switch strings.ToLower(strings.TrimSpace(string(status))) {
	case "docked", "waiting":
		return ShipBridgeStatusDocked
	case "complete", "completed", "done":
		return ShipBridgeStatusComplete
	case "halted", "failed":
		return ShipBridgeStatusHalted
	default:
		return ShipBridgeStatusLaunched
	}
}

func renderHealthDots(status ShipBridgeStatus) string {
	normalized := normalizeShipBridgeStatus(status)
	filled := 5
	switch normalized {
	case ShipBridgeStatusHalted:
		filled = 2
	case ShipBridgeStatusDocked:
		filled = 4
	case ShipBridgeStatusComplete:
		filled = 5
	}

	dots := make([]string, 0, 5)
	for i := 0; i < 5; i++ {
		if i < filled {
			dots = append(dots, lipgloss.NewStyle().Foreground(theme.GreenOkColor).Render(theme.IconWorking))
			continue
		}
		dots = append(dots, lipgloss.NewStyle().Foreground(theme.GalaxyGrayColor).Render(theme.IconWorking))
	}
	return strings.Join(dots, "")
}

func renderInlineWaveSummary(current int, total int, done int, missionsTotal int) string {
	waveCurrent := shipWaveNumber(current)
	waveTotal := total
	if waveTotal < waveCurrent {
		waveTotal = waveCurrent
	}

	completed := clampToZero(done)
	totalMissions := clampToZero(missionsTotal)
	if totalMissions > 0 && completed > totalMissions {
		completed = totalMissions
	}

	fraction := 0.0
	if totalMissions > 0 {
		fraction = float64(completed) / float64(totalMissions)
	}

	barWidth := 10
	filled := int(fraction * float64(barWidth))
	if filled > barWidth {
		filled = barWidth
	}
	if filled < 0 {
		filled = 0
	}

	filledBar := lipgloss.NewStyle().Foreground(theme.ButterscotchColor).Render(strings.Repeat("=", filled))
	emptyBar := lipgloss.NewStyle().Foreground(theme.GalaxyGrayColor).Render(strings.Repeat(".", barWidth-filled))

	return fmt.Sprintf("Wave %d of %d [%s%s] %d/%d", waveCurrent, waveTotal, filledBar, emptyBar, completed, totalMissions)
}
