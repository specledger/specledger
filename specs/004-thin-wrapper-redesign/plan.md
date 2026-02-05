# Implementation Plan: SpecLedger Thin Wrapper Architecture

**Branch**: `004-thin-wrapper-redesign` | **Date**: 2026-02-05 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/004-thin-wrapper-redesign/spec.md`

## Summary

Redesign SpecLedger as a thin wrapper CLI that orchestrates tool installation and project bootstrapping while delegating SDD workflows to user-chosen frameworks (Spec Kit or OpenSpec). This involves removing duplicate SDD commands, implementing prerequisite checking with auto-installation via mise, migrating from .mod to YAML metadata format, and supporting multiple optional SDD frameworks. The redesign focuses SpecLedger on its unique value propositions: bootstrap orchestration, dependency management, and tool neutrality.

## Technical Context

**Language/Version**: Go 1.24+
**Primary Dependencies**:
- cobra (CLI framework) - existing
- viper (configuration) - existing
- gopkg.in/yaml.v3 (YAML parsing)
- os/exec (running mise, bd, etc.)

**Storage**: Local filesystem (YAML metadata at `specledger/specledger.yaml`, dependency cache at `~/.specledger/cache/`)
**Testing**: Go standard testing (`go test`), integration tests with temp directories
**Target Platform**: Cross-platform CLI (Linux, macOS, Windows)
**Project Type**: Single (CLI binary)
**Performance Goals**:
- Bootstrap completes in <3 minutes including tool installation
- `sl doctor` executes in <2 seconds
- Dependency resolution for 95% of repos succeeds on first attempt

**Constraints**:
- Must not duplicate Spec Kit or OpenSpec functionality
- Must maintain backward compatibility with existing .mod files (migration path)
- Must work offline after initial tool installation

**Scale/Scope**:
- Support 2 SDD frameworks initially (Spec Kit, OpenSpec)
- Handle dependency graphs up to 50 external specs
- Single-binary CLI distribution

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

**Status**: No constitution file defined yet (template only). No violations possible.

**Note**: Project should establish constitution principles for:
- CLI design standards (flag naming, output formats)
- Error handling patterns
- Testing requirements
- Breaking change policy

## Project Structure

### Documentation (this feature)

```text
specs/004-thin-wrapper-redesign/
├── plan.md              # This file
├── research.md          # Phase 0 output (technology decisions, migration strategy)
├── data-model.md        # Phase 1 output (YAML schema, metadata structures)
├── quickstart.md        # Phase 1 output (developer guide)
├── contracts/           # Phase 1 output (YAML schema definitions)
└── tasks.md             # Phase 2 output (created by /specledger.tasks)
```

### Source Code (repository root)

```text
pkg/
├── cli/
│   ├── commands/
│   │   ├── bootstrap.go         # sl new, sl init (MODIFY)
│   │   ├── doctor.go            # sl doctor (NEW)
│   │   ├── deps.go              # sl deps (EXISTING - may need YAML updates)
│   │   └── migrate.go           # sl migrate (NEW - .mod to YAML)
│   ├── prerequisites/
│   │   └── checker.go           # Prerequisite validation (NEW)
│   ├── metadata/
│   │   ├── yaml.go              # YAML read/write (NEW)
│   │   ├── schema.go            # Metadata structures (NEW)
│   │   └── migration.go         # .mod to YAML migration (NEW)
│   └── tui/
│       └── sl_new.go            # Bootstrap TUI (MODIFY - remove playbook, add framework selector)
├── embedded/
│   ├── templates/
│   │   ├── .claude/commands/    # MODIFY - remove duplicate SDD commands
│   │   ├── mise.toml            # MODIFY - add framework options with comments
│   │   └── specledger/
│   │       └── specledger.yaml  # NEW - YAML template
│   └── playbooks/               # DELETE - unused playbook feature
└── models/                      # MAY ADD - for YAML schema structs

