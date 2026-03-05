# Feature Specification: Fix sl init on Windows

**Feature Branch**: `603-fix-init-windows`
**Created**: 2026-03-05
**Status**: Draft
**Input**: User description: "fix bug not sl init not work in windows (but work in macos and linux)"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Windows User Runs sl init Successfully (Priority: P1)

A developer on Windows runs `sl init` in an existing repository and the command completes without errors, setting up SpecLedger files the same way it does on macOS and Linux.

**Why this priority**: This is the core bug — `sl init` is completely broken on Windows. Every Windows user is blocked from using SpecLedger until this is fixed.

**Independent Test**: Run `sl init` in any directory on a Windows machine. The command should exit with code 0 and create `.specledger/`, `specledger/`, and `.claude/` directories.

**Acceptance Scenarios**:

1. **Given** a Windows machine with `sl` installed and a git repository, **When** the user runs `sl init`, **Then** the command completes successfully without error messages related to script execution or permissions.
2. **Given** the same scenario, **When** `sl init` completes, **Then** all expected files (specledger.yaml, playbook templates, skills) are created in the correct directories.
3. **Given** the same scenario, **When** `sl init` runs the post-init phase, **Then** it does not crash with an "exec format error" or "cannot execute binary file" or similar shell-script-execution failure on Windows.

---

### User Story 2 - Tool Availability Detection Works on Windows (Priority: P2)

When `sl init` detects whether optional tools (like `gum`) are available, it does so in a Windows-compatible way rather than relying on Unix-only shell built-ins.

**Why this priority**: Incorrect tool detection on Windows may cause silent failures or incorrect behavior that degrades the user experience, even if it doesn't fully block init.

**Independent Test**: Run `sl init` on Windows with `gum` not installed. The command should still proceed using the fallback plain-CLI mode, not crash or hang.

**Acceptance Scenarios**:

1. **Given** a Windows machine without `gum` installed, **When** `sl init` checks tool availability, **Then** the check works correctly and falls back to non-interactive mode without errors.
2. **Given** a Windows machine with `gum` installed and in PATH, **When** `sl init` checks tool availability, **Then** `gum` is correctly detected and used.

---

### User Story 3 - Cross-Platform Behavior Parity (Priority: P3)

The outcome of `sl init` on Windows is functionally identical to macOS and Linux — all the same files are created, metadata is written, and the initialization summary is displayed.

**Why this priority**: Consistency across platforms builds trust. Partial functionality on Windows creates confusion and support burden.

**Independent Test**: Run `sl init` on Windows and macOS in equivalent repositories. Compare the resulting file tree — they should be identical.

**Acceptance Scenarios**:

1. **Given** equivalent projects on Windows, macOS, and Linux, **When** `sl init` is run on each, **Then** the resulting directory structure and file contents are identical.
2. **Given** a playbook with post-init logic, **When** `sl init` runs on Windows, **Then** the equivalent post-init outcomes are achieved (even if the mechanism differs from running a `.sh` script).

---

### Edge Cases

- What happens when `bash` or `sh` is available on Windows (e.g., via Git Bash or WSL) but the post-init script still fails?
- How does the system behave when `sl init` is run inside a Git Bash terminal on Windows versus a native Windows Command Prompt or PowerShell?
- What happens when the Windows temp directory path contains spaces?
- What happens when `os.Chmod` is called on Windows (it succeeds silently but has no effect — this is not a crash but should be noted)?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The `sl init` command MUST complete successfully on Windows without errors related to shell script execution.
- **FR-002**: The post-init script execution mechanism MUST detect the current operating system and use a Windows-compatible approach when running on Windows.
- **FR-003**: If a Unix shell (bash, sh) is available on Windows (e.g., via Git for Windows or WSL), the system SHOULD use it to execute `init.sh`; otherwise it MUST skip the script execution gracefully.
- **FR-004**: Tool availability detection (e.g., checking if `gum` is installed) MUST use a method that works on Windows without relying on Unix-only shell built-ins (`command -v`).
- **FR-005**: When the post-init script is skipped on Windows (no shell available), the system MUST NOT display a failure message to the user — it MUST continue silently or with an informational note.
- **FR-006**: All files created by `sl init` (playbook templates, skill files, metadata YAML) MUST be written correctly on Windows, including those involving file permission calls that are no-ops on Windows.

### Key Entities

- **Post-Init Script**: The embedded `init.sh` shell script that runs after core SpecLedger files are set up. Currently executed directly — incompatible with Windows unless a shell is available.
- **Tool Availability Check**: The runtime check for optional tools (gum, etc.) that currently uses `command -v`, a Unix shell built-in unavailable on Windows as a standalone executable.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: `sl init` exits with code 0 on Windows in 100% of cases where it would succeed on macOS/Linux with equivalent inputs.
- **SC-002**: Zero user-visible error messages related to script execution, permissions, or shell availability appear when `sl init` runs on Windows.
- **SC-003**: The files created by `sl init` on Windows are byte-for-byte identical in content (ignoring line endings) to those created on macOS/Linux for the same project.
- **SC-004**: Tool detection (gum, etc.) produces correct results on Windows in both Command Prompt and Git Bash environments.

### Previous work

- **[135-fix-missing-chmod-x] Fix Executable Permissions for Template Files**: Added `isExecutableFile` helper and set `0755` permissions during file copy. Identified that `os.Chmod` is a no-op on Windows for execution bits — relevant context for understanding the permission-setting approach in `applyEmbeddedSkills`.
- **[135-fix-missing-chmod-x] Integration test for sl init --force**: Existing integration test for `sl init --force` that may need Windows variants.

## Dependencies & Assumptions

- **Assumption**: The primary failure mode is `runPostInitScript` in `pkg/cli/commands/bootstrap_helpers.go` — it calls `exec.Command(tmpFile.Name())` on a `.sh` temp file, which Windows cannot execute without a shell interpreter.
- **Assumption**: The secondary issue is `checkGum()` in `pkg/cli/tui/terminal.go` using `exec.Command("command", "-v", "gum")` — `command` is a Unix shell built-in, not an executable on Windows.
- **Assumption**: Git for Windows ships `sh.exe` and `bash.exe` that can serve as shell interpreters if available; WSL may also provide bash access.
- **Assumption**: The post-init script (`init.sh`) content is minimal (prints success message, exports env vars) — its logic could alternatively be inlined in Go to avoid shell dependency entirely.
- **No external spec dependency needed** for this fix — all required context is within the codebase.
