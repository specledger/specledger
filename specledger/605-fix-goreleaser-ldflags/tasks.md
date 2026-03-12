# Tasks Index: Fix GoReleaser Build Version Injection

Issue Graph Index into the tasks and phases for this feature implementation.
This index does **not contain tasks directly**—those are fully managed through `sl issue` CLI.

## Feature Tracking

* **Epic ID**: `SL-fbf6b2`
* **User Stories Source**: `specledger/605-fix-goreleaser-ldflags/spec.md`
* **Research Inputs**: `specledger/605-fix-goreleaser-ldflags/research.md`
* **Planning Details**: `specledger/605-fix-goreleaser-ldflags/plan.md`

## Issue Query Hints

```bash
# Find all open tasks for this feature
sl issue list --label spec:605-fix-goreleaser-ldflags --status open

# See all issues across specs
sl issue list --all --status open

# View issue details
sl issue show SL-fbf6b2

# Link dependencies
sl issue link [from-id] blocks [to-id]
```

## Tasks and Phases Structure

* **Epic**: SL-fbf6b2 → Fix GoReleaser Build Version Injection
* **Phases**:
  * **US1** (SL-a62939): Released Binary Reports Correct Version
    * SL-fab5cc: Fix GoReleaser ldflags variable names [P1] [FR-001]
    * SL-cc5908: Fix rootCmd.Version initialization timing [P1] [FR-002]
  * **US2** (SL-de702a): Update Checker Detects Correct Installed Version
    * SL-6e33b5: Verify GoReleaser snapshot build produces correct version [P1] [FR-003]

## Dependencies & Execution Order

```
SL-fab5cc (Fix ldflags)  ──┐
                            ├──→ SL-6e33b5 (Verify snapshot build)
SL-cc5908 (Fix rootCmd)  ──┘
```

- **SL-fab5cc** and **SL-cc5908** can be executed **in parallel** (different files, no shared state)
- **SL-6e33b5** is blocked by both SL-fab5cc and SL-cc5908 (verification requires both fixes)

## Definition of Done Summary

| Issue ID   | DoD Items |
|------------|-----------|
| SL-fab5cc  | - ldflags updated to use main.buildVersion<br>- ldflags updated to use main.buildCommit<br>- ldflags updated to use main.buildDate<br>- main.buildType=release unchanged (already correct) |
| SL-cc5908  | - Version field removed from rootCmd struct literal<br>- rootCmd.Version set in init() after version.Version population<br>- sl --version shows correct version in released binary |
| SL-6e33b5  | - goreleaser build --snapshot --clean succeeds<br>- Built binary --version shows snapshot version<br>- Commit hash is populated (not unknown)<br>- Build date is populated (not unknown) |

## Implementation Strategy

### MVP Scope

Both user stories are P1 and the entire fix is 3 tasks across 2 files. This feature is small enough to deliver as a single increment — no phased MVP needed.

**Suggested execution**:
1. Fix both SL-fab5cc and SL-cc5908 in parallel (2 min each)
2. Run SL-6e33b5 verification (goreleaser snapshot build)
3. Done

### Story Testability

- **US1**: Independently testable by building with `go build -ldflags` and checking `--version` output
- **US2**: Automatically resolved by US1 — existing `pkg/version/checker.go` logic works correctly once it receives real version strings

---

> This file is intentionally light and index-only. Implementation data lives in the issue store.
