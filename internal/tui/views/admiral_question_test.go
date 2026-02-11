package views

import (
	"context"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ship-commander/sc3/internal/admiral"
)

func TestRenderAdmiralQuestionModalIncludesBorderHeaderAndBadges(t *testing.T) {
	t.Parallel()

	rendered := RenderAdmiralQuestionModal(sampleAdmiralQuestionConfig())
	for _, expected := range []string{
		"ADMIRAL -- QUESTION FROM Cmdr. Data",
		"USS Enterprise",
		"COMMANDER",
		"TECHNICAL",
		"Question",
		"Select an option",
		"Or provide a custom response",
		"Broadcast to all crew?",
		AdmiralQuestionSkipOption,
		"Press Enter",
		"╔",
		"░",
	} {
		if !strings.Contains(rendered, expected) {
			t.Fatalf("rendered modal missing %q\n%s", expected, rendered)
		}
	}
}

func TestBuildAdmiralQuestionFormIncludesSelectInputConfirmAndSkipOption(t *testing.T) {
	t.Parallel()

	config := sampleAdmiralQuestionConfig()
	config.Response = AdmiralQuestionResponse{}
	response := config.Response

	form := BuildAdmiralQuestionForm(config, &response)
	if form == nil {
		t.Fatal("expected huh form instance")
	}

	if strings.TrimSpace(response.SelectedOption) == "" {
		t.Fatalf("expected selected option default to be set, got %q", response.SelectedOption)
	}

	options := buildAdmiralQuestionOptions(config.Options)
	if !containsString(options, AdmiralQuestionSkipOption) {
		t.Fatalf("expected skip option to be appended: %v", options)
	}
}

func TestAdmiralQuestionQuickActionForKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		key  tea.KeyMsg
		want AdmiralQuestionQuickAction
	}{
		{key: tea.KeyMsg{Type: tea.KeyUp}, want: AdmiralQuestionQuickActionSelectPrev},
		{key: tea.KeyMsg{Type: tea.KeyDown}, want: AdmiralQuestionQuickActionSelectNext},
		{key: tea.KeyMsg{Type: tea.KeyEnter}, want: AdmiralQuestionQuickActionSubmit},
		{key: tea.KeyMsg{Type: tea.KeyEsc}, want: AdmiralQuestionQuickActionDismiss},
		{key: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}, want: AdmiralQuestionQuickActionNone},
	}

	for _, tt := range tests {
		if got := AdmiralQuestionQuickActionForKey(tt.key); got != tt.want {
			t.Fatalf("quick action for key %q = %q, want %q", tt.key.String(), got, tt.want)
		}
	}
}

func TestAdmiralQuestionAnimationProfiles(t *testing.T) {
	t.Parallel()

	openFreq, openRatio := AdmiralQuestionAnimationConfig(AdmiralQuestionAnimationOpen)
	if openFreq != 6.0 || openRatio != 0.8 {
		t.Fatalf("open config = (%f, %f), want (6.0, 0.8)", openFreq, openRatio)
	}

	closeFreq, closeRatio := AdmiralQuestionAnimationConfig(AdmiralQuestionAnimationClose)
	if closeFreq != 10.0 || closeRatio != 1.0 {
		t.Fatalf("close config = (%f, %f), want (10.0, 1.0)", closeFreq, closeRatio)
	}

	opened, _ := ApplyAdmiralQuestionAnimation(0.0, 0.0, AdmiralQuestionAnimationOpen)
	if opened <= 0.0 {
		t.Fatalf("open spring position = %f, want > 0", opened)
	}

	closed, _ := ApplyAdmiralQuestionAnimation(1.0, 0.0, AdmiralQuestionAnimationClose)
	if closed >= 1.0 {
		t.Fatalf("close spring position = %f, want < 1", closed)
	}
}

func TestSubmitAdmiralQuestionAnswerResolvesQuestionGateOnSubmit(t *testing.T) {
	t.Parallel()

	gate := admiral.NewQuestionGate(1)
	question := admiral.AdmiralQuestion{
		QuestionID:     "Q-22",
		AskingAgent:    "Cmdr. Data",
		MissionID:      "MISSION-22",
		Domain:         "technical",
		QuestionText:   "Which session store should we use?",
		Options:        []string{"Redis", "SQLite"},
		AllowFreeText:  true,
		AllowBroadcast: true,
	}

	errCh := make(chan error, 1)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_, err := gate.Ask(ctx, question)
		errCh <- err
	}()

	select {
	case asked := <-gate.Questions():
		if asked.QuestionID != "Q-22" {
			t.Fatalf("asked question id = %q, want Q-22", asked.QuestionID)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for question gate event")
	}

	err := SubmitAdmiralQuestionAnswer(gate, question, AdmiralQuestionResponse{
		SelectedOption: "Redis",
		Broadcast:      true,
	})
	if err != nil {
		t.Fatalf("submit answer: %v", err)
	}

	select {
	case askErr := <-errCh:
		if askErr != nil {
			t.Fatalf("gate ask error: %v", askErr)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for question gate to resolve")
	}

	history := gate.History()
	if len(history) != 1 {
		t.Fatalf("history length = %d, want 1", len(history))
	}
	if history[0].Answer.SelectedOption != "Redis" {
		t.Fatalf("selected option = %q, want Redis", history[0].Answer.SelectedOption)
	}
	if !history[0].Answer.Broadcast {
		t.Fatal("expected broadcast=true")
	}
}

func sampleAdmiralQuestionConfig() AdmiralQuestionModalConfig {
	return AdmiralQuestionModalConfig{
		Width:          120,
		Height:         30,
		QuestionID:     "Q-22",
		AgentName:      "Cmdr. Data",
		AgentRole:      "Commander",
		Domain:         "technical",
		ShipName:       "USS Enterprise",
		MissionID:      "MISSION-22",
		QuestionText:   "The authentication module requires a session store. Which approach should I use?",
		Options:        []string{"Redis-backed session store", "Database-backed sessions"},
		AllowFreeText:  true,
		AllowBroadcast: true,
	}
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if strings.EqualFold(strings.TrimSpace(value), strings.TrimSpace(target)) {
			return true
		}
	}
	return false
}
