# Session 03 — Implementation Progress
**Date**: 2026-02-20
**Branch**: `136-revise-comments`
**Agent**: claude-sonnet-4-6
**Scope**: M6 fix (--dry-run) + sl-0r5 edge cases (all 6) + sl-8w6 new task creation + beads DB repair.

---

## What Was Built

### Modified Files

| File | Change |
|------|--------|
| `pkg/cli/commands/revise.go` | M6 fixed; all 6 edge cases implemented; `networkHint` helper added |

### Beads Tasks Closed This Session

```
sl-0r5  Edge case handling (all 6 cases + M6)        ✓
```

### Still Open

```
sl-ndj  Polish epic (blocked by sl-8w6)
sl-0e7  Feature epic (blocked by sl-8w6 + sl-ndj)
sl-8w6  Detect local branch behind remote + pull      ← NEXT (Priority 3, blocked by sl-0r5)
```

---

## Fixes and Edge Cases Implemented

### M6 — `--dry-run` unreachable flag — FIXED

**What was implemented**:
- `editAndConfirmPrompt(initial string, dryRun bool)` — added `dryRun bool` parameter
- After `current = edited` (editor closes), added short-circuit:
  ```go
  if dryRun {
      _, err := writePromptInteractive(current)
      return "", err
  }
  ```
- Call site in `runRevise` updated: `editAndConfirmPrompt(prompt, reviseDryRun)`
- `TODO(M6)` comment removed

### Case 1 — File not found

In `processComments`, after rendering the "File:" line:
```go
if _, statErr := os.Stat(c.FilePath); os.IsNotExist(statErr) {
    fmt.Printf("⚠ File not found locally: %s\n", c.FilePath)
    fileExists = false
}
```
`fileExists` flag gates the selected_text check (case 2).

**Output matches quickstart §"Error Handling"**:
```
⚠ File not found locally: specledger/009-feature-name/old-spec.md
```

### Case 2 — Selected text not in file

In `processComments`, after the file-exists check:
```go
if fileExists && c.SelectedText != "" {
    content, readErr := os.ReadFile(c.FilePath)
    if readErr == nil && !strings.Contains(string(content), c.SelectedText) {
        fmt.Println("⚠ Original selected text not found in current file version")
    }
}
```

**Output matches quickstart §3 "Edge case"**:
```
⚠ Original selected text not found in current file version
```

### Case 3 — Deleted remote branch

**Decision**: Option A (descriptive error, no retry loop) — confirmed by user.

In `checkoutIfNeeded`, `CheckoutRemoteTracking` failure path:
```go
return stashUsed, fmt.Errorf(
    "branch %q not found on remote (it may have been deleted)\nRun `sl revise` again to select a different branch.",
    targetBranch)
```

**Note**: Local-only branches are already handled by the `BranchExists=true → CheckoutBranch` path and never reach `CheckoutRemoteTracking`. This case only fires when the branch doesn't exist locally AND the remote checkout fails.

### Case 4 — Stash failure

In `checkoutIfNeeded`, `StashChanges` error path:
```go
return false, fmt.Errorf("stash failed — resolve manually before switching branches: %w", err)
```

### Case 5 — Network errors

Added `networkHint` helper:
```go
func networkHint(err error) error {
    if err == nil { return nil }
    if strings.Contains(err.Error(), "API error (") { return err }
    return fmt.Errorf("%w\nCheck your network connection and try again.", err)
}
```

Applied to all 6 client call error sites in `fetchComments` and `pickBranchFromAPI`.

API-level errors (contain "API error (NNN)") pass through unchanged. Network errors (connection refused, timeout, DNS failures) get the hint appended.

### Case 6 — Editor not found message

In `editAndConfirmPrompt`, updated the error print:
```go
if strings.Contains(err.Error(), "no editor found") {
    fmt.Println("⚠ No editor found ($EDITOR/$VISUAL not set, vi not available)")
} else {
    fmt.Printf("Editor unavailable (%v).\n", err)
}
```

**Output matches quickstart §4 "Editor not found"**:
```
⚠ No editor found ($EDITOR/$VISUAL not set, vi not available)
```

---

## Decisions Made This Session

### D6 — Case 3 approach: descriptive error vs retry loop

**Divergence identified**: Task `sl-0r5` said "re-show the branch list" on remote checkout failure, but `checkoutIfNeeded` has no access to the API client.

**Options offered** (via AskUserQuestion):

| Option | Description |
|--------|-------------|
| A | Descriptive error only — user re-runs manually ← **chosen** |
| B | Sentinel error + retry loop in `runRevise` — matches spec exactly but more code |

**User decision**: Option A — descriptive error.

### D7 — New task: detect local branch behind remote

**Identified during analysis**: When a local branch is behind its remote tracking branch:
- `BranchExists=true → CheckoutBranch` succeeds silently
- Agent runs on stale content
- `git push` fails with "rejected — non-fast-forward" after all agent work is done

**User decision**: Create new beads task `sl-8w6` in the polish phase, blocked by `sl-0r5`, lowest priority.

