# Feature Specification: SpecLedger Thin Wrapper Architecture

**Feature Branch**: `004-thin-wrapper-redesign`
**Created**: 2026-02-05
**Status**: Draft
**Input**: User description: "Redesign SpecLedger as a thin wrapper CLI that orchestrates tool installation and project bootstrapping, while delegating SDD workflows to user-chosen frameworks (Spec Kit or OpenSpec). Remove duplicate SDD commands, implement prerequisite checking with auto-installation via mise, and support multiple optional SDD frameworks. Use YAML for SpecLedger metadata instead of .mod files to contain dependency, SDD framework used, etc."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Install and Bootstrap New Project (Priority: P1)

As a developer starting a new project, I want to quickly bootstrap a project with my preferred SDD framework so that I can begin spec-driven development immediately.

**Why this priority**: This is the core value proposition of SpecLedger - simplifying project setup. Without this, SpecLedger has no purpose.

**Independent Test**: Can be fully tested by running `sl new`, selecting a framework, and verifying all tools are installed and project structure is created. Delivers immediate value by providing a working development environment.

**Acceptance Scenarios**:

1. **Given** a developer has installed SpecLedger, **When** they run `sl new` interactively, **Then** they are prompted to select an SDD framework (Spec Kit, OpenSpec, Both, or None)
2. **Given** the user selects an SDD framework, **When** the bootstrap completes, **Then** all required tools (mise, beads, perles) and the selected framework(s) are installed and configured
3. **Given** a CI/CD pipeline, **When** it runs `sl new --ci --project-name test --short-code t`, **Then** the project is created non-interactively with default settings
4. **Given** prerequisite tools are missing, **When** the user runs `sl new`, **Then** they are prompted to automatically install missing tools via mise

---

### User Story 2 - Check Tool Installation Status (Priority: P1)

As a developer, I want to verify that all required tools are properly installed so that I can troubleshoot setup issues before starting development.

**Why this priority**: Without a diagnostic tool, users will struggle to understand why commands fail. This is critical for successful onboarding.

**Independent Test**: Can be fully tested by running `sl doctor` and verifying it correctly reports the status of all tools. Delivers value by providing clear visibility into the development environment state.

**Acceptance Scenarios**:

1. **Given** all tools are installed, **When** the user runs `sl doctor`, **Then** a success report shows all tools with their versions
2. **Given** some tools are missing, **When** the user runs `sl doctor`, **Then** the report clearly indicates which tools are missing and provides installation instructions
3. **Given** multiple SDD frameworks are available, **When** the user runs `sl doctor`, **Then** the report shows which frameworks are installed (Spec Kit, OpenSpec, or both)

---

### User Story 3 - Manage Spec Dependencies (Priority: P2)

As a developer working on a project that extends external specifications, I want to declare and resolve spec dependencies so that AI agents can reference upstream specs during development.

**Why this priority**: Dependency management is a unique value-add of SpecLedger, but projects can function without it initially. It becomes important for larger projects with shared specifications.

**Independent Test**: Can be fully tested by adding, listing, and resolving dependencies using `sl deps` commands. Delivers value by enabling spec composition and reuse.

**Acceptance Scenarios**:

1. **Given** a SpecLedger project, **When** the user runs `sl deps add git@github.com:org/spec`, **Then** the dependency is recorded in the project metadata YAML
2. **Given** dependencies are declared, **When** the user runs `sl deps resolve`, **Then** all dependencies are fetched and cached locally
3. **Given** a project with dependencies, **When** an AI agent needs context, **Then** it can reference cached spec files from the dependency cache

---

### User Story 4 - Initialize SpecLedger in Existing Project (Priority: P2)

As a developer with an existing codebase, I want to add SpecLedger infrastructure to my current project so that I can adopt spec-driven development without creating a new repository.

**Why this priority**: Many users will want to adopt SpecLedger incrementally rather than starting fresh. This enables gradual migration.

**Independent Test**: Can be fully tested by running `sl init` in an existing directory and verifying SpecLedger files are created without disrupting existing code. Delivers value by lowering the barrier to adoption.

**Acceptance Scenarios**:

1. **Given** an existing Git repository, **When** the user runs `sl init`, **Then** SpecLedger directories (.beads, specledger) are created without modifying existing files
2. **Given** prerequisite tools are missing, **When** the user runs `sl init`, **Then** they are prompted to install missing tools
3. **Given** a specledger metadata YAML already exists, **When** the user runs `sl init --force`, **Then** the existing configuration is preserved or merged with defaults

