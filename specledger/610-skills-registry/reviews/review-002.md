## Specification Analysis Report

| ID | Category | Severity | Location(s) | Summary | Recommendation |
|----|----------|----------|-------------|---------|----------------|
| C1 | Constitution Alignment | CRITICAL | `.specledger/memory/constitution.md:29`, `.specledger/memory/constitution.md:65`, `specledger/610-skills-registry/plan.md:235`, issue `SL-430f49` | Plan explicitly treats Supabase validation as "not a conflict," but constitution Principle VII and workflow rules require local Supabase validation on feature branches before push. No explicit task enforces this gate. | Add an explicit polish task to run/verify local Supabase stack validation (or explicitly scoped exemption via constitution update process, outside this command). |
| I1 | Inconsistency | HIGH | issue `SL-04d7eb` (DoD + design), `specledger/610-skills-registry/spec.md:170` | Install task DoD says "agent paths resolved from constitution + registry," while spec FR-015 requires resolution from `specledger.yaml` configured agents. | Normalize wording in task DoD/design: source of truth is `specledger.yaml` + agent registry mapping. |
| U1 | Underspecification | HIGH | `specledger/610-skills-registry/spec.md:152`, `specledger/610-skills-registry/plan.md:18-20` | Spec has no explicit Non-Functional Requirements section, while plan introduces performance/operational constraints (`<2s`, no Node, public APIs only). | Add explicit NFRs in `spec.md` and map them to concrete tasks/tests. |
| G1 | Coverage Gap | MEDIUM | `specledger/610-skills-registry/spec.md:165-166`, issues `SL-9f28f2`, `SL-51a866`, `SL-279648`, `SL-745d54`, `SL-c06508`, `SL-4d17f2` | FR-010 (`--json` for all subcommands) and FR-011 (actionable network errors) are implemented by inference in command tasks but not consistently trace-labeled (`requirement:FR-010`, `requirement:FR-011`). | Add explicit requirement labels to all six command tasks for FR-010/FR-011 for deterministic traceability. |
| A1 | Ambiguity | MEDIUM | `specledger/610-skills-registry/spec.md:35`, `specledger/610-skills-registry/spec.md:74`, `specledger/610-skills-registry/spec.md:156` | Terms like "user-friendly error," "prominent warning," and "compact format" are only partially measurable. | Define display/error criteria (e.g., required fields, warning prefix, stderr template, max widths) in spec or contracts. |
| I2 | Inconsistency | MEDIUM | `specledger/610-skills-registry/tasks.md:126` | Tasks index says T010-T012 can run in parallel because they "modify only skill.go with no overlapping functions," but all target same file (`pkg/cli/commands/skill.go`), creating practical merge/conflict risk. | Keep logical parallelism but add serialization guidance or function ownership split to avoid integration collisions. |
| D1 | Duplication | LOW | issue pairs `SL-f97f5c` + `SL-aafde3`, `SL-38ebee` + `SL-50e427` | Parent feature issue and child task have near-identical title/DoD wording, which adds noise in reports. | Keep parent as outcome-level only; move execution details to child task wording. |
| U2 | Underspecification | LOW | `specledger/610-skills-registry/spec.md:182-193` | Success criteria numbering/order is non-sequential (`SC-008` appears before `SC-006`/`SC-007`), which weakens reference clarity in tooling/reviews. | Reorder SC list or renumber for deterministic referencing. |

### Coverage Summary Table

| Requirement Key | Has Task? | Task IDs | Notes |
|-----------------|-----------|----------|-------|
| `search-skills-registry` (FR-001) | Yes | `SL-2ce847`, `SL-9f28f2` | Explicit labels present |
| `search-support-limit` (FR-002) | Yes | `SL-9f28f2` | Explicit label present |
| `info-display-audit-partners` (FR-003) | Yes | `SL-2ce847`, `SL-279648` | Explicit labels present |
| `add-download-install-skill` (FR-004) | Yes | `SL-664ae4`, `SL-cd4b97`, `SL-04d7eb`, `SL-51a866` | Explicit labels present |
| `write-vercel-compatible-lock` (FR-005) | Yes | `SL-3a6930`, `SL-72c546`, `SL-04d7eb` | Explicit labels present |
| `telemetry-opt-out-gated` (FR-006) | Yes | `SL-9c864c` | Explicit label present |
| `remove-skill-cleanly` (FR-007) | Yes | `SL-c06508` | Explicit label present |
| `list-installed-from-lock` (FR-008) | Yes | `SL-745d54` | Explicit label present |
| `audit-installed-skills` (FR-009) | Yes | `SL-4d17f2` | Explicit label present |
| `json-output-all-subcommands` (FR-010) | Yes | `SL-9f28f2`, `SL-51a866`, `SL-279648`, `SL-745d54`, `SL-c06508`, `SL-4d17f2` | Inferred from design/AC; missing explicit requirement labels |
| `actionable-network-errors` (FR-011) | Yes | `SL-9f28f2`, `SL-51a866`, `SL-279648`, `SL-745d54`, `SL-c06508`, `SL-4d17f2` | Inferred from DoD item text; missing explicit requirement labels |
| `confirm-overwrite-on-add` (FR-012) | Yes | `SL-51a866` | Explicit label present |
| `telemetry-client-id-versioned` (FR-013) | Yes | `SL-9c864c` | Explicit label present |
| `audit-before-add-confirm` (FR-014) | Yes | `SL-51a866` | Explicit label present |
| `resolve-agent-targets-from-config` (FR-015) | Yes | `SL-04d7eb` | Explicit label present; wording inconsistency flagged in I1 |
| `ship-embedded-sl-skill-template` (FR-016) | Yes | `SL-50e427` | Explicit label present |

