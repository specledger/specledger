# Research: sl doctor revisited

## Prior Work

| Issue/PR | Summary | Relevance |
|----------|---------|-----------|
| #64 | SDD workflow streamline — bash CLI migration | Introduced `detectStaleFiles()` stub in `pkg/templates/updater.go` (lines 106-119). Builds valid command map but never scans filesystem. |
| #101 | Remove specledger.commit skill | Direct target of US-2. The skill was originally needed because the project-level PostToolUse hook was unreliable. Global hook now handles session capture. |
| #81 | Doctor fails from subdirectories | Direct target of US-3. `findProjectRoot()` exists in `deps.go` but doctor.go uses raw `os.Getwd()`. |
| #96 | Extract ContextDetector to shared package | Overlaps with `findProjectRoot()` extraction. ContextDetector resolves branch→spec; findProjectRoot resolves cwd→project. Different concerns but both benefit from a `pkg/project/` package. |
| #82 | Embedded skill template drift | FR-016 (DONE) — templates synced. FR-019 adds CI guard to prevent recurrence. |
| commit a215636 | Removed project-level PostToolUse hook | Hook "was not reliably firing (especially after mid-session settings changes)". Session capture moved into `/specledger.commit` skill. |
| commit c0dffe7 | Moved inline capture into specledger.commit | Consolidation step — skill became the sole capture mechanism. Now being reversed because global hook is reliable. |

## Decision: Shared `findProjectRoot()` package

**Chosen**: New `pkg/project/root.go` package
**Rationale**: `findProjectRoot()` walks up from cwd looking for `specledger.yaml`. Currently duplicated concern — `deps.go` has it, `doctor.go` needs it, `context_update.go` may need it. A shared package avoids the DRY violation.
**Alternatives considered**:
- Keep in `deps.go` and import from doctor → circular dependency risk (both in `commands` package)
- Put in `pkg/metadata/` → metadata already handles `HasYAMLMetadata()` checks but shouldn't own directory walking logic
- Put in `pkg/cli/project/` → unnecessarily deep nesting

## Decision: Stale file detection approach

**Chosen**: `os.ReadDir()` scan of `.claude/commands/` filtered by `specledger.*.md` glob pattern
**Rationale**: Simple, follows existing pattern. The `specledger.` prefix is CLI-owned (per spec clarification), so we only scan that namespace. No false positives from user commands.
**Alternatives considered**:
- Hash-based comparison (compare file content hashes against embedded) → overkill for deletion detection, useful for drift detection which `diff.go` already handles
- Track installed files in a ledger → complex state management, violates YAGNI

## Decision: --force flag behavior

**Chosen**: `--force` only applies to stale file deletion (not template overwrite, which already forces)
**Rationale**: The `--template` flag already forces overwrite of all non-protected files. `--force` adds the destructive action of deleting stale files. Two separate concerns, two separate flags.
**Alternatives considered**:
- Single `--force` that both overwrites and deletes → conflates two operations
- Interactive prompt per file → poor agent UX, not scriptable

## Decision: CLAUDE.md sentinel migration

**Chosen**: Refactor `updater.go` to use `MergeSentinelSection()` with two sentinel blocks
**Rationale**: `MergeSentinelSection()` already exists in `pkg/cli/playbooks/merge.go` and is tested. The current `<!-- MANUAL ADDITIONS -->` pattern is a bespoke reimplementation of the same concept. Unifying reduces code paths.
**Sentinel blocks**:
1. `# >>> specledger-generated` — session-start guidance (managed by `sl doctor --template`)
2. `# >>> specledger-context` — Active Technologies (managed by `sl context update`)
3. User content lives outside both sentinel blocks

**Alternatives considered**:
- Keep `<!-- MANUAL ADDITIONS -->` and nest sentinels inside → already happening, creates confusion
- Single sentinel block for everything → can't independently update technologies vs guidance

## Decision: Hook opt-out mechanism

**Chosen**: `session_capture: false` in `specledger.yaml`
**Rationale**: Project-level config keeps the opt-out visible to all team members. `sl auth hook --remove` sets this flag. `sl auth hook --install` clears it. `sl auth login` checks it.
**Alternatives considered**:
- Global `~/.specledger/config.yaml` → per-user, not visible to team
- Environment variable → ephemeral, easily forgotten
- Both global and project-level → complexity without clear benefit (YAGNI)

## Decision: Comment UUID prefix matching

**Chosen**: Load all comment IDs for current spec, filter by prefix, error on ambiguity
**Rationale**: Comment lists are small (typically <100 per spec). No performance concern. Same pattern as git commit hash resolution.
**Alternatives considered**:
- Server-side prefix search → requires API changes, overkill
- Client-side trie → unnecessary optimization for small datasets

## Decision: Checklist template sourcing (FR-018)

**Chosen**: Footer hint from `sl spec create` pointing to `.specledger/templates/checklist-template.md`
**Rationale**: The template is already copied to the project by `sl doctor --template`. Agents just need to know where to read it. A dedicated `sl spec checklist` command would violate YAGNI — the file is already on disk.
**Alternatives considered**:
- `sl spec checklist` command that outputs the template → adds a command for something a file read achieves
- Inline template in command prompt → current state, causes drift
