# Quickstart: Gitattributes Merge

## What This Feature Does

After `sl init`, your project's `.gitattributes` will contain `linguist-generated` markers that tell GitHub to collapse machine-generated files (`issues.jsonl`, `tasks.md`) in PR diffs.

## Usage

```bash
# New project — creates .gitattributes with sentinel block
sl init

# Existing project with .gitattributes — merges sentinel block (preserves your content)
sl init

# Re-init after upgrade — updates sentinel block content idempotently
sl init
```

## What Gets Added

```gitattributes
# Your existing content is preserved above...

# >>> specledger-generated
# Auto-managed by specledger - do not edit this section
specledger/*/issues.jsonl linguist-generated=true
specledger/*/tasks.md linguist-generated=true
# <<< specledger-generated
```

## What Files Are Collapsed in PRs

| File | Collapsed? | Why |
|------|-----------|-----|
| `issues.jsonl` | Yes | Machine-generated issue index |
| `tasks.md` | Yes | Machine-generated task index |
| `spec.md` | No | Reviewable design artifact |
| `plan.md` | No | Reviewable design artifact |
| `checklists/*.md` | No | Reviewable content |
| `research.md` | No | Reviewable content |
| `.claude/commands/*` | No | Reviewable content |
| `.claude/skills/*` | No | Reviewable content |

## Development

```bash
# Run merge tests
go test ./pkg/cli/playbooks/ -run TestMerge -v

# Build and test manually
go build -o sl ./cmd/sl/
mkdir /tmp/test-init && cd /tmp/test-init && git init
echo "*.pbxproj binary" > .gitattributes
/path/to/sl init
cat .gitattributes  # Should show both user content and sentinel block
```
