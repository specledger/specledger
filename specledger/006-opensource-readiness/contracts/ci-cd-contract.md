# CI/CD Contract

**Feature**: 006-opensource-readiness | **Version**: 1.0

## Overview

This contract defines the continuous integration and continuous delivery requirements for the SpecLedger project.

## CI Requirements

### Workflow File

**Location**: `.github/workflows/ci.yml`
**Trigger Events**:
- Push to `main` branch
- Pull requests to `main` branch
- Manual workflow dispatch

### Jobs

#### Job 1: Test

**Required Steps**:
1. Checkout code
2. Set up Go 1.24+
3. Download dependencies
4. Run all tests
5. Generate coverage report
6. Upload to Codecov

**Contract**:
```yaml
test:
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-go@v5
      with:
        go-version: '1.24'
    - run: go mod download
    - run: make test
    - run: make test-coverage
    - uses: codecov/codecov-action@v4
```

#### Job 2: Lint

**Required Steps**:
1. Checkout code
2. Set up Go 1.24+
3. Run golangci-lint

**Contract**:
```yaml
lint:
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-go@v5
      with:
        go-version: '1.24'
    - uses: golangci/golangci-lint-action@v6
```

#### Job 3: Format

**Required Steps**:
1. Checkout code
2. Set up Go 1.24+
3. Check Go formatting

**Contract**:
```yaml
format:
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-go@v5
      with:
        go-version: '1.24'
    - run: |
        if [ -n "$(gofmt -l .)" ]; then
          echo "Go code is not formatted"
          exit 1
        fi
```

### Quality Gates

All jobs MUST pass before merge is allowed:
- Test coverage: Minimum 70%
- Lint: Zero high-severity issues
- Format: No formatting differences

## CD Requirements

### Release Workflow

**Location**: `.github/workflows/release.yml`
**Trigger Events**:
- Git tag creation matching `v*.*.*`

### GoReleaser Configuration

**Location**: `.goreleaser.yaml`

**Required Builds**:

| OS | Architecture | Binary Name |
|----|--------------|-------------|
| linux | amd64 | specledger_linux_amd64 |
| linux | arm64 | specledger_linux_arm64 |
| darwin | amd64 | specledger_darwin_amd64 |
| darwin | arm64 | specledger_darwin_arm64 |
| windows | amd64 | specledger_windows_amd64.exe |

**Required Artifacts**:
- Binary for each platform
- SHA256 checksums file
- Homebrew formula update

**Contract**:
```yaml
builds:
  - env:
      - CGO_ENABLED=0
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w -X main.version={{.Version}}
    main: ./cmd/specledger

checksum:
  name_template: 'checksums.txt'

brews:
  - name: specledger
    tap:
      owner: specledger
      name: homebrew-specledger
    commit_author:
      name: specledger-bot
      email: bot@specledger.io
```

### Release Timeline

- Tag creation → Release published: < 10 minutes
- Homebrew formula updated: < 5 minutes after release
- All artifacts uploaded: < 10 minutes

## Badge Requirements

### Required Badges in README

All badges MUST be present at the top of README.md:

```markdown
[![Build Status](https://img.shields.io/github/actions/workflow/status/specledger/specledger/ci.yml?branch=main)](https://github.com/specledger/specledger/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/specledger/specledger)](https://goreportcard.com/report/github.com/specledger/specledger)
[![Coverage](https://img.shields.io/codecov/c/github/specledger/specledger)](https://codecov.io/gh/specledger/specledger)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Version](https://img.shields.io/github/v/release/specledger/specledger)](https://github.com/specledger/specledger/releases)
```

### Badge Validation

```bash
# Verify badges in README
grep -q "Build Status" README.md
grep -q "Go Report Card" README.md
grep -q "codecov" README.md
grep -q "License: MIT" README.md
grep -q "github/v/release" README.md
```

## golangci-lint Configuration

**Location**: `.golangci.yml`

**Enabled Linters**:
- gofmt
- govet
- staticcheck
- errcheck
- gosimple
- ineffassign
- unused
- gosec

**Configuration**:
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

run:
  timeout: 5m
  tests: true
```

## Testing Contract

### Test Requirements

1. All packages must have tests
2. Test coverage minimum: 70%
3. Integration tests must cover:
   - CLI commands
   - Git operations
   - Template processing

### Makefile Targets

```makefile
.PHONY: test test-coverage

test:
	go test -v ./...

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
```

## Acceptance Criteria

1. CI workflow runs on all pushes and PRs
2. All three jobs (test, lint, format) must pass
3. Coverage reports uploaded to Codecov
4. GoReleaser creates all required artifacts
5. Homebrew formula updated automatically
6. All badges display correctly in README
7. Release completes in under 10 minutes
8. No manual intervention required for releases
