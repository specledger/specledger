# Research: Fix SpecLedger Dependencies Integration

**Feature**: 008-fix-sl-deps
**Date**: 2026-02-09
**Status**: Phase 0 Complete

## Overview

This document consolidates research findings for implementing the SpecLedger dependencies integration fixes. The primary goal is to add proper artifact path configuration, complete the resolve functionality, and ensure Claude Code integration is comprehensive.

---

## Key Findings

### 1. Current Implementation State

#### Existing Code Locations

| Component | Location | Status |
|-----------|----------|--------|
| deps.go (CLI commands) | `pkg/cli/commands/deps.go` | Partially implemented |
| Metadata schema | `pkg/cli/metadata/schema.go` | Missing artifact_path |
| Git resolver | `internal/spec/resolver.go` | Exists, uses go-git/v5 |
| Models | `pkg/models/dependency.go` | Different structure than metadata |

#### What Works

- **sl deps add**: Functional, adds dependencies with URL, branch, path, alias
- **sl deps list**: Functional, displays all dependencies
- **sl deps remove**: Functional, removes by URL or alias
- **sl deps resolve**: Partially implemented - uses system git command (should use go-git)
- **Framework detection**: Auto-detects SpecKit/OpenSpec frameworks

#### What's Missing

1. **artifact_path field** - Not in `ProjectMetadata` or `Dependency` structs
2. **--artifact-path flag** - Not available on `sl deps add` for non-SpecLedger repos
3. **Auto-discovery** - No logic to read dependency's specledger.yaml for artifact_path
4. **Reference resolution** - No system to combine artifact_paths for cross-repo references
5. **Auto-download on add** - `sl deps add` doesn't automatically download dependencies
6. **sl new/init** - Don't configure artifact_path field
7. **`--alias` not required** - Should be required, not optional

---

## 2. Data Model Inconsistencies

### Two Different Dependency Structures

**pkg/models/dependency.go** (used by internal/spec/resolver.go):
```go
type Dependency struct {
    RepositoryURL string
    Version       string       // branch, tag, or commit
    SpecPath      string       // path within repo
    Alias         string
    Pinned        bool
    Transitive    []Dependency
}
```

**pkg/cli/metadata/schema.go** (used by deps.go):
```go
type Dependency struct {
    URL            string
    Branch         string
    Path           string       // currently used for spec file path
    Alias          string
    ResolvedCommit string
    Framework      FrameworkChoice
    ImportPath     string
}
```

**Decision**: Keep `metadata.Dependency` as the primary structure (used in YAML). Add `ArtifactPath` field. The `models.Dependency` appears to be from an earlier design that may be deprecated or used internally.

---

## 3. Architecture Decisions

### Artifact Path System

**Concept**: Each project declares where its artifacts (specs) are stored via `artifact_path` in `specledger.yaml`. Dependencies also have artifact paths (auto-discovered for SpecLedger repos, manual for others).

**Reference Resolution Formula**:
```
Full path = project.artifact_path + dependency.alias + ":" + artifact_name
```
(Note: The `dependency.path` field has been removed in favor of using `alias` as the reference path)

**Example**:
```yaml
# Project A's specledger.yaml
artifact_path: specledger/
dependencies:
  - url: git@github.com:org/project-b
    artifact_path: specs/      # auto-discovered from Project B
    path: project-b            # reference within project's artifact_path
    alias: project-b

# Reference "project-b:api.md" resolves to:
# specledger/project-b/api.md
```

### Cache Location

**Current Behavior**: `sl deps resolve` uses `~/.specledger/cache/<dir-name>/` (global cache)

**Code Evidence** (deps.go:329-330):
```go
homeDir, _ := os.UserHomeDir()
cacheDir = filepath.Join(homeDir, ".specledger", "cache", dirName)
```

**Note**: The `resolve-deps.md` command file incorrectly states it uses `specledger/deps/` - this should be updated to reflect the actual global cache location.

---

## 4. Git Operations

### Two Different Git Implementations

**deps.go** uses system git command:
```go
cmd := exec.Command("git", "-C", cacheDir, "rev-parse", "HEAD")
```

**internal/spec/resolver.go** uses go-git/v5:
```go
repo, err := git.PlainCloneContext(ctx, cachePath, false, &git.CloneOptions{
    URL:      dep.RepositoryURL,
    Depth:    1,
})
```

