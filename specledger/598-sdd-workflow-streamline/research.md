# Research: SDD Workflow Streamline

**Branch**: `598-sdd-workflow-streamline` | **Date**: 2026-03-01

## Prior Work

### Directly Related Specs

| Spec | Status | Relevance |
|------|--------|-----------|
| `136-revise-comments` | Mostly complete (3 open edge-case tasks) | **Primary**: `sl revise` command, PostgREST client, comment types, prompt template, agent launcher. Direct refactoring target for `sl comment` extraction. |
| `597-agent-model-config` | Complete | Agent config, `sl config` command, `BuildEnv()` in launcher. Config merge pattern reusable for new commands. |
| `596-doctor-version-update` | Complete | `sl doctor` template update flow. Foundation for `--template` stale command detection. |
| `595-issue-tree-ready` | Complete | `sl issue tree/ready` — model for Data CRUD pattern with tree rendering. |
| `591-issue-tracking-upgrade` | Complete | Issue JSONL store, context detection (`ContextDetector`). Foundation for `sl spec info` context resolution. |

### Open Issues (136-revise-comments)

| Issue ID | Title | Status | Relevance |
|----------|-------|--------|-----------|
| SL-5af4b2 | Revise Comments CLI Command | open | Superseded — comment CRUD moves to `sl comment` |
| SL-cafbb1 | Implement edge case handling | open | Carry forward to `sl comment` |
| SL-41d979 | Polish: Error handling, edge cases | open | Carry forward to `sl comment` + refactored revise |
| SL-9343e6 | Detect local branch behind remote | open | Relevant to `sl revise` launcher cleanup |

## Research Findings

### R1: Supabase `review_comments` Table Schema

**Decision**: Use existing schema as-is for `sl comment` CLI.
**Rationale**: Schema is stable, in production, with working PostgREST queries.

**Actual columns** (from `information_schema.columns`):

| Column | Type | Nullable | Default | Notes |
|--------|------|----------|---------|-------|
| id | uuid | NO | uuid_generate_v4() | PK |
| change_id | uuid | NO | — | FK → changes |
| content | text | NO | — | Comment body |
| start_line | integer | YES | — | Line range start |
| line | integer | YES | — | Line range end |
| is_resolved | boolean | NO | false | Resolution status |
| author_id | uuid | NO | — | FK → auth.users |
| created_at | timestamptz | YES | now() | |
| file_path | text | NO | — | Artifact file path |
| selected_text | text | YES | — | Highlighted passage |
| updated_at | timestamptz | YES | now() | |
| author_name | text | YES | — | Denormalized |
| author_email | text | YES | — | Denormalized |
| parent_comment_id | uuid | YES | — | Self-ref for threads |
| is_ai_generated | boolean | YES | — | **New** (not in 136 data model) |
| triggered_by_user_id | uuid | YES | — | **New** (not in 136 data model) |

**Relationship chain**: `projects → specs → changes → review_comments`

**Thread model**: `parent_comment_id` is self-referential. Replies have `parent_comment_id` set, no `selected_text`. Top-level comments have `parent_comment_id = NULL`.

**New columns** discovered (not in 136-revise-comments data model):
- `is_ai_generated` — tracks AI-authored comments
- `triggered_by_user_id` — tracks who triggered AI comment generation

These should be included in `sl comment` types but are read-only from CLI perspective.

### R2: Comment Threading Sample Data

**Decision**: CLI displays threads as parent → replies (flat, not nested).
**Rationale**: All production threads are 1-level deep (parent + direct replies). No nested reply-to-reply observed.

Sample thread:
```
Parent: "need clarify" (TrungDinh, on tasks.md, selected "Type")
  └── Reply: "this is great" (Ngoc Tran)
  └── Reply: "I think we need to adjust more" (Ngoc Tran)
```

AI-generated replies exist in production:
```
Parent: "..." (human reviewer)
  └── Reply: "Added US9: --summary flag..." (Claude Code, is_ai_generated implied by author_name)
```

### R3: Existing PostgREST Query Pattern (from `pkg/cli/revise/client.go`)

**Decision**: Extract PostgREST client from `pkg/cli/revise/` into a shared `pkg/cli/comment/` package.
**Rationale**: The 4-step query chain (project→spec→change→comments) is reusable. Comment CRUD should not depend on the revise package.

**Current query chain**:
1. `GET /projects?repo_owner=eq.{o}&repo_name=eq.{n}` → project_id
2. `GET /specs?project_id=eq.{pid}&spec_key=eq.{key}` → spec_id
3. `GET /changes?spec_id=eq.{sid}` → change_id
4. `GET /review_comments?change_id=eq.{cid}&is_resolved=eq.false&parent_comment_id=is.null` → comments
5. `GET /review_comments?change_id=eq.{cid}&parent_comment_id=not.is.null` → replies (separate call)

