# Data Model: Project Templates & Coding Agent Selection

**Feature**: 592-init-project-templates
**Date**: 2026-02-18

This document defines the core entities, relationships, and state management for template and agent selection in `sl new`.

---

## Core Entities

### 1. TemplateDefinition

Represents a business-defined project template with its metadata and characteristics.

**Fields**:

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `ID` | `string` | Unique template identifier (kebab-case) | Required, matches `^[a-z][a-z0-9-]*$` |
| `Name` | `string` | Human-readable display name | Required, 1-50 chars |
| `Description` | `string` | One-line description for selection UI | Required, 1-200 chars |
| `Characteristics` | `[]string` | Key technologies/features (e.g., ["Go", "React"]) | Optional, max 6 items |
| `Path` | `string` | Relative path to template directory in embedded FS | Required, non-empty |
| `IsDefault` | `bool` | Whether this is the default template | Required, only one template can be default |

**Go Struct** (`pkg/models/template.go`):

```go
// TemplateDefinition defines a project template with its metadata
type TemplateDefinition struct {
    ID              string   `yaml:"id" json:"id"`
    Name            string   `yaml:"name" json:"name"`
    Description     string   `yaml:"description" json:"description"`
    Characteristics []string `yaml:"characteristics,omitempty" json:"characteristics,omitempty"`
    Path            string   `yaml:"path" json:"path"`
    IsDefault       bool     `yaml:"is_default" json:"is_default"`
}

// Validate checks if the template definition is valid
func (t *TemplateDefinition) Validate() error {
    if t.ID == "" {
        return errors.New("template ID is required")
    }
    if !regexp.MustCompile(`^[a-z][a-z0-9-]*$`).MatchString(t.ID) {
        return fmt.Errorf("template ID must be kebab-case: %s", t.ID)
    }
    if t.Name == "" || len(t.Name) > 50 {
        return errors.New("template name must be 1-50 characters")
    }
    if t.Description == "" || len(t.Description) > 200 {
        return errors.New("template description must be 1-200 characters")
    }
    if len(t.Characteristics) > 6 {
        return errors.New("template can have at most 6 characteristics")
    }
    if t.Path == "" {
        return errors.New("template path is required")
    }
    return nil
}

// String returns a formatted string representation for logging
func (t *TemplateDefinition) String() string {
    return fmt.Sprintf("%s (%s)", t.Name, t.ID)
}
```

**Instances** (defined in `pkg/embedded/templates/manifest.yaml`):

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

  - id: batch-data
    name: Batch Data Processing
    description: Scheduled data pipelines with workflow orchestration
    characteristics: [Go, Airflow, DAGs, Cron]
    path: templates/batch-data
    is_default: false

  - id: realtime-workflow
    name: Real-Time Workflow
    description: Durable long-running workflow orchestration
    characteristics: [Go, Temporal, Workflows, Activities]
    path: templates/realtime-workflow
    is_default: false

  - id: ml-image
    name: ML Image Processing
    description: Machine learning pipeline for image classification
    characteristics: [Python, TensorFlow, Training, Inference]
    path: templates/ml-image
    is_default: false

  - id: realtime-data
    name: Real-Time Data Pipeline
    description: Streaming data ingestion and processing
    characteristics: [Go, Kafka, Streams, Feature Store]
    path: templates/realtime-data
    is_default: false
```

**State Transitions**: Immutable after loading from manifest.

**Relationships**:
- Used by `TUIModel` for template selection step
- Referenced in `ProjectMetadata.Project.Template` after creation
- Mapped to embedded filesystem path for file copying

---

### 2. AgentConfig

Represents a coding agent configuration with installation details.

**Fields**:

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `ID` | `string` | Unique agent identifier (kebab-case) | Required, matches `^[a-z][a-z0-9-]*$` |
| `Name` | `string` | Human-readable display name | Required, 1-50 chars |
| `Description` | `string` | One-line description for selection UI | Required, 1-200 chars |
| `ConfigDir` | `string` | Directory name for agent config (e.g., ".claude") | Optional (empty for "none") |

**Go Struct** (`pkg/models/agent.go`):

```go
// AgentConfig defines a coding agent configuration
type AgentConfig struct {
    ID          string `yaml:"id" json:"id"`
    Name        string `yaml:"name" json:"name"`
    Description string `yaml:"description" json:"description"`
    ConfigDir   string `yaml:"config_dir,omitempty" json:"config_dir,omitempty"`
}

