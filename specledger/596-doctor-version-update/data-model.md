# Data Model: Doctor Version and Template Update

**Feature**: 596-doctor-version-update
**Date**: 2026-02-20

## Entities

### 1. VersionInfo

Represents version information from GitHub Releases API.

```
VersionInfo
├── CurrentVersion: string      # Installed CLI version (e.g., "1.2.0")
├── LatestVersion: string       # Latest available version (e.g., "1.3.0")
├── LatestURL: string           # URL to release page
├── UpdateAvailable: bool       # true if LatestVersion > CurrentVersion
└── CheckedAt: time.Time        # When the check was performed
```

**Validation**:
- Versions must be valid semver (parse with `github.com/Masterminds/semver/v3`)
- `UpdateAvailable` computed by comparing versions

### 2. TemplateStatus

Represents the state of project templates relative to current CLI.

```
TemplateStatus
├── ProjectTemplateVersion: string   # Version stored in specledger.yaml (may be empty)
├── CLIVersion: string               # Current CLI version
├── UpdateAvailable: bool            # true if versions differ
├── CustomizedFiles: []string        # Files that differ from embedded originals
├── TotalFiles: int                  # Total template files in project
└── NeedsUpdate: bool                # true if UpdateAvailable and in project
```

**State Transitions**:
```
[No Project] ──> [Skip Template Check]
     │
[In Project] ──> [TemplateVersion missing?] ──yes──> [Offer Update]
     │                    │
     no                   │
     ↓                    │
[Compare Versions] <──────┘
     │
     ├── [Match] ──> [Templates Current]
     │
     └── [Mismatch] ──> [Offer Update]
                              │
                              ├── [Accept] ──> [Update Templates]
                              │                    │
                              │                    └── [Update TemplateVersion in YAML]
                              │
                              └── [Decline] ──> [Skip Update]
```

### 3. TemplateUpdateResult

Result of template update operation.

```
TemplateUpdateResult
├── Updated: []string          # Files that were updated
├── Skipped: []string          # Customized files that were skipped
├── Errors: []error            # Any errors encountered
├── NewVersion: string         # New template_version written to YAML
└── Success: bool              # true if no fatal errors
```

### 4. DoctorOutput (Extended)

Extension of existing `DoctorOutput` struct in doctor.go.

```
DoctorOutput
├── Status: string                    # "pass" or "fail"
├── Tools: []DoctorToolStatus         # Existing
├── Missing: []string                 # Existing
├── InstallInstructions: string       # Existing
│
├── CLIVersion: string                # NEW - Current CLI version
├── CLILatestVersion: string          # NEW - Latest available version
├── CLIUpdateAvailable: bool          # NEW - CLI update available
├── CLIUpdateInstructions: string     # NEW - How to update CLI
│
├── TemplateVersion: string           # NEW - Project template version
├── TemplateUpdateAvailable: bool     # NEW - Template update available
└── TemplateCustomizedFiles: []string # NEW - Customized files detected
```

## Schema Changes

### ProjectMetadata Extension

Add `TemplateVersion` field to existing `ProjectMetadata` struct.

**Before**:
```go
type ProjectMetadata struct {
    Version      string          `yaml:"version"`
    Project      ProjectInfo     `yaml:"project"`
    Playbook     PlaybookInfo    `yaml:"playbook"`
    TaskTracker  TaskTrackerInfo `yaml:"task_tracker,omitempty"`
    ArtifactPath string          `yaml:"artifact_path,omitempty"`
    Dependencies []Dependency    `yaml:"dependencies,omitempty"`
}
```

**After**:
```go
type ProjectMetadata struct {
    Version         string          `yaml:"version"`
    Project         ProjectInfo     `yaml:"project"`
    Playbook        PlaybookInfo    `yaml:"playbook"`
    TemplateVersion string          `yaml:"template_version,omitempty"` // NEW
    TaskTracker     TaskTrackerInfo `yaml:"task_tracker,omitempty"`
    ArtifactPath    string          `yaml:"artifact_path,omitempty"`
    Dependencies    []Dependency    `yaml:"dependencies,omitempty"`
}
```

**Migration**:
- Optional field - no migration required
- Missing `template_version` treated as "needs update"

## File Checksums

### Checksum Storage (Runtime Only)

Checksums are computed at runtime, not stored.

```
FileChecksum
├── Path: string        # Relative path from .claude/
├── Hash: []byte        # SHA-256 hash (32 bytes)
└── IsCustomized: bool  # true if hash differs from embedded
```

**Algorithm**:
```go
func ComputeChecksums(projectDir string, embeddedFS embed.FS) (map[string][]byte, error) {
    checksums := make(map[string][]byte)

    // Walk .claude/ directory
    filepath.Walk(filepath.Join(projectDir, ".claude"), func(path string, info os.FileInfo, err error) error {
        if info.IsDir() {
            return nil
        }

        relPath := strings.TrimPrefix(path, filepath.Join(projectDir, ".claude") + "/")
        content, err := os.ReadFile(path)
        if err != nil {
            return err
        }

        hash := sha256.Sum256(content)
        checksums[relPath] = hash[:]
        return nil
    })

    return checksums, nil
}
```

## API Contracts

### GitHub Releases API

**Request**:
```
GET https://api.github.com/repos/specledger/specledger/releases/latest
Accept: application/vnd.github.v3+json
User-Agent: SpecLedger-CLI/{version}
```

**Response (Success)**:
```json
{
  "tag_name": "v1.2.3",
  "name": "SpecLedger v1.2.3",
  "html_url": "https://github.com/specledger/specledger/releases/tag/v1.2.3",
  "published_at": "2026-02-20T10:00:00Z"
}
```

**Response (Error - Rate Limited)**:
```json
{
  "message": "API rate limit exceeded",
  "documentation_url": "https://docs.github.com/rest"
}
```

**Error Handling**:
- HTTP 403 (rate limit): Skip version check, continue
- HTTP 404: Skip version check, continue
- Network timeout: Skip version check, continue
- Parse error: Skip version check, continue

### Doctor JSON Output

**Request**:
```
sl doctor --json
```

**Response (All checks pass, updates available)**:
```json
{
  "status": "pass",
  "tools": [...],
  "cli_version": "1.0.0",
  "cli_latest_version": "1.2.0",
  "cli_update_available": true,
  "cli_update_instructions": "Download from https://github.com/specledger/specledger/releases/latest",
  "template_version": "0.9.0",
  "template_update_available": true,
  "template_customized_files": ["commands/custom-workflow.md"]
}
```

**Response (Network error, CLI check skipped)**:
```json
{
  "status": "pass",
  "tools": [...],
  "cli_version": "1.0.0",
  "cli_latest_version": "",
  "cli_update_available": false,
  "cli_update_instructions": "",
  "cli_check_error": "network timeout"
}
```
