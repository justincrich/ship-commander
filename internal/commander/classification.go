package commander

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/ship-commander/sc3/internal/telemetry"
	"gopkg.in/yaml.v3"
)

const (
	// MissionClassificationREDAlert routes mission execution through full TDD gates.
	MissionClassificationREDAlert = "RED_ALERT"
)

const (
	confidenceHigh   = "high"
	confidenceMedium = "medium"
	confidenceLow    = "low"
)

var (
	// ErrLowConfidenceClassification indicates the mission needs Admiral confirmation.
	ErrLowConfidenceClassification = errors.New("low-confidence classification requires admiral review")

	redAlertCriteria = map[string]struct{}{
		"business_logic": {},
		"api_changes":    {},
		"auth_security":  {},
		"data_integrity": {},
		"bug_fix":        {},
	}
	standardOpsCriteria = map[string]struct{}{
		"styling":                 {},
		"non_behavioral_refactor": {},
		"tooling":                 {},
		"documentation":           {},
	}
)

// ClassificationContext is the prompt context needed for mission classification.
type ClassificationContext struct {
	MissionID              string
	Title                  string
	UseCase                string
	FunctionalRequirements string
	DesignRequirements     string
	CommissionTitle        string
	Domain                 string
	Dependencies           []string
	Harness                string
	Model                  string
}

// ClassificationRationale describes why a mission was classified a specific way.
type ClassificationRationale struct {
	AffectsBehavior bool
	CriteriaMatched []string
	RiskAssessment  string
	Confidence      string
}

// ClassificationResult is the normalized classification output consumed by execution and review UIs.
type ClassificationResult struct {
	MissionID       string
	Title           string
	Classification  string
	Rationale       ClassificationRationale
	Harness         string
	Model           string
	RawYAMLResponse string
}

// RequiresAdmiralReview reports whether the current result requires explicit Admiral confirmation.
func (r ClassificationResult) RequiresAdmiralReview() bool {
	return normalizeConfidence(r.Rationale.Confidence) == confidenceLow
}

// ClassificationRequest is the harness invocation payload for risk classification.
type ClassificationRequest struct {
	Harness string
	Model   string
	Prompt  string
}

// ClassificationInvoker runs one mission-classification prompt through a configured harness/model.
type ClassificationInvoker interface {
	ClassifyMission(ctx context.Context, request ClassificationRequest) (string, error)
}

// Classifier classifies missions as RED_ALERT or STANDARD_OPS using a configured harness/model.
type Classifier struct {
	invoker ClassificationInvoker
}

// NewClassifier builds a mission classifier with the provided harness invoker.
func NewClassifier(invoker ClassificationInvoker) (*Classifier, error) {
	if invoker == nil {
		return nil, errors.New("classification invoker is required")
	}
	return &Classifier{invoker: invoker}, nil
}

// LowConfidenceClassificationError captures the parsed result when classification requires Admiral review.
type LowConfidenceClassificationError struct {
	Result ClassificationResult
}

func (e *LowConfidenceClassificationError) Error() string {
	return fmt.Sprintf(
		"classification for mission %q requires admiral review (confidence=%s)",
		e.Result.MissionID,
		normalizeConfidence(e.Result.Rationale.Confidence),
	)
}

// Is allows errors.Is(err, ErrLowConfidenceClassification) checks.
func (e *LowConfidenceClassificationError) Is(target error) bool {
	return target == ErrLowConfidenceClassification
}

// ClassifyMission renders the prompt, invokes the configured harness/model, and parses YAML output.
func (c *Classifier) ClassifyMission(ctx context.Context, input ClassificationContext) (ClassificationResult, error) {
	if c == nil {
		return ClassificationResult{}, errors.New("classifier is nil")
	}

	input.MissionID = strings.TrimSpace(input.MissionID)
	if input.MissionID == "" {
		return ClassificationResult{}, errors.New("mission id is required")
	}
	input.Harness = strings.TrimSpace(input.Harness)
	if input.Harness == "" {
		return ClassificationResult{}, errors.New("classification harness must be configured")
	}
	input.Model = strings.TrimSpace(input.Model)
	if input.Model == "" {
		return ClassificationResult{}, errors.New("classification model must be configured")
	}

	prompt, err := BuildClassificationPrompt(input)
	if err != nil {
		return ClassificationResult{}, err
	}

	classifyCtx, llmCall := telemetry.StartLLMCall(ctx, telemetry.LLMCallRequest{
		Operation: "mission_classification",
		ModelName: input.Model,
		Harness:   input.Harness,
		Prompt:    prompt,
	})

	rawResponse, err := c.invoker.ClassifyMission(classifyCtx, ClassificationRequest{
		Harness: input.Harness,
		Model:   input.Model,
		Prompt:  prompt,
	})
	if err != nil {
		llmCall.RecordError("classification_invoke_error", err.Error(), 0)
		llmCall.End("", nil, err)
		return ClassificationResult{}, fmt.Errorf("invoke classification harness: %w", err)
	}

	result, err := parseClassificationYAML(input, rawResponse)
	if err != nil {
		llmCall.RecordError("classification_parse_error", err.Error(), 0)
		llmCall.End(rawResponse, nil, err)
		return ClassificationResult{}, err
	}
	result.Harness = input.Harness
	result.Model = input.Model
	result.RawYAMLResponse = rawResponse
	llmCall.End(rawResponse, nil, nil)

	if result.RequiresAdmiralReview() {
		return result, &LowConfidenceClassificationError{Result: result}
	}
	return result, nil
}

