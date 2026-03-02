# Quickstart: Verifying Silent Session Capture

## Verify the Fix

### Test 1: No Credentials (should be silent)

1. Rename/remove `~/.specledger/credentials.json`
2. Make a git commit via Claude Code
3. Verify: zero output from `sl session capture` on stderr

### Test 2: No Project ID (should be silent)

1. Ensure `~/.specledger/credentials.json` exists with valid content
2. Work in a repo without `specledger.yaml` or without `project.id`
3. Make a git commit via Claude Code
4. Verify: zero output from `sl session capture` on stderr

### Test 3: Upload Failure (should log error)

1. Ensure credentials + project ID are set up
2. Disconnect network or use invalid token
3. Make a git commit via Claude Code
4. Verify: "Session queued for upload" message appears on stderr

### Test 4: Happy Path (should work as before)

1. Full setup: credentials + project ID + network
2. Make a git commit via Claude Code
3. Verify: "Session captured: ..." message appears on stderr

## Run Unit Tests

```bash
make test
```
