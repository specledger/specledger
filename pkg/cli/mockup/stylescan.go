package mockup

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// CSS framework detection patterns for package.json dependencies.
var cssFrameworkDeps = []struct {
	dep       string
	framework string
	approach  string
}{
	{"tailwindcss", "Tailwind CSS", "utility-first"},
	{"@tailwindcss/", "Tailwind CSS", "utility-first"},
	{"unocss", "UnoCSS", "utility-first"},
	{"@unocss/", "UnoCSS", "utility-first"},
	{"bootstrap", "Bootstrap", "utility-first"},
	{"@pandacss/dev", "Panda CSS", "utility-first"},
	{"@stylexjs/stylex", "StyleX", "css-in-js"},
	{"styled-components", "styled-components", "css-in-js"},
	{"@emotion/react", "Emotion", "css-in-js"},
	{"@emotion/styled", "Emotion", "css-in-js"},
	{"@stitches/react", "Stitches", "css-in-js"},
	{"@vanilla-extract/css", "Vanilla Extract", "css-in-js"},
	{"sass", "Sass/SCSS", "preprocessor"},
	{"less", "Less", "preprocessor"},
	{"lightningcss", "Lightning CSS", "preprocessor"},
}

// cssConfigFiles lists config files to detect CSS framework.
// Order matters: higher-priority frameworks listed first so they won't be
// overwritten by lower-priority ones (e.g. PostCSS is present alongside Tailwind).
var cssConfigFiles = []struct {
	file      string
	framework string
	stopIfSet bool // if true, stop scanning once a framework is already detected
}{
	{"tailwind.config.ts", "Tailwind CSS", false},
	{"tailwind.config.js", "Tailwind CSS", false},
	{"tailwind.config.mjs", "Tailwind CSS", false},
	{"uno.config.ts", "UnoCSS", false},
	{"uno.config.js", "UnoCSS", false},
	{"panda.config.ts", "Panda CSS", false},
	{"panda.config.mjs", "Panda CSS", false},
	{"postcss.config.js", "PostCSS", true},
	{"postcss.config.mjs", "PostCSS", true},
}

// componentLibDeps maps package.json dependencies to component library names.
var componentLibDeps = []struct {
	dep  string
	name string
}{
	{"@radix-ui/", "Radix UI"},
	{"@shadcn/ui", "shadcn/ui"},
	{"@mui/material", "MUI (Material UI)"},
	{"@mui/joy", "MUI Joy UI"},
	{"@chakra-ui/react", "Chakra UI"},
	{"antd", "Ant Design"},
	{"@ant-design/", "Ant Design"},
	{"@mantine/core", "Mantine"},
	{"@headlessui/react", "Headless UI"},
	{"@headlessui/vue", "Headless UI"},
	{"@nextui-org/react", "NextUI"},
	{"@heroicons/react", "Heroicons"},
	{"lucide-react", "Lucide Icons"},
	{"@phosphor-icons/react", "Phosphor Icons"},
	{"daisyui", "daisyUI"},
	{"flowbite", "Flowbite"},
	{"primereact", "PrimeReact"},
	{"primevue", "PrimeVue"},
	{"vuetify", "Vuetify"},
	{"element-plus", "Element Plus"},
	{"@ark-ui/react", "Ark UI"},
	{"@park-ui/", "Park UI"},
}

// tailwindColorRe matches simple hex/rgb color values in JS/TS object literals.
var tailwindColorRe = regexp.MustCompile(`'(#[0-9a-fA-F]{3,8}|rgba?\([^)]+\)|hsl[a]?\([^)]+\))'`)

// Global CSS file candidates to scan for variables and fonts.
var globalCSSCandidates = []string{
	"src/styles/globals.css",
	"src/styles/global.css",
	"src/app/globals.css",
	"app/globals.css",
	"styles/globals.css",
	"src/index.css",
	"src/styles.css",
	"src/global.css",
	"styles/global.css",
	"src/styles/variables.css",
	"src/styles/theme.css",
	"src/app/layout.css",
	"app/layout.css",
	"src/assets/css/main.css",
	"assets/css/main.css",
	"src/css/app.css",
}

