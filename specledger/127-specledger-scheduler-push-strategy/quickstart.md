# Quickstart: Push-Triggered Scheduler Strategy

**Feature**: 127-specledger-scheduler-push-strategy

## Developer Workflow

### 1. Install the push hook
```bash
sl hook install
# Output: Push hook installed at .git/hooks/pre-push
```

### 2. Create and plan a feature (existing workflow)
```bash
# Specify, clarify, plan, generate tasks (existing SpecLedger commands)
/specledger.specify
/specledger.plan
/specledger.tasks
```

### 3. Approve the spec
```bash
sl approve
# Output: Approved: 127-specledger-scheduler-push-strategy
```

### 4. Push to trigger implementation
```bash
git add -A && git commit -m "approve spec for implementation"
git push
# Hook detects approved spec, spawns sl implement in background
# sl implement delegates to: claude -p "/specledger.implement" --dangerously-skip-permissions
# Output: [SpecLedger] Triggering implementation for 127-specledger-scheduler-push-strategy
```

### 5. Review generated code
```bash
# Check implementation progress
git fetch origin
git diff 127-specledger-scheduler-push-strategy..127-specledger-scheduler-push-strategy/implement

# Check execution log
cat .specledger/logs/push-hook.log
```

## Management Commands

```bash
# Check hook status
sl hook status

# Remove hook
sl hook uninstall

# Reinstall (force overwrite)
sl hook install --force
```

## File Locations

| File | Purpose |
|------|---------|
| `.git/hooks/pre-push` | Git hook script (managed by sl) |
| `.specledger/exec.lock` | Execution lock (prevents duplicate runs) |
| `.specledger/logs/push-hook.log` | Hook execution history |
| `.specledger/logs/<feature>-result.md` | Implementation results summary |
| `.specledger/logs/<feature>-claude.log` | Claude CLI execution output |

## Troubleshooting

**Hook not triggering?**
1. Run `sl hook status` to verify installation
2. Check spec is approved: look for `**Status**: Approved` in spec.md
3. Verify branch name matches `NNN-feature-name` pattern
4. Verify `claude` CLI is available: `which claude`

**Implementation already running?**
- Check `.specledger/exec.lock` for PID
- If process is dead, run `sl lock reset` to clear the stale lock

**Want to see what happened?**
- Check `.specledger/logs/push-hook.log` for hook detection/trigger activity
- Check `.specledger/logs/<feature>-claude.log` for Claude CLI execution output
