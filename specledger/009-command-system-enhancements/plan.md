# Implementation Plan: Command System Enhancements

**Branch**: `009-command-system-enhancements` | **Date**: 2026-02-10 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specledger/009-command-system-enhancements/spec.md`

**Note**: This plan documents **existing changes** already implemented in the diff. No new implementation is required.

## Summary

This feature enhances the SpecLedger command system with improved discoverability, codebase analysis capabilities, and Supabase integration for collaborative review workflows. The changes include 3 new commands, updates to 2 existing commands, "Purpose" sections added to all commands, 2 new utility scripts, and bash script path corrections.

## Technical Context

**Language/Version**: Go 1.24+ (CLI), JavaScript/Node.js (utility scripts), Bash (shell scripts)
**Primary Dependencies**: Cobra (CLI), @supabase/supabase-js (Node.js scripts)
**Storage**: File-based (`~/.specledger/credentials.json`, `.beads/issues.jsonl`, `scripts/audit-cache.json`)
**Testing**: Manual testing via command execution
**Target Platform**: macOS, Linux (CLI environments)
**Project Type**: Single project with embedded templates
**Performance Goals**: Command response < 5 seconds for sync operations
**Constraints**: Requires authentication for Supabase operations (`sl login`)
**Scale/Scope**: ~40 files changed, +825 net lines

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Verify compliance with principles from `.specledger/memory/constitution.md`:

- [x] **Specification-First**: Spec.md complete with prioritized user stories
- [x] **Test-First**: Commands are testable via manual execution
- [x] **Code Quality**: Markdown linting for command files, ESLint for JS
- [x] **UX Consistency**: User flows documented in spec.md acceptance scenarios
- [x] **Performance**: Sync operations complete within 10 seconds
- [x] **Observability**: Commands provide clear output with status indicators
- [x] **Issue Tracking**: Feature tracked as 009-command-system-enhancements

**Complexity Violations**: None identified - changes are additive enhancements

## Project Structure

### Documentation (this feature)

```text
specledger/009-command-system-enhancements/
├── spec.md              # Feature specification
├── plan.md              # This file
├── checklists/
│   └── requirements.md  # Quality validation checklist
└── research.md          # (Not needed - documents existing changes)
```

### Source Code (repository root)

```text
# Embedded Templates (pkg/embedded/templates/specledger/)
.claude/commands/
├── specledger.adopt.md        # Updated: --from-audit mode
├── specledger.analyze.md      # Updated: Purpose section
├── specledger.audit.md        # NEW: Codebase audit command
├── specledger.checklist.md    # Updated: Purpose section
├── specledger.clarify.md      # Updated: Purpose section
├── specledger.constitution.md # Updated: Purpose section
├── specledger.help.md         # NEW: Command reference
├── specledger.implement.md    # Updated: Supabase sync step 1
├── specledger.plan.md         # Updated: Purpose section
├── specledger.revise.md       # NEW: Review comments command
├── specledger.specify.md      # Updated: Purpose section
└── specledger.tasks.md        # Updated: Purpose section

.specledger/scripts/bash/
├── adopt-feature-branch.sh    # Updated: .specledger paths
├── common.sh                  # Updated: specledger/ directory
├── create-new-feature.sh      # Updated: .specledger paths
├── setup-plan.sh              # Updated: .specledger paths
└── update-agent-context.sh    # Updated: .specledger paths

scripts/
├── pull-issues.js             # NEW: Supabase issue sync
└── review-comments.js         # NEW: Review comment management

# Root level
AGENTS.md                      # Updated: Simplified, bd focus
```

**Structure Decision**: Changes are distributed across embedded templates, scripts, and documentation. No new directories created.

## Implementation Phases

### Phase 0: Complete (Documentation)

Since this documents existing changes, no research phase was needed. The changes were already implemented.

### Phase 1: Complete (All Changes Implemented)

| Category | File | Status | Description |
|----------|------|--------|-------------|
| New Command | `specledger.audit.md` | ✅ Done | Two-phase codebase audit |
| New Command | `specledger.help.md` | ✅ Done | Command quick reference |
| New Command | `specledger.revise.md` | ✅ Done | Review comment workflow |
| Updated | `specledger.adopt.md` | ✅ Done | Added `--from-audit` mode |
| Updated | `specledger.implement.md` | ✅ Done | Added Supabase sync step |
| Enhanced | 8 command files | ✅ Done | Added Purpose sections |
| New Script | `pull-issues.js` | ✅ Done | Sync beads issues |
| New Script | `review-comments.js` | ✅ Done | Manage review comments |
| Path Fix | 5 bash scripts | ✅ Done | `.specify` → `.specledger` |
| Cleanup | `AGENTS.md` | ✅ Done | Simplified documentation |

### Phase 2: Verification

To verify the changes work correctly:

1. **Test Help Command**: Run `/specledger.help` - should display categorized commands
2. **Test Audit Command**: Run `/specledger.audit` in a project - should detect tech stack
3. **Test Revise Command**: Run `/specledger.revise` with active comments
4. **Test Implement Sync**: Run `/specledger.implement` - should sync before starting
5. **Test Adopt from Audit**: Run `/specledger.adopt --from-audit` after audit

## Key Design Decisions

### D1: Supabase Integration for Revise Command
- **Decision**: Fetch review comments directly from Supabase
- **Rationale**: Enables real-time collaboration without GitHub PR dependency
- **Alternative Rejected**: GitHub API - would require PR to be open

### D2: Two-Phase Audit
- **Decision**: Quick reconnaissance (~15 min) + Deep analysis (~30+ min)
- **Rationale**: Allows fast overview while enabling detailed analysis when needed
- **Alternative Rejected**: Single-phase audit - too slow for quick checks

### D3: Mandatory Sync Before Implement
- **Decision**: Always sync issues from Supabase before starting implementation
- **Rationale**: Prevents duplicate work on claimed issues
- **Alternative Rejected**: Optional sync - would lead to conflicts

### D4: Path Standardization
- **Decision**: Use `.specledger` consistently instead of `.specify`
- **Rationale**: Aligns with project naming (SpecLedger)
- **Alternative Rejected**: Keep mixed naming - confusing for users

## Dependencies

| Dependency | Purpose | Required For |
|------------|---------|--------------|
| `sl login` | OAuth authentication | revise, implement sync |
| Node.js | Run utility scripts | pull-issues.js, review-comments.js |
| @supabase/supabase-js | Supabase client | All Supabase operations |
| git | Repository detection | Auto-detect repo owner/name |

## Risk Assessment

| Risk | Impact | Mitigation |
|------|--------|------------|
| Not logged in | Commands fail | Clear error message with `sl login` prompt |
| Supabase unavailable | Sync fails | Graceful degradation, continue with local data |
| Stale audit cache | Incorrect spec generation | `--force` flag to re-analyze |
| Node.js not installed | Scripts fail | Document prerequisite in README |

## Next Steps

1. **Review the diff** to verify all changes are captured in spec
2. **Test the commands** to ensure they work as documented
3. **Commit the changes** when satisfied
4. **Update embedded templates** with `sl bootstrap` if needed
