# specledger Development Guidelines

Auto-generated from all feature plans. Last updated: 2026-02-05

## Active Technologies
- File system (templates embedded in codebase at `templates/`, copied to user projects) (001-embedded-templates)
- Go 1.24+ (current: 1.24.2) + Cobra (CLI), Bubble Tea (TUI), go-git (v5), YAML v3, GoReleaser (006-opensource-readiness)
- GitHub repository (https://github.com/specledger/specledger), Documentation hosted separately (006-opensource-readiness)
- Go 1.24+ + GoReleaser v2, GitHub Actions, Homebrew (007-release-delivery-fix)
- N/A (release artifacts stored in GitHub Releases) (007-release-delivery-fix)
- Go 1.24.2 + Cobra (CLI framework), net/http (callback server), encoding/json (credential storage) (008-cli-auth)
- File-based (`~/.specledger/credentials.json`) with 0600 permissions (008-cli-auth)
- Go 1.24+ (CLI), JavaScript/Node.js (utility scripts), Bash (shell scripts) + Cobra (CLI), @supabase/supabase-js (Node.js scripts) (009-command-system-enhancements)
- File-based (`~/.specledger/credentials.json`, `.beads/issues.jsonl`, `scripts/audit-cache.json`) (009-command-system-enhancements)

- Go 1.24+ (004-thin-wrapper-redesign)

## Project Structure

```text
src/
tests/
```

## Commands

# Add commands for Go 1.24+

## Code Style

Go 1.24+: Follow standard conventions

## Recent Changes
- 009-command-system-enhancements: Added Go 1.24+ (CLI), JavaScript/Node.js (utility scripts), Bash (shell scripts) + Cobra (CLI), @supabase/supabase-js (Node.js scripts)
- 008-cli-auth: Added Go 1.24.2 + Cobra (CLI framework), net/http (callback server), encoding/json (credential storage)
- 007-release-delivery-fix: Added Go 1.24+ + GoReleaser v2, GitHub Actions, Homebrew


<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
