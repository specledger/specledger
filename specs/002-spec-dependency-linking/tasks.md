# Tasks: Spec Dependency Linking

**Epic ID**: sl-62h
**Feature Branch**: `002-spec-dependency-linking`
**Generated**: 2026-01-30
**Spec**: `/specs/002-spec-dependency-linking/spec.md`
**Plan**: `/specs/002-spec-dependency-linking/plan.md`
**Research**: `/specs/002-spec-dependency-linking/research.md`
**Data Model**: `/specs/002-spec-dependency-linking/data-model.md`
**Contracts**: `/specs/002-spec-dependency-linking/contracts/`
**Quickstart**: `/specs/002-spec-dependency-linking/quickstart.md`

## Overview

This task breakdown implements a golang-style dependency management system for specifications, organized by user stories to enable independent implementation and testing. Each user story has its own phase with specific tasks, ensuring incremental delivery and clear progress tracking.

## Task Generation Summary

- **Total Tasks**: 32 tasks across 5 user stories
- **Story Testability**: All stories independently verifiable
- **Parallel Opportunities**: 8 identified across stories
- **MVP Scope**: User Story 1 (P1) - Dependency Declaration and Resolution

## User Stories Breakdown

### P1: User Story 1 - Declare External Spec Dependencies
**Goal**: Enable teams to declare and resolve external specification dependencies
**Independent Test**: Create `spec.mod`, resolve dependencies, verify `spec.sum` generation
**Priority**: Critical (0) - Foundation for all other features

**Phase Commands**:
```bash
# Setup phase
bd create "Setup Phase" --description "Project initialization and CLI framework setup" --type "feature" --deps "parent-child:sl-62h" --labels "spec:002-spec-dependency-linking,phase:setup,component:cli" --priority 1

# US1 phase
bd create "US1 - Dependency Declaration" --description "Implement dependency declaration and resolution" --type "feature" --deps "parent-child:sl-62h" --labels "spec:002-spec-dependency-linking,phase:us1,component:cli" --priority 1

# Tasks within US1
bd create "Initialize Go project structure" --description "Create Go module and basic project structure for dependency management" --type "task" --deps "parent-child:sl-US1-feature-id" --labels "spec:002-spec-dependency-linking,story:US1,component:setup" --priority 1 --design "Create cmd/main.go with cobra CLI framework, go.mod with dependencies: go-git/v4, cobra, viper, golang.org/x/crypto" --acceptance "Go module initialized, can run 'go mod tidy' successfully"
```

### P2: User Story 2 - Reference External Specs in Current Spec
**Goal**: Enable referencing specific sections from external specifications
**Independent Test**: Add references, validate they resolve correctly
**Priority**: High (1) - Builds on US1 foundation

**Phase Commands**:
```bash
bd create "US2 - External References" --description "Implement external specification reference validation" --type "feature" --deps "parent-child:sl-62h" --labels "spec:002-spec-dependency-linking,phase:us2,component:cli" --priority 2

# Tasks within US2
bd create "Implement reference parser" --description "Create parser for markdown links in spec.md files" --type "task" --deps "parent-child:sl-US2-feature-id" --labels "spec:002-spec-dependency-linking,story:US2,component:parser" --priority 2 --design "Parse [Name](repo-url#spec-id#section-id) syntax using regex and validate components" --acceptance "Can parse valid references and report errors for invalid syntax"
```

### P3: User Story 3 - Update and Pin Dependency Versions
**Goal**: Enable updating dependencies and pinning to specific versions
**Independent Test**: Update dependencies, verify version changes and hash updates
**Priority**: Normal (2) - Builds on US1 and US2

### P4: User Story 4 - Detect and Resolve Dependency Conflicts
**Goal**: Detect and help resolve dependency conflicts
**Independent Test**: Create conflicting dependencies, verify conflict detection
**Priority**: Normal (2) - Builds on US1

### P5: User Story 5 - Vendor Dependencies for Offline Use
**Goal**: Enable copying external dependencies for offline use
**Independent Test**: Vendor dependencies, verify offline functionality
**Priority**: Low (3) - Optional enhancement

