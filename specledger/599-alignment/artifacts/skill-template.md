# Skill Template

**Purpose**: Standard structure for skill markdown files
**Version**: 1.0 | **Date**: 2026-03-02

## What is a Skill?

Skills are passive context injection documents that provide domain knowledge to AI agents. They:
- Load progressively when relevant commands are referenced
- Teach patterns, decision criteria, and best practices
- Complement CLI `--help` without duplicating it
- Focus on "when to use" rather than "how to use"

## Required Structure

```markdown
# [Skill Name]

## Overview

[1-2 sentences explaining what this skill provides and its scope]

## When to Use [Skill Name] vs [Alternative]

### Use [Skill Name] when:
- [Condition 1]
- [Condition 2]
- [Condition 3]

### Use [Alternative] when:
- [Condition 1]
- [Condition 2]

**Key insight**: [One-line decision criterion]

## Key Concepts

### [Concept 1]
[Explanation with examples]

### [Concept 2]
[Explanation with examples]

## Decision Patterns

### [Pattern 1: When to do X]
**Scenario**: [Description]
**Approach**: [What to do]
**Example**:
\```bash
[Minimal CLI example]
\```

### [Pattern 2: When to do Y]
[Continue patterns...]

## CLI Reference

### Essential Commands

| Action | Command |
|--------|---------|
| [Action 1] | `sl <cmd> <subcmd> [flags]` |
| [Action 2] | `sl <cmd> <subcmd> [flags]` |

> **Full syntax**: See `sl <command> --help` for complete flag reference.

### Output Formats

| Flag | Use When |
|------|----------|
| `--json` | Programmatic parsing needed |
| (default) | Human-readable output |

## Patterns

### [Pattern Name]

[Description of the pattern with context]

\```bash
# Example workflow
sl <cmd> list --json
# Parse results, then:
sl <cmd> show <id> --json
\```

## Troubleshooting

**If [problem]**:
\```bash
[Diagnostic command]
\```

**If [error]**:
- [Solution 1]
- [Solution 2]
```

## Section Guidelines

### Overview

- Keep to 1-2 sentences
- Define the scope clearly
- Explain the value proposition

### When to Use

- Compare with alternatives (other skills, CLI, TodoWrite, etc.)
- Provide clear decision criteria
- Include a "key insight" one-liner

### Key Concepts

- Domain-specific knowledge
- Terms and definitions
- Mental models for understanding

### Decision Patterns

- Scenario-based guidance
- When to use which approach
- Minimal CLI examples (1-2 lines max)

### CLI Reference

- Table format for quick reference
- Link to `--help` for full syntax
- Don't duplicate all flags and options

### Patterns

- Workflow examples
- Common sequences
- Integration with other tools

## Example: Good Skill (sl-issue-tracking excerpt)

```markdown
# sl Issue Tracking

## Overview

`sl issue` is the built-in issue tracker for SpecLedger. Use it for
multi-session work with complex dependencies; use TodoWrite for simple
single-session tasks.

## When to Use sl issue vs TodoWrite

### Use sl issue when:
- **Multi-session work** - Tasks spanning multiple compaction cycles
- **Complex dependencies** - Work with blockers or prerequisites
- **Knowledge work** - Strategic documents or research

### Use TodoWrite when:
- **Single-session tasks** - Work that completes within current session
- **Linear execution** - Straightforward step-by-step tasks

**Key insight**: If resuming work after 2 weeks would be difficult without
sl issue, use sl issue.

## CLI Reference

### Essential Commands

| Action | Command |
|--------|---------|
| Create issue | `sl issue create --title "..." --type task` |
| List issues | `sl issue list --status open` |
| Show details | `sl issue show <id>` |
| Update issue | `sl issue update <id> --status in_progress` |

> **Full syntax**: See `sl issue --help` for complete flag reference.
```

## Example: Anti-Pattern

```markdown
# sl Issue Tracking

## Commands

### sl issue create

Create a new issue with the following syntax:

\```bash
sl issue create --title "TITLE" [--description "DESC"] [--type TYPE]
  [--priority PRIORITY] [--labels LABELS] [--parent PARENT_ID]
\```

Flags:
- `--title`: Required. The title of the issue (max 200 chars)
- `--description`: Optional. Full description (max 10000 chars)
- `--type`: One of: task, bug, feature, epic, chore (default: task)
[... exhaustive flag documentation ...]
```

**Why this is bad**: Duplicates `--help` content, no decision guidance, no context.

## Focus Guidelines

### DO Include

- Decision criteria (when to use X vs Y)
- Domain concepts and mental models
- Workflow patterns and sequences
- Integration with other tools
- Troubleshooting guidance
- Minimal CLI examples

### DON'T Include

- Exhaustive flag documentation (link to `--help`)
- Every possible command variation
- Content specific to one AI command
- Static configuration (use templates)
- Information that will quickly become outdated

## Validation Checklist

- [ ] Overview is 1-2 sentences
- [ ] "When to Use" section compares alternatives
- [ ] Key insight provides decision criterion
- [ ] CLI examples are minimal (1-2 lines)
- [ ] Links to `--help` for full syntax
- [ ] Focus on patterns, not syntax reference
- [ ] Troubleshooting section included
