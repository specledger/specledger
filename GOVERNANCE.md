# SpecLedger Project Governance

## Project Maintainers

The SpecLedger project is maintained by the SpecLedger core team.

## Decision Making

### Feature Proposals
1. Create a GitHub issue with the `proposal` tag
2. Discuss with the community
3. Maintainers review and approve/reject
4. Approved proposals move to specification phase

### Contribution Review
1. All contributions go through pull requests
2. At least one maintainer must approve
3. CI checks must pass
4. Follows CONTRIBUTING.md guidelines

## Release Process

1. Version bump in go.mod
2. Update CHANGELOG.md
3. Create git tag
4. GoReleaser creates release artifacts
5. Homebrew formula updated automatically

## Security

Security vulnerabilities should be reported privately per SECURITY.md.

## Code of Conduct

All community members must follow CODE_OF_CONDUCT.md.
