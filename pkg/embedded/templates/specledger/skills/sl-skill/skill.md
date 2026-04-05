---
name: sl-skill
description: Manage agent skills from the skills.sh registry — search, install, audit, and remove community-built skills
---
# sl Skill Management

**When to Load**: Triggered when tasks involve discovering, installing, removing, or auditing agent skills from the skills.sh registry, or when users ask about extending their agent's capabilities with community skills.

## Overview

`sl skill` manages agent skills from the Vercel skills.sh registry. Skills are SKILL.md files installed to agent-specific directories (e.g., `.claude/skills/`) and tracked in `skills-lock.json` for reproducibility.

## Subcommands

| Command | Purpose | Output Mode |
|---------|---------|-------------|
| `sl skill search <query>` | Search skills.sh registry by keyword (supports `--limit N`, default 10) | Compact table + footer hint |
| `sl skill add <source>` | Install skill(s) from a repository | Audit table + confirmation + progress |
| `sl skill info <source>` | Show skill metadata and security audit | Detail view with audit |
| `sl skill list` | List installed skills from lock file | Compact list + footer hint |
| `sl skill remove <skill-name>` | Remove an installed skill | Confirmation |
| `sl skill audit [skill-name]` | Run security audit on installed skills | 3-partner audit table |

## Source Formats

Skills are identified by source format:

| Format | Example | Description |
|--------|---------|-------------|
| `owner/repo` | `vercel-labs/agent-skills` | All skills from repo |
| `owner/repo@skill` | `vercel-labs/agent-skills@creating-pr` | Specific skill |
| Full HTTPS URL | `https://github.com/org/repo.git` | Git clone fallback |
| SSH URL | `git@github.com:org/repo.git` | Git clone fallback |

## Common Workflows

### Discover and Install a Skill

```bash
# Search for skills
sl skill search "commit"

# Search with a result limit
sl skill search "deploy" --limit 5

# Install a specific skill
sl skill add vercel-labs/agent-skills@creating-pr

# Install all skills from a repo (non-interactive)
sl skill add vercel-labs/agent-skills -y
```

### Manage Installed Skills

```bash
# List what's installed
sl skill list

# Check security
sl skill audit

# Remove a skill
sl skill remove creating-pr
```

### JSON Output for Scripting

All commands support `--json` for machine-readable output:

```bash
# Search results as JSON
sl skill search "deploy" --json

# Installed skills as JSON
sl skill list --json

# Audit results as JSON
sl skill audit --json
```

## Security Audit Interpretation

The `audit` and `add` commands show security assessments from three partners:

| Partner | What It Checks | Risk Levels |
|---------|---------------|-------------|
| **Gen** (ATH) | General threat intelligence | safe, low, medium, high, critical |
| **Socket** | Supply chain vulnerability alerts | Alert count (0 = clean) |
| **Snyk** | Known vulnerability scan | safe, low, medium, high, critical |

**Risk guidance**:
- **safe/low**: Generally safe to use
- **medium**: Review the audit details before using in production
- **high/critical**: Investigate before using — check the skill source for known issues

## Lock File

Skills are tracked in `skills-lock.json` (Vercel-compatible format):

```json
{
  "version": 1,
  "skills": {
    "skill-name": {
      "source": "owner/repo",
      "sourceType": "github",
      "computedHash": "sha256-hex"
    }
  }
}
```

This file should be committed to version control for reproducible skill installations across team members.

## Decision Criteria

### When to Use sl skill

- User asks to install, find, or manage agent skills
- User wants to extend their agent with community skills
- User asks about skills.sh or the skills registry
- User wants to audit installed skills for security

### When NOT to Use sl skill

- User wants to create a new skill from scratch (use skill-creator instead)
- User asks about `sl deps` (that's for spec dependencies, not skills)
- User wants to modify an installed skill directly (edit the SKILL.md file instead)
