# Task Breakdown: Fix SpecLedger Dependencies Integration

**Feature**: 008-fix-sl-deps
**Epic ID**: SL-7nu
**Branch**: `008-fix-sl-deps`
**Date**: 2026-02-09

## Overview

This document organizes all implementation tasks for the SpecLedger dependencies integration fix by user story, enabling independent implementation and testing of each story.

**Sources**:
- Specification: [spec.md](./spec.md)
- Implementation Plan: [plan.md](./plan.md)
- Research: [research.md](./research.md)
- Data Model: [data-model.md](./data-model.md)
- Contracts: [contracts/cli-api.md](./contracts/cli-api.md)
- Quickstart: [quickstart.md](./quickstart.md)

---

## Epic: SL-7nu - Fix SpecLedger Dependencies Integration

**Description**: Add artifact_path configuration to specledger.yaml for current project and dependencies. Auto-discover artifact paths from SpecLedger repos, support manual --artifact-path flag for non-SpecLedger repos. Implement auto-download on sl deps add. Update Claude Code integration with comprehensive skill documentation.

**Labels**: `spec:008-fix-sl-deps`, `component:cli`

---

## Task Filter Commands

```bash
# List all tasks for this feature
bd list --label "spec:008-fix-sl-deps"

# Show ready tasks (unblocked, unassigned)
bd ready --label "spec:008-fix-sl-deps"

# Show tasks by phase
bd list --label "spec:008-fix-sl-deps,phase:setup"
bd list --label "spec:008-fix-sl-deps,phase:foundational"
bd list --label "spec:008-fix-sl-deps,story:US1"
bd list --label "spec:008-fix-sl-deps,story:US2"

# Show tasks by story
bd list --label "spec:008-fix-sl-deps,story:US1" -n 10
```

---

## Phase 1: Setup (Project Initialization)

**Status**: Pending
**Feature**: Setup Phase
**Parent**: SL-7nu

### bd create Command

```bash
bd create "Setup Phase: Dependencies Integration" \
  --description "Initialize project infrastructure for dependencies integration feature" \
  --type feature \
  --deps "parent-child:SL-7nu" \
  --labels "spec:008-fix-sl-deps,phase:setup,component:infra" \
  --priority 1 \
  --design "Create pkg/deps/ package structure and prepare for artifact path implementation" \
  --acceptance "Package structure created, ready for implementation tasks"
```

### Tasks

#### T001: Create pkg/deps package structure

```bash
bd create "Create pkg/deps package structure" \
  --description "Create the new pkg/deps/ package directory with placeholder files for artifact path discovery, cache operations, and reference resolution" \
  --type task \
  --deps "parent-child:<setup-feature-id>" \
  --labels "spec:008-fix-sl-deps,phase:setup,component:deps" \
  --priority 1 \
  --design "Create pkg/deps/ directory with resolver.go, cache.go, reference.go placeholder files. Add package documentation." \
  --acceptance "pkg/deps/ directory exists with all placeholder files, package compiles"
```

#### T002: Update templates for embedded artifacts

```bash
bd create "Update embedded templates for Claude commands" \
  --description "Ensure .claude/commands/ and .claude/skills/ templates include the add-deps.md and remove-deps.md files, and specledger-deps/ skill structure" \
  --type task \
  --deps "parent-child:<setup-feature-id>" \
  --labels "spec:008-fix-sl-deps,phase:setup,component:templates" \
  --priority 2 \
  --design "Update pkg/embedded/templates/specledger/.claude/commands/ and pkg/embedded/templates/specledger/.claude/skills/specledger-deps/ with current versions" \
  --acceptance "Embedded templates include all required Claude Code integration files"
```

---

## Phase 2: Foundational (Blocking Prerequisites)

**Status**: Pending
**Feature**: Foundational Phase
**Parent**: SL-7nu

These tasks MUST complete before any user story can be implemented.

### bd create Command

```bash
bd create "Foundational Phase: Core Data Model Changes" \
  --description "Update metadata schema to support artifact_path for projects and dependencies" \
  --type feature \
  --deps "parent-child:SL-7nu" \
  --labels "spec:008-fix-sl-deps,phase:foundational,component:metadata" \
  --priority 1 \
  --design "Add ArtifactPath field to ProjectMetadata, remove Path field from Dependency, update validation and loading logic" \
  --acceptance "Metadata schema updated, backward compatible, all tests pass"
```

### Tasks

#### T003: Add ArtifactPath to ProjectMetadata schema

