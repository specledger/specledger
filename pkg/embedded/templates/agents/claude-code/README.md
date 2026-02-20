# Claude Code Configuration

This directory contains Claude Code configuration files that are copied to new projects when "Claude Code" is selected as the coding agent.

## Session Capture Integration

The `.claude/settings.json` file configures automatic session capture for SpecLedger integration:

```json
{
  "saveTranscripts": true,
  "transcriptsDirectory": "~/.claude/sessions",
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Bash",
        "command": "sl session capture"
      }
    ]
  }
}
```

### How It Works

1. **saveTranscripts**: Enables saving conversation transcripts locally
2. **transcriptsDirectory**: Stores transcripts in `~/.claude/sessions`
3. **PostToolUse Hook**: After each Bash command, runs `sl session capture` to:
   - Capture the current session state
   - Upload to Supabase for project tracking (if configured)
   - Associate with the project UUID for organization

### Benefits

- Automatic session capture without manual intervention
- Sessions organized by project UUID in cloud storage
- Enables team visibility and session playback
- Integrates with SpecLedger's checkpoint system

### Customization

You can modify `.claude/settings.json` in your project to:
- Disable session capture: Set `saveTranscripts: false`
- Change transcript location: Update `transcriptsDirectory`
- Add additional hooks: Extend the `hooks` configuration

For more information, see the [Claude Code documentation](https://docs.anthropic.com/claude-code).
