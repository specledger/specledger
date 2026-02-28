# Feature Specification: SDD Workflow Streamline

**Feature Branch**: `598-sdd-workflow-streamline`
**Created**: 2026-02-27
**Status**: Draft
**Input**: User description: "we want to streamline the workflow of sdd using specledger, and add/remove/modify ai skills/commands/ or sl cli command"

## Overview

Audit and consolidate the SDD (Specification-Driven Development) workflow components across four layers:

| Layer | Name | Runtime | Purpose |
|-------|------|---------|---------|
| 0 | Hooks | Invisible, event-driven | Auto-capture sessions on commit |
| 1 | `sl` CLI | Go binary, no AI needed | Data operations, CRUD, standalone tooling |
| 2 | AI Commands | Agent shell prompts | AI workflow orchestration (specify→implement) |
| 3 | Skills | Passive context injection | Domain knowledge, progressively loaded |

**Cross-layer interactions** (convenience patterns, not additional layers):
- L1→L0: `sl auth hook --install` configures hooks
- L1→L2: `sl revise` generates a prompt and launches an agent session (launcher pattern)
- L2→L1: AI commands call CLI tools (e.g., `/specledger.tasks` calls `sl issue create`)

Goal: Reduce redundancy, clarify layer responsibilities, streamline the developer experience, add new capabilities, and establish a CLI development constitution for future `sl` binary work.

**Core workflow is immutable**: specify→clarify→plan→tasks→implement. Playbooks customize content (skill bundles), not workflow shape.

**Future Work** (out of scope):
- **TUI tool**: Separate binary for human-focused interactive flows (`sl init` wizard, `sl revise` review flow). The `sl` CLI stays agent-focused (no PTY). Until TUI exists, interactive commands remain in `sl` as launcher shortcuts.
- **Playbook management**: Skill bundles (ML, backend, frontend, fullstack teams) designed in webapp/backend first, then flow back to CLI. `sl playbook` CLI frozen until then.
- **`/simplify` overlap**: Claude Code's built-in `/simplify` command handles PR-level code quality. Our `/specledger.audit` remains for codebase-level reconnaissance (onboarding, module graphs, tech stack discovery). These are complementary, not overlapping. Re-evaluate if `/simplify` expands scope.

## Clarifications

### Session 2026-02-27

- Q: How should `sl update` distinguish built-in (updatable) files from custom (preserved) files? → A: Filename matching (compare against list of embedded template names)
- Q: Should spike and checkpoint be implemented as AI commands or CLI commands? → A: AI commands (markdown files in `.opencode/commands/`)

### Session 2026-02-28 (cross-team review)

- Q: Should `sl update` be a separate command? → A: No — fold into `sl doctor --template` (template management is repo health, not a separate concern). See decision log D3.
- Q: Should analyze and audit be merged into `verify`? → A: No — keep separate. `analyze` (SpecKit brand) checks spec artifact consistency; `audit` scans the codebase. Different inputs, different stages. See decision log D4/D18.
- Q: How should non-conforming branch names resolve to spec folders? → A: Enhance `ContextDetector` with a 4-step fallback chain (regex → yaml alias → git heuristic → interactive prompt). Removes need for `adopt` command entirely. See decision log D9.
- Q: Where does review comment management belong? → A: New `sl comment` CLI subcommand (data CRUD). `clarify` AI command absorbs revise's spec-refinement logic. `sl revise` stays as CLI launcher shortcut until TUI. See decision log D2/D4.

## New Capabilities to Add

1. **Spike** (AI command) - Exploratory research for time-boxed investigations. Output: `specledger/<spec>/research/yyyy-mm-dd-<name>.md`
2. **Checkpoint** (AI command) - Implementation verification + session log with deviation tracking. Output: `specledger/<spec>/sessions/yyyy-mm-dd-<name>.md`
3. **`sl comment`** (CLI) - Review comment CRUD (list/show/reply/resolve/pull/push). Enables agents to manage comments with granular detail.
4. **`sl doctor --template`** enhancement - Update AI skills/commands to latest embedded templates (replaces proposed `sl update`)
5. **Context detection fallback** - Enhanced `ContextDetector` with yaml alias lookup + git heuristic (replaces `adopt` command)

## Current State Inventory

### Skills (2 → 3)
| Skill | Purpose | Status |
|-------|---------|--------|
| `sl-issue-tracking` | Teaches agent when/how to use `sl issue` | Existing |
| `specledger-deps` | Teaches agent when/how to use `sl deps` | Existing |
| `sl-comment` | Teaches agent when/how to use `sl comment` | **New** (D4) |

