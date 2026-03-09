# Feature Context Detection

## Overview

The CLI uses a **4-step fallback chain** for feature context detection (per 598-sdd-workflow-streamline US12, D9). This enables `sl` commands to work on branches that don't follow the SpecLedger naming convention (e.g., created by GitHub UI, Jira, or Linear).

## Detection Priority (4-Step Fallback Chain)

```
Step 0: SPECIFY_FEATURE env var (highest priority - manual override)
├─ Set: Use that feature name, skip all other steps
└─ Not set → Step 1

Step 1: Regex Match
├─ Branch matches ^\d{3,}-[a-z0-9-]+$
│   └─ Use branch name as feature
└─ No match → Step 2

Step 2: YAML Alias Lookup
├─ Check specledger.yaml for branch_aliases.<branch_name>
│   └─ Use aliased feature name
└─ No alias → Step 3

Step 3: Git Heuristic
├─ Diff branch against base (main/master)
├─ Find specledger/<dir>/ paths modified
│   ├─ Exactly one match → Use that feature
│   └─ Multiple/zero matches → Step 4
└─ Step 4

Step 4: Interactive Prompt (or --spec override)
├─ Interactive mode: List available specs, prompt user, save alias to yaml
└─ Non-interactive: Require --spec flag or fail
```

---

## Step 0: Environment Variable (Manual Override)

The `SPECIFY_FEATURE` environment variable bypasses all detection steps. Use it when:
- Working on non-feature branches (main, develop)
- Using custom branch naming (JIRA, Linear)
- In CI/CD pipelines
- Quickly switching between features

### Setting SPECIFY_FEATURE

```bash
# Set for current session
export SPECIFY_FEATURE=600-bash-cli-migration

# Or inline for single command
SPECIFY_FEATURE=600-bash-cli-migration sl spec info --json
```

### Use Cases

**Working on non-feature branches**:
```bash
$ git checkout main
$ export SPECIFY_FEATURE=600-bash-cli-migration
$ sl spec info --json
# Works! Shows context for 600-bash-cli-migration
```

**CI/CD pipelines**:
```bash
# In GitHub Actions, GitLab CI, etc.
export SPECIFY_FEATURE=${FEATURE_NUMBER}-$(echo "$FEATURE_NAME" | tr ' ' '-')
sl spec info --json
```

**Custom branch naming**:
```bash
# Your team uses JIRA-style branches
$ git checkout feature/PROJ-123
$ export SPECIFY_FEATURE=600-bash-cli-migration
$ sl spec info
# Works! Uses 600-bash-cli-migration context
```

---

## Step 1: Regex Match (Primary)

### Feature Branch Pattern

**Pattern**: `^\d{3,}-[a-z0-9-]+$` (3+ digits, hyphen, lowercase alphanumeric)

**Valid Examples**:
- ✅ `600-bash-cli-migration`
- ✅ `599-alignment`
- ✅ `100-login-feature`
- ✅ `9999-very-long-feature-name`