**Task created**: `sl-8w6` — "Detect local branch behind remote and offer pull before revise"
- Needs two new helpers in `pkg/cli/git/git.go`: `IsBehindRemote` and `PullBranch`
- Needs changes in `checkoutIfNeeded` (checkout path) and `runRevise` (no-checkout path)
- See `bd show sl-8w6` for full implementation spec

### Beads DB repair

`bd sync` failed mid-session with `sqlite3: database disk image is malformed`. Diagnosis:
- JSONL had 37 issues (correct, includes sl-8w6)
- SQLite had 36 issues (stale, missing sl-8w6)
- WAL files (`.beads/beads.db-shm`, `.beads/beads.db-wal`) were corrupted

**Fix**: Deleted `beads.db` + WAL files, ran `bd import -i .beads/issues.jsonl` to rebuild from JSONL (source of truth). 37 issues imported. `bd sync` succeeded.

---

## Current Command Flow (as implemented end of session03)

Same as session02 with the following changes:

```
Step 8: editAndConfirmPrompt(prompt, reviseDryRun)
         ↑ now takes dryRun bool
         If dryRun: skip menu, call writePromptInteractive directly

Step 3 (checkoutIfNeeded) — error paths:
  StashChanges failure → "stash failed — resolve manually before switching branches: ..."
  CheckoutRemoteTracking failure → "branch X not found on remote (may have been deleted)..."

processComments — per-comment display:
  After "File:" line:
    os.Stat(c.FilePath) → if missing: "⚠ File not found locally: ..."
    os.ReadFile + strings.Contains → if text missing: "⚠ Original selected text not found..."

fetchComments + pickBranchFromAPI — all client call errors:
  wrapped with networkHint() → appends "Check your network connection and try again."
  only when error does not already contain "API error ("

editAndConfirmPrompt — editor not found:
  "⚠ No editor found ($EDITOR/$VISUAL not set, vi not available)"
```

---

## What the Next Agent Must Do

### Priority 1 — sl-8w6: Detect local branch behind remote

This is the only remaining open task in the polish phase. sl-ndj and sl-0e7 cannot be closed until sl-8w6 is done.

Run `bd show sl-8w6` for full implementation spec. Key points:

1. **New helpers** in `pkg/cli/git/git.go`:
   - `IsBehindRemote(repoPath, branch string) (bool, int, error)` — compare local vs cached `origin/<branch>` ref
   - `PullBranch(repoPath string) error` — `exec git pull`

2. **Changes in `commands/revise.go`**:
   - In `checkoutIfNeeded`, after `CheckoutBranch` succeeds: check `IsBehindRemote`, offer `huh.NewConfirm` to pull
   - In `runRevise`, before Step 5 (fetchComments), for the no-checkout path: same check

3. **UX**: Confirm prompt default Yes (`[Y/n]`). Pull failure → clear error, no agent launched. Skip → warning printed.

### Priority 2 — Close epics

After sl-8w6:
```bash
bd close sl-ndj --reason="..."
bd close sl-0e7 --reason="..."
```

### Priority 3 — Session closeout

1. `go build ./...`
2. `go test ./pkg/cli/revise/... ./pkg/cli/git/...`
3. `bd sync`
4. Commit and push
5. Write `specledger/136-revise-comments/sessions/session04.md`

---

## Continuation Prompt

Copy-paste this as the opening message in the next session.

---

**AGENT INSTRUCTIONS — Session 04: sl-8w6 (behind-remote detection) + epic close**

You are continuing implementation of the `136-revise-comments` feature on branch `136-revise-comments` in the specledger CLI (Go 1.24.2 + Cobra + Bubble Tea + Huh + go-git).

### Load context FIRST (required before writing any code)

```
specledger/136-revise-comments/sessions/session03.md   ← decisions, sl-8w6 spec
specledger/136-revise-comments/plan.md                 ← authoritative design reference
pkg/cli/commands/revise.go                             ← current command handler
pkg/cli/git/git.go                                     ← shared git package (add helpers here)
```

Then run:

```bash
bd show sl-8w6
```

### Critical rules (carry forward from session02)

1. **Any divergence from plan.md or a task description MUST be confirmed by the user.** Use `AskUserQuestion` to present options.
2. **Use `AskUserQuestion` whenever you would otherwise present a list of options in plain text.**
3. **Do not mark any beads task as closed until `go build ./...` passes.**
4. **Do not implement beyond what was asked.**

### Work to complete

**Priority 1 — sl-8w6**: Add `IsBehindRemote` + `PullBranch` to `pkg/cli/git/git.go`. Add behind-remote check + pull confirm in `checkoutIfNeeded` (checkout path) and `runRevise` (no-checkout path). See `bd show sl-8w6` for full UX and acceptance criteria.

**Priority 2 — Close epics**:
```bash
bd close sl-ndj --reason="..."
bd close sl-0e7 --reason="..."
```

**Priority 3 — Session closeout**:
1. `go build ./...`
2. `go test ./pkg/cli/revise/... ./pkg/cli/git/...`
3. `bd sync`
4. Commit and push
5. Write `specledger/136-revise-comments/sessions/session04.md`
