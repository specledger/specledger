# Research: SDD Layer Alignment

**Feature**: 599-alignment | **Date**: 2026-03-02

## Reference: 598 Consolidation Summary

From 598-sdd-workflow-streamline spec, the following consolidation decisions were made:

### AI Commands: Remove (6)

| Command | Rationale | Absorbed By |
|---------|-----------|-------------|
| `resume` | Duplicate of `implement` | implement |
| `help` | Absorbed by `onboard` | onboard |
| `adopt` | Replaced by context detection (D9) | — |
| `add-deps` | Agent calls `sl deps` CLI directly | — |
| `remove-deps` | Agent calls `sl deps` CLI directly | — |
| `revise` | Absorbed by `clarify` | clarify |

### AI Commands: Rename (1)

| Old | New | Rationale |
|-----|-----|-----------|
| `analyze` | `verify` | OpenSpec terminology alignment |

### AI Commands: Convert to Skill (1)

| Command | Becomes | Rationale |
|---------|---------|-----------|
| `audit` | `sl-audit` skill | Codebase reconnaissance is passive context, not workflow orchestration |

### Final AI Command Count

- **Before**: 16 commands
- **After**: 11 commands (remove 6, rename 1, convert 1)
- **Per 598 SC-001**: "AI command count reduced from 16 to 11"

---

## Merge Analysis

### resume → implement

**Resume content to merge**:
- Check for in-progress tasks
- Continue from last checkpoint
- Update notes field with progress

**Implement changes needed**:
- Add "Resume or Start Fresh" section at start
- Check `sl issue list --status in_progress` for current spec
- If found, prompt to resume

### help → onboard

**Help content to merge**:
- Command overview list
- Core workflow description
- Quick reference

**Onboard changes needed**:
- Add "Command Overview" section after workflow intro
- List all commands with one-line descriptions
- Keep onboarding flow as primary purpose

### revise → clarify

**Revise content to merge**:
- Fetch open comments via `sl comment list`
- Process each comment
- Reply and resolve via `sl comment reply`/`sl comment resolve`

**Clarify changes needed**:
- Add comment processing after ambiguity scan
- Replace `sl revise --summary` with `sl comment list --status open --json`
- Add `sl comment reply` and `sl comment resolve` instructions
- Note: Depends on 598 Stream 1 (`sl comment` CLI)

---

## Layer Architecture Analysis

---

## Layer Architecture Analysis

From 598-sdd-workflow-streamline plan:

| Layer | Name | Runtime | Purpose | Key Characteristic |
|-------|------|---------|---------|-------------------|
| 0 | Hooks | Invisible, event-driven | Auto-capture sessions on commit | No AI, pure automation |
| 1 | `sl` CLI | Go binary, no AI needed | Data operations, CRUD, standalone tooling | JSON output, token-efficient |
| 2 | AI Commands | Agent shell prompts | AI workflow orchestration (specify→implement) | Calls L1, parses output |
| 3 | Skills | Passive context injection | Domain knowledge, progressively loaded | Teaches patterns, no execution |

### Cross-Layer Interactions (from 598)

- **L1→L0**: `sl auth hook --install` configures hooks (convenience pattern)
- **L1→L2**: `sl revise` generates prompt and launches agent (launcher pattern)
- **L2→L1**: AI commands call CLI tools (e.g., `/specledger.tasks` calls `sl issue create`)

## AI Commands Audit

**Location**: `.claude/commands/` (15 files, 2718 total lines)

| File | Lines | Classification | Notes |
|------|-------|----------------|-------|
| specledger.specify.md | 287 | pure-orchestration | Calls git, create-new-feature.sh (bash script) |
| specledger.tasks.md | 417 | pure-orchestration | Heavy `sl issue` usage, well-documented |
| specledger.audit.md | 301 | pure-orchestration | Git operations, file scanning |
| specledger.checklist.md | 293 | pure-orchestration | File operations |
| specledger.implement.md | 239 | pure-orchestration | Git, `sl issue`, file operations |
| specledger.analyze.md | 195 | pure-orchestration | File scanning, `sl issue` |
| specledger.clarify.md | 178 | **has-deprecated-pattern** | Uses `sl revise --summary` (should use `sl comment list`) |
| specledger.adopt.md | 154 | pure-orchestration | Git, file operations |
| specledger.add-deps.md | 134 | pure-orchestration | File operations |
| specledger.plan.md | 101 | pure-orchestration | Calls setup-plan.sh (bash script) |
| specledger.onboard.md | 103 | pure-orchestration | Workflow orchestration |
| specledger.remove-deps.md | 97 | pure-orchestration | File operations |
| specledger.constitution.md | 89 | pure-orchestration | File operations |
| specledger.help.md | 68 | pure-orchestration | Display only |
| specledger.resume.md | 62 | pure-orchestration | `sl issue` operations |

