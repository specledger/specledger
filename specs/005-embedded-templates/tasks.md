# Tasks: Embedded Templates

**Input**: Design documents from `/specs/005-embedded-templates/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Tests are NOT explicitly requested in the feature specification. Test tasks are optional and can be added if needed.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- Single project structure at repository root: `cmd/`, `pkg/`, `templates/`, `tests/`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [ ] T001 Create templates manifest file at templates/manifest.yaml with Spec Kit template metadata
- [ ] T002 Create pkg/embedded directory for embedded filesystem package

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core template infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [ ] T003 Create embedded filesystem in pkg/embedded/templates.go with //go:embed directives for templates/ folder
- [ ] T004 Create template source interface in pkg/cli/templates/source.go defining TemplateSource, List, Copy, Exists methods
- [ ] T005 Create template data structures in pkg/cli/templates/template.go with Template, TemplateManifest, CopyOptions, CopyResult, CopyError types
- [ ] T006 Create manifest parser in pkg/cli/templates/manifest.go with LoadManifest, ParseManifest functions
- [ ] T007 Create file copying utilities in pkg/cli/templates/copy.go with CopyTemplates, CopyFile, CopyDir functions
- [ ] T008 Create embedded source implementation in pkg/cli/templates/embedded.go with EmbeddedSource struct implementing TemplateSource interface
- [ ] T009 Create template manager in pkg/cli/templates/templates.go with ApplyToProject, ListTemplates functions
- [ ] T010 Create integration test file at tests/integration/templates_test.go with test framework setup

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Create Project with Embedded Templates (Priority: P1) 🎯 MVP

**Goal**: Automatically copy embedded Spec Kit playbook templates to new projects during `sl new` and `sl init`

**Independent Test**: Run `sl new --framework speckit` and verify that:
1. The `.claude/` folder contains SpecLedger commands and skills
2. The `specledger/scripts/` folder contains helper scripts
3. The `specledger/templates/` folder contains file templates
4. The `.beads/` folder is initialized with configuration
5. All template files match the embedded versions

### Implementation for User Story 1

- [ ] T011 [US1] Add applyEmbeddedTemplates function to pkg/cli/commands/bootstrap_helpers.go that calls templates.ApplyToProject with framework
- [ ] T012 [US1] Integrate template copying into runNew function in pkg/cli/commands/bootstrap.go after project directory creation
- [ ] T013 [US1] Integrate template copying into runInit function in pkg/cli/commands/bootstrap_init.go after metadata creation
- [ ] T014 [US1] Add framework-to-file-patterns mapping in pkg/cli/templates/embedded.go for speckit, openspec, both, none
- [ ] T015 [US1] Add template validation in pkg/cli/templates/embedded.go to check templates folder exists before copy
- [ ] T016 [US1] Add existing file handling in pkg/cli/templates/copy.go with skip logic and warning messages
- [ ] T017 [US1] Add copy result tracking in pkg/cli/templates/copy.go with FilesCopied, FilesSkipped, Errors counts
- [ ] T018 [US1] Add verbose output option in pkg/cli/templates/copy.go for debugging template copy operations
- [ ] T019 [US1] Add error handling in pkg/cli/templates/embedded.go for missing templates, permission errors, disk space
- [ ] T020 [US1] Write integration test in tests/integration/templates_test.go for template copying with speckit framework
- [ ] T021 [US1] Write integration test in tests/integration/templates_test.go for sl init with template copying
- [ ] T022 [US1] Write integration test in tests/integration/templates_test.go for existing file skip behavior

**Checkpoint**: At this point, User Story 1 should be fully functional - users can create projects with embedded Spec Kit templates automatically applied

---

## Phase 4: User Story 2 - List Available Template Playbooks (Priority: P2)

**Goal**: Provide `sl template list` command to show available embedded templates

**Independent Test**: Run `sl template list` and verify it displays "speckit" as an available embedded template with name, description, framework, and version

### Implementation for User Story 2

- [ ] T023 [P] [US2] Create template list command in pkg/cli/commands/templates.go with VarTemplateCmd and runListTemplates function
- [ ] T024 [US2] Add template list formatters in pkg/cli/templates/templates.go with FormatTable, FormatJSON functions
- [ ] T025 [US2] Add template list command to root command in cmd/main.go with proper registration
- [ ] T026 [US2] Add autocomplete support in pkg/cli/commands/templates.go for --framework flag completion
- [ ] T027 [US2] Write integration test in tests/integration/templates_test.go for template list command output
- [ ] T028 [US2] Write integration test in tests/integration/templates_test.go for template list JSON output

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently - users can create projects with templates and discover available templates

---

## Phase 5: User Story 3 - Future: Remote Template Support (Priority: P3)

**Goal**: Design architecture to support future remote template fetching without major refactoring

**Independent Test**: Architecture review - verify that TemplateSource interface can support a RemoteSource implementation

### Implementation for User Story 3

- [ ] T029 [P] [US3] Add RemoteSource struct stub in pkg/cli/templates/remote.go implementing TemplateSource interface
- [ ] T030 [US3] Add remote template source documentation in pkg/cli/templates/remote.go with future implementation notes
- [ ] T031 [US3] Add template cache directory setup in pkg/cli/templates/cache.go for future remote template storage
- [ ] T032 [US3] Write architecture validation in tests/integration/templates_test.go verifying interface supports both embedded and remote sources

**Checkpoint**: All user stories should now be independently functional - architecture is ready for future remote template support

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [ ] T033 [P] Update README.md with template copying documentation and examples
- [ ] T034 [P] Update FORK.md with embedded templates information
- [ ] T035 [P] Add template manifest validation to pkg/cli/templates/manifest.go with schema checks
- [ ] T036 [P] Add performance metrics to pkg/cli/templates/copy.go tracking copy duration
- [ ] T037 Run quickstart.md validation to verify all documented examples work

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-5)**: All depend on Foundational phase completion
  - User stories can then proceed in parallel (if staffed)
  - Or sequentially in priority order (US1 → US2 → US3)
- **Polish (Phase 6)**: Depends on desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P2)**: Can start after Foundational (Phase 2) - Independent of US1
- **User Story 3 (P3)**: Can start after Foundational (Phase 2) - Independent of US1 and US2

### Within Each User Story

- US1: Bootstrap integration (T011-T013) depends on core templates package (T003-T009)
- US2: Command creation (T023-T026) depends on templates package (T003-T009)
- US3: Architecture stubs (T029-T032) can be done in parallel with other stories

### Parallel Opportunities

- T003-T009: All foundational package files can be created in parallel (different files)
- T020-T022: US1 integration tests can run in parallel
- T023-T024: US2 command and formatter can be created in parallel
- T029-T031: US3 architecture stubs can be created in parallel
- Once Foundational phase completes, US1, US2, US3 can all be worked on in parallel by different team members

---

## Parallel Example: User Story 1

```bash
# Launch integration tests for User Story 1 together:
Task T020: "Write integration test for template copying with speckit framework"
Task T021: "Write integration test for sl init with template copying"
Task T022: "Write integration test for existing file skip behavior"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1
4. **STOP and VALIDATE**: Test User Story 1 independently with `sl new --framework speckit`
5. Verify all template files are copied correctly

### Incremental Delivery

1. Complete Setup + Foundational → Template infrastructure ready
2. Add User Story 1 → Test independently → Deploy/Demo (MVP - automatic template copying!)
3. Add User Story 2 → Test independently → Deploy/Demo (template discovery)
4. Add User Story 3 → Test independently → Deploy/Demo (architecture for remote templates)
5. Each story adds value without breaking previous stories

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: User Story 1 (template copying integration)
   - Developer B: User Story 2 (template list command)
   - Developer C: User Story 3 (remote template architecture)
3. Stories complete and integrate independently

---

## Notes

- [P] tasks = different files, no dependencies
- [US1], [US2], [US3] labels map task to specific user story for traceability
- Each user story should be independently completable and testable
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Templates are embedded via Go embed package - no external file dependencies at runtime
- Architecture designed to support future remote templates without refactoring
