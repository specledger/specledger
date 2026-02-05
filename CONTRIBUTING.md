# Contributing to SpecLedger

First off, thank you for considering contributing to SpecLedger! It's people like you that make SpecLedger such a great tool.

## How Can I Contribute?

### Reporting Bugs

Before creating bug reports, please check the existing issues as you might find that the problem has already been reported. When creating a bug report, please include:

- **Title**: A clear and descriptive title
- **Description**: A detailed description of the problem
- **Steps to Reproduce**: Step-by-step instructions to reproduce the issue
- **Expected Behavior**: What you expected to happen
- **Actual Behavior**: What actually happened
- **Environment**: 
  - OS and version
  - SpecLedger version (`sl --version`)
  - Go version (if building from source)
- **Screenshots/Logs**: If applicable

### Suggesting Enhancements

Enhancement suggestions are tracked as GitHub issues. When creating an enhancement suggestion, include:

- **Use Case**: What problem does this solve?
- **Proposed Solution**: How would you like it to work?
- **Alternatives**: What other solutions have you considered?
- **Examples**: Mockups, code examples, or documentation

### Pull Requests

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes with clear, descriptive commit messages
4. Add or update tests if applicable
5. Ensure the code passes all tests (`make test`)
6. Format your code (`make fmt`)
7. Commit your changes (`git commit -m 'Add some amazing feature'`)
8. Push to the branch (`git push origin feature/amazing-feature`)
9. Open a Pull Request

## Development Setup

### Prerequisites

- Go 1.21 or higher
- Make
- Git

### Building

```bash
# Clone the repository
git clone https://github.com/specledger/specledger.git
cd specledger

# Build
make build

# The binary will be at bin/sl
```

### Running Tests

```bash
# Run all tests
make test

# Run with coverage
go test -v -cover ./...
```

### Code Style

- Follow standard Go conventions
- Use `gofmt` for formatting
- Run `go vet` before committing
- Write clear comments for exported functions

## Project Structure

```
specledger/
├── cmd/                 # CLI entry point
├── pkg/
│   ├── cli/            # CLI commands and utilities
│   ├── models/         # Data models
│   └── embedded/       # Embedded templates
├── internal/           # Internal packages
├── scripts/            # Installation and utility scripts
└── templates/          # Project templates (embedded)
```

## Commit Messages

Follow conventional commits format:

- `feat:` New feature
- `fix:` Bug fix
- `docs:` Documentation changes
- `style:` Code style changes (formatting, etc.)
- `refactor:` Code refactoring
- `test:` Adding or updating tests
- `chore:` Build process or auxiliary tool changes

Example:
```
feat: add interactive mode for dependency selection

This allows users to interactively select which dependencies to resolve
when running `sl deps resolve --interactive`.
```

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

## Questions?

Feel free to open an issue with the `question` label, or start a discussion on GitHub Discussions.
