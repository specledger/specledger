# Windows Platform Testing Guide

**Date**: 2026-03-02
**Feature**: 600-bash-cli-migration
**Task**: SL-da1e8a

## Prerequisites

1. **Windows System**: Windows 10 or later
2. **Git for Windows**: Install from https://git-scm.com/download/win
3. **Go 1.24.2+**: Install from https://go.dev/dl/
4. **PowerShell or CMD**: Built-in Windows terminal

## Build Instructions

```powershell
# Clone repository
git clone https://github.com/specledger/specledger.git
cd specledger

# Checkout feature branch
git checkout 600-bash-cli-migration

# Build the CLI
go build -o sl.exe ./cmd/sl/

# Verify build
.\sl.exe version
```

## Test Plan

### Test 1: sl spec info Command

**Test Steps**:
```powershell
# 1. Navigate to feature directory
cd specledger\600-bash-cli-migration

# 2. Run basic command
..\..\sl.exe spec info

# 3. Test JSON output
..\..\sl.exe spec info --json

# 4. Test --require-plan flag
..\..\sl.exe spec info --require-plan

# 5. Test --include-tasks flag
..\..\sl.exe spec info --include-tasks --json

# 6. Test --paths-only flag
..\..\sl.exe spec info --paths-only --json
```

**Expected Results**:
- ✅ No "bash: command not found" errors
- ✅ No "jq: command not found" errors
- ✅ Paths use backslash (\) separators: `C:\path\to\specledger\600-feature`
- ✅ JSON output is valid and parseable
- ✅ All flags work correctly

### Test 2: sl spec create Command

**Test Steps**:
```powershell
# 1. Return to repo root
cd ..\..

# 2. Create test feature (use high number to avoid collision)
.\sl.exe spec create --number 999 --short-name "windows-test-feature" --json

# 3. Verify branch created
git branch --list "*999*"

# 4. Verify directory created
dir specledger\999-windows-test-feature

# 5. Verify spec.md created
type specledger\999-windows-test-feature\spec.md

# 6. Cleanup (switch back to feature branch)
git checkout 600-bash-cli-migration
git branch -D 999-windows-test-feature
Remove-Item -Recurse -Force specledger\999-windows-test-feature
```

**Expected Results**:
- ✅ Branch created successfully
- ✅ Directory created with correct path separators
- ✅ spec.md created from template
- ✅ JSON output shows correct paths
- ✅ No shell command errors

### Test 3: sl spec setup-plan Command

**Test Steps**:
```powershell
# 1. Navigate to feature directory
cd specledger\600-bash-cli-migration

# 2. Backup existing plan.md if needed
Copy-Item plan.md plan.md.bak

# 3. Remove plan.md
Remove-Item plan.md

# 4. Run setup-plan command
..\..\sl.exe spec setup-plan --json

# 5. Verify plan.md created
type plan.md

# 6. Test error on existing file
..\..\sl.exe spec setup-plan

# 7. Restore original if needed
Move-Item plan.md.bak plan.md -Force
```

**Expected Results**:
- ✅ plan.md created from template
- ✅ Error message when plan.md already exists
- ✅ JSON output shows correct path
- ✅ No file operation errors

### Test 4: sl context update Command

**Test Steps**:
```powershell
# 1. Navigate to feature directory
cd specledger\600-bash-cli-migration

# 2. Run context update for claude
..\..\sl.exe context update claude --json

# 3. Verify CLAUDE.md updated
type ..\..\CLAUDE.md

# 4. Check Active Technologies section
Select-String -Path ..\..\CLAUDE.md -Pattern "Active Technologies"

# 5. Test with different agent
..\..\sl.exe context update gemini --json

# 6. Verify GEMINI.md created (if it doesn't exist)
type ..\..\GEMINI.md
```

**Expected Results**:
- ✅ CLAUDE.md updated with Active Technologies
- ✅ Manual additions preserved (if any)
- ✅ No duplicate entries
- ✅ JSON output shows updated file
- ✅ Multiple agent types supported

## Path Separator Verification

**Test All Commands Output Paths**:
```powershell
# Run all commands with JSON output and check paths
.\sl.exe spec info --json | Select-String -Pattern "\\\\"
.\sl.exe spec create --number 998 --short-name "test" --json | Select-String -Pattern "\\\\"
.\sl.exe spec setup-plan --json | Select-String -Pattern "\\\\"
.\sl.exe context update claude --json | Select-String -Pattern "\\\\"
```