```bash
bd create "Add ArtifactPath field to ProjectMetadata" \
  --description "Add ArtifactPath field to pkg/cli/metadata/schema.go ProjectMetadata struct to specify where current project's artifacts are located" \
  --type task \
  --deps "parent-child:<foundational-feature-id>" \
  --labels "spec:008-fix-sl-deps,phase:foundational,story:US2,requirement:FR-005,component:metadata" \
  --priority 0 \
  --design "Update pkg/cli/metadata/schema.go: Add ArtifactPath string field to ProjectMetadata struct with yaml:artifact_path,omitempty tag. Add GetArtifactPath() helper method that returns field value or default 'specledger/'." \
  --acceptance "ProjectMetadata has ArtifactPath field, GetArtifactPath() helper exists, defaults to 'specledger/'"
```

#### T004: Remove Path field from Dependency schema

```bash
bd create "Remove Path field from Dependency struct" \
  --description "Remove the Path field from Dependency struct since alias is now used as the reference path within project's artifact_path" \
  --type task \
  --deps "parent-child:<foundational-feature-id>,after:T003" \
  --labels "spec:008-fix-sl-deps,phase:foundational,requirement:FR-011,component:metadata" \
  --priority 1 \
  --design "Update pkg/cli/metadata/schema.go: Remove Path field from Dependency struct. Update all code that references dep.Path to use dep.Alias instead." \
  --acceptance "Dependency struct has no Path field, all references updated to use Alias"
```

#### T005: Add ArtifactPath to Dependency schema

```bash
bd create "Add ArtifactPath field to Dependency struct" \
  --description "Add ArtifactPath field to Dependency struct for storing where artifacts are located within the dependency repository" \
  --type task \
  --deps "parent-child:<foundational-feature-id>,after:T003" \
  --labels "spec:008-fix-sl-deps,phase:foundational,requirement:FR-009,requirement:FR-010,component:metadata" \
  --priority 1 \
  --design "Update pkg/cli/metadata/schema.go: Add ArtifactPath string field to Dependency struct with yaml:artifact_path,omitempty tag. This stores auto-discovered or manually specified artifact path." \
  --acceptance "Dependency struct has ArtifactPath field, defaults to empty string"
```

#### T006: Update metadata validation for artifact_path

```bash
bd create "Add validation for artifact_path field" \
  --description "Add validation logic to ensure artifact_path values are valid relative paths without parent directory references" \
  --type task \
  --deps "parent-child:<foundational-feature-id>,after:T003,after:T005" \
  --labels "spec:008-fix-sl-deps,phase:foundational,requirement:FR-015,component:metadata" \
  --priority 2 \
  --design "Update pkg/cli/metadata/schema.go: Add validation in Validate() methods for both ProjectMetadata and Dependency. artifact_path must be relative (not starting with /), must not contain .. for security." \
  --acceptance "Validation rejects absolute paths and parent directory references, clear error messages"
```

#### T007: Ensure backward compatibility for artifact_path

```bash
bd create "Add backward compatibility for loading old specledger.yaml" \
  --description "Ensure projects without artifact_path field default to 'specledger/' for backward compatibility" \
  --type task \
  --deps "parent-child:<foundational-feature-id>,after:T003" \
  --labels "spec:008-fix-sl-deps,phase:foundational,component:metadata" \
  --priority 1 \
  --design "Update pkg/cli/metadata/ loader: When loading specledger.yaml, if artifact_path is empty or missing, set to 'specledger/'. Use omitempty in YAML tags so empty values aren't written." \
  --acceptance "Old specledger.yaml files load successfully with default artifact_path"
```

---

## Phase 3: User Story 1 - Add Dependencies Auto-Downloads (P1)

**Status**: Pending
**Story**: US1 - Add Dependencies Automatically Downloads to Cache
**Priority**: P1
**Goal**: Automatically download and cache dependency repositories when adding them

**Independent Test**: Can be tested by adding a dependency with `sl deps add` and verifying the repository is cloned to `~/.specledger/cache/<dir-name>/`

**Acceptance Criteria**:
1. `sl deps add <url> --alias <name>` automatically clones to cache
2. Error when adding duplicate dependency
3. Custom branches are checked out correctly

### bd create Command

```bash
bd create "User Story 1: Auto-Download on Add" \
  --description "Implement automatic download and caching of dependencies when they are added via sl deps add" \
  --type feature \
  --deps "parent-child:SL-7nu" \
  --labels "spec:008-fix-sl-deps,phase:us1,story:US1,component:deps" \
  --priority 1 \
  --design "Modify runAddDependency in pkg/cli/commands/deps.go to automatically download dependencies after adding to metadata. Use go-git/v5 library for Git operations." \
  --acceptance "Dependencies are automatically downloaded to ~/.specledger/cache/ on add, commit SHA resolved"
```

### Tasks

#### T008: Make --alias required for sl deps add

