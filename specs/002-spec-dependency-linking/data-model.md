# Data Model: Spec Dependency Linking

**Generated**: 2026-01-30
**Branch**: `002-spec-dependency-linking`

## Entity Overview

The spec dependency linking system manages external specification dependencies with version control, validation, and resolution capabilities. The following entities define the core data structures and relationships.

## Core Entities

### 1. SpecManifest

**Description**: Represents the dependency manifest file (`spec.mod`) that declares external specification dependencies.

**Fields**:
```go
type SpecManifest struct {
    Version     string           // Version of the manifest format
    Dependencies []Dependency   // List of declared dependencies
    ID         string           // Unique identifier for this spec (for external references)
    Path       string           // File path to this spec (relative to repo root)
    CreatedAt  time.Time        // When the manifest was created
    UpdatedAt  time.Time        // When the manifest was last updated
}
```

**Validation Rules**:
- Version must be semver-compatible (e.g., "1.0.0")
- Dependencies array cannot be empty if dependencies exist
- ID must be unique within the repository
- Path must be relative and not contain ".."
- CreatedAt and UpdatedAt must be valid timestamps

### 2. Dependency

**Description**: Represents a single external specification dependency declaration.

**Fields**:
```go
type Dependency struct {
    RepositoryURL string   // Git repository URL (HTTPS or SSH)
    Version      string   // Version constraint (branch, tag, commit hash, or semver)
    SpecPath     string   // Path to the specification file within the repository
    Alias        string   // Optional short alias for the dependency
    Pinned       bool     // Whether the dependency is pinned to a specific commit
    Transitive   []Dependency // Transitive dependencies discovered from external spec.mod
}
```

**Validation Rules**:
- RepositoryURL must be a valid Git URL (GitHub, GitLab, etc.)
- Version must be a valid branch name, tag, or commit hash (40 chars)
- SpecPath must be a relative path without ".."
- Alias must be alphanumeric with hyphens/underscores if provided
- Pinned is true when Version is a commit hash

### 3. SpecLockfile

**Description**: Represents the lockfile (`spec.sum`) that contains resolved dependency versions with cryptographic verification.

**Fields**:
```go
type SpecLockfile struct {
    Version    string             // Version of the lockfile format
    Entries    []LockfileEntry   // Resolved dependency entries
    Hashes     map[string]string // Additional content hashes for verification
    Timestamp  time.Time         // When the lockfile was generated
    Hash       string            // SHA-256 hash of the entire lockfile
}
```

**Validation Rules**:
- Version must match manifest version
- Entries must have unique RepositoryURL + SpecPath combinations
- Hashes must contain valid SHA-256 hashes
- Timestamp must be recent (< 24 hours old for auto-update checks)
- Hash must verify against the lockfile content

### 4. LockfileEntry

**Description**: Represents a single resolved dependency entry in the lockfile.

**Fields**:
```go
type LockfileEntry struct {
    RepositoryURL  string    // Repository URL
    CommitHash     string    // Git commit hash of the resolved version
    ContentHash    string    // SHA-256 hash of the spec file content
    SpecPath       string    // Path to the spec file
    Branch         string    // Resolved branch name (if applicable)
    Size           int64     // Size of the spec file in bytes
    FetchedAt      time.Time // When the entry was last fetched
}
```

**Validation Rules**:
- CommitHash must be a valid 40-character SHA-1 hash
- ContentHash must be a valid SHA-256 hash (64 chars)
- SpecPath must match the dependency declaration
- Size must be > 0 and reasonable (< 10MB per spec)
- FetchedAt must be within the cache expiration period

### 5. ExternalReference

**Description**: Represents a link from the current spec to a section in an external specification.

**Fields**:
```go
type ExternalReference struct {
    SourceLocation string    // Line location in the current spec (file:line)
    TargetSpecID   string    // Unique ID of the target external spec
    TargetSectionID string   // Section ID in the external spec
    DisplayText    string    // Text to display for the reference
    Validated      bool      // Whether the reference has been validated
    ValidationTime time.Time // When the reference was last validated
}
```

**Validation Rules**:
- SourceLocation must be in "file:line" format
- TargetSpecID must match an ID from an external manifest
- TargetSectionID must exist in the referenced spec
- DisplayText must not be empty
- Validated must be true before the spec can be used for planning

### 6. DependencyGraph

**Description**: Represents the complete dependency graph including transitive dependencies.

**Fields**:
```go
type DependencyGraph struct {
    Root      Dependency        // The root specification (this repo)
    Nodes     []GraphNode      // All dependency nodes
    Edges     []GraphEdge       // Dependencies between nodes
    Conflicts []Conflict       // Detected version conflicts
    Depth     int              // Maximum depth of dependency tree
    TotalSize int64            // Total size of all specs in KB
}
```

**Validation Rules**:
- Root must have no parent dependencies
- Nodes must have unique RepositoryURL + SpecPath combinations
- Edges must reference existing nodes
- Conflicts must have severity levels (error, warning, info)
- Depth must be within reasonable limits (< 20 for performance)

### 7. GraphNode

**Description**: Represents a single node in the dependency graph.

**Fields**:
```go
type GraphNode struct {
    ID          string        // Unique node identifier
    Dependency  Dependency    // The dependency this node represents
    Level       int           // Level in the dependency tree (0 = root)
    Children    []string      // IDs of child nodes
    Size        int64         // Size of this spec file
    State       NodeState     // Resolved, Missing, Conflicted
    CacheHit    bool          // Whether this was a cache hit
}
```

