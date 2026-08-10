# Revision Log: 011-streamline-onboarding spec.md

**Date**: 2026-02-21
**Reviewer feedback**: 8 comments on spec.md

## Comment Tracker

| # | Target Text | Status | Choice |
|---|-------------|--------|--------|
| 1 | "then are guided through creating their project constitution" | Done | Option A: Reorder only |
| 2 | "After confirming, SpecLedger creates the project with a populated constitution" | Done | Option B: TUI seeds, agent completes |
| 3 | "that guides them through the SpecLedger workflow" | Done | Option B: Explicit commit gate |
| 4 | "the system launches an exploration step..." | Done | Option B: Optional + terminology note |
| 5 | "After setup, the user is offered the option to launch..." | Done | Option A: Fix TUI ref + clarify agent context |
| 6 | "then leads the user through creating their first feature specification..." | Done | Option A: Constitution + suggest next steps |
| 7 | "Git initialization is offered as an optional step" | Done | Option A: Require git init |
| 8 | "Any partial setup is rolled back cleanly" | Done | Option B: Simple abort |

## Detailed Revision History

### Comment 1
**Feedback**: Constitution creation should come after agent shell selection (playbook informs constitution rules)
**Options presented**:
- A (chosen): Reorder only — agent preference moves into setup questions, constitution follows with playbook-informed principles
- B: Reorder + defer constitution finalization to agent session
- C: Minimal reorder — swap order with minimal wording change
**Applied**: Reordered flow in User Story 1 line 21. Agent preference now part of setup questions, constitution creation follows informed by playbook.

### Comment 2
**Feedback**: Constitution should not be fully populated in TUI — agent should finalize it using AskUserQuestion interaction
**Options presented**:
- A: Agent finalizes — project created with basic config, agent does all constitution work
- B (chosen): TUI seeds, agent completes — TUI provides seed principles, agent refines/finalizes interactively
- C: Fully defer to agent — no TUI constitution at all, 100% agent-driven
**Applied**: Changed "populated constitution" to "seed constitution principles" and added that the agent refines/finalizes interactively.

### Comment 3
**Feedback**: Constitution should be separate commit/PR before any feature work. Agent finalizes constitution first, commits, THEN proposes next steps.
**Options presented**:
- A: Append workflow detail — soft mention of commit-then-next-steps
- B (chosen): Explicit commit gate — strong language about standalone commit before feature work
- C: Minimal addition — brief inline note
**Applied**: Added "The finalized constitution is committed as a standalone commit before any feature work begins. The agent then guides the user to their first feature specification."

### Comment 4
**Feedback**: Codebase exploration should be optional (agent asks user first) due to token cost. Use agent-specific terminology in plan docs.
**Options presented**:
- A: Make optional with ask — agent asks before running audit, falls back to defaults
- B (chosen): Optional + terminology note — same as A plus note about plan-phase terminology
- C: Soft optional (default yes) — exploration is default but user can skip
**Applied**: Changed User Story 2 and User Story 3 audit text to be opt-in (agent offers, user chooses). Added plan-phase terminology note to User Story 2.

### Comment 5
**Feedback**: Agent is already running in the shell — it's not "launched after setup." Also User Story 4 had stale "constitution was created during TUI" text.
**Options presented**:
- A (chosen): Fix TUI reference + clarify agent context — rewrite User Story 4 opening to say agent operates within shell session, TUI only seeds, agent finalizes, constitution committed before features
- B: Lighter touch — just fix "created" to "seeded" and add "within the agent session"
**Applied**: Rewrote User Story 4 opening paragraph and acceptance scenarios 1-2 to align with Comments 2-3 (TUI seeds, agent finalizes, commit gate).

### Comment 6
**Feedback**: Onboarding should only ensure constitution is in place, then suggest `/specledger.specify` (after `/clear` for context) or provide docs links. Should NOT run full workflow.
**Options presented**:
- A (chosen): Constitution + suggest next steps — scope onboarding to constitution + next-step guidance, remove deep workflow scenarios
- B: Constitution + light guidance — keep light workflow mention, simplify scenarios
- C: Keep full workflow but mark optional — offer choice between walkthrough or docs
**Applied**: Rewrote User Story 4 ending to suggest `/clear` + `/specledger.specify`, replaced acceptance scenarios 3-6 with simpler next-step scenarios, updated test description.

### Comment 7
**Feedback**: speckit forces git init — git initialization is not optional
**Options presented**:
- A (chosen): Require git init — system initializes git automatically if none exists
- B: Force with notice — auto-init + inform user
- C: Error if no git — display error, user must init manually
**Applied**: Changed edge case answer to "system initializes git automatically." Also updated FR-019 and User Story 3 acceptance scenario 6 for consistency.

### Comment 8
**Feedback**: Rollback is YAGNI/KISS. Just abort — tell user to use git.
**Options presented**:
- A: Abort + git advice — abort cleanly with specific git commands
- B (chosen): Simple abort — abort immediately, user uses git to revert
- C: Abort + re-run safe — abort with note that re-running is idempotent
**Applied**: Replaced rollback language with simple "Setup aborts immediately. The user can use git to revert any partial changes."

