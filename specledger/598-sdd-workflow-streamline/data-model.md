# Data Model: SDD Workflow Streamline

**Branch**: `598-sdd-workflow-streamline` | **Date**: 2026-03-01

## Work Stream 1: `sl comment` (Comment CRUD)

### Entity: ReviewComment

Extracted from `pkg/cli/revise/types.go` and validated against Supabase `review_comments` table.

```go
// pkg/cli/comment/types.go

// ReviewComment maps to the Supabase review_comments table.
type ReviewComment struct {
    ID               string  `json:"id"`
    ChangeID         string  `json:"change_id"`
    FilePath         string  `json:"file_path"`
    Content          string  `json:"content"`
    SelectedText     *string `json:"selected_text"`
    Line             *int    `json:"line"`
    StartLine        *int    `json:"start_line"`
    IsResolved       bool    `json:"is_resolved"`
    AuthorID         string  `json:"author_id"`
    AuthorName       *string `json:"author_name"`
    AuthorEmail      *string `json:"author_email"`
    ParentCommentID  *string `json:"parent_comment_id"`
    IsAIGenerated    *bool   `json:"is_ai_generated"`
    TriggeredByUser  *string `json:"triggered_by_user_id"`
    CreatedAt        string  `json:"created_at"`
    UpdatedAt        string  `json:"updated_at"`
}

// CommentThread groups a parent comment with its replies.
type CommentThread struct {
    Parent  ReviewComment   `json:"parent"`
    Replies []ReviewComment `json:"replies"`
}

// CommentSummary is the compact JSON output for sl comment list (D21).
// Previews are truncated; reply_count replaces full thread data.
type CommentSummary struct {
    ID                  string `json:"id"`
    FilePath            string `json:"file_path"`
    ContentPreview      string `json:"content_preview"`       // First 120 chars
    SelectedTextPreview string `json:"selected_text_preview"` // First 80 chars
    AuthorName          string `json:"author_name"`
    ReplyCount          int    `json:"reply_count"`
    IsResolved          bool   `json:"is_resolved"`
    CreatedAt           string `json:"created_at"`
}

// CommentListOutput is the JSON output for sl comment list (compact).
type CommentListOutput struct {
    SpecKey    string           `json:"spec_key"`
    ChangeID   string           `json:"change_id"`
    Comments   []CommentSummary `json:"comments"`
    TotalCount int              `json:"total_count"`
    Hint       string           `json:"hint"` // Follow-up instruction for agents
}

// CommentDetail is the full JSON output for sl comment show.
// Includes complete content, full thread replies — no truncation.
type CommentDetail struct {
    ID              string         `json:"id"`
    FilePath        string         `json:"file_path"`
    Content         string         `json:"content"`
    SelectedText    *string        `json:"selected_text"`
    Line            *int           `json:"line"`
    StartLine       *int           `json:"start_line"`
    IsResolved      bool           `json:"is_resolved"`
    AuthorName      *string        `json:"author_name"`
    AuthorEmail     *string        `json:"author_email"`
    IsAIGenerated   *bool          `json:"is_ai_generated"`
    CreatedAt       string         `json:"created_at"`
    Replies         []ReplyDetail  `json:"replies"`
}

// ReplyDetail is a single reply in sl comment show output.
type ReplyDetail struct {
    ID         string `json:"id"`
    Content    string `json:"content"`
    AuthorName string `json:"author_name"`
    CreatedAt  string `json:"created_at"`
}

// CommentShowOutput is the JSON output for sl comment show.
type CommentShowOutput struct {
    Comments []CommentDetail `json:"comments"`
}

// CommentReplyInput is the input for sl comment reply.
type CommentReplyInput struct {
    ParentCommentID string `json:"parent_comment_id"`
    Content         string `json:"content"`
    AuthorName      string `json:"author_name"`
    AuthorEmail     string `json:"author_email"`
}

// CommentResolveInput is the input for sl comment resolve.
type CommentResolveInput struct {
    CommentID string `json:"comment_id"`
    Reason    string `json:"reason"` // Stored as a reply before resolving
}
```

### Entity: CommentClient (PostgREST)

Extracted from `pkg/cli/revise/client.go`. Shared across `sl comment` and `sl revise`.

```go
// pkg/cli/comment/client.go

// Client handles all review comment API operations via PostgREST.
type Client struct {
    BaseURL     string
    AnonKey     string
    AccessToken string
}

// Core operations:
// - GetProject(owner, name) → project_id
// - GetSpec(projectID, specKey) → spec_id
// - GetChange(specID) → change_id
// - ListComments(changeID, status) → []ReviewComment
// - ListReplies(changeID) → []ReviewComment
// - BuildThreads(comments, replies) → []CommentThread
// - ResolveComment(commentID) error
// - ResolveWithReplies(commentIDs []string) error
// - CreateReply(changeID, input CommentReplyInput) error
```