```bash
bd create "Make --alias flag required for sl deps add" \
  --description "Update sl deps add command to require the --alias flag instead of treating it as optional" \
  --type task \
  --deps "parent-child:<us1-feature-id>,after:T004" \
  --labels "spec:008-fix-sl-deps,story:US1,requirement:FR-011,component:cli" \
  --priority 0 \
  --design "Update pkg/cli/commands/deps.go: Change alias flag from optional to required. Add validation that alias is provided. Error message should be clear: 'alias is required (use --alias <name>)'" \
  --acceptance "sl deps add fails with clear error when --alias not provided"
```

#### T009: Implement auto-download in sl deps add

```bash
bd create "Implement auto-download functionality in sl deps add" \
  --description "After adding dependency to metadata, automatically clone the repository to ~/.specledger/cache/<alias>/" \
  --type task \
  --deps "parent-child:<us1-feature-id>,after:T008" \
  --labels "spec:008-fix-sl-deps,story:US1,requirement:FR-001,component:deps" \
  --priority 0 \
  --design "Update pkg/cli/commands/deps.go runAddDependency: After saving metadata, call download function that clones repo to ~/.specledger/cache/<alias>/. Use go-git/v5 PlainClone for Git operations. Resolve and store commit SHA." \
  --acceptance "sl deps add automatically downloads dependency, shows download progress, stores commit SHA"
```

#### T010: Implement go-git based clone operations

```bash
bd create "Implement Git clone using go-git/v5 library" \
  --description "Create Git clone functionality using go-git/v5 library instead of system git commands" \
  --type task \
  --deps "parent-child:<us1-feature-id>" \
  --labels "spec:008-fix-sl-deps,story:US1,component:deps" \
  --priority 1 \
  --design "Create pkg/deps/cache.go: Implement CloneDependency() function using go-git/v5. Handles shallow clone, branch checkout, commit resolution. Reuses patterns from internal/spec/resolver.go." \
  --acceptance "Dependencies can be cloned using go-git, commit SHA resolved and stored"
```

#### T011: Handle duplicate dependency detection

```bash
bd create "Add duplicate dependency detection in sl deps add" \
  --description "Detect and error when attempting to add a dependency that already exists (by URL or alias)" \
  --type task \
  --deps "parent-child:<us1-feature-id>,after:T008" \
  --labels "spec:008-fix-sl-deps,story:US1,component:cli" \
  --priority 2 \
  --design "Update pkg/cli/commands/deps.go runAddDependency: Check if URL or alias already exists in metadata.Dependencies before adding. Return error 'dependency already exists: <url>' or 'alias already exists: <alias>'" \
  --acceptance "Adding duplicate dependency fails with clear error message"
```

#### T012: Implement custom branch checkout

```bash
bd create "Support custom branch checkout during add" \
  --description "When adding a dependency with a specific branch, checkout that branch after cloning" \
  --type task \
  --deps "parent-child:<us1-feature-id>,after:T010" \
  --labels "spec:008-fix-sl-deps,story:US1,component:deps" \
  --priority 2 \
  --design "Update pkg/deps/cache.go: CloneDependency should accept branch parameter. After clone, use worktree.Checkout() to switch to specified branch. Default to 'main' if not specified." \
  --acceptance "Dependencies added with custom branch are checked out to that branch"
```

#### T013: Add network error handling for download

```bash
bd create "Handle network errors during dependency download" \
  --description "Gracefully handle network errors, authentication failures, and invalid repository URLs during download" \
  --type task \
  --deps "parent-child:<us1-feature-id>,after:T010" \
  --labels "spec:008-fix-sl-deps,story:US1,component:deps" \
  --priority 2 \
  --design "Update pkg/deps/cache.go: Wrap go-git operations with error handling. Return clear error messages for: auth failures, network timeouts, invalid URLs, repository not found. Log warnings for non-fatal issues." \
  --acceptance "Network errors return clear messages, invalid URLs are detected early"
```

---

## Phase 4: User Story 2 - Configure Artifact Path (P1)

**Status**: Pending
**Story**: US2 - Configure Artifact Path for Current Project
**Priority**: P1
**Goal**: Configure artifact_path in specledger.yaml for current project

**Independent Test**: Can be tested by running `sl new` and verifying artifact_path field exists

**Acceptance Criteria**:
1. `sl new` creates project with `artifact_path: specledger/`
2. `sl init` detects existing artifact directory
3. Tools use configured artifact_path

### bd create Command

```bash
bd create "User Story 2: Configure Project Artifact Path" \
  --description "Add artifact_path configuration to specledger.yaml for the current project, set by sl new and detected by sl init" \
  --type feature \
  --deps "parent-child:SL-7nu" \
  --labels "spec:008-fix-sl-deps,phase:us2,story:US2,component:cli" \
  --priority 1 \
  --design "Update pkg/cli/commands/new.go and init.go to set/detect artifact_path field in ProjectMetadata" \
  --acceptance "New projects have artifact_path: specledger/, init detects existing structure"
```

### Tasks

#### T014: Update sl new to set default artifact_path

