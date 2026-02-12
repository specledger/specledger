---
description: Display all available SpecLedger commands with descriptions and workflow guidance
---

## Purpose

Quick reference for all available SpecLedger commands, organized by workflow stage. Use this when you need to discover what commands are available or understand the recommended workflow.

## Output

### Core Workflow (specify → plan → tasks → implement)

| Command | Description |
|---------|-------------|
| `/specledger.specify` | Create feature specification from natural language description |
| `/specledger.plan` | Generate implementation plan from spec with architecture decisions |
| `/specledger.tasks` | Create actionable, dependency-ordered tasks from plan |
| `/specledger.implement` | Execute tasks in order following the task plan |

### Analysis & Validation

| Command | Description |
|---------|-------------|
| `/specledger.analyze` | Cross-artifact consistency check across spec, plan, and tasks |
| `/specledger.audit` | Complete codebase audit - tech stack detection, module discovery, and dependency graphs |
| `/specledger.clarify` | Identify and resolve spec ambiguities via targeted questions |
| `/specledger.checklist` | Generate custom validation checklist for the feature |

### Setup & Configuration

| Command | Description |
|---------|-------------|
| `/specledger.constitution` | Define project principles and coding standards |
| `/specledger.adopt` | Create spec from existing branch or audit output |

### Collaboration

| Command | Description |
|---------|-------------|
| `/specledger.revise` | Fetch and address review comments from Supabase |

## Workflow Guide

**New Feature Development:**
```
/specledger.specify → /specledger.plan → /specledger.tasks → /specledger.implement
```

**Existing Codebase Analysis:**
```
/specledger.audit → /specledger.adopt
```

**Spec Quality Improvement:**
```
/specledger.clarify → /specledger.checklist → /specledger.analyze
```

## Quick Tips

- Start with `/specledger.specify` for new features
- Run `/specledger.analyze` after task generation to verify consistency
