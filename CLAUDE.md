## Commits & PRs

For conventional commit types, version bump rules, and the release flow, see [docs/guides/release-flow.md](docs/guides/release-flow.md).

## Pre-push Checklist

- `make lint` — golangci-lint v2 (install: `mise install golangci-lint`)
- `make test` — unit tests
- `make fmt` — formatting
- `zizmor .github/workflows/` — validate GitHub Actions workflows when modifying them (if missing, suggest `mise install zizmor` to install)

<!-- >>> specledger-generated -->
<!-- Auto-managed by specledger - do not edit this section -->
## Active Technologies

- Cobra (CLI)
- Embedded filesystem (`pkg/embedded/`) + local file I/O
- Go embed FS
- File-based (YAML config files, JSONL for issues)
- Go 1.24.2
- `go test` with table-driven unit tests + integration tests building full `sl` binary
- go-git/v5
- Two-tier: (1) `dnaeon/go-vcr` v4 cassettes for API client unit tests, (2) `httptest.Server` for full CLI E2E integration tests. Endpoint base URLs configurable via ENV vars (`SKILLS_API_URL`, `SKILLS_AUDIT_URL`, `GITHUB_API_URL`) for testability.
- `skills-lock.json` (Vercel-compatible local lock file, project root)
- GoReleaser (build) + release-please
- crypto/sha256 (lock file hashing)
- dnaeon/go-vcr v4 (test only — VCR cassettes)
- gopkg.in/yaml.v3 (SKILL.md frontmatter)
- net/http (API client)
<!-- <<< specledger-generated -->
