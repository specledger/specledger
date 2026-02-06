# SpecLedger Migration Guide

## Overview

This guide helps you migrate from the legacy SpecLedger format (`.mod` files) to the new YAML-based metadata system (`specledger.yaml`).

## What's Changed?

### Before (Legacy)
```
specledger/
├── specledger.mod    # Plain text comments
└── specs/            # Specification files
```

### After (New)
```
specledger/
├── specledger.yaml   # Structured YAML metadata
├── specledger.mod.backup  # Backup of original file
└── specs/            # Specification files
```

## Key Improvements

1. **Structured Data**: YAML provides structured, validated metadata
2. **Framework Tracking**: Records which SDD framework you're using
3. **Dependency Lockfile**: Resolved commit SHAs for reproducibility
4. **Better Validation**: Schema validation prevents errors
5. **Machine-Readable**: Easy to parse by tools and AI agents

## Migration Steps

### Option 1: Automatic Migration (Recommended)

Run the migration command in your project directory:

```bash
cd your-project
sl migrate
```

**What happens**:
1. SpecLedger reads your `specledger.mod` file
2. Creates `specledger.yaml` with migrated data
3. Backs up original file to `specledger.spec.mod.backup`
4. Sets framework choice to `none` (you can change this)

### Option 2: Dry Run

Preview changes without modifying files:

```bash
sl migrate --dry-run
```

### Option 3: Manual Migration

If you prefer to migrate manually:

1. Create `specledger/specledger.yaml`:

```yaml
version: "1.0.0"

project:
  name: your-project-name      # From # Project: in .mod
  short_code: yp               # From # Short Code: in .mod
  created: "2024-01-01T00:00:00Z"  # File creation time
  modified: "2024-01-01T00:00:00Z"  # Current time
  version: "0.1.0"

framework:
  choice: none                 # or: speckit, openspec, both

dependencies: []              # Empty for now, or add from comments
```

2. Verify the file is valid:

```bash
sl doctor
```

## Post-Migration Tasks

### 1. Choose Your SDD Framework (Optional)

Edit `specledger.yaml` to set your preferred framework:

```yaml
framework:
  choice: speckit    # For GitHub Spec Kit
  # or
  choice: openspec   # For OpenSpec
  # or
  choice: both       # To use both frameworks
```

### 2. Remove Duplicate Commands

If you previously had SpecLedger's built-in SDD commands, they have been removed:

- `sl specify` → Use `specify` from Spec Kit
- `sl plan` → Use `specify plan` or OpenSpec
- `sl tasks` → Use `specify tasks` or OpenSpec
- `sl implement` → Use framework-specific commands

### 3. Verify Your Setup

```bash
# Check tool installation
sl doctor

# Verify dependencies
sl deps list

# Test framework (if you chose one)
specify --version    # For Spec Kit
openspec --version  # For OpenSpec
```

## Rollback

If you need to rollback to the `.mod` format:

```bash
# Restore from backup
mv specledger/specledger.spec.mod.backup specledger/specledger.mod

# Remove YAML file
rm specledger/specledger.yaml
```

## Compatibility

### SpecLedger 1.x
- ✅ Reads both `.mod` and `.yaml` formats
- ✅ Writes only `.yaml` for new projects
- ⚠️  Warns when `.mod` is detected

### SpecLedger 2.0 (Future)
- ✅ Auto-migrates `.mod` files on first run
- ✅ Supports reading `.mod` (read-only)
- ⚠️  Deprecates `.mod` format

### SpecLedger 3.0 (Future)
- ❌ Removes `.mod` support entirely

## Troubleshooting

### Migration Fails

**Problem**: `sl migrate` fails with error

**Solution**:
1. Check your `.mod` file has required fields:
   ```
   # Project: your-name
   # Short Code: yn
   ```
2. Ensure you're in the project root directory
3. Use `--dry-run` to preview what will happen

### Framework Not Detected

**Problem**: Framework tools show as not installed

**Solution**:
```bash
# Check what's installed
sl doctor

# Install via mise
mise install

# Or install manually
pipx install git+https://github.com/github/spec-kit.git  # Spec Kit
npm install -g @fission-ai/openspec                      # OpenSpec
```

### Dependencies Missing

**Problem**: Dependencies not showing after migration

**Solution**: The legacy `.mod` format didn't store dependencies. Add them manually:

```bash
sl deps add git@github.com:org/spec main spec.md --alias org
```

## Getting Help

If you encounter issues:

1. Run `sl doctor` to check your setup
2. Check the [ARCHITECTURE.md](ARCHITECTURE.md) for design details
3. Open an issue on GitHub with:
   - Your `.mod` file content (sanitized)
   - Error messages
   - `sl doctor` output

## Summary

| Aspect | Old (.mod) | New (.yaml) |
|--------|-----------|------------|
| Format | Plain text | Structured YAML |
| Validation | None | Schema validation |
| Framework tracking | No | Yes |
| Dependencies | Comments only | Full metadata with SHAs |
| Tool support | Legacy only | Modern + backward compatible |

**Ready to migrate?**

```bash
sl migrate
```

Questions? Join the discussion at https://github.com/specledger/specledger/discussions or visit https://specledger.io/docs
