# Implementation Plan: Project Templates & Coding Agent Selection

**Branch**: `592-init-project-templates` | **Date**: 2026-02-19 | **Spec**: [spec.md](../specs/592-init-project-templates/spec.md)
**Input**: Feature specification from `/specs/592-init-project-templates/spec.md`

## Summary

Enable developers to select from 7 business-defined project templates (General Purpose, Full-Stack, Batch Data, Real-Time Workflow, ML Image Processing, Real-Time Data Pipeline, AI Chatbot) and choose their preferred coding agent (Claude Code, OpenCode, None) during interactive `sl new`. The system generates a unique project ID (UUID) for each project to enable session storage and tracking in Supabase.

**Technical Approach**:
1. Extend embedded template system from feature 005 to support 7 distinct project types
2. Enhance TUI (from feature 011) to add template selection, agent selection, optional preview, and recommendations
3. Generate UUID v4 project IDs during creation for Supabase session storage namespacing
4. Implement keyword-based template recommendations (MVP) with upgrade path to embedding-based (production)
5. Support template parameterization for customizing service names, ports, module names
6. Provide CLI flags for non-interactive usage (`--template`, `--agent`, `--param`, `--preview-template`, `--recommend`)

---

## Technical Context

**Language/Version**: Go 1.24+

**Primary Dependencies**:
- `github.com/spf13/cobra` - CLI framework
- `github.com/charmbracelet/bubbletea v1.3.10` - TUI framework
- `github.com/charmbracelet/bubbles v0.21.1` - TUI components (list, viewport, textinput)
- `github.com/charmbracelet/lipgloss v1.1.0` - Terminal styling
- `github.com/google/uuid v1.6.0` - UUID generation (NEW)
- `gopkg.in/yaml.v3` - YAML parsing
- `encoding/json` - JSON formatting for .claude/settings.json (stdlib)
- Existing: `specledger/pkg/cli/*` packages

**Storage**: File system (templates embedded in binary via `embed.FS`, project metadata in `specledger.yaml`)

**Testing**: Go testing (`go test`), integration tests with temporary directories, TUI tests with BubbleTea test harness

**Target Platform**: Cross-platform (Linux, macOS, Windows) - same as SpecLedger CLI

**Project Type**: CLI tool with embedded resources and interactive TUI

**Performance Goals**:
- Template selection TUI flow: <60 seconds end-to-end (SC-001)
- Template preview display: <2 seconds (SC-007)
- Template recommendation: <1 second for keyword-based (MVP)
- UUID generation: <1ms per ID
- Parameter customization: Within same 60-second flow (SC-010)

**Constraints**:
- Must preserve existing `sl new` behavior when using General Purpose + Claude Code (backward compatibility)
- Must handle 7+ templates without overwhelming users (recommendations help)
- TUI must work on terminals with varying sizes (responsive layout)
- Offline-capable: No external API calls required for core functionality
- Non-interactive mode must support all features via CLI flags

**Scale/Scope**:
- 7 project templates (expandable to 15+ with architecture)
- 3 coding agent options
- 3-5 parameters per parameterized template
- Expected template files per type: 20-150 files
- Total embedded size estimate: ~10-20MB

---

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

**Note**: No active constitution file exists for SpecLedger project. Using general software engineering best practices as guidelines:

| Principle | Status | Notes |
|-----------|--------|-------|
| **Specification-First** | ✅ PASS | Spec.md complete with 8 prioritized user stories, 22 functional requirements |
| **Test-First** | ✅ PASS | Test strategy defined: integration tests for TUI flows, template copying, UUID generation |
| **Code Quality** | ✅ PASS | Go 1.24+ conventions, gofmt, golangci-lint |
| **UX Consistency** | ✅ PASS | Extends existing TUI patterns from feature 011, maintains gold #13 theme |
| **Performance** | ✅ PASS | Metrics defined: <60s flow, <2s preview, <1s recommendations |
| **Observability** | ✅ PASS | Structured logging for template operations, metadata tracking in specledger.yaml |
| **Simplicity** | ✅ PASS | Reuses Bubbles components, keyword-based recommendations (MVP), no over-engineering |
| **Future-Proof** | ✅ PASS | Architecture supports: more templates, embedding-based recommendations, remote templates |

**Complexity Violations**: None identified - proceeding to Phase 0.

---

## Project Structure

### Documentation (this feature)

