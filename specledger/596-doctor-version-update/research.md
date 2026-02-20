# Research: Doctor Version and Template Update

**Feature**: 596-doctor-version-update
**Date**: 2026-02-20

## Prior Work

### Related Features

1. **135-fix-missing-chmod-x**: Fixed executable permissions for templates during initialization. Relevant for understanding how templates are applied to projects.

2. **011-streamline-onboarding**: Added template embedding system and project initialization. Established the `pkg/embedded/embedded.go` pattern for embedding skills and templates.

3. **006-opensource-readiness**: Established GitHub Releases for distribution. The release workflow is already in place with GoReleaser.

### Existing Code Patterns

**Version Handling** (`cmd/sl/main.go`):
```go
var (
    version   = "dev"
    commit    = "unknown"
    date      = "unknown"
    buildType = "development"
)
```
Version is set at build time via ldflags from GoReleaser.

**Metadata Schema** (`pkg/cli/metadata/schema.go`):
- `ProjectMetadata` struct with `Project`, `Playbook`, `Dependencies` fields
- Playbook has `Version` field tracking which playbook version was applied
- Missing: No `TemplateVersion` field for tracking template version

**Doctor Command** (`pkg/cli/commands/doctor.go`):
- Checks core tools (mise) and framework tools (specify, openspec)
- Has `--json` flag for CI/CD support
- Uses `prerequisites.PrerequisiteCheck` for tool status

**Embedded Templates** (`pkg/embedded/embedded.go`):
```go
//go:embed all:skills
var SkillsFS embed.FS

//go:embed all:templates
var TemplatesFS embed.FS
```

**Template Application** (`pkg/cli/commands/bootstrap_helpers.go`):
- `applyEmbeddedSkills()` - Copies skills to `.claude/`
- `applyEmbeddedPlaybooks()` - Applies playbooks via `playbooks.ApplyToProject()`
- Already handles file permissions (0755 for executables, 0644 for regular)

## Research Findings

### 1. GitHub Releases API

**Decision**: Use GitHub REST API for version checking

**API Endpoint**:
```
GET https://api.github.com/repos/specledger/specledger/releases/latest
```

**Response** (relevant fields):
```json
{
  "tag_name": "v1.2.3",
  "name": "SpecLedger v1.2.3",
  "html_url": "https://github.com/specledger/specledger/releases/tag/v1.2.3",
  "assets": [...]
}
```

**Implementation Notes**:
- Use `net/http` with 5-second timeout
- Handle rate limiting (60 requests/hour for unauthenticated)
- Parse `tag_name` as semver (strip "v" prefix)
- Cache result for session duration to avoid repeated calls

**Alternatives Considered**:
- GoReleaser update notifications: Requires additional integration
- Homebrew version check: Only works for Homebrew installations

### 2. Template Version Storage

**Decision**: Add `template_version` field to `ProjectMetadata`

**Schema Change** (`pkg/cli/metadata/schema.go`):
```go
type ProjectMetadata struct {
    Version         string          `yaml:"version"`
    Project         ProjectInfo     `yaml:"project"`
    Playbook        PlaybookInfo    `yaml:"playbook"`
    TemplateVersion string          `yaml:"template_version,omitempty"` // NEW
    TaskTracker     TaskTrackerInfo `yaml:"task_tracker,omitempty"`
    ArtifactPath    string          `yaml:"artifact_path,omitempty"`
    Dependencies    []Dependency    `yaml:"dependencies,omitempty"`
}
```

**Migration Strategy**:
- If `template_version` is missing, assume templates need updating
- Set `template_version` on `sl init` and `sl doctor` template update
- Backward compatible (optional field)

**Alternatives Considered**:
- Checksum-based detection only: Less explicit, harder to debug
- Separate template manifest file: More files to manage

### 3. Custom File Detection

**Decision**: SHA-256 checksum comparison

**Algorithm**:
1. Walk `.claude/` directory
2. For each file, compute SHA-256 hash
3. Compare against hash of corresponding embedded file
4. If hashes differ, file is "customized"

**Implementation**:
```go
func IsCustomized(projectPath, relativePath string, embeddedFS embed.FS) (bool, error) {
    // Read project file
    projectContent, err := os.ReadFile(filepath.Join(projectPath, relativePath))
    if err != nil {
        return false, err
    }

    // Read embedded file
    embeddedContent, err := embeddedFS.ReadFile(filepath.Join("skills", relativePath))
    if err != nil {
        // File exists in project but not in embedded - definitely customized
        return true, nil
    }

    // Compare hashes
    projectHash := sha256.Sum256(projectContent)
    embeddedHash := sha256.Sum256(embeddedContent)

    return projectHash != embeddedHash, nil
}
```

**Performance**: SHA-256 of typical template files (< 100KB) is negligible.

**Alternatives Considered**:
- mtime comparison: Unreliable (git checkout changes mtime)
- User confirmation for each file: Too tedious for many files

### 4. Interactive Prompt

**Decision**: Use Bubble Tea TUI (consistent with existing patterns)

**Existing Pattern** (`pkg/cli/tui/`):
- Project already uses Bubble Tea for `sl init` and `sl new`
- `pkg/cli/tui/` has TUI components

**Update Prompt Design**:
```
┌─────────────────────────────────────────────────────────┐
│ Template Update Available                                │
│                                                         │
│ Your templates (v1.0.0) are older than CLI (v1.2.0).    │
│                                                         │
│ Update templates?                                        │
│   > Yes, update templates                               │
│     No, skip for now                                    │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

**Non-Interactive Mode** (`--json`):
- Skip template updates (no prompt)
- Include `template_update_available: true` in JSON output

### 5. Update Instructions

**Decision**: Method-specific instructions based on installation method

**Detection Heuristics**:
1. Check if `sl` is in Homebrew prefix → "Run `brew upgrade specledger`"
2. Check if `sl` is in GOPATH/bin → "Run `go install github.com/specledger/specledger/cmd/sl@latest`"
3. Default → "Download from https://github.com/specledger/specledger/releases/latest"

**Fallback Message**:
```
A new version (v1.2.0) is available!
You are currently on v1.0.0

Update with one of:
  brew upgrade specledger          # Homebrew
  go install .../cmd/sl@latest     # Go install
  https://github.com/.../releases  # Binary download
```

## Open Questions (Resolved)

All clarifications from spec.md have been addressed:
- Template version tracking: specledger.yaml field
- Customized file handling: Skip with summary
- Uncommitted changes: Warn but proceed
- Update flow: Interactive prompt (no flag)
