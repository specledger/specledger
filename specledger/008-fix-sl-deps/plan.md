# Implementation Plan: Fix SpecLedger Dependencies Integration

**Branch**: `008-fix-sl-deps` | **Date**: 2026-02-09 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specledger/008-fix-sl-deps/spec.md`

**Note**: This template is filled in by the `/specledger.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Fix the SpecLedger dependencies integration to properly support artifact path configuration, cross-repository artifact references, and complete Claude Code integration. The system will:
1. Add `artifact_path` field to `specledger.yaml` for current project metadata
2. Auto-discover artifact paths from SpecLedger repository dependencies
3. Support manual `--artifact-path` flag for non-SpecLedger repos
4. Complete the `sl deps resolve` command to download dependencies to cache
5. Create all 5 Claude Code command files for deps operations
6. Update skill documentation with comprehensive workflow guidance

## Technical Context

**Language/Version**: Go 1.24+
**Primary Dependencies**: Cobra (CLI), go-git v5 (Git operations), YAML v3 (config parsing)
**Storage**: File-based (specledger.yaml for metadata, ~/.specledger/cache/ for dependencies)
**Testing**: Go testing (unit tests in pkg/, integration tests in tests/)
**Target Platform**: macOS, Linux (CLI tool)
**Project Type**: Single project (CLI tool with internal packages)
**Performance Goals**:
- Dependency resolution: <5 seconds per dependency for typical repos
- Cache operations: <1 second for already-cached repos
- CLI commands: <500ms response time for non-network operations
**Constraints**:
- Must support offline operation after dependencies are cached
- Must handle network errors gracefully during resolve operations
- Must maintain backward compatibility with existing specledger.yaml files
**Scale/Scope**:
- Typical projects: 5-20 dependencies
- Cache storage: ~100MB per dependency (varies by repo size)
- CLI user base: Developers using SpecLedger for specification management

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Verify compliance with principles from `.specledger/memory/constitution.md`:

- [x] **Specification-First**: Spec.md complete with prioritized user stories (6 user stories, P1-P2)
- [x] **Test-First**: Test strategy defined (unit tests in pkg/, integration tests for deps operations)
- [x] **Code Quality**: Linting/formatting tools identified (gofmt, golangci-lint in Makefile)
- [x] **UX Consistency**: User flows documented in spec.md acceptance scenarios (18 scenarios across 6 stories)
- [x] **Performance**: Metrics defined in Technical Context (<5s per dependency, <500ms CLI response)
- [x] **Observability**: Logging/metrics strategy documented (CLI output, cache status, error messages)
- [ ] **Issue Tracking**: Beads epic created and linked to spec (NEEDS CLARIFICATION - use existing SL-26y or create new?)

**Complexity Violations** (if any, justify in Complexity Tracking table below):
- None identified

## Project Structure

### Documentation (this feature)

```text
specledger/008-fix-sl-deps/
├── plan.md              # This file (/specledger.plan command output)
├── research.md          # Phase 0 output (/specledger.plan command)
├── data-model.md        # Phase 1 output (/specledger.plan command)
├── quickstart.md        # Phase 1 output (/specledger.plan command)
├── contracts/           # Phase 1 output (/specledger.plan command)
└── tasks.md             # Phase 2 output (/specledger.tasks command - NOT created by /specledger.plan)
```

### Source Code (repository root)