**Auth pattern**: Bearer token + apikey header. Auto-refresh on 401/PGRST303 via `doWithRetry()`.

### R4: `sl revise` Architecture Analysis

**Decision**: Refactor `sl revise` into clean layers: extract comment CRUD to `pkg/cli/comment/`, keep launcher logic in `pkg/cli/commands/revise.go`.
**Rationale**: Current `revise.go` is ~1100 lines mixing data operations (fetch/resolve comments) with TUI (huh multi-select, lipgloss styling) and launcher logic (agent spawn, git stash). Clean layering per D2.

**Current architecture** (`pkg/cli/commands/revise.go` + `pkg/cli/revise/`):

| Concern | Current Location | Target Location |
|---------|-----------------|-----------------|
| Comment types/models | `pkg/cli/revise/types.go` | `pkg/cli/comment/types.go` |
| PostgREST client | `pkg/cli/revise/client.go` | `pkg/cli/comment/client.go` |
| Comment fetch/resolve | `pkg/cli/revise/client.go` | `pkg/cli/comment/client.go` |
| Prompt rendering | `pkg/cli/revise/prompt.go` | Keep in `pkg/cli/revise/` (revise-specific) |
| Editor integration | `pkg/cli/revise/editor.go` | Keep in `pkg/cli/revise/` (revise-specific) |
| Automation/fixtures | `pkg/cli/revise/automation.go` | Keep in `pkg/cli/revise/` (revise-specific) |
| CLI command (revise) | `pkg/cli/commands/revise.go` | Simplified launcher using `pkg/cli/comment/` |
| CLI command (comment) | — | **New**: `pkg/cli/commands/comment.go` |

### R5: Clarify Command Integration Point

**Decision**: Replace `sl revise --summary` call in clarify with `sl comment list --status open --json`.
**Rationale**: `sl comment list` is the canonical Data CRUD interface. `--summary` was a bridge solution.

**Current** (clarify.md step 2): Calls `sl revise --summary` for compact comment listing.
**Target**: Calls `sl comment list --status open --json` for structured comment data.

### R6: Bash Script Analysis for CLI Replacement

**Decision**: Create 4 new `sl` commands to replace 6 bash scripts.
**Rationale**: Cross-platform support (D10/D16), eliminate `jq`/`bash` dependencies.

#### Script → Command Mapping

| Script | Proposed Command | Complexity | Notes |
|--------|-----------------|------------|-------|
| `common.sh` | Absorbed into internal packages | N/A | Shared logic (repo root, branch detect, feature paths) → `internal/ref/` + `pkg/cli/` |
| `check-prerequisites.sh` | `sl spec info` | Medium | Path resolution + prerequisite validation + JSON output |
| `create-new-feature.sh` | `sl spec create` | High | Branch naming, stop-word filter, collision prevention, git branch create, template copy |
| `setup-plan.sh` | `sl spec setup-plan` | Low | Template copy + JSON path output |
| `update-agent-context.sh` | `sl context update` | High | Parse plan.md, update 17+ agent file formats, preserve manual additions |
| `adopt-feature-branch.sh` | Superseded by D9 | N/A | ContextDetector fallback chain replaces branch-map.json |

#### External Dependencies to Eliminate

| Dependency | Used By | Go Replacement |
|------------|---------|----------------|
| `jq` | common.sh, adopt-feature-branch.sh | `encoding/json` stdlib |
| `grep` | check-prerequisites.sh, update-agent-context.sh | `regexp`, `strings`, `bufio` |
| `sed` | update-agent-context.sh | `strings.Replace`, `regexp` |
| `git` CLI | create-new-feature.sh | `go-git/v5` (already a dependency) |

#### Key Implementation Details from Script Analysis

