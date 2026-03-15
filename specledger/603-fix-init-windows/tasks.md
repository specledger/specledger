# Tasks Index: Fix sl init on Windows

**Epic ID**: `SL-451d15`
**User Stories Source**: [spec.md](spec.md)
**Research Inputs**: [research.md](research.md)
**Planning Details**: [plan.md](plan.md)
**Data Model**: N/A (bug fix only)
**Contract Definitions**: N/A

## Query Hints

```bash
# All tasks for this feature
sl issue list --label "spec:603-fix-init-windows"

# Open tasks only
sl issue list --label "spec:603-fix-init-windows" --status open

# Tasks by user story
sl issue list --label "spec:603-fix-init-windows" --label "story:US1"
sl issue list --label "spec:603-fix-init-windows" --label "story:US2"

# View epic tree
sl issue show SL-451d15
```

## Structure

```
SL-451d15  Epic: Fix sl init on Windows
├── SL-12d91b  Foundational: Fix embed.FS path separators in manifest and playbook lookup
│   ├── SL-557dfa  Fix manifest.go: use path.Join for embed.FS manifest path
│   └── SL-fbc2ad  Fix embedded.go: use path.Join for ValidatePlaybooks embed.FS paths
│                  (depends on SL-557dfa - both edit embed.FS path construction logic)
│
├── SL-63a9d8  US1: sl init completes successfully on Windows  [blocked by SL-12d91b]
│   ├── SL-ef3870  Fix copy.go: use path.Join and strings.TrimPrefix
│   ├── SL-8102ed  Fix bootstrap_helpers.go: use path.Join for init.sh embed.FS path
│   ├── SL-576ff4  Add unit test: LoadManifest forward-slash paths
│   ├── SL-680bf2  Add unit test: embed.FS path separator in copy.go
│   └── SL-163cb7  Extend bootstrap_test.go: assert no playbook failure warnings
│
├── SL-e4a5b6  US2: Tool availability and post-init script on Windows  [blocked by SL-12d91b]
│   ├── SL-9e8ef1  Fix terminal.go checkGum: use exec.LookPath
│   └── SL-fe7aeb  Fix runPostInitScript: Windows shell detection
│
└── SL-26743b  Polish: make test + make vet (blocked by US1 and US2)
```

## Phase Dependencies

- **Foundational (SL-12d91b)** → blocks US1 and US2 phases
- **US1 (SL-63a9d8)** and **US2 (SL-e4a5b6)** → parallel after foundational
- **Polish (SL-26743b)** → blocked by both US1 and US2

## Parallel Opportunities

Within US1 (after foundational complete):
- `SL-ef3870` (copy.go) and `SL-8102ed` (bootstrap_helpers.go:449) → different files, fully parallel
- `SL-576ff4`, `SL-680bf2`, `SL-163cb7` (tests) → can be written in parallel with fixes

Within US2 (after foundational complete, parallel with US1):
- `SL-9e8ef1` (terminal.go) and `SL-fe7aeb` (bootstrap_helpers.go runPostInitScript) → different functions, parallel

## MVP Scope

**MVP = US1 (SL-63a9d8)** — fixes the primary error shown in the screenshot:
> `Playbook copying failed: failed to read manifest: open templates\manifest.yaml: file does not exist`
> `Error: failed to create project metadata: playbook name is required`

After MVP, US2 handles the secondary post-init script issue (graceful skip or shell execution).

## Definition of Done Summary

| Issue | DoD Items |
|-------|-----------|
| SL-557dfa | path.Join at line 12; path/filepath removed; path added; make build passes |
| SL-fbc2ad | path.Join at lines 90,97; path/filepath removed from embedded.go; make build passes |
| SL-ef3870 | path.Join for srcPath; filepath.Rel replaced with strings.TrimPrefix; make build passes |
| SL-8102ed | path.Join at line 449; path import added; filepath kept for real FS paths; make build passes |
| SL-576ff4 | manifest_test.go created; TestLoadManifestPathForwardSlash passes; make test passes |
| SL-680bf2 | TestEmbedFSPathSeparator added; forward slash resolves; backslash does not; make test passes |
| SL-163cb7 | Assertions added to TestBootstrapInitInExistingDirectory; make test passes |
| SL-9e8ef1 | checkGum uses exec.LookPath; exec.Command("command"...) removed; make build passes |
| SL-fe7aeb | runtime.GOOS branch added; findWindowsShell() added; graceful skip; make build passes |
| SL-26743b | make test: PASS; make vet: no errors; make fmt: no diffs |