Skills are lean, isolated, and progressively loaded. Each CLI domain gets its own skill. AI commands reference CLI tools briefly, which triggers the relevant skill to load.

### AI Commands (11 → 12)
| Command | Purpose | Change | Stage |
|---------|---------|--------|-------|
| `specify` | Create feature spec | Unchanged | Core pipeline |
| `clarify` | Clarify spec ambiguities + process review comments | **Absorbs** revise functionality (D4) | Core pipeline |
| `plan` | Create implementation plan | Unchanged | Core pipeline |
| `tasks` | Generate tasks | Unchanged | Core pipeline |
| `implement` | Implementation guidance | **Absorbs** resume (D4) | Core pipeline |
| `spike` | Exploratory research | **New** (D13) | Escape hatch (any stage) |
| `checkpoint` | Implementation verification + session log | **New** (D14) | During implement |
| `analyze` | Cross-artifact consistency check | **Reverted** from `verify` — keep SpecKit name (D4) | Quality validation |
| `audit` | Codebase reconnaissance | **Reverted** from `verify` — separate from analyze (D18) | Codebase analysis |
| `checklist` | Custom per-feature quality gates | **Kept** as optional standalone (D12) | Optional |
| `onboard` | Project onboarding | **Absorbs** help (D4) | Setup |
| `constitution` | Project constitution | Unchanged | Setup |

**Removed AI commands** (6):
- `resume` → duplicate of implement
- `help` → absorbed by onboard
- `adopt` → replaced by context detection fallback chain (D9)
- `add-deps` / `remove-deps` → agent calls `sl deps` CLI directly
- `revise` → absorbed by clarify; `sl revise` stays as CLI launcher

### CLI Commands (9 → 10)
| Command | Purpose | Pattern (D16) | Future: TUI? |
|---------|---------|---------------|--------------|
| `sl init` | Initialize project (new or existing) | Environment + Launcher (D7) | **Yes** |
| `sl deps` | Manage dependencies (add/remove/graph) | Data CRUD | No |
| `sl doctor` | System diagnostics + template updates | Environment + Template mgmt (D3) | No |
| `sl playbook` | List playbooks | Environment | Frozen until webapp (D17) |
| `sl auth` | Authentication | Environment | No |
| `sl session` | Session capture | Hook trigger | No |
| `sl issue` | Issue tracking | Data CRUD | No |
| `sl comment` | Review comment management | Data CRUD | **New** (D4) |
| `sl revise` | Launch agent with comment context | Launcher | **Yes** |
| `sl mockup` | Launch agent with design system context | Launcher | **Yes** |

> **Note**: TUI migration is out of scope for this spec. Commands marked for TUI will be addressed when the separate TUI tool is created. Until then, interactive flows stay in `sl` as launcher shortcuts.

### Hooks (Layer 0)
| Hook | Trigger | CLI Command |
|------|---------|-------------|
| `PostToolUse` (Bash matcher) | After `git commit` | `sl session capture` |

## Identified Overlaps (Resolved)

1. **Dependency Management**: Agent calls `sl deps` CLI directly. No AI command wrapper needed. `sl graph` folded into `sl deps graph` (D6).
2. **Analysis/Audit**: Kept separate — `analyze` = spec artifact consistency, `audit` = codebase reconnaissance. Different inputs, different stages (D4/D18). Claude Code `/simplify` covers PR-level code quality (complementary).
3. **Issue Tracking**: `sl-issue-tracking` skill teaches agent when/how to use `sl issue` CLI. Complementary, not duplicate (D5).
4. **Comment Management**: New `sl comment` CLI + `sl-comment` skill. Same pattern as issues (D4/D5).
5. **Template Updates**: `sl doctor --template` handles template lifecycle. Owns `specledger.` prefix, detects stale/deprecated commands (D3). No separate `sl update` needed.
6. **Branch Detection**: `ContextDetector` enhanced with fallback chain. No `adopt` command needed (D9).

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Inventory and Classify Workflow Components by Layer (Priority: P1)

A developer or maintainer needs a clear inventory of all workflow components **organized by layer** (Hooks, CLI, Commands, Skills) to understand responsibilities, overlaps, and established CLI patterns. (D1, D16)

> **Note**: This is a workflow review exercise, not the `/specledger.audit` AI command (which scans source code). Naming kept distinct to avoid confusion.

**Why this priority**: Cannot streamline without understanding current state. The 4-layer model and CLI constitution provide the framework for all subsequent consolidation decisions.

**Independent Test**: Can be fully tested by reviewing documentation that lists all components with their layer assignment, pattern classification, and cross-layer interactions.

