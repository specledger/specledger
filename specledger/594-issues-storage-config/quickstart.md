# Quickstart: Issues Storage Configuration

**Feature**: 594-issues-storage-config
**Date**: 2026-02-20

## Prerequisites

- Go 1.24.2 installed
- Repository cloned and on branch `594-issues-storage-config`
- `sl` binary built: `go build -o sl ./cmd/sl`

## Manual Testing

### Test 1: Lock File Naming

Verify lock files are named without leading dot.

```bash
# Build the binary
go build -o sl ./cmd/sl

# Create a test issue
./sl issue create --title "Test lock naming" --type task --spec 594-issues-storage-config

# Verify lock file name (should be issues.jsonl.lock, NOT .issues.jsonl.lock)
ls -la specledger/594-issues-storage-config/

# Expected: issues.jsonl and issues.jsonl.lock (no leading dot on lock)
```

### Test 2: Default Artifact Path

Verify default behavior with no custom artifact_path.

```bash
# Check current artifact_path setting
cat specledger/specledger.yaml | grep artifact_path

# Should show: artifact_path: specledger/

# Create issue
./sl issue create --title "Test default path" --type task

# List issues
./sl issue list

# Verify file location
ls -la specledger/594-issues-storage-config/issues.jsonl
```

### Test 3: Custom Artifact Path

Verify issues stored in custom artifact_path.

```bash
# Create a temporary directory for testing
mkdir -p /tmp/sl-test-project
cd /tmp/sl-test-project

# Initialize a new project with custom artifact_path
echo 'version: 1.0.0
project:
  name: test-project
  short_code: tp
  created: 2026-02-20T00:00:00Z
  modified: 2026-02-20T00:00:00Z
playbook:
  name: specledger
  version: 1.0.0
task_tracker:
  choice: builtin
artifact_path: docs/specs/' > specledger.yaml

# Create a spec directory
mkdir -p docs/specs/010-test-spec

# Create an issue
/path/to/sl issue create --title "Test custom path" --type task --spec 010-test-spec

# Verify file location
ls -la docs/specs/010-test-spec/

# Expected: issues.jsonl and issues.jsonl.lock in docs/specs/010-test-spec/
```

### Test 4: Gitignore Pattern

Verify lock files are ignored by git.

```bash
# Back to repo root
cd /path/to/specledger

# Add pattern to .gitignore if not present
grep -q "issues.jsonl.lock" .gitignore || echo "issues.jsonl.lock" >> .gitignore

# Create an issue to generate lock file
./sl issue create --title "Test gitignore" --type task

# Check git status
git status

# Expected: issues.jsonl.lock should NOT appear as untracked
```

### Test 5: List All with Custom Path

Verify --all flag respects custom artifact_path.

```bash
# In test project with custom path
cd /tmp/sl-test-project

# Create issues in multiple specs
mkdir -p docs/specs/020-another-spec
/path/to/sl issue create --title "First issue" --type task --spec 010-test-spec
/path/to/sl issue create --title "Second issue" --type task --spec 020-another-spec

# List all issues
/path/to/sl issue list --all

# Expected: Both issues listed, searched in docs/specs/
```

### Test 6: Missing specledger.yaml

Verify fallback to default path.

```bash
# Create temp directory without specledger.yaml
mkdir -p /tmp/sl-no-config
cd /tmp/sl-no-config

# Create a spec directory in default location
mkdir -p specledger/010-test

# Try to create issue (should use default path)
/path/to/sl issue create --title "Test fallback" --type task --spec 010-test

# Verify file location
ls -la specledger/010-test/

# Expected: issues.jsonl created in specledger/010-test/
```

## Cleanup

```bash
# Remove test directories
rm -rf /tmp/sl-test-project /tmp/sl-no-config
```

## Success Criteria

- [ ] Lock file named `issues.jsonl.lock` (no leading dot)
- [ ] Issues stored in configured `artifact_path`
- [ ] Default path (`specledger/`) works when no config
- [ ] Lock files ignored by git
- [ ] `--all` flag searches in correct artifact path
- [ ] Graceful fallback when specledger.yaml missing
