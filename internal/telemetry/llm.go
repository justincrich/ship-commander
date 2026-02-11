package telemetry

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strings"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const maxErrorMessageBytes = 512

var (
	sensitiveInlinePattern = regexp.MustCompile(`(?i)(api[_-]?key|token|password|secret|authorization)\s*[:=]\s*([^\s,;]+)`)
	bearerTokenPattern     = regexp.MustCompile(`(?i)\bbearer\s+[a-z0-9._\-]+`)
	openAITokenPattern     = regexp.MustCompile(`\bsk-[A-Za-z0-9]{10,}\b`)
)

// LLMCallRequest defines telemetry metadata for one LLM/harness interaction.
type LLMCallRequest struct {
	Operation    string
	ModelName    string
	Harness      string
	Prompt       string
	PromptTokens int
}

// LLMCall tracks one llm.call span lifecycle.
type LLMCall struct {
	span         trace.Span
	startedAt    time.Time
	promptTokens int

	mu        sync.Mutex
	toolCalls int
	ended     bool
}

type llmCallContextKey struct{}

// StartLLMCall starts an llm.call span and returns a context carrying the tracker.
func StartLLMCall(ctx context.Context, req LLMCallRequest) (context.Context, *LLMCall) {
	if ctx == nil {
		ctx = context.Background()
	}

	model := normalizeOrUnknown(req.ModelName)
	harness := normalizeOrUnknown(req.Harness)
	promptTokens := req.PromptTokens
	if promptTokens < 0 {
		promptTokens = 0
	}
	if promptTokens == 0 {
		promptTokens = EstimateTokenCount(req.Prompt)
	}

	attrs := []attribute.KeyValue{
		attribute.String("model_name", model),
		attribute.String("harness", harness),
		attribute.Int("prompt_tokens", promptTokens),
		attribute.String("prompt_hash", hashPrompt(req.Prompt)),
	}
	if operation := strings.TrimSpace(req.Operation); operation != "" {
		attrs = append(attrs, attribute.String("operation", operation))
	}

	spanCtx, span := otel.Tracer("sc3/telemetry/llm").Start(
		ctx,
		"llm.call",
		trace.WithAttributes(attrs...),
	)

	call := &LLMCall{
		span:         span,
		startedAt:    time.Now(),
		promptTokens: promptTokens,
	}

	return context.WithValue(spanCtx, llmCallContextKey{}, call), call
}

// LLMCallFromContext returns the llm call tracker if one exists on the context.
func LLMCallFromContext(ctx context.Context) *LLMCall {
	if ctx == nil {
		return nil
	}
	callValue := ctx.Value(llmCallContextKey{})
	call, ok := callValue.(*LLMCall)
	if !ok {
		return nil
	}
	return call
}

// RecordToolCall adds a tool-call event to the active llm span.
func (c *LLMCall) RecordToolCall(toolName string, duration time.Duration, success bool) {
	if c == nil || c.span == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.ended {
		return
	}
	c.toolCalls++

	durationMS := duration.Milliseconds()
	if durationMS < 0 {
		durationMS = 0
	}

	c.span.AddEvent(
		"llm.tool_call",
		trace.WithAttributes(
			attribute.String("tool_name", normalizeOrUnknown(toolName)),
			attribute.Int64("duration_ms", durationMS),
			attribute.Bool("success", success),
		),
	)
}

// RecordError adds a redacted llm error event to the active llm span.
func (c *LLMCall) RecordError(errorType string, errorMessage string, retryCount int) {
	if c == nil || c.span == nil {
		return
	}
	if retryCount < 0 {
		retryCount = 0
	}

	c.span.AddEvent(
		"llm.error",
		trace.WithAttributes(
			attribute.String("error_type", normalizeOrUnknown(errorType)),
			attribute.String("error_message", redactSecrets(errorMessage)),
			attribute.Int("retry_count", retryCount),
		),
	)
	c.span.SetStatus(codes.Error, normalizeOrUnknown(errorType))
}

// End finalizes the llm.call span with latency, token counts, and tool call count.
func (c *LLMCall) End(responseText string, responseTokens *int, err error) {
	if c == nil || c.span == nil {
		return
	}

	c.mu.Lock()
	if c.ended {
		c.mu.Unlock()
		return
	}
	c.ended = true
	toolCalls := c.toolCalls
	promptTokens := c.promptTokens
	c.mu.Unlock()

	durationMS := time.Since(c.startedAt).Milliseconds()
	if durationMS < 0 {
		durationMS = 0
	}

	resolvedResponseTokens, includeResponseTokens := resolveResponseTokens(responseText, responseTokens)
	totalTokens := promptTokens + resolvedResponseTokens

	attrs := []attribute.KeyValue{
		attribute.Int64("latency_ms", durationMS),
		attribute.Int("tool_calls_count", toolCalls),
		attribute.Int("total_tokens", totalTokens),
	}
	if includeResponseTokens {
		attrs = append(attrs, attribute.Int("response_tokens", resolvedResponseTokens))
	}
	c.span.SetAttributes(attrs...)

	if err != nil {
		c.span.RecordError(err)
		c.span.SetStatus(codes.Error, redactSecrets(err.Error()))
	} else {
		c.span.SetStatus(codes.Ok, "llm call completed")
	}
	c.span.End()
}

// EstimateTokenCount estimates token count using a deterministic words-to-tokens heuristic.
func EstimateTokenCount(text string) int {
	fields := strings.Fields(strings.TrimSpace(text))
	if len(fields) == 0 {
		return 0
	}
	estimated := (len(fields)*4 + 2) / 3
	if estimated < 1 {
		return 1
	}
	return estimated
}

func resolveResponseTokens(responseText string, responseTokens *int) (int, bool) {
	if responseTokens != nil {
		if *responseTokens < 0 {
			return 0, false
		}
		return *responseTokens, true
	}

	estimated := EstimateTokenCount(responseText)
	if estimated <= 0 {
		return 0, false
	}
	return estimated, true
}

func hashPrompt(prompt string) string {
	sum := sha256.Sum256([]byte(redactSecrets(prompt)))
	return hex.EncodeToString(sum[:])
}

func redactSecrets(input string) string {
	redacted := strings.TrimSpace(input)
	if redacted == "" {
		return ""
	}
	redacted = sensitiveInlinePattern.ReplaceAllString(redacted, "$1=<redacted>")
	redacted = bearerTokenPattern.ReplaceAllString(redacted, "bearer <redacted>")
	redacted = openAITokenPattern.ReplaceAllString(redacted, "<redacted>")
	if len(redacted) > maxErrorMessageBytes {
		return redacted[:maxErrorMessageBytes-len("...[truncated]")] + "...[truncated]"
	}
	return redacted
}

func normalizeOrUnknown(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "unknown"
	}
	return trimmed
}
