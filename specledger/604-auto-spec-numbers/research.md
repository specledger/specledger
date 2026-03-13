# Research: Auto-Generate Spec Numbers

## R1: Numbering Strategy

**Decision**: Sequential auto-increment (NNN format, zero-padded to 3 digits)

**Rationale**:
- Human-readable and familiar (matches existing convention)
- Simple implementation using existing `GetNextFeatureNum()`
- Sufficient for project scale (~600 features, 999 max before 4-digit overflow which is handled)

**Alternatives considered**:
- **Hash-based (8-char hex)**: Zero collision risk but poor readability. Overkill for single-team project.
- **Hybrid**: Both formats simultaneously — adds complexity without clear benefit during migration.

## R2: Collision Detection Sources

**Decision**: 3-layer check (local dirs → local branches → remote branches)

**Rationale**:
- Local dirs: Primary source, always available
- Local branches: Catches branches without matching dirs (e.g., deleted spec dir but branch still exists)
- Remote branches: Best-effort — catches numbers used by teammates. Fails silently on network issues.

**Alternatives considered**:
- **Local-only**: Misses remote collisions from teammates
- **Centralized registry**: Too complex, requires server infrastructure

## R3: Prior Work Analysis

| Feature | Relevant Code | Reuse |
|---------|--------------|-------|
| 600-bash-cli-migration | `GetNextFeatureNum()`, `CheckFeatureCollision()` | Extended with `GetNextAvailableNum()` |
| 600-bash-cli-migration | `spec_create.go` command | Modified to make `--number` optional |
| 601-cli-skills | `specledger.specify.md` | Simplified step 2 |
