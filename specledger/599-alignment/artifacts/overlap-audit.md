# Overlap Audit Report

**Feature**: 599-alignment | **Date**: 2026-03-02
**Scope**: AI Commands and Skills
**Reference**: 598-sdd-workflow-streamline spec for consolidation decisions

## Executive Summary

- **15 AI commands** audited in `.claude/commands/`
- **7 AI commands** to remove/absorb (per 598 spec)
- **~5 skills** audited in `pkg/embedded/`
- **1 deprecated pattern** identified (`sl revise --summary`)
- **2 bash script dependencies** (will be resolved by 598)

---

## AI Commands Consolidation (from 598)

### Commands to Remove (7)

| Command | Fate | Rationale (from 598 spec) |
|---------|------|---------------------------|
| `resume` | **Remove** | Duplicate of `implement` |
| `help` | **Remove** | Absorbed by `onboard` |
| `adopt` | **Remove** | Replaced by context detection fallback chain (D9) |
| `add-deps` | **Remove** | Agent calls `sl deps` CLI directly |
| `remove-deps` | **Remove** | Agent calls `sl deps` CLI directly |
| `revise` | **Remove** | Absorbed by `clarify`; `sl revise` stays as CLI launcher |
| `audit` | **Convert to skill** | Ship as `sl-audit` skill (codebase reconnaissance is passive context) |

### Commands to Rename

| Old Name | New Name | Rationale |
|----------|----------|-----------|
| `analyze` | `verify` | Aligned with OpenSpec terminology |

### Commands to Keep (8)

| Command | Purpose | Notes |
|---------|---------|-------|
| `specify` | Create feature specs | Core workflow |
| `clarify` | Resolve spec ambiguities + comments | Absorbs `revise` functionality |
| `plan` | Create implementation plan | Core workflow |
| `tasks` | Generate task list | Core workflow |
| `implement` | Execute tasks | Core workflow |
| `onboard` | Project onboarding | Absorbs `help` |
| `constitution` | Project constitution | Setup |
| `checklist` | Generate checklists | Utility |

### New Commands to Add (per 598)

| Command | Purpose | Priority |
|---------|---------|----------|
| `spike` | Exploratory research documents | US6 |
| `checkpoint` | Session capture verification | US7 |

---

## Current AI Commands Audit

### Classification Legend

| Classification | Definition |
|---------------|------------|
| pure-orchestration | AI command that orchestrates CLI calls without duplicating logic |
| has-business-logic | Contains logic that might belong in CLI |
| duplicates-CLI | Significant overlap with CLI functionality |
| has-deprecated-pattern | Uses outdated CLI patterns |

### Results with 598 Status

| File | Lines | Audit Classification | 598 Status | Notes |
|------|-------|---------------------|------------|-------|
| specledger.specify.md | 287 | pure-orchestration | **KEEP** | Bash script dep → 598 |
| specledger.tasks.md | 417 | pure-orchestration | **KEEP** | Well-structured |
| specledger.audit.md | 301 | pure-orchestration | **→ SKILL** | Convert to `sl-audit` skill |
| specledger.checklist.md | 293 | pure-orchestration | **KEEP** | Template generation |
| specledger.implement.md | 239 | pure-orchestration | **KEEP** | Core workflow |
| specledger.analyze.md | 195 | pure-orchestration | **RENAME** | → `verify` |
| specledger.clarify.md | 178 | has-deprecated-pattern | **KEEP** | Update `sl revise --summary` |
| specledger.adopt.md | 154 | pure-orchestration | **REMOVE** | Context detection replaces |
| specledger.add-deps.md | 134 | pure-orchestration | **REMOVE** | Agent calls `sl deps` directly |
| specledger.plan.md | 101 | pure-orchestration | **KEEP** | Bash script dep → 598 |
| specledger.onboard.md | 103 | pure-orchestration | **KEEP** | Absorbs `help` |
| specledger.remove-deps.md | 97 | pure-orchestration | **REMOVE** | Agent calls `sl deps` directly |
| specledger.constitution.md | 89 | pure-orchestration | **KEEP** | Setup command |
| specledger.help.md | 68 | pure-orchestration | **REMOVE** | Absorbed by `onboard` |
| specledger.resume.md | 62 | pure-orchestration | **REMOVE** | Duplicate of `implement` |

### Summary

