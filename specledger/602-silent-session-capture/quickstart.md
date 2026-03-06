# Quickstart: Verifying Silent Session Capture

## Setup

```bash
make build && make install
```

Then run `sl bootstrap` in your project to copy the new `/specledger.commit` command.

## Test 1: /specledger.commit with Auth + Project ID (happy path)

1. Ensure `~/.specledger/credentials.json` exists with valid tokens
2. Ensure `specledger.yaml` has a `project.id` field
3. Stage some changes: `git add .`
4. Run `/specledger.commit` in Claude Code (or type "commit giúp tôi")
5. **Verify**: Commit succeeds, session captured, pushed to remote

## Test 2: /specledger.commit without Auth (silent skip)

1. Rename credentials file: `mv ~/.specledger/credentials.json ~/.specledger/credentials.json.bak`
2. Stage some changes: `git add .`
3. Run `/specledger.commit`
4. **Verify**: Commit succeeds, pushed to remote, ZERO warnings on stderr
5. Restore: `mv ~/.specledger/credentials.json.bak ~/.specledger/credentials.json`

## Test 3: /specledger.commit with Auth but no Project ID (silent skip)

1. Ensure credentials exist
2. Remove project.id from specledger.yaml (or work in a repo without it)
3. Stage some changes and run `/specledger.commit`
4. **Verify**: Commit succeeds, pushed to remote, ZERO warnings on stderr

## Test 4: Upload Failure (error logged locally + sent to Sentry)

1. Ensure credentials + project ID exist
2. Set Supabase URL to invalid (or disconnect network after commit)
3. Stage changes and run `/specledger.commit`
4. **Verify**:
   - Commit and push still succeed
   - Session is queued locally
   - Error written to `~/.specledger/capture-errors.log`
   - Error appears in Sentry dashboard (if network reachable)

## Test 5: PostToolUse Hook Silent Skip

1. Remove credentials file
2. Make a direct `git commit -m "test"` in terminal (not through /specledger.commit)
3. **Verify**: Hook exits silently, no stderr output

## Test 6: sl session sync Retry

1. After Test 4, restore network
2. Run `sl session sync`
3. **Verify**: Queued sessions uploaded. If retry fails, error logged locally and sent to Sentry.

## Check Error Logs

```bash
# Local log
cat ~/.specledger/capture-errors.log

# Sentry — check dashboard for project errors

# Run tests
make test
```