**Decision**: Migrate deps.go to use go-git/v5 for consistency and better error handling. The resolver already exists and can be adapted.

---

## 5. Claude Code Integration

### Existing Command Files

| File | Status | Notes |
|------|--------|-------|
| specledger.add-deps.md | ✅ Exists | Review for artifact_path updates |
| specledger.remove-deps.md | ✅ Exists | Current |
| specledger.list-deps.md | ✅ Exists | Current |
| specledger.resolve-deps.md | ✅ Exists | **Needs update**: cache location incorrect |
| specledger.update-deps.md | ❌ Missing | **Needs creation** |

### Skill Documentation

**File**: `.claude/skills/specledger-deps/SKILL.md`
**Status**: Comprehensive but needs updates for:
- artifact_path concept
- Reference resolution
- Cache location clarification

---

## 6. Implementation Approach

### Phase 1: Data Model Changes

1. **Add ArtifactPath to ProjectMetadata**:
```go
type ProjectMetadata struct {
    // ... existing fields ...
    ArtifactPath string `yaml:"artifact_path,omitempty"` // NEW
    Dependencies []Dependency `yaml:"dependencies,omitempty"`
}
```

2. **Add ArtifactPath to Dependency**:
```go
type Dependency struct {
    // ... existing fields ...
    ArtifactPath string `yaml:"artifact_path,omitempty"` // NEW
}
```

3. **Update sl new/init**:
- Set default `artifact_path: specledger/` when creating new projects
- Detect existing artifact_path when running `sl init`

### Phase 2: Artifact Path Discovery

1. **Create pkg/deps/resolver.go**:
```go
// DetectArtifactPathFromSpecLedgerRepo reads specledger.yaml from a cloned dependency
func DetectArtifactPathFromSpecLedgerRepo(repoPath string) (string, error)

// DetectArtifactPathFromRemote clones, reads, and returns artifact_path
func DetectArtifactPathFromRemote(repoURL, branch, cacheDir string) (string, error)
```

2. **Update sl deps add**:
- Make `--alias` required (no longer optional)
- Remove third argument for path (use alias instead)
- Add `--artifact-path` flag for manual specification
- Auto-detect for SpecLedger repos (check if specledger.yaml exists)
- Auto-download/cache dependency on add (like `go mod`)
- Store result in dependency's artifact_path field

### Phase 3: Complete Resolve Implementation (for manual refresh)

1. **Migrate deps.go to use go-git/v5**:
- Replace `exec.Command("git", ...)` with go-git API calls
- Reuse/adapt logic from `internal/spec/resolver.go`

2. **Update sl deps resolve**:
- Ensure it downloads to `~/.specledger/cache/`
- Handle partial downloads and resume
- Update resolved commit in metadata
- Note: This is for manual refresh (like `go mod download`), auto-download happens on add

### Phase 4: Reference Resolution

1. **Create pkg/deps/reference.go**:
```go
// ResolveReference uses alias to find files
func ResolveReference(projectMeta *ProjectMetadata, depAlias, artifactName string) (string, error)
```

2. **Implement validation**:
- Check if resolved paths exist
- Provide clear error messages

### Phase 5: Claude Code Integration

