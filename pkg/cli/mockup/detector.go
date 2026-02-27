package mockup

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// configSignal maps a config file to its framework and confidence level.
type configSignal struct {
	file       string
	framework  FrameworkType
	confidence int
}

// tier1Signals are definitive config file indicators (highest confidence).
var tier1Signals = []configSignal{
	{"next.config.js", FrameworkNextJS, 99},
	{"next.config.ts", FrameworkNextJS, 99},
	{"next.config.mjs", FrameworkNextJS, 99},
	{"angular.json", FrameworkAngular, 98},
	{".angular.json", FrameworkAngular, 98},
	{"svelte.config.js", FrameworkSvelteKit, 97},
	{"svelte.config.ts", FrameworkSvelteKit, 97},
	{"nuxt.config.js", FrameworkNuxt, 97},
	{"nuxt.config.ts", FrameworkNuxt, 97},
	{"remix.config.js", FrameworkReact, 97},
	{"remix.config.ts", FrameworkReact, 97},
}

// depSignal maps a package.json dependency to its framework.
type depSignal struct {
	dep       string
	framework FrameworkType
}

// tier2Signals are package.json dependency indicators.
var tier2Signals = []depSignal{
	{"next", FrameworkNextJS},
	{"@angular/core", FrameworkAngular},
	{"nuxt", FrameworkNuxt},
	{"svelte", FrameworkSvelte},
	{"react", FrameworkReact},
	{"react-dom", FrameworkReact},
	{"vue", FrameworkVue},
}

// extensionSignal maps a file extension to its framework.
type extensionSignal struct {
	ext       string
	framework FrameworkType
}

// tier3Signals are file extension fallbacks.
var tier3Signals = []extensionSignal{
	{".tsx", FrameworkReact},
	{".jsx", FrameworkReact},
	{".vue", FrameworkVue},
	{".svelte", FrameworkSvelte},
}

// commonComponentDirs are directories commonly containing UI components.
var commonComponentDirs = []string{
	"src/components",
	"components",
	"app/components",
}

// DetectFramework detects the frontend framework in the given project path.
// It uses a 3-tier heuristic: config files > package.json > file extensions.
func DetectFramework(projectPath string) (*DetectionResult, error) {
	result := &DetectionResult{
		Framework: FrameworkUnknown,
	}

	// Tier 1: Config files (definitive)
	for _, sig := range tier1Signals {
		configPath := filepath.Join(projectPath, sig.file)
		if _, err := os.Stat(configPath); err == nil {
			result.IsFrontend = true
			result.Framework = sig.framework
			result.Confidence = sig.confidence
			result.ConfigFile = sig.file
			result.Indicators = append(result.Indicators, "config: "+sig.file)
			result.ComponentDirs = detectComponentDirs(projectPath)
			return result, nil
		}
	}

	// Tier 1b: vite.config + .vue files = Vue (Vite)
	for _, viteCfg := range []string{"vite.config.js", "vite.config.ts"} {
		if _, err := os.Stat(filepath.Join(projectPath, viteCfg)); err == nil {
			// Check if there are .vue files
			hasVue := hasFileWithExtension(projectPath, ".vue")
			if hasVue {
				result.IsFrontend = true
				result.Framework = FrameworkVue
				result.Confidence = 95
				result.ConfigFile = viteCfg
				result.Indicators = append(result.Indicators, "config: "+viteCfg, "extension: .vue")
				result.ComponentDirs = detectComponentDirs(projectPath)
				return result, nil
			}
			// Vite without .vue could be React with Vite
			result.Indicators = append(result.Indicators, "config: "+viteCfg)
		}
	}

	// Tier 2: package.json dependencies
	if allDeps := readPackageDeps(projectPath); allDeps != nil {
		for _, sig := range tier2Signals {
			if _, ok := allDeps[sig.dep]; ok {
				result.IsFrontend = true
				result.Framework = sig.framework
				result.Confidence = 85
				result.ConfigFile = "package.json"
				result.Indicators = append(result.Indicators, "dependency: "+sig.dep)
				result.ComponentDirs = detectComponentDirs(projectPath)
				return result, nil
			}
		}
	}

	// Tier 3: File extension scan (last resort)
	for _, sig := range tier3Signals {
		if hasFileWithExtension(projectPath, sig.ext) {
			result.IsFrontend = true
			result.Framework = sig.framework
			result.Confidence = 70
			result.Indicators = append(result.Indicators, "extension: "+sig.ext)
			result.ComponentDirs = detectComponentDirs(projectPath)
			return result, nil
		}
	}

	// Angular component.ts check
	if hasFileWithExtension(projectPath, ".component.ts") {
		result.IsFrontend = true
		result.Framework = FrameworkAngular
		result.Confidence = 70
		result.Indicators = append(result.Indicators, "extension: .component.ts")
		result.ComponentDirs = detectComponentDirs(projectPath)
		return result, nil
	}

	return result, nil
}

// hasFileWithExtension checks if any file with the given extension exists
// in the project (up to 2 levels deep), skipping node_modules/vendor/.git.
func hasFileWithExtension(projectPath, ext string) bool {
	found := false
	_ = filepath.WalkDir(projectPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return filepath.SkipDir
		}
		if found {
			return filepath.SkipAll
		}
		if d.IsDir() {
			name := d.Name()
			if shouldSkipDir(name) {
				return filepath.SkipDir
			}
			// Limit scan depth to 3 levels
			rel, _ := filepath.Rel(projectPath, path)
			if strings.Count(rel, string(filepath.Separator)) >= 3 {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(d.Name(), ext) {
			found = true
			return filepath.SkipAll
		}
		return nil
	})
	return found
}

// readPackageDeps reads package.json and returns a merged map of dependencies and devDependencies.
func readPackageDeps(projectPath string) map[string]string {
	pkgPath := filepath.Join(projectPath, "package.json")
	data, err := os.ReadFile(pkgPath)
	if err != nil {
		return nil
	}

	var pkg struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil
	}

	allDeps := make(map[string]string, len(pkg.Dependencies)+len(pkg.DevDependencies))
	for k, v := range pkg.Dependencies {
		allDeps[k] = v
	}
	for k, v := range pkg.DevDependencies {
		allDeps[k] = v
	}
	return allDeps
}

// detectComponentDirs finds existing component directories in the project.
func detectComponentDirs(projectPath string) []string {
	var dirs []string
	for _, dir := range commonComponentDirs {
		fullPath := filepath.Join(projectPath, dir)
		if info, err := os.Stat(fullPath); err == nil && info.IsDir() {
			dirs = append(dirs, dir)
		}
	}
	return dirs
}

// shouldSkipDir returns true if the directory should be excluded from scanning.
func shouldSkipDir(name string) bool {
	skip := map[string]bool{
		"node_modules": true,
		"vendor":       true,
		".git":         true,
		"dist":         true,
		"build":        true,
		".next":        true,
		".nuxt":        true,
		"coverage":     true,
		".cache":       true,
		"ui":           true, // base/primitive UI components — skip
	}
	if skip[name] {
		return true
	}
	// Skip <name>-ui directories (e.g., shadcn-ui, radix-ui)
	if strings.HasSuffix(name, "-ui") {
		return true
	}
	return false
}
