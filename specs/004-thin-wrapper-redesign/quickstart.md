# Quickstart Guide: SpecLedger Thin Wrapper Redesign

**Feature**: 004-thin-wrapper-redesign
**Date**: 2026-02-05

## Overview

This guide helps developers implement and test the SpecLedger thin wrapper architecture redesign.

## Prerequisites

- Go 1.24+ installed
- Git installed
- SpecLedger repository cloned
- Familiarity with Go, Cobra CLI framework, YAML

## Development Setup

### 1. Check Out Feature Branch

```bash
git checkout 004-thin-wrapper-redesign
```

### 2. Install Dependencies

```bash
# Install Go dependencies
go mod tidy

# Verify build
make build
```

### 3. Run Existing Tests

```bash
# Run all tests
make test

# Run specific package tests
go test ./pkg/cli/commands/...
go test ./pkg/cli/tui/...
```

## Implementation Checklist

Use this checklist to track implementation progress:

### Phase 1: Cleanup & Removal

- [ ] Remove `specledger.specify.md` from `.claude/commands/`
- [ ] Remove `specledger.plan.md` from `.claude/commands/`
- [ ] Remove `specledger.tasks.md` from `.claude/commands/`
- [ ] Remove `specledger.implement.md` from `.claude/commands/`
- [ ] Remove `specledger.analyze.md` from `.claude/commands/`
- [ ] Remove `specledger.clarify.md` from `.claude/commands/`
- [ ] Remove `specledger.checklist.md` from `.claude/commands/`
- [ ] Remove `specledger.constitution.md` from `.claude/commands/`
- [ ] Remove same files from `pkg/embedded/templates/.claude/commands/`
- [ ] Remove `playbookFlag` from `pkg/cli/commands/bootstrap.go`
- [ ] Remove playbook step from `pkg/cli/tui/sl_new.go`
- [ ] Remove `pkg/embedded/playbooks/` directory if it exists
- [ ] Verify no references remain: `grep -r "playbook" pkg/`
- [ ] Run tests to ensure no breakage

### Phase 2: Metadata System

- [ ] Create `pkg/cli/metadata/` directory
- [ ] Create `schema.go` with structs (ProjectMetadata, ProjectInfo, FrameworkInfo, Dependency)
- [ ] Create `yaml.go` with Load/Save functions
- [ ] Create `migration.go` with .mod→YAML conversion
- [ ] Write unit tests for YAML parsing
- [ ] Write unit tests for migration logic
- [ ] Create `pkg/cli/commands/migrate.go` for `sl migrate` command
- [ ] Update `pkg/cli/commands/deps.go` to use YAML instead of .mod
- [ ] Test end-to-end: create project, add dep, migrate, verify YAML

### Phase 3: Prerequisites Checker

- [ ] Create `pkg/cli/prerequisites/` directory
- [ ] Create `checker.go` with CheckPrerequisites() function
- [ ] Implement tool detection functions (isMiseInstalled, isCommandAvailable)
- [ ] Implement EnsurePrerequisites(interactive bool)
- [ ] Write unit tests for tool detection
- [ ] Create `pkg/cli/commands/doctor.go` for `sl doctor` command
- [ ] Add JSON output option to `sl doctor`
- [ ] Test with missing tools (uninstall mise temporarily)
- [ ] Test interactive prompts
- [ ] Test CI mode (non-interactive)

### Phase 4: Bootstrap Integration

- [ ] Update `pkg/cli/commands/bootstrap.go` to call EnsurePrerequisites()
- [ ] Update `pkg/cli/tui/sl_new.go` to add framework selection step
- [ ] Update `pkg/embedded/templates/mise.toml` with commented framework options
- [ ] Create `pkg/embedded/templates/specledger/specledger.yaml` template
- [ ] Update bootstrap to write YAML instead of .mod
- [ ] Update `sl init` command to use YAML
- [ ] Test `sl new` workflow end-to-end
- [ ] Test `sl init` in existing directory
- [ ] Test CI mode: `sl new --ci --project-name test --short-code t`

### Phase 5: Documentation

- [ ] Update README.md with new architecture section
- [ ] Create ARCHITECTURE.md
- [ ] Update installation instructions
- [ ] Add "Choosing an SDD Framework" section
- [ ] Document migration from .mod to YAML
- [ ] Update CONTRIBUTING.md if needed
- [ ] Update CLI help text for all commands
- [ ] Create migration guide for existing users

## Testing Guide

### Unit Tests

Run unit tests for specific packages:

```bash
# Test metadata package
go test -v ./pkg/cli/metadata/

# Test prerequisites package
go test -v ./pkg/cli/prerequisites/

# Test commands
go test -v ./pkg/cli/commands/
```

### Integration Tests

Create integration test scenarios:

