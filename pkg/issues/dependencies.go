package issues

import (
	"errors"
	"fmt"
)

// Dependency-related errors
var (
	ErrCyclicDependency   = errors.New("would create a circular dependency")
	ErrSelfDependency     = errors.New("cannot create dependency on self")
	ErrDependencyNotFound = errors.New("dependency target issue not found")
)

// LinkType represents the type of relationship between issues
type LinkType string

const (
	LinkBlocks  LinkType = "blocks"  // A blocks B (A must complete before B)
	LinkRelated LinkType = "related" // A and B are related
)

// IsValidLinkType checks if a link type is valid
func IsValidLinkType(t LinkType) bool {
	switch t {
	case LinkBlocks, LinkRelated:
		return true
	default:
		return false
	}
}

// DependencyTree represents the dependency tree for an issue
type DependencyTree struct {
	Issue     Issue
	BlockedBy []*DependencyTree
	Blocks    []*DependencyTree
	Children  []*DependencyTree // Parent-child hierarchy
}

// AddDependency creates a dependency link between two issues
func (s *Store) AddDependency(fromID, toID string, linkType LinkType) error {
	return s.WithLock(func() error {
		// Validate inputs
		if fromID == toID {
			return ErrSelfDependency
		}

		if !IsValidLinkType(linkType) {
			return fmt.Errorf("invalid link type: %s", linkType)
		}

		// Get all issues
		issues, err := s.readAllUnlocked()
		if err != nil {
			return err
		}

		// Find both issues
		var fromIssue, toIssue *Issue
		var fromIdx, toIdx int
		for i, issue := range issues {
			if issue.ID == fromID {
				fromIssue = issue
				fromIdx = i
			}
			if issue.ID == toID {
				toIssue = issue
				toIdx = i
			}
		}

		if fromIssue == nil {
			return fmt.Errorf("issue %s not found", fromID)
		}
		if toIssue == nil {
			return fmt.Errorf("issue %s not found", toID)
		}

		// For blocks: fromID blocks toID means toID is blocked by fromID
		if linkType == LinkBlocks {
			// Check for cycles: if we're adding "A blocks B", check if B already blocks A
			if contains(toIssue.Blocks, fromID) || contains(fromIssue.BlockedBy, toID) {
				return fmt.Errorf("%w: %s -> %s -> %s", ErrCyclicDependency, toID, fromID, toID)
			}

			// Check for deeper cycles using DFS
			if s.wouldCreateCycle(fromID, toID, issues) {
				return fmt.Errorf("%w: adding this dependency would create a cycle", ErrCyclicDependency)
			}

			// Add to fromIssue.Blocks
			if !contains(fromIssue.Blocks, toID) {
				fromIssue.Blocks = append(fromIssue.Blocks, toID)
			}

			// Add to toIssue.BlockedBy (bidirectional)
			if !contains(toIssue.BlockedBy, fromID) {
				toIssue.BlockedBy = append(toIssue.BlockedBy, fromID)
			}
		} else if linkType == LinkRelated {
			// For related, we could add a separate Related field, but for now
			// we just add to blocks/blocked_by with a note that it's a soft link
			// For simplicity, related links are stored in both directions
			if !contains(fromIssue.Blocks, toID) {
				fromIssue.Blocks = append(fromIssue.Blocks, toID)
			}
			if !contains(toIssue.Blocks, fromID) {
				toIssue.Blocks = append(toIssue.Blocks, fromID)
			}
		}

		// Update timestamps
		fromIssue.UpdatedAt = NowFunc()
		toIssue.UpdatedAt = NowFunc()

		// Update in slice
		issues[fromIdx] = fromIssue
		issues[toIdx] = toIssue

		// Write back
		return s.writeAllUnlocked(issues)
	})
}

// RemoveDependency removes a dependency link between two issues
func (s *Store) RemoveDependency(fromID, toID string, linkType LinkType) error {
	return s.WithLock(func() error {
		// Get all issues
		issues, err := s.readAllUnlocked()
		if err != nil {
			return err
		}

		// Find both issues
		var fromIssue, toIssue *Issue
		var fromIdx, toIdx int
		for i, issue := range issues {
			if issue.ID == fromID {
				fromIssue = issue
				fromIdx = i
			}
			if issue.ID == toID {
				toIssue = issue
				toIdx = i
			}
		}

		if fromIssue == nil {
			return fmt.Errorf("issue %s not found", fromID)
		}
		if toIssue == nil {
			return fmt.Errorf("issue %s not found", toID)
		}

		// Remove from fromIssue.Blocks
		fromIssue.Blocks = removeFromSlice(fromIssue.Blocks, toID)

		// Remove from toIssue.BlockedBy
		toIssue.BlockedBy = removeFromSlice(toIssue.BlockedBy, fromID)

		// Update timestamps
		fromIssue.UpdatedAt = NowFunc()
		toIssue.UpdatedAt = NowFunc()

		// Update in slice
		issues[fromIdx] = fromIssue
		issues[toIdx] = toIssue

		// Write back
		return s.writeAllUnlocked(issues)
	})
}