## Task Details

### Setup Phase (Phase 1)

T001 - Initialize Go project structure
```bash
bd create "Initialize Go project structure" --description "Create Go module and basic project structure for dependency management" --type "task" --deps "parent-child:sl-setup-id" --labels "spec:002-spec-dependency-linking,story:US1,component:setup" --priority 1 --design "Create cmd/main.go with cobra CLI framework, go.mod with dependencies: go-git/v4, cobra, viper, golang.org/x/crypto" --acceptance "Go module initialized, can run 'go mod tidy' successfully"
```

T002 - Setup CLI command structure
```bash
bd create "Setup CLI command structure" --description "Implement cobra CLI structure with subcommands for deps, refs, graph, vendor" --type "task" --deps "parent-child:sl-setup-id" --labels "spec:002-spec-dependency-linking,story:US1,component:cli" --priority 1 --design "Create root command with subcommands: deps add/list/resolve/update, refs validate/list, graph show, vendor" --acceptance "CLI shows help with all subcommands, basic structure works"
```

T003 - Setup configuration management
```bash
bd create "Setup configuration management" --description "Implement file-based configuration for auth tokens, cache settings, and tool preferences" --type "task" --deps "parent-child:sl-setup-id" --labels "spec:002-spec-dependency-linking,story:US1,component:config" --priority 1 --design "Use viper for config management, create .spec-config/auth.json and .spec-config/config.yaml" --acceptance "Can store and retrieve configuration values, supports environment variable overrides"
```

### Foundational Phase (Phase 2) - Must complete before any user story

T004 - Implement SpecManifest parser
```bash
bd create "Implement SpecManifest parser" --description "Create parser for spec.mod files with dependency declarations" --type "task" --priority 0 --design "Parse text format: require <repo-url> <version> <spec-path> with optional id <spec-id>" --acceptance "Can parse valid spec.mod and report errors for invalid syntax, supports all version formats"
```

T005 - Implement Lockfile entry structure
```bash
bd create "Implement Lockfile entry structure" --description "Create data structures for spec.sum entries with SHA-256 verification" --type "task" --priority 0 --design "Parse format: <repo-url> <commit-hash> <sha256-hash> <spec-path>" --acceptance "Can parse and validate lockfile entries, support hash verification"
```

T006 - Setup Git client abstraction
```bash
bd create "Setup Git client abstraction" --description "Create Git client interface using go-git for repository operations" --type "task" --priority 0 --design "Implement Repository interface with Clone, Fetch, Commit methods" --acceptance "Can clone repositories and fetch specific branches/commits"
```

T007 - Setup authentication framework
```bash
bd create "Setup authentication framework" --description "Implement authentication for private repositories using tokens and SSH keys" --type "task" --priority 0 --design "Support GitHub/GitLab tokens via environment variables and SSH keys from config" --acceptance "Can authenticate to private repositories with proper credentials"
```

T008 - Setup cache management
```bash
bd create "Setup cache management" --description "Implement LRU cache for resolved dependencies with SHA-256 verification" --type "task" --priority 0 --design "100MB cache with 1-hour TTL, store specs in .spec-cache/ directory" --acceptance "Cache speeds up repeated fetches, invalidates when content changes"
```

### User Story 1 Tasks (P1) - T009 to T016

T009 - Implement dependency declaration command
```bash
bd create "Implement dependency declaration command" --description "Create 'sl deps add' command to add external dependencies to spec.mod" --type "task" --deps "parent-child:sl-US1-id" --labels "spec:002-spec-dependency-linking,story:US1,component:cli" --priority 1 --design "Command: sl deps add <repo-url> [branch] [path] [--alias], create/update spec.mod" --acceptance "Can add dependencies with validation, preserve existing entries"
```

