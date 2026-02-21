# Feature Specification: Revise Comments CLI Command

**Feature Branch**: `136-revise-comments`
**Created**: 2026-02-19
**Status**: Draft
**Input**: User description: "As a User I want to revise comments that were added on to artifacts in specledger via an integrated `sl revise` command in the Go CLI binary."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Branch Selection and Comment Fetching (Priority: P1)

As a developer, I want to run `sl revise [optional-branch-name]` and have the command determine which branch to work on — prefilling the current branch if I'm already on one, or presenting a list of branches to choose from if I'm not — so that artifact comments are always fetched in the context of the correct branch.

**Why this priority**: This is the primary entry point. The branch determines which comments to fetch and which files to revise. Getting the branch right is a prerequisite for all other functionality.

**Independent Test**: Can be fully tested by running `sl revise` from various branch states (on a feature branch, on main, with an explicit argument) and verifying the correct branch is selected and comments are fetched.

**Acceptance Scenarios**:

1. **Given** I am authenticated and on branch `136-revise-comments`, **When** I run `sl revise` (no argument), **Then** the current branch is prefilled and I am asked to confirm or choose a different branch.
2. **Given** I am authenticated and on `main`, **When** I run `sl revise` (no argument), **Then** I see a list of branches (local and remote) that have unresolved artifact comments and I can select exactly one.
3. **Given** I run `sl revise 009-feature-name`, **When** the branch exists, **Then** that branch is used directly without a selection prompt.
4. **Given** I am not authenticated, **When** I run `sl revise`, **Then** I see an error directing me to run `sl auth login` first.
5. **Given** the selected branch is different from my current branch, **When** a checkout is needed, **Then** the system checks for uncommitted local changes and offers to stash them before switching.
6. **Given** I am on a feature branch and confirm it, **When** no unresolved comments exist, **Then** the command exits cleanly with "No unresolved comments found."

---

### User Story 2 - Artifact Multi-Select with Comment Counts (Priority: P2)

As a developer with fetched comments, I want to see a multi-select prompt listing only artifacts that have unresolved comments along with comment counts, so I can choose which artifacts to focus on.

**Why this priority**: Filtering to only artifacts with unresolved comments and showing counts lets the user triage efficiently. Without this, the user would have to process all comments linearly.

**Independent Test**: Can be tested by fetching comments for a spec with multiple artifact files, verifying only files with unresolved comments appear, selecting a subset, and confirming only those are queued for processing.

**Acceptance Scenarios**:

1. **Given** comments have been fetched and 3 artifacts have unresolved comments, **When** the artifact selection prompt appears, **Then** I see a multi-select list showing each artifact path and the count of unresolved comments (e.g., `spec.md (4 comments)`, `plan.md (2 comments)`).
2. **Given** comments have been fetched but all comments are already resolved, **When** the command reaches the selection step, **Then** it exits immediately with "All comments are resolved. Nothing to process."
3. **Given** I select 2 out of 3 artifacts, **When** I confirm selection, **Then** only comments from the 2 selected artifacts are queued for the processing loop.

---

### User Story 3 - Interactive Comment Processing Loop (Priority: P3)

As a developer reviewing comments, I want to iterate through each comment one at a time and choose to process it, skip it, or quit the loop, so I have full control over which feedback I address.

**Why this priority**: The core interaction loop. Seeing comment details with context and choosing an action is the main value proposition of the command.

**Independent Test**: Can be tested by entering the processing loop with multiple comments and verifying that "process" queues the comment for revision, "skip" excludes it from the final prompt, and "quit" exits the loop immediately.

**Acceptance Scenarios**:

1. **Given** I am in the processing loop and a comment is displayed, **When** I choose "process", **Then** I am prompted to optionally provide guidance text for the LLM, and the comment is added to the revision batch.
2. **Given** I am in the processing loop and a comment is displayed, **When** I choose "skip", **Then** the comment is excluded from the revision batch and the next comment is shown.
3. **Given** I am in the processing loop, **When** I choose "quit", **Then** the loop ends immediately and I proceed to the revision summary with only the comments I processed so far.
4. **Given** a comment is displayed, **Then** I see: the file path, the line number (if available), the selected text (if available), the comment content, the author, and the date.

---

### User Story 4 - Revision Prompt Generation and Editor Launch (Priority: P4)