**Acceptance Scenarios**:

1. **Given** the current codebase, **When** audit is complete, **Then** all components are documented with their **layer assignment** (Hook/CLI/Command/Skill) and **pattern classification** (Data CRUD, Launcher, Hook trigger, Environment, Template mgmt)
2. **Given** the component inventory, **When** overlaps are identified, **Then** each overlap is resolved with documented rationale referencing the decision log
3. **Given** the audit findings, **When** a CLI development constitution is drafted, **Then** it defines the 5 established patterns, review gates, and constraints (offline capability, cross-platform, layer boundaries)
4. **Given** the constitution, **When** each existing `sl` subcommand is classified, **Then** every command maps to at least one established pattern

---

### User Story 2 - Consolidate Dependency Commands (Priority: P1)

A developer wants a single AI command for dependency management that orchestrates the CLI, plus a consolidated `sl deps` CLI with graph functionality.

**Why this priority**: Dependency management has the most obvious overlap and user confusion.

**Independent Test**: Can be fully tested by verifying `deps` AI command orchestrates `sl deps` CLI and `sl deps graph` subcommand exists.

**Acceptance Scenarios**:

1. **Given** consolidated `deps` AI command, **When** user runs it, **Then** it orchestrates `sl deps add`/`remove`/`graph` CLI commands
2. **Given** the consolidated CLI, **When** user needs to view dependency graph, **Then** `sl deps graph` provides the visualization
3. **Given** the consolidated interface, **When** user needs to manage dependencies, **Then** the workflow is clear and documented

---

### User Story 3 - Clarify Analyze vs Audit Responsibilities (Priority: P2)

A developer wants clear, distinct purposes for `analyze` and `audit` commands so they know which to use and when. (D4, D18)

**Why this priority**: These commands serve different stages and have different inputs. Merging them (as previously proposed as `verify`) conflates codebase understanding with spec document validation.

**Independent Test**: Can be fully tested by running each command separately and verifying they have distinct inputs, outputs, and help text.

**Acceptance Scenarios**:

1. **Given** the `analyze` command, **When** a developer runs it after task generation, **Then** it reads spec.md, plan.md, and tasks.md and reports consistency issues (duplication, ambiguity, coverage gaps, constitution alignment)
2. **Given** the `audit` command, **When** a developer runs it on an unfamiliar codebase, **Then** it scans source code for tech stack, modules, dependency graphs and produces a JSON cache
3. **Given** both commands, **When** a developer reads their help text, **Then** the distinction is clear: `analyze` = "validate your design documents", `audit` = "understand this codebase"
4. **Given** Claude Code's built-in `/simplify`, **When** a developer needs PR-level code quality checks, **Then** they use `/simplify` (complementary, not overlapping with analyze or audit)

---

### User Story 4 - Update Skills to Complement CLI with Progressive Loading (Priority: P2)

A developer wants AI skills that enhance CLI functionality rather than duplicate it. Skills are lean, isolated, and progressively loaded — they teach the agent *when* and *how* to use CLI tools, triggered by contextual references in AI commands. (D5)

**Why this priority**: Improves developer experience by making skills additive. The progressive loading pattern ensures skills don't bloat the agent context.

**Independent Test**: Can be fully tested by reviewing skill content to ensure it references CLI commands rather than duplicating their logic, and by verifying a new `sl-comment` skill is created.

**Acceptance Scenarios**:

1. **Given** `sl-issue-tracking` skill and `sl issue` CLI, **When** skill is updated, **Then** skill provides AI context about issue workflow without duplicating CLI functionality
2. **Given** `specledger-deps` skill and `sl deps` CLI, **When** skill is updated, **Then** skill provides AI context about dependency workflow without duplicating CLI functionality
3. **Given** the new `sl comment` CLI, **When** a `sl-comment` skill is created, **Then** it follows the same pattern as `sl-issue-tracking` — teaches agent when/how to use `sl comment list/show/reply/resolve`
4. **Given** the `/specledger.clarify` AI command references review comments, **When** the agent encounters this reference, **Then** the `sl-comment` skill is progressively loaded to provide usage details

---

### User Story 5 - Document Consolidated Workflow (Priority: P3)

A developer wants clear documentation of the streamlined SDD workflow.

**Why this priority**: Documentation is important but follows the actual consolidation work.

**Independent Test**: Can be fully tested by reviewing updated documentation.

**Acceptance Scenarios**:

1. **Given** consolidated workflow, **When** documentation is updated, **Then** all skills, commands, and CLI commands are documented with their purposes
2. **Given** the documentation, **When** a new developer reads it, **Then** the workflow is clear and unambiguous

