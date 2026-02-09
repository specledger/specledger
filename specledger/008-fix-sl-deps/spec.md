# Feature Specification: Fix SpecLedger Dependencies Integration

**Feature Branch**: `008-fix-sl-deps`
**Created**: 2026-02-09
**Status**: Draft
**Input**: User description: "fix sl deps, .claude skills and command, specledger.yaml should contains specs path that is referred by the sl deps command"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Resolve Downloads Dependencies to Cache (Priority: P1)

As a developer using SpecLedger with external specification dependencies, I want the `sl deps resolve` command to download dependency repositories to `~/.specledger/cache` so that I can access them offline and for AI context.

**Why this priority**: This is the core functionality of the dependency system - without downloading dependencies, the entire feature doesn't work.

**Independent Test**: Can be fully tested by adding a dependency with `sl deps add`, running `sl deps resolve`, and verifying that the repository is cloned to `~/.specledger/cache/<dir-name>/`.

**Acceptance Scenarios**:

1. **Given** a project with dependencies declared in `specledger.yaml`, **When** I run `sl deps resolve`, **Then** each dependency should be cloned to `~/.specledger/cache/<dir-name>/`
2. **Given** a dependency that has already been resolved, **When** I run `sl deps resolve` again, **Then** it should update the repository to the latest commit on the configured branch
3. **Given** a dependency with a custom branch like `develop`, **When** I run `sl deps resolve`, **Then** the repository should be checked out to that branch

---

### User Story 2 - Configure Artifact Path for Current Project (Priority: P1)

As a developer setting up a SpecLedger project, I want `specledger.yaml` to have an `artifact_path` field that specifies where my project's artifacts are located so that tools and AI agents can find them.

**Why this priority**: Critical for project setup - without knowing where the artifacts are, the system cannot resolve references correctly.

**Independent Test**: Can be tested by creating a new project with `sl new` and verifying that `specledger.yaml` contains the `artifact_path` field pointing to the correct location.

**Acceptance Scenarios**:

1. **Given** I create a new project with `sl new`, **When** I inspect `specledger.yaml`, **Then** it should contain an `artifact_path` field pointing to the artifacts directory (default: `specledger/`)
2. **Given** I initialize SpecLedger in an existing repository with `sl init`, **When** I inspect `specledger.yaml`, **Then** it should detect and configure the `artifact_path` based on the existing structure
3. **Given** a project with custom artifacts location, **When** I update `specledger.yaml`, **Then** tools should use the configured `artifact_path` to find artifact files

---

### User Story 3 - Discover Artifact Path in Dependency Repositories (Priority: P1)

As a developer adding a dependency, I want the system to automatically find the artifact path in SpecLedger repositories and allow manual specification for non-SpecLedger repos.

**Why this priority**: Essential for resolving references correctly - SpecLedger repos have a known structure, but other repos need manual configuration.

**Independent Test**: Can be tested by adding a SpecLedger repo (auto-detect) and a non-SpecLedger repo (manual path).

**Acceptance Scenarios**:

1. **Given** I add a SpecLedger repository dependency, **When** the system resolves it, **Then** it should read the dependency's `specledger.yaml` to find its `artifact_path` automatically
2. **Given** I add a non-SpecLedger repository, **When** I run `sl deps add <url> --artifact-path <path>`, **Then** the system should store the manual artifact path for that dependency
3. **Given** a dependency with a detected artifact path, **When** I run `sl deps list`, **Then** it should show both the dependency URL and its artifact path

---

### User Story 4 - Reference Artifacts Across Repositories (Priority: P1)

As a developer working with dependencies, I want to reference artifacts from dependency repositories by combining my project's `artifact_path` with the dependency's artifact path so that the system can locate the correct files.

**Why this priority**: Core reference resolution - without this, the entire dependency system cannot work.

**Independent Test**: Can be tested by adding a dependency, then referencing an artifact from it and verifying the system resolves the correct file path.

**Acceptance Scenarios**:

1. **Given** my project has `artifact_path: specledger/`, **When** I reference `dependency-name:artifact.md`, **Then** the system should resolve to `specledger/dependency-name/artifact.md`
2. **Given** my project has `artifact_path: docs/specs/`, **When** I reference a dependency with artifact path `specs/api.md`, **Then** the system should resolve to `docs/specs/dependency-name/specs/api.md`
3. **Given** a dependency with no artifact path specified, **When** I reference it, **Then** the system should use the default artifact path

