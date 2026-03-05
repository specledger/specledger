# Implementation Plan: Fix sl init on Windows

**Branch**: `603-fix-init-windows` | **Date**: 2026-03-05 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specledger/603-fix-init-windows/spec.md`

## Summary

`sl init` (and `sl new`) fail on Windows because `filepath.Join` is used to construct paths for `embed.FS` lookups. On Windows, `filepath.Join` produces backslash-separated paths (`templates\manifest.yaml`), but Go's `embed.FS` requires forward slashes on all platforms. This causes the manifest to fail to load, the playbook name to be empty, and metadata creation to error. The fix replaces `filepath.Join` with `path.Join` (from the `path` package) wherever the result is used as an `embed.FS` path. A secondary fix addresses `checkGum()` using `command -v` (a Unix shell built-in unavailable on Windows) and the post-init shell script execution.

## Technical Context

**Language/Version**: Go 1.24.2
**Primary Dependencies**: `path` (stdlib), `path/filepath` (stdlib), `io/fs` (stdlib), `embed` (stdlib)
**Storage**: Embedded filesystem (`embed.FS`) for templates; real filesystem for output
**Testing**: `go test` — unit tests in `pkg/cli/playbooks/`, integration tests in `tests/integration/`
**Target Platform**: Windows, macOS, Linux (cross-platform fix)
**Project Type**: Single CLI binary
**Performance Goals**: N/A (correctness fix only)
**Constraints**: No new dependencies; backward-compatible on macOS/Linux
**Scale/Scope**: 5 files changed, ~15 lines modified

## Constitution Check

*The project constitution file is in template form (not yet populated). Evaluating against project principles from CLAUDE.md.*

- [x] **Specification-First**: spec.md complete with prioritized user stories
- [x] **Test-First**: New unit tests planned before implementation; existing integration tests validate behavior
- [x] **Code Quality**: `gofmt`, `go vet`, `golangci-lint` — same toolchain as rest of project
- [x] **UX Consistency**: Bug fix; no new user flows introduced
- [x] **Performance**: No performance impact — path string construction only
- [x] **Observability**: No new logging needed; existing warning messages cover graceful skip paths
- [x] **Issue Tracking**: Focused 1-day fix, no epic required

**Complexity Violations**: None. This is a minimal targeted fix.

## Project Structure

### Documentation (this feature)

```text
specledger/603-fix-init-windows/
├── plan.md              # This file
├── research.md          # Phase 0 output
└── tasks.md             # Phase 2 output (/specledger.tasks)
```

### Source Code (changed files only)

```text
pkg/cli/playbooks/
├── manifest.go          # Fix: path.Join for embed.FS path
├── embedded.go          # Fix: path.Join for embed.FS paths (x2)
├── copy.go              # Fix: path.Join + string-trim for srcPath/embed.FS comparisons
└── manifest_test.go     # New: LoadManifest succeeds on all platforms

pkg/cli/commands/
└── bootstrap_helpers.go # Fix: path.Join for init.sh embed.FS path; Windows shell detection

pkg/cli/tui/
└── terminal.go          # Fix: exec.LookPath instead of exec.Command("command", "-v", ...)