```text
# SpecLedger CLI structure (single project)
cmd/sl/
└── main.go              # Entry point

pkg/
├── cli/
│   ├── commands/
│   │   ├── deps.go              # DEPS: Complete resolve, add --artifact-path flag
│   │   ├── init.go              # DEPS: Add artifact_path detection
│   │   └── new.go               # DEPS: Add artifact_path initialization
│   └── metadata/
│       ├── schema.go            # DEPS: Add ArtifactPath to ProjectMetadata
│       ├── yaml.go              # Already handles DefaultMetadataFile
│       └── loader.go            # Load/save specledger.yaml
├── models/
│   └── dependency.go            # DEPS: Reconcile with metadata schema
└── deps/
    ├── resolver.go              # NEW: Artifact path discovery
    ├── cache.go                 # NEW: Cache operations for ~/.specledger/cache/
    └── reference.go             # NEW: Reference resolution logic

internal/
└── spec/
    └── resolver.go              # EXISTS: Go git operations, reuse for deps

tests/
├── integration/
│   └── deps_test.go             # DEPS: Integration tests for deps operations
└── unit/
    └── deps_test.go             # DEPS: Unit tests for new functions

.claude/
├── commands/
│   ├── specledger.add-deps.md    # EXISTS: Review and update
│   ├── specledger.remove-deps.md # EXISTS: Review and update
│   ├── specledger.list-deps.md   # NEW: Create
│   ├── specledger.resolve-deps.md # NEW: Create
│   └── specledger.update-deps.md  # NEW: Create
└── skills/
    └── specledger-deps/
        └── SKILL.md              # EXISTS: Update with comprehensive documentation
```

**Structure Decision**: Single CLI project with internal packages. Key modifications:
1. `pkg/cli/commands/deps.go` - Complete implementation
2. `pkg/cli/metadata/schema.go` - Add ArtifactPath field
3. `pkg/deps/` - New package for artifact path resolution
4. `.claude/commands/` - Create missing command files

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| N/A | | |

## Previous Work

### Related Features and Tasks

#### SL-26y: Release Delivery Fix
- **SL-m82**: Fix install script architecture detection
- **SL-vhp**: Simplify GoReleaser builds
- **Status**: Completed, released as v1.0.3, v1.0.4, v1.0.5
- **Relevance**: Release automation but not directly related to dependencies

#### Existing Dependencies Implementation
- **File**: `pkg/cli/commands/deps.go`
- **Status**: Partially implemented (add, list, remove work; resolve and update incomplete)
- **Gaps Identified**:
  1. No `artifact_path` field in `specledger.yaml`
  2. `sl deps resolve` doesn't download to cache properly
  3. Missing Claude Code command files for list, resolve, update
  4. No auto-discovery of artifact paths from SpecLedger repos

#### Existing Git Operations
- **File**: `internal/spec/resolver.go`
- **Status**: Implements go-git operations for spec resolution
- **Relevance**: Can be reused/adapted for deps resolve functionality

#### Data Model Inconsistencies
- **File**: `pkg/models/dependency.go` vs `pkg/cli/metadata/schema.go`
- **Issue**: Different structures for Dependency entity
- **Action Required**: Reconcile schemas and add ArtifactPath field

### Beads References

No specific beads identified yet. The feature spec references generic epic SL-26y but this may need a separate epic for the dependencies work.

**Decision**: Create new epic SL-deps for dependencies work, distinct from SL-26y (Release Delivery Fix).

---

## Phase 0: Research Summary

**Status**: ✅ Complete

**Output**: [research.md](./research.md)

**Key Decisions**:
1. Add `ArtifactPath` field to both `ProjectMetadata` and `Dependency` structs
2. Use `pkg/cli/metadata/schema.go` as primary data structure (models.Dependency may be deprecated)
3. Migrate deps.go to use go-git/v5 (currently uses system git command)
4. Create `pkg/deps/` package for artifact path discovery and reference resolution
5. Default `artifact_path` to `"specledger/"` for backward compatibility

**Clarifications Resolved**:
- Cache location confirmed: `~/.specledger/cache/` (global, not project-local)
- Two command files already exist: `list-deps.md` and `resolve-deps.md`
- Missing: `update-deps.md` command file needs creation
- Reference resolution to be implemented in Phase 2 (deferred from Phase 1)

---

## Phase 1: Design Summary

**Status**: ✅ Complete

