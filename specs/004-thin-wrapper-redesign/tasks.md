# Tasks: SpecLedger Thin Wrapper Architecture

**Feature**: 004-thin-wrapper-redesign
**Branch**: `004-thin-wrapper-redesign`
**Epic**: `sl-2n9`
**Created**: 2026-02-05

## Overview

This task breakdown implements the SpecLedger thin wrapper redesign by organizing work into independently testable user story phases. Each phase delivers a complete, verifiable increment.

## Source Documents

- **Specification**: [spec.md](./spec.md) - User stories and requirements
- **Implementation Plan**: [plan.md](./plan.md) - Technical approach and architecture
- **Data Model**: [data-model.md](./data-model.md) - YAML schema and entities
- **Research**: [research.md](./research.md) - Technology decisions
- **Contracts**: [contracts/](./contracts/) - API schemas

## Epic

**ID**: `sl-2n9`
**Title**: SpecLedger Thin Wrapper Architecture
**Description**: Redesign SpecLedger as a thin wrapper CLI that orchestrates tool installation and project bootstrapping while delegating SDD workflows to user-chosen frameworks (Spec Kit or OpenSpec).

**Labels**: `spec:004-thin-wrapper-redesign`, `component:cli`

## Task Organization

Tasks are organized by user story to enable independent implementation and testing. Each phase is a complete, independently verifiable increment.

### Beads Query Commands

```bash
# List all tasks for this feature
bd list --label "spec:004-thin-wrapper-redesign" -n 50

# Show ready-to-work tasks
bd ready --label "spec:004-thin-wrapper-redesign" -n 10

# Filter by user story
bd list --label "story:US1" -n 10
bd list --label "story:US2" -n 10

# Filter by phase
bd list --label "phase:setup" -n 10
bd list --label "phase:foundational" -n 10

# Show epic and its children
bd show sl-2n9 --children
```

## Phase 1: Setup and Cleanup

**Feature ID**: `sl-nh7`
**Purpose**: Remove duplicate SDD commands and unused playbook code
**Dependencies**: None (can start immediately)
**Story Mapping**: Primarily US5 (Framework Transparency), also supports US1

**Independent Test**: Verify no duplicate commands remain and codebase compiles successfully.

### Tasks Created

| Task ID | Title | Story | Priority | Status |
|---------|-------|-------|----------|--------|
| sl-pt7 | Remove duplicate specledger.specify command | US5 | P1 | open |
| sl-bzx | Remove duplicate specledger.plan command | US5 | P1 | open |
| sl-t13 | Remove duplicate specledger.tasks command | US5 | P1 | open |
| sl-5ey | Remove remaining duplicate SDD commands | US5 | P1 | open |
| sl-5k8 | Remove playbook code from TUI | US1 | P1 | open |
| sl-odx | Remove playbook flag from bootstrap command | US1 | P1 | open |

**Parallel Execution**: All tasks except sl-odx can run in parallel (different files). sl-odx depends on sl-5k8 completing.

**Acceptance Criteria**:
- All specledger.{specify,plan,tasks,implement,analyze,clarify,checklist,constitution}.md files removed
- Only specledger.{deps,adopt,resume}.md remain in commands/
- No playbook code in TUI or bootstrap
- Codebase compiles successfully
- All tests pass

## Phase 2: Foundational - Metadata System

**Feature ID**: `sl-nzj`
**Purpose**: Create YAML metadata infrastructure (blocks all user stories)
**Dependencies**: Phase 1 complete
**Story Mapping**: Foundation for US1, US3, US4

**Independent Test**: Can create, read, write, and migrate YAML metadata files.

### Tasks to Create

Use these Beads commands to create remaining foundational tasks:

```bash
# T007 - Already created as sl-bwv
# T008 - Already created as sl-0n1

# T009 - Implement YAML read/write functions
bd create "Implement YAML read/write functions" \
  --description "Create Load() and Save() functions in pkg/cli/metadata/yaml.go to read/write specledger.yaml files. Handle file I/O errors gracefully." \
  --design "LoadMetadata(dir string) reads YAML, unmarshals to ProjectMetadata struct, validates. SaveMetadata(dir string, meta *ProjectMetadata) marshals and writes atomically with temp file + rename pattern." \
  --acceptance "Can load valid YAML. Returns errors for invalid YAML. Can save metadata. Atomic writes (no partial files on crash). Unit tests pass." \
  --type task \
  --deps "parent-child:sl-nzj,blocks:sl-0n1" \
  --labels "spec:004-thin-wrapper-redesign,phase:foundational,story:US3,component:metadata,requirement:FR-008" \
  --priority 1

# T010 - Implement .mod to YAML migration
bd create "Implement .mod to YAML migration logic" \
  --description "Create migration.go with functions to parse legacy .mod files and convert to YAML format. Follow migration rules from data-model.md." \
  --design "ParseModFile() reads .mod, extracts project name/short code/created time. ConvertToYAML() creates ProjectMetadata with framework.choice='none', sets timestamps. Preserves .mod file (don't delete)." \
  --acceptance "Can parse valid .mod files. Correctly extracts metadata. Sets framework.choice to 'none'. Preserves original .mod. Unit tests with sample .mod files pass." \
  --type task \
  --deps "parent-child:sl-nzj,blocks:sl-0n1" \
  --labels "spec:004-thin-wrapper-redesign,phase:foundational,story:US4,component:metadata,requirement:FR-017" \
  --priority 1

# T011 - Add unit tests for metadata package
bd create "Add unit tests for metadata package" \
  --description "Create comprehensive unit tests for schema validation, YAML I/O, and migration logic in pkg/cli/metadata/." \
  --design "Create metadata_test.go. Test: valid/invalid YAML parsing, schema validation, migration from .mod, edge cases (empty files, corrupted YAML, missing fields). Use testdata/ directory for fixtures." \
  --acceptance "Test coverage >80%. All edge cases covered. Tests pass. Test data includes valid/invalid examples." \
  --type task \
  --deps "parent-child:sl-nzj,blocks:sl-0n1" \
  --labels "spec:004-thin-wrapper-redesign,phase:foundational,component:metadata" \
  --priority 2
```