T010 - Implement manifest validation
```bash
bd create "Implement manifest validation" --description "Validate spec.mod syntax and ensure no duplicate dependencies" --type "task" --deps "parent-child:sl-US1-id" --labels "spec:002-spec-dependency-linking,story:US1,component:validation" --priority 1 --design "Check URL format, version syntax, spec path validity, uniqueness constraints" --acceptance "Rejects malformed declarations, allows valid syntax"
```

T011 - Implement dependency resolution engine
```bash
bd create "Implement dependency resolution engine" --description "Create resolver that fetches external specs and generates spec.sum" --type "task" --deps "parent-child:sl-US1-id" --labels "spec:002-spec-dependency-linking,story:US1,component:resolver" --priority 1 --design "Fetch from Git, transitive dependency discovery, parallel fetching for performance" --acceptance "Resolves dependencies in <10s for single repo, <30s for 10 repos"
```

T012 - Implement lockfile generation
```bash
bd create "Implement lockfile generation" --description "Generate spec.sum with cryptographic hashes and commit hashes" --type "task" --deps "parent-child:sl-US1-id" --labels "spec:002-spec-dependency-linking,story:US1,component:crypto" --priority 1 --design "Calculate SHA-256 for each spec content, record exact commit hashes" --acceptance "Lockfile contains verified hashes, matches go.sum format"
```

T013 - Implement content verification
```bash
bd create "Implement content verification" --description "Verify fetched content against lockfile hashes before use" --type "task" --deps "parent-child:sl-US1-id" --labels "spec:002-spec-dependency-linking,story:US1,component:crypto" --priority 1 --design "Compare SHA-256 hashes, warn on mismatch, offer to regenerate lockfile" --acceptance "Detects tampered content, provides clear error messages"
```

T014 - Implement resolution command
```bash
bd create "Implement resolution command" --description "Create 'sl deps resolve' command to resolve all dependencies" --type "task" --deps "parent-child:sl-US1-id" --labels "spec:002-spec-dependency-linking,story:US1,component:cli" --priority 1 --design "Command: sl deps resolve [--no-cache] [--deep], validate existing spec.mod" --acceptance "Resolves all declared dependencies, generates spec.sum"
```

T015 - Add error handling and logging
```bash
bd create "Add error handling and logging" --description "Implement structured logging and error handling for all operations" --type "task" --deps "parent-child:sl-US1-id" --labels "spec:002-spec-dependency-linking,story:US1,component:logging" --priority 1 --design "Use structured logging with context, meaningful error messages, progress indicators" --acceptance "Clear error messages, logs key decision points"
```

T016 - Write unit tests for core functionality
```bash
bd create "Write unit tests for core functionality" --description "Create comprehensive unit tests for dependency resolution and verification" --type "task" --deps "parent-child:sl-US1-id" --labels "spec:002-spec-dependency-linking,story:US1,component:test" --priority 1 --design "Use testify for assertions, 90%+ coverage, mock Git operations for testing" --acceptance "All unit tests pass, coverage meets requirements"
```

### User Story 2 Tasks (P2) - T017 to T020

T017 - Implement reference parser
```bash
bd create "Implement reference parser" --description "Create parser for markdown links in spec.md files" --type "task" --deps "parent-child:sl-US2-id" --labels "spec:002-spec-dependency-linking,story:US2,component:parser" --priority 2 --design "Parse [Name](repo-url#spec-id#section-id) syntax using regex and validate components" --acceptance "Can parse valid references and report errors for invalid syntax"
```

T018 - Implement reference resolver
```bash
bd create "Implement reference resolver" --description "Resolve external references to actual content in dependent specs" --type "task" --deps "parent-child:sl-US2-id" --labels "spec:002-spec-dependency-linking,story:US2,component:resolver" --priority 2 --design "Look up spec-id in manifest, resolve section-id, fetch content from cache or remote" --acceptance "Can resolve references to existing sections, report non-existent references"
```

T019 - Implement reference validation command
```bash
bd create "Implement reference validation command" --description "Create 'sl refs validate' to validate all external references" --type "task" --deps "parent-child:sl-US2-id" --labels "spec:002-spec-dependency-linking,story:US2,component:cli" --priority 2 --design "Command: sl refs validate [--strict], check all references in spec.md" --acceptance "Validates all references in <5s, reports specific failures"
```

