# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

**Note**: This project uses [bd (beads)](https://github.com/steveyegge/beads) for issue tracking. Use `bd` commands instead of markdown TODOs. See AGENTS.md for workflow details.

## Project Overview

This is a neobrutalist landing page for TerraConstructs, a CDKTF L2 Constructs library. The page showcases TypeScript, Go, and Python code examples with interactive syntax highlighting and live Terraform synthesis demos.

## Beads (Issue Tracking) Best Practices

### ⚠️ CRITICAL: Avoid Context Exhaustion with bd list

**NEVER run `bd list` without rigorous filtering or limits!** This will kill your context:

```bash
# ❌ DANGER: Even CLI can consume 5k-15k tokens without filters
bd list  # Lists ALL issues with full descriptions

# ✅ SAFE: Always use specific filters
bd search "database" --limit 10
bd list --status open --priority 1 --limit 5

# ✅ BETTER: Use targeted queries
bd ready --limit 5 # CLI: Find unblocked issues
bd show sl-xxxx  # CLI: View one issue
```

### Workflow Pattern

1. Create issue: `bd create` with full description
2. Update progress: `bd comments add`
3. Update status: `bd update` for status/priority changes
4. Close issue: `bd close` with reason

### Common Commands

```bash
bd show sl-xxxx && bd comments sl-xxxx # View issue details + comments
bd ready --limit 5                         # Find issues ready to work on
```

## Recent Changes
- 002-spec-dependency-linking: Added [if applicable, e.g., PostgreSQL, CoreData, files or N/A]
