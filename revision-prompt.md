You are assisting with document revision for spec "127-specledger-scheduler-push-strategy".

## Artifacts to Revise
The following files contain comments that need to be addressed:
- specledger/127-specledger-scheduler-push-strategy/research.md
- specledger/127-specledger-scheduler-push-strategy/plan.md
- specledger/127-specledger-scheduler-push-strategy/plan.md

You have full access to read and edit these files in the workspace.

## Revision Strategy
Before making any edits:
1. Read ALL comments below
2. Identify thematic clusters — comments that share intent, topic, or require coordinated changes across artifacts
3. For each cluster, present a SINGLE AskUserQuestion covering all related comments with 2-3 distinct revision approaches
4. Apply coordinated edits across all impacted artifacts for the chosen approach
5. Track every option proposed and choices made in a dedicated revision log, create if it doesn't exist yet.

## Comments to Address

### Comment 1
- **File**: specledger/127-specledger-scheduler-push-strategy/research.md
- **Target**: "4. Execution Lock Strategy
Decision: PID-based lock file at .specledger/exec.lock using gofrs/flock (already a dependency).

Rationale:

gofrs/flock already used in issue store for JSONL locking

L..."
- **Feedback**: "Maybe we should have a command to let user reset/delete the lock file in cases where the file isn't delete properly"

### Comment 2
- **File**: specledger/127-specledger-scheduler-push-strategy/plan.md
- **Target**: "Phase 3: Push-Triggered Execution (P1 - US1)"
- **Feedback**: "Missing: error handling strategy for edge cases (missing sl binary, malformed spec, rejected push). Spec requires graceful failure (FR-011)."

### Comment 3
- **File**: specledger/127-specledger-scheduler-push-strategy/plan.md
- **Target**: "Execution lock management in pkg/cli/scheduler/lock.go"
- **Feedback**: "Underspecified: stale lock detection flow (FR-015). When/how is staleness checked? What's the removal + retry logic?"

## Important Instructions
- ALWAYS use AskUserQuestion before making any edit
- Group related comments together — do NOT address them one by one if they share a theme
- Present clear, distinct options for each thematic cluster
- When a thread has replies with additional context, factor that into your suggestions
- Apply edits incrementally, one cluster at a time across all impacted artifacts
- After all edits, summarize what was changed and which comments were addressed
- Do NOT modify files beyond what the comments request

Begin by reading all comments, identifying clusters, then processing the first cluster.
