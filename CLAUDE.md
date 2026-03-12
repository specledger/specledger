# Active Technologies

This file is auto-generated from plan.md. Manual additions are preserved below.

## Active Technologies

- Cobra (CLI)
- Go 1.24.2
- JSONL file store (pkg/issues)
- JSONL files per spec (`specledger/<spec>/issues.jsonl`)
- GoReleaser (build/release)
- N/A (configuration fix)

<!-- MANUAL ADDITIONS START -->

## Commit & Push Rules

- **NEVER** run `git commit` or `git push` directly. Always use the `/specledger.commit` skill for all commit and push operations. This ensures auth-aware session capture works correctly.

<!-- MANUAL ADDITIONS END -->
