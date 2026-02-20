# Feature Specification: Doctor Version and Template Update

**Feature Branch**: `596-doctor-version-update`
**Created**: 2026-02-20
**Status**: Draft
**Input**: User description: "sl doctor need to be able to check sl cli version and ask user to update sl cli and also update the templates in the project"

## Clarifications

### Session 2026-02-20

- Q: How should the system detect which CLI version created the project templates? → A: Store a `template_version` field in specledger.yaml metadata
- Q: How should customized template files be handled during updates? → A: Skip customized files, display summary of skipped files after update
- Q: Should uncommitted changes in .claude/ block template updates? → A: Warn about uncommitted changes but proceed with update
- Q: Should template update require a flag or be automatic? → A: Proactively offer interactive prompt when version mismatch detected, no flag needed

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Check CLI Version and Prompt for Update (Priority: P1)

As a SpecLedger user, I want `sl doctor` to check if my CLI version is up-to-date so that I know when I should update to get the latest features and bug fixes.

**Why this priority**: Version checking is the foundational capability - without it, users won't know when updates are available. This delivers immediate value by keeping users informed.

**Independent Test**: Can be fully tested by running `sl doctor` with an outdated CLI version and verifying that it displays version information and update instructions. Delivers value by informing users about available updates.

**Acceptance Scenarios**:

1. **Given** I have a SpecLedger CLI installed, **When** I run `sl doctor`, **Then** the CLI version is displayed alongside other diagnostic information
2. **Given** I have an outdated CLI version, **When** I run `sl doctor`, **Then** I see a clear message indicating a newer version is available with instructions on how to update
3. **Given** I have the latest CLI version, **When** I run `sl doctor`, **Then** I see a confirmation that my CLI is up-to-date
4. **Given** the version check cannot reach the remote server, **When** I run `sl doctor`, **Then** I see a graceful message indicating version check was skipped (not an error)

---

### User Story 2 - Update Project Templates (Priority: P1)

As a SpecLedger user, I want `sl doctor` to proactively offer to update my project's embedded templates (skills, commands) when a version mismatch is detected, so that my project stays in sync with the latest CLI's embedded resources.

**Why this priority**: Template updates are essential for keeping projects current with improvements to workflows, prompts, and configurations. Without this, projects become stale even if the CLI is updated.

**Independent Test**: Can be fully tested by initializing a project with an older CLI, updating the CLI, running `sl doctor`, and accepting the template update offer. Delivers value by keeping projects current.

**Acceptance Scenarios**:

1. **Given** I am in a SpecLedger project with outdated templates, **When** I run `sl doctor`, **Then** I see an interactive prompt offering to update templates
2. **Given** I am in a SpecLedger project and accept the template update offer, **When** the update completes, **Then** the embedded skills and templates in `.claude/` are updated to match the current CLI version and the template_version in specledger.yaml is updated
3. **Given** templates are already up-to-date, **When** I run `sl doctor`, **Then** I see a confirmation message that templates are current (no prompt needed)
4. **Given** I am not in a SpecLedger project, **When** I run `sl doctor`, **Then** template checking is skipped (no prompt offered)

---

### User Story 3 - Non-Interactive CI/CD Support (Priority: P2)

As a CI/CD pipeline maintainer, I want `sl doctor` to provide machine-readable output with version and template status so that I can automate environment validation.

**Why this priority**: CI/CD integration is important for teams but requires the foundational features (P1) to be in place first. Enables automated checks in pipelines.

**Independent Test**: Can be fully tested by running `sl doctor --json` in various scenarios and parsing the structured output. Delivers value for automation pipelines.

**Acceptance Scenarios**:

1. **Given** I run `sl doctor --json`, **When** the command completes, **Then** the output includes CLI version, latest available version, and template status in JSON format
2. **Given** CLI is outdated, **When** I run `sl doctor --json`, **Then** the JSON includes `"cli_update_available": true` and update details
3. **Given** templates are outdated and I run `sl doctor --json`, **Then** the JSON includes `"template_update_available": true` (template updates are not performed in JSON mode since no interactive prompt)

---

### Edge Cases

- **GitHub API rate-limited or unreachable**: Display a warning that version check was skipped, continue with other doctor checks. Do not block the command.
- **Customized template files**: Skip updating files that differ from embedded originals, display a summary of skipped files after update completes so users can manually merge if needed.
- **Uncommitted changes in `.claude/` directory**: Display a warning about uncommitted changes but proceed with the update. Users are responsible for their own git hygiene.
- **Project initialized with much older CLI version**: If template_version field is missing from specledger.yaml, assume templates need updating and offer the update prompt.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST display the currently installed CLI version in `sl doctor` output
- **FR-002**: System MUST query GitHub Releases API to check for the latest available CLI version
- **FR-003**: System MUST display a clear update message when a newer CLI version is available, including download instructions
- **FR-004**: System MUST display a confirmation message when the CLI is already at the latest version
- **FR-005**: System MUST gracefully handle network failures during version check without blocking the doctor command
- **FR-006**: System MUST detect if the current project's templates (`.claude/` directory) were created by an older CLI version by comparing the `template_version` stored in specledger.yaml against the current CLI version
- **FR-007**: System MUST proactively offer to update project templates when running `sl doctor` and a version mismatch is detected between CLI and project templates
- **FR-008**: System MUST update all embedded resources including skills (`.claude/commands/`, `.claude/skills/`) when user confirms the template update offer
- **FR-009**: System MUST preserve user-customized files during template updates by skipping them and displaying a summary of skipped files after the update completes
- **FR-010**: System MUST include version and template status information in `--json` output format
- **FR-011**: System MUST only offer template updates when run from within a SpecLedger project directory (detected via specledger.yaml presence)
- **FR-012**: System MUST store the template version in specledger.yaml when templates are initialized or updated

### Key Entities

- **CLI Version**: Represents the semantic version of the installed SpecLedger binary, including build metadata (commit, date)
- **Remote Version**: The latest available version from GitHub Releases, including download URLs
- **Template Version**: The CLI version that created/last updated the project templates, stored in specledger.yaml
- **Template Status**: Information about whether project templates match the current CLI version, including which files differ
- **Template Update Offer**: Interactive prompt displayed when version mismatch is detected, allowing user to confirm or skip template updates
- **Update Instructions**: Step-by-step guidance for updating the CLI (varies by installation method: Homebrew, binary download, go install)

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can see their CLI version status within 3 seconds of running `sl doctor`
- **SC-002**: Version check completes successfully in 95% of cases when network is available
- **SC-003**: Template updates complete in under 10 seconds for typical projects
- **SC-004**: 90% of users understand how to update their CLI after seeing the update message
- **SC-005**: No user-customized files in `.claude/` are lost during template updates
- **SC-006**: CI/CD pipelines can reliably parse `sl doctor --json` output

### Previous work

- **135-fix-missing-chmod-x**: Fixed executable permissions for templates after initialization
- **011-streamline-onboarding**: Added template embedding and project initialization
- **006-opensource-readiness**: Established GitHub Releases for distribution

### Epic: N/A - Standalone Feature
