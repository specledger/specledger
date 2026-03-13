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

## Skill Routing Rules

- **Commit/Push**: ALWAYS use `specledger.commit` skill when the user asks to commit, push, or save changes to git/github. Never run manual `git commit` or `git push` commands directly. This applies to ALL languages (Vietnamese, English, etc.) and ALL phrasings (e.g. "commit and push", "commit giúp tôi", "push to github", "commit for me", "save and push to github").

<!-- MANUAL ADDITIONS END -->
