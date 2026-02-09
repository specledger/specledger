# Data Model: Fix SpecLedger Dependencies Integration

**Feature**: 008-fix-sl-deps
**Date**: 2026-02-09
**Phase**: Phase 1 - Design & Contracts

## Overview

This document defines the data model changes required for the SpecLedger dependencies integration fix. The primary change is adding `artifact_path` to both project metadata and dependency structures.

---

## Core Entities

### 1. ProjectMetadata (Modified)

**Location**: `pkg/cli/metadata/schema.go`

**Changes**: Add `ArtifactPath` field

```go
type ProjectMetadata struct {
    Version      string          `yaml:"version"`
    Project      ProjectInfo     `yaml:"project"`
    Playbook     PlaybookInfo    `yaml:"playbook"`
    TaskTracker  TaskTrackerInfo `yaml:"task_tracker,omitempty"`
    ArtifactPath string          `yaml:"artifact_path,omitempty"` // NEW
    Dependencies []Dependency    `yaml:"dependencies,omitempty"`
}
```

**Fields**:
| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| artifact_path | string | No | `specledger/` | Path to artifacts directory relative to project root |

**Validation**:
- If specified, must be a valid relative path (not absolute)
- Must not contain `..` (parent directory references)
- Must not start with `/`

**Helper Method**:
```go
// GetArtifactPath returns the artifact path with default fallback
func (m *ProjectMetadata) GetArtifactPath() string {
    if m.ArtifactPath != "" {
        return m.ArtifactPath
    }
    return "specledger/"
}
```

### 2. Dependency (Modified)

**Location**: `pkg/cli/metadata/schema.go`

**Changes**: Add `ArtifactPath` field

```go
type Dependency struct {
    URL            string          `yaml:"url"`
    Branch         string          `yaml:"branch,omitempty"`
    Path           string          `yaml:"path,omitempty"`
    Alias          string          `yaml:"alias,omitempty"`
    ArtifactPath   string          `yaml:"artifact_path,omitempty"` // NEW
    ResolvedCommit string          `yaml:"resolved_commit,omitempty"`
    Framework      FrameworkChoice `yaml:"framework,omitempty"`
    ImportPath     string          `yaml:"import_path,omitempty"`
}
```

**Fields**:
| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| url | string | Yes | - | Git repository URL |
| branch | string | No | `main` | Git branch name |
| path | string | No | `<alias>` | Reference path within project's artifact_path |
| alias | string | No | `<generated>` | Short name for the dependency |
| **artifact_path** | **string** | **No** | **auto-discovered** | **Path to artifacts within dependency repo** |
| resolved_commit | string | No | - | Git commit SHA when resolved |
| framework | FrameworkChoice | No | `none` | Framework type (speckit, openspec, both, none) |
| import_path | string | No | - | AI import path format |

**Artifact Path Behavior**:
- **SpecLedger repos**: Auto-discovered by reading dependency's `specledger.yaml`
- **Non-SpecLedger repos**: Must be specified via `--artifact-path` flag
- If not specified and not auto-discoverable: defaults to empty string (root of repo)

### 3. Artifact Reference (New Concept)

**Concept**: A string reference format for cross-repository artifact references.

**Format**: `<dependency-alias>:<artifact-name>`

**Examples**:
- `api-specs:openapi.yaml` - References openapi.yaml from api-specs dependency
- `shared-specs:common/types.md` - References common/types.md from shared-specs

**Resolution**: `project.artifact_path + dependency.path + "/" + artifact-name`

**Example Resolution**:
```yaml
# Project specledger.yaml
artifact_path: specledger/
dependencies:
  - url: git@github.com:org/api-specs
    artifact_path: specs/
    path: api
    alias: api-specs

# Reference "api-specs:openapi.yaml"
# Resolves to: specledger/api/openapi.yaml
```

---

## State Diagrams

### Dependency Lifecycle

```
┌─────────────────────────────────────────────────────────────────────┐
│                        Dependency Lifecycle                          │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  ┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐     │
│  │   Add    │───▶│ Detect   │───▶│ Resolve  │───▶│ Resolved │     │
│  │          │    │ Artifact │    │ (Cache)  │    │          │     │
│  └──────────┘    └──────────┘    └──────────┘    └──────────┘     │
│                      │                                      │        │
│                      │                                      ▼        │
│                      │                                 ┌──────────┐ │
│                      │                                 │ Reference│ │
│                      │                                 │   Use    │ │
│                      │                                 └──────────┘ │
│                      ▼                                      ▲        │
│                 ┌──────────┐                                │        │
│                 │  Manual  │────────────────────────────────┘        │
│                 │  Flag    │                                         │
│                 └──────────┘                                         │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

**States**:
1. **Add**: Dependency added to specledger.yaml
2. **Detect Artifact**: System reads dependency's specledger.yaml (if SpecLedger repo)
3. **Resolve**: Dependency cloned to cache, commit SHA recorded
4. **Resolved**: Dependency ready for use, can be referenced
5. **Manual Flag**: User provides artifact_path manually for non-SpecLedger repos

---

## YAML Structure Examples

### Complete specledger.yaml with artifact_path

```yaml
version: "1.0.0"

project:
  name: my-service
  short_code: ms
  created: 2026-02-09T10:00:00Z
  modified: 2026-02-09T10:00:00Z
  version: "0.1.0"

artifact_path: specledger/

playbook:
  name: specledger
  version: "1.0.0"
  applied_at: 2026-02-09T10:00:00Z
  structure:
    - specledger/
    - .claude/

