# Research: Mockup Command

**Branch**: `598-mockup-command` | **Date**: 2026-02-27

## Prior Work

### Related Specs (from `sl issue list --all`)

| Spec | Relevance | Key Patterns to Reuse |
|------|-----------|----------------------|
| **597-issue-create-fields** | Most recent CLI command, establishes current patterns | Cobra command structure, flag handling, spec context detection |
| **011-streamline-onboarding** | Referenced in spec - extends onboarding flow | `sl init` integration point, TUI patterns |
| **591-issue-tracking-upgrade** | File-based storage patterns | JSONL storage patterns, `pkg/issues/context.go` for spec detection |
| **136-revise-comments** | Complex command with subpackage | `pkg/cli/revise/` pattern for domain logic separation |
| **596-doctor-version-update** | Recent command with file scanning | File traversal patterns |

### Existing Codebase Patterns

**Command Structure** (`pkg/cli/commands/`):
- Main command file: `mockup.go` with `VarMockupCmd`
- Subcommands: `mockup <spec-name>`, `mockup update`
- Registration in `cmd/sl/main.go`
- Domain logic in separate package: `pkg/cli/mockup/`

**Spec Context Detection** (`pkg/issues/context.go`):
- `NewContextDetector(".")` for detecting current spec from git branch
- Branch pattern: `###-feature-name` extracts spec name

**Init Integration** (`pkg/cli/commands/bootstrap.go`):
- `setupSpecLedgerProject()` function handles initialization
- Can add frontend detection and design system generation here

---

## Decision 1: Frontend Framework Detection

### Decision
Implement tiered detection with config files as primary signals.

### Rationale
Config files have >95% accuracy and are mandatory for frameworks to function. Package.json analysis is fallback for edge cases.

### Detection Heuristics (Priority Order)

**Tier 1: Framework Config Files (Definitive)**

| Framework | Config File(s) | Confidence |
|-----------|----------------|------------|
| Next.js | `next.config.js`, `next.config.ts`, `next.config.mjs` | 99% |
| Angular | `angular.json`, `.angular.json` | 98% |
| SvelteKit | `svelte.config.js` | 97% |
| Nuxt | `nuxt.config.js`, `nuxt.config.ts` | 97% |
| Remix | `remix.config.js`, `remix.config.ts` | 97% |
| Vue (Vite) | `vite.config.js` + `.vue` files | 95% |

**Tier 2: Package.json Dependencies (Fallback)**

```json
{
  "react": "^18.x",        // React project
  "react-dom": "^18.x",    // React project
  "vue": "^3.x",           // Vue project
  "@angular/core": "^17.x", // Angular project
  "svelte": "^4.x"         // Svelte project
}
```

**Tier 3: File Extension Scan (Last Resort)**

| Extension | Framework |
|-----------|-----------|
| `.tsx`, `.jsx` | React |
| `.vue` | Vue |
| `.svelte` | Svelte |
| `.component.ts` | Angular |

### Alternatives Considered

1. **AST Parsing** - Rejected: Too slow for initial detection, overkill for config file checks
2. **External API** - Rejected: Violates offline-capable constraint
3. **User Configuration Only** - Rejected: Adds setup friction, contradicts auto-detection goal

### Implementation

```go
// pkg/cli/mockup/detector.go

type FrameworkType string

const (
    FrameworkReact    FrameworkType = "react"
    FrameworkNextJS   FrameworkType = "nextjs"
    FrameworkVue      FrameworkType = "vue"
    FrameworkNuxt     FrameworkType = "nuxt"
    FrameworkSvelte   FrameworkType = "svelte"
    FrameworkAngular  FrameworkType = "angular"
    FrameworkUnknown  FrameworkType = "unknown"
)

type DetectionResult struct {
    IsFrontend   bool
    Framework    FrameworkType
    Confidence   int           // 0-100
    ComponentDir string        // e.g., "src/components"
    Indicators   []string      // What was detected
}

func DetectFramework(projectPath string) (*DetectionResult, error)
```

---

## Decision 2: CSS Token Extraction Strategy

### Decision
Extract global CSS tokens (colors, fonts, variables) instead of scanning individual components. The AI agent discovers and uses components via codebase search.

