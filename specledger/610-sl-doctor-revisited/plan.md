# Implementation Plan: sl doctor revisited

**Branch**: `610-sl-doctor-revisited` | **Date**: 2026-03-16 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `specledger/610-sl-doctor-revisited/spec.md`

## Summary

Fix and enhance `sl doctor` to: (1) complete the stubbed `detectStaleFiles()` implementation, (2) remove the deprecated `specledger.commit` skill from all layers, (3) fix subdirectory resolution by extracting `findProjectRoot()` to a shared package, (4) add `--check` and `--force` flags, (5) add `NEXT_STEPS` to scaffold command JSON output, (6) improve onboarding constitution prompts, (7) migrate CLAUDE.md to sentinel-based management, and (8) various UX improvements (comment UUID prefix matching, hook opt-out, decision log, `sl issue ready` in implement).

## Technical Context

**Language/Version**: Go 1.24.2
**Primary Dependencies**: Cobra (CLI), Go embed FS, `pkg/cli/playbooks/` (template management), `pkg/templates/` (update logic)
**Storage**: JSONL files per spec (`specledger/<spec>/issues.jsonl`), YAML config (`specledger.yaml`), embedded filesystem (`pkg/embedded/`)
**Testing**: `go test` with table-driven tests (unit), binary invocation via `exec.Command` (integration in `tests/integration/`)
**Target Platform**: macOS/Linux CLI
**Project Type**: Single Go module
**Performance Goals**: N/A — CLI tool, instant response expected
**Constraints**: No new dependencies; changes must be backward compatible with existing projects
**Scale/Scope**: ~27 FRs across 9 user stories; primarily refactoring existing code + filling in stubs

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] **Specification-First**: Spec.md complete with 9 prioritized user stories (P1-P3), 27 FRs, edge cases
- [x] **Test-First**: Test strategy defined — extend `tests/integration/doctor_test.go`, new unit tests for `detectStaleFiles()` and `findProjectRoot()` shared package
- [x] **Code Quality**: `make lint` (golangci-lint v2), `make fmt`
- [x] **UX Consistency**: Acceptance scenarios defined per user story with Given/When/Then
- [x] **Performance**: N/A — CLI operations, no performance targets needed
- [x] **Observability**: CLI uses structured JSON output (`--json`) for machine consumption; human mode uses color-coded status
- [ ] **Issue Tracking**: Epic to be created during task generation phase

**Complexity Violations**: None — all changes follow existing patterns and stay within the current architecture.

## Project Structure

### Documentation (this feature)

```text
specledger/610-sl-doctor-revisited/
├── plan.md              # This file
├── research.md          # Phase 0 output (below)
├── quickstart.md        # Phase 1 output — user scenarios for E2E validation
└── tasks.md             # Phase 2 output (/specledger.tasks)
```

Note: No `data-model.md` or `contracts/` needed — this feature modifies existing CLI internals without introducing new entities or APIs.

### Source Code (repository root)

```text
pkg/
├── cli/
│   ├── commands/
│   │   ├── doctor.go          # MODIFY: add --check, --force; use shared findProjectRoot()
│   │   ├── deps.go            # MODIFY: remove local findProjectRoot(), import shared
│   │   ├── comment.go         # MODIFY: add UUID prefix resolution
│   │   ├── auth.go            # MODIFY: add hook opt-out config check
│   │   ├── spec_create.go     # MODIFY: add NEXT_STEPS to JSON output + footer hints
│   │   └── spec_setup_plan.go # MODIFY: add NEXT_STEPS to JSON output + footer hints
│   ├── context/
│   │   └── updater.go         # MODIFY: refactor to use MergeSentinelSection()
│   ├── hooks/
│   │   └── claude.go          # MODIFY: check opt-out config before hook install
│   └── playbooks/
│       ├── copy.go            # MODIFY: return protected file list in result
│       └── merge.go           # (existing — used by refactored updater.go)
├── project/                   # NEW PACKAGE: shared findProjectRoot()
│   └── root.go                # Extract from deps.go, reuse in doctor.go
├── templates/
│   └── updater.go             # MODIFY: complete detectStaleFiles() stub, add --force delete
├── issues/
│   └── id.go                  # MODIFY: add PrefixMatch() for short UUID resolution
└── embedded/
    └── templates/
        ├── manifest.yaml      # MODIFY: remove specledger.commit entry
        └── specledger/
            ├── commands/
            │   └── specledger.commit.md    # DELETE
            │   └── specledger.onboard.md   # MODIFY: improve constitution prompts
            │   └── specledger.checkpoint.md # MODIFY: add Decision Log section
            │   └── specledger.implement.md  # MODIFY: use sl issue ready
            └── .specledger/
                └── memory/
                    └── constitution.md     # MODIFY: add KISS principle, remove /specledger.commit ref

.claude/commands/specledger.commit.md      # DELETE (local copy)
CLAUDE.md                                  # MODIFY: sentinel migration via FR-017
docs/design/README.md                      # MODIFY: remove commit from workflow diagram
docs/design/commands.md                    # MODIFY: audit for commit references

tests/
├── integration/
│   └── doctor_test.go         # MODIFY: add subdirectory + stale file test cases
└── issues/
    └── (existing tests)       # May need prefix match tests
```

