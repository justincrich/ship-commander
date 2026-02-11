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
	fleetOverviewCompactThreshold = 120
	fleetOverviewPanelGap         = 1
	fleetOverviewDefaultWidth     = 120
	fleetOverviewBridgeView       = "ship_bridge"
)

// FleetOverviewLayout is the responsive layout mode for Fleet Overview rendering.
type FleetOverviewLayout string

const (
	// FleetOverviewLayoutStandard renders list + preview side-by-side.
	FleetOverviewLayoutStandard FleetOverviewLayout = "standard"
	// FleetOverviewLayoutCompact renders only list content.
	FleetOverviewLayoutCompact FleetOverviewLayout = "compact"
)

// FleetOverviewCrewMember represents one crew member in preview content.
type FleetOverviewCrewMember struct {
	Name string
	Role string
}

// FleetOverviewShip is the full render payload for one ship card and preview.
type FleetOverviewShip struct {
	Name             string
	Class            string
	Status           string
	DirectiveTitle   string
	DirectiveSummary string
	CrewCount        int
	WaveCurrent      int
	WaveTotal        int
	MissionsDone     int
	MissionsTotal    int
	Crew             []FleetOverviewCrewMember
	RecentActivity   []string
}

// FleetOverviewConfig contains all render-time inputs for Fleet Overview.
type FleetOverviewConfig struct {
	Width              int
	Ships              []FleetOverviewShip
	SelectedIndex      int
	PendingMessages    int
	FleetHealthLabel   string
	ToolbarHighlighted int
}

// FleetOverviewQuickAction represents direct action keys for fleet overview.
type FleetOverviewQuickAction string

const (
	// FleetQuickActionNone indicates no quick action.
	FleetQuickActionNone FleetOverviewQuickAction = ""
	// FleetQuickActionNew commissions a ship.
	FleetQuickActionNew FleetOverviewQuickAction = "new"
	// FleetQuickActionDirective opens directive flow.
	FleetQuickActionDirective FleetOverviewQuickAction = "directive"
	// FleetQuickActionRoster opens roster flow.
	FleetQuickActionRoster FleetOverviewQuickAction = "roster"
	// FleetQuickActionInbox opens message center.
	FleetQuickActionInbox FleetOverviewQuickAction = "inbox"
	// FleetQuickActionSettings opens settings.
	FleetQuickActionSettings FleetOverviewQuickAction = "settings"
	// FleetQuickActionHelp opens help.
	FleetQuickActionHelp FleetOverviewQuickAction = "help"
	// FleetQuickActionQuit requests quit.
	FleetQuickActionQuit FleetOverviewQuickAction = "quit"
)

type fleetOverviewListItem struct {
	content string
}

func (item fleetOverviewListItem) FilterValue() string {
	return item.content
}

type fleetOverviewListDelegate struct{}

func (delegate fleetOverviewListDelegate) Height() int {
	return 7
}

func (delegate fleetOverviewListDelegate) Spacing() int {
	return 1
}

