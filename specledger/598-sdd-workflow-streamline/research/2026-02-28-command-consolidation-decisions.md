# Command Consolidation Decision Tracker

**Date**: 2026-02-28
**GitHub Issue**: https://github.com/specledger/specledger/issues/43
**Status**: In Progress

---

## D1: Layer Model

**Decision**: Adopt a 4-layer tooling model

| Layer | Name | Runtime | Purpose |
|-------|------|---------|---------|
| 0 | Hooks | Invisible, event-driven | Auto-capture sessions on commit |
| 1 | `sl` CLI | Go binary, no AI needed | Data operations, CRUD, standalone tooling |
| 2 | Slash commands | Claude Code prompts | AI workflow orchestration (specify→implement) |
| 3 | Skills | Passive context injection | Domain knowledge for agent decision-making |

**Cross-layer interactions**: Layers are not strictly isolated — the CLI (L1) can reach into other layers:
- L1→L0: `sl auth hook --install` configures hooks
- L1→L2: `sl revise` generates a prompt and spawns `claude --prompt` (launcher pattern)
- L2→L1: Slash commands call CLI tools (e.g., `/specledger.tasks` calls `sl issue create`)

These cross-layer interactions are **convenience patterns**, not additional layers.

**Note on Skills (Layer 3)**: Skills are installed per-repo based on tech stack. A future `sl skill` subcommand could discover and install skills (similar to [vercel-labs/skills find.ts](https://raw.githubusercontent.com/vercel-labs/skills/refs/heads/main/src/find.ts)). This is a separate proposal from command consolidation. This should be considered part of repo bootstrap and template management.

---

## D2: CLI vs Slash Command Responsibility

**Decision**: CLI = data operations, Slash command = AI reasoning

| Responsibility | Layer | Example |
|----------------|-------|---------|
| CRUD on entities (issues, comments, deps) | CLI (L1) | `sl comment reply`, `sl issue close` |
| Fetching/displaying data | CLI (L1) | `sl revise --summary`, `sl issue list` |
| Analyzing, reasoning, deciding | Slash command (L2) | `/specledger.clarify` analyzes gaps |
| Applying changes based on reasoning | Slash command (L2) | Agent edits files, then calls `sl comment resolve` |

**The revise middle ground**: `sl revise` (CLI) remains as an optional launcher that pre-filters comments and spawns an agent. The agent's behavior is defined by the slash command (`/specledger.clarify --comments`), and the agent uses `sl comment reply/resolve` for granular, detailed resolution. This eliminates the current tension where:
- The CLI tried to interpret agent results post-session (clunky)
- Comments were bulk-resolved without detail

**Key principle**: The agent should have proper CLI tools to perform data operations with full context, rather than the CLI guessing what the agent accomplished. When control flow resumes in the CLI, it should warn and prompt the user to let the agent do commits/resolve comments directly for better traceability and remove logic to resolve comments or do git operations from the CLI (except from the beginning where the CLI does help make sure the right branch is checked out and any changes are taken care (stash or allow user to cancel and commit first before relaunching revise)).

---

## D3: Template Management

**Side note**: The `/specledger.revise` slash command was removed from playbook templates after the `sl revise` CLI command was merged, but existing repos that were bootstrapped before the removal still have the stale command. Template management should:
- Own the `specledger.` prefix in `.claude/commands/`
- On `sl doctor --template` or `sl init`, detect and prompt to remove commands that are no longer in the playbook
- Never silently delete user-modified commands — prompt first

**Status**: Open — needs separate proposal for template lifecycle management.

---

## D4: Slash Command Consolidation (16 → 12)

**Decision**: Consolidate slash commands while preserving SpecKit nomenclature, add 2 new commands

### Final Command Set

| Command | Absorbs | Stage |
|---------|---------|-------|
| `/specledger.specify` | — | Core pipeline |
| `/specledger.clarify` | `/specledger.revise` | Core pipeline (agent uses `sl comment` for granular resolution) |
| `/specledger.plan` | — | Core pipeline |
| `/specledger.tasks` | — | Core pipeline |
| `/specledger.implement` | `/specledger.resume` | Core pipeline |
| `/specledger.spike` | — | **New** — exploratory research (D13) |
| `/specledger.checkpoint` | — | **New** — implementation verification + session log (D14) |
| `/specledger.analyze` | — | Quality validation (keeps SpecKit brand name) |
| `/specledger.checklist` | — | Optional quality gate (standalone, per-feature custom gates) |
| `/specledger.onboard` | `/specledger.help` | Setup |
| `/specledger.constitution` | — | Setup (runs rarely) |
| `/specledger.audit` | — | Codebase analysis |

### Removed (6 commands)

| Removed | Reason |
|---------|--------|
| `/specledger.resume` | Duplicate of `/specledger.implement` |
| `/specledger.revise` | Absorbed by `/specledger.clarify` + `sl comment` CLI |
| `/specledger.help` | Absorbed by `/specledger.onboard` welcome step |
| `/specledger.adopt` | Moves to CLI as `sl spec link` (branch→spec mapping is deterministic, not AI) |
| `/specledger.add-deps` | Agent calls `sl deps add` directly |
| `/specledger.remove-deps` | Agent calls `sl deps remove` directly |

### New CLI subcommand required

`sl comment list/show/reply/resolve/pull/push` — enables agents to manage review comments with full detail rather than bulk resolution.

---

## D5: Skills Architecture

**Decision**: Keep skills separate, lean, and progressively loaded

- Skills are silently injected into agent context as needed
- Each CLI domain gets its own skill (`sl-issue-tracking`, `specledger-deps`, future `sl-comment`)
- Skills must be focused and isolated — not consolidated into a monolith
- Slash commands reference CLI tools briefly (triggering skill loading), skills provide the deep usage patterns
- Skill discovery/install is a separate concern (future `sl skill` command, part of bootstrap/template management — see D19)
- **Two types of skills**: (1) CLI-embedded skills (`sl-issue-tracking`, `specledger-deps`, `sl-comment`) that ship with the `sl` binary and teach agents to use `sl` commands, and (2) external registry skills (e.g., vercel-labs/skills) that provide domain expertise (security, React best practices, etc.). `sl skill` would manage type (2); type (1) is handled by `sl doctor --template`.

**Pattern**: `sl comment` follows the same pattern as `sl issue`:
- No dedicated slash command mapping
- A `sl-comment` skill teaches the agent when/how to use `sl comment`
- `/specledger.clarify` references comments briefly, which triggers the skill to load progressively

---

## D6: CLI Command Cleanup

### `sl graph` → fold into `sl deps`

`sl graph` is currently stubbed ("coming soon") with `show`, `export`, `transitive` subcommands. Meanwhile `sl issue list --graph` already shows task dependency graphs. The `sl graph` command should be scoped to **spec dependency visualization** and folded into `sl deps` (e.g., `sl deps graph`), not kept as a top-level command.

### `sl playbook` — future org-level management

Currently only `sl playbook list`. Future: playbook management at org level via the SpecLedger web app. Note for proposal.

### `sl issue migrate` — keep indefinitely

Beads migration tooling stays as a safety net. Even with beads removed from this repo, other projects may still need it.

---

## D7: `sl init` → Onboard Launcher

**Decision**: `sl init` should offer to launch the onboarding workflow after filesystem setup

Flow:
1. `sl init` creates `.claude/`, `specledger/`, `specledger.yaml` (existing behavior)
2. **New**: Prompt user: "Launch guided onboarding? (Y/n)"
   - Default (Y): `sl init` spawns `claude --prompt "/specledger.onboard"` (launcher pattern, same as `sl revise`)
   - Opt-out (n): Print instructions showing how to call `/specledger.onboard` manually (advanced users)

This mirrors the `sl revise` launcher pattern (D2) and gives new users a seamless path from `sl init` to their first feature spec.

---

## D8: Full CLI ↔ Layer Mapping

### CLI commands — pure data ops (no slash command counterpart)

| CLI Command | Purpose | Notes |
|-------------|---------|-------|
| `sl auth` | Authentication management | L1→L0 cross-layer (`hook --install`) |
| `sl doctor` | Environment health check | Auto-updates CLI + templates |
| `sl version` | Version info | — |
| `sl playbook` | List/manage playbooks | Future: org-level management in web app (T3) |
| `sl deps` | Dependency management | Absorbs `sl graph` (D6) |
| `sl session` | AI session capture | Invisible via hooks (L0); list/get for audit |
| `sl new` | Project creation (TUI) | — |
| `sl issue` | Issue CRUD | Skill-driven agent usage |
| `sl comment` (new) | Comment CRUD | Skill-driven agent usage |
| ~~`sl spec link`~~ | ~~Branch→spec mapping~~ | Superseded: D9 revised — fallback chain in context detector + `specledger.yaml` aliases |

### CLI commands — launcher pattern (L1→L2 cross-layer)

| CLI Command | Launches | Agent Behavior Defined By |
|-------------|----------|--------------------------|
| `sl revise` | `claude --prompt` with comment context | `/specledger.clarify` + `sl-comment` skill |
| `sl init` (proposed D7) | `claude --prompt "/specledger.onboard"` | `/specledger.onboard` |

### Slash commands — pure AI reasoning (no CLI counterpart needed)

| Command | Stage | CLI Tools Called |
|---------|-------|-----------------|
| `/specledger.specify` | Core pipeline | `sl issue create` (optional) |
| `/specledger.clarify` | Core pipeline | `sl comment list/reply/resolve` |
| `/specledger.plan` | Core pipeline | — |
| `/specledger.tasks` | Core pipeline | `sl issue create/link` |
| `/specledger.implement` | Core pipeline | `sl issue update/close` |
| `/specledger.spike` | Research (any stage) | — |
| `/specledger.checkpoint` | Verification + session log | `sl issue show` (compare plan vs actual) |
| `/specledger.analyze` | Quality validation | `sl comment list --status open` (flag unresolved) |
| `/specledger.checklist` | Optional quality gate | — |
| `/specledger.onboard` | Setup | Orchestrates other slash commands |
| `/specledger.constitution` | Setup | — |
| `/specledger.audit` | Codebase analysis | — |

---

## D9: Adopt → No New Command, Improve Context Detection

**Decision**: Remove `/specledger.adopt` entirely. No new CLI command needed. Instead, enhance the existing `ContextDetector` with a robust fallback chain.

### Current State (fragile)

The context detector (`pkg/issues/context.go`) uses a single regex:

```go
var specBranchPattern = regexp.MustCompile(`^(\d{3,}-[a-z0-9-]+)$`)
```

If the branch name doesn't match `###-slug` → hard error. No fallback. This is why `/specledger.adopt` was created — to work around detection failures for branches created by external tools (GitHub UI, Jira, Linear, etc.).

### Proposed: Fallback Chain

Replace the single regex with a 4-step resolution chain:

```
1. Regex match (current)     → instant, deterministic
       ↓ fail
2. YAML alias lookup         → check specledger.yaml branch_aliases
       ↓ fail
3. Git heuristic             → diff branch vs base, find touched specledger/ dirs
       ↓ fail
4. Interactive prompt        → ask user, save alias to specledger.yaml
```

#### Step 1: Regex Match (existing)
Branch `598-sdd-workflow-streamline` → `specledger/598-sdd-workflow-streamline/`. Instant.

#### Step 2: YAML Alias Lookup (new)
Check `specledger.yaml` for saved mappings:

```yaml
branch_aliases:
  "feature/fix-login-bug": "042-auth-improvements"
  "johns-auth-work": "042-auth-improvements"
```

#### Step 3: Git Heuristic (new)
Diff the branch against its base and inspect which `specledger/` directories were modified:

```bash
git diff --name-only main...HEAD | grep '^specledger/' | cut -d/ -f2 | sort -u
```

Note: Review `git` commands like `ls-tree` to find if the `specledger/` has any changes (checksum change against base branch `specledger`) instead and identify which slug?Would that be faster than diff (stats/name-only) grep and cut/sort?

- **Exactly one result** → auto-resolve, offer to save alias
- **Multiple results** → prompt user to pick, save alias
- **Zero results** → fall through to step 4

This is the key insight: if someone is working on a non-conforming branch but has been editing `specledger/042-auth-improvements/spec.md`, the heuristic resolves it without any manual intervention.

#### Step 4: Interactive Prompt (new)
Last resort — list available spec directories, let user pick, save to `specledger.yaml`:

```
⚠ Can't determine which spec branch "feature/fix-login-bug" belongs to.

Available specs:
  1. 042-auth-improvements
  2. 043-payment-flow
  3. 044-dashboard-v2

Which spec does this branch belong to? [1-3, or 'skip']:
```

For non-interactive/CI: `--spec <slug>` flag already exists on most commands as a manual override.

### Why This Eliminates Adopt

- **No new command** — the mapping is a side-effect of the detection logic, triggered on-demand at the point of need
- **On-demand** — only prompts when a command actually needs spec context and can't find it
- **Discoverable** — user doesn't need to know about aliases; the system guides them at the point of failure
- **Version-controlled** — aliases stored in `specledger.yaml`, visible to the team
- **Lazy initialization** — mapping created once, auto-resolves for all future commands on that branch

### What `/specledger.adopt` was vs what this replaces

| Aspect | Old `/specledger.adopt` | New detection fallback |
|--------|------------------------|----------------------|
| Trigger | Manual slash command invocation | Automatic on detection failure |
| Layer | L2 (AI reasoning — overkill) | L1 (CLI detection logic) |
| Storage | Unknown/ad-hoc | `specledger.yaml` (version-controlled) |
| Git analysis | None | Diff-based heuristic |
| User friction | Must know about adopt + run it proactively | Zero — system prompts when needed |

---

## D10: Phase Out Bash Scripts

**Decision**: Replace `.specledger/scripts/bash/` scripts with `sl` CLI subcommands

Several slash commands (notably `/specledger.specify`) call bash scripts like `create-new-feature.sh` for deterministic operations. These scripts:
- Don't work on Windows (sometimes have PowerShell equivalents, sometimes not)
- Are a maintenance burden (two codepaths for the same operation)
- Should be replaced by cross-platform `sl` CLI commands

**Migration path**: Each bash script's functionality should be identified and absorbed into the appropriate `sl` subcommand. The CLI (Go binary) is already cross-platform.

**Status**: Needs inventory of all scripts and their CLI equivalents.

---

## D11: Playbooks = Skill Bundles, Core Workflow is Fixed

**Decision**: The core workflow (specify→clarify→plan→tasks→implement) is immutable. Playbooks customize content, not workflow shape.

| Concept | Owns | Does NOT own |
|---------|------|-------------|
| Core workflow | Pipeline stages (specify→implement) | — |
| Playbooks | Suggested skills, templates, constitution defaults | Workflow stage order or additions |
| Skills | Domain-specific patterns injected per-stage | Pipeline structure |

**Rationale**: The specify→plan→tasks→implement workflow has been tested across different teams (data, ML, platform) and works universally. Playbooks are a fast way to propose known skills and templates for a given tech stack or team type, allowing orgs to fine-tune without breaking the core workflow.

**Future**: Org-level playbook management in the SpecLedger web app.

---

## D12: Checklist — Keep as Optional Standalone

**Decision**: `/specledger.checklist` remains as a standalone optional command, not merged into `/specledger.analyze`

**Rationale**: Checklist was a late addition from SpecKit that never fully landed. Its purpose (custom per-feature quality gates on top of the constitution) is distinct from analyze (cross-artifact consistency). Keeping it optional lets teams that want custom gates use it without cluttering the core pipeline.

**Note**: May evolve to become a constitution-level feature in the future. Defer deeper design.

---

## D13: Spike — Exploratory Research Command

**Decision**: Add `/specledger.spike` as a new slash command for time-boxed exploratory research

Spikes are essential for reducing implementation risk by validating approaches early. They sit alongside the core pipeline — typically invoked between `plan` and `tasks`, or anytime during `implement` when an unknown surfaces.

**Output**: Research document at `specledger/<spec-id>/research/yyyy-mm-dd-<spike-name>.md`

**Contents**:
- Research question / hypothesis
- Approach explored
- Findings (with evidence — code snippets, API responses, benchmarks)
- Recommendations (proceed / pivot / abandon)
- Impact on spec/plan (what needs updating based on findings)

**Relationship to core pipeline**: Spikes are **not** a pipeline stage — they're an escape hatch. The core workflow (specify→plan→tasks→implement) remains immutable per D11. Spikes can be invoked at any point and their findings feed back into the relevant artifact.

---

## D14: Checkpoint — Implementation Verification + Session Log

**Decision**: Add `/specledger.checkpoint` as a new slash command that serves two purposes:

### Purpose 1: Implementation Verification
Verify current implementation aligns with the spec and plan. Detect drift early.

### Purpose 2: Session Log (Resolves T1)
At the end of an implementation session (or at meaningful milestones), the agent **looks back** at what was done and produces a structured log entry capturing:

- **Tasks worked on**: Which issues were addressed (by ID)
- **What was planned**: Original task description and acceptance criteria
- **What was actually done**: Summary of actual changes made
- **Divergences**: Where and why the implementation differs from the plan
- **Decisions made**: Key choices the agent or human made during the session
- **Unfinished work**: What was started but not completed, and why
- **Impact on downstream tasks**: Whether divergences affect other tasks

**Output**: Session file at `specledger/<spec-id>/sessions/yyyy-mm-dd-<checkpoint-name>.md`

**Why this matters for reviewers**: When a PR lands, reviewers can read the checkpoint logs to understand:
- Why the implementation looks different from the original plan
- What trade-offs were made and why
- Where to focus their review attention (divergence areas)
- What the agent decided vs what the human directed

**Trigger points**:
- End of implementation session (manual invocation)
- After completing a task phase
- When the agent detects significant drift from the plan
- Before context compaction (to preserve reasoning that would be lost)

**Relationship to `sl session capture` (L0)**: Session capture records the raw conversation transcript. Checkpoint produces a **curated, human-readable summary** with structured deviation tracking. They complement each other — session capture is forensic evidence, checkpoint is the executive summary.

---

## D15: Mockup Command — Launcher Pattern, Not New Workflow Stage

**Decision**: `sl mockup` is a CLI launcher command (same pattern as `sl revise`), not a new core pipeline stage.

The `598-mockup-command` branch has ~4,000 lines of Go implementation. It follows the launcher pattern (D2):
- CLI handles deterministic work: framework detection, component scanning, design system index, prompt building
- AI agent handles the creative work: generating the actual mockup (HTML/JSX)
- Post-session: commit/push flow

**Key observations from spike ([research/2026-02-28-spike-magic-patterns.md](/tmp/specledger-598/specledger/598-sdd-workflow-streamline/research/2026-02-28-spike-magic-patterns.md))**:
- Magic Patterns (SaaS) could complement but not replace the pipeline (offline requirement, spec integration)
- The current approach is sound for v1
- A `/specledger.mockup` slash command is NOT needed — the CLI launcher handles context gathering and the agent receives a structured prompt

**Where mockup fits in D8 mapping**: Launcher pattern (L1→L2), same row as `sl revise` and `sl init`.

---

## D16: CLI Development Constitution

**Decision**: The `sl` binary needs a constitution that classifies development patterns and enforces review gates.

### Established CLI Patterns

Any new `sl` subcommand MUST identify which pattern(s) it uses. If it introduces a new pattern, it requires wider team review.

| Pattern | Description | Examples | Constraints |
|---------|-------------|----------|-------------|
| **Data CRUD** | Deterministic operations on entities (create/read/update/delete) | `sl issue`, `sl deps`, `sl comment` (proposed) | Must work offline. No AI reasoning. Returns structured data. |
| **Launcher** | Pre-flight checks + context gathering + spawn AI agent | `sl revise`, `sl mockup`, `sl init` (proposed) | CLI does NOT interpret agent results. Agent owns commits/resolution. CLI only does pre-session setup. |
| **Hook trigger** | Invisible automation fired by agent shell events | `sl session capture` | Must be non-blocking. Must handle failures gracefully (queue for retry). |
| **Environment** | System health, auth, configuration | `sl doctor`, `sl auth`, `sl version`, `sl config` | Must be idempotent. Must work without a project context. |
| **Template management** | Install/update/remove agent shell templates (commands, skills) | `sl doctor --template`, `sl init` | Must detect stale/deprecated templates. Must prompt before destructive changes. Owns `specledger.` prefix. |

### Review Gates

| Scenario | Required Review |
|----------|----------------|
| New subcommand using existing pattern | Standard PR review |
| New subcommand introducing new pattern | Wider team review + constitution update |
| Subcommand that doesn't fit any pattern | Block until pattern is identified or constitution updated |
| Change to cross-layer interaction (e.g., new launcher) | Review by CLI + agent shell owners |

### Constitution Checks for PRs

When reviewing a PR that adds/modifies `sl` CLI code:

1. **Pattern classification**: Which pattern(s) does this use? Is it documented?
2. **Layer boundary**: Does it respect D2 (CLI = data ops, not AI reasoning)?
3. **Offline capability**: Does it work without network? If not, is there a fallback?
4. **Cross-platform**: Does it work on macOS, Linux, and Windows? (D10 — no bash scripts)
5. **Template ownership**: Does it touch `specledger.` prefixed files? If so, does it follow D3?
6. **Justification for deviation**: If it breaks any constraint, is there a documented reason?

---

## Identified Tensions (Need Wider Alignment)

### ~~T1: Implementation Tracking Gap~~ → Resolved by D14

Checkpoint command addresses this tension directly. The session log component produces structured deviation tracking that maps plan→actual for each implementation session.

### T2: CLI Launcher Pattern — How Far Does It Go?

**Tension**: `sl revise` and `sl init` (proposed) both use the launcher pattern (CLI spawns `claude --prompt`). Should this generalize to `sl specify`, `sl plan`, `sl implement`?

**Trade-offs**:
- Pro: Consistent UX — developers always start with `sl <verb>`
- Pro: CLI can do pre-flight checks (branch, auth, stash) before launching
- Con: Adds indirection — user could just type `/specledger.specify` directly
- Con: Each launcher needs to know what context to pre-fetch
- Con: Not enough usage data to know if this helps or hinders

**Status**: Deferred — need real user feedback on `sl revise` and `sl init` patterns first.

### T3: Playbooks vs Skills Boundary

**Tension**: If playbooks only bundle skills, what distinguishes a playbook from a `package.json`-like dependency list? Is there room for playbooks to own more (e.g., default constitution values, org-specific templates, CI/CD integration)?

**Status**: Deferred — depends on org-level playbook management design.

---

## Open Questions

| # | Question | Status |
|---|----------|--------|
| Q1 | Should the CLI launcher pattern generalize beyond `sl revise` and `sl init`? | Tension T2 — deferred |
| Q2 | How should `sl doctor --template` handle stale commands from older playbook versions? | D3 — needs template lifecycle proposal |
| Q3 | Should `sl skill` discover/install skills based on repo tech stack? | Deferred — separate proposal. See D19. |
| Q4 | Inventory of bash scripts to replace with CLI commands? | D10 — needs script audit |
| Q5 | ~~How to track plan→actual deviations during implementation?~~ | Resolved by D14 (checkpoint) |
| ~~Q6~~ | ~~Naming for branch→spec mapping CLI command?~~ | Resolved — D9 revised: no new command, context detector fallback chain |

---

## Decision Log

| # | Question | Decision | Date |
|---|----------|----------|------|
| D1 | Layer model | 4-layer (Hooks → CLI → Commands → Skills) with cross-layer interactions | 2026-02-28 |
| D2 | CLI vs slash command responsibility | CLI = data ops, Command = AI reasoning, Launcher = optional convenience | 2026-02-28 |
| D3 | Template management | `specledger.` prefix is owned by playbook; stale commands should be detected and prompted for removal | 2026-02-28 |
| D4 | Slash command consolidation | 16 → 12 commands (remove 6, add 2: spike + checkpoint) | 2026-02-28 |
| D5 | Skills architecture | Keep separate, lean, progressively loaded; no consolidation | 2026-02-28 |
| D6 | `sl graph` cleanup | Fold into `sl deps graph`; `sl issue list --graph` handles task graphs | 2026-02-28 |
| D7 | `sl init` → onboard launcher | Prompt user to launch onboarding after init (default yes, opt-out for advanced) | 2026-02-28 |
| D8 | Full CLI ↔ layer mapping | Complete mapping documented; spike + checkpoint added | 2026-02-28 |
| D9 | Adopt → context detection | No new command; enhance ContextDetector with 4-step fallback chain (regex → yaml alias → git heuristic → prompt) | 2026-02-28 |
| D10 | Phase out bash scripts | Replace with cross-platform `sl` CLI subcommands. Branch number generation must address collision issue ([#46](https://github.com/specledger/specledger/issues/46)). | 2026-02-28 |
| D11 | Core workflow immutability | specify→plan→tasks→implement is fixed; playbooks = skill bundles | 2026-02-28 |
| D12 | Checklist fate | Keep as optional standalone; not merged into analyze | 2026-02-28 |
| D13 | Spike command | New `/specledger.spike` for time-boxed exploratory research | 2026-02-28 |
| D14 | Checkpoint + session log | New `/specledger.checkpoint` for implementation verification + deviation tracking (resolves T1) | 2026-02-28 |
| D15 | Mockup command | Launcher pattern, not new pipeline stage. Spike: Magic Patterns is complementary, not replacement. | 2026-02-28 |
| D16 | CLI development constitution | 5 established patterns (Data CRUD, Launcher, Hook trigger, Environment, Template mgmt) with review gates | 2026-02-28 |
| D17 | Playbook frozen until webapp | `sl playbook` stays in CLI but no new features. Skill bundles (ML, backend, frontend, fullstack) designed in webapp first, then flows back to CLI. | 2026-02-28 |
| D18 | Audit vs `/simplify` | Keep `/specledger.audit` — it's codebase reconnaissance (onboarding, module graphs). Claude Code `/simplify` is complementary (PR-level code quality). | 2026-02-28 |
| D19 | `sl skill` command | Deferred. Future command for skill discovery/install based on repo tech stack (similar to vercel-labs/skills find.ts). Depends on playbook management (D17) and webapp design. Skills embedded in CLI (sl-issue-tracking, specledger-deps, sl-comment) are distinct from external skill registries. | 2026-02-28 |
| D20 | US1 "Audit" naming | US1 is a workflow review/inventory exercise, NOT the `/specledger.audit` AI command (which scans source code). Naming must stay distinct. Review carefully before finalizing. | 2026-02-28 |
