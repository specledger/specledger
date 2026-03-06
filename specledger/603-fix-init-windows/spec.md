# Feature Specification: Fix sl init on Windows

**Feature Branch**: `603-fix-init-windows`
**Created**: 2026-03-05
**Status**: Draft
**Input**: User description: "fix bug not sl init not work in windows (but work in macos and linux)"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Windows User Runs sl init Successfully (Priority: P1)

A developer on Windows runs `sl init` in an existing repository. Currently the command fails with "failed to read manifest: open templates\manifest.yaml: file does not exist", then "failed to create project metadata: playbook name is required". The fix makes `sl init` complete successfully with the same outcome as on macOS and Linux.

**Why this priority**: This is the core bug — `sl init` is completely broken on Windows for every user. No workaround exists.

**Independent Test**: Run `sl init` in any directory on a Windows machine. The command should exit with code 0, print "SpecLedger Initialized", and create `specledger/specledger.yaml`, `.specledger/`, and `.claude/` directories.

**Acceptance Scenarios**:

1. **Given** a Windows machine with `sl` installed and a git repository, **When** the user runs `sl init`, **Then** the command completes without any error about manifest files or playbook names.
2. **Given** the same scenario, **When** `sl init` completes, **Then** all expected files (`specledger.yaml`, playbook templates, skill files) are present in the correct directories.
3. **Given** the same scenario, **When** `sl init` runs, **Then** the "Copying Playbooks" phase succeeds and reports the playbook was applied, not a warning about failure.

---

### User Story 2 - Post-Init Script Runs or Skips Gracefully on Windows (Priority: P2)

After fixing the manifest loading, `sl init` may attempt to run the embedded `init.sh` post-init script. On Windows without a Unix shell, this should either run using an available shell (Git Bash, WSL) or skip silently — it must not crash or show a confusing error.

**Why this priority**: Once the manifest bug is fixed, this is the next failure point. The post-init script is a bash script that Windows cannot execute directly.

**Independent Test**: Run `sl init` on a Windows machine. After seeing "SpecLedger Initialized", verify no error messages appear from the post-init phase.

**Acceptance Scenarios**:

1. **Given** a Windows machine without Git Bash or WSL, **When** `sl init` reaches the post-init script phase, **Then** it skips the script silently without printing an error to the user.
2. **Given** a Windows machine with Git Bash installed, **When** `sl init` reaches the post-init script phase, **Then** it executes the script using `bash.exe` and completes successfully.

---

### Edge Cases (Acceptance Scenarios)

1. **Given** a Windows machine where the user's temp directory path contains spaces (e.g., `C:\Users\My Name\AppData\Local\Temp`), **When** the user runs `sl init`, **Then** the command completes successfully with exit code 0. *(Go's standard library handles paths with spaces correctly. No special handling required — this is a verification item, not a code change.)*
2. **Given** a Windows machine with multiple terminals available (Git Bash, PowerShell, Command Prompt), **When** the user runs `sl init` from any of these terminals, **Then** the core init output (created files and directories) is identical across all terminals. *(The core fix — embed.FS paths — is terminal-independent. The post-init script phase follows the FR-006 decision algorithm.)*
3. **`sl init --force` on an already-initialized project**: **Out of scope** for this fix. Behavior should be identical to current macOS/Linux `--force` behavior once the path fix is applied. If issues arise, they should be tracked as a separate issue.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: All path construction for lookups within the embedded filesystem (`embed.FS`) MUST use forward slashes, regardless of the host operating system.
- **FR-002**: The manifest file (`templates/manifest.yaml`) MUST be found and loaded successfully on Windows.
- **FR-003**: Playbook files MUST be copied to the target project directory successfully on Windows.
- **FR-004**: The metadata file (`specledger.yaml`) MUST be created successfully — this requires a valid playbook name, which depends on FR-002 and FR-003 succeeding.
- **FR-005**: The post-init script phase MUST NOT crash `sl init` on Windows. If no Unix shell is available, the phase MUST be skipped gracefully without a user-visible error.
- **FR-006**: If a Unix shell (`bash`, `sh`) is available on Windows (e.g., via Git for Windows), the post-init script SHOULD execute using that shell as the interpreter.

**FR-006 Decision Algorithm (Windows only):**

1. Look for `bash.exe` on PATH (covers Git for Windows / Git Bash).
2. If not found, look for `wsl.exe` on PATH and invoke `wsl bash -c <script>`.
3. If neither found, skip the post-init script silently (log at debug level only).

Rationale: Git Bash is the most common Unix shell on Windows developer machines. WSL requires more setup and may have filesystem path differences, so it is tried second.

### Key Entities

- **Embedded Filesystem (`embed.FS`)**: Go's embedded virtual filesystem. It always uses forward slashes (`/`) as path separators, regardless of the OS. Using `filepath.Join` to construct paths for it on Windows produces backslash paths that cause "file does not exist" errors.
- **Post-Init Script**: The embedded `init.sh` bash script executed after core setup. Currently invoked via `exec.Command(tmpFile.Name())` which fails on Windows without a shell interpreter.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: `sl init` exits with code 0 on Windows in 100% of cases where it would succeed on macOS/Linux with equivalent inputs.
- **SC-002**: The error messages "failed to read manifest" and "playbook name is required" no longer appear when running `sl init` on Windows.
- **SC-003**: The files created by `sl init` on Windows are functionally identical to those created on macOS/Linux for the same project.
- **SC-004**: No regression on macOS or Linux — existing behavior is preserved.

### Previous work

- **[135-fix-missing-chmod-x] Fix Executable Permissions for Template Files**: Added `isExecutableFile` helper and set permissions during file copy. Uses `filepath.Join` for embedded FS paths — same class of bug but did not manifest as a failure because the copy logic was adjusted separately.

## Dependencies & Assumptions

- **Root cause confirmed**: `filepath.Join` in `pkg/cli/playbooks/manifest.go:12`, `pkg/cli/playbooks/embedded.go:90,97`, and `pkg/cli/commands/bootstrap_helpers.go:449` produces OS-native path separators (backslash on Windows) when constructing paths for `embed.FS` lookups. `embed.FS` requires forward slashes on all platforms.
- **Fix**: Replace `filepath.Join` with `path.Join` (from the `path` package, not `path/filepath`) wherever the result is used as an `embed.FS` path. OS-native path construction (for real filesystem writes) should continue using `filepath.Join`.

**Fix pattern** (before/after):

```go
// BEFORE (broken on Windows)
p := filepath.Join("templates", name)
data, err := templateFS.ReadFile(p)

// AFTER (works on all platforms)
p := path.Join("templates", name)
data, err := templateFS.ReadFile(p)
```

**Affected locations** (exhaustive — only `embed.FS` paths need fixing):

- `pkg/cli/playbooks/manifest.go:12` — manifest lookup
- `pkg/cli/playbooks/embedded.go:90` — template file read
- `pkg/cli/playbooks/embedded.go:97` — template file read
- `pkg/cli/commands/bootstrap_helpers.go:449` — bootstrap path construction

Other uses of `filepath.Join` that write to the real filesystem are correct and should NOT be changed.
- **Secondary issue**: The post-init script is run via `exec.Command(tmpFile.Name())` where `tmpFile` is a `.sh` file. Windows cannot execute shell scripts directly; the fix should detect available shells or skip gracefully.
- **No external spec dependency needed** — all required context is within the codebase.
