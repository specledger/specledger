# Implementation Plan: Fix GoReleaser Build Version Injection

**Branch**: `605-fix-goreleaser-ldflags` | **Date**: 2026-03-10 | **Spec**: [spec.md](./spec.md)

## Summary

Fix two bugs preventing released CLI binaries from reporting correct version information: (1) GoReleaser ldflags reference wrong variable names, and (2) `rootCmd.Version` is set before ldflags values are applied.

## Technical Context

**Language/Version**: Go 1.24.2
**Primary Dependencies**: Cobra (CLI), GoReleaser (build/release)
**Storage**: N/A (configuration fix)
**Project Type**: CLI tool
**Target Platform**: Cross-platform (macOS, Linux, Windows)

## Constitution Check

Constitution is uninitialized (template placeholders only). No gates to evaluate.

## Phase 0: Research Summary

### Root Cause Analysis

**Bug 1: Variable Name Mismatch**

GoReleaser `.goreleaser.yaml` (line 47) sets:
```
-X main.version={{.Version}} -X main.commit={{.Commit}} -X main.date={{.Date}} -X main.buildType=release
```

But `cmd/sl/main.go` (lines 15-22) declares:
```go
var (
    buildVersion = "dev"
    buildCommit  = "unknown"
    buildDate    = "unknown"
    buildType    = "development"
)
```

The ldflags target `main.version`, `main.commit`, `main.date` — but the actual Go variables are `main.buildVersion`, `main.buildCommit`, `main.buildDate`. Only `main.buildType` matches correctly. Go's linker silently ignores ldflags that don't match existing variables, so the build succeeds but the variables retain their defaults.

**Bug 2: rootCmd.Version Initialization Timing**

`rootCmd` is declared as a package-level variable with `Version: version.GetVersion()` (line 48). In Go, package-level variable initialization occurs before `init()` runs. The `init()` function (lines 24-30) copies `buildVersion` → `version.Version`, but by that point `rootCmd.Version` has already been set to `"dev"` (the default from `pkg/version/version.go`).

Even after fixing Bug 1, `rootCmd.Version` would still show `"dev"` because:
1. ldflags set `buildVersion = "v1.2.3"` (link time)
2. `rootCmd` is initialized with `Version: version.GetVersion()` → returns `"dev"` (version.Version hasn't been updated yet)
3. `init()` runs: copies `buildVersion` → `version.Version` (too late for rootCmd)

### Decision

**Fix approach**: Two targeted changes in two files.

1. **`.goreleaser.yaml`**: Update ldflags to reference correct variable names (`main.buildVersion`, `main.buildCommit`, `main.buildDate`)
2. **`cmd/sl/main.go`**: Move `rootCmd.Version` assignment into `init()` after version package variables are set

**Alternatives considered**:
- Renaming Go variables to match GoReleaser convention (`version`, `commit`, `date`) — rejected because it would shadow the `version` package import and require broader refactoring
- Using a separate `SetVersion()` function in `main()` — rejected because `init()` already exists and runs at the right time

### Previous Work

- **596-doctor-version-update**: Built the version checking infrastructure (`pkg/version/`) that depends on correct version injection
- **007-release-delivery-fix / migrated spec**: Earlier attempts at fixing the release pipeline that didn't catch this mismatch
- **SL-a68e99**: Originally added `buildVersion` variable to `main.go`

## Phase 1: Design

### Changes Required

No new files, data models, or contracts needed. This is a two-file configuration/code fix.

**File 1: `.goreleaser.yaml`**
- Change ldflags line to use correct variable names
- Before: `-X main.version={{.Version}} -X main.commit={{.Commit}} -X main.date={{.Date}}`
- After: `-X main.buildVersion={{.Version}} -X main.buildCommit={{.Commit}} -X main.buildDate={{.Date}}`

**File 2: `cmd/sl/main.go`**
- Remove `Version: version.GetVersion()` from the `rootCmd` struct literal
- Add `rootCmd.Version = version.GetVersion()` inside `init()` after the version package variables are set

### Verification Strategy

- `goreleaser build --snapshot --clean` to verify ldflags are correctly injected
- Check built binary with `./sl --version` to confirm version output
- Existing `pkg/version` tests continue to pass (they test comparison logic, not injection)

## Phase 2: Work Breakdown

### Task 1: Fix GoReleaser ldflags variable names (FR-001)
- **File**: `.goreleaser.yaml`
- **Change**: Update ldflags to reference `main.buildVersion`, `main.buildCommit`, `main.buildDate`
- **Depends on**: Nothing

### Task 2: Fix rootCmd.Version initialization timing (FR-002)
- **File**: `cmd/sl/main.go`
- **Change**: Move `rootCmd.Version` assignment into `init()` after version package population
- **Depends on**: Nothing (can be done in parallel with Task 1)

### Task 3: Verify with GoReleaser snapshot build (FR-003)
- **Action**: Run `goreleaser build --snapshot --clean` and check binary version output
- **Depends on**: Tasks 1 and 2

## Success Criteria

- [ ] SC-001: `goreleaser build --snapshot` produces a binary where `sl --version` shows the snapshot version, not `dev`
- [ ] SC-002: GoReleaser ldflags in `.goreleaser.yaml` reference `main.buildVersion`, `main.buildCommit`, `main.buildDate` matching the Go variable declarations
- [ ] SC-003: `rootCmd.Version` is set inside `init()` after version package variables are populated
