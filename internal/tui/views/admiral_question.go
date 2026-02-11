package views

import (
	"errors"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/harmonica"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/ship-commander/sc3/internal/admiral"
	"github.com/ship-commander/sc3/internal/tui/theme"
)

const (
	admiralQuestionDefaultWidth      = 120
	admiralQuestionDefaultHeight     = 30
	admiralQuestionCompactThreshold  = 120
	admiralQuestionStandardWidthPct  = 0.60
	admiralQuestionCompactWidthPct   = 0.80
	admiralQuestionOpenFrequency     = 6.0
	admiralQuestionOpenDampingRatio  = 0.8
	admiralQuestionCloseFrequency    = 10.0
	admiralQuestionCloseDampingRatio = 1.0

	// AdmiralQuestionSkipOption is always appended to question options.
	AdmiralQuestionSkipOption = "Skip -- agent uses best judgment"
)

// AdmiralQuestionAnimationState identifies spring profile mode.
type AdmiralQuestionAnimationState string

const (
	// AdmiralQuestionAnimationOpen uses an under-damped spring for opening.
	AdmiralQuestionAnimationOpen AdmiralQuestionAnimationState = "open"
	// AdmiralQuestionAnimationClose uses a critically-damped spring for closing.
	AdmiralQuestionAnimationClose AdmiralQuestionAnimationState = "close"
)

// AdmiralQuestionQuickAction captures direct key actions for the modal.
type AdmiralQuestionQuickAction string

const (
	// AdmiralQuestionQuickActionNone indicates no modal action.
	AdmiralQuestionQuickActionNone AdmiralQuestionQuickAction = ""
	// AdmiralQuestionQuickActionSelectPrev moves selection up.
	AdmiralQuestionQuickActionSelectPrev AdmiralQuestionQuickAction = "select_prev"
	// AdmiralQuestionQuickActionSelectNext moves selection down.
	AdmiralQuestionQuickActionSelectNext AdmiralQuestionQuickAction = "select_next"
	// AdmiralQuestionQuickActionSubmit submits the response.
	AdmiralQuestionQuickActionSubmit AdmiralQuestionQuickAction = "submit"
	// AdmiralQuestionQuickActionDismiss dismisses the modal.
	AdmiralQuestionQuickActionDismiss AdmiralQuestionQuickAction = "dismiss"
)

// AdmiralQuestionResponse contains user-entered response fields.
type AdmiralQuestionResponse struct {
	SelectedOption string
	FreeText       string
	Broadcast      bool
}

// AdmiralQuestionModalConfig contains all render/configuration inputs for the modal.
type AdmiralQuestionModalConfig struct {
	Width          int
	Height         int
	QuestionID     string
	AgentName      string
	AgentRole      string
	Domain         string
	ShipName       string
	MissionID      string
	QuestionText   string
	Options        []string
	AllowFreeText  bool
	AllowBroadcast bool
	Response       AdmiralQuestionResponse
}

// AdmiralQuestionQuickActionForKey resolves direct key actions.
func AdmiralQuestionQuickActionForKey(msg tea.KeyMsg) AdmiralQuestionQuickAction {
	switch strings.ToLower(strings.TrimSpace(msg.String())) {
	case "up", "k":
		return AdmiralQuestionQuickActionSelectPrev
	case "down", "j":
		return AdmiralQuestionQuickActionSelectNext
	case "enter":
		return AdmiralQuestionQuickActionSubmit
	case "esc":
		return AdmiralQuestionQuickActionDismiss
	default:
		return AdmiralQuestionQuickActionNone
	}
}

// AdmiralQuestionAnimationConfig returns spring frequency and damping ratio for a phase.
func AdmiralQuestionAnimationConfig(state AdmiralQuestionAnimationState) (float64, float64) {
	if state == AdmiralQuestionAnimationClose {
		return admiralQuestionCloseFrequency, admiralQuestionCloseDampingRatio
	}
	return admiralQuestionOpenFrequency, admiralQuestionOpenDampingRatio
}

// AdmiralQuestionAnimationSpring builds the harmonica spring for a phase.
func AdmiralQuestionAnimationSpring(state AdmiralQuestionAnimationState) harmonica.Spring {
	frequency, ratio := AdmiralQuestionAnimationConfig(state)
	return harmonica.NewSpring(harmonica.FPS(60), frequency, ratio)
}

// ApplyAdmiralQuestionAnimation advances one spring step and returns updated position and velocity.
func ApplyAdmiralQuestionAnimation(position float64, velocity float64, state AdmiralQuestionAnimationState) (float64, float64) {
	spring := AdmiralQuestionAnimationSpring(state)
	target := 1.0
	if state == AdmiralQuestionAnimationClose {
		target = 0.0
	}
	return spring.Update(position, velocity, target)
}

