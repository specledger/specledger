---
description: Remove a specification dependency from the project
---

Remove an existing dependency by URL or alias.

## User Input

```text
$ARGUMENTS
```

**Argument is the dependency URL or alias to remove.**

## Quick Usage

```bash
sl deps remove git@github.com:org/specs.git
sl deps remove api-specs  # Using alias
```

## Execution Flow

1. **Identify dependency**:
   - Check if argument matches an existing URL or alias
   - Show dependency details for confirmation

2. **Confirm removal**:
   - Display what will be removed
   - Ask for confirmation unless `--force` flag used
   - Warn about potential impacts

3. **Remove from metadata**:
   - Update `specledger/specledger.yaml`
   - Remove dependency from dependencies array
   - Save updated metadata

4. **Report success**:
   - Show what was removed
   - Note: Local checkout in `specledger/deps/` is not removed

## Examples

```bash
# Remove by URL
sl deps remove git@github.com:org/specs.git

# Remove by alias
sl deps remove api-specs

# Remove without confirmation
sl deps remove git@github.com:org/specs.git --force
```

## Confirmation Prompt

```
Removing dependency:
  URL: git@github.com:org/specs.git
  Alias: api-specs
  Branch: main
  Path: openapi.yaml

This will remove the dependency from specledger/specledger.yaml.
Local checkout in specledger/deps/ will be preserved.

Remove? [y/N]:
```

## Error Handling

**Dependency not found:**
```
Error: dependency not found: git@github.com:org/specs.git
```
Solution: Check `sl deps list` for the correct URL/alias.

**Not in a SpecLedger project:**
```
Error: failed to find project root: not in a SpecLedger project
```
Solution: Navigate to your project directory.

## When to Remove Dependencies

Remove a dependency when:
- Specification is no longer referenced
- Project has been deprecated or moved
- Replacing with a different dependency
- Cleaning up unused dependencies

## After Removal

Consider:
- Run `sl deps list` to verify removal
- Check if any code references the removed spec
- Clean up local checkout: `rm -rf specledger/deps/<alias>`
