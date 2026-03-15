# Revision Log: 603-fix-init-windows

## Revision 1 — 2026-03-05

### Comments Addressed
| Comment | File | Target | Status |
|---------|------|--------|--------|
| 1 | spec.md | User Story 1 (fix scope specificity) | Resolved |
| 2 | spec.md | User Story 2 (shell detection algorithm) | Resolved |
| 3 | spec.md | Edge Cases (acceptance criteria) | Resolved |

### Clusters & Decisions

**Cluster A: Fix specificity + Edge case acceptance (Comments 1 & 3)**
- **Options presented:**
  1. Before/after code snippet + convert edge cases to acceptance scenarios *(chosen)*
  2. Checklist only + mark all edge cases out of scope
  3. Before/after only, leave edge cases as-is
- **Decision:** Option 1 — Added before/after code pattern and exhaustive affected locations list to Dependencies section. Converted edge cases to acceptance scenarios: spaces in temp path (MUST work), terminal independence (same behavior), `--force` (out of scope).

**Cluster B: Shell detection algorithm (Comment 2)**
- **Options presented:**
  1. Git Bash → WSL → skip *(chosen)*
  2. bash on PATH → skip (simple)
  3. Always skip on Windows
- **Decision:** Option 1 — Added FR-006 Decision Algorithm specifying precedence: `bash.exe` on PATH first, then `wsl.exe`, then skip silently with debug-level logging.

### Changes Made
- `spec.md` — Dependencies & Assumptions: Added fix pattern (before/after code snippet) and exhaustive affected locations list with guidance on what NOT to change
- `spec.md` — Requirements FR-006: Added decision algorithm for Windows shell detection with rationale
- `spec.md` — Edge Cases: Converted from open questions to numbered acceptance scenarios with defined expected behavior

---

## Revision 2 — 2026-03-05

### Comments Addressed (re-review)
| Comment | File | Status | Action |
|---------|------|--------|--------|
| 1 — Explicit fix locations & before/after | spec.md | Already resolved in Rev 1 | No change needed |
| 2 — Decision algorithm for post-init script | spec.md | Already resolved in Rev 1 | No change needed |
| 3 — Formalize Edge Cases to Given/When/Then | spec.md | Revised | Converted edge cases 1 & 2 to Given/When/Then format; edge case 3 kept as "Out of scope" |

### Clusters & Decisions

**Single cluster: Spec precision (all 3 comments)**
- **Options presented:**
  1. Only formalize Edge Cases *(chosen)* — Comments 1 & 2 already resolved in current spec
  2. Formalize Edge Cases + add cross-references between User Stories and Dependencies
  3. Full rework of all three with inlined details per section
- **Decision:** Option 1 — Edge cases 1 and 2 rewritten as Given/When/Then acceptance scenarios with parenthetical rationale. Edge case 3 unchanged (already marked out of scope).

### Changes Made
- `spec.md` — Edge Cases: Rewrote items 1 (spaces in temp path) and 2 (terminal independence) from descriptive bullets to Given/When/Then acceptance scenarios