---

### User Story 5 - Create Claude Code Command Files for All Deps Operations (Priority: P1)

As a developer using Claude Code, I want `.claude/commands/` files for all deps operations so that the AI agent can execute deps commands correctly.

**Why this priority**: Critical for AI integration - without command files, the AI agent doesn't know how to execute deps operations.

**Independent Test**: Can be tested by verifying that each deps subcommand has a corresponding `.claude/commands/specledger.<operation>.md` file.

**Acceptance Scenarios**:

1. **Given** the SpecLedger CLI with deps commands, **When** I inspect `.claude/commands/`, **Then** I should find files for: `specledger.add-deps`, `specledger.remove-deps`, `specledger.list-deps`, `specledger.resolve-deps`, `specledger.update-deps`
2. **Given** a command file like `specledger.resolve-deps.md`, **When** an AI agent reads it, **Then** it should contain clear instructions on how to execute the resolve operation
3. **Given** an AI agent following a command file, **When** it executes the command, **Then** the operation should complete successfully

---

### User Story 6 - Update SpecLedger-Deps Skill Documentation (Priority: P2)

As a developer using Claude Code, I want the `.claude/skills/specledger-deps/` skill to contain comprehensive documentation on how to use the deps system so that the AI agent can provide accurate assistance.

**Why this priority**: Important for AI assistance but the system can function without perfect documentation.

**Independent Test**: Can be tested by reading the skill file and verifying it contains clear explanations of all deps operations and workflows.

**Acceptance Scenarios**:

1. **Given** an AI agent with access to SpecLedger skills, **When** it references `specledger-deps`, **Then** it should understand how the deps system works
2. **Given** a developer asking about dependency management, **When** the AI agent uses the specledger-deps skill, **Then** it should provide accurate guidance on using deps commands
3. **Given** the specledger-deps skill, **When** I read its documentation, **Then** it should explain the relationship between `artifact_path`, cached dependencies, and reference resolution

---

### Edge Cases

- What happens when the specified artifact path doesn't exist in the dependency repository?
- What happens when resolving dependencies and the cache directory doesn't exist?
- How does the system handle network errors during `sl deps resolve`?
- What happens when the cache directory has insufficient disk space?
- What happens if a SpecLedger repo doesn't have a `specledger.yaml` file?
- What happens if an AI agent tries to use a deps command but the corresponding command file is missing?
- How does the system handle outdated skill or command documentation when the CLI behavior changes?

## Requirements *(mandatory)*

### Functional Requirements

**Dependency Resolution (Core)**

- **FR-001**: The `sl deps resolve` command MUST download/clone dependency repositories to `~/.specledger/cache/<dir-name>/`
- **FR-002**: The `sl deps resolve` command MUST handle partial downloads and resume interrupted downloads
- **FR-003**: When resolving, the system MUST update repositories to the latest commit on their configured branches
- **FR-004**: The system MUST use the alias (if provided) or generate a directory name for caching

**Artifact Path Configuration (Current Project)**

- **FR-005**: The `specledger.yaml` file MUST contain an `artifact_path` field that specifies where the current project's artifacts are located
- **FR-006**: The `artifact_path` field MUST default to a standard location like `specledger/` if not specified
- **FR-007**: The `sl new` command MUST initialize the `artifact_path` field when creating a new project
- **FR-008**: The `sl init` command MUST detect and configure the `artifact_path` field based on existing project structure

**Artifact Path Discovery (Dependencies)**

- **FR-009**: When adding a SpecLedger repository dependency, the system MUST read the dependency's `specledger.yaml` to automatically discover its `artifact_path`
- **FR-010**: When adding a non-SpecLedger repository, the system MUST allow manual specification of the artifact path via `--artifact-path` flag
- **FR-011**: The `sl deps add` command MUST accept an optional third argument for the reference path within the current project's `artifact_path` (default: same as dependency name)
- **FR-012**: The `sl deps list` command MUST display both the dependency URL and its artifact path (or discovered path)

**Reference Resolution**

