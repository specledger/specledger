package mockup

import "time"

// FrameworkType identifies the frontend framework in use.
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

// String returns the display name for the framework.
func (f FrameworkType) String() string {
	names := map[FrameworkType]string{
		FrameworkReact:     "React",
		FrameworkNextJS:    "Next.js",
		FrameworkVue:       "Vue",
		FrameworkNuxt:      "Nuxt",
		FrameworkSvelte:    "Svelte",
		FrameworkSvelteKit: "SvelteKit",
		FrameworkAngular:   "Angular",
		FrameworkUnknown:   "Unknown",
	}
	if name, ok := names[f]; ok {
		return name
	}
	return string(f)
}

// MockupFormat is the output format for the generated mockup file.
type MockupFormat string

const (
	MockupFormatHTML MockupFormat = "html"
	MockupFormatJSX  MockupFormat = "jsx"
)

// IsValid returns true if the format is supported.
func (f MockupFormat) IsValid() bool {
	return f == MockupFormatHTML || f == MockupFormatJSX
}

// DetectionResult is the result of frontend framework detection.
type DetectionResult struct {
	IsFrontend    bool          `json:"is_frontend"`
	Framework     FrameworkType `json:"framework"`
	Confidence    int           `json:"confidence"`
	ComponentDirs []string      `json:"component_dirs"`
	Indicators    []string      `json:"indicators"`
	ConfigFile    string        `json:"config_file"`
}

// PropInfo describes a single prop/input of a component.
type PropInfo struct {
	Name     string `yaml:"name" json:"name"`
	Type     string `yaml:"type,omitempty" json:"type,omitempty"`
	Required bool   `yaml:"required,omitempty" json:"required,omitempty"`
}

// Component is a single UI component discovered in the codebase.
type Component struct {
	Name        string     `yaml:"name" json:"name"`
	FilePath    string     `yaml:"path" json:"path"`
	Description string     `yaml:"description,omitempty" json:"description,omitempty"`
	Props       []PropInfo `yaml:"props,omitempty" json:"props,omitempty"`
	IsExternal  bool       `yaml:"external,omitempty" json:"external,omitempty"`
	Library     string     `yaml:"library,omitempty" json:"library,omitempty"`
}

// DesignSystem is the full design system index for a project.
type DesignSystem struct {
	Version       int           `yaml:"version" json:"version"`
	Framework     FrameworkType `yaml:"framework" json:"framework"`
	LastScanned   time.Time     `yaml:"last_scanned" json:"last_scanned"`
	ComponentDirs []string      `yaml:"component_dirs" json:"component_dirs"`
	ExternalLibs  []string      `yaml:"external_libs,omitempty" json:"external_libs,omitempty"`
	Components    []Component   `yaml:"components" json:"components"`
	ManualEntries []Component   `yaml:"manual_entries,omitempty" json:"manual_entries,omitempty"`
}

// ScanResult is the result of a component scanning operation.
type ScanResult struct {
	Components    []Component   `json:"components"`
	Framework     FrameworkType `json:"framework"`
	ComponentDirs []string      `json:"component_dirs"`
	ExternalLibs  []string      `json:"external_libs"`
	FilesScanned  int           `json:"files_scanned"`
	Errors        []ScanError   `json:"errors,omitempty"`
}

// ScanError represents a non-fatal scan issue.
type ScanError struct {
	FilePath string `json:"file_path"`
	Error    string `json:"error"`
}

// SpecContent is parsed content from a spec.md file.
type SpecContent struct {
	Title        string   `json:"title"`
	UserStories  []string `json:"user_stories"`
	Requirements []string `json:"requirements"`
	FullContent  string   `json:"full_content"`
}

// StyleInfo describes the project's CSS/styling patterns.
type StyleInfo struct {
	CSSFramework    string            `json:"css_framework"`
	Preprocessor    string            `json:"preprocessor,omitempty"`
	StylingApproach string            `json:"styling_approach"`
	ThemeColors     map[string]string `json:"theme_colors,omitempty"`
	FontFamilies    []string          `json:"font_families,omitempty"`
	CSSVariables    []string          `json:"css_variables,omitempty"`
	SampleImports   []string          `json:"sample_imports,omitempty"`
}

// PromptComponent is a component entry included in the agent prompt.
type PromptComponent struct {
	Name        string `json:"name"`
	FilePath    string `json:"path"`
	Description string `json:"description,omitempty"`
	Props       string `json:"props,omitempty"`
	IsExternal  bool   `json:"external,omitempty"`
	Library     string `json:"library,omitempty"`
}

// MockupPromptContext is the template rendering context for the AI agent prompt.
type MockupPromptContext struct {
	SpecName        string            `json:"spec_name"`
	SpecPath        string            `json:"spec_path"`
	SpecTitle       string            `json:"spec_title"`
	Framework       FrameworkType     `json:"framework"`
	Format          MockupFormat      `json:"format"`
	OutputDir       string            `json:"output_dir"`
	Components      []PromptComponent `json:"components"`
	ComponentTree   string            `json:"component_tree,omitempty"`
	ExternalLibs    []string          `json:"external_libs,omitempty"`
	HasDesignSystem bool              `json:"has_design_system"`
	Style           *StyleInfo        `json:"style,omitempty"`
	HasStyle        bool              `json:"has_style"`
}

// MockupResult is the JSON output for --json mode.
type MockupResult struct {
	Status              string `json:"status"`
	Framework           string `json:"framework"`
	SpecName            string `json:"spec_name"`
	MockupPath          string `json:"mockup_path"`
	PromptPath          string `json:"prompt_path,omitempty"`
	Format              string `json:"format"`
	DesignSystemCreated bool   `json:"design_system_created"`
	ComponentsScanned   int    `json:"components_scanned"`
	ComponentsSelected  int    `json:"components_selected"`
	AgentLaunched       bool   `json:"agent_launched"`
	Committed           bool   `json:"committed"`
}

// UpdateResult is the JSON output for mockup update --json mode.
type UpdateResult struct {
	Status            string `json:"status"`
	ComponentsTotal   int    `json:"components_total"`
	ComponentsAdded   int    `json:"components_added"`
	ComponentsRemoved int    `json:"components_removed"`
	ScanDurationMs    int64  `json:"scan_duration_ms"`
}