1. **Update specledger.add-deps.md** (add artifact-path flag, make alias required, note auto-download)
2. **Update specledger.remove-deps.md** (ensure it's current)
3. **Update specledger-deps/SKILL.md** (comprehensive docs for all commands: add, remove, list, update, resolve)

Note: Only `add-deps.md` and `remove-deps.md` command files are needed. Other operations (list, update, resolve) are documented in the skill for reference.

---

## 7. Backward Compatibility

### Migration Strategy

1. **Default artifact_path**: When loading old specledger.yaml files without artifact_path, default to `specledger/`
2. **Optional artifact_path**: The field should be `omitempty` in YAML tags
3. **Validation**: Don't fail if artifact_path is missing, use defaults

### Code Pattern
```go
// GetArtifactPath returns the artifact path, with default fallback
func (m *ProjectMetadata) GetArtifactPath() string {
    if m.ArtifactPath != "" {
        return m.ArtifactPath
    }
    return "specledger/" // default
}
```

---

## 8. Testing Strategy

### Unit Tests

- `pkg/cli/metadata/schema_test.go`: Test artifact_path field
- `pkg/deps/resolver_test.go`: Test artifact path detection
- `pkg/deps/reference_test.go`: Test reference resolution

### Integration Tests

- `tests/integration/deps_test.go`: Test full deps workflow
- Test with real Git repositories (public repos)
- Test cache operations

### Test Scenarios

1. Add SpecLedger repo → auto-detect artifact_path
2. Add non-SpecLedger repo → require --artifact-path flag
3. Resolve dependencies → verify cache location
4. Reference artifacts → verify path resolution
5. Backward compatibility → load old specledger.yaml files

---

## 9. Open Questions / Needs Clarification

### Issue Tracking (Constitution Check)

**Question**: Should we create a new Beads epic for dependencies fix or extend SL-26y?

**Recommendation**: Create a new epic (e.g., SL-deps) as this is distinct from release delivery work.

### Reference Resolution Priority

**Question**: Should reference resolution (FR-013 to FR-016) be implemented in this feature or deferred?

**Recommendation**: Implement core data model changes (artifact_path) and resolve command completion first. Reference resolution can be a follow-up feature as it requires more design around how artifacts are referenced in specs.

### Update Command Implementation

**Question**: The `sl deps update` command is currently a stub. What should it do?

**Options**:
1. Check for new commits on configured branches
2. Re-run resolve to update to latest commits
3. Interactive prompt for each dependency

**Recommendation**: Implement as re-running resolve with explicit confirmation for each dependency that would change.

---

## 10. Summary of Required Changes

| Priority | Component | Change | Files |
|----------|-----------|--------|-------|
| P1 | Metadata schema | Add ArtifactPath to ProjectMetadata, remove Path from Dependency | pkg/cli/metadata/schema.go |
| P1 | deps add | Make --alias required, add --artifact-path flag, auto-download | pkg/cli/commands/deps.go |
| P1 | deps resolve | Complete for manual refresh, use go-git | pkg/cli/commands/deps.go |
| P1 | sl new/init | Configure artifact_path | pkg/cli/commands/new.go, init.go |
| P1 | Claude commands | Update add-deps.md, keep remove-deps.md (no new command files) | .claude/commands/ |
| P1 | Skill docs | Comprehensive docs for all commands (add, remove, list, update, resolve) | .claude/skills/specledger-deps/ |
| P2 | Artifact discovery | Create resolver package | pkg/deps/resolver.go |
| P2 | Reference resolution | Create reference package | pkg/deps/reference.go |

---

## 11. Alternatives Considered

### Alternative 1: Store Full Artifact Path in Dependency

**Approach**: Store the full path (e.g., `specledger/dependency-name/`) in the dependency

**Rejected Because**: Less flexible, harder to change project structure, duplicates information

### Alternative 2: Use Configuration File for Artifact Paths

**Approach**: Separate `.specledger/artifacts.yaml` file

**Rejected Because**: Adds file complexity, specledger.yaml is already the project metadata file

### Alternative 3: No Default artifact_path

**Approach**: Require explicit artifact_path in every project

**Rejected Because**: Breaks backward compatibility, less convenient for standard projects

### Alternative 4: Separate `path` and `alias` fields

**Approach**: Keep both `path` (for reference location) and `alias` (for short name) as separate fields

**Rejected Because**: Unnecessary complexity - `alias` can serve both purposes. The `path` field was redundant since it defaulted to the `alias` value in most cases.

**Decision**: Remove `dependency.path` field and use `alias` as the reference path within the project's artifact_path. This simplifies the data model and reduces user confusion.

### Alternative 5: Separate command file for each deps operation

**Approach**: Create `.claude/commands/` files for all 5 deps operations (add, remove, list, resolve, update)

**Rejected Because**: Only `add` and `remove` are core operations that AI agents need explicit guidance for. `list`, `update`, and `resolve` are simple commands that can be documented in the skill file for reference.

**Decision**: Keep only `add-deps.md` and `remove-deps.md` command files. Document other operations in the `specledger-deps` skill.

---

## 12. Dependencies

### Go Packages Already Used

- `github.com/go-git/go-git/v5` - Git operations (already in go.mod)
- `gopkg.in/yaml.v3` - YAML parsing (already in go.mod)
- `github.com/spf13/cobra` - CLI framework (already in go.mod)

### No New External Dependencies Required

All required functionality can be implemented with existing dependencies.

---

**Next Steps**: Proceed to Phase 1 (Design & Contracts) to generate data-model.md, contracts/, and quickstart.md.
