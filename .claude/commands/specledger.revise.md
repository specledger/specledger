---
description: Fetch and address review comments from Supabase directly by file path
---

## User Input

```text
$ARGUMENTS
```

You **MUST** consider the user input before proceeding (if not empty).

## Purpose

This command fetches review comments for spec file(s) **directly from Supabase**. It queries by project/spec path to get all open review comments.

**When to use**:
- After pushing changes to GitHub
- When starting a new session and need to check for comments
- When team members have added comments to your specs
- To address feedback on your specifications

**Prerequisites**:
- Must be logged in via `sl login` (access token stored in `~/.specledger/credentials.json`)

## Input Options

### Option 1: No arguments (auto-detect from git remote)
```
/specledger.revise
```
→ Auto-detects repo owner/name from git remote and fetches all open review comments

### Option 2: Spec Folder Path
```
/specledger.revise "specledger/001-feature-name"
```
→ Filters comments for files in this folder

### Option 3: Explicit repo owner/name
```
/specledger.revise "owner/repo-name"
```
→ Fetches comments for this specific repository

## Execution Flow

### 1. Parse Arguments & Detect Repository

**Step 1a: Get repo info from git remote (if not provided)**:
```bash
# Get repo owner and name from git remote
git remote get-url origin
# Parse: https://github.com/OWNER/REPO.git or git@github.com:OWNER/REPO.git
```

**Step 1b: Parse arguments**:
- If `$ARGUMENTS` is empty → use auto-detected repo
- If `$ARGUMENTS` contains `/` without `specledger/` → treat as `owner/repo`
- If `$ARGUMENTS` starts with `specledger/` → treat as filter path

### 2. Query Supabase for Review Comments

Use the `scripts/review-comments.js` script (uses access token from `~/.specledger/credentials.json`):

```bash
# Query by project
node scripts/review-comments.js by-project <repo-owner> <repo-name>
```

**Expected output structure**:
```json
[
  {
    "change": {
      "id": "uuid",
      "head_branch": "change/spec-plan-tasks",
      "base_branch": "001-feature-name",
      "state": "open"
    },
    "comments": [
      {
        "id": "uuid",
        "content": "comment text",
        "file_path": "specledger/001-feature-name/spec.md",
        "selected_text": "text that was selected",
        "start_line": null,
        "line": null,
        "is_resolved": false,
        "author_id": "uuid",
        "created_at": "timestamp"
      }
    ]
  }
]
```

### 3. Display Comments Summary

Show all unresolved comments grouped by change/branch:

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📬 Review Comments for {repo_owner}/{repo_name}
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Change: {head_branch} → {base_branch} ({state})
Comments: {count}

┌─ {file_path}
│  Comment #{id (8 chars)}
│  Content: "{content}"
│  Selected: "{selected_text (truncated)}"
│  Resolved: {is_resolved}
└─
```

### 4. Process Each Comment Interactively

For each comment:

1. **Read the file** at `file_path`
2. **Find the selected_text** in the file (if available)
3. **Analyze the comment** content and context
4. **Generate 2-3 options** for addressing the feedback
5. **Use askUserQuestion** to get user preference
6. **Apply the edit** to the file
7. **Confirm** and move to next comment

**CRITICAL RULES:**
- MUST use askUserQuestion before making ANY edit
- If `selected_text` is provided, locate it in the file for context
- If `line` is provided, show that line in context
- Present clear, distinct options for each comment
- Apply edits incrementally, one comment at a time

### 5. Mark Comments as Resolved

After user confirms changes for each comment, mark it as resolved:

```bash
node scripts/review-comments.js resolve <comment-id>
```

### 6. Commit Changes (Optional)

After all comments are addressed:

```
📝 Ready to commit changes?

Options:
A) Yes, commit and push changes
B) No, I'll commit manually later
```

If user chooses A:
```bash
git add <modified-files>
git commit -m "feat: address review comments

Updated files: <list>
Comments resolved: <count>"
git push origin HEAD
```

### 7. Summary Report

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
✅ Review Session Complete
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📦 Repository: {repo_owner}/{repo_name}
🌿 Branch: {head_branch}
💬 Comments Addressed: {count}
📄 Files Updated: {count}

Files:
  ✓ {file_path} ({comment_count} comments)
  ...

✓ All comments marked as resolved
✓ Changes committed and pushed (if chosen)

Next Steps:
- View changes: git diff HEAD~1
- Continue with /specledger.implement
- Check new comments: /specledger.revise
```

## Error Handling

### Not logged in
**Error:** "Credentials file not found"
**Solution:** Run `sl login` to authenticate first

### Repository not found
**Error:** "Project not found: owner/repo"
**Solution:** Ensure repository is added to SpecLedger first

### No changes found
**Message:** "No open changes found"
**Meaning:** No active branches/PRs with review comments

### No comments found
**Message:** "No unresolved comments found"
**Meaning:** All comments have been resolved or no one has commented yet

### Script not found
**Error:** "Cannot find scripts/review-comments.js"
**Solution:** Ensure you're in the project root directory

## Notes

- Comments are fetched from `review_comments` table in Supabase
- Comments with `selected_text` show what text the reviewer highlighted
- Some comments are acknowledgments (like "good") and don't require file changes
- Works with any push method (git push, GitHub UI, etc.)
