# Research: CLI Unification

**Feature**: CLI Unification (003-cli-unification)
**Date**: 2026-01-30
**Status**: Complete

## Prior Work

Related work identified in Beads issue tracker:

### Epic: CLI Integration & Distribution (sl-29y)
- **sl-29y.1**: CLI Version Injection (setup phase)
- **sl-29y.2**: Enhance Makefile for CLI Build (setup phase)
- **sl-29y.3**: Update sl Script to Build CLI (setup phase)
- **sl-29y.4**: Add README.md for CLI Distribution
- **sl-29y.5**: Test CLI Available After sl Bootstrap
- **sl-29y.6**: Implement CLI Binary Build Verification in sl
- **sl-29y.7**: Test sl with Existing Projects
- **sl-29y.8**: Test All CLI Commands Work After sl
- **sl-29y.9**: Implement Cross-Platform CLI Builds
- **sl-29y.10**: Test Cross-Platform CLI Binaries
- **sl-29y.11**: Create Distribution Archive Targets
- **sl-29y.12**: Test Vendor Specs for Offline Development
- **sl-29y.13**: Implement Vendor Directory Clean Up

**Summary**: This epic covers CLI distribution, cross-platform builds, and testing. However, it focuses on making the existing `sl` script build the CLI, not integrating them into a single CLI tool with shared functionality. Our feature extends this work to unify the CLI and add TUI integration.

## Decision 1: CLI Framework Selection

**Decision**: Continue using Cobra (spf13/cobra) as the CLI framework.

**Rationale**:
- Project already uses Cobra (`cmd/main.go` confirms this)
- Excellent support for complex multi-level command hierarchies
- Built-in support for both interactive and non-interactive modes
- Pre-run hooks for detecting terminal capabilities
- Active maintenance with extensive documentation and community support
- Better suited than urfave/cli for applications requiring sophisticated command structures

**Alternatives Considered**:
- **urfave/cli**: Simpler API, excellent auto-completion, but lacks hierarchical command structure that our project needs
- **yargs** (Node.js): Not applicable - project uses Go
- **clap** (Rust): Not applicable - project uses Go

**Impact**: Minimal - project already committed to Cobra, no migration needed.

---

## Decision 2: TUI Integration Pattern

**Decision**: Use Bubble Tea library for TUI components with automatic fallback to plain CLI mode.

**Architecture Pattern**:

1. **Terminal Detection**:
   ```go
   func isInteractiveTerminal() bool {
     return term.IsTerminal(int(os.Stdin.Fd())) &&
            !strings.Contains(os.Getenv("TERM"), "dumb")
   }
   ```

2. **Hybrid Execution Paths**:
   - Interactive mode: Use Bubble Tea TUI for bootstrap and complex commands
   - Non-interactive mode: Use plain CLI with flags
   - CI detection: Use `--ci` flag or environment variable `CI=true`

3. **Command Structure**:
   - `sl bootstrap` or `sl new`: Interactive TUI for project creation
   - `sl deps [subcommands]`: Use TUI for complex operations, plain text for simple
   - `sl --help`, `sl --version`: Always show in plain text

**Rationale**:
- Bubble Tea is the de facto standard for Go TUIs
- Well-maintained with good documentation
- Easy to integrate with Cobra
- Provides automatic fallback patterns

**Alternatives Considered**:
- **Gum**: CLI wrapper around Bubble Tea - our project already uses gum, but for bootstrap we need more control
- **stdio-based TUIs**: Less modern, harder to maintain
- **Web-based UI**: Overkill for CLI tool

**Impact**: High - this is the core integration pattern for the unified CLI.

---

## Decision 3: Cross-Platform Distribution Strategy

**Decision**: Use GoReleaser with GitHub Actions for automated cross-platform binary distribution.

**Implementation**:

1. **GoReleaser Configuration** (`.goreleaser.yaml`):
   ```yaml
   builds:
     - env:
         - CGO_ENABLED=0
       goos:
         - linux
         - darwin
         - windows
       goarch:
         - amd64
         - arm64
   archives:
     - format: tar.gz
       format_overrides:
         - goos: windows
           format: zip
   ```
   - Cross-platform builds (Linux, macOS, Windows - amd64, arm64)
   - Package manager integrations (Homebrew, Chocolatey)
   - Docker container image builds

2. **GitHub Actions Release Workflow** (`.github/workflows/release.yml`):
   - Trigger on version tags
   - Run goreleaser with `release --clean`
   - Upload artifacts to GitHub Releases
   - Create Homebrew formula automatically

3. **Distribution Channels**:
   - **GitHub Releases**: Main distribution channel
   - **Self-Built**: `make build` for local development
   - **Self-Hosted**: Binary works when placed in any directory
   - **Homebrew**: `brew install specledger` for macOS
   - **npm/npx**: `npx @specledger/cli` for JS users

**Rationale**:
- GoReleaser is the industry standard for Go binary distribution
- Automates cross-platform builds, packaging, and releases
- Integrated with GitHub Actions for CI/CD
- Supports multiple package managers out of the box
- Well-documented and widely adopted

**Alternatives Considered**:
- **Manual make targets**: Too error-prone, doesn't handle all platforms well
- **Release scripts**: Custom solutions require more maintenance
- **Binaries only**: Missing package manager integrations

**Impact**: High - this enables the GitHub Releases and package manager features (User Stories 2 and 6).

---

## Decision 4: Dependency Handling Pattern

**Decision**: Implement dependency registry with multiple fallback levels and helpful error messages.

