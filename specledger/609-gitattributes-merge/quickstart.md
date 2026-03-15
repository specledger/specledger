# Quickstart: Gitattributes Merge

## What This Feature Does

After `sl init` or `sl doctor --template`, your project's `.gitattributes` will contain `linguist-generated` markers that tell GitHub to collapse machine-generated files (`issues.jsonl`, `tasks.md`) in PR diffs.

## Usage: `sl init` (new projects / first-time setup)

```bash
# New project — creates .gitattributes with sentinel block
sl init

# Existing project with .gitattributes — merges sentinel block (preserves your content)
sl init

# Re-init after upgrade — updates sentinel block content idempotently
sl init

# Force re-init — still merges (does NOT overwrite your .gitattributes)
sl init --force
```

## Usage: `sl doctor --template` (updating templates after CLI upgrade)

```bash
# Non-interactive: updates all templates including .gitattributes sentinel block
sl doctor --template

# Interactive: prompts to update if templates are outdated
sl doctor
# Output:
#   ⚠  Templates: v1.0.0 (CLI: v1.1.0)
#   Template update available: v1.0.0 -> v1.1.0
#   Apply template updates? [y/N]: y
#   ✓ Updated 15 templates (14 new, 1 overwritten)
#   Merged 1 file(s)
```

Both paths use the same merge logic — user content in `.gitattributes` is always preserved, only the sentinel-managed section is updated.

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
# Run unit tests for merge function
go test ./pkg/cli/playbooks/ -run TestMerge -v

# Run integration tests for sl init + sl doctor flows
go test ./tests/integration/ -run TestGitattributes -v

# Run all integration tests
go test ./tests/integration/ -v

# Build and test manually
go build -o sl ./cmd/sl/
mkdir /tmp/test-init && cd /tmp/test-init && git init
echo "*.pbxproj binary" > .gitattributes
/path/to/sl init --short-code test --ci
cat .gitattributes  # Should show both user content and sentinel block
```