var (
	cssVarDecl   = regexp.MustCompile(`--([a-zA-Z0-9_-]+)\s*:\s*([^;]+);`)
	cssColorVar  = regexp.MustCompile(`(?i)--(?:[a-z-]*(?:color|bg|background|foreground|primary|secondary|accent|border|muted|destructive|ring|card|popover)[a-z-]*)\s*:\s*([^;]+);`)
	cssFontDecl  = regexp.MustCompile(`(?i)font-family\s*:\s*([^;]+);`)
	styleImport  = regexp.MustCompile(`(?m)^import\s+.*(?:\.css|\.scss|\.less|\.module\.|styled|@emotion)`)
	tailwindBase = regexp.MustCompile(`@tailwind\s+base|@apply\s+`)
	// Tailwind v4 uses CSS-based config: @import "tailwindcss" and @theme blocks.
	tailwindV4Import = regexp.MustCompile(`@import\s+["']tailwindcss["']`)
	tailwindV4Theme  = regexp.MustCompile(`@theme\s*\{`)
)

// ScanStyles detects the project's CSS framework, variables, and styling patterns.
func ScanStyles(projectPath string) *StyleInfo {
	info := &StyleInfo{
		ThemeColors: make(map[string]string),
	}

	// 1. Check CSS config files (higher-priority entries first)
	for _, cfg := range cssConfigFiles {
		if cfg.stopIfSet && info.CSSFramework != "" {
			continue
		}
		if _, err := os.Stat(filepath.Join(projectPath, cfg.file)); err == nil {
			info.CSSFramework = cfg.framework
		}
	}

	// 1b. If Tailwind detected, parse its config to extract theme colors
	if info.CSSFramework == "Tailwind CSS" {
		scanTailwindConfig(projectPath, info)
	}

	// 1c. Check for Tailwind v4 CSS-based config (@import "tailwindcss", @theme)
	if info.CSSFramework == "" {
		for _, candidate := range globalCSSCandidates {
			fullPath := filepath.Join(projectPath, candidate)
			content, err := os.ReadFile(fullPath)
			if err != nil {
				continue
			}
			if tailwindV4Import.Match(content) || tailwindV4Theme.Match(content) {
				info.CSSFramework = "Tailwind CSS v4"
				info.StylingApproach = "utility-first"
				scanTailwindV4Theme(string(content), info)
				break
			}
		}
	}

	// 2. Check package.json deps for CSS framework and component libraries
	allDeps := readPackageDeps(projectPath)
	if allDeps != nil {
		for _, sig := range cssFrameworkDeps {
			for dep := range allDeps {
				if strings.HasPrefix(dep, sig.dep) || dep == sig.dep {
					if info.CSSFramework == "" {
						info.CSSFramework = sig.framework
					}
					if info.StylingApproach == "" {
						info.StylingApproach = sig.approach
					}
					switch sig.framework {
					case "Sass/SCSS":
						info.Preprocessor = "scss"
					case "Less":
						info.Preprocessor = "less"
					}
				}
			}
		}

		// Detect component libraries
		seen := make(map[string]bool)
		for _, lib := range componentLibDeps {
			for dep := range allDeps {
				if (strings.HasSuffix(lib.dep, "/") && strings.HasPrefix(dep, lib.dep)) || dep == lib.dep {
					if !seen[lib.name] {
						seen[lib.name] = true
						info.ComponentLibs = append(info.ComponentLibs, lib.name)
					}
				}
			}
		}

		// Detect shadcn/ui by components.json (shadcn doesn't always appear in deps)
		if _, err := os.Stat(filepath.Join(projectPath, "components.json")); err == nil {
			if !seen["shadcn/ui"] {
				info.ComponentLibs = append(info.ComponentLibs, "shadcn/ui")
			}
		}
	}

	// 3. Scan global CSS files for variables, colors, and fonts
	for _, candidate := range globalCSSCandidates {
		fullPath := filepath.Join(projectPath, candidate)
		if _, err := os.Stat(fullPath); err != nil {
			continue
		}
		scanCSSFile(fullPath, info)
	}

	// Also scan SCSS variable files
	for _, candidate := range []string{"src/styles/variables.scss", "src/styles/_variables.scss", "styles/variables.scss"} {
		fullPath := filepath.Join(projectPath, candidate)
		if _, err := os.Stat(fullPath); err != nil {
			continue
		}
		scanCSSFile(fullPath, info)
	}

	// 4. Detect styling approach from component files if not yet determined
	if info.StylingApproach == "" {
		info.StylingApproach = detectStylingApproach(projectPath)
	}

	// 5. Sample component imports for styling patterns
	info.SampleImports = sampleStyleImports(projectPath)

	// 6. Check for Tailwind directives in CSS files (confirms Tailwind usage)
	if info.CSSFramework == "" {
		for _, candidate := range globalCSSCandidates {
			fullPath := filepath.Join(projectPath, candidate)
			content, err := os.ReadFile(fullPath)
			if err != nil {
				continue
			}
			// Tailwind v4: @import "tailwindcss" or @theme {}
			if tailwindV4Import.Match(content) || tailwindV4Theme.Match(content) {
				info.CSSFramework = "Tailwind CSS v4"
				info.StylingApproach = "utility-first"
				break
			}
			// Tailwind v3: @tailwind base or @apply
			if tailwindBase.Match(content) {
				info.CSSFramework = "Tailwind CSS"
				info.StylingApproach = "utility-first"
				break
			}
		}
	}

	// Default approach if nothing detected
	if info.StylingApproach == "" && info.CSSFramework == "" {
		info.StylingApproach = "traditional"
	}

	return info
}

