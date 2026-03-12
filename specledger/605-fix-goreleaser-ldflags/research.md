# Research: Fix GoReleaser Build Version Injection

**Date**: 2026-03-10
**Feature**: 605-fix-goreleaser-ldflags

## Prior Work

- **596-doctor-version-update**: Built `pkg/version/` package with version comparison, GitHub Releases API client, and `sl doctor` version checking. All of this infrastructure works correctly but receives `"dev"` as input due to the ldflags mismatch.
- **007-release-delivery-fix (migrated)**: Earlier release pipeline fix that established the GoReleaser configuration. The ldflags were originally set with `main.version` style naming but the Go variables were later renamed to `buildVersion` style, creating the mismatch.
- **SL-a68e99**: Added `buildVersion` variable to `main.go`, introducing the naming convention that diverged from GoReleaser's ldflags.

## Findings

### Bug 1: ldflags Variable Name Mismatch

| GoReleaser ldflags target | Go variable name | Match? |
|---------------------------|------------------|--------|
| `main.version`            | `buildVersion`   | No     |
| `main.commit`             | `buildCommit`    | No     |
| `main.date`               | `buildDate`      | No     |
| `main.buildType`          | `buildType`      | Yes    |

Go's linker (`-X` flag) silently ignores ldflags that don't match existing package-level variables. No build error occurs — the variables simply retain their default values.

### Bug 2: rootCmd.Version Timing

Go initialization order within a package:
1. Package-level `var` declarations are evaluated (in source order)
2. `init()` functions run (in source order)

Since `rootCmd` is a package-level var with `Version: version.GetVersion()`, it captures the version **before** `init()` copies the ldflags values into the version package. Result: `rootCmd.Version` is always `"dev"`.

### Decision

- **Approach**: Fix both issues with minimal, targeted changes
- **Rationale**: The fix is straightforward — correct the variable names and move the timing. No architectural changes needed.
- **Alternatives considered**:
  - Rename Go variables to match `main.version` convention → rejected (shadows `version` package import)
  - Use `main()` instead of `init()` → rejected (`init()` already exists and is the idiomatic place)
