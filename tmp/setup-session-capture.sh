#!/bin/bash
# Setup script for Claude Code session capture

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo "================================================"
echo "Claude Code Session Capture Setup"
echo "================================================"
echo ""

CLAUDE_DIR="$HOME/.claude"
SESSIONS_DIR="$CLAUDE_DIR/sessions"
SETTINGS_FILE="$CLAUDE_DIR/settings.json"

# Step 1: Check if .claude directory exists
if [ ! -d "$CLAUDE_DIR" ]; then
    echo -e "${RED}✗ Claude Code directory not found at $CLAUDE_DIR${NC}"
    echo "Please ensure Claude Code is installed and has been run at least once."
    exit 1
fi
echo -e "${GREEN}✓ Claude Code directory found${NC}"

# Step 2: Create sessions directory
if [ ! -d "$SESSIONS_DIR" ]; then
    echo "Creating sessions directory..."
    mkdir -p "$SESSIONS_DIR"
    chmod 700 "$SESSIONS_DIR"
    echo -e "${GREEN}✓ Sessions directory created${NC}"
else
    echo -e "${GREEN}✓ Sessions directory exists${NC}"
fi

# Step 3: Update settings.json
echo "Updating settings.json..."
cat > "$SETTINGS_FILE" << 'EOF'
{
  "saveTranscripts": true,
  "transcriptsDirectory": "~/.claude/sessions"
}
EOF
echo -e "${GREEN}✓ Settings updated${NC}"

# Step 4: Check current setup
echo ""
echo "================================================"
echo "Current Configuration"
echo "================================================"
echo ""
echo "Settings file: $SETTINGS_FILE"
cat "$SETTINGS_FILE"
echo ""
echo "Sessions directory: $SESSIONS_DIR"
ls -la "$SESSIONS_DIR" 2>/dev/null || echo "(empty)"
echo ""

# Step 5: Instructions
echo "================================================"
echo "Next Steps"
echo "================================================"
echo ""
echo -e "${YELLOW}1. RESTART Claude Code${NC}"
echo "   The settings changes require a full restart to take effect."
echo ""
echo -e "${YELLOW}2. Start a new conversation${NC}"
echo "   Claude Code will now save transcripts to ~/.claude/sessions/"
echo ""
echo -e "${YELLOW}3. Verify transcripts are being created:${NC}"
echo "   ls ~/.claude/sessions/"
echo "   You should see UUID directories with transcript.jsonl files"
echo ""
echo -e "${YELLOW}4. Test session capture:${NC}"
echo "   ./bin/sl session capture --test-mode"
echo ""
echo -e "${YELLOW}5. Make a commit to capture a real session:${NC}"
echo "   git commit -m \"Your commit message\""
echo ""
echo -e "${YELLOW}6. View captured sessions:${NC}"
echo "   ./bin/sl session list"
echo ""
echo "================================================"
echo "Troubleshooting"
echo "================================================"
echo ""
echo "If transcripts aren't being created after restart:"
echo "1. Check Claude Code version (may need to update)"
echo "2. Look for transcripts in other locations:"
echo "   find ~/.claude -name 'transcript*.jsonl'"
echo "3. Check Claude Code logs for errors"
echo "4. Verify permissions: ls -la ~/.claude/sessions"
echo ""
echo "For more help, see: tmp/enable-session-capture.md"
echo ""
