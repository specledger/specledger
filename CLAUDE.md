# specledger Development Guidelines

Auto-generated from all feature plans. Last updated: 2026-02-05

## Active Technologies
- File system (templates embedded in codebase at `templates/`, copied to user projects) (001-embedded-templates)
- Go 1.24+ (current: 1.24.2) + Cobra (CLI), Bubble Tea (TUI), go-git (v5), YAML v3, GoReleaser (006-opensource-readiness)
- GitHub repository (https://github.com/specledger/specledger), Documentation hosted separately (006-opensource-readiness)
- Go 1.24+ + GoReleaser v2, GitHub Actions, Homebrew (007-release-delivery-fix)
- N/A (release artifacts stored in GitHub Releases) (007-release-delivery-fix)
- Go 1.24+ + Cobra (CLI), go-git v5 (Git operations), YAML v3 (config parsing) (008-fix-sl-deps)
- File-based (specledger.yaml for metadata, ~/.specledger/cache/ for dependencies) (008-fix-sl-deps)
- Go 1.24.2 + Cobra (CLI framework), net/http (callback server), encoding/json (credential storage) (008-cli-auth)
- File-based (`~/.specledger/credentials.json`) with 0600 permissions (008-cli-auth)
- Go 1.24+ (CLI), JavaScript/Node.js (utility scripts), Bash (shell scripts) + Cobra (CLI), @supabase/supabase-js (Node.js scripts) (009-command-system-enhancements)
- File-based (`~/.specledger/credentials.json`, `.beads/issues.jsonl`, `scripts/audit-cache.json`) (009-command-system-enhancements)
- Go 1.24.2 + Cobra (CLI), net/http (Supabase REST + Storage API), compress/gzip (compression), encoding/json (serialization), go-git/v5 (commit detection) (010-checkpoint-session-capture)
- Supabase Storage (session content as gzip JSON) + Supabase PostgreSQL (session metadata via PostgREST) + local filesystem (offline queue, delta state) (010-checkpoint-session-capture)
- Go 1.24+ + Cobra (CLI framework), go-git v5 (branch detection), crypto/sha256 (ID generation) (591-issue-tracking-upgrade)
- File-based JSONL at `specledger/<spec>/issues.jsonl` (per-spec storage) (591-issue-tracking-upgrade)
- Go 1.24.2 + Cobra (CLI), Bubble Tea + Bubbles + Lipgloss (TUI), go-git v5, YAML v3 (011-streamline-onboarding)
- File-based (`.specledger/memory/constitution.md`, `specledger/specledger.yaml`) (011-streamline-onboarding)
- Markdown (prompt files), Go 1.24+ (embedding system) + Existing `sl deps` CLI, `sl issue` CLI commands (592-prompt-updates)
- N/A (documentation updates only) (592-prompt-updates)
- Go 1.24.2 + Cobra (CLI), go-git v5, gofrs/flock (file locking), gopkg.in/yaml.v3 (594-issues-storage-config)
- File-based (JSONL for issues, file locks for concurrency) (594-issues-storage-config)

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
- 594-issues-storage-config: Added Go 1.24.2 + Cobra (CLI), go-git v5, gofrs/flock (file locking), gopkg.in/yaml.v3
- 592-prompt-updates: Added Markdown (prompt files), Go 1.24+ (embedding system) + Existing `sl deps` CLI, `sl issue` CLI commands
- 591-issue-tracking-upgrade: Added Go 1.24+ + Cobra (CLI framework), go-git v5 (branch detection), crypto/sha256 (ID generation)


<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
