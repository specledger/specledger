# Specification Analysis Report

| ID | Category | Severity | Location(s) | Summary | Recommendation |
|----|----------|----------|-------------|---------|----------------|
| C1 | Constitution Alignment | CRITICAL | `.specledger/memory/constitution.md:64`, `specledger/610-skills-registry/tasks.md:132` | Constitution requires a task-list verification task for command pattern classification; tasks index/issue set does not define an explicit standalone verification task. | Add a dedicated task (or explicit DoD item on an existing task) for Data CRUD pattern compliance verification. |
| C2 | Constitution Alignment | CRITICAL | `.specledger/memory/constitution.md:65`, `specledger/610-skills-registry/plan.md:145`, `specledger/610-skills-registry/tasks.md:130` | Constitution requires local Supabase stack validation before pushing; no explicit validation task/gate is present in plan/tasks for this feature branch. | Add an explicit validation task/gate (even if "no schema impact") to satisfy constitution workflow requirements. |
| I1 | Inconsistency | HIGH | `specledger/610-skills-registry/spec.md:12`, `specledger/610-skills-registry/spec.md:170`, `specledger/610-skills-registry/plan.md:179`, `specledger/610-skills-registry/research.md:29` | Source of truth for install target paths conflicts: spec says `specledger.yaml`; plan/tasks/research implement constitution+agent-registry resolution. | Pick one canonical source and align spec FR/clarifications + plan/tasks wording. |
| I2 | Inconsistency | HIGH | `specledger/610-skills-registry/spec.md:159`, `specledger/610-skills-registry/spec.md:50`, `specledger/610-skills-registry/plan.md:175`, `specledger/610-skills-registry/research/2026-04-05-skills-cli-source-analysis.md:70` | FR-004 states raw GitHub content path only, but user story/plan/tasks include non-GitHub and `git clone` fallback. | Broaden FR-004 language to cover both GitHub fast-path and git fallback. |
| G1 | Coverage Gap | MEDIUM | `specledger/610-skills-registry/spec.md:166`, `sl issue list --spec 610-skills-registry --type task --json` | FR-011 ("all subcommands actionable network errors") has weak explicit task traceability (mostly inferred via client + polish tests). | Add requirement label(s) or explicit DoD bullets tying FR-011 to each command/polish task. |
| U1 | Underspecification | MEDIUM | `specledger/610-skills-registry/spec.md:152`, `specledger/610-skills-registry/spec.md:182` | Spec has no dedicated Non-Functional Requirements section; NFR-like constraints are scattered across FRs/success criteria. | Add a compact NFR section (performance, reliability, security, observability) with measurable criteria. |
| A1 | Ambiguity | MEDIUM | `specledger/610-skills-registry/spec.md:186` | SC-001 ("discover in under 10 seconds") lacks test conditions/boundary assumptions (network, cache, percentile). | Define measurable test condition (e.g., p95 under stated network profile). |
| D1 | Duplication/Drift | LOW | `specledger/610-skills-registry/plan.md:72`, `specledger/610-skills-registry/plan.md:88` | Command file naming drifts between `skill.go` and `skills.go`. | Normalize to one filename reference to reduce execution ambiguity. |
| I3 | Terminology Drift | LOW | `specledger/610-skills-registry/spec.md:156`, `specledger/610-skills-registry/research.md:126`, `specledger/610-skills-registry/research/2026-04-05-skills-cli-source-analysis.md:47` | Research docs still use `sl skills` in places while spec/plan standardize on `sl skill`. | Normalize research wording to avoid confusion during implementation. |

## Coverage Summary Table