// BuildAdmiralQuestionForm constructs the canonical huh.Form for Admiral responses.
func BuildAdmiralQuestionForm(config AdmiralQuestionModalConfig, response *AdmiralQuestionResponse) *huh.Form {
	if response == nil {
		response = &AdmiralQuestionResponse{}
	}

	options := buildAdmiralQuestionOptions(config.Options)
	if strings.TrimSpace(response.SelectedOption) == "" && len(options) > 0 {
		response.SelectedOption = options[0]
	}

	questionSummary := strings.TrimSpace(config.QuestionText)
	if questionSummary == "" {
		questionSummary = "No question text provided."
	}
	if strings.TrimSpace(config.MissionID) != "" {
		questionSummary = questionSummary + "\n\nMission: " + strings.TrimSpace(config.MissionID)
	}

	fields := []huh.Field{
		huh.NewNote().
			Title("Question").
			Description(questionSummary),
		huh.NewSelect[string]().
			Title("Select an option").
			Options(huh.NewOptions(options...)...).
			Value(&response.SelectedOption),
		huh.NewInput().
			Title("Or provide a custom response").
			Placeholder("Type a custom response...").
			Value(&response.FreeText),
		huh.NewConfirm().
			Title("Broadcast to all crew?").
			Affirmative("Yes").
			Negative("No").
			Value(&response.Broadcast),
	}

	form := huh.NewForm(huh.NewGroup(fields...)).
		WithShowHelp(false).
		WithShowErrors(false).
		WithWidth(72)
	_ = form.Init()
	return form
}

// RenderAdmiralQuestionModal renders a centered overlay modal with dimmed background fill.
func RenderAdmiralQuestionModal(config AdmiralQuestionModalConfig) string {
	width := config.Width
	if width <= 0 {
		width = admiralQuestionDefaultWidth
	}

	height := config.Height
	if height <= 0 {
		height = admiralQuestionDefaultHeight
	}

	modalWidth := int(float64(width) * admiralQuestionStandardWidthPct)
	if width < admiralQuestionCompactThreshold {
		modalWidth = int(float64(width) * admiralQuestionCompactWidthPct)
	}
	modalWidth = max(50, modalWidth)

	header := renderAdmiralQuestionHeader(config)
	form := BuildAdmiralQuestionForm(config, &config.Response)
	formView := strings.TrimSpace(form.View())
	if formView == "" {
		formView = renderAdmiralQuestionFallbackForm(config)
	}
	questionText := renderMarkdown(strings.TrimSpace(config.QuestionText), max(40, modalWidth-10))
	if strings.TrimSpace(questionText) == "" {
		questionText = "No question text provided."
	}

	questionLabel := lipgloss.NewStyle().Foreground(theme.SpaceWhiteColor).Bold(true).Render("Question")
	hint := lipgloss.NewStyle().Foreground(theme.GalaxyGrayColor).Render("Press Enter to submit")

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		questionLabel,
		questionText,
		formView,
		hint,
	)

	modal := lipgloss.NewStyle().
		Width(modalWidth).
		Padding(0, 1).
		Border(lipgloss.DoubleBorder()).
		BorderForeground(theme.GoldColor).
		Render(content)

	return lipgloss.Place(
		width,
		height,
		lipgloss.Center,
		lipgloss.Center,
		modal,
		lipgloss.WithWhitespaceChars("░"),
		lipgloss.WithWhitespaceForeground(theme.GalaxyGrayColor),
	)
}

// SubmitAdmiralQuestionAnswer validates and submits a response into AdmiralQuestionGate.
func SubmitAdmiralQuestionAnswer(gate *admiral.QuestionGate, question admiral.AdmiralQuestion, response AdmiralQuestionResponse) error {
	if gate == nil {
		return errors.New("question gate is nil")
	}

	answer := admiral.AdmiralAnswer{
		QuestionID: strings.TrimSpace(question.QuestionID),
		Broadcast:  response.Broadcast,
	}

	selected := strings.TrimSpace(response.SelectedOption)
	freeText := strings.TrimSpace(response.FreeText)
	if strings.EqualFold(selected, AdmiralQuestionSkipOption) {
		answer.SkipFlag = true
	} else {
		answer.SelectedOption = selected
		answer.FreeText = freeText
	}

	if err := admiral.ValidateAnswer(question, answer); err != nil {
		return err
	}
	if err := gate.SubmitAnswer(answer); err != nil {
		return err
	}

	return nil
}

