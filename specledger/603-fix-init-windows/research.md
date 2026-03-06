# Research: Fix sl init on Windows

**Feature**: `603-fix-init-windows` | **Date**: 2026-03-05

## Prior Work

- **[135-fix-missing-chmod-x]**: Added `isExecutableFile` helper + `0755` perms during file copy. Touched `copy.go` and `bootstrap_helpers.go` but only for permission handling, not path separators. Confirms these files are the right place to fix.
- **[591-issue-tracking-upgrade]**: Removed beads setup from `init.sh`. Current `init.sh` only prints a success message and exports env vars — safe to skip on Windows without user impact.

## Decision: `path.Join` vs `filepath.Join` for `embed.FS`

**Decision**: Use `path.Join` (stdlib `path` package) for all paths passed to `embed.FS` methods (`ReadFile`, `Stat`, `ReadDir`). Keep `filepath.Join` for real filesystem paths (output directory creation, temp files, etc.).

**Rationale**: Go's `embed.FS` is a virtual filesystem with a contract that all paths use `/` as separator, regardless of the OS. From the Go spec: "An FS may be implemented by the operating system but also by in-memory file systems or by other systems." The `embed` package explicitly states paths use `/`. Using `filepath.Join` on Windows produces `\`-separated paths which `embed.FS` does not recognize.

**Alternatives considered**:
- `strings.ReplaceAll(path, "\\", "/")`: treats symptom; fragile if called repeatedly
- OS-specific code branches with `runtime.GOOS`: more code to maintain; `path.Join` already solves it

## Decision: Remove `gum` dependency, use `huh` library

**Decision**: Remove all `gum` (external CLI binary) usage. The project already uses `charmbracelet/huh` (Go library) for interactive prompts in `bootstrap_init.go` and `revise.go`. The `gum`-related code was dead:
- `pkg/cli/dependencies/` package was never imported — deleted entirely.
- `tui/terminal.go` had `checkGum()`/`IsGumAvailable()` that were never called — removed.

**Rationale**: `huh` is an embedded Go library providing the same Charmbracelet interactive prompts as `gum`, without requiring users to install an external binary. The codebase had already migrated to `huh` for actual prompt usage; the `gum` references were leftover dead code.

## Decision: Windows shell detection for `runPostInitScript`

**Decision**: On Windows, search for `bash`/`sh` via `exec.LookPath`; use as interpreter if found; skip gracefully if not.

**Rationale**: The `init.sh` content is minimal (5 lines of output). Skipping it on Windows without a shell has no functional impact. Users with Git for Windows (which ships `bash.exe`) get the full experience. This is better than porting the script to Go (unnecessary complexity) or requiring bash installation.

## Affected Code Map

```
embed.FS path bug (use path.Join):
  pkg/cli/playbooks/manifest.go:12          ← manifest path for ReadFile
  pkg/cli/playbooks/embedded.go:90          ← manifest path for Exists
  pkg/cli/playbooks/embedded.go:97          ← playbook path for Exists
  pkg/cli/playbooks/copy.go:19              ← srcPath for walk comparisons
  pkg/cli/playbooks/copy.go:57              ← filepath.Rel → strings.TrimPrefix
  pkg/cli/commands/bootstrap_helpers.go:449 ← init.sh path for ReadFile

Dead gum code (removed):
  pkg/cli/dependencies/registry.go          ← entire package deleted (never imported)
  pkg/cli/tui/terminal.go                   ← removed checkGum(), IsGumAvailable(), gumAvailable field

Windows shell execution:
  pkg/cli/commands/bootstrap_helpers.go:483 ← exec.Command on .sh file
```

## No External Dependencies

All fixes use Go stdlib only (`path`, `runtime`, `os/exec`). No new packages required.
