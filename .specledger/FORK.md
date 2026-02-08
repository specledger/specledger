# FORK notes

Notes on changes made to local files compared to upstream [github/spec-kit](https://github.com/github/spec-kit/).

## Directory Structure Changes

- **`.specify/` → `.specledger/`**: Renamed the configuration directory from `.specify/` to `.specledger/` to reflect the project branding.
- **`specs/` → `specledger/`**: Renamed the specifications directory from `specs/` to `specledger/` for consistency.

## Task Generation Updates

- Updated `speckit.tasks` prompt template and `tasks.md` template to use [beads](https://github.com/steveyegge/beads) CLI for task management instead of a linear checklist.
- Updated `speckit.analyze` prompt to reflect new task generation approach.
- Updated `speckit.implement` prompt to reflect new task generation approach.

## Task Tracking Updates

- Updated `speckit.specify` prompt to review previous work using beads queries.
- Updated `speckit.plan` prompt to research previous work first using beads queries.

## Scripts updates

### Enhanced create-new-feature.sh
- Added `--short-name` argument to allow custom branch short names (2-4 words)
- Added `--number` argument to manually specify branch number
- Added `--json` flag for JSON output mode
- Added intelligent stop word filtering for branch name generation
- Added GitHub branch name length validation (244-byte limit)
- Added support for both specs and git branches when determining next feature number

### New adopt-feature-branch.sh
- Created new script to adopt existing feature branches (e.g., from claude.ai/code)
- Supports mapping arbitrary branch names to specledger feature naming
- Creates `specledger/branch-map.json` for branch-to-feature mapping
- Same argument handling as create-new-feature.sh (--short-name, --number, --json)

### Updated common.sh
- Added `check_mapped_branch()` to check branch mappings
- Added `get_mapped_branch()` to retrieve mapped branch names
- Added branch mapping support in `get_feature_paths()`
- Updated `check_feature_branch()` to handle mapped branches
- Updated `get_current_branch()` to support SPECIFY_FEATURE environment variable
- Added `find_feature_dir_by_prefix()` for prefix-based feature directory lookup

### Updated update-agent-context.sh
- Merged upstream fixes and improvements
- Added specledger-specific project type detection
- Enhanced technology parsing from plan.md

## Template Changes

### spec-template.md
- Added CLAUDE.md reference section for project-specific instructions
- Updated to include specledger-specific guidance
- Enhanced user scenario structure with priority levels (P1, P2, P3)
- Added independent testability requirements for user stories

### plan-template.md
- Added Phase 0 (Research) and Phase 1 (Design) structure
- Enhanced Technical Context section with more fields
- Added Constitution Check gate
- Added Complexity Tracking table
- Added Previous Work section

### tasks-template.md
- Adapted for beads-based task management
- Added dependency tracking (blocks/blockedBy)
- Changed from linear checklist to dependency graph

## Branding Updates

- Changed all references from "spec-kit" to "specledger"
- Changed all references from "speckit" to "specledger"
- Updated URLs from github.com/github/spec-kit to specledger.io
- Updated repository references to github.com/specledger/specledger

## Project-Specific Configuration

- **CLAUDE.md**: Auto-generated from feature plans with active technologies and commands
- **specledger.yaml**: Project configuration file
- **branch-map.json**: Maps arbitrary branch names to specledger feature branches
- **spec-kit-version**: Tracks upstream spec-kit version for potential merges

## Current Upstream Version

Based on the templates and scripts, this fork tracks a version of spec-kit from approximately 2024-2025, with significant local modifications for:
1. Beads integration for task tracking
2. Branch adoption workflow
3. Enhanced branch naming with stop word filtering
4. Specledger branding and project structure
