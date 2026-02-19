# How to Enable Session Capturing in Claude Code

## What I Just Did ✅

I've configured Claude Code to save session transcripts:

1. **Updated `~/.claude/settings.json`** with:
   ```json
   {
     "saveTranscripts": true,
     "transcriptsDirectory": "/Users/ngoctran/.claude/sessions"
   }
   ```

2. **Created the sessions directory**: `~/.claude/sessions`

## Next Steps

### Step 1: Restart Claude Code
The settings changes require a restart to take effect:
```bash
# If using CLI
exit  # or Ctrl+D

# Then restart Claude Code
claude-code
```

### Step 2: Verify Transcripts Are Being Created

After restarting and starting a new conversation:

```bash
# Check if sessions directory has content
ls -la ~/.claude/sessions/

# You should see directories with UUIDs like:
# drwxr-xr-x  3 user  staff   96 Feb 18 16:00 abc123-def4-5678-90ab-cdef12345678
```

Each session directory should contain:
- `transcript.jsonl` - The conversation transcript

### Step 3: Test Session Capture

Once transcripts are being created, test the capture:

```bash
# Test mode to verify everything is working
./bin/sl session capture --test-mode
```

Expected output:
```
Running in test mode...
✓ Git repository detected
✓ Project ID found: 7109364a-2ebc-451f-b052-2fbe5453459e
✓ Authenticated as: user@example.com
✓ Claude Code sessions directory found
✓ Found transcript: /Users/ngoctran/.claude/sessions/abc123.../transcript.jsonl
✓ Session ID: abc123...

✅ Session capture system is configured correctly!
```

### Step 4: Make a Commit to Capture a Session

Now when you make a git commit with Claude Code, the session will be captured:

```bash
# Work with Claude Code as normal
# Make changes to your code
# Commit the changes

git add .
git commit -m "Your commit message"

# The session will be automatically captured!
```

## Troubleshooting

### Issue: Settings not taking effect

**Solution**: Make sure to fully restart Claude Code (not just start a new session)

### Issue: transcripts.jsonl vs transcript.jsonl

Some versions of Claude Code may use different naming conventions:
- `transcript.jsonl` (current standard)
- `transcripts.jsonl` (older versions)

Our code checks for `transcript.jsonl` in session subdirectories.

### Issue: Transcripts in different location

If your Claude Code stores transcripts elsewhere, you can:

1. Check where they actually are:
   ```bash
   find ~/.claude -name "*transcript*.jsonl" 2>/dev/null
   ```

2. Update the session capture code to look in the correct location, or symlink:
   ```bash
   ln -s /path/to/actual/transcripts ~/.claude/sessions
   ```

### Issue: No session ID directories created

If the sessions directory remains empty, Claude Code might be using a different storage mechanism. Alternative approaches:

#### Option A: Use hooks to capture (recommended)

Instead of relying on transcripts, we can modify the git hooks to capture sessions at commit time using Claude Code's hook system.

#### Option B: Check Claude Code version

```bash
# Check version
claude-code --version

# Update to latest version if needed
brew upgrade claude-code  # or your package manager
```

### Issue: Permission denied

```bash
# Ensure correct permissions
chmod 700 ~/.claude/sessions
chown -R $USER ~/.claude/sessions
```

## Alternative: Hook-Based Capture

If transcript-based capture doesn't work, we can use Claude Code's hook system:

### Create a post-tool hook

1. **Create hook configuration** at `.claude/hooks/hooks.json`:
   ```json
   {
     "tool-post": {
       "enabled": true,
       "command": "sl session capture",
       "stdin": "json"
     }
   }
   ```

2. This will automatically call `sl session capture` after each tool execution and pass the hook data via stdin.

3. When you run git commands through Claude Code, the session will be captured automatically.

## Verification Checklist

- [ ] Settings updated in `~/.claude/settings.json`
- [ ] Sessions directory created at `~/.claude/sessions`
- [ ] Claude Code restarted
- [ ] New session started
- [ ] Transcript file created in sessions directory
- [ ] `sl session capture --test-mode` passes
- [ ] Made test commit and session was captured
- [ ] `sl session list` shows captured sessions

## Additional Configuration

### Storage Location

You can customize where transcripts are saved by modifying settings.json:

```json
{
  "saveTranscripts": true,
  "transcriptsDirectory": "/custom/path/to/sessions"
}
```

### Privacy Settings

If you want to exclude certain patterns from transcripts:

```json
{
  "saveTranscripts": true,
  "transcriptsDirectory": "/Users/ngoctran/.claude/sessions",
  "transcriptExcludePatterns": [
    "password",
    "api_key",
    "secret"
  ]
}
```

## Testing Your Setup

Run this command to verify everything:

```bash
./bin/sl session capture --test-mode
```

This will check:
1. ✓ Git repository
2. ✓ Project ID
3. ✓ Authentication
4. ✓ Claude Code sessions directory
5. ✓ Active transcript files

Once all checks pass with ✓, you're ready to capture sessions!

## Using Session Capture

### Capture happens automatically

Session capture happens automatically when you commit:
```bash
# Just commit as normal
git commit -m "Add feature X"

# Session is captured behind the scenes
```

### View captured sessions

```bash
# List all sessions for current branch
sl session list

# Get specific session by commit hash
sl session get abc123

# Get session by task ID
sl session get SL-42
```

### Upload queued sessions

If a session failed to upload due to network issues:
```bash
sl session sync
```

## Questions?

If you encounter issues:
1. Check `sl session capture --test-mode` output
2. Verify settings.json is valid JSON
3. Check permissions on ~/.claude/sessions
4. Ensure Claude Code version is up to date
5. Check for transcript files: `find ~/.claude -name "transcript.jsonl"`
