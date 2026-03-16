# Requirements Checklist: sl doctor revisited

**Purpose**: Validate implementation completeness against spec FRs and plan phases
**Created**: 2026-03-16
**Feature**: [spec.md](../spec.md)

## Phase 1: Core Doctor Fixes (P1)

- [ ] CHK001 `findProjectRoot()` extracted to `pkg/project/root.go` and used by both `deps.go` and `doctor.go`
- [ ] CHK002 `detectStaleFiles()` uses `os.ReadDir()` to scan `.claude/commands/` for `specledger.*.md` not in manifest
- [ ] CHK003 `--force` flag deletes stale `specledger.*` files with per-file confirmation message
- [ ] CHK004 Without `--force`, stale files are warned about but never deleted
- [ ] CHK005 Custom commands (no `specledger.` prefix) are never flagged or deleted
- [ ] CHK006 `specledger.commit.md` removed from: embedded templates, manifest.yaml, `.claude/commands/`
- [ ] CHK007 Constitution line 66 (`/specledger.commit` reference) removed
- [ ] CHK008 `docs/design/README.md` lines 63, 77 updated (commit removed from workflow + escape hatches)
- [ ] CHK009 `docs/design/commands.md` audited for remaining commit references
- [ ] CHK010 `--check` flag: human-readable dry-run, no prompts, no changes, non-zero exit if updates needed
- [ ] CHK011 Protected files listed in `--template` output ("Skipped N protected files: ...")

## Phase 2: Scaffold & CLAUDE.md (P2)

- [ ] CHK012 `sl spec create --json` includes `NEXT_STEPS` field with template read instructions
- [ ] CHK013 `sl spec setup-plan --json` includes `NEXT_STEPS` field with template/checklist/constitution read instructions
- [ ] CHK014 `sl spec create` human mode prints footer hint for template path
- [ ] CHK015 `sl spec setup-plan` human mode prints footer hint
- [ ] CHK016 CLAUDE.md refactored: `<!-- MANUAL ADDITIONS -->` replaced with sentinel blocks
- [ ] CHK017 `# >>> specledger-generated` block contains session-start `sl doctor --check` guidance
- [ ] CHK018 Active Technologies managed via `# >>> specledger-context` sentinel (or equivalent)
- [ ] CHK019 Pre-push Checklist preserved outside sentinel blocks after migration
- [ ] CHK020 `sl auth hook --remove` persists opt-out (`session_capture: false`) in `specledger.yaml`
- [ ] CHK021 `sl auth login` respects opt-out — does NOT re-install hook
- [ ] CHK022 `sl auth hook --install` clears opt-out

## Phase 3: Template & UX (P3)

- [ ] CHK023 `specledger.onboard.md` constitution step asks for design principles, not tech inventory
- [ ] CHK024 Example categories provided: testing philosophy, code standards, deployment strategy, error handling
- [ ] CHK025 `specledger.checkpoint.md` includes `### Decision Log` with What/Why/Impact/Artifacts fields
- [ ] CHK026 `specledger.implement.md` uses `sl issue ready` for task selection (not `sl issue list --status in_progress`)
- [ ] CHK027 Comment subcommands (resolve, show, reply) accept short UUID prefixes
- [ ] CHK028 Ambiguous prefix returns error listing matching full UUIDs
- [ ] CHK029 Non-matching prefix returns "comment not found" with `sl comment list` suggestion
- [ ] CHK030 Constitution includes KISS principle (FR-026)

## Phase 4: CI & Quality Gates

- [ ] CHK031 CI includes drift guard: `make build && ./bin/sl doctor --template && git diff --exit-code`
- [ ] CHK032 Drift guard covers `.claude/commands/`, `.claude/skills/`, `.specledger/templates/`
- [ ] CHK033 `specledger.specify.md` references checklist template file instead of inline checklist structure

## CLI Changes

- [ ] CLI-001 `--check` flag: Template Management pattern — dry-run variant of existing `sl doctor`
- [ ] CLI-002 `--force` flag: Template Management pattern — destructive variant for stale cleanup
- [ ] CLI-003 Error messages from `findProjectRoot()` include: what failed, why, suggested fix
- [ ] CLI-004 `--check` human output uses footer hints for next step (`sl doctor --update --template`)
- [ ] CLI-005 `--check` and `--json` outputs are complete and pipeable
- [ ] CLI-006 Errors to stderr, data to stdout

## Template & Sync

- [ ] SYNC-001 After deleting `specledger.commit.md` from embedded, `sl doctor --template` correctly reports it as stale on existing projects
- [ ] SYNC-002 Runtime copies in `.claude/commands/` and `.claude/skills/` match embedded source after `sl doctor --template`
- [ ] SYNC-003 Modified skill/command templates (`onboard`, `checkpoint`, `implement`) synced between embedded and runtime

## Architecture Compliance

- [ ] ARCH-001 Agent Owns Outcomes — `--force` requires explicit user flag; no auto-deletion without consent
- [ ] ARCH-002 Cross-layer calls follow patterns: L1 CLI (`sl doctor`) manages templates; L2 commands read them
- [ ] ARCH-003 No Supabase changes in this spec — local stack not affected

## Testing

- [ ] TEST-001 `pkg/project/root_test.go` — subdirectory resolution, no project found, filesystem root
- [ ] TEST-002 `pkg/templates/updater_test.go` — stale detection: stale files, no stale, no commands dir, custom files
- [ ] TEST-003 Comment prefix match tests — exact, prefix, ambiguous, not found
- [ ] TEST-004 `tests/integration/doctor_test.go` — subdirectory, stale detection, --force, --check, protected files
- [ ] TEST-005 All existing tests pass (`make test && make test-integration`)

## Notes

- Constitution post-design re-check passed: YAGNI, DRY, Shortest Path, Simplicity, Contract-First Testing, Quickstart-Driven, Fail Fast all satisfied
- No data-model.md or contracts/ needed — internal CLI refactoring only
- No new Go dependencies required
