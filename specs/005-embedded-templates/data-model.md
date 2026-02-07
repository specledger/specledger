# Data Model: Embedded Templates

**Feature**: 001-embedded-templates
**Phase**: 1 - Design & Contracts
**Date**: 2026-02-07

## Overview

This document defines the data structures and entities for the embedded templates feature. The data model is designed to support both embedded and future remote template sources.

## Core Entities

### 1. TemplateSource

Represents a source of templates (embedded or remote).

```go
// pkg/cli/templates/source.go
type TemplateSource interface {
    // List returns all available templates from this source
    List() ([]Template, error)

    // Copy copies the specified template to the destination directory
    Copy(name string, destDir string, opts CopyOptions) error

    // Exists checks if a template exists in this source
    Exists(name string) bool
}
```

### 2. Template

Represents a single template/playbook that can be applied to a project.

```go
// pkg/cli/templates/template.go
type Template struct {
    // Name is the identifier for this template (e.g., "speckit", "openspec")
    Name string `yaml:"name" json:"name"`

    // Description explains what this template provides
    Description string `yaml:"description" json:"description"`

    // Framework is the SDD framework type (speckit, openspec, both, none)
    Framework string `yaml:"framework" json:"framework"`

    // Version is the template version
    Version string `yaml:"version" json:"version"`

    // Path is the relative path within the templates folder
    Path string `yaml:"path" json:"path"`

    // Patterns are glob patterns for files to include (optional, defaults to all)
    Patterns []string `yaml:"patterns,omitempty" json:"patterns,omitempty"`
}
```

### 3. TemplateManifest

The manifest file that lists all available templates.

```go
// pkg/cli/templates/manifest.go
type TemplateManifest struct {
    // Version is the manifest format version
    Version string `yaml:"version"`

    // Templates is the list of available templates
    Templates []Template `yaml:"templates"`
}
```

### 4. CopyOptions

Options for controlling template copying behavior.

```go
// pkg/cli/templates/copy.go
type CopyOptions struct {
    // DryRun if true, shows what would be copied without actually copying
    DryRun bool

    // Overwrite if true, overwrites existing files
    Overwrite bool

    // SkipExisting if true, skips existing files (default: true)
    SkipExisting bool

    // Verbose if true, prints each file being copied
    Verbose bool

    // Framework filters templates by framework type (optional)
    Framework string
}
```

### 5. CopyResult

Result of a template copy operation.

```go
// pkg/cli/templates/copy.go
type CopyResult struct {
    // FilesCopied is the count of files successfully copied
    FilesCopied int

    // FilesSkipped is the count of files skipped (already existed)
    FilesSkipped int

    // Errors is a list of errors encountered during copying
    Errors []CopyError

    // Duration is how long the copy operation took
    Duration time.Duration
}
```

### 6. CopyError

An error that occurred during copying with context.

```go
// pkg/cli/templates/copy.go
type CopyError struct {
    // Path is the file path where the error occurred
    Path string

    // Err is the underlying error
    Err error

    // IsWarning indicates if this is a non-fatal warning
    IsWarning bool
}
```

## Data Relationships

```
TemplateSource (interface)
    ├── EmbeddedSource (implements TemplateSource)
    │   ├── embed.FS (Go embedded filesystem)
    │   └── TemplateManifest
    │       └── []Template
    │
    └── RemoteSource (future)
        ├── HTTP Client
        └── TemplateManifest (fetched from URL)
            └── []Template

Template
    ├── Name (unique identifier)
    ├── Framework (speckit/openspec/both/none)
    ├── Path (source location)
    └── Patterns (file selection)

CopyOperation
    ├── Template (what to copy)
    ├── DestDir (where to copy)
    ├── Options (how to copy)
    └── Result (what happened)
```

## State Transitions

### Template Lifecycle

```
[Embedded in Binary] → [Discovered via List()] → [Selected by User]
                                                    ↓
                                    [Copied to Project Directory]
                                                    ↓
                                    [Immutable in Project]
                                                    ↓
                        [User May Modify] (SpecLedger doesn't track)
```

### Copy Operation Flow

```
[Start Copy]
      ↓
[Validate Template Exists]
      ↓
[Resolve File Patterns]
      ↓
[For Each File]
      ↓
  [File Exists?] ──Yes──→ [Skip or Warn] ──┐
      │ No                                │
      ↓                                   │
[Copy File]                               │
      ↓                                   │
[Record Result] ←─────────────────────────┘
      ↓
[Return CopyResult]
```

## Validation Rules

### Template Validation

| Field | Validation | Error Message |
|-------|-----------|---------------|
| Name | Required, alphanumeric with hyphens | "template name must be alphanumeric" |
| Framework | Must be: speckit, openspec, both, or none | "invalid framework type" |
| Path | Must exist in source | "template path not found" |
| Patterns | Valid glob patterns | "invalid glob pattern" |

### Copy Validation

| Check | Validation | Error Message |
|-------|-----------|---------------|
| DestDir | Must be writable directory | "destination directory not writable" |
| Template | Must exist in source | "template not found" |
| Disk Space | Sufficient space for copy | "insufficient disk space" |
| Permissions | Read permission on source, write on dest | "permission denied" |

## Storage Considerations

### Embedded Templates (Current)

- **Location**: Compiled into binary via `embed` package
- **Format**: Original file structure preserved in `embed.FS`
- **Size**: ~50KB for Spec Kit templates
- **Access**: Read-only, in-memory

### Remote Templates (Future)

- **Location**: Git repository or HTTPS URL
- **Format**: Tarball or git clone
- **Caching**: Local cache at `~/.specledger/template-cache/`
- **Access**: Download on first use, cached locally

### Project Templates (After Copy)

- **Location**: User's project directory
- **Format**: Regular files on disk
- **Ownership**: User owns and may modify
- **Lifecycle**: SpecLedger does not track or update after copy

## Edge Case Handling

### Missing Manifest File

If `templates/manifest.yaml` doesn't exist:
1. Scan `templates/` directory for subdirectories
2. Create synthetic Template entries with default values
3. Log warning about missing manifest

### Conflicting Files

When a template file already exists in the target project:
1. Default: Skip file with warning
2. If `--force-templates` flag: Overwrite file
3. Log all skipped/overwritten files

### Invalid Template Path

If a template's path in manifest doesn't exist:
1. Return error during validation
2. User must update manifest or fix template structure
3. List operation skips invalid templates

## Performance Considerations

### Memory Usage

- **Embedded**: Templates loaded into memory at startup (~50KB)
- **Copy Operation**: Streaming copy, minimal memory overhead
- **Large Templates**: Not anticipated (all text files)

### Copy Performance

- **Target**: < 1 second for typical template set
- **Optimization**: Parallel file copy for large template sets (future)
- **Monitoring**: Track CopyResult.Duration

## Future Extensions

### Template Variables

Support for variable substitution in templates (deferred):
```go
type TemplateVariables map[string]string

// In Template:
Variables map[string]string `yaml:"variables,omitempty"`

// Example:
// {{ .ProjectName }} → "myproject"
// {{ .ShortCode }} → "mp"
```

### Template Inheritance

Support for templates extending other templates (deferred):
```yaml
# Template can extend another
extends: "base-template"

# Override specific files
overrides:
  - "templates/spec-template.md"
```

### Template Versioning

Support for semantic versioning and updates (deferred):
```go
type TemplateVersion struct {
    Version    string
    UpdatedAt  time.Time
    Changelog  string
}
```