### Rationale
Component scanning is complex and error-prone (different patterns per framework, nested components, dynamic imports). The AI agent is better equipped to discover components via codebase search since it understands code context. The design system should focus on design tokens (colors, typography, spacing) that define the project's visual identity.

### CSS Token Sources

**Tailwind CSS**
```text
Files:
  tailwind.config.js
  tailwind.config.ts

Extract:
  - theme.colors
  - theme.fontFamily
  - theme.extend.colors
```

**CSS Variables**
```text
Files:
  src/index.css
  src/globals.css
  src/styles/*.css

Extract:
  --color-primary
  --color-secondary
  --font-family-*
  --spacing-*
```

**CSS-in-JS Themes**
```text
Files:
  src/theme.ts
  src/styles/theme.js

Detect:
  - styled-components ThemeProvider
  - emotion ThemeProvider
  - Chakra UI theme
```

### Implementation

```go
// pkg/cli/mockup/stylescan.go

type StyleInfo struct {
    CSSFramework    string            // e.g., "Tailwind CSS"
    Preprocessor    string            // e.g., "sass"
    StylingApproach string            // "utility-first", "css-in-js", "css-modules"
    ThemeColors     map[string]string // Extracted color tokens
    FontFamilies    []string          // Extracted font families
    CSSVariables    []string          // Extracted CSS custom properties
}

func ScanStyles(projectPath string) *StyleInfo
```

### Alternatives Considered

1. **Full Component Scanning** - Rejected: Complex, error-prone, AI agent does this better
2. **Design Token JSON Export** - Rejected: Adds tooling requirement
3. **Manual Configuration Only** - Rejected: Too much friction

---

## Decision 3: Design System Format

### Decision
Structured markdown with YAML frontmatter containing CSS tokens only. No component indexing — the AI agent discovers components via codebase search.

### Rationale
Markdown is human-editable (requirement FR-010), YAML frontmatter enables programmatic parsing. Focusing on CSS tokens (colors, fonts, variables) rather than component indexing keeps the design system lightweight and avoids complex scanning logic.

### Format

```markdown
---
version: 1
framework: react
last_scanned: 2026-02-27T10:30:00Z
external_libs:
  - "@mui/material"
  - "@chakra-ui/react"
style:
  css_framework: Tailwind CSS
  styling_approach: utility-first
  theme_colors:
    primary: "#3b82f6"
    secondary: "#64748b"
  font_families:
    - Inter
    - system-ui
  css_variables:
    - "--color-primary"
    - "--spacing-lg"
---

# Design System

This project uses **Tailwind CSS** with a utility-first approach.

## Theme Colors

| Name | Value |
|------|-------|
| primary | #3b82f6 |
| secondary | #64748b |

## Typography

- **Primary Font**: Inter
- **Fallback**: system-ui

## Notes

The AI agent will search the codebase to discover and use existing components.
Add any manual styling notes or conventions below.
```

### Alternatives Considered

1. **Component Indexing** - Rejected: Complex scanning, AI agent does this better
2. **Pure YAML** - Rejected: Less human-readable
3. **JSON** - Rejected: Not human-friendly for manual edits

---

## Decision 4: Mockup Output Format

### Decision
HTML or JSX mockup files with component annotations, stored at `specledger/<spec-name>/mockup.html` or `mockup.jsx`.

### Rationale
HTML provides an immediately previewable format that can be opened in any browser. JSX provides React-compatible component code that developers can directly reference or adapt. Both formats are version-controllable, diffable, and support component annotations linking to the design system index. A `--format` flag (`html` default, `jsx` optional) lets the user choose.

### Format — HTML (default)

