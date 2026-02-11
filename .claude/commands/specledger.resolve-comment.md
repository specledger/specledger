---
description: Resolve or delete a comment from Specledger.
---

## User Input

```text
$ARGUMENTS
```

You **MUST** consider the user input before proceeding (if not empty).

## Execution

When user calls `/specledger.resolve-comment`, follow these steps:

### Step 1: Check authentication

```bash
sl auth status
```

If not logged in → run `sl auth login` first.

### Step 2: Parse arguments

From `$ARGUMENTS`, extract:
- `--comment_id` or `-c`: ID of issue comment (integer)
- `--review_id` or `-r`: ID of review comment (UUID)
- `--skip` or `-s`: Skip processing, just mark as resolved

**Auto-detect**: If ID contains letters → UUID (review comment), otherwise → integer (issue comment)

If missing ID:
- Notify: "Please specify a comment ID"
- Show example usage
- Stop.

### Step 3: Fetch comment details

**Get credentials:**
```bash
SUPABASE_URL="https://iituikpbiesgofuraclk.supabase.co"
SUPABASE_ANON_KEY="sb_publishable_KpaZ2lKPu6eJ5WLqheu9_A_J9dYhGQb"
ACCESS_TOKEN=$(cat ~/.specledger/credentials.json | grep -o '"access_token": *"[^"]*"' | cut -d'"' -f4)
```

#### If Review Comment (UUID):
```bash
curl -s "${SUPABASE_URL}/rest/v1/review_comments?id=eq.${REVIEW_ID}&select=*" \
  -H "apikey: ${SUPABASE_ANON_KEY}" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}"
```

Get information:
- `file_path`: File being reviewed
- `selected_text`: Selected text snippet
- `content`: Comment/feedback content
- `is_resolved`: Current status

### Step 4: Analyze and address the review (IMPORTANT)

**This is the main step - DO NOT skip unless `--skip` flag is provided**

1. **Read the reviewed file:**
   ```
   Read file_path from comment
   ```

2. **Understand review feedback:**
   - Analyze `content` (comment content)
   - Look at `selected_text` for context
   - Determine what reviewer wants: clarify? fix? add? remove?

3. **Propose changes:**
   Display to user:
   ```
   📝 Review Comment Analysis
   ━━━━━━━━━━━━━━━━━━━━━━━━━━
   📁 File: [file_path]
   📌 Selected: "[selected_text]"
   💬 Feedback: "[content]"

   🔍 Analysis:
   [Explain what the reviewer wants]

   ✏️ Proposed changes:
   [Describe what needs to be edited]
   ```

4. **Perform edit (if needed):**
   - Use Edit tool to modify the file
   - Or ask user if unsure how to handle

5. **Confirm with user:**
   ```
   Do you want to mark this comment as resolved? (Y/n)
   ```

### Step 5: Mark as resolved

#### 5a. If Issue Comment (integer ID) → DELETE

```bash
curl -s -X DELETE "${SUPABASE_URL}/rest/v1/comments?id=eq.${COMMENT_ID}" \
  -H "apikey: ${SUPABASE_ANON_KEY}" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}"
```

#### 5b. If Review Comment (UUID) → UPDATE is_resolved = true

```bash
curl -s -X PATCH "${SUPABASE_URL}/rest/v1/review_comments?id=eq.${REVIEW_ID}" \
  -H "apikey: ${SUPABASE_ANON_KEY}" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}" \
  -H "Content-Type: application/json" \
  -H "Prefer: return=minimal" \
  -d '{"is_resolved": true}'
```

### Step 6: Handle response

**If success (HTTP 200/204):**

```text
✅ Review comment #[id] has been processed and marked as resolved.

Changes made:
- [List edits performed]
```

**If error:**
- 401: "Session expired. Run `sl auth login` again."
- 403: "You don't have permission to resolve this comment."
- 404: "Comment not found."

### Step 7: Show next actions

```text
Next steps:
- /specledger.fetch-comments to view remaining comments
```

## Example Usage

```text
# Resolve and process review feedback (default behavior)
/specledger.resolve-comment #54181d3b
/specledger.resolve-comment f030526a-1234-5678-9abc-def012345678

# Just mark as resolved, don't process content
/specledger.resolve-comment #54181d3b --skip
/specledger.resolve-comment -r f030526a -s

# Issue comments (will DELETE)
/specledger.resolve-comment -c 35
```

## Workflow Summary

```
┌─────────────────────────────────────────────────────────────┐
│  1. Fetch comment details                                    │
│  2. Read the reviewed file                                   │
│  3. Analyze: What does the reviewer want?                    │
│  4. Propose changes to address the feedback                  │
│  5. Edit file (with user confirmation)                       │
│  6. Mark comment as resolved                                 │
│  7. Show summary of changes made                             │
└─────────────────────────────────────────────────────────────┘
```

## Table Info

| Table | ID Type | Resolve Action |
|-------|---------|----------------|
| `comments` | integer | DELETE (no is_resolved column) |
| `review_comments` | UUID | UPDATE is_resolved = true |
