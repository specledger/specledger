package memory

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// RenderKnowledgeMarkdown generates knowledge.md from promoted entries.
// The output file is at basePath/../knowledge.md (i.e., .specledger/memory/knowledge.md).
func RenderKnowledgeMarkdown(store *Store, outputPath string) error {
	promoted, err := store.ListPromoted()
	if err != nil {
		return fmt.Errorf("failed to list promoted entries: %w", err)
	}

	content := renderMarkdown(promoted)

	// Ensure directory exists
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	return os.WriteFile(outputPath, []byte(content), 0600)
}

// DefaultKnowledgePath returns the default path for knowledge.md.
func DefaultKnowledgePath() string {
	return filepath.Join(".specledger", "memory", "knowledge.md")
}

// DefaultCachePath returns the default path for the cache directory.
func DefaultCachePath() string {
	return filepath.Join(".specledger", "memory", "cache")
}

func renderMarkdown(entries []*KnowledgeEntry) string {
	var sb strings.Builder

	sb.WriteString("# Project Knowledge Base\n\n")
	sb.WriteString("> Auto-generated from promoted knowledge entries. Do not edit manually.\n")
	sb.WriteString(fmt.Sprintf("> Last updated: %s\n\n", time.Now().UTC().Format(time.RFC3339)))

	if len(entries) == 0 {
		sb.WriteString("_No promoted knowledge entries yet._\n")
		return sb.String()
	}

	// Group entries by tag category
	groups := groupByFirstTag(entries)

	// Sort group names for consistent output
	var groupNames []string
	for name := range groups {
		groupNames = append(groupNames, name)
	}
	sort.Strings(groupNames)

	for _, groupName := range groupNames {
		groupEntries := groups[groupName]

		// Sort entries within group by composite score descending
		sort.Slice(groupEntries, func(i, j int) bool {
			return groupEntries[i].Scores.Composite > groupEntries[j].Scores.Composite
		})

		sb.WriteString(fmt.Sprintf("## %s\n\n", titleCase(groupName)))

		for _, entry := range groupEntries {
			sb.WriteString(fmt.Sprintf("### %s\n\n", entry.Title))
			sb.WriteString(entry.Description)
			sb.WriteString("\n\n")

			// Source and score metadata
			source := ""
			if entry.SourceSessionID != "" {
				source = fmt.Sprintf("session %s", entry.SourceSessionID)
			}
			if entry.SourceBranch != "" {
				if source != "" {
					source += " on "
				}
				source += entry.SourceBranch
			}
			if source == "" {
				source = "unknown"
			}

			sb.WriteString(fmt.Sprintf("_Source: %s | Score: %.1f | Tags: %s_\n\n",
				source, entry.Scores.Composite, strings.Join(entry.Tags, ", ")))
		}
	}

	return sb.String()
}

// titleCase capitalizes the first letter of a string.
func titleCase(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// groupByFirstTag groups entries by their first tag.
func groupByFirstTag(entries []*KnowledgeEntry) map[string][]*KnowledgeEntry {
	groups := make(map[string][]*KnowledgeEntry)
	for _, entry := range entries {
		key := "general"
		if len(entry.Tags) > 0 {
			key = entry.Tags[0]
		}
		groups[key] = append(groups[key], entry)
	}
	return groups
}
