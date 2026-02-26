# Session Capture Manual Test Scenarios

## Prerequisites

Before running tests:

```bash
# 1. Ensure CLI is built
make build

# 2. Login
sl auth login

# 3. Verify project has ID
cat specledger/specledger.yaml | grep "id:"

# 4. Verify Claude Code hook is configured
cat ~/.claude/settings.json | jq '.hooks.PostToolUse'
```

---

## Test Group 1: Session Capture (Core Flow)

### TC-1.1: Basic Commit Triggers Session Capture

**Preconditions:**
- Working in Claude Code
- Authenticated (`sl auth login`)
- Project has `project.id` in specledger.yaml

**Steps:**
1. Make a code change (e.g., add a comment to any file)
2. Stage the change: `git add <file>`
3. Commit: `git commit -m "test: session capture TC-1.1"`

**Expected:**
- Hook triggers `sl session capture`
- Output in stderr: `Session captured: <uuid> (<N> messages, <X> bytes)`
- No error blocking the commit

**Verify:**
```bash
sl session list --limit 1
# Should show commit with short hash matching TC-1.1 commit
```

---

### TC-1.2: Multiple Commits Capture Delta Only

**Preconditions:**
- Same as TC-1.1

**Steps:**
1. Make first change and commit: `git commit -m "test: delta capture part 1"`
2. Continue conversation with Claude
3. Make second change and commit: `git commit -m "test: delta capture part 2"`

**Expected:**
- First commit captures all messages from session start
- Second commit captures only NEW messages since first commit (delta)

**Verify:**
```bash
sl session list --limit 2
# Should show 2 sessions

sl session get <first-commit> --json | jq '.messages | length'
sl session get <second-commit> --json | jq '.messages | length'
# Second should have fewer messages (delta only)
```

---

### TC-1.3: Amend Commit Does NOT Trigger Capture

**Preconditions:**
- Same as TC-1.1
- At least one prior commit exists

**Steps:**
1. Run: `git commit --amend --no-edit`

**Expected:**
- Hook triggers but capture is SKIPPED
- No new session created

**Verify:**
```bash
sl session list --limit 1
# Should NOT show a new session for the amend
```

---

### TC-1.4: Non-Commit Git Commands Do NOT Trigger

**Steps:**
1. Run: `git status`
2. Run: `git add .`
3. Run: `git push`
4. Run: `git log --oneline -3`

**Expected:**
- Hook may trigger but capture is SKIPPED for each
- No "Session captured" output

---

### TC-1.5: Failed Commit Does NOT Trigger Capture

**Preconditions:**
- No staged changes

**Steps:**
1. Run: `git commit -m "test: should fail"` (with nothing staged)

**Expected:**
- Commit fails
- No session captured (hook checks `tool_success`)

---

## Test Group 2: Session List

### TC-2.1: List Sessions for Current Branch

**Preconditions:**
- At least one session captured on current branch

**Steps:**
```bash
sl session list
```

**Expected:**
- Table output with columns: COMMIT, MESSAGES, SIZE, STATUS, CAPTURED
- Sessions sorted by captured time (newest first)

---

### TC-2.2: List Sessions for Different Branch

**Steps:**
```bash
sl session list --feature main
```

**Expected:**
- Shows sessions for `main` branch only
- If no sessions on main: `No sessions found for branch 'main'`

---

### TC-2.3: Filter by Commit Hash (Partial)

**Preconditions:**
- Know a commit hash with captured session

**Steps:**
```bash
sl session list --commit abc123  # Use first 6-7 chars
```

**Expected:**
- Shows only sessions matching that commit hash

---

### TC-2.4: Filter by Task ID

**Preconditions:**
- Have a session with task_id set (optional field)

**Steps:**
```bash
sl session list --task SL-42
```

**Expected:**
- Shows only sessions with matching task_id
- Or empty if none match

---

### TC-2.5: JSON Output

**Steps:**
```bash
sl session list --json
```

**Expected:**
- Valid JSON array output
- Each session has: id, project_id, feature_branch, commit_hash, message_count, size_bytes, status, created_at

---

### TC-2.6: Limit Results

**Steps:**
```bash
sl session list --limit 3
```

**Expected:**
- Shows at most 3 sessions

---

### TC-2.7: Unauthenticated User

**Preconditions:**
- Logout: `sl auth logout`

**Steps:**
```bash
sl session list
```

**Expected:**
- Error: `authentication required: run 'sl auth login' first`

---

## Test Group 3: Session Get

### TC-3.1: Get Session by Commit Hash

**Preconditions:**
- Have a captured session with known commit hash

**Steps:**
```bash
sl session get abc1234  # Partial hash
```

**Expected:**
- Pretty-printed session output:
  - Session UUID
  - Branch name
  - Commit hash (full)
  - Author email
  - Capture timestamp
  - Message count
  - Separator line
  - Messages with [ROLE] timestamp and content

---

### TC-3.2: Get Session by Full UUID