---

### User Story 5 - Use Preferred SDD Framework Transparently (Priority: P1)

As a developer, I want to use my chosen SDD framework's commands (Spec Kit or OpenSpec) without SpecLedger interfering, so that I can follow the standard workflow of my preferred framework.

**Why this priority**: The core architectural decision is to delegate SDD workflows. If SpecLedger interferes or duplicates commands, the redesign fails its primary goal.

**Independent Test**: Can be fully tested by installing a framework via mise and verifying its native commands work without SpecLedger-specific wrappers. Delivers value by maintaining tool neutrality.

**Acceptance Scenarios**:

1. **Given** a user has installed Spec Kit, **When** they run `/speckit.specify`, **Then** the native Spec Kit command executes without SpecLedger modification
2. **Given** a user has installed OpenSpec, **When** they run `/opsx:new`, **Then** the native OpenSpec command executes without SpecLedger modification
3. **Given** a user has both frameworks installed, **When** they check available Claude commands, **Then** they see both `/speckit.*` and `/opsx.*` command sets without duplication

---

### Edge Cases

- What happens when a user tries to run `sl new` without mise installed? → Clear error message with installation instructions
- What happens when a user selects a framework but the installation fails? → Bootstrap is rolled back, error is displayed with troubleshooting steps
- What happens when a user runs `sl init` in a directory that already has SpecLedger? → Error message or confirmation prompt to overwrite
- What happens when mise.toml is manually edited incorrectly? → `sl doctor` detects and reports the issue
- What happens when a user wants to add a second framework after initial bootstrap? → They can uncomment the framework in mise.toml and run `mise install`
- What happens when the specledger.yaml is corrupted? → Commands that read metadata provide clear error messages and suggest repair steps
- What happens when dependencies reference unreachable Git repositories? → `sl deps resolve` reports the failure with the specific dependency that failed

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: SpecLedger CLI MUST provide a `new` command that bootstraps a project with user-selected SDD framework (Spec Kit, OpenSpec, Both, or None)
- **FR-002**: SpecLedger CLI MUST provide an `init` command that initializes SpecLedger infrastructure in an existing repository
- **FR-003**: SpecLedger CLI MUST provide a `deps` command group for managing spec dependencies (add, list, remove, resolve)
- **FR-004**: SpecLedger CLI MUST provide a `doctor` command that checks installation status of all required and optional tools
- **FR-005**: SpecLedger MUST check for mise installation before executing bootstrap commands
- **FR-006**: SpecLedger MUST prompt users to install missing core tools (mise, beads, perles) during bootstrap
- **FR-007**: SpecLedger MUST support interactive mode (TUI) and non-interactive mode (CI/CD) for all commands
- **FR-008**: SpecLedger MUST use YAML format for project metadata instead of .mod files
- **FR-009**: SpecLedger MUST remove all duplicate SDD workflow commands (specify, plan, tasks, implement, analyze, clarify, checklist, constitution)
- **FR-010**: SpecLedger MUST retain only SpecLedger-specific commands (deps, adopt, resume) in Claude command templates
- **FR-011**: SpecLedger MUST allow users to manually select SDD frameworks by editing mise.toml and running `mise install`
- **FR-012**: SpecLedger MUST support both Spec Kit and OpenSpec as optional, interchangeable frameworks
- **FR-013**: SpecLedger MUST store framework choice in project metadata YAML for documentation purposes
- **FR-014**: SpecLedger MUST cache resolved dependencies locally at `~/.specledger/cache/`
- **FR-015**: SpecLedger MUST verify tool installation via version checks (`--version` flags)
- **FR-016**: mise.toml configuration files MUST include comments explaining how to enable optional frameworks
- **FR-017**: SpecLedger MUST support upgrading existing .mod files to YAML format via migration command
- **FR-018**: Project metadata YAML MUST include fields for: project name, short code, dependencies, framework choice, created date, and version

### Key Entities *(include if feature involves data)*

