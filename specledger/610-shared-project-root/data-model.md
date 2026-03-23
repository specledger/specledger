# Data Model: Shared Project Root Resolution

**Date**: 2026-03-23
**Feature**: 610-shared-project-root

## Entities

This feature introduces no new data entities. It extracts and shares existing logic.

### Existing Entity: Project Root

- **Marker file**: `specledger/specledger.yaml` (path defined by `metadata.DefaultMetadataFile`)
- **Detection**: `metadata.HasYAMLMetadata(dir string) bool` — checks if the marker file exists at the given directory
- **Resolution**: Walk up directory tree from current directory, checking each level with `HasYAMLMetadata()`

### New Functions (no new types)

| Function | Package | Signature | Description |
|----------|---------|-----------|-------------|
| `FindProjectRootFrom` | `pkg/cli/metadata` | `(startDir string) (string, error)` | Core logic — walks up from startDir |
| `FindProjectRoot` | `pkg/cli/metadata` | `() (string, error)` | Convenience — calls FindProjectRootFrom with os.Getwd() |

### Error Cases

| Condition | Error Message |
|-----------|---------------|
| No specledger.yaml found up to filesystem root | `"not in a SpecLedger project (no specledger.yaml found). Run 'sl init' to create one, or navigate to a project directory."` |
| Cannot determine current directory | Wraps os.Getwd() error |
