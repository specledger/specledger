# AI Agent Integration Guide

## Overview

This guide explains how AI agents should handle feature context detection when branches don't exactly match feature names.

## CLI Behavior (Deterministic)

The `sl` CLI uses **exact matching only**:

```
1. Check SPECIFY_FEATURE env var
   └─ Use exact value if set
2. Check git branch name
   └─ Look for exact match in specledger/<branch-name>/
3. Return error if not found
```

**Why exact match?**
- ✅ Fast and deterministic
- ✅ No ambiguity
- ✅ Clear error messages
- ✅ Reliable for automation

## AI Agent Behavior (Fuzzy Matching)

When the CLI returns an error, AI agents should:

### Step 1: Detect Available Features

```bash
# List all available features
ls -d specledger/*/ 2>/dev/null | xargs -n1 basename
```

**Example output**:
```
598-sdd-workflow-streamline
599-alignment
600-bash-cli-migration
601-cli-skills
```

### Step 2: Extract Feature Metadata

```bash
# Get feature number and name
for dir in specledger/*/; do
  name=$(basename "$dir")
  num=$(echo "$name" | cut -d'-' -f1)
  desc=$(echo "$name" | cut -d'-' -f2-)
  echo "$num|$desc|$name"
done
```

**Example output**:
```
598|sdd-workflow-streamline|598-sdd-workflow-streamline
599|alignment|599-alignment
600|bash-cli-migration|600-bash-cli-migration
601|cli-skills|601-cli-skills
```

### Step 3: Fuzzy Match Algorithm

```python
def find_matching_features(branch_name, available_features):
    """
    Find features that might match the branch name.
    
    Args:
        branch_name: Current git branch (e.g., "feature/add-login")
        available_features: List of feature names (e.g., ["600-add-login", "599-auth"])
    
    Returns:
        List of (score, feature_name) tuples sorted by relevance
    """
    matches = []
    
    # Normalize branch name
    branch_parts = set(branch_name.lower().replace('-', ' ').replace('/', ' ').split())
    
    for feature in available_features:
        # Extract description (skip number)
        feature_desc = '-'.join(feature.split('-')[1:]).lower()
        feature_parts = set(feature_desc.replace('-', ' ').split())
        
        # Calculate similarity
        common = branch_parts & feature_parts
        score = len(common) / max(len(branch_parts), len(feature_parts))
        
        if score > 0.3:  # Threshold for relevance
            matches.append((score, feature))
    
    return sorted(matches, reverse=True)
```

### Step 4: Handle Matches

**Multiple matches found**:
```bash
# Branch: feature/add-oauth
# Matches:
#   1. 600-add-oauth-authentication (score: 0.75)
#   2. 599-oauth-integration (score: 0.50)

echo "Found multiple features matching '$branch_name':"
echo "1. 600-add-oauth-authentication (75% match)"
echo "2. 599-oauth-integration (50% match)"
echo ""
echo "Which feature are you working on?"
read -p "Enter number or feature name: " choice
```

**Single match found**:
```bash
# Branch: fix-auth-bug
# Match: 599-authentication-fixes (score: 0.60)

echo "Detected feature: 599-authentication-fixes"
echo "Set SPECIFY_FEATURE=599-authentication-fixes? [Y/n]"
read -p "> " confirm
if [[ "$confirm" != "n" ]]; then
    export SPECIFY_FEATURE=599-authentication-fixes
fi
```

**No matches found**:
```bash
echo "No features found matching branch: $branch_name"
echo ""
echo "Available features:"
ls -d specledger/*/ | xargs -n1 basename
echo ""
echo "Create a new feature or specify manually:"
echo "  sl spec create --number XXX --short-name 'description'"
echo "  export SPECIFY_FEATURE=XXX-description"
```

## Implementation Patterns

### Pattern 1: Automatic Detection (Confident)

