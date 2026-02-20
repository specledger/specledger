# Research: Project Templates & Coding Agent Selection

**Feature**: 592-init-project-templates
**Date**: 2026-02-18
**Status**: Complete

This document consolidates research findings for implementing template and agent selection in `sl new`.

---

## Prior Work

### Feature 011-streamline-onboarding (TUI Implementation)

**Location**: `pkg/cli/tui/sl_new.go`

**Key Insights**:
- Implemented 5-step TUI flow using Bubble Tea framework (Bubbles, Lipgloss)
- Pattern: `Model` struct with `step` field, `Update()` for events, `View()` for rendering
- Text input steps: project name, directory, short code
- Radio selection step: playbook choice with `›` cursor + `◉`/`○` buttons
- Confirmation review step: displays all selections before creating project
- Navigation: Enter (next), Ctrl+C (quit), ↑/↓ (selection)
- Styling: Lipgloss colors (gold primary, green success, red error)

**Reusable Components**:
- `textInput` from Bubbles (`github.com/charmbracelet/bubbles/textinput`)
- Radio selection pattern (selection index + arrow key navigation)
- Step-based state machine (stepProjectName, stepDirectory, etc.)
- Answers map (`map[string]string`) for collecting user input across steps

**Integration Point**: Add two new steps (stepTemplate, stepAgent) after stepShortCode and before stepPlaybook.

### Feature 005-embedded-templates (Template System)

**Location**: `pkg/cli/playbooks/`, `pkg/embedded/templates/`

**Key Insights**:
- Well-architected playbook system with `PlaybookSource` interface
- `EmbeddedSource` loads templates from embedded filesystem (`embed.FS`)
- Manifest-driven: YAML manifest (`manifest.yaml`) defines template patterns
- Copy strategy: Pattern matching with glob patterns, skip existing by default
- Templates compiled into binary via Go's `embed` package

**Current Manifest Structure**:
```yaml
playbooks:
  - name: specledger
    description: "SpecLedger playbook..."
    version: "1.0.0"
    path: "specledger"
    patterns: ["**"]
```

**Extension Path**: Add template definitions to manifest with directory structures per template type.

**Files**:
- `pkg/cli/playbooks/template.go` - Core interfaces (PlaybookSource, Playbook)
- `pkg/cli/playbooks/embedded.go` - EmbeddedSource implementation
- `pkg/cli/playbooks/copy.go` - File copying logic with pattern matching
- `pkg/embedded/templates/manifest.yaml` - Current manifest

### Feature 004-thin-wrapper-redesign (Metadata System)

**Location**: `pkg/cli/metadata/`

**Current Schema** (`schema.go`):
```go
type ProjectMetadata struct {
    Version      string           `yaml:"version"`
    Project      ProjectInfo      `yaml:"project"`
    Framework    FrameworkInfo    `yaml:"framework"`
    TaskTracker  TaskTrackerInfo  `yaml:"task_tracker,omitempty"`
    Dependencies []Dependency     `yaml:"dependencies,omitempty"`
}

type ProjectInfo struct {
    Name      string    `yaml:"name"`
    ShortCode string    `yaml:"short_code"`
    Created   time.Time `yaml:"created"`
    Modified  time.Time `yaml:"modified"`
    Version   string    `yaml:"version"`
}
```

**YAML Library**: `gopkg.in/yaml.v3` (v3.0.1)
- Standard marshaling/unmarshaling via `yaml.Marshal()` and `yaml.Unmarshal()`
- No custom `MarshalYAML`/`UnmarshalYAML` implementations
- Validation via `Validate()` method after unmarshal, before marshal
- Auto-update of `Modified` timestamp before saving

**Extension Required**: Add `ID`, `Template`, `Agent` fields to `ProjectInfo` struct.

---

## Research Findings

### 1. UUID Generation for Project ID (FR-007a)

**Decision**: Use `github.com/google/uuid` package

**Rationale**:
- Industry standard Go UUID library maintained by Google
- Based on RFC 9562 and DCE 1.1 standards
- Uses `crypto/rand` for cryptographically secure randomness
- Built-in YAML marshaling via `MarshalText()`/`UnmarshalText()` interfaces
- Zero additional code needed for YAML integration