```bash
bd create "Update sl new command to set artifact_path" \
  --description "When creating a new project with sl new, automatically set artifact_path to 'specledger/' in the metadata" \
  --type task \
  --deps "parent-child:<us2-feature-id>,after:T003" \
  --labels "spec:008-fix-sl-deps,story:US2,requirement:FR-007,component:cli" \
  --priority 1 \
  --design "Update pkg/cli/commands/new.go: After creating ProjectMetadata struct, set meta.ArtifactPath = 'specledger/' before saving to specledger.yaml." \
  --acceptance "sl new creates projects with artifact_path: specledger/ in metadata"
```

#### T015: Update sl init to detect artifact_path

```bash
bd create "Update sl init command to detect artifact_path" \
  --description "When initializing SpecLedger in an existing repo, detect common artifact directories and prompt user for confirmation" \
  --type task \
  --deps "parent-child:<us2-feature-id>,after:T003" \
  --labels "spec:008-fix-sl-deps,story:US2,requirement:FR-008,component:cli" \
  --priority 1 \
  --design "Update pkg/cli/commands/init.go: Scan for common directories: specledger/, specs/, docs/specs/, documentation/. Prompt user: 'Use <detected> as artifact path? [Y/n]'. Store selected path in metadata." \
  --acceptance "sl init detects common directories and sets artifact_path based on user choice"
```

---

## Phase 5: User Story 3 - Discover Artifact Path (P1)

**Status**: Pending
**Story**: US3 - Discover Artifact Path in Dependency Repositories
**Priority**: P1
**Goal**: Auto-discover artifact_path from SpecLedger repos, manual for others

**Independent Test**: Can be tested by adding SpecLedger repo (auto-detect) and non-SpecLedger repo (manual)

**Acceptance Criteria**:
1. SpecLedger repos: auto-discover artifact_path
2. Non-SpecLedger repos: require --artifact-path flag
3. `sl deps list` shows artifact_path

### bd create Command

```bash
bd create "User Story 3: Auto-Discover Artifact Path" \
  --description "Automatically discover artifact_path from SpecLedger repository dependencies, require manual --artifact-path flag for non-SpecLedger repos" \
  --type feature \
  --deps "parent-child:SL-7nu" \
  --labels "spec:008-fix-sl-deps,phase:us3,story:US3,component:deps" \
  --priority 1 \
  --design "Create pkg/deps/resolver.go with DetectArtifactPathFromSpecLedgerRepo() function. Update sl deps add to auto-detect or require --artifact-path flag." \
  --acceptance "SpecLedger repos have artifact_path auto-discovered, non-SpecLedger require --artifact-path"
```

### Tasks

#### T016: Create artifact path discovery functions

```bash
bd create "Create artifact path detection functions" \
  --description "Create pkg/deps/resolver.go with functions to detect artifact_path from cloned dependency repositories" \
  --type task \
  --deps "parent-child:<us3-feature-id>" \
  --labels "spec:008-fix-sl-deps,story:US3,requirement:FR-009,component:deps" \
  --priority 0 \
  --design "Create pkg/deps/resolver.go: Implement DetectArtifactPathFromSpecLedgerRepo(repoPath string) that reads specledger.yaml and returns artifact_path value. Implement DetectArtifactPathFromRemote() for full workflow." \
  --acceptance "resolver.go created with functions to read artifact_path from specledger.yaml"
```

#### T017: Integrate auto-discovery into sl deps add

```bash
bd create "Integrate artifact path auto-discovery into sl deps add" \
  --description "After cloning dependency, attempt to read its specledger.yaml to auto-discover artifact_path for SpecLedger repos" \
  --type task \
  --deps "parent-child:<us3-feature-id>,after:T009,after:T016" \
  --labels "spec:008-fix-sl-deps,story:US3,requirement:FR-009,component:cli" \
  --priority 0 \
  --design "Update pkg/cli/commands/deps.go runAddDependency: After downloading dependency, call DetectArtifactPathFromSpecLedgerRepo() on cached path. If specledger.yaml found and has artifact_path, store it. If not found for SpecLedger repo, warn user." \
  --acceptance "SpecLedger repo dependencies have artifact_path auto-discovered and stored"
```

#### T018: Add --artifact-path flag for manual specification

```bash
bd create "Add --artifact-path flag to sl deps add" \
  --description "Add --artifact-path flag for manually specifying artifact path when adding non-SpecLedger repository dependencies" \
  --type task \
  --deps "parent-child:<us3-feature-id>" \
  --labels "spec:008-fix-sl-deps,story:US3,requirement:FR-010,component:cli" \
  --priority 0 \
  --design "Update pkg/cli/commands/deps.go: Add --artifact-path string flag to VarAddCmd. In runAddDependency, if flag provided, store value directly to dep.ArtifactPath. If SpecLedger repo and flag not provided, auto-discover." \
  --acceptance "--artifact-path flag works, value stored in dependency metadata"
```

