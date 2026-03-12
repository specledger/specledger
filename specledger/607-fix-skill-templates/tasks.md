# Tasks: Fix Embedded Skill Templates

**Epic**: SL-30dfb7 - Fix Embedded Skill Templates
**Branch**: `607-fix-skill-templates`
**Spec**: [spec.md](./spec.md) | **Plan**: [plan.md](./plan.md)

## Issue Hierarchy

```
SL-30dfb7 [Epic] Fix Embedded Skill Templates
├── SL-990e00 [Feature] US1: Fix sl-deps Skill Content (P0)
│   └── SL-ff94b5 [Task] Rewrite sl-deps/skill.md with correct content
├── SL-718a53 [Feature] US2: Fix sl-audit Duplicates (P0)
│   └── SL-7bfc04 [Task] Remove duplicate sections from sl-audit/skill.md
├── SL-37d94a [Feature] US3: Update Manifest Descriptions (P1)
│   └── SL-94ac84 [Task] Update manifest.yaml skill descriptions
├── SL-39b0bf [Feature] US4: Remove Aspirational Content (P1)
│   └── SL-6149e3 [Task] Remove aspirational content from sl-audit/skill.md
└── SL-545bf2 [Feature] Polish: Verify and Test Changes (P2)
    └── SL-145827 [Task] Run tests and verify all skill changes
```

## Dependency Graph

```
                    ┌─────────────────┐
                    │   SL-30dfb7     │
                    │   (Epic)        │
                    └────────┬────────┘
                             │
         ┌───────────────────┼───────────────────┐
         │                   │                   │
         ▼                   ▼                   ▼
┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐
│   SL-990e00     │ │   SL-718a53     │ │   SL-39b0bf     │
│   US1: sl-deps  │ │   US2: sl-audit │ │   US4: aspir.   │
│   (P0)          │ │   (P0)          │ │   (P1)          │
└────────┬────────┘ └────────┬────────┘ └────────┬────────┘
         │                   │                   │
         ▼                   │                   │
┌─────────────────┐          │                   │
│   SL-37d94a     │          │                   │
│   US3: manifest │          │                   │
│   (P1)          │◄─────────┘                   │
└────────┬────────┘                              │
         │                                       │
         └───────────────────┬───────────────────┘
                             │
                             ▼
                    ┌─────────────────┐
                    │   SL-545bf2     │
                    │   Polish        │
                    │   (P2)          │
                    └─────────────────┘
```

## Task List by Phase

### Phase 1: US1 - Fix sl-deps Content (P0) — Parallel with US2

**Feature**: SL-990e00 | **Story**: US1 | **Priority**: P0 (Critical)

| Task ID | Title | Status | File |
|---------|-------|--------|------|
| SL-ff94b5 | Rewrite sl-deps/skill.md with correct content | ○ | `pkg/embedded/templates/specledger/skills/sl-deps/skill.md` |

**Independent Test**: Verify file contains `sl deps add/remove/list/resolve/link/unlink` commands, NOT `sl issue` commands.

---

### Phase 2: US2 - Fix sl-audit Duplicates (P0) — Parallel with US1

**Feature**: SL-718a53 | **Story**: US2 | **Priority**: P0 (Critical)

| Task ID | Title | Status | File |
|---------|-------|--------|------|
| SL-7bfc04 | Remove duplicate sections from sl-audit/skill.md | ○ | `pkg/embedded/templates/specledger/skills/sl-audit/skill.md` |

**Independent Test**: Verify file has no duplicate sections, line count ≤ 240.

---

### Phase 3: US3 - Update Manifest Descriptions (P1) — After US1

**Feature**: SL-37d94a | **Story**: US3 | **Priority**: P1 (High)

| Task ID | Title | Status | Dependencies |
|---------|-------|--------|--------------|
| SL-94ac84 | Update manifest.yaml skill descriptions | ○ | SL-ff94b5 |

**Independent Test**: Verify each skill description has ≥3 trigger keywords.

---

### Phase 4: US4 - Remove Aspirational Content (P1) — Parallel with US2

**Feature**: SL-39b0bf | **Story**: US4 | **Priority**: P1 (High)

| Task ID | Title | Status | File |
|---------|-------|--------|------|
| SL-6149e3 | Remove aspirational content from sl-audit/skill.md | ○ | `pkg/embedded/templates/specledger/skills/sl-audit/skill.md` |

**Independent Test**: Verify no `--force` flag or `scripts/audit-cache.json` references.

---

### Phase 5: Polish - Verify and Test (P2) — After All Implementation

**Feature**: SL-545bf2 | **Priority**: P2 (Normal)

| Task ID | Title | Status | Dependencies |
|---------|-------|--------|--------------|
| SL-145827 | Run tests and verify all skill changes | ○ | SL-ff94b5, SL-7bfc04, SL-94ac84, SL-6149e3 |

---

## Parallel Execution

Tasks that can run simultaneously (different files, no shared dependencies):

| Parallel Group | Tasks | Rationale |
|----------------|-------|-----------|
| Group A | SL-ff94b5, SL-7bfc04, SL-6149e3 | Different files (sl-deps vs sl-audit) |
| Group B | SL-94ac84 | After SL-ff94b5 completes |
| Group C | SL-145827 | After all implementation tasks |

## Commands

```bash
# View all issues for this feature
sl issue list --label "spec:607-fix-skill-templates"

# View issue tree with dependencies
sl issue list --label "spec:607-fix-skill-templates" --tree

# Find ready-to-work issues (unblocked)
sl issue ready --label "spec:607-fix-skill-templates"

# Show specific issue details
sl issue show SL-30dfb7 --tree
```

## Definition of Done Summary

| Issue ID | Title | DoD Items |
|----------|-------|-----------|
| SL-ff94b5 | Rewrite sl-deps/skill.md | sl deps add/remove/list/resolve/link/unlink documented; comparison section distinguishes from sl issue link; no sl issue commands except in comparison |
| SL-7bfc04 | Remove sl-audit duplicates | Duplicate CLI Reference removed; duplicate Troubleshooting removed; file ≤238 lines |
| SL-94ac84 | Update manifest descriptions | sl-audit has keywords; sl-comment has keywords; sl-deps mentions cross-repo; sl-issue-tracking mentions multi-session |
| SL-6149e3 | Remove aspirational content | No --force references; no audit-cache.json references; cache strategy is manual only |
| SL-145827 | Verify and test | go test passes; sl-deps verified; sl-audit verified; manifest verified |

## MVP Scope

**Minimum Viable Product**: US1 + US2 (SL-990e00 + SL-718a53)

These two user stories address the critical P0 issues:
1. Fix broken sl-deps content (wrong content)
2. Remove token-wasting duplicates in sl-audit

US3 and US4 are enhancements that improve triggering accuracy and content clarity but are not blocking issues.

## Success Criteria Mapping

| Success Criteria | Tasks |
|------------------|-------|
| SC-001: sl-deps content matches manifest | SL-ff94b5 |
| SC-002: sl-audit no duplicates | SL-7bfc04 |
| SC-003: Token count reduced 700+ | SL-7bfc04 |
| SC-004: 4 skill descriptions have 3+ keywords | SL-94ac84 |
| SC-005: No non-existent command references | SL-6149e3 |