**Implementation**:
```go
import "github.com/google/uuid"

// Generate UUID v4
projectID, err := uuid.NewRandom()
if err != nil {
    return fmt.Errorf("failed to generate UUID: %w", err)
}

// Or simpler version (panics on error)
projectID := uuid.New()
```

**YAML Format**:
```yaml
project:
    id: 550e8400-e29b-41d4-a716-446655440000
```

**Schema Extension**:
```go
type ProjectInfo struct {
    ID        uuid.UUID `yaml:"id"`           // NEW
    Name      string    `yaml:"name"`
    ShortCode string    `yaml:"short_code"`
    Template  string    `yaml:"template,omitempty"`  // NEW
    Agent     string    `yaml:"agent,omitempty"`     // NEW
    Created   time.Time `yaml:"created"`
    Modified  time.Time `yaml:"modified"`
    Version   string    `yaml:"version"`
}
```

**Alternatives Considered**:
- `github.com/satori/go.uuid`: Older, less actively maintained
- Manual UUID generation: Reinventing the wheel, potential security issues

**Installation**: `go get github.com/google/uuid`

### 2. Bubble Tea TUI Testing Strategy

**Decision**: Three-layer testing pyramid with `github.com/charmbracelet/x/exp/teatest`

**Rationale**:
- Teatest is the official testing library from Charm Bracelet
- Wraps `tea.Program` in controlled environment (headless terminal)
- Supports key simulation, output verification, golden file testing
- Direct model testing (bypassing program) provides fast unit tests
- Combination gives comprehensive coverage with fast feedback loop

**Testing Layers**:

1. **Unit Tests** (Fast, Focused)
   - Direct testing of `Update()` and `View()` methods
   - No program overhead, pure function testing
   - Table-driven tests for state transitions
   ```go
   func TestUpdateProjectName(t *testing.T) {
       m := InitialModel("/tmp")
       updatedModel, _ := m.Update(tea.KeyMsg{
           Type: tea.KeyRunes,
           Runes: []rune("test-project"),
       })
       model := updatedModel.(Model)
       assert.Equal(t, "test-project", model.textInput.Value())
   }
   ```

2. **Integration Tests** (Medium Speed, Key Bindings)
   - Teatest with `Send()` and `Type()` for key simulation
   - Verify navigation and state transitions
   ```go
   func TestPlaybookNavigation(t *testing.T) {
       tm := teatest.NewTestModel(t, InitialModel("/tmp"),
           teatest.WithInitialTermSize(80, 24))
       tm.Send(tea.KeyMsg{Type: tea.KeyDown})
       teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
           return bytes.Contains(bts, []byte("Select Playbook"))
       })
   }
   ```

3. **E2E Tests** (Slow, Comprehensive, Golden Files)
   - Full flow testing with golden file output comparison
   - Regression protection for UI rendering
   ```go
   func TestFullBootstrapFlow(t *testing.T) {
       tm := teatest.NewTestModel(t, m)
       tm.Type("my-project")
       tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
       // ... complete flow
       out, _ := io.ReadAll(tm.FinalOutput(t))
       teatest.RequireEqualOutput(t, out)
   }
   ```

**Critical Best Practices**:
- Set `lipgloss.SetColorProfile(termenv.Ascii)` in `init()` to prevent CI failures
- Add `*.golden -text` to `.gitattributes`
- Always specify terminal size: `teatest.WithInitialTermSize(80, 24)`
- Set timeouts: `teatest.WithFinalTimeout(time.Second*2)`

**Integration with Existing Tests**:
- Current integration tests use `--ci` mode to bypass TUI
- Add new `bootstrap_tui_test.go` alongside existing `bootstrap_test.go`
- Leverage existing `test_helper.go` utilities

**Alternatives Considered**:
- Catwalk (`github.com/knz/catwalk`): Data-driven testing, but teatest is official and more comprehensive
- Direct program testing without teatest: Too complex, no output capture

**Installation**: `go get github.com/charmbracelet/x/exp/teatest@latest`

### 3. OpenCode Configuration Format

