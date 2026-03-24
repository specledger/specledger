# API Contracts: Project Visibility

**Feature**: 610-project-visibility
**Date**: 2026-03-24

## PostgREST Endpoints

### Project Visibility

#### Get Project (public)
```
GET /rest/v1/projects?repo_owner=eq.{owner}&repo_name=eq.{name}&select=id,repo_owner,repo_name,default_branch,visibility
Headers:
  apikey: {anon_key}
  # No Authorization header needed for public projects
```

#### Update Project Visibility
```
PATCH /rest/v1/projects?id=eq.{project_id}
Headers:
  Authorization: Bearer {access_token}
  apikey: {anon_key}
  Content-Type: application/json
  Prefer: return=representation
Body:
  {"visibility": "public"}  # or "private"
```

**RLS**: Only project owners can update.

### Anonymous Comments

#### Create Anonymous Comment (public project)
```
POST /rest/v1/review_comments
Headers:
  apikey: {anon_key}
  Content-Type: application/json
  Prefer: return=representation
  # No Authorization header (anonymous)
Body:
  {
    "change_id": "{change_uuid}",
    "file_path": "specledger/610-project-visibility/spec.md",
    "content": "This looks great!",
    "author_name": "Jane Doe",
    "is_anonymous": true
  }
```

**RLS**: Allowed only on public projects. `author_id` is NULL, `author_name` required.

#### Create Registered Comment (public project, non-member)
```
POST /rest/v1/review_comments
Headers:
  Authorization: Bearer {access_token}
  apikey: {anon_key}
  Content-Type: application/json
  Prefer: return=representation
Body:
  {
    "change_id": "{change_uuid}",
    "file_path": "specledger/610-project-visibility/spec.md",
    "content": "Suggestion: consider adding rate limits",
    "author_id": "{user_uuid}",
    "author_name": "John Smith",
    "author_email": "john@example.com",
    "is_anonymous": false
  }
```

### Access Requests

#### Create Access Request
```
POST /rest/v1/access_requests
Headers:
  Authorization: Bearer {access_token}
  apikey: {anon_key}
  Content-Type: application/json
  Prefer: return=representation
Body:
  {
    "project_id": "{project_uuid}",
    "requester_id": "{user_uuid}",
    "message": "I'd like to contribute to this spec"
  }
```

**RLS**: Authenticated users only. Fails if already a member or has pending request.

#### List Access Requests (project owner)
```
GET /rest/v1/access_requests?project_id=eq.{project_id}&status=eq.pending&select=id,requester_id,message,status,created_at&order=created_at.asc
Headers:
  Authorization: Bearer {access_token}
  apikey: {anon_key}
```

**RLS**: Only project owners see requests for their projects.

#### Approve/Deny Access Request
```
PATCH /rest/v1/access_requests?id=eq.{request_id}
Headers:
  Authorization: Bearer {access_token}
  apikey: {anon_key}
  Content-Type: application/json
  Prefer: return=representation
Body:
  {
    "status": "approved",  # or "denied"
    "reviewed_by": "{owner_uuid}"
  }
```

**RLS**: Only project owners. On approve, a database trigger inserts into `project_members` with role='editor' and creates a notification.

### Notifications

#### List Unread Notifications
```
GET /rest/v1/notifications?user_id=eq.{user_id}&is_read=eq.false&select=id,type,title,body,reference_id,created_at&order=created_at.desc
Headers:
  Authorization: Bearer {access_token}
  apikey: {anon_key}
```

#### Mark Notification Read
```
PATCH /rest/v1/notifications?id=eq.{notification_id}
Headers:
  Authorization: Bearer {access_token}
  apikey: {anon_key}
  Content-Type: application/json
Body:
  {"is_read": true}
```

## CLI Commands (L1)

### `sl project visibility [public|private]`
- **Pattern**: Data CRUD (Environment subcommand)
- **No args**: Display current visibility
- **With arg**: Set visibility (owner only)
- **Output**: Compact: `Project visibility: public` / JSON: `{"visibility": "public"}`

### `sl access request [--message "..."]`
- **Pattern**: Data CRUD
- **Context**: Auto-detects project from current repo
- **Output**: Compact: `Access request sent to project owner` / JSON: full request object

### `sl access list`
- **Pattern**: Data CRUD
- **Context**: Auto-detects project, shows requests for owners
- **Output**: Compact: table of pending requests / JSON: array of request objects

### `sl access approve <request-id>`
- **Pattern**: Data CRUD
- **Output**: Compact: `Approved. {user} added as editor.` / JSON: updated request + member objects

### `sl access deny <request-id>`
- **Pattern**: Data CRUD
- **Output**: Compact: `Denied. {user} notified.` / JSON: updated request object

## Database Triggers

### On access_request status change to 'approved'
```sql
CREATE OR REPLACE FUNCTION handle_access_request_approval()
RETURNS TRIGGER AS $$
BEGIN
  IF NEW.status = 'approved' AND OLD.status = 'pending' THEN
    -- Add user as editor
    INSERT INTO project_members (project_id, user_id, role)
    VALUES (NEW.project_id, NEW.requester_id, 'editor')
    ON CONFLICT DO NOTHING;

    -- Notify requester
    INSERT INTO notifications (user_id, type, title, reference_id)
    VALUES (NEW.requester_id, 'access_approved', 'Your access request was approved', NEW.id);
  END IF;

  IF NEW.status = 'denied' AND OLD.status = 'pending' THEN
    -- Notify requester
    INSERT INTO notifications (user_id, type, title, reference_id)
    VALUES (NEW.requester_id, 'access_denied', 'Your access request was denied', NEW.id);
  END IF;

  RETURN NEW;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;
```

### On project visibility change to 'private'
```sql
CREATE OR REPLACE FUNCTION handle_visibility_change()
RETURNS TRIGGER AS $$
BEGIN
  IF NEW.visibility = 'private' AND OLD.visibility = 'public' THEN
    -- Cancel all pending access requests
    DELETE FROM access_requests
    WHERE project_id = NEW.id AND status = 'pending';
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;
```

## Error Responses

| Status | Scenario | Suggested Action |
|--------|----------|-----------------|
| 401 | Token expired | `sl auth login` |
| 403 | Not a project owner (visibility change) | "Only project owners can change visibility" |
| 403 | Not a project member (private project) | "This project is private. Request access with `sl access request`" |
| 409 | Duplicate pending access request | "You already have a pending request for this project" |
| 404 | Project not found or private | "Project not found. If private, request access with `sl access request`" |
