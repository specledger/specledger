---
description: Post a new comment on a specification file.
---

## User Input

```text
$ARGUMENTS
```

You **MUST** consider the user input before proceeding (if not empty).

## Execution

When user calls `/specledger.post-comment`, follow these steps:

### Step 1: Check authentication

```bash
sl auth status
```

If not logged in → run `sl auth login` first.

### Step 2: Parse arguments

From `$ARGUMENTS`, extract:
- `--file` or `-f`: File path (relative to repo root)
- `--line` or `-l`: Line number (optional)
- `--message` or `-m`: Comment content (required)
- `--selected` or `-s`: Selected text (optional, for context)

If missing `--file` or `--message`:
- Notify: "Please specify file and message"
- Show example usage
- Stop.

### Step 3: Get spec-key and change_id

```bash
SPEC_KEY=$(git branch --show-current)
```

**Get change_id for this spec:**
```bash
SUPABASE_URL="https://iituikpbiesgofuraclk.supabase.co"
SUPABASE_ANON_KEY="sb_publishable_KpaZ2lKPu6eJ5WLqheu9_A_J9dYhGQb"
ACCESS_TOKEN=$(python3 -c "import json; print(json.load(open('$HOME/.specledger/credentials.json'))['access_token'])")

# Get latest change_id for this spec
curl -s "${SUPABASE_URL}/rest/v1/changes?spec_key=eq.${SPEC_KEY}&select=id&order=created_at.desc&limit=1" \
  -H "apikey: ${SUPABASE_ANON_KEY}" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}"
```

If no change_id:
- Notify: "No change found for this spec. Please create a review on the web first."
- Stop.

### Step 4: Get user info from credentials

```bash
USER_EMAIL=$(python3 -c "import json; print(json.load(open('$HOME/.specledger/credentials.json'))['user_email'])")
USER_ID=$(python3 -c "import json; print(json.load(open('$HOME/.specledger/credentials.json'))['user_id'])")
```

Extract user name from JWT (or use email prefix):
```bash
USER_NAME=$(echo $USER_EMAIL | cut -d'@' -f1)
```

### Step 5: Post comment to Supabase

```bash
curl -s -X POST "${SUPABASE_URL}/rest/v1/review_comments" \
  -H "apikey: ${SUPABASE_ANON_KEY}" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}" \
  -H "Content-Type: application/json" \
  -H "Prefer: return=representation" \
  -d '{
    "change_id": "${CHANGE_ID}",
    "file_path": "${FILE_PATH}",
    "line": ${LINE_NUMBER},
    "selected_text": "${SELECTED_TEXT}",
    "content": "${MESSAGE}",
    "is_resolved": false,
    "author_id": "${USER_ID}",
    "author_name": "${USER_NAME}",
    "author_email": "${USER_EMAIL}"
  }'
```

### Step 6: Handle response

**If success (HTTP 201):**

```text
✅ Comment posted successfully!

📁 File: [file_path]
📍 Line: [line_number]
💬 Comment: "[message]"

Comment ID: [returned_id]
```

**If error:**
- 401: "Session expired. Run `sl auth login` again."
- 403: "You don't have permission to post comments."
- 400: "Invalid request. Check file path and message."

### Step 7: Show next actions

```text
Next steps:
- /specledger.fetch-comments to view comment list
```

## Example Usage

```text
# Post comment on a file
/specledger.post-comment --file specledger/009-xxx/spec.md --message "Need more details here"

# Post comment on specific line
/specledger.post-comment -f specledger/009-xxx/plan.md -l 42 -m "Consider alternative approach"

# Post comment with selected text context
/specledger.post-comment --file spec.md --line 10 --selected "authentication flow" --message "Should use OAuth2"
```

## Table Schema

### `review_comments`
| Column | Type | Required |
|--------|------|----------|
| id | UUID | auto |
| change_id | UUID | yes |
| file_path | string | yes |
| line | integer | no |
| selected_text | string | no |
| content | string | yes |
| is_resolved | boolean | default false |
| author_id | UUID | yes |
| author_name | string | yes |
| author_email | string | yes |
| created_at | timestamp | auto |
