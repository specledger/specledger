# Feature Specification: Issues Storage Configuration

**Feature Branch**: `594-issues-storage-config`
**Created**: 2026-02-20
**Status**: Draft
**Input**: User description: "rename .issues.jsonl.lock -> issues.jsonl.lock, add lock to gitignore, utilize artifact_path from specledger.yaml to store the issues.jsonl"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Lock File Naming Convention (Priority: P1)

As a SpecLedger user, I want lock files to follow standard naming conventions (without leading dots) so that they are more visible and easier to manage, and so they can be properly gitignored.

**Why this priority**: This is the foundation for the gitignore configuration. Without properly named lock files, the gitignore pattern cannot work correctly.

**Independent Test**: Can be fully tested by creating issues and verifying the lock file is named `issues.jsonl.lock` (not `.issues.jsonl.lock`) in the spec directory.

**Acceptance Scenarios**:

1. **Given** a spec with no existing issues, **When** I create an issue using `sl issue create`, **Then** the lock file is created as `issues.jsonl.lock` (without leading dot)
2. **Given** an existing spec with the old `.issues.jsonl.lock` file, **When** I perform any issue operation, **Then** the system uses the new naming convention
3. **Given** cross-spec operations (no spec context), **When** operations occur, **Then** the lock file is named `issues.jsonl.lock` at the base artifact path

---

### User Story 2 - Artifact Path Configuration (Priority: P1)

As a SpecLedger user with a custom artifact path configuration, I want issues to be stored in my configured `artifact_path` from specledger.yaml so that they follow my project's organization structure.

**Why this priority**: This is critical for projects using custom artifact paths. Without this, issues are always stored in the default location regardless of configuration.

**Independent Test**: Can be tested by setting `artifact_path: docs/specs/` in specledger.yaml, creating issues, and verifying they are stored at `docs/specs/<spec>/issues.jsonl`.

**Acceptance Scenarios**:

1. **Given** specledger.yaml has `artifact_path: docs/specs/`, **When** I create an issue for spec "010-my-feature", **Then** it is stored at `docs/specs/010-my-feature/issues.jsonl`
2. **Given** specledger.yaml has `artifact_path: specledger/` (default), **When** I create an issue, **Then** it is stored at `specledger/010-my-feature/issues.jsonl`
3. **Given** specledger.yaml has no artifact_path defined, **When** I create an issue, **Then** it defaults to `specledger/<spec>/issues.jsonl`
4. **Given** listing issues with `--all` flag, **When** a custom artifact_path is configured, **Then** all specs are searched within that configured path

---

### User Story 3 - Lock File Gitignore (Priority: P2)

As a SpecLedger user, I want lock files automatically ignored by git so that concurrent access control files don't appear in version control.

**Why this priority**: Clean version control is important but the core functionality must work first. This depends on User Story 1 being complete.

**Independent Test**: Can be tested by creating issues and running `git status` to verify lock files are not shown as untracked.

**Acceptance Scenarios**:

1. **Given** the SpecLedger project .gitignore, **When** I check its contents, **Then** `issues.jsonl.lock` is listed in the pattern
2. **Given** a new project created via `sl init`, **When** I check the generated .gitignore, **Then** it contains `issues.jsonl.lock`
3. **Given** I create an issue, **When** I run `git status`, **Then** the `issues.jsonl.lock` file is not shown as untracked

---

### Edge Cases

- What happens when users have both old `.issues.jsonl.lock` and new `issues.jsonl.lock` files? The system should prefer the new naming and clean up old files when possible
- What happens if specledger.yaml is malformed or unreadable? Fall back to default `specledger/` path
- What happens when artifact_path ends without a trailing slash? The system should handle both `docs/specs` and `docs/specs/` correctly
- What happens when artifact_path is set to a non-existent directory? Create the directory structure when needed (existing behavior)

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The lock file MUST be renamed from `.issues.jsonl.lock` to `issues.jsonl.lock` (removing leading dot)
- **FR-002**: The issue store MUST use the `artifact_path` value from `specledger.yaml` as the base directory for issue storage
- **FR-003**: If `artifact_path` is not specified in specledger.yaml, the default path MUST be `specledger/`
- **FR-004**: The project's root .gitignore MUST include `issues.jsonl.lock` pattern
- **FR-005**: The embedded templates for new projects MUST include `issues.jsonl.lock` in their .gitignore (when applicable)
- **FR-006**: All issue CLI commands (`create`, `list`, `show`, `update`, `close`, `link`, `unlink`, `migrate`, `repair`) MUST respect the configured artifact_path
- **FR-007**: Cross-spec operations (list --all, GetIssueAcrossSpecs) MUST search within the configured artifact_path
- **FR-008**: The artifact_path MUST be read from specledger.yaml at the project root during store initialization

### Key Entities

- **Issue Store**: File-based storage using JSONL format at `<artifact_path>/<spec>/issues.jsonl`
- **Lock File**: File-based lock at `<artifact_path>/<spec>/issues.jsonl.lock` for concurrent access control
- **Artifact Path Configuration**: The `artifact_path` field in specledger.yaml that defines where all spec artifacts (including issues) are stored

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can create and list issues when `artifact_path` is set to a custom location
- **SC-002**: Lock files are named without leading dots and match the pattern `issues.jsonl.lock`
- **SC-003**: Lock files do not appear in `git status` output after adding to .gitignore
- **SC-004**: All existing issue functionality works correctly with custom artifact_path configurations
- **SC-005**: Default behavior (no artifact_path specified) remains unchanged with storage at `specledger/`

### Previous work

- **591-issue-tracking-upgrade**: Original implementation of the issue tracking system - this feature refines the storage configuration
- **593-ticket-rename**: Related feature that renames `sl issue` to `sl ticket` - this feature focuses specifically on storage configuration improvements
