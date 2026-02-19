# sl Issue Tracking

## Overview

`sl issue` is the built-in issue tracker for SpecLedger. Use it for multi-session work with complex dependencies; use TodoWrite for simple single-session tasks.

## When to Use sl issue vs TodoWrite

### Use sl issue when:
- **Multi-session work** - Tasks spanning multiple compaction cycles or days
- **Complex dependencies** - Work with blockers, prerequisites, or hierarchical structure
- **Knowledge work** - Strategic documents, research, or tasks with fuzzy boundaries
- **Side quests** - Exploratory work that might pause the main task
- **Project memory** - Need to resume work after weeks away with full context

### Use TodoWrite when:
- **Single-session tasks** - Work that completes within current session
- **Linear execution** - Straightforward step-by-step tasks with no branching
- **Immediate context** - All information already in conversation
- **Simple tracking** - Just need a checklist to show progress

**Key insight**: If resuming work after 2 weeks would be difficult without sl issue, use sl issue. If the work can be picked up from a markdown skim, TodoWrite is sufficient.

## Session Start Protocol

**At session start, always check for available work:**

```bash
# Check ready-to-work issues
sl issue list --status open

# Check in-progress issues
sl issue list --status in_progress
```

**Report format:**
- "I can see X open issues: [summary]"
- "Issue Y is in_progress. Last session: [summary from notes]. Next: [from notes]. Should I continue with that?"

## Core Operations

### Essential Commands

**Create new issue:**
```bash
sl issue create --title "Fix login bug" --type bug
sl issue create --title "Add OAuth" --type feature --priority 0
sl issue create --title "Write tests" --description "Unit tests for auth module"
```

**List issues:**
```bash
sl issue list                      # All open issues
sl issue list --status in_progress # In-progress only
sl issue list --all                # All issues across all specs
sl issue list --label "phase:setup" # Filter by label
```

**Show issue details:**
```bash
sl issue show SL-abc123
```

**Update issue:**
```bash
sl issue update SL-abc123 --status in_progress
sl issue update SL-abc123 --priority 0
sl issue update SL-abc123 --notes "COMPLETED: Login endpoint. NEXT: Session middleware"
```

**Close issue:**
```bash
sl issue close SL-abc123 --reason "Implemented in PR #42"
```

### Issue Types

| Type | Description |
|------|-------------|
| `task` | Standard work item (default) |
| `bug` | Defect or problem |
| `feature` | New functionality |
| `epic` | Large work with subtasks |
| `chore` | Maintenance or cleanup |

### Priority Levels

| Priority | Description |
|----------|-------------|
| `0` | Critical (highest) |
| `1` | High |
| `2` | Normal (default) |
| `3` | Low |

### Labels

Use labels for categorization:
- `spec:<slug>` - Feature spec this issue belongs to
- `phase:<name>` - Setup, US1, polish, etc.
- `story:<id>` - User story traceability
- `component:<area>` - Mapping to plan-defined modules

## Progress Checkpointing

Update issue notes at these checkpoints:

**Critical triggers:**
- Context running low / approaching token limit
- Major milestone reached
- Hit a blocker
- Task transition or about to close issue

**Notes format:**
```
COMPLETED: Specific deliverables
IN PROGRESS: Current state + next immediate step
BLOCKERS: What's preventing progress
KEY DECISIONS: Important context or user guidance
NEXT: Immediate next action
```

**Example:**
```bash
sl issue update SL-abc123 --notes "COMPLETED: JWT auth with RS256. KEY DECISION: RS256 over HS256 per security review. IN PROGRESS: Password reset flow. BLOCKERS: Waiting on user decision for token expiry. NEXT: Implement rate limiting."
```

## Issue Lifecycle

### 1. Discovery Phase

During exploration or implementation, proactively file issues for:
- Bugs or problems discovered
- Potential improvements noticed
- Follow-up work identified
- Technical debt encountered

```bash
sl issue create --title "Found: auth doesn't handle profile permissions"
```

### 2. Execution Phase

Mark issues in_progress when starting work:
```bash
sl issue update SL-abc123 --status in_progress
```

Update throughout work and close when complete:
```bash
sl issue close SL-abc123 --reason "Implemented with tests passing"
```

### 3. Planning Phase

For complex multi-step work, structure issues with dependencies:

```bash
# Create epic
sl issue create --title "Implement user authentication" --type epic

# Create subtasks
sl issue create --title "Set up OAuth credentials" --type task
sl issue create --title "Implement authorization flow" --type task

# Link dependencies
sl issue link SL-epic blocks SL-credentials
sl issue link SL-credentials blocks SL-flow
```

## Dependency Management

**Link issues:**
```bash
sl issue link SL-abc123 blocks SL-def456  # abc123 blocks def456
sl issue link SL-abc123 related SL-xyz789 # related but not blocking
```

**View dependencies:**
```bash
sl issue show SL-abc123  # Shows linked issues
```

## Definition of Done

Before closing an issue, verify:
1. All acceptance criteria met
2. Tests pass (if applicable)
3. Documentation updated (if needed)
4. No blockers remaining

Use `sl issue show <id>` to review issue details before closing.

## Integration with TodoWrite

Both tools complement each other at different timescales:

**TodoWrite** (short-term working memory - this hour):
- Tactical execution checklist
- Marked completed as you go
- Ephemeral: Disappears when session ends

**sl issue** (long-term episodic memory - this week/month):
- Strategic objectives and context
- Key decisions and outcomes in notes field
- Persistent: Survives compaction and session boundaries

**Pattern:**
1. Session start: Read sl issue -> Create TodoWrite items for immediate actions
2. During work: Mark TodoWrite items completed as you go
3. Reach milestone: Update sl issue notes with outcomes + context
4. Session end: TodoWrite disappears, sl issue survives with enriched notes

## Troubleshooting

**If issues seem lost:**
```bash
sl issue list --all  # See all issues across specs
```

**If issue not found:**
- Use exact issue ID (case-sensitive)
- Check correct spec context (issues stored per-spec)

**Lock file issues:**
- Lock files (`.issues.jsonl.lock`) prevent concurrent access
- If stale, remove and retry
