# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.2.1](https://github.com/specledger/specledger/compare/v1.2.0...v1.2.1) (2026-04-10)


### Bug Fixes

* **skills:** resolve default branch for repos not using main ([#173](https://github.com/specledger/specledger/issues/173)) ([3ab14c4](https://github.com/specledger/specledger/commit/3ab14c4bdf7e5a0b79146bc6316b998b527d9cfd))

## [1.2.0](https://github.com/specledger/specledger/compare/v1.1.0...v1.2.0) (2026-04-09)


### Features

* **skill:** add sl skill command with skills.sh registry integration ([#167](https://github.com/specledger/specledger/issues/167)) ([9eddd7c](https://github.com/specledger/specledger/commit/9eddd7c315ab15c5a24b0be0e5a7329e61c4ba7f))


### Bug Fixes

* enable go-git worktree support across all repo open calls ([#165](https://github.com/specledger/specledger/issues/165)) ([de7f007](https://github.com/specledger/specledger/commit/de7f007f01da9f775bdff10d3273a251d5202df7))
* resolve dependabot CI failures and bump dependencies ([#160](https://github.com/specledger/specledger/issues/160)) ([aac6eef](https://github.com/specledger/specledger/commit/aac6eef8f2de40922bb0f4db2ee5e78c1e84d09e))
* **skills:** add verify and checkpoint prompts to spec-driven workflow ([#169](https://github.com/specledger/specledger/issues/169)) ([07d9d8e](https://github.com/specledger/specledger/commit/07d9d8e89e7eaec1d3924760dffd2516d6a245b4))

## [1.1.0](https://github.com/specledger/specledger/compare/v1.0.65...v1.1.0) (2026-04-04)


### Features

* **init:** add --agent flag for non-interactive agent selection ([#157](https://github.com/specledger/specledger/issues/157)) ([87aad06](https://github.com/specledger/specledger/commit/87aad06e7f81ffa37dfe4f99c54d30333f7ceaab))


### Bug Fixes

* detect missing git repository in sl init and offer to initialize ([#152](https://github.com/specledger/specledger/issues/152)) ([aefd202](https://github.com/specledger/specledger/commit/aefd2028b18b284976e0784c874fe8263461fff6))
* improve agent prompt behavior for task tracking, interactivity, and template reading ([#154](https://github.com/specledger/specledger/issues/154)) ([5741a08](https://github.com/specledger/specledger/commit/5741a087ee714c19ffc1b42157690f55c0dce422))
* improve embedded skill templates — fix duplicates, optimize triggering, reduce tokens ([#158](https://github.com/specledger/specledger/issues/158)) ([46ea82f](https://github.com/specledger/specledger/commit/46ea82f0d12190adbb3561da9f34fd7b097c4c6f))
* prevent template updates from overwriting user-customized files ([#148](https://github.com/specledger/specledger/issues/148)) ([d1a6e51](https://github.com/specledger/specledger/commit/d1a6e5147e0d6361c2dc3f3f08e112302251cbf9))
* remove beads from onboarding prompt ([#143](https://github.com/specledger/specledger/issues/143)) ([818d2fe](https://github.com/specledger/specledger/commit/818d2fedc509c69ed187704a86f6e564ec5d1c8a))
* steer onboarding constitution toward design principles, not tech stack ([#155](https://github.com/specledger/specledger/issues/155)) ([b8c3f60](https://github.com/specledger/specledger/commit/b8c3f6002302ad12e4e28607af73c3c9fc29e6d3))
* support SSH remote aliases and add --repo flag for manual override ([#151](https://github.com/specledger/specledger/issues/151)) ([f3d17f9](https://github.com/specledger/specledger/commit/f3d17f94fea20fc982aac5eebecc3d5996a1ed61))

## [Unreleased]

### Changed
- (no changes yet)

## [1.0.1] - 2026-02-09

### Fixed
- Fixed golangci-lint v1.64.8 compatibility issues (removed deprecated `version` field and `check-shadowing` setting)
- Removed unused code identified by linter (15 items across 8 files)
- Added gosec `#nosec` annotations for appropriate cases (subprocess execution, file permissions)
- Fixed errcheck issues (ignored return values appropriately)

### Changed
- Updated Makefile test targets to exclude integration tests from CI pipeline
- Added `make lint` target for running golangci-lint
- Added `make test-integration` target for running integration tests separately
- Integration tests now skip when prerequisites (mise, beads, perles) are not installed

## [1.0.0] - 2026-01-31

## [1.0.0] - 2026-01-31

### Added
- CLI unification from bash `sl` script and Go `specledger` CLI
- Interactive TUI for project bootstrap using Bubble Tea
- Non-interactive bootstrap with flags (--project-name, --short-code, etc.)
- Specification dependency management commands (add, list, remove, resolve, update)
- GitHub Releases distribution with GoReleaser
- Cross-platform binary builds (Linux, macOS, Windows) for multiple architectures
- Installation scripts for all platforms (bash, PowerShell)
- Debug logging system
- CLI configuration system (~/.config/specledger/config.yaml)
- Error handling with actionable suggestions
- Local dependency caching at ~/.specledger/cache/
- Dependency manifest (specledger.mod) and lockfile (specledger.sum)
- SHA-256 content hashing for cryptographic verification

### Security
- No hardcoded credentials
- Cryptographic verification of cached dependencies
- All configuration is optional
- No data is transmitted externally

### Documentation
- README with installation instructions
- Migration guide from legacy scripts
- Embedded templates for new projects
- AGENTS.md for SpecLedger workflow documentation

### Technical
- Cobra CLI framework for command structure
- Bubble Tea for terminal UI
- Go 1.21+ with go:embed for template embedding
- Cross-platform build support via GoReleaser
- Dependency management via Go modules
