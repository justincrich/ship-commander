package admiral

import (
	"context"
	"testing"
	"time"
)

func TestQuestionGateAskBlocksUntilMatchingAnswerAndPersistsHistory(t *testing.T) {
	t.Parallel()

	gate := NewQuestionGate(1)
	question := AdmiralQuestion{
		QuestionID:    "Q-1",
		AskingAgent:   "captain",
		Domain:        "functional",
		QuestionText:  "Proceed with proposed scope?",
		Options:       []string{"Proceed", "Hold"},
		AllowFreeText: true,
	}

	done := make(chan struct{})
	var answer AdmiralAnswer
	var err error
	go func() {
		defer close(done)
		answer, err = gate.Ask(context.Background(), question)
	}()

	select {
	case asked := <-gate.Questions():
		if asked.QuestionID != "Q-1" {
			t.Fatalf("question id = %q, want Q-1", asked.QuestionID)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for surfaced question")
	}

	if err := gate.SubmitAnswer(AdmiralAnswer{
		QuestionID:     "Q-1",
		SelectedOption: "Proceed",
	}); err != nil {
		t.Fatalf("submit answer: %v", err)
	}

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for ask to complete")
	}

	if err != nil {
		t.Fatalf("ask: %v", err)
	}
	if answer.SelectedOption != "Proceed" {
		t.Fatalf("selected option = %q, want Proceed", answer.SelectedOption)
	}

	history := gate.History()
	if len(history) != 1 {
		t.Fatalf("history entries = %d, want 1", len(history))
	}
	if history[0].QuestionID != "Q-1" {
		t.Fatalf("history question id = %q, want Q-1", history[0].QuestionID)
	}
}

func TestValidateAnswerSupportsOptionFreeTextAndSkip(t *testing.T) {
	t.Parallel()

	question := AdmiralQuestion{
		QuestionID:     "Q-2",
		AskingAgent:    "commander",
		Domain:         "technical",
		QuestionText:   "Choose execution mode",
		Options:        []string{"strict", "fast"},
		AllowFreeText:  true,
		AllowBroadcast: true,
	}

	tests := []struct {
		name   string
		answer AdmiralAnswer
	}{
		{
			name: "selected option",
			answer: AdmiralAnswer{
				QuestionID:     "Q-2",
				SelectedOption: "strict",
			},
		},
		{
			name: "free text",
			answer: AdmiralAnswer{
				QuestionID: "Q-2",
				FreeText:   "use hybrid mode",
			},
		},
		{
			name: "skip",
			answer: AdmiralAnswer{
				QuestionID: "Q-2",
				SkipFlag:   true,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := ValidateAnswer(question, tt.answer); err != nil {
				t.Fatalf("validate answer: %v", err)
			}
		})
	}
}