```text
specs/592-init-project-templates/
├── spec.md                    # Feature specification (8 user stories, 22 FRs)
├── plan.md                    # This file (implementation plan)
├── research.md                # Phase 0 output (TUI, UUID, recommendations research)
├── data-model.md              # Phase 1 output (data structures)
├── quickstart.md              # Phase 1 output (developer guide)
├── contracts/                 # Phase 1 output (Go interfaces)
│   └── playbooks.go           # PlaybookSource interface extensions
├── RESEARCH.md                # Industry research (6 general templates)
├── RESEARCH-CHATBOT.md        # Industry research (chatbot template)
├── RESEARCH-TUI-PATTERNS.md   # TUI architecture (1,150 lines)
├── IMPLEMENTATION-GUIDE.md    # TUI code patterns (1,250 lines)
├── TUI-QUICK-REFERENCE.md     # Quick lookup (550 lines)
├── RESEARCH-RECOMMENDATIONS.md # Recommendation algorithms (804 lines)
├── RECOMMENDATIONS-SUMMARY.md  # Decision guide (309 lines)
├── KEYWORD-PROFILES.yaml      # Keyword config for all 7 templates
├── INDEX-RECOMMENDATIONS.md   # Navigation for recommendation research
└── checklists/
    └── requirements.md        # Specification quality validation
```

### Source Code (repository root)

```text
specledger/
├── cmd/
│   └── main.go                    # CLI entry point (no changes)
│
├── pkg/
│   ├── cli/
│   │   ├── commands/              # CLI commands
│   │   │   ├── bootstrap.go       # Modified: Add UUID generation
│   │   │   ├── bootstrap_helpers.go # Modified: Template+agent application
│   │   │   └── playbooks.go       # Modified: Add preview, recommend flags
│   │   │
│   │   ├── playbooks/             # Template management (feature 005 foundation)
│   │   │   ├── template.go        # Modified: Add template metadata fields
│   │   │   ├── templates.go       # Modified: Add preview, recommend methods
│   │   │   ├── manifest.go        # Modified: Parse template characteristics
│   │   │   ├── copy.go            # Modified: Parameter substitution
│   │   │   ├── embedded.go        # Unchanged: Embedded template loading
│   │   │   ├── remote.go          # Unchanged: Stub for future
│   │   │   ├── cache.go           # Unchanged: Caching infrastructure
│   │   │   ├── path.go            # Unchanged: Path utilities
│   │   │   ├── recommend.go       # NEW: Template recommendation engine
│   │   │   └── parameters.go      # NEW: Parameter collection & substitution
│   │   │
│   │   ├── metadata/              # Project metadata system
│   │   │   ├── schema.go          # Modified: Add ID, Template, Agent, Parameters fields
│   │   │   ├── yaml.go            # Modified: Generate UUID in NewProjectMetadata
│   │   │   └── migration.go       # Modified: Handle migration for projects without UUID
│   │   │
│   │   └── tui/                   # Terminal UI
│   │       └── sl_new.go          # Modified: Add template/agent/param steps
│   │
│   └── embedded/                  # Embedded resources
│       └── embedded.go            # Modified: Embed new template directories
│
├── pkg/embedded/templates/        # Template source (embedded in binary)
│   ├── manifest.yaml              # Modified: Define 7 templates with metadata
│   ├── general-purpose/           # NEW: Default template (current behavior)
│   ├── full-stack/                # NEW: Go backend + React frontend
│   ├── batch-data-processing/     # NEW: Airflow-style pipelines
│   ├── real-time-workflow/        # NEW: Temporal-style workflows
│   ├── ml-image-processing/       # NEW: MLOps structure
│   ├── real-time-data-pipeline/   # NEW: Kafka streaming
│   └── ai-chatbot/                # NEW: Multi-platform chatbot
│
└── tests/
    └── integration/
        ├── templates_test.go      # Modified: Test all 7 templates
        ├── tui_test.go            # NEW: TUI flow tests
        ├── uuid_test.go           # NEW: UUID generation tests
        └── recommend_test.go      # NEW: Recommendation algorithm tests
```

**Structure Decision**: Single project structure (CLI tool). This feature extends the existing SpecLedger CLI with:
1. **TUI enhancements** in `pkg/cli/tui/sl_new.go` - Add 4 new steps (template, agent, params, recommendations)
2. **Playbook extensions** in `pkg/cli/playbooks/` - Add recommendation engine and parameter handling
3. **Metadata extensions** in `pkg/cli/metadata/` - Add UUID, template, agent tracking
4. **7 new template directories** in `pkg/embedded/templates/` - Each with constitution, directory structure, placeholder files

---

## Complexity Tracking

> No constitution violations - this section intentionally left empty.

---

## Phase 0: Outline & Research (COMPLETE)

### Research Tasks Completed

