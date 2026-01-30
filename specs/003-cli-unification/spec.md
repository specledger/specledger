# Feature Specification: CLI Unification

**Feature Branch**: `003-cli-unification`
**Created**: 2026-01-30
**Status**: Draft
**Input**: User description: "integrate specledger bootstrap script ./sl as single cli tools with specledger, but keep TUI for interactive bootstrap. Also take care of the distribution of the cli tools: github release, self-built, uvx/npx style, package manager, etc"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Single Unified CLI Tool (Priority: P1)

Users should be able to interact with SpecLedger using a single unified CLI tool that provides both project bootstrap and specification dependency management capabilities. This eliminates the need for separate `sl` and `specledger` commands.

**Why this priority**: This is the core unification goal. Without a single CLI, users have to learn and maintain two separate tools that share common functionality. The TUI for bootstrap interactions remains, providing a consistent user experience.

**Independent Test**: Can be fully tested by installing the unified CLI and verifying it supports both bootstrap and dependency management commands without requiring a separate `sl` script.

**Acceptance Scenarios**:

1. **Given** a user has the unified CLI installed, **When** they run `sl` in a new directory, **Then** they see the interactive TUI bootstrap prompts for project creation
2. **Given** a user has an existing SpecLedger project, **When** they run `sl deps list`, **Then** they see the same dependency listing output as `specledger deps list`
3. **Given** a user is running the unified CLI, **When** they run `sl --version`, **Then** they see a single version message consistent with the Go CLI
4. **Given** a user needs to bootstrap a new project, **When** they run `sl new`, **Then** they see the same TUI experience as the original `sl` script

---

### User Story 2 - GitHub Releases (Priority: P1)

Users should be able to install the unified CLI directly from GitHub releases, ensuring they always have access to the latest stable version without manual compilation.

**Why this priority**: GitHub releases provide the most common and expected distribution channel for CLI tools. Users expect to be able to download and use binaries directly from GitHub releases.

**Independent Test**: Can be fully tested by downloading a CLI binary from a GitHub release and verifying it executes correctly and provides expected help output.

**Acceptance Scenarios**:

1. **Given** a user on macOS, **When** they run the installation script provided in releases, **Then** they have a functioning `sl` command in their PATH
2. **Given** a user on Linux, **When** they download the corresponding release binary, **Then** they can execute `sl --help` without errors
3. **Given** a user on Windows, **When** they download the release binary, **Then** they can execute `sl --help` without errors
4. **Given** a new release is published, **When** a user runs the installation script, **Then** they get the latest version

---

### User Story 3 - Self-Built Binaries (Priority: P2)

Users should be able to build and install the CLI from source on their local machines using standard build tools (Go, make).

**Why this priority**: Self-building is essential for development workflows and for users who want to customize or contribute to the project. It provides flexibility and control.

**Independent Test**: Can be fully tested by cloning the repository and building the CLI using `make build`, then verifying the resulting binary executes correctly.

**Acceptance Scenarios**:

1. **Given** a user has the source code, **When** they run `make build`, **Then** they get a binary at `bin/sl`
2. **Given** a user runs the built binary, **When** they execute `sl --help`, **Then** they see comprehensive help text
3. **Given** a user builds the CLI, **When** they run `sl new` in a new directory, **Then** the interactive TUI bootstrap works correctly

---

### User Story 4 - Self-Hosted / Local Binaries (Priority: P2)

Users should be able to use the CLI without installing dependencies by placing the binary directly in their project directory or a local installation location.

**Why this priority**: This provides a portable option for users who don't want to modify their PATH or install system-wide tools. Useful for CI/CD and temporary setups.

**Independent Test**: Can be fully tested by running the binary from a non-PATH location with `./sl --help` and verifying it works.

**Acceptance Scenarios**:

1. **Given** a user downloads the binary, **When** they place it in a project directory and run `./sl new`, **Then** the bootstrap TUI works correctly
2. **Given** a user has a binary in a non-PATH location, **When** they run it with `~/path/to/sl --help`, **Then** they see help output
3. **Given** the binary is run from within a SpecLedger project, **When** they run `sl deps list`, **Then** they see dependency listing for that project

---

### User Story 5 - UVX Style (Priority: P2)

Users should be able to use the CLI via a standalone executable without setup, similar to how tools like `uvx` work, though this may require first-time download.

**Why this priority**: UVX-style execution provides the fastest path to trying the tool without installation. Users can immediately test functionality before deciding to install.

**Independent Test**: Can be fully tested by executing the standalone executable (either pre-provided or generated) and verifying help output works.

**Acceptance Scenarios**:

1. **Given** a user has the standalone executable URL, **When** they run it, **Then** they see help output on first execution
2. **Given** a user runs the standalone CLI, **When** they execute `sl new`, **Then** they see the interactive bootstrap prompts
3. **Given** the standalone CLI is run without a project, **When** they try to use project-specific commands like `sl deps list`, **Then** they see an appropriate error message

---

### User Story 6 - Package Manager Integration (Priority: P3)

Users should be able to install the CLI via common package managers (Homebrew, npx, etc.) where available.

**Why this priority**: Package managers provide a familiar installation method for many users. However, this depends on external package manager ecosystems and may not be immediately possible for all platforms.

**Independent Test**: Can be fully tested by installing via the package manager and verifying the CLI is accessible and functional.

**Acceptance Scenarios**:

1. **Given** a macOS user, **When** they run `brew install specledger`, **Then** they have access to the `sl` command
2. **Given** a JavaScript/TypeScript user, **When** they run `npx @specledger/cli`, **Then** they can execute `sl` commands
3. **Given** a user installs via package manager, **When** they run `sl --version`, **Then** they see the correct version

