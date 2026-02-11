package admiral

import (
	"bytes"
	"context"
	"testing"
	"time"
)

func TestApprovalConsumerApproveAndFeedbackFlow(t *testing.T) {
	t.Parallel()

	input := bytes.NewBufferString("a\n")
	output := &bytes.Buffer{}
	gate := NewApprovalGate(1)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	StartApprovalGateStdioConsumer(ctx, gate, input, output)

	resp, err := gate.AwaitDecision(context.Background(), ApprovalRequest{
		CommissionID:    "C-1",
		MissionManifest: []Mission{{ID: "M-1", Title: "One", Classification: "RED_ALERT"}},
		Iteration:       1,
		MaxIterations:   1,
	})
	if err != nil {
		t.Fatalf("await decision: %v", err)
	}
	if resp.Decision != ApprovalDecisionApproved {
		t.Fatalf("decision = %q, want %q", resp.Decision, ApprovalDecisionApproved)
	}

	input.Reset()
	input.WriteString("f\nline1\nline2\n\n")
	resp, err = gate.AwaitDecision(context.Background(), ApprovalRequest{
		CommissionID:    "C-2",
		MissionManifest: []Mission{{ID: "M-2", Title: "Two", Classification: "STANDARD_OPS"}},
		Iteration:       1,
		MaxIterations:   2,
	})
	if err != nil {
		t.Fatalf("await feedback decision: %v", err)
	}
	if resp.Decision != ApprovalDecisionFeedback {
		t.Fatalf("decision = %q, want %q", resp.Decision, ApprovalDecisionFeedback)
	}
	if resp.FeedbackText != "line1\nline2" {
		t.Fatalf("feedback = %q, want %q", resp.FeedbackText, "line1\\nline2")
	}
}

func TestWaveReviewPromptUsesContinueFeedbackHalt(t *testing.T) {
	t.Parallel()

	gate := NewApprovalGate(1)
	output := &bytes.Buffer{}
	input := bytes.NewBufferString("h\nstop now\n\n")
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	StartApprovalGateStdioConsumer(ctx, gate, input, output)

	resp, err := gate.AwaitDecision(context.Background(), ApprovalRequest{
		CommissionID:    "C-WAVE",
		MissionManifest: []Mission{{ID: "M-1", Title: "One", Classification: "RED_ALERT"}},
		Iteration:       1,
		MaxIterations:   1,
		WaveReview: &WaveReview{
			WaveIndex:  1,
			DemoTokens: map[string]string{"M-1": "demo"},
		},
	})
	if err != nil {
		t.Fatalf("await wave review: %v", err)
	}
	if resp.Decision != ApprovalDecisionHalted {
		t.Fatalf("decision = %q, want %q", resp.Decision, ApprovalDecisionHalted)
	}
	if resp.FeedbackText != "stop now" {
		t.Fatalf("feedback = %q, want stop now", resp.FeedbackText)
	}
}

func TestQuestionConsumerSupportsOptionAndFreeText(t *testing.T) {
	t.Parallel()

	gate := NewQuestionGate(1)
	output := &bytes.Buffer{}
	input := bytes.NewBufferString("2\n")
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	StartQuestionGateStdioConsumer(ctx, gate, input, output)

	answer, err := gate.Ask(context.Background(), AdmiralQuestion{
		QuestionID:    "Q-1",
		AskingAgent:   "captain",
		Domain:        "backend",
		QuestionText:  "Pick strategy",
		Options:       []string{"A", "B"},
		AllowFreeText: true,
	})
	if err != nil {
		t.Fatalf("ask option: %v", err)
	}
	if answer.SelectedOption != "B" {
		t.Fatalf("selected = %q, want B", answer.SelectedOption)
	}

	input.Reset()
	input.WriteString("custom answer\ny\n")
	answer, err = gate.Ask(context.Background(), AdmiralQuestion{
		QuestionID:     "Q-2",
		AskingAgent:    "commander",
		Domain:         "ui",
		QuestionText:   "Any notes?",
		Options:        []string{"Yes", "No"},
		AllowFreeText:  true,
		AllowBroadcast: true,
	})
	if err != nil {
		t.Fatalf("ask free text: %v", err)
	}
	if answer.FreeText != "custom answer" {
		t.Fatalf("free text = %q, want custom answer", answer.FreeText)
	}
	if !answer.Broadcast {
		t.Fatal("expected broadcast true")
	}
}

func TestCombinedConsumersTerminateOnCancel(t *testing.T) {
	t.Parallel()

	approval := NewApprovalGate(1)
	questions := NewQuestionGate(1)
	input := bytes.NewBufferString("")
	output := &bytes.Buffer{}
	ctx, cancel := context.WithCancel(context.Background())
	done := StartAdmiralStdioConsumers(ctx, approval, questions, input, output)
	cancel()

	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("expected consumers to stop after cancel")
	}
}