#### T019: Add validation for non-SpecLedger repos

```bash
bd create "Validate artifact_path for non-SpecLedger dependencies" \
  --description "Require --artifact-path flag for non-SpecLedger repositories to ensure artifact path is explicitly specified" \
  --type task \
  --deps "parent-child:<us3-feature-id>,after:T017,after:T018" \
  --labels "spec:008-fix-sl-deps,story:US3,requirement:FR-010,component:cli" \
  --priority 2 \
  --design "Update pkg/cli/commands/deps.go: After auto-discovery attempt, if dependency is not a SpecLedger repo (no specledger.yaml found) and --artifact-path not provided, error: 'artifact_path must be specified for non-SpecLedger repositories (use --artifact-path <path>)'" \
  --acceptance "Non-SpecLedger repos require --artifact-path flag, clear error if missing"
```

#### T020: Update sl deps list to show artifact_path

```bash
bd create "Update sl deps list to display artifact_path" \
  --description "Modify sl deps list output to show the artifact_path for each dependency" \
  --type task \
  --deps "parent-child:<us3-feature-id>" \
  --labels "spec:008-fix-sl-deps,story:US3,requirement:FR-012,component:cli" \
  --priority 2 \
  --design "Update pkg/cli/commands/deps.go runListDependencies: Add line showing 'Artifact Path: <path>' for each dependency. Show '(auto-discovered)' if auto-detected from SpecLedger repo." \
  --acceptance "sl deps list shows artifact_path for each dependency"
```

---

## Phase 6: User Story 4 - Reference Artifacts (P1)

**Status**: Pending
**Story**: US4 - Reference Artifacts Across Repositories
**Priority**: P1
**Goal**: Reference artifacts using `alias:artifact` syntax

**Independent Test**: Add dependency, reference artifact, verify correct path resolution

**Acceptance Criteria**:
1. `alias:artifact.md` resolves to `artifact_path/alias/artifact.md`
2. Nested paths work correctly
3. Clear errors for invalid references

### bd create Command

```bash
bd create "User Story 4: Cross-Repository Artifact References" \
  --description "Implement reference resolution for cross-repository artifact references using alias:artifact syntax" \
  --type feature \
  --deps "parent-child:SL-7nu" \
  --labels "spec:008-fix-sl-deps,phase:us4,story:US4,component:deps" \
  --priority 2 \
  --design "Create pkg/deps/reference.go with ResolveArtifactReference() function that combines project.artifact_path + dependency.alias + artifact_name" \
  --acceptance "Artifact references resolve correctly, clear errors for missing files"
```

### Tasks

#### T021: Create reference resolution function

```bash
bd create "Create ResolveArtifactReference function" \
  --description "Create function to resolve alias:artifact references to full file paths using project.artifact_path and dependency.alias" \
  --type task \
  --deps "parent-child:<us4-feature-id>,after:T003" \
  --labels "spec:008-fix-sl-deps,story:US4,requirement:FR-013,component:deps" \
  --priority 1 \
  --design "Create pkg/deps/reference.go: Implement ResolveArtifactReference(projectMeta *ProjectMetadata, reference string) (string, error). Parse 'alias:artifact' format. Find dependency by alias. Build path: project.artifact_path + dep.Alias + artifact." \
  --acceptance "References are resolved to correct file paths"
```

#### T022: Add validation for resolved paths

```bash
bd create "Add validation for resolved artifact paths" \
  --description "Validate that resolved artifact paths point to existing files, return clear errors for missing files" \
  --type task \
  --deps "parent-child:<us4-feature-id>,after:T021" \
  --labels "spec:008-fix-sl-deps,story:US4,requirement:FR-015,component:deps" \
  --priority 2 \
  --design "Update pkg/deps/reference.go: After resolving path, use os.Stat() to check if file exists. If not found, return error: 'artifact not found: <full_path>'. Include both resolved path and cache location in error for debugging." \
  --acceptance "Missing artifacts return clear error messages with full paths"
```

#### T023: Add support for nested artifact paths

```bash
bd create "Add support for nested artifact path references" \
  --description "Handle references to artifacts within subdirectories of the dependency's artifact_path" \
  --type task \
  --deps "parent-child:<us4-feature-id>,after:T021" \
  --labels "spec:008-fix-sl-deps,story:US4,requirement:FR-014,component:deps" \
  --priority 2 \
  --design "Update pkg/deps/reference.go: Parse artifact name to support nested paths like 'subdir/file.md'. When resolving, if dependency has artifact_path, construct: project.artifact_path + alias + dependency.artifact_path + artifact." \
  --acceptance "Nested references like 'api-docs:openapi/api.yaml' resolve correctly"
```

