# Data Model: CLI Unification

**Feature**: CLI Unification (003-cli-unification)
**Date**: 2026-01-30

## Overview

This document defines the data structures and entities for the unified CLI tool. The data model is primarily about the CLI binary itself and its configuration, with minimal in-memory state during operation.

## Key Entities

### CLI Binary

**Description**: The unified executable that provides all SpecLedger functionality (bootstrap TUI + dependency management).

**Fields**:
- `Name`: String - The CLI name (default: `sl`, alias: `specledger`)
- `Version`: String - Semantic version (e.g., `1.0.0`)
- `Platform`: String - Target platform (linux, darwin, windows)
- `Architecture`: String - Target architecture (amd64, arm64)
- `DistributionChannel`: Enum - Distribution method (github-release, self-built, self-hosted, uvx, package-manager)
- `ExecutablePath`: String - Absolute path to the binary
- `InstallLocation`: String - Where the binary is installed (PATH, current dir, etc.)
- `LastUpdated`: DateTime - When the binary was built
- `BuildHash`: String - Git commit hash for reproducible builds

**Relationships**:
- `BootstrapProject` (one binary can bootstrap many projects)
- `DependencySpecification` (binary manages dependencies in projects)

**Validation Rules**:
- `Name` must be non-empty string
- `Version` must follow semantic versioning format (semver)
- `Platform` must be one of: linux, darwin, windows
- `Architecture` must be one of: amd64, arm64

**Lifecycle**:
1. Built via `make build` or GoReleaser
2. Distributed via GitHub releases or self-hosted
3. Installed by user (download, copy, brew install, etc.)
4. Executed to bootstrap projects or manage dependencies
5. No persistent state kept by binary

---

### Bootstrap Project

**Description**: A newly created SpecLedger project containing configuration files for issue tracking, AI agents, and tool management.

**Fields**:
- `ProjectName`: String - Human-readable name (e.g., "my-project")
- `ShortCode`: String - 2-4 character prefix for Beads issues (e.g., "abc")
- `Playbook`: String - Type of project (default, data-science, platform-engineering)
- `AgentShell`: String - AI agent shell (claude-code)
- `DemoDir`: String - Directory where project is created (default: ~/demos)
- `Created`: DateTime - When the project was created
- `ConfigPath`: String - Path to `.specify/memory/constitution.md`
- `BeadsConfigPath`: String - Path to `.beads/config.yaml`
- `TemplatesPath`: String - Path to `.specify/templates/`

**Relationships**:
- `CLI Binary` (one binary can bootstrap many projects)
- `DependencySpecification` (projects may declare external specs)

**Validation Rules**:
- `ProjectName`: alphanumeric, hyphens, underscores only
- `ShortCode`: 2-4 lowercase letters only
- `Playbook`: must be from predefined list
- `AgentShell`: must be from predefined list (currently only "claude-code")

**Lifecycle**:
1. User invokes `sl new` with TUI or flags
2. CLI validates inputs
3. CLI creates directory and copies template files
4. CLI initializes git repository
5. CLI configures Beads and tooling
6. Project is ready for use

**State Transitions**:
```
Initializing → Validating Inputs → Creating Directory → Copying Templates → Configuring Tools → Git Init → Complete
```

---

### Dependency Specification

**Description**: An external specification declared in a project with associated cryptographic verification.

**Fields**:
- `RepositoryURL`: String - URL to the specification repository (e.g., "git@github.com:org/spec.git")
- `Branch`: String - Git branch to use (default: main)
- `SpecPath`: String - Path to the spec file in the repository (e.g., "specs/sdd-control-plane.md")
- `Alias`: String (optional) - Local alias for referencing (e.g., "sdd")
- `Signature`: String - Cryptographic signature of the spec
- `Checksum`: String - File checksum for integrity
- `AddedAt`: DateTime - When the dependency was added
- `UpdatedAt`: DateTime - When the dependency was last verified
- `LockStatus`: Enum - Lock status (locked, unlocked, verification-failed)
- `Status`: Enum - Status (active, removed, failed)

**Relationships**:
- `BootstrapProject` (one project can have many dependencies)
- `DependencySpecification` (dependencies can reference other dependencies transitively)

**Validation Rules**:
- `RepositoryURL`: Must be a valid git repository URL
- `Branch`: Must be a valid git branch name
- `SpecPath`: Must be a valid file path
- `Alias`: 1-20 lowercase alphanumeric characters
- `Signature`: Must be a valid cryptographic signature
- `LockStatus`: Must be one of: locked, unlocked, verification-failed
- `Status`: Must be one of: active, removed, failed

**Lifecycle**:
1. User adds dependency via `sl deps add`
2. CLI fetches spec and verifies signature
3. CLI calculates checksum
4. CLI stores in lockfile (`.specledger/lockfile.yaml`)
5. Dependency is tracked and updated via `sl deps update`

**State Transitions**:
```
Adding → Fetching → Verifying Signature → Calculating Checksum → Storing in Lockfile → Active
```

---

### Dependency Lockfile

**Description**: The lockfile that stores the current state of all dependency specifications in a project.

