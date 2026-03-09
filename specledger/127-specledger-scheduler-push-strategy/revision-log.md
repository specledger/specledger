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
