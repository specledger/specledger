# Tasks Index: Push-Triggered Scheduler Strategy

Issue Graph Index into the tasks and phases for this feature implementation.
This index does **not contain tasks directly**—those are fully managed through `sl issue` CLI.

## Feature Tracking

* **Epic ID**: `SL-022cfd`
* **User Stories Source**: `specledger/127-specledger-scheduler-push-strategy/spec.md`
* **Research Inputs**: `specledger/127-specledger-scheduler-push-strategy/research.md`
* **Planning Details**: `specledger/127-specledger-scheduler-push-strategy/plan.md`
* **Data Model**: `specledger/127-specledger-scheduler-push-strategy/data-model.md`
* **Contract Definitions**: `specledger/127-specledger-scheduler-push-strategy/contracts/`

## Issue Query Hints

```bash
# Find all open tasks for this feature
sl issue list --label "spec:127-specledger-scheduler-push-strategy" --status open

# See ready tasks (no blockers)
sl issue ready

# View issue details
sl issue show <issue-id>

# Filter by story
sl issue list --label "story:US0"
sl issue list --label "story:US1"
sl issue list --label "story:US2"
sl issue list --label "story:US3"
```

## Tasks and Phases Structure

```
Epic: SL-022cfd (Push-Triggered Scheduler Strategy)
│
├── Phase 1 - Setup: SL-c85ab6
│   ├── T001: SL-267e9b - Create scheduler and spec package structure
│   └── T002: SL-7564e2 - Register new command stubs in main.go
│
├── Phase 2 - Foundational: SL-b8fbc1 (blocked by Setup)
│   ├── T003: SL-ecb00a - Implement spec status read/write helpers [P]
│   ├── T004: SL-28c8b7 - Implement ExecutionLock management [P]
│   └── T005: SL-256286 - Implement approved spec detection logic (blocked by T003)
│
├── Phase 3 - US0 Approval Command (P1): SL-66050d (blocked by Foundational)
│   └── T006: SL-31a2dc - Implement sl approve command (blocked by T003)
│
├── Phase 4 - US2 Hook Management (P2): SL-997e06 (blocked by Foundational)
│   ├── T007: SL-bfc36b - Implement git hook install/uninstall/status logic
│   └── T008: SL-c134ce - Implement sl hook commands (blocked by T007)
│
├── Phase 5 - US1 Push-Triggered Execution (P1): SL-de808a (blocked by US0, US2)
│   ├── T009: SL-2a296c - Implement Claude CLI process spawning executor
│   ├── T010: SL-8ec018 - Implement sl implement command (blocked by T004, T009)
│   ├── T011: SL-bfae4a - Implement sl hook execute subcommand (blocked by T004, T005)
│   └── T012: SL-bf7864 - Implement sl lock reset/status commands (blocked by T004)
│
├── Phase 6 - US3 Logging & Observability (P3): SL-adb436 (blocked by US1)
│   ├── T013: SL-9b2048 - Implement structured hook execution log writer
│   └── T014: SL-745262 - Add recent error display to sl hook status (blocked by T013)
│
└── Phase 7 - Polish: SL-47cd1f (blocked by US3)
    ├── T015: SL-c9cf75 - Integration test: approve command workflow
    └── T016: SL-938ba6 - Integration test: hook install/uninstall lifecycle
```

## Convention Summary

| Type    | Description                  | Labels                                 |
| ------- | ---------------------------- | -------------------------------------- |
| epic    | Full feature epic            | `spec:127-specledger-scheduler-push-strategy` |
| feature | Implementation phase / story | `phase:<name>`, `story:<US#>`          |
| task    | Implementation task          | `component:<area>`, `requirement:<FR>` |

## Dependency Graph

```
Setup (SL-c85ab6)
  └──► Foundational (SL-b8fbc1)
         ├──► US0: Approve (SL-66050d) ──┐
         └──► US2: Hook Mgmt (SL-997e06) ┤
                                          ▼
                              US1: Execution (SL-de808a)
                                          │
                                          ▼
                              US3: Logging (SL-adb436)
                                          │
                                          ▼
                              Polish (SL-47cd1f)
```

**Parallel opportunities**:
- T003 (spec status) and T004 (lock) can run in parallel within Foundational
- T007 (githook library) within US2 can start as soon as Foundational completes
- US0 and US2 can run in parallel (both only depend on Foundational)
- T009 (executor) and T012 (lock commands) can run in parallel within US1
- T015 and T016 (integration tests) can run in parallel within Polish

## Implementation Strategy

### MVP Scope (US0 + US2 + US1)
1. **Setup**: Package structure and command stubs
2. **Foundational**: Spec status helpers + execution lock
3. **US0**: `sl approve` command — enables approval gate
4. **US2**: `sl hook install/uninstall/status` — enables hook management
5. **US1**: `sl implement` + `sl hook execute` — enables push-triggered execution

**MVP delivers**: Full push-triggered implementation workflow (approve → push → auto-implement)

### Incremental Delivery
6. **US3**: Logging & observability — adds structured logging and error visibility
7. **Polish**: Integration tests — adds end-to-end verification

## Definition of Done Summary

| Issue ID | DoD Items |
|----------|-----------|
| SL-267e9b | pkg/cli/scheduler/ created, stub files, pkg/cli/spec/status.go, go build succeeds |
| SL-7564e2 | approve.go, hook.go, implement.go, lock.go stubs, registered in main.go, go build succeeds |
| SL-ecb00a | ReadStatus, WriteStatus, handle missing/malformed spec |
| SL-28c8b7 | ExecutionLock struct, AcquireLock, CheckLock, ReleaseLock, error if held, unit tests |
| SL-256286 | DetectApprovedSpec, branch pattern matching, ReadStatus integration, non-feature branch handling, unit tests |
| SL-31a2dc | VarApproveCmd, artifact validation, status update, error messages, auto-detect, unit tests |
| SL-bfc36b | InstallPushHook, UninstallPushHook, HasPushHook, preserve existing hooks, chmod +x, unit tests |
| SL-c134ce | hookInstallCmd, hookUninstallCmd, hookStatusCmd, exit codes, unit tests |
| SL-2a296c | SpawnClaudeCLI, exec.LookPath, Setpgid, log redirect, unit tests |
| SL-8ec018 | VarImplementCmd, lock acquire, Claude spawn, lock release, result summary, error handling, exit code passthrough |
| SL-bfae4a | hookExecuteCmd, detector integration, lock check, detached spawn, always exit 0, logging |
| SL-bf7864 | lockResetCmd, lockStatusCmd, --json output, graceful no-lock handling |
| SL-9b2048 | WriteHookLog, structured format, log rotation (50), auto-create dir, unit tests |
| SL-745262 | ReadRecentErrors, display last 5, graceful empty/missing |
| SL-c9cf75 | Integration tests for all approve scenarios |
| SL-938ba6 | Integration tests for all hook management scenarios |

---

> This file is intentionally light and index-only. Implementation data lives in the issue store. Update this file only to point humans and agents to canonical query paths and feature references.
