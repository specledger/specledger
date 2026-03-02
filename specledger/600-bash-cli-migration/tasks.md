---
description: "Task list for Bash Script to Go CLI Migration"
---

# Tasks Index: Bash Script to Go CLI Migration

Task list for replacing 6 bash scripts with 4 Go CLI commands for cross-platform support.

## Feature Tracking

* **Issue Epic ID**: `SL-7f4ea3`
* **User Stories Source**: `specledger/600-bash-cli-migration/spec.md`
* **Research Inputs**: `specledger/600-bash-cli-migration/research.md`
* **Planning Details**: `specledger/600-bash-cli-migration/plan.md`
* **Data Model**: N/A (file-based operations only)
* **Contract Definitions**: N/A (CLI commands, not API endpoints)

## Issue Query Commands

Use the `sl issue` CLI to query and manipulate issues:

```bash
# Find all open tasks for this feature
sl issue list --label spec:600-bash-cli-migration --status open

# Find ready tasks (no blocking dependencies)
sl issue list --label spec:600-bash-cli-migration --status open --json | jq '.[] | select(.blocked_by | length == 0)'

# View issues by component
sl issue list --label 'component:cli' --label 'spec:600-bash-cli-migration'

# View all phases (feature-type issues)
sl issue list --type feature --label 'spec:600-bash-cli-migration'

# Show dependency tree for epic
sl issue show SL-7f4ea3 --tree

# Get specific task details
sl issue show SL-958887
```

## Tasks and Phases Structure

This feature follows a 3-level hierarchy:

* **Epic**: SL-7f4ea3 → Bash Script to Go CLI Migration
* **Features/Phases**: Implementation phases (setup, foundational, user stories, polish)
  * Each phase = a user story or technical milestone
* **Tasks**: Implementation tasks within each phase

## Convention Summary

| Type    | Description                  | Labels                                      |
| ------- | ---------------------------- | ------------------------------------------- |
| epic    | Full feature epic            | `spec:600-bash-cli-migration`               |
| feature | Implementation phase / story | `phase:[name]`, `story:[US#]`               |
| task    | Implementation task          | `component:[area]`, `requirement:[FR-###]`  |

## Phase Overview

| Phase | Issue ID | Type | Priority | Story | Status |
|-------|----------|------|----------|-------|--------|
| Setup | SL-e12f56 | feature | 1 | - | Open |
| Foundational | SL-4f80fb | feature | 0 | US0 | Open |
| US1: sl spec info | SL-043360 | feature | 1 | US1 | Open |
| US2: sl spec create | SL-d2f3bc | feature | 1 | US2 | Open |
| US3: sl spec setup-plan | SL-61161a | feature | 2 | US3 | Open |
| US4: sl context update | SL-1ab18f | feature | 1 | US4 | Open |
| US5: Cross-Platform | SL-92fb0b | feature | 2 | US5 | Open |
| Polish | SL-9a0e47 | feature | 3 | - | Open |

---

## Phase 1: Setup

**Issue**: SL-e12f56 | **Priority**: 1 | **Status**: Open

**Purpose**: Verify project structure and dependencies before implementation

**Tasks**:

| ID | Title | Status | Component |
|----|-------|--------|-----------|
| SL-41daff | Verify pkg/cli/spec/ directory structure | Open | infra |
| SL-aa3426 | Verify pkg/cli/context/ directory structure | Open | infra |
| SL-8a8228 | Verify Cobra and go-git/v5 dependencies | Open | infra |
| SL-aad91f | Verify template files exist | Open | infra |

**Independent Test**: All directories exist, dependencies installed, templates accessible

**View tasks**:
```bash
sl issue show SL-e12f56 --tree
```

---

## Phase 2: Foundational - Core Packages

**Issue**: SL-4f80fb | **Priority**: 0 (Critical) | **Status**: Open

**Purpose**: Create core packages that ALL user stories depend on. MUST complete before any user story implementation.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

**Tasks**:

| ID | Title | Status | Component | Requirements |
|----|-------|--------|-----------|--------------|
| SL-e8c313 | Create pkg/cli/spec/detector.go | Open | cli | FR-001 |
| SL-6d857e | Create pkg/cli/spec/paths.go | Open | cli | FR-001 |
| SL-03c3fe | Create pkg/cli/spec/branch.go | Open | cli | FR-003, FR-004 |