tests/integration/
└── bootstrap_test.go    # Extend: assert no "Playbook copying failed" in sl init output
```

## Complexity Tracking

> No violations.

---

## Phase 0: Research

### Prior Work

- **[135-fix-missing-chmod-x]**: Fixed `copyEmbeddedFile` and `applyEmbeddedSkills` to set `0755` permissions. Touches same files but did not address path separator issue. Confirms pattern is pre-existing.
- **[591-issue-tracking-upgrade]**: Removed beads from `init.sh`. Current `init.sh` is minimal (prints success message, exports env vars) — safe to skip on Windows.

### Root Cause Analysis

| Location | Bug | Impact on Windows |
|----------|-----|-------------------|
| `manifest.go:12` | `filepath.Join("templates", "manifest.yaml")` → `templates\manifest.yaml` | `ReadFile` fails — manifest not loaded |
| `embedded.go:90` | Same pattern in `ValidatePlaybooks` | Validation fails silently |
| `embedded.go:97` | `filepath.Join(templatesDir, pb.Path)` for playbook path check | Playbook existence check fails |
| `copy.go:19` | `filepath.Join(srcDir, playbook.Path)` used as `srcPath` for embed.FS string comparisons | `strings.HasPrefix(path, srcPath+"/")` always false → all files skipped |
| `copy.go:57` | `filepath.Rel(srcPath, path)` where srcPath has `\` and path has `/` | Returns wrong result; relPath calculation breaks |
| `bootstrap_helpers.go:449` | `filepath.Join("templates", playbookName, "init.sh")` for `TemplatesFS.ReadFile` | Script not found on Windows |
| `terminal.go:93` | `exec.Command("command", "-v", "gum")` — `command` is a Unix shell built-in | Returns error on Windows (not found) |
| `bootstrap_helpers.go:483` | `exec.Command(tmpFile.Name())` on a `.sh` temp file | Windows cannot execute shell scripts directly |

### Decisions

| Decision | Rationale | Alternative Rejected |
|----------|-----------|---------------------|
| Use `path.Join` for embed.FS paths | `embed.FS` always uses `/`; `path` package always produces `/`-separated paths on all OS | `strings.Replace("\\", "/")`: treats symptom not cause; brittle |
| Use `exec.LookPath("gum")` for tool detection | Standard cross-platform Go stdlib idiom | OS-branching with `where` vs `which`: more code, harder to test |
| Try bash/sh as shell interpreter on Windows; skip if unavailable | init.sh is minimal; skip is safe; Git for Windows users still get the script | Port init.sh logic to Go: over-engineering for a 5-line script |

---

## Phase 1: Design & Contracts

### No API contracts needed

This is a targeted bug fix. No new CLI surface, no new data model, no new API endpoints.

### Precise Change Specification

#### Fix 1 — `pkg/cli/playbooks/manifest.go`

```diff
 import (
     "fmt"
-    "path/filepath"
+    "path"

     "gopkg.in/yaml.v3"
 )

 func LoadManifest(templatesDir string) (*PlaybookManifest, error) {
-    manifestPath := filepath.Join(templatesDir, "manifest.yaml")
+    manifestPath := path.Join(templatesDir, "manifest.yaml")
```

#### Fix 2 — `pkg/cli/playbooks/embedded.go`

```diff
 import (
     "fmt"
-    "path/filepath"
+    "path"
 )

-    manifestPath := filepath.Join(s.templatesDir, "manifest.yaml")
+    manifestPath := path.Join(s.templatesDir, "manifest.yaml")

-        playbookPath := filepath.Join(s.templatesDir, pb.Path)
+        playbookPath := path.Join(s.templatesDir, pb.Path)
```

#### Fix 3 — `pkg/cli/playbooks/copy.go`

```diff
 import (
     "fmt"
     "io/fs"
     "os"
+    "path"
     "path/filepath"
     "strings"
     "time"
 )

-    srcPath := filepath.Join(srcDir, playbook.Path)
+    srcPath := path.Join(srcDir, playbook.Path)

     // Replace filepath.Rel (OS-native) with string trim (embed.FS-safe)
-    relPath, err := filepath.Rel(srcPath, path)
-    if err != nil {
-        result.Errors = append(...)
-        return nil
-    }
+    relPath := strings.TrimPrefix(path, srcPath+"/")
```

> `filepath.Rel` is replaced by `strings.TrimPrefix` because:
> - We already checked `strings.HasPrefix(path, srcPath+"/")` on line 52
> - `path.Rel` does not exist in the `path` stdlib package
> - The TrimPrefix approach is simpler and correct given the prior HasPrefix guard

#### Fix 4 — `pkg/cli/commands/bootstrap_helpers.go`

Two changes in this file:

**4a — embed.FS path for init.sh (line 449)**:
```diff
+    "path"
     ...
-    initScriptPath := filepath.Join("templates", playbookName, "init.sh")
+    initScriptPath := path.Join("templates", playbookName, "init.sh")
```

**4b — Windows-compatible shell execution (lines ~476–503)**:
```diff
+    "runtime"
     ...
-    // #nosec G204 -- tmpFile.Name() is from os.CreateTemp, safe path
-    cmd := exec.Command(tmpFile.Name())
+    // #nosec G204 -- tmpFile.Name() is from os.CreateTemp, safe path
+    var cmd *exec.Cmd
+    if runtime.GOOS == "windows" {
+        shell := findWindowsShell()
+        if shell == "" {
+            // No Unix shell available on Windows — skip post-init script gracefully
+            return
+        }
+        cmd = exec.Command(shell, tmpFile.Name())
+    } else {
+        cmd = exec.Command(tmpFile.Name())
+    }
```

Add new helper function:
```go
// findWindowsShell looks for bash or sh shipped with Git for Windows or similar.
// Returns the path to the shell executable, or empty string if not found.
func findWindowsShell() string {
    for _, shell := range []string{"bash", "sh"} {
        if p, err := exec.LookPath(shell); err == nil {
            return p
        }
    }
    return ""
}
```

#### Fix 5 — `pkg/cli/tui/terminal.go`

```diff
 func checkGum() bool {
-    cmd := exec.Command("command", "-v", "gum")
-    return cmd.Run() == nil
+    _, err := exec.LookPath("gum")
+    return err == nil
 }
```

### Test Plan

#### New: `pkg/cli/playbooks/manifest_test.go`

```go
// TestLoadManifestPathForwardSlash verifies LoadManifest works on all platforms
// (catches regression where filepath.Join produced backslash paths for embed.FS)
func TestLoadManifestPathForwardSlash(t *testing.T) {
    manifest, err := LoadManifest("templates")
    if err != nil {
        t.Fatalf("LoadManifest failed (embed.FS path separator bug?): %v", err)
    }
    if len(manifest.Playbooks) == 0 {
        t.Error("Expected at least one playbook in manifest")
    }
}
```

#### New/Extend: `pkg/cli/playbooks/copy_test.go`

```go
// TestEmbedFSPathSeparator verifies embed.FS paths use forward slashes
func TestEmbedFSPathSeparator(t *testing.T) {
    // Forward slash must work
    if !Exists("templates/specledger") {
        t.Error("templates/specledger should exist (forward slash required)")
    }
    // Backslash must NOT work (catches regression)
    if Exists(`templates\specledger`) {
        t.Error(`templates\specledger should NOT exist in embed.FS`)
    }
}
```

#### Extend: `tests/integration/bootstrap_test.go`

Add assertion to `TestBootstrapInitInExistingDirectory`:
```go
if strings.Contains(string(output), "Playbook copying failed") {
    t.Errorf("sl init reported playbook failure: %s", string(output))
}
if strings.Contains(string(output), "playbook name is required") {
    t.Errorf("sl init failed with metadata error: %s", string(output))
}
```