**Acceptance Criteria**:
- pkg/cli/metadata/ package compiles
- Can load/save YAML metadata
- Can migrate .mod to YAML
- Unit tests pass with >80% coverage
- Matches schema in data-model.md

## Phase 3: Foundational - Prerequisites Checker

**Feature ID**: To be created
**Purpose**: Tool detection and installation prompting (blocks US1, US2)
**Dependencies**: None (parallel to Phase 2)
**Story Mapping**: Foundation for US1, US2

**Independent Test**: Can detect installed/missing tools and prompt for installation.

### Tasks to Create

```bash
# Create feature
bd create "Phase 3: Foundational - Prerequisites Checker" \
  --description "Create tool detection and installation infrastructure. Checks for mise, bd, perles, and optional frameworks. Provides interactive prompts or auto-install based on mode." \
  --type feature \
  --deps "parent-child:sl-2n9" \
  --labels "spec:004-thin-wrapper-redesign,phase:foundational,component:prerequisites" \
  --priority 0

# T012 - Create prerequisites package
bd create "Create prerequisites package structure" \
  --description "Create pkg/cli/prerequisites/ directory with checker.go for tool detection logic." \
  --design "Create directory pkg/cli/prerequisites/. Create checker.go with package prerequisites declaration. Add Tool struct and detection functions." \
  --acceptance "Directory exists. checker.go compiles. Package declaration present." \
  --type task \
  --deps "parent-child:<FEATURE_ID>" \
  --labels "spec:004-thin-wrapper-redesign,phase:foundational,component:prerequisites" \
  --priority 1

# T013 - Implement tool detection
bd create "Implement tool detection functions" \
  --description "Create functions to detect installed tools (mise, bd, perles, specify, openspec) via PATH lookup and --version checks." \
  --design "isMiseInstalled() uses exec.LookPath. isCommandAvailable(cmd string) checks PATH and runs cmd --version. GetToolStatus(name string) returns ToolStatus struct with version info. Use os/exec package." \
  --acceptance "Can detect mise, bd, perles, specify, openspec. Returns version strings. Handles missing tools gracefully. Unit tests pass." \
  --type task \
  --deps "parent-child:<FEATURE_ID>,blocks:<T012>" \
  --labels "spec:004-thin-wrapper-redesign,phase:foundational,story:US2,component:prerequisites,requirement:FR-015" \
  --priority 1

# T014 - Implement prerequisite checking
bd create "Implement CheckPrerequisites function" \
  --description "Create CheckPrerequisites() that validates mise is installed and returns clear error if missing." \
  --design "CheckPrerequisites() checks for mise only (core requirement). Returns formatted error with installation instructions if missing. Does not check frameworks (optional)." \
  --acceptance "Returns nil if mise installed. Returns error with install instructions if mise missing. Does not check frameworks. Unit tests pass." \
  --type task \
  --deps "parent-child:<FEATURE_ID>,blocks:<T013>" \
  --labels "spec:004-thin-wrapper-redesign,phase:foundational,story:US1,component:prerequisites,requirement:FR-005" \
  --priority 1

# T015 - Implement EnsurePrerequisites with prompts
bd create "Implement EnsurePrerequisites with interactive prompts" \
  --description "Create EnsurePrerequisites(interactive bool) that checks tools and prompts for installation in interactive mode, auto-installs in CI mode." \
  --design "In interactive mode: detect missing tools, prompt user 'Install via mise? [Y/n]', call mise install if yes. In CI mode: auto-call mise install. Use tui.ConfirmPrompt for interactive prompts." \
  --acceptance "Interactive mode prompts user. CI mode auto-installs. Handles mise install failures. Returns clear errors. Integration tests pass." \
  --type task \
  --deps "parent-child:<FEATURE_ID>,blocks:<T014>" \
  --labels "spec:004-thin-wrapper-redesign,phase:foundational,story:US1,component:prerequisites,requirement:FR-006,requirement:FR-007" \
  --priority 1

# T016 - Add unit tests for prerequisites
bd create "Add unit tests for prerequisites package" \
  --description "Create comprehensive unit tests for tool detection and prerequisite checking." \
  --design "Create checker_test.go. Test: tool detection with mock exec.LookPath, missing/present tools, version parsing, error messages. Mock file system for PATH checks." \
  --acceptance "Test coverage >80%. All tool detection scenarios covered. Tests pass." \
  --type task \
  --deps "parent-child:<FEATURE_ID>,blocks:<T013>" \
  --labels "spec:004-thin-wrapper-redesign,phase:foundational,component:prerequisites" \
  --priority 2
```

