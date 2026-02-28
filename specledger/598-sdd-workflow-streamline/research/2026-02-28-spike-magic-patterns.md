# Spike: Magic Patterns for Mockup Generation

**Date**: 2026-02-28
**Author**: so0k + Claude
**Status**: Complete
**Time-boxed**: 30 minutes
**Related**: `598-mockup-command` branch, `598-sdd-workflow-streamline` spec

---

## Research Question

Can [Magic Patterns](https://www.magicpatterns.com/) simplify or replace the custom mockup generation pipeline built in the `598-mockup-command` branch (~4,000 lines of Go code)?

## What is Magic Patterns?

Magic Patterns is a **SaaS AI prototype generator** that creates UI components from:
- Text prompts (natural language descriptions)
- Images (screenshots, wireframes)
- Figma designs (import/export)

It outputs **real code** (React, HTML/Tailwind, ShadCN, Chakra UI, Mantine) grounded in your project's design system.

### Key Capabilities

| Capability | How it works |
|------------|-------------|
| Design system import | Upload via Figma, Storybook, or GitHub sync |
| Component generation | Text prompt → React/HTML code with your components |
| Framework support | React, HTML/Tailwind, ShadCN, Chakra UI, Mantine |
| GitHub sync | Two-way sync between Magic Patterns and repo |
| MCP server | [ryanleecode/magic-patterns-mcp](https://github.com/ryanleecode/magic-patterns-mcp) — `create_design` tool |
| Figma integration | Bidirectional: import designs as code, export code back to Figma |
| Data connectors | Notion, Linear, PostHog, Granola (OAuth) |
| Output | Source files (React/HTML/CSS), compiled assets, preview URLs, editor links |

### MCP Server (`create_design` tool)

```json
{
  "prompt": "Create a pricing page using @PricingCard and @CTAButton",
  "mode": "best",       // or "fast"
  "presetId": "html-tailwind"  // or "shadcn", "chakra", "mantine"
}
```

Returns: design ID, editor URL, preview URL, source files, compiled assets.

## Current Mockup Branch (`598-mockup-command`)

The current implementation is ~4,000 lines of Go code:

| Package | Lines | Purpose |
|---------|-------|---------|
| `pkg/cli/mockup/detector.go` | ~242 | Frontend framework detection (React/Vue/Svelte/Angular) |
| `pkg/cli/mockup/scanner.go` | ~465 | Component scanning from source files |
| `pkg/cli/mockup/designsystem.go` | ~508 | Read/write `specledger/design_system.md` |
| `pkg/cli/mockup/stylescan.go` | ~317 | Style/theme scanning |
| `pkg/cli/mockup/specparser.go` | ~134 | Parse spec.md for mockup context |
| `pkg/cli/mockup/mockupprompt.go` | ~85 | Prompt template rendering |
| `pkg/cli/commands/mockup.go` | ~765 | TUI command with huh forms |
| `pkg/cli/prompt/` | ~107 | Shared editor/prompt utilities |

Flow: detect framework → scan components → build/read design_system.md → parse spec → build prompt → editor review → launch AI agent → commit/push.

## Comparison

| Aspect | Current `sl mockup` | Magic Patterns |
|--------|---------------------|----------------|
| Design system source | Scans codebase (Go code) | Import from Figma/Storybook/GitHub |
| Component detection | Custom Go scanner (~465 lines) | Built-in, supports popular libraries |
| Framework detection | Custom Go heuristics (~242 lines) | Preset selection (html-tailwind, shadcn, chakra, mantine) |
| Mockup generation | Delegates to AI agent (local) | Delegates to Magic Patterns API (cloud) |
| Output format | HTML or JSX file | React/HTML/CSS source + preview URL + editor |
| Integration | CLI only | MCP server, Figma, GitHub sync |
| Offline | Yes | No (SaaS) |
| Cost | Free (uses local AI agent) | Unknown (SaaS pricing not public) |
| Design system format | Custom markdown (`design_system.md`) | Magic Patterns internal format |
| Self-hosted | Yes | No |

## Findings

### Could Magic Patterns replace the custom pipeline?

**Partially, but not fully.** Here's why:

**What Magic Patterns does better:**
- Design system management is more mature (Figma/Storybook integration)
- Component library awareness is broader (catalog of popular React libraries)
- Output includes live preview URLs and an editor — better for iteration
- MCP server means it could be called from within an agent session directly

**What blocks full replacement:**
1. **SaaS dependency** — Magic Patterns is cloud-only. The `sl mockup` command works offline with local AI agents. Adding a SaaS dependency conflicts with SpecLedger's design principle of CLI-first, offline-capable tooling.
2. **Pricing unknown** — No public pricing for API/MCP access. Could be expensive at scale.
3. **Design system format lock-in** — Magic Patterns uses its own internal format. The current `design_system.md` is portable and human-readable.
4. **Framework coverage** — Magic Patterns focuses on React ecosystem (ShadCN, Chakra, Mantine). The current scanner supports Vue, Svelte, Angular too.
5. **No spec integration** — Magic Patterns takes free-form prompts. The `sl mockup` pipeline builds structured prompts from `spec.md` user stories + design system components. This spec-grounded approach is core to SpecLedger's value proposition.

### Could Magic Patterns complement the current pipeline?

**Yes.** Two integration points:

1. **MCP tool as alternative backend** — Instead of launching a local AI agent, `sl mockup` could offer `--backend magic-patterns` that calls the MCP `create_design` tool. The Go code still handles detection, scanning, prompt building — but delegates generation to Magic Patterns instead of a local agent.

2. **Design system import** — If a team already has their design system in Magic Patterns (via Figma/Storybook), `sl mockup` could import from Magic Patterns rather than scanning the codebase. This would be a new import source alongside the existing scanner.

## User Feedback on Current `sl mockup` Implementation

The initial implementor (Nathan/Ngoc) reported that the current `598-mockup-command` implementation (~4,000 lines of Go):
- **Didn't work well** — mockup quality was poor
- **Resource intensive** — the Go-side scanning/detection is heavyweight for what it produces
- Magic Patterns has **not been tried yet** as an alternative

This changes the calculus significantly. The question shifts from "can Magic Patterns replace a working pipeline?" to "can Magic Patterns provide a better foundation for a pipeline that isn't working well?"

## External Dependencies Introduced by Magic Patterns

| Dependency | Required? | Concern |
|------------|-----------|---------|
| Magic Patterns account | Yes | SaaS vendor lock-in |
| Internet connection | Yes | **Breaks offline-first principle** |
| Figma account | No (optional) | Only needed for Figma↔MP sync |
| Storybook | No (optional) | Only needed for Storybook↔MP import |
| GitHub OAuth | No (optional) | Only needed for GitHub↔MP two-way sync |
| API pricing | Unknown | No public pricing — cost risk at scale |

## Recommendations

| Option | Recommendation | Rationale |
|--------|---------------|-----------|
| Replace with Magic Patterns | **No** | SaaS dependency, offline requirement, unknown pricing |
| Add as optional backend | **Evaluate** | Current implementation isn't working well. MCP integration is low effort. Could be offered as `--backend magic-patterns` flag for teams that have accounts. |
| Rethink the approach | **Yes** | The current Go-heavy scanning approach (4k lines) produced poor results. Consider whether the problem is the scanning logic or the prompt quality. A simpler approach might be: minimal Go scaffolding → richer prompt template → let the AI agent do the heavy lifting (component discovery + mockup generation together). |
| Import design system from MP | **Defer** | Only relevant if teams adopt Magic Patterns separately |

### Key Insight

The current implementation puts too much logic in Go (framework detection, component scanning, style scanning) and too little in the AI agent prompt. The launcher pattern (D2) says CLI = data ops, AI = reasoning. Component discovery and mockup generation are both **reasoning tasks** — they should be in the agent's prompt, not in Go code. The Go side should only handle: detect spec context, check for existing design_system.md, build prompt, launch agent.

### Impact on spec/plan

The `598-mockup-command` spec should be revisited given the poor results feedback. Consider a v2 approach that:
1. Drastically reduces Go code (remove scanner/stylescan, simplify to prompt building)
2. Moves component discovery into the AI agent's responsibilities
3. Optionally supports Magic Patterns as a cloud backend via MCP

## Sources

- [Magic Patterns](https://www.magicpatterns.com/)
- [Magic Patterns Design Systems](https://www.magicpatterns.com/docs/documentation/design-systems/overview)
- [Magic Patterns Integrations](https://www.magicpatterns.com/docs/documentation/integrations/overview)
- [Magic Patterns MCP Server](https://github.com/ryanleecode/magic-patterns-mcp)
- [Magic Patterns Catalog (OSS)](https://github.com/magicpatterns/catalog)
- [Magic Patterns Guide](https://codeparrot.ai/blogs/magic-patterns-ai-a-complete-guide-tutorial)
