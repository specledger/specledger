# Implementation Plan: Issue Create Fields Enhancement

**Branch**: `597-issue-create-fields` | **Date**: 2026-02-24 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `specledger/597-issue-create-fields/spec.md`

## Summary

Enhance `sl issue` CLI commands to expose all JSONL-supported fields (acceptance_criteria, definition_of_done, design, notes, parentId) via create/update flags, update specledger.tasks and specledger.implement prompts to utilize these structured fields, and improve task blocking tree relations for proper parent-child hierarchies.

## Technical Context

**Language/Version**: Go 1.24.2
**Primary Dependencies**: Cobra (CLI), go-git v5, YAML v3, Supabase (GoTrue, PostgREST)
**Storage**: File-based JSONL for issues
**Testing**: Go standard testing, make test
**Target Platform**: macOS/Linux CLI
**Project Type**: Single project (CLI tool)
**Performance Goals**: Standard CLI response times (<100ms for local operations)
**Constraints**: Local file system, offline-capable
**Scale/Scope**: Single-user CLI tool, hundreds of issues per project

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] **Specification-First**: Spec.md complete with 6 prioritized user stories
- [x] **Test-First**: Manual test scenarios defined in quickstart.md
- [x] **Code Quality**: Linting via golangci-lint, formatting via gofmt
- [x] **UX Consistency**: User flows documented in spec.md acceptance scenarios
- [x] **Performance**: CLI response time targets met (<100ms)
- [x] **Observability**: Standard CLI error output
- [x] **Issue Tracking**: Epic SL-bb6f5b created and linked

**Complexity Violations**: None identified

## Project Structure

### Documentation (this feature)

```text
specledger/597-issue-create-fields/
├── plan.md              # This file
├── research.md          # Prior work + Beads analysis
├── research/            # Detailed research files
│   └── beads-dependency-implementation.md
├── data-model.md        # Issue model field mappings
├── quickstart.md        # Test scenarios
├── contracts/
│   └── cli-contract.md  # CLI command specifications
└── tasks.md             # Task index (via /specledger.tasks)
```

### Source Code (repository root)

```text
cmd/sl/main.go                 # CLI entrypoint
pkg/cli/commands/
├── issue.go                   # issue create/update/show commands
├── tasks.go                   # /specledger.tasks prompt
└── implement.go               # /specledger.implement prompt
pkg/issues/
├── issue.go                   # Issue model (add ParentID field)
├── store.go                   # JSONL storage
└── validation.go              # Cycle detection for parent-child
pkg/embedded/
├── tasks.md                   # Embedded tasks prompt template
└── implement.md               # Embedded implement prompt template
.claude/commands/
├── specledger.tasks.md        # Tasks prompt (repo-local)
└── specledger.implement.md    # Implement prompt (repo-local)
```

**Structure Decision**: Single project with CLI commands in `pkg/cli/commands/`, models in `pkg/issues/`, and embedded templates in `pkg/embedded/`.

## Implementation Phases

### Phase 1: US1 - Create Issues with Complete Field Set (P1)

**Goal**: Add --acceptance-criteria, --dod, --design, --notes flags to `sl issue create`

**Files**:
- `pkg/cli/commands/issue.go` - Add flags to issueCreateCmd
- `pkg/issues/issue.go` - No changes (fields already exist)

**Tasks**:
1. Add --acceptance-criteria string flag
2. Add --dod StringArray flag (repeatable)
3. Add --design string flag
4. Add --notes string flag
5. Populate issue fields from flags in runIssueCreate
6. Update issue show display for new fields

### Phase 2: US2 - Update Definition of Done (P2)

**Goal**: Add --dod, --check-dod, --uncheck-dod flags to `sl issue update`

**Depends on**: Phase 1

**Files**:
- `pkg/cli/commands/issue.go` - Add flags to issueUpdateCmd
- `pkg/issues/issue.go` - No changes (methods already exist)

**Tasks**:
1. Add --dod StringArray flag for replacement
2. Add --check-dod string flag
3. Add --uncheck-dod string flag
4. Handle DoD operations in runIssueUpdate

### Phase 3: US6 - Parent-Child Relationships (P2)

**Goal**: Add --parent flag with single parent constraint and cycle detection

**Depends on**: Phase 1

**Files**:
- `pkg/cli/commands/issue.go` - Add --parent flag
- `pkg/issues/issue.go` - Add ParentID field, SetParent method
- `pkg/issues/validation.go` - Add cycle detection

**Tasks**:
1. Add ParentID *string field to Issue struct
2. Add --parent flag to issueCreateCmd and issueUpdateCmd
3. Implement single parent constraint validation
4. Implement cycle detection for parent-child relationships
5. Update issue show to display parent
6. Update issue show --tree to display children

### Phase 4: US3 - Tasks Generated with Proper Blocking Relations (P2)

**Goal**: Ensure generated tasks have correct blocking relationships

**Depends on**: Phase 1, Phase 3

**Files**:
- `pkg/cli/commands/tasks.go` - Update prompt
- `.claude/commands/specledger.tasks.md` - Update prompt
- `pkg/embedded/tasks.md` - Update embedded template

**Tasks**:
1. Review existing blocking logic in task generation
2. Add explicit blocking rules to prompt
3. Ensure parallelizable tasks are not blocked

### Phase 5: US4 - Tasks Prompt Utilizes Structured Fields (P3)

**Goal**: Update tasks prompt to use --acceptance-criteria, --dod, --design, --parent flags

**Depends on**: Phase 1, Phase 2, Phase 3

**Files**:
- `.claude/commands/specledger.tasks.md`
- `pkg/embedded/tasks.md`

**Tasks**:
1. Update CLI examples to use new flags
2. Add design field population instruction
3. Add parent-child hierarchy instruction

### Phase 6: US5 - Implement Prompt Utilizes Structured Fields (P3)

**Goal**: Update implement prompt to read design, verify AC, check DoD items

**Depends on**: Phase 1, Phase 2

**Files**:
- `.claude/commands/specledger.implement.md`
- `pkg/embedded/implement.md`

**Tasks**:
1. Add field reading instructions
2. Add AC verification instruction
3. Add --check-dod instruction

### Phase 7: Polish - Tests (P3)

**Goal**: Add unit tests for CLI flag changes

**Depends on**: All implementation phases

**Files**:
- `pkg/cli/commands/issue_test.go`

**Tasks**:
1. Add tests for create flags
2. Add tests for update flags
3. Add tests for parent-child validation

## Dependency Graph

```
Phase 1 (US1 - Create fields)
├── Phase 2 (US2 - Update DoD)
├── Phase 3 (US6 - Parent-Child)
└── Phase 4 (US3 - Blocking relations) [depends on 1, 3]├── Phase 5 (US4 - Tasks prompt) [depends on 1, 2, 3]
└── Phase 6 (US5 - Implement prompt) [depends on 1, 2]
    └── Phase 7 (Polish - Tests) [depends on all]
```

## Parallel Execution Opportunities

- Phase 2 and Phase 3 can run in parallel (both depend only on Phase 1)
- Phase 5 and Phase 6 can run in parallel after Phase 2 and Phase 3

## MVP Scope (Recommended)

Implement Phase 1 (US1) + Phase 3 (US6) for minimum viable functionality:
- Create issues with all 5 new flags (AC, DoD, Design, Notes, Parent)
- Display fields in issue show
- Parent-child relationships with validation

This enables immediate value: users can create fully-specified issues with hierarchies in one command.