```bash
#!/bin/bash
# Auto-detect feature with high confidence (>80%)

branch=$(git rev-parse --abbrev-ref HEAD)
features=$(ls -d specledger/*/ 2>/dev/null | xargs -n1 basename)

# Use AI matching logic here
match=$(fuzzy_match "$branch" "$features" --threshold 0.8)

if [[ -n "$match" ]]; then
    export SPECIFY_FEATURE="$match"
    echo "Auto-detected feature: $match"
else
    # Fall back to interactive prompt
    interactive_feature_selection "$branch" "$features"
fi
```

### Pattern 2: Interactive Prompt (Uncertain)

```bash
#!/bin/bash
# Interactive prompt for uncertain matches

branch=$(git rev-parse --abbrev-ref HEAD
features=$(ls -d specledger/*/ 2>/dev/null | xargs -n1 basename)

matches=$(fuzzy_match "$branch" "$features" --threshold 0.3)

if [[ $(echo "$matches" | wc -l) -eq 1 ]]; then
    # Single match - ask for confirmation
    feature=$(echo "$matches" | head -1)
    echo "Detected feature: $feature"
    read -p "Use this feature? [Y/n] " confirm
    [[ "$confirm" != "n" ]] && export SPECIFY_FEATURE="$feature"
else
    # Multiple matches - let user choose
    echo "Multiple features match branch '$branch':"
    echo "$matches" | nl
    read -p "Select feature (1-$(echo "$matches" | wc -l)): " num
    feature=$(echo "$matches" | sed -n "${num}p")
    export SPECIFY_FEATURE="$feature"
fi
```

### Pattern 3: Specify Command (AI Agent)

```bash
#!/bin/bash
# .specledger/commands/specify-command.sh

# This is called by AI agents when feature context is ambiguous

branch=$(git rev-parse --abbrev-ref HEAD 2>/dev/null)

# Try exact match first
if [[ -d "specledger/$branch" ]]; then
    export SPECIFY_FEATURE="$branch"
    echo "✓ Using exact match: $branch"
    exit 0
fi

# Fuzzy match
features=$(ls -d specledger/*/ 2>/dev/null | xargs -n1 basename)
branch_words=$(echo "$branch" | tr '-/' ' ' | tr '[:upper:]' '[:lower:]')

best_match=""
best_score=0

for feature in $features; do
    feature_desc=$(echo "$feature" | cut -d'-' -f2- | tr '-' ' ' | tr '[:upper:]' '[:lower:]')
    
    # Count matching words
    score=0
    for word in $branch_words; do
        if echo "$feature_desc" | grep -qw "$word"; then
            ((score++))
        fi
    done
    
    if [[ $score -gt $best_score ]]; then
        best_score=$score
        best_match="$feature"
    fi
done

if [[ $best_score -gt 0 ]]; then
    echo "Detected feature: $best_match (matched $best_score words)"
    read -p "Use this feature? [Y/n] " confirm
    if [[ "$confirm" != "n" ]]; then
        export SPECIFY_FEATURE="$best_match"
        echo "✓ Set SPECIFY_FEATURE=$best_match"
    fi
else
    echo "No matching features found"
    echo "Available features:"
    echo "$features" | sed 's/^/  - /'
    echo ""
    echo "Set manually: export SPECIFY_FEATURE=<feature-name>"
fi
```

## Example Workflows

### Example 1: Feature Branch with Slight Variation

**Branch**: `feature/600-bash-migration` (missing "cli")
**Available**: `600-bash-cli-migration`

**AI Agent Logic**:
```bash
# Normalize and compare
branch_norm="bash migration"
feature_norm="bash cli migration"

# Match: 2/3 words = 66% match
# Action: Suggest with confirmation
```

**Output**:
```
Detected feature: 600-bash-cli-migration (66% match)
Use this feature? [Y/n] > y
✓ Set SPECIFY_FEATURE=600-bash-cli-migration
```

### Example 2: Completely Different Name

