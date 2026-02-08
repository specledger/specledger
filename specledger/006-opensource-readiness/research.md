# Research: Open Source Readiness

**Feature**: 006-opensource-readiness | **Date**: 2026-02-09

## Overview

This document consolidates research findings for the Open Source Readiness feature. The SpecLedger project is already significantly prepared for open source release with most infrastructure in place.

## Prior Work

### Existing Open Source Infrastructure

The project already has substantial open source infrastructure from previous features:

- **001-sdd-control-plane**: Core specification-driven development framework
- **002-spec-dependency-linking**: Dependency management between specifications
- **003-cli-integration/unification**: CLI interface and command structure
- **004-thin-wrapper-redesign**: Architecture improvements with Go 1.24+
- **005-embedded-templates**: Template system for project initialization

### Current State Assessment

**Already Complete** (from previous features):
- MIT License file
- CODE_OF_CONDUCT.md
- SECURITY.md
- CONTRIBUTING.md
- GoReleaser configuration (.goreleaser.yaml)
- Homebrew formula (homebrew/specledger.rb)
- GitHub release workflow
- Installation scripts (Unix and Windows)
- Comprehensive README.md
- CHANGELOG.md
- GitHub issue/PR templates

**Needs to be Added** (this feature):
- NOTICE file for third-party attributions
- Quality checks in CI (golangci-lint, formatting)
- Test coverage badge
- Go version update in CI (1.21 → 1.24)
- Dedicated documentation structure for specledger.io/docs
- README badges (build status, version, license, coverage)
- GOVERNANCE.md for project decision-making

## Decisions

### D1: License Choice
**Decision**: MIT License

**Rationale**:
- Already established in the codebase
- Permissive, widely accepted in open source community
- Minimal attribution requirements
- Compatible with most other licenses

**Alternatives Considered**:
- Apache 2.0: More comprehensive patent protection but longer
- BSD 3-Clause: Similar to MIT but with more restrictive endorsement clause
- GPL: Rejected due to copyleft restrictions

### D2: CI/CD Platform
**Decision**: GitHub Actions

**Rationale**:
- Already in use for releases
- Native GitHub integration
- Free for public repositories
- Comprehensive workflow support

**Alternatives Considered**:
- CircleCI: Additional service, not native
- GitLab CI: Requires GitLab hosting
- Travis CI: Less reliable, requires external service

### D3: Linting/Formatting Tool
**Decision**: golangci-lint

**Rationale**:
- Industry standard for Go projects
- Fast, comprehensive, configurable
- Supports multiple linters in one tool
- Easy CI integration

**Alternatives Considered**:
- staticcheck: Good but limited to one tool
- go vet + gofmt: Too basic, not comprehensive enough

### D4: Documentation Platform
**Decision**: Static site with specledger.io domain

**Rationale**:
- Already established (specledger.io, specledger.io/docs)
- Full control over content and hosting
- Can use Hugo, Jekyll, or similar static site generators
- Fast, reliable, cost-effective

**Alternatives Considered**:
- GitHub Pages: Would require github.io subdomain
- Read the Docs: External dependency
- Notion/docs.google.com: Not ideal for public documentation

### D5: Release Automation
**Decision**: GoReleaser (already configured)

**Rationale**:
- Already in place and working
- Industry standard for Go projects
- Supports multi-platform builds
- Homebrew tap integration built-in

**Alternatives Considered**:
- Custom build scripts: More maintenance
- GitHub Releases only: Manual, error-prone

## Third-Party Dependencies

### Main Dependencies

All dependencies are from well-maintained open source projects with compatible licenses:

| Package | License | Purpose | Attribution Needed |
|---------|---------|---------|-------------------|
| github.com/spf13/cobra | Apache 2.0 | CLI framework | No (permissive) |
| github.com/charmbracelet/bubbletea | MIT | Terminal UI | No (same license) |
| github.com/go-git/go-git/v5 | Apache 2.0 | Git operations | No (permissive) |
| gopkg.in/yaml.v3 | Apache 2.0 | YAML parsing | No (permissive) |

**Conclusion**: No NOTICE file strictly required due to permissive licenses, but will create one for transparency as best practice.

## Quality Tools Configuration

### golangci-lint

Recommended configuration based on Go best practices:

```yaml
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
```

### Test Coverage

Target: 70%+ coverage for new code
Tool: `go test -coverprofile=coverage.out`
Badge: Coverage percentage displayed in README

## Documentation Structure

### Proposed Layout

```
docs/
├── index.md           # Main landing page
├── user/
│   ├── getting-started.md
│   ├── commands.md
│   ├── sdd-framework.md
│   └── troubleshooting.md
├── contributor/
│   ├── setup.md
│   ├── architecture.md
│   ├── development.md
│   └── testing.md
└── governance/
    ├── roadmap.md
    ├── decision-making.md
    └── maintenance.md
```

### README Badges

```markdown
[![Build Status](https://img.shields.io/github/actions/workflow/status/specledger/specledger/ci.yml?branch=main)](https://github.com/specledger/specledger/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/specledger/specledger)](https://goreportcard.com/report/github.com/specledger/specledger)
[![Coverage](https://img.shields.io/codecov/c/github/specledger/specledger)](https://codecov.io/gh/specledger/specledger)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Homebrew](https://img.shields.io/badge/dynamic/json?color=orange&label=homebrew&query=specledger&url=https://formulae.brew.sh/api/formula/specledger.json)](https://formulae.brew.sh/formula/specledger)
[![Version](https://img.shields.io/github/v/release/specledger/specledger)](https://github.com/specledger/specledger/releases)
```

## Implementation Gaps

Based on codebase exploration, the following items need to be implemented for 006-opensource-readiness:

### P1 Items (Must Have)
1. Create NOTICE file
2. Add CI workflow for quality checks
3. Add README badges
4. Create GOVERNANCE.md
5. Update Go version in CI to 1.24

### P2 Items (Should Have)
6. Set up Codecov for coverage tracking
7. Create docs/ directory structure
8. Add documentation deployment workflow
9. Improve unit test coverage

### P3 Items (Nice to Have)
10. Add performance benchmarks
11. Create architecture diagrams
12. Add contributor recognition (AUTHORS.md)

## Security Considerations

### Dependency Scanning
- GitHub Dependabot already enabled
- Monthly security reviews recommended
- govulncheck integration in CI

### Vulnerability Reporting
- SECURITY.md already in place
- Private vulnerability reporting via GitHub
- Security advisory policy documented

## Performance Targets

Based on spec requirements:
- Homebrew installation: < 2 minutes ✅ (already fast)
- Release automation: < 10 minutes ✅ (GoReleaser is efficient)
- CI feedback: < 5 minutes ✅ (GitHub Actions is fast)

## Next Steps

1. Update plan.md with research findings
2. Create data-model.md (if applicable)
3. Create contracts/ directory (if applicable)
4. Create quickstart.md
5. Run agent context update script
