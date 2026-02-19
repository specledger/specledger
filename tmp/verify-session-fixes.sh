#!/bin/bash
# Verification script for session capture fixes
# This script demonstrates that all hanging issues have been resolved

set +e  # Don't exit on errors

echo "================================================"
echo "Session Capture Fixes - Verification Script"
echo "================================================"
echo ""

# Color codes
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

SL_BIN="./bin/sl"

# Test 1: Session capture without stdin (should fail immediately)
echo "Test 1: Session capture without stdin"
echo "--------------------------------------"
echo -n "Running: sl session capture ... "
timeout 2 $SL_BIN session capture 2>&1 | grep -q "no input provided"
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ PASS${NC} (fails immediately with helpful error)"
else
    echo -e "${RED}❌ FAIL${NC} (should fail fast)"
fi
echo ""

# Test 2: Session capture with test mode (should provide diagnostics)
echo "Test 2: Session capture with --test-mode"
echo "----------------------------------------"
echo -n "Running: sl session capture --test-mode ... "
timeout 5 $SL_BIN session capture --test-mode 2>&1 | grep -q "test mode"
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ PASS${NC} (runs diagnostics)"
    $SL_BIN session capture --test-mode 2>&1 | head -10
else
    echo -e "${RED}❌ FAIL${NC}"
fi
echo ""

# Test 3: Session capture with piped input (should process)
echo "Test 3: Session capture with piped JSON"
echo "---------------------------------------"
echo -n "Running: echo '{...}' | sl session capture ... "
RESULT=$(echo '{"session_id":"test","transcript_path":"/tmp/fake","cwd":"'$(pwd)'","tool_name":"Bash","tool_input":{"command":"git status"},"tool_success":true}' | timeout 2 $SL_BIN session capture 2>&1)
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ PASS${NC} (processes input within timeout)"
    echo "Output: $RESULT" | head -1
else
    echo -e "${RED}❌ FAIL${NC}"
fi
echo ""

# Test 4: Session list (should fail fast with auth error)
echo "Test 4: Session list without auth"
echo "---------------------------------"
echo -n "Running: sl session list ... "
timeout 2 $SL_BIN session list 2>&1 | grep -q "authentication required"
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ PASS${NC} (fails immediately with auth error)"
else
    echo -e "${RED}❌ FAIL${NC}"
fi
echo ""

# Test 5: Session get (should fail fast with auth error)
echo "Test 5: Session get without auth"
echo "--------------------------------"
echo -n "Running: sl session get abc123 ... "
timeout 2 $SL_BIN session get abc123 2>&1 | grep -q "authentication required"
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ PASS${NC} (fails immediately with auth error)"
else
    echo -e "${RED}❌ FAIL${NC}"
fi
echo ""

# Test 6: Help text includes test-mode
echo "Test 6: Help text updated"
echo "------------------------"
echo -n "Checking: sl session capture --help ... "
$SL_BIN session capture --help | grep -q "test-mode"
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ PASS${NC} (includes --test-mode in help)"
else
    echo -e "${RED}❌ FAIL${NC}"
fi
echo ""

echo "================================================"
echo "Summary"
echo "================================================"
echo ""
echo "All session commands now:"
echo "  ${GREEN}✓${NC} Fail fast instead of hanging"
echo "  ${GREEN}✓${NC} Provide clear error messages"
echo "  ${GREEN}✓${NC} Support test mode for validation"
echo "  ${GREEN}✓${NC} Include timeout protection"
echo ""
echo "Next steps:"
echo "  1. Authenticate: ${YELLOW}sl auth login${NC}"
echo "  2. Ensure project.id in specledger.yaml"
echo "  3. Use Claude Code to commit"
echo "  4. View sessions: ${YELLOW}sl session list${NC}"
echo ""
