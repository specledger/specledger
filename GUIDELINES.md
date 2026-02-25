# SpecLedger Usage Guidelines

Best practices for using SpecLedger to build features methodically with human-in-the-loop review at every stage.

## Core Workflow

```
specify → UI review → revise → plan → UI review → revise → tasks → analyze → UI review → revise → implement
```

Every stage produces artifacts that **must be reviewed in the SpecLedger UI** before proceeding. Never skip the review step.

## Stage 1: Specify

Run `/specledger.specify` with a detailed feature description.

**Prompting tips for thorough specs:**

- Provide context about who the users are and what problem you're solving
- Mention known constraints, integrations, and edge cases upfront
- Reference external specs with `deps:alias-name` or `@alias-name` syntax if the feature depends on external contracts or APIs
- If you have prior research or requirements docs, add them as dependencies first with `sl deps add`

**After specify completes:**

1. Push the branch and open the spec in the SpecLedger UI
2. Review the spec with your team — leave comments on unclear areas, missing requirements, or scope concerns
3. Run `/specledger.revise` to fetch and address all review comments with AI assistance
4. Repeat review/revise until the spec is solid

## Stage 2: Plan

Run `/specledger.plan` with your tech stack context (e.g., "I am building with Go and PostgreSQL").

**Prompting tips for thorough research:**

- Tell the AI to research specific areas: "Research best practices for X", "Investigate tradeoffs between A and B"
- The plan phase spawns research agents — give it enough context to investigate properly
- If the plan's `research.md` feels shallow, ask follow-up questions before accepting
- Check that the data model, contracts, and architecture decisions align with your spec

**After plan completes:**

1. Push and review `plan.md`, `research.md`, `data-model.md`, and `contracts/` in the UI
2. Pay special attention to architecture decisions and tech stack choices
3. Leave comments on anything that seems off or needs deeper investigation
4. Run `/specledger.revise` to address review comments
5. Repeat until the plan is approved

## Stage 3: Generate Tasks

Run `/specledger.tasks` to generate the task breakdown.

**Important: Generate tasks cleanly.** The task generation should produce a clean `tasks.md` file with dependency-ordered tasks. Do not use `sl issue` commands during task generation — the tasks file is the source of truth at this stage.

## Stage 4: Analyze (Before Implementation)

**Always run `/specledger.analyze` after task generation and before implementation.** This is a critical quality gate.

The analyze command performs a read-only cross-artifact consistency check across `spec.md`, `plan.md`, and `tasks.md`. It will identify:

- Requirements with no associated tasks (coverage gaps)
- Tasks with no mapped requirement (orphan tasks)
- Terminology drift between artifacts
- Ambiguous or vague acceptance criteria
- Constitution violations
- Conflicting requirements or task ordering issues

**How to act on the analysis:**

- **CRITICAL issues** — must be resolved before proceeding to implementation
- **HIGH issues** — should be resolved; skip only with explicit justification
- **MEDIUM/LOW issues** — resolve if time permits; document as known gaps if skipping

If the analysis reveals gaps, update the relevant artifacts (spec, plan, or tasks) and re-run `/specledger.analyze` to confirm fixes.

## Stage 5: Final Review Before Implementation

After analyze passes cleanly:

1. Push all artifacts and review the complete task list in the UI
2. Verify task ordering, dependencies, and acceptance criteria make sense
3. Leave comments on tasks that need refinement
4. Run `/specledger.revise` to address final comments
5. Only proceed to implementation when the team has approved

## Stage 6: Implement

Run `/specledger.implement` to execute the task plan.

Implementation follows the task order, respects dependencies, and tracks progress through `sl issue` commands.

## Using `/specledger.revise`

The revise command is your bridge between UI reviews and AI-assisted fixes.

**Workflow:**

1. Team reviews artifacts in the SpecLedger UI and leaves comments
2. Run `/specledger.revise` in your coding agent
3. The command fetches all unresolved comments from Supabase
4. For each comment, the AI analyzes the feedback, proposes changes, and asks for your confirmation
5. After approval, edits are applied and comments are marked as resolved
6. Push the updated artifacts for another round of review if needed

**Tips:**

- Run `/specledger.revise` at the start of each session to catch new comments
- Use `/specledger.revise --resolve <id>` to address a specific comment
- Use `/specledger.revise --post -f <file> -m <message>` to leave comments from the CLI

## Prompting for Thorough Research

The quality of SpecLedger output depends heavily on the prompts you provide. Here are patterns for getting deeper research:

**During specify:**

```
/specledger.specify Build a real-time collaboration system for document editing.
Users need to see each other's cursors and edits within 100ms.
Must handle offline editing and conflict resolution.
Consider CRDT vs OT approaches. Must integrate with our existing auth system.
```

**During plan:**

```
/specledger.plan I am building with TypeScript, Next.js, and PostgreSQL.
Research CRDT libraries compatible with our stack.
Investigate Yjs vs Automerge for our use case.
Consider WebSocket vs SSE for real-time sync.
```

The more specific your constraints and questions, the deeper the research agents will dig.

## Quick Reference

| Stage     | Command                 | Review After?       | Key Check                    |
| --------- | ----------------------- | ------------------- | ---------------------------- |
| Specify   | `/specledger.specify`   | Yes - UI review     | Spec covers all requirements |
| Clarify   | `/specledger.clarify`   | Optional            | Ambiguities resolved         |
| Plan      | `/specledger.plan`      | Yes - UI review     | Architecture decisions sound |
| Tasks     | `/specledger.tasks`     | After analyze       | Tasks are dependency-ordered |
| Analyze   | `/specledger.analyze`   | Yes - UI review     | No CRITICAL gaps             |
| Revise    | `/specledger.revise`    | After each cycle    | All comments addressed       |
| Implement | `/specledger.implement` | Post-implementation | Features match spec          |

## Anti-Patterns

- **Skipping UI review** — Running specify-plan-tasks-implement without human review produces misaligned features
- **Skipping analyze** — Going straight from tasks to implement misses coverage gaps and inconsistencies
- **Vague prompts** — "Build a login page" produces a shallow spec; provide constraints, edge cases, and user context
- **Ignoring CRITICAL findings** — Analyze flags them for a reason; resolve before implementing
- **Using issue tracking during task generation** — Keep task generation clean; issue tracking starts during implementation