**Structure Decision**: Single Go module. New `pkg/project/` package for shared utilities (`findProjectRoot`). All other changes modify existing files.

## Complexity Tracking

No violations. All changes follow existing patterns:
- `pkg/project/root.go` extracts existing code from `deps.go` — no new abstraction
- Sentinel merge reuses existing `MergeSentinelSection()` from `playbooks/merge.go`
- UUID prefix matching follows the same pattern as git commit hash resolution

## Phase Breakdown

### Phase 1: Core Doctor Fixes (P1 — US-1, US-2, US-3)

**Goal**: Fix the three P1 items — stale detection, commit removal, subdirectory resolution.

1. **Extract `findProjectRoot()` to `pkg/project/root.go`**
   - Move from `deps.go` lines 960-985
   - Update `deps.go` to import from `pkg/project`
   - Update `doctor.go` to use shared `findProjectRoot()` instead of `os.Getwd()`
   - Unit test in `pkg/project/root_test.go`

2. **Complete `detectStaleFiles()` in `pkg/templates/updater.go`**
   - Replace stub (lines 106-119) with actual `os.ReadDir()` scan
   - Filter for `specledger.*.md` pattern
   - Compare against manifest commands
   - Populate `result.Stale` field
   - Add `--force` flag to `doctor.go` for deletion
   - Unit test with table-driven cases

3. **Remove `specledger.commit` from all layers**
   - Delete `pkg/embedded/templates/specledger/commands/specledger.commit.md`
   - Remove entry from `manifest.yaml`
   - Delete `.claude/commands/specledger.commit.md`
   - Remove constitution reference (line 66)
   - Update `docs/design/README.md` (lines 63, 77)
   - Audit `docs/design/commands.md`

4. **Add `--check` flag to `sl doctor`**
   - Human-readable dry-run: reports status, no prompts, no changes
   - Exits non-zero if updates needed
   - Works from subdirectories (uses shared `findProjectRoot()`)

5. **Add `--force` flag behavior**
   - With `--template --force`: delete stale files + confirmation message
   - Without `--force`: warn only (current default behavior, now actually working)

6. **Show protected files in `--template` output**
   - Modify `CopyPlaybooks()` to track skipped protected files
   - Show in doctor output: "Skipped N protected files: constitution.md, AGENTS.md"

### Phase 2: Scaffold & CLAUDE.md Improvements (P2 — US-3, US-4, US-9)

**Goal**: Improve CLI output for agent consumption and migrate CLAUDE.md management.

7. **Add `NEXT_STEPS` to `sl spec create --json`**
   - Add field to JSON output struct
   - Include: read spec template before writing
   - Add footer hint in human mode

8. **Add `NEXT_STEPS` to `sl spec setup-plan --json`**
   - Add field to JSON output struct
   - Include: read plan template, checklist template, constitution
   - Add footer hint in human mode

9. **Migrate CLAUDE.md to sentinel-based management (FR-017)**
   - Refactor `pkg/cli/context/updater.go` to use `MergeSentinelSection()`
   - Replace `<!-- MANUAL ADDITIONS START/END -->` with sentinel blocks
   - Use `# >>> specledger-generated` for session-start guidance
   - Use `# >>> specledger-context` for Active Technologies
   - Preserve user content (Pre-push Checklist) outside sentinel blocks
   - Add `sl doctor --check` recommendation to managed section (FR-021)

10. **Hook opt-out config (FR-022)**
    - Check `specledger.yaml` for `session_capture: false`
    - `sl auth hook --remove` persists opt-out to config
    - `sl auth hook --install` clears opt-out
    - `sl auth login` respects opt-out

