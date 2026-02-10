---
description: View comments on current spec from Specledger.
---

## User Input

```text
$ARGUMENTS
```

## Execution

### Step 1: Get spec-key

**Nếu có argument `--spec`**: dùng giá trị đó

**Nếu không có argument**: lấy từ git branch
```bash
SPEC_KEY=$(git branch --show-current)
```

Spec-key = tên branch = tên folder trong `specledger/`

### Step 2: Check authentication

```bash
sl auth status
```

Nếu chưa login → chạy `sl auth login` trước.

### Step 3: Fetch comments từ Supabase

Fetch từ **2 tables**:

**Lấy credentials (KHÔNG đọc file trực tiếp):**
```bash
SUPABASE_URL=$(sl auth supabase --url)
SUPABASE_ANON_KEY=$(sl auth supabase --key)
ACCESS_TOKEN=$(sl auth token)
```

#### 3a. Issue comments (table: `comments`)
```bash
curl -s "${SUPABASE_URL}/rest/v1/comments?select=*&issue_id=eq.${SPEC_KEY}" \
  -H "apikey: ${SUPABASE_ANON_KEY}" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}"
```

#### 3b. Review comments (table: `review_comments`)
```bash
curl -s "${SUPABASE_URL}/rest/v1/review_comments?select=*&file_path=like.*${SPEC_KEY}*" \
  -H "apikey: ${SUPABASE_ANON_KEY}" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}"
```

**Lưu ý**: Sử dụng `sl auth token` thay vì đọc file `~/.specledger/credentials.json` để bảo mật token.

### Step 4: Render comment list

```text
📄 Spec: 009-add-login-and-comment-commands

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📋 ISSUE COMMENTS (issue_id: 009-add-login-and-comment-commands)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

#35 | Bao Lam | 2026-01-30
    "Simplified adopt.md..."

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📝 REVIEW COMMENTS (file-level)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

#f030526a | Son Vo | 2026-02-09 | ⏳ unresolved
    📁 File: specs/008-xxx/spec.md
    📌 Selected: "Refactor và cải thiện..."
    💬 Comment: "Sai dự án"

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📊 Total: 1 issue comment, 2 review comments
```

## Example Usage

```text
/specledger.fetch-comments                    # Dùng branch hiện tại
/specledger.fetch-comments --spec other-spec  # Chỉ định spec khác (optional)
```

## Table Schemas

### `comments` (Issue comments)
| Column | Type |
|--------|------|
| id | integer |
| issue_id | string |
| author | string |
| text | string |
| created_at | timestamp |

### `review_comments` (File review comments)
| Column | Type |
|--------|------|
| id | UUID |
| change_id | UUID |
| file_path | string |
| selected_text | string |
| content | string |
| is_resolved | boolean |
| author_name | string |
| author_email | string |
| created_at | timestamp |
