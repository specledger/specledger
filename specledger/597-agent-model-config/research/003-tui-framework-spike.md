# Research: TUI Framework Spike — Config Editor & Reusable Shell

**Date**: 2026-02-21
**Feature**: 597-agent-model-config
**Purpose**: Evaluate whether SpecLedger should build a reusable TUI shell (tree nav, expand/collapse, panes, inline editing) using Bubble Tea primitives, adopt an alternative framework (tview), or defer full TUI to a future spec.

---

## 1. Current State of SpecLedger TUI

### Codebase Analysis

| File | Lines | Pattern |
|------|-------|---------|
| `pkg/cli/tui/sl_new.go` | 551 | 7-step bootstrap wizard |
| `pkg/cli/tui/sl_init.go` | 363 | Dynamic-step init wizard |
| `pkg/cli/tui/terminal.go` | 156 | Mode detection + fallback prompts |
| **Total** | **1,070** | Step-based form wizards only |

### Key Findings

- **95% hand-rolled UI** — only `textinput.Model` from bubbles is used. Radio buttons (◉/○), checkboxes ([x]/[ ]), cursor indicators (›) are all manual string building.
- **No reusable components** — radio button rendering is duplicated 4 times across files. Styles are hardcoded globals. No theme abstraction.
- **Two architectural patterns**: `sl_new.go` uses hardcoded step constants (iota), `sl_init.go` uses a dynamic `[]InitStep` array. The array pattern is better and extensible.
- **Bubbles adoption: 1 of 20 available components** — `textinput` only. No list, table, viewport, help, or filepicker components used.
- **Refactoring potential: ~40-50%** of existing code could move to a generic wizard/form framework, reducing future wizard implementations from 300+ lines to ~100.

### What's Missing for a Config Editor

The current TUI only supports **sequential form flows** (step 1 → step 2 → ... → done). A config editor needs:

| Capability | Current TUI | Config Editor Needs |
|---|---|---|
| Persistent full-screen view | No (sequential steps) | Yes |
| Tree/section navigation | No | Yes (categories → keys) |
| Expand/collapse sections | No | Yes (agent.env map) |
| Inline value editing | Partial (textinput only) | Yes (string, bool, enum, map) |
| Scope indicators | No | Yes (global/local/personal) |
| Multi-pane layout | No | Yes (nav + editor) |

---

## 2. Bubble Tea Ecosystem — Available Components

### Official (charmbracelet)

| Component | Package | Interactive? | Useful For |
|---|---|---|---|
| `textinput` | bubbles | Yes | String value editing |
| `textarea` | bubbles | Yes | Multi-line editing |
| `list` | bubbles | Yes | Key/section navigation, filtering |
| `table` | bubbles | Yes | Key-value display (row selection only, no cell editing) |
| `viewport` | bubbles | Yes | Scrollable content pane |
| `help` | bubbles | Yes | Keybinding help bar |
| `filepicker` | bubbles | Yes | File path selection |
| `tree` | lipgloss/tree | **Render only** | Tree display (no keyboard nav, no expand/collapse interaction) |
| `huh` (v0.8.0) | huh | Yes | Form fields: Input, Select, Confirm, MultiSelect, FilePicker, Text |

**Key gap**: Lipgloss `tree` is a renderer, not an interactive widget. There is no official interactive tree component in bubbles. PR #639 (by dlvhdr) adds one, but it's blocked on bubbles v2.0.0 release.

### huh Form Library (already in go.mod)

`huh` v0.8.0 is already an indirect dependency. It provides:
- `Input` — single-line text (string config keys)
- `Select[T]` — single selection (enum config keys)
- `Confirm` — yes/no (bool config keys)
- `MultiSelect[T]` — multiple selection (string-list keys)
- `Text` — multi-line (with editor support)
- `FilePicker` — file selection
- Validation: `ValidateNotEmpty`, `ValidateOneOf`, custom validators
- Layout: `LayoutColumns()`, `LayoutGrid()`
- Themes: built-in + custom via lipgloss

**Limitation**: huh is a **form wizard** — it steps through groups of fields sequentially. It does not provide persistent navigation, tree views, or expand/collapse. It handles the *editing* part, not the *browsing* part.

### Community Tree Components

