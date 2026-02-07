# FORK notes

Notes on SpecLedger's relationship to [github/spec-kit](https://github.com/github/spec-kit/).

## Overview

SpecLedger is a **thin wrapper** and project bootstrap tool that integrates with Spec Kit and other SDD frameworks. While not a direct fork, SpecLedger incorporates concepts and workflows from Spec Kit.

## Key Differences

### SpecLedger vs Spec Kit

- **SpecLedger**: Platform-agnostic CLI for project bootstrap and SDD framework setup
- **Spec Kit**: Comprehensive SDD framework with structured workflows

SpecLedger:
- Creates new projects and initializes them with Spec Kit or OpenSpec
- Manages dependencies between spec repositories
- Provides AI-friendly import paths for linked specs
- **Copies embedded templates** to new projects (Claude Code commands, skills, scripts)
- Delegates actual SDD workflows to the chosen framework

### Shared Concepts

SpecLedger adopts these Spec Kit concepts:
- **Phase-gated workflow**: spec → plan → tasks → implement
- **Beads for issue tracking**: Dependency-aware task management
- **AI-first design**: Optimized for Claude Code and other AI agents
- **Embedded templates**: Ships with Spec Kit playbook templates for immediate use

## Template Origin

The templates in `templates/` are adapted from Spec Kit with modifications:

### Task Generation Updates
- Updated prompt templates and `tasks.md` to use [beads](https://github.com/steelywing/beads) CLI for task management
- Updated analyze and implement prompts for the new task generation approach

### Task Tracking Updates
- Updated prompts to review previous work using beads queries
- Updated plan prompt to research previous work using beads

### Script Updates
- Modified scripts to accept arguments for branch short name and branch number
- Added adopt-feature-branch script to adopt existing feature branches
- Updated common.sh for feature branch mapping

## Embedded Templates

SpecLedger ships with embedded Spec Kit playbook templates:

- **Claude Code commands**: `.claude/commands/specledger.*.md`
- **Claude Code skills**: `.claude/skills/bd-issue-tracking/`
- **Helper scripts**: `specledger/scripts/bash/`
- **File templates**: `specledger/templates/`
- **Issue tracker**: `.beads/` configuration

These templates are **automatically copied** to new projects during `sl new` and `sl init`.

## Template Management

```bash
# List available embedded templates
sl template list

# Create project with templates (automatic)
sl new --framework speckit

# Templates are copied to:
#   - .claude/       (Claude Code integration)
#   - .beads/       (Issue tracker)
#   - specledger/   (Scripts and file templates)
```

## Integration

When you create a SpecLedger project with `sl new --framework speckit`:

1. SpecLedger creates the project structure
2. SpecLedger **copies embedded templates** to the project
3. SpecLedger installs Spec Kit via mise
4. SpecLedger runs `specify init --here --ai claude --force --script sh --no-git`
5. Your project is ready for Spec Kit workflows

## See Also

- [Spec Kit Repository](https://github.com/github/spec-kit/)
- [Beads Issue Tracker](https://github.com/steelywing/beads)
