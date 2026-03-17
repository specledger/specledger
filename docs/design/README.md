# SpecLedger Design Documentation

Architecture and design principles for the SpecLedger 4-layer tooling model.

---

## 4-Layer Model

SpecLedger organizes its tooling into four layers, each with distinct responsibilities:

| Layer | Name | Runtime | Purpose | Design Doc |
|-------|------|---------|---------|------------|
| **L0** | Hooks | Invisible, event-driven | Auto-capture sessions on commit | [hooks.md](hooks.md) |
| **L1** | `sl` CLI | Go binary, no AI needed | Data operations, CRUD, standalone tooling | [cli.md](cli.md) |
| **L2** | AI Commands | Agent shell prompts (`/specledger.*`) | AI workflow orchestration (specify → implement) | [commands.md](commands.md) |
| **L3** | Skills | Passive context injection | Domain knowledge for agent decision-making | [skills.md](skills.md) |

### Layer Responsibilities

```
L3 Skills      "How should I think about this?"    Passive context, loaded on-demand
L2 Commands    "What workflow should I follow?"     Multi-step orchestration
L1 CLI         "Do this specific operation"         Deterministic data ops
L0 Hooks       "Capture this automatically"         Invisible automation
```

---

## Cross-Layer Interactions

Layers are not strictly isolated. These convenience patterns enable the layers to work together:

```
L2 → L1    AI commands call CLI tools
           /specledger.clarify calls `sl comment list --json`

L1 → L2    CLI launcher spawns agent session
           `sl revise` gathers context, launches `claude --prompt`

L1 → L0    CLI installs hook configuration
           `sl auth hook --install` writes to .claude/settings.json

L0 → L1    Hooks invoke CLI commands in hook context
           PostToolUse hook runs `sl session capture`

L2 → L3    Commands trigger skill loading
           /specledger.clarify mentions `sl comment` → sl-comment skill loads
```

**Rule**: Cross-layer calls are one-directional convenience patterns, not bidirectional dependencies. L3 never calls L2. L0 never calls L2. The data flows up (L0 → L1) and orchestration flows down (L2 → L1).

---

## Core Workflow

The specify → implement pipeline is immutable across all projects. Playbooks customize content (skill bundles, templates), not workflow shape.

```
specify → clarify → plan → tasks → implement
                      │       │         │
                      │       │         ├─ checkpoint (session handover, progress capture,
                      │       │         │              deviation tracking between sessions)
                      │       │         └─ commit (auth-aware session capture)
                      │       │
                      │       ├─ spike (time-boxed research before committing to approach)
                      │       └─ checklist (custom quality gates for task verification)
                      │
                      └─ verify (cross-validate spec ↔ plan ↔ tasks consistency;
                                 recommended after all three artifacts exist)
```

**Stage-aligned escape hatches**:
- **spike** — Planning phase. Research technologies/patterns before committing to a plan or task approach.
- **checklist** — Planning/tasks phase. Custom quality gates for task verification.
- **verify** — Post-planning. Cross-validates spec ↔ plan ↔ tasks for consistency. Strongly recommended once all three artifacts exist.
- **checkpoint** — Implementation phase. Captures session progress, deviations from plan, and handover notes when context window is exhausted or between sessions. Creates structured documentation of what was planned vs what was actually done.
- **commit** — Implementation phase. Auth-aware git commit with automatic session capture via hooks.
- **constitution** / **onboard** — Setup-time, before the pipeline starts.

See [commands.md](commands.md) for the full command inventory and pipeline details.

---

## Key Design Principles

Each layer doc contains detailed principles. Here are the cross-cutting ones:

### 1. CLI is the Agent's Primary Interface
The `sl` CLI is the hub. AI commands (L2) call it for data operations. Skills (L3) teach the agent how to use it. Hooks (L0) invoke it for automation. Every data operation goes through L1.

### 2. Progressive Disclosure
Information is discovered on-demand, not preloaded:
- **CLI**: `--help` → subcommand usage → specific params (see [cli.md](cli.md))
- **Skills**: Loaded when triggered, not at session start (see [skills.md](skills.md))
- **Commands**: Each stage reveals the next via handoffs (see [commands.md](commands.md))

### 3. Agent Owns Outcomes
The CLI provides tools; the agent makes decisions. The CLI does NOT interpret agent results, auto-resolve comments, or make commits post-session. When control returns to the CLI after an agent session:
- **Detect**: Check if commits were made, comments were resolved
- **Warn**: Print clear guidance if actions were missed
- **Guide**: Print resume command so user can re-enter the session
- **No auto-action**: The CLI warns, the agent acts

