# Contracts Directory

**Feature**: 006-opensource-readiness

This directory contains contracts defining the requirements for various components of the Open Source Readiness feature.

## Contracts

### [legal-files-contract.md](./legal-files-contract.md)

Defines the required legal files for open source compliance:
- LICENSE (MIT)
- NOTICE (third-party attributions)
- CODE_OF_CONDUCT.md
- SECURITY.md
- CONTRIBUTING.md
- GOVERNANCE.md
- Source file license headers

### [ci-cd-contract.md](./ci-cd-contract.md)

Defines continuous integration and delivery requirements:
- CI workflow (test, lint, format jobs)
- GoReleaser configuration
- Release automation
- README badges
- golangci-lint configuration

## Contract Validation

Each contract includes:
1. Required file locations
2. Expected content structure
3. Validation scripts
4. Acceptance criteria

## Usage

These contracts serve as:
1. **Implementation guides** - What needs to be built
2. **Testing contracts** - How to verify compliance
3. **Documentation** - What the system should do

## No API Contracts

This feature does not define API contracts as it is focused on:
- Documentation
- Legal compliance
- Infrastructure/CI/CD
- Community governance

The actual SpecLedger CLI API contracts are defined in the core feature specifications.
