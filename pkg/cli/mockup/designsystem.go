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
	manualStart    = "<!-- MANUAL -->"
	manualEnd      = "<!-- /MANUAL -->"
)

// LoadDesignSystem reads and parses a design_system.md file.
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

	// Parse markdown body for components
	ds.Components = parseComponentsFromMarkdown(content)
	ds.ManualEntries = parseManualEntries(content)

	return ds, nil
}

// WriteDesignSystem writes a DesignSystem to a design_system.md file.
func WriteDesignSystem(path string, ds *DesignSystem) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	ds.LastScanned = time.Now()

	var sb strings.Builder

	// Write YAML frontmatter
	sb.WriteString(frontmatterSep + "\n")

	frontmatter := struct {
		Version       int           `yaml:"version"`
		Framework     FrameworkType `yaml:"framework"`
		LastScanned   time.Time     `yaml:"last_scanned"`
		ComponentDirs []string      `yaml:"component_dirs"`
		ExternalLibs  []string      `yaml:"external_libs,omitempty"`
	}{
		Version:       ds.Version,
		Framework:     ds.Framework,
		LastScanned:   ds.LastScanned,
		ComponentDirs: ds.ComponentDirs,
		ExternalLibs:  ds.ExternalLibs,
	}

	yamlData, err := yaml.Marshal(frontmatter)
	if err != nil {
		return fmt.Errorf("failed to marshal frontmatter: %w", err)
	}
	sb.Write(yamlData)
	sb.WriteString(frontmatterSep + "\n\n")

	// Write markdown body
	sb.WriteString("# Design System Index\n\n")

	if len(ds.Components) > 0 {
		sb.WriteString("## Project Components\n\n")
		sb.WriteString("```\n")
		sb.WriteString(renderComponentTree(ds.Components))
		sb.WriteString("```\n\n")
	}

	// Write external libraries section
	if len(ds.ExternalLibs) > 0 {
		sb.WriteString("## External Libraries\n\n")
		for _, lib := range ds.ExternalLibs {
			sb.WriteString(fmt.Sprintf("- %s\n", lib))
		}
		sb.WriteString("\n")
	}

	// Write manual entries section
	if len(ds.ManualEntries) > 0 {
		sb.WriteString(manualStart + "\n\n")
		sb.WriteString("## Manual Entries\n\n")
		for _, c := range ds.ManualEntries {
			sb.WriteString(fmt.Sprintf("### %s\n", c.Name))
			if c.FilePath != "" {
				sb.WriteString(fmt.Sprintf("- **Path**: `%s`\n", c.FilePath))
			}
			if c.Description != "" {
				sb.WriteString(fmt.Sprintf("- **Description**: %s\n", c.Description))
			}
			sb.WriteString("\n")
		}
		sb.WriteString(manualEnd + "\n")
	}

	return os.WriteFile(path, []byte(sb.String()), 0600)
}

// MergeDesignSystem merges scan results into an existing design system,
// preserving manual entries. Returns (added, removed) counts.
func MergeDesignSystem(existing *DesignSystem, scanResult *ScanResult) (added, removed int) {
	// Build a set of existing component names
	existingNames := make(map[string]struct{}, len(existing.Components))
	for _, c := range existing.Components {
		existingNames[c.Name] = struct{}{}
	}

	// Build a set of scanned component names
	scannedNames := make(map[string]struct{}, len(scanResult.Components))
	for _, c := range scanResult.Components {
		scannedNames[c.Name] = struct{}{}
	}

	// Count added (in scan but not in existing)
	for name := range scannedNames {
		if _, exists := existingNames[name]; !exists {
			added++
		}
	}

	// Count removed (in existing but not in scan)
	for name := range existingNames {
		if _, exists := scannedNames[name]; !exists {
			removed++
		}
	}

	// Replace components with scan results, keep manual entries
	existing.Components = scanResult.Components
	existing.ComponentDirs = scanResult.ComponentDirs
	existing.ExternalLibs = scanResult.ExternalLibs
	existing.LastScanned = time.Now()

	return added, removed
}

// treeNode is used to build a directory tree for display.
type treeNode struct {
	name      string
	component *Component // nil for directories
	children  []*treeNode
}

