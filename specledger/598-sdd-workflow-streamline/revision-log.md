# Revision Log: 598-sdd-workflow-streamline

## Cluster A - Dependency Management (Comments 1, 7)

**Question**: How should dependency management commands be consolidated?

**Options Presented**:
1. Merge all into sl deps (Recommended) - Remove add-deps/remove-deps AI commands, add sl deps graph
2. Keep add-deps/remove-deps as single AI command - Merge into one deps AI command, add sl deps graph
3. Keep current structure, only merge sl graph - Minimal change

**Choice Made**: Option 2 - Keep add-deps/remove-deps as single AI command

**Changes Applied**:
- Merged `add-deps` and `remove-deps` AI commands into single `deps` command
- Merged `sl graph` CLI into `sl deps graph` subcommand
- Updated Commands table (15 → 14 commands)
- Updated CLI Commands table (11 → 10 commands)
- Updated Overlaps section
- Updated User Story 2 to reflect single deps AI command

---

## Cluster B - Analysis/Audit Commands (Comment 2)

**Question**: How should the analyze and audit commands be consolidated?

**Options Presented**:
1. Merge into single verify command (Recommended) - New verify command combining both
2. Keep separate with clearer scope - analyze = quick, audit = deep
3. Remove analyze, keep audit only

**Choice Made**: Option 1 - Merge into single verify command

**Changes Applied**:
- Merged `analyze` and `audit` AI commands into single `verify` command
- Updated Commands table (14 → 13 commands)
- Updated Overlaps section
- Updated User Story 3 to reflect verify command

---

## Cluster C - Commands to Remove (Comments 3, 4)

**Question**: Should checklist and resume commands be removed?

**Options Presented**:
1. Remove both (Recommended) - Checklist via manual/WebUI, resume rarely used
2. Remove resume only - Keep checklist for custom checklists
3. Keep both - Niche use cases may be valuable

**Choice Made**: Option 1 - Remove both

**Changes Applied**:
- Removed `checklist` AI command
- Removed `resume` AI command
- Updated Commands table (13 → 11 commands)
- Updated Overview counts

---

## Cluster D - Project Initialization Commands (Comments 5, 6)

**Question**: How should sl bootstrap and sl init be consolidated?

**Options Presented**:
1. Merge into single sl init (Recommended) - Detects existing repo vs new project
2. Rename bootstrap to sl new, keep separate - Two commands for two use cases
3. Keep current structure - No changes

**Choice Made**: Option 1 - Merge into single sl init

**Changes Applied**:
- Merged `sl bootstrap` and `sl init` into single `sl init` command
- Updated CLI Commands table (10 → 9 commands)
- Updated Future Work note to reflect single init command

---

## Cluster E - Playbook Command (Comment 8)

**Question**: Should sl playbook be kept or removed?

**Options Presented**:
1. Keep sl playbook (Recommended) - For fetching and modifying playbook
2. Remove sl playbook - Handled via WebUI and sl init
3. Keep for fetching only - Modification only via WebUI

**Choice Made**: Option 1 - Keep sl playbook

**Changes Applied**:
- No changes - `sl playbook` retained for fetching playbook data and modifying repository playbook

---

---

## Cluster F - Drop Audit, Rename Analyze → Verify (Comments 4, 10, 11)

**Question**: Comments 4, 10, 11 converge on audit/analyze commands. Author guidance: drop `audit` (ship as skill), rename `analyze` to `verify` (aligned with OpenSpec).

**Options Presented**:
1. Drop audit + rename analyze→verify (Recommended) - Remove audit as AI command, ship as `sl-audit` skill. Rename `analyze` to `verify`. Update all references. Add SpecKit migration note.
2. Drop audit only, keep analyze name - Remove audit, but keep SpecKit `analyze` branding.
3. Merge audit into verify (single command) - Combine both into one `verify` command.

**Choice Made**: Option 1 - Drop audit + rename analyze→verify

**Changes Applied (spec.md)**:
- Future Work: updated `/specledger.audit` reference to `sl-audit` skill
- Clarifications: updated Q&A to reflect new decision
- Skills table: 2→4, added `sl-audit` skill
- AI Commands table: 11→11 (removed `audit` row, renamed `analyze` to `verify`)
- Removed AI commands list: added `audit` (7 total removals)
- Overlaps section: updated Analysis/Audit entry
- US1 note: updated `/specledger.audit` → `sl-audit` skill reference
- US3: rewritten for `verify` + `sl-audit` skill
- FR-006/FR-007/FR-008: rewritten for verify + audit skill
- SC-001: updated count from 12 to 11

**Changes Applied (research doc)**:
- D4 command table: renamed analyze→verify, struck through audit
- D8 mapping: renamed analyze→verify, struck through audit
- D12: updated analyze→verify reference
- D12 decision log: updated
- D18 decision log: updated to reflect audit→skill
- D20 decision log: updated reference

---

## Cluster G - Rename specledger-deps → sl-deps (Comment 3)

**Question**: Rename `specledger-deps` skill to `sl-deps` for consistency with `sl-issue-tracking`, `sl-comment` naming convention.

