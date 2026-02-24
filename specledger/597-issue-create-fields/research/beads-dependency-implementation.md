# Beads Dependency & Parent-Child Implementation Research

**Date:** 2026-02-24
**Source:** https://github.com/steveyegge/beads
**Purpose:** Understanding patterns for implementing dependency relationships and parent-child hierarchies in a Go-based issue tracker

---

## Executive Summary

Beads implements dependencies using a **separate `dependencies` table** with typed edges, combined with **hierarchical ID-based parent-child relationships**. The system uses Dolt (a versioned SQL database) for storage, providing transactional integrity and git-like versioning.

### Key Findings

1. **Dependencies are stored in a separate table** with typed relationships (`blocks`, `parent-child`, `related`, etc.)
2. **Parent-child uses two mechanisms**:
   - **Hierarchical IDs**: Child IDs are derived from parent (e.g., `bd-abc.1`, `bd-abc.1.2`)
   - **Explicit dependencies**: A `parent-child` dependency type for non-ID-based hierarchy
3. **Cycle detection is SQL-based** using recursive CTEs
4. **Single parent is enforced by ID hierarchy** - a child ID can only have one parent by design

---

## 1. Data Structures

### 1.1 Dependency Struct

**File:** `internal/types/types.go`

```go
// Dependency represents a relationship between issues
type Dependency struct {
    IssueID     string         `json:"issue_id"`      // The dependent issue
    DependsOnID string         `json:"depends_on_id"` // What it depends on
    Type        DependencyType `json:"type"`          // Relationship type
    CreatedAt   time.Time      `json:"created_at"`
    CreatedBy   string         `json:"created_by,omitempty"`
    Metadata    string         `json:"metadata,omitempty"`   // JSON blob for type-specific data
    ThreadID    string         `json:"thread_id,omitempty"`  // Groups conversation edges
}
```

### 1.2 Dependency Types

**File:** `internal/types/types.go`

```go
type DependencyType string

const (
    // Workflow types (affect ready work calculation)
    DepBlocks            DependencyType = "blocks"             // Hard dependency
    DepParentChild       DependencyType = "parent-child"       // Hierarchy
    DepConditionalBlocks DependencyType = "conditional-blocks" // B runs only if A fails
    DepWaitsFor          DependencyType = "waits-for"          // Fanout gate

    // Association types (soft links)
    DepRelated           DependencyType = "related"
    DepDiscoveredFrom    DependencyType = "discovered-from"

    // Graph link types
    DepRepliesTo  DependencyType = "replies-to"
    DepRelatesTo  DependencyType = "relates-to"
    DepDuplicates DependencyType = "duplicates"
    DepSupersedes DependencyType = "supersedes"

    // Entity types (HOP foundation)
    DepAuthoredBy DependencyType = "authored-by"
    DepAssignedTo DependencyType = "assigned-to"
    DepApprovedBy DependencyType = "approved-by"
    DepAttests    DependencyType = "attests"

    // Reference types
    DepTracks    DependencyType = "tracks"
    DepUntil     DependencyType = "until"
    DepCausedBy  DependencyType = "caused-by"
    DepValidates DependencyType = "validates"
    DepDelegatedFrom DependencyType = "delegated-from"
)
```

### 1.3 Blocking vs Non-Blocking Dependencies

```go
// AffectsReadyWork returns true if this dependency type blocks work
func (d DependencyType) AffectsReadyWork() bool {
    return d == DepBlocks || d == DepParentChild || d == DepConditionalBlocks || d == DepWaitsFor
}
```

**Key insight:** Only `blocks`, `parent-child`, `conditional-blocks`, and `waits-for` affect the ready queue. Other types are soft links for organization.

---

## 2. Storage Implementation

### 2.1 Database Schema

**File:** `internal/storage/dolt/schema.go` (inferred from queries)

```sql
CREATE TABLE dependencies (
    issue_id VARCHAR(255) NOT NULL,
    depends_on_id VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,
    created_at DATETIME NOT NULL,
    created_by VARCHAR(255),
    metadata TEXT,
    thread_id VARCHAR(255),
    PRIMARY KEY (issue_id, depends_on_id)
);

CREATE TABLE child_counters (
    parent_id VARCHAR(255) PRIMARY KEY,
    last_child INT NOT NULL
);
```

### 2.2 AddDependency Implementation

**File:** `internal/storage/dolt/dependencies.go`

