---
description: Fetch and address review comments from Supabase using CURL
---

## User Input

```text
$ARGUMENTS
```

You **MUST** consider the user input before proceeding (if not empty).

## Purpose

This command fetches, displays, processes, and resolves review comments directly from Supabase using CURL. It combines fetch, resolve, and post functionality into a single workflow.

**When to use**:
- After pushing changes to GitHub
- When starting a new session and need to check for comments
- When team members have added comments to your specs
- To address feedback on your specifications

**Prerequisites**:
- Must be logged in via `sl auth login`

## Input Options

| Usage | Description |
|-------|-------------|
| `/specledger.revise` | Auto-detect spec from current branch |
| `/specledger.revise 009-feature-name` | Specify spec-key explicitly |
| `/specledger.revise --resolve <id>` | Resolve specific comment only |
| `/specledger.revise --post -f <file> -m <msg>` | Post new comment |

## Execution Flow

### Step 1: Check Authentication

```bash
sl auth status
```

If not logged in → run `sl auth login` first.

### Step 2: Get Credentials (IMPORTANT)

**Option A: Use CLI commands (preferred, if available):**
```bash
SUPABASE_URL=$(sl auth supabase --url)
SUPABASE_ANON_KEY=$(sl auth supabase --key)
ACCESS_TOKEN=$(sl auth token)
```

**Option B: Fallback - extract from credentials file:**
```bash
SUPABASE_URL="https://iituikpbiesgofuraclk.supabase.co"
SUPABASE_ANON_KEY="sb_publishable_KpaZ2lKPu6eJ5WLqheu9_A_J9dYhGQb"
ACCESS_TOKEN=$(python3 -c "import json; print(json.load(open('$HOME/.specledger/credentials.json'))['access_token'])")
```

**Note:** If CLI commands fail (TUI mode), use the fallback method.

### Step 3: Determine Spec-Key

**If `$ARGUMENTS` contains spec-key**: use that value
**If no argument**: get from git branch

```bash
SPEC_KEY=$(git branch --show-current)
```

Spec-key = branch name = folder name in `specledger/`

---

## Mode A: Fetch & Process Comments (Default)

### A1. Fetch Comments from Supabase

Query **2 tables**:

#### Issue Comments (table: `comments`)
```bash
curl -s "${SUPABASE_URL}/rest/v1/comments?select=*&issue_id=eq.${SPEC_KEY}" \
  -H "apikey: ${SUPABASE_ANON_KEY}" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}"
```

#### Review Comments (table: `review_comments`)
```bash
curl -s "${SUPABASE_URL}/rest/v1/review_comments?select=*&file_path=like.*${SPEC_KEY}*&is_resolved=eq.false" \
  -H "apikey: ${SUPABASE_ANON_KEY}" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}"
```

### A2. Display Comments Summary (MANDATORY)

**CRITICAL: You MUST display all fetched comments to the user before processing.**

After fetching, immediately output this summary:

```text
📄 Spec: {SPEC_KEY}

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📋 ISSUE COMMENTS (issue_id: {SPEC_KEY})
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

#{id} | {author} | {created_at}
    "{text}"

(If empty: display "(No issue comments)")

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📝 REVIEW COMMENTS (file-level)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

#{id (8 chars)} | {author_name} | {created_at} | ⏳ unresolved
    📁 File: {file_path}
    📍 Line: {line} (if available)
    📌 Selected: "{selected_text}" (if available)
    💬 Comment: "{content}"

(If empty: display "(No unresolved review comments)")

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📊 Total: {n} issue comments, {m} review comments
```

**Rules for display:**
- ALWAYS show the summary, even if there are no comments
- Display ALL comments from both tables before asking any questions
- Use the exact format above for consistency
- If no comments found, still show the summary with "(No comments)" messages

### A3. Process Each Comment Interactively

For each **unresolved** comment:

1. **Read the file** at `file_path`
2. **Find the `selected_text`** in the file (if available)
3. **Analyze the comment** content and context
4. **Display analysis:**
   ```text
   📝 Review Comment Analysis
   ━━━━━━━━━━━━━━━━━━━━━━━━━━
   📁 File: {file_path}
   📌 Selected: "{selected_text}"
   💬 Feedback: "{content}"

   🔍 Analysis:
   {Explain what the reviewer wants}

   ✏️ Proposed changes:
   {Describe what needs to be edited}
   ```
5. **Use AskUserQuestion** to get user preference (2-3 options)
6. **Apply the edit** to the file
7. **Mark as resolved** (see A4)

**CRITICAL RULES:**
- MUST use AskUserQuestion before making ANY edit
- If `selected_text` is provided, locate it in the file for context
- Some comments are acknowledgments (like "good") and don't require file changes
- Apply edits incrementally, one comment at a time

### A4. Mark Comments as Resolved

#### Issue Comment (integer ID) → DELETE
```bash
curl -s -X DELETE "${SUPABASE_URL}/rest/v1/comments?id=eq.${COMMENT_ID}" \
  -H "apikey: ${SUPABASE_ANON_KEY}" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}"
```

#### Review Comment (UUID) → UPDATE is_resolved = true
```bash
curl -s -X PATCH "${SUPABASE_URL}/rest/v1/review_comments?id=eq.${REVIEW_ID}" \
  -H "apikey: ${SUPABASE_ANON_KEY}" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}" \
  -H "Content-Type: application/json" \
  -H "Prefer: return=minimal" \
  -d '{"is_resolved": true}'
```

### A5. Commit Changes (Optional)

After all comments are addressed:

```text
📝 Ready to commit changes?

Options:
A) Yes, commit and push changes
B) No, I'll commit manually later
```

