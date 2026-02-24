# Research: Issue Create Fields Enhancement

**Feature**: 597-issue-create-fields
**Date**: 2026-02-22

## Prior Work

### Related Features

| Feature | Description | Relevance |
|---------|-------------|-----------|
| 591-issue-tracking-upgrade | Issue model with all JSONL fields | Model already supports target fields |
| 595-issue-tree-ready | Dependency tree and ready commands | Blocking/dependency logic exists |

### Existing Code Analysis

**`pkg/issues/issue.go`**:
- `Issue` struct already has: `AcceptanceCriteria`, `DefinitionOfDone`, `Design`, `Notes`
- `IssueUpdate` struct already has: `AcceptanceCriteria`, `DefinitionOfDone`, `CheckDoDItem`, `UncheckDoDItem`
- `DefinitionOfDone` struct with `CheckItem()` and `UncheckItem()` methods already implement exact match

**`pkg/cli/commands/issue.go`**:
- `issueCreateCmd` flags: `--title`, `--description`, `--type`, `--priority`, `--labels`, `--spec`, `--force`
- `issueUpdateCmd` flags: `--title`, `--description`, `--status`, `--priority`, `--assignee`, `--notes`, `--design`, `--acceptance-criteria`, `--add-label`, `--remove-label`
- Missing: `--dod`, `--check-dod`, `--uncheck-dod` on update; `--dod`, `--design`, `--acceptance-criteria`, `--notes` on create

## Decisions

### 1. Cobra StringArray for --dod Flag

**Decision**: Use `StringArrayVar` for repeated `--dod` flags

**Rationale**:
- Standard Cobra pattern for repeated string flags
- Allows natural CLI usage: `--dod "Item 1" --dod "Item 2"`
- Avoids comma-splitting complexity (items can contain commas)

**Implementation**:
```go
var issueDoDFlag []string
issueCreateCmd.Flags().StringArrayVar(&issueDoDFlag, "dod", []string{}, "Definition of Done items (can be repeated)")
```

### 2. Exact Text Matching for DoD Operations

**Decision**: Use exact string match (case-sensitive, no whitespace normalization)

**Rationale**:
- Matches existing `CheckItem()` implementation in `pkg/issues/issue.go`
- Predictable behavior - no surprises from normalization
- Clear error message when item not found

**Implementation**: No changes needed to existing `CheckItem()`/`UncheckItem()` methods

### 3. Error Message Format

**Decision**: Return error with format `"DoD item not found: '<text>'"`

**Rationale**:
- Clear identification of what text was searched
- Single quotes distinguish the search text from error message
- Consistent with CLI error patterns

### 4. Prompt Template Strategy

**Decision**: Update both `.claude/commands/` and `pkg/embedded/` copies

**Rationale**:
- `.claude/commands/` - used by this repository
- `pkg/embedded/` - copied to user projects via `sl init`
- Both must stay in sync to ensure consistent behavior

## Alternatives Considered

| Alternative | Rejected Because |
|-------------|------------------|
| Comma-separated `--dod` flag | Items might contain commas; less intuitive UX |
| Case-insensitive DoD matching | Could match wrong item; unpredictable |
| Auto-add DoD item on --check-dod | Hides typos; violates explicit is better than implicit |

## External System Analysis: Beads

**Source**: https://github.com/steveyegge/beads

### Overview

Beads is an AI-assisted issue tracking system designed for small teams. It uses Dolt (a versioned SQL database) for storage and integrates with AI agents via Model Context Protocol (MCP).

### Supported Fields

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Unique identifier (format: `B-xxxxxx`) |
| `title` | string | Issue title |
| `type` | string | Issue type: `task`, `epic`, `bug`, `feature`, `story` |
| `status` | string | Status: `open`, `in_progress`, `closed` |
| `priority` | int | Priority level (0=highest) |
| `created` | timestamp | Creation time |
| `modified` | timestamp | Last modification time |
| `external_ref` | string | External reference (e.g., Jira ticket) |
| `reason` | string | Why this issue exists |
| `assignee` | string | Assigned user |

### Fields NOT Supported (Gap Analysis)

| Field | SpecLedger Has | Beads Has | Impact |
|-------|----------------|-----------|--------|
| `acceptance_criteria` | ✓ | ✗ | Beads users cannot specify acceptance criteria |
| `definition_of_done` | ✓ | ✗ | Beads lacks checklist-style DoD tracking |
| `design` | ✓ | ✗ | Beads has no design notes field |
| `notes` | ✓ | ✗ | Beads has no implementation notes field |
| `labels` | ✓ | ✗ | Beads lacks tagging/labeling system |
| `spec` | ✓ | ✗ | Beads has no spec context linking |

---

## Beads Dependency & Parent-Child Implementation

**Source**: Deep source code analysis from https://github.com/steveyegge/beads
**Storage**: Dolt (versioned SQL database), not JSONL

### Dependency Model

Beads uses a **separate `dependencies` table** with typed edges:

```go
type Dependency struct {
    IssueID     string         // The dependent issue (child)
    DependsOnID string         // What it depends on (parent)
    Type        DependencyType // Relationship type
    CreatedAt   time.Time
    CreatedBy   string
    Metadata    string         // JSON blob for type-specific data
}
```

### Dependency Types