```go
func (s *DoltStore) AddDependency(ctx context.Context, dep *types.Dependency, actor string) error {
    tx, err := s.db.BeginTx(ctx, nil)
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %w", err)
    }
    defer func() { _ = tx.Rollback() }()

    // 1. Validate source issue exists
    var issueExists int
    if err := tx.QueryRowContext(ctx,
        `SELECT COUNT(*) FROM issues WHERE id = ?`, dep.IssueID).Scan(&issueExists); err != nil {
        return fmt.Errorf("failed to check issue existence: %w", err)
    }
    if issueExists == 0 {
        return fmt.Errorf("issue %s not found", dep.IssueID)
    }

    // 2. Validate target issue exists (skip for external references)
    if !strings.HasPrefix(dep.DependsOnID, "external:") {
        var targetExists int
        if err := tx.QueryRowContext(ctx,
            `SELECT COUNT(*) FROM issues WHERE id = ?`, dep.DependsOnID).Scan(&targetExists); err != nil {
            return fmt.Errorf("failed to check target issue existence: %w", err)
        }
        if targetExists == 0 {
            return fmt.Errorf("issue %s not found", dep.DependsOnID)
        }
    }

    // 3. Cycle detection for blocking dependency types
    if dep.Type == types.DepBlocks {
        var reachable int
        if err := tx.QueryRowContext(ctx, `
            WITH RECURSIVE reachable AS (
                SELECT ? AS node, 0 AS depth
                UNION ALL
                SELECT d.depends_on_id, r.depth + 1
                FROM reachable r
                JOIN dependencies d ON d.issue_id = r.node
                WHERE d.type = 'blocks'
                  AND r.depth < 100
            )
            SELECT COUNT(*) FROM reachable WHERE node = ?
        `, dep.DependsOnID, dep.IssueID).Scan(&reachable); err != nil {
            return fmt.Errorf("failed to check for dependency cycle: %w", err)
        }
        if reachable > 0 {
            return fmt.Errorf("adding dependency would create a cycle")
        }
    }

    // 4. Insert with upsert semantics
    if _, err := tx.ExecContext(ctx, `
        INSERT INTO dependencies (issue_id, depends_on_id, type, created_at, created_by, metadata, thread_id)
        VALUES (?, ?, ?, NOW(), ?, ?, ?)
        ON DUPLICATE KEY UPDATE type = VALUES(type), metadata = VALUES(metadata)
    `, dep.IssueID, dep.DependsOnID, dep.Type, actor, metadata, dep.ThreadID); err != nil {
        return fmt.Errorf("failed to add dependency: %w", err)
    }

    s.invalidateBlockedIDsCache()
    return tx.Commit()
}
```

**Key points:**
1. Uses explicit transactions for atomicity
2. Validates both issues exist (referential integrity)
3. Cycle detection uses recursive CTE with depth limit (100)
4. Upsert semantics for idempotency

---

## 3. Parent-Child Implementation

### 3.1 Two Mechanisms for Hierarchy

Beads uses **two complementary mechanisms** for parent-child relationships:

#### Mechanism 1: Hierarchical IDs (Primary)

**File:** `internal/types/id_generator.go`

```go
// GenerateChildID creates a hierarchical child ID
// Format: parent.N (e.g., "bd-af78e9a2.1", "bd-af78e9a2.1.2")
func GenerateChildID(parentID string, childNumber int) string {
    return fmt.Sprintf("%s.%d", parentID, childNumber)
}

// ParseHierarchicalID extracts the parent ID and depth from a hierarchical ID
// Returns: (rootID, parentID, depth)
//
// Examples:
//   "bd-af78e9a2"     -> ("bd-af78e9a2", "", 0)
//   "bd-af78e9a2.1"   -> ("bd-af78e9a2", "bd-af78e9a2", 1)
//   "bd-af78e9a2.1.2" -> ("bd-af78e9a2", "bd-af78e9a2.1", 2)
func ParseHierarchicalID(id string) (rootID, parentID string, depth int) {
    depth = 0
    lastDot := -1
    for i, ch := range id {
        if ch == '.' {
            depth++
            lastDot = i
        }
    }

    if depth == 0 {
        return id, "", 0
    }

    firstDot := -1
    for i, ch := range id {
        if ch == '.' {
            firstDot = i
            break
        }
    }
    rootID = id[:firstDot]
    parentID = id[:lastDot]

    return rootID, parentID, depth
}

// MaxHierarchyDepth is the maximum nesting level (3)
const MaxHierarchyDepth = 3

func CheckHierarchyDepth(parentID string, maxDepth int) error {
    if maxDepth < 1 {
        maxDepth = MaxHierarchyDepth
    }
    depth := 0
    for _, ch := range parentID {
        if ch == '.' {
            depth++
        }
    }
    if depth >= maxDepth {
        return fmt.Errorf("maximum hierarchy depth (%d) exceeded for parent %s", maxDepth, parentID)
    }
    return nil
}
```

#### Mechanism 2: Explicit Parent-Child Dependency

The `parent-child` dependency type can be used for non-ID-based hierarchy:

