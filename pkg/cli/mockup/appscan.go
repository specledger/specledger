package mockup

import (
	"os"
	"path/filepath"
	"strings"
)

// frameworkLayoutConfig defines where to look for layout files per framework.
type frameworkLayoutConfig struct {
	router    string
	dirs      []string
	filenames []string // exact filenames to match (e.g., "layout.tsx")
	filter    string   // substring filter for extension-based scan
	exts      []string
}

var frameworkLayoutConfigs = map[FrameworkType]frameworkLayoutConfig{
	FrameworkNextJS: {
		router:    "app-router",
		dirs:      []string{"app", "src/app"},
		filenames: []string{"layout.tsx", "layout.jsx", "layout.ts", "layout.js"},
		exts:      []string{".tsx", ".jsx", ".ts", ".js"},
	},
	FrameworkNuxt: {
		router: "file-based",
		dirs:   []string{"layouts"},
		exts:   []string{".vue"},
	},
	FrameworkSvelteKit: {
		router:    "file-based",
		dirs:      []string{"src/routes"},
		filenames: []string{"+layout.svelte", "+layout.ts", "+layout.server.ts"},
		exts:      []string{".svelte", ".ts"},
	},
	FrameworkAstro: {
		router: "file-based",
		dirs:   []string{"src/layouts"},
		exts:   []string{".astro"},
	},
	FrameworkRemix: {
		router:    "file-based",
		dirs:      []string{"app"},
		filenames: []string{"root.tsx", "root.jsx"},
		exts:      []string{".tsx", ".jsx"},
	},
	FrameworkReact: {
		router: "component-based",
		dirs:   []string{"src/layouts", "src/components/layouts"},
		filter: "layout",
		exts:   []string{".tsx", ".jsx"},
	},
	FrameworkVue: {
		router: "component-based",
		dirs:   []string{"src/layouts", "layouts"},
		exts:   []string{".vue"},
	},
	FrameworkSvelte: {
		router: "component-based",
		dirs:   []string{"src/layouts"},
		filter: "layout",
		exts:   []string{".svelte"},
	},
	FrameworkAngular: {
		router: "module-based",
		dirs:   []string{"src/app"},
		filter: "layout",
		exts:   []string{".ts"},
	},
	FrameworkSolid: {
		router: "component-based",
		dirs:   []string{"src/layouts"},
		filter: "layout",
		exts:   []string{".tsx", ".jsx"},
	},
	FrameworkQwik: {
		router:    "file-based",
		dirs:      []string{"src/routes"},
		filenames: []string{"layout.tsx"},
		exts:      []string{".tsx"},
	},
}

// ScanAppStructure detects the project's layout files based on the detected framework.
// Returns nil if no layouts are found.
func ScanAppStructure(projectPath string, framework FrameworkType) *AppStructure {
	cfg, ok := frameworkLayoutConfigs[framework]
	if !ok {
		return nil
	}

	app := &AppStructure{
		Router: cfg.router,
	}

	// Detect Next.js Pages Router vs App Router
	if framework == FrameworkNextJS {
		router := detectNextJSRouter(projectPath)
		app.Router = router
		if router == "pages-router" {
			cfg.dirs = []string{"pages", "src/pages"}
			cfg.filenames = []string{"_app.tsx", "_app.jsx", "_app.ts", "_app.js", "_document.tsx", "_document.jsx"}
		}
	}

	// Scan layouts
	if len(cfg.filenames) > 0 {
		app.Layouts = scanByFilename(projectPath, cfg.dirs, cfg.filenames)
	} else {
		app.Layouts = scanByExtension(projectPath, cfg.dirs, cfg.exts, cfg.filter)
	}

	// Scan component directories
	app.Components = scanComponents(projectPath)

	// Scan global stylesheets (css, scss)
	app.GlobalStyles = scanGlobalStyles(projectPath)

	if len(app.Layouts) == 0 && len(app.GlobalStyles) == 0 && len(app.Components) == 0 {
		return nil
	}

	return app
}

// scanComponents finds component files in common component directories.
// Only indexes 1 level deep to keep the tree concise.
func scanComponents(projectPath string) []string {
	componentExts := map[string]bool{
		".tsx": true, ".jsx": true, ".vue": true, ".svelte": true, ".astro": true,
	}

	var results []string
	seen := make(map[string]bool)

	for _, dir := range commonComponentDirs {
		fullDir := filepath.Join(projectPath, dir)
		entries, err := os.ReadDir(fullDir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() {
				// Include subdirectory name (e.g., "ui/") without listing contents
				rel := filepath.Join(dir, e.Name())
				if !seen[rel] && len(results) < 50 {
					seen[rel] = true
					results = append(results, rel+"/")
				}
				continue
			}
			if !componentExts[filepath.Ext(e.Name())] {
				continue
			}
			rel := filepath.Join(dir, e.Name())
			if !seen[rel] && len(results) < 50 {
				seen[rel] = true
				results = append(results, rel)
			}
		}
	}

	return results
}

