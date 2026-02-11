package state

import (
	"errors"
	"fmt"
	"strings"
)

const (
	// ClassificationREDAlert requires tests plus commands or diff refs.
	ClassificationREDAlert = "RED_ALERT"
	// ClassificationStandardOps requires at least one proof artifact.
	ClassificationStandardOps = "STANDARD_OPS"
)

// ClassifiedMission captures the proof-relevant mission metadata.
type ClassifiedMission struct {
	ID             string
	Classification string
}

// DemoTokenProof is the normalized proof payload extracted from a demo token.
type DemoTokenProof struct {
	Tests       []string
	Commands    []string
	ManualSteps []string
	DiffRefs    []string
}

// ClassificationProofError describes why proof evidence did not satisfy mission classification rules.
type ClassificationProofError struct {
	MissionID      string
	Classification string
	Reason         string
}

func (e *ClassificationProofError) Error() string {
	return fmt.Sprintf(
		"classification proof validation failed for mission %q (%s): %s",
		e.MissionID,
		e.Classification,
		e.Reason,
	)
}

// ValidateClassificationProof enforces deterministic proof requirements by mission classification.
func ValidateClassificationProof(mission ClassifiedMission, token DemoTokenProof) error {
	missionID := strings.TrimSpace(mission.ID)
	if missionID == "" {
		return errors.New("mission id must not be empty")
	}

	classification := strings.ToUpper(strings.TrimSpace(mission.Classification))
	if classification == "" {
		return &ClassificationProofError{
			MissionID:      missionID,
			Classification: classification,
			Reason:         "classification must not be empty",
		}
	}

	hasTests := len(normalizeEvidence(token.Tests)) > 0
	hasCommands := len(normalizeEvidence(token.Commands)) > 0
	hasManualSteps := len(normalizeEvidence(token.ManualSteps)) > 0
	hasDiffRefs := len(normalizeEvidence(token.DiffRefs)) > 0

	switch classification {
	case ClassificationREDAlert:
		if !hasTests {
			return &ClassificationProofError{
				MissionID:      missionID,
				Classification: classification,
				Reason:         "RED_ALERT requires a tests section",
			}
		}
		if !hasCommands && !hasDiffRefs {
			return &ClassificationProofError{
				MissionID:      missionID,
				Classification: classification,
				Reason:         "RED_ALERT requires commands or diff_refs in addition to tests",
			}
		}
	case ClassificationStandardOps:
		if !hasCommands && !hasManualSteps && !hasDiffRefs {
			return &ClassificationProofError{
				MissionID:      missionID,
				Classification: classification,
				Reason:         "STANDARD_OPS requires commands, manual_steps, or diff_refs",
			}
		}
	default:
		return &ClassificationProofError{
			MissionID:      missionID,
			Classification: classification,
			Reason:         "unsupported classification",
		}
	}

	return nil
}

func normalizeEvidence(items []string) []string {
	normalized := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		normalized = append(normalized, item)
	}
	return normalized
}
