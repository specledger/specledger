package mockup

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Framework-specific glob patterns for component files.
var frameworkGlobs = map[FrameworkType][]string{
	FrameworkReact:     {"**/*.tsx", "**/*.jsx"},
	FrameworkNextJS:    {"**/*.tsx", "**/*.jsx"},
	FrameworkVue:       {"**/*.vue"},
	FrameworkNuxt:      {"**/*.vue"},
	FrameworkSvelte:    {"**/*.svelte"},
	FrameworkSvelteKit: {"**/*.svelte"},
	FrameworkAngular:   {"**/*.component.ts"},
}

// Component identification regex patterns per framework.
var (
	reactExportDefault  = regexp.MustCompile(`export\s+default\s+function\s+(\w+)`)
	reactExportConst    = regexp.MustCompile(`export\s+(?:const|function)\s+(\w+)`)
	reactPropsType      = regexp.MustCompile(`(?:Props|Props\s*=)\s*\{([^}]+)\}`)
	reactPropsInterface = regexp.MustCompile(`interface\s+\w*Props\s*\{([^}]+)\}`)

	vueDefineProps = regexp.MustCompile(`defineProps<\{([^}]+)\}>`)
	vuePropsOption = regexp.MustCompile(`props:\s*\{([^}]+)\}`)

	svelteExportLet = regexp.MustCompile(`export\s+let\s+(\w+)`)

	angularComponent = regexp.MustCompile(`@Component\(\{`)
	angularSelector  = regexp.MustCompile(`selector:\s*['"]([^'"]+)['"]`)
	angularInput     = regexp.MustCompile(`@Input\(\)\s+(\w+)`)
)

// externalLibPatterns maps import patterns to library names.
var externalLibPatterns = map[string]string{
	"@mui/material":     "@mui/material",
	"antd":              "antd",
	"@chakra-ui/react":  "@chakra-ui/react",
	"@headlessui/react": "@headlessui/react",
	"@radix-ui/react-":  "@radix-ui/react",
	"@mantine/core":     "@mantine/core",
}

// ScanComponents scans the project for UI components based on the detected framework.
func ScanComponents(projectPath string, framework FrameworkType) (*ScanResult, error) {
	start := time.Now()
	result := &ScanResult{
		Framework: framework,
	}

	exts := frameworkGlobs[framework]
	if len(exts) == 0 {
		// Fallback: scan for common frontend extensions
		exts = []string{"**/*.tsx", "**/*.jsx", "**/*.vue", "**/*.svelte"}
	}

	// Collect component dirs and scan files
	compDirSet := make(map[string]struct{})
	extLibSet := make(map[string]struct{})
	var scanErrors []ScanError
	filesScanned := 0

	for _, dir := range commonComponentDirs {
		fullDir := filepath.Join(projectPath, dir)
		if _, err := os.Stat(fullDir); os.IsNotExist(err) {
			continue
		}

		compDirSet[dir] = struct{}{}

		err := filepath.WalkDir(fullDir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return filepath.SkipDir
			}
			if d.IsDir() {
				if shouldSkipDir(d.Name()) {
					return filepath.SkipDir
				}
				return nil
			}

			if !matchesExtensions(d.Name(), exts) {
				return nil
			}

			filesScanned++

			relPath, _ := filepath.Rel(projectPath, path)
			comp, extLibs, scanErr := scanFile(path, relPath, framework)
			if scanErr != nil {
				scanErrors = append(scanErrors, ScanError{
					FilePath: relPath,
					Error:    scanErr.Error(),
				})
				return nil
			}

			if comp != nil {
				result.Components = append(result.Components, *comp)
			}
			for _, lib := range extLibs {
				extLibSet[lib] = struct{}{}
			}

			return nil
		})
		if err != nil {
			scanErrors = append(scanErrors, ScanError{
				FilePath: dir,
				Error:    err.Error(),
			})
		}
	}

	// Also scan external libraries from package.json
	pkgExtLibs := scanExternalLibraries(projectPath)
	for _, lib := range pkgExtLibs {
		extLibSet[lib] = struct{}{}
	}

	// Convert sets to slices
	for dir := range compDirSet {
		result.ComponentDirs = append(result.ComponentDirs, dir)
	}
	for lib := range extLibSet {
		result.ExternalLibs = append(result.ExternalLibs, lib)
	}

	result.FilesScanned = filesScanned
	result.Errors = scanErrors
	_ = time.Since(start) // scan duration tracked externally

	return result, nil
}