**Acceptance Criteria**:
- pkg/cli/prerequisites/ package compiles
- Can detect all required and optional tools
- Interactive and CI modes work correctly
- Unit tests pass with >80% coverage

## Phase 4: User Story 2 - Check Tool Installation Status (Priority P1)

**Feature ID**: To be created
**Purpose**: Implement `sl doctor` command for diagnostics
**Dependencies**: Phase 3 complete
**Story**: US2

**Independent Test**: Run `sl doctor` and verify tool status is reported correctly.

### Tasks to Create

```bash
# Create feature
bd create "Phase 4: User Story 2 - sl doctor Command" \
  --description "Implement 'sl doctor' diagnostic command that checks and reports installation status of all tools. Delivers US2 independently." \
  --type feature \
  --deps "parent-child:sl-2n9,blocks:<PHASE3_FEATURE_ID>" \
  --labels "spec:004-thin-wrapper-redesign,phase:us2,story:US2,component:cli" \
  --priority 0

# T017 - Create doctor command
bd create "Create sl doctor command file" \
  --description "Create pkg/cli/commands/doctor.go with Cobra command for 'sl doctor'. Implement basic command structure." \
  --design "Create doctor.go. Define DoctorCmd with cobra.Command. Add Use, Short, Long descriptions. Register command in cli.go. Add --json flag for JSON output option." \
  --acceptance "Command compiles. sl doctor --help works. Command registered in root command. --json flag present." \
  --type task \
  --deps "parent-child:<FEATURE_ID>" \
  --labels "spec:004-thin-wrapper-redesign,phase:us2,story:US2,component:cli,requirement:FR-004" \
  --priority 1

# T018 - Implement tool status checking
bd create "Implement tool status checking in doctor command" \
  --description "Call prerequisites package functions to check all tools (core + frameworks) and collect status information." \
  --design "Import pkg/cli/prerequisites. Call detection functions for mise, bd, perles, specify, openspec. Collect ToolStatus structs. Determine overall pass/fail status (pass only if all core tools present)." \
  --acceptance "Checks all 5 tools. Correctly categorizes core vs framework. Pass/fail status accurate. Handles missing tools gracefully." \
  --type task \
  --deps "parent-child:<FEATURE_ID>,blocks:<T017>" \
  --labels "spec:004-thin-wrapper-redesign,phase:us2,story:US2,component:cli,requirement:FR-004" \
  --priority 1

# T019 - Add human-readable output
bd create "Add human-readable output to doctor command" \
  --description "Format tool status as human-readable output with checkmarks, versions, and installation instructions." \
  --design "Print 'SpecLedger Environment Check' header. For each tool: print ✅ name (version) if installed, ❌ name (not installed) if missing. Group by Core Tools and SDD Frameworks. If failures, print installation instructions at end." \
  --acceptance "Output is clear and easy to read. Uses checkmarks. Shows versions. Groups tools logically. Provides install instructions for missing tools. Matches design in plan.md." \
  --type task \
  --deps "parent-child:<FEATURE_ID>,blocks:<T018>" \
  --labels "spec:004-thin-wrapper-redesign,phase:us2,story:US2,component:cli" \
  --priority 1

# T020 - Add JSON output option
bd create "Add JSON output option to doctor command" \
  --description "When --json flag is used, output tool status as structured JSON matching contracts/doctor-output.json schema." \
  --design "If --json flag set: marshal tool statuses to JSON matching schema in contracts/doctor-output.json. Include status, tools array, missing array, install_instructions. Print to stdout. Exit with code 0 (pass) or 1 (fail)." \
  --acceptance "JSON output valid. Matches schema in contracts/doctor-output.json. Can be parsed by jq. Exit codes correct." \
  --type task \
  --deps "parent-child:<FEATURE_ID>,blocks:<T018>" \
  --labels "spec:004-thin-wrapper-redesign,phase:us2,story:US2,component:cli,requirement:FR-004" \
  --priority 1

# T021 - Add integration test for doctor command
bd create "Add integration test for sl doctor" \
  --description "Create integration test that runs sl doctor in test environment and verifies output." \
  --design "Create tests/integration/doctor_test.go. Test scenarios: all tools present, some missing, --json flag output. Mock tool installation for testing. Verify exit codes and output format." \
  --acceptance "Integration tests pass. Tests both human and JSON output. Covers success and failure cases." \
  --type task \
  --deps "parent-child:<FEATURE_ID>,blocks:<T019>,blocks:<T020>" \
  --labels "spec:004-thin-wrapper-redesign,phase:us2,story:US2,component:cli" \
  --priority 2
```

**US2 Acceptance**: `sl doctor` command works, shows tool status, provides JSON output, helps users troubleshoot.

**Checkpoint**: ✅ US2 complete and independently testable

## Phase 5: User Story 1 - Install and Bootstrap New Project (Priority P1)

**Feature ID**: To be created
**Purpose**: Update `sl new` to use prerequisites checker, add framework selection, use YAML metadata
**Dependencies**: Phase 2 (metadata), Phase 3 (prerequisites) complete
**Story**: US1

**Independent Test**: Run `sl new`, select framework, verify project created with YAML metadata and tools installed.

### Tasks to Create