---

## Phase 7: User Story 5 - Claude Code Commands (P1)

**Status**: Pending
**Story**: US5 - Create Claude Code Command Files for Core Deps Operations
**Priority**: P1
**Goal**: Update add-deps.md and remove-deps.md command files

**Independent Test**: Verify command files exist and contain clear instructions

**Acceptance Criteria**:
1. add-deps.md updated with new flags and behavior
2. remove-deps.md is current
3. Files are in both project and embedded templates

### bd create Command

```bash
bd create "User Story 5: Update Claude Code Command Files" \
  --description "Update .claude/commands/ files for add-deps and remove-deps with new flags, required alias, and auto-download behavior" \
  --type feature \
  --deps "parent-child:SL-7nu" \
  --labels "spec:008-fix-sl-deps,phase:us5,story:US5,component:claude" \
  --priority 2 \
  --design "Update .claude/commands/specledger.add-deps.md and .claude/commands/specledger.remove-deps.md with current command signatures including --artifact-path flag, required --alias, and auto-download behavior" \
  --acceptance "Command files reflect current CLI behavior, embedded templates updated"
```

### Tasks

#### T024: Update specledger.add-deps.md command file

```bash
bd create "Update specledger.add-deps.md command file" \
  --description "Update the Claude Code command file for add-deps to reflect new signature: required --alias, --artifact-path flag, auto-download behavior" \
  --type task \
  --deps "parent-child:<us5-feature-id>" \
  --labels "spec:008-fix-sl-deps,story:US5,requirement:FR-017,component:claude" \
  --priority 1 \
  --design "Update .claude/commands/specledger.add-deps.md: Update usage to show --alias is required. Add --artifact-path flag documentation. Explain auto-download happens automatically. Add examples for SpecLedger vs non-SpecLedger repos." \
  --acceptance "add-deps.md reflects current CLI behavior with all new flags"
```

#### T025: Update embedded add-deps.md template

```bash
bd create "Update embedded specledger.add-deps.md template" \
  --description "Update the embedded template so new projects created by sl new include the updated add-deps.md command file" \
  --type task \
  --deps "parent-child:<us5-feature-id>,after:T024" \
  --labels "spec:008-fix-sl-deps,story:US5,component:templates,component:claude" \
  --priority 2 \
  --design "Copy updated .claude/commands/specledger.add-deps.md to pkg/embedded/templates/specledger/.claude/commands/specledger.add-deps.md" \
  --acceptance "Embedded template matches project command file"
```

#### T026: Review and update specledger.remove-deps.md

```bash
bd create "Review and update specledger.remove-deps.md command file" \
  --description "Review the remove-deps.md command file for accuracy and update if needed to reflect current behavior" \
  --type task \
  --deps "parent-child:<us5-feature-id>" \
  --labels "spec:008-fix-sl-deps,story:US5,requirement:FR-017,component:claude" \
  --priority 2 \
  --design "Review .claude/commands/specledger.remove-deps.md: Ensure examples and usage are current. Verify error messages match implementation. Update if needed." \
  --acceptance "remove-deps.md is current and accurate"
```

---

## Phase 8: User Story 6 - Update Skill Documentation (P2)

**Status**: Pending
**Story**: US6 - Update SpecLedger-Deps Skill Documentation
**Priority**: P2
**Goal**: Comprehensive documentation for all deps commands

**Independent Test**: Read skill file and verify all commands documented

**Acceptance Criteria**:
1. All deps commands documented (add, remove, list, update, resolve)
2. artifact_path concept explained
3. Reference resolution explained

### bd create Command

```bash
bd create "User Story 6: Update SpecLedger-Deps Skill Documentation" \
  --description "Update .claude/skills/specledger-deps/SKILL.md with comprehensive documentation for all deps commands, artifact_path discovery, and reference resolution" \
  --type feature \
  --deps "parent-child:SL-7nu" \
  --labels "spec:008-fix-sl-deps,phase:us6,story:US6,component:claude" \
  --priority 2 \
  --design "Update .claude/skills/specledger-deps/SKILL.md with: artifact_path concept explanation, all 5 deps commands (add, remove, list, update, resolve), reference resolution format, workflow examples" \
  --acceptance "Skill file contains comprehensive documentation for entire deps workflow"
```

### Tasks

#### T027: Update skill with artifact_path documentation

```bash
bd create "Add artifact_path concept documentation to skill" \
  --description "Add section explaining the artifact_path concept: what it is, how it's used, difference between project and dependency artifact_path" \
  --type task \
  --deps "parent-child:<us6-feature-id>" \
  --labels "spec:008-fixsl-deps,story:US6,requirement:FR-019,component:claude" \
  --priority 1 \
  --design "Update .claude/skills/specledger-deps/SKILL.md: Add 'Understanding artifact_path' section explaining: project artifact_path (where YOUR artifacts are), dependency artifact_path (where DEPENDENCY artifacts are), reference resolution combining both." \
  --acceptance "Skill has clear explanation of artifact_path concept with examples"
```

