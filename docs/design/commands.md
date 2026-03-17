# AI Commands Design Principles (Layer 2)

AI commands (slash commands) are agent shell prompts that orchestrate multi-step workflows. They live in `.claude/commands/specledger.*.md` and are the primary interface for specification-driven development.

---

## Rules of Thumb

1. **Commands orchestrate, CLI executes** — commands call `sl` CLI tools for data operations; they don't duplicate CLI logic
2. **Core workflow is immutable** — specify → clarify → plan → tasks → implement. Commands customize behavior within stages, not stage order
3. **One command, one workflow stage** — each command maps to exactly one stage in the pipeline (or an escape hatch like spike/checkpoint)
4. **Commands reference CLI tools briefly** — this triggers skill loading (L3) for full usage context
5. **Use `sl spec info --json` for context** — don't re-implement branch detection or path resolution

---

## Core Workflow Pipeline

The specify → implement pipeline is immutable. Commands customize content within each stage.

```
specify → clarify → plan → tasks → implement
   │         │        │       │         │
   │         │        │       │         └─ Execute tasks, resume from checkpoint
   │         │        │       └─ Generate dependency-ordered tasks.md
   │         │        └─ Create implementation plan from spec
   │         └─ Resolve ambiguities + process review comments
   └─ Create feature spec from description
```

**Escape hatches** (usable at any stage):
- `spike` — time-boxed exploratory research
- `checkpoint` — verify implementation progress + session log
- `verify` — cross-artifact consistency check
- `checklist` — custom per-feature quality gates
- `constitution` — project principles (setup-time)
- `onboard` — guided walkthrough (setup-time)

---

## Command Anatomy

Every command follows the same structure:

```markdown
---
description: One-line purpose
handoffs:
  - label: Next Stage
    agent: specledger.next
    prompt: Proceed to next stage
---

## User Input
$ARGUMENTS

## Purpose
What this command does and when to use it.

## Outline
Step-by-step execution instructions for the agent.
```

### Key Design Rules

**Step 1 is always context detection**:
```markdown
1. Run `sl spec info --json --paths-only` from repo root **once**.
   Parse: FEATURE_DIR, FEATURE_SPEC, PLAN_FILE, TASKS_FILE.
```

**CLI tools are the data interface** — commands never directly query Supabase, read git state, or parse file paths. They call `sl` CLI commands and parse their output.

**Handoffs define the next stage** — the `handoffs` frontmatter allows the agent shell to suggest the next command after completion.

---

## L2 → L1 Interaction Pattern

Commands call CLI tools for all data operations. This is the primary cross-layer interaction.

```markdown
## In a command prompt:

# Context detection (L1)
sl spec info --json --paths-only

# Comment management (L1, triggers sl-comment skill at L3)
sl comment list --status open --json
sl comment resolve <id> --reason "Resolved: narrowed scope"

# Issue tracking (L1, triggers sl-issue-tracking skill at L3)
sl issue list --status in_progress
sl issue close <id>
```

**Rule**: When a command references a CLI tool, it should mention it briefly (e.g., "fetch open comments using `sl comment list`"). This brief mention triggers the corresponding skill (L3) to load, providing the agent with full usage patterns.

---

## Command Inventory

| Command | Stage | Purpose |
|---------|-------|---------|
| `specify` | Entry | Create feature spec from description |
| `clarify` | Refinement | Resolve ambiguities + process review comments |
| `plan` | Design | Create implementation plan |
| `tasks` | Planning | Generate dependency-ordered tasks |
| `implement` | Execution | Execute tasks, with resume support |
| `verify` | Validation | Cross-artifact consistency check |
| `spike` | Escape hatch | Time-boxed exploratory research |
| `checkpoint` | Escape hatch | Implementation progress + session log |
| `checklist` | Escape hatch | Custom quality gates |
| `onboard` | Setup | Guided project walkthrough |
| `constitution` | Setup | Project principles |
| `commit` | Utility | Auth-aware commit + session capture |

---

## Template Lifecycle

Commands are embedded in the `sl` binary as Go `embed.FS` templates and copied to projects via `sl doctor --template`. See [cli.md — Template Management pattern](cli.md#pattern-classification).

**Source of truth**: `pkg/embedded/templates/specledger/commands/specledger.*.md`
**Runtime copy**: `.claude/commands/specledger.*.md`

Both must stay in sync. Changes to command prompts must update the embedded template first, then the runtime copy.

---

## Anti-Patterns

### AP-01: Command duplicates CLI logic
```markdown
# Wrong: command re-implements branch detection
Run `git branch --show-current` and parse the branch name to find the spec directory...

# Right: command calls CLI
Run `sl spec info --json --paths-only` to get feature context.
```

### AP-02: Command resolves comments without audit trail
```markdown
# Wrong: silent resolution
Run `sl comment resolve <id>` for each addressed comment.

# Right: reason-based resolution
Run `sl comment resolve <id> --reason "Resolved: <summary>"` for each addressed comment.
```

### AP-03: Command hardcodes paths
```markdown
# Wrong: assumes directory structure
Read the spec at `specledger/601-cli-skills/spec.md`

# Right: uses CLI-provided paths
Parse FEATURE_SPEC from `sl spec info --json --paths-only` output.
```

---

## References

- [4-Layer Model Overview](README.md)
- [CLI Design](cli.md) — L1 tools that commands call
- [Skills Design](skills.md) — L3 context that commands trigger
- [598 Spec](../../specledger/598-sdd-workflow-streamline/spec.md) — Original command consolidation decisions
