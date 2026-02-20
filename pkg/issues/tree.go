package issues

import (
	"fmt"
	"strings"
)

// TreeRenderOptions configures the tree output format
type TreeRenderOptions struct {
	MaxDepth   int  // Maximum depth to render (default: 10)
	ShowStatus bool // Include status indicator (default: true)
	TitleWidth int  // Max title width before truncation (default: 40)
	ShowSpec   bool // Show spec context for cross-spec trees (default: false)
}

// DefaultTreeRenderOptions returns the default options
func DefaultTreeRenderOptions() TreeRenderOptions {
	return TreeRenderOptions{
		MaxDepth:   10,
		ShowStatus: true,
		TitleWidth: 40,
		ShowSpec:   false,
	}
}

// TreeRenderer handles tree output formatting
type TreeRenderer struct {
	options TreeRenderOptions
}

// NewTreeRenderer creates a new tree renderer with the given options
func NewTreeRenderer(opts TreeRenderOptions) *TreeRenderer {
	return &TreeRenderer{options: opts}
}

// truncate truncates a string to the given length
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// formatIssue formats an issue for display in the tree
func (r *TreeRenderer) formatIssue(issue Issue) string {
	var sb strings.Builder

	sb.WriteString(issue.ID)

	if r.options.ShowStatus {
		sb.WriteString(" [")
		sb.WriteString(string(issue.Status))
		sb.WriteString("]")
	}

	sb.WriteString(" ")
	sb.WriteString(truncate(issue.Title, r.options.TitleWidth))

	if r.options.ShowSpec && issue.SpecContext != "" {
		sb.WriteString(" (")
		sb.WriteString(issue.SpecContext)
		sb.WriteString(")")
	}

	return sb.String()
}

// Render renders a single dependency tree
func (r *TreeRenderer) Render(tree *DependencyTree) string {
	return r.renderTree(tree, "", true, 0, make(map[string]bool))
}

// RenderForest renders multiple dependency trees
func (r *TreeRenderer) RenderForest(trees []*DependencyTree) string {
	var sb strings.Builder

	for i, tree := range trees {
		isLast := i == len(trees)-1
		sb.WriteString(r.renderTree(tree, "", isLast, 0, make(map[string]bool)))
		if !isLast {
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// renderTree recursively renders a tree with proper indentation
func (r *TreeRenderer) renderTree(tree *DependencyTree, prefix string, isLast bool, depth int, visited map[string]bool) string {
	if depth > r.options.MaxDepth {
		return prefix + "...\n"
	}

	var sb strings.Builder

	// Check for cycles
	if visited[tree.Issue.ID] {
		sb.WriteString(prefix)
		if isLast {
			sb.WriteString("└── ")
		} else {
			sb.WriteString("├── ")
		}
		sb.WriteString(r.formatIssue(tree.Issue))
		sb.WriteString(" ⚠ (cycle)\n")
		return sb.String()
	}

	// Mark as visited
	visited[tree.Issue.ID] = true
	defer func() { delete(visited, tree.Issue.ID) }()

	// Render current node
	sb.WriteString(prefix)
	if isLast {
		sb.WriteString("└── ")
	} else if depth > 0 {
		sb.WriteString("├── ")
	}
	sb.WriteString(r.formatIssue(tree.Issue))
	sb.WriteString("\n")

	// Render children (Blocks)
	if len(tree.Blocks) > 0 {
		newPrefix := prefix
		if depth > 0 {
			if isLast {
				newPrefix += "    "
			} else {
				newPrefix += "│   "
			}
		}

		for i, child := range tree.Blocks {
			childIsLast := i == len(tree.Blocks)-1
			sb.WriteString(r.renderTree(child, newPrefix, childIsLast, depth+1, visited))
		}
	}

	return sb.String()
}

// RenderWithRoot renders a tree with a root label
func (r *TreeRenderer) RenderWithRoot(rootLabel string, trees []*DependencyTree, totalIssues int) string {
	var sb strings.Builder

	// Render root
	sb.WriteString(rootLabel)
	if totalIssues > 0 {
		sb.WriteString(fmt.Sprintf(" (%d issues)", totalIssues))
	}
	sb.WriteString("\n")

	// Render each tree as a child of root
	for i, tree := range trees {
		isLast := i == len(trees)-1
		childOutput := r.renderTree(tree, "", isLast, 1, make(map[string]bool))

		// Add proper prefix to each line
		lines := strings.Split(strings.TrimRight(childOutput, "\n"), "\n")
		for j, line := range lines {
			if j == 0 {
				if isLast {
					sb.WriteString("└── ")
				} else {
					sb.WriteString("├── ")
				}
			} else {
				if isLast {
					sb.WriteString("    ")
				} else {
					sb.WriteString("│   ")
				}
			}
			sb.WriteString(line)
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// DetectCycles detects cycles in the dependency graph and returns them
func DetectCycles(trees []*DependencyTree) [][]string {
	var cycles [][]string
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var dfs func(tree *DependencyTree, path []string)
	dfs = func(tree *DependencyTree, path []string) {
		id := tree.Issue.ID

		if recStack[id] {
			// Found a cycle
			cycleStart := -1
			for i, p := range path {
				if p == id {
					cycleStart = i
					break
				}
			}
			if cycleStart >= 0 {
				cycle := make([]string, len(path)-cycleStart)
				copy(cycle, path[cycleStart:])
				cycle = append(cycle, id)
				cycles = append(cycles, cycle)
			}
			return
		}

		if visited[id] {
			return
		}

		visited[id] = true
		recStack[id] = true
		path = append(path, id)

		for _, child := range tree.Blocks {
			dfs(child, path)
		}

		recStack[id] = false
	}

	for _, tree := range trees {
		visited = make(map[string]bool)
		recStack = make(map[string]bool)
		dfs(tree, []string{})
	}

	return cycles
}

// FormatCycleWarning formats a cycle warning message
func FormatCycleWarning(cycles [][]string) string {
	if len(cycles) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("⚠ Warning: Cyclic dependencies detected\n")
	for _, cycle := range cycles {
		sb.WriteString("  Cycle: ")
		for i, id := range cycle {
			if i > 0 {
				sb.WriteString(" → ")
			}
			sb.WriteString(id)
		}
		sb.WriteString("\n")
	}
	sb.WriteString("\n")
	return sb.String()
}