**Key Deliverables**:
- `FeatureContext` struct and `DetectFeatureContext()` function
- Path resolution functions using `filepath.Join()`
- Branch name generation with stop-word filtering and 244-byte truncation

**Independent Test**: Go packages compile and basic functions work in isolation

**View tasks**:
```bash
sl issue show SL-4f80fb --tree
```

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: US1 - sl spec info Command (Priority: P1) 🎯 MVP

**Issue**: SL-043360 | **Priority**: 1 | **Status**: Open

**Goal**: Replace `check-prerequisites.sh` with `sl spec info` command that outputs JSON with feature paths and validation

**User Story**: As an AI agent, I need `sl spec info` to get feature paths and prerequisite validation so I can verify feature state before proceeding.

**Independent Test**: Run `sl spec info --json` and verify JSON output with FEATURE_DIR and AVAILABLE_DOCS fields

**Tasks**:

| ID | Title | Status | Component | Requirements | Blocked By |
|----|-------|--------|-----------|--------------|------------|
| SL-b60bb1 | Create pkg/cli/commands/spec.go parent command | Open | cli | FR-001 | Foundational |
| SL-958887 | Create pkg/cli/commands/spec_info.go implementation | Open | cli | FR-001, FR-002 | SL-b60bb1 |
| SL-5c02db | Register spec commands in cmd/sl/main.go | Open | cli | FR-001 | SL-958887 |

**Key Features**:
- JSON output with `--json` flag
- Validation flags: `--require-plan`, `--require-tasks`
- Doc discovery: `--include-tasks`
- Minimal output: `--paths-only`

**View tasks**:
```bash
sl issue show SL-043360 --tree
```

**Checkpoint**: sl spec info command fully functional and testable

---

## Phase 4: US2 - sl spec create Command (Priority: P1)

**Issue**: SL-d2f3bc | **Priority**: 1 | **Status**: Open

**Goal**: Replace `create-new-feature.sh` with `sl spec create` command that generates feature branches and spec directories

**User Story**: As a developer, I need `sl spec create` to generate a feature branch and spec directory so I can start a new feature without bash scripts.

**Independent Test**: Run `sl spec create --number 999 --short-name "test" --json` and verify branch and spec directory created

**Tasks**:

| ID | Title | Status | Component | Requirements | Blocked By |
|----|-------|--------|-----------|--------------|------------|
| SL-0c39e1 | Create pkg/cli/commands/spec_create.go implementation | Open | cli | FR-003, FR-004, FR-005 | Foundational (SL-03c3fe) |
| SL-74133d | Implement feature number collision detection | Open | cli | FR-005 | - |

**Key Features**:
- Branch name generation with stop-word filtering
- Acronym preservation (OAuth2, API, JWT)
- 244-byte branch name truncation
- Collision detection with existing features
- Template copying for spec.md

**View tasks**:
```bash
sl issue show SL-d2f3bc --tree
```

**Checkpoint**: sl spec create command fully functional, can create new features

---

## Phase 5: US3 - sl spec setup-plan Command (Priority: P2)

**Issue**: SL-61161a | **Priority**: 2 | **Status**: Open

**Goal**: Replace `setup-plan.sh` with `sl spec setup-plan` command that copies plan templates

**User Story**: As a developer, I need `sl spec setup-plan` to copy plan templates so I can begin planning without bash scripts.

**Independent Test**: Run `sl spec setup-plan --json` and verify plan.md created with template content

**Tasks**:

| ID | Title | Status | Component | Requirements | Blocked By |
|----|-------|--------|-----------|--------------|------------|
| SL-42ac0c | Create pkg/cli/commands/spec_setup_plan.go implementation | Open | cli | FR-006 | Foundational |

**Key Features**:
- Copy plan template from embedded files
- Error if plan.md already exists
- JSON output with PLAN_FILE path

**View tasks**:
```bash
sl issue show SL-61161a --tree
```

**Checkpoint**: sl spec setup-plan command functional

---

## Phase 6: US4 - sl context update Command (Priority: P1)

**Issue**: SL-1ab18f | **Priority**: 1 | **Status**: Open

**Goal**: Replace `update-agent-context.sh` with `sl context update` command that updates agent context files with plan metadata

