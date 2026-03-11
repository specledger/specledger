# Revision Log: 127-specledger-scheduler-push-strategy

## Revision Session 2026-03-10

### Cluster: Background Execution Safety & Feedback

**Comments addressed**: Comment 1 (race condition / locking), Comment 2 (result delivery / notification)

#### Options Proposed

**Locking mechanism (Comment 1)**:
1. **File lock with PID** — `.specledger/exec.lock` with PID check + stale lock cleanup ✅ Selected
2. **Status-field guard only** — Use spec.md Status field ("Implementing") as the guard
3. **Both lock file + status field** — Belt-and-suspenders approach

**Result delivery (Comment 2)**:
1. **Auto-commit to sub-branch** — Commit to `<feature-branch>/implement`, working tree untouched ✅ Selected
2. **Unstaged changes + notification** — Write to working tree, developer stages manually
3. **Auto-commit to current branch** — Commit directly with `[sl-impl]` prefix

#### Changes Applied

| File | Change | Comments Addressed |
|------|--------|--------------------|
| spec.md | Revised FR-007: concrete lock-file-with-PID mechanism | Comment 1 |
| spec.md | Added FR-015: stale lock detection and cleanup | Comment 1 |
| spec.md | Added FR-016: sub-branch commit + result summary | Comment 2 |
| spec.md | Added SC-006: developer reviews via git diff on sub-branch | Comment 2 |
| spec.md | Updated User Story 1 scenarios 3-5: lock-based guard + stale lock + sub-branch delivery | Comments 1 & 2 |

## Revision Session 2026-03-11

### Cluster: Execution Lock Robustness & Error Recovery

**Comments addressed**: Comment 1 (lock reset command), Comment 2 (error handling strategy), Comment 3 (stale lock detection flow)

#### Options Proposed

**Lock Recovery (Comments 1 & 3)**:
1. **Auto-detect + manual fallback** — PID signal 0 check on every invocation + `sl lock reset` for manual override
2. **Auto-detect + timeout expiry** — PID check + 30min max-age, fully automatic, no user command
3. **Manual-only via `sl lock reset`** — No auto stale detection; if lock exists, execution blocked; user runs `sl lock reset` to clear ✅ Selected

**Error Strategy (Comment 2)**:
1. **Silent log + exit 0** — Never block push, all errors logged to push-hook.log, `sl hook status` shows last 5 errors ✅ Selected
2. **Stderr warning + exit 0** — Same but also prints one-line warnings to terminal during push

#### Changes Applied

| File | Change | Comments Addressed |
|------|--------|--------------------|
| research.md §4 | Rewrote Execution Lock Strategy: removed auto stale detection, added `sl lock reset` and `sl lock status` commands, updated alternatives considered | Comment 1, 3 |
| plan.md Phase 3 | Expanded lock management with manual recovery flow (`sl lock reset`, `sl lock status`) | Comment 3 |
| plan.md Phase 3 | Added full error handling strategy: never block push, edge case table (missing binary, malformed spec, lock contention, spawn failure), `sl hook status` error display | Comment 2 |
| plan.md Source Code | Added `lock.go` to `pkg/cli/commands/` in project structure | Comment 1 |