| 598 Action | Count | Files |
|------------|-------|-------|
| KEEP | 8 | specify, tasks, checklist, implement, clarify, plan, onboard, constitution |
| REMOVE | 5 | resume, help, adopt, add-deps, remove-deps |
| RENAME | 1 | analyze → verify |
| → SKILL | 1 | audit → sl-audit skill |
| **Total** | 15 | |

### Action Items (Implementation Order)

| Phase | File | Action | Depends On |
|-------|------|--------|------------|
| 1 | specledger.clarify.md | Replace `sl revise --summary` → `sl comment list` | 598 Stream 1 |
| 1 | specledger.specify.md | Replace `create-new-feature.sh` → `sl spec create` | 598 Stream 2 |
| 1 | specledger.plan.md | Replace `setup-plan.sh` → `sl spec setup-plan` | 598 Stream 2 |
| 2 | specledger.analyze.md | Rename to `verify.md` | None |
| 2 | specledger.audit.md | Convert to `sl-audit` skill | None |
| 2 | specledger.resume.md | Remove (merge into implement) | None |
| 2 | specledger.help.md | Remove (merge into onboard) | None |
| 2 | specledger.adopt.md | Remove | Context detection in 598 |
| 2 | specledger.add-deps.md | Remove | Agent calls `sl deps` |
| 2 | specledger.remove-deps.md | Remove | Agent calls `sl deps` |

---

## Skills Audit

### Classification Legend

| Classification | Definition |
|---------------|------------|
| unique-knowledge | Provides domain knowledge not available elsewhere |
| duplicates-help | Significantly duplicates CLI --help content |
| outdated | References deprecated patterns or commands |

### Results

| File | Location | Classification | Notes |
|------|----------|----------------|-------|
| sl-issue-tracking/skill.md | templates/skills/ | unique-knowledge | Model skill: decision criteria + patterns |
| specledger-deps/SKILL.md | skills/skills/ | unique-knowledge | Dependency management patterns |
| commands/*.md (15 files) | skills/commands/ | **investigate** | Appears to be copies of AI commands |

### Summary

| Classification | Count | Notes |
|---------------|-------|-------|
| unique-knowledge | 2 | Good model skills |
| investigate | 15 | Possible duplication |
| duplicates-help | 0 | None identified |
| outdated | 0 | None identified |

### Action Items

| Priority | Item | Action |
|----------|------|--------|
| Medium | skills/commands/*.md | Investigate if duplication or intentional embedded copy |
| Low | Create sl-comment skill | Document `sl comment` patterns (598 Stream 1) |
| Low | Create sl-spec skill | Document `sl spec` patterns (598 Stream 2) |

---

## Detailed Findings

### 1. Deprecated Pattern in specledger.clarify.md

**Current (line 39)**:
```bash
sl revise --summary
```

**Should be (after 598)**:
```bash
sl comment list --status open --json
```

**Impact**: Medium - Command will still work but uses deprecated pattern

**Recommendation**: Update after 598 Stream 1 (`sl comment` CLI) is implemented

### 2. Bash Script Dependencies

**Files affected**:
- `specledger.specify.md` → `.specledger/scripts/bash/create-new-feature.sh`
- `specledger.plan.md` → `.specledger/scripts/bash/setup-plan.sh`

**Impact**: Low - Scripts work, but not cross-platform

**Recommendation**: Update after 598 Stream 2 (`sl spec` CLI) is implemented

### 3. Skill/Command Duplication Question

**Observation**: `pkg/embedded/skills/commands/` contains 15 files matching AI commands

**Possible explanations**:
1. Intentional: Embedded templates for new project initialization
2. Accidental: Historical duplication not cleaned up
3. Different purpose: Skill versions of commands

**Recommendation**: Investigate and document purpose

---

## Recommendations Summary

### Immediate (This Feature)

1. Document layer responsibilities (artifacts/layer-responsibilities.md) ✓
2. Document overlap audit findings (this file) ✓
3. Create AI command template ✓
4. Create skill template ✓

### After 598 Stream 1 (sl comment)

1. Update `specledger.clarify.md` to use `sl comment list`
2. Create `sl-comment` skill

### After 598 Stream 2 (sl spec)

1. Update `specledger.specify.md` to use `sl spec create`
2. Update `specledger.plan.md` to use `sl spec setup-plan`
3. Create `sl-spec` skill

### Future

1. Investigate `pkg/embedded/skills/commands/` purpose
2. Consider skill loading trigger documentation
3. Add cross-layer interaction examples