**Branch**: `fix-login-bug`
**Available**: `599-authentication-system`, `600-bash-cli-migration`

**AI Agent Logic**:
```bash
# Compare keywords
branch_keywords: [fix, login, bug]
feature1_keywords: [authentication, system]  # 0 match
feature2_keywords: [bash, cli, migration]     # 0 match

# No good matches
# Action: Show all features, let user choose
```

**Output**:
```
No automatic match found for branch: fix-login-bug

Available features:
  1. 599-authentication-system
  2. 600-bash-cli-migration

Select feature or create new: _
```

### Example 3: Multiple Candidates

**Branch**: `oauth-integration`
**Available**: `599-oauth-login`, `600-oauth-api`, `601-auth-fixes`

**AI Agent Logic**:
```bash
# Match scores
599-oauth-login:     50% (oauth matches)
600-oauth-api:       50% (oauth matches)
601-auth-fixes:      0%  (no match)

# Multiple candidates
# Action: Present choices ranked by score
```

**Output**:
```
Multiple features match 'oauth-integration':

  1. 599-oauth-login (50% match)
  2. 600-oauth-api (50% match)
  3. Other (specify manually)

Which feature are you working on? [1-3] > 1
✓ Set SPECIFY_FEATURE=599-oauth-login
```

## Integration with AI Prompts

### Claude/Gemini/Copilot Integration

When AI agent detects CLI error:

```markdown
The user is on branch 'feature/add-oauth' but no exact feature match found.

Available features:
- 600-add-oauth-authentication
- 599-oauth-integration
- 601-auth-fixes

I detected likely matches:
1. 600-add-oauth-authentication (75% match - has "add" and "oauth")
2. 599-oauth-integration (50% match - has "oauth")

Which feature are you working on? Or should I:
- Create a new feature with `sl spec create`
- Set SPECIFY_FEATURE manually
```

### Automated AI Response

```python
def handle_feature_detection_error(error_message, branch_name):
    """AI agent handles feature detection failure"""
    
    # Parse available features
    features = list_available_features()
    
    # Fuzzy match
    matches = fuzzy_match(branch_name, features)
    
    if len(matches) == 1 and matches[0].score > 0.7:
        # High confidence - auto-set
        feature = matches[0].name
        return f"export SPECIFY_FEATURE={feature}"
    
    elif len(matches) > 0:
        # Multiple matches - ask user
        options = "\n".join([f"{i+1}. {m.name} ({m.score}% match)" 
                            for i, m in enumerate(matches)])
        return f"Found multiple matches:\n{options}\n\nWhich feature?"
    
    else:
        # No matches - suggest creation
        return f"No matching features found. Create new with:\nsl spec create --number XXX --short-name '{branch_name}'"
```

## Best Practices

### ✅ DO

- Use fuzzy matching in AI prompts/agents
- Present multiple options when uncertain
- Confirm before auto-setting SPECIFY_FEATURE
- Provide clear feedback on match confidence
- Allow manual override

### ❌ DON'T

- Add fuzzy matching to CLI (keep it deterministic)
- Auto-set SPECIFY_FEATURE without confirmation (unless >90% confident)
- Hide match scores from user
- Make assumptions about user intent

## Configuration

### AI Agent Config

```yaml
# .specledger/ai-config.yaml
feature_detection:
  fuzzy_matching:
    enabled: true
    threshold: 0.3        # Minimum match score
    auto_confirm: 0.8     # Auto-set if >80% match
    max_suggestions: 5    # Maximum matches to show
  
  keywords:
    ignore:
      - "feature"
      - "fix"
      - "add"
      - "update"
      - "the"
      - "a"
      - "an"
```

## Summary

**CLI**: Deterministic, exact match only, fast
**AI Agent**: Fuzzy matching, interactive, intelligent

This separation ensures:
- ✅ CLI remains simple and reliable
- ✅ AI agents provide smart assistance
- ✅ Users maintain control
- ✅ Clear separation of concerns