```go
// From queries.go - filtering by parent
if filter.ParentID != nil {
    parentID := *filter.ParentID
    whereClauses = append(whereClauses,
        "(id IN (SELECT issue_id FROM dependencies WHERE type = 'parent-child' AND depends_on_id = ?) OR id LIKE CONCAT(?, '.%'))")
    args = append(args, parentID, parentID)
}
```

### 3.2 Child ID Generation

**File:** `internal/storage/dolt/queries.go`

```go
// GetNextChildID returns the next available child ID for a parent
func (s *DoltStore) GetNextChildID(ctx context.Context, parentID string) (string, error) {
    tx, err := s.db.BeginTx(ctx, nil)
    if err != nil {
        return "", err
    }
    defer tx.Rollback()

    // Get or create counter
    var lastChild int
    err = tx.QueryRowContext(ctx,
        "SELECT last_child FROM child_counters WHERE parent_id = ?", parentID).Scan(&lastChild)
    if err == sql.ErrNoRows {
        lastChild = 0
    } else if err != nil {
        return "", err
    }

    nextChild := lastChild + 1

    // Upsert counter
    _, err = tx.ExecContext(ctx, `
        INSERT INTO child_counters (parent_id, last_child) VALUES (?, ?)
        ON DUPLICATE KEY UPDATE last_child = ?
    `, parentID, nextChild, nextChild)
    if err != nil {
        return "", err
    }

    if err := tx.Commit(); err != nil {
        return "", err
    }

    return fmt.Sprintf("%s.%d", parentID, nextChild), nil
}
```

### 3.3 Creating a Child Issue

**File:** `cmd/bd/create.go`

```go
// If parent is specified, generate child ID
if parentID != "" {
    ctx := rootCtx
    // Validate parent exists before generating child ID
    _, err := store.GetIssue(ctx, parentID)
    if err != nil {
        if errors.Is(err, storage.ErrNotFound) {
            FatalError("parent issue %s not found", parentID)
        }
        FatalError("failed to check parent issue: %v", err)
    }
    childID, err := store.GetNextChildID(ctx, parentID)
    if err != nil {
        FatalError("%v", err)
    }
    explicitID = childID
}
```

---

## 4. Key Design Decisions

### 4.1 Single Parent Constraint

**How is single parent enforced?**

The single parent constraint is enforced by the **ID hierarchy design**:
- A child ID like `bd-abc.1` can only have one parent (`bd-abc`) by construction
- The ID itself encodes the parent relationship
- You cannot have `bd-abc.1` be a child of both `bd-abc` and `bd-xyz`

For explicit `parent-child` dependencies (non-ID-based), the constraint is NOT enforced at the database level - an issue could theoretically have multiple parent-child dependencies pointing to different parents.

### 4.2 Automatic Children Updates

**Does setting a parent automatically update the parent's children array?**

**No.** Beads does NOT maintain a denormalized `children` array on the parent issue. Instead:

1. Children are queried dynamically using:
   - ID pattern matching: `id LIKE CONCAT(parent_id, '.%')`
   - Dependency table: `SELECT issue_id FROM dependencies WHERE type = 'parent-child' AND depends_on_id = ?`

2. This avoids data synchronization issues and keeps the model simple

### 4.3 Referential Integrity

**Are there referential integrity checks?**