```bash
# Create feature
bd create "Phase 5: User Story 1 - Enhanced Bootstrap" \
  --description "Update sl new command to check prerequisites, add framework selection, and use YAML metadata. Delivers US1 independently." \
  --type feature \
  --deps "parent-child:sl-2n9,blocks:<PHASE2_FEATURE_ID>,blocks:<PHASE3_FEATURE_ID>" \
  --labels "spec:004-thin-wrapper-redesign,phase:us1,story:US1,component:cli" \
  --priority 0

# T022 - Add framework selection to TUI
bd create "Add framework selection step to TUI" \
  --description "Add framework selection step to pkg/cli/tui/sl_new.go where playbook was removed. Let user choose Spec Kit, OpenSpec, Both, or None." \
  --design "Add stepFramework constant. Create viewFramework() method with framework options. Add getFrameworkOptions() returning ['Spec Kit', 'OpenSpec', 'Both', 'None']. Update step transitions. Store choice in answers['framework']." \
  --acceptance "TUI shows framework selection step. All 4 options available. Selection stored. TUI flows correctly. Code compiles." \
  --type task \
  --deps "parent-child:<FEATURE_ID>" \
  --labels "spec:004-thin-wrapper-redesign,phase:us1,story:US1,component:tui,requirement:FR-001" \
  --priority 1

# T023 - Integrate prerequisites check in bootstrap
bd create "Integrate prerequisite check in bootstrap command" \
  --description "Call prerequisites.EnsurePrerequisites() at start of bootstrap (both interactive and CI modes)." \
  --design "In pkg/cli/commands/bootstrap.go: Import pkg/cli/prerequisites. In VarBootstrapCmd.RunE, call prerequisites.EnsurePrerequisites(true) for interactive. In runBootstrapNonInteractive, call prerequisites.EnsurePrerequisites(false). Handle errors and abort if prerequisites fail." \
  --acceptance "Prerequisites checked before bootstrap. Interactive mode prompts user. CI mode auto-installs. Bootstrap aborts if prerequisites fail. Error messages clear." \
  --type task \
  --deps "parent-child:<FEATURE_ID>" \
  --labels "spec:004-thin-wrapper-redesign,phase:us1,story:US1,component:cli,requirement:FR-005,requirement:FR-006" \
  --priority 1

# T024 - Update bootstrap to use YAML metadata
bd create "Update bootstrap to write YAML metadata" \
  --description "Replace .mod file writing with YAML metadata using new metadata package." \
  --design "In bootstrap.go: Import pkg/cli/metadata. Create ProjectMetadata struct from user inputs (name, short code, framework choice). Set timestamps to time.Now(). Call metadata.SaveMetadata() instead of writing .mod file. Handle framework choice from TUI." \
  --acceptance "Bootstrap writes specledger/specledger.yaml. YAML is valid. Contains all required fields. Framework choice recorded. No .mod file written for new projects." \
  --type task \
  --deps "parent-child:<FEATURE_ID>,blocks:<T022>" \
  --labels "spec:004-thin-wrapper-redesign,phase:us1,story:US1,component:cli,requirement:FR-008,requirement:FR-013" \
  --priority 1

# T025 - Update mise.toml template with framework options
bd create "Update mise.toml template with commented framework options" \
  --description "Add commented-out framework options to pkg/embedded/templates/mise.toml with clear instructions." \
  --design "Add comments above core tools section. Add commented lines for Spec Kit (pipx) and OpenSpec (npm) with explanations. Add instructions: 'Uncomment the framework(s) you want, then run: mise install'. Follow pattern from research.md." \
  --acceptance "Template has clear comments. Framework options commented out. Instructions present. Valid TOML syntax." \
  --type task \
  --deps "parent-child:<FEATURE_ID>" \
  --labels "spec:004-thin-wrapper-redesign,phase:us1,story:US1,component:templates,requirement:FR-011,requirement:FR-016" \
  --priority 1

# T026 - Create YAML metadata template
bd create "Create specledger.yaml template" \
  --description "Create pkg/embedded/templates/specledger/specledger.yaml template file." \
  --design "Create template with placeholder values: version '1.0.0', project fields with {{.ProjectName}}/{{.ShortCode}}/{{.Created}} placeholders, framework.choice with {{.FrameworkChoice}} placeholder, empty dependencies array. Follow schema in contracts/specledger-schema.yaml." \
  --acceptance "Template file exists. Valid YAML syntax. Has all required fields. Placeholders present for dynamic values." \
  --type task \
  --deps "parent-child:<FEATURE_ID>" \
  --labels "spec:004-thin-wrapper-redesign,phase:us1,story:US1,component:templates,requirement:FR-008" \
  --priority 1

# T027 - Update CI mode to support new metadata
bd create "Update non-interactive bootstrap for YAML" \
  --description "Ensure runBootstrapNonInteractive() uses YAML metadata with default framework choice 'none'." \
  --design "In runBootstrapNonInteractive: Set framework choice to 'none' if not specified. Call metadata.SaveMetadata() with appropriate values. Ensure prerequisite check happens. Test with --ci flag." \
  --acceptance "sl new --ci creates YAML metadata. Framework choice defaults to 'none'. Prerequisites checked in CI mode. Integration test passes." \
  --type task \
  --deps "parent-child:<FEATURE_ID>,blocks:<T024>" \
  --labels "spec:004-thin-wrapper-redesign,phase:us1,story:US1,component:cli,requirement:FR-007" \
  --priority 1

# T028 - Add integration test for bootstrap
bd create "Add integration test for sl new workflow" \
  --description "Create comprehensive integration test for bootstrap workflow (interactive and CI modes)." \
  --design "Create tests/integration/bootstrap_test.go. Test scenarios: interactive mode with framework selection, CI mode, prerequisite prompts, YAML metadata creation, framework choices (all 4 options). Use temp directories. Mock user input." \
  --acceptance "Integration tests pass. Covers interactive and CI modes. Tests all framework choices. Verifies YAML metadata created correctly." \
  --type task \
  --deps "parent-child:<FEATURE_ID>,blocks:<T024>,blocks:<T027>" \
  --labels "spec:004-thin-wrapper-redesign,phase:us1,story:US1,component:cli" \
  --priority 2
```

