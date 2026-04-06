---
description: Guided onboarding workflow - walks through the full SpecLedger process from constitution to implementation
---

**User Interaction**: Whenever you need input, clarification, or a decision from the user, use the **AskUserQuestion** tool directly. Do not output questions as plain text and stop — always use the interactive tool for proper UX.

## Purpose

Guide a new user through the complete SpecLedger workflow after project setup. This command is designed to be run immediately after `sl new` or `sl init` launches the coding agent.

## When to use

- Right after `sl new` or `sl init` launches this agent session
- When a user wants a guided walkthrough of the SpecLedger process
- When onboarding a new team member to a SpecLedger project

## Workflow

### Step 1: Constitution Check

Check if a populated project constitution exists at `.specledger/memory/constitution.md`.

**If constitution exists** (populated, no `[PLACEHOLDER]` tokens):
- Confirm: "Your project constitution is in place."
- Show a brief summary of the core principles.
- Proceed to Step 2.

**If constitution is missing or unfilled** (contains `[PLACEHOLDER]` tokens or file doesn't exist):
- Explain: "Your project doesn't have a constitution yet. Let's create one by analyzing your codebase."
- Run `/specledger.audit` to build familiarity with the codebase structure and patterns.
- Run `/specledger.constitution` to propose guiding principles for the project.
  - **Important**: The constitution captures high-level software design principles (e.g., YAGNI, test-first, simplicity, contract-driven design) — NOT technology selections discovered during the audit. The audit provides codebase familiarity; principles come from how the team wants to approach software design.
- Wait for the user to review and approve the constitution before proceeding.

**Commit Suggestion**: After the constitution is created or confirmed, use AskUserQuestion to suggest:

> **Constitution Ready — Commit Setup?**
>
> Your project constitution is in place. It's a good idea to commit this setup before starting feature work.
>
> **Would you like to commit the constitution and setup files now?**

If the user accepts, commit the constitution and any setup files. If skipped, proceed to Step 2.

### Step 2: Welcome & Orientation

Present a brief welcome message:

> **Welcome to SpecLedger!**
>
> SpecLedger helps you build features methodically:
> 1. **Specify** - Describe what you want to build
> 2. **Clarify** - Resolve ambiguities in the spec
> 3. **Plan** - Design the implementation approach
> 4. **Tasks** - Generate actionable, ordered tasks
> 5. **Review** - You review tasks before any code is written
> 6. **Verify** - Cross-check spec/plan/task consistency (recommended)
> 7. **Implement** - Execute the tasks
> 8. **Checkpoint** - Review implementation against plan (recommended)
>
> Let's start by describing your first feature!

### Step 2.5: Command Overview

Before we begin, here's a quick reference of the available SpecLedger commands:

**Core Workflow Commands:**
| Command | Description |
|---------|-------------|
| `/specledger.specify` | Create a feature specification from user requirements |
| `/specledger.clarify` | Resolve ambiguities and answer spec questions |
| `/specledger.plan` | Generate technical implementation plan |
| `/specledger.tasks` | Create ordered, dependency-linked task list |
| `/specledger.verify` | Cross-artifact consistency and quality check |
| `/specledger.implement` | Execute implementation tasks in order |
| `/specledger.checkpoint` | Divergence review during/after implementation |

**Utility Commands:**
| Command | Description |
|---------|-------------|
| `/specledger.constitution` | Create or update project guiding principles |
| `/specledger.checklist` | Generate quality checklists for specs |

**Skills (auto-loaded context):**
- `sl-issue-tracking` - Issue management patterns and best practices
- `sl-audit` - Codebase reconnaissance and module discovery

### Step 3: Feature Description

Use AskUserQuestion to ask: **"What feature would you like to build? Describe it in a few sentences."**

Use their description as input for the next step.

### Step 4: Specification

Run `/specledger.specify` with the user's feature description.

After completion, briefly explain what was created and where the spec file is located.

### Step 5: Clarification

Run `/specledger.clarify` to identify and resolve any ambiguities in the specification.

Walk the user through answering clarification questions.

### Step 6: Planning

Run `/specledger.plan` to generate the implementation plan with architecture decisions, data models, and contracts.

Briefly summarize the key design decisions.

### Step 7: Task Generation

Run `/specledger.tasks` to create the actionable, dependency-ordered task list.

### Step 8: Review Pause (CRITICAL)

**STOP HERE.** Do NOT proceed to implementation without explicit user approval.

Present the generated tasks to the user and use AskUserQuestion to ask:

> **Task Review**
>
> I've generated the implementation tasks. Please review them:
> - Use `sl issue list --all` to see all tasks
> - Use `sl issue list --tree` to see the dependency graph
>
> **Would you like to proceed with implementation, or would you like to modify any tasks first?**

### Step 8.5: Verification (Recommended)

After the user approves the tasks, recommend running verification before implementation.

Use AskUserQuestion to ask:

> **Pre-Implementation Verification**
>
> Before we start coding, it's strongly recommended to run `/specledger.verify` to check that your spec, plan, and tasks are consistent and complete. At least one verify review should exist before implementation begins.
>
> **Would you like to run verification now, or skip and go straight to implementation?**

If the user chooses to verify:
- Run `/specledger.verify` to perform cross-artifact consistency analysis.
- If CRITICAL issues are found, recommend resolving them before proceeding.
- After verification completes (or if only LOW/MEDIUM issues), ask if they want to proceed to implementation.

If the user skips, proceed directly to Step 9.

### Step 9: Implementation

Only after the user has reviewed and approved the tasks:

Run `/specledger.implement` to begin executing the tasks in order.

### Step 10: Post-Implementation Checkpoint (Recommended)

After implementation completes, recommend a checkpoint review.

Use AskUserQuestion to ask:

> **Implementation Complete — Checkpoint Recommended**
>
> All tasks have been implemented. Before wrapping up, it's strongly recommended to run `/specledger.checkpoint` for a divergence review. This will compare your implementation against the plan, flag any gaps, and offer an adversarial code review.
>
> **Would you like to run a checkpoint now?**

If the user accepts:
- Run `/specledger.checkpoint` to perform the divergence review.
- If the checkpoint offers an adversarial review agent, let the user decide whether to run it.

If the user declines, summarize what was completed and note the checkpoint was skipped.

## Important Notes

- **Never skip the review pause** in Step 8. The user must always approve tasks before implementation begins.
- **Verify and checkpoint are optional but recommended.** If a user skips them during onboarding, note that they can always run `/specledger.verify` or `/specledger.checkpoint` independently later.
- If the user wants to modify tasks, help them update the specledger issues before proceeding.
- If the user wants to stop at any point, respect that and summarize what was completed.
- Each step builds on the previous one - don't skip steps unless the user explicitly asks.
