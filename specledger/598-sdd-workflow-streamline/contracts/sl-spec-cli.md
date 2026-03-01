# CLI Contract: `sl spec` + `sl context`

**Pattern**: Data CRUD / Environment (D16) | **Layer**: L1 (CLI)

## `sl spec info`

Replaces: `check-prerequisites.sh`

```
sl spec info [--json] [--require-plan] [--require-tasks] [--include-tasks] [--paths-only] [--spec <key>]
```

**Flags**:
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--json` | bool | false | JSON output (for AI command consumption) |
| `--require-plan` | bool | false | Error if plan.md doesn't exist |
| `--require-tasks` | bool | false | Error if tasks.md doesn't exist |
| `--include-tasks` | bool | false | Include tasks.md in AVAILABLE_DOCS |
| `--paths-only` | bool | false | Only output paths, skip validation |
| `--spec` | string | auto | Override spec key (bypass detection) |

**JSON Output** (`--paths-only`):
```json
{
  "REPO_ROOT": "/path/to/repo",
  "BRANCH": "598-sdd-workflow",
  "FEATURE_DIR": "/path/to/repo/specledger/598-sdd-workflow",
  "FEATURE_SPEC": "/path/to/repo/specledger/598-sdd-workflow/spec.md",
  "IMPL_PLAN": "/path/to/repo/specledger/598-sdd-workflow/plan.md",
  "TASKS": "/path/to/repo/specledger/598-sdd-workflow/tasks.md",
  "HAS_GIT": true
}
```

**JSON Output** (validation mode):
```json
{
  "FEATURE_DIR": "/path/to/repo/specledger/598-sdd-workflow",
  "AVAILABLE_DOCS": ["research.md", "data-model.md", "contracts/", "tasks.md"]
}
```

**Compatibility**: Output fields MUST match `check-prerequisites.sh --json` exactly so AI command templates work unchanged during migration.

---

## `sl spec create`

Replaces: `create-new-feature.sh`

```
sl spec create [--json] [--number N] [--short-name "name"] "feature description"
```

**Flags**:
| Flag | Type | Required | Description |
|------|------|----------|-------------|
| `--json` | bool | no | JSON output |
| `--number` | int | no | Override auto-detected branch number |
| `--short-name` | string | no | Override auto-generated short name |

**Positional**: Feature description (required)

**Behavior**:
1. Scan `specledger/*/` dirs for highest numeric prefix
2. Check remote branches for collision prevention ([#46])
3. Generate short name from description (stop-word filter, 3-4 words)
4. Format: `{number}-{short-name}` (e.g., `599-comment-crud`)
5. Enforce GitHub 244-byte limit (truncate with warning)
6. Create git branch (if git available)
7. Create `specledger/{branch}/` directory
8. Copy spec template to `specledger/{branch}/spec.md`

**JSON Output**:
```json
{
  "BRANCH_NAME": "599-comment-crud",
  "SPEC_FILE": "/path/to/repo/specledger/599-comment-crud/spec.md",
  "FEATURE_NUM": "599"
}
```

**Stop Words** (filtered from description):
i, a, an, the, to, for, of, in, on, at, by, with, from, is, are, was, were, be, been, being, have, has, had, do, does, did, will, would, should, could, can, may, might, must, shall, this, that, these, those, my, your, our, their, want, need, add, get, set

---

## `sl spec setup-plan`

Replaces: `setup-plan.sh`

```
sl spec setup-plan [--json] [--spec <key>]
```

**Behavior**:
1. Resolve feature paths (via ContextDetector)
2. Validate branch is a feature branch
3. Create feature directory if not exists
4. Copy plan template from `.specledger/templates/plan-template.md` to `specledger/{branch}/plan.md`

**JSON Output**:
```json
{
  "FEATURE_SPEC": "/path/to/repo/specledger/598-sdd/spec.md",
  "IMPL_PLAN": "/path/to/repo/specledger/598-sdd/plan.md",
  "SPECS_DIR": "/path/to/repo/specledger/598-sdd",
  "BRANCH": "598-sdd",
  "HAS_GIT": true
}
```

---

## `sl context update`

Replaces: `update-agent-context.sh`

```
sl context update [agent-type] [--spec <key>]
```

**Positional**: Agent type (optional). If omitted, updates all existing agent files.

**Supported agent types**: claude, gemini, copilot, cursor-agent, qwen, opencode, codex, windsurf, kilocode, auggie, roo, codebuddy, qoder, amp, shai, q, bob

**Behavior**:
1. Resolve feature paths and validate plan.md exists
2. Parse plan.md for `**Language/Version**:`, `**Primary Dependencies**:`, `**Storage**:`, `**Project Type**:` fields
3. For each agent file:
   - If file doesn't exist: create from `.specledger/templates/agent-file-template.md` with substitutions
   - If file exists: update tech stack in `## Active Technologies` section
   - Preserve content between `<!-- MANUAL ADDITIONS START -->` and `<!-- MANUAL ADDITIONS END -->` markers
4. Update `## Recent Changes` section with `- {branch}: {description}`
5. Atomic write: temp file → rename

**Agent File Mappings**:
| Agent | File Path |
|-------|-----------|
| claude | `CLAUDE.md` |
| gemini | `GEMINI.md` |
| copilot | `.github/agents/copilot-instructions.md` |
| cursor-agent | `.cursor/rules/specify-rules.mdc` |
| opencode/codex/amp/q/bob | `AGENTS.md` |
| windsurf | `.windsurf/rules/specify-rules.md` |
| Others | `{AGENT_TYPE}.md` |

**Text Output** (default):
```
Updated CLAUDE.md (updated)
Created GEMINI.md (created)
Skipped .cursor/rules/specify-rules.mdc (not found, use --create to initialize)
```