- **Project Metadata**: Represents a SpecLedger project configuration stored in `specledger/specledger.yaml`, including project name, short code, dependency list, selected SDD framework(s), creation timestamp, and SpecLedger version
- **Dependency**: Represents an external spec dependency with fields for Git URL, branch/tag, file path, alias, and resolved commit hash
- **Tool Status**: Represents the installation state of a development tool (mise, beads, perles, specify, openspec) including presence check, version information, and installation path
- **Framework Choice**: Enumeration of SDD framework options (speckit, openspec, both, none) recorded in project metadata for documentation

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Developers can bootstrap a new SpecLedger project in under 3 minutes including automatic tool installation
- **SC-002**: `sl doctor` command executes in under 2 seconds and provides actionable diagnostics
- **SC-003**: Developers can switch between Spec Kit and OpenSpec frameworks without recreating their project
- **SC-004**: 100% of duplicate SDD commands are removed from SpecLedger command templates (specledger.specify, specledger.plan, etc.)
- **SC-005**: Project metadata migration from .mod to YAML completes without data loss for all existing projects
- **SC-006**: CI/CD pipelines can bootstrap projects non-interactively without human intervention
- **SC-007**: Prerequisite check failures provide clear error messages with installation instructions 100% of the time
- **SC-008**: Dependency resolution succeeds for 95% of valid Git repository URLs on first attempt

## Context & Background *(optional)*

### Problem Statement

SpecLedger originally attempted to replace GitHub Spec Kit by duplicating all SDD workflow commands (`specledger.specify`, `specledger.plan`, etc.). This created several problems:

1. **Maintenance burden**: Every Spec Kit update requires manual synchronization
2. **User confusion**: Unclear whether to use `/speckit.*` or `/specledger.*` commands
3. **Framework lock-in**: Users cannot choose alternative SDD frameworks like OpenSpec
4. **Fragmentation**: The community is split between GitHub Spec Kit, OpenSpec, and SpecLedger implementations

The current metadata format (.mod files) is also limited:
- Text-based format is hard to parse programmatically
- No structured fields for framework choice or advanced dependency metadata
- Cannot represent complex dependency relationships

### Why This Matters

SpecLedger should focus on its unique value propositions:
1. **Bootstrap orchestration**: Simplify tool installation and project setup
2. **Dependency management**: Enable spec composition and reuse across repositories
3. **Tool neutrality**: Support multiple SDD frameworks rather than replacing them

By becoming a thin wrapper, SpecLedger can:
- Reduce maintenance overhead (no need to sync with upstream frameworks)
- Increase adoption (users can choose their preferred framework)
- Focus development effort on unique features (dependency management)

## Assumptions *(optional)*

- Users are familiar with command-line tools and Git workflows
- Users have internet access for installing tools via mise and fetching dependencies
- Users prefer YAML for configuration due to widespread adoption in devops tooling
- mise is an acceptable dependency manager for development tools (alternative to manual installation)
- The SpecLedger community values tool neutrality over opinionated workflows
- Existing SpecLedger projects using .mod files are willing to migrate to YAML format
- Spec Kit and OpenSpec will remain the dominant SDD frameworks for the foreseeable future
- Claude Code, Cursor, and similar AI coding assistants support custom slash commands via `.claude/commands/` directory

## Dependencies *(optional)*

### External Dependencies

- **mise**: Required for managing tool installations (beads, perles, frameworks)
- **Git**: Required for cloning dependencies and version control
- **GitHub Spec Kit** (optional): One of the supported SDD frameworks
- **OpenSpec** (optional): One of the supported SDD frameworks
- **Beads** (bd): Required for issue/task tracking
- **Perles**: Required for workflow composition

### Internal Dependencies

- This feature modifies core SpecLedger CLI commands and must be coordinated with any in-progress features
- Migration from .mod to YAML format affects any code that reads/writes project metadata
- Removal of duplicate SDD commands may break existing workflows for users who rely on `specledger.*` commands

## Previous Work *(optional)*

### Related Issues

(Query Beads for related issues - this will be filled by the agent executing this spec)

### Related Branches

- Branch `003-cli-unification`: Unified the SpecLedger CLI into a single binary - this work builds on that foundation
- Branch `002-spec-dependency-linking`: Established the dependency management system - this work replaces .mod format with YAML

## Out of Scope *(optional)*

The following are explicitly not included in this redesign:

- Creating a new SDD framework to compete with Spec Kit or OpenSpec
- Modifying or extending Spec Kit or OpenSpec functionality
- Building integrations with specific AI coding assistants beyond Claude command templates
- Implementing visual UIs or web interfaces for SpecLedger
- Automated spec-to-code generation or AI-powered planning
- Support for non-Git dependency sources (npm, local filesystems, etc.)
- Advanced dependency resolution features (version constraints, conflict resolution, etc.)
