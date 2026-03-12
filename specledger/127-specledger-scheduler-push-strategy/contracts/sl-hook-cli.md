# CLI Contract: sl hook, sl approve & sl implement

**Feature**: 127-specledger-scheduler-push-strategy
**Date**: 2026-03-10

## New Commands

### sl approve

Mark a feature spec as approved, gating it for push-triggered implementation.

```
sl approve [--spec <spec-context>]
```

**Flags**:
| Flag     | Type   | Default          | Description                    |
|----------|--------|------------------|--------------------------------|
| --spec   | string | auto-detect from branch | Feature spec context to approve |

**Behavior**:
1. Resolve spec context (from `--spec` flag or current branch name)
2. Validate artifacts exist and are non-empty:
   - `specledger/<spec>/spec.md`
   - `specledger/<spec>/plan.md`
   - `specledger/<spec>/tasks.md`
3. Read spec.md, find `**Status**: <value>` line
4. If already "Approved", print message and exit 0
5. If "Draft", replace with "Approved" and write back
6. Print success message

**Exit codes**:
| Code | Meaning                                    |
|------|--------------------------------------------|
| 0    | Success (approved or already approved)     |
| 1    | Missing/empty artifacts (lists which ones) |
| 2    | Spec not found                             |

**Output examples**:
```
$ sl approve
Approved: 127-specledger-scheduler-push-strategy

$ sl approve
Already approved: 127-specledger-scheduler-push-strategy

$ sl approve
Error: cannot approve - missing artifacts:
  - specledger/127-specledger-scheduler-push-strategy/plan.md (not found)
  - specledger/127-specledger-scheduler-push-strategy/tasks.md (empty)
```

---

### sl hook install

Install the SpecLedger pre-push git hook.

```
sl hook install [--force]
```

**Flags**:
| Flag    | Type | Default | Description                         |
|---------|------|---------|-------------------------------------|
| --force | bool | false   | Overwrite existing SpecLedger block |

**Behavior**:
1. Check `.git/hooks/pre-push` exists
2. If exists, check for `# BEGIN SPECLEDGER PUSH HOOK` marker
   - If marker found and `--force` not set: print "already installed" and exit 0
   - If marker found and `--force` set: remove old block, add new block
3. If no marker: append SpecLedger block after existing content
4. If file doesn't exist: create with shebang + SpecLedger block
5. `chmod +x .git/hooks/pre-push`

**Exit codes**:
| Code | Meaning                        |
|------|--------------------------------|
| 0    | Success                        |
| 1    | Not a git repository           |

---

### sl hook uninstall

Remove the SpecLedger pre-push git hook.

```
sl hook uninstall
```

**Behavior**:
1. Read `.git/hooks/pre-push`
2. Remove content between `# BEGIN SPECLEDGER PUSH HOOK` and `# END SPECLEDGER PUSH HOOK` (inclusive)
3. If file is now empty (or only shebang), delete file
4. If no marker found, print "not installed" and exit 0

**Exit codes**:
| Code | Meaning                        |
|------|--------------------------------|
| 0    | Success (or not installed)     |

---

### sl hook status

Check if the SpecLedger push hook is installed.

```
sl hook status [--json]
```

**Flags**:
| Flag   | Type | Default | Description       |
|--------|------|---------|-------------------|
| --json | bool | false   | JSON output format |

**Output**:
```
$ sl hook status
Push hook: installed
Location: .git/hooks/pre-push

$ sl hook status --json
{"installed": true, "path": ".git/hooks/pre-push"}

$ sl hook status
Push hook: not installed
```

---

### sl hook execute (Internal)

Called by the pre-push hook script. Not intended for direct user invocation.

```
sl hook execute --event pre-push
```

**Behavior**:
1. Read current branch name
2. Check if branch matches feature pattern (`NNN-feature-name`)
   - If not: log "non-feature branch", exit 0
3. Resolve spec directory from branch name
4. Read spec.md, check `**Status**: Approved`
   - If not approved: log "not approved", exit 0
5. Check `.specledger/exec.lock`:
   - If exists and PID is alive: log "already running", exit 0
   - If exists and PID is dead: remove stale lock, log warning
6. Spawn `sl implement --feature <spec>` as detached background process
7. Log action to `.specledger/logs/push-hook.log`
8. Exit 0 (never block the push)

**Exit codes**:
| Code | Meaning                                |
|------|----------------------------------------|
| 0    | Always (errors are logged, not raised) |

---

### sl implement

Execute implementation for an approved feature by delegating to the Claude CLI. This is the core execution command spawned by `sl hook execute`.

```
sl implement --feature <spec-context>
```

**Flags**:
| Flag      | Type   | Default          | Description                          |
|-----------|--------|------------------|--------------------------------------|
| --feature | string | auto-detect from branch | Feature spec context to implement |

**Behavior**:
1. Resolve spec context (from `--feature` flag or current branch name)
2. Verify `claude` CLI is available in PATH (`exec.LookPath("claude")`)
3. Verify `.claude/commands/specledger.implement.md` exists
4. Acquire execution lock (write `.specledger/exec.lock` with PID, feature, timestamp)
   - If lock already held: print error and exit 1
5. Spawn Claude CLI: `claude -p "/specledger.implement" --dangerously-skip-permissions`
   - Working directory: project root
   - stdout/stderr redirected to `.specledger/logs/<feature>-claude.log`
6. Wait for Claude CLI process to complete
7. On completion (success or failure):
   - Remove `.specledger/exec.lock`
   - Write result summary to `.specledger/logs/<feature>-result.md`
8. Exit with Claude CLI's exit code

**Exit codes**:
| Code | Meaning                                       |
|------|-----------------------------------------------|
| 0    | Implementation completed successfully         |
| 1    | Lock held / missing prerequisites / claude not found |
| *    | Passthrough from Claude CLI exit code         |

**Output examples**:
```
$ sl implement --feature 127-specledger-scheduler-push-strategy
Implementing: 127-specledger-scheduler-push-strategy
Claude CLI log: .specledger/logs/127-specledger-scheduler-push-strategy-claude.log
Implementation complete.

$ sl implement
Error: claude CLI not found in PATH. Install Claude Code CLI first.

$ sl implement
Error: execution lock held (PID 12345, feature: 127-specledger-scheduler-push-strategy)
Run 'sl lock reset' to clear if the process is no longer running.
```

---

### sl lock reset

Manually remove the execution lock file. Used for recovery when a lock is left behind after a crash.

```
sl lock reset
```

**Behavior**:
1. If `.specledger/exec.lock` exists: remove it, print confirmation
2. If no lock file: print "no lock found"

**Exit codes**:
| Code | Meaning   |
|------|-----------|
| 0    | Always    |

---

### sl lock status

Display current execution lock information.

```
sl lock status [--json]
```

**Output**:
```
$ sl lock status
Lock held:
  PID: 12345
  Feature: 127-specledger-scheduler-push-strategy
  Started: 2026-03-10T14:30:00Z

$ sl lock status
No active execution lock.
```

---

## Hook Script Template

```bash
#!/bin/sh
# BEGIN SPECLEDGER PUSH HOOK
# Installed by: sl hook install
# Do not edit this block manually. Use 'sl hook uninstall' to remove.
sl hook execute --event pre-push "$@" 2>/dev/null || true
# END SPECLEDGER PUSH HOOK
```

## Execution Lock File Format

```json
{
  "pid": 12345,
  "feature": "127-specledger-scheduler-push-strategy",
  "started_at": "2026-03-10T14:30:00Z"
}
```