**Preconditions:**
- Know a session ID from `sl session list --json`

**Steps:**
```bash
sl session get 550e8400-e29b-41d4-a716-446655440000
```

**Expected:**
- Same output as TC-3.1

---

### TC-3.3: Get Session by Task ID

**Preconditions:**
- Have a session with task_id set

**Steps:**
```bash
sl session get SL-42
```

**Expected:**
- Session content for that task

---

### TC-3.4: JSON Output

**Steps:**
```bash
sl session get abc1234 --json
```

**Expected:**
- Full session content as JSON
- Valid JSON structure with session_id, messages array, etc.

---

### TC-3.5: Raw Output (gzip)

**Steps:**
```bash
sl session get abc1234 --raw > /tmp/session.json.gz
gunzip -c /tmp/session.json.gz | jq .
```

**Expected:**
- Raw gzip bytes written to stdout
- Can decompress to valid JSON

---

### TC-3.6: Session Not Found

**Steps:**
```bash
sl session get nonexistent123
```

**Expected:**
- Error: `session not found: nonexistent123`

---

## Test Group 4: Session Sync (Offline Queue)

### TC-4.1: Check Queue Status (Empty)

**Preconditions:**
- No queued sessions

**Steps:**
```bash
sl session sync --status
```

**Expected:**
- Output: `No sessions in queue`

---

### TC-4.2: Sync Empty Queue

**Steps:**
```bash
sl session sync
```

**Expected:**
- Output: `No queued sessions to sync`

---

### TC-4.3: Queue Status with Pending Sessions

**Preconditions:**
- Simulate network failure during capture (manual: add file to `~/.specledger/session-queue/`)

**Steps:**
```bash
sl session sync --status
```

**Expected:**
- Shows count of queued sessions
- Lists each with session ID (truncated), commit hash, retry count

---

### TC-4.4: Sync Queued Sessions

**Preconditions:**
- Have queued sessions

**Steps:**
```bash
sl session sync
```

**Expected:**
- Output: `Uploaded N session(s)`
- Queue emptied after successful upload

**Verify:**
```bash
sl session sync --status
# Should show: No sessions in queue
```

---

### TC-4.5: Sync JSON Output

**Steps:**
```bash
sl session sync --json
```

**Expected:**
- JSON with: uploaded, failed, skipped, errors array

---

## Test Group 5: Test Mode

### TC-5.1: Run Test Mode in Claude Code

**Preconditions:**
- Running inside Claude Code session

**Steps:**
```bash
sl session capture --test-mode
```

**Expected:**
```
Running in test mode...
✓ Git repository detected
✓ Project ID found: <uuid>
✓ Authenticated as: <email>
✓ Claude Code sessions directory found
✓ Found transcript: /path/to/transcript.jsonl
✓ Session ID: <uuid>

📝 Simulating git commit hook...

⚠️  Test mode simulates the capture flow but won't create a real session.
✅ Session capture system is configured correctly!
```

---

### TC-5.2: Test Mode Outside Claude Code

**Preconditions:**
- Run from regular terminal (not Claude Code)

**Steps:**
```bash
sl session capture --test-mode
```

**Expected:**
- Error or warning about missing transcript path
- `CLAUDE_CODE_TRANSCRIPT_PATH` not set

---

## Test Group 6: Project Initialization

### TC-6.1: Init New Project (No specledger.yaml)

**Preconditions:**
- Fresh repo without specledger.yaml
- Or: `rm specledger/specledger.yaml` to simulate

**Steps:**
```bash
sl init
```

**Expected:**
- Prompts for project name (or uses repo name)
- Creates `specledger/specledger.yaml` with `project.id` set
- Registers project in Supabase (if authenticated)

**Verify:**
```bash
cat specledger/specledger.yaml | grep "id:"
# Should show a UUID
```

---

### TC-6.2: Init When Already Initialized

**Preconditions:**
- specledger.yaml already exists with project.id

**Steps:**
```bash
sl init
```

**Expected:**
- Warning: project already initialized
- Or: option to re-initialize/update

---

### TC-6.3: Session Commands Without Project ID

**Preconditions:**
- Remove project.id from specledger.yaml (keep file, clear id value)

**Steps:**
```bash
sl session list
```

**Expected:**
- Error: `project not configured`
- Hint: `Run 'sl init' to initialize the project`

---

### TC-6.4: Init Without Authentication

**Preconditions:**
- Logout: `sl auth logout`

**Steps:**
```bash
sl init
```

**Expected:**
- Either: Prompts to login first
- Or: Creates local config, warns that project won't be registered remotely

---

## Test Group 7: Error Handling

### TC-7.1: No Project ID (Error Message)

**Preconditions:**
- Remove/rename specledger.yaml or clear project.id

**Steps:**
```bash
sl session list
```

**Expected:**
- Error: `project not configured: no specledger.yaml found`
- Hint about running `sl init`

---