**US1 Acceptance**: `sl new` command checks prerequisites, lets users choose framework, creates YAML metadata, works in CI mode.

**Checkpoint**: ✅ US1 complete and independently testable

## Phase 6: User Story 4 - Initialize in Existing Project (Priority P2)

**Feature ID**: To be created
**Purpose**: Update `sl init` command for YAML and prerequisites
**Dependencies**: Phase 2, Phase 3 complete
**Story**: US4

**Independent Test**: Run `sl init` in existing directory, verify SpecLedger files added without disrupting existing code.

### Tasks to Create

```bash
# Create feature
bd create "Phase 6: User Story 4 - sl init for Existing Projects" \
  --description "Update sl init command to use YAML metadata and check prerequisites. Delivers US4 independently." \
  --type feature \
  --deps "parent-child:sl-2n9,blocks:<PHASE2_FEATURE_ID>,blocks:<PHASE3_FEATURE_ID>" \
  --labels "spec:004-thin-wrapper-redesign,phase:us4,story:US4,component:cli" \
  --priority 1

# T029 - Integrate prerequisites in init command
bd create "Add prerequisite check to sl init" \
  --description "Call prerequisites.EnsurePrerequisites() in VarInitCmd.RunE before initializing project." \
  --design "In pkg/cli/commands/bootstrap.go VarInitCmd.RunE: Call prerequisites.EnsurePrerequisites(true) at start. Handle errors and abort if prerequisites fail. Use same error messages as sl new." \
  --acceptance "Prerequisites checked before init. Prompts user for missing tools. Aborts gracefully if prerequisites fail." \
  --type task \
  --deps "parent-child:<FEATURE_ID>" \
  --labels "spec:004-thin-wrapper-redesign,phase:us4,story:US4,component:cli,requirement:FR-002" \
  --priority 1

# T030 - Update init to use YAML metadata
bd create "Update sl init to write YAML metadata" \
  --description "Replace .mod writing in runInit() with YAML metadata creation." \
  --design "In runInit(): Prompt for framework choice if not provided via flag. Create ProjectMetadata with framework.choice='none' as default. Call metadata.SaveMetadata(). Derive project name from directory name. Prompt for short code." \
  --acceptance "sl init writes specledger/specledger.yaml. Prompts for short code. Framework choice defaults to 'none'. YAML valid. No .mod written." \
  --type task \
  --deps "parent-child:<FEATURE_ID>,blocks:<T029>" \
  --labels "spec:004-thin-wrapper-redesign,phase:us4,story:US4,component:cli,requirement:FR-002,requirement:FR-008" \
  --priority 1

# T031 - Handle existing YAML with --force
bd create "Add --force flag handling for existing YAML" \
  --description "When specledger/specledger.yaml exists, prompt or use --force flag to overwrite/merge." \
  --design "Check if specledger/specledger.yaml exists. If yes and no --force: prompt 'SpecLedger already initialized. Overwrite? [y/N]'. If --force or user confirms: load existing, merge with new values, save. Preserve existing dependencies." \
  --acceptance "Detects existing YAML. Prompts user appropriately. --force flag skips prompt. Existing data preserved/merged. Error handling for corrupted YAML." \
  --type task \
  --deps "parent-child:<FEATURE_ID>,blocks:<T030>" \
  --labels "spec:004-thin-wrapper-redesign,phase:us4,story:US4,component:cli,requirement:FR-002" \
  --priority 1

# T032 - Add integration test for init command
bd create "Add integration test for sl init" \
  --description "Create integration test that runs sl init in existing directory and verifies behavior." \
  --design "Create tests/integration/init_test.go. Test scenarios: init in empty directory, init in existing git repo, init with existing YAML, --force flag behavior. Verify no existing files modified (except .beads/, specledger/)." \
  --acceptance "Integration tests pass. Tests all init scenarios. Verifies existing files not modified. YAML metadata created correctly." \
  --type task \
  --deps "parent-child:<FEATURE_ID>,blocks:<T031>" \
  --labels "spec:004-thin-wrapper-redesign,phase:us4,story:US4,component:cli" \
  --priority 2
```

**US4 Acceptance**: `sl init` command works in existing projects, creates YAML metadata, checks prerequisites, handles existing config.

**Checkpoint**: ✅ US4 complete and independently testable

## Phase 7: User Story 3 - Manage Spec Dependencies (Priority P2)

**Feature ID**: To be created
**Purpose**: Update `sl deps` command to use YAML format
**Dependencies**: Phase 2 complete
**Story**: US3

**Independent Test**: Run `sl deps add/list/resolve` and verify dependencies stored in YAML.

### Tasks to Create