### 4. Error as Navigation
Every error across all layers should guide toward the correct action:
- **L1 CLI**: Error + suggested fix command (see [cli.md — Error Messages](cli.md#principle-2-error-messages-as-navigation))
- **L0 Hooks**: Silent failure + local cache (see [hooks.md — Silent Resilience](hooks.md#core-principle-silent-resilience))
- **L2 Commands**: Prerequisite validation before execution
- **L3 Skills**: Error handling tables with cause + solution

---

## Governing Principles

These design docs and the [Project Constitution](../../.specledger/memory/constitution.md) operate at **different levels of concern**:

| | Constitution | Design Docs |
|---|---|---|
| **Scope** | How we plan and build software | How the software behaves at runtime |
| **Audience** | Developers deciding *what to build* | Developers deciding *how the CLI responds* |
| **Example** | "Fail Fast, Fix Forward" — surface errors early in implementation, don't silently swallow failures | "Error Messages as Navigation" — every CLI error includes what failed, why, and a suggested fix command |
| **Governs** | Development workflow, testing strategy, branch hygiene | CLI output format, cross-layer interactions, command patterns |

The constitution covers:
- **YAGNI, DRY, Simplicity** — What to build and how much complexity to accept
- **Contract-First Testing** — How API contracts are snapshotted and validated
- **Supabase Local Stack** — Infrastructure requirements for every feature branch
- **Quickstart-Driven Validation** — How user scenarios map to E2E tests

These design docs cover:
- **Progressive Discovery** — Information is revealed on-demand, not preloaded
- **Error Messages as Navigation** — Runtime UX: every error guides toward the correct action
- **Two-Level Output Design** — JSON for agents, compact for humans
- **Agent Owns Outcomes** — CLI provides tools; the agent makes decisions

**These are complementary, not overlapping.** A constitutional principle like "Fail Fast" tells you to design your implementation to surface errors early. A design principle like "Error as Navigation" tells you how the CLI should format those errors at runtime. When a design decision conflicts with a constitutional principle, the constitution takes precedence.

---

## Document Index

| Document | Covers |
|----------|--------|
| [cli.md](cli.md) | CLI design principles, pattern classification, output format, error design, anti-patterns |
| [hooks.md](hooks.md) | Hook selection, silent resilience, session capture, available Claude Code hooks |
| [commands.md](commands.md) | AI command anatomy, core pipeline, L2→L1 interaction, template lifecycle |
| [skills.md](skills.md) | Skill anatomy, progressive loading, cross-layer alignment, skill creator reference |
| [testing.md](testing.md) | Layer-by-layer testing strategy, quickstart-driven E2E, contract snapshots |
| [tech-debts.md](tech-debts.md) | Implementation gaps found during design review (scratchpad, to be filed as GH issues) |

---

## Decision History

These design decisions originated in feature specs and are now canonicalized in the docs above. The decision IDs (D1, D4, etc.) are from the [598-sdd-workflow-streamline spec](../../specledger/598-sdd-workflow-streamline/spec.md) where the 4-layer model was first designed.

| ID | Decision | Summary | Canonical Location |
|----|----------|---------|-------------------|
| D1 | 4-layer tooling model | Organize tooling into Hooks (L0), CLI (L1), Commands (L2), Skills (L3) with defined responsibilities and cross-layer interaction rules | This README |
| D2 | Clarify absorbs revise | `/specledger.clarify` handles both spec ambiguity and review comment processing. `sl revise` stays as CLI launcher. | [commands.md](commands.md) |
| D3 | Template lifecycle via `sl doctor` | `sl doctor --template` manages command/skill updates. Detects stale templates, prompts before destructive changes. No separate `sl update` command. | [cli.md — Template Management](cli.md#pattern-classification) |
| D4 | Extract `sl comment` CLI + skill | Review comment management as standalone Data CRUD CLI (`sl comment`) with complementary skill (`sl-comment`). Same pattern as `sl issue` / `sl-issue-tracking`. | [cli.md](cli.md), [skills.md](skills.md) |
| D5 | Skills complement, don't duplicate | Skills teach the agent *when* and *how* to use CLI tools. They never duplicate CLI logic or orchestrate workflows. | [skills.md](skills.md) |
| D9 | Context detection fallback chain | `ContextDetector` resolves branch → spec via: regex match → yaml alias → git heuristic → interactive prompt. Replaces `adopt` command. | [cli.md — AP-04](cli.md#ap-04-duplicating-contextdetector-per-package) |
| D11 | Core workflow is immutable | specify → clarify → plan → tasks → implement pipeline cannot be modified by playbooks. Playbooks customize content (skill bundles), not workflow shape. | [commands.md](commands.md) |
| D13 | Spike command | Time-boxed exploratory research. Output: `specledger/<spec>/research/yyyy-mm-dd-<topic>.md`. Escape hatch during planning. | [commands.md](commands.md) |
| D14 | Checkpoint + session log | Implementation verification with structured deviation tracking. Captures planned vs actual, decisions made, unfinished work. Critical for session handover. | [commands.md](commands.md) |
| D16 | CLI development constitution | 5 command patterns (Data CRUD, Launcher, Hook Trigger, Environment, Template Management) with failure mode constraints and review expectations. | [cli.md — Pattern Classification](cli.md#pattern-classification) |
| D21 | Token-efficient output | Human output: compact (truncated previews, counts, footer hints). JSON output: complete (full data, pipeable). Token efficiency via workflow pattern (list → show), not field truncation. | [cli.md — Two-Level Output](cli.md#principle-3-two-level-output-design) |

**Source specs** (historical, decisions are now canonicalized in design docs):
- [specledger/598-sdd-workflow-streamline/spec.md](../../specledger/598-sdd-workflow-streamline/spec.md) — Original decisions D1-D21
- [specledger/599-alignment/spec.md](../../specledger/599-alignment/spec.md) — Command consolidation (15 → 9)
- [specledger/601-cli-skills/research.md](../../specledger/601-cli-skills/research.md) — CLI extraction and skill design research
- [specledger/010-checkpoint-session-capture/spike-hooks.md](../../specledger/010-checkpoint-session-capture/spike-hooks.md) — Hook approach evaluation
- [specledger/602-silent-session-capture/research.md](../../specledger/602-silent-session-capture/research.md) — Silent failure and error logging research