**Branch number generation** (`create-new-feature.sh`):
- Scans `specledger/*/` directories for highest numeric prefix
- Uses `10#$number` for base-10 (prevents octal interpretation)
- Stop-word list: 41 common words filtered from description
- Acronym detection: uppercase words in original description
- GitHub 244-byte branch name limit enforcement
- **Gap**: Only checks local dirs, not remote branches ([#46](https://github.com/specledger/specledger/issues/46))

**Agent context update** (`update-agent-context.sh`):
- Parses `**FieldName**: value` patterns from plan.md
- 17 agent file mappings (claude, gemini, copilot, cursor, etc.)
- Template substitution with `[PLACEHOLDER]` patterns
- Manual additions preserved between `<!-- MANUAL ADDITIONS START/END -->` markers
- Language-specific command generation (Python, Rust, JS/TS, Go, etc.)
- Atomic file updates via temp file + mv

**Prerequisite validation** (`check-prerequisites.sh`):
- Checks feature dir, plan.md, tasks.md existence
- Optional doc discovery: research.md, data-model.md, contracts/, quickstart.md
- JSON output mode with `FEATURE_DIR` and `AVAILABLE_DOCS` fields

### R7: `sl-comment` Skill Design

**Decision**: Create lean skill following `sl-issue-tracking` pattern.
**Rationale**: Skills teach agents *when* and *how* to use CLI tools (D5). Progressive loading triggered by comment references in AI commands.

**Skill outline** (high level):
- Trigger: AI command mentions "review comments", "comment", or references `sl comment`
- Content: Usage patterns for `sl comment list/show/reply/resolve`
- Pattern: Same as `sl-issue-tracking` skill structure
- Key teaching: When to use `--json` for structured output, thread context for replies, resolution reasons

### R8: Clarify → Comment Integration (High Level)

**Decision**: Clarify absorbs revise's comment-processing intelligence.
**Rationale**: Per D4, `/specledger.clarify` becomes the single AI command for spec refinement (ambiguity + comments).

**Changes to clarify** (high level only — details in implementation):
1. Replace `sl revise --summary` call with `sl comment list --status open --json`
2. Structured JSON enables richer comment analysis (thread context, selected_text, file_path grouping)
3. After addressing each comment: `sl comment reply <id> --content "..."` + `sl comment resolve <id> --reason "..."`
4. `sl-comment` skill progressively loaded when clarify references comments

### R9: `update-agent-context.sh` CLAUDE.md Output Quality

**Decision**: The Go replacement (`sl context update`) must NOT replicate the bash script's append-only behavior.
**Rationale**: Running the bash script on this plan produced a degraded CLAUDE.md with critical quality issues.

**Observed problems** after running `update-agent-context.sh claude` on the 598 plan:

| Problem | Example | Root Cause |
|---------|---------|------------|
| **Duplicate entries** | 6 lines in "Active Technologies" all saying "Go 1.24.2 + Cobra" with minor variations | Script appends without reading existing entries |
| **Near-identical lines** | `597-issue-create-fields` produced 4 separate entries that overlap | No deduplication logic |
| **Stale "Recent Changes"** | Only shows `597-issue-create-fields`, missing `598-sdd-workflow-streamline` | Script failed to add entry to Recent Changes section |
| **Branch-tag noise** | Every line tagged `(597-issue-create-fields)` or `(598-sdd-workflow-streamline)` | Branch name used as differentiator instead of useful description |
| **Unbounded growth** | Active Technologies grows by 2-4 lines per feature, never pruned | No cap or rotation logic |

**Actual CLAUDE.md "Active Technologies" after script ran**:
```
- Go 1.24.2 + Cobra (CLI), YAML v3 (config), JSONL (storage) (597-issue-create-fields)
- File-based JSONL at `specledger/<spec>/issues.jsonl` (597-issue-create-fields)
- Go 1.24.2 + Cobra (CLI), go-git v5, YAML v3, Supabase (GoTrue, PostgREST) (597-issue-create-fields)
- File-based JSONL for issues (597-issue-create-fields)
- Go 1.24.2 + Cobra (CLI), go-git v5 (git), Bubble Tea + Bubbles + Lipgloss (TUI), YAML v3 (config) (598-sdd-workflow-streamline)
- Supabase PostgREST (review_comments table), file-based (JSONL for issues, YAML for config) (598-sdd-workflow-streamline)
```

**Requirements for `sl context update` Go replacement**:

| Behavior | Bash (don't copy) | Go (implement instead) |
|----------|-------------------|----------------------|
| Active Technologies | Blind append | Read existing entries, deduplicate, merge supersets |
| Recent Changes | Sometimes missed | Always add when Active Technologies changes |
| Entry format | Branch name tag | Concise tech description |
| Growth | Unbounded | Cap Active Technologies to ~5 entries; Recent Changes to last 3 |
| Idempotency | Running twice = duplicates | Running twice = no change |

## Alternatives Considered

### A1: Keep Comment CRUD in `pkg/cli/revise/`
**Rejected**: Violates D2 (CLI = data CRUD, separate from launcher logic). `sl comment` needs to be independently usable.

### A2: Build `sl comment` from scratch without reusing revise code
**Rejected**: `pkg/cli/revise/client.go` has battle-tested PostgREST client with auth retry, token refresh, and correct query patterns. Extract and refactor, don't rewrite.

### A3: Port bash scripts using `os/exec` to call `git` CLI
**Rejected**: `go-git/v5` is already a dependency. Direct Go implementation is more portable and testable. Exception: Some git operations may still need CLI fallback (e.g., `git checkout -b` for branch creation with tracking).

### A4: Merge `sl spec` subcommands into existing commands
**Rejected**: `sl spec info/create/setup-plan` forms a coherent command group for specification lifecycle. Adding to `sl init` or `sl doctor` would muddy their responsibilities.
