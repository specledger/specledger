# Feature Specification: Fix Executable Permissions for Template Scripts

**Feature Branch**: `011-fix-chmod`
**Created**: 2026-02-17
**Status**: Draft
**Input**: User description: "fix chmod +x command missing in sl init/sl new script that cause that copied template scripts cannot be executed"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Scripts Work Immediately After Bootstrap (Priority: P1)

As a developer using SpecLedger, when I run `sl init` or `sl new` to bootstrap a new project, I expect all utility scripts copied from templates to be immediately executable without needing to manually run `chmod +x`.

**Why this priority**: This is the core bug fix - users cannot use the copied scripts without this fix, blocking their workflow completely.

**Independent Test**: Can be fully tested by running `sl init` on a new project and immediately executing `.specledger/scripts/bash/create-new-feature.sh --help` without any permission errors.

**Acceptance Scenarios**:

1. **Given** a fresh directory, **When** user runs `sl init my-project`, **Then** all `.sh` files in `.specledger/scripts/bash/` have execute permission (`-rwxr-xr-x` or similar)
2. **Given** a project bootstrapped with `sl new`, **When** user runs any script in `.specledger/scripts/bash/` directly, **Then** the script executes without "Permission denied" error
3. **Given** a project with existing scripts, **When** user re-runs `sl init --force`, **Then** scripts remain executable after being overwritten

---

### User Story 2 - Existing Projects Can Be Fixed (Priority: P2)

As a developer with existing SpecLedger projects, when I update to a version with this fix and re-run `sl init --force`, my scripts should become executable.

**Why this priority**: Allows existing users to benefit from the fix without creating new projects.

**Independent Test**: Can be tested by having an existing project with non-executable scripts, running `sl init --force`, and verifying scripts are now executable.

**Acceptance Scenarios**:

1. **Given** an existing project with non-executable scripts, **When** user runs `sl init --force`, **Then** all scripts are updated with execute permissions
2. **Given** an existing project, **When** user applies the fix, **Then** scripts work identically to a fresh bootstrap

---

### Edge Cases

- What happens when a script file already has execute permission? The system should preserve it and not cause errors.
- What happens when copying to a filesystem that doesn't support execute permissions (e.g., some network mounts)? The system should complete successfully, and the permission error is deferred to execution time.
- What happens with non-.sh files that have shebangs? Files with shebangs indicating executability should also receive execute permissions.

## Clarifications

### Session 2026-02-17

- Q: When re-running `sl init --force` on an existing project with non-executable scripts, what behavior do you want? → A: Always set +x (ensures consistency and fixes the bug for existing projects)

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST set execute permission (owner+group+other executable bits) on all files with `.sh` extension when copying from embedded templates
- **FR-002**: The system MUST set execute permission on files containing a shebang line (`#!`) in the first line when copying from embedded templates
- **FR-003**: The system MUST apply execute permissions consistently in both `sl init` and `sl new` commands
- **FR-004**: The system MUST apply execute permissions when copying playbook templates via `CopyPlaybooks` function
- **FR-005**: The system MUST apply execute permissions when copying embedded skills via `applyEmbeddedSkills` function
- **FR-006**: The system MUST set execute permissions even when overwriting existing files with `--force` flag (always ensures scripts are executable)

### Dependencies & Assumptions

- Assumes target filesystem supports Unix file permissions
- Assumes embedded template files are correctly identified by extension or content
- No external specifications need to be added as dependencies

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: All shell scripts (`.sh` files) in a bootstrapped project have execute permission immediately after `sl init` or `sl new` completes
- **SC-002**: Users can run any script in `.specledger/scripts/bash/` directly without first running `chmod +x`
- **SC-003**: Zero "Permission denied" errors when attempting to execute copied template scripts on supported filesystems
- **SC-004**: The fix applies to both new projects and existing projects using `--force` re-initialization

### Previous work

No related previous work found in Beads issue tracker.
