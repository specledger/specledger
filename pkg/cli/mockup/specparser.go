package mockup

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

var (
	h1Pattern    = regexp.MustCompile(`^#\s+(.+)`)
	h2Pattern    = regexp.MustCompile(`^##\s+(.+)`)
	h3Pattern    = regexp.MustCompile(`^###\s+(.+)`)
	storyPattern = regexp.MustCompile(`(?i)user\s+stor`)
	reqPattern   = regexp.MustCompile(`(?i)functional\s+requirement`)
)

// ParseSpec reads and parses a spec.md file into SpecContent.
func ParseSpec(specPath string) (*SpecContent, error) {
	data, err := os.ReadFile(specPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read spec: %w", err)
	}

	content := string(data)
	if strings.TrimSpace(content) == "" {
		return nil, fmt.Errorf("spec file is empty")
	}

	sc := &SpecContent{
		FullContent: content,
	}

	lines := strings.Split(content, "\n")

	// Extract title from first H1
	for _, line := range lines {
		if m := h1Pattern.FindStringSubmatch(line); len(m) > 1 {
			sc.Title = strings.TrimSpace(m[1])
			break
		}
	}

	// Extract user stories and requirements
	sc.UserStories = extractSection(lines, storyPattern)
	sc.Requirements = extractSection(lines, reqPattern)

	return sc, nil
}

// extractSection finds sections matching the header pattern and collects their content.
func extractSection(lines []string, headerPattern *regexp.Regexp) []string {
	var results []string
	inSection := false
	var currentItem strings.Builder
	depth := 0 // H2=2, H3=3

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check if this is a section header matching our pattern
		if h2m := h2Pattern.FindStringSubmatch(line); len(h2m) > 1 {
			if inSection {
				// New H2 header: flush current section
				if currentItem.Len() > 0 {
					results = append(results, strings.TrimSpace(currentItem.String()))
					currentItem.Reset()
				}
			}
			if headerPattern.MatchString(h2m[1]) {
				inSection = true
				depth = 2
				continue
			}
			inSection = false
			continue
		}

		if h3m := h3Pattern.FindStringSubmatch(line); len(h3m) > 1 {
			if inSection && depth >= 2 {
				// New H3 inside our section: start a new item
				if currentItem.Len() > 0 {
					results = append(results, strings.TrimSpace(currentItem.String()))
					currentItem.Reset()
				}
				currentItem.WriteString(h3m[1])
				currentItem.WriteString(": ")
				continue
			}
			if headerPattern.MatchString(h3m[1]) {
				inSection = true
				depth = 3
				continue
			}
		}

		if !inSection {
			continue
		}

		// Collect content within the section
		if trimmed == "" {
			continue
		}

		// Collect bullet points and descriptions
		if strings.HasPrefix(trimmed, "-") || strings.HasPrefix(trimmed, "*") {
			item := strings.TrimLeft(trimmed, "-* ")
			if currentItem.Len() > 0 {
				currentItem.WriteString(" ")
			}
			currentItem.WriteString(item)
		} else if strings.HasPrefix(trimmed, "**") {
			// Bold text like "**As a** user..."
			if currentItem.Len() > 0 {
				results = append(results, strings.TrimSpace(currentItem.String()))
				currentItem.Reset()
			}
			currentItem.WriteString(trimmed)
		} else {
			if currentItem.Len() > 0 {
				currentItem.WriteString(" ")
			}
			currentItem.WriteString(trimmed)
		}
	}

	// Flush last item
	if currentItem.Len() > 0 {
		results = append(results, strings.TrimSpace(currentItem.String()))
	}

	return results
}
