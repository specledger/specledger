# Data Model: Open Source Readiness

**Feature**: 006-opensource-readiness | **Date**: 2026-02-09

## Overview

This feature is primarily concerned with documentation, legal compliance, and infrastructure rather than data structures. However, there are several key entities and their relationships that are relevant to the open source readiness of the project.

## Key Entities

### Legal Documents

#### LICENSE
| Field | Type | Description |
|-------|------|-------------|
| type | string | License identifier (e.g., "MIT") |
| copyright | string | Copyright holder(s) |
| year | string | Copyright year(s) |
| permissions | string[] | Granted permissions |
| conditions | string[] | Conditions imposed |
| limitations | string[] | Limitations of liability |

#### NOTICE
| Field | Type | Description |
|-------|------|-------------|
| project | string | Project name |
| third_party_attribution | ThirdPartyAttribution[] | List of third-party works |

### ThirdPartyAttribution
| Field | Type | Description |
|-------|------|-------------|
| name | string | Name of the third-party work |
| author | string | Author/owner of the work |
| license | string | License under which it's used |
| url | string | Source URL |

### Governance

#### Maintainer
| Field | Type | Description |
|-------|------|-------------|
| name | string | Maintainer's name |
| role | string | Role in the project |
| responsibilities | string[] | Areas of responsibility |
| contact | string | Preferred contact method |

#### Contribution
| Field | Type | Description |
|-------|------|-------------|
| contributor | string | Contributor's name |
| contribution | string | Description of contribution |
| date | datetime | Date of contribution |
| license_grant | string | License grant for the contribution |

### Release

#### Release
| Field | Type | Description |
|-------|------|-------------|
| version | string | Semantic version (MAJOR.MINOR.PATCH) |
| tag | string | Git tag |
| date | datetime | Release date |
| platforms | Platform[] | Supported platforms |
| artifacts | Artifact[] | Release artifacts |

#### Platform
| Field | Type | Description |
|-------|------|-------------|
| os | string | Operating system (linux, darwin, windows) |
| arch | string | Architecture (amd64, arm64, arm) |
| binary | string | Binary filename |

#### Artifact
| Field | Type | Description |
|-------|------|-------------|
| name | string | Artifact name |
| type | string | Artifact type (binary, archive, checksum) |
| url | string | Download URL |
| checksum | string | SHA256 checksum |

### Documentation

#### DocumentationPage
| Field | Type | Description |
|-------|------|-------------|
| path | string | URL path (e.g., /docs/getting-started) |
| title | string | Page title |
| content | string | Page content (markdown) |
| last_updated | datetime | Last update timestamp |
| version | string | Applicable version |

#### Badge
| Field | Type | Description |
|-------|------|-------------|
| type | string | Badge type (build, coverage, license, version) |
| url | string | Badge image URL |
| link | string | Link target |
| status | string | Current status |

## Entity Relationships

```
┌─────────────┐
│   LICENSE   │
└─────────────┘
       │
       │ grants
       ▼
┌─────────────────────┐
│   Contribution      │───┐
└─────────────────────┘   │
                          │ contributed_by
                          ▼
                    ┌──────────────┐
                    │ Contributor  │
                    └──────────────┘

┌─────────────┐       ┌──────────────┐
│   NOTICE    │───────│ThirdPartyAttr│
└─────────────┘       └──────────────┘

┌─────────────┐
│   Release   │
└─────────────┘
       │
       │ contains
       ▼
┌─────────────────────┐
│     Platform        │───┐
└─────────────────────┘   │
                          │ builds
                          ▼
                    ┌──────────────┐
                    │   Artifact   │
                    └──────────────┘

┌───────────────────┐
│ DocumentationPage │
└───────────────────┘
       │
       │ referenced_by
       ▼
┌──────────────┐
│    Badge     │
└──────────────┘

┌──────────────┐
│  Maintainer  │
└──────────────┘
       │
       │ governs
       ▼
┌───────────────────┐
│     Project       │
└───────────────────┘
```

## State Transitions

### Release State
```
draft ──→ tagged ──→ building ──→ published
            ↑                      │
            └──────────────────────┘
                  (rollback)
```

### Contribution State
```
proposed ──→ reviewed ──→ accepted ──→ integrated
              │                       │
              └───────────────────────┘
                    (rejected)
```

## Validation Rules

1. **LICENSE**: Must be one of the approved open source licenses (MIT, Apache-2.0, BSD-3-Clause)
2. **Release.version**: Must follow semantic versioning (MAJOR.MINOR.PATCH)
3. **Artifact.checksum**: Must be valid SHA256 hash
4. **DocumentationPage.path**: Must be unique, must start with /docs/
5. **Badge.url**: Must be a valid HTTPS URL to a badge image
6. **Contribution.license_grant**: Must be compatible with project license

## Storage Notes

- **Legal files** (LICENSE, NOTICE, GOVERNANCE): Stored as plain text in repository root
- **Documentation**: Stored as markdown in `docs/` directory
- **CI/CD configurations**: Stored as YAML in `.github/workflows/`
- **Release metadata**: Stored in CHANGELOG.md and Git tags
- **Badges**: Embedded in README.md as markdown links

## No Database Required

This feature does not require a database. All data is stored in:
1. Git repository (source code, documentation, legal files)
2. GitHub (issues, releases, actions)
3. specledger.io (published documentation)
