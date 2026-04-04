# Active Technologies

This file is auto-generated from plan.md. Manual additions are preserved below.

## Active Technologies

- Cobra (CLI)
- Embedded filesystem (`pkg/embedded/`) + local file I/O
- File-based (YAML config files, JSONL for issues)
- Go 1.24.2
- Go embed FS
- Go testing package (`_test.go` files)
- go-git/v5
- JSONL file store (pkg/issues)
- JSONL files per spec (`specledger/<spec>/issues.jsonl`)
- GoReleaser (build/release)
- `go test` with table-driven tests
- YAML v3 (config)

<!-- MANUAL ADDITIONS START -->

## Commits & PRs

For conventional commit types, version bump rules, and the release flow, see [docs/guides/release-flow.md](docs/guides/release-flow.md).

## Pre-push Checklist

- `make lint` — golangci-lint v2 (install: `mise install golangci-lint`)
- `make test` — unit tests
- `make fmt` — formatting
- `zizmor .github/workflows/` — validate GitHub Actions workflows when modifying them (install: `mise install zizmor`)

<!-- MANUAL ADDITIONS END -->
