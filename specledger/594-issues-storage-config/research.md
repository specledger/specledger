# Research: Issues Storage Configuration

**Feature**: 594-issues-storage-config
**Date**: 2026-02-20

## Prior Work

### Related Features

- **591-issue-tracking-upgrade**: Original implementation of the issue tracking system. Created `pkg/issues/store.go` with lock file at `.issues.jsonl.lock` and hardcoded base path `specledger/`.
- **593-ticket-rename**: Related feature for renaming `sl issue` to `sl ticket`. This feature (594) focuses specifically on storage configuration.

### Existing Infrastructure

| Component | Location | Status |
|-----------|----------|--------|
| `ProjectMetadata.GetArtifactPath()` | `pkg/cli/metadata/schema.go:142` | Ready to use |
| `ProjectMetadata.ArtifactPath` field | `pkg/cli/metadata/schema.go:18` | Already defined |
| `metadata.LoadProjectMetadata()` | `pkg/cli/metadata/yaml.go` | Ready to use |
| File lock implementation | `pkg/issues/store.go` (gofrs/flock) | Working |
| JSONL store | `pkg/issues/store.go` | Working |

## Research Findings

### 1. Lock File Naming Convention

**Decision**: Change from `.issues.jsonl.lock` to `issues.jsonl.lock`

**Rationale**:
- Leading dot conventionally denotes "hidden" files in Unix systems
- Lock files should be visible to users for debugging purposes
- Standard lock file naming (e.g., `package-lock.json`, `yarn.lock`) doesn't use leading dots
- Enables simple `.gitignore` pattern matching

**Alternatives Considered**:
- Keep `.issues.jsonl.lock`: Rejected - hidden files harder to debug
- Use `.lock` extension (`issues.jsonl.lock`): Rejected - lock is for the JSONL file, not an extension

### 2. Artifact Path Integration

**Decision**: Load `artifact_path` from `specledger.yaml` via `metadata.LoadProjectMetadata()`

**Rationale**:
- `GetArtifactPath()` method already exists with default fallback to `specledger/`
- Consistent with how other commands (deps) resolve artifact paths
- Single source of truth for path configuration

**Implementation Pattern**:
```go
// In issue commands
meta, err := metadata.LoadProjectMetadata(".")
if err != nil {
    // Fall back to default on error
    basePath = "specledger/"
} else {
    basePath = meta.GetArtifactPath()
}

store, err := issues.NewStore(issues.StoreOptions{
    BasePath:    basePath,
    SpecContext: specContext,
})
```

### 3. Gitignore Pattern

**Decision**: Add `issues.jsonl.lock` to project .gitignore

**Rationale**:
- Lock files are process-specific and should not be committed
- Consistent with other lock file practices (e.g., `.DS_Store`, `*.swp`)
- Simple pattern without wildcards for clarity

**Embedded Templates**:
- Check if embedded templates include a .gitignore
- If not, document in quickstart.md for new project setup

### 4. Backward Compatibility

**Decision**: Handle old `.issues.jsonl.lock` files gracefully

**Rationale**:
- Users may have existing lock files after upgrade
- System should prefer new naming and clean up old files on access

**Implementation**:
- On store creation, check if old lock file exists
- If new lock doesn't exist but old does, attempt to acquire old lock
- Release and delete old lock, create new lock
- This migration happens transparently during normal operations

## Open Questions Resolved

| Question | Resolution |
|----------|------------|
| How to load artifact_path? | Use existing `metadata.LoadProjectMetadata()` and `GetArtifactPath()` |
| Handle missing specledger.yaml? | Fall back to default `specledger/` path |
| Handle malformed yaml? | Fall back to default path, log warning |
| Migrate old lock files? | Yes, transparent migration on first access |

## Dependencies

No new dependencies required. All infrastructure exists in:
- `pkg/cli/metadata/` - Configuration loading
- `pkg/issues/` - Store implementation
- `gofrs/flock` - Already in go.mod
