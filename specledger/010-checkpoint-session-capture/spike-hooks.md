# Spike: Session Capture Approaches

**Date**: 2026-02-23
**Author**: Implementation team
**Status**: Complete

---

## Problem

When developers use AI assistants (Claude Code) to write code, the conversation that led to each commit is lost. This prevents:
- **Audit**: Reviewers can't see AI reasoning behind changes
- **Retrospectives**: Teams can't analyze AI-assisted patterns
- **Data mining**: No data for process improvement

**Constraint**: We need automatic capture at commit time without requiring developers to change their workflow.

---

## Executive Summary

We evaluated three approaches for capturing AI conversation sessions:
1. **Claude Code Hooks** (selected)
2. **AI-triggered command** (rejected)
3. **Wrapper binary** (rejected)

**Recommendation**: Use Claude Code's `PostToolUse` hook with matcher `Bash` to detect `git commit` commands. This provides automatic capture with direct transcript access and minimal user friction.

---

## Available Claude Code Hooks

Reference: [Claude Code Hooks Guide](https://code.claude.com/docs/en/hooks-guide)

| Hook | Trigger | Has Transcript Access | Suitable for Session Capture |
|------|---------|----------------------|------------------------------|
| `SessionStart` | New Claude Code session begins | ✓ | ❌ No checkpoint yet |
| `UserPromptSubmit` | User submits a prompt | ✓ | ❌ Too early, no work done |
| `PreToolUse` | Before tool execution | ✓ | ❌ Before commit happens |
| `PermissionRequest` | Tool needs permission | ✓ | ❌ Not commit-related |
| **`PostToolUse`** | **After tool completes** | **✓** | **✓ Can detect commit** |
| `PostToolUseFailure` | Tool execution failed | ✓ | ❌ Failed commits |
| `Notification` | System notification | ✓ | ❌ Not commit-related |
| `SubagentStart` | Subagent spawned | ✓ | ❌ Not commit-related |
| `SubagentStop` | Subagent completed | ✓ | ❌ Not commit-related |
| `Stop` | Session manually stopped | ✓ | ⚠️ Partial, no checkpoint |
| `TeammateIdle` | Teammate mode idle | ✓ | ❌ Not commit-related |
| `TaskCompleted` | Task finished | ✓ | ⚠️ Could work for tasks |
| `ConfigChange` | Settings changed | ✗ | ❌ Not relevant |
| `WorktreeCreate` | Worktree created | ✗ | ❌ Not relevant |
| `WorktreeRemove` | Worktree removed | ✗ | ❌ Not relevant |
| `PreCompact` | Before context compression | ✓ | ❌ Not a checkpoint |
| `SessionEnd` | Session ends | ✓ | ⚠️ No specific commit |

### Hook Selection: PostToolUse

**Rationale:**
- Fires immediately after `git commit` succeeds
- Provides `transcript_path` for extracting conversation
- Provides `session_id` for tracking deltas
- Can filter by `tool_name: Bash` and inspect `tool_input.command`
- Non-blocking (async hook execution)

**Hook Input Format** (discovered via testing):
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

## Alternative Approaches Evaluated

### Option 2: AI-Triggered Command

**Concept**: Add instruction to system prompts telling the AI to run `sl session capture <session-id>` after commits.

**Implementation:**
```markdown
# CLAUDE.md addition
After completing a git commit, run: `sl session capture $SESSION_ID`
```

**Pros:**
- Works across any AI tool (not Claude Code specific)
- AI can pass context about what was committed
- More explicit in conversation flow

**Cons:**
- ❌ Relies on AI compliance (may forget or skip)
- ❌ Session ID not available to AI (internal to Claude Code)
- ❌ No direct transcript access from command
- ❌ Requires modifying all project CLAUDE.md files
- ❌ Different AI models may not follow instructions consistently

**Verdict**: **Rejected** - Too unreliable, session ID not accessible to AI.

---

### Option 3: Wrapper Binary (`sl claude`)

**Concept**: Launch Claude Code through `sl` wrapper that captures session on exit.

**Implementation:**
```go
// cmd/sl/commands/claude.go
func runClaude(cmd *cobra.Command, args []string) error {
    // Start Claude Code as subprocess
    proc := exec.Command("claude", args...)
    proc.Stdin = os.Stdin
    proc.Stdout = os.Stdout
    proc.Stderr = os.Stderr

    err := proc.Run()

    // On exit, find and capture the session
    sessionID := findLatestSession()
    captureSession(sessionID)

    return err
}
```

**Pros:**
- Guaranteed capture on session end
- Works regardless of hook configuration
- Can capture entire session (not just delta)

**Cons:**
- ❌ Requires users to change workflow (`sl claude` instead of `claude`)
- ❌ Only captures on session end, not per-commit
- ❌ Complex subprocess management (PTY handling, signal forwarding)
- ❌ Claude Code already has a "resume" feature that complicates session boundaries
- ❌ No way to know which commits happened during session
- ❌ Miss beads task boundaries

**Verdict**: **Rejected** - Requires workflow change, only captures at session end, loses commit-to-session mapping.

---

## Detailed Comparison

| Criteria | PostToolUse Hook | AI-Triggered Command | Wrapper Binary |
|----------|------------------|---------------------|----------------|
| Automatic capture | ✓ Yes | ❌ AI may forget | ✓ Yes |
| Per-commit granularity | ✓ Yes | ⚠️ If AI complies | ❌ Session-level only |
| Transcript access | ✓ Direct path | ❌ Not available | ⚠️ After session ends |
| User workflow change | ✓ None | ⚠️ Update CLAUDE.md | ❌ Must use `sl claude` |
| Claude Code dependency | Yes | No | Yes |
| Delta capture support | ✓ Yes | ❌ No session tracking | ❌ No intermediate state |
| Beads task support | ✓ Extensible | ⚠️ If AI complies | ❌ No task boundaries |
| Error handling | ✓ Hook retry | ❌ No retry | ✓ Exit code handling |

---

## Implementation Details for Selected Approach

### Configuration

`.claude/settings.json`:
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

### Trigger Logic

The `sl session capture` command:
1. Reads hook JSON from stdin
2. Checks if `tool_name == "Bash"` and command matches `git commit` (excluding `--amend`)
3. If no, exits silently (no capture needed)
4. If yes, extracts delta from transcript since last capture
5. Compresses and uploads to Supabase Storage
6. Records metadata in database

### Delta Tracking

State file at `~/.specledger/session-state.json`:
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

## Future Considerations

### Beads Task Capture

The `TaskCompleted` hook could be used for beads task sessions:
```json
{
  "hooks": {
    "TaskCompleted": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "sl session capture --task"
          }
        ]
      }
    ]
  }
}
```

### Multi-AI Support

If future requirements need non-Claude-Code support:
1. Option 2 (AI-triggered) could be revisited with explicit session ID parameter
2. A hybrid approach: hooks for Claude Code, manual command for others

### Revisiting the Wrapper Approach

The wrapper binary approach (`sl claude`) was rejected for initial implementation, but may be worth revisiting if:
- We develop more "guided flow" commands following the `sl revise` pattern
- Users are already invoking Claude through `sl` for specific workflows
- We need session capture for task-specific contexts (not just commits)

In that case, the workflow change concern diminishes because users are already using `sl` as their entry point.

---

## Conclusion

**Selected: PostToolUse Hook with Bash matcher**

This approach provides:
- ✓ Automatic, transparent capture
- ✓ Per-commit granularity
- ✓ Direct transcript access
- ✓ No user workflow changes
- ✓ Delta capture for efficient storage
- ✓ Extensible to beads tasks via additional hooks

The main trade-off is Claude Code dependency, which is acceptable given:
- specledger is designed for Claude Code workflows
- The team uses Claude Code as the primary AI coding tool
- Alternative approaches have significant reliability or UX issues
