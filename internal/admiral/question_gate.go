package admiral

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

const defaultGateBuffer = 1

// AdmiralQuestion is the normalized question payload sent from a planning agent to the Admiral.
//
//nolint:revive // Field names are specified by the issue contract.
type AdmiralQuestion struct {
	QuestionID     string
	AskingAgent    string
	MissionID      string
	Domain         string
	QuestionText   string
	Options        []string
	AllowFreeText  bool
	AllowBroadcast bool
}

// AdmiralAnswer is the Admiral's response payload for a question.
//
//nolint:revive // Field names are specified by the issue contract.
type AdmiralAnswer struct {
	QuestionID     string
	SelectedOption string
	FreeText       string
	Broadcast      bool
	SkipFlag       bool
}

// QuestionRecord captures one persisted question/answer pair linked by QuestionID.
type QuestionRecord struct {
	QuestionID string
	Question   AdmiralQuestion
	Answer     AdmiralAnswer
	AskedAt    time.Time
	AnsweredAt time.Time
}

// QuestionGate is a channel-based gate that blocks planning progress until an Admiral answer arrives.
type QuestionGate struct {
	questions chan AdmiralQuestion
	answers   chan AdmiralAnswer
	now       func() time.Time

	mu      sync.Mutex
	history []QuestionRecord
}

// NewQuestionGate constructs a new blocking Admiral question gate.
func NewQuestionGate(bufferSize int) *QuestionGate {
	if bufferSize <= 0 {
		bufferSize = defaultGateBuffer
	}
	return &QuestionGate{
		questions: make(chan AdmiralQuestion, bufferSize),
		answers:   make(chan AdmiralAnswer, bufferSize),
		now:       time.Now,
		history:   make([]QuestionRecord, 0),
	}
}

// Questions exposes surfaced Admiral questions for subscribers (for example, TUI modal handling).
func (g *QuestionGate) Questions() <-chan AdmiralQuestion {
	return g.questions
}

// SubmitAnswer publishes one Admiral answer into the gate.
func (g *QuestionGate) SubmitAnswer(answer AdmiralAnswer) error {
	if g == nil {
		return errors.New("question gate is nil")
	}

	answer.QuestionID = strings.TrimSpace(answer.QuestionID)
	answer.SelectedOption = strings.TrimSpace(answer.SelectedOption)
	answer.FreeText = strings.TrimSpace(answer.FreeText)
	if answer.QuestionID == "" {
		return errors.New("question id is required")
	}

	g.answers <- answer
	return nil
}

// Ask surfaces a question and blocks until the matching answer is received or context is canceled.
func (g *QuestionGate) Ask(ctx context.Context, question AdmiralQuestion) (AdmiralAnswer, error) {
	if g == nil {
		return AdmiralAnswer{}, errors.New("question gate is nil")
	}

	normalized, err := normalizeQuestion(question)
	if err != nil {
		return AdmiralAnswer{}, err
	}
	askedAt := g.now().UTC()

	select {
	case g.questions <- normalized:
	case <-ctx.Done():
		return AdmiralAnswer{}, ctx.Err()
	}

	for {
		select {
		case answer := <-g.answers:
			answer = normalizeAnswer(answer)
			if answer.QuestionID != normalized.QuestionID {
				continue
			}
			record := QuestionRecord{
				QuestionID: normalized.QuestionID,
				Question:   normalized,
				Answer:     answer,
				AskedAt:    askedAt,
				AnsweredAt: g.now().UTC(),
			}
			g.mu.Lock()
			g.history = append(g.history, record)
			g.mu.Unlock()

			return answer, nil
		case <-ctx.Done():
			return AdmiralAnswer{}, ctx.Err()
		}
	}
}

// History returns a copy of persisted question/answer records.
func (g *QuestionGate) History() []QuestionRecord {
	if g == nil {
		return nil
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	history := make([]QuestionRecord, len(g.history))
	copy(history, g.history)
	return history
}

func normalizeQuestion(question AdmiralQuestion) (AdmiralQuestion, error) {
	question.QuestionID = strings.TrimSpace(question.QuestionID)
	question.AskingAgent = strings.TrimSpace(question.AskingAgent)
	question.MissionID = strings.TrimSpace(question.MissionID)
	question.Domain = strings.TrimSpace(question.Domain)
	question.QuestionText = strings.TrimSpace(question.QuestionText)
	if question.QuestionID == "" {
		return AdmiralQuestion{}, errors.New("question id is required")
	}
	if question.AskingAgent == "" {
		return AdmiralQuestion{}, errors.New("asking agent is required")
	}
	if question.QuestionText == "" {
		return AdmiralQuestion{}, errors.New("question text is required")
	}

	options := make([]string, 0, len(question.Options))
	for _, option := range question.Options {
		option = strings.TrimSpace(option)
		if option == "" {
			continue
		}
		options = append(options, option)
	}
	question.Options = options

	return question, nil
}

func normalizeAnswer(answer AdmiralAnswer) AdmiralAnswer {
	answer.QuestionID = strings.TrimSpace(answer.QuestionID)
	answer.SelectedOption = strings.TrimSpace(answer.SelectedOption)
	answer.FreeText = strings.TrimSpace(answer.FreeText)
	return answer
}

// ValidateAnswer checks whether an answer shape is valid for a specific question.
func ValidateAnswer(question AdmiralQuestion, answer AdmiralAnswer) error {
	question, err := normalizeQuestion(question)
	if err != nil {
		return err
	}
	answer = normalizeAnswer(answer)

	if answer.QuestionID == "" {
		return errors.New("question id is required")
	}
	if answer.QuestionID != question.QuestionID {
		return fmt.Errorf("answer question id %q does not match %q", answer.QuestionID, question.QuestionID)
	}
	if answer.SkipFlag {
		return nil
	}

	if answer.SelectedOption != "" {
		for _, option := range question.Options {
			if option == answer.SelectedOption {
				return nil
			}
		}
		return fmt.Errorf("selected option %q not found in question options", answer.SelectedOption)
	}

	if answer.FreeText != "" {
		if !question.AllowFreeText {
			return errors.New("free text is not allowed for this question")
		}
		return nil
	}

	return errors.New("answer must set selected option, free text, or skip flag")
}
