# Hooks Design Principles (Layer 0)

Hooks are invisible, event-driven automations triggered by agent shell events. They run without user interaction and must never disrupt the developer experience.

> **Decision history**: Three approaches were evaluated for session capture: (1) Claude Code PostToolUse hook — selected, (2) AI-triggered command (rejected: relies on AI compliance, session ID not accessible to AI), (3) Wrapper binary `sl claude` (rejected: requires workflow change, only captures at session end). See [spike-hooks.md](../../specledger/010-checkpoint-session-capture/spike-hooks.md) for the full evaluation.

---

## Rules of Thumb

1. **Never block the agent session** — hooks are fire-and-forget
2. **Cache on failure, retry later** — offline/auth failures must not surface as errors
3. **Local file first, remote second** — write locally (guaranteed), then sync (best-effort)
4. **No user workflow changes** — hooks are transparent; users don't invoke them
5. **One trigger, one purpose** — each hook configuration maps to a single operation

---

## Core Principle: Silent Resilience

A broken hook breaks the entire agent session UX. Unlike CLI commands (L1) that should fail loudly with guidance, hooks must fail silently and recover gracefully.

```
# Wrong: hook surfaces error to agent
[stderr] session capture failed: auth token expired

# Right: hook caches locally and retries
→ Writes to ~/.specledger/capture-queue.jsonl
→ Next successful auth triggers queue flush
→ User never sees the failure
```

**Error logging** (two-tier):
1. **Local file** (`~/.specledger/capture-errors.log`): JSONL format, written first (guaranteed). For immediate user troubleshooting. Includes: timestamp, user_id, session_id, error_message, branch, commit_hash, retry_count.
2. **Remote** (Sentry): Best-effort, non-blocking. For team-level visibility, aggregation, alerting. Sentry failure never blocks the hook.

**Why Sentry over Supabase for error logging**: Logging errors to Supabase adds load to the same system that handles app data. Sentry is purpose-built for error aggregation, deduplication, and alerting. A Supabase `session_capture_errors` table was rejected — it would compete with production data for quota and require custom dashboards for what Sentry provides out of the box.

---

## Available Claude Code Hooks

Reference: [Claude Code Hooks Guide](https://code.claude.com/docs/en/hooks-guide)

| Hook | Trigger | Transcript Access | Current Use |
|------|---------|-------------------|-------------|
| `PostToolUse` | After tool execution completes | Yes | **Session capture** (Bash matcher for `git commit`) |
| `SessionStart` | New session begins | Yes | — |
| `PreToolUse` | Before tool execution | Yes | — |
| `Stop` | Session manually stopped | Yes | — |
| `TaskCompleted` | Task finished | Yes | Future: beads task capture |
| `PreCompact` | Before context compression | Yes | Future: reasoning preservation |
| `SessionEnd` | Session ends | Yes | Future: final checkpoint |

### Hook Input Format

Hooks receive JSON on stdin with session context:

```json
{
  "session_id": "abc-123",
  "transcript_path": "/home/user/.claude/projects/project/session.jsonl",
  "cwd": "/home/user/project",
  "hook_event_name": "PostToolUse",
  "tool_name": "Bash",
  "tool_input": {"command": "git commit -m 'message'"},
  "tool_response": {"stdout": "...", "stderr": "", "interrupted": false},
  "tool_use_id": "toolu_123"
}
```

---

## Current Hook: Session Capture

**Configuration** (`.claude/settings.json`):
```json
{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "command": "sl session capture"
          }
        ]
      }
    ]
  }
}
```

**Trigger logic** (`sl session capture`):
1. Read hook JSON from stdin
2. Check `tool_name == "Bash"` and command matches `git commit` (excluding `--amend`)
3. If no match → exit silently (exit 0)
4. Check credentials → if missing, exit silently (no error)
5. Extract transcript delta since last capture
6. Write to local cache (guaranteed)
7. Upload to Supabase Storage (best-effort)
8. Record metadata in database (best-effort)

**Delta tracking** — state file at `~/.specledger/session-state.json`:
```json
{
  "sessions": {
    "abc-123": {
      "last_offset": 45678,
      "last_commit": "abc123def",
      "transcript_path": "/home/user/.claude/projects/proj/abc-123.jsonl"
    }
  }
}
```

---

## Design Constraints

### Hooks vs CLI Commands

| Aspect | Hooks (L0) | CLI Commands (L1) |
|--------|-----------|-------------------|
| Invocation | Automatic, event-driven | Explicit, by user or agent |
| Failure mode | Silent cache + retry | Loud error + guidance |
| User visibility | Invisible | Visible output |
| Blocking | Never | May block for result |

### Cross-Layer Interactions

- **L1→L0**: `sl auth hook --install` configures hook entries in `.claude/settings.json`
- **L0→L1**: Hooks invoke `sl session capture` (a CLI command run in hook context)

The hook fires the CLI command, but the CLI command behaves differently in hook context — it checks stdin for hook JSON and applies silent-failure semantics.

---

## Future Hook Considerations

| Hook | Potential Use | Status |
|------|--------------|--------|
| `TaskCompleted` | Beads task session capture | Planned |
| `PreCompact` | Preserve reasoning before context compression | Investigating |
| `SessionEnd` | Final checkpoint / unresolved comment warning | Investigating |
| `Stop` | Partial session capture on manual stop | Future |

---

## References

**Design docs**:
- [4-Layer Model Overview](README.md)
- [CLI Design](cli.md)
- [Claude Code Hooks Guide](https://code.claude.com/docs/en/hooks-guide)

**Historical research** (decisions are inlined above; these are appendices for full context):
- [Session Capture Spike](../../specledger/010-checkpoint-session-capture/spike-hooks.md) — Evaluated 3 approaches, selected PostToolUse
- [Silent Session Capture Research](../../specledger/602-silent-session-capture/research.md) — Error logging, Sentry integration, slash command strategy