---

### User Story 6 - Create Spike for Exploratory Research (Priority: P1)

A developer needs to perform time-boxed exploratory research (spike) on a technical question or approach before committing to implementation. The spike results are captured in a structured format for future reference.

**Why this priority**: Spikes are essential for reducing implementation risk by validating approaches early.

**Independent Test**: Can be fully tested by running the spike command and verifying a research document is created in `[spec]/research/` with date-prefixed filename.

**Acceptance Scenarios**:

1. **Given** an active feature spec, **When** user runs spike command with a topic, **Then** a research file is created at `specledger/[spec-id]/research/yyyy-mm-dd-[spike-name].md`
2. **Given** the spike command, **When** spike completes, **Then** the file contains research question, approach, findings, and recommendations
3. **Given** multiple spikes on same spec, **When** listing research, **Then** all spikes are visible in chronological order

---

### User Story 7 - Checkpoint Implementation Progress + Session Log (Priority: P1)

A developer wants to verify that current implementation aligns with the spec and produce a structured session log that captures where and why the implementation diverges from the plan. This enables reviewers to understand the implementation without reading full session transcripts. (D14)

**Why this priority**: Checkpoints provide safety and visibility into implementation progress. The session log solves the implementation tracking gap — without it, reviewers can't tell what changed from the original plan without reading raw conversation transcripts.

**Independent Test**: Can be fully tested by running checkpoint command and verifying a session file is created with both implementation verification and structured deviation tracking.

**Acceptance Scenarios**:

1. **Given** an active feature with implementation in progress, **When** user runs checkpoint command, **Then** a session file is created at `specledger/[spec-id]/sessions/yyyy-mm-dd-[session-name].md`
2. **Given** the checkpoint command, **When** checkpoint runs, **Then** the file contains: spec compliance status, changed files list, implementation notes, and any deviations found
3. **Given** git changes since last checkpoint, **When** checkpoint runs, **Then** file changes are summarized with diff statistics
4. **Given** multiple checkpoints on same spec, **When** listing sessions, **Then** all checkpoints are visible in chronological order
5. **Given** the session log section, **When** checkpoint runs, **Then** the agent looks back and documents: tasks worked on (by issue ID), what was planned vs what was actually done, divergences with justifications, key decisions made during the session, unfinished work and why, and impact on downstream tasks
6. **Given** a PR with checkpoint logs, **When** a reviewer reads them, **Then** they can understand why the implementation looks different from the original plan without reading raw session transcripts

**Relationship to `sl session capture`** (Layer 0): Session capture records the raw conversation transcript automatically via hooks. Checkpoint produces a **curated, human-readable summary** with structured deviation tracking. They complement each other — session capture is forensic evidence, checkpoint is the executive summary.

---

### User Story 8 - Template Lifecycle via `sl doctor --template` (Priority: P2)

A developer wants to update their project's AI skills and commands to the latest embedded template versions when a new version of SpecLedger is released, without losing project-specific customizations. Template management is part of repo health (`sl doctor`), not a separate command. (D3)

**Why this priority**: Keeps projects up-to-date with workflow improvements, but less urgent than core workflow commands.

**Independent Test**: Can be fully tested by running `sl doctor --template` and verifying skills/commands are updated while custom files are preserved, and stale commands are detected.

**Acceptance Scenarios**:

1. **Given** a project with existing skills/commands, **When** user runs `sl doctor --template`, **Then** built-in skills and commands are updated to latest embedded versions
2. **Given** a project with custom (non-built-in) skills/commands, **When** user runs `sl doctor --template`, **Then** custom files are preserved unchanged
3. **Given** a built-in skill/command that was locally modified, **When** user runs `sl doctor --template`, **Then** user is prompted to keep local or use updated template
4. **Given** `sl doctor --template --dry-run`, **When** run, **Then** shows what would be updated without making changes
5. **Given** a project with `specledger.` prefixed commands that are no longer in the current playbook, **When** user runs `sl doctor --template`, **Then** stale commands are detected and user is prompted to remove them (with option to keep if experimental)
6. **Given** a `specledger.` prefixed command with a non-standard prefix (e.g., `specledger-beta.`), **When** user runs `sl doctor --template`, **Then** it is treated as user-owned and not flagged for removal

---

### User Story 9 - Update CLI README After Streamlining (Priority: P2)

After the CLI command streamlining is complete and stable, the project README.md must be updated to reflect the new simplified command structure and workflow. The README should highlight the streamlined workflow, show clear examples of each command, and guide new users through the improved developer experience.

