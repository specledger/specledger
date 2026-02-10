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
- `--comment_id` hoặc `-c`: ID của issue comment (integer)
- `--review_id` hoặc `-r`: ID của review comment (UUID)
- `--skip` hoặc `-s`: Bỏ qua việc xử lý, chỉ mark as resolved

**Auto-detect**: Nếu ID chứa chữ cái → là UUID (review comment), ngược lại → integer (issue comment)

Nếu thiếu ID:
- Thông báo: "Vui lòng chỉ định comment ID"
- Hiển thị example usage
- Dừng lại.

### Step 3: Fetch comment details

**Lấy credentials:**
```bash
SUPABASE_URL="https://iituikpbiesgofuraclk.supabase.co"
SUPABASE_ANON_KEY="sb_publishable_KpaZ2lKPu6eJ5WLqheu9_A_J9dYhGQb"
ACCESS_TOKEN=$(cat ~/.specledger/credentials.json | grep -o '"access_token": *"[^"]*"' | cut -d'"' -f4)
```

#### Nếu là Review Comment (UUID):
```bash
curl -s "${SUPABASE_URL}/rest/v1/review_comments?id=eq.${REVIEW_ID}&select=*" \
  -H "apikey: ${SUPABASE_ANON_KEY}" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}"
```

Lấy thông tin:
- `file_path`: File được review
- `selected_text`: Đoạn text được chọn
- `content`: Nội dung comment/feedback
- `is_resolved`: Trạng thái hiện tại

### Step 4: Analyze and address the review (QUAN TRỌNG)

**Đây là bước chính - KHÔNG được bỏ qua trừ khi có flag `--skip`**

1. **Đọc file được review:**
   ```
   Read file_path từ comment
   ```

2. **Hiểu review feedback:**
   - Phân tích `content` (nội dung comment)
   - Xem `selected_text` để hiểu context
   - Xác định reviewer muốn gì: clarify? fix? add? remove?

3. **Đề xuất thay đổi:**
   Hiển thị cho user:
   ```
   📝 Review Comment Analysis
   ━━━━━━━━━━━━━━━━━━━━━━━━━━
   📁 File: [file_path]
   📌 Selected: "[selected_text]"
   💬 Feedback: "[content]"

   🔍 Phân tích:
   [Giải thích reviewer muốn gì]

   ✏️ Đề xuất thay đổi:
   [Mô tả những gì cần edit]
   ```

4. **Thực hiện edit (nếu cần):**
   - Sử dụng Edit tool để sửa file
   - Hoặc hỏi user nếu không chắc chắn cách xử lý

5. **Confirm với user:**
   ```
   Bạn có muốn mark comment này là resolved? (Y/n)
   ```

### Step 5: Mark as resolved

#### 5a. Nếu là Issue Comment (integer ID) → DELETE

```bash
curl -s -X DELETE "${SUPABASE_URL}/rest/v1/comments?id=eq.${COMMENT_ID}" \
  -H "apikey: ${SUPABASE_ANON_KEY}" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}"
```

#### 5b. Nếu là Review Comment (UUID) → UPDATE is_resolved = true

```bash
curl -s -X PATCH "${SUPABASE_URL}/rest/v1/review_comments?id=eq.${REVIEW_ID}" \
  -H "apikey: ${SUPABASE_ANON_KEY}" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}" \
  -H "Content-Type: application/json" \
  -H "Prefer: return=minimal" \
  -d '{"is_resolved": true}'
```

### Step 6: Handle response

**Nếu success (HTTP 200/204):**

```text
✅ Review comment #[id] đã được xử lý và đánh dấu resolved.

Thay đổi đã thực hiện:
- [Liệt kê các edit đã làm]
```

**Nếu error:**
- 401: "Phiên đăng nhập hết hạn. Chạy `sl auth login` lại."
- 403: "Bạn không có quyền resolve comment này."
- 404: "Comment không tồn tại."

### Step 7: Show next actions

```text
Tiếp theo:
- /specledger.fetch-comments để xem danh sách comments còn lại
```

## Example Usage

```text
# Resolve và xử lý review feedback (default behavior)
/specledger.resolve-comment #54181d3b
/specledger.resolve-comment f030526a-1234-5678-9abc-def012345678

# Chỉ mark as resolved, không xử lý content
/specledger.resolve-comment #54181d3b --skip
/specledger.resolve-comment -r f030526a -s

# Issue comments (sẽ DELETE)
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
| `comments` | integer | DELETE (không có is_resolved column) |
| `review_comments` | UUID | UPDATE is_resolved = true |
