# Data Model: Issues Storage Configuration

**Feature**: 594-issues-storage-config
**Date**: 2026-02-20

## Entity Changes

### Store (pkg/issues/store.go)

The `Store` struct remains unchanged. The `StoreOptions` struct is already designed for this:

```go
type StoreOptions struct {
    BasePath    string // Base path to specledger directory (default: "specledger")
    SpecContext string // Spec context (e.g., "010-my-feature"), empty for cross-spec mode
}
```

**Changes Required**:
1. Update lock file path construction in `NewStore()`:
   - From: `.issues.jsonl.lock`
   - To: `issues.jsonl.lock`

2. Add optional migration logic for old lock files (deferred - handled on access)

### Path Resolution Flow

```text
┌─────────────────────────────────────────────────────────────┐
│                    Issue Command                             │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│           metadata.LoadProjectMetadata(".")                  │
│           Returns: ProjectMetadata or error                  │
└─────────────────────────────────────────────────────────────┘
                              │
                    ┌─────────┴─────────┐
                    │                   │
              Success              Error
                    │                   │
                    ▼                   ▼
        meta.GetArtifactPath()    "specledger/"
                    │                   │
                    └─────────┬─────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│              issues.NewStore(StoreOptions{                   │
│                  BasePath: artifactPath,                     │
│                  SpecContext: detectedSpec,                  │
│              })                                              │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│           File paths constructed:                            │
│   Issues: <basePath>/<spec>/issues.jsonl                     │
│   Lock:   <basePath>/<spec>/issues.jsonl.lock                │
└─────────────────────────────────────────────────────────────┘
```

## File Locations

### Before (Current)

| File Type | Path Pattern |
|-----------|--------------|
| Issues data | `specledger/<spec>/issues.jsonl` |
| Lock file | `specledger/<spec>/.issues.jsonl.lock` |

### After (This Feature)

| File Type | Path Pattern |
|-----------|--------------|
| Issues data | `<artifact_path>/<spec>/issues.jsonl` |
| Lock file | `<artifact_path>/<spec>/issues.jsonl.lock` |

### Examples

| artifact_path | Spec | Issues File | Lock File |
|---------------|------|-------------|-----------|
| `specledger/` (default) | `010-my-feature` | `specledger/010-my-feature/issues.jsonl` | `specledger/010-my-feature/issues.jsonl.lock` |
| `docs/specs/` | `010-my-feature` | `docs/specs/010-my-feature/issues.jsonl` | `docs/specs/010-my-feature/issues.jsonl.lock` |
| `specs/` | `594-issues-storage-config` | `specs/594-issues-storage-config/issues.jsonl` | `specs/594-issues-storage-config/issues.jsonl.lock` |

## Lock File Migration

Old lock files (`.issues.jsonl.lock`) are handled transparently:

1. On `NewStore()`, check if new lock file exists
2. If not, check if old lock file exists
3. If old exists, attempt rename to new location
4. If rename fails (locked by another process), use old lock temporarily
5. New operations create new lock file

**Note**: Migration is best-effort. If old lock is held, the system continues to work.