**Decision**: Create `.opencode/` directory structure analogous to `.claude/`

**Rationale**:
- OpenCode is a distinct AI coding agent (separate from OpenDevin/OpenHands)
- Uses directory structure very similar to Claude Code for easy migration
- Supports backward compatibility by reading `.claude/` directories
- Has explicit JSON schema validation for configuration

**Directory Structure**:
```
.opencode/
├── commands/          # Custom slash commands (*.md files)
├── skills/            # Reusable instruction sets (*/SKILL.md)
├── agents/            # Custom agent personas (*.md) [OpenCode-specific]
├── modes/             # Operational modes [optional]
└── plugins/           # External plugins [optional]

opencode.json          # Primary config with JSON schema
AGENTS.md              # Project context (same as Claude Code)
```

**Key Differences from Claude Code**:

| Aspect | Claude Code | OpenCode |
|--------|-------------|----------|
| Config Directory | `.claude/` | `.opencode/` |
| Config File | `settings.local.json` | `opencode.json` (with schema) |
| Commands | `.claude/commands/*.md` | `.opencode/commands/*.md` |
| Skills | `.claude/skills/*/SKILL.md` | `.opencode/skills/*/SKILL.md` |
| Agents | System prompts | `.opencode/agents/*.md` (explicit) |
| Context File | `AGENTS.md` | `AGENTS.md` (same) |
| Backward Compat | N/A | Reads `.claude/` directories |

**Files to Create for OpenCode**:

1. **Required**:
   - `.opencode/` directory
   - `opencode.json` with schema reference
   - `AGENTS.md` (project context)

2. **Recommended** (ported from Claude Code):
   - `.opencode/commands/*.md` (all specledger.* and speckit.* commands)
   - `.opencode/skills/*/SKILL.md` (specledger-deps, issue-tracking, etc.)

**Example `opencode.json`**:
```json
{
  "$schema": "https://opencode.ai/config.json",
  "commands": {},
  "instructions": []
}
```

**Template Porting Strategy**:
- Commands: Same markdown format, copy `.claude/commands/*.md` → `.opencode/commands/*.md`
- Skills: Identical `SKILL.md` format, copy directly
- Settings: Translate permissions from `.claude/settings.local.json` to `opencode.json` tool configs
- Agent-specific files: Omit `.claude/settings.local.json`, use `opencode.json` instead

**Alternatives Considered**:
- Single config directory for all agents: Reduces clarity, harder to maintain agent-specific files
- No OpenCode support: Limits adoption by teams using alternative agents

**Configuration Schema**: `https://opencode.ai/config.json`

### 4. Go Code Quality Tools

**Decision**: Use standard Go toolchain with golangci-lint

**Rationale**:
- Already in use across the codebase
- Zero additional dependencies
- Industry standard for Go projects

**Tools**:
- `gofmt`: Code formatting (built-in)
- `go vet`: Static analysis (built-in)
- `golangci-lint`: Meta-linter running 50+ linters (verify in CI)

**Verification Needed**: Check if `golangci-lint` is configured in `.golangci.yml` and CI pipeline.

### 5. Structured Logging Strategy

**Decision**: Use standard library `log/slog` for structured logging

**Rationale**:
- Added in Go 1.21, now standard in Go 1.24+
- Zero dependencies
- Structured logging with levels (Debug, Info, Warn, Error)
- Context-aware logging

**Implementation**:
```go
import "log/slog"

// Log template selection
slog.Info("template selected",
    "template", selectedTemplate,
    "project", projectName)

// Log errors
slog.Error("failed to create project",
    "error", err,
    "directory", projectDir)
```

**Log Events**:
- Template selection
- Agent selection
- Project creation start/completion
- File copying operations
- Errors and warnings

**Alternatives Considered**:
- `logrus`: Third-party, unnecessary dependency
- `zap`: High performance, overkill for CLI tool
- Plain `log` package: No structured logging, no levels

---

## Architecture Decisions

### Template Definition Model

**Decision**: Extend manifest system with template metadata