// wouldCreateCycle checks if adding a dependency would create a cycle using DFS
func (s *Store) wouldCreateCycle(fromID, toID string, issues []*Issue) bool {
	// Build adjacency map
	adj := make(map[string][]string)
	for _, issue := range issues {
		adj[issue.ID] = append(adj[issue.ID], issue.Blocks...)
	}

	// Add the proposed edge
	adj[fromID] = append(adj[fromID], toID)

	// Check for cycle using DFS from toID
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var dfs func(node string) bool
	dfs = func(node string) bool {
		visited[node] = true
		recStack[node] = true

		for _, neighbor := range adj[node] {
			if !visited[neighbor] {
				if dfs(neighbor) {
					return true
				}
			} else if recStack[neighbor] {
				return true
			}
		}

		recStack[node] = false
		return false
	}

	// Start DFS from toID to see if we can reach fromID
	return dfs(toID)
}

// DetectCycles checks for circular dependencies in the entire issue set
func (s *Store) DetectCycles() ([][]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	issues, err := s.readAllUnlocked()
	if err != nil {
		return nil, err
	}

	// Build adjacency map
	adj := make(map[string][]string)
	for _, issue := range issues {
		adj[issue.ID] = append(adj[issue.ID], issue.Blocks...)
	}

	var cycles [][]string
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	path := []string{}

	var dfs func(node string)
	dfs = func(node string) {
		visited[node] = true
		recStack[node] = true
		path = append(path, node)

		for _, neighbor := range adj[node] {
			if !visited[neighbor] {
				dfs(neighbor)
			} else if recStack[neighbor] {
				// Found a cycle
				cycleStart := -1
				for i, n := range path {
					if n == neighbor {
						cycleStart = i
						break
					}
				}
				if cycleStart >= 0 {
					cycle := make([]string, len(path)-cycleStart)
					copy(cycle, path[cycleStart:])
					cycle = append(cycle, neighbor)
					cycles = append(cycles, cycle)
				}
			}
		}

		path = path[:len(path)-1]
		recStack[node] = false
	}

	for _, issue := range issues {
		if !visited[issue.ID] {
			dfs(issue.ID)
		}
	}

	return cycles, nil
}

// GetDependencyTree returns the full dependency tree for an issue
func (s *Store) GetDependencyTree(id string) (*DependencyTree, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	issues, err := s.readAllUnlocked()
	if err != nil {
		return nil, err
	}

	// Build issue map
	issueMap := make(map[string]*Issue)
	for _, issue := range issues {
		issueMap[issue.ID] = issue
	}

	// Find the issue
	issue, ok := issueMap[id]
	if !ok {
		return nil, ErrIssueNotFound
	}

	// Build tree recursively
	tree := &DependencyTree{Issue: *issue}
	tree.BlockedBy = s.buildDependencySubtree(issue.BlockedBy, issueMap, make(map[string]bool))
	tree.Blocks = s.buildDependencySubtree(issue.Blocks, issueMap, make(map[string]bool))

	return tree, nil
}

func (s *Store) buildDependencySubtree(ids []string, issueMap map[string]*Issue, visited map[string]bool) []*DependencyTree {
	var trees []*DependencyTree
	for _, id := range ids {
		if visited[id] {
			continue // Avoid infinite recursion
		}
		visited[id] = true

		issue, ok := issueMap[id]
		if !ok {
			continue
		}

		tree := &DependencyTree{Issue: *issue}
		tree.BlockedBy = s.buildDependencySubtree(issue.BlockedBy, issueMap, visited)
		tree.Blocks = s.buildDependencySubtree(issue.Blocks, issueMap, visited)
		trees = append(trees, tree)
	}
	return trees
}

// GetBlockedIssues returns all issues that are currently blocked
func (s *Store) GetBlockedIssues() ([]Issue, error) {
	filter := ListFilter{Blocked: true}
	return s.List(filter)
}

// GetHierarchyForest returns a forest of trees based on parent-child relationships.
// Root nodes are issues without a parent. Each tree includes all descendants.
func (s *Store) GetHierarchyForest() ([]*DependencyTree, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	issues, err := s.readAllUnlocked()
	if err != nil {
		return nil, err
	}

	// Build issue map
	issueMap := make(map[string]*Issue)
	for _, issue := range issues {
		issueMap[issue.ID] = issue
	}

	// Find root issues (no parent)
	var roots []*DependencyTree
	hasParent := make(map[string]bool)

	for _, issue := range issues {
		if issue.ParentID != nil && *issue.ParentID != "" {
			hasParent[issue.ID] = true
		}
	}

	// Build trees for root issues
	visited := make(map[string]bool)
	for _, issue := range issues {
		if !hasParent[issue.ID] {
			tree := s.buildHierarchyTree(issue.ID, issueMap, visited)
			if tree != nil {
				roots = append(roots, tree)
			}
		}
	}

	return roots, nil
}

// buildHierarchyTree recursively builds a hierarchy tree for an issue
func (s *Store) buildHierarchyTree(id string, issueMap map[string]*Issue, visited map[string]bool) *DependencyTree {
	if visited[id] {
		return nil // Avoid cycles
	}
	visited[id] = true

	issue, ok := issueMap[id]
	if !ok {
		return nil
	}

	tree := &DependencyTree{Issue: *issue}

	// Find children (issues that have this issue as parent)
	for _, potentialChild := range issueMap {
		if potentialChild.ParentID != nil && *potentialChild.ParentID == id {
			childTree := s.buildHierarchyTree(potentialChild.ID, issueMap, visited)
			if childTree != nil {
				tree.Children = append(tree.Children, childTree)
			}
		}
	}

	return tree
}

// Helper function to remove a string from a slice
func removeFromSlice(slice []string, item string) []string {
	var result []string
	for _, s := range slice {
		if s != item {
			result = append(result, s)
		}
	}
	return result
}
