# Research: Push-Triggered Scheduler Strategy

**Feature**: 127-specledger-scheduler-push-strategy
**Date**: 2026-03-10

## Prior Work

### Feature 010: Checkpoint Session Capture (Complete)
- Established Claude Code hook infrastructure in `pkg/cli/hooks/claude.go`
- Pattern: load settings JSON -> install/uninstall hooks -> check status
- Uses `~/.claude/settings.json` for Claude Code PostToolUse hooks
- **Relevance**: Similar install/uninstall/status pattern, but this feature targets git hooks (`.git/hooks/`) instead of Claude Code hooks

### Feature 598: SDD Workflow Streamline (Draft)
- Defined 4-layer architecture: Hooks (L0) -> CLI (L1) -> AI Commands (L2) -> Skills (L3)
- Hook management pattern: `sl auth hook --install`
- **Relevance**: Establishes the conceptual layer model. Push hooks are L0 (invisible, event-driven). The `sl approve` command is L1 (CLI).

### Existing Hook Infrastructure
- `pkg/cli/hooks/claude.go` (281 lines) - manages Claude Code hooks only
- `pkg/cli/commands/auth.go` - exposes `sl auth hook --install/--remove`
- No existing git hook management code

### Existing Command Patterns
- Commands follow Cobra pattern in `pkg/cli/commands/*.go`
- Each file defines `Var<Name>Cmd` + subcommands + `init()` registration
- Root command in `cmd/sl/main.go` registers via `rootCmd.AddCommand()`
- No existing `sl approve` or `sl implement` CLI commands

## Research Topics

### 1. Git Pre-Push Hook Mechanics

**Decision**: Use `pre-push` git hook to detect approved specs and spawn `sl implement` as a detached background process.

**Rationale**:
- `pre-push` runs before data is transmitted to remote, receives remote name and URL on stdin plus refs on stdin
- It can spawn background processes without blocking the push (using `nohup` + `&` or `os/exec.Cmd.Start()` with process group detach)
- No `post-push` hook exists in git, so `pre-push` with background spawn is the correct approach
- The hook script itself must exit 0 to not block the push

**Alternatives considered**:
- `post-commit` hook: Too early - triggers on every commit, not just push. Would cause unnecessary implementation runs.
- CI/CD webhook: Requires server-side infrastructure; SpecLedger is a local-first tool.
- File watcher (fsnotify): Continuous process overhead; less intuitive trigger than push.

**Implementation approach**:
```bash
#!/bin/sh
# .git/hooks/pre-push
# SpecLedger push-triggered implementation
nohup sl hook execute --event pre-push "$@" > /dev/null 2>&1 &
```
The actual logic lives in Go (`sl hook execute`), not in the shell script. This keeps the hook script minimal and testable.

### 2. Hook Installation Strategy (Preserving Existing Hooks)

**Decision**: Append a SpecLedger marker block to existing pre-push hook, or create new one if none exists.

**Rationale**:
- Many projects already have pre-push hooks (linting, test runners)
- Replacing would break existing workflows
- Using marker comments (`# BEGIN SPECLEDGER` / `# END SPECLEDGER`) allows clean install/uninstall

**Alternatives considered**:
- Husky-style hook runner: Overkill for a single hook integration
- Symlink-based: Fragile, doesn't compose with other hooks
- `.git/hooks/pre-push.d/` directory approach: Not standard git, requires a dispatcher script

**Implementation approach**:
```go
// Install: read existing hook, append marked block, write back, chmod +x
// Uninstall: read hook, remove marked block, write back (remove file if empty)
// Status: check if marked block exists in hook file
```

### 3. Background Process Detachment on macOS/Linux

**Decision**: Use Go's `os/exec` with `SysProcAttr{Setpgid: true}` + redirect stdout/stderr to log file.

**Rationale**:
- `Setpgid: true` creates a new process group so the child survives parent exit
- Works on both macOS (Darwin) and Linux
- The shell script wrapper uses `nohup ... &` as a belt-and-suspenders approach
- Log file at `.specledger/logs/push-hook.log` captures all output

**Alternatives considered**:
- `syscall.SysProcAttr{Setsid: true}`: Creates new session; works but Setpgid is sufficient
- Go goroutine (no subprocess): Would block the hook; push would hang

### 4. Execution Lock Strategy

**Decision**: Simple lock file at `.specledger/exec.lock` with manual recovery via `sl lock reset`.

**Rationale**:
- `gofrs/flock` already used in issue store for JSONL locking
- Lock file contains JSON: `{"pid": 12345, "feature": "127-specledger-scheduler-push-strategy", "started_at": "..."}`
- Lock created by `sl implement`, not by the hook script
- No automatic stale lock detection — keeps implementation simple and predictable

**Lock handling behavior**:
- On `sl hook execute` / `sl implement`: if `exec.lock` exists, skip execution and log message
- No PID checking or timeout-based expiry

**Manual recovery commands**:
- `sl lock reset` — removes `exec.lock` unconditionally; user runs this when a lock is left behind after a crash or kill
- `sl lock status` — displays current lock info (PID, feature, started_at) or reports no active lock

**Alternatives considered**:
- Advisory file lock only (no PID): Can't inspect which process holds the lock
- Auto-detect stale via PID + signal 0: Adds complexity; PID recycling can cause false positives
- Timeout-based expiry: Arbitrary timeout may be too short for long implementations or too long for crashes
- Socket-based lock: Unnecessary complexity for single-process coordination
- Database lock: No database in SpecLedger architecture

### 5. Sub-Branch Commit Strategy

**Decision**: Use go-git (already a dependency) to commit to `<feature-branch>/implement` without modifying the working tree.

**Rationale**:
- go-git/v5 is already used throughout the project for git operations
- Can create commits on a branch without checkout using in-memory index
- Developer's working tree stays clean; they review via `git diff <feature>..<feature>/implement`

**Implementation approach**:
- Create/checkout orphan-like branch from current HEAD
- Use worktree-less commit (go-git supports this via plumbing)
- Or: use `git worktree add` to a temp dir, commit there, remove worktree

**Alternatives considered**:
- Stash/checkout/commit/checkout-back: Risky; could lose uncommitted work
- Patch file output: Less integrated; harder to review

### 6. Spec Status Field Parsing

**Decision**: Parse `**Status**: <value>` from spec.md using simple regex/string matching.

**Rationale**:
- The status is a well-defined markdown pattern already used by SpecLedger
- `internal/spec/` package already has spec parsing capabilities
- No need for a full markdown parser; line-by-line scan is sufficient

**Implementation**:
```go
// Pattern: **Status**: Approved
// Read spec.md, find line matching `**Status**:`, extract value
// For approve: replace "Draft" with "Approved"
```