type classificationYAML struct {
	MissionID      string `yaml:"mission_id"`
	Title          string `yaml:"title"`
	Classification string `yaml:"classification"`
	Confidence     string `yaml:"confidence"`
	Rationale      struct {
		AffectsBehavior bool     `yaml:"affects_behavior"`
		CriteriaMatched []string `yaml:"criteria_matched"`
		RiskAssessment  string   `yaml:"risk_assessment"`
		Confidence      string   `yaml:"confidence"`
	} `yaml:"rationale"`
}

func parseClassificationYAML(input ClassificationContext, rawResponse string) (ClassificationResult, error) {
	trimmed := strings.TrimSpace(rawResponse)
	if trimmed == "" {
		return ClassificationResult{}, errors.New("classification response is empty")
	}

	var parsed classificationYAML
	if err := yaml.Unmarshal([]byte(trimmed), &parsed); err != nil {
		return ClassificationResult{}, fmt.Errorf("parse classification YAML: %w", err)
	}

	result := ClassificationResult{
		MissionID:      firstNonEmpty(strings.TrimSpace(parsed.MissionID), strings.TrimSpace(input.MissionID)),
		Title:          firstNonEmpty(strings.TrimSpace(parsed.Title), strings.TrimSpace(input.Title), strings.TrimSpace(input.MissionID)),
		Classification: normalizeClassification(parsed.Classification),
		Rationale: ClassificationRationale{
			AffectsBehavior: parsed.Rationale.AffectsBehavior,
			CriteriaMatched: normalizeCriteria(parsed.Rationale.CriteriaMatched),
			RiskAssessment:  strings.TrimSpace(parsed.Rationale.RiskAssessment),
			Confidence:      normalizeConfidence(firstNonEmpty(parsed.Rationale.Confidence, parsed.Confidence)),
		},
	}

	if result.Classification == "" {
		result.Classification = MissionClassificationREDAlert
	}
	if result.Rationale.Confidence == "" {
		result.Rationale.Confidence = confidenceLow
	}
	if result.Rationale.RiskAssessment == "" {
		result.Rationale.RiskAssessment = "No rationale supplied by classifier."
	}

	if err := validateClassificationResult(result); err != nil {
		return ClassificationResult{}, err
	}
	return result, nil
}

func validateClassificationResult(result ClassificationResult) error {
	if strings.TrimSpace(result.MissionID) == "" {
		return errors.New("classification response missing mission_id")
	}

	switch result.Classification {
	case MissionClassificationREDAlert:
	case MissionClassificationStandardOps:
	default:
		return fmt.Errorf("unsupported mission classification %q", result.Classification)
	}

	if len(result.Rationale.CriteriaMatched) == 0 {
		return errors.New("classification response must include at least one criteria_matched entry")
	}

	for _, criterion := range result.Rationale.CriteriaMatched {
		if _, ok := redAlertCriteria[criterion]; ok {
			if result.Classification == MissionClassificationStandardOps {
				return fmt.Errorf(
					"criterion %q requires %s classification",
					criterion,
					MissionClassificationREDAlert,
				)
			}
			continue
		}
		if _, ok := standardOpsCriteria[criterion]; ok {
			continue
		}
		return fmt.Errorf("unsupported classification criterion %q", criterion)
	}

	if result.Classification == MissionClassificationStandardOps {
		for _, criterion := range result.Rationale.CriteriaMatched {
			if _, ok := redAlertCriteria[criterion]; ok {
				return fmt.Errorf(
					"STANDARD_OPS classification is invalid with RED_ALERT criterion %q",
					criterion,
				)
			}
		}
	}

	return nil
}

func normalizeCriteria(criteria []string) []string {
	out := make([]string, 0, len(criteria))
	for _, criterion := range criteria {
		normalized := strings.ToLower(strings.TrimSpace(criterion))
		if normalized == "" {
			continue
		}
		if !containsString(out, normalized) {
			out = append(out, normalized)
		}
	}
	return out
}

func normalizeClassification(value string) string {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case MissionClassificationREDAlert:
		return MissionClassificationREDAlert
	case MissionClassificationStandardOps:
		return MissionClassificationStandardOps
	default:
		return ""
	}
}

func normalizeConfidence(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case confidenceHigh:
		return confidenceHigh
	case confidenceMedium:
		return confidenceMedium
	case confidenceLow:
		return confidenceLow
	default:
		return ""
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

func containsString(values []string, candidate string) bool {
	for _, value := range values {
		if value == candidate {
			return true
		}
	}
	return false
}