**Validation Rules**:
- ID must be unique across the graph
- Level must be non-negative
- Children must reference existing node IDs
- State must be one of: Resolved, Missing, Conflicted
- Size must be accurate (0 for missing nodes)

### 8. GraphEdge

**Description**: Represents a dependency relationship between two nodes.

**Fields**:
```go
type GraphEdge struct {
    From      string        // Source node ID
    To        string        // Target node ID
    Type      EdgeType      // Direct, Transitive, Conflict
    Weight    int           // Priority or conflict severity
    Metadata  map[string]string // Additional context
}
```

**Validation Rules**:
- From and To must reference existing node IDs
- Type must be one of: Direct, Transitive, Conflict
- Weight must be appropriate for the type
- Metadata must contain relevant context

### 9. Conflict

**Description**: Represents a dependency conflict that needs resolution.

**Fields**:
```go
type Conflict struct {
    Type       ConflictType    // Version, Circular, Missing
    Severity   ConflictSeverity // Error, Warning, Info
    Description string         // Human-readable description
    Affects    []string        // List of affected node IDs
    Suggestions []Suggestion   // Possible resolutions
    Resolved   bool            // Whether the conflict has been resolved
}
```

**Validation Rules**:
- Type must be one of: Version, Circular, Missing
- Severity must match the conflict type
- Description must be clear and actionable
- Affects must list all impacted nodes
- Suggestions must be viable resolutions

### 10. VendoredSpec

**Description**: Represents a local copy of an external specification stored in the vendor directory.

**Fields**:
```go
type VendoredSpec struct {
    OriginalURL     string    // Original repository URL
    CommitHash      string    // Git commit hash
    LocalPath       string    // Path in specs/vendor/
    ContentHash     string    // SHA-256 hash of content
    VendorPath      string    // Path within vendor directory
    CopiedAt        time.Time // When copied to vendor
    Size            int64     // Size of the file
    UpToDate        bool      // Whether it matches the remote version
}
```

**Validation Rules**:
- OriginalURL must be valid and accessible
- CommitHash must be valid
- LocalPath must be within specs/vendor/
- ContentHash must match the original
- VendorPath must preserve the original structure
- UpToDate must be checked against remote

## Relationships

### 1. Manifest → Dependencies (One-to-Many)
- A SpecManifest contains multiple Dependency entries
- Each Dependency belongs to exactly one SpecManifest
- Relationship enforced through file structure

### 2. Dependency → Transitive Dependencies (Self-Reference)
- A Dependency can have multiple transitive dependencies
- Transitive dependencies are discovered from external spec.mod files
- Forms a tree structure with potential cycles

### 3. Lockfile → Entries (One-to-Many)
- A SpecLockfile contains multiple LockfileEntry objects
- Each Entry represents a resolved dependency
- Relationship enforced through file structure

### 4. Entry → Dependency (Reference)
- A LockfileEntry corresponds to a Dependency declaration
- The entry contains the resolved version information
- Used for validation and conflict detection

### 5. Reference → External Spec (Many-to-One)
- Multiple ExternalReferences can point to the same external spec
- References are validated against the resolved dependencies
- Forms a web of interconnected specifications

### 6. Graph → Nodes (One-to-Many)
- A DependencyGraph contains multiple GraphNode objects
- Each Node represents a dependency in the graph
- Relationship computed during graph construction

### 7. Graph → Edges (One-to-Many)
- A DependencyGraph contains multiple GraphEdge objects
- Edges define relationships between nodes
- Relationship computed during graph traversal

### 8. Conflict → Graph (Composition)
- Conflicts exist within the context of a DependencyGraph
- Multiple conflicts can exist in a single graph
- Relationship computed during conflict detection

## State Transitions

### 1. Dependency Resolution Flow
```
Declared → Resolving → Resolved → Validated
    ↓         ↓         ↓
  Missing → Error → Conflict
```

### 2. Reference Validation Flow
```
Unvalidated → Validating → Validated → Stale
     ↓            ↓         ↓        ↑
   Invalid   → Error → Refresh
```

### 3. Vendoring Flow
```
External → Copying → Copied → Outdated → Updated
     ↓         ↓        ↓         ↓        ↓
   Error   → Failed → Valid → Valid → Synced
```

## Data Persistence

### File Storage Locations
- **spec.mod**: `specs/<NNN>-feature>/spec.mod`
- **spec.sum**: `specs/<NNN>-feature>/spec.sum`
- **Cache**: `specs/<NNN>-feature>/.spec-cache/`
- **Vendor**: `specs/vendor/`
- **Config**: `.spec-config/auth.json`

### Serialization Formats
- **JSON**: For configuration and metadata
- **Text**: For spec.mod and spec.sum files (human-readable)
- **Binary**: For cached Git objects (go-git internal)

### Validation Hooks
- File parsing validation (syntax checking)
- Cross-reference validation (external references)
- Content integrity validation (SHA-256 verification)
- Graph validation (cycle detection)

## Performance Considerations

### Indexing Strategies
- Repository URL index for fast lookups
- Commit hash index for version tracking
- Content hash index for deduplication
- Graph adjacency list for traversal

### Caching Layers
- Memory cache for frequently accessed specs
- File system cache for resolved dependencies
- Git object cache for repository operations

### Batch Operations
- Parallel resolution of multiple dependencies
- Bulk validation of external references
- Efficient graph algorithms for conflict detection