// renderComponentTree builds an ASCII directory tree from components.
func renderComponentTree(components []Component) string {
	root := &treeNode{}

	for i := range components {
		c := &components[i]
		parts := strings.Split(filepath.ToSlash(c.FilePath), "/")
		node := root
		for j, part := range parts {
			isFile := j == len(parts)-1
			// Find or create child
			var child *treeNode
			for _, ch := range node.children {
				if ch.name == part {
					child = ch
					break
				}
			}
			if child == nil {
				child = &treeNode{name: part}
				if isFile {
					child.component = c
				}
				node.children = append(node.children, child)
			}
			node = child
		}
	}

	var sb strings.Builder
	// Render from each top-level entry
	for i, child := range root.children {
		isLast := i == len(root.children)-1
		renderTreeNode(&sb, child, "", isLast)
	}
	return sb.String()
}

func renderTreeNode(sb *strings.Builder, node *treeNode, prefix string, isLast bool) {
	connector := "├── "
	if isLast {
		connector = "└── "
	}

	sb.WriteString(prefix + connector + node.name)

	if node.component != nil {
		// File node — append component info
		sb.WriteString("  →  " + node.component.Name)
		if len(node.component.Props) > 0 {
			propNames := make([]string, 0, len(node.component.Props))
			for _, p := range node.component.Props {
				s := p.Name
				if p.Type != "" {
					s += ": " + p.Type
				}
				propNames = append(propNames, s)
			}
			sb.WriteString(" (" + strings.Join(propNames, ", ") + ")")
		}
	} else if len(node.children) > 0 {
		sb.WriteString("/")
	}
	sb.WriteString("\n")

	// Render children
	childPrefix := prefix
	if isLast {
		childPrefix += "    "
	} else {
		childPrefix += "│   "
	}
	for i, child := range node.children {
		childIsLast := i == len(node.children)-1
		renderTreeNode(sb, child, childPrefix, childIsLast)
	}
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

// treeLineRegex matches a tree line like:
//   "├── Button.tsx  →  Button (variant: string, onClick: func)"
var treeArrow = "→"

// parseComponentsFromMarkdown extracts components from the markdown body.
// Supports both tree format (new) and legacy ### format.
func parseComponentsFromMarkdown(content string) []Component {
	// Check if content has tree format (code block with → arrows)
	if strings.Contains(content, treeArrow) && strings.Contains(content, "```") {
		return parseComponentsFromTree(content)
	}
	return parseComponentsFromLegacy(content)
}

// parseComponentsFromTree parses the directory tree format.
func parseComponentsFromTree(content string) []Component {
	var components []Component
	inCodeBlock := false
	inManual := false

	// dirStack tracks the directory at each depth level
	type depthEntry struct {
		depth int
		name  string
	}
	var dirStack []depthEntry

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if trimmed == manualStart {
			inManual = true
			continue
		}
		if trimmed == manualEnd {
			inManual = false
			continue
		}
		if inManual {
			continue
		}

		if trimmed == "```" {
			inCodeBlock = !inCodeBlock
			if inCodeBlock {
				dirStack = nil
			}
			continue
		}
		if !inCodeBlock {
			continue
		}

		// Find the position of tree connector (├── or └──)
		connIdx := strings.Index(line, "├── ")
		if connIdx < 0 {
			connIdx = strings.Index(line, "└── ")
		}
		if connIdx < 0 {
			continue
		}

		// Depth is determined by connector position (each level = 4 chars)
		depth := connIdx / 4

		// Extract the entry text after the connector
		entry := line[connIdx+len("├── "):]
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}

		// Trim dirStack to current depth
		for len(dirStack) > 0 && dirStack[len(dirStack)-1].depth >= depth {
			dirStack = dirStack[:len(dirStack)-1]
		}

		// Check if it's a component line (has →)
		if strings.Contains(entry, treeArrow) {
			parts := strings.SplitN(entry, treeArrow, 2)
			filename := strings.TrimSpace(parts[0])
			compInfo := strings.TrimSpace(parts[1])

			// Build file path from dir stack + filename
			var pathParts []string
			for _, de := range dirStack {
				pathParts = append(pathParts, de.name)
			}
			pathParts = append(pathParts, filename)
			filePath := strings.Join(pathParts, "/")

			comp := Component{
				FilePath: filePath,
			}

			// Parse "Name (prop1: type1, prop2: type2)"
			if idx := strings.Index(compInfo, "("); idx > 0 && strings.HasSuffix(compInfo, ")") {
				comp.Name = strings.TrimSpace(compInfo[:idx])
				propsStr := compInfo[idx+1 : len(compInfo)-1]
				for _, p := range strings.Split(propsStr, ",") {
					p = strings.TrimSpace(p)
					if p == "" {
						continue
					}
					propParts := strings.SplitN(p, ":", 2)
					pi := PropInfo{Name: strings.TrimSpace(propParts[0])}
					if len(propParts) > 1 {
						pi.Type = strings.TrimSpace(propParts[1])
					}
					comp.Props = append(comp.Props, pi)
				}
			} else {
				comp.Name = compInfo
			}

			components = append(components, comp)
		} else {
			// Directory entry (ends with /)
			dirName := strings.TrimSuffix(entry, "/")
			dirStack = append(dirStack, depthEntry{depth: depth, name: dirName})
		}
	}

	return components
}