**Task 1: TUI Template Selection Patterns**
- Researched Bubble Tea list component, viewport for preview, split layout patterns
- Evaluated state machine extension for multi-step flows (template → agent → params → confirm)
- Analyzed recommendation UI patterns (showing 2-3 suggestions with reasoning)
- Evaluated parameter collection approaches (textinput vs huh forms)
- **Output**: 3 comprehensive docs (3,950 lines total)
  - `RESEARCH-TUI-PATTERNS.md` - Architecture guide (1,150 lines)
  - `IMPLEMENTATION-GUIDE.md` - Code patterns (1,250 lines)
  - `TUI-QUICK-REFERENCE.md` - Quick lookup (550 lines)

**Task 2: UUID Generation**
- Researched UUID libraries (google/uuid recommended over gofrs/uuid, satori/go.uuid)
- Evaluated UUID v4 generation patterns, YAML serialization, validation
- Analyzed migration strategy for existing projects without UUIDs
- **Output**: Comprehensive research in `research.md` Decision 2

**Task 3: Template Recommendation Algorithm**
- Evaluated 4 options: keyword-based, embedding-based, LLM-based, simple regex
- **Decision**: Two-phase approach
  - Phase 1 (MVP): Keyword-based matching (40-60% accuracy, <1ms, $0 cost)
  - Phase 2 (Production): Embedding-based (75-85% accuracy, 100-500ms, upgrade when needed)
- Created keyword profiles for all 7 templates with weighted terms
- **Output**: 3 comprehensive docs (1,485 lines total)
  - `RESEARCH-RECOMMENDATIONS.md` - Technical analysis (804 lines)
  - `RECOMMENDATIONS-SUMMARY.md` - Decision guide (309 lines)
  - `KEYWORD-PROFILES.yaml` - Ready-to-use config (372 lines)

### Prior Work Summary

**Feature 005: Embedded Templates** (Foundation)
- Established `pkg/cli/playbooks/` package architecture
- Implemented manifest-based template discovery
- Created file copying with pattern matching
- Embedded filesystem with `//go:embed`

**Feature 011: Streamline Onboarding** (TUI Foundation)
- Implemented 5-step TUI flow with Bubble Tea
- State machine pattern in `sl_new.go`
- Text input and radio selection components
- Gold #13 theme styling

**Feature 004: Thin Wrapper Redesign** (Metadata Foundation)
- Created `pkg/cli/metadata/` with ProjectMetadata struct
- Defined `specledger.yaml` schema
- Established CLI command structure

---

## Phase 1: Design & Contracts (COMPLETE - From Previous Run)

**Prerequisites**: Research complete ✅

### Artifacts Created

**1. data-model.md** (28,009 bytes)
- Entity definitions: ProjectTemplate, CodingAgentConfiguration, ProjectID, TemplateParameter, TemplateRecommendation
- Extended ProjectMetadata struct with ID, Template, Agent, Parameters fields
- Validation rules and state transitions
- Field definitions and relationships