// scanFile extracts component information from a single file.
func scanFile(path, relPath string, framework FrameworkType) (*Component, []string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}

	text := string(content)
	var extLibs []string

	// Detect external library usage
	for pattern, lib := range externalLibPatterns {
		if strings.Contains(text, pattern) {
			extLibs = append(extLibs, lib)
		}
	}

	var comp *Component

	switch framework {
	case FrameworkReact, FrameworkNextJS:
		comp = scanReactComponent(text, relPath)
	case FrameworkVue, FrameworkNuxt:
		comp = scanVueComponent(text, relPath)
	case FrameworkSvelte, FrameworkSvelteKit:
		comp = scanSvelteComponent(text, relPath)
	case FrameworkAngular:
		comp = scanAngularComponent(text, relPath)
	default:
		// Try React patterns as default
		comp = scanReactComponent(text, relPath)
	}

	return comp, extLibs, nil
}

// scanReactComponent extracts component info from a React/Next.js file.
func scanReactComponent(content, relPath string) *Component {
	// Try export default function ComponentName
	if m := reactExportDefault.FindStringSubmatch(content); len(m) > 1 {
		comp := &Component{
			Name:     m[1],
			FilePath: relPath,
			Props:    extractReactProps(content),
		}
		comp.Description = inferDescription(comp.Name, relPath)
		return comp
	}

	// Try export const/function ComponentName
	if m := reactExportConst.FindStringSubmatch(content); len(m) > 1 {
		name := m[1]
		// Skip non-component exports (lowercase first letter, common utility names)
		if len(name) > 0 && name[0] >= 'A' && name[0] <= 'Z' {
			comp := &Component{
				Name:     name,
				FilePath: relPath,
				Props:    extractReactProps(content),
			}
			comp.Description = inferDescription(comp.Name, relPath)
			return comp
		}
	}

	// Fallback: use filename as component name if it contains JSX
	if strings.Contains(content, "return (") && (strings.Contains(content, "<div") || strings.Contains(content, "<>")) {
		name := componentNameFromPath(relPath)
		if name != "" {
			return &Component{
				Name:        name,
				FilePath:    relPath,
				Description: inferDescription(name, relPath),
			}
		}
	}

	return nil
}

// extractReactProps extracts props from React component content.
func extractReactProps(content string) []PropInfo {
	var props []PropInfo

	// Try interface Props pattern
	if m := reactPropsInterface.FindStringSubmatch(content); len(m) > 1 {
		props = parsePropsBlock(m[1])
	}

	// Try type Props pattern
	if len(props) == 0 {
		if m := reactPropsType.FindStringSubmatch(content); len(m) > 1 {
			props = parsePropsBlock(m[1])
		}
	}

	return props
}

// parsePropsBlock parses a TypeScript-style props block into PropInfo entries.
func parsePropsBlock(block string) []PropInfo {
	var props []PropInfo
	scanner := bufio.NewScanner(strings.NewReader(block))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}
		// Parse "name?: type" or "name: type"
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		name := strings.TrimSpace(parts[0])
		name = strings.TrimSuffix(name, "?")
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}

		propType := strings.TrimSpace(parts[1])
		propType = strings.TrimSuffix(propType, ";")
		propType = strings.TrimSuffix(propType, ",")
		propType = strings.TrimSpace(propType)

		required := !strings.Contains(parts[0], "?")

		props = append(props, PropInfo{
			Name:     name,
			Type:     propType,
			Required: required,
		})
	}
	return props
}

// scanVueComponent extracts component info from a Vue file.
func scanVueComponent(content, relPath string) *Component {
	if !strings.Contains(content, "<template>") && !strings.Contains(content, "<template ") {
		return nil
	}

	name := componentNameFromPath(relPath)
	if name == "" {
		return nil
	}

	comp := &Component{
		Name:     name,
		FilePath: relPath,
	}

	// Extract props
	if m := vueDefineProps.FindStringSubmatch(content); len(m) > 1 {
		comp.Props = parsePropsBlock(m[1])
	} else if m := vuePropsOption.FindStringSubmatch(content); len(m) > 1 {
		comp.Props = parsePropsBlock(m[1])
	}

	comp.Description = inferDescription(name, relPath)
	return comp
}