| Requirement Key | Has Task? | Task IDs | Notes |
|-----------------|-----------|----------|-------|
| `sl-skill-search-query-display-results` (FR-001) | Yes | SL-2ce847, SL-9f28f2 | Covered by client + search command |
| `sl-skill-search-limit-flag` (FR-002) | Yes | SL-9f28f2 | Explicit |
| `sl-skill-info-show-3-partner-audit` (FR-003) | Yes | SL-2ce847, SL-279648 | Explicit |
| `sl-skill-add-download-and-save-skill` (FR-004) | Yes | SL-664ae4, SL-cd4b97, SL-04d7eb, SL-51a866 | Coverage exists but FR wording conflicts with plan/tasks |
| `sl-skill-add-update-vercel-lock-schema` (FR-005) | Yes | SL-3a6930, SL-72c546, SL-04d7eb | Explicit |
| `sl-skill-add-telemetry-with-opt-out` (FR-006) | Yes | SL-9c864c, SL-51a866 | Explicit |
| `sl-skill-remove-delete-and-update-lock` (FR-007) | Yes | SL-04d7eb, SL-c06508 | Explicit |
| `sl-skill-list-read-lock-and-display` (FR-008) | Yes | SL-745d54 | Explicit |
| `sl-skill-audit-query-and-display` (FR-009) | Yes | SL-4d17f2 | Explicit |
| `all-subcommands-support-json` (FR-010) | Yes | SL-9f28f2, SL-51a866, SL-279648, SL-745d54, SL-c06508, SL-4d17f2, SL-ae3b59 | Good coverage |
| `all-subcommands-network-errors-actionable` (FR-011) | Yes (inferred) | SL-2ce847, SL-9f28f2, SL-ae3b59 | Label traceability weak |
| `add-overwrite-confirmation` (FR-012) | Yes | SL-51a866 | Explicit |
| `telemetry-version-specledger` (FR-013) | Yes | SL-9c864c | Explicit |
| `add-show-audit-before-confirmation` (FR-014) | Yes | SL-51a866 | Explicit |
| `install-paths-from-configured-agents` (FR-015) | Yes | SL-04d7eb | Conflicts on config source (`specledger.yaml` vs constitution) |
| `ship-embedded-sl-skill-template` (FR-016) | Yes | SL-50e427 | Explicit |

## Constitution Alignment Issues

- C1: Missing explicit task-level command pattern verification gate (required by constitution workflow).
- C2: Missing explicit local Supabase validation gate in plan/tasks for branch workflow compliance.

## Unmapped Tasks

- SL-aafde3 (setup infrastructure)
- SL-0c84c4 (VCR cassette recording)
- SL-ae3b59 (integration/polish; indirectly covers multiple FRs but unlabeled)

## Metrics

- Total Requirements: 16
- Total Tasks: 17
- Coverage % (requirements with >=1 mapped task): 100%
- Ambiguity Count: 1
- Duplication Count: 1
- Critical Issues Count: 2

## Next Actions

- Resolve CRITICAL items before `/specledger.implement`.
- Align requirement text + implementation plan for install path source and source-fetch strategy.
- Add explicit constitution compliance tasks/gates, then re-run verification.

Suggested commands:

1. `sl issue create --title "Add Data CRUD pattern compliance verification task" --type task --priority 1`
2. `sl issue create --title "Add Supabase local stack validation gate for feature branch" --type task --priority 1`
3. Re-run `/specledger.tasks` (or update issue DoD/labels) to restore explicit traceability for FR-011 and constitution gates.

---

## Appendix: Resolutions Applied (2026-04-05)

| ID | Severity | Fix Applied |
|----|----------|-------------|
| C1 | CRITICAL | Added "Data CRUD pattern compliance verified" DoD to SL-ae3b59 (E2E tests task) |
| C2 | CRITICAL | Added "no Supabase schema changes required" DoD to SL-430f49 (Polish feature) |
| I1 | HIGH | Fixed plan.md + research.md to use `specledger.yaml` (not constitution) per #147. Updated SL-04d7eb design. |
| I2 | HIGH | Broadened FR-004 in spec.md to cover both GitHub API and `git clone` fetch methods |
| G1 | MEDIUM | Added FR-011 network error DoD to all 6 command tasks (SL-9f28f2, SL-51a866, SL-279648, SL-745d54, SL-c06508, SL-4d17f2) |
| U1 | MEDIUM | Kept as-is — NFRs implied by CLI design principles doc and constitution |
| A1 | MEDIUM | Reworded SC-001 in spec.md to remove timing claim, replaced with "single command invocation" |
| D1 | LOW | Fixed `skills.go` → `skill.go` in plan.md (2 occurrences) |
| I3 | LOW | Normalized `sl skills` → `sl skill` across all 3 research docs (18 occurrences) |