**Fields**:
- `Version`: String - Lockfile version (e.g., "1.0.0")
- `ProjectName`: String - Name of the project
- `ShortCode`: String - Beads short code for the project
- `Dependencies`: Array<DependencySpecification> - List of declared dependencies
- `CreatedAt`: DateTime - When lockfile was created
- `UpdatedAt`: DateTime - When lockfile was last modified
- `VerificationState`: String - Overall verification state of all dependencies

**Relationships**:
- `BootstrapProject` (one project has one lockfile)

**Validation Rules**:
- `Version`: Must follow semver
- `ProjectName`: Must match project name
- `ShortCode`: Must match project short code
- `Dependencies`: Must be non-empty array of valid DependencySpecification

**Lifecycle**:
1. Created on first `sl deps add` command
2. Updated on `sl deps update` or `sl deps remove`
3. Validated on `sl deps resolve` or `sl deps list`
4. Read on project initialization

**File Location**: `.specledger/lockfile.yaml`

**State Transitions**:
```
Empty → Created (on first add) → Updated (on change) → Validated (on resolve)
```

---

### CLI Configuration

**Description**: Configuration file for the CLI tool itself (user preferences, settings).

**Fields**:
- `DefaultProjectDir`: String - Default directory for new projects (default: ~/demos)
- `PreferredShell`: String - Preferred shell for AI agents (default: bash, zsh)
- `TUIEnabled`: Boolean - Whether TUI is enabled by default (default: true)
- `AutoInstallDependencies`: Boolean - Whether to auto-install missing dependencies (default: false)
- `FallbackToPlainCLI`: Boolean - Whether to fallback to plain CLI when TUI fails (default: true)
- `LogLevel`: String - Logging level (debug, info, warn, error)
- `Theme`: String - TUI theme (default: default)
- `Language`: String - Language for CLI output (en, de, es, etc.)

**Relationships**:
- `CLI Binary` (configuration is per-user, stored in user's home directory)

**Validation Rules**:
- `DefaultProjectDir`: Must be an absolute path
- `PreferredShell`: Must be one of: bash, zsh, fish
- `TUIEnabled`: Boolean
- `AutoInstallDependencies`: Boolean
- `FallbackToPlainCLI`: Boolean
- `LogLevel`: Must be one of: debug, info, warn, error
- `Theme`: Must be one of: default, dark, light
- `Language`: Must be a valid ISO 639-1 language code

**Lifecycle**:
1. Created on first run (user config)
2. Modified by user via configuration file
3. Read by CLI on startup

**File Location**: `~/.config/specledger/config.yaml`

**State Transitions**:
```
Empty → Created → Modified → Loaded on startup
```

---

## Data Flow Diagrams

### Bootstrap Flow

```
User → sl new (TUI or flags)
     → CLI validates inputs
     → CLI checks dependencies (gum, mise)
     → CLI creates directory
     → CLI copies template files
     → CLI initializes git repo
     → CLI configures Beads
     → CLI installs tools (mise)
     → CLI creates first issue
     → User has working project
```

### Dependency Management Flow

```
User → sl deps add <repo>
     → CLI fetches spec from repo
     → CLI verifies signature
     → CLI calculates checksum
     → CLI stores in lockfile
     → User can reference spec in code
```

---

## In-Memory State During Operation

The CLI maintains minimal in-memory state during command execution:

1. **Validation State**:
   - Track which validation checks have passed
   - Store validation errors for reporting

2. **Progress State** (for long-running operations):
   - Current step of operation
   - Progress percentage
   - Current sub-task

3. **User Input State** (for non-interactive bootstrap):
   - Project name
   - Short code
   - Playbook selection
   - Agent shell selection

4. **Dependency Resolution State**:
   - Currently resolving dependency
   - Transitive dependency tracking
   - Verification status

This state is ephemeral (cleared after command completes) and not persisted.

---

## Error States

### CLI Execution Errors

| Error Type | Condition | User Impact |
|------------|-----------|-------------|
| `ErrCommandNotFound` | Invalid command | Show help with list of valid commands |
| `ErrDependencyMissing` | gum/mise not found | Prompt user to install or continue without TUI |
| `ErrPermissionDenied` | Cannot write to directory | Show error with alternative location suggestion |
| `ErrInvalidInput` | Invalid command-line flags | Show error with correct usage |
| `ErrProjectExists` | Bootstrap in existing project | Prompt user to confirm or choose different name |
| `ErrTemplateNotFound` | Template file missing | Show error and stop execution |
| `ErrNetworkError` | Cannot fetch spec | Show error with retry suggestion |
| `ErrSignatureInvalid` | Spec signature verification failed | Show error and mark as failed status |

### Exit Codes

| Exit Code | Meaning |
|-----------|---------|
| 0 | Success |
| 1 | Any error occurred |
| 130 | User cancellation (Ctrl+C) |

---

## Summary

This data model defines:

1. **CLI Binary** - The executable artifact and its properties
2. **Bootstrap Project** - Structure of newly created SpecLedger projects
3. **Dependency Specification** - External specs with cryptographic verification
4. **Dependency Lockfile** - State management for dependencies
5. **CLI Configuration** - User preferences for CLI behavior

The model is minimal and focused, with clear validation rules and lifecycle transitions. No complex data storage is required as the CLI is a stateless tool that works with existing project files.
