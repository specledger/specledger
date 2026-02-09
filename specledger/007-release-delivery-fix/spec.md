# Feature Specification: Release Delivery Fix

**Feature Branch**: `007-release-delivery-fix`
**Created**: 2025-02-09
**Status**: Draft
**Input**: User description: "fix release make sure that release works well, user can install from binary, go get, shell script and homebrew"

**Scope Note**: This feature focuses on macOS (darwin) as the primary development target. Linux and Windows support will be added in future iterations.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Binary Download from GitHub Releases (Priority: P1)

As a user, I want to download a pre-built binary from the GitHub releases page so that I can quickly install and run the tool without compiling from source.

**Why this priority**: Direct binary download is the most reliable installation method and serves as the foundation for all other installation methods.

**Independent Test**: Can be fully tested by downloading a binary from a release, extracting it, and verifying the `sl` command runs correctly.

**Acceptance Scenarios**:

1. **Given** a user visits the GitHub releases page, **When** they download the binary for their platform, **Then** the downloaded file should be a valid archive containing the `sl` binary
2. **Given** a user downloads the macOS binary (darwin_amd64), **When** they extract and run it, **Then** the binary should execute without errors
3. **Given** a user downloads the Linux binary (linux_amd64), **When** they extract and run it, **Then** the binary should execute without errors
4. **Given** a user downloads the Windows binary (windows_amd64), **When** they extract and run it, **Then** the binary should execute without errors
5. **Given** a user downloads a release binary, **When** they verify the checksum, **Then** the checksum should match the provided checksums.txt file

---

### User Story 2 - Shell Script Installation (Priority: P1)

As a user, I want to install the tool with a single curl command so that I can quickly get started without manually downloading and extracting files.

**Why this priority**: One-line installation is a standard expectation for CLI tools and significantly reduces adoption friction.

**Independent Test**: Can be fully tested by running the install script from a clean system and verifying the `sl` command is available afterward.

**Acceptance Scenarios**:

1. **Given** a user runs the install script on macOS, **When** the script completes, **Then** the `sl` binary should be installed in a directory on their PATH
2. **Given** a user runs the install script on Linux, **When** the script completes, **Then** the `sl` binary should be installed in a directory on their PATH
3. **Given** a user already has `sl` installed, **When** they run the install script again, **Then** it should overwrite the existing binary with the latest version
4. **Given** the install script runs, **When** it downloads the binary, **Then** it should verify the checksum before installation
5. **Given** the installation directory is not on PATH, **When** the script completes, **Then** it should provide instructions to add the directory to PATH

---

### User Story 3 - Homebrew Installation (Priority: P1)

As a macOS user, I want to install the tool via Homebrew so that I can use my preferred package manager and easily update the tool.

**Why this priority**: Homebrew is the de facto standard package manager for macOS and many developers expect it to be available.

**Independent Test**: Can be fully tested by tapping the repository and installing via brew, then verifying the `sl` command works.

**Acceptance Scenarios**:

1. **Given** a user taps the homebrew-specledger repository, **When** they run `brew install specledger`, **Then** the installation should complete successfully
2. **Given** a user has specledger installed via Homebrew, **When** they run `brew upgrade specledger`, **Then** they should receive the latest version
3. **Given** a user installs via Homebrew, **When** they run `sl --version`, **Then** it should display the correct version number
4. **Given** a new release is published, **When** the GoReleaser automation runs, **Then** it should automatically update the Homebrew formula in the tap repository

---

### User Story 4 - Go Install (Priority: P2)

As a Go developer, I want to install the tool using `go install` so that I can install it alongside other Go tools without additional setup.

**Why this priority**: Go developers often prefer using `go install` for CLI tools as it integrates with their existing Go workflow.

**Independent Test**: Can be fully tested by running `go install` and verifying the binary is installed to GOPATH/bin or GOBIN.

**Acceptance Scenarios**:

1. **Given** a user has Go installed, **When** they run `go install github.com/specledger/specledger/cmd@latest`, **Then** the binary should be compiled and installed
2. **Given** a user runs `go install`, **When** the installation completes, **Then** the `sl` command should be available from their shell
3. **Given** a user installs via `go install`, **When** they run `sl --version`, **Then** it should display the correct version information
4. **Given** a Go repository is tagged with a version, **When** a user runs `go install` with that version, **Then** they should receive that specific version

---

### User Story 5 - Release Automation (Priority: P1)

As a maintainer, I want to create releases by pushing a git tag so that all build artifacts are automatically generated and published.

**Why this priority**: Automated releases reduce manual effort, ensure consistency, and prevent human error during the release process.

**Independent Test**: Can be fully tested by pushing a tag and verifying all artifacts are created and published correctly.