### Key Findings

1. **Deprecated Pattern**: `specledger.clarify.md` uses `sl revise --summary` which should be `sl comment list --status open --json` after 598 is complete

2. **Bash Script Dependencies**: Two commands still call bash scripts:
   - `specledger.specify.md` → `create-new-feature.sh`
   - `specledger.plan.md` → `setup-plan.sh`
   - These will be replaced by `sl spec create` and `sl spec setup-plan` in 598

3. **Consistent CLI Usage**: All commands use `sl issue` properly with `--json` implicit in command design

4. **Template Consistency**: All commands follow similar structure (frontmatter, purpose, outline)

## Skills Audit

**Location**: `pkg/embedded/skills/` and `pkg/embedded/templates/`

| File | Classification | Notes |
|------|----------------|-------|
| sl-issue-tracking/skill.md | unique-knowledge | Good model: domain knowledge + when to use, CLI syntax documented but comprehensive |
| specledger-deps/SKILL.md | unique-knowledge | Dependency management domain knowledge |
| commands/*.md (15 files) | duplicates-commands | These appear to be copies of AI commands in skills location - potential duplication |

### Key Findings

1. **Model Skill**: `sl-issue-tracking/skill.md` is well-structured with:
   - Clear "When to Use" section
   - Decision criteria (sl issue vs TodoWrite)
   - CLI reference with examples
   - Integration patterns

2. **Potential Duplication**: The `pkg/embedded/skills/commands/` directory contains copies of AI commands - this may be intentional (embedded templates) or accidental duplication

3. **Template Location**: `pkg/embedded/templates/specledger/.claude/skills/` contains the actual skill templates

## Decision Criteria Draft

### L0 - Hooks (Event-Driven Automation)

**Use when:**
- Action should happen automatically without user/agent initiation
- Response to git events (commit, push, merge)
- No decision logic needed

**Don't use when:**
- Action requires user input
- Action needs AI reasoning
- Action should be user-initiated

### L1 - CLI (Data Operations)

**Use when:**
- CRUD operations on data (create, read, update, delete)
- Standalone tooling that works without AI
- Cross-platform compatibility needed
- Output should be parseable (JSON)

**Don't use when:**
- Complex multi-step orchestration needed
- AI reasoning required between steps
- Workflow involves user interaction

### L2 - AI Commands (Workflow Orchestration)

**Use when:**
- Multi-step workflow with decision points
- AI reasoning needed between operations
- Orchestrating multiple CLI calls
- Generating content (specs, plans, code)

**Don't use when:**
- Single data operation (use CLI)
- No AI reasoning needed
- Pure automation (use hooks)

### L3 - Skills (Domain Knowledge)

**Use when:**
- Teaching patterns and best practices
- Providing decision criteria
- Context that should load progressively
- Domain-specific knowledge

**Don't use when:**
- Executing operations (use CLI or AI command)
- Duplicate of CLI help text
- Static configuration (use templates)

## Recommendations

### High Priority

1. **Update specledger.clarify.md**: Replace `sl revise --summary` with `sl comment list --status open --json` after 598 Stream 2 complete

2. **Create sl-comment skill**: Document `sl comment` usage patterns (referenced in 598 plan but not implemented)

3. **Create sl-spec skill**: Document `sl spec` usage patterns

### Medium Priority

4. **Standardize AI command template**: Document required sections based on existing patterns

5. **Create skill template**: Based on sl-issue-tracking model

### Low Priority

6. **Audit embedded skills/commands duplication**: Determine if `pkg/embedded/skills/commands/` is intentional or should be removed

## Open Questions

1. Should skills contain CLI syntax examples or link to `--help`?
   - Current model (sl-issue-tracking) includes syntax
   - Trade-off: completeness vs. maintenance burden

2. What triggers skill loading?
   - Needs investigation of skill loading mechanism
   - Important for cross-layer interaction documentation

3. How are embedded templates vs. active files kept in sync?
   - `pkg/embedded/templates/` contains copies of commands
   - Need to understand sync mechanism