**Options Presented**:
1. Rename specledger-deps → sl-deps (Recommended) - Aligns all skill names to `sl-*` prefix convention.
2. Keep specledger-deps as-is - Defer rename to implementation.

**Choice Made**: Option 1 - Rename specledger-deps → sl-deps

**Changes Applied**:
- spec.md: All 3 occurrences of `specledger-deps` → `sl-deps` (skills table, US4 scenarios, key entities)
- research doc: All 3 occurrences in D5, D19

---

## Cluster H - Spike Template (Comment 1)

**Question**: Add structured spike template to the spec definition.

**Options Presented**:
1. Add template in US6 acceptance scenarios (Recommended)
2. Add template under FR-012/FR-013
3. Add as a new standalone section

**Choice Made**: User override — create separate template file, reference from spec

**Changes Applied**:
- Created `research/spike-template.md` with full template structure (objective, investigation plan, research & findings, prototype results, decision/recommendation) and quick workflow
- spec.md: Updated New Capability #1 to reference template file
- spec.md: Updated FR-013 to reference template file

---

## Cluster I - Checklist Clarification (Comment 5)

**Question**: Clarify that checklist is "constitution but on spec level" — per-feature quality gates.

**Options Presented**:
1. Add clarification inline in command table + Key Entities (Recommended)
2. Add clarification in D12 research doc only

**Choice Made**: Option 2 - Add clarification in D12 research doc only

**Changes Applied**:
- research doc D12: Added clarification paragraph explaining constitution vs checklist scope (project-wide vs feature-level guardrails)

---

## Cluster J - Session Lifecycle Hooks (Comment 6)

**Question**: Consider additional hooks beyond PostToolUse (e.g., SessionEnd) for checkpoint reliability.

**Options Presented**:
1. Add future hooks note in Hooks section + edge case (Recommended)
2. Add as a new User Story (P3)
3. Add to Dependencies & Assumptions only

**Choice Made**: Option 2 - Add as new User Story (P3)

**Changes Applied**:
- spec.md: Added User Story 15 (P3) — Session Lifecycle Hooks for Checkpoint Reliability
- Covers SessionEnd, PreContextCompaction, PostSessionCapture hooks
- Includes acceptance scenarios for detection, warning, and guidance when sessions end without commits/resolution

---

## Cluster K - Research Doc Corrections (Comments 7, 8, 9)

**Comment 7**: `--comments` flag removed from D2 `/specledger.clarify --comments` reference.
**Comment 8**: Updated D2 post-session CLI behavior with conditional warning logic (no commit AND/OR unresolved comments → warn + print resume command).
**Comment 9**: Open Questions updated per user guidance:
- Q1: Kept deferred (no GH issue)
- Q2: Clarified `specledger.` prefix ownership, experimental commands must use different prefix
- Q3: Noted to file as GH issue for `sl skill` command group
- Q4: Added explanation of what "script audit" means
- Q5/Q6: Already resolved, no change

---

## Cluster L - No Change (Comment 2)

**Comment**: "after updating the template would it reload other generated files in sl onboard?"
**Author Guidance**: "needs no change, just ensure we resolve it when done"
**Thread**: Confirmed `sl doctor` will be the one-stop shop for repo health management.

**No changes applied** — resolved by existing design.

---

## Cluster M - Bash Script Audit & CLI Replacement (User Request)

**Question**: Add a user story for the script audit — enumerate all bash scripts, document what each does, and map to `sl` CLI equivalents.

**Changes Applied**:
- spec.md: Added User Story 16 (P1) — Bash Script Audit and CLI Replacement
  - Full inventory of 6 bash scripts with purpose, callers, and proposed `sl` equivalents
  - 4 new `sl` commands: `sl spec info`, `sl spec create`, `sl spec setup-plan`, `sl context update`
  - 1 script superseded by D9 (`adopt-feature-branch.sh`)
  - 1 script absorbed into internal packages (`common.sh`)
  - 7 acceptance scenarios covering JSON compatibility, cross-platform, branch collision prevention
- spec.md: Added FR-036 through FR-040 under "Bash Script Migration (US13, US16)"
- spec.md: Updated CLI Commands table from 9→10 to 9→14 with the 4 new commands

---

## Summary

### AI Commands: 16 → 11 (5 reduction)
- Merged: `add-deps` + `remove-deps` → `deps`
- Renamed: `analyze` → `verify` (OpenSpec alignment)
- Moved to skill: `audit` → `sl-audit` skill
- Removed: `checklist`, `resume`

### Skills: 2 → 4 (2 additions)
- Added: `sl-comment` (new)
- Added: `sl-audit` (moved from AI command)

### CLI Commands: 9 → 14 (5 additions)
- Merged: `sl bootstrap` + `sl init` → `sl init`
- Merged: `sl graph` → `sl deps graph`
- Retained: `sl playbook`
- Added: `sl spec info` (replaces `check-prerequisites.sh`)
- Added: `sl spec create` (replaces `create-new-feature.sh`)
- Added: `sl spec setup-plan` (replaces `setup-plan.sh`)
- Added: `sl context update` (replaces `update-agent-context.sh`)

### Total Commands: AI 11 + CLI 14 + Skills 4 + Hooks 1 = 30 components