**Registry Pattern**:
```go
type DependencyRegistry struct {
    gum    *gumClient
    mise   *miseClient
    local  *localResolver
}

func (r *DependencyRegistry) Resolve(name string) (Dependency, error) {
    // Level 1: Check installed locally
    if dep := r.tryLocal(name); dep != nil {
        return dep, nil
    }

    // Level 2: Check via mise
    if dep := r.tryMise(name); dep != nil {
        return dep, nil
    }

    // Level 3: Provide error with installation instructions
    return nil, fmt.Errorf("dependency %s not found. Install with: %s", name, getInstallCommand(name))
}
```

**Fallback Levels**:
1. **Check installed**: Verify `gum` and `mise` are in PATH
2. **Interactive prompt**: Ask user if they want to install
3. **Download fallback**: Optionally download temporary binaries
4. **Plain CLI**: Provide alternative without TUI
5. **Error with instructions**: Clear error message

**Rationale**:
- Provides good UX (stops at interactive prompt instead of failing)
- Follows user's clarification (interactive fallback)
- Clear error messages prevent silent failures
- Consistent with the "no silent failures" success criterion

**Alternatives Considered**:
- **Fail fast**: Exit immediately - too harsh for end users
- **Auto-install**: Download automatically - security concern, user preference unclear
- **No fallback**: Just error - poor UX for our target audience

**Impact**: Medium - affects all bootstrap and TUI workflows.

---

## Decision 5: Command Structure Design

**Decision**: Unified command structure with `sl` as primary name and `specledger` alias.

**Command Hierarchy**:
```
sl [COMMAND] [OPTIONS]

Commands:
  new / bootstrap        Start interactive TUI for project bootstrap
  deps [SUBCOMMAND]      Dependency management
  refs [SUBCOMMAND]      Reference validation
  graph [SUBCOMMAND]     Graph visualization
  vendor [SUBCOMMAND]    Vendor management
  conflict [SUBCOMMAND]  Conflict resolution
  update [SUBCOMMAND]    Update dependencies
  --help                 Show help
  --version              Show version
```

**Flags**:
- `--help` / `-h`: Show help
- `--version` / `-v`: Show version
- `--ci`: Force non-interactive mode
- `--simple`: Force plain CLI mode (no TUI)
- `--project-name`: Set project name (non-interactive bootstrap)
- `--short-code`: Set short code (non-interactive bootstrap)
- `--dry-run`: Preview changes without executing

**Alias Support**:
- `specledger` commands work as aliases
- `specledger new` = `sl new`
- `specledger deps list` = `sl deps list`

**Rationale**:
- `sl` is shorter and more intuitive for bootstrap operations
- `specledger` alias for backward compatibility with existing docs
- Hierarchical structure matches existing Go CLI patterns
- Clear separation between project bootstrap and dependency management

**Alternatives Considered**:
- **Keep separate**: `sl` for bootstrap, `specledger` for everything - doesn't meet unification goal
- **Just sl**: Drop `specledger` alias - breaks backward compatibility
- **Multiple entry points**: Separate binaries - more complex deployment

**Impact**: Medium - defines the core user experience of the unified CLI.

---

## Decision 6: Exit Code Strategy

**Decision**: Use standard exit codes 0 (success) and 1 (any failure).

**Rationale**:
- Simple, standard approach
- Consistent with user clarification
- All failures result in non-zero exit code
- No need for granular error codes (keeps code simpler)

**Exit Code Table**:
| Scenario | Exit Code |
|----------|-----------|
| Success | 0 |
| Any error | 1 |
| Missing required flags | 1 |
| Invalid command | 1 |
| Dependency not found | 1 |
| Permission denied | 1 |
| User cancellation (Ctrl+C) | 130 (handled separately) |

**Alternatives Considered**:
- **Semantic codes**: 0=success, 1=general, 2=usage, 3=deps - adds complexity, unclear benefit
- **Custom codes**: Project-specific codes - harder for users to interpret

**Impact**: Low - implementation detail with minimal user-facing impact.

---

## Decision 7: Observability Strategy

**Decision**: Debug-level logging to stderr, no structured metrics.

**Rationale**:
- Follows user clarification (debug-level logging only)
- Simple, effective for troubleshooting
- No overhead of metrics collection
- Stderr is standard for tooling

**Logging Strategy**:
- **Level**: Debug only (no info/warn/error by default)
- **Output**: stderr (standard for CLIs)
- **Format**: Simple text, human-readable
- **Content**: Decision points, errors, important state changes

**Example**:
```
2026/01/30 10:30:45 [DEBUG] Starting bootstrap with project name "myproject"
2026/01/30 10:30:45 [DEBUG] Checking for gum dependency... not found
2026/01/30 10:30:45 [DEBUG] Asking user about dependency installation
2026/01/30 10:30:50 [DEBUG] User chose to install gum
2026/01/30 10:30:50 [INFO] Installing gum...
```

**Alternatives Considered**:
- **All levels**: Info/warn/error + debug - adds overhead
- **Structured JSON**: Machine-readable - unnecessary for CLI tool
- **No logging**: Harder to debug issues

**Impact**: Low - implementation detail, not user-facing.

---

## Research Summary

All research decisions have been made based on industry best practices, existing project context, and user clarifications. The decisions are:

1. **Cobra framework** - Already in use, perfect for this use case
2. **Bubble Tea for TUI** - Industry standard, integrates well with Cobra
3. **GoReleaser for distribution** - Industry standard, automates everything
4. **Dependency registry pattern** - Provides good UX with multiple fallbacks
5. **Unified command structure** - Meets unification goal with backward compatibility
6. **Simple exit codes** - Standard practice, matches user preference
7. **Debug logging only** - Simple, effective for troubleshooting

No NEEDS CLARIFICATION markers remain. All research complete.
