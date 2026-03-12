# Feature Specification: Fix GoReleaser Build Version Injection

**Feature Branch**: `605-fix-goreleaser-ldflags`
**Created**: 2026-03-10
**Status**: Draft
**Input**: User description: "https://github.com/specledger/specledger/issues/65"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Released Binary Reports Correct Version (Priority: P1)

A user downloads a released binary of the specledger CLI. When they run `sl --version`, the CLI displays the actual release version (e.g., `v1.2.3`), not `dev`.

**Why this priority**: This is the core problem. Without correct version reporting, users cannot verify what version they have installed, and automated update checks cannot function correctly.

**Independent Test**: Can be fully tested by building a release binary with GoReleaser and verifying `sl --version` outputs the correct semantic version, commit hash, and build date.

**Acceptance Scenarios**:

1. **Given** a binary built via GoReleaser release pipeline, **When** the user runs `sl --version`, **Then** the output displays the tagged release version (e.g., `v1.2.3`), not `dev`.
2. **Given** a binary built via GoReleaser release pipeline, **When** the user runs `sl --version`, **Then** the output includes the commit hash and build date from the release.

---

### User Story 2 - Update Checker Detects Correct Installed Version (Priority: P1)

A user runs `sl doctor --update` to check for CLI updates. The system correctly compares the installed version against the latest available release and only prompts for update when a newer version actually exists.

**Why this priority**: The update mechanism is broken because the binary always reports `dev`, causing the update checker to always think an update is available regardless of actual version.

**Independent Test**: Can be fully tested by running `sl doctor` with a correctly versioned binary and verifying it reports "up to date" when on the latest release.

**Acceptance Scenarios**:

1. **Given** a released binary running the latest version, **When** the user runs `sl doctor`, **Then** the version check reports "up to date" and does not prompt for an update.
2. **Given** a released binary running an older version, **When** the user runs `sl doctor`, **Then** the version check correctly identifies the newer version and prompts for update.
3. **Given** a development build (unreleased), **When** the user runs `sl doctor`, **Then** the version check gracefully handles the `dev` version identifier.

---

### Edge Cases

- What happens when a user builds from source without GoReleaser (e.g., `go install`)?
  - The version should remain `dev` as a sensible default, and the update checker should handle this gracefully without errors.
- What happens if the GoReleaser configuration references variables that don't exist in the Go source?
  - The build should fail at release time, making the mismatch immediately visible in CI rather than silently producing broken binaries.
- What happens if the version is set correctly but `rootCmd.Version` is initialized before the ldflags values are applied?
  - The root command version must be set after variable initialization to ensure ldflags values are used.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The GoReleaser linker flags (ldflags) configuration MUST reference the correct Go variable names as declared in the CLI entry point source file.
- **FR-002**: The root command's version field MUST be set after build-time variables have been initialized, ensuring injected values take effect.
- **FR-003**: Released binaries MUST display the tagged release version, commit hash, and build date when queried for version information.
- **FR-004**: Development builds (without ldflags injection) MUST default to `dev` as the version identifier.
- **FR-005**: The update checker MUST correctly compare the installed version against available releases when the binary has a properly injected version.

### Key Entities

- **Build Variables**: The Go variables (`buildVersion`, `buildCommit`, `buildDate`) that receive values via linker flags at build time.
- **GoReleaser Configuration**: The `.goreleaser.yaml` file that defines how ldflags are passed during the release build process.
- **Root Command**: The Cobra root command whose `Version` field controls `--version` output.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users running a released binary see the correct semantic version (not `dev`) 100% of the time when checking `sl --version`.
- **SC-002**: The update checker correctly identifies whether the installed version is current, reducing false "update available" notifications to zero for users on the latest release.
- **SC-003**: The GoReleaser release pipeline produces binaries with correct version metadata on every release without manual intervention.

### Previous work

### Epic: SL-826d4e - Doctor Version and Template Update (596-doctor-version-update)

- **SL-9524be - Create pkg/version package with Version comparison**: Established the version package and comparison logic used by the update checker.
- **SL-faa3a6 - Implement GitHub Releases API client**: Implemented the mechanism for fetching latest release information for version comparison.
- **SL-445a25 - Add CLI version section to doctor output**: Added version display to `sl doctor` output.
- **SL-a9a5e9 - Add update instructions logic**: Implemented the update prompting logic that depends on accurate version reporting.

### Related Closed Issues (migrated spec)

- **SL-a68e99 - Add version variable to main.go**: Originally added the version variable to the CLI entry point.
- **SL-313a4e - Verify GoReleaser dry-run**: Previous verification of GoReleaser configuration that did not catch the variable name mismatch.
- **SL-a3f192 - Verify GoReleaser configuration**: Another prior verification pass.
- **SL-efe7c6 - Release Delivery Fix (007-release-delivery-fix)**: Earlier release pipeline fix effort.

## Assumptions

- The GoReleaser variable name mismatch is the root cause; no other build pipeline issues affect version injection.
- The `rootCmd.Version` initialization timing issue is a secondary fix that must be addressed alongside the ldflags correction.
- The existing version comparison logic in `pkg/version` is correct and will work properly once it receives actual version strings instead of `dev`.
