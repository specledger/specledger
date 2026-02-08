#!/usr/bin/env bash
# init.sh - Post-initialization script for SpecLedger playbook
# This script runs automatically after 'sl init' or 'sl new'
#
# Environment variables available from specledger.yaml:
#   SPECLEDGER_PROJECT_NAME        - Project name
#   SPECLEDGER_PROJECT_SHORT_CODE  - Short code for issue IDs
#   SPECLEDGER_PROJECT_VERSION     - Project version
#   SPECLEDGER_PLAYBOOK_NAME       - Playbook name
#   SPECLEDGER_PLAYBOOK_VERSION    - Playbook version
#
# This script is executed from the template directory but runs commands
# in the target project directory ($SPECLEDGER_PROJECT_ROOT).

set -e

# Target project root is passed as environment variable
PROJECT_ROOT="${SPECLEDGER_PROJECT_ROOT:-}"

if [ -z "$PROJECT_ROOT" ]; then
    echo "Error: SPECLEDGER_PROJECT_ROOT not set"
    exit 1
fi

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
    cd "$PROJECT_ROOT"

    # Export specledger environment variables for child scripts
    export SPECLEDGER_PROJECT_NAME="${SPECLEDGER_PROJECT_NAME:-}"
    export SPECLEDGER_PROJECT_SHORT_CODE="${SPECLEDGER_PROJECT_SHORT_CODE:-}"
    export SPECLEDGER_PROJECT_VERSION="${SPECLEDGER_PROJECT_VERSION:-}"
    export SPECLEDGER_PLAYBOOK_NAME="${SPECLEDGER_PLAYBOOK_NAME:-}"
    export SPECLEDGER_PLAYBOOK_VERSION="${SPECLEDGER_PLAYBOOK_VERSION:-}"

    # Run setup-beads.sh if it exists in the target project
    if [ -f ".specledger/scripts/bash/setup-beads.sh" ]; then
        bash .specledger/scripts/bash/setup-beads.sh
    else
        print_warning "setup-beads.sh not found - skipping beads initialization"
    fi

    echo
    print_success "SpecLedger initialization complete!"
}

main "$@"