func renderAdmiralQuestionHeader(config AdmiralQuestionModalConfig) string {
	agent := strings.TrimSpace(config.AgentName)
	if agent == "" {
		agent = "Unknown Agent"
	}
	ship := strings.TrimSpace(config.ShipName)
	if ship == "" {
		ship = "Unknown Ship"
	}
	role := strings.TrimSpace(config.AgentRole)
	if role == "" {
		role = "Unassigned"
	}
	domain := strings.TrimSpace(config.Domain)
	if domain == "" {
		domain = "general"
	}

	title := lipgloss.NewStyle().
		Foreground(theme.ButterscotchColor).
		Bold(true).
		Render(fmt.Sprintf("ADMIRAL -- QUESTION FROM %s", agent))

	shipLabel := lipgloss.NewStyle().Foreground(theme.GalaxyGrayColor).Faint(true).Render(ship)
	meta := lipgloss.JoinHorizontal(
		lipgloss.Left,
		shipLabel,
		"  ",
		renderRoleBadge(role),
		" ",
		renderDomainBadge(domain),
	)

	return lipgloss.JoinVertical(lipgloss.Left, title, meta)
}

func renderRoleBadge(role string) string {
	normalized := strings.ToLower(strings.TrimSpace(role))
	bg := theme.ButterscotchColor
	switch normalized {
	case "commander":
		bg = theme.BlueColor
	case "captain":
		bg = theme.GoldColor
	case "ensign":
		bg = theme.ButterscotchColor
	}
	return lipgloss.NewStyle().
		Foreground(theme.BlackColor).
		Background(bg).
		Bold(true).
		Padding(0, 1).
		Render(strings.ToUpper(strings.TrimSpace(role)))
}

func renderDomainBadge(domain string) string {
	normalized := strings.ToLower(strings.TrimSpace(domain))
	bg := theme.GalaxyGrayColor
	switch normalized {
	case "functional":
		bg = theme.BlueColor
	case "technical":
		bg = theme.ButterscotchColor
	case "design":
		bg = theme.PurpleColor
	}
	return lipgloss.NewStyle().
		Foreground(theme.BlackColor).
		Background(bg).
		Bold(true).
		Padding(0, 1).
		Render(strings.ToUpper(strings.TrimSpace(domain)))
}

func renderAdmiralQuestionFallbackForm(config AdmiralQuestionModalConfig) string {
	options := buildAdmiralQuestionOptions(config.Options)
	lines := make([]string, 0, len(options)+8)
	lines = append(lines, lipgloss.NewStyle().Foreground(theme.SpaceWhiteColor).Bold(true).Render("Select an option"))

	selected := strings.TrimSpace(config.Response.SelectedOption)
	if selected == "" && len(options) > 0 {
		selected = options[0]
	}
	for _, option := range options {
		prefix := "  "
		style := lipgloss.NewStyle().Foreground(theme.LightGrayColor)
		if strings.EqualFold(selected, option) {
			prefix = "> "
			style = lipgloss.NewStyle().Foreground(theme.MoonlitVioletColor).Bold(true)
		}
		lines = append(lines, style.Render(prefix+option))
	}

	lines = append(lines, "")
	lines = append(lines, lipgloss.NewStyle().Foreground(theme.GalaxyGrayColor).Render("Or provide a custom response"))
	value := strings.TrimSpace(config.Response.FreeText)
	if value == "" {
		value = "Type a custom response..."
	}
	lines = append(lines, lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(theme.GalaxyGrayColor).Render(value))

	check := "[ ]"
	if config.Response.Broadcast {
		check = "[x]"
	}
	lines = append(lines, "")
	lines = append(lines, lipgloss.NewStyle().Foreground(theme.LightGrayColor).Render(check+" Broadcast to all crew?"))

	return strings.Join(lines, "\n")
}

func buildAdmiralQuestionOptions(options []string) []string {
	dedup := make(map[string]struct{}, len(options)+1)
	ordered := make([]string, 0, len(options)+1)
	for _, option := range options {
		trimmed := strings.TrimSpace(option)
		if trimmed == "" {
			continue
		}
		key := strings.ToLower(trimmed)
		if _, exists := dedup[key]; exists {
			continue
		}
		dedup[key] = struct{}{}
		ordered = append(ordered, trimmed)
	}
	if _, exists := dedup[strings.ToLower(AdmiralQuestionSkipOption)]; !exists {
		ordered = append(ordered, AdmiralQuestionSkipOption)
	}
	return ordered
}
