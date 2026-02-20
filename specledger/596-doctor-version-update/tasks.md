# Tasks Index: Doctor Version and Template Update

Issue tracking for this feature is managed through the built-in `sl issue` system.

## Feature Tracking

* **Epic ID**: `SL-826d4e`
* **User Stories Source**: `specledger/596-doctor-version-update/spec.md`
* **Research Inputs**: `specledger/596-doctor-version-update/research.md`
* **Planning Details**: `specledger/596-doctor-version-update/plan.md`
* **Data Model**: `specledger/596-doctor-version-update/data-model.md`
* **Quickstart**: `specledger/596-doctor-version-update/quickstart.md`

## Query Commands

```bash
# List all issues for this feature
sl issue list --label spec:596-doctor-version-update

# List open issues
sl issue list --label spec:596-doctor-version-update --status open

# View issue tree
sl issue show --tree SL-826d4e

# List ready-to-work issues
sl issue ready --label spec:596-doctor-version-update
```

## Phases Overview

| Phase | Issue ID | Type | Description |
|-------|----------|------|-------------|
| Epic | SL-826d4e | epic | Doctor Version and Template Update |
| Foundational | SL-be5e0e | feature | Schema and Types |
| US1 | SL-011a55 | feature | CLI Version Check and Update Prompt |
| US2 | SL-a00083 | feature | Project Template Update |
| US3 | SL-fb291a | feature | CI/CD JSON Output Support |
| Polish | SL-0d2400 | feature | Documentation and Validation |

## Task Summary by User Story

### Foundational: Schema and Types (SL-be5e0e)

| Issue ID | Title | Priority | Status |
|----------|-------|----------|--------|
| SL-e206f1 | Add TemplateVersion field to ProjectMetadata | 0 | open |
| SL-9524be | Create pkg/version package with VersionInfo type | 0 | open |

### US1: CLI Version Check (SL-011a55) - P1 🎯 MVP

| Issue ID | Title | Priority | Status |
|----------|-------|----------|--------|
| SL-faa3a6 | Implement GitHub Releases API client | 1 | open |
| SL-445a25 | Add CLI version section to doctor output | 1 | open |
| SL-a9a5e9 | Add update instructions logic | 1 | open |

**Independent Test**: Run `sl doctor` with outdated CLI and verify version info displayed.

### US2: Template Update (SL-a00083) - P1

| Issue ID | Title | Priority | Status |
|----------|-------|----------|--------|
| SL-fd58b1 | Create pkg/templates package with TemplateStatus type | 1 | open |
| SL-20ecb8 | Implement template diff and customized file detection | 1 | open |
| SL-02616f | Implement template updater with skip logic | 1 | open |
| SL-8c624a | Add interactive template update prompt to doctor | 1 | open |

**Independent Test**: Initialize project with older CLI, update CLI, run `sl doctor`, accept template update.

### US3: CI/CD JSON Support (SL-fb291a) - P2

| Issue ID | Title | Priority | Status |
|----------|-------|----------|--------|
| SL-5e7876 | Extend DoctorOutput struct with version fields | 2 | open |
| SL-dda5e1 | Update JSON output to include version/template info | 2 | open |

**Independent Test**: Run `sl doctor --json` and parse structured output.

### Polish: Documentation (SL-0d2400) - P3

| Issue ID | Title | Priority | Status |
|----------|-------|----------|--------|
| SL-74478b | Update CLAUDE.md with feature details | 3 | open |
| SL-041a6b | Run quickstart.md validation tests | 3 | open |

## Dependency Graph

```
SL-e206f1 ─────────────────────────────────┬──> SL-445a25 ──┬──> SL-5e7876 ──> SL-dda5e1 ──┬──> SL-74478b
(TemplateVersion field)                    │                │                               │
                                           │                │                               └──> SL-041a6b
SL-9524be ──> SL-faa3a6 ──────────────────┤                │
(VersionInfo type)   (GitHub API)          │                └──> SL-8c624a <── SL-02616f <── SL-20ecb8 <── SL-fd58b1
                     │                     │                      ↑              (Updater)        (Diff)         (Status type)
                     └──> SL-445a25 ───────┘                      │
                     (Version display)                            │
                                                                 │
                                           SL-e206f1 ────────────┘
```

## Definition of Done Summary

| Issue ID | DoD Items |
|----------|-----------|
| SL-e206f1 | TemplateVersion field added, yaml tag correct, NewProjectMetadata updated, tests pass |
| SL-9524be | pkg/version/ created, VersionInfo struct, CompareVersions function |
| SL-faa3a6 | checker.go created, CheckLatestVersion implemented, timeout/error handling |
| SL-445a25 | CLI version section added, version check integrated, error states handled |
| SL-a9a5e9 | instructions.go created, GetUpdateInstructions function, method detection |
| SL-fd58b1 | pkg/templates/ created, TemplateStatus struct, CheckTemplateStatus function |
| SL-20ecb8 | diff.go created, IsFileCustomized, FindCustomizedFiles, SHA-256 working |
| SL-02616f | updater.go created, UpdateTemplates, skip logic, permissions, YAML update |
| SL-8c624a | Template status check, interactive prompt, uncommitted warning, update execution |
| SL-5e7876 | DoctorOutput struct extended, fields populated |
| SL-dda5e1 | Version/template in JSON, template updates skipped in JSON mode |
| SL-74478b | CLAUDE.md updated with technologies and feature |
| SL-041a6b | Version check verified, JSON output verified, template update verified |

## MVP Scope

**Minimum Viable Product**: Foundational + US1 (3 issues)
- SL-e206f1: Add TemplateVersion field
- SL-9524be: Create pkg/version package
- SL-faa3a6: Implement GitHub API client
- SL-445a25: Add CLI version section
- SL-a9a5e9: Add update instructions

This delivers the core value: users can see their CLI version and know when to update.

## Execution Strategy

1. **Foundational First**: Complete SL-e206f1 and SL-9524be in parallel
2. **US1 MVP**: Complete SL-faa3a6, SL-445a25, SL-a9a5e9 for CLI version checking
3. **US2 Parallel**: Start US2 tasks after foundational, can run alongside US1
4. **US3 After US1+US2**: JSON output requires version check and template update complete
5. **Polish Last**: Documentation and validation after all features complete
