package commander

import (
	"bytes"
	"embed"
	"fmt"
	"strings"
	"text/template"
)

//go:embed prompts/*.tmpl
var promptTemplatesFS embed.FS

var promptTemplates = template.Must(template.ParseFS(promptTemplatesFS, "prompts/*.tmpl"))

// PlanningPromptContext contains planner prompt inputs.
type PlanningPromptContext struct {
	CommissionTitle string
	CommissionPRD   string
	UseCases        []string
}

// ImplementerPromptContext contains mission context for implementer-phase prompts.
type ImplementerPromptContext struct {
	MissionID           string
	Title               string
	Classification      string
	UseCases            []string
	WorktreePath        string
	AcceptanceCriterion string
	MissionSpec         string
	PriorContext        string
	GateFeedback        string
	ValidationCommands  []string
}

// ReviewerPromptContext contains reviewer prompt inputs.
type ReviewerPromptContext struct {
	MissionID          string
	Title              string
	Classification     string
	AcceptanceCriteria []string
	GateEvidence       []string
	CodeDiff           string
	DemoTokenContent   string
}

// BuildClassificationPrompt renders the commander mission-risk prompt with mission context.
func BuildClassificationPrompt(input ClassificationContext) (string, error) {
	renderInput := struct {
		MissionID              string
		Title                  string
		UseCase                string
		CommissionTitle        string
		Domain                 string
		DependenciesText       string
		FunctionalRequirements string
		DesignRequirements     string
	}{
		MissionID:              strings.TrimSpace(input.MissionID),
		Title:                  strings.TrimSpace(input.Title),
		UseCase:                strings.TrimSpace(input.UseCase),
		CommissionTitle:        strings.TrimSpace(input.CommissionTitle),
		Domain:                 strings.TrimSpace(input.Domain),
		DependenciesText:       joinDependencies(input.Dependencies),
		FunctionalRequirements: strings.TrimSpace(input.FunctionalRequirements),
		DesignRequirements:     strings.TrimSpace(input.DesignRequirements),
	}

	if renderInput.MissionID == "" {
		return "", fmt.Errorf("mission id is required for classification prompt")
	}
	if renderInput.Title == "" {
		renderInput.Title = renderInput.MissionID
	}
	if renderInput.FunctionalRequirements == "" {
		renderInput.FunctionalRequirements = "(none provided)"
	}
	if renderInput.DesignRequirements == "" {
		renderInput.DesignRequirements = "(none provided)"
	}
	if renderInput.DependenciesText == "" {
		renderInput.DependenciesText = "(none)"
	}

	return renderTemplate("classification.tmpl", renderInput)
}

// BuildPlanningPrompt renders the planning-agent prompt.
func BuildPlanningPrompt(input PlanningPromptContext) (string, error) {
	renderInput := struct {
		CommissionTitle string
		CommissionPRD   string
		UseCasesText    string
	}{
		CommissionTitle: strings.TrimSpace(input.CommissionTitle),
		CommissionPRD:   strings.TrimSpace(input.CommissionPRD),
		UseCasesText:    joinLines(input.UseCases),
	}
	if renderInput.CommissionTitle == "" {
		return "", fmt.Errorf("commission title is required for planning prompt")
	}
	if renderInput.CommissionPRD == "" {
		renderInput.CommissionPRD = "(none provided)"
	}
	if renderInput.UseCasesText == "" {
		renderInput.UseCasesText = "(none provided)"
	}
	return renderTemplate("planning.tmpl", renderInput)
}

// BuildREDPrompt renders the RED-phase implementer prompt.
func BuildREDPrompt(input ImplementerPromptContext) (string, error) {
	return buildImplementerPrompt("red.tmpl", input)
}

// BuildGREENPrompt renders the GREEN-phase implementer prompt.
func BuildGREENPrompt(input ImplementerPromptContext) (string, error) {
	return buildImplementerPrompt("green.tmpl", input)
}

// BuildREFACTORPrompt renders the REFACTOR-phase implementer prompt.
func BuildREFACTORPrompt(input ImplementerPromptContext) (string, error) {
	return buildImplementerPrompt("refactor.tmpl", input)
}

// BuildStandardOpsPrompt renders the STANDARD_OPS implementer prompt.
func BuildStandardOpsPrompt(input ImplementerPromptContext) (string, error) {
	return buildImplementerPrompt("standard_ops.tmpl", input)
}