// scanTailwindConfig parses a tailwind.config.ts/js file and extracts theme color tokens.
// It uses regex-based extraction (no JS evaluation) to find color name/value pairs
// in the theme.extend.colors or theme.colors blocks.
func scanTailwindConfig(projectPath string, info *StyleInfo) {
	candidates := []string{"tailwind.config.ts", "tailwind.config.js", "tailwind.config.mjs"}
	var content []byte
	for _, name := range candidates {
		data, err := os.ReadFile(filepath.Join(projectPath, name))
		if err == nil {
			content = data
			break
		}
	}
	if len(content) == 0 {
		return
	}

	text := string(content)

	// Locate the first colors block (theme.colors or theme.extend.colors)
	colorsIdx := strings.Index(text, "colors:")
	if colorsIdx == -1 {
		colorsIdx = strings.Index(text, "colors :")
	}
	if colorsIdx == -1 {
		return
	}

	// Bounded window after "colors:" to avoid scanning the whole file
	window := text[colorsIdx:]
	if len(window) > 8000 {
		window = window[:8000]
	}

	// keyValueRe matches lines like:  500: '#6366f1',  or  brand: '#4f46e5',
	keyValueRe := regexp.MustCompile(`(\w+)\s*:\s*` + tailwindColorRe.String())
	// groupKeyRe matches object group openers like:  brand: {
	groupKeyRe := regexp.MustCompile(`^\s*(\w+)\s*:\s*\{`)
	// shadeRe matches Tailwind shade keys (numeric or DEFAULT)
	shadeRe := regexp.MustCompile(`^(\d+|DEFAULT)$`)

	var colorPrefix string
	for _, line := range strings.Split(window, "\n") {
		if g := groupKeyRe.FindStringSubmatch(line); len(g) > 1 {
			name := g[1]
			if name != "colors" && name != "extend" && name != "theme" {
				colorPrefix = name
			}
		}
		if kv := keyValueRe.FindStringSubmatch(line); len(kv) > 2 {
			key, val := kv[1], kv[2]
			var tokenName string
			if shadeRe.MatchString(key) && colorPrefix != "" {
				if key == "DEFAULT" {
					tokenName = colorPrefix
				} else {
					tokenName = colorPrefix + "-" + key
				}
			} else {
				tokenName = key
				colorPrefix = key
			}
			if _, exists := info.ThemeColors[tokenName]; !exists {
				info.ThemeColors[tokenName] = val
			}
		}
	}
}

// scanTailwindV4Theme extracts color tokens from Tailwind v4 @theme {} blocks.
// Tailwind v4 defines theme tokens as CSS custom properties inside @theme { ... }.
func scanTailwindV4Theme(content string, info *StyleInfo) {
	themeIdx := strings.Index(content, "@theme")
	if themeIdx == -1 {
		return
	}
	// Find the opening brace
	braceStart := strings.Index(content[themeIdx:], "{")
	if braceStart == -1 {
		return
	}
	start := themeIdx + braceStart + 1

	// Find matching closing brace
	depth := 1
	end := start
	for end < len(content) && depth > 0 {
		switch content[end] {
		case '{':
			depth++
		case '}':
			depth--
		}
		if depth > 0 {
			end++
		}
	}
	if depth != 0 {
		return
	}

	block := content[start:end]
	for _, line := range strings.Split(block, "\n") {
		if matches := cssVarDecl.FindStringSubmatch(strings.TrimSpace(line)); len(matches) > 2 {
			varName := "--" + matches[1]
			value := strings.TrimSpace(matches[2])
			if _, exists := info.ThemeColors[varName]; !exists {
				info.ThemeColors[varName] = value
			}
		}
	}
}

