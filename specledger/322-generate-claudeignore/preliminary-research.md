# Preliminary Research: `.claudeignore` vs `permissions.deny`

**Spec**: 322-generate-claudeignore
**Date**: 2026-02-26

## Key Finding

**`.claudeignore` is a deprecated/transitional feature.** The current recommended approach is using `permissions.deny` in `.claude/settings.json`, which provides:
- Stronger enforcement (blocks read operations, not just discovery)
- Fine-grained control with gitignore-style patterns
- Scope-aware configuration (user, project, local, managed)

## Purpose

| Purpose | Description |
|---------|-------------|
| **Token Efficiency** | Exclude large directories (`node_modules/`, `dist/`, `build/`) to prevent context window bloat and reduce API costs |
| **Sensitive Data Exclusion** | Block access to `.env`, `.env.*`, `secrets/**`, credentials files to prevent Claude from reading or exposing sensitive information |

## Known Issues

1. **Session caching:** When resuming sessions after adding ignore patterns, old cached content may still be loaded
2. **File watching:** File watcher may still process excluded directories like `node_modules/`, potentially causing OOM crashes
3. **`.gitignore` alone is insufficient for security** - Claude Code can still read gitignored files unless explicitly blocked

## Recommended Modern Approach

```json
// .claude/settings.json
{
  "respectGitignore": true,
  "permissions": {
    "deny": [
      "Read(./.env)",
      "Read(./.env.*)",
      "Read(./secrets/**)",
      "Read(./node_modules/**)",
      "Read(./dist/**)"
    ]
  }
}
```

## Conclusion

This spec should generate `permissions.deny` entries in `.claude/settings.json` instead of (or in addition to) `.claudeignore`.

## Sources

1. **Official Documentation:**
   - Claude Code Settings: https://docs.anthropic.com/en/docs/claude-code/settings
   - Claude Code Permissions: https://code.claude.com/docs/en/permissions.md
   - Claude Code Security: https://code.claude.com/docs/en/security.md

2. **GitHub Repository:**
   - anthropics/claude-code: https://github.com/anthropics/claude-code

3. **GitHub Issues:**
   - Issue #23033: Feature request for `--reapply-ignore` flag
   - Issue #24185: Bug report about `.env` files being read despite `.gitignore`
   - Issue #27863: Bug report about OOM crashes with `node_modules`
