---
description: Critical divergence review — compare implementation against plan artifacts, flag force-closed issues, and surface gaps. Updates session log at FEATURE_DIR/sessions/<branch-name>-checkpoint.md
---

## User Input

```text
$ARGUMENTS
```

You **MUST** consider the user input before proceeding (if not empty).

**Execution Tracking**: Before starting work, create a task list (using the TaskCreate tool) covering all execution steps in this workflow. If `$ARGUMENTS` contains user-specified actions beyond the standard workflow, place those tasks where they logically fit: before setup steps if arguments change what gets set up, or after all standard steps if arguments extend the workflow. Update task status as you complete each step.

**User Interaction**: Whenever you need input, clarification, or a decision from the user, use the **AskUserQuestion** tool directly. Do not output questions as plain text and stop — always use the interactive tool for proper UX.

## Purpose

Perform a critical divergence review of the current implementation state against plan artifacts. Your job is to **find problems, not confirm success**. Surface plan drift, force-closed issues, uncovered requirements, and implementation gaps that human reviewers need to know about before merge.

**When to use**: During or after implementation to catch drift, before handoff, or before merging.

## Framing

Adopt an adversarial reviewer mindset. Assume the implementation has gaps until proven otherwise. The completed task list is already visible in the issue tracker — your value is in finding what it **doesn't** show.

## Outline

Goal: Identify divergences between planned and actual implementation, classify them, and produce an actionable review.

Execution steps:

1. Run `sl spec info --json --paths-only` to get `FEATURE_DIR` and `BRANCH`.

2. Gather implementation state:

   ```bash
   # All closed issues with full details
   sl issue list --status closed --json

   # Remaining work
   sl issue list --status open --json
   sl issue list --status in_progress --json

   # Uncommitted changes
   git status --porcelain
   ```

   **Detect force-closed issues**: Iterate closed issues. Any issue where `definition_of_done` exists and contains items with `checked: false` was force-closed (DoD bypassed). Flag every one of these — they are the highest-signal findings.

3. Run project tests and checks:
   - Consult the project's `CLAUDE.md` (or equivalent) for the canonical test/lint/format commands.
   - If no project-level instructions exist, detect the project type and use conventional commands:
     - **Go**: `go test ./...`
     - **Node (npm/pnpm/yarn)**: check `package.json` for `test`, `lint`, `format:check` scripts and run those that exist
     - **Python**: `pytest` or the configured test runner
     - **Other**: look for a `Makefile`, `justfile`, or CI config for test commands
   - If no test runner is configured, state that explicitly — do not fabricate a test step
   - All executed checks must pass (exit code 0) for a clean checkpoint
   - If any check fails, report failures and include them as CRITICAL divergences

