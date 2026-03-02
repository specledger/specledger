# Research: Silent Session Capture

**Feature**: 598-silent-session-capture
**Date**: 2026-03-02

## Prior Work

- **010-checkpoint-session-capture**: Original session capture implementation. Established the hook → capture → upload → queue flow. This fix modifies the early-exit behavior of that flow.
- No related issues found in issue tracker for this specific bug.

## Research Findings

### R1: Current Error Behavior Analysis

**Decision**: The current `Capture()` function has two problematic early-exit paths that produce stderr output when they shouldn't.

**Current flow order** (in `capture.go` `Capture()`):
1. Check git commit → silent skip (OK)
2. Check tool success → silent skip (OK)
3. Get commit hash/branch → error if fails (OK, git broken)
4. **Get project ID → prints 3-line warning to stderr** (PROBLEM)
5. Get transcript → graceful degradation (OK)
6. **Load credentials → sets result.Error** (PROBLEM)
7. Build, compress, upload → errors/queue (OK)

**Problem**: Steps 4 and 6 treat "user hasn't set up" the same as "something broke". These are not errors - they are expected states for users who haven't opted in.

### R2: Reordering - Credentials Before Project ID

**Decision**: Move credentials check (step 6) to before project ID check (step 4).

**Rationale**:
- Credentials check is a pure local file read (fast, no network)
- `GetProjectIDFromRemote()` internally calls `auth.GetValidAccessToken()` anyway, so it needs credentials
- Fail-fast: if no credentials, skip everything else immediately
- No functional change for the happy path

**Alternatives considered**:
- Keep order, just suppress output → Wasteful (still attempts project ID lookup without credentials)
- Add `Silent` flag to CaptureResult → Over-engineering for this case

### R3: Silent Skip Strategy

**Decision**: Use nil error returns for expected skip cases. Return `result` with `Captured=false` and `Error=nil`.

**Rationale**:
- In `commands/session.go`, `runSessionCapture()` only prints to stderr when `result.Error != nil`
- If we return nil error, the command handler already does the right thing (silent exit code 0)
- No changes needed in the command handler
- This is idiomatic Go: nil error = no problem

**Alternatives considered**:
- Add `Silent bool` field to `CaptureResult` → Adds complexity, new field to maintain
- Log at debug level instead of warning → No debug logging infrastructure exists in the project

### R4: Test Coverage Gap

**Decision**: Add unit tests specifically for the silent-skip paths.

**Rationale**:
- Current `capture_test.go` only tests helper functions (`IsGitCommit`, `ParseHookInput`, etc.)
- No tests cover the `Capture()` function directly
- The new behavior (silent skip) must be verified by tests
- Tests need to mock `auth.LoadCredentials()` and `GetProjectID()` since they depend on filesystem

**Challenge**: `Capture()` directly calls `auth.LoadCredentials()` and filesystem functions, making it hard to unit test without refactoring. Options:
1. Test via integration (set up temp dirs, credential files) - matches existing project patterns
2. Test the behavioral contract: given specific env, verify no stderr output

### R5: Stderr Output Removal

**Decision**: Remove all `fmt.Fprintf(os.Stderr, ...)` calls in the no-credentials and no-project-ID paths of `Capture()`.

**Lines to change**:
- Lines 246-248: Remove 3 `fmt.Fprintf(os.Stderr, ...)` calls for project ID error
- Lines 290-291: Remove `result.Error = fmt.Errorf(...)` for credentials error

**What stays**: Stderr output for actual failures (upload fail, metadata fail, queue fail) remains unchanged. The `runSessionCapture` handler in commands/session.go needs no changes.
