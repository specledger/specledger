# Research: Built-In Issue Tracker

**Feature**: 591-issue-tracking-upgrade | **Date**: 2026-02-18

## Prior Work

### Related Features in Codebase

1. **009-command-system-enhancements**: Established Go/Cobra CLI patterns in `pkg/cli/commands/`. Commands follow a consistent pattern:
   - Command definition in `VarXxxCmd` cobra command
   - Flags defined in `init()` function
   - RunE function with error handling
   - Use of `pkg/cli/ui` for consistent output formatting

2. **010-checkpoint-session-capture**: Established patterns for file-based storage with JSON serialization and go-git for repository operations.

3. **Existing Beads Integration** (to be removed):
   - `pkg/cli/prerequisites/checker.go`: Lists `bd` (beads) and `perles` as required core tools
   - `pkg/embedded/templates/specledger/.specledger/scripts/bash/setup-beads.sh`: Initializes beads database
   - `pkg/embedded/templates/specledger/.claude/skills/bd-issue-tracking/`: Beads skill definitions
   - `pkg/embedded/skills/commands/specledger.implement.md`: Uses `bd` commands for issue management
   - `pkg/embedded/skills/commands/specledger.tasks.md`: Creates issues via `bd create`

## Technical Decisions

### 1. Issue ID Generation

**Decision**: SHA-256 based IDs with format `SL-<6-char-hex>`

**Rationale**:
- Deterministic: Same inputs always produce same ID (useful for deduplication)
- Collision-resistant: 16.7M possible values, < 0.01% collision for 100K issues
- No central counter needed (unlike sequential IDs)
- Timestamp precision (nanoseconds) prevents identical IDs

**Alternatives Considered**:
- Sequential counter: Requires per-spec state management, not deterministic
- UUID: 36 characters too long for CLI usage
- Full SHA-256: 64 characters too long

**Implementation**:
```go
// ID format: SL-<6-char-hex>
// Generated from: SHA-256(spec_context + title + created_at)
func GenerateIssueID(specContext, title string, createdAt time.Time) string {
    data := fmt.Sprintf("%s|%s|%d", specContext, title, createdAt.UnixNano())
    hash := sha256.Sum256([]byte(data))
    return "SL-" + hex.EncodeToString(hash[:3]) // First 3 bytes = 6 hex chars
}
```

### 2. Per-Spec Storage

**Decision**: Store issues at `specledger/<spec>/issues.jsonl`

**Rationale**:
- Eliminates cross-branch merge conflicts (each feature branch has its own issues)
- Aligns with feature-centric workflow
- Natural scoping for issue context
- Smaller files = faster operations

**Alternatives Considered**:
- Single file at root: Conflicts when merging branches
- SQLite database: Requires database management, not human-readable
- Git-based storage: Too slow for interactive CLI

### 3. Duplicate Detection Algorithm

**Decision**: Levenshtein distance for string similarity

**Rationale**:
- Standard algorithm, well-tested implementations available
- Works well for short strings like issue titles
- Configurable threshold (80% similarity = warning)

**Alternatives Considered**:
- Jaro-Winkler: Better for names, not titles
- Semantic similarity (ML): Overkill, requires model
- Exact hash match: Too strict, misses similar issues

**Implementation**:
```go
// Using existing library: github.com/texttheater/golang-levenshtein/levenshtein
func CheckDuplicate(newTitle string, existingIssues []Issue) []Issue {
    var duplicates []Issue
    for _, issue := range existingIssues {
        ratio := levenshtein.RatioForStrings([]rune(newTitle), []rune(issue.Title))
        if ratio >= 0.8 { // 80% similarity threshold
            duplicates = append(duplicates, issue)
        }
    }
    return duplicates
}
```

### 4. File Locking Strategy

**Decision**: Use `github.com/gofrs/flock` for cross-platform file locking

**Rationale**:
- Simple API, well-maintained
- Cross-platform (Linux, macOS, Windows)
- Non-blocking option for read operations

**Implementation Pattern**:
```go
func (s *Store) WithLock(fn func() error) error {
    lock := flock.New(s.lockPath)
    locked, err := lock.TryLock()
    if err != nil {
        return err
    }
    if !locked {
        return errors.New("file is locked by another process")
    }
    defer lock.Unlock()
    return fn()
}
```

### 5. Migration Strategy

**Decision**: Map Beads issues to spec directories based on metadata and branch association, then clean up Beads dependencies

**Rationale**:
- Preserve all existing issue data
- Maintain issue history and dependencies
- One-time migration, no ongoing compatibility
- Clean removal of Beads/Perles from project configuration

