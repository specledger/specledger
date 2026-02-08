# FORK notes

Notes on changes made to local files compared to upstream [github/spec-kit](https://github.com/github/spec-kit/).

## Task Generation Updates

- Updated `speckit.tasks` prompt template and `tasks.md` template to use [beads](https://github.com/steveyegge/beads) CLI for task management instead of a linear checklist.
- Updated `speckit.analyze` prompt to reflect new task generation approach.
- Updated `speckit.implement` prompt to reflect new task generation approach.

## Task Tracking Updates

- Updated `speckit.specify` prompt to review previous work using beads queries.
- Updated `speckit.plan` prompt to research previous work first using beads queries.

## Scripts updates

- Update to `.specify/scripts/bash/create-new-feature.sh` to accept arguments for branch short name and branch number.
- Updated `speckit.specify` to use new script arguments.
- Updated `update-agent-context.sh` with upstream fixes.
- Added new script `.specify/scripts/bash/adopt-feature-branch.sh` to help adopt existing feature branches (e.g. from https://claude.ia/code)
- Added new prompt `speckit.adopt` to adopt existing feature branches using the new script.
- Updated `.specify/scripts/bash/common.sh` to support feature branch mapping for adopted branches.
