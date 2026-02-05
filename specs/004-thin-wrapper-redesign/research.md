# Research & Decisions: SpecLedger Thin Wrapper Architecture

**Date**: 2026-02-05
**Feature**: 004-thin-wrapper-redesign

## Prior Work

### Beads Issues Query Results

```bash
# No beads issues found related to "thin wrapper", "framework", "bootstrap", or "mise"
# This is a new architectural direction for the project
```

### Related Git Branches

- `003-cli-unification`: Merged the standalone `sl` bash script and `specledger` binary into a single Go binary. This redesign builds on that unified CLI foundation.
- `002-spec-dependency-linking`: Established the `.mod` file format for dependency tracking. This redesign replaces `.mod` with YAML.

## Research Findings

### 1. YAML Schema Design for Go CLIs

**Research Summary**:
- Surveyed popular Go CLI tools: kubectl (Kubernetes), helm, docker-compose, GitHub CLI
- Common pattern: Use Go structs with YAML tags for marshaling/unmarshaling
- Validation approaches: struct tags (most common), JSON Schema validators (overkill for simple schemas)

**Decision**: Use Go structs with `yaml` tags and custom validation methods
**Rationale**:
- Native Go approach, no external dependencies for validation
- Clear type safety at compile time
- Easy to extend with custom validation logic
- gopkg.in/yaml.v3 is mature and widely used

**Alternatives Considered**:
- JSON Schema validation: Too heavy, adds complexity
- TOML format: Less common for configuration, harder to parse nested structures
- Keep .mod format: No structured parsing, hard to extend

**Implementation Pattern**:
```go
type ProjectMetadata struct {
    Version  string        `yaml:"version"`
    Project  ProjectInfo   `yaml:"project"`
    Framework FrameworkInfo `yaml:"framework"`
    Dependencies []Dependency `yaml:"dependencies,omitempty"`
}

func (m *ProjectMetadata) Validate() error {
    // Custom validation logic
}
```

### 2. Prerequisite Checking Patterns

**Research Summary**:
- Analyzed: kubectl (checks k8s cluster), terraform (checks providers), aws-cli (checks credentials)
- Common patterns: Check via `--version` flag, use `exec.Command().Run()`, clear error messages with actionable steps

**Decision**: Interactive prompts in TUI mode, auto-install in CI mode
**Rationale**:
- User-friendly: Asks permission before running installers
- CI-friendly: Non-blocking in automated environments
- Clear error messages with copy-paste install commands

**Alternatives Considered**:
- Silent auto-install: Too aggressive, users may not want automatic installation
- Fail-only approach: Frustrating for new users, slows onboarding
- Manual-only: Adds friction, defeats purpose of orchestration tool

**Implementation Pattern**:
```go
func EnsurePrerequisites(interactive bool) error {
    missing := checkMissingTools()
    if len(missing) == 0 {
        return nil
    }

    if interactive {
        fmt.Printf("Missing tools: %s\n", strings.Join(missing, ", "))
        if promptYesNo("Install now via mise?") {
            return runMiseInstall()
        }
        return errors.New("prerequisites not met")
    }

    // CI mode: auto-install
    return runMiseInstall()
}
```

### 3. Framework Detection Strategy

**Research Summary**:
- Options: Check PATH for binaries, parse mise.toml, query mise itself
- Trade-offs: PATH check is fast but unreliable (symlinks, multiple versions), mise query is authoritative but slower

