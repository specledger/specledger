# Research: Embedded Templates

**Feature**: 005-embedded-templates
**Phase**: 0 - Research & Technical Decisions
**Date**: 2026-02-07

## Overview

This document captures research findings and technical decisions for implementing embedded template support in SpecLedger. All technical questions from the plan have been resolved.

## Decision 1: Template Storage Mechanism

**Question**: How should templates be embedded in the SpecLedger binary?

### Options Evaluated

| Option | Description | Pros | Cons |
|--------|-------------|------|------|
| A: `embed` package (Go 1.16+) | Compile templates into binary | Single binary distribution, faster access, no external dependencies | Binary size increases (~50KB), requires rebuild for template changes |
| B: Runtime file path | Read from `templates/` folder at install location | Easy template updates, smaller binary | Requires templates folder in correct location, deployment complexity |
| C: Hybrid | Check embed first, fallback to files | Maximum flexibility | More complex logic, two paths to test |

### Decision: **Option A - `embed` package**

**Rationale**:
1. **Single Binary Distribution**: SpecLedger is distributed as a single binary (via GoReleaser, Homebrew, npm). Embedded templates maintain this simplicity.
2. **Fast Access**: No filesystem lookup overhead - templates are in memory.
3. **Reliability**: No issues with missing template folders or incorrect installation paths.
4. **Binary Size**: Templates are small text files (~50KB total) - negligible impact on binary size.
5. **Future Remote Templates**: The architecture can support remote templates by adding a fetch function alongside the embedded provider.

**Implementation**:
```go
// pkg/embedded/templates.go
package embedded

import "embed"

//go:embed templates/*
var TemplatesFS embed.FS
```

## Decision 2: Template Copying Strategy

**Question**: How should templates be copied to the target project?

### Options Evaluated

| Option | Description | Pros | Cons |
|--------|-------------|------|------|
| A: Copy all | Copy entire templates folder | Simple, no filtering needed | May copy unwanted files |
| B: Framework-specific | Copy only based on `--framework` flag | Targeted, smaller footprint | Requires mapping framework to files |
| C: Manifest-driven | Use manifest file to specify what to copy | Flexible, per-framework control | Requires maintaining manifest files |

### Decision: **Option B - Framework-specific with fallback**

**Rationale**:
1. **User Intent**: When user specifies `--framework speckit`, they expect Spec Kit templates only.
2. **Cleaner Projects**: Users don't want OpenSpec templates cluttering their Spec Kit project.
3. **Future-Proof**: Easy to add new frameworks by creating new framework → file mappings.
4. **Fallback to Copy All**: If `--framework` is `none` or `both`, copy all templates.

**Implementation**:
```go
// pkg/cli/templates/templates.go
type FrameworkTemplate struct {
    Framework string
    Patterns  []string  // Glob patterns for files to include
}

var frameworkTemplates = map[string]FrameworkTemplate{
    "speckit": {
        Patterns: []string{
            "specledger/**",
            ".claude/commands/specledger.adopt.md",
            ".claude/commands/specledger.specify.md",
            // ... other Spec Kit specific files
        },
    },
    "openspec": {
        Patterns: []string{
            // Future: OpenSpec specific files
        },
    },
}
```

## Decision 3: Handling Existing Files

**Question**: What should happen when template files already exist in the target project?

### Options Evaluated

| Option | Description | Pros | Cons |
|--------|-------------|------|------|
| A: Skip all | Don't overwrite any existing files | Safe, preserves user changes | Template updates never applied |
| B: Overwrite all | Replace all existing files | Templates always up-to-date | Destroys user customizations |
| C: Smart merge | Check if file differs, ask or skip | Best of both worlds | More complex, requires diff logic |
| D: Skip with warning | Skip existing files, log warning | Safe, informative | May miss important updates |

### Decision: **Option D - Skip with warning (current phase)**

**Rationale**:
1. **User Safety**: Preserve user's existing files and customizations.
2. **Simplicity**: No complex merge logic for MVP.
3. **Transparency**: Warnings inform users what was skipped.
4. **Future Enhancement**: Can add `--force-templates` flag for overwrite behavior later.

**Implementation**:
```go
if fileExists(targetPath) {
    ui.PrintWarning(fmt.Sprintf("Skipped existing file: %s", targetPath))
    continue
}
```

## Decision 4: Template Discovery Mechanism

**Question**: How should SpecLedger discover available embedded templates?

### Options Evaluated