---

### Edge Cases

- **CI/CD non-interactive environments**: The CLI MUST provide a fully non-interactive mode that allows project creation via flags (e.g., `sl new --project-name myproject --short-code abc`) without requiring TUI. When invoked in non-interactive mode, the TUI is skipped entirely.

- **Missing dependencies (gum, mise)**: The CLI MUST provide interactive fallback. When required TUI dependencies are missing, the CLI asks the user whether to install them or proceed with limited functionality. If the user chooses to install, the CLI provides clear instructions and optionally attempts installation.

- **Conflicting flags**: The CLI MUST prioritize `--help` over all other flags. If both `--help` and `--version` are specified, `--help` takes precedence and displays help with version information.

- **Non-Project directory**: When run from a directory without necessary configuration files, the CLI displays a clear error message indicating that it must be run within a SpecLedger project for dependency management commands.

- **Bootstrap in existing project**: When bootstrap is invoked in a directory that already contains a SpecLedger project, the CLI must detect this and either refuse to proceed with a clear error message, or offer a confirmation prompt to overwrite/initialize.

- **Permission errors**: When writing to user directories (e.g., `~/demos`), the CLI displays a clear error message explaining the permission issue and suggests alternative locations (current directory, project directory).

- **Incomplete template files**: When the bootstrap encounters incomplete template files, the CLI fails with a clear error message indicating which file is missing or incomplete and stops execution to prevent corrupted project creation.

## Clarifications

### Session 2026-01-30

- Q: What should be the minimum level of observability for the unified CLI (logging, metrics, or both)? → A: Debug-level logging only
- Q: When required external dependencies (gum, mise) are missing, what should the CLI's behavior be? → A: Provide interactive fallback
- Q: For CI/CD non-interactive environments, how should the CLI behave? → A: Provide fully non-interactive mode
- Q: What exit codes should the CLI use for different failure scenarios? → A: Standard 0/1 only

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The unified CLI MUST be named `sl` for bootstrap operations and MUST be aliased to accept `specledger` for backward compatibility with existing documentation
- **FR-002**: The CLI MUST provide a `bootstrap` or `new` command that invokes the interactive TUI for project creation
- **FR-003**: The CLI MUST support all current `specledger` dependency management commands (deps add, list, resolve, update, remove)
- **FR-004**: The CLI MUST provide `--help` and `--version` flags with consistent output across all commands
- **FR-004**: The CLI MUST provide `--help` and `--version` flags with consistent output across all commands
- **FR-005**: The CLI MUST provide a fully non-interactive mode that allows project creation via flags (e.g., `sl new --project-name myproject --short-code abc`) without requiring TUI
- **FR-006**: The CLI MUST detect when it's running in a non-interactive environment and skip TUI, using command-line flags instead
- **FR-007**: The CLI MUST include pre-flight checks for required dependencies (gum, mise) and provide interactive fallback to ask users whether to install dependencies or proceed with limited functionality
- **FR-008**: The CLI MUST log debug-level output to stderr for troubleshooting purposes
- **FR-008**: GitHub releases MUST include binaries for macOS (Darwin), Linux, and Windows
- **FR-009**: The CLI MUST provide a self-hosted binary distribution that works when placed in any directory
- **FR-010**: The CLI MUST provide a build mechanism (via Makefile) that produces both `specledger` and `sl` aliases
- **FR-011**: The CLI MUST provide standalone executable links that can be executed without setup
- **FR-012**: The CLI MUST support package manager installations where applicable (Homebrew, npx, etc.)
- **FR-013**: The CLI MUST use exit code 0 for successful operations and exit code 1 for any failure (standard error handling)

### Key Entities

- **CLI Binary**: The executable artifact that provides all SpecLedger functionality (bootstrap TUI + dependency management)
- **Bootstrap Project**: A newly created SpecLedger project containing configuration files for issue tracking, AI agents, and tool management
- **Dependency Specification**: An external specification declared in a project with associated cryptographic verification
- **Distribution Channel**: A method for users to obtain the CLI binary (GitHub releases, self-built, self-hosted, UVX-style, package manager)

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can complete project bootstrap using only the unified CLI (`sl` command) in under 3 minutes
- **SC-002**: 95% of users can install the CLI from GitHub releases without manual compilation
- **SC-003**: The unified CLI supports all existing `specledger` dependency management commands with identical functionality
- **SC-004**: 90% of users successfully bootstrap a new project on first attempt using the CLI
- **SC-005**: Users can run the CLI from both PATH locations and non-PATH locations without additional setup
- **SC-006**: The unified CLI works across macOS, Linux, and Windows without platform-specific code changes
- **SC-007**: In non-interactive environments (CI/CD), users can bootstrap projects using command-line flags in under 1 minute
- **SC-008**: 100% of bootstrap operations complete with a clear error message when pre-flight checks fail (no silent failures)

### Previous work

No previous work requires duplication for this feature.

### Epic: CLI-UN-001 - Unified CLI Architecture

- **Merge shell and Go CLI**: Consolidate the bash bootstrap script and Go CLI into a single CLI with shared functionality
- **TUI Integration**: Move the interactive TUI prompts from the bash script to the Go CLI
- **Command Structure**: Define the unified command structure (bootstrap/new, deps, refs, graph, vendor, tools)

### Epic: CLI-UN-002 - Distribution Channels

- **GitHub Releases**: Set up GitHub releases with cross-platform binaries
- **Self-Built**: Configure Makefile for local builds
- **Self-Hosted**: Provide standalone binary distribution strategy
- **UVX Style**: Implement standalone executable distribution
- **Package Managers**: Create package manifests for Homebrew, npx, etc.
