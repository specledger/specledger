# CLI Contract: Issue Commands

**Feature**: 597-issue-create-fields
**Date**: 2026-02-24

## sl issue create

### Synopsis

```bash
sl issue create --title <string> [flags]
```

### Required Flags

| Flag | Type | Description |
|------|------|-------------|
| `--title` | string | Issue title (required) |

### Optional Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--description` | string | "" | Issue description |
| `--type` | string | "task" | Issue type: epic, feature, task, bug |
| `-p, --priority` | int | 2 | Priority: 0-5 (0=highest) |
| `--labels` | string | "" | Comma-separated labels |
| `--spec` | string | "" | Override spec context |
| `--force` | bool | false | Skip duplicate detection |
| `--acceptance-criteria` | string | "" | Acceptance criteria text |
| `--dod` | []string | [] | Definition of Done items (repeatable) |
| `--design` | string | "" | Design notes/approach |
| `--notes` | string | "" | Implementation notes |
| `--parent` | string | "" | Parent issue ID |

### Output

**Success (non-JSON)**:
```
✓ Created issue SL-xxxxxx
  Title: <title>
  Type: <type>
  Priority: <priority>
  Spec: <spec-context>

View: sl issue show SL-xxxxxx
```

**Success (--json)**:
```json
{
  "id": "SL-xxxxxx",
  "title": "...",
  "acceptance_criteria": "...",
  "definition_of_done": {
    "items": [
      {"item": "Item 1", "checked": false},
      {"item": "Item 2", "checked": false}
    ]
  },
  "design": "...",
  "notes": "...",
  "parentId": "SL-yyyyyy"
}
```

### Error Cases

| Condition | Exit Code | Error Message |
|-----------|-----------|---------------|
| Missing --title | 1 | `title is required` |
| Invalid --type | 1 | `invalid issue type: <value>` |
| Invalid --priority | 1 | `invalid priority: must be 0-5` |
| Invalid --parent | 1 | `parent issue not found: <id>` |
| Self as parent | 1 | `cannot set self as parent` |

---

## sl issue update

### Synopsis

```bash
sl issue update <issue-id> [flags]
```

### Arguments

| Argument | Type | Description |
|----------|------|-------------|
| issue-id | string | Issue ID (format: SL-xxxxxx) |

### Optional Flags

| Flag | Type | Description |
|------|------|-------------|
| `--title` | string | Update title |
| `--description` | string | Update description |
| `--status` | string | Update status: open, in_progress, closed |
| `-p, --priority` | int | Update priority |
| `--assignee` | string | Update assignee |
| `--notes` | string | Update notes |
| `--design` | string | Update design notes |
| `--acceptance-criteria` | string | Update acceptance criteria |
| `--add-label` | string | Add a label |
| `--remove-label` | string | Remove a label |
| `--dod` | []string | Replace Definition of Done items (repeatable) |
| `--check-dod` | string | Mark DoD item as checked (exact match) |
| `--uncheck-dod` | string | Mark DoD item as unchecked (exact match) |
| `--parent` | string | Set parent issue ID (empty string to clear) |

### Output

**Success**:
```
✓ Updated issue SL-xxxxxx
```

### Error Cases

| Condition | Exit Code | Error Message |
|-----------|-----------|---------------|
| Invalid issue-id format | 1 | `invalid issue ID: <id>` |
| Issue not found | 1 | `failed to get issue: ...` |
| DoD item not found | 1 | `DoD item not found: '<text>'` |
| Already has parent | 1 | `issue already has a parent, remove existing parent first` |
| Self as parent | 1 | `cannot set self as parent` |
| Circular parent | 1 | `circular parent-child relationship detected` |
| Parent not found | 1 | `parent issue not found: <id>` |

---

## sl issue show

### Synopsis

```bash
sl issue show <issue-id> [flags]
```

### Arguments

| Argument | Type | Description |
|----------|------|-------------|
| issue-id | string | Issue ID (format: SL-xxxxxx) |

### Flags

| Flag | Type | Description |
|------|------|-------------|
| `--json` | bool | Output as JSON |
| `--tree` | bool | Show dependency tree with children |

### Output Format (non-JSON)

```
Issue: SL-xxxxxx
  Title: <title>
  Type: <type>
  Status: <status>
  Priority: <priority> (<priority-label>)
  Spec: <spec-context>
  Parent: SL-yyyyyy (if set)

Description:
  <description>

Acceptance Criteria:
  <acceptance_criteria>

Design:
  <design>

Definition of Done:
  [x] <checked item> (verified: <timestamp>)
  [ ] <unchecked item>

Notes:
  <notes>

Labels:
  - <label1>
  - <label2>

Created: <timestamp>
Updated: <timestamp>
Closed: <timestamp>  (if closed)
```

**Field Display Rules**:
- Empty fields are omitted from display
- Acceptance Criteria, Design, Notes shown only if populated
- Definition of Done shown only if items exist
- Labels shown only if any exist
- Parent shown only if set

### Tree Output Format (--tree)

```
Issue: SL-xxxxx (Epic)
├── Issue: SL-yyyyy (Feature)
│   ├── Issue: SL-zzzz1 (Task 1) [P1]
│   └── Issue: SL-zzzz2 (Task 2) [P2]
└── Issue: SL-wwww (Feature 2) [P2]
```

**Children Ordering**:
1. Priority (descending - higher priority first)
2. Creation order / ID (ascending)

---

## Behavior Contracts

### DoD Text Matching

- **Pattern**: Exact string match
- **Case sensitivity**: Yes (case-sensitive)
- **Whitespace**: No normalization (trailing/leading spaces must match)
- **Error format**: `DoD item not found: '<provided-text>'`

### DoD Replacement

- `--dod` on update **replaces** entire Definition of Done
- Previous items are **not preserved**
- Items created as unchecked (checked=false, verified_at=null)

### DoD Check/Uncheck

- `--check-dod` sets: checked=true, verified_at=now()
- `--uncheck-dod` sets: checked=false, verified_at=null
- Both require exact text match
- Both return error if item not found

### Parent-Child Relationships

- `--parent` on create sets the parent immediately
- `--parent` on update:
  - With valid ID: Sets parent (fails if already has parent)
  - With empty string: Clears parent (idempotent)
- **Single parent constraint**: An issue can only have ONE parent
- **Self-parent prevention**: Cannot set self as parent
- **Cycle prevention**: Cannot create circular parent chains
- **Hierarchy depth**: Unlimited (no maximum depth)
- **Children query**: Computed at read time, not stored

### Error Messages

| Scenario | Error Message |
|----------|---------------|
| Set second parent | `issue already has a parent, remove existing parent first` |
| Set self as parent | `cannot set self as parent` |
| Create circular chain | `circular parent-child relationship detected` |
| Parent doesn't exist | `parent issue not found: <id>` |
