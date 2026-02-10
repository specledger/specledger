---
description: Resolve or delete a comment from Specledger.
---

## User Input

```text
$ARGUMENTS
```

You **MUST** consider the user input before proceeding (if not empty).

## Execution

Khi user gọi `/specledger.resolve-comment`, thực hiện các bước sau:

### Step 1: Check authentication

```bash
sl auth status
```

Nếu chưa login → chạy `sl auth login` trước.

### Step 2: Parse arguments

Từ `$ARGUMENTS`, extract:
- `--comment_id` hoặc `-c`: ID của issue comment (integer) → sẽ DELETE
- `--review_id` hoặc `-r`: ID của review comment (UUID) → sẽ UPDATE is_resolved = true

**Auto-detect**: Nếu ID chứa `-` (dấu gạch ngang) → là UUID (review comment), ngược lại → integer (issue comment)

Nếu thiếu ID:
- Thông báo: "Vui lòng chỉ định comment ID"
- Hiển thị example usage
- Dừng lại.

### Step 3: Resolve comment via Supabase API

**Lấy credentials (KHÔNG đọc file trực tiếp):**
```bash
SUPABASE_URL=$(sl auth supabase --url)
SUPABASE_ANON_KEY=$(sl auth supabase --key)
ACCESS_TOKEN=$(sl auth token)
```

**Lưu ý**: Sử dụng `sl auth token` thay vì đọc file `~/.specledger/credentials.json` để bảo mật token.

#### 3a. Nếu là Issue Comment (integer ID) → DELETE

```bash
COMMENT_ID=35

curl -s -X DELETE "${SUPABASE_URL}/rest/v1/comments?id=eq.${COMMENT_ID}" \
  -H "apikey: ${SUPABASE_ANON_KEY}" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}"
```

#### 3b. Nếu là Review Comment (UUID) → UPDATE is_resolved = true

```bash
REVIEW_ID="f030526a-xxxx-xxxx-xxxx-xxxxxxxxxxxx"

curl -s -X PATCH "${SUPABASE_URL}/rest/v1/review_comments?id=eq.${REVIEW_ID}" \
  -H "apikey: ${SUPABASE_ANON_KEY}" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"is_resolved": true}'
```

### Step 4: Handle response

**Nếu success (HTTP 200/204):**

Với Issue Comment:
```text
✅ Issue comment #35 đã được xóa.
```

Với Review Comment:
```text
✅ Review comment #f030526a đã được đánh dấu resolved.
```

**Nếu error:**
- 401: "Phiên đăng nhập hết hạn. Chạy `sl auth login` lại."
- 403: "Bạn không có quyền resolve comment này."
- 404: "Comment không tồn tại."

### Step 5: Show next actions

```text
Tiếp theo:
- /specledger.comment để xem danh sách comments còn lại
```

## Example Usage

```text
# Issue comments (integer ID) - sẽ DELETE
/specledger.resolve-comment --comment_id 35
/specledger.resolve-comment -c 36

# Review comments (UUID) - sẽ UPDATE is_resolved = true
/specledger.resolve-comment --review_id f030526a-1234-5678-9abc-def012345678
/specledger.resolve-comment -r f030526a

# Auto-detect based on ID format
/specledger.resolve-comment 35           # integer → issue comment
/specledger.resolve-comment f030526a     # contains letters → review comment
```

## Table Info

| Table | ID Type | Resolve Action |
|-------|---------|----------------|
| `comments` | integer | DELETE (không có is_resolved column) |
| `review_comments` | UUID | UPDATE is_resolved = true |
