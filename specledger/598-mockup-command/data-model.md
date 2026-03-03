# Data Model: Mockup Command

**Branch**: `598-mockup-command` | **Date**: 2026-02-27 | **Updated**: 2026-03-03

## Entities

### 1. FrameworkType (Enum)

Identifies the frontend framework in use.

```go
type FrameworkType string

const (
    FrameworkReact     FrameworkType = "react"
    FrameworkNextJS    FrameworkType = "nextjs"
    FrameworkVue       FrameworkType = "vue"
    FrameworkNuxt      FrameworkType = "nuxt"
    FrameworkSvelte    FrameworkType = "svelte"
    FrameworkSvelteKit FrameworkType = "sveltekit"
    FrameworkAngular   FrameworkType = "angular"
    FrameworkUnknown   FrameworkType = "unknown"
)
```

**Validation Rules**:
- Must be one of the defined constants
- `unknown` is valid for non-frontend projects or when using `--force`

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

---

### 3. MockupFormat (Enum)

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
- Defaults to `html` if not specified via `--format` flag

---

### 4. DesignSystem

The design system document containing CSS tokens and styling conventions. Does NOT index individual components — the AI agent discovers those via codebase search.

```go
type DesignSystem struct {
    Version      int           `yaml:"version" json:"version"`
    Framework    FrameworkType `yaml:"framework" json:"framework"`
    LastScanned  time.Time     `yaml:"last_scanned" json:"last_scanned"`
    ExternalLibs []string      `yaml:"external_libs,omitempty" json:"external_libs,omitempty"`
    Style        *StyleInfo    `yaml:"style,omitempty" json:"style,omitempty"`
}
```

**Validation Rules**:
- `Version` must be `1` (current schema version)
- `Framework` must be valid `FrameworkType`
- `Style` contains extracted CSS tokens and styling patterns

**State Transitions**:
```
[Not Exists] --sl mockup--> [Generated]
[Generated] --sl mockup update--> [Updated]
```

---

### 5. StyleInfo

Describes the project's CSS/styling patterns and design tokens.

```go
type StyleInfo struct {
    CSSFramework    string            `json:"css_framework"`              // e.g., "Tailwind CSS", "Bootstrap"
    Preprocessor    string            `json:"preprocessor,omitempty"`     // e.g., "sass", "less"
    StylingApproach string            `json:"styling_approach"`           // e.g., "utility-first", "css-in-js", "css-modules"
    ThemeColors     map[string]string `json:"theme_colors,omitempty"`     // Extracted color tokens
    FontFamilies    []string          `json:"font_families,omitempty"`    // Extracted font families
    CSSVariables    []string          `json:"css_variables,omitempty"`    // Extracted CSS custom properties
    SampleImports   []string          `json:"sample_imports,omitempty"`   // Sample import patterns
}
```

**Validation Rules**:
- At least one of `CSSFramework`, `StylingApproach`, or `ThemeColors` should be non-empty

---

### 6. SpecContent

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

### 7. MockupPromptContext

Template rendering context for the AI agent prompt.

```go
type MockupPromptContext struct {
    SpecName        string        `json:"spec_name"`
    SpecPath        string        `json:"spec_path"`
    SpecTitle       string        `json:"spec_title"`
    Framework       FrameworkType `json:"framework"`
    Format          MockupFormat  `json:"format"`
    OutputPath      string        `json:"output_path"`
    ExternalLibs    []string      `json:"external_libs,omitempty"`
    HasDesignSystem bool          `json:"has_design_system"`
    Style           *StyleInfo    `json:"style,omitempty"`
    HasStyle        bool          `json:"has_style"`
}
```

**Validation Rules**:
- `SpecName` is required, must match an existing spec directory
- `OutputPath` is the path to the expected mockup output file
- `Format` must be valid `MockupFormat`

---

### 8. MockupResult

JSON output for `sl mockup --json` mode.

```go
type MockupResult struct {
    Status              string `json:"status"`
    Framework           string `json:"framework"`
    SpecName            string `json:"spec_name"`
    MockupPath          string `json:"mockup_path"`
    PromptPath          string `json:"prompt_path,omitempty"`
    Format              string `json:"format"`
    DesignSystemCreated bool   `json:"design_system_created"`
    AgentLaunched       bool   `json:"agent_launched"`
    Committed           bool   `json:"committed"`
}
```

---

### 9. UpdateResult

JSON output for `sl mockup update --json` mode.

```go
type UpdateResult struct {
    Status         string `json:"status"`
    ScanDurationMs int64  `json:"scan_duration_ms"`
}
```

---

## File Locations

| Artifact | Path | Format |
|----------|------|--------|
| Design System | `.specledger/memory/design-system.md` | Markdown + YAML frontmatter |
| Mockup Output (HTML) | `specledger/<spec-name>/mockup.html` | HTML (generated by AI agent) |
| Mockup Output (JSX) | `specledger/<spec-name>/mockup.jsx` | JSX (generated by AI agent) |
| Agent Prompt | `specledger/<spec-name>/mockup-prompt.md` | Markdown (generated by `--dry-run`) |
| Prompt Template | `pkg/cli/mockup/prompt.tmpl` | Go template (embedded) |

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
                           | StyleInfo        |
                           | (CSS tokens)     |
                           +------------------+
                                   |
                                   v
+------------------+       +----------------------+
| Spec.md          |------>| MockupPromptContext  |
| (parsed into     |       | (template context)   |
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
                           | Mockup File      |
                           | (HTML or JSX)    |
                           +------------------+
```

### Prompt Rendering Flow

```
SpecContent ─────────┐
                     ├──→ MockupPromptContext ──→ prompt.tmpl ──→ Agent Prompt
StyleInfo ───────────┤                                                │
  (CSS tokens,       │                                                v
   theme colors)     │                                           AI Agent
                     │                                                │
DetectionResult ─────┘                                     reads primary sources
  (Framework)                                                         │
                                                                      v
                                                   ┌──────────────────────────────┐
                                                   │ Primary Sources (READ FIRST) │
                                                   │ 1. spec.md                   │
                                                   │ 2. requirements.md           │
                                                   │ 3. data-model.md             │
                                                   └──────────────────────────────┘
                                                                      │
                                                                      v
                                                                Mockup File
```

---

## Schema Versioning

The design system file uses a `version` field in YAML frontmatter:

```yaml
---
version: 1
framework: react
last_scanned: 2026-02-27T10:00:00Z
---
```

**Migration Strategy**:
- Version 1: Initial schema (CSS tokens only, no component indexing)
- Future versions: Add `migrate()` function to upgrade older schemas
- Unknown versions: Warn user but attempt to parse as latest
