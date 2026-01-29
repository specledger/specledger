# Research Report: Spec Dependency Linking

**Generated**: 2026-01-30
**Branch**: `002-spec-dependency-linking`

## Prior Work

### 1. **Epic: 001-sdd-control-plane - SpecLedger SDD Control Plane**
- **Status**: Draft (created 2025-12-22)
- **Core Infrastructure**: Specification creation, versioning, and collaboration
- **Key Foundation Features**:
  - Specification Management: Structured format for user stories and requirements
  - Branching and Versioning: Git-based version tracking with commit hash support
  - Multi-User Collaboration: Concurrent access with conflict resolution
  - Audit Trail: Complete history tracking for specification changes

### 2. **Cross-Repository Specification Sharing Foundation**
- **External References**: Existing support for external spec references in FR-011, FR-012
- **Version Tracking**: Infrastructure for pinning specific versions (FR-032)
- **Reference Integrity**: System for maintaining referential integrity (FR-036)

### 3. **Specification Lifecycle Management**
- **Clarification System**: Structured Q&A for requirements refinement
- **Planning Integration**: Link between specifications and implementation
- **Task Generation**: Beads-based task management with dependency tracking

### 4. **Authentication Framework**
- **Private Repository Support**: Token and SSH key authentication (FR-026, FR-028)
- **Security**: Credential management without plaintext storage

### Key Foundation for Extension
The 001 specification provides essential building blocks:
- Specification format structure
- Git-based versioning system
- Markdown link syntax for references
- Authentication framework for private repositories

## Technical Decisions

### Language: Go 1.21+

**Chosen**: Go 1.21+ for CLI implementation

**Why Chosen**:
- **Performance**: Compiled nature meets SC-003 (<30s for 10 repos) and SC-001 (<10s single dep)
- **Binary Distribution**: Single binary without runtime dependencies (aligns with existing `sl` script)
- **Git Integration**: Excellent support through go-git library
- **Memory Efficiency**: Superior for handling multiple repository operations (SC-010 - 50 deps)
- **Existing Infrastructure**: Project already uses Go tools (`bd` beads tool)

**Alternatives Considered**:
- **Node.js**: Strong npm ecosystem but higher memory usage and startup overhead
- **Rust**: Excellent performance but steeper learning curve and larger binaries
- **Python**: Easy development but slower for crypto operations

### Primary Dependencies

**Git Operations**: `github.com/go-git/go-git/v4`
- Pure Go Git implementation
- Supports HTTP(S) and SSH protocols
- No external Git binary dependency
- Built-in authentication for private repos

**CLI Framework**: `github.com/spf13/cobra` + `github.com/spf13/viper`
- Standard Go CLI framework with subcommand support
- Configuration management for auth tokens
- Flag parsing and help system

**Cryptographic**: `golang.org/x/crypto` + `hashicorp/go-getter`
- SHA-256 implementation for spec.sum verification
- URL handling and resource fetching

**Authentication**: `golang.org/x/oauth2` + `golang.org/x/ssh`
- OAuth2 for GitHub/GitLab tokens
- SSH client support for private repositories

### Storage Architecture

**File System Structure**:
```
specs/
├── <NNN>-<feature-name>/
│   ├── spec.md          # Main specification
│   ├── spec.mod         # Dependency manifest
│   ├── spec.sum         # Lockfile with hashes
│   └── .spec-cache/     # Local cache directory
specs/vendor/             # Vendored dependencies (P5 feature)
└── .spec-config/        # Configuration and tokens
    ├── auth.json        # Authentication tokens
    └── cache.db         # Cache metadata (SQLite)
```

**Cache Strategy**:
- 100MB LRU cache limit
- SHA-256 hash as cache key
- 1 hour expiration for cached specs
- Git ignored to avoid accidental commits

### Testing Framework

**Tools**: `go-unit`, `testify`, `httptest`
- 90%+ line coverage for critical paths
- 100% coverage for cryptographic operations
- Performance benchmarks for SC-001 to SC-010 criteria
- Security tests for authentication bypass scenarios

### Target Platform

**Distribution**: GitHub releases, Homebrew tap, APT/Yum repos
**Build Targets**: darwin/amd64, linux/amd64, windows/amd64
**Static Binaries**: No installation required
**Memory Limit**: 512MB max for dependency resolution

## Implementation Architecture

### Core Components
```go
cmd/main.go (cobra CLI)
├── spec/
│   ├── parser.go      # spec.mod parsing
│   ├── resolver.go    # Dependency resolution
│   └── validator.go   # Reference validation
├── git/
│   ├── client.go      # Git client
│   ├── auth.go        # Authentication
│   └── cache.go       # Cache management
├── crypto/
│   ├── hash.go        # SHA-256 calculation
│   └── verify.go      # Content verification
└── config/
    ├── file.go        # File-based config
    └── auth.go        # Token management
```

### Performance Optimizations
- **Parallel Fetching**: Concurrent Git operations for multiple repos
- **Shallow Clones**: `--depth 1` for most operations
- **Stream Processing**: Line-by-line spec parsing
- **Connection Pooling**: Reuse HTTP connections

### Security Considerations
- **Token Encryption**: Encrypted storage for auth tokens
- **Tamper Detection**: SHA-256 verification mandatory before use
- **Certificate Pinning**: Optional for critical repositories
- **Input Validation**: Strict validation of all user inputs

### Integration Points
- **Beads Issue Tracker**: Use `bd` for issue management
- **Claude Skills**: Integrate with `/specledger.*` commands
- **Mise Configuration**: Add tool to `mise.toml`

## Performance Targets (from Specification)

| Metric | Target | Implementation Strategy |
|--------|--------|-------------------------|
| SC-001 | <10s single dependency | Shallow clone, parallel hash calculation |
| SC-002 | <5s reference validation | Pre-resolved cache, indexed lookups |
| SC-003 | <30s for 10 repos | Parallel fetching, connection pooling |
| SC-010 | 50 transitive deps | Efficient graph traversal, caching |
| Memory | <512MB | Stream processing, bounded cache |

## Command Structure

```bash
# Primary commands
specledger deps add <repo-url> [branch] [path]      # Add dependency
specledger deps resolve                             # Resolve dependencies
specledger deps update                              # Update dependencies
specledger deps validate                             # Validate references

# Secondary commands
specledger deps vendor                               # Vendor dependencies
specledger deps graph                                # Show dependency graph
specledger deps clean                               # Clean cache
```

## Risk Mitigation

1. **Performance**: Go's compiled nature and go-git's efficiency ensure SC-003 compliance
2. **Security**: Multiple layers of authentication and content verification
3. **Compatibility**: Single binary deployment follows existing project patterns
4. **Maintainability**: Standard Go project structure with clear separation of concerns
5. **Testing**: Comprehensive test coverage for all critical paths

## Research Sources

1. Go Module Dependency Management Best Practices 2026
2. Node.js CLI Tools vs Go Performance Comparison 2026
3. CLI Architecture Patterns in Go
4. Go-git Library Documentation
5. SpecLedger Existing Architecture Documentation