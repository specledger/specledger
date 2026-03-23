# Feature Specification: Shared Project Root Resolution

**Feature Branch**: `610-shared-project-root`
**Created**: 2026-03-23
**Status**: Draft
**Input**: User description: "Extract findProjectRoot to shared utility and fix doctor command subdirectory resolution (GitHub issue #81)"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Run Commands from Subdirectories (Priority: P1)

A user working inside a SpecLedger project navigates into a subdirectory (e.g., `pkg/cli/commands/`) and runs `sl doctor --template`. The command should locate the project root by walking up the directory tree and function correctly, rather than failing with "not in a SpecLedger project."

**Why this priority**: This is the core bug reported in issue #81. Users expect all `sl` commands to work regardless of their current working directory within the project, consistent with how tools like `git` behave.

**Independent Test**: Can be tested by running `sl doctor --template` from any subdirectory within a SpecLedger project and verifying it produces correct output.

**Acceptance Scenarios**:

1. **Given** a user is in a subdirectory of a SpecLedger project, **When** they run `sl doctor --template`, **Then** the command locates the project root and executes successfully.
2. **Given** a user is in the project root directory, **When** they run `sl doctor --template`, **Then** the command continues to work as before (no regression).
3. **Given** a user is in a directory outside any SpecLedger project, **When** they run `sl doctor --template`, **Then** a clear error message indicates no project was found.

---

### User Story 2 - Consistent Project Root Detection Across All Commands (Priority: P2)

All commands that need project context (doctor, session, comment, revise, mockup, spec-info, context-update, spec-create, spec-setup-plan) use the same shared utility for finding the project root, ensuring consistent behavior regardless of which command is invoked.

**Why this priority**: The shared utility prevents the same bug from recurring in other commands and eliminates code duplication. Without this, each command would need individual fixes.

**Independent Test**: Can be tested by running various `sl` commands from subdirectories and verifying they all resolve the project root correctly.

**Acceptance Scenarios**:

1. **Given** a user is in a subdirectory, **When** they run any project-aware `sl` command (e.g., `sl session list`, `sl comment add`), **Then** the command finds the project root and functions correctly.
2. **Given** two different commands are run from the same subdirectory, **When** both need the project root, **Then** they both resolve to the same directory.

---

### User Story 3 - Filesystem Boundary Safety (Priority: P3)

The project root search stops at filesystem boundaries (e.g., the root directory `/`) to avoid traversing outside the user's project tree or causing performance issues on deeply nested paths.

**Why this priority**: This is a safety concern — without a boundary, the search could walk up to `/` on every invocation from a non-project directory, causing unnecessary filesystem operations.

**Independent Test**: Can be tested by running `sl doctor` from a directory that is not within any SpecLedger project and verifying the command fails promptly with a clear error.

**Acceptance Scenarios**:

1. **Given** a user is in `/tmp` (no SpecLedger project above), **When** they run `sl doctor`, **Then** the command returns an error within a reasonable time without traversing to the filesystem root.
2. **Given** a user is at the filesystem root `/`, **When** they run any `sl` command, **Then** a clear "not in a SpecLedger project" error is shown.

---

### Edge Cases

- What happens when the `specledger.yaml` file exists but is malformed or empty?
- How does the system handle symbolic links in the directory path?
- What happens when the user has read permissions on the subdirectory but not on a parent directory?
- What happens when a nested project exists (a SpecLedger project inside another SpecLedger project)?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST provide a shared utility function for resolving the project root directory from any location within a project.
- **FR-002**: The project root resolution MUST walk up the directory tree from the current working directory, checking each level for the presence of `specledger.yaml`.
- **FR-003**: The project root search MUST stop at the filesystem root (`/`) and return a clear error if no project root is found.
- **FR-004**: The `sl doctor --template` command MUST use the shared project root resolution instead of `os.Getwd()` directly.
- **FR-005**: The `sl doctor` human output function MUST use the shared project root resolution instead of `os.Getwd()` directly.
- **FR-006**: All other commands that currently use `os.Getwd()` for project context MUST be audited and updated to use the shared utility where appropriate.
- **FR-007**: The shared utility MUST return the absolute path of the project root directory.
- **FR-008**: Error messages from the shared utility MUST clearly indicate that no SpecLedger project was found and suggest the user navigate to or initialize a project.

### Key Entities

- **Project Root**: The directory containing `specledger.yaml`, representing the top-level directory of a SpecLedger project.
- **Shared Utility**: A reusable function accessible by all commands that need project directory context.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: All `sl` commands that require project context work correctly when run from any subdirectory within a SpecLedger project.
- **SC-002**: The `sl doctor --template` command no longer fails with "not in a SpecLedger project" when run from a subdirectory (the specific bug in issue #81).
- **SC-003**: Only one implementation of project root resolution exists in the codebase — no duplicated logic across commands.
- **SC-004**: Error messages for "project not found" are consistent across all commands that use the shared utility.

### Previous work

- **GitHub Issue #81**: `sl doctor --template` fails to find project root from subdirectories — the bug that motivated this feature.
- **`findProjectRoot()` in deps.go**: Existing working implementation that walks up the directory tree, currently private to the deps command.

## Dependencies & Assumptions

### Assumptions

- The presence of `specledger.yaml` is the definitive marker for a SpecLedger project root (consistent with existing behavior in `deps.go`).
- Commands like `bootstrap` that intentionally create new projects may still use `os.Getwd()` directly, as they operate on the current directory by design.
- Nested SpecLedger projects are not a supported scenario — the first `specledger.yaml` found while walking up is the project root.
- Symbolic links are resolved using standard OS path resolution before walking up the tree.