```html
<!-- specledger/042-user-registration/mockup.html -->
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Mockup: User Registration Flow</title>
  <style>
    /* Inline styles for self-contained preview */
    body { font-family: system-ui, sans-serif; margin: 0; padding: 20px; }
    .screen { border: 1px solid #ccc; padding: 24px; margin-bottom: 32px; max-width: 480px; }
    .component-ref { color: #666; font-size: 12px; }
  </style>
</head>
<body>
  <h1>Mockup: User Registration Flow</h1>
  <p>Spec: <code>specledger/042-user-registration/spec.md</code></p>
  <p>Generated: 2026-02-27</p>

  <!-- Screen 1: Registration Form -->
  <section class="screen">
    <header><!-- Navbar → src/components/Navbar.tsx --></header>
    <h2>Create Account</h2>
    <label>Email</label>
    <input type="email" placeholder="email" /> <!-- TextField → @mui/material -->
    <label>Password</label>
    <input type="password" placeholder="password" /> <!-- TextField → @mui/material -->
    <button>Sign Up</button> <!-- Button → src/components/Button.tsx -->
    <p>Already have an account? <a href="#">Login</a></p> <!-- Link → src/components/Link.tsx -->
  </section>

  <!-- Component Mapping -->
  <!--
  | UI Element      | Component  | Source                      |
  |-----------------|------------|-----------------------------|
  | Email input     | TextField  | @mui/material               |
  | Password input  | TextField  | @mui/material               |
  | Submit button   | Button     | src/components/Button.tsx    |
  | Login link      | Link       | src/components/Link.tsx      |
  -->

  <!-- User Flow:
  1. User enters email and password
  2. User clicks "Sign Up"
  3. System validates input (FR-003)
  4. On success: redirect to dashboard
  5. On error: display inline error messages
  -->
</body>
</html>
```

### Format — JSX (with `--format jsx`)

```jsx
// specledger/042-user-registration/mockup.jsx
// Mockup: User Registration Flow
// Spec: specledger/042-user-registration/spec.md
// Generated: 2026-02-27

import React from "react";
// Component references from design system:
// TextField → @mui/material
// Button → src/components/Button.tsx
// Link → src/components/Link.tsx

export default function MockupUserRegistration() {
  return (
    <div style={{ fontFamily: "system-ui, sans-serif", padding: 20 }}>
      <h1>User Registration Flow</h1>

      {/* Screen 1: Registration Form */}
      <section style={{ border: "1px solid #ccc", padding: 24, maxWidth: 480 }}>
        {/* Navbar → src/components/Navbar.tsx */}
        <header />
        <h2>Create Account</h2>
        <label>Email</label>
        {/* TextField → @mui/material */}
        <input type="email" placeholder="email" />
        <label>Password</label>
        {/* TextField → @mui/material */}
        <input type="password" placeholder="password" />
        {/* Button → src/components/Button.tsx */}
        <button>Sign Up</button>
        <p>
          Already have an account? <a href="#">Login</a>
          {/* Link → src/components/Link.tsx */}
        </p>
      </section>

      {/* User Flow:
        1. User enters email and password
        2. User clicks "Sign Up"
        3. System validates input (FR-003)
        4. On success: redirect to dashboard
        5. On error: display inline error messages
      */}
    </div>
  );
}
```

### Alternatives Considered

1. **ASCII Markdown** - Rejected: Not previewable in browser, less useful for frontend developers
2. **SVG/Image Generation** - Rejected: Requires rendering engine, not diffable
3. **Figma Export** - Rejected: External dependency, out of scope

---

## Decision 5: Init Integration Approach

### Decision
Add frontend detection to `setupSpecLedgerProject()` function with conditional design system generation.

### Rationale
Minimal changes to existing flow. Detection happens after base initialization, design system is created only if frontend detected.

### Integration Point

```go
// In pkg/cli/commands/bootstrap.go
func setupSpecLedgerProject(...) {
    // ... existing setup code ...

    // NEW: Frontend detection and design system init
    if result, _ := mockup.DetectFramework(projectPath); result.IsFrontend {
        if err := mockup.InitializeDesignSystem(projectPath, result); err != nil {
            // Log warning but don't fail init
            l.Warn("Could not initialize design system: %v", err)
        }
    }
}
```

---

## External Dependencies

No new external Go dependencies required. Implementation uses:
- Standard library: `path/filepath`, `os`, `regexp`, `strings`
- Existing: `gopkg.in/yaml.v3` for YAML parsing

---

## Open Questions (Resolved)

| Question | Resolution |
|----------|------------|
| How to handle monorepos? | Detect frontend in current working directory, flag if ambiguous |
| Performance for large codebases? | Limit scan depth, skip node_modules/vendor, cache results |
| Manual edits preservation? | Use markers `<!-- MANUAL -->` ... `<!-- /MANUAL -->` for preserved sections |