- **FR-013**: When referencing artifacts from dependencies, the system MUST resolve paths by combining the current project's `artifact_path` with the dependency's artifact path
- **FR-014**: The system MUST support nested references (e.g., `artifact_path: specledger/` with dependency artifact path `specs/` resolves to `specledger/<dep-name>/specs/`)
- **FR-015**: The system MUST validate that resolved artifact paths point to existing files
- **FR-016**: The system MUST provide clear error messages when artifact paths cannot be resolved

**Claude Code Integration**

- **FR-017**: The `.claude/commands/` directory MUST contain command files for all deps operations: `specledger.add-deps`, `specledger.remove-deps`, `specledger.list-deps`, `specledger.resolve-deps`, `specledger.update-deps`
- **FR-018**: Each command file MUST provide clear instructions that the AI agent can follow to execute the deps operation
- **FR-019**: The `.claude/skills/specledger-deps/` skill MUST contain comprehensive documentation on dependency management workflow, including artifact_path discovery and reference resolution
- **FR-020**: Command files and skills MUST be kept in sync with CLI behavior changes

### Key Entities

- **Dependency**: Represents an external specification repository with properties:
  - `url`: Git repository URL
  - `branch`: Git branch (default: main)
  - `artifact_path`: The path within the dependency repo where artifacts are stored (auto-discovered for SpecLedger repos, manual for others)
  - `path`: Reference path within the current project's `artifact_path` to this dependency (default: same as dependency name)
  - `alias`: Optional short name for the dependency
  - `importPath`: Generated path for AI context
  - `resolvedCommit`: The Git commit SHA that was resolved

- **SpecLedger Project Metadata**: Stored in `specledger.yaml` with properties:
  - `artifact_path`: Path to the current project's artifacts directory (e.g., `specledger/`, `docs/specs/`)
  - `dependencies`: List of external specification dependencies

- **Cached Repository**: Local storage at `~/.specledger/cache/<dir-name>/` containing cloned dependency repository

- **Claude Command File**: A markdown file in `.claude/commands/` that provides instructions for the AI agent to execute a specific SpecLedger command

- **Claude Skill File**: A markdown file in `.claude/skills/` that provides comprehensive documentation on a feature or capability

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Developers can run `sl deps resolve` and dependencies are downloaded to `~/.specledger/cache/`
- **SC-002**: The `specledger.yaml` file contains an `artifact_path` field that correctly points to the project's artifacts directory
- **SC-003**: When adding SpecLedger repository dependencies, the system automatically discovers their `artifact_path`
- **SC-004**: When adding non-SpecLedger repository dependencies, users can specify `--artifact-path` flag
- **SC-005**: Artifact references are resolved correctly by combining project's `artifact_path` with dependency's artifact path
- **SC-006**: All 5 deps operations have corresponding `.claude/commands/specledger.<operation>.md` files
- **SC-007**: The `.claude/skills/specledger-deps/` skill contains comprehensive documentation on the deps workflow
- **SC-008**: Dependencies are successfully downloaded and cached in 95% of cases (excluding network errors)

### Previous work

### Epic: SL-26y - Release Delivery Fix

- **SL-m82**: Fix install script architecture detection - Related to installation but not dependencies
- **SL-vhp**: Simplify GoReleaser builds - Related to release delivery but not dependencies

### Related Commands

- **VarDepsCmd**: The main dependencies command with subcommands for add, list, remove, resolve, update
- **VarResolveCmd**: The resolve command that downloads/clones dependency repositories
- **metadata.LoadFromProject**: Loads the `specledger.yaml` file from project directory
- **metadata.SaveToProject**: Saves dependencies back to `specledger.yaml`

### Current Issues Identified

1. The `.claude/commands/` directory only contains `specledger.add-deps.md` and `specledger.remove-deps.md`, missing commands for: `list-deps`, `resolve-deps`, `update-deps`
2. The `specledger.yaml` structure may not have an `artifact_path` field for the current project
3. Dependencies in `specledger.yaml` may not be storing artifact_path information for non-SpecLedger repos
4. The system may not be reading dependency `specledger.yaml` files to discover artifact paths
5. The `sl deps resolve` command may not be correctly downloading to `~/.specledger/cache/`
6. The `.claude/skills/specledger-deps/` skill may not contain comprehensive documentation