**Why this priority**: Documentation is essential for adoption. Users discovering SpecLedger via the README need accurate, up-to-date information about the simplified commands. P2 because the commands must be implemented and stable first, but documentation should follow closely.

**Independent Test**: Can be tested by reviewing the updated README.md to verify it documents all streamlined commands with accurate usage examples.

**Acceptance Scenarios**:

1. **Given** the CLI streamlining feature is complete, **When** the README is updated, **Then** it documents all remaining commands with usage examples.
2. **Given** a new user reading the README, **When** they follow the quickstart instructions, **Then** the commands shown match the actual CLI behavior.
3. **Given** the README is updated, **When** a user searches for workflow instructions, **Then** the README provides a clear, step-by-step guide.

---

### User Story 10 - Add CHANGELOG.md for AI Commands/Skills Templates (Priority: P3)

When embedded AI commands (`.opencode/commands/`) and skills (`.opencode/skills/`) templates are updated, a CHANGELOG.md should be included in the embedded templates to track these changes. This allows projects initialized with SpecLedger to understand what changes have been made to their command/skill files over time and decide whether to adopt updates.

**Why this priority**: Provides transparency for template evolution but is not critical to the core workflow experience. P3 because it's a nice-to-have for maintainability but doesn't block user adoption.

**Independent Test**: Can be tested by verifying a CHANGELOG.md exists in the embedded templates and documents changes to command/skill files with version references.

**Acceptance Scenarios**:

1. **Given** the embedded templates directory, **When** a CHANGELOG.md is added, **Then** it lists all AI command and skill files with their initial versions.
2. **Given** a command or skill template is modified, **When** the change is released, **Then** the CHANGELOG.md is updated with the change description, affected files, and version.
3. **Given** a user running `sl bootstrap` or `sl init`, **When** templates are copied to their project, **Then** the CHANGELOG.md is included so they can track template changes.
4. **Given** a user reviewing their project's CHANGELOG.md, **When** they want to update their commands/skills, **Then** the CHANGELOG provides enough information to decide which changes to adopt.

---

### User Story 11 - Review Comment Management via `sl comment` (Priority: P1)

A developer working with an AI agent needs fine-grained control over review comments — listing, replying to, and resolving them with detailed context. Currently, comment management is done via raw curl calls to Supabase (fragile, auth-dependent) or bulk-resolved without detail by `sl revise` post-session. (D2, D4)

**Why this priority**: Without `sl comment`, agents can't resolve comments with meaningful replies. This blocks the `clarify` command from absorbing revise's comment-handling functionality and prevents granular comment workflows.

**Independent Test**: Can be fully tested by creating review comments in the app, then using `sl comment` CLI to list, reply, and resolve them.

**Acceptance Scenarios**:

1. **Given** unresolved review comments on a spec, **When** user runs `sl comment list --status open`, **Then** all unresolved comments are displayed with file path, selected text, author, and content
2. **Given** a specific comment ID, **When** user runs `sl comment show <id>`, **Then** full comment details including thread context (parent/replies) are displayed
3. **Given** a comment ID and reply text, **When** user runs `sl comment reply <id> --content "Addressed in commit abc123"`, **Then** a reply is posted and linked to the parent comment
4. **Given** a comment ID, **When** user runs `sl comment resolve <id> --reason "Fixed in PR #42"`, **Then** the comment is marked resolved with the reason recorded
5. **Given** an AI agent running `/specledger.clarify`, **When** the agent processes review comments, **Then** it uses `sl comment reply` and `sl comment resolve` for each comment individually (not bulk resolution)
6. **Given** the `sl-comment` skill is installed, **When** the agent encounters a reference to review comments in any AI command, **Then** the skill provides usage patterns for `sl comment` CLI

**CLI pattern**: Data CRUD (D16). Works offline against local cache; `sl comment pull/push` for remote sync.

---

### User Story 12 - Context Detection Fallback Chain (Priority: P2)

A developer working on a branch that doesn't follow SpecLedger's naming convention (e.g., created by GitHub UI, Jira, or Linear) needs `sl` commands to automatically detect which spec the branch belongs to, without running a separate `adopt` command. (D9)

**Why this priority**: The current `ContextDetector` is a single regex (`^\d{3,}-[a-z0-9-]+$`). Any non-conforming branch fails with a hard error, forcing users to use `--spec` flags everywhere. This is the root cause of why `adopt` was created.

**Independent Test**: Can be tested by creating a branch with a non-standard name, editing files in a known spec directory, and verifying auto-detection works.

**Acceptance Scenarios**:

