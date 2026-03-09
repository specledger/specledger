package revise

import "github.com/specledger/specledger/pkg/cli/comment"

type ReviewComment = comment.ReviewComment

type ThreadReply = comment.ThreadReply

type ProcessedComment struct {
	Comment  ReviewComment
	Guidance string
	Index    int
}

type RevisionContext struct {
	SpecKey  string
	Comments []PromptComment
}

// PromptComment is a single comment entry in the revision prompt template.
type PromptComment struct {
	Index    int    // 1-based display index
	ID       string // Comment UUID (internal, for resolution)
	FilePath string
	Target   string // selected_text, "Line N", or "General"
	Feedback string // Comment content
	Guidance string // Optional user guidance
	Replies  []ThreadReply
}

// AutoFixture is the fixture file structure for non-interactive automation mode.
type AutoFixture struct {
	Branch   string           `json:"branch"`
	Comments []FixtureComment `json:"comments"`
}

// FixtureComment is a single comment entry in the automation fixture.
type FixtureComment struct {
	FilePath     string `json:"file_path"`
	SelectedText string `json:"selected_text"`
	Guidance     string `json:"guidance"`
}
