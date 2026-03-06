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
	FrameworkAstro     FrameworkType = "astro"
	FrameworkSolid     FrameworkType = "solid"
	FrameworkQwik      FrameworkType = "qwik"
	FrameworkRemix     FrameworkType = "remix"
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
		FrameworkAstro:     "Astro",
		FrameworkSolid:     "SolidJS",
		FrameworkQwik:      "Qwik",
		FrameworkRemix:     "Remix",
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

// DesignSystem contains the project's design tokens and styling conventions.
// Does NOT index components - the AI agent discovers those via codebase search.
type DesignSystem struct {
	Version      int           `yaml:"version" json:"version"`
	Framework    FrameworkType `yaml:"framework" json:"framework"`
	LastScanned  time.Time     `yaml:"last_scanned" json:"last_scanned"`
	ExternalLibs []string      `yaml:"external_libs,omitempty" json:"external_libs,omitempty"`
	Style        *StyleInfo    `yaml:"style,omitempty" json:"style,omitempty"`
	AppStructure *AppStructure `yaml:"app_structure,omitempty" json:"app_structure,omitempty"`
}

// AppStructure describes the project's layout and routing structure.
type AppStructure struct {
	Router       string   `yaml:"router,omitempty" json:"router,omitempty"`
	Layouts      []string `yaml:"layouts,omitempty" json:"layouts,omitempty"`
	Components   []string `yaml:"components,omitempty" json:"components,omitempty"`
	GlobalStyles []string `yaml:"global_styles,omitempty" json:"global_styles,omitempty"`
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
	CSSFramework    string            `yaml:"css_framework" json:"css_framework"`
	Preprocessor    string            `yaml:"preprocessor,omitempty" json:"preprocessor,omitempty"`
	StylingApproach string            `yaml:"styling_approach" json:"styling_approach"`
	ThemeColors     map[string]string `yaml:"theme_colors,omitempty" json:"theme_colors,omitempty"`
	FontFamilies    []string          `yaml:"font_families,omitempty" json:"font_families,omitempty"`
	CSSVariables    []string          `yaml:"css_variables,omitempty" json:"css_variables,omitempty"`
	SampleImports   []string          `yaml:"sample_imports,omitempty" json:"sample_imports,omitempty"`
	ComponentLibs   []string          `yaml:"component_libs,omitempty" json:"component_libs,omitempty"`
}

// MockupPromptContext is the template rendering context for the AI agent prompt.
type MockupPromptContext struct {
	SpecName   string        `json:"spec_name"`
	SpecPath   string        `json:"spec_path"`
	SpecTitle  string        `json:"spec_title"`
	Framework  FrameworkType `json:"framework"`
	Format     MockupFormat  `json:"format"`
	OutputPath string        `json:"output_path"`
	UserPrompt string        `json:"user_prompt,omitempty"`
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
	AgentLaunched       bool   `json:"agent_launched"`
	Committed           bool   `json:"committed"`
}

// UpdateResult is the JSON output for mockup update --json mode.
type UpdateResult struct {
	Status         string `json:"status"`
	ScanDurationMs int64  `json:"scan_duration_ms"`
}
