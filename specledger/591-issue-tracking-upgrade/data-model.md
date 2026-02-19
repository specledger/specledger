# Data Model: Built-In Issue Tracker

**Feature**: 591-issue-tracking-upgrade | **Date**: 2026-02-18

## Entity Overview

```
┌─────────────────────────────────────────────────────────────┐
│                        Issue                                 │
├─────────────────────────────────────────────────────────────┤
│ id: string (SL-xxxxxx)                                      │
│ title: string                                               │
│ description: string                                         │
│ status: enum (open, in_progress, closed)                    │
│ priority: int (0-5)                                         │
│ issue_type: enum (epic, feature, task, bug)                 │
│ spec_context: string (e.g., "010-my-feature")               │
│ created_at: timestamp                                       │
│ updated_at: timestamp                                       │
│ closed_at: timestamp (nullable)                             │
│ definition_of_done: DefinitionOfDone (optional)             │
│ blocked_by: []string (issue IDs)                            │
│ blocks: []string (issue IDs)                                │
│ labels: []string                                            │
│ assignee: string (optional)                                 │
│ notes: string (optional)                                    │
│ design: string (optional)                                   │
│ acceptance_criteria: string (optional)                      │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                    DefinitionOfDone                          │
├─────────────────────────────────────────────────────────────┤
│ items: []ChecklistItem                                      │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                     ChecklistItem                            │
├─────────────────────────────────────────────────────────────┤
│ item: string                                                │
│ checked: bool                                               │
│ verified_at: timestamp (nullable)                           │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                      IssueStore                              │
├─────────────────────────────────────────────────────────────┤
│ path: string (specledger/<spec>/issues.jsonl)               │
│ issues: []Issue                                             │
│ lock: flock (file lock for concurrent access)               │
└─────────────────────────────────────────────────────────────┘
```

## Entity Definitions

### Issue

The core tracking unit representing a piece of work.

```go
type IssueStatus string

const (
    StatusOpen       IssueStatus = "open"
    StatusInProgress IssueStatus = "in_progress"
    StatusClosed     IssueStatus = "closed"
)

type IssueType string

const (
    TypeEpic    IssueType = "epic"
    TypeFeature IssueType = "feature"
    TypeTask    IssueType = "task"
    TypeBug     IssueType = "bug"
)

type Issue struct {
    // Required fields
    ID          string      `json:"id"`
    Title       string      `json:"title"`
    Description string      `json:"description,omitempty"`
    Status      IssueStatus `json:"status"`
    Priority    int         `json:"priority"` // 0=highest, 5=lowest
    IssueType   IssueType   `json:"issue_type"`
    SpecContext string      `json:"spec_context"`
    CreatedAt   time.Time   `json:"created_at"`
    UpdatedAt   time.Time   `json:"updated_at"`

    // Optional fields
    ClosedAt            *time.Time        `json:"closed_at,omitempty"`
    DefinitionOfDone    *DefinitionOfDone `json:"definition_of_done,omitempty"`
    BlockedBy           []string          `json:"blocked_by,omitempty"`  // Issue IDs
    Blocks              []string          `json:"blocks,omitempty"`      // Issue IDs
    Labels              []string          `json:"labels,omitempty"`
    Assignee            string            `json:"assignee,omitempty"`
    Notes               string            `json:"notes,omitempty"`
    Design              string            `json:"design,omitempty"`
    AcceptanceCriteria  string            `json:"acceptance_criteria,omitempty"`
}
```

**Validation Rules**:
- `ID`: Must match pattern `^SL-[a-f0-9]{6}$`
- `Title`: Required, 1-200 characters
- `Status`: Must be one of `open`, `in_progress`, `closed`
- `Priority`: Must be 0-5
- `IssueType`: Must be one of `epic`, `feature`, `task`, `bug`
- `SpecContext`: Must match pattern `^\d{3,}-[a-z0-9-]+$` (e.g., "010-my-feature")
- `Priority`: 0 = highest (critical), 5 = lowest