#### T028: Document all deps commands in skill

```bash
bd create "Document all deps operations in skill file" \
  --description "Add comprehensive documentation for all 5 deps commands: add, remove, list, update, resolve with examples and usage patterns" \
  --type task \
  --deps "parent-child:<us6-feature-id>" \
  --labels "spec:008-fix-sl-deps,story:US6,requirement:FR-019,component:claude" \
  --priority 1 \
  --design "Update .claude/skills/specledger-deps/SKILL.md: Add sections for each command: sl deps add (with auto-download), sl deps remove, sl deps list, sl deps update, sl deps resolve (for manual refresh). Include examples and when to use each." \
  --acceptance "All 5 deps commands documented in skill with examples"
```

#### T029: Document reference resolution format

```bash
bd create "Document alias:artifact reference format in skill" \
  --description "Add documentation for the cross-repository reference format and how resolution works" \
  --type task \
  --deps "parent-child:<us6-feature-id>,after:T021" \
  --labels "spec:008-fix-sl-deps,story:US6,requirement:FR-019,component:claude" \
  --priority 2 \
  --design "Update .claude/skills/specledger-deps/SKILL.md: Add 'Referencing Artifacts' section explaining <alias>:<artifact> format, resolution formula (project.artifact_path + alias + artifact), examples with different artifact_path configurations." \
  --acceptance "Reference format documented with clear examples"
```

#### T030: Update embedded skill template

```bash
bd create "Update embedded specledger-deps skill template" \
  --description "Update the embedded skill template so new projects include the updated deps documentation" \
  --type task \
  --deps "parent-child:<us6-feature-id>,after:T027,after:T028,after:T029" \
  --labels "spec:008-fix-sl-deps,story:US6,component:templates,component:claude" \
  --priority 2 \
  --design "Copy updated .claude/skills/specledger-deps/ to pkg/embedded/templates/specledger/.claude/skills/specledger-deps/" \
  --acceptance "Embedded skill template matches project skill file"
```

---

## Phase 9: Polish & Cross-Cutting Concerns

**Status**: Pending
**Feature**: Polish Phase
**Parent**: SL-7nu

### bd create Command

```bash
bd create "Polish Phase: Integration, Testing, Documentation" \
  --description "Complete implementation with integration tests, unit tests, and final documentation updates" \
  --type feature \
  --deps "parent-child:SL-7nu" \
  --labels "spec:008-fix-sl-deps,phase:polish,component:infra" \
  --priority 2 \
  --design "Complete all testing, update remaining documentation, verify all acceptance criteria met" \
  --acceptance "All tests pass, documentation complete, ready for release"
```

### Tasks

#### T031: Complete sl deps update implementation

```bash
bd create "Complete sl deps update command implementation" \
  --description "Finish the stub implementation of sl deps update to check for and apply updates from remote repositories" \
  --type task \
  --deps "parent-child:<polish-feature-id>" \
  --labels "spec:008-fix-sl-deps,phase:polish,component:cli" \
  --priority 2 \
  --design "Update pkg/cli/commands/deps.go runUpdateDependencies: For each dependency, fetch from remote, compare with resolved_commit. If newer, prompt for update. Support --yes flag for auto-update. Update commit SHA in metadata." \
  --acceptance "sl deps update checks for and applies updates from remote"
```

#### T032: Write unit tests for metadata changes

```bash
bd create "Write unit tests for metadata schema changes" \
  --description "Add unit tests for ArtifactPath field in ProjectMetadata and Dependency structs" \
  --type task \
  --deps "parent-child:<polish-feature-id>,after:T003,after:T004,after:T005,after:T006" \
  --labels "spec:008-fix-sl-deps,phase:polish,component:test" \
  --priority 2 \
  --design "Create pkg/cli/metadata/schema_test.go: Test GetArtifactPath() default behavior, test ArtifactPath validation (rejects absolute paths, ..), test Dependency ArtifactPath validation, test backward compatibility loading." \
  --acceptance "Unit tests cover all metadata schema changes, >80% coverage"
```

#### T033: Write unit tests for deps package

```bash
bd create "Write unit tests for deps package functions" \
  --description "Add unit tests for resolver, cache, and reference functions in pkg/deps/" \
  --type task \
  --deps "parent-child:<polish-feature-id>,after:T010,after:T016,after:T021" \
  --labels "spec:008-fix-sl-deps,phase:polish,component:test" \
  --priority 2 \
  --design "Create pkg/deps/resolver_test.go, cache_test.go, reference_test.go: Test artifact path detection, test clone operations, test reference resolution with various inputs, test error handling." \
  --acceptance "Unit tests for pkg/deps/ package, >80% coverage"
```