// parseComponentsFromLegacy parses the old ### heading format for backwards compatibility.
func parseComponentsFromLegacy(content string) []Component {
	var components []Component
	inManual := false

	lines := strings.Split(content, "\n")
	var currentComp *Component

	for i := 0; i < len(lines); i++ {
		line := lines[i]

		if strings.TrimSpace(line) == manualStart {
			inManual = true
			continue
		}
		if strings.TrimSpace(line) == manualEnd {
			inManual = false
			continue
		}
		if inManual {
			continue
		}

		if strings.HasPrefix(line, "### ") {
			if currentComp != nil {
				components = append(components, *currentComp)
			}
			name := strings.TrimPrefix(line, "### ")
			name = strings.TrimSpace(name)
			currentComp = &Component{Name: name}
			continue
		}

		if currentComp == nil {
			continue
		}

		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "- **Path**:") {
			val := extractMarkdownValue(trimmed, "- **Path**:")
			currentComp.FilePath = strings.Trim(val, "`")
		}
		if strings.HasPrefix(trimmed, "- **Description**:") {
			currentComp.Description = extractMarkdownValue(trimmed, "- **Description**:")
		}
		if strings.HasPrefix(trimmed, "- **Props**:") {
			propsStr := extractMarkdownValue(trimmed, "- **Props**:")
			propNames := strings.Split(propsStr, ",")
			for _, p := range propNames {
				p = strings.TrimSpace(p)
				p = strings.Trim(p, "`")
				if p != "" {
					currentComp.Props = append(currentComp.Props, PropInfo{Name: p})
				}
			}
		}
	}

	if currentComp != nil && currentComp.FilePath != "" {
		components = append(components, *currentComp)
	}

	return components
}

// parseManualEntries extracts manually-added components from between MANUAL markers.
func parseManualEntries(content string) []Component {
	startIdx := strings.Index(content, manualStart)
	endIdx := strings.Index(content, manualEnd)
	if startIdx < 0 || endIdx < 0 || endIdx <= startIdx {
		return nil
	}

	manualBlock := content[startIdx+len(manualStart) : endIdx]
	var components []Component
	var currentComp *Component

	lines := strings.Split(manualBlock, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "### ") {
			if currentComp != nil {
				components = append(components, *currentComp)
			}
			name := strings.TrimPrefix(line, "### ")
			name = strings.TrimSpace(name)
			currentComp = &Component{Name: name}
			continue
		}

		if currentComp == nil {
			continue
		}

		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "- **Path**:") {
			currentComp.FilePath = strings.Trim(extractMarkdownValue(trimmed, "- **Path**:"), "`")
		}
		if strings.HasPrefix(trimmed, "- **Description**:") {
			currentComp.Description = extractMarkdownValue(trimmed, "- **Description**:")
		}
	}

	if currentComp != nil {
		components = append(components, *currentComp)
	}

	return components
}

// extractMarkdownValue extracts the value after a markdown label prefix.
func extractMarkdownValue(line, prefix string) string {
	return strings.TrimSpace(strings.TrimPrefix(line, prefix))
}
