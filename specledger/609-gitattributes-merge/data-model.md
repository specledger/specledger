# Data Model: Gitattributes Merge

## Entities

### Sentinel Block

A delimited section within a text file managed by specledger.

| Attribute | Type | Description |
|-----------|------|-------------|
| BeginMarker | string (const) | `# >>> specledger-generated` |
| EndMarker | string (const) | `# <<< specledger-generated` |
| ManagedContent | string | Content between markers (from embedded template) |

**Lifecycle**: Created on first `sl init` → Updated on subsequent `sl init` runs → Never deleted by specledger

### Mergeable File Declaration

A manifest entry that opts a structure file into merge (vs copy) behavior.

| Attribute | Type | Description |
|-----------|------|-------------|
| Path | string | Relative path in playbook structure (e.g., `.gitattributes`) |

**Storage**: `mergeable` list in `manifest.yaml` under playbook definition

### Extended Playbook Struct

```go
type Playbook struct {
    // ... existing fields ...
    Mergeable []string `yaml:"mergeable,omitempty"` // NEW
}
```

### Extended CopyResult Struct

```go
type CopyResult struct {
    // ... existing fields ...
    FilesMerged int // NEW
}
```

## State Transitions

```
.gitattributes state:

  [Not Exists] --sl init--> [Created with sentinel block]
  [Exists, no sentinels] --sl init--> [Appended sentinel block]
  [Exists, valid sentinels] --sl init--> [Sentinel section replaced]
  [Exists, malformed sentinel] --sl init--> [Begin-to-EOF replaced with sentinel block]
```

## No Contracts Needed

This feature is a CLI-internal change with no API surface. All changes are within the playbook copy system and embedded templates.
