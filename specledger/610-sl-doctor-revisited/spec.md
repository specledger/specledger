# Feature Specification: sl doctor revisited

**Feature Branch**: `610-sl-doctor-revisited`
**Created**: 2026-03-16
**Status**: Draft
**Input**: User description: "Fix #101 (remove /specledger.commit command), ensure doctor template management works from subdirectories (#81), and detectStaleFiles works as expected"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Stale template detection after CLI upgrade (Priority: P1)

A developer upgrades their `sl` CLI binary to a new version where a command has been removed from the embedded templates (e.g., `specledger.commit`). They run `sl doctor --template` to sync their project templates. The system copies the current set of templates and warns them about stale files (commands that exist on disk but are no longer in the manifest) so they can manually clean them up.

**Why this priority**: Without working stale detection, removed commands silently persist in user projects forever, creating confusion when Claude Code continues to trigger deprecated skills.

**Independent Test**: Can be fully tested by creating a temp project, placing an extra `specledger.*.md` file in `.claude/commands/`, running `sl doctor --template`, and verifying the stale file warning appears.

**Acceptance Scenarios**:

1. **Given** a project with `.claude/commands/specledger.commit.md` that is not in the manifest, **When** the user runs `sl doctor --template`, **Then** the output includes a warning listing `specledger.commit.md` as stale with a recommendation to re-run with `--force` to delete.
2. **Given** stale files detected, **When** the user runs `sl doctor --template --force`, **Then** all stale `specledger.*.md` files are deleted with a confirmation message listing each removed file.
3. **Given** a project where all `.claude/commands/specledger.*.md` files match the manifest, **When** the user runs `sl doctor --template`, **Then** no stale file warnings are shown.
4. **Given** a project with custom non-specledger commands (e.g., `my-deploy.md`) in `.claude/commands/`, **When** the user runs `sl doctor --template` (with or without `--force`), **Then** custom commands are never flagged or deleted.

---

### User Story 2 - Remove specledger.commit skill (Priority: P1)

A developer using SpecLedger no longer needs the `/specledger.commit` skill because the L0 PostToolUse hook (`sl session capture`) already handles session capture automatically when `git commit` is detected via regex. The redundant skill and all references to it are removed from embedded templates, project files, and documentation.

**Historical context**: The PostToolUse hook was originally project-level (`.claude/settings.json`), but was removed in commit `a215636` (Mar 4, 2026) because it "was not reliably firing (especially after mid-session settings changes)". Inline capture was then moved INTO the `/specledger.commit` skill (commit `c0dffe7`). Currently `sl auth login` installs the hook to **global** `~/.claude/settings.json`. Both mechanisms coexist — removing the skill leaves the global hook as the sole session capture mechanism. The global installation must include an opt-out config option so users who don't want session capture can disable it without the hook being re-added on every `sl auth login`.

**Why this priority**: The skill actively causes friction — CLAUDE.md mandates its use, overriding Claude's natural git workflow. Removing it unblocks normal commit behavior while the L0 hook continues to provide session capture.

**Independent Test**: Can be verified by confirming `specledger.commit.md` is absent from embedded templates, manifest, `.claude/commands/`, CLAUDE.md, and constitution. The L0 hook at `~/.claude/settings.json` (PostToolUse → Bash → `sl session capture`) remains active and correctly detects `git commit` via regex.

**Acceptance Scenarios**:

1. **Given** the updated CLI binary, **When** a user runs `sl doctor --template`, **Then** `specledger.commit.md` is not deployed to `.claude/commands/` and any existing copy is reported as stale.
2. **Given** the updated project, **When** Claude Code processes a `git commit` bash command, **Then** the PostToolUse hook fires `sl session capture` which detects the commit via `gitCommitPattern` regex and captures the session (no `/specledger.commit` skill needed).
3. **Given** the updated CLAUDE.md, **When** Claude Code reads project instructions, **Then** there is no "Commit & Push Rules" section mandating `/specledger.commit` use.

---

### User Story 3 - Doctor works from subdirectories (Priority: P2)

A developer is working in a subdirectory of their SpecLedger project (e.g., `pkg/cli/commands/`) and runs `sl doctor` or `sl doctor --template`. The command finds the project root by walking up the directory tree (looking for `specledger.yaml`) rather than failing because the current directory doesn't contain it.