4. Compare implementation against plan artifacts:

   Read the following from `FEATURE_DIR` (skip any that don't exist):

   **From spec.md:**
   - Functional requirements (FR-xxx or numbered list)
   - User stories and their acceptance criteria
   - Non-functional requirements
   - Edge cases

   **From plan.md:**
   - Phases and their deliverables
   - Project structure (expected files/components)
   - Architecture decisions and constraints

   **From data-model.md** (if present):
   - Entity names and key fields
   - Validation rules
   - Relationships

   **From quickstart.md** (if present):
   - Integration scenarios
   - Expected output formats

   For each artifact claim, check:
   - Is there a closed issue that covers this requirement?
   - Does the implementation match the specification? (Check actual code if uncertain.)
   - Are there planned files/components that don't exist?
   - Are there data model entities defined but not implemented, or implemented differently?
   - Are there quickstart scenarios not validated by tests?

5. Classify each divergence:

   **Severity** (use same scale as `/specledger.verify`):
   - **CRITICAL**: Missing core requirement, failing tests, security/compliance gap
   - **HIGH**: Force-closed issue with significant unchecked DoD, requirement partially implemented, test gap for critical path
   - **MEDIUM**: Data model drift, terminology inconsistency, undocumented architecture change
   - **LOW**: Minor format difference, non-critical edge case not covered

   **Type** — check the issue's `notes` field and any Decision Log entries in the session log:
   - **conscious**: Divergence is documented somewhere (issue notes, decision log, commit message)
   - **oversight**: No documentation found — this was likely missed

6. Update session log:
   - Create `FEATURE_DIR/sessions/` directory if it doesn't exist
   - **Determine output file based on scope**:
     - **Phase-scoped checkpoint**: If `$ARGUMENTS` indicates a phase scope (e.g., `"Verify phase:setup issues only"`), write to `FEATURE_DIR/sessions/<branch-name>-checkpoint-<phase-name>.md`. One file per phase, overwriting any prior phase-scoped checkpoint for the same phase.
     - **Full checkpoint** (no phase scope): Append a timestamped entry to `FEATURE_DIR/sessions/<branch-name>-checkpoint.md`.
   - Use the entry format below

   ```markdown
   ## Divergence Review: YYYY-MM-DD HH:MM

   ### Divergences

   | # | Severity | Type | Category | Issue/Artifact | Description |
   |---|----------|------|----------|----------------|-------------|
   | 1 | HIGH | oversight | Missing requirement | spec.md FR-003 | Rate limiting not implemented by any closed issue |
   | 2 | MEDIUM | conscious | Data model drift | SL-xxx / data-model.md | Field renamed from X to Y (documented in issue notes) |

   ### Force-Closed Issues (DoD Bypassed)

   | Issue | Title | Unchecked DoD Items | Risk |
   |-------|-------|---------------------|------|
   | SL-xxx | Add validation | "Integration test passes" unchecked | HIGH — no test coverage |

   ### Issues Encountered & Resolutions
   - <What went wrong> → <How it was resolved or worked around>

   ### Items Requiring Action Before Merge
   1. [CRITICAL] Fix <specific gap> — <why it matters>
   2. [HIGH] Write test for <scenario> — <what's at risk>

   ### Tests & Checks
   - Status: PASS/FAIL/SKIPPED
   - Commands run: <list of commands executed>
   - Failures: <details if any>

   ### Progress Summary
   - Closed: N issues
   - In Progress: N issues
   - Open/Remaining: N issues
   - Force-Closed: N issues (DoD bypassed)

   ### Uncommitted Changes
   - <File paths or "None">

   ---
   ```

7. Report divergence summary to the user:
   - Lead with divergence count and severity breakdown
   - Show the divergence table and force-closed issues table
   - List items requiring action
   - End with test status and progress numbers
   - If CRITICAL divergences exist, recommend resolving before merge

8. Offer adversarial deep-dive agent:

   After reporting your findings, **always offer** to launch an independent adversarial review agent. This agent runs in a fresh context with no knowledge of the implementation session — it cannot rationalize shortcuts or inherit anchoring bias from prior decisions.

   **When all issues are closed** (zero open, zero in-progress), **strongly recommend** running the adversarial agent as a best practice before merge. This is the highest-value moment: the work appears complete, so the risk of undetected drift is greatest.

   Otherwise, present it as an optional next step — useful when the checkpoint is mid-implementation and more sessions are expected.

   Offer the user this prompt (adjust `FEATURE_DIR` and `BRANCH` from step 1):

   Use AskUserQuestion to ask: **"Would you like me to launch an independent adversarial review agent?"** This runs in a separate context with no memory of this session — it reviews the code and artifacts cold.

   If the user accepts, launch an Agent with this prompt:

   ~~~
   You are an adversarial code reviewer. Your job is to find problems, not confirm success.

   ## Context
   - Feature directory: {FEATURE_DIR}
   - Branch: {BRANCH}
   - This review is context-free by design — you have no prior knowledge of
     implementation decisions or tradeoffs made during development.

   ## Instructions

   1. Run `sl spec setup-plan --json` from repo root and parse JSON for FEATURE_SPEC, IMPL_PLAN, SPECS_DIR, BRANCH
   2. Read the spec, plan, and any design artifacts in {FEATURE_DIR}.
   3. Run `sl issue list --status closed --json` and `sl issue list --status open --json`.
   4. Flag any closed issue where `definition_of_done` contains `checked: false` items.
   5. Read the actual implementation code on this branch. For each requirement and
      planned deliverable, verify it exists and behaves as specified.
   6. Run the project's test/lint commands (check CLAUDE.md for canonical commands).
   7. Produce a findings report:
      - Divergences (severity + conscious/oversight classification)
      - Force-closed issues with unchecked DoD
      - Code quality concerns (dead code, missing error handling, untested paths)
      - Requirements with no corresponding implementation
      - Implementation that has no corresponding requirement (scope creep)
   8. Be specific: cite file paths, line numbers, issue IDs, and artifact references.
   9. Report findings only — do not fix anything.
   ~~~

## Behavior Rules

- **Lead with divergences, not accomplishments** — the progress summary is an appendix
- **Flag every force-closed issue** — unchecked DoD on a closed issue is always worth reporting
- **Classify every divergence** as conscious or oversight by checking issue notes and decision logs
- **If zero divergences found**, report that explicitly — this is a positive signal worth stating, not a default
- All executed tests/checks must pass for a clean checkpoint
- Don't auto-commit — prompt user instead
- If CRITICAL divergences exist, strongly recommend resolving before merge
- If no progress since last checkpoint, report "no changes detected"
- Include file paths for uncommitted changes

## Example Usage

```bash
# Critical divergence review after implementation
/specledger.checkpoint

# Review with specific focus area
/specledger.checkpoint "Focus on data model alignment and test coverage"

# Pre-merge divergence review
/specledger.checkpoint "Pre-merge review for PR #42"

# Checkpoint with known context
/specledger.checkpoint "We switched from go-vcr to httptest — flag that as conscious"
```

## Session Log Format

Session logs are stored at `FEATURE_DIR/sessions/<branch-name>-checkpoint.md`:

```markdown
# Session Log: <branch-name>

## Divergence Review: 2026-03-05 14:30

### Divergences

| # | Severity | Type | Category | Issue/Artifact | Description |
|---|----------|------|----------|----------------|-------------|
| 1 | HIGH | oversight | Missing requirement | spec.md FR-009 | JSONL fallback on 404 not implemented — only shows warning |
| 2 | LOW | conscious | Architecture change | plan.md Phase 2 | Used httptest instead of go-vcr cassettes (documented in SL-6a0837 notes) |
| 3 | MEDIUM | oversight | Test gap | quickstart.md Scenario 12 | TestPlanShowCacheReuse never written |

### Force-Closed Issues (DoD Bypassed)

| Issue | Title | Unchecked DoD Items | Risk |
|-------|-------|---------------------|------|
| SL-6a0837 | go-vcr cassette setup | "Cassette file created", "Replay test passes" | LOW — httptest approach covers same ground |
| SL-d45f35 | TestPlanShowCacheReuse | "Test implemented", "Cache hit verified" | MEDIUM — no test for cache reuse path |

### Issues Encountered & Resolutions
- TestParsePlanJSONSensitive failed: sensitive values compared equal → added isSensitive flag
- TestRunCancelJSON mock returned non-cancelable state → fixed mock to return cancelable first

### Items Requiring Action Before Merge
1. [HIGH] Fix Scenario 11 JSONL fallback (spec.md FR-009 requires it)
2. [MEDIUM] Write TestPlanShowCacheReuse or document why it's deferred
3. [MEDIUM] Verify formatAttrValue output matches quickstart scenarios

### Tests & Checks
- Status: PASS
- Commands run: go test ./pkg/cli/commands/... ./pkg/plan/...
- 21 tests passing

### Progress Summary
- Closed: 33 issues
- In Progress: 0 issues
- Open/Remaining: 0 issues
- Force-Closed: 7 issues (DoD bypassed)

### Uncommitted Changes
- None

---
```
