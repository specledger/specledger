# Session Capture Fixes - Summary

## Problems Identified

1. `sl session capture` hung indefinitely when run manually without stdin input
2. No validation to check if stdin was available before attempting to read
3. No way to test the session capture flow manually without setting up full hook integration

## Implemented Solutions

### 1. Added Timeout for Stdin Reads ✅

**File**: `pkg/cli/session/capture.go`

**Changes**:
- Replaced blocking `os.ReadFile("/dev/stdin")` with timeout-based goroutine
- Added 5-second timeout using `select` with `time.After()`
- If no data arrives within 5 seconds, returns clear error message

**Code**:
```go
resultChan := make(chan readResult, 1)
go func() {
    data, err := io.ReadAll(os.Stdin)
    resultChan <- readResult{data: data, err: err}
}()

select {
case result := <-resultChan:
    // Process data
case <-time.After(5 * time.Second):
    return &CaptureResult{Error: fmt.Errorf("timeout waiting for stdin input (waited 5 seconds)")}
}
```

### 2. Added Stdin Availability Check ✅

**File**: `pkg/cli/session/capture.go`

**Changes**:
- Added check using `os.Stdin.Stat()` to detect if stdin is a terminal (TTY)
- If stdin is a TTY (meaning no pipe/redirect), immediately fails with helpful error
- Error message suggests using `--test-mode` for manual testing

**Code**:
```go
stat, err := os.Stdin.Stat()
if err != nil {
    return &CaptureResult{Error: fmt.Errorf("failed to stat stdin: %w", err)}
}

if (stat.Mode() & os.ModeCharDevice) != 0 {
    return &CaptureResult{
        Error: fmt.Errorf("no input provided: this command reads hook JSON from stdin (use --test-mode for manual testing)"),
    }
}
```

### 3. Created Test Mode ✅

**Files**:
- `pkg/cli/session/capture.go` (added `CaptureTestMode()` function)
- `pkg/cli/commands/session.go` (added `--test-mode` flag)

**Features**:
- Interactive validation of session capture prerequisites
- Checks for:
  - Git repository
  - Project ID in specledger.yaml
  - Authentication status
  - Claude Code sessions directory
  - Active transcript files
- Provides step-by-step feedback with ✓ checkmarks
- Shows helpful next steps if setup is incomplete

**Usage**:
```bash
sl session capture --test-mode
```

**Example Output**:
```
Running in test mode...
✓ Git repository detected
✓ Project ID found: 7109364a-2ebc-451f-b052-2fbe5453459e
✓ Authenticated as: user@example.com
✓ Claude Code sessions directory found
✓ Found transcript: /Users/user/.claude/sessions/abc123/transcript.jsonl
✓ Session ID: abc123

📝 Simulating git commit hook...

⚠️  Test mode simulates the capture flow but won't create a real session.
To capture a real session, make a git commit while using Claude Code.

✅ Session capture system is configured correctly!

Next steps:
  1. Work on your code with Claude Code
  2. Make a git commit
  3. Session will be automatically captured

To view sessions: sl session list
```

## Testing Results

### Before Fixes:
```bash
$ sl session capture
# Hangs indefinitely... 😞
```

### After Fixes:

**1. Running without stdin (immediate failure):**
```bash
$ sl session capture
Session capture warning: no input provided: this command reads hook JSON from stdin (use --test-mode for manual testing)
# Returns immediately! ✅
```

**2. Running with piped input (works correctly):**
```bash
$ echo '{"session_id":"test","transcript_path":"/path/to/transcript",...}' | sl session capture
Session capture warning: transcript not found: stat /path/to/transcript: no such file or directory
# Processes input and validates! ✅
```

**3. Running in test mode (interactive validation):**
```bash
$ sl session capture --test-mode
Running in test mode...
✓ Git repository detected
✓ Project ID found: 7109364a-2ebc-451f-b052-2fbe5453459e
Session capture warning: not authenticated. Run: sl auth login
# Validates setup and provides guidance! ✅
```

## Files Modified

1. `pkg/cli/session/capture.go` - Added stdin validation, timeout, and test mode
2. `pkg/cli/commands/session.go` - Added --test-mode flag and updated help text

## Additional Benefits

- **Better UX**: Clear error messages guide users instead of silent hangs
- **Faster feedback**: Fails in <1 second instead of hanging forever
- **Debuggability**: Test mode helps diagnose setup issues
- **Documentation**: Updated help text explains proper usage

## Backward Compatibility

✅ **Fully backward compatible**
- Hook-based calls (with stdin) work exactly as before
- Only manual invocations without stdin now fail fast with helpful errors
- Test mode is opt-in via flag