**Decision**: Use `command -v <tool>` for detection, rely on PATH (don't copy commands)
**Rationale**:
- Frameworks (Spec Kit, OpenSpec) manage their own Claude commands
- SpecLedger shouldn't duplicate or modify framework-provided commands
- Users install frameworks via mise, which handles PATH automatically
- Simpler architecture: SpecLedger doesn't need to know about framework internals

**Alternatives Considered**:
- Copy framework commands to `.claude/commands/`: Creates duplication, versioning issues
- Parse mise.toml: Doesn't reflect actual installation state (user may have installed manually)
- Query mise programmatically: Adds complexity, mise CLI is sufficient

**Implementation Pattern**:
```go
func DetectFrameworks() []string {
    frameworks := []string{}
    if isCommandAvailable("specify") {
        frameworks = append(frameworks, "speckit")
    }
    if isCommandAvailable("openspec") {
        frameworks = append(frameworks, "openspec")
    }
    return frameworks
}

func isCommandAvailable(cmd string) bool {
    _, err := exec.LookPath(cmd)
    return err == nil
}
```

### 4. Backward Compatibility Strategy

**Research Summary**:
- Studied: Python 2→3 migration, Angular version upgrades, Kubernetes API deprecations
- Best practices: Provide migration tooling, support both formats during transition, clear deprecation warnings

**Decision**: Explicit `sl migrate` command + automatic detection with warning
**Rationale**:
- Users control when migration happens
- Automatic detection prevents silent failures
- Clear migration path documented

**Migration Timeline**:
- Version 1.x: Support both `.mod` and `.yaml`, warn on `.mod` usage
- Version 2.0: Deprecate `.mod`, auto-migrate on first run with confirmation
- Version 3.0: Remove `.mod` support entirely

**Alternatives Considered**:
- Silent auto-migration: Risky, users may not expect file changes
- Hard cutover: Too disruptive, alienates existing users
- Support both forever: Maintenance burden, confusing for new users

**Implementation Pattern**:
```go
func LoadMetadata(dir string) (*ProjectMetadata, error) {
    yamlPath := filepath.Join(dir, "specledger", "specledger.yaml")
    modPath := filepath.Join(dir, "specledger", "specledger.mod")

    if fileExists(yamlPath) {
        return loadYAML(yamlPath)
    }

    if fileExists(modPath) {
        fmt.Println("⚠️  WARNING: .mod format is deprecated. Run 'sl migrate' to convert to YAML.")
        return loadMod(modPath)
    }

    return nil, errors.New("no metadata file found")
}
```

### 5. mise Integration Patterns

**Research Summary**:
- mise documentation: Supports ubi (GitHub binaries), pipx (Python), npm (Node.js), cargo (Rust)
- Best practice: Use comments in mise.toml to guide users, don't validate syntax (mise itself does this)

**Decision**: Commented-out framework options in template, no syntax validation
**Rationale**:
- Users uncomment what they want (explicit opt-in)
- mise validates syntax itself (don't duplicate validation)
- Clear inline documentation in comments

**Template Pattern**:
```toml
[tools]
# Core tools (required)
node = "22"
"ubi:steveyegge/beads" = { version = "0.28.0", exe = "bd" }
"ubi:zjrosen/perles" = "0.2.11"

# SDD Frameworks (optional - uncomment to enable)
# Uncomment ONE or BOTH frameworks below, then run: mise install

# GitHub Spec Kit (Python-based, structured phases)
# "pipx:git+https://github.com/github/spec-kit.git" = "latest"

# OpenSpec (Node-based, lightweight iteration)
# "npm:@fission-ai/openspec" = "latest"
```

**Alternatives Considered**:
- Interactive prompts during bootstrap: Adds friction, users may want to decide later
- Separate config file: Unnecessary, mise.toml is the source of truth
- Validate mise.toml syntax: Duplicates mise's responsibility, adds complexity

## Technology Decisions

### YAML Library
**Chosen**: `gopkg.in/yaml.v3`
**Reasoning**: Most mature Go YAML library, supports anchors/aliases, good error messages

### CLI Framework
**Chosen**: cobra (existing)
**Reasoning**: Already in use, no need to change

### Testing Strategy
**Chosen**: Go standard `testing` package + temp directories for integration tests
**Reasoning**: No need for heavy test frameworks, standard library is sufficient

## Migration Strategy

### Phase 1: Dual Support (v1.x)
- Read both .mod and .yaml
- Write only .yaml for new projects
- Warn users when .mod is detected
- Provide `sl migrate` command

### Phase 2: Auto-Migration (v2.0)
- On first run with .mod, offer to auto-migrate
- Backup original .mod file
- Continue to support reading .mod (read-only)

### Phase 3: Deprecation (v3.0)
- Remove .mod parsing code
- Fail with helpful message if .mod found

## Open Questions

None - all research complete.

## Approved Decisions

✅ YAML format with Go structs and custom validation
✅ Interactive prerequisite prompts (TUI) + auto-install (CI)
✅ Framework detection via PATH (don't copy commands)
✅ Explicit `sl migrate` command with automatic detection/warning
✅ Commented mise.toml template (no syntax validation)