**Structure**:
```go
type TemplateDefinition struct {
    ID              string   `yaml:"id"`              // "general-purpose", "full-stack"
    Name            string   `yaml:"name"`            // "General Purpose"
    Description     string   `yaml:"description"`     // One-line description
    Characteristics []string `yaml:"characteristics"` // ["Go", "CLI", "Single Binary"]
    Path            string   `yaml:"path"`            // "templates/general-purpose"
    IsDefault       bool     `yaml:"is_default"`      // true for general-purpose
}
```

**Manifest Extension**:
```yaml
templates:
  - id: general-purpose
    name: General Purpose
    description: Default template for any project type
    characteristics: [Go, CLI, Single Binary]
    path: templates/general-purpose
    is_default: true

  - id: full-stack
    name: Full-Stack Application
    description: Go backend with TypeScript/React frontend
    characteristics: [Go, TypeScript, React, REST API]
    path: templates/full-stack
    is_default: false
```

### Agent Configuration Model

**Structure**:
```go
type AgentConfig struct {
    ID          string `yaml:"id"`          // "claude-code", "opencode", "none"
    Name        string `yaml:"name"`        // "Claude Code"
    Description string `yaml:"description"` // Brief description
    ConfigDir   string `yaml:"config_dir"`  // ".claude", ".opencode", "" for none
}
```

**Hardcoded Options** (not in manifest):
```go
var supportedAgents = []AgentConfig{
    {
        ID:          "claude-code",
        Name:        "Claude Code",
        Description: "Anthropic's official CLI with commands and skills",
        ConfigDir:   ".claude",
    },
    {
        ID:          "opencode",
        Name:        "OpenCode",
        Description: "Open-source AI coding agent with LLM flexibility",
        ConfigDir:   ".opencode",
    },
    {
        ID:          "none",
        Name:        "None",
        Description: "Agent-agnostic setup (no agent-specific files)",
        ConfigDir:   "",
    },
}
```

### TUI Flow Extension

**New Steps**:
1. `stepProjectName` (existing)
2. `stepDirectory` (existing)
3. `stepShortCode` (existing)
4. **`stepTemplate`** (NEW) - Radio selection from 6 templates
5. **`stepAgent`** (NEW) - Radio selection from 3 agents
6. `stepPlaybook` (existing) - may be deprecated if templates replace playbooks
7. `stepConfirm` (existing)

**Selection Display Pattern**:
```
Select Project Template

  ◉ General Purpose
    Default template for any project type
    Tech: Go, CLI, Single Binary

  ○ Full-Stack Application
    Go backend with TypeScript/React frontend
    Tech: Go, TypeScript, React, REST API

  ○ Batch Data Processing
    Scheduled data pipelines with workflow orchestration
    Tech: Go, Airflow, DAGs

[↑/↓ to navigate, Enter to select]
```

---

## Implementation Checklist

**Phase 1: Data Model & Metadata**
- [ ] Add `github.com/google/uuid` dependency
- [ ] Extend `ProjectInfo` struct with `ID`, `Template`, `Agent` fields
- [ ] Create `TemplateDefinition` struct in `pkg/models/`
- [ ] Create `AgentConfig` struct in `pkg/models/`
- [ ] Update `metadata.Validate()` to validate new fields
- [ ] Update metadata version to "1.1.0"

**Phase 1: Template System**
- [ ] Define 6 template directories in `templates/`:
  - `general-purpose/` (copy current specledger playbook)
  - `full-stack/` (backend/ + frontend/ dirs)
  - `batch-data/` (dags/, data/, orchestration/)
  - `realtime-workflow/` (workflows/, activities/, workers/)
  - `ml-image/` (data/, training/, validation/, inference/)
  - `realtime-data/` (ingestion/, processing/, catalog/, features/, models/)
- [ ] Extend `manifest.yaml` with template definitions
- [ ] Update `EmbeddedSource` to load template list from manifest

**Phase 1: Agent Configuration**
- [ ] Create `supportedAgents` slice in `pkg/models/agent.go`
- [ ] Create OpenCode template structure in `templates/opencode/` with:
  - `.opencode/commands/` (port from `.claude/commands/`)
  - `.opencode/skills/` (port from `.claude/skills/`)
  - `opencode.json` with schema reference
  - `AGENTS.md` (shared with Claude Code)