```bash
# Test bootstrap workflow
go test -v ./tests/integration/ -run TestBootstrap

# Test migration workflow
go test -v ./tests/integration/ -run TestMigration

# Test doctor command
go test -v ./tests/integration/ -run TestDoctor
```

### Manual Testing Checklist

#### Test: `sl new` (Interactive)

1. Run `./bin/sl new`
2. Enter project name: `test-project`
3. Enter short code: `tp`
4. Select framework: `Spec Kit` (or `OpenSpec`, `Both`, `None`)
5. Confirm installation when prompted
6. Verify:
   - [ ] Project directory created
   - [ ] `specledger/specledger.yaml` exists and is valid YAML
   - [ ] Framework choice recorded in YAML
   - [ ] mise.toml copied with correct comments
   - [ ] Tools installed if prompted

#### Test: `sl new` (CI Mode)

```bash
./bin/sl new --ci --project-name ci-test --short-code ct
```

Verify:
- [ ] Project created without prompts
- [ ] Default framework choice (`none`)
- [ ] YAML file created correctly

#### Test: `sl doctor`

```bash
./bin/sl doctor
```

Verify:
- [ ] Reports all core tools (mise, bd, perles)
- [ ] Reports framework tools (specify, openspec)
- [ ] Clear status indicators (✅/❌)
- [ ] Version numbers shown for installed tools
- [ ] Install instructions shown for missing tools

#### Test: `sl doctor --json`

```bash
./bin/sl doctor --json
```

Verify:
- [ ] Valid JSON output
- [ ] Matches schema in `contracts/doctor-output.json`

#### Test: `sl migrate`

1. Create a test project with old .mod file:
   ```bash
   mkdir test-migrate
   cd test-migrate
   mkdir specledger
   cat > specledger/specledger.mod <<EOF
   # SpecLedger Dependency Manifest v1.0.0
   # Project: test-migrate
   # Short Code: tm
   EOF
   ```

2. Run migration:
   ```bash
   ../bin/sl migrate
   ```

3. Verify:
   - [ ] `specledger.yaml` created
   - [ ] Project name and short code migrated correctly
   - [ ] Framework choice set to `none`
   - [ ] Original .mod file preserved
   - [ ] Success message printed

#### Test: `sl deps` (YAML format)

```bash
cd test-project
./bin/sl deps add git@github.com:example/spec main spec.md --alias example
./bin/sl deps list
```

Verify:
- [ ] Dependency added to `specledger.yaml`
- [ ] YAML remains valid
- [ ] Dependency listed correctly

## Debugging Tips

### Problem: YAML parsing errors

```bash
# Validate YAML syntax
go run ./cmd/main.go doctor --debug

# Check YAML structure
cat specledger/specledger.yaml | python3 -m yaml
```

### Problem: Tool detection fails

```bash
# Check PATH
echo $PATH

# Manually check for tools
which mise
which bd
which specify
which openspec
```

### Problem: Migration fails

```bash
# Enable debug logging
SL_LOG_LEVEL=debug ./bin/sl migrate

# Check .mod file format
cat specledger/specledger.mod
```

## Common Development Tasks

### Add New Field to Metadata Schema

1. Update `pkg/cli/metadata/schema.go`
2. Update `contracts/specledger-schema.yaml`
3. Add YAML tag to struct field
4. Update validation if needed
5. Add unit tests
6. Update documentation

### Add New Prerequisite Tool

1. Update `pkg/cli/prerequisites/checker.go`
2. Add tool to `RequiredTools` or `OptionalTools`
3. Update `sl doctor` output
4. Add to integration tests
5. Update documentation

### Change Framework Options

1. Update `pkg/embedded/templates/mise.toml`
2. Update comments to explain new options
3. Update `pkg/cli/tui/sl_new.go` if adding to TUI
4. Update documentation
5. Test installation with new option

## Before Submitting PR

Run this checklist before creating a pull request:

- [ ] All unit tests pass: `make test`
- [ ] Integration tests pass: `go test ./tests/integration/...`
- [ ] Code formatted: `make fmt`
- [ ] No linter errors: `make vet`
- [ ] Documentation updated (README, ARCHITECTURE)
- [ ] Migration guide written
- [ ] CLI help text updated
- [ ] Manual testing completed (all scenarios above)
- [ ] No breaking changes for existing users (or documented)
- [ ] YAML schema validated
- [ ] Example projects tested

## Getting Help

- **Design questions**: Review `spec.md` and `plan.md`
- **Implementation questions**: Check `data-model.md` and `research.md`
- **Testing issues**: See integration test examples in `tests/`
- **Build issues**: Check `Makefile` and `go.mod`

## Next Steps

After completing implementation:

1. Create PR from `004-thin-wrapper-redesign` to `main`
2. Request review from maintainers
3. Address feedback
4. Merge when approved
5. Create release notes
6. Announce migration guide to users
