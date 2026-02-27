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
	{"bootstrap", "Bootstrap", "utility-first"},
	{"styled-components", "styled-components", "css-in-js"},
	{"@emotion/react", "Emotion", "css-in-js"},
	{"@emotion/styled", "Emotion", "css-in-js"},
	{"@stitches/react", "Stitches", "css-in-js"},
	{"@vanilla-extract/css", "Vanilla Extract", "css-in-js"},
	{"sass", "Sass/SCSS", "preprocessor"},
	{"less", "Less", "preprocessor"},
}

// CSS config files to check for.
var cssConfigFiles = []struct {
	file      string
	framework string
}{
	{"tailwind.config.js", "Tailwind CSS"},
	{"tailwind.config.ts", "Tailwind CSS"},
	{"tailwind.config.mjs", "Tailwind CSS"},
	{"postcss.config.js", "PostCSS"},
	{"postcss.config.mjs", "PostCSS"},
}

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
}

var (
	cssVarDecl   = regexp.MustCompile(`--([a-zA-Z0-9_-]+)\s*:\s*([^;]+);`)
	cssColorVar  = regexp.MustCompile(`(?i)--(?:[a-z-]*(?:color|bg|background|foreground|primary|secondary|accent|border|muted|destructive|ring|card|popover)[a-z-]*)\s*:\s*([^;]+);`)
	cssFontDecl  = regexp.MustCompile(`(?i)font-family\s*:\s*([^;]+);`)
	styleImport  = regexp.MustCompile("(?m)^import\\s+.*(?:\\.css|\\.scss|\\.less|\\.module\\.|styled|@emotion)")
	tailwindBase = regexp.MustCompile(`@tailwind\s+base|@apply\s+`)
)

// ScanStyles detects the project's CSS framework, variables, and styling patterns.
func ScanStyles(projectPath string) *StyleInfo {
	info := &StyleInfo{
		ThemeColors: make(map[string]string),
	}

	// 1. Check CSS config files
	for _, cfg := range cssConfigFiles {
		if _, err := os.Stat(filepath.Join(projectPath, cfg.file)); err == nil {
			info.CSSFramework = cfg.framework
		}
	}

	// 2. Check package.json deps for CSS framework
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
					if sig.framework == "Sass/SCSS" {
						info.Preprocessor = "scss"
					} else if sig.framework == "Less" {
						info.Preprocessor = "less"
					}
				}
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

// scanCSSFile extracts CSS variables, colors, and font families from a CSS/SCSS file.
func scanCSSFile(path string, info *StyleInfo) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

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
			if ext != ".tsx" && ext != ".jsx" && ext != ".vue" && ext != ".svelte" {
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
			if strings.Contains(text, "styled.") || strings.Contains(text, "css`") || strings.Contains(text, "@emotion") {
				counts["css-in-js"]++
			}
			if strings.Contains(text, "className=\"") && (strings.Contains(text, "flex ") || strings.Contains(text, "bg-") || strings.Contains(text, "text-")) {
				counts["utility"]++
			}
			if strings.Contains(text, "import './") || strings.Contains(text, "import \"./") {
				if strings.Contains(text, ".css") && !strings.Contains(text, ".module.css") {
					counts["traditional"]++
				}
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
			if ext != ".tsx" && ext != ".jsx" {
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
