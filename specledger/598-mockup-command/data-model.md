# Data Model: Mockup Command

**Branch**: `598-mockup-command` | **Date**: 2026-02-27 | **Updated**: 2026-03-06

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
    FrameworkAstro     FrameworkType = "astro"
    FrameworkSolid     FrameworkType = "solid"
    FrameworkQwik      FrameworkType = "qwik"
    FrameworkRemix     FrameworkType = "remix"
    FrameworkUnknown   FrameworkType = "unknown"
)
```

**Validation Rules**:
- Must be one of the defined constants
- `unknown` is valid for non-frontend projects or when using `--force`
- Includes Astro, SolidJS, Qwik, and Remix as distinct framework types (Remix was previously mapped to React)

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
    Version      int            `yaml:"version" json:"version"`
    Framework    FrameworkType  `yaml:"framework" json:"framework"`
    LastScanned  time.Time      `yaml:"last_scanned" json:"last_scanned"`
    ExternalLibs []string       `yaml:"external_libs,omitempty" json:"external_libs,omitempty"`
    Style        *StyleInfo     `yaml:"style,omitempty" json:"style,omitempty"`
    AppStructure *AppStructure  `yaml:"app_structure,omitempty" json:"app_structure,omitempty"`
}
```

**Validation Rules**:
- `Version` must be `1` (current schema version)
- `Framework` must be valid `FrameworkType`
- `Style` contains extracted CSS tokens and styling patterns
- `AppStructure` contains layout files, component directories, and global stylesheets discovered by `ScanAppStructure()`

**State Transitions**:
```
[Not Exists] --sl mockup--> [Generated]
[Generated] --sl mockup update--> [Updated]
```

---

### 5. AppStructure

Describes the project's layout and routing structure, discovered by `ScanAppStructure()`.

```go
type AppStructure struct {
    Router       string   `yaml:"router,omitempty" json:"router,omitempty"`         // e.g., "app-router", "pages-router", "file-based", "component-based", "module-based"
    Layouts      []string `yaml:"layouts,omitempty" json:"layouts,omitempty"`       // Layout file paths (e.g., "app/layout.tsx")
    Components   []string `yaml:"components,omitempty" json:"components,omitempty"` // Component file/dir paths (max 50)
    GlobalStyles []string `yaml:"global_styles,omitempty" json:"global_styles,omitempty"` // Global stylesheet paths (max 30)
}
```

**Validation Rules**:
- `Router` is framework-specific: Next.js uses "app-router"/"pages-router", Nuxt/SvelteKit/Astro/Remix/Qwik use "file-based", React/Vue/Svelte/Solid use "component-based", Angular uses "module-based"
- Returns `nil` if no layouts, components, or global styles are found
- Layout detection is framework-aware (e.g., Next.js looks for `layout.tsx` in `app/`, SvelteKit looks for `+layout.svelte` in `src/routes/`)

---

### 6. StyleInfo

Describes the project's CSS/styling patterns and design tokens.

```go
type StyleInfo struct {
    CSSFramework    string            `yaml:"css_framework" json:"css_framework"`
    Preprocessor    string            `yaml:"preprocessor,omitempty" json:"preprocessor,omitempty"`
    StylingApproach string            `yaml:"styling_approach" json:"styling_approach"`
    ThemeColors     map[string]string `yaml:"theme_colors,omitempty" json:"theme_colors,omitempty"`
    FontFamilies    []string          `yaml:"font_families,omitempty" json:"font_families,omitempty"`
    CSSVariables    []string          `yaml:"css_variables,omitempty" json:"css_variables,omitempty"`
    SampleImports   []string          `yaml:"sample_imports,omitempty" json:"sample_imports,omitempty"`
    ComponentLibs   []string          `yaml:"component_libs,omitempty" json:"component_libs,omitempty"`
}
```

**Validation Rules**:
- At least one of `CSSFramework`, `StylingApproach`, or `ThemeColors` should be non-empty
- `ComponentLibs` lists detected UI component libraries (e.g., "shadcn/ui", "MUI (Material UI)", "Chakra UI")
- `CSSFramework` may include version info (e.g., "Tailwind CSS v4" for CSS-based config)

---

### 7. SpecContent

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

### 8. MockupPromptContext

Template rendering context for the AI agent prompt. Style and design system data is **not embedded** in the prompt — the template instructs the agent to read `.specledger/memory/design-system.md` directly.

```go
type MockupPromptContext struct {
    SpecName   string        `json:"spec_name"`
    SpecPath   string        `json:"spec_path"`
    SpecTitle  string        `json:"spec_title"`
    Framework  FrameworkType `json:"framework"`
    Format     MockupFormat  `json:"format"`
    OutputPath string        `json:"output_path"`
    UserPrompt string        `json:"user_prompt,omitempty"`
}
```

**Validation Rules**:
- `SpecName` is required, must match an existing spec directory
- `OutputPath` is the path to the expected mockup output file
- `Format` must be valid `MockupFormat`
- `UserPrompt` contains additional user instructions from positional args

---

### 9. MockupResult

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

### 10. UpdateResult

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
| App Scanner | `pkg/cli/mockup/appscan.go` | Go source (layout/component/style discovery) |

---

## Relationships

```
+------------------+       +------------------+
| DetectionResult  |------>| DesignSystem     |
| (detects)        |       | (creates/updates)|
+------------------+       +------------------+
                                   |
                           contains ├──────────────┐
                                   v              v
                           +------------------+  +------------------+
                           | StyleInfo        |  | AppStructure     |
                           | (CSS tokens,     |  | (layouts, comps, |
                           |  component libs) |  |  global styles)  |
                           +------------------+  +------------------+
                                   |
                                   v
+------------------+       +----------------------+
| Spec.md          |------>| MockupPromptContext  |
| (parsed into     |       | (template context —  |
|  SpecContent)    |       |  no style data)      |
+------------------+       +----------------------+
                                  |
                                  | rendered via prompt.tmpl
                                  v
                           +------------------+
                           | Agent Prompt     |
                           | (instructs agent |
                           |  to READ design- |
                           |  system.md)      |
                           +------------------+
                                  |
                                  | launched via launcher
                                  v
                           +------------------+
                           | AI Agent         |
                           | (reads design-   |
                           |  system.md +     |
                           |  generates)      |
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
DetectionResult ─────┘     (no style data —                          │
  (Framework)               agent reads it)                          v
                                                                AI Agent
                                                                     │
                                                          reads sources at runtime
                                                                     │
                                                                     v
                                                   ┌──────────────────────────────┐
                                                   │ Sources (READ FIRST)         │
                                                   │ 1. spec.md                   │
                                                   │ 2. requirements.md           │
                                                   │ 3. data-model.md             │
                                                   │ 4. design-system.md (NEW)    │
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
