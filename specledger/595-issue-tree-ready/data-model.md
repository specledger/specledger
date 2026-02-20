# Data Model: Issue Tree View and Ready Command

**Feature**: 595-issue-tree-ready
**Date**: 2026-02-20

## Existing Entities (No Changes)

### Issue

The Issue entity already exists in `pkg/issues/issue.go` with all needed fields:

| Field | Type | Description | Usage for This Feature |
|-------|------|-------------|------------------------|
| ID | string | Unique identifier (SL-xxxxxx) | Display in tree |
| Title | string | Issue title | Display in tree (truncated) |
| Status | IssueStatus | open/in_progress/closed | Display in tree, ready state |
| Priority | int | 0-5 priority | Sorting in ready list |
| BlockedBy | []string | Issue IDs blocking this | Ready state computation |
| Blocks | []string | Issue IDs this blocks | Tree structure |
| SpecContext | string | Spec identifier | Cross-spec queries |

### DependencyTree

Already exists in `pkg/issues/dependencies.go`:

```go
type DependencyTree struct {
    Issue     Issue
    BlockedBy []*DependencyTree
    Blocks    []*DependencyTree
}
```

Used by existing `GetDependencyTree()` method.

### ListFilter

Already exists in `pkg/issues/issue.go`:

```go
type ListFilter struct {
    Status      *IssueStatus
    IssueType   *IssueType
    Priority    *int
    Labels      []string
    SpecContext string
    All         bool
    Blocked     bool  // Existing - shows only blocked issues
}
```

## New/Modified Entities

### ListFilter Extension

Add `Ready` field to existing filter:

```go
type ListFilter struct {
    // ... existing fields ...
    Ready       bool   // NEW - show only ready issues (not blocked)
}
```

### TreeRenderOptions (New)

Configuration for tree rendering:

```go
type TreeRenderOptions struct {
    MaxDepth     int    // Maximum depth to render (default: 10)
    ShowStatus   bool   // Include status indicator (default: true)
    TitleWidth   int    // Max title width before truncation (default: 40)
    ShowSpec     bool   // Show spec context for cross-spec trees (default: false)
}
```

### TreeRenderer (New)

Handles tree output formatting:

```go
type TreeRenderer struct {
    options TreeRenderOptions
}

func (r *TreeRenderer) Render(tree *DependencyTree) string
func (r *TreeRenderer) RenderForest(trees []*DependencyTree) string
```

### ReadyIssue (Virtual)

Computed view of ready issues (not persisted):

```go
// ReadyIssue represents an issue ready for work
// This is a virtual/computed entity, not stored
type ReadyIssue struct {
    Issue        Issue
    BlockedBy    []Blocker // Empty if truly ready
}

type Blocker struct {
    ID      string
    Title   string
    Status  IssueStatus
}
```

## Computed Properties

### IsReady() Method

Add to Issue entity:

```go
func (i *Issue) IsReady(allIssues map[string]*Issue) bool {
    // Not ready if closed
    if i.Status == StatusClosed {
        return false
    }

    // Ready if no blockers
    if len(i.BlockedBy) == 0 {
        return true
    }

    // Ready if all blockers are closed
    for _, blockerID := range i.BlockedBy {
        blocker, exists := allIssues[blockerID]
        if !exists || blocker.Status != StatusClosed {
            return false
        }
    }
    return true
}
```

### GetBlockers() Method

Add to Issue entity:

```go
func (i *Issue) GetBlockers(allIssues map[string]*Issue) []Blocker {
    var blockers []Blocker
    for _, blockerID := range i.BlockedBy {
        if blocker, exists := allIssues[blockerID]; exists {
            blockers = append(blockers, Blocker{
                ID:     blocker.ID,
                Title:  blocker.Title,
                Status: blocker.Status,
            })
        }
    }
    return blockers
}
```

## Store Extensions

### ListReady() Method

Add to Store:

```go
func (s *Store) ListReady(filter ListFilter) ([]ReadyIssue, error) {
    // Load all issues
    issues, err := s.readAllUnlocked()
    if err != nil {
        return nil, err
    }

    // Build lookup map
    issueMap := make(map[string]*Issue)
    for _, issue := range issues {
        issueMap[issue.ID] = issue
    }

    // Filter ready issues
    var ready []ReadyIssue
    for _, issue := range issues {
        if issue.IsReady(issueMap) {
            ready = append(ready, ReadyIssue{
                Issue:     *issue,
                BlockedBy: []Blocker{}, // Empty for ready issues
            })
        }
    }

    return ready, nil
}
```

### GetReadyIssuesAcrossSpecs() Function

Add to store package:

```go
func GetReadyIssuesAcrossSpecs(basePath string, filter ListFilter) ([]ReadyIssue, error) {
    // Similar to ListAllSpecs but filters for ready state
    specs, err := listSpecDirs(basePath)
    if err != nil {
        return nil, err
    }

    var allReady []ReadyIssue
    for _, spec := range specs {
        store, err := NewStore(StoreOptions{
            BasePath:    basePath,
            SpecContext: spec,
        })
        if err != nil {
            continue
        }

        ready, err := store.ListReady(filter)
        if err != nil {
            continue
        }

        allReady = append(allReady, ready...)
    }

    return allReady, nil
}
```

## CLI Output Formats

### Tree Output

```
595-issue-tree-ready (5 issues)
├── SL-abc123 [open] Implement tree renderer
│   ├── SL-def456 [open] Add ASCII tree characters
│   └── SL-ghi789 [in_progress] Add cycle detection
└── SL-jkl012 [closed] Design tree structure ⚠ cycle
    └── SL-abc123 [open] Implement tree renderer
```

### Ready Output (Table)

```
ID         TITLE                          STATUS       PRIORITY
SL-abc123  Implement tree renderer        open         1
SL-ghi789  Add cycle detection            in_progress  2
```

### Ready Output (JSON)

```json
[
  {
    "id": "SL-abc123",
    "title": "Implement tree renderer",
    "status": "open",
    "priority": 1,
    "spec_context": "595-issue-tree-ready"
  }
]
```

### Blocked Message Output

```
No ready issues found.

Blocked issues:
  SL-xxx789 "Setup database" blocked by:
    - SL-abc123 "Create connection" (open)
```

## State Transitions

### Ready State Changes

```
Issue Created (open, no blockers) → READY

Blocker Added:
  READY → NOT READY (if blocker not closed)

Blocker Closed:
  NOT READY → READY (if all blockers now closed)

Issue Closed:
  READY/NOT READY → Not considered for ready list
```

### Tree Display States

```
No Dependencies → Flat list under root
Dependencies Exist → Hierarchical tree
Cycle Detected → Warning + partial tree
Broken Reference → ⚠ indicator on node
```
