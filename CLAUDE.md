# specledger Development Guidelines

## Tech Stack
- **Language:** Go 1.24+ (current: 1.24.2)
- **CLI Framework:** Cobra
- **TUI:** Bubble Tea + Bubbles + Lipgloss
- **Git:** go-git v5
- **Config:** YAML v3
- **Auth/API:** Supabase (GoTrue for auth, PostgREST for data, Storage for files)
- **Build/Release:** GoReleaser v2, GitHub Actions, Homebrew
- **Storage:** File-based (JSONL for issues, JSON for credentials, YAML for config)

## Project Structure
```
cmd/sl/          # CLI entrypoint
pkg/cli/         # CLI commands and auth
pkg/embedded/    # Embedded templates and skills
pkg/issues/      # Issue tracking
pkg/models/      # Data models
internal/        # Internal packages (ref, spec)
templates/       # Embedded templates copied to user projects
.github/         # CI workflows
```

## Commands
```bash
make build          # Build binary to bin/sl
make test           # Run unit tests
make test-coverage  # Run tests with coverage report
make fmt            # Format code
make vet            # Vet code
make lint           # Run golangci-lint
make install        # Install sl to $GOBIN
```

## Code Style
- Follow standard Go conventions (`gofmt`, `go vet`)
- Linting: golangci-lint with govet, staticcheck, errcheck, ineffassign, unused, gosec
- **Note:** CI uses golangci-lint v1 config (`.golangci.yml`). Local v2 requires `--no-config` flag with inline linters

## CI Checks (`.github/workflows/ci.yml`)
1. **test** — `make test` + `make test-coverage` + Codecov upload
2. **lint** — `golangci-lint-action@v6`
3. **format** — `gofmt -l .` (fails if unformatted)

<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->

## Active Technologies
- Go 1.24.2 + Cobra (CLI), YAML v3 (config), JSONL (storage) (597-issue-create-fields)
- File-based JSONL at `specledger/<spec>/issues.jsonl` (597-issue-create-fields)
- Go 1.24.2 + Cobra (CLI), go-git v5, YAML v3, Supabase (GoTrue, PostgREST) (597-issue-create-fields)
- File-based JSONL for issues (597-issue-create-fields)
- Go 1.24.2 + Cobra (CLI), go-git v5 (git), gopkg.in/yaml.v3 (YAML parsing) (598-mockup-command)
- File-based — Markdown with YAML frontmatter (`design_system.md`), Markdown (`mockup.md`) (598-mockup-command)

## Recent Changes
- 597-issue-create-fields: Added Go 1.24.2 + Cobra (CLI), YAML v3 (config), JSONL (storage)