1. **Given** a branch named `598-sdd-workflow-streamline`, **When** `ContextDetector` runs, **Then** it resolves via regex match (existing behavior, step 1)
2. **Given** a branch named `feature/fix-login` with an alias in `specledger.yaml`, **When** `ContextDetector` runs, **Then** it resolves via yaml alias lookup (step 2)
3. **Given** a branch named `johns-auth-work` with commits touching `specledger/042-auth-improvements/`, **When** `ContextDetector` runs, **Then** it resolves via git heuristic — diffing branch against base, finding exactly one `specledger/` directory modified (step 3)
4. **Given** a branch with no regex match, no alias, and no git heuristic result, **When** `ContextDetector` runs in interactive mode, **Then** it lists available spec directories and prompts the user to pick one, saving the alias to `specledger.yaml` (step 4)
5. **Given** a saved alias in `specledger.yaml`, **When** any future `sl` command runs on that branch, **Then** it auto-resolves without re-prompting
6. **Given** non-interactive mode (CI, `--spec` flag), **When** detection fails, **Then** the `--spec` flag overrides all detection steps

**Note**: Full technical details of the 4-step fallback chain are in the decision log (D9). Implementation should also evaluate `git ls-tree` as an alternative to `git diff --name-only` for the heuristic step.

---

### User Story 13 - Phase Out Bash Scripts (Priority: P2)

A developer on Windows (or any non-Unix platform) needs `sl` commands to work without depending on bash scripts. Currently, several AI commands call bash scripts in `.specledger/scripts/bash/` for deterministic operations (e.g., `create-new-feature.sh`). These don't work on Windows and are a maintenance burden. (D10)

**Why this priority**: Cross-platform support is a constraint in the CLI constitution (D16). Bash scripts break this constraint and create two codepaths for the same operation.

**Independent Test**: Can be tested by verifying all bash script functionality is available via `sl` CLI commands, and that `sl` works on Windows without bash.

**Acceptance Scenarios**:

1. **Given** the current bash scripts in `.specledger/scripts/bash/`, **When** an inventory is completed, **Then** each script's functionality is documented with its proposed `sl` CLI equivalent
2. **Given** `create-new-feature.sh`, **When** its functionality is absorbed into `sl specify` or `sl init`, **Then** the bash script is no longer called by any AI command
3. **Given** all bash scripts have CLI equivalents, **When** a developer runs `sl` on Windows, **Then** all commands work without bash dependency
4. **Given** an AI command template that previously called a bash script, **When** the template is updated, **Then** it calls the equivalent `sl` CLI command instead

---

### User Story 14 - Clarify Absorbs Revise Comment Handling (Priority: P2)

A developer wants a single AI command (`clarify`) that handles both spec ambiguity resolution and review comment processing. The `sl revise` CLI command stays as a launcher shortcut (pre-filters comments, launches agent) until TUI work begins. (D2, D4)

**Why this priority**: Two separate comment-handling workflows (revise slash command + clarify) creates confusion. Consolidating the AI reasoning into `clarify` while keeping `sl revise` as a CLI launcher creates a clean separation.

**Independent Test**: Can be tested by running `/specledger.clarify` with review comments present, and verifying it processes them using `sl comment` CLI.

**Acceptance Scenarios**:

1. **Given** unresolved review comments on a spec, **When** user runs `/specledger.clarify`, **Then** the command detects open comments (via `sl comment list --status open`) and offers to process them alongside spec ambiguity scanning
2. **Given** the clarify command is processing a review comment, **When** the agent decides how to address it, **Then** it uses `sl comment reply <id>` to post a detailed response and `sl comment resolve <id>` to mark it resolved
3. **Given** `sl revise` CLI is run from the terminal, **When** it pre-filters comments and launches the agent, **Then** the agent's behavior is defined by `/specledger.clarify` (not a separate `/specledger.revise` command)
4. **Given** the old `/specledger.revise` AI command template, **When** consolidation is complete, **Then** it is removed from the playbook templates (and detected as stale by `sl doctor --template` per D3)

**Note**: `sl revise` stays as a CLI launcher shortcut — it does pre-flight (branch checkout, stash handling) and context gathering (fetch comments), then hands off to the agent. The CLI does NOT interpret agent results post-session. Until TUI work begins, this launcher pattern is the primary entry point for comment-driven revision workflows.

---

### Edge Cases