T020 - Write integration tests for references
```bash
bd create "Write integration tests for references" --description "Create integration tests for external reference validation" --type "task" --deps "parent-child:sl-US2-id" --labels "spec:002-spec-dependency-linking,story:US2,component:test" --priority 2 --design "Test reference resolution with actual Git repositories, validate edge cases" --acceptance "Tests pass for all reference scenarios including broken links"
```

### User Story 3 Tasks (P3) - T021 to T024

T021 - Implement version update logic
```bash
bd create "Implement version update logic" --description "Create logic to update dependencies to latest compatible versions" --type "task" --deps "parent-child:sl-US3-id" --labels "spec:002-spec-dependency-linking,story:US3,component:resolver" --priority 2 --design "Support semantic version constraints, fetch latest compatible versions" --acceptance "Can update dependencies while respecting version constraints"
```

T022 - Implement update command
```bash
bd create "Implement update command" --description "Create 'sl deps update' to update dependencies to latest versions" --type "task" --deps "parent-child:sl-US3-id" --labels "spec:002-spec-dependency-linking,story:US3,component:cli" --priority 2 --design "Command: sl deps update [--force] [repo-url], show version differences" --acceptance "Updates dependencies in <2 minutes, shows migration notes"
```

T023 - Implement diff display
```bash
bd create "Implement diff display" --description "Show differences between current and updated dependency versions" --type "task" --deps "parent-child:sl-US3-id" --labels "spec:002-spec-dependency-linking,story:US3,component:ui" --priority 2 --design "Display version changes with context, highlight breaking changes" --acceptance "Clear diff output, warns about potential breaking changes"
```

T024 - Write tests for version management
```bash
bd create "Write tests for version management" --description "Create tests for dependency update and version constraint handling" --type "task" --deps "parent-child:sl-US3-id" --labels "spec:002-spec-dependency-linking,story:US3,component:test" --priority 2 --design "Test semantic version parsing, update logic, conflict scenarios" --acceptance "All version management tests pass"
```

### User Story 4 Tasks (P4) - T025 to T28

T025 - Implement conflict detection
```bash
bd create "Implement conflict detection" --description "Detect version conflicts and circular dependencies in dependency graph" --type "task" --deps "parent-child:sl-US4-id" --labels "spec:002-spec-dependency-linking,story:US4,component:resolver" --priority 2 --design "Analyze dependency graph for conflicts, detect cycles, suggest resolutions" --acceptance "Detects all conflicts with 100% accuracy before resolution"
```

T026 - Implement conflict resolution suggestions
```bash
bd create "Implement conflict resolution suggestions" --description "Provide actionable suggestions for resolving dependency conflicts" --type "task" --deps "parent-child:sl-US4-id" --labels "spec:002-spec-dependency-linking,story:US4,component:ui" --priority 2 --design "Suggest update all, use compatible version, create local copy options" --acceptance "Clear suggestions for each conflict type"
```

T027 - Implement conflict command
```bash
bd create "Implement conflict command" --description "Create 'sl conflicts check' to detect and resolve dependency conflicts" --type "task" --deps "parent-child:sl-US4-id" --labels "spec:002-spec-dependency-linking,story:US4,component:cli" --priority 2 --design "Command: sl conflicts check [--resolve], show conflict report" --acceptance "Reports all conflicts with resolution options"
```

T028 - Write tests for conflict detection
```bash
bd create "Write tests for conflict detection" --description "Create tests for conflict detection and resolution scenarios" --type "task" --deps "parent-child:sl-US4-id" --labels "spec:002-spec-dependency-linking,story:US4,component:test" --priority 2 --design "Test various conflict scenarios, circular dependencies, version mismatches" --acceptance "All conflict detection tests pass"
```

### User Story 5 Tasks (P5) - T029 to T32

