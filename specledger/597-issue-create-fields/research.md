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

### Dependency Model

Beads uses a **separate`dependencies`table** (not embedded in issues):

```go
type Dependency struct {
    ID          uuid.UUID `db:"id" json:"id"`
    IssueID     uuid.UUID `db:"issue_id" json:"issue_id"`
    DependsOn  uuid.UUID `db:"depends_on" json:"depends_on"`
    Type        string    `db:"type" json:"type"` // "parent-child", "blocks", etc.
    CreatedAt   time.Time `db:"created_at" json:"created_at"`
}
```

### Dependency Types

| Type | Description | CLI |
|------|-------------|-----|
| `blocks` | Issue A blocks Issue B | `--blocks` |
| `blocked-by` | Issue A is blocked by Issue B | `--blocked-by` |
| `parent-child` | Parent/child relationship | `--parent` / `--child` |

### Parent-Child Relationship

Key implementation details from Beads:

1. **Child ID references Parent**: The child stores the parent ID
2. **Single Parent Constraint**: An issue can only have ONE parent
3. **No Circular References**: `issue A -> issue B -> issue A` cycles prevented
4. **Automatic Transitive Closure**: Child inherits parent's dependencies

```go
// SetParent assigns parent (one parent only!)
func (i *Issue) SetParent(parent *Issue) error {
    if i.ParentID != nil {
        return errors.New("issue already has a parent")
    }
    // Prevent cycles
    if parent.ID == i.ID {
        return errors.New("cannot set self as parent")
    }
    // ... dependency created via `parent-child` type
}
```

### Single Parent Enforcement

```go
// GetParent returns the parent of an issue
// Returns nil if issue has no parent
func (s *IssueStore) GetParent(ctx context.Context, id uuid.UUID) (*Issue, error) {
    dep, err := s.GetDependency(ctx, id, "parent-child", "blocked-by")
    // Only ONE dependency of type "parent-child"
    // child --depends-on--> parent
}
```

### Dependency Table Schema

```sql
CREATE TABLE dependencies (
    id UUID PRIMARY KEY,
    issue_id UUID NOT NULL,        -- child issue
    depends_on UUID NOT NULL,     -- parent issue
    type VARCHAR(50) NOT NULL,    -- 'parent-child', 'blocks', etc.
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(issue_id, depends_on, type)  -- prevents duplicates
);

-- Example: Issue A is parent of Issue B
-- issue_id = B (child), depends_on = A (parent), type = 'parent-child'
```

### Recommendations for Specledger

Based on Beads implementation analysis:

| Aspect | Beads Approach | Recommended for Specledger |
|--------|----------------|-----------------------------|
| Storage | Separate `dependencies` table | Use `Issue.dependencies` array |
| Parent limit | ONE parent per issue | ONE parent per issue (enforced in API) |
| Circular refs | Prevented at API level | Prevented at API level |
| Query pattern | Child → parent reference | Store `parentId` on child |

### CLI Commands (from Beads)

```bash
# Set parent
be issue set --parent <issue-id> --child <issue-id>

# Set child (reverse)
be issue set --child <issue-id> --parent <issue-id>

# View dependencies
be issue view <issue-id> --dependencies

# Graph view
be issue tree <issue-id> --dependencies
```

### Implications for Current Task

The current `Issue.dependencies` field in Specledger should:

1. Support `blocks` dependencies (already planned)
2. Add `parent-child` dependency type
3. Enforce single parent via API validation
4. Prevent circular references in API layer

### Key Implementation Pattern

```go
// Dependency represents a single dependency
type Dependency struct {
    Type      string `json:"type"`      // "blocks", "parent-child", etc.
    IssueID   string `json:"issueId"`   --child--
    DependsOn string `json:"dependsOn"` --parent--
    CreatedAt string `json:"createdAt"`
}
```

---

### Implications for Current Task

The current spec already defines `Issue.dependencies` array. We should ensure:

1. **Single Parent**: API validation ensures only ONE parent
2. **Cycle Prevention**: API validation prevents circular references
3. **Dependency Types**: Support `blocks` and `parent-child` types

### CLI Commands (for Specledger)

```bash
# Set parent (one parent only)
sl issue set --parent <issue-id> <child-issue-id>

# Add blocks dependency
sl issue add <issue-id> --blocks <issue-id> --type "blocks"

# Remove parent
sl issue set --parent "" <child-issue-id>
```

### Impact on Current Implementation

| Current Implementation | Beads Pattern | Recommendation |
|------------------------|---------------|----------------|
| `Issue.dependencies` array | Separate `dependencies` table | Keep array approach (simpler) |
| No `parentId` field | Child → parent reference | Add `parentId` field for easier querying |
| No single parent constraint | ONE parent per issue | Add API validation |
| No circular reference check | Prevented at API level | Add API validation |

---

### Source Code Analysis

**Key files from Beads:**
- `/internal/issues/dependencies.go` - Set/Get parent logic
- `/internal/issues/issues.go` - Issue model with dependencies
- `/internal/issues/store.go` - Dependency queries

**Full analysis:** `docs/beads-dependency-analysis.md` (auto-generated by research agent)

## No NEEDS CLARIFICATION Items

All technical decisions resolved through code analysis and spec clarifications.
