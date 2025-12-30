#!/usr/bin/env bash
#
# sl - Bootstrap a new SpecLedger project
#
# This script creates a new project with all SpecLedger infrastructure:
# - Claude Code skills and commands
# - Beads issue tracker
# - SpecKit templates and scripts
# - Tool configuration (mise)
#

set -e

# ============================================================================
# Configuration
# ============================================================================

DEMO_DIR=""
REPO_ROOT=""

# ============================================================================
# Error Handling
# ============================================================================

cleanup_on_error() {
    local exit_code=$?

    # Only cleanup on error (non-zero exit) or Ctrl+C
    if [[ $exit_code -eq 0 ]]; then
        return 0
    fi

    if [[ $exit_code -eq 130 ]]; then
        # User pressed Ctrl+C, just exit
        echo ""
        exit 130
    fi

    if [[ -n "$DEMO_DIR" ]] && [[ -d "$DEMO_DIR" ]]; then
        if command -v gum &> /dev/null; then
            gum log --level error "Cleaning up $DEMO_DIR"
        else
            echo "ERROR: Cleaning up $DEMO_DIR"
        fi
        rm -rf "$DEMO_DIR"
    fi
}

trap cleanup_on_error EXIT

# ============================================================================
# Pre-flight Checks
# ============================================================================

check_repo_root() {
    if [[ ! -d "templates" ]] || [[ ! -f "mise.toml" ]]; then
        echo "ERROR: Must run from specledger repository root"
        echo "Expected to find templates/ directory and mise.toml file"
        exit 1
    fi
}

check_gum_available() {
    if ! command -v gum &> /dev/null; then
        cat <<EOF
ERROR: gum not found. Install from https://github.com/charmbracelet/gum
  • macOS:  brew install gum
  • Linux:  go install github.com/charmbracelet/gum@latest
  • Or see: https://github.com/charmbracelet/gum#installation
EOF
        exit 1
    fi
}

check_mise_available() {
    if ! command -v mise &> /dev/null; then
        gum log --level error "mise not found. Install from https://mise.jdx.dev"
        echo "  • macOS:  brew install mise"
        echo "  • Linux:  curl https://mise.run | sh"
        echo "  • Or see: https://mise.jdx.dev/getting-started.html"
        exit 1
    fi
}

ensure_demos_dir() {
    mkdir -p ~/demos
}

# ============================================================================
# Interactive Prompts
# ============================================================================

prompt_project_name() {
    local error_msg=""

    while true; do
        # Show error from previous iteration if any
        if [[ -n "$error_msg" ]]; then
            gum log --level error "$error_msg"
            echo ""
        fi

        name=$(gum input --placeholder "Project name (alphanumeric, hyphens, underscores)") || exit 130

        # Clear error for this iteration
        error_msg=""

        # Check if input is empty
        if [[ -z "$name" ]]; then
            error_msg="Project name cannot be empty"
            continue
        fi

        # Validation
        if [[ ! "$name" =~ ^[a-zA-Z0-9_-]+$ ]]; then
            error_msg="Name must be alphanumeric, hyphens, or underscores only"
            continue
        fi

        if [[ -d ~/demos/$name ]]; then
            error_msg="~/demos/$name already exists"
            continue
        fi

        echo "$name"
        return 0
    done
}

prompt_short_code() {
    local project_name="$1"
    local error_msg=""

    # Generate default: first 3 letters, lowercase
    local default_code
    default_code=$(echo "$project_name" | tr '[:upper:]' '[:lower:]' | head -c 3)

    while true; do
        # Show error from previous iteration if any
        if [[ -n "$error_msg" ]]; then
            gum log --level error "$error_msg"
            echo ""
        fi

        local code
        code=$(gum input --header "Short code for Beads issue prefix (2-4 letters):" --prompt "> " --placeholder "$default_code" --value "$default_code") || exit 130

        # Clear error for this iteration
        error_msg=""

        # Check if input is empty
        if [[ -z "$code" ]]; then
            error_msg="Short code cannot be empty"
            continue
        fi

        # Convert to lowercase
        code=$(echo "$code" | tr '[:upper:]' '[:lower:]')

        # Validation: 2-4 lowercase letters only
        if [[ ! "$code" =~ ^[a-z]{2,4}$ ]]; then
            error_msg="Short code must be 2-4 lowercase letters only"
            continue
        fi

        echo "$code"
        return 0
    done
}