```bash
# Create feature
bd create "Phase 7: User Story 3 - YAML-based Dependencies" \
  --description "Update sl deps commands to read/write YAML metadata instead of .mod format. Delivers US3 independently." \
  --type feature \
  --deps "parent-child:sl-2n9,blocks:<PHASE2_FEATURE_ID>" \
  --labels "spec:004-thin-wrapper-redesign,phase:us3,story:US3,component:cli" \
  --priority 1

# T033 - Update deps add to use YAML
bd create "Update sl deps add for YAML format" \
  --description "Modify pkg/cli/commands/deps.go add command to load YAML, append dependency, save YAML." \
  --design "In deps add handler: Call metadata.LoadMetadata() to read current YAML. Create Dependency struct from args. Append to Dependencies array. Call metadata.SaveMetadata(). Handle errors gracefully." \
  --acceptance "sl deps add records dependency in YAML. YAML remains valid. Duplicate URLs detected. Error handling works." \
  --type task \
  --deps "parent-child:<FEATURE_ID>" \
  --labels "spec:004-thin-wrapper-redesign,phase:us3,story:US3,component:cli,requirement:FR-003,requirement:FR-014" \
  --priority 1

# T034 - Update deps list to use YAML
bd create "Update sl deps list for YAML format" \
  --description "Modify deps list command to read from YAML metadata and display dependencies." \
  --design "In deps list handler: Call metadata.LoadMetadata(). Iterate over Dependencies array. Print formatted list with URL, branch, alias, resolved commit. Handle empty dependencies gracefully." \
  --acceptance "sl deps list shows dependencies from YAML. Format matches existing output. Works with empty dependencies. Handles missing/corrupted YAML." \
  --type task \
  --deps "parent-child:<FEATURE_ID>" \
  --labels "spec:004-thin-wrapper-redesign,phase:us3,story:US3,component:cli,requirement:FR-003" \
  --priority 1

# T035 - Update deps resolve to use YAML
bd create "Update sl deps resolve for YAML format" \
  --description "Modify deps resolve command to read dependencies from YAML, fetch them, update resolved_commit in YAML." \
  --design "In deps resolve handler: Load YAML. For each dependency: git clone to cache directory, record commit SHA. Update dependency.ResolvedCommit. Save updated YAML. Handle git errors gracefully." \
  --acceptance "sl deps resolve fetches dependencies. Updates resolved_commit in YAML. Handles git errors. Cache directory created at ~/.specledger/cache/. YAML updated correctly." \
  --type task \
  --deps "parent-child:<FEATURE_ID>,blocks:<T033>" \
  --labels "spec:004-thin-wrapper-redesign,phase:us3,story:US3,component:cli,requirement:FR-003,requirement:FR-014" \
  --priority 1

# T036 - Update deps remove to use YAML
bd create "Update sl deps remove for YAML format" \
  --description "Modify deps remove command to load YAML, remove dependency by URL or alias, save YAML." \
  --design "In deps remove handler: Load YAML. Find dependency by URL or alias. Remove from Dependencies array. Save updated YAML. Print confirmation message." \
  --acceptance "sl deps remove deletes dependency from YAML. Works with URL or alias. Handles non-existent dependencies gracefully. YAML remains valid." \
  --type task \
  --deps "parent-child:<FEATURE_ID>" \
  --labels "spec:004-thin-wrapper-redesign,phase:us3,story:US3,component:cli,requirement:FR-003" \
  --priority 1

# T037 - Add integration test for deps commands
bd create "Add integration test for sl deps workflow" \
  --description "Create integration test covering full dependency workflow: add, list, resolve, remove." \
  --design "Create tests/integration/deps_test.go. Test: add dependency, verify YAML updated, list dependencies, resolve (mock git), verify resolved_commit, remove dependency. Use temp project directory." \
  --acceptance "Integration tests pass. Tests full workflow. Verifies YAML correctness at each step. Handles errors appropriately." \
  --type task \
  --deps "parent-child:<FEATURE_ID>,blocks:<T033>,blocks:<T035>" \
  --labels "spec:004-thin-wrapper-redesign,phase:us3,story:US3,component:cli" \
  --priority 2
```

**US3 Acceptance**: `sl deps` commands work with YAML format, dependencies cached correctly, no regression in functionality.

**Checkpoint**: ✅ US3 complete and independently testable

## Phase 8: Migration Support

**Feature ID**: To be created
**Purpose**: Create `sl migrate` command for .mod to YAML migration
**Dependencies**: Phase 2 complete
**Story**: Supports US1, US3, US4 (backward compatibility)

**Independent Test**: Run `sl migrate` on project with .mod file, verify YAML created correctly.

### Tasks to Create

