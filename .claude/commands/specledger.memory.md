---
description: Extract structured knowledge from session transcripts using AI-powered analysis
---

## User Input

```text
$ARGUMENTS
```

You **MUST** consider the user input before proceeding (if not empty).

## Purpose

AI-powered knowledge extraction from session transcripts. Analyzes conversations to identify patterns, decisions, debugging insights, and conventions, then produces structured knowledge entries stored in the local knowledge cache.

**When to use**: After a productive session to capture learned knowledge for future sessions.

## Modes

The command supports four modes. If no mode specified, default to `summarize`.

### Mode: summarize (default)

Extract key decisions, patterns, and insights from the current or specified session.

1. Read the current session transcript (or use `sl session get <id>` if a session ID is provided)
2. Analyze assistant messages for:
   - Architectural decisions and their rationale
   - Debugging insights and root cause analyses
   - Code patterns and conventions established
   - Mistakes made and corrections applied
3. For each significant finding, create a KnowledgeEntry:
   - **title**: Concise description (max 200 chars)
   - **description**: Full context and details
   - **tags**: Category tags (e.g., "debugging", "architecture", "convention", "pattern")
   - **scores**: Assess Impact (0-10) and Specificity (0-10). Set Recurrence to 1.0 for new entries.
4. Store entries in `.specledger/memory/cache/entries.jsonl` using `sl memory` or by writing JSONL directly
5. If composite score >= 7.0, auto-promote the entry
6. Report what was extracted

### Mode: tag

Assign or update category tags on existing knowledge entries.

1. List current entries with `sl memory show`
2. Analyze entry titles and descriptions
3. Suggest appropriate tags based on content
4. Update entries with new tags

### Mode: patterns

Identify recurring patterns across multiple sessions.

1. Read multiple session transcripts
2. Cross-reference findings to detect recurring themes
3. Increase Recurrence scores for patterns found in multiple sessions
4. Merge duplicate entries

### Mode: synthesize

Merge related entries into consolidated knowledge.

1. List all entries and group by similarity
2. Propose merges for closely related entries
3. Create consolidated entries with combined descriptions
4. Archive the source entries

## Output Format

For each extracted entry, display:

```
[NEW] Title: <title>
      Tags: <tag1>, <tag2>
      Score: R:<recurrence> I:<impact> S:<specificity> = <composite>
      Status: <candidate|promoted>
```

If no significant knowledge found, report:
```
No significant knowledge patterns found in this session.
```

## Storage

- Entries: `.specledger/memory/cache/entries.jsonl`
- Promoted: `.specledger/memory/knowledge.md` (auto-generated)

## Guidelines

- Only extract genuinely useful knowledge — avoid low-quality filler entries
- Prefer fewer, high-quality entries over many weak ones
- Score Impact based on how much the insight would help future sessions
- Score Specificity based on whether this is project-specific (high) or generic (low)
- Set Recurrence to 1.0 for new entries; it increases when patterns recur
