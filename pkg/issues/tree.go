package issues

import (
	"fmt"
	"strings"
)

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorGray   = "\033[90m"
)

// TreeRenderOptions configures the tree output format
type TreeRenderOptions struct {
	MaxDepth    int  // Maximum depth to render (default: 10)
	ShowStatus  bool // Include status indicator (default: true)
	TitleWidth  int  // Max title width before truncation (default: 40)
	ShowSpec    bool // Show spec context for cross-spec trees (default: false)
	ShowType    bool // Show issue type (default: true)
	ShowPriority bool // Show priority (default: true)
	Color       bool // Use colors (default: true)
}

// DefaultTreeRenderOptions returns the default options
func DefaultTreeRenderOptions() TreeRenderOptions {
	return TreeRenderOptions{
		MaxDepth:     10,
		ShowStatus:   true,
		TitleWidth:   40,
		ShowSpec:     false,
		ShowType:     true,
		ShowPriority: true,
		Color:        true,
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

// colorize applies color if enabled
func (r *TreeRenderer) colorize(text, color string) string {
	if !r.options.Color {
		return text
	}
	return color + text + colorReset
}

// formatPriority returns a colored priority indicator
func (r *TreeRenderer) formatPriority(priority int) string {
	var indicator string
	var color string

	switch priority {
	case 0:
		indicator = "P0"
		color = colorRed
	case 1:
		indicator = "P1"
		color = colorYellow
	case 2:
		indicator = "P2"
		color = colorGreen
	case 3:
		indicator = "P3"
		color = colorCyan
	default:
		indicator = fmt.Sprintf("P%d", priority)
		color = colorGray
	}

	return r.colorize("["+indicator+"]", color)
}

// formatType returns a colored type indicator
func (r *TreeRenderer) formatType(issueType IssueType) string {
	var indicator string
	var color string

	switch issueType {
	case TypeEpic:
		indicator = "E"
		color = colorPurple
	case TypeFeature:
		indicator = "F"
		color = colorGreen
	case TypeTask:
		indicator = "T"
		color = colorBlue
	case TypeBug:
		indicator = "B"
		color = colorRed
	default:
		indicator = "?"
		color = colorGray
	}

	return r.colorize("["+indicator+"]", color)
}

// formatStatus returns a colored status indicator
func (r *TreeRenderer) formatStatus(status IssueStatus) string {
	var indicator string
	var color string

	switch status {
	case StatusOpen:
		indicator = "○"
		color = colorGreen
	case StatusInProgress:
		indicator = "◐"
		color = colorYellow
	case StatusClosed:
		indicator = "●"
		color = colorGray
	default:
		indicator = "?"
		color = colorGray
	}

	return r.colorize(indicator, color)
}

// formatIssue formats an issue for display in the tree
func (r *TreeRenderer) formatIssue(issue Issue) string {
	var sb strings.Builder

	// Type indicator
	if r.options.ShowType {
		sb.WriteString(r.formatType(issue.IssueType))
	}

	// Priority
	if r.options.ShowPriority {
		sb.WriteString(r.formatPriority(issue.Priority))
	}

	// ID (bolded)
	sb.WriteString(" ")
	sb.WriteString(r.colorize(issue.ID, "\033[1m")) // bold

	// Title
	sb.WriteString(" ")
	sb.WriteString(truncate(issue.Title, r.options.TitleWidth))

	// Status indicator
	if r.options.ShowStatus {
		sb.WriteString(" ")
		sb.WriteString(r.formatStatus(issue.Status))
	}

	// Spec context
	if r.options.ShowSpec && issue.SpecContext != "" {
		sb.WriteString(" (")
		sb.WriteString(issue.SpecContext)
		sb.WriteString(")")
	}

	return sb.String()
}

// FormatIssueSimple is a public method to format an issue for display
func (r *TreeRenderer) FormatIssueSimple(issue Issue) string {
	return r.formatIssue(issue)
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
		if depth == 0 {
			// Already has prefix from RenderWithRoot, just add cycle marker
			sb.WriteString(" ⚠ (cycle)\n")
		} else {
			if isLast {
				sb.WriteString("└── ")
			} else {
				sb.WriteString("├── ")
			}
			sb.WriteString(r.formatIssue(tree.Issue))
			sb.WriteString(" ⚠ (cycle)\n")
		}
		return sb.String()
	}

	// Mark as visited
	visited[tree.Issue.ID] = true
	defer func() { delete(visited, tree.Issue.ID) }()

	// Render current node
	if depth == 0 {
		// Prefix already contains the tree marker from RenderWithRoot
		sb.WriteString(prefix)
		sb.WriteString(r.formatIssue(tree.Issue))
	} else {
		sb.WriteString(prefix)
		if isLast {
			sb.WriteString("└── ")
		} else {
			sb.WriteString("├── ")
		}
		sb.WriteString(r.formatIssue(tree.Issue))
	}
	sb.WriteString("\n")

	// Render children (Blocks)
	if len(tree.Blocks) > 0 {
		var newPrefix string
		if depth == 0 {
			// First level children - determine continuation prefix
			if isLast {
				newPrefix = "    "
			} else {
				newPrefix = "│   "
			}
		} else {
			if isLast {
				newPrefix = prefix + "    "
			} else {
				newPrefix = prefix + "│   "
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

	// Render each tree as a child of root (depth 0, with prefix from root)
	for i, tree := range trees {
		isLast := i == len(trees)-1
		var prefix string
		if isLast {
			prefix = "└── "
		} else {
			prefix = "├── "
		}
		sb.WriteString(r.renderTree(tree, prefix, isLast, 0, make(map[string]bool)))
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
