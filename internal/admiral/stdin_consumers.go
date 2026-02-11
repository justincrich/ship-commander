package admiral

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
)

// StartApprovalGateStdioConsumer starts a goroutine that consumes approval requests and answers via stdio.
func StartApprovalGateStdioConsumer(
	ctx context.Context,
	gate *ApprovalGate,
	input io.Reader,
	output io.Writer,
) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		defer close(done)
		if gate == nil || input == nil || output == nil {
			return
		}
		reader := bufio.NewReader(input)
		for {
			select {
			case <-ctx.Done():
				return
			case request := <-gate.Requests():
				response := readApprovalResponse(reader, output, request)
				if err := gate.Respond(response); err != nil {
					writef(output, "invalid approval response: %v\n", err)
				}
			}
		}
	}()
	return done
}

// StartQuestionGateStdioConsumer starts a goroutine that consumes admiral questions and answers via stdio.
func StartQuestionGateStdioConsumer(
	ctx context.Context,
	gate *QuestionGate,
	input io.Reader,
	output io.Writer,
) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		defer close(done)
		if gate == nil || input == nil || output == nil {
			return
		}
		reader := bufio.NewReader(input)
		for {
			select {
			case <-ctx.Done():
				return
			case question := <-gate.Questions():
				answer := readQuestionAnswer(reader, output, question)
				if err := gate.SubmitAnswer(answer); err != nil {
					writef(output, "invalid answer: %v\n", err)
				}
			}
		}
	}()
	return done
}

// StartAdmiralStdioConsumers launches both stdin consumers and returns one done channel that closes after both exit.
func StartAdmiralStdioConsumers(
	ctx context.Context,
	approvalGate *ApprovalGate,
	questionGate *QuestionGate,
	input io.Reader,
	output io.Writer,
) <-chan struct{} {
	done := make(chan struct{})
	if input == nil || output == nil {
		close(done)
		return done
	}
	locked := &lockedReader{reader: bufio.NewReader(input)}
	approvalDone := StartApprovalGateStdioConsumer(ctx, approvalGate, locked, output)
	questionDone := StartQuestionGateStdioConsumer(ctx, questionGate, locked, output)
	go func() {
		defer close(done)
		<-approvalDone
		<-questionDone
	}()
	return done
}

type lockedReader struct {
	mu     sync.Mutex
	reader *bufio.Reader
}

func (r *lockedReader) Read(p []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.reader == nil {
		return 0, io.EOF
	}
	return r.reader.Read(p)
}

func readApprovalResponse(reader *bufio.Reader, output io.Writer, request ApprovalRequest) ApprovalResponse {
	if request.WaveReview != nil {
		renderWaveReviewPrompt(output, request)
		choice := strings.ToLower(strings.TrimSpace(readLine(reader)))
		switch choice {
		case "h":
			return ApprovalResponse{Decision: ApprovalDecisionHalted, FeedbackText: strings.TrimSpace(readMultiline(reader, output, "halt reason (optional):"))}
		case "f":
			feedback := strings.TrimSpace(readMultiline(reader, output, "feedback (blank line to finish):"))
			return ApprovalResponse{Decision: ApprovalDecisionFeedback, FeedbackText: feedback}
		default:
			return ApprovalResponse{Decision: ApprovalDecisionApproved}
		}
	}

	renderApprovalPrompt(output, request)
	choice := strings.ToLower(strings.TrimSpace(readLine(reader)))
	switch choice {
	case "s":
		feedback := strings.TrimSpace(readMultiline(reader, output, "shelve note (optional):"))
		return ApprovalResponse{Decision: ApprovalDecisionShelved, FeedbackText: feedback}
	case "f":
		feedback := strings.TrimSpace(readMultiline(reader, output, "feedback (blank line to finish):"))
		return ApprovalResponse{Decision: ApprovalDecisionFeedback, FeedbackText: feedback}
	default:
		return ApprovalResponse{Decision: ApprovalDecisionApproved}
	}
}

func readQuestionAnswer(reader *bufio.Reader, output io.Writer, question AdmiralQuestion) AdmiralAnswer {
	renderQuestionPrompt(output, question)
	line := strings.TrimSpace(readLine(reader))
	answer := AdmiralAnswer{QuestionID: question.QuestionID}
	if line == "" {
		answer.SkipFlag = true
		return answer
	}

	if index, err := strconv.Atoi(line); err == nil {
		if index >= 1 && index <= len(question.Options) {
			answer.SelectedOption = question.Options[index-1]
			return answer
		}
	}

	for _, option := range question.Options {
		if strings.EqualFold(option, line) {
			answer.SelectedOption = option
			return answer
		}
	}

	answer.FreeText = line
	if question.AllowBroadcast {
		write(output, "broadcast to all agents? [y/N]: ")
		broadcastChoice := strings.ToLower(strings.TrimSpace(readLine(reader)))
		answer.Broadcast = broadcastChoice == "y" || broadcastChoice == "yes"
	}
	return answer
}

func renderApprovalPrompt(output io.Writer, request ApprovalRequest) {
	writef(output, "Commission %s approval request (iteration %d/%d)\n", request.CommissionID, request.Iteration, request.MaxIterations)
	writeln(output, "Mission manifest:")
	for _, mission := range request.MissionManifest {
		writef(output, "- %s: %s (%s)\n", mission.ID, mission.Title, mission.Classification)
	}
	writeln(output, "Choose: [a]pprove, [f]eedback, [s]helve")
	write(output, "> ")
}

func renderWaveReviewPrompt(output io.Writer, request ApprovalRequest) {
	writef(output, "Wave review for commission %s wave %d\n", request.CommissionID, request.WaveReview.WaveIndex)
	writeln(output, "Demo tokens:")
	for missionID := range request.WaveReview.DemoTokens {
		writef(output, "- %s\n", missionID)
	}
	writeln(output, "Choose: [c]ontinue, [f]eedback, [h]alt")
	write(output, "> ")
}

func renderQuestionPrompt(output io.Writer, question AdmiralQuestion) {
	writef(output, "Question (%s) from %s [domain=%s]\n", question.QuestionID, question.AskingAgent, question.Domain)
	writeln(output, question.QuestionText)
	for idx, option := range question.Options {
		writef(output, "%d) %s\n", idx+1, option)
	}
	if question.AllowFreeText {
		writeln(output, "Enter option number/name, free text, or blank to skip")
	} else {
		writeln(output, "Enter option number/name or blank to skip")
	}
	write(output, "> ")
}

func readLine(reader *bufio.Reader) string {
	line, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return ""
	}
	return strings.TrimRight(line, "\r\n")
}

func readMultiline(reader *bufio.Reader, output io.Writer, prompt string) string {
	if strings.TrimSpace(prompt) != "" {
		writef(output, "%s\n", prompt)
	}
	lines := make([]string, 0)
	for {
		line := readLine(reader)
		if strings.TrimSpace(line) == "" {
			break
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func write(output io.Writer, text string) {
	if output == nil {
		return
	}
	if _, err := io.WriteString(output, text); err != nil {
		return
	}
}

func writeln(output io.Writer, text string) {
	write(output, text+"\n")
}

func writef(output io.Writer, format string, values ...any) {
	if output == nil {
		return
	}
	if _, err := fmt.Fprintf(output, format, values...); err != nil {
		return
	}
}
