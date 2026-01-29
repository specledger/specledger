# SpecLedger

Specification dependency management CLI tool for managing external specification dependencies across repositories.

## Features

- **Dependency Declaration**: Add external specification dependencies to your project
- **Dependency Resolution**: Fetch and verify external specifications with cryptographic hashing
- **Reference Validation**: Validate markdown links to external spec sections
- **Dependency Graph**: Visualize and export dependency relationships
- **Vendor Support**: Copy dependencies for offline use

## Project Structure

```
specledger/
├── cmd/
│   └── main.go           # CLI entry point
├── pkg/
│   └── cli/
│       └── commands/
│           ├── deps.go   # Dependency management commands
│           ├── refs.go   # Reference validation commands
│           ├── graph.go  # Graph visualization commands
│           └── vendor.go # Vendoring commands
├── internal/              # Internal logic (to be implemented)
├── tests/                 # Test fixtures and tests (to be implemented)
├── go.mod                 # Go module definition
├── Makefile               # Build and development targets
└── .gitignore             # Git ignore patterns
```

## Installation

### Build from Source

```bash
make build
```

This creates `bin/specledger` binary.

### Run

```bash
./bin/specledger --help
```

## Usage

### Dependency Commands

```bash
# Add a dependency
sl deps add <repo-url> [branch] [spec-path] [--alias <name>]

# List dependencies
sl deps list [--include-transitive]

# Resolve dependencies
sl deps resolve [--no-cache] [--deep]

# Update dependencies
sl deps update [--force] [repo-url]

# Remove a dependency
sl deps remove <repo-url> <spec-path>
```

### Reference Commands

```bash
# Validate references
sl refs validate [--strict] [--spec-path <path>]

# List references
sl refs list
```

### Graph Commands

```bash
# Show dependency graph
sl graph show [--format <format>] [--include-transitive]

# Export graph to file
sl graph export --format <format> --output <file>

# Show transitive dependencies
sl graph transitive [--depth <n>]
```

### Vendor Commands

```bash
# Vendor dependencies
sl vendor --output <path>

# Update vendored dependencies
sl vendor update [--vendor-path <path>] [--force]

# Clean vendored dependencies
sl vendor clean
```

## Development

### Build and Test

```bash
make build        # Build the binary
make test         # Run tests
make test-coverage  # Generate coverage report
make fmt          # Format code
make vet          # Run go vet
```

### Available Platforms

```bash
make build-all    # Build for linux, darwin, windows
```

## Project Status

**Current Phase**: Setup and CLI framework

Implemented:
- ✅ Go project initialization (go.mod)
- ✅ Cobra CLI framework
- ✅ Command structure (deps, refs, graph, vendor)
- ✅ Basic command help and flags
- ✅ .gitignore and Makefile
- ✅ Project structure

To be implemented (see tasks.md):
- 🔨 Dependency declaration and manifest parsing
- 🔨 Dependency resolution with Git integration
- 🔨 Cryptographic hash verification (spec.sum)
- 🔨 Reference validation
- 🔨 Cache management
- 🔨 Authentication framework

## License

See LICENSE file for details.