- What happens to existing projects using removed commands? Document migration path.
- How to handle commands that are removed but still referenced in external docs? Add deprecation notices.
- What happens when spike research directory doesn't exist? Create it automatically.
- What happens when checkpoint finds no changes since last checkpoint? Still create session file noting "no changes".
- What happens when checkpoint detects spec violations? Flag violations prominently in session file.
- What happens when `sl doctor --template` finds no updates available? Display "Already up to date" message.
- What happens when `sl doctor --template` on a project never initialized? Error with suggestion to run `sl init` first.
- What happens when `sl comment list` has no comments? Display "No review comments found" and exit cleanly.
- What happens when context detection heuristic finds multiple spec directories modified? Prompt user to pick one and save alias.
- What happens when a bash script has platform-specific logic that can't be trivially ported? Document the gap and prioritize for the next release.

## Requirements *(mandatory)*

### Functional Requirements

**Layer Model & Constitution (US1)**
- **FR-001**: System MUST have documented inventory of all skills, commands, and CLI commands organized by layer (Hook/CLI/Command/Skill) with pattern classification (D1, D16)
- **FR-002**: A CLI development constitution MUST define the 5 established patterns (Data CRUD, Launcher, Hook trigger, Environment, Template mgmt) with review gates and constraints
- **FR-003**: All remaining commands MUST have clear, non-overlapping purposes with a single documented layer assignment

**Dependency Management (US2)**
- **FR-004**: Agent MUST call `sl deps` CLI directly for dependency operations. No AI command wrapper needed (D4)
- **FR-005**: `sl deps graph` MUST provide spec dependency visualization (absorbing `sl graph`) (D6)

**Analysis & Audit (US3)**
- **FR-006**: `analyze` and `audit` MUST remain separate commands with distinct inputs and outputs (D4, D18)
- **FR-007**: `analyze` MUST check spec artifacts (spec.md, plan.md, tasks.md) for consistency, coverage gaps, and constitution alignment
- **FR-008**: `audit` MUST scan source code for tech stack, modules, and dependency graphs

**Skills (US4)**
- **FR-009**: Skills MUST complement CLI functionality, not duplicate it (D5)
- **FR-010**: A `sl-comment` skill MUST be created following the same pattern as `sl-issue-tracking` (D5)
- **FR-011**: Skills MUST be progressively loaded — AI commands reference CLI tools briefly, triggering skill injection

**Spike (US6)**
- **FR-012**: System MUST provide spike AI command to create exploratory research documents in `specledger/[spec-id]/research/yyyy-mm-dd-[name].md`
- **FR-013**: Spike files MUST include research question, approach explored, findings, recommendations, and impact on spec/plan

**Checkpoint + Session Log (US7)**
- **FR-014**: System MUST provide checkpoint AI command to verify implementation against specs in `specledger/[spec-id]/sessions/yyyy-mm-dd-[name].md`
- **FR-015**: Checkpoint files MUST include spec compliance status, changed files summary, implementation notes, and deviations found
- **FR-016**: Checkpoint MUST detect and summarize git file changes since last checkpoint (or branch creation)
- **FR-017**: Checkpoint session log MUST capture: tasks worked on (by issue ID), planned vs actual, divergences with justifications, key decisions, unfinished work, and impact on downstream tasks (D14)
- **FR-018**: Both spike and checkpoint MUST auto-create target directories if they don't exist

**Template Lifecycle (US8)**
- **FR-019**: `sl doctor --template` MUST update built-in skills and commands to latest embedded template versions (D3)
- **FR-020**: `sl doctor --template` MUST preserve custom (non-built-in) skills and commands (detected by filename matching against embedded template list)
- **FR-021**: `sl doctor --template` MUST prompt for conflict resolution when built-in files were locally modified
- **FR-022**: `sl doctor --template --dry-run` MUST show pending updates without making changes
- **FR-023**: `sl doctor --template` MUST detect stale `specledger.` prefixed commands no longer in the playbook and prompt for removal (D3)

**Review Comment Management (US11)**
- **FR-024**: System MUST provide `sl comment` CLI with subcommands: list, show, reply, resolve (D4)
- **FR-025**: `sl comment` MUST support `--status open|resolved|all` filtering
- **FR-026**: `sl comment reply` MUST link replies to parent comments with detailed content
- **FR-027**: `sl comment resolve` MUST record a resolution reason

**Context Detection (US12)**
- **FR-028**: `ContextDetector` MUST implement a 4-step fallback chain: regex match → yaml alias → git heuristic → interactive prompt (D9)
- **FR-029**: Branch aliases MUST be stored in `specledger.yaml` and version-controlled
- **FR-030**: Non-interactive mode (`--spec` flag) MUST override all detection steps

