package mockup

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
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

	// Write markdown body — human-readable summary only.
	// All data lives in the YAML frontmatter (machine-readable, used by LoadDesignSystem).
	// The markdown body is a concise overview to avoid duplication.
	sb.WriteString("# Design System\n\n")
	sb.WriteString("> All design tokens are stored in the YAML frontmatter above.\n")
	sb.WriteString("> The AI agent reads the frontmatter directly — this summary is for humans.\n\n")

	// Summary table
	sb.WriteString("## Overview\n\n")
	sb.WriteString("| Field | Value |\n")
	sb.WriteString("| ----- | ----- |\n")
	sb.WriteString(fmt.Sprintf("| Framework | %s |\n", ds.Framework.String()))

	if ds.Style != nil {
		if ds.Style.CSSFramework != "" {
			sb.WriteString(fmt.Sprintf("| CSS | %s (%s) |\n", ds.Style.CSSFramework, ds.Style.StylingApproach))
		}
		if ds.Style.Preprocessor != "" {
			sb.WriteString(fmt.Sprintf("| Preprocessor | %s |\n", ds.Style.Preprocessor))
		}
		if len(ds.Style.ThemeColors) > 0 {
			sb.WriteString(fmt.Sprintf("| Theme Colors | %d tokens |\n", len(ds.Style.ThemeColors)))
		}
		if len(ds.Style.FontFamilies) > 0 {
			sb.WriteString(fmt.Sprintf("| Fonts | %s |\n", strings.Join(ds.Style.FontFamilies, "; ")))
		}
		if len(ds.Style.CSSVariables) > 0 {
			sb.WriteString(fmt.Sprintf("| CSS Variables | %d |\n", len(ds.Style.CSSVariables)))
		}
	}
	if len(ds.ExternalLibs) > 0 {
		sb.WriteString(fmt.Sprintf("| UI Libraries | %s |\n", strings.Join(ds.ExternalLibs, ", ")))
	}
	if ds.Style != nil && len(ds.Style.ComponentLibs) > 0 {
		sb.WriteString(fmt.Sprintf("| Component Libs | %s |\n", strings.Join(ds.Style.ComponentLibs, ", ")))
	}
	if ds.AppStructure != nil {
		if ds.AppStructure.Router != "" {
			sb.WriteString(fmt.Sprintf("| Router | %s |\n", ds.AppStructure.Router))
		}
		sb.WriteString(fmt.Sprintf("| Layouts | %d |\n", len(ds.AppStructure.Layouts)))
		if len(ds.AppStructure.Components) > 0 {
			sb.WriteString(fmt.Sprintf("| Components | %d |\n", len(ds.AppStructure.Components)))
		}
		if len(ds.AppStructure.GlobalStyles) > 0 {
			sb.WriteString(fmt.Sprintf("| Global Styles | %d |\n", len(ds.AppStructure.GlobalStyles)))
		}
	}
	sb.WriteString("\n")

	// App structure tree
	if ds.AppStructure != nil {
		allPaths := append(ds.AppStructure.Layouts, ds.AppStructure.Components...)
		allPaths = append(allPaths, ds.AppStructure.GlobalStyles...)
		if len(allPaths) > 0 {
			sb.WriteString("## App Structure\n\n")
			sb.WriteString("```\n")
			sb.WriteString(buildDirTree(allPaths))
			sb.WriteString("```\n\n")
		}
	}

	// Color palette
	if ds.Style != nil && len(ds.Style.ThemeColors) > 0 {
		sb.WriteString("## Color Palette\n\n")
		colorNames := make([]string, 0, len(ds.Style.ThemeColors))
		for name := range ds.Style.ThemeColors {
			colorNames = append(colorNames, name)
		}
		sort.Strings(colorNames)
		for _, name := range colorNames {
			sb.WriteString(fmt.Sprintf("- `%s` → `%s`\n", name, ds.Style.ThemeColors[name]))
		}
		sb.WriteString("\n")
	}

	return os.WriteFile(path, []byte(sb.String()), 0600)
}

// buildDirTree renders a list of file paths as an ASCII directory tree.
func buildDirTree(paths []string) string {
	sort.Strings(paths)

	type node struct {
		name     string
		children []*node
	}

	root := &node{}
	findOrAdd := func(parent *node, name string) *node {
		for _, c := range parent.children {
			if c.name == name {
				return c
			}
		}
		child := &node{name: name}
		parent.children = append(parent.children, child)
		return child
	}

	for _, p := range paths {
		parts := strings.Split(filepath.ToSlash(p), "/")
		cur := root
		for _, part := range parts {
			cur = findOrAdd(cur, part)
		}
	}

	var sb strings.Builder
	var render func(n *node, prefix string, isLast bool)
	render = func(n *node, prefix string, isLast bool) {
		if n.name != "" {
			connector := "├── "
			if isLast {
				connector = "└── "
			}
			sb.WriteString(prefix + connector + n.name + "\n")
			if isLast {
				prefix += "    "
			} else {
				prefix += "│   "
			}
		}
		for i, c := range n.children {
			render(c, prefix, i == len(n.children)-1)
		}
	}

	for i, c := range root.children {
		render(c, "", i == len(root.children)-1)
	}

	return sb.String()
}

// parseFrontmatter extracts and parses the YAML frontmatter from markdown content.
func parseFrontmatter(content string) (*DesignSystem, error) {
	lines := strings.Split(content, "\n")
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
