# Data Model: SpecLedger Thin Wrapper Architecture

**Date**: 2026-02-05
**Feature**: 004-thin-wrapper-redesign

## Overview

This document defines the data structures for SpecLedger's YAML-based metadata system, prerequisite checking, and framework detection.

## Core Entities

### 1. Project Metadata

**Storage**: `specledger/specledger.yaml`
**Purpose**: Project configuration and dependency tracking

#### Fields

| Field | Type | Required | Description | Validation Rules |
|-------|------|----------|-------------|------------------|
| `version` | string | Yes | Schema version (semver) | Must be "1.0.0" for this design |
| `project.name` | string | Yes | Project name | Non-empty, alphanumeric + hyphens |
| `project.short_code` | string | Yes | Beads issue prefix | 2-10 characters, alphanumeric |
| `project.created` | timestamp | Yes | Project creation time | ISO8601 format |
| `project.modified` | timestamp | Yes | Last modification time | ISO8601 format, >= created |
| `project.version` | string | Yes | Project version | Semver format |
| `framework.choice` | enum | Yes | Selected SDD framework | One of: `speckit`, `openspec`, `both`, `none` |
| `framework.installed_at` | timestamp | No | Framework installation time | ISO8601 format |
| `dependencies` | array | No | List of spec dependencies | Can be empty array |

#### Relationships

- A Project has 0 or more Dependencies
- A Project has exactly 1 Framework choice

#### Example

```yaml
version: "1.0.0"

project:
  name: my-api
  short_code: mapi
  created: "2026-02-05T10:30:00Z"
  modified: "2026-02-05T10:30:00Z"
  version: "0.1.0"

framework:
  choice: speckit
  installed_at: "2026-02-05T10:30:00Z"

dependencies: []
```

### 2. Dependency Entry

**Storage**: Within `specledger/specledger.yaml` under `dependencies` array
**Purpose**: Track external spec dependencies with lockfile behavior

#### Fields

| Field | Type | Required | Description | Validation Rules |
|-------|------|----------|-------------|------------------|
| `url` | string | Yes | Git repository URL | Valid git URL (ssh or https) |
| `branch` | string | No | Branch or tag name | Defaults to "main" |
| `path` | string | No | File path within repo | Defaults to "spec.md" |
| `alias` | string | No | Short reference name | Alphanumeric + hyphens |
| `resolved_commit` | string | No | SHA hash of resolved commit | 40-character hex string |

#### Relationships

- A Dependency belongs to exactly 1 Project
- A Dependency references exactly 1 external Git repository

#### State Transitions

```
[Declared] --resolve--> [Resolved]
    |                        |
    |                        |
    +--update--------> [Resolved with new commit]
```

#### Example

```yaml
dependencies:
  - url: git@github.com:org/auth-spec
    branch: v1.0
    path: specs/authentication.md
    alias: auth
    resolved_commit: abc123def456...

  - url: https://github.com/org/api-spec
    # branch defaults to "main"
    # path defaults to "spec.md"
    alias: upstream-api
    resolved_commit: 789xyz012...
```

### 3. Tool Status

**Storage**: In-memory only (not persisted)
**Purpose**: Track installation status of required and optional tools for `sl doctor` command

#### Fields

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Tool identifier (mise, bd, perles, specify, openspec) |
| `installed` | boolean | Whether tool is found in PATH |
| `version` | string | Version string from `--version` output (if installed) |
| `path` | string | Full path to executable (if installed) |
| `category` | enum | One of: `core`, `framework` |

#### Relationships

- Tool Status is computed at runtime (not stored)
- Derived from `exec.LookPath()` and `--version` checks

#### Example (JSON output from `sl doctor`)

```json
{
  "status": "pass",
  "tools": [
    {
      "name": "mise",
      "installed": true,
      "version": "v2024.1.0",
      "path": "/usr/local/bin/mise",
      "category": "core"
    },
    {
      "name": "bd",
      "installed": true,
      "version": "0.28.0",
      "path": "/Users/user/.local/share/mise/installs/ubi-steveyegge-beads/0.28.0/bin/bd",
      "category": "core"
    },
    {
      "name": "specify",
      "installed": false,
      "version": "",
      "path": "",
      "category": "framework"
    }
  ]
}
```

### 4. Framework Choice

**Storage**: Within `specledger/specledger.yaml` under `framework.choice`
**Purpose**: Document which SDD framework(s) the user intends to use

#### Valid Values

| Value | Meaning | Use Case |
|-------|---------|----------|
| `speckit` | GitHub Spec Kit only | User prefers structured, phase-gated workflow |
| `openspec` | OpenSpec only | User prefers lightweight, iterative workflow |
| `both` | Both frameworks | User wants flexibility to use either approach |
| `none` | No SDD framework | User only wants SpecLedger for dependencies/bootstrap |

#### Behavior

- This field is **documentation only** - SpecLedger does not enforce framework usage
- Recorded during `sl new` based on user selection
- Can be changed manually by editing YAML file
- `sl doctor` reports which frameworks are actually installed (may differ from choice)

## Entity Lifecycle

### Project Metadata Lifecycle

1. **Creation**: `sl new` or `sl init` creates `specledger/specledger.yaml`
2. **Modification**: `sl deps add/remove` updates `dependencies` array
3. **Migration**: `sl migrate` converts existing `.mod` file to `.yaml`
4. **Reading**: All commands read metadata on startup

### Dependency Lifecycle

