# Data Model: Mockup Command

**Branch**: `598-mockup-command` | **Date**: 2026-02-27

## Entities

### 1. FrameworkType (Enum)

Identifies the frontend framework in use.

```go
type FrameworkType string

const (
    FrameworkReact    FrameworkType = "react"
    FrameworkNextJS   FrameworkType = "nextjs"
    FrameworkVue      FrameworkType = "vue"
    FrameworkNuxt     FrameworkType = "nuxt"
    FrameworkSvelte   FrameworkType = "svelte"
    FrameworkSvelteKit FrameworkType = "sveltekit"
    FrameworkAngular  FrameworkType = "angular"
    FrameworkUnknown  FrameworkType = "unknown"
)
```

**Validation Rules**:
- Must be one of the defined constants
- `unknown` is valid for non-frontend projects

---

### 2. DetectionResult

Result of frontend framework detection.

```go
type DetectionResult struct {
    IsFrontend    bool          `json:"is_frontend"`
    Framework     FrameworkType `json:"framework"`
    Confidence    int           `json:"confidence"`     // 0-100
    ComponentDirs []string      `json:"component_dirs"` // Detected component directories
    Indicators    []string      `json:"indicators"`     // What triggered detection
    ConfigFile    string        `json:"config_file"`    // Primary config file found
}
```

**Validation Rules**:
- `Confidence` must be 0-100
- `IsFrontend` is `true` only if `Confidence >= 70`
- `ComponentDirs` may be empty if framework detected but no components found

**State Transitions**:
- N/A (stateless result object)

---

### 3. Component

A single UI component discovered in the codebase.

```go
type Component struct {
    Name        string     `yaml:"name" json:"name"`
    FilePath    string     `yaml:"path" json:"path"`
    Description string     `yaml:"description,omitempty" json:"description,omitempty"`
    Props       []PropInfo `yaml:"props,omitempty" json:"props,omitempty"`
    IsExternal  bool       `yaml:"external,omitempty" json:"external,omitempty"`
    Library     string     `yaml:"library,omitempty" json:"library,omitempty"`
}

type PropInfo struct {
    Name     string `yaml:"name" json:"name"`
    Type     string `yaml:"type,omitempty" json:"type,omitempty"`
    Required bool   `yaml:"required,omitempty" json:"required,omitempty"`
}
```

**Validation Rules**:
- `Name` is required, must be valid identifier (alphanumeric + underscore)
- `FilePath` is required for project components, empty for external
- `IsExternal` and `Library` must both be set if component is from external library

---

### 4. DesignSystem

The full design system index for a project.

```go
type DesignSystem struct {
    Version       int         `yaml:"version" json:"version"`
    Framework     FrameworkType `yaml:"framework" json:"framework"`
    LastScanned   time.Time   `yaml:"last_scanned" json:"last_scanned"`
    ComponentDirs []string    `yaml:"component_dirs" json:"component_dirs"`
    ExternalLibs  []string    `yaml:"external_libs,omitempty" json:"external_libs,omitempty"`
    Components    []Component `yaml:"components" json:"components"`
    ManualEntries []Component `yaml:"manual_entries,omitempty" json:"manual_entries,omitempty"`
}
```

**Validation Rules**:
- `Version` must be `1` (current schema version)
- `Framework` must be valid `FrameworkType`
- `Components` may be empty for newly initialized projects
- `ManualEntries` preserved across `sl mockup update` operations

**State Transitions**:
```
[Not Exists] --sl mockup--> [Generated]
[Generated] --sl mockup update--> [Updated]
[Updated] --manual edit--> [Modified]
[Modified] --sl mockup update--> [Merged]
```

---

### 5. MockupFormat (Enum)

Output format for the generated mockup file.

```go
type MockupFormat string

const (
    MockupFormatHTML MockupFormat = "html"
    MockupFormatJSX  MockupFormat = "jsx"
)
```

**Validation Rules**:
- Must be one of the defined constants
- Defaults to `html` if not specified

---

### 6. Mockup

Generated mockup artifact for a feature spec.

```go
type Mockup struct {
    SpecName      string         `yaml:"spec_name" json:"spec_name"`
    SpecPath      string         `yaml:"spec_path" json:"spec_path"`
    Generated     time.Time      `yaml:"generated" json:"generated"`
    Framework     FrameworkType  `yaml:"framework" json:"framework"`
    Format        MockupFormat   `yaml:"format" json:"format"`
    Screens       []Screen       `yaml:"screens" json:"screens"`
    ComponentMap  []ComponentRef `yaml:"component_map" json:"component_map"`
}

type Screen struct {
    Name        string `yaml:"name" json:"name"`
    Description string `yaml:"description,omitempty" json:"description,omitempty"`
    Content     string `yaml:"content" json:"content"`   // HTML or JSX content for this screen
    Flow        string `yaml:"flow,omitempty" json:"flow,omitempty"` // User flow description
}

type ComponentRef struct {
    UIElement      string `yaml:"ui_element" json:"ui_element"`
    ComponentName  string `yaml:"component" json:"component"`
    Source         string `yaml:"source" json:"source"` // File path or library name
}
```

**Validation Rules**:
- `SpecName` must match an existing spec directory
- `Screens` must have at least one entry
- `Format` must be valid `MockupFormat`
- `Content` must be valid HTML or JSX depending on `Format`

