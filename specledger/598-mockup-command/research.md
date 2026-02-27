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

## Decision 2: Component Scanning Strategy

### Decision
Use glob patterns + regex content analysis per framework.

### Rationale
Each framework has distinct file patterns and component identification markers. A unified scanner with framework-specific handlers provides consistency while supporting variations.

### Scanning Patterns by Framework

**React/Next.js**

```text
Globs:
  src/components/**/*.{tsx,jsx}
  components/**/*.{tsx,jsx}
  app/components/**/*.{tsx,jsx}

Component Identification:
  - export default function ComponentName
  - export const ComponentName =
  - Return statement contains JSX (<div>, <Component>)
```

**Vue**

```text
Globs:
  src/components/**/*.vue
  components/**/*.vue

Component Identification:
  - <template> tag present
  - <script setup> or export default {}
  - defineProps<> for props extraction
```

**Svelte**

```text
Globs:
  src/components/**/*.svelte
  src/lib/**/*.svelte

Component Identification:
  - <script> tag present
  - export let propName for props
```

**Angular**

```text
Globs:
  src/app/**/*.component.ts

Component Identification:
  - @Component({ decorator
  - @Input() for props
  - selector: 'app-name' for component name
```

### Third-Party Library Detection

```text
Material UI:  import { X } from '@mui/material'
Ant Design:   import { X } from 'antd'
Chakra UI:    import { X } from '@chakra-ui/react'
Headless UI:  import { X } from '@headlessui/react'
Radix UI:     import { X } from '@radix-ui/react-*'
```

### Alternatives Considered

1. **AST Parsing** - Rejected: Complex, slow, framework-specific AST libraries needed
2. **Component Library Manifest** - Rejected: Not all projects have this
3. **Source Map Analysis** - Rejected: Requires build step

### Implementation

```go
// pkg/cli/mockup/scanner.go

type Component struct {
    Name        string
    FilePath    string
    Description string      // Auto-generated from file/context
    Props       []PropInfo  // Extracted props/inputs
    IsExternal  bool        // From third-party library
    Library     string      // e.g., "@mui/material"
}

type ScanResult struct {
    Components      []Component
    Framework       FrameworkType
    ComponentDirs   []string
    ExternalLibs    []string
}

func ScanComponents(projectPath string, framework FrameworkType) (*ScanResult, error)
```

---

## Decision 3: Design System Index Format

### Decision
Structured markdown with YAML frontmatter for machine readability.

### Rationale
Markdown is human-editable (requirement FR-010), YAML frontmatter enables programmatic parsing while keeping the document readable.

### Format

```markdown
---
version: 1
framework: react
last_scanned: 2026-02-27T10:30:00Z
component_dirs:
  - src/components
external_libs:
  - "@mui/material"
  - "@chakra-ui/react"
---

# Design System Index

## Project Components

### Button
- **Path**: `src/components/Button/Button.tsx`
- **Props**: `variant`, `size`, `disabled`, `onClick`
- **Description**: Primary action button with multiple variants

### Card
- **Path**: `src/components/Card/Card.tsx`
- **Props**: `title`, `children`, `elevated`
- **Description**: Container component for grouped content

## External Library Components

### @mui/material

Used components: `TextField`, `Dialog`, `Snackbar`, `Tooltip`

### @chakra-ui/react

Used components: `Box`, `Flex`, `Button`, `Input`
```

### Alternatives Considered

1. **Pure YAML** - Rejected: Less human-readable, harder to edit manually
2. **JSON** - Rejected: Not human-friendly for manual edits
3. **Custom DSL** - Rejected: Learning curve, no tooling support

---

## Decision 4: Mockup Output Format

### Decision
ASCII-style markdown mockups with component annotations.

### Rationale
Text-based mockups are version-controllable, diffable, and don't require external rendering. Component annotations link to design system index.

### Format

```markdown
# Mockup: User Registration Flow

**Spec**: `specledger/042-user-registration/spec.md`
**Generated**: 2026-02-27

## Screen 1: Registration Form

```text
+------------------------------------------+
|  HEADER [Navbar]                         |
+------------------------------------------+
|                                          |
|   Create Account                         |
|   ─────────────────                      |
|                                          |
|   Email                                  |
|   [TextField: email] ← @mui/material     |
|                                          |
|   Password                               |
|   [TextField: password]                  |
|                                          |
|   [Button: "Sign Up"] ← src/Button.tsx   |
|                                          |
|   Already have an account? [Link: Login] |
|                                          |
+------------------------------------------+
```

### Component Mapping

| UI Element | Design System Component | Source |
|------------|------------------------|--------|
| Email input | TextField | @mui/material |
| Password input | TextField | @mui/material |
| Submit button | Button | src/components/Button.tsx |
| Login link | Link | src/components/Link.tsx |

### User Flow

1. User enters email and password
2. User clicks "Sign Up"
3. System validates input (FR-003)
4. On success: redirect to dashboard
5. On error: display inline error messages

---

## Screen 2: Success Confirmation
...
```

### Alternatives Considered

1. **SVG/Image Generation** - Rejected: Requires rendering engine, not diffable
2. **HTML Mockups** - Rejected: Heavy for text-based workflow
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