**User Story**: As a developer, I need `sl context update` to parse plan.md and update agent files so AI assistants have current feature context.

**Independent Test**: Run `sl context update claude` and verify CLAUDE.md updated with plan metadata

**Tasks**:

| ID | Title | Status | Component | Requirements | Blocked By |
|----|-------|--------|-----------|--------------|------------|
| SL-0394dd | Create pkg/cli/context/parser.go | Open | cli | FR-007 | Foundational |
| SL-183067 | Create pkg/cli/context/updater.go | Open | cli | FR-008, FR-009 | SL-0394dd |
| SL-1052e6 | Create pkg/cli/commands/context.go parent command | Open | cli | FR-007 | - |
| SL-6f7f0c | Create pkg/cli/commands/context_update.go implementation | Open | cli | FR-007, FR-008, FR-009 | SL-183067, SL-1052e6 |

**Key Features**:
- Parse Technical Context from plan.md
- Update 17+ agent file types (CLAUDE.md, GEMINI.md, etc.)
- Preserve manual additions between markers
- Deduplicate entries (not append)
- Marker format: `<!-- MANUAL ADDITIONS START -->` / `<!-- MANUAL ADDITIONS END -->`

**View tasks**:
```bash
sl issue show SL-1ab18f --tree
```

**Checkpoint**: sl context update command fully functional, all agent files can be updated

---

## Phase 7: US5 - Cross-Platform Support (Priority: P2)

**Issue**: SL-92fb0b | **Priority**: 2 | **Status**: Open

**Goal**: Verify all commands work on macOS, Linux, Windows without bash/jq/sed/grep dependencies

**User Story**: As a Windows developer, I need all `sl` commands to work without bash so I can use SpecLedger on any platform.

**Independent Test**: Run all 4 new commands on Windows and verify no bash/jq/sed/grep errors

**Tasks**:

| ID | Title | Status | Component | Requirements | Blocked By |
|----|-------|--------|-----------|--------------|------------|
| SL-1d09c9 | Verify all path operations use filepath package | Open | cli | FR-010 | All US phases |
| SL-da1e8a | Test commands on Windows platform | Open | cli | FR-010 | All US phases |
| SL-c2f9a0 | Verify JSON output identical across platforms | Open | cli | FR-010 | All US phases |

**Key Features**:
- All paths use `filepath.Join()`
- No shell invocation in code paths
- Identical JSON output on all platforms
- Use go-git/v5 for git operations (already cross-platform)

**View tasks**:
```bash
sl issue show SL-92fb0b --tree
```

**Checkpoint**: All commands work on Windows without external dependencies

---

## Phase 8: Polish & Cross-Cutting Concerns

**Issue**: SL-9a0e47 | **Priority**: 3 | **Status**: Open

**Purpose**: Documentation, examples, integration testing, and bash script retention

**Tasks**:

| ID | Title | Status | Component | Blocked By |
|----|-------|--------|-----------|------------|
| SL-f69213 | Update documentation with new CLI commands | Open | docs | All US phases |
| SL-db4983 | Add usage examples for each command | Open | docs | All US phases |
| SL-303900 | Verify all commands work together | Open | cli | All US phases |
| SL-c46fee | Retain bash scripts as fallback | Open | infra | All US phases |

**Key Deliverables**:
- README.md updated with all 4 commands
- Usage examples for common workflows
- End-to-end integration test
- Bash scripts retained with deprecation notices

**View tasks**:
```bash
sl issue show SL-9a0e47 --tree
```

---

## Dependencies & Execution Order

### Phase Dependencies

```
Setup (SL-e12f56)
  └─→ Foundational (SL-4f80fb) [CRITICAL - blocks all US]
       ├─→ US1 (SL-043360) [P1] ──┐
       ├─→ US2 (SL-d2f3bc) [P1] ──┤
       ├─→ US3 (SL-61161a) [P2] ──┼─→ Polish (SL-9a0e47)
       ├─→ US4 (SL-1ab18f) [P1] ──┤
       └─→ US5 (SL-92fb0b) [P2] ──┘
```

### Parallel Execution Opportunities

**After Foundational Phase Completes**:
- ✅ US1 (sl spec info) - Can start immediately
- ✅ US2 (sl spec create) - Can start immediately (parallel to US1)
- ✅ US4 (sl context update) - Can start immediately (parallel to US1/US2)
- ⏸️ US3 (sl spec setup-plan) - Lower priority (P2), can wait
- ⏸️ US5 (Cross-Platform) - Lower priority (P2), can wait

