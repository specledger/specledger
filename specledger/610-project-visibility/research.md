# Research: Project Visibility (Public/Private)

**Feature**: 610-project-visibility
**Date**: 2026-03-24

## Prior Work

### CLI Authentication (008-cli-auth)
- Browser-based OAuth login via `sl auth login`
- Credentials stored at `~/.specledger/credentials.json` (0600 permissions)
- Token refresh with proactive 30-second-before-expiry strategy
- AuthProvider interface for testability (`pkg/cli/auth/`)

### Comment Management (009-add-login-and-comment-commands, 136-revise-comments)
- `sl comment list|show|reply|resolve` commands implemented
- PostgREST client in `pkg/cli/comment/client.go` with `DoWithRetry()` for 401 handling
- ReviewComment table: `review_comments` with `author_id`, `author_name`, `author_email`
- RLS policies enforce project membership for SELECT/INSERT/UPDATE
- Comments linked via chain: projects → specs → changes → review_comments

### Existing Database Schema
- **projects**: `id`, `repo_owner`, `repo_name`, `default_branch`
- **specs**: `id`, `project_id`, `spec_key`, `phase`
- **changes**: `id`, `spec_id`, `head_branch`, `base_branch`, `state`
- **review_comments**: Full comment model with thread support
- **sessions**: Session capture with project membership RLS
- **project_members**: Referenced in RLS policies (exists in Supabase, used for membership checks)

## Research Decisions

### R1: Visibility Storage

**Decision**: Add `visibility` column to existing `projects` table.
**Rationale**: Project visibility is a simple property of the project entity — no need for a separate table. A single TEXT column with CHECK constraint is sufficient.
**Alternatives considered**:
- Separate `project_settings` table → Over-engineering for a single field (YAGNI)
- Boolean `is_public` column → TEXT enum is more readable and extensible if needed

### R2: Anonymous Comment Identity

**Decision**: Extend `review_comments` to support anonymous comments by making `author_id` nullable and adding `is_anonymous` flag.
**Rationale**: The existing comment system requires `author_id` (FK to auth.users). Anonymous users have no auth.users record, so `author_id` must be nullable. The `author_name` field already exists and can hold the display name.
**Alternatives considered**:
- Create a separate `anonymous_comments` table → Splits comment queries, violates DRY
- Create ghost/system user for anonymous comments → Adds unnecessary complexity, misrepresents identity
- Use `is_ai_generated` field to indicate anonymous → Semantic mismatch

### R3: Access Request Storage

**Decision**: New `access_requests` table with status workflow.
**Rationale**: Access requests are a distinct entity with their own lifecycle (pending → approved/denied). They don't fit into existing tables.
**Alternatives considered**:
- Reuse `project_members` with a "pending" role → Mixes approved and pending states, complicates RLS

### R4: RLS Policy Strategy for Public Projects

**Decision**: Modify existing RLS policies to check project visibility. Public projects allow SELECT for all authenticated AND anonymous users. INSERT (comments) checks visibility + requires display name for anonymous.
**Rationale**: Supabase RLS is the security boundary. The `anon` key (already available) can be used for unauthenticated access to public projects.
**Alternatives considered**:
- Application-level access control only → Bypasses Supabase security model, violates contract-first testing principle

### R5: Rate Limiting for Anonymous Comments

**Decision**: Implement via Supabase Edge Function with IP-based rate limiting.
**Rationale**: RLS policies cannot enforce rate limits. A lightweight Edge Function wrapping the comment insert provides the spam protection without a full middleware layer.
**Alternatives considered**:
- Client-side rate limiting → Easily bypassed
- pgREST middleware → Not available in Supabase hosted
- No rate limiting initially → Acceptable MVP, defer to later. **Selected for MVP per YAGNI.**

### R6: Notification Delivery

**Decision**: In-platform notification via new `notifications` table. MVP: poll-based (no real-time push).
**Rationale**: Spec states notifications are within SpecLedger platform. A simple table with unread/read status is the minimum viable approach.
**Alternatives considered**:
- Supabase Realtime subscriptions → More complex, not needed for MVP
- Email notifications → Out of scope per spec assumptions
- Skip notifications in MVP → Access request workflow needs at least basic notification to be usable

### R7: CLI vs Web for Access Requests

**Decision**: Both CLI (`sl access request|approve|deny|list`) and web UI support. CLI follows existing Data CRUD pattern.
**Rationale**: Follows the 4-layer model — L1 CLI provides data operations, web UI provides the visual interface. Consistent with existing `sl comment` and `sl issue` patterns.
**Alternatives considered**:
- Web-only → Breaks the CLI-first principle