func (delegate fleetOverviewListDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd {
	return nil
}

func (delegate fleetOverviewListDelegate) Render(writer io.Writer, _ list.Model, _ int, listItem list.Item) {
	item, ok := listItem.(fleetOverviewListItem)
	if !ok {
		return
	}
	if _, err := io.WriteString(writer, item.content); err != nil {
		return
	}
}

// ResolveFleetOverviewLayout returns the responsive mode for the given width.
func ResolveFleetOverviewLayout(width int) FleetOverviewLayout {
	if width > 0 && width < fleetOverviewCompactThreshold {
		return FleetOverviewLayoutCompact
	}
	return FleetOverviewLayoutStandard
}

// FleetOverviewToolbarButtons returns the canonical wayfinding buttons for the view.
func FleetOverviewToolbarButtons() []components.ToolbarButton {
	return []components.ToolbarButton{
		{Key: "n", Label: "New", Enabled: true},
		{Key: "d", Label: "Directive", Enabled: true},
		{Key: "r", Label: "Roster", Enabled: true},
		{Key: "i", Label: "Inbox", Enabled: true},
		{Key: "s", Label: "Settings", Enabled: true},
		{Key: "?", Label: "Help", Enabled: true},
		{Key: "q", Label: "Quit", Enabled: true},
	}
}

// FleetOverviewEnterTarget returns the drill-down destination for Enter key actions.
func FleetOverviewEnterTarget() string {
	return fleetOverviewBridgeView
}

// FleetOverviewQuickActionForKey resolves direct action keys.
func FleetOverviewQuickActionForKey(msg tea.KeyMsg) FleetOverviewQuickAction {
	switch strings.ToLower(strings.TrimSpace(msg.String())) {
	case "n":
		return FleetQuickActionNew
	case "d":
		return FleetQuickActionDirective
	case "r":
		return FleetQuickActionRoster
	case "i":
		return FleetQuickActionInbox
	case "s":
		return FleetQuickActionSettings
	case "?":
		return FleetQuickActionHelp
	case "q":
		return FleetQuickActionQuit
	default:
		return FleetQuickActionNone
	}
}

// NextFleetSelection returns the next selection index for Up/Down navigation.
func NextFleetSelection(ships []FleetOverviewShip, current int) int {
	if len(ships) == 0 {
		return -1
	}
	if current < 0 || current >= len(ships) {
		return 0
	}
	return (current + 1) % len(ships)
}

// PreviousFleetSelection returns the previous selection index for Up/Down navigation.
func PreviousFleetSelection(ships []FleetOverviewShip, current int) int {
	if len(ships) == 0 {
		return -1
	}
	if current < 0 || current >= len(ships) {
		return 0
	}
	return (current - 1 + len(ships)) % len(ships)
}

// RenderFleetOverview renders the Fleet Overview view in standard or compact mode.
func RenderFleetOverview(config FleetOverviewConfig) string {
	width := config.Width
	if width <= 0 {
		width = fleetOverviewDefaultWidth
	}

	layout := ResolveFleetOverviewLayout(width)
	ships := sortShipsForOverview(config.Ships)
	selectedIndex := normalizeSelectedIndex(config.SelectedIndex, len(ships))
	selectedShip, hasSelected := selectedShip(ships, selectedIndex)

	header := renderFleetHeader(ships, config.PendingMessages, config.FleetHealthLabel)
	toolbar := components.RenderNavigableToolbar(FleetOverviewToolbarButtons(), config.ToolbarHighlighted)

	if layout == FleetOverviewLayoutCompact {
		content := renderShipListPanel(ships, selectedIndex, width)
		return lipgloss.JoinVertical(lipgloss.Left, header, content, toolbar)
	}

	leftWidth := int(float64(width)*0.6) - fleetOverviewPanelGap
	if leftWidth < 48 {
		leftWidth = 48
	}
	rightWidth := width - leftWidth - fleetOverviewPanelGap
	if rightWidth < 32 {
		rightWidth = 32
	}

	listPanel := lipgloss.NewStyle().Width(leftWidth).Render(renderShipListPanel(ships, selectedIndex, leftWidth))
	previewPanel := lipgloss.NewStyle().Width(rightWidth).Render(renderShipPreviewPanel(selectedShip, hasSelected, rightWidth))
	content := lipgloss.JoinHorizontal(lipgloss.Top, listPanel, lipgloss.NewStyle().Width(fleetOverviewPanelGap).Render(""), previewPanel)

	return lipgloss.JoinVertical(lipgloss.Left, header, content, toolbar)
}

func renderFleetHeader(ships []FleetOverviewShip, pendingMessages int, healthLabel string) string {
	totalShips := len(ships)
	launched := 0
	done := 0
	total := 0
	for _, ship := range ships {
		if shipStatusRank(ship.Status) == 0 {
			launched++
		}
		done += clampToZero(ship.MissionsDone)
		total += clampToZero(ship.MissionsTotal)
	}

	completion := 0
	if total > 0 {
		completion = int((float64(done) / float64(total)) * 100)
	}
	if healthLabel == "" {
		healthLabel = "Optimal"
	}

	title := lipgloss.NewStyle().
		Foreground(theme.GoldColor).
		Bold(true).
		Render("FLEET COMMAND")

	messageBadge := ""
	if pendingMessages > 0 {
		messageBadge = " " + lipgloss.NewStyle().
			Background(theme.PinkColor).
			Foreground(theme.BlackColor).
			Bold(true).
			Render(fmt.Sprintf("[%d]", pendingMessages))
	}

	metrics := strings.Join([]string{
		fmt.Sprintf("Ships: %d", totalShips),
		fmt.Sprintf("Launched: %d", launched),
		"Messages:" + messageBadge,
		fmt.Sprintf("Fleet Health: %s %s", lipgloss.NewStyle().Foreground(theme.GreenOkColor).Render("●"), strings.TrimSpace(healthLabel)),
		fmt.Sprintf("Completion: %d%%", completion),
	}, "   ")

	return theme.PanelBorder.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			title,
			lipgloss.NewStyle().Foreground(theme.SpaceWhiteColor).Render(metrics),
		),
	)
}

