package demo

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidatorValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		mission            Mission
		writeToken         bool
		tokenContent       string
		worktreeFiles      map[string]string
		wantValid          bool
		wantReasonContains string
	}{
		{
			name:               "fails when demo token file is missing",
			mission:            Mission{ID: "MISSION-42", Classification: ClassificationREDAlert},
			writeToken:         false,
			wantValid:          false,
			wantReasonContains: "does not exist",
		},
		{
			name:       "fails when required frontmatter field is missing",
			mission:    Mission{ID: "MISSION-42", Classification: ClassificationREDAlert},
			writeToken: true,
			tokenContent: tokenMarkdown(
				"MISSION-42",
				ClassificationREDAlert,
				[]string{
					"### commands",
					"- `go test ./...`",
					"",
					"### tests",
					"- file: `internal/demo/validator_test.go`",
					"  - added_tests:",
					"    - \"TestValidatorValidate\"",
					"  - passing: true",
				},
				map[string]string{
					"title":          "Validate demo token",
					"classification": ClassificationREDAlert,
					"status":         "complete",
					"created_at":     "2026-02-10T14:30:00Z",
					"agent_id":       "",
				},
			),
			wantValid:          false,
			wantReasonContains: "agent_id",
		},
		{
			name:       "fails when mission id does not match",
			mission:    Mission{ID: "MISSION-42", Classification: ClassificationREDAlert},
			writeToken: true,
			tokenContent: tokenMarkdown(
				"MISSION-24",
				ClassificationREDAlert,
				[]string{
					"### commands",
					"- `go test ./...`",
					"",
					"### tests",
					"- file: `internal/demo/validator_test.go`",
					"  - added_tests:",
					"    - \"TestValidatorValidate\"",
					"  - passing: true",
				},
				nil,
			),
			wantValid:          false,
			wantReasonContains: "does not match mission",
		},
		{
			name:       "fails when diff_refs points to non-existent file",
			mission:    Mission{ID: "MISSION-42", Classification: ClassificationREDAlert},
			writeToken: true,
			tokenContent: tokenMarkdown(
				"MISSION-42",
				ClassificationREDAlert,
				[]string{
					"### tests",
					"- file: `internal/demo/validator_test.go`",
					"  - added_tests:",
					"    - \"TestValidatorValidate\"",
					"  - passing: true",
					"",
					"### diff_refs",
					"- `internal/demo/missing.go` — reference not present",
				},
				nil,
			),
			wantValid:          false,
			wantReasonContains: "diff_ref file not found",
		},
		{
			name:       "fails RED_ALERT without tests section",
			mission:    Mission{ID: "MISSION-42", Classification: ClassificationREDAlert},
			writeToken: true,
			tokenContent: tokenMarkdown(
				"MISSION-42",
				ClassificationREDAlert,
				[]string{
					"### commands",
					"- `go test ./...`",
				},
				nil,
			),
			wantValid:          false,
			wantReasonContains: "requires a tests section",
		},
		{
			name:       "fails RED_ALERT without commands or diff_refs",
			mission:    Mission{ID: "MISSION-42", Classification: ClassificationREDAlert},
			writeToken: true,
			tokenContent: tokenMarkdown(
				"MISSION-42",
				ClassificationREDAlert,
				[]string{
					"### tests",
					"- file: `internal/demo/validator_test.go`",
					"  - added_tests:",
					"    - \"TestValidatorValidate\"",
					"  - passing: true",
				},
				nil,
			),
			wantValid:          false,
			wantReasonContains: "requires commands or diff_refs",
		},
		{
			name:       "passes RED_ALERT with tests and commands evidence",
			mission:    Mission{ID: "MISSION-42", Classification: ClassificationREDAlert},
			writeToken: true,
			tokenContent: tokenMarkdown(
				"MISSION-42",
				ClassificationREDAlert,
				[]string{
					"### commands",
					"- `go test ./internal/demo/...`",
					"  - exit_code: 0",
					"  - summary: \"validator tests passed\"",
					"",
					"### tests",
					"- file: `internal/demo/validator_test.go`",
					"  - added_tests:",
					"    - \"TestValidatorValidate\"",
					"  - passing: true",
				},
				nil,
			),
			wantValid: true,
		},
		{
			name:       "passes RED_ALERT with tests and diff_refs evidence",
			mission:    Mission{ID: "MISSION-42", Classification: ClassificationREDAlert},
			writeToken: true,
			tokenContent: tokenMarkdown(
				"MISSION-42",
				ClassificationREDAlert,
				[]string{
					"### tests",
					"- file: `internal/demo/validator_test.go`",
					"  - added_tests:",
					"    - \"TestValidatorValidate\"",
					"  - passing: true",
					"",
					"### diff_refs",
					"- `internal/demo/validator.go` — validator implementation",
				},
				nil,
			),
			worktreeFiles: map[string]string{
				"internal/demo/validator.go": "package demo\n",
			},
			wantValid: true,
		},
		{
			name:       "fails STANDARD_OPS without evidence section",
			mission:    Mission{ID: "MISSION-42", Classification: ClassificationStandardOps},
			writeToken: true,
			tokenContent: tokenMarkdown(
				"MISSION-42",
				ClassificationStandardOps,
				[]string{
					"### notes",
					"- housekeeping only",
				},
				nil,
			),
			wantValid:          false,
			wantReasonContains: "STANDARD_OPS requires",
		},
		{
			name:       "passes STANDARD_OPS with manual_steps evidence",
			mission:    Mission{ID: "MISSION-42", Classification: ClassificationStandardOps},
			writeToken: true,
			tokenContent: tokenMarkdown(
				"MISSION-42",
				ClassificationStandardOps,
				[]string{
					"### manual_steps",
					"1. Run `make lint`.",
					"2. Run `make test`.",
					"3. Verify command exits 0.",
				},
				nil,
			),
			wantValid: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()

			for relPath, contents := range tt.worktreeFiles {
				absPath := filepath.Join(root, relPath)
				require.NoError(t, os.MkdirAll(filepath.Dir(absPath), 0o750))
				require.NoError(t, os.WriteFile(absPath, []byte(contents), 0o600))
			}

			if tt.writeToken {
				writeDemoToken(t, root, tt.mission.ID, tt.tokenContent)
			}

			validator := NewValidator()
			result := validator.Validate(context.Background(), tt.mission, root)

			assert.Equal(t, tt.wantValid, result.Valid)
			assert.Equal(t, filepath.Join(root, "demo", "MISSION-"+tt.mission.ID+".md"), result.TokenPath)
			if tt.wantValid {
				assert.Empty(t, result.Reason)
			} else {
				assert.Contains(t, result.Reason, tt.wantReasonContains)
			}
		})
	}
}

