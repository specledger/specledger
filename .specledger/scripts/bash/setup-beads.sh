#!/usr/bin/env bash
# setup-beads.sh - Initialize beads issue tracking system
# This script sets up the beads database after SpecLedger initialization
#
# Environment variables (passed from init.sh):
#   SPECLEDGER_PROJECT_SHORT_CODE - Short code for issue IDs (e.g., "sl", "myproj")

set -e

# Source common functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/common.sh"

# Colors for output
export RED='\033[0;31m'
export GREEN='\033[0;32m'
export YELLOW='\033[0;33m'
export BLUE='\033[0;34m'
export NC='\033[0m' # No Color

print_section() {
    echo -e "${BLUE}==>${NC} $1"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

main() {
    print_section "Setting up Beads Issue Tracking"

    # Check if bd is installed
    if ! command -v bd &> /dev/null; then
        print_error "bd (beads) is not installed"
        echo "  Please install beads first:"
        echo "  mise install bd   # or: ubi steveyegge/beads"
        exit 1
    fi

    # Check if already initialized
    if [ -d ".beads" ] && [ -f ".beads/metadata.json" ]; then
        print_warning "Beads directory already exists"
        read -p "Re-initialize beads? This will reset the database. [y/N]: " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            print_warning "Skipping beads initialization"
            exit 0
        fi
        rm -rf .beads
    fi

    # Get project short code from environment variable (set by init.sh from specledger.yaml)
    SHORT_CODE="${SPECLEDGER_PROJECT_SHORT_CODE:-}"

    # Fallback: try to read from specledger.yaml if env var not set
    if [ -z "$SHORT_CODE" ] && [ -f "specledger/specledger.yaml" ]; then
        SHORT_CODE=$(grep "short_code:" specledger/specledger.yaml | cut -d'"' -f2)
    fi

    # Default fallback
    if [ -z "$SHORT_CODE" ]; then
        print_warning "Could not determine short code"
        SHORT_CODE="sl"
    fi

    echo "  Short code: ${SHORT_CODE}"
    echo

    # Run bd init with the project prefix
    print_section "Initializing beads database"
    if bd init --prefix "$SHORT_CODE" 2>/dev/null || bd init; then
        print_success "Beads initialized successfully"
    else
        print_error "Failed to initialize beads"
        exit 1
    fi

    echo
    print_success "Beads is ready to use!"
    echo
    echo "Next steps:"
    echo "  bd create           Create a new issue"
    echo "  bd ready            Find work to do"
    echo "  bd --help           Show all commands"
}

main "$@"
