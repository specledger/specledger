# JSON Output Verification Report

**Date**: 2026-03-02
**Feature**: 600-bash-cli-migration
**Task**: SL-c2f9a0
**Platform Tested**: Linux (Ubuntu)

## JSON Output Format

All 4 commands output JSON in a consistent format:

### Common Patterns

1. **Field Names**: Uppercase with underscores (e.g., `FEATURE_DIR`, `BRANCH_NAME`)
2. **Indentation**: 2-space indentation
3. **No HTML Escaping**: `encoder.SetEscapeHTML(false)`
4. **Arrays**: Used for lists (e.g., `AVAILABLE_DOCS`, `UPDATED_FILES`)

## Command JSON Outputs

### 1. sl spec info --json

**Structure**:
```json
{
  "FEATURE_DIR": "/path/to/specledger/600-feature",
  "BRANCH": "600-feature-name",
  "FEATURE_SPEC": "/path/to/specledger/600-feature/spec.md",
  "AVAILABLE_DOCS": [
    "research.md",
    "data-model.md"
  ]
}
```

**Fields**:
- `FEATURE_DIR` (string): Absolute path to feature directory
- `BRANCH` (string): Feature branch name
- `FEATURE_SPEC` (string): Absolute path to spec.md
- `AVAILABLE_DOCS` (array of strings): Relative paths to documentation files

**Flags**:
- `--json`: Output as JSON
- `--require-plan`: No JSON output change (error on missing plan.md)
- `--require-tasks`: No JSON output change (error on missing tasks.md)
- `--include-tasks`: Adds "tasks.md" to AVAILABLE_DOCS array
- `--paths-only`: Omits AVAILABLE_DOCS field entirely

**Variations**:
```json
// With --include-tasks
{
  "FEATURE_DIR": "/path/to/specledger/600-feature",
  "BRANCH": "600-feature-name",
  "FEATURE_SPEC": "/path/to/specledger/600-feature/spec.md",
  "AVAILABLE_DOCS": [
    "research.md",
    "tasks.md"
  ]
}

// With --paths-only
{
  "FEATURE_DIR": "/path/to/specledger/600-feature",
  "BRANCH": "600-feature-name",
  "FEATURE_SPEC": "/path/to/specledger/600-feature/spec.md"
}
```

### 2. sl spec create --json

**Structure**:
```json
{
  "BRANCH_NAME": "600-feature-name",
  "FEATURE_DIR": "/path/to/specledger/600-feature-name",
  "SPEC_FILE": "/path/to/specledger/600-feature-name/spec.md",
  "FEATURE_NUM": "600"
}
```

**Fields**:
- `BRANCH_NAME` (string): Created branch name
- `FEATURE_DIR` (string): Absolute path to feature directory
- `SPEC_FILE` (string): Absolute path to spec.md
- `FEATURE_NUM` (string): Feature number (as string)

**Flags**:
- `--json`: Output as JSON
- `--number`: Required, no JSON output change
- `--short-name`: Required, no JSON output change

### 3. sl spec setup-plan --json

**Structure**:
```json
{
  "PLAN_FILE": "/path/to/specledger/600-feature/plan.md"
}
```

**Fields**:
- `PLAN_FILE` (string): Absolute path to created plan.md

**Flags**:
- `--json`: Output as JSON

### 4. sl context update --json

**Structure**:
```json
{
  "UPDATED_FILES": [
    "/path/to/CLAUDE.md"
  ]
}
```

**Fields**:
- `UPDATED_FILES` (array of strings): Absolute paths to updated agent files

**Flags**:
- `--json`: Output as JSON
- `--agent`: Agent type (claude, gemini, etc.), no JSON output change

## JSON Validation

### Linux Test Results ✅

All JSON outputs validated on Linux:

```bash
# sl spec info --json
$ ./sl spec info --json | jq .
{
  "FEATURE_DIR": "/home/me322/project/specledger/specledger/600-bash-cli-migration",
  "BRANCH": "600-bash-cli-migration",
  "FEATURE_SPEC": "/home/me322/project/specledger/specledger/600-bash-cli-migration/spec.md",
  "AVAILABLE_DOCS": [
    "checklists/requirements.md",
    "research.md",
    "sessions/2026-03-02-session-1.md",
    ...
  ]
}

# sl spec create --json
$ ./sl spec create --number 998 --short-name "json-test" --json | jq .
{
  "BRANCH_NAME": "998-json-test",
  "FEATURE_DIR": "/home/me322/project/specledger/specledger/998-json-test",
  "SPEC_FILE": "/home/me322/project/specledger/specledger/998-json-test/spec.md",
  "FEATURE_NUM": "998"
}

# sl spec setup-plan --json
$ ./sl spec setup-plan --json | jq .
{
  "PLAN_FILE": "/home/me322/project/specledger/specledger/998-json-test/plan.md"
}

# sl context update --json
$ ./sl context update claude --json | jq .
{
  "UPDATED_FILES": [
    "/home/me322/project/specledger/CLAUDE.md"
  ]
}
```