**2. contracts/** (Go interfaces)
- PlaybookSource interface extensions for preview and recommendations
- TemplateRecommender interface
- ParameterCollector interface

**3. quickstart.md** (20,075 bytes)
- Developer setup guide
- Example workflows for each user story
- Testing instructions
- Common patterns and troubleshooting

### Claude Settings File Structure (FR-023)

When Claude Code is selected as the coding agent, the system creates `.claude/settings.json` with session capture configuration:

```json
{
  "saveTranscripts": true,
  "transcriptsDirectory": "~/.claude/sessions",
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "command": "sl session capture"
          }
        ]
      }
    ]
  }
}
```

---

## Phase 2: Task Breakdown (Next Step)

**Prerequisites**: Phase 1 complete ✅

### Execution

Run `/specledger.tasks` command to generate `tasks.md` with:
- Dependency-ordered implementation tasks
- Effort estimates per task
- Phase breakdown (MVP → Enhancements)
- Acceptance criteria per task

**Expected Phases**:
1. **Foundation** (3-4 days): UUID generation, metadata extensions, manifest updates, .claude/settings.json creation
2. **TUI Basic Selection** (2-3 days): Template list, agent list, state machine integration
3. **Template Creation** (4-5 days): Create 7 template directories with structures
4. **Preview & Recommendations** (2-3 days): Viewport preview, keyword-based recommendations
5. **Parameterization** (2-3 days): Parameter collection, file substitution
6. **CLI Flags & Non-Interactive** (2 days): Flag parsing, non-interactive mode
7. **Testing & Polish** (3-4 days): Integration tests, documentation, edge cases

**Total Estimate**: 18-25 days

---

## Implementation Roadmap

### Phase 1: Foundation & MVP (Week 1-2)
- UUID generation in metadata
- Extend ProjectMetadata schema
- Create manifest.yaml with 7 template definitions
- Update TUI state machine (add template and agent steps)
- Basic template list with bubbles/list component
- Basic agent selection
- Template file copying for all 7 types
- Create `.claude/settings.json` with project ID when Claude Code is selected (FR-023)

### Phase 2: Rich Display & Preview (Week 2-3)
- Viewport-based preview pane
- Split layout (list on left, preview on right)
- Directory tree visualization
- Toggle preview with Tab key
- Lazy load preview content

### Phase 3: Recommendations & Parameters (Week 3-4)
- Keyword-based recommendation engine
- KEYWORD-PROFILES.yaml integration
- Recommendation display in TUI (optional step)
- Parameter collection (textinput steps)
- Parameter substitution in template files
- Store parameters in specledger.yaml

### Phase 4: CLI & Non-Interactive (Week 4)
- `--template <name>` flag
- `--agent <name>` flag
- `--param <key>=<value>` flags
- `--preview-template <name>` command
- `--recommend "<description>"` command
- `--list-templates` command
- Non-interactive mode validation

### Phase 5: Testing & Polish (Week 4-5)
- Integration tests for all 7 templates
- TUI flow tests
- UUID generation tests
- Recommendation algorithm tests
- Documentation updates
- Edge case handling

---

## Success Criteria Validation

From spec.md Success Criteria section:

- **SC-001**: <60 seconds TUI flow
  - **Validation**: Time each step in integration tests, sum must be <60s

- **SC-002**: All 7 templates produce valid, non-empty structures
  - **Validation**: Test each template creation, verify key directories exist

- **SC-003**: 100% backward compatibility with General Purpose + Claude Code
  - **Validation**: Compare output with current `sl new` behavior byte-for-byte

- **SC-004**: Non-interactive mode completes without prompts
  - **Validation**: Test `sl new --template X --agent Y` in CI environment

- **SC-004a**: Every project has unique UUID
  - **Validation**: Create 1000 projects, verify no duplicate IDs

- **SC-005**: Template structures are distinct and recognizable
  - **Validation**: Manual review + automated checks for key directories per template

- **SC-006**: Agent selection produces different configuration files
  - **Validation**: Compare `.claude/` vs OpenCode config vs none for same template

- **SC-007**: Preview displays in <2 seconds
  - **Validation**: Time viewport rendering for each template

- **SC-008**: Recommendations achieve ≥80% relevance
  - **Validation**: Test 20+ descriptions, subjective evaluation by developers

- **SC-009**: ≥3 key parameters customizable
  - **Validation**: Count parameters for each parameterized template

- **SC-010**: Parameter customization within 60-second flow
  - **Validation**: Time complete flow including parameter steps

---

## Dependencies & External References

**Go Module Dependencies** (to add):
```bash
go get github.com/google/uuid@v1.6.0
```

**Existing Dependencies** (already in go.mod):
- charmbracelet/bubbletea v1.3.10
- charmbracelet/bubbles v0.21.1
- charmbracelet/lipgloss v1.1.0
- spf13/cobra
- gopkg.in/yaml.v3

**Optional Phase 2 Dependencies**:
- `charmbracelet/huh` - Form builder for enhanced parameter collection

**External References** (no sl deps add needed):
- Research documents are internal
- No external API specifications required
- All templates are self-contained

---

## Risk Register

| Risk | Probability | Impact | Mitigation |
|------|------------|---------|------------|
| TUI complexity overwhelms users | Medium | Medium | Start with simple list, add preview in Phase 2 |
| Recommendation accuracy <40% | Low | Medium | Keyword profiles are well-researched, upgrade to embeddings if needed |
| Parameter collection UX poor | Medium | Low | Make parameters optional with sensible defaults |
| Template maintenance overhead | High | Medium | Clear documentation, consistent structure across templates |
| UUID collision | Very Low | High | Use google/uuid (industry standard), collision practically impossible |
| Non-interactive mode gaps | Low | Medium | Comprehensive flag coverage from day 1 |

---

## Next Steps

1. ✅ Phase 0: Research complete (TUI patterns, UUID, recommendations)
2. ✅ Phase 1: Design complete (data-model.md, contracts/, quickstart.md)
3. **Next**: Run `/specledger.tasks` to generate implementation task breakdown
4. Begin Phase 1 implementation (Foundation & MVP)
5. Iterate through phases with continuous testing

---

**Status**: Planning complete, ready for task generation and implementation.
