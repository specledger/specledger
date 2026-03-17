# Skills Design Principles (Layer 3)

Skills are passive context injection — domain knowledge loaded into the agent's context when triggered by relevant references in commands (L2) or user conversation.

> **Building skills**: For creating, editing, and testing skills, use the `skill-creator` skill at [`.claude/skills/skill-creator/`](../../.claude/skills/skill-creator/). This document covers design principles and architecture, not implementation mechanics.

---

## Rules of Thumb

1. **Skills teach, CLI does** — skills explain *when* and *how* to use CLI tools, never duplicate their logic
2. **Progressive loading** — skills load on-demand when triggered, not upfront
3. **One CLI domain per skill** — `sl-comment` skill for `sl comment`, `sl-issue-tracking` for `sl issue`
4. **Lean and isolated** — each skill is a single markdown file with focused context
5. **Include decision criteria** — help the agent choose between subcommands, not just list them

---

## Core Principle: Complementary, Not Redundant

Skills exist at Layer 3 because they provide *passive context* — domain knowledge that helps the agent make better decisions. They do NOT orchestrate workflows (that's L2 commands) or perform operations (that's L1 CLI).

```
Layer 1 (CLI):   sl comment list --json         ← Does the work
Layer 2 (Cmd):   /specledger.clarify             ← Orchestrates when to call it
Layer 3 (Skill): sl-comment skill                ← Teaches the agent how and why
```

**Trigger pattern**: AI commands reference CLI tools briefly → agent shell detects the reference → loads the corresponding skill → agent has full usage context.

---

## Skill Anatomy

Every skill follows this structure:

```markdown
# {skill-name} Skill

**When to Load**: Triggered when [context description]

## Overview
What this skill provides.

## Subcommands
| Command | Purpose | Output Mode |

## Decision Criteria
### When to Use X vs Y
Help the agent choose between options.

## JSON Parsing Examples
Show the agent how to parse CLI output.

## Workflow Patterns
Common multi-step sequences.

## Error Handling
| Error | Cause | Solution |

## Token Efficiency
How the CLI minimizes context usage.
```

### Required Sections

| Section | Purpose | Why Required |
|---------|---------|-------------|
| Overview | What the skill covers | Agent needs to know if the skill is relevant |
| Subcommands | Available CLI commands | Quick reference table |
| Decision Criteria | When to use which command | Prevents wrong tool choice |
| JSON Parsing Examples | How to extract data from `--json` output | Agents need concrete patterns |
| Workflow Patterns | Multi-step sequences | Shows the intended usage flow |
| Error Handling | Common errors and fixes | Aligns with CLI error-as-navigation principle |
| Token Efficiency | Output budget notes | Helps agent choose compact vs full output |

---

## Skill Inventory

| Skill | CLI Domain | Triggers On |
|-------|-----------|-------------|
| `sl-comment` | `sl comment` (list/show/reply/resolve) | "review comments", `sl comment` references |
| `sl-issue-tracking` | `sl issue` (create/list/close/update) | "issues", "tasks", `sl issue` references |
| `sl-deps` | `sl deps` (add/remove/graph) | "dependencies", `sl deps` references |
| `sl-audit` | Codebase reconnaissance | Tech stack discovery, module analysis |

---

## Progressive Loading

Skills are NOT loaded at session start. They are injected when the agent encounters a relevant trigger:

1. AI command says: "fetch open comments using `sl comment list`"
2. Agent shell detects `sl comment` reference
3. `sl-comment` skill is loaded into context
4. Agent now knows: subcommands, JSON format, workflow patterns, error handling

This keeps the agent's initial context lean. A session that never touches comments never loads the comment skill.

---

## Cross-Layer Alignment

Skills must stay synchronized with CLI changes:

| When CLI changes... | Skill must... |
|---|---|
| New subcommand added | Add to subcommands table + decision criteria |
| Flag added/changed (e.g., `--reason`) | Update workflow patterns + examples |
| Output format changed | Update JSON parsing examples |
| Error message changed | Update error handling table |

**Template lifecycle**: Skills are embedded in the `sl` binary (`pkg/embedded/templates/specledger/skills/`) and distributed via `sl doctor --template`. Both the embedded source and runtime copy (`.claude/skills/`) must stay in sync.

---

## Anti-Patterns

### AP-01: Skill duplicates CLI help
```markdown
# Wrong: re-documents all flags
The `--status` flag accepts: open, resolved, all. The `--json` flag outputs JSON...

# Right: teaches decision-making
Use `--status open` when scanning for actionable feedback.
Use `sl comment show <id>` for full thread context before replying.
```

### AP-02: Skill includes implementation logic
```markdown
# Wrong: skill contains code patterns
Parse the JSON response and extract the `change_id` field, then POST to...

# Right: skill teaches workflow
1. List comments → identify actionable ones
2. Show specific comment → understand full context
3. Reply with resolution → resolve with reason
```

---

## References

- [4-Layer Model Overview](README.md)
- [CLI Design](cli.md) — L1 tools that skills document
- [Commands Design](commands.md) — L2 commands that trigger skill loading
- [Skill Creator](../../.claude/skills/skill-creator/) — Tool for building and testing skills
