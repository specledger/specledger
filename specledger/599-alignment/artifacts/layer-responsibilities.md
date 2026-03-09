# SDD Layer Responsibilities

**Purpose**: Decision criteria for placing functionality in the correct layer
**Reference**: Extends CLI Constitution from 598-sdd-workflow-streamline

## Quick Reference

| Layer | Name | One-Liner | Owns |
|-------|------|-----------|------|
| L0 | Hooks | "Invisible automation" | Git event responses |
| L1 | CLI | "Data operations" | CRUD, standalone tooling |
| L2 | AI Commands | "Workflow orchestration" | Multi-step AI workflows |
| L3 | Skills | "Domain knowledge" | Patterns, decision criteria |

## L0 - Hooks

**Runtime**: Invisible, event-driven
**Purpose**: Auto-capture sessions on commit

### Use When

- Action should happen automatically without user/agent initiation
- Response to git events (commit, push, merge, branch creation)
- No decision logic needed - pure automation
- Side effects are always desirable (e.g., session capture)

### Don't Use When

- Action requires user input or confirmation
- Action needs AI reasoning or content generation
- Action should be user-initiated (put in CLI)
- Different behavior needed based on context

### Examples

| Good | Bad |
|------|-----|
| Auto-capture session on commit | Prompt user for commit message |
| Validate commit message format | Generate commit message |
| Update context files on branch switch | Ask user which files to update |

### Anti-Patterns

- Hooks that require network access (slow down git operations)
- Hooks that prompt for user input
- Hooks that fail silently without recovery

---

## L1 - CLI (`sl`)

**Runtime**: Go binary, no AI needed
**Purpose**: Data operations, CRUD, standalone tooling

### Use When

- CRUD operations on data (create, read, update, delete)
- Standalone tooling that works without AI agent
- Cross-platform compatibility needed (macOS, Linux, Windows)
- Output should be machine-parseable (JSON)
- Operations are deterministic and repeatable
- No PTY required (agent-friendly)

### Don't Use When

- Complex multi-step orchestration with decision points
- AI reasoning required between operations
- Content generation (specs, plans, code)
- Workflow involves interactive user choices

### CLI Constitution (from 598)

| ID | Principle |
|----|-----------|
| D16 | Data CRUD pattern - list/show/create/update/delete |
| D21 | Token-efficient output - compact lists, full show |
| JSON | `--json` flag for all commands |
| Offline | Most commands work offline (except API calls) |

### Examples

| Good | Bad |
|------|-----|
| `sl issue create --title "Bug"` | AI command that creates issues with reasoning |
| `sl spec info --json` | AI command that parses spec manually |
| `sl comment list --status open` | AI command that fetches and formats comments |
| `sl doctor --check` | Interactive wizard in CLI |

### Anti-Patterns

- CLI commands that require PTY (breaks agent usage)
- Verbose output without `--json` alternative
- Business logic that belongs in AI orchestration
- Commands that generate content (put in AI command)

---

## L2 - AI Commands

**Runtime**: Agent shell prompts (Claude, etc.)
**Purpose**: AI workflow orchestration (specify→implement)

### Use When

- Multi-step workflow with decision points
- AI reasoning needed between operations
- Orchestrating multiple CLI calls with context
- Generating content (specs, plans, code, documentation)
- Workflow state needs to be tracked across steps
- User interaction through natural language

### Don't Use When

- Single data operation (use CLI)
- No AI reasoning needed - pure automation
- Operation is deterministic (use CLI)
- Action should happen automatically (use hooks)

### Standard Structure

```markdown
---
description: [One-line description]
handoffs:
  - label: [Handoff label]
    agent: [Target agent]
    prompt: [Prompt template]
---

## Purpose
[What this command does]

## When to Use
[Scenarios]

## Outline
1. [Step with CLI calls]
2. [Step with AI reasoning]

## Behavior Rules
[Constraints and error handling]
```

### Examples

