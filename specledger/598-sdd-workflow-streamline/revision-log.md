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

## Summary

### AI Commands: 15 → 11 (4 reduction)
- Merged: `add-deps` + `remove-deps` → `deps`
- Merged: `analyze` + `audit` → `verify`
- Removed: `checklist`, `resume`

### CLI Commands: 11 → 9 (2 reduction)
- Merged: `sl bootstrap` + `sl init` → `sl init`
- Merged: `sl graph` → `sl deps graph`
- Retained: `sl playbook`

### Total Reduction: 6 commands (26% reduction)