**Validation**: All outputs pass `jq .` validation ✅

### Cross-Platform Expectations

**Expected Differences**:
1. **Path separators**:
   - Linux/macOS: `/home/user/project/specledger/600-feature`
   - Windows: `C:\Users\user\project\specledger\600-feature`

2. **Absolute paths**:
   - Different root paths on different platforms
   - All other fields identical

**Expected Identical Fields**:
1. **Field names**: All uppercase with underscores
2. **Field types**: Strings and arrays
3. **Array formats**: Same structure
4. **JSON structure**: Identical nesting
5. **Number formats**: Feature numbers as strings
6. **Relative paths**: AVAILABLE_DOCS entries are relative

### Verification Checklist

- [x] JSON is valid on Linux
- [x] Field names are consistent
- [x] Field types are consistent
- [x] Arrays formatted correctly
- [x] No platform-specific assumptions
- [x] No locale-dependent formatting
- [x] No timezone-dependent values
- [ ] JSON valid on macOS (pending manual test)
- [ ] JSON valid on Windows (pending manual test)
- [ ] Path separators correct on each platform (pending manual test)

## Platform-Specific Notes

### Linux (Tested) ✅
- Paths use forward slash (/)
- JSON encoding: UTF-8
- jq validation: PASS
- Python json.tool validation: PASS

### macOS (Pending)
- Expected: Forward slash (/) paths
- Expected: Identical JSON structure
- Expected: Valid JSON output

### Windows (Pending)
- Expected: Backslash (\) paths
- Expected: Identical JSON structure
- Expected: Valid JSON output
- Expected: PowerShell ConvertFrom-Json works

## JSON Encoding Details

**Implementation**:
```go
encoder := json.NewEncoder(os.Stdout)
encoder.SetEscapeHTML(false)  // No HTML escaping
encoder.SetIndent("", "  ")    // 2-space indentation
if err := encoder.Encode(output); err != nil {
    return fmt.Errorf("failed to encode JSON output: %w", err)
}
```

**Character Encoding**: UTF-8
**HTML Escaping**: Disabled
**Indentation**: 2 spaces
**Line Endings**: Platform-native (Go handles this)

## Testing Recommendations

### Automated Testing
1. Run all 4 commands with --json
2. Pipe output through `jq .` (Linux/macOS)
3. Pipe output through PowerShell `ConvertFrom-Json` (Windows)
4. Compare field names programmatically
5. Validate JSON schema

### Manual Testing
1. Run commands on each platform
2. Save JSON outputs to files
3. Compare with diff tool
4. Verify only paths differ
5. Verify JSON parsing works

## Conclusion

**Status**: ✅ PASS (Linux)

All JSON outputs on Linux:
- ✅ Are valid JSON
- ✅ Follow consistent format
- ✅ Use correct field names
- ✅ Have correct data types
- ✅ Are parseable by jq
- ✅ Have no platform-specific assumptions

**Next Steps**:
1. Manual testing on macOS
2. Manual testing on Windows (see windows-testing-guide.md)
3. Compare JSON outputs across platforms
4. Close SL-c2f9a0 task with verification results

## Appendix: JSON Schemas

### SpecInfoOutput Schema
```json
{
  "type": "object",
  "properties": {
    "FEATURE_DIR": {"type": "string"},
    "BRANCH": {"type": "string"},
    "FEATURE_SPEC": {"type": "string"},
    "AVAILABLE_DOCS": {
      "type": "array",
      "items": {"type": "string"}
    }
  },
  "required": ["FEATURE_DIR", "BRANCH", "FEATURE_SPEC"]
}
```

### SpecCreateOutput Schema
```json
{
  "type": "object",
  "properties": {
    "BRANCH_NAME": {"type": "string"},
    "FEATURE_DIR": {"type": "string"},
    "SPEC_FILE": {"type": "string"},
    "FEATURE_NUM": {"type": "string"}
  },
  "required": ["BRANCH_NAME", "FEATURE_DIR", "SPEC_FILE", "FEATURE_NUM"]
}
```

### SpecSetupPlanOutput Schema
```json
{
  "type": "object",
  "properties": {
    "PLAN_FILE": {"type": "string"}
  },
  "required": ["PLAN_FILE"]
}
```

### ContextUpdateOutput Schema
```json
{
  "type": "object",
  "properties": {
    "UPDATED_FILES": {
      "type": "array",
      "items": {"type": "string"}
    }
  },
  "required": ["UPDATED_FILES"]
}
```
