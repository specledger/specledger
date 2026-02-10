# Feature Specification: CLI Authentication and Comment Management

**Feature Branch**: `009-add-login-and-comment-commands`
**Created**: 2026-02-10
**Status**: Draft
**Input**: User description: "Add login, logout, comment, and resolve-comment CLI commands based on authentication integration guide and comment resolution design"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - CLI Login via Browser (Priority: P1)

A user wants to authenticate the Specledger CLI to access protected features like viewing and resolving comments on specifications.

**Why this priority**: Authentication is the foundation for all protected operations. Without login capability, users cannot access any authenticated features.

**Independent Test**: Can be fully tested by running the login command, completing browser authentication, pasting the token, and verifying session file creation. Delivers immediate value as authentication gateway.

**Acceptance Scenarios**:

1. **Given** the user is not logged in, **When** they run `/specledger.login`, **Then** they see instructions with a URL to open in their browser
2. **Given** the user has opened the auth URL and logged in, **When** they copy the token and paste it into the CLI prompt, **Then** the CLI validates the token format and saves the session
3. **Given** the session is saved, **When** the user checks `~/.specledger/session.json`, **Then** the file exists with permission 600 (owner read/write only)
4. **Given** the user pastes an invalid token format, **When** the CLI validates it, **Then** the CLI shows an error message "Invalid session format" and does not save

---

### User Story 2 - View Comments on Specification (Priority: P2)

A logged-in user wants to view comments left by reviewers on a specification file so they can see feedback that needs addressing.

**Why this priority**: Viewing comments is the prerequisite for resolving them. Users need to see what feedback exists before taking action.

**Independent Test**: Can be fully tested by running the comment view command on a file with existing comments and verifying the output format displays correctly.

**Acceptance Scenarios**:

1. **Given** the user is logged in and a spec file has comments, **When** they run `/specledger.comment` with a file path, **Then** they see a formatted list of comments with ID, type icon, author, status, and content
2. **Given** the user is logged in but the file has no comments, **When** they run `/specledger.comment`, **Then** they see a message indicating no comments found
3. **Given** the user is not logged in, **When** they run `/specledger.comment`, **Then** they see an error "Not logged in. Run `/specledger.login` first."

---

### User Story 3 - Resolve Comment (Priority: P3)

A logged-in user wants to mark a comment as resolved after addressing the feedback, so reviewers know their feedback has been handled.

**Why this priority**: Resolving comments completes the feedback loop and enables workflow progression. Depends on ability to view comments first.

**Independent Test**: Can be fully tested by resolving a specific comment by ID and verifying the status change in the system.

**Acceptance Scenarios**:

1. **Given** the user is logged in and a comment exists with status "open", **When** they run `/specledger.resolve-comment` with the comment ID, **Then** the comment status changes to "resolved" and they see a success confirmation
2. **Given** the user tries to resolve a non-existent comment ID, **When** they run `/specledger.resolve-comment`, **Then** they see an error "Comment not found"
3. **Given** the user is not logged in, **When** they run `/specledger.resolve-comment`, **Then** they see an error "Not logged in. Run `/specledger.login` first."

---

### User Story 4 - CLI Logout (Priority: P4)

A user wants to log out of the Specledger CLI to clear their credentials from the local machine.

**Why this priority**: Logout is a secondary operation used less frequently. Important for security hygiene but not blocking core workflows.

**Independent Test**: Can be fully tested by running logout command and verifying session file is removed.

**Acceptance Scenarios**:

1. **Given** the user is logged in, **When** they run `/specledger.logout`, **Then** the session file at `~/.specledger/session.json` is deleted and they see a confirmation message
2. **Given** the user is not logged in, **When** they run `/specledger.logout`, **Then** they see a message indicating they are not logged in (no error, graceful handling)

---

### Edge Cases

- What happens when the session file exists but is corrupted/invalid JSON? → Show clear error and suggest re-login
- What happens when the session token has expired? → Show clear error and suggest re-login (token expiry check based on `expires_at` field)
- What happens when the Supabase server is unreachable during comment operations? → Show network error with retry suggestion
- What happens when user lacks permissions to create `~/.specledger/` directory? → Show permission error with guidance
- What happens when multiple users share the same machine? → Each user has isolated session in their home directory

## Requirements *(mandatory)*

### Functional Requirements

**Authentication:**
- **FR-001**: System MUST provide a login command that displays a URL for browser-based authentication
- **FR-002**: System MUST accept a JSON token pasted by the user containing `access_token`, `refresh_token`, `expires_at`, and `user_id` fields
- **FR-003**: System MUST validate that the pasted token contains all required fields before saving
- **FR-004**: System MUST store the session at `~/.specledger/session.json` with file permission 600
- **FR-005**: System MUST create the `~/.specledger/` directory if it does not exist
- **FR-006**: System MUST provide a logout command that removes the session file
- **FR-007**: System MUST provide an `requireAuth` helper that checks for valid session before protected operations

**Comment Operations:**
- **FR-008**: System MUST provide a command to view comments on a specification file
- **FR-009**: System MUST display comments with: ID, type indicator, author, status, and content
- **FR-010**: System MUST provide a command to resolve a comment by its ID
- **FR-011**: System MUST update comment status to "resolved" and record who resolved it
- **FR-012**: System MUST attach the Authorization header with Bearer token for all Supabase API calls

**Security:**
- **FR-013**: System MUST NOT log tokens to console or files
- **FR-014**: System MUST NOT commit session.json to version control (enforce via .gitignore)
- **FR-015**: System MUST NOT verify JWT signature locally (server validates)

### Key Entities

- **Session**: User authentication state containing access_token, refresh_token, expires_at (unix timestamp), and user_id
- **Comment**: Feedback on a specification file with ID, type (issue/suggestion), author, status (open/resolved), content, and file reference

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can complete the login flow (open URL → authenticate → paste token) in under 2 minutes
- **SC-002**: Users can view comments on any specification file within 5 seconds of running the command
- **SC-003**: 100% of resolve operations correctly update comment status when valid comment ID is provided
- **SC-004**: Session files have correct 600 permissions on all supported platforms (macOS, Linux)
- **SC-005**: All protected commands show clear "not logged in" message when session is missing

### Previous work

### Epic: SL-008 - CLI Authentication

- **008-cli-auth**: Initial implementation of browser-based OAuth authentication for the CLI with session storage
