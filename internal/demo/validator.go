package demo

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ship-commander/sc3/internal/state"
	"gopkg.in/yaml.v3"
)

const (
	// ClassificationREDAlert requires tests plus command or diff evidence.
	ClassificationREDAlert = "RED_ALERT"
	// ClassificationStandardOps requires at least one evidence section.
	ClassificationStandardOps = "STANDARD_OPS"
)

var backtickPathPattern = regexp.MustCompile("`([^`]+)`")

// Mission identifies the mission whose demo token must be validated.
type Mission struct {
	ID             string
	Classification string
}

// ValidationResult is the deterministic pass/fail result for a demo token.
type ValidationResult struct {
	Valid     bool
	Reason    string
	TokenPath string
}

// Validator validates demo token files against the V1 schema rules.
type Validator struct{}

// NewValidator creates a demo token validator.
func NewValidator() *Validator {
	return &Validator{}
}

// Validate checks demo/MISSION-<id>.md in the mission worktree and returns pass/fail evidence.
func (v *Validator) Validate(ctx context.Context, mission Mission, worktreePath string) ValidationResult {
	tokenPath, err := tokenPathForMission(worktreePath, mission.ID)
	if err != nil {
		return failResult("", err.Error())
	}
	if err := ctx.Err(); err != nil {
		return failResult(tokenPath, fmt.Sprintf("validation canceled: %v", err))
	}

	raw, readResult, ok := readTokenFile(tokenPath)
	if !ok {
		return readResult
	}
	frontmatter, body, parseResult, ok := parseToken(raw, tokenPath)
	if !ok {
		return parseResult
	}
	classification, metadataResult, ok := validateMetadata(frontmatter, mission, tokenPath)
	if !ok {
		return metadataResult
	}

	sections := parseSections(body)
	hasDiffRefs, diffResult, ok := validateDiffRefs(sections, worktreePath, tokenPath)
	if !ok {
		return diffResult
	}

	return validateEvidenceRequirements(sections, classification, hasDiffRefs, tokenPath)
}

func failResult(tokenPath, reason string) ValidationResult {
	return ValidationResult{
		Valid:     false,
		Reason:    reason,
		TokenPath: tokenPath,
	}
}

func tokenPathForMission(worktreePath, missionID string) (string, error) {
	trimmedWorktree := strings.TrimSpace(worktreePath)
	if trimmedWorktree == "" {
		return "", fmt.Errorf("worktree path must not be empty")
	}
	trimmedMissionID := strings.TrimSpace(missionID)
	if trimmedMissionID == "" {
		return "", fmt.Errorf("mission id must not be empty")
	}
	if strings.Contains(trimmedMissionID, "/") || strings.Contains(trimmedMissionID, "\\") {
		return "", fmt.Errorf("mission id must not contain path separators")
	}
	return filepath.Join(trimmedWorktree, "demo", "MISSION-"+trimmedMissionID+".md"), nil
}

func readTokenFile(tokenPath string) (string, ValidationResult, bool) {
	// #nosec G304 -- tokenPath is deterministic from validated worktree root + mission id.
	raw, err := os.ReadFile(tokenPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", failResult(tokenPath, fmt.Sprintf("demo token file does not exist at %s", tokenPath)), false
		}
		return "", failResult(tokenPath, fmt.Sprintf("read demo token file: %v", err)), false
	}
	return string(raw), ValidationResult{}, true
}

func parseToken(raw, tokenPath string) (map[string]any, string, ValidationResult, bool) {
	frontmatterText, body, ok := splitFrontmatter(raw)
	if !ok {
		return nil, "", failResult(tokenPath, "demo token file has invalid or missing YAML frontmatter"), false
	}

	var frontmatter map[string]any
	if err := yaml.Unmarshal([]byte(frontmatterText), &frontmatter); err != nil {
		return nil, "", failResult(tokenPath, fmt.Sprintf("parse YAML frontmatter: %v", err)), false
	}
	return frontmatter, body, ValidationResult{}, true
}

func validateMetadata(frontmatter map[string]any, mission Mission, tokenPath string) (string, ValidationResult, bool) {
	requiredFields := []string{"mission_id", "title", "classification", "status", "created_at", "agent_id"}
	for _, key := range requiredFields {
		if _, ok := requiredString(frontmatter, key); !ok {
			return "", failResult(tokenPath, fmt.Sprintf("missing required frontmatter field %q", key)), false
		}
	}

	missionID, _ := requiredString(frontmatter, "mission_id")
	if missionID != strings.TrimSpace(mission.ID) {
		return "", failResult(tokenPath, fmt.Sprintf("mission_id %q does not match mission %q", missionID, strings.TrimSpace(mission.ID))), false
	}

	frontmatterClassification, _ := requiredString(frontmatter, "classification")
	classification := normalizeClassification(mission.Classification, frontmatterClassification)
	if classification == "" {
		return "", failResult(tokenPath, "classification must not be empty"), false
	}
	if !isSupportedClassification(classification) {
		return "", failResult(tokenPath, fmt.Sprintf("unsupported classification %q", classification)), false
	}
	if strings.TrimSpace(mission.Classification) != "" &&
		!strings.EqualFold(strings.TrimSpace(mission.Classification), frontmatterClassification) {
		return "", failResult(tokenPath, fmt.Sprintf(
			"classification %q does not match mission classification %q",
			frontmatterClassification,
			strings.TrimSpace(mission.Classification),
		)), false
	}

	return classification, ValidationResult{}, true
}

