# Data Model: Session-to-Knowledge Memory Pipeline

## Entities

### KnowledgeEntry

A structured piece of organizational knowledge extracted from session transcripts.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| id | string (UUID) | yes | Unique identifier |
| title | string | yes | Short descriptive title (max 200 chars) |
| description | string | yes | Full knowledge content |
| tags | string[] | yes | Category tags (e.g., "debugging", "architecture", "convention") |
| source_session_id | string | no | Session ID the entry was extracted from |
| source_branch | string | no | Feature branch where entry was discovered |
| scores | Score | yes | Three-axis scoring object |
| status | EntryStatus | yes | candidate, promoted, or archived |
| created_at | timestamp | yes | When the entry was first created |
| updated_at | timestamp | yes | When the entry was last modified |
| recurrence_count | int | yes | Number of sessions where this pattern appeared |

### Score

Scoring dimensions for a knowledge entry.

| Field | Type | Range | Description |
|-------|------|-------|-------------|
| recurrence | float | 0-10 | How often the pattern appears across sessions |
| impact | float | 0-10 | How much the pattern affects outcomes |
| specificity | float | 0-10 | How project-specific (10) vs. generic (0) |
| composite | float | 0-10 | Average of recurrence, impact, specificity |

### EntryStatus

| Value | Description |
|-------|-------------|
| candidate | Extracted but not yet promoted (below threshold or pending review) |
| promoted | Active in the knowledge base, injected into agent context |
| archived | Removed from active knowledge base, retained for history |

## Relationships

```
SessionTranscript (existing)
    ↓ extraction (AI-powered, L2 command)
KnowledgeEntry (candidate)
    ↓ scoring (composite ≥ 7.0)
KnowledgeEntry (promoted)
    ↓ rendering
knowledge.md (agent context)
```

## Validation Rules

- **title**: 1-200 characters, non-empty
- **description**: 1-5000 characters, non-empty
- **tags**: 1-10 tags, each 1-50 characters
- **scores**: Each axis 0.0-10.0
- **composite**: Calculated field, not user-settable
- **status transitions**: candidate → promoted, promoted → archived, candidate → archived, promoted → candidate (demote)

## Storage

### Local (JSONL)

File: `.specledger/memory/cache/entries.jsonl`

One JSON line per entry. Same pattern as `issues.jsonl`.

### Local (Generated Markdown)

File: `.specledger/memory/knowledge.md`

Auto-generated from promoted entries. Format:

```markdown
# Project Knowledge Base

> Auto-generated from promoted knowledge entries. Do not edit manually.
> Last updated: 2026-03-11T10:00:00Z

## Conventions

### [Entry Title]
[Entry description]
_Source: session abc123 on feature-branch | Score: 8.5_

## Patterns

### [Entry Title]
...
```

### Cloud (Supabase)

Table: `knowledge_entries`

Mirrors the KnowledgeEntry struct with additional `project_id` and `user_id` columns for multi-tenant access control.

## Promotion Threshold

- **Auto-promote**: composite score ≥ 7.0
- **Manual promote**: User can promote any entry regardless of score
- **Manual demote**: User can demote promoted entries back to candidate