| Option | Description | Pros | Cons |
|--------|-------------|------|------|
| A: Hardcoded list | List templates in code | Simple, type-safe | Requires code changes for new templates |
| B: Directory scanning | Scan embedded FS for subdirectories | Automatic discovery | May find non-template directories |
| C: Manifest file | Read `templates/manifest.yaml` | Flexible, metadata-rich | Requires maintaining manifest |

### Decision: **Option C - Manifest file (with B as fallback)**

**Rationale**:
1. **Metadata**: Manifest can store template name, description, framework type, version.
2. **Future-Proof**: Easy to add remote templates by fetching their manifest.
3. **Fallback**: If manifest missing, scan directories for basic functionality.
4. **UX**: `sl template list` can show rich information from manifest.

**Manifest Structure**:
```yaml
# templates/manifest.yaml
version: "1.0"
templates:
  - name: speckit
    description: "Spec Kit playbook templates"
    framework: "speckit"
    version: "1.0.0"
    path: "specledger"
  - name: openspec
    description: "OpenSpec playbook templates (future)"
    framework: "openspec"
    version: "0.1.0"
    path: "openspec"
```

## Decision 5: File Copying Implementation

**Question**: What library/method to use for copying files and directories?

### Options Evaluated

| Option | Description | Pros | Cons |
|--------|-------------|------|------|
| A: `io.Copy` | Manual implementation | Full control, no dependencies | More code, handle edge cases manually |
| B: `os.ReadFile` + `os.WriteFile` | Simple file operations | Simple for single files | Manual directory recursion |
| C: `embed.FS` + Walk pattern | Use embed.Walk to iterate | Built-in, no external deps | Manual copy implementation |

### Decision: **Option C - `embed.FS.Walk` pattern**

**Rationale**:
1. **No External Dependencies**: Uses only Go standard library.
2. **Works with embed.FS**: Direct integration with Decision 1.
3. **Pattern Established**: Common pattern in Go ecosystem for embedded files.
4. **Cross-Platform**: Works on all platforms SpecLedger supports.

**Implementation**:
```go
func CopyTemplates(fsys fs.FS, srcDir, destDir string, patterns []string) error {
    return fs.WalkDir(fsys, srcDir, func(path string, d fs.DirEntry, err error) error {
        // Copy logic here
    })
}
```

## Decision 6: Integration Points

**Question**: Where in the existing code should template copying be integrated?

### Research: Existing Bootstrap Flow

After analyzing the existing codebase:
- `pkg/cli/commands/bootstrap.go` - Main entry point for `sl new`
- `pkg/cli/commands/bootstrap_helpers.go` - Helper functions for project creation
- `pkg/cli/commands/bootstrap_init.go` - Handles `sl init`

**Integration Points**:
1. **After project directory creation** in `bootstrap.go` - Copy templates before framework initialization
2. **After specledger.yaml creation** in `bootstrap_init.go` - Copy templates after metadata setup

### Decision: **Add to `bootstrap_helpers.go`**

**Rationale**:
1. **Centralized**: `bootstrap_helpers.go` already has helper functions for bootstrap
2. **Reusability**: Both `sl new` and `sl init` call the same helpers
3. **Testability**: Helper functions can be tested independently

**Integration Point**:
```go
// In bootstrap_helpers.go
func applyEmbeddedTemplates(projectPath, framework string) error {
    // Call template package to copy templates
    return templates.ApplyToProject(projectPath, framework)
}
```

## Summary of Technical Decisions

| Decision | Choice | Key Rationale |
|----------|--------|---------------|
| Template Storage | Go `embed` package | Single binary distribution, reliable |
| Copying Strategy | Framework-specific with patterns | Targeted, future-proof |
| Existing Files | Skip with warning | User safety, transparency |
| Template Discovery | Manifest file (with scan fallback) | Metadata-rich, supports remote templates |
| File Copying | `embed.FS.Walk` pattern | No dependencies, cross-platform |
| Integration Point | `bootstrap_helpers.go` | Centralized, reusable |

## Architecture for Future Remote Templates

The decisions above support the future requirement for remote templates:

```go
// Future interface for template sources
type TemplateSource interface {
    List() ([]Template, error)
    Copy(name, destDir string) error
}

// Embedded implementation (current)
type EmbeddedSource struct {
    fsys embed.FS
}

// Remote implementation (future)
type RemoteSource struct {
    baseURL string
    client  *http.Client
}
```

This design allows adding remote templates without refactoring the core logic.

## References

- Go embed package: https://pkg.go.dev/embed
- Go fs.WalkDir: https://pkg.go.dev/io/fs#WalkDir
- Existing SpecLedger code: `pkg/cli/commands/` and `pkg/cli/metadata/`