As a developer who has selected comments to process, I want to see a single combined revision prompt that merges all processed comments (across all selected artifacts) with my guidance, and I want to edit this prompt in my preferred editor before sending it to my coding agent.

**Why this priority**: The prompt is the bridge between human review feedback and automated code changes. Allowing the user to refine the prompt ensures the agent receives clear, well-scoped instructions.

**Independent Test**: Can be tested by processing comments, verifying a prompt is generated from an embedded template, launching the configured editor, modifying the prompt, and confirming the modified text is displayed after the editor closes.

**Acceptance Scenarios**:

1. **Given** I have processed 3 comments with guidance, **When** the revision summary is shown, **Then** I see each comment's file path, comment content, selected text, and my guidance organized clearly.
2. **Given** the revision prompt has been generated, **When** I choose to edit it, **Then** my shell-configured editor opens with the prompt content in a temporary file.
3. **Given** I have edited the prompt in my editor and saved/closed it, **When** the editor exits, **Then** I see the modified prompt text displayed in the terminal and can confirm or re-edit it.
4. **Given** a prompt has been generated, **When** a token count estimate is computed, **Then** I see the approximate token count with a warning if the prompt is very short (under 100 tokens — may lack context) or very long (over 8000 tokens — may reduce agent effectiveness).

---

### User Story 5 - Launch Coding Agent with Prompt (Priority: P5)

As a developer with a finalized revision prompt, I want to launch my configured coding agent (e.g., Claude Code) with that prompt, so the agent can make the requested changes to my artifacts.

**Why this priority**: This is the payoff — automated revision of artifacts based on structured feedback. Without this, the user would need to manually copy-paste the prompt.

**Independent Test**: Can be tested by generating a revision prompt and verifying the configured agent command is invoked with the prompt content provided as input.

**Acceptance Scenarios**:

1. **Given** I have a finalized prompt, **When** I choose to launch the coding agent, **Then** the configured agent command is executed with the prompt content provided as input.
2. **Given** no coding agent is configured, **When** I choose to launch, **Then** I see a message explaining how to configure one and am prompted for a filename to write the prompt to.
3. **Given** the coding agent has completed its work and exited with file changes, **When** control returns to `sl revise`, **Then** I see a summary of changed files and am offered to commit and push before proceeding to comment resolution.
4. **Given** the coding agent has exited with no file changes on disk (agent may have committed itself, or no changes were needed), **When** control returns to `sl revise`, **Then** I skip the commit/push step and proceed directly to comment resolution.

---

### User Story 6 - Commit, Push, and Comment Resolution After Agent Work (Priority: P6)

As a developer whose coding agent has finished making changes, I want help committing and pushing those changes, and then selectively marking comments as resolved, so the full feedback loop is closed in one flow.

**Why this priority**: Closing the feedback loop — commit, push, resolve — is important for team collaboration. Offering to commit/push after the agent exits removes friction and ensures the revised artifacts reach the remote before comments are resolved.

**Independent Test**: Can be tested by completing an agent session with file changes, verifying the commit/push prompt appears, then choosing to resolve some comments and confirming the correct API calls are made.

**Acceptance Scenarios**:

1. **Given** the agent has exited and files have been modified, **When** control returns to `sl revise`, **Then** I see a summary of changed files (like `git status`) and am offered to commit and push.
2. **Given** I choose to commit, **When** the file list is shown, **Then** I see a multi-select of changed files to stage, can confirm or skip individual files, enter a commit message, and push to the current branch.
2a. **Given** I choose to skip committing entirely, **Then** no files are staged and the flow proceeds to resolution.
3. **Given** I choose to skip committing, **When** I proceed to the resolve step, **Then** I see a warning that resolving comments without pushing changes may lead to inconsistencies, with options to proceed or defer.
4. **Given** the commit/push step is complete (or skipped), **When** I reach the resolve step, **Then** I see a multi-select list of the processed comments and can choose which ones to mark as resolved.
5. **Given** I select 2 of 3 comments to resolve, **When** I confirm, **Then** only those 2 are marked as resolved in the remote system.
6. **Given** I choose to defer resolution entirely, **When** the command exits, **Then** I see a reminder: "Unresolved comments remain. Re-run `sl revise` after pushing to resolve them."