T029 - Implement vendoring logic
```bash
bd create "Implement vendoring logic" --description "Create logic to copy external dependencies to local vendor directory" --type "task" --deps "parent-child:sl-US5-id" --labels "spec:002-spec-dependency-linking,story:US5,component:vendor" --priority 3 --design "Copy specs to specs/vendor/, preserve directory structure, update spec.sum" --acceptance "Copies all dependencies with preserved structure"
```

T030 - Implement vendor command
```bash
bd create "Implement vendor command" --description "Create 'sl vendor' to vendor external dependencies" --type "task" --deps "parent-child:sl-US5-id" --labels "spec:002-spec-dependency-linking,story:US5,component:cli" --priority 3 --design "Command: sl vendor [--output=specs/vendor], support incremental updates" --acceptance "Vendors dependencies in <60 seconds for 20 specs"
```

T031 - Implement offline reference resolution
```bash
bd create "Implement offline reference resolution" --description "Enable reference resolution using vendored copies when offline" --type "task" --deps "parent-child:sl-US5-id" --labels "spec:002-spec-dependency-linking,story:US5,component:resolver" --priority 3 --design "Check vendor directory first, fallback to cache, warn about stale copies" --acceptance "Works completely offline with vendored dependencies"
```

T032 - Write tests for vendoring
```bash
bd create "Write tests for vendoring" --description "Create tests for dependency vending and offline functionality" --type "task" --deps "parent-child:sl-US5-id" --labels "spec:002-spec-dependency-linking,story:US5,component:test" --priority 3 --design "Test vending process, offline resolution, sync scenarios" --acceptance "All vendoring tests pass, including edge cases"
```

## Parallel Execution Examples

### User Story 1 Parallel Tasks
- T009 (CLI command) and T011 (resolution engine) can run in parallel
- T012 (lockfile) depends on T011
- T013 (verification) depends on T012
- T014 (command) depends on T013

### User Story 2 Parallel Tasks
- T017 (parser) and T018 (resolver) can run in parallel
- T019 (command) depends on both
- T020 (tests) depend on T019

### User Story 3 Parallel Tasks
- T021 (update logic) and T022 (command) can run in parallel
- T023 (diff) depends on T022
- T024 (tests) depend on T023

### User Story 4 Parallel Tasks
- T025 (detection) and T026 (suggestions) can run in parallel
- T027 (command) depends on both
- T028 (tests) depend on T027

### User Story 5 Parallel Tasks
- T029 (vendor logic) and T030 (command) can run in parallel
- T031 (offline) depends on T029
- T032 (tests) depend on T031

## Beads Query Commands

### Filter by Feature
```bash
# All tasks for this feature
bd list --label "spec:002-spec-dependency-linking" -n 32

# Ready tasks to work on
bd ready --label "spec:002-spec-dependency-linking" -n 10

# Tasks by priority
bd list --label "spec:002-spec-dependency-linking" --priority 0 -n 5  # Critical
bd list --label "spec:002-spec-dependency-linking" --priority 1 -n 10 # High
bd list --label "spec:002-spec-dependency-linking" --priority 2 -n 10 # Normal
bd list --label "spec:002-spec-dependency-linking" --priority 3 -n 5  # Low
```

### Filter by Phase
```bash
# Setup phase
bd list --label "phase:setup" -n 3

# User Story 1
bd list --label "phase:us1" -n 8

# User Story 2
bd list --label "phase:us2" -n 4

# User Story 3
bd list --label "phase:us3" -n 4

# User Story 4
bd list --label "phase:us4" -n 4

# User Story 5
bd list --label "phase:us5" -n 4

# Polish phase
bd list --label "phase:polish" -n 3
```

### Filter by Component
```bash
# CLI components
bd list --label "component:cli" -n 8

# Resolver components
bd list --label "component:resolver" -n 8

# Validation components
bd list --label "component:validation" -n 4

# Testing components
bd list --label "component:test" -n 5

# Configuration components
bd list --label "component:config" -n 2
```