// Validate checks if the agent config is valid
func (a *AgentConfig) Validate() error {
    if a.ID == "" {
        return errors.New("agent ID is required")
    }
    if !regexp.MustCompile(`^[a-z][a-z0-9-]*$`).MatchString(a.ID) {
        return fmt.Errorf("agent ID must be kebab-case: %s", a.ID)
    }
    if a.Name == "" || len(a.Name) > 50 {
        return errors.New("agent name must be 1-50 characters")
    }
    if a.Description == "" || len(a.Description) > 200 {
        return errors.New("agent description must be 1-200 characters")
    }
    return nil
}

// HasConfig returns true if this agent requires configuration files
func (a *AgentConfig) HasConfig() bool {
    return a.ConfigDir != ""
}

// String returns a formatted string representation for logging
func (a *AgentConfig) String() string {
    return fmt.Sprintf("%s (%s)", a.Name, a.ID)
}
```

**Instances** (hardcoded in `pkg/models/agent.go`):

```go
// SupportedAgents returns the list of supported coding agents
func SupportedAgents() []AgentConfig {
    return []AgentConfig{
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
}

// GetAgentByID returns the agent config for a given ID
func GetAgentByID(id string) (*AgentConfig, error) {
    for _, agent := range SupportedAgents() {
        if agent.ID == id {
            return &agent, nil
        }
    }
    return nil, fmt.Errorf("unknown agent: %s", id)
}

// DefaultAgent returns the default agent (Claude Code)
func DefaultAgent() AgentConfig {
    return SupportedAgents()[0] // claude-code
}
```

**State Transitions**: Immutable, predefined list.

**Relationships**:
- Used by `TUIModel` for agent selection step
- Referenced in `ProjectMetadata.Project.Agent` after creation
- Mapped to agent-specific file templates for installation

---

### 3. ProjectMetadata (Extended)

Existing entity extended with new fields for template, agent, and project ID.

**New Fields in `ProjectInfo`**:

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `ID` | `uuid.UUID` | Unique project identifier (UUID v4) | Required, auto-generated |
| `Template` | `string` | Selected template ID | Optional (omitempty), must match a valid template ID if provided |
| `Agent` | `string` | Selected agent ID | Optional (omitempty), must match a valid agent ID if provided |

**Updated Go Struct** (`pkg/cli/metadata/schema.go`):

```go
import "github.com/google/uuid"

type ProjectInfo struct {
    ID        uuid.UUID `yaml:"id"`                   // NEW: Unique project identifier
    Name      string    `yaml:"name"`
    ShortCode string    `yaml:"short_code"`
    Template  string    `yaml:"template,omitempty"`   // NEW: Template ID (e.g., "full-stack")
    Agent     string    `yaml:"agent,omitempty"`      // NEW: Agent ID (e.g., "claude-code")
    Created   time.Time `yaml:"created"`
    Modified  time.Time `yaml:"modified"`
    Version   string    `yaml:"version"`
}

type ProjectMetadata struct {
    Version      string           `yaml:"version"` // Bump to "1.1.0"
    Project      ProjectInfo      `yaml:"project"`
    Playbook     PlaybookInfo     `yaml:"playbook"`
    Framework    FrameworkInfo    `yaml:"framework,omitempty"`
    TaskTracker  TaskTrackerInfo  `yaml:"task_tracker,omitempty"`
    Dependencies []Dependency     `yaml:"dependencies,omitempty"`
}
```

**Updated Validation** (`pkg/cli/metadata/schema.go`):

```go
func (m *ProjectMetadata) Validate() error {
    // Existing validations...

    // Validate project ID
    if m.Project.ID == uuid.Nil {
        return errors.New("project ID is required")
    }

    // Validate template ID if provided
    if m.Project.Template != "" {
        // Validate against known templates (optional strict validation)
        // For now, just check non-empty
    }

    // Validate agent ID if provided
    if m.Project.Agent != "" {
        // Validate against known agents (optional strict validation)
        // For now, just check non-empty
    }

    // Version must be 1.1.0 or higher for new fields
    if m.Project.Template != "" || m.Project.Agent != "" {
        if m.Version < "1.1.0" {
            return errors.New("metadata version must be 1.1.0 or higher for template/agent fields")
        }
    }

    return nil
}
```

**Example YAML**:

```yaml
version: 1.1.0
project:
    id: 550e8400-e29b-41d4-a716-446655440000
    name: my-project
    short_code: mp
    template: full-stack
    agent: claude-code
    created: 2026-02-18T17:05:59+07:00
    modified: 2026-02-18T17:05:59+07:00
    version: 0.1.0
playbook:
    name: specledger
    version: 1.0.0
    applied_at: 2026-02-18T17:05:59+07:00
task_tracker:
    choice: beads
    enabled_at: 2026-02-18T17:05:59+07:00
```

**State Transitions**:
- `ID`: Generated once at project creation, never changes
- `Template`, `Agent`: Set once at project creation, never changes
- `Modified`: Updated on every metadata save

**Relationships**:
- `Template` references `TemplateDefinition.ID`
- `Agent` references `AgentConfig.ID`
- Used by `sl session list` to map project ID to Supabase bucket

---

### 4. TUIModel (Extended)

Extended TUI model to support template and agent selection steps.

**New Fields**:

| Field | Type | Description |
|-------|------|-------------|
| `templates` | `[]TemplateDefinition` | Available templates loaded from manifest |
| `selectedTemplateIndex` | `int` | Currently highlighted template (0-based) |
| `agents` | `[]AgentConfig` | Available agents (hardcoded list) |
| `selectedAgentIndex` | `int` | Currently highlighted agent (0-based) |

**New Step Constants**:

```go
const (
    stepProjectName = iota
    stepDirectory
    stepShortCode
    stepTemplate        // NEW
    stepAgent           // NEW
    stepPlaybook        // May be deprecated/skipped if template replaces playbook
    stepConfirm
    stepDone
)
```

**Extended Go Struct** (`pkg/cli/tui/sl_new.go`):

```go
type Model struct {
    // Existing fields...
    step                  int
    textInput             textinput.Model
    answers               map[string]string
    showingError          string
    quitting              bool

    // NEW: Template selection
    templates             []models.TemplateDefinition
    selectedTemplateIndex int

    // NEW: Agent selection
    agents                []models.AgentConfig
    selectedAgentIndex    int
}

// InitialModel creates a new TUI model with templates and agents loaded
func InitialModel(cwd string) Model {
    templates, err := loadTemplatesFromManifest()
    if err != nil {
        // Log error and use default template only
        templates = []models.TemplateDefinition{
            {
                ID:          "general-purpose",
                Name:        "General Purpose",
                Description: "Default template for any project type",
                IsDefault:   true,
            },
        }
    }

    agents := models.SupportedAgents()

    return Model{
        step:                  stepProjectName,
        textInput:             textinput.New(),
        answers:               make(map[string]string),
        templates:             templates,
        selectedTemplateIndex: findDefaultTemplateIndex(templates),
        agents:                agents,
        selectedAgentIndex:    0, // Claude Code is default
    }
}

// findDefaultTemplateIndex returns the index of the default template
func findDefaultTemplateIndex(templates []models.TemplateDefinition) int {
    for i, t := range templates {
        if t.IsDefault {
            return i
        }
    }
    return 0 // Fallback to first template
}
```

**State Transitions**:

| From Step | Key Press | To Step | Action |
|-----------|-----------|---------|--------|
| `stepShortCode` | Enter (valid) | `stepTemplate` | Store short_code, initialize template selection |
| `stepTemplate` | ↑ | `stepTemplate` | Decrement `selectedTemplateIndex` (with wrap) |
| `stepTemplate` | ↓ | `stepTemplate` | Increment `selectedTemplateIndex` (with wrap) |
| `stepTemplate` | Enter | `stepAgent` | Store template ID, initialize agent selection |
| `stepAgent` | ↑ | `stepAgent` | Decrement `selectedAgentIndex` (with wrap) |
| `stepAgent` | ↓ | `stepAgent` | Increment `selectedAgentIndex` (with wrap) |
| `stepAgent` | Enter | `stepPlaybook` or `stepConfirm` | Store agent ID, proceed to next step |
| Any step | Ctrl+C | `stepDone` | Set `quitting = true` |

**Answers Map Keys**:

| Key | Value | Set At |
|-----|-------|--------|
| `"project_name"` | Project name string | `stepProjectName` |
| `"project_dir"` | Directory path | `stepDirectory` |
| `"short_code"` | Short code (≤4 chars) | `stepShortCode` |
| `"template"` | Template ID (e.g., "full-stack") | `stepTemplate` |
| `"agent"` | Agent ID (e.g., "claude-code") | `stepAgent` |
| `"playbook"` | Playbook name (if still used) | `stepPlaybook` |

**Relationships**:
- Loads `templates` from `TemplateDefinition` via manifest
- Loads `agents` from `AgentConfig.SupportedAgents()`
- Stores selections in `answers` map
- Consumed by bootstrap command to create project

---

## Relationships Diagram

```
┌─────────────────────────────┐
│  TemplateDefinition         │
│  (manifest.yaml)            │
│  - id: "full-stack"         │
│  - name: "Full-Stack App"   │
│  - path: "templates/..."    │
└──────────────┬──────────────┘
               │ loads
               │
               ▼
┌─────────────────────────────┐        ┌─────────────────────────────┐
│  TUIModel                   │        │  AgentConfig                │
│  (sl_new.go)                │        │  (hardcoded list)           │
│  - templates []TD           │        │  - id: "claude-code"        │
│  - selectedTemplateIndex    │◄───────┤  - name: "Claude Code"      │
│  - agents []AC              │  loads │  - config_dir: ".claude"    │
│  - selectedAgentIndex       │        └─────────────────────────────┘
│  - answers map[string]string│
└──────────────┬──────────────┘
               │ saves to
               │
               ▼
┌─────────────────────────────┐
│  ProjectMetadata            │
│  (specledger.yaml)          │
│  project:                   │
│    id: UUID                 │───────► Used by Supabase session storage
│    template: "full-stack"   │───────► References TemplateDefinition.id
│    agent: "claude-code"     │───────► References AgentConfig.id
└─────────────────────────────┘
               │
               │ used by
               ▼
┌─────────────────────────────┐
│  Bootstrap Command          │
│  (commands/bootstrap.go)    │
│  - Reads answers from TUI   │
│  - Generates UUID            │
│  - Copies template files    │
│  - Creates agent config dir │
│  - Writes specledger.yaml   │
└─────────────────────────────┘
```

---

## Data Flow

### 1. TUI Initialization

```
User runs: sl new
  │
  ▼
InitialModel()
  ├─► Load templates from manifest.yaml
  ├─► Load agents from SupportedAgents()
  ├─► Set selectedTemplateIndex to default template
  ├─► Set selectedAgentIndex to 0 (Claude Code)
  └─► Return Model with templates, agents loaded
```

### 2. Template Selection

```
User navigates with ↑/↓
  │
  ▼
Update() handles tea.KeyUp / tea.KeyDown
  ├─► Increment/decrement selectedTemplateIndex
  ├─► Wrap around (0 → len-1, len-1 → 0)
  └─► Re-render View() with new selection

User presses Enter
  │
  ▼
Update() handles tea.KeyEnter
  ├─► Validate template exists (always true)
  ├─► Store templates[selectedTemplateIndex].ID → answers["template"]
  ├─► Transition to stepAgent
  └─► Initialize agent selection UI
```

### 3. Agent Selection

```
User navigates with ↑/↓
  │
  ▼
Update() handles tea.KeyUp / tea.KeyDown
  ├─► Increment/decrement selectedAgentIndex
  ├─► Wrap around (0 → len-1, len-1 → 0)
  └─► Re-render View() with new selection

User presses Enter
  │
  ▼
Update() handles tea.KeyEnter
  ├─► Validate agent exists (always true)
  ├─► Store agents[selectedAgentIndex].ID → answers["agent"]
  ├─► Transition to stepConfirm
  └─► Display confirmation review with all selections
```

### 4. Project Creation

```
User confirms at stepConfirm
  │
  ▼
Bootstrap command receives answers map
  │
  ├─► Generate UUID: uuid.New()
  │
  ├─► Read template ID from answers["template"]
  │   ├─► Load TemplateDefinition from manifest
  │   └─► Copy files from templates/<template-id>/ to project dir
  │
  ├─► Read agent ID from answers["agent"]
  │   ├─► Load AgentConfig from SupportedAgents()
  │   └─► Copy agent config files to project dir (if HasConfig())
  │
  ├─► Create ProjectMetadata
  │   ├─► Set project.id = UUID
  │   ├─► Set project.template = template ID
  │   ├─► Set project.agent = agent ID
  │   └─► Set project.created = time.Now()
  │
  └─► Write specledger.yaml
      └─► Success: return nil
```

---

## Validation Rules

### Template Selection

| Rule | Validation | Error Message |
|------|------------|---------------|
| Template exists | Selected index within bounds | "Invalid template selection" |
| Template ID not empty | `template.ID != ""` | "Template ID is required" |
| Template path exists | Path exists in embedded FS | "Template path not found: {path}" |

### Agent Selection

| Rule | Validation | Error Message |
|------|------------|---------------|
| Agent exists | Selected index within bounds | "Invalid agent selection" |
| Agent ID not empty | `agent.ID != ""` | "Agent ID is required" |

### Metadata Creation

| Rule | Validation | Error Message |
|------|------------|---------------|
| UUID generation succeeds | `uuid.New()` doesn't panic | "Failed to generate project ID" |
| Template ID valid | Exists in manifest | "Unknown template: {template}" (warning, not fatal) |
| Agent ID valid | Exists in SupportedAgents() | "Unknown agent: {agent}" (warning, not fatal) |
| Metadata version | Version ≥ "1.1.0" if template/agent set | "Metadata version must be 1.1.0+" |

---

## State Management

### TUI State

**Immutable After Loading**:
- `templates []TemplateDefinition` - Loaded once from manifest
- `agents []AgentConfig` - Loaded once from hardcoded list

**Mutable**:
- `selectedTemplateIndex int` - Updated by arrow key navigation
- `selectedAgentIndex int` - Updated by arrow key navigation
- `step int` - Updated by Enter key to progress through steps
- `answers map[string]string` - Updated at end of each step

**Transient**:
- `showingError string` - Set on validation failure, cleared on next input

### Metadata State

**Write-Once**:
- `project.id` - Generated at creation, never changes
- `project.template` - Set at creation, never changes
- `project.agent` - Set at creation, never changes
- `project.created` - Set at creation, never changes

**Updated on Save**:
- `project.modified` - Auto-updated by `metadata.Save()`

---

## Error Handling

### TUI Errors

| Scenario | Handling | User Experience |
|----------|----------|-----------------|
| Manifest load failure | Use default template only, log error | User sees "General Purpose" template, warning logged |
| No templates in manifest | Use default template only, log error | User sees "General Purpose" template, warning logged |
| No agents available | Use hardcoded list (cannot fail) | N/A |

### Bootstrap Errors

| Scenario | Handling | User Experience |
|----------|----------|-----------------|
| UUID generation failure | Panic (should never happen) | Fatal error, exit |
| Template path not found | Return error, abort creation | "Template not found: {path}" |
| Agent config copy failure | Log warning, continue | Warning logged, project created without agent files |
| Metadata write failure | Return error, abort creation | "Failed to create specledger.yaml: {error}" |

### CLI Flag Errors

| Scenario | Handling | User Experience |
|----------|----------|-----------------|
| Invalid `--template` value | Print error + available templates, exit | "Unknown template: {value}\nAvailable: ..." |
| Invalid `--agent` value | Print error + available agents, exit | "Unknown agent: {value}\nAvailable: ..." |
| Non-interactive without flags | Print error + usage, exit | "Interactive terminal required or use --template/--agent" |

---

## Performance Characteristics

### TUI Rendering

- Template list size: 6 items (fixed)
- Agent list size: 3 items (fixed)
- Render time: <5ms per frame (Bubble Tea handles efficiently)
- Memory: <1MB for TUI state

### File Operations

- Template copying: O(n) where n = number of files in template
- Largest template: ~100 files (full-stack with backend + frontend)
- Copy time: <3 seconds for largest template

### Metadata I/O

- Read: <1ms (small YAML file, ~1KB)
- Write: <10ms (YAML marshal + disk write)
- UUID generation: <1μs (crypto/rand is fast)

---

## Backward Compatibility

### Metadata Version Migration

**Version 1.0.0** (current):
```yaml
version: 1.0.0
project:
    name: specledger
    short_code: sp
    created: 2026-02-18T17:05:59+07:00
    modified: 2026-02-18T17:05:59+07:00
    version: 0.1.0
```

**Version 1.1.0** (with new fields):
```yaml
version: 1.1.0
project:
    id: 550e8400-e29b-41d4-a716-446655440000  # NEW
    name: specledger
    short_code: sp
    template: general-purpose                   # NEW (optional)
    agent: claude-code                          # NEW (optional)
    created: 2026-02-18T17:05:59+07:00
    modified: 2026-02-18T17:05:59+07:00
    version: 0.1.0
```

**Migration Strategy**:
- Old projects without `id`: Generate UUID on first load, update metadata
- Old projects without `template`/`agent`: Infer from directory structure or leave empty
- No breaking changes: `omitempty` tags allow fields to be absent

### Default Behavior Preservation

**Current `sl new` behavior**:
- Prompts for: project name, directory, short code, playbook
- Creates: General Purpose structure with Claude Code config

**New `sl new` behavior with backward compatibility**:
- Default template: "General Purpose" (selected by default)
- Default agent: "Claude Code" (selected by default)
- Result: Identical to current behavior if user accepts defaults

**Test Case** (SC-003):
```
User runs: sl new
User enters: name="test", dir="/tmp", shortcode="t"
User presses: Enter (accept default template)
User presses: Enter (accept default agent)

Expected: Project structure identical to old sl new
```

---

## Security Considerations

### UUID Generation

- Uses `crypto/rand` via `github.com/google/uuid` (cryptographically secure)
- UUID v4 has 122 bits of randomness (collision probability: negligible)
- No predictable patterns or sequential IDs

### Template Path Traversal

- Template paths restricted to embedded filesystem (no user-provided paths)
- Manifest validated at build time (paths must exist)
- No risk of path traversal attacks

### Agent Config Injection

- Agent IDs validated against hardcoded list (no arbitrary agent loading)
- Config directories created in project root only (no arbitrary paths)
- No risk of injecting malicious agent configurations

---

## Testing Strategy

### Unit Tests

**TemplateDefinition**:
- Test `Validate()` with valid/invalid IDs, names, descriptions
- Test `String()` formatting
- Test default template detection

**AgentConfig**:
- Test `Validate()` with valid/invalid IDs, names
- Test `HasConfig()` for agents with/without config dirs
- Test `GetAgentByID()` with valid/invalid IDs
- Test `DefaultAgent()` returns Claude Code

**ProjectMetadata**:
- Test UUID marshaling/unmarshaling to YAML
- Test validation with/without new fields
- Test version enforcement (1.1.0 required for new fields)

### Integration Tests

**TUI Template Selection**:
- Test arrow key navigation wraps correctly
- Test Enter key stores correct template ID
- Test default template is pre-selected

**TUI Agent Selection**:
- Test arrow key navigation wraps correctly
- Test Enter key stores correct agent ID
- Test default agent (Claude Code) is pre-selected

**Bootstrap Command**:
- Test UUID generation and storage
- Test template copying for all 6 templates
- Test agent config creation for all 3 agents
- Test backward compatibility (general-purpose + claude-code)

### End-to-End Tests

**Full Flow**:
- Test complete TUI flow with template/agent selection
- Test golden file output for UI rendering
- Test non-interactive mode with `--template` and `--agent` flags
- Test `--list-templates` flag

**Backward Compatibility**:
- Test old `sl new` behavior with defaults produces identical result
- Test loading old metadata (v1.0.0) doesn't fail

---

## Summary

This data model extends the existing SpecLedger architecture with minimal changes:

1. **New entities**: `TemplateDefinition`, `AgentConfig` (simple structs with validation)
2. **Extended entity**: `ProjectMetadata.Project` (3 new fields: ID, Template, Agent)
3. **Extended TUI**: `TUIModel` (4 new fields: templates, selectedTemplateIndex, agents, selectedAgentIndex)
4. **New steps**: `stepTemplate`, `stepAgent` (following existing TUI patterns)

All entities are immutable after loading except for TUI selection state. Validation is enforced at every boundary (TUI input, metadata save, CLI flags). Backward compatibility is preserved through default values and optional fields.