---

### User Story 7 - Branch Checkout with Stash Handling (Priority: P7)

As a developer who selects a branch different from my current one, I want the command to safely switch me to that branch — detecting uncommitted changes and offering to stash them first — so I can revise artifacts without losing work in progress.

**Why this priority**: This is a convenience and safety feature. The core workflow (US1-US6) works when the user is already on the correct branch. This story handles the branch-switching edge case gracefully.

**Independent Test**: Can be tested by having uncommitted changes on `main`, running `sl revise` and selecting a different branch, verifying the stash prompt appears, and confirming the checkout succeeds after stashing.

**Acceptance Scenarios**:

1. **Given** I am on `main` with no uncommitted changes and select branch `009-feature-name`, **When** I confirm the selection, **Then** the system checks out `009-feature-name` and proceeds to fetch comments.
2. **Given** I am on `main` with uncommitted changes and select a different branch, **When** the checkout is needed, **Then** I see a warning listing the uncommitted changes and am offered to stash, abort, or continue (risking conflicts).
3. **Given** I choose to stash, **When** the checkout completes and the revise session ends, **Then** I see a reminder: "You have stashed changes. Run `git stash pop` to restore them."
4. **Given** I select a branch that only exists on remote, **When** the checkout is needed, **Then** the system fetches the remote branch and creates a local tracking branch.

---

### User Story 8 - Automation Mode (Priority: P8)

As a developer or CI pipeline, I want to run `sl revise --auto <fixture.json>` with a pre-prepared fixture file that specifies which comments to process and optional guidance per comment, so the revise flow can run non-interactively for testing or batch processing.

**Why this priority**: Enables CI integration and repeatable testing. Not required for the core interactive workflow but makes the tool scriptable and testable.

**Independent Test**: Can be tested by creating a fixture JSON file with comment IDs and guidance strings, running `sl revise --auto fixture.json`, and verifying the correct comments are processed with the specified guidance, the prompt is generated, and the agent is launched without interactive prompts.

**Acceptance Scenarios**:

1. **Given** I have a fixture file mapping comments to guidance strings, **When** I run `sl revise --auto fixture.json`, **Then** the command skips all interactive prompts, processes only the specified comments with their guidance, generates the prompt, and prints it to stdout. No agent is launched and no comments are resolved.
2. **Given** the fixture file references a comment that doesn't exist or is already resolved, **When** the command processes it, **Then** it logs a warning and skips that entry.
3. **Given** I want to validate prompt generation in a test suite, **When** I run `sl revise --auto fixture.json`, **Then** the stdout output is deterministic for a given set of comments and guidance, making it suitable for snapshot testing.

---

### User Story 9 - Summary Flag for Non-Interactive Comment Listing (Priority: P9)

As a coding agent (e.g., Claude Code running `/specledger.clarify`), I want to run `sl revise --summary` to get a compact, non-interactive listing of unresolved comments with file paths, line ranges, and comment text — so the agent can incorporate reviewer feedback into its analysis without launching another interactive session.

**Why this priority**: Enables integration with the `/specledger.clarify` workflow. When a user runs `/specledger.clarify` inside Claude Code, the clarify prompt instructs the agent to call `sl revise --summary` to fetch reviewer comments alongside its own spec ambiguity analysis. This makes reviewer feedback available to the clarify workflow without requiring a separate interactive revise session.

**Independent Test**: Can be tested by running `sl revise --summary` and verifying the output is a compact, machine-readable listing of unresolved comments to stdout.

**Acceptance Scenarios**:

1. **Given** I run `sl revise --summary`, **When** unresolved comments exist, **Then** I see a compact listing to stdout with one entry per comment showing: file path, line range (if available), selected text (truncated), and comment content (truncated).
2. **Given** I run `sl revise --summary` and authentication fails, **When** the API call fails, **Then** the command exits silently with a non-zero exit code (no error output to stdout), so the calling agent can gracefully fall back to local-only analysis.
3. **Given** the `/specledger.clarify` prompt includes instructions to call `sl revise --summary`, **When** the agent executes it, **Then** the agent can present the comments to the user via AskUserQuestion multi-select to choose which reviewer feedback to incorporate into the clarification session.

---

### Edge Cases

