package commission

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yuin/goldmark"
	gast "github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	extast "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/text"
)

// ParseFile reads and parses a PRD markdown file into a Commission.
func ParseFile(ctx context.Context, path string) (*Commission, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read PRD file %q: %w", path, err)
	}

	title := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	return ParseMarkdown(ctx, title, string(content))
}

// ParseMarkdown parses markdown content into a Commission model.
func ParseMarkdown(ctx context.Context, title, markdown string) (*Commission, error) {
	if strings.TrimSpace(markdown) == "" {
		return nil, fmt.Errorf("markdown content is empty")
	}

	parser := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
	)

	source := []byte(markdown)
	doc := parser.Parser().Parse(text.NewReader(source))
	useCases := extractUseCases(source, doc)
	criteria := extractAcceptanceCriteria(source, doc)
	functionalGroups := extractFunctionalGroups(source, doc)
	scope := extractScopeBoundaries(source, doc)

	_ = ctx
	return &Commission{
		Title:              title,
		Status:             StatusPlanning,
		UseCases:           useCases,
		AcceptanceCriteria: criteria,
		FunctionalGroups:   functionalGroups,
		ScopeBoundaries:    scope,
		PRDContent:         markdown,
		CreatedAt:          time.Now().UTC(),
	}, nil
}

func extractUseCases(source []byte, doc gast.Node) []UseCase {
	useCases := make([]UseCase, 0)

	_ = gast.Walk(doc, func(node gast.Node, entering bool) (gast.WalkStatus, error) {
		if !entering {
			return gast.WalkContinue, nil
		}

		table, ok := node.(*extast.Table)
		if !ok {
			return gast.WalkContinue, nil
		}

		header, ok := table.FirstChild().(*extast.TableHeader)
		if !ok {
			return gast.WalkSkipChildren, nil
		}

		headers := tableRowCells(source, header)
		ucIDIndex := findHeader(headers, "uc id")
		if ucIDIndex == -1 {
			return gast.WalkSkipChildren, nil
		}

		titleIndex := findHeaderAny(headers, []string{"title", "use case", "use case title"})
		descriptionIndex := findHeaderAny(headers, []string{"description", "details"})

		for row := header.NextSibling(); row != nil; row = row.NextSibling() {
			tableRow, ok := row.(*extast.TableRow)
			if !ok {
				continue
			}

			cells := tableRowCells(source, tableRow)
			ucID := cellAt(cells, ucIDIndex)
			if ucID == "" {
				continue
			}

			title := cellAt(cells, titleIndex)
			if title == "" {
				title = ucID
			}

			useCases = append(useCases, UseCase{
				ID:          ucID,
				Title:       title,
				Description: cellAt(cells, descriptionIndex),
			})
		}

		return gast.WalkSkipChildren, nil
	})

	return useCases
}

func extractAcceptanceCriteria(source []byte, doc gast.Node) []AC {
	criteria := make([]AC, 0)
	acIndex := 1

	_ = gast.Walk(doc, func(node gast.Node, entering bool) (gast.WalkStatus, error) {
		if !entering {
			return gast.WalkContinue, nil
		}

		item, ok := node.(*gast.ListItem)
		if !ok || !containsTaskCheckbox(item) {
			return gast.WalkContinue, nil
		}

		description := strings.TrimSpace(plainText(source, item))
		if description == "" {
			return gast.WalkContinue, nil
		}

		criteria = append(criteria, AC{
			ID:          fmt.Sprintf("AC-%03d", acIndex),
			Description: description,
			Status:      "open",
		})
		acIndex++
		return gast.WalkSkipChildren, nil
	})

	return criteria
}

func extractFunctionalGroups(source []byte, doc gast.Node) []string {
	groups := make([]string, 0)

	_ = gast.Walk(doc, func(node gast.Node, entering bool) (gast.WalkStatus, error) {
		if !entering {
			return gast.WalkContinue, nil
		}

		heading, ok := node.(*gast.Heading)
		if !ok || heading.Level != 2 {
			return gast.WalkContinue, nil
		}

		text := strings.TrimSpace(plainText(source, heading))
		if text != "" {
			groups = append(groups, text)
		}
		return gast.WalkSkipChildren, nil
	})

	return groups
}

func extractScopeBoundaries(source []byte, doc gast.Node) ScopeConfig {
	scope := ScopeConfig{
		InScope:    []string{},
		OutOfScope: []string{},
	}

	var currentHeading string
	for node := doc.FirstChild(); node != nil; node = node.NextSibling() {
		if heading, ok := node.(*gast.Heading); ok && heading.Level == 2 {
			currentHeading = normalizeHeader(strings.TrimSpace(plainText(source, heading)))
			continue
		}

		list, ok := node.(*gast.List)
		if !ok {
			continue
		}

		switch currentHeading {
		case "in scope":
			scope.InScope = append(scope.InScope, listItemsText(source, list)...)
		case "out of scope":
			scope.OutOfScope = append(scope.OutOfScope, listItemsText(source, list)...)
		}
	}

	return scope
}

func tableRowCells(source []byte, row gast.Node) []string {
	cells := make([]string, 0)
	for child := row.FirstChild(); child != nil; child = child.NextSibling() {
		if _, ok := child.(*extast.TableCell); !ok {
			continue
		}
		cells = append(cells, strings.TrimSpace(plainText(source, child)))
	}
	return cells
}

func listItemsText(source []byte, list *gast.List) []string {
	items := make([]string, 0)
	for child := list.FirstChild(); child != nil; child = child.NextSibling() {
		item, ok := child.(*gast.ListItem)
		if !ok {
			continue
		}
		text := strings.TrimSpace(plainText(source, item))
		if text != "" {
			items = append(items, text)
		}
	}
	return items
}

func containsTaskCheckbox(node gast.Node) bool {
	hasCheckbox := false
	_ = gast.Walk(node, func(inner gast.Node, entering bool) (gast.WalkStatus, error) {
		if entering {
			if _, ok := inner.(*extast.TaskCheckBox); ok {
				hasCheckbox = true
				return gast.WalkStop, nil
			}
		}
		return gast.WalkContinue, nil
	})
	return hasCheckbox
}

func plainText(source []byte, node gast.Node) string {
	var builder strings.Builder
	_ = gast.Walk(node, func(inner gast.Node, entering bool) (gast.WalkStatus, error) {
		if !entering {
			return gast.WalkContinue, nil
		}

		switch value := inner.(type) {
		case *gast.Text:
			builder.Write(value.Segment.Value(source))
			if value.HardLineBreak() || value.SoftLineBreak() {
				builder.WriteByte(' ')
			}
		case *gast.String:
			builder.Write(value.Value)
		}

		return gast.WalkContinue, nil
	})

	return normalizeWhitespace(builder.String())
}

func normalizeHeader(input string) string {
	return normalizeWhitespace(strings.ToLower(input))
}

func normalizeWhitespace(input string) string {
	return strings.Join(strings.Fields(input), " ")
}

func findHeader(headers []string, target string) int {
	return findHeaderAny(headers, []string{target})
}

func findHeaderAny(headers []string, targets []string) int {
	for i, header := range headers {
		normalized := normalizeHeader(header)
		for _, target := range targets {
			if normalized == normalizeHeader(target) {
				return i
			}
		}
	}
	return -1
}

func cellAt(cells []string, index int) string {
	if index < 0 || index >= len(cells) {
		return ""
	}
	return strings.TrimSpace(cells[index])
}