**Outputs**:
- [data-model.md](./data-model.md) - Entity definitions, state diagrams, validation rules
- [contracts/cli-api.md](./contracts/cli-api.md) - Command signatures, flags, error handling
- [quickstart.md](./quickstart.md) - User guide with examples and troubleshooting

**Design Decisions**:
1. **Reference Format**: `<alias>:<artifact-name>` for cross-repo references
2. **Resolution Formula**: `project.artifact_path + dependency.path + "/" + artifact_name`
3. **Auto-Discovery**: Read dependency's specledger.yaml to find artifact_path for SpecLedger repos
4. **Manual Flag**: `--artifact-path` flag for non-SpecLedger repos
5. **Backward Compatibility**: `omitempty` YAML tags, default values for missing fields

---

## Implementation Roadmap

### Priority 1 (Mandatory)

| Task | File | Description |
|------|------|-------------|
| Add ArtifactPath | `pkg/cli/metadata/schema.go` | Add field to ProjectMetadata and Dependency |
| Update sl new | `pkg/cli/commands/new.go` | Set default artifact_path |
| Update sl init | `pkg/cli/commands/init.go` | Detect and configure artifact_path |
| Add --artifact-path flag | `pkg/cli/commands/deps.go` | New flag for add command |
| Implement discovery | `pkg/deps/resolver.go` | Auto-detect artifact_path from SpecLedger repos |
| Complete resolve | `pkg/cli/commands/deps.go` | Use go-git, handle errors properly |
| Create update-deps.md | `.claude/commands/` | New command file |
| Update resolve-deps.md | `.claude/commands/` | Fix cache location documentation |
| Update skill docs | `.claude/skills/specledger-deps/` | Add artifact_path documentation |

### Priority 2 (Important)

| Task | File | Description |
|------|------|-------------|
| Implement update | `pkg/cli/commands/deps.go` | Complete the stub implementation |
| Reference resolution | `pkg/deps/reference.go` | Resolve alias:artifact references |
| Validation helpers | `pkg/cli/metadata/validator.go` | Validate artifact_path values |
| Unit tests | `pkg/**/*_test.go` | Test new functionality |
| Integration tests | `tests/integration/deps_test.go` | Test full workflow |

---

## Constitution Check (Post-Design)

*Re-evaluated after Phase 1 design*

- [x] **Specification-First**: Spec.md complete with prioritized user stories (6 user stories, P1-P2)
- [x] **Test-First**: Test strategy defined (unit and integration tests planned in roadmap)
- [x] **Code Quality**: Linting/formatting tools identified (gofmt, golangci-lint in Makefile)
- [x] **UX Consistency**: User flows documented in spec.md acceptance scenarios (18 scenarios)
- [x] **Performance**: Metrics defined (<5s per dependency, <500ms CLI response)
- [x] **Observability**: Logging/metrics strategy documented (CLI output, cache status)
- [x] **Issue Tracking**: New epic SL-deps to be created for dependencies work

**No complexity violations identified.**

---

## Generated Artifacts

| Artifact | Path | Description |
|----------|------|-------------|
| Plan | `plan.md` | This file |
| Research | `research.md` | Research findings and decisions |
| Data Model | `data-model.md` | Entity definitions and state diagrams |
| Contracts | `contracts/cli-api.md` | CLI API specifications |
| Quickstart | `quickstart.md` | User guide and examples |
| Tasks | `tasks.md` | **NOT YET GENERATED** - Run `/specledger.tasks` |

---

## Next Steps

1. **Review this plan** with stakeholders to confirm approach
2. **Run `/specledger.tasks`** to generate the task breakdown (beads)
3. **Run `/specledger.implement`** to execute the implementation (or implement manually)

---

**Plan Status**: ✅ Phase 0 Complete | ✅ Phase 1 Complete | ⏳ Phase 2 Pending (tasks.md)

**Last Updated**: 2026-02-09