---

### 7. ScanResult

Result of component scanning operation.

```go
type ScanResult struct {
    Components    []Component   `json:"components"`
    Framework     FrameworkType `json:"framework"`
    ComponentDirs []string      `json:"component_dirs"`
    ExternalLibs  []string      `json:"external_libs"`
    ScanDuration  time.Duration `json:"scan_duration"`
    FilesScanned  int           `json:"files_scanned"`
    Errors        []ScanError   `json:"errors,omitempty"`
}

type ScanError struct {
    FilePath string `json:"file_path"`
    Error    string `json:"error"`
}
```

**Validation Rules**:
- `Framework` must be valid `FrameworkType`
- `Errors` contains non-fatal scan issues (e.g., permission denied on one file)

---

### 8. MockupPromptContext

Template rendering context for the AI agent prompt.

```go
type MockupPromptContext struct {
    SpecName       string            `json:"spec_name"`
    SpecContent    SpecContent       `json:"spec_content"`
    Framework      FrameworkType     `json:"framework"`
    Format         MockupFormat      `json:"format"`
    OutputPath     string            `json:"output_path"`
    Components     []PromptComponent `json:"components"`
    ExternalLibs   []string          `json:"external_libs,omitempty"`
    HasDesignSystem bool             `json:"has_design_system"`
}
```

**Validation Rules**:
- `SpecName` is required, must match an existing spec directory
- `Components` is the user-selected subset from the design system
- `OutputPath` is the full path to the expected mockup output file
- `Format` must be valid `MockupFormat`

---

### 9. PromptComponent

A component entry included in the agent prompt (subset of Component with prompt-relevant fields).

```go
type PromptComponent struct {
    Name        string `json:"name"`
    FilePath    string `json:"path"`
    Description string `json:"description,omitempty"`
    Props       string `json:"props,omitempty"` // Formatted props summary for prompt
    IsExternal  bool   `json:"external,omitempty"`
    Library     string `json:"library,omitempty"`
}
```

**Validation Rules**:
- `Name` is required
- `Props` is a human-readable summary (e.g., `"variant: string, onClick: func, disabled: bool"`) not the full `[]PropInfo`

---

### 10. SpecContent

Parsed content from a `spec.md` file used to build the agent prompt.

```go
type SpecContent struct {
    Title        string   `json:"title"`
    UserStories  []string `json:"user_stories"`
    Requirements []string `json:"requirements"`
    FullContent  string   `json:"full_content"`
}
```

**Validation Rules**:
- `Title` is extracted from the first H1 heading
- `UserStories` extracted from `### User Story` sections
- `Requirements` extracted from `### Functional Requirements` section
- `FullContent` is the raw markdown text (used as fallback if structured parsing fails)

---

## File Locations

| Artifact | Path | Format |
|----------|------|--------|
| Design System Index | `specledger/design_system.md` | Markdown + YAML frontmatter |
| Mockup Output (HTML) | `specledger/<spec-name>/mockup.html` | HTML (generated by AI agent) |
| Mockup Output (JSX) | `specledger/<spec-name>/mockup.jsx` | JSX (generated by AI agent) |
| Agent Prompt | `specledger/<spec-name>/mockup-prompt.md` | Markdown (generated by `--dry-run`) |
| Prompt Template | `pkg/cli/mockup/prompt.tmpl` | Go template (embedded) |
| Project Config | `specledger/specledger.yaml` | YAML |

---

## Relationships

```
+------------------+       +------------------+
| DetectionResult  |------>| DesignSystem     |
| (detects)        |       | (creates/updates)|
+------------------+       +------------------+
                                   |
                                   | contains
                                   v
                           +------------------+
                           | Component[]      |
                           +------------------+
                                   |
                                   | user selects subset
                                   v
+------------------+       +----------------------+
| Spec.md          |------>| MockupPromptContext   |
| (parsed into     |       | (template context)    |
|  SpecContent)    |       +----------------------+
+------------------+              |
                                  | rendered via prompt.tmpl
                                  v
                           +------------------+
                           | Agent Prompt     |
                           | (markdown text)  |
                           +------------------+
                                  |
                                  | launched via launcher
                                  v
                           +------------------+
                           | AI Agent         |
                           | (generates)      |
                           +------------------+
                                  |
                                  | produces
                                  v
                           +------------------+
                           | Mockup           |
                           | (HTML or JSX)    |
                           +------------------+
                                  |
                                  | contains
                                  v
                           +------------------+
                           | Screen[]         |
                           | ComponentRef[]   |
                           +------------------+
```

### Prompt Rendering Flow

```
SpecContent ─────────┐
                     ├──→ MockupPromptContext ──→ prompt.tmpl ──→ Agent Prompt
DesignSystem ────────┤                                                │
  (selected          │                                                v
   PromptComponent[])│                                          AI Agent
                     │                                                │
DetectionResult ─────┘                                                v
  (Framework, Format)                                           Mockup File
```

---

## Schema Versioning

The design system file uses a `version` field in YAML frontmatter:

```yaml
---
version: 1
---
```

**Migration Strategy**:
- Version 1: Initial schema (this implementation)
- Future versions: Add `migrate()` function to upgrade older schemas
- Unknown versions: Warn user but attempt to parse as latest
