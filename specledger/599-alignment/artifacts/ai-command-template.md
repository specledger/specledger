# AI Command Template

**Purpose**: Standard structure for `.claude/commands/*.md` files
**Version**: 1.0 | **Date**: 2026-03-02

## Required Structure

```markdown
---
description: [One-line description shown in command list]
handoffs:
  - label: [UI button label]
    agent: [Target agent name]
    prompt: [Prompt template for handoff]
    send: [true/false - auto-send or require user action]
---

## User Input

\```text
$ARGUMENTS
\```

You **MUST** consider the user input before proceeding (if not empty).

## Purpose

[1-2 sentences explaining what this command does and its value]

**When to use**: [Brief guidance on when this command is appropriate]

## Outline

1. **[Phase/Step Name]**:
   [Description with specific actions]

   \```bash
   [CLI command with --json flag if applicable]
   \```

2. **[Next Step]**:
   [Continue numbered steps...]

## Behavior Rules

- [Constraint 1: what the AI must do]
- [Constraint 2: what the AI must NOT do]
- [Error handling approach]
- [User interaction guidelines]
```

## Section Guidelines

### Frontmatter

| Field | Required | Description |
|-------|----------|-------------|
| `description` | Yes | One-line description for command list |
| `handoffs` | No | List of handoff options to other agents |
| `handoffs[].label` | If handoffs | UI button label |
| `handoffs[].agent` | If handoffs | Target agent identifier |
| `handoffs[].prompt` | If handoffs | Prompt template (can use $ARGUMENTS) |
| `handoffs[].send` | No | Auto-send (true) or require user action (false/default) |

### Purpose Section

- Keep to 1-2 sentences
- Focus on WHAT the command accomplishes
- Include "When to use" as a quick reference

### Outline Section

- Use numbered steps for sequential execution
- Include bash commands in code blocks
- Mark optional steps with "(optional)" or "(if applicable)"
- Use `--json` for CLI commands that return structured data
- Document expected output format

### Behavior Rules Section

Common patterns:
- Error handling: "If X fails, do Y"
- User interaction: "Ask user before Z"
- Constraints: "Never do X", "Always Y"
- Termination: "Stop if Z occurs"

## CLI Invocation Patterns

### Standard Pattern

```bash
# Get structured data
sl <command> <subcommand> --json

# Parse in AI context
# JSON is automatically parsed, no need for jq
```

### Error Handling

```markdown
If the command fails:
1. Check error message in stderr
2. If auth error, guide user to `sl auth login`
3. If not found, verify the resource exists
4. Report error to user with context
```

### Token Efficiency

- Use `--json` for structured output (more parseable)
- Use list commands for overviews, show commands for details
- Don't request full data if summary suffices

## Example: Good Command

```markdown
---
description: Create or update the feature specification from a description
handoffs:
  - label: Build Technical Plan
    agent: specledger.plan
    prompt: Create a plan for the spec. I am building with...
---

## User Input

\```text
$ARGUMENTS
\```

## Purpose

Create feature specifications from natural language descriptions.

**When to use**: At the beginning of any new feature development.

## Outline

1. **Generate branch name**:
   Analyze the feature description and create a 2-4 word short name.

2. **Check current state**:
   \```bash
   git rev-parse --abbrev-ref HEAD
   \```

3. **Create feature branch**:
   \```bash
   .specledger/scripts/bash/create-new-feature.sh --json --number N --short-name "name"
   \```

## Behavior Rules

- Never create a feature branch if already on one
- Always check for existing branches with same short-name
- Maximum 3 [NEEDS CLARIFICATION] markers
- Stop if user denies branch creation
```

## Example: Anti-Pattern

```markdown
## Outline

1. Parse the spec file and extract all requirements:
   First, read the file line by line, then for each line,
   check if it starts with "FR-", then extract the text...

   [Too much business logic - should be in CLI]

2. Format the output in a pretty table:
   [Parsing/formatting logic that duplicates CLI]

## Behavior Rules

- Generate comprehensive documentation for every field
  [Too vague - should be specific]
```

## Validation Checklist

- [ ] Description is one line, under 80 characters
- [ ] Purpose section is 1-2 sentences
- [ ] Outline has numbered steps
- [ ] CLI commands use `--json` for structured data
- [ ] Behavior rules are specific and actionable
- [ ] No business logic that belongs in CLI
- [ ] Error handling is documented
