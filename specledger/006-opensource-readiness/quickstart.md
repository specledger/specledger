# Quickstart: Open Source Readiness

**Feature**: 006-opensource-readiness | **Date**: 2026-02-09

## Overview

This quickstart guide covers the implementation of the Open Source Readiness feature for SpecLedger. Most infrastructure is already in place from previous features, so this feature primarily involves adding missing documentation, legal files, and CI/CD quality checks.

## Prerequisites

- Go 1.24+ installed
- Git configured with GitHub credentials
- Access to `github.com/specledger/specledger` repository
- Access to `github.com/specledger/homebrew-specledger` repository
- Domain access to `specledger.io` (for documentation)

## What's Already Done

Before starting, verify these existing files are present:

```bash
# Verify existing legal files
ls -la LICENSE CODE_OF_CONDUCT.md SECURITY.md CONTRIBUTING.md

# Verify existing CI/CD
ls -la .github/workflows/ .goreleaser.yaml

# Verify existing documentation
ls -la README.md CHANGELOG.md
```

## Implementation Steps

### Step 1: Add NOTICE File (P1)

Create a NOTICE file at repository root for transparency about third-party dependencies:

```bash
cat > NOTICE << 'EOF'
SpecLedger
Copyright 2025 SpecLedger Contributors

This project includes third-party software:

- Cobra (Apache 2.0) - https://github.com/spf13/cobra
- Bubble Tea (MIT) - https://github.com/charmbracelet/bubbletea
- go-git (Apache 2.0) - https://github.com/go-git/go-git
- YAML v3 (Apache 2.0) - https://github.com/go-yaml/yaml

See full dependency list in go.mod and go.sum.
EOF
```

### Step 2: Create GOVERNANCE.md (P1)

Create project governance documentation:

```bash
cat > GOVERNANCE.md << 'EOF'
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
EOF
```

### Step 3: Add CI Quality Workflow (P1)

Create a CI workflow for quality checks:

```bash
mkdir -p .github/workflows

cat > .github/workflows/ci.yml << 'EOF'
name: CI

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.24'

      - name: Download dependencies
        run: go mod download

      - name: Run tests
        run: make test

      - name: Run coverage
        run: make test-coverage

      - name: Upload coverage
        uses: codecov/codecov-action@v4
        with:
          file: ./coverage.out

  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.24'

      - name: Run golangci-lint
        uses: golangci/golangci-lint-action@v6
        with:
          version: latest

  format:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.24'

      - name: Check formatting
        run: |
          if [ -n "$(gofmt -l .)" ]; then
            echo "Go code is not formatted:"
            gofmt -d .
            exit 1
          fi
EOF
```

### Step 4: Add golangci-lint Configuration (P1)

Create golangci-lint configuration:

```bash
cat > .golangci.yml << 'EOF'
linters:
  enable:
    - gofmt
    - govet
    - staticcheck
    - errcheck
    - gosimple
    - ineffassign
    - unused
    - gosec

linters-settings:
  govet:
    check-shadowing: true
  gofmt:
    simplify: true

run:
  timeout: 5m
  tests: true
EOF
```

### Step 5: Update README with Badges (P1)

Add badges to the top of README.md:

```markdown
[![Build Status](https://img.shields.io/github/actions/workflow/status/specledger/specledger/ci.yml?branch=main)](https://github.com/specledger/specledger/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/specledger/specledger)](https://goreportcard.com/report/github.com/specledger/specledger)
[![Coverage](https://img.shields.io/codecov/c/github/specledger/specledger)](https://codecov.io/gh/specledger/specledger)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Version](https://img.shields.io/github/v/release/specledger/specledger)](https://github.com/specledger/specledger/releases)

# SpecLedger
...
```

### Step 6: Create Documentation Structure (P2)

Create the docs directory structure:

```bash
mkdir -p docs/{user,contributor,governance}

# Create index
cat > docs/index.md << 'EOF'
# SpecLedger Documentation

Welcome to the SpecLedger documentation.

## User Guide

- [Getting Started](user/getting-started.md)
- [Commands](user/commands.md)
- [SDD Framework](user/sdd-framework.md)
- [Troubleshooting](user/troubleshooting.md)

## Contributor Guide

- [Setup](contributor/setup.md)
- [Architecture](contributor/architecture.md)
- [Development](contributor/development.md)
- [Testing](contributor/testing.md)

## Governance

- [Roadmap](governance/roadmap.md)
- [Decision Making](governance/decision-making.md)
- [Maintenance](governance/maintenance.md)
EOF

# Copy relevant sections from README to docs
# (This is a simplified example - actual content migration would be more detailed)
```

### Step 7: Update GoReleaser Configuration (P1)

Ensure .goreleaser.yaml includes all necessary configuration:

```bash
# Verify .goreleaser.yaml has:
# - Build targets for linux, darwin, windows (amd64, arm64)
# - Homebrew tap configuration
# - Checksum generation
# - GitHub release creation
```

### Step 8: Setup Codecov (P2)

1. Sign up at https://codecov.io
2. Add the `specledger/specledger` repository
3. The CI workflow already includes codecov upload

## Verification

After implementing all steps, verify:

```bash
# 1. Check legal files exist
ls -la LICENSE NOTICE CODE_OF_CONDUCT.md SECURITY.md CONTRIBUTING.md GOVERNANCE.md

# 2. Verify CI workflow
cat .github/workflows/ci.yml

# 3. Verify badges in README
head -20 README.md

# 4. Run local tests
make test

# 5. Run linting
golangci-lint run

# 6. Check formatting
gofmt -l .

# 7. Verify documentation structure
ls -la docs/
```

## Deployment

### Documentation Deployment

For deploying to specledger.io/docs, you have several options:

**Option 1: GitHub Pages**
- Create `docs/.gitignore` with appropriate exclusions
- Use a static site generator (Hugo, Jekyll)
- Deploy via GitHub Actions

**Option 2: Separate repository**
- Create `specledger/documentation` repository
- Deploy via separate CI/CD pipeline

**Option 3: CDN deployment**
- Build static site
- Deploy to CDN (Cloudflare Pages, Netlify, Vercel)

## Testing the Release Process

Before the actual release, test the process:

```bash
# 1. Create a test tag
git tag -a v0.0.0-test -m "Test release"

# 2. Run GoReleaser in dry-run mode
goreleaser release --clean --skip-publish --skip-sign

# 3. Verify artifacts are created correctly
ls -la dist/

# 4. Clean up test tag
git tag -d v0.0.0-test
```

## Rollback Plan

If issues arise after release:

1. **Homebrew**: Revert formula in homebrew-specledger tap
2. **GitHub Releases**: Yank the release, create new patch version
3. **Documentation**: Revert documentation deployment

## Success Criteria

- [ ] All legal files present and complete
- [ ] CI workflow passes on all branches
- [ ] Badges display correctly in README
- [ ] Documentation is accessible at specledger.io/docs
- [ ] GoReleaser builds successfully
- [ ] Homebrew tap is functional
- [ ] Coverage reports are visible on Codecov

## Next Steps

After implementing this feature:

1. Run `/specledger.tasks` to generate implementation tasks
2. Create GitHub issues for tracking
3. Assign priority and effort estimates
4. Begin implementation following the task order
