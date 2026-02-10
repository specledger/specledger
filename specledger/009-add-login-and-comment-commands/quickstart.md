# Quickstart: CLI Authentication and Comment Commands

**Feature Branch**: `009-add-login-and-comment-commands`

## Overview

This feature adds four Claude Code slash commands for authentication and comment management:

| Command | Purpose |
|---------|---------|
| `/specledger.login` | Authenticate via browser token paste |
| `/specledger.logout` | Remove local session |
| `/specledger.comment` | View comments on a spec file |
| `/specledger.resolve-comment` | Mark a comment as resolved |

## Prerequisites

1. **Claude Code CLI** installed and configured
2. **Supabase backend** with `comments` table and RLS policies (provided by Specledger)
3. **Auth UI** deployed at `https://specledger.io/cli/auth`

---

## Quick Test

### 1. Login

```bash
# In Claude Code, run:
/specledger.login
```

This will:
1. Open `https://specledger.io/cli/auth` in your browser
2. Prompt you to paste the JSON token
3. Save session to `~/.specledger/session.json`

**Manual Test (without Auth UI)**:
```bash
# Create a test session file
mkdir -p ~/.specledger
cat > ~/.specledger/session.json << 'EOF'
{
  "access_token": "test-access-token",
  "refresh_token": "test-refresh-token",
  "expires_at": 9999999999,
  "user_id": "test-user-id"
}
EOF
chmod 600 ~/.specledger/session.json
```

### 2. View Comments

```bash
/specledger.comment --file auth.md
```

Expected output:
```text
📄 auth.md

[1] ❗ Nam (open)
    "Thiếu error case cho expired token"

[2] 💡 Tram (open)
    "Phần refresh token nên mô tả rõ hơn"

Actions:
[r N] Resolve comment #N
[a]   Resolve all addressed
[q]   Quit
```

### 3. Resolve Comment

```bash
/specledger.resolve-comment --comment_id:1
```

### 4. Logout

```bash
/specledger.logout
```

---

## Command Locations

All commands are defined in `.claude/commands/`:

```text
.claude/commands/
├── specledger.login.md
├── specledger.logout.md
├── specledger.comment.md
└── specledger.resolve-comment.md
```

---

## Session File Format

**Location**: `~/.specledger/session.json`
**Permissions**: `600`

```json
{
  "access_token": "JWT_ACCESS_TOKEN",
  "refresh_token": "JWT_REFRESH_TOKEN",
  "expires_at": 1712345678,
  "user_id": "uuid"
}
```

**Validation**:
- All four fields required
- `expires_at` is Unix timestamp (seconds)
- File permission must be 600

---

## API Endpoints (Supabase)

### Fetch Comments
```bash
curl -s 'https://specledger.supabase.co/rest/v1/comments?select=*&file_path=eq.auth.md' \
  -H "apikey: YOUR_ANON_KEY" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

### Resolve Comment
```bash
curl -s -X PATCH 'https://specledger.supabase.co/rest/v1/comments?id=eq.COMMENT_UUID' \
  -H "apikey: YOUR_ANON_KEY" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"status": "resolved", "resolved_by": "YOUR_USER_ID"}'
```

---

## Troubleshooting

### "Not logged in" Error
```bash
# Check if session exists
cat ~/.specledger/session.json

# If missing or invalid, run login again
/specledger.login
```

### Permission Denied
```bash
# Ensure correct file permissions
chmod 600 ~/.specledger/session.json
```

### Token Expired
```bash
# Check expires_at vs current time
cat ~/.specledger/session.json | jq '.expires_at'
date +%s  # Current Unix timestamp

# If expired, re-login
/specledger.login
```

### Supabase Errors
- **401**: Token invalid or expired → re-login
- **403**: RLS policy violation → check user permissions
- **404**: Comment not found → verify comment ID

---

## Development Notes

### Modifying Commands

1. Edit the relevant `.claude/commands/specledger.*.md` file
2. Test in Claude Code with `/specledger.<command>`
3. No build step needed - commands are interpreted by Claude

### Adding New Commands

1. Create `.claude/commands/specledger.newcommand.md`
2. Add YAML frontmatter with description
3. Define execution steps in markdown
4. Register in skills if needed

### Testing Without Backend

Create mock responses:
```bash
# Mock comment list
echo '[{"id":"1","content":"Test comment","type":"critical","status":"open","author_name":"Test"}]' \
  > /tmp/mock-comments.json
```

---

## Security Reminders

1. **Never commit session.json** - add to .gitignore
2. **Never log tokens** - even in debug output
3. **Always use 600 permissions** - prevent other users from reading
4. **RLS is the guard** - CLI only attaches token, backend enforces access
