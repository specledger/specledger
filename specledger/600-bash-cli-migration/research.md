# Research: Bash Script to Go CLI Migration

**Feature**: 600-bash-cli-migration | **Date**: 2026-03-02

## Reference: 598 Bash Script Analysis

From 598-sdd-workflow-streamline research, the following bash scripts need migration:

### Scripts to Replace

| Script | Lines | Go Command | Complexity |
|--------|-------|------------|------------|
| `common.sh` | 205 | Absorbed into `pkg/cli/spec/` | Medium |
| `check-prerequisites.sh` | 167 | `sl spec info` | Medium |
| `create-new-feature.sh` | 307 | `sl spec create` | High |
| `setup-plan.sh` | ~50 | `sl spec setup-plan` | Low |
| `update-agent-context.sh` | 800 | `sl context update` | High |
| `adopt-feature-branch.sh` | ~100 | Superseded by ContextDetector | N/A |

---

## Script Analysis

### common.sh - Core Functions

**Purpose**: Shared functions for all bash scripts

**Key Functions**:
1. `get_repo_root()` - Find git toplevel or fallback to script location
2. `get_current_branch()` - Get current branch from git or SPECIFY_FEATURE env
3. `has_git()` - Check if git is available
4. `check_feature_branch()` - Validate branch name pattern `^\d{3,}-`
5. `get_feature_paths()` - Generate all feature-related paths
6. `check_mapped_branch()` - Check branch-map.json for aliases

**Go Migration**:
```go
// pkg/cli/spec/detector.go
type FeatureContext struct {
    RepoRoot     string
    Branch       string
    FeatureDir   string
    SpecFile     string
    PlanFile     string
    TasksFile    string
    HasGit       bool
}

func DetectFeatureContext() (*FeatureContext, error)
```

**Dependencies Eliminated**:
- `git` CLI → `go-git/v5` (already in `pkg/cli/git/git.go`)
- `jq` for branch-map.json → `encoding/json`

---

### check-prerequisites.sh - Feature Validation

**Purpose**: Validate feature state before proceeding

**Key Features**:
- JSON output mode with `--json`
- Path-only mode with `--paths-only`
- Requirement flags: `--require-plan`, `--require-tasks`
- Doc discovery: `--include-tasks`

**Output Structure**:
```json
{
  "FEATURE_DIR": "/path/to/specledger/600-feature",
  "AVAILABLE_DOCS": ["research.md", "data-model.md"]
}
```

**Go Migration**:
```go
// pkg/cli/commands/spec_info.go
type SpecInfoOutput struct {
    FeatureDir    string   `json:"FEATURE_DIR"`
    Branch        string   `json:"BRANCH"`
    FeatureSpec   string   `json:"FEATURE_SPEC"`
    AvailableDocs []string `json:"AVAILABLE_DOCS,omitempty"`
}

func NewSpecInfoCmd() *cobra.Command
```

**Edge Cases**:
- Detached HEAD → Error with suggestion
- Missing spec directory → Error with guidance
- Non-feature branch → Error (unless mapped in branch-map.json)

---

### create-new-feature.sh - Branch Creation

**Purpose**: Create new feature branch and spec directory

**Key Features**:
1. **Branch name generation**:
   - Stop-word filtering (the, a, an, to, for, of, etc.)
   - Acronym preservation (OAuth2, API, JWT)
   - 244-byte GitHub limit enforcement
   - Smart word selection (3-4 meaningful words)

2. **Number assignment**:
   - Auto-detect highest from specs directory
   - Manual override with `--number`

3. **Template handling**:
   - Copy spec-template.md to new spec.md

**Stop Words List** (from script):
```
i, a, an, the, to, for, of, in, on, at, by, with, from,
is, are, was, were, be, been, being, have, has, had,
do, does, did, will, would, should, could, can, may,
might, must, shall, this, that, these, those, my, your,
our, their, want, need, add, get, set
```

**Go Migration**:
```go
// pkg/cli/spec/branch.go
var StopWords = map[string]bool{
    "i": true, "a": true, "an": true, "the": true,
    // ... full list
}

func GenerateBranchName(description string, number int) string
func FilterStopWords(words []string) []string
func PreserveAcronyms(original, word string) string
func TruncateToLimit(name string, maxBytes int) string
```

**244-byte Truncation Logic**:
```go
const MaxBranchLength = 244

func TruncateToLimit(name string, maxBytes int) string {
    if len(name) <= maxBytes {
        return name
    }
    // Preserve feature number prefix (###-)
    parts := strings.SplitN(name, "-", 2)
    if len(parts) != 2 {
        return name[:maxBytes]
    }
    prefix := parts[0] + "-"
    maxSuffix := maxBytes - len(prefix)
    truncated := parts[1][:maxSuffix]
    // Remove trailing hyphen
    truncated = strings.TrimRight(truncated, "-")
    return prefix + truncated
}
```

---

### setup-plan.sh - Plan Template

**Purpose**: Copy plan template to feature directory

**Key Features**:
- Simple file copy operation
- Template from `.specledger/templates/plan-template.md`
- Error if plan.md already exists

**Go Migration**:
```go
// pkg/cli/commands/spec_setup_plan.go
func NewSpecSetupPlanCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "setup-plan",
        Short: "Copy plan template to feature directory",
        RunE: func(cmd *cobra.Command, args []string) error {
            // 1. Detect feature context
            // 2. Check if plan.md exists (error if so)
            // 3. Copy template from embedded/templates/
            // 4. Output JSON if --json flag
        },
    }
}
```

**Template Source**: Use embedded templates from `pkg/embedded/templates/`