### Filter by Story
```bash
# User Story 1 tasks
bd list --label "story:US1" -n 8

# User Story 2 tasks
bd list --label "story:US2" -n 4

# User Story 3 tasks
bd list --label "story:US3" -n 4

# User Story 4 tasks
bd list --label "story:US4" -n 4

# User Story 5 tasks
bd list --label "story:US5" -n 4
```

## MVP and Incremental Delivery

### MVP Scope (User Story 1)
**Deliverable**: Basic dependency declaration and resolution system
**Key Tasks**: T001-T008 (Setup), T009-T016 (US1)
**Success Criteria**:
- Can declare external dependencies with `sl deps add`
- Can resolve dependencies with `sl deps resolve`
- Generates `spec.sum` with cryptographic verification
- Meets SC-001 (<10s) and SC-003 (<30s) performance targets

### Incremental Delivery

**Phase 2** (After MVP): User Story 2 - External References
**Phase 3**: User Story 3 - Version Management
**Phase 4**: User Story 4 - Conflict Detection
**Phase 5**: User Story 5 - Vendoring

## Implementation Strategy

### Critical Path
Setup (T001-T008) → US1 Core (T009-T011) → US1 Completion (T012-T16)

### Dependencies
- All user stories depend on Setup and Foundational phases
- US2 depends on US1 completion
- US3, US4, US5 can be implemented in any order after US1
- Each story is independently testable once prerequisites are met

### Test-First Approach
Each story includes specific test tasks before implementation:
- US1: T016 unit tests before completing core functionality
- US2: T020 integration tests before reference validation
- US3: T024 version management tests
- US4: T028 conflict detection tests
- US5: T032 vendoring tests

## Quality Gates

### Constitution Compliance
- ✅ Specification-First: Complete spec with prioritized user stories
- ✅ Test-First: Test strategy defined for each user story
- ✅ Code Quality: Go linting with golangci-lab, 90%+ coverage
- ✅ UX Consistency: CLI commands follow established patterns
- ✅ Performance: SC-001 to SC-010 targets defined
- ✅ Observability: Structured logging with context
- ✅ Issue Tracking: Beads issue tracking with clear dependencies

### Performance Validation
- SC-001: <10s single dependency resolution (T011)
- SC-002: <5s reference validation (T019)
- SC-003: <30s for 10 repositories (T011)
- SC-004: 100% conflict detection (T025)
- SC-005: <2 minutes dependency updates (T022)
- SC-006: <60 seconds vendoring (T030)
- SC-007: Cryptographic verification (T013)
- SC-008: 90% first success rate (T014)
- SC-009: Authentication handling (T007)
- SC-010: 50 transitive dependencies (T011)

## Task Dependencies Summary

| Task | Depends On | Critical Path |
|------|------------|---------------|
| T001 | None | Setup |
| T002 | T001 | Setup |
| T003 | T001 | Setup |
| T004 | T001, T006 | Foundational |
| T005 | T001, T006 | Foundational |
| T006 | T001, T002, T003 | Foundational |
| T007 | T001, T002, T003 | Foundational |
| T008 | T001, T002, T003, T006 | Foundational |
| T009 | T004, T006, T007, T008 | US1 |
| T010 | T009 | US1 |
| T011 | T004, T006, T007, T008 | US1 |
| T012 | T011 | US1 |
| T013 | T012 | US1 |
| T014 | T013 | US1 |
| T015 | T014 | US1 |
| T016 | T014 | US1 |
| T017 | T004, T005, T008 | US2 |
| T018 | T004, T005, T008 | US2 |
| T019 | T017, T018 | US2 |
| T020 | T019 | US2 |
| ... | ... | ... |

## Notes for Implementation

1. **Interface Design**: Each component should have clear interfaces defined in pkg/
2. **Error Handling**: Use structured errors with context for better debugging
3. **Performance**: Use goroutines for parallel operations where beneficial
4. **Testing**: Mock external dependencies for unit tests
5. **Documentation**: Update inline docs as features are implemented
6. **Backward Compatibility**: Maintain compatibility with existing sl bootstrap script