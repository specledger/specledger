# Feature Specification: Project Visibility (Public/Private)

**Feature Branch**: `610-project-visibility`
**Created**: 2026-03-24
**Status**: Draft
**Input**: User description: "Add public/private visibility for SpecLedger projects. Public projects allow registered users and anonymous users to view and comment (anonymous users must provide a display name). Edit access requires permission request (registered users only)."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Project Owner Sets Visibility (Priority: P1)

As a project owner, I want to set my project as public or private so that I can control who has access to view and interact with my project artifacts.

**Why this priority**: This is the foundational capability — without visibility settings, no other access control behavior can be enforced.

**Independent Test**: Can be fully tested by creating a project and toggling its visibility setting. Delivers immediate value by giving owners control over project exposure.

**Acceptance Scenarios**:

1. **Given** a project owner is on the project settings page, **When** they set visibility to "public", **Then** the project becomes viewable by anyone (registered users and anonymous visitors).
2. **Given** a project owner is on the project settings page, **When** they set visibility to "private", **Then** only project members can access the project.
3. **Given** a project is set to public, **When** the owner changes it to private, **Then** all non-member access is immediately revoked and pending access requests are cancelled.

---

### User Story 2 - Public Project Browsing and Commenting (Priority: P1)

As any user (registered or anonymous), I want to view and comment on public projects so that I can contribute feedback and participate in the specification process without needing membership or even an account.

**Why this priority**: This is the core value proposition of public projects — enabling the broadest possible community engagement with specifications, including users who haven't registered yet.

**Independent Test**: Can be tested by: (a) logging in as a non-member user, viewing specs, and posting a comment; (b) visiting as an anonymous user, entering a display name, and posting a comment.

**Acceptance Scenarios**:

1. **Given** a registered user is not a member of a public project, **When** they navigate to the project, **Then** they can view all project specifications and artifacts in read-only mode.
2. **Given** a registered user is viewing a public project, **When** they submit a comment on a specification, **Then** the comment is saved with their account name and visible to all project participants.
3. **Given** an anonymous (not logged in) user navigates to a public project, **When** they view the project, **Then** they can see all specifications and artifacts in read-only mode without needing to log in.
4. **Given** an anonymous user is viewing a public project, **When** they want to post a comment, **Then** they are prompted to enter a display name before submitting.
5. **Given** an anonymous user has entered a display name, **When** they submit a comment, **Then** the comment is saved with the provided display name and marked as an anonymous comment.
6. **Given** a registered user is not a member of a private project, **When** they try to access the project, **Then** they receive an "access denied" message.
7. **Given** an anonymous user tries to access a private project, **When** they navigate to the project URL, **Then** they receive an "access denied" message.

---

### User Story 3 - Request Edit Access to Public Project (Priority: P2)

As a registered user viewing a public project, I want to request edit access so that I can contribute changes to specifications and artifacts.

**Why this priority**: Enables a contributor workflow where interested users can escalate from viewer to editor through a managed approval process.

**Independent Test**: Can be tested by requesting access as a non-member, then having a project owner approve/deny the request.

**Acceptance Scenarios**:

1. **Given** a registered user is viewing a public project they are not a member of, **When** they click "Request Edit Access", **Then** an access request is sent to the project owner.
2. **Given** a project owner has a pending access request, **When** they approve the request, **Then** the requesting user gains edit permissions on project artifacts.
3. **Given** a project owner has a pending access request, **When** they deny the request, **Then** the requesting user is notified and retains view-only access.
4. **Given** a user already has a pending access request, **When** they try to request again, **Then** they are informed that a request is already pending.

---

### User Story 4 - Manage Access Requests (Priority: P2)

As a project owner, I want to view and manage pending access requests so that I can control who gains edit permissions on my project.

**Why this priority**: Completes the access request workflow by providing owners with management tools.

**Independent Test**: Can be tested by having multiple users request access, then reviewing and acting on each request.

**Acceptance Scenarios**:

1. **Given** a project owner navigates to access management, **When** there are pending requests, **Then** they see a list with requester name, date, and optional message.
2. **Given** a project owner views access requests, **When** they approve a request, **Then** the user is added as a project member with edit permissions.
3. **Given** a project owner views access requests, **When** they deny a request, **Then** the request is removed and the user is notified.

---

### Edge Cases

