package mockup

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	frontmatterSep = "---"
)

// LoadDesignSystem reads and parses a design-system.md file.
func LoadDesignSystem(path string) (*DesignSystem, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read design system: %w", err)
	}

	content := string(data)

	// Parse YAML frontmatter
	ds, err := parseFrontmatter(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse design system frontmatter: %w", err)
	}

	return ds, nil
}

// WriteDesignSystem writes a DesignSystem to a design-system.md file.
func WriteDesignSystem(path string, ds *DesignSystem) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	ds.LastScanned = time.Now()

	var sb strings.Builder

	// Write YAML frontmatter
	sb.WriteString(frontmatterSep + "\n")

	yamlData, err := yaml.Marshal(ds)
	if err != nil {
		return fmt.Errorf("failed to marshal frontmatter: %w", err)
	}
	sb.Write(yamlData)
	sb.WriteString(frontmatterSep + "\n\n")

	// Write markdown body
	sb.WriteString("# Design System\n\n")
	sb.WriteString("This file documents the project's design tokens and styling conventions.\n")
	sb.WriteString("The AI agent will search the codebase for existing components.\n\n")

	// Write external libraries section
	if len(ds.ExternalLibs) > 0 {
		sb.WriteString("## UI Libraries\n\n")
		for _, lib := range ds.ExternalLibs {
			sb.WriteString(fmt.Sprintf("- %s\n", lib))
		}
		sb.WriteString("\n")
	}

	// Write styling info if available
	if ds.Style != nil {
		sb.WriteString("## Styling\n\n")
		if ds.Style.CSSFramework != "" {
			sb.WriteString(fmt.Sprintf("- **CSS Framework**: %s\n", ds.Style.CSSFramework))
		}
		if ds.Style.StylingApproach != "" {
			sb.WriteString(fmt.Sprintf("- **Approach**: %s\n", ds.Style.StylingApproach))
		}
		if ds.Style.Preprocessor != "" {
			sb.WriteString(fmt.Sprintf("- **Preprocessor**: %s\n", ds.Style.Preprocessor))
		}
		sb.WriteString("\n")

		if len(ds.Style.ThemeColors) > 0 {
			sb.WriteString("### Theme Colors\n\n")
			for name, value := range ds.Style.ThemeColors {
				sb.WriteString(fmt.Sprintf("- `%s`: `%s`\n", name, value))
			}
			sb.WriteString("\n")
		}

		if len(ds.Style.FontFamilies) > 0 {
			sb.WriteString("### Fonts\n\n")
			for _, font := range ds.Style.FontFamilies {
				sb.WriteString(fmt.Sprintf("- %s\n", font))
			}
			sb.WriteString("\n")
		}

		if len(ds.Style.CSSVariables) > 0 {
			sb.WriteString("### CSS Variables\n\n")
			for _, v := range ds.Style.CSSVariables {
				sb.WriteString(fmt.Sprintf("- `%s`\n", v))
			}
			sb.WriteString("\n")
		}
	}

	return os.WriteFile(path, []byte(sb.String()), 0600)
}

// parseFrontmatter extracts and parses the YAML frontmatter from markdown content.
func parseFrontmatter(content string) (*DesignSystem, error) {
	lines := strings.SplitN(content, "\n", -1)
	if len(lines) < 3 || strings.TrimSpace(lines[0]) != frontmatterSep {
		return nil, fmt.Errorf("no frontmatter found")
	}

	// Find closing ---
	endIdx := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == frontmatterSep {
			endIdx = i
			break
		}
	}
	if endIdx < 0 {
		return nil, fmt.Errorf("unclosed frontmatter")
	}

	yamlContent := strings.Join(lines[1:endIdx], "\n")

	var ds DesignSystem
	if err := yaml.Unmarshal([]byte(yamlContent), &ds); err != nil {
		return nil, fmt.Errorf("invalid YAML frontmatter: %w", err)
	}

	return &ds, nil
}
