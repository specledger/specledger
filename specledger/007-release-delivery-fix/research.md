# Research: Release Delivery Fix

**Feature**: 007-release-delivery-fix
**Date**: 2025-02-09

## Prior Work

### From Spec 006-opensource-readiness

The open-source readiness feature established the baseline release infrastructure:
- GoReleaser v2 configuration for automated releases
- GitHub Actions workflow for release automation
- Install script for one-line installation
- Homebrew tap configuration (created but not published)

**Completed Tasks**:
- SL-hlt: Setup Codecov integration (completed during this branch)
- SL-rfk: Verify GoReleaser configuration (completed but had deprecation warnings)

**Open Tasks**:
- SL-m1x: Dry-run release verification
- SL-4tk: Test release process dry-run

### Recent Fixes (Branch 007-build-delivery-setup)

During this session, the following fixes were applied:
- Fixed GoReleaser v2 deprecation warnings (`archives.format`, `brews` vs `homebrew_casks`)
- Fixed archive naming from `sl_*` to `specledger_*`
- Fixed broken `windows_arm_7` target by splitting builds
- Added Codecov configuration

## Decision: Platform Scope - macOS Only

**Decision**: Focus on macOS (darwin) with amd64 and arm64 architectures only.

**Rationale**:
1. User explicitly requested "for now focus first on mac development"
2. Reduces testing scope for initial validation
3. macOS is the primary development environment for the target audience
4. Linux and Windows can be added in future iterations once macOS flow is validated

**Implementation**: Simplify GoReleaser config to build only darwin_amd64 and darwin_arm64.

## Decision: Archive Structure

**Investigation**: GoReleaser default archive structure

When GoReleaser creates archives with `name_template: 'specledger_{{ .Version }}_{{ .Os }}_{{ .Arch }}'`, it creates:
- Archive file: `specledger_1.0.2_darwin_amd64.tar.gz`
- When extracted, contains: `sl` binary directly at root (not in a subdirectory)

**Issue Found**: The install script at line 134 looks for `$extract_dir/sl` which is correct, but at line 132 it references a nested path for windows which doesn't match the actual structure.

**Decision**: Keep archive structure as-is (flat with `sl` at root), but verify install script matches.

## Decision: Architecture Detection on macOS

**Investigation**: How to detect Intel vs Apple Silicon

macOS architecture detection methods:
1. `uname -m` returns `x86_64` for Intel, `arm64` for Apple Silicon
2. `arch` command also returns architecture
3. Go's `runtime.GOARCH` returns `amd64` or `arm64`

**Decision**: Use `uname -m` in bash script and map:
- `x86_64` → `amd64`
- `arm64` → `arm64`

## Decision: Homebrew Tap Repository

**Investigation**: Does `github.com/specledger/homebrew-specledger` exist?

The GoReleaser config references `repository: owner: specledger, name: homebrew-specledger` with `skip_upload: auto`.

**Finding**: The repository likely doesn't exist yet (would need manual creation or GitHub API).

**Decision**:
1. Keep `skip_upload: auto` for now (skips upload if repo doesn't exist)
2. Document need to create the repository
3. Provide instructions for manual creation or creation via GitHub CLI/API

## Decision: Version Flag Implementation

**Investigation**: How to add version display to Go CLI

GoReleaser supports version injection via ldflags:
```yaml
ldflags:
  - -s -w -X main.version={{.Version}} -X main.commit={{.Commit}} -X main.date={{.Date}}
```

The CLI needs a `--version` flag that displays this information.

**Implementation**: Add version variable in main.go and a flag to display it.

## Decision: Checksum Verification

**Investigation**: How to verify binary checksums

GoReleaser generates `checksums.txt` containing SHA256 hashes of all archives.

**Implementation**:
1. Install script should download `checksums.txt` along with the binary
2. Use `shasum -a 256 -c` to verify the downloaded archive
3. Fail installation if checksum doesn't match

## Alternatives Considered

### Alternative 1: Use GoReleaser's Homebrew formula generation with default tap

**Rejected**: Custom tap (`homebrew-specledger`) provides better branding and control than using default Homebrew core.

### Alternative 2: Support all platforms (Linux, Windows, macOS) in this feature

**Rejected**: User explicitly requested macOS focus. Multi-platform support increases testing complexity and should be a separate feature.

### Alternative 3: Use different installation method (e.g., install via Go modules only)

**Rejected**: Users expect multiple installation options (binary, Homebrew, script) for CLI tools. Single method reduces adoption.

## Open Questions

1. **Homebrew Tap Creation**: Who will create the `homebrew-specledger` repository? Needs to be done before Homebrew installation works.

2. **Testing on Apple Silicon**: Do we have access to Apple Silicon hardware for testing the arm64 build?

3. **Release Frequency**: What is the intended release cadence? (Affects whether to automate full tap management)

## Summary of Required Changes

1. **GoReleaser config**: Simplify to macOS-only builds
2. **Install script**: Add arm64 detection and checksum verification
3. **CLI code**: Add `--version` flag
4. **Homebrew tap**: Create repository (manual or via API)
5. **Documentation**: Update README with verified instructions