func renderShipListPanel(ships []FleetOverviewShip, selectedIndex int, width int) string {
	if len(ships) == 0 {
		empty := lipgloss.NewStyle().
			Foreground(theme.ButterscotchColor).
			Bold(true).
			Render("Your fleet is empty. Press [n] to commission your first starship.")
		return theme.PanelBorder.Render(panelWithTitle("Ship List", empty))
	}

	items := make([]list.Item, 0, len(ships))
	for i, ship := range ships {
		items = append(items, fleetOverviewListItem{
			content: renderShipCard(ship, i == selectedIndex, width-6),
		})
	}

	listHeight := len(items)*8 + 1
	if listHeight < 12 {
		listHeight = 12
	}
	model := list.New(items, fleetOverviewListDelegate{}, width-4, listHeight)
	model.SetShowTitle(false)
	model.SetShowStatusBar(false)
	model.SetShowHelp(false)
	model.SetShowPagination(false)
	model.SetFilteringEnabled(false)
	if len(items) > 0 {
		model.Select(selectedIndex)
	}
	content := model.View()
	return theme.PanelBorder.Render(panelWithTitle("Ship List", content))
}

func renderShipCard(ship FleetOverviewShip, selected bool, width int) string {
	name := strings.TrimSpace(ship.Name)
	if name == "" {
		name = "Unnamed Ship"
	}
	class := strings.TrimSpace(ship.Class)
	if class == "" {
		class = "Unknown-class"
	}

	directive := strings.TrimSpace(ship.DirectiveTitle)
	directiveStyle := lipgloss.NewStyle().Foreground(theme.BlueColor)
	if directive == "" {
		directive = "No Directive"
		directiveStyle = lipgloss.NewStyle().Foreground(theme.GalaxyGrayColor).Faint(true)
	}

	status := components.RenderStatusBadge(mapStatusToBadge(ship.Status), components.WithBadgeBold(true))
	progress := components.RenderWaveProgressBar(components.WaveProgressBarConfig{
		WaveNumber: shipWaveNumber(ship.WaveCurrent),
		Completed:  clampToZero(ship.MissionsDone),
		Total:      clampToZero(ship.MissionsTotal),
		Width:      18,
	})

	titleRow := lipgloss.JoinHorizontal(
		lipgloss.Left,
		lipgloss.NewStyle().Foreground(theme.ButterscotchColor).Bold(true).Render(name),
		"  ",
		lipgloss.NewStyle().Foreground(theme.GalaxyGrayColor).Faint(true).Render(class),
		"  ",
		status,
	)

	body := lipgloss.JoinVertical(
		lipgloss.Left,
		titleRow,
		directiveStyle.Render(directive),
		lipgloss.NewStyle().Foreground(theme.LightGrayColor).Render(fmt.Sprintf("Crew: %d   Mission %d/%d", clampToZero(ship.CrewCount), clampToZero(ship.MissionsDone), clampToZero(ship.MissionsTotal))),
		lipgloss.NewStyle().Foreground(theme.SpaceWhiteColor).Render(progress),
	)

	cardStyle := theme.PanelBorder
	if selected {
		cardStyle = theme.PanelBorderFocused
	}
	if width > 0 {
		body = lipgloss.NewStyle().Width(width).Render(body)
	}
	return cardStyle.Render(body)
}

