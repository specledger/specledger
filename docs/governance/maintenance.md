# Maintenance Policy

Release schedule, support commitments, and maintenance practices.

## Release Schedule

### Version Policy

SpecLedger follows Semantic Versioning (SemVer):
- **Major version (X.0.0)**: Breaking changes
- **Minor version (0.X.0)**: New features, backward compatible
- **Patch version (0.0.X)**: Bug fixes, backward compatible

### Release Cadence

- **Patch releases**: As needed for critical fixes
- **Minor releases**: Monthly or when feature set warrants
- **Major releases**: Quarterly or when breaking changes accumulate

### Release Process

1. Version bump in `go.mod`
2. Update `CHANGELOG.md` with changes
3. Create git tag: `git tag -a v0.2.0 -m "Release v0.2.0"`
4. Push tag: `git push origin v0.2.0`
5. GoReleaser automatically:
   - Builds binaries for all platforms
   - Creates GitHub release
   - Updates Homebrew formula

## Support Commitments

### Supported Versions

| Version | Support Level | End of Life |
|---------|---------------|-------------|
| 0.2.x | Full support | 3 months after next minor release |
| 0.1.x | Maintenance only | When 0.3.0 releases |
| < 0.1 | Best effort | No active support |

### Support Scope

**Full Support**:
- Bug fixes
- Security patches
- Feature requests (within scope)
- Documentation updates

**Maintenance Only**:
- Critical bug fixes
- Security patches
- No new features

### Response Times

| Issue Type | Response Time |
|------------|---------------|
| Critical security | 48 hours |
| High priority bugs | 1 week |
| Medium priority bugs | 2 weeks |
| Low priority | Best effort |
| Feature requests | Roadmap consideration |

## Security Policy

### Vulnerability Reporting

Report vulnerabilities privately per [SECURITY.md](../SECURITY.md):
1. Do not use public issues
2. Send details to maintainers
3. Allow 90 days for fix before public disclosure

### Security Updates

- Critical vulnerabilities: Patch within 7 days
- High severity: Patch within 30 days
- Medium/Low severity: Next minor version or as needed

### Dependency Scanning

- GitHub Dependabot enabled
- Monthly security reviews
- govulncheck integrated in CI
- Vulnerabilities addressed promptly

## Quality Standards

### Code Quality

- All PRs must pass CI (tests, linting, formatting)
- Code coverage target: 70%+
- golangci-lint with 8 linters enabled
- No high-severity vulnerabilities

### Documentation Standards

- All features documented in specledger.io/docs
- Changelog updated for each release
- README reflects current state
- API docs for code libraries (if applicable)

### Performance Standards

- CLI commands complete in <5 seconds
- Dependency resolution completes in <30 seconds
- Project creation completes in <2 minutes

## Breaking Changes

### Deprecation Process

1. Feature marked as deprecated in documentation
2. Deprecation period: 3 months minimum
3. Removal announcement in changelog
4. Removal in next minor version

### Migration Guide

For breaking changes, provide:
- What changed and why
- Migration steps
- Code examples if applicable
- Rollback options if migration fails

## Backwards Compatibility

### What We Guarantee

- CLI command signatures (non-experimental flags)
- YAML configuration schema
- Dependency metadata format
- Plugin/extension APIs

### What May Change

- Experimental features (marked as such)
- Internal implementation details
- Output formats (unless documented)
- TUI behavior (unless documented)

## Maintenance Tasks

### Regular Maintenance

- **Weekly**: Review triage backlog
- **Monthly**: Security dependency updates
- **Quarterly**: Release planning and roadmap review
- **As needed**: Bug fixes and patches

### Monitoring

- GitHub Issues for bugs and feature requests
- GitHub Discussions for questions
- CI/CD health monitoring
- Dependency vulnerability scanning

## Contributing to Maintenance

### Becoming a Maintainer

1. Active contributions over 6 months
2. Successful completion of 5+ non-trivial PRs
3. Understanding of codebase and architecture
4. Agreement with governance principles

### Maintainer Responsibilities

- Review proposals and PRs
- Release management
- Security vulnerability handling
- Community support
- Documentation maintenance

## Emeritus Maintainers

Maintainers who step down may become emeritus:
- Granted lifetime access to private discussions
- Consulted on historical context
- No binding vote (unless active maintainer)

## Contact

For maintenance-related questions:
- **Issues**: [GitHub Issues](https://github.com/specledger/specledger/issues)
- **Security**: [SECURITY.md](../SECURITY.md)
- **Governance**: [GOVERNANCE.md](../GOVERNANCE.md)