func validateDiffRefs(sections map[string]string, worktreePath, tokenPath string) (bool, ValidationResult, bool) {
	content, ok := sections["diff_refs"]
	if !ok {
		return false, ValidationResult{}, true
	}

	diffRefs := parseDiffRefs(content)
	if len(diffRefs) == 0 {
		return false, failResult(tokenPath, "diff_refs section has no file references"), false
	}

	for _, diffRef := range diffRefs {
		cleanPath, err := safeRelativePath(diffRef)
		if err != nil {
			return false, failResult(tokenPath, fmt.Sprintf("invalid diff_ref %q: %v", diffRef, err)), false
		}
		targetPath := filepath.Join(worktreePath, cleanPath)
		if _, err := os.Stat(targetPath); err != nil {
			if os.IsNotExist(err) {
				return false, failResult(tokenPath, fmt.Sprintf("diff_ref file not found: %s", cleanPath)), false
			}
			return false, failResult(tokenPath, fmt.Sprintf("check diff_ref %s: %v", cleanPath, err)), false
		}
	}

	return true, ValidationResult{}, true
}

func validateEvidenceRequirements(
	sections map[string]string,
	classification string,
	hasDiffRefs bool,
	tokenPath string,
) ValidationResult {
	hasCommands := hasSectionContent(sections, "commands")
	hasTests := hasSectionContent(sections, "tests")
	hasManualSteps := hasSectionContent(sections, "manual_steps")

	err := state.ValidateClassificationProof(
		state.ClassifiedMission{
			ID:             "mission",
			Classification: classification,
		},
		state.DemoTokenProof{
			Tests:       boolToEvidenceSlice(hasTests),
			Commands:    boolToEvidenceSlice(hasCommands),
			ManualSteps: boolToEvidenceSlice(hasManualSteps),
			DiffRefs:    boolToEvidenceSlice(hasDiffRefs),
		},
	)
	if err != nil {
		return failResult(tokenPath, extractValidationReason(err))
	}

	return ValidationResult{Valid: true, TokenPath: tokenPath}
}

func boolToEvidenceSlice(ok bool) []string {
	if !ok {
		return nil
	}
	return []string{"present"}
}

func extractValidationReason(err error) string {
	var proofErr *state.ClassificationProofError
	if !errors.As(err, &proofErr) {
		return err.Error()
	}
	return proofErr.Reason
}

func splitFrontmatter(markdown string) (string, string, bool) {
	normalized := strings.ReplaceAll(markdown, "\r\n", "\n")
	lines := strings.Split(normalized, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return "", "", false
	}

	for idx := 1; idx < len(lines); idx++ {
		if strings.TrimSpace(lines[idx]) == "---" {
			frontmatter := strings.Join(lines[1:idx], "\n")
			body := strings.Join(lines[idx+1:], "\n")
			return frontmatter, body, true
		}
	}

	return "", "", false
}

func requiredString(frontmatter map[string]any, key string) (string, bool) {
	raw, ok := frontmatter[key]
	if !ok {
		return "", false
	}
	value, ok := raw.(string)
	if !ok {
		return "", false
	}
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", false
	}
	return trimmed, true
}

func normalizeClassification(missionClassification, frontmatterClassification string) string {
	if strings.TrimSpace(missionClassification) != "" {
		return strings.ToUpper(strings.TrimSpace(missionClassification))
	}
	return strings.ToUpper(strings.TrimSpace(frontmatterClassification))
}

func isSupportedClassification(classification string) bool {
	return classification == ClassificationREDAlert || classification == ClassificationStandardOps
}

func parseSections(body string) map[string]string {
	normalized := strings.ReplaceAll(body, "\r\n", "\n")
	lines := strings.Split(normalized, "\n")

	sections := make(map[string]string)
	currentName := ""
	var currentLines []string

	flush := func() {
		if currentName == "" {
			return
		}
		sections[currentName] = strings.TrimSpace(strings.Join(currentLines, "\n"))
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "### ") {
			flush()
			currentName = normalizeSectionName(strings.TrimSpace(strings.TrimPrefix(trimmed, "### ")))
			currentLines = nil
			continue
		}
		if currentName != "" {
			currentLines = append(currentLines, line)
		}
	}
	flush()

	return sections
}

func normalizeSectionName(name string) string {
	return strings.ToLower(strings.ReplaceAll(name, " ", "_"))
}

func hasSectionContent(sections map[string]string, name string) bool {
	content, ok := sections[name]
	return ok && strings.TrimSpace(content) != ""
}

func parseDiffRefs(content string) []string {
	normalized := strings.ReplaceAll(content, "\r\n", "\n")
	lines := strings.Split(normalized, "\n")
	refs := make([]string, 0, len(lines))

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || (!strings.HasPrefix(trimmed, "- ") && !strings.HasPrefix(trimmed, "* ")) {
			continue
		}

		candidate := extractDiffRefPath(trimmed[2:])
		if candidate != "" {
			refs = append(refs, candidate)
		}
	}

	return refs
}

func extractDiffRefPath(listItem string) string {
	if matches := backtickPathPattern.FindStringSubmatch(listItem); len(matches) == 2 {
		return strings.TrimSpace(matches[1])
	}

	candidate := strings.TrimSpace(listItem)
	for _, separator := range []string{" — ", " - ", " – ", ":"} {
		if idx := strings.Index(candidate, separator); idx >= 0 {
			candidate = strings.TrimSpace(candidate[:idx])
			break
		}
	}
	return strings.Trim(candidate, "\"'")
}

func safeRelativePath(path string) (string, error) {
	clean := filepath.Clean(strings.TrimSpace(path))
	if clean == "." || clean == "" {
		return "", fmt.Errorf("path is empty")
	}
	if filepath.IsAbs(clean) {
		return "", fmt.Errorf("path must be relative")
	}
	if clean == ".." || strings.HasPrefix(clean, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("path escapes worktree")
	}
	return clean, nil
}