- What happens when a public project with active external commenters is switched to private? External comments remain visible but users lose the ability to add new comments.
- What happens when a user whose access request is pending gets invited directly by the owner? The pending request is automatically resolved as approved.
- What happens when a project owner removes a member who was granted access via request? The member loses edit access and reverts to view-only (if project is public).
- How does the system handle a project owner trying to set visibility when they don't have owner-level permissions? Only users with the "owner" role can change project visibility.
- What happens if an anonymous user enters a display name that matches an existing registered user's name? The comment is still posted with the display name but clearly marked as "Guest" to avoid impersonation.
- Can anonymous users comment without any rate limiting? Anonymous comments should have basic rate limiting (e.g., per IP) to prevent spam.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST support two project visibility states: "public" and "private".
- **FR-002**: System MUST default new projects to "private" visibility.
- **FR-003**: Only project owners MUST be able to change project visibility settings.
- **FR-004**: Public projects MUST allow any user (registered or anonymous) to view all specifications, artifacts, and comments.
- **FR-005**: Public projects MUST allow registered users to post comments using their account name.
- **FR-005a**: Public projects MUST allow anonymous (not logged in) users to post comments after providing a display name.
- **FR-005b**: Comments from anonymous users MUST be visually distinguished from comments by registered users (e.g., labeled "Anonymous" or "Guest").
- **FR-006**: Public projects MUST NOT allow non-members to edit specifications or artifacts without explicit permission.
- **FR-007**: System MUST provide a mechanism for registered non-member users to request edit access on public projects. Anonymous users MUST register and log in before requesting edit access.
- **FR-008**: System MUST notify project owners when a new access request is submitted.
- **FR-009**: Project owners MUST be able to approve or deny access requests.
- **FR-010**: System MUST notify requesting users when their access request is approved or denied.
- **FR-011**: When a project is changed from public to private, system MUST immediately revoke all non-member access.
- **FR-012**: System MUST prevent duplicate access requests from the same user for the same project.
- **FR-013**: Private projects MUST only be accessible to explicitly added project members.
- **FR-014**: Anonymous users MUST provide a non-empty display name before posting a comment.
- **FR-015**: Anonymous users MUST NOT be able to request edit access — they must register and log in first.

### Key Entities

- **Project Visibility**: A property of a project indicating whether it is "public" or "private". Controls the default access level for non-member users.
- **Access Request**: A record representing a non-member user's request for edit permissions on a public project. Contains requester identity, target project, request date, status (pending/approved/denied), and optional message.
- **Project Member**: A user with explicit access to a project. Members have roles (owner, editor, viewer) that determine their permissions.
- **Anonymous Commenter**: A non-authenticated visitor who provides a display name to post comments on public projects. Has no persistent identity, cannot request edit access, and their comments are labeled as "Guest".

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Project owners can toggle between public and private visibility in under 10 seconds.
- **SC-002**: Non-member users can view a public project's specifications within 3 seconds of navigation.
- **SC-003**: Access requests are delivered to project owners within 30 seconds of submission.
- **SC-004**: 95% of access request approvals/denials are reflected in user permissions within 10 seconds.
- **SC-005**: Switching from public to private immediately prevents new non-member access on the next request.

### Previous work

- **CLI Authentication (008-cli-auth)**: Existing auth infrastructure for user identity and session management.
- **Comment Management (136-revise-comments)**: Existing comment system with PostgREST client and auth integration.

## Dependencies & Assumptions

### Dependencies

- Existing authentication system (login, session tokens) is functional and stable.
- User registration system exists and provides user identity.

### Assumptions

- "Registered user" means a user who has completed SpecLedger registration and can authenticate.
- "Anonymous user" means a visitor who has not logged in — they have no persistent identity across sessions.
- Project visibility is a project-level setting, not per-specification.
- Comment permissions on public projects follow the same rules as existing comment functionality — no moderation queue for comments (direct post).
- Anonymous comments are posted immediately (no moderation queue) but are labeled distinctly from registered user comments.
- Anonymous user display names are not unique or validated against existing accounts — they are free-text labels for attribution only.
- Access requests include only a simple approve/deny workflow (no role selection during request — granted users receive "editor" role by default).
- Only registered users can request edit access — anonymous users must register first.
- Notifications are delivered within the SpecLedger platform (no external email/push notifications required in initial version).
- Basic rate limiting for anonymous comments is assumed to prevent spam abuse.
