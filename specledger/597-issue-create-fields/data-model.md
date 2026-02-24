# Data Model: Issue Create Fields Enhancement

**Feature**: 597-issue-create-fields
**Date**: 2026-02-24

## Overview

This feature exposes existing fields in the Issue model via CLI and adds a new `ParentID` field for parent-child relationships.

## Entities

### Issue (existing + ParentID addition)

Location: `pkg/issues/issue.go`

```go
type Issue struct {
    // Required fields
    ID          string      `json:"id"`
    Title       string      `json:"title"`
    Description string      `json:"description,omitempty"`
    Status      IssueStatus `json:"status"`
    Priority    int         `json:"priority"`      // 0=highest, 5=lowest
    IssueType   IssueType   `json:"issue_type"`
    SpecContext string      `json:"spec_context"`
    CreatedAt   time.Time   `json:"created_at"`
    UpdatedAt   time.Time   `json:"updated_at"`

    // Optional fields (being exposed via CLI)
    ClosedAt           *time.Time        `json:"closed_at,omitempty"`
    DefinitionOfDone   *DefinitionOfDone `json:"definition_of_done,omitempty"`
    BlockedBy          []string          `json:"blocked_by,omitempty"`  // Issue IDs
    Blocks             []string          `json:"blocks,omitempty"`      // Issue IDs
    Labels             []string          `json:"labels,omitempty"`
    Assignee           string            `json:"assignee,omitempty"`
    Notes              string            `json:"notes,omitempty"`
    Design             string            `json:"design,omitempty"`
    AcceptanceCriteria string            `json:"acceptance_criteria,omitempty"`

    // NEW: Parent-child relationship
    ParentID          *string           `json:"parentId,omitempty"`    // ID of parent issue

    // Migration metadata
    BeadsMigration *BeadsMigration `json:"beads_migration,omitempty"`
}
```

### DefinitionOfDone (existing - no changes)

```go
type DefinitionOfDone struct {
    Items []ChecklistItem `json:"items"`
}

type ChecklistItem struct {
    Item       string     `json:"item"`
    Checked    bool       `json:"checked"`
    VerifiedAt *time.Time `json:"verified_at,omitempty"`
}
```

### IssueUpdate (existing + ParentID addition)

```go
type IssueUpdate struct {
    Title              *string
    Description        *string
    Status             *IssueStatus
    Priority           *int
    IssueType          *IssueType
    Assignee           *string
    Notes              *string
    Design             *string
    AcceptanceCriteria *string
    Labels             *[]string
    AddLabels          []string
    RemoveLabels       []string
    BlockedBy          *[]string
    Blocks             *[]string
    DefinitionOfDone   *DefinitionOfDone  // Replace entire DoD
    CheckDoDItem       string             // Item to mark as checked
    UncheckDoDItem     string             // Item to mark as unchecked
    ParentID           *string            // NEW: Set or clear parent
}
```

## Field to Flag Mapping

### Issue Create

| Field | CLI Flag | Type | Notes |
|-------|----------|------|-------|
| AcceptanceCriteria | `--acceptance-criteria` | string | Single value |
| DefinitionOfDone | `--dod` | []string (repeated) | Creates unchecked items |
| Design | `--design` | string | Single value |
| Notes | `--notes` | string | Single value |
| ParentID | `--parent` | string | ID of parent issue |

### Issue Update

| Field | CLI Flag | Type | Notes |
|-------|----------|------|-------|
| DefinitionOfDone | `--dod` | []string (repeated) | Replaces entire DoD |
| (DoD check) | `--check-dod` | string | Exact match, sets checked=true, verified_at=now |
| (DoD uncheck) | `--uncheck-dod` | string | Exact match, sets checked=false, verified_at=nil |
| ParentID | `--parent` | string | Set parent (empty string to clear) |

## Validation Rules

### DoD Operations

1. **DoD item text matching**: Exact match (case-sensitive, no whitespace normalization)
2. **Error on not found**: `--check-dod` and `--uncheck-dod` return error if item text not found
3. **DoD replacement**: `--dod` on update replaces entire DoD, not additive

### Parent-Child Relationships

1. **Single parent constraint**: An issue can only have ONE parent
   - Error: `"issue already has a parent, remove existing parent first"`
2. **Self-parent prevention**: An issue cannot be its own parent
   - Error: `"cannot set self as parent"`
3. **Circular reference prevention**: Parent chain cannot form a cycle
   - Error: `"circular parent-child relationship detected"`
4. **Parent existence check**: Parent issue must exist
   - Error: `"parent issue not found: <id>"`
5. **Hierarchy depth**: Unlimited (no maximum depth restriction)
6. **Idempotent clear**: `--parent ""` on issue without parent succeeds silently

## State Transitions

### DefinitionOfDone Item

```
[unchecked] --check-dod--> [checked + verified_at]
[checked] --uncheck-dod--> [unchecked + verified_at=nil]
```

### Parent-Child Relationship

```
[no parent] --parent SL-xxx--> [parentId=SL-xxx]
[parentId=SL-xxx] --parent ""--> [no parent]
[parentId=SL-xxx] --parent SL-yyy--> ERROR (must clear first)
```

### Issue (no changes to existing transitions)

```
open --(status change)--> in_progress --(status change)--> closed
```

Note: All DoD items being checked does NOT auto-close the issue. Explicit `sl issue close` required.

## Computed Fields

### Children (read-time computation)

Children are NOT stored in the Issue model. They are computed at read time by querying all issues where `parentId == this.ID`.

**Ordering**: Children are returned ordered by:
1. Priority (descending - higher priority first)
2. Creation order / ID (ascending)

```go
// Pseudo-code for GetChildren
func (s *Store) GetChildren(parentID string) ([]Issue, error) {
    issues := s.GetAllIssues()
    var children []Issue
    for _, issue := range issues {
        if issue.ParentID != nil && *issue.ParentID == parentID {
            children = append(children, issue)
        }
    }
    // Sort by priority (desc), then by ID (asc)
    sort.Slice(children, func(i, j int) bool {
        if children[i].Priority != children[j].Priority {
            return children[i].Priority < children[j].Priority
        }
        return children[i].ID < children[j].ID
    })
    return children, nil
}
```

## Cycle Detection Algorithm

```go
// HasCircularParent checks if setting parent would create a cycle
func (s *Store) HasCircularParent(issueID, parentID string) bool {
    visited := make(map[string]bool)
    current := parentID

    for current != "" {
        if current == issueID {
            return true // Cycle detected
        }
        if visited[current] {
            break // Already visited, no cycle through this path
        }
        visited[current] = true

        parent := s.GetIssue(current)
        if parent == nil || parent.ParentID == nil {
            break
        }
        current = *parent.ParentID
    }
    return false
}
```
