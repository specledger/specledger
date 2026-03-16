# Active Technologies

This file is auto-generated from plan.md. Manual additions are preserved below.

## Active Technologies

- Cobra (CLI)
- Go 1.24.2
- Go embed FS
- JSONL files per spec (`specledger/<spec>/issues.jsonl`), YAML config (`specledger.yaml`), embedded filesystem (`pkg/embedded/`)
- `go test` with table-driven tests (unit), binary invocation via `exec.Command` (integration in `tests/integration/`)
- `pkg/cli/playbooks/` (template management)
- `pkg/templates/` (update logic)

<!-- MANUAL ADDITIONS START -->

# >>> specledger-generated
# Auto-managed by specledger - do not edit this section

## Session Start

- Run `sl doctor --json` to verify CLI version and template freshness. If `cli_update_available` or `template_update_available` is true, suggest running `sl doctor --update --template` to resolve.

# <<< specledger-generated

## Pre-push Checklist

- `make lint` — golangci-lint v2 (install: `mise install golangci-lint`)
- `make test` — unit tests
- `make fmt` — formatting

<!-- MANUAL ADDITIONS END -->
