# Cross-Layer Interactions

**Purpose**: Document how SDD layers interact with examples
**Reference**: 598-sdd-workflow-streamline plan

## Interaction Patterns Overview

```
┌─────────────────────────────────────────────────────────────┐
│                      L2: AI Commands                        │
│  (Orchestrates workflow, calls L1, parses output)          │
│                         │                                   │
│         ┌───────────────┼───────────────┐                  │
│         ▼               ▼               ▼                  │
│    ┌─────────┐    ┌─────────┐    ┌─────────┐              │
│    │ sl issue│    │sl comment│   │ sl spec │              │
│    └────┬────┘    └────┬────┘    └────┬────┘              │
└─────────┼──────────────┼──────────────┼───────────────────┘
          │              │              │
          ▼              ▼              ▼
┌─────────────────────────────────────────────────────────────┐
│                      L1: CLI (sl)                           │
│  (Data operations, CRUD, JSON output)                      │
│                         │                                   │
│                    ┌────┴────┐                             │
│                    ▼         ▼                             │
│              ┌─────────┐ ┌─────────┐                       │
│              │git hooks│ │ config  │                       │
│              └────┬────┘ └─────────┘                       │
└───────────────────┼─────────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────────────────────┐
│                      L0: Hooks                              │
│  (Event-driven automation)                                 │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                      L3: Skills                            │
│  (Passive context injection, loaded on demand)             │
└─────────────────────────────────────────────────────────────┘
```

---

## L2 → L1: Launcher Pattern

AI commands invoke CLI for data operations, then parse the output.

### Pattern

```markdown
## Outline

1. **Fetch data using CLI**:
   \```bash
   sl <command> list --json
   \```

2. **Parse JSON output**:
   The AI automatically parses the JSON structure. Example output:
   \```json
   [
     {"id": "SL-abc123", "title": "Feature", "status": "open"}
   ]
   \```

3. **Process and act**:
   Based on the data, the AI takes appropriate action.
```

### Example: /specledger.implement

From `.claude/commands/specledger.implement.md`:

```markdown
## Core Operations

| Action | Command |
|--------|---------|
| Find ready tasks | `sl issue ready` |
| List open issues | `sl issue list --status open` |
| Show issue details | `sl issue show <id>` |
| Update issue | `sl issue update <id> --status in_progress` |
| Close issue | `sl issue close <id> --reason "Completed"` |
```

The AI command orchestrates multiple CLI calls:
1. `sl issue ready` → Find work
2. `sl issue show <id>` → Get details
3. `sl issue update <id> --status in_progress` → Claim work
4. [Implement the task]
5. `sl issue close <id> --reason "..."` → Complete work

### Token Efficiency (D21 from 598)

CLI commands follow token-efficient output patterns:

| Command | Output | Use Case |
|---------|--------|----------|
| `sl <cmd> list` | Compact overview | Scan many items |
| `sl <cmd> show <id>` | Full details | Drill into one item |

```bash
# Compact: list returns truncated previews
sl comment list --status open
# Output: 5 comments, IDs: SL-c1 SL-c2 SL-c3 SL-c4 SL-c5

# Full: show returns complete content
sl comment show SL-c1 --json
# Output: {"id":"SL-c1", "content": "...", "thread": [...]}
```

---

## L3: Skill Loading

Skills are loaded progressively when relevant commands are referenced.

### Trigger Patterns

| Trigger | Skill Loaded |
|---------|-------------|
| Reference to `sl issue` | sl-issue-tracking |
| Reference to `sl comment` | sl-comment (after 598) |
| Reference to `sl spec` | sl-spec (after 598) |

### Example: sl-issue-tracking

When an AI command references `sl issue`, the skill provides:

```markdown
## When to Use sl issue vs TodoWrite

### Use sl issue when:
- Multi-session work
- Complex dependencies
- Knowledge work

### Use TodoWrite when:
- Single-session tasks
- Linear execution
```

This context helps the AI make better decisions about which tool to use.

### Skill Content Focus

Skills should provide:
- ✅ Decision criteria (when to use)
- ✅ Domain concepts
- ✅ Workflow patterns
- ❌ Exhaustive CLI syntax (link to `--help`)

---

## L1 → L0: Hook Configuration

CLI commands can configure hooks for convenience.

### Pattern

```bash
# CLI command that configures hooks
sl auth hook --install
```

This convenience pattern allows users to configure L0 (hooks) through L1 (CLI) rather than manually editing `.git/hooks/`.

### Current Hooks

| Hook | Purpose | Trigger |
|------|---------|---------|
| post-commit | Session capture | After each commit |

### Implementation

From 598 plan:
> L1→L0: `sl auth hook --install` configures hooks (convenience pattern)

---

## L1 → L2: Launcher Pattern (Reverse)

Some CLI commands launch AI agents.

### Pattern

```bash
# CLI generates prompt and launches agent
sl revise
```

From 598 plan:
> L1→L2: `sl revise` generates a prompt and launches an agent session (launcher pattern)

### How It Works

1. CLI command runs
2. CLI generates a prompt based on current context
3. CLI launches AI agent with the prompt
4. Agent takes over execution

This is a convenience pattern for starting AI workflows from the CLI.

---

## Interaction Decision Matrix

| From | To | Pattern | Example |
|------|-----|---------|---------|
| L2 | L1 | Call + parse | AI calls `sl issue list --json` |
| L2 | L3 | Reference trigger | AI mentions `sl issue`, skill loads |
| L1 | L0 | Configure | `sl auth hook --install` |
| L1 | L2 | Launch | `sl revise` generates prompt, launches agent |
| L0 | (none) | Autonomous | Hooks run without interaction |

---

## Best Practices

### For AI Command Authors

1. **Always use `--json`** for structured data
2. **Don't parse human output** - it changes format
3. **Document expected JSON structure** in command
4. **Handle CLI errors gracefully** - check stderr

### For Skill Authors

1. **Focus on decisions** not syntax
2. **Link to `--help`** for full CLI reference
3. **Provide comparison criteria** vs alternatives
4. **Include workflow patterns** not just commands

### For CLI Authors

1. **Always provide `--json`** output
2. **Compact by default** (D21), full on show
3. **No PTY required** - agent-friendly
4. **Errors to stderr**, data to stdout

---

## Code Examples

### AI Command Calling CLI

```markdown
## Outline

1. **Get feature context**:
   \```bash
   sl spec info --json --include-tasks
   \```

   Parse FEATURE_DIR, FEATURE_SPEC, AVAILABLE_DOCS from JSON.

2. **Load current tasks**:
   \```bash
   sl issue list --status open --label "spec:<slug>"
   \```

3. **Process each task**:
   For each task ID from the list:
   \```bash
   sl issue show <id> --json
   \```
```

### Skill Reference in Command

```markdown
## Purpose

Execute tasks using `sl issue` for tracking.

> See sl-issue-tracking skill for when to use sl issue vs TodoWrite.
```

### Error Handling

```markdown
## Behavior Rules

- If `sl issue show <id>` fails, check if issue exists
- If auth error, guide user to `sl auth login`
- If network error, retry once then report
```