// scanCSSFile extracts CSS variables, colors, and font families from a CSS/SCSS file.
func scanCSSFile(path string, info *StyleInfo) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	varCount := 0
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Extract CSS custom properties (limit to avoid noise)
		if matches := cssVarDecl.FindStringSubmatch(line); len(matches) > 2 && varCount < 50 {
			varName := "--" + matches[1]
			info.CSSVariables = append(info.CSSVariables, varName)
			varCount++
		}

		// Extract color-related variables
		if matches := cssColorVar.FindStringSubmatch(line); len(matches) > 0 {
			fullMatch := cssVarDecl.FindStringSubmatch(line)
			if len(fullMatch) > 2 {
				info.ThemeColors["--"+fullMatch[1]] = strings.TrimSpace(fullMatch[2])
			}
		}

		// Extract font-family declarations
		if matches := cssFontDecl.FindStringSubmatch(line); len(matches) > 1 {
			font := strings.TrimSpace(matches[1])
			if !containsString(info.FontFamilies, font) {
				info.FontFamilies = append(info.FontFamilies, font)
			}
		}
	}
}

// detectStylingApproach samples component files to determine the CSS approach.
func detectStylingApproach(projectPath string) string {
	counts := map[string]int{
		"css-modules": 0,
		"css-in-js":   0,
		"utility":     0,
		"traditional": 0,
	}

	for _, dir := range commonComponentDirs {
		fullDir := filepath.Join(projectPath, dir)
		if _, err := os.Stat(fullDir); err != nil {
			continue
		}

		sampled := 0
		_ = filepath.WalkDir(fullDir, func(path string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				if d != nil && d.IsDir() && shouldSkipDir(d.Name()) {
					return filepath.SkipDir
				}
				return nil
			}
			if sampled >= 10 {
				return filepath.SkipAll
			}

			ext := filepath.Ext(d.Name())
			if ext != ".tsx" && ext != ".jsx" && ext != ".vue" && ext != ".svelte" && ext != ".astro" {
				return nil
			}

			content, readErr := os.ReadFile(path)
			if readErr != nil {
				return nil
			}
			text := string(content)
			sampled++

			if strings.Contains(text, ".module.css") || strings.Contains(text, ".module.scss") {
				counts["css-modules"]++
			}
			if strings.Contains(text, "styled.") || strings.Contains(text, "css`") || strings.Contains(text, "@emotion") || strings.Contains(text, "stylex.create") {
				counts["css-in-js"]++
			}
			// Utility-first: className with Tailwind-like classes, or class: with UnoCSS/Tailwind
			if (strings.Contains(text, "className=\"") || strings.Contains(text, "class=\"") || strings.Contains(text, "class:")) &&
				(strings.Contains(text, "flex ") || strings.Contains(text, "bg-") || strings.Contains(text, "text-") || strings.Contains(text, "p-") || strings.Contains(text, "m-")) {
				counts["utility"]++
			}
			if strings.Contains(text, "import './") || strings.Contains(text, "import \"./") {
				if strings.Contains(text, ".css") && !strings.Contains(text, ".module.css") {
					counts["traditional"]++
				}
			}
			// Scoped styles in Vue/Svelte/Astro
			if strings.Contains(text, "<style") && (strings.Contains(text, "scoped") || ext == ".svelte" || ext == ".astro") {
				counts["css-modules"]++ // scoped styles are similar to CSS modules
			}

			return nil
		})
	}

	best := ""
	bestCount := 0
	for approach, count := range counts {
		if count > bestCount {
			best = approach
			bestCount = count
		}
	}

	if best == "utility" {
		return "utility-first"
	}
	return best
}

// sampleStyleImports collects unique styling import patterns from component files.
func sampleStyleImports(projectPath string) []string {
	seen := make(map[string]struct{})
	var imports []string

	for _, dir := range commonComponentDirs {
		fullDir := filepath.Join(projectPath, dir)
		if _, err := os.Stat(fullDir); err != nil {
			continue
		}

		sampled := 0
		_ = filepath.WalkDir(fullDir, func(path string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				if d != nil && d.IsDir() && shouldSkipDir(d.Name()) {
					return filepath.SkipDir
				}
				return nil
			}
			if sampled >= 5 || len(imports) >= 5 {
				return filepath.SkipAll
			}

			ext := filepath.Ext(d.Name())
			if ext != ".tsx" && ext != ".jsx" && ext != ".vue" && ext != ".svelte" && ext != ".astro" {
				return nil
			}

			content, readErr := os.ReadFile(path)
			if readErr != nil {
				return nil
			}
			sampled++

			matches := styleImport.FindAllString(string(content), 3)
			for _, m := range matches {
				m = strings.TrimSpace(m)
				if _, ok := seen[m]; !ok && len(imports) < 5 {
					seen[m] = struct{}{}
					imports = append(imports, m)
				}
			}
			return nil
		})
	}

	return imports
}

func containsString(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