// BuildReviewerPrompt renders the independent reviewer prompt.
func BuildReviewerPrompt(input ReviewerPromptContext) (string, error) {
	renderInput := struct {
		MissionID              string
		Title                  string
		Classification         string
		AcceptanceCriteriaText string
		GateEvidenceText       string
		CodeDiff               string
		DemoTokenContent       string
	}{
		MissionID:              strings.TrimSpace(input.MissionID),
		Title:                  strings.TrimSpace(input.Title),
		Classification:         strings.TrimSpace(input.Classification),
		AcceptanceCriteriaText: joinLines(input.AcceptanceCriteria),
		GateEvidenceText:       joinLines(input.GateEvidence),
		CodeDiff:               strings.TrimSpace(input.CodeDiff),
		DemoTokenContent:       strings.TrimSpace(input.DemoTokenContent),
	}
	if renderInput.MissionID == "" {
		return "", fmt.Errorf("mission id is required for reviewer prompt")
	}
	if renderInput.Title == "" {
		renderInput.Title = renderInput.MissionID
	}
	if renderInput.Classification == "" {
		renderInput.Classification = MissionClassificationREDAlert
	}
	if renderInput.AcceptanceCriteriaText == "" {
		renderInput.AcceptanceCriteriaText = "(none provided)"
	}
	if renderInput.GateEvidenceText == "" {
		renderInput.GateEvidenceText = "(none provided)"
	}
	if renderInput.CodeDiff == "" {
		renderInput.CodeDiff = "(none provided)"
	}
	if renderInput.DemoTokenContent == "" {
		renderInput.DemoTokenContent = "(none provided)"
	}
	return renderTemplate("reviewer.tmpl", renderInput)
}

func buildImplementerPrompt(templateName string, input ImplementerPromptContext) (string, error) {
	renderInput := struct {
		MissionID              string
		Title                  string
		Classification         string
		UseCasesText           string
		WorktreePath           string
		AcceptanceCriterion    string
		MissionSpec            string
		PriorContext           string
		GateFeedback           string
		ValidationCommandsText string
		DemoTokenInstruction   string
	}{
		MissionID:              strings.TrimSpace(input.MissionID),
		Title:                  strings.TrimSpace(input.Title),
		Classification:         strings.TrimSpace(input.Classification),
		UseCasesText:           joinLines(input.UseCases),
		WorktreePath:           strings.TrimSpace(input.WorktreePath),
		AcceptanceCriterion:    strings.TrimSpace(input.AcceptanceCriterion),
		MissionSpec:            strings.TrimSpace(input.MissionSpec),
		PriorContext:           strings.TrimSpace(input.PriorContext),
		GateFeedback:           strings.TrimSpace(input.GateFeedback),
		ValidationCommandsText: joinLines(input.ValidationCommands),
	}

	if renderInput.MissionID == "" {
		return "", fmt.Errorf("mission id is required for implementer prompt")
	}
	if renderInput.Title == "" {
		renderInput.Title = renderInput.MissionID
	}
	if renderInput.Classification == "" {
		renderInput.Classification = MissionClassificationREDAlert
	}
	if renderInput.UseCasesText == "" {
		renderInput.UseCasesText = "(none provided)"
	}
	if renderInput.WorktreePath == "" {
		renderInput.WorktreePath = "."
	}
	if renderInput.AcceptanceCriterion == "" {
		renderInput.AcceptanceCriterion = "(none provided)"
	}
	if renderInput.MissionSpec == "" {
		renderInput.MissionSpec = "(none provided)"
	}
	if renderInput.PriorContext == "" {
		renderInput.PriorContext = "(none provided)"
	}
	if renderInput.GateFeedback == "" {
		renderInput.GateFeedback = "(none provided)"
	}
	if renderInput.ValidationCommandsText == "" {
		renderInput.ValidationCommandsText = "(none provided)"
	}

	demoInstruction, err := renderTemplate("demo_token_instruction.tmpl", struct {
		MissionID      string
		Title          string
		Classification string
	}{
		MissionID:      renderInput.MissionID,
		Title:          renderInput.Title,
		Classification: renderInput.Classification,
	})
	if err != nil {
		return "", fmt.Errorf("render demo token instruction: %w", err)
	}
	renderInput.DemoTokenInstruction = demoInstruction

	return renderTemplate(templateName, renderInput)
}

func renderTemplate(templateName string, data any) (string, error) {
	var prompt bytes.Buffer
	if err := promptTemplates.ExecuteTemplate(&prompt, templateName, data); err != nil {
		return "", fmt.Errorf("render %s: %w", templateName, err)
	}
	return prompt.String(), nil
}

func joinDependencies(dependencies []string) string {
	normalized := make([]string, 0, len(dependencies))
	for _, dependency := range dependencies {
		dependency = strings.TrimSpace(dependency)
		if dependency == "" {
			continue
		}
		normalized = append(normalized, dependency)
	}
	return strings.Join(normalized, ", ")
}

func joinLines(values []string) string {
	normalized := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		normalized = append(normalized, value)
	}
	if len(normalized) == 0 {
		return ""
	}
	return "- " + strings.Join(normalized, "\n- ")
}
