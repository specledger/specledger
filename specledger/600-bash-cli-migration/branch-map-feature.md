# Branch Map Feature

## Overview

The `branch-map.json` file allows you to map non-feature branches to feature contexts. This is useful when:
- Working on branches that don't follow the `###-description` pattern
- Using custom branch naming conventions
- Mapping main/master to a specific feature for development

## Location

The branch map file should be located at:
```
<repo-root>/.specledger/branch-map.json
```

## Format

```json
{
  "branchToFeature": {
    "main": "600-bash-cli-migration",
    "develop": "599-alignment",
    "custom-branch-name": "598-sdd-workflow-streamline"
  }
}
```

## How It Works

When `DetectFeatureContext()` is called:

1. **First**: Check if current branch matches `###-description` pattern
2. **If not**: Look for branch in `branch-map.json`
3. **If found**: Use the mapped feature name for context
4. **If not found**: Return error (not a feature branch)

## Example Scenarios

### Scenario 1: Working on main branch

You're on `main` branch but want to work on feature `600-bash-cli-migration`:

**branch-map.json**:
```json
{
  "branchToFeature": {
    "main": "600-bash-cli-migration"
  }
}
```

**Result**: `sl spec info` will show context for `600-bash-cli-migration`

### Scenario 2: Custom branch naming

Your team uses `feature/ISSUE-123` naming:

**branch-map.json**:
```json
{
  "branchToFeature": {
    "feature/ISSUE-123": "600-bash-cli-migration",
    "feature/ISSUE-456": "601-another-feature"
  }
}
```

### Scenario 3: Development branch

Map `develop` to a specific feature:

**branch-map.json**:
```json
{
  "branchToFeature": {
    "develop": "599-alignment"
  }
}
```

## When to Use

### ✅ Good Use Cases

- **Main branch development**: Temporarily working on `main` while developing a feature
- **Custom naming conventions**: Team uses different branch naming patterns
- **Legacy branches**: Existing branches that don't match specledger pattern
- **CI/CD branches**: Map `staging`, `production` to features

### ❌ Not Recommended

- **Permanent mappings**: Better to rename branches to match pattern
- **Multiple features**: Don't map one branch to multiple features
- **Team-wide use**: This is intended for individual developer convenience

## Commands That Use Branch Map

All commands that call `DetectFeatureContext()`:

- ✅ `sl spec info`
- ✅ `sl spec setup-plan`
- ✅ `sl context update`
- ⏸️ `sl spec create` (uses collision detection, not context)

## Implementation Details

### File: pkg/cli/spec/detector.go

```go
func DetectFeatureContext(workDir string) (*FeatureContext, error) {
    // ... get branch ...
    
    featureBranch := branch
    if !isFeatureBranch(branch) {
        // Check branch-map.json
        mappedFeature, err := checkBranchMap(repoRoot, branch)
        if err != nil {
            return nil, fmt.Errorf("not a feature branch: %q", branch)
        }
        if mappedFeature != "" {
            featureBranch = mappedFeature
        } else {
            return nil, fmt.Errorf("not a feature branch: %q", branch)
        }
    }
    
    // Use featureBranch for paths...
}
```

### Error Handling

1. **Branch-map.json doesn't exist**: Silently ignore, proceed with branch name check
2. **Branch-map.json is malformed**: Return error with parse details
3. **Branch not in map**: Return error (not a feature branch)

## Security Considerations

- File should be gitignored if containing personal mappings
- Can be committed if containing team-wide conventions
- No sensitive data should be in branch-map.json

## Sample Files

### Minimal branch-map.json

```json
{
  "branchToFeature": {
    "main": "600-bash-cli-migration"
  }
}
```

### Multiple Mappings

```json
{
  "branchToFeature": {
    "main": "600-bash-cli-migration",
    "develop": "599-alignment",
    "staging": "598-sdd-workflow-streamline"
  }
}
```

### With Comments (Not Valid JSON)

```json
{
  "branchToFeature": {
    "main": "600-bash-cli-migration",
    "develop": "599-alignment"
  }
}
```

Note: JSON doesn't support comments. Use a separate README if documentation is needed.

## Troubleshooting

### Error: "not a feature branch"

**Cause**: Branch doesn't match pattern and isn't in branch-map.json

**Solutions**:
1. Add branch to branch-map.json
2. Checkout a feature branch (`###-description`)
3. Create feature branch with `sl spec create`

### Error: "failed to parse branch-map.json"

**Cause**: JSON syntax error in file

**Solution**: Validate JSON with `jq . branch-map.json` or online validator

### Mapped feature not found

**Cause**: Feature directory doesn't exist

**Solution**: Ensure `specledger/<feature-name>/` directory exists

## Migration from Bash Scripts

**Old behavior** (check_mapped_branch in common.sh):
```bash
check_mapped_branch() {
    local branch="$1"
    local map_file="$REPO_ROOT/.specledger/branch-map.json"
    
    if [[ -f "$map_file" ]]; then
        local mapped=$(jq -r ".branchToFeature[\"$branch\"] // empty" "$map_file")
        if [[ -n "$mapped" ]]; then
            echo "$mapped"
            return 0
        fi
    fi
    return 1
}
```

**New behavior** (detector.go):
```go
func checkBranchMap(repoRoot, branch string) (string, error) {
    branchMapPath := filepath.Join(repoRoot, ".specledger", "branch-map.json")
    
    data, err := os.ReadFile(branchMapPath)
    if err != nil {
        if os.IsNotExist(err) {
            return "", nil
        }
        return "", fmt.Errorf("failed to read branch-map.json: %w", err)
    }
    
    var branchMap BranchMap
    if err := json.Unmarshal(data, &branchMap); err != nil {
        return "", fmt.Errorf("failed to parse branch-map.json: %w", err)
    }
    
    if branchMap.BranchToFeature == nil {
        return "", nil
    }
    
    return branchMap.BranchToFeature[branch], nil
}
```

## Benefits

1. **Flexibility**: Work on any branch naming convention
2. **Compatibility**: Support existing team workflows
3. **Convenience**: No need to rename branches
4. **Opt-in**: Only used when needed, doesn't affect standard workflow

## Related Files

- **Implementation**: `pkg/cli/spec/detector.go`
- **Config file**: `.specledger/branch-map.json`
- **Research**: `specledger/600-bash-cli-migration/research.md`