- What happens when the API returns an authentication error mid-session (token expired)? All API calls should auto-retry on 401/PGRST303: call `auth.GetValidAccessToken()` (which auto-refreshes the token) and retry the request once. This is critical because the agent session can last a long time, making token expiry likely before the resolve step.
- What happens when a comment references a file that no longer exists locally? The comment should still be displayed but flagged as "File not found locally" and the user can choose to skip or process it without file context.
- What happens when multiple users are resolving comments concurrently? Resolving an already-resolved comment should be a no-op (idempotent).
- What happens when the user's editor command fails or is not found? Fall back to displaying the prompt in the terminal and prompting for a filename to write it to.
- What happens when network connectivity is lost during comment fetching? Display a clear error with guidance to retry, and ensure no partial state is persisted.
- What happens when a comment's `selected_text` is no longer present in the file (artifact was modified)? Display the comment with a note: "Original selected text not found in current file version" and show surrounding context if `line` number is available.
- What happens when the user selects a branch that has been deleted remotely? Display an error and re-show the branch list.
- What happens when `git stash` fails (e.g., conflicts)? Abort the branch switch and inform the user to resolve manually.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST provide `sl revise` as a subcommand in the CLI tool.
- **FR-002**: When on a feature branch, system MUST prefill the current branch and allow the user to confirm or choose a different branch.
- **FR-003**: When not on a feature branch (e.g., `main`), system MUST present a list of branches with unresolved artifact comments for the user to select from.
- **FR-003a**: System MUST accept an optional branch name as a positional argument (e.g., `sl revise 009-feature-name`) to skip the selection prompt.
- **FR-003b**: When the selected branch differs from the current branch, system MUST detect uncommitted changes and offer to stash before switching.
- **FR-003c**: System MUST support checking out remote-only branches by fetching and creating a local tracking branch.
- **FR-004**: System MUST verify authentication status before making any Supabase API calls, and display a clear error with remediation steps if not authenticated.
- **FR-005**: System MUST fetch all unresolved artifact comments for the determined spec key.
- **FR-006**: System MUST group fetched comments by artifact (file path) and display only artifacts with unresolved comments in a multi-select TUI prompt.
- **FR-007**: System MUST show the count of unresolved comments next to each artifact in the selection prompt.
- **FR-008**: System MUST exit immediately with a clear message when no unresolved comments exist for the spec.
- **FR-009**: System MUST provide an interactive processing loop with three actions per comment: process, skip, and quit.
- **FR-010**: System MUST display comment details (file path, line, selected text, content, author, date) when presenting each comment in the loop.
- **FR-011**: System MUST allow the user to optionally provide free-text guidance when choosing to "process" a comment.
- **FR-012**: System MUST generate a single combined revision prompt from an embedded template, merging all processed comments (across all selected artifacts) with their context, selected text, file references, and user guidance.
- **FR-013**: System MUST launch the user's configured editor (`$EDITOR` or `$VISUAL`, defaulting to `vi`) to allow modification of the generated prompt.
- **FR-014**: System MUST display the final prompt text after the editor closes and allow the user to confirm, re-edit, or cancel.
- **FR-015**: System MUST provide an estimated token count using a simple character-based heuristic (~3.5 characters per token, Anthropic's recommended local estimate) with warnings for prompts under 100 tokens or over 8000 tokens.
- **FR-016**: System MUST launch the configured coding agent with the finalized prompt content when the user chooses to proceed.
- **FR-017**: System MUST present a multi-select prompt after agent exit allowing the user to choose which processed comments to mark as resolved.
- **FR-018**: System MUST resolve artifact comments by marking them as resolved in the remote system.
- **FR-019**: System MUST detect uncommitted Git changes before the resolve step and warn the user about potential inconsistencies.
- **FR-020**: System MUST handle API errors (401, 403, 404, expired tokens) with clear error messages and remediation guidance.
- **FR-022**: All API calls MUST auto-retry on 401/PGRST303 errors by refreshing the access token and retrying the request once, to handle token expiry during long agent sessions.
- **FR-023**: System MUST support `--auto <fixture.json>` flag for non-interactive automation. In auto mode: process only the comments specified in the fixture, generate the prompt, and print it to stdout. No agent is launched and no comments are resolved. This enables snapshot testing of prompt generation.
- **FR-024**: System MUST support `--dry-run` flag (for interactive mode) that outputs the generated prompt to a file instead of launching the agent, and does not resolve comments.
- **FR-025**: System MUST support `--summary` flag that outputs a compact, non-interactive listing of unresolved comments to stdout and exits. On auth failure, exit silently with non-zero exit code (no error to stdout).
- **FR-021**: The session MUST exit after one complete cycle (fetch → select → process → prompt → agent → commit/push → resolve). Users re-run `sl revise` for remaining unresolved comments. No `--resolve` flag is needed; the full flow handles resolution with the ability to skip/quit the processing loop to reach resolution quickly.

### Key Entities

- **Artifact Comment**: A piece of review feedback attached to a specific artifact file path (and optionally a line and selected text) within a spec. Has a resolution status. Distinct from issue comments, which are beads task-tracking comments and are not in scope for the revise workflow.
- **Artifact**: A file within a spec folder (e.g., `spec.md`, `plan.md`, `tasks.md`) that can have comments attached to it.
- **Revision Prompt**: A generated text document combining comment context, file references, and user guidance, intended as input for a coding agent.
- **Coding Agent**: An external tool (e.g., Claude Code) configured by the user to process revision prompts and make changes to files.
- **Automation Fixture**: A JSON file specifying which comments to process and optional guidance per comment, enabling non-interactive/CI usage of `sl revise`.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can fetch and view all comments for a specification in under 5 seconds on a standard internet connection.
- **SC-002**: *(Dropped)* ~~Full flow under 10 minutes~~ — not meaningful to measure since interactive steps (comment selection, guidance, agent loop) are user-paced. Performance focus is on SC-001 (fetch speed) and prompt generation (Go template rendering, negligible).
- **SC-003**: 90% of users can successfully run `sl revise` on their first attempt without consulting documentation, when authenticated and on a feature branch.
- **SC-004**: Zero data loss — no comments are accidentally resolved without explicit user confirmation.
- **SC-005**: The command correctly auto-detects the spec key from the branch name in 100% of cases where the branch follows the `###-name` pattern.
- **SC-006**: Token count estimates are within 20% accuracy of the actual token count when processed by the coding agent.

### Previous work

- **specledger.revise.md**: Existing Claude Code slash command (`.claude/commands/specledger.revise.md`) that performs comment fetching and processing via shell commands and CURL. This feature replaces it with a native Go CLI command.
- **008-cli-auth**: Authentication system (`sl auth login`, `sl auth status`, `sl auth token`) that this feature depends on for Supabase API access.
- **009-add-login-and-comment-commands**: Prior work on login and comment-related CLI commands.

## Clarifications

### Session 2026-02-20

- Q: Should the system generate one combined prompt or one per artifact? → A: One combined prompt for all processed comments (single agent session). Token warnings (FR-015) help the user gauge prompt size.
- Q: Should `sl revise --resolve` exist as a standalone mode? → A: No. There is no `--resolve` flag. Users always go through the full `sl revise` flow. To reach resolution quickly, skip/quit the processing loop. Comment IDs are never exposed to the user.
- Q: Should the session loop back for more comments after one agent run? → A: No. Exit after one cycle. Users re-run `sl revise` for remaining comments, ensuring a fresh fetch of current state each time.

## Dependencies & Assumptions

### Dependencies

- Authenticated Supabase session via `sl auth login` (feature 008-cli-auth)
- Remote artifact comment storage with resolution tracking (currently Supabase)
- User's shell must have `$EDITOR` or `$VISUAL` set (or `vi` available as fallback)

### Assumptions

- The spec key used for querying artifact comments is derived from the Git branch name (e.g., branch `136-revise-comments` → spec key `136-revise-comments`)
- Comments are only posted through the SpecLedger web UI or CLI, so their format is predictable
- The coding agent accepts prompt content via stdin or a file path argument
- Token count estimation uses a simple heuristic (~3.5 characters per token, per Anthropic's recommended local estimate) as a rough guide, not a precise tokenizer
- The user configures their preferred coding agent via `specledger.yaml` (under `agent.command`) or the `SPECLEDGER_AGENT` environment variable. `specledger.yaml` is the canonical config source; the env var overrides it.
- Network connectivity is available for Supabase API calls; offline mode is not supported for this command