**State Transitions**:
```
                 create
    ┌─────────────────────────┐
    │                         ▼
┌───┴───┐               ┌───────────┐
│ open  │──────────────▶│in_progress│
└───┬───┘  start work   └─────┬─────┘
    │                         │
    │ close (skip)            │ close
    │                         │
    ▼                         ▼
┌───────┐               ┌───────┐
│closed │◀──────────────│closed │
└───────┘   complete    └───────┘
```

### DefinitionOfDone

Optional checklist that must be completed before an issue can be closed.

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

**Validation**:
- `DefinitionOfDone` is optional on issues
- If present and any `Checked == false`, `sl issue close` fails without `--force`
- Malformed `definition_of_done` field logs warning and skips validation

### IssueStore

Manages JSONL file operations with file locking.

```go
type Store struct {
    path     string // Path to issues.jsonl
    lockPath string // Path to .issues.jsonl.lock
}

// Core operations
func NewStore(specContext string) (*Store, error)
func (s *Store) Create(issue *Issue) error
func (s *Store) Get(id string) (*Issue, error)
func (s *Store) List(filter ListFilter) ([]Issue, error)
func (s *Store) Update(id string, updates IssueUpdate) error
func (s *Store) Delete(id string) error

// ListFilter for querying issues
type ListFilter struct {
    Status     *IssueStatus
    IssueType  *IssueType
    Priority   *int
    Labels     []string
    SpecContext string  // Empty = all specs
    All        bool     // Search across all specs
}

// IssueUpdate for partial updates
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
    DefinitionOfDone   *DefinitionOfDone
}
```

### Dependency Management

```go
// LinkType for issue relationships
type LinkType string

const (
    LinkBlocks    LinkType = "blocks"     // A blocks B (A must complete before B)
    LinkRelated   LinkType = "related"    // A and B are related
)

// AddDependency creates a link between two issues
func (s *Store) AddDependency(fromID, toID string, linkType LinkType) error

// RemoveDependency removes a link between two issues
func (s *Store) RemoveDependency(fromID, toID string) error

// DetectCycles checks for circular dependencies
func (s *Store) DetectCycles() ([][]string, error)

// GetDependencyTree returns the full dependency tree for an issue
func (s *Store) GetDependencyTree(id string) (*DependencyTree, error)

type DependencyTree struct {
    Issue     Issue
    BlockedBy []DependencyTree
    Blocks    []DependencyTree
}
```

**Cycle Detection Algorithm**:
- Use depth-first search (DFS) to detect cycles
- Cycle creation fails with error: "Cannot create dependency: would create cycle A → B → C → A"

### ID Generation

```go
// GenerateIssueID creates a deterministic, globally unique issue ID
func GenerateIssueID(specContext, title string, createdAt time.Time) string {
    data := fmt.Sprintf("%s|%s|%d", specContext, title, createdAt.UnixNano())
    hash := sha256.Sum256([]byte(data))
    return "SL-" + hex.EncodeToString(hash[:3])
}

// ParseIssueID validates an issue ID format
func ParseIssueID(id string) (string, error) {
    if !strings.HasPrefix(id, "SL-") {
        return "", errors.New("issue ID must start with 'SL-'")
    }
    hexPart := strings.TrimPrefix(id, "SL-")
    if len(hexPart) != 6 {
        return "", errors.New("issue ID must have 6 hex characters after 'SL-'")
    }
    if _, err := hex.DecodeString(hexPart); err != nil {
        return "", errors.New("issue ID must contain valid hex characters")
    }
    return id, nil
}
```

**Collision Probability**:
- 6 hex characters = 16,777,216 possible values
- Birthday problem: P(collision) ≈ n²/(2 * 16.7M)
- For 100,000 issues: P ≈ 0.003% (well under 0.01% threshold)

## JSONL File Format

### File Structure