**Why this priority**: This is a usability bug (#81) that forces users to `cd` to the project root before running doctor. Other commands like `sl deps list` already handle this correctly.

**Independent Test**: Can be tested by `cd`-ing into a subdirectory and running `sl doctor --template`, verifying it succeeds.

**Acceptance Scenarios**:

1. **Given** a SpecLedger project at `/path/to/project/`, **When** the user runs `sl doctor --template` from `/path/to/project/pkg/cli/`, **Then** the command succeeds and updates templates correctly.
2. **Given** a directory that is not inside any SpecLedger project, **When** the user runs `sl doctor --template`, **Then** the command returns a clear error: "not in a SpecLedger project (no specledger.yaml found)".
3. **Given** a SpecLedger project, **When** the user runs `sl doctor` (interactive mode) from a subdirectory, **Then** the template status section correctly identifies the project and offers updates.
4. **Given** a SpecLedger project with outdated templates, **When** the user runs `sl doctor --check`, **Then** the command prints human-readable status showing what's outdated and exits non-zero, without prompting or making changes.

---

### User Story 4 - CLI scaffold commands include footer hints and JSON next-steps (Priority: P2)

A developer (or AI agent) runs `sl spec create` to scaffold a new feature. The CLI creates the branch and spec template file, but its output doesn't tell the agent what to do next — leading to errors like "File has not been read yet" when the agent tries to write without first reading the template. The same problem exists for `sl spec setup-plan`. Per the CLI design principle (Principle 3: Two-Level Output), human output MUST end with footer hints and JSON output MUST be complete — scaffold commands should follow this established pattern.

**Why this priority**: Without footer hints, agents skip reading templates and either fail (Write tool error) or produce specs/plans that don't follow the template structure. This creates a poor first-run experience especially during onboarding.

**Independent Test**: Run `sl spec create --json` and verify the output includes a `NEXT_STEPS` field. Run `sl spec create` (human mode) and verify footer hints are printed.

**Acceptance Scenarios**:

1. **Given** a user runs `sl spec create --json`, **When** the output is returned, **Then** it includes a `NEXT_STEPS` field with instructions to read the spec template at `.specledger/templates/spec-template.md` before writing `SPEC_FILE`.
2. **Given** a user runs `sl spec setup-plan --json`, **When** the output is returned, **Then** it includes a `NEXT_STEPS` field with instructions to read the plan template, checklist template, and constitution before writing `PLAN_FILE`.
3. **Given** a user runs `sl spec create` (human mode), **When** the output is printed, **Then** a footer hint is printed: `→ Read '.specledger/templates/spec-template.md' before writing the spec`.

---

### User Story 5 - Onboarding produces meaningful constitution principles (Priority: P3)

A new SpecLedger user runs the onboarding workflow (`/specledger.onboard`). The constitution step currently triggers `sl-audit` which discovers the tech stack and produces overly technical principles (e.g., "Use Go 1.24", "Use Cobra CLI"). Instead, the onboarding should guide the user toward high-level software design principles (e.g., YAGNI, DRY, testing strategy, deployment philosophy) that govern how the team builds software, not what tools they use.

**Why this priority**: The constitution is foundational — it governs all subsequent specs, plans, and reviews. A too-technical constitution fails its purpose and requires experienced users to manually correct it. New users won't know to fix it.

**Independent Test**: Run the onboarding flow and verify the constitution prompt asks for software design principles, not tech stack details.

**Acceptance Scenarios**:

1. **Given** a new project running `/specledger.onboard`, **When** the constitution step executes, **Then** the prompt explicitly asks for high-level software design principles (e.g., testing philosophy, code review standards, deployment strategy) rather than technology inventory.
2. **Given** the onboarding constitution prompt, **When** the agent explores the codebase for context, **Then** it uses the codebase to inform principle suggestions but frames them as development practices, not tool choices.
3. **Given** a generated constitution, **When** the user reviews it, **Then** principles are actionable guidelines (e.g., "Fail Fast, Fix Forward — surface errors early with clear messages") rather than technology declarations (e.g., "Use Go 1.24.2 with Cobra CLI").

---

### User Story 6 - Checkpoint captures implementation decisions (Priority: P3)

During implementation, developers make decisions that diverge from the plan (e.g., a whitespace normalization choice not identified during planning). The `/specledger.checkpoint` command should have a structured `### Decision Log` section to capture these decisions so they don't get lost between sessions.

**Why this priority**: Decisions made during implementation are high-value context that gets lost without structured capture. Currently the checkpoint template has `### Notes` but no dedicated place for decisions.

**Independent Test**: Run `/specledger.checkpoint` after closing tasks and verify the decision log section is present and prompts for divergences.

**Acceptance Scenarios**:

1. **Given** a checkpoint is triggered after closing tasks, **When** the agent generates the checkpoint, **Then** the `### Decision Log` section prompts: "Did implementation diverge from plan/spec? Were assumptions invalidated?"
2. **Given** a decision was made during implementation, **When** the user records it in the decision log, **Then** the entry includes What, Why, Impact level, and Artifacts affected.

---

### User Story 7 - Implement command uses sl issue ready for task selection (Priority: P3)

The `/specledger.implement` command currently uses `sl issue list --status in_progress` for task selection. It should use `sl issue ready` which handles dependency resolution automatically — finding tasks whose blockers are all satisfied.

**Why this priority**: `sl issue ready` already exists and handles the complexity of dependency checking. Using it aligns with the documented command inventory and provides better UX.

**Independent Test**: Run `/specledger.implement` and verify it calls `sl issue ready` for picking the next task.

**Acceptance Scenarios**:

1. **Given** the implement workflow starts, **When** it picks the next task to work on, **Then** it uses `sl issue ready` (not `sl issue list --status in_progress`).
2. **Given** a task has unsatisfied blockers, **When** `sl issue ready` is called, **Then** that task is not offered as the next work item.

---

### User Story 8 - Comment commands accept short UUID prefixes (Priority: P3)

A developer runs `sl comment list` which shows truncated comment IDs (e.g., `fda6ac86-cab`). When they try to resolve a comment using this truncated ID, the command fails because it requires the full UUID. All `sl comment` subcommands that accept `<comment-id>` should resolve short prefixes to full UUIDs, like git does with commit hashes.

**Why this priority**: This is a usability bug discovered during clarification of this spec. It affects all comment operations (resolve, show, reply) and makes the CLI frustrating to use.

**Independent Test**: Run `sl comment resolve <short-prefix> --reason "test"` and verify it resolves to the full UUID and succeeds.

**Acceptance Scenarios**:

1. **Given** a comment with ID `fda6ac86-cab4-4850-...`, **When** the user runs `sl comment resolve fda6ac86 --reason "test"`, **Then** it resolves the prefix to the full UUID and succeeds.
2. **Given** two comments whose IDs share a prefix, **When** the user provides the ambiguous prefix, **Then** the command errors with "ambiguous comment ID, matches: ..." listing the full IDs.
3. **Given** no comment matches the prefix, **When** the user provides a non-matching prefix, **Then** the command errors with "comment not found" and suggests `sl comment list`.

---

### User Story 9 - Hook installation respects opt-out config (Priority: P2)

A developer who doesn't want session capture runs `sl auth hook --remove` to uninstall the PostToolUse hook. Currently, the next `sl auth login` silently re-installs it. The CLI should respect an opt-out so users don't have to remove the hook after every login.

**Why this priority**: Re-installing after explicit removal is "virus-like" behavior. Users must be able to control what gets installed to their global Claude settings.

**Independent Test**: Remove the hook, set the opt-out config, run `sl auth login`, verify the hook is NOT re-installed.

**Acceptance Scenarios**:

1. **Given** a user runs `sl auth hook --remove`, **When** the opt-out is persisted (e.g., in `~/.specledger/config.yaml` or `specledger.yaml`), **Then** subsequent `sl auth login` calls do NOT re-install the hook.
2. **Given** a user with opt-out enabled, **When** they run `sl auth hook --install`, **Then** the hook is installed AND the opt-out is cleared (explicit install overrides opt-out).
3. **Given** a fresh project with no opt-out config, **When** the user runs `sl auth login`, **Then** the hook is installed as before (backward compatible).

---

### Edge Cases

- What happens when `.claude/commands/` directory doesn't exist? Stale detection silently returns with no error.
- What happens when a user has `specledger.commit.md` AND it was customized? Stale warning shown. `--force` deletes it with a warning that anything not tracked by version control will be lost without backup. Files under git are recoverable via `git checkout`.
- What happens when `sl doctor --template` is run outside any git repository? Fails with: "Not in a git directory — are you sure you're in a SpecLedger project?"
- What happens when multiple stale files exist? All are listed in the warning output. All deleted when `--force` is used.
- What happens when `findProjectRoot()` reaches filesystem root without finding `specledger.yaml`? Returns clear error with navigation guidance per cli.md Principle 2.
- What happens when `sl spec create` is called but `.specledger/templates/` doesn't exist on disk? CLI reads directly from embedded templates (they are always available in the binary). No fallback needed — embedded is the source of truth. Templates on disk are copies for user reference only.
- What happens when onboarding is run on a project that already has a constitution? The existing constitution is preserved (protected in manifest).
- What happens when `sl doctor --template` overwrites a user-modified runtime template? Embedded is source of truth — runtime copies are always overwritten. Users who customized templates rely on git to resolve conflicts (KISS).

## Clarifications

### Session 2026-03-16

- Q: Should `sl doctor --template` offer `--force` to auto-delete stale `specledger.*` files? → A: Yes — warn by default, `--force` deletes with confirmation. The `specledger.` prefix is CLI-owned.
- Q: Should hook installation stay global or move to project-level? → A: Keep global but add opt-out config so users can disable session capture without the hook being re-added on every `sl auth login`.
- Q: Should users be able to modify runtime templates, and what happens on conflict? → A: Embedded is source of truth. `sl doctor --template` always overwrites. Users who customize rely on git for conflict resolution (KISS).
- Q: Should #84 (decision log) and #92 (sl issue ready) be added to this spec? → A: Yes, as P3 stories. Also add KISS to the project constitution.
- Q: Should FR-008 also audit `docs/design/commands.md`? → A: Yes, audit both README.md and commands.md.
- Q: Should the "protected" concept be visible to users? → A: Yes, show in `sl doctor --template` output which files were skipped as protected.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: `detectStaleFiles()` MUST scan `.claude/commands/` using `os.ReadDir()` and report files matching `specledger.*.md` that are not in the playbook manifest
- **FR-002**: By default, stale files MUST be warned about but NOT auto-deleted. With `--force` flag, stale `specledger.*` files MUST be deleted with a confirmation message listing each removed file. The `specledger.` prefix is CLI-owned; custom commands (no prefix) are never touched
- **FR-003**: `specledger.commit` MUST be removed from the playbook manifest
- **FR-004**: The embedded template file `specledger.commit.md` MUST be deleted
- **FR-005**: The project-level `.claude/commands/specledger.commit.md` MUST be deleted
- **FR-006**: ~~CLAUDE.md "Commit & Push Rules" section MUST be removed~~ **DONE** — replaced with session-start `sl doctor` reminder inside specledger sentinels. The `<!-- MANUAL ADDITIONS -->` markers will be fully removed when FR-017 (sentinel unification) is implemented; user content (Pre-push Checklist) will be preserved outside sentinel blocks
- **FR-007**: Constitution line referencing `/specledger.commit` MUST be removed
- **FR-008**: Design docs MUST remove `commit` from the workflow diagram and escape hatches list in `docs/design/README.md` (lines 63, 77) AND audit `docs/design/commands.md` for any remaining commit command references
- **FR-009**: `performTemplateUpdate()` and `outputDoctorHuman()` MUST use a shared `findProjectRoot()` utility instead of `os.Getwd()`
- **FR-010**: The existing `findProjectRoot()` in `deps.go` MUST be extracted to a shared package and reused by both `doctor.go` and `deps.go`
- **FR-011**: `sl spec create --json` output MUST include a `NEXT_STEPS` field instructing agents to read `.specledger/templates/spec-template.md` before writing the spec file
- **FR-012**: `sl spec setup-plan --json` output MUST include a `NEXT_STEPS` field instructing agents to read the plan template, checklist template, and constitution before writing
- **FR-013**: `sl spec create` and `sl spec setup-plan` human mode output MUST print next-step guidance after file paths
- **FR-014**: The onboarding constitution prompt (`specledger.onboard.md` and/or `specledger.constitution.md`) MUST guide toward high-level software design principles, not technology inventory
- **FR-015**: The onboarding constitution prompt MUST provide example principle categories (testing philosophy, code standards, deployment strategy, error handling approach) to steer agents away from tech-stack enumeration
- **FR-016**: ~~All embedded templates MUST be synced with their runtime copies~~ **DONE** — `checklist-template.md` and `specledger.tasks.md` embedded copies have been synced to match runtime. The CI drift guard (FR-019) prevents this from recurring
- **FR-017**: `sl context update` (`pkg/cli/context/updater.go`) MUST be refactored to use `MergeSentinelSection()` from `pkg/cli/playbooks/merge.go` instead of its bespoke `<!-- MANUAL ADDITIONS -->` marker system. Both implement the same pattern (managed section + preserved user content) but the sentinel approach is explicit, composable, and already manifest-driven. After refactoring, CLAUDE.md becomes a regular mergeable file — `sl doctor --template` manages the specledger-generated section, `sl context update` manages Active Technologies within a second sentinel block, and user content lives outside both
- **FR-018**: The CLI MUST provide a way for agents to retrieve the current checklist template — either via `sl spec create` footer hints pointing to `.specledger/templates/checklist-template.md`, or via a dedicated `sl spec checklist` command that outputs the latest embedded checklist template. The hardcoded checklist structure in agent command prompts should be removed in favor of reading the CLI-provided template
- **FR-019**: CI MUST include a template drift guard: `make build && ./bin/sl doctor --template` followed by `git diff --exit-code` on template-managed paths (`.claude/commands/`, `.claude/skills/`, `.specledger/templates/`). If drift is detected, the CI check fails — forcing contributors to update embedded templates when they change runtime copies (or vice versa)
- **FR-020**: `sl doctor` MUST support a `--check` flag: human-readable dry-run that reports CLI version status and template freshness without prompting or making changes. Exits non-zero if updates are needed. This is the flag CLAUDE.md should recommend at session start (not `--json` which is for CI piping)
- **FR-021**: The CLAUDE.md managed section (injected by FR-017) MUST recommend `sl doctor --check` and suggest `sl doctor --update --template` if outdated
- **FR-022**: `sl auth login` MUST respect an opt-out config (e.g., `session_capture: false` in `specledger.yaml` or `~/.specledger/config.yaml`) to prevent the PostToolUse hook from being re-installed on every login. `sl auth hook --remove` should persist the opt-out so the user doesn't have to remove it again
- **FR-023**: `sl doctor --template` output MUST list protected files separately (e.g., "Skipped N protected files: constitution.md, AGENTS.md") so users can see what won't be overwritten
- **FR-024**: The `/specledger.checkpoint` command template MUST include a `### Decision Log` section that captures implementation decisions diverging from plan/spec/tasks — structured as: What, Why, Impact (minimal/moderate/significant), Artifacts affected
- **FR-025**: The `/specledger.implement` command template MUST use `sl issue ready` instead of `sl issue list --status in_progress` for task selection. Resume logic may still check in-progress tasks, but picking the next task should use the readiness-based command
- **FR-026**: The project constitution MUST include a KISS (Keep It Simple, Stupid) principle — embedded templates on disk are copies of the CLI binary's source of truth; users who customize rely on git for conflict resolution; no complex merge infrastructure needed
- **FR-027**: All `sl comment` subcommands that accept `<comment-id>` MUST support short UUID prefix matching — load all comment IDs for the current spec, find those starting with the given prefix, error if 0 or >1 match, use the full UUID if exactly 1 match

### Key Entities

- **Playbook Manifest** (`manifest.yaml`): Defines which commands/skills are canonical — source of truth for stale detection
- **Template Update Result** (`TemplateUpdateResult`): Carries `Stale` field populated by `detectStaleFiles()`, consumed by doctor's UI output

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Running `sl doctor --template` from any subdirectory within a SpecLedger project succeeds
- **SC-002**: After removing `specledger.commit` from the manifest, running `sl doctor --template` on a project with the old file produces a stale warning
- **SC-003**: `specledger.commit.md` is absent from the compiled binary's embedded templates
- **SC-004**: No references to `/specledger.commit` remain in CLAUDE.md, constitution, or design docs
- **SC-005**: All existing tests pass and new stale detection unit tests validate the logic
- **SC-006**: Lint passes with no new warnings
- **SC-007**: `sl spec create --json` output includes `NEXT_STEPS` with template read instructions; agents using this output read the template before writing
- **SC-008**: `sl spec setup-plan --json` output includes `NEXT_STEPS` with template read instructions
- **SC-009**: Onboarding-generated constitution contains principles about how to build software (testing, code quality, deployment), not what tools to use
- **SC-010**: ~~Embedded templates match runtime copies~~ **DONE** — verified zero drift. CI drift guard (FR-019) prevents recurrence
- **SC-011**: After `sl doctor --template`, CLAUDE.md contains a `# >>> specledger-generated` section with a session-start `sl doctor` reminder, separate from the `<!-- MANUAL ADDITIONS -->` block
- **SC-012**: The `/specledger.specify` checklist is sourced from the CLI's embedded template, not hardcoded in the command prompt
- **SC-013**: `sl doctor --template --force` deletes stale `specledger.*` files and reports each deletion
- **SC-014**: After `sl auth hook --remove`, re-running `sl auth login` does NOT re-install the hook (opt-out persisted)
- **SC-015**: `sl doctor --template` output lists protected files that were skipped
- **SC-016**: Checkpoint template includes `### Decision Log` with structured fields
- **SC-017**: Implement command uses `sl issue ready` for task selection
- **SC-018**: `sl comment resolve <short-prefix>` succeeds when prefix is unambiguous

### Previous work

- **#64**: feat: SDD workflow streamline - bash CLI migration + spec verification (merged — introduced stale detection stub that was never completed)

---

## GitHub Issues Addressed

This spec directly addresses the following open issues:

| Issue | Title | Relationship |
|-------|-------|-------------|
| [#101](https://github.com/specledger/specledger/issues/101) | chore: remove specledger.commit skill | US-2: Remove skill, manifest entry, docs references |
| [#81](https://github.com/specledger/specledger/issues/81) | bug: sl doctor --template fails to find project root from subdirectories | US-3: Extract shared `findProjectRoot()`, fix doctor.go |
| [#90](https://github.com/specledger/specledger/issues/90) | Improve agent prompts: read template files before writing content | US-4: Add `NEXT_STEPS` to `sl spec create` and `sl spec setup-plan` output |
| [#91](https://github.com/specledger/specledger/issues/91) | Onboarding explore/audit for principles is too technical | US-5: Fix constitution prompt to guide toward design principles |
| [#96](https://github.com/specledger/specledger/issues/96) | refactor: extract ContextDetector to shared package and standardize --spec flag | FR-010: Overlaps with extracting `findProjectRoot()` to shared package |
| [#82](https://github.com/specledger/specledger/issues/82) | Improve embedded skill templates: fix duplicates, optimize triggering | FR-016: Sync embedded checklist template drift |
| [#84](https://github.com/specledger/specledger/issues/84) | Add decision log section to checkpoint prompt template | US-6: Decision log in checkpoint template |
| [#92](https://github.com/specledger/specledger/issues/92) | Use `sl issue ready` in implement command | US-7: Replace task selection with readiness-based command |
| [#106](https://github.com/specledger/specledger/issues/106) | sl comment resolve requires full UUID, should accept short prefix | US-8: Prefix-matching UUID resolution for comment commands |

---

## Planning Phase Notes

> **Testing debt**: The plan phase MUST include a task phase to align the test infrastructure with the prescription in `tests/README.md`. The README defines a 3-tier testing strategy (Unit → Integration → E2E) with several "Planned additions" that are not yet implemented. For this spec specifically:
>
> - **Unit tests**: New `detectStaleFiles()` logic needs table-driven tests. New `findProjectRoot()` shared utility needs tests.
> - **Integration tests**: `tests/integration/doctor_test.go` already exists — extend it to cover subdirectory resolution and stale file detection (real binary invocation).
> - **go-vcr cassettes**: Not applicable to this spec (no pgREST interactions), but the plan should note the gap for future specs.
> - **E2E tests**: `tests/e2e/` directory does not exist yet. This spec should create it if quickstart scenarios are defined, or at minimum note the gap.
>
> The plan must not add tests that violate the constitution's testing tiers (Principle VI). Specifically: no hand-crafted `httptest` mocks where go-vcr cassettes are prescribed, and no unit tests pretending to be integration tests.
>
> **CLAUDE.md migration**: The plan MUST include a phase to migrate CLAUDE.md from the bespoke `<!-- MANUAL ADDITIONS START/END -->` pattern to the standard `MergeSentinelSection()` pattern. During this migration:
> - The temporary `sl doctor --json` Session Start block (currently inside specledger sentinels within manual additions) MUST be removed — it will be replaced by the `--check` flag guidance once FR-020 is implemented
> - Existing user content from manual additions (e.g., Pre-push Checklist) MUST be preserved outside the sentinel blocks
> - The `sl context update` code MUST be refactored to use a second sentinel block (e.g., `# >>> specledger-context`) for Active Technologies instead of regenerating the entire file above manual additions