---

### update-agent-context.sh - Agent File Updates

**Purpose**: Update AI agent context files with plan metadata

**Key Features**:
1. **Plan parsing**: Extract Technical Context fields
   - Language/Version
   - Primary Dependencies
   - Storage
   - Project Type

2. **Agent file management**:
   - Support 17+ agent types
   - Create new files from template
   - Update existing files with preservation

3. **Marker preservation**:
   - `<!-- MANUAL ADDITIONS START -->`
   - `<!-- MANUAL ADDITIONS END -->`
   - Content between markers preserved

4. **Deduplication**:
   - Don't append duplicate entries
   - Check existing entries before adding

**Agent File Paths** (from script):
| Agent | File Path |
|-------|-----------|
| Claude | `CLAUDE.md` |
| Gemini | `GEMINI.md` |
| Copilot | `.github/agents/copilot-instructions.md` |
| Cursor | `.cursor/rules/specify-rules.mdc` |
| Qwen | `QWEN.md` |
| Windsurf | `.windsurf/rules/specify-rules.md` |
| Kilo Code | `.kilocode/rules/specify-rules.md` |
| Auggie | `.augment/rules/specify-rules.md` |
| Roo Code | `.roo/rules/specify-rules.md` |
| CodeBuddy | `CODEBUDDY.md` |
| Qoder | `QODER.md` |
| SHAI | `SHAI.md` |
| Amazon Q | `AGENTS.md` |
| IBM Bob | `AGENTS.md` |
| opencode | `AGENTS.md` |
| Codex | `AGENTS.md` |

**Go Migration**:
```go
// pkg/cli/context/parser.go
type TechnicalContext struct {
    Language    string
    Framework   string
    Storage     string
    ProjectType string
}

func ParseTechnicalContext(planPath string) (*TechnicalContext, error)

// pkg/cli/context/updater.go
type AgentUpdater struct {
    AgentType string
    FilePath  string
}

func (u *AgentUpdater) Update(ctx *TechnicalContext) error
func (u *AgentUpdater) PreserveManualAdditions(content string) string
func (u *AgentUpdater) DeduplicateEntries(entries []string) []string
```

**Marker Preservation Logic**:
```go
func PreserveManualAdditions(content string, newContent string) string {
    markerStart := "<!-- MANUAL ADDITIONS START -->"
    markerEnd := "<!-- MANUAL ADDITIONS END -->"

    // Extract manual additions from existing content
    startIdx := strings.Index(content, markerStart)
    endIdx := strings.Index(content, markerEnd)

    if startIdx == -1 || endIdx == -1 {
        // No markers, add them
        return newContent + "\n\n" + markerStart + "\n" + markerEnd
    }

    manualContent := content[startIdx:endIdx+len(markerEnd)]

    // Insert into new content
    newIdx := strings.Index(newContent, markerStart)
    if newIdx != -1 {
        // Replace placeholder with preserved content
        return newContent[:newIdx] + manualContent + newContent[strings.Index(newContent, markerEnd)+len(markerEnd):]
    }

    return newContent
}
```

---

## Existing Go Code Patterns

### pkg/cli/git/git.go (Reuse)

Already implements needed functions:
- `GetCurrentBranch(repoPath string) (string, error)`
- `BranchExists(repoPath, name string) (bool, error)`
- `CheckoutBranch(repoPath, name string) error`
- `IsFeatureBranch(name string) bool`

### pkg/cli/prerequisites/checker.go (Pattern)

Pattern for tool checking:
```go
type Tool struct {
    Name        string
    DisplayName string
    Category    metadata.ToolCategory
}

func CheckTool(tool Tool) ToolCheckResult
```

### pkg/cli/commands/*.go (Pattern)

Cobra command pattern:
```go
func NewExampleCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "example",
        Short: "Short description",
        RunE: func(cmd *cobra.Command, args []string) error {
            // Implementation
            return nil
        },
    }
    cmd.Flags().BoolP("json", "j", false, "Output in JSON format")
    return cmd
}
```

---

## Dependencies Eliminated

| Dependency | Used By | Go Replacement |
|------------|---------|----------------|
| `jq` | common.sh (branch-map.json parsing) | `encoding/json` |
| `grep` | check-prerequisites.sh (file checks) | `os.Stat()` |
| `grep` | create-new-feature.sh (number extraction) | `regexp` package |
| `grep` | update-agent-context.sh (field extraction) | `strings` + `regexp` |
| `sed` | create-new-feature.sh (name cleaning) | `strings.Replace`, `regexp` |
| `sed` | update-agent-context.sh (template substitution) | `strings.Replace` |
| `git` CLI | create-new-feature.sh (branch creation) | `go-git/v5` |

---

## Recommendations

### High Priority

1. **Create pkg/cli/spec package first** - Foundational for all commands
2. **Implement sl spec info** - Used by all other commands
3. **Implement sl spec create** - Used by /specledger.specify

### Medium Priority

4. **Implement sl context update** - Used by /specledger.plan
5. **Implement sl spec setup-plan** - Simple, used by /specledger.plan

### Low Priority

6. **Add comprehensive tests** - Unit tests for each package
7. **Add Windows CI** - GitHub Actions matrix for cross-platform

---

## Open Questions

1. Should we keep bash scripts as fallback during transition?
   - **Recommendation**: Yes, until AI commands updated in 599

2. How to handle branch-map.json migration?
   - **Recommendation**: Keep JSON format, parse with encoding/json

3. Should templates be embedded or file-based?
   - **Recommendation**: Use existing embedded templates in `pkg/embedded/`