### Phase 3: Template & UX Improvements (P3 — US-5, US-6, US-7, US-8)

**Goal**: Improve onboarding, checkpoint, implement, and comment UX.

11. **Improve onboarding constitution prompts (FR-014, FR-015)**
    - Modify `specledger.onboard.md` constitution step
    - Guide toward design principles, not tech inventory
    - Provide example categories: testing philosophy, code standards, deployment strategy

12. **Add Decision Log to checkpoint template (FR-024)**
    - Add `### Decision Log` section to `specledger.checkpoint.md`
    - Structure: What, Why, Impact level, Artifacts affected

13. **Use `sl issue ready` in implement template (FR-025)**
    - Replace `sl issue list --status in_progress` with `sl issue ready` for task selection
    - Keep resume logic using in-progress check

14. **Comment UUID prefix matching (FR-027)**
    - Add `PrefixMatch()` to `pkg/issues/` or `pkg/cli/comment/`
    - Load all comment IDs, find prefix matches
    - Error on 0 or >1 matches
    - Apply to resolve, show, reply subcommands

15. **Add KISS to constitution (FR-026)**
    - Add principle to embedded constitution template
    - Embedded templates are source of truth; users customize via git

### Phase 4: CI & Quality Gates (P2)

16. **CI template drift guard (FR-019)**
    - Add CI step: `make build && ./bin/sl doctor --template && git diff --exit-code`
    - Cover `.claude/commands/`, `.claude/skills/`, `.specledger/templates/`

17. **Checklist template sourcing (FR-018)**
    - Add footer hint from `sl spec create` pointing to `.specledger/templates/checklist-template.md`
    - Remove hardcoded checklist structure from `specledger.specify.md` command prompt
    - Agent reads template file instead

### Phase 5: Testing

18. **Unit tests**
    - `pkg/project/root_test.go` — findProjectRoot with subdirs, no project, filesystem root
    - `pkg/templates/updater_test.go` — detectStaleFiles with stale files, no stale, no commands dir, custom files
    - `pkg/issues/id_test.go` or comment prefix match tests — exact match, prefix match, ambiguous, not found

19. **Integration tests**
    - Extend `tests/integration/doctor_test.go`:
      - `TestDoctorFromSubdirectory` — run from pkg/cli/, verify success
      - `TestDoctorStaleFileDetection` — place extra specledger.*.md, verify warning
      - `TestDoctorStaleFileForceDelete` — verify --force deletes stale files
      - `TestDoctorCheckFlag` — verify --check exits non-zero when outdated, zero when current
      - `TestDoctorProtectedFileDisplay` — verify protected files listed

20. **Quickstart scenario validation**
    - Translate quickstart.md scenarios to integration test assertions
    - File gaps as issues for future E2E directory creation

## Previous Work

- **#64**: feat: SDD workflow streamline — introduced `detectStaleFiles()` stub that was never completed
- **#101**: chore: remove specledger.commit skill — directly addressed by this spec (US-2)
- **#81**: bug: sl doctor fails from subdirectories — directly addressed (US-3)
- **#90**: Improve agent prompts — directly addressed (US-4)
- **#91**: Onboarding too technical — directly addressed (US-5)
- **#96**: Extract ContextDetector to shared package — overlaps with `findProjectRoot()` extraction (FR-010)
- **#82**: Improve embedded skill templates — FR-016 (DONE), FR-019 (CI guard)
- **#84**: Decision log in checkpoint — directly addressed (US-6)
- **#92**: Use sl issue ready — directly addressed (US-7)
- **#106**: Comment short UUID prefix — directly addressed (US-8)

## External Dependencies

None. All changes are internal to the `sl` CLI binary. No new Go dependencies required.

## Risk Assessment

| Risk | Mitigation |
|------|-----------|
| CLAUDE.md sentinel migration breaks existing projects | Test migration path: existing manual additions content preserved outside sentinels |
| Removing specledger.commit while hook is unreliable | Global hook at `~/.claude/settings.json` is working (installed by `sl auth login`); spec US-9 adds opt-out for users who don't want it |
| findProjectRoot() extraction breaks deps command | Integration tests for deps already exist in `tests/integration/deps_test.go` — run before and after |
| Stale file deletion with --force is destructive | Only deletes `specledger.*` prefix files; custom commands never touched; git recovery always available |
