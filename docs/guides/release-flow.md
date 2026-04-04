# Release Flow

Releases are managed by [release-please](https://github.com/googleapis/release-please). Version bumps are determined automatically from conventional commit messages on `main`.

## How it works

1. PRs are squash-merged to `main`. The PR title becomes the commit message.
2. Release-please analyzes new commits and opens (or updates) a release PR.
3. The release PR updates `CHANGELOG.md` and bumps the version.
4. When the release PR is merged, a GitHub Release and tag are created, and GoReleaser builds and publishes binaries.

## Conventional commit types

PR titles must follow the format: `type(scope): description` or `type: description`.

### Version-bumping types

| Type | Version bump | Description |
|------|-------------|-------------|
| `feat` | minor (0.x → 0.x+1) | New feature |
| `fix` | patch (0.0.x → 0.0.x+1) | Bug fix |
| `perf` | patch | Performance improvement |
| `deps` | patch | Dependency update |
| `revert` | patch | Revert a previous commit |

A commit with `!` after the type (e.g., `feat!: remove old API`) or a `BREAKING CHANGE:` footer triggers a **minor** bump (pre-1.0 behavior via `bump-minor-pre-major` config).

### Non-bumping types

These are included in the release PR changelog under "Miscellaneous Chores" but do **not** trigger a version bump on their own. A release PR is only created when at least one version-bumping commit is present.

| Type | Description |
|------|-------------|
| `chore` | Maintenance tasks, config changes |
| `docs` | Documentation only |
| `style` | Formatting, whitespace |
| `refactor` | Code restructuring (no behavior change) |
| `test` | Adding or updating tests |
| `build` | Build system or external dependencies |
| `ci` | CI/CD configuration |

## Examples

```
feat: add spec diff command              → minor bump, appears in "Features"
feat(audit): support JSON output         → minor bump, scoped to "audit"
fix: correct YAML parsing for nested keys → patch bump, appears in "Bug Fixes"
feat!: drop support for Go 1.22          → minor bump (breaking), appears in "Features" with ⚠ marker
chore: update linter config              → no bump, included in changelog
ci: migrate to release-please            → no bump, included in changelog
docs: update installation guide          → no bump, included in changelog
```

## Configuration

- `release-please-config.json` — release-please behavior (release type, tag format, bump rules)
- `.release-please-manifest.json` — tracks current released version
- `.github/workflows/release.yml` — workflow (release-please + goreleaser)
- `.github/workflows/pr-title.yml` — enforces conventional commit format on PR titles