**Recommended Execution Order** (if single developer):
1. Setup → Foundational (MUST complete first)
2. US1 (sl spec info) - Most commonly used by other commands
3. US2 (sl spec create) + US4 (sl context update) - Can parallelize if team available
4. US3 (sl spec setup-plan) - Simple, quick win
5. US5 (Cross-Platform) - Verification and testing
6. Polish - Documentation and cleanup

### Within Each User Story

- Foundational packages → Command implementation → Registration
- Models/packages before command logic
- Command logic before registration
- Each story independently testable upon completion

---

## Definition of Done Summary

| Issue ID | DoD Items |
|----------|-----------|
| SL-41daff | • Directory exists<br>• Proper package structure |
| SL-aa3426 | • Directory exists<br>• Proper package structure |
| SL-8a8228 | • Cobra in go.mod<br>• go-git/v5 in go.mod<br>• Dependencies installed |
| SL-aad91f | • plan-template.md exists<br>• spec-template.md exists<br>• Templates accessible via embedded FS |
| SL-e8c313 | • FeatureContext struct defined<br>• DetectFeatureContext() implemented<br>• Uses go-git/v5 for git ops<br>• Handles detached HEAD<br>• Returns error on non-feature branch |
| SL-6d857e | • GetFeatureDir() implemented<br>• GetSpecFile() implemented<br>• GetPlanFile() implemented<br>• GetTasksFile() implemented<br>• DiscoverDocs() implemented<br>• Uses filepath.Join() everywhere |
| SL-03c3fe | • StopWords map populated<br>• GenerateBranchName() implemented<br>• FilterStopWords() implemented<br>• PreserveAcronyms() implemented<br>• TruncateToLimit() implemented<br>• 244-byte limit enforced |
| SL-b60bb1 | • NewSpecCmd() function created<br>• Command structure follows Cobra patterns |
| SL-958887 | • NewSpecInfoCmd() implemented<br>• --json flag works<br>• --require-plan flag works<br>• --require-tasks flag works<br>• --include-tasks flag works<br>• --paths-only flag works<br>• Uses pkg/cli/spec packages |
| SL-5c02db | • Commands imported<br>• Commands added to root<br>• sl spec info works |
| SL-0c39e1 | • NewSpecCreateCmd() implemented<br>• --json flag works<br>• --number flag works<br>• --short-name flag works<br>• Branch created with go-git/v5<br>• Spec directory created<br>• Template copied<br>• Collision detection implemented |
| SL-74133d | • CheckFeatureCollision() implemented<br>• Checks local features<br>• Checks local branches<br>• Checks remote branches<br>• Returns error on collision |
| SL-42ac0c | • NewSpecSetupPlanCmd() implemented<br>• --json flag works<br>• Template copied correctly<br>• Error on existing plan.md<br>• Uses embedded templates |
| SL-0394dd | • TechnicalContext struct defined<br>• ParseTechnicalContext() implemented<br>• Parses Language/Version<br>• Parses Primary Dependencies<br>• Parses Storage<br>• Parses Project Type<br>• Handles malformed fields |
| SL-183067 | • AgentUpdater struct defined<br>• Update() implemented<br>• PreserveManualAdditions() implemented<br>• DeduplicateEntries() implemented<br>• Markers preserved<br>• Deduplication works<br>• Atomic write |
| SL-1052e6 | • NewContextCmd() function created<br>• Command structure follows Cobra patterns |
| SL-6f7f0c | • NewContextUpdateCmd() implemented<br>• --json flag works<br>• --agent flag works<br>• 17+ agent types supported<br>• Uses pkg/cli/context packages<br>• JSON output correct |
| SL-1d09c9 | • All paths use filepath.Join()<br>• No hardcoded separators<br>• Code review passed |
| SL-da1e8a | • sl spec info works on Windows<br>• sl spec create works on Windows<br>• sl spec setup-plan works on Windows<br>• sl context update works on Windows<br>• No external dependencies<br>• Path separators correct |
| SL-c2f9a0 | • JSON output captured on all platforms<br>• JSON format compared<br>• JSON is valid<br>• No unexpected differences |
| SL-f69213 | • README.md updated<br>• All 4 commands documented<br>• Flags documented<br>• JSON examples included |
| SL-db4983 | • Examples for sl spec info<br>• Examples for sl spec create<br>• Examples for sl spec setup-plan<br>• Examples for sl context update |
| SL-303900 | • Full workflow tested<br>• All commands work together<br>• No errors |
| SL-c46fee | • Scripts not deleted<br>• Deprecation notices added<br>• Scripts functional<br>• Documentation updated |