prompt_playbook() {
    echo "" >&2

    local playbook
    playbook=$(gum choose --header "Select Playbook:" \
        "Default (General SWE)" \
        "Data Science" \
        "Platform Engineering" \
        "Custom")

    if [[ "$playbook" == "Advanced"* ]]; then
        gum log --level warn "Advanced playbook not yet implemented"
        gum log --level warn "Falling back to Default"
        echo "Default (General SWE)"
    fi
}

prompt_agent_shell() {
    echo "" >&2

    local shell
    shell=$(gum choose --header "Select Agent Shell:" \
        "Claude Code" \
        "Gemini CLI" \
        "Codex")

    if [[ "$shell" != "Claude Code"* ]]; then
        gum log --level warn "Only Claude Code is currently implemented"
        gum log --level warn "Falling back to Claude Code"
        echo "Claude Code"
    else
        echo "Claude Code"
    fi
}

# ============================================================================
# Bootstrap File Operations
# ============================================================================

copy_bootstrap_files() {
    local repo_root="$1"
    local demo_dir="$2"

    # Copy from templates/
    if [[ -d "$repo_root/templates/.claude" ]]; then
        cp -R "$repo_root/templates/.claude" "$demo_dir/"
    else
        gum style --foreground 196 "ERROR: templates/.claude not found"
        return 1
    fi

    if [[ -d "$repo_root/templates/.beads" ]]; then
        cp -R "$repo_root/templates/.beads" "$demo_dir/"
    else
        gum style --foreground 196 "ERROR: templates/.beads not found"
        return 1
    fi

    if [[ -d "$repo_root/templates/specledger" ]]; then
        cp -R "$repo_root/templates/specledger" "$demo_dir/.specify"
    else
        gum style --foreground 196 "ERROR: templates/specledger not found"
        return 1
    fi

    # Copy individual files from templates/
    local template_files=(
        "mise.toml"
        "AGENTS.md"
        ".gitattributes"
    )

    for file in "${template_files[@]}"; do
        if [[ -f "$repo_root/templates/$file" ]]; then
            cp "$repo_root/templates/$file" "$demo_dir/"
        else
            gum log --level error "templates/$file not found"
            return 1
        fi
    done

    # Verify critical files exist
    local critical_files=(
        ".specify/templates/spec-template.md"
        ".specify/memory/constitution.md"
        ".beads/config.yaml"
        ".claude/commands/specledger.specify.md"
        ".claude/skills/bd-issue-tracking/SKILL.md"
        "mise.toml"
        "AGENTS.md"
        ".gitattributes"
    )

    for file in "${critical_files[@]}"; do
        if [[ ! -f "$demo_dir/$file" ]]; then
            gum log --level error "Critical file missing: $file"
            return 1
        fi
    done

    gum log --level info "Bootstrap files copied successfully"
}

# ============================================================================
# Configuration Updates
# ============================================================================

update_beads_config() {
    local demo_dir="$1"
    local short_code="$2"

    # Update issue-prefix in config.yaml
    if [[ -f "$demo_dir/.beads/config.yaml" ]]; then
        sed -i.bak "s/^issue-prefix: .*/issue-prefix: \"$short_code\"/" "$demo_dir/.beads/config.yaml"
        rm -f "$demo_dir/.beads/config.yaml.bak"
    fi
}

reset_state_files() {
    local demo_dir="$1"

    # Empty beads issues
    if [[ -f "$demo_dir/.beads/issues.jsonl" ]]; then
        echo "" > "$demo_dir/.beads/issues.jsonl"
    fi

    # Update metadata
    if [[ -d "$demo_dir/.beads" ]]; then
        cat > "$demo_dir/.beads/metadata.json" <<'EOF'
{
  "database": "beads.db",
  "jsonl_export": "issues.jsonl",
  "last_bd_version": "0.28.0"
}
EOF
    fi

    gum log --level info "Configuration updated successfully"
}

# ============================================================================
# Git Initialization
# ============================================================================

initialize_git_repo() {
    local demo_dir="$1"
    local project_name="$2"

    cd "$demo_dir"

    git init > /dev/null 2>&1

    # Configure beads merge driver
    git config merge.beads.driver "bd merge %O %A %B %P"
    git config merge.beads.name "Beads 3-way JSONL merge"

    # Initial commit
    git add . > /dev/null 2>&1
    git commit -m "chore: Bootstrap SpecLedger project

Generated with sl script.
Project: $project_name

Includes:
- SpecLedger templates and scripts
- Beads issue tracker
- Claude Code skills
- Mise tool configuration" > /dev/null 2>&1

    cd - > /dev/null

    gum log --level info "Git repository initialized"
}