### Constitution Alignment Issues

- `C1` is open and blocking: Supabase validation gate is not explicitly represented as an executable task despite constitutional MUST-level workflow constraints.

### Unmapped Tasks

- No fully unmapped implementation tasks after inference.
- Traceability weak spots (missing explicit FR labels): `SL-aafde3`, `SL-0c84c4`, `SL-ae3b59` (enablement/testing tasks), plus FR-010/FR-011 command tasks relying on inferred mapping.

### Metrics

- Total Requirements: **16** (functional), **0 explicit NFR entries**
- Total Tasks (issue type `task`): **17**
- Coverage % (requirements with >=1 mapped task): **100% inferred** / **87.5% explicit-label traceability**
- Ambiguity Count: **1**
- Duplication Count: **1**
- Critical Issues Count: **1**

## Next Actions

- Since a **CRITICAL** issue exists, resolve before `/specledger.implement`.
- Suggested commands:
  1. `sl issue create --title "Add Supabase validation gate task for 610-skills-registry" --type task --priority 1`
  2. `sl issue create --title "Normalize FR-015 agent path wording in install task" --type task --priority 1`
  3. `sl issue create --title "Add explicit FR-010/FR-011 labels to all skill command tasks" --type task --priority 2`
  4. `sl issue create --title "Add explicit NFR section to spec (performance/ops constraints)" --type task --priority 2`

## Remediation Edits (Top 10, no file changes applied)

1. Add a new polish task: "Validate local Supabase stack before push for this feature branch" with clear DoD command evidence.
2. Update `SL-04d7eb` DoD phrase to "agent paths resolved from `specledger.yaml` configured agents + registry mapping."
3. Add `requirement:FR-010` to all six command tasks (`SL-9f28f2`, `SL-51a866`, `SL-279648`, `SL-745d54`, `SL-c06508`, `SL-4d17f2`).
4. Add `requirement:FR-011` to those same six command tasks.
5. Add explicit NFR section to `spec.md` with measurable performance/error-output criteria from `plan.md`.
6. Add measurable definition for "prominent warning" (e.g., required header + risk source + caution line).
7. Add measurable definition for "user-friendly/actionable errors" (3-part structure: what failed, likely cause, suggested command).
8. Adjust `tasks.md` parallelization note for T010-T012 to acknowledge same-file coordination risk.
9. De-duplicate parent/child issue wording for setup and US7 to reduce review noise.
10. Normalize success criteria order in `spec.md` (SC numbering sequence) to improve deterministic references.

---

## Appendix: Resolutions Applied (2026-04-05)

| ID | Severity | Resolution |
|----|----------|------------|
| C1 | CRITICAL | Ignored — Supabase validation already has DoD gate on SL-430f49 from review-001. sl skill has zero Supabase interaction. |
| I1 | HIGH | **Fixed**: SL-04d7eb DoD updated to "specledger.yaml + agent registry (NOT constitution)". Verified via `sl issue show`. |
| U1 | HIGH | Ignored — NFR section deliberately kept out per user decision (review-001 A1). NFRs implied by CLI design principles. |
| G1 | MEDIUM | **Fixed**: Added `requirement:FR-010` and `requirement:FR-011` labels to all 6 command tasks. |
| A1 | MEDIUM | **Fixed**: Replaced vague terms in spec: "user-friendly error" → 3-part stderr template, "prominent warning" → `⚠ Warning:` format with risk details, "compact format" → `{source}@{name}  {installs}` column layout. Added concrete error examples to quickstart.md. |
| I2 | MEDIUM | **Fixed**: Added function ownership guidance to tasks.md for T010-T012 parallel work. |
| D1 | LOW | Acknowledged — parent/child wording overlap is a creation artifact. Not worth updating. |
| U2 | LOW | Acknowledged — SC numbering is stable and referenced. Renumbering would break existing review references. |
