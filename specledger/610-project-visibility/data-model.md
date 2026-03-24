# Data Model: Project Visibility (Public/Private)

**Feature**: 610-project-visibility
**Date**: 2026-03-24

## Schema Changes

### Modified: `projects` table

Add visibility column to existing table.

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `visibility` | TEXT | NOT NULL, DEFAULT 'private', CHECK(IN 'public','private') | Project visibility setting |

**Migration**: ALTER TABLE projects ADD COLUMN visibility TEXT NOT NULL DEFAULT 'private' CHECK (visibility IN ('public', 'private'));

### New: `access_requests` table

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `id` | UUID | PK, DEFAULT gen_random_uuid() | Unique request identifier |
| `project_id` | UUID | FK → projects.id, NOT NULL | Target project |
| `requester_id` | UUID | FK → auth.users.id, NOT NULL | User requesting access |
| `message` | TEXT | nullable | Optional message from requester |
| `status` | TEXT | NOT NULL, DEFAULT 'pending', CHECK(IN 'pending','approved','denied') | Request status |
| `reviewed_by` | UUID | FK → auth.users.id, nullable | Owner who reviewed the request |
| `created_at` | TIMESTAMPTZ | NOT NULL, DEFAULT now() | When request was created |
| `updated_at` | TIMESTAMPTZ | NOT NULL, DEFAULT now() | When request was last updated |

**Constraints**:
- UNIQUE: `(project_id, requester_id)` WHERE `status = 'pending'` — prevents duplicate pending requests
- `reviewed_by` must be a project owner (enforced by RLS)

**Indexes**:
- `idx_access_requests_project_status` ON `(project_id, status)`
- `idx_access_requests_requester` ON `(requester_id, created_at DESC)`

### New: `notifications` table

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `id` | UUID | PK, DEFAULT gen_random_uuid() | Unique notification ID |
| `user_id` | UUID | FK → auth.users.id, NOT NULL | Recipient user |
| `type` | TEXT | NOT NULL | Notification type (e.g., 'access_request', 'access_approved', 'access_denied') |
| `title` | TEXT | NOT NULL | Short notification title |
| `body` | TEXT | nullable | Notification detail |
| `reference_id` | UUID | nullable | Related entity ID (e.g., access_request.id) |
| `is_read` | BOOLEAN | NOT NULL, DEFAULT false | Read status |
| `created_at` | TIMESTAMPTZ | NOT NULL, DEFAULT now() | When notification was created |

**Indexes**:
- `idx_notifications_user_unread` ON `(user_id, created_at DESC)` WHERE `is_read = false`

### Modified: `review_comments` table

Make `author_id` nullable and add `is_anonymous` flag.

| Field | Type | Change | Description |
|-------|------|--------|-------------|
| `author_id` | UUID | ALTER: DROP NOT NULL | Nullable for anonymous comments |
| `is_anonymous` | BOOLEAN | ADD: NOT NULL DEFAULT false | Distinguishes anonymous from registered comments |

**Migration**:
```sql
ALTER TABLE review_comments ALTER COLUMN author_id DROP NOT NULL;
ALTER TABLE review_comments ADD COLUMN is_anonymous BOOLEAN NOT NULL DEFAULT false;
```

**Constraint**: CHECK ((is_anonymous = false AND author_id IS NOT NULL) OR (is_anonymous = true AND author_name IS NOT NULL))

## RLS Policies

### projects (modified)

- **SELECT**: Allow all users (authenticated + anon) for public projects. Private projects require project membership.
```sql
-- Public projects: anyone can view
CREATE POLICY "public_projects_select" ON projects
  FOR SELECT USING (visibility = 'public');

-- Private projects: members only
CREATE POLICY "private_projects_select" ON projects
  FOR SELECT USING (
    visibility = 'private'
    AND EXISTS (
      SELECT 1 FROM project_members
      WHERE project_members.project_id = projects.id
      AND project_members.user_id = auth.uid()
    )
  );
```

- **UPDATE** (visibility field): Owner only.
```sql
CREATE POLICY "projects_update_visibility" ON projects
  FOR UPDATE USING (
    EXISTS (
      SELECT 1 FROM project_members
      WHERE project_members.project_id = projects.id
      AND project_members.user_id = auth.uid()
      AND project_members.role = 'owner'
    )
  );
```

### review_comments (modified)

- **SELECT**: Members for private projects. Anyone for public projects.
```sql
CREATE POLICY "comments_select_public" ON review_comments
  FOR SELECT USING (
    EXISTS (
      SELECT 1 FROM changes
      JOIN specs ON specs.id = changes.spec_id
      JOIN projects ON projects.id = specs.project_id
      WHERE changes.id = review_comments.change_id
      AND projects.visibility = 'public'
    )
  );
```

- **INSERT**: Authenticated members for private. Authenticated + anon for public (with anonymous flag).
```sql
-- Registered users on public projects
CREATE POLICY "comments_insert_public_registered" ON review_comments
  FOR INSERT WITH CHECK (
    is_anonymous = false
    AND author_id = auth.uid()
    AND EXISTS (
      SELECT 1 FROM changes
      JOIN specs ON specs.id = changes.spec_id
      JOIN projects ON projects.id = specs.project_id
      WHERE changes.id = review_comments.change_id
      AND projects.visibility = 'public'
    )
  );

-- Anonymous users on public projects
CREATE POLICY "comments_insert_public_anonymous" ON review_comments
  FOR INSERT WITH CHECK (
    is_anonymous = true
    AND author_id IS NULL
    AND author_name IS NOT NULL
    AND author_name != ''
    AND EXISTS (
      SELECT 1 FROM changes
      JOIN specs ON specs.id = changes.spec_id
      JOIN projects ON projects.id = specs.project_id
      WHERE changes.id = review_comments.change_id
      AND projects.visibility = 'public'
    )
  );
```

### access_requests

- **SELECT**: Requester can see own requests. Project owners can see all requests for their projects.
- **INSERT**: Authenticated users only. Cannot request for projects they already belong to.
- **UPDATE**: Project owners only (for approve/deny).

### notifications

- **SELECT**: Users can only see their own notifications.
- **UPDATE**: Users can only update (mark read) their own notifications.
- **INSERT**: System-triggered only (via database triggers or Edge Functions).

## State Transitions

### Access Request Lifecycle

```
[New Request] → pending → approved → (user added to project_members as 'editor')
                       → denied   → (no membership change)

[Project goes private] → all pending requests → cancelled (deleted)
```

### Project Visibility Toggle

```
private → public:  No data migration needed. RLS policies immediately grant broader access.
public → private:  Cancel all pending access requests. Non-members immediately lose access.
```

## Entity Relationships

```
projects (1) ──── (*) specs ──── (*) changes ──── (*) review_comments
    │                                                      │
    │ visibility                                     is_anonymous
    │                                                author_name (display name)
    │
    ├──── (*) project_members (role: owner|editor|viewer)
    │
    ├──── (*) access_requests (status: pending|approved|denied)
    │              │
    │              └──── (1) notifications (type: access_request|access_approved|access_denied)
    │
    └──── (*) notifications (via user_id)
```
