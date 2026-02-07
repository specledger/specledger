# Quickstart: Embedded Templates

**Feature**: 001-embedded-templates
**Phase**: 1 - Design & Contracts
**Date**: 2026-02-07

## Overview

This guide shows how to use the embedded templates feature in SpecLedger after implementation.

## Prerequisites

- SpecLedger CLI installed (`sl` command available)
- Go 1.24+ (for development)
- Git (for project initialization)

## Basic Usage

### Create a New Project with Templates

```bash
# Interactive mode (select framework in TUI)
sl new

# Non-interactive with Spec Kit framework
sl new --ci --project-name myproject --short-code mp --framework speckit

# With custom directory
sl new --ci --project-name myproject --short-code mp --project-dir ~/projects --framework speckit
```

**What happens**:
1. Project directory is created
2. `specledger/specledger.yaml` is created
3. Templates are automatically copied from embedded Spec Kit playbook:
   - `.claude/commands/` - Claude Code commands
   - `.claude/skills/` - Claude Code skills
   - `specledger/scripts/` - Helper bash scripts
   - `specledger/templates/` - File templates
   - `.beads/` - Issue tracker configuration

### Initialize an Existing Repository

```bash
# Navigate to your repository
cd my-existing-project

# Initialize with Spec Kit templates
sl init --framework speckit
```

**What happens**:
1. `specledger/specledger.yaml` is created
2. Templates are copied to the existing repository

### List Available Templates

```bash
# List all embedded templates
sl template list

# Output example:
# Available Templates:
#   speckit    Spec Kit playbook templates    Framework: speckit    Version: 1.0.0
```

## Template Structure

After creating a project, your directory structure includes:

```text
myproject/
├── .claude/
│   ├── commands/
│   │   ├── specledger.adopt.md
│   │   ├── specledger.specify.md
│   │   ├── specledger.plan.md
│   │   ├── specledger.tasks.md
│   │   └── ...
│   └── skills/
│       └── bd-issue-tracking/
│           ├── SKILL.md
│           └── references/
├── .beads/
│   ├── config.yaml
│   ├── issues.jsonl
│   └── metadata.json
├── specledger/
│   ├── memory/
│   │   └── constitution.md
│   ├── scripts/
│   │   ├── bash/
│   │   │   ├── create-new-feature.sh
│   │   │   ├── adopt-feature-branch.sh
│   │   │   └── ...
│   └── templates/
│       ├── spec-template.md
│       ├── plan-template.md
│       └── tasks-template.md
└── specledger/
    └── specledger.yaml
```

## CLI Reference

### `sl new` - Create New Project

```bash
sl new [flags]

Flags:
  --ci                      Non-interactive mode (required for CI/CD)
  --project-name string     Project name (required in CI mode)
  --short-code string       Short code for issues (required in CI mode)
  --project-dir string      Parent directory (default: from config)
  --framework string        SDD framework: speckit, openspec, both, none (default: speckit)
```

### `sl init` - Initialize Existing Repository

```bash
sl init [flags]

Flags:
  --framework string    SDD framework: speckit, openspec, both, none (default: speckit)
  --short-code string   Short code for issues
```

### `sl template list` - List Available Templates

```bash
sl template list [flags]

Flags:
  --json    Output in JSON format
```

## Development Quickstart

### Project Structure for Template Development

```text
specledger/
├── pkg/
│   ├── cli/
│   │   └── templates/          # Template management package
│   │       ├── templates.go    # Core template operations
│   │       ├── manifest.go     # Manifest parsing
│   │       ├── copy.go         # File copying utilities
│   │       └── source.go       # Template source interface
│   └── embedded/              # Embedded templates
│       └── templates.go       #go:embed directives
├── templates/                 # Template source files
│   ├── manifest.yaml          # Template metadata
│   ├── specledger/            # Spec Kit playbook
│   └── .claude/               # Claude Code integration
└── cmd/
    └── main.go                # CLI entry point
```

### Adding a New Template

1. Create template directory structure in `templates/`:
   ```bash
   mkdir -p templates/my-template/{scripts,templates}
   ```

2. Add template files to the directory

3. Update `templates/manifest.yaml`:
   ```yaml
   templates:
     - name: my-template
       description: "My custom template"
       framework: "none"
       version: "1.0.0"
       path: "my-template"
   ```

4. Rebuild SpecLedger:
   ```bash
   go build ./cmd/main.go
   ```

5. Test the new template:
   ```bash
   sl new --ci --project-name test --short-code ts --framework my-template
   ```

### Testing Template Copying

```bash
# Run template tests
go test ./pkg/cli/templates/...

# Run integration tests
go test ./tests/integration/templates_test.go
```

## Troubleshooting

### Templates Not Copied

**Problem**: Project created but no template files present

**Solutions**:
1. Check framework selection: `sl template list`
2. Verify embedded templates exist: `sl template list --json`
3. Check for errors during creation
4. Ensure `--framework` flag matches available template

### Existing Files Not Overwritten

**Problem**: Modified template files not updated

**Explanation**: By default, SpecLedger skips existing files to preserve user changes

**Solutions**:
1. Manually delete files and re-run `sl init`
2. Use `--force-templates` flag (future feature)
3. Manually update specific files

### Template Not Found

**Problem**: "template not found" error

**Solutions**:
1. Check available templates: `sl template list`
2. Verify manifest.yaml exists and is valid
4. Rebuild SpecLedger after adding templates

## Next Steps

After creating a project with templates:

1. **Customize Templates**: Edit template files in your project
2. **Create First Spec**: Use Claude Code commands in `.claude/commands/`
3. **Track Issues**: Use beads for task management
4. **Review Documentation**: Check `templates/AGENTS.md` for workflow guidance

## Examples

### Complete Workflow: New Spec Kit Project

```bash
# 1. Create new project with Spec Kit templates
sl new --ci \
  --project-name user-auth \
  --short-code ua \
  --framework speckit \
  --project-dir ~/projects

# 2. Navigate to project
cd ~/projects/user-auth

# 3. Review template files
ls -la .claude/ specledger/ .beads/

# 4. Create first spec
# (Use Claude Code: /speckit.specify "Implement user authentication")

# 5. Check issues
bd ls

# 6. Start development
```

### Custom Template Development

```bash
# 1. Create custom template structure
mkdir -p templates/company-playbook/{.claude,specledger}

# 2. Add company-specific files
echo "# Company Constitution" > templates/company-playbook/specledger/memory/constitution.md

# 3. Update manifest
cat >> templates/manifest.yaml << EOF
  - name: company
    description: "Company-specific playbook"
    framework: "speckit"
    version: "1.0.0"
    path: "company-playbook"
EOF

# 4. Rebuild and test
go build ./cmd/main.go
sl new --ci --project-name test --short-code ts --framework company
```
