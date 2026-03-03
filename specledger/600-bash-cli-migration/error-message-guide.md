# Feature Detection Error Messages

## CLI Error Messages (for AI Agents to Parse)

The CLI returns structured error messages that AI agents can parse to provide fuzzy matching.

### Error Message Format

```
Error: <error_type>
Branch: <current_branch>
Feature Dir: <expected_path>

Available features:
- <feature_1>
- <feature_2>
- ...

Suggestion: <helpful_suggestion>
```

### Example 1: Feature Directory Not Found

```
Error: feature directory not found
Branch: feature/add-oauth
Feature Dir: /path/to/repo/specledger/feature/add-oauth

Available features:
- 600-bash-cli-migration
- 599-alignment
- 598-sdd-workflow-streamline
- 601-cli-skills

Suggestion: Set SPECIFY_FEATURE to thefeature-name` or checkout a matching branch
```

**AI Agent Action**:
1. Parse "Branch: feature/add-oauth"
2. Parse "Available features" list
3. Fuzzy match "add-oauth" against features
4. Present options to user

### Example 2: Detached HEAD

```
Error: detached HEAD state
Branch: HEAD (detached)
Commit: abc12345

Suggestion: Checkout a feature branch or set SPECIFY_FEATURE environment variable
```

**AI Agent Action**:
1. Detect detached HEAD state
2. List recent branches: `git branch --sort=-committerdate | head -5`
3. Suggest checkout or set SPECIFY_FEATURE

### Example 3: Not a Git Repository

```
Error: failed to open git repository
Path: /current/working/directory

Suggestion: Run from within a git repository
```

**AI Agent Action**:
1. Check if parent directory is a git repo
2. Suggest navigating to git repo root
3. Or suggest `git init` if appropriate

## Parsing Error Messages (for AI Agents)

### Python Example

```python
import re

def parse_cli_error(error_message):
    """Parse CLI error to extract actionable information"""
    
    result = {
        'error_type': None,
        'branch': None,
        'available_features': [],
        'suggestion': None
    }
    
    # Extract error type
    if match := re.search(r'^Error: (.+)$', error_message, re.MULTILINE):
        result['error_type'] = match.group(1).strip()
    
    # Extract branch
    if match := re.search(r'^Branch: (.+)$', error_message, re.MULTILINE):
        result['branch'] = match.group(1).strip()
    
    # Extract available features
    features_section = re.search(r'Available features:\n((?:- .+\n?)+)', error_message)
    if features_section:
        features_text = features_section.group(1)
        result['available_features'] = [
            line.strip('- ').strip()
            for line in features_text.split('\n')
            if line.strip().startswith('-')
        ]
    
    # Extract suggestion
    if match := re.search(r'^Suggestion: (.+)$', error_message, re.MULTILINE):
        result['suggestion'] = match.group(1).strip()
    
    return result

# Usage
error_msg = """
Error: feature directory not found
Branch: feature/add-oauth
Feature Dir: /path/to/repo/specledger/feature/add-oauth

Available features:
- 600-bash-cli-migration
- 599-alignment
- 601-cli-skills

Suggestion: Set SPECIFY_FEATURE or checkout matching branch
"""

parsed = parse_cli_error(error_msg)
print(parsed['branch'])              # "feature/add-oauth"
print(parsed['available_features'])  # ["600-bash-cli-migration", "599-alignment", "601-cli-skills"]
```

### Bash Example

```bash
#!/bin/bash
# Parse CLI error and do fuzzy matching

parse_error() {
    local error_msg="$1"
    
    # Extract branch
    branch=$(echo "$error_msg" | grep "^Branch:" | cut -d' ' -f2-)
    
    # Extract available features
    mapfile -t features < <(echo "$error_msg" | grep "^- " | sed 's/^- //')
    
    # Fuzzy match
    echo "Current branch: $branch"
    echo "Available features:"
    printf '  %s\n' "${features[@]}"
    
    # Do fuzzy matching here...
}

# Capture error
error=$(sl spec info 2>&1)
if [[ $? -ne 0 ]]; then
    parse_error "$error"
fi
```

## Fuzzy Matching Algorithm

```python
def fuzzy_match_features(branch_name, available_features):
    """
    Match branch name to features using keyword overlap.
    
    Returns list of (feature, score) tuples sorted by relevance.
    """
    # Normalize branch name
    branch_keywords = set(
        branch_name.lower()
        .replace('-', ' ')
        .replace('/', ' ')
        .split()
    )
    
    # Filter out common words
    stop_words = {'feature', 'fix', 'add', 'update', 'the', 'a', 'an', 'for', 'to'}
    branch_keywords -= stop_words
    
    matches = []
    
    for feature in available_features:
        # Extract description (remove number prefix)
        desc = '-'.join(feature.split('-')[1:]).lower()
        feature_keywords = set(desc.replace('-', ' ').split())
        feature_keywords -= stop_words
        
        # Calculate overlap
        common = branch_keywords & feature_keywords
        if not common:
            continue
        
        # Score = overlap / max(keywords)
        score = len(common) / max(len(branch_keywords), len(feature_keywords))
        
        if score > 0.2:  # Minimum threshold
            matches.append((feature, score))
    
    return sorted(matches, key=lambda x: x[1], reverse=True)

# Example
branch = "feature/add-oauth"
features = ["600-oauth-integration", "599-oauth-login", "601-auth-fixes", "600-bash-cli-migration"]

matches = fuzzy_match_features(branch, features)
# Returns: [("600-oauth-integration", 0.67), ("599-oauth-login", 0.5)]
```

## AI Agent Response Templates

### Template 1: Single Match Found

```markdown
I detected that you're on branch `{branch}` but there's no exact feature match.

However, I found a similar feature:
- **{feature}** ({score}% match)

Would you like me to:
1. Set `SPECIFY_FEATURE={feature}` for this session
2. Checkout the feature branch
3. Create a new feature

Which would you prefer?
```

### Template 2: Multiple Matches Found

```markdown
I detected that you're on branch `{branch}` but there's no exact feature match.

I found several similar features:
1. **{feature_1}** ({score_1}% match)
2. **{feature_2}** ({score_2}% match)
3. **{feature_3}** ({score_3}% match)

Which feature are you working on? (Enter 1-3 or type feature name)
```

### Template 3: No Matches Found

```markdown
I detected that you're on branch `{branch}` but there's no matching feature.

Available features:
- {feature_1}
- {feature_2}
- {feature_3}

Would you like me to:
1. Create a new feature with `sl spec create`
2. Set `SPECIFY_FEATURE` to an existing feature
3. List all features with more details

What would you like to do?
```

## Integration Checklist

For AI agents integrating feature detection:

- [ ] Parse CLI error messages
- [ ] Extract branch name and available features
- [ ] Implement fuzzy matching algorithm
- [ ] Present options to user with match scores
- [ ] Allow manual override
- [ ] Set SPECIFY_FEATURE based on user choice
- [ ] Provide helpful suggestions

## Error Message Standard

All CLI errors follow this format:

```
Error: <type>
[Context lines]
Available features:
- <list>
Suggestion: <action>
```

**Context lines** vary by error type:
- `feature directory not found`: Branch, Feature Dir
- `detached HEAD state`: Commit hash
- `not a git repository`: Current path

This structured format makes parsing reliable for AI agents.