- [ ] Update agent context scripts to support OpenCode

**Phase 1: TUI Updates**
- [ ] Add `stepTemplate` constant
- [ ] Add `stepAgent` constant
- [ ] Add `selectedTemplateIndex` field to `Model`
- [ ] Add `selectedAgentIndex` field to `Model`
- [ ] Add `templates []TemplateDefinition` field to `Model`
- [ ] Add `agents []AgentConfig` field to `Model`
- [ ] Implement `Update()` handler for `stepTemplate` (arrow keys + Enter)
- [ ] Implement `View()` renderer for `stepTemplate`
- [ ] Implement `Update()` handler for `stepAgent`
- [ ] Implement `View()` renderer for `stepAgent`
- [ ] Update confirmation review to display template and agent selections

**Phase 1: Bootstrap Integration**
- [ ] Update `bootstrap.go` to read template/agent from answers map
- [ ] Implement template directory copying based on selection
- [ ] Implement agent config directory creation based on selection
- [ ] Generate UUID and write to `specledger.yaml`
- [ ] Record template and agent in `specledger.yaml`

**Phase 2: CLI Flags**
- [ ] Add `--template <name>` flag to `sl new`
- [ ] Add `--agent <name>` flag to `sl new`
- [ ] Add `--list-templates` flag to `sl new`
- [ ] Implement non-interactive flow (skip TUI when flags provided)
- [ ] Validate flag values against supported options
- [ ] Implement TTY detection and require flags in non-interactive mode

**Phase 2: Testing**
- [ ] Add `github.com/charmbracelet/x/exp/teatest` dependency
- [ ] Create `pkg/cli/tui/sl_new_test.go` with unit tests
- [ ] Create `tests/integration/bootstrap_tui_test.go` with integration tests
- [ ] Add golden file tests for full TUI flow
- [ ] Set `lipgloss.SetColorProfile(termenv.Ascii)` in test init
- [ ] Add `*.golden -text` to `.gitattributes`
- [ ] Test all 6 templates create correct structures
- [ ] Test all 3 agents create correct config directories
- [ ] Test backward compatibility (general-purpose + claude-code = current behavior)
- [ ] Test non-interactive mode with flags
- [ ] Test TTY detection and error message

**Phase 3: Documentation & Polish**
- [ ] Add template descriptions to each template's README
- [ ] Document `--template`, `--agent`, `--list-templates` flags in CLI help
- [ ] Update project README with template catalog
- [ ] Add structured logging with `log/slog` for template/agent selection
- [ ] Verify `golangci-lint` configuration
- [ ] Update CLAUDE.md with new Go dependencies

---

## Open Questions

None - all NEEDS CLARIFICATION items resolved.

---

## References

**Dependencies**:
- `github.com/google/uuid` - UUID generation
- `gopkg.in/yaml.v3` - YAML marshaling (already in use)
- `github.com/charmbracelet/bubbletea` - TUI framework (already in use)
- `github.com/charmbracelet/bubbles` - TUI components (already in use)
- `github.com/charmbracelet/lipgloss` - TUI styling (already in use)
- `github.com/charmbracelet/x/exp/teatest` - TUI testing (new)
- `log/slog` - Structured logging (standard library)

**External Resources**:
- [google/uuid Package Documentation](https://pkg.go.dev/github.com/google/uuid)
- [OpenCode Documentation](https://opencode.ai/docs/)
- [Teatest Testing Guide](https://charm.land/blog/teatest/)
- [Bubble Tea Documentation](https://github.com/charmbracelet/bubbletea)
- [slog Package Documentation](https://pkg.go.dev/log/slog)

**Internal References**:
- Feature 011-streamline-onboarding: TUI implementation patterns
- Feature 005-embedded-templates: Template embedding and copying
- Feature 004-thin-wrapper-redesign: Metadata schema and validation
- `pkg/cli/tui/sl_new.go` - Current TUI implementation
- `pkg/cli/metadata/schema.go` - Current metadata schema
- `pkg/cli/playbooks/template.go` - Playbook system interfaces
