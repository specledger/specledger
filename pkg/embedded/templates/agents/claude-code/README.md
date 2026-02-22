# Claude Code Configuration

This directory contains Claude Code configuration files that are copied to new projects when "Claude Code" is selected as the coding agent.

## Session Capture Integration

The `.claude/settings.json` file configures a PostToolUse hook for SpecLedger session capture:

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

### How It Works

The **PostToolUse Hook** runs `sl session capture` after each Bash command to:
- Capture the current session state
- Upload to Supabase for project tracking (if configured)
- Associate with the project UUID for organization

### Verification

To verify the hook is working:
```bash
sl session list
```

This shows captured sessions for the current branch.

### Benefits

- Automatic session capture without manual intervention
- Sessions organized by project UUID in cloud storage
- Enables team visibility and session playback
- Integrates with SpecLedger's checkpoint system

### Customization

You can modify `.claude/settings.json` in your project to:
- Disable session capture: Remove the PostToolUse hook
- Add additional hooks: Extend the `hooks` configuration

For more information, see the [Claude Code documentation](https://docs.anthropic.com/claude-code).