**Mapping Logic**:
1. Parse `.beads/issues.jsonl` for all issues
2. For each issue, extract branch/feature context from:
   - Issue labels (e.g., `spec:006-authz-authn-rbac`)
   - Issue description (branch references)
   - Git branch history (if available)
3. Write to appropriate `specledger/<spec>/issues.jsonl`
4. Unmapped issues go to `specledger/migrated/` with warning

**Cleanup Logic** (after successful migration):
1. Remove `.beads/` directory and all contents
2. Update `mise.toml`:
   - Remove `beads` entry from `[tools]` section
   - Remove `perles` entry from `[tools]` section
3. Preserve migration log in `specledger/.migration-log` for audit trail

**Implementation**:
```go
func (m *Migrator) Cleanup() error {
    // Remove .beads directory
    if err := os.RemoveAll(".beads"); err != nil {
        return fmt.Errorf("failed to remove .beads: %w", err)
    }

    // Update mise.toml
    misePath := "mise.toml"
    content, err := os.ReadFile(misePath)
    if err != nil {
        return nil // mise.toml may not exist, that's ok
    }

    lines := strings.Split(string(content), "\n")
    var newLines []string
    for _, line := range lines {
        // Skip lines containing beads or perles tool entries
        if strings.Contains(line, "beads") || strings.Contains(line, "perles") {
            if strings.HasPrefix(strings.TrimSpace(line), "ubi:") ||
               strings.Contains(line, "= ") {
                continue // Skip this line
            }
        }
        newLines = append(newLines, line)
    }

    return os.WriteFile(misePath, []byte(strings.Join(newLines, "\n")), 0644)
}
```

### 6. Definition of Done Validation

**Decision**: Checklist format in `definition_of_done` field

**Rationale**:
- Human-readable checklist format
- Can be programmatically parsed
- Optional field (skip if absent)

**Format**:
```json
{
  "definition_of_done": [
    {"item": "Unit tests written", "checked": true},
    {"item": "Code review approved", "checked": false},
    {"item": "Documentation updated", "checked": false}
  ]
}
```

## Best Practices Research

### JSONL File Operations

**Pattern**: Append-only writes with periodic compaction
- New issues: Append to end of file (O(1))
- Updates: Rewrite entire file (acceptable for < 1000 issues)
- Compaction: Remove closed/deleted issues periodically

**Error Handling**:
- Corruption recovery: Parse line-by-line, skip invalid JSON
- Backup before write: Keep `.bak` file
- Atomic rename: Write to temp file, then rename

### CLI Design Patterns (from existing codebase)

**Command Structure**:
```
sl issue create --title "..." --type task --priority 1
sl issue list [--status open] [--type bug] [--all] [--spec 010-foo]
sl issue show SL-a3f5d8
sl issue update SL-a3f5d8 --status in_progress
sl issue close SL-a3f5d8 [--force]
sl issue link SL-a3f5d8 blocks SL-b4e6f9
sl issue migrate
sl issue repair
```

**Output Formats**:
- Default: Human-readable table (use `pkg/cli/ui` patterns)
- `--json`: Structured JSON for programmatic use
- `--quiet`: IDs only for scripting

## Constraints & Considerations

### Performance
- File read: < 50ms for 1000 issues
- File write: < 50ms for 1000 issues
- Duplicate check: < 20ms for 1000 issues (O(n) scan)

### Compatibility
- Go 1.24+ required
- No CGO dependencies for portability
- Works on Windows (file locking via flock)

### Security
- No sensitive data in issue files
- File permissions: 0644 (readable by user/group)
- No network operations required

## Dependencies to Add

```go
require (
    github.com/texttheater/golang-levenshtein v1.0.1  // Duplicate detection
    github.com/gofrs/flock v0.12.1                    // File locking
)
```

## Files to Remove/Modify

### Remove
- `pkg/embedded/templates/specledger/.specledger/scripts/bash/setup-beads.sh`
- `pkg/embedded/templates/specledger/.claude/skills/bd-issue-tracking/` (entire directory)

### Modify
- `pkg/cli/prerequisites/checker.go`: Remove `bd` and `perles` from CoreTools
- `pkg/embedded/templates/specledger/init.sh`: Remove beads setup call
- `pkg/embedded/skills/commands/specledger.implement.md`: Replace `bd` with `sl issue`
- `pkg/embedded/skills/commands/specledger.tasks.md`: Replace `bd create` with `sl issue create`