func writeDemoToken(t *testing.T, worktreeRoot, missionID, content string) {
	t.Helper()
	tokenPath := filepath.Join(worktreeRoot, "demo", "MISSION-"+missionID+".md")
	require.NoError(t, os.MkdirAll(filepath.Dir(tokenPath), 0o750))
	require.NoError(t, os.WriteFile(tokenPath, []byte(content), 0o600))
}

func tokenMarkdown(missionID, classification string, bodyLines []string, frontmatterOverrides map[string]string) string {
	frontmatter := map[string]string{
		"title":          "Validate demo token artifact",
		"classification": classification,
		"status":         "complete",
		"created_at":     "2026-02-10T14:30:00Z",
		"agent_id":       "ensign-42",
	}
	for key, value := range frontmatterOverrides {
		frontmatter[key] = value
	}

	var frontmatterBuilder strings.Builder
	frontmatterBuilder.WriteString("---\n")
	frontmatterBuilder.WriteString(fmt.Sprintf("mission_id: %q\n", missionID))
	for _, key := range []string{"title", "classification", "status", "created_at", "agent_id"} {
		if value, ok := frontmatter[key]; ok {
			frontmatterBuilder.WriteString(fmt.Sprintf("%s: %q\n", key, value))
		}
	}
	frontmatterBuilder.WriteString("---\n\n")
	frontmatterBuilder.WriteString("## Evidence\n\n")
	frontmatterBuilder.WriteString(strings.Join(bodyLines, "\n"))
	frontmatterBuilder.WriteString("\n")

	return frontmatterBuilder.String()
}
