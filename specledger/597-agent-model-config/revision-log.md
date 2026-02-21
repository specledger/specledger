# Revision Log: 597-agent-model-config spec.md

## Revision 1 — 2026-02-21

### Comments Addressed

| Comment | Cluster | Resolution |
|---------|---------|------------|
| 1 (user/team preference conflicts) | A: Precedence | Added 3-tier local hierarchy (personal-local > team-local > global) with personal overrides gitignored |
| 2 (local always takes precedence?) | A: Precedence | Clarified full precedence model in User Story 2 with 5 acceptance scenarios |
| 8 (profiles scope creep) | A: Precedence | Kept profiles at P2 — they solve quick-switching use case distinct from hierarchy |
| 3 (agent.env for arbitrary env) | B: Env Schema | Added FR-007a for `agent.env` catch-all map for arbitrary env var injection |
| 6 (map/dictionary config) | B: Env Schema | Added string-map type to Configuration Key entity, map rendering to TUI FR-009, map editing to User Story 3 |
| 11 (pass os environ + override) | B: Env Schema | Updated FR-007: subprocess inherits process env, config values override. Added assumption. |
| 9 (editing masked fields) | C: Secrets | Kept minimal: masking + file permissions. TUI reveal/hide deferred to planning phase |
| 10 (sensitive data out of scope?) | C: Secrets | Kept minimal FRs. Added assumption that secrets management integration is out of scope but design should not preclude it |
| 4 (define CLI usage in spec?) | D: Scope | Removed specific CLI syntax from acceptance scenarios. UX details deferred to plan/research phase |
| 5 (how to set profile config?) | D: Scope | Deferred to planning phase — quickstart/UX design |
| 7 (TUI needs spike) | D: Scope | Added assumption that TUI requires feasibility spike during planning |
| 12 (constitution migration concern) | D: Scope | Softened FR-013 from MUST to SHOULD, added clarification note |

### Options Presented & Choices Made

| Cluster | Options Offered | Choice |
|---------|----------------|--------|
| A: Precedence | 1) Clarify existing model with 5-tier, 2) Add personal project layer, 3) Add conflict prompting | **Clarify existing model** (with 5-tier hierarchy including personal-local) |
| Profiles | 1) Keep as P3, 2) Remove entirely, 3) Keep as P2 | **Keep as P2** |
| B: Env config | 1) Add agent.env catch-all, 2) Explicit keys only, 3) Defer to research | **Add agent.env catch-all** |
| C: Secrets | 1) Keep minimal + future path, 2) Remove from spec, 3) Full secrets support | **Keep minimal + future path** |
| D: Scope | 1) Clean up + defer details, 2) Aggressive descoping, 3) Keep as-is + annotations | **Clean up + defer details** |

### User Guidance
- spec.md captures WHAT (user stories), not HOW (UX design, CLI syntax)
- UX design details belong in plan/research phase (quickstart document)
- spec-ux.md was a draft mixing user stories with UX design — not git tracked, will be deleted