#### T034: Write integration tests for deps workflow

```bash
bd create "Write integration tests for full deps workflow" \
  --description "Add integration tests that test the complete dependency workflow: add, list, remove, update" \
  --type task \
  --deps "parent-child:<polish-feature-id>,after:T032,after:T033" \
  --labels "spec:008-fix-sl-deps,phase:polish,component:test" \
  --priority 2 \
  --design "Create tests/integration/deps_test.go: Test full workflow with real Git repositories (public repos), test auto-download, test artifact path discovery, test reference resolution, test error cases." \
  --acceptance "Integration tests cover main workflows, can run against test repos"
```

#### T035: Update README with artifact_path documentation

```bash
bd create "Update README with artifact_path usage documentation" \
  --description "Update project README to document the new artifact_path feature and how to use dependencies" \
  --type task \
  --deps "parent-child:<polish-feature-id>,after:T027,after:T028" \
  --labels "spec:008-fix-sl-deps,phase:polish,component:docs" \
  --priority 3 \
  --design "Update README.md: Add section on artifact_path configuration, add examples of using dependencies, document reference format, include quickstart link" \
  --acceptance "README documents artifact_path and dependencies features"
```

---

## Summary Statistics

| Metric | Count |
|--------|-------|
| **Total User Stories** | 6 (5 P1, 1 P2) |
| **Total Tasks** | 35 |
| **Total Phases** | 9 |
| **Setup Tasks** | 2 |
| **Foundational Tasks** | 5 |
| **US1 (Auto-Download)** | 6 tasks |
| **US2 (Artifact Path Config)** | 2 tasks |
| **US3 (Auto-Discovery)** | 5 tasks |
| **US4 (References)** | 3 tasks |
| **US5 (Command Files)** | 3 tasks |
| **US6 (Skill Docs)** | 4 tasks |
| **Polish Tasks** | 5 tasks |

---

## Story Testability

| Story | Independent Test | Parallelizable |
|-------|------------------|---------------|
| US1 - Auto-Download | ✅ Yes - Add dep, verify cache | ✅ Yes (after foundational) |
| US2 - Artifact Path Config | ✅ Yes - Run sl new, verify field | ✅ Yes (after foundational) |
| US3 - Auto-Discovery | ✅ Yes - Add SpecLedger/non-SpecLedger dep | ✅ Yes (after US1) |
| US4 - References | ✅ Yes - Add dep, reference artifact | ✅ Yes (after US3) |
| US5 - Command Files | ✅ Yes - Verify file content | ✅ Yes (can run parallel) |
| US6 - Skill Docs | ✅ Yes - Read skill, verify docs | ✅ Yes (can run parallel) |

---

## Dependencies Between Stories

```
Setup → Foundational → US1 → US3 → US4
                    → US2 ↗
                    → US5 (parallel, after US1)
                    → US6 (parallel, after US4)
                    → Polish (after all stories)
```

**Key Dependencies**:
- **Foundational** MUST complete before any story (metadata schema changes)
- **US1** provides auto-download used by US3
- **US3** provides artifact_path discovery used by US4
- **US5** and **US6** can run in parallel after US1
- **Polish** runs after all stories complete

---

## MVP Scope

**Recommended MVP**: US1 (Auto-Download) + US2 (Artifact Path Config) + Foundational

This MVP delivers:
1. Core data model changes (artifact_path fields)
2. Automatic dependency download on add
3. Project artifact_path configuration
4. Basic Claude Code integration (add/remove commands)

**Time Estimate**: ~3-5 days for MVP

**Incremental Delivery**: After MVP, add US3 (auto-discovery), then US4 (references), then polish.

---

## Suggested Execution Order

1. **Sprint 1** (MVP Core):
   - Setup (T001-T002)
   - Foundational (T003-T007)
   - US2: Artifact Path Config (T014-T015)

2. **Sprint 2** (Auto-Download):
   - US1: Auto-Download (T008-T013)

3. **Sprint 3** (Auto-Discovery):
   - US3: Auto-Discovery (T016-T020)

4. **Sprint 4** (References & Polish):
   - US4: References (T021-T023)
   - US5: Command Files (T024-T026)
   - US6: Skill Docs (T027-T030)

5. **Sprint 5** (Final Polish):
   - Polish (T031-T035)

---

## Next Steps

1. **Execute tasks**: Use `bd ready --label "spec:008-fix-sl-deps"` to see next available tasks
2. **Track progress**: Use `bd list --label "spec:008-fix-sl-deps"` to see all tasks
3. **Start implementation**: Begin with Phase 1 (Setup) or Phase 2 (Foundational)

**Suggested first command**:
```bash
bd ready --label "spec:008-fix-sl-deps,phase:setup"
```