```bash
# Create feature
bd create "Phase 8: Migration Support" \
  --description "Create sl migrate command to convert existing .mod files to YAML format. Ensures backward compatibility." \
  --type feature \
  --deps "parent-child:sl-2n9,blocks:<PHASE2_FEATURE_ID>" \
  --labels "spec:004-thin-wrapper-redesign,phase:migration,component:cli" \
  --priority 1

# T038 - Create migrate command
bd create "Create sl migrate command file" \
  --description "Create pkg/cli/commands/migrate.go with Cobra command for 'sl migrate'." \
  --design "Create migrate.go. Define MigrateCmd with cobra.Command. Add Use, Short, Long descriptions. Register command in cli.go. Add --dry-run flag to preview without writing." \
  --acceptance "Command compiles. sl migrate --help works. Command registered. --dry-run flag present." \
  --type task \
  --deps "parent-child:<FEATURE_ID>" \
  --labels "spec:004-thin-wrapper-redesign,phase:migration,component:cli,requirement:FR-017" \
  --priority 1

# T039 - Implement migration logic in command
bd create "Implement migration logic in migrate command" \
  --description "Call migration.ConvertToYAML() from metadata package to perform migration." \
  --design "In migrate handler: Check for specledger/specledger.mod. If not found, error. If specledger.yaml exists, prompt to overwrite. Call migration.ParseModFile() and migration.ConvertToYAML(). If not --dry-run, call metadata.SaveMetadata(). Print success message with path to new YAML. Preserve original .mod file." \
  --acceptance "Migrates .mod to YAML. Preserves original .mod. Prompts before overwriting existing YAML. --dry-run works. Success message printed." \
  --type task \
  --deps "parent-child:<FEATURE_ID>,blocks:<T038>" \
  --labels "spec:004-thin-wrapper-redesign,phase:migration,component:cli,requirement:FR-017" \
  --priority 1

# T040 - Add automatic detection with warning
bd create "Add .mod detection with deprecation warning" \
  --description "Update metadata.LoadMetadata() to detect .mod files and print deprecation warning." \
  --design "In metadata/yaml.go LoadMetadata(): If YAML not found but .mod found, print warning: '⚠️ .mod format is deprecated. Run sl migrate to convert to YAML.' Then call migration.ParseModFile() to load data temporarily. Don't auto-migrate (user control)." \
  --acceptance "Detects .mod files. Prints clear deprecation warning. Still loads .mod data (read-only). Does not auto-migrate. Warning message helpful." \
  --type task \
  --deps "parent-child:<FEATURE_ID>" \
  --labels "spec:004-thin-wrapper-redesign,phase:migration,component:metadata,requirement:FR-017" \
  --priority 1

# T041 - Add integration test for migration
bd create "Add integration test for sl migrate" \
  --description "Create integration test that migrates sample .mod files and verifies YAML correctness." \
  --design "Create tests/integration/migrate_test.go. Create sample .mod files in testdata/. Test: migrate valid .mod, verify YAML fields, test --dry-run, test with existing YAML, test with missing .mod. Verify original .mod preserved." \
  --acceptance "Integration tests pass. Tests all migration scenarios. Verifies YAML correctness. Tests error handling." \
  --type task \
  --deps "parent-child:<FEATURE_ID>,blocks:<T039>" \
  --labels "spec:004-thin-wrapper-redesign,phase:migration,component:cli" \
  --priority 2
```

**Migration Acceptance**: `sl migrate` works, .mod files detected with warnings, backward compatibility maintained.

**Checkpoint**: ✅ Migration support complete

## Phase 9: Documentation and Polish

**Feature ID**: To be created
**Purpose**: Update documentation to reflect new architecture
**Dependencies**: All previous phases complete
**Story**: Supports all user stories (documentation)

**Independent Test**: Review documentation for accuracy and completeness.

### Tasks to Create

```bash
# Create feature
bd create "Phase 9: Documentation and Polish" \
  --description "Update README, create ARCHITECTURE.md, update help text, and finalize user-facing documentation." \
  --type feature \
  --deps "parent-child:sl-2n9,blocks:<ALL_PREVIOUS_FEATURES>" \
  --labels "spec:004-thin-wrapper-redesign,phase:polish,component:docs" \
  --priority 2

# T042 - Update README.md
bd create "Update README with new architecture" \
  --description "Update README.md to document thin wrapper architecture, framework choices, and new YAML format." \
  --design "Add Architecture section explaining orchestration role. Update Installation section with framework selection info. Add 'Choosing an SDD Framework' section comparing Spec Kit vs OpenSpec. Document YAML metadata format. Update Quick Start examples. Remove references to removed commands." \
  --acceptance "README accurate. Explains thin wrapper concept. Documents framework choices. Shows YAML examples. No references to removed commands. Clear migration guide." \
  --type task \
  --deps "parent-child:<FEATURE_ID>" \
  --labels "spec:004-thin-wrapper-redesign,phase:polish,component:docs" \
  --priority 2

# T043 - Create ARCHITECTURE.md
bd create "Create ARCHITECTURE.md design document" \
  --description "Create comprehensive architecture documentation explaining design decisions and component relationships." \
  --design "Create ARCHITECTURE.md following structure from plan.md: Design Philosophy, Component Diagram, Decision Records (why thin wrapper, why mise, why YAML), Extension Points. Document relationships between SpecLedger, frameworks, and tools." \
  --acceptance "ARCHITECTURE.md created. Explains all design decisions. Includes diagrams. Documents extension points. Clear and comprehensive." \
  --type task \
  --deps "parent-child:<FEATURE_ID>" \
  --labels "spec:004-thin-wrapper-redesign,phase:polish,component:docs" \
  --priority 2

# T044 - Update CLI help text
bd create "Update CLI help text for all commands" \
  --description "Review and update help text for sl new, sl init, sl doctor, sl migrate, sl deps to reflect new functionality." \
  --design "Update Short and Long descriptions in each command file. Ensure flags documented. Add examples to Long descriptions. Update root command description. Remove references to removed commands." \
  --acceptance "sl --help accurate. Each command help up-to-date. Examples provided. No references to removed commands. Help text clear and useful." \
  --type task \
  --deps "parent-child:<FEATURE_ID>" \
  --labels "spec:004-thin-wrapper-redesign,phase:polish,component:docs" \
  --priority 2

# T045 - Create migration guide
bd create "Create migration guide for existing users" \
  --description "Write migration guide documenting upgrade path from old SpecLedger to new thin wrapper version." \
  --design "Create MIGRATION.md. Document: what's changed, breaking changes, migration steps (run sl migrate), framework selection process, .mod deprecation timeline, FAQ. Provide troubleshooting tips." \
  --acceptance "MIGRATION.md created. Clear upgrade path. Documents breaking changes. Provides examples. FAQ answers common questions." \
  --type task \
  --deps "parent-child:<FEATURE_ID>" \
  --labels "spec:004-thin-wrapper-redesign,phase:polish,component:docs" \
  --priority 2

# T046 - Update CONTRIBUTING.md if needed
bd create "Review and update CONTRIBUTING.md" \
  --description "Update contributor guidelines if new patterns or requirements introduced." \
  --design "Review CONTRIBUTING.md. Update if: new testing requirements, new package structure patterns, new documentation standards. Ensure guidelines match new architecture." \
  --acceptance "CONTRIBUTING.md accurate. Reflects new structure. Testing guidelines up-to-date. No outdated information." \
  --type task \
  --deps "parent-child:<FEATURE_ID>" \
  --labels "spec:004-thin-wrapper-redesign,phase:polish,component:docs" \
  --priority 2
```