**Invalid Examples**:
- ❌ `main` (no digits)
- ❌ `develop` (no digits)
- ❌ `feature-123` (digits after hyphen)
- ❌ `12-too-short` (only 2 digits)
- ❌ `feature/PROJ-123` (doesn't match pattern)

### Example Usage

```bash
$ git checkout 600-bash-cli-migration
$ sl spec info
# Works automatically! No env var needed
```

---

## Step 2: YAML Alias Lookup

Branch aliases are stored in `specledger.yaml` and version-controlled. This allows teams to use any branch naming convention while still mapping to spec directories.

### specledger.yaml Format

```yaml
branch_aliases:
  feature/PROJ-123: 600-bash-cli-migration
  johns-auth-work: 042-auth-improvements
  bugfix/login-redirect: 100-login-feature
```

### Example Usage

```bash
$ git checkout feature/PROJ-123
$ sl spec info
# Looks up alias: feature/PROJ-123 → 600-bash-cli-migration
# Shows context for 600-bash-cli-migration
```

### Adding Aliases

Aliases are automatically added by Step 4 (interactive prompt) or can be manually edited:

```bash
# Edit specledger.yaml directly
vim specledger.yaml

# Or use sl spec alias (future command)
sl spec alias feature/PROJ-123 600-bash-cli-migration
```

---

## Step 3: Git Heuristic

When no alias exists, the CLI analyzes git history to infer which spec the branch is working on.

### Algorithm

1. Find base branch (main, master, or develop)
2. Run `git diff --name-only base...HEAD`
3. Parse paths matching `specledger/<dir>/`
4. If exactly one unique directory found, use it
5. Otherwise, proceed to Step 4

### Example

```bash
$ git checkout johns-auth-work
$ sl spec info
# Diffs against main, finds:
#   specledger/042-auth-improvements/spec.md
#   specledger/042-auth-improvements/plan.md
# Exactly one spec directory → uses 042-auth-improvements
```

### Limitations

- Fails if branch touches multiple specs
- Fails if no specledger files modified
- May use wrong spec if base branch is incorrect

---

## Step 4: Interactive Prompt

When all automatic detection fails, the CLI prompts the user (in interactive mode).

### Behavior

```bash
$ sl spec info
Could not auto-detect feature context for branch: "johns-branch"

Available specs:
  1. 598-sdd-workflow-streamline
  2. 599-alignment
  3. 600-bash-cli-migration
  4. 601-cli-skills

Which spec are you working on? [1-4]: 3

Saving alias to specledger.yaml...
Feature: 600-bash-cli-migration
FEATURE_DIR: /path/to/repo/specledger/600-bash-cli-migration
...
```

### Non-Interactive Mode

In CI/CD or when `--spec` flag is provided:

```bash
$ sl spec info --spec 600-bash-cli-migration
# Uses provided spec, skips detection
```

```bash
$ sl spec info  # No --spec, non-interactive
Error: Could not detect feature context. Use --spec flag to specify.
```

---

## Feature Context Fields

When detected successfully, `FeatureContext` contains:

```go
type FeatureContext struct {
    RepoRoot   string  // /path/to/repo
    Branch     string  // 600-bash-cli-migration
    FeatureDir string  // /path/to/repo/specledger/600-bash-cli-migration
    SpecFile   string  // /path/to/repo/specledger/600-bash-cli-migration/spec.md
    PlanFile   string  // /path/to/repo/specledger/600-bash-cli-migration/plan.md
    TasksFile  string  // /path/to/repo/specledger/600-bash-cli-migration/tasks.md
    HasGit     bool    // true
}
```

---

## Commands Using Feature Context

All these commands use `DetectFeatureContext()`:

| Command | Uses Context For |
|---------|-----------------|
| `sl spec info` | Display feature context |
| `sl spec setup-plan` | Create plan.md in feature directory |
| `sl context update` | Update agent files with plan metadata |
| `sl issue list` | Filter issues by spec |
| `sl comment list` | Filter comments by spec |
| `sl spec create` | Creates new features (doesn't need context) |

---

## Error Handling

### Error: Detached HEAD

```bash
$ git checkout v1.0.0
$ sl spec info
Error: detached HEAD state - please checkout a feature branch or set SPECIFY_FEATURE env var
```

**Solutions**:
```bash
# Option 1: Checkout feature branch
git checkout 600-bash-cli-migration

# Option 2: Set env var
export SPECIFY_FEATURE=600-bash-cli-migration
```

### Error: Non-Feature Branch (No Alias, No Heuristic Match)

```bash
$ git checkout johns-branch
$ sl spec info
Could not auto-detect feature context for branch: "johns-branch"
Available specs: ...
Which spec are you working on?
```

### Error: Feature Directory Not Found

```bash
$ sl spec info
Error: specledger/600-feature directory does not exist
```

**Solution**:
```bash
# Create feature directory
sl spec create --number 600 --short-name "feature"
```

---

## Implementation Details

### File: pkg/cli/spec/detector.go

```go
func DetectFeatureContext(workDir string) (*FeatureContext, error) {
    // Step 0: Check SPECIFY_FEATURE env var
    if featureBranch := os.Getenv("SPECIFY_FEATURE"); featureBranch != "" {
        return buildContext(repoRoot, featureBranch)
    }

    // Get current branch
    branch := getCurrentBranch(repo)

    // Step 1: Regex match
    if isFeatureBranch(branch) {
        return buildContext(repoRoot, branch)
    }

    // Step 2: YAML alias lookup
    if alias := lookupAlias(branch, config); alias != "" {
        return buildContext(repoRoot, alias)
    }

    // Step 3: Git heuristic
    if spec := inferFromGitDiff(repo, branch); spec != "" {
        return buildContext(repoRoot, spec)
    }

    // Step 4: Interactive prompt (or fail in non-interactive)
    if isInteractive() {
        spec := promptUser(availableSpecs)
        saveAlias(branch, spec, config)
        return buildContext(repoRoot, spec)
    }

    return nil, fmt.Errorf("could not detect feature context. Use --spec flag")
}
```

---

## Best Practices

### ✅ DO

```bash
# Use feature branch naming convention when possible
git checkout -b 600-bash-cli-migration

# Set SPECIFY_FEATURE for non-feature branches
git checkout main
export SPECIFY_FEATURE=600-bash-cli-migration

# Let interactive prompt save aliases for you
sl spec info  # Will prompt and save if needed
```

### ❌ DON'T

```bash
# Don't set SPECIFY_FEATURE permanently in shell profile
echo "export SPECIFY_FEATURE=600-bash-cli-migration" >> ~/.bashrc

# Don't skip alias saves - let the system learn your workflow
```

---

## Comparison: Detection Methods

| Method | Priority | Persistence | Use Case |
|--------|----------|-------------|----------|
| SPECIFY_FEATURE | 0 (highest) | Session | CI/CD, quick override |
| Regex match | 1 | None | Standard workflow |
| YAML alias | 2 | Version-controlled | Custom branch names |
| Git heuristic | 3 | None | Inferring from work |
| Interactive | 4 | Saved to yaml | First-time setup |

---

## Performance

| Step | Time | Notes |
|------|------|-------|
| Env var check | <1ms | String comparison |
| Regex match | <1ms | Pattern match |
| YAML lookup | <5ms | File read + parse |
| Git heuristic | ~50ms | `git diff` execution |
| Interactive | User time | Only on first run |

**Typical detection time**: <10ms (steps 0-2 cover 95% of cases)

---

## Summary

The 4-step fallback chain ensures `sl` commands work regardless of branch naming:

1. **Standard workflow**: Name branches `###-description` → automatic detection
2. **Custom naming**: Add aliases to `specledger.yaml` → works like standard
3. **One-off work**: Git heuristic infers from changes
4. **First time**: Interactive prompt saves alias for future

**Key benefits**:
- ✅ Works with any branch naming convention
- ✅ No manual `adopt` command needed
- ✅ Aliases are version-controlled (team-shared)
- ✅ Graceful fallback for edge cases