If user chooses A:
```bash
git add <modified-files>
git commit -m "feat: address review comments

Updated files: <list>
Comments resolved: <count>"
git push origin HEAD
```

### A6. Summary Report

```text
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
✅ Review Session Complete
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📄 Spec: {SPEC_KEY}
💬 Comments Addressed: {count}
📄 Files Updated: {count}

Files:
  ✓ {file_path} ({comment_count} comments)
  ...

Next Steps:
- View changes: git diff HEAD~1
- Continue with /specledger.implement
- Check new comments: /specledger.revise
```

---

## Mode B: Resolve Specific Comment

When `$ARGUMENTS` contains `--resolve <id>`:

### B1. Parse ID Type

- If ID contains letters → UUID (review comment)
- If ID is numeric → integer (issue comment)

### B2. Fetch Comment Details

#### Review Comment:
```bash
curl -s "${SUPABASE_URL}/rest/v1/review_comments?id=eq.${REVIEW_ID}&select=*" \
  -H "apikey: ${SUPABASE_ANON_KEY}" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}"
```

### B3. Display Comment Details (MANDATORY)

**MUST display the fetched comment before processing:**

```text
📝 Comment Details
━━━━━━━━━━━━━━━━━━━━━━━━━━
🆔 ID: {id}
📁 File: {file_path}
📍 Line: {line}
📌 Selected: "{selected_text}"
💬 Content: "{content}"
👤 Author: {author_name}
📅 Created: {created_at}
```

### B4. Process & Resolve

Same as A3-A4 but for single comment.

---

## Mode C: Post New Comment

When `$ARGUMENTS` contains `--post`:

### C1. Parse Arguments

- `--file` or `-f`: File path (required)
- `--message` or `-m`: Comment content (required)
- `--line` or `-l`: Line number (optional)
- `--selected` or `-s`: Selected text (optional)

### C2. Get Change ID

**Option A: Get from existing review_comments (preferred)**
```bash
# Get change_id from an existing comment on this spec
curl -s "${SUPABASE_URL}/rest/v1/review_comments?select=change_id&file_path=like.*${SPEC_KEY}*&limit=1" \
  -H "apikey: ${SUPABASE_ANON_KEY}" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}"
```

**Option B: Query changes table by base_branch**
```bash
curl -s "${SUPABASE_URL}/rest/v1/changes?select=id&base_branch=eq.${SPEC_KEY}&order=created_at.desc&limit=1" \
  -H "apikey: ${SUPABASE_ANON_KEY}" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}"
```

If no change_id → "No change found for this spec. Please create a review on the web first."

### C3. Get User Info

```bash
USER_ID=$(sl auth user --id)
USER_EMAIL=$(sl auth user --email)
USER_NAME=$(echo $USER_EMAIL | cut -d'@' -f1)
```

### C4. Post Comment

```bash
curl -s -X POST "${SUPABASE_URL}/rest/v1/review_comments" \
  -H "apikey: ${SUPABASE_ANON_KEY}" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}" \
  -H "Content-Type: application/json" \
  -H "Prefer: return=representation" \
  -d '{
    "change_id": "'${CHANGE_ID}'",
    "file_path": "'${FILE_PATH}'",
    "line": '${LINE_NUMBER:-null}',
    "selected_text": "'${SELECTED_TEXT}'",
    "content": "'${MESSAGE}'",
    "is_resolved": false,
    "author_id": "'${USER_ID}'",
    "author_name": "'${USER_NAME}'",
    "author_email": "'${USER_EMAIL}'"
  }'
```

### C5. Display Posted Comment (MANDATORY)

**MUST display the posted comment to confirm success:**

```text
✅ Comment posted successfully!

━━━━━━━━━━━━━━━━━━━━━━━━━━
🆔 Comment ID: {returned_id}
📁 File: {file_path}
📍 Line: {line_number}
📌 Selected: "{selected_text}" (if provided)
💬 Content: "{message}"
👤 Author: {author_name}
📅 Created: {created_at}
━━━━━━━━━━━━━━━━━━━━━━━━━━

Next: /specledger.revise to see all comments
```

---

## Table Schemas

### `comments` (Issue comments)
| Column | Type | Resolve Action |
|--------|------|----------------|
| id | integer | DELETE |
| issue_id | string | |
| author | string | |
| text | string | |
| created_at | timestamp | |

### `review_comments` (File review comments)
| Column | Type | Resolve Action |
|--------|------|----------------|
| id | UUID | UPDATE is_resolved=true |
| change_id | UUID | |
| file_path | string | |
| selected_text | string | |
| content | string | |
| line | integer | |
| is_resolved | boolean | |
| author_id | UUID | |
| author_name | string | |
| author_email | string | |
| created_at | timestamp | |

---

## Error Handling

| Error | Cause | Solution |
|-------|-------|----------|
| JWT expired | Token expired | Run `sl login` to refresh token |
| 401 Unauthorized | Session expired | Run `sl login` again |
| 403 Forbidden | No permission | Check access rights |
| 404 Not Found | Comment/spec not found | Verify ID/spec-key |
| PGRST303 | JWT expired | Run `sl login` to refresh token |
| No comments | All resolved | Nothing to do |
| Credentials file not found | Not logged in | Run `sl login` first |

---

## Example Usage

```bash
# Default: fetch and process all unresolved comments
/specledger.revise

# Specify spec explicitly
/specledger.revise 009-feature-name

# Resolve specific comment
/specledger.revise --resolve f030526a

# Post new comment
/specledger.revise --post -f specledger/009-xxx/spec.md -m "Need more details"
/specledger.revise --post -f spec.md -l 42 -m "Consider alternative"
```
