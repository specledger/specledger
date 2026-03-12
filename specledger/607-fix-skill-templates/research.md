# Research: Fix Embedded Skill Templates

**Feature**: 607-fix-skill-templates
**Date**: 2026-03-12

## Prior Work

### Related Features

| Feature | Description | Relevance |
|---------|-------------|-----------|
| 008-fix-sl-deps | Previous attempt to fix sl-deps skill | May have been incomplete or overwritten |
| 598-sdd-workflow-streamline | D21 token efficiency principle | Skills should focus on decision patterns, not command reference |
| 599-alignment | Converted audit to skill format | Created sl-audit skill structure |
| 601-cli-skills | CLI skills architecture | Progressive disclosure design patterns |

### Related Issues (from tracker)

- **SL-3137ed**: Delete specledger.audit.md (closed)
- **SL-2e8168**: Create sl-audit skill (closed)
- **SL-91c1c7**: Delete specledger.add-deps.md (closed)
- **SL-32d9fb**: US3: Convert Audit to Skill (closed)

## Current State Analysis

### sl-deps/skill.md (CRITICAL)

**Finding**: Contains exact duplicate of `sl-issue-tracking/skill.md` content

| Metric | Value |
|--------|-------|
| Lines | 227 |
| Content | 100% wrong (issue tracking instead of deps) |
| Token waste | ~2,300 tokens per load |

**Root Cause**: Likely copy-paste error during skill creation. The manifest.yaml correctly describes it as "Manage specification dependencies with sl deps commands" but the content doesn't match.

**Required Content** (from `sl deps --help`):
- `sl deps add` - Add a specification dependency
- `sl deps remove` - Remove a specification dependency
- `sl deps list` - List all dependencies
- `sl deps resolve` - Resolve dependency status
- `sl deps link` - Link local specs
- `sl deps unlink` - Unlink local specs

### sl-audit/skill.md (HIGH)

**Finding**: Lines 239-271 are exact duplicates of lines 206-238

| Metric | Value |
|--------|-------|
| Total lines | 271 |
| Duplicate lines | 33 (lines 239-271) |
| Token waste | ~700 tokens per load |

**Root Cause**: Likely merge conflict or copy-paste error.

**Fix**: Delete lines 239-271

### manifest.yaml Skill Descriptions (MEDIUM)

**Finding**: Descriptions lack trigger keywords for reliable Claude Code activation

| Skill | Current Description | Issue |
|-------|---------------------|-------|
| sl-audit | "Codebase reconnaissance for understanding project structure" | Missing keywords: "audit", "recon", "explore" |
| sl-comment | "Issue and comment management with sl comment commands" | Good - has command name |
| sl-deps | "Manage specification dependencies with sl deps commands" | Good - has command name, but needs "cross-repo", "multi-repo" |
| sl-issue-tracking | "Issue tracking patterns and best practices" | Missing: "multi-session", "task tracking", "inter-issue" |

### sl-audit Aspirational Content (MEDIUM)

**Finding**: References non-existent functionality

| Content | Issue |
|---------|-------|
| `--force` flag | Command doesn't exist |
| `scripts/audit-cache.json` | Automated cache not implemented |
| Cache strategy section | Implies automation that doesn't exist |

**Fix**: Keep manual reconnaissance patterns, remove automated cache references

## Decisions

### D1: sl-deps Skill Content Structure

**Decision**: Follow `sl-comment` skill structure as reference model

**Rationale**:
- `sl-comment/skill.md` is well-structured (138 lines, ~1,400 tokens)
- Uses progressive disclosure: decision criteria first, command reference second
- Includes workflow patterns and error handling

**Structure to adopt**:
1. Overview (when to use)
2. Subcommands table
3. Decision criteria
4. Workflow patterns
5. Error handling

### D2: Distinguish sl deps vs sl issue link

**Decision**: Add explicit comparison section to sl-deps skill

**Rationale**: GitHub issue #82 notes conflation between:
- `sl deps` - Cross-repo specification dependencies (multi-repo)
- `sl issue link` - Inter-issue dependencies (within single repo)

**Content to add**:
```markdown
## When to Use sl deps vs sl issue link

### Use sl deps when:
- Cross-repo specification dependencies
- Multi-repo dependency resolution
- Artifact caching between projects
- Spec imports from external sources

### Use sl issue link when:
- Inter-issue dependencies within same spec
- Blocking relationships between tasks
- Related issues without blocking

**Key distinction**: sl deps is for REPO-level dependencies, sl issue link is for TASK-level dependencies.
```

### D3: Skill Description Format

**Decision**: Add 3+ trigger keywords to each skill description in manifest.yaml

**Format**: "[Primary purpose] with [command names]. Use for [use cases]."

**Examples**:
- sl-deps: "Manage cross-repo specification dependencies with sl deps add/list/resolve/link. Use for multi-repo dependency resolution, artifact caching, and spec imports."
- sl-issue-tracking: "Track multi-session work with sl issue create/list/show/update/close/ready. Use for task tracking, inter-issue dependency management, and progress checkpointing across sessions."

## Alternatives Considered

| Alternative | Rejected Because |
|-------------|------------------|
| Rewrite all skills from scratch | Existing content mostly correct; only sl-deps needs full rewrite |
| Add CLI subcommand for skill discovery | Out of scope (P1 recommendation from issue) |
| Implement automated cache for sl-audit | Out of scope (skill provides manual patterns) |

## References

- GitHub Issue #82: https://github.com/specledger/specledger/issues/82
- sl-comment/skill.md: Reference model for skill structure
- D21 principle from 598-sdd-workflow-streamline: Token efficiency through progressive disclosure