tests/
├── integration/
│   ├── bootstrap_test.go        # Test sl new workflow
│   ├── doctor_test.go           # Test sl doctor
│   └── migrate_test.go          # Test .mod to YAML migration
└── unit/
    ├── prerequisites_test.go    # Test prerequisite checker
    └── metadata_test.go         # Test YAML parsing/generation
```

**Structure Decision**: Single project structure is maintained since SpecLedger is a standalone CLI tool. The existing `pkg/cli/` structure is extended with new packages for prerequisites checking and metadata management. Embedded templates are cleaned up by removing duplicate commands and unused playbook code.

## Complexity Tracking

**Status**: No constitution violations - this section not needed.

## Phase 0: Research & Decisions

### Research Tasks

1. **YAML Schema Design**
   - Research: Best practices for YAML configuration schemas in Go CLI tools
   - Decision needed: Schema validation approach (struct tags vs external validator)
   - Decision needed: How to handle migration of existing .mod metadata

2. **Prerequisite Checking Patterns**
   - Research: How other CLI tools check for dependencies (kubectl, terraform, etc.)
   - Decision needed: Interactive vs silent mode behavior for missing tools
   - Decision needed: Retry logic for failed installations

3. **Framework Detection Strategy**
   - Research: How to detect which SDD framework a user has installed
   - Decision needed: Should SpecLedger copy framework commands to `.claude/commands/` or rely on PATH?
   - Decision needed: Handle version conflicts between frameworks

4. **Backward Compatibility Strategy**
   - Research: Migration patterns for CLI tools changing configuration formats
   - Decision needed: Automatic migration vs explicit `sl migrate` command
   - Decision needed: Support both formats during transition period?

5. **mise Integration Patterns**
   - Research: How mise backend selection works (ubi vs pipx vs npm)
   - Research: Best practices for mise.toml comments and optional tools
   - Decision needed: Should SpecLedger validate mise.toml syntax?

### Beads Previous Work Query

```bash
# Query related beads issues
bd search "bootstrap" --limit 10
bd search "dependencies" --limit 10
bd search "mise" --limit 5
```

**Note**: Previous work section will be populated after beads query execution.

## Phase 1: Design Artifacts

### Phase 1.1: Data Model

**File**: `data-model.md`

#### Core Entities

1. **Project Metadata** (specledger.yaml)
   - Project identification (name, short_code, version)
   - Framework choice (speckit, openspec, both, none)
   - Dependencies list
   - Creation and modification timestamps

2. **Dependency Entry**
   - Git URL
   - Branch/tag reference
   - File path within repository
   - Alias for referencing
   - Resolved commit hash (lockfile behavior)

3. **Tool Status**
   - Tool name (mise, bd, perles, specify, openspec)
   - Detected installation (boolean)
   - Version string
   - Installation path

### Phase 1.2: Contracts

**Directory**: `contracts/`

#### specledger.yaml Schema

```yaml
# contracts/specledger-schema.yaml
version: "1.0.0"
project:
  name: string          # required
  short_code: string    # required, 2-10 chars
  created: timestamp    # required, ISO8601
  modified: timestamp   # required, ISO8601
  version: string       # required, semver

framework:
  choice: enum          # required: speckit | openspec | both | none
  installed_at: timestamp # optional

dependencies:
  - url: string         # required, git URL
    branch: string      # optional, default: main
    path: string        # optional, default: spec.md
    alias: string       # optional
    resolved_commit: string # optional, SHA hash
