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

### 5. Mockup

Generated mockup artifact for a feature spec.

```go
type Mockup struct {
    SpecName      string         `yaml:"spec_name" json:"spec_name"`
    SpecPath      string         `yaml:"spec_path" json:"spec_path"`
    Generated     time.Time      `yaml:"generated" json:"generated"`
    Framework     FrameworkType  `yaml:"framework" json:"framework"`
    Screens       []Screen       `yaml:"screens" json:"screens"`
    ComponentMap  []ComponentRef `yaml:"component_map" json:"component_map"`
}

type Screen struct {
    Name        string `yaml:"name" json:"name"`
    Description string `yaml:"description,omitempty" json:"description,omitempty"`
    ASCII       string `yaml:"ascii" json:"ascii"`       // ASCII art representation
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
- `ASCII` must be valid ASCII art (no binary characters)

---

### 6. ScanResult

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

## File Locations

| Artifact | Path | Format |
|----------|------|--------|
| Design System Index | `specledger/design_system.md` | Markdown + YAML frontmatter |
| Mockup Output | `specledger/<spec-name>/mockup.md` | Markdown |
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
                                   | referenced by
                                   v
+------------------+       +------------------+
| Spec.md          |------>| Mockup           |
| (input)          |       | (generated)      |
+------------------+       +------------------+
                                   |
                                   | contains
                                   v
                           +------------------+
                           | Screen[]         |
                           | ComponentRef[]   |
                           +------------------+
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