```
specledger/010-my-feature/issues.jsonl
```

Each line is a complete JSON object:

```jsonl
{"id":"SL-a3f5d8","title":"Add validation","description":"Implement input validation","status":"open","priority":1,"issue_type":"task","spec_context":"010-my-feature","created_at":"2026-02-18T10:00:00Z","updated_at":"2026-02-18T10:00:00Z","labels":["component:api"]}
{"id":"SL-b4e6f9","title":"Fix auth bug","description":"Auth fails on edge case","status":"in_progress","priority":0,"issue_type":"bug","spec_context":"010-my-feature","created_at":"2026-02-18T11:00:00Z","updated_at":"2026-02-18T12:00:00Z","blocked_by":["SL-a3f5d8"]}
```

### File Operations

**Read**: Parse each line as JSON, skip invalid lines with warning
**Write (Create)**: Append new line to end of file
**Write (Update)**: Rewrite entire file with modified issues
**Lock**: Use `.issues.jsonl.lock` for concurrent access protection

### Auto-Merge Strategy

For merge conflicts in `issues.jsonl`:
1. Parse both versions line by line
2. Deduplicate by issue ID
3. Keep both versions' changes where possible (e.g., different fields updated)
4. For conflicting field values, prefer incoming branch (standard git merge)
5. Write merged result

## Beads Migration Mapping

### Source Format (Beads)

```json
{
  "id": "sl-1",
  "title": "...",
  "status": "open",
  "priority": 1,
  "type": "task",
  "labels": ["spec:010-my-feature"],
  ...
}
```

### Target Format (sl issue)

```json
{
  "id": "SL-a3f5d8",
  "title": "...",
  "status": "open",
  "priority": 1,
  "issue_type": "task",
  "spec_context": "010-my-feature",
  "labels": ["spec:010-my-feature"],
  "beads_migration": {
    "original_id": "sl-1",
    "migrated_at": "2026-02-18T10:00:00Z"
  }
  ...
}
```

### Migration Logic

1. **ID Mapping**: Generate new SHA-256 ID, store original ID in `beads_migration.original_id`
2. **Spec Context**: Extract from `labels` matching `spec:###-name` pattern
3. **Fallback**: Issues without spec context → `specledger/migrated/issues.jsonl`
4. **Dependencies**: Map Beads dependencies to new IDs using original ID mapping
5. **Cleanup**: After successful migration, remove Beads dependencies:
   - Delete `.beads/` directory
   - Remove `beads` and `perles` entries from `mise.toml`
   - Create migration log at `specledger/.migration-log`

```go
// Migrator handles migration from Beads to sl issue format
type Migrator struct {
    beadsPath    string // Path to .beads/issues.jsonl
    artifactPath string // Path to specledger/ directory
    dryRun       bool
    keepBeads    bool
}

// Cleanup removes Beads dependencies after successful migration
func (m *Migrator) Cleanup() error {
    if m.keepBeads {
        return nil
    }

    // 1. Remove .beads directory
    if err := os.RemoveAll(".beads"); err != nil {
        return fmt.Errorf("failed to remove .beads: %w", err)
    }

    // 2. Update mise.toml to remove beads and perles
    if err := m.removeFromMiseToml(); err != nil {
        return fmt.Errorf("failed to update mise.toml: %w", err)
    }

    // 3. Write migration log
    return m.writeMigrationLog()
}
```

## Relationships

```
Issue 1 ──blocks──▶ Issue 2
   │                  │
   │                  └──▶ Issue 3 (blocked by Issue 2)
   │
   └──▶ Issue 4 (also blocked by Issue 1)

Issue 5 ◀──related──▶ Issue 6
```

**Dependency Rules**:
1. Cannot create self-referential dependency (A blocks A)
2. Cannot create cycles (A blocks B blocks C blocks A)
3. `blocked_by` and `blocks` are bidirectional (maintain both sides)
4. Deleting an issue removes all its dependencies
