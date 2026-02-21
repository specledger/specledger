package revise

import (
	"encoding/json"
	"fmt"
	"os"
)

// ParseFixture reads and parses a fixture JSON file for --auto mode.
func ParseFixture(path string) (*AutoFixture, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read fixture file %q: %w", path, err)
	}

	var fixture AutoFixture
	if err := json.Unmarshal(data, &fixture); err != nil {
		return nil, fmt.Errorf("failed to parse fixture file %q: %w", path, err)
	}

	return &fixture, nil
}

// MatchFixtureComments matches fixture entries to fetched review comments by file_path + selected_text.
// Returns matched ProcessedComments (with guidance from fixture) and warning messages for unmatched entries.
// Warnings are intended for stderr so they do not contaminate stdout prompt output.
func MatchFixtureComments(fixture *AutoFixture, comments []ReviewComment) ([]ProcessedComment, []string) {
	processed := make([]ProcessedComment, 0, len(fixture.Comments))
	warnings := make([]string, 0)

	for _, fc := range fixture.Comments {
		match := findComment(comments, fc.FilePath, fc.SelectedText)
		if match == nil {
			warnings = append(warnings, fmt.Sprintf(
				"no comment matched: file_path=%q selected_text=%q",
				fc.FilePath, fc.SelectedText,
			))
			continue
		}

		processed = append(processed, ProcessedComment{
			Comment:  *match,
			Guidance: fc.Guidance,
			Index:    len(processed) + 1,
		})
	}

	return processed, warnings
}

// findComment returns the first comment matching both file_path and selected_text, or nil.
func findComment(comments []ReviewComment, filePath, selectedText string) *ReviewComment {
	for i := range comments {
		if comments[i].FilePath == filePath && comments[i].SelectedText == selectedText {
			return &comments[i]
		}
	}
	return nil
}