```go
type DependencyType string

const (
    // Workflow types (affect ready work calculation)
    DepBlocks            DependencyType = "blocks"             // Hard dependency
    DepParentChild       DependencyType = "parent-child"       // Hierarchy
    DepConditionalBlocks DependencyType = "conditional-blocks" // B runs only if A fails
    DepWaitsFor          DependencyType = "waits-for"          // Fanout gate

    // Association types (soft links - don't block)
    DepRelated           DependencyType = "related"
    DepDiscoveredFrom    DependencyType = "discovered-from"
    DepRelatesTo         DependencyType = "relates-to"
    DepDuplicates        DependencyType = "duplicates"
    DepSupersedes        DependencyType = "supersedes"
)
```

**Key insight**: Only `blocks`, `parent-child`, `conditional-blocks`, and `waits-for` affect the ready queue. Other types are soft links for organization.

### Parent-Child Implementation

Beads uses **TWO complementary mechanisms** for parent-child:

#### Mechanism 1: Hierarchical IDs (Primary)

Child IDs are derived from parent using dot notation:

```go
// GenerateChildID creates a hierarchical child ID
// Format: parent.N (e.g., "bd-af78e9a2.1", "bd-af78e9a2.1.2")
func GenerateChildID(parentID string, childNumber int) string {
    return fmt.Sprintf("%s.%d", parentID, childNumber)
}
```

**Single parent is enforced by ID hierarchy** - a child ID like `bd-abc.1` can only have one parent (`bd-abc`) by construction.

#### Mechanism 2: Explicit Parent-Child Dependency

For non-ID-based hierarchy, the `parent-child` dependency type is used:

```sql
-- Query children by parent
SELECT issue_id FROM dependencies
WHERE type = 'parent-child' AND depends_on_id = ?
```

### Cycle Detection

Beads uses recursive CTEs with depth limit:

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

### Key Design Decisions from Beads

| Question | Beads Answer |
|----------|--------------|
| How is single parent enforced? | By ID hierarchy design - `bd-abc.1` can only be child of `bd-abc` |
| Does setting parent update children array? | No - children are queried dynamically |
| Are there referential integrity checks? | Yes - validates both issues exist before creating dependency |
| How are circular references prevented? | 1) ID hierarchy makes cycles impossible, 2) Recursive CTE detection |

### Beads CLI Commands

```bash
# Add a blocking dependency
bd dep add bd-42 bd-41                    # bd-42 depends on (blocked by) bd-41

# Add with type
bd dep add bd-42 bd-41 --type parent-child
bd dep add bd-42 bd-41 --type related

# Create with parent (generates hierarchical ID)
bd create "Subtask title" --parent bd-42
# Results in ID like: bd-42.1, bd-42.2, etc.

# Dependency tree
bd dep tree bd-42 --format=mermaid
```

### Recommendations for SpecLedger

Based on Beads analysis, here's how to implement parent-child in SpecLedger:

| Aspect | Beads Approach | SpecLedger Recommendation |
|--------|----------------|---------------------------|
| Storage | Separate `dependencies` table | Keep `Issue.BlockedBy` array (simpler for JSONL) |
| Parent field | Hierarchical ID or dependency record | Add `ParentID *string` field to Issue |
| Single parent | Enforced by ID hierarchy | Enforce in API - only ONE parent allowed |
| Children query | Dynamic query | Compute `Children []string` at read time |
| Cycle prevention | Recursive CTE | Graph traversal in Go code |

### Proposed SpecLedger Implementation

```go
// Issue model additions
type Issue struct {
    // ... existing fields ...

    // Parent relationship (single parent only)
    ParentID   *string  `json:"parentId,omitempty"`   // ID of parent issue

    // Blocking relationships (existing)
    BlockedBy  []string `json:"blockedBy,omitempty"`  // Issues blocking this one
    Blocks     []string `json:"blocks,omitempty"`     // Issues this one blocks
}

// Validation: single parent constraint
func (i *Issue) SetParent(parentID string) error {
    if i.ParentID != nil && *i.ParentID != "" {
        return errors.New("issue already has a parent, remove existing parent first")
    }
    if parentID == i.ID {
        return errors.New("cannot set self as parent")
    }
    i.ParentID = &parentID
    return nil
}
```

### CLI Commands for SpecLedger

```bash
# Set parent (one parent only)
sl issue update SL-001 --parent SL-002

# Remove parent
sl issue update SL-001 --parent ""

# Add blocking dependency
sl issue link SL-001 blocks SL-002

# View with dependency tree
sl issue show SL-001 --tree
```

### Source Files Reference (Beads)

| File | Purpose |
|------|---------|
| `internal/types/types.go` | Core type definitions (Issue, Dependency, DependencyType) |
| `internal/types/id_generator.go` | Hierarchical ID generation and parsing |
| `internal/storage/dolt/dependencies.go` | Dependency CRUD operations |
| `internal/storage/dolt/queries.go` | Query builders including GetNextChildID |
| `cmd/bd/dep.go` | Dependency CLI commands |
| `cmd/bd/create.go` | Issue creation with parent support |

**Full analysis:** `/Users/arielmiki/Workspace/aidev/specledger/specledger/research/beads-dependency-implementation.md`

## No NEEDS CLARIFICATION Items

All technical decisions resolved through code analysis and spec clarifications.
