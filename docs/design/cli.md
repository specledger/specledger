# CLI Design Principles

Design guidelines for the `sl` CLI binary (Layer 1). The CLI is the agent's primary interface for data operations — it must be self-documenting, token-efficient, and navigable without external documentation.

> **Source**: Principles adapted from [Manus backend lead's CLI design post](https://www.reddit.com/r/LocalLLaMA/comments/1rrisqn/i_was_backend_lead_at_manus_after_building_agents/) and refined against SpecLedger's 4-layer model (see [docs/design/README.md](README.md)).

---

## Rules of Thumb

Quick reference for contributors. Every `sl` subcommand must satisfy these:

1. **Every subcommand must have `--help` with at least 2 usage examples**
2. **Every error must suggest a fix** — what failed, why, what to run instead
3. **Human output is compact** — truncated previews, counts, footer hints
4. **JSON output is complete** — full data, pipeable to `jq`, no truncation
5. **Errors to stderr, data to stdout** — enables clean piping
6. **Positional args for simple cases** — reserve flags for optional/complex params
7. **`--spec` flag for spec override** — auto-detect via ContextDetector by default
8. **Classify against a pattern before merging** — see [Pattern Classification](#pattern-classification)

---

## Principle 1: Progressive Discovery

A well-designed CLI doesn't require reading documentation — `--help` tells you everything. The agent discovers on-demand, each level providing just enough for the next step.

The agent doesn't need to load all documentation at once, but discovers details on-demand as it goes deeper.

**Level 0: Tool description → command list injection**
The agent knows what `sl` subcommands exist from Cobra's auto-generated help. No need to preload all docs.

**Level 1: `sl <command>` (no args) → subcommand usage**
When the agent is interested in a command, it just calls it. No arguments? The command returns its own usage:

When the agent runs a command without args, it gets its subcommand list:
```
$ sl comment
Manage review comments

Subcommands:
  list     List review comments (compact or JSON format)
  show     Show full comment details with thread replies
  reply    Reply to a comment thread
  resolve  Mark comments as resolved
```

**Level 2: `sl <command> <subcommand>` (missing args) → specific params**

The agent decides to use `sl comment resolve` but isn't sure about the format? It drills down:

```
$ sl comment resolve
Error: requires at least 1 arg(s), only received 0

Usage:
  sl comment resolve <comment-id> [--reason "text"] [--json]

Examples:
  sl comment resolve abc123 --reason "Fixed in PR #42"
  sl comment resolve abc123 def456 --reason "Batch resolved"
```
Progressive disclosure: **overview (injected) → usage (explored) → parameters (drilled down).** The agent discovers on-demand, each level providing just enough information for the next step.

This is fundamentally different from stuffing 3,000 words of tool documentation into the system prompt. Most of that information is irrelevant most of the time — pure context waste. Progressive help lets the agent decide when it needs more.

This also imposes a requirement on command design: **every command and subcommand must have complete help output.** It's not just for humans — it's for the agent. A good help message means one-shot success. A missing one means a blind guess.

**Rule**: Every subcommand MUST have a `Long` description with at least 2 `Examples` in its Cobra definition. The agent should be able to construct correct commands from `--help` alone.

---

## Principle 2: Error Messages as Navigation

Agents will make mistakes. The key isn't preventing errors — it's **making every error point to the right direction.**

Agents can't Google. Every error must contain both "what went wrong" and "what to do instead."

```
# Bad: raw API error with no context
ResolveComment: API error (403): {"code":"42501","message":"new row violates row-level security"}

# Good: actionable guidance that preserves the raw error
ResolveComment failed (403): row-level security violation — missing required fields.
→ Run 'sl comment resolve --verbose ...' to see the full API response
→ This usually means auth credentials are stale. Try 'sl auth login' first.
```

```
# Bad: guidance that masks the real error
ResolveComment failed: auth token expired.
→ Run 'sl auth login' to re-authenticate, then retry.
# ↑ If the real error was NOT auth, the user is now debugging the wrong thing

# Good: preserve the actual error, suggest based on common causes
ResolveComment failed (403): row-level security violation.
→ Common cause: missing auth or stale token. Run 'sl auth login' to refresh.
→ If auth is valid, run with --verbose to see the full API response.
```

```
# Bad: missing context
failed to get project: project not found

# Good: next step
Project not found for 'specledger/specledger'.
→ Check repo remote with 'git remote -v'
→ Or specify manually: sl comment list --spec 601-cli-skills
```

**Rule**: Every CLI error MUST include:
1. What failed (operation name + HTTP status if API)
2. Why it failed — the **actual** error, not a guess. Never swallow the raw error.
3. Suggested fix based on common causes
4. Escape hatch to see more detail (e.g., `--verbose` flag)

> **stderr is the information agents need most, precisely when commands fail. Never drop it.**

All error/warning output goes to stderr. All data output goes to stdout. This enables clean piping: `sl comment list --json 2>/dev/null | jq ...`

Never replace the actual error with a guess about the cause. The guidance should *supplement* the raw error, not *replace* it. If an agent sees "auth token expired" but the real problem was a missing `change_id` field, it will waste cycles re-authenticating instead of fixing the actual issue.

### Desire Paths

Errors are also a signal about how agents *want* to use the CLI. The concept of [desire paths](https://en.wikipedia.org/wiki/Desire_path) — trails worn by users taking shortcuts — applies to CLI design:

- If agents repeatedly try `sl comment "new comment"` instead of `sl comment add "new comment"`, that failed path is a desire path. Consider paving it with an alias or shorthand.
- If an error pattern repeats across multiple agent sessions, consider whether the failed path should become a real command instead of just redirecting.

> **Future**: Systematic desire path detection requires logging/monitoring agent CLI usage patterns. Consider adding opt-in usage telemetry to identify common failure patterns.

---

## Principle 3: Two-Level Output Design

### Human Output (default): Compact and Budget-Conscious

Human output is designed for quick scanning — truncated previews, counts instead of nested data, footer hints for next steps.

```
$ sl comment list
df77879f | spec.md:42 | cái này có cần thiết không | Son Vo | 2 replies
f846962e | spec.md:15 | non existent command | so0k | 1 reply

2 comment(s) across 1 artifact(s)
→ Use 'sl comment show <id>' for full details and thread context
```

**Truncation rules** (human output only):
- Content: first 80 chars, append `...` if truncated
- Newlines replaced with spaces
- Reply count shown as `N replies`, not full thread

**Footer hints**: Every compact output MUST end with a hint line suggesting the drill-down command. This is the agent's navigation cue.

### JSON Output (`--json`): Complete and Pipeable

JSON output includes full, untruncated data. The consumer (agent or `jq` pipe) decides what to extract. No truncation, no previews.

```bash
# Agent workflow: list overview, then selectively drill down
sl comment list --json | jq '.[].id'                    # scan IDs
sl comment show abc123 --json | jq '.replies[].content'  # drill into one
```

**Rule**: JSON output is the "execution layer" — complete, structured, pipeable. Human output is the "presentation layer" — budget-conscious, hinted. Token efficiency is achieved by the *workflow pattern* (list → show), not by truncating JSON.

### Exit Codes

Standard Unix convention. No duration metadata (our CLI runs locally against Supabase API; duration varies by network, not command cost).

| Code | Meaning |
|------|---------|
| 0 | Success (including empty results) |
| 1 | Error (auth, network, invalid args) |

### Binary Output

Not applicable — the `sl` CLI produces structured text/JSON only. If a future command needs to reference binary artifacts (images, PDFs), it MUST output the file path, not the binary content.

---

## Pattern Classification

Every `sl` subcommand MUST identify which pattern(s) it follows. This classification drives design constraints and review expectations. See the [checklist template](../../.specledger/templates/checklist-template.md) for the pattern gate checklist used during spec reviews.

| Pattern | Purpose | Examples | Constraints |
|---------|---------|----------|-------------|
| **Data CRUD** | Deterministic operations on entities | `sl issue`, `sl deps`, `sl comment` | No AI reasoning. Returns structured data. `--spec` flag for override, ContextDetector for auto-detect. |
| **Launcher** | Pre-flight + context gathering + spawn agent | `sl revise`, `sl mockup`, `sl init` | CLI does NOT interpret agent results. Agent owns commits/resolution. CLI only does pre-session setup. |
| **Hook Trigger** | Invisible automation on agent shell events | `sl session capture` | MUST be non-blocking. MUST handle failures gracefully (cache + retry, never block UX). See [docs/design/hooks.md](hooks.md). |
| **Environment** | System health, auth, configuration | `sl doctor`, `sl auth`, `sl version` | MUST be idempotent. MUST work without a project context. |
| **Template Management** | Install/update/remove agent shell templates | `sl doctor --template` | MUST detect stale/deprecated templates. MUST prompt before destructive changes. Owns `specledger.` prefix. |

### Failure Mode Constraints

Each pattern has a distinct failure mode expectation:

| Pattern | Failure Mode |
|---------|-------------|
| **Data CRUD** | MUST fail with clear error + suggested fix. Agent needs inline feedback. |
| **Launcher** | MUST validate prerequisites before spawning agent. Fail early with guidance. |
| **Hook Trigger** | MUST handle failures gracefully — cache locally and retry. Never block UX. A broken hook breaks the entire agent session. |
| **Environment** | MUST be idempotent and safe to retry. |
| **Template Management** | MUST prompt before destructive changes. `--dry-run` for preview. |

### Offline Behavior

- **Local-first commands** (`sl issue`, `sl deps`): Work on local JSONL files. Fully offline.
- **API-backed commands** (`sl comment`): Require Supabase connectivity. MUST fail with clear error when offline.
- **Hook trigger** (`sl session capture`): MUST cache locally when offline/unauthenticated and retry when connectivity is restored. A hook failure should never surface as an error to the user.

---

## Anti-Patterns

Common implementation deviations found in practice. Avoid these.

### AP-01: Full content in compact list output
```
# Wrong: sends full content in human list output
fmt.Printf("%s | %s\n", c.ID, c.Content)  // 500 chars of content

# Right: truncate in human output
content := truncate(c.Content, 80)
fmt.Printf("%s | %s\n", c.ID, content)
```

### AP-02: Silent resolution without audit trail
```
# Wrong: resolve with no reason
sl comment resolve abc123

# Right: reason required, CLI posts reply automatically
sl comment resolve abc123 --reason "Fixed in PR #42"
```

### AP-03: Swallowing raw errors behind guidance
```
# Wrong: raw API dump with no guidance
return fmt.Errorf("API error (%d): %s", resp.StatusCode, body)

# Also wrong: guidance that replaces the actual error
return fmt.Errorf("ResolveComment failed: auth token expired.\n→ Run 'sl auth login'")
// ↑ The real error might NOT be auth — now the agent debugs the wrong thing

# Right: preserve raw error + add guidance based on common causes
return fmt.Errorf("ResolveComment failed (%d): %s\n→ Common cause: stale auth. Run 'sl auth login' to refresh.\n→ Run with --verbose for full API response", resp.StatusCode, parseError(body))
```

### AP-04: Duplicating ContextDetector per package
```
# Wrong: each package re-implements branch detection
branch, _ := cligit.GetCurrentBranch(cwd)  // manual in sl comment
specKey = currentBranch                      // raw branch name

# Right: shared ContextDetector with fallback chain
detector := context.NewContextDetector(cwd)
specKey, err := detector.DetectSpec()        // regex → alias → git heuristic
```

---

## Architecture: Execution vs Presentation

Inside the CLI (L1), there are two conceptual layers:

```
┌─────────────────────────────────────────────┐
│  Presentation: LLM/Human-facing output      │  ← --json (full) vs human (compact + hints)
│  Truncation | Footer hints | stderr/stdout  │
├─────────────────────────────────────────────┤
│  Execution: Go business logic               │  ← Cobra routing, API calls, data transforms
│  Pattern constraints | ContextDetector      │
└─────────────────────────────────────────────┘
```

The execution layer handles command routing, API calls, and data operations. The presentation layer formats output for the consumer (human or agent). These are not separate packages — they're a design concern within each command's `runXxx()` function.

**Key rule**: Execution logic must not depend on output format. The same data path should serve both `--json` and human output.

---

## References

- [4-Layer Model Overview](README.md)
- [Hooks Design](hooks.md)
- [Commands Design](commands.md)
- [Skills Design](skills.md)
- [Tech Debt Scratchpad](tech-debts.md)
- [External: Manus CLI design post](https://www.reddit.com/r/LocalLLaMA/comments/1rrisqn/i_was_backend_lead_at_manus_after_building_agents/)
- [External: Rewrite your CLI for AI agents](https://justin.poehnelt.com/posts/rewrite-your-cli-for-ai-agents/)
- [External: Desire Paths (Wikipedia)](https://en.wikipedia.org/wiki/Desire_path)
