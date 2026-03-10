# Implementation Plan: Auto-Generate Spec Numbers

**Branch**: `604-auto-spec-numbers` | **Date**: 2026-03-10 | **Spec**: [spec.md](./spec.md)

## Summary

Make the `--number` flag optional in `sl spec create` by auto-generating collision-free feature numbers. Adds `GetNextAvailableNum()` to scan local dirs, local branches, and remote branches. Updates AI skill docs to remove manual number lookup.

## Technical Context

**Language/Version**: Go 1.24.2
**Primary Dependencies**: go-git/v5 (git operations), cobra (CLI)
**Storage**: File-based (specledger/ directories)
**Project Type**: CLI tool
**Target Platform**: Cross-platform (Windows, macOS, Linux)

## Phase 0: Research Summary

### R1: Auto-increment vs Hash-based IDs

- **Decision**: Sequential auto-increment (Option A from issue #66)
- **Rationale**: Simpler, human-readable, sufficient for current project scale (~600 features). Hash-based IDs (Option B) add complexity without proportional benefit.
- **Alternatives considered**: 8-char hex prefix (Option B), hybrid support (Option C)

### R2: Collision detection scope

- **Decision**: 3-layer check — local directories, local branches, remote branches
- **Rationale**: Covers all possible collision sources. Remote check is best-effort (network failures don't block).
- **Alternatives considered**: Local-only (too narrow), centralized registry (too complex)

### R3: Prior work

- **600-bash-cli-migration**: `sl spec create` command, `GetNextFeatureNum()`, `CheckFeatureCollision()` already exist
- **601-cli-skills**: `specledger.specify` skill references `--number` flag
- Only need to extend existing code, not rewrite

## Phase 1: Design

### Files to modify

| File | Change |
|------|--------|
| `pkg/cli/commands/spec_create.go` | Make `--number` optional, add auto-generate logic, add `FEATURE_ID` output |
| `pkg/cli/spec/collision.go` | Add `GetNextAvailableNum()` function |
| `.claude/commands/specledger.specify.md` | Simplify step 2 (remove manual number scan) |
| `pkg/embedded/templates/specledger/commands/specledger.specify.md` | Same as above (embedded template) |

### New files

| File | Purpose |
|------|---------|
| `pkg/cli/spec/collision_test.go` | Unit tests for `GetNextFeatureNum`, `GetNextAvailableNum`, `CheckFeatureCollision` |

### Architecture

```
sl spec create --short-name "feature"
     │
     ├─ --number provided?
     │   ├─ YES → CheckFeatureCollision(number)
     │   │         ├─ OK → use number
     │   │         └─ COLLISION → suggest GetNextAvailableNum()
     │   └─ NO  → GetNextAvailableNum()
     │             ├─ GetNextFeatureNum() → scan specledger/ dirs → max+1
     │             └─ Loop: CheckFeatureCollision() until no collision
     │                       (local dirs + local branches + remote branches)
     │
     └─ GenerateBranchName(shortName, number) → create branch + spec dir
```

## Phase 2: Work Breakdown

### Task 1: Make --number optional (FR-001, FR-003)
- Remove `--number` required validation
- When empty, call `GetNextAvailableNum()`
- Print auto-assigned number to stderr (non-JSON mode)

### Task 2: Add GetNextAvailableNum (FR-002)
- Start from `GetNextFeatureNum()` result
- Loop `CheckFeatureCollision()` up to 100 attempts
- Increment on collision until clear number found

### Task 3: Collision suggestion (FR-004)
- When manual `--number` collides, suggest next available
- Error message includes both collision detail and suggestion

### Task 4: FEATURE_ID in JSON output (FR-005)
- Add `FeatureID` field to `SpecCreateOutput` struct
- Set to `branchName` (e.g., "604-auto-spec-numbers")

### Task 5: Update AI skill docs (FR-007)
- Simplify step 2 in specledger.specify
- Remove manual directory scanning instructions
- Document new JSON output fields

### Task 6: Unit tests
- `TestGetNextFeatureNum_EmptyDir`
- `TestGetNextFeatureNum_WithExistingFeatures`
- `TestGetNextFeatureNum_SkipsNonFeatureDirs`
- `TestGetNextAvailableNum_NoCollision`
- `TestGetNextAvailableNum_SkipsCollisions`
- `TestCheckFeatureCollision_NoCollision`
- `TestCheckFeatureCollision_HasCollision`
- `TestParseFeatureNum`

## Success Criteria

- [x] SC-001: `sl spec create --short-name "test"` succeeds without `--number`
- [x] SC-002: Auto-assigned numbers skip collisions (verified by tests)
- [x] SC-003: AI skill docs updated, no manual number logic needed
- [x] SC-004: `sl spec create --number 42 --short-name "test"` still works
- [x] SC-005: Feature creation completes in <5 seconds (including remote check)