### TC-7.2: Invalid specledger.yaml

**Preconditions:**
- Corrupt specledger.yaml (invalid YAML)

**Steps:**
```bash
sl session list
```

**Expected:**
- Error parsing YAML

---

### TC-7.3: Network Failure During Capture

**Preconditions:**
- Disconnect network before commit

**Steps:**
1. Disable network
2. Make a commit

**Expected:**
- Session queued locally
- Output: `Session queued for upload: <uuid>`
- Commit still succeeds (capture doesn't block)

**Verify:**
```bash
sl session sync --status
# Should show 1 queued session
```

---

### TC-7.4: Expired Token

**Preconditions:**
- Have expired access token

**Steps:**
```bash
sl session list
```

**Expected:**
- Token refresh attempted
- If refresh succeeds: command works
- If refresh fails: `authentication required: run 'sl auth login' first`

---

## Test Group 8: Edge Cases

### TC-8.1: Very Large Session (Many Messages)

**Preconditions:**
- Long conversation with 100+ messages

**Steps:**
1. Make a commit

**Expected:**
- Capture succeeds
- Compression effective (check size in `sl session list`)

---

### TC-8.2: Unicode/Non-ASCII Content

**Preconditions:**
- Conversation contains emoji, Vietnamese, Chinese, etc.

**Steps:**
1. Make a commit

**Expected:**
- Capture succeeds
- `sl session get <commit>` displays correctly

---

### TC-8.3: Empty Transcript (No Messages)

**Preconditions:**
- Start fresh Claude Code session
- Immediately commit without any conversation

**Steps:**
1. `git commit -m "empty session test"`

**Expected:**
- Either: Session captured with 0 messages
- Or: Capture skipped (no new messages)

---

### TC-8.4: Concurrent Commits (Race Condition)

**Steps:**
1. Run two commits in quick succession (different terminals)

**Expected:**
- Both captures succeed
- No data corruption

---

### TC-8.5: Branch with Special Characters

**Steps:**
1. Create branch: `git checkout -b "feature/test-123_foo"`
2. Make a commit

**Expected:**
- Session captured with correct branch name
- `sl session list --feature "feature/test-123_foo"` works

---

## Cleanup

After testing:

```bash
# Revert test commits
git reset --soft HEAD~<N>  # N = number of test commits
git checkout .

# Or create a cleanup commit
git add . && git commit -m "test: cleanup session capture tests"
```

---

## Test Summary Checklist

| Group | Test | Status |
|-------|------|--------|
| 1. Capture | TC-1.1 Basic commit | ☐ |
| 1. Capture | TC-1.2 Delta capture | ☐ |
| 1. Capture | TC-1.3 Amend skip | ☐ |
| 1. Capture | TC-1.4 Non-commit skip | ☐ |
| 1. Capture | TC-1.5 Failed commit skip | ☐ |
| 2. List | TC-2.1 Current branch | ☐ |
| 2. List | TC-2.2 Different branch | ☐ |
| 2. List | TC-2.3 Filter by commit | ☐ |
| 2. List | TC-2.4 Filter by task | ☐ |
| 2. List | TC-2.5 JSON output | ☐ |
| 2. List | TC-2.6 Limit | ☐ |
| 2. List | TC-2.7 Unauth error | ☐ |
| 3. Get | TC-3.1 By commit hash | ☐ |
| 3. Get | TC-3.2 By UUID | ☐ |
| 3. Get | TC-3.3 By task ID | ☐ |
| 3. Get | TC-3.4 JSON output | ☐ |
| 3. Get | TC-3.5 Raw output | ☐ |
| 3. Get | TC-3.6 Not found | ☐ |
| 4. Sync | TC-4.1 Status empty | ☐ |
| 4. Sync | TC-4.2 Sync empty | ☐ |
| 4. Sync | TC-4.3 Status pending | ☐ |
| 4. Sync | TC-4.4 Sync queued | ☐ |
| 4. Sync | TC-4.5 JSON output | ☐ |
| 5. Test | TC-5.1 In Claude Code | ☐ |
| 5. Test | TC-5.2 Outside | ☐ |
| 6. Init | TC-6.1 Init new project | ☐ |
| 6. Init | TC-6.2 Init already done | ☐ |
| 6. Init | TC-6.3 Commands without ID | ☐ |
| 6. Init | TC-6.4 Init without auth | ☐ |
| 7. Error | TC-7.1 No project ID | ☐ |
| 7. Error | TC-7.2 Invalid YAML | ☐ |
| 7. Error | TC-7.3 Network failure | ☐ |
| 7. Error | TC-7.4 Expired token | ☐ |
| 8. Edge | TC-8.1 Large session | ☐ |
| 8. Edge | TC-8.2 Unicode | ☐ |
| 8. Edge | TC-8.3 Empty transcript | ☐ |
| 8. Edge | TC-8.4 Concurrent | ☐ |
| 8. Edge | TC-8.5 Special branch | ☐ |