### Relationship: Supabase Query Chain

```
projects (repo_owner, repo_name)
    └── specs (project_id, spec_key)
        └── changes (spec_id, state="open")
            └── review_comments (change_id)
                └── review_comments (parent_comment_id) [self-ref: replies]
```

## Work Stream 2: `sl spec` + `sl context` (Bash Replacement)

### Entity: FeatureInfo

Output of `sl spec info`, replacing `check-prerequisites.sh`.

```go
// pkg/cli/spec/types.go

// FeatureInfo contains resolved feature paths and available docs.
type FeatureInfo struct {
    RepoRoot      string   `json:"REPO_ROOT"`
    Branch        string   `json:"BRANCH"`
    FeatureDir    string   `json:"FEATURE_DIR"`
    FeatureSpec   string   `json:"FEATURE_SPEC"`
    ImplPlan      string   `json:"IMPL_PLAN"`
    Tasks         string   `json:"TASKS"`
    HasGit        bool     `json:"HAS_GIT"`
    AvailableDocs []string `json:"AVAILABLE_DOCS,omitempty"`
}
```

### Entity: FeatureCreateResult

Output of `sl spec create`, replacing `create-new-feature.sh`.

```go
// pkg/cli/spec/types.go

// FeatureCreateResult is returned after creating a new feature.
type FeatureCreateResult struct {
    BranchName string `json:"BRANCH_NAME"`
    SpecFile   string `json:"SPEC_FILE"`
    FeatureNum string `json:"FEATURE_NUM"`
}

// BranchNameConfig controls branch name generation.
type BranchNameConfig struct {
    Description string
    ShortName   string // Optional override
    Number      int    // Optional override (0 = auto)
    MaxLength   int    // Default: 244 (GitHub limit)
}
```

### Entity: PlanSetupResult

Output of `sl spec setup-plan`, replacing `setup-plan.sh`.

```go
// pkg/cli/spec/types.go

// PlanSetupResult is returned after setting up the plan template.
type PlanSetupResult struct {
    FeatureSpec string `json:"FEATURE_SPEC"`
    ImplPlan    string `json:"IMPL_PLAN"`
    SpecsDir    string `json:"SPECS_DIR"`
    Branch      string `json:"BRANCH"`
    HasGit      bool   `json:"HAS_GIT"`
}
```

### Entity: AgentContextUpdate

Used by `sl context update`, replacing `update-agent-context.sh`.

```go
// pkg/cli/context/types.go

// PlanMetadata is extracted from plan.md for agent context updates.
type PlanMetadata struct {
    Language     string `json:"language"`
    Dependencies string `json:"dependencies"`
    Storage      string `json:"storage"`
    ProjectType  string `json:"project_type"`
}

// AgentFileMapping maps agent types to their context file paths.
type AgentFileMapping struct {
    AgentType string // e.g., "claude", "gemini", "copilot"
    FilePath  string // e.g., "CLAUDE.md", ".github/agents/copilot-instructions.md"
}

// UpdateResult reports what was changed during context update.
type UpdateResult struct {
    AgentType string `json:"agent_type"`
    FilePath  string `json:"file_path"`
    Action    string `json:"action"` // "created", "updated", "skipped"
}
```

## Validation Rules

### Comment Operations
- `content` is required (non-empty) for replies
- `reason` for resolve is optional but recommended (stored as reply content before resolution)
- Cascade resolution: resolving a parent also resolves all its replies
- Auth required for all write operations (reply, resolve)
- Read operations require auth (PostgREST RLS)

### Feature Creation
- Branch number: must be unique across local `specledger/*/` dirs AND remote branches
- Short name: 2-4 words, lowercase, hyphenated, no stop words
- Branch name: max 244 bytes (GitHub limit), truncated with warning if exceeded
- Spec directory: `specledger/<number>-<short-name>/`
- Spec file: copied from `.specledger/templates/spec-template.md`

### Agent Context Update
- Plan.md must exist and be readable
- Manual additions between `<!-- MANUAL ADDITIONS START/END -->` markers are preserved
- Atomic updates via temp file + rename
- Agent file created from template if not exists, updated if exists

## State Transitions

### Comment Lifecycle
```
[created] → is_resolved=false (open)
    ├── reply added → still open (replies inherit change_id)
    └── resolved → is_resolved=true
        └── cascade: all replies also resolved
```

### Feature Lifecycle (spec commands)
```
sl spec create → branch created + specledger/<num>-<name>/spec.md
    └── sl spec setup-plan → plan.md copied from template
        └── sl context update → agent files updated from plan data
```
