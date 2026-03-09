# Contract: /specledger.commit Slash Command

## Command File

**Source**: `pkg/embedded/skills/commands/specledger.commit.md`
**Distributed to**: `.claude/commands/specledger.commit.md`

## Interface

```
/specledger.commit [optional commit message]
```

**Input**: Optional commit message as `$ARGUMENTS`. If empty, agent analyzes changes and generates a message.

## Workflow Steps

```
1. Check staged changes (git status --porcelain)
   └─ If nothing staged → prompt user to stage, exit

2. Analyze changes (git diff --cached)
   └─ Generate commit message if not provided

3. Check auth status
   └─ Read ~/.specledger/credentials.json
   └─ Check project ID from specledger.yaml
   └─ Set flags: has_auth, has_project_id

4. Commit (git commit -m "...")
   └─ PostToolUse hook runs sl session capture automatically
   └─ If no auth: hook silently skips
   └─ If auth + project ID: hook captures session
   └─ If capture fails: session queued + error logged

5. Push (git push origin <branch>)
   └─ Always attempt push regardless of capture status
   └─ If push fails: show error to user

6. Show summary
   └─ Commit hash
   └─ Branch name
   └─ Session capture status (captured/skipped/queued)
   └─ Any errors logged
```

## Auth Decision Matrix

| Has Credentials | Has Project ID | Session Capture | Error Logging |
|----------------|----------------|-----------------|---------------|
| No | - | Skip silently | None |
| Yes | No | Skip silently | None |
| Yes | Yes | Attempt | On failure: local file + Sentry |
