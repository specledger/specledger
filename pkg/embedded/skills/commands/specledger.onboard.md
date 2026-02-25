---
description: Guided onboarding workflow - walks through the full SpecLedger process from constitution to implementation
---

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
- Run `/specledger.audit` to analyze the existing codebase (languages, frameworks, patterns, conventions).
- Run `/specledger.constitution` to propose tailored guiding principles based on the audit findings.
- Wait for the user to review and approve the constitution before proceeding.

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
> 6. **Implement** - Execute the tasks
>
> Let's start by describing your first feature!

### Step 3: Feature Description

Ask the user: **"What feature would you like to build? Describe it in a few sentences."**

Wait for the user's response. Use their description as input for the next step.

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

Present the generated tasks to the user and ask:

> **Task Review**
>
> I've generated the implementation tasks. Please review them:
> - Use `bd list --label "spec:<feature>" --limit 10` to see all tasks
> - Use `bd dep tree --reverse <epic-id>` to see the dependency graph
>
> **Would you like to proceed with implementation, or would you like to modify any tasks first?**

Wait for the user to explicitly confirm before proceeding.

### Step 9: Implementation

Only after the user has reviewed and approved the tasks:

Run `/specledger.implement` to begin executing the tasks in order.

## Important Notes

- **Never skip the review pause** in Step 8. The user must always approve tasks before implementation begins.
- If the user wants to modify tasks, help them update the issues before proceeding.
- If the user wants to stop at any point, respect that and summarize what was completed.
- Each step builds on the previous one - don't skip steps unless the user explicitly asks.
