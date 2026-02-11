package commander

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

const missionClassificationPromptTemplate = `You are the Commander in Ship Commander 3.

Assess mission risk for execution routing:
- RED_ALERT: behavior/correctness changes (full TDD)
- STANDARD_OPS: non-behavioral work (fast path)

Mission Context
- mission_id: {{ .MissionID }}
- title: {{ .Title }}
- use_case: {{ .UseCase }}
- commission_title: {{ .CommissionTitle }}
- domain: {{ .Domain }}
- dependencies: {{ .DependenciesText }}

Functional Requirements (Captain)
{{ .FunctionalRequirements }}

Design Requirements (Design Officer)
{{ .DesignRequirements }}

Decision Rules
1. If behavior/correctness is affected, classify RED_ALERT.
2. Otherwise classify STANDARD_OPS.
3. If uncertain, default to RED_ALERT.

Criteria
- RED_ALERT: business_logic, api_changes, auth_security, data_integrity, bug_fix
- STANDARD_OPS: styling, non_behavioral_refactor, tooling, documentation

Return ONLY YAML with this shape:
mission_id: "<mission id>"
title: "<mission title>"
classification: "RED_ALERT" | "STANDARD_OPS"
rationale:
  affects_behavior: true | false
  criteria_matched:
    - "<criterion>"
  risk_assessment: "<2-3 sentence explanation>"
  confidence: "high" | "medium" | "low"
confidence: "high" | "medium" | "low"
`

var classificationPromptTmpl = template.Must(
	template.New("mission-classification").Parse(missionClassificationPromptTemplate),
)

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

	var prompt bytes.Buffer
	if err := classificationPromptTmpl.Execute(&prompt, renderInput); err != nil {
		return "", fmt.Errorf("render classification prompt: %w", err)
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
