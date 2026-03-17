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

## Commit & Push Rules

- **NEVER** run `git commit` or `git push` directly. Always use the `/specledger.commit` skill for all commit and push operations. This ensures auth-aware session capture works correctly.

## Pre-push Checklist

- `make lint` — golangci-lint v2 (install: `mise install golangci-lint`)
- `make test` — unit tests
- `make fmt` — formatting

<!-- MANUAL ADDITIONS END -->