// scanSvelteComponent extracts component info from a Svelte file.
func scanSvelteComponent(content, relPath string) *Component {
	name := componentNameFromPath(relPath)
	if name == "" {
		return nil
	}

	comp := &Component{
		Name:     name,
		FilePath: relPath,
	}

	// Extract props from export let declarations
	matches := svelteExportLet.FindAllStringSubmatch(content, -1)
	for _, m := range matches {
		if len(m) > 1 {
			comp.Props = append(comp.Props, PropInfo{Name: m[1]})
		}
	}

	comp.Description = inferDescription(name, relPath)
	return comp
}

// scanAngularComponent extracts component info from an Angular component file.
func scanAngularComponent(content, relPath string) *Component {
	if !angularComponent.MatchString(content) {
		return nil
	}

	name := componentNameFromPath(relPath)
	if m := angularSelector.FindStringSubmatch(content); len(m) > 1 {
		// Convert selector to PascalCase name
		name = selectorToName(m[1])
	}
	if name == "" {
		return nil
	}

	comp := &Component{
		Name:     name,
		FilePath: relPath,
	}

	// Extract @Input() props
	matches := angularInput.FindAllStringSubmatch(content, -1)
	for _, m := range matches {
		if len(m) > 1 {
			comp.Props = append(comp.Props, PropInfo{Name: m[1]})
		}
	}

	comp.Description = inferDescription(name, relPath)
	return comp
}

// componentNameFromPath extracts a PascalCase component name from the file path.
func componentNameFromPath(relPath string) string {
	base := filepath.Base(relPath)
	// Remove extension(s)
	name := strings.TrimSuffix(base, filepath.Ext(base))
	// Remove .component suffix for Angular
	name = strings.TrimSuffix(name, ".component")

	if name == "" || name == "index" {
		// Use parent directory name
		dir := filepath.Dir(relPath)
		name = filepath.Base(dir)
	}

	// Convert to PascalCase
	return toPascalCase(name)
}

// toPascalCase converts a kebab-case or snake_case string to PascalCase.
func toPascalCase(s string) string {
	if s == "" {
		return ""
	}
	words := strings.FieldsFunc(s, func(r rune) bool {
		return r == '-' || r == '_' || r == '.'
	})
	var result strings.Builder
	for _, w := range words {
		if len(w) == 0 {
			continue
		}
		result.WriteString(strings.ToUpper(w[:1]))
		if len(w) > 1 {
			result.WriteString(w[1:])
		}
	}
	return result.String()
}

// selectorToName converts an Angular selector like "app-user-card" to "UserCard".
func selectorToName(selector string) string {
	// Remove "app-" prefix
	name := strings.TrimPrefix(selector, "app-")
	return toPascalCase(name)
}

// inferDescription generates a brief description based on the component name and path.
func inferDescription(name, relPath string) string {
	dir := filepath.Dir(relPath)
	parts := strings.Split(dir, string(filepath.Separator))
	for _, p := range parts {
		lower := strings.ToLower(p)
		switch lower {
		case "layout", "layouts":
			return "Layout component"
		case "forms", "form":
			return "Form component"
		case "common", "shared", "ui":
			return "Shared UI component"
		}
	}
	return name + " component"
}

// matchesExtensions checks if a filename matches any of the glob extensions.
func matchesExtensions(filename string, patterns []string) bool {
	for _, pattern := range patterns {
		ext := strings.TrimPrefix(pattern, "**/*")
		if strings.HasSuffix(filename, ext) {
			return true
		}
	}
	return false
}

// scanExternalLibraries checks package.json for known UI library dependencies.
func scanExternalLibraries(projectPath string) []string {
	allDeps := readPackageDeps(projectPath)
	if allDeps == nil {
		return nil
	}

	var libs []string
	for dep := range allDeps {
		for pattern, lib := range externalLibPatterns {
			if strings.HasPrefix(dep, pattern) {
				libs = append(libs, lib)
				break
			}
		}
	}

	return libs
}

// ScanResultToDesignSystem converts a ScanResult into a DesignSystem.
func ScanResultToDesignSystem(sr *ScanResult, framework FrameworkType) *DesignSystem {
	return &DesignSystem{
		Version:       1,
		Framework:     framework,
		LastScanned:   time.Now(),
		ComponentDirs: sr.ComponentDirs,
		ExternalLibs:  sr.ExternalLibs,
		Components:    sr.Components,
	}
}
