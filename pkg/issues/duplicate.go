package issues

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/texttheater/golang-levenshtein/levenshtein"
)

// DefaultSimilarityThreshold is the default threshold for duplicate detection (80%)
const DefaultSimilarityThreshold = 0.8

// DuplicateResult contains information about a potential duplicate
type DuplicateResult struct {
	Issue      Issue
	Similarity float64
}

// CalculateSimilarity calculates the Levenshtein similarity ratio between two strings
func CalculateSimilarity(s1, s2 string) float64 {
	if s1 == s2 {
		return 1.0
	}
	if s1 == "" || s2 == "" {
		return 0.0
	}

	// Normalize strings for comparison
	s1 = strings.ToLower(strings.TrimSpace(s1))
	s2 = strings.ToLower(strings.TrimSpace(s2))

	// Calculate Levenshtein distance
	distance := levenshtein.DistanceForStrings([]rune(s1), []rune(s2), levenshtein.DefaultOptions)

	// Convert to similarity ratio
	maxLen := max(len(s1), len(s2))
	if maxLen == 0 {
		return 1.0
	}

	return 1.0 - float64(distance)/float64(maxLen)
}

// FindSimilarIssues finds issues with similar titles to the given title
func FindSimilarIssues(title string, issues []Issue, threshold float64) []DuplicateResult {
	var duplicates []DuplicateResult

	title = strings.ToLower(strings.TrimSpace(title))

	for _, issue := range issues {
		issueTitle := strings.ToLower(strings.TrimSpace(issue.Title))
		similarity := CalculateSimilarity(title, issueTitle)

		if similarity >= threshold {
			duplicates = append(duplicates, DuplicateResult{
				Issue:      issue,
				Similarity: similarity,
			})
		}
	}

	return duplicates
}

// FindSimilarIssuesAcrossSpecs finds similar issues across all specs
func FindSimilarIssuesAcrossSpecs(title string, store *Store, threshold float64) ([]DuplicateResult, error) {
	// Get all issues from the store
	filter := ListFilter{All: true}
	issues, err := store.List(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list issues: %w", err)
	}

	return FindSimilarIssues(title, issues, threshold), nil
}

// FormatDuplicateWarning formats a warning message for duplicate issues
func FormatDuplicateWarning(duplicates []DuplicateResult) string {
	if len(duplicates) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("Potential duplicate issues found:\n\n")

	for _, dup := range duplicates {
		sb.WriteString(fmt.Sprintf("  %s (%.0f%% similar)\n", dup.Issue.ID, dup.Similarity*100))
		sb.WriteString(fmt.Sprintf("    Title: %s\n", dup.Issue.Title))
		sb.WriteString(fmt.Sprintf("    Spec: %s\n", dup.Issue.SpecContext))
		sb.WriteString(fmt.Sprintf("    Status: %s\n", dup.Issue.Status))
		sb.WriteString("\n")
	}

	sb.WriteString("Use --force to create anyway.\n")

	return sb.String()
}

// CheckDuplicateResult contains the result of a duplicate check
type CheckDuplicateResult struct {
	HasDuplicates bool
	Duplicates    []DuplicateResult
}

// CheckDuplicatesForCreate checks for duplicates when creating a new issue
func CheckDuplicatesForCreate(title, specContext string, threshold float64) (*CheckDuplicateResult, error) {
	// First check within the same spec
	store, err := NewStore(StoreOptions{SpecContext: specContext})
	if err != nil {
		return nil, fmt.Errorf("failed to create store: %w", err)
	}

	issues, err := store.List(ListFilter{})
	if err != nil {
		return nil, fmt.Errorf("failed to list issues: %w", err)
	}

	duplicates := FindSimilarIssues(title, issues, threshold)

	return &CheckDuplicateResult{
		HasDuplicates: len(duplicates) > 0,
		Duplicates:    duplicates,
	}, nil
}

// CheckDuplicatesAcrossAllSpecs checks for duplicates across all specs
func CheckDuplicatesAcrossAllSpecs(title string, threshold float64) (*CheckDuplicateResult, error) {
	// This requires iterating through all spec directories
	basePath := "specledger"
	entries, err := listSpecDirectories(basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to list spec directories: %w", err)
	}

	var allDuplicates []DuplicateResult

	for _, specContext := range entries {
		store, err := NewStore(StoreOptions{
			BasePath:    basePath,
			SpecContext: specContext,
		})
		if err != nil {
			continue
		}

		issues, err := store.List(ListFilter{})
		if err != nil {
			continue
		}

		duplicates := FindSimilarIssues(title, issues, threshold)
		allDuplicates = append(allDuplicates, duplicates...)
	}

	return &CheckDuplicateResult{
		HasDuplicates: len(allDuplicates) > 0,
		Duplicates:    allDuplicates,
	}, nil
}

// listSpecDirectories lists all spec directories in the base path
func listSpecDirectories(basePath string) ([]string, error) {
	entries, err := os.ReadDir(basePath)
	if err != nil {
		return nil, err
	}

	var specs []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		// Check if it matches the spec pattern (###-name)
		if specBranchPattern.MatchString(name) {
			// Check if issues.jsonl exists
			issuesPath := filepath.Join(basePath, name, "issues.jsonl")
			if _, err := os.Stat(issuesPath); err == nil {
				specs = append(specs, name)
			}
		}
	}

	return specs, nil
}