**Documentation Acceptance**: All documentation updated, accurate, and helpful for users and contributors.

**Checkpoint**: ✅ Documentation complete

## Implementation Strategy

### MVP Scope (Minimum Viable Product)

**Recommended MVP**: User Story 2 (sl doctor) + User Story 1 (sl new with prerequisites)

**Rationale**:
- US2 (sl doctor) is self-contained and provides immediate value for troubleshooting
- US1 (sl new) is the core user flow and depends on foundational work
- Together they demonstrate the thin wrapper architecture working end-to-end
- Can be released and tested by early adopters before completing other stories

**MVP Phases**:
1. Phase 1: Setup and Cleanup (remove old code)
2. Phase 2: Foundational - Metadata System (YAML infrastructure)
3. Phase 3: Foundational - Prerequisites Checker (tool detection)
4. Phase 4: US2 - sl doctor (diagnostic tool)
5. Phase 5: US1 - Enhanced Bootstrap (main user flow)

**Post-MVP Increments**:
- Increment 2: US4 (sl init) - existing project adoption
- Increment 3: US3 (sl deps) + Migration Support - dependency management
- Increment 4: Documentation and Polish

### Parallel Execution Opportunities

**Within Phase 1** (Setup): All cleanup tasks can run in parallel except sl-odx depends on sl-5k8

**Across Phases**:
- Phase 2 (Metadata) and Phase 3 (Prerequisites) are independent → **Run in parallel**
- After Phases 2+3 complete: Phase 4 (US2), Phase 5 (US1), Phase 6 (US4), Phase 7 (US3) can run in parallel (different commands)

**Maximum Parallelism**:
```
Phase 1 (Setup)
    ├─ Parallel: sl-pt7, sl-bzx, sl-t13, sl-5ey, sl-5k8
    └─ Sequential: sl-odx (after sl-5k8)

Phase 2 (Metadata) ║ Phase 3 (Prerequisites)  ← Parallel phases
    ↓                    ↓
Phase 4 (US2) ║ Phase 5 (US1) ║ Phase 6 (US4) ║ Phase 7 (US3)  ← All parallel

Phase 8 (Migration) depends on Phase 2 only

Phase 9 (Polish) depends on all
```

### Story Independence Verification

| Story | Dependencies | Can Deploy Independently? | Independent Test |
|-------|--------------|---------------------------|------------------|
| **US5** | None | ✅ Yes (Phase 1) | Verify commands removed, codebase compiles |
| **US2** | Prerequisites pkg | ✅ Yes | Run `sl doctor`, verify tool status reported |
| **US1** | Metadata + Prerequisites | ✅ Yes | Run `sl new`, verify project created with YAML |
| **US4** | Metadata + Prerequisites | ✅ Yes | Run `sl init`, verify SpecLedger added to existing project |
| **US3** | Metadata only | ✅ Yes | Run `sl deps add/list/resolve`, verify YAML updated |

**All user stories are independently deployable and testable** ✅

## Summary

**Total Phases**: 9
**Total Tasks**: ~46 (exact count after all created in Beads)
**Parallel Phases**: 2 (Metadata + Prerequisites)
**MVP Phases**: 5 (Phases 1-5)
**Estimated Parallel Execution Time**: ~40% reduction with 4 parallel workers

**Critical Path**:
Phase 1 → Phase 2 → Phase 5 (US1) → Phase 9

**Success Metrics**:
- 100% duplicate commands removed ✓
- YAML metadata system functional ✓
- Prerequisites checker works in interactive/CI modes ✓
- All 5 user stories independently testable ✓
- Migration path from .mod to YAML provided ✓

## Next Steps

1. **Create remaining Beads tasks** using commands provided in each phase
2. **Prioritize MVP** (Phases 1-5) for initial implementation
3. **Assign tasks** to developers or AI agents
4. **Execute in parallel** where possible (Phases 2+3, then 4+5+6+7)
5. **Test each user story** independently after its phase completes
6. **Deploy incrementally** (MVP first, then additional stories)

Use `bd ready --label "spec:004-thin-wrapper-redesign"` to find next available tasks to work on.
