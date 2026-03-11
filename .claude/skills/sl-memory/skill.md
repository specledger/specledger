# sl-memory Skill

## Overview

Passive knowledge injection skill that loads the project's accumulated knowledge base into agent context at session start. This knowledge includes patterns, conventions, debugging insights, and architectural decisions extracted from previous AI sessions.

**Layer**: L3 (Skill) - Passive context injection
**Use when**: Starting a new session on a project that has accumulated knowledge entries

## When to Load

Load this skill when:
- Starting a new session on a feature branch
- Need context about project conventions and patterns
- Want to avoid repeating mistakes from previous sessions
- Working on code that has established patterns

**Don't load when**:
- No knowledge base exists for the project
- Working on a brand new project with no history

## Knowledge Base

The knowledge base is stored at `.specledger/memory/knowledge.md` and is auto-generated from promoted knowledge entries. It contains:

- **Conventions**: Coding standards, naming patterns, architectural rules
- **Patterns**: Recurring solutions, design patterns specific to this project
- **Debugging**: Known issues, common pitfalls, troubleshooting guides
- **Decisions**: Architectural decisions and their rationale

## CLI Commands

| Action | Command |
|--------|---------|
| View all entries | `sl memory show` |
| Promote entry | `sl memory promote <id>` |
| Demote entry | `sl memory demote <id>` |
| Delete entry | `sl memory delete <id>` |
| Sync to cloud | `sl memory sync` |
| Pull from cloud | `sl memory pull` |

## Extraction

Use the `/specledger.memory` command to extract knowledge from session transcripts:

```
/specledger.memory summarize   # Extract key decisions and patterns
/specledger.memory tag         # Assign category tags
/specledger.memory patterns    # Identify recurring patterns
/specledger.memory synthesize  # Merge related entries
```

## Scoring System

Entries are scored on three axes (0-10 each):
- **Recurrence**: How often the pattern appears across sessions
- **Impact**: How much the pattern affects outcomes
- **Specificity**: How project-specific vs. generic

Composite score = average of three axes. Entries with composite >= 7.0 are auto-promoted.

## Knowledge File Location

```
.specledger/memory/
├── cache/
│   └── entries.jsonl    # All knowledge entries (JSONL)
└── knowledge.md         # Promoted entries (auto-generated markdown)
```