| Library | Stars | License | Active | Features |
|---|---|---|---|---|
| [Digital-Shane/treeview](https://github.com/Digital-Shane/treeview) | 79 | **GPL-3.0** | Yes (2025) | Full: search, filter, expand/collapse, viewport, lipgloss styling |
| [mariusor/bubbles-tree](https://github.com/mariusor/bubbles-tree) | 26 | MIT | Moderate (2025) | Interface-based nodes, expand/collapse, customizable symbols |
| [savannahostrowski/tree-bubble](https://github.com/savannahostrowski/tree-bubble) | 31 | MIT | No (2023) | Basic expand/collapse, keyboard nav |
| [mistakenelf/teacup](https://github.com/mistakenelf/teacup) | 263 | MIT | No (2023) | Filetree, statusbar, code viewer, markdown renderer |
| Official bubbles PR #639 | N/A | MIT | Pending v2.1.0 | Full tree widget — approved but not merged |

**Best MIT-licensed option today**: `mariusor/bubbles-tree` (interface-based, moderate maintenance).

### Community Layout Libraries

| Library | Stars | License | Pattern |
|---|---|---|---|
| [treilik/bubbleboxer](https://github.com/treilik/bubbleboxer) | 79 | MIT | Layout tree for side-by-side panes |
| [winder/bubblelayout](https://github.com/winder/bubblelayout) | 23 | MIT | Declarative grid with docking |
| [KevM/bubbleo](https://github.com/KevM/bubbleo) | 68 | — | Navigation stacks + breadcrumbs |

**Reality**: Most production Bubble Tea apps hand-roll layouts with `lipgloss.JoinHorizontal`/`JoinVertical` and manual dimension math.

---

## 3. Reference Implementations

### Best Examples of Complex Bubble Tea TUIs

| Project | Stars | Layout | Tree Nav | Architecture |
|---|---|---|---|---|
| [dlvhdr/diffnav](https://github.com/dlvhdr/diffnav) | 739 | 2-pane (tree sidebar + diff) | File tree with expand/collapse | Custom, MIT, active Feb 2026 |
| [leg100/pug](https://github.com/leg100/pug) | 663 | 3-pane (explorer + content + details) | Hierarchical module explorer | Custom, [excellent blog post](https://leg100.github.io/en/posts/building-bubbletea-programs/) |
| [mistakenelf/fm](https://github.com/mistakenelf/fm) | 628 | Multi-pane file manager | Filetree using teacup | Custom, active Dec 2025 |
| [dlvhdr/gh-dash](https://github.com/dlvhdr/gh-dash) | 10,200 | Multi-section dashboard | Tab/section nav | All custom with lipgloss |
| [charmbracelet/soft-serve](https://github.com/charmbracelet/soft-serve) | 6,600 | Multi-pane git browser | Repo + file browsing | First-party Charm |

### Key Architectural Pattern (from leg100's pug blog)

1. **Model tree**: Root model routes messages and composites views. Child models handle specific panes.
2. **Message routing**: Global (quit, help) → current model (active pane) → broadcast (WindowSizeMsg)
3. **Layout composition**: `lipgloss.JoinHorizontal`/`JoinVertical` with dynamic dimension calculation
4. **Navigation**: Stack-based (push/pop) or focus-based (tab between panes)
5. **Split panes**: A split model manages left/right panes with adjustable proportions

---

## 4. Alternative Framework: tview

[rivo/tview](https://github.com/rivo/tview) (13.3k stars, active, Aug 2025 release) is the only Go TUI framework with a **production-ready interactive TreeView** out of the box.

### tview Widget Set (relevant to config editor)

| Widget | Built-in? | Description |
|---|---|---|
| `TreeView` | Yes | Interactive tree with expand/collapse, selection callbacks, per-node styling |
| `Form` | Yes | Input fields, dropdowns, checkboxes, buttons |
| `Table` | Yes | Selectable cells, fixed rows/columns, virtual tables |
| `Flex` | Yes | Flexbox layout (horizontal/vertical) |
| `Grid` | Yes | CSS-grid-like layout |
| `Pages` | Yes | Tab/stack view switching |
| `List` | Yes | Selectable items with descriptions |
| `Modal` | Yes | Dialog overlays |

### Compatibility with Bubble Tea

**Cannot coexist in the same terminal session.** tview uses tcell and an imperative event loop; Bubble Tea uses its own renderer and MVU architecture. Options:

1. **Full migration to tview** — rewrite existing wizards. High effort, loses Bubble Tea ecosystem.
2. **Hybrid: tview for full-screen TUI, Bubble Tea for wizards** — tear down one framework before starting the other. Awkward but possible.
3. **Stay with Bubble Tea** — build custom tree/layout, more work upfront but consistent stack.

---

## 5. Effort Estimates

### Option A: Build Custom TUI Shell in Bubble Tea

Build a reusable `pkg/cli/tui/shell` package with tree navigation, pane management, and inline editing.

| Component | Effort | Reference |
|---|---|---|
| Interactive tree model (expand/collapse, cursor, keyboard) | 2-3 days | Based on mariusor/bubbles-tree pattern |
| Two-pane layout model (nav + editor, focus switching) | 1-2 days | Based on diffnav/pug pattern |
| Config key-value renderer (scope indicators, masking) | 1 day | Custom |
| Editor integration (huh forms per key type) | 1-2 days | huh embedding in Bubble Tea |
| String-map editor (add/remove/edit entries) | 1 day | Custom list + huh.Input |
| Theme system (extract from existing) | 0.5 day | Refactor |
| **Total for config editor** | **6-10 days** | |
| Refactor existing wizards to use shared components | 2-3 days | Quick win |
| **Total including wizard refactor** | **8-13 days** | |

**Reuse potential**: The tree nav, pane layout, and theme components are directly reusable for a revise TUI (comment list + file preview + action selection).

### Option B: Adopt tview for Full-Screen TUI

| Component | Effort | Notes |
|---|---|---|
| Config editor with TreeView + Form | 2-3 days | tview widgets are ready-made |
| Integration layer (tear down Bubble Tea, launch tview) | 1-2 days | Framework switching |
| Rewrite existing wizards to tview | 3-5 days | Optional, could keep Bubble Tea |
| **Total for config editor** | **3-5 days** | |
| **Total with full migration** | **6-10 days** | |

**Risk**: Maintaining two TUI frameworks increases complexity. tview community is smaller than Charm.

### Option C: Defer TUI, CLI-Only Config

| Component | Effort | Notes |
|---|---|---|
| `sl config set/get/show/unset` CLI subcommands | 2-3 days | Cobra commands only |
| `sl config profile` subcommands | 1-2 days | Cobra commands |
| Config merge/resolve logic | 2-3 days | Core feature |
| **Total** | **5-8 days** | |

**Tradeoff**: Fastest to ship. Users manage config entirely through CLI subcommands. TUI becomes a separate future spec that also covers the revise flow.

---

## 6. Recommendation

### Decision Matrix

| Criteria | Option A (Custom BT) | Option B (tview) | Option C (Defer) |
|---|---|---|---|
| Effort for config editor | 6-10 days | 3-5 days | 0 (no TUI) |
| Reuse for revise TUI | High | Medium (if tview) | None |
| Consistency with codebase | High (same stack) | Low (two stacks) | N/A |
| Widget richness | Medium (custom) | High (built-in) | N/A |
| Long-term maintenance | One stack | Two stacks | One stack |
| Ships core value fastest | No | No | **Yes** |

### Recommended: Option C now, Option A later

**Ship CLI-only config (Option C) in this spec. Create a separate TUI spec that covers both config editor and revise flow (Option A).**

Rationale:
1. The core value of 597-agent-model-config is **persistent config replacing shell aliases** — this is fully achievable with CLI subcommands alone (P1 + P2 user stories).
2. The TUI is explicitly P3 in the spec and was already flagged as needing a spike.
3. Building a reusable TUI shell is 6-10 days of work that benefits both config and revise — it deserves its own spec with proper user stories for both use cases.
4. Staying on Bubble Tea (not adopting tview) avoids framework fragmentation. The official bubbles tree component (PR #639) may land before the TUI spec is implemented.
5. The existing wizard TUI has ~40-50% refactoring potential that should be done as part of the TUI spec, not bolted onto a config feature.

### If TUI is built later, the recommended architecture is:

```
pkg/cli/tui/
├── shell/              # NEW: Reusable TUI shell
│   ├── tree.go         # Interactive tree model (expand/collapse)
│   ├── pane.go         # Two-pane layout (nav + content)
│   ├── editor.go       # Type-dispatched editor (huh integration)
│   ├── theme.go        # Shared styles and colors
│   └── shell.go        # Root model compositing tree + panes
├── config_editor.go    # Config-specific: key registry → tree, editing
├── revise_viewer.go    # Revise-specific: comment list, file preview
├── sl_new.go           # REFACTOR: Use shared components
└── sl_init.go          # REFACTOR: Use shared components
```

---

## 7. Action Items for 597-agent-model-config

1. **Remove User Story 3 (Interactive TUI) from this spec** — move to a new TUI spec
2. **Remove FR-008, FR-009, FR-020** (TUI-specific requirements)
3. **Remove SC-003** (TUI accessibility success criterion) or reword to CLI-only
4. **Keep all P1 and P2 stories** — CLI config, profiles, local/global hierarchy
5. **Create a new beads issue** for the future TUI spec covering config editor + revise flow
6. **Update plan.md** — remove TUI source files from project structure, remove huh from direct dependencies
