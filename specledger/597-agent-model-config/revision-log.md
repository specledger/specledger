# Revision Log: 597-agent-model-config

## Revision 4 — 2026-02-22

### Comments Addressed

| Comment | Cluster | Resolution |
|---------|---------|------------|
| Comment 1 (0600 perms with other values) | Sensitive field permissions | Dropped per-field 0600 file permissions; guardrails reduced to display masking + scope warning |
| Comment 2 (single field controls perms?) | Sensitive field permissions | Same — removed all per-field permission references |
| Comment 4 (weird to define per field) | Sensitive field permissions | Same — file permissions no longer defined per field |
| Comment 3 (mask except last 4 chars?) | Display masking | Updated masking format to `****[last4]` (already in plan.md, aligned in data-model.md) |
| Comment 5 (config path library) | Config path consolidation | Confirmed stdlib `os.UserHomeDir()`. Consolidated global config path from `~/.config/specledger/` to `~/.specledger/` to align with credentials |
| Comment 6 (tui_enabled — low pri) | Low-priority annotations | Added inline note: `tui_enabled` is currently unused placeholder in codebase |
| Comment 7 (specledger.yaml path — low pri) | Low-priority annotations | Added inline note: path stays as-is, moving to `.specledger/` is high-impact (200+ refs), out of scope |
| Comment 8 (launcher consumers) | Launcher call sites | Added call site inventory to plan.md: 3 consumers (sl new, sl init, sl revise), no code duplication, no merge conflict risk |

### Options Presented & Choices Made

| Cluster | Options Offered | Choice |
|---------|----------------|--------|
| Sensitive field permissions (C1,2,4) | 1) Drop file permissions entirely, 2) Keep 0600 at file level only, 3) Auto-route sensitive fields to personal scope | **Drop file permissions entirely** — guardrails are display masking + scope warning only |
| Config path library (C5) | 1) Document stdlib as-is, 2) Recommend consolidating paths, 3) Adopt os.UserConfigDir() | **Recommend consolidating paths** → then user directed: consolidate under `~/.specledger/` now |
| Low-priority items (C6,7) | 1) Add inline annotations only, 2) Remove tui_enabled from example, 3) Flag both as future work | **Add inline annotations only** |
| Launcher call sites (C8) | 1) Add call site inventory to plan.md, 2) Minimal note only | **Add call site inventory to plan.md** |

### Files Modified

| File | Change |
|------|--------|
| `data-model.md` | Removed 0600 from ConfigKeyDef.Sensitive description; removed 0600 from AgentConfig preamble; removed "File permissions" bullet from sensitive tag drives (3→2 behaviors); updated masking format to `****[last4]`; added tui_enabled annotation; added specledger.yaml path annotation; consolidated global config path to `~/.specledger/config.yaml` |
| `plan.md` | Updated sensitive field guardrails from 3 to 2 behaviors; removed 0600 from constraints; consolidated global config path to `~/.specledger/config.yaml`; added launcher consumer call site inventory |
| `spec.md` | Updated global config path to `~/.specledger/config.yaml`; updated assumption about path resolution (stdlib, no third-party lib) |
| `quickstart.md` | Updated global config path to `~/.specledger/config.yaml` |
| `research.md` | Updated global config path to `~/.specledger/config.yaml` in R1 rationale and merge algorithm |
| `research/002-config-precedence-patterns.md` | Updated global config path in precedence table |

### User Guidance
- Per-field 0600 file permissions are fully removed — the gitignored `specledger.local.yaml` is the security mechanism for sensitive values
- Global config path consolidated from `~/.config/specledger/config.yaml` to `~/.specledger/config.yaml` — implementation must update `getConfigPath()` in `pkg/cli/config/config.go`
- No `.config/specledger` references remain in the feature artifacts (verified via full sweep)
- Launcher package has no code duplication — all 3 agent-launching commands use the shared `launcher` package

---

## Revision 3 — 2026-02-22

### Comments Addressed

| Comment | Cluster | Resolution |
|---------|---------|------------|
| quickstart-comment-1 (drop "Migration from Shell Alias") | Migration removal | Full purge of CONSTITUTION.md migration feature across all artifacts |

### Options Presented & Choices Made

| Cluster | Options Offered | Choice |
|---------|----------------|--------|
| Migration removal scope | 1) Full purge across all artifacts (Recommended), 2) Quickstart + spec only, 3) Quickstart only | **Full purge across all artifacts** — removed all migration references from quickstart, spec, plan, data-model, research, and checklist |

### Files Modified

| File | Change |
|------|--------|
| `quickstart.md` | Removed Section 5 "Migration from Shell Alias"; renumbered Section 6 → 5 |
| `spec.md` | Removed FR-013 (CONSTITUTION.md migration), SC-006 (migration success criterion), edge case about CONSTITUTION.md migration, assumption about migration clarification |
| `plan.md` | Removed `migration/` package from project structure; removed migration mention from structure decision |
| `data-model.md` | Removed "Migration Lifecycle" state transition section |
| `research.md` | Removed R8 (CONSTITUTION.md Migration); renumbered R9 → R8 (TUI), R10 → R9 (Sensitive Values) |
| `checklists/requirements.md` | Updated FR count (13 → 12), edge case count (6 → 5), removed migration mention |

### User Guidance
- CONSTITUTION.md agent preference detection and migration is fully descoped from this feature
- The `ReadAgentPreference()` function in `bootstrap_helpers.go` remains untouched — it serves existing bootstrap flows, not this config feature

---

## Revision 2 — 2026-02-22

### Comments Addressed

| Comment | Cluster | Resolution |
|---------|---------|------------|
| quickstart-1 (secrets in config / separate file / interpolation) | Secrets & Scope Flags | Clarified with `--personal` flag for sensitive values, Go struct tags as guardrails, warnings on git-tracked scope |

### Options Presented & Choices Made

| Cluster | Options Offered | Choice |
|---------|----------------|--------|
| Secrets handling | 1) Auto-route sensitive keys to personal-local, 2) Separate secrets file + interpolation, 3) Clarify existing design + guardrails | **Clarify existing design + guardrails** — minimal achievable with struct tags, CLI warning, and pre-commit hook recommendation |
| Flag naming | 1) `--personal` (recommended), 2) `--local`, 3) `--private` | **`--personal`** — already used in quickstart, clear intent |

### Files Modified

| File | Change |
|------|--------|
| `quickstart.md` §1 | Changed `sl config set agent.auth-token` → `sl config set --personal agent.auth-token`; added secrets warning callout |
| `quickstart.md` §2 | Changed profile auth-token example to use `--personal` |
| `data-model.md` AgentConfig | Added `Sensitive` column to field table; added Go struct tag convention section explaining `sensitive:"true"` drives masking, permissions, and scope warnings |
| `spec.md` FR-004 | Expanded to include `--personal` flag alongside `--global`; documented default targets team-local |
| `spec.md` Assumptions | Added sensitive struct tag guardrail assumption; added secret interpolation as future extensibility note |
| `plan.md` | Added "Design Decisions" section with CLI Scope Flags table and Sensitive Field Guardrails subsection |

### User Guidance
- Secrets management integration (interpolation, vault backends) remains out of scope — struct tags + warnings are best-effort guardrails
- Teams should adopt pre-commit secret detection tools as defense-in-depth
- `--personal` flag is the canonical way to store sensitive values in gitignored config

---

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