---

## Success Criteria Alignment

| Success Criteria | Related Issues | Verification |
|------------------|----------------|--------------|
| SC-001: 4 new CLI commands available | SL-043360, SL-d2f3bc, SL-61161a, SL-1ab18f | All 4 phases complete |
| SC-002: 0 bash script dependencies | SL-4f80fb, SL-92fb0b | No jq/grep/sed/bash in code paths |
| SC-003: Identical JSON output on all platforms | SL-92fb0b | Cross-platform tests pass |
| SC-004: Branch name handling correct | SL-03c3fe, SL-0c39e1 | Stop-words, acronyms, truncation tests |
| SC-005: Context update preserves and deduplicates | SL-183067, SL-6f7f0c | Marker preservation, deduplication tests |
| SC-006: AI commands updated | Deferred to 599-alignment | Bash scripts retained as fallback |

---

## Implementation Strategy

### MVP Scope (Minimum Viable Product)

**Phase 1 MVP**: Setup + Foundational + US1 (sl spec info)
- Delivers: Basic feature context detection and JSON output
- Enables: AI agents can query feature state
- Effort: ~2-3 days

**Phase 2 MVP**: Add US2 (sl spec create) + US4 (sl context update)
- Delivers: Feature creation and agent context updates
- Enables: Full feature workflow without bash scripts
- Effort: ~2-3 days additional

**Full Feature**: All 5 user stories + Polish
- Delivers: Complete bash script replacement
- Enables: Cross-platform support, documentation
- Effort: ~1 week total

### Incremental Delivery

Each user story is independently testable and deliverable:

1. **US1 (sl spec info)**: Can be released immediately - useful for AI agents
2. **US2 (sl spec create)**: Can be released next - useful for developers
3. **US4 (sl context update)**: Can be released next - useful for AI workflows
4. **US3 (sl spec setup-plan)**: Can be released when ready - simplifies planning
5. **US5 (Cross-Platform)**: Can be verified incrementally during development

---

## Notes for Implementation

### Parallel Work Assumptions

When working on tasks in parallel, assume:
- **Foundational packages are complete**: detector.go, paths.go, branch.go are available
- **Each US operates independently**: No shared state between commands
- **File paths are absolute**: All commands use resolved absolute paths
- **Error handling is consistent**: Follow existing patterns in pkg/cli/

### Testing Strategy

**No tests explicitly requested** in spec.md, so focus on:
- Manual testing per acceptance criteria
- Cross-platform verification (US5)
- End-to-end workflow testing (Polish phase)

If tests are desired, add test tasks to each US phase before implementation.

### Bash Script Retention

Keep bash scripts functional during transition:
- Add deprecation notices pointing to new Go commands
- Plan removal in 599-alignment when AI commands updated
- Scripts provide fallback if issues arise

---

## Quick Reference

### All Tasks by Status

```bash
# Open tasks
sl issue list --label spec:600-bash-cli-migration --status open

# In progress
sl issue list --label spec:600-bash-cli-migration --status in_progress

# Completed
sl issue list --label spec:600-bash-cli-migration --status closed
```

### Tasks by Component

```bash
# CLI implementation
sl issue list --label component:cli --label spec:600-bash-cli-migration

# Infrastructure
sl issue list --label component:infra --label spec:600-bash-cli-migration

# Documentation
sl issue list --label component:docs --label spec:600-bash-cli-migration
```

### Tasks by User Story

```bash
# US1 tasks
sl issue list --label story:US1 --label spec:600-bash-cli-migration

# US2 tasks
sl issue list --label story:US2 --label spec:600-bash-cli-migration

# US4 tasks
sl issue list --label story:US4 --label spec:600-bash-cli-migration
```

---

> This file is an index. Implementation data lives in the issue tracking system. Update this file only to provide context and query examples.