1. **Declaration**: `sl deps add <url>` adds dependency without resolving
2. **Resolution**: `sl deps resolve` clones repo and records commit SHA
3. **Update**: `sl deps update` fetches latest commit and updates SHA
4. **Removal**: `sl deps remove <url/alias>` deletes from array

### Tool Status Lifecycle

1. **Detection**: `sl doctor` or prerequisite check queries PATH and `--version`
2. **Installation**: User runs `mise install` (external to SpecLedger)
3. **Re-detection**: Next command run picks up newly installed tools

## Validation Rules

### Project Name Validation

```go
func ValidateProjectName(name string) error {
    if len(name) == 0 {
        return errors.New("project name cannot be empty")
    }
    if !regexp.MustCompile(`^[a-zA-Z0-9-]+$`).MatchString(name) {
        return errors.New("project name must contain only alphanumeric characters and hyphens")
    }
    return nil
}
```

### Short Code Validation

```go
func ValidateShortCode(code string) error {
    if len(code) < 2 || len(code) > 10 {
        return errors.New("short code must be 2-10 characters")
    }
    if !regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString(code) {
        return errors.New("short code must contain only alphanumeric characters")
    }
    return nil
}
```

### Git URL Validation

```go
func ValidateGitURL(url string) error {
    sshPattern := `^git@[^:]+:[^/]+/.+\.git$|^git@[^:]+:[^/]+/[^/]+$`
    httpsPattern := `^https://[^/]+/[^/]+/.+$`

    if regexp.MustCompile(sshPattern).MatchString(url) {
        return nil
    }
    if regexp.MustCompile(httpsPattern).MatchString(url) {
        return nil
    }

    return errors.New("url must be valid git SSH or HTTPS URL")
}
```

### Commit SHA Validation

```go
func ValidateCommitSHA(sha string) error {
    if len(sha) != 40 {
        return errors.New("commit SHA must be 40 characters")
    }
    if !regexp.MustCompile(`^[a-f0-9]+$`).MatchString(sha) {
        return errors.New("commit SHA must contain only hexadecimal characters")
    }
    return nil
}
```

## Migration from .mod Format

### .mod Format (Legacy)

```
# SpecLedger Dependency Manifest v1.0.0
# Generated by sl init on 2026-02-05
# Project: specledger
# Short Code: sl
#
# To add dependencies, use:
#   sl deps add git@github.com:org/spec main spec.md --alias alias

# Dependencies are listed below (none yet)
```

### YAML Format (New)

```yaml
version: "1.0.0"

project:
  name: specledger
  short_code: sl
  created: "2026-02-05T10:30:00Z"
  modified: "2026-02-05T10:30:00Z"
  version: "0.1.0"

framework:
  choice: none
  # No installed_at because .mod predates framework tracking

dependencies: []
```

### Migration Rules

1. Parse .mod file to extract: project name, short code
2. Set `framework.choice` to `none` (default for migrated projects)
3. Set `created` timestamp to .mod file creation time
4. Set `modified` timestamp to current time
5. Set `version` to "0.1.0" (default for migrated projects)
6. Parse any dependency comments (format TBD - .mod doesn't store dependencies yet)
7. Write YAML file alongside .mod (don't delete .mod)
8. Print success message with path to new YAML file

## Storage Patterns

### File Locations

- **Project metadata**: `<project-root>/specledger/specledger.yaml`
- **Dependency cache**: `~/.specledger/cache/<domain>/<org>/<repo>/<commit>/`
- **Lockfile** (future): `<project-root>/specledger/specledger.lock` (if needed)

### Cache Directory Structure

```
~/.specledger/
└── cache/
    └── github.com/
        └── org/
            └── repo/
                └── abc123.../
                    ├── spec.md
                    ├── plan.md
                    └── ...
```

## Go Struct Definitions

```go
// ProjectMetadata represents specledger.yaml
type ProjectMetadata struct {
    Version      string            `yaml:"version"`
    Project      ProjectInfo       `yaml:"project"`
    Framework    FrameworkInfo     `yaml:"framework"`
    Dependencies []Dependency      `yaml:"dependencies,omitempty"`
}

// ProjectInfo contains project identification
type ProjectInfo struct {
    Name      string    `yaml:"name"`
    ShortCode string    `yaml:"short_code"`
    Created   time.Time `yaml:"created"`
    Modified  time.Time `yaml:"modified"`
    Version   string    `yaml:"version"`
}

// FrameworkInfo records SDD framework choice
type FrameworkInfo struct {
    Choice      FrameworkChoice  `yaml:"choice"`
    InstalledAt *time.Time       `yaml:"installed_at,omitempty"`
}

// FrameworkChoice is an enum
type FrameworkChoice string

const (
    FrameworkSpecKit   FrameworkChoice = "speckit"
    FrameworkOpenSpec  FrameworkChoice = "openspec"
    FrameworkBoth      FrameworkChoice = "both"
    FrameworkNone      FrameworkChoice = "none"
)

// Dependency represents an external spec dependency
type Dependency struct {
    URL            string  `yaml:"url"`
    Branch         string  `yaml:"branch,omitempty"`
    Path           string  `yaml:"path,omitempty"`
    Alias          string  `yaml:"alias,omitempty"`
    ResolvedCommit string  `yaml:"resolved_commit,omitempty"`
}

// ToolStatus represents runtime tool detection (not persisted)
type ToolStatus struct {
    Name      string
    Installed bool
    Version   string
    Path      string
    Category  ToolCategory
}

// ToolCategory is an enum
type ToolCategory string

const (
    ToolCategoryCore      ToolCategory = "core"
    ToolCategoryFramework ToolCategory = "framework"
)
```
