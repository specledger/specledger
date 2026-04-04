---
name: creating-pr
description: Orchestrate the full PR creation workflow — commit, branch, push, create PR via gh CLI, and optionally poll for Qodo code review results. Use this skill whenever the user asks to "create a PR", "open a PR", "submit a PR", "commit and push and create a PR", or any variation of making a pull request. Also trigger when the user says "let's PR this" or "ship it" in a context where there are uncommitted or unpushed changes.
---

# Creating a PR

This skill walks through the complete pull request workflow for this repository, from staging changes through to optionally waiting for automated code review results from Qodo.

The reason this skill exists is that PR creation involves several coordinated steps where getting the details right matters — conventional commit format, structured PR bodies, and the opportunity to catch review feedback before context-switching away. Following this workflow means the PR is ready for human review the moment it lands.

## Workflow

### Step 1: Pre-flight

Run these in parallel to understand what you're working with:

```bash
git status                  # untracked + modified files
git diff                    # unstaged changes
git diff --cached           # staged changes
git log --oneline -5        # recent commit style
```

Before proceeding, verify there are actual changes to commit. If the working tree is clean, tell the user and stop.

### Step 2: Commit

Stage the relevant files — prefer naming specific files over `git add -A` to avoid accidentally including sensitive files (`.env`, credentials, local settings).

Write a conventional commit message following the project's release flow (see `docs/guides/release-flow.md`). The format is `type(scope): description` or `type: description`.

**Version-bumping types:** `feat` (minor), `fix` (patch), `perf` (patch), `deps` (patch), `revert` (patch)
**Non-bumping types:** `chore`, `docs`, `style`, `refactor`, `test`, `build`, `ci`

End every commit message with:
```
Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>
```

Use a HEREDOC for the commit message to preserve formatting:
```bash
git commit -m "$(cat <<'EOF'
type(scope): short description

Longer explanation of why, not what.

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>
EOF
)"
```

### Step 3: Branch and push

If currently on `main`, create a feature branch first — never commit directly to main. Branch naming convention: `type/short-description` (e.g., `fix/yaml-parsing`, `ci/harden-workflows`).

```bash
git checkout -b <branch-name>   # only if on main
git push -u origin <branch-name>
```

### Step 4: Create the PR

Use `gh pr create` with a structured body. PR titles follow the same conventional commit format (they become the squash-merge commit message on main).

```bash
gh pr create --title "type(scope): description" --body "$(cat <<'EOF'
## Summary
<1-3 bullet points explaining what and why>

## Test plan
[Describe how changes were verified]

🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
)"
```

Capture the returned PR URL — `gh pr create` prints it to stdout.

### Step 5: Poll for Qodo review (optional)

After the PR is created, ask the user:

> "PR created at {url}. Want me to wait for the Qodo code review?"

If yes, run the review poller in the background. This frees up the conversation — the agent gets notified when the review lands.

```bash
node scripts/review-poll.js <pr-url> --interval 15 --timeout 300
```

Run this with `run_in_background: true` on the Bash tool. When the background task completes, you'll be notified automatically. At that point:

1. Read the output (the Qodo review body in markdown)
2. Parse out any findings — look for sections marked `Action required` or `Review recommended`
3. Summarize the findings for the user, highlighting any bugs or actionable items
4. If there are issues to fix, offer to address them

If the user says no to polling, you're done — just return the PR URL.

## Script reference

**`scripts/review-poll.js`** — Polls a GitHub PR for the Qodo bot's code review comment.

```
node scripts/review-poll.js <pr-url> [--interval 15] [--timeout 300]
```

- Accepts `https://github.com/owner/repo/pull/123` or `owner/repo#123`
- Shells out to `gh api` (no npm dependencies)
- Detects the Qodo bot comment by author login (contains "qodo")
- Distinguishes placeholder ("Looking for bugs? Check back in a few minutes") from the final review (contains `Bugs`, `Rule violations`, `Action required`, etc.)
- Exit 0 = review found (body on stdout), exit 1 = timeout, exit 2 = bad args

## Decision patterns

### When to create a new branch

- On `main` → always create a branch
- On an existing feature branch → stay on it (the user likely has in-progress work)
- Branch already tracks remote and is up to date → just push new commits

### Commit granularity

- One logical change = one commit. Don't bundle unrelated fixes.
- If the user explicitly asks for a single commit covering multiple changes, that's fine.
- If a pre-commit hook fails, fix the issue and create a NEW commit (don't amend — the failed commit didn't happen).

### PR body depth

- Small changes (typos, config tweaks) → brief summary, minimal test plan
- Feature work → explain the why, list affected areas, thorough test plan
- Security/CI changes → explain the threat model or reasoning, link to relevant incidents or docs