| Good | Bad |
|------|-----|
| `/specledger.specify` - generates spec from description | AI command that just calls `sl issue create` |
| `/specledger.plan` - researches and creates implementation plan | AI command that formats CLI output |
| `/specledger.clarify` - asks questions, updates spec | AI command that does single file read |

### Anti-Patterns

- AI commands that duplicate CLI functionality
- Business logic in AI command that should be in CLI
- Parsing human-readable CLI output (use `--json`)
- Embedding extensive CLI syntax (link to `--help`)

---

## L3 - Skills

**Runtime**: Passive context injection
**Purpose**: Domain knowledge, progressively loaded

### Use When

- Teaching patterns and best practices
- Providing decision criteria (when to use X vs Y)
- Context that should load progressively (not always present)
- Domain-specific knowledge that applies across commands
- Reference material for AI reasoning

### Don't Use When

- Executing operations (use CLI or AI command)
- Duplicate of CLI help text
- Static configuration (use templates)
- Information that's only relevant to one command

### Standard Structure

```markdown
# [Skill Name]

## Overview
[What this skill provides]

## When to Use [vs Alternative]
[Decision criteria]

## Key Concepts
[Domain knowledge]

## Decision Patterns
[When to use which approach]

## CLI Reference
[Links to --help, minimal syntax examples]

## Patterns
[Best practices and examples]
```

### Examples

| Good | Bad |
|------|-----|
| sl-issue-tracking: when to use sl issue vs TodoWrite | Skill that just lists `sl issue` commands |
| Skill explaining SDD workflow phases | Skill that duplicates spec.md content |
| Skill with decision flowchart for auth methods | Skill with exhaustive CLI syntax reference |

### Anti-Patterns

- Duplicating CLI `--help` content
- Information that belongs in a template
- Content specific to a single command
- Syntax reference without decision guidance

---

## Decision Flowchart

```
START: New functionality needed
  │
  ├─ Should it happen automatically on git events?
  │   └─ YES → L0 (Hooks)
  │
  ├─ Is it a CRUD operation or standalone tool?
  │   └─ YES → L1 (CLI)
  │
  ├─ Does it need AI reasoning or content generation?
  │   └─ YES → L2 (AI Command)
  │
  ├─ Is it reference knowledge or patterns?
  │   └─ YES → L3 (Skill)
  │
  └─ None of the above? Re-evaluate requirements.
```

## Exceptions and Edge Cases

### Consolidated Commands (from 598)

The 598-sdd-workflow-streamline spec consolidates AI commands:

| Command | Layer | Decision | Rationale |
|---------|-------|----------|-----------|
| `add-deps` / `remove-deps` | L2 → L1 | Remove AI command | Agent calls `sl deps` CLI directly |
| `audit` | L2 → L3 | Convert to skill | Codebase reconnaissance is passive context |
| `revise` | L2 → L2 | Absorb into `clarify` | Comment processing belongs with spec refinement |
| `resume` | L2 | Remove | Duplicate of `implement` |
| `help` | L2 | Remove | Absorbed by `onboard` |
| `adopt` | L2 | Remove | Context detection fallback chain replaces it |
| `analyze` | L2 | Rename to `verify` | OpenSpec terminology alignment |

### Cross-Layer Convenience Patterns

Some patterns span layers for convenience:

| Pattern | Example | Justification |
|---------|---------|---------------|
| L1→L0 | `sl auth hook --install` | CLI configures hooks |
| L1→L2 | `sl revise` (launcher) | CLI generates prompt, launches agent |
| L2→L1 | AI commands call CLI | AI orchestrates CLI operations |

### When Business Logic Can't Move to CLI

If an AI command has business logic that truly can't move to CLI:
1. Document the exception in the command file
2. Explain why CLI is insufficient
3. Consider if CLI extension is possible in future

### When Skills Need CLI Syntax

If a skill needs CLI syntax examples for clarity:
1. Keep examples minimal (1-2 lines)
2. Link to `sl <command> --help` for full syntax
3. Focus on decision criteria, not exhaustive reference