func renderShipPreviewPanel(ship FleetOverviewShip, hasSelected bool, width int) string {
	if !hasSelected {
		return theme.PanelBorder.Render(panelWithTitle("Ship Preview", lipgloss.NewStyle().Foreground(theme.GalaxyGrayColor).Faint(true).Render("Select a ship to preview details.")))
	}

	name := strings.TrimSpace(ship.Name)
	if name == "" {
		name = "Unnamed Ship"
	}
	class := strings.TrimSpace(ship.Class)
	if class == "" {
		class = "Unknown-class"
	}

	directive := strings.TrimSpace(ship.DirectiveSummary)
	if directive == "" {
		directive = strings.TrimSpace(ship.DirectiveTitle)
	}
	if directive == "" {
		directive = "No directive assigned."
	}

	crewLines := []string{"Crew Roster"}
	if len(ship.Crew) == 0 {
		crewLines = append(crewLines, lipgloss.NewStyle().Foreground(theme.GalaxyGrayColor).Faint(true).Render("No crew assigned"))
	} else {
		for _, member := range ship.Crew {
			crewLines = append(crewLines, fmt.Sprintf("%s (%s)", strings.TrimSpace(member.Name), strings.TrimSpace(member.Role)))
		}
	}

	activityLines := []string{"Recent Activity"}
	if len(ship.RecentActivity) == 0 {
		activityLines = append(activityLines, lipgloss.NewStyle().Foreground(theme.GalaxyGrayColor).Faint(true).Render("No recent activity"))
	} else {
		limit := len(ship.RecentActivity)
		if limit > 5 {
			limit = 5
		}
		for i := 0; i < limit; i++ {
			activityLines = append(activityLines, ship.RecentActivity[i])
		}
	}

	progress := components.RenderWaveProgressBar(components.WaveProgressBarConfig{
		WaveNumber: shipWaveNumber(ship.WaveCurrent),
		Completed:  clampToZero(ship.MissionsDone),
		Total:      clampToZero(ship.MissionsTotal),
		Width:      16,
	})

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.NewStyle().Foreground(theme.ButterscotchColor).Bold(true).Render(name),
		lipgloss.NewStyle().Foreground(theme.GalaxyGrayColor).Render(class),
		"",
		lipgloss.NewStyle().Foreground(theme.GoldColor).Bold(true).Render("DIRECTIVE"),
		lipgloss.NewStyle().Foreground(theme.SpaceWhiteColor).Render(directive),
		"",
		lipgloss.NewStyle().Foreground(theme.GoldColor).Bold(true).Render(strings.Join(crewLines, "\n")),
		"",
		lipgloss.NewStyle().Foreground(theme.GoldColor).Bold(true).Render("MISSION PROGRESS"),
		lipgloss.NewStyle().Foreground(theme.SpaceWhiteColor).Render(progress),
		"",
		lipgloss.NewStyle().Foreground(theme.GoldColor).Bold(true).Render(strings.Join(activityLines, "\n")),
	)

	if width > 0 {
		content = lipgloss.NewStyle().Width(width - 4).Render(content)
	}
	return theme.PanelBorder.Render(panelWithTitle("Ship Preview", content))
}

func panelWithTitle(title string, content string) string {
	return lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.NewStyle().Foreground(theme.GalaxyGrayColor).Bold(true).Render(title),
		content,
	)
}

func sortShipsForOverview(ships []FleetOverviewShip) []FleetOverviewShip {
	sorted := append([]FleetOverviewShip(nil), ships...)
	sort.SliceStable(sorted, func(i int, j int) bool {
		leftRank := shipStatusRank(sorted[i].Status)
		rightRank := shipStatusRank(sorted[j].Status)
		if leftRank != rightRank {
			return leftRank < rightRank
		}
		return strings.ToLower(strings.TrimSpace(sorted[i].Name)) < strings.ToLower(strings.TrimSpace(sorted[j].Name))
	})
	return sorted
}

func shipStatusRank(status string) int {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "launched", "running":
		return 0
	case "docked", "waiting":
		return 1
	case "complete", "completed", "done":
		return 2
	default:
		return 3
	}
}

func mapStatusToBadge(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "launched", "running":
		return "running"
	case "docked", "waiting":
		return "waiting"
	case "complete", "completed", "done":
		return "done"
	case "halted", "failed":
		return "halted"
	default:
		return "waiting"
	}
}

func selectedShip(ships []FleetOverviewShip, index int) (FleetOverviewShip, bool) {
	if index < 0 || index >= len(ships) {
		return FleetOverviewShip{}, false
	}
	return ships[index], true
}

func normalizeSelectedIndex(index int, count int) int {
	if count == 0 {
		return -1
	}
	if index < 0 || index >= count {
		return 0
	}
	return index
}

func shipWaveNumber(value int) int {
	if value <= 0 {
		return 1
	}
	return value
}

func clampToZero(value int) int {
	if value < 0 {
		return 0
	}
	return value
}
