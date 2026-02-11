package commission

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestParseMarkdownExtractsUseCasesCriteriaAndGroups(t *testing.T) {
	t.Parallel()

	markdown := `
# Ship Commander 3 PRD

## Commission Management

| UC ID | Title | Description |
|-------|-------|-------------|
| UC-COMM-01 | Parse PRD | Parse markdown into Commission |
| UC-COMM-02 | Persist Commission | Save commission using bd create |

## Acceptance Criteria

- [ ] Parser extracts use cases from markdown table
- [x] Parser extracts acceptance criteria from checkbox list

## In Scope
- Commission parsing
- Commission persistence

## Out of Scope
- UI implementation
`

	commission, err := ParseMarkdown(context.Background(), "PRD Title", markdown)
	if err != nil {
		t.Fatalf("parse markdown: %v", err)
	}

	if commission.Status != StatusPlanning {
		t.Fatalf("status = %q, want %q", commission.Status, StatusPlanning)
	}
	if len(commission.UseCases) != 2 {
		t.Fatalf("use cases = %d, want 2", len(commission.UseCases))
	}
	if commission.UseCases[0].ID != "UC-COMM-01" {
		t.Fatalf("use case id = %q, want %q", commission.UseCases[0].ID, "UC-COMM-01")
	}
	if commission.UseCases[1].Title != "Persist Commission" {
		t.Fatalf("use case title = %q, want %q", commission.UseCases[1].Title, "Persist Commission")
	}
	if len(commission.AcceptanceCriteria) != 2 {
		t.Fatalf("acceptance criteria = %d, want 2", len(commission.AcceptanceCriteria))
	}
	if commission.AcceptanceCriteria[0].Description != "Parser extracts use cases from markdown table" {
		t.Fatalf("first acceptance criterion = %q", commission.AcceptanceCriteria[0].Description)
	}
	if len(commission.FunctionalGroups) < 2 {
		t.Fatalf("functional groups = %v, want at least 2 entries", commission.FunctionalGroups)
	}
	if commission.FunctionalGroups[0] != "Commission Management" {
		t.Fatalf("first functional group = %q, want %q", commission.FunctionalGroups[0], "Commission Management")
	}
	if len(commission.ScopeBoundaries.InScope) != 2 {
		t.Fatalf("in-scope count = %d, want 2", len(commission.ScopeBoundaries.InScope))
	}
	if len(commission.ScopeBoundaries.OutOfScope) != 1 {
		t.Fatalf("out-of-scope count = %d, want 1", len(commission.ScopeBoundaries.OutOfScope))
	}
}

func TestParseFileUsesFilenameAsFallbackTitle(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "mission-prd.md")
	markdown := `
## Core
| UC ID | Title |
| --- | --- |
| UC-COMM-01 | Parse PRD |
`
	if err := os.WriteFile(path, []byte(markdown), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	commission, err := ParseFile(context.Background(), path)
	if err != nil {
		t.Fatalf("parse file: %v", err)
	}
	if commission.Title != "mission-prd" {
		t.Fatalf("title = %q, want %q", commission.Title, "mission-prd")
	}
	if len(commission.UseCases) != 1 {
		t.Fatalf("use cases = %d, want 1", len(commission.UseCases))
	}
}
