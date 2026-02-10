# Research: CLI Authentication and Comment Management

**Feature Branch**: `009-add-login-and-comment-commands`
**Created**: 2026-02-10

## Prior Work

### Related Features

| Feature | Status | Relationship |
|---------|--------|--------------|
| 008-cli-auth | Completed | Browser-based OAuth for CLI binary (`sl auth login/logout/status/refresh`) |

### Key Finding: Two Authentication Systems

The project has **two distinct authentication systems**:

1. **CLI Binary Commands** (`sl auth ...`)
   - Location: `pkg/cli/commands/auth.go`, `pkg/cli/auth/`
   - Flow: Browser → OAuth callback server → save credentials
   - Storage: `~/.specledger/credentials.json`
   - Format: `expires_in` (duration) + `created_at` (timestamp)

2. **Claude Code Slash Commands** (`/specledger.login`, etc.)
   - Location: `.claude/commands/specledger.*.md`
   - Flow: User copies token from browser → pastes into Claude
   - Storage: `~/.specledger/session.json`
   - Format: `expires_at` (absolute timestamp)

These are intentionally separate:
- CLI commands require a running callback server for OAuth
- Slash commands need a simpler token-paste flow that works within Claude's context

---

## Technical Decisions

### Decision 1: Slash Command Implementation

**Question**: How should the slash commands be implemented?

**Decision**: Markdown-based command definitions in `.claude/commands/`

**Rationale**:
- Claude Code executes these as agent workflows, not compiled code
- Markdown commands define step-by-step instructions for Claude
- Shell commands (curl, file operations) are executed via Claude's Bash tool
- No Go code changes needed for slash commands

**Alternatives Rejected**:
- Implementing as Go CLI commands: Would duplicate 008-cli-auth functionality
- Using a shared library: Slash commands run in Claude's context, not as Go binaries

---

### Decision 2: Session Storage Format

**Question**: Should slash commands use the existing credentials.json or separate session.json?

**Decision**: Separate `~/.specledger/session.json` with simplified format

**Session Format**:
```json
{
  "access_token": "JWT_ACCESS_TOKEN",
  "refresh_token": "JWT_REFRESH_TOKEN",
  "expires_at": 1712345678,
  "user_id": "uuid"
}
```

**Rationale**:
- `expires_at` (absolute timestamp) is simpler than `expires_in` + `created_at` calculation
- Matches the Auth UI contract defined in design documents
- Separation prevents interference between CLI and Claude workflows
- Easier for Claude to validate (just compare `expires_at` with current time)

**Alternatives Rejected**:
- Reusing credentials.json: Format mismatch, CLI relies on its specific structure
- Converting between formats: Adds complexity with no benefit

---

### Decision 3: Comment API Integration

**Question**: How should comment operations interact with the backend?

**Decision**: Direct Supabase REST API calls via curl

**API Pattern**:
```bash
# Fetch comments
curl -s 'https://specledger.supabase.co/rest/v1/comments?select=*&file_path=eq.$FILE' \
  -H "apikey: $SUPABASE_ANON_KEY" \
  -H "Authorization: Bearer $ACCESS_TOKEN"

# Resolve comment
curl -s -X PATCH 'https://specledger.supabase.co/rest/v1/comments?id=eq.$ID' \
  -H "apikey: $SUPABASE_ANON_KEY" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"status": "resolved", "resolved_by": "$USER_ID"}'
```

**Rationale**:
- Supabase REST API is already the backend (per design docs)
- RLS policies handle authorization based on JWT
- No custom backend code needed
- curl commands are portable and debuggable

**Alternatives Rejected**:
- GraphQL: Supabase uses REST/PostgREST primarily
- Custom Go client: Not needed for slash commands (Claude runs curl)

---

### Decision 4: Auth URL Configuration

**Question**: How should the auth URL be configured?

**Decision**: Use production URL with environment variable override

**URLs**:
- Production: `https://specledger.io/cli/auth`
- Development: Override via browser URL or env variable
- Supabase: `https://specledger.supabase.co`

**Rationale**:
- Matches existing CLI auth URL pattern
- Environment variables allow local development
- No hardcoded secrets (anon key is public)

---

## Technical Context Summary

Based on research, the Technical Context for plan.md:

| Field | Value |
|-------|-------|
| Language/Version | Go 1.24+ (existing CLI), Markdown (slash commands) |
| Primary Dependencies | Cobra (CLI), Supabase REST API, curl (slash commands) |
| Storage | File-based (`~/.specledger/session.json`), Supabase (comments) |
| Testing | Manual testing via Claude Code execution |
| Target Platform | macOS, Linux (WSL compatible) |
| Project Type | CLI tool with Claude Code integration |
| Performance Goals | N/A (interactive user commands) |
| Constraints | No secrets in code, file permission 600, RLS for auth |

---

## Existing Implementation Review

### Files Already Created (from git status)

```text
A  .claude/commands/specledger.comment.md     # View comments command
A  .claude/commands/specledger.login.md       # Login command
A  .claude/commands/specledger.logout.md      # Logout command
A  .claude/commands/specledger.resolve-comment.md  # Resolve comment command
```

### Implementation Status

| Command | File Created | Documented | Tested |
|---------|--------------|------------|--------|
| /specledger.login | Yes | Yes | No |
| /specledger.logout | Yes | Yes | No |
| /specledger.comment | Yes | Yes | No |
| /specledger.resolve-comment | Yes | Yes | No |

### Key Observations

1. **Commands are already defined** - The `.claude/commands/` files contain complete implementation instructions
2. **No Go code needed** - These are Claude Code agent workflows, not CLI binaries
3. **Testing gap** - Commands need validation with actual Supabase backend

---

## Open Questions (Out of Scope for 009)

1. **Token refresh**: Design mentions auto-refresh but marks as "out of scope for this phase"
2. **Supabase RLS policies**: Assumes backend has proper `comments` table with RLS
3. **Auth UI**: Placeholder URL used; actual UI deployment is separate concern

---

## Recommendations

1. **Validate existing slash commands** - Run through login/comment flows manually
2. **Document Supabase requirements** - List required tables, columns, RLS policies
3. **Add error handling** - Current commands assume happy path; need robust error messages