```

#### Prerequisites Check Contract

**Input**: None (checks current environment)

**Output** (JSON):
```json
{
  "status": "pass" | "fail",
  "core_tools": {
    "mise": {"installed": true, "version": "v2024.1.0"},
    "bd": {"installed": true, "version": "0.28.0"},
    "perles": {"installed": true, "version": "0.2.11"}
  },
  "frameworks": {
    "speckit": {"installed": false},
    "openspec": {"installed": true, "version": "1.0.0"}
  },
  "missing": ["mise"],
  "install_instructions": "curl https://mise.jdx.dev/install.sh | sh"
}
```

### Phase 1.3: Quickstart Guide

**File**: `quickstart.md`

Contents:
- Developer setup instructions
- How to test the redesign locally
- How to run integration tests
- Migration guide for .mod to YAML
- Testing checklist before PR

## Phase 2: Implementation Phases

### Phase 2.1: Cleanup & Removal (Task Group 1)

**Dependencies**: None

**Deliverables**:
- Remove duplicate SDD commands from `.claude/commands/`
- Remove `specledger.specify`, `specledger.plan`, `specledger.tasks`, `specledger.implement`, `specledger.analyze`, `specledger.clarify`, `specledger.checklist`, `specledger.constitution`
- Keep only: `specledger.deps`, `specledger.adopt`, `specledger.resume`
- Remove unused playbook code from TUI and bootstrap
- Remove `playbookFlag` variable and related logic

**Testing**: Verify no references to removed commands remain in codebase

### Phase 2.2: Metadata System (Task Group 2)

**Dependencies**: Phase 2.1 complete

**Deliverables**:
- Create `pkg/cli/metadata/` package
- Implement YAML schema structures
- Implement read/write functions
- Create migration logic from .mod to YAML
- Add `sl migrate` command
- Update `sl deps` to use YAML format

**Testing**: Unit tests for YAML parsing, migration tests with sample .mod files

### Phase 2.3: Prerequisites Checker (Task Group 3)

**Dependencies**: None (can run parallel to Phase 2.2)

**Deliverables**:
- Create `pkg/cli/prerequisites/checker.go`
- Implement `CheckPrerequisites()` function
- Implement `EnsurePrerequisites(interactive bool)` function
- Helper functions for tool detection
- Create `sl doctor` command

**Testing**: Unit tests for tool detection, integration tests for install prompts

### Phase 2.4: Bootstrap Integration (Task Group 4)

**Dependencies**: Phase 2.2 and 2.3 complete

**Deliverables**:
- Update `bootstrap.go` to call prerequisite checker
- Update TUI to add framework selection (remove playbook)
- Update mise.toml template with framework options and comments
- Update template specledger.yaml file
- Modify `sl init` to use new metadata format

**Testing**: Integration tests for `sl new` and `sl init` workflows

### Phase 2.5: Documentation & Migration (Task Group 5)

**Dependencies**: All previous phases complete

**Deliverables**:
- Update README.md with new architecture
- Create ARCHITECTURE.md design doc
- Update CONTRIBUTING.md if needed
- Create migration guide for existing users
- Update CLI help text

**Testing**: Documentation review, user acceptance testing

## Risks & Mitigation

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Breaking changes for existing users | High | High | Provide automatic migration, support .mod format during transition |
| mise installation failures | Medium | Medium | Clear error messages, fallback instructions for manual installation |
| Framework detection false positives | Low | Medium | Use version checks, not just PATH presence |
| YAML parsing edge cases | Medium | Low | Comprehensive unit tests, schema validation |
| Users confused by framework choice | High | Medium | Clear documentation, sensible defaults (none), `sl doctor` diagnostics |

## Success Metrics

- 100% of duplicate SDD commands removed
- `sl new` bootstrap completes in <3 minutes
- `sl doctor` executes in <2 seconds
- All existing .mod files migrate successfully to YAML
- Zero regression in existing `sl deps` functionality
- CI/CD integration tests pass on Linux, macOS, Windows

## Next Steps

After this plan is approved:
1. Execute Phase 0 research and document findings in `research.md`
2. Complete Phase 1 design artifacts (`data-model.md`, `contracts/`, `quickstart.md`)
3. Run agent context update script
4. Proceed to `/specledger.tasks` for detailed task breakdown