**Acceptance Scenarios**:

1. **Given** a maintainer pushes a version tag (e.g., v1.0.0), **When** the GitHub Actions workflow triggers, **Then** it should build binaries for all supported platforms
2. **Given** the release workflow completes, **When** users visit the releases page, **Then** they should see a new release with all platform binaries attached
3. **Given** the release workflow runs, **When** it completes, **Then** a checksums.txt file should be generated and attached to the release
4. **Given** the release workflow runs, **When** it completes, **Then** the Homebrew formula should be updated in the tap repository
5. **Given** a release tag is pushed, **When** the workflow runs, **Then** it should complete without errors or deprecation warnings

---

### Edge Cases

- What happens when a user tries to install on an unsupported platform (e.g., FreeBSD, ARM v6)?
- How does the install script handle permission errors when writing to the installation directory?
- What happens when the GitHub releases page is temporarily unavailable?
- How does the system handle a corrupted download during installation?
- What happens when a user has an old version of the tool and tries to upgrade?
- How are checksums verified during installation?
- What happens when GoReleaser configuration has syntax errors?
- How does the system handle concurrent releases being created?
- What happens when the Homebrew tap repository doesn't exist?
- How does the install script handle different shell environments (bash, zsh, fish)?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST build and publish binaries for macOS (amd64, arm64)
- **FR-002**: The system MUST generate SHA256 checksums for all release binaries
- **FR-003**: The system MUST attach all binaries and checksums to GitHub releases automatically when a version tag is pushed
- **FR-004**: The install script MUST download the correct binary for macOS (amd64 or arm64 based on system architecture)
- **FR-005**: The install script MUST verify the binary checksum before installation
- **FR-006**: The install script MUST install the binary to a user-writable directory (~/.local/bin by default)
- **FR-007**: The install script MUST provide clear instructions if the installation directory is not on PATH
- **FR-008**: The system MUST provide a Homebrew formula for macOS users
- **FR-009**: The system MUST automatically update the Homebrew formula when a new release is published
- **FR-010**: The `go install` command MUST work for installing the tool from source
- **FR-011**: The binary MUST display version information when run with `--version` flag
- **FR-012**: Release archives MUST be named following the pattern `specledger_VERSION_OS_ARCH` (e.g., `specledger_1.0.2_darwin_amd64.tar.gz`)
- **FR-013**: macOS releases MUST use tar.gz format
- **FR-014**: The GoReleaser configuration MUST use valid syntax for version 2
- **FR-015**: Release binaries MUST be compiled with CGO_ENABLED=0 for portability

### Key Entities

- **Release Binary**: A compiled executable for a specific platform and architecture
- **Install Script**: A shell script that automates downloading and installing the correct binary
- **Homebrew Formula**: A Ruby file that defines how Homebrew installs and manages the software
- **Homebrew Tap**: A GitHub repository containing Homebrew formulas
- **GoReleaser**: A tool that automates building and releasing Go projects
- **Checksum File**: A text file containing SHA256 hashes of all release binaries for verification
- **Version Tag**: A git tag following semantic versioning (e.g., v1.0.0) that triggers release automation

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can install the tool via shell script on macOS in under 1 minute on a typical broadband connection
- **SC-002**: Users can install the tool via Homebrew on macOS in under 2 minutes including the tap command
- **SC-003**: All installation methods (binary download, script, Homebrew, go install) produce a working binary on macOS
- **SC-004**: Release automation completes within 5 minutes of pushing a version tag for macOS builds
- **SC-005**: 100% of macOS release binaries pass checksum verification
- **SC-006**: Installation script works on macOS (both Intel and Apple Silicon)
- **SC-007**: New releases are available via all installation methods within 10 minutes of tagging
- **SC-008**: Release artifacts include binaries for both darwin_amd64 and darwin_arm64
- **SC-009**: The install script provides helpful error messages for all failure scenarios
- **SC-010**: Users can upgrade to a new version using the same installation method

### Previous work

### Epic: 006 - Open Source Readiness

- **Verify GoReleaser configuration**: Initial setup of GoReleaser for automated releases (completed but had deprecation warnings)
- **Dry-run release verification**: Task to test the release process (open)
- **Test release process dry-run**: Additional verification task (open)

### Related Issues

- **GoReleaser v2 deprecation warnings**: Recent fixes addressed deprecated `archives.format`, `brews` → `homebrew_casks`, and broken `windows_arm_7` target
- **Archive naming mismatch**: Fixed binary naming from `sl_*` to `specledger_*` to match install script expectations
- **Homebrew configuration**: Changed from `homebrew_casks` (GUI apps) back to `brews` (CLI tools)
