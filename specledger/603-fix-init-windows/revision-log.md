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