dependencies:
  # SpecLedger repository - artifact_path auto-discovered
  - url: git@github.com:org/platform-specs
    branch: main
    artifact_path: specs/
    path: platform
    alias: platform
    resolved_commit: abc123def456789abc123def456789abc123def4
    framework: both
    import_path: @platform/spec

  # Non-SpecLedger repository - artifact_path specified manually
  - url: https://github.com/external/api-docs
    branch: v2.0
    artifact_path: docs/openapi/
    path: api-docs
    alias: api-docs
    resolved_commit: def456abc123789def456abc123789def456abc12
    framework: none
```

### Minimal specledger.yaml (backward compatible)

```yaml
version: "1.0.0"

project:
  name: simple-project
  short_code: sp
  created: 2026-02-09T10:00:00Z
  modified: 2026-02-09T10:00:00Z
  version: "0.1.0"

# artifact_path defaults to "specledger/"

playbook:
  name: specledger
  version: "1.0.0"

dependencies:
  - url: git@github.com:org/specs
    branch: main
    path: specs
    alias: specs
```

---

## Reference Resolution Algorithm

### Input
- `projectMeta`: ProjectMetadata
- `reference`: string in format `<alias>:<artifact>`

### Algorithm
```go
func ResolveArtifactReference(projectMeta *ProjectMetadata, reference string) (string, error) {
    // Parse reference
    parts := strings.SplitN(reference, ":", 2)
    if len(parts) != 2 {
        return "", fmt.Errorf("invalid reference format: %s (expected alias:artifact)", reference)
    }

    alias := parts[0]
    artifactName := parts[1]

    // Find dependency by alias
    var dep *Dependency
    for i := range projectMeta.Dependencies {
        if projectMeta.Dependencies[i].Alias == alias {
            dep = &projectMeta.Dependencies[i]
            break
        }
    }

    if dep == nil {
        return "", fmt.Errorf("dependency not found: %s", alias)
    }

    // Build path: project.artifact_path + dep.path + "/" + artifactName
    artifactBase := projectMeta.GetArtifactPath()
    depPath := dep.Path
    if depPath == "" {
        depPath = dep.Alias
    }

    fullPath := filepath.Join(artifactBase, depPath, artifactName)

    // Validate path exists
    if _, err := os.Stat(fullPath); os.IsNotExist(err) {
        return "", fmt.Errorf("artifact not found: %s", fullPath)
    }

    return fullPath, nil
}
```

### Example Walkthrough

**Input**:
- Project artifact_path: `specledger/`
- Dependency: `{alias: "api", path: "api-specs", artifact_path: "specs/"}`
- Reference: `api:openapi.yaml`

**Steps**:
1. Parse reference: alias="api", artifact="openapi.yaml"
2. Find dependency with alias="api"
3. Build path: `specledger/` + `api-specs/` + `openapi.yaml`
4. Result: `specledger/api-specs/openapi.yaml`

---

## Validation Rules

### ProjectMetadata Validation

```go
func (m *ProjectMetadata) Validate() error {
    // ... existing validation ...

    // New: Validate artifact_path if specified
    if m.ArtifactPath != "" {
        if strings.HasPrefix(m.ArtifactPath, "/") {
            return errors.New("artifact_path must be relative, not absolute")
        }
        if strings.Contains(m.ArtifactPath, "..") {
            return errors.New("artifact_path cannot contain parent directory references")
        }
    }

    return nil
}
```

### Dependency Validation

```go
func (d *Dependency) Validate() error {
    // ... existing validation ...

    // New: artifact_path validation
    if d.ArtifactPath != "" {
        if strings.HasPrefix(d.ArtifactPath, "/") {
            return errors.New("dependency artifact_path must be relative")
        }
        if strings.Contains(d.ArtifactPath, "..") {
            return errors.New("dependency artifact_path cannot contain parent references")
        }
    }

    return nil
}
```

---

## Backward Compatibility

### Migration Rules

1. **Loading old specledger.yaml without artifact_path**:
   - Default to `"specledger/"`
   - No validation error

2. **Dependencies without artifact_path**:
   - SpecLedger repos: Auto-discover when resolving
   - Non-SpecLedger repos: Leave empty, treat as root directory

3. **YAML omitempty**:
   - Both artifact_path fields use `yaml:"artifact_path,omitempty"`
   - Empty values won't be written to YAML

### Migration Code

```go
// LoadFromProject with backward compatibility
func LoadFromProject(projectDir string) (*ProjectMetadata, error) {
    meta, err := loadRaw(projectDir)
    if err != nil {
        return nil, err
    }

    // Ensure artifact_path has a default
    if meta.ArtifactPath == "" {
        meta.ArtifactPath = "specledger/"
    }

    return meta, nil
}
```

---

## File Structure After Changes

```
pkg/
├── cli/
│   ├── commands/
│   │   ├── deps.go              # Updated: --artifact-path flag, go-git
│   │   ├── init.go              # Updated: Detect artifact_path
│   │   └── new.go               # Updated: Set default artifact_path
│   └── metadata/
│       ├── schema.go            # Updated: Add ArtifactPath fields
│       ├── yaml.go              # Existing: Load/Save
│       └── validator.go         # NEW: Validation helpers
├── deps/
│   ├── resolver.go              # NEW: Artifact path detection
│   ├── cache.go                 # NEW: Cache operations
│   └── reference.go             # NEW: Reference resolution
└── models/
    └── dependency.go            # EXISTING: May be deprecated
```

---

**Next**: Generate API contracts in `contracts/` directory.
