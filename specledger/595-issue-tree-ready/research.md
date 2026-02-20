# Research: Issue Tree View and Ready Command

**Feature**: 595-issue-tree-ready
**Date**: 2026-02-20

## Prior Work Analysis

### Related Features

| Feature | Description | Relevance |
|---------|-------------|-----------|
| 591-issue-tracking-upgrade | Built-in issue tracking with JSONL storage, dependencies, CLI commands | Foundation - Issue entity, Store, dependencies |
| 594-issues-storage-config | Configurable artifact paths, file locking | Storage layer already supports cross-spec queries |

### Existing Code Assets

| File | What Exists | What's Missing |
|------|-------------|----------------|
| `pkg/issues/issue.go` | Issue entity with BlockedBy/Blocks fields, ListFilter with Blocked flag | No Ready filter, no IsReady method |
| `pkg/issues/dependencies.go` | DependencyTree struct, GetDependencyTree(), DetectCycles(), cycle detection | No tree rendering, no ready state computation |
| `pkg/issues/store.go` | Store with List(), Get(), file locking, ListAllSpecs() | No ListReady() method |
| `pkg/cli/commands/issue.go` | --tree flag (defined but not implemented), --blocked flag | Tree output not rendered, no ready command |
| `pkg/embedded/skills/commands/specledger.implement.md` | Uses `sl issue list --status open` for task selection | Should use ready command instead |

## Technical Decisions

### Decision 1: Tree Rendering Approach

**Decision**: ASCII tree characters for terminal output

**Rationale**:
- Maximum terminal compatibility (works in all shells)
- Standard visual representation used by tools like `tree`, `npm list`
- No external dependencies required
- Bubble Tea TUI is overkill for simple tree output

**Alternatives Considered**:
| Approach | Pros | Cons | Rejected Because |
|----------|------|------|------------------|
| Bubble Tea TUI | Rich interactivity | Heavy dependency, complex | Overkill for static tree |
| JSON output only | Simple | Not human-friendly | Users want visual tree |
| Unicode box drawing | More characters | Windows compatibility issues | ASCII is safer |

**Implementation Pattern**:
```
SL-abc123 [open] Setup database
├── SL-def456 [open] Create schema
│   └── SL-ghi789 [open] Add indexes
└── SL-jkl012 [closed] Configure connection
    └── SL-mno345 [open] Test queries
```

### Decision 2: Ready State Computation

**Decision**: Compute ready state at query time (no caching)

**Rationale**:
- Simplicity - no cache invalidation logic needed
- Ready state changes frequently as issues close
- Typical spec has < 50 issues, computation is fast
- Avoids complexity of tracking dependent issue changes

**Ready State Definition**:
```
Issue is "ready" when:
  - Status is "open" OR "in_progress"
  - AND BlockedBy array is empty
  - OR ALL issues in BlockedBy have status "closed"
```

**Alternatives Considered**:
| Approach | Pros | Cons | Rejected Because |
|----------|------|------|------------------|
| Cached ready flag | O(1) lookup | Cache invalidation complexity | Over-engineering for small data |
| Materialized view | Fast queries | Requires storage changes | Breaks JSONL simplicity |
| Background computation | Always fresh | Async complexity | Not needed for this scale |

### Decision 3: Implement Workflow Integration

**Decision**: Update prompt template to use `sl issue ready` command

**Rationale**:
- Minimal change - update command reference in template
- Consistent with existing workflow patterns
- AI agent receives ready tasks automatically

**Implementation**:
Change from:
```
sl issue list --status open
```
To:
```
sl issue ready
```

When no ready tasks:
```
sl issue ready
# Output: "No ready issues. All open issues are blocked."
# Show: "Blocked by: SL-xxx (open), SL-yyy (in_progress)"
```

## Tree Rendering Details

### Character Set

| Character | Usage |
|-----------|-------|
| `├─` | Branch with more siblings below |
| `└─` | Last branch (no more siblings) |
| `│  ` | Vertical connector for nested items |
| `   ` | Indentation for last item's children |

### Algorithm

1. Build adjacency list from issues (Blocks relationships)
2. Find root nodes (issues not blocked by anything)
3. For each root, recursively render tree
4. Track visited nodes to handle cycles gracefully
5. Limit depth to prevent stack overflow (max 10 levels)

### Output Format

```
SL-ID [STATUS] Title (truncated to 40 chars)
├── CHILD-1 [STATUS] Child title...
│   └── GRANDCHILD [STATUS] Grandchild...
└── CHILD-2 [STATUS] Another child...
```

### Cycle Detection

When cycle detected:
1. Display warning at top of output
2. Show cycle path: `Cycle: SL-a → SL-b → SL-a`
3. Continue rendering non-cyclic portions
4. Mark cyclic nodes with `⚠` indicator

## Ready Command Details

### Command Signature

```bash
sl issue ready [flags]

Flags:
  --all     List ready issues across all specs
  --json    Output as JSON
```

### Output Format (table)

```
ID         TITLE                          STATUS       PRIORITY  SPEC
SL-abc123  Ready to work task             open         1         595-issue-tree-ready
SL-def456  Another unblocked task         in_progress  2         595-issue-tree-ready
```

### Output Format (JSON)

```json
[
  {
    "id": "SL-abc123",
    "title": "Ready to work task",
    "status": "open",
    "priority": 1,
    "spec_context": "595-issue-tree-ready"
  }
]
```

### Blocked Message

When all issues are blocked:
```
No ready issues found.

Blocked issues:
  SL-abc123 "Setup database" is blocked by:
    - SL-xyz789 "Create connection" (open)
  SL-def456 "Add tests" is blocked by:
    - SL-xyz789 "Create connection" (open)
    - SL-qwe012 "Write schema" (in_progress)
```

## Performance Considerations

### Ready State Query

For N issues with average M blockers:
- Load all issues: O(N)
- For each issue, check blockers: O(N * M)
- Total: O(N * M)

With N=100, M=3: ~300 operations, < 10ms

### Tree Rendering

For N issues in tree:
- Build adjacency: O(N)
- Find roots: O(N)
- Render: O(N)
- Total: O(N)

With N=100: < 100ms including I/O

## Test Strategy

### Unit Tests

| Test | Description |
|------|-------------|
| `TestIsReady` | Verify ready state computation for various dependency scenarios |
| `TestRenderTree` | Verify tree output formatting |
| `TestCycleDetection` | Verify cycles are detected and handled |
| `TestTruncation` | Verify title truncation at 40 chars |

### Integration Tests

| Test | Description |
|------|-------------|
| `TestReadyCommand` | End-to-end ready command with real issue store |
| `TestTreeCommand` | End-to-end tree command with dependencies |
| `TestCrossSpecReady` | Ready command with --all flag |

## Constraints

1. **Backward Compatibility**: Existing CLI flags must continue to work
2. **No Storage Changes**: Use existing JSONL format, no schema migrations
3. **Terminal Width**: Assume 80-char minimum, truncate long titles
4. **No External Deps**: Use only standard library + existing dependencies
