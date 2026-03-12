# Quickstart: Fix Embedded Skill Templates

**Feature**: 607-fix-skill-templates
**Time Estimate**: 30 minutes

## Prerequisites

- [ ] Go 1.24.2 installed
- [ ] Repository checked out on branch `607-fix-skill-templates`
- [ ] Familiarity with `sl deps` commands (run `sl deps --help`)

## Quick Reference

### Files to Modify

| File | Action | Lines Changed |
|------|--------|---------------|
| `pkg/embedded/templates/specledger/skills/sl-deps/skill.md` | **REPLACE** | All 227 lines |
| `pkg/embedded/templates/specledger/skills/sl-audit/skill.md` | **DELETE** | Lines 239-271 (33 lines) |
| `pkg/embedded/templates/manifest.yaml` | **UPDATE** | 4 skill descriptions |

### Verification Commands

```bash
# Verify no duplicate content in sl-audit
wc -l pkg/embedded/templates/specledger/skills/sl-audit/skill.md
# Expected: ~238 lines (down from 271)

# Verify sl-deps contains correct commands
grep "sl deps" pkg/embedded/templates/specledger/skills/sl-deps/skill.md
# Expected: Multiple matches for add/remove/list/resolve/link/unlink

# Verify sl-deps does NOT contain issue commands
grep "sl issue" pkg/embedded/templates/specledger/skills/sl-deps/skill.md
# Expected: No matches (or only in comparison section)

# Run tests
go test ./pkg/embedded/... -v
```

## Implementation Steps

### Step 1: Fix sl-deps/skill.md

1. Read `pkg/embedded/templates/specledger/skills/sl-comment/skill.md` as reference
2. Rewrite `pkg/embedded/templates/specledger/skills/sl-deps/skill.md` with:
   - Overview of `sl deps` purpose
   - Subcommands table (add/remove/list/resolve/link/unlink)
   - Decision criteria (when to use deps vs issue link)
   - Workflow patterns
   - Error handling

### Step 2: Fix sl-audit/skill.md

1. Open `pkg/embedded/templates/specledger/skills/sl-audit/skill.md`
2. Delete lines 239-271 (duplicate "CLI Reference" and "Troubleshooting" sections)
3. Remove `--force` flag reference (non-existent)
4. Simplify cache strategy section to reference manual paths only

### Step 3: Update manifest.yaml Descriptions

Update skill descriptions in `pkg/embedded/templates/manifest.yaml`:

```yaml
skills:
  - name: sl-audit
    description: "Codebase reconnaissance with tech stack detection and module analysis. Use for understanding unfamiliar codebases, architecture validation, and entry point discovery."
  - name: sl-comment
    description: "Review comment management with sl comment list/show/reply/resolve. Use for addressing review feedback, thread management, and comment resolution."
  - name: sl-deps
    description: "Manage cross-repo specification dependencies with sl deps add/list/resolve/link. Use for multi-repo dependency resolution, artifact caching, and spec imports."
  - name: sl-issue-tracking
    description: "Track multi-session work with sl issue create/list/show/update/close/ready. Use for task tracking, inter-issue dependency management, and progress checkpointing."
```

### Step 4: Verify and Test

```bash
# Run embedded template tests
go test ./pkg/embedded/... -v

# Verify templates embed correctly
go run . doctor --template
```

## Success Criteria Checklist

- [ ] sl-deps/skill.md describes all 6 `sl deps` subcommands
- [ ] sl-deps/skill.md does NOT contain `sl issue` commands (except in comparison)
- [ ] sl-audit/skill.md has no duplicate sections
- [ ] sl-audit/skill.md ≤ 240 lines
- [ ] manifest.yaml has 4 updated skill descriptions with trigger keywords
- [ ] `go test ./pkg/embedded/...` passes