// scanGlobalStyles finds top-level CSS/SCSS files in common locations.
func scanGlobalStyles(projectPath string) []string {
	candidates := []string{
		"src", "app", "styles", "css", "assets", "src/styles", "src/assets", "src/css",
	}

	styleExts := map[string]bool{
		".css": true, ".scss": true, ".sass": true, ".less": true,
	}

	var results []string
	seen := make(map[string]bool)

	// Check root-level style files first
	entries, err := os.ReadDir(projectPath)
	if err == nil {
		for _, e := range entries {
			if !e.IsDir() && styleExts[filepath.Ext(e.Name())] {
				if !seen[e.Name()] && len(results) < 30 {
					seen[e.Name()] = true
					results = append(results, e.Name())
				}
			}
		}
	}

	// Check candidate directories (1 level deep only)
	for _, dir := range candidates {
		fullDir := filepath.Join(projectPath, dir)
		entries, err := os.ReadDir(fullDir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			if !styleExts[filepath.Ext(e.Name())] {
				continue
			}
			rel := filepath.Join(dir, e.Name())
			if !seen[rel] && len(results) < 30 {
				seen[rel] = true
				results = append(results, rel)
			}
		}
	}

	return results
}

// detectNextJSRouter checks whether a Next.js project uses App Router or Pages Router.
func detectNextJSRouter(projectPath string) string {
	for _, dir := range []string{"app", "src/app"} {
		layoutPath := filepath.Join(projectPath, dir)
		if info, err := os.Stat(layoutPath); err == nil && info.IsDir() {
			for _, name := range []string{"layout.tsx", "layout.jsx", "layout.ts", "layout.js"} {
				if _, err := os.Stat(filepath.Join(layoutPath, name)); err == nil {
					return "app-router"
				}
			}
		}
	}
	for _, dir := range []string{"pages", "src/pages"} {
		if info, err := os.Stat(filepath.Join(projectPath, dir)); err == nil && info.IsDir() {
			return "pages-router"
		}
	}
	return "app-router"
}

// scanByFilename walks directories looking for specific filenames.
func scanByFilename(projectPath string, dirs []string, filenames []string) []string {
	nameSet := make(map[string]bool, len(filenames))
	for _, f := range filenames {
		nameSet[f] = true
	}

	var results []string
	seen := make(map[string]bool)

	for _, dir := range dirs {
		fullDir := filepath.Join(projectPath, dir)
		if _, err := os.Stat(fullDir); err != nil {
			continue
		}
		_ = filepath.WalkDir(fullDir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return filepath.SkipDir
			}
			if d.IsDir() {
				if shouldSkipDir(d.Name()) {
					return filepath.SkipDir
				}
				rel, _ := filepath.Rel(fullDir, path)
				if strings.Count(rel, string(filepath.Separator)) >= 5 {
					return filepath.SkipDir
				}
				return nil
			}
			if nameSet[d.Name()] {
				rel, _ := filepath.Rel(projectPath, path)
				if !seen[rel] && len(results) < 50 {
					seen[rel] = true
					results = append(results, rel)
				}
			}
			return nil
		})
	}
	return results
}

// scanByExtension walks directories looking for files with matching extensions.
func scanByExtension(projectPath string, dirs []string, exts []string, filter string) []string {
	extSet := make(map[string]bool, len(exts))
	for _, e := range exts {
		extSet[e] = true
	}

	var results []string
	seen := make(map[string]bool)

	for _, dir := range dirs {
		fullDir := filepath.Join(projectPath, dir)
		if _, err := os.Stat(fullDir); err != nil {
			continue
		}
		_ = filepath.WalkDir(fullDir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return filepath.SkipDir
			}
			if d.IsDir() {
				if shouldSkipDir(d.Name()) {
					return filepath.SkipDir
				}
				rel, _ := filepath.Rel(fullDir, path)
				if strings.Count(rel, string(filepath.Separator)) >= 5 {
					return filepath.SkipDir
				}
				return nil
			}
			if !extSet[filepath.Ext(d.Name())] {
				return nil
			}
			if filter != "" && !strings.Contains(strings.ToLower(d.Name()), filter) {
				return nil
			}
			rel, _ := filepath.Rel(projectPath, path)
			if !seen[rel] && len(results) < 50 {
				seen[rel] = true
				results = append(results, rel)
			}
			return nil
		})
	}
	return results
}
