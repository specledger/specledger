---
description: Post a new comment on a specification file.
---

## User Input

```text
$ARGUMENTS
```

You **MUST** consider the user input before proceeding (if not empty).

## Execution

Khi user gọi `/specledger.post-comment`, thực hiện các bước sau:

### Step 1: Check authentication

```bash
sl auth status
```

Nếu chưa login → chạy `sl auth login` trước.

### Step 2: Parse arguments

Từ `$ARGUMENTS`, extract:
- `--file` hoặc `-f`: File path (relative to repo root)
- `--line` hoặc `-l`: Line number (optional)
- `--message` hoặc `-m`: Comment content (required)
- `--selected` hoặc `-s`: Selected text (optional, for context)

Nếu thiếu `--file` hoặc `--message`:
- Thông báo: "Vui lòng chỉ định file và message"
- Hiển thị example usage
- Dừng lại.

### Step 3: Get spec-key and change_id

```bash
SPEC_KEY=$(git branch --show-current)
```

**Lấy change_id cho spec này:**
```bash
SUPABASE_URL="https://iituikpbiesgofuraclk.supabase.co"
SUPABASE_ANON_KEY="sb_publishable_KpaZ2lKPu6eJ5WLqheu9_A_J9dYhGQb"
ACCESS_TOKEN=$(python3 -c "import json; print(json.load(open('$HOME/.specledger/credentials.json'))['access_token'])")

# Get latest change_id for this spec
curl -s "${SUPABASE_URL}/rest/v1/changes?spec_key=eq.${SPEC_KEY}&select=id&order=created_at.desc&limit=1" \
  -H "apikey: ${SUPABASE_ANON_KEY}" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}"
```

Nếu không có change_id:
- Thông báo: "Không tìm thấy change cho spec này. Vui lòng tạo review trên web trước."
- Dừng lại.

### Step 4: Get user info from credentials

```bash
USER_EMAIL=$(python3 -c "import json; print(json.load(open('$HOME/.specledger/credentials.json'))['user_email'])")
USER_ID=$(python3 -c "import json; print(json.load(open('$HOME/.specledger/credentials.json'))['user_id'])")
```

Extract user name from JWT (hoặc dùng email prefix):
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

**Nếu success (HTTP 201):**

```text
✅ Comment posted successfully!

📁 File: [file_path]
📍 Line: [line_number]
💬 Comment: "[message]"

Comment ID: [returned_id]
```

**Nếu error:**
- 401: "Phiên đăng nhập hết hạn. Chạy `sl auth login` lại."
- 403: "Bạn không có quyền post comment."
- 400: "Invalid request. Check file path and message."

### Step 7: Show next actions

```text
Tiếp theo:
- /specledger.fetch-comments để xem danh sách comments
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