**Bash Script Migration (US13)**
- **FR-031**: All bash scripts in `.specledger/scripts/bash/` MUST have documented `sl` CLI equivalents (D10)
- **FR-032**: Branch number generation (currently in bash) MUST prevent numeric prefix collisions when migrated to `sl` CLI ([#46](https://github.com/specledger/specledger/issues/46))

**Clarify + Revise (US14)**
- **FR-033**: `/specledger.clarify` MUST detect open review comments (via `sl comment list --status open`) and offer to process them alongside spec ambiguity scanning (D4)
- **FR-034**: Removed/deprecated AI commands MUST have migration documentation (D3)
- **FR-035**: Updated workflow MUST be documented in AGENTS.md or equivalent

### Key Entities

- **Layer**: One of four tiers in the tooling model — Hook (L0), CLI (L1), Command (L2), Skill (L3) (D1)
- **CLI Pattern**: One of 5 established patterns for `sl` subcommands — Data CRUD, Launcher, Hook trigger, Environment, Template mgmt (D16)
- **Skill**: AI context file that guides agent behavior for specific CLI domains. Lean, isolated, progressively loaded. (L3)
- **Command**: AI command file that orchestrates multi-step workflows using AI reasoning. (L2)
- **CLI Command**: Go binary command that performs deterministic data operations. (L1)
- **Launcher**: CLI command that gathers context and spawns an AI agent session (L1→L2 cross-layer). (D2)
- **Spike**: Time-boxed exploratory research document stored in `research/` folder with date prefix
- **Checkpoint**: Implementation verification + session log document stored in `sessions/` folder with date prefix (D14)
- **Branch Alias**: Mapping from non-conforming branch name to spec slug, stored in `specledger.yaml` (D9)
- **Built-in template**: Skills or commands embedded in the `sl` binary, owned by the `specledger.` prefix (D3)

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: AI command count reduced from 16 to 12 (remove 6, add 2: spike + checkpoint)
- **SC-002**: Zero overlapping purposes between layers — each component has exactly one layer assignment
- **SC-003**: Each `sl` CLI command maps to at least one established pattern (D16)
- **SC-004**: New developers can understand the 4-layer model and core workflow in under 5 minutes of reading
- **SC-005**: Spike command creates research document in under 2 minutes
- **SC-006**: Checkpoint command captures implementation state + session log in under 30 seconds
- **SC-007**: All spike and checkpoint files follow consistent `yyyy-mm-dd-[name].md` naming convention
- **SC-008**: `sl comment` provides granular comment management (no more bulk resolution)
- **SC-009**: `ContextDetector` resolves non-conforming branch names without user running a separate command
- **SC-010**: All `sl` commands work on macOS, Linux, and Windows without bash dependency

### Previous work

Existing infrastructure in agent shell skills (2 skills) and commands (16 commands across `.claude/commands/` and `.opencode/commands/`). CLI commands in `pkg/cli/commands/` (9 commands).

Existing session capture in `sl session capture` (Layer 0 hook) stores in `specledger/[spec-id]/sessions/` but lacks checkpoint verification and session log features.

Cross-team decision log with 20 decisions: [research/2026-02-28-command-consolidation-decisions.md](research/2026-02-28-command-consolidation-decisions.md)

### Dependencies & Assumptions

**Out of Scope** (future spec):
- TUI tool for human-focused interactive flows (`sl init` wizard, `sl revise` review) (D2)
- Playbook management in webapp — skill bundles for ML/backend/frontend/fullstack teams (D17)
- `sl skill` command for external skill discovery/install (D19)
- Mockup command improvements — pending feedback on current implementation (D15)

**Assumptions**:
- CLI commands (`sl deps`, `sl issue`, `sl comment`) are the source of truth for data operations (D2)
- AI commands orchestrate workflows using AI reasoning, calling CLI tools for data ops (D2)
- Skills provide domain knowledge, progressively loaded, not operational instructions (D5)
- Core workflow (specify→clarify→plan→tasks→implement) is immutable — playbooks customize content, not stages (D11)
- Spike and checkpoint are AI commands (agent shell command templates)
- Git is available for checkpoint change detection and context detection heuristic (D9)
- `specledger.` prefix in command filenames is owned by the playbook (D3)

**Dependencies**:
- Existing agent shell directory structure (`.claude/` or `.opencode/`)
- Current CLI command implementations in `pkg/cli/commands/`
- Supabase `review_comments` table for `sl comment` (columns: id, change_id, content, file_path, start_line, line, selected_text, is_resolved, author_id, parent_comment_id)
- Git for checkpoint change detection and context detection fallback chain
- Branch number collision prevention when migrating bash scripts ([#46](https://github.com/specledger/specledger/issues/46))