**Expected Results**:
- ✅ All paths use backslash (\) on Windows
- ✅ No forward slash (/) in file paths
- ✅ No mixed separators

## Error Handling Tests

### Test Non-Feature Branch
```powershell
# Switch to main branch
git checkout main

# Try to run spec commands (should error)
.\sl.exe spec info

# Expected: Error message about not being on feature branch
```

### Test Missing Files
```powershell
# Navigate to feature directory
cd specledger\600-bash-cli-migration

# Temporarily rename spec.md
Rename-Item spec.md spec.md.tmp

# Try --require-plan flag
..\..\sl.exe spec info --require-plan

# Expected: Error about plan.md missing

# Restore file
Rename-Item spec.md.tmp spec.md
```

## Performance Tests

**Measure execution time**:
```powershell
Measure-Command { .\sl.exe spec info --json }
Measure-Command { .\sl.exe spec create --number 997 --short-name "perf-test" --json }
Measure-Command { .\sl.exe context update claude --json }
```

**Expected Results**:
- ✅ All commands execute in < 1 second
- ✅ No noticeable delays

## Dependency Verification

**Verify no external dependencies**:
```powershell
# Try commands without bash in PATH
$env:PATH = "C:\Windows\System32;C:\Go\bin"

# Run all commands
.\sl.exe spec info --json
.\sl.exe spec create --number 996 --short-name "no-bash" --json
.\sl.exe spec setup-plan --json
.\sl.exe context update claude --json

# All should work without bash, jq, grep, sed
```

**Expected Results**:
- ✅ Commands work without bash
- ✅ Commands work without jq
- ✅ Commands work without grep
- ✅ Commands work without sed

## JSON Output Validation

**Validate JSON is parseable**:
```powershell
# Install jq for validation (optional)
# Or use PowerShell's ConvertFrom-Json

.\sl.exe spec info --json | ConvertFrom-Json
.\sl.exe spec create --number 995 --short-name "json-test" --json | ConvertFrom-Json
.\sl.exe spec setup-plan --json | ConvertFrom-Json
.\sl.exe context update claude --json | ConvertFrom-Json
```

**Expected Results**:
- ✅ All JSON outputs are valid
- ✅ PowerShell can parse them
- ✅ Fields are correctly named

## Cleanup

```powershell
# Remove test branches
git branch -D 995-json-test
git branch -D 996-no-bash
git branch -D 997-perf-test
git branch -D 998-test
git branch -D 999-windows-test-feature

# Remove test directories
Remove-Item -Recurse -Force specledger\995-json-test
Remove-Item -Recurse -Force specledger\996-no-bash
Remove-Item -Recurse -Force specledger\997-perf-test
Remove-Item -Recurse -Force specledger\998-test
Remove-Item -Recurse -Force specledger\999-windows-test-feature

# Return to feature branch
git checkout 600-bash-cli-migration
```

## Test Results Template

| Test | Command | Status | Notes |
|------|---------|--------|-------|
| 1 | sl spec info | ⏸️ Pending | Test on Windows |
| 2 | sl spec create | ⏸️ Pending | Test on Windows |
| 3 | sl spec setup-plan | ⏸️ Pending | Test on Windows |
| 4 | sl context update | ⏸️ Pending | Test on Windows |
| 5 | Path separators | ⏸️ Pending | Verify backslash |
| 6 | No bash dependency | ⏸️ Pending | Remove bash from PATH |
| 7 | No jq dependency | ⏸️ Pending | Remove jq from PATH |
| 8 | JSON validity | ⏸️ Pending | PowerShell parsing |
| 9 | Performance | ⏸️ Pending | < 1 second each |
| 10 | Error handling | ⏸️ Pending | Non-feature branch |

## Notes for Tester

1. **Run all tests in order** to ensure proper cleanup
2. **Save test output** for documentation
3. **Report any errors** with full error messages
4. **Note path separator format** in all outputs
5. **Test on both PowerShell and CMD** if possible
6. **Test on fresh Windows install** if available

## Next Steps After Testing

1. Document test results
2. Report any issues found
3. Update cross-platform verification if needed
4. Close SL-da1e8a task with test results
5. Proceed to SL-c2f9a0 (JSON output verification)