**Yes.** The `AddDependency` function:
1. Validates that `issue_id` exists in the issues table
2. Validates that `depends_on_id` exists (unless it's an external reference)
3. Returns an error if either issue is not found

### 4.4 Circular Reference Prevention

**How does Beads prevent circular parent-child relationships?**

**Two mechanisms:**

1. **ID hierarchy makes cycles impossible** - A child ID like `bd-abc.1` cannot be a parent of `bd-abc` because the ID structure doesn't allow it

2. **Cycle detection for `blocks` dependencies** - Uses recursive CTE:
```sql
WITH RECURSIVE reachable AS (
    SELECT ? AS node, 0 AS depth
    UNION ALL
    SELECT d.depends_on_id, r.depth + 1
    FROM reachable r
    JOIN dependencies d ON d.issue_id = r.node
    WHERE d.type = 'blocks'
      AND r.depth < 100  -- Depth limit prevents infinite loops
)
SELECT COUNT(*) FROM reachable WHERE node = ?
```

3. **Child-to-parent dependency anti-pattern check** - Explicitly blocks adding a dependency from child to parent:
```go
// From cmd/bd/dep.go
if isChildOf(fromID, toID) {
    FatalErrorRespectJSON("cannot add dependency: %s is already a child of %s. Children inherit dependency on parent completion via hierarchy. Adding an explicit dependency would create a deadlock", fromID, toID)
}
```

---

## 5. CLI Commands

### 5.1 Dependency Management

```bash
# Add a blocking dependency
bd dep add bd-42 bd-41                    # bd-42 depends on (blocked by) bd-41
bd dep bd-xyz --blocks bd-abc             # bd-xyz blocks bd-abc

# Add with type
bd dep add bd-42 bd-41 --type parent-child
bd dep add bd-42 bd-41 --type related
bd dep add bd-42 bd-41 --type discovered-from

# Remove dependency
bd dep remove bd-42 bd-41

# List dependencies
bd dep list bd-42                         # What bd-42 depends on
bd dep list bd-42 --direction=up          # What depends on bd-42
bd dep list bd-42 --type=blocks           # Filter by type

# Dependency tree
bd dep tree bd-42                         # Show dependency tree
bd dep tree bd-42 --direction=up          # Show dependent tree
bd dep tree bd-42 --format=mermaid        # Mermaid.js output

# Cycle detection
bd dep cycles
```

### 5.2 Creating Children

```bash
# Create with parent (generates hierarchical ID)
bd create "Subtask title" --parent bd-42

# Results in ID like: bd-42.1, bd-42.2, etc.
```

---

## 6. Recommendations for SpecLedger

Based on this research, here are recommendations for implementing similar functionality in SpecLedger:

### 6.1 Dependency Model

1. **Use a separate `dependencies` table** with typed relationships
2. **Implement cycle detection** using recursive CTEs or graph traversal
3. **Support both blocking and non-blocking** dependency types
4. **Store metadata as JSON** for extensibility

### 6.2 Parent-Child Model

1. **Consider hierarchical IDs** for implicit parent-child relationships
2. **Use a `child_counters` table** for sequential child ID generation
3. **Query children dynamically** rather than storing a denormalized array
4. **Enforce max depth** (Beads uses 3 levels)

### 6.3 Blocking Tree for Issues

For the `sl issue` command enhancement requested:

1. **Add `blocks` and `blocked_by` fields** to the issue model (computed, not stored)
2. **Create explicit dependency records** in a `dependencies` table
3. **Implement `sl issue link` command** for creating dependencies:
   ```bash
   sl issue link SL-001 blocks SL-002
   sl issue link SL-001 blocked-by SL-003
   ```
4. **Generate dependency tree** in tasks.md showing blocking relationships
5. **Validate feature-type issues** have proper blocking tree structure

### 6.4 JSONL Storage Pattern

Since SpecLedger uses JSONL, the dependency record format would be:

```jsonl
{"type": "dependency", "issue_id": "SL-001", "depends_on_id": "SL-002", "dep_type": "blocks", "created_at": "2026-02-24T12:00:00Z", "created_by": "user@example.com"}
```

Query pattern for finding blockers:
```bash
# Find what blocks SL-001
jq 'select(.type == "dependency" and .issue_id == "SL-001" and .dep_type == "blocks") | .depends_on_id' issues.jsonl

# Find what SL-001 blocks
jq 'select(.type == "dependency" and .depends_on_id == "SL-001" and .dep_type == "blocks") | .issue_id' issues.jsonl
```

---

## 7. Source Files Reference

| File | Purpose |
|------|---------|
| `internal/types/types.go` | Core type definitions (Issue, Dependency, DependencyType) |
| `internal/types/id_generator.go` | Hierarchical ID generation and parsing |
| `internal/storage/dolt/dependencies.go` | Dependency CRUD operations |
| `internal/storage/dolt/queries.go` | Query builders including GetNextChildID |
| `internal/validation/issue.go` | Issue validation functions |
| `cmd/bd/dep.go` | Dependency CLI commands |
| `cmd/bd/create.go` | Issue creation with parent support |
| `beads.go` | Public API exports |

---

## 8. Appendix: Code Snippets

### A. Checking if ID is a Child

```go
// isChildOf returns true if childID is a hierarchical child of parentID
func isChildOf(childID, parentID string) bool {
    _, actualParentID, depth := types.ParseHierarchicalID(childID)
    if depth == 0 {
        return false
    }
    if actualParentID == parentID {
        return true
    }
    return strings.HasPrefix(childID, parentID+".")
}
```

### B. Cycle Detection Warning

```go
func warnIfCyclesExist(s *dolt.DoltStore) {
    cycles, err := s.DetectCycles(rootCtx)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Warning: Failed to check for cycles: %v\n", err)
        return
    }
    if len(cycles) == 0 {
        return
    }
    fmt.Fprintf(os.Stderr, "\n%s Warning: Dependency cycle detected!\n", ui.RenderWarn("⚠"))
    // ... print cycle details
}
```

### C. Parent Filtering in Queries

```go
// Parent filtering: filter children by parent issue
if filter.ParentID != nil {
    parentID := *filter.ParentID
    whereClauses = append(whereClauses,
        "(id IN (SELECT issue_id FROM dependencies WHERE type = 'parent-child' AND depends_on_id = ?) OR id LIKE CONCAT(?, '.%'))")
    args = append(args, parentID, parentID)
}
```