# ============================================================================
# Tool Installation
# ============================================================================

install_tools() {
    local demo_dir="$1"

    echo ""
    gum spin --spinner dot --title "Installing tools (bd, perles, ...)..." \
        -- bash -c "cd '$demo_dir' && mise trust && mise install"

    gum log --level info "Tools installed successfully"
}

initialize_beads() {
    local demo_dir="$1"

    echo ""
    gum spin --spinner dot --title "Initializing Beads issue tracker..." \
        -- bash -c "cd '$demo_dir' && bd init"

    gum log --level info "Beads initialized successfully"
}

# ============================================================================
# Next Steps Output
# ============================================================================

print_next_steps() {
    local demo_dir="$1"
    local project_name="$2"

    echo ""
    gum style \
        --foreground 212 --border double --align center \
        --width 70 --margin "1" --padding "1 2" \
        "✓ Project successfully created!"

    echo ""
    gum style --bold "Project location:" --foreground 212
    echo "  ~/demos/$project_name"
    echo ""

    gum style --bold --underline "Next Steps:"
    echo ""

    gum style --foreground 212 "1. Open the project in Claude Code:"
    gum style --italic "   cd ~/demos/$project_name"
    gum style --italic "   claude"
    echo ""

    gum style --foreground 212 "2. Create your project constitution (copy-paste this prompt):"
    echo ""
    gum style \
        --border rounded --padding "1 2" --width 65 \
        "Create principles focused on YAGNI, KISS, user" \
        "experience consistency, and performance requirements." \
        "Include governance for how these principles should" \
        "guide technical decisions and implementation choices."
    echo ""

    gum style --foreground 212 "3. Start your first feature workflow:"
    gum style --italic "   (Feature prompt to be added - see demo script for details)"
    echo ""

    gum style --bold --underline "Available SpecLedger Commands:"
    cat <<'EOF'
  • /specledger.specify      - Create feature specification
  • /specledger.clarify      - Resolve spec ambiguities
  • /specledger.plan         - Generate implementation plan
  • /specledger.tasks        - Create task breakdown
  • /specledger.implement    - Execute tasks
EOF
    echo ""

    gum style --bold --underline "Issue Tracking (Beads):"
    cat <<'EOF'
  • bd ready             - Find unblocked work
  • bd create            - Create new issue
  • bd show <id>         - View issue details
  • bd sync              - Sync with git remote
EOF
    echo ""

    gum style --bold --underline "TUI Dashboard:"
    echo "  • perles               - Launch kanban board"
    echo ""

    gum style --bold --underline "Learn More:"
    cat <<'EOF'
  • AGENTS.md            - Workflow documentation
  • CLAUDE.md            - Best practices
  • .beads/README.md     - Issue tracking guide
EOF
    echo ""
}

# ============================================================================
# Main Execution
# ============================================================================

main() {
    # Pre-flight checks (must be first)
    check_repo_root
    check_gum_available
    check_mise_available

    # Get repository root (needed for prompts)
    REPO_ROOT=$(git rev-parse --show-toplevel)

    # Ensure demos directory exists
    ensure_demos_dir

    # Display header
    gum style \
        --foreground 212 --bold --border normal --align center \
        --width 60 --margin "1 0" --padding "1 2" \
        "SpecLedger bootstrapper"

    echo ""

    # Interactive prompts
    PROJECT_NAME=$(prompt_project_name)
    SHORT_CODE=$(prompt_short_code "$PROJECT_NAME")
    PLAYBOOK=$(prompt_playbook)
    AGENT_SHELL=$(prompt_agent_shell)

    echo ""
    gum log --level info "Creating project: $PROJECT_NAME (beads prefix: $SHORT_CODE)"
    echo ""

    # Create demo directory
    DEMO_DIR=~/demos/$PROJECT_NAME
    mkdir -p "$DEMO_DIR"

    # Bootstrap operations
    copy_bootstrap_files "$REPO_ROOT" "$DEMO_DIR"
    update_beads_config "$DEMO_DIR" "$SHORT_CODE"
    reset_state_files "$DEMO_DIR"
    initialize_git_repo "$DEMO_DIR" "$PROJECT_NAME"
    install_tools "$DEMO_DIR"
    initialize_beads "$DEMO_DIR"

    # Print next steps
    print_next_steps "$DEMO_DIR" "$PROJECT_NAME"
}

# Run main function
main "$@"
